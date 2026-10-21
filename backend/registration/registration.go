// Service registration orchestrates new user registration.
package registration

import (
	"bytes"
	"context"
	"embed"
	"errors"
	htmltemplate "html/template"
	"net"
	"net/mail"
	"os"
	"strings"
	texttemplate "text/template"
	"time"

	authsvc "encore.app/auth"
	mailssvc "encore.app/mails"
	profilessvc "encore.app/profiles"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
)

//go:embed mails/*.tmpl
var mailTemplates embed.FS

var db = sqldb.NewDatabase("registration", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

var profilesDB = sqldb.Named("profiles")

const (
	welcomeTokenTTL        = 24 * time.Hour
	passwordResetTokenTTL  = time.Hour
	defaultFrontendBaseURL = "http://localhost:9000/#"
)

var frontendBaseURL = defaultFrontendBaseURL

func init() {
	if value := strings.TrimSpace(os.Getenv("FRONTEND_BASE_URL")); value != "" {
		frontendBaseURL = value
	}
}

type RegisterParams struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	Password    string `json:"password" encore:"sensitive"`
}

type RegisterResponse struct {
	Profile *profilessvc.Profile `json:"profile"`
}

type WelcomeResponse struct {
	Email     string    `json:"email"`
	Confirmed bool      `json:"confirmed"`
	UsedAt    time.Time `json:"used_at"`
}

type PasswordResetRequestParams struct {
	Email string `json:"email"`
}

func (p *PasswordResetRequestParams) Validate() error {
	if p == nil {
		return &errs.Error{Code: errs.InvalidArgument, Message: "password reset request is required"}
	}
	_, err := normalizeEmail(p.Email)
	return err
}

type PasswordResetRequestResponse struct {
	EmailSent bool `json:"email_sent"`
}

type PasswordResetConfirmParams struct {
	Password string `json:"password" encore:"sensitive"`
}

func (p *PasswordResetConfirmParams) Validate() error {
	if p == nil {
		return &errs.Error{Code: errs.InvalidArgument, Message: "password reset confirm params are required"}
	}
	return authsvc.ValidatePassword(p.Password)
}

type PasswordResetConfirmResponse struct {
	Email string `json:"email"`
	Reset bool   `json:"reset"`
}

type EmailAvailabilityResponse struct {
	Email     string `json:"email"`
	Available bool   `json:"available"`
	MXValid   bool   `json:"mx_valid"`
}

