package marketplace

import (
	"kisanlink-ecom/internal/common"
	"kisanlink-ecom/internal/services/marketplace"

	"github.com/gin-gonic/gin"
)

// NotificationHandler handles notification-related HTTP requests
type NotificationHandler struct {
	notificationService marketplace.RealTimeNotificationService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notificationService marketplace.RealTimeNotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// HandleWebSocket handles WebSocket connection requests for real-time notifications
func (h *NotificationHandler) HandleWebSocket(c *gin.Context) {
	h.notificationService.HandleWebSocketConnection(c)
}

// GetNotificationPreferences retrieves user notification preferences
func (h *NotificationHandler) GetNotificationPreferences(c *gin.Context) {
	// Extract user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	// Get notification preferences
	preferences, err := h.notificationService.GetUserNotificationPreferences(c.Request.Context(), userID.(string))
	if err != nil {
		common.InternalServerError(c, "PREFERENCES_RETRIEVAL_FAILED", "Failed to get notification preferences", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, preferences, nil)
}

// UpdateNotificationPreferences updates user notification preferences
func (h *NotificationHandler) UpdateNotificationPreferences(c *gin.Context) {
	// Extract user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	// Parse request body
	var preferences marketplace.NotificationPreferences
	if err := c.ShouldBindJSON(&preferences); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Set user ID
	preferences.UserID = userID.(string)

	// Update notification preferences
	if err := h.notificationService.UpdateUserNotificationPreferences(c.Request.Context(), userID.(string), &preferences); err != nil {
		common.InternalServerError(c, "PREFERENCES_UPDATE_FAILED", "Failed to update notification preferences", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, preferences, nil)
}

// TestNotification sends a test notification to the user (for testing purposes)
func (h *NotificationHandler) TestNotification(c *gin.Context) {
	// Extract user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	// Create test notification
	testNotification := &marketplace.RealTimeNotificationEvent{
		ID:       "test_notification",
		Type:     "TEST",
		Title:    "Test Notification",
		Message:  "This is a test notification to verify your WebSocket connection",
		Priority: "LOW",
	}

	// Send test notification
	if err := h.notificationService.BroadcastToUser(userID.(string), testNotification); err != nil {
		common.InternalServerError(c, "NOTIFICATION_SEND_FAILED", "Failed to send test notification", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, nil, nil)
}
