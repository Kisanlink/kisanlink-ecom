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

// TaxType represents the type of tax applied
type TaxType string

const (
	TaxTypeGST    TaxType = "gst"    // Goods and Services Tax
	TaxTypeCGST   TaxType = "cgst"   // Central GST
	TaxTypeSGST   TaxType = "sgst"   // State GST
	TaxTypeIGST   TaxType = "igst"   // Integrated GST
	TaxTypeVAT    TaxType = "vat"    // Value Added Tax
	TaxTypeExcise TaxType = "excise" // Excise Duty
	TaxTypeCustom TaxType = "custom" // Custom Duty
)

// Tax represents a tax component for a catalog item
type Tax struct {
	Type        TaxType `json:"type" gorm:"type:varchar(20);not null"`
	Percentage  float64 `json:"percentage" gorm:"type:decimal(5,2);not null"`
	IsActive    bool    `json:"is_active" gorm:"default:true"`
	Description string  `json:"description" gorm:"type:varchar(255)"`
}

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
	Tags        []string        `json:"tags" gorm:"type:jsonb"`
	Images      []string        `json:"images" gorm:"type:jsonb"`

	// Taxation fields
	Taxes        []Tax `json:"taxes" gorm:"type:jsonb"`
	TaxInclusive bool  `json:"tax_inclusive" gorm:"default:false"` // Whether price includes taxes
}

// NewCatalogItem creates a new CatalogItem instance
func NewCatalogItem(itemType CatalogItemType, orgID, sku, name, description, category, subcategory, uom string, price float64, currency string) *CatalogItem {
	return &CatalogItem{
		BaseModel:    base.NewBaseModel("CAT", hash.Medium),
		Type:         itemType,
		OrgID:        orgID,
		SKU:          sku,
		Name:         name,
		Description:  description,
		Category:     category,
		Subcategory:  subcategory,
		UOM:          uom,
		Price:        price,
		Currency:     currency,
		IsActive:     true,
		Tags:         []string{},
		Images:       []string{},
		Taxes:        []Tax{},
		TaxInclusive: false,
	}
}

// CalculateTaxAmount calculates the total tax amount for the item
func (c *CatalogItem) CalculateTaxAmount() float64 {
	var totalTax float64
	for _, tax := range c.Taxes {
		if tax.IsActive {
			totalTax += (c.Price * tax.Percentage) / 100
		}
	}
	return totalTax
}

// GetPriceWithTax returns the price including taxes
func (c *CatalogItem) GetPriceWithTax() float64 {
	if c.TaxInclusive {
		return c.Price
	}
	return c.Price + c.CalculateTaxAmount()
}

// GetPriceWithoutTax returns the price excluding taxes
func (c *CatalogItem) GetPriceWithoutTax() float64 {
	if c.TaxInclusive {
		return c.Price - c.CalculateTaxAmount()
	}
	return c.Price
}

// Product represents a physical commodity, seed, fertilizer, or tool
type Product struct {
	CatalogItem
	StockQty    int `json:"stock_qty" gorm:"type:int;default:0"`
	MinOrderQty int `json:"min_order_qty" gorm:"type:int;default:1"`
	MaxOrderQty int `json:"max_order_qty" gorm:"type:int;default:1000"`
}

// NewProduct creates a new Product instance
func NewProduct(orgID, sku, name, description, category, subcategory, uom string, price float64, currency string, stockQty, minOrderQty, maxOrderQty int) *Product {
	return &Product{
		CatalogItem: *NewCatalogItem(CatalogItemTypeProduct, orgID, sku, name, description, category, subcategory, uom, price, currency),
		StockQty:    stockQty,
		MinOrderQty: minOrderQty,
		MaxOrderQty: maxOrderQty,
	}
}

// ServiceOffering represents a service like drone spraying, tractor, harvester, etc.
type ServiceOffering struct {
	CatalogItem
	Duration string `json:"duration" gorm:"type:varchar(100)"` // e.g., "2 hours", "1 day"
	Location string `json:"location" gorm:"type:varchar(255)"`
}

// NewServiceOffering creates a new ServiceOffering instance
func NewServiceOffering(orgID, sku, name, description, category, subcategory, uom string, price float64, currency string, duration, location string) *ServiceOffering {
	return &ServiceOffering{
		CatalogItem: *NewCatalogItem(CatalogItemTypeService, orgID, sku, name, description, category, subcategory, uom, price, currency),
		Duration:    duration,
		Location:    location,
	}
}

