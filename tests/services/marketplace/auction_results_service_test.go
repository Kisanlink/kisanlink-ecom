package marketplace

import (
	"context"
	"testing"
	"time"

	"kisanlink-ecom/entities/models/marketplace"
	"kisanlink-ecom/internal/common"
	marketplaceService "kisanlink-ecom/internal/services/marketplace"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuctionEventRepository is a mock implementation of AuctionEventRepositoryInterface
type MockAuctionEventRepository struct {
	mock.Mock
}

func (m *MockAuctionEventRepository) Create(ctx context.Context, event *marketplace.AuctionEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockAuctionEventRepository) GetByID(ctx context.Context, id string) (*marketplace.AuctionEvent, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*marketplace.AuctionEvent), args.Error(1)
}

func (m *MockAuctionEventRepository) GetByListingID(ctx context.Context, listingID string, pagination *common.PaginationRequest) ([]*marketplace.AuctionEvent, int, error) {
	args := m.Called(ctx, listingID, pagination)
	return args.Get(0).([]*marketplace.AuctionEvent), args.Int(1), args.Error(2)
}

func (m *MockAuctionEventRepository) GetByEventType(ctx context.Context, eventType marketplace.AuctionEventType, pagination *common.PaginationRequest) ([]*marketplace.AuctionEvent, int, error) {
	args := m.Called(ctx, eventType, pagination)
	return args.Get(0).([]*marketplace.AuctionEvent), args.Int(1), args.Error(2)
}

// MockBidVisibilityService is a mock implementation of BidVisibilityServiceInterface
type MockBidVisibilityService struct {
	mock.Mock
}

func (m *MockBidVisibilityService) ApplyBidVisibilityRules(ctx context.Context, listing *marketplace.Listing, bids []*marketplace.Bid, viewerID string, viewerOrgID string) (*marketplaceService.FilteredBidHistory, error) {
	args := m.Called(ctx, listing, bids, viewerID, viewerOrgID)
	return args.Get(0).(*marketplaceService.FilteredBidHistory), args.Error(1)
}

func (m *MockBidVisibilityService) GetVisibleBidHistory(ctx context.Context, listingID string, viewerID string, viewerOrgID string, pagination *common.PaginationParams) (*marketplaceService.FilteredBidHistory, error) {
	args := m.Called(ctx, listingID, viewerID, viewerOrgID, pagination)
	return args.Get(0).(*marketplaceService.FilteredBidHistory), args.Error(1)
}

func (m *MockBidVisibilityService) GetVisibleBidStatistics(ctx context.Context, listingID string, viewerID string, viewerOrgID string) (*marketplaceService.FilteredBidStatistics, error) {
	args := m.Called(ctx, listingID, viewerID, viewerOrgID)
	return args.Get(0).(*marketplaceService.FilteredBidStatistics), args.Error(1)
}

func (m *MockBidVisibilityService) FilterBidForRealTimeUpdate(ctx context.Context, listing *marketplace.Listing, bid *marketplace.Bid, viewerID string, viewerOrgID string) (*marketplaceService.FilteredBid, error) {
	args := m.Called(ctx, listing, bid, viewerID, viewerOrgID)
	return args.Get(0).(*marketplaceService.FilteredBid), args.Error(1)
}

func (m *MockBidVisibilityService) FilterBidsForListing(ctx context.Context, listing *marketplace.Listing, bids []*marketplace.Bid, viewerID string, viewerOrgID string) ([]*marketplaceService.FilteredBid, error) {
	args := m.Called(ctx, listing, bids, viewerID, viewerOrgID)
	return args.Get(0).([]*marketplaceService.FilteredBid), args.Error(1)
}

func (m *MockBidVisibilityService) CanViewBidDetails(ctx context.Context, listing *marketplace.Listing, bid *marketplace.Bid, viewerID string, viewerOrgID string) bool {
	args := m.Called(ctx, listing, bid, viewerID, viewerOrgID)
	return args.Bool(0)
}

func (m *MockBidVisibilityService) CanViewListing(ctx context.Context, listing *marketplace.Listing, viewerID string, viewerOrgID string) bool {
	args := m.Called(ctx, listing, viewerID, viewerOrgID)
	return args.Bool(0)
}

func (m *MockBidVisibilityService) GetAuctionVisibilityState(ctx context.Context, listing *marketplace.Listing, viewerID string, viewerOrgID string) (*marketplaceService.AuctionVisibilityState, error) {
	args := m.Called(ctx, listing, viewerID, viewerOrgID)
	return args.Get(0).(*marketplaceService.AuctionVisibilityState), args.Error(1)
}

