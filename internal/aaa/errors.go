package aaa

import (
	"errors"
	"fmt"
)

// Common errors
var (
	// Service errors
	ErrServiceUnavailable = errors.New("AAA service unavailable")
	ErrTooManyRequests    = errors.New("too many requests")
	ErrRateLimited        = errors.New("rate limited")
	ErrTimeout            = errors.New("request timeout")
	ErrCircuitOpen        = errors.New("circuit breaker is open")

	// Address errors
	ErrAddressNotFound   = errors.New("address not found")
	ErrInvalidAddress    = errors.New("invalid address")
	ErrAddressExists     = errors.New("address already exists")
	ErrInvalidAddressID  = errors.New("invalid address ID")
	ErrAddressValidation = errors.New("address validation failed")

	// Connection errors
	ErrNoConnections    = errors.New("no available connections")
	ErrConnectionClosed = errors.New("connection closed")
	ErrDialFailed       = errors.New("failed to dial AAA service")
	ErrPoolShutdown     = errors.New("connection pool shut down")

	// Auth errors
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrUnauthorized    = errors.New("unauthorized")

	// Request errors
	ErrInvalidRequest = errors.New("invalid request")
	ErrMissingField   = errors.New("required field missing")
	ErrInvalidInput   = errors.New("invalid input")

	// Internal errors
	ErrInternal = errors.New("internal error")
)

// ErrInvalidConfig represents a configuration validation error
type ErrInvalidConfig struct {
	Field   string
	Message string
}

func (e ErrInvalidConfig) Error() string {
	return fmt.Sprintf("invalid config: %s - %s", e.Field, e.Message)
}

// ErrRetryExhausted indicates all retry attempts have been exhausted
type ErrRetryExhausted struct {
	Attempts int
	LastErr  error
}

func (e ErrRetryExhausted) Error() string {
	return fmt.Sprintf("retry exhausted after %d attempts: %v", e.Attempts, e.LastErr)
}

func (e ErrRetryExhausted) Unwrap() error {
	return e.LastErr
}

// IsRetryable determines if an error should trigger a retry
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Retryable errors
	switch {
	case errors.Is(err, ErrServiceUnavailable):
		return true
	case errors.Is(err, ErrTimeout):
		return true
	case errors.Is(err, ErrConnectionClosed):
		return true
	case errors.Is(err, ErrNoConnections):
		return true
	case errors.Is(err, ErrTooManyRequests):
		return false // Don't retry rate limits
	case errors.Is(err, ErrInvalidRequest):
		return false // Don't retry invalid requests
	case errors.Is(err, ErrInvalidAddress):
		return false
	case errors.Is(err, ErrAddressNotFound):
		return false
	case errors.Is(err, ErrUnauthenticated):
		return false
	case errors.Is(err, ErrUnauthorized):
		return false
	default:
		return false
	}
}
