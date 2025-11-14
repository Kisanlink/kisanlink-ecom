package marketplace

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/marketplace"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// RealTimeNotificationService handles real-time WebSocket notifications for marketplace events
type RealTimeNotificationService interface {
	// WebSocket connections
	HandleWebSocketConnection(c *gin.Context)
	BroadcastToListing(listingID string, notification *RealTimeNotificationEvent) error
	BroadcastToUser(userID string, notification *RealTimeNotificationEvent) error

	// Event notifications
	NotifyBidPlaced(ctx context.Context, listing *marketplace.Listing, bid *marketplace.Bid, previousHighest *float64) error
	NotifyBidOutbid(ctx context.Context, listing *marketplace.Listing, outbidBid *marketplace.Bid, newHighestBid *marketplace.Bid) error
	NotifyAuctionClosing(ctx context.Context, listing *marketplace.Listing, minutesRemaining int) error
	NotifyAuctionClosed(ctx context.Context, listing *marketplace.Listing, winningBid *marketplace.Bid) error
	NotifyAutoBidTriggered(ctx context.Context, listing *marketplace.Listing, originalBid *marketplace.Bid, autoBid *marketplace.Bid) error

	// Email/SMS notifications
	SendEmailNotification(ctx context.Context, userID string, notification *EmailNotification) error
	SendSMSNotification(ctx context.Context, userID string, notification *SMSNotification) error

	// User preferences
	GetUserNotificationPreferences(ctx context.Context, userID string) (*NotificationPreferences, error)
	UpdateUserNotificationPreferences(ctx context.Context, userID string, preferences *NotificationPreferences) error
}

