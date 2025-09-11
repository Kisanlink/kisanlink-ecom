package discounts

import (
    "time"

    "github.com/Kisanlink/kisanlink-db/pkg/base"
    "github.com/shopspring/decimal"
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
    base.BaseModel
    OrgID       string `json:"org_id" gorm:"type:varchar(255);not null;index"` // Organization that offers the discount
    Code        string `json:"code" gorm:"type:varchar(50);unique;not null"`   // Discount code (e.g., "SAVE20")
    Name        string `json:"name" gorm:"type:varchar(255);not null"`         // Discount name
    Description string `json:"description" gorm:"type:text"`                   // Discount description

    // Discount details
    DiscountType DiscountType    `json:"discount_type" gorm:"type:varchar(20);not null;index"`
    Value        decimal.Decimal `json:"value" gorm:"type:decimal(12,2);not null"` // Discount value (percentage or fixed amount)
    Currency     string          `json:"currency" gorm:"type:varchar(3);not null;default:'INR'"`

    // Applicability
    EntityType  string `json:"entity_type" gorm:"type:varchar(50);index"`  // "product", "service", "labour", "all"
    Category    string `json:"category" gorm:"type:varchar(100);index"`    // Product category this applies to
    Subcategory string `json:"subcategory" gorm:"type:varchar(100);index"` // Product subcategory
    EntityIDs   string `json:"entity_ids" gorm:"type:jsonb"`               // Specific entity IDs this applies to

    // Conditions
    MinOrderAmount *decimal.Decimal `json:"min_order_amount" gorm:"type:decimal(12,2)"` // Minimum order amount required
    MaxOrderAmount *decimal.Decimal `json:"max_order_amount" gorm:"type:decimal(12,2)"` // Maximum order amount for discount
    MinQuantity    *decimal.Decimal `json:"min_quantity" gorm:"type:decimal(10,3)"`     // Minimum quantity required
    MaxQuantity    *decimal.Decimal `json:"max_quantity" gorm:"type:decimal(10,3)"`     // Maximum quantity for discount

    // Usage limits
    MaxUsage        *int `json:"max_usage" gorm:"type:int"`          // Maximum times this discount can be used
    MaxUsagePerUser *int `json:"max_usage_per_user" gorm:"type:int"` // Maximum times per user

    // Validity period
    ValidFrom *time.Time `json:"valid_from" gorm:"type:timestamp"`
    ValidTo   *time.Time `json:"valid_to" gorm:"type:timestamp"`

    // Status
    IsActive bool `json:"is_active" gorm:"default:true"`

    // Metadata
    Metadata string `json:"metadata" gorm:"type:jsonb"` // Additional discount metadata
}

// TableName returns the table name for GORM
func (Discount) TableName() string {
    return "discounts"
}

