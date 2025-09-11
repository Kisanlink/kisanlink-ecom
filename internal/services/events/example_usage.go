package events

import (
    "context"
    "log"

    "github.com/shopspring/decimal"
    "github.com/sirupsen/logrus"
    "gorm.io/gorm"

    "kisanlink-ecom/entities/models/catalog"
    "kisanlink-ecom/entities/models/orders"
    "kisanlink-ecom/internal/repositories/events"
)

// ExampleUsage demonstrates how to use the event publishing service with routing
func ExampleUsage(db *gorm.DB, sqsClient SQSClient, queueURL string) {
    // Initialize dependencies
    logger := logrus.New()
    outboxRepo := events.NewOutboxRepository(db)
    eventFactory := NewEventFactory()

    // Set up routing
    orgRouter := NewDefaultOrganizationRouter()
    serviceRegistry := NewDefaultServiceRegistry()
    eventRouter := NewDefaultEventRouter(orgRouter, serviceRegistry)

    // Configure organization endpoints
    orgRouter.RegisterOrganizationEndpoint("buyer-org-123", Recipient{
        Type:     RecipientTypeOrganization,
        ID:       "buyer-org-123",
        QueueURL: "https://sqs.region.amazonaws.com/account/buyer-org-123-events",
    })

    orgRouter.RegisterOrganizationEndpoint("seller-org-456", Recipient{
        Type:     RecipientTypeOrganization,
        ID:       "seller-org-456",
        QueueURL: "https://sqs.region.amazonaws.com/account/seller-org-456-events",
    })

    // Add partner relationships
    orgRouter.AddPartnerRelationship("seller-org-456", "buyer-org-123")

    // Add custom routing rules
    eventRouter.AddRoutingRule(RoutingRule{
        EventType:     "OrderCreated",
        AggregateType: "order",
        Recipients: []Recipient{
            {Type: RecipientTypeService, ID: "logistics"},
            {Type: RecipientTypeService, ID: "analytics"},
        },
        Conditions: []Condition{
            {
                Field:    "total_amount",
                Operator: "gt",
                Value:    1000.0, // High-value orders get special routing
            },
        },
    })

    // Create the outbox publisher with routing
    publisher := NewOutboxPublisher(outboxRepo, sqsClient, queueURL, logger, eventFactory, eventRouter)

    // Example 1: Publishing an order created event (will be routed to multiple recipients)
    order := orders.NewOrder("buyer-org-123", "seller-org-456", "user-789")
    order.OrderNumber = "ORD-2024-001"

    ctx := context.Background()
    if err := publisher.PublishOrderCreated(ctx, order); err != nil {
        log.Printf("Failed to publish order created event: %v", err)
    }

    // Example 2: Publishing a catalog item created event (will be routed to discovery and pricing services)
    catalogItem := catalog.NewCatalogItem("org-123", catalog.CatalogItemTypeProduct, "Organic Tomatoes", decimal.NewFromFloat(50.0))

    if err := publisher.PublishCatalogItemCreated(ctx, catalogItem); err != nil {
        log.Printf("Failed to publish catalog item created event: %v", err)
    }

    // Example 3: Publishing an inventory adjustment event (will be routed to inventory and analytics services)
    if err := publisher.PublishInventoryAdjusted(ctx, "lot-123", "item-456", 100.0, 95.0, "Sale adjustment"); err != nil {
        log.Printf("Failed to publish inventory adjusted event: %v", err)
    }

    // Example 4: Processing unpublished events (for retry mechanism)
    if err := publisher.ProcessUnpublishedEvents(ctx); err != nil {
        log.Printf("Failed to process unpublished events: %v", err)
    }
}

// ExampleBackgroundProcessor demonstrates how to set up background processing
func ExampleBackgroundProcessor(publisher EventPublisher, outboxRepo events.OutboxRepository) {
    logger := logrus.New()

    config := &EventPublishingConfig{
        ProcessingInterval:  30, // Process every 30 seconds
        CleanupIntervalDays: 7,  // Clean up events older than 7 days
    }

    processor := NewEventProcessor(publisher, config, logger)

    ctx := context.Background()

    // Start the background processor (this would typically run in a goroutine)
    go processor.Start(ctx)

    // Start the cleanup worker (this would also run in a goroutine)
    go processor.StartCleanupWorker(ctx, outboxRepo)

    // The processor will continue running until the context is cancelled
    // or processor.Stop() is called
}

// ExampleRoutingConfiguration shows different routing strategies
func ExampleRoutingConfiguration() {
    orgRouter := NewDefaultOrganizationRouter()
    serviceRegistry := NewDefaultServiceRegistry()
    eventRouter := NewDefaultEventRouter(orgRouter, serviceRegistry)

    // 1. Organization-specific routing
    orgRouter.RegisterOrganizationEndpoint("fpo-maharashtra-001", Recipient{
        Type:     RecipientTypeOrganization,
        ID:       "fpo-maharashtra-001",
        QueueURL: "https://sqs.ap-south-1.amazonaws.com/account/fpo-maharashtra-001-events",
    })

    // 2. Service-specific routing
    serviceRegistry.RegisterService("payment-processor", Recipient{
        Type:     RecipientTypeService,
        ID:       "payment-processor",
        QueueURL: "https://sqs.ap-south-1.amazonaws.com/account/payment-service-events",
    })

    // 3. Conditional routing based on event data
    eventRouter.AddRoutingRule(RoutingRule{
        EventType:     "OrderCreated",
        AggregateType: "order",
        Recipients: []Recipient{
            {Type: RecipientTypeService, ID: "payment-processor"},
            {Type: RecipientTypeService, ID: "fraud-detection"},
        },
        Conditions: []Condition{
            {
                Field:    "total_amount",
                Operator: "gt",
                Value:    10000.0, // High-value orders need fraud detection
            },
        },
    })

    // 4. Geographic routing
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

    // 5. Partner network routing
    orgRouter.AddPartnerRelationship("fpo-maharashtra-001", "logistics-partner-west")
    orgRouter.AddPartnerRelationship("fpo-maharashtra-001", "cold-storage-mumbai")
}
