package orders

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/shopspring/decimal"
)

// OrderStatus represents the possible order states
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusRefunded  OrderStatus = "refunded"
	// always extensible
)

// Address represents shipping/billing address
type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// Order represents the main order entity
type Order struct {
	base.BaseModel

	OrderNumber          string      `json:"order_number" gorm:"type:varchar(50);uniqueIndex;not null"`
	Status               OrderStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`
	BuyerOrganizationID  string      `json:"buyer_organization_id" gorm:"type:varchar(255);not null;index"`
	SellerOrganizationID string      `json:"seller_organization_id" gorm:"type:varchar(255);not null;index"`
	BuyerUserID          string      `json:"buyer_user_id" gorm:"type:varchar(255);not null"`

	// Financial Information
	SubtotalAmount decimal.Decimal `json:"subtotal_amount" gorm:"type:decimal(12,2);not null;default:0.00"`
	TaxAmount      decimal.Decimal `json:"tax_amount" gorm:"type:decimal(12,2);not null;default:0.00"`
	DiscountAmount decimal.Decimal `json:"discount_amount" gorm:"type:decimal(12,2);not null;default:0.00"`
	ShippingAmount decimal.Decimal `json:"shipping_amount" gorm:"type:decimal(12,2);not null;default:0.00"`
	TotalAmount    decimal.Decimal `json:"total_amount" gorm:"type:decimal(12,2);not null;default:0.00"`

	// Shipping Information
	ShippingAddress       string     `json:"shipping_address" gorm:"type:jsonb"`
	EstimatedDeliveryDate *time.Time `json:"estimated_delivery_date"`
	ActualDeliveryDate    *time.Time `json:"actual_delivery_date"`

	// Metadata
	Notes    string `json:"notes" gorm:"type:text"`
	Metadata string `json:"metadata" gorm:"type:jsonb"`

	// Relationships
	Items         []OrderItem          `json:"items,omitempty" gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE"`
	StatusHistory []OrderStatusHistory `json:"status_history,omitempty" gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE"`
}

// TableName returns the table name for GORM
func (Order) TableName() string {
	return "orders"
}

// NewOrder creates a new Order instance
func NewOrder(buyerOrgID, sellerOrgID, buyerUserID string) *Order {
	return &Order{
		BaseModel:            *base.NewBaseModel("ORD", "large"),
		OrderNumber:          generateOrderNumber(), // kisanlink-db.hash.GenerateRandom() -> choice (use this)
		Status:               OrderStatusPending,
		BuyerOrganizationID:  buyerOrgID,
		SellerOrganizationID: sellerOrgID,
		BuyerUserID:          buyerUserID,
		SubtotalAmount:       decimal.Zero,
		TaxAmount:            decimal.Zero,
		DiscountAmount:       decimal.Zero,
		ShippingAmount:       decimal.Zero,
		TotalAmount:          decimal.Zero,
	}
}

// choice: kisanlink-db hashing
// generateOrderNumber generates a unique order number
func generateOrderNumber() string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%d", timestamp)
}

// OrderItem represents individual items in an order
type OrderItem struct {
	base.BaseModel

	OrderID         string `json:"order_id" gorm:"type:varchar(255);not null;index"`
	CatalogItemID   string `json:"catalog_item_id" gorm:"type:varchar(255);not null;index"`
	CatalogItemType string `json:"catalog_item_type" gorm:"type:varchar(20);not null"` // 'product', 'service', 'labour'
	CatalogItemName string `json:"catalog_item_name" gorm:"type:varchar(255);not null"`
	CatalogItemSKU  string `json:"catalog_item_sku" gorm:"type:varchar(100)"`

	// Quantity and Pricing
	Quantity   decimal.Decimal `json:"quantity" gorm:"type:decimal(10,3);not null"`
	UnitPrice  decimal.Decimal `json:"unit_price" gorm:"type:decimal(10,2);not null"`
	TotalPrice decimal.Decimal `json:"total_price" gorm:"type:decimal(12,2);not null"`

	// Tax and Discounts
	TaxRate        decimal.Decimal `json:"tax_rate" gorm:"type:decimal(5,4);default:0.0000"`
	TaxAmount      decimal.Decimal `json:"tax_amount" gorm:"type:decimal(10,2);default:0.00"`
	DiscountRate   decimal.Decimal `json:"discount_rate" gorm:"type:decimal(5,4);default:0.0000"`
	DiscountAmount decimal.Decimal `json:"discount_amount" gorm:"type:decimal(10,2);default:0.00"`

	// Metadata
	Metadata string `json:"metadata" gorm:"type:jsonb"`
}

