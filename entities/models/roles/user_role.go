package roles

import (
	"time"
)

// UserRole represents the relationship between users and roles in the e-commerce service
// This maps AAA service users to e-commerce specific roles
type UserRole struct {
	Id       string `json:"id" gorm:"primaryKey;type:varchar(255)"`
	UserId   string `json:"user_id" gorm:"type:varchar(255);not null;index"` // Reference to AAA service user ID
	RoleId   string `json:"role_id" gorm:"type:varchar(255);not null;index"` // Reference to e-commerce role
	IsActive bool   `json:"is_active" gorm:"default:true"`

	// Role assignment context
	AssignedBy string     `json:"assigned_by" gorm:"type:varchar(255);index"` // Who assigned this role
	AssignedAt time.Time  `json:"assigned_at" gorm:"autoCreateTime"`
	ExpiresAt  *time.Time `json:"expires_at"`             // Optional role expiration
	Notes      string     `json:"notes" gorm:"type:text"` // Assignment notes

	// Timestamps
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (UserRole) TableName() string {
	return "user_roles"
}

// NewUserRole creates a new user-role assignment
func NewUserRole(userId, roleId, assignedBy string) *UserRole {
	return &UserRole{
		UserId:     userId,
		RoleId:     roleId,
		AssignedBy: assignedBy,
		IsActive:   true,
	}
}

// SetExpiration sets an expiration date for this role assignment
func (ur *UserRole) SetExpiration(expiresAt time.Time) {
	ur.ExpiresAt = &expiresAt
}

// IsExpired checks if this role assignment has expired
func (ur *UserRole) IsExpired() bool {
	if ur.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*ur.ExpiresAt)
}

// IsValid checks if this role assignment is currently valid
func (ur *UserRole) IsValid() bool {
	return ur.IsActive && !ur.IsExpired()
}
