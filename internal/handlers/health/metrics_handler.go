package health

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/Kisanlink/kisanlink-ecom/internal/database"

	"github.com/gin-gonic/gin"
)

// MetricsHandler handles metrics endpoints
type MetricsHandler struct{}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{}
}

// MetricsResponse represents system metrics
type MetricsResponse struct {
	System   SystemMetrics   `json:"system"`
	Database DatabaseMetrics `json:"database"`
	HTTP     HTTPMetrics     `json:"http"`
}

// SystemMetrics represents system-level metrics
type SystemMetrics struct {
	Uptime        string `json:"uptime"`
	GoVersion     string `json:"go_version"`
	NumGoroutines int    `json:"num_goroutines"`
	NumCPU        int    `json:"num_cpu"`
	MemoryUsage   Memory `json:"memory_usage"`
	GCStats       GC     `json:"gc_stats"`
}

// Memory represents memory usage metrics
type Memory struct {
	Alloc        uint64 `json:"alloc_bytes"`
	TotalAlloc   uint64 `json:"total_alloc_bytes"`
	Sys          uint64 `json:"sys_bytes"`
	NumGC        uint32 `json:"num_gc"`
	HeapAlloc    uint64 `json:"heap_alloc_bytes"`
	HeapSys      uint64 `json:"heap_sys_bytes"`
	HeapInuse    uint64 `json:"heap_inuse_bytes"`
	HeapReleased uint64 `json:"heap_released_bytes"`
}

// GC represents garbage collection metrics
type GC struct {
	NumGC      uint32 `json:"num_gc"`
	PauseTotal int64  `json:"pause_total_ns"`
	LastGC     string `json:"last_gc"`
	NextGC     uint64 `json:"next_gc_bytes"`
}

// DatabaseMetrics represents database-related metrics
type DatabaseMetrics struct {
	Status      string `json:"status"`
	Connections int    `json:"connections"`
	Latency     int64  `json:"latency_ms"`
	Errors      int64  `json:"errors"`
}

// HTTPMetrics represents HTTP-related metrics
type HTTPMetrics struct {
	RequestsTotal   int64   `json:"requests_total"`
	RequestsPerSec  float64 `json:"requests_per_sec"`
	AvgResponseTime float64 `json:"avg_response_time_ms"`
	ErrorRate       float64 `json:"error_rate"`
}

var (
	startTime     = time.Now()
	requestsTotal int64
	errorsTotal   int64
	totalLatency  time.Duration
	requestCount  int64
)

// GetMetrics handles the metrics endpoint
// @Summary Get system metrics
// @Description Get comprehensive system metrics including memory, database, and HTTP stats
// @Tags health
// @Produce json
// @Success 200 {object} MetricsResponse
// @Failure 503 {object} MetricsResponse
// @Router /metrics [get]
func (h *MetricsHandler) GetMetrics(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	metrics := MetricsResponse{
		System:   h.getSystemMetrics(),
		Database: h.getDatabaseMetrics(ctx, c),
		HTTP:     h.getHTTPMetrics(),
	}

	// Determine overall health status
	statusCode := http.StatusOK
	if metrics.Database.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, metrics)
}

// getSystemMetrics collects system-level metrics
func (h *MetricsHandler) getSystemMetrics() SystemMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	uptime := time.Since(startTime)

	return SystemMetrics{
		Uptime:        uptime.String(),
		GoVersion:     runtime.Version(),
		NumGoroutines: runtime.NumGoroutine(),
		NumCPU:        runtime.NumCPU(),
		MemoryUsage: Memory{
			Alloc:        m.Alloc,
			TotalAlloc:   m.TotalAlloc,
			Sys:          m.Sys,
			NumGC:        m.NumGC,
			HeapAlloc:    m.HeapAlloc,
			HeapSys:      m.HeapSys,
			HeapInuse:    m.HeapInuse,
			HeapReleased: m.HeapReleased,
		},
		GCStats: GC{
			NumGC:      m.NumGC,
			PauseTotal: int64(m.PauseTotalNs),
			LastGC:     time.Unix(0, int64(m.LastGC)).Format(time.RFC3339),
			NextGC:     m.NextGC,
		},
	}
}

// getDatabaseMetrics collects database-related metrics
func (h *MetricsHandler) getDatabaseMetrics(ctx context.Context, c *gin.Context) DatabaseMetrics {
	metrics := DatabaseMetrics{
		Status:      "unknown",
		Connections: 0,
		Latency:     0,
		Errors:      errorsTotal,
	}

	// Check if database manager is available
	if dbManager, exists := c.Get("dbManager"); exists {
		if dm, ok := dbManager.(*database.DatabaseManager); ok {
			start := time.Now()
			err := dm.HealthCheck(ctx)
			latency := time.Since(start)

			metrics.Latency = latency.Milliseconds()
			if err != nil {
				metrics.Status = "unhealthy"
			} else {
				metrics.Status = "healthy"
			}
		}
	}

	return metrics
}

