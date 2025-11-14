// Package services provides repository operations for service-related entities.
package services

import (
	"context"
	"fmt"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/services"
	"github.com/Kisanlink/kisanlink-ecom/internal/repositories/common"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// SLARepository handles service SLA operations using the database manager
type SLARepository struct {
	*common.BaseRepository
	dbManager db.DBManager
}

// NewSLARepository creates a new service SLA repository
func NewSLARepository(dbManager db.DBManager) *SLARepository {
	return &SLARepository{
		BaseRepository: common.NewBaseRepository(dbManager),
		dbManager:      dbManager,
	}
}

// Create creates a new service SLA
func (r *SLARepository) Create(ctx context.Context, sla *services.SLA) error {
	return r.dbManager.Create(ctx, sla)
}

// GetByID retrieves a service SLA by ID with soft delete filtering
func (r *SLARepository) GetByID(ctx context.Context, id string) (*services.SLA, error) {
	var item services.SLA
	opts := common.QueryOptionsFromContext(ctx)

	if opts.IncludeDeleted {
		if err := r.dbManager.GetByID(ctx, id, &item); err != nil {
			return nil, fmt.Errorf("failed to get service SLA: %w", err)
		}
		return &item, nil
	}

	// Filter out deleted items
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "id", Operator: base.OpEqual, Value: id},
	}

	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*services.SLA
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get service SLA: %w", err)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("service SLA not found")
	}

	return items[0], nil
}

// Update updates an existing service SLA
func (r *SLARepository) Update(ctx context.Context, sla *services.SLA) error {
	return r.dbManager.Update(ctx, sla)
}

// Delete soft deletes a service SLA
func (r *SLARepository) Delete(ctx context.Context, id string) error {
	item := &services.SLA{}
	return r.dbManager.Delete(ctx, id, item)
}

// List retrieves service SLAs with filtering and pagination
func (r *SLARepository) List(ctx context.Context, filter *base.Filter, limit, offset int) ([]*services.SLA, int, error) {
	dbFilter := filter
	if dbFilter == nil {
		dbFilter = base.NewFilter()
	}

	// Apply soft delete filtering
	dbFilter = r.ApplyQueryOptions(ctx, dbFilter)

	dbFilter.Limit = limit
	dbFilter.Offset = offset

	var items []*services.SLA
	if err := r.dbManager.List(ctx, dbFilter, &items); err != nil {
		return nil, 0, fmt.Errorf("failed to list service SLAs: %w", err)
	}

	total, err := r.dbManager.Count(ctx, dbFilter, &services.SLA{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count service SLAs: %w", err)
	}

	return items, int(total), nil
}

// GetByOrganizationID retrieves all SLAs for an organization (implements interface method)
func (r *SLARepository) GetByOrganizationID(ctx context.Context, orgID string) ([]*services.SLA, error) {
	return r.GetByOrganizationIDFiltered(ctx, orgID, false)
}

// GetByOrganizationIDFiltered retrieves all SLAs for an organization with optional active filtering
func (r *SLARepository) GetByOrganizationIDFiltered(ctx context.Context, orgID string, activeOnly bool) ([]*services.SLA, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "organization_id", Operator: base.OpEqual, Value: orgID},
	}

	// Filter for active SLAs only if requested
	if activeOnly {
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "is_active",
			Operator: base.OpEqual,
			Value:    true,
		})
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*services.SLA
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get SLAs by organization ID: %w", err)
	}

	return items, nil
}

// GetByCatalogItemID retrieves all SLAs for a catalog item (implements interface method)
func (r *SLARepository) GetByCatalogItemID(ctx context.Context, catalogItemID string) ([]*services.SLA, error) {
	return r.GetByCatalogItemIDFiltered(ctx, catalogItemID, false)
}

// GetByCatalogItemIDFiltered retrieves all SLAs for a catalog item with optional active filtering
func (r *SLARepository) GetByCatalogItemIDFiltered(ctx context.Context, catalogItemID string, activeOnly bool) ([]*services.SLA, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "catalog_item_id", Operator: base.OpEqual, Value: catalogItemID},
	}

	// Filter for active SLAs only if requested
	if activeOnly {
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "is_active",
			Operator: base.OpEqual,
			Value:    true,
		})
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*services.SLA
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get SLAs by catalog item ID: %w", err)
	}

	return items, nil
}

