package marketplace

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// BackgroundJobServiceInterface defines the interface for background job operations
type BackgroundJobServiceInterface interface {
	// Job lifecycle
	Start(ctx context.Context) error
	Stop() error
	IsRunning() bool

	// Job management
	RegisterJob(name string, job BackgroundJob) error
	UnregisterJob(name string) error
	GetRegisteredJobs() []string

	// Manual execution
	ExecuteJob(ctx context.Context, jobName string) error
	ExecuteAllJobs(ctx context.Context) error

	// Status and monitoring
	GetJobStatus(jobName string) *JobStatus
	GetAllJobStatuses() map[string]*JobStatus
}

// BackgroundJob defines the interface for individual background jobs
type BackgroundJob interface {
	Execute(ctx context.Context) error
	GetName() string
	GetDescription() string
	GetInterval() time.Duration
	IsEnabled() bool
}

// JobStatus represents the status of a background job
type JobStatus struct {
	Name             string        `json:"name"`
	Description      string        `json:"description"`
	Enabled          bool          `json:"enabled"`
	Running          bool          `json:"running"`
	LastRun          *time.Time    `json:"last_run,omitempty"`
	LastDuration     time.Duration `json:"last_duration"`
	LastError        string        `json:"last_error,omitempty"`
	TotalRuns        int64         `json:"total_runs"`
	SuccessfulRuns   int64         `json:"successful_runs"`
	FailedRuns       int64         `json:"failed_runs"`
	NextScheduledRun *time.Time    `json:"next_scheduled_run,omitempty"`
}

// BackgroundJobService manages background jobs for the marketplace
type BackgroundJobService struct {
	jobs    map[string]BackgroundJob
	status  map[string]*JobStatus
	tickers map[string]*time.Ticker
	logger  *logrus.Logger

	// Service state
	running   bool
	stopChans map[string]chan struct{}
	mutex     sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewBackgroundJobService creates a new background job service
func NewBackgroundJobService(logger *logrus.Logger) BackgroundJobServiceInterface {
	if logger == nil {
		logger = logrus.New()
	}

	return &BackgroundJobService{
		jobs:      make(map[string]BackgroundJob),
		status:    make(map[string]*JobStatus),
		tickers:   make(map[string]*time.Ticker),
		stopChans: make(map[string]chan struct{}),
		logger:    logger,
		running:   false,
	}
}

// Start begins running all registered background jobs
func (s *BackgroundJobService) Start(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		return fmt.Errorf("background job service is already running")
	}

	s.logger.Info("Starting background job service")

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.running = true

	// Start all registered jobs
	for name, job := range s.jobs {
		if job.IsEnabled() {
			if err := s.startJob(name, job); err != nil {
				s.logger.WithError(err).WithField("job", name).Error("Failed to start job")
				continue
			}
		}
	}

	s.logger.WithField("job_count", len(s.jobs)).Info("Background job service started")
	return nil
}

// Stop stops all running background jobs
func (s *BackgroundJobService) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return fmt.Errorf("background job service is not running")
	}

	s.logger.Info("Stopping background job service")

	// Stop all jobs
	for name := range s.jobs {
		s.stopJob(name)
	}

	// Cancel the context
	if s.cancel != nil {
		s.cancel()
	}

	s.running = false
	s.logger.Info("Background job service stopped")
	return nil
}

// IsRunning returns whether the service is currently running
func (s *BackgroundJobService) IsRunning() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.running
}

// RegisterJob registers a new background job
func (s *BackgroundJobService) RegisterJob(name string, job BackgroundJob) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.jobs[name]; exists {
		return fmt.Errorf("job with name '%s' already registered", name)
	}

	s.jobs[name] = job
	s.status[name] = &JobStatus{
		Name:        name,
		Description: job.GetDescription(),
		Enabled:     job.IsEnabled(),
		Running:     false,
		TotalRuns:   0,
	}

	s.logger.WithFields(logrus.Fields{
		"job":         name,
		"description": job.GetDescription(),
		"interval":    job.GetInterval(),
		"enabled":     job.IsEnabled(),
	}).Info("Registered background job")

	// If service is running and job is enabled, start it
	if s.running && job.IsEnabled() {
		if err := s.startJob(name, job); err != nil {
			s.logger.WithError(err).WithField("job", name).Error("Failed to start newly registered job")
		}
	}

	return nil
}

