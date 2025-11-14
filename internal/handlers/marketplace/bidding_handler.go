package marketplace

import (
	"strconv"
	"time"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/marketplace"
	"github.com/Kisanlink/kisanlink-ecom/internal/common"
	"github.com/Kisanlink/kisanlink-ecom/internal/middleware"
	marketplaceService "github.com/Kisanlink/kisanlink-ecom/internal/services/marketplace"
	"github.com/Kisanlink/kisanlink-ecom/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// BiddingHandler handles marketplace bidding operations
type BiddingHandler struct {
	marketplaceServices *marketplaceService.MarketplaceServices
	visibilityService   marketplaceService.BidVisibilityServiceInterface
	resultsService      marketplaceService.AuctionResultsServiceInterface
}

// NewBiddingHandler creates a new bidding handler
func NewBiddingHandler(
	marketplaceServices *marketplaceService.MarketplaceServices,
	visibilityService marketplaceService.BidVisibilityServiceInterface,
	resultsService marketplaceService.AuctionResultsServiceInterface,
) *BiddingHandler {
	return &BiddingHandler{
		marketplaceServices: marketplaceServices,
		visibilityService:   visibilityService,
		resultsService:      resultsService,
	}
}

// PlaceBidRequest represents the request to place a bid
type PlaceBidRequest struct {
	BidAmount    string  `json:"bid_amount" binding:"required" example:"1200.00"`
	Quantity     string  `json:"quantity" binding:"required" example:"100.5"`
	Message      string  `json:"message,omitempty" example:"Interested in bulk purchase"`
	AutoBidLimit *string `json:"auto_bid_limit,omitempty" example:"1500.00"`
}

// BidResponse represents the response for bid operations
type BidResponse struct {
	ID           string                `json:"id" example:"uuid-123"`
	BidID        string                `json:"bid_id" example:"BID_1234567890"`
	ListingID    string                `json:"listing_id" example:"LST_1234567890"`
	BidderID     string                `json:"bidder_id" example:"USER_456"`
	BidAmount    string                `json:"bid_amount" example:"1200.00"`
	Currency     string                `json:"currency" example:"INR"`
	Quantity     string                `json:"quantity" example:"100.5"`
	Message      string                `json:"message,omitempty" example:"Interested in bulk purchase"`
	AutoBidLimit *string               `json:"auto_bid_limit,omitempty" example:"1500.00"`
	IsAutoBid    bool                  `json:"is_auto_bid" example:"false"`
	Status       marketplace.BidStatus `json:"status" example:"ACTIVE"`
	IsHighestBid bool                  `json:"is_highest_bid" example:"true"`
	IsOwnBid     bool                  `json:"is_own_bid" example:"false"`
	PlacedAt     string                `json:"placed_at" example:"2024-01-14T10:30:00Z"`
	OutbidAt     *string               `json:"outbid_at,omitempty" example:"2024-01-14T11:30:00Z"`
}

// BidListResponse represents the response for bid list operations
type BidListResponse struct {
	Bids       []BidResponse         `json:"bids"`
	Pagination common.PaginationMeta `json:"pagination"`
}

// BidHistoryResponse represents the bid history for a listing
type BidHistoryResponse struct {
	ListingID     string        `json:"listing_id" example:"LST_1234567890"`
	TotalBids     int           `json:"total_bids" example:"5"`
	HighestBid    *BidResponse  `json:"highest_bid,omitempty"`
	Bids          []BidResponse `json:"bids"`
	LastUpdated   string        `json:"last_updated" example:"2024-01-14T10:30:00Z"`
	CanViewBids   bool          `json:"can_view_bids" example:"true"`
	BidVisibility string        `json:"bid_visibility" example:"FULL"`
}

// AnonymousBidResponse represents an anonymized bid for visibility control
type AnonymousBidResponse struct {
	BidAmount    string `json:"bid_amount" example:"1200.00"`
	AnonymousID  string `json:"anonymous_id" example:"Bidder #1"`
	PlacedAt     string `json:"placed_at" example:"2024-01-14T10:30:00Z"`
	IsHighestBid bool   `json:"is_highest_bid" example:"true"`
	IsAutoBid    bool   `json:"is_auto_bid" example:"false"`
}

// RevealedAuctionResults represents auction results with visibility filtering applied
type RevealedAuctionResults struct {
	ListingID        string                    `json:"listing_id" example:"LST_1234567890"`
	Status           marketplace.ListingStatus `json:"status" example:"CLOSED"`
	WinnerID         *string                   `json:"winner_id,omitempty" example:"USER_456"`
	WinningBid       *string                   `json:"winning_bid,omitempty" example:"1500.00"`
	TotalBids        int                       `json:"total_bids" example:"12"`
	AuctionEndTime   string                    `json:"auction_end_time" example:"2024-01-14T15:00:00Z"`
	ResultsAvailable bool                      `json:"results_available" example:"true"`
}

// AuctionSummary represents a high-level summary of auction results
type AuctionSummary struct {
	ListingID    string                    `json:"listing_id" example:"LST_1234567890"`
	Title        string                    `json:"title" example:"Auction for Product PROD_123"`
	Status       marketplace.ListingStatus `json:"status" example:"CLOSED"`
	StartTime    string                    `json:"start_time" example:"2024-01-14T10:00:00Z"`
	EndTime      string                    `json:"end_time" example:"2024-01-14T15:00:00Z"`
	TotalBids    int                       `json:"total_bids" example:"12"`
	Participants int                       `json:"participants" example:"8"`
	WinnerID     *string                   `json:"winner_id,omitempty" example:"USER_456"`
	WinningBid   *string                   `json:"winning_bid,omitempty" example:"1500.00"`
}

// HistoricalBidData represents historical bid data with proper filtering
type HistoricalBidData struct {
	ListingID     string        `json:"listing_id" example:"LST_1234567890"`
	RequestedBy   string        `json:"requested_by" example:"USER_123"`
	TotalBids     int           `json:"total_bids" example:"12"`
	Bids          []BidResponse `json:"bids"`
	TimeRange     string        `json:"time_range" example:"2024-01-14T10:00:00Z to 2024-01-14T15:00:00Z"`
	DataAvailable bool          `json:"data_available" example:"true"`
	AccessLevel   string        `json:"access_level" example:"FULL"`
}

// FilteredBidStatistics represents bid statistics with applied visibility rules
type FilteredBidStatistics struct {
	ListingID       string  `json:"listing_id" example:"LST_1234567890"`
	TotalBids       int     `json:"total_bids" example:"12"`
	VisibleBids     int     `json:"visible_bids" example:"8"`
	HighestBid      *string `json:"highest_bid,omitempty" example:"1500.00"`
	AverageBid      *string `json:"average_bid,omitempty" example:"1200.00"`
	BidRange        *string `json:"bid_range,omitempty" example:"800.00 - 1500.00"`
	ParticipantInfo string  `json:"participant_info" example:"8 unique bidders"`
}

// PlaceBid places a new bid on a listing
// @Summary Place a bid on a marketplace listing
// @Description Place a new bid on an active marketplace listing
// @Tags Marketplace
// @Accept json
// @Produce json
// @Param X-Organization-ID header string true "Organization ID"
// @Param id path string true "Listing ID"
// @Param bid body PlaceBidRequest true "Bid placement request"
// @Success 201 {object} common.Response{data=BidResponse} "Bid placed successfully"
// @Failure 400 {object} common.Response{error=common.ResponseError} "Invalid request data"
// @Failure 401 {object} common.Response{error=common.ResponseError} "Unauthorized"
// @Failure 403 {object} common.Response{error=common.ResponseError} "Access denied"
// @Failure 404 {object} common.Response{error=common.ResponseError} "Listing not found"
// @Failure 409 {object} common.Response{error=common.ResponseError} "Bid conflict (too low, auction ended, etc.)"
// @Failure 500 {object} common.Response{error=common.ResponseError} "Internal server error"
// @Router /api/v1/marketplace/listings/{id}/bids [post]
// @Security BearerAuth
func (h *BiddingHandler) PlaceBid(c *gin.Context) {
	listingID := c.Param("id")
	if listingID == "" {
		common.BadRequest(c, "INVALID_REQUEST", "Listing ID is required", nil)
		return
	}

	var req PlaceBidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "VALIDATION_ERROR", "Invalid input data", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	// Extract user information from context
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		common.BadRequest(c, "MISSING_ORG", "Organization ID not found in context", nil)
		return
	}

	// Parse decimal values
	bidAmount, err := decimal.NewFromString(req.BidAmount)
	if err != nil {
		common.BadRequest(c, "VALIDATION_ERROR", "Invalid input data", map[string]interface{}{
			"fields": map[string]string{"bid_amount": "must be a valid decimal number"},
		})
		return
	}

	quantity, err := decimal.NewFromString(req.Quantity)
	if err != nil {
		common.BadRequest(c, "VALIDATION_ERROR", "Invalid input data", map[string]interface{}{
			"fields": map[string]string{"quantity": "must be a valid decimal number"},
		})
		return
	}

	// Parse auto-bid limit if provided (for future use)
	if req.AutoBidLimit != nil {
		_, err := decimal.NewFromString(*req.AutoBidLimit)
		if err != nil {
			common.BadRequest(c, "VALIDATION_ERROR", "Invalid input data", map[string]interface{}{
				"fields": map[string]string{"auto_bid_limit": "must be a valid decimal number"},
			})
			return
		}
		// Auto-bid limit validation passed, but not used in regular bid placement
	}

	// Create service request
	serviceReq := &marketplaceService.PlaceBidRequest{
		ListingID:     listingID,
		BidderID:      userID.(string),
		BidAmount:     bidAmount,
		Quantity:      quantity,
		Message:       req.Message,
		PaymentMethod: "", // Could be added to request if needed
	}

	// Place bid
	bid, err := h.marketplaceServices.GetBiddingService().PlaceBid(c.Request.Context(), serviceReq)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	// Convert to response format
	response := h.convertToBidResponse(bid)
	common.Created(c, response, nil)
}

