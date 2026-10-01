package advertising

import "testing"

func TestCreateAdvertisingParamsValidate(t *testing.T) {
	tests := []struct {
		name    string
		params  *CreateAdvertisingParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "empty email",
			params: &CreateAdvertisingParams{
				Nome:  "Azienda Rossi",
				Email: "",
			},
			wantErr: true,
		},
		{
			name: "invalid email format",
			params: &CreateAdvertisingParams{
				Nome:  "Azienda Rossi",
				Email: "invalid-email-address",
			},
			wantErr: true,
		},
		{
			name: "email missing domain dot",
			params: &CreateAdvertisingParams{
				Nome:  "Azienda Rossi",
				Email: "test@domain",
			},
			wantErr: true,
		},
		{
			name: "empty name",
			params: &CreateAdvertisingParams{
				Nome:  "   ",
				Email: "azienda@rossi.it",
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			params: &CreateAdvertisingParams{
				Nome:   "Azienda Rossi",
				Email:  "azienda@rossi.it",
				Status: "pending",
			},
			wantErr: true,
		},
		{
			name: "valid with default active status",
			params: &CreateAdvertisingParams{
				Nome:  "  Azienda Rossi  ",
				Email: " AZIENDA@rossi.it ",
			},
			wantErr: false,
		},
		{
			name: "valid with explicit inactive status",
			params: &CreateAdvertisingParams{
				Nome:   "Azienda Bianchi",
				Email:  "info@bianchi.com",
				Status: StatusInactive,
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
				if tt.params.Status != StatusActive && tt.params.Status != StatusInactive {
					t.Errorf("expected Status to be active or inactive, got %s", tt.params.Status)
				}
				if tt.params.Email != "azienda@rossi.it" && tt.params.Email != "info@bianchi.com" {
					t.Errorf("expected email to be normalized, got %s", tt.params.Email)
				}
			}
		})
	}
}
