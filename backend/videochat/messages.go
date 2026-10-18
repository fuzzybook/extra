package videochat

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/types/uuid"
)

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
	Kind            string `json:"kind,omitempty"` // facoltativo, default 'text'
}

type ListMessagesParams struct {
	Limit int `json:"limit,omitempty" query:"limit"`
}

type ListMessagesResponse struct {
	Messages []*ChatMessage `json:"messages"`
}

// SendMessage invia e memorizza un messaggio di chat nella stanza in modo sicuro e idempotente.
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

	kind := "text"
	if p.Kind == "system" {
		kind = "system"
	}

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

	return &created, nil
}

// ListMessages recupera lo storico cronologico dei messaggi di chat della stanza.
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
		FROM videochat_messages
		WHERE room_name = $1
		ORDER BY created_at ASC
		LIMIT $2
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
	if list == nil {
		list = []*ChatMessage{}
	}

	return &ListMessagesResponse{Messages: list}, nil
}
