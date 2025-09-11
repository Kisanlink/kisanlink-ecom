package catalog

import (
    "fmt"
    "time"

    "github.com/Kisanlink/kisanlink-db/pkg/base"
    "github.com/shopspring/decimal"
)

// InventoryStatus represents the status of an inventory lot
type InventoryStatus string

const (
    InventoryStatusAvailable InventoryStatus = "available"
    InventoryStatusReserved  InventoryStatus = "reserved"
    InventoryStatusSold      InventoryStatus = "sold"
    InventoryStatusExpired   InventoryStatus = "expired"
    InventoryStatusDamaged   InventoryStatus = "damaged"
)

// InventoryLot represents product inventory authority
type InventoryLot struct {
    base.BaseModel

    // Product Reference
    CatalogItemID  string `json:"catalog_item_id" gorm:"type:varchar(255);not null;index"`
    OrganizationID string `json:"organization_id" gorm:"type:varchar(255);not null;index"`

    // Lot Information
    LotNumber   string `json:"lot_number" gorm:"type:varchar(100);not null"`
    BatchNumber string `json:"batch_number" gorm:"type:varchar(100)"`

    // Quantity Management
    InitialQuantity   decimal.Decimal `json:"initial_quantity" gorm:"type:decimal(12,3);not null"`
    AvailableQuantity decimal.Decimal `json:"available_quantity" gorm:"type:decimal(12,3);not null"`
    ReservedQuantity  decimal.Decimal `json:"reserved_quantity" gorm:"type:decimal(12,3);default:0"`
    SoldQuantity      decimal.Decimal `json:"sold_quantity" gorm:"type:decimal(12,3);default:0"`

    // Quality & Compliance
    QualityGrade string     `json:"quality_grade" gorm:"type:varchar(50)"`
    HarvestDate  *time.Time `json:"harvest_date" gorm:"type:date"`
    ExpiryDate   *time.Time `json:"expiry_date" gorm:"type:date"`

    // Location
    WarehouseLocation string `json:"warehouse_location" gorm:"type:varchar(255)"`
    StorageConditions string `json:"storage_conditions" gorm:"type:jsonb"`

    // Pricing Override
    LotPrice *decimal.Decimal `json:"lot_price" gorm:"type:decimal(12,2)"` // Override catalog base price

    // Status
    Status InventoryStatus `json:"status" gorm:"type:varchar(20);default:'available'"`

    // Metadata
    Metadata string `json:"metadata" gorm:"type:jsonb"`

    // Relationships
    CatalogItem *CatalogItem `json:"catalog_item,omitempty" gorm:"foreignKey:CatalogItemID"`
}

// TableName returns the table name for GORM
func (InventoryLot) TableName() string {
    return "inventory_lots"
}

// NewInventoryLot creates a new inventory lot
func NewInventoryLot(catalogItemID, orgID, lotNumber string, initialQuantity decimal.Decimal) *InventoryLot {
    return &InventoryLot{
        BaseModel:         *base.NewBaseModel("LOT", "large"),
        CatalogItemID:     catalogItemID,
        OrganizationID:    orgID,
        LotNumber:         lotNumber,
        InitialQuantity:   initialQuantity,
        AvailableQuantity: initialQuantity,
        ReservedQuantity:  decimal.Zero,
        SoldQuantity:      decimal.Zero,
        Status:            InventoryStatusAvailable,
    }
}

// CanReserve checks if the lot can reserve the requested quantity
func (i *InventoryLot) CanReserve(quantity decimal.Decimal) bool {
    return i.Status == InventoryStatusAvailable && i.AvailableQuantity.GreaterThanOrEqual(quantity)
}

// Reserve reserves the specified quantity
func (i *InventoryLot) Reserve(quantity decimal.Decimal) error {
    if !i.CanReserve(quantity) {
        return fmt.Errorf("insufficient available quantity: requested %s, available %s",
            quantity.String(), i.AvailableQuantity.String())
    }

    i.AvailableQuantity = i.AvailableQuantity.Sub(quantity)
    i.ReservedQuantity = i.ReservedQuantity.Add(quantity)

    if i.AvailableQuantity.IsZero() {
        i.Status = InventoryStatusReserved
    }

    return nil
}

// Release releases the specified reserved quantity back to available
func (i *InventoryLot) Release(quantity decimal.Decimal) error {
    if i.ReservedQuantity.LessThan(quantity) {
        return fmt.Errorf("insufficient reserved quantity: requested %s, reserved %s",
            quantity.String(), i.ReservedQuantity.String())
    }

    i.ReservedQuantity = i.ReservedQuantity.Sub(quantity)
    i.AvailableQuantity = i.AvailableQuantity.Add(quantity)

    if i.Status == InventoryStatusReserved && i.AvailableQuantity.GreaterThan(decimal.Zero) {
        i.Status = InventoryStatusAvailable
    }

    return nil
}

// Sell converts reserved quantity to sold
func (i *InventoryLot) Sell(quantity decimal.Decimal) error {
    if i.ReservedQuantity.LessThan(quantity) {
        return fmt.Errorf("insufficient reserved quantity: requested %s, reserved %s",
            quantity.String(), i.ReservedQuantity.String())
    }

    i.ReservedQuantity = i.ReservedQuantity.Sub(quantity)
    i.SoldQuantity = i.SoldQuantity.Add(quantity)

    // Check if lot is completely sold
    totalRemaining := i.AvailableQuantity.Add(i.ReservedQuantity)
    if totalRemaining.IsZero() {
        i.Status = InventoryStatusSold
    } else if i.AvailableQuantity.GreaterThan(decimal.Zero) {
        i.Status = InventoryStatusAvailable
    }

    return nil
}

// IsExpired checks if the lot has expired
func (i *InventoryLot) IsExpired() bool {
    if i.ExpiryDate == nil {
        return false
    }
    return time.Now().After(*i.ExpiryDate)
}

// GetEffectivePrice returns the effective price (lot price or catalog base price)
func (i *InventoryLot) GetEffectivePrice() decimal.Decimal {
    if i.LotPrice != nil {
        return *i.LotPrice
    }
    if i.CatalogItem != nil {
        return i.CatalogItem.BasePrice
    }
    return decimal.Zero
}
