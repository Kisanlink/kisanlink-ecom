// Package actors provides repository operations for actor-related entities.
package actors

import (
	"context"
	"fmt"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/actors"
	"github.com/Kisanlink/kisanlink-ecom/internal/repositories/common"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// CollaboratorRepository handles organization collaborator operations using the database manager
type CollaboratorRepository struct {
	*common.BaseRepository
	dbManager db.DBManager
}

// NewCollaboratorRepository creates a new organization collaborator repository
func NewCollaboratorRepository(dbManager db.DBManager) *CollaboratorRepository {
	return &CollaboratorRepository{
		BaseRepository: common.NewBaseRepository(dbManager),
		dbManager:      dbManager,
	}
}

// Create creates a new organization collaborator
func (r *CollaboratorRepository) Create(ctx context.Context, collaborator *actors.Collaborator) error {
	return r.dbManager.Create(ctx, collaborator)
}

// GetByID retrieves an organization collaborator by ID with soft delete filtering
func (r *CollaboratorRepository) GetByID(ctx context.Context, id string) (*actors.Collaborator, error) {
	var item actors.Collaborator
	opts := common.QueryOptionsFromContext(ctx)

	if opts.IncludeDeleted {
		if err := r.dbManager.GetByID(ctx, id, &item); err != nil {
			return nil, fmt.Errorf("failed to get organization collaborator: %w", err)
		}
		return &item, nil
	}

	// Filter out deleted items
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "id", Operator: base.OpEqual, Value: id},
	}

	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*actors.Collaborator
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get organization collaborator: %w", err)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("organization collaborator not found")
	}

	return items[0], nil
}

// Update updates an existing organization collaborator
func (r *CollaboratorRepository) Update(ctx context.Context, collaborator *actors.Collaborator) error {
	return r.dbManager.Update(ctx, collaborator)
}

// Delete soft deletes an organization collaborator
func (r *CollaboratorRepository) Delete(ctx context.Context, id string) error {
	item := &actors.Collaborator{}
	return r.dbManager.Delete(ctx, id, item)
}

// List retrieves organization collaborators with filtering and pagination
func (r *CollaboratorRepository) List(ctx context.Context, filter *base.Filter, limit, offset int) ([]*actors.Collaborator, int, error) {
	dbFilter := filter
	if dbFilter == nil {
		dbFilter = base.NewFilter()
	}

	// Apply soft delete filtering
	dbFilter = r.ApplyQueryOptions(ctx, dbFilter)

	dbFilter.Limit = limit
	dbFilter.Offset = offset

	var items []*actors.Collaborator
	if err := r.dbManager.List(ctx, dbFilter, &items); err != nil {
		return nil, 0, fmt.Errorf("failed to list organization collaborators: %w", err)
	}

	total, err := r.dbManager.Count(ctx, dbFilter, &actors.Collaborator{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count organization collaborators: %w", err)
	}

	return items, int(total), nil
}

// GetByContextOrganizationID retrieves all collaborators for a context organization (implements interface method)
func (r *CollaboratorRepository) GetByContextOrganizationID(ctx context.Context, contextOrgID string) ([]*actors.Collaborator, error) {
	return r.GetByContextOrganizationIDFiltered(ctx, contextOrgID, nil)
}

// GetByContextOrganizationIDFiltered retrieves all collaborators for a context organization with optional status filtering
func (r *CollaboratorRepository) GetByContextOrganizationIDFiltered(ctx context.Context, contextOrgID string, statusFilter *actors.CollaboratorStatus) ([]*actors.Collaborator, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "context_organization_id", Operator: base.OpEqual, Value: contextOrgID},
	}

	// Filter by status if provided
	if statusFilter != nil {
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "status",
			Operator: base.OpEqual,
			Value:    string(*statusFilter),
		})
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*actors.Collaborator
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get collaborators by context organization: %w", err)
	}

	return items, nil
}

