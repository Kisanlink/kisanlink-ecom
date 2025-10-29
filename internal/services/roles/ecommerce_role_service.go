// Package roles provides service layer for ecommerce role management
package roles

import (
	"context"
	"fmt"

	"kisanlink-ecom/entities/models/roles"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// EcommerceRoleRepositoryInterface defines repository operations
type EcommerceRoleRepositoryInterface interface {
	Create(ctx context.Context, ecomRole *roles.EcommerceRole) error
	GetByID(ctx context.Context, id string) (*roles.EcommerceRole, error)
	GetByOrganizationID(ctx context.Context, organizationID string) ([]*roles.EcommerceRole, error)
	Update(ctx context.Context, ecomRole *roles.EcommerceRole) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter *base.Filter, limit, offset int) ([]*roles.EcommerceRole, int, error)
}

// EcommerceRoleServiceInterface defines service operations
type EcommerceRoleServiceInterface interface {
	CreateEcommerceRole(ctx context.Context, aaaRoleID, roleName, description string, orgID *string) (*roles.EcommerceRole, error)
	GetEcommerceRole(ctx context.Context, id string) (*roles.EcommerceRole, error)
	GetOrganizationRoles(ctx context.Context, organizationID string) ([]*roles.EcommerceRole, error)
	UpdateEcommerceRole(ctx context.Context, ecomRole *roles.EcommerceRole) (*roles.EcommerceRole, error)
	UpdatePermissions(ctx context.Context, id string, permissions map[string]bool) (*roles.EcommerceRole, error)
	SetOrganizationScope(ctx context.Context, id, organizationID string) (*roles.EcommerceRole, error)
	DeleteEcommerceRole(ctx context.Context, id string) error
	ListEcommerceRoles(ctx context.Context, limit, offset int) ([]*roles.EcommerceRole, int, error)
	CheckPermission(ctx context.Context, roleID, permission string) (bool, error)
	SyncFromAAA(ctx context.Context, aaaRoleID string) (*roles.EcommerceRole, error)
}

// EcommerceRoleService provides business logic for ecommerce role management
type EcommerceRoleService struct {
	repo EcommerceRoleRepositoryInterface
}

// NewEcommerceRoleService creates a new ecommerce role service
func NewEcommerceRoleService(repo EcommerceRoleRepositoryInterface) *EcommerceRoleService {
	return &EcommerceRoleService{repo: repo}
}

// CreateEcommerceRole creates a new ecommerce role
func (s *EcommerceRoleService) CreateEcommerceRole(ctx context.Context, aaaRoleID, roleName, description string, orgID *string) (*roles.EcommerceRole, error) {
	// Validate input
	if aaaRoleID == "" {
		return nil, fmt.Errorf("AAA role ID is required")
	}
	if roleName == "" {
		return nil, fmt.Errorf("role name is required")
	}

	// Check if role already exists
	existing, err := s.repo.GetByID(ctx, aaaRoleID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("ecommerce role with AAA role ID %s already exists", aaaRoleID)
	}

	// Create new ecommerce role
	ecomRole := roles.NewEcommerceRole(aaaRoleID, roleName, description)

	if orgID != nil && *orgID != "" {
		ecomRole.SetOrganizationScope(*orgID)
	}

	if err := s.repo.Create(ctx, ecomRole); err != nil {
		return nil, fmt.Errorf("failed to create ecommerce role: %w", err)
	}

	return ecomRole, nil
}

// GetEcommerceRole retrieves an ecommerce role by ID
func (s *EcommerceRoleService) GetEcommerceRole(ctx context.Context, id string) (*roles.EcommerceRole, error) {
	if id == "" {
		return nil, fmt.Errorf("ecommerce role ID is required")
	}

	ecomRole, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get ecommerce role: %w", err)
	}
	if ecomRole == nil {
		return nil, fmt.Errorf("ecommerce role not found")
	}

	return ecomRole, nil
}

// GetOrganizationRoles retrieves all ecommerce roles for an organization
func (s *EcommerceRoleService) GetOrganizationRoles(ctx context.Context, organizationID string) ([]*roles.EcommerceRole, error) {
	if organizationID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	ecomRoles, err := s.repo.GetByOrganizationID(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization roles: %w", err)
	}

	return ecomRoles, nil
}

// UpdateEcommerceRole updates an existing ecommerce role
func (s *EcommerceRoleService) UpdateEcommerceRole(ctx context.Context, ecomRole *roles.EcommerceRole) (*roles.EcommerceRole, error) {
	if ecomRole == nil {
		return nil, fmt.Errorf("ecommerce role cannot be nil")
	}

	// Validate exists
	existing, err := s.repo.GetByID(ctx, ecomRole.AAARoleID)
	if err != nil || existing == nil {
		return nil, fmt.Errorf("ecommerce role not found")
	}

	// Validate role name
	if ecomRole.RoleName == "" {
		return nil, fmt.Errorf("role name is required")
	}

	if err := s.repo.Update(ctx, ecomRole); err != nil {
		return nil, fmt.Errorf("failed to update ecommerce role: %w", err)
	}

	return ecomRole, nil
}

