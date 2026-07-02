package auth

type role string

const (
	roleUser  role = "user"
	roleAdmin role = "admin"
)

func (r role) String() string {
	return string(r)
}

func (r role) IsValid() bool {
	return r == roleUser || r == roleAdmin
}

func (r role) IsAdmin() bool {
	return r == roleAdmin
}

func (r role) IsUser() bool {
	return r == roleUser
}

// RoleOption describes a valid role for display in admin UIs.
type RoleOption struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Roles returns every valid role with a display name and its underlying value.
func Roles() []RoleOption {
	return []RoleOption{
		{Name: "User", Value: string(roleUser)},
		{Name: "Admin", Value: string(roleAdmin)},
	}
}

// IsValidRole reports whether value is a valid role.
func IsValidRole(value string) bool {
	return role(value).IsValid()
}
