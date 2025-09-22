package marketplace

import (
	"context"
	"fmt"
	"sync"
	"time"

	"kisanlink-ecom/entities/models/marketplace"
	"kisanlink-ecom/internal/common"
	marketplaceRepo "kisanlink-ecom/internal/repositories/marketplace"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"github.com/shopspring/decimal"
)

// BiddingServiceInterface defines the interface for bidding operations
type BiddingServiceInterface interface {
	// Bid Management
	PlaceBid(ctx context.Context, req *PlaceBidRequest) (*marketplace.Bid, error)
	GetBidHistory(ctx context.Context, listingID string, viewerID string, viewerOrgID string, pagination *common.PaginationParams) (*marketplace.BidHistory, error)
	GetUserBids(ctx context.Context, userID string, filter *marketplace.BidFilter, pagination *common.PaginationParams) ([]*marketplace.Bid, int, error)
	GetBid(ctx context.Context, bidID string, viewerID string) (*marketplace.Bid, error)

	// Auto-bidding
	EnableAutoBidding(ctx context.Context, req *EnableAutoBiddingRequest) (*marketplace.Bid, error)
	DisableAutoBidding(ctx context.Context, userID string, listingID string) error
	ProcessAutoBidding(ctx context.Context, listingID string, newBid *marketplace.Bid) ([]*marketplace.Bid, error)

	// Bid Status Management
	GetBidStatistics(ctx context.Context, listingID string) (*marketplace.BidStatistics, error)
	GetHighestBid(ctx context.Context, listingID string) (*marketplace.Bid, error)

	// Admin operations
	RemoveBid(ctx context.Context, bidID string, reason string, adminID string) error
	GetAllBids(ctx context.Context, filter *marketplace.BidFilter, pagination *common.PaginationParams) ([]*marketplace.Bid, int, error)
}

// PlaceBidRequest represents the request to place a new bid
type PlaceBidRequest struct {
	ListingID     string          `json:"listing_id" validate:"required"`
	BidderID      string          `json:"bidder_id" validate:"required"`
	BidAmount     decimal.Decimal `json:"bid_amount" validate:"required,gt=0"`
	Quantity      decimal.Decimal `json:"quantity" validate:"required,gt=0"`
	Message       string          `json:"message,omitempty"`
	PaymentMethod string          `json:"payment_method,omitempty"`
}

// EnableAutoBiddingRequest represents the request to enable auto-bidding
type EnableAutoBiddingRequest struct {
	ListingID    string          `json:"listing_id" validate:"required"`
	BidderID     string          `json:"bidder_id" validate:"required"`
	MaxBidAmount decimal.Decimal `json:"max_bid_amount" validate:"required,gt=0"`
	Increment    decimal.Decimal `json:"increment" validate:"required,gt=0"`
}

// BiddingService provides business logic for bidding operations
type BiddingService struct {
	bidRepo         marketplaceRepo.BidRepository
	listingRepo     marketplaceRepo.ListingRepository
	eventService    EventServiceInterface
	notificationSvc NotificationServiceInterface
	lockService     DistributedLockService
	cacheService    CacheService
	queryOptimizer  QueryOptimizer
	connectionPool  ConnectionPoolManager

	// Concurrency control (deprecated in favor of distributed locking)
	bidLocks map[string]*sync.RWMutex
	locksMux sync.RWMutex
}

// BiddingNotificationServiceInterface defines the interface for bidding notification operations
type BiddingNotificationServiceInterface interface {
	NotifyBidPlaced(ctx context.Context, bid *marketplace.Bid, listing *marketplace.Listing) error
	NotifyBidOutbid(ctx context.Context, outbidBid *marketplace.Bid, newBid *marketplace.Bid, listing *marketplace.Listing) error
	NotifyAutoBidTriggered(ctx context.Context, autoBid *marketplace.Bid, triggeringBid *marketplace.Bid, listing *marketplace.Listing) error
}

