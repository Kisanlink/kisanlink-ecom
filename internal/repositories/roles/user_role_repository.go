// Package roles provides repository operations for role-related entities.
package roles

import (
	"context"
	"fmt"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/roles"
	"github.com/Kisanlink/kisanlink-ecom/internal/repositories/common"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// UserRoleRepository handles user role operations using the database manager
type UserRoleRepository struct {
	*common.BaseRepository
	dbManager db.DBManager
}

// NewUserRoleRepository creates a new user role repository
func NewUserRoleRepository(dbManager db.DBManager) *UserRoleRepository {
	return &UserRoleRepository{
		BaseRepository: common.NewBaseRepository(dbManager),
		dbManager:      dbManager,
	}
}

// Create creates a new user role assignment
func (r *UserRoleRepository) Create(ctx context.Context, userRole *roles.UserRole) error {
	return r.dbManager.Create(ctx, userRole)
}

// GetByID retrieves a user role by ID with soft delete filtering
func (r *UserRoleRepository) GetByID(ctx context.Context, id string) (*roles.UserRole, error) {
	var item roles.UserRole
	opts := common.QueryOptionsFromContext(ctx)

	if opts.IncludeDeleted {
		if err := r.dbManager.GetByID(ctx, id, &item); err != nil {
			return nil, fmt.Errorf("failed to get user role: %w", err)
		}
		return &item, nil
	}

	// Filter out deleted items
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "id", Operator: base.OpEqual, Value: id},
	}

	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*roles.UserRole
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get user role: %w", err)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("user role not found")
	}

	return items[0], nil
}

// Update updates an existing user role assignment
func (r *UserRoleRepository) Update(ctx context.Context, userRole *roles.UserRole) error {
	return r.dbManager.Update(ctx, userRole)
}

// Delete soft deletes a user role assignment
func (r *UserRoleRepository) Delete(ctx context.Context, id string) error {
	item := &roles.UserRole{}
	return r.dbManager.Delete(ctx, id, item)
}

// List retrieves user roles with filtering and pagination
func (r *UserRoleRepository) List(ctx context.Context, filter *base.Filter, limit, offset int) ([]*roles.UserRole, int, error) {
	dbFilter := filter
	if dbFilter == nil {
		dbFilter = base.NewFilter()
	}

	// Apply soft delete filtering
	dbFilter = r.ApplyQueryOptions(ctx, dbFilter)

	dbFilter.Limit = limit
	dbFilter.Offset = offset

	var items []*roles.UserRole
	if err := r.dbManager.List(ctx, dbFilter, &items); err != nil {
		return nil, 0, fmt.Errorf("failed to list user roles: %w", err)
	}

	total, err := r.dbManager.Count(ctx, dbFilter, &roles.UserRole{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count user roles: %w", err)
	}

	return items, int(total), nil
}

// GetByUserID retrieves all roles assigned to a user (implements interface method)
func (r *UserRoleRepository) GetByUserID(ctx context.Context, userID string) ([]*roles.UserRole, error) {
	return r.GetUserRoles(ctx, userID, false)
}

// GetUserRoles retrieves all roles assigned to a user
func (r *UserRoleRepository) GetUserRoles(ctx context.Context, userID string, activeOnly bool) ([]*roles.UserRole, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "user_id", Operator: base.OpEqual, Value: userID},
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

	var items []*roles.UserRole
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	return items, nil
}

// GetByRoleID retrieves all users with a specific role (implements interface method)
func (r *UserRoleRepository) GetByRoleID(ctx context.Context, roleID string) ([]*roles.UserRole, error) {
	return r.GetRoleUsers(ctx, roleID, false)
}

// GetRoleUsers retrieves all users with a specific role
func (r *UserRoleRepository) GetRoleUsers(ctx context.Context, roleID string, activeOnly bool) ([]*roles.UserRole, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "role_id", Operator: base.OpEqual, Value: roleID},
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

	var items []*roles.UserRole
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get role users: %w", err)
	}

	return items, nil
}

// GetActiveUserRole retrieves the active role assignment for a user-role combination
func (r *UserRoleRepository) GetActiveUserRole(ctx context.Context, userID, roleID string) (*roles.UserRole, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "user_id", Operator: base.OpEqual, Value: userID},
		{Field: "role_id", Operator: base.OpEqual, Value: roleID},
		{Field: "is_active", Operator: base.OpEqual, Value: true},
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*roles.UserRole
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get active user role: %w", err)
	}

	if len(items) == 0 {
		return nil, nil
	}

	return items[0], nil
}

// DeactivateUserRole deactivates a user role assignment
func (r *UserRoleRepository) DeactivateUserRole(ctx context.Context, id string) error {
	userRole, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user role for deactivation: %w", err)
	}

	userRole.IsActive = false
	return r.Update(ctx, userRole)
}

// ActivateUserRole activates a user role assignment
func (r *UserRoleRepository) ActivateUserRole(ctx context.Context, id string) error {
	userRole, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user role for activation: %w", err)
	}

	userRole.IsActive = true
	return r.Update(ctx, userRole)
}
