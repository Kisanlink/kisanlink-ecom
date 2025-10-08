package actors

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// AAAEntityType represents the type of AAA entity
type AAAEntityType string

const (
	AAAEntityTypeUser         AAAEntityType = "USER"
	AAAEntityTypeOrganization AAAEntityType = "ORGANIZATION"
)

// Customer represents a customer entity that references AAA service entities
type Customer struct {
	base.BaseModel

	// AAA Service References
	AAAEntityID   string        `json:"aaa_entity_id" gorm:"type:varchar(255);not null;index:idx_customers_aaa_entity;index:idx_customers_aaa_composite,priority:1"`
	AAAEntityType AAAEntityType `json:"aaa_entity_type" gorm:"type:varchar(20);not null;index:idx_customers_aaa_type;index:idx_customers_aaa_composite,priority:2;check:aaa_entity_type IN ('USER', 'ORGANIZATION')"`

	// Internal References (optional, for caching/performance)
	UserID         *string `json:"user_id" gorm:"type:varchar(255);index:idx_customers_user"`        // Reference to user ID from AAA
	OrganizationID *string `json:"organization_id" gorm:"type:varchar(255);index:idx_customers_org"` // Reference to org ID from AAA

	// Customer-specific fields
	CustomerCode string `json:"customer_code" gorm:"type:varchar(100);not null;uniqueIndex:idx_customers_customer_code"`
	DisplayName  string `json:"display_name" gorm:"type:varchar(255);index:idx_customers_search_name"`
	Email        string `json:"email" gorm:"type:varchar(255);index:idx_customers_email"`
	Phone        string `json:"phone" gorm:"type:varchar(20);index:idx_customers_phone"`
	Status       string `json:"status" gorm:"type:varchar(20);not null;default:'active';index:idx_customers_status;check:status IN ('active', 'inactive', 'suspended')"`
	IsVerified   bool   `json:"is_verified" gorm:"not null;default:false;index:idx_customers_verified"`

	// Business Information
	BusinessType      string `json:"business_type" gorm:"type:varchar(50);check:business_type IN ('individual', 'business', 'cooperative', 'fpo') OR business_type IS NULL"` // individual, business, cooperative, etc.
	TaxID             string `json:"tax_id" gorm:"type:varchar(50);uniqueIndex:idx_customers_tax_id,where:tax_id IS NOT NULL AND tax_id != ''"`
	RegistrationID    string `json:"registration_id" gorm:"type:varchar(100);uniqueIndex:idx_customers_registration_id,where:registration_id IS NOT NULL AND registration_id != ''"`
	PreferredLanguage string `json:"preferred_language" gorm:"type:varchar(10);not null;default:'en'"`

	// Address Information (stored as JSON for flexibility)
	BillingAddress  string `json:"billing_address" gorm:"type:jsonb"`
	ShippingAddress string `json:"shipping_address" gorm:"type:jsonb"`

	// Metadata
	Metadata string `json:"metadata" gorm:"type:jsonb;index:idx_customers_metadata"`

	// Audit fields
	CreatedBy string `json:"created_by" gorm:"type:varchar(255);not null"`
	UpdatedBy string `json:"updated_by" gorm:"type:varchar(255);not null"`

	// Soft delete support
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index:idx_customers_deleted"`
}

// TableName returns the table name for GORM
func (Customer) TableName() string {
	return "customers"
}

// NewCustomer creates a new Customer instance
func NewCustomer(aaaEntityID string, aaaEntityType AAAEntityType, customerCode string) *Customer {
	return &Customer{
		BaseModel:     *base.NewBaseModel("CUST", "large"),
		AAAEntityID:   aaaEntityID,
		AAAEntityType: aaaEntityType,
		CustomerCode:  customerCode,
		Status:        "active",
		IsVerified:    false,
	}
}

// NewUserCustomer creates a new Customer for a user entity
func NewUserCustomer(aaaEntityID, userID, customerCode string) *Customer {
	customer := NewCustomer(aaaEntityID, AAAEntityTypeUser, customerCode)
	customer.UserID = &userID
	return customer
}

// NewOrganizationCustomer creates a new Customer for an organization entity
func NewOrganizationCustomer(aaaEntityID, organizationID, customerCode string) *Customer {
	customer := NewCustomer(aaaEntityID, AAAEntityTypeOrganization, customerCode)
	customer.OrganizationID = &organizationID
	return customer
}

// IsUserCustomer checks if the customer is associated with a user
func (c *Customer) IsUserCustomer() bool {
	return c.AAAEntityType == AAAEntityTypeUser
}

// IsOrganizationCustomer checks if the customer is associated with an organization
func (c *Customer) IsOrganizationCustomer() bool {
	return c.AAAEntityType == AAAEntityTypeOrganization
}

// GetEntityID returns the appropriate entity ID based on type
func (c *Customer) GetEntityID() string {
	if c.IsUserCustomer() && c.UserID != nil {
		return *c.UserID
	}
	if c.IsOrganizationCustomer() && c.OrganizationID != nil {
		return *c.OrganizationID
	}
	return c.AAAEntityID
}
