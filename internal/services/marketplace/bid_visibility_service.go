package marketplace

import (
	"context"
	"fmt"
	"time"

	"kisanlink-ecom/entities/models/marketplace"
	"kisanlink-ecom/internal/common"
)

// BidVisibilityServiceInterface defines the interface for bid visibility operations
type BidVisibilityServiceInterface interface {
	// Visibility Rules Engine
	ApplyBidVisibilityRules(ctx context.Context, listing *marketplace.Listing, bids []*marketplace.Bid, viewerID string, viewerOrgID string) (*FilteredBidHistory, error)
	GetVisibleBidHistory(ctx context.Context, listingID string, viewerID string, viewerOrgID string, pagination *common.PaginationParams) (*FilteredBidHistory, error)
	GetVisibleBidStatistics(ctx context.Context, listingID string, viewerID string, viewerOrgID string) (*FilteredBidStatistics, error)

	// Real-time Filtering
	FilterBidForRealTimeUpdate(ctx context.Context, listing *marketplace.Listing, bid *marketplace.Bid, viewerID string, viewerOrgID string) (*FilteredBid, error)
	FilterBidsForListing(ctx context.Context, listing *marketplace.Listing, bids []*marketplace.Bid, viewerID string, viewerOrgID string) ([]*FilteredBid, error)

	// Access Control
	CanViewBidDetails(ctx context.Context, listing *marketplace.Listing, bid *marketplace.Bid, viewerID string, viewerOrgID string) bool
	CanViewListing(ctx context.Context, listing *marketplace.Listing, viewerID string, viewerOrgID string) bool

	// Auction State Visibility
	GetAuctionVisibilityState(ctx context.Context, listing *marketplace.Listing, viewerID string, viewerOrgID string) (*AuctionVisibilityState, error)
}

// FilteredBidHistory represents bid history with applied visibility rules
type FilteredBidHistory struct {
	ListingID       string                 `json:"listing_id"`
	TotalBids       int                    `json:"total_bids"`
	VisibleBids     int                    `json:"visible_bids"`
	HighestBid      *FilteredBid           `json:"highest_bid,omitempty"`
	Bids            []*FilteredBid         `json:"bids"`
	VisibilityLevel BidVisibilityLevel     `json:"visibility_level"`
	AuctionState    AuctionVisibilityState `json:"auction_state"`
	LastUpdated     time.Time              `json:"last_updated"`
}

// FilteredBid represents a bid with applied visibility filtering
type FilteredBid struct {
	BidID           string                 `json:"bid_id,omitempty"`
	BidAmount       *string                `json:"bid_amount,omitempty"` // Pointer to allow null for hidden amounts
	AnonymousID     string                 `json:"anonymous_id,omitempty"`
	PlacedAt        *time.Time             `json:"placed_at,omitempty"`
	IsHighestBid    *bool                  `json:"is_highest_bid,omitempty"`
	IsAutoBid       *bool                  `json:"is_auto_bid,omitempty"`
	Status          *marketplace.BidStatus `json:"status,omitempty"`
	IsOwnBid        bool                   `json:"is_own_bid"`
	VisibilityLevel BidVisibilityLevel     `json:"visibility_level"`
}

// FilteredBidStatistics represents bid statistics with applied visibility rules
type FilteredBidStatistics struct {
	ListingID       string                 `json:"listing_id"`
	TotalBids       int                    `json:"total_bids"`
	UniqueBidders   *int                   `json:"unique_bidders,omitempty"`
	HighestBid      *string                `json:"highest_bid,omitempty"`
	AverageBid      *string                `json:"average_bid,omitempty"`
	LastBidTime     *time.Time             `json:"last_bid_time,omitempty"`
	AutoBidCount    *int                   `json:"auto_bid_count,omitempty"`
	VisibilityLevel BidVisibilityLevel     `json:"visibility_level"`
	AuctionState    AuctionVisibilityState `json:"auction_state"`
}

// BidVisibilityLevel represents the effective visibility level for a viewer
type BidVisibilityLevel string

