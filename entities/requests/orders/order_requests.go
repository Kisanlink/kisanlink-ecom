package orders

import (
	"kisanlink-ecom/entities/models/orders"

	"github.com/shopspring/decimal"
)

// CreateOrderRequest represents the request to create an order
type CreateOrderRequest struct {
	BuyerOrganizationID  string                   `json:"buyer_organization_id" binding:"required" validate:"required,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
	SellerOrganizationID string                   `json:"seller_organization_id" binding:"required" validate:"required,uuid4" example:"123e4567-e89b-12d3-a456-426614174001"`
	Items                []CreateOrderItemRequest `json:"items" binding:"required,min=1,dive" validate:"required,min=1,dive"`
	ShippingAddress      *Address                 `json:"shipping_address" validate:"omitempty"`
	Notes                string                   `json:"notes" validate:"omitempty,max=1000" example:"Special delivery instructions"`
	Metadata             map[string]interface{}   `json:"metadata" validate:"omitempty"`
}

// CreateOrderItemRequest represents an item in the order creation request
type CreateOrderItemRequest struct {
	CatalogItemID   string          `json:"catalog_item_id" binding:"required" validate:"required,uuid4" example:"123e4567-e89b-12d3-a456-426614174002"`
	CatalogItemType string          `json:"catalog_item_type" binding:"required,oneof=product service labour" validate:"required,oneof=product service labour" example:"product"`
	Quantity        decimal.Decimal `json:"quantity" binding:"required" validate:"required,gt=0" example:"10.5"`
	UnitPrice       decimal.Decimal `json:"unit_price" binding:"required" validate:"required,gte=0" example:"25.50"`
	Notes           string          `json:"notes" validate:"omitempty,max=500" example:"Organic variety preferred"`
}

// UpdateOrderStatusRequest represents the request to update order status
type UpdateOrderStatusRequest struct {
	Status orders.OrderStatus `json:"status" binding:"required,oneof=pending confirmed paid shipped delivered completed cancelled refunded" validate:"required,oneof=pending confirmed paid shipped delivered completed cancelled refunded" example:"confirmed"`
	Reason string             `json:"reason" binding:"required,max=500" validate:"required,min=1,max=500" example:"Payment confirmed by bank"`
}

// UpdateOrderRequest represents the request to update an order
type UpdateOrderRequest struct {
	Status               *orders.OrderStatus      `json:"status" binding:"omitempty,oneof=pending confirmed paid shipped delivered completed cancelled refunded" validate:"omitempty,oneof=pending confirmed paid shipped delivered completed cancelled refunded" example:"confirmed"`
	Reason               *string                  `json:"reason" binding:"omitempty,max=500" validate:"omitempty,min=1,max=500" example:"Payment confirmed"`
	Notes                *string                  `json:"notes" binding:"omitempty,max=1000" validate:"omitempty,max=1000" example:"Updated delivery instructions"`
	Metadata             map[string]interface{}   `json:"metadata" validate:"omitempty"`
	BuyerOrganizationID  *string                  `json:"buyer_organization_id" binding:"omitempty" validate:"omitempty,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
	SellerOrganizationID *string                  `json:"seller_organization_id" binding:"omitempty" validate:"omitempty,uuid4" example:"123e4567-e89b-12d3-a456-426614174001"`
	Items                []CreateOrderItemRequest `json:"items" binding:"omitempty,min=1,dive" validate:"omitempty,min=1,dive"`
	ShippingAddress      *Address                 `json:"shipping_address" validate:"omitempty"`
}

