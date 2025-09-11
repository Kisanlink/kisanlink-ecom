package auth

import (
    "context"
    "fmt"
    "sync"
)

// MockAAAClient is a mock implementation of AAAClient for testing
type MockAAAClient struct {
    users       map[string]*AAAUser
    roles       map[string][]*AAARole
    permissions map[string][]string
    mutex       sync.RWMutex
}

// NewMockAAAClient creates a new mock AAA client for testing
func NewMockAAAClient() *MockAAAClient {
    return &MockAAAClient{
        users:       make(map[string]*AAAUser),
        roles:       make(map[string][]*AAARole),
        permissions: make(map[string][]string),
    }
}

// CreateUser creates a user in the mock store
func (m *MockAAAClient) CreateUser(ctx context.Context, user *AAAUser) (*AAAUser, error) {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    // Generate ID if not provided
    if user.ID == "" {
        user.ID = fmt.Sprintf("user_%d", len(m.users)+1)
    }

    m.users[user.ID] = user
    return user, nil
}

// GetUser retrieves a user from the mock store
func (m *MockAAAClient) GetUser(ctx context.Context, userID string) (*AAAUser, error) {
    m.mutex.RLock()
    defer m.mutex.RUnlock()

    user, exists := m.users[userID]
    if !exists {
        return nil, fmt.Errorf("user not found")
    }
    return user, nil
}

// UpdateUser updates a user in the mock store
func (m *MockAAAClient) UpdateUser(ctx context.Context, user *AAAUser) (*AAAUser, error) {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    if _, exists := m.users[user.ID]; !exists {
        return nil, fmt.Errorf("user not found")
    }

    m.users[user.ID] = user
    return user, nil
}

// DeleteUser deletes a user from the mock store
func (m *MockAAAClient) DeleteUser(ctx context.Context, userID string) error {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    if _, exists := m.users[userID]; !exists {
        return fmt.Errorf("user not found")
    }

    delete(m.users, userID)
    delete(m.roles, userID)
    delete(m.permissions, userID)
    return nil
}

// AuthenticateUser authenticates a user (mock implementation)
func (m *MockAAAClient) AuthenticateUser(ctx context.Context, username, password string) (*AuthenticationResponse, error) {
    m.mutex.RLock()
    defer m.mutex.RUnlock()

    for _, user := range m.users {
        if user.Username == username {
            // Mock authentication - always succeed for existing users
            return &AuthenticationResponse{
                AccessToken:  "mock-access-token",
                RefreshToken: "mock-refresh-token",
                ExpiresIn:    3600,
                TokenType:    "Bearer",
                User:         user,
                UserContext: &UserContext{
                    UserID:           user.ID,
                    Username:         user.Username,
                    Email:            user.Email,
                    OrganizationID:   user.OrganizationID,
                    OrganizationName: user.OrganizationName,
                    IsActive:         user.IsActive,
                },
            }, nil
        }
    }

    return nil, fmt.Errorf("authentication failed")
}

// RefreshToken refreshes a token (mock implementation)
func (m *MockAAAClient) RefreshToken(ctx context.Context, refreshToken string) (*AuthenticationResponse, error) {
    // Mock implementation - return new tokens
    return &AuthenticationResponse{
        AccessToken:  "new-mock-access-token",
        RefreshToken: "new-mock-refresh-token",
        ExpiresIn:    3600,
        TokenType:    "Bearer",
    }, nil
}

// ValidateJWT validates a JWT token (mock implementation)
func (m *MockAAAClient) ValidateJWT(ctx context.Context, token string) (bool, error) {
    // Mock implementation - check for valid mock token patterns
    if len(token) == 0 {
        return false, nil
    }

    // Accept tokens that start with "mock-" or "Bearer"
    if token == "invalid-token" {
        return false, nil
    }

    return true, nil
}

// GetUserFromToken gets user context from a token (mock implementation)
func (m *MockAAAClient) GetUserFromToken(ctx context.Context, token string) (*UserContext, error) {
    // Mock implementation - return first user
    m.mutex.RLock()
    defer m.mutex.RUnlock()

    for _, user := range m.users {
        return &UserContext{
            UserID:           user.ID,
            Username:         user.Username,
            Email:            user.Email,
            OrganizationID:   user.OrganizationID,
            OrganizationName: user.OrganizationName,
            IsActive:         user.IsActive,
        }, nil
    }

    return &UserContext{
        UserID:   "mock-user-id",
        Username: "mock-user",
        Email:    "mock@example.com",
        IsActive: true,
    }, nil
}

// EvaluatePermission evaluates a permission (mock implementation)
func (m *MockAAAClient) EvaluatePermission(ctx context.Context, userID, resource, action string) (bool, error) {
    // Mock implementation - always allow
    return true, nil
}

// EvaluateResourcePermission evaluates a resource-specific permission (mock implementation)
func (m *MockAAAClient) EvaluateResourcePermission(ctx context.Context, userID, resourceType, resourceID, action string) (bool, error) {
    // Mock implementation - always allow
    return true, nil
}

// BulkEvaluatePermissions evaluates multiple permissions (mock implementation)
func (m *MockAAAClient) BulkEvaluatePermissions(ctx context.Context, userID string, permissions []PermissionCheck) ([]PermissionResult, error) {
    results := make([]PermissionResult, len(permissions))
    for i, perm := range permissions {
        results[i] = PermissionResult{
            Resource: perm.Resource,
            Action:   perm.Action,
            Allowed:  true,
            Reason:   "mock implementation",
        }
    }
    return results, nil
}

// GetUserRoles gets user roles (mock implementation)
func (m *MockAAAClient) GetUserRoles(ctx context.Context, userID string) ([]*AAARole, error) {
    m.mutex.RLock()
    defer m.mutex.RUnlock()

    roles, exists := m.roles[userID]
    if !exists {
        return []*AAARole{}, nil
    }
    return roles, nil
}

// AssignRole assigns a role to a user (mock implementation)
func (m *MockAAAClient) AssignRole(ctx context.Context, userID, roleID string) error {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    role := &AAARole{
        ID:       roleID,
        Name:     "mock-role",
        IsActive: true,
    }

    m.roles[userID] = append(m.roles[userID], role)
    return nil
}

// RemoveRole removes a role from a user (mock implementation)
func (m *MockAAAClient) RemoveRole(ctx context.Context, userID, roleID string) error {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    roles := m.roles[userID]
    for i, role := range roles {
        if role.ID == roleID {
            m.roles[userID] = append(roles[:i], roles[i+1:]...)
            break
        }
    }
    return nil
}

// GetUserPermissions gets user permissions (mock implementation)
func (m *MockAAAClient) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
    m.mutex.RLock()
    defer m.mutex.RUnlock()

    permissions, exists := m.permissions[userID]
    if !exists {
        return []string{}, nil
    }
    return permissions, nil
}

// ValidateUserOrganization validates user organization membership (mock implementation)
func (m *MockAAAClient) ValidateUserOrganization(ctx context.Context, userID, orgID string) (bool, error) {
    m.mutex.RLock()
    defer m.mutex.RUnlock()

    user, exists := m.users[userID]
    if !exists {
        return false, fmt.Errorf("user not found")
    }

    return user.OrganizationID == orgID, nil
}

// HealthCheck performs a health check (mock implementation)
func (m *MockAAAClient) HealthCheck(ctx context.Context) error {
    return nil
}

// Close closes the mock client (mock implementation)
func (m *MockAAAClient) Close() error {
    return nil
}
