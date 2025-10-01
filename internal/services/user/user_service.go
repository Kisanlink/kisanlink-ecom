package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"kisanlink-ecom/entities/models/roles"
	"kisanlink-ecom/entities/models/user"
	"kisanlink-ecom/entities/requests/auth"
	authService "kisanlink-ecom/internal/auth"
)

// UserService handles user-related business logic and AAA service integration
type UserService struct {
	aaaClient authService.Client
	// TODO: Add user repository when available
}

// NewUserService creates a new user service instance
func NewUserService(aaaClient authService.Client) *UserService {
	return &UserService{
		aaaClient: aaaClient,
	}
}

// CreateUser creates a new user account by registering with AAA service
func (s *UserService) CreateUser(ctx context.Context, req auth.RegisterRequest) (*user.User, error) {
	// First, register the user with AAA service
	aaaUser, err := s.registerWithAAA(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to register with AAA service: %w", err)
	}

	// Create local user reference
	localUser := user.NewUser(
		aaaUser.UserID,
		req.Username,
		req.Email,
		req.FirstName,
		req.LastName,
		req.Phone,
	)

	// TODO: Save local user to database
	// For now, return the user object
	return localUser, nil
}

// LoginUser authenticates a user via AAA service and returns a token
func (s *UserService) LoginUser(ctx context.Context, username, password string) (*user.User, string, error) {
	// Authenticate with AAA service
	authResponse, err := s.authenticateWithAAA(ctx, username, password)
	if err != nil {
		return nil, "", fmt.Errorf("authentication failed: %w", err)
	}

	// Get or create local user reference
	localUser, err := s.getOrCreateLocalUser(ctx, authResponse.UserContext)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get local user: %w", err)
	}

	// Update last login time
	now := time.Now()
	localUser.LastLoginAt = &now
	// TODO: Save updated user to database

	return localUser, authResponse.AccessToken, nil
}

// GetUserByID retrieves a user by ID, fetching latest data from AAA service
func (s *UserService) GetUserByID(ctx context.Context, userID string) (*user.User, error) {
	// TODO: Get from local database first
	// For now, fetch from AAA service
	aaaUser, err := s.getUserFromAAA(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from AAA service: %w", err)
	}

	// Convert to local user format
	localUser := user.NewUser(
		aaaUser.UserID,
		aaaUser.Username,
		aaaUser.Email,
		"", // FirstName not available in UserContext
		"", // LastName not available in UserContext
		"", // Phone not available in UserContext
	)

	return localUser, nil
}

// GetUserRoles retrieves the roles assigned to a user
func (s *UserService) GetUserRoles(ctx context.Context, userID string) ([]*roles.EcommerceRole, error) {
	// TODO: Implement role retrieval from local database
	// This would join user_roles, ecommerce_roles, and organization_roles tables
	return []*roles.EcommerceRole{}, nil
}

// AssignRole assigns a role to a user
func (s *UserService) AssignRole(ctx context.Context, userID, roleID, assignedBy string) error {
	// TODO: Implement role assignment
	// This would create a user_role record and validate permissions
	return nil
}

// registerWithAAA registers a user with the AAA service
func (s *UserService) registerWithAAA(ctx context.Context, req auth.RegisterRequest) (*authService.UserContext, error) {
	if s.aaaClient == nil {
		// Mock implementation when AAA client is not available
		return &authService.UserContext{
			UserID:   generateID(),
			Username: req.Username,
			Email:    req.Email,
			IsActive: true,
		}, nil
	}

	// TODO: Implement actual AAA service registration
	// This would call the AAA service gRPC endpoint
	return nil, fmt.Errorf("AAA service integration not implemented yet")
}

// authenticateWithAAA authenticates a user with the AAA service
func (s *UserService) authenticateWithAAA(ctx context.Context, username, password string) (*authService.AuthenticationResponse, error) {
	if s.aaaClient == nil {
		// Mock implementation when AAA client is not available
		userContext := &authService.UserContext{
			UserID:   generateID(),
			Username: username,
			Email:    username + "@example.com",
			IsActive: true,
		}

		authResponse := &authService.AuthenticationResponse{
			AccessToken:  generateToken(),
			RefreshToken: generateToken(),
			ExpiresIn:    3600,
			TokenType:    "Bearer",
			UserContext:  userContext,
		}

		return authResponse, nil
	}

	// TODO: Implement actual AAA service authentication
	// This would call the AAA service gRPC endpoint
	return nil, fmt.Errorf("AAA service authentication not implemented yet")
}

// getUserFromAAA retrieves user data from the AAA service
func (s *UserService) getUserFromAAA(ctx context.Context, userID string) (*authService.UserContext, error) {
	if s.aaaClient == nil {
		// Mock implementation when AAA client is not available
		return &authService.UserContext{
			UserID:   userID,
			Username: "mock-user",
			IsActive: true,
		}, nil
	}

	// TODO: Implement actual AAA service user retrieval
	// This would call the AAA service gRPC endpoint
	return nil, fmt.Errorf("AAA service integration not implemented yet")
}

// getOrCreateLocalUser gets or creates a local user reference
func (s *UserService) getOrCreateLocalUser(ctx context.Context, userContext *authService.UserContext) (*user.User, error) {
	// TODO: Check if user exists in local database
	// If not, create new local user reference
	localUser := user.NewUser(
		userContext.UserID,
		userContext.Username,
		userContext.Email,
		"", // FirstName not available in UserContext
		"", // LastName not available in UserContext
		"", // Phone not available in UserContext
	)

	return localUser, nil
}

// generateID generates a random ID
func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// generateToken generates a mock JWT token
func generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return "mock-jwt-" + hex.EncodeToString(bytes)
}
