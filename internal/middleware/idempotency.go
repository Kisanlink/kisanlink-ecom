package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"kisanlink-ecom/entities/models/common"
	commonErrors "kisanlink-ecom/internal/common"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

const (
	// IdempotencyKeyHeader is the header name for idempotency keys
	IdempotencyKeyHeader = "Idempotency-Key"

	// IfMatchHeader is the header name for optimistic locking
	IfMatchHeader = "If-Match"

	// ETagHeader is the header name for entity tags
	ETagHeader = "ETag"

	// DefaultIdempotencyTTL is the default TTL for idempotency keys (24 hours)
	DefaultIdempotencyTTL = 24 * time.Hour

	// MaxIdempotencyKeyLength is the maximum allowed length for idempotency keys
	MaxIdempotencyKeyLength = 255
)

// IdempotencyConfig holds configuration for idempotency middleware
type IdempotencyConfig struct {
	// TTL for idempotency keys in cache
	TTL time.Duration

	// Cache for storing idempotency keys
	Cache *cache.Cache

	// RequiredMethods specifies which HTTP methods require idempotency keys
	RequiredMethods []string

	// OptionalMethods specifies which HTTP methods support optional idempotency keys
	OptionalMethods []string

	// EnableOptimisticLocking enables optimistic locking support
	EnableOptimisticLocking bool

	// Logger for logging idempotency events
	Logger *logrus.Logger
}

// DefaultIdempotencyConfig returns default idempotency configuration
func DefaultIdempotencyConfig() *IdempotencyConfig {
	return &IdempotencyConfig{
		TTL:                     DefaultIdempotencyTTL,
		Cache:                   cache.New(DefaultIdempotencyTTL, 10*time.Minute), // 10 minute cleanup interval
		RequiredMethods:         []string{"POST"},
		OptionalMethods:         []string{"PUT", "PATCH"},
		EnableOptimisticLocking: true,
		Logger:                  logrus.StandardLogger(),
	}
}

// IdempotencyMiddleware provides idempotency and optimistic locking support
type IdempotencyMiddleware struct {
	config *IdempotencyConfig
}

// NewIdempotencyMiddleware creates a new idempotency middleware
func NewIdempotencyMiddleware(config *IdempotencyConfig) *IdempotencyMiddleware {
	if config == nil {
		config = DefaultIdempotencyConfig()
	}

	if config.Cache == nil {
		config.Cache = cache.New(config.TTL, 10*time.Minute)
	}

	return &IdempotencyMiddleware{
		config: config,
	}
}

// IdempotencyHandler provides idempotency support for write operations
func (m *IdempotencyMiddleware) IdempotencyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		// Check if idempotency is required or supported for this method
		isRequired := m.isMethodRequired(method)
		isSupported := isRequired || m.isMethodSupported(method)

		if !isSupported {
			c.Next()
			return
		}

		// Get idempotency key from header
		idempotencyKey := c.GetHeader(IdempotencyKeyHeader)

		// Validate idempotency key
		if isRequired && idempotencyKey == "" {
			m.handleError(c, commonErrors.NewValidationError(
				fmt.Sprintf("Idempotency-Key header is required for %s requests", method),
			).WithField("header", IdempotencyKeyHeader))
			return
		}

		if idempotencyKey != "" {
			if err := m.validateIdempotencyKey(idempotencyKey); err != nil {
				m.handleError(c, commonErrors.NewValidationError(
					"Invalid Idempotency-Key format",
				).WithField("idempotency_key", err.Error()))
				return
			}

			// Check for existing response
			if cachedResponse, found := m.getCachedResponse(idempotencyKey); found {
				m.config.Logger.WithFields(logrus.Fields{
					"idempotency_key": idempotencyKey,
					"method":          method,
					"path":            c.Request.URL.Path,
				}).Info("Returning cached idempotent response")

				// Return cached response
				m.returnCachedResponse(c, cachedResponse)
				return
			}

			// Store idempotency key in context for later use
			c.Set("idempotency_key", idempotencyKey)
		}

		c.Next()

		// After request processing, cache the response if idempotency key was provided
		if idempotencyKey != "" && c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			m.cacheResponse(idempotencyKey, c)
		}
	}
}

// OptimisticLockingHandler provides optimistic locking support
func (m *IdempotencyMiddleware) OptimisticLockingHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.EnableOptimisticLocking {
			c.Next()
			return
		}

		method := c.Request.Method

		// Only apply optimistic locking to update operations
		if method != "PUT" && method != "PATCH" {
			c.Next()
			return
		}

		// Get If-Match header (ETag for optimistic locking)
		ifMatch := c.GetHeader(IfMatchHeader)

		if ifMatch != "" {
			// Validate ETag format
			if err := m.validateETag(ifMatch); err != nil {
				m.handleError(c, commonErrors.NewValidationError(
					"Invalid If-Match header format",
				).WithField("if_match", err.Error()))
				return
			}

			// Store ETag in context for service layer to use
			c.Set("if_match_etag", ifMatch)
		}

		c.Next()
	}
}

// GenerateETag generates an ETag for a given entity version
func (m *IdempotencyMiddleware) GenerateETag(entityID string, version int64, updatedAt time.Time) string {
	// Create a hash based on entity ID, version, and updated timestamp
	data := fmt.Sprintf("%s:%d:%d", entityID, version, updatedAt.Unix())
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf(`"%s"`, hex.EncodeToString(hash[:8])) // Use first 8 bytes for shorter ETag
}

