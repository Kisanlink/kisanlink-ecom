package resilience

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	// StateClosed means the circuit breaker is closed and requests pass through
	StateClosed CircuitBreakerState = iota
	// StateOpen means the circuit breaker is open and requests are rejected
	StateOpen
	// StateHalfOpen means the circuit breaker allows limited requests to test recovery
	StateHalfOpen
)

// String returns the string representation of the circuit breaker state
func (s CircuitBreakerState) String() string {
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

// CircuitBreakerConfig holds the configuration for a circuit breaker
type CircuitBreakerConfig struct {
	// Name identifies the circuit breaker
	Name string
	// MaxRequests is the maximum number of requests allowed in half-open state
	MaxRequests uint32
	// Interval is the cyclic period in closed state for clearing counts
	Interval time.Duration
	// Timeout is the period after which the circuit breaker transitions from open to half-open
	Timeout time.Duration
	// ReadyToTrip returns true if the circuit breaker should trip (open)
	ReadyToTrip func(counts Counts) bool
	// OnStateChange is called when the state changes
	OnStateChange func(name string, from, to CircuitBreakerState)
}

// DefaultCircuitBreakerConfig returns a default circuit breaker configuration
func DefaultCircuitBreakerConfig(name string) *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		Name:        name,
		MaxRequests: 1,
		Interval:    60 * time.Second,
		Timeout:     60 * time.Second,
		ReadyToTrip: func(counts Counts) bool {
			return counts.Requests >= 3 && counts.TotalFailures >= 2
		},
		OnStateChange: func(name string, from, to CircuitBreakerState) {
			// Default no-op state change handler
		},
	}
}

// Counts holds the statistics of a circuit breaker
type Counts struct {
	Requests             uint32
	TotalSuccesses       uint32
	TotalFailures        uint32
	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
}

// Reset resets all counts to zero
func (c *Counts) Reset() {
	c.Requests = 0
	c.TotalSuccesses = 0
	c.TotalFailures = 0
	c.ConsecutiveSuccesses = 0
	c.ConsecutiveFailures = 0
}

// OnRequest increments the request count
func (c *Counts) OnRequest() {
	c.Requests++
}

// OnSuccess increments the success count and resets consecutive failures
func (c *Counts) OnSuccess() {
	c.TotalSuccesses++
	c.ConsecutiveSuccesses++
	c.ConsecutiveFailures = 0
}

// OnFailure increments the failure count and resets consecutive successes
func (c *Counts) OnFailure() {
	c.TotalFailures++
	c.ConsecutiveFailures++
	c.ConsecutiveSuccesses = 0
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name          string
	maxRequests   uint32
	interval      time.Duration
	timeout       time.Duration
	readyToTrip   func(counts Counts) bool
	onStateChange func(name string, from, to CircuitBreakerState)

	mutex      sync.Mutex
	state      CircuitBreakerState
	generation uint64
	counts     Counts
	expiry     time.Time
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig("default")
	}

	cb := &CircuitBreaker{
		name:          config.Name,
		maxRequests:   config.MaxRequests,
		interval:      config.Interval,
		timeout:       config.Timeout,
		readyToTrip:   config.ReadyToTrip,
		onStateChange: config.OnStateChange,
		state:         StateClosed,
		expiry:        time.Now().Add(config.Interval),
	}

	return cb
}

// Name returns the name of the circuit breaker
func (cb *CircuitBreaker) Name() string {
	return cb.name
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, _ := cb.currentState(now)
	return state
}

// Counts returns the current counts of the circuit breaker
func (cb *CircuitBreaker) Counts() Counts {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	return cb.counts
}

// Execute runs the given function if the circuit breaker allows it
func (cb *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	generation, err := cb.beforeRequest()
	if err != nil {
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			cb.afterRequest(generation, false)
			panic(r)
		}
	}()

	result, err := fn()
	cb.afterRequest(generation, err == nil)
	return result, err
}

// ExecuteContext runs the given function with context if the circuit breaker allows it
func (cb *CircuitBreaker) ExecuteContext(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	generation, err := cb.beforeRequest()
	if err != nil {
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			cb.afterRequest(generation, false)
			panic(r)
		}
	}()

	result, err := fn(ctx)
	cb.afterRequest(generation, err == nil)
	return result, err
}

// Call executes the given function and returns its result, wrapping it with circuit breaker logic
func (cb *CircuitBreaker) Call(fn func() error) error {
	_, err := cb.Execute(func() (interface{}, error) {
		return nil, fn()
	})
	return err
}

