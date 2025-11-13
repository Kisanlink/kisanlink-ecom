package marketplace

import (
	"context"
	"time"
)

// NotificationServiceInterface defines the interface for external notification services
type NotificationServiceInterface interface {
	// Basic notification methods
	SendNotification(ctx context.Context, request NotificationRequest) error
	SendBatchNotifications(ctx context.Context, requests []NotificationRequest) error

	// Auction-specific notification methods
	NotifyBidPlaced(ctx context.Context, bid interface{}, listing interface{}) error
	NotifyBidOutbid(ctx context.Context, outbidBid interface{}, newBid interface{}, listing interface{}) error
	NotifyAutoBidTriggered(ctx context.Context, autoBid interface{}, triggeringBid interface{}, listing interface{}) error
	NotifyAuctionWon(ctx context.Context, listingID string, userID string, amount string) error
	NotifyAuctionExpired(ctx context.Context, listingID string, sellerID string, winnerID string) error
}

// NotificationRequest represents a notification request
type NotificationRequest struct {
	Type      string                 `json:"type"`
	Recipient string                 `json:"recipient"`
	Subject   string                 `json:"subject"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Priority  string                 `json:"priority"`
	Channel   string                 `json:"channel,omitempty"`
}

// WebSocketNotificationService defines the interface for WebSocket-based real-time notifications
type WebSocketNotificationService interface {
	// WebSocket connection management
	HandleWebSocketConnection(userID string, conn interface{}) error
	CloseConnection(userID string) error

	// Real-time broadcasting
	BroadcastToUser(userID string, notification interface{}) error
	BroadcastToListing(listingID string, notification interface{}) error
	BroadcastToAll(notification interface{}) error

	// Subscription management
	SubscribeToListing(userID string, listingID string) error
	UnsubscribeFromListing(userID string, listingID string) error

	// Connection status
	IsUserConnected(userID string) bool
	GetConnectedUsers() []string
}

// NotificationPreferencesService defines the interface for managing user notification preferences
type NotificationPreferencesService interface {
	GetPreferences(ctx context.Context, userID string) (*NotificationPreferences, error)
	UpdatePreferences(ctx context.Context, userID string, preferences *NotificationPreferences) error
	GetDefaultPreferences() *NotificationPreferences
}

// NotificationPreferences represents user notification preferences
type NotificationPreferences struct {
	UserID           string          `json:"user_id"`
	WebSocketEnabled bool            `json:"websocket_enabled"`
	EmailEnabled     bool            `json:"email_enabled"`
	SMSEnabled       bool            `json:"sms_enabled"`
	PushEnabled      bool            `json:"push_enabled"`
	EventPreferences map[string]bool `json:"event_preferences"`
	QuietHours       *QuietHours     `json:"quiet_hours,omitempty"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// QuietHours represents time periods when notifications should be suppressed
type QuietHours struct {
	Enabled   bool      `json:"enabled"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Timezone  string    `json:"timezone"`
}

// RealTimeNotificationEvent represents a real-time notification event
type RealTimeNotificationEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	ListingID string                 `json:"listing_id,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Priority  string                 `json:"priority"`
}
