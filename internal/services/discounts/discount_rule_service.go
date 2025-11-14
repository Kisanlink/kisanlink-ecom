// Package discounts provides service layer for discount rule management
package discounts

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/discounts"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// DiscountRuleRepositoryInterface defines repository operations
type DiscountRuleRepositoryInterface interface {
	Create(ctx context.Context, rule *discounts.DiscountRule) error
	GetByID(ctx context.Context, id string) (*discounts.DiscountRule, error)
	GetByOrganizationID(ctx context.Context, orgID string) ([]*discounts.DiscountRule, error)
	Update(ctx context.Context, rule *discounts.DiscountRule) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter *base.Filter, limit, offset int) ([]*discounts.DiscountRule, int, error)
}

// DiscountRuleServiceInterface defines service operations
type DiscountRuleServiceInterface interface {
	CreateDiscountRule(ctx context.Context, orgID, name, ruleType string, priority int) (*discounts.DiscountRule, error)
	GetDiscountRule(ctx context.Context, id string) (*discounts.DiscountRule, error)
	GetOrganizationRules(ctx context.Context, orgID string) ([]*discounts.DiscountRule, error)
	UpdateDiscountRule(ctx context.Context, rule *discounts.DiscountRule) (*discounts.DiscountRule, error)
	UpdateRuleData(ctx context.Context, id string, ruleData map[string]interface{}) (*discounts.DiscountRule, error)
	UpdateConditions(ctx context.Context, id string, conditions map[string]interface{}) (*discounts.DiscountRule, error)
	UpdatePriority(ctx context.Context, id string, priority int) (*discounts.DiscountRule, error)
	DeleteDiscountRule(ctx context.Context, id string) error
	ListDiscountRules(ctx context.Context, limit, offset int) ([]*discounts.DiscountRule, int, error)
	GetActiveRules(ctx context.Context, orgID string) ([]*discounts.DiscountRule, error)
	GetRulesByPriority(ctx context.Context, orgID string) ([]*discounts.DiscountRule, error)
	CheckConflict(ctx context.Context, ruleID string, otherRuleID string) (bool, error)
	IsApplicable(ctx context.Context, ruleID string, conditionsData map[string]interface{}) (bool, error)
}

// DiscountRuleService provides business logic for discount rule management
type DiscountRuleService struct {
	repo DiscountRuleRepositoryInterface
}

// NewDiscountRuleService creates a new discount rule service
func NewDiscountRuleService(repo DiscountRuleRepositoryInterface) *DiscountRuleService {
	return &DiscountRuleService{repo: repo}
}

// CreateDiscountRule creates a new discount rule
func (s *DiscountRuleService) CreateDiscountRule(ctx context.Context, orgID, name, ruleType string, priority int) (*discounts.DiscountRule, error) {
	// Validate input
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if ruleType == "" {
		return nil, fmt.Errorf("rule type is required")
	}

	// Validate rule type
	validTypes := map[string]bool{
		"stackable":   true,
		"exclusive":   true,
		"conditional": true,
	}
	if !validTypes[ruleType] {
		return nil, fmt.Errorf("invalid rule type: %s, must be one of: stackable, exclusive, conditional", ruleType)
	}

	// Validate priority
	if priority < 0 {
		return nil, fmt.Errorf("priority cannot be negative")
	}

	// Create new discount rule
	rule := discounts.NewDiscountRule(orgID, name, ruleType)
	rule.Priority = priority

	if err := s.repo.Create(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to create discount rule: %w", err)
	}

	return rule, nil
}

// GetDiscountRule retrieves a discount rule by ID
func (s *DiscountRuleService) GetDiscountRule(ctx context.Context, id string) (*discounts.DiscountRule, error) {
	if id == "" {
		return nil, fmt.Errorf("discount rule ID is required")
	}

	rule, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get discount rule: %w", err)
	}
	if rule == nil {
		return nil, fmt.Errorf("discount rule not found")
	}

	return rule, nil
}

// GetOrganizationRules retrieves all discount rules for an organization
func (s *DiscountRuleService) GetOrganizationRules(ctx context.Context, orgID string) ([]*discounts.DiscountRule, error) {
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	rules, err := s.repo.GetByOrganizationID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization rules: %w", err)
	}

	return rules, nil
}

// GetActiveRules retrieves all active discount rules for an organization
func (s *DiscountRuleService) GetActiveRules(ctx context.Context, orgID string) ([]*discounts.DiscountRule, error) {
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	allRules, err := s.repo.GetByOrganizationID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization rules: %w", err)
	}

	// Filter to only active rules
	activeRules := make([]*discounts.DiscountRule, 0)
	for _, rule := range allRules {
		if rule.IsActive {
			activeRules = append(activeRules, rule)
		}
	}

	return activeRules, nil
}

// GetRulesByPriority retrieves all active rules for an organization sorted by priority (highest first)
func (s *DiscountRuleService) GetRulesByPriority(ctx context.Context, orgID string) ([]*discounts.DiscountRule, error) {
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	activeRules, err := s.GetActiveRules(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Sort by priority (highest first) - using bubble sort for simplicity
	n := len(activeRules)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if activeRules[j].Priority < activeRules[j+1].Priority {
				activeRules[j], activeRules[j+1] = activeRules[j+1], activeRules[j]
			}
		}
	}

	return activeRules, nil
}

