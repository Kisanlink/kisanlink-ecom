package resilience

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net"
	"syscall"
	"time"
)

// RetryConfig holds the configuration for retry policy
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int
	// InitialInterval is the initial wait time before the first retry
	InitialInterval time.Duration
	// MaxInterval is the maximum wait time between retries
	MaxInterval time.Duration
	// Multiplier is the factor by which the interval increases after each retry
	Multiplier float64
	// RandomizationFactor adds randomness to retry intervals to avoid thundering herd
	RandomizationFactor float64
	// RetryableErrorChecker determines if an error is retryable
	RetryableErrorChecker func(error) bool
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:            3,
		InitialInterval:       100 * time.Millisecond,
		MaxInterval:           10 * time.Second,
		Multiplier:            2.0,
		RandomizationFactor:   0.1,
		RetryableErrorChecker: DefaultRetryableErrorChecker,
	}
}

// RetryPolicy implements retry logic with exponential backoff and jitter
type RetryPolicy struct {
	config *RetryConfig
	rand   *rand.Rand
}

// NewRetryPolicy creates a new retry policy with the given configuration
func NewRetryPolicy(config *RetryConfig) *RetryPolicy {
	if config == nil {
		config = DefaultRetryConfig()
	}

	return &RetryPolicy{
		config: config,
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Execute executes the given function with retry logic
func (rp *RetryPolicy) Execute(fn func() error) error {
	return rp.ExecuteContext(context.Background(), func(context.Context) error {
		return fn()
	})
}

// ExecuteContext executes the given function with retry logic and context
func (rp *RetryPolicy) ExecuteContext(ctx context.Context, fn func(context.Context) error) error {
	var lastErr error

	for attempt := 0; attempt <= rp.config.MaxRetries; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute the function
		err := fn(ctx)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !rp.config.RetryableErrorChecker(err) {
			return err
		}

		// Don't sleep after the last attempt
		if attempt == rp.config.MaxRetries {
			break
		}

		// Calculate backoff delay
		delay := rp.calculateBackoff(attempt)

		// Sleep with context cancellation support
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", rp.config.MaxRetries, lastErr)
}

// ExecuteWithResult executes a function that returns a result with retry logic
func (rp *RetryPolicy) ExecuteWithResult(fn func() (interface{}, error)) (interface{}, error) {
	return rp.ExecuteWithResultContext(context.Background(), func(context.Context) (interface{}, error) {
		return fn()
	})
}

// ExecuteWithResultContext executes a function that returns a result with retry logic and context
func (rp *RetryPolicy) ExecuteWithResultContext(ctx context.Context, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	var lastErr error
	var result interface{}

	for attempt := 0; attempt <= rp.config.MaxRetries; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Execute the function
		var err error
		result, err = fn(ctx)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if error is retryable
		if !rp.config.RetryableErrorChecker(err) {
			return nil, err
		}

		// Don't sleep after the last attempt
		if attempt == rp.config.MaxRetries {
			break
		}

		// Calculate backoff delay
		delay := rp.calculateBackoff(attempt)

		// Sleep with context cancellation support
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return nil, fmt.Errorf("operation failed after %d retries: %w", rp.config.MaxRetries, lastErr)
}

// calculateBackoff calculates the backoff delay for the given attempt
func (rp *RetryPolicy) calculateBackoff(attempt int) time.Duration {
	// Calculate exponential backoff
	backoff := float64(rp.config.InitialInterval) * math.Pow(rp.config.Multiplier, float64(attempt))

	// Apply jitter to avoid thundering herd
	if rp.config.RandomizationFactor > 0 {
		jitter := backoff * rp.config.RandomizationFactor
		backoff = backoff - jitter + (rp.rand.Float64() * 2 * jitter)
	}

	// Ensure backoff doesn't exceed max interval
	if backoff > float64(rp.config.MaxInterval) {
		backoff = float64(rp.config.MaxInterval)
	}

	return time.Duration(backoff)
}

// DefaultRetryableErrorChecker determines if an error is retryable by default
func DefaultRetryableErrorChecker(err error) bool {
	if err == nil {
		return false
	}

	// Check for network errors that are typically retryable
	if netErr, ok := err.(net.Error); ok {
		// Timeout errors are retryable
		if netErr.Timeout() {
			return true
		}
		// Temporary errors are retryable
		if netErr.Temporary() {
			return true
		}
	}

	// Check for specific syscall errors that are retryable
	if opErr, ok := err.(*net.OpError); ok {
		if syscallErr, ok := opErr.Err.(*syscall.Errno); ok {
			switch *syscallErr {
			case syscall.ECONNREFUSED, syscall.ECONNRESET, syscall.ETIMEDOUT:
				return true
			}
		}
	}

	// Check for context errors - these are not retryable
	if err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}

	// Check for circuit breaker errors - these are not retryable
	if IsCircuitBreakerError(err) {
		return false
	}

	// By default, don't retry unknown errors
	return false
}

// TimeoutConfig holds configuration for timeout handling
type TimeoutConfig struct {
	// Timeout is the maximum duration for an operation
	Timeout time.Duration
}

// DefaultTimeoutConfig returns a default timeout configuration
func DefaultTimeoutConfig() *TimeoutConfig {
	return &TimeoutConfig{
		Timeout: 30 * time.Second,
	}
}

// TimeoutWrapper wraps operations with timeout handling
type TimeoutWrapper struct {
	config *TimeoutConfig
}

// NewTimeoutWrapper creates a new timeout wrapper
func NewTimeoutWrapper(config *TimeoutConfig) *TimeoutWrapper {
	if config == nil {
		config = DefaultTimeoutConfig()
	}

	return &TimeoutWrapper{
		config: config,
	}
}

// Execute executes a function with timeout
func (tw *TimeoutWrapper) Execute(fn func() error) error {
	return tw.ExecuteContext(context.Background(), func(context.Context) error {
		return fn()
	})
}

// ExecuteContext executes a function with timeout and context
func (tw *TimeoutWrapper) ExecuteContext(ctx context.Context, fn func(context.Context) error) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, tw.config.Timeout)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		done <- fn(timeoutCtx)
	}()

	select {
	case err := <-done:
		return err
	case <-timeoutCtx.Done():
		return timeoutCtx.Err()
	}
}