// CallContext executes the given function with context, wrapping it with circuit breaker logic
func (cb *CircuitBreaker) CallContext(ctx context.Context, fn func(ctx context.Context) error) error {
	_, err := cb.ExecuteContext(ctx, func(ctx context.Context) (interface{}, error) {
		return nil, fn(ctx)
	})
	return err
}

// beforeRequest checks if the request should be allowed and updates state
func (cb *CircuitBreaker) beforeRequest() (uint64, error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)

	if state == StateOpen {
		return generation, ErrCircuitBreakerOpen
	} else if state == StateHalfOpen && cb.counts.Requests >= cb.maxRequests {
		return generation, ErrCircuitBreakerOpen
	}

	cb.counts.OnRequest()
	return generation, nil
}

// afterRequest updates the circuit breaker state based on the request result
func (cb *CircuitBreaker) afterRequest(before uint64, success bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)
	if generation != before {
		return
	}

	if success {
		cb.onSuccess(state, now)
	} else {
		cb.onFailure(state, now)
	}
}

// currentState returns the current state and generation
func (cb *CircuitBreaker) currentState(now time.Time) (CircuitBreakerState, uint64) {
	switch cb.state {
	case StateClosed:
		if !cb.expiry.IsZero() && cb.expiry.Before(now) {
			cb.toNewGeneration(now)
		}
	case StateOpen:
		if cb.expiry.Before(now) {
			cb.setState(StateHalfOpen, now)
		}
	}
	return cb.state, cb.generation
}

// onSuccess handles successful requests
func (cb *CircuitBreaker) onSuccess(state CircuitBreakerState, now time.Time) {
	cb.counts.OnSuccess()

	if state == StateHalfOpen {
		cb.setState(StateClosed, now)
	}
}

// onFailure handles failed requests
func (cb *CircuitBreaker) onFailure(state CircuitBreakerState, now time.Time) {
	cb.counts.OnFailure()

	switch state {
	case StateClosed:
		if cb.readyToTrip(cb.counts) {
			cb.setState(StateOpen, now)
		}
	case StateHalfOpen:
		cb.setState(StateOpen, now)
	}
}

// setState changes the state of the circuit breaker
func (cb *CircuitBreaker) setState(state CircuitBreakerState, now time.Time) {
	if cb.state == state {
		return
	}

	prev := cb.state
	cb.state = state

	cb.toNewGeneration(now)

	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, prev, state)
	}
}

// toNewGeneration moves to a new generation and resets counters
func (cb *CircuitBreaker) toNewGeneration(now time.Time) {
	cb.generation++
	cb.counts.Reset()

	var zero time.Time
	switch cb.state {
	case StateClosed:
		if cb.interval == 0 {
			cb.expiry = zero
		} else {
			cb.expiry = now.Add(cb.interval)
		}
	case StateOpen:
		cb.expiry = now.Add(cb.timeout)
	default: // StateHalfOpen
		cb.expiry = zero
	}
}

// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	mutex    sync.RWMutex
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager() *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
	}
}

// GetOrCreate gets an existing circuit breaker or creates a new one
func (cbm *CircuitBreakerManager) GetOrCreate(name string, config *CircuitBreakerConfig) *CircuitBreaker {
	cbm.mutex.RLock()
	if cb, exists := cbm.breakers[name]; exists {
		cbm.mutex.RUnlock()
		return cb
	}
	cbm.mutex.RUnlock()

	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()

	// Double-check after acquiring write lock
	if cb, exists := cbm.breakers[name]; exists {
		return cb
	}

	if config == nil {
		config = DefaultCircuitBreakerConfig(name)
	}
	config.Name = name

	cb := NewCircuitBreaker(config)
	cbm.breakers[name] = cb
	return cb
}

// Get retrieves a circuit breaker by name
func (cbm *CircuitBreakerManager) Get(name string) (*CircuitBreaker, bool) {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()

	cb, exists := cbm.breakers[name]
	return cb, exists
}

// List returns all circuit breakers with their current state
func (cbm *CircuitBreakerManager) List() map[string]CircuitBreakerInfo {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()

	result := make(map[string]CircuitBreakerInfo)
	for name, cb := range cbm.breakers {
		result[name] = CircuitBreakerInfo{
			Name:   name,
			State:  cb.State(),
			Counts: cb.Counts(),
		}
	}
	return result
}

