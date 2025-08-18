package order

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// CreateOrderRequest represents request to create an order.
// @Description Request structure for creating a new order
type CreateOrderRequest struct {
	base.BaseRequest
	BuyerID    string                   `json:"buyer_id" validate:"required" example:"USER123456789"`
	BuyerType  string                   `json:"buyer_type" validate:"required" example:"user"`
	SellerID   string                   `json:"seller_id" validate:"required" example:"ORG123456789"`
	SellerType string                   `json:"seller_type" validate:"required" example:"fpo"`
	Items      []CreateOrderItemRequest `json:"items" validate:"required,min=1"`
	Notes      string                   `json:"notes"`
	Currency   string                   `json:"currency" validate:"required,len=3" example:"INR"`
}

// CreateOrderItemRequest represents a request to create an order item
type CreateOrderItemRequest struct {
	ItemID    string    `json:"item_id" validate:"required" example:"CAT123456789"`
	Type      OrderType `json:"type" validate:"required" example:"product"`
	Quantity  float64   `json:"quantity" validate:"required,min=0.01" example:"2.5"`
	UnitPrice float64   `json:"unit_price" validate:"required,min=0" example:"14.99"`
	Notes     string    `json:"notes"`
}

// UpdateOrderRequest represents request to update an order.
// @Description Request structure for updating an existing order
type UpdateOrderRequest struct {
	base.BaseRequest
	Status OrderStatus `json:"status" validate:"required" example:"confirmed"`
}

// UpdateOrderStatusRequest represents a request to update order status
type UpdateOrderStatusRequest struct {
	Status           OrderStatus `json:"status" validate:"required"`
	ExpectedDelivery *time.Time  `json:"expected_delivery"`
	Notes            string      `json:"notes"`
}

// OrderFilter represents filters for order queries
type OrderFilter struct {
	BuyerID    string        `json:"buyer_id"`
	BuyerType  string        `json:"buyer_type"`
	SellerID   string        `json:"seller_id"`
	SellerType string        `json:"seller_type"`
	Statuses   []OrderStatus `json:"statuses"`
	DateFrom   string        `json:"date_from"`
	DateTo     string        `json:"date_to"`
	MinAmount  *float64      `json:"min_amount"`
	MaxAmount  *float64      `json:"max_amount"`
	Currency   string        `json:"currency"`
}

// OrderSummary represents a summary of an order
type OrderSummary struct {
	ID               string      `json:"id"`
	OrderNumber      string      `json:"order_number"`
	BuyerID          string      `json:"buyer_id"`
	BuyerType        string      `json:"buyer_type"`
	SellerID         string      `json:"seller_id"`
	SellerType       string      `json:"seller_type"`
	Status           OrderStatus `json:"status"`
	TotalAmount      float64     `json:"total_amount"`
	Currency         string      `json:"currency"`
	ItemCount        int         `json:"item_count"`
	ExpectedDelivery *time.Time  `json:"expected_delivery"`
	CreatedAt        time.Time   `json:"created_at"`
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
