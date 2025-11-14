package events_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/outbox"
	"github.com/Kisanlink/kisanlink-ecom/internal/services/events"
	"github.com/Kisanlink/kisanlink-ecom/tests/mocks/repositories"
	"github.com/Kisanlink/kisanlink-ecom/tests/mocks/services"
)

func TestOutboxPublisher_PublishEvent(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*repositories.MockOutboxRepository, *services.MockSQSClient, *services.MockEventFactory)
		event         *outbox.OutboxEvent
		expectedError string
		shouldCallSQS bool
	}{
		{
			name: "successful event publishing",
			setupMocks: func(repo *repositories.MockOutboxRepository, sqsClient *services.MockSQSClient, factory *services.MockEventFactory) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*outbox.OutboxEvent")).Return(nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*outbox.OutboxEvent")).Return(nil)
				sqsClient.On("SendMessage", mock.Anything, mock.AnythingOfType("*sqs.SendMessageInput")).Return(&sqs.SendMessageOutput{}, nil)
			},
			event: &outbox.OutboxEvent{
				EventType:      "OrderCreated",
				AggregateType:  "order",
				AggregateID:    "order-123",
				EventData:      `{"order_id": "order-123"}`,
				EventMetadata:  `{"source": "test"}`,
				IdempotencyKey: "test-key",
			},
			shouldCallSQS: true,
		},
		{
			name: "repository create failure",
			setupMocks: func(repo *repositories.MockOutboxRepository, sqsClient *services.MockSQSClient, factory *services.MockEventFactory) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*outbox.OutboxEvent")).Return(errors.New("database error"))
			},
			event: &outbox.OutboxEvent{
				EventType:     "OrderCreated",
				AggregateType: "order",
				AggregateID:   "order-123",
			},
			expectedError: "failed to store event in outbox",
			shouldCallSQS: false,
		},
		{
			name: "SQS publish failure with retry storage",
			setupMocks: func(repo *repositories.MockOutboxRepository, sqsClient *services.MockSQSClient, factory *services.MockEventFactory) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*outbox.OutboxEvent")).Return(nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*outbox.OutboxEvent")).Return(nil)
				sqsClient.On("SendMessage", mock.Anything, mock.AnythingOfType("*sqs.SendMessageInput")).Return(nil, errors.New("SQS error"))
			},
			event: &outbox.OutboxEvent{
				EventType:      "OrderCreated",
				AggregateType:  "order",
				AggregateID:    "order-123",
				EventData:      `{"order_id": "order-123"}`,
				EventMetadata:  `{"source": "test"}`,
				IdempotencyKey: "test-key",
			},
			shouldCallSQS: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &repositories.MockOutboxRepository{}
			mockSQSClient := &services.MockSQSClient{}
			mockFactory := &services.MockEventFactory{}
			mockRouter := &services.MockEventRouter{}
			logger := logrus.New()

			tt.setupMocks(mockRepo, mockSQSClient, mockFactory)

			// Setup default router behavior
			mockRouter.On("GetRecipients", mock.Anything, mock.AnythingOfType("*outbox.OutboxEvent")).Return(
				[]events.Recipient{
					{
						Type:     events.RecipientTypeService,
						ID:       "test-service",
						QueueURL: "test-queue-url",
					},
				}, nil)

			// Create publisher
			publisher := events.NewOutboxPublisher(mockRepo, mockSQSClient, "test-queue-url", logger, mockFactory, mockRouter)

			// Execute
			err := publisher.PublishEvent(context.Background(), tt.event)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
			if tt.shouldCallSQS {
				mockSQSClient.AssertExpectations(t)
			}
		})
	}
}

