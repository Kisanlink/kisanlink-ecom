package auth

import "time"

// AAAUser represents a user from the AAA service
type AAAUser struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	IsActive  bool   `json:"is_active"`
	// Organization information
	OrganizationID   string `json:"organization_id"`
	OrganizationName string `json:"organization_name"`
}

// AAARole represents a role from the AAA service
type AAARole struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Scope       string `json:"scope"`
	IsActive    bool   `json:"is_active"`
}

// AAAPermission represents a permission from the AAA service
type AAAPermission struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Resource    string   `json:"resource"`
	Actions     []string `json:"actions"`
	IsActive    bool     `json:"is_active"`
}

// JWTClaims represents the claims extracted from a JWT token
type JWTClaims struct {
	UserID           string    `json:"sub"`
	Username         string    `json:"username"`
	Email            string    `json:"email"`
	OrganizationID   string    `json:"org_id"`
	OrganizationName string    `json:"org_name"`
	Roles            []string  `json:"roles"`
	Permissions      []string  `json:"permissions"`
	IssuedAt         time.Time `json:"iat"`
	ExpiresAt        time.Time `json:"exp"`
	Issuer           string    `json:"iss"`
	Audience         string    `json:"aud"`
}

// UserContext represents the authenticated user context
type UserContext struct {
	UserID           string
	Username         string
	Email            string
	OrganizationID   string
	OrganizationName string
	Roles            []string
	Permissions      []string
	IsActive         bool
}

// TokenValidationResult represents the result of token validation
type TokenValidationResult struct {
	Valid       bool
	Claims      *JWTClaims
	UserContext *UserContext
	Error       error
}

// AuthenticationRequest represents an authentication request
type AuthenticationRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// AuthenticationResponse represents an authentication response
type AuthenticationResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
	TokenType    string       `json:"token_type"`
	User         *AAAUser     `json:"user"`
	UserContext  *UserContext `json:"user_context"`
}

// PermissionRequest represents a permission evaluation request
type PermissionRequest struct {
	UserID   string                 `json:"user_id"`
	Resource string                 `json:"resource"`
	Action   string                 `json:"action"`
	Context  map[string]interface{} `json:"context,omitempty"`
}
