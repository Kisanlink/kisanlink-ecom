package events

import (
	"context"
	"kisanlink-ecom/internal/repositories/events"
	"time"

	"github.com/sirupsen/logrus"
)

// EventProcessor handles background processing of unpublished events
type EventProcessor struct {
	publisher EventPublisher
	config    *EventPublishingConfig
	logger    *logrus.Logger
	stopCh    chan struct{}
}

// NewEventProcessor creates a new event processor
func NewEventProcessor(
	publisher EventPublisher,
	config *EventPublishingConfig,
	logger *logrus.Logger,
) *EventProcessor {
	return &EventProcessor{
		publisher: publisher,
		config:    config,
		logger:    logger,
		stopCh:    make(chan struct{}),
	}
}

// Start begins the background processing of unpublished events
func (p *EventProcessor) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(p.config.ProcessingInterval) * time.Second)
	defer ticker.Stop()

	p.logger.Info("Event processor started")

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("Event processor stopped due to context cancellation")
			return
		case <-p.stopCh:
			p.logger.Info("Event processor stopped")
			return
		case <-ticker.C:
			if err := p.processUnpublishedEvents(ctx); err != nil {
				p.logger.WithError(err).Error("Failed to process unpublished events")
			}
		}
	}
}

// Stop stops the event processor
func (p *EventProcessor) Stop() {
	close(p.stopCh)
}

// processUnpublishedEvents processes a batch of unpublished events
func (p *EventProcessor) processUnpublishedEvents(ctx context.Context) error {
	p.logger.Debug("Processing unpublished events")

	if err := p.publisher.ProcessUnpublishedEvents(ctx); err != nil {
		return err
	}

	return nil
}

// StartCleanupWorker starts a background worker to clean up old published events
func (p *EventProcessor) StartCleanupWorker(ctx context.Context, outboxRepo events.OutboxRepository) {
	ticker := time.NewTicker(24 * time.Hour) // Run cleanup daily
	defer ticker.Stop()

	p.logger.Info("Event cleanup worker started")

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("Event cleanup worker stopped due to context cancellation")
			return
		case <-p.stopCh:
			p.logger.Info("Event cleanup worker stopped")
			return
		case <-ticker.C:
			if err := p.cleanupOldEvents(ctx, outboxRepo); err != nil {
				p.logger.WithError(err).Error("Failed to cleanup old events")
			}
		}
	}
}

// cleanupOldEvents removes old published events
func (p *EventProcessor) cleanupOldEvents(ctx context.Context, outboxRepo events.OutboxRepository) error {
	p.logger.Debug("Cleaning up old published events")

	if err := outboxRepo.DeletePublished(ctx, p.config.CleanupIntervalDays); err != nil {
		return err
	}

	p.logger.Info("Old published events cleaned up successfully")
	return nil
}
