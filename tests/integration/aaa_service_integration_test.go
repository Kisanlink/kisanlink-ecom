package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"kisanlink-ecom/internal/auth"
	"kisanlink-ecom/tests/testutils"

	"github.com/stretchr/testify/suite"
)

// AAAServiceIntegrationTestSuite tests AAA service integration with mock service implementation
type AAAServiceIntegrationTestSuite struct {
	suite.Suite
	mockAAAService *MockAAAService
	aaaClient      auth.AAAClient
	testUsers      map[string]*auth.AAAUser
	testTokens     map[string]string
}

// MockAAAService simulates the AAA service for integration testing
type MockAAAService struct {
	users         map[string]*auth.AAAUser
	tokens        map[string]*auth.UserContext
	permissions   map[string]map[string]bool // userID -> resource:action -> allowed
	organizations map[string][]string        // userID -> orgIDs
	isHealthy     bool
	responseDelay time.Duration
}

func NewMockAAAService() *MockAAAService {
	return &MockAAAService{
		users:         make(map[string]*auth.AAAUser),
		tokens:        make(map[string]*auth.UserContext),
		permissions:   make(map[string]map[string]bool),
		organizations: make(map[string][]string),
		isHealthy:     true,
		responseDelay: 0,
	}
}

func (m *MockAAAService) SetHealthy(healthy bool) {
	m.isHealthy = healthy
}

func (m *MockAAAService) SetResponseDelay(delay time.Duration) {
	m.responseDelay = delay
}

func (m *MockAAAService) AddUser(user *auth.AAAUser) {
	m.users[user.ID] = user
}

func (m *MockAAAService) AddToken(token string, userContext *auth.UserContext) {
	m.tokens[token] = userContext
}

func (m *MockAAAService) SetPermission(userID, resource, action string, allowed bool) {
	if m.permissions[userID] == nil {
		m.permissions[userID] = make(map[string]bool)
	}
	m.permissions[userID][resource+":"+action] = allowed
}

func (m *MockAAAService) AddUserToOrganization(userID, orgID string) {
	m.organizations[userID] = append(m.organizations[userID], orgID)
}

func (m *MockAAAService) simulateDelay() {
	if m.responseDelay > 0 {
		time.Sleep(m.responseDelay)
	}
}

func (m *MockAAAService) checkHealth() error {
	if !m.isHealthy {
		return errors.New("AAA service is unhealthy")
	}
	return nil
}

// MockAAAClient implements AAAClient interface for testing
type MockAAAClient struct {
	service *MockAAAService
}

func NewMockAAAClientWithService(service *MockAAAService) *MockAAAClient {
	return &MockAAAClient{service: service}
}

func (c *MockAAAClient) CreateUser(ctx context.Context, user *auth.AAAUser) (*auth.AAAUser, error) {
	c.service.simulateDelay()
	if err := c.service.checkHealth(); err != nil {
		return nil, err
	}

	if _, exists := c.service.users[user.ID]; exists {
		return nil, errors.New("user already exists")
	}

	c.service.AddUser(user)
	return user, nil
}

func (c *MockAAAClient) GetUser(ctx context.Context, userID string) (*auth.AAAUser, error) {
	c.service.simulateDelay()
	if err := c.service.checkHealth(); err != nil {
		return nil, err
	}

	if user, exists := c.service.users[userID]; exists {
		return user, nil
	}
	return nil, errors.New("user not found")
}

func (c *MockAAAClient) UpdateUser(ctx context.Context, user *auth.AAAUser) (*auth.AAAUser, error) {
	c.service.simulateDelay()
	if err := c.service.checkHealth(); err != nil {
		return nil, err
	}

	if _, exists := c.service.users[user.ID]; !exists {
		return nil, errors.New("user not found")
	}

	c.service.users[user.ID] = user
	return user, nil
}

func (c *MockAAAClient) DeleteUser(ctx context.Context, userID string) error {
	c.service.simulateDelay()
	if err := c.service.checkHealth(); err != nil {
		return err
	}

	if _, exists := c.service.users[userID]; !exists {
		return errors.New("user not found")
	}

	delete(c.service.users, userID)
	return nil
}

