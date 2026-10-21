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
var broadcastersDB = sqldb.Named("broadcasters")

type Profile struct {
	UserID      uuid.UUID          `json:"user_id"`
	Email       string             `json:"email"`
	DisplayName string             `json:"display_name"`
	AvatarURL   string             `json:"avatar_url"`
	MemberType  string             `json:"member_type"`
	Broadcaster *uuid.UUID         `json:"broadcaster"`
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
		SELECT user_id, email, display_name, avatar_url, role, status, member_type, broadcaster, created_at, updated_at
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
		if err := rows.Scan(&p.UserID, &p.Email, &p.DisplayName, &p.AvatarURL, &p.Role, &p.Status, &p.MemberType, &p.Broadcaster, &p.CreatedAt, &p.UpdatedAt); err != nil {
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
		SELECT email, display_name, avatar_url, role, status, member_type, broadcaster, created_at, updated_at FROM user_profiles WHERE user_id = $1
	`, userID).Scan(&p.Email, &p.DisplayName, &p.AvatarURL, &p.Role, &p.Status, &p.MemberType, &p.Broadcaster, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{Code: errs.NotFound, Message: "profile not found"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to fetch profile")
	}
	return &p, nil
}

// InsertProfile creates a new user profile through profiles.Insert, so the
// same validation, email normalization and credential setup apply.
//
//encore:api auth method=POST path=/admin/profiles
func InsertProfile(ctx context.Context, p *RegisterParams) (*Profile, error) {
	created, err := profilessvc.Insert(ctx, &profilessvc.RegisterParams{
		Email:       p.Email,
		DisplayName: p.DisplayName,
		AvatarURL:   p.AvatarURL,
		Password:    p.Password,
	})
	if err != nil {
		return nil, err
	}
	return fromServiceProfile(created), nil
}

func fromServiceProfile(p *profilessvc.Profile) *Profile {
	return &Profile{
		UserID:      p.UserID,
		Email:       p.Email,
		DisplayName: p.DisplayName,
		AvatarURL:   p.AvatarURL,
		MemberType:  p.MemberType,
		Broadcaster: p.Broadcaster,
		Role:        p.Role,
		Status:      p.Status,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// UpdateProfile replaces the data of any user's profile.
//
//encore:api auth method=PUT path=/admin/profiles/:userID
func UpdateProfile(ctx context.Context, userID uuid.UUID, p *ProfileParams) (*Profile, error) {
	params := profilessvc.ProfileParams{DisplayName: p.DisplayName, AvatarURL: p.AvatarURL}
	if err := params.Validate(); err != nil {
		return nil, err
	}
	p.DisplayName, p.AvatarURL = params.DisplayName, params.AvatarURL
	profile := Profile{UserID: userID, DisplayName: p.DisplayName, AvatarURL: p.AvatarURL}
	err := profilesDB.QueryRow(ctx, `
		UPDATE user_profiles
		SET display_name = $2, avatar_url = $3, updated_at = NOW()
		WHERE user_id = $1
		RETURNING email, role, status, member_type, broadcaster, created_at, updated_at
	`, userID, p.DisplayName, p.AvatarURL).Scan(&profile.Email, &profile.Role, &profile.Status, &profile.MemberType, &profile.Broadcaster, &profile.CreatedAt, &profile.UpdatedAt)
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
		RETURNING email, display_name, avatar_url, role, member_type, broadcaster, created_at, updated_at
	`, userID, p.Status).Scan(&profile.Email, &profile.DisplayName, &profile.AvatarURL, &profile.Role, &profile.MemberType, &profile.Broadcaster, &profile.CreatedAt, &profile.UpdatedAt)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{Code: errs.NotFound, Message: "profile not found"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to update profile status")
	}
	if p.Status != profilessvc.StatusActive {
		if err := authsvc.RevokeUserSessions(ctx, userID); err != nil {
			return nil, err
		}
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
		RETURNING email, display_name, avatar_url, status, member_type, broadcaster, created_at, updated_at
	`, userID, p.Role).Scan(&profile.Email, &profile.DisplayName, &profile.AvatarURL, &profile.Status, &profile.MemberType, &profile.Broadcaster, &profile.CreatedAt, &profile.UpdatedAt)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{Code: errs.NotFound, Message: "profile not found"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to update profile role")
	}
	return &profile, nil
}

type UpdateProfilePasswordParams struct {
	Password string `json:"password" encore:"sensitive"`
}

// UpdateProfilePassword sets the password of any user's profile.
//
//encore:api auth method=PUT path=/admin/profiles/:userID/password
func UpdateProfilePassword(ctx context.Context, userID uuid.UUID, p *UpdateProfilePasswordParams) error {
	if err := authsvc.SetUserPassword(ctx, userID, &authsvc.SetUserPasswordParams{UserID: userID, Password: p.Password}); err != nil {
		return err
	}
	return nil
}

type UpdateProfileMemberTypeParams struct {
	MemberType string `json:"member_type"`
}

func (p *UpdateProfileMemberTypeParams) Validate() error {
	if !profilessvc.IsValidMemberType(p.MemberType) {
		return &errs.Error{Code: errs.InvalidArgument, Message: "invalid member type"}
	}
	return nil
}

// UpdateProfileMemberType sets the member type of any user's profile.
// If changing away from broadcaster, disables broadcaster and sets used=false.
// If changing to broadcaster, creates broadcaster with used=true and enabled=false.
//
//encore:api auth method=PUT path=/admin/profiles/:userID/member-type
func UpdateProfileMemberType(ctx context.Context, userID uuid.UUID, p *UpdateProfileMemberTypeParams) (*Profile, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	profile := Profile{UserID: userID}
	var (
		currentMemberType  string
		currentBroadcaster *uuid.UUID
	)

	err := profilesDB.QueryRow(ctx, `
		SELECT email, display_name, avatar_url, role, status, member_type, broadcaster, created_at, updated_at
		FROM user_profiles
		WHERE user_id = $1
	`, userID).Scan(
		&profile.Email,
		&profile.DisplayName,
		&profile.AvatarURL,
		&profile.Role,
		&profile.Status,
		&currentMemberType,
		&currentBroadcaster,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{Code: errs.NotFound, Message: "profile not found"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to fetch profile")
	}

	targetBroadcasterID := currentBroadcaster

	if currentMemberType == profilessvc.MemberTypeBroadcaster && p.MemberType != profilessvc.MemberTypeBroadcaster {
		// Moving AWAY from broadcaster:
		// "se è broadcaster e si vuole cambiarlo chiede conferma e se affermativa si disabilita e set used false al broadcaster"
		_, err := broadcastersDB.Exec(ctx, `
			UPDATE broadcasters
			SET enabled = FALSE, used = FALSE, updated_at = NOW()
			WHERE user_id = $1
		`, userID)
		if err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "failed to disable broadcaster")
		}
		targetBroadcasterID = nil
	} else if currentMemberType != profilessvc.MemberTypeBroadcaster && p.MemberType == profilessvc.MemberTypeBroadcaster {
		// Moving TO broadcaster:
		// "se invece il nuovo tipo si crea il broadcaster con used true enabled false"
		var existingID uuid.UUID
		err := broadcastersDB.QueryRow(ctx, `SELECT id FROM broadcasters WHERE user_id = $1`, userID).Scan(&existingID)
		if err == nil {
			_, err = broadcastersDB.Exec(ctx, `
				UPDATE broadcasters
				SET used = TRUE, enabled = FALSE, updated_at = NOW()
				WHERE id = $1
			`, existingID)
			if err != nil {
				return nil, errs.WrapCode(err, errs.Internal, "failed to update broadcaster status")
			}
			targetBroadcasterID = &existingID
		} else if errors.Is(err, sqldb.ErrNoRows) {
			newID, err := uuid.NewV4()
			if err != nil {
				return nil, errs.WrapCode(err, errs.Internal, "failed to generate broadcaster id")
			}
			now := time.Now()
			_, err = broadcastersDB.Exec(ctx, `
				INSERT INTO broadcasters (
					id, user_id, display_name, category, category_color, bio, certified,
					availability, phone, services, languages, attributes, hours, photos,
					enabled, used, created_at, updated_at
				) VALUES (
					$1, $2, $3, 'General', '#63f18eff', '', FALSE,
					'available', '', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb, '[]'::jsonb, '[]'::jsonb,
					FALSE, TRUE, $4, $4
				)
			`, newID, userID, profile.DisplayName, now)
			if err != nil {
				return nil, errs.WrapCode(err, errs.Internal, "failed to create broadcaster record")
			}
			targetBroadcasterID = &newID
		} else {
			return nil, errs.WrapCode(err, errs.Internal, "failed to check existing broadcaster")
		}
	}

	err = profilesDB.QueryRow(ctx, `
		UPDATE user_profiles
		SET member_type = $2, broadcaster = $3, updated_at = NOW()
		WHERE user_id = $1
		RETURNING updated_at
	`, userID, p.MemberType, targetBroadcasterID).Scan(&profile.UpdatedAt)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to update profile member type")
	}

	profile.MemberType = p.MemberType
	profile.Broadcaster = targetBroadcasterID
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
	return profilessvc.CleanupDeletedUser(ctx, userID)
}

type PersonalData struct {
	UserID    uuid.UUID `json:"user_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Address   string    `json:"address"`
	Cap       string    `json:"cap"`
	City      string    `json:"city"`
	Country   string    `json:"country"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PersonalDataParams struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Address   string `json:"address"`
	Cap       string `json:"cap"`
	City      string `json:"city"`
	Country   string `json:"country"`
}

// GetPersonalData returns the personal data for the given user.
//
//encore:api auth method=GET path=/admin/profiles/:userID/personal-data
func GetPersonalData(ctx context.Context, userID uuid.UUID) (*PersonalData, error) {
	pd := PersonalData{UserID: userID}
	err := profilesDB.QueryRow(ctx, `
		SELECT first_name, last_name, address, cap, city, country, created_at, updated_at
		FROM personal_data WHERE user_id = $1
	`, userID).Scan(&pd.FirstName, &pd.LastName, &pd.Address, &pd.Cap, &pd.City, &pd.Country, &pd.CreatedAt, &pd.UpdatedAt)
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
	params := profilessvc.PersonalDataParams(*p)
	if err := params.Validate(); err != nil {
		return nil, err
	}
	*p = PersonalDataParams(params)
	pd := PersonalData{
		UserID:    userID,
		FirstName: p.FirstName,
		LastName:  p.LastName,
		Address:   p.Address,
		Cap:       p.Cap,
		City:      p.City,
		Country:   p.Country,
	}
	err := profilesDB.QueryRow(ctx, `
		INSERT INTO personal_data (user_id, first_name, last_name, address, cap, city, country)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id) DO UPDATE
		SET first_name = EXCLUDED.first_name,
		    last_name = EXCLUDED.last_name,
		    address = EXCLUDED.address,
		    cap = EXCLUDED.cap,
		    city = EXCLUDED.city,
		    country = EXCLUDED.country,
		    updated_at = NOW()
		RETURNING created_at, updated_at
	`, userID, p.FirstName, p.LastName, p.Address, p.Cap, p.City, p.Country).Scan(&pd.CreatedAt, &pd.UpdatedAt)
	if isForeignKeyViolation(err) {
		return nil, &errs.Error{Code: errs.NotFound, Message: "profile not found"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to save personal data")
	}
	return &pd, nil
}

// isForeignKeyViolation reports whether err is a Postgres foreign key constraint violation.
func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
