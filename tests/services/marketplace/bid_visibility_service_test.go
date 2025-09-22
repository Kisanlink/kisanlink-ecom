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

// MockBidRepository is a mock implementation of BidRepositoryInterface
type MockBidRepository struct {
	mock.Mock
}

func (m *MockBidRepository) Create(ctx context.Context, bid *marketplace.Bid) error {
	args := m.Called(ctx, bid)
	return args.Error(0)
}

func (m *MockBidRepository) GetByID(ctx context.Context, id string) (*marketplace.Bid, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*marketplace.Bid), args.Error(1)
}

func (m *MockBidRepository) GetByBidID(ctx context.Context, bidID string) (*marketplace.Bid, error) {
	args := m.Called(ctx, bidID)
	return args.Get(0).(*marketplace.Bid), args.Error(1)
}

func (m *MockBidRepository) Update(ctx context.Context, bid *marketplace.Bid) error {
	args := m.Called(ctx, bid)
	return args.Error(0)
}

func (m *MockBidRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBidRepository) GetBidHistory(ctx context.Context, listingID string, bidVisibility marketplace.BidVisibility, viewerID string, pagination *common.PaginationRequest) ([]*marketplace.Bid, int, error) {
	args := m.Called(ctx, listingID, bidVisibility, viewerID, pagination)
	return args.Get(0).([]*marketplace.Bid), args.Int(1), args.Error(2)
}

func (m *MockBidRepository) GetUserBids(ctx context.Context, userID string, filter *marketplace.BidFilter, pagination *common.PaginationRequest) ([]*marketplace.Bid, int, error) {
	args := m.Called(ctx, userID, filter, pagination)
	return args.Get(0).([]*marketplace.Bid), args.Int(1), args.Error(2)
}

func (m *MockBidRepository) GetAllBids(ctx context.Context, filter *marketplace.BidFilter, pagination *common.PaginationRequest) ([]*marketplace.Bid, int, error) {
	args := m.Called(ctx, filter, pagination)
	return args.Get(0).([]*marketplace.Bid), args.Int(1), args.Error(2)
}

func (m *MockBidRepository) GetHighestBid(ctx context.Context, listingID string) (*marketplace.Bid, error) {
	args := m.Called(ctx, listingID)
	return args.Get(0).(*marketplace.Bid), args.Error(1)
}

func (m *MockBidRepository) GetBidStatistics(ctx context.Context, listingID string) (*marketplace.BidStatistics, error) {
	args := m.Called(ctx, listingID)
	return args.Get(0).(*marketplace.BidStatistics), args.Error(1)
}

func (m *MockBidRepository) GetUserAutoBids(ctx context.Context, userID string, listingID string) ([]*marketplace.Bid, error) {
	args := m.Called(ctx, userID, listingID)
	return args.Get(0).([]*marketplace.Bid), args.Error(1)
}

func (m *MockBidRepository) GetAutoBidsForListing(ctx context.Context, listingID string) ([]*marketplace.Bid, error) {
	args := m.Called(ctx, listingID)
	return args.Get(0).([]*marketplace.Bid), args.Error(1)
}

func (m *MockBidRepository) UpdateAutoBidStatus(ctx context.Context, bidID string, status marketplace.BidStatus) error {
	args := m.Called(ctx, bidID, status)
	return args.Error(0)
}

func (m *MockBidRepository) PlaceBidAtomic(ctx context.Context, bid *marketplace.Bid, listingID string) (*marketplace.Bid, error) {
	args := m.Called(ctx, bid, listingID)
	return args.Get(0).(*marketplace.Bid), args.Error(1)
}

func (m *MockBidRepository) RemoveBid(ctx context.Context, bidID string, reason string, adminID string) error {
	args := m.Called(ctx, bidID, reason, adminID)
	return args.Error(0)
}

func (m *MockBidRepository) ExpireBidsForListing(ctx context.Context, listingID string) error {
	args := m.Called(ctx, listingID)
	return args.Error(0)
}

