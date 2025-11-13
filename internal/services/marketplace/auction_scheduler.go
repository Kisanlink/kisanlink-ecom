package marketplace

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// AuctionSchedulerInterface defines the interface for auction scheduling operations
type AuctionSchedulerInterface interface {
	// Scheduler lifecycle
	Start(ctx context.Context) error
	Stop() error
	IsRunning() bool

	// Configuration
	SetInterval(interval time.Duration)
	GetInterval() time.Duration
	SetBatchSize(size int)
	GetBatchSize() int

	// Manual processing
	ProcessExpiredAuctionsNow(ctx context.Context) (*AuctionProcessingResult, error)

	// Status and metrics
	GetLastProcessingResult() *AuctionProcessingResult
	GetProcessingStats() *SchedulerStats
}

// SchedulerConfig holds configuration for the auction scheduler
type SchedulerConfig struct {
	ProcessingInterval time.Duration `json:"processing_interval"`
	BatchSize          int           `json:"batch_size"`
	MaxRetries         int           `json:"max_retries"`
	RetryDelay         time.Duration `json:"retry_delay"`
	EnableMetrics      bool          `json:"enable_metrics"`
}

// DefaultSchedulerConfig returns the default scheduler configuration
func DefaultSchedulerConfig() *SchedulerConfig {
	return &SchedulerConfig{
		ProcessingInterval: 5 * time.Minute,  // Process every 5 minutes
		BatchSize:          50,               // Process up to 50 auctions at once
		MaxRetries:         3,                // Retry failed auctions up to 3 times
		RetryDelay:         30 * time.Second, // Wait 30 seconds between retries
		EnableMetrics:      true,
	}
}

