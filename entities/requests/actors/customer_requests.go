package actors

import (
    "kisanlink-ecom/entities/models/actors"
)

// CreateCustomerRequest represents the request to create a customer
type CreateCustomerRequest struct {
    AAAEntityID       string                 `json:"aaa_entity_id" binding:"required"`
    AAAEntityType     actors.AAAEntityType   `json:"aaa_entity_type" binding:"required,oneof=USER ORGANIZATION"`
    UserID            *string                `json:"user_id" binding:"omitempty"`
    OrganizationID    *string                `json:"organization_id" binding:"omitempty"`
    CustomerCode      string                 `json:"customer_code" binding:"required,min=3,max=100"`
    DisplayName       string                 `json:"display_name" binding:"omitempty,max=255"`
    Email             string                 `json:"email" binding:"omitempty,email"`
    Phone             string                 `json:"phone" binding:"omitempty,max=20"`
    BusinessType      string                 `json:"business_type" binding:"omitempty,max=50"`
    TaxID             string                 `json:"tax_id" binding:"omitempty,max=50"`
    RegistrationID    string                 `json:"registration_id" binding:"omitempty,max=100"`
    PreferredLanguage string                 `json:"preferred_language" binding:"omitempty,len=2"`
    BillingAddress    map[string]interface{} `json:"billing_address" binding:"omitempty"`
    ShippingAddress   map[string]interface{} `json:"shipping_address" binding:"omitempty"`
    Metadata          map[string]interface{} `json:"metadata" binding:"omitempty"`
}

// UpdateCustomerRequest represents the request to update customer information
type UpdateCustomerRequest struct {
    DisplayName       *string                `json:"display_name" binding:"omitempty,max=255"`
    Email             *string                `json:"email" binding:"omitempty,email"`
    Phone             *string                `json:"phone" binding:"omitempty,max=20"`
    Status            *string                `json:"status" binding:"omitempty,oneof=active inactive suspended"`
    IsVerified        *bool                  `json:"is_verified"`
    BusinessType      *string                `json:"business_type" binding:"omitempty,max=50"`
    TaxID             *string                `json:"tax_id" binding:"omitempty,max=50"`
    RegistrationID    *string                `json:"registration_id" binding:"omitempty,max=100"`
    PreferredLanguage *string                `json:"preferred_language" binding:"omitempty,len=2"`
    BillingAddress    map[string]interface{} `json:"billing_address" binding:"omitempty"`
    ShippingAddress   map[string]interface{} `json:"shipping_address" binding:"omitempty"`
    Metadata          map[string]interface{} `json:"metadata" binding:"omitempty"`
}

// ListCustomersRequest represents the request to list customers
type ListCustomersRequest struct {
    Page          int                   `form:"page" binding:"omitempty,min=1"`
    PageSize      int                   `form:"page_size" binding:"omitempty,min=1,max=100"`
    Status        *string               `form:"status" binding:"omitempty,oneof=active inactive suspended"`
    AAAEntityType *actors.AAAEntityType `form:"aaa_entity_type" binding:"omitempty,oneof=USER ORGANIZATION"`
    BusinessType  *string               `form:"business_type"`
    IsVerified    *bool                 `form:"is_verified"`
    Search        *string               `form:"search"`
    SortBy        *string               `form:"sort_by" binding:"omitempty,oneof=created_at updated_at customer_code display_name"`
    SortOrder     *string               `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

// CustomerFilter represents filters for customer queries (internal use)
type CustomerFilter struct {
    Status        *string
    AAAEntityType *actors.AAAEntityType
    BusinessType  *string
    IsVerified    *bool
    Search        *string
    Page          int
    PageSize      int
    SortBy        string
    SortOrder     string
}
