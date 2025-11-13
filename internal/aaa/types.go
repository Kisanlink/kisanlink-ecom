package aaa

import "time"

// Address represents an address from AAA service
type Address struct {
	ID         string
	EntityID   string // Maps to user_id in AAA
	EntityType string // Always "user" from AAA
	Line1      string
	Line2      string
	City       string
	State      string
	Country    string
	PostalCode string
	Type       AddressType
	IsPrimary  bool
	IsVerified bool // Maps to is_active in AAA
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// AddressType represents the type of address
type AddressType string

// Address type constants
const (
	AddressTypeHome      AddressType = "HOME"
	AddressTypeBusiness  AddressType = "BUSINESS"
	AddressTypeBilling   AddressType = "BILLING"
	AddressTypeShipping  AddressType = "SHIPPING"
	AddressTypeWarehouse AddressType = "WAREHOUSE"
)

// CreateAddressRequest represents a request to create an address
type CreateAddressRequest struct {
	EntityID   string // user_id in AAA
	EntityType string // ignored, AAA uses user_id only
	Line1      string
	Line2      string
	City       string
	State      string
	Country    string
	PostalCode string
	Type       AddressType
	IsPrimary  bool
}

// UpdateAddressRequest represents a request to update an address
type UpdateAddressRequest struct {
	ID         string
	Line1      string
	Line2      string
	City       string
	State      string
	Country    string
	PostalCode string
	Type       AddressType
}

// Validate validates CreateAddressRequest
func (r *CreateAddressRequest) Validate() error {
	if r.EntityID == "" {
		return ErrMissingField
	}
	if r.EntityType == "" {
		return ErrMissingField
	}
	if r.Line1 == "" {
		return ErrMissingField
	}
	if r.City == "" {
		return ErrMissingField
	}
	if r.State == "" {
		return ErrMissingField
	}
	if r.Country == "" {
		return ErrMissingField
	}
	if r.PostalCode == "" {
		return ErrMissingField
	}
	return nil
}

// Validate validates UpdateAddressRequest
func (r *UpdateAddressRequest) Validate() error {
	if r.ID == "" {
		return ErrInvalidAddressID
	}
	return nil
}
