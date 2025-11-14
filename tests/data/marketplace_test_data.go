package data

import (
	"time"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/marketplace"

	"github.com/shopspring/decimal"
)

// MarketplaceTestData provides test data for marketplace entities
type MarketplaceTestData struct {
	Listings []marketplace.Listing
	Bids     []marketplace.Bid
	Events   []marketplace.AuctionEvent
}

// GetMarketplaceTestData returns a set of test data for marketplace testing
func GetMarketplaceTestData() *MarketplaceTestData {
	// Create test listings
	listing1 := *marketplace.NewListing(
		"PROD_WHEAT_001",
		"SELLER_FARMER_001",
		"ORG_FARM_COOP_001",
		decimal.NewFromFloat(1000.0),   // 1000 kg
		decimal.NewFromFloat(25000.00), // ₹25,000 asking price
		decimal.NewFromFloat(20000.00), // ₹20,000 minimum bid
		72,                             // 3 days
	)
	listing1.Visibility = marketplace.VisibilityPublic
	listing1.AuctionType = marketplace.AuctionTypeOpen
	listing1.BidVisibility = marketplace.BidVisibilityFull

	listing2 := *marketplace.NewListing(
		"PROD_RICE_002",
		"SELLER_FARMER_002",
		"ORG_FARM_COOP_002",
		decimal.NewFromFloat(500.0),    // 500 kg
		decimal.NewFromFloat(15000.00), // ₹15,000 asking price
		decimal.NewFromFloat(12000.00), // ₹12,000 minimum bid
		48,                             // 2 days
	)
	listing2.Visibility = marketplace.VisibilityOrganization
	listing2.AuctionType = marketplace.AuctionTypeClosed
	listing2.BidVisibility = marketplace.BidVisibilityPartial

	listing3 := *marketplace.NewListing(
		"SERV_TRACTOR_001",
		"SELLER_SERVICE_001",
		"ORG_AGRI_SERVICES_001",
		decimal.NewFromFloat(8.0),     // 8 hours
		decimal.NewFromFloat(4000.00), // ₹4,000 asking price
		decimal.NewFromFloat(3000.00), // ₹3,000 minimum bid
		24,                            // 1 day
	)
	listing3.Visibility = marketplace.VisibilityNetwork
	listing3.AuctionType = marketplace.AuctionTypeOpen
	listing3.BidVisibility = marketplace.BidVisibilityMinimal

	// Create test bids
	bid1 := *marketplace.NewBid(
		listing1.ListingID,
		"BUYER_TRADER_001",
		decimal.NewFromFloat(21000.00),
		decimal.NewFromFloat(1000.0),
		"Interested in bulk purchase",
	)

	bid2 := *marketplace.NewBid(
		listing1.ListingID,
		"BUYER_TRADER_002",
		decimal.NewFromFloat(22500.00),
		decimal.NewFromFloat(1000.0),
		"Can pick up immediately",
	)
	bid2.SetAsHighestBid()

	// Auto-bid example
	bid3 := *marketplace.NewAutoBid(
		listing1.ListingID,
		"BUYER_TRADER_003",
		decimal.NewFromFloat(23000.00),
		decimal.NewFromFloat(1000.0),
		decimal.NewFromFloat(26000.00), // Auto-bid limit
		bid1.BidID,
	)

	bid4 := *marketplace.NewBid(
		listing2.ListingID,
		"BUYER_MILL_001",
		decimal.NewFromFloat(13500.00),
		decimal.NewFromFloat(500.0),
		"Quality rice needed",
	)
	bid4.SetAsHighestBid()

	bid5 := *marketplace.NewBid(
		listing3.ListingID,
		"BUYER_FARMER_001",
		decimal.NewFromFloat(3200.00),
		decimal.NewFromFloat(8.0),
		"Need tractor service urgently",
	)
	bid5.SetAsHighestBid()

	// Create test events
	event1 := *marketplace.NewAuctionEvent(
		listing1.ListingID,
		marketplace.EventListingCreated,
		"SELLER_FARMER_001",
		marketplace.ActorTypeUser,
		nil,
	)

	event2 := *marketplace.NewAuctionEvent(
		listing1.ListingID,
		marketplace.EventBidPlaced,
		"BUYER_TRADER_001",
		marketplace.ActorTypeUser,
		nil,
	)

	event3 := *marketplace.NewAuctionEvent(
		listing1.ListingID,
		marketplace.EventBidPlaced,
		"BUYER_TRADER_002",
		marketplace.ActorTypeUser,
		nil,
	)

	event4 := *marketplace.NewAuctionEvent(
		listing1.ListingID,
		marketplace.EventAutoBidTriggered,
		"SYSTEM",
		marketplace.ActorTypeSystem,
		nil,
	)

	return &MarketplaceTestData{
		Listings: []marketplace.Listing{listing1, listing2, listing3},
		Bids:     []marketplace.Bid{bid1, bid2, bid3, bid4, bid5},
		Events:   []marketplace.AuctionEvent{event1, event2, event3, event4},
	}
}