// UnregisterJob unregisters a background job
func (s *BackgroundJobService) UnregisterJob(name string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.jobs[name]; !exists {
		return fmt.Errorf("job with name '%s' not found", name)
	}

	// Stop the job if it's running
	s.stopJob(name)

	// Remove from maps
	delete(s.jobs, name)
	delete(s.status, name)

	s.logger.WithField("job", name).Info("Unregistered background job")
	return nil
}

// GetRegisteredJobs returns a list of all registered job names
func (s *BackgroundJobService) GetRegisteredJobs() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	jobs := make([]string, 0, len(s.jobs))
	for name := range s.jobs {
		jobs = append(jobs, name)
	}
	return jobs
}

// ExecuteJob manually executes a specific job
func (s *BackgroundJobService) ExecuteJob(ctx context.Context, jobName string) error {
	s.mutex.RLock()
	job, exists := s.jobs[jobName]
	status := s.status[jobName]
	s.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("job '%s' not found", jobName)
	}

	s.logger.WithField("job", jobName).Info("Manually executing job")
	return s.executeJob(ctx, jobName, job, status)
}

// ExecuteAllJobs manually executes all registered jobs
func (s *BackgroundJobService) ExecuteAllJobs(ctx context.Context) error {
	s.mutex.RLock()
	jobs := make(map[string]BackgroundJob)
	statuses := make(map[string]*JobStatus)
	for name, job := range s.jobs {
		jobs[name] = job
		statuses[name] = s.status[name]
	}
	s.mutex.RUnlock()

	s.logger.Info("Manually executing all jobs")

	var lastError error
	for name, job := range jobs {
		if err := s.executeJob(ctx, name, job, statuses[name]); err != nil {
			s.logger.WithError(err).WithField("job", name).Error("Failed to execute job")
			lastError = err
		}
	}

	return lastError
}

// GetJobStatus returns the status of a specific job
func (s *BackgroundJobService) GetJobStatus(jobName string) *JobStatus {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	status, exists := s.status[jobName]
	if !exists {
		return nil
	}

	// Return a copy to avoid race conditions
	statusCopy := *status
	return &statusCopy
}

// GetAllJobStatuses returns the status of all jobs
func (s *BackgroundJobService) GetAllJobStatuses() map[string]*JobStatus {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	statuses := make(map[string]*JobStatus)
	for name, status := range s.status {
		// Return copies to avoid race conditions
		statusCopy := *status
		statuses[name] = &statusCopy
	}
	return statuses
}

// startJob starts a specific job
func (s *BackgroundJobService) startJob(name string, job BackgroundJob) error {
	if !job.IsEnabled() {
		return fmt.Errorf("job '%s' is disabled", name)
	}

	interval := job.GetInterval()
	if interval <= 0 {
		return fmt.Errorf("job '%s' has invalid interval: %v", name, interval)
	}

	// Create ticker and stop channel
	ticker := time.NewTicker(interval)
	stopChan := make(chan struct{})

	s.tickers[name] = ticker
	s.stopChans[name] = stopChan

	// Calculate next scheduled run
	nextRun := time.Now().Add(interval)
	s.status[name].NextScheduledRun = &nextRun

	// Start job goroutine
	go s.jobLoop(name, job, ticker, stopChan)

	s.logger.WithFields(logrus.Fields{
		"job":      name,
		"interval": interval,
		"next_run": nextRun,
	}).Info("Started background job")

	return nil
}

// stopJob stops a specific job
func (s *BackgroundJobService) stopJob(name string) {
	// Stop ticker
	if ticker, exists := s.tickers[name]; exists {
		ticker.Stop()
		delete(s.tickers, name)
	}

	// Signal stop
	if stopChan, exists := s.stopChans[name]; exists {
		close(stopChan)
		delete(s.stopChans, name)
	}

	// Update status
	if status, exists := s.status[name]; exists {
		status.Running = false
		status.NextScheduledRun = nil
	}

	s.logger.WithField("job", name).Info("Stopped background job")
}

