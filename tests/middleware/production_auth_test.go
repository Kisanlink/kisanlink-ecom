package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kisanlink-ecom/internal/auth"
	"kisanlink-ecom/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProductionAAAClient implements auth.Client for testing
type MockProductionAAAClient struct {
	mock.Mock
}

func (m *MockProductionAAAClient) ValidateToken(ctx context.Context, token string) (*auth.TokenClaims, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.TokenClaims), args.Error(1)
}

func (m *MockProductionAAAClient) Authorize(ctx context.Context, req *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.AuthorizeResponse), args.Error(1)
}

// Add missing methods to satisfy the interface
func (m *MockProductionAAAClient) CreateUser(ctx context.Context, user *auth.AAAUser) (*auth.AAAUser, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.AAAUser), args.Error(1)
}

func (m *MockProductionAAAClient) GetUser(ctx context.Context, userID string) (*auth.AAAUser, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.AAAUser), args.Error(1)
}

func (m *MockProductionAAAClient) UpdateUser(ctx context.Context, user *auth.AAAUser) (*auth.AAAUser, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.AAAUser), args.Error(1)
}

func (m *MockProductionAAAClient) DeleteUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockProductionAAAClient) GetUserRoles(ctx context.Context, userID string) ([]*auth.AAARole, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*auth.AAARole), args.Error(1)
}

func (m *MockProductionAAAClient) BulkEvaluatePermissions(ctx context.Context, userID string, permissions []auth.PermissionCheck) ([]auth.PermissionResult, error) {
	args := m.Called(ctx, userID, permissions)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]auth.PermissionResult), args.Error(1)
}

func (m *MockProductionAAAClient) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockProductionAAAClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockProductionAAAClient) ValidateJWT(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}

func (m *MockProductionAAAClient) AuthenticateUser(ctx context.Context, req *auth.AuthenticationRequest) (*auth.AuthenticationResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.AuthenticationResponse), args.Error(1)
}

func (m *MockProductionAAAClient) RefreshToken(ctx context.Context, refreshToken string) (*auth.AuthenticationResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.AuthenticationResponse), args.Error(1)
}

func (m *MockProductionAAAClient) EvaluatePermission(ctx context.Context, userID, resource, action string) (bool, error) {
	args := m.Called(ctx, userID, resource, action)
	return args.Bool(0), args.Error(1)
}

func (m *MockProductionAAAClient) EvaluateResourcePermission(ctx context.Context, userID, resource, action, resourceID string) (bool, error) {
	args := m.Called(ctx, userID, resource, action, resourceID)
	return args.Bool(0), args.Error(1)
}

func (m *MockProductionAAAClient) GetUserFromToken(ctx context.Context, token string) (*auth.AAAUser, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.AAAUser), args.Error(1)
}

