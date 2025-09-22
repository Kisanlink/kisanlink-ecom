package catalog

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// LotStatus represents the status of an inventory lot
type LotStatus string

const (
	LotStatusActive   LotStatus = "active"
	LotStatusReserved LotStatus = "reserved"
	LotStatusSold     LotStatus = "sold"
	LotStatusExpired  LotStatus = "expired"
	LotStatusDamaged  LotStatus = "damaged"
	LotStatusReturned LotStatus = "returned"
)

// InventoryLot represents a specific lot/batch of inventory
type InventoryLot struct {
	base.BaseModel

	// Tenant isolation
	OrganizationID string `json:"organization_id" gorm:"type:varchar(255);not null;index:idx_tenant"`

	// Parent catalog item or variant
	CatalogItemID *string `json:"catalog_item_id" gorm:"type:varchar(255);index:idx_catalog_item"`
	VariantID     *string `json:"variant_id" gorm:"type:varchar(255);index:idx_variant"`

	// Lot identification
	LotNumber    string `json:"lot_number" gorm:"type:varchar(100);not null;uniqueIndex:idx_org_lot_number"`
	BatchNumber  string `json:"batch_number" gorm:"type:varchar(100);index:idx_batch"`
	SerialNumber string `json:"serial_number" gorm:"type:varchar(100);uniqueIndex:idx_serial,where:serial_number IS NOT NULL AND serial_number != ''"`

	// Quantity and units
	Quantity      decimal.Decimal `json:"quantity" gorm:"type:decimal(12,3);not null;check:quantity > 0"`
	ReservedQty   decimal.Decimal `json:"reserved_qty" gorm:"type:decimal(12,3);not null;default:0;check:reserved_qty >= 0"`
	AvailableQty  decimal.Decimal `json:"available_qty" gorm:"type:decimal(12,3);not null;check:available_qty >= 0"`
	UnitOfMeasure string          `json:"unit_of_measure" gorm:"type:varchar(50);not null"`

	// Status and condition
	Status    LotStatus `json:"status" gorm:"type:varchar(20);not null;default:'active';index:idx_status;check:status IN ('active', 'reserved', 'sold', 'expired', 'damaged', 'returned')"`
	Condition string    `json:"condition" gorm:"type:varchar(50)"` // new, used, refurbished, damaged

	// Dates
	ManufacturedDate *time.Time `json:"manufactured_date" gorm:"type:timestamp"`
	ExpiryDate       *time.Time `json:"expiry_date" gorm:"type:timestamp;index:idx_expiry"`
	ReceivedDate     *time.Time `json:"received_date" gorm:"type:timestamp"`

	// Location and storage
	Location  string `json:"location" gorm:"type:varchar(255);index:idx_location"`
	Warehouse string `json:"warehouse" gorm:"type:varchar(255);index:idx_warehouse"`
	Zone      string `json:"zone" gorm:"type:varchar(100)"`
	Aisle     string `json:"aisle" gorm:"type:varchar(50)"`
	Shelf     string `json:"shelf" gorm:"type:varchar(50)"`
	Bin       string `json:"bin" gorm:"type:varchar(50)"`

	// Cost and pricing
	UnitCost  decimal.Decimal `json:"unit_cost" gorm:"type:decimal(12,2);check:unit_cost >= 0"`
	TotalCost decimal.Decimal `json:"total_cost" gorm:"type:decimal(12,2);check:total_cost >= 0"`
	Currency  string          `json:"currency" gorm:"type:varchar(3);not null;default:'INR'"`

	// Supplier information
	SupplierID      string `json:"supplier_id" gorm:"type:varchar(255);index:idx_supplier"`
	PurchaseOrderID string `json:"purchase_order_id" gorm:"type:varchar(255);index:idx_po"`

	// Quality and compliance
	QualityGrade    string `json:"quality_grade" gorm:"type:varchar(50)"`     // A, B, C, Premium, Standard
	CertificationID string `json:"certification_id" gorm:"type:varchar(100)"` // Organic, Fair Trade, etc.
	TestResults     string `json:"test_results" gorm:"type:jsonb"`            // Lab test results

	// Tracking and traceability
	OriginLocation string     `json:"origin_location" gorm:"type:varchar(255)"` // Farm, factory location
	HarvestDate    *time.Time `json:"harvest_date" gorm:"type:timestamp"`       // For agricultural products
	ProcessingDate *time.Time `json:"processing_date" gorm:"type:timestamp"`

	// Metadata
	Notes    string `json:"notes" gorm:"type:text"`
	Metadata string `json:"metadata" gorm:"type:jsonb;index:idx_inventory_lots_metadata"`

	// Audit fields
	CreatedBy string `json:"created_by" gorm:"type:varchar(255);not null"`
	UpdatedBy string `json:"updated_by" gorm:"type:varchar(255);not null"`

	// Soft delete support
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index:idx_inventory_lots_deleted"`

	// Relationships
	CatalogItem *CatalogItem `json:"catalog_item,omitempty" gorm:"foreignKey:CatalogItemID"`
	Variant     *Variant     `json:"variant,omitempty" gorm:"foreignKey:VariantID"`
}

