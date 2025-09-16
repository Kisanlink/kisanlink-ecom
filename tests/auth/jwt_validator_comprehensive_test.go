package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"kisanlink-ecom/internal/auth"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAAAClient for JWT validator testing
type MockAAAClient struct {
	mock.Mock
}

func (m *MockAAAClient) ValidateJWT(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}

func (m *MockAAAClient) GetUserFromToken(ctx context.Context, token string) (*auth.UserContext, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.UserContext), args.Error(1)
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
func (m *MockAAAClient) EvaluatePermission(ctx context.Context, userID, resource, action string) (bool, error) {
	return false, nil
}
func (m *MockAAAClient) EvaluateResourcePermission(ctx context.Context, userID, resourceType, resourceID, action string) (bool, error) {
	return false, nil
}
func (m *MockAAAClient) BulkEvaluatePermissions(ctx context.Context, userID string, permissions []auth.PermissionCheck) ([]auth.PermissionResult, error) {
	return nil, nil
}
func (m *MockAAAClient) ValidateUserOrganization(ctx context.Context, userID, orgID string) (bool, error) {
	return false, nil
}
func (m *MockAAAClient) HealthCheck(ctx context.Context) error {
	return nil
}
func (m *MockAAAClient) Close() error {
	return nil
}

// Test helper functions
func createTestJWTClaims() *auth.JWTClaims {
	return &auth.JWTClaims{
		UserID:           "user-123",
		Username:         "testuser",
		Email:            "test@example.com",
		OrganizationID:   "org-123",
		OrganizationName: "Test Organization",
		Roles:            []string{"user", "collaborator"},
		Permissions:      []string{"read", "write", "create_catalog"},
		IssuedAt:         time.Now(),
		ExpiresAt:        time.Now().Add(1 * time.Hour),
		Issuer:           "test-issuer",
		Audience:         "test-audience",
	}
}

func createTestUserContextJWT() *auth.UserContext {
	return &auth.UserContext{
		UserID:           "user-123",
		Username:         "testuser",
		Email:            "test@example.com",
		OrganizationID:   "org-123",
		OrganizationName: "Test Organization",
		Roles:            []string{"user", "collaborator"},
		Permissions:      []string{"read", "write", "create_catalog"},
		IsActive:         true,
	}
}

func createValidJWTToken() string {
	// Create a simple JWT token for testing
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":         "user-123",
		"username":    "testuser",
		"email":       "test@example.com",
		"org_id":      "org-123",
		"org_name":    "Test Organization",
		"roles":       []string{"user", "collaborator"},
		"permissions": []string{"read", "write", "create_catalog"},
		"iat":         time.Now().Unix(),
		"exp":         time.Now().Add(1 * time.Hour).Unix(),
		"iss":         "test-issuer",
		"aud":         "test-audience",
	})

	tokenString, _ := token.SignedString([]byte("test-secret"))
	return tokenString
}

func createExpiredJWTToken() string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      "user-123",
		"username": "testuser",
		"email":    "test@example.com",
		"iat":      time.Now().Add(-2 * time.Hour).Unix(),
		"exp":      time.Now().Add(-1 * time.Hour).Unix(), // Expired 1 hour ago
		"iss":      "test-issuer",
		"aud":      "test-audience",
	})

	tokenString, _ := token.SignedString([]byte("test-secret"))
	return tokenString
}

func createMalformedJWTToken() string {
	return "malformed.jwt.token.string"
}

// Test NewJWTValidator with comprehensive scenarios
func TestNewJWTValidator_Success(t *testing.T) {
	mockClient := &MockAAAClient{}
	validator := auth.NewJWTValidator(mockClient, "test-issuer", "test-audience")

	assert.NotNil(t, validator)
}

func TestNewJWTValidator_NilClient(t *testing.T) {
	validator := auth.NewJWTValidator(nil, "test-issuer", "test-audience")

	assert.NotNil(t, validator) // Should still create validator but will fail on validation
}

func TestNewJWTValidator_EmptyIssuerAudience(t *testing.T) {
	mockClient := &MockAAAClient{}
	validator := auth.NewJWTValidator(mockClient, "", "")

	assert.NotNil(t, validator) // Should create validator without issuer/audience validation
}

// Test ValidateToken with comprehensive scenarios
func TestJWTValidator_ValidateToken_Success(t *testing.T) {
	mockClient := &MockAAAClient{}
	validator := auth.NewJWTValidator(mockClient, "test-issuer", "test-audience")

	validToken := createValidJWTToken()
	mockClient.On("ValidateJWT", mock.Anything, validToken).Return(true, nil)

	result, err := validator.ValidateToken(context.Background(), validToken)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.Nil(t, result.Error)
	assert.NotNil(t, result.Claims)
	assert.NotNil(t, result.UserContext)
	assert.Equal(t, "user-123", result.Claims.UserID)
	assert.Equal(t, "testuser", result.Claims.Username)

	mockClient.AssertExpectations(t)
}

