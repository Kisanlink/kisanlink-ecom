package data

import (
	"github.com/Kisanlink/kisanlink-ecom/entities/models/discounts"
)

// CreateTestDiscountRule creates a test discount rule with default values
func CreateTestDiscountRule() *discounts.DiscountRule {
	return &discounts.DiscountRule{
		OrgID:       "org-test-id-123",
		Name:        "Test Discount Rule",
		Description: "Test discount rule for unit testing",
		RuleType:    "stackable",
		Priority:    5,
		IsActive:    true,
		RuleData:    "{}",
		Conditions:  "{}",
	}
}

// CreateTestDiscountRuleWithID creates a test discount rule with a specific ID
func CreateTestDiscountRuleWithID(id string) *discounts.DiscountRule {
	rule := CreateTestDiscountRule()
	rule.ID = id
	return rule
}

// CreateTestDiscountRuleWithOrgID creates a test discount rule with a specific organization ID
func CreateTestDiscountRuleWithOrgID(orgID string) *discounts.DiscountRule {
	rule := CreateTestDiscountRule()
	rule.OrgID = orgID
	return rule
}

// CreateTestDiscountRuleWithName creates a test discount rule with a specific name
func CreateTestDiscountRuleWithName(orgID, name string) *discounts.DiscountRule {
	rule := CreateTestDiscountRule()
	rule.OrgID = orgID
	rule.Name = name
	return rule
}

// CreateTestStackableDiscountRule creates a stackable discount rule
func CreateTestStackableDiscountRule() *discounts.DiscountRule {
	rule := CreateTestDiscountRule()
	rule.RuleType = "stackable"
	return rule
}

// CreateTestExclusiveDiscountRule creates an exclusive discount rule
func CreateTestExclusiveDiscountRule() *discounts.DiscountRule {
	rule := CreateTestDiscountRule()
	rule.RuleType = "exclusive"
	rule.Priority = 10
	return rule
}

// CreateTestConditionalDiscountRule creates a conditional discount rule
func CreateTestConditionalDiscountRule() *discounts.DiscountRule {
	rule := CreateTestDiscountRule()
	rule.RuleType = "conditional"
	rule.Conditions = `{"min_amount": 1000, "min_quantity": 5}`
	return rule
}

// CreateTestInactiveDiscountRule creates an inactive discount rule
func CreateTestInactiveDiscountRule() *discounts.DiscountRule {
	rule := CreateTestDiscountRule()
	rule.IsActive = false
	return rule
}

// CreateTestDiscountRuleWithPriority creates a discount rule with a specific priority
func CreateTestDiscountRuleWithPriority(priority int) *discounts.DiscountRule {
	rule := CreateTestDiscountRule()
	rule.Priority = priority
	return rule
}

// CreateTestDiscountRuleWithType creates a discount rule with a specific type
func CreateTestDiscountRuleWithType(ruleType string) *discounts.DiscountRule {
	rule := CreateTestDiscountRule()
	rule.RuleType = ruleType
	return rule
}

// CreateTestDiscountRulesArray creates a slice of test discount rules
func CreateTestDiscountRulesArray(count int, orgID string) []*discounts.DiscountRule {
	rules := make([]*discounts.DiscountRule, count)
	for i := 0; i < count; i++ {
		rules[i] = CreateTestDiscountRuleWithOrgID(orgID)
		rules[i].Name = generateDiscountRuleName(i)
		rules[i].Priority = i + 1
	}
	return rules
}

// CreateTestDiscountRulesArrayByType creates discount rules of a specific type
func CreateTestDiscountRulesArrayByType(count int, orgID, ruleType string) []*discounts.DiscountRule {
	rules := make([]*discounts.DiscountRule, count)
	for i := 0; i < count; i++ {
		rules[i] = CreateTestDiscountRuleWithOrgID(orgID)
		rules[i].Name = generateDiscountRuleName(i)
		rules[i].RuleType = ruleType
		rules[i].Priority = i + 1
	}
	return rules
}

// Helper functions
func generateDiscountRuleName(index int) string {
	return "Test Discount Rule " + string(rune('A'+index))
}
