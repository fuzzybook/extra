package videochat

import (
	"context"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
)

var db = sqldb.NewDatabase("videochat", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

type BroadcastSession struct {
	ID        int64     `json:"id"`
	OwnerID   string    `json:"owner_id"`
	SID       string    `json:"sid"`
	Name      string    `json:"name"`
	Action    string    `json:"action"` // "OPEN" o "CLOSE"
	CreatedAt time.Time `json:"created_at"`
}

type ListSessionsParams struct {
	OwnerID string `json:"owner_id,omitempty" query:"owner_id"`
	Limit   int    `json:"limit,omitempty" query:"limit"`
}

type ListSessionsResponse struct {
	Sessions []*BroadcastSession `json:"sessions"`
}

// logBroadcastSession inserisce un record di apertura o chiusura trasmissione in modo idempotente.
func logBroadcastSession(ctx context.Context, ownerID, sid, name, action string) error {
	if ownerID == "" || sid == "" || name == "" {
		return nil
	}

	if action == "OPEN" {
		var exists bool
		_ = db.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM broadcast_sessions 
				WHERE sid = $1 AND action = 'OPEN'
			)
		`, sid).Scan(&exists)
		if exists {
			return nil
		}
	} else if action == "CLOSE" {
		var exists bool
		_ = db.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM broadcast_sessions 
				WHERE sid = $1 AND action = 'CLOSE'
			)
		`, sid).Scan(&exists)
		if exists {
			return nil
		}
	}

	_, err := db.Exec(ctx, `
		INSERT INTO broadcast_sessions (owner_id, sid, name, action, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`, ownerID, sid, name, action)
	if err != nil {
		rlog.Error("failed to insert broadcast session", "err", err, "owner_id", ownerID, "sid", sid, "action", action)
		return err
	}
	rlog.Info("logged broadcast session", "owner_id", ownerID, "sid", sid, "name", name, "action", action)
	return nil
}

// syncClosedSessions rileva le sessioni che risultavano aperte ma la cui stanza LiveKit non è più attiva,
// e registra automaticamente la chiusura (CLOSE).
func syncClosedSessions(ctx context.Context, activeSIDs map[string]bool) {
	rows, err := db.Query(ctx, `
		SELECT o.owner_id, o.sid, o.name
		FROM broadcast_sessions o
		WHERE o.action = 'OPEN'
		  AND NOT EXISTS (
		      SELECT 1 FROM broadcast_sessions c
		      WHERE c.sid = o.sid AND c.action = 'CLOSE'
		  )
	`)
	if err != nil {
		return
	}
	defer rows.Close()

	type toClose struct {
		ownerID, sid, name string
	}
	var list []toClose
	for rows.Next() {
		var c toClose
		if err := rows.Scan(&c.ownerID, &c.sid, &c.name); err == nil {
			if !activeSIDs[c.sid] {
				list = append(list, c)
			}
		}
	}
	for _, item := range list {
		_ = logBroadcastSession(ctx, item.ownerID, item.sid, item.name, "CLOSE")
	}
}

// ListSessions restituisce lo storico delle sessioni di trasmissione (aperture/chiusure).
//
//encore:api auth method=GET path=/videochat/sessions
func ListSessions(ctx context.Context, p *ListSessionsParams) (*ListSessionsResponse, error) {
	uid, ok := auth.UserID()
	if !ok {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "must be logged in"}
	}

	limit := 50
	if p != nil && p.Limit > 0 {
		limit = p.Limit
	}
	if limit > 100 {
		limit = 100
	}

	ownerFilter := string(uid)
	// Se admin, può filtrare su un altro owner se specificato
	if err := requireAdmin(); err == nil && p != nil && p.OwnerID != "" {
		ownerFilter = p.OwnerID
	}

	rows, err := db.Query(ctx, `
		SELECT id, owner_id, sid, name, action, created_at
		FROM broadcast_sessions
		WHERE ($1 = '' OR owner_id = $1)
		ORDER BY created_at DESC, id DESC
		LIMIT $2
	`, ownerFilter, limit)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to list broadcast sessions")
	}
	defer rows.Close()

	var sessions []*BroadcastSession
	for rows.Next() {
		var s BroadcastSession
		if err := rows.Scan(&s.ID, &s.OwnerID, &s.SID, &s.Name, &s.Action, &s.CreatedAt); err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "failed to scan broadcast session")
		}
		sessions = append(sessions, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "error iterating broadcast sessions")
	}

	if sessions == nil {
		sessions = []*BroadcastSession{}
	}

	return &ListSessionsResponse{Sessions: sessions}, nil
}
