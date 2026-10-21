package videochat

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/cron"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/livekit/protocol/livekit"
)

const speakRequestTTL = 5 * time.Minute
const speakerLeaseTTL = 30 * time.Minute

type SpeakRequest struct {
	ID                       string     `json:"id"`
	RoomName                 string     `json:"room_name"`
	RequesterUserID          string     `json:"requester_user_id"`
	RequesterLivekitIdentity string     `json:"requester_livekit_identity"`
	Status                   string     `json:"status"` // pending, approved, rejected, cancelled, expired, revoked
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
	ExpiresAt                time.Time  `json:"expires_at"`
	DecidedAt                *time.Time `json:"decided_at,omitempty"`
	DecidedByUserID          *string    `json:"decided_by_user_id,omitempty"`
}

type ListSpeakRequestsParams struct {
	Status string `json:"status,omitempty" query:"status"`
	Limit  int    `json:"limit,omitempty" query:"limit"`
}

type ListSpeakRequestsResponse struct {
	Requests []*SpeakRequest `json:"requests"`
}

type ApproveSpeakResponse struct {
	RequestID           string `json:"request_id"`
	ParticipantIdentity string `json:"participant_identity"`
	Status              string `json:"status"`
	CanPublish          bool   `json:"can_publish"`
}

type SpeakerStateResponse struct {
	RoomName        string        `json:"room_name"`
	IsSlotAvailable bool          `json:"is_slot_available"`
	ActiveSpeaker   *string       `json:"active_speaker,omitempty"`
	MyRequest       *SpeakRequest `json:"my_request,omitempty"`
}

// CreateSpeakRequest crea una richiesta di intervento da parte di un viewer nella stanza.
//
//encore:api auth method=POST path=/videochat/rooms/:roomName/speak-requests
func CreateSpeakRequest(ctx context.Context, roomName string) (*SpeakRequest, error) {
	uid, ok := auth.UserID()
	if !ok {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "must be logged in"}
	}
	if err := validateRoomName(roomName); err != nil {
		return nil, err
	}

	client, _, err := liveKitClient()
	if err != nil {
		return nil, err
	}
	ctxTimeout, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	// Verifica esistenza e stato della stanza LiveKit
	_, metadata, err := findBroadcast(ctxTimeout, client, roomName)
	if err != nil {
		return nil, err
	}

	// L'owner ha già permessi completi di pubblicazione
	if metadata.OwnerID == string(uid) {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "owner already has publishing permissions"}
	}

	// Le richieste pending scadute passano a 'expired': altrimenti l'indice
	// unico parziale (status = 'pending') bloccherebbe ogni nuova richiesta.
	if _, err := db.Exec(ctx, `
		UPDATE videochat_speak_requests
		SET status = 'expired', updated_at = NOW()
		WHERE room_name = $1 AND requester_user_id = $2 AND status = 'pending' AND expires_at <= NOW()
	`, roomName, string(uid)); err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to expire stale speak requests")
	}

	// Idempotenza: controlla se l'utente ha già una richiesta 'pending' attiva non scaduta
	existing, err := pendingSpeakRequest(ctx, roomName, string(uid))
	if err == nil {
		return existing, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, errs.WrapCode(err, errs.Internal, "failed to check existing speak request")
	}

	reqID, err := uuid.NewV4()
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to generate request id")
	}

	now := time.Now().UTC()
	expiresAt := now.Add(speakRequestTTL)

	// Una richiesta concorrente può aver vinto la corsa: in caso di violazione
	// dell'indice unico si restituisce quella già creata.
	var created SpeakRequest
	err = db.QueryRow(ctx, `
		INSERT INTO videochat_speak_requests (
			id, room_name, requester_user_id, requester_livekit_identity, status, created_at, updated_at, expires_at
		) VALUES ($1, $2, $3, $4, 'pending', $5, $5, $6)
		RETURNING id, room_name, requester_user_id, requester_livekit_identity, status, created_at, updated_at, expires_at
	`, reqID.String(), roomName, string(uid), string(uid), now, expiresAt).Scan(
		&created.ID, &created.RoomName, &created.RequesterUserID, &created.RequesterLivekitIdentity,
		&created.Status, &created.CreatedAt, &created.UpdatedAt, &created.ExpiresAt,
	)
	if isUniqueViolation(err) {
		if existing, err := pendingSpeakRequest(ctx, roomName, string(uid)); err == nil {
			return existing, nil
		}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to create speak request")
	}

	rlog.Info("created speak request", "request_id", created.ID, "room", roomName, "user", uid)
	return &created, nil
}

