package broadcasters

import (
	"context"
	"strings"
	"time"

	authsvc "encore.app/auth"
	"encore.dev/beta/errs"
)

// maxChatLogData limita la dimensione di un singolo evento salvato.
const maxChatLogData = 4 << 10 // 4 KB

type ChatLogEntry struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Room      string    `json:"room"`
	Type      string    `json:"type"`
	Data      string    `json:"data"`
}

type AddChatLogParams struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type ListChatLogParams struct {
	Limit int    `json:"limit,omitempty" query:"limit"`
	Type  string `json:"type,omitempty" query:"type"`
}

type ListChatLogResponse struct {
	Messages []*ChatLogEntry `json:"messages"`
}

// AddChatLog inserisce un nuovo evento nella tabella chatlog (registro di
// audit degli eventi della stanza, es. accensione di camera e microfono).
// Richiede autenticazione; il tipo CMD è riservato agli admin. I messaggi di
// chat passano invece da videochat.SendMessage.
//
//encore:api auth method=POST path=/broadcasters/id/:id/chatlog
func AddChatLog(ctx context.Context, id string, p *AddChatLogParams) (*ChatLogEntry, error) {
	if strings.TrimSpace(id) == "" || len(id) > 128 {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "id stanza obbligatorio (massimo 128 caratteri)"}
	}
	if p == nil || strings.TrimSpace(p.Data) == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "data obbligatorio"}
	}
	if len(p.Data) > maxChatLogData {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "data troppo grande (massimo 4 KB)"}
	}

	msgType := strings.ToUpper(strings.TrimSpace(p.Type))
	if msgType == "" || msgType == "ADMIN_ALERT" {
		msgType = "MESSAGE"
	} else if msgType == "ALERT" {
		msgType = "ALLERT"
	}
	switch msgType {
	case "ACTION", "MESSAGE", "ALLERT":
	case "CMD":
		if !authsvc.IsAdmin() {
			return nil, &errs.Error{Code: errs.PermissionDenied, Message: "il tipo CMD è riservato agli amministratori"}
		}
	default:
		msgType = "MESSAGE"
	}

	entry := &ChatLogEntry{
		Room: id,
		Type: msgType,
		Data: p.Data,
	}

	err := db.QueryRow(ctx, `
		INSERT INTO chatlog (room, type, data, timestamp)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, timestamp
	`, entry.Room, entry.Type, entry.Data).Scan(&entry.ID, &entry.Timestamp)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "impossibile salvare il messaggio di chat")
	}

	return entry, nil
}

// ListChatLog preleva gli ultimi eventi per la stanza (default 30) e li ordina
// in ordine cronologico crescente (dal più vecchio al più recente). Solo admin.
//
//encore:api auth method=GET path=/broadcasters/id/:id/chatlog
func ListChatLog(ctx context.Context, id string, p *ListChatLogParams) (*ListChatLogResponse, error) {
	if err := authsvc.RequireAdmin(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(id) == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "id stanza obbligatorio"}
	}

	limit := 30
	filterType := ""
	if p != nil {
		if p.Limit > 0 {
			limit = p.Limit
		}
		if t := strings.TrimSpace(p.Type); t != "" {
			filterType = strings.ToUpper(t)
		}
	}
	if limit > 100 {
		limit = 100
	}

	query := `
		SELECT id, timestamp, room, type, data
		FROM (
			SELECT id, timestamp, room, type, data
			FROM chatlog
			WHERE room = $1 AND ($3 = '' OR type = $3)
			ORDER BY timestamp DESC, id DESC
			LIMIT $2
		) sub
		ORDER BY timestamp ASC, id ASC
	`

	rows, err := db.Query(ctx, query, id, limit, filterType)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "impossibile recuperare i messaggi di chat")
	}
	defer rows.Close()

	var list []*ChatLogEntry
	for rows.Next() {
		var e ChatLogEntry
		if err := rows.Scan(&e.ID, &e.Timestamp, &e.Room, &e.Type, &e.Data); err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "errore nella lettura dei messaggi")
		}
		list = append(list, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "errore iterazione messaggi")
	}

	if list == nil {
		list = []*ChatLogEntry{}
	}

	return &ListChatLogResponse{Messages: list}, nil
}
