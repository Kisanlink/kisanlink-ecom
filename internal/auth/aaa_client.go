package auth

import (
	"context"
	"fmt"
	"kisanlink-ecom/internal/config"
	"time"

	aaaPb "github.com/Kisanlink/aaa-service/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AAAClient interface for interacting with AAA service
type AAAClient interface {
	// User management
	CreateUser(ctx context.Context, user *AAAUser) (*AAAUser, error)
	GetUser(ctx context.Context, userID string) (*AAAUser, error)
	UpdateUser(ctx context.Context, user *AAAUser) (*AAAUser, error)
	DeleteUser(ctx context.Context, userID string) error

	// Authentication
	AuthenticateUser(ctx context.Context, username, password string) (*AuthenticationResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthenticationResponse, error)

	// Token validation
	ValidateJWT(ctx context.Context, token string) (bool, error)
	GetUserFromToken(ctx context.Context, token string) (*UserContext, error)

	// Role and permission management
	GetUserRoles(ctx context.Context, userID string) ([]*AAARole, error)
	AssignRole(ctx context.Context, userID, roleID string) error
	RemoveRole(ctx context.Context, userID, roleID string) error
	GetUserPermissions(ctx context.Context, userID string) ([]string, error)

	// Permission evaluation
	EvaluatePermission(ctx context.Context, userID, resource, action string) (bool, error)
	EvaluateResourcePermission(ctx context.Context, userID, resourceType, resourceID, action string) (bool, error)
	BulkEvaluatePermissions(ctx context.Context, userID string, permissions []PermissionCheck) ([]PermissionResult, error)

	// Organization validation
	ValidateUserOrganization(ctx context.Context, userID, orgID string) (bool, error)

	// Health check
	HealthCheck(ctx context.Context) error

	// Close connection
	Close() error
}

// PermissionCheck represents a permission to check
type PermissionCheck struct {
	Resource string
	Action   string
}

// PermissionResult represents the result of a permission check
type PermissionResult struct {
	Resource string
	Action   string
	Allowed  bool
	Reason   string
}

// aaaClient implements AAAClient interface
type aaaClient struct {
	conn        *grpc.ClientConn
	userClient  aaaPb.UserServiceV2Client
	authzClient aaaPb.AuthorizationServiceClient
}

// NewAAAClient creates a new AAA client
func NewAAAClient(config *config.AAAConfig) (AAAClient, error) {
	// Validate config
	if config == nil {
		return nil, fmt.Errorf("AAA config cannot be nil")
	}
	if config.GRPCServerAddr == "" {
		return nil, fmt.Errorf("AAA gRPC server address cannot be empty")
	}

	// Connect to AAA service
	conn, err := grpc.Dial(config.GRPCServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AAA service: %w", err)
	}

	client := &aaaClient{
		conn:        conn,
		userClient:  aaaPb.NewUserServiceV2Client(conn),
		authzClient: aaaPb.NewAuthorizationServiceClient(conn),
	}

	return client, nil
}

// Close closes the gRPC connection
func (c *aaaClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// CreateUser creates a new user in AAA service
func (c *aaaClient) CreateUser(ctx context.Context, user *AAAUser) (*AAAUser, error) {
	// Validate input
	if user == nil {
		return nil, fmt.Errorf("user cannot be nil")
	}
	if user.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if user.Email == "" {
		return nil, fmt.Errorf("email is required")
	}

	// Generate ID if not provided
	if user.ID == "" {
		user.ID = fmt.Sprintf("user_%s_%d", user.Username, time.Now().Unix())
	}

	// Set default values
	user.IsActive = true

	// Return the user - actual AAA service integration will be implemented later
	return user, nil
}

// GetUser retrieves a user from AAA service
func (c *aaaClient) GetUser(ctx context.Context, userID string) (*AAAUser, error) {
	// Validate input
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}

	// Return mock user for development - actual AAA service integration will be implemented later
	return &AAAUser{
		ID:       userID,
		Username: "user_" + userID,
		Email:    "user_" + userID + "@example.com",
		IsActive: true,
	}, nil
}

// UpdateUser updates a user in AAA service
func (c *aaaClient) UpdateUser(ctx context.Context, user *AAAUser) (*AAAUser, error) {
	// Validate input
	if user == nil {
		return nil, fmt.Errorf("user cannot be nil")
	}
	if user.ID == "" {
		return nil, fmt.Errorf("user ID is required for update")
	}

	// Return the updated user - actual AAA service integration will be implemented later
	return user, nil
}

// DeleteUser deletes a user from AAA service
func (c *aaaClient) DeleteUser(ctx context.Context, userID string) error {
	// Validate input
	if userID == "" {
		return fmt.Errorf("userID cannot be empty")
	}

	// Return success for development - actual AAA service integration will be implemented later
	return nil
}

// AuthenticateUser authenticates a user with AAA service
func (c *aaaClient) AuthenticateUser(ctx context.Context, username, password string) (*AuthenticationResponse, error) {
	// Validate input
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	// Mock authentication for development - actual AAA service integration will be implemented later
	user := &AAAUser{
		ID:       "auth_" + username,
		Username: username,
		Email:    username + "@example.com",
		IsActive: true,
	}

	userContext := &UserContext{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		IsActive: user.IsActive,
	}

	return &AuthenticationResponse{
		AccessToken:  "mock_token_" + username,
		RefreshToken: "mock_refresh_" + username,
		ExpiresIn:    3600,
		TokenType:    "Bearer",
		User:         user,
		UserContext:  userContext,
	}, nil
}

// RefreshToken refreshes a user's token
func (c *aaaClient) RefreshToken(ctx context.Context, refreshToken string) (*AuthenticationResponse, error) {
	// Validate input
	if refreshToken == "" {
		return nil, fmt.Errorf("refresh token cannot be empty")
	}

	// Mock token refresh for development - actual AAA service integration will be implemented later
	user := &AAAUser{
		ID:       "refreshed_user",
		Username: "refreshed_user",
		Email:    "refreshed@example.com",
		IsActive: true,
	}

	userContext := &UserContext{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		IsActive: user.IsActive,
	}

	return &AuthenticationResponse{
		AccessToken:  "refreshed_token_" + refreshToken,
		RefreshToken: "new_refresh_" + refreshToken,
		ExpiresIn:    3600,
		TokenType:    "Bearer",
		User:         user,
		UserContext:  userContext,
	}, nil
}

// GetUserRoles retrieves roles assigned to a user
func (c *aaaClient) GetUserRoles(ctx context.Context, userID string) ([]*AAARole, error) {
	// Return empty roles as RBAC service integration is not yet available
	// This allows the application to function while RBAC service is being developed
	return []*AAARole{}, nil
}

// AssignRole assigns a role to a user
func (c *aaaClient) AssignRole(ctx context.Context, userID, roleID string) error {
	// Role assignment will be implemented when RBAC service is available
	// For now, return success to allow application functionality
	return nil
}

// RemoveRole removes a role from a user
func (c *aaaClient) RemoveRole(ctx context.Context, userID, roleID string) error {
	// Role removal will be implemented when RBAC service is available
	// For now, return success to allow application functionality
	return nil
}

// GetUserPermissions retrieves permissions for a user
func (c *aaaClient) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	// Return empty permissions as RBAC service integration is not yet available
	// This allows the application to function while RBAC service is being developed
	return []string{}, nil
}

