package roles

import (
	"context"
	"fmt"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/roles"
	"github.com/Kisanlink/kisanlink-ecom/internal/repositories/common"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// EcommerceRoleRepository handles ecommerce role operations using the database manager
type EcommerceRoleRepository struct {
	*common.BaseRepository
	dbManager db.DBManager
}

// NewEcommerceRoleRepository creates a new ecommerce role repository
func NewEcommerceRoleRepository(dbManager db.DBManager) *EcommerceRoleRepository {
	return &EcommerceRoleRepository{
		BaseRepository: common.NewBaseRepository(dbManager),
		dbManager:      dbManager,
	}
}

// Create creates a new ecommerce role
func (r *EcommerceRoleRepository) Create(ctx context.Context, role *roles.EcommerceRole) error {
	return r.dbManager.Create(ctx, role)
}

// GetByID retrieves an ecommerce role by ID (AAA Role ID) with soft delete filtering
func (r *EcommerceRoleRepository) GetByID(ctx context.Context, id string) (*roles.EcommerceRole, error) {
	var item roles.EcommerceRole
	opts := common.QueryOptionsFromContext(ctx)

	if opts.IncludeDeleted {
		if err := r.dbManager.GetByID(ctx, id, &item); err != nil {
			return nil, fmt.Errorf("failed to get ecommerce role: %w", err)
		}
		return &item, nil
	}

	// Filter out deleted items
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "id", Operator: base.OpEqual, Value: id},
	}

	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*roles.EcommerceRole
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get ecommerce role: %w", err)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("ecommerce role not found")
	}

	return items[0], nil
}

// Update updates an existing ecommerce role
func (r *EcommerceRoleRepository) Update(ctx context.Context, role *roles.EcommerceRole) error {
	return r.dbManager.Update(ctx, role)
}

// Delete soft deletes an ecommerce role
func (r *EcommerceRoleRepository) Delete(ctx context.Context, id string) error {
	item := &roles.EcommerceRole{}
	return r.dbManager.Delete(ctx, id, item)
}

// List retrieves ecommerce roles with filtering and pagination
func (r *EcommerceRoleRepository) List(ctx context.Context, filter *base.Filter, limit, offset int) ([]*roles.EcommerceRole, int, error) {
	dbFilter := filter
	if dbFilter == nil {
		dbFilter = base.NewFilter()
	}

	// Apply soft delete filtering
	dbFilter = r.ApplyQueryOptions(ctx, dbFilter)

	dbFilter.Limit = limit
	dbFilter.Offset = offset

	var items []*roles.EcommerceRole
	if err := r.dbManager.List(ctx, dbFilter, &items); err != nil {
		return nil, 0, fmt.Errorf("failed to list ecommerce roles: %w", err)
	}

	total, err := r.dbManager.Count(ctx, dbFilter, &roles.EcommerceRole{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count ecommerce roles: %w", err)
	}

	return items, int(total), nil
}

// GetByRoleName retrieves an ecommerce role by role name
func (r *EcommerceRoleRepository) GetByRoleName(ctx context.Context, roleName string) (*roles.EcommerceRole, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "role_name", Operator: base.OpEqual, Value: roleName},
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*roles.EcommerceRole
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get ecommerce role by name: %w", err)
	}

	if len(items) == 0 {
		return nil, nil
	}

	return items[0], nil
}

// GetByOrganizationID retrieves all roles scoped to a specific organization (implements interface method)
func (r *EcommerceRoleRepository) GetByOrganizationID(ctx context.Context, organizationID string) ([]*roles.EcommerceRole, error) {
	return r.GetOrgScopedRoles(ctx, organizationID, false)
}

// GetOrgScopedRoles retrieves all roles scoped to a specific organization
func (r *EcommerceRoleRepository) GetOrgScopedRoles(ctx context.Context, organizationID string, activeOnly bool) ([]*roles.EcommerceRole, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "organization_id", Operator: base.OpEqual, Value: organizationID},
	}

	// Filter for active roles only if requested
	if activeOnly {
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "is_active",
			Operator: base.OpEqual,
			Value:    true,
		})
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*roles.EcommerceRole
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get org-scoped roles: %w", err)
	}

	return items, nil
}

// GetPlatformRoles retrieves all platform-wide roles (not org-scoped)
func (r *EcommerceRoleRepository) GetPlatformRoles(ctx context.Context, activeOnly bool) ([]*roles.EcommerceRole, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "organization_id", Operator: base.OpIsNull, Value: nil},
	}

	// Filter for active roles only if requested
	if activeOnly {
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "is_active",
			Operator: base.OpEqual,
			Value:    true,
		})
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*roles.EcommerceRole
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get platform roles: %w", err)
	}

	return items, nil
}

// GetRolesWithPermission retrieves all roles that have a specific permission
func (r *EcommerceRoleRepository) GetRolesWithPermission(ctx context.Context, permission string, activeOnly bool) ([]*roles.EcommerceRole, error) {
	// Get all roles and filter by permission in memory
	// This is necessary because permission checks are method-based, not field-based
	var filter *base.Filter
	if activeOnly {
		filter = base.NewFilter()
		filter.Group.Conditions = []base.FilterCondition{
			{Field: "is_active", Operator: base.OpEqual, Value: true},
		}
	}

	// Apply soft delete filtering
	if filter == nil {
		filter = base.NewFilter()
	}
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*roles.EcommerceRole
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	// Filter by permission
	var result []*roles.EcommerceRole
	for _, role := range items {
		if role.HasPermission(permission) {
			result = append(result, role)
		}
	}

	return result, nil
}

// DeactivateRole deactivates an ecommerce role
func (r *EcommerceRoleRepository) DeactivateRole(ctx context.Context, id string) error {
	role, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get ecommerce role for deactivation: %w", err)
	}

	role.IsActive = false
	return r.Update(ctx, role)
}

// ActivateRole activates an ecommerce role
func (r *EcommerceRoleRepository) ActivateRole(ctx context.Context, id string) error {
	role, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get ecommerce role for activation: %w", err)
	}

	role.IsActive = true
	return r.Update(ctx, role)
}

// SyncFromAAA synchronizes role data from the AAA service
// This is a placeholder - actual AAA sync logic would be implemented in the service layer
func (r *EcommerceRoleRepository) SyncFromAAA(ctx context.Context, aaaRoleID, roleName, description string) error {
	// Check if role already exists
	existingRole, err := r.GetByID(ctx, aaaRoleID)
	if err == nil && existingRole != nil {
		// Update existing role
		existingRole.RoleName = roleName
		existingRole.Description = description
		return r.Update(ctx, existingRole)
	}

	// Create new role
	newRole := roles.NewEcommerceRole(aaaRoleID, roleName, description)
	return r.Create(ctx, newRole)
}
