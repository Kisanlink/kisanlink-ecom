# Event Publishing Service with Recipient Routing

This document explains how the Event Publishing Service determines recipients and routes messages in the federated network architecture.

## Overview

The Event Publishing Service uses a sophisticated routing system to determine who should receive each event. Instead of publishing to a single queue, it can route events to multiple recipients based on:

1. **Event Type & Content** - Different events go to different services
2. **Organization Relationships** - Partner organizations receive relevant events
3. **Service Subscriptions** - Services subscribe to specific event types
4. **Custom Routing Rules** - Conditional routing based on event data

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────────┐
│   Event Source  │───▶│  Event Router    │───▶│  Multiple Recipients│
│  (Order/Catalog)│    │                  │    │                     │
└─────────────────┘    │ • Organization   │    │ • Organizations     │
                       │   Router         │    │ • Services          │
                       │ • Service        │    │ • Partners          │
                       │   Registry       │    │ • Webhooks          │
                       │ • Routing Rules  │    │                     │
                       └──────────────────┘    └─────────────────────┘
```

## Recipient Types

### 1. Organization Recipients

Organizations in the federated network that should receive events:

```go
Recipient{
    Type:     RecipientTypeOrganization,
    ID:       "fpo-maharashtra-001",
    QueueURL: "https://sqs.ap-south-1.amazonaws.com/account/fpo-maharashtra-001-events",
}
```

**Use Cases:**

- Order events go to buyer and seller organizations
- Catalog updates go to partner organizations
- Inventory changes go to supply chain partners

### 2. Service Recipients

Federated services that process specific event types:

```go
Recipient{
    Type:     RecipientTypeService,
    ID:       "logistics",
    QueueURL: "https://sqs.region.amazonaws.com/account/logistics-service-events",
}
```

**Default Services:**

- **Logistics Service** - Order and shipping events
- **Analytics Service** - All events for reporting
- **Pricing Engine** - Catalog and inventory events
- **Discovery Service** - Catalog events for search
- **Inventory Service** - Inventory and order events
- **Notification Service** - Order status updates

### 3. Partner Recipients

External partners with specific integration needs:

```go
Recipient{
    Type:       RecipientTypePartner,
    ID:         "payment-gateway",
    WebhookURL: "https://partner.com/webhooks/kisanlink-events",
}
```

## Routing Strategies

### 1. Default Routing

Events are automatically routed based on their type:

| Event Type           | Default Recipients                              |
| -------------------- | ----------------------------------------------- |
| `OrderCreated`       | Buyer org, Seller org, Logistics, Analytics     |
| `OrderStatusUpdated` | Buyer org, Seller org, Logistics, Notifications |
| `CatalogItemCreated` | Discovery, Pricing, Partner orgs                |
| `CatalogItemUpdated` | Discovery, Pricing, Partner orgs                |
| `InventoryAdjusted`  | Inventory, Analytics                            |

### 2. Organization-Based Routing

Organizations can have multiple endpoints and partner relationships:

```go
// Register organization endpoints
orgRouter.RegisterOrganizationEndpoint("fpo-001", Recipient{
    Type:     RecipientTypeOrganization,
    ID:       "fpo-001",
    QueueURL: "https://sqs.region.amazonaws.com/account/fpo-001-events",
})

// Add partner relationships
orgRouter.AddPartnerRelationship("fpo-001", "logistics-partner-west")
orgRouter.AddPartnerRelationship("fpo-001", "cold-storage-mumbai")
```

### 3. Service Subscription Routing

Services subscribe to specific event types:

```go
// Analytics service subscribes to all events
serviceRegistry.SubscribeToEvent("OrderCreated", analyticsService)
serviceRegistry.SubscribeToEvent("CatalogItemCreated", analyticsService)
serviceRegistry.SubscribeToEvent("InventoryAdjusted", analyticsService)

