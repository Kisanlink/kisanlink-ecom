# Event Publishing Service - Recipient Routing Summary

## How the Service Determines Recipients

The Event Publishing Service uses a sophisticated **multi-layered routing system** to determine who should receive each event. Here's how it works:

### 🎯 **Routing Architecture**

```
Event → Event Router → Multiple Recipients
  ↓         ↓              ↓
Order    Organization   • Buyer Org Queue
Created  Router +       • Seller Org Queue
         Service        • Logistics Service
         Registry       • Analytics Service
                       • Payment Service (conditional)
```

### 📋 **Recipient Determination Process**

1. **Event Analysis** - Router examines event type, content, and metadata
2. **Rule Matching** - Checks custom routing rules with conditions
3. **Default Routing** - Applies built-in routing logic for event types
4. **Organization Lookup** - Finds endpoints for involved organizations
5. **Service Subscription** - Identifies subscribed federated services
6. **Deduplication** - Removes duplicate recipients
7. **Message Delivery** - Sends to all unique recipients

### 🔄 **Routing Strategies**

#### 1. **Event-Type Based Routing**

Different events automatically go to different services:

```go
// Order events → Logistics, Analytics, Organizations
"OrderCreated" → [logistics-service, analytics-service, buyer-org, seller-org]

// Catalog events → Discovery, Pricing, Partner Organizations
"CatalogItemCreated" → [discovery-service, pricing-engine, partner-orgs]

// Inventory events → Inventory Service, Analytics
"InventoryAdjusted" → [inventory-service, analytics-service]
```

#### 2. **Organization-Based Routing**

Events are routed to relevant organizations:

```go
// Order events go to buyer and seller organizations
order := {
  "buyer_organization_id": "fpo-maharashtra-001",
  "seller_organization_id": "fpo-gujarat-002"
}
// → Routes to both organization queues
```

#### 3. **Service Subscription Routing**

Services subscribe to specific event types:

```go
// Analytics service subscribes to ALL events
analyticsService.SubscribeTo("OrderCreated", "CatalogItemCreated", "InventoryAdjusted")

// Logistics only cares about orders
logisticsService.SubscribeTo("OrderCreated", "OrderStatusUpdated")
```

#### 4. **Conditional Routing**

Smart routing based on event content:

```go
// High-value orders get special treatment
if order.total_amount > 10000 {
  recipients.add("fraud-detection-service")
  recipients.add("premium-logistics-service")
}

// Regional routing
if organization.region == "north-india" {
  recipients.add("regional-discovery-north")
}
```

### 🏗️ **Implementation Components**

#### **EventRouter Interface**

```go
type EventRouter interface {
    GetRecipients(ctx context.Context, event *OutboxEvent) ([]Recipient, error)
    AddRoutingRule(rule RoutingRule) error
    RemoveRoutingRule(eventType, aggregateType string) error
}
```

#### **OrganizationRouter**

- Manages organization endpoints
- Handles partner relationships
- Enforces data sharing permissions

#### **ServiceRegistry**

- Maintains federated service endpoints
- Manages event subscriptions
- Handles service discovery

### 📨 **Message Structure**

Each message includes recipient context:

```json
{
  "MessageAttributes": {
    "EventType": "OrderCreated",
    "RecipientType": "organization",
    "RecipientID": "fpo-maharashtra-001"
  },
  "MessageBody": {
    "event_id": "EVT123",
    "data": { "order_id": "order-456" }
  }
}
```

### 🎛️ **Configuration Examples**

#### **Basic Setup**

```go
// Initialize routing
orgRouter := NewDefaultOrganizationRouter()
serviceRegistry := NewDefaultServiceRegistry()
eventRouter := NewDefaultEventRouter(orgRouter, serviceRegistry)

// Create publisher with routing
publisher := NewOutboxPublisher(repo, sqsClient, queueURL, logger, factory, eventRouter)
```

#### **Custom Routing Rules**

```go
// Route high-value orders to fraud detection
eventRouter.AddRoutingRule(RoutingRule{
    EventType: "OrderCreated",
    Recipients: []Recipient{{Type: "service", ID: "fraud-detection"}},
    Conditions: []Condition{{Field: "total_amount", Operator: "gt", Value: 10000}},
})
```

#### **Organization Endpoints**

```go
// Register organization endpoints
orgRouter.RegisterOrganizationEndpoint("fpo-001", Recipient{
    Type: "organization",
    ID: "fpo-001",
    QueueURL: "https://sqs.region.amazonaws.com/account/fpo-001-events",
})
```

### 🔍 **Recipient Discovery Flow**

For an `OrderCreated` event:

1. **Parse Event Data** - Extract buyer_org_id, seller_org_id, total_amount
2. **Check Custom Rules** - Match against conditional routing rules
3. **Organization Routing** - Find endpoints for buyer and seller orgs
4. **Service Subscriptions** - Get all services subscribed to "OrderCreated"
5. **Default Services** - Add logistics, analytics (built-in defaults)
6. **Partner Networks** - Include partner organizations if authorized
7. **Deduplicate** - Remove duplicate recipients
8. **Deliver** - Send to all unique recipients

### 📊 **Benefits**

- **🎯 Targeted Delivery** - Events only go to interested parties
- **🔧 Flexible Configuration** - Easy to add new recipients and rules
- **🛡️ Security** - Organization-scoped routing respects boundaries
- **📈 Scalable** - Handles complex federated network topologies
- **🔄 Reliable** - Partial failures don't block other recipients
- **📝 Auditable** - Full logging of routing decisions

### 🚀 **Real-World Example**

When a farmer places a ₹15,000 order for organic seeds:

```
OrderCreated Event
├── Buyer Organization (FPO Maharashtra) ✓
├── Seller Organization (Seed Company Gujarat) ✓
├── Logistics Service (for shipping) ✓
├── Analytics Service (for reporting) ✓
├── Fraud Detection Service (high-value order) ✓
├── Premium Logistics (high-value order) ✓
├── Payment Gateway (credit card payment) ✓
└── Regional Discovery North (seller region) ✓

Result: 8 recipients receive the event with appropriate context
```

This ensures that every stakeholder in the agricultural supply chain gets the information they need, when they need it, without overwhelming them with irrelevant events.
