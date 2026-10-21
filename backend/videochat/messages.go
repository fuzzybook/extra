package videochat

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	authsvc "encore.app/auth"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/pubsub"
	"encore.dev/types/uuid"
)

// RoomEvent rappresenta un evento (messaggio o comando) della videochat distribuito via Pub/Sub.
type RoomEvent struct {
	EventID         string    `json:"event_id"`
	RoomName        string    `json:"room_name"`
	Type            string    `json:"type"` // "chat.message", "command.alert", "command.mute", "command.clear"
	SenderUserID    string    `json:"sender_user_id"`
	SenderName      string    `json:"sender_name"`
	SenderAvatar    string    `json:"sender_avatar,omitempty"`
	Body            string    `json:"body"`
	TargetID        string    `json:"target_id,omitempty"`
	ClientMessageID string    `json:"client_message_id,omitempty"`
	OccurredAt      time.Time `json:"occurred_at"`
}

// RoomEvents è il topic Pub/Sub Encore per tutti gli eventi e comandi della videochat.
// Permette in modo disaccoppiato l'elaborazione asincrona, moderazione, filtri e audit dei messaggi.
var RoomEvents = pubsub.NewTopic[*RoomEvent]("videochat-room-events", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

// Subscription per elaborare e filtrare i messaggi e comandi della stanza.
var _ = pubsub.NewSubscription(
	RoomEvents, "filter-and-process-room-events",
	pubsub.SubscriptionConfig[*RoomEvent]{
		Handler: handleRoomEvent,
	},
)

// handleRoomEvent è il punto di estensione centrale per il filtraggio e la gestione dei messaggi e comandi.
func handleRoomEvent(ctx context.Context, ev *RoomEvent) error {
	// 1. Filtraggio e moderazione (es. profanity filter, spam detection)
	if ev.Type == "chat.message" {
		_ = sanitizeMessageBody(ev.Body)
	}

	// 2. Gestione comandi di sistema/moderazione
	if strings.HasPrefix(ev.Type, "command.") {
		// Log o esecuzione di azioni di moderazione/audit
	}

	return nil
}

// sanitizeMessageBody normalizza e prepara il testo per futuri filtri di moderazione.
func sanitizeMessageBody(raw string) string {
	cleaned := strings.TrimSpace(raw)
	return cleaned
}

type ChatMessage struct {
	ID              string    `json:"id"`
	RoomName        string    `json:"room_name"`
	SenderUserID    string    `json:"sender_user_id"`
	SenderName      string    `json:"sender_name"`
	SenderAvatar    string    `json:"sender_avatar,omitempty"`
	Kind            string    `json:"kind"` // 'text' | 'system'
	Body            string    `json:"body"`
	ClientMessageID string    `json:"client_message_id"`
	CreatedAt       time.Time `json:"created_at"`
}

type SendMessageParams struct {
	ClientMessageID string `json:"client_message_id"`
	Text            string `json:"text"`
	Kind            string `json:"kind,omitempty"` // ignorato: i client inviano solo 'text'
}

type ListMessagesParams struct {
	Limit int `json:"limit,omitempty" query:"limit"`
}

type ListMessagesResponse struct {
	Messages []*ChatMessage `json:"messages"`
}

// allowedRoomCommands è l'elenco chiuso dei comandi di moderazione.
var allowedRoomCommands = map[string]bool{"alert": true, "mute": true, "clear": true, "filter": true}

type SendCommandParams struct {
	Command  string `json:"command"` // 'alert' | 'mute' | 'clear' | 'filter'
	TargetID string `json:"target_id,omitempty"`
	Payload  string `json:"payload,omitempty"`
}

type RoomCommandResponse struct {
	EventID    string    `json:"event_id"`
	RoomName   string    `json:"room_name"`
	Command    string    `json:"command"`
	Success    bool      `json:"success"`
	OccurredAt time.Time `json:"occurred_at"`
}

// SendMessage invia e memorizza un messaggio di chat nella stanza, pubblicando l'evento su Pub/Sub.
//
//encore:api auth method=POST path=/videochat/rooms/:roomName/messages
func SendMessage(ctx context.Context, roomName string, p *SendMessageParams) (*ChatMessage, error) {
	uid, ok := auth.UserID()
	if !ok {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "must be logged in"}
	}
	if err := validateRoomName(roomName); err != nil {
		return nil, err
	}

	text := strings.TrimSpace(p.Text)
	if text == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "message text cannot be empty"}
	}
	if len(text) > 2000 {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "message text exceeds maximum length of 2000 characters"}
	}

	clientMsgID := strings.TrimSpace(p.ClientMessageID)
	if clientMsgID == "" {
		newID, _ := uuid.NewV4()
		clientMsgID = newID.String()
	}

	client, _, err := liveKitClient()
	if err != nil {
		return nil, err
	}
	ctxTimeout, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	// Verifica che la stanza esista e sia aperta
	_, _, err = findBroadcast(ctxTimeout, client, roomName)
	if err != nil {
		return nil, err
	}

	// I messaggi "system" sono creati solo dal server: il campo Kind inviato
	// dal client viene ignorato.
	kind := "text"

	// Recupera display name e avatar del mittente da profilesDB
	senderName := "Utente"
	senderAvatar := ""
	_ = profilesDB.QueryRow(ctx, `
		SELECT display_name, avatar_url 
		FROM user_profiles 
		WHERE user_id = $1
	`, string(uid)).Scan(&senderName, &senderAvatar)

	msgID, err := uuid.NewV4()
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to generate message id")
	}

	// Idempotenza: controlla se il messaggio con lo stesso client_message_id esiste già
	var existing ChatMessage
	err = db.QueryRow(ctx, `
		SELECT id, room_name, sender_user_id, sender_name, sender_avatar, kind, body, client_message_id, created_at
		FROM videochat_messages
		WHERE room_name = $1 AND sender_user_id = $2 AND client_message_id = $3
	`, roomName, string(uid), clientMsgID).Scan(
		&existing.ID, &existing.RoomName, &existing.SenderUserID, &existing.SenderName,
		&existing.SenderAvatar, &existing.Kind, &existing.Body, &existing.ClientMessageID, &existing.CreatedAt,
	)
	if err == nil {
		return &existing, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, errs.WrapCode(err, errs.Internal, "failed to check existing message")
	}

	now := time.Now().UTC()
	var created ChatMessage
	err = db.QueryRow(ctx, `
		INSERT INTO videochat_messages (
			id, room_name, sender_user_id, sender_name, sender_avatar, kind, body, client_message_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, room_name, sender_user_id, sender_name, sender_avatar, kind, body, client_message_id, created_at
	`, msgID.String(), roomName, string(uid), senderName, senderAvatar, kind, text, clientMsgID, now).Scan(
		&created.ID, &created.RoomName, &created.SenderUserID, &created.SenderName,
		&created.SenderAvatar, &created.Kind, &created.Body, &created.ClientMessageID, &created.CreatedAt,
	)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to save chat message")
	}

	// Pubblica l'evento sul topic Pub/Sub Encore per consentire il filtraggio e la gestione asincrona
	_, _ = RoomEvents.Publish(ctx, &RoomEvent{
		EventID:         created.ID,
		RoomName:        created.RoomName,
		Type:            "chat.message",
		SenderUserID:    created.SenderUserID,
		SenderName:      created.SenderName,
		SenderAvatar:    created.SenderAvatar,
		Body:            created.Body,
		ClientMessageID: created.ClientMessageID,
		OccurredAt:      created.CreatedAt,
	})

	return &created, nil
}