// NewBiddingService creates a new bidding service with performance optimizations
func NewBiddingService(
	bidRepo marketplaceRepo.BidRepository,
	listingRepo marketplaceRepo.ListingRepository,
	eventService EventServiceInterface,
	notificationSvc NotificationServiceInterface,
) BiddingServiceInterface {
	cacheService := NewCacheService()
	queryOptimizer := NewQueryOptimizer(bidRepo, listingRepo, cacheService)

	// Create connection pool (in production, this would use actual DB factory)
	connectionPool, _ := NewConnectionPoolManager(
		DefaultConnectionPoolConfig(),
		func() (db.DBManager, error) {
			// This would return actual DB manager in production
			return nil, fmt.Errorf("connection pool not implemented for this demo")
		},
	)

	return &BiddingService{
		bidRepo:         bidRepo,
		listingRepo:     listingRepo,
		eventService:    eventService,
		notificationSvc: notificationSvc,
		lockService:     NewDistributedLockService(),
		cacheService:    cacheService,
		queryOptimizer:  queryOptimizer,
		connectionPool:  connectionPool,
		bidLocks:        make(map[string]*sync.RWMutex),
		locksMux:        sync.RWMutex{},
	}
}

// PlaceBid places a bid with atomic processing and validation using distributed locking
func (s *BiddingService) PlaceBid(ctx context.Context, req *PlaceBidRequest) (*marketplace.Bid, error) {
	// Validate request
	if err := s.validatePlaceBidRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get listing and validate
	listing, err := s.listingRepo.GetByListingID(ctx, req.ListingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get listing: %w", err)
	}
	if listing == nil {
		return nil, fmt.Errorf("listing not found: %s", req.ListingID)
	}

	// Validate listing state
	if err := s.validateListingForBidding(listing, req.BidderID); err != nil {
		return nil, err
	}

	// Validate bid amount against minimum bid
	if req.BidAmount.LessThan(listing.MinimumBid) {
		return nil, fmt.Errorf("bid amount %s is below minimum bid %s", req.BidAmount.String(), listing.MinimumBid.String())
	}

	// Create atomic bid operation
	operation := &PlaceBidOperation{
		BidRepo:       s.bidRepo,
		ListingRepo:   s.listingRepo,
		ListingID:     req.ListingID,
		BidderID:      req.BidderID,
		BidAmount:     req.BidAmount,
		Quantity:      req.Quantity,
		Message:       req.Message,
		PaymentMethod: req.PaymentMethod,
		RetryPolicy:   DefaultRetryPolicy(),
	}

	// Execute atomic bid placement with distributed locking
	if err := s.lockService.ExecuteAtomicBidOperation(ctx, req.ListingID, operation); err != nil {
		return nil, fmt.Errorf("failed to place bid atomically: %w", err)
	}

	// Invalidate cache entries that are now stale
	_ = s.cacheService.InvalidateListingCache(ctx, req.ListingID)

	// Get the placed bid to return to caller
	placedBid, err := s.bidRepo.GetHighestBid(ctx, req.ListingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get placed bid: %w", err)
	}

	// Record bid placement event
	if s.eventService != nil {
		eventData := map[string]interface{}{
			"bid_id":      placedBid.BidID,
			"listing_id":  placedBid.ListingID,
			"bidder_id":   placedBid.BidderID,
			"bid_amount":  placedBid.BidAmount,
			"quantity":    placedBid.Quantity,
			"is_auto_bid": placedBid.IsAutoBid,
		}
		_ = s.eventService.RecordListingEvent(ctx, req.ListingID, marketplace.EventBidPlaced, eventData, req.BidderID)
	}

	// Send notifications
	if s.notificationSvc != nil {
		_ = s.notificationSvc.NotifyBidPlaced(ctx, placedBid, listing)

		// Get previous highest bid to notify if outbid
		// Note: This is a simplified approach - in production, we'd track this in the atomic operation
		bids, _, err := s.bidRepo.GetBidHistory(ctx, req.ListingID, marketplace.BidVisibilityFull, req.BidderID, &common.PaginationRequest{Limit: 2})
		if err == nil && len(bids) > 1 {
			previousHighest := bids[1] // Second highest is the previous highest
			if previousHighest.BidderID != placedBid.BidderID {
				_ = s.notificationSvc.NotifyBidOutbid(ctx, previousHighest, placedBid, listing)
			}
		}
	}

	// Process auto-bidding for other users asynchronously
	go func() {
		autoBids, err := s.ProcessAutoBidding(context.Background(), req.ListingID, placedBid)
		if err != nil {
			fmt.Printf("Warning: auto-bidding processing failed for listing %s: %v\n", req.ListingID, err)
		} else if len(autoBids) > 0 {
			fmt.Printf("Info: processed %d auto-bids for listing %s\n", len(autoBids), req.ListingID)
		}
	}()

	return placedBid, nil
}

