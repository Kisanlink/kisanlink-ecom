package marketplace

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// AuctionAutomationServiceInterface defines the interface for auction automation operations
type AuctionAutomationServiceInterface interface {
	// Service lifecycle
	Start(ctx context.Context) error
	Stop() error
	IsRunning() bool

	// Configuration
	Configure(config *AutomationConfig) error
	GetConfiguration() *AutomationConfig

	// Manual operations
	ProcessExpiredAuctionsNow(ctx context.Context) (*AuctionProcessingResult, error)
	CleanupCompletedAuctionsNow(ctx context.Context) (*CleanupResult, error)

	// Status and monitoring
	GetStatus() *AutomationStatus
	GetMetrics() *AutomationMetrics
}

// AutomationConfig holds configuration for auction automation
type AutomationConfig struct {
	// Expiry processing configuration
	ExpiryProcessingEnabled  bool          `json:"expiry_processing_enabled"`
	ExpiryProcessingInterval time.Duration `json:"expiry_processing_interval"`
	ExpiryBatchSize          int           `json:"expiry_batch_size"`

	// Cleanup configuration
	CleanupEnabled         bool          `json:"cleanup_enabled"`
	CleanupInterval        time.Duration `json:"cleanup_interval"`
	CleanupRetentionPeriod time.Duration `json:"cleanup_retention_period"`

	// Notification configuration
	NotificationsEnabled bool                  `json:"notifications_enabled"`
	NotificationChannels []NotificationChannel `json:"notification_channels"`

	// Performance configuration
	MaxConcurrentJobs int           `json:"max_concurrent_jobs"`
	JobTimeout        time.Duration `json:"job_timeout"`
	RetryAttempts     int           `json:"retry_attempts"`
	RetryDelay        time.Duration `json:"retry_delay"`
}

// DefaultAutomationConfig returns the default automation configuration
func DefaultAutomationConfig() *AutomationConfig {
	return &AutomationConfig{
		ExpiryProcessingEnabled:  true,
		ExpiryProcessingInterval: 5 * time.Minute,
		ExpiryBatchSize:          50,
		CleanupEnabled:           true,
		CleanupInterval:          24 * time.Hour,
		CleanupRetentionPeriod:   30 * 24 * time.Hour, // 30 days
		NotificationsEnabled:     true,
		NotificationChannels:     []NotificationChannel{NotificationChannelEmail, NotificationChannelInApp},
		MaxConcurrentJobs:        5,
		JobTimeout:               30 * time.Minute,
		RetryAttempts:            3,
		RetryDelay:               30 * time.Second,
	}
}

// AutomationStatus represents the current status of automation services
type AutomationStatus struct {
	IsRunning               bool       `json:"is_running"`
	StartTime               *time.Time `json:"start_time,omitempty"`
	ExpiryProcessingEnabled bool       `json:"expiry_processing_enabled"`
	CleanupEnabled          bool       `json:"cleanup_enabled"`
	NotificationsEnabled    bool       `json:"notifications_enabled"`
	ActiveJobs              int        `json:"active_jobs"`
	LastExpiryProcessing    *time.Time `json:"last_expiry_processing,omitempty"`
	LastCleanup             *time.Time `json:"last_cleanup,omitempty"`
	NextScheduledExpiry     *time.Time `json:"next_scheduled_expiry,omitempty"`
	NextScheduledCleanup    *time.Time `json:"next_scheduled_cleanup,omitempty"`
}

// AutomationMetrics represents metrics for automation operations
type AutomationMetrics struct {
	// Expiry processing metrics
	TotalExpiryRuns         int64         `json:"total_expiry_runs"`
	TotalAuctionsProcessed  int64         `json:"total_auctions_processed"`
	TotalAuctionsSuccessful int64         `json:"total_auctions_successful"`
	TotalAuctionsFailed     int64         `json:"total_auctions_failed"`
	AverageProcessingTime   time.Duration `json:"average_processing_time"`

	// Cleanup metrics
	TotalCleanupRuns   int64         `json:"total_cleanup_runs"`
	TotalItemsCleaned  int64         `json:"total_items_cleaned"`
	AverageCleanupTime time.Duration `json:"average_cleanup_time"`

	// Notification metrics
	TotalNotificationsSent   int64 `json:"total_notifications_sent"`
	TotalNotificationsFailed int64 `json:"total_notifications_failed"`

	// Error metrics
	TotalErrors   int64      `json:"total_errors"`
	LastError     string     `json:"last_error,omitempty"`
	LastErrorTime *time.Time `json:"last_error_time,omitempty"`
}

