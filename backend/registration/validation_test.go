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

func TestRegisterParamsProfileParams(t *testing.T) {
	valid := func() *RegisterParams {
		return &RegisterParams{
			Email:       "alice@example.com",
			DisplayName: "Alice",
			Password:    "validpass",
			Language:    "it",
			MemberType:  "broadcaster",
			Name:        " Alice  Rossi ",
			Phone:       "+39 333 1234567",
			Country:     "it",
			Address:     "Via Roma 1\nMilano",
			Newsletter:  true,
			AcceptTerms: true,
		}
	}

	for name, mutate := range map[string]func(p *RegisterParams){
		"terms not accepted": func(p *RegisterParams) { p.AcceptTerms = false },
		"no language":        func(p *RegisterParams) { p.Language = "" },
		"unknown language":   func(p *RegisterParams) { p.Language = "es" },
		"no member type":     func(p *RegisterParams) { p.MemberType = "" },
		"admin member type":  func(p *RegisterParams) { p.MemberType = "admin" },
		"no name":            func(p *RegisterParams) { p.Name = "  " },
		"invalid phone":      func(p *RegisterParams) { p.Phone = "call me" },
		"invalid country":    func(p *RegisterParams) { p.Country = "Italy" },
		"no display name":    func(p *RegisterParams) { p.DisplayName = "" },
	} {
		p := valid()
		mutate(p)
		if _, err := p.profileParams(); err == nil {
			t.Errorf("profileParams() accepted %s", name)
		}
	}

	// Telefono, paese e indirizzo sono facoltativi.
	p := valid()
	p.Phone, p.Country, p.Address = "", "", ""
	if _, err := p.profileParams(); err != nil {
		t.Fatalf("profileParams() without optional fields = %v", err)
	}

	got, err := valid().profileParams()
	if err != nil {
		t.Fatalf("profileParams() = %v", err)
	}
	if !got.TermsAccepted || got.Language != "it" || got.RequestedMemberType != "broadcaster" || !got.Newsletter {
		t.Errorf("profileParams() = %+v", got)
	}
	if got.Personal == nil || got.Personal.Name != "Alice Rossi" || got.Personal.Country != "IT" || got.Personal.Phone != "+39 333 1234567" {
		t.Errorf("personal = %+v", got.Personal)
	}
}