const (
	VisibilityLevelFull    BidVisibilityLevel = "FULL"    // All bid details visible
	VisibilityLevelPartial BidVisibilityLevel = "PARTIAL" // Limited bid details
	VisibilityLevelMinimal BidVisibilityLevel = "MINIMAL" // Only bid count
	VisibilityLevelHidden  BidVisibilityLevel = "HIDDEN"  // No bid information
	VisibilityLevelOwn     BidVisibilityLevel = "OWN"     // Own bids always visible
)

// AuctionVisibilityState represents the current state of auction visibility
type AuctionVisibilityState struct {
	IsActive      bool                      `json:"is_active"`
	IsExpired     bool                      `json:"is_expired"`
	AuctionType   marketplace.AuctionType   `json:"auction_type"`
	BidVisibility marketplace.BidVisibility `json:"bid_visibility"`
	TimeRemaining *time.Duration            `json:"time_remaining,omitempty"`
	CanRevealBids bool                      `json:"can_reveal_bids"`
	RevealReason  string                    `json:"reveal_reason,omitempty"`
}

// BidVisibilityService provides business logic for bid visibility and transparency
type BidVisibilityService struct {
	bidRepo       BidRepositoryInterface
	listingRepo   ListingRepositoryInterface
	accessControl AccessControlServiceInterface
	eventService  EventServiceInterface
}

// NewBidVisibilityService creates a new bid visibility service
func NewBidVisibilityService(
	bidRepo BidRepositoryInterface,
	listingRepo ListingRepositoryInterface,
	accessControl AccessControlServiceInterface,
	eventService EventServiceInterface,
) BidVisibilityServiceInterface {
	return &BidVisibilityService{
		bidRepo:       bidRepo,
		listingRepo:   listingRepo,
		accessControl: accessControl,
		eventService:  eventService,
	}
}

// ApplyBidVisibilityRules applies visibility rules to a set of bids based on auction configuration
func (s *BidVisibilityService) ApplyBidVisibilityRules(ctx context.Context, listing *marketplace.Listing, bids []*marketplace.Bid, viewerID string, viewerOrgID string) (*FilteredBidHistory, error) {
	// Check if viewer can access this listing
	if !s.CanViewListing(ctx, listing, viewerID, viewerOrgID) {
		return nil, fmt.Errorf("access denied: listing not visible to viewer")
	}

	// Get auction visibility state
	auctionState, err := s.GetAuctionVisibilityState(ctx, listing, viewerID, viewerOrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get auction visibility state: %w", err)
	}

	// Determine effective visibility level
	visibilityLevel := s.determineEffectiveVisibilityLevel(listing, auctionState)

	// Filter bids based on visibility rules
	filteredBids, err := s.FilterBidsForListing(ctx, listing, bids, viewerID, viewerOrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to filter bids: %w", err)
	}

	// Find highest bid
	var highestBid *FilteredBid
	for _, bid := range filteredBids {
		if bid.IsHighestBid != nil && *bid.IsHighestBid {
			highestBid = bid
			break
		}
	}

	// Build filtered history
	filteredHistory := &FilteredBidHistory{
		ListingID:       listing.ListingID,
		TotalBids:       listing.BidCount,
		VisibleBids:     len(filteredBids),
		HighestBid:      highestBid,
		Bids:            filteredBids,
		VisibilityLevel: visibilityLevel,
		AuctionState:    *auctionState,
		LastUpdated:     time.Now(),
	}

	return filteredHistory, nil
}

// GetVisibleBidHistory retrieves bid history with visibility filtering applied
func (s *BidVisibilityService) GetVisibleBidHistory(ctx context.Context, listingID string, viewerID string, viewerOrgID string, pagination *common.PaginationParams) (*FilteredBidHistory, error) {
	// Get listing
	listing, err := s.listingRepo.GetByListingID(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get listing: %w", err)
	}
	if listing == nil {
		return nil, fmt.Errorf("listing not found: %s", listingID)
	}

	// Get bids based on pagination
	paginationReq := &common.PaginationRequest{
		Limit:  pagination.Limit,
		Offset: pagination.CalculateOffset(),
	}

	bids, _, err := s.bidRepo.GetBidHistory(ctx, listingID, listing.BidVisibility, viewerID, paginationReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get bid history: %w", err)
	}

	// Apply visibility rules
	return s.ApplyBidVisibilityRules(ctx, listing, bids, viewerID, viewerOrgID)
}

