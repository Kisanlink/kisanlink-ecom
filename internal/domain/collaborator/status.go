package collaborator

// Status represents the collaborator status
type Status string

// Status constants define all valid collaborator states
const (
	StatusPending     Status = "PENDING"
	StatusVerified    Status = "VERIFIED"
	StatusActive      Status = "ACTIVE"
	StatusOnHold      Status = "ON_HOLD"
	StatusSuspended   Status = "SUSPENDED"
	StatusInactive    Status = "INACTIVE"
	StatusRejected    Status = "REJECTED"
	StatusDeactivated Status = "DEACTIVATED"
)

// IsValid returns true if the status is valid
func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusVerified, StatusActive, StatusOnHold,
		StatusSuspended, StatusInactive, StatusRejected, StatusDeactivated:
		return true
	}
	return false
}

// String returns the string representation
func (s Status) String() string {
	return string(s)
}
