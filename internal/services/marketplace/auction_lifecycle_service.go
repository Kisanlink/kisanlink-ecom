package marketplace

import (
	"context"
	"fmt"
	"time"

	"kisanlink-ecom/entities/models/marketplace"

	"github.com/shopspring/decimal"
)

// AuctionLifecycleServiceInterface defines the interface for auction lifecycle management
type AuctionLifecycleServiceInterface interface {
	// Auction Expiry Processing
	ProcessExpiredAuctions(ctx context.Context) (*AuctionProcessingResult, error)
	ProcessSingleExpiredAuction(ctx context.Context, listingID string) (*AuctionResult, error)

	// Auction Closure
	CloseAuction(ctx context.Context, listingID string, reason string) (*AuctionResult, error)
	DetermineWinner(ctx context.Context, listingID string) (*marketplace.Bid, error)

	// Order Integration
	CreateOrderFromWinningBid(ctx context.Context, winningBid *marketplace.Bid, listing *marketplace.Listing) (*OrderCreationResult, error)

	// Scheduled Jobs
	StartAuctionExpiryScheduler(ctx context.Context, interval time.Duration) error
	StopAuctionExpiryScheduler() error
}

// AuctionProcessingResult represents the result of processing multiple expired auctions
type AuctionProcessingResult struct {
	ProcessedCount  int              `json:"processed_count"`
	SuccessfulCount int              `json:"successful_count"`
	FailedCount     int              `json:"failed_count"`
	Results         []*AuctionResult `json:"results"`
	ProcessingTime  time.Duration    `json:"processing_time"`
	Errors          []string         `json:"errors,omitempty"`
}

// AuctionResult represents the result of processing a single auction
type AuctionResult struct {
	ListingID     string                    `json:"listing_id"`
	Status        marketplace.ListingStatus `json:"status"`
	WinningBid    *marketplace.Bid          `json:"winning_bid,omitempty"`
	TotalBids     int                       `json:"total_bids"`
	UniqueBidders int                       `json:"unique_bidders"`
	OrderCreated  bool                      `json:"order_created"`
	OrderID       string                    `json:"order_id,omitempty"`
	ProcessedAt   time.Time                 `json:"processed_at"`
	Error         string                    `json:"error,omitempty"`
}