func TestJWTValidator_ValidateToken_EmptyToken(t *testing.T) {
	mockClient := &MockAAAClient{}
	validator := auth.NewJWTValidator(mockClient, "test-issuer", "test-audience")

	result, err := validator.ValidateToken(context.Background(), "")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Error(), "empty token")

	mockClient.AssertNotCalled(t, "ValidateJWT")
}

func TestJWTValidator_ValidateToken_WhitespaceToken(t *testing.T) {
	mockClient := &MockAAAClient{}
	validator := auth.NewJWTValidator(mockClient, "test-issuer", "test-audience")

	result, err := validator.ValidateToken(context.Background(), "   \t\n   ")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Error(), "empty token")

	mockClient.AssertNotCalled(t, "ValidateJWT")
}

func TestJWTValidator_ValidateToken_MalformedToken(t *testing.T) {
	mockClient := &MockAAAClient{}
	validator := auth.NewJWTValidator(mockClient, "test-issuer", "test-audience")

	malformedToken := createMalformedJWTToken()

	result, err := validator.ValidateToken(context.Background(), malformedToken)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Error(), "failed to parse token")

	mockClient.AssertNotCalled(t, "ValidateJWT")
}

func TestJWTValidator_ValidateToken_AAAServiceFailure(t *testing.T) {
	mockClient := &MockAAAClient{}
	validator := auth.NewJWTValidator(mockClient, "test-issuer", "test-audience")

	validToken := createValidJWTToken()
	mockClient.On("ValidateJWT", mock.Anything, validToken).Return(false, errors.New("AAA service unavailable"))

	result, err := validator.ValidateToken(context.Background(), validToken)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Error(), "AAA service validation failed")

	mockClient.AssertExpectations(t)
}

func TestJWTValidator_ValidateToken_AAAServiceRejectsToken(t *testing.T) {
	mockClient := &MockAAAClient{}
	validator := auth.NewJWTValidator(mockClient, "test-issuer", "test-audience")

	validToken := createValidJWTToken()
	mockClient.On("ValidateJWT", mock.Anything, validToken).Return(false, nil)

	result, err := validator.ValidateToken(context.Background(), validToken)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Error(), "token validation failed")

	mockClient.AssertExpectations(t)
}

func TestJWTValidator_ValidateToken_InvalidIssuer(t *testing.T) {
	mockClient := &MockAAAClient{}
	validator := auth.NewJWTValidator(mockClient, "expected-issuer", "test-audience")

	// Create token with different issuer
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      "user-123",
		"username": "testuser",
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
		"iss":      "wrong-issuer", // Different issuer
		"aud":      "test-audience",
	})
	tokenString, _ := token.SignedString([]byte("test-secret"))

	result, err := validator.ValidateToken(context.Background(), tokenString)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Error(), "invalid issuer")

	mockClient.AssertNotCalled(t, "ValidateJWT")
}

func TestJWTValidator_ValidateToken_InvalidAudience(t *testing.T) {
	mockClient := &MockAAAClient{}
	validator := auth.NewJWTValidator(mockClient, "test-issuer", "expected-audience")

	// Create token with different audience
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      "user-123",
		"username": "testuser",
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
		"iss":      "test-issuer",
		"aud":      "wrong-audience", // Different audience
	})
	tokenString, _ := token.SignedString([]byte("test-secret"))

	result, err := validator.ValidateToken(context.Background(), tokenString)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Error(), "invalid audience")

	mockClient.AssertNotCalled(t, "ValidateJWT")
}

func TestJWTValidator_ValidateToken_MissingSubject(t *testing.T) {
	mockClient := &MockAAAClient{}
	validator := auth.NewJWTValidator(mockClient, "test-issuer", "test-audience")

	// Create token without subject claim
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "testuser",
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
		"iss":      "test-issuer",
		"aud":      "test-audience",
		// Missing "sub" claim
	})
	tokenString, _ := token.SignedString([]byte("test-secret"))

	result, err := validator.ValidateToken(context.Background(), tokenString)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Error(), "missing or invalid subject claim")

	mockClient.AssertNotCalled(t, "ValidateJWT")
}

