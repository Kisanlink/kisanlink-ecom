package marketplace

import marketplaceService "github.com/Kisanlink/kisanlink-ecom/internal/services/marketplace"

// MarketplaceHandler aggregates all marketplace-related handlers
type MarketplaceHandler struct {
	Listing      *ListingHandler
	Bidding      *BiddingHandler
	Admin        *AdminHandler
	Notification *NotificationHandler
}

// NewMarketplaceHandler creates a new marketplace handler with all sub-handlers
func NewMarketplaceHandler(marketplaceServices *marketplaceService.MarketplaceServices) *MarketplaceHandler {
	return &MarketplaceHandler{
		Listing:      NewListingHandler(marketplaceServices),
		Bidding:      NewBiddingHandler(marketplaceServices, marketplaceServices.GetBidVisibilityService(), marketplaceServices.GetAuctionResultsService()),
		Admin:        NewAdminHandler(marketplaceServices),
		Notification: NewNotificationHandler(marketplaceServices.GetRealTimeNotificationService()),
	}
}

// GetListingHandler returns the listing handler
func (h *MarketplaceHandler) GetListingHandler() *ListingHandler {
	return h.Listing
}

// GetBiddingHandler returns the bidding handler
func (h *MarketplaceHandler) GetBiddingHandler() *BiddingHandler {
	return h.Bidding
}

// GetAdminHandler returns the admin handler
func (h *MarketplaceHandler) GetAdminHandler() *AdminHandler {
	return h.Admin
}

// GetNotificationHandler returns the notification handler
func (h *MarketplaceHandler) GetNotificationHandler() *NotificationHandler {
	return h.Notification
}
