package observability

import (
	"context"
	"time"

	"github.com/Kisanlink/kisanlink-ecom/internal/cache"
	"github.com/Kisanlink/kisanlink-ecom/internal/resilience"
)

// MetricsAwareCacheWrapper wraps cache operations with metrics collection
type MetricsAwareCacheWrapper struct {
	cache              cache.Cache
	metricsIntegration *MetricsIntegration
	cacheLevel         string
}

// NewMetricsAwareCacheWrapper creates a cache wrapper with metrics
func NewMetricsAwareCacheWrapper(cache cache.Cache, metricsIntegration *MetricsIntegration, level string) *MetricsAwareCacheWrapper {
	return &MetricsAwareCacheWrapper{
		cache:              cache,
		metricsIntegration: metricsIntegration,
		cacheLevel:         level,
	}
}

// Get implements cache.Cache interface with metrics
func (w *MetricsAwareCacheWrapper) Get(ctx context.Context, key string) (interface{}, bool, error) {
	start := time.Now()
	value, found, err := w.cache.Get(ctx, key)
	duration := time.Since(start)

	if err != nil || !found {
		w.metricsIntegration.RecordCacheMiss(ctx, w.cacheLevel, "get", duration)
	} else {
		w.metricsIntegration.RecordCacheHit(ctx, w.cacheLevel, "get", duration)
	}

	return value, found, err
}

// Set implements cache.Cache interface with metrics
func (w *MetricsAwareCacheWrapper) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	start := time.Now()
	err := w.cache.Set(ctx, key, value, ttl)
	duration := time.Since(start)

	w.metricsIntegration.RecordCacheHit(ctx, w.cacheLevel, "set", duration)
	return err
}

// Delete implements cache.Cache interface with metrics
func (w *MetricsAwareCacheWrapper) Delete(ctx context.Context, key string) error {
	start := time.Now()
	err := w.cache.Delete(ctx, key)
	duration := time.Since(start)

	w.metricsIntegration.RecordCacheHit(ctx, w.cacheLevel, "delete", duration)
	return err
}

// Exists implements cache.Cache interface with metrics
func (w *MetricsAwareCacheWrapper) Exists(ctx context.Context, key string) (bool, error) {
	start := time.Now()
	exists, err := w.cache.Exists(ctx, key)
	duration := time.Since(start)

	if exists {
		w.metricsIntegration.RecordCacheHit(ctx, w.cacheLevel, "exists", duration)
	} else {
		w.metricsIntegration.RecordCacheMiss(ctx, w.cacheLevel, "exists", duration)
	}

	return exists, err
}

// TTL implements cache.Cache interface with metrics
func (w *MetricsAwareCacheWrapper) TTL(ctx context.Context, key string) (time.Duration, error) {
	start := time.Now()
	ttl, err := w.cache.TTL(ctx, key)
	duration := time.Since(start)

	w.metricsIntegration.RecordCacheHit(ctx, w.cacheLevel, "ttl", duration)
	return ttl, err
}

// Clear implements cache.Cache interface with metrics
func (w *MetricsAwareCacheWrapper) Clear(ctx context.Context) error {
	start := time.Now()
	err := w.cache.Clear(ctx)
	duration := time.Since(start)

	w.metricsIntegration.RecordCacheHit(ctx, w.cacheLevel, "clear", duration)
	return err
}

// Stats implements cache.Cache interface
func (w *MetricsAwareCacheWrapper) Stats() cache.CacheStats {
	return w.cache.Stats()
}

// MetricsAwareCircuitBreakerWrapper wraps circuit breaker operations with metrics
type MetricsAwareCircuitBreakerWrapper struct {
	circuitBreaker     *resilience.CircuitBreaker
	metricsIntegration *MetricsIntegration
	name               string
}

// NewMetricsAwareCircuitBreakerWrapper creates a circuit breaker wrapper with metrics
func NewMetricsAwareCircuitBreakerWrapper(cb *resilience.CircuitBreaker, metricsIntegration *MetricsIntegration) *MetricsAwareCircuitBreakerWrapper {
	wrapper := &MetricsAwareCircuitBreakerWrapper{
		circuitBreaker:     cb,
		metricsIntegration: metricsIntegration,
		name:               cb.Name(),
	}

	// Record initial state
	wrapper.recordCurrentState(context.Background())

	return wrapper
}