func TestJWTValidator_ValidateToken_MissingExpiration(t *testing.T) {
	mockClient := &MockAAAClient{}
	validator := auth.NewJWTValidator(mockClient, "test-issuer", "test-audience")

	// Create token without expiration claim
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      "user-123",
		"username": "testuser",
		"iat":      time.Now().Unix(),
		"iss":      "test-issuer",
		"aud":      "test-audience",
		// Missing "exp" claim
	})
	tokenString, _ := token.SignedString([]byte("test-secret"))

	result, err := validator.ValidateToken(context.Background(), tokenString)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Error(), "missing or invalid expiration claim")

	mockClient.AssertNotCalled(t, "ValidateJWT")
}

// Test ValidateTokenOffline with comprehensive scenarios
func TestJWTValidator_ValidateTokenOffline_NotImplemented(t *testing.T) {
	mockClient := &MockAAAClient{}
	validator := auth.NewJWTValidator(mockClient, "test-issuer", "test-audience")

	validToken := createValidJWTToken()

	result, err := validator.ValidateTokenOffline(validToken)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Error(), "offline validation requires public key configuration")
}

func TestJWTValidator_ValidateTokenOffline_MalformedToken(t *testing.T) {
	mockClient := &MockAAAClient{}
	validator := auth.NewJWTValidator(mockClient, "test-issuer", "test-audience")

	malformedToken := createMalformedJWTToken()

	result, err := validator.ValidateTokenOffline(malformedToken)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Error(), "offline token validation failed")
}

// Test ExtractClaimsWithoutValidation with comprehensive scenarios
func TestJWTValidator_ExtractClaimsWithoutValidation_Success(t *testing.T) {
	mockClient := &MockAAAClient{}
	validator := auth.NewJWTValidator(mockClient, "test-issuer", "test-audience")

	validToken := createValidJWTToken()

	claims, err := validator.ExtractClaimsWithoutValidation(validToken)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "org-123", claims.OrganizationID)
	assert.Equal(t, "Test Organization", claims.OrganizationName)
	assert.Contains(t, claims.Roles, "user")
	assert.Contains(t, claims.Roles, "collaborator")
	assert.Contains(t, claims.Permissions, "read")
	assert.Contains(t, claims.Permissions, "write")
}

func TestJWTValidator_ExtractClaimsWithoutValidation_MalformedToken(t *testing.T) {
	mockClient := &MockAAAClient{}
	validator := auth.NewJWTValidator(mockClient, "test-issuer", "test-audience")

	malformedToken := createMalformedJWTToken()

	claims, err := validator.ExtractClaimsWithoutValidation(malformedToken)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "failed to parse token")
}

func TestJWTValidator_ExtractClaimsWithoutValidation_EmptyToken(t *testing.T) {
	mockClient := &MockAAAClient{}
	validator := auth.NewJWTValidator(mockClient, "test-issuer", "test-audience")

	claims, err := validator.ExtractClaimsWithoutValidation("")

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "failed to parse token")
}

// Test claims conversion scenarios
func TestJWTValidator_ConvertClaims_SingleRole(t *testing.T) {
	mockClient := &MockAAAClient{}
	validator := auth.NewJWTValidator(mockClient, "test-issuer", "test-audience")

	// Create token with single role as string instead of array
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      "user-123",
		"username": "testuser",
		"roles":    "admin", // Single role as string
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
		"iss":      "test-issuer",
		"aud":      "test-audience",
	})
	tokenString, _ := token.SignedString([]byte("test-secret"))

	claims, err := validator.ExtractClaimsWithoutValidation(tokenString)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Len(t, claims.Roles, 1)
	assert.Equal(t, "admin", claims.Roles[0])
}

func TestJWTValidator_ConvertClaims_NumericClaims(t *testing.T) {
	mockClient := &MockAAAClient{}
	validator := auth.NewJWTValidator(mockClient, "test-issuer", "test-audience")

	now := time.Now()
	exp := now.Add(1 * time.Hour)

	// Create token with numeric timestamp claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      "user-123",
		"username": "testuser",
		"iat":      float64(now.Unix()), // Numeric timestamp
		"exp":      float64(exp.Unix()), // Numeric timestamp
		"iss":      "test-issuer",
		"aud":      "test-audience",
	})
	tokenString, _ := token.SignedString([]byte("test-secret"))

	claims, err := validator.ExtractClaimsWithoutValidation(tokenString)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.WithinDuration(t, now, claims.IssuedAt, 1*time.Second)
	assert.WithinDuration(t, exp, claims.ExpiresAt, 1*time.Second)
}

