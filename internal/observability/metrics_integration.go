package observability

import (
	"context"
	"fmt"
	"time"

	"kisanlink-ecom/internal/resilience"

	"go.opentelemetry.io/otel/attribute"
	metricapi "go.opentelemetry.io/otel/metric"
)

// MetricsIntegration handles metrics collection across all system components
type MetricsIntegration struct {
	telemetryManager *TelemetryManager
	metricsCollector *MetricsCollector

	// System component metrics
	httpMetrics           *HTTPMetrics
	dbMetrics             *DatabaseMetrics
	cacheMetrics          *CacheMetrics
	circuitBreakerMetrics *CircuitBreakerMetrics
	rateLimiterMetrics    *RateLimiterMetrics
	businessMetrics       *BusinessMetrics
}

// NewMetricsIntegration creates a new metrics integration instance
func NewMetricsIntegration(telemetryManager *TelemetryManager) (*MetricsIntegration, error) {
	metricsCollector, err := NewMetricsCollector(telemetryManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics collector: %w", err)
	}

	mi := &MetricsIntegration{
		telemetryManager: telemetryManager,
		metricsCollector: metricsCollector,
	}

	// Initialize component metrics
	if err := mi.initializeComponentMetrics(); err != nil {
		return nil, fmt.Errorf("failed to initialize component metrics: %w", err)
	}

	return mi, nil
}

// HTTPMetrics tracks HTTP-related metrics
type HTTPMetrics struct {
	requestsTotal      metricapi.Int64Counter
	requestDuration    metricapi.Float64Histogram
	requestsInFlight   metricapi.Int64UpDownCounter
	responseSize       metricapi.Int64Histogram
	requestSize        metricapi.Int64Histogram
	requestsByStatus   metricapi.Int64Counter
	requestsByMethod   metricapi.Int64Counter
	requestsByEndpoint metricapi.Int64Counter
}

// DatabaseMetrics tracks database-related metrics
type DatabaseMetrics struct {
	queriesTotal        metricapi.Int64Counter
	queryDuration       metricapi.Float64Histogram
	connections         metricapi.Int64UpDownCounter
	transactionsTotal   metricapi.Int64Counter
	transactionDuration metricapi.Float64Histogram
	deadlocks           metricapi.Int64Counter
	poolStats           metricapi.Int64UpDownCounter
}

// CacheMetrics tracks cache-related metrics
type CacheMetrics struct {
	hits              metricapi.Int64Counter
	misses            metricapi.Int64Counter
	operations        metricapi.Int64Counter
	operationDuration metricapi.Float64Histogram
	size              metricapi.Int64UpDownCounter
	evictions         metricapi.Int64Counter
	levelPromotions   metricapi.Int64Counter
	levelDemotions    metricapi.Int64Counter
}

// CircuitBreakerMetrics tracks circuit breaker metrics
type CircuitBreakerMetrics struct {
	state            metricapi.Int64UpDownCounter
	requestsAllowed  metricapi.Int64Counter
	requestsRejected metricapi.Int64Counter
	stateTransitions metricapi.Int64Counter
	failureRate      metricapi.Float64Histogram
	successRate      metricapi.Float64Histogram
}

// RateLimiterMetrics tracks rate limiting metrics
type RateLimiterMetrics struct {
	requestsAllowed  metricapi.Int64Counter
	requestsRejected metricapi.Int64Counter
	bucketsCreated   metricapi.Int64Counter
	bucketsCleanedUp metricapi.Int64Counter
	currentTokens    metricapi.Float64UpDownCounter
	limitUtilization metricapi.Float64Histogram
}

// BusinessMetrics tracks business-specific metrics
type BusinessMetrics struct {
	ordersCreated     metricapi.Int64Counter
	ordersCompleted   metricapi.Int64Counter
	ordersCancelled   metricapi.Int64Counter
	productsViewed    metricapi.Int64Counter
	productsCreated   metricapi.Int64Counter
	revenue           metricapi.Float64Counter
	userRegistrations metricapi.Int64Counter
	activeUsers       metricapi.Int64UpDownCounter
}