// pendingSpeakRequest returns the caller's active pending request in the room.
func pendingSpeakRequest(ctx context.Context, roomName, userID string) (*SpeakRequest, error) {
	var r SpeakRequest
	err := db.QueryRow(ctx, `
		SELECT id, room_name, requester_user_id, requester_livekit_identity, status,
		       created_at, updated_at, expires_at, decided_at, decided_by_user_id
		FROM videochat_speak_requests
		WHERE room_name = $1 AND requester_user_id = $2 AND status = 'pending' AND expires_at > NOW()
	`, roomName, userID).Scan(
		&r.ID, &r.RoomName, &r.RequesterUserID, &r.RequesterLivekitIdentity,
		&r.Status, &r.CreatedAt, &r.UpdatedAt, &r.ExpiresAt,
		&r.DecidedAt, &r.DecidedByUserID,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// ListSpeakRequests elenca le richieste di intervento della stanza. Richiede l'owner.
//
//encore:api auth method=GET path=/videochat/rooms/:roomName/speak-requests
func ListSpeakRequests(ctx context.Context, roomName string, p *ListSpeakRequestsParams) (*ListSpeakRequestsResponse, error) {
	if _, ok := auth.UserID(); !ok {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "must be logged in"}
	}
	if err := validateRoomName(roomName); err != nil {
		return nil, err
	}

	client, _, err := liveKitClient()
	if err != nil {
		return nil, err
	}
	ctxTimeout, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	if err := requireRoomOwner(ctxTimeout, client, roomName); err != nil {
		return nil, err
	}

	limit := 50
	if p.Limit > 0 && p.Limit <= 100 {
		limit = p.Limit
	}

	var rows *sqldb.Rows
	if p.Status != "" {
		rows, err = db.Query(ctx, `
			SELECT id, room_name, requester_user_id, requester_livekit_identity, status,
			       created_at, updated_at, expires_at, decided_at, decided_by_user_id
			FROM videochat_speak_requests
			WHERE room_name = $1 AND status = $2
			ORDER BY created_at DESC
			LIMIT $3
		`, roomName, p.Status, limit)
	} else {
		rows, err = db.Query(ctx, `
			SELECT id, room_name, requester_user_id, requester_livekit_identity, status,
			       created_at, updated_at, expires_at, decided_at, decided_by_user_id
			FROM videochat_speak_requests
			WHERE room_name = $1
			ORDER BY created_at DESC
			LIMIT $2
		`, roomName, limit)
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to query speak requests")
	}
	defer rows.Close()

	var list []*SpeakRequest
	for rows.Next() {
		var r SpeakRequest
		if err := rows.Scan(
			&r.ID, &r.RoomName, &r.RequesterUserID, &r.RequesterLivekitIdentity, &r.Status,
			&r.CreatedAt, &r.UpdatedAt, &r.ExpiresAt, &r.DecidedAt, &r.DecidedByUserID,
		); err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "failed to scan speak request")
		}
		list = append(list, &r)
	}
	if list == nil {
		list = []*SpeakRequest{}
	}

	return &ListSpeakRequestsResponse{Requests: list}, nil
}