// EmailNotification represents an email notification
type EmailNotification struct {
	To       string                 `json:"to"`
	Subject  string                 `json:"subject"`
	Body     string                 `json:"body"`
	Template string                 `json:"template,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

// SMSNotification represents an SMS notification
type SMSNotification struct {
	To      string `json:"to"`
	Message string `json:"message"`
}

// WebSocketConnection represents a WebSocket connection
type WebSocketConnection struct {
	UserID     string
	ListingIDs []string
	Conn       *websocket.Conn
	Send       chan *RealTimeNotificationEvent
	Hub        *NotificationHub
}

// NotificationHub manages WebSocket connections
type NotificationHub struct {
	// Registered connections by user ID
	userConnections map[string]*WebSocketConnection

	// Registered connections by listing ID
	listingConnections map[string]map[string]*WebSocketConnection

	// Register requests from connections
	register chan *WebSocketConnection

	// Unregister requests from connections
	unregister chan *WebSocketConnection

	// Broadcast notifications
	broadcast chan *RealTimeNotificationEvent

	// Mutex for thread safety
	mutex sync.RWMutex
}

// realTimeNotificationService implements the RealTimeNotificationService interface
type realTimeNotificationService struct {
	hub *NotificationHub
	// In a real implementation, these would be injected dependencies
	// emailService EmailService
	// smsService   SMSService
	// userService  UserService
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// In production, implement proper origin checking
		return true
	},
}

// NewRealTimeNotificationService creates a new real-time notification service
func NewRealTimeNotificationService() RealTimeNotificationService {
	hub := &NotificationHub{
		userConnections:    make(map[string]*WebSocketConnection),
		listingConnections: make(map[string]map[string]*WebSocketConnection),
		register:           make(chan *WebSocketConnection),
		unregister:         make(chan *WebSocketConnection),
		broadcast:          make(chan *RealTimeNotificationEvent),
	}

	// Start the hub
	go hub.run()

	return &realTimeNotificationService{
		hub: hub,
	}
}

// HandleWebSocketConnection handles WebSocket connection requests
func (s *realTimeNotificationService) HandleWebSocketConnection(c *gin.Context) {
	// Extract user ID from context (should be set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket connection: %v", err)
		return
	}

	// Create WebSocket connection
	wsConn := &WebSocketConnection{
		UserID: userID.(string),
		Conn:   conn,
		Send:   make(chan *RealTimeNotificationEvent, 256),
		Hub:    s.hub,
	}

	// Register the connection
	s.hub.register <- wsConn

	// Start goroutines for reading and writing
	go wsConn.writePump()
	go wsConn.readPump()
}

// BroadcastToListing broadcasts a notification to all users watching a listing
func (s *realTimeNotificationService) BroadcastToListing(listingID string, notification *RealTimeNotificationEvent) error {
	notification.ListingID = listingID
	notification.Timestamp = time.Now()

	s.hub.mutex.RLock()
	connections, exists := s.hub.listingConnections[listingID]
	s.hub.mutex.RUnlock()

	if !exists {
		return nil // No connections for this listing
	}

	for _, conn := range connections {
		select {
		case conn.Send <- notification:
		default:
			// Connection is blocked, close it
			close(conn.Send)
			delete(connections, conn.UserID)
		}
	}

	return nil
}

// BroadcastToUser broadcasts a notification to a specific user
func (s *realTimeNotificationService) BroadcastToUser(userID string, notification *RealTimeNotificationEvent) error {
	notification.UserID = userID
	notification.Timestamp = time.Now()

	s.hub.mutex.RLock()
	conn, exists := s.hub.userConnections[userID]
	s.hub.mutex.RUnlock()

	if !exists {
		return nil // User not connected
	}

	select {
	case conn.Send <- notification:
	default:
		// Connection is blocked, close it
		close(conn.Send)
		delete(s.hub.userConnections, userID)
	}

	return nil
}

// NotifyBidPlaced sends notifications when a bid is placed
func (s *realTimeNotificationService) NotifyBidPlaced(ctx context.Context, listing *marketplace.Listing, bid *marketplace.Bid, previousHighest *float64) error {
	notification := &RealTimeNotificationEvent{
		ID:        fmt.Sprintf("bid_placed_%s_%d", bid.BidID, time.Now().Unix()),
		Type:      "BID_PLACED",
		ListingID: listing.ListingID,
		Title:     "New Bid Placed",
		Message:   fmt.Sprintf("A new bid of %s %.2f has been placed", bid.Currency, bid.BidAmount.InexactFloat64()),
		Priority:  "MEDIUM",
		Data: map[string]interface{}{
			"bid_id":     bid.BidID,
			"bid_amount": bid.BidAmount.InexactFloat64(),
			"currency":   bid.Currency,
			"bidder_id":  bid.BidderID,
		},
	}

	// Broadcast to all users watching this listing
	if err := s.BroadcastToListing(listing.ListingID, notification); err != nil {
		return fmt.Errorf("failed to broadcast bid placed notification: %w", err)
	}

	return nil
}

// NotifyBidOutbid sends notifications when a bid is outbid
func (s *realTimeNotificationService) NotifyBidOutbid(ctx context.Context, listing *marketplace.Listing, outbidBid *marketplace.Bid, newHighestBid *marketplace.Bid) error {
	notification := &RealTimeNotificationEvent{
		ID:        fmt.Sprintf("bid_outbid_%s_%d", outbidBid.BidID, time.Now().Unix()),
		Type:      "BID_OUTBID",
		ListingID: listing.ListingID,
		Title:     "Your Bid Has Been Outbid",
		Message: fmt.Sprintf("Your bid of %s %.2f has been outbid by a higher bid of %s %.2f",
			outbidBid.Currency, outbidBid.BidAmount.InexactFloat64(),
			newHighestBid.Currency, newHighestBid.BidAmount.InexactFloat64()),
		Priority: "HIGH",
		Data: map[string]interface{}{
			"outbid_bid_id":      outbidBid.BidID,
			"outbid_amount":      outbidBid.BidAmount.InexactFloat64(),
			"new_highest_amount": newHighestBid.BidAmount.InexactFloat64(),
		},
	}

	// Send notification to the outbid user
	if err := s.BroadcastToUser(outbidBid.BidderID, notification); err != nil {
		return fmt.Errorf("failed to broadcast outbid notification: %w", err)
	}

	return nil
}

// NotifyAuctionClosing sends notifications when an auction is about to close
func (s *realTimeNotificationService) NotifyAuctionClosing(ctx context.Context, listing *marketplace.Listing, minutesRemaining int) error {
	notification := &RealTimeNotificationEvent{
		ID:        fmt.Sprintf("auction_closing_%s_%d", listing.ListingID, time.Now().Unix()),
		Type:      "AUCTION_CLOSING",
		ListingID: listing.ListingID,
		Title:     "Auction Closing Soon",
		Message:   fmt.Sprintf("Auction for %s closes in %d minutes", listing.ProductID, minutesRemaining),
		Priority:  "HIGH",
		Data: map[string]interface{}{
			"minutes_remaining": minutesRemaining,
			"expires_at":        listing.ExpiresAt,
		},
	}

	// Broadcast to all users watching this listing
	if err := s.BroadcastToListing(listing.ListingID, notification); err != nil {
		return fmt.Errorf("failed to broadcast auction closing notification: %w", err)
	}

	return nil
}

// NotifyAuctionClosed sends notifications when an auction closes
func (s *realTimeNotificationService) NotifyAuctionClosed(ctx context.Context, listing *marketplace.Listing, winningBid *marketplace.Bid) error {
	var notification *RealTimeNotificationEvent

	if winningBid != nil {
		notification = &RealTimeNotificationEvent{
			ID:        fmt.Sprintf("auction_closed_%s_%d", listing.ListingID, time.Now().Unix()),
			Type:      "AUCTION_CLOSED",
			ListingID: listing.ListingID,
			Title:     "Auction Closed",
			Message:   fmt.Sprintf("Auction closed with winning bid of %s %.2f", winningBid.Currency, winningBid.BidAmount.InexactFloat64()),
			Priority:  "HIGH",
			Data: map[string]interface{}{
				"winning_bid_id":    winningBid.BidID,
				"winning_amount":    winningBid.BidAmount.InexactFloat64(),
				"winning_bidder_id": winningBid.BidderID,
			},
		}

		// Send special notification to winner
		winnerNotification := &RealTimeNotificationEvent{
			ID:        fmt.Sprintf("auction_won_%s_%d", listing.ListingID, time.Now().Unix()),
			Type:      "AUCTION_WON",
			ListingID: listing.ListingID,
			Title:     "Congratulations! You Won the Auction",
			Message:   fmt.Sprintf("You won the auction with a bid of %s %.2f", winningBid.Currency, winningBid.BidAmount.InexactFloat64()),
			Priority:  "URGENT",
			Data: map[string]interface{}{
				"winning_bid_id": winningBid.BidID,
				"winning_amount": winningBid.BidAmount.InexactFloat64(),
			},
		}

		if err := s.BroadcastToUser(winningBid.BidderID, winnerNotification); err != nil {
			return fmt.Errorf("failed to broadcast winner notification: %w", err)
		}
	} else {
		notification = &RealTimeNotificationEvent{
			ID:        fmt.Sprintf("auction_closed_%s_%d", listing.ListingID, time.Now().Unix()),
			Type:      "AUCTION_CLOSED",
			ListingID: listing.ListingID,
			Title:     "Auction Closed",
			Message:   "Auction closed with no bids",
			Priority:  "MEDIUM",
		}
	}

	// Broadcast to all users watching this listing
	if err := s.BroadcastToListing(listing.ListingID, notification); err != nil {
		return fmt.Errorf("failed to broadcast auction closed notification: %w", err)
	}

	return nil
}

// NotifyAutoBidTriggered sends notifications when an auto-bid is triggered
func (s *realTimeNotificationService) NotifyAutoBidTriggered(ctx context.Context, listing *marketplace.Listing, originalBid *marketplace.Bid, autoBid *marketplace.Bid) error {
	notification := &RealTimeNotificationEvent{
		ID:        fmt.Sprintf("auto_bid_%s_%d", autoBid.BidID, time.Now().Unix()),
		Type:      "AUTO_BID_TRIGGERED",
		ListingID: listing.ListingID,
		Title:     "Auto-Bid Triggered",
		Message:   fmt.Sprintf("Your auto-bid placed a bid of %s %.2f", autoBid.Currency, autoBid.BidAmount.InexactFloat64()),
		Priority:  "MEDIUM",
		Data: map[string]interface{}{
			"auto_bid_id":     autoBid.BidID,
			"auto_bid_amount": autoBid.BidAmount.InexactFloat64(),
			"trigger_bid_id":  originalBid.BidID,
		},
	}

	// Send notification to the auto-bidder
	if err := s.BroadcastToUser(autoBid.BidderID, notification); err != nil {
		return fmt.Errorf("failed to broadcast auto-bid notification: %w", err)
	}

	return nil
}

// SendEmailNotification sends an email notification (placeholder implementation)
func (s *realTimeNotificationService) SendEmailNotification(ctx context.Context, userID string, notification *EmailNotification) error {
	// In a real implementation, this would integrate with an email service like SendGrid, AWS SES, etc.
	log.Printf("Sending email notification to user %s: %s", userID, notification.Subject)
	return nil
}

// SendSMSNotification sends an SMS notification (placeholder implementation)
func (s *realTimeNotificationService) SendSMSNotification(ctx context.Context, userID string, notification *SMSNotification) error {
	// In a real implementation, this would integrate with an SMS service like Twilio, AWS SNS, etc.
	log.Printf("Sending SMS notification to user %s: %s", userID, notification.Message)
	return nil
}

// GetUserNotificationPreferences retrieves user notification preferences (placeholder implementation)
func (s *realTimeNotificationService) GetUserNotificationPreferences(ctx context.Context, userID string) (*NotificationPreferences, error) {
	// In a real implementation, this would fetch from database
	return &NotificationPreferences{
		UserID:           userID,
		WebSocketEnabled: true,
		EmailEnabled:     true,
		SMSEnabled:       false,
		EventPreferences: map[string]bool{
			"BID_PLACED":         true,
			"BID_OUTBID":         true,
			"AUCTION_CLOSING":    true,
			"AUCTION_CLOSED":     true,
			"AUTO_BID_TRIGGERED": true,
		},
		UpdatedAt: time.Now(),
	}, nil
}

// UpdateUserNotificationPreferences updates user notification preferences (placeholder implementation)
func (s *realTimeNotificationService) UpdateUserNotificationPreferences(ctx context.Context, userID string, preferences *NotificationPreferences) error {
	// In a real implementation, this would save to database
	log.Printf("Updating notification preferences for user %s", userID)
	return nil
}

// NotificationHub methods

// run starts the notification hub
func (h *NotificationHub) run() {
	for {
		select {
		case conn := <-h.register:
			h.registerConnection(conn)
		case conn := <-h.unregister:
			h.unregisterConnection(conn)
		case notification := <-h.broadcast:
			h.broadcastNotification(notification)
		}
	}
}

// registerConnection registers a new WebSocket connection
func (h *NotificationHub) registerConnection(conn *WebSocketConnection) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Register by user ID
	h.userConnections[conn.UserID] = conn

	// Register by listing IDs (if any)
	for _, listingID := range conn.ListingIDs {
		if h.listingConnections[listingID] == nil {
			h.listingConnections[listingID] = make(map[string]*WebSocketConnection)
		}
		h.listingConnections[listingID][conn.UserID] = conn
	}

	log.Printf("WebSocket connection registered for user %s", conn.UserID)
}

// unregisterConnection unregisters a WebSocket connection
func (h *NotificationHub) unregisterConnection(conn *WebSocketConnection) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Unregister by user ID
	delete(h.userConnections, conn.UserID)

	// Unregister by listing IDs
	for _, listingID := range conn.ListingIDs {
		if connections, exists := h.listingConnections[listingID]; exists {
			delete(connections, conn.UserID)
			if len(connections) == 0 {
				delete(h.listingConnections, listingID)
			}
		}
	}

	// Close the send channel
	close(conn.Send)

	log.Printf("WebSocket connection unregistered for user %s", conn.UserID)
}

// broadcastNotification broadcasts a notification to all relevant connections
func (h *NotificationHub) broadcastNotification(notification *RealTimeNotificationEvent) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	// Broadcast to specific user if UserID is set
	if notification.UserID != "" {
		if conn, exists := h.userConnections[notification.UserID]; exists {
			select {
			case conn.Send <- notification:
			default:
				close(conn.Send)
				delete(h.userConnections, notification.UserID)
			}
		}
		return
	}

	// Broadcast to all users watching the listing
	if notification.ListingID != "" {
		if connections, exists := h.listingConnections[notification.ListingID]; exists {
			for userID, conn := range connections {
				select {
				case conn.Send <- notification:
				default:
					close(conn.Send)
					delete(connections, userID)
				}
			}
		}
	}
}

// WebSocketConnection methods

// readPump pumps messages from the WebSocket connection
func (c *WebSocketConnection) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	// Set read deadline and pong handler
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		// Read message from WebSocket
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming messages (e.g., subscription to listings)
		c.handleMessage(message)
	}
}

// writePump pumps messages to the WebSocket connection
func (c *WebSocketConnection) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case notification, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Send notification as JSON
			if err := c.Conn.WriteJSON(notification); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage handles incoming WebSocket messages
func (c *WebSocketConnection) handleMessage(message []byte) {
	// Parse message to handle subscription requests
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Failed to parse WebSocket message: %v", err)
		return
	}

	// Handle subscription to listings
	if action, ok := msg["action"].(string); ok && action == "subscribe" {
		if listingID, ok := msg["listing_id"].(string); ok {
			c.subscribeToListing(listingID)
		}
	}

	// Handle unsubscription from listings
	if action, ok := msg["action"].(string); ok && action == "unsubscribe" {
		if listingID, ok := msg["listing_id"].(string); ok {
			c.unsubscribeFromListing(listingID)
		}
	}
}

// subscribeToListing subscribes the connection to a specific listing
func (c *WebSocketConnection) subscribeToListing(listingID string) {
	c.Hub.mutex.Lock()
	defer c.Hub.mutex.Unlock()

	// Add to listing connections
	if c.Hub.listingConnections[listingID] == nil {
		c.Hub.listingConnections[listingID] = make(map[string]*WebSocketConnection)
	}
	c.Hub.listingConnections[listingID][c.UserID] = c

	// Add to connection's listing IDs
	for _, id := range c.ListingIDs {
		if id == listingID {
			return // Already subscribed
		}
	}
	c.ListingIDs = append(c.ListingIDs, listingID)

	log.Printf("User %s subscribed to listing %s", c.UserID, listingID)
}

// unsubscribeFromListing unsubscribes the connection from a specific listing
func (c *WebSocketConnection) unsubscribeFromListing(listingID string) {
	c.Hub.mutex.Lock()
	defer c.Hub.mutex.Unlock()

	// Remove from listing connections
	if connections, exists := c.Hub.listingConnections[listingID]; exists {
		delete(connections, c.UserID)
		if len(connections) == 0 {
			delete(c.Hub.listingConnections, listingID)
		}
	}

	// Remove from connection's listing IDs
	for i, id := range c.ListingIDs {
		if id == listingID {
			c.ListingIDs = append(c.ListingIDs[:i], c.ListingIDs[i+1:]...)
			break
		}
	}

	log.Printf("User %s unsubscribed from listing %s", c.UserID, listingID)
}
