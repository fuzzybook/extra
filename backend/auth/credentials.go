package auth

import (
	"context"
	"errors"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
	"golang.org/x/crypto/bcrypt"
)

type RegisterParams struct {
	UserID   uuid.UUID `json:"user_id"`
	Password string    `json:"password" encore:"sensitive"`
}

// Register stores the initial password hash for a newly created profile.
// Called directly (service-to-service) by profiles.Insert during registration,
// since that flow needs to know immediately whether credential setup succeeded.
//
//encore:api private method=POST path=/auth/credentials
func Register(ctx context.Context, p *RegisterParams) error {
	hash, err := hashPassword(p.Password)
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to hash password")
	}
	_, err = db.Exec(ctx, `
		INSERT INTO credentials (user_id, password_hash) VALUES ($1, $2)
	`, p.UserID, hash)
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to store credentials")
	}
	return nil
}

type SetPasswordParams struct {
	CurrentPassword string `json:"current_password" encore:"sensitive"`
	Password        string `json:"password" encore:"sensitive"`
}

type SetUserPasswordParams struct {
	UserID   uuid.UUID `json:"user_id"`
	Password string    `json:"password" encore:"sensitive"`
}

// SetPassword changes the password for the authenticated user.
//
//encore:api auth method=PUT path=/auth/password
func SetPassword(ctx context.Context, p *SetPasswordParams) error {
	uid, ok := auth.UserID()
	if !ok {
		return &errs.Error{Code: errs.Unauthenticated, Message: "missing auth"}
	}
	userID, err := uuid.FromString(string(uid))
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "invalid user id")
	}

	var currentHash string
	err = db.QueryRow(ctx, `
		SELECT password_hash FROM credentials WHERE user_id = $1
	`, userID).Scan(&currentHash)
	if errors.Is(err, sqldb.ErrNoRows) {
		return &errs.Error{Code: errs.NotFound, Message: "credentials not found"}
	}
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to fetch credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), pepperedPassword(p.CurrentPassword)); err != nil {
		return &errs.Error{Code: errs.Unauthenticated, Message: "invalid current password"}
	}

	hash, err := hashPassword(p.Password)
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to hash password")
	}
	_, err = db.Exec(ctx, `
		UPDATE credentials SET password_hash = $2, updated_at = NOW() WHERE user_id = $1
	`, userID, hash)
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to update password")
	}
	return nil
}

// SetUserPassword changes a user's password from trusted internal flows.
//
//encore:api private method=PUT path=/auth/internal/users/:userID/password
func SetUserPassword(ctx context.Context, userID uuid.UUID, p *SetUserPasswordParams) error {
	if p.UserID != uuid.Nil && p.UserID != userID {
		return &errs.Error{Code: errs.InvalidArgument, Message: "user id mismatch"}
	}

	hash, err := hashPassword(p.Password)
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to hash password")
	}

	res, err := db.Exec(ctx, `
		UPDATE credentials SET password_hash = $2, updated_at = NOW() WHERE user_id = $1
	`, userID, hash)
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to update password")
	}
	if res.RowsAffected() == 0 {
		return &errs.Error{Code: errs.NotFound, Message: "credentials not found"}
	}
	return nil
}
