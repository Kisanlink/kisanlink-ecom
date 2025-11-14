package services

import (
	"github.com/stretchr/testify/mock"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/catalog"
	"github.com/Kisanlink/kisanlink-ecom/entities/models/orders"
	"github.com/Kisanlink/kisanlink-ecom/entities/models/outbox"
)

// MockEventFactory is a mock implementation of EventFactoryInterface
type MockEventFactory struct {
	mock.Mock
}

// CreateOrderCreatedEvent mocks the CreateOrderCreatedEvent method
func (m *MockEventFactory) CreateOrderCreatedEvent(order *orders.Order) (*outbox.OutboxEvent, error) {
	args := m.Called(order)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*outbox.OutboxEvent), args.Error(1)
}

// CreateOrderStatusUpdatedEvent mocks the CreateOrderStatusUpdatedEvent method
func (m *MockEventFactory) CreateOrderStatusUpdatedEvent(order *orders.Order, previousStatus string) (*outbox.OutboxEvent, error) {
	args := m.Called(order, previousStatus)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*outbox.OutboxEvent), args.Error(1)
}

// CreateCatalogItemCreatedEvent mocks the CreateCatalogItemCreatedEvent method
func (m *MockEventFactory) CreateCatalogItemCreatedEvent(catalogItem *catalog.CatalogItem) (*outbox.OutboxEvent, error) {
	args := m.Called(catalogItem)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*outbox.OutboxEvent), args.Error(1)
}

// CreateCatalogItemUpdatedEvent mocks the CreateCatalogItemUpdatedEvent method
func (m *MockEventFactory) CreateCatalogItemUpdatedEvent(catalogItem *catalog.CatalogItem) (*outbox.OutboxEvent, error) {
	args := m.Called(catalogItem)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*outbox.OutboxEvent), args.Error(1)
}

// CreateInventoryAdjustedEvent mocks the CreateInventoryAdjustedEvent method
func (m *MockEventFactory) CreateInventoryAdjustedEvent(lotID, catalogItemID string, previousQuantity, newQuantity float64, reason string) (*outbox.OutboxEvent, error) {
	args := m.Called(lotID, catalogItemID, previousQuantity, newQuantity, reason)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*outbox.OutboxEvent), args.Error(1)
}