// MockNotificationService is a mock implementation of NotificationServiceInterface
type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) NotifyBidPlaced(ctx context.Context, bid interface{}, listing interface{}) error {
	args := m.Called(ctx, bid, listing)
	return args.Error(0)
}

func (m *MockNotificationService) NotifyBidOutbid(ctx context.Context, outbidBid interface{}, newBid interface{}, listing interface{}) error {
	args := m.Called(ctx, outbidBid, newBid, listing)
	return args.Error(0)
}

func (m *MockNotificationService) NotifyAutoBidTriggered(ctx context.Context, autoBid interface{}, triggeringBid interface{}, listing interface{}) error {
	args := m.Called(ctx, autoBid, triggeringBid, listing)
	return args.Error(0)
}

func (m *MockNotificationService) NotifyAuctionWon(ctx context.Context, listingID string, userID string, amount string) error {
	args := m.Called(ctx, listingID, userID, amount)
	return args.Error(0)
}

func (m *MockNotificationService) NotifyAuctionExpired(ctx context.Context, listingID string, sellerID string, winnerID string) error {
	args := m.Called(ctx, listingID, sellerID, winnerID)
	return args.Error(0)
}

func (m *MockNotificationService) SendNotification(ctx context.Context, request marketplaceService.NotificationRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockNotificationService) SendBatchNotifications(ctx context.Context, requests []marketplaceService.NotificationRequest) error {
	args := m.Called(ctx, requests)
	return args.Error(0)
}

// Test helper functions for auction results

func createCompletedTestListing() *marketplace.Listing {
	closedAt := time.Now().Add(-1 * time.Hour)
	return &marketplace.Listing{
		ListingID:       "LST_123",
		ProductID:       "PROD_456",
		SellerID:        "SELLER_789",
		OrganizationID:  "ORG_123",
		Quantity:        decimal.NewFromFloat(100),
		AskingPrice:     decimal.NewFromFloat(1000),
		MinimumBid:      decimal.NewFromFloat(900),
		Currency:        "INR",
		ListingDuration: 24,
		ExpiresAt:       time.Now().Add(-1 * time.Hour),
		Status:          marketplace.ListingStatusClosed,
		BidCount:        5,
		Visibility:      marketplace.VisibilityPublic,
		AuctionType:     marketplace.AuctionTypeOpen,
		BidVisibility:   marketplace.BidVisibilityFull,
		ClosedAt:        &closedAt,
	}
}

func createTestBidsForResults() []*marketplace.Bid {
	now := time.Now()
	return []*marketplace.Bid{
		{
			BidID:        "BID_1",
			ListingID:    "LST_123",
			BidderID:     "USER_1",
			BidAmount:    decimal.NewFromFloat(950),
			Currency:     "INR",
			Quantity:     decimal.NewFromFloat(100),
			IsAutoBid:    false,
			Status:       marketplace.BidStatusOutbid,
			IsHighestBid: false,
			PlacedAt:     now.Add(-4 * time.Hour),
		},
		{
			BidID:        "BID_2",
			ListingID:    "LST_123",
			BidderID:     "USER_2",
			BidAmount:    decimal.NewFromFloat(1000),
			Currency:     "INR",
			Quantity:     decimal.NewFromFloat(100),
			IsAutoBid:    false,
			Status:       marketplace.BidStatusOutbid,
			IsHighestBid: false,
			PlacedAt:     now.Add(-3 * time.Hour),
		},
		{
			BidID:        "BID_3",
			ListingID:    "LST_123",
			BidderID:     "USER_1",
			BidAmount:    decimal.NewFromFloat(1050),
			Currency:     "INR",
			Quantity:     decimal.NewFromFloat(100),
			IsAutoBid:    true,
			Status:       marketplace.BidStatusOutbid,
			IsHighestBid: false,
			PlacedAt:     now.Add(-2 * time.Hour),
		},
		{
			BidID:        "BID_4",
			ListingID:    "LST_123",
			BidderID:     "USER_3",
			BidAmount:    decimal.NewFromFloat(1100),
			Currency:     "INR",
			Quantity:     decimal.NewFromFloat(100),
			IsAutoBid:    false,
			Status:       marketplace.BidStatusOutbid,
			IsHighestBid: false,
			PlacedAt:     now.Add(-1*time.Hour - 30*time.Minute),
		},
		{
			BidID:        "BID_5",
			ListingID:    "LST_123",
			BidderID:     "USER_2",
			BidAmount:    decimal.NewFromFloat(1200),
			Currency:     "INR",
			Quantity:     decimal.NewFromFloat(100),
			IsAutoBid:    false,
			Status:       marketplace.BidStatusWinning,
			IsHighestBid: true,
			PlacedAt:     now.Add(-1 * time.Hour),
		},
	}
}

