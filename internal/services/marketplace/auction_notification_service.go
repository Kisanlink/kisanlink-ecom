package marketplace

import (
	"context"
	"fmt"
	"time"

	"kisanlink-ecom/entities/models/marketplace"

	"github.com/sirupsen/logrus"
)

// AuctionNotificationServiceInterface defines the interface for auction notification operations
type AuctionNotificationServiceInterface interface {
	// Winner notifications
	NotifyAuctionWinner(ctx context.Context, listing *marketplace.Listing, winningBid *marketplace.Bid) error
	NotifyAuctionLosers(ctx context.Context, listing *marketplace.Listing, losingBids []*marketplace.Bid) error

	// Seller notifications
	NotifySellerAuctionCompleted(ctx context.Context, listing *marketplace.Listing, winningBid *marketplace.Bid) error
	NotifySellerAuctionExpiredNoBids(ctx context.Context, listing *marketplace.Listing) error

	// Bidder notifications
	NotifyBidderOutbid(ctx context.Context, listing *marketplace.Listing, outbidBid *marketplace.Bid, newHighestBid *marketplace.Bid) error
	NotifyBidderAuctionEnding(ctx context.Context, listing *marketplace.Listing, bid *marketplace.Bid, timeRemaining time.Duration) error

	// Order notifications
	NotifyOrderCreated(ctx context.Context, orderID string, listing *marketplace.Listing, winningBid *marketplace.Bid) error

	// Batch notifications
	SendBatchNotifications(ctx context.Context, notifications []*AuctionNotification) error
}

