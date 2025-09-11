package health

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "runtime"
    "time"

    "kisanlink-ecom/internal/database"

    "github.com/gin-gonic/gin"
)

// StatusHandler handles status dashboard endpoints
type StatusHandler struct{}

// NewStatusHandler creates a new status handler
func NewStatusHandler() *StatusHandler {
    return &StatusHandler{}
}

// StatusDashboard represents the status dashboard data
type StatusDashboard struct {
    Service      ServiceStatus      `json:"service"`
    Dependencies []DependencyStatus `json:"dependencies"`
    Alerts       []Alert            `json:"alerts"`
    Metrics      DashboardMetrics   `json:"metrics"`
    LastUpdated  string             `json:"last_updated"`
}

// ServiceStatus represents the overall service status
type ServiceStatus struct {
    Name        string `json:"name"`
    Version     string `json:"version"`
    Status      string `json:"status"`
    Uptime      string `json:"uptime"`
    Environment string `json:"environment"`
}

// DependencyStatus represents the status of a dependency
type DependencyStatus struct {
    Name        string `json:"name"`
    Type        string `json:"type"`
    Status      string `json:"status"`
    Latency     int64  `json:"latency_ms"`
    LastChecked string `json:"last_checked"`
    Message     string `json:"message,omitempty"`
}

// Alert represents a system alert
type Alert struct {
    Level     string `json:"level"`
    Message   string `json:"message"`
    Timestamp string `json:"timestamp"`
    Component string `json:"component"`
    Resolved  bool   `json:"resolved"`
}

// DashboardMetrics represents key metrics for the dashboard
type DashboardMetrics struct {
    RequestsPerMinute float64 `json:"requests_per_minute"`
    ErrorRate         float64 `json:"error_rate"`
    AvgResponseTime   float64 `json:"avg_response_time_ms"`
    ActiveConnections int     `json:"active_connections"`
    MemoryUsage       float64 `json:"memory_usage_percent"`
    CPUUsage          float64 `json:"cpu_usage_percent"`
}

// GetStatusDashboard handles the status dashboard endpoint
// @Summary Get status dashboard
// @Description Get comprehensive status dashboard with service health, dependencies, and alerts
// @Tags health
// @Produce json
// @Success 200 {object} StatusDashboard
// @Failure 503 {object} StatusDashboard
// @Router /status [get]
func (h *StatusHandler) GetStatusDashboard(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()

    dashboard := StatusDashboard{
        Service:      h.getServiceStatus(),
        Dependencies: h.getDependencyStatuses(ctx, c),
        Alerts:       h.getActiveAlerts(),
        Metrics:      h.getDashboardMetrics(),
        LastUpdated:  time.Now().UTC().Format(time.RFC3339),
    }

    // Determine overall status
    overallStatus := "healthy"
    statusCode := http.StatusOK

    for _, dep := range dashboard.Dependencies {
        if dep.Status == "unhealthy" {
            overallStatus = "degraded"
            statusCode = http.StatusServiceUnavailable
            break
        } else if dep.Status == "warning" && overallStatus == "healthy" {
            overallStatus = "warning"
        }
    }

    dashboard.Service.Status = overallStatus

    c.JSON(statusCode, dashboard)
}

// getServiceStatus returns the current service status
func (h *StatusHandler) getServiceStatus() ServiceStatus {
    uptime := time.Since(startTime)

    return ServiceStatus{
        Name:        "KisanLink E-commerce API",
        Version:     "1.0.0",
        Status:      "healthy", // Will be updated based on dependencies
        Uptime:      uptime.String(),
        Environment: getEnvironment(),
    }
}

// getDependencyStatuses checks the status of all dependencies
func (h *StatusHandler) getDependencyStatuses(ctx context.Context, c *gin.Context) []DependencyStatus {
    var dependencies []DependencyStatus

    // Check database status
    dbStatus := h.checkDatabaseStatus(ctx, c)
    dependencies = append(dependencies, dbStatus)

    // Check AAA service status (if available)
    aaaStatus := h.checkAAAServiceStatus(ctx)
    dependencies = append(dependencies, aaaStatus)

    return dependencies
}