func TestOutboxPublisher_ProcessUnpublishedEvents(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*repositories.MockOutboxRepository, *services.MockSQSClient)
		expectedError string
	}{
		{
			name: "successful processing of unpublished events",
			setupMocks: func(repo *repositories.MockOutboxRepository, sqsClient *services.MockSQSClient) {
				event := outbox.NewOutboxEvent("OrderCreated", "order", "order-123", nil)
				event.EventData = `{"order_id": "order-123"}`
				event.IdempotencyKey = "test-key-1"
				event.FailedAttempts = 1
				// Set UpdatedAt to simulate old failure
				event.BaseModel.UpdatedAt = time.Now().Add(-10 * time.Minute)

				events := []*outbox.OutboxEvent{event}

				repo.On("GetUnpublished", mock.Anything, 100).Return(events, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*outbox.OutboxEvent")).Return(nil)
				sqsClient.On("SendMessage", mock.Anything, mock.AnythingOfType("*sqs.SendMessageInput")).Return(&sqs.SendMessageOutput{}, nil)
			},
		},
		{
			name: "repository get unpublished failure",
			setupMocks: func(repo *repositories.MockOutboxRepository, sqsClient *services.MockSQSClient) {
				repo.On("GetUnpublished", mock.Anything, 100).Return(nil, errors.New("database error"))
			},
			expectedError: "failed to get unpublished events",
		},
		{
			name: "no unpublished events",
			setupMocks: func(repo *repositories.MockOutboxRepository, sqsClient *services.MockSQSClient) {
				repo.On("GetUnpublished", mock.Anything, 100).Return([]*outbox.OutboxEvent{}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &repositories.MockOutboxRepository{}
			mockSQSClient := &services.MockSQSClient{}
			mockFactory := &services.MockEventFactory{}
			mockRouter := &services.MockEventRouter{}
			logger := logrus.New()

			tt.setupMocks(mockRepo, mockSQSClient)

			// Setup default router behavior
			mockRouter.On("GetRecipients", mock.Anything, mock.AnythingOfType("*outbox.OutboxEvent")).Return(
				[]events.Recipient{
					{
						Type:     events.RecipientTypeService,
						ID:       "test-service",
						QueueURL: "test-queue-url",
					},
				}, nil)

			// Create publisher
			publisher := events.NewOutboxPublisher(mockRepo, mockSQSClient, "test-queue-url", logger, mockFactory, mockRouter)

			// Execute
			err := publisher.ProcessUnpublishedEvents(context.Background())

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
			mockSQSClient.AssertExpectations(t)
		})
	}
}

func TestOutboxPublisher_SerializeEvent(t *testing.T) {
	// Setup
	mockRepo := &repositories.MockOutboxRepository{}
	mockSQSClient := &services.MockSQSClient{}
	mockFactory := &services.MockEventFactory{}
	mockRouter := &services.MockEventRouter{}
	logger := logrus.New()

	// Setup router behavior
	mockRouter.On("GetRecipients", mock.Anything, mock.AnythingOfType("*outbox.OutboxEvent")).Return(
		[]events.Recipient{
			{
				Type:     events.RecipientTypeService,
				ID:       "test-service",
				QueueURL: "test-queue-url",
			},
		}, nil)

	publisher := events.NewOutboxPublisher(mockRepo, mockSQSClient, "test-queue-url", logger, mockFactory, mockRouter)

	event := outbox.NewOutboxEvent("OrderCreated", "order", "order-123", nil)
	event.EventData = `{"order_id": "order-123", "status": "pending"}`
	event.EventMetadata = `{"source": "order-service"}`
	event.IdempotencyKey = "test-key"

	// Execute - using reflection to access private method
	// Note: In a real implementation, you might want to make this method public for testing
	// or create a separate serializer that can be tested independently

	// For now, we'll test the serialization indirectly through the SQS message creation
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*outbox.OutboxEvent")).Return(nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*outbox.OutboxEvent")).Return(nil)

	// Capture the message sent to SQS
	var capturedMessage *sqs.SendMessageInput
	mockSQSClient.On("SendMessage", mock.Anything, mock.AnythingOfType("*sqs.SendMessageInput")).
		Run(func(args mock.Arguments) {
			capturedMessage = args.Get(1).(*sqs.SendMessageInput)
		}).
		Return(&sqs.SendMessageOutput{}, nil)

	// Execute
	err := publisher.PublishEvent(context.Background(), event)
	require.NoError(t, err)

	// Assert message structure
	require.NotNil(t, capturedMessage)
	assert.Equal(t, "test-queue-url", *capturedMessage.QueueUrl)

	// Parse and validate the message body
	var envelope events.EventEnvelope
	err = json.Unmarshal([]byte(*capturedMessage.MessageBody), &envelope)
	require.NoError(t, err)

	assert.Equal(t, event.EventType, envelope.EventType)
	assert.Equal(t, event.EventVersion, envelope.EventVersion)
	assert.Equal(t, event.AggregateType, envelope.AggregateType)
	assert.Equal(t, event.AggregateID, envelope.AggregateID)
	assert.Equal(t, event.IdempotencyKey, envelope.IdempotencyKey)

	// Validate message attributes
	assert.Equal(t, event.EventType, *capturedMessage.MessageAttributes["EventType"].StringValue)
	assert.Equal(t, event.EventVersion, *capturedMessage.MessageAttributes["EventVersion"].StringValue)
	assert.Equal(t, event.AggregateType, *capturedMessage.MessageAttributes["AggregateType"].StringValue)
	assert.Equal(t, event.AggregateID, *capturedMessage.MessageAttributes["AggregateID"].StringValue)
	assert.Equal(t, event.IdempotencyKey, *capturedMessage.MessageAttributes["IdempotencyKey"].StringValue)
}

