package discounts

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// DiscountType represents the type of discount
type DiscountType string

const (
	DiscountTypePercentage DiscountType = "percentage" // Percentage discount
	DiscountTypeFixed      DiscountType = "fixed"      // Fixed amount discount
	DiscountTypeBuyOne     DiscountType = "buy_one"    // Buy one get one
	DiscountTypeBulk       DiscountType = "bulk"       // Bulk purchase discount
	DiscountTypeSeasonal   DiscountType = "seasonal"   // Seasonal discount
	DiscountTypeLoyalty    DiscountType = "loyalty"    // Loyalty program discount
	DiscountTypeReferral   DiscountType = "referral"   // Referral discount
	DiscountTypeFirstTime  DiscountType = "first_time" // First-time customer discount
)

// Discount represents a discount configuration
type Discount struct {
	*base.BaseModel
	OrgID       string `json:"org_id" gorm:"type:varchar(255);not null;index"` // Organization that offers the discount
	Code        string `json:"code" gorm:"type:varchar(50);unique;not null"`   // Discount code (e.g., "SAVE20")
	Name        string `json:"name" gorm:"type:varchar(255);not null"`         // Discount name
	Description string `json:"description" gorm:"type:text"`                   // Discount description

	// Discount details
	DiscountType DiscountType `json:"discount_type" gorm:"type:varchar(20);not null;index"`
	Value        float64      `json:"value" gorm:"type:decimal(10,2);not null"` // Discount value (percentage or fixed amount)
	Currency     string       `json:"currency" gorm:"type:varchar(3);not null;default:'INR'"`

	// Applicability
	EntityType  string `json:"entity_type" gorm:"type:varchar(50);index"`  // "product", "service", "labour", "all"
	Category    string `json:"category" gorm:"type:varchar(100);index"`    // Product category this applies to
	Subcategory string `json:"subcategory" gorm:"type:varchar(100);index"` // Product subcategory
	EntityIDs   string `json:"entity_ids" gorm:"type:jsonb"`               // Specific entity IDs this applies to

	// Conditions
	MinOrderAmount *float64 `json:"min_order_amount" gorm:"type:decimal(10,2)"` // Minimum order amount required
	MaxOrderAmount *float64 `json:"max_order_amount" gorm:"type:decimal(10,2)"` // Maximum order amount for discount
	MinQuantity    *float64 `json:"min_quantity" gorm:"type:decimal(10,2)"`     // Minimum quantity required
	MaxQuantity    *float64 `json:"max_quantity" gorm:"type:decimal(10,2)"`     // Maximum quantity for discount

	// Usage limits
	MaxUsage        *int `json:"max_usage" gorm:"type:int"`          // Maximum times this discount can be used
	MaxUsagePerUser *int `json:"max_usage_per_user" gorm:"type:int"` // Maximum times per user

	// Validity period
	ValidFrom *time.Time `json:"valid_from" gorm:"type:timestamp"`
	ValidTo   *time.Time `json:"valid_to" gorm:"type:timestamp"`

	// Status
	IsActive bool `json:"is_active" gorm:"default:true"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata" gorm:"type:jsonb"` // Additional discount metadata
}

// NewDiscount creates a new Discount instance
func NewDiscount(orgID, code, name string, discountType DiscountType, value float64) *Discount {
	return &Discount{
		BaseModel:    base.NewBaseModel("DISC", hash.Medium),
		OrgID:        orgID,
		Code:         code,
		Name:         name,
		DiscountType: discountType,
		Value:        value,
		Currency:     "INR",
		IsActive:     true,
	}
}

// IsValid checks if the discount is currently valid
func (d *Discount) IsValid() bool {
	now := time.Now()

	if !d.IsActive {
		return false
	}

	if d.ValidFrom != nil && now.Before(*d.ValidFrom) {
		return false
	}

	if d.ValidTo != nil && now.After(*d.ValidTo) {
		return false
	}

	return true
}

// IsValidForOrder checks if the discount is valid for the given order
func (d *Discount) IsValidForOrder(orderAmount, quantity float64) bool {
	if !d.IsValid() {
		return false
	}

	if d.MinOrderAmount != nil && orderAmount < *d.MinOrderAmount {
		return false
	}

	if d.MaxOrderAmount != nil && orderAmount > *d.MaxOrderAmount {
		return false
	}

	if d.MinQuantity != nil && quantity < *d.MinQuantity {
		return false
	}

	if d.MaxQuantity != nil && quantity > *d.MaxQuantity {
		return false
	}

	return true
}