// GetListingBids retrieves bids for a specific listing
// @Summary Get bids for a marketplace listing
// @Description Retrieve bids for a specific marketplace listing with visibility filtering
// @Tags Marketplace
// @Accept json
// @Produce json
// @Param X-Organization-ID header string true "Organization ID"
// @Param id path string true "Listing ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} common.Response{data=BidHistoryResponse} "Bids retrieved successfully"
// @Failure 400 {object} common.Response{error=common.ResponseError} "Invalid request parameters"
// @Failure 401 {object} common.Response{error=common.ResponseError} "Unauthorized"
// @Failure 403 {object} common.Response{error=common.ResponseError} "Access denied"
// @Failure 404 {object} common.Response{error=common.ResponseError} "Listing not found"
// @Failure 500 {object} common.Response{error=common.ResponseError} "Internal server error"
// @Router /api/v1/marketplace/listings/{id}/bids [get]
// @Security BearerAuth
func (h *BiddingHandler) GetListingBids(c *gin.Context) {
	// Extract query options (includes deleted items if user is admin and include_deleted=true)
	middleware.ExtractQueryOptions(c)

	listingID := c.Param("id")
	if listingID == "" {
		common.BadRequest(c, "INVALID_REQUEST", "Listing ID is required", nil)
		return
	}

	// Extract user information from context
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

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

	// Get bid history with visibility filtering using the new visibility service
	filteredHistory, err := h.visibilityService.GetVisibleBidHistory(c.Request.Context(), listingID, userID.(string), orgID, pagination)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	// Convert to response format
	response := h.convertToFilteredBidHistoryResponse(filteredHistory, pagination)
	common.Success(c, response, nil)
}

