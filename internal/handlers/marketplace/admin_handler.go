package marketplace

import (
	"fmt"
	"strconv"
	"time"

	"kisanlink-ecom/entities/models/marketplace"
	"kisanlink-ecom/internal/common"
	marketplaceService "kisanlink-ecom/internal/services/marketplace"
	"kisanlink-ecom/internal/utils"

	"github.com/gin-gonic/gin"
)

// AdminHandler handles marketplace admin operations
type AdminHandler struct {
	marketplaceServices *marketplaceService.MarketplaceServices
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(marketplaceServices *marketplaceService.MarketplaceServices) *AdminHandler {
	return &AdminHandler{
		marketplaceServices: marketplaceServices,
	}
}

// ForceCloseListingRequest represents the request to force close a listing
type ForceCloseListingRequest struct {
	Reason string `json:"reason" binding:"required" example:"Policy violation - fraudulent listing"`
}

// RemoveBidRequest represents the request to remove a bid
type RemoveBidRequest struct {
	Reason string `json:"reason" binding:"required" example:"Fraudulent bid detected"`
}

// MarketplaceStatsResponse represents marketplace statistics
type MarketplaceStatsResponse struct {
	TotalListings         int     `json:"total_listings" example:"1250"`
	ActiveListings        int     `json:"active_listings" example:"450"`
	ClosedListings        int     `json:"closed_listings" example:"600"`
	ExpiredListings       int     `json:"expired_listings" example:"200"`
	TotalBids             int     `json:"total_bids" example:"5670"`
	AverageBidsPerListing float64 `json:"average_bids_per_listing" example:"4.5"`
	TotalValue            string  `json:"total_value" example:"2500000.00"`
	AverageListingValue   string  `json:"average_listing_value" example:"2000.00"`
}

// GetAllListings retrieves all marketplace listings (admin operation)
// @Summary Get all marketplace listings (Admin)
// @Description Retrieve all marketplace listings with admin privileges - no visibility restrictions
// @Tags Admin
// @Accept json
// @Produce json
// @Param status query string false "Filter by status" Enums(ACTIVE,CLOSED,EXPIRED,CANCELLED)
// @Param seller_id query string false "Filter by seller ID"
// @Param organization_id query string false "Filter by organization ID"
// @Param product_id query string false "Filter by product ID"
// @Param visibility query string false "Filter by visibility" Enums(PRIVATE,PUBLIC,NETWORK,ORGANIZATION)
// @Param auction_type query string false "Filter by auction type" Enums(OPEN,CLOSED)
// @Param min_price query string false "Filter by minimum price"
// @Param max_price query string false "Filter by maximum price"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} common.Response{data=ListingListResponse} "Listings retrieved successfully"
// @Failure 400 {object} common.Response{error=common.ResponseError} "Invalid request parameters"
// @Failure 401 {object} common.Response{error=common.ResponseError} "Unauthorized"
// @Failure 403 {object} common.Response{error=common.ResponseError} "Insufficient privileges"
// @Failure 500 {object} common.Response{error=common.ResponseError} "Internal server error"
// @Router /admin/v1/marketplace/listings [get]
// @Security BearerAuth
func (h *AdminHandler) GetAllListings(c *gin.Context) {
	// Verify admin role
	if !h.isAdmin(c) {
		common.Forbidden(c, "FORBIDDEN", "Access denied", nil)
		return
	}

	// Parse pagination parameters
	pagination, err := h.parsePaginationParams(c)
	if err != nil {
		common.BadRequest(c, "INVALID_PAGINATION", "Invalid pagination parameters", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	// Parse filter parameters
	filter, err := h.parseListingFilters(c)
	if err != nil {
		common.BadRequest(c, "INVALID_FILTER", "Invalid filter parameters", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	// Get all listings (admin operation - no visibility restrictions)
	listings, total, err := h.marketplaceServices.GetListingService().GetAllListings(c.Request.Context(), filter, pagination)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	// Convert to response format
	listingResponses := make([]ListingResponse, len(listings))
	for i, listing := range listings {
		listingResponses[i] = h.convertToListingResponse(listing)
	}

	common.Success(c, gin.H{
		"listings": listingResponses,
	}, &common.ResponseMeta{
		Pagination: common.NewPaginationMeta(pagination.Page, pagination.Limit, total),
	})
}

// ForceCloseListing force closes a listing (admin operation)
// @Summary Force close a marketplace listing (Admin)
// @Description Force close a marketplace listing regardless of its current state
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path string true "Listing ID"
// @Param request body ForceCloseListingRequest true "Force close request"
// @Success 200 {object} common.Response{data=ListingResponse} "Listing force closed successfully"
// @Failure 400 {object} common.Response{error=common.ResponseError} "Invalid request data"
// @Failure 401 {object} common.Response{error=common.ResponseError} "Unauthorized"
// @Failure 403 {object} common.Response{error=common.ResponseError} "Insufficient privileges"
// @Failure 404 {object} common.Response{error=common.ResponseError} "Listing not found"
// @Failure 500 {object} common.Response{error=common.ResponseError} "Internal server error"
// @Router /admin/v1/marketplace/listings/{id}/force-close [post]
// @Security BearerAuth
func (h *AdminHandler) ForceCloseListing(c *gin.Context) {
	// Verify admin role
	if !h.isAdmin(c) {
		common.Forbidden(c, "FORBIDDEN", "Access denied", nil)
		return
	}

	listingID := c.Param("id")
	if listingID == "" {
		common.BadRequest(c, "INVALID_REQUEST", "Listing ID is required", nil)
		return
	}

	var req ForceCloseListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "VALIDATION_ERROR", "Invalid input data", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	// Extract admin user ID from context
	adminID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	// Force close listing
	listing, err := h.marketplaceServices.GetListingService().ForceCloseListing(c.Request.Context(), listingID, req.Reason, adminID.(string))
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	// Convert to response format
	response := h.convertToListingResponse(listing)
	common.Success(c, response, nil)
}

// RemoveBid removes a fraudulent bid (admin operation)
// @Summary Remove a bid (Admin)
// @Description Remove a fraudulent or policy-violating bid from the marketplace
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path string true "Bid ID"
// @Param request body RemoveBidRequest true "Remove bid request"
// @Success 200 {object} common.Response "Bid removed successfully"
// @Failure 400 {object} common.Response{error=common.ResponseError} "Invalid request data"
// @Failure 401 {object} common.Response{error=common.ResponseError} "Unauthorized"
// @Failure 403 {object} common.Response{error=common.ResponseError} "Insufficient privileges"
// @Failure 404 {object} common.Response{error=common.ResponseError} "Bid not found"
// @Failure 500 {object} common.Response{error=common.ResponseError} "Internal server error"
// @Router /admin/v1/marketplace/bids/{id} [delete]
// @Security BearerAuth
func (h *AdminHandler) RemoveBid(c *gin.Context) {
	// Verify admin role
	if !h.isAdmin(c) {
		common.Forbidden(c, "FORBIDDEN", "Access denied", nil)
		return
	}

	bidID := c.Param("id")
	if bidID == "" {
		common.BadRequest(c, "INVALID_REQUEST", "Bid ID is required", nil)
		return
	}

	var req RemoveBidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "VALIDATION_ERROR", "Invalid input data", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	// Extract admin user ID from context
	adminID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	// Remove bid
	err := h.marketplaceServices.GetBiddingService().RemoveBid(c.Request.Context(), bidID, req.Reason, adminID.(string))
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	common.Success(c, nil, nil)
}

// GetMarketplaceStats retrieves marketplace statistics (admin operation)
// @Summary Get marketplace statistics (Admin)
// @Description Retrieve comprehensive marketplace statistics and analytics
// @Tags Admin
// @Accept json
// @Produce json
// @Param time_range query string false "Time range for statistics" Enums(24h,7d,30d,90d,1y) default(30d)
// @Success 200 {object} common.Response{data=MarketplaceStatsResponse} "Statistics retrieved successfully"
// @Failure 401 {object} common.Response{error=common.ResponseError} "Unauthorized"
// @Failure 403 {object} common.Response{error=common.ResponseError} "Insufficient privileges"
// @Failure 500 {object} common.Response{error=common.ResponseError} "Internal server error"
// @Router /admin/v1/marketplace/stats [get]
// @Security BearerAuth
func (h *AdminHandler) GetMarketplaceStats(c *gin.Context) {
	// Verify admin role
	if !h.isAdmin(c) {
		common.Forbidden(c, "FORBIDDEN", "Access denied", nil)
		return
	}

	timeRange := c.Query("time_range")
	if timeRange == "" {
		timeRange = "30d"
	}

	// Get marketplace statistics
	stats, err := h.marketplaceServices.GetAnalyticsService().GetMarketplaceStats(c.Request.Context(), timeRange)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	// Convert to response format
	response := MarketplaceStatsResponse{
		TotalListings:         stats.TotalListings,
		ActiveListings:        stats.ActiveListings,
		ClosedListings:        stats.ClosedListings,
		ExpiredListings:       stats.ExpiredListings,
		TotalBids:             stats.TotalBids,
		AverageBidsPerListing: stats.AverageBidsPerListing,
		TotalValue:            stats.TotalValue.String(),
		AverageListingValue:   stats.AverageListingValue.String(),
	}

	common.Success(c, response, nil)
}

// GetAuditLog retrieves marketplace audit log (admin operation)
// @Summary Get marketplace audit log (Admin)
// @Description Retrieve audit log of marketplace activities for compliance and monitoring
// @Tags Admin
// @Accept json
// @Produce json
// @Param listing_id query string false "Filter by listing ID"
// @Param event_type query string false "Filter by event type" Enums(LISTING_CREATED,BID_PLACED,BID_OUTBID,LISTING_CLOSED,LISTING_EXPIRED)
// @Param actor_id query string false "Filter by actor ID"
// @Param time_from query string false "Filter events from this time (RFC3339 format)"
// @Param time_to query string false "Filter events to this time (RFC3339 format)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Success 200 {object} common.Response{data=AuditLogResponse} "Audit log retrieved successfully"
// @Failure 400 {object} common.Response{error=common.ResponseError} "Invalid request parameters"
// @Failure 401 {object} common.Response{error=common.ResponseError} "Unauthorized"
// @Failure 403 {object} common.Response{error=common.ResponseError} "Insufficient privileges"
// @Failure 500 {object} common.Response{error=common.ResponseError} "Internal server error"
// @Router /admin/v1/marketplace/audit-log [get]
// @Security BearerAuth
func (h *AdminHandler) GetAuditLog(c *gin.Context) {
	// Verify admin role
	if !h.isAdmin(c) {
		common.Forbidden(c, "FORBIDDEN", "Access denied", nil)
		return
	}

	// Parse pagination parameters
	pagination, err := h.parsePaginationParams(c)
	if err != nil {
		common.BadRequest(c, "INVALID_PAGINATION", "Invalid pagination parameters", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	// Parse filter parameters
	filter, err := h.parseEventFilters(c)
	if err != nil {
		common.BadRequest(c, "INVALID_FILTER", "Invalid filter parameters", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	// Get audit log
	events, total, err := h.marketplaceServices.GetAuditService().GetAuditLog(c.Request.Context(), filter, pagination)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	// Convert to response format
	eventResponses := make([]AuditEventResponse, len(events))
	for i, event := range events {
		eventResponses[i] = h.convertToAuditEventResponse(event)
	}

	common.Success(c, gin.H{
		"events": eventResponses,
	}, &common.ResponseMeta{
		Pagination: common.NewPaginationMeta(pagination.Page, pagination.Limit, total),
	})
}

// Helper methods

// isAdmin checks if the current user has admin privileges
func (h *AdminHandler) isAdmin(c *gin.Context) bool {
	// Extract user role from context (set by auth middleware)
	userRole, exists := c.Get("user_role")
	if !exists {
		return false
	}

	// Check if user has admin role
	return userRole == "admin" || userRole == "super_admin"
}

// convertToListingResponse converts a listing model to response format
func (h *AdminHandler) convertToListingResponse(listing *marketplace.Listing) ListingResponse {
	response := ListingResponse{
		ID:                  listing.ID,
		ListingID:           listing.ListingID,
		ProductID:           listing.ProductID,
		SellerID:            listing.SellerID,
		OrganizationID:      listing.OrganizationID,
		Quantity:            listing.Quantity.String(),
		AskingPrice:         listing.AskingPrice.String(),
		MinimumBid:          listing.MinimumBid.String(),
		Currency:            listing.Currency,
		ListingDuration:     listing.ListingDuration,
		ExpiresAt:           listing.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		Status:              listing.Status,
		CurrentHighestBidID: listing.CurrentHighestBidID,
		BidCount:            listing.BidCount,
		Visibility:          listing.Visibility,
		AuctionType:         listing.AuctionType,
		BidVisibility:       listing.BidVisibility,
		TermsConditions:     listing.TermsConditions,
		TimeRemaining:       h.formatTimeRemaining(listing.GetTimeRemaining()),
		CreatedAt:           listing.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:           listing.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Set optional fields
	if listing.ClosedAt != nil {
		closedAt := listing.ClosedAt.Format("2006-01-02T15:04:05Z")
		response.ClosedAt = &closedAt
	}

	if listing.CloseReason != nil {
		response.CloseReason = listing.CloseReason
	}

	// Parse pickup location
	if location, err := listing.GetPickupLocation(); err == nil && location != nil {
		response.PickupLocation = location
	}

	return response
}

// convertToAuditEventResponse converts an audit event to response format
func (h *AdminHandler) convertToAuditEventResponse(event *marketplace.AuctionEvent) AuditEventResponse {
	return AuditEventResponse{
		ID:        event.ID,
		EventID:   event.EventID,
		ListingID: event.ListingID,
		EventType: event.EventType,
		ActorID:   event.ActorID,
		ActorType: event.ActorType,
		Timestamp: event.Timestamp.Format("2006-01-02T15:04:05Z"),
		Summary:   h.generateEventSummary(event),
	}
}

// generateEventSummary generates a human-readable summary for an audit event
func (h *AdminHandler) generateEventSummary(event *marketplace.AuctionEvent) string {
	switch event.EventType {
	case marketplace.EventListingCreated:
		return fmt.Sprintf("Listing %s created by %s", event.ListingID, event.ActorID)
	case marketplace.EventBidPlaced:
		return fmt.Sprintf("Bid placed on listing %s by %s", event.ListingID, event.ActorID)
	case marketplace.EventBidOutbid:
		return fmt.Sprintf("Bid outbid on listing %s", event.ListingID)
	case marketplace.EventListingClosed:
		return fmt.Sprintf("Listing %s closed by %s", event.ListingID, event.ActorID)
	case marketplace.EventListingExpired:
		return fmt.Sprintf("Listing %s expired", event.ListingID)
	default:
		return fmt.Sprintf("Event %s on listing %s", event.EventType, event.ListingID)
	}
}

// parsePaginationParams parses pagination parameters from query string
func (h *AdminHandler) parsePaginationParams(c *gin.Context) (*common.PaginationParams, error) {
	page := 1
	limit := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	return &common.PaginationParams{
		Page:  page,
		Limit: limit,
	}, nil
}

// parseListingFilters parses filter parameters from query string
func (h *AdminHandler) parseListingFilters(c *gin.Context) (*marketplace.ListingFilter, error) {
	filter := &marketplace.ListingFilter{}

	// Status filter
	if statusStr := c.Query("status"); statusStr != "" {
		status := marketplace.ListingStatus(statusStr)
		filter.Status = &status
	}

	// Seller ID filter
	if sellerID := c.Query("seller_id"); sellerID != "" {
		filter.SellerID = sellerID
	}

	// Organization ID filter
	if organizationID := c.Query("organization_id"); organizationID != "" {
		filter.OrganizationID = organizationID
	}

	// Product ID filter
	if productID := c.Query("product_id"); productID != "" {
		filter.ProductID = productID
	}

	// Visibility filter
	if visibilityStr := c.Query("visibility"); visibilityStr != "" {
		visibility := marketplace.ListingVisibility(visibilityStr)
		filter.Visibility = &visibility
	}

	// Auction type filter
	if auctionTypeStr := c.Query("auction_type"); auctionTypeStr != "" {
		auctionType := marketplace.AuctionType(auctionTypeStr)
		filter.AuctionType = &auctionType
	}

	return filter, nil
}

// parseEventFilters parses event filter parameters from query string
func (h *AdminHandler) parseEventFilters(c *gin.Context) (*marketplace.EventFilter, error) {
	filter := &marketplace.EventFilter{}

	// Listing ID filter
	if listingID := c.Query("listing_id"); listingID != "" {
		filter.ListingID = listingID
	}

	// Event type filter
	if eventTypeStr := c.Query("event_type"); eventTypeStr != "" {
		eventType := marketplace.AuctionEventType(eventTypeStr)
		filter.EventType = &eventType
	}

	// Actor ID filter
	if actorID := c.Query("actor_id"); actorID != "" {
		filter.ActorID = actorID
	}

	// Time range filters would be parsed here
	// Implementation depends on the specific time format requirements

	return filter, nil
}

// formatTimeRemaining formats time duration to human-readable string
func (h *AdminHandler) formatTimeRemaining(duration time.Duration) string {
	if duration <= 0 {
		return "Expired"
	}

	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60

	if hours > 24 {
		days := hours / 24
		remainingHours := hours % 24
		if remainingHours > 0 {
			return fmt.Sprintf("%dd %dh", days, remainingHours)
		}
		return fmt.Sprintf("%dd", days)
	}

	if hours > 0 {
		if minutes > 0 {
			return fmt.Sprintf("%dh %dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}

	return fmt.Sprintf("%dm", minutes)
}

// Response types for admin operations

// AuditLogResponse represents the response for audit log operations
type AuditLogResponse struct {
	Events     []AuditEventResponse  `json:"events"`
	Pagination common.PaginationMeta `json:"pagination"`
}

// AuditEventResponse represents an audit event in the response
type AuditEventResponse struct {
	ID        string                       `json:"id" example:"uuid-123"`
	EventID   string                       `json:"event_id" example:"EVT_1234567890"`
	ListingID string                       `json:"listing_id" example:"LST_1234567890"`
	EventType marketplace.AuctionEventType `json:"event_type" example:"BID_PLACED"`
	ActorID   string                       `json:"actor_id" example:"USER_456"`
	ActorType marketplace.ActorType        `json:"actor_type" example:"USER"`
	Timestamp string                       `json:"timestamp" example:"2024-01-14T10:30:00Z"`
	Summary   string                       `json:"summary" example:"Bid placed on listing LST_1234567890 by USER_456"`
}