// LabourOffering represents manual labor services like weeding, tilling, sowing, etc.
type LabourOffering struct {
	CatalogItem
	RateType   string   `json:"rate_type" gorm:"type:varchar(50);not null"` // "per_hour", "per_day", "per_task"
	Skills     []string `json:"skills" gorm:"type:jsonb"`
	Experience string   `json:"experience" gorm:"type:varchar(255)"`
	Location   string   `json:"location" gorm:"type:varchar(255)"`
}

// NewLabourOffering creates a new LabourOffering instance
func NewLabourOffering(orgID, sku, name, description, category, subcategory, uom string, price float64, currency string, rateType, experience, location string, skills []string) *LabourOffering {
	return &LabourOffering{
		CatalogItem: *NewCatalogItem(CatalogItemTypeLabour, orgID, sku, name, description, category, subcategory, uom, price, currency),
		RateType:    rateType,
		Skills:      skills,
		Experience:  experience,
		Location:    location,
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
		BaseModel:     base.NewBaseModel("INV", hash.Medium),
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
		BaseModel:     base.NewBaseModel("SLOT", hash.Small),
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
		BaseModel:     base.NewBaseModel("LAB", hash.Small),
		LabourID:      labourID,
		ProviderOrgID: providerOrgID,
		Availability:  availability,
	}
}

// Request and response models for API operations

// CreateCatalogItemRequest represents a request to create a catalog item
type CreateCatalogItemRequest struct {
	Type         CatalogItemType `json:"type" validate:"required"`
	OrgID        string          `json:"org_id" validate:"required"`
	SKU          string          `json:"sku" validate:"required"`
	Name         string          `json:"name" validate:"required"`
	Description  string          `json:"description"`
	Category     string          `json:"category" validate:"required"`
	Subcategory  string          `json:"subcategory"`
	UOM          string          `json:"uom" validate:"required"`
	Price        float64         `json:"price" validate:"required,min=0"`
	Currency     string          `json:"currency" validate:"required,len=3"`
	IsActive     bool            `json:"is_active"`
	Tags         []string        `json:"tags"`
	Images       []string        `json:"images"`
	TaxInclusive bool            `json:"tax_inclusive"`
	Taxes        []Tax           `json:"taxes"`

	// Product-specific fields
	StockQty    *int `json:"stock_qty"`
	MinOrderQty *int `json:"min_order_qty"`
	MaxOrderQty *int `json:"max_order_qty"`

	// Service-specific fields
	Duration *string `json:"duration"`
	Location *string `json:"location"`

	// Labour-specific fields
	RateType   *string   `json:"rate_type"`
	Skills     *[]string `json:"skills"`
	Experience *string   `json:"experience"`
}

// UpdateCatalogItemRequest represents a request to update a catalog item
type UpdateCatalogItemRequest struct {
	Name         *string   `json:"name"`
	Description  *string   `json:"description"`
	Category     *string   `json:"category"`
	Subcategory  *string   `json:"subcategory"`
	UOM          *string   `json:"uom"`
	Price        *float64  `json:"price" validate:"omitempty,min=0"`
	Currency     *string   `json:"currency" validate:"omitempty,len=3"`
	IsActive     *bool     `json:"is_active"`
	Tags         *[]string `json:"tags"`
	Images       *[]string `json:"images"`
	TaxInclusive *bool     `json:"tax_inclusive"`
	Taxes        *[]Tax    `json:"taxes"`

	// Product-specific fields
	StockQty    *int `json:"stock_qty"`
	MinOrderQty *int `json:"min_order_qty"`
	MaxOrderQty *int `json:"max_order_qty"`

	// Service-specific fields
	Duration *string `json:"duration"`
	Location *string `json:"location"`

	// Labour-specific fields
	RateType   *string   `json:"rate_type"`
	Skills     *[]string `json:"skills"`
	Experience *string   `json:"experience"`
}

// CatalogFilter represents filters for catalog queries
type CatalogFilter struct {
	Search      string            `json:"search"`
	Category    string            `json:"category"`
	Subcategory string            `json:"subcategory"`
	OrgID       string            `json:"org_id"`
	Type        CatalogItemType   `json:"type"`
	IsActive    *bool             `json:"is_active"`
	MinPrice    *float64          `json:"min_price"`
	MaxPrice    *float64          `json:"max_price"`
	Currency    string            `json:"currency"`
	Tags        []string          `json:"tags"`
	Extra       map[string]string `json:"extra"`
}
