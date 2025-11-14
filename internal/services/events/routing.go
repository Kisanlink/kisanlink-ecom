package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/outbox"
)

// RecipientType defines the type of recipient
type RecipientType string

const (
	RecipientTypeOrganization RecipientType = "organization"
	RecipientTypeService      RecipientType = "service"
	RecipientTypePartner      RecipientType = "partner"
	RecipientTypeAll          RecipientType = "all"
)

// Recipient represents an event recipient
type Recipient struct {
	Type       RecipientType `json:"type"`
	ID         string        `json:"id"`
	QueueURL   string        `json:"queue_url,omitempty"`
	TopicARN   string        `json:"topic_arn,omitempty"`
	WebhookURL string        `json:"webhook_url,omitempty"`
}

// RoutingRule defines how events should be routed
type RoutingRule struct {
	EventType     string      `json:"event_type"`
	AggregateType string      `json:"aggregate_type"`
	Recipients    []Recipient `json:"recipients"`
	Conditions    []Condition `json:"conditions,omitempty"`
}

// Condition defines conditional routing based on event data
type Condition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, in, contains
	Value    interface{} `json:"value"`
}

// EventRouter handles routing logic for events
type EventRouter interface {
	// GetRecipients returns the list of recipients for an event
	GetRecipients(ctx context.Context, event *outbox.OutboxEvent) ([]Recipient, error)

	// AddRoutingRule adds a new routing rule
	AddRoutingRule(rule RoutingRule) error

	// RemoveRoutingRule removes a routing rule
	RemoveRoutingRule(eventType, aggregateType string) error
}

// DefaultEventRouter implements basic routing logic
type DefaultEventRouter struct {
	rules              map[string]RoutingRule // key: eventType:aggregateType
	organizationRouter OrganizationRouter
	serviceRegistry    ServiceRegistry
}

// OrganizationRouter handles organization-specific routing
type OrganizationRouter interface {
	// GetOrganizationEndpoints returns the endpoints for an organization
	GetOrganizationEndpoints(ctx context.Context, orgID string) ([]Recipient, error)

	// GetPartnerOrganizations returns partner organizations that should receive events
	GetPartnerOrganizations(ctx context.Context, sourceOrgID string, eventType string) ([]string, error)
}

// ServiceRegistry manages federated service endpoints
type ServiceRegistry interface {
	// GetServiceEndpoints returns endpoints for specific services
	GetServiceEndpoints(ctx context.Context, serviceType string) ([]Recipient, error)

	// GetSubscribedServices returns services subscribed to specific event types
	GetSubscribedServices(ctx context.Context, eventType string) ([]Recipient, error)
}

// NewDefaultEventRouter creates a new default event router
func NewDefaultEventRouter(orgRouter OrganizationRouter, serviceRegistry ServiceRegistry) *DefaultEventRouter {
	return &DefaultEventRouter{
		rules:              make(map[string]RoutingRule),
		organizationRouter: orgRouter,
		serviceRegistry:    serviceRegistry,
	}
}

// GetRecipients returns the list of recipients for an event
func (r *DefaultEventRouter) GetRecipients(ctx context.Context, event *outbox.OutboxEvent) ([]Recipient, error) {
	var allRecipients []Recipient

	// 1. Check explicit routing rules
	ruleKey := fmt.Sprintf("%s:%s", event.EventType, event.AggregateType)
	if rule, exists := r.rules[ruleKey]; exists {
		recipients, err := r.evaluateRule(ctx, event, rule)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate routing rule: %w", err)
		}
		allRecipients = append(allRecipients, recipients...)
	}

	// 2. Apply default routing based on event type
	defaultRecipients, err := r.getDefaultRecipients(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("failed to get default recipients: %w", err)
	}
	allRecipients = append(allRecipients, defaultRecipients...)

	// 3. Deduplicate recipients
	return r.deduplicateRecipients(allRecipients), nil
}

// evaluateRule evaluates a routing rule and returns matching recipients
func (r *DefaultEventRouter) evaluateRule(ctx context.Context, event *outbox.OutboxEvent, rule RoutingRule) ([]Recipient, error) {
	// Check if conditions are met
	if !r.evaluateConditions(event, rule.Conditions) {
		return nil, nil
	}

	var recipients []Recipient

	for _, recipient := range rule.Recipients {
		switch recipient.Type {
		case RecipientTypeOrganization:
			// Get organization-specific endpoints
			orgRecipients, err := r.organizationRouter.GetOrganizationEndpoints(ctx, recipient.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get organization endpoints: %w", err)
			}
			recipients = append(recipients, orgRecipients...)

		case RecipientTypeService:
			// Get service-specific endpoints
			serviceRecipients, err := r.serviceRegistry.GetServiceEndpoints(ctx, recipient.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get service endpoints: %w", err)
			}
			recipients = append(recipients, serviceRecipients...)

		case RecipientTypeAll:
			// Broadcast to all subscribed services
			allRecipients, err := r.serviceRegistry.GetSubscribedServices(ctx, event.EventType)
			if err != nil {
				return nil, fmt.Errorf("failed to get subscribed services: %w", err)
			}
			recipients = append(recipients, allRecipients...)

		default:
			// Direct recipient
			recipients = append(recipients, recipient)
		}
	}

	return recipients, nil
}

