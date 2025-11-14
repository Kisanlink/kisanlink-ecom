package common_test

import (
	"net/http/httptest"
	"testing"

	"github.com/Kisanlink/kisanlink-ecom/internal/common"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetUserRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupContext   func(*gin.Context)
		expectedRoles  []string
		expectedExists bool
	}{
		{
			name: "snake_case key (primary)",
			setupContext: func(c *gin.Context) {
				c.Set("user_roles", []string{"admin", "user"})
			},
			expectedRoles:  []string{"admin", "user"},
			expectedExists: true,
		},
		{
			name: "camelCase key (fallback)",
			setupContext: func(c *gin.Context) {
				c.Set("userRoles", []string{"super_admin"})
			},
			expectedRoles:  []string{"super_admin"},
			expectedExists: true,
		},
		{
			name: "snake_case takes priority over camelCase",
			setupContext: func(c *gin.Context) {
				c.Set("user_roles", []string{"admin"})
				c.Set("userRoles", []string{"user"})
			},
			expectedRoles:  []string{"admin"},
			expectedExists: true,
		},
		{
			name: "no roles set",
			setupContext: func(_ *gin.Context) {
				// Don't set any roles
			},
			expectedRoles:  nil,
			expectedExists: false,
		},
		{
			name: "invalid type stored",
			setupContext: func(c *gin.Context) {
				c.Set("user_roles", "not-a-slice")
			},
			expectedRoles:  nil,
			expectedExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup context
			tt.setupContext(c)

			// Test GetUserRoles
			roles, exists := common.GetUserRoles(c)

			assert.Equal(t, tt.expectedExists, exists)
			assert.Equal(t, tt.expectedRoles, roles)
		})
	}
}

func TestIsAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		roles    []string
		key      string
		expected bool
	}{
		{
			name:     "admin role with snake_case key",
			roles:    []string{"admin"},
			key:      "user_roles",
			expected: true,
		},
		{
			name:     "super_admin role with snake_case key",
			roles:    []string{"super_admin"},
			key:      "user_roles",
			expected: true,
		},
		{
			name:     "admin role with camelCase key",
			roles:    []string{"admin"},
			key:      "userRoles",
			expected: true,
		},
		{
			name:     "super_admin role with camelCase key",
			roles:    []string{"super_admin"},
			key:      "userRoles",
			expected: true,
		},
		{
			name:     "multiple roles including admin",
			roles:    []string{"user", "admin", "editor"},
			key:      "user_roles",
			expected: true,
		},
		{
			name:     "non-admin roles",
			roles:    []string{"user", "editor"},
			key:      "user_roles",
			expected: false,
		},
		{
			name:     "no roles",
			roles:    []string{},
			key:      "user_roles",
			expected: false,
		},
		{
			name:     "roles not set",
			roles:    nil,
			key:      "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup context
			if tt.key != "" {
				c.Set(tt.key, tt.roles)
			}

			// Test IsAdmin
			result := common.IsAdmin(c)

			assert.Equal(t, tt.expected, result)
		})
	}
}
