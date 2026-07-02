// Service profiles stores and serves user profile information.
package profiles

import (
	"context"
	"errors"
	"time"

	authsvc "encore.app/auth"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

var db = sqldb.NewDatabase("profiles", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

type Profile struct {
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	IsArtist    bool      `json:"is_artist"`
	Role        string    `json:"role"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProfileParams struct {
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

type RegisterParams struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	Password    string `json:"password" encore:"sensitive"`
}

// Insert registers a new profile with a server-generated UUID v4, delegating
// credential setup to the auth service.
//
//encore:api public method=POST path=/profiles
func Insert(ctx context.Context, p *RegisterParams) (*Profile, error) {
	userID, err := uuid.NewV4()
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to generate profile id")
	}
	profile := Profile{UserID: userID, Email: p.Email, DisplayName: p.DisplayName, AvatarURL: p.AvatarURL}
	err = db.QueryRow(ctx, `
		INSERT INTO user_profiles (user_id, email, display_name, avatar_url)
		VALUES ($1, $2, $3, $4)
		RETURNING role, status, is_artist, created_at, updated_at
	`, userID, p.Email, p.DisplayName, p.AvatarURL).Scan(&profile.Role, &profile.Status, &profile.IsArtist, &profile.CreatedAt, &profile.UpdatedAt)
	if isUniqueViolation(err) {
		return nil, &errs.Error{Code: errs.AlreadyExists, Message: "email already registered"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to create profile")
	}
	if err := authsvc.Register(ctx, &authsvc.RegisterParams{UserID: userID, Password: p.Password}); err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to register credentials")
	}
	return &profile, nil
}

// Update replaces the data of the authenticated user's own profile.
//
//encore:api auth method=PUT path=/profiles
func Update(ctx context.Context, p *ProfileParams) (*Profile, error) {
	uid, ok := auth.UserID()
	if !ok {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "cannot act on another user's profile"}
	}
	userID, err := uuid.FromString(string(uid))
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "invalid user id")
	}
	if err := requireSelf(userID); err != nil {
		return nil, err
	}
	profile := Profile{UserID: userID, DisplayName: p.DisplayName, AvatarURL: p.AvatarURL}
	err = db.QueryRow(ctx, `
		UPDATE user_profiles
		SET display_name = $2, avatar_url = $3, updated_at = NOW()
		WHERE user_id = $1
		RETURNING email, role, status, is_artist, created_at, updated_at
	`, userID, p.DisplayName, p.AvatarURL).Scan(&profile.Email, &profile.Role, &profile.Status, &profile.IsArtist, &profile.CreatedAt, &profile.UpdatedAt)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{Code: errs.NotFound, Message: "profile not found"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to update profile")
	}
	return &profile, nil
}

// Get returns the profile for the given user. For the caller's own profile
// (including when not authenticated), see MyProfile.
//
//encore:api public method=GET path=/profiles/profile
func Get(ctx context.Context) (*Profile, error) {
	uid, ok := auth.UserID()
	if !ok {
		return nil, nil
	}
	userID, err := uuid.FromString(string(uid))
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "invalid user id")
	}
	p := Profile{UserID: userID}
	err = db.QueryRow(ctx, `
		SELECT email, display_name, avatar_url, role, status, is_artist, created_at, updated_at FROM user_profiles WHERE user_id = $1
	`, userID).Scan(&p.Email, &p.DisplayName, &p.AvatarURL, &p.Role, &p.Status, &p.IsArtist, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{Code: errs.NotFound, Message: "profile not found"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to fetch profile")
	}
	return &p, nil
}

// Delete removes the authenticated user's own profile.
//
//encore:api auth method=DELETE path=/profiles
func Delete(ctx context.Context) error {
	uid, ok := auth.UserID()
	if !ok {
		return &errs.Error{Code: errs.PermissionDenied, Message: "cannot act on another user's profile"}
	}
	userID, err := uuid.FromString(string(uid))
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "invalid user id")
	}
	if err := requireSelf(userID); err != nil {
		return err
	}
	res, err := db.Exec(ctx, `
		DELETE FROM user_profiles WHERE user_id = $1
	`, userID)
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to delete profile")
	}
	if res.RowsAffected() == 0 {
		return &errs.Error{Code: errs.NotFound, Message: "profile not found"}
	}
	return nil
}

// requireSelf returns an error unless the authenticated caller is the user identified by userID.
func requireSelf(userID uuid.UUID) error {
	if uid, ok := auth.UserID(); !ok || uid != auth.UID(userID.String()) {
		return &errs.Error{Code: errs.PermissionDenied, Message: "cannot act on another user's profile"}
	}
	return nil
}

// isUniqueViolation reports whether err is a Postgres unique constraint violation.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