// TableName returns the table name for GORM
func (OrderItem) TableName() string {
	return "order_items"
}

// NewOrderItem creates a new OrderItem instance
func NewOrderItem(orderID, catalogItemID, catalogItemType, catalogItemName, catalogItemSKU string, quantity, unitPrice decimal.Decimal) *OrderItem {
	return &OrderItem{
		BaseModel:       *base.NewBaseModel("ITEM", "large"),
		OrderID:         orderID,
		CatalogItemID:   catalogItemID,
		CatalogItemType: catalogItemType,
		CatalogItemName: catalogItemName,
		CatalogItemSKU:  catalogItemSKU,
		Quantity:        quantity,
		UnitPrice:       unitPrice,
		TotalPrice:      quantity.Mul(unitPrice),
		TaxRate:         decimal.Zero,
		TaxAmount:       decimal.Zero,
		DiscountRate:    decimal.Zero,
		DiscountAmount:  decimal.Zero,
	}
}

// OrderStatusHistory represents the audit trail of order status changes
type OrderStatusHistory struct {
	base.BaseModel

	OrderID                 string      `json:"order_id" gorm:"type:varchar(255);not null;index"`
	FromStatus              *string     `json:"from_status" gorm:"type:varchar(20)"`
	ToStatus                OrderStatus `json:"to_status" gorm:"type:varchar(20);not null"`
	Reason                  string      `json:"reason" gorm:"type:varchar(500)"`
	ChangedByUserID         string      `json:"changed_by_user_id" gorm:"type:varchar(255);not null"`
	ChangedByOrganizationID string      `json:"changed_by_organization_id" gorm:"type:varchar(255)"`
	Metadata                string      `json:"metadata" gorm:"type:jsonb"`
}

// TableName returns the table name for GORM
func (OrderStatusHistory) TableName() string {
	return "order_status_history"
}

// NewOrderStatusHistory creates a new status history entry
func NewOrderStatusHistory(orderID string, fromStatus *OrderStatus, toStatus OrderStatus, reason, changedByUserID, changedByOrgID string) *OrderStatusHistory {
	var fromStatusStr *string
	if fromStatus != nil {
		str := string(*fromStatus)
		fromStatusStr = &str
	}

	return &OrderStatusHistory{
		BaseModel:               *base.NewBaseModel("HIST", "large"),
		OrderID:                 orderID,
		FromStatus:              fromStatusStr,
		ToStatus:                toStatus,
		Reason:                  reason,
		ChangedByUserID:         changedByUserID,
		ChangedByOrganizationID: changedByOrgID,
	}
}

// Order methods

// CalculateTotal calculates the total amount for the order
func (o *Order) CalculateTotal() {
	var subtotal decimal.Decimal
	for _, item := range o.Items {
		subtotal = subtotal.Add(item.TotalPrice)
	}
	o.SubtotalAmount = subtotal
	o.TotalAmount = subtotal.Add(o.TaxAmount).Add(o.ShippingAmount).Sub(o.DiscountAmount)
}

