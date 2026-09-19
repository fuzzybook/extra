package auth

import (
	"context"
	"errors"

	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
)

type RegisterParams struct {
	UserID   uuid.UUID `json:"user_id"`
	Password string    `json:"password" encore:"sensitive"`
}

func (p *RegisterParams) Validate() error {
	if p == nil {
		return &errs.Error{Code: errs.InvalidArgument, Message: "register params are required"}
	}
	if p.UserID == uuid.Nil {
		return &errs.Error{Code: errs.InvalidArgument, Message: "user id is required"}
	}
	return validatePassword(p.Password)
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

func (p *SetPasswordParams) Validate() error {
	if p == nil {
		return &errs.Error{Code: errs.InvalidArgument, Message: "password params are required"}
	}
	if err := validatePassword(p.CurrentPassword); err != nil {
		return err
	}
	return validatePassword(p.Password)
}

func (p *SetUserPasswordParams) Validate() error {
	if p == nil {
		return &errs.Error{Code: errs.InvalidArgument, Message: "user password params are required"}
	}
	return validatePassword(p.Password)
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
	userID, err := authenticatedUserID()
	if err != nil {
		return err
	}
	return setPasswordWithCurrentCheck(ctx, userID, p.CurrentPassword, p.Password)
}

// SetUserPassword changes a user's password from trusted internal flows.
//
//encore:api private method=PUT path=/auth/internal/users/:userID/password
func SetUserPassword(ctx context.Context, userID uuid.UUID, p *SetUserPasswordParams) error {
	if p.UserID != uuid.Nil && p.UserID != userID {
		return &errs.Error{Code: errs.InvalidArgument, Message: "user id mismatch"}
	}
	return replacePassword(ctx, userID, p.Password)
}

func setPasswordWithCurrentCheck(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to begin password update transaction")
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var currentHash string
	err = tx.QueryRow(ctx, `
		SELECT password_hash FROM credentials WHERE user_id = $1 FOR UPDATE
	`, userID).Scan(&currentHash)
	if errors.Is(err, sqldb.ErrNoRows) {
		return &errs.Error{Code: errs.NotFound, Message: "credentials not found"}
	}
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to fetch credentials")
	}
	if err = verifyPassword(currentHash, currentPassword); err != nil {
		return &errs.Error{Code: errs.Unauthenticated, Message: "invalid current password"}
	}
	if isPasswordReuse(currentHash, newPassword) {
		return &errs.Error{Code: errs.InvalidArgument, Message: "new password must be different from the current password"}
	}

	hash, err := hashPassword(newPassword)
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to hash password")
	}
	_, err = tx.Exec(ctx, `
		UPDATE credentials SET password_hash = $2, updated_at = NOW() WHERE user_id = $1
	`, userID, hash)
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to update password")
	}
	if err = tx.Commit(); err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to commit password update")
	}
	return nil
}

func replacePassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to begin password update transaction")
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	hash, err := hashPassword(newPassword)
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to hash password")
	}
	res, err := tx.Exec(ctx, `
		UPDATE credentials SET password_hash = $2, updated_at = NOW() WHERE user_id = $1
	`, userID, hash)
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to update password")
	}
	if res.RowsAffected() == 0 {
		return &errs.Error{Code: errs.NotFound, Message: "credentials not found"}
	}
	if err = tx.Commit(); err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to commit password update")
	}
	return nil
}