// getDefaultRecipients applies default routing logic
func (r *DefaultEventRouter) getDefaultRecipients(ctx context.Context, event *outbox.OutboxEvent) ([]Recipient, error) {
	var recipients []Recipient

	// Parse event data to extract organization information
	eventData, err := r.parseEventData(event)
	if err != nil {
		return nil, fmt.Errorf("failed to parse event data: %w", err)
	}

	switch event.EventType {
	case "OrderCreated", "OrderStatusUpdated":
		// Route to buyer and seller organizations
		if buyerOrgID, ok := eventData["buyer_organization_id"].(string); ok {
			orgRecipients, err := r.organizationRouter.GetOrganizationEndpoints(ctx, buyerOrgID)
			if err == nil {
				recipients = append(recipients, orgRecipients...)
			}
		}

		if sellerOrgID, ok := eventData["seller_organization_id"].(string); ok {
			orgRecipients, err := r.organizationRouter.GetOrganizationEndpoints(ctx, sellerOrgID)
			if err == nil {
				recipients = append(recipients, orgRecipients...)
			}
		}

		// Route to logistics and analytics services
		logisticsRecipients, _ := r.serviceRegistry.GetServiceEndpoints(ctx, "logistics")
		analyticsRecipients, _ := r.serviceRegistry.GetServiceEndpoints(ctx, "analytics")
		recipients = append(recipients, logisticsRecipients...)
		recipients = append(recipients, analyticsRecipients...)

	case "CatalogItemCreated", "CatalogItemUpdated":
		// Route to discovery and pricing services
		discoveryRecipients, _ := r.serviceRegistry.GetServiceEndpoints(ctx, "discovery")
		pricingRecipients, _ := r.serviceRegistry.GetServiceEndpoints(ctx, "pricing")
		recipients = append(recipients, discoveryRecipients...)
		recipients = append(recipients, pricingRecipients...)

		// Route to partner organizations based on visibility
		if orgID, ok := eventData["organization_id"].(string); ok {
			partnerOrgs, err := r.organizationRouter.GetPartnerOrganizations(ctx, orgID, event.EventType)
			if err == nil {
				for _, partnerOrgID := range partnerOrgs {
					orgRecipients, err := r.organizationRouter.GetOrganizationEndpoints(ctx, partnerOrgID)
					if err == nil {
						recipients = append(recipients, orgRecipients...)
					}
				}
			}
		}

	case "InventoryAdjusted":
		// Route to inventory and analytics services
		inventoryRecipients, _ := r.serviceRegistry.GetServiceEndpoints(ctx, "inventory")
		analyticsRecipients, _ := r.serviceRegistry.GetServiceEndpoints(ctx, "analytics")
		recipients = append(recipients, inventoryRecipients...)
		recipients = append(recipients, analyticsRecipients...)
	}

	return recipients, nil
}

// evaluateConditions checks if all conditions are met
func (r *DefaultEventRouter) evaluateConditions(event *outbox.OutboxEvent, conditions []Condition) bool {
	if len(conditions) == 0 {
		return true
	}

	eventData, err := r.parseEventData(event)
	if err != nil {
		return false
	}

	for _, condition := range conditions {
		if !r.evaluateCondition(eventData, condition) {
			return false
		}
	}

	return true
}

// evaluateCondition evaluates a single condition
func (r *DefaultEventRouter) evaluateCondition(eventData map[string]interface{}, condition Condition) bool {
	fieldValue, exists := eventData[condition.Field]
	if !exists {
		return false
	}

	switch condition.Operator {
	case "eq":
		return fieldValue == condition.Value
	case "ne":
		return fieldValue != condition.Value
	case "in":
		if values, ok := condition.Value.([]interface{}); ok {
			for _, v := range values {
				if fieldValue == v {
					return true
				}
			}
		}
		return false
	case "contains":
		if str, ok := fieldValue.(string); ok {
			if substr, ok := condition.Value.(string); ok {
				return contains(str, substr)
			}
		}
		return false
	default:
		return false
	}
}

// parseEventData parses the event data JSON
func (r *DefaultEventRouter) parseEventData(event *outbox.OutboxEvent) (map[string]interface{}, error) {
	var eventData map[string]interface{}
	if err := json.Unmarshal([]byte(event.EventData), &eventData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
	}
	return eventData, nil
}

// deduplicateRecipients removes duplicate recipients
func (r *DefaultEventRouter) deduplicateRecipients(recipients []Recipient) []Recipient {
	seen := make(map[string]bool)
	var unique []Recipient

	for _, recipient := range recipients {
		key := fmt.Sprintf("%s:%s:%s", recipient.Type, recipient.ID, recipient.QueueURL)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, recipient)
		}
	}

	return unique
}

// AddRoutingRule adds a new routing rule
func (r *DefaultEventRouter) AddRoutingRule(rule RoutingRule) error {
	key := fmt.Sprintf("%s:%s", rule.EventType, rule.AggregateType)
	r.rules[key] = rule
	return nil
}

// RemoveRoutingRule removes a routing rule
func (r *DefaultEventRouter) RemoveRoutingRule(eventType, aggregateType string) error {
	key := fmt.Sprintf("%s:%s", eventType, aggregateType)
	delete(r.rules, key)
	return nil
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