// NewDiscount creates a new Discount instance
func NewDiscount(orgID, code, name string, discountType DiscountType, value decimal.Decimal) *Discount {
    return &Discount{
        BaseModel:    *base.NewBaseModel("DISC", "medium"),
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
func (d *Discount) IsValidForOrder(orderAmount, quantity decimal.Decimal) bool {
    if !d.IsValid() {
        return false
    }

    if d.MinOrderAmount != nil && orderAmount.LessThan(*d.MinOrderAmount) {
        return false
    }

    if d.MaxOrderAmount != nil && orderAmount.GreaterThan(*d.MaxOrderAmount) {
        return false
    }

    if d.MinQuantity != nil && quantity.LessThan(*d.MinQuantity) {
        return false
    }

    if d.MaxQuantity != nil && quantity.GreaterThan(*d.MaxQuantity) {
        return false
    }

    return true
}

// CalculateDiscount calculates the discount amount for a given order amount
func (d *Discount) CalculateDiscount(orderAmount decimal.Decimal) decimal.Decimal {
    if !d.IsValid() {
        return decimal.Zero
    }

    switch d.DiscountType {
    case DiscountTypePercentage:
        return orderAmount.Mul(d.Value).Div(decimal.NewFromInt(100))
    case DiscountTypeFixed:
        if d.Value.GreaterThan(orderAmount) {
            return orderAmount // Can't discount more than the order amount
        }
        return d.Value
    default:
        return decimal.Zero
    }
}

// DiscountRule represents complex discount rules
type DiscountRule struct {
    base.BaseModel
    OrgID       string `json:"org_id" gorm:"type:varchar(255);not null;index"`
    Name        string `json:"name" gorm:"type:varchar(255);not null"`
    Description string `json:"description" gorm:"type:text"`
    RuleType    string `json:"rule_type" gorm:"type:varchar(50);not null"` // "stackable", "exclusive", "conditional"
    RuleData    string `json:"rule_data" gorm:"type:jsonb"`                // Rule-specific data
    Priority    int    `json:"priority" gorm:"default:0"`                  // Higher priority rules are applied first
    IsActive    bool   `json:"is_active" gorm:"default:true"`

    // Conditions
    Conditions string `json:"conditions" gorm:"type:jsonb"` // When this rule applies
}

// TableName returns the table name for GORM
func (DiscountRule) TableName() string {
    return "discount_rules"
}

// NewDiscountRule creates a new DiscountRule instance
func NewDiscountRule(orgID, name, ruleType string) *DiscountRule {
    return &DiscountRule{
        BaseModel: *base.NewBaseModel("DISCRULE", "medium"),
        OrgID:     orgID,
        Name:      name,
        RuleType:  ruleType,
        IsActive:  true,
        Priority:  0,
    }
}

// DiscountUsage represents usage tracking for discounts
type DiscountUsage struct {
    base.BaseModel
    DiscountID string          `json:"discount_id" gorm:"type:varchar(255);not null;index"` // References discounts table
    UserID     string          `json:"user_id" gorm:"type:varchar(255);not null;index"`     // User who used the discount
    OrderID    string          `json:"order_id" gorm:"type:varchar(255);not null;index"`    // Order where discount was applied
    Amount     decimal.Decimal `json:"amount" gorm:"type:decimal(12,2);not null"`           // Discount amount applied
    Currency   string          `json:"currency" gorm:"type:varchar(3);not null;default:'INR'"`
    UsedAt     time.Time       `json:"used_at" gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name for GORM
func (DiscountUsage) TableName() string {
    return "discount_usage"
}

// NewDiscountUsage creates a new DiscountUsage instance
func NewDiscountUsage(discountID, userID, orderID string, amount decimal.Decimal, currency string) *DiscountUsage {
    return &DiscountUsage{
        BaseModel:  *base.NewBaseModel("DISCUSE", "medium"),
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
    OrderAmount       decimal.Decimal            `json:"order_amount"`       // Original order amount
    DiscountAmount    decimal.Decimal            `json:"discount_amount"`    // Total discount amount
    FinalAmount       decimal.Decimal            `json:"final_amount"`       // Order amount after discount
    AppliedDiscounts  []string                   `json:"applied_discounts"`  // Applied discount codes
    DiscountBreakdown map[string]decimal.Decimal `json:"discount_breakdown"` // Breakdown by discount code
    AppliedRules      []string                   `json:"applied_rules"`      // Applied discount rules
}

// NewDiscountCalculation creates a new DiscountCalculation instance
func NewDiscountCalculation(orderAmount decimal.Decimal) *DiscountCalculation {
    return &DiscountCalculation{
        OrderAmount:       orderAmount,
        DiscountAmount:    decimal.Zero,
        FinalAmount:       orderAmount,
        AppliedDiscounts:  make([]string, 0),
        DiscountBreakdown: make(map[string]decimal.Decimal),
        AppliedRules:      make([]string, 0),
    }
}

// AddDiscount adds a discount amount for a specific discount code
func (dc *DiscountCalculation) AddDiscount(discountCode string, amount decimal.Decimal) {
    dc.DiscountBreakdown[discountCode] = amount
    dc.DiscountAmount = dc.DiscountAmount.Add(amount)
    dc.FinalAmount = dc.OrderAmount.Sub(dc.DiscountAmount)

    // Ensure final amount doesn't go below zero
    if dc.FinalAmount.LessThan(decimal.Zero) {
        dc.FinalAmount = decimal.Zero
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
