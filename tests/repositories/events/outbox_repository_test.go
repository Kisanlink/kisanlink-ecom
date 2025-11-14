package events_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/outbox"
	"github.com/Kisanlink/kisanlink-ecom/internal/repositories/events"
	"github.com/Kisanlink/kisanlink-ecom/tests/testutils"
)

func TestOutboxRepository_Create(t *testing.T) {
	// Setup test database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, db)

	repo := events.NewOutboxRepository(db)
	ctx := context.Background()

	// Test data
	event := outbox.NewOutboxEvent("OrderCreated", "order", "order-123", nil)
	event.EventData = `{"order_id": "order-123"}`
	event.EventMetadata = `{"source": "test"}`
	event.IdempotencyKey = "test-key-123"

	// Execute
	err := repo.Create(ctx, event)

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, event.ID)
	assert.NotZero(t, event.CreatedAt)
	assert.NotZero(t, event.UpdatedAt)

	// Verify in database
	var dbEvent outbox.OutboxEvent
	err = db.Where("id = ?", event.ID).First(&dbEvent).Error
	require.NoError(t, err)
	assert.Equal(t, event.EventType, dbEvent.EventType)
	assert.Equal(t, event.AggregateID, dbEvent.AggregateID)
	assert.Equal(t, event.IdempotencyKey, dbEvent.IdempotencyKey)
}

func TestOutboxRepository_Update(t *testing.T) {
	// Setup test database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, db)

	repo := events.NewOutboxRepository(db)
	ctx := context.Background()

	// Create initial event
	event := outbox.NewOutboxEvent("OrderCreated", "order", "order-123", nil)
	event.EventData = `{"order_id": "order-123"}`
	event.IdempotencyKey = "test-key-123"

	err := repo.Create(ctx, event)
	require.NoError(t, err)

	// Update event
	now := time.Now()
	event.PublishedAt = &now
	event.FailedAttempts = 1
	event.LastError = "test error"

	// Execute
	err = repo.Update(ctx, event)

	// Assert
	require.NoError(t, err)

	// Verify in database
	var dbEvent outbox.OutboxEvent
	err = db.Where("id = ?", event.ID).First(&dbEvent).Error
	require.NoError(t, err)
	assert.NotNil(t, dbEvent.PublishedAt)
	assert.Equal(t, 1, dbEvent.FailedAttempts)
	assert.Equal(t, "test error", dbEvent.LastError)
}

func TestOutboxRepository_GetByID(t *testing.T) {
	// Setup test database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, db)

	repo := events.NewOutboxRepository(db)
	ctx := context.Background()

	// Create test event
	event := outbox.NewOutboxEvent("OrderCreated", "order", "order-123", nil)
	event.EventData = `{"order_id": "order-123"}`
	event.IdempotencyKey = "test-key-123"

	err := repo.Create(ctx, event)
	require.NoError(t, err)

	// Execute
	retrievedEvent, err := repo.GetByID(ctx, event.ID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, retrievedEvent)
	assert.Equal(t, event.ID, retrievedEvent.ID)
	assert.Equal(t, event.EventType, retrievedEvent.EventType)
	assert.Equal(t, event.AggregateID, retrievedEvent.AggregateID)
}

func TestOutboxRepository_GetByID_NotFound(t *testing.T) {
	// Setup test database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, db)

	repo := events.NewOutboxRepository(db)
	ctx := context.Background()

	// Execute
	retrievedEvent, err := repo.GetByID(ctx, "non-existent-id")

	// Assert
	require.Error(t, err)
	assert.Nil(t, retrievedEvent)
	assert.Contains(t, err.Error(), "outbox event not found")
}