func (m *MockBidRepository) GetBidRanking(ctx context.Context, listingID string, limit int) ([]*marketplace.Bid, error) {
	args := m.Called(ctx, listingID, limit)
	return args.Get(0).([]*marketplace.Bid), args.Error(1)
}

// MockListingRepository is a mock implementation of ListingRepositoryInterface
type MockListingRepository struct {
	mock.Mock
}

func (m *MockListingRepository) Create(ctx context.Context, listing *marketplace.Listing) error {
	args := m.Called(ctx, listing)
	return args.Error(0)
}

func (m *MockListingRepository) GetByID(ctx context.Context, id string) (*marketplace.Listing, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*marketplace.Listing), args.Error(1)
}

func (m *MockListingRepository) GetByListingID(ctx context.Context, listingID string) (*marketplace.Listing, error) {
	args := m.Called(ctx, listingID)
	return args.Get(0).(*marketplace.Listing), args.Error(1)
}

func (m *MockListingRepository) Update(ctx context.Context, listing *marketplace.Listing) error {
	args := m.Called(ctx, listing)
	return args.Error(0)
}

func (m *MockListingRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockListingRepository) GetActiveListings(ctx context.Context, viewerOrgID string, filter *marketplace.ListingFilter, pagination *common.PaginationRequest) ([]*marketplace.Listing, int, error) {
	args := m.Called(ctx, viewerOrgID, filter, pagination)
	return args.Get(0).([]*marketplace.Listing), args.Int(1), args.Error(2)
}

func (m *MockListingRepository) GetExpiredListings(ctx context.Context, limit int) ([]*marketplace.Listing, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*marketplace.Listing), args.Error(1)
}

func (m *MockListingRepository) GetUserListings(ctx context.Context, userID string, filter *marketplace.ListingFilter, pagination *common.PaginationRequest) ([]*marketplace.Listing, int, error) {
	args := m.Called(ctx, userID, filter, pagination)
	return args.Get(0).([]*marketplace.Listing), args.Int(1), args.Error(2)
}

func (m *MockListingRepository) GetAllListings(ctx context.Context, filter *marketplace.ListingFilter, pagination *common.PaginationRequest) ([]*marketplace.Listing, int, error) {
	args := m.Called(ctx, filter, pagination)
	return args.Get(0).([]*marketplace.Listing), args.Int(1), args.Error(2)
}

func (m *MockListingRepository) IncrementBidCount(ctx context.Context, listingID string) error {
	args := m.Called(ctx, listingID)
	return args.Error(0)
}

func (m *MockListingRepository) UpdateHighestBid(ctx context.Context, listingID string, bidID string) error {
	args := m.Called(ctx, listingID, bidID)
	return args.Error(0)
}

func (m *MockListingRepository) UpdateStatus(ctx context.Context, listingID string, status marketplace.ListingStatus, reason string) error {
	args := m.Called(ctx, listingID, status, reason)
	return args.Error(0)
}

// MockAccessControlService is a mock implementation of AccessControlServiceInterface
type MockAccessControlService struct {
	mock.Mock
}

func (m *MockAccessControlService) CanViewListing(ctx context.Context, listing *marketplace.Listing, userID string, orgID string) bool {
	args := m.Called(ctx, listing, userID, orgID)
	return args.Bool(0)
}

