package events

import (
	"context"
)

// DefaultServiceRegistry implements service discovery and routing
type DefaultServiceRegistry struct {
	// Service endpoints by service type
	serviceEndpoints map[string][]Recipient

	// Event subscriptions by event type
	eventSubscriptions map[string][]Recipient
}

// NewDefaultServiceRegistry creates a new service registry
func NewDefaultServiceRegistry() *DefaultServiceRegistry {
	registry := &DefaultServiceRegistry{
		serviceEndpoints:   make(map[string][]Recipient),
		eventSubscriptions: make(map[string][]Recipient),
	}

	// Initialize with default federated services
	registry.initializeDefaultServices()

	return registry
}

// GetServiceEndpoints returns endpoints for specific services
func (r *DefaultServiceRegistry) GetServiceEndpoints(ctx context.Context, serviceType string) ([]Recipient, error) {
	endpoints, exists := r.serviceEndpoints[serviceType]
	if !exists {
		return []Recipient{}, nil
	}
	return endpoints, nil
}

// GetSubscribedServices returns services subscribed to specific event types
func (r *DefaultServiceRegistry) GetSubscribedServices(ctx context.Context, eventType string) ([]Recipient, error) {
	subscribers, exists := r.eventSubscriptions[eventType]
	if !exists {
		return []Recipient{}, nil
	}
	return subscribers, nil
}

// RegisterService registers a service endpoint
func (r *DefaultServiceRegistry) RegisterService(serviceType string, recipient Recipient) {
	if r.serviceEndpoints[serviceType] == nil {
		r.serviceEndpoints[serviceType] = []Recipient{}
	}
	r.serviceEndpoints[serviceType] = append(r.serviceEndpoints[serviceType], recipient)
}

// SubscribeToEvent subscribes a service to an event type
func (r *DefaultServiceRegistry) SubscribeToEvent(eventType string, recipient Recipient) {
	if r.eventSubscriptions[eventType] == nil {
		r.eventSubscriptions[eventType] = []Recipient{}
	}
	r.eventSubscriptions[eventType] = append(r.eventSubscriptions[eventType], recipient)
}

// initializeDefaultServices sets up default federated services
func (r *DefaultServiceRegistry) initializeDefaultServices() {
	// Logistics Service
	logisticsService := Recipient{
		Type:     RecipientTypeService,
		ID:       "logistics",
		QueueURL: "https://sqs.region.amazonaws.com/account/logistics-service-events",
	}
	r.RegisterService("logistics", logisticsService)
	r.SubscribeToEvent("OrderCreated", logisticsService)
	r.SubscribeToEvent("OrderStatusUpdated", logisticsService)

	// Analytics Service
	analyticsService := Recipient{
		Type:     RecipientTypeService,
		ID:       "analytics",
		QueueURL: "https://sqs.region.amazonaws.com/account/analytics-service-events",
	}
	r.RegisterService("analytics", analyticsService)
	r.SubscribeToEvent("OrderCreated", analyticsService)
	r.SubscribeToEvent("OrderStatusUpdated", analyticsService)
	r.SubscribeToEvent("CatalogItemCreated", analyticsService)
	r.SubscribeToEvent("CatalogItemUpdated", analyticsService)
	r.SubscribeToEvent("InventoryAdjusted", analyticsService)

	// Pricing Engine Service
	pricingService := Recipient{
		Type:     RecipientTypeService,
		ID:       "pricing",
		QueueURL: "https://sqs.region.amazonaws.com/account/pricing-engine-events",
	}
	r.RegisterService("pricing", pricingService)
	r.SubscribeToEvent("CatalogItemCreated", pricingService)
	r.SubscribeToEvent("CatalogItemUpdated", pricingService)
	r.SubscribeToEvent("InventoryAdjusted", pricingService)

	// Discovery Service
	discoveryService := Recipient{
		Type:     RecipientTypeService,
		ID:       "discovery",
		QueueURL: "https://sqs.region.amazonaws.com/account/discovery-service-events",
	}
	r.RegisterService("discovery", discoveryService)
	r.SubscribeToEvent("CatalogItemCreated", discoveryService)
	r.SubscribeToEvent("CatalogItemUpdated", discoveryService)

	// Inventory Service
	inventoryService := Recipient{
		Type:     RecipientTypeService,
		ID:       "inventory",
		QueueURL: "https://sqs.region.amazonaws.com/account/inventory-service-events",
	}
	r.RegisterService("inventory", inventoryService)
	r.SubscribeToEvent("InventoryAdjusted", inventoryService)
	r.SubscribeToEvent("OrderCreated", inventoryService)

	// Notification Service
	notificationService := Recipient{
		Type:     RecipientTypeService,
		ID:       "notifications",
		QueueURL: "https://sqs.region.amazonaws.com/account/notification-service-events",
	}
	r.RegisterService("notifications", notificationService)
	r.SubscribeToEvent("OrderCreated", notificationService)
	r.SubscribeToEvent("OrderStatusUpdated", notificationService)
}

// GetAllServices returns all registered services
func (r *DefaultServiceRegistry) GetAllServices() map[string][]Recipient {
	return r.serviceEndpoints
}

// GetAllSubscriptions returns all event subscriptions
func (r *DefaultServiceRegistry) GetAllSubscriptions() map[string][]Recipient {
	return r.eventSubscriptions
}

// UnregisterService removes a service endpoint
func (r *DefaultServiceRegistry) UnregisterService(serviceType string, recipientID string) {
	endpoints := r.serviceEndpoints[serviceType]
	for i, endpoint := range endpoints {
		if endpoint.ID == recipientID {
			r.serviceEndpoints[serviceType] = append(endpoints[:i], endpoints[i+1:]...)
			break
		}
	}
}

// UnsubscribeFromEvent removes a service subscription
func (r *DefaultServiceRegistry) UnsubscribeFromEvent(eventType string, recipientID string) {
	subscribers := r.eventSubscriptions[eventType]
	for i, subscriber := range subscribers {
		if subscriber.ID == recipientID {
			r.eventSubscriptions[eventType] = append(subscribers[:i], subscribers[i+1:]...)
			break
		}
	}
}
