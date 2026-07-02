package auth

import (
	"context"
	"errors"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/pubsub"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
	"golang.org/x/crypto/bcrypt"
)

type LoginParams struct {
	UserID   uuid.UUID `json:"user_id"`
	Password string    `json:"password" encore:"sensitive"`
}

type LoginResponse struct {
	Token     uuid.UUID `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Login verifies a user's password and issues a new session token.
//
// It is private — external clients authenticate via the public
// profiles.Login, which talks to this service asynchronously over Pub/Sub
// (see LoginRequests/LoginResults below) rather than calling it directly.
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
	if err := bcrypt.CompareHashAndPassword([]byte(hash), pepperedPassword(p.Password)); err != nil {
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

// LoginRequested is published by profiles.Login to ask this service to
// authenticate a user asynchronously — the public endpoint lives in profiles,
// but the credential check and session issuance happen here.
type LoginRequested struct {
	RequestID uuid.UUID
	UserID    uuid.UUID
	Password  string `encore:"sensitive"`
}

var LoginRequests = pubsub.NewTopic[*LoginRequested]("login-requests", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

// LoginCompleted carries the outcome of a LoginRequested message back to
// whichever profiles instance is waiting on the matching RequestID.
type LoginCompleted struct {
	RequestID uuid.UUID
	OK        bool
	Token     uuid.UUID
	ExpiresAt time.Time
}

var LoginResults = pubsub.NewTopic[*LoginCompleted]("login-results", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

var _ = pubsub.NewSubscription(
	LoginRequests, "authenticate",
	pubsub.SubscriptionConfig[*LoginRequested]{
		Handler: handleLoginRequested,
	},
)

func handleLoginRequested(ctx context.Context, ev *LoginRequested) error {
	result := &LoginCompleted{RequestID: ev.RequestID}
	if resp, err := Login(ctx, &LoginParams{UserID: ev.UserID, Password: ev.Password}); err == nil {
		result.OK = true
		result.Token = resp.Token
		result.ExpiresAt = resp.ExpiresAt
	}
	_, err := LoginResults.Publish(ctx, result)
	return err
}
