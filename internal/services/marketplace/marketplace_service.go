package marketplace

import (
	"kisanlink-ecom/internal/repositories/marketplace"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// MarketplaceServices aggregates all marketplace services
type MarketplaceServices struct {
	Listing              ListingServiceInterface
	Bidding              BiddingServiceInterface
	AuctionLifecycle     AuctionLifecycleServiceInterface
	AccessControl        AccessControlServiceInterface
	Analytics            AnalyticsServiceInterface
	Audit                AuditServiceInterface
	Automation           AuctionAutomationServiceInterface
	Cleanup              AuctionCleanupServiceInterface
	Notification         AuctionNotificationServiceInterface
	RealTimeNotification RealTimeNotificationService
	BidVisibility        BidVisibilityServiceInterface
	AuctionResults       AuctionResultsServiceInterface

	// Repositories (exposed for middleware)
	Repositories *marketplace.MarketplaceRepositories
}

// MarketplaceServiceConfig holds configuration for marketplace services
type MarketplaceServiceConfig struct {
	DBManager        db.DBManager
	EventService     EventServiceInterface
	NotificationSvc  NotificationServiceInterface
	InventoryService InventoryServiceInterface
	OrderService     OrderServiceInterface
	NetworkService   NetworkServiceInterface
	OrgService       OrganizationServiceInterface
}

// NewMarketplaceServices creates a new set of marketplace services
func NewMarketplaceServices(config *MarketplaceServiceConfig) *MarketplaceServices {
	// Initialize repositories
	repos := marketplace.NewMarketplaceRepositories(config.DBManager)

	// Initialize services
	listingService := NewListingService(
		repos.Listing,
		config.InventoryService,
		config.EventService,
	)

	biddingService := NewBiddingService(
		repos.Bid,
		repos.Listing,
		config.EventService,
		config.NotificationSvc,
	)

	auctionLifecycleService := NewAuctionLifecycleService(
		repos.Listing,
		repos.Bid,
		config.EventService,
		config.OrderService,
		config.NotificationSvc,
		config.InventoryService,
	)

	accessControlService := NewAccessControlService(
		repos.Listing,
		config.EventService,
		config.NetworkService,
		config.OrgService,
	)

	analyticsService := NewAnalyticsService(
		repos.Listing,
		repos.Bid,
		repos.AuctionEvent,
	)

	auditService := NewAuditService(
		repos.AuctionEvent,
		repos.Listing,
		repos.Bid,
	)

	// Initialize notification service
	notificationService := NewAuctionNotificationService(
		config.NotificationSvc,
		nil, // logger will be created internally
	)

	// Initialize real-time notification service
	realTimeNotificationService := NewRealTimeNotificationService()

	// Initialize cleanup service
	cleanupService := NewAuctionCleanupService(
		repos.Listing,
		repos.Bid,
		repos.AuctionEvent,
		config.EventService,
		nil, // logger will be created internally
	)

	// Initialize automation service
	automationService := NewAuctionAutomationService(
		auctionLifecycleService,
		cleanupService,
		notificationService,
		nil, // logger will be created internally
	)

	// Initialize bid visibility service
	bidVisibilityService := NewBidVisibilityService(
		repos.Bid,
		repos.Listing,
		accessControlService,
		config.EventService,
	)

	// Initialize auction results service
	auctionResultsService := NewAuctionResultsService(
		repos.Bid,
		repos.Listing,
		repos.AuctionEvent,
		bidVisibilityService,
		config.NotificationSvc,
		config.EventService,
	)

	return &MarketplaceServices{
		Listing:              listingService,
		Bidding:              biddingService,
		AuctionLifecycle:     auctionLifecycleService,
		AccessControl:        accessControlService,
		Analytics:            analyticsService,
		Audit:                auditService,
		Automation:           automationService,
		Cleanup:              cleanupService,
		Notification:         notificationService,
		RealTimeNotification: realTimeNotificationService,
		BidVisibility:        bidVisibilityService,
		AuctionResults:       auctionResultsService,
		Repositories:         repos,
	}
}

// GetListingService returns the listing service
func (s *MarketplaceServices) GetListingService() ListingServiceInterface {
	return s.Listing
}

// GetBiddingService returns the bidding service
func (s *MarketplaceServices) GetBiddingService() BiddingServiceInterface {
	return s.Bidding
}

// GetAuctionLifecycleService returns the auction lifecycle service
func (s *MarketplaceServices) GetAuctionLifecycleService() AuctionLifecycleServiceInterface {
	return s.AuctionLifecycle
}

// GetAccessControlService returns the access control service
func (s *MarketplaceServices) GetAccessControlService() AccessControlServiceInterface {
	return s.AccessControl
}

// GetAnalyticsService returns the analytics service
func (s *MarketplaceServices) GetAnalyticsService() AnalyticsServiceInterface {
	return s.Analytics
}

// GetAuditService returns the audit service
func (s *MarketplaceServices) GetAuditService() AuditServiceInterface {
	return s.Audit
}

// GetAutomationService returns the automation service
func (s *MarketplaceServices) GetAutomationService() AuctionAutomationServiceInterface {
	return s.Automation
}

// GetCleanupService returns the cleanup service
func (s *MarketplaceServices) GetCleanupService() AuctionCleanupServiceInterface {
	return s.Cleanup
}

// GetNotificationService returns the notification service
func (s *MarketplaceServices) GetNotificationService() AuctionNotificationServiceInterface {
	return s.Notification
}

// GetRealTimeNotificationService returns the real-time notification service
func (s *MarketplaceServices) GetRealTimeNotificationService() RealTimeNotificationService {
	return s.RealTimeNotification
}

// GetBidVisibilityService returns the bid visibility service
func (s *MarketplaceServices) GetBidVisibilityService() BidVisibilityServiceInterface {
	return s.BidVisibility
}

// GetAuctionResultsService returns the auction results service
func (s *MarketplaceServices) GetAuctionResultsService() AuctionResultsServiceInterface {
	return s.AuctionResults
}
