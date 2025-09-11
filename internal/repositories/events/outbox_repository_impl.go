package events

import (
    "context"
    "fmt"
    "time"

    "gorm.io/gorm"

    "kisanlink-ecom/entities/models/outbox"
)

// outboxRepositoryImpl implements the OutboxRepository interface
type outboxRepositoryImpl struct {
    db *gorm.DB
}

// NewOutboxRepository creates a new outbox repository
func NewOutboxRepository(db *gorm.DB) OutboxRepository {
    return &outboxRepositoryImpl{
        db: db,
    }
}

// Create stores a new outbox event
func (r *outboxRepositoryImpl) Create(ctx context.Context, event *outbox.OutboxEvent) error {
    if err := r.db.WithContext(ctx).Create(event).Error; err != nil {
        return fmt.Errorf("failed to create outbox event: %w", err)
    }
    return nil
}

// Update updates an existing outbox event
func (r *outboxRepositoryImpl) Update(ctx context.Context, event *outbox.OutboxEvent) error {
    if err := r.db.WithContext(ctx).Save(event).Error; err != nil {
        return fmt.Errorf("failed to update outbox event: %w", err)
    }
    return nil
}

// GetByID retrieves an outbox event by ID
func (r *outboxRepositoryImpl) GetByID(ctx context.Context, id string) (*outbox.OutboxEvent, error) {
    var event outbox.OutboxEvent
    if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&event).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, fmt.Errorf("outbox event not found: %s", id)
        }
        return nil, fmt.Errorf("failed to get outbox event: %w", err)
    }
    return &event, nil
}

// GetUnpublished retrieves unpublished events for retry processing
func (r *outboxRepositoryImpl) GetUnpublished(ctx context.Context, limit int) ([]*outbox.OutboxEvent, error) {
    var events []*outbox.OutboxEvent

    query := r.db.WithContext(ctx).
        Where("published_at IS NULL AND deleted_at IS NULL").
        Order("created_at ASC")

    if limit > 0 {
        query = query.Limit(limit)
    }

    if err := query.Find(&events).Error; err != nil {
        return nil, fmt.Errorf("failed to get unpublished events: %w", err)
    }

    return events, nil
}

// GetByIdempotencyKey retrieves an event by idempotency key for deduplication
func (r *outboxRepositoryImpl) GetByIdempotencyKey(ctx context.Context, key string) (*outbox.OutboxEvent, error) {
    var event outbox.OutboxEvent
    if err := r.db.WithContext(ctx).Where("idempotency_key = ? AND deleted_at IS NULL", key).First(&event).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, nil // Not found is not an error for deduplication check
        }
        return nil, fmt.Errorf("failed to get event by idempotency key: %w", err)
    }
    return &event, nil
}

// DeletePublished removes published events older than specified duration
func (r *outboxRepositoryImpl) DeletePublished(ctx context.Context, olderThanDays int) error {
    cutoffDate := time.Now().AddDate(0, 0, -olderThanDays)

    result := r.db.WithContext(ctx).
        Where("published_at IS NOT NULL AND published_at < ?", cutoffDate).
        Delete(&outbox.OutboxEvent{})

    if result.Error != nil {
        return fmt.Errorf("failed to delete old published events: %w", result.Error)
    }

    return nil
}

// GetEventsByAggregate retrieves events for a specific aggregate
func (r *outboxRepositoryImpl) GetEventsByAggregate(ctx context.Context, aggregateType, aggregateID string, limit int) ([]*outbox.OutboxEvent, error) {
    var events []*outbox.OutboxEvent

    query := r.db.WithContext(ctx).
        Where("aggregate_type = ? AND aggregate_id = ? AND deleted_at IS NULL", aggregateType, aggregateID).
        Order("created_at ASC")

    if limit > 0 {
        query = query.Limit(limit)
    }

    if err := query.Find(&events).Error; err != nil {
        return nil, fmt.Errorf("failed to get events by aggregate: %w", err)
    }

    return events, nil
}
