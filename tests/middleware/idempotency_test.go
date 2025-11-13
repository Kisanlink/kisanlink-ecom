package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kisanlink-ecom/entities/models/catalog"
	"kisanlink-ecom/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdempotencyMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		idempotencyKey string
		expectError    bool
		errorCode      string
	}{
		{
			name:           "POST with valid idempotency key",
			method:         "POST",
			idempotencyKey: "test-key-123",
			expectError:    false,
		},
		{
			name:           "POST without idempotency key should fail",
			method:         "POST",
			idempotencyKey: "",
			expectError:    true,
			errorCode:      "VALIDATION_FAILED",
		},
		{
			name:           "PUT with optional idempotency key",
			method:         "PUT",
			idempotencyKey: "test-key-456",
			expectError:    false,
		},
		{
			name:           "PUT without idempotency key should pass",
			method:         "PUT",
			idempotencyKey: "",
			expectError:    false,
		},
		{
			name:           "GET should not require idempotency key",
			method:         "GET",
			idempotencyKey: "",
			expectError:    false,
		},
		{
			name:           "POST with invalid idempotency key",
			method:         "POST",
			idempotencyKey: "invalid key with spaces!",
			expectError:    true,
			errorCode:      "VALIDATION_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			config := middleware.DefaultIdempotencyConfig()
			idempotencyMiddleware := middleware.NewIdempotencyMiddleware(config)

			router := gin.New()
			router.Use(middleware.ResponseCapture())
			router.Use(idempotencyMiddleware.IdempotencyHandler())

			// Test handler
			router.Any("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Create request
			req := httptest.NewRequest(tt.method, "/test", bytes.NewBuffer([]byte(`{"test": "data"}`)))
			req.Header.Set("Content-Type", "application/json")

			if tt.idempotencyKey != "" {
				req.Header.Set(middleware.IdempotencyKeyHeader, tt.idempotencyKey)
			}

			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assertions
			if tt.expectError {
				assert.Equal(t, http.StatusBadRequest, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.False(t, response["success"].(bool))
				if tt.errorCode != "" {
					errorObj := response["error"].(map[string]interface{})
					assert.Equal(t, tt.errorCode, errorObj["code"])
				}
			} else {
				assert.Equal(t, http.StatusOK, w.Code)
			}
		})
	}
}

func TestIdempotencyKeyValidation(t *testing.T) {
	config := middleware.DefaultIdempotencyConfig()
	idempotencyMiddleware := middleware.NewIdempotencyMiddleware(config)
	print(idempotencyMiddleware)

	tests := []struct {
		name        string
		key         string
		expectError bool
	}{
		{
			name:        "Valid alphanumeric key",
			key:         "abc123",
			expectError: false,
		},
		{
			name:        "Valid key with hyphens",
			key:         "test-key-123",
			expectError: false,
		},
		{
			name:        "Valid key with underscores",
			key:         "test_key_123",
			expectError: false,
		},
		{
			name:        "Empty key",
			key:         "",
			expectError: true,
		},
		{
			name:        "Key with spaces",
			key:         "test key",
			expectError: true,
		},
		{
			name:        "Key with special characters",
			key:         "test@key!",
			expectError: true,
		},
		{
			name:        "Too long key",
			key:         string(make([]byte, 300)), // Longer than MaxIdempotencyKeyLength
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := middleware.NewIdempotencyKeyGenerator()
			err := generator.ValidateKey(tt.key)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOptimisticLockingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		method      string
		ifMatch     string
		expectError bool
	}{
		{
			name:        "PUT with valid ETag",
			method:      "PUT",
			ifMatch:     `"1234567890abcdef"`,
			expectError: false,
		},
		{
			name:        "PATCH with valid ETag",
			method:      "PATCH",
			ifMatch:     `"abcdef1234567890"`,
			expectError: false,
		},
		{
			name:        "PUT without ETag should pass",
			method:      "PUT",
			ifMatch:     "",
			expectError: false,
		},
		{
			name:        "GET should not process ETag",
			method:      "GET",
			ifMatch:     `"1234567890abcdef"`,
			expectError: false,
		},
		{
			name:        "PUT with invalid ETag format",
			method:      "PUT",
			ifMatch:     "invalid-etag",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			config := middleware.DefaultIdempotencyConfig()
			idempotencyMiddleware := middleware.NewIdempotencyMiddleware(config)

			router := gin.New()
			router.Use(idempotencyMiddleware.OptimisticLockingHandler())

			// Test handler
			router.Any("/test", func(c *gin.Context) {
				// Check if ETag was stored in context
				if etag, exists := c.Get("if_match_etag"); exists {
					c.JSON(http.StatusOK, gin.H{"etag": etag})
				} else {
					c.JSON(http.StatusOK, gin.H{"message": "no etag"})
				}
			})

			// Create request
			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.ifMatch != "" {
				req.Header.Set(middleware.IfMatchHeader, tt.ifMatch)
			}

			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assertions
			if tt.expectError {
				assert.Equal(t, http.StatusBadRequest, w.Code)
			} else {
				assert.Equal(t, http.StatusOK, w.Code)
			}
		})
	}
}

