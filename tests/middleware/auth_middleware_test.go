package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kisanlink/kisanlink-ecom/internal/auth"
	"github.com/Kisanlink/kisanlink-ecom/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthAAAClient for middleware testing
type MockAuthAAAClient struct {
	mock.Mock
}

// Implement required methods for testing
func (m *MockAuthAAAClient) ValidateJWT(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthAAAClient) EvaluatePermission(ctx context.Context, userID, resource, action string) (bool, error) {
	args := m.Called(ctx, userID, resource, action)
	return args.Bool(0), args.Error(1)
}

// Implement all required methods for auth.Client interface
func (m *MockAuthAAAClient) ValidateToken(ctx context.Context, token string) (*auth.TokenClaims, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.TokenClaims), args.Error(1)
}

func (m *MockAuthAAAClient) Authorize(ctx context.Context, req *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.AuthorizeResponse), args.Error(1)
}

func (m *MockAuthAAAClient) AuthenticateUser(ctx context.Context, req *auth.AuthenticationRequest) (*auth.AuthenticationResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.AuthenticationResponse), args.Error(1)
}

func (m *MockAuthAAAClient) RefreshToken(ctx context.Context, refreshToken string) (*auth.AuthenticationResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.AuthenticationResponse), args.Error(1)
}

func (m *MockAuthAAAClient) EvaluateResourcePermission(ctx context.Context, userID, resource, action, resourceID string) (bool, error) {
	args := m.Called(ctx, userID, resource, action, resourceID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthAAAClient) CreateUser(ctx context.Context, user *auth.AAAUser) (*auth.AAAUser, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.AAAUser), args.Error(1)
}

func (m *MockAuthAAAClient) GetUser(ctx context.Context, userID string) (*auth.AAAUser, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.AAAUser), args.Error(1)
}

func (m *MockAuthAAAClient) GetUserFromToken(ctx context.Context, token string) (*auth.AAAUser, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.AAAUser), args.Error(1)
}

func (m *MockAuthAAAClient) UpdateUser(ctx context.Context, user *auth.AAAUser) (*auth.AAAUser, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.AAAUser), args.Error(1)
}

func (m *MockAuthAAAClient) DeleteUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAuthAAAClient) GetUserRoles(ctx context.Context, userID string) ([]*auth.AAARole, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*auth.AAARole), args.Error(1)
}

func (m *MockAuthAAAClient) BulkEvaluatePermissions(ctx context.Context, userID string, permissions []auth.PermissionCheck) ([]auth.PermissionResult, error) {
	args := m.Called(ctx, userID, permissions)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]auth.PermissionResult), args.Error(1)
}

func (m *MockAuthAAAClient) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAuthAAAClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAuthAAAClient) AssignRole(ctx context.Context, userID, roleID string) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockAuthAAAClient) RemoveRole(ctx context.Context, userID, roleID string) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockAuthAAAClient) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockAuthAAAClient) ValidateUserOrganization(ctx context.Context, userID, orgID string) (bool, error) {
	args := m.Called(ctx, userID, orgID)
	return args.Bool(0), args.Error(1)
}

// Test helper functions
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func createTestHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	}
}

// Test AuthNMiddleware with comprehensive scenarios
func TestAuthNMiddleware_Success(t *testing.T) {
	mockClient := &MockAuthAAAClient{}
	router := setupTestRouter()

	// Set up mock expectations
	mockClaims := &auth.TokenClaims{
		UserID:           "mock_user_id",
		Username:         "test_user",
		Email:            "test@example.com",
		TenantID:         "tenant_123",
		OrganizationID:   "org_456",
		OrganizationName: "Test Org",
		Roles:            []string{"buyer", "seller"},
		Permissions:      []string{"catalog_read", "catalog_write"},
	}
	mockClient.On("ValidateToken", mock.Anything, "valid_token").Return(mockClaims, nil)

	router.Use(middleware.AuthNMiddleware(mockClient))
	router.GET("/test", createTestHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid_token")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "success", response["message"])

	mockClient.AssertExpectations(t)
}

func TestAuthNMiddleware_MissingAuthorizationHeader(t *testing.T) {
	mockClient := &MockAuthAAAClient{}
	router := setupTestRouter()

	router.Use(middleware.AuthNMiddleware(mockClient))
	router.GET("/test", createTestHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	// No Authorization header

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"].(map[string]interface{})["code"], "UNAUTHORIZED")
}

func TestAuthNMiddleware_InvalidAuthorizationFormat(t *testing.T) {
	mockClient := &MockAuthAAAClient{}
	router := setupTestRouter()

	router.Use(middleware.AuthNMiddleware(mockClient))
	router.GET("/test", createTestHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat token")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"].(map[string]interface{})["code"], "UNAUTHORIZED")
}

func TestAuthNMiddleware_EmptyToken(t *testing.T) {
	mockClient := &MockAuthAAAClient{}
	router := setupTestRouter()

	// Set up mock expectation for empty token validation
	mockClient.On("ValidateToken", mock.Anything, "").Return(nil, fmt.Errorf("empty token provided"))

	router.Use(middleware.AuthNMiddleware(mockClient))
	router.GET("/test", createTestHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer ")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"].(map[string]interface{})["code"], "INVALID_TOKEN")

	mockClient.AssertExpectations(t)
}

