package middleware

import (
	"fmt"
	"net/http"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/common"
	marketplaceModels "github.com/Kisanlink/kisanlink-ecom/entities/models/marketplace"
	"github.com/Kisanlink/kisanlink-ecom/internal/auth"
	"github.com/Kisanlink/kisanlink-ecom/internal/repositories/marketplace"

	"github.com/gin-gonic/gin"
)

// Marketplace-specific resource and action constants
const (
	// Marketplace Resources
	ResourceMarketplaceListing = "marketplace_listing"
	ResourceMarketplaceBid     = "marketplace_bid"
	ResourceMarketplaceAuction = "marketplace_auction"
	ResourceMarketplaceAdmin   = "marketplace_admin"

	// Marketplace Actions
	ActionCreateListing     = "create_listing"
	ActionViewListing       = "view_listing"
	ActionUpdateListing     = "update_listing"
	ActionDeleteListing     = "delete_listing"
	ActionCloseListing      = "close_listing"
	ActionPlaceBid          = "place_bid"
	ActionViewBid           = "view_bid"
	ActionViewBidHistory    = "view_bid_history"
	ActionManageAuction     = "manage_auction"
	ActionAdminOverride     = "admin_override"
	ActionForceCloseListing = "force_close_listing"
	ActionRemoveBid         = "remove_bid"
	ActionViewAnalytics     = "view_analytics"

	// Marketplace Roles
	RoleMarketplaceAdmin  = "marketplace_admin"
	RoleMarketplaceLister = "marketplace_lister"
	RoleMarketplaceBuyer  = "marketplace_buyer"
	RoleMarketplaceViewer = "marketplace_viewer"
)

// MarketplaceAuthMiddleware provides marketplace-specific authorization
type MarketplaceAuthMiddleware struct {
	aaaClient   auth.Client
	listingRepo marketplace.ListingRepository
	bidRepo     marketplace.BidRepository
}

// NewMarketplaceAuthMiddleware creates a new marketplace authorization middleware
func NewMarketplaceAuthMiddleware(
	aaaClient auth.Client,
	listingRepo marketplace.ListingRepository,
	bidRepo marketplace.BidRepository,
) *MarketplaceAuthMiddleware {
	return &MarketplaceAuthMiddleware{
		aaaClient:   aaaClient,
		listingRepo: listingRepo,
		bidRepo:     bidRepo,
	}
}

// RequireMarketplaceAuth ensures user is authenticated and has marketplace access
func (m *MarketplaceAuthMiddleware) RequireMarketplaceAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userContext, exists := GetUserContext(c)
		if !exists {
			m.handleUnauthorized(c, "Authentication required for marketplace access")
			return
		}

		// Validate organization ID is present for marketplace operations
		orgID := GetOrganizationID(c)
		if orgID == "" {
			m.handleUnauthorized(c, "Organization ID required for marketplace operations")
			return
		}

		// Validate user belongs to the organization
		if userContext.OrganizationID != orgID {
			req := &auth.AuthorizeRequest{
				UserID:     userContext.UserID,
				TenantID:   userContext.TenantID,
				Resource:   "organization",
				Action:     "access",
				ResourceID: orgID,
			}
			resp, err := m.aaaClient.Authorize(c.Request.Context(), req)
			if err != nil {
				m.handleAuthError(c, "Organization validation failed", err)
				return
			}
			allowed := resp.Allowed
			if !allowed {
				m.handleForbidden(c, "Access denied to organization marketplace")
				return
			}
		}

		// Store validated organization ID in context
		c.Set("validated_org_id", orgID)
		c.Next()
	}
}

