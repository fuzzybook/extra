package profiles

import (
	"context"
	"errors"
	"time"

	"encore.app/auth"
	"encore.dev/beta/errs"
	"encore.dev/pubsub"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
)

const loginTimeout = 10 * time.Second

type LoginParams struct {
	UserEmail string `json:"user_email"`
	Password  string `json:"password" encore:"sensitive"`
}

type LoginResponse struct {
	Token     uuid.UUID `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

var loginRequests = pubsub.TopicRef[pubsub.Publisher[*auth.LoginRequested]](auth.LoginRequests)

// Login is the public entry point for authenticating a user. It first
// resolves the given email to a user ID via the local profiles database.
// The actual credential check lives in the auth service; rather than calling
// it directly, this hands the request off over Pub/Sub (publishing to
// auth.LoginRequests) and waits for the matching auth.LoginCompleted reply
// on auth.LoginResults before responding to the caller.
//
//encore:api public method=POST path=/auth/login
func Login(ctx context.Context, p *LoginParams) (*LoginResponse, error) {
	var userID uuid.UUID
	err := db.QueryRow(ctx, `
		SELECT user_id FROM user_profiles WHERE email = $1
	`, p.UserEmail).Scan(&userID)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "invalid credentials"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to look up user")
	}

	requestID, err := uuid.NewV4()
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to generate request id")
	}

	if _, err := loginRequests.Publish(ctx, &auth.LoginRequested{
		RequestID: requestID,
		UserID:    userID,
		Password:  p.Password,
	}); err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to publish login request")
	}

	result, err := pollLoginResult(ctx, func(ctx context.Context) (*auth.LoginCompleted, error) {
		var ok bool
		var token uuid.UUID
		var expiresAt time.Time
		err := authDB.QueryRow(ctx, `
			SELECT ok, token, expires_at FROM login_results WHERE request_id = $1
		`, requestID).Scan(&ok, &token, &expiresAt)
		if errors.Is(err, sqldb.ErrNoRows) {
			return nil, sqldb.ErrNoRows
		}
		if err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "failed to read login result")
		}
		return &auth.LoginCompleted{RequestID: requestID, OK: ok, Token: token, ExpiresAt: expiresAt}, nil
	}, loginTimeout, 50*time.Millisecond)
	if err != nil {
		return nil, err
	}
	if !result.OK {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "invalid credentials"}
	}
	return &LoginResponse{Token: result.Token, ExpiresAt: result.ExpiresAt}, nil
}

func pollLoginResult(ctx context.Context, fetch func(context.Context) (*auth.LoginCompleted, error), timeout, interval time.Duration) (*auth.LoginCompleted, error) {
	deadline := time.Now().Add(timeout)
	for {
		result, err := fetch(ctx)
		if err == nil {
			return result, nil
		}
		if !errors.Is(err, sqldb.ErrNoRows) {
			return nil, err
		}
		if time.Now().After(deadline) {
			return nil, &errs.Error{Code: errs.DeadlineExceeded, Message: "login timed out"}
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}
	}
}