// GetVisibleBidStatistics retrieves bid statistics with visibility filtering applied
func (s *BidVisibilityService) GetVisibleBidStatistics(ctx context.Context, listingID string, viewerID string, viewerOrgID string) (*FilteredBidStatistics, error) {
	// Get listing
	listing, err := s.listingRepo.GetByListingID(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get listing: %w", err)
	}
	if listing == nil {
		return nil, fmt.Errorf("listing not found: %s", listingID)
	}

	// Check access
	if !s.CanViewListing(ctx, listing, viewerID, viewerOrgID) {
		return nil, fmt.Errorf("access denied: listing not visible to viewer")
	}

	// Get auction state
	auctionState, err := s.GetAuctionVisibilityState(ctx, listing, viewerID, viewerOrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get auction visibility state: %w", err)
	}

	// Get raw statistics
	rawStats, err := s.bidRepo.GetBidStatistics(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bid statistics: %w", err)
	}

	// Apply visibility filtering to statistics
	visibilityLevel := s.determineEffectiveVisibilityLevel(listing, auctionState)
	filteredStats := s.filterBidStatistics(rawStats, visibilityLevel, auctionState)

	return filteredStats, nil
}

// FilterBidForRealTimeUpdate filters a single bid for real-time updates
func (s *BidVisibilityService) FilterBidForRealTimeUpdate(ctx context.Context, listing *marketplace.Listing, bid *marketplace.Bid, viewerID string, viewerOrgID string) (*FilteredBid, error) {
	// Check access
	if !s.CanViewListing(ctx, listing, viewerID, viewerOrgID) {
		return nil, fmt.Errorf("access denied: listing not visible to viewer")
	}

	// Get auction state
	auctionState, err := s.GetAuctionVisibilityState(ctx, listing, viewerID, viewerOrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get auction visibility state: %w", err)
	}

	// Filter the bid
	filteredBid := s.filterSingleBid(bid, listing, auctionState, viewerID)
	return filteredBid, nil
}

// FilterBidsForListing filters multiple bids for a listing
func (s *BidVisibilityService) FilterBidsForListing(ctx context.Context, listing *marketplace.Listing, bids []*marketplace.Bid, viewerID string, viewerOrgID string) ([]*FilteredBid, error) {
	// Get auction state
	auctionState, err := s.GetAuctionVisibilityState(ctx, listing, viewerID, viewerOrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get auction visibility state: %w", err)
	}

	// Filter each bid
	filteredBids := make([]*FilteredBid, 0, len(bids))
	for _, bid := range bids {
		filteredBid := s.filterSingleBid(bid, listing, auctionState, viewerID)
		if filteredBid != nil {
			filteredBids = append(filteredBids, filteredBid)
		}
	}

	return filteredBids, nil
}

// CanViewBidDetails checks if a viewer can see detailed bid information
func (s *BidVisibilityService) CanViewBidDetails(ctx context.Context, listing *marketplace.Listing, bid *marketplace.Bid, viewerID string, viewerOrgID string) bool {
	// Always allow viewing own bids
	if bid.BidderID == viewerID {
		return true
	}

	// Check listing access
	if !s.CanViewListing(ctx, listing, viewerID, viewerOrgID) {
		return false
	}

	// Get auction state
	auctionState, err := s.GetAuctionVisibilityState(ctx, listing, viewerID, viewerOrgID)
	if err != nil {
		return false
	}

	// Apply visibility rules
	visibilityLevel := s.determineEffectiveVisibilityLevel(listing, auctionState)

	switch visibilityLevel {
	case VisibilityLevelFull:
		return true
	case VisibilityLevelPartial:
		return bid.IsHighestBid // Only show highest bid details
	case VisibilityLevelMinimal, VisibilityLevelHidden:
		return false
	default:
		return false
	}
}