// AuctionAutomationService provides comprehensive auction automation
type AuctionAutomationService struct {
	config               *AutomationConfig
	lifecycleService     AuctionLifecycleServiceInterface
	cleanupService       AuctionCleanupServiceInterface
	notificationService  AuctionNotificationServiceInterface
	backgroundJobService BackgroundJobServiceInterface
	logger               *logrus.Logger

	// Service state
	running   bool
	startTime *time.Time
	metrics   *AutomationMetrics
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewAuctionAutomationService creates a new auction automation service
func NewAuctionAutomationService(
	lifecycleService AuctionLifecycleServiceInterface,
	cleanupService AuctionCleanupServiceInterface,
	notificationService AuctionNotificationServiceInterface,
	logger *logrus.Logger,
) AuctionAutomationServiceInterface {
	if logger == nil {
		logger = logrus.New()
	}

	backgroundJobService := NewBackgroundJobService(logger)

	return &AuctionAutomationService{
		config:               DefaultAutomationConfig(),
		lifecycleService:     lifecycleService,
		cleanupService:       cleanupService,
		notificationService:  notificationService,
		backgroundJobService: backgroundJobService,
		logger:               logger,
		running:              false,
		metrics: &AutomationMetrics{
			TotalExpiryRuns:          0,
			TotalAuctionsProcessed:   0,
			TotalAuctionsSuccessful:  0,
			TotalAuctionsFailed:      0,
			TotalCleanupRuns:         0,
			TotalItemsCleaned:        0,
			TotalNotificationsSent:   0,
			TotalNotificationsFailed: 0,
			TotalErrors:              0,
		},
	}
}

// Start begins the auction automation services
func (s *AuctionAutomationService) Start(ctx context.Context) error {
	if s.running {
		return fmt.Errorf("auction automation service is already running")
	}

	s.logger.Info("Starting auction automation service")

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.running = true
	now := time.Now()
	s.startTime = &now

	// Start background job service
	if err := s.backgroundJobService.Start(s.ctx); err != nil {
		return fmt.Errorf("failed to start background job service: %w", err)
	}

	// Register auction expiry job
	if s.config.ExpiryProcessingEnabled {
		expiryJob := NewAuctionExpiryJob(
			s.lifecycleService,
			s.config.ExpiryProcessingInterval,
			s.logger,
		)
		if err := s.backgroundJobService.RegisterJob("auction_expiry", expiryJob); err != nil {
			s.logger.WithError(err).Error("Failed to register auction expiry job")
		}
	}

	// Register auction cleanup job
	if s.config.CleanupEnabled {
		cleanupJob := NewAuctionCleanupJob(
			s.cleanupService,
			s.config.CleanupInterval,
			s.config.CleanupRetentionPeriod,
			s.logger,
		)
		if err := s.backgroundJobService.RegisterJob("auction_cleanup", cleanupJob); err != nil {
			s.logger.WithError(err).Error("Failed to register auction cleanup job")
		}
	}

	s.logger.WithFields(logrus.Fields{
		"expiry_processing_enabled": s.config.ExpiryProcessingEnabled,
		"cleanup_enabled":           s.config.CleanupEnabled,
		"notifications_enabled":     s.config.NotificationsEnabled,
		"expiry_interval":           s.config.ExpiryProcessingInterval,
		"cleanup_interval":          s.config.CleanupInterval,
	}).Info("Auction automation service started successfully")

	return nil
}

// Stop stops the auction automation services
func (s *AuctionAutomationService) Stop() error {
	if !s.running {
		return fmt.Errorf("auction automation service is not running")
	}

	s.logger.Info("Stopping auction automation service")

	// Stop background job service
	if err := s.backgroundJobService.Stop(); err != nil {
		s.logger.WithError(err).Error("Failed to stop background job service")
	}

	// Cancel context
	if s.cancel != nil {
		s.cancel()
	}

	s.running = false
	s.startTime = nil

	s.logger.Info("Auction automation service stopped successfully")
	return nil
}

// IsRunning returns whether the automation service is currently running
func (s *AuctionAutomationService) IsRunning() bool {
	return s.running
}

// Configure updates the automation configuration
func (s *AuctionAutomationService) Configure(config *AutomationConfig) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	s.logger.WithFields(logrus.Fields{
		"old_expiry_interval":  s.config.ExpiryProcessingInterval,
		"new_expiry_interval":  config.ExpiryProcessingInterval,
		"old_cleanup_interval": s.config.CleanupInterval,
		"new_cleanup_interval": config.CleanupInterval,
	}).Info("Updating automation configuration")

	s.config = config

	// If running, restart with new configuration
	if s.running {
		s.logger.Info("Restarting automation service with new configuration")
		ctx := s.ctx
		if err := s.Stop(); err != nil {
			return fmt.Errorf("failed to stop service for reconfiguration: %w", err)
		}
		if err := s.Start(ctx); err != nil {
			return fmt.Errorf("failed to restart service with new configuration: %w", err)
		}
	}

	return nil
}