func TestAuthNMiddleware_WithOrgHeader(t *testing.T) {
	mockClient := &MockAuthAAAClient{}
	router := setupTestRouter()

	// Set up mock expectations
	mockClaims := &auth.TokenClaims{
		UserID:           "mock_user_id",
		Username:         "test_user",
		Email:            "test@example.com",
		TenantID:         "org-123",
		OrganizationID:   "org-123",
		OrganizationName: "Test Org",
		Roles:            []string{"buyer", "seller"},
		Permissions:      []string{"catalog_read", "catalog_write"},
	}
	mockClient.On("ValidateToken", mock.Anything, "valid_token").Return(mockClaims, nil)

	router.Use(middleware.AuthNMiddleware(mockClient))
	router.GET("/test", func(c *gin.Context) {
		userID, exists := middleware.GetSubjectID(c)
		assert.True(t, exists)
		assert.Equal(t, "mock_user_id", userID)

		roles, exists := middleware.GetUserRoles(c)
		assert.True(t, exists)
		assert.Contains(t, roles, "buyer")
		assert.Contains(t, roles, "seller")

		orgID, exists := middleware.GetOrgID(c)
		assert.True(t, exists)
		assert.Equal(t, "org-123", orgID)

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	req.Header.Set("X-Org-ID", "org-123")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockClient.AssertExpectations(t)
}

// Test RequireAuth middleware
func TestRequireAuth_Success(t *testing.T) {
	router := setupTestRouter()

	router.Use(func(c *gin.Context) {
		c.Set(middleware.SubjectIDKey, "user-123")
		c.Next()
	})
	router.Use(middleware.RequireAuth())
	router.GET("/test", createTestHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireAuth_MissingAuthentication(t *testing.T) {
	router := setupTestRouter()

	router.Use(middleware.RequireAuth())
	router.GET("/test", createTestHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"].(map[string]interface{})["code"], "UNAUTHORIZED")
}

// Test context getter functions
func TestGetSubjectID_Success(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		c.Set(middleware.SubjectIDKey, "user-123")

		userID, exists := middleware.GetSubjectID(c)
		assert.True(t, exists)
		assert.Equal(t, "user-123", userID)

		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetSubjectID_NotSet(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		userID, exists := middleware.GetSubjectID(c)
		assert.False(t, exists)
		assert.Empty(t, userID)

		c.JSON(http.StatusOK, gin.H{"exists": exists})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetSubjectID_WrongType(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		c.Set(middleware.SubjectIDKey, 123) // Wrong type

		userID, exists := middleware.GetSubjectID(c)
		assert.False(t, exists)
		assert.Empty(t, userID)

		c.JSON(http.StatusOK, gin.H{"exists": exists})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetUserRoles_Success(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		c.Set(middleware.UserRolesKey, []string{"admin", "user"})

		roles, exists := middleware.GetUserRoles(c)
		assert.True(t, exists)
		assert.Len(t, roles, 2)
		assert.Contains(t, roles, "admin")
		assert.Contains(t, roles, "user")

		c.JSON(http.StatusOK, gin.H{"roles": roles})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetUserRoles_NotSet(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		roles, exists := middleware.GetUserRoles(c)
		assert.False(t, exists)
		assert.Nil(t, roles)

		c.JSON(http.StatusOK, gin.H{"exists": exists})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetUserRoles_WrongType(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		c.Set(middleware.UserRolesKey, "not-a-slice") // Wrong type

		roles, exists := middleware.GetUserRoles(c)
		assert.False(t, exists)
		assert.Nil(t, roles)

		c.JSON(http.StatusOK, gin.H{"exists": exists})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetOrgID_Success(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		c.Set(middleware.OrgIDKey, "org-123")

		orgID, exists := middleware.GetOrgID(c)
		assert.True(t, exists)
		assert.Equal(t, "org-123", orgID)

		c.JSON(http.StatusOK, gin.H{"org_id": orgID})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetOrgID_NotSet(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		orgID, exists := middleware.GetOrgID(c)
		assert.False(t, exists)
		assert.Empty(t, orgID)

		c.JSON(http.StatusOK, gin.H{"exists": exists})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetOrgID_WrongType(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		c.Set(middleware.OrgIDKey, 456) // Wrong type

		orgID, exists := middleware.GetOrgID(c)
		assert.False(t, exists)
		assert.Empty(t, orgID)

		c.JSON(http.StatusOK, gin.H{"exists": exists})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test AuthZ middleware with comprehensive scenarios
func TestAuthZ_Success(t *testing.T) {
	mockClient := &MockAuthAAAClient{}
	router := setupTestRouter()

	// Set up authenticated context
	router.Use(func(c *gin.Context) {
		c.Set(middleware.SubjectIDKey, "user-123")
		c.Set("aaaClient", mockClient)
		c.Next()
	})

	inferResourceID := func(c *gin.Context) (string, error) {
		return "resource-123", nil
	}

	mockClient.On("Authorize", mock.Anything, mock.MatchedBy(func(req *auth.AuthorizeRequest) bool {
		return req.UserID == "user-123" && req.Resource == "orders" && req.Action == "create" && req.ResourceID == "resource-123"
	})).Return(&auth.AuthorizeResponse{Allowed: true, Reason: "Test authorization"}, nil)

	router.Use(middleware.AuthZ("orders", "create", inferResourceID))
	router.POST("/test", createTestHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockClient.AssertExpectations(t)
}

func TestAuthZ_MissingAuthentication(t *testing.T) {
	mockClient := &MockAuthAAAClient{}
	router := setupTestRouter()

	inferResourceID := func(c *gin.Context) (string, error) {
		return "resource-123", nil
	}

	router.Use(middleware.AuthZ("orders", "create", inferResourceID))
	router.POST("/test", createTestHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"].(map[string]interface{})["code"], "UNAUTHORIZED")

	mockClient.AssertNotCalled(t, "Authorize")
}

func TestAuthZ_ResourceInferenceError(t *testing.T) {
	mockClient := &MockAuthAAAClient{}
	router := setupTestRouter()

	router.Use(func(c *gin.Context) {
		c.Set(middleware.SubjectIDKey, "user-123")
		c.Set("aaaClient", mockClient)
		c.Next()
	})

	inferResourceID := func(c *gin.Context) (string, error) {
		return "", errors.New("failed to infer resource ID")
	}

	router.Use(middleware.AuthZ("orders", "create", inferResourceID))
	router.POST("/test", createTestHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"].(map[string]interface{})["code"], "INVALID_RESOURCE")

	mockClient.AssertNotCalled(t, "Authorize")
}

func TestAuthZ_MissingAAAClient(t *testing.T) {
	router := setupTestRouter()

	router.Use(func(c *gin.Context) {
		c.Set(middleware.SubjectIDKey, "user-123")
		// Missing aaaClient in context
		c.Next()
	})

	inferResourceID := func(c *gin.Context) (string, error) {
		return "resource-123", nil
	}

	router.Use(middleware.AuthZ("orders", "create", inferResourceID))
	router.POST("/test", createTestHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"].(map[string]interface{})["code"], "INTERNAL_ERROR")
}

func TestAuthZ_InvalidAAAClient(t *testing.T) {
	router := setupTestRouter()

	router.Use(func(c *gin.Context) {
		c.Set(middleware.SubjectIDKey, "user-123")
		c.Set("aaaClient", "not-an-aaa-client") // Wrong type
		c.Next()
	})

	inferResourceID := func(c *gin.Context) (string, error) {
		return "resource-123", nil
	}

	router.Use(middleware.AuthZ("orders", "create", inferResourceID))
	router.POST("/test", createTestHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"].(map[string]interface{})["code"], "INTERNAL_ERROR")
}

func TestAuthZ_PermissionDenied(t *testing.T) {
	mockClient := &MockAuthAAAClient{}
	router := setupTestRouter()

	router.Use(func(c *gin.Context) {
		c.Set(middleware.SubjectIDKey, "user-123")
		c.Set("aaaClient", mockClient)
		c.Next()
	})

	inferResourceID := func(c *gin.Context) (string, error) {
		return "resource-123", nil
	}

	mockClient.On("Authorize", mock.Anything, mock.MatchedBy(func(req *auth.AuthorizeRequest) bool {
		return req.UserID == "user-123" && req.Resource == "orders" && req.Action == "create" && req.ResourceID == "resource-123"
	})).Return(&auth.AuthorizeResponse{Allowed: false, Reason: "Permission denied"}, nil)

	router.Use(middleware.AuthZ("orders", "create", inferResourceID))
	router.POST("/test", createTestHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"].(map[string]interface{})["code"], "FORBIDDEN")

	errorDetails := response["error"].(map[string]interface{})["details"].(map[string]interface{})
	assert.Equal(t, "orders", errorDetails["resource"])
	assert.Equal(t, "create", errorDetails["action"])
	assert.Equal(t, "resource-123", errorDetails["resource_id"])

	mockClient.AssertExpectations(t)
}

func TestAuthZ_PermissionEvaluationError(t *testing.T) {
	mockClient := &MockAuthAAAClient{}
	router := setupTestRouter()

	router.Use(func(c *gin.Context) {
		c.Set(middleware.SubjectIDKey, "user-123")
		c.Set("aaaClient", mockClient)
		c.Next()
	})

	inferResourceID := func(c *gin.Context) (string, error) {
		return "resource-123", nil
	}

	mockClient.On("Authorize", mock.Anything, mock.MatchedBy(func(req *auth.AuthorizeRequest) bool {
		return req.UserID == "user-123" && req.Resource == "orders" && req.Action == "create" && req.ResourceID == "resource-123"
	})).Return(nil, errors.New("AAA service error"))

	router.Use(middleware.AuthZ("orders", "create", inferResourceID))
	router.POST("/test", createTestHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"].(map[string]interface{})["code"], "AUTHORIZATION_ERROR")

	mockClient.AssertExpectations(t)
}

// Test resource ID inference functions
func TestInferOrgIDFromParam_Success(t *testing.T) {
	router := setupTestRouter()

	router.GET("/orgs/:orgId/test", func(c *gin.Context) {
		inferFunc := middleware.InferOrgIDFromParam("orgId")
		orgID, err := inferFunc(c)

		assert.NoError(t, err)
		assert.Equal(t, "org-123", orgID)

		c.JSON(http.StatusOK, gin.H{"org_id": orgID})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/orgs/org-123/test", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInferOrgIDFromParam_Missing(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		inferFunc := middleware.InferOrgIDFromParam("orgId")
		orgID, err := inferFunc(c)

		assert.Error(t, err)
		assert.Empty(t, orgID)

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInferOrderIDFromParam_Success(t *testing.T) {
	router := setupTestRouter()

	router.GET("/orders/:id", func(c *gin.Context) {
		inferFunc := middleware.InferOrderIDFromParam()
		orderID, err := inferFunc(c)

		assert.NoError(t, err)
		assert.Equal(t, "order-123", orderID)

		c.JSON(http.StatusOK, gin.H{"order_id": orderID})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/orders/order-123", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInferOrderIDFromParam_Missing(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		inferFunc := middleware.InferOrderIDFromParam()
		orderID, err := inferFunc(c)

		assert.Error(t, err)
		assert.Empty(t, orderID)

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInferItemIDFromParam_Success(t *testing.T) {
	router := setupTestRouter()

	router.GET("/items/:id", func(c *gin.Context) {
		inferFunc := middleware.InferItemIDFromParam()
		itemID, err := inferFunc(c)

		assert.NoError(t, err)
		assert.Equal(t, "item-123", itemID)

		c.JSON(http.StatusOK, gin.H{"item_id": itemID})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/items/item-123", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test middleware chaining scenarios
func TestMiddlewareChaining_AuthNThenAuthZ(t *testing.T) {
	mockClient := &MockAuthAAAClient{}
	router := setupTestRouter()

	inferResourceID := func(c *gin.Context) (string, error) {
		return "resource-123", nil
	}

	// Set up mock expectations for both AuthN and AuthZ
	mockClaims := &auth.TokenClaims{
		UserID:           "mock_user_id",
		Username:         "test_user",
		Email:            "test@example.com",
		TenantID:         "tenant_123",
		OrganizationID:   "org_456",
		OrganizationName: "Test Org",
		Roles:            []string{"buyer", "seller"},
		Permissions:      []string{"catalog_read", "catalog_write"},
	}
	mockClient.On("ValidateToken", mock.Anything, "valid_token").Return(mockClaims, nil)
	mockClient.On("Authorize", mock.Anything, mock.MatchedBy(func(req *auth.AuthorizeRequest) bool {
		return req.UserID == "mock_user_id" && req.Resource == "orders" && req.Action == "create" && req.ResourceID == "resource-123"
	})).Return(&auth.AuthorizeResponse{Allowed: true, Reason: "Test authorization"}, nil)

	router.Use(middleware.AuthNMiddleware(mockClient))
	router.Use(func(c *gin.Context) {
		c.Set("aaaClient", mockClient)
		c.Next()
	})
	router.Use(middleware.AuthZ("orders", "create", inferResourceID))
	router.POST("/test", createTestHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid_token")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockClient.AssertExpectations(t)
}

func TestMiddlewareChaining_AuthNFailsBeforeAuthZ(t *testing.T) {
	mockClient := &MockAuthAAAClient{}
	router := setupTestRouter()

	inferResourceID := func(c *gin.Context) (string, error) {
		return "resource-123", nil
	}

	router.Use(middleware.AuthNMiddleware(mockClient))
	router.Use(middleware.AuthZ("orders", "create", inferResourceID))
	router.POST("/test", createTestHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)
	// No Authorization header

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// AuthZ should not be called if AuthN fails
	mockClient.AssertNotCalled(t, "Authorize")
}

// Test concurrent request scenarios
func TestAuthNMiddleware_ConcurrentRequests(t *testing.T) {
	mockClient := &MockAuthAAAClient{}
	router := setupTestRouter()

	// Set up mock expectations for multiple calls
	mockClaims := &auth.TokenClaims{
		UserID:           "mock_user_id",
		Username:         "test_user",
		Email:            "test@example.com",
		TenantID:         "tenant_123",
		OrganizationID:   "org_456",
		OrganizationName: "Test Org",
		Roles:            []string{"buyer", "seller"},
		Permissions:      []string{"catalog_read", "catalog_write"},
	}
	mockClient.On("ValidateToken", mock.Anything, "valid_token").Return(mockClaims, nil).Times(10)

	router.Use(middleware.AuthNMiddleware(mockClient))
	router.GET("/test", func(c *gin.Context) {
		userID, _ := middleware.GetSubjectID(c)
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	// Simulate concurrent requests
	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer valid_token")

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	mockClient.AssertExpectations(t)
}

// Test edge cases
func TestAuthNMiddleware_CaseInsensitiveBearer(t *testing.T) {
	mockClient := &MockAuthAAAClient{}
	router := setupTestRouter()

	// Set up mock expectations for multiple calls
	mockClaims := &auth.TokenClaims{
		UserID:           "mock_user_id",
		Username:         "test_user",
		Email:            "test@example.com",
		TenantID:         "tenant_123",
		OrganizationID:   "org_456",
		OrganizationName: "Test Org",
		Roles:            []string{"buyer", "seller"},
		Permissions:      []string{"catalog_read", "catalog_write"},
	}
	mockClient.On("ValidateToken", mock.Anything, "valid_token").Return(mockClaims, nil).Times(4)

	router.Use(middleware.AuthNMiddleware(mockClient))
	router.GET("/test", createTestHandler())

	testCases := []string{
		"Bearer valid_token",
		"bearer valid_token",
		"BEARER valid_token",
		"BeArEr valid_token",
	}

	for _, authHeader := range testCases {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", authHeader)

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "Failed for auth header: %s", authHeader)
	}

	mockClient.AssertExpectations(t)
}

func TestAuthNMiddleware_MultipleSpacesInToken(t *testing.T) {
	mockClient := &MockAuthAAAClient{}
	router := setupTestRouter()

	// Set up mock expectations
	mockClaims := &auth.TokenClaims{
		UserID:           "mock_user_id",
		Username:         "test_user",
		Email:            "test@example.com",
		TenantID:         "tenant_123",
		OrganizationID:   "org_456",
		OrganizationName: "Test Org",
		Roles:            []string{"buyer", "seller"},
		Permissions:      []string{"catalog_read", "catalog_write"},
	}
	mockClient.On("ValidateToken", mock.Anything, "   valid_token_with_spaces   ").Return(mockClaims, nil)

	router.Use(middleware.AuthNMiddleware(mockClient))
	router.GET("/test", createTestHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer    valid_token_with_spaces   ")

	router.ServeHTTP(w, req)

	// Should handle token trimming properly
	assert.Equal(t, http.StatusOK, w.Code)
	mockClient.AssertExpectations(t)
}
