package auth_test

import (
	"context"
	"testing"
	"time"

	"kisanlink-ecom/internal/auth"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAAAClient for testing
type MockAAAClient struct {
	mock.Mock
}

func (m *MockAAAClient) CreateUser(ctx context.Context, user *auth.AAAUser) (*auth.AAAUser, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(*auth.AAAUser), args.Error(1)
}

func (m *MockAAAClient) GetUser(ctx context.Context, userID string) (*auth.AAAUser, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*auth.AAAUser), args.Error(1)
}

func (m *MockAAAClient) UpdateUser(ctx context.Context, user *auth.AAAUser) (*auth.AAAUser, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(*auth.AAAUser), args.Error(1)
}

func (m *MockAAAClient) DeleteUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAAAClient) AuthenticateUser(ctx context.Context, username, password string) (*auth.AuthenticationResponse, error) {
	args := m.Called(ctx, username, password)
	return args.Get(0).(*auth.AuthenticationResponse), args.Error(1)
}

func (m *MockAAAClient) RefreshToken(ctx context.Context, refreshToken string) (*auth.AuthenticationResponse, error) {
	args := m.Called(ctx, refreshToken)
	return args.Get(0).(*auth.AuthenticationResponse), args.Error(1)
}

func (m *MockAAAClient) ValidateJWT(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}

func (m *MockAAAClient) GetUserFromToken(ctx context.Context, token string) (*auth.UserContext, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(*auth.UserContext), args.Error(1)
}

func (m *MockAAAClient) EvaluatePermission(ctx context.Context, userID, resource, action string) (bool, error) {
	args := m.Called(ctx, userID, resource, action)
	return args.Bool(0), args.Error(1)
}

func (m *MockAAAClient) EvaluateResourcePermission(ctx context.Context, userID, resourceType, resourceID, action string) (bool, error) {
	args := m.Called(ctx, userID, resourceType, resourceID, action)
	return args.Bool(0), args.Error(1)
}

func (m *MockAAAClient) BulkEvaluatePermissions(ctx context.Context, userID string, permissions []auth.PermissionCheck) ([]auth.PermissionResult, error) {
	args := m.Called(ctx, userID, permissions)
	return args.Get(0).([]auth.PermissionResult), args.Error(1)
}

func (m *MockAAAClient) GetUserRoles(ctx context.Context, userID string) ([]*auth.AAARole, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*auth.AAARole), args.Error(1)
}

func (m *MockAAAClient) AssignRole(ctx context.Context, userID, roleID string) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockAAAClient) RemoveRole(ctx context.Context, userID, roleID string) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockAAAClient) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockAAAClient) ValidateUserOrganization(ctx context.Context, userID, orgID string) (bool, error) {
	args := m.Called(ctx, userID, orgID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAAAClient) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAAAClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestJWTValidator_ValidateToken_Success(t *testing.T) {
	// Setup
	mockAAAClient := new(MockAAAClient)
	validator := auth.NewJWTValidator(mockAAAClient, "test-issuer", "test-audience")
	ctx := context.Background()

	// Create a test JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":         "user123",
		"username":    "testuser",
		"email":       "test@example.com",
		"org_id":      "org123",
		"org_name":    "Test Organization",
		"roles":       []string{"user", "admin"},
		"permissions": []string{"read", "write"},
		"iat":         time.Now().Unix(),
		"exp":         time.Now().Add(time.Hour).Unix(),
		"iss":         "test-issuer",
		"aud":         "test-audience",
	})

	tokenString, err := token.SignedString([]byte("secret"))
	assert.NoError(t, err)

	// Mock expectations
	mockAAAClient.On("ValidateJWT", ctx, tokenString).Return(true, nil)

	// Test
	result, err := validator.ValidateToken(ctx, tokenString)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.NotNil(t, result.Claims)
	assert.NotNil(t, result.UserContext)
	assert.Equal(t, "user123", result.UserContext.UserID)
	assert.Equal(t, "testuser", result.UserContext.Username)
	assert.Equal(t, "test@example.com", result.UserContext.Email)
	assert.Equal(t, "org123", result.UserContext.OrganizationID)
	assert.Equal(t, "Test Organization", result.UserContext.OrganizationName)
	assert.Equal(t, []string{"user", "admin"}, result.UserContext.Roles)
	assert.Equal(t, []string{"read", "write"}, result.UserContext.Permissions)

	mockAAAClient.AssertExpectations(t)
}

func TestJWTValidator_ValidateToken_EmptyToken(t *testing.T) {
	// Setup
	mockAAAClient := new(MockAAAClient)
	validator := auth.NewJWTValidator(mockAAAClient, "test-issuer", "test-audience")
	ctx := context.Background()

	// Test with empty token
	result, err := validator.ValidateToken(ctx, "")

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Error.Error(), "empty token")

	mockAAAClient.AssertExpectations(t)
}

