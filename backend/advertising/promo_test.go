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