// GetBySLAType retrieves all SLAs of a specific type
func (r *SLARepository) GetBySLAType(ctx context.Context, orgID string, slaType services.SLAType, activeOnly bool) ([]*services.SLA, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "organization_id", Operator: base.OpEqual, Value: orgID},
		{Field: "type", Operator: base.OpEqual, Value: string(slaType)},
	}

	// Filter for active SLAs only if requested
	if activeOnly {
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "is_active",
			Operator: base.OpEqual,
			Value:    true,
		})
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*services.SLA
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get SLAs by type: %w", err)
	}

	return items, nil
}

// GetResponseTimeSLAs retrieves all response time SLAs for an organization
func (r *SLARepository) GetResponseTimeSLAs(ctx context.Context, orgID string, activeOnly bool) ([]*services.SLA, error) {
	return r.GetBySLAType(ctx, orgID, services.SLATypeResponse, activeOnly)
}

// GetAvailabilitySLAs retrieves all availability SLAs for an organization
func (r *SLARepository) GetAvailabilitySLAs(ctx context.Context, orgID string, activeOnly bool) ([]*services.SLA, error) {
	return r.GetBySLAType(ctx, orgID, services.SLATypeAvailability, activeOnly)
}

// GetResolutionTimeSLAs retrieves all resolution time SLAs for an organization
func (r *SLARepository) GetResolutionTimeSLAs(ctx context.Context, orgID string, activeOnly bool) ([]*services.SLA, error) {
	return r.GetBySLAType(ctx, orgID, services.SLATypeResolution, activeOnly)
}

// GetPerformanceSLAs retrieves all performance SLAs for an organization
func (r *SLARepository) GetPerformanceSLAs(ctx context.Context, orgID string, activeOnly bool) ([]*services.SLA, error) {
	return r.GetBySLAType(ctx, orgID, services.SLATypePerformance, activeOnly)
}

// GetByName retrieves an SLA by name within an organization
func (r *SLARepository) GetByName(ctx context.Context, orgID, name string) (*services.SLA, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "organization_id", Operator: base.OpEqual, Value: orgID},
		{Field: "name", Operator: base.OpEqual, Value: name},
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*services.SLA
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get SLA by name: %w", err)
	}

	if len(items) == 0 {
		return nil, nil
	}

	return items[0], nil
}

// GetSLAsByCatalogAndType retrieves SLAs for a catalog item filtered by type
func (r *SLARepository) GetSLAsByCatalogAndType(ctx context.Context, catalogItemID string, slaType services.SLAType, activeOnly bool) ([]*services.SLA, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "catalog_item_id", Operator: base.OpEqual, Value: catalogItemID},
		{Field: "type", Operator: base.OpEqual, Value: string(slaType)},
	}

	// Filter for active SLAs only if requested
	if activeOnly {
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "is_active",
			Operator: base.OpEqual,
			Value:    true,
		})
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*services.SLA
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get SLAs by catalog and type: %w", err)
	}

	return items, nil
}

// DeactivateSLA deactivates a service SLA
func (r *SLARepository) DeactivateSLA(ctx context.Context, id string) error {
	sla, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get service SLA for deactivation: %w", err)
	}

	sla.IsActive = false
	return r.Update(ctx, sla)
}

// ActivateSLA activates a service SLA
func (r *SLARepository) ActivateSLA(ctx context.Context, id string) error {
	sla, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get service SLA for activation: %w", err)
	}

	sla.IsActive = true
	return r.Update(ctx, sla)
}

// UpdateThresholds updates warning and critical thresholds for an SLA
func (r *SLARepository) UpdateThresholds(ctx context.Context, id string, warningThreshold, criticalThreshold *float64) error {
	sla, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get service SLA for threshold update: %w", err)
	}

	if warningThreshold != nil {
		sla.WarningThreshold = warningThreshold
	}
	if criticalThreshold != nil {
		sla.CriticalThreshold = criticalThreshold
	}

	return r.Update(ctx, sla)
}

// GetViolatedSLAs retrieves SLAs that might be violated (this is a helper that would be used with monitoring data)
// In real implementation, this would check against actual metrics
func (r *SLARepository) GetViolatedSLAs(ctx context.Context, orgID string) ([]*services.SLA, error) {
	// Get all active SLAs for the organization
	_, err := r.GetByOrganizationIDFiltered(ctx, orgID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get SLAs for violation check: %w", err)
	}

	// In a real implementation, we would check actual metrics against SLA targets
	// For now, return empty list as placeholder
	return []*services.SLA{}, nil
}