// CalculateDiscount calculates the discount amount for a given order amount
func (d *Discount) CalculateDiscount(orderAmount float64) float64 {
	if !d.IsValid() {
		return 0.0
	}

	switch d.DiscountType {
	case DiscountTypePercentage:
		return (orderAmount * d.Value) / 100.0
	case DiscountTypeFixed:
		if d.Value > orderAmount {
			return orderAmount // Can't discount more than the order amount
		}
		return d.Value
	default:
		return 0.0
	}
}

// DiscountRule represents complex discount rules
type DiscountRule struct {
	*base.BaseModel
	OrgID       string                 `json:"org_id" gorm:"type:varchar(255);not null;index"`
	Name        string                 `json:"name" gorm:"type:varchar(255);not null"`
	Description string                 `json:"description" gorm:"type:text"`
	RuleType    string                 `json:"rule_type" gorm:"type:varchar(50);not null"` // "stackable", "exclusive", "conditional"
	RuleData    map[string]interface{} `json:"rule_data" gorm:"type:jsonb"`                // Rule-specific data
	Priority    int                    `json:"priority" gorm:"default:0"`                  // Higher priority rules are applied first
	IsActive    bool                   `json:"is_active" gorm:"default:true"`

	// Conditions
	Conditions map[string]interface{} `json:"conditions" gorm:"type:jsonb"` // When this rule applies
}

// NewDiscountRule creates a new DiscountRule instance
func NewDiscountRule(orgID, name, ruleType string) *DiscountRule {
	return &DiscountRule{
		BaseModel: base.NewBaseModel("DISCRULE", hash.Medium),
		OrgID:     orgID,
		Name:      name,
		RuleType:  ruleType,
		IsActive:  true,
		Priority:  0,
	}
}

// DiscountUsage represents usage tracking for discounts
type DiscountUsage struct {
	*base.BaseModel
	DiscountID string    `json:"discount_id" gorm:"type:varchar(255);not null;index"` // References discounts table
	UserID     string    `json:"user_id" gorm:"type:varchar(255);not null;index"`     // User who used the discount
	OrderID    string    `json:"order_id" gorm:"type:varchar(255);not null;index"`    // Order where discount was applied
	Amount     float64   `json:"amount" gorm:"type:decimal(10,2);not null"`           // Discount amount applied
	Currency   string    `json:"currency" gorm:"type:varchar(3);not null;default:'INR'"`
	UsedAt     time.Time `json:"used_at" gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// NewDiscountUsage creates a new DiscountUsage instance
func NewDiscountUsage(discountID, userID, orderID string, amount float64, currency string) *DiscountUsage {
	return &DiscountUsage{
		BaseModel:  base.NewBaseModel("DISCUSE", hash.Medium),
		DiscountID: discountID,
		UserID:     userID,
		OrderID:    orderID,
		Amount:     amount,
		Currency:   currency,
		UsedAt:     time.Now(),
	}
}

// DiscountCalculation represents a discount calculation result
type DiscountCalculation struct {
	OrderAmount       float64            `json:"order_amount"`       // Original order amount
	DiscountAmount    float64            `json:"discount_amount"`    // Total discount amount
	FinalAmount       float64            `json:"final_amount"`       // Order amount after discount
	AppliedDiscounts  []string           `json:"applied_discounts"`  // Applied discount codes
	DiscountBreakdown map[string]float64 `json:"discount_breakdown"` // Breakdown by discount code
	AppliedRules      []string           `json:"applied_rules"`      // Applied discount rules
}

// NewDiscountCalculation creates a new DiscountCalculation instance
func NewDiscountCalculation(orderAmount float64) *DiscountCalculation {
	return &DiscountCalculation{
		OrderAmount:       orderAmount,
		DiscountAmount:    0.0,
		FinalAmount:       orderAmount,
		AppliedDiscounts:  make([]string, 0),
		DiscountBreakdown: make(map[string]float64),
		AppliedRules:      make([]string, 0),
	}
}

// AddDiscount adds a discount amount for a specific discount code
func (dc *DiscountCalculation) AddDiscount(discountCode string, amount float64) {
	dc.DiscountBreakdown[discountCode] = amount
	dc.DiscountAmount += amount
	dc.FinalAmount = dc.OrderAmount - dc.DiscountAmount

	// Ensure final amount doesn't go below zero
	if dc.FinalAmount < 0 {
		dc.FinalAmount = 0
		dc.DiscountAmount = dc.OrderAmount
	}
}

// AddAppliedDiscount adds an applied discount code
func (dc *DiscountCalculation) AddAppliedDiscount(discountCode string) {
	dc.AppliedDiscounts = append(dc.AppliedDiscounts, discountCode)
}

// AddRule adds an applied discount rule
func (dc *DiscountCalculation) AddRule(ruleID string) {
	dc.AppliedRules = append(dc.AppliedRules, ruleID)
}
