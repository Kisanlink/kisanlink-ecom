package events

import (
	"context"

	"kisanlink-ecom/entities/models/outbox"
)

// OutboxRepository defines the interface for outbox event persistence
type OutboxRepository interface {
	// Create stores a new outbox event
	Create(ctx context.Context, event *outbox.OutboxEvent) error

	// Update updates an existing outbox event
	Update(ctx context.Context, event *outbox.OutboxEvent) error

	// GetByID retrieves an outbox event by ID
	GetByID(ctx context.Context, id string) (*outbox.OutboxEvent, error)

	// GetUnpublished retrieves unpublished events for retry processing
	GetUnpublished(ctx context.Context, limit int) ([]*outbox.OutboxEvent, error)

	// GetByIdempotencyKey retrieves an event by idempotency key for deduplication
	GetByIdempotencyKey(ctx context.Context, key string) (*outbox.OutboxEvent, error)

	// DeletePublished removes published events older than specified duration
	DeletePublished(ctx context.Context, olderThanDays int) error

	// GetEventsByAggregate retrieves events for a specific aggregate
	GetEventsByAggregate(ctx context.Context, aggregateType, aggregateID string, limit int) ([]*outbox.OutboxEvent, error)
}
