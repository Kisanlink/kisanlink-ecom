package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"kisanlink-ecom/entities/models/common"
	"kisanlink-ecom/internal/resilience"

	"github.com/gin-gonic/gin"
)

// ResilienceMiddleware provides endpoints for monitoring and controlling resilience components
type ResilienceMiddleware struct {
	// No fields needed for now
}

// NewResilienceMiddleware creates a new resilience middleware
func NewResilienceMiddleware() *ResilienceMiddleware {
	return &ResilienceMiddleware{}
}

// GetCircuitBreakersStatus returns the status of all circuit breakers
func (m *ResilienceMiddleware) GetCircuitBreakersStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := resilience.GetCircuitBreakerStats()
		metrics := resilience.GetCircuitBreakerMetrics()

		response := gin.H{
			"summary":   stats,
			"details":   metrics,
			"timestamp": time.Now().UTC(),
		}

		c.JSON(http.StatusOK, common.APIResponse{
			Success: true,
			Data:    response,
		})
	}
}

// GetCircuitBreakerDetails returns detailed information about a specific circuit breaker
func (m *ResilienceMiddleware) GetCircuitBreakerDetails() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, common.APIResponse{
				Success: false,
				Error: &common.APIError{
					Code:    "MISSING_NAME",
					Message: "Circuit breaker name is required",
				},
			})
			return
		}

		cb, exists := resilience.GetCircuitBreaker(name), true
		if cb == nil {
			c.JSON(http.StatusNotFound, common.APIResponse{
				Success: false,
				Error: &common.APIError{
					Code:    "NOT_FOUND",
					Message: "Circuit breaker not found",
				},
			})
			return
		}

		if !exists {
			c.JSON(http.StatusNotFound, common.APIResponse{
				Success: false,
				Error: &common.APIError{
					Code:    "NOT_FOUND",
					Message: "Circuit breaker not found",
				},
			})
			return
		}

		info := resilience.CircuitBreakerInfo{
			Name:   cb.Name(),
			State:  cb.State(),
			Counts: cb.Counts(),
		}

		c.JSON(http.StatusOK, common.APIResponse{
			Success: true,
			Data:    info,
		})
	}
}

// ResetCircuitBreaker resets a specific circuit breaker to closed state
func (m *ResilienceMiddleware) ResetCircuitBreaker() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, common.APIResponse{
				Success: false,
				Error: &common.APIError{
					Code:    "MISSING_NAME",
					Message: "Circuit breaker name is required",
				},
			})
			return
		}

		err := resilience.ResetCircuitBreaker(name)
		if err != nil {
			c.JSON(http.StatusNotFound, common.APIResponse{
				Success: false,
				Error: &common.APIError{
					Code:    "RESET_FAILED",
					Message: err.Error(),
				},
			})
			return
		}

		c.JSON(http.StatusOK, common.APIResponse{
			Success: true,
			Data: gin.H{
				"message": "Circuit breaker reset successfully",
				"name":    name,
			},
		})
	}
}

// ForceOpenCircuitBreaker forces a specific circuit breaker to open state
func (m *ResilienceMiddleware) ForceOpenCircuitBreaker() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, common.APIResponse{
				Success: false,
				Error: &common.APIError{
					Code:    "MISSING_NAME",
					Message: "Circuit breaker name is required",
				},
			})
			return
		}

		err := resilience.ForceOpenCircuitBreaker(name)
		if err != nil {
			c.JSON(http.StatusNotFound, common.APIResponse{
				Success: false,
				Error: &common.APIError{
					Code:    "FORCE_OPEN_FAILED",
					Message: err.Error(),
				},
			})
			return
		}

		c.JSON(http.StatusOK, common.APIResponse{
			Success: true,
			Data: gin.H{
				"message": "Circuit breaker forced open successfully",
				"name":    name,
			},
		})
	}
}

// CircuitBreakerHealthCheck provides a health check endpoint for circuit breakers
func (m *ResilienceMiddleware) CircuitBreakerHealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		healthy, message := resilience.CircuitBreakerHealthCheck()

		status := http.StatusOK
		if !healthy {
			status = http.StatusServiceUnavailable
		}

		c.JSON(status, common.APIResponse{
			Success: healthy,
			Data: gin.H{
				"healthy":   healthy,
				"message":   message,
				"timestamp": time.Now().UTC(),
			},
		})
	}
}

// RequestResilienceMiddleware applies circuit breaker to incoming requests
type RequestResilienceMiddleware struct {
	cbName string
	config *resilience.CircuitBreakerConfig
}

// NewRequestResilienceMiddleware creates middleware that applies circuit breaker to requests
func NewRequestResilienceMiddleware(circuitBreakerName string, config *resilience.CircuitBreakerConfig) *RequestResilienceMiddleware {
	return &RequestResilienceMiddleware{
		cbName: circuitBreakerName,
		config: config,
	}
}

