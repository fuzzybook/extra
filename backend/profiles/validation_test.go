package profiles

import "testing"

func TestProfileParamsValidate(t *testing.T) {
	if err := (&ProfileParams{DisplayName: "", AvatarURL: ""}).Validate(); err == nil {
		t.Fatal("Validate() should reject empty display name")
	}
	if err := (&ProfileParams{DisplayName: "  Alice  ", AvatarURL: "https://example.com/avatar.png"}).Validate(); err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}
}

func TestRegisterParamsValidate(t *testing.T) {
	if err := (&RegisterParams{Email: "", DisplayName: "Alice", Password: "validpass"}).Validate(); err == nil {
		t.Fatal("Validate() should reject empty email")
	}
	if err := (&RegisterParams{Email: "alice@example.com", DisplayName: "Alice", Password: "weak"}).Validate(); err == nil {
		t.Fatal("Validate() should reject weak password")
	}
	if err := (&RegisterParams{Email: " Alice@Example.com ", DisplayName: " Alice ", Password: "validpass"}).Validate(); err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}
}

func TestPersonalDataParamsValidate(t *testing.T) {
	if err := (&PersonalDataParams{FirstName: "", LastName: "Rossi", Address: "Main 1", Cap: "20100", City: "Milan", Country: "IT"}).Validate(); err == nil {
		t.Fatal("Validate() should reject empty first name")
	}
	if err := (&PersonalDataParams{FirstName: "Mario", LastName: "Rossi", Address: "Main 1", Cap: "20", City: "Milan", Country: "IT"}).Validate(); err == nil {
		t.Fatal("Validate() should reject invalid cap")
	}
	if err := (&PersonalDataParams{FirstName: "  Mario  ", LastName: "  Rossi  ", Address: "  Main 1  ", Cap: "20100", City: "  Milan  ", Country: " IT "}).Validate(); err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}
}