// CancelSpeakRequest annulla una richiesta ancora pending. Può farlo il richiedente o l'owner.
//
//encore:api auth method=DELETE path=/videochat/rooms/:roomName/speak-requests/:requestId
func CancelSpeakRequest(ctx context.Context, roomName, requestId string) error {
	uid, ok := auth.UserID()
	if !ok {
		return &errs.Error{Code: errs.Unauthenticated, Message: "must be logged in"}
	}
	if err := validateRoomName(roomName); err != nil {
		return err
	}

	client, _, err := liveKitClient()
	if err != nil {
		return err
	}
	ctxTimeout, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	room, metadata, err := findBroadcast(ctxTimeout, client, roomName)
	if err != nil {
		return err
	}
	_ = room

	isOwner := metadata.OwnerID == string(uid)

	var requesterID, status string
	err = db.QueryRow(ctx, `
		SELECT requester_user_id, status 
		FROM videochat_speak_requests 
		WHERE id = $1 AND room_name = $2
	`, requestId, roomName).Scan(&requesterID, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return &errs.Error{Code: errs.NotFound, Message: "speak request not found"}
	} else if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to query request")
	}

	if !isOwner && requesterID != string(uid) {
		return &errs.Error{Code: errs.PermissionDenied, Message: "not allowed to cancel this request"}
	}

	if status != "pending" {
		return &errs.Error{Code: errs.FailedPrecondition, Message: "request is not in pending state"}
	}

	_, err = db.Exec(ctx, `
		UPDATE videochat_speak_requests
		SET status = 'cancelled', updated_at = NOW()
		WHERE id = $1
	`, requestId)
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to cancel speak request")
	}

	return nil
}