// RequireListingPermission validates permissions for listing operations
func (m *MarketplaceAuthMiddleware) RequireListingPermission(action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userContext, exists := GetUserContext(c)
		if !exists {
			m.handleUnauthorized(c, "Authentication required")
			return
		}

		// Check permission with AAA service
		req := &auth.AuthorizeRequest{
			UserID:   userContext.UserID,
			TenantID: userContext.TenantID,
			Resource: ResourceMarketplaceListing,
			Action:   action,
		}
		resp, err := m.aaaClient.Authorize(c.Request.Context(), req)
		if err != nil {
			m.handleAuthError(c, "Permission evaluation failed", err)
			return
		}
		allowed := resp.Allowed

		if !allowed {
			m.handleForbidden(c, fmt.Sprintf("Permission denied for action '%s' on marketplace listings", action))
			return
		}

		c.Next()
	}
}

// RequireBiddingPermission validates permissions for bidding operations
func (m *MarketplaceAuthMiddleware) RequireBiddingPermission(action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userContext, exists := GetUserContext(c)
		if !exists {
			m.handleUnauthorized(c, "Authentication required")
			return
		}

		// Check permission with AAA service
		req := &auth.AuthorizeRequest{
			UserID:   userContext.UserID,
			TenantID: userContext.TenantID,
			Resource: ResourceMarketplaceBid,
			Action:   action,
		}
		resp, err := m.aaaClient.Authorize(c.Request.Context(), req)
		if err != nil {
			m.handleAuthError(c, "Permission evaluation failed", err)
			return
		}
		allowed := resp.Allowed

		if !allowed {
			m.handleForbidden(c, fmt.Sprintf("Permission denied for action '%s' on marketplace bids", action))
			return
		}

		c.Next()
	}
}

// RequireAdminPermission validates admin permissions for marketplace operations
func (m *MarketplaceAuthMiddleware) RequireAdminPermission(action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userContext, exists := GetUserContext(c)
		if !exists {
			m.handleUnauthorized(c, "Authentication required")
			return
		}

		// Check if user has admin role
		hasAdminRole := false
		for _, role := range userContext.Roles {
			if role == RoleMarketplaceAdmin || role == "admin" || role == "super_admin" {
				hasAdminRole = true
				break
			}
		}

		if !hasAdminRole {
			m.handleForbidden(c, "Admin role required for this operation")
			return
		}

		// Check specific admin permission
		req := &auth.AuthorizeRequest{
			UserID:   userContext.UserID,
			TenantID: userContext.TenantID,
			Resource: ResourceMarketplaceAdmin,
			Action:   action,
		}
		resp, err := m.aaaClient.Authorize(c.Request.Context(), req)
		if err != nil {
			m.handleAuthError(c, "Admin permission evaluation failed", err)
			return
		}
		allowed := resp.Allowed

		if !allowed {
			m.handleForbidden(c, fmt.Sprintf("Admin permission denied for action '%s'", action))
			return
		}

		c.Next()
	}
}

// RequireListingOwnership validates that user owns the listing or has admin privileges
func (m *MarketplaceAuthMiddleware) RequireListingOwnership() gin.HandlerFunc {
	return func(c *gin.Context) {
		userContext, exists := GetUserContext(c)
		if !exists {
			m.handleUnauthorized(c, "Authentication required")
			return
		}

		// Get listing ID from URL parameter
		listingID := c.Param("id")
		if listingID == "" {
			m.handleBadRequest(c, "Listing ID required")
			return
		}

		// Check if user has admin role (admins bypass ownership check)
		hasAdminRole := false
		for _, role := range userContext.Roles {
			if role == RoleMarketplaceAdmin || role == "admin" || role == "super_admin" {
				hasAdminRole = true
				break
			}
		}

		if hasAdminRole {
			c.Set("is_admin_override", true)
			c.Next()
			return
		}

		// Query the listing to verify ownership
		listing, err := m.listingRepo.GetByID(c.Request.Context(), listingID)
		if err != nil {
			m.handleAuthError(c, "Failed to retrieve listing", err)
			return
		}

		// Verify user is the seller
		if listing.SellerID != userContext.UserID {
			m.handleForbidden(c, "You do not have permission to modify this listing")
			return
		}

		// Store ownership validation in context
		c.Set("listing_ownership_validated", true)
		c.Set("validated_listing", listing)
		c.Next()
	}
}

