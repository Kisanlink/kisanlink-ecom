package roles

import (
	"context"
	"fmt"
	"time"

	"kisanlink-ecom/entities/models/roles"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// UserRoleRepositoryInterface defines repository operations
type UserRoleRepositoryInterface interface {
	Create(ctx context.Context, userRole *roles.UserRole) error
	GetByID(ctx context.Context, id string) (*roles.UserRole, error)
	GetByUserID(ctx context.Context, userID string) ([]*roles.UserRole, error)
	GetByRoleID(ctx context.Context, roleID string) ([]*roles.UserRole, error)
	Update(ctx context.Context, userRole *roles.UserRole) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter *base.Filter, limit, offset int) ([]*roles.UserRole, int, error)
}

// UserRoleServiceInterface defines service operations
type UserRoleServiceInterface interface {
	AssignRole(ctx context.Context, userID, roleID, assignedBy string, expiresAt *time.Time, notes string) (*roles.UserRole, error)
	GetUserRole(ctx context.Context, id string) (*roles.UserRole, error)
	GetUserRoles(ctx context.Context, userID string) ([]*roles.UserRole, error)
	GetRoleUsers(ctx context.Context, roleID string) ([]*roles.UserRole, error)
	UpdateUserRole(ctx context.Context, userRole *roles.UserRole) (*roles.UserRole, error)
	RevokeRole(ctx context.Context, id string) error
	DeactivateRole(ctx context.Context, id string) error
	ActivateRole(ctx context.Context, id string) error
	ListUserRoles(ctx context.Context, limit, offset int) ([]*roles.UserRole, int, error)
	GetActiveUserRoles(ctx context.Context, userID string) ([]*roles.UserRole, error)
}

// UserRoleService provides business logic for user role management
type UserRoleService struct {
	repo UserRoleRepositoryInterface
}

// NewUserRoleService creates a new user role service
func NewUserRoleService(repo UserRoleRepositoryInterface) *UserRoleService {
	return &UserRoleService{repo: repo}
}

// AssignRole assigns a role to a user with optional expiration
func (s *UserRoleService) AssignRole(ctx context.Context, userID, roleID, assignedBy string, expiresAt *time.Time, notes string) (*roles.UserRole, error) {
	// Validate input
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if roleID == "" {
		return nil, fmt.Errorf("role ID is required")
	}
	if assignedBy == "" {
		return nil, fmt.Errorf("assigned by is required")
	}

	// Check if user already has this role (active)
	existing, err := s.repo.GetByUserID(ctx, userID)
	if err == nil && existing != nil {
		for _, ur := range existing {
			if ur.RoleId == roleID && ur.IsActive && !ur.IsExpired() {
				return nil, fmt.Errorf("user already has this role")
			}
		}
	}

	// Create new role assignment
	userRole := roles.NewUserRole(userID, roleID, assignedBy)
	userRole.Notes = notes

	if expiresAt != nil {
		userRole.SetExpiration(*expiresAt)
	}

	if err := s.repo.Create(ctx, userRole); err != nil {
		return nil, fmt.Errorf("failed to assign role: %w", err)
	}

	return userRole, nil
}

// GetUserRole retrieves a user role by ID
func (s *UserRoleService) GetUserRole(ctx context.Context, id string) (*roles.UserRole, error) {
	if id == "" {
		return nil, fmt.Errorf("user role ID is required")
	}

	userRole, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user role: %w", err)
	}
	if userRole == nil {
		return nil, fmt.Errorf("user role not found")
	}

	return userRole, nil
}

// GetUserRoles retrieves all roles for a user
func (s *UserRoleService) GetUserRoles(ctx context.Context, userID string) ([]*roles.UserRole, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	userRoles, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	return userRoles, nil
}

// GetActiveUserRoles retrieves all active and non-expired roles for a user
func (s *UserRoleService) GetActiveUserRoles(ctx context.Context, userID string) ([]*roles.UserRole, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	allRoles, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Filter to only active and non-expired roles
	activeRoles := make([]*roles.UserRole, 0)
	for _, ur := range allRoles {
		if ur.IsValid() {
			activeRoles = append(activeRoles, ur)
		}
	}

	return activeRoles, nil
}

// GetRoleUsers retrieves all users with a specific role
func (s *UserRoleService) GetRoleUsers(ctx context.Context, roleID string) ([]*roles.UserRole, error) {
	if roleID == "" {
		return nil, fmt.Errorf("role ID is required")
	}

	userRoles, err := s.repo.GetByRoleID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role users: %w", err)
	}

	return userRoles, nil
}

// UpdateUserRole updates an existing user role assignment
func (s *UserRoleService) UpdateUserRole(ctx context.Context, userRole *roles.UserRole) (*roles.UserRole, error) {
	if userRole == nil {
		return nil, fmt.Errorf("user role cannot be nil")
	}

	// Validate exists
	existing, err := s.repo.GetByID(ctx, userRole.Id)
	if err != nil || existing == nil {
		return nil, fmt.Errorf("user role not found")
	}

	// Validate expiration date if set
	if userRole.ExpiresAt != nil && userRole.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("expiration date cannot be in the past")
	}

	if err := s.repo.Update(ctx, userRole); err != nil {
		return nil, fmt.Errorf("failed to update user role: %w", err)
	}

	return userRole, nil
}

// RevokeRole revokes a role assignment (soft delete)
func (s *UserRoleService) RevokeRole(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("user role ID is required")
	}

	// Check exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("user role not found")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}

	return nil
}

// DeactivateRole deactivates a role assignment without deleting it
func (s *UserRoleService) DeactivateRole(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("user role ID is required")
	}

	userRole, err := s.repo.GetByID(ctx, id)
	if err != nil || userRole == nil {
		return fmt.Errorf("user role not found")
	}

	userRole.IsActive = false

	if err := s.repo.Update(ctx, userRole); err != nil {
		return fmt.Errorf("failed to deactivate role: %w", err)
	}

	return nil
}

// ActivateRole activates a previously deactivated role assignment
func (s *UserRoleService) ActivateRole(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("user role ID is required")
	}

	userRole, err := s.repo.GetByID(ctx, id)
	if err != nil || userRole == nil {
		return fmt.Errorf("user role not found")
	}

	// Check if role has expired
	if userRole.IsExpired() {
		return fmt.Errorf("cannot activate expired role")
	}

	userRole.IsActive = true

	if err := s.repo.Update(ctx, userRole); err != nil {
		return fmt.Errorf("failed to activate role: %w", err)
	}

	return nil
}

// ListUserRoles retrieves user roles with pagination
func (s *UserRoleService) ListUserRoles(ctx context.Context, limit, offset int) ([]*roles.UserRole, int, error) {
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	userRoles, total, err := s.repo.List(ctx, nil, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list user roles: %w", err)
	}

	return userRoles, total, nil
}
