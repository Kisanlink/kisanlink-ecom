package outbox

import (
	"fmt"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// OutboxEvent represents events for federated network publishing
type OutboxEvent struct {
	base.BaseModel

	// Event Identification
	EventType     string `json:"event_type" gorm:"type:varchar(100);not null;index"`          // e.g., 'OrderCreated', 'CatalogItemUpdated'
	EventVersion  string `json:"event_version" gorm:"type:varchar(10);not null;default:'v1'"` // Event schema version
	AggregateType string `json:"aggregate_type" gorm:"type:varchar(50);not null;index"`       // 'order', 'catalog_item', 'inventory_lot'
	AggregateID   string `json:"aggregate_id" gorm:"type:varchar(255);not null;index"`        // ID of the entity that changed

	// Event Content
	EventData     string `json:"event_data" gorm:"type:jsonb;not null"` // Event payload
	EventMetadata string `json:"event_metadata" gorm:"type:jsonb"`      // Additional metadata

	// Deduplication
	IdempotencyKey string `json:"idempotency_key" gorm:"type:varchar(255);uniqueIndex"` // For deduplication

	// Publishing Status
	PublishedAt    *time.Time `json:"published_at" gorm:"type:timestamp"` // When event was published
	FailedAttempts int        `json:"failed_attempts" gorm:"default:0"`   // Number of failed publish attempts
	LastError      string     `json:"last_error" gorm:"type:text"`        // Last error message
}

// TableName returns the table name for GORM
func (OutboxEvent) TableName() string {
	return "outbox_events"
}

// NewOutboxEvent creates a new outbox event
func NewOutboxEvent(eventType, aggregateType, aggregateID string, eventData map[string]interface{}) *OutboxEvent {
	return &OutboxEvent{
		BaseModel:      *base.NewBaseModel("EVT", "large"),
		EventType:      eventType,
		EventVersion:   "v1",
		AggregateType:  aggregateType,
		AggregateID:    aggregateID,
		IdempotencyKey: generateIdempotencyKey(eventType, aggregateID),
		FailedAttempts: 0,
	}
}

// generateIdempotencyKey creates a unique key for deduplication
func generateIdempotencyKey(eventType, aggregateID string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s:%s:%d", eventType, aggregateID, timestamp)
}

// IsPublished checks if the event has been published
func (e *OutboxEvent) IsPublished() bool {
	return e.PublishedAt != nil
}

// MarkAsPublished marks the event as successfully published
func (e *OutboxEvent) MarkAsPublished() {
	now := time.Now()
	e.PublishedAt = &now
}

// RecordFailure records a failed publish attempt
func (e *OutboxEvent) RecordFailure(errorMsg string) {
	e.FailedAttempts++
	e.LastError = errorMsg
}

// ShouldRetry determines if the event should be retried based on failure count
func (e *OutboxEvent) ShouldRetry() bool {
	maxRetries := 5 // Configure as needed
	return e.FailedAttempts < maxRetries
}