// Apply applies circuit breaker to the request pipeline
func (m *RequestResilienceMiddleware) Apply() gin.HandlerFunc {
	cb := resilience.GetCircuitBreakerWithConfig(m.cbName, m.config)

	return func(c *gin.Context) {
		err := cb.CallContext(c.Request.Context(), func(ctx context.Context) error {
			c.Next()

			// Consider 5xx status codes as failures
			if c.Writer.Status() >= 500 {
				return fmt.Errorf("server error: %d", c.Writer.Status())
			}

			return nil
		})

		if err != nil {
			if resilience.IsCircuitBreakerError(err) {
				c.JSON(http.StatusServiceUnavailable, common.APIResponse{
					Success: false,
					Error: &common.APIError{
						Code:    "SERVICE_UNAVAILABLE",
						Message: "Service temporarily unavailable due to circuit breaker",
						Details: err.Error(),
					},
				})
				c.Abort()
				return
			}
		}
	}
}

// ServiceResilienceWrapper wraps service calls with resilience patterns
type ServiceResilienceWrapper struct {
	serviceName string
	wrapper     *resilience.ExternalServiceWrapper
}

// NewServiceResilienceWrapper creates a new service resilience wrapper
func NewServiceResilienceWrapper(serviceName string, config *resilience.ServiceConfig) *ServiceResilienceWrapper {
	wrapper := resilience.NewExternalServiceWrapper(serviceName, config)

	return &ServiceResilienceWrapper{
		serviceName: serviceName,
		wrapper:     wrapper,
	}
}

// WrapCall wraps a service call with resilience patterns
func (w *ServiceResilienceWrapper) WrapCall(c *gin.Context, operation func(ctx context.Context) error) error {
	return w.wrapper.Call(c.Request.Context(), operation)
}

// WrapCallWithResult wraps a service call that returns a result
func (w *ServiceResilienceWrapper) WrapCallWithResult(c *gin.Context, operation func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	return w.wrapper.CallWithResult(c.Request.Context(), operation)
}

// DatabaseResilienceWrapper wraps database operations with resilience
type DatabaseResilienceWrapper struct {
	wrapper *resilience.DatabaseWrapper
}

// NewDatabaseResilienceWrapper creates a new database resilience wrapper
func NewDatabaseResilienceWrapper(config *resilience.ServiceConfig) *DatabaseResilienceWrapper {
	wrapper := resilience.NewDatabaseWrapper(config)

	return &DatabaseResilienceWrapper{
		wrapper: wrapper,
	}
}

// WrapDBCall wraps a database call with resilience patterns
func (w *DatabaseResilienceWrapper) WrapDBCall(c *gin.Context, operation func(ctx context.Context) error) error {
	return w.wrapper.Execute(c.Request.Context(), operation)
}

// WrapDBCallWithResult wraps a database call that returns a result
func (w *DatabaseResilienceWrapper) WrapDBCallWithResult(c *gin.Context, operation func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	return w.wrapper.ExecuteWithResult(c.Request.Context(), operation)
}

// ResilienceConfig holds configuration for resilience middleware
type ResilienceConfig struct {
	// EnableCircuitBreaker enables circuit breaker for requests
	EnableCircuitBreaker bool
	// CircuitBreakerName is the name of the circuit breaker
	CircuitBreakerName string
	// CircuitBreakerConfig is the configuration for the circuit breaker
	CircuitBreakerConfig *resilience.CircuitBreakerConfig
	// EnableRequestLogging enables request logging for resilience monitoring
	EnableRequestLogging bool
}

// DefaultResilienceConfig returns default resilience configuration
func DefaultResilienceConfig() *ResilienceConfig {
	return &ResilienceConfig{
		EnableCircuitBreaker: true,
		CircuitBreakerName:   "default_request_cb",
		CircuitBreakerConfig: resilience.DefaultCircuitBreakerConfig("default_request_cb"),
		EnableRequestLogging: true,
	}
}

// ResilienceMonitoringMiddleware provides comprehensive resilience monitoring
type ResilienceMonitoringMiddleware struct {
	config *ResilienceConfig
}

// NewResilienceMonitoringMiddleware creates a new resilience monitoring middleware
func NewResilienceMonitoringMiddleware(config *ResilienceConfig) *ResilienceMonitoringMiddleware {
	if config == nil {
		config = DefaultResilienceConfig()
	}

	return &ResilienceMonitoringMiddleware{
		config: config,
	}
}

// Monitor provides monitoring for request resilience
func (m *ResilienceMonitoringMiddleware) Monitor() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Log request start if enabled
		if m.config.EnableRequestLogging {
			c.Header("X-Request-Start", start.Format(time.RFC3339))
		}

		c.Next()

		// Log request completion
		duration := time.Since(start)
		if m.config.EnableRequestLogging {
			c.Header("X-Request-Duration", duration.String())

			// Add circuit breaker status if enabled
			if m.config.EnableCircuitBreaker {
				cb := resilience.GetCircuitBreaker(m.config.CircuitBreakerName)
				if cb != nil {
					c.Header("X-Circuit-Breaker-State", cb.State().String())
				}
			}
		}
	}
}

// Add route group for resilience endpoints
func (m *ResilienceMiddleware) RegisterRoutes(router *gin.RouterGroup) {
	resilience := router.Group("/resilience")
	{
		resilience.GET("/circuit-breakers", m.GetCircuitBreakersStatus())
		resilience.GET("/circuit-breakers/:name", m.GetCircuitBreakerDetails())
		resilience.POST("/circuit-breakers/:name/reset", m.ResetCircuitBreaker())
		resilience.POST("/circuit-breakers/:name/force-open", m.ForceOpenCircuitBreaker())
		resilience.GET("/health", m.CircuitBreakerHealthCheck())
	}
}
