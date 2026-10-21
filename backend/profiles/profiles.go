// Service profiles stores and serves user profile information.
package profiles

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	authsvc "encore.app/auth"
	broadcasterssvc "encore.app/broadcasters"
	"encore.app/validate"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

var db = sqldb.NewDatabase("profiles", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

const (
	MemberTypeBroadcaster = "broadcaster"
	MemberTypeVisitor     = "visitor"
	MemberTypeAffiliate   = "affiliate"
	MemberTypeAdvertiser  = "advertiser"
	MemberTypeAdmin       = "admin"
)

type MemberTypeOption struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func MemberTypes() []MemberTypeOption {
	return []MemberTypeOption{
		{Name: "Broadcaster", Value: MemberTypeBroadcaster},
		{Name: "Visitor", Value: MemberTypeVisitor},
		{Name: "Affiliate", Value: MemberTypeAffiliate},
		{Name: "Advertiser", Value: MemberTypeAdvertiser},
		{Name: "Admin", Value: MemberTypeAdmin},
	}
}

func IsValidMemberType(value string) bool {
	switch value {
	case MemberTypeBroadcaster, MemberTypeVisitor, MemberTypeAffiliate, MemberTypeAdvertiser, MemberTypeAdmin:
		return true
	default:
		return false
	}
}

type Profile struct {
	UserID      uuid.UUID  `json:"user_id"`
	Email       string     `json:"email"`
	DisplayName string     `json:"display_name"`
	AvatarURL   string     `json:"avatar_url"`
	MemberType  string     `json:"member_type"`
	Broadcaster *uuid.UUID `json:"broadcaster"`
	Role        string     `json:"role"`
	Status      Status     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
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

func (p *RegisterParams) Validate() error {
	if p == nil {
		return &errs.Error{Code: errs.InvalidArgument, Message: "register params are required"}
	}
	email, err := normalizeEmailValue(p.Email)
	if err != nil {
		return err
	}
	p.Email = email
	name := normalizeText(p.DisplayName)
	if name == "" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "display name is required"}
	}
	p.DisplayName = name
	if p.AvatarURL != "" {
		if err := validate.HTTPURL(p.AvatarURL); err != nil {
			return &errs.Error{Code: errs.InvalidArgument, Message: "avatar url is invalid"}
		}
	}
	return authsvc.ValidatePassword(p.Password)
}

func (p *ProfileParams) Validate() error {
	if p == nil {
		return &errs.Error{Code: errs.InvalidArgument, Message: "profile params are required"}
	}
	name := normalizeText(p.DisplayName)
	if name == "" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "display name is required"}
	}
	if p.AvatarURL != "" {
		if err := validate.HTTPURL(p.AvatarURL); err != nil {
			return &errs.Error{Code: errs.InvalidArgument, Message: "avatar url is invalid"}
		}
	}
	p.DisplayName = name
	return nil
}

func normalizeEmailValue(email string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return "", &errs.Error{Code: errs.InvalidArgument, Message: "email is required"}
	}
	parsed, err := mail.ParseAddress(email)
	if err != nil {
		return "", &errs.Error{Code: errs.InvalidArgument, Message: "invalid email"}
	}
	return strings.TrimSpace(strings.ToLower(parsed.Address)), nil
}

func normalizeText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	parts := strings.Fields(value)
	return strings.Join(parts, " ")
}

// Insert registers a new profile with a server-generated UUID v4, delegating
// credential setup to the auth service. It is private: the public sign-up
// flow is registration.Register (validation, welcome email), and
// administrators go through admin.InsertProfile.
//
//encore:api private method=POST path=/profiles
func Insert(ctx context.Context, p *RegisterParams) (*Profile, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	userID, err := uuid.NewV4()
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to generate profile id")
	}
	profile := Profile{UserID: userID, Email: p.Email, DisplayName: p.DisplayName, AvatarURL: p.AvatarURL}
	err = db.QueryRow(ctx, `
		INSERT INTO user_profiles (user_id, email, display_name, avatar_url)
		VALUES ($1, $2, $3, $4)
		RETURNING role, status, member_type, broadcaster, created_at, updated_at
	`, userID, p.Email, p.DisplayName, p.AvatarURL).Scan(&profile.Role, &profile.Status, &profile.MemberType, &profile.Broadcaster, &profile.CreatedAt, &profile.UpdatedAt)
	if isUniqueViolation(err) {
		return nil, &errs.Error{Code: errs.AlreadyExists, Message: "email already registered"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to create profile")
	}
	if err := authsvc.Register(ctx, &authsvc.RegisterParams{UserID: userID, Password: p.Password}); err != nil {
		// Without credentials the profile is unusable and would block the email.
		_, _ = db.Exec(ctx, `DELETE FROM user_profiles WHERE user_id = $1`, userID)
		return nil, errs.WrapCode(err, errs.Internal, "failed to register credentials")
	}
	return &profile, nil
}

// Update replaces the data of the authenticated user's own profile.
//
//encore:api auth method=PUT path=/profiles
func Update(ctx context.Context, p *ProfileParams) (*Profile, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
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
	return CleanupDeletedUser(ctx, userID)
}

// CleanupDeletedUser removes what other services keep for a deleted profile:
// credentials and sessions (auth) and the broadcaster record. Called by
// Delete and by admin.DeleteProfile.
//
//encore:api private method=POST path=/profiles/internal/:userID/cleanup
func CleanupDeletedUser(ctx context.Context, userID uuid.UUID) error {
	if err := authsvc.DeleteCredentials(ctx, userID); err != nil {
		return err
	}
	return broadcasterssvc.DeleteByUser(ctx, userID)
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