// Remove removes a circuit breaker
func (cbm *CircuitBreakerManager) Remove(name string) bool {
	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()

	if _, exists := cbm.breakers[name]; exists {
		delete(cbm.breakers, name)
		return true
	}
	return false
}

// Clear removes all circuit breakers
func (cbm *CircuitBreakerManager) Clear() {
	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()

	cbm.breakers = make(map[string]*CircuitBreaker)
}

// CircuitBreakerInfo provides information about a circuit breaker
type CircuitBreakerInfo struct {
	Name   string              `json:"name"`
	State  CircuitBreakerState `json:"state"`
	Counts Counts              `json:"counts"`
}

// Predefined errors
var (
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
	ErrTooManyRequests    = errors.New("too many requests")
)

// Global circuit breaker manager instance
var defaultManager = NewCircuitBreakerManager()

// GetCircuitBreaker gets or creates a circuit breaker with default configuration
func GetCircuitBreaker(name string) *CircuitBreaker {
	return defaultManager.GetOrCreate(name, nil)
}

// GetCircuitBreakerWithConfig gets or creates a circuit breaker with custom configuration
func GetCircuitBreakerWithConfig(name string, config *CircuitBreakerConfig) *CircuitBreaker {
	return defaultManager.GetOrCreate(name, config)
}

// ListCircuitBreakers returns information about all circuit breakers
func ListCircuitBreakers() map[string]CircuitBreakerInfo {
	return defaultManager.List()
}

// RemoveCircuitBreaker removes a circuit breaker
func RemoveCircuitBreaker(name string) bool {
	return defaultManager.Remove(name)
}

// ClearCircuitBreakers removes all circuit breakers
func ClearCircuitBreakers() {
	defaultManager.Clear()
}

// Convenience functions for common operations

// ExecuteWithCircuitBreaker executes a function with circuit breaker protection
func ExecuteWithCircuitBreaker(name string, fn func() error) error {
	cb := GetCircuitBreaker(name)
	return cb.Call(fn)
}

// ExecuteWithCircuitBreakerContext executes a function with circuit breaker protection and context
func ExecuteWithCircuitBreakerContext(ctx context.Context, name string, fn func(ctx context.Context) error) error {
	cb := GetCircuitBreaker(name)
	return cb.CallContext(ctx, fn)
}

// ExecuteWithCustomCircuitBreaker executes a function with custom circuit breaker configuration
func ExecuteWithCustomCircuitBreaker(name string, config *CircuitBreakerConfig, fn func() error) error {
	cb := GetCircuitBreakerWithConfig(name, config)
	return cb.Call(fn)
}

// IsCircuitBreakerError checks if an error is a circuit breaker error
func IsCircuitBreakerError(err error) bool {
	return errors.Is(err, ErrCircuitBreakerOpen) || errors.Is(err, ErrTooManyRequests)
}

// CircuitBreakerStats provides statistics about circuit breaker usage
type CircuitBreakerStats struct {
	TotalBreakers    int                           `json:"total_breakers"`
	OpenBreakers     int                           `json:"open_breakers"`
	HalfOpenBreakers int                           `json:"half_open_breakers"`
	ClosedBreakers   int                           `json:"closed_breakers"`
	Breakers         map[string]CircuitBreakerInfo `json:"breakers"`
}

// GetCircuitBreakerStats returns statistics about all circuit breakers
func GetCircuitBreakerStats() CircuitBreakerStats {
	breakers := ListCircuitBreakers()
	stats := CircuitBreakerStats{
		TotalBreakers: len(breakers),
		Breakers:      breakers,
	}

	for _, info := range breakers {
		switch info.State {
		case StateOpen:
			stats.OpenBreakers++
		case StateHalfOpen:
			stats.HalfOpenBreakers++
		case StateClosed:
			stats.ClosedBreakers++
		}
	}

	return stats
}

// ResetCircuitBreaker resets a circuit breaker to closed state
func ResetCircuitBreaker(name string) error {
	cbm := defaultManager
	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()

	cb, exists := cbm.breakers[name]
	if !exists {
		return fmt.Errorf("circuit breaker %s not found", name)
	}

	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.setState(StateClosed, time.Now())
	return nil
}

// ForceOpenCircuitBreaker forces a circuit breaker to open state
func ForceOpenCircuitBreaker(name string) error {
	cbm := defaultManager
	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()

	cb, exists := cbm.breakers[name]
	if !exists {
		return fmt.Errorf("circuit breaker %s not found", name)
	}

	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.setState(StateOpen, time.Now())
	return nil
}
