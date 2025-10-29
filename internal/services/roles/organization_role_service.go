package roles

import (
	"context"
	"encoding/json"
	"fmt"

	"kisanlink-ecom/entities/models/roles"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// OrganizationRoleRepositoryInterface defines repository operations
type OrganizationRoleRepositoryInterface interface {
	Create(ctx context.Context, orgRole *roles.OrganizationRole) error
	GetByID(ctx context.Context, id string) (*roles.OrganizationRole, error)
	GetByOrganizationID(ctx context.Context, organizationID string) ([]*roles.OrganizationRole, error)
	GetByAAARoleID(ctx context.Context, aaaRoleID string) ([]*roles.OrganizationRole, error)
	GetDefaultRole(ctx context.Context, organizationID string) (*roles.OrganizationRole, error)
	Update(ctx context.Context, orgRole *roles.OrganizationRole) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter *base.Filter, limit, offset int) ([]*roles.OrganizationRole, int, error)
}

// OrganizationRoleServiceInterface defines service operations
type OrganizationRoleServiceInterface interface {
	CreateOrganizationRole(ctx context.Context, organizationID, aaaRoleID, configuredBy string, customPermissions map[string]bool, maxUsers *int) (*roles.OrganizationRole, error)
	GetOrganizationRole(ctx context.Context, id string) (*roles.OrganizationRole, error)
	GetOrganizationRoles(ctx context.Context, organizationID string) ([]*roles.OrganizationRole, error)
	GetDefaultRole(ctx context.Context, organizationID string) (*roles.OrganizationRole, error)
	SetAsDefault(ctx context.Context, organizationID, roleID string) error
	UpdateOrganizationRole(ctx context.Context, orgRole *roles.OrganizationRole) (*roles.OrganizationRole, error)
	UpdateCustomPermissions(ctx context.Context, id string, permissions map[string]bool) (*roles.OrganizationRole, error)
	UpdateMaxUsers(ctx context.Context, id string, maxUsers int) (*roles.OrganizationRole, error)
	DeleteOrganizationRole(ctx context.Context, id string) error
	ListOrganizationRoles(ctx context.Context, limit, offset int) ([]*roles.OrganizationRole, int, error)
}

// OrganizationRoleService provides business logic for organization role management
type OrganizationRoleService struct {
	repo OrganizationRoleRepositoryInterface
}

// NewOrganizationRoleService creates a new organization role service
func NewOrganizationRoleService(repo OrganizationRoleRepositoryInterface) *OrganizationRoleService {
	return &OrganizationRoleService{repo: repo}
}

// CreateOrganizationRole creates a new organization-specific role configuration
func (s *OrganizationRoleService) CreateOrganizationRole(ctx context.Context, organizationID, aaaRoleID, configuredBy string, customPermissions map[string]bool, maxUsers *int) (*roles.OrganizationRole, error) {
	// Validate input
	if organizationID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}
	if aaaRoleID == "" {
		return nil, fmt.Errorf("AAA role ID is required")
	}
	if configuredBy == "" {
		return nil, fmt.Errorf("configured by is required")
	}

	// Check if organization already has this role configured
	existing, err := s.repo.GetByOrganizationID(ctx, organizationID)
	if err == nil && existing != nil {
		for _, or := range existing {
			if or.AAARoleID == aaaRoleID && or.IsActive {
				return nil, fmt.Errorf("organization already has this role configured")
			}
		}
	}

	// Create new organization role
	orgRole := roles.NewOrganizationRole(organizationID, aaaRoleID, configuredBy)

	// Set custom permissions if provided
	if len(customPermissions) > 0 {
		permJSON, err := json.Marshal(customPermissions)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal custom permissions: %w", err)
		}
		orgRole.CustomPermissions = string(permJSON)
	}

	// Set max users if provided
	if maxUsers != nil {
		if *maxUsers <= 0 {
			return nil, fmt.Errorf("max users must be greater than 0")
		}
		orgRole.SetMaxUsers(*maxUsers)
	}

	if err := s.repo.Create(ctx, orgRole); err != nil {
		return nil, fmt.Errorf("failed to create organization role: %w", err)
	}

	return orgRole, nil
}

// GetOrganizationRole retrieves an organization role by ID
func (s *OrganizationRoleService) GetOrganizationRole(ctx context.Context, id string) (*roles.OrganizationRole, error) {
	if id == "" {
		return nil, fmt.Errorf("organization role ID is required")
	}

	orgRole, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization role: %w", err)
	}
	if orgRole == nil {
		return nil, fmt.Errorf("organization role not found")
	}

	return orgRole, nil
}

// GetOrganizationRoles retrieves all roles for an organization
func (s *OrganizationRoleService) GetOrganizationRoles(ctx context.Context, organizationID string) ([]*roles.OrganizationRole, error) {
	if organizationID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	orgRoles, err := s.repo.GetByOrganizationID(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization roles: %w", err)
	}

	return orgRoles, nil
}

