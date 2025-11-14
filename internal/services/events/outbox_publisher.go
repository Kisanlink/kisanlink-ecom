package events

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/sirupsen/logrus"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/catalog"
	"github.com/Kisanlink/kisanlink-ecom/entities/models/orders"
	"github.com/Kisanlink/kisanlink-ecom/entities/models/outbox"
	"github.com/Kisanlink/kisanlink-ecom/internal/repositories/events"
)

// OutboxPublisher handles reliable event publishing using the outbox pattern
type OutboxPublisher struct {
	outboxRepo   events.OutboxRepository
	sqsClient    SQSClient
	queueURL     string // Default queue URL (fallback)
	logger       *logrus.Logger
	eventFactory EventFactoryInterface
	eventRouter  EventRouter
}

// NewOutboxPublisher creates a new outbox publisher
func NewOutboxPublisher(
	outboxRepo events.OutboxRepository,
	sqsClient SQSClient,
	queueURL string,
	logger *logrus.Logger,
	eventFactory EventFactoryInterface,
	eventRouter EventRouter,
) *OutboxPublisher {
	return &OutboxPublisher{
		outboxRepo:   outboxRepo,
		sqsClient:    sqsClient,
		queueURL:     queueURL,
		logger:       logger,
		eventFactory: eventFactory,
		eventRouter:  eventRouter,
	}
}

// PublishEvent publishes an event to the outbox for reliable delivery
func (p *OutboxPublisher) PublishEvent(ctx context.Context, event *outbox.OutboxEvent) error {
	// First, store the event in the outbox table (transactional)
	if err := p.outboxRepo.Create(ctx, event); err != nil {
		p.logger.WithError(err).Error("Failed to store event in outbox")
		return fmt.Errorf("failed to store event in outbox: %w", err)
	}

	// Attempt immediate publishing
	if err := p.publishToSQS(ctx, event); err != nil {
		p.logger.WithError(err).WithField("event_id", event.ID).Warn("Failed to publish event immediately, will retry later")
		event.RecordFailure(err.Error())
		if updateErr := p.outboxRepo.Update(ctx, event); updateErr != nil {
			p.logger.WithError(updateErr).Error("Failed to update event failure status")
		}
		return nil // Don't return error as event is stored for retry
	}

	// Mark as published
	event.MarkAsPublished()
	if err := p.outboxRepo.Update(ctx, event); err != nil {
		p.logger.WithError(err).Error("Failed to mark event as published")
		return fmt.Errorf("failed to mark event as published: %w", err)
	}

	p.logger.WithField("event_id", event.ID).Info("Event published successfully")
	return nil
}

// ProcessUnpublishedEvents processes events that failed to publish
func (p *OutboxPublisher) ProcessUnpublishedEvents(ctx context.Context) error {
	events, err := p.outboxRepo.GetUnpublished(ctx, 100) // Process in batches
	if err != nil {
		return fmt.Errorf("failed to get unpublished events: %w", err)
	}

	for _, event := range events {
		if !event.ShouldRetry() {
			p.logger.WithField("event_id", event.ID).Warn("Event exceeded max retry attempts")
			continue
		}

		// Calculate exponential backoff delay
		delay := p.calculateBackoffDelay(event.FailedAttempts)
		if time.Since(event.UpdatedAt) < delay {
			continue // Not ready for retry yet
		}

		if err := p.publishToSQS(ctx, event); err != nil {
			p.logger.WithError(err).WithField("event_id", event.ID).Error("Failed to retry event publishing")
			event.RecordFailure(err.Error())
			if updateErr := p.outboxRepo.Update(ctx, event); updateErr != nil {
				p.logger.WithError(updateErr).Error("Failed to update event failure status")
			}
			continue
		}

		// Mark as published
		event.MarkAsPublished()
		if err := p.outboxRepo.Update(ctx, event); err != nil {
			p.logger.WithError(err).Error("Failed to mark event as published")
			continue
		}

		p.logger.WithField("event_id", event.ID).Info("Event retry published successfully")
	}

	return nil
}

// publishToSQS publishes an event to Amazon SQS
func (p *OutboxPublisher) publishToSQS(ctx context.Context, event *outbox.OutboxEvent) error {
	// Get recipients for this event
	recipients, err := p.eventRouter.GetRecipients(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to get event recipients: %w", err)
	}

	// If no recipients found, use default queue
	if len(recipients) == 0 {
		recipients = []Recipient{
			{
				Type:     RecipientTypeService,
				ID:       "default",
				QueueURL: p.queueURL,
			},
		}
	}

	// Serialize event data once
	eventPayload, err := p.serializeEvent(event)
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// Send to all recipients
	var publishErrors []error
	successCount := 0

	for _, recipient := range recipients {
		if err := p.publishToRecipient(ctx, event, eventPayload, recipient); err != nil {
			p.logger.WithError(err).WithFields(logrus.Fields{
				"event_id":       event.ID,
				"recipient_type": recipient.Type,
				"recipient_id":   recipient.ID,
			}).Error("Failed to publish to recipient")
			publishErrors = append(publishErrors, err)
		} else {
			successCount++
		}
	}

	// Consider the publish successful if at least one recipient received it
	if successCount > 0 {
		p.logger.WithFields(logrus.Fields{
			"event_id":         event.ID,
			"success_count":    successCount,
			"total_recipients": len(recipients),
		}).Info("Event published to recipients")
		return nil
	}

	// All recipients failed
	return fmt.Errorf("failed to publish to any recipients: %v", publishErrors)
}

