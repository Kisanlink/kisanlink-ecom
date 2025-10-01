package catalog

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// CatalogItemType represents the type of catalog item
type CatalogItemType string

const (
	CatalogItemTypeProduct  CatalogItemType = "PRODUCT"
	CatalogItemTypeService  CatalogItemType = "SERVICE"
	CatalogItemTypeLabour   CatalogItemType = "LABOUR"
	CatalogItemTypeContract CatalogItemType = "CONTRACT"
)

// VisibilityType represents the visibility level of catalog items
type VisibilityType string

const (
	VisibilityPrivate VisibilityType = "PRIVATE"
	VisibilityOrg     VisibilityType = "ORG"
	VisibilityNetwork VisibilityType = "NETWORK"
	VisibilityPublic  VisibilityType = "PUBLIC"
)

// CatalogItem represents the base catalog entity with tenant isolation
type CatalogItem struct {
	base.BaseModel

	// Tenant isolation (using OrganizationID as TenantID for compatibility)
	OrganizationID string `json:"organization_id" gorm:"type:varchar(255);not null;index:idx_tenant;index:idx_tenant_type_status,priority:1;index:idx_tenant_category,priority:1;index:idx_tenant_vendor,priority:1"`

	// Item Classification
	ItemType    CatalogItemType `json:"item_type" gorm:"type:varchar(20);not null;index:idx_tenant_type_status,priority:2"`
	CategoryID  string          `json:"category_id" gorm:"type:varchar(255);index:idx_tenant_category,priority:2"`
	Category    string          `json:"category" gorm:"type:varchar(100)"`
	Subcategory string          `json:"subcategory" gorm:"type:varchar(100)"`

	// Basic Information
	Name          string `json:"name" gorm:"type:varchar(255);not null;index:idx_catalog_items_search_name"`
	Description   string `json:"description" gorm:"type:text;index:idx_search_desc"`
	SKU           string `json:"sku" gorm:"type:varchar(100);uniqueIndex:idx_org_sku,where:sku IS NOT NULL AND sku != ''"`
	UnitOfMeasure string `json:"unit_of_measure" gorm:"type:varchar(50)"`

	// Vendor/FPO Information
	VendorID string `json:"vendor_id" gorm:"type:varchar(255);index:idx_tenant_vendor,priority:3"`

	// Pricing (Authoritative)
	BasePrice decimal.Decimal `json:"base_price" gorm:"type:decimal(12,2);not null;check:base_price >= 0"`
	Currency  string          `json:"currency" gorm:"type:varchar(3);not null;default:'INR'"`

	// Availability & Visibility (PublishState)
	IsActive   bool           `json:"is_active" gorm:"not null;default:true;index:idx_tenant_type_status,priority:3"`
	Visibility VisibilityType `json:"visibility" gorm:"type:varchar(20);not null;default:'PRIVATE'"`

	// Metadata
	Tags       pq.StringArray `json:"tags" gorm:"type:text[];index:idx_tags"`
	Attributes string         `json:"attributes" gorm:"type:jsonb;index:idx_attributes"`
	Images     pq.StringArray `json:"images" gorm:"type:text[]"`

	// Audit fields
	CreatedBy string `json:"created_by" gorm:"type:varchar(255);not null"`
	UpdatedBy string `json:"updated_by" gorm:"type:varchar(255);not null"`
	Version   int64  `json:"version" gorm:"not null;default:1"`

	// Soft delete support
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index:idx_catalog_items_deleted"`

	// Note: Relationships are managed via separate queries to avoid circular imports:
	// - VendorID references actors.Vendor
	// - CategoryID references catalog.Category
	// - Media via media.Media (CatalogItemID)
	// - Variants via catalog.Variant (CatalogItemID)
	// - Availability via catalog.Availability (CatalogItemID)
	// - Pricing via pricing.Price (EntityID)
}

// VersionedEntity interface implementation for optimistic locking
func (c *CatalogItem) GetID() string {
	return c.ID
}

func (c *CatalogItem) GetVersion() int64 {
	return c.Version
}

func (c *CatalogItem) GetUpdatedAt() time.Time {
	return c.UpdatedAt
}

func (c *CatalogItem) IncrementVersion() {
	c.Version++
}

// Product represents a product catalog item
type Product struct {
	CatalogItem

	// Product-specific fields
	Weight        *decimal.Decimal `json:"weight" gorm:"type:decimal(10,3);check:weight IS NULL OR weight >= 0"`
	Dimensions    string           `json:"dimensions" gorm:"type:jsonb"` // {length, width, height}
	Perishable    bool             `json:"perishable" gorm:"not null;default:false"`
	ShelfLifeDays *int             `json:"shelf_life_days" gorm:"check:shelf_life_days IS NULL OR shelf_life_days > 0"`
}

// Service represents a service catalog item
type Service struct {
	CatalogItem

	// Service-specific fields
	DurationMinutes *int   `json:"duration_minutes" gorm:"check:duration_minutes IS NULL OR duration_minutes > 0"`
	ServiceArea     string `json:"service_area" gorm:"type:jsonb"` // geographic coverage
}

// Labour represents a labour catalog item
type Labour struct {
	CatalogItem

	// Labour-specific fields
	SkillLevel string           `json:"skill_level" gorm:"type:varchar(50)"`
	HourlyRate *decimal.Decimal `json:"hourly_rate" gorm:"type:decimal(10,2);check:hourly_rate IS NULL OR hourly_rate >= 0"`
	Skills     pq.StringArray   `json:"skills" gorm:"type:text[];default:'{}'"`
	Experience int              `json:"experience" gorm:"not null;default:0;check:experience >= 0"`
	RateType   string           `json:"rate_type" gorm:"type:varchar(20);not null;default:'hourly';check:rate_type IN ('hourly', 'daily', 'weekly', 'monthly')"`
}

// Contract represents a contract catalog item
type Contract struct {
	CatalogItem

	// Contract-specific fields
	Term      string     `json:"term" gorm:"type:varchar(100);default:'TBD'"`
	Duration  int        `json:"duration" gorm:"default:1;check:duration > 0"`
	StartDate *time.Time `json:"start_date" gorm:"type:timestamp"`
	EndDate   *time.Time `json:"end_date" gorm:"type:timestamp"`
	Terms     string     `json:"terms" gorm:"type:text"`
}

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

// TableName returns the table name for GORM
func (Contract) TableName() string {
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
		RateType:    "hourly",
	}
}

// NewContract creates a new contract catalog item
func NewContract(orgID string, name string, basePrice decimal.Decimal, term string, duration int) *Contract {
	return &Contract{
		CatalogItem: *NewCatalogItem(orgID, CatalogItemTypeContract, name, basePrice),
		Term:        term,
		Duration:    duration,
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

// IsContract checks if the item is a contract
func (c *CatalogItem) IsContract() bool {
	return c.ItemType == CatalogItemTypeContract
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