// Logistics only cares about orders
serviceRegistry.SubscribeToEvent("OrderCreated", logisticsService)
serviceRegistry.SubscribeToEvent("OrderStatusUpdated", logisticsService)
```

### 4. Conditional Routing Rules

Custom rules based on event content:

```go
// High-value orders get special treatment
eventRouter.AddRoutingRule(RoutingRule{
    EventType:     "OrderCreated",
    AggregateType: "order",
    Recipients: []Recipient{
        {Type: RecipientTypeService, ID: "fraud-detection"},
        {Type: RecipientTypeService, ID: "premium-logistics"},
    },
    Conditions: []Condition{
        {
            Field:    "total_amount",
            Operator: "gt",
            Value:    10000.0,
        },
    },
})

// Regional routing
eventRouter.AddRoutingRule(RoutingRule{
    EventType:     "CatalogItemCreated",
    AggregateType: "catalog_item",
    Recipients: []Recipient{
        {Type: RecipientTypeService, ID: "regional-discovery-north"},
    },
    Conditions: []Condition{
        {
            Field:    "organization_region",
            Operator: "in",
            Value:    []string{"punjab", "haryana", "uttar-pradesh"},
        },
    },
})
```

## Event Message Structure

Each message includes recipient information:

```json
{
  "event_id": "EVT1234567890",
  "event_type": "OrderCreated",
  "event_version": "v1",
  "aggregate_type": "order",
  "aggregate_id": "order-123",
  "idempotency_key": "OrderCreated:order-123:1234567890",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "order_id": "order-123",
    "buyer_organization_id": "fpo-001",
    "seller_organization_id": "fpo-002",
    "total_amount": 1500.0
  },
  "metadata": {
    "source": "order-management-system",
    "version": "v1",
    "correlation_id": "order-order-123"
  }
}
```

**SQS Message Attributes:**

- `EventType` - Type of event
- `EventVersion` - Schema version
- `AggregateType` - Entity type
- `AggregateID` - Entity ID
- `IdempotencyKey` - Deduplication key
- `RecipientType` - Type of recipient
- `RecipientID` - Recipient identifier

## Configuration Examples

### Basic Setup

```go
// Initialize routing components
orgRouter := NewDefaultOrganizationRouter()
serviceRegistry := NewDefaultServiceRegistry()
eventRouter := NewDefaultEventRouter(orgRouter, serviceRegistry)

// Create publisher with routing
publisher := NewOutboxPublisher(
    outboxRepo,
    sqsClient,
    defaultQueueURL,
    logger,
    eventFactory,
    eventRouter,
)
```

### Advanced Configuration

```go
// Configure organization endpoints
orgRouter.RegisterOrganizationEndpoint("fpo-maharashtra-001", Recipient{
    Type:     RecipientTypeOrganization,
    ID:       "fpo-maharashtra-001",
    QueueURL: "https://sqs.ap-south-1.amazonaws.com/account/fpo-maharashtra-001-events",
})

// Add custom services
serviceRegistry.RegisterService("payment-processor", Recipient{
    Type:     RecipientTypeService,
    ID:       "payment-processor",
    QueueURL: "https://sqs.region.amazonaws.com/account/payment-service-events",
})

// Subscribe to events
serviceRegistry.SubscribeToEvent("OrderCreated", paymentService)

// Add conditional routing
eventRouter.AddRoutingRule(RoutingRule{
    EventType:     "OrderCreated",
    AggregateType: "order",
    Recipients: []Recipient{
        {Type: RecipientTypeService, ID: "payment-processor"},
    },
    Conditions: []Condition{
        {Field: "payment_method", Operator: "eq", Value: "credit_card"},
    },
})
```

## Benefits

1. **Scalability** - Events are only sent to interested parties
2. **Flexibility** - Easy to add new recipients and routing rules
3. **Reliability** - Partial failures don't block all recipients
4. **Security** - Organization-scoped routing respects data boundaries
5. **Performance** - Conditional routing reduces unnecessary traffic

## Monitoring & Observability

The system logs detailed information about routing decisions:

```
INFO Event published to recipients event_id=EVT123 success_count=3 total_recipients=4
WARN Failed to publish to recipient event_id=EVT123 recipient_type=service recipient_id=analytics error="queue not found"
```

This allows you to monitor:

- Routing effectiveness
- Recipient availability
- Message delivery success rates
- Performance bottlenecks
