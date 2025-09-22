package marketplace

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// PerformanceMonitor tracks and reports on marketplace performance metrics
type PerformanceMonitor interface {
	// Metric recording
	RecordBidPlacement(ctx context.Context, duration time.Duration, success bool)
	RecordQueryExecution(ctx context.Context, queryType string, duration time.Duration, cacheHit bool)
	RecordConcurrentBids(ctx context.Context, listingID string, concurrentCount int)

	// Performance reporting
	GetPerformanceReport() PerformanceReport
	GetRealTimeMetrics() RealTimeMetrics

	// Alerting
	CheckPerformanceThresholds() []PerformanceAlert
	SetThresholds(thresholds PerformanceThresholds)

	// Reset metrics
	ResetMetrics()
}

// PerformanceReport provides comprehensive performance analysis
type PerformanceReport struct {
	ReportPeriod time.Duration
	GeneratedAt  time.Time

	// Bid placement metrics
	TotalBidsPlaced int64
	SuccessfulBids  int64
	FailedBids      int64
	SuccessRate     float64
	AverageBidTime  time.Duration
	P95BidTime      time.Duration
	P99BidTime      time.Duration

	// Query performance
	TotalQueries     int64
	CacheHitRate     float64
	AverageQueryTime time.Duration
	QueryTimesByType map[string]time.Duration

	// Concurrency metrics
	MaxConcurrentBids    int
	AverageConcurrency   float64
	ConcurrencyByListing map[string]int

	// System health
	ErrorRate           float64
	ThroughputPerSecond float64
	ResourceUtilization ResourceUtilization
}

// RealTimeMetrics provides current system performance
type RealTimeMetrics struct {
	CurrentTime         time.Time
	ActiveBidOperations int
	QueuedOperations    int
	CacheHitRate        float64
	AverageResponseTime time.Duration
	ErrorsPerMinute     int
	ThroughputPerMinute int
}

// PerformanceAlert represents a performance threshold violation
type PerformanceAlert struct {
	AlertType   AlertType
	Severity    AlertSeverity
	Message     string
	Threshold   interface{}
	ActualValue interface{}
	Timestamp   time.Time
}

// AlertType defines the type of performance alert
type AlertType string

const (
	AlertTypeBidLatency    AlertType = "BID_LATENCY"
	AlertTypeErrorRate     AlertType = "ERROR_RATE"
	AlertTypeCacheHitRate  AlertType = "CACHE_HIT_RATE"
	AlertTypeConcurrency   AlertType = "CONCURRENCY"
	AlertTypeThroughput    AlertType = "THROUGHPUT"
	AlertTypeResourceUsage AlertType = "RESOURCE_USAGE"
)

// AlertSeverity defines the severity level of alerts
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "INFO"
	AlertSeverityWarning  AlertSeverity = "WARNING"
	AlertSeverityCritical AlertSeverity = "CRITICAL"
)

// PerformanceThresholds defines performance alert thresholds
type PerformanceThresholds struct {
	MaxBidLatency     time.Duration
	MinCacheHitRate   float64
	MaxErrorRate      float64
	MaxConcurrentBids int
	MinThroughput     float64
	MaxMemoryUsage    int64
}

// ResourceUtilization tracks system resource usage
type ResourceUtilization struct {
	CPUUsage    float64
	MemoryUsage int64
	DiskUsage   int64
	NetworkIO   int64
}

// MetricEntry represents a single performance metric
type MetricEntry struct {
	Timestamp  time.Time
	MetricType string
	Value      float64
	Success    bool
	CacheHit   bool
	Duration   time.Duration
	Metadata   map[string]interface{}
}