// CanViewListing checks if a viewer can access a listing
func (s *BidVisibilityService) CanViewListing(ctx context.Context, listing *marketplace.Listing, viewerID string, viewerOrgID string) bool {
	if s.accessControl != nil {
		return s.accessControl.CanViewListing(ctx, listing, viewerID, viewerOrgID)
	}

	// Fallback to basic visibility check
	return s.basicVisibilityCheck(listing, viewerOrgID)
}

// GetAuctionVisibilityState determines the current visibility state of an auction
func (s *BidVisibilityService) GetAuctionVisibilityState(ctx context.Context, listing *marketplace.Listing, viewerID string, viewerOrgID string) (*AuctionVisibilityState, error) {
	now := time.Now()
	isActive := listing.Status == marketplace.ListingStatusActive && now.Before(listing.ExpiresAt)
	isExpired := now.After(listing.ExpiresAt) || listing.Status != marketplace.ListingStatusActive

	var timeRemaining *time.Duration
	if isActive {
		remaining := time.Until(listing.ExpiresAt)
		timeRemaining = &remaining
	}

	// Determine if bids can be revealed
	canRevealBids, revealReason := s.canRevealBids(listing, isActive, isExpired, viewerID)

	state := &AuctionVisibilityState{
		IsActive:      isActive,
		IsExpired:     isExpired,
		AuctionType:   listing.AuctionType,
		BidVisibility: listing.BidVisibility,
		TimeRemaining: timeRemaining,
		CanRevealBids: canRevealBids,
		RevealReason:  revealReason,
	}

	return state, nil
}

// Helper methods

// determineEffectiveVisibilityLevel determines the effective visibility level based on auction configuration and state
func (s *BidVisibilityService) determineEffectiveVisibilityLevel(listing *marketplace.Listing, auctionState *AuctionVisibilityState) BidVisibilityLevel {
	// If auction is closed/expired and bids can be revealed, use full visibility
	if auctionState.CanRevealBids {
		return VisibilityLevelFull
	}

	// For active auctions, apply configured visibility rules
	switch listing.AuctionType {
	case marketplace.AuctionTypeOpen:
		// Open auctions show bids based on bid visibility setting
		switch listing.BidVisibility {
		case marketplace.BidVisibilityFull:
			return VisibilityLevelFull
		case marketplace.BidVisibilityPartial:
			return VisibilityLevelPartial
		case marketplace.BidVisibilityMinimal:
			return VisibilityLevelMinimal
		case marketplace.BidVisibilityHidden:
			return VisibilityLevelHidden
		}
	case marketplace.AuctionTypeClosed:
		// Closed auctions hide bid amounts during active period
		switch listing.BidVisibility {
		case marketplace.BidVisibilityFull, marketplace.BidVisibilityPartial:
			return VisibilityLevelMinimal // Only show bid count
		case marketplace.BidVisibilityMinimal:
			return VisibilityLevelMinimal
		case marketplace.BidVisibilityHidden:
			return VisibilityLevelHidden
		}
	}

	return VisibilityLevelHidden
}

// filterSingleBid applies visibility filtering to a single bid
func (s *BidVisibilityService) filterSingleBid(bid *marketplace.Bid, listing *marketplace.Listing, auctionState *AuctionVisibilityState, viewerID string) *FilteredBid {
	isOwnBid := bid.BidderID == viewerID

	// Always show full details for own bids
	if isOwnBid {
		bidAmountStr := bid.BidAmount.String()
		return &FilteredBid{
			BidID:           bid.BidID,
			BidAmount:       &bidAmountStr,
			AnonymousID:     "Your Bid",
			PlacedAt:        &bid.PlacedAt,
			IsHighestBid:    &bid.IsHighestBid,
			IsAutoBid:       &bid.IsAutoBid,
			Status:          &bid.Status,
			IsOwnBid:        true,
			VisibilityLevel: VisibilityLevelOwn,
		}
	}

	// Apply visibility rules for other bids
	visibilityLevel := s.determineEffectiveVisibilityLevel(listing, auctionState)

	filteredBid := &FilteredBid{
		IsOwnBid:        false,
		VisibilityLevel: visibilityLevel,
		AnonymousID:     s.generateAnonymousID(bid.BidderID),
	}

	switch visibilityLevel {
	case VisibilityLevelFull:
		bidAmountStr := bid.BidAmount.String()
		filteredBid.BidID = bid.BidID
		filteredBid.BidAmount = &bidAmountStr
		filteredBid.PlacedAt = &bid.PlacedAt
		filteredBid.IsHighestBid = &bid.IsHighestBid
		filteredBid.IsAutoBid = &bid.IsAutoBid
		filteredBid.Status = &bid.Status

	case VisibilityLevelPartial:
		// Only show highest bid amount and basic info
		if bid.IsHighestBid {
			bidAmountStr := bid.BidAmount.String()
			filteredBid.BidAmount = &bidAmountStr
			filteredBid.IsHighestBid = &bid.IsHighestBid
		}
		filteredBid.PlacedAt = &bid.PlacedAt

	case VisibilityLevelMinimal:
		// Only show that a bid exists
		filteredBid.PlacedAt = &bid.PlacedAt

	case VisibilityLevelHidden:
		// Don't show this bid at all
		return nil
	}

	return filteredBid
}

