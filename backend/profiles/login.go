package profiles

import (
	"context"
	"errors"
	"time"

	"encore.app/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
)

type LoginParams struct {
	UserEmail string `json:"user_email"`
	Password  string `json:"password" encore:"sensitive"`
}

type LoginResponse struct {
	Token     uuid.UUID `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Login is the public entry point for authenticating a user. It resolves the
// given email to a user ID via the local profiles database, rejects profiles
// that are not active, and then asks the auth service to check the password
// and issue a session.
//
//encore:api public method=POST path=/auth/login
func Login(ctx context.Context, p *LoginParams) (*LoginResponse, error) {
	email, err := normalizeEmailValue(p.UserEmail)
	if err != nil {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "invalid credentials"}
	}

	var userID uuid.UUID
	var status Status
	err = db.QueryRow(ctx, `
		SELECT user_id, status FROM user_profiles WHERE lower(email) = $1
	`, email).Scan(&userID, &status)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "invalid credentials"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to look up user")
	}

	// Suspended, banned or deleted profiles get the same answer as a wrong
	// password, so the endpoint does not reveal the status of an account.
	if status != StatusActive {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "invalid credentials"}
	}

	resp, err := auth.Login(ctx, &auth.LoginParams{UserID: userID, Password: p.Password})
	if err != nil {
		var e *errs.Error
		if errors.As(err, &e) && e.Code != errs.Unauthenticated && e.Code != errs.InvalidArgument {
			return nil, err
		}
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "invalid credentials"}
	}
	return &LoginResponse{Token: resp.Token, ExpiresAt: resp.ExpiresAt}, nil
}
