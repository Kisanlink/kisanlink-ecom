package auth

import (
	"kisanlink-ecom/internal/models/user"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// LoginRequest represents login request.
// @Description Request structure for user login
type LoginRequest struct {
	base.BaseRequest
	Username string `json:"username" validate:"required" example:"john_doe"`
	Password string `json:"password" validate:"required" example:"password123"`
}

// LoginResponse represents login response.
// @Description Response structure for successful login
type LoginResponse struct {
	Token     string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt int64     `json:"expires_at" example:"1640995200"`
	User      user.User `json:"user"`
}

// SimpleUser represents a simplified user structure for API responses.
// @Description Simplified user structure for API responses
type SimpleUser struct {
	ID       string        `json:"id" example:"USER123456789"`
	Username string        `json:"username" example:"john_doe"`
	Email    string        `json:"email" example:"john@example.com"`
	FullName string        `json:"full_name" example:"John Doe"`
	Role     user.UserRole `json:"role" example:"customer"`
	Status   user.Status   `json:"status" example:"active"`
}

// SimpleLoginResponse represents a simplified login response.
// @Description Simplified login response structure
type SimpleLoginResponse struct {
	Token     string     `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt int64      `json:"expires_at" example:"1640995200"`
	User      SimpleUser `json:"user"`
}
