package observability

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	traceapi "go.opentelemetry.io/otel/trace"
)

// TelemetryConfig holds configuration for OpenTelemetry
type TelemetryConfig struct {
	// ServiceName is the name of the service
	ServiceName string
	// ServiceVersion is the version of the service
	ServiceVersion string
	// Environment is the deployment environment (dev, staging, prod)
	Environment string
	// JaegerEndpoint is the Jaeger collector endpoint
	JaegerEndpoint string
	// OTLPEndpoint is the OTLP collector endpoint
	OTLPEndpoint string
	// EnableTracing enables distributed tracing
	EnableTracing bool
	// EnableMetrics enables metrics collection
	EnableMetrics bool
	// SamplingRatio is the trace sampling ratio (0.0 to 1.0)
	SamplingRatio float64
	// Headers for OTLP exporter
	Headers map[string]string
}

// DefaultTelemetryConfig returns default telemetry configuration
func DefaultTelemetryConfig() *TelemetryConfig {
	return &TelemetryConfig{
		ServiceName:    "kisanlink-ecom",
		ServiceVersion: "1.0.0",
		Environment:    getEnv("ENVIRONMENT", "development"),
		JaegerEndpoint: getEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
		OTLPEndpoint:   getEnv("OTLP_ENDPOINT", "http://localhost:4318"),
		EnableTracing:  true,
		EnableMetrics:  true,
		SamplingRatio:  0.1, // 10% sampling by default
		Headers:        make(map[string]string),
	}
}

// TelemetryManager manages OpenTelemetry setup and instrumentation
type TelemetryManager struct {
	config        *TelemetryConfig
	tracer        traceapi.Tracer
	meter         metric.Meter
	traceProvider *trace.TracerProvider
	shutdown      func(context.Context) error
}

// NewTelemetryManager creates a new telemetry manager
func NewTelemetryManager(config *TelemetryConfig) (*TelemetryManager, error) {
	if config == nil {
		config = DefaultTelemetryConfig()
	}

	tm := &TelemetryManager{
		config: config,
	}

	if err := tm.setupTracing(); err != nil {
		return nil, fmt.Errorf("failed to setup tracing: %w", err)
	}

	if err := tm.setupMetrics(); err != nil {
		return nil, fmt.Errorf("failed to setup metrics: %w", err)
	}

	return tm, nil
}

// setupTracing initializes distributed tracing
func (tm *TelemetryManager) setupTracing() error {
	if !tm.config.EnableTracing {
		return nil
	}

	// Create resource
	res, err := tm.createResource()
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace exporter
	exporter, err := tm.createTraceExporter()
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create trace provider
	tm.traceProvider = trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
		trace.WithSampler(trace.TraceIDRatioBased(tm.config.SamplingRatio)),
	)

	// Set global trace provider
	otel.SetTracerProvider(tm.traceProvider)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Create tracer
	tm.tracer = otel.Tracer(tm.config.ServiceName)

	// Set shutdown function
	tm.shutdown = tm.traceProvider.Shutdown

	return nil
}

// setupMetrics initializes metrics collection
func (tm *TelemetryManager) setupMetrics() error {
	if !tm.config.EnableMetrics {
		return nil
	}

	// Create meter
	tm.meter = otel.Meter(tm.config.ServiceName)

	return nil
}

// createResource creates an OpenTelemetry resource
func (tm *TelemetryManager) createResource() (*resource.Resource, error) {
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(tm.config.ServiceName),
			semconv.ServiceVersion(tm.config.ServiceVersion),
			semconv.DeploymentEnvironment(tm.config.Environment),
			attribute.String("service.instance.id", getHostname()),
		),
	)
}

// createTraceExporter creates a trace exporter based on configuration
func (tm *TelemetryManager) createTraceExporter() (trace.SpanExporter, error) {
	// Try OTLP first if configured
	if tm.config.OTLPEndpoint != "" {
		return tm.createOTLPExporter()
	}

	// Fall back to Jaeger if configured
	if tm.config.JaegerEndpoint != "" {
		return tm.createJaegerExporter()
	}

	// Default to stdout exporter for development
	return stdouttrace.New(stdouttrace.WithPrettyPrint())
}

// createOTLPExporter creates an OTLP trace exporter
func (tm *TelemetryManager) createOTLPExporter() (trace.SpanExporter, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(tm.config.OTLPEndpoint),
	}

	if len(tm.config.Headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(tm.config.Headers))
	}

	return otlptrace.New(
		context.Background(),
		otlptracehttp.NewClient(opts...),
	)
}

// createJaegerExporter creates a Jaeger trace exporter
func (tm *TelemetryManager) createJaegerExporter() (trace.SpanExporter, error) {
	return jaeger.New(jaeger.WithCollectorEndpoint(
		jaeger.WithEndpoint(tm.config.JaegerEndpoint),
	))
}

