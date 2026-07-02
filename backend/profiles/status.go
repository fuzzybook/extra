package profiles

type Status int

const (
	StatusActive Status = iota
	StatusInactive
	StatusSuspended
	StatusDeleted
	StatusPending
	StatusWaitingDeletion
	StatusBanned
)

// StatusOption describes a valid profile status for display in admin UIs.
type StatusOption struct {
	Name      string `json:"name"`
	Value     Status `json:"value"`
	Updatable bool   `json:"updatable"`
}

// Statuses returns every valid profile status with a display name and its underlying value.
func Statuses() []StatusOption {
	return []StatusOption{
		{Name: "Active", Value: StatusActive, Updatable: true},
		{Name: "Inactive", Value: StatusInactive, Updatable: true},
		{Name: "Suspended", Value: StatusSuspended, Updatable: true},
		{Name: "Deleted", Value: StatusDeleted, Updatable: false},
		{Name: "Pending", Value: StatusPending, Updatable: true},
		{Name: "Waiting deletion", Value: StatusWaitingDeletion, Updatable: false},
		{Name: "Banned", Value: StatusBanned, Updatable: true},
	}
}

// IsValid reports whether s is one of the defined Status values.
func (s Status) IsValid() bool {
	return s >= StatusActive && s <= StatusBanned
}
