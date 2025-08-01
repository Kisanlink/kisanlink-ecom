package user

import (
	"fmt"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// User represents a user in the system.
// @Description User model representing a user in the system
type User struct {
	base.BaseModel
	Username     string   `json:"username" validate:"required,min=3,max=50" gorm:"uniqueIndex;type:varchar(100)" example:"john_doe"`
	Email        string   `json:"email" validate:"required,email" gorm:"uniqueIndex;type:varchar(255)" example:"john@example.com"`
	FullName     string   `json:"full_name" validate:"required,min=2,max=100" gorm:"type:varchar(255)" example:"John Doe"`
	PasswordHash string   `json:"-" gorm:"type:varchar(255);not null"`
	Role         UserRole `json:"role" validate:"required" gorm:"type:varchar(50);not null" example:"customer"`
	Status       Status   `json:"status" validate:"required" gorm:"type:varchar(50);default:'active'" example:"active"`
}

// NewUser creates a new User with initialized fields.
func NewUser(username, email, fullName string, role UserRole) *User {
	baseModel := base.NewBaseModel("USER", hash.Medium)
	return &User{
		BaseModel: *baseModel,
		Username:  username,
		Email:     email,
		FullName:  fullName,
		Role:      role,
		Status:    StatusActive,
	}
}

// BeforeCreate implements the base model interface.
func (u *User) BeforeCreate() error {
	if err := u.BaseModel.BeforeCreate(); err != nil {
		return err
	}

	if u.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if u.Email == "" {
		return fmt.Errorf("email cannot be empty")
	}
	if u.FullName == "" {
		return fmt.Errorf("full name cannot be empty")
	}

	return nil
}

// UserRole represents user roles in the system.
type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleCustomer UserRole = "customer"
	RoleVendor   UserRole = "vendor"
)

// Status represents general status for entities.
type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusDeleted  Status = "deleted"
)