func TestOutboxRepository_GetUnpublished(t *testing.T) {
	// Setup test database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, db)

	repo := events.NewOutboxRepository(db)
	ctx := context.Background()

	// Create test events
	unpublishedEvent1 := outbox.NewOutboxEvent("OrderCreated", "order", "order-1", nil)
	unpublishedEvent1.EventData = `{"order_id": "order-1"}`
	unpublishedEvent1.IdempotencyKey = "key-1"

	time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	unpublishedEvent2 := outbox.NewOutboxEvent("OrderCreated", "order", "order-2", nil)
	unpublishedEvent2.EventData = `{"order_id": "order-2"}`
	unpublishedEvent2.IdempotencyKey = "key-2"

	time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	publishedEvent := outbox.NewOutboxEvent("OrderCreated", "order", "order-3", nil)
	publishedEvent.EventData = `{"order_id": "order-3"}`
	publishedEvent.IdempotencyKey = "key-3"

	// Create events
	err := repo.Create(ctx, unpublishedEvent1)
	require.NoError(t, err)
	err = repo.Create(ctx, unpublishedEvent2)
	require.NoError(t, err)
	err = repo.Create(ctx, publishedEvent)
	require.NoError(t, err)

	// Mark one as published
	now := time.Now()
	publishedEvent.PublishedAt = &now
	err = repo.Update(ctx, publishedEvent)
	require.NoError(t, err)

	// Execute
	unpublishedEvents, err := repo.GetUnpublished(ctx, 10)

	// Assert
	require.NoError(t, err)
	assert.Len(t, unpublishedEvents, 2)

	// Verify only unpublished events are returned
	for _, event := range unpublishedEvents {
		assert.Nil(t, event.PublishedAt)
	}
}

func TestOutboxRepository_GetUnpublished_WithLimit(t *testing.T) {
	// Setup test database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, db)

	repo := events.NewOutboxRepository(db)
	ctx := context.Background()

	// Create multiple unpublished events
	for i := 0; i < 5; i++ {
		event := outbox.NewOutboxEvent("OrderCreated", "order", fmt.Sprintf("order-%d", i), nil)
		event.EventData = fmt.Sprintf(`{"order_id": "order-%d"}`, i)
		event.IdempotencyKey = fmt.Sprintf("key-%d", i)
		err := repo.Create(ctx, event)
		require.NoError(t, err)
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	}

	// Execute with limit
	unpublishedEvents, err := repo.GetUnpublished(ctx, 3)

	// Assert
	require.NoError(t, err)
	assert.Len(t, unpublishedEvents, 3)
}

func TestOutboxRepository_GetByIdempotencyKey(t *testing.T) {
	// Setup test database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, db)

	repo := events.NewOutboxRepository(db)
	ctx := context.Background()

	// Create test event
	event := outbox.NewOutboxEvent("OrderCreated", "order", "order-123", nil)
	event.EventData = `{"order_id": "order-123"}`
	event.IdempotencyKey = "unique-key-123"

	err := repo.Create(ctx, event)
	require.NoError(t, err)

	// Execute
	retrievedEvent, err := repo.GetByIdempotencyKey(ctx, "unique-key-123")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, retrievedEvent)
	assert.Equal(t, event.ID, retrievedEvent.ID)
	assert.Equal(t, event.IdempotencyKey, retrievedEvent.IdempotencyKey)
}

func TestOutboxRepository_GetByIdempotencyKey_NotFound(t *testing.T) {
	// Setup test database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, db)

	repo := events.NewOutboxRepository(db)
	ctx := context.Background()

	// Execute
	retrievedEvent, err := repo.GetByIdempotencyKey(ctx, "non-existent-key")

	// Assert
	require.NoError(t, err)
	assert.Nil(t, retrievedEvent) // Should return nil, not error for deduplication check
}