func (m *MockAccessControlService) CanBidOnListing(ctx context.Context, listingID string, userID string, orgID string) (bool, error) {
	args := m.Called(ctx, listingID, userID, orgID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAccessControlService) CanEditListing(ctx context.Context, listingID string, userID string, orgID string) (bool, error) {
	args := m.Called(ctx, listingID, userID, orgID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAccessControlService) InviteParticipant(ctx context.Context, req *marketplaceService.InviteParticipantRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockAccessControlService) RemoveParticipant(ctx context.Context, req *marketplaceService.RemoveParticipantRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockAccessControlService) GetInvitedParticipants(ctx context.Context, listingID string, pagination *common.PaginationParams) ([]*marketplaceService.AuctionParticipant, int, error) {
	args := m.Called(ctx, listingID, pagination)
	return args.Get(0).([]*marketplaceService.AuctionParticipant), args.Int(1), args.Error(2)
}

func (m *MockAccessControlService) IsParticipantInvited(ctx context.Context, listingID string, userID string) (bool, error) {
	args := m.Called(ctx, listingID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAccessControlService) ValidateNetworkMembership(ctx context.Context, userOrgID string, listingOrgID string) (bool, error) {
	args := m.Called(ctx, userOrgID, listingOrgID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAccessControlService) ValidateOrganizationMembership(ctx context.Context, userOrgID string, listingOrgID string) (bool, error) {
	args := m.Called(ctx, userOrgID, listingOrgID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAccessControlService) ValidateVisibilityConfiguration(ctx context.Context, listing *marketplace.Listing) error {
	args := m.Called(ctx, listing)
	return args.Error(0)
}

func (m *MockAccessControlService) GetVisibilityRules(ctx context.Context, listingID string) (*marketplaceService.VisibilityRules, error) {
	args := m.Called(ctx, listingID)
	return args.Get(0).(*marketplaceService.VisibilityRules), args.Error(1)
}

func (m *MockAccessControlService) LogAccessAttempt(ctx context.Context, req *marketplaceService.AccessAttemptLog) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockAccessControlService) GetAccessLogs(ctx context.Context, listingID string, pagination *common.PaginationParams) ([]*marketplaceService.AccessAttemptLog, int, error) {
	args := m.Called(ctx, listingID, pagination)
	return args.Get(0).([]*marketplaceService.AccessAttemptLog), args.Int(1), args.Error(2)
}

// MockEventService is a mock implementation of EventServiceInterface
type MockEventService struct {
	mock.Mock
}

func (m *MockEventService) RecordListingEvent(ctx context.Context, listingID string, eventType marketplace.AuctionEventType, eventData map[string]interface{}, actorID string) error {
	args := m.Called(ctx, listingID, eventType, eventData, actorID)
	return args.Error(0)
}

func (m *MockEventService) GetListingEvents(ctx context.Context, listingID string, pagination *common.PaginationRequest) ([]*marketplace.AuctionEvent, int, error) {
	args := m.Called(ctx, listingID, pagination)
	return args.Get(0).([]*marketplace.AuctionEvent), args.Int(1), args.Error(2)
}

// Test helper functions

func createTestListing(auctionType marketplace.AuctionType, bidVisibility marketplace.BidVisibility, visibility marketplace.ListingVisibility) *marketplace.Listing {
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
		ExpiresAt:       time.Now().Add(24 * time.Hour),
		Status:          marketplace.ListingStatusActive,
		BidCount:        3,
		Visibility:      visibility,
		AuctionType:     auctionType,
		BidVisibility:   bidVisibility,
	}
}

func createTestBid(bidderID string, amount float64, isHighest bool) *marketplace.Bid {
	return &marketplace.Bid{
		BidID:        "BID_" + bidderID,
		ListingID:    "LST_123",
		BidderID:     bidderID,
		BidAmount:    decimal.NewFromFloat(amount),
		Currency:     "INR",
		Quantity:     decimal.NewFromFloat(100),
		IsAutoBid:    false,
		Status:       marketplace.BidStatusActive,
		IsHighestBid: isHighest,
		PlacedAt:     time.Now(),
	}
}

func createTestBidStatistics() *marketplace.BidStatistics {
	return &marketplace.BidStatistics{
		ListingID:     "LST_123",
		TotalBids:     5,
		UniqueBidders: 3,
		HighestBid:    decimal.NewFromFloat(1200),
		AverageBid:    decimal.NewFromFloat(1050),
		BidIncrement:  decimal.NewFromFloat(50),
		AutoBidCount:  2,
	}
}

// Test cases

func TestBidVisibilityService_ApplyBidVisibilityRules_OpenAuctionFullVisibility(t *testing.T) {
	// Setup
	mockBidRepo := new(MockBidRepository)
	mockListingRepo := new(MockListingRepository)
	mockAccessControl := new(MockAccessControlService)
	mockEventService := new(MockEventService)

	service := marketplaceService.NewBidVisibilityService(
		mockBidRepo,
		mockListingRepo,
		mockAccessControl,
		mockEventService,
	)

	// Test data
	listing := createTestListing(marketplace.AuctionTypeOpen, marketplace.BidVisibilityFull, marketplace.VisibilityPublic)
	bids := []*marketplace.Bid{
		createTestBid("USER_1", 1000, false),
		createTestBid("USER_2", 1100, false),
		createTestBid("USER_3", 1200, true),
	}

	// Mock expectations
	mockAccessControl.On("CanViewListing", mock.Anything, listing, "VIEWER_1", "ORG_123").Return(true)

	// Execute
	ctx := context.Background()
	result, err := service.ApplyBidVisibilityRules(ctx, listing, bids, "VIEWER_1", "ORG_123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "LST_123", result.ListingID)
	assert.Equal(t, 3, result.TotalBids)
	assert.Equal(t, 3, result.VisibleBids)
	assert.Equal(t, marketplaceService.VisibilityLevelFull, result.VisibilityLevel)
	assert.Len(t, result.Bids, 3)

	// Check that all bid details are visible
	for _, filteredBid := range result.Bids {
		assert.NotNil(t, filteredBid.BidAmount)
		assert.NotNil(t, filteredBid.PlacedAt)
		assert.NotNil(t, filteredBid.IsHighestBid)
		assert.NotNil(t, filteredBid.IsAutoBid)
	}

	mockAccessControl.AssertExpectations(t)
}

func TestBidVisibilityService_ApplyBidVisibilityRules_ClosedAuctionMinimalVisibility(t *testing.T) {
	// Setup
	mockBidRepo := new(MockBidRepository)
	mockListingRepo := new(MockListingRepository)
	mockAccessControl := new(MockAccessControlService)
	mockEventService := new(MockEventService)

	service := marketplaceService.NewBidVisibilityService(
		mockBidRepo,
		mockListingRepo,
		mockAccessControl,
		mockEventService,
	)

	// Test data - closed auction with minimal visibility
	listing := createTestListing(marketplace.AuctionTypeClosed, marketplace.BidVisibilityMinimal, marketplace.VisibilityPublic)
	bids := []*marketplace.Bid{
		createTestBid("USER_1", 1000, false),
		createTestBid("USER_2", 1100, false),
		createTestBid("USER_3", 1200, true),
	}

	// Mock expectations
	mockAccessControl.On("CanViewListing", mock.Anything, listing, "VIEWER_1", "ORG_123").Return(true)

	// Execute
	ctx := context.Background()
	result, err := service.ApplyBidVisibilityRules(ctx, listing, bids, "VIEWER_1", "ORG_123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, marketplaceService.VisibilityLevelMinimal, result.VisibilityLevel)
	assert.Len(t, result.Bids, 3)

	// Check that bid amounts are hidden but timestamps are visible
	for _, filteredBid := range result.Bids {
		if !filteredBid.IsOwnBid {
			assert.Nil(t, filteredBid.BidAmount)   // Amounts should be hidden
			assert.NotNil(t, filteredBid.PlacedAt) // Timestamps should be visible
		}
	}

	mockAccessControl.AssertExpectations(t)
}

func TestBidVisibilityService_ApplyBidVisibilityRules_OwnBidsAlwaysVisible(t *testing.T) {
	// Setup
	mockBidRepo := new(MockBidRepository)
	mockListingRepo := new(MockListingRepository)
	mockAccessControl := new(MockAccessControlService)
	mockEventService := new(MockEventService)

	service := marketplaceService.NewBidVisibilityService(
		mockBidRepo,
		mockListingRepo,
		mockAccessControl,
		mockEventService,
	)

	// Test data - hidden visibility but viewer has own bids
	listing := createTestListing(marketplace.AuctionTypeClosed, marketplace.BidVisibilityHidden, marketplace.VisibilityPublic)
	bids := []*marketplace.Bid{
		createTestBid("VIEWER_1", 1000, false), // Viewer's own bid
		createTestBid("USER_2", 1100, false),
		createTestBid("USER_3", 1200, true),
	}

	// Mock expectations
	mockAccessControl.On("CanViewListing", mock.Anything, listing, "VIEWER_1", "ORG_123").Return(true)

	// Execute
	ctx := context.Background()
	result, err := service.ApplyBidVisibilityRules(ctx, listing, bids, "VIEWER_1", "ORG_123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Find the viewer's own bid
	var ownBid *marketplaceService.FilteredBid
	for _, filteredBid := range result.Bids {
		if filteredBid.IsOwnBid {
			ownBid = filteredBid
			break
		}
	}

	// Own bid should be fully visible regardless of visibility settings
	assert.NotNil(t, ownBid)
	assert.Equal(t, marketplaceService.VisibilityLevelOwn, ownBid.VisibilityLevel)
	assert.NotNil(t, ownBid.BidAmount)
	assert.Equal(t, "Your Bid", ownBid.AnonymousID)

	mockAccessControl.AssertExpectations(t)
}

func TestBidVisibilityService_GetVisibleBidStatistics_FullVisibility(t *testing.T) {
	// Setup
	mockBidRepo := new(MockBidRepository)
	mockListingRepo := new(MockListingRepository)
	mockAccessControl := new(MockAccessControlService)
	mockEventService := new(MockEventService)

	service := marketplaceService.NewBidVisibilityService(
		mockBidRepo,
		mockListingRepo,
		mockAccessControl,
		mockEventService,
	)

	// Test data
	listing := createTestListing(marketplace.AuctionTypeOpen, marketplace.BidVisibilityFull, marketplace.VisibilityPublic)
	stats := createTestBidStatistics()

	// Mock expectations
	mockListingRepo.On("GetByListingID", mock.Anything, "LST_123").Return(listing, nil)
	mockAccessControl.On("CanViewListing", mock.Anything, listing, "VIEWER_1", "ORG_123").Return(true)
	mockBidRepo.On("GetBidStatistics", mock.Anything, "LST_123").Return(stats, nil)

	// Execute
	ctx := context.Background()
	result, err := service.GetVisibleBidStatistics(ctx, "LST_123", "VIEWER_1", "ORG_123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "LST_123", result.ListingID)
	assert.Equal(t, 5, result.TotalBids)
	assert.Equal(t, marketplaceService.VisibilityLevelFull, result.VisibilityLevel)

	// All statistics should be visible
	assert.NotNil(t, result.UniqueBidders)
	assert.Equal(t, 3, *result.UniqueBidders)
	assert.NotNil(t, result.HighestBid)
	assert.Equal(t, "1200", *result.HighestBid)
	assert.NotNil(t, result.AverageBid)
	assert.NotNil(t, result.AutoBidCount)

	mockListingRepo.AssertExpectations(t)
	mockAccessControl.AssertExpectations(t)
	mockBidRepo.AssertExpectations(t)
}

func TestBidVisibilityService_GetVisibleBidStatistics_HiddenVisibility(t *testing.T) {
	// Setup
	mockBidRepo := new(MockBidRepository)
	mockListingRepo := new(MockListingRepository)
	mockAccessControl := new(MockAccessControlService)
	mockEventService := new(MockEventService)

	service := marketplaceService.NewBidVisibilityService(
		mockBidRepo,
		mockListingRepo,
		mockAccessControl,
		mockEventService,
	)

	// Test data
	listing := createTestListing(marketplace.AuctionTypeClosed, marketplace.BidVisibilityHidden, marketplace.VisibilityPublic)
	stats := createTestBidStatistics()

	// Mock expectations
	mockListingRepo.On("GetByListingID", mock.Anything, "LST_123").Return(listing, nil)
	mockAccessControl.On("CanViewListing", mock.Anything, listing, "VIEWER_1", "ORG_123").Return(true)
	mockBidRepo.On("GetBidStatistics", mock.Anything, "LST_123").Return(stats, nil)

	// Execute
	ctx := context.Background()
	result, err := service.GetVisibleBidStatistics(ctx, "LST_123", "VIEWER_1", "ORG_123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "LST_123", result.ListingID)
	assert.Equal(t, 0, result.TotalBids) // Should be hidden
	assert.Equal(t, marketplaceService.VisibilityLevelHidden, result.VisibilityLevel)

	// All detailed statistics should be hidden
	assert.Nil(t, result.UniqueBidders)
	assert.Nil(t, result.HighestBid)
	assert.Nil(t, result.AverageBid)
	assert.Nil(t, result.AutoBidCount)

	mockListingRepo.AssertExpectations(t)
	mockAccessControl.AssertExpectations(t)
	mockBidRepo.AssertExpectations(t)
}

func TestBidVisibilityService_CanViewBidDetails_OwnerCanAlwaysView(t *testing.T) {
	// Setup
	mockBidRepo := new(MockBidRepository)
	mockListingRepo := new(MockListingRepository)
	mockAccessControl := new(MockAccessControlService)
	mockEventService := new(MockEventService)

	service := marketplaceService.NewBidVisibilityService(
		mockBidRepo,
		mockListingRepo,
		mockAccessControl,
		mockEventService,
	)

	// Test data
	listing := createTestListing(marketplace.AuctionTypeClosed, marketplace.BidVisibilityHidden, marketplace.VisibilityPublic)
	bid := createTestBid("VIEWER_1", 1000, false)

	// Execute
	ctx := context.Background()
	canView := service.CanViewBidDetails(ctx, listing, bid, "VIEWER_1", "ORG_123")

	// Assert - owner should always be able to view their own bid details
	assert.True(t, canView)
}

func TestBidVisibilityService_CanViewBidDetails_NonOwnerHiddenVisibility(t *testing.T) {
	// Setup
	mockBidRepo := new(MockBidRepository)
	mockListingRepo := new(MockListingRepository)
	mockAccessControl := new(MockAccessControlService)
	mockEventService := new(MockEventService)

	service := marketplaceService.NewBidVisibilityService(
		mockBidRepo,
		mockListingRepo,
		mockAccessControl,
		mockEventService,
	)

	// Test data
	listing := createTestListing(marketplace.AuctionTypeClosed, marketplace.BidVisibilityHidden, marketplace.VisibilityPublic)
	bid := createTestBid("OTHER_USER", 1000, false)

	// Mock expectations
	mockAccessControl.On("CanViewListing", mock.Anything, listing, "VIEWER_1", "ORG_123").Return(true)

	// Execute
	ctx := context.Background()
	canView := service.CanViewBidDetails(ctx, listing, bid, "VIEWER_1", "ORG_123")

	// Assert - non-owner should not be able to view bid details with hidden visibility
	assert.False(t, canView)

	mockAccessControl.AssertExpectations(t)
}

func TestBidVisibilityService_GetAuctionVisibilityState_ActiveAuction(t *testing.T) {
	// Setup
	mockBidRepo := new(MockBidRepository)
	mockListingRepo := new(MockListingRepository)
	mockAccessControl := new(MockAccessControlService)
	mockEventService := new(MockEventService)

	service := marketplaceService.NewBidVisibilityService(
		mockBidRepo,
		mockListingRepo,
		mockAccessControl,
		mockEventService,
	)

	// Test data - active auction
	listing := createTestListing(marketplace.AuctionTypeOpen, marketplace.BidVisibilityFull, marketplace.VisibilityPublic)

	// Execute
	ctx := context.Background()
	state, err := service.GetAuctionVisibilityState(ctx, listing, "VIEWER_1", "ORG_123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, state)
	assert.True(t, state.IsActive)
	assert.False(t, state.IsExpired)
	assert.Equal(t, marketplace.AuctionTypeOpen, state.AuctionType)
	assert.Equal(t, marketplace.BidVisibilityFull, state.BidVisibility)
	assert.NotNil(t, state.TimeRemaining)
	assert.True(t, state.CanRevealBids) // Open auction should allow bid revelation
	assert.Equal(t, "open_auction", state.RevealReason)
}

func TestBidVisibilityService_GetAuctionVisibilityState_ExpiredAuction(t *testing.T) {
	// Setup
	mockBidRepo := new(MockBidRepository)
	mockListingRepo := new(MockListingRepository)
	mockAccessControl := new(MockAccessControlService)
	mockEventService := new(MockEventService)

	service := marketplaceService.NewBidVisibilityService(
		mockBidRepo,
		mockListingRepo,
		mockAccessControl,
		mockEventService,
	)

	// Test data - expired auction
	listing := createTestListing(marketplace.AuctionTypeClosed, marketplace.BidVisibilityHidden, marketplace.VisibilityPublic)
	listing.ExpiresAt = time.Now().Add(-1 * time.Hour) // Expired 1 hour ago
	listing.Status = marketplace.ListingStatusExpired

	// Execute
	ctx := context.Background()
	state, err := service.GetAuctionVisibilityState(ctx, listing, "VIEWER_1", "ORG_123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, state)
	assert.False(t, state.IsActive)
	assert.True(t, state.IsExpired)
	assert.True(t, state.CanRevealBids) // Expired auction should allow bid revelation
	assert.Equal(t, "auction_ended", state.RevealReason)
	assert.Nil(t, state.TimeRemaining)
}

func TestBidVisibilityService_CanViewListing_PublicVisibility(t *testing.T) {
	// Setup
	mockBidRepo := new(MockBidRepository)
	mockListingRepo := new(MockListingRepository)
	mockAccessControl := new(MockAccessControlService)
	mockEventService := new(MockEventService)

	service := marketplaceService.NewBidVisibilityService(
		mockBidRepo,
		mockListingRepo,
		mockAccessControl,
		mockEventService,
	)

	// Test data
	listing := createTestListing(marketplace.AuctionTypeOpen, marketplace.BidVisibilityFull, marketplace.VisibilityPublic)

	// Mock expectations
	mockAccessControl.On("CanViewListing", mock.Anything, listing, "VIEWER_1", "ORG_123").Return(true)

	// Execute
	ctx := context.Background()
	canView := service.CanViewListing(ctx, listing, "VIEWER_1", "ORG_123")

	// Assert
	assert.True(t, canView)

	mockAccessControl.AssertExpectations(t)
}

func TestBidVisibilityService_CanViewListing_PrivateVisibilityDenied(t *testing.T) {
	// Setup
	mockBidRepo := new(MockBidRepository)
	mockListingRepo := new(MockListingRepository)
	mockAccessControl := new(MockAccessControlService)
	mockEventService := new(MockEventService)

	service := marketplaceService.NewBidVisibilityService(
		mockBidRepo,
		mockListingRepo,
		mockAccessControl,
		mockEventService,
	)

	// Test data
	listing := createTestListing(marketplace.AuctionTypeOpen, marketplace.BidVisibilityFull, marketplace.VisibilityPrivate)

	// Mock expectations
	mockAccessControl.On("CanViewListing", mock.Anything, listing, "VIEWER_1", "ORG_456").Return(false)

	// Execute
	ctx := context.Background()
	canView := service.CanViewListing(ctx, listing, "VIEWER_1", "ORG_456")

	// Assert
	assert.False(t, canView)

	mockAccessControl.AssertExpectations(t)
}
