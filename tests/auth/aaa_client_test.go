package auth_test

import (
    "context"
    "testing"

    "kisanlink-ecom/internal/auth"

    "github.com/stretchr/testify/assert"
)

func TestMockAAAClient_CreateUser(t *testing.T) {
    // Setup
    client := auth.NewMockAAAClient()
    ctx := context.Background()

    user := &auth.AAAUser{
        Username:         "testuser",
        Email:            "test@example.com",
        FirstName:        "Test",
        LastName:         "User",
        Phone:            "+1234567890",
        IsActive:         true,
        OrganizationID:   "org123",
        OrganizationName: "Test Organization",
    }

    // Test
    createdUser, err := client.CreateUser(ctx, user)

    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, createdUser)
    assert.NotEmpty(t, createdUser.ID)
    assert.Equal(t, user.Username, createdUser.Username)
    assert.Equal(t, user.Email, createdUser.Email)
    assert.Equal(t, user.FirstName, createdUser.FirstName)
    assert.Equal(t, user.LastName, createdUser.LastName)
    assert.Equal(t, user.Phone, createdUser.Phone)
    assert.Equal(t, user.IsActive, createdUser.IsActive)
    assert.Equal(t, user.OrganizationID, createdUser.OrganizationID)
    assert.Equal(t, user.OrganizationName, createdUser.OrganizationName)
}

func TestMockAAAClient_GetUser_Success(t *testing.T) {
    // Setup
    client := auth.NewMockAAAClient()
    ctx := context.Background()

    // Create a user first
    originalUser := &auth.AAAUser{
        Username:         "testuser",
        Email:            "test@example.com",
        FirstName:        "Test",
        LastName:         "User",
        IsActive:         true,
        OrganizationID:   "org123",
        OrganizationName: "Test Organization",
    }

    createdUser, err := client.CreateUser(ctx, originalUser)
    assert.NoError(t, err)

    // Test
    retrievedUser, err := client.GetUser(ctx, createdUser.ID)

    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, retrievedUser)
    assert.Equal(t, createdUser.ID, retrievedUser.ID)
    assert.Equal(t, createdUser.Username, retrievedUser.Username)
    assert.Equal(t, createdUser.Email, retrievedUser.Email)
}

func TestMockAAAClient_GetUser_NotFound(t *testing.T) {
    // Setup
    client := auth.NewMockAAAClient()
    ctx := context.Background()

    // Test
    user, err := client.GetUser(ctx, "nonexistent-user")

    // Assertions
    assert.Error(t, err)
    assert.Nil(t, user)
    assert.Contains(t, err.Error(), "user not found")
}

func TestMockAAAClient_UpdateUser(t *testing.T) {
    // Setup
    client := auth.NewMockAAAClient()
    ctx := context.Background()

    // Create a user first
    originalUser := &auth.AAAUser{
        Username:  "testuser",
        Email:     "test@example.com",
        FirstName: "Test",
        LastName:  "User",
        IsActive:  true,
    }

    createdUser, err := client.CreateUser(ctx, originalUser)
    assert.NoError(t, err)

    // Update the user
    createdUser.Email = "updated@example.com"
    createdUser.FirstName = "Updated"

    // Test
    updatedUser, err := client.UpdateUser(ctx, createdUser)

    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, updatedUser)
    assert.Equal(t, createdUser.ID, updatedUser.ID)
    assert.Equal(t, "updated@example.com", updatedUser.Email)
    assert.Equal(t, "Updated", updatedUser.FirstName)
}

func TestMockAAAClient_DeleteUser(t *testing.T) {
    // Setup
    client := auth.NewMockAAAClient()
    ctx := context.Background()

    // Create a user first
    user := &auth.AAAUser{
        Username: "testuser",
        Email:    "test@example.com",
        IsActive: true,
    }

    createdUser, err := client.CreateUser(ctx, user)
    assert.NoError(t, err)

    // Test
    err = client.DeleteUser(ctx, createdUser.ID)
    assert.NoError(t, err)

    // Verify user is deleted
    deletedUser, err := client.GetUser(ctx, createdUser.ID)
    assert.Error(t, err)
    assert.Nil(t, deletedUser)
}

func TestMockAAAClient_AuthenticateUser_Success(t *testing.T) {
    // Setup
    client := auth.NewMockAAAClient()
    ctx := context.Background()

    // Create a user first
    user := &auth.AAAUser{
        Username:         "testuser",
        Email:            "test@example.com",
        IsActive:         true,
        OrganizationID:   "org123",
        OrganizationName: "Test Organization",
    }

    createdUser, err := client.CreateUser(ctx, user)
    assert.NoError(t, err)

    // Test
    authResponse, err := client.AuthenticateUser(ctx, createdUser.Username, "password")

    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, authResponse)
    assert.Equal(t, "mock-access-token", authResponse.AccessToken)
    assert.Equal(t, "mock-refresh-token", authResponse.RefreshToken)
    assert.Equal(t, int64(3600), authResponse.ExpiresIn)
    assert.Equal(t, "Bearer", authResponse.TokenType)
    assert.NotNil(t, authResponse.User)
    assert.NotNil(t, authResponse.UserContext)
    assert.Equal(t, createdUser.ID, authResponse.User.ID)
    assert.Equal(t, createdUser.Username, authResponse.User.Username)
}