// initializeComponentMetrics initializes all component-specific metrics
func (mi *MetricsIntegration) initializeComponentMetrics() error {
	meter := mi.telemetryManager.GetMeter()

	// Initialize HTTP metrics
	if err := mi.initializeHTTPMetrics(meter); err != nil {
		return fmt.Errorf("failed to initialize HTTP metrics: %w", err)
	}

	// Initialize database metrics
	if err := mi.initializeDatabaseMetrics(meter); err != nil {
		return fmt.Errorf("failed to initialize database metrics: %w", err)
	}

	// Initialize cache metrics
	if err := mi.initializeCacheMetrics(meter); err != nil {
		return fmt.Errorf("failed to initialize cache metrics: %w", err)
	}

	// Initialize circuit breaker metrics
	if err := mi.initializeCircuitBreakerMetrics(meter); err != nil {
		return fmt.Errorf("failed to initialize circuit breaker metrics: %w", err)
	}

	// Initialize rate limiter metrics
	if err := mi.initializeRateLimiterMetrics(meter); err != nil {
		return fmt.Errorf("failed to initialize rate limiter metrics: %w", err)
	}

	// Initialize business metrics
	if err := mi.initializeBusinessMetrics(meter); err != nil {
		return fmt.Errorf("failed to initialize business metrics: %w", err)
	}

	return nil
}