// OrderCreationResult represents the result of creating an order from a winning bid
type OrderCreationResult struct {
	OrderID   string          `json:"order_id"`
	BidID     string          `json:"bid_id"`
	ListingID string          `json:"listing_id"`
	BuyerID   string          `json:"buyer_id"`
	SellerID  string          `json:"seller_id"`
	Amount    decimal.Decimal `json:"amount"`
	Quantity  decimal.Decimal `json:"quantity"`
	Status    string          `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
	ExpiresAt *time.Time      `json:"expires_at,omitempty"`
}

// OrderServiceInterface defines the interface for order operations
type OrderServiceInterface interface {
	CreateOrderFromAuction(ctx context.Context, req *CreateOrderFromAuctionRequest) (*OrderCreationResult, error)
}

// CreateOrderFromAuctionRequest represents the request to create an order from an auction
type CreateOrderFromAuctionRequest struct {
	ListingID    string          `json:"listing_id"`
	WinningBidID string          `json:"winning_bid_id"`
	BuyerID      string          `json:"buyer_id"`
	SellerID     string          `json:"seller_id"`
	ProductID    string          `json:"product_id"`
	Quantity     decimal.Decimal `json:"quantity"`
	Price        decimal.Decimal `json:"price"`
	Currency     string          `json:"currency"`
}

// AuctionLifecycleService provides business logic for auction lifecycle management
type AuctionLifecycleService struct {
	listingRepo      ListingRepositoryInterface
	bidRepo          BidRepositoryInterface
	eventService     EventServiceInterface
	orderService     OrderServiceInterface
	notificationSvc  NotificationServiceInterface
	inventoryService InventoryServiceInterface

	// Scheduler control
	schedulerRunning bool
	schedulerStop    chan bool
}

// NewAuctionLifecycleService creates a new auction lifecycle service
func NewAuctionLifecycleService(
	listingRepo ListingRepositoryInterface,
	bidRepo BidRepositoryInterface,
	eventService EventServiceInterface,
	orderService OrderServiceInterface,
	notificationSvc NotificationServiceInterface,
	inventoryService InventoryServiceInterface,
) AuctionLifecycleServiceInterface {
	return &AuctionLifecycleService{
		listingRepo:      listingRepo,
		bidRepo:          bidRepo,
		eventService:     eventService,
		orderService:     orderService,
		notificationSvc:  notificationSvc,
		inventoryService: inventoryService,
		schedulerRunning: false,
		schedulerStop:    make(chan bool),
	}
}

// ProcessExpiredAuctions processes all expired auctions automatically
func (s *AuctionLifecycleService) ProcessExpiredAuctions(ctx context.Context) (*AuctionProcessingResult, error) {
	startTime := time.Now()

	// Get expired listings (limit to 100 at a time)
	expiredListings, err := s.listingRepo.GetExpiredListings(ctx, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get expired listings: %w", err)
	}

	result := &AuctionProcessingResult{
		ProcessedCount:  len(expiredListings),
		SuccessfulCount: 0,
		FailedCount:     0,
		Results:         make([]*AuctionResult, 0, len(expiredListings)),
		ProcessingTime:  0,
		Errors:          make([]string, 0),
	}

	// Process each expired listing
	for _, listing := range expiredListings {
		auctionResult, err := s.ProcessSingleExpiredAuction(ctx, listing.ListingID)
		if err != nil {
			result.FailedCount++
			result.Errors = append(result.Errors, fmt.Sprintf("Listing %s: %v", listing.ListingID, err))

			// Create error result
			auctionResult = &AuctionResult{
				ListingID:   listing.ListingID,
				Status:      listing.Status,
				ProcessedAt: time.Now(),
				Error:       err.Error(),
			}
		} else {
			result.SuccessfulCount++
		}

		result.Results = append(result.Results, auctionResult)
	}

	result.ProcessingTime = time.Since(startTime)
	return result, nil
}

// ProcessSingleExpiredAuction processes a single expired auction
func (s *AuctionLifecycleService) ProcessSingleExpiredAuction(ctx context.Context, listingID string) (*AuctionResult, error) {
	// Get the listing
	listing, err := s.listingRepo.GetByListingID(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get listing: %w", err)
	}
	if listing == nil {
		return nil, fmt.Errorf("listing not found: %s", listingID)
	}

	// Check if listing is actually expired and active
	if listing.Status != marketplace.ListingStatusActive || !listing.IsExpired() {
		return nil, fmt.Errorf("listing %s is not expired or not active", listingID)
	}

	// Get bid statistics
	bidStats, err := s.bidRepo.GetBidStatistics(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bid statistics: %w", err)
	}

	result := &AuctionResult{
		ListingID:     listingID,
		TotalBids:     bidStats.TotalBids,
		UniqueBidders: bidStats.UniqueBidders,
		ProcessedAt:   time.Now(),
	}

	// Determine winner and close auction
	if bidStats.TotalBids == 0 {
		// No bids - mark as expired with no bids
		if err := listing.UpdateStatus(marketplace.ListingStatusExpiredNoBids, "Auction expired with no bids"); err != nil {
			return nil, fmt.Errorf("failed to update listing status: %w", err)
		}
		result.Status = marketplace.ListingStatusExpiredNoBids
	} else {
		// Determine winner
		winningBid, err := s.DetermineWinner(ctx, listingID)
		if err != nil {
			return nil, fmt.Errorf("failed to determine winner: %w", err)
		}

		if winningBid == nil {
			// No valid winner found
			if err := listing.UpdateStatus(marketplace.ListingStatusExpired, "Auction expired with no valid winner"); err != nil {
				return nil, fmt.Errorf("failed to update listing status: %w", err)
			}
			result.Status = marketplace.ListingStatusExpired
		} else {
			// Mark winning bid
			if err := winningBid.UpdateStatus(marketplace.BidStatusWinning); err != nil {
				return nil, fmt.Errorf("failed to update winning bid status: %w", err)
			}
			if err := s.bidRepo.Update(ctx, winningBid); err != nil {
				return nil, fmt.Errorf("failed to save winning bid: %w", err)
			}

			// Close auction
			if err := listing.UpdateStatus(marketplace.ListingStatusClosed, "Auction closed - winner determined"); err != nil {
				return nil, fmt.Errorf("failed to update listing status: %w", err)
			}

			result.Status = marketplace.ListingStatusClosed
			result.WinningBid = winningBid

			// Create order opportunity
			if s.orderService != nil {
				orderResult, err := s.CreateOrderFromWinningBid(ctx, winningBid, listing)
				if err != nil {
					// Log error but don't fail the auction closure
					fmt.Printf("Warning: failed to create order for winning bid %s: %v\n", winningBid.BidID, err)
				} else {
					result.OrderCreated = true
					result.OrderID = orderResult.OrderID
				}
			}
		}
	}

	// Expire all remaining active bids
	if err := s.bidRepo.ExpireBidsForListing(ctx, listingID); err != nil {
		fmt.Printf("Warning: failed to expire bids for listing %s: %v\n", listingID, err)
	}

	// Save updated listing
	if err := s.listingRepo.Update(ctx, listing); err != nil {
		return nil, fmt.Errorf("failed to save listing: %w", err)
	}

	// Release inventory
	if s.inventoryService != nil {
		_ = s.inventoryService.ReleaseInventory(ctx, listing.ProductID, listing.Quantity, listing.ListingID)
	}

	// Record auction closure event
	if s.eventService != nil {
		eventData := map[string]interface{}{
			"listing_id":     listingID,
			"final_status":   result.Status,
			"total_bids":     result.TotalBids,
			"unique_bidders": result.UniqueBidders,
			"order_created":  result.OrderCreated,
		}
		if result.WinningBid != nil {
			eventData["winning_bid_id"] = result.WinningBid.BidID
			eventData["winning_amount"] = result.WinningBid.BidAmount
			eventData["winner_id"] = result.WinningBid.BidderID
		}
		_ = s.eventService.RecordListingEvent(ctx, listingID, marketplace.EventListingExpired, eventData, "system")
	}

	// Send notifications
	if s.notificationSvc != nil {
		if result.WinningBid != nil {
			// Notify winner
			_ = s.notificationSvc.NotifyAuctionWon(ctx, listingID, result.WinningBid.BidderID, result.WinningBid.BidAmount.String())
			// Notify seller about successful auction
			_ = s.notificationSvc.NotifyAuctionExpired(ctx, listingID, listing.SellerID, result.WinningBid.BidderID)
		} else {
			// Notify seller about auction with no valid winner
			_ = s.notificationSvc.NotifyAuctionExpired(ctx, listingID, listing.SellerID, "")
		}
	}

	return result, nil
}

// CloseAuction manually closes an auction and determines the winner
func (s *AuctionLifecycleService) CloseAuction(ctx context.Context, listingID string, reason string) (*AuctionResult, error) {
	// Get the listing
	listing, err := s.listingRepo.GetByListingID(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get listing: %w", err)
	}
	if listing == nil {
		return nil, fmt.Errorf("listing not found: %s", listingID)
	}

	// Check if listing can be closed
	if listing.Status != marketplace.ListingStatusActive {
		return nil, fmt.Errorf("listing %s cannot be closed: current status is %s", listingID, listing.Status)
	}

	// Get bid statistics
	bidStats, err := s.bidRepo.GetBidStatistics(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bid statistics: %w", err)
	}

	result := &AuctionResult{
		ListingID:     listingID,
		TotalBids:     bidStats.TotalBids,
		UniqueBidders: bidStats.UniqueBidders,
		ProcessedAt:   time.Now(),
	}

	// Determine winner if there are bids
	if bidStats.TotalBids > 0 {
		winningBid, err := s.DetermineWinner(ctx, listingID)
		if err != nil {
			return nil, fmt.Errorf("failed to determine winner: %w", err)
		}

		if winningBid != nil {
			// Mark winning bid
			if err := winningBid.UpdateStatus(marketplace.BidStatusWinning); err != nil {
				return nil, fmt.Errorf("failed to update winning bid status: %w", err)
			}
			if err := s.bidRepo.Update(ctx, winningBid); err != nil {
				return nil, fmt.Errorf("failed to save winning bid: %w", err)
			}
			result.WinningBid = winningBid
		}
	}

	// Close the listing
	closeReason := reason
	if closeReason == "" {
		closeReason = "Auction manually closed"
	}

	if err := listing.UpdateStatus(marketplace.ListingStatusClosed, closeReason); err != nil {
		return nil, fmt.Errorf("failed to update listing status: %w", err)
	}
	result.Status = marketplace.ListingStatusClosed

	// Expire all remaining active bids
	if err := s.bidRepo.ExpireBidsForListing(ctx, listingID); err != nil {
		fmt.Printf("Warning: failed to expire bids for listing %s: %v\n", listingID, err)
	}

	// Save updated listing
	if err := s.listingRepo.Update(ctx, listing); err != nil {
		return nil, fmt.Errorf("failed to save listing: %w", err)
	}

	// Create order opportunity if there's a winner
	if result.WinningBid != nil && s.orderService != nil {
		orderResult, err := s.CreateOrderFromWinningBid(ctx, result.WinningBid, listing)
		if err != nil {
			fmt.Printf("Warning: failed to create order for winning bid %s: %v\n", result.WinningBid.BidID, err)
		} else {
			result.OrderCreated = true
			result.OrderID = orderResult.OrderID
		}
	}

	// Release inventory
	if s.inventoryService != nil {
		_ = s.inventoryService.ReleaseInventory(ctx, listing.ProductID, listing.Quantity, listing.ListingID)
	}

	// Record auction closure event
	if s.eventService != nil {
		eventData := map[string]interface{}{
			"listing_id":     listingID,
			"close_reason":   closeReason,
			"manual_close":   true,
			"total_bids":     result.TotalBids,
			"unique_bidders": result.UniqueBidders,
		}
		if result.WinningBid != nil {
			eventData["winning_bid_id"] = result.WinningBid.BidID
			eventData["winning_amount"] = result.WinningBid.BidAmount
			eventData["winner_id"] = result.WinningBid.BidderID
		}
		_ = s.eventService.RecordListingEvent(ctx, listingID, marketplace.EventListingClosed, eventData, "system")
	}

	return result, nil
}

// DetermineWinner determines the winning bid for an auction
func (s *AuctionLifecycleService) DetermineWinner(ctx context.Context, listingID string) (*marketplace.Bid, error) {
	// Get the highest bid
	highestBid, err := s.bidRepo.GetHighestBid(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get highest bid: %w", err)
	}

	if highestBid == nil {
		return nil, nil // No bids
	}

	// Validate that the highest bid is still active
	if highestBid.Status != marketplace.BidStatusActive && highestBid.Status != marketplace.BidStatusWinning {
		// Get the next highest active bid
		ranking, err := s.bidRepo.GetBidRanking(ctx, listingID, 10)
		if err != nil {
			return nil, fmt.Errorf("failed to get bid ranking: %w", err)
		}

		for _, bid := range ranking {
			if bid.Status == marketplace.BidStatusActive {
				return bid, nil
			}
		}

		return nil, nil // No active bids found
	}

	return highestBid, nil
}

// CreateOrderFromWinningBid creates an order opportunity for the winning bidder
func (s *AuctionLifecycleService) CreateOrderFromWinningBid(ctx context.Context, winningBid *marketplace.Bid, listing *marketplace.Listing) (*OrderCreationResult, error) {
	if s.orderService == nil {
		return nil, fmt.Errorf("order service not available")
	}

	// Create order request
	orderReq := &CreateOrderFromAuctionRequest{
		ListingID:    listing.ListingID,
		WinningBidID: winningBid.BidID,
		BuyerID:      winningBid.BidderID,
		SellerID:     listing.SellerID,
		ProductID:    listing.ProductID,
		Quantity:     winningBid.Quantity,
		Price:        winningBid.BidAmount,
		Currency:     winningBid.Currency,
	}

	// Create the order
	orderResult, err := s.orderService.CreateOrderFromAuction(ctx, orderReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create order from auction: %w", err)
	}

	return orderResult, nil
}

// StartAuctionExpiryScheduler starts the scheduled job for processing expired auctions
func (s *AuctionLifecycleService) StartAuctionExpiryScheduler(ctx context.Context, interval time.Duration) error {
	if s.schedulerRunning {
		return fmt.Errorf("auction expiry scheduler is already running")
	}

	s.schedulerRunning = true

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Process expired auctions
				result, err := s.ProcessExpiredAuctions(ctx)
				if err != nil {
					fmt.Printf("Error processing expired auctions: %v\n", err)
				} else if result.ProcessedCount > 0 {
					fmt.Printf("Processed %d expired auctions: %d successful, %d failed\n",
						result.ProcessedCount, result.SuccessfulCount, result.FailedCount)
				}

			case <-s.schedulerStop:
				fmt.Println("Auction expiry scheduler stopped")
				return
			}
		}
	}()

	fmt.Printf("Auction expiry scheduler started with interval: %v\n", interval)
	return nil
}

// StopAuctionExpiryScheduler stops the scheduled job for processing expired auctions
func (s *AuctionLifecycleService) StopAuctionExpiryScheduler() error {
	if !s.schedulerRunning {
		return fmt.Errorf("auction expiry scheduler is not running")
	}

	s.schedulerStop <- true
	s.schedulerRunning = false
	return nil
}