func (c *MockAAAClient) AuthenticateUser(ctx context.Context, username, password string) (*auth.AuthenticationResponse, error) {
	c.service.simulateDelay()
	if err := c.service.checkHealth(); err != nil {
		return nil, err
	}

	// Find user by username
	for _, user := range c.service.users {
		if user.Username == username && user.IsActive {
			// Simulate password check (in real implementation, would verify password)
			if password == "correct_password" {
				token := "token_" + user.ID + "_" + time.Now().Format("20060102150405")
				userContext := &auth.UserContext{
					UserID:   user.ID,
					Username: user.Username,
					Email:    user.Email,
					IsActive: user.IsActive,
				}
				c.service.AddToken(token, userContext)

				return &auth.AuthenticationResponse{
					AccessToken:  token,
					RefreshToken: "refresh_" + token,
					ExpiresIn:    3600,
					TokenType:    "Bearer",
					User:         user,
					UserContext:  userContext,
				}, nil
			}
		}
	}

	return nil, errors.New("invalid credentials")
}

func (c *MockAAAClient) RefreshToken(ctx context.Context, refreshToken string) (*auth.AuthenticationResponse, error) {
	c.service.simulateDelay()
	if err := c.service.checkHealth(); err != nil {
		return nil, err
	}

	// Simulate token refresh
	if refreshToken != "" {
		newToken := "refreshed_" + refreshToken + "_" + time.Now().Format("20060102150405")
		userContext := testutils.CreateTestUserContext()
		c.service.AddToken(newToken, userContext)

		return &auth.AuthenticationResponse{
			AccessToken:  newToken,
			RefreshToken: "new_refresh_" + newToken,
			ExpiresIn:    3600,
			TokenType:    "Bearer",
			User:         testutils.CreateTestAAAUser(),
			UserContext:  userContext,
		}, nil
	}

	return nil, errors.New("invalid refresh token")
}

func (c *MockAAAClient) ValidateJWT(ctx context.Context, token string) (bool, error) {
	c.service.simulateDelay()
	if err := c.service.checkHealth(); err != nil {
		return false, err
	}

	_, exists := c.service.tokens[token]
	return exists, nil
}

func (c *MockAAAClient) GetUserFromToken(ctx context.Context, token string) (*auth.UserContext, error) {
	c.service.simulateDelay()
	if err := c.service.checkHealth(); err != nil {
		return nil, err
	}

	if userContext, exists := c.service.tokens[token]; exists {
		return userContext, nil
	}
	return nil, errors.New("invalid token")
}

func (c *MockAAAClient) EvaluatePermission(ctx context.Context, userID, resource, action string) (bool, error) {
	c.service.simulateDelay()
	if err := c.service.checkHealth(); err != nil {
		return false, err
	}

	if userPerms, exists := c.service.permissions[userID]; exists {
		if allowed, exists := userPerms[resource+":"+action]; exists {
			return allowed, nil
		}
	}
	return false, nil
}

func (c *MockAAAClient) EvaluateResourcePermission(ctx context.Context, userID, resourceType, resourceID, action string) (bool, error) {
	return c.EvaluatePermission(ctx, userID, resourceType, action)
}

func (c *MockAAAClient) ValidateUserOrganization(ctx context.Context, userID, orgID string) (bool, error) {
	c.service.simulateDelay()
	if err := c.service.checkHealth(); err != nil {
		return false, err
	}

	if orgs, exists := c.service.organizations[userID]; exists {
		for _, org := range orgs {
			if org == orgID {
				return true, nil
			}
		}
	}
	return false, nil
}

func (c *MockAAAClient) BulkEvaluatePermissions(ctx context.Context, userID string, permissions []auth.PermissionCheck) ([]auth.PermissionResult, error) {
	c.service.simulateDelay()
	if err := c.service.checkHealth(); err != nil {
		return nil, err
	}

	results := make([]auth.PermissionResult, len(permissions))
	for i, perm := range permissions {
		allowed, _ := c.EvaluatePermission(ctx, userID, perm.Resource, perm.Action)
		results[i] = auth.PermissionResult{
			Resource: perm.Resource,
			Action:   perm.Action,
			Allowed:  allowed,
			Reason:   "Mock evaluation",
		}
	}
	return results, nil
}

