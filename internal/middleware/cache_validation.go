package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	// IfNoneMatchHeader is the header name for conditional requests
	IfNoneMatchHeader = "If-None-Match"
)

// CacheValidationMiddleware provides ETag-based cache validation
type CacheValidationMiddleware struct {
	logger *logrus.Logger
}

// NewCacheValidationMiddleware creates a new cache validation middleware
func NewCacheValidationMiddleware(logger *logrus.Logger) *CacheValidationMiddleware {
	if logger == nil {
		logger = logrus.StandardLogger()
	}

	return &CacheValidationMiddleware{
		logger: logger,
	}
}

// CacheValidationHandler provides ETag-based cache validation for GET requests
func (m *CacheValidationMiddleware) CacheValidationHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to GET requests
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		// Get If-None-Match header
		ifNoneMatch := c.GetHeader(IfNoneMatchHeader)
		if ifNoneMatch == "" {
			c.Next()
			return
		}

		// Store If-None-Match value in context for handlers to use
		c.Set("if_none_match", ifNoneMatch)

		m.logger.WithFields(logrus.Fields{
			"method":        c.Request.Method,
			"path":          c.Request.URL.Path,
			"if_none_match": ifNoneMatch,
		}).Debug("Processing conditional GET request")

		c.Next()
	}
}

// CheckETagMatch checks if the provided ETag matches the If-None-Match header
// Returns true if ETags match (indicating client has current version)
func (m *CacheValidationMiddleware) CheckETagMatch(c *gin.Context, currentETag string) bool {
	ifNoneMatch, exists := c.Get("if_none_match")
	if !exists {
		return false
	}

	ifNoneMatchStr, ok := ifNoneMatch.(string)
	if !ok {
		return false
	}

	// Handle multiple ETags in If-None-Match (comma-separated)
	etags := m.parseIfNoneMatch(ifNoneMatchStr)

	// Check if current ETag matches any of the provided ETags
	for _, etag := range etags {
		if etag == "*" || etag == currentETag {
			m.logger.WithFields(logrus.Fields{
				"current_etag":  currentETag,
				"if_none_match": ifNoneMatchStr,
				"matched_etag":  etag,
			}).Debug("ETag match found - returning 304 Not Modified")
			return true
		}
	}

	return false
}

// ReturnNotModified returns a 304 Not Modified response
func (m *CacheValidationMiddleware) ReturnNotModified(c *gin.Context, etag string) {
	// Set ETag header
	c.Header(ETagHeader, etag)

	// Set cache control headers
	c.Header("Cache-Control", "private, must-revalidate")

	m.logger.WithFields(logrus.Fields{
		"path":   c.Request.URL.Path,
		"etag":   etag,
		"status": http.StatusNotModified,
	}).Info("Returning 304 Not Modified response")

	c.Status(http.StatusNotModified)
	c.Abort()
}

// parseIfNoneMatch parses the If-None-Match header value
// Handles both single ETags and comma-separated lists
func (m *CacheValidationMiddleware) parseIfNoneMatch(ifNoneMatch string) []string {
	// Handle wildcard
	if strings.TrimSpace(ifNoneMatch) == "*" {
		return []string{"*"}
	}

	// Split by comma and clean up each ETag
	parts := strings.Split(ifNoneMatch, ",")
	etags := make([]string, 0, len(parts))

	for _, part := range parts {
		etag := strings.TrimSpace(part)
		// Remove quotes if present
		etag = strings.Trim(etag, `"`)
		if etag != "" {
			// Add quotes back for consistent format
			etags = append(etags, `"`+etag+`"`)
		}
	}

	return etags
}

// ValidateETagFormat validates ETag format
func (m *CacheValidationMiddleware) ValidateETagFormat(etag string) error {
	if etag == "" {
		return nil // Empty ETag is valid (no caching)
	}

	// ETag should be quoted
	if !strings.HasPrefix(etag, `"`) || !strings.HasSuffix(etag, `"`) {
		return &ValidationError{
			Field:   "etag",
			Message: "ETag must be quoted",
		}
	}

	// Extract content without quotes
	content := strings.Trim(etag, `"`)
	if len(content) == 0 {
		return &ValidationError{
			Field:   "etag",
			Message: "ETag content cannot be empty",
		}
	}

	return nil
}

// ValidationError represents an ETag validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// ETagGenerator provides utilities for generating ETags
type ETagGenerator struct {
	middleware *IdempotencyMiddleware
}

// NewETagGenerator creates a new ETag generator
func NewETagGenerator(middleware *IdempotencyMiddleware) *ETagGenerator {
	return &ETagGenerator{
		middleware: middleware,
	}
}

// GenerateForEntity generates an ETag for an entity with version and timestamp
func (g *ETagGenerator) GenerateForEntity(entityID string, version int64, updatedAt interface{}) string {
	if g.middleware == nil {
		// Fallback implementation if middleware is not available
		return g.generateSimpleETag(entityID, version)
	}

	// Use the existing middleware ETag generation
	switch v := updatedAt.(type) {
	case time.Time:
		return g.middleware.GenerateETag(entityID, version, v)
	case string:
		// Parse timestamp string if needed
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return g.middleware.GenerateETag(entityID, version, t)
		}
		return g.generateSimpleETag(entityID, version)
	default:
		// Use middleware's ETag generation if timestamp interface is available
		if timestamp, ok := updatedAt.(interface{ Unix() int64 }); ok {
			t := time.Unix(timestamp.Unix(), 0)
			return g.middleware.GenerateETag(entityID, version, t)
		}
		return g.generateSimpleETag(entityID, version)
	}
}

// generateSimpleETag generates a simple ETag based on entity ID and version
func (g *ETagGenerator) generateSimpleETag(entityID string, version int64) string {
	// Simple ETag format: "entityID-version"
	return fmt.Sprintf(`"%s-%d"`, entityID, version)
}

// CacheControlConfig holds cache control configuration
type CacheControlConfig struct {
	MaxAge         int  // Max age in seconds
	Private        bool // Private cache
	MustRevalidate bool // Must revalidate
	NoCache        bool // No cache
	NoStore        bool // No store
}

// DefaultCacheControlConfig returns default cache control configuration
func DefaultCacheControlConfig() *CacheControlConfig {
	return &CacheControlConfig{
		MaxAge:         300, // 5 minutes
		Private:        true,
		MustRevalidate: true,
		NoCache:        false,
		NoStore:        false,
	}
}

// SetCacheControlHeaders sets appropriate cache control headers
func (m *CacheValidationMiddleware) SetCacheControlHeaders(c *gin.Context, config *CacheControlConfig) {
	if config == nil {
		config = DefaultCacheControlConfig()
	}

	var directives []string

	if config.NoStore {
		directives = append(directives, "no-store")
	} else if config.NoCache {
		directives = append(directives, "no-cache")
	} else {
		if config.Private {
			directives = append(directives, "private")
		} else {
			directives = append(directives, "public")
		}

		if config.MaxAge > 0 {
			directives = append(directives, "max-age="+strconv.Itoa(config.MaxAge))
		}

		if config.MustRevalidate {
			directives = append(directives, "must-revalidate")
		}
	}

	if len(directives) > 0 {
		c.Header("Cache-Control", strings.Join(directives, ", "))
	}
}
