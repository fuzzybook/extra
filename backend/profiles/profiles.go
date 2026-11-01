// Service profiles stores and serves user profile information.
package profiles

import (
	"context"
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

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
	MemberTypeClub        = "club"
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
		{Name: "Club", Value: MemberTypeClub},
		{Name: "Admin", Value: MemberTypeAdmin},
	}
}

func IsValidMemberType(value string) bool {
	switch value {
	case MemberTypeBroadcaster, MemberTypeVisitor, MemberTypeAffiliate, MemberTypeAdvertiser, MemberTypeClub, MemberTypeAdmin:
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

	// Dati del modulo di registrazione pubblico, facoltativi qui (gli admin
	// creano profili senza): registration.Register li rende obbligatori.
	Language            string          `json:"language,omitempty" encore:"optional"`
	RequestedMemberType string          `json:"requested_member_type,omitempty" encore:"optional"`
	Newsletter          bool            `json:"newsletter,omitempty" encore:"optional"`
	TermsAccepted       bool            `json:"terms_accepted,omitempty" encore:"optional"`
	Personal            *SignupPersonal `json:"personal,omitempty" encore:"optional"`
}

// SignupPersonal sono i dati privati raccolti alla registrazione: il nome è
// obbligatorio, gli altri campi facoltativi.
type SignupPersonal struct {
	Name    string `json:"name"`
	Phone   string `json:"phone,omitempty" encore:"optional"`
	Country string `json:"country,omitempty" encore:"optional"`
	Address string `json:"address,omitempty" encore:"optional"`
}

// Languages sono le lingue del sito; la prima è quella predefinita.
var Languages = []string{"en", "it", "fr", "de"}

// SignupMemberTypes sono i tipi di membro che si possono richiedere alla
// registrazione (non admin).
var SignupMemberTypes = []string{MemberTypeVisitor, MemberTypeBroadcaster, MemberTypeAffiliate, MemberTypeAdvertiser, MemberTypeClub}

// contactRequiredMemberTypes devono indicare telefono, paese e indirizzo.
var contactRequiredMemberTypes = []string{MemberTypeAdvertiser, MemberTypeClub}

// signupCountries sono i paesi accettati alla registrazione: gli stati
// europei (ISO 3166-1 alpha-2, più XK per il Kosovo).
var signupCountries = strings.Fields(`AD AL AT BA BE BG BY CH CY CZ DE DK EE ES FI FR GB GR HR HU IE IS IT LI LT LU
	LV MC MD ME MK MT NL NO PL PT RO RS RU SE SI SK SM UA VA XK`)

var (
	phonePattern   = regexp.MustCompile(`^\+?[0-9 ()./-]{5,30}$`)
	countryPattern = regexp.MustCompile(`^[A-Z]{2}$`)
)

const (
	maxNameLength    = 120
	maxAddressLength = 500
)

func contains(values []string, v string) bool {
	for _, x := range values {
		if x == v {
			return true
		}
	}
	return false
}

// initialMemberType è il tipo di membro assegnato alla registrazione: quello
// richiesto, tranne broadcaster che dà subito accesso alle stanze e quindi
// resta una richiesta da confermare da un admin.
func initialMemberType(requested string) string {
	if requested == "" || requested == MemberTypeBroadcaster {
		return MemberTypeVisitor
	}
	return requested
}

func (p *SignupPersonal) validate() error {
	p.Name = normalizeText(p.Name)
	p.Phone = normalizeText(p.Phone)
	p.Country = strings.ToUpper(strings.TrimSpace(p.Country))
	p.Address = strings.TrimSpace(p.Address)
	if p.Name == "" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "name is required"}
	}
	if utf8.RuneCountInString(p.Name) > maxNameLength {
		return &errs.Error{Code: errs.InvalidArgument, Message: "name is too long"}
	}
	if p.Phone != "" && !phonePattern.MatchString(p.Phone) {
		return &errs.Error{Code: errs.InvalidArgument, Message: "phone number is invalid"}
	}
	if p.Country != "" && !countryPattern.MatchString(p.Country) {
		return &errs.Error{Code: errs.InvalidArgument, Message: "country must be a 2-letter code"}
	}
	if p.Country != "" && !contains(signupCountries, p.Country) {
		return &errs.Error{Code: errs.InvalidArgument, Message: "country must be a European state"}
	}
	if utf8.RuneCountInString(p.Address) > maxAddressLength {
		return &errs.Error{Code: errs.InvalidArgument, Message: "address is too long"}
	}
	return nil
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
	p.Language = strings.ToLower(strings.TrimSpace(p.Language))
	if p.Language != "" && !contains(Languages, p.Language) {
		return &errs.Error{Code: errs.InvalidArgument, Message: "language is not supported"}
	}
	p.RequestedMemberType = strings.TrimSpace(p.RequestedMemberType)
	if p.RequestedMemberType != "" && !contains(SignupMemberTypes, p.RequestedMemberType) {
		return &errs.Error{Code: errs.InvalidArgument, Message: "member type is not valid"}
	}
	if p.Personal != nil {
		if err := p.Personal.validate(); err != nil {
			return err
		}
	}
	if contains(contactRequiredMemberTypes, p.RequestedMemberType) {
		if p.Personal == nil || p.Personal.Phone == "" || p.Personal.Country == "" || p.Personal.Address == "" {
			return &errs.Error{Code: errs.InvalidArgument, Message: "phone, country and address are required for advertisers and clubs"}
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
	language := p.Language
	if language == "" {
		language = Languages[0]
	}
	var termsAcceptedAt *time.Time
	if p.TermsAccepted {
		now := time.Now()
		termsAcceptedAt = &now
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to create profile")
	}
	defer func() { _ = tx.Rollback() }()
	err = tx.QueryRow(ctx, `
		INSERT INTO user_profiles (
			user_id, email, display_name, avatar_url,
			member_type, requested_member_type, language, newsletter, terms_accepted_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING role, status, member_type, broadcaster, created_at, updated_at
	`, userID, p.Email, p.DisplayName, p.AvatarURL,
		initialMemberType(p.RequestedMemberType), p.RequestedMemberType, language, p.Newsletter, termsAcceptedAt,
	).Scan(&profile.Role, &profile.Status, &profile.MemberType, &profile.Broadcaster, &profile.CreatedAt, &profile.UpdatedAt)
	if isUniqueViolation(err) {
		return nil, &errs.Error{Code: errs.AlreadyExists, Message: "email already registered"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to create profile")
	}
	if pd := p.Personal; pd != nil {
		if _, err := tx.Exec(ctx, `
			INSERT INTO personal_data (user_id, first_name, phone, country, address)
			VALUES ($1, $2, $3, $4, $5)
		`, userID, pd.Name, pd.Phone, pd.Country, pd.Address); err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "failed to save personal data")
		}
	}
	if err := tx.Commit(); err != nil {
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