// GetActiveListings returns only active listings from test data
func (mtd *MarketplaceTestData) GetActiveListings() []marketplace.Listing {
	var activeListings []marketplace.Listing
	for _, listing := range mtd.Listings {
		if listing.Status == marketplace.ListingStatusActive {
			activeListings = append(activeListings, listing)
		}
	}
	return activeListings
}

// GetBidsForListing returns bids for a specific listing
func (mtd *MarketplaceTestData) GetBidsForListing(listingID string) []marketplace.Bid {
	var bids []marketplace.Bid
	for _, bid := range mtd.Bids {
		if bid.ListingID == listingID {
			bids = append(bids, bid)
		}
	}
	return bids
}

// GetEventsForListing returns events for a specific listing
func (mtd *MarketplaceTestData) GetEventsForListing(listingID string) []marketplace.AuctionEvent {
	var events []marketplace.AuctionEvent
	for _, event := range mtd.Events {
		if event.ListingID == listingID {
			events = append(events, event)
		}
	}
	return events
}

// CreateExpiredListing creates a listing that has already expired
func CreateExpiredListing() *marketplace.Listing {
	listing := marketplace.NewListing(
		"PROD_EXPIRED_001",
		"SELLER_EXPIRED_001",
		"ORG_EXPIRED_001",
		decimal.NewFromFloat(100.0),
		decimal.NewFromFloat(5000.00),
		decimal.NewFromFloat(4000.00),
		1, // 1 hour duration
	)

	// Set expiry to past
	listing.ExpiresAt = time.Now().Add(-2 * time.Hour)
	listing.Status = marketplace.ListingStatusExpired
	now := time.Now()
	listing.ClosedAt = &now
	reason := "Auction expired"
	listing.CloseReason = &reason

	return listing
}

// CreateClosedListing creates a listing that has been manually closed
func CreateClosedListing() *marketplace.Listing {
	listing := marketplace.NewListing(
		"PROD_CLOSED_001",
		"SELLER_CLOSED_001",
		"ORG_CLOSED_001",
		decimal.NewFromFloat(200.0),
		decimal.NewFromFloat(10000.00),
		decimal.NewFromFloat(8000.00),
		24,
	)

	listing.Status = marketplace.ListingStatusClosed
	now := time.Now()
	listing.ClosedAt = &now
	reason := "Seller closed early"
	listing.CloseReason = &reason

	return listing
}

// CreatePrivateListing creates a private listing for testing visibility
func CreatePrivateListing() *marketplace.Listing {
	listing := marketplace.NewListing(
		"PROD_PRIVATE_001",
		"SELLER_PRIVATE_001",
		"ORG_PRIVATE_001",
		decimal.NewFromFloat(50.0),
		decimal.NewFromFloat(2500.00),
		decimal.NewFromFloat(2000.00),
		12,
	)

	listing.Visibility = marketplace.VisibilityPrivate
	listing.AuctionType = marketplace.AuctionTypeClosed
	listing.BidVisibility = marketplace.BidVisibilityHidden

	return listing
}