// RequireBidOwnership validates that user owns the bid or has admin privileges
func (m *MarketplaceAuthMiddleware) RequireBidOwnership() gin.HandlerFunc {
	return func(c *gin.Context) {
		userContext, exists := GetUserContext(c)
		if !exists {
			m.handleUnauthorized(c, "Authentication required")
			return
		}

		// Get bid ID from URL parameter
		bidID := c.Param("id")
		if bidID == "" {
			m.handleBadRequest(c, "Bid ID required")
			return
		}

		// Check if user has admin role (admins bypass ownership check)
		hasAdminRole := false
		for _, role := range userContext.Roles {
			if role == RoleMarketplaceAdmin || role == "admin" || role == "super_admin" {
				hasAdminRole = true
				break
			}
		}

		if hasAdminRole {
			c.Set("is_admin_override", true)
			c.Next()
			return
		}

		// Query the bid to verify ownership
		bid, err := m.bidRepo.GetByID(c.Request.Context(), bidID)
		if err != nil {
			m.handleAuthError(c, "Failed to retrieve bid", err)
			return
		}

		// Verify user is the bidder
		if bid.BidderID != userContext.UserID {
			m.handleForbidden(c, "You do not have permission to modify this bid")
			return
		}

		// Store ownership validation in context
		c.Set("bid_ownership_validated", true)
		c.Set("validated_bid", bid)
		c.Next()
	}
}

// ValidateListingVisibility validates user can access listing based on visibility settings
func (m *MarketplaceAuthMiddleware) ValidateListingVisibility() gin.HandlerFunc {
	return func(c *gin.Context) {
		userContext, exists := GetUserContext(c)
		if !exists {
			m.handleUnauthorized(c, "Authentication required")
			return
		}

		// Get listing ID from URL parameter
		listingID := c.Param("id")
		if listingID == "" {
			m.handleBadRequest(c, "Listing ID required")
			return
		}

		// Query the listing to check visibility
		listing, err := m.listingRepo.GetByID(c.Request.Context(), listingID)
		if err != nil {
			m.handleAuthError(c, "Failed to retrieve listing", err)
			return
		}

		// Check if user has admin role (admins can access all listings)
		hasAdminRole := false
		for _, role := range userContext.Roles {
			if role == RoleMarketplaceAdmin || role == "admin" || role == "super_admin" {
				hasAdminRole = true
				break
			}
		}

		if hasAdminRole {
			c.Set("is_admin_override", true)
			c.Set("validated_listing", listing)
			c.Next()
			return
		}

		// Validate access based on visibility settings
		switch listing.Visibility {
		case marketplaceModels.VisibilityPublic:
			// PUBLIC: All authenticated users can access
			c.Set("validated_listing", listing)
			c.Next()
			return

		case marketplaceModels.VisibilityOrganization:
			// ORGANIZATION: Only same organization
			if userContext.OrganizationID != listing.OrganizationID {
				m.handleForbidden(c, "This listing is only visible to members of the listing organization")
				return
			}

		case marketplaceModels.VisibilityNetwork:
			// NETWORK: Partner organizations (requires AAA check for organization relationships)
			if userContext.OrganizationID == listing.OrganizationID {
				// Same org always has access
				c.Set("validated_listing", listing)
				c.Next()
				return
			}

			// Check if user's organization is a partner with the listing organization
			req := &auth.AuthorizeRequest{
				UserID:     userContext.UserID,
				TenantID:   userContext.TenantID,
				Resource:   "organization",
				Action:     "access_network",
				ResourceID: listing.OrganizationID,
			}
			resp, err := m.aaaClient.Authorize(c.Request.Context(), req)
			if err != nil {
				m.handleAuthError(c, "Network access validation failed", err)
				return
			}
			if !resp.Allowed {
				m.handleForbidden(c, "This listing is only visible to network partner organizations")
				return
			}

		case marketplaceModels.VisibilityPrivate:
			// PRIVATE: Only invited participants (seller + invited bidders)
			// Check if user is the seller
			if listing.SellerID == userContext.UserID {
				c.Set("validated_listing", listing)
				c.Next()
				return
			}

			// Check if user has placed a bid (indicating they were invited)
			bids, _, err := m.bidRepo.GetUserBids(c.Request.Context(), userContext.UserID, &marketplaceModels.BidFilter{
				ListingID: listingID,
			}, nil)
			if err != nil {
				m.handleAuthError(c, "Failed to check bid history", err)
				return
			}

			if len(bids) == 0 {
				m.handleForbidden(c, "This is a private listing and you have not been invited")
				return
			}

		default:
			m.handleAuthError(c, "Invalid listing visibility setting", fmt.Errorf("unknown visibility: %s", listing.Visibility))
			return
		}

		// Store validated listing in context
		c.Set("visibility_validated", true)
		c.Set("validated_listing", listing)
		c.Next()
	}
}

