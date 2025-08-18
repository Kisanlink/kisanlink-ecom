package order

import (
	"fmt"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// OrderType represents the type of order item
type OrderType string

const (
	OrderTypeProduct OrderType = "product"
	OrderTypeService OrderType = "service"
	OrderTypeLabour  OrderType = "labour"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusDraft      OrderStatus = "draft"
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusFulfilled  OrderStatus = "fulfilled"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusRefunded   OrderStatus = "refunded"
)

// BuyerType represents the type of buyer
type BuyerType string

const (
	BuyerTypeUser         BuyerType = "user"
	BuyerTypeFPO          BuyerType = "fpo"
	BuyerTypeCollaborator BuyerType = "collaborator"
	BuyerTypeOrganization BuyerType = "organization"
)

// SellerType represents the type of seller
type SellerType string

const (
	SellerTypeFPO          SellerType = "fpo"
	SellerTypeCollaborator SellerType = "collaborator"
	SellerTypePlatform     SellerType = "platform"
)

// Order represents a customer order
type Order struct {
	*base.BaseModel
	OrderNumber      string      `json:"order_number" gorm:"type:varchar(255);uniqueIndex;not null"`
	BuyerID          string      `json:"buyer_id" gorm:"type:varchar(255);not null;index"`  // References users or organizations table
	BuyerType        string      `json:"buyer_type" gorm:"type:varchar(20);not null"`       // "user" or "organization"
	SellerID         string      `json:"seller_id" gorm:"type:varchar(255);not null;index"` // References organizations table
	SellerType       string      `json:"seller_type" gorm:"type:varchar(20);not null"`      // "organization" or "individual"
	Status           OrderStatus `json:"status" gorm:"type:varchar(20);not null;default:'draft'"`
	PaymentIntentID  *string     `json:"payment_intent_id" gorm:"type:varchar(255)"`
	Notes            string      `json:"notes" gorm:"type:text"`
	TotalAmount      float64     `json:"total_amount" gorm:"type:decimal(10,2);not null"`
	Currency         string      `json:"currency" gorm:"type:varchar(3);not null;default:'INR'"`
	Subtotal         float64     `json:"subtotal" gorm:"type:decimal(10,2);not null"`
	TaxAmount        float64     `json:"tax_amount" gorm:"type:decimal(10,2);default:0"`
	ShippingAmount   float64     `json:"shipping_amount" gorm:"type:decimal(10,2);default:0"`
	DiscountAmount   float64     `json:"discount_amount" gorm:"type:decimal(10,2);default:0"`
	ExpectedDelivery *time.Time  `json:"expected_delivery"`
	DeliveredAt      *time.Time  `json:"delivered_at"`

	// Relationships
	Items []OrderItem `json:"items" gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE"`
}

// NewOrder creates a new Order instance
func NewOrder(buyerID, buyerType, sellerID, sellerType string, currency string) *Order {
	return &Order{
		BaseModel:      base.NewBaseModel("ORD", hash.Large), // Large size for orders (high volume)
		OrderNumber:    generateOrderNumber(),
		BuyerID:        buyerID,
		BuyerType:      buyerType,
		SellerID:       sellerID,
		SellerType:     sellerType,
		Status:         OrderStatusDraft,
		Currency:       currency,
		TotalAmount:    0.0,
		Subtotal:       0.0,
		TaxAmount:      0.0,
		ShippingAmount: 0.0,
		DiscountAmount: 0.0,
	}
}

// generateOrderNumber generates a unique order number
func generateOrderNumber() string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("ORD%d", timestamp)
}

// OrderItem represents an item within an order
type OrderItem struct {
	*base.BaseModel
	OrderID     string    `json:"order_id" gorm:"type:varchar(255);not null;index"` // References orders table
	ItemID      string    `json:"item_id" gorm:"type:varchar(255);not null;index"`  // References catalog_items table
	Type        OrderType `json:"type" gorm:"type:varchar(20);not null"`
	SKU         string    `json:"sku" gorm:"type:varchar(100);not null"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null"`
	Description string    `json:"description" gorm:"type:text"`
	Quantity    float64   `json:"quantity" gorm:"type:decimal(10,2);not null"`
	UnitPrice   float64   `json:"unit_price" gorm:"type:decimal(10,2);not null"`
	Currency    string    `json:"currency" gorm:"type:varchar(3);not null;default:'INR'"`
	TotalPrice  float64   `json:"total_price" gorm:"type:decimal(10,2);not null"`
	UOM         string    `json:"uom" gorm:"type:varchar(50)"`
	Notes       string    `json:"notes" gorm:"type:text"`
}

// NewOrderItem creates a new OrderItem instance
func NewOrderItem(orderID, itemID string, itemType OrderType, sku, name, description string, quantity, unitPrice float64, currency, uom string) *OrderItem {
	return &OrderItem{
		BaseModel:   base.NewBaseModel("ITEM", hash.Large), // Large size for order items (high volume)
		OrderID:     orderID,
		ItemID:      itemID,
		Type:        itemType,
		SKU:         sku,
		Name:        name,
		Description: description,
		Quantity:    quantity,
		UnitPrice:   unitPrice,
		Currency:    currency,
		TotalPrice:  quantity * unitPrice,
		UOM:         uom,
	}
}

// CalculateTotal calculates the total amount for the order
func (o *Order) CalculateTotal() {
	var subtotal float64
	for _, item := range o.Items {
		subtotal += item.UnitPrice * item.Quantity
	}
	o.Subtotal = subtotal
	o.TotalAmount = subtotal + o.TaxAmount + o.ShippingAmount - o.DiscountAmount
}

// CanTransitionTo checks if the order can transition to the given status
func (o *Order) CanTransitionTo(newStatus OrderStatus) bool {
	switch o.Status {
	case OrderStatusDraft:
		return newStatus == OrderStatusPending || newStatus == OrderStatusCancelled
	case OrderStatusPending:
		return newStatus == OrderStatusConfirmed || newStatus == OrderStatusCancelled
	case OrderStatusConfirmed:
		return newStatus == OrderStatusProcessing || newStatus == OrderStatusCancelled
	case OrderStatusProcessing:
		return newStatus == OrderStatusShipped || newStatus == OrderStatusCancelled
	case OrderStatusShipped:
		return newStatus == OrderStatusDelivered || newStatus == OrderStatusCancelled
	case OrderStatusDelivered:
		return newStatus == OrderStatusRefunded
	case OrderStatusFulfilled, OrderStatusCancelled, OrderStatusRefunded:
		return false // Terminal states
	default:
		return false
	}
}

// AddItem adds an item to the order
func (o *Order) AddItem(item *OrderItem) {
	o.Items = append(o.Items, *item)
	o.CalculateTotal()
}

// RemoveItem removes an item from the order by ID
func (o *Order) RemoveItem(itemID string) {
	for i, item := range o.Items {
		if item.ID == itemID {
			o.Items = append(o.Items[:i], o.Items[i+1:]...)
			o.CalculateTotal()
			break
		}
	}
}
