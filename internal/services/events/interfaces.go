package events

import (
	"context"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/catalog"
	"github.com/Kisanlink/kisanlink-ecom/entities/models/orders"
	"github.com/Kisanlink/kisanlink-ecom/entities/models/outbox"
)

// EventPublisher defines the interface for publishing domain events
type EventPublisher interface {
	// PublishEvent publishes an event to the outbox for reliable delivery
	PublishEvent(ctx context.Context, event *outbox.OutboxEvent) error

	// ProcessUnpublishedEvents processes events that failed to publish
	ProcessUnpublishedEvents(ctx context.Context) error

	// PublishOrderCreated publishes an order created event
	PublishOrderCreated(ctx context.Context, order *orders.Order) error

	// PublishOrderStatusUpdated publishes an order status updated event
	PublishOrderStatusUpdated(ctx context.Context, order *orders.Order, previousStatus string) error

	// PublishCatalogItemCreated publishes a catalog item created event
	PublishCatalogItemCreated(ctx context.Context, catalogItem *catalog.CatalogItem) error

	// PublishCatalogItemUpdated publishes a catalog item updated event
	PublishCatalogItemUpdated(ctx context.Context, catalogItem *catalog.CatalogItem) error

	// PublishInventoryAdjusted publishes an inventory adjusted event
	PublishInventoryAdjusted(ctx context.Context, lotID, catalogItemID string, previousQuantity, newQuantity float64, reason string) error
}

// EventFactory defines the interface for creating domain events
type EventFactoryInterface interface {
	// CreateOrderCreatedEvent creates an event for order creation
	CreateOrderCreatedEvent(order *orders.Order) (*outbox.OutboxEvent, error)

	// CreateOrderStatusUpdatedEvent creates an event for order status updates
	CreateOrderStatusUpdatedEvent(order *orders.Order, previousStatus string) (*outbox.OutboxEvent, error)

	// CreateCatalogItemCreatedEvent creates an event for catalog item creation
	CreateCatalogItemCreatedEvent(catalogItem *catalog.CatalogItem) (*outbox.OutboxEvent, error)

	// CreateCatalogItemUpdatedEvent creates an event for catalog item updates
	CreateCatalogItemUpdatedEvent(catalogItem *catalog.CatalogItem) (*outbox.OutboxEvent, error)

	// CreateInventoryAdjustedEvent creates an event for inventory adjustments
	CreateInventoryAdjustedEvent(lotID, catalogItemID string, previousQuantity, newQuantity float64, reason string) (*outbox.OutboxEvent, error)
}
