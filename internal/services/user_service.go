package services

import (
	"context"
	"fmt"
	"log"

	"kisanlink-ecom/internal/models/user"
)

// UserService handles user-related business logic using gRPC client
type UserService struct {
	grpcClient *GRPCClient
}

// NewUserService creates a new user service instance
func NewUserService(grpcClient *GRPCClient) *UserService {
	return &UserService{
		grpcClient: grpcClient,
	}
}

// CreateUser creates a new user via gRPC
func (s *UserService) CreateUser(ctx context.Context, req user.CreateUserRequest) (*user.User, error) {
	// Convert user role to role IDs (you may need to map your roles to aaa-service roles)
	userRoleIDs := []string{}

	// For now, we'll create a default role. In a real implementation, you'd map roles properly
	switch req.Role {
	case user.RoleAdmin:
		userRoleIDs = append(userRoleIDs, "admin-role-id") // You'd get this from aaa-service
	case user.RoleCustomer:
		userRoleIDs = append(userRoleIDs, "customer-role-id")
	case user.RoleVendor:
		userRoleIDs = append(userRoleIDs, "vendor-role-id")
	}

	// Call gRPC service
	resp, err := s.grpcClient.CreateUser(ctx, req.Username, req.Password, userRoleIDs)
	if err != nil {
		log.Printf("Error creating user via gRPC: %v", err)
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("gRPC service returned status %d: %s", resp.StatusCode, resp.Message)
	}

	// Convert gRPC response to our user model
	userModel := &user.User{
		Username: resp.User.Username,
		Status:   user.StatusActive,
	}

	// Set the ID from gRPC response
	if resp.User.Id != "" {
		userModel.ID = resp.User.Id
	}

	return userModel, nil
}

// GetUserByID retrieves a user by ID via gRPC
func (s *UserService) GetUserByID(ctx context.Context, userID string) (*user.User, error) {
	resp, err := s.grpcClient.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("Error getting user via gRPC: %v", err)
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("gRPC service returned status %d: %s", resp.StatusCode, resp.Message)
	}

	// Convert gRPC response to our user model
	userModel := &user.User{
		Username: resp.User.Username,
		Status:   user.StatusActive,
	}

	if resp.User.Id != "" {
		userModel.ID = resp.User.Id
	}

	return userModel, nil
}

// GetAllUsers retrieves all users via gRPC
func (s *UserService) GetAllUsers(ctx context.Context) ([]*user.User, error) {
	resp, err := s.grpcClient.GetAllUsers(ctx)
	if err != nil {
		log.Printf("Error getting all users via gRPC: %v", err)
		return nil, fmt.Errorf("failed to get users: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("gRPC service returned status %d: %s", resp.StatusCode, resp.Message)
	}

	// Convert gRPC response to our user models
	users := make([]*user.User, 0, len(resp.Users))
	for _, grpcUser := range resp.Users {
		userModel := &user.User{
			Username: grpcUser.Username,
			Status:   user.StatusActive,
		}

		if grpcUser.Id != "" {
			userModel.ID = grpcUser.Id
		}

		users = append(users, userModel)
	}

	return users, nil
}

// UpdateUser updates a user via gRPC
func (s *UserService) UpdateUser(ctx context.Context, userID string, req user.UpdateUserRequest) (*user.User, error) {
	// Convert user role to role IDs
	userRoleIDs := []string{}

	switch req.Role {
	case user.RoleAdmin:
		userRoleIDs = append(userRoleIDs, "admin-role-id")
	case user.RoleCustomer:
		userRoleIDs = append(userRoleIDs, "customer-role-id")
	case user.RoleVendor:
		userRoleIDs = append(userRoleIDs, "vendor-role-id")
	}

	// Call gRPC service - using V2 method for enhanced functionality
	resp, err := s.grpcClient.UpdateUserV2(ctx, userID, req.Username, "", "", "", true, userRoleIDs)
	if err != nil {
		log.Printf("Error updating user via gRPC: %v", err)
		return nil, fmt.Errorf("failed to update user: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("gRPC service returned status %d: %s", resp.StatusCode, resp.Message)
	}

	// Convert gRPC response to our user model
	userModel := &user.User{
		Username: resp.User.Username,
		Status:   user.StatusActive,
	}

	if resp.User.Id != "" {
		userModel.ID = resp.User.Id
	}

	return userModel, nil
}

// DeleteUser deletes a user via gRPC
func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	resp, err := s.grpcClient.DeleteUser(ctx, userID)
	if err != nil {
		log.Printf("Error deleting user via gRPC: %v", err)
		return fmt.Errorf("failed to delete user: %v", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("gRPC service returned status %d: %s", resp.StatusCode, resp.Message)
	}

	return nil
}

// LoginUser authenticates a user via gRPC
func (s *UserService) LoginUser(ctx context.Context, username, password string) (*user.User, string, error) {
	resp, err := s.grpcClient.LoginUser(ctx, username, password)
	if err != nil {
		log.Printf("Error logging in user via gRPC: %v", err)
		return nil, "", fmt.Errorf("failed to login user: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("gRPC service returned status %d: %s", resp.StatusCode, resp.Message)
	}

	// Convert gRPC response to our user model
	userModel := &user.User{
		Username: resp.User.Username,
		Status:   user.StatusActive,
	}

	if resp.User.Id != "" {
		userModel.ID = resp.User.Id
	}

	return userModel, resp.Token, nil
}