func (c *MockAAAClient) HealthCheck(ctx context.Context) error {
	c.service.simulateDelay()
	return c.service.checkHealth()
}

// Implement other required methods as no-ops
func (c *MockAAAClient) GetUserRoles(ctx context.Context, userID string) ([]*auth.AAARole, error) {
	return []*auth.AAARole{}, nil
}
func (c *MockAAAClient) AssignRole(ctx context.Context, userID, roleID string) error { return nil }
func (c *MockAAAClient) RemoveRole(ctx context.Context, userID, roleID string) error { return nil }
func (c *MockAAAClient) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	return []string{}, nil
}
func (c *MockAAAClient) Close() error { return nil }

// SetupSuite initializes the test suite
func (suite *AAAServiceIntegrationTestSuite) SetupSuite() {
	// Initialize mock AAA service
	suite.mockAAAService = NewMockAAAService()
	suite.aaaClient = NewMockAAAClientWithService(suite.mockAAAService)

	// Initialize test data
	suite.testUsers = make(map[string]*auth.AAAUser)
	suite.testTokens = make(map[string]string)

	// Setup test users and permissions
	suite.setupTestData()
}

func (suite *AAAServiceIntegrationTestSuite) setupTestData() {
	// Create test users
	testUser1 := testutils.CreateTestAAAUser()
	testUser1.ID = "test-user-1"
	testUser1.Username = "testuser1"
	testUser1.Email = "testuser1@example.com"

	testUser2 := testutils.CreateTestAAAUser()
	testUser2.ID = "test-user-2"
	testUser2.Username = "testuser2"
	testUser2.Email = "testuser2@example.com"

	suite.testUsers["test-user-1"] = testUser1
	suite.testUsers["test-user-2"] = testUser2

	// Add users to mock service
	suite.mockAAAService.AddUser(testUser1)
	suite.mockAAAService.AddUser(testUser2)

	// Setup permissions
	suite.mockAAAService.SetPermission("test-user-1", "orders", "create", true)
	suite.mockAAAService.SetPermission("test-user-1", "orders", "read", true)
	suite.mockAAAService.SetPermission("test-user-1", "catalog", "read", true)

	suite.mockAAAService.SetPermission("test-user-2", "catalog", "create", true)
	suite.mockAAAService.SetPermission("test-user-2", "catalog", "update", true)
	suite.mockAAAService.SetPermission("test-user-2", "orders", "read", true)

	// Setup organizations
	suite.mockAAAService.AddUserToOrganization("test-user-1", testutils.TestOrgID)
	suite.mockAAAService.AddUserToOrganization("test-user-2", testutils.TestSellerOrgID)
}

// TearDownSuite cleans up after all tests
func (suite *AAAServiceIntegrationTestSuite) TearDownSuite() {
	if suite.aaaClient != nil {
		suite.aaaClient.Close()
	}
}

// SetupTest runs before each test
func (suite *AAAServiceIntegrationTestSuite) SetupTest() {
	// Reset service health
	suite.mockAAAService.SetHealthy(true)
	suite.mockAAAService.SetResponseDelay(0)
}

// Test user management operations
func (suite *AAAServiceIntegrationTestSuite) TestUserManagementOperations() {
	ctx := context.Background()

	suite.T().Log("Testing user creation")
	suite.testUserCreation(ctx)

	suite.T().Log("Testing user retrieval")
	suite.testUserRetrieval(ctx)

	suite.T().Log("Testing user update")
	suite.testUserUpdate(ctx)

	suite.T().Log("Testing user deletion")
	suite.testUserDeletion(ctx)
}

func (suite *AAAServiceIntegrationTestSuite) testUserCreation(ctx context.Context) {
	newUser := &auth.AAAUser{
		ID:       "new-user-123",
		Username: "newuser",
		Email:    "newuser@example.com",
		IsActive: true,
	}

	createdUser, err := suite.aaaClient.CreateUser(ctx, newUser)
	suite.NoError(err)
	suite.NotNil(createdUser)
	suite.Equal(newUser.ID, createdUser.ID)
	suite.Equal(newUser.Username, createdUser.Username)
	suite.Equal(newUser.Email, createdUser.Email)

	// Test duplicate user creation
	duplicateUser, err := suite.aaaClient.CreateUser(ctx, newUser)
	suite.Error(err)
	suite.Nil(duplicateUser)
	suite.Contains(err.Error(), "already exists")
}