// jobLoop is the main loop for a background job
func (s *BackgroundJobService) jobLoop(name string, job BackgroundJob, ticker *time.Ticker, stopChan chan struct{}) {
	status := s.status[name]

	for {
		select {
		case <-ticker.C:
			// Execute the job
			if err := s.executeJob(s.ctx, name, job, status); err != nil {
				s.logger.WithError(err).WithField("job", name).Error("Job execution failed")
			}

			// Update next scheduled run
			s.mutex.Lock()
			nextRun := time.Now().Add(job.GetInterval())
			status.NextScheduledRun = &nextRun
			s.mutex.Unlock()

		case <-stopChan:
			s.logger.WithField("job", name).Info("Job loop stopping")
			return

		case <-s.ctx.Done():
			s.logger.WithField("job", name).Info("Job loop cancelled by context")
			return
		}
	}
}

// executeJob executes a single job and updates its status
func (s *BackgroundJobService) executeJob(ctx context.Context, name string, job BackgroundJob, status *JobStatus) error {
	startTime := time.Now()

	s.mutex.Lock()
	status.Running = true
	status.TotalRuns++
	s.mutex.Unlock()

	s.logger.WithField("job", name).Debug("Executing job")

	// Execute the job
	err := job.Execute(ctx)

	// Update status
	s.mutex.Lock()
	status.Running = false
	status.LastRun = &startTime
	status.LastDuration = time.Since(startTime)

	if err != nil {
		status.FailedRuns++
		status.LastError = err.Error()
		s.logger.WithError(err).WithField("job", name).Error("Job execution failed")
	} else {
		status.SuccessfulRuns++
		status.LastError = ""
		s.logger.WithFields(logrus.Fields{
			"job":      name,
			"duration": status.LastDuration,
		}).Debug("Job executed successfully")
	}
	s.mutex.Unlock()

	return err
}

// AuctionExpiryJob implements the BackgroundJob interface for auction expiry processing
type AuctionExpiryJob struct {
	name             string
	description      string
	interval         time.Duration
	enabled          bool
	lifecycleService AuctionLifecycleServiceInterface
	logger           *logrus.Logger
}

// NewAuctionExpiryJob creates a new auction expiry background job
func NewAuctionExpiryJob(
	lifecycleService AuctionLifecycleServiceInterface,
	interval time.Duration,
	logger *logrus.Logger,
) BackgroundJob {
	if logger == nil {
		logger = logrus.New()
	}

	return &AuctionExpiryJob{
		name:             "auction_expiry",
		description:      "Processes expired auctions and determines winners",
		interval:         interval,
		enabled:          true,
		lifecycleService: lifecycleService,
		logger:           logger,
	}
}

// Execute processes expired auctions
func (j *AuctionExpiryJob) Execute(ctx context.Context) error {
	j.logger.Debug("Starting auction expiry processing")

	result, err := j.lifecycleService.ProcessExpiredAuctions(ctx)
	if err != nil {
		return fmt.Errorf("failed to process expired auctions: %w", err)
	}

	if result.ProcessedCount > 0 {
		j.logger.WithFields(logrus.Fields{
			"processed":  result.ProcessedCount,
			"successful": result.SuccessfulCount,
			"failed":     result.FailedCount,
			"duration":   result.ProcessingTime,
		}).Info("Processed expired auctions")

		// Log any errors
		if len(result.Errors) > 0 {
			for _, errMsg := range result.Errors {
				j.logger.WithField("error", errMsg).Warn("Auction processing error")
			}
		}
	}

	return nil
}

// GetName returns the job name
func (j *AuctionExpiryJob) GetName() string {
	return j.name
}

// GetDescription returns the job description
func (j *AuctionExpiryJob) GetDescription() string {
	return j.description
}

// GetInterval returns the job execution interval
func (j *AuctionExpiryJob) GetInterval() time.Duration {
	return j.interval
}

// IsEnabled returns whether the job is enabled
func (j *AuctionExpiryJob) IsEnabled() bool {
	return j.enabled
}

// SetEnabled enables or disables the job
func (j *AuctionExpiryJob) SetEnabled(enabled bool) {
	j.enabled = enabled
}

// SetInterval updates the job execution interval
func (j *AuctionExpiryJob) SetInterval(interval time.Duration) {
	j.interval = interval
}
