package common

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// OptimisticLockError represents an optimistic locking conflict
type OptimisticLockError struct {
	EntityID       string    `json:"entity_id"`
	ExpectedETag   string    `json:"expected_etag"`
	CurrentETag    string    `json:"current_etag"`
	CurrentVersion int64     `json:"current_version"`
	LastModified   time.Time `json:"last_modified"`
}

func (e *OptimisticLockError) Error() string {
	return fmt.Sprintf("optimistic locking conflict for entity %s: expected ETag %s, current ETag %s",
		e.EntityID, e.ExpectedETag, e.CurrentETag)
}

// NewOptimisticLockError creates a new optimistic locking error
func NewOptimisticLockError(entityID, expectedETag, currentETag string, currentVersion int64, lastModified time.Time) *OptimisticLockError {
	return &OptimisticLockError{
		EntityID:       entityID,
		ExpectedETag:   expectedETag,
		CurrentETag:    currentETag,
		CurrentVersion: currentVersion,
		LastModified:   lastModified,
	}
}

// ToAppError converts OptimisticLockError to AppError
func (e *OptimisticLockError) ToAppError() *AppError {
	return &AppError{
		Code:       ErrorCodeBusinessRuleViolation,
		Message:    "Optimistic locking conflict - entity has been modified by another request",
		Details:    e.Error(),
		StatusCode: http.StatusPreconditionFailed,
		Context: map[string]interface{}{
			"entity_id":       e.EntityID,
			"expected_etag":   e.ExpectedETag,
			"current_etag":    e.CurrentETag,
			"current_version": e.CurrentVersion,
			"last_modified":   e.LastModified,
		},
	}
}

// VersionedEntity represents an entity with version information for optimistic locking
type VersionedEntity interface {
	GetID() string
	GetVersion() int64
	GetUpdatedAt() time.Time
	IncrementVersion()
}

// OptimisticLockManager manages optimistic locking operations
type OptimisticLockManager struct{}

// NewOptimisticLockManager creates a new optimistic lock manager
func NewOptimisticLockManager() *OptimisticLockManager {
	return &OptimisticLockManager{}
}

// ValidateVersion validates that the entity version matches the expected version
func (m *OptimisticLockManager) ValidateVersion(entity VersionedEntity, expectedVersion int64) error {
	if entity.GetVersion() != expectedVersion {
		return NewOptimisticLockError(
			entity.GetID(),
			fmt.Sprintf("version-%d", expectedVersion),
			fmt.Sprintf("version-%d", entity.GetVersion()),
			entity.GetVersion(),
			entity.GetUpdatedAt(),
		)
	}
	return nil
}

// ValidateETag validates that the entity ETag matches the expected ETag
func (m *OptimisticLockManager) ValidateETag(entity VersionedEntity, expectedETag, currentETag string) error {
	if expectedETag != currentETag {
		return NewOptimisticLockError(
			entity.GetID(),
			expectedETag,
			currentETag,
			entity.GetVersion(),
			entity.GetUpdatedAt(),
		)
	}
	return nil
}

// PrepareForUpdate prepares an entity for update by incrementing version
func (m *OptimisticLockManager) PrepareForUpdate(entity VersionedEntity) {
	entity.IncrementVersion()
}

// ExtractVersionFromContext extracts version information from gin context
func (m *OptimisticLockManager) ExtractVersionFromContext(c *gin.Context) (int64, bool) {
	if version, exists := c.Get("expected_version"); exists {
		if v, ok := version.(int64); ok {
			return v, true
		}
	}
	return 0, false
}

// ExtractETagFromContext extracts ETag from gin context
func (m *OptimisticLockManager) ExtractETagFromContext(c *gin.Context) (string, bool) {
	if etag, exists := c.Get("if_match_etag"); exists {
		if e, ok := etag.(string); ok {
			return e, true
		}
	}
	return "", false
}

// IdempotencyManager manages idempotency operations
type IdempotencyManager struct{}

// NewIdempotencyManager creates a new idempotency manager
func NewIdempotencyManager() *IdempotencyManager {
	return &IdempotencyManager{}
}

// ExtractIdempotencyKey extracts idempotency key from gin context
func (m *IdempotencyManager) ExtractIdempotencyKey(c *gin.Context) (string, bool) {
	if key, exists := c.Get("idempotency_key"); exists {
		if k, ok := key.(string); ok {
			return k, true
		}
	}
	return "", false
}

// IsIdempotentRequest checks if the request has an idempotency key
func (m *IdempotencyManager) IsIdempotentRequest(c *gin.Context) bool {
	_, exists := m.ExtractIdempotencyKey(c)
	return exists
}

// ConflictError represents a conflict error for idempotency or optimistic locking
type ConflictError struct {
	Type    string                 `json:"type"` // "idempotency" or "optimistic_lock"
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details"`
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("%s conflict: %s", e.Type, e.Message)
}

// ToAppError converts ConflictError to AppError
func (e *ConflictError) ToAppError() *AppError {
	var code ErrorCode
	var statusCode int

	switch e.Type {
	case "idempotency":
		code = ErrorCodeDuplicateRecord
		statusCode = http.StatusConflict
	case "optimistic_lock":
		code = ErrorCodeBusinessRuleViolation
		statusCode = http.StatusPreconditionFailed
	default:
		code = ErrorCodeBusinessRuleViolation
		statusCode = http.StatusConflict
	}

	return &AppError{
		Code:       code,
		Message:    e.Message,
		Details:    fmt.Sprintf("%s conflict", e.Type),
		StatusCode: statusCode,
		Context:    e.Details,
	}
}
