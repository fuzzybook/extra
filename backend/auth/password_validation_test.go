package auth

import (
	"testing"

	authpkg "encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/types/uuid"
)

func TestValidatePasswordRejectsShortAndEmptyValues(t *testing.T) {
	for _, tc := range []struct {
		name     string
		password string
	}{
		{name: "empty", password: ""},
		{name: "spaces", password: "   "},
		{name: "short", password: "short"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePassword(tc.password)
			if err == nil {
				t.Fatalf("validatePassword(%q) = nil, want error", tc.password)
			}
			e, ok := err.(*errs.Error)
			if !ok {
				t.Fatalf("validatePassword(%q) returned %T, want *errs.Error", tc.password, err)
			}
			if e.Code != errs.InvalidArgument {
				t.Fatalf("validatePassword(%q) returned code %v, want %v", tc.password, e.Code, errs.InvalidArgument)
			}
		})
	}
}

func TestLoginParamsValidate(t *testing.T) {
	uid, err := uuid.NewV4()
	if err != nil {
		t.Fatalf("uuid.NewV4() failed: %v", err)
	}

	if err := (&LoginParams{UserID: uuid.Nil, Password: "validpass"}).Validate(); err == nil {
		t.Fatal("Validate() should reject nil user id")
	}
	if err := (&LoginParams{UserID: uid, Password: "short"}).Validate(); err == nil {
		t.Fatal("Validate() should reject short password")
	}
	if err := (&LoginParams{UserID: uid, Password: "validpass"}).Validate(); err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}
}

func TestRegisterParamsValidate(t *testing.T) {
	uid, err := uuid.NewV4()
	if err != nil {
		t.Fatalf("uuid.NewV4() failed: %v", err)
	}

	if err := (&RegisterParams{UserID: uuid.Nil, Password: "validpass"}).Validate(); err == nil {
		t.Fatal("Validate() should reject nil user id")
	}
	if err := (&RegisterParams{UserID: uid, Password: "weak"}).Validate(); err == nil {
		t.Fatal("Validate() should reject weak password")
	}
}

func TestPepperedPasswordRequiresConfiguredSecret(t *testing.T) {
	old := secrets.PasswordPepper
	secrets.PasswordPepper = ""
	defer func() { secrets.PasswordPepper = old }()

	if _, err := pepperedPassword("validpass"); err == nil {
		t.Fatal("pepperedPassword() should fail when PasswordPepper is empty")
	}
}

func TestPasswordReuseDetection(t *testing.T) {
	old := secrets.PasswordPepper
	secrets.PasswordPepper = "pepper-test"
	defer func() { secrets.PasswordPepper = old }()

	hash, err := hashPassword("this-is-a-valid-password")
	if err != nil {
		t.Fatalf("hashPassword() failed: %v", err)
	}
	if !isPasswordReuse(hash, "this-is-a-valid-password") {
		t.Fatal("isPasswordReuse() should return true for the same password")
	}
	if isPasswordReuse(hash, "another-very-different-pass") {
		t.Fatal("isPasswordReuse() should return false for a different password")
	}
}

func TestParseUserID(t *testing.T) {
	uid, err := uuid.NewV4()
	if err != nil {
		t.Fatalf("uuid.NewV4() failed: %v", err)
	}

	parsed, err := parseUserID(authpkg.UID(uid.String()))
	if err != nil {
		t.Fatalf("parseUserID() returned unexpected error: %v", err)
	}
	if parsed != uid {
		t.Fatalf("parseUserID() mismatch: got %v, want %v", parsed, uid)
	}

	if _, err := parseUserID(authpkg.UID("not-a-uuid")); err == nil {
		t.Fatal("parseUserID() should reject invalid UUID strings")
	}
}