// publishToRecipient publishes an event to a specific recipient
func (p *OutboxPublisher) publishToRecipient(ctx context.Context, event *outbox.OutboxEvent, eventPayload string, recipient Recipient) error {
	// Determine the target queue URL
	queueURL := recipient.QueueURL
	if queueURL == "" {
		queueURL = p.queueURL // Fallback to default
	}

	// Create SQS message with recipient information
	message := &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(eventPayload),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"EventType": {
				DataType:    aws.String("String"),
				StringValue: aws.String(event.EventType),
			},
			"EventVersion": {
				DataType:    aws.String("String"),
				StringValue: aws.String(event.EventVersion),
			},
			"AggregateType": {
				DataType:    aws.String("String"),
				StringValue: aws.String(event.AggregateType),
			},
			"AggregateID": {
				DataType:    aws.String("String"),
				StringValue: aws.String(event.AggregateID),
			},
			"IdempotencyKey": {
				DataType:    aws.String("String"),
				StringValue: aws.String(event.IdempotencyKey),
			},
			"RecipientType": {
				DataType:    aws.String("String"),
				StringValue: aws.String(string(recipient.Type)),
			},
			"RecipientID": {
				DataType:    aws.String("String"),
				StringValue: aws.String(recipient.ID),
			},
		},
		MessageDeduplicationId: aws.String(fmt.Sprintf("%s:%s", event.IdempotencyKey, recipient.ID)), // For FIFO queues
		MessageGroupId:         aws.String(event.AggregateType),                                      // For FIFO queues
	}

	// Send message to SQS
	_, err := p.sqsClient.SendMessage(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send message to recipient %s (%s): %w", recipient.ID, queueURL, err)
	}

	return nil
}

// serializeEvent serializes an outbox event for publishing
func (p *OutboxPublisher) serializeEvent(event *outbox.OutboxEvent) (string, error) {
	eventEnvelope := EventEnvelope{
		EventID:        event.ID,
		EventType:      event.EventType,
		EventVersion:   event.EventVersion,
		AggregateType:  event.AggregateType,
		AggregateID:    event.AggregateID,
		IdempotencyKey: event.IdempotencyKey,
		Timestamp:      event.CreatedAt,
		Data:           json.RawMessage(event.EventData),
		Metadata:       json.RawMessage(event.EventMetadata),
	}

	serialized, err := json.Marshal(eventEnvelope)
	if err != nil {
		return "", fmt.Errorf("failed to marshal event envelope: %w", err)
	}

	return string(serialized), nil
}

// calculateBackoffDelay calculates exponential backoff delay
func (p *OutboxPublisher) calculateBackoffDelay(attempts int) time.Duration {
	// Exponential backoff: 2^attempts seconds, max 5 minutes
	delay := time.Duration(math.Pow(2, float64(attempts))) * time.Second
	maxDelay := 5 * time.Minute
	if delay > maxDelay {
		delay = maxDelay
	}
	return delay
}

// EventEnvelope represents the structure of events published to the federated network
type EventEnvelope struct {
	EventID        string          `json:"event_id"`
	EventType      string          `json:"event_type"`
	EventVersion   string          `json:"event_version"`
	AggregateType  string          `json:"aggregate_type"`
	AggregateID    string          `json:"aggregate_id"`
	IdempotencyKey string          `json:"idempotency_key"`
	Timestamp      time.Time       `json:"timestamp"`
	Data           json.RawMessage `json:"data"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
}

// PublishOrderCreated publishes an order created event
func (p *OutboxPublisher) PublishOrderCreated(ctx context.Context, order *orders.Order) error {
	event, err := p.eventFactory.CreateOrderCreatedEvent(order)
	if err != nil {
		return fmt.Errorf("failed to create order created event: %w", err)
	}

	return p.PublishEvent(ctx, event)
}

// PublishOrderStatusUpdated publishes an order status updated event
func (p *OutboxPublisher) PublishOrderStatusUpdated(ctx context.Context, order *orders.Order, previousStatus string) error {
	event, err := p.eventFactory.CreateOrderStatusUpdatedEvent(order, previousStatus)
	if err != nil {
		return fmt.Errorf("failed to create order status updated event: %w", err)
	}

	return p.PublishEvent(ctx, event)
}

// PublishCatalogItemCreated publishes a catalog item created event
func (p *OutboxPublisher) PublishCatalogItemCreated(ctx context.Context, catalogItem *catalog.CatalogItem) error {
	event, err := p.eventFactory.CreateCatalogItemCreatedEvent(catalogItem)
	if err != nil {
		return fmt.Errorf("failed to create catalog item created event: %w", err)
	}

	return p.PublishEvent(ctx, event)
}

// PublishCatalogItemUpdated publishes a catalog item updated event
func (p *OutboxPublisher) PublishCatalogItemUpdated(ctx context.Context, catalogItem *catalog.CatalogItem) error {
	event, err := p.eventFactory.CreateCatalogItemUpdatedEvent(catalogItem)
	if err != nil {
		return fmt.Errorf("failed to create catalog item updated event: %w", err)
	}

	return p.PublishEvent(ctx, event)
}

// PublishInventoryAdjusted publishes an inventory adjusted event
func (p *OutboxPublisher) PublishInventoryAdjusted(ctx context.Context, lotID, catalogItemID string, previousQuantity, newQuantity float64, reason string) error {
	event, err := p.eventFactory.CreateInventoryAdjustedEvent(lotID, catalogItemID, previousQuantity, newQuantity, reason)
	if err != nil {
		return fmt.Errorf("failed to create inventory adjusted event: %w", err)
	}

	return p.PublishEvent(ctx, event)
}
