package actors

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
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
	AAAEntityID   string        `json:"aaa_entity_id" gorm:"type:varchar(255);not null;index"`
	AAAEntityType AAAEntityType `json:"aaa_entity_type" gorm:"type:varchar(20);not null;index"`

	// Internal References (optional, for caching/performance)
	UserID         *string `json:"user_id" gorm:"type:varchar(255);index"`         // Reference to user ID from AAA
	OrganizationID *string `json:"organization_id" gorm:"type:varchar(255);index"` // Reference to org ID from AAA

	// Customer-specific fields
	CustomerCode string `json:"customer_code" gorm:"type:varchar(100);not null;uniqueIndex"`
	DisplayName  string `json:"display_name" gorm:"type:varchar(255)"`
	Email        string `json:"email" gorm:"type:varchar(255)"`
	Phone        string `json:"phone" gorm:"type:varchar(20)"`
	Status       string `json:"status" gorm:"type:varchar(20);default:'active'"`
	IsVerified   bool   `json:"is_verified" gorm:"default:false"`

	// Business Information
	BusinessType      string `json:"business_type" gorm:"type:varchar(50)"` // individual, business, cooperative, etc.
	TaxID             string `json:"tax_id" gorm:"type:varchar(50)"`
	RegistrationID    string `json:"registration_id" gorm:"type:varchar(100)"`
	PreferredLanguage string `json:"preferred_language" gorm:"type:varchar(10);default:'en'"`

	// Address Information (stored as JSON for flexibility)
	BillingAddress  string `json:"billing_address" gorm:"type:jsonb"`
	ShippingAddress string `json:"shipping_address" gorm:"type:jsonb"`

	// Metadata
	Metadata string `json:"metadata" gorm:"type:jsonb"`
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