// getHTTPMetrics collects HTTP-related metrics
func (h *MetricsHandler) getHTTPMetrics() HTTPMetrics {
	uptime := time.Since(startTime)
	requestsPerSec := float64(requestsTotal) / uptime.Seconds()

	var avgResponseTime float64
	if requestCount > 0 {
		avgResponseTime = float64(totalLatency.Nanoseconds()) / float64(requestCount) / 1e6 // Convert to milliseconds
	}

	var errorRate float64
	if requestsTotal > 0 {
		errorRate = float64(errorsTotal) / float64(requestsTotal) * 100
	}

	return HTTPMetrics{
		RequestsTotal:   requestsTotal,
		RequestsPerSec:  requestsPerSec,
		AvgResponseTime: avgResponseTime,
		ErrorRate:       errorRate,
	}
}

// PrometheusMetrics handles Prometheus-style metrics endpoint
// @Summary Get Prometheus metrics
// @Description Get metrics in Prometheus format for scraping
// @Tags health
// @Produce text/plain
// @Success 200 {string} string "Prometheus metrics"
// @Router /metrics/prometheus [get]
func (h *MetricsHandler) PrometheusMetrics(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	uptime := time.Since(startTime)

	// Generate Prometheus-style metrics
	metrics := `# HELP kisanlink_ecom_uptime_seconds Total uptime of the service in seconds
# TYPE kisanlink_ecom_uptime_seconds counter
kisanlink_ecom_uptime_seconds ` + formatFloat(uptime.Seconds()) + `

# HELP kisanlink_ecom_goroutines Number of goroutines
# TYPE kisanlink_ecom_goroutines gauge
kisanlink_ecom_goroutines ` + formatInt(runtime.NumGoroutine()) + `

# HELP kisanlink_ecom_memory_alloc_bytes Bytes allocated and still in use
# TYPE kisanlink_ecom_memory_alloc_bytes gauge
kisanlink_ecom_memory_alloc_bytes ` + formatUint64(m.Alloc) + `

# HELP kisanlink_ecom_memory_sys_bytes Bytes obtained from system
# TYPE kisanlink_ecom_memory_sys_bytes gauge
kisanlink_ecom_memory_sys_bytes ` + formatUint64(m.Sys) + `

# HELP kisanlink_ecom_gc_runs_total Total number of GC runs
# TYPE kisanlink_ecom_gc_runs_total counter
kisanlink_ecom_gc_runs_total ` + formatUint32(m.NumGC) + `

# HELP kisanlink_ecom_http_requests_total Total number of HTTP requests
# TYPE kisanlink_ecom_http_requests_total counter
kisanlink_ecom_http_requests_total ` + formatInt64(requestsTotal) + `

# HELP kisanlink_ecom_http_errors_total Total number of HTTP errors
# TYPE kisanlink_ecom_http_errors_total counter
kisanlink_ecom_http_errors_total ` + formatInt64(errorsTotal) + `
`

	// Add database metrics if available
	if dbManager, exists := c.Get("dbManager"); exists {
		if dm, ok := dbManager.(*database.DatabaseManager); ok {
			start := time.Now()
			err := dm.HealthCheck(ctx)
			latency := time.Since(start)

			dbStatus := 1.0
			if err != nil {
				dbStatus = 0.0
			}

			metrics += `
# HELP kisanlink_ecom_database_up Database availability (1 = up, 0 = down)
# TYPE kisanlink_ecom_database_up gauge
kisanlink_ecom_database_up ` + formatFloat(dbStatus) + `

# HELP kisanlink_ecom_database_latency_seconds Database query latency in seconds
# TYPE kisanlink_ecom_database_latency_seconds gauge
kisanlink_ecom_database_latency_seconds ` + formatFloat(latency.Seconds()) + `
`
		}
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, metrics)
}

// MetricsMiddleware tracks HTTP metrics
func (h *MetricsHandler) MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		// Update metrics
		requestsTotal++
		requestCount++
		totalLatency += time.Since(start)

		// Count errors (4xx and 5xx status codes)
		if c.Writer.Status() >= 400 {
			errorsTotal++
		}
	}
}

// Helper functions for formatting metrics
func formatFloat(f float64) string {
	return fmt.Sprintf("%.6f", f)
}

func formatInt(i int) string {
	return fmt.Sprintf("%d", i)
}

func formatInt64(i int64) string {
	return fmt.Sprintf("%d", i)
}

func formatUint32(i uint32) string {
	return fmt.Sprintf("%d", i)
}

func formatUint64(i uint64) string {
	return fmt.Sprintf("%d", i)
}
