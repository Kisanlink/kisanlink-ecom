package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Kisanlink/kisanlink-ecom/internal/health"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	healthManager *health.HealthManager
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(healthManager *health.HealthManager) *HealthHandler {
	return &HealthHandler{
		healthManager: healthManager,
	}
}

// RegisterRoutes registers health check routes
func (h *HealthHandler) RegisterRoutes(router *gin.RouterGroup) {
	healthGroup := router.Group("/health")
	{
		healthGroup.GET("", h.GetHealth)
		healthGroup.GET("/", h.GetHealth)
		healthGroup.GET("/detailed", h.GetDetailedHealth)
		healthGroup.GET("/live", h.GetLiveness)
		healthGroup.GET("/ready", h.GetReadiness)
		healthGroup.GET("/components", h.GetComponentsHealth)
		healthGroup.GET("/components/:component", h.GetComponentHealth)
	}
}

// GetHealth returns basic health status
// @Summary Get basic health status
// @Description Returns basic health status of the system
// @Tags health
// @Produce json
// @Success 200 {object} health.HealthResult "System is healthy"
// @Success 503 {object} health.HealthResult "System is unhealthy"
// @Router /health [get]
func (h *HealthHandler) GetHealth(c *gin.Context) {
	systemHealth := h.healthManager.GetLastHealthCheck()
	result := systemHealth.IsHealthy()

	if result.Healthy {
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusServiceUnavailable, result)
	}
}

// GetDetailedHealth returns detailed health status
// @Summary Get detailed health status
// @Description Returns detailed health status of all system components
// @Tags health
// @Produce json
// @Success 200 {object} health.SystemHealth "Detailed system health"
// @Success 503 {object} health.SystemHealth "System is unhealthy"
// @Router /health/detailed [get]
func (h *HealthHandler) GetDetailedHealth(c *gin.Context) {
	// Check if we should force a fresh health check
	forceCheck := c.Query("force") == "true"

	var systemHealth health.SystemHealth
	if forceCheck {
		systemHealth = h.healthManager.CheckHealth(c.Request.Context())
	} else {
		systemHealth = h.healthManager.GetLastHealthCheck()
	}

	statusCode := http.StatusOK
	if systemHealth.Status == health.StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, systemHealth)
}

// GetLiveness returns liveness probe status
// @Summary Get liveness probe status
// @Description Returns liveness probe status for Kubernetes
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{} "Service is alive"
// @Success 503 {object} map[string]interface{} "Service is not alive"
// @Router /health/live [get]
func (h *HealthHandler) GetLiveness(c *gin.Context) {
	// Liveness probe - basic check that the service is running
	// This should only fail if the service is completely broken
	response := gin.H{
		"status":    "alive",
		"timestamp": time.Now().UTC(),
		"uptime":    time.Since(time.Now().Add(-time.Hour)).String(), // Placeholder uptime
	}

	c.JSON(http.StatusOK, response)
}

// GetReadiness returns readiness probe status
// @Summary Get readiness probe status
// @Description Returns readiness probe status for Kubernetes
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{} "Service is ready"
// @Success 503 {object} map[string]interface{} "Service is not ready"
// @Router /health/ready [get]
func (h *HealthHandler) GetReadiness(c *gin.Context) {
	// Readiness probe - check if service is ready to handle traffic
	systemHealth := h.healthManager.GetLastHealthCheck()

	ready := systemHealth.Status == health.StatusHealthy || systemHealth.Status == health.StatusDegraded

	response := gin.H{
		"status":    map[bool]string{true: "ready", false: "not_ready"}[ready],
		"timestamp": time.Now().UTC(),
	}

	if ready {
		c.JSON(http.StatusOK, response)
	} else {
		response["reason"] = "One or more critical components are unhealthy"
		c.JSON(http.StatusServiceUnavailable, response)
	}
}

// GetComponentsHealth returns health status of all components
// @Summary Get health status of all components
// @Description Returns health status of all registered components
// @Tags health
// @Produce json
// @Success 200 {object} map[string]health.ComponentHealth "Components health status"
// @Router /health/components [get]
func (h *HealthHandler) GetComponentsHealth(c *gin.Context) {
	systemHealth := h.healthManager.GetLastHealthCheck()

	// Add summary statistics
	response := gin.H{
		"components": systemHealth.Components,
		"summary": gin.H{
			"total":        len(systemHealth.Components),
			"healthy":      h.countComponentsByStatus(systemHealth.Components, health.StatusHealthy),
			"degraded":     h.countComponentsByStatus(systemHealth.Components, health.StatusDegraded),
			"unhealthy":    h.countComponentsByStatus(systemHealth.Components, health.StatusUnhealthy),
			"last_updated": systemHealth.Timestamp,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetComponentHealth returns health status of a specific component
// @Summary Get health status of a specific component
// @Description Returns health status of a specific component by name
// @Tags health
// @Param component path string true "Component name"
// @Produce json
// @Success 200 {object} health.ComponentHealth "Component health status"
// @Success 404 {object} map[string]string "Component not found"
// @Router /health/components/{component} [get]
func (h *HealthHandler) GetComponentHealth(c *gin.Context) {
	componentName := c.Param("component")

	systemHealth := h.healthManager.GetLastHealthCheck()

	if component, exists := systemHealth.Components[componentName]; exists {
		c.JSON(http.StatusOK, component)
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"error":     "Component not found",
			"component": componentName,
		})
	}
}

// countComponentsByStatus counts components by their health status
func (h *HealthHandler) countComponentsByStatus(components map[string]health.ComponentHealth, status health.HealthStatus) int {
	count := 0
	for _, component := range components {
		if component.Status == status {
			count++
		}
	}
	return count
}

// HealthMiddleware provides health check information in response headers
func HealthMiddleware(healthManager *health.HealthManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add health status to response headers
		systemHealth := healthManager.GetLastHealthCheck()

		c.Header("X-Health-Status", string(systemHealth.Status))
		c.Header("X-Health-Version", systemHealth.Version)
		c.Header("X-Health-Timestamp", systemHealth.Timestamp.Format(time.RFC3339))

		// Add component count headers
		c.Header("X-Health-Components-Total", strconv.Itoa(len(systemHealth.Components)))
		c.Header("X-Health-Components-Healthy", strconv.Itoa(countComponentsByStatus(systemHealth.Components, health.StatusHealthy)))
		c.Header("X-Health-Components-Unhealthy", strconv.Itoa(countComponentsByStatus(systemHealth.Components, health.StatusUnhealthy)))

		c.Next()
	}
}

// Helper function for counting components by status
func countComponentsByStatus(components map[string]health.ComponentHealth, status health.HealthStatus) int {
	count := 0
	for _, component := range components {
		if component.Status == status {
			count++
		}
	}
	return count
}
