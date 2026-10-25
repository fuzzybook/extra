package advertising

import (
	"context"
	"testing"
	"time"

	authsvc "encore.app/auth"
	encoreauth "encore.dev/beta/auth"
	"encore.dev/et"
	"encore.dev/types/uuid"
)

func TestCreatePromoParamsValidate(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		params  *CreatePromoParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "invalid negative position",
			params: &CreatePromoParams{
				Position: -1,
			},
			wantErr: true,
		},
		{
			name: "invalid position > 2",
			params: &CreatePromoParams{
				Position: 3,
			},
			wantErr: true,
		},
		{
			name: "valid PromoLeft with week section",
			params: &CreatePromoParams{
				Position:   PromoLeft,
				Section:    string(PromoLeftWeek),
				ActiveFrom: &now,
			},
			wantErr: false,
		},
		{
			name: "valid PromoRight1 with show section",
			params: &CreatePromoParams{
				Position: PromoRight1,
				Section:  string(PromoRight1Show),
			},
			wantErr: false,
		},
		{
			name: "valid PromoRight2 with items and top section",
			params: &CreatePromoParams{
				Position: PromoRight2,
				Section:  string(PromoRight2Top),
				Item: []PromoItem{
					{
						Title: "Banner Estivo",
						Content: []PromoContent{
							{
								ImageURL: "https://example.com/banner.jpg",
								URL:      "https://example.com/landing",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid section for PromoLeft",
			params: &CreatePromoParams{
				Position: PromoLeft,
				Section:  "invalid-section",
			},
			wantErr: true,
		},
		{
			name: "invalid section for PromoRight1",
			params: &CreatePromoParams{
				Position: PromoRight1,
				Section:  string(PromoLeftWeek),
			},
			wantErr: true,
		},
		{
			name: "invalid section for PromoRight2",
			params: &CreatePromoParams{
				Position: PromoRight2,
				Section:  string(PromoRight1Show),
			},
			wantErr: true,
		},
		{
			name: "invalid week < 1",
			params: &CreatePromoParams{
				Position: PromoLeft,
				Section:  string(PromoLeftWeek),
				Weeks:    []int{0},
			},
			wantErr: true,
		},
		{
			name: "invalid week > 53",
			params: &CreatePromoParams{
				Position: PromoLeft,
				Section:  string(PromoLeftWeek),
				Weeks:    []int{54},
			},
			wantErr: true,
		},
		{
			name: "valid weeks",
			params: &CreatePromoParams{
				Position: PromoLeft,
				Section:  string(PromoLeftWeek),
				Weeks:    []int{1, 10, 52},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && tt.params.Item == nil {
				t.Errorf("expected Item slice to be initialized, got nil")
			}
		})
	}
}

func TestUpdatePromoParamsValidate(t *testing.T) {
	tests := []struct {
		name    string
		params  *UpdatePromoParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "invalid negative position",
			params: &UpdatePromoParams{
				Position: -1,
			},
			wantErr: true,
		},
		{
			name: "valid PromoLeft week",
			params: &UpdatePromoParams{
				Position: PromoLeft,
				Section:  string(PromoLeftWeek),
			},
			wantErr: false,
		},
		{
			name: "valid PromoRight1 show with items",
			params: &UpdatePromoParams{
				Position: PromoRight1,
				Section:  string(PromoRight1Show),
				Item: []PromoItem{
					{Title: "Updated title"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid section for PromoRight1",
			params: &UpdatePromoParams{
				Position: PromoRight1,
				Section:  "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid week in update",
			params: &UpdatePromoParams{
				Position: PromoRight1,
				Section:  string(PromoRight1Show),
				Weeks:    []int{0},
			},
			wantErr: true,
		},
		{
			name: "valid weeks in update",
			params: &UpdatePromoParams{
				Position: PromoRight1,
				Section:  string(PromoRight1Show),
				Weeks:    []int{5, 6, 7},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && tt.params.Item == nil {
				t.Errorf("expected Item slice to be initialized, got nil")
			}
		})
	}
}

func TestGetSectionOrder(t *testing.T) {
	// PromoLeft order: week (1), raccomended (2), discover (3)
	if ord := getSectionOrder(PromoLeft, string(PromoLeftWeek)); ord != 1 {
		t.Errorf("expected 1 for PromoLeftWeek, got %d", ord)
	}
	if ord := getSectionOrder(PromoLeft, string(PromoLeftRaccomended)); ord != 2 {
		t.Errorf("expected 2 for PromoLeftRaccomended, got %d", ord)
	}
	if ord := getSectionOrder(PromoLeft, string(PromoLeftDiscover)); ord != 3 {
		t.Errorf("expected 3 for PromoLeftDiscover, got %d", ord)
	}
	if ord := getSectionOrder(PromoLeft, "unknown"); ord != 99 {
		t.Errorf("expected 99 for unknown PromoLeft section, got %d", ord)
	}

	// PromoRight1 order: show (1), service (2)
	if ord := getSectionOrder(PromoRight1, string(PromoRight1Show)); ord != 1 {
		t.Errorf("expected 1 for PromoRight1Show, got %d", ord)
	}
	if ord := getSectionOrder(PromoRight1, string(PromoRight1Service)); ord != 2 {
		t.Errorf("expected 2 for PromoRight1Service, got %d", ord)
	}

	// PromoRight2 order: top (1), selected (2)
	if ord := getSectionOrder(PromoRight2, string(PromoRight2Top)); ord != 1 {
		t.Errorf("expected 1 for PromoRight2Top, got %d", ord)
	}
	if ord := getSectionOrder(PromoRight2, string(PromoRight2Selected)); ord != 2 {
		t.Errorf("expected 2 for PromoRight2Selected, got %d", ord)
	}
}

func TestIsPromoActiveInWeek(t *testing.T) {
	// All year (empty weeks)
	allYear := &Promo{Year: 2026, Weeks: []int{}}
	if !isPromoActiveInWeek(allYear, 2026, 10) {
		t.Errorf("expected all-year promo to be active in week 10")
	}
	if isPromoActiveInWeek(allYear, 2025, 10) {
		t.Errorf("expected promo for year 2026 to not match year 2025")
	}

	// Specific weeks
	specificWeeks := &Promo{Year: 2026, Weeks: []int{10, 11, 12}}
	if !isPromoActiveInWeek(specificWeeks, 2026, 11) {
		t.Errorf("expected promo to be active in week 11")
	}
	if isPromoActiveInWeek(specificWeeks, 2026, 13) {
		t.Errorf("expected promo to NOT be active in week 13")
	}
}

func TestPromoParamsRequireBroadcaster(t *testing.T) {
	// Senza broadcaster il contenuto non è valido.
	noBroadcaster := &CreatePromoParams{Position: PromoLeft, Section: string(PromoLeftWeek), Item: []PromoItem{{Content: []PromoContent{{}}}}}
	if err := noBroadcaster.Validate(); err == nil {
		t.Error("CreatePromoParams.Validate() accepted a content without broadcaster")
	}

	for _, shortID := range []string{"", "abc", "abc-12", "abcdefg", "àbc123"} {
		create := &CreatePromoParams{Position: PromoLeft, Section: string(PromoLeftWeek), Item: []PromoItem{{Content: []PromoContent{{Broadcaster: &PromoBroadcaster{ShortID: shortID}}}}}}
		if err := create.Validate(); err == nil {
			t.Errorf("CreatePromoParams.Validate() accepted short_id %q", shortID)
		}
		update := &UpdatePromoParams{Position: PromoLeft, Section: string(PromoLeftWeek), Item: []PromoItem{{Content: []PromoContent{{Broadcaster: &PromoBroadcaster{ShortID: shortID}}}}}}
		if err := update.Validate(); err == nil {
			t.Errorf("UpdatePromoParams.Validate() accepted short_id %q", shortID)
		}
	}

	// Lo short_id viene normalizzato e i dati del broadcaster inviati dal
	// client vengono scartati (sono solo in lettura).
	p := &CreatePromoParams{Position: PromoLeft, Section: string(PromoLeftWeek), Item: []PromoItem{{Content: []PromoContent{{
		Broadcaster: &PromoBroadcaster{ShortID: " Ab12Cd ", DisplayName: "x", ImageURL: "javascript:alert(1)"},
	}}}}}
	if err := p.Validate(); err != nil {
		t.Fatalf("Validate() = %v", err)
	}
	c := p.Item[0].Content[0]
	if c.Broadcaster == nil || *c.Broadcaster != (PromoBroadcaster{ShortID: "ab12cd"}) {
		t.Fatalf("content after Validate = %+v, want only broadcaster.short_id = ab12cd", c.Broadcaster)
	}
}

func TestPromoParamsRightImageContent(t *testing.T) {
	right := func(c PromoContent) *CreatePromoParams {
		return &CreatePromoParams{Position: PromoRight1, Section: string(PromoRight1Show), Item: []PromoItem{{Content: []PromoContent{c}}}}
	}

	// L'immagine è obbligatoria; immagine e link solo http/https.
	for name, c := range map[string]PromoContent{
		"no image":         {URL: "https://example.com"},
		"javascript image": {ImageURL: "javascript:alert(1)"},
		"javascript link":  {ImageURL: "https://example.com/a.jpg", URL: "javascript:alert(1)"},
		"broadcaster only": {Broadcaster: &PromoBroadcaster{ShortID: "ab12cd"}},
		"relative image":   {ImageURL: "/images/a.jpg"},
	} {
		if err := right(c).Validate(); err == nil {
			t.Errorf("Validate() accepted right content %q", name)
		}
	}

	// Il broadcaster inviato su una promo right viene scartato, new_window
	// resta solo con un link.
	p := right(PromoContent{
		ImageURL:    " https://example.com/a.jpg ",
		NewWindow:   true,
		MobileOnly:  true,
		Broadcaster: &PromoBroadcaster{ShortID: "ab12cd"},
	})
	if err := p.Validate(); err != nil {
		t.Fatalf("Validate() = %v", err)
	}
	want := PromoContent{ImageURL: "https://example.com/a.jpg", MobileOnly: true}
	if got := p.Item[0].Content[0]; got.Broadcaster != nil || got.ImageURL != want.ImageURL || got.URL != "" || got.NewWindow || !got.MobileOnly {
		t.Fatalf("content after Validate = %+v, want %+v", got, want)
	}

	// Su PromoLeft immagine e link inviati dal client vengono scartati.
	left := &UpdatePromoParams{Position: PromoLeft, Section: string(PromoLeftWeek), Item: []PromoItem{{Content: []PromoContent{{
		Broadcaster: &PromoBroadcaster{ShortID: "ab12cd"},
		ImageURL:    "https://example.com/a.jpg",
		URL:         "https://example.com",
		NewWindow:   true,
	}}}}}
	if err := left.Validate(); err != nil {
		t.Fatalf("Validate() = %v", err)
	}
	if got := left.Item[0].Content[0]; got.ImageURL != "" || got.URL != "" || got.NewWindow {
		t.Fatalf("left content after Validate = %+v, want only the broadcaster", got)
	}
}

func TestRightPromoImageLinkRoundTrip(t *testing.T) {
	adminID, _ := uuid.NewV4()
	et.OverrideAuthInfo(encoreauth.UID(adminID.String()), &authsvc.AuthData{Role: "admin"})
	defer et.OverrideAuthInfo("", nil)
	ctx := context.Background()

	created, err := CreatePromo(ctx, &CreatePromoParams{
		Position: PromoRight2,
		Section:  string(PromoRight2Top),
		Item: []PromoItem{{Title: "Banner", Content: []PromoContent{{
			ImageURL:  "https://example.com/banner.jpg",
			URL:       "https://example.com/landing",
			NewWindow: true,
		}}}},
	})
	if err != nil {
		t.Fatalf("CreatePromo() = %v", err)
	}
	want := PromoContent{ImageURL: "https://example.com/banner.jpg", URL: "https://example.com/landing", NewWindow: true}

	updated, err := UpdatePromo(ctx, created.ID, &UpdatePromoParams{
		Position: PromoRight2,
		Section:  string(PromoRight2Selected),
		Item:     created.Item,
	})
	if err != nil {
		t.Fatalf("UpdatePromo() = %v", err)
	}
	if got := updated.Item[0].Content[0]; got != want {
		t.Fatalf("updated content = %+v, want %+v", got, want)
	}

	// Lettura da admin e da utente non loggato: il contenuto senza
	// broadcaster resta com'è.
	for _, asAdmin := range []bool{true, false} {
		if !asAdmin {
			et.OverrideAuthInfo("", nil)
		}
		list, err := ListPromotions(ctx)
		if err != nil {
			t.Fatalf("ListPromotions(admin=%v) = %v", asAdmin, err)
		}
		found := false
		for _, p := range list.Promotions {
			if p.ID == created.ID {
				found = true
				if got := p.Item[0].Content[0]; got != want {
					t.Errorf("listed content (admin=%v) = %+v, want %+v", asAdmin, got, want)
				}
			}
		}
		if !found {
			t.Errorf("promo %d not listed (admin=%v)", created.ID, asAdmin)
		}
	}
}
