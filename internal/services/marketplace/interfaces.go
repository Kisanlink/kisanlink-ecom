package marketplace

import (
	"context"
	"time"

	"kisanlink-ecom/entities/models/marketplace"
	"kisanlink-ecom/internal/common"

	"github.com/shopspring/decimal"
)

// Repository Interfaces

// BidRepositoryInterface defines the interface for bid data access
type BidRepositoryInterface interface {
	// Basic CRUD operations
	Create(ctx context.Context, bid *marketplace.Bid) error
	GetByID(ctx context.Context, id string) (*marketplace.Bid, error)
	GetByBidID(ctx context.Context, bidID string) (*marketplace.Bid, error)
	Update(ctx context.Context, bid *marketplace.Bid) error
	Delete(ctx context.Context, id string) error

	// Bid queries
	GetBidHistory(ctx context.Context, listingID string, bidVisibility marketplace.BidVisibility, viewerID string, pagination *common.PaginationRequest) ([]*marketplace.Bid, int, error)
	GetUserBids(ctx context.Context, userID string, filter *marketplace.BidFilter, pagination *common.PaginationRequest) ([]*marketplace.Bid, int, error)
	GetAllBids(ctx context.Context, filter *marketplace.BidFilter, pagination *common.PaginationRequest) ([]*marketplace.Bid, int, error)
	GetHighestBid(ctx context.Context, listingID string) (*marketplace.Bid, error)
	GetBidStatistics(ctx context.Context, listingID string) (*marketplace.BidStatistics, error)

	// Auto-bidding operations
	GetUserAutoBids(ctx context.Context, userID string, listingID string) ([]*marketplace.Bid, error)
	GetAutoBidsForListing(ctx context.Context, listingID string) ([]*marketplace.Bid, error)
	UpdateAutoBidStatus(ctx context.Context, bidID string, status marketplace.BidStatus) error

	// Atomic operations
	PlaceBidAtomic(ctx context.Context, bid *marketplace.Bid, listingID string) (*marketplace.Bid, error)
	RemoveBid(ctx context.Context, bidID string, reason string, adminID string) error
	ExpireBidsForListing(ctx context.Context, listingID string) error
	GetBidRanking(ctx context.Context, listingID string, limit int) ([]*marketplace.Bid, error)
}

// ListingRepositoryInterface defines the interface for listing data access
type ListingRepositoryInterface interface {
	// Basic CRUD operations
	Create(ctx context.Context, listing *marketplace.Listing) error
	GetByID(ctx context.Context, id string) (*marketplace.Listing, error)
	GetByListingID(ctx context.Context, listingID string) (*marketplace.Listing, error)
	Update(ctx context.Context, listing *marketplace.Listing) error
	Delete(ctx context.Context, id string) error

	// Listing queries
	GetActiveListings(ctx context.Context, viewerOrgID string, filter *marketplace.ListingFilter, pagination *common.PaginationRequest) ([]*marketplace.Listing, int, error)
	GetExpiredListings(ctx context.Context, limit int) ([]*marketplace.Listing, error)
	GetUserListings(ctx context.Context, userID string, filter *marketplace.ListingFilter, pagination *common.PaginationRequest) ([]*marketplace.Listing, int, error)
	GetAllListings(ctx context.Context, filter *marketplace.ListingFilter, pagination *common.PaginationRequest) ([]*marketplace.Listing, int, error)

	// Listing updates
	IncrementBidCount(ctx context.Context, listingID string) error
	UpdateHighestBid(ctx context.Context, listingID string, bidID string) error
	UpdateStatus(ctx context.Context, listingID string, status marketplace.ListingStatus, reason string) error
}

// AuctionEventRepositoryInterface defines the interface for auction event data access
type AuctionEventRepositoryInterface interface {
	Create(ctx context.Context, event *marketplace.AuctionEvent) error
	GetByID(ctx context.Context, id string) (*marketplace.AuctionEvent, error)
	GetByListingID(ctx context.Context, listingID string, pagination *common.PaginationRequest) ([]*marketplace.AuctionEvent, int, error)
	GetByEventType(ctx context.Context, eventType marketplace.AuctionEventType, pagination *common.PaginationRequest) ([]*marketplace.AuctionEvent, int, error)
}

// Service Interfaces

// EventServiceInterface defines the interface for event operations
type EventServiceInterface interface {
	RecordListingEvent(ctx context.Context, listingID string, eventType marketplace.AuctionEventType, eventData map[string]interface{}, actorID string) error
	GetListingEvents(ctx context.Context, listingID string, pagination *common.PaginationRequest) ([]*marketplace.AuctionEvent, int, error)
}

// MarketplaceAccessControlInterface defines the interface for marketplace access control operations
type MarketplaceAccessControlInterface interface {
	CanViewListing(ctx context.Context, listing *marketplace.Listing, viewerID string, viewerOrgID string) bool
	CanBidOnListing(ctx context.Context, listing *marketplace.Listing, bidderID string, bidderOrgID string) bool
	CanManageListing(ctx context.Context, listing *marketplace.Listing, userID string, userOrgID string) bool
	IsInvitedParticipant(ctx context.Context, listingID string, userID string) bool
	IsNetworkMember(ctx context.Context, organizationID string, userID string) bool
	IsSameOrganization(ctx context.Context, organizationID string, userID string) bool
}

// InventoryServiceInterface defines the interface for inventory operations
type InventoryServiceInterface interface {
	CheckAvailability(ctx context.Context, productID string, quantity decimal.Decimal, organizationID string) (bool, error)
	ReserveInventory(ctx context.Context, productID string, quantity decimal.Decimal, organizationID string, reservationID string) error
	ReleaseReservation(ctx context.Context, reservationID string) error
	ReleaseInventory(ctx context.Context, productID string, quantity decimal.Decimal, reservationID string) error
}

