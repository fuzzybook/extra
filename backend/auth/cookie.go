package auth

import (
	"fmt"
	"time"

	"encore.dev"
	"encore.dev/types/uuid"
)

// sessionCookieName must match the cookie read by AuthParams.
const sessionCookieName = "session"

// SessionCookie returns the Set-Cookie value that stores a session token in
// an HttpOnly cookie, so scripts in the page (and any XSS) cannot read it.
// SameSite=Lax keeps the cookie off cross-site POSTs (CSRF). Secure is left
// out only when running locally over plain http.
func SessionCookie(token uuid.UUID, expiresAt time.Time) string {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	return fmt.Sprintf("%s=%s; Path=/; Max-Age=%d; HttpOnly; SameSite=Lax%s",
		sessionCookieName, token, maxAge, secureAttr())
}

// ClearSessionCookie returns the Set-Cookie value that deletes the session cookie.
func ClearSessionCookie() string {
	return fmt.Sprintf("%s=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax%s", sessionCookieName, secureAttr())
}

func secureAttr() string {
	if encore.Meta().Environment.Cloud == encore.CloudLocal {
		return ""
	}
	return "; Secure"
}
