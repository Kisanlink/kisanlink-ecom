package marketplace

import (
	"context"
	"testing"
	"time"

	marketplaceModels "github.com/Kisanlink/kisanlink-ecom/entities/models/marketplace"
	"github.com/Kisanlink/kisanlink-ecom/internal/services/marketplace"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuctionLifecycleService is a mock implementation of AuctionLifecycleServiceInterface
type MockAuctionLifecycleService struct {
	mock.Mock
}

func (m *MockAuctionLifecycleService) ProcessExpiredAuctions(ctx context.Context) (*marketplace.AuctionProcessingResult, error) {
	args := m.Called(ctx)
	return args.Get(0).(*marketplace.AuctionProcessingResult), args.Error(1)
}

func (m *MockAuctionLifecycleService) ProcessSingleExpiredAuction(ctx context.Context, listingID string) (*marketplace.AuctionResult, error) {
	args := m.Called(ctx, listingID)
	return args.Get(0).(*marketplace.AuctionResult), args.Error(1)
}

func (m *MockAuctionLifecycleService) CloseAuction(ctx context.Context, listingID string, reason string) (*marketplace.AuctionResult, error) {
	args := m.Called(ctx, listingID, reason)
	return args.Get(0).(*marketplace.AuctionResult), args.Error(1)
}

func (m *MockAuctionLifecycleService) DetermineWinner(ctx context.Context, listingID string) (*marketplaceModels.Bid, error) {
	args := m.Called(ctx, listingID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*marketplaceModels.Bid), args.Error(1)
}

func (m *MockAuctionLifecycleService) CreateOrderFromWinningBid(ctx context.Context, winningBid *marketplaceModels.Bid, listing *marketplaceModels.Listing) (*marketplace.OrderCreationResult, error) {
	args := m.Called(ctx, winningBid, listing)
	return args.Get(0).(*marketplace.OrderCreationResult), args.Error(1)
}

func (m *MockAuctionLifecycleService) StartAuctionExpiryScheduler(ctx context.Context, interval time.Duration) error {
	args := m.Called(ctx, interval)
	return args.Error(0)
}

func (m *MockAuctionLifecycleService) StopAuctionExpiryScheduler() error {
	args := m.Called()
	return args.Error(0)
}

// MockAuctionCleanupService is a mock implementation of AuctionCleanupServiceInterface
type MockAuctionCleanupService struct {
	mock.Mock
}

func (m *MockAuctionCleanupService) CleanupCompletedAuctions(ctx context.Context, olderThan time.Duration) (*marketplace.CleanupResult, error) {
	args := m.Called(ctx, olderThan)
	return args.Get(0).(*marketplace.CleanupResult), args.Error(1)
}

func (m *MockAuctionCleanupService) CleanupExpiredBids(ctx context.Context, olderThan time.Duration) (*marketplace.CleanupResult, error) {
	args := m.Called(ctx, olderThan)
	return args.Get(0).(*marketplace.CleanupResult), args.Error(1)
}

func (m *MockAuctionCleanupService) CleanupAuctionEvents(ctx context.Context, olderThan time.Duration) (*marketplace.CleanupResult, error) {
	args := m.Called(ctx, olderThan)
	return args.Get(0).(*marketplace.CleanupResult), args.Error(1)
}

func (m *MockAuctionCleanupService) ArchiveCompletedAuctions(ctx context.Context, olderThan time.Duration) (*marketplace.ArchiveResult, error) {
	args := m.Called(ctx, olderThan)
	return args.Get(0).(*marketplace.ArchiveResult), args.Error(1)
}

func (m *MockAuctionCleanupService) OptimizeDatabase(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAuctionCleanupService) UpdateStatistics(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockAuctionNotificationService is a mock implementation of AuctionNotificationServiceInterface
type MockAuctionNotificationService struct {
	mock.Mock
}

func (m *MockAuctionNotificationService) NotifyAuctionWinner(ctx context.Context, listing *marketplaceModels.Listing, winningBid *marketplaceModels.Bid) error {
	args := m.Called(ctx, listing, winningBid)
	return args.Error(0)
}

func (m *MockAuctionNotificationService) NotifyAuctionLosers(ctx context.Context, listing *marketplaceModels.Listing, losingBids []*marketplaceModels.Bid) error {
	args := m.Called(ctx, listing, losingBids)
	return args.Error(0)
}

func (m *MockAuctionNotificationService) NotifySellerAuctionCompleted(ctx context.Context, listing *marketplaceModels.Listing, winningBid *marketplaceModels.Bid) error {
	args := m.Called(ctx, listing, winningBid)
	return args.Error(0)
}

func (m *MockAuctionNotificationService) NotifySellerAuctionExpiredNoBids(ctx context.Context, listing *marketplaceModels.Listing) error {
	args := m.Called(ctx, listing)
	return args.Error(0)
}

func (m *MockAuctionNotificationService) NotifyBidderOutbid(ctx context.Context, listing *marketplaceModels.Listing, outbidBid *marketplaceModels.Bid, newHighestBid *marketplaceModels.Bid) error {
	args := m.Called(ctx, listing, outbidBid, newHighestBid)
	return args.Error(0)
}

func (m *MockAuctionNotificationService) NotifyBidderAuctionEnding(ctx context.Context, listing *marketplaceModels.Listing, bid *marketplaceModels.Bid, timeRemaining time.Duration) error {
	args := m.Called(ctx, listing, bid, timeRemaining)
	return args.Error(0)
}

func (m *MockAuctionNotificationService) NotifyOrderCreated(ctx context.Context, orderID string, listing *marketplaceModels.Listing, winningBid *marketplaceModels.Bid) error {
	args := m.Called(ctx, orderID, listing, winningBid)
	return args.Error(0)
}

func (m *MockAuctionNotificationService) SendBatchNotifications(ctx context.Context, notifications []*marketplace.AuctionNotification) error {
	args := m.Called(ctx, notifications)
	return args.Error(0)
}

func TestAuctionAutomationService_StartStop(t *testing.T) {
	// Create mock services
	mockLifecycle := &MockAuctionLifecycleService{}
	mockCleanup := &MockAuctionCleanupService{}
	mockNotification := &MockAuctionNotificationService{}

	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce log noise in tests

	// Create automation service
	automationService := marketplace.NewAuctionAutomationService(
		mockLifecycle,
		mockCleanup,
		mockNotification,
		logger,
	)

	// Test initial state
	assert.False(t, automationService.IsRunning())

	// Test start
	ctx := context.Background()
	err := automationService.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, automationService.IsRunning())

	// Test double start (should fail)
	err = automationService.Start(ctx)
	assert.Error(t, err)

	// Test stop
	err = automationService.Stop()
	assert.NoError(t, err)
	assert.False(t, automationService.IsRunning())

	// Test double stop (should fail)
	err = automationService.Stop()
	assert.Error(t, err)
}

func TestAuctionAutomationService_Configuration(t *testing.T) {
	// Create mock services
	mockLifecycle := &MockAuctionLifecycleService{}
	mockCleanup := &MockAuctionCleanupService{}
	mockNotification := &MockAuctionNotificationService{}

	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Create automation service
	automationService := marketplace.NewAuctionAutomationService(
		mockLifecycle,
		mockCleanup,
		mockNotification,
		logger,
	)

	// Test default configuration
	config := automationService.GetConfiguration()
	assert.NotNil(t, config)
	assert.True(t, config.ExpiryProcessingEnabled)
	assert.True(t, config.CleanupEnabled)
	assert.True(t, config.NotificationsEnabled)

	// Test configuration update
	newConfig := &marketplace.AutomationConfig{
		ExpiryProcessingEnabled:  false,
		ExpiryProcessingInterval: 10 * time.Minute,
		ExpiryBatchSize:          100,
		CleanupEnabled:           false,
		CleanupInterval:          48 * time.Hour,
		CleanupRetentionPeriod:   60 * 24 * time.Hour,
		NotificationsEnabled:     false,
		NotificationChannels:     []marketplace.NotificationChannel{marketplace.NotificationChannelEmail},
		MaxConcurrentJobs:        10,
		JobTimeout:               60 * time.Minute,
		RetryAttempts:            5,
		RetryDelay:               60 * time.Second,
	}

	err := automationService.Configure(newConfig)
	assert.NoError(t, err)

	// Verify configuration was updated
	updatedConfig := automationService.GetConfiguration()
	assert.Equal(t, newConfig.ExpiryProcessingEnabled, updatedConfig.ExpiryProcessingEnabled)
	assert.Equal(t, newConfig.ExpiryProcessingInterval, updatedConfig.ExpiryProcessingInterval)
	assert.Equal(t, newConfig.CleanupEnabled, updatedConfig.CleanupEnabled)
	assert.Equal(t, newConfig.NotificationsEnabled, updatedConfig.NotificationsEnabled)

	// Test nil configuration (should fail)
	err = automationService.Configure(nil)
	assert.Error(t, err)
}

func TestAuctionAutomationService_ManualOperations(t *testing.T) {
	// Create mock services
	mockLifecycle := &MockAuctionLifecycleService{}
	mockCleanup := &MockAuctionCleanupService{}
	mockNotification := &MockAuctionNotificationService{}

	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Create automation service
	automationService := marketplace.NewAuctionAutomationService(
		mockLifecycle,
		mockCleanup,
		mockNotification,
		logger,
	)

	ctx := context.Background()

	// Test manual expiry processing
	expectedResult := &marketplace.AuctionProcessingResult{
		ProcessedCount:  5,
		SuccessfulCount: 4,
		FailedCount:     1,
		ProcessingTime:  100 * time.Millisecond,
		Results:         []*marketplace.AuctionResult{},
		Errors:          []string{},
	}

	mockLifecycle.On("ProcessExpiredAuctions", ctx).Return(expectedResult, nil)

	result, err := automationService.ProcessExpiredAuctionsNow(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)

	// Test manual cleanup
	expectedCleanupResult := &marketplace.CleanupResult{
		Operation:      "cleanup_completed_auctions",
		ProcessedCount: 10,
		CleanedCount:   8,
		ErrorCount:     2,
		ProcessingTime: 200 * time.Millisecond,
		Errors:         []string{},
	}

	mockCleanup.On("CleanupCompletedAuctions", ctx, mock.AnythingOfType("time.Duration")).Return(expectedCleanupResult, nil)

	cleanupResult, err := automationService.CleanupCompletedAuctionsNow(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedCleanupResult, cleanupResult)

	// Verify mock expectations
	mockLifecycle.AssertExpectations(t)
	mockCleanup.AssertExpectations(t)
}

func TestAuctionAutomationService_StatusAndMetrics(t *testing.T) {
	// Create mock services
	mockLifecycle := &MockAuctionLifecycleService{}
	mockCleanup := &MockAuctionCleanupService{}
	mockNotification := &MockAuctionNotificationService{}

	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Create automation service
	automationService := marketplace.NewAuctionAutomationService(
		mockLifecycle,
		mockCleanup,
		mockNotification,
		logger,
	)

	// Test initial status
	status := automationService.GetStatus()
	assert.NotNil(t, status)
	assert.False(t, status.IsRunning)
	assert.Nil(t, status.StartTime)

	// Test initial metrics
	metrics := automationService.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(0), metrics.TotalExpiryRuns)
	assert.Equal(t, int64(0), metrics.TotalAuctionsProcessed)

	// Start service and check status
	ctx := context.Background()
	err := automationService.Start(ctx)
	assert.NoError(t, err)

	status = automationService.GetStatus()
	assert.True(t, status.IsRunning)
	assert.NotNil(t, status.StartTime)

	// Stop service
	err = automationService.Stop()
	assert.NoError(t, err)

	status = automationService.GetStatus()
	assert.False(t, status.IsRunning)
}