// GetBidHistory retrieves bid history with visibility filtering
func (s *BiddingService) GetBidHistory(ctx context.Context, listingID string, viewerID string, viewerOrgID string, pagination *common.PaginationParams) (*marketplace.BidHistory, error) {
	// Get listing to check visibility settings
	listing, err := s.listingRepo.GetByListingID(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get listing: %w", err)
	}
	if listing == nil {
		return nil, fmt.Errorf("listing not found: %s", listingID)
	}

	// Check if viewer can access this listing
	if !s.canViewListing(listing, viewerOrgID) {
		return nil, fmt.Errorf("access denied: listing not visible")
	}

	// Convert pagination params
	paginationReq := &common.PaginationRequest{
		Limit:  pagination.Limit,
		Offset: pagination.CalculateOffset(),
	}

	// Get bid history based on visibility settings
	bids, total, err := s.bidRepo.GetBidHistory(ctx, listingID, listing.BidVisibility, viewerID, paginationReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get bid history: %w", err)
	}

	// Get highest bid
	highestBid, _ := s.bidRepo.GetHighestBid(ctx, listingID)

	// Build bid history response
	bidHistory := &marketplace.BidHistory{
		ListingID:   listingID,
		TotalBids:   total,
		Bids:        make([]marketplace.BidSummary, len(bids)),
		LastUpdated: time.Now(),
	}

	if highestBid != nil {
		bidHistory.HighestBid = &marketplace.BidSummary{
			ID:           highestBid.ID,
			BidID:        highestBid.BidID,
			ListingID:    highestBid.ListingID,
			BidderID:     s.anonymizeBidderID(highestBid.BidderID, viewerID),
			BidAmount:    highestBid.BidAmount,
			Status:       highestBid.Status,
			IsHighestBid: highestBid.IsHighestBid,
			IsAutoBid:    highestBid.IsAutoBid,
			PlacedAt:     highestBid.PlacedAt,
			OutbidAt:     highestBid.OutbidAt,
		}
	}

	// Convert bids to summaries with appropriate anonymization
	for i, bid := range bids {
		bidHistory.Bids[i] = marketplace.BidSummary{
			ID:           bid.ID,
			BidID:        bid.BidID,
			ListingID:    bid.ListingID,
			BidderID:     s.anonymizeBidderID(bid.BidderID, viewerID),
			BidAmount:    bid.BidAmount,
			Status:       bid.Status,
			IsHighestBid: bid.IsHighestBid,
			IsAutoBid:    bid.IsAutoBid,
			PlacedAt:     bid.PlacedAt,
			OutbidAt:     bid.OutbidAt,
		}
	}

	return bidHistory, nil
}

// GetUserBids retrieves all bids for a specific user
func (s *BiddingService) GetUserBids(ctx context.Context, userID string, filter *marketplace.BidFilter, pagination *common.PaginationParams) ([]*marketplace.Bid, int, error) {
	// Convert pagination params
	paginationReq := &common.PaginationRequest{
		Limit:  pagination.Limit,
		Offset: pagination.CalculateOffset(),
	}

	// Get user bids
	bids, total, err := s.bidRepo.GetUserBids(ctx, userID, filter, paginationReq)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user bids: %w", err)
	}

	return bids, total, nil
}

// GetBid retrieves a specific bid with access control
func (s *BiddingService) GetBid(ctx context.Context, bidID string, viewerID string) (*marketplace.Bid, error) {
	// Get bid
	bid, err := s.bidRepo.GetByBidID(ctx, bidID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bid: %w", err)
	}
	if bid == nil {
		return nil, fmt.Errorf("bid not found: %s", bidID)
	}

	// Check access - user can only see their own bids or admin can see all
	if bid.BidderID != viewerID {
		// Additional admin check would go here
		return nil, fmt.Errorf("access denied: cannot view bid %s", bidID)
	}

	return bid, nil
}

