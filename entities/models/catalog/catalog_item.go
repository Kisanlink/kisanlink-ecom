package catalog

import (
    "github.com/Kisanlink/kisanlink-db/pkg/base"
    "github.com/lib/pq"
    "github.com/shopspring/decimal"
)

// CatalogItemType represents the type of catalog item
type CatalogItemType string

const (
    CatalogItemTypeProduct CatalogItemType = "PRODUCT"
    CatalogItemTypeService CatalogItemType = "SERVICE"
    CatalogItemTypeLabour  CatalogItemType = "LABOUR"
    // CatalogItemTypeContract CatalogItemType = "CONTRACT"
)

// VisibilityType represents the visibility level of catalog items
type VisibilityType string

const (
    VisibilityPrivate VisibilityType = "PRIVATE"
    VisibilityOrg     VisibilityType = "ORG"
    VisibilityNetwork VisibilityType = "NETWORK"
    VisibilityPublic  VisibilityType = "PUBLIC"
)

// CatalogItem represents the base catalog entity
type CatalogItem struct {
    base.BaseModel

    // Organization Scoping
    OrganizationID string `json:"organization_id" gorm:"type:varchar(255);not null;index"`

    // Item Classification
    ItemType    CatalogItemType `json:"item_type" gorm:"type:varchar(20);not null;index"`
    Category    string          `json:"category" gorm:"type:varchar(100)"`
    Subcategory string          `json:"subcategory" gorm:"type:varchar(100)"`

    // Basic Information
    Name          string `json:"name" gorm:"type:varchar(255);not null"`
    Description   string `json:"description" gorm:"type:text"`
    SKU           string `json:"sku" gorm:"type:varchar(100)"`
    UnitOfMeasure string `json:"unit_of_measure" gorm:"type:varchar(50)"`

    // Pricing (Authoritative)
    BasePrice decimal.Decimal `json:"base_price" gorm:"type:decimal(12,2);not null"`
    Currency  string          `json:"currency" gorm:"type:varchar(3);default:'INR'"`

    // Availability & Visibility
    IsActive   bool           `json:"is_active" gorm:"default:true"`
    Visibility VisibilityType `json:"visibility" gorm:"type:varchar(20);default:'PRIVATE'"`

    // Metadata
    Tags       pq.StringArray `json:"tags" gorm:"type:text[]"`
    Attributes string         `json:"attributes" gorm:"type:jsonb"`
    Images     pq.StringArray `json:"images" gorm:"type:text[]"`
}

// Product represents a product catalog item
type Product struct {
    CatalogItem

    // Product-specific fields
    Weight        *decimal.Decimal `json:"weight" gorm:"type:decimal(10,3)"`
    Dimensions    string           `json:"dimensions" gorm:"type:jsonb"` // {length, width, height}
    Perishable    bool             `json:"perishable" gorm:"default:false"`
    ShelfLifeDays *int             `json:"shelf_life_days"`
}

// Service represents a service catalog item
type Service struct {
    CatalogItem

    // Service-specific fields
    DurationMinutes *int   `json:"duration_minutes"`
    ServiceArea     string `json:"service_area" gorm:"type:jsonb"` // geographic coverage
}

// Labour represents a labour catalog item
type Labour struct {
    CatalogItem

    // Labour-specific fields
    SkillLevel string           `json:"skill_level" gorm:"type:varchar(50)"`
    HourlyRate *decimal.Decimal `json:"hourly_rate" gorm:"type:decimal(10,2)"`
}

// CONTRACT

// TableName returns the table name for GORM
func (CatalogItem) TableName() string {
    return "catalog_items"
}

// TableName returns the table name for GORM
func (Product) TableName() string {
    return "catalog_items"
}

// TableName returns the table name for GORM
func (Service) TableName() string {
    return "catalog_items"
}

// TableName returns the table name for GORM
func (Labour) TableName() string {
    return "catalog_items"
}

// NewCatalogItem creates a new catalog item
func NewCatalogItem(orgID string, itemType CatalogItemType, name string, basePrice decimal.Decimal) *CatalogItem {
    return &CatalogItem{
        BaseModel:      *base.NewBaseModel("CAT", "large"),
        OrganizationID: orgID,
        ItemType:       itemType,
        Name:           name,
        BasePrice:      basePrice,
        Currency:       "INR",
        IsActive:       true,
        Visibility:     VisibilityPrivate,
    }
}

// NewProduct creates a new product catalog item
func NewProduct(orgID string, name string, basePrice decimal.Decimal) *Product {
    return &Product{
        CatalogItem: *NewCatalogItem(orgID, CatalogItemTypeProduct, name, basePrice),
        Perishable:  false,
    }
}

// NewService creates a new service catalog item
func NewService(orgID string, name string, basePrice decimal.Decimal) *Service {
    return &Service{
        CatalogItem: *NewCatalogItem(orgID, CatalogItemTypeService, name, basePrice),
    }
}

// NewLabour creates a new labour catalog item
func NewLabour(orgID string, name string, basePrice decimal.Decimal) *Labour {
    return &Labour{
        CatalogItem: *NewCatalogItem(orgID, CatalogItemTypeLabour, name, basePrice),
    }
}

// IsProduct checks if the item is a product
func (c *CatalogItem) IsProduct() bool {
    return c.ItemType == CatalogItemTypeProduct
}

// IsService checks if the item is a service
func (c *CatalogItem) IsService() bool {
    return c.ItemType == CatalogItemTypeService
}

// IsLabour checks if the item is labour
func (c *CatalogItem) IsLabour() bool {
    return c.ItemType == CatalogItemTypeLabour
}

// CanBeViewedBy checks if the item can be viewed by the given organization
func (c *CatalogItem) CanBeViewedBy(viewerOrgID string) bool {
    switch c.Visibility {
    case VisibilityPrivate:
        return c.OrganizationID == viewerOrgID
    case VisibilityOrg:
        return c.OrganizationID == viewerOrgID
    case VisibilityNetwork, VisibilityPublic:
        return true
    default:
        return false
    }
}