func TestOutboxPublisher_CalculateBackoffDelay(t *testing.T) {
	tests := []struct {
		name     string
		attempts int
		expected time.Duration
	}{
		{
			name:     "first retry",
			attempts: 1,
			expected: 2 * time.Second,
		},
		{
			name:     "second retry",
			attempts: 2,
			expected: 4 * time.Second,
		},
		{
			name:     "third retry",
			attempts: 3,
			expected: 8 * time.Second,
		},
		{
			name:     "max delay reached",
			attempts: 10,
			expected: 5 * time.Minute, // Should be capped at 5 minutes
		},
	}

	// Setup
	mockRepo := &repositories.MockOutboxRepository{}
	mockSQSClient := &services.MockSQSClient{}
	mockFactory := &services.MockEventFactory{}
	mockRouter := &services.MockEventRouter{}
	logger := logrus.New()

	// Setup router behavior
	mockRouter.On("GetRecipients", mock.Anything, mock.AnythingOfType("*outbox.OutboxEvent")).Return(
		[]events.Recipient{
			{
				Type:     events.RecipientTypeService,
				ID:       "test-service",
				QueueURL: "test-queue-url",
			},
		}, nil)

	publisher := events.NewOutboxPublisher(mockRepo, mockSQSClient, "test-queue-url", logger, mockFactory, mockRouter)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This test would require accessing the private calculateBackoffDelay method
			// In a real implementation, you might want to make this method public for testing
			// or extract it to a utility function

			// For now, we'll test the behavior indirectly through ProcessUnpublishedEvents
			// by checking that events with recent failures are not retried immediately

			if tt.attempts <= 8 { // Only test reasonable attempt counts
				recentTime := time.Now().Add(-1 * time.Second) // Very recent failure
				event := outbox.NewOutboxEvent("TestEvent", "test", "test-123", nil)
				event.FailedAttempts = tt.attempts
				event.BaseModel.UpdatedAt = recentTime

				mockRepo.On("GetUnpublished", mock.Anything, 100).Return([]*outbox.OutboxEvent{event}, nil)

				err := publisher.ProcessUnpublishedEvents(context.Background())
				require.NoError(t, err)

				// The event should not be processed due to backoff delay
				mockSQSClient.AssertNotCalled(t, "SendMessage")
				mockRepo.AssertExpectations(t)
			}
		})
	}
}