// CheckEmail reports whether an email can be used for a new registration.
//
//encore:api public method=GET path=/registration/email/:email
func CheckEmail(ctx context.Context, email string) (*EmailAvailabilityResponse, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return nil, err
	}

	domain, ok := emailDomain(email)
	if !ok {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "invalid email domain"}
	}
	mxValid, err := hasValidMX(domain)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to check email domain")
	}

	var exists bool
	err = profilesDB.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM user_profiles WHERE lower(email) = $1)
	`, email).Scan(&exists)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to check email")
	}

	return &EmailAvailabilityResponse{Email: email, Available: !exists, MXValid: mxValid}, nil
}

// Register creates a new user and records the welcome transactional email.
//
//encore:api public method=POST path=/registration
func Register(ctx context.Context, p *RegisterParams) (*RegisterResponse, error) {
	if p == nil {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "register params are required"}
	}
	if err := (&profilessvc.RegisterParams{
		Email:       p.Email,
		DisplayName: p.DisplayName,
		AvatarURL:   p.AvatarURL,
		Password:    p.Password,
	}).Validate(); err != nil {
		return nil, err
	}
	pruneExpiredTokens(ctx)
	profile, err := profilessvc.Insert(ctx, &profilessvc.RegisterParams{
		Email:       p.Email,
		DisplayName: p.DisplayName,
		AvatarURL:   p.AvatarURL,
		Password:    p.Password,
	})
	if err != nil {
		return nil, err
	}

	token, err := createWelcomeToken(ctx, profile)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to create welcome token")
	}

	mail, err := welcomeMail(profile, token)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to render welcome email")
	}
	if _, err := mailssvc.SendTransactional(ctx, mail); err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to queue welcome email")
	}

	return &RegisterResponse{Profile: profile}, nil
}

// ConfirmWelcome validates the welcome email token.
//
//encore:api public method=POST path=/registration/welcome/:token
func ConfirmWelcome(ctx context.Context, token uuid.UUID) (*WelcomeResponse, error) {
	var email string
	var expiresAt time.Time
	var usedAt *time.Time
	err := db.QueryRow(ctx, `
		SELECT email, expires_at, used_at FROM welcome_tokens WHERE token = $1
	`, token).Scan(&email, &expiresAt, &usedAt)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{Code: errs.NotFound, Message: "welcome token not found"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to load welcome token")
	}
	if usedAt != nil {
		return &WelcomeResponse{Email: email, Confirmed: true, UsedAt: *usedAt}, nil
	}
	if time.Now().After(expiresAt) {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "welcome token has expired"}
	}

	var confirmedAt time.Time
	err = db.QueryRow(ctx, `
		UPDATE welcome_tokens SET used_at = NOW() WHERE token = $1
		RETURNING used_at
	`, token).Scan(&confirmedAt)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to confirm welcome token")
	}
	return &WelcomeResponse{Email: email, Confirmed: true, UsedAt: confirmedAt}, nil
}

// RequestPasswordReset creates a reset token and queues an email when the account exists.
//
//encore:api public method=POST path=/registration/password-reset
func RequestPasswordReset(ctx context.Context, p *PasswordResetRequestParams) (*PasswordResetRequestResponse, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	email, err := normalizeEmail(p.Email)
	if err != nil {
		return nil, err
	}
	pruneExpiredTokens(ctx)

	var profile profilessvc.Profile
	err = profilesDB.QueryRow(ctx, `
		SELECT user_id, email, display_name, avatar_url, role, status, member_type, created_at, updated_at
		FROM user_profiles
		WHERE lower(email) = $1
	`, email).Scan(
		&profile.UserID,
		&profile.Email,
		&profile.DisplayName,
		&profile.AvatarURL,
		&profile.Role,
		&profile.Status,
		&profile.MemberType,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if errors.Is(err, sqldb.ErrNoRows) {
		return &PasswordResetRequestResponse{EmailSent: true}, nil
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to load profile")
	}

	token, err := createPasswordResetToken(ctx, &profile)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to create password reset token")
	}

	mail, err := passwordResetMail(&profile, token)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to render password reset email")
	}
	if _, err := mailssvc.SendTransactional(ctx, mail); err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to queue password reset email")
	}

	return &PasswordResetRequestResponse{EmailSent: true}, nil
}

// ConfirmPasswordReset validates a reset token and changes the account password.
//
//encore:api public method=POST path=/registration/password-reset/:token
func ConfirmPasswordReset(ctx context.Context, token uuid.UUID, p *PasswordResetConfirmParams) (*PasswordResetConfirmResponse, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	// Consume the token first with a single conditional UPDATE, so two
	// concurrent requests with the same token cannot both succeed.
	var userID uuid.UUID
	var email string
	err := db.QueryRow(ctx, `
		UPDATE password_reset_tokens SET used_at = NOW()
		WHERE token = $1 AND used_at IS NULL AND expires_at > NOW()
		RETURNING user_id, email
	`, token).Scan(&userID, &email)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "password reset token is invalid, expired or already used"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to consume password reset token")
	}

	// SetUserPassword also revokes every existing session of the user.
	if err := authsvc.SetUserPassword(ctx, userID, &authsvc.SetUserPasswordParams{UserID: userID, Password: p.Password}); err != nil {
		// Give the token back so the user can retry with the same link.
		_, _ = db.Exec(ctx, `UPDATE password_reset_tokens SET used_at = NULL WHERE token = $1`, token)
		return nil, err
	}

	return &PasswordResetConfirmResponse{Email: email, Reset: true}, nil
}

func createWelcomeToken(ctx context.Context, profile *profilessvc.Profile) (uuid.UUID, error) {
	token, err := uuid.NewV4()
	if err != nil {
		return uuid.Nil, err
	}
	_, err = db.Exec(ctx, `
		INSERT INTO welcome_tokens (token, user_id, email, expires_at)
		VALUES ($1, $2, $3, $4)
	`, token, profile.UserID, profile.Email, time.Now().Add(welcomeTokenTTL))
	return token, err
}

func createPasswordResetToken(ctx context.Context, profile *profilessvc.Profile) (uuid.UUID, error) {
	token, err := uuid.NewV4()
	if err != nil {
		return uuid.Nil, err
	}
	_, err = db.Exec(ctx, `
		INSERT INTO password_reset_tokens (token, user_id, email, expires_at)
		VALUES ($1, $2, $3, $4)
	`, token, profile.UserID, profile.Email, time.Now().Add(passwordResetTokenTTL))
	return token, err
}

func normalizeEmail(email string) (string, error) {
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

func pruneExpiredTokens(ctx context.Context) {
	_, _ = db.Exec(ctx, `DELETE FROM welcome_tokens WHERE expires_at < NOW() OR used_at IS NOT NULL`)
	_, _ = db.Exec(ctx, `DELETE FROM password_reset_tokens WHERE expires_at < NOW() OR used_at IS NOT NULL`)
}

func emailDomain(email string) (string, bool) {
	_, domain, ok := strings.Cut(email, "@")
	domain = strings.TrimSpace(strings.TrimSuffix(domain, "."))
	return domain, ok && domain != ""
}

func hasValidMX(domain string) (bool, error) {
	records, err := net.LookupMX(domain)
	if err != nil {
		if dnsErr, ok := err.(*net.DNSError); ok && dnsErr.IsNotFound {
			return false, nil
		}
		return false, err
	}
	for _, record := range records {
		if strings.TrimSpace(record.Host) != "." {
			return true, nil
		}
	}
	return false, nil
}

func welcomeMail(profile *profilessvc.Profile, token uuid.UUID) (*mailssvc.TransactionalMailParams, error) {
	name := profile.DisplayName
	if name == "" {
		name = profile.Email
	}

	data := welcomeMailData{
		Name:       name,
		Email:      profile.Email,
		UserID:     profile.UserID.String(),
		WelcomeURL: frontendBaseURL + "/welcome?token=" + token.String(),
	}

	subject, err := renderTextTemplate("mails/welcome_subject.tmpl", data)
	if err != nil {
		return nil, err
	}
	textBody, err := renderTextTemplate("mails/welcome_text.tmpl", data)
	if err != nil {
		return nil, err
	}
	htmlBody, err := renderHTMLTemplate("mails/welcome_html.tmpl", data)
	if err != nil {
		return nil, err
	}

	return &mailssvc.TransactionalMailParams{
		To:       profile.Email,
		Subject:  subject,
		TextBody: textBody,
		HTMLBody: htmlBody,
		Metadata: map[string]string{
			"kind":    "welcome",
			"user_id": profile.UserID.String(),
		},
	}, nil
}

type welcomeMailData struct {
	Name       string
	Email      string
	UserID     string
	WelcomeURL string
}

func passwordResetMail(profile *profilessvc.Profile, token uuid.UUID) (*mailssvc.TransactionalMailParams, error) {
	name := profile.DisplayName
	if name == "" {
		name = profile.Email
	}

	data := passwordResetMailData{
		Name:             name,
		Email:            profile.Email,
		UserID:           profile.UserID.String(),
		PasswordResetURL: frontendBaseURL + "/password-reset?token=" + token.String(),
	}

	subject, err := renderTextTemplate("mails/password_reset_subject.tmpl", data)
	if err != nil {
		return nil, err
	}
	textBody, err := renderTextTemplate("mails/password_reset_text.tmpl", data)
	if err != nil {
		return nil, err
	}
	htmlBody, err := renderHTMLTemplate("mails/password_reset_html.tmpl", data)
	if err != nil {
		return nil, err
	}

	return &mailssvc.TransactionalMailParams{
		To:       profile.Email,
		Subject:  subject,
		TextBody: textBody,
		HTMLBody: htmlBody,
		Metadata: map[string]string{
			"kind":    "password_reset",
			"user_id": profile.UserID.String(),
		},
	}, nil
}

type passwordResetMailData struct {
	Name             string
	Email            string
	UserID           string
	PasswordResetURL string
}

func renderTextTemplate(name string, data any) (string, error) {
	tmpl, err := texttemplate.ParseFS(mailTemplates, name)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

func renderHTMLTemplate(name string, data any) (string, error) {
	tmpl, err := htmltemplate.ParseFS(mailTemplates, name)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}