// Execute wraps circuit breaker execution with metrics
func (w *MetricsAwareCircuitBreakerWrapper) Execute(fn func() (interface{}, error)) (interface{}, error) {
	ctx := context.Background()

	// Check if request will be allowed
	state := w.circuitBreaker.State()
	if state == resilience.StateOpen {
		w.metricsIntegration.RecordCircuitBreakerRequest(ctx, w.name, false)
		return w.circuitBreaker.Execute(fn)
	}

	w.metricsIntegration.RecordCircuitBreakerRequest(ctx, w.name, true)
	result, err := w.circuitBreaker.Execute(fn)

	// Record state after execution (might have changed)
	w.recordCurrentState(ctx)

	return result, err
}

// ExecuteContext wraps circuit breaker execution with context and metrics
func (w *MetricsAwareCircuitBreakerWrapper) ExecuteContext(ctx context.Context, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	// Check if request will be allowed
	state := w.circuitBreaker.State()
	if state == resilience.StateOpen {
		w.metricsIntegration.RecordCircuitBreakerRequest(ctx, w.name, false)
		return w.circuitBreaker.ExecuteContext(ctx, fn)
	}

	w.metricsIntegration.RecordCircuitBreakerRequest(ctx, w.name, true)
	result, err := w.circuitBreaker.ExecuteContext(ctx, fn)

	// Record state after execution (might have changed)
	w.recordCurrentState(ctx)

	return result, err
}

// Call wraps circuit breaker call with metrics
func (w *MetricsAwareCircuitBreakerWrapper) Call(fn func() error) error {
	ctx := context.Background()

	// Check if request will be allowed
	state := w.circuitBreaker.State()
	if state == resilience.StateOpen {
		w.metricsIntegration.RecordCircuitBreakerRequest(ctx, w.name, false)
		return w.circuitBreaker.Call(fn)
	}

	w.metricsIntegration.RecordCircuitBreakerRequest(ctx, w.name, true)
	err := w.circuitBreaker.Call(fn)

	// Record state after execution (might have changed)
	w.recordCurrentState(ctx)

	return err
}

// CallContext wraps circuit breaker call with context and metrics
func (w *MetricsAwareCircuitBreakerWrapper) CallContext(ctx context.Context, fn func(context.Context) error) error {
	// Check if request will be allowed
	state := w.circuitBreaker.State()
	if state == resilience.StateOpen {
		w.metricsIntegration.RecordCircuitBreakerRequest(ctx, w.name, false)
		return w.circuitBreaker.CallContext(ctx, fn)
	}

	w.metricsIntegration.RecordCircuitBreakerRequest(ctx, w.name, true)
	err := w.circuitBreaker.CallContext(ctx, fn)

	// Record state after execution (might have changed)
	w.recordCurrentState(ctx)

	return err
}

// Name returns the circuit breaker name
func (w *MetricsAwareCircuitBreakerWrapper) Name() string {
	return w.circuitBreaker.Name()
}

// State returns the current circuit breaker state
func (w *MetricsAwareCircuitBreakerWrapper) State() resilience.CircuitBreakerState {
	return w.circuitBreaker.State()
}

// Counts returns the circuit breaker counts
func (w *MetricsAwareCircuitBreakerWrapper) Counts() resilience.Counts {
	return w.circuitBreaker.Counts()
}

// recordCurrentState records the current circuit breaker state as a metric
func (w *MetricsAwareCircuitBreakerWrapper) recordCurrentState(ctx context.Context) {
	state := w.circuitBreaker.State()
	w.metricsIntegration.RecordCircuitBreakerState(ctx, w.name, state)
}

// MetricsAwareMultiLevelCacheWrapper wraps multi-level cache with metrics
type MetricsAwareMultiLevelCacheWrapper struct {
	cache              *cache.MultiLevelCache
	metricsIntegration *MetricsIntegration
}

// NewMetricsAwareMultiLevelCacheWrapper creates a multi-level cache wrapper with metrics
func NewMetricsAwareMultiLevelCacheWrapper(cache *cache.MultiLevelCache, metricsIntegration *MetricsIntegration) *MetricsAwareMultiLevelCacheWrapper {
	return &MetricsAwareMultiLevelCacheWrapper{
		cache:              cache,
		metricsIntegration: metricsIntegration,
	}
}

