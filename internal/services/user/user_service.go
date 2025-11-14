package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/roles"
	"github.com/Kisanlink/kisanlink-ecom/entities/models/user"
	"github.com/Kisanlink/kisanlink-ecom/entities/requests/auth"
	authService "github.com/Kisanlink/kisanlink-ecom/internal/auth"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// RepositoryInterface defines repository operations for user entities
type RepositoryInterface interface {
	Create(ctx context.Context, user *user.User) error
	GetByID(ctx context.Context, id string) (*user.User, error)
	GetByUsername(ctx context.Context, username string) (*user.User, error)
	GetByEmail(ctx context.Context, email string) (*user.User, error)
	Update(ctx context.Context, user *user.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter *base.Filter, limit, offset int) ([]*user.User, int, error)
}

// ServiceInterface defines service operations for user management
type ServiceInterface interface {
	CreateUser(ctx context.Context, req auth.RegisterRequest) (*user.User, error)
	GetUserByID(ctx context.Context, userID string) (*user.User, error)
	GetUserByUsername(ctx context.Context, username string) (*user.User, error)
	GetUserByEmail(ctx context.Context, email string) (*user.User, error)
	UpdateUser(ctx context.Context, user *user.User) (*user.User, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, limit, offset int) ([]*user.User, int, error)
	LoginUser(ctx context.Context, username, password string) (*user.User, string, error)
	GetUserRoles(ctx context.Context, userID string) ([]*roles.EcommerceRole, error)
	AssignRole(ctx context.Context, userID, roleID, assignedBy string) error
}

// UserService handles user-related business logic and AAA service integration
type UserService struct {
	aaaClient authService.Client
	repo      RepositoryInterface
}

// NewUserService creates a new user service instance
func NewUserService(aaaClient authService.Client, repo RepositoryInterface) *UserService {
	return &UserService{
		aaaClient: aaaClient,
		repo:      repo,
	}
}

// CreateUser creates a new user account by registering with AAA service
func (s *UserService) CreateUser(ctx context.Context, req auth.RegisterRequest) (*user.User, error) {
	// Validate input
	if err := s.validateUserRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if username already exists
	if s.repo != nil {
		existing, err := s.repo.GetByUsername(ctx, req.Username)
		if err == nil && existing != nil {
			return nil, fmt.Errorf("username %s already exists", req.Username)
		}
	}

	// Check if email already exists
	if s.repo != nil {
		existing, err := s.repo.GetByEmail(ctx, req.Email)
		if err == nil && existing != nil {
			return nil, fmt.Errorf("email %s already exists", req.Email)
		}
	}

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

	// Save local user to database
	if s.repo != nil {
		if err := s.repo.Create(ctx, localUser); err != nil {
			return nil, fmt.Errorf("failed to create local user: %w", err)
		}
	}

	return localUser, nil
}

// LoginUser authenticates a user via AAA service and returns a token
func (s *UserService) LoginUser(ctx context.Context, username, password string) (*user.User, string, error) {
	// Validate credentials
	if username == "" || password == "" {
		return nil, "", fmt.Errorf("username and password are required")
	}

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

	// Save updated user to database
	if s.repo != nil {
		if err := s.repo.Update(ctx, localUser); err != nil {
			return nil, "", fmt.Errorf("failed to update user login time: %w", err)
		}
	}

	return localUser, authResponse.AccessToken, nil
}

// GetUserByID retrieves a user by ID, fetching latest data from AAA service
func (s *UserService) GetUserByID(ctx context.Context, userID string) (*user.User, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Get from local database first
	if s.repo != nil {
		localUser, err := s.repo.GetByID(ctx, userID)
		if err == nil && localUser != nil {
			return localUser, nil
		}
	}

	// Fallback to AAA service
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

// GetUserByUsername retrieves a user by username
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*user.User, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}

	if s.repo == nil {
		return nil, fmt.Errorf("repository not available")
	}

	localUser, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if localUser == nil {
		return nil, fmt.Errorf("user not found")
	}

	return localUser, nil
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	if s.repo == nil {
		return nil, fmt.Errorf("repository not available")
	}

	localUser, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if localUser == nil {
		return nil, fmt.Errorf("user not found")
	}

	return localUser, nil
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(ctx context.Context, usr *user.User) (*user.User, error) {
	if usr == nil {
		return nil, fmt.Errorf("user cannot be nil")
	}

	if s.repo == nil {
		return nil, fmt.Errorf("repository not available")
	}

	// Validate exists
	existing, err := s.repo.GetByID(ctx, usr.Id)
	if err != nil || existing == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Validate user data
	if err := s.validateUser(usr); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if err := s.repo.Update(ctx, usr); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return usr, nil
}

// DeleteUser soft deletes a user
func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("user ID is required")
	}

	if s.repo == nil {
		return fmt.Errorf("repository not available")
	}

	// Check exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ListUsers retrieves users with pagination
func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]*user.User, int, error) {
	if s.repo == nil {
		return nil, 0, fmt.Errorf("repository not available")
	}

	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	users, total, err := s.repo.List(ctx, nil, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
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
	// Check if user exists in local database
	if s.repo != nil {
		existing, err := s.repo.GetByID(ctx, userContext.UserID)
		if err == nil && existing != nil {
			return existing, nil
		}
	}

	// Create new local user reference
	localUser := user.NewUser(
		userContext.UserID,
		userContext.Username,
		userContext.Email,
		"", // FirstName not available in UserContext
		"", // LastName not available in UserContext
		"", // Phone not available in UserContext
	)

	// Save to database
	if s.repo != nil {
		if err := s.repo.Create(ctx, localUser); err != nil {
			return nil, fmt.Errorf("failed to create local user: %w", err)
		}
	}

	return localUser, nil
}

// validateUserRequest validates user registration request
func (s *UserService) validateUserRequest(req auth.RegisterRequest) error {
	if req.Username == "" {
		return fmt.Errorf("username is required")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.FirstName == "" {
		return fmt.Errorf("first name is required")
	}
	if req.LastName == "" {
		return fmt.Errorf("last name is required")
	}
	return nil
}

// validateUser validates user data
func (s *UserService) validateUser(usr *user.User) error {
	if usr.Username == "" {
		return fmt.Errorf("username is required")
	}
	if usr.Email == "" {
		return fmt.Errorf("email is required")
	}
	if usr.FirstName == "" {
		return fmt.Errorf("first name is required")
	}
	if usr.LastName == "" {
		return fmt.Errorf("last name is required")
	}
	return nil
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
