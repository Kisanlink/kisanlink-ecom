package events_test

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/catalog"
	"github.com/Kisanlink/kisanlink-ecom/entities/models/orders"
	"github.com/Kisanlink/kisanlink-ecom/internal/services/events"
	"github.com/Kisanlink/kisanlink-ecom/tests/data"
)

func TestEventFactory_CreateOrderCreatedEvent(t *testing.T) {
	// Setup
	factory := events.NewEventFactory()

	order := orders.NewOrder("org-buyer", "org-seller", "user-123")
	order.OrderNumber = "ORD-2024-001"
	order.TotalAmount = decimal.NewFromFloat(150.50)
	order.Status = orders.OrderStatusPending

	item := orders.NewOrderItem(order.ID, "item-1", "product", "Test Item", "SKU-1", decimal.NewFromFloat(2), decimal.NewFromFloat(75.25))
	order.Items = []orders.OrderItem{*item}

	// Execute
	event, err := factory.CreateOrderCreatedEvent(order)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, event)

	assert.Equal(t, "OrderCreated", event.EventType)
	assert.Equal(t, "v1", event.EventVersion)
	assert.Equal(t, "order", event.AggregateType)
	assert.Equal(t, order.ID, event.AggregateID)
	assert.NotEmpty(t, event.IdempotencyKey)

	// Validate event data
	var eventData events.OrderCreatedEventData
	err = json.Unmarshal([]byte(event.EventData), &eventData)
	require.NoError(t, err)

	assert.Equal(t, order.ID, eventData.OrderID)
	assert.Equal(t, order.OrderNumber, eventData.OrderNumber)
	assert.Equal(t, order.BuyerOrganizationID, eventData.BuyerOrganizationID)
	assert.Equal(t, order.SellerOrganizationID, eventData.SellerOrganizationID)
	assert.True(t, order.TotalAmount.Equal(eventData.TotalAmount))
	assert.Equal(t, string(order.Status), eventData.Status)
	assert.Len(t, eventData.Items, 1)

	// Validate item data
	itemData := eventData.Items[0]
	assert.Equal(t, order.Items[0].CatalogItemID, itemData.CatalogItemID)
	assert.Equal(t, order.Items[0].CatalogItemType, itemData.CatalogItemType)
	assert.True(t, order.Items[0].Quantity.Equal(itemData.Quantity))
	assert.True(t, order.Items[0].UnitPrice.Equal(itemData.UnitPrice))
	assert.True(t, order.Items[0].TotalPrice.Equal(itemData.TotalPrice))

	// Validate metadata
	var metadata events.EventMetadata
	err = json.Unmarshal([]byte(event.EventMetadata), &metadata)
	require.NoError(t, err)

	assert.Equal(t, "order-management-system", metadata.Source)
	assert.Equal(t, "v1", metadata.Version)
	assert.Contains(t, metadata.CorrelationID, order.ID)
}

func TestEventFactory_CreateOrderStatusUpdatedEvent(t *testing.T) {
	// Setup
	factory := events.NewEventFactory()

	order := orders.NewOrder("org-buyer", "org-seller", "user-123")
	order.OrderNumber = "ORD-2024-001"
	order.Status = orders.OrderStatusConfirmed

	previousStatus := string(orders.OrderStatusPending)

	// Execute
	event, err := factory.CreateOrderStatusUpdatedEvent(order, previousStatus)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, event)

	assert.Equal(t, "OrderStatusUpdated", event.EventType)
	assert.Equal(t, "v1", event.EventVersion)
	assert.Equal(t, "order", event.AggregateType)
	assert.Equal(t, order.ID, event.AggregateID)

	// Validate event data
	var eventData events.OrderStatusUpdatedEventData
	err = json.Unmarshal([]byte(event.EventData), &eventData)
	require.NoError(t, err)

	assert.Equal(t, order.ID, eventData.OrderID)
	assert.Equal(t, order.OrderNumber, eventData.OrderNumber)
	assert.Equal(t, previousStatus, eventData.PreviousStatus)
	assert.Equal(t, string(order.Status), eventData.NewStatus)
}

func TestEventFactory_CreateCatalogItemCreatedEvent(t *testing.T) {
	// Setup
	factory := events.NewEventFactory()

	catalogItem := catalog.NewCatalogItem("org-123", catalog.CatalogItemTypeProduct, "Test Product", decimal.NewFromFloat(99.99))
	catalogItem.Description = "A test product"
	catalogItem.Visibility = catalog.VisibilityPublic

	// Execute
	event, err := factory.CreateCatalogItemCreatedEvent(catalogItem)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, event)

	assert.Equal(t, "CatalogItemCreated", event.EventType)
	assert.Equal(t, "v1", event.EventVersion)
	assert.Equal(t, "catalog_item", event.AggregateType)
	assert.Equal(t, catalogItem.ID, event.AggregateID)

	// Validate event data
	var eventData events.CatalogItemCreatedEventData
	err = json.Unmarshal([]byte(event.EventData), &eventData)
	require.NoError(t, err)

	assert.Equal(t, catalogItem.ID, eventData.CatalogItemID)
	assert.Equal(t, catalogItem.ID, eventData.GlobalID) // Using ID as GlobalID
	assert.Equal(t, catalogItem.OrganizationID, eventData.OrganizationID)
	assert.Equal(t, string(catalogItem.ItemType), eventData.ItemType)
	assert.Equal(t, catalogItem.Name, eventData.Name)
	assert.Equal(t, catalogItem.Description, eventData.Description)
	assert.True(t, catalogItem.BasePrice.Equal(eventData.BasePrice))
	assert.Equal(t, catalogItem.Currency, eventData.Currency)
	assert.Equal(t, catalogItem.IsActive, eventData.IsActive)
	assert.Equal(t, string(catalogItem.Visibility), eventData.Visibility)
}