// ApproveSpeakRequest approva una richiesta pendente, riserva lo slot e promuove il partecipante su LiveKit senza riconnessione.
//
//encore:api auth method=POST path=/videochat/rooms/:roomName/speak-requests/:requestId/approve
func ApproveSpeakRequest(ctx context.Context, roomName, requestId string) (*ApproveSpeakResponse, error) {
	uid, ok := auth.UserID()
	if !ok {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "must be logged in"}
	}
	if err := validateRoomName(roomName); err != nil {
		return nil, err
	}

	client, _, err := liveKitClient()
	if err != nil {
		return nil, err
	}
	ctxTimeout, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	if err := requireRoomOwner(ctxTimeout, client, roomName); err != nil {
		return nil, err
	}

	// Fase 1: Lock advisory su database e prenotazione dello slot
	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to begin transaction")
	}
	defer tx.Rollback()

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, "videochat/speaker/"+roomName); err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to acquire speaker slot lock")
	}

	var req SpeakRequest
	err = tx.QueryRow(ctx, `
		SELECT id, room_name, requester_user_id, requester_livekit_identity, status, expires_at
		FROM videochat_speak_requests
		WHERE id = $1 AND room_name = $2
	`, requestId, roomName).Scan(
		&req.ID, &req.RoomName, &req.RequesterUserID, &req.RequesterLivekitIdentity, &req.Status, &req.ExpiresAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, &errs.Error{Code: errs.NotFound, Message: "speak request not found"}
	} else if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to fetch request")
	}

	if req.Status != "pending" {
		return nil, &errs.Error{Code: errs.FailedPrecondition, Message: "request is not pending"}
	}
	if time.Now().UTC().After(req.ExpiresAt) {
		_, _ = tx.Exec(ctx, `UPDATE videochat_speak_requests SET status = 'expired', updated_at = NOW() WHERE id = $1`, req.ID)
		_ = tx.Commit()
		return nil, &errs.Error{Code: errs.FailedPrecondition, Message: "speak request has expired"}
	}

	// Controlla se lo slot è attualmente occupato da un altro speaker attivo o prenotato
	var currentSlotState sql.NullString
	var currentSlotSpeaker sql.NullString
	var leaseExpires sql.NullTime
	_ = tx.QueryRow(ctx, `
		SELECT state, speaker_user_id, lease_expires_at
		FROM videochat_speaker_slots
		WHERE room_name = $1
	`, roomName).Scan(&currentSlotState, &currentSlotSpeaker, &leaseExpires)

	if currentSlotState.Valid && currentSlotState.String != "free" {
		if !leaseExpires.Valid || leaseExpires.Time.After(time.Now().UTC()) {
			return nil, &errs.Error{Code: errs.AlreadyExists, Message: "speaker slot is currently occupied by another participant"}
		}
	}

	now := time.Now().UTC()
	leaseEnd := now.Add(speakerLeaseTTL)

	// Riserva lo slot temporaneamente (state = 'reserved')
	_, err = tx.Exec(ctx, `
		INSERT INTO videochat_speaker_slots (
			room_name, speaker_user_id, speaker_livekit_identity, request_id, state, updated_at, lease_expires_at
		) VALUES ($1, $2, $3, $4, 'reserved', $5, $6)
		ON CONFLICT (room_name) DO UPDATE SET
			speaker_user_id = EXCLUDED.speaker_user_id,
			speaker_livekit_identity = EXCLUDED.speaker_livekit_identity,
			request_id = EXCLUDED.request_id,
			state = 'reserved',
			updated_at = EXCLUDED.updated_at,
			lease_expires_at = EXCLUDED.lease_expires_at
	`, roomName, req.RequesterUserID, req.RequesterLivekitIdentity, req.ID, now, leaseEnd)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to reserve speaker slot")
	}

	if err := tx.Commit(); err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to commit slot reservation")
	}

	// Fase 2: Chiamata LiveKit per aggiornare i permessi del partecipante
	// Assegna CanPublish=true (con limitazione a Camera e Microfono) senza riconnessione
	_, lkErr := client.UpdateParticipant(ctxTimeout, &livekit.UpdateParticipantRequest{
		Room:     roomName,
		Identity: req.RequesterLivekitIdentity,
		Permission: &livekit.ParticipantPermission{
			CanPublish:     true,
			CanSubscribe:   true,
			CanPublishData: false,
			CanPublishSources: []livekit.TrackSource{
				livekit.TrackSource_CAMERA,
				livekit.TrackSource_MICROPHONE,
			},
		},
	})

	if lkErr != nil {
		// Compensazione: rilascia lo slot riservato
		rlog.Error("livekit update participant failed, releasing slot", "err", lkErr, "identity", req.RequesterLivekitIdentity)
		_, _ = db.Exec(ctx, `
			UPDATE videochat_speaker_slots
			SET state = 'free', speaker_user_id = NULL, speaker_livekit_identity = NULL, request_id = NULL, updated_at = NOW()
			WHERE room_name = $1 AND request_id = $2
		`, roomName, req.ID)
		return nil, liveKitError(lkErr)
	}

	// Fase 3: Conferma stato 'approved' e slot 'active'
	decidedBy := string(uid)
	_, err = db.Exec(ctx, `
		UPDATE videochat_speak_requests
		SET status = 'approved', decided_at = $1, decided_by_user_id = $2, updated_at = $1
		WHERE id = $3
	`, now, decidedBy, req.ID)
	if err != nil {
		rlog.Error("failed to set request approved in DB", "err", err)
	}

	_, err = db.Exec(ctx, `
		UPDATE videochat_speaker_slots
		SET state = 'active', updated_at = NOW()
		WHERE room_name = $1 AND request_id = $2
	`, roomName, req.ID)
	if err != nil {
		rlog.Error("failed to set slot active in DB", "err", err)
	}

	rlog.Info("approved speak request", "request_id", req.ID, "room", roomName, "identity", req.RequesterLivekitIdentity)

	return &ApproveSpeakResponse{
		RequestID:           req.ID,
		ParticipantIdentity: req.RequesterLivekitIdentity,
		Status:              "approved",
		CanPublish:          true,
	}, nil
}

