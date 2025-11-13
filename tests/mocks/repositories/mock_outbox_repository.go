package repositories

import (
	"context"

	"github.com/stretchr/testify/mock"

	"kisanlink-ecom/entities/models/outbox"
)

// MockOutboxRepository is a mock implementation of OutboxRepository
type MockOutboxRepository struct {
	mock.Mock
}

// Create mocks the Create method
func (m *MockOutboxRepository) Create(ctx context.Context, event *outbox.OutboxEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// Update mocks the Update method
func (m *MockOutboxRepository) Update(ctx context.Context, event *outbox.OutboxEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// GetByID mocks the GetByID method
func (m *MockOutboxRepository) GetByID(ctx context.Context, id string) (*outbox.OutboxEvent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*outbox.OutboxEvent), args.Error(1)
}

// GetUnpublished mocks the GetUnpublished method
func (m *MockOutboxRepository) GetUnpublished(ctx context.Context, limit int) ([]*outbox.OutboxEvent, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*outbox.OutboxEvent), args.Error(1)
}

// GetByIdempotencyKey mocks the GetByIdempotencyKey method
func (m *MockOutboxRepository) GetByIdempotencyKey(ctx context.Context, key string) (*outbox.OutboxEvent, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*outbox.OutboxEvent), args.Error(1)
}

// DeletePublished mocks the DeletePublished method
func (m *MockOutboxRepository) DeletePublished(ctx context.Context, olderThanDays int) error {
	args := m.Called(ctx, olderThanDays)
	return args.Error(0)
}

// GetEventsByAggregate mocks the GetEventsByAggregate method
func (m *MockOutboxRepository) GetEventsByAggregate(ctx context.Context, aggregateType, aggregateID string, limit int) ([]*outbox.OutboxEvent, error) {
	args := m.Called(ctx, aggregateType, aggregateID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*outbox.OutboxEvent), args.Error(1)
}
