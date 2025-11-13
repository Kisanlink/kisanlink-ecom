package aaa

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// CircuitState represents the state of the circuit breaker
type CircuitState int

const (
	// StateClosed allows all requests through
	StateClosed CircuitState = iota
	// StateOpen rejects all requests
	StateOpen
	// StateHalfOpen allows a limited number of requests to test if service recovered
	StateHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name   string
	config CircuitBreakerConfig
	logger *logrus.Logger

	mu                   sync.RWMutex
	state                CircuitState
	generation           uint64
	counts               *counts
	expiry               time.Time
	consecutiveSuccesses uint32
}

// counts tracks success and failure counts
type counts struct {
	requests         uint32
	totalSuccesses   uint32
	totalFailures    uint32
	consecutiveFails uint32
}

func newCounts() *counts {
	return &counts{}
}

func (c *counts) onSuccess() {
	c.requests++
	c.totalSuccesses++
	c.consecutiveFails = 0
}

func (c *counts) onFailure() {
	c.requests++
	c.totalFailures++
	c.consecutiveFails++
}

func (c *counts) clear() {
	c.requests = 0
	c.totalSuccesses = 0
	c.totalFailures = 0
	c.consecutiveFails = 0
}

func (c *counts) failureRatio() float64 {
	if c.requests == 0 {
		return 0.0
	}
	return float64(c.totalFailures) / float64(c.requests)
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, config CircuitBreakerConfig, logger *logrus.Logger) *CircuitBreaker {
	return &CircuitBreaker{
		name:       name,
		config:     config,
		logger:     logger,
		state:      StateClosed,
		counts:     newCounts(),
		generation: 0,
	}
}

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(context.Context) error) error {
	generation, err := cb.beforeRequest()
	if err != nil {
		return err
	}

	// Execute the function
	err = fn(ctx)

	// Record result
	cb.afterRequest(generation, err)

	return err
}

// beforeRequest checks if the request can proceed
func (cb *CircuitBreaker) beforeRequest() (uint64, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)

	if state == StateOpen {
		return generation, ErrCircuitOpen
	} else if state == StateHalfOpen && cb.counts.requests >= cb.config.MaxRequests {
		return generation, ErrTooManyRequests
	}

	// Don't increment count here - will be done in onSuccess/onFailure
	return generation, nil
}

// afterRequest records the result of a request
func (cb *CircuitBreaker) afterRequest(generation uint64, err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state, currGen := cb.currentState(now)

	// Ignore if generation doesn't match (state changed during request)
	if generation != currGen {
		return
	}

	if err == nil {
		cb.onSuccess(state)
	} else {
		cb.onFailure(state)
	}
}

// onSuccess handles successful requests
func (cb *CircuitBreaker) onSuccess(state CircuitState) {
	cb.counts.onSuccess()

	if state == StateHalfOpen {
		cb.consecutiveSuccesses++

		cb.logger.WithFields(logrus.Fields{
			"circuit":             cb.name,
			"consecutive_success": cb.consecutiveSuccesses,
			"max_requests":        cb.config.MaxRequests,
		}).Debug("Circuit breaker success in half-open state")

		// Transition to CLOSED after required consecutive successes
		if cb.consecutiveSuccesses >= cb.config.MaxRequests {
			cb.setState(StateClosed, time.Now())
			cb.logger.WithField("circuit", cb.name).Info("Circuit breaker CLOSED after recovery")
		}
	}
}

// onFailure handles failed requests
func (cb *CircuitBreaker) onFailure(state CircuitState) {
	cb.counts.onFailure()

	switch state {
	case StateClosed:
		// Check if we should open the circuit
		if cb.shouldTrip() {
			cb.setState(StateOpen, time.Now())
			cb.logger.WithFields(logrus.Fields{
				"circuit":       cb.name,
				"requests":      cb.counts.requests,
				"failures":      cb.counts.totalFailures,
				"failure_ratio": cb.counts.failureRatio(),
			}).Warn("Circuit breaker OPEN due to failures")
		}

	case StateHalfOpen:
		// Immediately reopen on any failure in half-open state
		cb.setState(StateOpen, time.Now())
		cb.logger.WithField("circuit", cb.name).Warn("Circuit breaker OPEN again after failure in HALF_OPEN")
	}
}

// shouldTrip determines if the circuit should trip to open state
func (cb *CircuitBreaker) shouldTrip() bool {
	// Need minimum number of requests before considering failure ratio
	if cb.counts.requests < cb.config.MinRequests {
		return false
	}

	// Check failure ratio
	failureRatio := cb.counts.failureRatio()
	return failureRatio >= cb.config.FailureRatio
}

// currentState returns the current state and generation
func (cb *CircuitBreaker) currentState(now time.Time) (CircuitState, uint64) {
	switch cb.state {
	case StateClosed:
		// Check if interval expired - reset counts
		if cb.expiry.IsZero() {
			// Initialize expiry on first request
			cb.expiry = now.Add(cb.config.Interval)
		} else if cb.expiry.Before(now) {
			cb.counts.clear()
			cb.expiry = now.Add(cb.config.Interval)
		}

	case StateOpen:
		// Check if timeout expired - transition to half-open
		if cb.expiry.Before(now) {
			cb.setState(StateHalfOpen, now)
			cb.logger.WithField("circuit", cb.name).Info("Circuit breaker transitioned to HALF_OPEN")
		}
	}

	return cb.state, cb.generation
}

// setState changes the state and updates related fields
func (cb *CircuitBreaker) setState(state CircuitState, now time.Time) {
	if cb.state == state {
		return
	}

	prevState := cb.state
	cb.state = state
	cb.generation++

	switch state {
	case StateClosed:
		cb.counts.clear()
		cb.expiry = now.Add(cb.config.Interval)
		cb.consecutiveSuccesses = 0

	case StateOpen:
		cb.expiry = now.Add(cb.config.Timeout)

	case StateHalfOpen:
		cb.counts.clear()
		cb.consecutiveSuccesses = 0
	}

	cb.logger.WithFields(logrus.Fields{
		"circuit": cb.name,
		"from":    prevState.String(),
		"to":      state.String(),
	}).Info("Circuit breaker state changed")
}

// State returns the current state
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	state, _ := cb.currentState(time.Now())
	return state
}

// Counts returns a copy of the current counts
func (cb *CircuitBreaker) Counts() (requests, successes, failures uint32, failureRatio float64) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.counts.requests, cb.counts.totalSuccesses, cb.counts.totalFailures, cb.counts.failureRatio()
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.setState(StateClosed, time.Now())
	cb.logger.WithField("circuit", cb.name).Info("Circuit breaker manually reset")
}

// String returns a string representation of the circuit breaker
func (cb *CircuitBreaker) String() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return fmt.Sprintf("CircuitBreaker{name=%s, state=%s, requests=%d, failures=%d, ratio=%.2f}",
		cb.name, cb.state, cb.counts.requests, cb.counts.totalFailures, cb.counts.failureRatio())
}
