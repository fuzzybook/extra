package broadcasters

import (
	"context"
	"testing"

	"encore.app/config"
)

func TestCreateBroadcasterParamsValidate(t *testing.T) {
	tests := []struct {
		name    string
		params  *CreateBroadcasterParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "empty display name",
			params: &CreateBroadcasterParams{
				DisplayName: "",
			},
			wantErr: true,
		},
		{
			name: "whitespace display name",
			params: &CreateBroadcasterParams{
				DisplayName: "   ",
			},
			wantErr: true,
		},
		{
			name: "valid broadcaster params with defaults applied",
			params: &CreateBroadcasterParams{
				DisplayName: "DJ Rossi",
			},
			wantErr: false,
		},
		{
			name: "valid full broadcaster params",
			params: &CreateBroadcasterParams{
				DisplayName:   "Live Streamer Alpha",
				Category:      "Music",
				CategoryColor: "#ec4899",
				Bio:           "Official channel for live stream sets.",
				Certified:     true,
				Availability:  "live",
				Phone:         "+39 02 1234567",
				Services:      []string{"Live DJ", "Video chat"},
				Languages:     []string{"IT", "EN"},
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
			if err == nil {
				if tt.params.Category == "" {
					t.Errorf("expected default Category, got empty string")
				}
				if tt.params.CategoryColor == "" {
					t.Errorf("expected default CategoryColor, got empty string")
				}
				if tt.params.Availability == "" {
					t.Errorf("expected default Availability, got empty string")
				}
			}
		})
	}
}

func TestUpdateBroadcasterParamsValidate(t *testing.T) {
	tests := []struct {
		name    string
		params  *UpdateBroadcasterParams
		wantErr bool
	}{
		{
			name:    "nil update params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "empty display name in update",
			params: &UpdateBroadcasterParams{
				DisplayName: "",
			},
			wantErr: true,
		},
		{
			name: "valid update params",
			params: &UpdateBroadcasterParams{
				DisplayName: "New Artist Name",
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
		})
	}
}

func TestBroadcasterImageConfig(t *testing.T) {
	if BroadcasterImageSize.Width != 600 {
		t.Errorf("expected default image width 600, got %d", BroadcasterImageSize.Width)
	}
	if BroadcasterImageSize.Proportion != "2:3" {
		t.Errorf("expected default image proportion 2:3, got %s", BroadcasterImageSize.Proportion)
	}
}

func TestGetCategoriesFromSiteConfig(t *testing.T) {
	resp, err := GetCategories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Categories) == 0 {
		t.Errorf("expected at least 1 category from SiteConfig, got 0")
	}
	foundGeneral := false
	for _, c := range resp.Categories {
		if c.Name == "General" {
			foundGeneral = true
			break
		}
	}
	if !foundGeneral {
		t.Errorf("expected to find General category in SiteConfig")
	}
}

func TestValidateHours(t *testing.T) {
	// Valid slots
	valid := []BroadcasterHours{
		{
			Day: "Lunedì",
			Hours: []BroadcasterTimeSlot{
				{From: "09:00", To: "12:00"},
				{From: "14:00", To: "18:00"},
			},
		},
	}
	if err := ValidateHours(valid); err != nil {
		t.Errorf("expected valid hours, got error: %v", err)
	}

	// Invalid: From >= To
	invalidRange := []BroadcasterHours{
		{
			Day: "Martedì",
			Hours: []BroadcasterTimeSlot{
				{From: "18:00", To: "12:00"},
			},
		},
	}
	if err := ValidateHours(invalidRange); err == nil {
		t.Errorf("expected error for from >= to, got nil")
	}

	// Invalid: overlapping slots
	overlapping := []BroadcasterHours{
		{
			Day: "Mercoledì",
			Hours: []BroadcasterTimeSlot{
				{From: "10:00", To: "13:00"},
				{From: "12:00", To: "15:00"},
			},
		},
	}
	if err := ValidateHours(overlapping); err == nil {
		t.Errorf("expected error for overlapping hours, got nil")
	}
}

func TestGetBroadcasterAttributePresetsFromSiteConfig(t *testing.T) {
	resp, err := GetBroadcasterAttributePresets(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Presets) == 0 {
		t.Errorf("expected at least 1 attribute preset from SiteConfig, got 0")
	}
	foundGenre := false
	for _, p := range resp.Presets {
		if p == "Genere Musicale" {
			foundGenre = true
			break
		}
	}
	if !foundGenre {
		t.Errorf("expected to find 'Genere Musicale' in attribute presets")
	}
}

func TestGetAvailabilityOptionsFromSiteConfig(t *testing.T) {
	resp, err := GetAvailabilityOptions(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := config.GetAvailabilityOptions()
	if len(resp.Options) != len(expected) {
		t.Fatalf("expected %d availability options from SiteConfig, got %d", len(expected), len(resp.Options))
	}
	for i, opt := range resp.Options {
		if opt != expected[i] {
			t.Errorf("expected option %d to be '%s', got '%s'", i, expected[i], opt)
		}
	}
}

func TestGetBroadcasterServicesPresetsFromSiteConfig(t *testing.T) {
	resp, err := GetBroadcasterServicesPresets(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := config.GetBroadcasterServices()
	if len(resp.Services) != len(expected) {
		t.Fatalf("expected %d services presets from SiteConfig, got %d", len(expected), len(resp.Services))
	}
	for i, svc := range resp.Services {
		if svc != expected[i] {
			t.Errorf("expected service %d to be '%s', got '%s'", i, expected[i], svc)
		}
	}
}