// ExecuteWithResult executes a function that returns a result with timeout
func (tw *TimeoutWrapper) ExecuteWithResult(fn func() (interface{}, error)) (interface{}, error) {
	return tw.ExecuteWithResultContext(context.Background(), func(context.Context) (interface{}, error) {
		return fn()
	})
}

// ExecuteWithResultContext executes a function that returns a result with timeout and context
func (tw *TimeoutWrapper) ExecuteWithResultContext(ctx context.Context, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, tw.config.Timeout)
	defer cancel()

	type result struct {
		value interface{}
		err   error
	}

	done := make(chan result, 1)

	go func() {
		value, err := fn(timeoutCtx)
		done <- result{value: value, err: err}
	}()

	select {
	case res := <-done:
		return res.value, res.err
	case <-timeoutCtx.Done():
		return nil, timeoutCtx.Err()
	}
}

// RetryStats holds statistics about retry operations
type RetryStats struct {
	TotalAttempts   int           `json:"total_attempts"`
	SuccessAttempts int           `json:"success_attempts"`
	FailedAttempts  int           `json:"failed_attempts"`
	AvgAttempts     float64       `json:"avg_attempts"`
	MaxDelay        time.Duration `json:"max_delay"`
	TotalDelay      time.Duration `json:"total_delay"`
}

// StatsCollectingRetryPolicy wraps RetryPolicy to collect statistics
type StatsCollectingRetryPolicy struct {
	policy *RetryPolicy
	stats  RetryStats
}

// NewStatsCollectingRetryPolicy creates a retry policy that collects statistics
func NewStatsCollectingRetryPolicy(config *RetryConfig) *StatsCollectingRetryPolicy {
	return &StatsCollectingRetryPolicy{
		policy: NewRetryPolicy(config),
	}
}