// TableName returns the table name for GORM
func (InventoryLot) TableName() string {
	return "inventory_lots"
}

// NewInventoryLot creates a new inventory lot
func NewInventoryLot(orgID, lotNumber string, quantity decimal.Decimal, unitOfMeasure string) *InventoryLot {
	return &InventoryLot{
		BaseModel:      *base.NewBaseModel("LOT", "large"),
		OrganizationID: orgID,
		LotNumber:      lotNumber,
		Quantity:       quantity,
		AvailableQty:   quantity,
		ReservedQty:    decimal.Zero,
		UnitOfMeasure:  unitOfMeasure,
		Status:         LotStatusActive,
		Currency:       "INR",
	}
}

// IsAvailable checks if the lot has available quantity
func (l *InventoryLot) IsAvailable() bool {
	return l.Status == LotStatusActive && l.AvailableQty.GreaterThan(decimal.Zero)
}

// IsExpired checks if the lot has expired
func (l *InventoryLot) IsExpired() bool {
	if l.ExpiryDate == nil {
		return false
	}
	return time.Now().After(*l.ExpiryDate)
}

// ReserveQuantity reserves a specific quantity from the lot
func (l *InventoryLot) ReserveQuantity(qty decimal.Decimal) bool {
	if l.AvailableQty.LessThan(qty) {
		return false
	}

	l.AvailableQty = l.AvailableQty.Sub(qty)
	l.ReservedQty = l.ReservedQty.Add(qty)
	return true
}

// ReleaseReservation releases a reserved quantity back to available
func (l *InventoryLot) ReleaseReservation(qty decimal.Decimal) bool {
	if l.ReservedQty.LessThan(qty) {
		return false
	}

	l.ReservedQty = l.ReservedQty.Sub(qty)
	l.AvailableQty = l.AvailableQty.Add(qty)
	return true
}

// ConsumeQuantity consumes quantity from reserved stock
func (l *InventoryLot) ConsumeQuantity(qty decimal.Decimal) bool {
	if l.ReservedQty.LessThan(qty) {
		return false
	}

	l.ReservedQty = l.ReservedQty.Sub(qty)
	l.Quantity = l.Quantity.Sub(qty)

	// Update status if lot is empty
	if l.Quantity.IsZero() {
		l.Status = LotStatusSold
	}

	return true
}

// GetTotalValue returns the total value of the lot
func (l *InventoryLot) GetTotalValue() decimal.Decimal {
	return l.UnitCost.Mul(l.Quantity)
}

// GetLocationString returns the full location string
func (l *InventoryLot) GetLocationString() string {
	location := l.Warehouse
	if l.Zone != "" {
		location += " - " + l.Zone
	}
	if l.Aisle != "" {
		location += " - " + l.Aisle
	}
	if l.Shelf != "" {
		location += " - " + l.Shelf
	}
	if l.Bin != "" {
		location += " - " + l.Bin
	}
	return location
}