func (suite *AAAServiceIntegrationTestSuite) testUserRetrieval(ctx context.Context) {
	// Test existing user
	user, err := suite.aaaClient.GetUser(ctx, "test-user-1")
	suite.NoError(err)
	suite.NotNil(user)
	suite.Equal("test-user-1", user.ID)
	suite.Equal("testuser1", user.Username)

	// Test non-existent user
	nonExistentUser, err := suite.aaaClient.GetUser(ctx, "non-existent")
	suite.Error(err)
	suite.Nil(nonExistentUser)
	suite.Contains(err.Error(), "not found")
}

func (suite *AAAServiceIntegrationTestSuite) testUserUpdate(ctx context.Context) {
	// Get existing user
	user, err := suite.aaaClient.GetUser(ctx, "test-user-1")
	suite.NoError(err)

	// Update user
	user.Email = "updated@example.com"
	updatedUser, err := suite.aaaClient.UpdateUser(ctx, user)
	suite.NoError(err)
	suite.NotNil(updatedUser)
	suite.Equal("updated@example.com", updatedUser.Email)

	// Verify update
	retrievedUser, err := suite.aaaClient.GetUser(ctx, "test-user-1")
	suite.NoError(err)
	suite.Equal("updated@example.com", retrievedUser.Email)

	// Test updating non-existent user
	nonExistentUser := &auth.AAAUser{ID: "non-existent"}
	_, err = suite.aaaClient.UpdateUser(ctx, nonExistentUser)
	suite.Error(err)
	suite.Contains(err.Error(), "not found")
}

func (suite *AAAServiceIntegrationTestSuite) testUserDeletion(ctx context.Context) {
	// Create user to delete
	userToDelete := &auth.AAAUser{
		ID:       "user-to-delete",
		Username: "deleteuser",
		Email:    "delete@example.com",
		IsActive: true,
	}

	_, err := suite.aaaClient.CreateUser(ctx, userToDelete)
	suite.NoError(err)

	// Delete user
	err = suite.aaaClient.DeleteUser(ctx, userToDelete.ID)
	suite.NoError(err)

	// Verify deletion
	deletedUser, err := suite.aaaClient.GetUser(ctx, userToDelete.ID)
	suite.Error(err)
	suite.Nil(deletedUser)

	// Test deleting non-existent user
	err = suite.aaaClient.DeleteUser(ctx, "non-existent")
	suite.Error(err)
	suite.Contains(err.Error(), "not found")
}

// Test authentication operations
func (suite *AAAServiceIntegrationTestSuite) TestAuthenticationOperations() {
	ctx := context.Background()

	suite.T().Log("Testing user authentication")
	suite.testUserAuthentication(ctx)

	suite.T().Log("Testing token validation")
	suite.testTokenValidation(ctx)

	suite.T().Log("Testing token refresh")
	suite.testTokenRefresh(ctx)

	suite.T().Log("Testing user context from token")
	suite.testUserContextFromToken(ctx)
}

func (suite *AAAServiceIntegrationTestSuite) testUserAuthentication(ctx context.Context) {
	// Test successful authentication
	authResponse, err := suite.aaaClient.AuthenticateUser(ctx, "testuser1", "correct_password")
	suite.NoError(err)
	suite.NotNil(authResponse)
	suite.NotEmpty(authResponse.AccessToken)
	suite.NotEmpty(authResponse.RefreshToken)
	suite.Equal("Bearer", authResponse.TokenType)
	suite.Greater(authResponse.ExpiresIn, int64(0))
	suite.NotNil(authResponse.User)
	suite.NotNil(authResponse.UserContext)

	// Store token for later tests
	suite.testTokens["testuser1"] = authResponse.AccessToken

	// Test failed authentication
	failedAuth, err := suite.aaaClient.AuthenticateUser(ctx, "testuser1", "wrong_password")
	suite.Error(err)
	suite.Nil(failedAuth)
	suite.Contains(err.Error(), "invalid credentials")

	// Test authentication with non-existent user
	nonExistentAuth, err := suite.aaaClient.AuthenticateUser(ctx, "nonexistent", "password")
	suite.Error(err)
	suite.Nil(nonExistentAuth)
}

