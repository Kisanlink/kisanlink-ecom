package auth

import (
	"context"
	"fmt"
)

// MockAAAClient is a mock implementation of AAAClient for testing
type MockAAAClient struct {
	users       map[string]*AAAUser
	roles       map[string][]*AAARole
	permissions map[string][]PermissionResult
	healthError error
}

// NewMockAAAClient creates a new mock AAA client
func NewMockAAAClient() *MockAAAClient {
	return &MockAAAClient{
		users:       make(map[string]*AAAUser),
		roles:       make(map[string][]*AAARole),
		permissions: make(map[string][]PermissionResult),
	}
}

// ValidateToken validates a JWT token (mock implementation)
func (m *MockAAAClient) ValidateToken(ctx context.Context, token string) (*TokenClaims, error) {
	if token == "" {
		return nil, fmt.Errorf("empty token")
	}

	// Return mock claims
	return &TokenClaims{
		UserID:           "test-user-123",
		Username:         "testuser",
		Email:            "test@example.com",
		TenantID:         "test-tenant",
		OrganizationID:   "test-org",
		OrganizationName: "Test Organization",
		Roles:            []string{"user"},
		Permissions:      []string{"read", "write"},
		IssuedAt:         1640995200, // 2022-01-01
		ExpiresAt:        1640998800, // 2022-01-01 + 1 hour
		Issuer:           "test-issuer",
		Audience:         "test-audience",
	}, nil
}

// Authorize checks authorization (mock implementation)
func (m *MockAAAClient) Authorize(ctx context.Context, req *AuthorizeRequest) (*AuthorizeResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("authorize request cannot be nil")
	}

	// Mock authorization - allow all for testing
	return &AuthorizeResponse{
		Allowed: true,
		Reason:  "Mock authorization",
	}, nil
}

// CreateUser creates a user (mock implementation)
func (m *MockAAAClient) CreateUser(ctx context.Context, user *AAAUser) (*AAAUser, error) {
	if user == nil {
		return nil, fmt.Errorf("user cannot be nil")
	}

	m.users[user.ID] = user
	return user, nil
}

// GetUser retrieves a user (mock implementation)
func (m *MockAAAClient) GetUser(ctx context.Context, userID string) (*AAAUser, error) {
	if user, exists := m.users[userID]; exists {
		return user, nil
	}
	return nil, fmt.Errorf("user not found")
}

// UpdateUser updates a user (mock implementation)
func (m *MockAAAClient) UpdateUser(ctx context.Context, user *AAAUser) (*AAAUser, error) {
	if user == nil {
		return nil, fmt.Errorf("user cannot be nil")
	}

	m.users[user.ID] = user
	return user, nil
}

// DeleteUser deletes a user (mock implementation)
func (m *MockAAAClient) DeleteUser(ctx context.Context, userID string) error {
	delete(m.users, userID)
	return nil
}

// GetUserRoles retrieves user roles (mock implementation)
func (m *MockAAAClient) GetUserRoles(ctx context.Context, userID string) ([]*AAARole, error) {
	if roles, exists := m.roles[userID]; exists {
		return roles, nil
	}
	return []*AAARole{}, nil
}

// ValidateJWT validates a JWT token (mock implementation)
func (m *MockAAAClient) ValidateJWT(ctx context.Context, token string) (bool, error) {
	_, err := m.ValidateToken(ctx, token)
	return err == nil, err
}

// AuthenticateUser authenticates a user (mock implementation)
func (m *MockAAAClient) AuthenticateUser(ctx context.Context, req *AuthenticationRequest) (*AuthenticationResponse, error) {
	if req == nil || req.Username == "" {
		return nil, fmt.Errorf("invalid authentication request")
	}

	return &AuthenticationResponse{
		AccessToken:  "mock_access_token_" + req.Username,
		RefreshToken: "mock_refresh_token_" + req.Username,
		ExpiresIn:    3600,
		TokenType:    "Bearer",
		UserContext: &UserContext{
			UserID:   "mock_user_" + req.Username,
			Username: req.Username,
			Email:    req.Username + "@example.com",
			IsActive: true,
		},
	}, nil
}

// RefreshToken refreshes a token (mock implementation)
func (m *MockAAAClient) RefreshToken(ctx context.Context, refreshToken string) (*AuthenticationResponse, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("invalid refresh token")
	}

	return &AuthenticationResponse{
		AccessToken:  "new_mock_access_token",
		RefreshToken: "new_mock_refresh_token",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
	}, nil
}

// EvaluatePermission evaluates a single permission (mock implementation)
func (m *MockAAAClient) EvaluatePermission(ctx context.Context, userID, resource, action string) (bool, error) {
	// Mock evaluation - allow all for testing
	return true, nil
}

// EvaluateResourcePermission evaluates resource permission (mock implementation)
func (m *MockAAAClient) EvaluateResourcePermission(ctx context.Context, userID, resource, action, resourceID string) (bool, error) {
	// Mock evaluation - allow all for testing
	return true, nil
}

// GetUserFromToken retrieves user from token (mock implementation)
func (m *MockAAAClient) GetUserFromToken(ctx context.Context, token string) (*AAAUser, error) {
	if token == "" {
		return nil, fmt.Errorf("empty token")
	}

	return &AAAUser{
		ID:       "mock_user_from_token",
		Username: "mock_user",
		Email:    "mock@example.com",
		IsActive: true,
	}, nil
}

// BulkEvaluatePermissions evaluates permissions (mock implementation)
func (m *MockAAAClient) BulkEvaluatePermissions(ctx context.Context, userID string, permissions []PermissionCheck) ([]PermissionResult, error) {
	if results, exists := m.permissions[userID]; exists {
		return results, nil
	}

	// Return all permissions as allowed by default
	results := make([]PermissionResult, len(permissions))
	for i, perm := range permissions {
		results[i] = PermissionResult{
			Resource: perm.Resource,
			Action:   perm.Action,
			Allowed:  true,
			Reason:   "Mock permission",
		}
	}
	return results, nil
}

// HealthCheck performs health check (mock implementation)
func (m *MockAAAClient) HealthCheck(ctx context.Context) error {
	return m.healthError
}

// Close closes the connection (mock implementation)
func (m *MockAAAClient) Close() error {
	return nil
}

// SetHealthError sets the health check error for testing
func (m *MockAAAClient) SetHealthError(err error) {
	m.healthError = err
}

// AddUser adds a user to the mock store
func (m *MockAAAClient) AddUser(user *AAAUser) {
	m.users[user.ID] = user
}

// AddUserRoles adds roles for a user
func (m *MockAAAClient) AddUserRoles(userID string, roles []*AAARole) {
	m.roles[userID] = roles
}

// SetUserPermissions sets permissions for a user
func (m *MockAAAClient) SetUserPermissions(userID string, permissions []PermissionResult) {
	m.permissions[userID] = permissions
}
