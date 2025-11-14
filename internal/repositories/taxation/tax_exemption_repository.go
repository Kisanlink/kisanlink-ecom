// Package taxation provides repository operations for taxation-related entities.
package taxation

import (
	"context"
	"fmt"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/taxation"
	"github.com/Kisanlink/kisanlink-ecom/internal/repositories/common"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// TaxExemptionRepository handles tax exemption operations using the database manager
type TaxExemptionRepository struct {
	*common.BaseRepository
	dbManager db.DBManager
}

// NewTaxExemptionRepository creates a new tax exemption repository
func NewTaxExemptionRepository(dbManager db.DBManager) *TaxExemptionRepository {
	return &TaxExemptionRepository{
		BaseRepository: common.NewBaseRepository(dbManager),
		dbManager:      dbManager,
	}
}

// Create creates a new tax exemption
func (r *TaxExemptionRepository) Create(ctx context.Context, exemption *taxation.TaxExemption) error {
	return r.dbManager.Create(ctx, exemption)
}

// GetByID retrieves a tax exemption by ID with soft delete filtering
func (r *TaxExemptionRepository) GetByID(ctx context.Context, id string) (*taxation.TaxExemption, error) {
	var item taxation.TaxExemption
	opts := common.QueryOptionsFromContext(ctx)

	if opts.IncludeDeleted {
		if err := r.dbManager.GetByID(ctx, id, &item); err != nil {
			return nil, fmt.Errorf("failed to get tax exemption: %w", err)
		}
		return &item, nil
	}

	// Filter out deleted items
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "id", Operator: base.OpEqual, Value: id},
	}

	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*taxation.TaxExemption
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get tax exemption: %w", err)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("tax exemption not found")
	}

	return items[0], nil
}

// Update updates an existing tax exemption
func (r *TaxExemptionRepository) Update(ctx context.Context, exemption *taxation.TaxExemption) error {
	return r.dbManager.Update(ctx, exemption)
}

// Delete soft deletes a tax exemption
func (r *TaxExemptionRepository) Delete(ctx context.Context, id string) error {
	item := &taxation.TaxExemption{}
	return r.dbManager.Delete(ctx, id, item)
}

// List retrieves tax exemptions with filtering and pagination
func (r *TaxExemptionRepository) List(ctx context.Context, filter *base.Filter, limit, offset int) ([]*taxation.TaxExemption, int, error) {
	dbFilter := filter
	if dbFilter == nil {
		dbFilter = base.NewFilter()
	}

	// Apply soft delete filtering
	dbFilter = r.ApplyQueryOptions(ctx, dbFilter)

	dbFilter.Limit = limit
	dbFilter.Offset = offset

	var items []*taxation.TaxExemption
	if err := r.dbManager.List(ctx, dbFilter, &items); err != nil {
		return nil, 0, fmt.Errorf("failed to list tax exemptions: %w", err)
	}

	total, err := r.dbManager.Count(ctx, dbFilter, &taxation.TaxExemption{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tax exemptions: %w", err)
	}

	return items, int(total), nil
}

// GetByExemptionID retrieves a tax exemption by exemption ID
func (r *TaxExemptionRepository) GetByExemptionID(ctx context.Context, exemptionID string) (*taxation.TaxExemption, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "exemption_id", Operator: base.OpEqual, Value: exemptionID},
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*taxation.TaxExemption
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get tax exemption by exemption ID: %w", err)
	}

	if len(items) == 0 {
		return nil, nil
	}

	return items[0], nil
}

// GetByOrganizationID retrieves all tax exemptions for an organization (implements interface method)
func (r *TaxExemptionRepository) GetByOrganizationID(ctx context.Context, orgID string) ([]*taxation.TaxExemption, error) {
	return r.GetByOrgID(ctx, orgID, false)
}

