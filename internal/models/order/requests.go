package order

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// CreateOrderRequest represents request to create an order.
// @Description Request structure for creating a new order
type CreateOrderRequest struct {
	base.BaseRequest
	UserID   string      `json:"user_id" validate:"required" example:"USER123456789"`
	Items    []OrderItem `json:"items" validate:"required,min=1" example:"[{\"product_id\":\"PRODUCT123456789\",\"quantity\":2,\"price\":14.99}]"`
	Currency string      `json:"currency" validate:"required,len=3" example:"USD"`
}

// UpdateOrderRequest represents request to update an order.
// @Description Request structure for updating an existing order
type UpdateOrderRequest struct {
	base.BaseRequest
	Status OrderStatus `json:"status" validate:"required" example:"confirmed"`
}

// SimpleOrderItem represents a simplified order item structure for API responses.
// @Description Simplified order item structure for API responses
type SimpleOrderItem struct {
	ProductID string  `json:"product_id" validate:"required" example:"PRODUCT123456789"`
	Quantity  int     `json:"quantity" validate:"required,min=1" example:"2"`
	Price     float64 `json:"price" validate:"required,min=0" example:"14.99"`
}

// SimpleOrder represents a simplified order structure for API responses.
// @Description Simplified order structure for API responses
type SimpleOrder struct {
	ID         string            `json:"id" example:"ORDER123456789"`
	UserID     string            `json:"user_id" example:"USER123456789"`
	Items      []SimpleOrderItem `json:"items"`
	TotalPrice float64           `json:"total_price" example:"29.99"`
	Currency   string            `json:"currency" example:"USD"`
	Status     OrderStatus       `json:"status" example:"pending"`
}
