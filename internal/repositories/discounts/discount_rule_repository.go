// Package discounts provides repository operations for discount-related entities.
package discounts

import (
	"context"
	"fmt"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/discounts"
	"github.com/Kisanlink/kisanlink-ecom/internal/repositories/common"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// DiscountRuleRepository handles discount rule operations using the database manager
type DiscountRuleRepository struct {
	*common.BaseRepository
	dbManager db.DBManager
}

// NewDiscountRuleRepository creates a new discount rule repository
func NewDiscountRuleRepository(dbManager db.DBManager) *DiscountRuleRepository {
	return &DiscountRuleRepository{
		BaseRepository: common.NewBaseRepository(dbManager),
		dbManager:      dbManager,
	}
}

// Create creates a new discount rule
func (r *DiscountRuleRepository) Create(ctx context.Context, rule *discounts.DiscountRule) error {
	return r.dbManager.Create(ctx, rule)
}

// GetByID retrieves a discount rule by ID with soft delete filtering
func (r *DiscountRuleRepository) GetByID(ctx context.Context, id string) (*discounts.DiscountRule, error) {
	var item discounts.DiscountRule
	opts := common.QueryOptionsFromContext(ctx)

	if opts.IncludeDeleted {
		if err := r.dbManager.GetByID(ctx, id, &item); err != nil {
			return nil, fmt.Errorf("failed to get discount rule: %w", err)
		}
		return &item, nil
	}

	// Filter out deleted items
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "id", Operator: base.OpEqual, Value: id},
	}

	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*discounts.DiscountRule
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get discount rule: %w", err)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("discount rule not found")
	}

	return items[0], nil
}

// Update updates an existing discount rule
func (r *DiscountRuleRepository) Update(ctx context.Context, rule *discounts.DiscountRule) error {
	return r.dbManager.Update(ctx, rule)
}

// Delete soft deletes a discount rule
func (r *DiscountRuleRepository) Delete(ctx context.Context, id string) error {
	item := &discounts.DiscountRule{}
	return r.dbManager.Delete(ctx, id, item)
}

// List retrieves discount rules with filtering and pagination
func (r *DiscountRuleRepository) List(ctx context.Context, filter *base.Filter, limit, offset int) ([]*discounts.DiscountRule, int, error) {
	dbFilter := filter
	if dbFilter == nil {
		dbFilter = base.NewFilter()
	}

	// Apply soft delete filtering
	dbFilter = r.ApplyQueryOptions(ctx, dbFilter)

	dbFilter.Limit = limit
	dbFilter.Offset = offset

	var items []*discounts.DiscountRule
	if err := r.dbManager.List(ctx, dbFilter, &items); err != nil {
		return nil, 0, fmt.Errorf("failed to list discount rules: %w", err)
	}

	total, err := r.dbManager.Count(ctx, dbFilter, &discounts.DiscountRule{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count discount rules: %w", err)
	}

	return items, int(total), nil
}

// GetByOrganizationID retrieves all discount rules for an organization (implements interface method)
func (r *DiscountRuleRepository) GetByOrganizationID(ctx context.Context, orgID string) ([]*discounts.DiscountRule, error) {
	return r.GetByOrgID(ctx, orgID, false)
}

// GetByOrgID retrieves all discount rules for an organization
func (r *DiscountRuleRepository) GetByOrgID(ctx context.Context, orgID string, activeOnly bool) ([]*discounts.DiscountRule, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "org_id", Operator: base.OpEqual, Value: orgID},
	}

	// Filter for active rules only if requested
	if activeOnly {
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "is_active",
			Operator: base.OpEqual,
			Value:    true,
		})
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*discounts.DiscountRule
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get discount rules by org ID: %w", err)
	}

	return items, nil
}

// GetByRuleType retrieves all discount rules of a specific type
func (r *DiscountRuleRepository) GetByRuleType(ctx context.Context, orgID, ruleType string, activeOnly bool) ([]*discounts.DiscountRule, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "org_id", Operator: base.OpEqual, Value: orgID},
		{Field: "rule_type", Operator: base.OpEqual, Value: ruleType},
	}

	// Filter for active rules only if requested
	if activeOnly {
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "is_active",
			Operator: base.OpEqual,
			Value:    true,
		})
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*discounts.DiscountRule
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get discount rules by type: %w", err)
	}

	return items, nil
}