// ExecuteContext executes a function with retry logic and collects statistics
func (srp *StatsCollectingRetryPolicy) ExecuteContext(ctx context.Context, fn func(context.Context) error) error {
	startTime := time.Now()
	attempts := 0

	err := srp.policy.ExecuteContext(ctx, func(ctx context.Context) error {
		attempts++
		return fn(ctx)
	})

	// Update statistics
	srp.stats.TotalAttempts += attempts
	if err == nil {
		srp.stats.SuccessAttempts++
	} else {
		srp.stats.FailedAttempts++
	}

	totalOps := srp.stats.SuccessAttempts + srp.stats.FailedAttempts
	if totalOps > 0 {
		srp.stats.AvgAttempts = float64(srp.stats.TotalAttempts) / float64(totalOps)
	}

	delay := time.Since(startTime)
	srp.stats.TotalDelay += delay
	if delay > srp.stats.MaxDelay {
		srp.stats.MaxDelay = delay
	}

	return err
}

// GetStats returns the collected retry statistics
func (srp *StatsCollectingRetryPolicy) GetStats() RetryStats {
	return srp.stats
}

// ResetStats resets the collected statistics
func (srp *StatsCollectingRetryPolicy) ResetStats() {
	srp.stats = RetryStats{}
}

// Convenience functions for common retry patterns

// RetryWithExponentialBackoff retries a function with exponential backoff
func RetryWithExponentialBackoff(ctx context.Context, maxRetries int, initialDelay time.Duration, fn func() error) error {
	config := &RetryConfig{
		MaxRetries:            maxRetries,
		InitialInterval:       initialDelay,
		MaxInterval:           30 * time.Second,
		Multiplier:            2.0,
		RandomizationFactor:   0.1,
		RetryableErrorChecker: DefaultRetryableErrorChecker,
	}

	policy := NewRetryPolicy(config)
	return policy.ExecuteContext(ctx, func(context.Context) error {
		return fn()
	})
}

// RetryWithLinearBackoff retries a function with linear backoff
func RetryWithLinearBackoff(ctx context.Context, maxRetries int, delay time.Duration, fn func() error) error {
	config := &RetryConfig{
		MaxRetries:            maxRetries,
		InitialInterval:       delay,
		MaxInterval:           delay * 10,
		Multiplier:            1.0, // Linear backoff
		RandomizationFactor:   0.1,
		RetryableErrorChecker: DefaultRetryableErrorChecker,
	}

	policy := NewRetryPolicy(config)
	return policy.ExecuteContext(ctx, func(context.Context) error {
		return fn()
	})
}

// RetryWithFixedDelay retries a function with fixed delay between attempts
func RetryWithFixedDelay(ctx context.Context, maxRetries int, delay time.Duration, fn func() error) error {
	config := &RetryConfig{
		MaxRetries:            maxRetries,
		InitialInterval:       delay,
		MaxInterval:           delay,
		Multiplier:            1.0,
		RandomizationFactor:   0.0, // No jitter for fixed delay
		RetryableErrorChecker: DefaultRetryableErrorChecker,
	}

	policy := NewRetryPolicy(config)
	return policy.ExecuteContext(ctx, func(context.Context) error {
		return fn()
	})
}

// RetryableError wraps an error to indicate it should be retried
type RetryableError struct {
	Err error
}

func (re RetryableError) Error() string {
	return re.Err.Error()
}

func (re RetryableError) Unwrap() error {
	return re.Err
}

// NewRetryableError creates a new retryable error
func NewRetryableError(err error) RetryableError {
	return RetryableError{Err: err}
}

// IsRetryableError checks if an error is marked as retryable
func IsRetryableError(err error) bool {
	var retryableErr RetryableError
	return errors.As(err, &retryableErr)
}

// NonRetryableError wraps an error to indicate it should not be retried
type NonRetryableError struct {
	Err error
}

func (nre NonRetryableError) Error() string {
	return nre.Err.Error()
}

func (nre NonRetryableError) Unwrap() error {
	return nre.Err
}

// NewNonRetryableError creates a new non-retryable error
func NewNonRetryableError(err error) NonRetryableError {
	return NonRetryableError{Err: err}
}

// IsNonRetryableError checks if an error is marked as non-retryable
func IsNonRetryableError(err error) bool {
	var nonRetryableErr NonRetryableError
	return errors.As(err, &nonRetryableErr)
}

// Enhanced error checker that considers explicit retry markers
func EnhancedRetryableErrorChecker(err error) bool {
	// Check for explicit non-retryable marker
	if IsNonRetryableError(err) {
		return false
	}

	// Check for explicit retryable marker
	if IsRetryableError(err) {
		return true
	}

	// Fall back to default logic
	return DefaultRetryableErrorChecker(err)
}
