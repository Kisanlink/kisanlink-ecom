package pricing

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// PriceType represents the type of price
type PriceType string

const (
	PriceTypeBase      PriceType = "base"      // Base price
	PriceTypeSale      PriceType = "sale"      // Sale price
	PriceTypeWholesale PriceType = "wholesale" // Wholesale price
	PriceTypeBulk      PriceType = "bulk"      // Bulk purchase price
	PriceTypeSeasonal  PriceType = "seasonal"  // Seasonal pricing
	PriceTypeDynamic   PriceType = "dynamic"   // Dynamic pricing
)

// Currency represents supported currencies
type Currency string

const (
	CurrencyINR Currency = "INR" // Indian Rupee
	CurrencyUSD Currency = "USD" // US Dollar
	CurrencyEUR Currency = "EUR" // Euro
)

// Price represents a price for a product or service
type Price struct {
	base.BaseModel
	EntityID   string `json:"entity_id" gorm:"type:varchar(255);not null;index:idx_prices_entity;index:idx_prices_org_entity,priority:2"`                                  // References catalog item, service, etc.
	EntityType string `json:"entity_type" gorm:"type:varchar(50);not null;index:idx_prices_entity_type;check:entity_type IN ('product', 'service', 'labour', 'contract')"` // "product", "service", "labour"
	OrgID      string `json:"org_id" gorm:"type:varchar(255);not null;index:idx_prices_org;index:idx_prices_org_entity,priority:1"`                                        // Organization that sets the price

	PriceType PriceType       `json:"price_type" gorm:"type:varchar(20);not null;default:'base';check:price_type IN ('base', 'sale', 'wholesale', 'bulk', 'seasonal', 'dynamic')"`
	Amount    decimal.Decimal `json:"amount" gorm:"type:decimal(12,2);not null;check:amount >= 0"`
	Currency  Currency        `json:"currency" gorm:"type:varchar(3);not null;default:'INR';check:currency IN ('INR', 'USD', 'EUR')"`

	// Pricing rules
	MinQuantity *decimal.Decimal `json:"min_quantity" gorm:"type:decimal(10,3);check:min_quantity IS NULL OR min_quantity >= 0"` // Minimum quantity for this price
	MaxQuantity *decimal.Decimal `json:"max_quantity" gorm:"type:decimal(10,3);check:max_quantity IS NULL OR max_quantity > 0"`  // Maximum quantity for this price

	// Validity period
	ValidFrom *time.Time `json:"valid_from" gorm:"type:timestamp"`
	ValidTo   *time.Time `json:"valid_to" gorm:"type:timestamp"`

	// Status
	IsActive bool `json:"is_active" gorm:"not null;default:true;index:idx_prices_active"`

	// Metadata
	Description string `json:"description" gorm:"type:text"`
	Metadata    string `json:"metadata" gorm:"type:jsonb;index:idx_prices_metadata"` // Additional pricing metadata

	// Audit fields
	CreatedBy string `json:"created_by" gorm:"type:varchar(255);not null"`
	UpdatedBy string `json:"updated_by" gorm:"type:varchar(255);not null"`

	// Soft delete support
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index:idx_prices_deleted"`
}

// TableName returns the table name for GORM
func (Price) TableName() string {
	return "prices"
}

// NewPrice creates a new Price instance
func NewPrice(entityID, entityType, orgID string, priceType PriceType, amount decimal.Decimal, currency Currency) *Price {
	return &Price{
		BaseModel:  *base.NewBaseModel("PRICE", "medium"),
		EntityID:   entityID,
		EntityType: entityType,
		OrgID:      orgID,
		PriceType:  priceType,
		Amount:     amount,
		Currency:   currency,
		IsActive:   true,
	}
}

// IsValid checks if the price is currently valid
func (p *Price) IsValid() bool {
	now := time.Now()

	if !p.IsActive {
		return false
	}

	if p.ValidFrom != nil && now.Before(*p.ValidFrom) {
		return false
	}

	if p.ValidTo != nil && now.After(*p.ValidTo) {
		return false
	}

	return true
}