// EnableAutoBidding enables auto-bidding for a user on a listing
func (s *BiddingService) EnableAutoBidding(ctx context.Context, req *EnableAutoBiddingRequest) (*marketplace.Bid, error) {
	// Validate request
	if err := s.validateAutoBiddingRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get listing and validate
	listing, err := s.listingRepo.GetByListingID(ctx, req.ListingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get listing: %w", err)
	}
	if listing == nil {
		return nil, fmt.Errorf("listing not found: %s", req.ListingID)
	}

	// Validate listing state
	if err := s.validateListingForBidding(listing, req.BidderID); err != nil {
		return nil, err
	}

	// Check if user already has auto-bidding enabled
	existingAutoBids, err := s.bidRepo.GetUserAutoBids(ctx, req.BidderID, req.ListingID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing auto-bids: %w", err)
	}
	if len(existingAutoBids) > 0 {
		return nil, fmt.Errorf("auto-bidding already enabled for user %s on listing %s", req.BidderID, req.ListingID)
	}

	// Get current highest bid to determine initial bid amount
	currentHighest, err := s.bidRepo.GetHighestBid(ctx, req.ListingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current highest bid: %w", err)
	}

	// Calculate initial auto-bid amount
	initialBidAmount := listing.MinimumBid
	if currentHighest != nil {
		initialBidAmount = currentHighest.BidAmount.Add(req.Increment)
	}

	// Validate that initial bid is within auto-bid limit
	if initialBidAmount.GreaterThan(req.MaxBidAmount) {
		return nil, fmt.Errorf("initial auto-bid amount %s exceeds maximum limit %s", initialBidAmount.String(), req.MaxBidAmount.String())
	}

	// Create auto-bid
	autoBid := marketplace.NewAutoBid(req.ListingID, req.BidderID, initialBidAmount, listing.Quantity, req.MaxBidAmount, "")

	// Place the auto-bid
	placedBid, err := s.bidRepo.PlaceBidAtomic(ctx, autoBid, req.ListingID)
	if err != nil {
		return nil, fmt.Errorf("failed to place auto-bid: %w", err)
	}

	// Update listing counters
	_ = s.listingRepo.IncrementBidCount(ctx, req.ListingID)
	_ = s.listingRepo.UpdateHighestBid(ctx, req.ListingID, placedBid.BidID)

	// Record auto-bid event
	if s.eventService != nil {
		eventData := map[string]interface{}{
			"bid_id":         placedBid.BidID,
			"listing_id":     placedBid.ListingID,
			"bidder_id":      placedBid.BidderID,
			"bid_amount":     placedBid.BidAmount,
			"auto_bid_limit": req.MaxBidAmount,
			"increment":      req.Increment,
		}
		_ = s.eventService.RecordListingEvent(ctx, req.ListingID, marketplace.EventAutoBidTriggered, eventData, req.BidderID)
	}

	return placedBid, nil
}

// DisableAutoBidding disables auto-bidding for a user on a listing
func (s *BiddingService) DisableAutoBidding(ctx context.Context, userID string, listingID string) error {
	// Get user's auto-bids for this listing
	autoBids, err := s.bidRepo.GetUserAutoBids(ctx, userID, listingID)
	if err != nil {
		return fmt.Errorf("failed to get user auto-bids: %w", err)
	}

	if len(autoBids) == 0 {
		return fmt.Errorf("no active auto-bidding found for user %s on listing %s", userID, listingID)
	}

	// Disable all auto-bids
	for _, autoBid := range autoBids {
		if err := s.bidRepo.UpdateAutoBidStatus(ctx, autoBid.BidID, marketplace.BidStatusExpired); err != nil {
			return fmt.Errorf("failed to disable auto-bid %s: %w", autoBid.BidID, err)
		}
	}

	return nil
}

