package marketplace_test

import (
	"testing"
	"time"

	"kisanlink-ecom/entities/models/marketplace"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewListing(t *testing.T) {
	// Test data
	productID := "PROD_123"
	sellerID := "USER_456"
	organizationID := "ORG_789"
	quantity := decimal.NewFromFloat(100.5)
	askingPrice := decimal.NewFromFloat(1000.00)
	minimumBid := decimal.NewFromFloat(800.00)
	durationHours := 24

	// Create new listing
	listing := marketplace.NewListing(productID, sellerID, organizationID, quantity, askingPrice, minimumBid, durationHours)

	// Assertions
	assert.NotEmpty(t, listing.ID)
	assert.NotEmpty(t, listing.ListingID)
	assert.Equal(t, productID, listing.ProductID)
	assert.Equal(t, sellerID, listing.SellerID)
	assert.Equal(t, organizationID, listing.OrganizationID)
	assert.True(t, quantity.Equal(listing.Quantity))
	assert.True(t, askingPrice.Equal(listing.AskingPrice))
	assert.True(t, minimumBid.Equal(listing.MinimumBid))
	assert.Equal(t, "INR", listing.Currency)
	assert.Equal(t, durationHours, listing.ListingDuration)
	assert.Equal(t, marketplace.ListingStatusActive, listing.Status)
	assert.Equal(t, 0, listing.BidCount)
	assert.Equal(t, marketplace.VisibilityPublic, listing.Visibility)
	assert.Equal(t, marketplace.AuctionTypeOpen, listing.AuctionType)
	assert.Equal(t, marketplace.BidVisibilityFull, listing.BidVisibility)
	assert.Equal(t, "AUCTION", listing.ListingType)
	assert.True(t, listing.ExpiresAt.After(time.Now()))
}

func TestListingStatusTransitions(t *testing.T) {
	listing := marketplace.NewListing("PROD_123", "USER_456", "ORG_789",
		decimal.NewFromFloat(100), decimal.NewFromFloat(1000), decimal.NewFromFloat(800), 24)

	// Test valid transitions from ACTIVE
	assert.True(t, listing.CanTransitionTo(marketplace.ListingStatusClosed))
	assert.True(t, listing.CanTransitionTo(marketplace.ListingStatusExpired))
	assert.True(t, listing.CanTransitionTo(marketplace.ListingStatusCancelled))
	assert.True(t, listing.CanTransitionTo(marketplace.ListingStatusExpiredNoBids))

	// Test invalid transition
	assert.False(t, listing.CanTransitionTo(marketplace.ListingStatusActive))

	// Test status update
	err := listing.UpdateStatus(marketplace.ListingStatusClosed, "Manual closure")
	assert.NoError(t, err)
	assert.Equal(t, marketplace.ListingStatusClosed, listing.Status)
	assert.NotNil(t, listing.ClosedAt)
	assert.NotNil(t, listing.CloseReason)
	assert.Equal(t, "Manual closure", *listing.CloseReason)

	// Test terminal state - no further transitions allowed
	assert.False(t, listing.CanTransitionTo(marketplace.ListingStatusActive))
	assert.False(t, listing.CanTransitionTo(marketplace.ListingStatusExpired))
}

func TestListingActiveAndExpired(t *testing.T) {
	// Create listing that expires in 1 hour
	listing := marketplace.NewListing("PROD_123", "USER_456", "ORG_789",
		decimal.NewFromFloat(100), decimal.NewFromFloat(1000), decimal.NewFromFloat(800), 1)

	// Should be active initially
	assert.True(t, listing.IsActive())
	assert.False(t, listing.IsExpired())
	assert.True(t, listing.CanAcceptBids())

	// Simulate expiry by setting expires_at to past
	listing.ExpiresAt = time.Now().Add(-1 * time.Hour)

	// Should now be expired
	assert.False(t, listing.IsActive())
	assert.True(t, listing.IsExpired())
	assert.False(t, listing.CanAcceptBids())
}

func TestListingPickupLocation(t *testing.T) {
	listing := marketplace.NewListing("PROD_123", "USER_456", "ORG_789",
		decimal.NewFromFloat(100), decimal.NewFromFloat(1000), decimal.NewFromFloat(800), 24)

	// Test setting pickup location
	location := &marketplace.Location{
		Address:    "123 Farm Road",
		City:       "Bangalore",
		State:      "Karnataka",
		PostalCode: "560001",
		Country:    "India",
		Latitude:   func() *float64 { f := 12.9716; return &f }(),
		Longitude:  func() *float64 { f := 77.5946; return &f }(),
	}

	err := listing.SetPickupLocation(location)
	assert.NoError(t, err)
	assert.NotEmpty(t, listing.PickupLocation)

	// Test getting pickup location
	retrievedLocation, err := listing.GetPickupLocation()
	assert.NoError(t, err)
	assert.NotNil(t, retrievedLocation)
	assert.Equal(t, location.Address, retrievedLocation.Address)
	assert.Equal(t, location.City, retrievedLocation.City)
	assert.Equal(t, location.State, retrievedLocation.State)
	assert.Equal(t, *location.Latitude, *retrievedLocation.Latitude)
	assert.Equal(t, *location.Longitude, *retrievedLocation.Longitude)

	// Test setting nil location
	err = listing.SetPickupLocation(nil)
	assert.NoError(t, err)
	assert.Empty(t, listing.PickupLocation)

	retrievedLocation, err = listing.GetPickupLocation()
	assert.NoError(t, err)
	assert.Nil(t, retrievedLocation)
}

func TestListingVisibilityAccess(t *testing.T) {
	listing := marketplace.NewListing("PROD_123", "USER_456", "ORG_789",
		decimal.NewFromFloat(100), decimal.NewFromFloat(1000), decimal.NewFromFloat(800), 24)

	// Test public visibility
	listing.Visibility = marketplace.VisibilityPublic
	assert.True(t, listing.CanBeViewedBy("ANY_ORG"))
	assert.True(t, listing.CanBeViewedBy("ORG_789"))

	// Test organization visibility
	listing.Visibility = marketplace.VisibilityOrganization
	assert.True(t, listing.CanBeViewedBy("ORG_789"))
	assert.False(t, listing.CanBeViewedBy("OTHER_ORG"))

	// Test private visibility
	listing.Visibility = marketplace.VisibilityPrivate
	assert.True(t, listing.CanBeViewedBy("ORG_789"))
	assert.False(t, listing.CanBeViewedBy("OTHER_ORG"))

	// Test network visibility (simplified - always true for now)
	listing.Visibility = marketplace.VisibilityNetwork
	assert.True(t, listing.CanBeViewedBy("ANY_ORG"))
}

func TestNewBid(t *testing.T) {
	// Test data
	listingID := "LST_123"
	bidderID := "USER_789"
	bidAmount := decimal.NewFromFloat(900.00)
	quantity := decimal.NewFromFloat(50.0)
	message := "Interested in this product"

	// Create new bid
	bid := marketplace.NewBid(listingID, bidderID, bidAmount, quantity, message)

	// Assertions
	assert.NotEmpty(t, bid.ID)
	assert.NotEmpty(t, bid.BidID)
	assert.Equal(t, listingID, bid.ListingID)
	assert.Equal(t, bidderID, bid.BidderID)
	assert.True(t, bidAmount.Equal(bid.BidAmount))
	assert.Equal(t, "INR", bid.Currency)
	assert.True(t, quantity.Equal(bid.Quantity))
	assert.Equal(t, message, bid.Message)
	assert.False(t, bid.IsAutoBid)
	assert.Equal(t, marketplace.BidStatusActive, bid.Status)
	assert.False(t, bid.IsHighestBid)
	assert.True(t, bid.PlacedAt.Before(time.Now().Add(1*time.Second)))
}

func TestNewAutoBid(t *testing.T) {
	listingID := "LST_123"
	bidderID := "USER_789"
	bidAmount := decimal.NewFromFloat(900.00)
	quantity := decimal.NewFromFloat(50.0)
	autoBidLimit := decimal.NewFromFloat(1200.00)
	parentBidID := "BID_PARENT_123"

	// Create auto-bid
	bid := marketplace.NewAutoBid(listingID, bidderID, bidAmount, quantity, autoBidLimit, parentBidID)

	// Assertions
	assert.True(t, bid.IsAutoBid)
	assert.NotNil(t, bid.AutoBidLimit)
	assert.True(t, autoBidLimit.Equal(*bid.AutoBidLimit))
	assert.NotNil(t, bid.ParentBidID)
	assert.Equal(t, parentBidID, *bid.ParentBidID)
	assert.Equal(t, "Auto-bid", bid.Message)
}

func TestBidStatusTransitions(t *testing.T) {
	bid := marketplace.NewBid("LST_123", "USER_789", decimal.NewFromFloat(900), decimal.NewFromFloat(50), "Test bid")

	// Test valid transitions from ACTIVE
	assert.True(t, bid.CanTransitionTo(marketplace.BidStatusOutbid))
	assert.True(t, bid.CanTransitionTo(marketplace.BidStatusWinning))
	assert.True(t, bid.CanTransitionTo(marketplace.BidStatusExpired))
	assert.True(t, bid.CanTransitionTo(marketplace.BidStatusRemoved))

	// Test status update to WINNING
	err := bid.UpdateStatus(marketplace.BidStatusWinning)
	assert.NoError(t, err)
	assert.Equal(t, marketplace.BidStatusWinning, bid.Status)
	assert.True(t, bid.IsHighestBid)

	// Test transition from WINNING to OUTBID
	err = bid.UpdateStatus(marketplace.BidStatusOutbid)
	assert.NoError(t, err)
	assert.Equal(t, marketplace.BidStatusOutbid, bid.Status)
	assert.False(t, bid.IsHighestBid)
	assert.NotNil(t, bid.OutbidAt)
}

func TestBidAutoBidCapacity(t *testing.T) {
	autoBidLimit := decimal.NewFromFloat(1200.00)
	bid := marketplace.NewAutoBid("LST_123", "USER_789", decimal.NewFromFloat(900),
		decimal.NewFromFloat(50), autoBidLimit, "BID_PARENT_123")

	// Test auto-bid capacity
	assert.True(t, bid.HasAutoBidCapacity(decimal.NewFromFloat(1000.00)))
	assert.False(t, bid.HasAutoBidCapacity(decimal.NewFromFloat(1300.00)))

	// Test next auto-bid amount calculation
	minimumIncrement := decimal.NewFromFloat(10.00)

	// Should bid 1010 (1000 + 10)
	nextAmount := bid.GetNextAutoBidAmount(decimal.NewFromFloat(1000.00), minimumIncrement)
	assert.True(t, decimal.NewFromFloat(1010.00).Equal(nextAmount))

	// Should bid up to limit (1200) when competing bid is 1195
	nextAmount = bid.GetNextAutoBidAmount(decimal.NewFromFloat(1195.00), minimumIncrement)
	assert.True(t, autoBidLimit.Equal(nextAmount))

	// Should return zero when competing bid exceeds limit
	nextAmount = bid.GetNextAutoBidAmount(decimal.NewFromFloat(1250.00), minimumIncrement)
	assert.True(t, decimal.Zero.Equal(nextAmount))
}

func TestBidHelperMethods(t *testing.T) {
	bid := marketplace.NewBid("LST_123", "USER_789", decimal.NewFromFloat(900), decimal.NewFromFloat(50), "Test bid")

	// Test initial state
	assert.True(t, bid.IsActive())
	assert.False(t, bid.IsWinning())
	assert.True(t, bid.CanBeOutbid())

	// Test setting as highest bid
	bid.SetAsHighestBid()
	assert.True(t, bid.IsHighestBid)
	assert.Equal(t, marketplace.BidStatusWinning, bid.Status)
	assert.True(t, bid.IsWinning())

	// Test outbid
	err := bid.Outbid()
	assert.NoError(t, err)
	assert.Equal(t, marketplace.BidStatusOutbid, bid.Status)
	assert.False(t, bid.IsWinning())
	assert.False(t, bid.CanBeOutbid())
}

func TestAuctionEvent(t *testing.T) {
	listingID := "LST_123"
	actorID := "USER_456"
	eventType := marketplace.EventBidPlaced

	// Create auction event
	event := marketplace.NewAuctionEvent(listingID, eventType, actorID, marketplace.ActorTypeUser, nil)

	// Assertions
	assert.NotEmpty(t, event.ID)
	assert.NotEmpty(t, event.EventID)
	assert.Equal(t, listingID, event.ListingID)
	assert.Equal(t, eventType, event.EventType)
	assert.Equal(t, actorID, event.ActorID)
	assert.Equal(t, marketplace.ActorTypeUser, event.ActorType)
	assert.True(t, event.Timestamp.Before(time.Now().Add(1*time.Second)))
}

func TestAuctionEventData(t *testing.T) {
	event := marketplace.NewAuctionEvent("LST_123", marketplace.EventBidPlaced, "USER_456", marketplace.ActorTypeUser, nil)

	// Test setting event data
	eventData := map[string]interface{}{
		"bid_id":     "BID_789",
		"bid_amount": 950.00,
		"message":    "Test bid",
	}

	err := event.SetEventData(eventData)
	assert.NoError(t, err)
	assert.NotEmpty(t, event.EventData)

	// Test getting event data
	retrievedData, err := event.GetEventData()
	assert.NoError(t, err)
	assert.NotNil(t, retrievedData)
	assert.Equal(t, "BID_789", retrievedData["bid_id"])
	assert.Equal(t, 950.00, retrievedData["bid_amount"])
	assert.Equal(t, "Test bid", retrievedData["message"])

	// Test setting nil data
	err = event.SetEventData(nil)
	assert.NoError(t, err)
	assert.Empty(t, event.EventData)

	retrievedData, err = event.GetEventData()
	assert.NoError(t, err)
	assert.Nil(t, retrievedData)
}

func TestCreateSpecificEvents(t *testing.T) {
	// Test listing created event
	listing := marketplace.NewListing("PROD_123", "USER_456", "ORG_789",
		decimal.NewFromFloat(100), decimal.NewFromFloat(1000), decimal.NewFromFloat(800), 24)

	event, err := marketplace.CreateListingCreatedEvent(listing.ListingID, "USER_456", listing)
	require.NoError(t, err)
	assert.Equal(t, marketplace.EventListingCreated, event.EventType)
	assert.NotEmpty(t, event.EventData)

	// Test bid placed event
	bid := marketplace.NewBid(listing.ListingID, "USER_789", decimal.NewFromFloat(900), decimal.NewFromFloat(50), "Test bid")
	previousHighest := 850.00

	event, err = marketplace.CreateBidPlacedEvent(listing.ListingID, "USER_789", bid, &previousHighest)
	require.NoError(t, err)
	assert.Equal(t, marketplace.EventBidPlaced, event.EventType)
	assert.NotEmpty(t, event.EventData)

	// Test listing closed event
	event, err = marketplace.CreateListingClosedEvent(listing.ListingID, "SYSTEM", marketplace.ActorTypeSystem,
		"Auction expired", bid, 5, marketplace.ListingStatusExpired)
	require.NoError(t, err)
	assert.Equal(t, marketplace.EventListingClosed, event.EventType)
	assert.Equal(t, marketplace.ActorTypeSystem, event.ActorType)
	assert.NotEmpty(t, event.EventData)
}