// UpdateDiscountRule updates an existing discount rule
func (s *DiscountRuleService) UpdateDiscountRule(ctx context.Context, rule *discounts.DiscountRule) (*discounts.DiscountRule, error) {
	if rule == nil {
		return nil, fmt.Errorf("discount rule cannot be nil")
	}

	// Validate exists
	existing, err := s.repo.GetByID(ctx, rule.ID)
	if err != nil || existing == nil {
		return nil, fmt.Errorf("discount rule not found")
	}

	// Validate priority
	if rule.Priority < 0 {
		return nil, fmt.Errorf("priority cannot be negative")
	}

	if err := s.repo.Update(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to update discount rule: %w", err)
	}

	return rule, nil
}

// UpdateRuleData updates the rule data for a discount rule
func (s *DiscountRuleService) UpdateRuleData(ctx context.Context, id string, ruleData map[string]interface{}) (*discounts.DiscountRule, error) {
	if id == "" {
		return nil, fmt.Errorf("discount rule ID is required")
	}
	if ruleData == nil {
		return nil, fmt.Errorf("rule data cannot be nil")
	}

	rule, err := s.repo.GetByID(ctx, id)
	if err != nil || rule == nil {
		return nil, fmt.Errorf("discount rule not found")
	}

	// Marshal rule data to JSON
	ruleJSON, err := json.Marshal(ruleData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal rule data: %w", err)
	}

	rule.RuleData = string(ruleJSON)

	if err := s.repo.Update(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to update rule data: %w", err)
	}

	return rule, nil
}

// UpdateConditions updates the conditions for a discount rule
func (s *DiscountRuleService) UpdateConditions(ctx context.Context, id string, conditions map[string]interface{}) (*discounts.DiscountRule, error) {
	if id == "" {
		return nil, fmt.Errorf("discount rule ID is required")
	}
	if conditions == nil {
		return nil, fmt.Errorf("conditions cannot be nil")
	}

	rule, err := s.repo.GetByID(ctx, id)
	if err != nil || rule == nil {
		return nil, fmt.Errorf("discount rule not found")
	}

	// Marshal conditions to JSON
	condJSON, err := json.Marshal(conditions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal conditions: %w", err)
	}

	rule.Conditions = string(condJSON)

	if err := s.repo.Update(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to update conditions: %w", err)
	}

	return rule, nil
}

// UpdatePriority updates the priority of a discount rule
func (s *DiscountRuleService) UpdatePriority(ctx context.Context, id string, priority int) (*discounts.DiscountRule, error) {
	if id == "" {
		return nil, fmt.Errorf("discount rule ID is required")
	}
	if priority < 0 {
		return nil, fmt.Errorf("priority cannot be negative")
	}

	rule, err := s.repo.GetByID(ctx, id)
	if err != nil || rule == nil {
		return nil, fmt.Errorf("discount rule not found")
	}

	rule.Priority = priority

	if err := s.repo.Update(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to update priority: %w", err)
	}

	return rule, nil
}

// DeleteDiscountRule soft deletes a discount rule
func (s *DiscountRuleService) DeleteDiscountRule(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("discount rule ID is required")
	}

	// Check exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("discount rule not found")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete discount rule: %w", err)
	}

	return nil
}

// ListDiscountRules retrieves discount rules with pagination
func (s *DiscountRuleService) ListDiscountRules(ctx context.Context, limit, offset int) ([]*discounts.DiscountRule, int, error) {
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	rules, total, err := s.repo.List(ctx, nil, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list discount rules: %w", err)
	}

	return rules, total, nil
}

// CheckConflict checks if two discount rules conflict with each other
func (s *DiscountRuleService) CheckConflict(ctx context.Context, ruleID string, otherRuleID string) (bool, error) {
	if ruleID == "" || otherRuleID == "" {
		return false, fmt.Errorf("both rule IDs are required")
	}

	rule1, err := s.repo.GetByID(ctx, ruleID)
	if err != nil || rule1 == nil {
		return false, fmt.Errorf("first discount rule not found")
	}

	rule2, err := s.repo.GetByID(ctx, otherRuleID)
	if err != nil || rule2 == nil {
		return false, fmt.Errorf("second discount rule not found")
	}

	// Exclusive rules conflict with all other rules
	if rule1.RuleType == "exclusive" || rule2.RuleType == "exclusive" {
		return true, nil
	}

	// Stackable rules don't conflict with each other
	if rule1.RuleType == "stackable" && rule2.RuleType == "stackable" {
		return false, nil
	}

	// Conditional rules may conflict - would need to check conditions
	// For now, assume they don't conflict
	return false, nil
}

// IsApplicable checks if a discount rule is applicable given certain conditions
func (s *DiscountRuleService) IsApplicable(ctx context.Context, ruleID string, conditionsData map[string]interface{}) (bool, error) {
	if ruleID == "" {
		return false, fmt.Errorf("discount rule ID is required")
	}

	rule, err := s.repo.GetByID(ctx, ruleID)
	if err != nil || rule == nil {
		return false, fmt.Errorf("discount rule not found")
	}

	// Check if rule is active
	if !rule.IsActive {
		return false, nil
	}

	// If no conditions are set, rule is always applicable
	if rule.Conditions == "" || rule.Conditions == "{}" {
		return true, nil
	}

	// Parse rule conditions
	var ruleConditions map[string]interface{}
	if err := json.Unmarshal([]byte(rule.Conditions), &ruleConditions); err != nil {
		return false, fmt.Errorf("failed to parse rule conditions: %w", err)
	}

	// Check if all rule conditions are met
	// This is a simplified implementation - actual logic would be more complex
	for key, expectedValue := range ruleConditions {
		if actualValue, ok := conditionsData[key]; ok {
			if actualValue != expectedValue {
				return false, nil
			}
		} else {
			// Required condition is missing
			return false, nil
		}
	}

	return true, nil
}
