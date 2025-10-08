package catalog

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// AvailabilityType represents different types of availability
type AvailabilityType string

const (
	AvailabilityInStock      AvailabilityType = "in_stock"
	AvailabilityOutOfStock   AvailabilityType = "out_of_stock"
	AvailabilityPreOrder     AvailabilityType = "pre_order"
	AvailabilityBackOrder    AvailabilityType = "back_order"
	AvailabilityDiscontinued AvailabilityType = "discontinued"
)

// Availability represents availability and inventory data
type Availability struct {
	base.BaseModel

	// Tenant isolation
	OrganizationID string `json:"organization_id" gorm:"type:varchar(255);not null;index:idx_availability_tenant"`

	// Parent catalog item or variant
	CatalogItemID *string `json:"catalog_item_id" gorm:"type:varchar(255);index:idx_availability_catalog_item"`
	VariantID     *string `json:"variant_id" gorm:"type:varchar(255);index:idx_availability_variant"`

	// Availability details
	Type   AvailabilityType `json:"type" gorm:"type:varchar(20);not null;check:type IN ('in_stock', 'out_of_stock', 'pre_order', 'back_order', 'discontinued')"`
	Status string           `json:"status" gorm:"type:varchar(50);not null;default:'available';index:idx_availability_status"`

	// Inventory
	QuantityAvailable *decimal.Decimal `json:"quantity_available" gorm:"type:decimal(12,3);check:quantity_available IS NULL OR quantity_available >= 0"`
	QuantityReserved  *decimal.Decimal `json:"quantity_reserved" gorm:"type:decimal(12,3);not null;default:0;check:quantity_reserved >= 0"`
	ReorderLevel      *decimal.Decimal `json:"reorder_level" gorm:"type:decimal(12,3);check:reorder_level IS NULL OR reorder_level >= 0"`
	MaxOrderQuantity  *decimal.Decimal `json:"max_order_quantity" gorm:"type:decimal(12,3);check:max_order_quantity IS NULL OR max_order_quantity > 0"`

	// Location
	Location  string `json:"location" gorm:"type:varchar(255);index:idx_availability_location"`
	Warehouse string `json:"warehouse" gorm:"type:varchar(255);index:idx_availability_warehouse"`

	// Time-based availability
	AvailableFrom *time.Time `json:"available_from" gorm:"type:timestamp"`
	AvailableTo   *time.Time `json:"available_to" gorm:"type:timestamp"`

	// Lead time for restocking
	LeadTimeDays *int `json:"lead_time_days" gorm:"check:lead_time_days IS NULL OR lead_time_days >= 0"`

	// Audit fields
	CreatedBy string `json:"created_by" gorm:"type:varchar(255);not null"`
	UpdatedBy string `json:"updated_by" gorm:"type:varchar(255);not null"`

	// Soft delete support
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index:idx_availability_deleted"`

	// Relationships
	CatalogItem *CatalogItem `json:"catalog_item,omitempty" gorm:"foreignKey:CatalogItemID"`
	Variant     *Variant     `json:"variant,omitempty" gorm:"foreignKey:VariantID"`
}

// TableName returns the table name for GORM
func (Availability) TableName() string {
	return "availability"
}

// NewAvailability creates a new availability record
func NewAvailability(orgID string, availabilityType AvailabilityType) *Availability {
	return &Availability{
		BaseModel:      *base.NewBaseModel("AVL", "large"),
		OrganizationID: orgID,
		Type:           availabilityType,
		Status:         "available",
	}
}

// IsAvailable checks if the item is currently available
func (a *Availability) IsAvailable() bool {
	if a.Type == AvailabilityOutOfStock || a.Type == AvailabilityDiscontinued {
		return false
	}

	now := time.Now()

	if a.AvailableFrom != nil && now.Before(*a.AvailableFrom) {
		return false
	}

	if a.AvailableTo != nil && now.After(*a.AvailableTo) {
		return false
	}

	return true
}

// GetAvailableQuantity returns the available quantity
func (a *Availability) GetAvailableQuantity() decimal.Decimal {
	if a.QuantityAvailable == nil {
		return decimal.Zero
	}

	available := *a.QuantityAvailable
	if a.QuantityReserved != nil {
		available = available.Sub(*a.QuantityReserved)
	}

	return available
}

// NeedsReorder checks if the item needs to be reordered
func (a *Availability) NeedsReorder() bool {
	if a.ReorderLevel == nil || a.QuantityAvailable == nil {
		return false
	}

	return a.QuantityAvailable.LessThan(*a.ReorderLevel)
}
