package user

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// CreateUserRequest represents request to create a user.
// @Description Request structure for creating a new user
type CreateUserRequest struct {
	base.BaseRequest
	Username string   `json:"username" validate:"required,min=3,max=50" example:"john_doe"`
	Email    string   `json:"email" validate:"required,email" example:"john@example.com"`
	FullName string   `json:"full_name" validate:"required,min=2,max=100" example:"John Doe"`
	Role     UserRole `json:"role" validate:"required" example:"customer"`
	Password string   `json:"password" validate:"required,min=8" example:"password123"`
}

// UpdateUserRequest represents request to update a user.
// @Description Request structure for updating an existing user
type UpdateUserRequest struct {
	base.BaseRequest
	Username string   `json:"username,omitempty" validate:"omitempty,min=3,max=50" example:"john_doe"`
	Email    string   `json:"email,omitempty" validate:"omitempty,email" example:"john@example.com"`
	FullName string   `json:"full_name,omitempty" validate:"omitempty,min=2,max=100" example:"John Doe"`
	Role     UserRole `json:"role,omitempty" example:"customer"`
	Status   Status   `json:"status,omitempty" example:"active"`
}

// SimpleUser represents a simplified user structure for API responses.
// @Description Simplified user structure for API responses
type SimpleUser struct {
	ID       string   `json:"id" example:"USER123456789"`
	Username string   `json:"username" example:"john_doe"`
	Email    string   `json:"email" example:"john@example.com"`
	FullName string   `json:"full_name" example:"John Doe"`
	Role     UserRole `json:"role" example:"customer"`
	Status   Status   `json:"status" example:"active"`
}