// GetByPriority retrieves discount rules sorted by priority (highest first)
func (r *DiscountRuleRepository) GetByPriority(ctx context.Context, orgID string, activeOnly bool) ([]*discounts.DiscountRule, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "org_id", Operator: base.OpEqual, Value: orgID},
	}

	// Filter for active rules only if requested
	if activeOnly {
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "is_active",
			Operator: base.OpEqual,
			Value:    true,
		})
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	// Sort by priority descending (highest priority first)
	filter.Sort = []base.SortField{
		{Field: "priority", Direction: "DESC"},
	}

	var items []*discounts.DiscountRule
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get discount rules by priority: %w", err)
	}

	return items, nil
}

// GetStackableRules retrieves all stackable discount rules
func (r *DiscountRuleRepository) GetStackableRules(ctx context.Context, orgID string, activeOnly bool) ([]*discounts.DiscountRule, error) {
	return r.GetByRuleType(ctx, orgID, "stackable", activeOnly)
}

// GetExclusiveRules retrieves all exclusive discount rules
func (r *DiscountRuleRepository) GetExclusiveRules(ctx context.Context, orgID string, activeOnly bool) ([]*discounts.DiscountRule, error) {
	return r.GetByRuleType(ctx, orgID, "exclusive", activeOnly)
}

// GetConditionalRules retrieves all conditional discount rules
func (r *DiscountRuleRepository) GetConditionalRules(ctx context.Context, orgID string, activeOnly bool) ([]*discounts.DiscountRule, error) {
	return r.GetByRuleType(ctx, orgID, "conditional", activeOnly)
}

// GetByName retrieves a discount rule by name
func (r *DiscountRuleRepository) GetByName(ctx context.Context, orgID, name string) (*discounts.DiscountRule, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{Field: "org_id", Operator: base.OpEqual, Value: orgID},
		{Field: "name", Operator: base.OpEqual, Value: name},
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var items []*discounts.DiscountRule
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to get discount rule by name: %w", err)
	}

	if len(items) == 0 {
		return nil, nil
	}

	return items[0], nil
}

// UpdatePriority updates the priority of a discount rule
func (r *DiscountRuleRepository) UpdatePriority(ctx context.Context, id string, priority int) error {
	rule, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get discount rule for priority update: %w", err)
	}

	rule.Priority = priority
	return r.Update(ctx, rule)
}

// DeactivateRule deactivates a discount rule
func (r *DiscountRuleRepository) DeactivateRule(ctx context.Context, id string) error {
	rule, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get discount rule for deactivation: %w", err)
	}

	rule.IsActive = false
	return r.Update(ctx, rule)
}

// ActivateRule activates a discount rule
func (r *DiscountRuleRepository) ActivateRule(ctx context.Context, id string) error {
	rule, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get discount rule for activation: %w", err)
	}

	rule.IsActive = true
	return r.Update(ctx, rule)
}

// GetApplicableRules retrieves rules applicable to a specific context
// This is a helper method that would be enhanced in the service layer with actual condition checking
func (r *DiscountRuleRepository) GetApplicableRules(ctx context.Context, orgID string) ([]*discounts.DiscountRule, error) {
	// Get all active rules sorted by priority
	rules, err := r.GetByPriority(ctx, orgID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get applicable rules: %w", err)
	}

	// In a real implementation, we would check conditions here
	// For now, return all active rules
	return rules, nil
}

// BulkUpdatePriority updates priorities for multiple rules in a single transaction
func (r *DiscountRuleRepository) BulkUpdatePriority(ctx context.Context, updates map[string]int) error {
	for id, priority := range updates {
		if err := r.UpdatePriority(ctx, id, priority); err != nil {
			return fmt.Errorf("failed to update priority for rule %s: %w", id, err)
		}
	}
	return nil
}