// GetByAAAEntityID retrieves all collaborators by AAA entity ID and type (implements interface method)
func (r *CollaboratorRepository) GetByAAAEntityID(ctx context.Context, aaaEntityID string, aaaEntityType actors.AAAEntityType) ([]*actors.Collaborator, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "aaa_entity_id", Operator: base.OpEqual, Value: aaaEntityID},
		{Field: "aaa_entity_type", Operator: base.OpEqual, Value: string(aaaEntityType)},
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*actors.Collaborator
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get collaborators by AAA entity: %w", err)
	}

	return items, nil
}

// GetByAAAEntityIDInOrg retrieves a collaborator by AAA entity ID, type, and organization
func (r *CollaboratorRepository) GetByAAAEntityIDInOrg(ctx context.Context, aaaEntityID string, aaaEntityType actors.AAAEntityType, contextOrgID string) (*actors.Collaborator, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "aaa_entity_id", Operator: base.OpEqual, Value: aaaEntityID},
		{Field: "aaa_entity_type", Operator: base.OpEqual, Value: string(aaaEntityType)},
		{Field: "context_organization_id", Operator: base.OpEqual, Value: contextOrgID},
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*actors.Collaborator
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get collaborator by AAA entity in organization: %w", err)
	}

	if len(items) == 0 {
		return nil, nil
	}

	return items[0], nil
}

// GetByUserID retrieves all collaborations for a user across organizations
func (r *CollaboratorRepository) GetByUserID(ctx context.Context, userID string, statusFilter *actors.CollaboratorStatus) ([]*actors.Collaborator, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "user_id", Operator: base.OpEqual, Value: userID},
	}

	// Filter by status if provided
	if statusFilter != nil {
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "status",
			Operator: base.OpEqual,
			Value:    string(*statusFilter),
		})
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*actors.Collaborator
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get collaborators by user ID: %w", err)
	}

	return items, nil
}

// GetByRole retrieves all collaborators with a specific role in an organization
func (r *CollaboratorRepository) GetByRole(ctx context.Context, contextOrgID string, role actors.CollaboratorRole, statusFilter *actors.CollaboratorStatus) ([]*actors.Collaborator, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "context_organization_id", Operator: base.OpEqual, Value: contextOrgID},
		{Field: "role", Operator: base.OpEqual, Value: string(role)},
	}

	// Filter by status if provided
	if statusFilter != nil {
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "status",
			Operator: base.OpEqual,
			Value:    string(*statusFilter),
		})
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*actors.Collaborator
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get collaborators by role: %w", err)
	}

	return items, nil
}

// GetActiveCollaborators retrieves all active collaborators for an organization
func (r *CollaboratorRepository) GetActiveCollaborators(ctx context.Context, contextOrgID string) ([]*actors.Collaborator, error) {
	activeStatus := actors.CollaboratorStatusActive
	return r.GetByContextOrganizationIDFiltered(ctx, contextOrgID, &activeStatus)
}

// GetByDepartment retrieves all collaborators in a specific department
func (r *CollaboratorRepository) GetByDepartment(ctx context.Context, contextOrgID, department string, statusFilter *actors.CollaboratorStatus) ([]*actors.Collaborator, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "context_organization_id", Operator: base.OpEqual, Value: contextOrgID},
		{Field: "department", Operator: base.OpEqual, Value: department},
	}

	// Filter by status if provided
	if statusFilter != nil {
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "status",
			Operator: base.OpEqual,
			Value:    string(*statusFilter),
		})
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*actors.Collaborator
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get collaborators by department: %w", err)
	}

	return items, nil
}

// GetByEmployeeID retrieves a collaborator by employee ID
func (r *CollaboratorRepository) GetByEmployeeID(ctx context.Context, employeeID string) (*actors.Collaborator, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "employee_id", Operator: base.OpEqual, Value: employeeID},
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*actors.Collaborator
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get collaborator by employee ID: %w", err)
	}

	if len(items) == 0 {
		return nil, nil
	}

	return items[0], nil
}