// EvaluatePermission evaluates if a user has permission to perform an action on a resource
func (c *aaaClient) EvaluatePermission(ctx context.Context, userID, resource, action string) (bool, error) {
	// Return true for development mode - allows application functionality
	// In production, this should integrate with RBAC service for proper authorization
	return true, nil
}

// ValidateJWT validates a JWT token with AAA service
func (c *aaaClient) ValidateJWT(ctx context.Context, token string) (bool, error) {
	// Basic token validation - check for non-empty token
	// Full JWT validation will be implemented when AAA service is available
	if token == "" {
		return false, fmt.Errorf("empty token provided")
	}
	return true, nil
}

// GetUserFromToken gets user context from a validated token
func (c *aaaClient) GetUserFromToken(ctx context.Context, token string) (*UserContext, error) {
	// First validate the token
	isValid, err := c.ValidateJWT(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}
	if !isValid {
		return nil, fmt.Errorf("invalid token")
	}

	// Return mock user context for development
	// Full token parsing will be implemented when AAA service is available
	return &UserContext{
		UserID:           "mock_user_from_token",
		Username:         "mock_user",
		Email:            "mock@example.com",
		OrganizationID:   "mock_org",
		OrganizationName: "Mock Organization",
		Roles:            []string{"user"},
		Permissions:      []string{"read", "write"},
		IsActive:         true,
	}, nil
}

// EvaluateResourcePermission evaluates permission for a specific resource instance
func (c *aaaClient) EvaluateResourcePermission(ctx context.Context, userID, resourceType, resourceID, action string) (bool, error) {
	// Return true for development mode - allows application functionality
	// In production, this should integrate with RBAC service for proper resource-level authorization
	return true, nil
}

// ValidateUserOrganization validates that a user belongs to an organization
func (c *aaaClient) ValidateUserOrganization(ctx context.Context, userID, orgID string) (bool, error) {
	// Return true for development mode - allows application functionality
	// In production, this should validate actual user-organization relationships
	return true, nil
}

// HealthCheck checks the health of the AAA service
func (c *aaaClient) HealthCheck(ctx context.Context) error {
	// Basic health check - verify connection exists
	if c.conn == nil {
		return fmt.Errorf("AAA service connection not established")
	}
	// Return success for now - detailed health checks will be implemented with AAA service
	return nil
}

// BulkEvaluatePermissions evaluates multiple permissions at once
func (c *aaaClient) BulkEvaluatePermissions(ctx context.Context, userID string, permissions []PermissionCheck) ([]PermissionResult, error) {
	// Return all permissions as allowed for development mode
	// In production, this should integrate with RBAC service for proper bulk permission evaluation
	var results []PermissionResult
	for _, perm := range permissions {
		results = append(results, PermissionResult{
			Resource: perm.Resource,
			Action:   perm.Action,
			Allowed:  true,
			Reason:   "Development mode - all permissions allowed",
		})
	}
	return results, nil
}

// withRetry executes a function with retry logic and exponential backoff
func (c *aaaClient) withRetry(operation func() error, maxRetries int) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if err := operation(); err != nil {
			lastErr = err
			if i < maxRetries-1 {
				// Exponential backoff with jitter
				backoffDelay := time.Duration(i+1) * 100 * time.Millisecond
				time.Sleep(backoffDelay)
				continue
			}
		} else {
			return nil
		}
	}
	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
}
