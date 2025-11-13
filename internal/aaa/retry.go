package aaa

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxAttempts    int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	BackoffFactor  float64
	JitterFactor   float64 // 0.0 to 1.0, adds randomness to prevent thundering herd
}

// RetryExecutor handles retry logic with exponential backoff and jitter
type RetryExecutor struct {
	config RetryConfig
	logger *logrus.Logger
}

// NewRetryExecutor creates a new retry executor
func NewRetryExecutor(config RetryConfig, logger *logrus.Logger) *RetryExecutor {
	if config.JitterFactor == 0 {
		config.JitterFactor = 0.1 // Default 10% jitter
	}
	return &RetryExecutor{
		config: config,
		logger: logger,
	}
}

// Execute runs the given function with retry logic
func (r *RetryExecutor) Execute(ctx context.Context, operation string, fn func(context.Context) error) error {
	var lastErr error
	backoff := r.config.InitialBackoff

	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		// Execute the function
		err := fn(ctx)

		if err == nil {
			// Success
			if attempt > 1 {
				r.logger.WithFields(logrus.Fields{
					"operation": operation,
					"attempt":   attempt,
				}).Info("Operation succeeded after retry")
			}
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !IsRetryable(err) {
			r.logger.WithFields(logrus.Fields{
				"operation": operation,
				"attempt":   attempt,
				"error":     err,
			}).Debug("Non-retryable error, failing fast")
			return err
		}

		// Check if we've exhausted attempts
		if attempt >= r.config.MaxAttempts {
			r.logger.WithFields(logrus.Fields{
				"operation": operation,
				"attempts":  attempt,
				"error":     err,
			}).Warn("Max retry attempts exhausted")
			return ErrRetryExhausted{
				Attempts: attempt,
				LastErr:  err,
			}
		}

		// Calculate backoff with jitter
		sleep := r.calculateBackoff(backoff, attempt)

		// Check context before sleeping
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleep):
			// Continue to next attempt
		}

		r.logger.WithFields(logrus.Fields{
			"operation": operation,
			"attempt":   attempt,
			"error":     err,
			"backoff":   sleep,
		}).Debug("Retrying operation after backoff")

		// Increase backoff for next attempt
		backoff = r.nextBackoff(backoff)
	}

	return ErrRetryExhausted{
		Attempts: r.config.MaxAttempts,
		LastErr:  lastErr,
	}
}

// calculateBackoff calculates the backoff duration with jitter
func (r *RetryExecutor) calculateBackoff(baseBackoff time.Duration, _ int) time.Duration {
	// Add jitter to prevent thundering herd
	//nolint:gosec // Using math/rand for jitter is acceptable here
	jitter := 1.0 + (rand.Float64()*2-1)*r.config.JitterFactor // ±jitterFactor

	backoff := float64(baseBackoff) * jitter
	return time.Duration(backoff)
}

// nextBackoff calculates the next backoff duration using exponential backoff
func (r *RetryExecutor) nextBackoff(current time.Duration) time.Duration {
	next := time.Duration(float64(current) * r.config.BackoffFactor)

	// Cap at max backoff
	if next > r.config.MaxBackoff {
		return r.config.MaxBackoff
	}

	return next
}

// CalculateBackoffSequence returns the sequence of backoff durations for testing
func (r *RetryExecutor) CalculateBackoffSequence() []time.Duration {
	sequence := make([]time.Duration, 0, r.config.MaxAttempts-1)
	backoff := r.config.InitialBackoff

	for i := 1; i < r.config.MaxAttempts; i++ {
		sequence = append(sequence, backoff)
		backoff = r.nextBackoff(backoff)
	}

	return sequence
}

// WithExponentialBackoff is a helper to create a standard exponential backoff config
func WithExponentialBackoff(maxAttempts int, initialBackoff, maxBackoff time.Duration) RetryConfig {
	return RetryConfig{
		MaxAttempts:    maxAttempts,
		InitialBackoff: initialBackoff,
		MaxBackoff:     maxBackoff,
		BackoffFactor:  2.0,
		JitterFactor:   0.1,
	}
}

// WithLinearBackoff is a helper to create a linear backoff config
func WithLinearBackoff(maxAttempts int, backoff time.Duration) RetryConfig {
	return RetryConfig{
		MaxAttempts:    maxAttempts,
		InitialBackoff: backoff,
		MaxBackoff:     backoff * time.Duration(maxAttempts),
		BackoffFactor:  1.0,
		JitterFactor:   0.1,
	}
}

// WithFixedBackoff is a helper to create a fixed backoff config
func WithFixedBackoff(maxAttempts int, backoff time.Duration) RetryConfig {
	return RetryConfig{
		MaxAttempts:    maxAttempts,
		InitialBackoff: backoff,
		MaxBackoff:     backoff,
		BackoffFactor:  1.0,
		JitterFactor:   0.0, // No jitter for fixed backoff
	}
}

// calculateExponentialBackoff is a utility function for calculating exponential backoff
//
//nolint:unused // Utility function kept for future use
func calculateExponentialBackoff(attempt int, baseDelay, maxDelay time.Duration, factor float64) time.Duration {
	backoff := float64(baseDelay) * math.Pow(factor, float64(attempt-1))
	if backoff > float64(maxDelay) {
		backoff = float64(maxDelay)
	}
	return time.Duration(backoff)
}