// ProcessAutoBidding processes auto-bidding logic with configurable limits and triggers
func (s *BiddingService) ProcessAutoBidding(ctx context.Context, listingID string, newBid *marketplace.Bid) ([]*marketplace.Bid, error) {
	// Don't process auto-bidding for auto-bids themselves
	if newBid.IsAutoBid {
		return nil, nil
	}

	// Get all active auto-bids for this listing
	autoBids, err := s.bidRepo.GetAutoBidsForListing(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get auto-bids: %w", err)
	}

	if len(autoBids) == 0 {
		return nil, nil // No auto-bids to process
	}

	// Get listing for minimum increment calculation
	listing, err := s.listingRepo.GetByListingID(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get listing: %w", err)
	}

	var processedBids []*marketplace.Bid
	minimumIncrement := decimal.NewFromFloat(1.0) // Default increment

	// Process each auto-bid
	for _, autoBid := range autoBids {
		// Skip auto-bids from the same user who placed the new bid
		if autoBid.BidderID == newBid.BidderID {
			continue
		}

		// Check if auto-bid can respond to this new bid
		nextBidAmount := autoBid.GetNextAutoBidAmount(newBid.BidAmount, minimumIncrement)
		if nextBidAmount.IsZero() {
			// Auto-bid limit reached, disable it
			_ = s.bidRepo.UpdateAutoBidStatus(ctx, autoBid.BidID, marketplace.BidStatusExpired)
			continue
		}

		// Create new auto-bid response
		responseBid := marketplace.NewAutoBid(
			listingID,
			autoBid.BidderID,
			nextBidAmount,
			listing.Quantity,
			*autoBid.AutoBidLimit,
			autoBid.BidID,
		)

		// Place the auto-bid response
		placedBid, err := s.bidRepo.PlaceBidAtomic(ctx, responseBid, listingID)
		if err != nil {
			fmt.Printf("Warning: failed to place auto-bid response: %v\n", err)
			continue
		}

		processedBids = append(processedBids, placedBid)

		// Update listing counters
		_ = s.listingRepo.IncrementBidCount(ctx, listingID)
		_ = s.listingRepo.UpdateHighestBid(ctx, listingID, placedBid.BidID)

		// Record auto-bid event
		if s.eventService != nil {
			eventData := map[string]interface{}{
				"bid_id":         placedBid.BidID,
				"listing_id":     placedBid.ListingID,
				"bidder_id":      placedBid.BidderID,
				"bid_amount":     placedBid.BidAmount,
				"triggering_bid": newBid.BidID,
				"parent_bid":     autoBid.BidID,
			}
			_ = s.eventService.RecordListingEvent(ctx, listingID, marketplace.EventAutoBidTriggered, eventData, autoBid.BidderID)
		}

		// Send notification
		if s.notificationSvc != nil {
			_ = s.notificationSvc.NotifyAutoBidTriggered(ctx, placedBid, newBid, listing)
		}
	}

	return processedBids, nil
}

// GetBidStatistics retrieves comprehensive bid statistics for a listing with caching
func (s *BiddingService) GetBidStatistics(ctx context.Context, listingID string) (*marketplace.BidStatistics, error) {
	// Try cache first
	if stats, err := s.cacheService.GetBidStatistics(ctx, listingID); err == nil {
		return stats, nil
	}

	// Cache miss, get from repository
	stats, err := s.bidRepo.GetBidStatistics(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bid statistics: %w", err)
	}

	// Cache the result for 2 minutes
	if stats != nil {
		_ = s.cacheService.SetBidStatistics(ctx, listingID, stats, 2*time.Minute)
	}

	return stats, nil
}

// GetHighestBid retrieves the current highest bid for a listing with optimizations
func (s *BiddingService) GetHighestBid(ctx context.Context, listingID string) (*marketplace.Bid, error) {
	// Use query optimizer for better performance
	bid, err := s.queryOptimizer.GetHighestBidOptimized(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get highest bid: %w", err)
	}
	return bid, nil
}

// RemoveBid removes a bid (admin operation)
func (s *BiddingService) RemoveBid(ctx context.Context, bidID string, reason string, adminID string) error {
	// Get the bid to be removed
	bid, err := s.bidRepo.GetByBidID(ctx, bidID)
	if err != nil {
		return fmt.Errorf("failed to get bid: %w", err)
	}
	if bid == nil {
		return fmt.Errorf("bid not found: %s", bidID)
	}

	// Remove the bid
	if err := s.bidRepo.RemoveBid(ctx, bidID, reason, adminID); err != nil {
		return fmt.Errorf("failed to remove bid: %w", err)
	}

	// If this was the highest bid, update the listing
	if bid.IsHighestBid {
		// Get the new highest bid
		newHighest, err := s.bidRepo.GetHighestBid(ctx, bid.ListingID)
		if err != nil {
			return fmt.Errorf("failed to get new highest bid: %w", err)
		}

		if newHighest != nil {
			_ = s.listingRepo.UpdateHighestBid(ctx, bid.ListingID, newHighest.BidID)
		} else {
			// No more bids, clear highest bid
			_ = s.listingRepo.UpdateHighestBid(ctx, bid.ListingID, "")
		}
	}

	return nil
}