// IsValidForQuantity checks if the price is valid for the given quantity
func (p *Price) IsValidForQuantity(quantity decimal.Decimal) bool {
	if p.MinQuantity != nil && quantity.LessThan(*p.MinQuantity) {
		return false
	}

	if p.MaxQuantity != nil && quantity.GreaterThan(*p.MaxQuantity) {
		return false
	}

	return true
}

// PriceTier represents quantity-based pricing tiers
type PriceTier struct {
	base.BaseModel
	PriceID   string           `json:"price_id" gorm:"type:varchar(255);not null;index:idx_price_tiers_price"` // References prices table
	MinQty    decimal.Decimal  `json:"min_qty" gorm:"type:decimal(10,3);not null;check:min_qty >= 0"`
	MaxQty    *decimal.Decimal `json:"max_qty" gorm:"type:decimal(10,3);check:max_qty IS NULL OR max_qty > min_qty"` // NULL means unlimited
	UnitPrice decimal.Decimal  `json:"unit_price" gorm:"type:decimal(12,2);not null;check:unit_price >= 0"`
	Currency  Currency         `json:"currency" gorm:"type:varchar(3);not null;default:'INR';check:currency IN ('INR', 'USD', 'EUR')"`

	// Audit fields
	CreatedBy string `json:"created_by" gorm:"type:varchar(255);not null"`
	UpdatedBy string `json:"updated_by" gorm:"type:varchar(255);not null"`

	// Soft delete support
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index:idx_price_tiers_deleted"`
}

// TableName returns the table name for GORM
func (PriceTier) TableName() string {
	return "price_tiers"
}

// NewPriceTier creates a new PriceTier instance
func NewPriceTier(priceID string, minQty, unitPrice decimal.Decimal, currency Currency) *PriceTier {
	return &PriceTier{
		BaseModel: *base.NewBaseModel("TIER", "small"),
		PriceID:   priceID,
		MinQty:    minQty,
		UnitPrice: unitPrice,
		Currency:  currency,
	}
}

// PriceRule represents complex pricing rules
type PriceRule struct {
	base.BaseModel
	OrgID       string `json:"org_id" gorm:"type:varchar(255);not null;index:idx_price_rules_org"`
	Name        string `json:"name" gorm:"type:varchar(255);not null"`
	Description string `json:"description" gorm:"type:text"`
	RuleType    string `json:"rule_type" gorm:"type:varchar(50);not null;check:rule_type IN ('percentage', 'fixed', 'formula')"` // "percentage", "fixed", "formula"
	RuleData    string `json:"rule_data" gorm:"type:jsonb;index:idx_price_rules_rule_data"`                                      // Rule-specific data
	Priority    int    `json:"priority" gorm:"not null;default:0;index:idx_price_rules_priority"`                                // Higher priority rules are applied first
	IsActive    bool   `json:"is_active" gorm:"not null;default:true;index:idx_price_rules_active"`

	// Conditions
	Conditions string `json:"conditions" gorm:"type:jsonb;index:idx_price_rules_conditions"` // When this rule applies

	// Audit fields
	CreatedBy string `json:"created_by" gorm:"type:varchar(255);not null"`
	UpdatedBy string `json:"updated_by" gorm:"type:varchar(255);not null"`

	// Soft delete support
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index:idx_price_rules_deleted"`
}

// TableName returns the table name for GORM
func (PriceRule) TableName() string {
	return "price_rules"
}

// NewPriceRule creates a new PriceRule instance
func NewPriceRule(orgID, name, ruleType string) *PriceRule {
	return &PriceRule{
		BaseModel: *base.NewBaseModel("RULE", "medium"),
		OrgID:     orgID,
		Name:      name,
		RuleType:  ruleType,
		IsActive:  true,
		Priority:  0,
	}
}