// GetBid retrieves a specific bid by ID
// @Summary Get a specific bid
// @Description Retrieve details of a specific bid by its ID
// @Tags Marketplace
// @Accept json
// @Produce json
// @Param X-Organization-ID header string true "Organization ID"
// @Param id path string true "Bid ID"
// @Success 200 {object} common.Response{data=BidResponse} "Bid retrieved successfully"
// @Failure 400 {object} common.Response{error=common.ResponseError} "Invalid request"
// @Failure 401 {object} common.Response{error=common.ResponseError} "Unauthorized"
// @Failure 403 {object} common.Response{error=common.ResponseError} "Access denied"
// @Failure 404 {object} common.Response{error=common.ResponseError} "Bid not found"
// @Failure 500 {object} common.Response{error=common.ResponseError} "Internal server error"
// @Router /api/v1/marketplace/bids/{id} [get]
// @Security BearerAuth
func (h *BiddingHandler) GetBid(c *gin.Context) {
	bidID := c.Param("id")
	if bidID == "" {
		common.BadRequest(c, "INVALID_REQUEST", "Bid ID is required", nil)
		return
	}

	// Extract user information from context
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		common.BadRequest(c, "MISSING_ORG", "Organization ID not found in context", nil)
		return
	}

	// Get bid
	bid, err := h.marketplaceServices.GetBiddingService().GetBid(c.Request.Context(), bidID, userID.(string))
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	// Convert to response format
	response := h.convertToBidResponse(bid)
	common.Success(c, response, nil)
}

