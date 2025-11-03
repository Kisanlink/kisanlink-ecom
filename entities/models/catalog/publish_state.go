package catalog

import (
	"errors"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Publish state validation errors
var (
	ErrInvalidProductID     = errors.New("product ID is required")
	ErrInvalidPlatformFee   = errors.New("platform fee must be between 0 and 100")
	ErrNegativeDeliveryCost = errors.New("delivery cost cannot be negative")
)

// PublishState tracks the publishing status and FPO-specific configuration for products
type PublishState struct {
	base.BaseModel

	// Product reference
	ProductID string `json:"product_id" gorm:"type:varchar(255);not null;uniqueIndex:idx_publish_states_product_id"`

	// FPO Access Control - stored as JSONB array of FPO organization IDs
	FPOAccessList []string `json:"fpo_access_list" gorm:"type:jsonb;not null;default:'[]';index:idx_publish_states_fpo_access"`

	// Delivery Costs - JSONB map of FPO ID -> delivery cost
	// Example: {"fpo_org_123": 50.00, "fpo_org_456": 75.00}
	DeliveryCosts map[string]decimal.Decimal `json:"delivery_costs" gorm:"type:jsonb;not null;default:'{}'"`

	// Platform commission as percentage (0-100)
	PlatformFeePercent decimal.Decimal `json:"platform_fee_percent" gorm:"type:decimal(5,2);not null;default:10.00;check:platform_fee_percent >= 0 AND platform_fee_percent <= 100"`

	// Publishing metadata
	PublishedAt *time.Time `json:"published_at" gorm:"type:timestamp;index:idx_publish_states_published_at"`
	PublishedBy string     `json:"published_by" gorm:"type:varchar(255)"`

	// Audit fields
	CreatedBy string `json:"created_by" gorm:"type:varchar(255);not null"`
	UpdatedBy string `json:"updated_by" gorm:"type:varchar(255);not null"`

	// Soft delete support
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index:idx_publish_states_deleted"`
}

// FPOPricingDetail represents the calculated pricing for a specific FPO
type FPOPricingDetail struct {
	BasePrice        decimal.Decimal `json:"base_price"`
	DeliveryCost     decimal.Decimal `json:"delivery_cost"`
	CommissionAmount decimal.Decimal `json:"commission_amount"`
	RetailPrice      decimal.Decimal `json:"retail_price"`
	PriceLockedUntil *time.Time      `json:"price_locked_until,omitempty"`
}

// ProductWithFPOPricing represents a product enriched with FPO-specific pricing
type ProductWithFPOPricing struct {
	Product

	// FPO-specific pricing details
	Pricing FPOPricingDetail `json:"pricing"`
}

// TableName returns the table name for GORM
func (PublishState) TableName() string {
	return "publish_states"
}

// NewPublishState creates a new publish state with validation
func NewPublishState(productID string, publishedBy string) *PublishState {
	now := time.Now()
	return &PublishState{
		BaseModel:          *base.NewBaseModel("PUB", "medium"),
		ProductID:          productID,
		FPOAccessList:      []string{},
		DeliveryCosts:      make(map[string]decimal.Decimal),
		PlatformFeePercent: decimal.NewFromFloat(10.00), // Default 10%
		PublishedAt:        &now,
		PublishedBy:        publishedBy,
		CreatedBy:          publishedBy,
		UpdatedBy:          publishedBy,
	}
}

// HasFPOAccess checks if a specific FPO has access to the product
func (ps *PublishState) HasFPOAccess(fpoOrgID string) bool {
	for _, id := range ps.FPOAccessList {
		if id == fpoOrgID {
			return true
		}
	}
	return false
}

// AddFPOAccess adds an FPO to the access list if not already present
func (ps *PublishState) AddFPOAccess(fpoOrgID string) {
	if !ps.HasFPOAccess(fpoOrgID) {
		ps.FPOAccessList = append(ps.FPOAccessList, fpoOrgID)
	}
}

// RemoveFPOAccess removes an FPO from the access list
func (ps *PublishState) RemoveFPOAccess(fpoOrgID string) {
	for i, id := range ps.FPOAccessList {
		if id == fpoOrgID {
			ps.FPOAccessList = append(ps.FPOAccessList[:i], ps.FPOAccessList[i+1:]...)
			return
		}
	}
}

// SetDeliveryCost sets the delivery cost for a specific FPO
func (ps *PublishState) SetDeliveryCost(fpoOrgID string, cost decimal.Decimal) {
	if ps.DeliveryCosts == nil {
		ps.DeliveryCosts = make(map[string]decimal.Decimal)
	}
	ps.DeliveryCosts[fpoOrgID] = cost
}

// GetDeliveryCost retrieves the delivery cost for a specific FPO
func (ps *PublishState) GetDeliveryCost(fpoOrgID string) decimal.Decimal {
	if cost, exists := ps.DeliveryCosts[fpoOrgID]; exists {
		return cost
	}
	return decimal.Zero
}

// CalculateRetailPrice calculates the FPO retail price given base price and delivery cost
func (ps *PublishState) CalculateRetailPrice(basePrice, deliveryCost decimal.Decimal) decimal.Decimal {
	// Commission = BasePrice * (PlatformFeePercent / 100)
	commission := basePrice.Mul(ps.PlatformFeePercent).Div(decimal.NewFromInt(100))

	// RetailPrice = BasePrice + DeliveryCost + Commission
	return basePrice.Add(deliveryCost).Add(commission)
}

// GetPricingForFPO calculates complete pricing details for a specific FPO
func (ps *PublishState) GetPricingForFPO(basePrice decimal.Decimal, fpoOrgID string, priceLockMinutes int) FPOPricingDetail {
	deliveryCost := ps.GetDeliveryCost(fpoOrgID)
	commission := basePrice.Mul(ps.PlatformFeePercent).Div(decimal.NewFromInt(100))
	retailPrice := ps.CalculateRetailPrice(basePrice, deliveryCost)

	var priceLock *time.Time
	if priceLockMinutes > 0 {
		lockTime := time.Now().Add(time.Duration(priceLockMinutes) * time.Minute)
		priceLock = &lockTime
	}

	return FPOPricingDetail{
		BasePrice:        basePrice,
		DeliveryCost:     deliveryCost,
		CommissionAmount: commission,
		RetailPrice:      retailPrice,
		PriceLockedUntil: priceLock,
	}
}

// IsPublished checks if the product is currently published
func (ps *PublishState) IsPublished() bool {
	return ps.PublishedAt != nil && ps.DeletedAt.Time.IsZero()
}

// Validate validates the publish state data
func (ps *PublishState) Validate() error {
	if ps.ProductID == "" {
		return ErrInvalidProductID
	}

	if ps.PlatformFeePercent.LessThan(decimal.Zero) || ps.PlatformFeePercent.GreaterThan(decimal.NewFromInt(100)) {
		return ErrInvalidPlatformFee
	}

	// Validate all delivery costs are non-negative
	for _, cost := range ps.DeliveryCosts {
		if cost.LessThan(decimal.Zero) {
			return ErrNegativeDeliveryCost
		}
	}

	return nil
}
