// Service admin provides administrative endpoints for managing application data.
package admin

import (
	"context"
	"errors"
	"time"

	authsvc "encore.app/auth"
	profilessvc "encore.app/profiles"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

var profilesDB = sqldb.Named("profiles")

type Profile struct {
	UserID      uuid.UUID          `json:"user_id"`
	Email       string             `json:"email"`
	DisplayName string             `json:"display_name"`
	AvatarURL   string             `json:"avatar_url"`
	IsArtist    bool               `json:"is_artist"`
	Role        string             `json:"role"`
	Status      profilessvc.Status `json:"status"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
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

type ListProfilesResponse struct {
	Profiles []*Profile `json:"profiles"`
}

// ListProfiles returns all user profiles, ordered by user ID.
//
//encore:api auth method=GET path=/admin/profiles
func ListProfiles(ctx context.Context) (*ListProfilesResponse, error) {
	rows, err := profilesDB.Query(ctx, `
		SELECT user_id, email, display_name, avatar_url, role, status, is_artist, created_at, updated_at
		FROM user_profiles
		ORDER BY user_id
	`)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to list profiles")
	}
	defer rows.Close()

	profiles := []*Profile{}
	for rows.Next() {
		var p Profile
		if err := rows.Scan(&p.UserID, &p.Email, &p.DisplayName, &p.AvatarURL, &p.Role, &p.Status, &p.IsArtist, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "failed to scan profile")
		}
		profiles = append(profiles, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to list profiles")
	}
	return &ListProfilesResponse{Profiles: profiles}, nil
}

// GetProfile returns the profile for the given user.
//
//encore:api auth method=GET path=/admin/profiles/:userID
func GetProfile(ctx context.Context, userID uuid.UUID) (*Profile, error) {
	p := Profile{UserID: userID}
	err := profilesDB.QueryRow(ctx, `
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

// InsertProfile creates a new user profile with a server-generated UUID v4,
// delegating credential setup to the auth service.
//
//encore:api auth method=POST path=/admin/profiles
func InsertProfile(ctx context.Context, p *RegisterParams) (*Profile, error) {
	userID, err := uuid.NewV4()
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to generate profile id")
	}
	profile := Profile{UserID: userID, Email: p.Email, DisplayName: p.DisplayName, AvatarURL: p.AvatarURL}
	err = profilesDB.QueryRow(ctx, `
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

// UpdateProfile replaces the data of any user's profile.
//
//encore:api auth method=PUT path=/admin/profiles/:userID
func UpdateProfile(ctx context.Context, userID uuid.UUID, p *ProfileParams) (*Profile, error) {
	profile := Profile{UserID: userID, DisplayName: p.DisplayName, AvatarURL: p.AvatarURL}
	err := profilesDB.QueryRow(ctx, `
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

type UpdateProfileStatusParams struct {
	Status profilessvc.Status `json:"status"`
}

func (p *UpdateProfileStatusParams) Validate() error {
	if !p.Status.IsValid() {
		return &errs.Error{Code: errs.InvalidArgument, Message: "invalid status"}
	}
	return nil
}

// UpdateProfileStatus sets the status of any user's profile.
//
//encore:api auth method=PUT path=/admin/profiles/:userID/status
func UpdateProfileStatus(ctx context.Context, userID uuid.UUID, p *UpdateProfileStatusParams) (*Profile, error) {
	profile := Profile{UserID: userID, Status: p.Status}
	err := profilesDB.QueryRow(ctx, `
		UPDATE user_profiles
		SET status = $2, updated_at = NOW()
		WHERE user_id = $1
		RETURNING email, display_name, avatar_url, role, is_artist, created_at, updated_at
	`, userID, p.Status).Scan(&profile.Email, &profile.DisplayName, &profile.AvatarURL, &profile.Role, &profile.IsArtist, &profile.CreatedAt, &profile.UpdatedAt)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{Code: errs.NotFound, Message: "profile not found"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to update profile status")
	}
	return &profile, nil
}

type UpdateProfileRoleParams struct {
	Role string `json:"role"`
}

func (p *UpdateProfileRoleParams) Validate() error {
	if !authsvc.IsValidRole(p.Role) {
		return &errs.Error{Code: errs.InvalidArgument, Message: "invalid role"}
	}
	return nil
}

// UpdateProfileRole sets the role of any user's profile.
//
//encore:api auth method=PUT path=/admin/profiles/:userID/role
func UpdateProfileRole(ctx context.Context, userID uuid.UUID, p *UpdateProfileRoleParams) (*Profile, error) {
	profile := Profile{UserID: userID, Role: p.Role}
	err := profilesDB.QueryRow(ctx, `
		UPDATE user_profiles
		SET role = $2, updated_at = NOW()
		WHERE user_id = $1
		RETURNING email, display_name, avatar_url, status, is_artist, created_at, updated_at
	`, userID, p.Role).Scan(&profile.Email, &profile.DisplayName, &profile.AvatarURL, &profile.Status, &profile.IsArtist, &profile.CreatedAt, &profile.UpdatedAt)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{Code: errs.NotFound, Message: "profile not found"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to update profile role")
	}
	return &profile, nil
}

type UpdateProfileArtistParams struct {
	IsArtist bool `json:"is_artist"`
}

// UpdateProfileArtist sets the artist flag of any user's profile.
//
//encore:api auth method=PUT path=/admin/profiles/:userID/artist
func UpdateProfileArtist(ctx context.Context, userID uuid.UUID, p *UpdateProfileArtistParams) (*Profile, error) {
	profile := Profile{UserID: userID, IsArtist: p.IsArtist}
	err := profilesDB.QueryRow(ctx, `
		UPDATE user_profiles
		SET is_artist = $2, updated_at = NOW()
		WHERE user_id = $1
		RETURNING email, display_name, avatar_url, role, status, created_at, updated_at
	`, userID, p.IsArtist).Scan(&profile.Email, &profile.DisplayName, &profile.AvatarURL, &profile.Role, &profile.Status, &profile.CreatedAt, &profile.UpdatedAt)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{Code: errs.NotFound, Message: "profile not found"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to update profile artist flag")
	}
	return &profile, nil
}

// DeleteProfile removes any user's profile.
//
//encore:api auth method=DELETE path=/admin/profiles/:userID
func DeleteProfile(ctx context.Context, userID uuid.UUID) error {
	res, err := profilesDB.Exec(ctx, `
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

type PersonalData struct {
	UserID    uuid.UUID `json:"user_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Address   string    `json:"address"`
	City      string    `json:"city"`
	Country   string    `json:"country"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PersonalDataParams struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Address   string `json:"address"`
	City      string `json:"city"`
	Country   string `json:"country"`
}

// GetPersonalData returns the personal data for the given user.
//
//encore:api auth method=GET path=/admin/profiles/:userID/personal-data
func GetPersonalData(ctx context.Context, userID uuid.UUID) (*PersonalData, error) {
	pd := PersonalData{UserID: userID}
	err := profilesDB.QueryRow(ctx, `
		SELECT first_name, last_name, address, city, country, created_at, updated_at
		FROM personal_data WHERE user_id = $1
	`, userID).Scan(&pd.FirstName, &pd.LastName, &pd.Address, &pd.City, &pd.Country, &pd.CreatedAt, &pd.UpdatedAt)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{Code: errs.NotFound, Message: "personal data not found"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to fetch personal data")
	}
	return &pd, nil
}

// UpsertPersonalData creates or replaces the personal data for the given user.
//
//encore:api auth method=PUT path=/admin/profiles/:userID/personal-data
func UpsertPersonalData(ctx context.Context, userID uuid.UUID, p *PersonalDataParams) (*PersonalData, error) {
	pd := PersonalData{
		UserID:    userID,
		FirstName: p.FirstName,
		LastName:  p.LastName,
		Address:   p.Address,
		City:      p.City,
		Country:   p.Country,
	}
	err := profilesDB.QueryRow(ctx, `
		INSERT INTO personal_data (user_id, first_name, last_name, address, city, country)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) DO UPDATE
		SET first_name = EXCLUDED.first_name,
		    last_name = EXCLUDED.last_name,
		    address = EXCLUDED.address,
		    city = EXCLUDED.city,
		    country = EXCLUDED.country,
		    updated_at = NOW()
		RETURNING created_at, updated_at
	`, userID, p.FirstName, p.LastName, p.Address, p.City, p.Country).Scan(&pd.CreatedAt, &pd.UpdatedAt)
	if isForeignKeyViolation(err) {
		return nil, &errs.Error{Code: errs.NotFound, Message: "profile not found"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to save personal data")
	}
	return &pd, nil
}

// isUniqueViolation reports whether err is a Postgres unique constraint violation.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// isForeignKeyViolation reports whether err is a Postgres foreign key constraint violation.
func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
