package videochat

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"

	"encore.app/broadcasters"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
	lkauth "github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

// LiveStatus dice se il broadcaster della pagina pubblica /br/:shortID è in
// diretta. RoomName serve a chi ha fatto login per chiedere un token con
// CreateToken; RoomSID è la chiave della chat della diretta.
type LiveStatus struct {
	Live        bool   `json:"live"`
	DisplayName string `json:"display_name"`
	RoomName    string `json:"room_name,omitempty"`
	RoomSID     string `json:"room_sid,omitempty"`
}

// liveBroadcast trova la stanza del broadcaster con quello short_id. La
// stanza è in diretta solo se il proprietario è collegato e pubblica almeno
// una traccia; altrimenti room è nil.
func liveBroadcast(ctx context.Context, client *lksdk.RoomServiceClient, shortID string) (*broadcasters.LiveOwner, *livekit.Room, error) {
	owner, err := broadcasters.GetLiveOwner(ctx, shortID)
	if err != nil {
		if errs.Code(err) == errs.NotFound {
			return nil, nil, &errs.Error{Code: errs.NotFound, Message: "broadcaster not found"}
		}
		return nil, nil, err
	}
	result, err := client.ListRooms(ctx, &livekit.ListRoomsRequest{})
	if err != nil {
		return nil, nil, liveKitError(err)
	}
	for _, room := range result.Rooms {
		metadata, ok := readBroadcast(room)
		if !ok || metadata.OwnerID != owner.UserID {
			continue
		}
		participants, err := client.ListParticipants(ctx, &livekit.ListParticipantsRequest{Room: room.Name})
		if err != nil {
			return nil, nil, liveKitError(err)
		}
		for _, p := range participants.Participants {
			if p.Identity == owner.UserID && len(p.Tracks) > 0 {
				return owner, room, nil
			}
		}
	}
	return owner, nil, nil
}

// GetLive restituisce lo stato della diretta di un broadcaster per il sito
// pubblico. Non richiede login.
//
//encore:api public method=GET path=/videochat/live/:shortID
func GetLive(ctx context.Context, shortID string) (*LiveStatus, error) {
	client, _, err := liveKitClient()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()
	owner, room, err := liveBroadcast(ctx, client, shortID)
	if err != nil {
		return nil, err
	}
	status := &LiveStatus{DisplayName: owner.DisplayName}
	if room != nil {
		status.Live = true
		status.RoomName = room.Name
		status.RoomSID = room.Sid
	}
	return status, nil
}

// CreateGuestToken permette a un visitatore senza login di guardare una
// diretta: può solo ricevere audio, video e messaggi, non pubblica nulla, non
// scrive in chat ed è nascosto agli altri partecipanti.
//
//encore:api public method=POST path=/videochat/live/:shortID/token
func CreateGuestToken(ctx context.Context, shortID string) (*TokenResponse, error) {
	client, serverURL, err := liveKitClient()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()
	owner, room, err := liveBroadcast(ctx, client, shortID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, &errs.Error{Code: errs.FailedPrecondition, Message: "broadcaster is not live"}
	}
	suffix := make([]byte, 8)
	if _, err := rand.Read(suffix); err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to create guest identity")
	}
	identity := "guest-" + hex.EncodeToString(suffix)
	token, err := signGuestToken(secrets.LiveKitAPIKey, secrets.LiveKitAPISecret, identity, room.Name)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to generate LiveKit token")
	}
	return &TokenResponse{OwnerID: owner.UserID, Role: "guest",
		ServerURL: serverURL, ParticipantToken: token, RoomName: room.Name,
		Identity: identity, ExpiresIn: int64(tokenTTL.Seconds()),
	}, nil
}

// ListLiveMessages restituisce gli ultimi messaggi della chat di una diretta
// in corso, in sola lettura per i visitatori senza login.
//
//encore:api public method=GET path=/videochat/live/:shortID/messages
func ListLiveMessages(ctx context.Context, shortID string, p *ListMessagesParams) (*ListMessagesResponse, error) {
	client, _, err := liveKitClient()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()
	_, room, err := liveBroadcast(ctx, client, shortID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return &ListMessagesResponse{Messages: []*ChatMessage{}}, nil
	}
	limit := 50
	if p.Limit > 0 && p.Limit <= 100 {
		limit = p.Limit
	}
	return recentMessages(ctx, room.Sid, limit)
}

func signGuestToken(key, secret, identity, room string) (string, error) {
	grant := &lkauth.VideoGrant{RoomJoin: true, Room: room, Hidden: true}
	grant.SetCanPublish(false)
	grant.SetCanSubscribe(true)
	grant.SetCanPublishData(false)
	grant.SetCanUpdateOwnMetadata(false)
	return lkauth.NewAccessToken(key, secret).
		SetIdentity(identity).
		SetName("Guest").
		SetValidFor(tokenTTL).
		SetVideoGrant(grant).
		ToJWT()
}