func (suite *AAAServiceIntegrationTestSuite) testTokenValidation(ctx context.Context) {
	// Get valid token
	token := suite.testTokens["testuser1"]
	suite.NotEmpty(token)

	// Test valid token
	isValid, err := suite.aaaClient.ValidateJWT(ctx, token)
	suite.NoError(err)
	suite.True(isValid)

	// Test invalid token
	isValid, err = suite.aaaClient.ValidateJWT(ctx, "invalid_token")
	suite.NoError(err)
	suite.False(isValid)

	// Test empty token
	isValid, err = suite.aaaClient.ValidateJWT(ctx, "")
	suite.NoError(err)
	suite.False(isValid)
}

func (suite *AAAServiceIntegrationTestSuite) testTokenRefresh(ctx context.Context) {
	// Test successful token refresh
	refreshResponse, err := suite.aaaClient.RefreshToken(ctx, "valid_refresh_token")
	suite.NoError(err)
	suite.NotNil(refreshResponse)
	suite.NotEmpty(refreshResponse.AccessToken)
	suite.NotEmpty(refreshResponse.RefreshToken)
	suite.Contains(refreshResponse.AccessToken, "refreshed_")

	// Test failed token refresh
	failedRefresh, err := suite.aaaClient.RefreshToken(ctx, "")
	suite.Error(err)
	suite.Nil(failedRefresh)
	suite.Contains(err.Error(), "invalid refresh token")
}

func (suite *AAAServiceIntegrationTestSuite) testUserContextFromToken(ctx context.Context) {
	// Get valid token
	token := suite.testTokens["testuser1"]
	suite.NotEmpty(token)

	// Test getting user context from valid token
	userContext, err := suite.aaaClient.GetUserFromToken(ctx, token)
	suite.NoError(err)
	suite.NotNil(userContext)
	suite.NotEmpty(userContext.UserID)
	suite.NotEmpty(userContext.Username)
	suite.True(userContext.IsActive)

	// Test getting user context from invalid token
	invalidContext, err := suite.aaaClient.GetUserFromToken(ctx, "invalid_token")
	suite.Error(err)
	suite.Nil(invalidContext)
	suite.Contains(err.Error(), "invalid token")
}

// Test authorization operations
func (suite *AAAServiceIntegrationTestSuite) TestAuthorizationOperations() {
	ctx := context.Background()

	suite.T().Log("Testing permission evaluation")
	suite.testPermissionEvaluation(ctx)

	suite.T().Log("Testing bulk permission evaluation")
	suite.testBulkPermissionEvaluation(ctx)

	suite.T().Log("Testing organization validation")
	suite.testOrganizationValidation(ctx)
}

func (suite *AAAServiceIntegrationTestSuite) testPermissionEvaluation(ctx context.Context) {
	// Test allowed permission
	allowed, err := suite.aaaClient.EvaluatePermission(ctx, "test-user-1", "orders", "create")
	suite.NoError(err)
	suite.True(allowed)

	// Test denied permission
	denied, err := suite.aaaClient.EvaluatePermission(ctx, "test-user-1", "catalog", "create")
	suite.NoError(err)
	suite.False(denied)

	// Test permission for non-existent user
	nonExistent, err := suite.aaaClient.EvaluatePermission(ctx, "non-existent", "orders", "create")
	suite.NoError(err)
	suite.False(nonExistent)

	// Test resource permission evaluation
	resourceAllowed, err := suite.aaaClient.EvaluateResourcePermission(ctx, "test-user-1", "order", "order-123", "read")
	suite.NoError(err)
	suite.True(resourceAllowed)
}

func (suite *AAAServiceIntegrationTestSuite) testBulkPermissionEvaluation(ctx context.Context) {
	permissions := []auth.PermissionCheck{
		{Resource: "orders", Action: "create"},
		{Resource: "orders", Action: "read"},
		{Resource: "catalog", Action: "create"},
		{Resource: "catalog", Action: "read"},
	}

	results, err := suite.aaaClient.BulkEvaluatePermissions(ctx, "test-user-1", permissions)
	suite.NoError(err)
	suite.Len(results, 4)

	// Verify individual results
	suite.True(results[0].Allowed)  // orders:create - allowed
	suite.True(results[1].Allowed)  // orders:read - allowed
	suite.False(results[2].Allowed) // catalog:create - denied
	suite.True(results[3].Allowed)  // catalog:read - allowed

	for _, result := range results {
		suite.NotEmpty(result.Resource)
		suite.NotEmpty(result.Action)
		suite.NotEmpty(result.Reason)
	}
}

