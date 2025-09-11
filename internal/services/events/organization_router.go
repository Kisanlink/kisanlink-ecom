package events

import (
    "context"
    "fmt"
)

// DefaultOrganizationRouter implements organization-specific routing
type DefaultOrganizationRouter struct {
    // This could be backed by a database or configuration service
    organizationEndpoints map[string][]Recipient
    partnerRelationships  map[string][]string // orgID -> partner orgIDs
}

// NewDefaultOrganizationRouter creates a new organization router
func NewDefaultOrganizationRouter() *DefaultOrganizationRouter {
    return &DefaultOrganizationRouter{
        organizationEndpoints: make(map[string][]Recipient),
        partnerRelationships:  make(map[string][]string),
    }
}

// GetOrganizationEndpoints returns the endpoints for an organization
func (r *DefaultOrganizationRouter) GetOrganizationEndpoints(ctx context.Context, orgID string) ([]Recipient, error) {
    endpoints, exists := r.organizationEndpoints[orgID]
    if !exists {
        // Return default endpoints or fetch from configuration service
        return r.getDefaultEndpoints(orgID), nil
    }
    return endpoints, nil
}

// GetPartnerOrganizations returns partner organizations that should receive events
func (r *DefaultOrganizationRouter) GetPartnerOrganizations(ctx context.Context, sourceOrgID string, eventType string) ([]string, error) {
    partners, exists := r.partnerRelationships[sourceOrgID]
    if !exists {
        return []string{}, nil
    }

    // Filter partners based on event type and permissions
    var authorizedPartners []string
    for _, partnerID := range partners {
        if r.isAuthorizedForEvent(sourceOrgID, partnerID, eventType) {
            authorizedPartners = append(authorizedPartners, partnerID)
        }
    }

    return authorizedPartners, nil
}

// RegisterOrganizationEndpoint registers an endpoint for an organization
func (r *DefaultOrganizationRouter) RegisterOrganizationEndpoint(orgID string, recipient Recipient) {
    if r.organizationEndpoints[orgID] == nil {
        r.organizationEndpoints[orgID] = []Recipient{}
    }
    r.organizationEndpoints[orgID] = append(r.organizationEndpoints[orgID], recipient)
}

// AddPartnerRelationship adds a partner relationship
func (r *DefaultOrganizationRouter) AddPartnerRelationship(orgID, partnerID string) {
    if r.partnerRelationships[orgID] == nil {
        r.partnerRelationships[orgID] = []string{}
    }
    r.partnerRelationships[orgID] = append(r.partnerRelationships[orgID], partnerID)
}

// getDefaultEndpoints returns default endpoints for an organization
func (r *DefaultOrganizationRouter) getDefaultEndpoints(orgID string) []Recipient {
    // This could be configured based on organization type, region, etc.
    return []Recipient{
        {
            Type:     RecipientTypeOrganization,
            ID:       orgID,
            QueueURL: fmt.Sprintf("https://sqs.region.amazonaws.com/account/org-%s-events", orgID),
        },
    }
}

// isAuthorizedForEvent checks if a partner is authorized to receive specific events
func (r *DefaultOrganizationRouter) isAuthorizedForEvent(sourceOrgID, partnerID, eventType string) bool {
    // Implement authorization logic based on:
    // - Partnership agreements
    // - Event type permissions
    // - Data sharing policies

    switch eventType {
    case "CatalogItemCreated", "CatalogItemUpdated":
        // Only share catalog events with trusted partners
        return r.isTrustedPartner(sourceOrgID, partnerID)
    case "OrderCreated":
        // Share order events only if partner is involved in the transaction
        return true // Simplified - would check if partner is buyer/seller
    case "InventoryAdjusted":
        // Inventory events are typically private
        return false
    default:
        return false
    }
}

// isTrustedPartner checks if an organization is a trusted partner
func (r *DefaultOrganizationRouter) isTrustedPartner(orgID, partnerID string) bool {
    // This would typically check a database or configuration
    // For now, return true for demonstration
    return true
}