// MyLive dice a chi ha fatto login se è un broadcaster e se può andare in
// diretta dal sito pubblico. Abilitazione e nome della stanza vengono dal
// personal setup del profilo (personal_setup.can_broadcast e room_name).
type MyLive struct {
	// IsBroadcaster: utente broadcaster (member_type, profilo broadcaster o
	// personal setup); il sito mostra il comando "Apri stanza".
	IsBroadcaster bool `json:"is_broadcaster"`
	// CanBroadcast: broadcasting abilitato nel personal setup, con un nome stanza.
	CanBroadcast bool   `json:"can_broadcast"`
	ShortID      string `json:"short_id,omitempty"`
	DisplayName  string `json:"display_name,omitempty"`
	RoomName     string `json:"room_name,omitempty"`
	RoomSID      string `json:"room_sid,omitempty"`
}

// myLive legge il personal setup di chi chiama e, se c'è, il suo profilo
// broadcaster abilitato (per lo short_id della pagina pubblica).
func myLive(ctx context.Context) (*MyLive, string, error) {
	uid, ok := auth.UserID()
	if !ok {
		return nil, "", &errs.Error{Code: errs.Unauthenticated, Message: "must be logged in"}
	}
	userID, err := uuid.FromString(string(uid))
	if err != nil {
		return &MyLive{}, string(uid), nil
	}
	var (
		memberType   string
		displayName  string
		canBroadcast bool
		roomName     string
	)
	err = profilesDB.QueryRow(ctx, `
		SELECT p.member_type, p.display_name, COALESCE(s.can_broadcast, FALSE), COALESCE(s.room_name, '')
		FROM user_profiles p
		LEFT JOIN personal_setup s ON s.user_id = p.user_id
		WHERE p.user_id = $1
	`, userID).Scan(&memberType, &displayName, &canBroadcast, &roomName)
	if errors.Is(err, sqldb.ErrNoRows) {
		return &MyLive{}, string(uid), nil
	}
	if err != nil {
		return nil, "", errs.WrapCode(err, errs.Internal, "failed to read personal setup")
	}
	live := &MyLive{
		IsBroadcaster: memberType == "broadcaster" || canBroadcast,
		CanBroadcast:  canBroadcast && validateRoomName(roomName) == nil,
		DisplayName:   displayName,
	}
	if live.CanBroadcast {
		live.RoomName = roomName
	}
	owner, err := broadcasters.GetLiveOwnerByUser(ctx, userID)
	switch {
	case err == nil:
		live.IsBroadcaster = true
		live.ShortID = owner.ShortID
	case errs.Code(err) != errs.NotFound:
		return nil, "", err
	}
	return live, string(uid), nil
}

// GetMyLive dice se l'utente collegato è un broadcaster, se il broadcasting è
// abilitato e qual è il nome della sua stanza. Non crea stanze.
//
//encore:api auth method=GET path=/videochat/my-live
func GetMyLive(ctx context.Context) (*MyLive, error) {
	live, _, err := myLive(ctx)
	return live, err
}

// OpenMyLive restituisce la stanza del broadcaster collegato e la crea, con il
// nome del personal setup, se non esiste ancora. Il proprietario è sempre chi
// chiama; senza broadcasting abilitato risponde PermissionDenied.
//
//encore:api auth method=POST path=/videochat/my-live
func OpenMyLive(ctx context.Context) (*MyLive, error) {
	live, ownerID, err := myLive(ctx)
	if err != nil {
		return nil, err
	}
	if !live.CanBroadcast {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "broadcasting not enabled"}
	}
	client, _, err := liveKitClient()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	// Stesso lock di OpenRoom: una sola creazione alla volta per nome stanza.
	tx, err := profilesDB.Begin(ctx)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to open room")
	}
	defer tx.Rollback()
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, "videochat/open/"+live.RoomName); err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to lock room creation")
	}
	result, err := client.ListRooms(ctx, &livekit.ListRoomsRequest{})
	if err != nil {
		return nil, liveKitError(err)
	}
	for _, room := range result.Rooms {
		if room.Name != live.RoomName {
			continue
		}
		if metadata, ok := readBroadcast(room); !ok || metadata.OwnerID != ownerID {
			return nil, &errs.Error{Code: errs.FailedPrecondition, Message: "room name is already used by another room"}
		}
		_ = logBroadcastSession(ctx, ownerID, room.Sid, room.Name, "OPEN")
		live.RoomSID = room.Sid
		return live, nil
	}
	// Come OpenRoom: la stanza si mostra con il suo nome.
	metadata, err := json.Marshal(broadcastMetadata{Mode: "broadcast", OwnerID: ownerID, DisplayName: live.RoomName})
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to encode room metadata")
	}
	room, err := client.CreateRoom(ctx, &livekit.CreateRoomRequest{Name: live.RoomName, Metadata: string(metadata), EmptyTimeout: 300})
	if err != nil {
		return nil, liveKitError(err)
	}
	_ = logBroadcastSession(ctx, ownerID, room.Sid, room.Name, "OPEN")
	live.RoomSID = room.Sid
	return live, nil
}
