package events

import (
	"encoding/json"
	"fmt"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/catalog"
	"github.com/Kisanlink/kisanlink-ecom/entities/models/orders"
	"github.com/Kisanlink/kisanlink-ecom/entities/models/outbox"
)

// EventFactory creates outbox events for different domain entities
type EventFactory struct{}

// NewEventFactory creates a new event factory
func NewEventFactory() *EventFactory {
	return &EventFactory{}
}

// CreateOrderCreatedEvent creates an event for order creation
func (f *EventFactory) CreateOrderCreatedEvent(order *orders.Order) (*outbox.OutboxEvent, error) {
	eventData := OrderCreatedEventData{
		OrderID:              order.ID,
		OrderNumber:          order.OrderNumber,
		BuyerOrganizationID:  order.BuyerOrganizationID,
		SellerOrganizationID: order.SellerOrganizationID,
		TotalAmount:          order.TotalAmount,
		Status:               string(order.Status),
		CreatedAt:            order.CreatedAt,
		Items:                make([]OrderItemEventData, len(order.Items)),
	}

	// Convert order items
	for i, item := range order.Items {
		eventData.Items[i] = OrderItemEventData{
			CatalogItemID:   item.CatalogItemID,
			CatalogItemType: item.CatalogItemType,
			Quantity:        item.Quantity,
			UnitPrice:       item.UnitPrice,
			TotalPrice:      item.TotalPrice,
		}
	}

	eventDataJSON, err := json.Marshal(eventData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order created event data: %w", err)
	}

	metadata := EventMetadata{
		Source:        "order-management-system",
		Version:       "v1",
		CorrelationID: fmt.Sprintf("order-%s", order.ID),
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event metadata: %w", err)
	}

	event := outbox.NewOutboxEvent("OrderCreated", "order", order.ID, nil)
	event.EventData = string(eventDataJSON)
	event.EventMetadata = string(metadataJSON)

	return event, nil
}

// CreateOrderStatusUpdatedEvent creates an event for order status updates
func (f *EventFactory) CreateOrderStatusUpdatedEvent(order *orders.Order, previousStatus string) (*outbox.OutboxEvent, error) {
	eventData := OrderStatusUpdatedEventData{
		OrderID:        order.ID,
		OrderNumber:    order.OrderNumber,
		PreviousStatus: previousStatus,
		NewStatus:      string(order.Status),
		UpdatedAt:      order.UpdatedAt,
	}

	eventDataJSON, err := json.Marshal(eventData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order status updated event data: %w", err)
	}

	metadata := EventMetadata{
		Source:        "order-management-system",
		Version:       "v1",
		CorrelationID: fmt.Sprintf("order-%s", order.ID),
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event metadata: %w", err)
	}

	event := outbox.NewOutboxEvent("OrderStatusUpdated", "order", order.ID, nil)
	event.EventData = string(eventDataJSON)
	event.EventMetadata = string(metadataJSON)

	return event, nil
}

// CreateCatalogItemCreatedEvent creates an event for catalog item creation
func (f *EventFactory) CreateCatalogItemCreatedEvent(catalogItem *catalog.CatalogItem) (*outbox.OutboxEvent, error) {
	eventData := CatalogItemCreatedEventData{
		CatalogItemID:  catalogItem.ID,
		GlobalID:       catalogItem.ID, // Use ID as GlobalID for now
		OrganizationID: catalogItem.OrganizationID,
		ItemType:       string(catalogItem.ItemType),
		Name:           catalogItem.Name,
		Description:    catalogItem.Description,
		BasePrice:      catalogItem.BasePrice,
		Currency:       catalogItem.Currency,
		IsActive:       catalogItem.IsActive,
		Visibility:     string(catalogItem.Visibility),
		CreatedAt:      catalogItem.CreatedAt,
	}

	eventDataJSON, err := json.Marshal(eventData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal catalog item created event data: %w", err)
	}

	metadata := EventMetadata{
		Source:        "order-management-system",
		Version:       "v1",
		CorrelationID: fmt.Sprintf("catalog-%s", catalogItem.ID),
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event metadata: %w", err)
	}

	event := outbox.NewOutboxEvent("CatalogItemCreated", "catalog_item", catalogItem.ID, nil)
	event.EventData = string(eventDataJSON)
	event.EventMetadata = string(metadataJSON)

	return event, nil
}

// CreateCatalogItemUpdatedEvent creates an event for catalog item updates
func (f *EventFactory) CreateCatalogItemUpdatedEvent(catalogItem *catalog.CatalogItem) (*outbox.OutboxEvent, error) {
	eventData := CatalogItemUpdatedEventData{
		CatalogItemID:  catalogItem.ID,
		GlobalID:       catalogItem.ID, // Use ID as GlobalID for now
		OrganizationID: catalogItem.OrganizationID,
		ItemType:       string(catalogItem.ItemType),
		Name:           catalogItem.Name,
		Description:    catalogItem.Description,
		BasePrice:      catalogItem.BasePrice,
		Currency:       catalogItem.Currency,
		IsActive:       catalogItem.IsActive,
		Visibility:     string(catalogItem.Visibility),
		UpdatedAt:      catalogItem.UpdatedAt,
	}

	eventDataJSON, err := json.Marshal(eventData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal catalog item updated event data: %w", err)
	}

	metadata := EventMetadata{
		Source:        "order-management-system",
		Version:       "v1",
		CorrelationID: fmt.Sprintf("catalog-%s", catalogItem.ID),
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event metadata: %w", err)
	}

	event := outbox.NewOutboxEvent("CatalogItemUpdated", "catalog_item", catalogItem.ID, nil)
	event.EventData = string(eventDataJSON)
	event.EventMetadata = string(metadataJSON)

	return event, nil
}

// CreateInventoryAdjustedEvent creates an event for inventory adjustments
func (f *EventFactory) CreateInventoryAdjustedEvent(lotID, catalogItemID string, previousQuantity, newQuantity float64, reason string) (*outbox.OutboxEvent, error) {
	eventData := InventoryAdjustedEventData{
		LotID:            lotID,
		CatalogItemID:    catalogItemID,
		PreviousQuantity: previousQuantity,
		NewQuantity:      newQuantity,
		AdjustmentAmount: newQuantity - previousQuantity,
		Reason:           reason,
	}

	eventDataJSON, err := json.Marshal(eventData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inventory adjusted event data: %w", err)
	}

	metadata := EventMetadata{
		Source:        "order-management-system",
		Version:       "v1",
		CorrelationID: fmt.Sprintf("inventory-%s", lotID),
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event metadata: %w", err)
	}

	event := outbox.NewOutboxEvent("InventoryAdjusted", "inventory_lot", lotID, nil)
	event.EventData = string(eventDataJSON)
	event.EventMetadata = string(metadataJSON)

	return event, nil
}