// GetConfiguration returns the current automation configuration
func (s *AuctionAutomationService) GetConfiguration() *AutomationConfig {
	// Return a copy to prevent external modification
	configCopy := *s.config
	return &configCopy
}

// ProcessExpiredAuctionsNow manually triggers processing of expired auctions
func (s *AuctionAutomationService) ProcessExpiredAuctionsNow(ctx context.Context) (*AuctionProcessingResult, error) {
	s.logger.Info("Manual processing of expired auctions triggered")

	startTime := time.Now()
	result, err := s.lifecycleService.ProcessExpiredAuctions(ctx)

	// Update metrics
	s.metrics.TotalExpiryRuns++
	if err != nil {
		s.metrics.TotalErrors++
		s.metrics.LastError = err.Error()
		s.metrics.LastErrorTime = &startTime
	} else {
		s.metrics.TotalAuctionsProcessed += int64(result.ProcessedCount)
		s.metrics.TotalAuctionsSuccessful += int64(result.SuccessfulCount)
		s.metrics.TotalAuctionsFailed += int64(result.FailedCount)

		// Update average processing time
		if s.metrics.TotalExpiryRuns > 0 {
			totalTime := s.metrics.AverageProcessingTime * time.Duration(s.metrics.TotalExpiryRuns-1)
			s.metrics.AverageProcessingTime = (totalTime + result.ProcessingTime) / time.Duration(s.metrics.TotalExpiryRuns)
		} else {
			s.metrics.AverageProcessingTime = result.ProcessingTime
		}
	}

	return result, err
}

// CleanupCompletedAuctionsNow manually triggers cleanup of completed auctions
func (s *AuctionAutomationService) CleanupCompletedAuctionsNow(ctx context.Context) (*CleanupResult, error) {
	s.logger.Info("Manual cleanup of completed auctions triggered")

	startTime := time.Now()
	result, err := s.cleanupService.CleanupCompletedAuctions(ctx, s.config.CleanupRetentionPeriod)

	// Update metrics
	s.metrics.TotalCleanupRuns++
	if err != nil {
		s.metrics.TotalErrors++
		s.metrics.LastError = err.Error()
		s.metrics.LastErrorTime = &startTime
	} else {
		s.metrics.TotalItemsCleaned += int64(result.CleanedCount)

		// Update average cleanup time
		if s.metrics.TotalCleanupRuns > 0 {
			totalTime := s.metrics.AverageCleanupTime * time.Duration(s.metrics.TotalCleanupRuns-1)
			s.metrics.AverageCleanupTime = (totalTime + result.ProcessingTime) / time.Duration(s.metrics.TotalCleanupRuns)
		} else {
			s.metrics.AverageCleanupTime = result.ProcessingTime
		}
	}

	return result, err
}

// GetStatus returns the current status of automation services
func (s *AuctionAutomationService) GetStatus() *AutomationStatus {
	status := &AutomationStatus{
		IsRunning:               s.running,
		StartTime:               s.startTime,
		ExpiryProcessingEnabled: s.config.ExpiryProcessingEnabled,
		CleanupEnabled:          s.config.CleanupEnabled,
		NotificationsEnabled:    s.config.NotificationsEnabled,
		ActiveJobs:              0,
	}

	// Get job statuses from background service
	if s.backgroundJobService != nil {
		jobStatuses := s.backgroundJobService.GetAllJobStatuses()

		for _, jobStatus := range jobStatuses {
			if jobStatus.Running {
				status.ActiveJobs++
			}

			// Set last run times and next scheduled times
			if jobStatus.Name == "auction_expiry" {
				status.LastExpiryProcessing = jobStatus.LastRun
				status.NextScheduledExpiry = jobStatus.NextScheduledRun
			} else if jobStatus.Name == "auction_cleanup" {
				status.LastCleanup = jobStatus.LastRun
				status.NextScheduledCleanup = jobStatus.NextScheduledRun
			}
		}
	}

	return status
}

// GetMetrics returns current automation metrics
func (s *AuctionAutomationService) GetMetrics() *AutomationMetrics {
	// Return a copy to prevent external modification
	metricsCopy := *s.metrics
	return &metricsCopy
}

// Helper method to update notification metrics
func (s *AuctionAutomationService) updateNotificationMetrics(sent int64, failed int64) {
	s.metrics.TotalNotificationsSent += sent
	s.metrics.TotalNotificationsFailed += failed
}
