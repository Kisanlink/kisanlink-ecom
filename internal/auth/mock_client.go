package auth

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// MockClient implements Client interface for development and testing
type MockClient struct {
	connected bool
}

// NewMockClient creates a new mock AAA client
func NewMockClient() Client {
	return &MockClient{
		connected: true,
	}
}

// ValidateToken validates a JWT token (mock implementation)
func (m *MockClient) ValidateToken(ctx context.Context, token string) (*TokenClaims, error) {
	if token == "" {
		return nil, fmt.Errorf("empty token provided")
	}

	// Mock token validation - accept any non-empty token
	if !strings.HasPrefix(token, "Bearer ") && !strings.HasPrefix(token, "mock_") {
		return nil, fmt.Errorf("invalid token format")
	}

	// Return mock claims
	now := time.Now().Unix()
	claims := &TokenClaims{
		UserID:           "mock_user_123",
		Username:         "mock_user",
		Email:            "mock@example.com",
		TenantID:         "mock_tenant_456",
		OrganizationID:   "mock_org_789",
		OrganizationName: "Mock Organization",
		Roles:            []string{"user", "catalog_manager"},
		Permissions:      []string{"catalog:read", "catalog:write", "catalog:publish"},
		IssuedAt:         now,
		ExpiresAt:        now + 3600, // 1 hour from now
		Issuer:           "mock_aaa_service",
		Audience:         "kisanlink_ecom",
	}

	return claims, nil
}

// Authorize checks if a user has permission to perform an action (mock implementation)
func (m *MockClient) Authorize(ctx context.Context, req *AuthorizeRequest) (*AuthorizeResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("authorize request cannot be nil")
	}

	// Mock authorization - allow all requests in development
	return &AuthorizeResponse{
		Allowed: true,
		Reason:  "Mock authorization - all permissions granted for development",
	}, nil
}

// HealthCheck checks the health of the mock service
func (m *MockClient) HealthCheck(ctx context.Context) error {
	if !m.connected {
		return fmt.Errorf("mock AAA service is not connected")
	}
	return nil
}

// CreateUser creates a user (mock implementation)
func (m *MockClient) CreateUser(ctx context.Context, user *AAAUser) (*AAAUser, error) {
	if user == nil {
		return nil, fmt.Errorf("user cannot be nil")
	}
	// Return the user as-is for mock
	return user, nil
}

// GetUser retrieves a user (mock implementation)
func (m *MockClient) GetUser(ctx context.Context, userID string) (*AAAUser, error) {
	// Return a mock user
	return &AAAUser{
		ID:       userID,
		Username: "mock_user",
		Email:    "mock@example.com",
		IsActive: true,
	}, nil
}

// UpdateUser updates a user (mock implementation)
func (m *MockClient) UpdateUser(ctx context.Context, user *AAAUser) (*AAAUser, error) {
	if user == nil {
		return nil, fmt.Errorf("user cannot be nil")
	}
	// Return the user as-is for mock
	return user, nil
}

// DeleteUser deletes a user (mock implementation)
func (m *MockClient) DeleteUser(ctx context.Context, userID string) error {
	// Mock deletion - always succeeds
	return nil
}

// GetUserRoles retrieves user roles (mock implementation)
func (m *MockClient) GetUserRoles(ctx context.Context, userID string) ([]*AAARole, error) {
	// Return mock roles
	return []*AAARole{
		{
			ID:          "role_1",
			Name:        "user",
			Description: "Basic user role",
			Permissions: []string{"read"},
		},
	}, nil
}

// ValidateJWT validates a JWT token (mock implementation)
func (m *MockClient) ValidateJWT(ctx context.Context, token string) (bool, error) {
	_, err := m.ValidateToken(ctx, token)
	return err == nil, err
}

// AuthenticateUser authenticates a user (mock implementation)
func (m *MockClient) AuthenticateUser(ctx context.Context, req *AuthenticationRequest) (*AuthenticationResponse, error) {
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
func (m *MockClient) RefreshToken(ctx context.Context, refreshToken string) (*AuthenticationResponse, error) {
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
func (m *MockClient) EvaluatePermission(ctx context.Context, userID, resource, action string) (bool, error) {
	// Mock evaluation - allow all for testing
	return true, nil
}

// EvaluateResourcePermission evaluates resource permission (mock implementation)
func (m *MockClient) EvaluateResourcePermission(ctx context.Context, userID, resource, action, resourceID string) (bool, error) {
	// Mock evaluation - allow all for testing
	return true, nil
}

// GetUserFromToken retrieves user from token (mock implementation)
func (m *MockClient) GetUserFromToken(ctx context.Context, token string) (*AAAUser, error) {
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

// BulkEvaluatePermissions evaluates multiple permissions (mock implementation)
func (m *MockClient) BulkEvaluatePermissions(ctx context.Context, userID string, permissions []PermissionCheck) ([]PermissionResult, error) {
	// Mock evaluation - allow all permissions
	results := make([]PermissionResult, len(permissions))
	for i, perm := range permissions {
		results[i] = PermissionResult{
			Resource: perm.Resource,
			Action:   perm.Action,
			Allowed:  true,
			Reason:   "Mock permission - allowed for development",
		}
	}
	return results, nil
}

// Close closes the mock connection
func (m *MockClient) Close() error {
	m.connected = false
	return nil
}