func (suite *AAAServiceIntegrationTestSuite) testOrganizationValidation(ctx context.Context) {
	// Test valid organization membership
	isValid, err := suite.aaaClient.ValidateUserOrganization(ctx, "test-user-1", testutils.TestOrgID)
	suite.NoError(err)
	suite.True(isValid)

	// Test invalid organization membership
	isInvalid, err := suite.aaaClient.ValidateUserOrganization(ctx, "test-user-1", "invalid-org")
	suite.NoError(err)
	suite.False(isInvalid)

	// Test non-existent user
	nonExistent, err := suite.aaaClient.ValidateUserOrganization(ctx, "non-existent", testutils.TestOrgID)
	suite.NoError(err)
	suite.False(nonExistent)
}

// Test service health and reliability
func (suite *AAAServiceIntegrationTestSuite) TestServiceHealthAndReliability() {
	ctx := context.Background()

	suite.T().Log("Testing service health check")
	suite.testServiceHealthCheck(ctx)

	suite.T().Log("Testing service unavailability handling")
	suite.testServiceUnavailability(ctx)

	suite.T().Log("Testing service recovery")
	suite.testServiceRecovery(ctx)

	suite.T().Log("Testing timeout handling")
	suite.testTimeoutHandling(ctx)
}

func (suite *AAAServiceIntegrationTestSuite) testServiceHealthCheck(ctx context.Context) {
	// Test healthy service
	err := suite.aaaClient.HealthCheck(ctx)
	suite.NoError(err)

	// Make service unhealthy
	suite.mockAAAService.SetHealthy(false)

	// Test unhealthy service
	err = suite.aaaClient.HealthCheck(ctx)
	suite.Error(err)
	suite.Contains(err.Error(), "unhealthy")

	// Restore health
	suite.mockAAAService.SetHealthy(true)
}

func (suite *AAAServiceIntegrationTestSuite) testServiceUnavailability(ctx context.Context) {
	// Make service unavailable
	suite.mockAAAService.SetHealthy(false)

	// Test operations fail when service is unavailable
	user, err := suite.aaaClient.GetUser(ctx, "test-user-1")
	suite.Error(err)
	suite.Nil(user)

	authResponse, err := suite.aaaClient.AuthenticateUser(ctx, "testuser1", "correct_password")
	suite.Error(err)
	suite.Nil(authResponse)

	allowed, err := suite.aaaClient.EvaluatePermission(ctx, "test-user-1", "orders", "create")
	suite.Error(err)
	suite.False(allowed)
}

func (suite *AAAServiceIntegrationTestSuite) testServiceRecovery(ctx context.Context) {
	// Service is currently unhealthy from previous test
	suite.False(suite.mockAAAService.isHealthy)

	// Restore service health
	suite.mockAAAService.SetHealthy(true)

	// Test operations work again
	user, err := suite.aaaClient.GetUser(ctx, "test-user-1")
	suite.NoError(err)
	suite.NotNil(user)

	allowed, err := suite.aaaClient.EvaluatePermission(ctx, "test-user-1", "orders", "create")
	suite.NoError(err)
	suite.True(allowed)
}

func (suite *AAAServiceIntegrationTestSuite) testTimeoutHandling(ctx context.Context) {
	// Set long response delay
	suite.mockAAAService.SetResponseDelay(100 * time.Millisecond)

	// Create context with short timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()

	// Test operation times out
	start := time.Now()
	user, err := suite.aaaClient.GetUser(timeoutCtx, "test-user-1")
	duration := time.Since(start)

	// Should timeout quickly
	suite.Less(duration, 100*time.Millisecond)
	suite.Error(err)
	suite.Nil(user)

	// Reset delay
	suite.mockAAAService.SetResponseDelay(0)

	// Test operation works with normal context
	user, err = suite.aaaClient.GetUser(ctx, "test-user-1")
	suite.NoError(err)
	suite.NotNil(user)
}

