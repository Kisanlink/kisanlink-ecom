package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"kisanlink-ecom/internal/auth"
	"kisanlink-ecom/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAAAClient for middleware testing
type MockAAAClient struct {
	mock.Mock
}

// Implement required methods for testing
func (m *MockAAAClient) ValidateJWT(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}

func (m *MockAAAClient) EvaluatePermission(ctx context.Context, userID, resource, action string) (bool, error) {
	args := m.Called(ctx, userID, resource, action)
	return args.Bool(0), args.Error(1)
}

// Implement other required methods as no-ops for testing
func (m *MockAAAClient) CreateUser(ctx context.Context, user *auth.AAAUser) (*auth.AAAUser, error) {
	return nil, nil
}
func (m *MockAAAClient) GetUser(ctx context.Context, userID string) (*auth.AAAUser, error) {
	return nil, nil
}
func (m *MockAAAClient) UpdateUser(ctx context.Context, user *auth.AAAUser) (*auth.AAAUser, error) {
	return nil, nil
}
func (m *MockAAAClient) DeleteUser(ctx context.Context, userID string) error {
	return nil
}
func (m *MockAAAClient) AuthenticateUser(ctx context.Context, username, password string) (*auth.AuthenticationResponse, error) {
	return nil, nil
}
func (m *MockAAAClient) RefreshToken(ctx context.Context, refreshToken string) (*auth.AuthenticationResponse, error) {
	return nil, nil
}
func (m *MockAAAClient) GetUserFromToken(ctx context.Context, token string) (*auth.UserContext, error) {
	return nil, nil
}
func (m *MockAAAClient) GetUserRoles(ctx context.Context, userID string) ([]*auth.AAARole, error) {
	return nil, nil
}
func (m *MockAAAClient) AssignRole(ctx context.Context, userID, roleID string) error {
	return nil
}
func (m *MockAAAClient) RemoveRole(ctx context.Context, userID, roleID string) error {
	return nil
}
func (m *MockAAAClient) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	return nil, nil
}
func (m *MockAAAClient) EvaluateResourcePermission(ctx context.Context, userID, resourceType, resourceID, action string) (bool, error) {
	return false, nil
}
func (m *MockAAAClient) BulkEvaluatePermissions(ctx context.Context, userID string, permissions []auth.PermissionCheck) ([]auth.PermissionResult, error) {
	return nil, nil
}
func (m *MockAAAClient) ValidateUserOrganization(ctx context.Context, userID, orgID string) (bool, error) {
	args := m.Called(ctx, userID, orgID)
	return args.Bool(0), args.Error(1)
}
func (m *MockAAAClient) HealthCheck(ctx context.Context) error {
	return nil
}
func (m *MockAAAClient) Close() error {
	return nil
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
	mockClient := &MockAAAClient{}
	router := setupTestRouter()

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
}

func TestAuthNMiddleware_MissingAuthorizationHeader(t *testing.T) {
	mockClient := &MockAAAClient{}
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
	mockClient := &MockAAAClient{}
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
	mockClient := &MockAAAClient{}
	router := setupTestRouter()

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
}

func TestAuthNMiddleware_WithOrgHeader(t *testing.T) {
	mockClient := &MockAAAClient{}
	router := setupTestRouter()

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
	mockClient := &MockAAAClient{}
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

	mockClient.On("EvaluatePermission", mock.Anything, "user-123", "orders", "create").Return(true, nil)

	router.Use(middleware.AuthZ("orders", "create", inferResourceID))
	router.POST("/test", createTestHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockClient.AssertExpectations(t)
}

func TestAuthZ_MissingAuthentication(t *testing.T) {
	mockClient := &MockAAAClient{}
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

	mockClient.AssertNotCalled(t, "EvaluatePermission")
}

func TestAuthZ_ResourceInferenceError(t *testing.T) {
	mockClient := &MockAAAClient{}
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

	mockClient.AssertNotCalled(t, "EvaluatePermission")
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
	mockClient := &MockAAAClient{}
	router := setupTestRouter()

	router.Use(func(c *gin.Context) {
		c.Set(middleware.SubjectIDKey, "user-123")
		c.Set("aaaClient", mockClient)
		c.Next()
	})

	inferResourceID := func(c *gin.Context) (string, error) {
		return "resource-123", nil
	}

	mockClient.On("EvaluatePermission", mock.Anything, "user-123", "orders", "create").Return(false, nil)

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
	mockClient := &MockAAAClient{}
	router := setupTestRouter()

	router.Use(func(c *gin.Context) {
		c.Set(middleware.SubjectIDKey, "user-123")
		c.Set("aaaClient", mockClient)
		c.Next()
	})

	inferResourceID := func(c *gin.Context) (string, error) {
		return "resource-123", nil
	}

	mockClient.On("EvaluatePermission", mock.Anything, "user-123", "orders", "create").Return(false, errors.New("AAA service error"))

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
	mockClient := &MockAAAClient{}
	router := setupTestRouter()

	inferResourceID := func(c *gin.Context) (string, error) {
		return "resource-123", nil
	}

	mockClient.On("EvaluatePermission", mock.Anything, "mock_user_id", "orders", "create").Return(true, nil)

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
	mockClient := &MockAAAClient{}
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
	mockClient.AssertNotCalled(t, "EvaluatePermission")
}

// Test concurrent request scenarios
func TestAuthNMiddleware_ConcurrentRequests(t *testing.T) {
	mockClient := &MockAAAClient{}
	router := setupTestRouter()

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
}

// Test edge cases
func TestAuthNMiddleware_CaseInsensitiveBearer(t *testing.T) {
	mockClient := &MockAAAClient{}
	router := setupTestRouter()

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
}

func TestAuthNMiddleware_MultipleSpacesInToken(t *testing.T) {
	mockClient := &MockAAAClient{}
	router := setupTestRouter()

	router.Use(middleware.AuthNMiddleware(mockClient))
	router.GET("/test", createTestHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer    valid_token_with_spaces   ")

	router.ServeHTTP(w, req)

	// Should handle token trimming properly
	assert.Equal(t, http.StatusOK, w.Code)
}
