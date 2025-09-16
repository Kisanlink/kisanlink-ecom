package middleware

import (
	"strconv"
	"time"

	"kisanlink-ecom/internal/observability"

	"github.com/gin-gonic/gin"
)

// MetricsMiddleware collects HTTP metrics for observability
type MetricsMiddleware struct {
	metricsIntegration *observability.MetricsIntegration
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware(metricsIntegration *observability.MetricsIntegration) *MetricsMiddleware {
	return &MetricsMiddleware{
		metricsIntegration: metricsIntegration,
	}
}

// Handler returns the middleware handler function
func (m *MetricsMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Record request start
		m.metricsIntegration.RecordHTTPRequestStart(c.Request.Context())

		// Process request
		c.Next()

		// Calculate metrics
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method

		// Get request and response sizes
		requestSize := c.Request.ContentLength
		if requestSize < 0 {
			requestSize = 0
		}
		responseSize := int64(c.Writer.Size())

		// Record metrics
		m.metricsIntegration.RecordHTTPRequest(
			c.Request.Context(),
			method,
			path,
			statusCode,
			duration,
			requestSize,
			responseSize,
		)

		// Record request end
		m.metricsIntegration.RecordHTTPRequestEnd(c.Request.Context())

		// Add metrics headers for debugging (optional)
		if gin.Mode() == gin.DebugMode {
			c.Header("X-Response-Time", duration.String())
			c.Header("X-Request-Size", strconv.FormatInt(requestSize, 10))
			c.Header("X-Response-Size", strconv.FormatInt(responseSize, 10))
		}
	}
}

// RecordBusinessMetric records business-specific metrics
func (m *MetricsMiddleware) RecordBusinessMetric(c *gin.Context, metricType string, value interface{}) {
	switch metricType {
	case "order_created":
		if orderData, ok := value.(map[string]interface{}); ok {
			userID, _ := orderData["user_id"].(string)
			amount, _ := orderData["amount"].(float64)
			m.metricsIntegration.RecordOrderCreated(c.Request.Context(), userID, amount)
		}
	case "order_completed":
		if userID, ok := value.(string); ok {
			m.metricsIntegration.RecordOrderCompleted(c.Request.Context(), userID)
		}
	case "product_view":
		if productData, ok := value.(map[string]interface{}); ok {
			productID, _ := productData["product_id"].(string)
			userID, _ := productData["user_id"].(string)
			m.metricsIntegration.RecordProductView(c.Request.Context(), productID, userID)
		}
	case "user_registration":
		m.metricsIntegration.RecordUserRegistration(c.Request.Context())
	}
}

// GetMetricsMiddleware helper function to get metrics middleware from context
func GetMetricsMiddleware(c *gin.Context) *MetricsMiddleware {
	if middleware, exists := c.Get("metrics_middleware"); exists {
		if metricsMiddleware, ok := middleware.(*MetricsMiddleware); ok {
			return metricsMiddleware
		}
	}
	return nil
}

// SetMetricsMiddleware helper function to set metrics middleware in context
func SetMetricsMiddleware(c *gin.Context, middleware *MetricsMiddleware) {
	c.Set("metrics_middleware", middleware)
}