// ParseETag extracts version information from ETag (if possible)
func (m *IdempotencyMiddleware) ParseETag(etag string) (string, error) {
	// Remove quotes if present
	etag = strings.Trim(etag, `"`)

	// Validate hex format
	if len(etag) != 16 { // 8 bytes = 16 hex characters
		return "", fmt.Errorf("invalid ETag format")
	}

	// Validate hex characters
	if _, err := hex.DecodeString(etag); err != nil {
		return "", fmt.Errorf("invalid ETag hex format: %w", err)
	}

	return etag, nil
}

// SetETagHeader sets the ETag header in the response
func (m *IdempotencyMiddleware) SetETagHeader(c *gin.Context, entityID string, version int64, updatedAt time.Time) {
	etag := m.GenerateETag(entityID, version, updatedAt)
	c.Header(ETagHeader, etag)
}

// CheckOptimisticLock validates optimistic locking constraints
func (m *IdempotencyMiddleware) CheckOptimisticLock(c *gin.Context, currentVersion int64, currentUpdatedAt time.Time, entityID string) error {
	ifMatch, exists := c.Get("if_match_etag")
	if !exists {
		// No optimistic locking requested
		return nil
	}

	expectedETag, ok := ifMatch.(string)
	if !ok {
		return commonErrors.NewValidationError("Invalid If-Match header type")
	}

	// Generate current ETag
	currentETag := m.GenerateETag(entityID, currentVersion, currentUpdatedAt)

	// Compare ETags
	if expectedETag != currentETag {
		return commonErrors.NewAppError(
			commonErrors.ErrorCodeBusinessRuleViolation,
			"Optimistic locking conflict - entity has been modified",
			http.StatusPreconditionFailed,
		).WithContext("expected_etag", expectedETag).
			WithContext("current_etag", currentETag).
			WithContext("entity_id", entityID)
	}

	return nil
}

// Helper methods

func (m *IdempotencyMiddleware) isMethodRequired(method string) bool {
	for _, required := range m.config.RequiredMethods {
		if method == required {
			return true
		}
	}
	return false
}

func (m *IdempotencyMiddleware) isMethodSupported(method string) bool {
	for _, supported := range m.config.OptionalMethods {
		if method == supported {
			return true
		}
	}
	return false
}

func (m *IdempotencyMiddleware) validateIdempotencyKey(key string) error {
	if len(key) == 0 {
		return fmt.Errorf("idempotency key cannot be empty")
	}

	if len(key) > MaxIdempotencyKeyLength {
		return fmt.Errorf("idempotency key too long (max %d characters)", MaxIdempotencyKeyLength)
	}

	// Check for valid characters (alphanumeric, hyphens, underscores)
	for _, char := range key {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return fmt.Errorf("idempotency key contains invalid characters (only alphanumeric, hyphens, and underscores allowed)")
		}
	}

	return nil
}

func (m *IdempotencyMiddleware) validateETag(etag string) error {
	if etag == "" {
		return fmt.Errorf("ETag cannot be empty")
	}

	// Parse and validate ETag
	_, err := m.ParseETag(etag)
	return err
}

func (m *IdempotencyMiddleware) getCachedResponse(idempotencyKey string) (*CachedResponseData, bool) {
	// Try to get cached response from cache
	if data, found := m.config.Cache.Get(idempotencyKey); found {
		if cachedResponse, ok := data.(*CachedResponseData); ok {
			return cachedResponse, true
		}
	}
	return nil, false
}

func (m *IdempotencyMiddleware) cacheResponse(idempotencyKey string, c *gin.Context) {
	// Create cached response data
	cachedData := NewCachedResponseData(c)

	// Store in cache with TTL
	m.config.Cache.Set(idempotencyKey, cachedData, m.config.TTL)

	m.config.Logger.WithFields(logrus.Fields{
		"idempotency_key": idempotencyKey,
		"status_code":     cachedData.StatusCode,
	}).Debug("Cached idempotent response")
}

func (m *IdempotencyMiddleware) returnCachedResponse(c *gin.Context, cached *CachedResponseData) {
	// Apply cached response to context
	cached.ApplyToContext(c)
	c.Abort()
}

func (m *IdempotencyMiddleware) handleError(c *gin.Context, err *commonErrors.AppError) {
	c.AbortWithStatusJSON(err.StatusCode, common.APIResponse{
		Success: false,
		Error: &common.APIError{
			Code:    string(err.Code),
			Message: err.Message,
			Details: err.Details,
			Fields:  err.Fields,
			Context: err.Context,
		},
	})
}

// IdempotencyKeyGenerator provides utilities for generating idempotency keys
type IdempotencyKeyGenerator struct{}

// NewIdempotencyKeyGenerator creates a new idempotency key generator
func NewIdempotencyKeyGenerator() *IdempotencyKeyGenerator {
	return &IdempotencyKeyGenerator{}
}

// GenerateKey generates a unique idempotency key based on request content
func (g *IdempotencyKeyGenerator) GenerateKey(userID, operation string, content []byte) string {
	// Create a hash of user ID, operation, and content
	data := fmt.Sprintf("%s:%s:%s", userID, operation, string(content))
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes for reasonable length
}

// GenerateKeyWithTimestamp generates a key with timestamp for uniqueness
func (g *IdempotencyKeyGenerator) GenerateKeyWithTimestamp(userID, operation string) string {
	timestamp := time.Now().UnixNano()
	data := fmt.Sprintf("%s:%s:%d", userID, operation, timestamp)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}

// ValidateKey validates an idempotency key format
func (g *IdempotencyKeyGenerator) ValidateKey(key string) error {
	if len(key) == 0 {
		return fmt.Errorf("idempotency key cannot be empty")
	}

	if len(key) > MaxIdempotencyKeyLength {
		return fmt.Errorf("idempotency key too long")
	}

	// Additional validation logic can be added here
	return nil
}