// GetMyBids retrieves bids placed by the authenticated user
// @Summary Get my bids
// @Description Retrieve bids placed by the authenticated user across all listings
// @Tags Marketplace
// @Accept json
// @Produce json
// @Param status query string false "Filter by bid status" Enums(ACTIVE,OUTBID,WINNING,EXPIRED,REMOVED)
// @Param listing_id query string false "Filter by listing ID"
// @Param is_highest_bid query bool false "Filter by highest bid status"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} common.Response{data=BidListResponse} "Bids retrieved successfully"
// @Failure 400 {object} common.Response{error=common.ResponseError} "Invalid request parameters"
// @Failure 401 {object} common.Response{error=common.ResponseError} "Unauthorized"
// @Failure 500 {object} common.Response{error=common.ResponseError} "Internal server error"
// @Router /api/v1/marketplace/bids/my-bids [get]
// @Security BearerAuth
func (h *BiddingHandler) GetMyBids(c *gin.Context) {
	// Extract query options (includes deleted items if user is admin and include_deleted=true)
	middleware.ExtractQueryOptions(c)

	// Extract user information from context
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "UNAUTHORIZED", "Authentication required", nil)
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
	filter, err := h.parseBidFilters(c)
	if err != nil {
		common.BadRequest(c, "INVALID_FILTER", "Invalid filter parameters", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	// Get user bids
	bids, total, err := h.marketplaceServices.GetBiddingService().GetUserBids(c.Request.Context(), userID.(string), filter, pagination)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	// Convert to response format
	bidResponses := make([]BidResponse, len(bids))
	for i, bid := range bids {
		bidResponses[i] = h.convertToBidResponse(bid)
	}

	common.Success(c, gin.H{
		"bids": bidResponses,
	}, &common.ResponseMeta{
		Pagination: common.NewPaginationMeta(pagination.Page, pagination.Limit, total),
	})
}

// Helper methods

// convertToBidResponse converts a bid model to response format
func (h *BiddingHandler) convertToBidResponse(bid *marketplace.Bid) BidResponse {
	response := BidResponse{
		ID:           bid.ID,
		BidID:        bid.BidID,
		ListingID:    bid.ListingID,
		BidderID:     bid.BidderID,
		BidAmount:    bid.BidAmount.String(),
		Currency:     bid.Currency,
		Quantity:     bid.Quantity.String(),
		Message:      bid.Message,
		IsAutoBid:    bid.IsAutoBid,
		Status:       bid.Status,
		IsHighestBid: bid.IsHighestBid,
		PlacedAt:     bid.PlacedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Set optional fields
	if bid.AutoBidLimit != nil {
		autoBidLimit := bid.AutoBidLimit.String()
		response.AutoBidLimit = &autoBidLimit
	}

	if bid.OutbidAt != nil {
		outbidAt := bid.OutbidAt.Format("2006-01-02T15:04:05Z")
		response.OutbidAt = &outbidAt
	}

	return response
}

// convertToBidHistoryResponse converts bid history to response format
func (h *BiddingHandler) convertToBidHistoryResponse(bidHistory *marketplace.BidHistory, pagination *common.PaginationParams) BidHistoryResponse {
	response := BidHistoryResponse{
		ListingID:     bidHistory.ListingID,
		TotalBids:     bidHistory.TotalBids,
		Bids:          make([]BidResponse, len(bidHistory.Bids)),
		LastUpdated:   bidHistory.LastUpdated.Format("2006-01-02T15:04:05Z"),
		CanViewBids:   true,   // This would be determined by visibility rules
		BidVisibility: "FULL", // This would come from the listing configuration
	}

	// Convert highest bid if available
	if bidHistory.HighestBid != nil {
		highestBid := h.convertBidSummaryToResponse(bidHistory.HighestBid)
		response.HighestBid = &highestBid
	}

	// Convert bids
	for i, bidSummary := range bidHistory.Bids {
		response.Bids[i] = h.convertBidSummaryToResponse(&bidSummary)
	}

	return response
}

// convertBidSummaryToResponse converts a bid summary to response format
func (h *BiddingHandler) convertBidSummaryToResponse(bidSummary *marketplace.BidSummary) BidResponse {
	response := BidResponse{
		ID:           bidSummary.ID,
		BidID:        bidSummary.BidID,
		ListingID:    bidSummary.ListingID,
		BidderID:     bidSummary.BidderID,
		BidAmount:    bidSummary.BidAmount.String(),
		Currency:     "INR", // Default currency
		Quantity:     "1",   // Default quantity for summary
		IsAutoBid:    bidSummary.IsAutoBid,
		Status:       bidSummary.Status,
		IsHighestBid: bidSummary.IsHighestBid,
		PlacedAt:     bidSummary.PlacedAt.Format("2006-01-02T15:04:05Z"),
	}

	if bidSummary.OutbidAt != nil {
		outbidAt := bidSummary.OutbidAt.Format("2006-01-02T15:04:05Z")
		response.OutbidAt = &outbidAt
	}

	return response
}

// parsePaginationParams parses pagination parameters from query string
func (h *BiddingHandler) parsePaginationParams(c *gin.Context) (*common.PaginationParams, error) {
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

// parseBidFilters parses filter parameters from query string
func (h *BiddingHandler) parseBidFilters(c *gin.Context) (*marketplace.BidFilter, error) {
	filter := &marketplace.BidFilter{}

	// Status filter
	if statusStr := c.Query("status"); statusStr != "" {
		status := marketplace.BidStatus(statusStr)
		filter.Status = &status
	}

	// Listing ID filter
	if listingID := c.Query("listing_id"); listingID != "" {
		filter.ListingID = listingID
	}

	// Highest bid filter
	if isHighestBidStr := c.Query("is_highest_bid"); isHighestBidStr != "" {
		if isHighestBid, err := strconv.ParseBool(isHighestBidStr); err == nil {
			filter.IsHighestBid = &isHighestBid
		}
	}

	// Auto bid filter
	if isAutoBidStr := c.Query("is_auto_bid"); isAutoBidStr != "" {
		if isAutoBid, err := strconv.ParseBool(isAutoBidStr); err == nil {
			filter.IsAutoBid = &isAutoBid
		}
	}

	return filter, nil
}

// convertToFilteredBidHistoryResponse converts filtered bid history to response format
func (h *BiddingHandler) convertToFilteredBidHistoryResponse(filteredHistory *marketplaceService.FilteredBidHistory, pagination *common.PaginationParams) BidHistoryResponse {
	response := BidHistoryResponse{
		ListingID:     filteredHistory.ListingID,
		TotalBids:     filteredHistory.TotalBids,
		Bids:          make([]BidResponse, len(filteredHistory.Bids)),
		LastUpdated:   filteredHistory.LastUpdated.Format("2006-01-02T15:04:05Z"),
		CanViewBids:   filteredHistory.VisibleBids > 0,
		BidVisibility: string(filteredHistory.VisibilityLevel),
	}

	// Convert highest bid if available
	if filteredHistory.HighestBid != nil {
		highestBid := h.convertFilteredBidToResponse(filteredHistory.HighestBid)
		response.HighestBid = &highestBid
	}

	// Convert filtered bids
	for i, filteredBid := range filteredHistory.Bids {
		response.Bids[i] = h.convertFilteredBidToResponse(filteredBid)
	}

	return response
}

// convertFilteredBidToResponse converts a filtered bid to response format
func (h *BiddingHandler) convertFilteredBidToResponse(filteredBid *marketplaceService.FilteredBid) BidResponse {
	response := BidResponse{
		BidderID: filteredBid.AnonymousID, // Use anonymous ID for display
		Currency: "INR",                   // Default currency
		Quantity: "1",                     // Default quantity
		IsOwnBid: filteredBid.IsOwnBid,
	}

	// Set fields based on visibility level
	if filteredBid.BidID != "" {
		response.BidID = filteredBid.BidID
	}

	if filteredBid.BidAmount != nil {
		response.BidAmount = *filteredBid.BidAmount
	}

	if filteredBid.PlacedAt != nil {
		response.PlacedAt = filteredBid.PlacedAt.Format("2006-01-02T15:04:05Z")
	}

	if filteredBid.IsHighestBid != nil {
		response.IsHighestBid = *filteredBid.IsHighestBid
	}

	if filteredBid.IsAutoBid != nil {
		response.IsAutoBid = *filteredBid.IsAutoBid
	}

	if filteredBid.Status != nil {
		response.Status = *filteredBid.Status
	}

	return response
}

// GetAuctionResults retrieves auction results with proper visibility filtering
// @Summary Get auction results
// @Description Retrieve auction results with visibility filtering based on auction configuration
// @Tags Marketplace
// @Accept json
// @Produce json
// @Param X-Organization-ID header string true "Organization ID"
// @Param id path string true "Listing ID"
// @Success 200 {object} common.Response{data=RevealedAuctionResults} "Auction results retrieved successfully"
// @Failure 400 {object} common.Response{error=common.ResponseError} "Invalid request"
// @Failure 401 {object} common.Response{error=common.ResponseError} "Unauthorized"
// @Failure 403 {object} common.Response{error=common.ResponseError} "Access denied"
// @Failure 404 {object} common.Response{error=common.ResponseError} "Listing not found"
// @Failure 500 {object} common.Response{error=common.ResponseError} "Internal server error"
// @Router /api/v1/marketplace/listings/{id}/results [get]
// @Security BearerAuth
func (h *BiddingHandler) GetAuctionResults(c *gin.Context) {
	listingID := c.Param("id")
	if listingID == "" {
		common.BadRequest(c, "INVALID_REQUEST", "Listing ID is required", nil)
		return
	}

	// Extract user information from context
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		common.BadRequest(c, "MISSING_ORG", "Organization ID not found in context", nil)
		return
	}

	// Get auction results with visibility filtering
	results, err := h.resultsService.RevealAuctionResults(c.Request.Context(), listingID, userID.(string), orgID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	common.Success(c, results, nil)
}

// GetAuctionSummary retrieves a high-level auction summary
// @Summary Get auction summary
// @Description Retrieve a high-level summary of auction results
// @Tags Marketplace
// @Accept json
// @Produce json
// @Param X-Organization-ID header string true "Organization ID"
// @Param id path string true "Listing ID"
// @Success 200 {object} common.Response{data=AuctionSummary} "Auction summary retrieved successfully"
// @Failure 400 {object} common.Response{error=common.ResponseError} "Invalid request"
// @Failure 401 {object} common.Response{error=common.ResponseError} "Unauthorized"
// @Failure 403 {object} common.Response{error=common.ResponseError} "Access denied"
// @Failure 404 {object} common.Response{error=common.ResponseError} "Listing not found"
// @Failure 500 {object} common.Response{error=common.ResponseError} "Internal server error"
// @Router /api/v1/marketplace/listings/{id}/summary [get]
// @Security BearerAuth
func (h *BiddingHandler) GetAuctionSummary(c *gin.Context) {
	listingID := c.Param("id")
	if listingID == "" {
		common.BadRequest(c, "INVALID_REQUEST", "Listing ID is required", nil)
		return
	}

	// Extract user information from context
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		common.BadRequest(c, "MISSING_ORG", "Organization ID not found in context", nil)
		return
	}

	// Get auction summary
	summary, err := h.resultsService.GetAuctionSummary(c.Request.Context(), listingID, userID.(string), orgID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	common.Success(c, summary, nil)
}

// GetHistoricalBidData retrieves historical bid data with proper filtering
// @Summary Get historical bid data
// @Description Retrieve historical bid data with proper access control and filtering
// @Tags Marketplace
// @Accept json
// @Produce json
// @Param X-Organization-ID header string true "Organization ID"
// @Param id path string true "Listing ID"
// @Param start_time query string false "Start time filter (RFC3339 format)"
// @Param end_time query string false "End time filter (RFC3339 format)"
// @Param bidder_id query string false "Filter by specific bidder ID"
// @Param include_auto_bids query bool false "Include auto bids" default(true)
// @Param sort_order query string false "Sort order (asc/desc)" default(desc)
// @Success 200 {object} common.Response{data=HistoricalBidData} "Historical bid data retrieved successfully"
// @Failure 400 {object} common.Response{error=common.ResponseError} "Invalid request"
// @Failure 401 {object} common.Response{error=common.ResponseError} "Unauthorized"
// @Failure 403 {object} common.Response{error=common.ResponseError} "Access denied"
// @Failure 404 {object} common.Response{error=common.ResponseError} "Listing not found"
// @Failure 500 {object} common.Response{error=common.ResponseError} "Internal server error"
// @Router /api/v1/marketplace/listings/{id}/historical-bids [get]
// @Security BearerAuth
func (h *BiddingHandler) GetHistoricalBidData(c *gin.Context) {
	listingID := c.Param("id")
	if listingID == "" {
		common.BadRequest(c, "INVALID_REQUEST", "Listing ID is required", nil)
		return
	}

	// Extract user information from context
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		common.BadRequest(c, "MISSING_ORG", "Organization ID not found in context", nil)
		return
	}

	// Parse historical data filters
	filter, err := h.parseHistoricalDataFilters(c)
	if err != nil {
		common.BadRequest(c, "INVALID_FILTER", "Invalid filter parameters", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	// Get historical bid data
	historicalData, err := h.resultsService.GetHistoricalBidData(c.Request.Context(), listingID, userID.(string), orgID, filter)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	common.Success(c, historicalData, nil)
}

// GetBidStatistics retrieves bid statistics with visibility filtering
// @Summary Get bid statistics
// @Description Retrieve bid statistics with visibility filtering applied
// @Tags Marketplace
// @Accept json
// @Produce json
// @Param X-Organization-ID header string true "Organization ID"
// @Param id path string true "Listing ID"
// @Success 200 {object} common.Response{data=FilteredBidStatistics} "Bid statistics retrieved successfully"
// @Failure 400 {object} common.Response{error=common.ResponseError} "Invalid request"
// @Failure 401 {object} common.Response{error=common.ResponseError} "Unauthorized"
// @Failure 403 {object} common.Response{error=common.ResponseError} "Access denied"
// @Failure 404 {object} common.Response{error=common.ResponseError} "Listing not found"
// @Failure 500 {object} common.Response{error=common.ResponseError} "Internal server error"
// @Router /api/v1/marketplace/listings/{id}/statistics [get]
// @Security BearerAuth
func (h *BiddingHandler) GetBidStatistics(c *gin.Context) {
	listingID := c.Param("id")
	if listingID == "" {
		common.BadRequest(c, "INVALID_REQUEST", "Listing ID is required", nil)
		return
	}

	// Extract user information from context
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		common.BadRequest(c, "MISSING_ORG", "Organization ID not found in context", nil)
		return
	}

	// Get filtered bid statistics
	statistics, err := h.visibilityService.GetVisibleBidStatistics(c.Request.Context(), listingID, userID.(string), orgID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	common.Success(c, statistics, nil)
}

// parseHistoricalDataFilters parses historical data filter parameters from query string
func (h *BiddingHandler) parseHistoricalDataFilters(c *gin.Context) (*marketplaceService.HistoricalDataFilter, error) {
	filter := &marketplaceService.HistoricalDataFilter{
		IncludeAutoBids: true,   // Default to including auto bids
		SortOrder:       "desc", // Default to descending order
	}

	// Parse time filters
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filter.StartTime = &startTime
		}
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filter.EndTime = &endTime
		}
	}

	// Parse bidder ID filter
	if bidderID := c.Query("bidder_id"); bidderID != "" {
		filter.BidderID = bidderID
	}

	// Parse amount filters
	if minAmountStr := c.Query("min_amount"); minAmountStr != "" {
		if minAmount, err := decimal.NewFromString(minAmountStr); err == nil {
			filter.MinAmount = &minAmount
		}
	}

	if maxAmountStr := c.Query("max_amount"); maxAmountStr != "" {
		if maxAmount, err := decimal.NewFromString(maxAmountStr); err == nil {
			filter.MaxAmount = &maxAmount
		}
	}

	// Parse boolean filters
	if includeAutoBidsStr := c.Query("include_auto_bids"); includeAutoBidsStr != "" {
		if includeAutoBids, err := strconv.ParseBool(includeAutoBidsStr); err == nil {
			filter.IncludeAutoBids = includeAutoBids
		}
	}

	// Parse sort order
	if sortOrder := c.Query("sort_order"); sortOrder == "asc" || sortOrder == "desc" {
		filter.SortOrder = sortOrder
	}

	return filter, nil
}
