package perm

import (
	"context"
	"fmt"
)

// MockClient provides a mock implementation of the permission client for testing
type MockClient struct {
	permissions map[string]bool
	users       map[string]string // token -> userID mapping
}

// NewMockClient creates a new mock permission client
func NewMockClient() *MockClient {
	return &MockClient{
		permissions: make(map[string]bool),
		users:       make(map[string]string),
	}
}

// SetPermission sets a permission for testing
func (m *MockClient) SetPermission(subjectID, resourceType, resourceID, action string, allowed bool) {
	key := fmt.Sprintf("%s:%s:%s:%s", subjectID, resourceType, resourceID, action)
	m.permissions[key] = allowed
}

// SetUserToken sets a user token mapping for testing
func (m *MockClient) SetUserToken(token, userID string) {
	m.users[token] = userID
}

// Check implements the permission checking logic
func (m *MockClient) Check(ctx context.Context, subjectID, resourceType, resourceID, action string) (bool, error) {
	key := fmt.Sprintf("%s:%s:%s:%s", subjectID, resourceType, resourceID, action)

	// Default to deny if not explicitly set
	if allowed, exists := m.permissions[key]; exists {
		return allowed, nil
	}

	// Allow platform admin for all actions
	if subjectID == "platform-admin" {
		return true, nil
	}

	// Allow org members to view their own org resources
	if action == "view" && resourceType == "org" && subjectID == resourceID {
		return true, nil
	}

	return false, nil
}

// ValidateToken implements token validation
func (m *MockClient) ValidateToken(ctx context.Context, token string) (string, error) {
	if userID, exists := m.users[token]; exists {
		return userID, nil
	}

	// Default test tokens
	switch token {
	case "valid-token":
		return "test-user-1", nil
	case "admin-token":
		return "platform-admin", nil
	case "collaborator-token":
		return "collaborator-user", nil
	case "fpo-admin-token":
		return "fpo-admin", nil
	case "farmer-token":
		return "farmer-user", nil
	default:
		return "", fmt.Errorf("invalid token: %s", token)
	}
}
