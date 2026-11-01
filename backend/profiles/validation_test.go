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

func TestInitialMemberType(t *testing.T) {
	for requested, want := range map[string]string{
		"":                    MemberTypeVisitor,
		MemberTypeVisitor:     MemberTypeVisitor,
		MemberTypeBroadcaster: MemberTypeVisitor,
		MemberTypeAffiliate:   MemberTypeAffiliate,
		MemberTypeAdvertiser:  MemberTypeAdvertiser,
		MemberTypeClub:        MemberTypeClub,
	} {
		if got := initialMemberType(requested); got != want {
			t.Errorf("initialMemberType(%q) = %q, want %q", requested, got, want)
		}
	}
}

func TestRegisterParamsContactAndCountry(t *testing.T) {
	base := func(memberType string, personal *SignupPersonal) *RegisterParams {
		return &RegisterParams{Email: "club@example.com", DisplayName: "Club", Password: "validpass", RequestedMemberType: memberType, Personal: personal}
	}
	full := func() *SignupPersonal {
		return &SignupPersonal{Name: "Club Srl", Phone: "+39 02 1234567", Country: "it", Address: "Via Roma 1, Milano"}
	}

	// Inserzionisti e club devono indicare telefono, paese e indirizzo.
	for _, memberType := range []string{MemberTypeAdvertiser, MemberTypeClub} {
		if err := base(memberType, nil).Validate(); err == nil {
			t.Errorf("%s without personal data should be rejected", memberType)
		}
		for field, clear := range map[string]func(*SignupPersonal){
			"phone":   func(p *SignupPersonal) { p.Phone = "" },
			"country": func(p *SignupPersonal) { p.Country = "" },
			"address": func(p *SignupPersonal) { p.Address = "  " },
		} {
			personal := full()
			clear(personal)
			if err := base(memberType, personal).Validate(); err == nil {
				t.Errorf("%s without %s should be rejected", memberType, field)
			}
		}
		if err := base(memberType, full()).Validate(); err != nil {
			t.Errorf("%s with full contact data: %v", memberType, err)
		}
	}
	// Per gli altri tipi restano facoltativi.
	if err := base(MemberTypeVisitor, &SignupPersonal{Name: "Mario"}).Validate(); err != nil {
		t.Errorf("visitor without contact data: %v", err)
	}

	// Paese: solo stati europei.
	for country, ok := range map[string]bool{"IT": true, "ch": true, "XK": true, "GB": true, "US": false, "BR": false, "TR": false} {
		personal := full()
		personal.Country = country
		err := base(MemberTypeClub, personal).Validate()
		if ok != (err == nil) {
			t.Errorf("country %s: err = %v, want accepted %v", country, err, ok)
		}
	}
}