func TestMockAAAClient_AuthenticateUser_UserNotFound(t *testing.T) {
    // Setup
    client := auth.NewMockAAAClient()
    ctx := context.Background()

    // Test with non-existent user
    authResponse, err := client.AuthenticateUser(ctx, "nonexistent", "password")

    // Assertions
    assert.Error(t, err)
    assert.Nil(t, authResponse)
    assert.Contains(t, err.Error(), "authentication failed")
}

func TestMockAAAClient_RefreshToken(t *testing.T) {
    // Setup
    client := auth.NewMockAAAClient()
    ctx := context.Background()

    // Test
    authResponse, err := client.RefreshToken(ctx, "mock-refresh-token")

    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, authResponse)
    assert.Equal(t, "new-mock-access-token", authResponse.AccessToken)
    assert.Equal(t, "new-mock-refresh-token", authResponse.RefreshToken)
    assert.Equal(t, int64(3600), authResponse.ExpiresIn)
    assert.Equal(t, "Bearer", authResponse.TokenType)
}

func TestMockAAAClient_ValidateJWT(t *testing.T) {
    // Setup
    client := auth.NewMockAAAClient()
    ctx := context.Background()

    tests := []struct {
        name     string
        token    string
        expected bool
    }{
        {
            name:     "valid mock token",
            token:    "mock-access-token",
            expected: true,
        },
        {
            name:     "valid bearer token",
            token:    "Bearer mock-access-token",
            expected: true,
        },
        {
            name:     "invalid token",
            token:    "invalid-token",
            expected: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test
            valid, err := client.ValidateJWT(ctx, tt.token)

            // Assertions
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, valid)
        })
    }
}

func TestMockAAAClient_GetUserFromToken(t *testing.T) {
    // Setup
    client := auth.NewMockAAAClient()
    ctx := context.Background()

    // Create a user first
    user := &auth.AAAUser{
        Username:         "testuser",
        Email:            "test@example.com",
        IsActive:         true,
        OrganizationID:   "org123",
        OrganizationName: "Test Organization",
    }

    createdUser, err := client.CreateUser(ctx, user)
    assert.NoError(t, err)

    // Test
    userContext, err := client.GetUserFromToken(ctx, "mock-token")

    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, userContext)
    assert.Equal(t, createdUser.ID, userContext.UserID)
    assert.Equal(t, createdUser.Username, userContext.Username)
    assert.Equal(t, createdUser.Email, userContext.Email)
    assert.Equal(t, createdUser.OrganizationID, userContext.OrganizationID)
    assert.Equal(t, createdUser.OrganizationName, userContext.OrganizationName)
    assert.True(t, userContext.IsActive)
}

func TestMockAAAClient_EvaluatePermission(t *testing.T) {
    // Setup
    client := auth.NewMockAAAClient()
    ctx := context.Background()

    // Test
    allowed, err := client.EvaluatePermission(ctx, "user123", "orders", "create")

    // Assertions
    assert.NoError(t, err)
    assert.True(t, allowed) // Mock implementation always allows
}

func TestMockAAAClient_EvaluateResourcePermission(t *testing.T) {
    // Setup
    client := auth.NewMockAAAClient()
    ctx := context.Background()

    // Test
    allowed, err := client.EvaluateResourcePermission(ctx, "user123", "order", "order123", "view")

    // Assertions
    assert.NoError(t, err)
    assert.True(t, allowed) // Mock implementation always allows
}

func TestMockAAAClient_BulkEvaluatePermissions(t *testing.T) {
    // Setup
    client := auth.NewMockAAAClient()
    ctx := context.Background()

    permissions := []auth.PermissionCheck{
        {Resource: "orders", Action: "create"},
        {Resource: "products", Action: "read"},
        {Resource: "users", Action: "update"},
    }

    // Test
    results, err := client.BulkEvaluatePermissions(ctx, "user123", permissions)

    // Assertions
    assert.NoError(t, err)
    assert.Len(t, results, 3)

    for i, result := range results {
        assert.Equal(t, permissions[i].Resource, result.Resource)
        assert.Equal(t, permissions[i].Action, result.Action)
        assert.True(t, result.Allowed) // Mock implementation always allows
        assert.Equal(t, "mock implementation", result.Reason)
    }
}

