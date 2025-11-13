package health

import (
	"context"
	"net/http"
	"time"

	"kisanlink-ecom/entities/models/common"
	"kisanlink-ecom/internal/database"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// DetailedHealthResponse represents a detailed health check response
type DetailedHealthResponse struct {
	Status       string                 `json:"status"`
	Message      string                 `json:"message"`
	Version      string                 `json:"version"`
	Timestamp    string                 `json:"timestamp"`
	Dependencies map[string]HealthCheck `json:"dependencies,omitempty"`
}

// HealthCheck represents the health status of a dependency
type HealthCheck struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

// HealthCheck handles the basic health check endpoint
// @Summary Health check
// @Description Check if the service is running
// @Tags health
// @Produce json
// @Success 200 {object} common.HealthResponse
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	response := common.HealthResponse{
		Status:  "ok",
		Message: "KisanLink E-commerce API is running",
		Version: "1.0.0",
	}

	c.JSON(http.StatusOK, response)
}

// DetailedHealthCheck handles the detailed health check endpoint with dependency checks
// @Summary Detailed health check
// @Description Check if the service and its dependencies are healthy
// @Tags health
// @Produce json
// @Success 200 {object} DetailedHealthResponse
// @Failure 503 {object} DetailedHealthResponse
// @Router /health/detailed [get]
func (h *HealthHandler) DetailedHealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	response := DetailedHealthResponse{
		Status:       "ok",
		Message:      "KisanLink E-commerce API is running",
		Version:      "1.0.0",
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Dependencies: make(map[string]HealthCheck),
	}

	overallHealthy := true

	// Check database health
	if dbManager, exists := c.Get("dbManager"); exists {
		if dm, ok := dbManager.(*database.DatabaseManager); ok {
			start := time.Now()
			err := dm.HealthCheck(ctx)
			latency := time.Since(start)

			if err != nil {
				response.Dependencies["database"] = HealthCheck{
					Status:  "unhealthy",
					Message: err.Error(),
					Latency: latency.String(),
				}
				overallHealthy = false
			} else {
				response.Dependencies["database"] = HealthCheck{
					Status:  "healthy",
					Message: "Database connections are working",
					Latency: latency.String(),
				}
			}
		} else {
			response.Dependencies["database"] = HealthCheck{
				Status:  "unknown",
				Message: "Database manager not available",
			}
			overallHealthy = false
		}
	} else {
		response.Dependencies["database"] = HealthCheck{
			Status:  "unknown",
			Message: "Database manager not found in context",
		}
		overallHealthy = false
	}

	// Set overall status
	if !overallHealthy {
		response.Status = "degraded"
		response.Message = "Some dependencies are unhealthy"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ReadinessCheck handles the readiness probe endpoint for Kubernetes
// @Summary Readiness check
// @Description Check if the service is ready to accept traffic
// @Tags health
// @Produce json
// @Success 200 {object} common.HealthResponse
// @Failure 503 {object} common.HealthResponse
// @Router /ready [get]
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	// Check if database is ready
	if dbManager, exists := c.Get("dbManager"); exists {
		if dm, ok := dbManager.(*database.DatabaseManager); ok {
			if err := dm.HealthCheck(ctx); err != nil {
				response := common.HealthResponse{
					Status:  "not ready",
					Message: "Database is not ready",
					Version: "1.0.0",
				}
				c.JSON(http.StatusServiceUnavailable, response)
				return
			}
		}
	}

	response := common.HealthResponse{
		Status:  "ready",
		Message: "Service is ready to accept traffic",
		Version: "1.0.0",
	}

	c.JSON(http.StatusOK, response)
}