// UpdatePermissions updates the permissions for an ecommerce role
func (s *EcommerceRoleService) UpdatePermissions(ctx context.Context, id string, permissions map[string]bool) (*roles.EcommerceRole, error) {
	if id == "" {
		return nil, fmt.Errorf("ecommerce role ID is required")
	}
	if permissions == nil {
		return nil, fmt.Errorf("permissions cannot be nil")
	}

	ecomRole, err := s.repo.GetByID(ctx, id)
	if err != nil || ecomRole == nil {
		return nil, fmt.Errorf("ecommerce role not found")
	}

	// Update permissions based on the map
	if val, ok := permissions["can_manage_catalog"]; ok {
		ecomRole.CanManageCatalog = val
	}
	if val, ok := permissions["can_manage_orders"]; ok {
		ecomRole.CanManageOrders = val
	}
	if val, ok := permissions["can_manage_inventory"]; ok {
		ecomRole.CanManageInventory = val
	}
	if val, ok := permissions["can_manage_pricing"]; ok {
		ecomRole.CanManagePricing = val
	}
	if val, ok := permissions["can_manage_customers"]; ok {
		ecomRole.CanManageCustomers = val
	}
	if val, ok := permissions["can_view_analytics"]; ok {
		ecomRole.CanViewAnalytics = val
	}
	if val, ok := permissions["can_manage_users"]; ok {
		ecomRole.CanManageUsers = val
	}
	if val, ok := permissions["can_manage_settings"]; ok {
		ecomRole.CanManageSettings = val
	}

	if err := s.repo.Update(ctx, ecomRole); err != nil {
		return nil, fmt.Errorf("failed to update permissions: %w", err)
	}

	return ecomRole, nil
}

// SetOrganizationScope sets the organization scope for an ecommerce role
func (s *EcommerceRoleService) SetOrganizationScope(ctx context.Context, id, organizationID string) (*roles.EcommerceRole, error) {
	if id == "" {
		return nil, fmt.Errorf("ecommerce role ID is required")
	}
	if organizationID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	ecomRole, err := s.repo.GetByID(ctx, id)
	if err != nil || ecomRole == nil {
		return nil, fmt.Errorf("ecommerce role not found")
	}

	ecomRole.SetOrganizationScope(organizationID)

	if err := s.repo.Update(ctx, ecomRole); err != nil {
		return nil, fmt.Errorf("failed to set organization scope: %w", err)
	}

	return ecomRole, nil
}

// DeleteEcommerceRole soft deletes an ecommerce role
func (s *EcommerceRoleService) DeleteEcommerceRole(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("ecommerce role ID is required")
	}

	// Check exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("ecommerce role not found")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete ecommerce role: %w", err)
	}

	return nil
}

// ListEcommerceRoles retrieves ecommerce roles with pagination
func (s *EcommerceRoleService) ListEcommerceRoles(ctx context.Context, limit, offset int) ([]*roles.EcommerceRole, int, error) {
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	ecomRoles, total, err := s.repo.List(ctx, nil, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list ecommerce roles: %w", err)
	}

	return ecomRoles, total, nil
}

// CheckPermission checks if a role has a specific permission
func (s *EcommerceRoleService) CheckPermission(ctx context.Context, roleID, permission string) (bool, error) {
	if roleID == "" {
		return false, fmt.Errorf("role ID is required")
	}
	if permission == "" {
		return false, fmt.Errorf("permission is required")
	}

	ecomRole, err := s.repo.GetByID(ctx, roleID)
	if err != nil || ecomRole == nil {
		return false, fmt.Errorf("ecommerce role not found")
	}

	if !ecomRole.IsActive {
		return false, nil
	}

	return ecomRole.HasPermission(permission), nil
}

// SyncFromAAA syncs role data from AAA service
func (s *EcommerceRoleService) SyncFromAAA(ctx context.Context, aaaRoleID string) (*roles.EcommerceRole, error) {
	if aaaRoleID == "" {
		return nil, fmt.Errorf("AAA role ID is required")
	}

	// Get existing role
	ecomRole, err := s.repo.GetByID(ctx, aaaRoleID)
	if err != nil || ecomRole == nil {
		return nil, fmt.Errorf("ecommerce role not found")
	}

	// In a real implementation, this would fetch updated role data from AAA service
	// For now, we just return the existing role
	// TODO: Implement AAA service integration

	return ecomRole, nil
}