// Get implements cache interface with multi-level metrics
func (w *MetricsAwareMultiLevelCacheWrapper) Get(ctx context.Context, key string) (interface{}, bool, error) {
	start := time.Now()
	value, found, err := w.cache.Get(ctx, key)
	duration := time.Since(start)

	// The multi-level cache will handle which level was hit
	// For now, we record it as a general cache operation
	if err != nil || !found {
		w.metricsIntegration.RecordCacheMiss(ctx, "multi-level", "get", duration)
	} else {
		w.metricsIntegration.RecordCacheHit(ctx, "multi-level", "get", duration)
	}

	return value, found, err
}

// Set implements cache interface with multi-level metrics
func (w *MetricsAwareMultiLevelCacheWrapper) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	start := time.Now()
	err := w.cache.Set(ctx, key, value, ttl)
	duration := time.Since(start)

	w.metricsIntegration.RecordCacheHit(ctx, "multi-level", "set", duration)
	return err
}

// Delete implements cache interface with multi-level metrics
func (w *MetricsAwareMultiLevelCacheWrapper) Delete(ctx context.Context, key string) error {
	start := time.Now()
	err := w.cache.Delete(ctx, key)
	duration := time.Since(start)

	w.metricsIntegration.RecordCacheHit(ctx, "multi-level", "delete", duration)
	return err
}

// Exists implements cache interface with multi-level metrics
func (w *MetricsAwareMultiLevelCacheWrapper) Exists(ctx context.Context, key string) (bool, error) {
	start := time.Now()
	exists, err := w.cache.Exists(ctx, key)
	duration := time.Since(start)

	if exists {
		w.metricsIntegration.RecordCacheHit(ctx, "multi-level", "exists", duration)
	} else {
		w.metricsIntegration.RecordCacheMiss(ctx, "multi-level", "exists", duration)
	}

	return exists, err
}

// TTL implements cache interface with multi-level metrics
func (w *MetricsAwareMultiLevelCacheWrapper) TTL(ctx context.Context, key string) (time.Duration, error) {
	start := time.Now()
	ttl, err := w.cache.TTL(ctx, key)
	duration := time.Since(start)

	w.metricsIntegration.RecordCacheHit(ctx, "multi-level", "ttl", duration)
	return ttl, err
}

// Clear implements cache interface with multi-level metrics
func (w *MetricsAwareMultiLevelCacheWrapper) Clear(ctx context.Context) error {
	start := time.Now()
	err := w.cache.Clear(ctx)
	duration := time.Since(start)

	w.metricsIntegration.RecordCacheHit(ctx, "multi-level", "clear", duration)
	return err
}

// Stats implements cache interface - returns empty stats for now
func (w *MetricsAwareMultiLevelCacheWrapper) Stats() cache.CacheStats {
	// Return empty stats since MultiLevelCache.Stats() returns a map
	return cache.CacheStats{}
}

// Promote records cache level promotion
func (w *MetricsAwareMultiLevelCacheWrapper) RecordPromotion(ctx context.Context, fromLevel, toLevel string) {
	w.metricsIntegration.RecordCacheEviction(ctx, "promotion:"+fromLevel+"->"+toLevel)
}

// Demote records cache level demotion
func (w *MetricsAwareMultiLevelCacheWrapper) RecordDemotion(ctx context.Context, fromLevel, toLevel string) {
	w.metricsIntegration.RecordCacheEviction(ctx, "demotion:"+fromLevel+"->"+toLevel)
}

// MetricsCollectionConfig provides configuration for metrics collection
type MetricsCollectionConfig struct {
	// Enable specific metric types
	EnableHTTPMetrics           bool
	EnableDatabaseMetrics       bool
	EnableCacheMetrics          bool
	EnableCircuitBreakerMetrics bool
	EnableRateLimiterMetrics    bool
	EnableBusinessMetrics       bool

	// Collection intervals
	MetricsCollectionInterval time.Duration
	HealthCheckInterval       time.Duration

	// Metric retention and sampling
	MetricRetentionPeriod time.Duration
	SamplingRate          float64
}

// DefaultMetricsCollectionConfig returns default metrics collection configuration
func DefaultMetricsCollectionConfig() *MetricsCollectionConfig {
	return &MetricsCollectionConfig{
		EnableHTTPMetrics:           true,
		EnableDatabaseMetrics:       true,
		EnableCacheMetrics:          true,
		EnableCircuitBreakerMetrics: true,
		EnableRateLimiterMetrics:    true,
		EnableBusinessMetrics:       true,

		MetricsCollectionInterval: 30 * time.Second,
		HealthCheckInterval:       60 * time.Second,

		MetricRetentionPeriod: 24 * time.Hour,
		SamplingRate:          1.0, // 100% sampling by default
	}
}
