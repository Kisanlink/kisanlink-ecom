package catalog

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Variant represents product/service variants
type Variant struct {
	base.BaseModel

	// Tenant isolation
	OrganizationID string `json:"organization_id" gorm:"type:varchar(255);not null;index:idx_tenant;index:idx_org_sku,priority:1"`

	// Parent catalog item
	CatalogItemID string `json:"catalog_item_id" gorm:"type:varchar(255);not null;index:idx_catalog_item"`

	// Variant details
	Name        string `json:"name" gorm:"type:varchar(255);not null"`
	SKU         string `json:"sku" gorm:"type:varchar(100);not null;uniqueIndex:idx_org_sku,priority:2"`
	Description string `json:"description" gorm:"type:text"`

	// Pricing
	Price    decimal.Decimal `json:"price" gorm:"type:decimal(12,2);not null;check:price >= 0"`
	Currency string          `json:"currency" gorm:"type:varchar(3);not null;default:'INR'"`

	// Variant attributes (size, color, etc.)
	Attributes string `json:"attributes" gorm:"type:jsonb;index:idx_attributes"`

	// Availability
	IsActive  bool `json:"is_active" gorm:"not null;default:true;index:idx_active"`
	IsDefault bool `json:"is_default" gorm:"not null;default:false"`
	SortOrder int  `json:"sort_order" gorm:"not null;default:0"`

	// Images specific to this variant
	Images pq.StringArray `json:"images" gorm:"type:text[]"`

	// Audit fields
	CreatedBy string `json:"created_by" gorm:"type:varchar(255);not null"`
	UpdatedBy string `json:"updated_by" gorm:"type:varchar(255);not null"`

	// Soft delete support
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index:idx_variants_deleted"`

	// Relationships
	CatalogItem *CatalogItem `json:"catalog_item,omitempty" gorm:"foreignKey:CatalogItemID"`
}

// TableName returns the table name for GORM
func (Variant) TableName() string {
	return "variants"
}

// NewVariant creates a new variant
func NewVariant(orgID, catalogItemID, name, sku string, price decimal.Decimal) *Variant {
	return &Variant{
		BaseModel:      *base.NewBaseModel("VAR", "large"),
		OrganizationID: orgID,
		CatalogItemID:  catalogItemID,
		Name:           name,
		SKU:            sku,
		Price:          price,
		Currency:       "INR",
		IsActive:       true,
		IsDefault:      false,
		SortOrder:      0,
	}
}

// IsAvailable checks if the variant is available
func (v *Variant) IsAvailable() bool {
	return v.IsActive
}