// GetDefaultRole retrieves the default role for new users in an organization
func (s *OrganizationRoleService) GetDefaultRole(ctx context.Context, organizationID string) (*roles.OrganizationRole, error) {
	if organizationID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	orgRole, err := s.repo.GetDefaultRole(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get default role: %w", err)
	}
	if orgRole == nil {
		return nil, fmt.Errorf("no default role configured for organization")
	}

	return orgRole, nil
}

// SetAsDefault sets a role as the default for new users in an organization
func (s *OrganizationRoleService) SetAsDefault(ctx context.Context, organizationID, roleID string) error {
	if organizationID == "" {
		return fmt.Errorf("organization ID is required")
	}
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}

	// Get the role to set as default
	targetRole, err := s.repo.GetByID(ctx, roleID)
	if err != nil || targetRole == nil {
		return fmt.Errorf("role not found")
	}

	// Verify role belongs to the organization
	if targetRole.OrganizationID != organizationID {
		return fmt.Errorf("role does not belong to organization")
	}

	// Get all organization roles
	allRoles, err := s.repo.GetByOrganizationID(ctx, organizationID)
	if err != nil {
		return fmt.Errorf("failed to get organization roles: %w", err)
	}

	// Unset IsDefault for all other roles
	for _, role := range allRoles {
		if role.ID != roleID && role.IsDefault {
			role.IsDefault = false
			if err := s.repo.Update(ctx, role); err != nil {
				return fmt.Errorf("failed to unset previous default role: %w", err)
			}
		}
	}

	// Set the target role as default
	targetRole.SetAsDefault()
	if err := s.repo.Update(ctx, targetRole); err != nil {
		return fmt.Errorf("failed to set default role: %w", err)
	}

	return nil
}

// UpdateOrganizationRole updates an existing organization role
func (s *OrganizationRoleService) UpdateOrganizationRole(ctx context.Context, orgRole *roles.OrganizationRole) (*roles.OrganizationRole, error) {
	if orgRole == nil {
		return nil, fmt.Errorf("organization role cannot be nil")
	}

	// Validate exists
	existing, err := s.repo.GetByID(ctx, orgRole.ID)
	if err != nil || existing == nil {
		return nil, fmt.Errorf("organization role not found")
	}

	// Validate max users if set
	if orgRole.MaxUsers != nil && *orgRole.MaxUsers <= 0 {
		return nil, fmt.Errorf("max users must be greater than 0")
	}

	if err := s.repo.Update(ctx, orgRole); err != nil {
		return nil, fmt.Errorf("failed to update organization role: %w", err)
	}

	return orgRole, nil
}

// UpdateCustomPermissions updates custom permissions for an organization role
func (s *OrganizationRoleService) UpdateCustomPermissions(ctx context.Context, id string, permissions map[string]bool) (*roles.OrganizationRole, error) {
	if id == "" {
		return nil, fmt.Errorf("organization role ID is required")
	}
	if permissions == nil {
		return nil, fmt.Errorf("permissions cannot be nil")
	}

	orgRole, err := s.repo.GetByID(ctx, id)
	if err != nil || orgRole == nil {
		return nil, fmt.Errorf("organization role not found")
	}

	// Marshal permissions to JSON
	permJSON, err := json.Marshal(permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal permissions: %w", err)
	}

	orgRole.CustomPermissions = string(permJSON)

	if err := s.repo.Update(ctx, orgRole); err != nil {
		return nil, fmt.Errorf("failed to update permissions: %w", err)
	}

	return orgRole, nil
}

// UpdateMaxUsers updates the maximum users limit for an organization role
func (s *OrganizationRoleService) UpdateMaxUsers(ctx context.Context, id string, maxUsers int) (*roles.OrganizationRole, error) {
	if id == "" {
		return nil, fmt.Errorf("organization role ID is required")
	}
	if maxUsers <= 0 {
		return nil, fmt.Errorf("max users must be greater than 0")
	}

	orgRole, err := s.repo.GetByID(ctx, id)
	if err != nil || orgRole == nil {
		return nil, fmt.Errorf("organization role not found")
	}

	orgRole.SetMaxUsers(maxUsers)

	if err := s.repo.Update(ctx, orgRole); err != nil {
		return nil, fmt.Errorf("failed to update max users: %w", err)
	}

	return orgRole, nil
}

// DeleteOrganizationRole soft deletes an organization role
func (s *OrganizationRoleService) DeleteOrganizationRole(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("organization role ID is required")
	}

	// Check exists
	orgRole, err := s.repo.GetByID(ctx, id)
	if err != nil || orgRole == nil {
		return fmt.Errorf("organization role not found")
	}

	// Prevent deletion of default role
	if orgRole.IsDefault {
		return fmt.Errorf("cannot delete default role, set another role as default first")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete organization role: %w", err)
	}

	return nil
}

// ListOrganizationRoles retrieves organization roles with pagination
func (s *OrganizationRoleService) ListOrganizationRoles(ctx context.Context, limit, offset int) ([]*roles.OrganizationRole, int, error) {
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	orgRoles, total, err := s.repo.List(ctx, nil, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list organization roles: %w", err)
	}

	return orgRoles, total, nil
}