// AuctionNotification represents a notification to be sent
type AuctionNotification struct {
	ID          string                 `json:"id"`
	Type        NotificationType       `json:"type"`
	RecipientID string                 `json:"recipient_id"`
	Subject     string                 `json:"subject"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data"`
	Priority    NotificationPriority   `json:"priority"`
	Channels    []NotificationChannel  `json:"channels"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeAuctionWon           NotificationType = "AUCTION_WON"
	NotificationTypeAuctionLost          NotificationType = "AUCTION_LOST"
	NotificationTypeAuctionCompleted     NotificationType = "AUCTION_COMPLETED"
	NotificationTypeAuctionExpiredNoBids NotificationType = "AUCTION_EXPIRED_NO_BIDS"
	NotificationTypeBidOutbid            NotificationType = "BID_OUTBID"
	NotificationTypeAuctionEnding        NotificationType = "AUCTION_ENDING"
	NotificationTypeOrderCreatedFromBid  NotificationType = "ORDER_CREATED_FROM_BID"
)

// NotificationPriority represents the priority level of notifications
type NotificationPriority string

const (
	NotificationPriorityHigh   NotificationPriority = "HIGH"
	NotificationPriorityMedium NotificationPriority = "MEDIUM"
	NotificationPriorityLow    NotificationPriority = "LOW"
)

// NotificationChannel represents the delivery channel for notifications
type NotificationChannel string

const (
	NotificationChannelEmail   NotificationChannel = "EMAIL"
	NotificationChannelSMS     NotificationChannel = "SMS"
	NotificationChannelPush    NotificationChannel = "PUSH"
	NotificationChannelInApp   NotificationChannel = "IN_APP"
	NotificationChannelWebhook NotificationChannel = "WEBHOOK"
)

// NotificationTemplate represents a template for generating notifications
type NotificationTemplate struct {
	Type     NotificationType      `json:"type"`
	Subject  string                `json:"subject"`
	Message  string                `json:"message"`
	Channels []NotificationChannel `json:"channels"`
	Priority NotificationPriority  `json:"priority"`
}

// AuctionNotificationService provides notification services for auction events
type AuctionNotificationService struct {
	notificationSvc NotificationServiceInterface
	templates       map[NotificationType]*NotificationTemplate
	logger          *logrus.Logger
}

// NewAuctionNotificationService creates a new auction notification service
func NewAuctionNotificationService(
	notificationSvc NotificationServiceInterface,
	logger *logrus.Logger,
) AuctionNotificationServiceInterface {
	if logger == nil {
		logger = logrus.New()
	}

	service := &AuctionNotificationService{
		notificationSvc: notificationSvc,
		templates:       make(map[NotificationType]*NotificationTemplate),
		logger:          logger,
	}

	// Initialize default templates
	service.initializeDefaultTemplates()

	return service
}

// NotifyAuctionWinner sends notification to the auction winner
func (s *AuctionNotificationService) NotifyAuctionWinner(ctx context.Context, listing *marketplace.Listing, winningBid *marketplace.Bid) error {
	template := s.templates[NotificationTypeAuctionWon]
	if template == nil {
		return fmt.Errorf("no template found for auction won notification")
	}

	notification := &AuctionNotification{
		ID:          generateNotificationID(),
		Type:        NotificationTypeAuctionWon,
		RecipientID: winningBid.BidderID,
		Subject:     s.formatTemplate(template.Subject, listing, winningBid, nil),
		Message:     s.formatTemplate(template.Message, listing, winningBid, nil),
		Priority:    template.Priority,
		Channels:    template.Channels,
		Data: map[string]interface{}{
			"listing_id": listing.ListingID,
			"bid_id":     winningBid.BidID,
			"bid_amount": winningBid.BidAmount.String(),
			"product_id": listing.ProductID,
			"quantity":   winningBid.Quantity.String(),
			"currency":   winningBid.Currency,
			"seller_id":  listing.SellerID,
		},
		CreatedAt: time.Now(),
	}

	return s.sendNotification(ctx, notification)
}

// NotifyAuctionLosers sends notifications to all losing bidders
func (s *AuctionNotificationService) NotifyAuctionLosers(ctx context.Context, listing *marketplace.Listing, losingBids []*marketplace.Bid) error {
	template := s.templates[NotificationTypeAuctionLost]
	if template == nil {
		return fmt.Errorf("no template found for auction lost notification")
	}

	var notifications []*AuctionNotification
	for _, bid := range losingBids {
		notification := &AuctionNotification{
			ID:          generateNotificationID(),
			Type:        NotificationTypeAuctionLost,
			RecipientID: bid.BidderID,
			Subject:     s.formatTemplate(template.Subject, listing, bid, nil),
			Message:     s.formatTemplate(template.Message, listing, bid, nil),
			Priority:    template.Priority,
			Channels:    template.Channels,
			Data: map[string]interface{}{
				"listing_id": listing.ListingID,
				"bid_id":     bid.BidID,
				"bid_amount": bid.BidAmount.String(),
				"product_id": listing.ProductID,
				"seller_id":  listing.SellerID,
			},
			CreatedAt: time.Now(),
		}
		notifications = append(notifications, notification)
	}

	return s.SendBatchNotifications(ctx, notifications)
}

// NotifySellerAuctionCompleted sends notification to seller when auction completes with winner
func (s *AuctionNotificationService) NotifySellerAuctionCompleted(ctx context.Context, listing *marketplace.Listing, winningBid *marketplace.Bid) error {
	template := s.templates[NotificationTypeAuctionCompleted]
	if template == nil {
		return fmt.Errorf("no template found for auction completed notification")
	}

	notification := &AuctionNotification{
		ID:          generateNotificationID(),
		Type:        NotificationTypeAuctionCompleted,
		RecipientID: listing.SellerID,
		Subject:     s.formatTemplate(template.Subject, listing, winningBid, nil),
		Message:     s.formatTemplate(template.Message, listing, winningBid, nil),
		Priority:    template.Priority,
		Channels:    template.Channels,
		Data: map[string]interface{}{
			"listing_id":     listing.ListingID,
			"winning_bid_id": winningBid.BidID,
			"winning_amount": winningBid.BidAmount.String(),
			"winner_id":      winningBid.BidderID,
			"product_id":     listing.ProductID,
			"quantity":       winningBid.Quantity.String(),
			"currency":       winningBid.Currency,
			"total_bids":     listing.BidCount,
		},
		CreatedAt: time.Now(),
	}

	return s.sendNotification(ctx, notification)
}

// NotifySellerAuctionExpiredNoBids sends notification to seller when auction expires without bids
func (s *AuctionNotificationService) NotifySellerAuctionExpiredNoBids(ctx context.Context, listing *marketplace.Listing) error {
	template := s.templates[NotificationTypeAuctionExpiredNoBids]
	if template == nil {
		return fmt.Errorf("no template found for auction expired no bids notification")
	}

	notification := &AuctionNotification{
		ID:          generateNotificationID(),
		Type:        NotificationTypeAuctionExpiredNoBids,
		RecipientID: listing.SellerID,
		Subject:     s.formatTemplate(template.Subject, listing, nil, nil),
		Message:     s.formatTemplate(template.Message, listing, nil, nil),
		Priority:    template.Priority,
		Channels:    template.Channels,
		Data: map[string]interface{}{
			"listing_id":   listing.ListingID,
			"product_id":   listing.ProductID,
			"asking_price": listing.AskingPrice.String(),
			"minimum_bid":  listing.MinimumBid.String(),
			"currency":     listing.Currency,
		},
		CreatedAt: time.Now(),
	}

	return s.sendNotification(ctx, notification)
}

// NotifyBidderOutbid sends notification when a bidder is outbid
func (s *AuctionNotificationService) NotifyBidderOutbid(ctx context.Context, listing *marketplace.Listing, outbidBid *marketplace.Bid, newHighestBid *marketplace.Bid) error {
	template := s.templates[NotificationTypeBidOutbid]
	if template == nil {
		return fmt.Errorf("no template found for bid outbid notification")
	}

	notification := &AuctionNotification{
		ID:          generateNotificationID(),
		Type:        NotificationTypeBidOutbid,
		RecipientID: outbidBid.BidderID,
		Subject:     s.formatTemplate(template.Subject, listing, outbidBid, newHighestBid),
		Message:     s.formatTemplate(template.Message, listing, outbidBid, newHighestBid),
		Priority:    template.Priority,
		Channels:    template.Channels,
		Data: map[string]interface{}{
			"listing_id":         listing.ListingID,
			"outbid_bid_id":      outbidBid.BidID,
			"outbid_amount":      outbidBid.BidAmount.String(),
			"new_highest_amount": newHighestBid.BidAmount.String(),
			"product_id":         listing.ProductID,
			"time_remaining":     listing.GetTimeRemaining().String(),
		},
		CreatedAt: time.Now(),
	}

	return s.sendNotification(ctx, notification)
}

// NotifyBidderAuctionEnding sends notification when auction is ending soon
func (s *AuctionNotificationService) NotifyBidderAuctionEnding(ctx context.Context, listing *marketplace.Listing, bid *marketplace.Bid, timeRemaining time.Duration) error {
	template := s.templates[NotificationTypeAuctionEnding]
	if template == nil {
		return fmt.Errorf("no template found for auction ending notification")
	}

	notification := &AuctionNotification{
		ID:          generateNotificationID(),
		Type:        NotificationTypeAuctionEnding,
		RecipientID: bid.BidderID,
		Subject:     s.formatTemplate(template.Subject, listing, bid, nil),
		Message:     s.formatTemplate(template.Message, listing, bid, nil),
		Priority:    template.Priority,
		Channels:    template.Channels,
		Data: map[string]interface{}{
			"listing_id":     listing.ListingID,
			"bid_id":         bid.BidID,
			"current_amount": bid.BidAmount.String(),
			"product_id":     listing.ProductID,
			"time_remaining": timeRemaining.String(),
			"is_highest":     bid.IsHighestBid,
		},
		CreatedAt: time.Now(),
	}

	return s.sendNotification(ctx, notification)
}

// NotifyOrderCreated sends notification when an order is created from winning bid
func (s *AuctionNotificationService) NotifyOrderCreated(ctx context.Context, orderID string, listing *marketplace.Listing, winningBid *marketplace.Bid) error {
	template := s.templates[NotificationTypeOrderCreatedFromBid]
	if template == nil {
		return fmt.Errorf("no template found for order created notification")
	}

	// Notify buyer
	buyerNotification := &AuctionNotification{
		ID:          generateNotificationID(),
		Type:        NotificationTypeOrderCreatedFromBid,
		RecipientID: winningBid.BidderID,
		Subject:     s.formatTemplate(template.Subject, listing, winningBid, nil),
		Message:     s.formatTemplate(template.Message, listing, winningBid, nil),
		Priority:    template.Priority,
		Channels:    template.Channels,
		Data: map[string]interface{}{
			"order_id":   orderID,
			"listing_id": listing.ListingID,
			"bid_id":     winningBid.BidID,
			"amount":     winningBid.BidAmount.String(),
			"product_id": listing.ProductID,
			"quantity":   winningBid.Quantity.String(),
			"seller_id":  listing.SellerID,
			"role":       "buyer",
		},
		CreatedAt: time.Now(),
	}

	// Notify seller
	sellerNotification := &AuctionNotification{
		ID:          generateNotificationID(),
		Type:        NotificationTypeOrderCreatedFromBid,
		RecipientID: listing.SellerID,
		Subject:     s.formatTemplate(template.Subject, listing, winningBid, nil),
		Message:     s.formatTemplate(template.Message, listing, winningBid, nil),
		Priority:    template.Priority,
		Channels:    template.Channels,
		Data: map[string]interface{}{
			"order_id":   orderID,
			"listing_id": listing.ListingID,
			"bid_id":     winningBid.BidID,
			"amount":     winningBid.BidAmount.String(),
			"product_id": listing.ProductID,
			"quantity":   winningBid.Quantity.String(),
			"buyer_id":   winningBid.BidderID,
			"role":       "seller",
		},
		CreatedAt: time.Now(),
	}

	return s.SendBatchNotifications(ctx, []*AuctionNotification{buyerNotification, sellerNotification})
}

// SendBatchNotifications sends multiple notifications in batch
func (s *AuctionNotificationService) SendBatchNotifications(ctx context.Context, notifications []*AuctionNotification) error {
	if len(notifications) == 0 {
		return nil
	}

	s.logger.WithField("count", len(notifications)).Debug("Sending batch notifications")

	// Convert to notification requests
	var requests []NotificationRequest
	for _, notification := range notifications {
		for _ = range notification.Channels {
			request := NotificationRequest{
				Type:      string(notification.Type),
				Recipient: notification.RecipientID,
				Subject:   notification.Subject,
				Message:   notification.Message,
				Data:      notification.Data,
				Priority:  string(notification.Priority),
			}
			requests = append(requests, request)
		}
	}

	// Send through notification service
	if s.notificationSvc != nil {
		return s.notificationSvc.SendBatchNotifications(ctx, requests)
	}

	// Fallback: log notifications if no service available
	for _, notification := range notifications {
		s.logger.WithFields(logrus.Fields{
			"type":      notification.Type,
			"recipient": notification.RecipientID,
			"subject":   notification.Subject,
		}).Info("Notification (service unavailable)")
	}

	return nil
}

// sendNotification sends a single notification
func (s *AuctionNotificationService) sendNotification(ctx context.Context, notification *AuctionNotification) error {
	return s.SendBatchNotifications(ctx, []*AuctionNotification{notification})
}

// initializeDefaultTemplates sets up default notification templates
func (s *AuctionNotificationService) initializeDefaultTemplates() {
	s.templates[NotificationTypeAuctionWon] = &NotificationTemplate{
		Type:     NotificationTypeAuctionWon,
		Subject:  "Congratulations! You won the auction for {{.ProductID}}",
		Message:  "You have won the auction for {{.ProductID}} with your bid of {{.BidAmount}} {{.Currency}}. An order will be created for you to complete the purchase.",
		Channels: []NotificationChannel{NotificationChannelEmail, NotificationChannelInApp, NotificationChannelPush},
		Priority: NotificationPriorityHigh,
	}

	s.templates[NotificationTypeAuctionLost] = &NotificationTemplate{
		Type:     NotificationTypeAuctionLost,
		Subject:  "Auction ended - {{.ProductID}}",
		Message:  "The auction for {{.ProductID}} has ended. Unfortunately, your bid of {{.BidAmount}} {{.Currency}} was not the winning bid.",
		Channels: []NotificationChannel{NotificationChannelEmail, NotificationChannelInApp},
		Priority: NotificationPriorityMedium,
	}

	s.templates[NotificationTypeAuctionCompleted] = &NotificationTemplate{
		Type:     NotificationTypeAuctionCompleted,
		Subject:  "Your auction has completed - {{.ProductID}}",
		Message:  "Your auction for {{.ProductID}} has completed successfully! The winning bid was {{.WinningAmount}} {{.Currency}} from {{.TotalBids}} total bids.",
		Channels: []NotificationChannel{NotificationChannelEmail, NotificationChannelInApp, NotificationChannelPush},
		Priority: NotificationPriorityHigh,
	}

	s.templates[NotificationTypeAuctionExpiredNoBids] = &NotificationTemplate{
		Type:     NotificationTypeAuctionExpiredNoBids,
		Subject:  "Your auction expired without bids - {{.ProductID}}",
		Message:  "Your auction for {{.ProductID}} has expired without receiving any bids. You may want to consider relisting with a lower starting price.",
		Channels: []NotificationChannel{NotificationChannelEmail, NotificationChannelInApp},
		Priority: NotificationPriorityMedium,
	}

	s.templates[NotificationTypeBidOutbid] = &NotificationTemplate{
		Type:     NotificationTypeBidOutbid,
		Subject:  "You've been outbid on {{.ProductID}}",
		Message:  "Your bid of {{.OutbidAmount}} {{.Currency}} on {{.ProductID}} has been outbid. The current highest bid is {{.NewHighestAmount}} {{.Currency}}.",
		Channels: []NotificationChannel{NotificationChannelEmail, NotificationChannelInApp, NotificationChannelPush},
		Priority: NotificationPriorityMedium,
	}

	s.templates[NotificationTypeAuctionEnding] = &NotificationTemplate{
		Type:     NotificationTypeAuctionEnding,
		Subject:  "Auction ending soon - {{.ProductID}}",
		Message:  "The auction for {{.ProductID}} is ending in {{.TimeRemaining}}. Your current bid is {{.CurrentAmount}} {{.Currency}}.",
		Channels: []NotificationChannel{NotificationChannelInApp, NotificationChannelPush},
		Priority: NotificationPriorityMedium,
	}

	s.templates[NotificationTypeOrderCreatedFromBid] = &NotificationTemplate{
		Type:     NotificationTypeOrderCreatedFromBid,
		Subject:  "Order created from your winning bid - {{.ProductID}}",
		Message:  "An order has been created for your winning bid on {{.ProductID}}. Please complete the payment to finalize your purchase.",
		Channels: []NotificationChannel{NotificationChannelEmail, NotificationChannelInApp, NotificationChannelPush},
		Priority: NotificationPriorityHigh,
	}
}

// formatTemplate formats a template string with listing and bid data
func (s *AuctionNotificationService) formatTemplate(template string, listing *marketplace.Listing, bid *marketplace.Bid, otherBid *marketplace.Bid) string {
	// This is a simplified template formatting
	// In production, you would use a proper template engine like text/template
	result := template

	if listing != nil {
		result = replaceTemplateVar(result, "ProductID", listing.ProductID)
		result = replaceTemplateVar(result, "Currency", listing.Currency)
		result = replaceTemplateVar(result, "AskingPrice", listing.AskingPrice.String())
		result = replaceTemplateVar(result, "MinimumBid", listing.MinimumBid.String())
		result = replaceTemplateVar(result, "TotalBids", fmt.Sprintf("%d", listing.BidCount))
		result = replaceTemplateVar(result, "TimeRemaining", listing.GetTimeRemaining().String())
	}

	if bid != nil {
		result = replaceTemplateVar(result, "BidAmount", bid.BidAmount.String())
		result = replaceTemplateVar(result, "OutbidAmount", bid.BidAmount.String())
		result = replaceTemplateVar(result, "CurrentAmount", bid.BidAmount.String())
		result = replaceTemplateVar(result, "WinningAmount", bid.BidAmount.String())
		result = replaceTemplateVar(result, "Quantity", bid.Quantity.String())
	}

	if otherBid != nil {
		result = replaceTemplateVar(result, "NewHighestAmount", otherBid.BidAmount.String())
	}

	return result
}

// replaceTemplateVar replaces template variables in the format {{.VarName}}
func replaceTemplateVar(template, varName, value string) string {
	placeholder := fmt.Sprintf("{{.%s}}", varName)
	return fmt.Sprintf(template, placeholder, value)
}

// generateNotificationID generates a unique notification ID
func generateNotificationID() string {
	return fmt.Sprintf("NOTIF_%d", time.Now().UnixNano())
}