// Test concurrent operations
func (suite *AAAServiceIntegrationTestSuite) TestConcurrentOperations() {
	ctx := context.Background()

	suite.T().Log("Testing concurrent authentication")
	suite.testConcurrentAuthentication(ctx)

	suite.T().Log("Testing concurrent permission evaluation")
	suite.testConcurrentPermissionEvaluation(ctx)
}

func (suite *AAAServiceIntegrationTestSuite) testConcurrentAuthentication(ctx context.Context) {
	concurrency := 10
	done := make(chan error, concurrency)

	// Perform concurrent authentication requests
	for i := 0; i < concurrency; i++ {
		go func(index int) {
			authResponse, err := suite.aaaClient.AuthenticateUser(ctx, "testuser1", "correct_password")
			if err == nil && authResponse != nil {
				done <- nil
			} else {
				done <- err
			}
		}(i)
	}

	// Wait for all operations to complete
	successCount := 0
	for i := 0; i < concurrency; i++ {
		err := <-done
		if err == nil {
			successCount++
		}
	}

	// All operations should succeed
	suite.Equal(concurrency, successCount, "All concurrent authentication requests should succeed")
}

func (suite *AAAServiceIntegrationTestSuite) testConcurrentPermissionEvaluation(ctx context.Context) {
	concurrency := 10
	done := make(chan bool, concurrency)

	// Perform concurrent permission evaluations
	for i := 0; i < concurrency; i++ {
		go func(index int) {
			allowed, err := suite.aaaClient.EvaluatePermission(ctx, "test-user-1", "orders", "create")
			if err == nil {
				done <- allowed
			} else {
				done <- false
			}
		}(i)
	}

	// Wait for all operations to complete
	successCount := 0
	for i := 0; i < concurrency; i++ {
		allowed := <-done
		if allowed {
			successCount++
		}
	}

	// All operations should succeed and return true
	suite.Equal(concurrency, successCount, "All concurrent permission evaluations should succeed")
}

// Test error scenarios and edge cases
func (suite *AAAServiceIntegrationTestSuite) TestErrorScenariosAndEdgeCases() {
	ctx := context.Background()

	suite.T().Log("Testing invalid input handling")
	suite.testInvalidInputHandling(ctx)

	suite.T().Log("Testing rate limiting simulation")
	suite.testRateLimitingSimulation(ctx)
}

func (suite *AAAServiceIntegrationTestSuite) testInvalidInputHandling(ctx context.Context) {
	// Test nil user creation
	nilUser, err := suite.aaaClient.CreateUser(ctx, nil)
	suite.Error(err)
	suite.Nil(nilUser)

	// Test empty user ID
	emptyUser, err := suite.aaaClient.GetUser(ctx, "")
	suite.Error(err)
	suite.Nil(emptyUser)

	// Test empty credentials
	emptyAuth, err := suite.aaaClient.AuthenticateUser(ctx, "", "")
	suite.Error(err)
	suite.Nil(emptyAuth)

	// Test empty permission parameters
	allowed, err := suite.aaaClient.EvaluatePermission(ctx, "", "", "")
	suite.Error(err)
	suite.False(allowed)
}

func (suite *AAAServiceIntegrationTestSuite) testRateLimitingSimulation(ctx context.Context) {
	// Simulate rate limiting by adding delay
	suite.mockAAAService.SetResponseDelay(10 * time.Millisecond)

	requestCount := 5
	start := time.Now()

	// Make multiple requests
	for i := 0; i < requestCount; i++ {
		_, err := suite.aaaClient.GetUser(ctx, "test-user-1")
		suite.NoError(err)
	}

	duration := time.Since(start)
	expectedMinDuration := time.Duration(requestCount) * 10 * time.Millisecond

	// Should take at least the expected time due to delays
	suite.GreaterOrEqual(duration, expectedMinDuration, "Requests should be rate limited")

	// Reset delay
	suite.mockAAAService.SetResponseDelay(0)
}

// Run the AAA service integration test suite
func TestAAAServiceIntegrationSuite(t *testing.T) {
	suite.Run(t, new(AAAServiceIntegrationTestSuite))
}
