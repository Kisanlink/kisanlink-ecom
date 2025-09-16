package events

import (
	"time"

	"github.com/shopspring/decimal"
)

// EventMetadata contains common metadata for all events
type EventMetadata struct {
	Source        string `json:"source"`
	Version       string `json:"version"`
	CorrelationID string `json:"correlation_id"`
}

// OrderCreatedEventData represents the data for order created events
type OrderCreatedEventData struct {
	OrderID              string               `json:"order_id"`
	OrderNumber          string               `json:"order_number"`
	BuyerOrganizationID  string               `json:"buyer_organization_id"`
	SellerOrganizationID string               `json:"seller_organization_id"`
	TotalAmount          decimal.Decimal      `json:"total_amount"`
	Status               string               `json:"status"`
	CreatedAt            time.Time            `json:"created_at"`
	Items                []OrderItemEventData `json:"items"`
}

// OrderItemEventData represents order item data in events
type OrderItemEventData struct {
	CatalogItemID   string          `json:"catalog_item_id"`
	CatalogItemType string          `json:"catalog_item_type"`
	Quantity        decimal.Decimal `json:"quantity"`
	UnitPrice       decimal.Decimal `json:"unit_price"`
	TotalPrice      decimal.Decimal `json:"total_price"`
}

// OrderStatusUpdatedEventData represents the data for order status update events
type OrderStatusUpdatedEventData struct {
	OrderID        string    `json:"order_id"`
	OrderNumber    string    `json:"order_number"`
	PreviousStatus string    `json:"previous_status"`
	NewStatus      string    `json:"new_status"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CatalogItemCreatedEventData represents the data for catalog item created events
type CatalogItemCreatedEventData struct {
	CatalogItemID  string          `json:"catalog_item_id"`
	GlobalID       string          `json:"global_id"`
	OrganizationID string          `json:"organization_id"`
	ItemType       string          `json:"item_type"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	BasePrice      decimal.Decimal `json:"base_price"`
	Currency       string          `json:"currency"`
	IsActive       bool            `json:"is_active"`
	Visibility     string          `json:"visibility"`
	CreatedAt      time.Time       `json:"created_at"`
}

// CatalogItemUpdatedEventData represents the data for catalog item updated events
type CatalogItemUpdatedEventData struct {
	CatalogItemID  string          `json:"catalog_item_id"`
	GlobalID       string          `json:"global_id"`
	OrganizationID string          `json:"organization_id"`
	ItemType       string          `json:"item_type"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	BasePrice      decimal.Decimal `json:"base_price"`
	Currency       string          `json:"currency"`
	IsActive       bool            `json:"is_active"`
	Visibility     string          `json:"visibility"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// InventoryAdjustedEventData represents the data for inventory adjustment events
type InventoryAdjustedEventData struct {
	LotID            string  `json:"lot_id"`
	CatalogItemID    string  `json:"catalog_item_id"`
	PreviousQuantity float64 `json:"previous_quantity"`
	NewQuantity      float64 `json:"new_quantity"`
	AdjustmentAmount float64 `json:"adjustment_amount"`
	Reason           string  `json:"reason"`
}