// RejectSpeakRequest rifiuta una richiesta di intervento pendente. Richiede l'owner.
//
//encore:api auth method=POST path=/videochat/rooms/:roomName/speak-requests/:requestId/reject
func RejectSpeakRequest(ctx context.Context, roomName, requestId string) error {
	uid, ok := auth.UserID()
	if !ok {
		return &errs.Error{Code: errs.Unauthenticated, Message: "must be logged in"}
	}
	if err := validateRoomName(roomName); err != nil {
		return err
	}

	client, _, err := liveKitClient()
	if err != nil {
		return err
	}
	ctxTimeout, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	if err := requireRoomOwner(ctxTimeout, client, roomName); err != nil {
		return err
	}

	now := time.Now().UTC()
	decidedBy := string(uid)
	result, err := db.Exec(ctx, `
		UPDATE videochat_speak_requests
		SET status = 'rejected', decided_at = $1, decided_by_user_id = $2, updated_at = $1
		WHERE id = $3 AND room_name = $4 AND status = 'pending'
	`, now, decidedBy, requestId, roomName)
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to reject speak request")
	}

	rows := result.RowsAffected()
	if rows == 0 {
		return &errs.Error{Code: errs.NotFound, Message: "pending speak request not found"}
	}

	return nil
}

// RevokeSpeaker revoca i permessi di pubblicazione ad un partecipante e libera lo slot speaker. Richiede l'owner.
//
//encore:api auth method=POST path=/videochat/rooms/:roomName/participants/:identity/revoke
func RevokeSpeaker(ctx context.Context, roomName, identity string) error {
	if _, ok := auth.UserID(); !ok {
		return &errs.Error{Code: errs.Unauthenticated, Message: "must be logged in"}
	}
	if err := validateRoomName(roomName); err != nil {
		return err
	}

	client, _, err := liveKitClient()
	if err != nil {
		return err
	}
	ctxTimeout, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	if err := requireRoomOwner(ctxTimeout, client, roomName); err != nil {
		return err
	}

	// Chiamata LiveKit per revocare CanPublish
	_, lkErr := client.UpdateParticipant(ctxTimeout, &livekit.UpdateParticipantRequest{
		Room:     roomName,
		Identity: identity,
		Permission: &livekit.ParticipantPermission{
			CanPublish:     false,
			CanSubscribe:   true,
			CanPublishData: false,
		},
	})
	if lkErr != nil {
		rlog.Warn("livekit update participant error during revoke (participant may have disconnected)", "err", lkErr, "identity", identity)
	}

	// Libera lo slot e aggiorna lo stato
	_, _ = db.Exec(ctx, `
		UPDATE videochat_speaker_slots
		SET state = 'free', speaker_user_id = NULL, speaker_livekit_identity = NULL, request_id = NULL, updated_at = NOW()
		WHERE room_name = $1 AND (speaker_livekit_identity = $2 OR speaker_user_id = $2)
	`, roomName, identity)

	_, _ = db.Exec(ctx, `
		UPDATE videochat_speak_requests
		SET status = 'revoked', updated_at = NOW()
		WHERE room_name = $1 AND requester_livekit_identity = $2 AND status = 'approved'
	`, roomName, identity)

	rlog.Info("revoked speaker permissions", "room", roomName, "identity", identity)
	return nil
}

