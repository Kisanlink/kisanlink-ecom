package services

import (
	"context"

	"github.com/stretchr/testify/mock"

	"kisanlink-ecom/entities/models/outbox"
	"kisanlink-ecom/internal/services/events"
)

// MockEventRouter is a mock implementation of EventRouter
type MockEventRouter struct {
	mock.Mock
}

// Ensure MockEventRouter implements the EventRouter interface
var _ events.EventRouter = (*MockEventRouter)(nil)

// GetRecipients mocks the GetRecipients method
func (m *MockEventRouter) GetRecipients(ctx context.Context, event *outbox.OutboxEvent) ([]events.Recipient, error) {
	args := m.Called(ctx, event)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]events.Recipient), args.Error(1)
}

// AddRoutingRule mocks the AddRoutingRule method
func (m *MockEventRouter) AddRoutingRule(rule events.RoutingRule) error {
	args := m.Called(rule)
	return args.Error(0)
}

// RemoveRoutingRule mocks the RemoveRoutingRule method
func (m *MockEventRouter) RemoveRoutingRule(eventType, aggregateType string) error {
	args := m.Called(eventType, aggregateType)
	return args.Error(0)
}