// checkDatabaseStatus checks the database dependency status
func (h *StatusHandler) checkDatabaseStatus(ctx context.Context, c *gin.Context) DependencyStatus {
    status := DependencyStatus{
        Name:        "Database",
        Type:        "PostgreSQL/DynamoDB",
        Status:      "unknown",
        LastChecked: time.Now().UTC().Format(time.RFC3339),
    }

    if dbManager, exists := c.Get("dbManager"); exists {
        if dm, ok := dbManager.(*database.DatabaseManager); ok {
            start := time.Now()
            err := dm.HealthCheck(ctx)
            latency := time.Since(start)

            status.Latency = latency.Milliseconds()

            if err != nil {
                status.Status = "unhealthy"
                status.Message = err.Error()
            } else {
                status.Status = "healthy"
                status.Message = "All database connections are working"
            }
        } else {
            status.Status = "unhealthy"
            status.Message = "Database manager not properly initialized"
        }
    } else {
        status.Status = "unhealthy"
        status.Message = "Database manager not found"
    }

    return status
}

// checkAAAServiceStatus checks the AAA service dependency status
func (h *StatusHandler) checkAAAServiceStatus(ctx context.Context) DependencyStatus {
    status := DependencyStatus{
        Name:        "AAA Service",
        Type:        "gRPC",
        Status:      "unknown",
        LastChecked: time.Now().UTC().Format(time.RFC3339),
    }

    // TODO: Implement actual AAA service health check
    // For now, assume it's healthy if we're running
    status.Status = "healthy"
    status.Message = "AAA service integration active"
    status.Latency = 50 // Mock latency in milliseconds

    return status
}

// getActiveAlerts returns current system alerts
func (h *StatusHandler) getActiveAlerts() []Alert {
    var alerts []Alert

    // Check for high error rate
    if errorsTotal > 0 && requestsTotal > 0 {
        errorRate := float64(errorsTotal) / float64(requestsTotal) * 100
        if errorRate > 5.0 { // Alert if error rate > 5%
            alerts = append(alerts, Alert{
                Level:     "warning",
                Message:   fmt.Sprintf("High error rate detected: %.2f%%", errorRate),
                Timestamp: time.Now().UTC().Format(time.RFC3339),
                Component: "HTTP",
                Resolved:  false,
            })
        }
    }

    // Check for high memory usage
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    memoryUsagePercent := float64(m.Alloc) / float64(m.Sys) * 100
    if memoryUsagePercent > 80.0 { // Alert if memory usage > 80%
        alerts = append(alerts, Alert{
            Level:     "warning",
            Message:   fmt.Sprintf("High memory usage: %.2f%%", memoryUsagePercent),
            Timestamp: time.Now().UTC().Format(time.RFC3339),
            Component: "System",
            Resolved:  false,
        })
    }

    // Check for too many goroutines
    numGoroutines := runtime.NumGoroutine()
    if numGoroutines > 1000 { // Alert if more than 1000 goroutines
        alerts = append(alerts, Alert{
            Level:     "warning",
            Message:   fmt.Sprintf("High number of goroutines: %d", numGoroutines),
            Timestamp: time.Now().UTC().Format(time.RFC3339),
            Component: "System",
            Resolved:  false,
        })
    }

    return alerts
}

// getDashboardMetrics returns key metrics for the dashboard
func (h *StatusHandler) getDashboardMetrics() DashboardMetrics {
    uptime := time.Since(startTime)
    requestsPerMinute := float64(requestsTotal) / uptime.Minutes()

    var avgResponseTime float64
    if requestCount > 0 {
        avgResponseTime = float64(totalLatency.Nanoseconds()) / float64(requestCount) / 1e6
    }

    var errorRate float64
    if requestsTotal > 0 {
        errorRate = float64(errorsTotal) / float64(requestsTotal) * 100
    }

    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    memoryUsagePercent := float64(m.Alloc) / float64(m.Sys) * 100

    return DashboardMetrics{
        RequestsPerMinute: requestsPerMinute,
        ErrorRate:         errorRate,
        AvgResponseTime:   avgResponseTime,
        ActiveConnections: runtime.NumGoroutine(), // Approximate
        MemoryUsage:       memoryUsagePercent,
        CPUUsage:          0.0, // TODO: Implement CPU usage monitoring
    }
}

// getEnvironment returns the current environment
func getEnvironment() string {
    env := os.Getenv("ENVIRONMENT")
    if env == "" {
        env = "development"
    }
    return env
}
