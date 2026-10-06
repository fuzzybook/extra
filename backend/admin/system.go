package admin

import (
	"context"

	authsvc "encore.app/auth"
	profilessvc "encore.app/profiles"
)

type SystemOptionsResponse struct {
	Roles       []authsvc.RoleOption          `json:"roles"`
	Statuses    []profilessvc.StatusOption    `json:"statuses"`
	MemberTypes []profilessvc.MemberTypeOption `json:"member_types"`
}

// GetSystemOptions returns the valid roles, profile statuses, and member types,
// each with a display name and underlying value, for use in admin UI dropdowns.
//
//encore:api auth method=GET path=/admin/system/options
func GetSystemOptions(ctx context.Context) (*SystemOptionsResponse, error) {
	return &SystemOptionsResponse{
		Roles:       authsvc.Roles(),
		Statuses:    profilessvc.Statuses(),
		MemberTypes: profilessvc.MemberTypes(),
	}, nil
}

