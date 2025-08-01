package order

import (
	"fmt"

	"kisanlink-ecom/internal/models/product"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Order represents an order in the system.
// @Description Order model representing a customer order
type Order struct {
	base.BaseModel
	UserID     string      `json:"user_id" validate:"required" gorm:"type:varchar(255);not null;index" example:"USER123456789"`
	Items      []OrderItem `json:"items" validate:"required,min=1" gorm:"serializer:json"`
	TotalPrice float64     `json:"total_price" validate:"required,min=0" gorm:"type:decimal(10,2);not null" example:"29.99"`
	Currency   string      `json:"currency" validate:"required,len=3" gorm:"type:varchar(3);default:'USD'" example:"USD"`
	Status     OrderStatus `json:"status" validate:"required" gorm:"type:varchar(50);default:'pending'" example:"pending"`
}

// NewOrder creates a new Order with initialized fields.
func NewOrder(userID string, items []OrderItem, totalPrice float64, currency string) *Order {
	baseModel := base.NewBaseModel("ORDER", hash.Medium)
	return &Order{
		BaseModel:  *baseModel,
		UserID:     userID,
		Items:      items,
		TotalPrice: totalPrice,
		Currency:   currency,
		Status:     OrderStatusPending,
	}
}

// BeforeCreate implements the base model interface.
func (o *Order) BeforeCreate() error {
	if err := o.BaseModel.BeforeCreate(); err != nil {
		return err
	}

	if o.UserID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	if len(o.Items) == 0 {
		return fmt.Errorf("order must have at least one item")
	}
	if o.TotalPrice < 0 {
		return fmt.Errorf("total price cannot be negative")
	}

	return nil
}

// OrderItem represents an item in an order.
// @Description OrderItem model representing an item within an order
type OrderItem struct {
	ProductID string           `json:"product_id" validate:"required" example:"PRODUCT123456789"`
	Quantity  int              `json:"quantity" validate:"required,min=1" example:"2"`
	Price     float64          `json:"price" validate:"required,min=0" example:"14.99"`
	Product   *product.Product `json:"product,omitempty"`
}

// OrderStatus represents order status.
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)