func TestProductionAuthMiddleware_JWTAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		token          string
		setupMock      func(*MockProductionAAAClient)
		expectedStatus int
		expectedError  string
	}{
		{
			name:  "Valid token",
			token: "valid-token",
			setupMock: func(m *MockProductionAAAClient) {
				m.On("ValidateToken", mock.Anything, "valid-token").Return(&auth.TokenClaims{
					UserID:         "user123",
					Username:       "testuser",
					Email:          "test@example.com",
					TenantID:       "tenant123",
					OrganizationID: "org123",
					Roles:          []string{"user"},
					Permissions:    []string{"catalog_read"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing token",
			token:          "",
			setupMock:      func(m *MockProductionAAAClient) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authentication token required",
		},
		{
			name:  "Invalid token",
			token: "invalid-token",
			setupMock: func(m *MockProductionAAAClient) {
				m.On("ValidateToken", mock.Anything, "invalid-token").Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Token validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockProductionAAAClient{}
			tt.setupMock(mockClient)

			authMiddleware := middleware.NewProductionAuthMiddleware(mockClient, nil)

			router := gin.New()
			router.Use(authMiddleware.JWTAuthenticationMiddleware())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.False(t, response["success"].(bool))
				assert.Contains(t, response["error"].(map[string]interface{})["message"].(string), tt.expectedError)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestProductionAuthMiddleware_Authorization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		resource       string
		action         string
		setupMock      func(*MockProductionAAAClient)
		expectedStatus int
		expectedError  string
	}{
		{
			name:     "Authorized access",
			resource: "catalog",
			action:   "read",
			setupMock: func(m *MockProductionAAAClient) {
				m.On("ValidateToken", mock.Anything, "valid-token").Return(&auth.TokenClaims{
					UserID:         "user123",
					TenantID:       "tenant123",
					OrganizationID: "org123",
				}, nil)
				m.On("Authorize", mock.Anything, mock.MatchedBy(func(req *auth.AuthorizeRequest) bool {
					return req.UserID == "user123" && req.Resource == "catalog" && req.Action == "read"
				})).Return(&auth.AuthorizeResponse{Allowed: true}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "Unauthorized access",
			resource: "catalog",
			action:   "delete",
			setupMock: func(m *MockProductionAAAClient) {
				m.On("ValidateToken", mock.Anything, "valid-token").Return(&auth.TokenClaims{
					UserID:         "user123",
					TenantID:       "tenant123",
					OrganizationID: "org123",
				}, nil)
				m.On("Authorize", mock.Anything, mock.MatchedBy(func(req *auth.AuthorizeRequest) bool {
					return req.UserID == "user123" && req.Resource == "catalog" && req.Action == "delete"
				})).Return(&auth.AuthorizeResponse{Allowed: false, Reason: "insufficient permissions"}, nil)
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Insufficient permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockProductionAAAClient{}
			tt.setupMock(mockClient)

			authMiddleware := middleware.NewProductionAuthMiddleware(mockClient, nil)

			router := gin.New()
			router.Use(authMiddleware.JWTAuthenticationMiddleware())
			router.Use(authMiddleware.RequirePermission(tt.resource, tt.action))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer valid-token")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.False(t, response["success"].(bool))
				assert.Contains(t, response["error"].(map[string]interface{})["message"].(string), tt.expectedError)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestProductionAuthMiddleware_PermissionCaching(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockClient := &MockProductionAAAClient{}

	// Setup mock to return valid token
	mockClient.On("ValidateToken", mock.Anything, "valid-token").Return(&auth.TokenClaims{
		UserID:         "user123",
		TenantID:       "tenant123",
		OrganizationID: "org123",
	}, nil)

	// Setup mock to return authorized response - should only be called once due to caching
	mockClient.On("Authorize", mock.Anything, mock.MatchedBy(func(req *auth.AuthorizeRequest) bool {
		return req.UserID == "user123" && req.Resource == "catalog" && req.Action == "read"
	})).Return(&auth.AuthorizeResponse{Allowed: true}, nil).Once()

	config := &middleware.ProductionAuthConfig{
		PermissionCacheTTL: 1 * time.Second, // Short TTL for testing
	}
	authMiddleware := middleware.NewProductionAuthMiddleware(mockClient, config)

	router := gin.New()
	router.Use(authMiddleware.JWTAuthenticationMiddleware())
	router.Use(authMiddleware.RequirePermission("catalog", "read"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// First request - should call AAA service
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("Authorization", "Bearer valid-token")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request - should use cache
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("Authorization", "Bearer valid-token")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	mockClient.AssertExpectations(t)
}

func TestProductionAuthMiddleware_SkipAuthPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockClient := &MockProductionAAAClient{}
	authMiddleware := middleware.NewProductionAuthMiddleware(mockClient, nil)

	router := gin.New()
	router.Use(authMiddleware.JWTAuthenticationMiddleware())
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "protected"})
	})

	// Health endpoint should not require authentication
	req1 := httptest.NewRequest("GET", "/health", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Protected endpoint should require authentication
	req2 := httptest.NewRequest("GET", "/protected", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusUnauthorized, w2.Code)

	mockClient.AssertExpectations(t)
}

func TestProductionAuthMiddleware_OptionalAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockClient := &MockProductionAAAClient{}
	mockClient.On("ValidateToken", mock.Anything, "valid-token").Return(&auth.TokenClaims{
		UserID:         "user123",
		TenantID:       "tenant123",
		OrganizationID: "org123",
	}, nil)

	authMiddleware := middleware.NewProductionAuthMiddleware(mockClient, nil)

	router := gin.New()
	router.Use(authMiddleware.OptionalAuth())
	router.GET("/test", func(c *gin.Context) {
		userContext, exists := middleware.GetUserContext(c)
		if exists {
			c.JSON(http.StatusOK, gin.H{"user_id": userContext.UserID})
		} else {
			c.JSON(http.StatusOK, gin.H{"user_id": "anonymous"})
		}
	})

	// Request with valid token
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("Authorization", "Bearer valid-token")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	var response1 map[string]interface{}
	err := json.Unmarshal(w1.Body.Bytes(), &response1)
	assert.NoError(t, err)
	assert.Equal(t, "user123", response1["user_id"])

	// Request without token
	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	var response2 map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &response2)
	assert.NoError(t, err)
	assert.Equal(t, "anonymous", response2["user_id"])

	mockClient.AssertExpectations(t)
}

func TestProductionAuthMiddleware_ResourcePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockClient := &MockProductionAAAClient{}
	mockClient.On("ValidateToken", mock.Anything, "valid-token").Return(&auth.TokenClaims{
		UserID:         "user123",
		TenantID:       "tenant123",
		OrganizationID: "org123",
	}, nil)

	mockClient.On("Authorize", mock.Anything, mock.MatchedBy(func(req *auth.AuthorizeRequest) bool {
		return req.UserID == "user123" && req.Resource == "catalog" && req.Action == "update" && req.ResourceID == "catalog123"
	})).Return(&auth.AuthorizeResponse{Allowed: true}, nil)

	authMiddleware := middleware.NewProductionAuthMiddleware(mockClient, nil)

	router := gin.New()
	router.Use(authMiddleware.JWTAuthenticationMiddleware())
	router.Use(authMiddleware.RequireResourcePermission("catalog", "update", middleware.ExtractIDFromParam("id")))
	router.PUT("/catalog/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "updated"})
	})

	req := httptest.NewRequest("PUT", "/catalog/catalog123", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockClient.AssertExpectations(t)
}

func TestProductionAuthMiddleware_ConvenienceMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockClient := &MockProductionAAAClient{}
	mockClient.On("ValidateToken", mock.Anything, "valid-token").Return(&auth.TokenClaims{
		UserID:         "user123",
		TenantID:       "tenant123",
		OrganizationID: "org123",
	}, nil)

	mockClient.On("Authorize", mock.Anything, mock.MatchedBy(func(req *auth.AuthorizeRequest) bool {
		return req.Resource == "catalog" && req.Action == "read"
	})).Return(&auth.AuthorizeResponse{Allowed: true}, nil)

	authMiddleware := middleware.NewProductionAuthMiddleware(mockClient, nil)

	router := gin.New()
	router.Use(authMiddleware.JWTAuthenticationMiddleware())
	router.GET("/catalog", authMiddleware.RequireCatalogRead(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "catalog list"})
	})

	req := httptest.NewRequest("GET", "/catalog", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockClient.AssertExpectations(t)
}