func TestETagGeneration(t *testing.T) {
	config := middleware.DefaultIdempotencyConfig()
	idempotencyMiddleware := middleware.NewIdempotencyMiddleware(config)

	entityID := "test-entity-123"
	version := int64(5)
	updatedAt := time.Now()

	// Generate ETag
	etag1 := idempotencyMiddleware.GenerateETag(entityID, version, updatedAt)
	etag2 := idempotencyMiddleware.GenerateETag(entityID, version, updatedAt)

	// Same inputs should generate same ETag
	assert.Equal(t, etag1, etag2)

	// Different version should generate different ETag
	etag3 := idempotencyMiddleware.GenerateETag(entityID, version+1, updatedAt)
	assert.NotEqual(t, etag1, etag3)

	// Different timestamp should generate different ETag
	etag4 := idempotencyMiddleware.GenerateETag(entityID, version, updatedAt.Add(time.Second))
	assert.NotEqual(t, etag1, etag4)

	// Validate ETag format (should be quoted hex string)
	assert.True(t, len(etag1) > 2)
	assert.Equal(t, byte('"'), etag1[0])
	assert.Equal(t, byte('"'), etag1[len(etag1)-1])
}

func TestETagParsing(t *testing.T) {
	config := middleware.DefaultIdempotencyConfig()
	idempotencyMiddleware := middleware.NewIdempotencyMiddleware(config)

	tests := []struct {
		name        string
		etag        string
		expectError bool
	}{
		{
			name:        "Valid quoted ETag",
			etag:        `"1234567890abcdef"`,
			expectError: false,
		},
		{
			name:        "Valid unquoted ETag",
			etag:        "1234567890abcdef",
			expectError: false,
		},
		{
			name:        "Invalid ETag - too short",
			etag:        `"123"`,
			expectError: true,
		},
		{
			name:        "Invalid ETag - too long",
			etag:        `"1234567890abcdef123"`,
			expectError: true,
		},
		{
			name:        "Invalid ETag - non-hex characters",
			etag:        `"1234567890abcdefg"`,
			expectError: true,
		},
		{
			name:        "Empty ETag",
			etag:        "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := idempotencyMiddleware.ParseETag(tt.etag)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCatalogItemVersionedEntity(t *testing.T) {
	// Create a test catalog item
	item := catalog.NewCatalogItem("org-123", catalog.CatalogItemTypeProduct, "Test Product", decimal.NewFromFloat(100.0))

	// Test VersionedEntity interface implementation
	assert.Equal(t, item.ID, item.GetID())
	assert.Equal(t, int64(1), item.GetVersion()) // Default version should be 1
	assert.False(t, item.GetUpdatedAt().IsZero())

	// Test version increment
	originalVersion := item.GetVersion()
	item.IncrementVersion()
	assert.Equal(t, originalVersion+1, item.GetVersion())
}

func TestIdempotencyKeyGenerator(t *testing.T) {
	generator := middleware.NewIdempotencyKeyGenerator()

	userID := "user-123"
	operation := "create-catalog"
	content := []byte(`{"name": "test", "price": 100}`)

	// Test key generation
	key1 := generator.GenerateKey(userID, operation, content)
	key2 := generator.GenerateKey(userID, operation, content)

	// Same inputs should generate same key
	assert.Equal(t, key1, key2)

	// Different content should generate different key
	key3 := generator.GenerateKey(userID, operation, []byte(`{"name": "different", "price": 200}`))
	assert.NotEqual(t, key1, key3)

	// Test key with timestamp (should be unique)
	key4 := generator.GenerateKeyWithTimestamp(userID, operation)
	key5 := generator.GenerateKeyWithTimestamp(userID, operation)
	assert.NotEqual(t, key4, key5)

	// Validate generated keys
	assert.NoError(t, generator.ValidateKey(key1))
	assert.NoError(t, generator.ValidateKey(key4))
}
