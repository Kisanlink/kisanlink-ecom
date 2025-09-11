package common

import (
    "context"
    "fmt"
    "math"
    "math/rand/v2"
    "time"
)

// RetryConfig defines retry behavior
type RetryConfig struct {
    MaxAttempts     int           `json:"max_attempts"`
    InitialDelay    time.Duration `json:"initial_delay"`
    MaxDelay        time.Duration `json:"max_delay"`
    BackoffFactor   float64       `json:"backoff_factor"`
    Jitter          bool          `json:"jitter"`
    RetryableErrors []error       `json:"-"`
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() *RetryConfig {
    return &RetryConfig{
        MaxAttempts:   3,
        InitialDelay:  100 * time.Millisecond,
        MaxDelay:      5 * time.Second,
        BackoffFactor: 2.0,
        Jitter:        true,
        RetryableErrors: []error{
            ErrDatabaseConnection,
            ErrDatabaseQuery,
            ErrExternalServiceUnavailable,
            ErrExternalServiceTimeout,
        },
    }
}

// RetryableOperation represents an operation that can be retried
type RetryableOperation func(ctx context.Context, attempt int) error

// Retry executes an operation with retry logic
func Retry(ctx context.Context, config *RetryConfig, operation RetryableOperation) error {
    if config == nil {
        config = DefaultRetryConfig()
    }

    var lastErr error

    for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
        // Execute the operation
        err := operation(ctx, attempt)
        if err == nil {
            return nil // Success
        }

        lastErr = err

        // Check if error is retryable
        if !isRetryableError(err, config.RetryableErrors) {
            return err // Non-retryable error
        }

        // Don't wait after the last attempt
        if attempt == config.MaxAttempts {
            break
        }

        // Check if context is cancelled
        if ctx.Err() != nil {
            return ctx.Err()
        }

        // Calculate delay for next attempt
        delay := calculateDelay(attempt, config)

        // Wait before next attempt
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(delay):
            // Continue to next attempt
        }
    }

    return fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// RetryWithResult executes an operation with retry logic and returns a result
func RetryWithResult[T any](ctx context.Context, config *RetryConfig, operation func(ctx context.Context, attempt int) (T, error)) (T, error) {
    var zero T
    var lastErr error

    for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
        // Execute the operation
        result, err := operation(ctx, attempt)
        if err == nil {
            return result, nil // Success
        }

        lastErr = err

        // Check if error is retryable
        if !isRetryableError(err, config.RetryableErrors) {
            return zero, err // Non-retryable error
        }

        // Don't wait after the last attempt
        if attempt == config.MaxAttempts {
            break
        }

        // Check if context is cancelled
        if ctx.Err() != nil {
            return zero, ctx.Err()
        }

        // Calculate delay for next attempt
        delay := calculateDelay(attempt, config)

        // Wait before next attempt
        select {
        case <-ctx.Done():
            return zero, ctx.Err()
        case <-time.After(delay):
            // Continue to next attempt
        }
    }

    return zero, fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// isRetryableError checks if an error is retryable
func isRetryableError(err error, retryableErrors []error) bool {
    if len(retryableErrors) == 0 {
        // Default retryable errors
        return IsError(err, ErrDatabaseConnection) ||
            IsError(err, ErrDatabaseQuery) ||
            IsError(err, ErrExternalServiceUnavailable) ||
            IsError(err, ErrExternalServiceTimeout)
    }

    for _, retryableErr := range retryableErrors {
        if IsError(err, retryableErr) {
            return true
        }
    }

    // Check for AppError with retryable status codes
    if appErr, ok := IsAppError(err); ok {
        switch appErr.StatusCode {
        case 500, 502, 503, 504, 408, 429: // Server errors and timeouts
            return true
        }
    }

    return false
}

// calculateDelay calculates the delay for the next retry attempt
func calculateDelay(attempt int, config *RetryConfig) time.Duration {
    // Calculate exponential backoff
    delay := float64(config.InitialDelay) * math.Pow(config.BackoffFactor, float64(attempt-1))

    // Apply maximum delay limit
    if delay > float64(config.MaxDelay) {
        delay = float64(config.MaxDelay)
    }

    // Add jitter if enabled
    if config.Jitter {
        // Add random jitter up to 10% of the delay
        jitter := delay * 0.1 * (2*rand.Float64() - 1) // Random value between -10% and +10%
        delay += jitter
    }

    // Ensure delay is not negative
    if delay < 0 {
        delay = float64(config.InitialDelay)
    }

    return time.Duration(delay)
}

// CircuitBreakerConfig defines circuit breaker behavior
type CircuitBreakerConfig struct {
    FailureThreshold int           `json:"failure_threshold"`
    RecoveryTimeout  time.Duration `json:"recovery_timeout"`
    MaxRequests      int           `json:"max_requests"`
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
    CircuitBreakerClosed CircuitBreakerState = iota
    CircuitBreakerOpen
    CircuitBreakerHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
    config       *CircuitBreakerConfig
    state        CircuitBreakerState
    failures     int
    lastFailTime time.Time
    requests     int
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
    if config == nil {
        config = &CircuitBreakerConfig{
            FailureThreshold: 5,
            RecoveryTimeout:  30 * time.Second,
            MaxRequests:      3,
        }
    }

    return &CircuitBreaker{
        config: config,
        state:  CircuitBreakerClosed,
    }
}

// Execute executes an operation through the circuit breaker
func (cb *CircuitBreaker) Execute(ctx context.Context, operation RetryableOperation) error {
    if !cb.canExecute() {
        return NewExternalServiceError("circuit breaker", fmt.Errorf("circuit breaker is open"))
    }

    err := operation(ctx, 1)
    cb.recordResult(err)
    return err
}

// canExecute checks if the operation can be executed
func (cb *CircuitBreaker) canExecute() bool {
    switch cb.state {
    case CircuitBreakerClosed:
        return true
    case CircuitBreakerOpen:
        if time.Since(cb.lastFailTime) > cb.config.RecoveryTimeout {
            cb.state = CircuitBreakerHalfOpen
            cb.requests = 0
            return true
        }
        return false
    case CircuitBreakerHalfOpen:
        return cb.requests < cb.config.MaxRequests
    default:
        return false
    }
}

// recordResult records the result of an operation
func (cb *CircuitBreaker) recordResult(err error) {
    switch cb.state {
    case CircuitBreakerClosed:
        if err != nil {
            cb.failures++
            if cb.failures >= cb.config.FailureThreshold {
                cb.state = CircuitBreakerOpen
                cb.lastFailTime = time.Now()
            }
        } else {
            cb.failures = 0
        }
    case CircuitBreakerHalfOpen:
        cb.requests++
        if err != nil {
            cb.state = CircuitBreakerOpen
            cb.lastFailTime = time.Now()
            cb.failures = cb.config.FailureThreshold
        } else if cb.requests >= cb.config.MaxRequests {
            cb.state = CircuitBreakerClosed
            cb.failures = 0
        }
    }
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
    return cb.state
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
    cb.state = CircuitBreakerClosed
    cb.failures = 0
    cb.requests = 0
}