// performanceMonitor implements PerformanceMonitor
type performanceMonitor struct {
	metrics    []MetricEntry
	metricsMux sync.RWMutex

	thresholds PerformanceThresholds
	threshMux  sync.RWMutex

	// Real-time counters
	activeBids     int64
	queuedOps      int64
	recentErrors   []time.Time
	recentRequests []time.Time
	counterMux     sync.RWMutex

	startTime time.Time
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor() PerformanceMonitor {
	monitor := &performanceMonitor{
		metrics:        make([]MetricEntry, 0, 10000),
		recentErrors:   make([]time.Time, 0, 1000),
		recentRequests: make([]time.Time, 0, 1000),
		startTime:      time.Now(),
		thresholds: PerformanceThresholds{
			MaxBidLatency:     500 * time.Millisecond,
			MinCacheHitRate:   0.8,
			MaxErrorRate:      0.05,
			MaxConcurrentBids: 100,
			MinThroughput:     10.0,
			MaxMemoryUsage:    1024 * 1024 * 1024, // 1GB
		},
	}

	// Start cleanup goroutine
	go monitor.cleanupOldMetrics()

	return monitor
}

// RecordBidPlacement records a bid placement operation
func (p *performanceMonitor) RecordBidPlacement(ctx context.Context, duration time.Duration, success bool) {
	entry := MetricEntry{
		Timestamp:  time.Now(),
		MetricType: "bid_placement",
		Value:      float64(duration.Nanoseconds()),
		Success:    success,
		Duration:   duration,
		Metadata: map[string]interface{}{
			"operation": "place_bid",
		},
	}

	p.addMetric(entry)

	// Update real-time counters
	p.counterMux.Lock()
	p.recentRequests = append(p.recentRequests, time.Now())
	if !success {
		p.recentErrors = append(p.recentErrors, time.Now())
	}
	p.counterMux.Unlock()
}

// RecordQueryExecution records a database query execution
func (p *performanceMonitor) RecordQueryExecution(ctx context.Context, queryType string, duration time.Duration, cacheHit bool) {
	entry := MetricEntry{
		Timestamp:  time.Now(),
		MetricType: "query_execution",
		Value:      float64(duration.Nanoseconds()),
		Success:    true,
		CacheHit:   cacheHit,
		Duration:   duration,
		Metadata: map[string]interface{}{
			"query_type": queryType,
		},
	}

	p.addMetric(entry)
}

// RecordConcurrentBids records concurrent bid activity
func (p *performanceMonitor) RecordConcurrentBids(ctx context.Context, listingID string, concurrentCount int) {
	entry := MetricEntry{
		Timestamp:  time.Now(),
		MetricType: "concurrent_bids",
		Value:      float64(concurrentCount),
		Success:    true,
		Metadata: map[string]interface{}{
			"listing_id": listingID,
		},
	}

	p.addMetric(entry)
}

// GetPerformanceReport generates a comprehensive performance report
func (p *performanceMonitor) GetPerformanceReport() PerformanceReport {
	p.metricsMux.RLock()
	defer p.metricsMux.RUnlock()

	now := time.Now()
	reportPeriod := now.Sub(p.startTime)

	report := PerformanceReport{
		ReportPeriod:         reportPeriod,
		GeneratedAt:          now,
		QueryTimesByType:     make(map[string]time.Duration),
		ConcurrencyByListing: make(map[string]int),
	}

	// Analyze bid placement metrics
	bidMetrics := p.filterMetrics("bid_placement")
	report.TotalBidsPlaced = int64(len(bidMetrics))

	var successfulBids int64
	var totalBidTime time.Duration
	bidTimes := make([]time.Duration, 0, len(bidMetrics))

	for _, metric := range bidMetrics {
		if metric.Success {
			successfulBids++
		}
		totalBidTime += metric.Duration
		bidTimes = append(bidTimes, metric.Duration)
	}

	report.SuccessfulBids = successfulBids
	report.FailedBids = report.TotalBidsPlaced - successfulBids

	if report.TotalBidsPlaced > 0 {
		report.SuccessRate = float64(successfulBids) / float64(report.TotalBidsPlaced)
		report.AverageBidTime = totalBidTime / time.Duration(report.TotalBidsPlaced)

		// Calculate percentiles
		if len(bidTimes) > 0 {
			report.P95BidTime = p.calculatePercentile(bidTimes, 0.95)
			report.P99BidTime = p.calculatePercentile(bidTimes, 0.99)
		}
	}

	// Analyze query metrics
	queryMetrics := p.filterMetrics("query_execution")
	report.TotalQueries = int64(len(queryMetrics))

	var cacheHits int64
	var totalQueryTime time.Duration
	queryTypeMap := make(map[string][]time.Duration)

	for _, metric := range queryMetrics {
		if metric.CacheHit {
			cacheHits++
		}
		totalQueryTime += metric.Duration

		if queryType, ok := metric.Metadata["query_type"].(string); ok {
			if _, exists := queryTypeMap[queryType]; !exists {
				queryTypeMap[queryType] = make([]time.Duration, 0)
			}
			queryTypeMap[queryType] = append(queryTypeMap[queryType], metric.Duration)
		}
	}

	if report.TotalQueries > 0 {
		report.CacheHitRate = float64(cacheHits) / float64(report.TotalQueries)
		report.AverageQueryTime = totalQueryTime / time.Duration(report.TotalQueries)
	}

	// Calculate average query time by type
	for queryType, times := range queryTypeMap {
		var total time.Duration
		for _, t := range times {
			total += t
		}
		report.QueryTimesByType[queryType] = total / time.Duration(len(times))
	}

	// Analyze concurrency metrics
	concurrencyMetrics := p.filterMetrics("concurrent_bids")
	var maxConcurrency int
	var totalConcurrency float64

	for _, metric := range concurrencyMetrics {
		concurrency := int(metric.Value)
		if concurrency > maxConcurrency {
			maxConcurrency = concurrency
		}
		totalConcurrency += metric.Value

		if listingID, ok := metric.Metadata["listing_id"].(string); ok {
			if existing, exists := report.ConcurrencyByListing[listingID]; !exists || concurrency > existing {
				report.ConcurrencyByListing[listingID] = concurrency
			}
		}
	}

	report.MaxConcurrentBids = maxConcurrency
	if len(concurrencyMetrics) > 0 {
		report.AverageConcurrency = totalConcurrency / float64(len(concurrencyMetrics))
	}

	// Calculate error rate and throughput
	if report.TotalBidsPlaced > 0 {
		report.ErrorRate = float64(report.FailedBids) / float64(report.TotalBidsPlaced)
		report.ThroughputPerSecond = float64(report.TotalBidsPlaced) / reportPeriod.Seconds()
	}

	return report
}

// GetRealTimeMetrics returns current real-time performance metrics
func (p *performanceMonitor) GetRealTimeMetrics() RealTimeMetrics {
	p.counterMux.RLock()
	defer p.counterMux.RUnlock()

	now := time.Now()
	oneMinuteAgo := now.Add(-1 * time.Minute)

	// Count recent errors and requests
	var recentErrorCount, recentRequestCount int

	for _, errorTime := range p.recentErrors {
		if errorTime.After(oneMinuteAgo) {
			recentErrorCount++
		}
	}

	for _, requestTime := range p.recentRequests {
		if requestTime.After(oneMinuteAgo) {
			recentRequestCount++
		}
	}

	// Calculate cache hit rate from recent metrics
	recentMetrics := p.getRecentMetrics(1 * time.Minute)
	var cacheHits, totalQueries int
	var totalResponseTime time.Duration

	for _, metric := range recentMetrics {
		if metric.MetricType == "query_execution" {
			totalQueries++
			if metric.CacheHit {
				cacheHits++
			}
			totalResponseTime += metric.Duration
		}
	}

	var cacheHitRate float64
	var avgResponseTime time.Duration

	if totalQueries > 0 {
		cacheHitRate = float64(cacheHits) / float64(totalQueries)
		avgResponseTime = totalResponseTime / time.Duration(totalQueries)
	}

	return RealTimeMetrics{
		CurrentTime:         now,
		ActiveBidOperations: int(p.activeBids),
		QueuedOperations:    int(p.queuedOps),
		CacheHitRate:        cacheHitRate,
		AverageResponseTime: avgResponseTime,
		ErrorsPerMinute:     recentErrorCount,
		ThroughputPerMinute: recentRequestCount,
	}
}

// CheckPerformanceThresholds checks for performance threshold violations
func (p *performanceMonitor) CheckPerformanceThresholds() []PerformanceAlert {
	p.threshMux.RLock()
	thresholds := p.thresholds
	p.threshMux.RUnlock()

	var alerts []PerformanceAlert
	realTimeMetrics := p.GetRealTimeMetrics()

	// Check bid latency
	if realTimeMetrics.AverageResponseTime > thresholds.MaxBidLatency {
		alerts = append(alerts, PerformanceAlert{
			AlertType:   AlertTypeBidLatency,
			Severity:    AlertSeverityWarning,
			Message:     fmt.Sprintf("Average response time (%v) exceeds threshold (%v)", realTimeMetrics.AverageResponseTime, thresholds.MaxBidLatency),
			Threshold:   thresholds.MaxBidLatency,
			ActualValue: realTimeMetrics.AverageResponseTime,
			Timestamp:   time.Now(),
		})
	}

	// Check cache hit rate
	if realTimeMetrics.CacheHitRate < thresholds.MinCacheHitRate {
		alerts = append(alerts, PerformanceAlert{
			AlertType:   AlertTypeCacheHitRate,
			Severity:    AlertSeverityWarning,
			Message:     fmt.Sprintf("Cache hit rate (%.2f) below threshold (%.2f)", realTimeMetrics.CacheHitRate, thresholds.MinCacheHitRate),
			Threshold:   thresholds.MinCacheHitRate,
			ActualValue: realTimeMetrics.CacheHitRate,
			Timestamp:   time.Now(),
		})
	}

	// Check error rate
	errorRate := float64(realTimeMetrics.ErrorsPerMinute) / float64(realTimeMetrics.ThroughputPerMinute)
	if errorRate > thresholds.MaxErrorRate {
		alerts = append(alerts, PerformanceAlert{
			AlertType:   AlertTypeErrorRate,
			Severity:    AlertSeverityCritical,
			Message:     fmt.Sprintf("Error rate (%.2f) exceeds threshold (%.2f)", errorRate, thresholds.MaxErrorRate),
			Threshold:   thresholds.MaxErrorRate,
			ActualValue: errorRate,
			Timestamp:   time.Now(),
		})
	}

	return alerts
}

// SetThresholds updates performance alert thresholds
func (p *performanceMonitor) SetThresholds(thresholds PerformanceThresholds) {
	p.threshMux.Lock()
	defer p.threshMux.Unlock()
	p.thresholds = thresholds
}

// ResetMetrics clears all collected metrics
func (p *performanceMonitor) ResetMetrics() {
	p.metricsMux.Lock()
	defer p.metricsMux.Unlock()

	p.metrics = make([]MetricEntry, 0, 10000)
	p.startTime = time.Now()

	p.counterMux.Lock()
	p.recentErrors = make([]time.Time, 0, 1000)
	p.recentRequests = make([]time.Time, 0, 1000)
	p.counterMux.Unlock()
}

// Helper methods

// addMetric adds a metric entry to the collection
func (p *performanceMonitor) addMetric(entry MetricEntry) {
	p.metricsMux.Lock()
	defer p.metricsMux.Unlock()

	p.metrics = append(p.metrics, entry)

	// Keep only the last 10000 metrics
	if len(p.metrics) > 10000 {
		p.metrics = p.metrics[1:]
	}
}

// filterMetrics returns metrics of a specific type
func (p *performanceMonitor) filterMetrics(metricType string) []MetricEntry {
	var filtered []MetricEntry

	for _, metric := range p.metrics {
		if metric.MetricType == metricType {
			filtered = append(filtered, metric)
		}
	}

	return filtered
}

// getRecentMetrics returns metrics from the last specified duration
func (p *performanceMonitor) getRecentMetrics(duration time.Duration) []MetricEntry {
	cutoff := time.Now().Add(-duration)
	var recent []MetricEntry

	for _, metric := range p.metrics {
		if metric.Timestamp.After(cutoff) {
			recent = append(recent, metric)
		}
	}

	return recent
}

// calculatePercentile calculates the specified percentile of durations
func (p *performanceMonitor) calculatePercentile(durations []time.Duration, percentile float64) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	// Simple percentile calculation (in production, use a proper sorting algorithm)
	index := int(float64(len(durations)) * percentile)
	if index >= len(durations) {
		index = len(durations) - 1
	}

	return durations[index]
}

// cleanupOldMetrics periodically removes old metrics to prevent memory leaks
func (p *performanceMonitor) cleanupOldMetrics() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cutoff := time.Now().Add(-1 * time.Hour)

		p.metricsMux.Lock()

		var kept []MetricEntry
		for _, metric := range p.metrics {
			if metric.Timestamp.After(cutoff) {
				kept = append(kept, metric)
			}
		}

		p.metrics = kept
		p.metricsMux.Unlock()

		// Cleanup recent error/request tracking
		p.counterMux.Lock()

		var recentErrors []time.Time
		for _, errorTime := range p.recentErrors {
			if errorTime.After(cutoff) {
				recentErrors = append(recentErrors, errorTime)
			}
		}
		p.recentErrors = recentErrors

		var recentRequests []time.Time
		for _, requestTime := range p.recentRequests {
			if requestTime.After(cutoff) {
				recentRequests = append(recentRequests, requestTime)
			}
		}
		p.recentRequests = recentRequests

		p.counterMux.Unlock()
	}
}