// GetTracer returns the tracer instance
func (tm *TelemetryManager) GetTracer() traceapi.Tracer {
	return tm.tracer
}

// GetMeter returns the meter instance
func (tm *TelemetryManager) GetMeter() metric.Meter {
	return tm.meter
}

// Shutdown gracefully shuts down telemetry
func (tm *TelemetryManager) Shutdown(ctx context.Context) error {
	if tm.shutdown != nil {
		return tm.shutdown(ctx)
	}
	return nil
}

// TracingMiddleware returns Gin middleware for tracing
func (tm *TelemetryManager) TracingMiddleware() gin.HandlerFunc {
	if !tm.config.EnableTracing {
		return gin.HandlerFunc(func(c *gin.Context) {
			c.Next()
		})
	}

	return otelgin.Middleware(tm.config.ServiceName)
}

// CustomTracingMiddleware provides custom tracing with additional attributes
func (tm *TelemetryManager) CustomTracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !tm.config.EnableTracing {
			c.Next()
			return
		}

		ctx := c.Request.Context()
		spanName := fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())

		ctx, span := tm.tracer.Start(ctx, spanName,
			traceapi.WithAttributes(
				semconv.HTTPMethod(c.Request.Method),
				semconv.HTTPRoute(c.FullPath()),
				semconv.HTTPTarget(c.Request.URL.Path),
				semconv.HTTPScheme(c.Request.URL.Scheme),
				semconv.HTTPUserAgent(c.Request.UserAgent()),
				semconv.HTTPClientIP(c.ClientIP()),
			),
		)

		// Add custom attributes
		if userID, exists := c.Get("user_id"); exists {
			span.SetAttributes(attribute.String("user.id", userID.(string)))
		}

		if orgID, exists := c.Get("organization_id"); exists {
			span.SetAttributes(attribute.String("organization.id", orgID.(string)))
		}

		// Store span in context
		c.Request = c.Request.WithContext(ctx)

		defer func() {
			// Set response attributes
			span.SetAttributes(
				semconv.HTTPStatusCode(c.Writer.Status()),
			)

			// Set status based on response
			if c.Writer.Status() >= 400 {
				span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", c.Writer.Status()))
			}

			span.End()
		}()

		c.Next()
	}
}

// InstrumentationHelper provides utilities for manual instrumentation
type InstrumentationHelper struct {
	tracer traceapi.Tracer
	meter  metric.Meter
}

// NewInstrumentationHelper creates a new instrumentation helper
func NewInstrumentationHelper(tm *TelemetryManager) *InstrumentationHelper {
	return &InstrumentationHelper{
		tracer: tm.GetTracer(),
		meter:  tm.GetMeter(),
	}
}

// StartSpan starts a new span with the given name and attributes
func (ih *InstrumentationHelper) StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, traceapi.Span) {
	return ih.tracer.Start(ctx, name, traceapi.WithAttributes(attrs...))
}

// AddEvent adds an event to the current span
func (ih *InstrumentationHelper) AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := traceapi.SpanFromContext(ctx)
	span.AddEvent(name, traceapi.WithAttributes(attrs...))
}

// SetAttribute sets an attribute on the current span
func (ih *InstrumentationHelper) SetAttribute(ctx context.Context, key string, value interface{}) {
	span := traceapi.SpanFromContext(ctx)
	switch v := value.(type) {
	case string:
		span.SetAttributes(attribute.String(key, v))
	case int:
		span.SetAttributes(attribute.Int(key, v))
	case int64:
		span.SetAttributes(attribute.Int64(key, v))
	case float64:
		span.SetAttributes(attribute.Float64(key, v))
	case bool:
		span.SetAttributes(attribute.Bool(key, v))
	}
}

// RecordError records an error on the current span
func (ih *InstrumentationHelper) RecordError(ctx context.Context, err error, attrs ...attribute.KeyValue) {
	span := traceapi.SpanFromContext(ctx)
	span.RecordError(err, traceapi.WithAttributes(attrs...))
	span.SetStatus(codes.Error, err.Error())
}