func TestOutboxRepository_DeletePublished(t *testing.T) {
	// Setup test database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, db)

	repo := events.NewOutboxRepository(db)
	ctx := context.Background()

	// Create old published event
	oldEvent := outbox.NewOutboxEvent("OrderCreated", "order", "order-old", nil)
	oldEvent.EventData = `{"order_id": "order-old"}`
	oldEvent.IdempotencyKey = "key-old"
	err := repo.Create(ctx, oldEvent)
	require.NoError(t, err)

	// Mark as published with old timestamp
	oldTime := time.Now().AddDate(0, 0, -10) // 10 days ago
	oldEvent.PublishedAt = &oldTime
	err = repo.Update(ctx, oldEvent)
	require.NoError(t, err)

	// Create recent published event
	time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	recentEvent := outbox.NewOutboxEvent("OrderCreated", "order", "order-recent", nil)
	recentEvent.EventData = `{"order_id": "order-recent"}`
	recentEvent.IdempotencyKey = "key-recent"
	err = repo.Create(ctx, recentEvent)
	require.NoError(t, err)

	// Mark as published with recent timestamp
	recentTime := time.Now().AddDate(0, 0, -1) // 1 day ago
	recentEvent.PublishedAt = &recentTime
	err = repo.Update(ctx, recentEvent)
	require.NoError(t, err)

	// Execute - delete events older than 7 days
	err = repo.DeletePublished(ctx, 7)

	// Assert
	require.NoError(t, err)

	// Verify old event is deleted
	var count int64
	err = db.Model(&outbox.OutboxEvent{}).Where("id = ?", oldEvent.ID).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Verify recent event still exists
	err = db.Model(&outbox.OutboxEvent{}).Where("id = ?", recentEvent.ID).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestOutboxRepository_GetEventsByAggregate(t *testing.T) {
	// Setup test database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, db)

	repo := events.NewOutboxRepository(db)
	ctx := context.Background()

	// Create events for different aggregates
	orderEvent1 := outbox.NewOutboxEvent("OrderCreated", "order", "order-123", nil)
	orderEvent1.EventData = `{"order_id": "order-123"}`
	orderEvent1.IdempotencyKey = "key-1"

	time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	orderEvent2 := outbox.NewOutboxEvent("OrderStatusUpdated", "order", "order-123", nil)
	orderEvent2.EventData = `{"order_id": "order-123"}`
	orderEvent2.IdempotencyKey = "key-2"

	time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	catalogEvent := outbox.NewOutboxEvent("CatalogItemCreated", "catalog_item", "item-456", nil)
	catalogEvent.EventData = `{"item_id": "item-456"}`
	catalogEvent.IdempotencyKey = "key-3"

	// Create events
	err := repo.Create(ctx, orderEvent1)
	require.NoError(t, err)
	err = repo.Create(ctx, orderEvent2)
	require.NoError(t, err)
	err = repo.Create(ctx, catalogEvent)
	require.NoError(t, err)

	// Execute
	orderEvents, err := repo.GetEventsByAggregate(ctx, "order", "order-123", 10)

	// Assert
	require.NoError(t, err)
	assert.Len(t, orderEvents, 2)

	// Verify all events are for the correct aggregate
	for _, event := range orderEvents {
		assert.Equal(t, "order", event.AggregateType)
		assert.Equal(t, "order-123", event.AggregateID)
	}
}

func TestOutboxRepository_GetEventsByAggregate_WithLimit(t *testing.T) {
	// Setup test database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, db)

	repo := events.NewOutboxRepository(db)
	ctx := context.Background()

	// Create multiple events for the same aggregate
	for i := 0; i < 5; i++ {
		event := outbox.NewOutboxEvent("OrderStatusUpdated", "order", "order-123", nil)
		event.EventData = fmt.Sprintf(`{"status": "status-%d"}`, i)
		event.IdempotencyKey = fmt.Sprintf("key-%d", i)
		err := repo.Create(ctx, event)
		require.NoError(t, err)
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	}

	// Execute with limit
	events, err := repo.GetEventsByAggregate(ctx, "order", "order-123", 3)

	// Assert
	require.NoError(t, err)
	assert.Len(t, events, 3)
}
