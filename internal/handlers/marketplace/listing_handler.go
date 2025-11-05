package marketplace

import (
	"fmt"
	"strconv"
	"time"

	"kisanlink-ecom/entities/models/marketplace"
	"kisanlink-ecom/internal/common"
	"kisanlink-ecom/internal/middleware"
	marketplaceService "kisanlink-ecom/internal/services/marketplace"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// ListingHandler handles marketplace listing operations
type ListingHandler struct {
	marketplaceServices *marketplaceService.MarketplaceServices
}

// NewListingHandler creates a new listing handler
func NewListingHandler(marketplaceServices *marketplaceService.MarketplaceServices) *ListingHandler {
	return &ListingHandler{
		marketplaceServices: marketplaceServices,
	}
}

// CreateListingRequest represents the request to create a new listing
type CreateListingRequest struct {
	ProductID       string                        `json:"product_id" binding:"required" example:"PROD_123"`
	Quantity        string                        `json:"quantity" binding:"required" example:"100.5"`
	AskingPrice     string                        `json:"asking_price" binding:"required" example:"1500.00"`
	MinimumBid      string                        `json:"minimum_bid" binding:"required" example:"1000.00"`
	Currency        string                        `json:"currency" example:"INR"`
	ListingDuration int                           `json:"listing_duration_hours" binding:"required,min=1,max=168" example:"24"`
	Visibility      marketplace.ListingVisibility `json:"visibility" example:"PUBLIC"`
	AuctionType     marketplace.AuctionType       `json:"auction_type" example:"OPEN"`
	BidVisibility   marketplace.BidVisibility     `json:"bid_visibility" example:"FULL"`
	PickupLocation  *marketplace.Location         `json:"pickup_location,omitempty"`
	TermsConditions string                        `json:"terms_conditions,omitempty" example:"Pickup within 7 days"`
}

// UpdateListingRequest represents the request to update an existing listing
type UpdateListingRequest struct {
	AskingPrice     *string                        `json:"asking_price,omitempty" example:"1600.00"`
	MinimumBid      *string                        `json:"minimum_bid,omitempty" example:"1100.00"`
	Visibility      *marketplace.ListingVisibility `json:"visibility,omitempty" example:"ORGANIZATION"`
	BidVisibility   *marketplace.BidVisibility     `json:"bid_visibility,omitempty" example:"PARTIAL"`
	PickupLocation  *marketplace.Location          `json:"pickup_location,omitempty"`
	TermsConditions *string                        `json:"terms_conditions,omitempty" example:"Updated terms"`
}

// CloseListingRequest represents the request to close a listing
type CloseListingRequest struct {
	Reason string `json:"reason" binding:"required" example:"Sold through other channel"`
}

// ListingResponse represents the response for listing operations
type ListingResponse struct {
	ID                  string                        `json:"id" example:"uuid-123"`
	ListingID           string                        `json:"listing_id" example:"LST_1234567890"`
	ProductID           string                        `json:"product_id" example:"PROD_123"`
	SellerID            string                        `json:"seller_id" example:"USER_456"`
	OrganizationID      string                        `json:"organization_id" example:"ORG_789"`
	Quantity            string                        `json:"quantity" example:"100.5"`
	AskingPrice         string                        `json:"asking_price" example:"1500.00"`
	MinimumBid          string                        `json:"minimum_bid" example:"1000.00"`
	Currency            string                        `json:"currency" example:"INR"`
	ListingDuration     int                           `json:"listing_duration_hours" example:"24"`
	ExpiresAt           string                        `json:"expires_at" example:"2024-01-15T10:30:00Z"`
	Status              marketplace.ListingStatus     `json:"status" example:"ACTIVE"`
	CurrentHighestBidID *string                       `json:"current_highest_bid_id,omitempty" example:"BID_987654321"`
	BidCount            int                           `json:"bid_count" example:"5"`
	Visibility          marketplace.ListingVisibility `json:"visibility" example:"PUBLIC"`
	AuctionType         marketplace.AuctionType       `json:"auction_type" example:"OPEN"`
	BidVisibility       marketplace.BidVisibility     `json:"bid_visibility" example:"FULL"`
	PickupLocation      *marketplace.Location         `json:"pickup_location,omitempty"`
	TermsConditions     string                        `json:"terms_conditions,omitempty" example:"Pickup within 7 days"`
	TimeRemaining       string                        `json:"time_remaining" example:"23h 45m"`
	CreatedAt           string                        `json:"created_at" example:"2024-01-14T10:30:00Z"`
	UpdatedAt           string                        `json:"updated_at" example:"2024-01-14T10:30:00Z"`
	ClosedAt            *string                       `json:"closed_at,omitempty" example:"2024-01-15T09:30:00Z"`
	CloseReason         *string                       `json:"close_reason,omitempty" example:"Auction completed"`
}

// ListingListResponse represents the response for listing list operations
type ListingListResponse struct {
	Listings   []ListingResponse     `json:"listings"`
	Pagination common.PaginationMeta `json:"pagination"`
}

// CreateListing creates a new marketplace listing
func (h *ListingHandler) CreateListing(c *gin.Context) {
	var req CreateListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "VALIDATION_ERROR", "Invalid input data", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	// Extract user information from context
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		common.BadRequest(c, "MISSING_ORG", "Organization ID not found in context", nil)
		return
	}

	// Parse decimal values
	quantity, err := decimal.NewFromString(req.Quantity)
	if err != nil {
		common.BadRequest(c, "INVALID_QUANTITY", "Invalid quantity format", map[string]interface{}{
			"details": "must be a valid decimal number",
		})
		return
	}

	askingPrice, err := decimal.NewFromString(req.AskingPrice)
	if err != nil {
		common.BadRequest(c, "INVALID_ASKING_PRICE", "Invalid asking price format", map[string]interface{}{
			"details": "must be a valid decimal number",
		})
		return
	}

	minimumBid, err := decimal.NewFromString(req.MinimumBid)
	if err != nil {
		common.BadRequest(c, "INVALID_MINIMUM_BID", "Invalid minimum bid format", map[string]interface{}{
			"details": "must be a valid decimal number",
		})
		return
	}

	// Create service request
	serviceReq := &marketplaceService.CreateListingRequest{
		ProductID:       req.ProductID,
		SellerID:        userID.(string),
		OrganizationID:  orgID,
		Quantity:        quantity,
		AskingPrice:     askingPrice,
		MinimumBid:      minimumBid,
		Currency:        req.Currency,
		ListingDuration: req.ListingDuration,
		Visibility:      req.Visibility,
		AuctionType:     req.AuctionType,
		BidVisibility:   req.BidVisibility,
		PickupLocation:  req.PickupLocation,
		TermsConditions: req.TermsConditions,
	}

	// Create listing
	listing, err := h.marketplaceServices.GetListingService().CreateListing(c.Request.Context(), serviceReq)
	if err != nil {
		common.InternalServerError(c, "LISTING_CREATION_FAILED", "Failed to create listing", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	// Convert to response format
	response := h.convertToListingResponse(listing)
	common.Created(c, response, nil)
}

// GetListing retrieves a specific listing by ID
func (h *ListingHandler) GetListing(c *gin.Context) {
	listingID := c.Param("id")
	if listingID == "" {
		common.BadRequest(c, "MISSING_LISTING_ID", "Listing ID is required", nil)
		return
	}

	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		common.BadRequest(c, "MISSING_ORG", "Organization ID not found in context", nil)
		return
	}

	// Get listing
	listing, err := h.marketplaceServices.GetListingService().GetListing(c.Request.Context(), listingID, orgID)
	if err != nil {
		if err.Error() == "listing not found: "+listingID {
			common.NotFound(c, "LISTING_NOT_FOUND", "Listing not found", nil)
			return
		}
		if err.Error() == "access denied: listing not visible to organization "+orgID {
			common.Forbidden(c, "ACCESS_DENIED", "Access denied", nil)
			return
		}
		common.InternalServerError(c, "LISTING_RETRIEVAL_FAILED", "Failed to retrieve listing", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	// Convert to response format
	response := h.convertToListingResponse(listing)
	common.Success(c, response, nil)
}

// GetActiveListings retrieves active marketplace listings
func (h *ListingHandler) GetActiveListings(c *gin.Context) {
	// Extract query options (includes deleted items if user is admin and include_deleted=true)
	middleware.ExtractQueryOptions(c)

	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		common.BadRequest(c, "MISSING_ORG", "Organization ID not found in context", nil)
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
		common.BadRequest(c, "INVALID_FILTERS", "Invalid filter parameters", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	// Get active listings
	listings, total, err := h.marketplaceServices.GetListingService().GetActiveListings(c.Request.Context(), orgID, filter, pagination)
	if err != nil {
		common.InternalServerError(c, "LISTINGS_RETRIEVAL_FAILED", "Failed to retrieve listings", map[string]interface{}{
			"details": err.Error(),
		})
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

// UpdateListing updates an existing listing
func (h *ListingHandler) UpdateListing(c *gin.Context) {
	listingID := c.Param("id")
	if listingID == "" {
		common.BadRequest(c, "MISSING_LISTING_ID", "Listing ID is required", nil)
		return
	}

	var req UpdateListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "VALIDATION_ERROR", "Invalid input data", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	// Extract user information from context
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		common.BadRequest(c, "MISSING_ORG", "Organization ID not found in context", nil)
		return
	}

	// Convert request to service format
	serviceReq := &marketplaceService.UpdateListingRequest{
		Visibility:      req.Visibility,
		BidVisibility:   req.BidVisibility,
		PickupLocation:  req.PickupLocation,
		TermsConditions: req.TermsConditions,
	}

	// Parse decimal values if provided
	if req.AskingPrice != nil {
		askingPrice, err := decimal.NewFromString(*req.AskingPrice)
		if err != nil {
			common.BadRequest(c, "INVALID_ASKING_PRICE", "Invalid asking price format", map[string]interface{}{
				"details": "must be a valid decimal number",
			})
			return
		}
		serviceReq.AskingPrice = &askingPrice
	}

	if req.MinimumBid != nil {
		minimumBid, err := decimal.NewFromString(*req.MinimumBid)
		if err != nil {
			common.BadRequest(c, "INVALID_MINIMUM_BID", "Invalid minimum bid format", map[string]interface{}{
				"details": "must be a valid decimal number",
			})
			return
		}
		serviceReq.MinimumBid = &minimumBid
	}

	// Update listing
	listing, err := h.marketplaceServices.GetListingService().UpdateListing(c.Request.Context(), listingID, serviceReq, userID.(string), orgID)
	if err != nil {
		if err.Error() == "listing not found: "+listingID {
			common.NotFound(c, "LISTING_NOT_FOUND", "Listing not found", nil)
			return
		}
		if err.Error() == "access denied: user "+userID.(string)+" cannot update listing "+listingID {
			common.Forbidden(c, "ACCESS_DENIED", "Access denied", nil)
			return
		}
		common.InternalServerError(c, "LISTING_UPDATE_FAILED", "Failed to update listing", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	// Convert to response format
	response := h.convertToListingResponse(listing)
	common.Success(c, response, nil)
}

// CloseListing closes an existing listing
func (h *ListingHandler) CloseListing(c *gin.Context) {
	listingID := c.Param("id")
	if listingID == "" {
		common.BadRequest(c, "MISSING_LISTING_ID", "Listing ID is required", nil)
		return
	}

	var req CloseListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "VALIDATION_ERROR", "Invalid input data", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	// Extract user information from context
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		common.BadRequest(c, "MISSING_ORG", "Organization ID not found in context", nil)
		return
	}

	// Close listing
	listing, err := h.marketplaceServices.GetListingService().CloseListing(c.Request.Context(), listingID, req.Reason, userID.(string), orgID)
	if err != nil {
		if err.Error() == "listing not found: "+listingID {
			common.NotFound(c, "LISTING_NOT_FOUND", "Listing not found", nil)
			return
		}
		if err.Error() == "access denied: user "+userID.(string)+" cannot close listing "+listingID {
			common.Forbidden(c, "ACCESS_DENIED", "Access denied", nil)
			return
		}
		common.InternalServerError(c, "LISTING_CLOSE_FAILED", "Failed to close listing", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	// Convert to response format
	response := h.convertToListingResponse(listing)
	common.Success(c, response, nil)
}

// GetMyListings retrieves listings for the authenticated user
func (h *ListingHandler) GetMyListings(c *gin.Context) {
	// Extract query options (includes deleted items if user is admin and include_deleted=true)
	middleware.ExtractQueryOptions(c)

	// Extract user information from context
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "UNAUTHORIZED", "User not authenticated", nil)
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
		common.BadRequest(c, "INVALID_FILTERS", "Invalid filter parameters", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	// Get seller listings
	listings, total, err := h.marketplaceServices.GetListingService().GetSellerListings(c.Request.Context(), userID.(string), filter, pagination)
	if err != nil {
		common.InternalServerError(c, "LISTINGS_RETRIEVAL_FAILED", "Failed to retrieve listings", map[string]interface{}{
			"details": err.Error(),
		})
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

// Helper methods

// convertToListingResponse converts a listing model to response format
func (h *ListingHandler) convertToListingResponse(listing *marketplace.Listing) ListingResponse {
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

// parsePaginationParams parses pagination parameters from query string
func (h *ListingHandler) parsePaginationParams(c *gin.Context) (*common.PaginationParams, error) {
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
func (h *ListingHandler) parseListingFilters(c *gin.Context) (*marketplace.ListingFilter, error) {
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

	// Price range filters
	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		if minPrice, err := decimal.NewFromString(minPriceStr); err == nil {
			filter.MinPrice = &minPrice
		}
	}

	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		if maxPrice, err := decimal.NewFromString(maxPriceStr); err == nil {
			filter.MaxPrice = &maxPrice
		}
	}

	return filter, nil
}

// formatTimeRemaining formats time duration to human-readable string
func (h *ListingHandler) formatTimeRemaining(duration time.Duration) string {
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