// GetSpeakerState restituisce lo stato dello slot speaker della stanza e l'eventuale richiesta del chiamante.
//
//encore:api auth method=GET path=/videochat/rooms/:roomName/speaker-state
func GetSpeakerState(ctx context.Context, roomName string) (*SpeakerStateResponse, error) {
	uid, ok := auth.UserID()
	if !ok {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "must be logged in"}
	}
	if err := validateRoomName(roomName); err != nil {
		return nil, err
	}

	var state, activeSpeaker sql.NullString
	var leaseExpires sql.NullTime
	_ = db.QueryRow(ctx, `
		SELECT state, speaker_livekit_identity, lease_expires_at
		FROM videochat_speaker_slots
		WHERE room_name = $1
	`, roomName).Scan(&state, &activeSpeaker, &leaseExpires)

	slotAvailable := true
	if state.Valid && state.String != "free" {
		if !leaseExpires.Valid || leaseExpires.Time.After(time.Now().UTC()) {
			slotAvailable = false
		}
	}

	var speakerPtr *string
	if !slotAvailable && activeSpeaker.Valid && activeSpeaker.String != "" {
		s := activeSpeaker.String
		speakerPtr = &s
	}

	// Recupera l'ultima richiesta attiva del chiamante
	var myReq SpeakRequest
	err := db.QueryRow(ctx, `
		SELECT id, room_name, requester_user_id, requester_livekit_identity, status,
		       created_at, updated_at, expires_at, decided_at, decided_by_user_id
		FROM videochat_speak_requests
		WHERE room_name = $1 AND requester_user_id = $2 AND status IN ('pending', 'approved')
		ORDER BY created_at DESC
		LIMIT 1
	`, roomName, string(uid)).Scan(
		&myReq.ID, &myReq.RoomName, &myReq.RequesterUserID, &myReq.RequesterLivekitIdentity,
		&myReq.Status, &myReq.CreatedAt, &myReq.UpdatedAt, &myReq.ExpiresAt,
		&myReq.DecidedAt, &myReq.DecidedByUserID,
	)

	var myReqPtr *SpeakRequest
	if err == nil {
		myReqPtr = &myReq
	}

	return &SpeakerStateResponse{
		RoomName:        roomName,
		IsSlotAvailable: slotAvailable,
		ActiveSpeaker:   speakerPtr,
		MyRequest:       myReqPtr,
	}, nil
}

var _ = cron.NewJob("expire-speaker-leases", cron.JobConfig{
	Title:    "Revoke guest speakers whose lease has expired",
	Every:    1 * cron.Minute,
	Endpoint: ExpireSpeakerLeases,
})

// ExpireSpeakerLeases revoca su LiveKit il permesso di pubblicare agli speaker
// il cui lease è scaduto e libera il relativo slot, così non possono esserci
// due speaker contemporanei.
//
//encore:api private
func ExpireSpeakerLeases(ctx context.Context) error {
	rows, err := db.Query(ctx, `
		SELECT room_name, COALESCE(speaker_livekit_identity, '')
		FROM videochat_speaker_slots
		WHERE state <> 'free' AND lease_expires_at IS NOT NULL AND lease_expires_at < NOW()
	`)
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to query expired speaker leases")
	}
	type expired struct{ room, identity string }
	var list []expired
	for rows.Next() {
		var e expired
		if err := rows.Scan(&e.room, &e.identity); err != nil {
			rows.Close()
			return errs.WrapCode(err, errs.Internal, "failed to scan expired speaker lease")
		}
		list = append(list, e)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to iterate expired speaker leases")
	}
	if len(list) == 0 {
		return nil
	}

	client, _, err := liveKitClient()
	if err != nil {
		return err
	}
	for _, e := range list {
		if e.identity != "" {
			ctxTimeout, cancel := context.WithTimeout(ctx, apiTimeout)
			_, lkErr := client.UpdateParticipant(ctxTimeout, &livekit.UpdateParticipantRequest{
				Room:     e.room,
				Identity: e.identity,
				Permission: &livekit.ParticipantPermission{
					CanPublish:     false,
					CanSubscribe:   true,
					CanPublishData: false,
				},
			})
			cancel()
			if lkErr != nil {
				// Il partecipante potrebbe essersi già disconnesso: lo slot va liberato comunque.
				rlog.Warn("livekit revoke on lease expiry failed", "err", lkErr, "room", e.room, "identity", e.identity)
			}
		}
		_, _ = db.Exec(ctx, `
			UPDATE videochat_speaker_slots
			SET state = 'free', speaker_user_id = NULL, speaker_livekit_identity = NULL, request_id = NULL, updated_at = NOW()
			WHERE room_name = $1 AND state <> 'free' AND lease_expires_at < NOW()
		`, e.room)
		_, _ = db.Exec(ctx, `
			UPDATE videochat_speak_requests
			SET status = 'revoked', updated_at = NOW()
			WHERE room_name = $1 AND requester_livekit_identity = $2 AND status = 'approved'
		`, e.room, e.identity)
		rlog.Info("speaker lease expired", "room", e.room, "identity", e.identity)
	}
	return nil
}