// RequireMarketplaceRole validates user has specific marketplace role
func (m *MarketplaceAuthMiddleware) RequireMarketplaceRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userContext, exists := GetUserContext(c)
		if !exists {
			m.handleUnauthorized(c, "Authentication required")
			return
		}

		// Check if user has any of the required marketplace roles
		userRoles := make(map[string]bool)
		for _, role := range userContext.Roles {
			userRoles[role] = true
		}

		for _, requiredRole := range requiredRoles {
			if userRoles[requiredRole] {
				c.Set("marketplace_role", requiredRole)
				c.Next()
				return
			}
		}

		m.handleForbidden(c, fmt.Sprintf("Marketplace role required. Required roles: %v", requiredRoles))
	}
}

// ValidateOrganizationAccess validates organization-level access for marketplace operations
func (m *MarketplaceAuthMiddleware) ValidateOrganizationAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		userContext, exists := GetUserContext(c)
		if !exists {
			m.handleUnauthorized(c, "Authentication required")
			return
		}

		// Get organization ID from parameter or query first (for explicit cross-org access)
		orgID := c.Param("org_id")
		if orgID == "" {
			orgID = c.Query("org_id")
		}
		// If not explicitly provided, use from context (user's own org from token)
		if orgID == "" {
			orgID = GetOrganizationID(c)
		}

		if orgID == "" {
			m.handleBadRequest(c, "Organization ID required")
			return
		}

		// Validate user has access to this organization
		if userContext.OrganizationID != orgID {
			req := &auth.AuthorizeRequest{
				UserID:     userContext.UserID,
				TenantID:   userContext.TenantID,
				Resource:   "organization",
				Action:     "access",
				ResourceID: orgID,
			}
			resp, err := m.aaaClient.Authorize(c.Request.Context(), req)
			if err != nil {
				m.handleAuthError(c, "Organization access validation failed", err)
				return
			}
			allowed := resp.Allowed
			if !allowed {
				m.handleForbidden(c, fmt.Sprintf("Access denied to organization %s", orgID))
				return
			}
		}

		c.Set("validated_org_id", orgID)
		c.Next()
	}
}

// PreventSelfBidding prevents users from bidding on their own listings
func (m *MarketplaceAuthMiddleware) PreventSelfBidding() gin.HandlerFunc {
	return func(c *gin.Context) {
		userContext, exists := GetUserContext(c)
		if !exists {
			m.handleUnauthorized(c, "Authentication required")
			return
		}

		// Get listing ID from URL parameter or request body
		listingID := c.Param("id")
		if listingID == "" {
			// For POST requests, listing_id might be in the request body
			listingID = c.Param("listing_id")
		}
		if listingID == "" {
			m.handleBadRequest(c, "Listing ID required")
			return
		}

		// Query the listing to check ownership
		listing, err := m.listingRepo.GetByID(c.Request.Context(), listingID)
		if err != nil {
			m.handleAuthError(c, "Failed to retrieve listing", err)
			return
		}

		// Prevent self-bidding: user cannot bid on their own listing
		if listing.SellerID == userContext.UserID {
			m.handleForbidden(c, "You cannot place bids on your own listings")
			return
		}

		// Store validated listing in context for use by handler
		c.Set("self_bidding_validated", true)
		c.Set("validated_listing", listing)
		c.Next()
	}
}