// ListOrdersRequest represents the request to list orders
type ListOrdersRequest struct {
	// Pagination
	Page     int `form:"page" binding:"omitempty,min=1" validate:"omitempty,min=1" example:"1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100" validate:"omitempty,min=1,max=100" example:"20"`

	// Filtering
	Status               *orders.OrderStatus `form:"status" binding:"omitempty,oneof=pending confirmed paid shipped delivered completed cancelled refunded" validate:"omitempty,oneof=pending confirmed paid shipped delivered completed cancelled refunded" example:"pending"`
	BuyerOrganizationID  *string             `form:"buyer_organization_id" validate:"omitempty,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
	SellerOrganizationID *string             `form:"seller_organization_id" validate:"omitempty,uuid4" example:"123e4567-e89b-12d3-a456-426614174001"`
	BuyerUserID          *string             `form:"buyer_user_id" validate:"omitempty,uuid4" example:"123e4567-e89b-12d3-a456-426614174002"`
	IsAdmin              *bool               `json:"-" form:"-"` // Set internally by service layer, not from request params

	// Date range
	CreatedAfter  *string `form:"created_after" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00" example:"2024-01-01T00:00:00Z"`
	CreatedBefore *string `form:"created_before" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00" example:"2024-12-31T23:59:59Z"`

	// Amount range
	MinAmount *decimal.Decimal `form:"min_amount" binding:"omitempty" validate:"omitempty,gte=0" example:"0.00"`
	MaxAmount *decimal.Decimal `form:"max_amount" binding:"omitempty" validate:"omitempty,gte=0" example:"1000.00"`

	// Search
	Search *string `form:"search" validate:"omitempty,max=255" example:"organic tomatoes"`

	// Sorting
	SortBy    *string `form:"sort_by" binding:"omitempty,oneof=created_at updated_at total_amount order_number" validate:"omitempty,oneof=created_at updated_at total_amount order_number" example:"created_at"`
	SortOrder *string `form:"sort_order" binding:"omitempty,oneof=asc desc" validate:"omitempty,oneof=asc desc" example:"desc"`

	// Include relationships
	IncludeItems   *bool `form:"include_items" example:"true"`
	IncludeHistory *bool `form:"include_history" example:"false"`
}

// Address represents shipping/billing address in requests
type Address struct {
	Street     string `json:"street" binding:"required,max=255" validate:"required,min=1,max=255" example:"123 Farm Road"`
	City       string `json:"city" binding:"required,max=100" validate:"required,min=1,max=100" example:"Rural City"`
	State      string `json:"state" binding:"required,max=100" validate:"required,min=1,max=100" example:"Maharashtra"`
	PostalCode string `json:"postal_code" binding:"required,max=20" validate:"required,min=1,max=20" example:"411001"`
	Country    string `json:"country" binding:"required,max=100" validate:"required,min=1,max=100" example:"India"`
}

// OrderFilters represents filters for order queries (internal use)
type OrderFilters struct {
	Status               *orders.OrderStatus
	BuyerOrganizationID  *string
	SellerOrganizationID *string
	BuyerUserID          *string
	CreatedAfter         *string
	CreatedBefore        *string
	MinAmount            *decimal.Decimal
	MaxAmount            *decimal.Decimal
	Search               *string
	Page                 int
	PageSize             int
	SortBy               string
	SortOrder            string
	IncludeItems         bool
	IncludeHistory       bool
}

// ToOrderFilters converts ListOrdersRequest to OrderFilters
func (r *ListOrdersRequest) ToOrderFilters() *OrderFilters {
	filters := &OrderFilters{
		Status:               r.Status,
		BuyerOrganizationID:  r.BuyerOrganizationID,
		SellerOrganizationID: r.SellerOrganizationID,
		BuyerUserID:          r.BuyerUserID,
		CreatedAfter:         r.CreatedAfter,
		CreatedBefore:        r.CreatedBefore,
		MinAmount:            r.MinAmount,
		MaxAmount:            r.MaxAmount,
		Search:               r.Search,
		Page:                 r.Page,
		PageSize:             r.PageSize,
		IncludeItems:         r.IncludeItems != nil && *r.IncludeItems,
		IncludeHistory:       r.IncludeHistory != nil && *r.IncludeHistory,
	}

	// Set defaults
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.PageSize <= 0 {
		filters.PageSize = 20
	}

	// Set sorting
	if r.SortBy != nil {
		filters.SortBy = *r.SortBy
	} else {
		filters.SortBy = "created_at"
	}

	if r.SortOrder != nil {
		filters.SortOrder = *r.SortOrder
	} else {
		filters.SortOrder = "desc"
	}

	return filters
}
