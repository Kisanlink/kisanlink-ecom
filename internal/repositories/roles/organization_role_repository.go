package roles

import (
	"context"
	"fmt"

	"kisanlink-ecom/entities/models/roles"
	"kisanlink-ecom/internal/repositories/common"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// OrganizationRoleRepository handles organization role operations using the database manager
type OrganizationRoleRepository struct {
	*common.BaseRepository
	dbManager db.DBManager
}

// NewOrganizationRoleRepository creates a new organization role repository
func NewOrganizationRoleRepository(dbManager db.DBManager) *OrganizationRoleRepository {
	return &OrganizationRoleRepository{
		BaseRepository: common.NewBaseRepository(dbManager),
		dbManager:      dbManager,
	}
}

// Create creates a new organization role configuration
func (r *OrganizationRoleRepository) Create(ctx context.Context, orgRole *roles.OrganizationRole) error {
	return r.dbManager.Create(ctx, orgRole)
}

// GetByID retrieves an organization role by ID with soft delete filtering
func (r *OrganizationRoleRepository) GetByID(ctx context.Context, id string) (*roles.OrganizationRole, error) {
	var item roles.OrganizationRole
	opts := common.QueryOptionsFromContext(ctx)

	if opts.IncludeDeleted {
		if err := r.dbManager.GetByID(ctx, id, &item); err != nil {
			return nil, fmt.Errorf("failed to get organization role: %w", err)
		}
		return &item, nil
	}

	// Filter out deleted items
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "id", Operator: base.OpEqual, Value: id},
	}

	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*roles.OrganizationRole
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get organization role: %w", err)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("organization role not found")
	}

	return items[0], nil
}

// Update updates an existing organization role configuration
func (r *OrganizationRoleRepository) Update(ctx context.Context, orgRole *roles.OrganizationRole) error {
	return r.dbManager.Update(ctx, orgRole)
}

// Delete soft deletes an organization role configuration
func (r *OrganizationRoleRepository) Delete(ctx context.Context, id string) error {
	item := &roles.OrganizationRole{}
	return r.dbManager.Delete(ctx, id, item)
}

// List retrieves organization roles with filtering and pagination
func (r *OrganizationRoleRepository) List(ctx context.Context, filter *base.Filter, limit, offset int) ([]*roles.OrganizationRole, int, error) {
	dbFilter := filter
	if dbFilter == nil {
		dbFilter = base.NewFilter()
	}

	// Apply soft delete filtering
	dbFilter = r.ApplyQueryOptions(ctx, dbFilter)

	dbFilter.Limit = limit
	dbFilter.Offset = offset

	var items []*roles.OrganizationRole
	if err := r.dbManager.List(ctx, dbFilter, &items); err != nil {
		return nil, 0, fmt.Errorf("failed to list organization roles: %w", err)
	}

	total, err := r.dbManager.Count(ctx, dbFilter, &roles.OrganizationRole{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count organization roles: %w", err)
	}

	return items, int(total), nil
}

// GetByOrganizationID retrieves all roles configured for an organization (implements interface method)
func (r *OrganizationRoleRepository) GetByOrganizationID(ctx context.Context, organizationID string) ([]*roles.OrganizationRole, error) {
	return r.GetOrgRoles(ctx, organizationID, false)
}

// GetOrgRoles retrieves all roles configured for an organization
func (r *OrganizationRoleRepository) GetOrgRoles(ctx context.Context, organizationID string, activeOnly bool) ([]*roles.OrganizationRole, error) {
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

	var items []*roles.OrganizationRole
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get organization roles: %w", err)
	}

	return items, nil
}

// GetDefaultRole retrieves the default role for new users in an organization
func (r *OrganizationRoleRepository) GetDefaultRole(ctx context.Context, organizationID string) (*roles.OrganizationRole, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "organization_id", Operator: base.OpEqual, Value: organizationID},
		{Field: "is_default", Operator: base.OpEqual, Value: true},
		{Field: "is_active", Operator: base.OpEqual, Value: true},
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*roles.OrganizationRole
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get default role: %w", err)
	}

	if len(items) == 0 {
		return nil, nil
	}

	return items[0], nil
}

// SetAsDefault sets a role as the default for an organization (unsets other defaults)
func (r *OrganizationRoleRepository) SetAsDefault(ctx context.Context, organizationID, roleID string) error {
	// First, get all roles for the organization
	orgRoles, err := r.GetOrgRoles(ctx, organizationID, false)
	if err != nil {
		return fmt.Errorf("failed to get organization roles: %w", err)
	}

	// Unset default flag for all roles
	for _, role := range orgRoles {
		if role.IsDefault {
			role.IsDefault = false
			if err := r.Update(ctx, role); err != nil {
				return fmt.Errorf("failed to unset default flag: %w", err)
			}
		}
	}

	// Set the new default role
	targetRole, err := r.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get target role: %w", err)
	}

	if targetRole.OrganizationID != organizationID {
		return fmt.Errorf("role does not belong to organization")
	}

	targetRole.IsDefault = true
	if err := r.Update(ctx, targetRole); err != nil {
		return fmt.Errorf("failed to set default flag: %w", err)
	}

	return nil
}

// GetByAAARoleID retrieves all organization roles for a specific AAA role (implements interface method)
func (r *OrganizationRoleRepository) GetByAAARoleID(ctx context.Context, aaaRoleID string) ([]*roles.OrganizationRole, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "aaa_role_id", Operator: base.OpEqual, Value: aaaRoleID},
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*roles.OrganizationRole
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get organization roles by AAA role ID: %w", err)
	}

	return items, nil
}

// GetByOrgAndAAARoleID retrieves an organization role by organization ID and AAA role ID
func (r *OrganizationRoleRepository) GetByOrgAndAAARoleID(ctx context.Context, organizationID, aaaRoleID string) (*roles.OrganizationRole, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "organization_id", Operator: base.OpEqual, Value: organizationID},
		{Field: "aaa_role_id", Operator: base.OpEqual, Value: aaaRoleID},
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*roles.OrganizationRole
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get organization role by AAA role ID: %w", err)
	}

	if len(items) == 0 {
		return nil, nil
	}

	return items[0], nil
}

// DeactivateOrgRole deactivates an organization role
func (r *OrganizationRoleRepository) DeactivateOrgRole(ctx context.Context, id string) error {
	orgRole, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get organization role for deactivation: %w", err)
	}

	orgRole.IsActive = false
	return r.Update(ctx, orgRole)
}

// ActivateOrgRole activates an organization role
func (r *OrganizationRoleRepository) ActivateOrgRole(ctx context.Context, id string) error {
	orgRole, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get organization role for activation: %w", err)
	}

	orgRole.IsActive = true
	return r.Update(ctx, orgRole)
}