// Helper methods for error handling

func (m *MarketplaceAuthMiddleware) handleUnauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, common.APIResponse{
		Success: false,
		Error: &common.APIError{
			Code:    "UNAUTHORIZED",
			Message: message,
		},
	})
}

func (m *MarketplaceAuthMiddleware) handleForbidden(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusForbidden, common.APIResponse{
		Success: false,
		Error: &common.APIError{
			Code:    "FORBIDDEN",
			Message: message,
		},
	})
}

func (m *MarketplaceAuthMiddleware) handleBadRequest(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, common.APIResponse{
		Success: false,
		Error: &common.APIError{
			Code:    "BAD_REQUEST",
			Message: message,
		},
	})
}

func (m *MarketplaceAuthMiddleware) handleAuthError(c *gin.Context, message string, err error) {
	details := ""
	if err != nil {
		details = err.Error()
	}

	c.AbortWithStatusJSON(http.StatusInternalServerError, common.APIResponse{
		Success: false,
		Error: &common.APIError{
			Code:    "AUTHORIZATION_ERROR",
			Message: message,
			Details: details,
		},
	})
}

// Utility functions for extracting context values

// GetValidatedOrgID retrieves the validated organization ID from context
func GetValidatedOrgID(c *gin.Context) (string, bool) {
	orgID, exists := c.Get("validated_org_id")
	if !exists {
		return "", false
	}
	if id, ok := orgID.(string); ok {
		return id, true
	}
	return "", false
}

// GetMarketplaceRole retrieves the marketplace role from context
func GetMarketplaceRole(c *gin.Context) (string, bool) {
	role, exists := c.Get("marketplace_role")
	if !exists {
		return "", false
	}
	if r, ok := role.(string); ok {
		return r, true
	}
	return "", false
}

// IsListingOwnershipRequired checks if listing ownership validation is required
func IsListingOwnershipRequired(c *gin.Context) (bool, string) {
	required, exists := c.Get("require_listing_ownership")
	if !exists {
		return false, ""
	}
	if req, ok := required.(bool); ok && req {
		if listingID, exists := c.Get("listing_id_for_ownership"); exists {
			if id, ok := listingID.(string); ok {
				return true, id
			}
		}
	}
	return false, ""
}

// IsBidOwnershipRequired checks if bid ownership validation is required
func IsBidOwnershipRequired(c *gin.Context) (bool, string) {
	required, exists := c.Get("require_bid_ownership")
	if !exists {
		return false, ""
	}
	if req, ok := required.(bool); ok && req {
		if bidID, exists := c.Get("bid_id_for_ownership"); exists {
			if id, ok := bidID.(string); ok {
				return true, id
			}
		}
	}
	return false, ""
}

// IsVisibilityCheckRequired checks if listing visibility validation is required
func IsVisibilityCheckRequired(c *gin.Context) (bool, string) {
	required, exists := c.Get("require_visibility_check")
	if !exists {
		return false, ""
	}
	if req, ok := required.(bool); ok && req {
		if listingID, exists := c.Get("listing_id_for_visibility"); exists {
			if id, ok := listingID.(string); ok {
				return true, id
			}
		}
	}
	return false, ""
}

// IsSelfBiddingPrevented checks if self-bidding prevention is required
func IsSelfBiddingPrevented(c *gin.Context) (bool, string) {
	prevented, exists := c.Get("prevent_self_bidding")
	if !exists {
		return false, ""
	}
	if prev, ok := prevented.(bool); ok && prev {
		if listingID, exists := c.Get("listing_id_for_self_bid_check"); exists {
			if id, ok := listingID.(string); ok {
				return true, id
			}
		}
	}
	return false, ""
}
