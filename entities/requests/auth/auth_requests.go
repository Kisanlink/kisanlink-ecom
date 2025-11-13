package auth

// LoginRequest represents the request to login a user
type LoginRequest struct {
	Username string `json:"username" binding:"required,max=255"`
	Password string `json:"password" binding:"required,min=6,max=255"`
}

// RefreshTokenRequest represents the request to refresh an access token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ChangePasswordRequest represents the request to change user password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6,max=255"`
}

// ForgotPasswordRequest represents the request to initiate password reset
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents the request to reset password with token
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=255"`
}

// RegisterRequest represents the request to register a new user
type RegisterRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=50"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6,max=255"`
	FirstName string `json:"first_name" binding:"required,max=100"`
	LastName  string `json:"last_name" binding:"required,max=100"`
	Phone     string `json:"phone" binding:"omitempty,max=20"`
}