// External Service Interfaces (to avoid circular dependencies)

// Request/Response Types

// CreateListingRequest represents the request to create a new listing
type CreateListingRequest struct {
	ProductID       string                        `json:"product_id" validate:"required"`
	SellerID        string                        `json:"seller_id" validate:"required"`
	OrganizationID  string                        `json:"organization_id" validate:"required"`
	Quantity        decimal.Decimal               `json:"quantity" validate:"required,gt=0"`
	AskingPrice     decimal.Decimal               `json:"asking_price" validate:"required,gt=0"`
	MinimumBid      decimal.Decimal               `json:"minimum_bid" validate:"required,gt=0"`
	Currency        string                        `json:"currency,omitempty"`
	ListingDuration int                           `json:"listing_duration_hours" validate:"required,min=1,max=168"`
	Visibility      marketplace.ListingVisibility `json:"visibility,omitempty"`
	AuctionType     marketplace.AuctionType       `json:"auction_type,omitempty"`
	BidVisibility   marketplace.BidVisibility     `json:"bid_visibility,omitempty"`
	PickupLocation  *marketplace.Location         `json:"pickup_location,omitempty"`
	TermsConditions string                        `json:"terms_conditions,omitempty"`
}

// UpdateListingRequest represents the request to update a listing
type UpdateListingRequest struct {
	AskingPrice     *decimal.Decimal               `json:"asking_price,omitempty"`
	MinimumBid      *decimal.Decimal               `json:"minimum_bid,omitempty"`
	ListingDuration *int                           `json:"listing_duration_hours,omitempty"`
	Visibility      *marketplace.ListingVisibility `json:"visibility,omitempty"`
	AuctionType     *marketplace.AuctionType       `json:"auction_type,omitempty"`
	BidVisibility   *marketplace.BidVisibility     `json:"bid_visibility,omitempty"`
	PickupLocation  *marketplace.Location          `json:"pickup_location,omitempty"`
	TermsConditions *string                        `json:"terms_conditions,omitempty"`
}

// AuctionCloseResult represents the result of closing an auction
type AuctionCloseResult struct {
	Listing     *marketplace.Listing `json:"listing"`
	WinningBid  *marketplace.Bid     `json:"winning_bid,omitempty"`
	OrderID     string               `json:"order_id,omitempty"`
	CloseReason string               `json:"close_reason"`
	ClosedAt    time.Time            `json:"closed_at"`
}

// Analytics Types

// AnalyticsFilter represents filters for analytics queries
type AnalyticsFilter struct {
	StartDate      *time.Time `json:"start_date,omitempty"`
	EndDate        *time.Time `json:"end_date,omitempty"`
	OrganizationID string     `json:"organization_id,omitempty"`
	ProductID      string     `json:"product_id,omitempty"`
	UserID         string     `json:"user_id,omitempty"`
}

// MarketplaceMetrics represents overall marketplace metrics
type MarketplaceMetrics struct {
	TotalListings     int             `json:"total_listings"`
	ActiveListings    int             `json:"active_listings"`
	TotalBids         int             `json:"total_bids"`
	TotalVolume       decimal.Decimal `json:"total_volume"`
	AveragePrice      decimal.Decimal `json:"average_price"`
	CompletionRate    float64         `json:"completion_rate"`
	ParticipationRate float64         `json:"participation_rate"`
}

// ListingAnalytics represents analytics for a specific listing
type ListingAnalytics struct {
	ListingID        string          `json:"listing_id"`
	ViewCount        int             `json:"view_count"`
	BidCount         int             `json:"bid_count"`
	UniqueBidders    int             `json:"unique_bidders"`
	HighestBid       decimal.Decimal `json:"highest_bid"`
	AverageBid       decimal.Decimal `json:"average_bid"`
	BidProgression   []BidPoint      `json:"bid_progression"`
	CompetitionLevel string          `json:"competition_level"`
}

// BiddingAnalytics represents bidding analytics for a user
type BiddingAnalytics struct {
	UserID             string          `json:"user_id"`
	TotalBids          int             `json:"total_bids"`
	WinningBids        int             `json:"winning_bids"`
	WinRate            float64         `json:"win_rate"`
	TotalSpent         decimal.Decimal `json:"total_spent"`
	AverageBidAmount   decimal.Decimal `json:"average_bid_amount"`
	FavoriteCategories []string        `json:"favorite_categories"`
}

// OrganizationPerformanceMetrics represents performance metrics for an organization
type OrganizationPerformanceMetrics struct {
	OrganizationID       string          `json:"organization_id"`
	ListingsCreated      int             `json:"listings_created"`
	ListingsCompleted    int             `json:"listings_completed"`
	TotalRevenue         decimal.Decimal `json:"total_revenue"`
	AveragePrice         decimal.Decimal `json:"average_price"`
	CustomerSatisfaction float64         `json:"customer_satisfaction"`
}

// BidPoint represents a point in bid progression
type BidPoint struct {
	Timestamp time.Time       `json:"timestamp"`
	BidAmount decimal.Decimal `json:"bid_amount"`
	BidderID  string          `json:"bidder_id"`
}

// AuditFilter represents filters for audit queries
type AuditFilter struct {
	EventType *marketplace.AuctionEventType `json:"event_type,omitempty"`
	ActorID   string                        `json:"actor_id,omitempty"`
	StartDate *time.Time                    `json:"start_date,omitempty"`
	EndDate   *time.Time                    `json:"end_date,omitempty"`
}