// CreateHighValueListing creates a high-value listing for testing
func CreateHighValueListing() *marketplace.Listing {
	listing := marketplace.NewListing(
		"PROD_HIGH_VALUE_001",
		"SELLER_PREMIUM_001",
		"ORG_PREMIUM_001",
		decimal.NewFromFloat(10000.0),   // 10 tons
		decimal.NewFromFloat(500000.00), // ₹5,00,000 asking price
		decimal.NewFromFloat(400000.00), // ₹4,00,000 minimum bid
		168,                             // 1 week
	)

	listing.Visibility = marketplace.VisibilityNetwork
	listing.AuctionType = marketplace.AuctionTypeOpen
	listing.BidVisibility = marketplace.BidVisibilityFull

	return listing
}

// MarketplaceScenarios provides different test scenarios
type MarketplaceScenarios struct{}

// GetCompetitiveBiddingScenario returns data for testing competitive bidding
func (ms *MarketplaceScenarios) GetCompetitiveBiddingScenario() (*marketplace.Listing, []marketplace.Bid) {
	listing := marketplace.NewListing(
		"PROD_COMPETITIVE_001",
		"SELLER_COMPETITIVE_001",
		"ORG_COMPETITIVE_001",
		decimal.NewFromFloat(1000.0),
		decimal.NewFromFloat(30000.00),
		decimal.NewFromFloat(25000.00),
		48,
	)

	bids := []marketplace.Bid{
		*marketplace.NewBid(listing.ListingID, "BUYER_001", decimal.NewFromFloat(25500.00), decimal.NewFromFloat(1000.0), "Initial bid"),
		*marketplace.NewBid(listing.ListingID, "BUYER_002", decimal.NewFromFloat(26000.00), decimal.NewFromFloat(1000.0), "Counter bid"),
		*marketplace.NewBid(listing.ListingID, "BUYER_003", decimal.NewFromFloat(26500.00), decimal.NewFromFloat(1000.0), "Higher bid"),
		*marketplace.NewBid(listing.ListingID, "BUYER_001", decimal.NewFromFloat(27000.00), decimal.NewFromFloat(1000.0), "Competitive response"),
		*marketplace.NewBid(listing.ListingID, "BUYER_004", decimal.NewFromFloat(28000.00), decimal.NewFromFloat(1000.0), "Final bid"),
	}

	// Set the last bid as highest
	bids[len(bids)-1].SetAsHighestBid()

	return listing, bids
}

// GetAutoBiddingScenario returns data for testing auto-bidding functionality
func (ms *MarketplaceScenarios) GetAutoBiddingScenario() (*marketplace.Listing, []marketplace.Bid) {
	listing := marketplace.NewListing(
		"PROD_AUTO_BID_001",
		"SELLER_AUTO_001",
		"ORG_AUTO_001",
		decimal.NewFromFloat(500.0),
		decimal.NewFromFloat(20000.00),
		decimal.NewFromFloat(15000.00),
		24,
	)

	initialBid := marketplace.NewBid(listing.ListingID, "BUYER_MANUAL_001", decimal.NewFromFloat(15500.00), decimal.NewFromFloat(500.0), "Manual bid")

	autoBid1 := marketplace.NewAutoBid(listing.ListingID, "BUYER_AUTO_001", decimal.NewFromFloat(16000.00), decimal.NewFromFloat(500.0), decimal.NewFromFloat(19000.00), initialBid.BidID)

	competingBid := marketplace.NewBid(listing.ListingID, "BUYER_MANUAL_002", decimal.NewFromFloat(17000.00), decimal.NewFromFloat(500.0), "Competing bid")

	// Auto-bid response
	autoBid2 := marketplace.NewAutoBid(listing.ListingID, "BUYER_AUTO_001", decimal.NewFromFloat(17500.00), decimal.NewFromFloat(500.0), decimal.NewFromFloat(19000.00), autoBid1.BidID)
	autoBid2.SetAsHighestBid()

	bids := []marketplace.Bid{*initialBid, *autoBid1, *competingBid, *autoBid2}

	return listing, bids
}