func TestJWTValidator_ValidateToken_AAAServiceFailure(t *testing.T) {
	// Setup
	mockAAAClient := new(MockAAAClient)
	validator := auth.NewJWTValidator(mockAAAClient, "test-issuer", "test-audience")
	ctx := context.Background()

	// Create a test JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      "user123",
		"username": "testuser",
		"email":    "test@example.com",
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(time.Hour).Unix(),
		"iss":      "test-issuer",
		"aud":      "test-audience",
	})

	tokenString, err := token.SignedString([]byte("secret"))
	assert.NoError(t, err)

	// Mock expectations - AAA service failure
	mockAAAClient.On("ValidateJWT", ctx, tokenString).Return(false, assert.AnError)

	// Test
	result, err := validator.ValidateToken(ctx, tokenString)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Error.Error(), "AAA service validation failed")

	mockAAAClient.AssertExpectations(t)
}

func TestJWTValidator_ValidateToken_InvalidClaims(t *testing.T) {
	// Setup
	mockAAAClient := new(MockAAAClient)
	validator := auth.NewJWTValidator(mockAAAClient, "test-issuer", "test-audience")
	ctx := context.Background()

	// Create a token with missing subject claim
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "testuser",
		"email":    "test@example.com",
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(time.Hour).Unix(),
		"iss":      "test-issuer",
		"aud":      "test-audience",
	})

	tokenString, err := token.SignedString([]byte("secret"))
	assert.NoError(t, err)

	// Test
	result, err := validator.ValidateToken(ctx, tokenString)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Error.Error(), "failed to convert claims")

	mockAAAClient.AssertExpectations(t)
}

func TestJWTValidator_ValidateToken_WrongIssuer(t *testing.T) {
	// Setup
	mockAAAClient := new(MockAAAClient)
	validator := auth.NewJWTValidator(mockAAAClient, "expected-issuer", "test-audience")
	ctx := context.Background()

	// Create a token with wrong issuer
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      "user123",
		"username": "testuser",
		"email":    "test@example.com",
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(time.Hour).Unix(),
		"iss":      "wrong-issuer",
		"aud":      "test-audience",
	})

	tokenString, err := token.SignedString([]byte("secret"))
	assert.NoError(t, err)

	// Test
	result, err := validator.ValidateToken(ctx, tokenString)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Error.Error(), "invalid issuer")

	mockAAAClient.AssertExpectations(t)
}

func TestJWTValidator_ValidateTokenOffline_Success(t *testing.T) {
	// Setup
	mockAAAClient := new(MockAAAClient)
	validator := auth.NewJWTValidator(mockAAAClient, "test-issuer", "test-audience")

	// Create a test JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":         "user123",
		"username":    "testuser",
		"email":       "test@example.com",
		"org_id":      "org123",
		"org_name":    "Test Organization",
		"roles":       []string{"user", "admin"},
		"permissions": []string{"read", "write"},
		"iat":         time.Now().Unix(),
		"exp":         time.Now().Add(time.Hour).Unix(),
		"iss":         "test-issuer",
		"aud":         "test-audience",
	})

	tokenString, err := token.SignedString([]byte("secret"))
	assert.NoError(t, err)

	// Test (this should fail since we don't have the public key for offline validation)
	result, err := validator.ValidateTokenOffline(tokenString)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Error.Error(), "offline validation requires public key configuration")

	mockAAAClient.AssertExpectations(t)
}

func TestJWTValidator_ExtractClaimsWithoutValidation(t *testing.T) {
	// Setup
	mockAAAClient := new(MockAAAClient)
	validator := auth.NewJWTValidator(mockAAAClient, "test-issuer", "test-audience")

	// Create a test JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":         "user123",
		"username":    "testuser",
		"email":       "test@example.com",
		"org_id":      "org123",
		"org_name":    "Test Organization",
		"roles":       []string{"user", "admin"},
		"permissions": []string{"read", "write"},
		"iat":         time.Now().Unix(),
		"exp":         time.Now().Add(time.Hour).Unix(),
		"iss":         "test-issuer",
		"aud":         "test-audience",
	})

	tokenString, err := token.SignedString([]byte("secret"))
	assert.NoError(t, err)

	// Test
	claims, err := validator.ExtractClaimsWithoutValidation(tokenString)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "user123", claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "org123", claims.OrganizationID)
	assert.Equal(t, "Test Organization", claims.OrganizationName)
	assert.Equal(t, []string{"user", "admin"}, claims.Roles)
	assert.Equal(t, []string{"read", "write"}, claims.Permissions)
	assert.Equal(t, "test-issuer", claims.Issuer)
	assert.Equal(t, "test-audience", claims.Audience)

	mockAAAClient.AssertExpectations(t)
}

func TestJWTValidator_ExtractClaimsWithoutValidation_InvalidToken(t *testing.T) {
	// Setup
	mockAAAClient := new(MockAAAClient)
	validator := auth.NewJWTValidator(mockAAAClient, "test-issuer", "test-audience")

	// Test with invalid token
	claims, err := validator.ExtractClaimsWithoutValidation("invalid-token")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "failed to parse token")

	mockAAAClient.AssertExpectations(t)
}

func TestNewJWTValidator(t *testing.T) {
	// Setup
	mockAAAClient := new(MockAAAClient)

	// Test
	validator := auth.NewJWTValidator(mockAAAClient, "test-issuer", "test-audience")

	// Assertions
	assert.NotNil(t, validator)

	mockAAAClient.AssertExpectations(t)
}
