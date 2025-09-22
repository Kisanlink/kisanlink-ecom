package marketplace

import (
	"net/http"

	"kisanlink-ecom/internal/services/marketplace"
	"kisanlink-ecom/internal/utils"

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
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", "")
		return
	}

	// Get notification preferences
	preferences, err := h.notificationService.GetUserNotificationPreferences(c.Request.Context(), userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "PREFERENCES_RETRIEVAL_FAILED", "Failed to get notification preferences", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Notification preferences retrieved successfully", preferences)
}

// UpdateNotificationPreferences updates user notification preferences
func (h *NotificationHandler) UpdateNotificationPreferences(c *gin.Context) {
	// Extract user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", "")
		return
	}

	// Parse request body
	var preferences marketplace.NotificationPreferences
	if err := c.ShouldBindJSON(&preferences); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err.Error())
		return
	}

	// Set user ID
	preferences.UserID = userID.(string)

	// Update notification preferences
	if err := h.notificationService.UpdateUserNotificationPreferences(c.Request.Context(), userID.(string), &preferences); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "PREFERENCES_UPDATE_FAILED", "Failed to update notification preferences", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Notification preferences updated successfully", preferences)
}

// TestNotification sends a test notification to the user (for testing purposes)
func (h *NotificationHandler) TestNotification(c *gin.Context) {
	// Extract user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", "")
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
		utils.ErrorResponse(c, http.StatusInternalServerError, "NOTIFICATION_SEND_FAILED", "Failed to send test notification", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Test notification sent successfully", nil)
}
