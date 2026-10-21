package advertising

import (
	"testing"
	"time"
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
								Width:    300,
								Height:   250,
								ImageURL: "https://example.com/banner.jpg",
								URL:      "https://example.com",
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


func TestPromoParamsRejectNonHTTPURLs(t *testing.T) {
	for _, field := range []func(*PromoContent){
		func(c *PromoContent) { c.URL = "javascript:alert(document.cookie)" },
		func(c *PromoContent) { c.ImageURL = "data:image/svg+xml,<svg onload=alert(1)>" },
		func(c *PromoContent) { c.MobileImageURL = "vbscript:x" },
	} {
		c := PromoContent{URL: "https://example.com", ImageURL: "https://example.com/i.png"}
		field(&c)
		create := &CreatePromoParams{Position: PromoLeft, Section: string(PromoLeftWeek), Item: []PromoItem{{Content: []PromoContent{c}}}}
		if err := create.Validate(); err == nil {
			t.Errorf("CreatePromoParams.Validate() accepted %+v", c)
		}
		update := &UpdatePromoParams{Position: PromoLeft, Section: string(PromoLeftWeek), Item: []PromoItem{{Content: []PromoContent{c}}}}
		if err := update.Validate(); err == nil {
			t.Errorf("UpdatePromoParams.Validate() accepted %+v", c)
		}
	}
}