// MetricsCollector collects custom metrics
type MetricsCollector struct {
	meter               metric.Meter
	requestCounter      metric.Int64Counter
	requestDuration     metric.Float64Histogram
	errorCounter        metric.Int64Counter
	cacheHitCounter     metric.Int64Counter
	cacheMissCounter    metric.Int64Counter
	dbQueryDuration     metric.Float64Histogram
	circuitBreakerState metric.Int64UpDownCounter
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(tm *TelemetryManager) (*MetricsCollector, error) {
	meter := tm.GetMeter()

	requestCounter, err := meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		return nil, err
	}

	requestDuration, err := meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
	)
	if err != nil {
		return nil, err
	}

	errorCounter, err := meter.Int64Counter(
		"errors_total",
		metric.WithDescription("Total number of errors"),
	)
	if err != nil {
		return nil, err
	}

	cacheHitCounter, err := meter.Int64Counter(
		"cache_hits_total",
		metric.WithDescription("Total number of cache hits"),
	)
	if err != nil {
		return nil, err
	}

	cacheMissCounter, err := meter.Int64Counter(
		"cache_misses_total",
		metric.WithDescription("Total number of cache misses"),
	)
	if err != nil {
		return nil, err
	}

	dbQueryDuration, err := meter.Float64Histogram(
		"db_query_duration_seconds",
		metric.WithDescription("Database query duration in seconds"),
	)
	if err != nil {
		return nil, err
	}

	circuitBreakerState, err := meter.Int64UpDownCounter(
		"circuit_breaker_state",
		metric.WithDescription("Circuit breaker state (0=closed, 1=open, 2=half-open)"),
	)
	if err != nil {
		return nil, err
	}

	return &MetricsCollector{
		meter:               meter,
		requestCounter:      requestCounter,
		requestDuration:     requestDuration,
		errorCounter:        errorCounter,
		cacheHitCounter:     cacheHitCounter,
		cacheMissCounter:    cacheMissCounter,
		dbQueryDuration:     dbQueryDuration,
		circuitBreakerState: circuitBreakerState,
	}, nil
}

// RecordHTTPRequest records HTTP request metrics
func (mc *MetricsCollector) RecordHTTPRequest(ctx context.Context, method, route string, statusCode int, duration time.Duration) {
	attrs := []attribute.KeyValue{
		attribute.String("method", method),
		attribute.String("route", route),
		attribute.Int("status_code", statusCode),
	}

	mc.requestCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	mc.requestDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}

// RecordError records error metrics
func (mc *MetricsCollector) RecordError(ctx context.Context, errorType, component string) {
	attrs := []attribute.KeyValue{
		attribute.String("error_type", errorType),
		attribute.String("component", component),
	}

	mc.errorCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordCacheHit records cache hit metrics
func (mc *MetricsCollector) RecordCacheHit(ctx context.Context, cacheLevel string) {
	attrs := []attribute.KeyValue{
		attribute.String("cache_level", cacheLevel),
	}

	mc.cacheHitCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordCacheMiss records cache miss metrics
func (mc *MetricsCollector) RecordCacheMiss(ctx context.Context, cacheLevel string) {
	attrs := []attribute.KeyValue{
		attribute.String("cache_level", cacheLevel),
	}

	mc.cacheMissCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordDBQuery records database query metrics
func (mc *MetricsCollector) RecordDBQuery(ctx context.Context, operation string, duration time.Duration) {
	attrs := []attribute.KeyValue{
		attribute.String("operation", operation),
	}

	mc.dbQueryDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}

// RecordCircuitBreakerState records circuit breaker state changes
func (mc *MetricsCollector) RecordCircuitBreakerState(ctx context.Context, name string, state int64) {
	attrs := []attribute.KeyValue{
		attribute.String("circuit_breaker", name),
	}

	mc.circuitBreakerState.Add(ctx, state, metric.WithAttributes(attrs...))
}

// MetricsMiddleware provides metrics collection for HTTP requests
func (mc *MetricsCollector) MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		mc.RecordHTTPRequest(
			c.Request.Context(),
			c.Request.Method,
			c.FullPath(),
			c.Writer.Status(),
			duration,
		)

		// Record errors for 4xx and 5xx responses
		if c.Writer.Status() >= 400 {
			errorType := "client_error"
			if c.Writer.Status() >= 500 {
				errorType = "server_error"
			}
			mc.RecordError(c.Request.Context(), errorType, "http")
		}
	}
}

// Global telemetry manager instance
var globalTelemetryManager *TelemetryManager
var globalMetricsCollector *MetricsCollector
var globalInstrumentationHelper *InstrumentationHelper

// InitTelemetry initializes global telemetry
func InitTelemetry(config *TelemetryConfig) error {
	var err error

	globalTelemetryManager, err = NewTelemetryManager(config)
	if err != nil {
		return err
	}

	globalMetricsCollector, err = NewMetricsCollector(globalTelemetryManager)
	if err != nil {
		return err
	}

	globalInstrumentationHelper = NewInstrumentationHelper(globalTelemetryManager)

	return nil
}

// GetTelemetryManager returns the global telemetry manager
func GetTelemetryManager() *TelemetryManager {
	return globalTelemetryManager
}

// GetMetricsCollector returns the global metrics collector
func GetMetricsCollector() *MetricsCollector {
	return globalMetricsCollector
}

// GetInstrumentationHelper returns the global instrumentation helper
func GetInstrumentationHelper() *InstrumentationHelper {
	return globalInstrumentationHelper
}

// ShutdownTelemetry gracefully shuts down telemetry
func ShutdownTelemetry(ctx context.Context) error {
	if globalTelemetryManager != nil {
		return globalTelemetryManager.Shutdown(ctx)
	}
	return nil
}

// Utility functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}
