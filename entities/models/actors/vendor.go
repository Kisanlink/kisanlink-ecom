package actors

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// VendorType represents the type of vendor
type VendorType string

const (
	VendorTypeIndividual  VendorType = "individual"
	VendorTypeFPO         VendorType = "fpo" // Farmer Producer Organization
	VendorTypeCorporate   VendorType = "corporate"
	VendorTypeCooperative VendorType = "cooperative"
)

// VendorStatus represents the status of vendor
type VendorStatus string

const (
	VendorStatusActive    VendorStatus = "active"
	VendorStatusInactive  VendorStatus = "inactive"
	VendorStatusSuspended VendorStatus = "suspended"
	VendorStatusPending   VendorStatus = "pending"
)

// Vendor represents vendor/FPO information
type Vendor struct {
	base.BaseModel

	// Tenant isolation
	OrganizationID string `json:"organization_id" gorm:"type:varchar(255);not null;index:idx_tenant;index:idx_org_status,priority:1"`

	// Vendor details
	Name        string       `json:"name" gorm:"type:varchar(255);not null;index:idx_vendors_search_name"`
	Type        VendorType   `json:"type" gorm:"type:varchar(20);not null;check:type IN ('individual', 'fpo', 'corporate', 'cooperative')"`
	Status      VendorStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending';index:idx_org_status,priority:2;check:status IN ('active', 'inactive', 'suspended', 'pending')"`
	Description string       `json:"description" gorm:"type:text"`

	// Contact information
	ContactPerson string `json:"contact_person" gorm:"type:varchar(255)"`
	Email         string `json:"email" gorm:"type:varchar(255);index:idx_email"`
	Phone         string `json:"phone" gorm:"type:varchar(20);index:idx_phone"`

	// Address
	Address    string `json:"address" gorm:"type:text"`
	City       string `json:"city" gorm:"type:varchar(100);index:idx_city"`
	State      string `json:"state" gorm:"type:varchar(100);index:idx_state"`
	Country    string `json:"country" gorm:"type:varchar(100);not null;default:'India'"`
	PostalCode string `json:"postal_code" gorm:"type:varchar(20)"`

	// Business information
	BusinessLicense string         `json:"business_license" gorm:"type:varchar(100);uniqueIndex:idx_business_license,where:business_license IS NOT NULL AND business_license != ''"`
	TaxID           string         `json:"tax_id" gorm:"type:varchar(50);uniqueIndex:idx_tax_id,where:tax_id IS NOT NULL AND tax_id != ''"`
	Categories      pq.StringArray `json:"categories" gorm:"type:text[];index:idx_categories"`

	// Rating and verification
	Rating     *float64 `json:"rating" gorm:"type:decimal(3,2);check:rating IS NULL OR (rating >= 0 AND rating <= 5)"`
	IsVerified bool     `json:"is_verified" gorm:"not null;default:false;index:idx_verified"`
	VerifiedAt *string  `json:"verified_at" gorm:"type:timestamp"`

	// Metadata
	Metadata string `json:"metadata" gorm:"type:jsonb;index:idx_vendors_metadata"`

	// Audit fields
	CreatedBy string `json:"created_by" gorm:"type:varchar(255);not null"`
	UpdatedBy string `json:"updated_by" gorm:"type:varchar(255);not null"`

	// Soft delete support
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index:idx_vendors_deleted"`
}

// TableName returns the table name for GORM
func (Vendor) TableName() string {
	return "vendors"
}

// NewVendor creates a new vendor
func NewVendor(orgID, name string, vendorType VendorType) *Vendor {
	return &Vendor{
		BaseModel:      *base.NewBaseModel("VEN", "large"),
		OrganizationID: orgID,
		Name:           name,
		Type:           vendorType,
		Status:         VendorStatusPending,
		Country:        "India",
		IsVerified:     false,
	}
}

// IsActive checks if the vendor is active
func (v *Vendor) IsActive() bool {
	return v.Status == VendorStatusActive
}

// IsFPO checks if the vendor is an FPO
func (v *Vendor) IsFPO() bool {
	return v.Type == VendorTypeFPO
}

// GetFullAddress returns the complete address
func (v *Vendor) GetFullAddress() string {
	address := v.Address
	if v.City != "" {
		address += ", " + v.City
	}
	if v.State != "" {
		address += ", " + v.State
	}
	if v.PostalCode != "" {
		address += " - " + v.PostalCode
	}
	if v.Country != "" {
		address += ", " + v.Country
	}
	return address
}