// SchedulerStats represents statistics about scheduler operations
type SchedulerStats struct {
	StartTime             time.Time     `json:"start_time"`
	TotalRuns             int64         `json:"total_runs"`
	TotalProcessed        int64         `json:"total_processed"`
	TotalSuccessful       int64         `json:"total_successful"`
	TotalFailed           int64         `json:"total_failed"`
	LastRunTime           time.Time     `json:"last_run_time"`
	LastProcessingTime    time.Duration `json:"last_processing_time"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	IsRunning             bool          `json:"is_running"`
}

// AuctionScheduler provides scheduled processing of expired auctions
type AuctionScheduler struct {
	config           *SchedulerConfig
	lifecycleService AuctionLifecycleServiceInterface
	logger           *logrus.Logger

	// Scheduler state
	running  bool
	stopChan chan struct{}
	ticker   *time.Ticker
	mutex    sync.RWMutex

	// Statistics and results
	stats               *SchedulerStats
	lastResult          *AuctionProcessingResult
	totalProcessingTime time.Duration
}

// NewAuctionScheduler creates a new auction scheduler
func NewAuctionScheduler(
	config *SchedulerConfig,
	lifecycleService AuctionLifecycleServiceInterface,
	logger *logrus.Logger,
) AuctionSchedulerInterface {
	if config == nil {
		config = DefaultSchedulerConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	return &AuctionScheduler{
		config:           config,
		lifecycleService: lifecycleService,
		logger:           logger,
		running:          false,
		stopChan:         make(chan struct{}),
		stats: &SchedulerStats{
			StartTime: time.Now(),
			IsRunning: false,
		},
	}
}

// Start begins the scheduled processing of expired auctions
func (s *AuctionScheduler) Start(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		return fmt.Errorf("auction scheduler is already running")
	}

	s.logger.WithFields(logrus.Fields{
		"interval":   s.config.ProcessingInterval,
		"batch_size": s.config.BatchSize,
	}).Info("Starting auction scheduler")

	s.running = true
	s.stats.IsRunning = true
	s.stats.StartTime = time.Now()
	s.ticker = time.NewTicker(s.config.ProcessingInterval)

	// Start the scheduler goroutine
	go s.schedulerLoop(ctx)

	s.logger.Info("Auction scheduler started successfully")
	return nil
}

// Stop stops the scheduled processing
func (s *AuctionScheduler) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return fmt.Errorf("auction scheduler is not running")
	}

	s.logger.Info("Stopping auction scheduler")

	// Signal stop and wait for goroutine to finish
	close(s.stopChan)
	if s.ticker != nil {
		s.ticker.Stop()
	}

	s.running = false
	s.stats.IsRunning = false

	s.logger.Info("Auction scheduler stopped successfully")
	return nil
}

// IsRunning returns whether the scheduler is currently running
func (s *AuctionScheduler) IsRunning() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.running
}

// SetInterval updates the processing interval
func (s *AuctionScheduler) SetInterval(interval time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.config.ProcessingInterval = interval

	// If running, restart the ticker with new interval
	if s.running && s.ticker != nil {
		s.ticker.Stop()
		s.ticker = time.NewTicker(interval)
	}

	s.logger.WithField("new_interval", interval).Info("Updated scheduler interval")
}

// GetInterval returns the current processing interval
func (s *AuctionScheduler) GetInterval() time.Duration {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.config.ProcessingInterval
}

// SetBatchSize updates the batch size for processing
func (s *AuctionScheduler) SetBatchSize(size int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.config.BatchSize = size
	s.logger.WithField("new_batch_size", size).Info("Updated scheduler batch size")
}

// GetBatchSize returns the current batch size
func (s *AuctionScheduler) GetBatchSize() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.config.BatchSize
}

// ProcessExpiredAuctionsNow manually triggers processing of expired auctions
func (s *AuctionScheduler) ProcessExpiredAuctionsNow(ctx context.Context) (*AuctionProcessingResult, error) {
	s.logger.Info("Manual processing of expired auctions triggered")
	return s.processExpiredAuctions(ctx)
}

// GetLastProcessingResult returns the result of the last processing run
func (s *AuctionScheduler) GetLastProcessingResult() *AuctionProcessingResult {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.lastResult
}

// GetProcessingStats returns current scheduler statistics
func (s *AuctionScheduler) GetProcessingStats() *SchedulerStats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Calculate average processing time
	avgProcessingTime := time.Duration(0)
	if s.stats.TotalRuns > 0 {
		avgProcessingTime = s.totalProcessingTime / time.Duration(s.stats.TotalRuns)
	}

	return &SchedulerStats{
		StartTime:             s.stats.StartTime,
		TotalRuns:             s.stats.TotalRuns,
		TotalProcessed:        s.stats.TotalProcessed,
		TotalSuccessful:       s.stats.TotalSuccessful,
		TotalFailed:           s.stats.TotalFailed,
		LastRunTime:           s.stats.LastRunTime,
		LastProcessingTime:    s.stats.LastProcessingTime,
		AverageProcessingTime: avgProcessingTime,
		IsRunning:             s.stats.IsRunning,
	}
}

// schedulerLoop is the main scheduler loop that runs in a goroutine
func (s *AuctionScheduler) schedulerLoop(ctx context.Context) {
	s.logger.Info("Scheduler loop started")

	for {
		select {
		case <-s.ticker.C:
			// Process expired auctions
			result, err := s.processExpiredAuctions(ctx)
			if err != nil {
				s.logger.WithError(err).Error("Failed to process expired auctions")
			} else if result.ProcessedCount > 0 {
				s.logger.WithFields(logrus.Fields{
					"processed":  result.ProcessedCount,
					"successful": result.SuccessfulCount,
					"failed":     result.FailedCount,
					"duration":   result.ProcessingTime,
				}).Info("Processed expired auctions")
			}

		case <-s.stopChan:
			s.logger.Info("Scheduler loop stopping")
			return

		case <-ctx.Done():
			s.logger.Info("Scheduler loop cancelled by context")
			return
		}
	}
}

// processExpiredAuctions processes expired auctions and updates statistics
func (s *AuctionScheduler) processExpiredAuctions(ctx context.Context) (*AuctionProcessingResult, error) {
	startTime := time.Now()

	// Process expired auctions through the lifecycle service
	result, err := s.lifecycleService.ProcessExpiredAuctions(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to process expired auctions")
		return nil, err
	}

	// Update statistics
	s.mutex.Lock()
	s.stats.TotalRuns++
	s.stats.TotalProcessed += int64(result.ProcessedCount)
	s.stats.TotalSuccessful += int64(result.SuccessfulCount)
	s.stats.TotalFailed += int64(result.FailedCount)
	s.stats.LastRunTime = startTime
	s.stats.LastProcessingTime = result.ProcessingTime
	s.totalProcessingTime += result.ProcessingTime
	s.lastResult = result
	s.mutex.Unlock()

	// Log processing results
	if result.ProcessedCount > 0 {
		s.logger.WithFields(logrus.Fields{
			"processed":       result.ProcessedCount,
			"successful":      result.SuccessfulCount,
			"failed":          result.FailedCount,
			"processing_time": result.ProcessingTime,
		}).Info("Completed processing expired auctions")

		// Log any errors
		if len(result.Errors) > 0 {
			for _, errMsg := range result.Errors {
				s.logger.WithField("error", errMsg).Warn("Auction processing error")
			}
		}
	}

	return result, nil
}

// Use AuctionNotificationServiceInterface from auction_notification_service.go

// Use NotificationRequest from notification_interfaces.go

// NotificationTypes constants for scheduler
const (
	NotificationTypeAuctionExpired = "AUCTION_EXPIRED"
	NotificationTypeOrderCreated   = "ORDER_CREATED"
)
