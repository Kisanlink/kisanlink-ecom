package interceptors

import (
	"context"
	"errors"
)

// MockJWTValidator is a mock implementation for testing
type MockJWTValidator struct {
	ValidateFunc func(ctx context.Context, token string) (map[string]interface{}, error)
}

// ValidateToken validates a JWT token (mock implementation)
func (m *MockJWTValidator) ValidateToken(ctx context.Context, token string) (map[string]interface{}, error) {
	if m.ValidateFunc != nil {
		return m.ValidateFunc(ctx, token)
	}

	// Default mock behavior: accept all tokens
	if token == "" {
		return nil, errors.New("empty token")
	}

	return map[string]interface{}{
		"user_id": "test-user",
		"fpo_id":  "test-fpo",
		"roles":   []string{"ADMIN"},
	}, nil
}