// ListMessages recupera gli ultimi messaggi di chat della stanza, in ordine cronologico.
//
//encore:api auth method=GET path=/videochat/rooms/:roomName/messages
func ListMessages(ctx context.Context, roomName string, p *ListMessagesParams) (*ListMessagesResponse, error) {
	if _, ok := auth.UserID(); !ok {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "must be logged in"}
	}
	if err := validateRoomName(roomName); err != nil {
		return nil, err
	}

	limit := 50
	if p.Limit > 0 && p.Limit <= 100 {
		limit = p.Limit
	}

	rows, err := db.Query(ctx, `
		SELECT id, room_name, sender_user_id, sender_name, sender_avatar, kind, body, client_message_id, created_at
		FROM (
			SELECT id, room_name, sender_user_id, sender_name, sender_avatar, kind, body, client_message_id, created_at
			FROM videochat_messages
			WHERE room_name = $1
			ORDER BY created_at DESC, id DESC
			LIMIT $2
		) latest
		ORDER BY created_at ASC, id ASC
	`, roomName, limit)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to query messages")
	}
	defer rows.Close()

	var list []*ChatMessage
	for rows.Next() {
		var m ChatMessage
		if err := rows.Scan(
			&m.ID, &m.RoomName, &m.SenderUserID, &m.SenderName, &m.SenderAvatar,
			&m.Kind, &m.Body, &m.ClientMessageID, &m.CreatedAt,
		); err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "failed to scan chat message")
		}
		list = append(list, &m)
	}
	if err := rows.Err(); err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to iterate chat messages")
	}
	if list == nil {
		list = []*ChatMessage{}
	}

	return &ListMessagesResponse{Messages: list}, nil
}

// SendCommand pubblica un comando amministrativo o di moderazione sul topic Pub/Sub della stanza.
// Richiede il proprietario della stanza o un admin.
//
//encore:api auth method=POST path=/videochat/rooms/:roomName/commands
func SendCommand(ctx context.Context, roomName string, p *SendCommandParams) (*RoomCommandResponse, error) {
	uid, ok := auth.UserID()
	if !ok {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "must be logged in"}
	}
	if err := validateRoomName(roomName); err != nil {
		return nil, err
	}
	command := strings.TrimSpace(p.Command)
	if !allowedRoomCommands[command] {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "unknown command: allowed are alert, mute, clear, filter"}
	}
	if len(p.Payload) > 2000 {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "command payload exceeds maximum length of 2000 characters"}
	}

	// Solo il proprietario della stanza o un admin possono moderarla.
	if !authsvc.IsAdmin() {
		client, _, err := liveKitClient()
		if err != nil {
			return nil, err
		}
		ctxTimeout, cancel := context.WithTimeout(ctx, apiTimeout)
		defer cancel()
		if err := requireRoomOwner(ctxTimeout, client, roomName); err != nil {
			return nil, err
		}
	}

	senderName := "Admin"
	_ = profilesDB.QueryRow(ctx, `
		SELECT display_name FROM user_profiles WHERE user_id = $1
	`, string(uid)).Scan(&senderName)

	cmdID, _ := uuid.NewV4()
	now := time.Now().UTC()

	// Pubblicazione su Pub/Sub
	_, err := RoomEvents.Publish(ctx, &RoomEvent{
		EventID:      cmdID.String(),
		RoomName:     roomName,
		Type:         "command." + command,
		SenderUserID: string(uid),
		SenderName:   senderName,
		Body:         p.Payload,
		TargetID:     p.TargetID,
		OccurredAt:   now,
	})
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to publish room command")
	}

	return &RoomCommandResponse{
		EventID:    cmdID.String(),
		RoomName:   roomName,
		Command:    command,
		Success:    true,
		OccurredAt: now,
	}, nil
}