// initializeHTTPMetrics creates HTTP-related metrics
func (mi *MetricsIntegration) initializeHTTPMetrics(meter metricapi.Meter) error {
	var err error
	mi.httpMetrics = &HTTPMetrics{}

	mi.httpMetrics.requestsTotal, err = meter.Int64Counter(
		"http_requests_total",
		metricapi.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		return err
	}

	mi.httpMetrics.requestDuration, err = meter.Float64Histogram(
		"http_request_duration_seconds",
		metricapi.WithDescription("HTTP request duration in seconds"),
		metricapi.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	mi.httpMetrics.requestsInFlight, err = meter.Int64UpDownCounter(
		"http_requests_in_flight",
		metricapi.WithDescription("Number of HTTP requests currently being processed"),
	)
	if err != nil {
		return err
	}

	mi.httpMetrics.responseSize, err = meter.Int64Histogram(
		"http_response_size_bytes",
		metricapi.WithDescription("Size of HTTP response in bytes"),
		metricapi.WithUnit("By"),
	)
	if err != nil {
		return err
	}

	mi.httpMetrics.requestSize, err = meter.Int64Histogram(
		"http_request_size_bytes",
		metricapi.WithDescription("Size of HTTP request in bytes"),
		metricapi.WithUnit("By"),
	)
	if err != nil {
		return err
	}

	mi.httpMetrics.requestsByStatus, err = meter.Int64Counter(
		"http_requests_by_status_total",
		metricapi.WithDescription("Total HTTP requests by status code"),
	)
	if err != nil {
		return err
	}

	mi.httpMetrics.requestsByMethod, err = meter.Int64Counter(
		"http_requests_by_method_total",
		metricapi.WithDescription("Total HTTP requests by method"),
	)
	if err != nil {
		return err
	}

	mi.httpMetrics.requestsByEndpoint, err = meter.Int64Counter(
		"http_requests_by_endpoint_total",
		metricapi.WithDescription("Total HTTP requests by endpoint"),
	)
	if err != nil {
		return err
	}

	return nil
}

// initializeDatabaseMetrics creates database-related metrics
func (mi *MetricsIntegration) initializeDatabaseMetrics(meter metricapi.Meter) error {
	var err error
	mi.dbMetrics = &DatabaseMetrics{}

	mi.dbMetrics.queriesTotal, err = meter.Int64Counter(
		"db_queries_total",
		metricapi.WithDescription("Total number of database queries"),
	)
	if err != nil {
		return err
	}

	mi.dbMetrics.queryDuration, err = meter.Float64Histogram(
		"db_query_duration_seconds",
		metricapi.WithDescription("Database query duration in seconds"),
		metricapi.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	mi.dbMetrics.connections, err = meter.Int64UpDownCounter(
		"db_connections_active",
		metricapi.WithDescription("Number of active database connections"),
	)
	if err != nil {
		return err
	}

	mi.dbMetrics.transactionsTotal, err = meter.Int64Counter(
		"db_transactions_total",
		metricapi.WithDescription("Total number of database transactions"),
	)
	if err != nil {
		return err
	}

	mi.dbMetrics.transactionDuration, err = meter.Float64Histogram(
		"db_transaction_duration_seconds",
		metricapi.WithDescription("Database transaction duration in seconds"),
		metricapi.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	mi.dbMetrics.deadlocks, err = meter.Int64Counter(
		"db_deadlocks_total",
		metricapi.WithDescription("Total number of database deadlocks"),
	)
	if err != nil {
		return err
	}

	mi.dbMetrics.poolStats, err = meter.Int64UpDownCounter(
		"db_pool_connections",
		metricapi.WithDescription("Database connection pool statistics"),
	)
	if err != nil {
		return err
	}

	return nil
}

// initializeCacheMetrics creates cache-related metrics
func (mi *MetricsIntegration) initializeCacheMetrics(meter metricapi.Meter) error {
	var err error
	mi.cacheMetrics = &CacheMetrics{}

	mi.cacheMetrics.hits, err = meter.Int64Counter(
		"cache_hits_total",
		metricapi.WithDescription("Total number of cache hits"),
	)
	if err != nil {
		return err
	}

	mi.cacheMetrics.misses, err = meter.Int64Counter(
		"cache_misses_total",
		metricapi.WithDescription("Total number of cache misses"),
	)
	if err != nil {
		return err
	}

	mi.cacheMetrics.operations, err = meter.Int64Counter(
		"cache_operations_total",
		metricapi.WithDescription("Total number of cache operations"),
	)
	if err != nil {
		return err
	}

	mi.cacheMetrics.operationDuration, err = meter.Float64Histogram(
		"cache_operation_duration_seconds",
		metricapi.WithDescription("Cache operation duration in seconds"),
		metricapi.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	mi.cacheMetrics.size, err = meter.Int64UpDownCounter(
		"cache_size_items",
		metricapi.WithDescription("Current number of items in cache"),
	)
	if err != nil {
		return err
	}

	mi.cacheMetrics.evictions, err = meter.Int64Counter(
		"cache_evictions_total",
		metricapi.WithDescription("Total number of cache evictions"),
	)
	if err != nil {
		return err
	}

	mi.cacheMetrics.levelPromotions, err = meter.Int64Counter(
		"cache_level_promotions_total",
		metricapi.WithDescription("Total number of cache level promotions"),
	)
	if err != nil {
		return err
	}

	mi.cacheMetrics.levelDemotions, err = meter.Int64Counter(
		"cache_level_demotions_total",
		metricapi.WithDescription("Total number of cache level demotions"),
	)
	if err != nil {
		return err
	}

	return nil
}

// initializeCircuitBreakerMetrics creates circuit breaker metrics
func (mi *MetricsIntegration) initializeCircuitBreakerMetrics(meter metricapi.Meter) error {
	var err error
	mi.circuitBreakerMetrics = &CircuitBreakerMetrics{}

	mi.circuitBreakerMetrics.state, err = meter.Int64UpDownCounter(
		"circuit_breaker_state",
		metricapi.WithDescription("Circuit breaker state (0=closed, 1=open, 2=half-open)"),
	)
	if err != nil {
		return err
	}

	mi.circuitBreakerMetrics.requestsAllowed, err = meter.Int64Counter(
		"circuit_breaker_requests_allowed_total",
		metricapi.WithDescription("Total requests allowed by circuit breaker"),
	)
	if err != nil {
		return err
	}

	mi.circuitBreakerMetrics.requestsRejected, err = meter.Int64Counter(
		"circuit_breaker_requests_rejected_total",
		metricapi.WithDescription("Total requests rejected by circuit breaker"),
	)
	if err != nil {
		return err
	}

	mi.circuitBreakerMetrics.stateTransitions, err = meter.Int64Counter(
		"circuit_breaker_state_transitions_total",
		metricapi.WithDescription("Total circuit breaker state transitions"),
	)
	if err != nil {
		return err
	}

	mi.circuitBreakerMetrics.failureRate, err = meter.Float64Histogram(
		"circuit_breaker_failure_rate",
		metricapi.WithDescription("Circuit breaker failure rate"),
	)
	if err != nil {
		return err
	}

	mi.circuitBreakerMetrics.successRate, err = meter.Float64Histogram(
		"circuit_breaker_success_rate",
		metricapi.WithDescription("Circuit breaker success rate"),
	)
	if err != nil {
		return err
	}

	return nil
}

// initializeRateLimiterMetrics creates rate limiter metrics
func (mi *MetricsIntegration) initializeRateLimiterMetrics(meter metricapi.Meter) error {
	var err error
	mi.rateLimiterMetrics = &RateLimiterMetrics{}

	mi.rateLimiterMetrics.requestsAllowed, err = meter.Int64Counter(
		"rate_limiter_requests_allowed_total",
		metricapi.WithDescription("Total requests allowed by rate limiter"),
	)
	if err != nil {
		return err
	}

	mi.rateLimiterMetrics.requestsRejected, err = meter.Int64Counter(
		"rate_limiter_requests_rejected_total",
		metricapi.WithDescription("Total requests rejected by rate limiter"),
	)
	if err != nil {
		return err
	}

	mi.rateLimiterMetrics.bucketsCreated, err = meter.Int64Counter(
		"rate_limiter_buckets_created_total",
		metricapi.WithDescription("Total rate limiter buckets created"),
	)
	if err != nil {
		return err
	}

	mi.rateLimiterMetrics.bucketsCleanedUp, err = meter.Int64Counter(
		"rate_limiter_buckets_cleaned_up_total",
		metricapi.WithDescription("Total rate limiter buckets cleaned up"),
	)
	if err != nil {
		return err
	}

	mi.rateLimiterMetrics.currentTokens, err = meter.Float64UpDownCounter(
		"rate_limiter_current_tokens",
		metricapi.WithDescription("Current number of tokens in rate limiter"),
	)
	if err != nil {
		return err
	}

	mi.rateLimiterMetrics.limitUtilization, err = meter.Float64Histogram(
		"rate_limiter_utilization",
		metricapi.WithDescription("Rate limiter utilization percentage"),
	)
	if err != nil {
		return err
	}

	return nil
}

// initializeBusinessMetrics creates business-specific metrics
func (mi *MetricsIntegration) initializeBusinessMetrics(meter metricapi.Meter) error {
	var err error
	mi.businessMetrics = &BusinessMetrics{}

	mi.businessMetrics.ordersCreated, err = meter.Int64Counter(
		"business_orders_created_total",
		metricapi.WithDescription("Total number of orders created"),
	)
	if err != nil {
		return err
	}

	mi.businessMetrics.ordersCompleted, err = meter.Int64Counter(
		"business_orders_completed_total",
		metricapi.WithDescription("Total number of orders completed"),
	)
	if err != nil {
		return err
	}

	mi.businessMetrics.ordersCancelled, err = meter.Int64Counter(
		"business_orders_cancelled_total",
		metricapi.WithDescription("Total number of orders cancelled"),
	)
	if err != nil {
		return err
	}

	mi.businessMetrics.productsViewed, err = meter.Int64Counter(
		"business_products_viewed_total",
		metricapi.WithDescription("Total number of product views"),
	)
	if err != nil {
		return err
	}

	mi.businessMetrics.productsCreated, err = meter.Int64Counter(
		"business_products_created_total",
		metricapi.WithDescription("Total number of products created"),
	)
	if err != nil {
		return err
	}

	mi.businessMetrics.revenue, err = meter.Float64Counter(
		"business_revenue_total",
		metricapi.WithDescription("Total revenue"),
	)
	if err != nil {
		return err
	}

	mi.businessMetrics.userRegistrations, err = meter.Int64Counter(
		"business_user_registrations_total",
		metricapi.WithDescription("Total number of user registrations"),
	)
	if err != nil {
		return err
	}

	mi.businessMetrics.activeUsers, err = meter.Int64UpDownCounter(
		"business_active_users",
		metricapi.WithDescription("Number of currently active users"),
	)
	if err != nil {
		return err
	}

	return nil
}

// HTTP Metrics Methods

// RecordHTTPRequest records an HTTP request
func (mi *MetricsIntegration) RecordHTTPRequest(ctx context.Context, method, endpoint string, statusCode int, duration time.Duration, requestSize, responseSize int64) {
	attrs := []attribute.KeyValue{
		attribute.String("method", method),
		attribute.String("endpoint", endpoint),
		attribute.Int("status_code", statusCode),
	}

	mi.httpMetrics.requestsTotal.Add(ctx, 1, metricapi.WithAttributes(attrs...))
	mi.httpMetrics.requestDuration.Record(ctx, duration.Seconds(), metricapi.WithAttributes(attrs...))
	mi.httpMetrics.requestSize.Record(ctx, requestSize, metricapi.WithAttributes(attrs...))
	mi.httpMetrics.responseSize.Record(ctx, responseSize, metricapi.WithAttributes(attrs...))

	mi.httpMetrics.requestsByStatus.Add(ctx, 1, metricapi.WithAttributes(attribute.Int("status_code", statusCode)))
	mi.httpMetrics.requestsByMethod.Add(ctx, 1, metricapi.WithAttributes(attribute.String("method", method)))
	mi.httpMetrics.requestsByEndpoint.Add(ctx, 1, metricapi.WithAttributes(attribute.String("endpoint", endpoint)))
}

// RecordHTTPRequestStart records the start of an HTTP request
func (mi *MetricsIntegration) RecordHTTPRequestStart(ctx context.Context) {
	mi.httpMetrics.requestsInFlight.Add(ctx, 1)
}

// RecordHTTPRequestEnd records the end of an HTTP request
func (mi *MetricsIntegration) RecordHTTPRequestEnd(ctx context.Context) {
	mi.httpMetrics.requestsInFlight.Add(ctx, -1)
}

// Database Metrics Methods

// RecordDatabaseQuery records a database query
func (mi *MetricsIntegration) RecordDatabaseQuery(ctx context.Context, operation string, duration time.Duration, success bool) {
	attrs := []attribute.KeyValue{
		attribute.String("operation", operation),
		attribute.Bool("success", success),
	}

	mi.dbMetrics.queriesTotal.Add(ctx, 1, metricapi.WithAttributes(attrs...))
	mi.dbMetrics.queryDuration.Record(ctx, duration.Seconds(), metricapi.WithAttributes(attrs...))
}

// RecordDatabaseTransaction records a database transaction
func (mi *MetricsIntegration) RecordDatabaseTransaction(ctx context.Context, duration time.Duration, success bool) {
	attrs := []attribute.KeyValue{
		attribute.Bool("success", success),
	}

	mi.dbMetrics.transactionsTotal.Add(ctx, 1, metricapi.WithAttributes(attrs...))
	mi.dbMetrics.transactionDuration.Record(ctx, duration.Seconds(), metricapi.WithAttributes(attrs...))
}

// RecordDatabaseDeadlock records a database deadlock
func (mi *MetricsIntegration) RecordDatabaseDeadlock(ctx context.Context) {
	mi.dbMetrics.deadlocks.Add(ctx, 1)
}

// Cache Metrics Methods

// RecordCacheHit records a cache hit
func (mi *MetricsIntegration) RecordCacheHit(ctx context.Context, cacheLevel string, operation string, duration time.Duration) {
	attrs := []attribute.KeyValue{
		attribute.String("level", cacheLevel),
		attribute.String("operation", operation),
	}

	mi.cacheMetrics.hits.Add(ctx, 1, metricapi.WithAttributes(attrs...))
	mi.cacheMetrics.operations.Add(ctx, 1, metricapi.WithAttributes(attrs...))
	mi.cacheMetrics.operationDuration.Record(ctx, duration.Seconds(), metricapi.WithAttributes(attrs...))
}

// RecordCacheMiss records a cache miss
func (mi *MetricsIntegration) RecordCacheMiss(ctx context.Context, cacheLevel string, operation string, duration time.Duration) {
	attrs := []attribute.KeyValue{
		attribute.String("level", cacheLevel),
		attribute.String("operation", operation),
	}

	mi.cacheMetrics.misses.Add(ctx, 1, metricapi.WithAttributes(attrs...))
	mi.cacheMetrics.operations.Add(ctx, 1, metricapi.WithAttributes(attrs...))
	mi.cacheMetrics.operationDuration.Record(ctx, duration.Seconds(), metricapi.WithAttributes(attrs...))
}

// RecordCacheEviction records a cache eviction
func (mi *MetricsIntegration) RecordCacheEviction(ctx context.Context, cacheLevel string) {
	attrs := []attribute.KeyValue{
		attribute.String("level", cacheLevel),
	}

	mi.cacheMetrics.evictions.Add(ctx, 1, metricapi.WithAttributes(attrs...))
}

// Circuit Breaker Metrics Methods

// RecordCircuitBreakerState records circuit breaker state
func (mi *MetricsIntegration) RecordCircuitBreakerState(ctx context.Context, name string, state resilience.CircuitBreakerState) {
	attrs := []attribute.KeyValue{
		attribute.String("circuit_breaker", name),
	}

	stateValue := int64(state)
	mi.circuitBreakerMetrics.state.Add(ctx, stateValue, metricapi.WithAttributes(attrs...))
}

// RecordCircuitBreakerRequest records a circuit breaker request
func (mi *MetricsIntegration) RecordCircuitBreakerRequest(ctx context.Context, name string, allowed bool) {
	attrs := []attribute.KeyValue{
		attribute.String("circuit_breaker", name),
	}

	if allowed {
		mi.circuitBreakerMetrics.requestsAllowed.Add(ctx, 1, metricapi.WithAttributes(attrs...))
	} else {
		mi.circuitBreakerMetrics.requestsRejected.Add(ctx, 1, metricapi.WithAttributes(attrs...))
	}
}

// Rate Limiter Metrics Methods

// RecordRateLimiterRequest records a rate limiter request
func (mi *MetricsIntegration) RecordRateLimiterRequest(ctx context.Context, limiterType string, allowed bool) {
	attrs := []attribute.KeyValue{
		attribute.String("limiter_type", limiterType),
	}

	if allowed {
		mi.rateLimiterMetrics.requestsAllowed.Add(ctx, 1, metricapi.WithAttributes(attrs...))
	} else {
		mi.rateLimiterMetrics.requestsRejected.Add(ctx, 1, metricapi.WithAttributes(attrs...))
	}
}

// Business Metrics Methods

// RecordOrderCreated records a new order creation
func (mi *MetricsIntegration) RecordOrderCreated(ctx context.Context, userID string, amount float64) {
	attrs := []attribute.KeyValue{
		attribute.String("user_id", userID),
	}

	mi.businessMetrics.ordersCreated.Add(ctx, 1, metricapi.WithAttributes(attrs...))
	mi.businessMetrics.revenue.Add(ctx, amount)
}

// RecordOrderCompleted records an order completion
func (mi *MetricsIntegration) RecordOrderCompleted(ctx context.Context, userID string) {
	attrs := []attribute.KeyValue{
		attribute.String("user_id", userID),
	}

	mi.businessMetrics.ordersCompleted.Add(ctx, 1, metricapi.WithAttributes(attrs...))
}

// RecordProductView records a product view
func (mi *MetricsIntegration) RecordProductView(ctx context.Context, productID, userID string) {
	attrs := []attribute.KeyValue{
		attribute.String("product_id", productID),
		attribute.String("user_id", userID),
	}

	mi.businessMetrics.productsViewed.Add(ctx, 1, metricapi.WithAttributes(attrs...))
}

// RecordUserRegistration records a new user registration
func (mi *MetricsIntegration) RecordUserRegistration(ctx context.Context) {
	mi.businessMetrics.userRegistrations.Add(ctx, 1)
}

// GetMetricsCollector returns the metrics collector for direct access
func (mi *MetricsIntegration) GetMetricsCollector() *MetricsCollector {
	return mi.metricsCollector
}

// GetHTTPMetrics returns HTTP metrics for direct access
func (mi *MetricsIntegration) GetHTTPMetrics() *HTTPMetrics {
	return mi.httpMetrics
}

// GetDatabaseMetrics returns database metrics for direct access
func (mi *MetricsIntegration) GetDatabaseMetrics() *DatabaseMetrics {
	return mi.dbMetrics
}

// GetCacheMetrics returns cache metrics for direct access
func (mi *MetricsIntegration) GetCacheMetrics() *CacheMetrics {
	return mi.cacheMetrics
}

// GetCircuitBreakerMetrics returns circuit breaker metrics for direct access
func (mi *MetricsIntegration) GetCircuitBreakerMetrics() *CircuitBreakerMetrics {
	return mi.circuitBreakerMetrics
}

// GetRateLimiterMetrics returns rate limiter metrics for direct access
func (mi *MetricsIntegration) GetRateLimiterMetrics() *RateLimiterMetrics {
	return mi.rateLimiterMetrics
}

// GetBusinessMetrics returns business metrics for direct access
func (mi *MetricsIntegration) GetBusinessMetrics() *BusinessMetrics {
	return mi.businessMetrics
}