func TestMockAAAClient_RoleManagement(t *testing.T) {
    // Setup
    client := auth.NewMockAAAClient()
    ctx := context.Background()

    // Create a user first
    user := &auth.AAAUser{
        Username: "testuser",
        Email:    "test@example.com",
        IsActive: true,
    }

    createdUser, err := client.CreateUser(ctx, user)
    assert.NoError(t, err)

    // Test AssignRole
    err = client.AssignRole(ctx, createdUser.ID, "admin")
    assert.NoError(t, err)

    err = client.AssignRole(ctx, createdUser.ID, "user")
    assert.NoError(t, err)

    // Test GetUserRoles
    roles, err := client.GetUserRoles(ctx, createdUser.ID)
    assert.NoError(t, err)
    assert.Len(t, roles, 2)

    // Test RemoveRole
    err = client.RemoveRole(ctx, createdUser.ID, "admin")
    assert.NoError(t, err)

    // Verify role was removed
    roles, err = client.GetUserRoles(ctx, createdUser.ID)
    assert.NoError(t, err)
    assert.Len(t, roles, 1)
    assert.Equal(t, "user", roles[0].ID)
}

func TestMockAAAClient_GetUserPermissions(t *testing.T) {
    // Setup
    client := auth.NewMockAAAClient()
    ctx := context.Background()

    // Test (empty permissions for new user)
    permissions, err := client.GetUserPermissions(ctx, "user123")

    // Assertions
    assert.NoError(t, err)
    assert.Empty(t, permissions) // Mock returns empty by default
}

func TestMockAAAClient_ValidateUserOrganization(t *testing.T) {
    // Setup
    client := auth.NewMockAAAClient()
    ctx := context.Background()

    // Create a user first
    user := &auth.AAAUser{
        Username:       "testuser",
        Email:          "test@example.com",
        IsActive:       true,
        OrganizationID: "org123",
    }

    createdUser, err := client.CreateUser(ctx, user)
    assert.NoError(t, err)

    // Test with correct organization
    valid, err := client.ValidateUserOrganization(ctx, createdUser.ID, "org123")
    assert.NoError(t, err)
    assert.True(t, valid)

    // Test with wrong organization
    valid, err = client.ValidateUserOrganization(ctx, createdUser.ID, "wrong-org")
    assert.NoError(t, err)
    assert.False(t, valid)

    // Test with non-existent user
    valid, err = client.ValidateUserOrganization(ctx, "nonexistent", "org123")
    assert.Error(t, err)
    assert.False(t, valid)
    assert.Contains(t, err.Error(), "user not found")
}

func TestMockAAAClient_HealthCheck(t *testing.T) {
    // Setup
    client := auth.NewMockAAAClient()
    ctx := context.Background()

    // Test
    err := client.HealthCheck(ctx)

    // Assertions
    assert.NoError(t, err) // Mock always returns success
}

func TestMockAAAClient_Close(t *testing.T) {
    // Setup
    client := auth.NewMockAAAClient()

    // Test
    err := client.Close()

    // Assertions
    assert.NoError(t, err) // Mock always returns success
}

func TestCircuitBreaker_Success(t *testing.T) {
    // This test would require exposing CircuitBreaker or testing it indirectly
    // through the EnhancedAAAClient, which would require a more complex setup
    // For now, we'll skip this test as it would require significant refactoring
    t.Skip("CircuitBreaker testing requires refactoring for testability")
}

func TestCircuitBreaker_MaxFailures(t *testing.T) {
    t.Skip("CircuitBreaker testing requires refactoring for testability")
}

func TestEnhancedAAAClient_Integration(t *testing.T) {
    // This would test the full enhanced client but requires a running AAA service
    // or more sophisticated mocking of the gRPC calls
    t.Skip("Integration test requires running AAA service")
}

// Benchmark tests
func BenchmarkMockAAAClient_AuthenticateUser(b *testing.B) {
    client := auth.NewMockAAAClient()
    ctx := context.Background()

    // Create a user first
    user := &auth.AAAUser{
        Username: "testuser",
        Email:    "test@example.com",
        IsActive: true,
    }

    createdUser, err := client.CreateUser(ctx, user)
    if err != nil {
        b.Fatal(err)
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := client.AuthenticateUser(ctx, createdUser.Username, "password")
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkMockAAAClient_ValidateJWT(b *testing.B) {
    client := auth.NewMockAAAClient()
    ctx := context.Background()
    token := "mock-access-token"

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := client.ValidateJWT(ctx, token)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkMockAAAClient_EvaluatePermission(b *testing.B) {
    client := auth.NewMockAAAClient()
    ctx := context.Background()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := client.EvaluatePermission(ctx, "user123", "orders", "create")
        if err != nil {
            b.Fatal(err)
        }
    }
}