func TestJWTValidator_ConvertClaims_OptionalFields(t *testing.T) {
	mockClient := &MockAAAClient{}
	validator := auth.NewJWTValidator(mockClient, "", "") // No issuer/audience validation

	// Create token with minimal required claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user-123",
		"exp": float64(time.Now().Add(1 * time.Hour).Unix()),
		// Missing optional fields: username, email, org_id, etc.
	})
	tokenString, _ := token.SignedString([]byte("test-secret"))

	claims, err := validator.ExtractClaimsWithoutValidation(tokenString)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Empty(t, claims.Username)
	assert.Empty(t, claims.Email)
	assert.Empty(t, claims.OrganizationID)
	assert.Empty(t, claims.Roles)
	assert.Empty(t, claims.Permissions)
}

// Test context cancellation scenarios
func TestJWTValidator_ValidateToken_ContextCancellation(t *testing.T) {
	mockClient := &MockAAAClient{}
	validator := auth.NewJWTValidator(mockClient, "test-issuer", "test-audience")

	validToken := createValidJWTToken()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockClient.On("ValidateJWT", mock.Anything, validToken).Return(false, errors.New("context canceled"))

	result, err := validator.ValidateToken(ctx, validToken)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Error(), "AAA service validation failed")

	mockClient.AssertExpectations(t)
}

func TestJWTValidator_ValidateToken_ContextTimeout(t *testing.T) {
	mockClient := &MockAAAClient{}
	validator := auth.NewJWTValidator(mockClient, "test-issuer", "test-audience")

	validToken := createValidJWTToken()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(1 * time.Millisecond) // Ensure timeout

	mockClient.On("ValidateJWT", mock.Anything, validToken).Return(false, errors.New("context deadline exceeded"))

	result, err := validator.ValidateToken(ctx, validToken)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Error(), "AAA service validation failed")

	mockClient.AssertExpectations(t)
}

// Test clock skew scenarios
func TestJWTValidator_ValidateTokenOffline_ClockSkew(t *testing.T) {
	mockClient := &MockAAAClient{}
	validator := auth.NewJWTValidator(mockClient, "test-issuer", "test-audience")

	// Create token that's slightly expired but within clock skew tolerance
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      "user-123",
		"username": "testuser",
		"iat":      time.Now().Add(-1 * time.Hour).Unix(),
		"exp":      time.Now().Add(-2 * time.Minute).Unix(), // Expired 2 minutes ago (within 5 min clock skew)
		"iss":      "test-issuer",
		"aud":      "test-audience",
	})
	tokenString, _ := token.SignedString([]byte("test-secret"))

	result, err := validator.ValidateTokenOffline(tokenString)

	// Should still fail due to missing public key configuration
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Error.Error(), "offline validation requires public key configuration")
}

// Test edge cases with malformed claims
func TestJWTValidator_ConvertClaims_InvalidClaimsStructure(t *testing.T) {
	mockClient := &MockAAAClient{}
	validator := auth.NewJWTValidator(mockClient, "test-issuer", "test-audience")

	// Create token with invalid claims structure
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":         123,                                     // Should be string
		"exp":         "not-a-number",                          // Should be numeric
		"roles":       map[string]interface{}{"invalid": true}, // Should be array or string
		"permissions": 12345,                                   // Should be array
	})
	tokenString, _ := token.SignedString([]byte("test-secret"))

	claims, err := validator.ExtractClaimsWithoutValidation(tokenString)

	assert.Error(t, err)
	assert.Nil(t, claims)
	// Should fail due to invalid subject claim type
}

// Test performance with large tokens
func TestJWTValidator_ValidateToken_LargeToken(t *testing.T) {
	mockClient := &MockAAAClient{}
	validator := auth.NewJWTValidator(mockClient, "test-issuer", "test-audience")

	// Create token with large claims
	largeRoles := make([]string, 100)
	largePermissions := make([]string, 200)
	for i := 0; i < 100; i++ {
		largeRoles[i] = "role_" + string(rune(i))
	}
	for i := 0; i < 200; i++ {
		largePermissions[i] = "permission_" + string(rune(i))
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":         "user-123",
		"username":    "testuser",
		"roles":       largeRoles,
		"permissions": largePermissions,
		"iat":         time.Now().Unix(),
		"exp":         time.Now().Add(1 * time.Hour).Unix(),
		"iss":         "test-issuer",
		"aud":         "test-audience",
	})
	tokenString, _ := token.SignedString([]byte("test-secret"))

	mockClient.On("ValidateJWT", mock.Anything, tokenString).Return(true, nil)

	start := time.Now()
	result, err := validator.ValidateToken(context.Background(), tokenString)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.Len(t, result.Claims.Roles, 100)
	assert.Len(t, result.Claims.Permissions, 200)
	assert.Less(t, duration, 100*time.Millisecond) // Should be fast

	mockClient.AssertExpectations(t)
}
