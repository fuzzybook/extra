package registration

import "testing"

func TestNormalizeEmail(t *testing.T) {
	got, err := normalizeEmail("  Alice@Example.com  ")
	if err != nil {
		t.Fatalf("normalizeEmail() returned unexpected error: %v", err)
	}
	if got != "alice@example.com" {
		t.Fatalf("normalizeEmail() = %q, want %q", got, "alice@example.com")
	}

	if _, err := normalizeEmail(" "); err == nil {
		t.Fatal("normalizeEmail() should reject empty email")
	}
}

func TestPasswordResetConfirmParamsValidate(t *testing.T) {
	if err := (&PasswordResetConfirmParams{Password: "weak"}).Validate(); err == nil {
		t.Fatal("Validate() should reject weak password")
	}
	if err := (&PasswordResetConfirmParams{Password: "validpass"}).Validate(); err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}
}
