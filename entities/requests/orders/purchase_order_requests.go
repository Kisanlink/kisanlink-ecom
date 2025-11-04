package orders

import (
	"kisanlink-ecom/entities/models/orders"

	"github.com/shopspring/decimal"
)

// CreatePurchaseOrderRequest represents the request to create a manual purchase order
type CreatePurchaseOrderRequest struct {
	VendorName      string                   `json:"vendor_name" binding:"required,max=255" validate:"required,min=1,max=255" example:"ABC Suppliers"`
	VendorContact   string                   `json:"vendor_contact" binding:"omitempty,max=100" validate:"omitempty,max=100" example:"+91-9876543210"`
	VendorID        *string                  `json:"vendor_id" binding:"omitempty" validate:"omitempty,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
	Items           []CreatePOItemRequest    `json:"items" binding:"required,min=1,dive" validate:"required,min=1,dive"`
	DeliveryAddress *Address                 `json:"delivery_address" validate:"omitempty"`
	Notes           string                   `json:"notes" validate:"omitempty,max=1000" example:"Urgent delivery required"`
	Metadata        map[string]interface{}   `json:"metadata" validate:"omitempty"`
}

// CreatePOItemRequest represents an item in a purchase order
type CreatePOItemRequest struct {
	ProductName string          `json:"product_name" binding:"required,max=255" validate:"required,min=1,max=255" example:"Organic Wheat"`
	ProductSKU  string          `json:"product_sku" binding:"omitempty,max=100" validate:"omitempty,max=100" example:"SKU-001"`
	HSNCode     string          `json:"hsn_code" binding:"omitempty,max=20" validate:"omitempty,max=20" example:"1001"`
	Quantity    decimal.Decimal `json:"quantity" binding:"required" validate:"required,gt=0" example:"100.0"`
	UnitPrice   decimal.Decimal `json:"unit_price" binding:"required" validate:"required,gte=0" example:"50.00"`
	GSTPercent  decimal.Decimal `json:"gst_percent" binding:"omitempty" validate:"omitempty,gte=0,lte=100" example:"18.0"`
}

// UpdatePurchaseOrderStatusRequest represents the request to update PO status
type UpdatePurchaseOrderStatusRequest struct {
	Status orders.POStatus `json:"status" binding:"required,oneof=PLACED CONFIRMED DELIVERED PAID CANCELLED" validate:"required,oneof=PLACED CONFIRMED DELIVERED PAID CANCELLED" example:"CONFIRMED"`
	Reason string          `json:"reason" binding:"omitempty,max=500" validate:"omitempty,max=500" example:"Vendor confirmed delivery date"`
}

// ListPurchaseOrdersRequest represents the request to list purchase orders
type ListPurchaseOrdersRequest struct {
	// Pagination
	Page     int `form:"page" binding:"omitempty,min=1" validate:"omitempty,min=1" example:"1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100" validate:"omitempty,min=1,max=100" example:"20"`

	// Filtering
	Status    *orders.POStatus `form:"status" binding:"omitempty,oneof=PLACED CONFIRMED DELIVERED PAID CANCELLED" validate:"omitempty,oneof=PLACED CONFIRMED DELIVERED PAID CANCELLED" example:"PLACED"`
	Source    *orders.POSource `form:"source" binding:"omitempty,oneof=KISANLINK MANUAL" validate:"omitempty,oneof=KISANLINK MANUAL" example:"MANUAL"`
	VendorID  *string          `form:"vendor_id" validate:"omitempty,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`

	// Date range
	CreatedAfter  *string `form:"created_after" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00" example:"2024-01-01T00:00:00Z"`
	CreatedBefore *string `form:"created_before" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00" example:"2024-12-31T23:59:59Z"`

	// Search
	Search *string `form:"search" validate:"omitempty,max=255" example:"wheat"`

	// Sorting
	SortBy    *string `form:"sort_by" binding:"omitempty,oneof=created_at updated_at total_amount po_number" validate:"omitempty,oneof=created_at updated_at total_amount po_number" example:"created_at"`
	SortOrder *string `form:"sort_order" binding:"omitempty,oneof=asc desc" validate:"omitempty,oneof=asc desc" example:"desc"`

	// Include relationships
	IncludeItems *bool `form:"include_items" example:"true"`
	IncludeGRNs  *bool `form:"include_grns" example:"false"`
}

// CreateGRNRequest represents the request to create a goods received note
type CreateGRNRequest struct {
	Items []CreateGRNItemRequest `json:"items" binding:"required,min=1,dive" validate:"required,min=1,dive"`
	Notes string                 `json:"notes" validate:"omitempty,max=1000" example:"All items received in good condition"`
}

// CreateGRNItemRequest represents an item in a GRN
type CreateGRNItemRequest struct {
	POItemID         string                 `json:"po_item_id" binding:"required" validate:"required,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
	QuantityReceived decimal.Decimal        `json:"quantity_received" binding:"required" validate:"required,gt=0" example:"100.0"`
	Condition        orders.ItemCondition   `json:"condition" binding:"required,oneof=GOOD DAMAGED PARTIAL" validate:"required,oneof=GOOD DAMAGED PARTIAL" example:"GOOD"`
	Notes            string                 `json:"notes" validate:"omitempty,max=500" example:"No issues noted"`
}

// UpdateGRNStatusRequest represents the request to update GRN status
type UpdateGRNStatusRequest struct {
	Status orders.GRNStatus `json:"status" binding:"required,oneof=PENDING CONFIRMED REJECTED" validate:"required,oneof=PENDING CONFIRMED REJECTED" example:"CONFIRMED"`
	Reason string           `json:"reason" binding:"omitempty,max=500" validate:"omitempty,max=500" example:"Quality verification completed"`
}
