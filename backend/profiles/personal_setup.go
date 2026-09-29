package profiles

import (
	"context"
	"errors"
	"strings"
	"unicode"

	authsvc "encore.app/auth"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
)

// PersonalSetup holds the broadcasting configuration linked to a user profile.
type PersonalSetup struct {
	UserID       uuid.UUID `json:"user_id"`
	CanBroadcast bool      `json:"can_broadcast"`
	RoomName     string    `json:"room_name"`
}

type PersonalSetupParams struct {
	CanBroadcast bool   `json:"can_broadcast"`
	RoomName     string `json:"room_name"`
}

func (p *PersonalSetupParams) Validate() error {
	if p == nil {
		return &errs.Error{Code: errs.InvalidArgument, Message: "personal setup params are required"}
	}
	if !p.CanBroadcast {
		p.RoomName = ""
		return nil
	}
	p.RoomName = strings.TrimSpace(p.RoomName)
	if p.RoomName == "" || len(p.RoomName) > 128 || strings.ContainsAny(p.RoomName, "/\\") || strings.ContainsFunc(p.RoomName, unicode.IsControl) {
		return &errs.Error{Code: errs.InvalidArgument, Message: "broadcasting requires a room name of 1-128 bytes without slashes or control characters"}
	}
	return nil
}

// GetPersonalSetup returns the caller's setup, defaulting to broadcasting disabled.
//
//encore:api auth method=GET path=/profiles/personal-setup
func GetPersonalSetup(ctx context.Context) (*PersonalSetup, error) {
	userID, err := authedUserID()
	if err != nil {
		return nil, err
	}
	return getPersonalSetup(ctx, userID)
}

// GetUserPersonalSetup lets administrators read the selected user's setup.
//
//encore:api auth method=GET path=/profiles/personal-setup/:userID
func GetUserPersonalSetup(ctx context.Context, userID uuid.UUID) (*PersonalSetup, error) {
	if _, ok := auth.UserID(); !ok {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "authentication required"}
	}
	data, _ := auth.Data().(*authsvc.AuthData)
	if data == nil || !data.Role.IsAdmin() {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin role required"}
	}
	return getPersonalSetup(ctx, userID)
}

func getPersonalSetup(ctx context.Context, userID uuid.UUID) (*PersonalSetup, error) {
	setup := PersonalSetup{UserID: userID}
	err := db.QueryRow(ctx, `
		SELECT COALESCE(s.can_broadcast, FALSE), COALESCE(s.room_name, '')
		FROM user_profiles p
		LEFT JOIN personal_setup s ON s.user_id = p.user_id
		WHERE p.user_id = $1
	`, userID).Scan(&setup.CanBroadcast, &setup.RoomName)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{Code: errs.NotFound, Message: "profile not found"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to fetch personal setup")
	}
	return &setup, nil
}

// UpsertPersonalSetup lets administrators assign or revoke a user's broadcasting setup.
// This stores configuration only; it does not create or delete a LiveKit room.
//
//encore:api auth method=PUT path=/profiles/personal-setup/:userID
func UpsertPersonalSetup(ctx context.Context, userID uuid.UUID, p *PersonalSetupParams) (*PersonalSetup, error) {
	if _, ok := auth.UserID(); !ok {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "authentication required"}
	}
	data, _ := auth.Data().(*authsvc.AuthData)
	if data == nil || !data.Role.IsAdmin() {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin role required"}
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	setup := PersonalSetup{UserID: userID}
	err := db.QueryRow(ctx, `
		INSERT INTO personal_setup (user_id, can_broadcast, room_name)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE
		SET can_broadcast = EXCLUDED.can_broadcast,
		    room_name = EXCLUDED.room_name
		RETURNING can_broadcast, room_name
	`, userID, p.CanBroadcast, p.RoomName).Scan(&setup.CanBroadcast, &setup.RoomName)
	if isForeignKeyViolation(err) {
		return nil, &errs.Error{Code: errs.NotFound, Message: "profile not found"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to save personal setup")
	}
	return &setup, nil
}
