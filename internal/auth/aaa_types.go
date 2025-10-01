package auth

import "time"

// RoleDetails represents full role information from JWT
type RoleDetails struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Scope       string   `json:"scope,omitempty"`
	Description string   `json:"description,omitempty"`
	ParentID    string   `json:"parent_id,omitempty"`
	IsActive    bool     `json:"is_active"`
	Permissions []string `json:"permissions,omitempty"`
}

// UserRoleDetails represents user-role association from JWT
type UserRoleDetails struct {
	ID       string       `json:"id"`
	UserID   string       `json:"user_id"`
	RoleID   string       `json:"role_id"`
	IsActive bool         `json:"is_active"`
	Role     *RoleDetails `json:"role,omitempty"`
}

// OrganizationDetails represents organization information from JWT
type OrganizationDetails struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	TenantID string `json:"tenant_id,omitempty"`
}

// GroupDetails represents group information from JWT
type GroupDetails struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	OrganizationID string `json:"organization_id,omitempty"`
	Description    string `json:"description,omitempty"`
}

// UserContextDetails represents nested user_context from JWT v2.0
type UserContextDetails struct {
	ID            string                 `json:"id"`
	Username      string                 `json:"username"`
	PhoneNumber   string                 `json:"phone_number,omitempty"`
	CountryCode   string                 `json:"country_code,omitempty"`
	IsValidated   bool                   `json:"is_validated"`
	Roles         []*RoleDetails         `json:"roles,omitempty"`
	Organizations []*OrganizationDetails `json:"organizations,omitempty"`
	Groups        []*GroupDetails        `json:"groups,omitempty"`
}

// JWTClaims represents JWT token claims (for backward compatibility)
type JWTClaims struct {
	UserID           string
	Username         string
	Email            string
	TenantID         string
	OrganizationID   string
	OrganizationName string
	Roles            []string
	Permissions      []string
	IssuedAt         time.Time
	ExpiresAt        time.Time
	Issuer           string
	Audience         string
}

// UserContext represents the authenticated user context
type UserContext struct {
	UserID           string                 `json:"user_id"`
	Username         string                 `json:"username"`
	Email            string                 `json:"email,omitempty"`
	PhoneNumber      string                 `json:"phone_number,omitempty"`
	CountryCode      string                 `json:"country_code,omitempty"`
	TenantID         string                 `json:"tenant_id,omitempty"`
	OrganizationID   string                 `json:"organization_id,omitempty"`
	OrganizationName string                 `json:"organization_name,omitempty"`
	Roles            []string               `json:"roles,omitempty"`
	RoleIDs          []string               `json:"role_ids,omitempty"`
	Permissions      []string               `json:"permissions,omitempty"`
	Scopes           []string               `json:"scopes,omitempty"`
	IsActive         bool                   `json:"is_active"`
	IsValidated      bool                   `json:"is_validated"`
	UserRoles        []*UserRoleDetails     `json:"user_roles,omitempty"`
	Organizations    []*OrganizationDetails `json:"organizations,omitempty"`
	Groups           []*GroupDetails        `json:"groups,omitempty"`
	TokenType        string                 `json:"token_type,omitempty"`
	TokenVersion     string                 `json:"token_version,omitempty"`
	Subject          string                 `json:"sub,omitempty"`
	SessionID        string                 `json:"session_id,omitempty"`
	JTI              string                 `json:"jti,omitempty"`
	Issuer           string                 `json:"iss,omitempty"`
	Audience         string                 `json:"aud,omitempty"`
	TenantContext    map[string]interface{} `json:"tenant_context,omitempty"`
	UserContextData  *UserContextDetails    `json:"user_context,omitempty"`
}

// TokenValidationResult represents the result of token validation
type TokenValidationResult struct {
	Valid       bool
	Claims      *TokenClaims
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
	UserContext  *UserContext `json:"user_context"`
	User         *AAAUser     `json:"user"` // For backward compatibility
}

// PermissionRequest represents a permission evaluation request
type PermissionRequest struct {
	UserID   string         `json:"user_id"`
	Resource string         `json:"resource"`
	Action   string         `json:"action"`
	Context  map[string]any `json:"context,omitempty"`
}

// AAAUser represents a user in the AAA system
type AAAUser struct {
	ID               string    `json:"id"`
	Username         string    `json:"username"`
	Email            string    `json:"email"`
	Phone            string    `json:"phone"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	TenantID         string    `json:"tenant_id"`
	OrganizationID   string    `json:"organization_id"`
	OrganizationName string    `json:"organization_name"`
	Roles            []string  `json:"roles"`
	Permissions      []string  `json:"permissions"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// AAARole represents a role in the AAA system
type AAARole struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	TenantID    string   `json:"tenant_id"`
}

// PermissionCheck represents a permission check request
type PermissionCheck struct {
	Resource string         `json:"resource"`
	Action   string         `json:"action"`
	Context  map[string]any `json:"context,omitempty"`
}

// PermissionResult represents the result of a permission check
type PermissionResult struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Allowed  bool   `json:"allowed"`
	Reason   string `json:"reason,omitempty"`
}
