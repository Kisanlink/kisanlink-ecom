package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Kisanlink/kisanlink-ecom/internal/auth"
	"github.com/Kisanlink/kisanlink-ecom/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRBACMiddleware_RequirePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		resource       string
		action         string
		userContext    *auth.UserContext
		permissionResp bool
		permissionErr  error
		expectedStatus int
	}{
		{
			name:     "valid permission allows access",
			resource: "catalog",
			action:   "read",
			userContext: &auth.UserContext{
				UserID:         "user123",
				OrganizationID: "org123",
				IsActive:       true,
			},
			permissionResp: true,
			expectedStatus: 200,
		},
		{
			name:     "denied permission blocks access",
			resource: "catalog",
			action:   "write",
			userContext: &auth.UserContext{
				UserID:         "user123",
				OrganizationID: "org123",
				IsActive:       true,
			},
			permissionResp: false,
			expectedStatus: 403,
		},
		{
			name:           "missing user context returns unauthorized",
			resource:       "catalog",
			action:         "read",
			userContext:    nil,
			expectedStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAAA := new(MockProductionAAAClient)

			if tt.userContext != nil {
				mockAAA.On("Authorize", mock.Anything, mock.MatchedBy(func(req *auth.AuthorizeRequest) bool {
					return req.UserID == tt.userContext.UserID && req.Resource == tt.resource && req.Action == tt.action
				})).Return(&auth.AuthorizeResponse{Allowed: tt.permissionResp}, tt.permissionErr)
			}

			rbacMiddleware := middleware.NewRBACMiddleware(mockAAA, nil)

			router := gin.New()
			router.Use(func(c *gin.Context) {
				if tt.userContext != nil {
					c.Set("user_context", tt.userContext)
				}
				c.Next()
			})
			router.GET("/test", rbacMiddleware.RequirePermission(tt.resource, tt.action), func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "success"})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockAAA.AssertExpectations(t)
		})
	}
}

func TestRBACMiddleware_RequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requiredRoles  []string
		userRoles      []string
		expectedStatus int
	}{
		{
			name:           "user with admin role gets access",
			requiredRoles:  []string{"admin"},
			userRoles:      []string{"admin", "user"},
			expectedStatus: 200,
		},
		{
			name:           "user without required role gets denied",
			requiredRoles:  []string{"admin"},
			userRoles:      []string{"user"},
			expectedStatus: 403,
		},
		{
			name:           "user with any of multiple required roles gets access",
			requiredRoles:  []string{"admin", "moderator"},
			userRoles:      []string{"moderator"},
			expectedStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAAA := new(MockProductionAAAClient)
			rbacMiddleware := middleware.NewRBACMiddleware(mockAAA, nil)

			userContext := &auth.UserContext{
				UserID:         "user123",
				OrganizationID: "org123",
				Roles:          tt.userRoles,
				IsActive:       true,
			}

			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("user_context", userContext)
				c.Next()
			})
			router.GET("/test", rbacMiddleware.RequireRole(tt.requiredRoles...), func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "success"})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRBACMiddleware_RequireOrganization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		requiredOrgID   string
		userOrgID       string
		crossOrgAllowed bool
		expectedStatus  int
		shouldCallAAA   bool
	}{
		{
			name:           "same organization allows access",
			requiredOrgID:  "org123",
			userOrgID:      "org123",
			expectedStatus: 200,
			shouldCallAAA:  false,
		},
		{
			name:            "different organization with AAA approval allows access",
			requiredOrgID:   "org456",
			userOrgID:       "org123",
			crossOrgAllowed: true,
			expectedStatus:  200,
			shouldCallAAA:   true,
		},
		{
			name:            "different organization with AAA denial blocks access",
			requiredOrgID:   "org456",
			userOrgID:       "org123",
			crossOrgAllowed: false,
			expectedStatus:  403,
			shouldCallAAA:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAAA := new(MockProductionAAAClient)

			if tt.shouldCallAAA {
				mockAAA.On("Authorize", mock.Anything, mock.MatchedBy(func(req *auth.AuthorizeRequest) bool {
					return req.UserID == "user123" && req.Resource == "organization" && req.ResourceID == tt.requiredOrgID
				})).Return(&auth.AuthorizeResponse{Allowed: tt.crossOrgAllowed}, nil)
			}

			rbacMiddleware := middleware.NewRBACMiddleware(mockAAA, nil)

			userContext := &auth.UserContext{
				UserID:         "user123",
				OrganizationID: tt.userOrgID,
				IsActive:       true,
			}

			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("user_context", userContext)
				c.Next()
			})
			router.GET("/test", rbacMiddleware.RequireOrganization(tt.requiredOrgID), func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "success"})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockAAA.AssertExpectations(t)
		})
	}
}

func TestRBACMiddleware_SkipAuthPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &middleware.RBACConfig{
		SkipAuthPaths: []string{"/health", "/docs"},
		CacheEnabled:  false,
	}

	mockAAA := new(MockProductionAAAClient)
	rbacMiddleware := middleware.NewRBACMiddleware(mockAAA, config)

	router := gin.New()
	router.GET("/health", rbacMiddleware.RequirePermission("test", "read"), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "health ok"})
	})
	router.GET("/secure", rbacMiddleware.RequirePermission("test", "read"), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "secure ok"})
	})

	// Test skip path
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// Test secure path without user context
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/secure", nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, 401, w2.Code)
}

func TestRBACMiddleware_CachePermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &middleware.RBACConfig{
		CacheEnabled: true,
		CacheTTL:     1 * time.Minute,
	}

	mockAAA := new(MockProductionAAAClient)

	// Mock should be called only once due to caching
	mockAAA.On("Authorize", mock.Anything, mock.MatchedBy(func(req *auth.AuthorizeRequest) bool {
		return req.UserID == "user123" && req.Resource == "catalog" && req.Action == "read"
	})).Return(&auth.AuthorizeResponse{Allowed: true}, nil).Once()

	rbacMiddleware := middleware.NewRBACMiddleware(mockAAA, config)

	userContext := &auth.UserContext{
		UserID:         "user123",
		OrganizationID: "org123",
		IsActive:       true,
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_context", userContext)
		c.Next()
	})
	router.GET("/test", rbacMiddleware.RequirePermission("catalog", "read"), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// First request - should call AAA service
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// Second request - should use cache
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)

	mockAAA.AssertExpectations(t)
}
