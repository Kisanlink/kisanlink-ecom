package catalog

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// CatalogItemType represents the type of catalog item
type CatalogItemType string

const (
	CatalogItemTypeProduct CatalogItemType = "product"
	CatalogItemTypeService CatalogItemType = "service"
	CatalogItemTypeLabour  CatalogItemType = "labour"
)

// CatalogItem represents a generic catalog item
type CatalogItem struct {
	*base.BaseModel
	Type        CatalogItemType `json:"type" gorm:"type:varchar(20);not null"`
	OrgID       string          `json:"org_id" gorm:"type:varchar(255);not null;index"` // References organizations table
	SKU         string          `json:"sku" gorm:"type:varchar(100);uniqueIndex;not null"`
	Name        string          `json:"name" gorm:"type:varchar(255);not null"`
	Description string          `json:"description" gorm:"type:text"`
	Category    string          `json:"category" gorm:"type:varchar(100);not null"`
	Subcategory string          `json:"subcategory" gorm:"type:varchar(100)"`
	UOM         string          `json:"uom" gorm:"type:varchar(50);not null"`
	Price       float64         `json:"price" gorm:"type:decimal(10,2);not null"`
	Currency    string          `json:"currency" gorm:"type:varchar(3);not null;default:'INR'"`
	IsActive    bool            `json:"is_active" gorm:"default:true"`
}

// NewCatalogItem creates a new CatalogItem instance
func NewCatalogItem(itemType CatalogItemType, orgID, sku, name, category, uom string, price float64) *CatalogItem {
	return &CatalogItem{
		BaseModel: base.NewBaseModel("CAT", hash.Large), // Large size for catalog items (high volume)
		Type:      itemType,
		OrgID:     orgID,
		SKU:       sku,
		Name:      name,
		Category:  category,
		UOM:       uom,
		Price:     price,
		Currency:  "INR",
		IsActive:  true,
	}
}

// Product represents a physical commodity, seed, fertilizer, or tool
type Product struct {
	CatalogItem
	// Additional product-specific fields can be added here
}

// NewProduct creates a new Product instance
func NewProduct(orgID, sku, name, category, uom string, price float64) *Product {
	return &Product{
		CatalogItem: *NewCatalogItem(CatalogItemTypeProduct, orgID, sku, name, category, uom, price),
	}
}

// ServiceOffering represents a service like drone spraying, tractor, harvester, etc.
type ServiceOffering struct {
	CatalogItem
	// Additional service-specific fields can be added here
}

// NewServiceOffering creates a new ServiceOffering instance
func NewServiceOffering(orgID, sku, name, category, uom string, price float64) *ServiceOffering {
	return &ServiceOffering{
		CatalogItem: *NewCatalogItem(CatalogItemTypeService, orgID, sku, name, category, uom, price),
	}
}

// LabourOffering represents manual labor services like weeding, tilling, sowing, etc.
type LabourOffering struct {
	CatalogItem
	// Additional labour-specific fields can be added here
}

// NewLabourOffering creates a new LabourOffering instance
func NewLabourOffering(orgID, sku, name, category, uom string, price float64) *LabourOffering {
	return &LabourOffering{
		CatalogItem: *NewCatalogItem(CatalogItemTypeLabour, orgID, sku, name, category, uom, price),
	}
}

// InventoryLot represents inventory for products
type InventoryLot struct {
	*base.BaseModel
	CatalogItemID string     `json:"catalog_item_id" gorm:"type:varchar(255);not null;index"` // References catalog_items table
	Quantity      float64    `json:"quantity" gorm:"type:decimal(10,2);not null"`
	UOM           string     `json:"uom" gorm:"type:varchar(50);not null"`
	LocationID    *string    `json:"location_id" gorm:"type:varchar(255)"` // References locations table if exists
	ExpiresAt     *time.Time `json:"expires_at"`
}

// NewInventoryLot creates a new InventoryLot instance
func NewInventoryLot(catalogItemID, uom string, quantity float64) *InventoryLot {
	return &InventoryLot{
		BaseModel:     base.NewBaseModel("INV", hash.Large), // Large size for inventory (high volume)
		CatalogItemID: catalogItemID,
		Quantity:      quantity,
		UOM:           uom,
	}
}

// ServiceSlot represents available time slots for services
type ServiceSlot struct {
	*base.BaseModel
	ServiceID     string    `json:"service_id" gorm:"type:varchar(255);not null;index"`      // References catalog_items table
	ProviderOrgID string    `json:"provider_org_id" gorm:"type:varchar(255);not null;index"` // References organizations table
	SlotStartsAt  time.Time `json:"slot_starts_at" gorm:"not null"`
	SlotEndsAt    time.Time `json:"slot_ends_at" gorm:"not null"`
	Capacity      int       `json:"capacity" gorm:"not null;default:1"`
}

// NewServiceSlot creates a new ServiceSlot instance
func NewServiceSlot(serviceID, providerOrgID string, slotStartsAt, slotEndsAt time.Time, capacity int) *ServiceSlot {
	return &ServiceSlot{
		BaseModel:     base.NewBaseModel("SLOT", hash.Large), // Large size for service slots (high volume)
		ServiceID:     serviceID,
		ProviderOrgID: providerOrgID,
		SlotStartsAt:  slotStartsAt,
		SlotEndsAt:    slotEndsAt,
		Capacity:      capacity,
	}
}

// LabourPool represents available labour for labour offerings
type LabourPool struct {
	*base.BaseModel
	LabourID      string `json:"labour_id" gorm:"type:varchar(255);not null;index"`       // References catalog_items table
	ProviderOrgID string `json:"provider_org_id" gorm:"type:varchar(255);not null;index"` // References organizations table
	Availability  string `json:"availability" gorm:"type:jsonb"`                          // JSON string for availability schedule
}

// NewLabourPool creates a new LabourPool instance
func NewLabourPool(labourID, providerOrgID, availability string) *LabourPool {
	return &LabourPool{
		BaseModel:     base.NewBaseModel("LAB", hash.Medium), // Medium size for labour pool
		LabourID:      labourID,
		ProviderOrgID: providerOrgID,
		Availability:  availability,
	}
}