func TestEventFactory_CreateCatalogItemUpdatedEvent(t *testing.T) {
	// Setup
	factory := events.NewEventFactory()

	catalogItem := catalog.NewCatalogItem("org-123", catalog.CatalogItemTypeProduct, "Updated Test Product", decimal.NewFromFloat(109.99))
	catalogItem.Description = "An updated test product"
	catalogItem.Visibility = catalog.VisibilityPrivate

	// Execute
	event, err := factory.CreateCatalogItemUpdatedEvent(catalogItem)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, event)

	assert.Equal(t, "CatalogItemUpdated", event.EventType)
	assert.Equal(t, "v1", event.EventVersion)
	assert.Equal(t, "catalog_item", event.AggregateType)
	assert.Equal(t, catalogItem.ID, event.AggregateID)

	// Validate event data
	var eventData events.CatalogItemUpdatedEventData
	err = json.Unmarshal([]byte(event.EventData), &eventData)
	require.NoError(t, err)

	assert.Equal(t, catalogItem.ID, eventData.CatalogItemID)
	assert.Equal(t, catalogItem.ID, eventData.GlobalID) // Using ID as GlobalID
	assert.Equal(t, catalogItem.OrganizationID, eventData.OrganizationID)
	assert.Equal(t, string(catalogItem.ItemType), eventData.ItemType)
	assert.Equal(t, catalogItem.Name, eventData.Name)
	assert.Equal(t, catalogItem.Description, eventData.Description)
	assert.True(t, catalogItem.BasePrice.Equal(eventData.BasePrice))
	assert.Equal(t, catalogItem.Currency, eventData.Currency)
	assert.Equal(t, catalogItem.IsActive, eventData.IsActive)
	assert.Equal(t, string(catalogItem.Visibility), eventData.Visibility)
}

func TestEventFactory_CreateInventoryAdjustedEvent(t *testing.T) {
	// Setup
	factory := events.NewEventFactory()

	lotID := "lot-123"
	catalogItemID := "item-123"
	previousQuantity := 100.0
	newQuantity := 95.0
	reason := "Sale adjustment"

	// Execute
	event, err := factory.CreateInventoryAdjustedEvent(lotID, catalogItemID, previousQuantity, newQuantity, reason)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, event)

	assert.Equal(t, "InventoryAdjusted", event.EventType)
	assert.Equal(t, "v1", event.EventVersion)
	assert.Equal(t, "inventory_lot", event.AggregateType)
	assert.Equal(t, lotID, event.AggregateID)

	// Validate event data
	var eventData events.InventoryAdjustedEventData
	err = json.Unmarshal([]byte(event.EventData), &eventData)
	require.NoError(t, err)

	assert.Equal(t, lotID, eventData.LotID)
	assert.Equal(t, catalogItemID, eventData.CatalogItemID)
	assert.Equal(t, previousQuantity, eventData.PreviousQuantity)
	assert.Equal(t, newQuantity, eventData.NewQuantity)
	assert.Equal(t, newQuantity-previousQuantity, eventData.AdjustmentAmount)
	assert.Equal(t, reason, eventData.Reason)

	// Validate metadata
	var metadata events.EventMetadata
	err = json.Unmarshal([]byte(event.EventMetadata), &metadata)
	require.NoError(t, err)

	assert.Equal(t, "order-management-system", metadata.Source)
	assert.Equal(t, "v1", metadata.Version)
	assert.Contains(t, metadata.CorrelationID, lotID)
}

func TestEventFactory_EventSerialization(t *testing.T) {
	// Test that all event types can be properly serialized and deserialized
	factory := events.NewEventFactory()

	// Test with complex order data
	order := data.CreateTestOrder()

	event, err := factory.CreateOrderCreatedEvent(order)
	require.NoError(t, err)

	// Ensure the event data is valid JSON
	var eventData map[string]interface{}
	err = json.Unmarshal([]byte(event.EventData), &eventData)
	require.NoError(t, err)

	// Ensure the metadata is valid JSON
	var metadata map[string]interface{}
	err = json.Unmarshal([]byte(event.EventMetadata), &metadata)
	require.NoError(t, err)

	// Verify required fields are present
	assert.Contains(t, eventData, "order_id")
	assert.Contains(t, eventData, "order_number")
	assert.Contains(t, eventData, "total_amount")
	assert.Contains(t, eventData, "items")

	assert.Contains(t, metadata, "source")
	assert.Contains(t, metadata, "version")
	assert.Contains(t, metadata, "correlation_id")
}