// GetAllBids retrieves all bids (admin operation)
func (s *BiddingService) GetAllBids(ctx context.Context, filter *marketplace.BidFilter, pagination *common.PaginationParams) ([]*marketplace.Bid, int, error) {
	// Convert pagination params
	paginationReq := &common.PaginationRequest{
		Limit:  pagination.Limit,
		Offset: pagination.CalculateOffset(),
	}

	// Get all bids
	bids, total, err := s.bidRepo.GetAllBids(ctx, filter, paginationReq)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get all bids: %w", err)
	}

	return bids, total, nil
}

// Helper methods

// validatePlaceBidRequest validates the place bid request
func (s *BiddingService) validatePlaceBidRequest(req *PlaceBidRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.ListingID == "" {
		return fmt.Errorf("listing ID is required")
	}

	if req.BidderID == "" {
		return fmt.Errorf("bidder ID is required")
	}

	if req.BidAmount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("bid amount must be greater than zero")
	}

	if req.Quantity.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("quantity must be greater than zero")
	}

	return nil
}

// validateAutoBiddingRequest validates the auto-bidding request
func (s *BiddingService) validateAutoBiddingRequest(req *EnableAutoBiddingRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.ListingID == "" {
		return fmt.Errorf("listing ID is required")
	}

	if req.BidderID == "" {
		return fmt.Errorf("bidder ID is required")
	}

	if req.MaxBidAmount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("max bid amount must be greater than zero")
	}

	if req.Increment.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("increment must be greater than zero")
	}

	return nil
}

// validateListingForBidding validates that a listing can accept bids
func (s *BiddingService) validateListingForBidding(listing *marketplace.Listing, bidderID string) error {
	// Check if listing is active and not expired
	if !listing.CanAcceptBids() {
		return fmt.Errorf("listing %s cannot accept bids: status=%s, expired=%v", listing.ListingID, listing.Status, listing.IsExpired())
	}

	// Check if bidder is trying to bid on their own listing
	if listing.SellerID == bidderID {
		return fmt.Errorf("cannot bid on your own listing")
	}

	return nil
}

// validateBidAmount validates the bid amount against minimum and current highest
func (s *BiddingService) validateBidAmount(bidAmount, minimumBid decimal.Decimal, currentHighest *marketplace.Bid) error {
	// Check against minimum bid
	if bidAmount.LessThan(minimumBid) {
		return fmt.Errorf("bid amount %s is below minimum bid %s", bidAmount.String(), minimumBid.String())
	}

	// Check against current highest bid
	if currentHighest != nil && bidAmount.LessThanOrEqual(currentHighest.BidAmount) {
		return fmt.Errorf("bid amount %s must be higher than current highest bid %s", bidAmount.String(), currentHighest.BidAmount.String())
	}

	return nil
}

// canViewListing checks if a viewer can access a listing based on visibility
func (s *BiddingService) canViewListing(listing *marketplace.Listing, viewerOrgID string) bool {
	switch listing.Visibility {
	case marketplace.VisibilityPrivate:
		// For private listings, additional invitation logic would be needed
		return listing.OrganizationID == viewerOrgID
	case marketplace.VisibilityPublic:
		return true
	case marketplace.VisibilityNetwork:
		// For network visibility, network membership logic would be needed
		return true
	case marketplace.VisibilityOrganization:
		return listing.OrganizationID == viewerOrgID
	default:
		return false
	}
}

// anonymizeBidderID anonymizes bidder ID for display purposes
func (s *BiddingService) anonymizeBidderID(bidderID, viewerID string) string {
	// If viewer is the bidder, show their own ID
	if bidderID == viewerID {
		return bidderID
	}

	// Otherwise, return anonymized ID
	// In a real implementation, this would use a consistent anonymization scheme
	return fmt.Sprintf("Bidder_%s", bidderID[len(bidderID)-4:])
}

// getListingLock gets or creates a lock for a specific listing
func (s *BiddingService) getListingLock(listingID string) *sync.RWMutex {
	s.locksMux.Lock()
	defer s.locksMux.Unlock()

	if lock, exists := s.bidLocks[listingID]; exists {
		return lock
	}

	lock := &sync.RWMutex{}
	s.bidLocks[listingID] = lock
	return lock
}
