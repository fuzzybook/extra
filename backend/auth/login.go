package auth

import (
	"context"
	"errors"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/cron"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
)

type LoginParams struct {
	UserID   uuid.UUID `json:"user_id"`
	Password string    `json:"password" encore:"sensitive"`
}

func (p *LoginParams) Validate() error {
	if p == nil {
		return &errs.Error{Code: errs.InvalidArgument, Message: "login params are required"}
	}
	if p.UserID == uuid.Nil {
		return &errs.Error{Code: errs.InvalidArgument, Message: "user id is required"}
	}
	return validatePassword(p.Password)
}

type LoginResponse struct {
	Token     uuid.UUID `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Login verifies a user's password and issues a new session token.
//
// It is private — external clients authenticate via the public
// profiles.Login, which resolves the email to a user ID and calls this directly.
//
//encore:api private method=POST path=/auth/internal/login
func Login(ctx context.Context, p *LoginParams) (*LoginResponse, error) {
	var hash string
	err := db.QueryRow(ctx, `
		SELECT password_hash FROM credentials WHERE user_id = $1
	`, p.UserID).Scan(&hash)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "invalid credentials"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to fetch credentials")
	}
	if err := verifyPassword(hash, p.Password); err != nil {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "invalid credentials"}
	}

	token, err := uuid.NewV4()
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to generate session token")
	}
	expiresAt := time.Now().Add(sessionTTL)
	_, err = db.Exec(ctx, `
		INSERT INTO sessions (token, user_id, expires_at) VALUES ($1, $2, $3)
	`, token, p.UserID, expiresAt)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to create session")
	}
	return &LoginResponse{Token: token, ExpiresAt: expiresAt}, nil
}

var _ = cron.NewJob("cleanup-expired-sessions", cron.JobConfig{
	Title:    "Cleanup expired sessions",
	Every:    1 * cron.Hour,
	Endpoint: CleanupExpiredSessions,
})

// CleanupExpiredSessions removes expired sessions from the database.
//
//encore:api private
func CleanupExpiredSessions(ctx context.Context) error {
	_, err := db.Exec(ctx, `DELETE FROM sessions WHERE expires_at <= NOW()`)
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to purge expired sessions")
	}
	return nil
}

// Logout revokes the session used to authenticate the current request.
//
//encore:api auth method=POST path=/auth/logout
func Logout(ctx context.Context) error {
	data, _ := auth.Data().(*AuthData)
	if data == nil {
		return nil
	}
	_, err := db.Exec(ctx, `DELETE FROM sessions WHERE token = $1`, data.SessionToken)
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to revoke session")
	}
	return nil
}