// GetUserCollaborators retrieves all user-type collaborators for an organization
func (r *CollaboratorRepository) GetUserCollaborators(ctx context.Context, contextOrgID string, statusFilter *actors.CollaboratorStatus) ([]*actors.Collaborator, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "context_organization_id", Operator: base.OpEqual, Value: contextOrgID},
		{Field: "aaa_entity_type", Operator: base.OpEqual, Value: string(actors.AAAEntityTypeUser)},
	}

	// Filter by status if provided
	if statusFilter != nil {
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "status",
			Operator: base.OpEqual,
			Value:    string(*statusFilter),
		})
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*actors.Collaborator
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get user collaborators: %w", err)
	}

	return items, nil
}

// GetOrganizationCollaborators retrieves all organization-type collaborators for an organization
func (r *CollaboratorRepository) GetOrganizationCollaborators(ctx context.Context, contextOrgID string, statusFilter *actors.CollaboratorStatus) ([]*actors.Collaborator, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "context_organization_id", Operator: base.OpEqual, Value: contextOrgID},
		{Field: "aaa_entity_type", Operator: base.OpEqual, Value: string(actors.AAAEntityTypeOrganization)},
	}

	// Filter by status if provided
	if statusFilter != nil {
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "status",
			Operator: base.OpEqual,
			Value:    string(*statusFilter),
		})
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*actors.Collaborator
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get organization collaborators: %w", err)
	}

	return items, nil
}

// ActivateCollaborator activates a collaborator
func (r *CollaboratorRepository) ActivateCollaborator(ctx context.Context, id string) error {
	collaborator, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get collaborator for activation: %w", err)
	}

	collaborator.Activate()
	return r.Update(ctx, collaborator)
}

// DeactivateCollaborator deactivates a collaborator
func (r *CollaboratorRepository) DeactivateCollaborator(ctx context.Context, id string) error {
	collaborator, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get collaborator for deactivation: %w", err)
	}

	collaborator.Deactivate()
	return r.Update(ctx, collaborator)
}

// SuspendCollaborator suspends a collaborator
func (r *CollaboratorRepository) SuspendCollaborator(ctx context.Context, id string) error {
	collaborator, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get collaborator for suspension: %w", err)
	}

	collaborator.Suspend()
	return r.Update(ctx, collaborator)
}

// UpdateAccessLevel updates the access level of a collaborator
func (r *CollaboratorRepository) UpdateAccessLevel(ctx context.Context, id string, accessLevel int) error {
	if accessLevel < 1 || accessLevel > 10 {
		return fmt.Errorf("invalid access level: must be between 1 and 10")
	}

	collaborator, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get collaborator for access level update: %w", err)
	}

	collaborator.AccessLevel = accessLevel
	return r.Update(ctx, collaborator)
}

// UpdateRole updates the role of a collaborator
func (r *CollaboratorRepository) UpdateRole(ctx context.Context, id string, role actors.CollaboratorRole) error {
	collaborator, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get collaborator for role update: %w", err)
	}

	collaborator.Role = role
	return r.Update(ctx, collaborator)
}

// SearchCollaborators searches for collaborators by name or email
func (r *CollaboratorRepository) SearchCollaborators(ctx context.Context, contextOrgID, searchTerm string, statusFilter *actors.CollaboratorStatus) ([]*actors.Collaborator, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "context_organization_id", Operator: base.OpEqual, Value: contextOrgID},
	}

	// Add search conditions for display_name and email
	searchGroup := base.FilterGroup{
		Logic: base.LogicOr,
		Conditions: []base.FilterCondition{
			{Field: "display_name", Operator: base.OpLike, Value: "%" + searchTerm + "%"},
			{Field: "email", Operator: base.OpLike, Value: "%" + searchTerm + "%"},
		},
	}
	filter.Group.Groups = []base.FilterGroup{searchGroup}

	// Filter by status if provided
	if statusFilter != nil {
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "status",
			Operator: base.OpEqual,
			Value:    string(*statusFilter),
		})
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*actors.Collaborator
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to search collaborators: %w", err)
	}

	return items, nil
}
