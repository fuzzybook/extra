package profiles

import (
	"context"
	"testing"
	"time"
)

func TestInsertStoresSignupData(t *testing.T) {
	ctx := context.Background()
	email := "signup-" + time.Now().Format("150405.000000") + "@example.com"
	profile, err := Insert(ctx, &RegisterParams{
		Email:               email,
		DisplayName:         "Nick",
		Password:            "validpass",
		Language:            "fr",
		RequestedMemberType: MemberTypeBroadcaster,
		Newsletter:          true,
		TermsAccepted:       true,
		Personal:            &SignupPersonal{Name: "Anna Rossi", Phone: "+41 79 123 45 67", Country: "ch", Address: "Rue 1"},
	})
	if err != nil {
		t.Fatalf("Insert() = %v", err)
	}
	// Chi chiede broadcaster resta visitor finché un admin non conferma.
	if profile.MemberType != MemberTypeVisitor {
		t.Errorf("member_type = %q, want visitor", profile.MemberType)
	}

	var (
		language, requested string
		newsletter          bool
		termsAccepted       *time.Time
	)
	if err := db.QueryRow(ctx, `
		SELECT language, requested_member_type, newsletter, terms_accepted_at
		FROM user_profiles WHERE user_id = $1
	`, profile.UserID).Scan(&language, &requested, &newsletter, &termsAccepted); err != nil {
		t.Fatalf("select profile = %v", err)
	}
	if language != "fr" || requested != MemberTypeBroadcaster || !newsletter || termsAccepted == nil {
		t.Errorf("profile row = %q %q %v %v", language, requested, newsletter, termsAccepted)
	}

	var name, phone, country, address string
	if err := db.QueryRow(ctx, `
		SELECT first_name, phone, country, address FROM personal_data WHERE user_id = $1
	`, profile.UserID).Scan(&name, &phone, &country, &address); err != nil {
		t.Fatalf("select personal data = %v", err)
	}
	if name != "Anna Rossi" || phone != "+41 79 123 45 67" || country != "CH" || address != "Rue 1" {
		t.Errorf("personal data = %q %q %q %q", name, phone, country, address)
	}

	// Un profilo creato senza dati di registrazione usa i valori predefiniti.
	plain, err := Insert(ctx, &RegisterParams{Email: "plain-" + email, DisplayName: "Plain", Password: "validpass"})
	if err != nil {
		t.Fatalf("Insert(plain) = %v", err)
	}
	if err := db.QueryRow(ctx, `
		SELECT language, requested_member_type, terms_accepted_at FROM user_profiles WHERE user_id = $1
	`, plain.UserID).Scan(&language, &requested, &termsAccepted); err != nil {
		t.Fatalf("select plain = %v", err)
	}
	if language != "en" || requested != "" || termsAccepted != nil || plain.MemberType != MemberTypeVisitor {
		t.Errorf("plain profile = %q %q %v %q", language, requested, termsAccepted, plain.MemberType)
	}
}
