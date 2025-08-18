package user

import (
	"fmt"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// EntityType represents the type of entity in aaa-service
type EntityType string

const (
	EntityTypeUser         EntityType = "user"
	EntityTypeOrganization EntityType = "organization"
)

// Customer represents a customer in the ecommerce system
// This is a reference to a user in aaa-service
type Customer struct {
	*base.BaseModel
	// Reference to aaa-service
	AAAEntityID   string     `json:"aaa_entity_id" gorm:"size:255;not null;index" validate:"required"`
	AAAEntityType EntityType `json:"aaa_entity_type" gorm:"size:50;not null;default:'user'" validate:"required"`

	// Ecommerce specific fields
	CustomerCode string `json:"customer_code" gorm:"size:100;unique;not null" validate:"required"`
	Status       string `json:"status" gorm:"size:50;not null;default:'active'" validate:"required"`
	IsVerified   bool   `json:"is_verified" gorm:"default:false"`

	// Optional fields that can be cached from aaa-service
	DisplayName string `json:"display_name" gorm:"size:255"`
	Email       string `json:"email" gorm:"size:255"`
	Phone       string `json:"phone" gorm:"size:20"`
}

// Collaborator represents a collaborator in the ecommerce system
// This can reference either a user or organization in aaa-service
type Collaborator struct {
	*base.BaseModel
	// Reference to aaa-service
	AAAEntityID   string     `json:"aaa_entity_id" gorm:"size:255;not null;index" validate:"required"`
	AAAEntityType EntityType `json:"aaa_entity_type" gorm:"size:50;not null" validate:"required"`

	// Ecommerce specific fields
	CollaboratorCode string `json:"collaborator_code" gorm:"size:100;unique;not null" validate:"required"`
	Role             string `json:"role" gorm:"size:50;not null" validate:"required"`
	Status           string `json:"status" gorm:"size:50;not null;default:'active'" validate:"required"`
	Permissions      string `json:"permissions" gorm:"type:jsonb"` // JSON array of permissions

	// Optional fields that can be cached from aaa-service
	DisplayName string `json:"display_name" gorm:"size:255"`
	Email       string `json:"email" gorm:"size:255"`
}

// Vendor represents a vendor in the ecommerce system
// This references an organization in aaa-service
type Vendor struct {
	*base.BaseModel
	// Reference to aaa-service
	AAAEntityID   string     `json:"aaa_entity_id" gorm:"size:255;not null;index" validate:"required"`
	AAAEntityType EntityType `json:"aaa_entity_type" gorm:"size:50;not null;default:'organization'" validate:"required"`

	// Ecommerce specific fields
	VendorCode   string  `json:"vendor_code" gorm:"size:100;unique;not null" validate:"required"`
	BusinessName string  `json:"business_name" gorm:"size:255;not null" validate:"required"`
	Status       string  `json:"status" gorm:"size:50;not null;default:'active'" validate:"required"`
	IsVerified   bool    `json:"is_verified" gorm:"default:false"`
	Commission   float64 `json:"commission" gorm:"type:decimal(5,2);default:0.00"` // Commission percentage

	// Optional fields that can be cached from aaa-service
	Description string `json:"description" gorm:"type:text"`
	Address     string `json:"address" gorm:"type:text"`
}

// Status constants
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusPending  = "pending"
	StatusBlocked  = "blocked"
)

// NewCustomer creates a new Customer instance
func NewCustomer(aaaEntityID string, customerCode string) *Customer {
	return &Customer{
		BaseModel:     base.NewBaseModel("CUST", hash.Medium), // Medium size for customer IDs
		AAAEntityID:   aaaEntityID,
		AAAEntityType: EntityTypeUser,
		CustomerCode:  customerCode,
		Status:        StatusActive,
		IsVerified:    false,
	}
}

// NewCollaborator creates a new Collaborator instance
func NewCollaborator(aaaEntityID string, aaaEntityType EntityType, collaboratorCode string, role string) *Collaborator {
	return &Collaborator{
		BaseModel:        base.NewBaseModel("COLL", hash.Medium), // Medium size for collaborator IDs
		AAAEntityID:      aaaEntityID,
		AAAEntityType:    aaaEntityType,
		CollaboratorCode: collaboratorCode,
		Role:             role,
		Status:           StatusActive,
	}
}

// NewVendor creates a new Vendor instance
func NewVendor(aaaEntityID string, businessName string, vendorCode string) *Vendor {
	return &Vendor{
		BaseModel:     base.NewBaseModel("VEND", hash.Medium), // Medium size for vendor IDs
		AAAEntityID:   aaaEntityID,
		AAAEntityType: EntityTypeOrganization,
		BusinessName:  businessName,
		VendorCode:    vendorCode,
		Status:        StatusActive,
		IsVerified:    false,
		Commission:    0.00,
	}
}

// BeforeCreate validation for Customer
func (c *Customer) BeforeCreate() error {
	if err := c.BaseModel.BeforeCreate(); err != nil {
		return err
	}

	if c.AAAEntityID == "" {
		return fmt.Errorf("aaa entity ID is required")
	}
	if c.CustomerCode == "" {
		return fmt.Errorf("customer code is required")
	}

	return nil
}

// BeforeCreate validation for Collaborator
func (c *Collaborator) BeforeCreate() error {
	if err := c.BaseModel.BeforeCreate(); err != nil {
		return err
	}

	if c.AAAEntityID == "" {
		return fmt.Errorf("aaa entity ID is required")
	}
	if c.CollaboratorCode == "" {
		return fmt.Errorf("collaborator code is required")
	}
	if c.Role == "" {
		return fmt.Errorf("role is required")
	}

	return nil
}

// BeforeCreate validation for Vendor
func (v *Vendor) BeforeCreate() error {
	if err := v.BaseModel.BeforeCreate(); err != nil {
		return err
	}

	if v.AAAEntityID == "" {
		return fmt.Errorf("aaa entity ID is required")
	}
	if v.BusinessName == "" {
		return fmt.Errorf("business name is required")
	}
	if v.VendorCode == "" {
		return fmt.Errorf("vendor code is required")
	}

	return nil
}

// BeforeUpdate, BeforeDelete, BeforeSoftDelete implementations
func (c *Customer) BeforeUpdate() error     { return c.BaseModel.BeforeUpdate() }
func (c *Customer) BeforeDelete() error     { return c.BaseModel.BeforeDelete() }
func (c *Customer) BeforeSoftDelete() error { return c.BaseModel.BeforeSoftDelete() }

func (c *Collaborator) BeforeUpdate() error     { return c.BaseModel.BeforeUpdate() }
func (c *Collaborator) BeforeDelete() error     { return c.BaseModel.BeforeDelete() }
func (c *Collaborator) BeforeSoftDelete() error { return c.BaseModel.BeforeSoftDelete() }

func (v *Vendor) BeforeUpdate() error     { return v.BaseModel.BeforeUpdate() }
func (v *Vendor) BeforeDelete() error     { return v.BaseModel.BeforeDelete() }
func (v *Vendor) BeforeSoftDelete() error { return v.BaseModel.BeforeSoftDelete() }
