package marketplace

import (
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// MarketplaceRepositories aggregates all marketplace repositories
type MarketplaceRepositories struct {
	Listing      ListingRepository
	Bid          BidRepository
	AuctionEvent AuctionEventRepository
}

// NewMarketplaceRepositories creates a new set of marketplace repositories
func NewMarketplaceRepositories(dbManager db.DBManager) *MarketplaceRepositories {
	return &MarketplaceRepositories{
		Listing:      NewListingRepository(dbManager),
		Bid:          NewBidRepository(dbManager),
		AuctionEvent: NewAuctionEventRepository(dbManager),
	}
}

// Repository interfaces are defined in their respective files:
// - ListingRepository in listing_repository.go
// - BidRepository in bid_repository.go
// - AuctionEventRepository in auction_event_repository.go
