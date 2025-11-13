package notifications

import (
	"context"
	"fmt"
	"log"
	"time"
)

// NotificationChannel represents the type of notification channel
type NotificationChannel string

const (
	ChannelEmail   NotificationChannel = "EMAIL"
	ChannelSMS     NotificationChannel = "SMS"
	ChannelWebhook NotificationChannel = "WEBHOOK"
)

// NotificationPayload represents the data for a notification
type NotificationPayload struct {
	RecipientID string                 `json:"recipient_id"`
	Subject     string                 `json:"subject"`
	Message     string                 `json:"message"`
	Channel     NotificationChannel    `json:"channel"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NotificationService defines the interface for notification operations
type NotificationService interface {
	SendNotification(ctx context.Context, payload *NotificationPayload) error
	SendBulkNotifications(ctx context.Context, payloads []*NotificationPayload) error
}

// notificationService implements the NotificationService interface
type notificationService struct {
	// Future: Add email, SMS, webhook clients here
}

// NewNotificationService creates a new notification service
func NewNotificationService() NotificationService {
	return &notificationService{}
}

// SendNotification sends a single notification
// MVP implementation: Logs notifications to stdout
func (s *notificationService) SendNotification(ctx context.Context, payload *NotificationPayload) error {
	if payload == nil {
		return fmt.Errorf("notification payload cannot be nil")
	}

	// Validate payload
	if payload.RecipientID == "" {
		return fmt.Errorf("recipient ID is required")
	}
	if payload.Message == "" {
		return fmt.Errorf("message is required")
	}

	// MVP: Log notification instead of sending
	log.Printf("[NOTIFICATION] Channel: %s | Recipient: %s | Subject: %s | Message: %s | Metadata: %v",
		payload.Channel,
		payload.RecipientID,
		payload.Subject,
		payload.Message,
		payload.Metadata,
	)

	// Future implementation will send actual notifications based on channel
	switch payload.Channel {
	case ChannelEmail:
		// TODO: Implement email sending
		log.Printf("[EMAIL] To: %s | Subject: %s", payload.RecipientID, payload.Subject)
	case ChannelSMS:
		// TODO: Implement SMS sending
		log.Printf("[SMS] To: %s | Message: %s", payload.RecipientID, payload.Message)
	case ChannelWebhook:
		// TODO: Implement webhook posting
		log.Printf("[WEBHOOK] URL: %s | Payload: %v", payload.RecipientID, payload.Metadata)
	default:
		return fmt.Errorf("unsupported notification channel: %s", payload.Channel)
	}

	return nil
}

// SendBulkNotifications sends multiple notifications
func (s *notificationService) SendBulkNotifications(ctx context.Context, payloads []*NotificationPayload) error {
	if len(payloads) == 0 {
		return nil
	}

	log.Printf("[BULK_NOTIFICATION] Sending %d notifications", len(payloads))

	var errors []error
	for i, payload := range payloads {
		if err := s.SendNotification(ctx, payload); err != nil {
			log.Printf("[BULK_NOTIFICATION] Failed to send notification %d: %v", i, err)
			errors = append(errors, err)
		}
		// Small delay to avoid overwhelming the system (MVP)
		time.Sleep(10 * time.Millisecond)
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send %d out of %d notifications", len(errors), len(payloads))
	}

	log.Printf("[BULK_NOTIFICATION] Successfully sent %d notifications", len(payloads))
	return nil
}

// Helper functions for creating notification payloads

// NewEmailNotification creates an email notification payload
func NewEmailNotification(recipientEmail, subject, message string, metadata map[string]interface{}) *NotificationPayload {
	return &NotificationPayload{
		RecipientID: recipientEmail,
		Subject:     subject,
		Message:     message,
		Channel:     ChannelEmail,
		Metadata:    metadata,
	}
}

// NewSMSNotification creates an SMS notification payload
func NewSMSNotification(phoneNumber, message string, metadata map[string]interface{}) *NotificationPayload {
	return &NotificationPayload{
		RecipientID: phoneNumber,
		Subject:     "",
		Message:     message,
		Channel:     ChannelSMS,
		Metadata:    metadata,
	}
}

// NewWebhookNotification creates a webhook notification payload
func NewWebhookNotification(webhookURL, message string, metadata map[string]interface{}) *NotificationPayload {
	return &NotificationPayload{
		RecipientID: webhookURL,
		Subject:     "",
		Message:     message,
		Channel:     ChannelWebhook,
		Metadata:    metadata,
	}
}