// CanTransitionTo checks if the order can transition to the given status
func (o *Order) CanTransitionTo(newStatus OrderStatus) bool {
	switch o.Status {
	case OrderStatusPending:
		return newStatus == OrderStatusConfirmed || newStatus == OrderStatusCancelled
	case OrderStatusConfirmed:
		return newStatus == OrderStatusPaid || newStatus == OrderStatusCancelled
	case OrderStatusPaid:
		return newStatus == OrderStatusShipped || newStatus == OrderStatusCancelled
	case OrderStatusShipped:
		return newStatus == OrderStatusDelivered || newStatus == OrderStatusCancelled
	case OrderStatusDelivered:
		return newStatus == OrderStatusCompleted || newStatus == OrderStatusRefunded
	case OrderStatusCompleted, OrderStatusCancelled, OrderStatusRefunded:
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

// UpdateStatus updates the order status and creates a history entry
func (o *Order) UpdateStatus(newStatus OrderStatus, reason, changedByUserID, changedByOrgID string) error {
	if !o.CanTransitionTo(newStatus) {
		return fmt.Errorf("invalid status transition from %s to %s", o.Status, newStatus)
	}

	oldStatus := o.Status
	o.Status = newStatus

	// Create history entry
	history := NewOrderStatusHistory(o.ID, &oldStatus, newStatus, reason, changedByUserID, changedByOrgID)
	o.StatusHistory = append(o.StatusHistory, *history)

	return nil
}

// GetValidTransitions returns the valid status transitions from current status
func (o *Order) GetValidTransitions() []OrderStatus {
	switch o.Status {
	case OrderStatusPending:
		return []OrderStatus{OrderStatusConfirmed, OrderStatusCancelled}
	case OrderStatusConfirmed:
		return []OrderStatus{OrderStatusPaid, OrderStatusCancelled}
	case OrderStatusPaid:
		return []OrderStatus{OrderStatusShipped, OrderStatusCancelled}
	case OrderStatusShipped:
		return []OrderStatus{OrderStatusDelivered, OrderStatusCancelled}
	case OrderStatusDelivered:
		return []OrderStatus{OrderStatusCompleted, OrderStatusRefunded}
	default:
		return []OrderStatus{}
	}
}

// OrderSummary represents a summary view of an order
type OrderSummary struct {
	ID                string          `json:"id"`
	OrderNumber       string          `json:"order_number"`
	Status            OrderStatus     `json:"status"`
	TotalAmount       decimal.Decimal `json:"total_amount"`
	ItemCount         int             `json:"item_count"`
	BuyerOrgID        string          `json:"buyer_org_id"`
	SellerOrgID       string          `json:"seller_org_id"`
	CreatedAt         time.Time       `json:"created_at"`
	EstimatedDelivery *time.Time      `json:"estimated_delivery,omitempty"`
}

// OrderFilter represents filters for order queries
type OrderFilter struct {
	Status    *OrderStatus     `json:"status,omitempty"`
	BuyerID   string           `json:"buyer_id,omitempty"`
	SellerID  string           `json:"seller_id,omitempty"`
	MinAmount *decimal.Decimal `json:"min_amount,omitempty"`
	MaxAmount *decimal.Decimal `json:"max_amount,omitempty"`
	DateFrom  *time.Time       `json:"date_from,omitempty"`
	DateTo    *time.Time       `json:"date_to,omitempty"`
}

// Address handling methods

// SetShippingAddress sets the shipping address from an Address struct
func (o *Order) SetShippingAddress(address *Address) error {
	if address == nil {
		o.ShippingAddress = ""
		return nil
	}

	addressJSON, err := json.Marshal(address)
	if err != nil {
		return fmt.Errorf("failed to marshal shipping address: %w", err)
	}

	o.ShippingAddress = string(addressJSON)
	return nil
}

// GetShippingAddress returns the shipping address as an Address struct
func (o *Order) GetShippingAddress() (*Address, error) {
	if o.ShippingAddress == "" {
		return nil, nil
	}

	var address Address
	if err := json.Unmarshal([]byte(o.ShippingAddress), &address); err != nil {
		return nil, fmt.Errorf("failed to unmarshal shipping address: %w", err)
	}

	return &address, nil
}

// Metadata handling methods

// SetMetadata sets the metadata from a map
func (o *Order) SetMetadata(metadata map[string]interface{}) error {
	if metadata == nil {
		o.Metadata = ""
		return nil
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	o.Metadata = string(metadataJSON)
	return nil
}

// GetMetadata returns the metadata as a map
func (o *Order) GetMetadata() (map[string]interface{}, error) {
	if o.Metadata == "" {
		return nil, nil
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(o.Metadata), &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return metadata, nil
}

// OrderItem metadata handling methods

// SetMetadata sets the metadata from a map
func (oi *OrderItem) SetMetadata(metadata map[string]interface{}) error {
	if metadata == nil {
		oi.Metadata = ""
		return nil
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	oi.Metadata = string(metadataJSON)
	return nil
}

// GetMetadata returns the metadata as a map
func (oi *OrderItem) GetMetadata() (map[string]interface{}, error) {
	if oi.Metadata == "" {
		return nil, nil
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(oi.Metadata), &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return metadata, nil
}

// OrderStatusHistory metadata handling methods

// SetMetadata sets the metadata from a map
func (osh *OrderStatusHistory) SetMetadata(metadata map[string]interface{}) error {
	if metadata == nil {
		osh.Metadata = ""
		return nil
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	osh.Metadata = string(metadataJSON)
	return nil
}

// GetMetadata returns the metadata as a map
func (osh *OrderStatusHistory) GetMetadata() (map[string]interface{}, error) {
	if osh.Metadata == "" {
		return nil, nil
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(osh.Metadata), &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return metadata, nil
}