// GetByOrgID retrieves all tax exemptions for an organization
func (r *TaxExemptionRepository) GetByOrgID(ctx context.Context, orgID string, activeOnly bool) ([]*taxation.TaxExemption, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "org_id", Operator: base.OpEqual, Value: orgID},
	}

	// Filter for active exemptions only if requested
	if activeOnly {
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "is_active",
			Operator: base.OpEqual,
			Value:    true,
		})
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*taxation.TaxExemption
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get tax exemptions by org ID: %w", err)
	}

	return items, nil
}

// GetValidExemptions retrieves all currently valid tax exemptions for an organization
func (r *TaxExemptionRepository) GetValidExemptions(ctx context.Context, orgID string) ([]*taxation.TaxExemption, error) {
	exemptions, err := r.GetByOrgID(ctx, orgID, true)
	if err != nil {
		return nil, err
	}

	// Filter by validity in memory since IsValid() is a method
	var validExemptions []*taxation.TaxExemption
	for _, exemption := range exemptions {
		if exemption.IsValid() {
			validExemptions = append(validExemptions, exemption)
		}
	}

	return validExemptions, nil
}

// GetApplicableExemptions retrieves exemptions applicable to a specific entity
func (r *TaxExemptionRepository) GetApplicableExemptions(ctx context.Context, orgID, entityType, category, hsnCode string) ([]*taxation.TaxExemption, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "org_id", Operator: base.OpEqual, Value: orgID},
		{Field: "is_active", Operator: base.OpEqual, Value: true},
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*taxation.TaxExemption
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get applicable exemptions: %w", err)
	}

	// Filter by applicability and validity in memory
	var applicableExemptions []*taxation.TaxExemption
	for _, exemption := range items {
		if !exemption.IsValid() {
			continue
		}

		// Check if exemption applies to this entity
		if exemption.EntityType != "" && exemption.EntityType != "all" && exemption.EntityType != entityType {
			continue
		}

		if exemption.Category != "" && exemption.Category != category {
			continue
		}

		if exemption.HSNCode != "" && exemption.HSNCode != hsnCode {
			continue
		}

		applicableExemptions = append(applicableExemptions, exemption)
	}

	return applicableExemptions, nil
}

// GetByEntityType retrieves all exemptions for a specific entity type
func (r *TaxExemptionRepository) GetByEntityType(ctx context.Context, orgID, entityType string, activeOnly bool) ([]*taxation.TaxExemption, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "org_id", Operator: base.OpEqual, Value: orgID},
		{Field: "entity_type", Operator: base.OpEqual, Value: entityType},
	}

	// Filter for active exemptions only if requested
	if activeOnly {
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "is_active",
			Operator: base.OpEqual,
			Value:    true,
		})
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*taxation.TaxExemption
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get exemptions by entity type: %w", err)
	}

	return items, nil
}

// GetByHSNCode retrieves all exemptions for a specific HSN code
func (r *TaxExemptionRepository) GetByHSNCode(ctx context.Context, orgID, hsnCode string, activeOnly bool) ([]*taxation.TaxExemption, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "org_id", Operator: base.OpEqual, Value: orgID},
		{Field: "hsn_code", Operator: base.OpEqual, Value: hsnCode},
	}

	// Filter for active exemptions only if requested
	if activeOnly {
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "is_active",
			Operator: base.OpEqual,
			Value:    true,
		})
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*taxation.TaxExemption
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get exemptions by HSN code: %w", err)
	}

	return items, nil
}

// DeactivateExemption deactivates a tax exemption
func (r *TaxExemptionRepository) DeactivateExemption(ctx context.Context, id string) error {
	exemption, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get tax exemption for deactivation: %w", err)
	}

	exemption.IsActive = false
	return r.Update(ctx, exemption)
}

// ActivateExemption activates a tax exemption
func (r *TaxExemptionRepository) ActivateExemption(ctx context.Context, id string) error {
	exemption, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get tax exemption for activation: %w", err)
	}

	exemption.IsActive = true
	return r.Update(ctx, exemption)
}
