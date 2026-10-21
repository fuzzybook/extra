// Service auth owns authentication: credentials, sessions, login and logout.
package auth

import (
	"context"
	"crypto/sha256"
	"errors"
	"net/http"
	"strings"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
	"golang.org/x/crypto/bcrypt"
)

var db = sqldb.NewDatabase("auth", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

var profilesDB = sqldb.Named("profiles")

// PasswordPepper is mixed into every password before hashing, on top of
// bcrypt's per-password salt, so leaked password hashes are useless without it.
var secrets struct {
	PasswordPepper string
}

const sessionTTL = 7 * 24 * time.Hour

// profileStatusActive mirrors profiles.StatusActive; auth cannot import
// profiles because profiles already depends on auth.
const profileStatusActive = 0
const minPasswordLength = 8

type AuthParams struct {
	SessionCookie *http.Cookie `cookie:"session"`
	Authorization string       `header:"Authorization"`
}

// AuthData is exposed to authenticated handlers via auth.Data().
type AuthData struct {
	SessionToken uuid.UUID
	Role         role
}

// AuthHandler authenticates a request by looking up its session token — taken
// from the session cookie, or as a fallback the "Bearer <token>" Authorization
// header — against active sessions.
//
//encore:authhandler
func AuthHandler(ctx context.Context, p *AuthParams) (auth.UID, *AuthData, error) {
	token, ok := sessionToken(p)
	if !ok {
		return "", nil, &errs.Error{Code: errs.Unauthenticated, Message: "missing session credentials"}
	}

	var userID uuid.UUID
	err := db.QueryRow(ctx, `
		SELECT user_id FROM sessions WHERE token = $1 AND expires_at > NOW()
	`, token).Scan(&userID)
	if errors.Is(err, sqldb.ErrNoRows) {
		return "", nil, &errs.Error{Code: errs.Unauthenticated, Message: "invalid or expired session"}
	}
	if err != nil {
		return "", nil, errs.WrapCode(err, errs.Internal, "failed to validate session")
	}

	// Sessions of deleted, suspended or banned profiles must not authenticate.
	var roleStr string
	var status int
	err = profilesDB.QueryRow(ctx, `
		SELECT role, status FROM user_profiles WHERE user_id = $1
	`, userID).Scan(&roleStr, &status)
	if errors.Is(err, sqldb.ErrNoRows) || (err == nil && status != profileStatusActive) {
		return "", nil, &errs.Error{Code: errs.Unauthenticated, Message: "account is not active"}
	}
	if err != nil {
		return "", nil, errs.WrapCode(err, errs.Internal, "failed to load user role")
	}
	r := role(roleStr)
	if !r.IsValid() {
		r = roleUser
	}

	return auth.UID(userID.String()), &AuthData{SessionToken: token, Role: r}, nil
}

func sessionToken(p *AuthParams) (uuid.UUID, bool) {
	if p.SessionCookie != nil {
		if id, err := uuid.FromString(p.SessionCookie.Value); err == nil {
			return id, true
		}
	}
	if rest, ok := strings.CutPrefix(p.Authorization, "Bearer "); ok {
		if id, err := uuid.FromString(rest); err == nil {
			return id, true
		}
	}
	return uuid.Nil, false
}

// pepperedPassword pre-hashes the password together with the application-wide
// pepper secret using SHA-256, producing a fixed-size digest. This both mixes
// in the pepper and keeps the input to bcrypt within its 72-byte limit
// regardless of the original password's length.
func pepperedPassword(password string) ([]byte, error) {
	if strings.TrimSpace(secrets.PasswordPepper) == "" {
		return nil, &errs.Error{Code: errs.Internal, Message: "password pepper is not configured"}
	}
	sum := sha256.Sum256([]byte(password + secrets.PasswordPepper))
	return sum[:], nil
}

func validatePassword(password string) error {
	if strings.TrimSpace(password) == "" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "password is required"}
	}
	if len(password) < minPasswordLength {
		return &errs.Error{Code: errs.InvalidArgument, Message: "password must be at least 8 characters long"}
	}
	return nil
}

// ValidatePassword exposes the shared password rules to other services.
func ValidatePassword(password string) error {
	return validatePassword(password)
}

func authenticatedUserID() (uuid.UUID, error) {
	uid, ok := auth.UserID()
	if !ok {
		return uuid.Nil, &errs.Error{Code: errs.Unauthenticated, Message: "missing auth"}
	}
	return parseUserID(uid)
}

func parseUserID(uid auth.UID) (uuid.UUID, error) {
	userID, err := uuid.FromString(string(uid))
	if err != nil {
		return uuid.Nil, errs.WrapCode(err, errs.Internal, "invalid user id")
	}
	return userID, nil
}

func verifyPassword(hash, password string) error {
	peppered, err := pepperedPassword(password)
	if err != nil {
		return err
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), peppered)
}

func isPasswordReuse(currentHash, newPassword string) bool {
	return verifyPassword(currentHash, newPassword) == nil
}

func hashPassword(password string) (string, error) {
	peppered, err := pepperedPassword(password)
	if err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword(peppered, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