// filterBidStatistics applies visibility filtering to bid statistics
func (s *BidVisibilityService) filterBidStatistics(rawStats *marketplace.BidStatistics, visibilityLevel BidVisibilityLevel, auctionState *AuctionVisibilityState) *FilteredBidStatistics {
	filteredStats := &FilteredBidStatistics{
		ListingID:       rawStats.ListingID,
		TotalBids:       rawStats.TotalBids,
		VisibilityLevel: visibilityLevel,
		AuctionState:    *auctionState,
	}

	switch visibilityLevel {
	case VisibilityLevelFull:
		uniqueBidders := rawStats.UniqueBidders
		highestBid := rawStats.HighestBid.String()
		averageBid := rawStats.AverageBid.String()
		autoBidCount := rawStats.AutoBidCount

		filteredStats.UniqueBidders = &uniqueBidders
		filteredStats.HighestBid = &highestBid
		filteredStats.AverageBid = &averageBid
		filteredStats.LastBidTime = rawStats.LastBidTime
		filteredStats.AutoBidCount = &autoBidCount

	case VisibilityLevelPartial:
		highestBid := rawStats.HighestBid.String()
		filteredStats.HighestBid = &highestBid
		filteredStats.LastBidTime = rawStats.LastBidTime

	case VisibilityLevelMinimal:
		// Only total bid count is shown (already set)

	case VisibilityLevelHidden:
		// Reset total bids to 0 for hidden visibility
		filteredStats.TotalBids = 0
	}

	return filteredStats
}

// canRevealBids determines if bid information can be revealed based on auction state
func (s *BidVisibilityService) canRevealBids(listing *marketplace.Listing, isActive bool, isExpired bool, viewerID string) (bool, string) {
	// Always reveal for listing owner
	if listing.SellerID == viewerID {
		return true, "listing_owner"
	}

	// Reveal when auction is closed/expired
	if isExpired || listing.Status != marketplace.ListingStatusActive {
		return true, "auction_ended"
	}

	// For active auctions, follow auction type rules
	if isActive && listing.AuctionType == marketplace.AuctionTypeOpen {
		return true, "open_auction"
	}

	return false, "auction_active"
}

// basicVisibilityCheck performs basic listing visibility check
func (s *BidVisibilityService) basicVisibilityCheck(listing *marketplace.Listing, viewerOrgID string) bool {
	switch listing.Visibility {
	case marketplace.VisibilityPrivate:
		return listing.OrganizationID == viewerOrgID
	case marketplace.VisibilityPublic:
		return true
	case marketplace.VisibilityNetwork:
		return true // Simplified - would need network logic
	case marketplace.VisibilityOrganization:
		return listing.OrganizationID == viewerOrgID
	default:
		return false
	}
}

// generateAnonymousID generates a consistent anonymous ID for a bidder
func (s *BidVisibilityService) generateAnonymousID(bidderID string) string {
	// Simple hash-based anonymization - in production, use proper anonymization
	hash := 0
	for _, char := range bidderID {
		hash = hash*31 + int(char)
	}
	if hash < 0 {
		hash = -hash
	}
	return fmt.Sprintf("Bidder #%d", (hash%1000)+1)
}