// Test cases for AuctionResultsService

func TestAuctionResultsService_ProcessAuctionResults_SuccessfulAuction(t *testing.T) {
	// Setup
	mockBidRepo := new(MockBidRepository)
	mockListingRepo := new(MockListingRepository)
	mockEventRepo := new(MockAuctionEventRepository)
	mockVisibilityService := new(MockBidVisibilityService)
	mockNotificationService := new(MockNotificationService)
	mockEventService := new(MockEventService)

	service := marketplaceService.NewAuctionResultsService(
		mockBidRepo,
		mockListingRepo,
		mockEventRepo,
		mockVisibilityService,
		mockNotificationService,
		mockEventService,
	)

	// Test data
	listing := createCompletedTestListing()
	bids := createTestBidsForResults()
	winningBid := bids[4] // Last bid is the winning bid

	// Mock expectations
	mockListingRepo.On("GetByListingID", mock.Anything, "LST_123").Return(listing, nil)
	mockBidRepo.On("GetBidHistory", mock.Anything, "LST_123", marketplace.BidVisibilityFull, "", mock.AnythingOfType("*common.PaginationRequest")).Return(bids, 5, nil)
	mockBidRepo.On("GetHighestBid", mock.Anything, "LST_123").Return(winningBid, nil)

	// Execute
	ctx := context.Background()
	results, err := service.ProcessAuctionResults(ctx, "LST_123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, "LST_123", results.ListingID)
	assert.Equal(t, marketplace.ListingStatusClosed, results.Status)
	assert.NotNil(t, results.WinningBid)
	assert.Equal(t, "BID_5", results.WinningBid.BidID)
	assert.Equal(t, 5, results.TotalBids)
	assert.Equal(t, 3, results.UniqueBidders) // USER_1, USER_2, USER_3
	assert.Equal(t, decimal.NewFromFloat(1200), results.FinalPrice)
	assert.Equal(t, decimal.NewFromFloat(300), results.PriceIncrease) // 1200 - 900
	assert.Len(t, results.BidProgression, 5)
	assert.Len(t, results.ParticipantStats, 3)

	// Check participant stats
	var user2Stats *marketplaceService.ParticipantStats
	for _, stats := range results.ParticipantStats {
		if stats.BidderID == "USER_2" {
			user2Stats = &stats
			break
		}
	}
	assert.NotNil(t, user2Stats)
	assert.True(t, user2Stats.IsWinner)
	assert.Equal(t, 2, user2Stats.TotalBids)
	assert.Equal(t, decimal.NewFromFloat(1200), user2Stats.HighestBid)

	mockListingRepo.AssertExpectations(t)
	mockBidRepo.AssertExpectations(t)
}

func TestAuctionResultsService_ProcessAuctionResults_NoWinner(t *testing.T) {
	// Setup
	mockBidRepo := new(MockBidRepository)
	mockListingRepo := new(MockListingRepository)
	mockEventRepo := new(MockAuctionEventRepository)
	mockVisibilityService := new(MockBidVisibilityService)
	mockNotificationService := new(MockNotificationService)
	mockEventService := new(MockEventService)

	service := marketplaceService.NewAuctionResultsService(
		mockBidRepo,
		mockListingRepo,
		mockEventRepo,
		mockVisibilityService,
		mockNotificationService,
		mockEventService,
	)

	// Test data - auction with no bids
	listing := createCompletedTestListing()
	listing.Status = marketplace.ListingStatusExpiredNoBids
	listing.BidCount = 0

	// Mock expectations
	mockListingRepo.On("GetByListingID", mock.Anything, "LST_123").Return(listing, nil)
	mockBidRepo.On("GetBidHistory", mock.Anything, "LST_123", marketplace.BidVisibilityFull, "", mock.AnythingOfType("*common.PaginationRequest")).Return([]*marketplace.Bid{}, 0, nil)
	mockBidRepo.On("GetHighestBid", mock.Anything, "LST_123").Return(nil, nil)

	// Execute
	ctx := context.Background()
	results, err := service.ProcessAuctionResults(ctx, "LST_123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, "LST_123", results.ListingID)
	assert.Equal(t, marketplace.ListingStatusExpiredNoBids, results.Status)
	assert.Nil(t, results.WinningBid)
	assert.Equal(t, 0, results.TotalBids)
	assert.Equal(t, 0, results.UniqueBidders)
	assert.Equal(t, decimal.NewFromFloat(900), results.FinalPrice) // Minimum bid
	assert.Equal(t, decimal.Zero, results.PriceIncrease)
	assert.Empty(t, results.BidProgression)
	assert.Empty(t, results.ParticipantStats)

	mockListingRepo.AssertExpectations(t)
	mockBidRepo.AssertExpectations(t)
}

func TestAuctionResultsService_RevealAuctionResults_WithVisibilityFiltering(t *testing.T) {
	// Setup
	mockBidRepo := new(MockBidRepository)
	mockListingRepo := new(MockListingRepository)
	mockEventRepo := new(MockAuctionEventRepository)
	mockVisibilityService := new(MockBidVisibilityService)
	mockNotificationService := new(MockNotificationService)
	mockEventService := new(MockEventService)

	service := marketplaceService.NewAuctionResultsService(
		mockBidRepo,
		mockListingRepo,
		mockEventRepo,
		mockVisibilityService,
		mockNotificationService,
		mockEventService,
	)

	// Test data
	listing := createCompletedTestListing()
	bids := createTestBidsForResults()
	winningBid := bids[4]

	auctionState := &marketplaceService.AuctionVisibilityState{
		IsActive:      false,
		IsExpired:     true,
		AuctionType:   marketplace.AuctionTypeOpen,
		BidVisibility: marketplace.BidVisibilityFull,
		CanRevealBids: true,
		RevealReason:  "auction_ended",
	}

	// Mock expectations
	mockListingRepo.On("GetByListingID", mock.Anything, "LST_123").Return(listing, nil).Times(2)
	mockBidRepo.On("GetBidHistory", mock.Anything, "LST_123", marketplace.BidVisibilityFull, "", mock.AnythingOfType("*common.PaginationRequest")).Return(bids, 5, nil)
	mockBidRepo.On("GetHighestBid", mock.Anything, "LST_123").Return(winningBid, nil)
	mockVisibilityService.On("CanViewListing", mock.Anything, listing, "VIEWER_1", "ORG_123").Return(true)
	mockVisibilityService.On("GetAuctionVisibilityState", mock.Anything, listing, "VIEWER_1", "ORG_123").Return(auctionState, nil)

	// Execute
	ctx := context.Background()
	results, err := service.RevealAuctionResults(ctx, "LST_123", "VIEWER_1", "ORG_123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, "LST_123", results.ListingID)
	assert.Equal(t, marketplace.ListingStatusClosed, results.Status)
	assert.NotNil(t, results.FinalPrice)
	assert.Equal(t, decimal.NewFromFloat(1200), *results.FinalPrice)
	assert.Equal(t, marketplaceService.VisibilityLevelFull, results.VisibilityLevel)
	assert.Equal(t, "auction_ended", results.RevealReason)

	mockListingRepo.AssertExpectations(t)
	mockBidRepo.AssertExpectations(t)
	mockVisibilityService.AssertExpectations(t)
}

func TestAuctionResultsService_GetAuctionSummary_SuccessfulAuction(t *testing.T) {
	// Setup
	mockBidRepo := new(MockBidRepository)
	mockListingRepo := new(MockListingRepository)
	mockEventRepo := new(MockAuctionEventRepository)
	mockVisibilityService := new(MockBidVisibilityService)
	mockNotificationService := new(MockNotificationService)
	mockEventService := new(MockEventService)

	service := marketplaceService.NewAuctionResultsService(
		mockBidRepo,
		mockListingRepo,
		mockEventRepo,
		mockVisibilityService,
		mockNotificationService,
		mockEventService,
	)

	// Test data
	listing := createCompletedTestListing()
	bids := createTestBidsForResults()
	winningBid := bids[4]

	// Mock expectations
	mockListingRepo.On("GetByListingID", mock.Anything, "LST_123").Return(listing, nil).Times(2)
	mockBidRepo.On("GetBidHistory", mock.Anything, "LST_123", marketplace.BidVisibilityFull, "", mock.AnythingOfType("*common.PaginationRequest")).Return(bids, 5, nil)
	mockBidRepo.On("GetHighestBid", mock.Anything, "LST_123").Return(winningBid, nil)
	mockVisibilityService.On("CanViewListing", mock.Anything, listing, "VIEWER_1", "ORG_123").Return(true)

	// Execute
	ctx := context.Background()
	summary, err := service.GetAuctionSummary(ctx, "LST_123", "VIEWER_1", "ORG_123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, "LST_123", summary.ListingID)
	assert.Equal(t, marketplace.ListingStatusClosed, summary.Status)
	assert.Equal(t, marketplaceService.OutcomeSuccessful, summary.Outcome)
	assert.NotNil(t, summary.WinnerInfo)
	assert.Equal(t, decimal.NewFromFloat(1200), summary.WinnerInfo.WinningAmount)
	assert.False(t, summary.WinnerInfo.IsAutoBid)
	assert.Equal(t, 5, summary.FinalMetrics.TotalBids)
	assert.Equal(t, 3, summary.FinalMetrics.UniqueBidders)
	assert.Equal(t, decimal.NewFromFloat(1200), summary.FinalMetrics.FinalPrice)
	assert.Equal(t, decimal.NewFromFloat(300), summary.FinalMetrics.PriceIncrease)
	assert.NotEmpty(t, summary.KeyEvents)

	// Check key events
	hasAuctionStarted := false
	hasWinningBid := false
	hasAuctionEnded := false
	for _, event := range summary.KeyEvents {
		switch event.EventType {
		case "AUCTION_STARTED":
			hasAuctionStarted = true
		case "WINNING_BID":
			hasWinningBid = true
		case "AUCTION_ENDED":
			hasAuctionEnded = true
		}
	}
	assert.True(t, hasAuctionStarted)
	assert.True(t, hasWinningBid)
	assert.True(t, hasAuctionEnded)

	mockListingRepo.AssertExpectations(t)
	mockBidRepo.AssertExpectations(t)
	mockVisibilityService.AssertExpectations(t)
}

func TestAuctionResultsService_GetHistoricalBidData_WithFiltering(t *testing.T) {
	// Setup
	mockBidRepo := new(MockBidRepository)
	mockListingRepo := new(MockListingRepository)
	mockEventRepo := new(MockAuctionEventRepository)
	mockVisibilityService := new(MockBidVisibilityService)
	mockNotificationService := new(MockNotificationService)
	mockEventService := new(MockEventService)

	service := marketplaceService.NewAuctionResultsService(
		mockBidRepo,
		mockListingRepo,
		mockEventRepo,
		mockVisibilityService,
		mockNotificationService,
		mockEventService,
	)

	// Test data
	listing := createCompletedTestListing()
	bids := createTestBidsForResults()

	filter := &marketplaceService.HistoricalDataFilter{
		BidderID:        "USER_2",
		IncludeAutoBids: false,
		SortOrder:       "desc",
	}

	// Mock expectations
	mockListingRepo.On("GetByListingID", mock.Anything, "LST_123").Return(listing, nil)
	mockBidRepo.On("GetBidHistory", mock.Anything, "LST_123", marketplace.BidVisibilityFull, "VIEWER_1", mock.AnythingOfType("*common.PaginationRequest")).Return(bids, 5, nil)

	// Execute
	ctx := context.Background()
	historicalData, err := service.GetHistoricalBidData(ctx, "LST_123", "VIEWER_1", "ORG_123", filter)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, historicalData)
	assert.Equal(t, "LST_123", historicalData.ListingID)
	assert.Equal(t, "VIEWER_1", historicalData.RequestedBy)
	assert.Equal(t, marketplaceService.HistoricalAccessFull, historicalData.AccessLevel) // Seller should have full access
	assert.Equal(t, 5, historicalData.TotalRecords)
	assert.LessOrEqual(t, historicalData.FilteredRecords, historicalData.TotalRecords)
	assert.NotEmpty(t, historicalData.BidHistory)

	mockListingRepo.AssertExpectations(t)
	mockBidRepo.AssertExpectations(t)
}

func TestAuctionResultsService_ValidateAuctionResults_ValidResults(t *testing.T) {
	// Setup
	mockBidRepo := new(MockBidRepository)
	mockListingRepo := new(MockListingRepository)
	mockEventRepo := new(MockAuctionEventRepository)
	mockVisibilityService := new(MockBidVisibilityService)
	mockNotificationService := new(MockNotificationService)
	mockEventService := new(MockEventService)

	service := marketplaceService.NewAuctionResultsService(
		mockBidRepo,
		mockListingRepo,
		mockEventRepo,
		mockVisibilityService,
		mockNotificationService,
		mockEventService,
	)

	// Test data
	listing := createCompletedTestListing()
	bids := createTestBidsForResults()
	winningBid := bids[4]

	// Mock expectations
	mockListingRepo.On("GetByListingID", mock.Anything, "LST_123").Return(listing, nil)
	mockBidRepo.On("GetBidHistory", mock.Anything, "LST_123", marketplace.BidVisibilityFull, "", mock.AnythingOfType("*common.PaginationRequest")).Return(bids, 5, nil)
	mockBidRepo.On("GetHighestBid", mock.Anything, "LST_123").Return(winningBid, nil)

	// Execute
	ctx := context.Background()
	validation, err := service.ValidateAuctionResults(ctx, "LST_123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, validation)
	assert.True(t, validation.IsValid)
	assert.Empty(t, validation.ValidationErrors)
	assert.NotEmpty(t, validation.CheckedAt)
	assert.Equal(t, "SYSTEM", validation.CheckedBy)

	// Should have warnings for single bidder scenarios or other edge cases
	// In this case, we have multiple bidders so no warnings expected for that

	mockListingRepo.AssertExpectations(t)
	mockBidRepo.AssertExpectations(t)
}

func TestAuctionResultsService_NotifyWinner_Success(t *testing.T) {
	// Setup
	mockBidRepo := new(MockBidRepository)
	mockListingRepo := new(MockListingRepository)
	mockEventRepo := new(MockAuctionEventRepository)
	mockVisibilityService := new(MockBidVisibilityService)
	mockNotificationService := new(MockNotificationService)
	mockEventService := new(MockEventService)

	service := marketplaceService.NewAuctionResultsService(
		mockBidRepo,
		mockListingRepo,
		mockEventRepo,
		mockVisibilityService,
		mockNotificationService,
		mockEventService,
	)

	// Test data
	listing := createCompletedTestListing()
	winningBid := createTestBidsForResults()[4]

	// Mock expectations
	mockEventService.On("RecordListingEvent", mock.Anything, "LST_123", "WINNER_NOTIFIED", mock.AnythingOfType("map[string]interface {}"), "SYSTEM").Return(nil)
	mockNotificationService.On("NotifyBidPlaced", mock.Anything, winningBid, listing).Return(nil)

	// Execute
	ctx := context.Background()
	err := service.NotifyWinner(ctx, listing, winningBid)

	// Assert
	assert.NoError(t, err)

	mockEventService.AssertExpectations(t)
	mockNotificationService.AssertExpectations(t)
}

func TestAuctionResultsService_NotifyWinner_NoWinner(t *testing.T) {
	// Setup
	mockBidRepo := new(MockBidRepository)
	mockListingRepo := new(MockListingRepository)
	mockEventRepo := new(MockAuctionEventRepository)
	mockVisibilityService := new(MockBidVisibilityService)
	mockNotificationService := new(MockNotificationService)
	mockEventService := new(MockEventService)

	service := marketplaceService.NewAuctionResultsService(
		mockBidRepo,
		mockListingRepo,
		mockEventRepo,
		mockVisibilityService,
		mockNotificationService,
		mockEventService,
	)

	// Test data
	listing := createCompletedTestListing()

	// Execute - no winner to notify
	ctx := context.Background()
	err := service.NotifyWinner(ctx, listing, nil)

	// Assert - should not error when no winner
	assert.NoError(t, err)

	// No expectations should be called since there's no winner
}

func TestAuctionResultsService_ReconcileResults_Success(t *testing.T) {
	// Setup
	mockBidRepo := new(MockBidRepository)
	mockListingRepo := new(MockListingRepository)
	mockEventRepo := new(MockAuctionEventRepository)
	mockVisibilityService := new(MockBidVisibilityService)
	mockNotificationService := new(MockNotificationService)
	mockEventService := new(MockEventService)

	service := marketplaceService.NewAuctionResultsService(
		mockBidRepo,
		mockListingRepo,
		mockEventRepo,
		mockVisibilityService,
		mockNotificationService,
		mockEventService,
	)

	// Execute
	ctx := context.Background()
	reconciliation, err := service.ReconcileResults(ctx, "LST_123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, reconciliation)
	assert.NotEmpty(t, reconciliation.ReconciliationID)
	assert.Equal(t, marketplaceService.ReconciliationCompleted, reconciliation.Status)
	assert.Equal(t, "SYSTEM", reconciliation.ReconciledBy)
	assert.NotEmpty(t, reconciliation.ReconciledAt)
}
