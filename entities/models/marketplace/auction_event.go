package marketplace

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// AuctionEventType represents the type of auction event
type AuctionEventType string

const (
	EventListingCreated   AuctionEventType = "LISTING_CREATED"
	EventBidPlaced        AuctionEventType = "BID_PLACED"
	EventBidOutbid        AuctionEventType = "BID_OUTBID"
	EventAutoBidTriggered AuctionEventType = "AUTO_BID_TRIGGERED"
	EventListingClosed    AuctionEventType = "LISTING_CLOSED"
	EventListingExpired   AuctionEventType = "LISTING_EXPIRED"
	EventListingCancelled AuctionEventType = "LISTING_CANCELLED"
	EventBidRemoved       AuctionEventType = "BID_REMOVED"
	EventListingUpdated   AuctionEventType = "LISTING_UPDATED"
)

// ActorType represents the type of actor performing the action
type ActorType string

const (
	ActorTypeUser   ActorType = "USER"
	ActorTypeSystem ActorType = "SYSTEM"
	ActorTypeAdmin  ActorType = "ADMIN"
)

// AuctionEvent represents an event in the auction lifecycle for audit trail
type AuctionEvent struct {
	base.BaseModel

	EventID   string           `json:"event_id" gorm:"type:varchar(50);uniqueIndex;not null"`
	ListingID string           `json:"listing_id" gorm:"type:varchar(50);not null;index"`
	EventType AuctionEventType `json:"event_type" gorm:"type:varchar(50);not null;index"`
	EventData string           `json:"event_data" gorm:"type:jsonb"`
	ActorID   string           `json:"actor_id" gorm:"type:varchar(50);not null"`
	ActorType ActorType        `json:"actor_type" gorm:"type:varchar(20);not null"`
	Timestamp time.Time        `json:"timestamp" gorm:"not null;index"`
}

// TableName returns the table name for GORM
func (AuctionEvent) TableName() string {
	return "auction_events"
}

// NewAuctionEvent creates a new auction event
func NewAuctionEvent(listingID string, eventType AuctionEventType, actorID string, actorType ActorType, eventData interface{}) *AuctionEvent {
	return &AuctionEvent{
		BaseModel: *base.NewBaseModel("EVT", "large"),
		EventID:   generateEventID(),
		ListingID: listingID,
		EventType: eventType,
		ActorID:   actorID,
		ActorType: actorType,
		Timestamp: time.Now(),
	}
}

// generateEventID generates a unique event ID
func generateEventID() string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("EVT_%d", timestamp)
}

// SetEventData sets the event data from any interface
func (ae *AuctionEvent) SetEventData(data interface{}) error {
	if data == nil {
		ae.EventData = ""
		return nil
	}

	eventDataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	ae.EventData = string(eventDataJSON)
	return nil
}

// GetEventData returns the event data as a map
func (ae *AuctionEvent) GetEventData() (map[string]interface{}, error) {
	if ae.EventData == "" {
		return nil, nil
	}

	var eventData map[string]interface{}
	if err := json.Unmarshal([]byte(ae.EventData), &eventData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	return eventData, nil
}

// Event Data Structures for different event types

// ListingCreatedEventData represents data for listing creation events
type ListingCreatedEventData struct {
	ProductID       string  `json:"product_id"`
	AskingPrice     float64 `json:"asking_price"`
	MinimumBid      float64 `json:"minimum_bid"`
	DurationHours   int     `json:"duration_hours"`
	Visibility      string  `json:"visibility"`
	AuctionType     string  `json:"auction_type"`
	BidVisibility   string  `json:"bid_visibility"`
}

// BidPlacedEventData represents data for bid placement events
type BidPlacedEventData struct {
	BidID           string  `json:"bid_id"`
	BidderID        string  `json:"bidder_id"`
	BidAmount       float64 `json:"bid_amount"`
	PreviousHighest *float64 `json:"previous_highest,omitempty"`
	IsAutoBid       bool    `json:"is_auto_bid"`
	Message         string  `json:"message,omitempty"`
}

// BidOutbidEventData represents data for bid outbid events
type BidOutbidEventData struct {
	OutbidBidID     string  `json:"outbid_bid_id"`
	OutbidBidderID  string  `json:"outbid_bidder_id"`
	OutbidAmount    float64 `json:"outbid_amount"`
	NewHighestBidID string  `json:"new_highest_bid_id"`
	NewHighestAmount float64 `json:"new_highest_amount"`
}

// AutoBidTriggeredEventData represents data for auto-bid trigger events
type AutoBidTriggeredEventData struct {
	OriginalBidID    string  `json:"original_bid_id"`
	AutoBidID        string  `json:"auto_bid_id"`
	TriggerBidAmount float64 `json:"trigger_bid_amount"`
	AutoBidAmount    float64 `json:"auto_bid_amount"`
	AutoBidLimit     float64 `json:"auto_bid_limit"`
}

// ListingClosedEventData represents data for listing closure events
type ListingClosedEventData struct {
	CloseReason     string  `json:"close_reason"`
	WinningBidID    *string `json:"winning_bid_id,omitempty"`
	WinningBidderID *string `json:"winning_bidder_id,omitempty"`
	WinningAmount   *float64 `json:"winning_amount,omitempty"`
	TotalBids       int     `json:"total_bids"`
	FinalStatus     string  `json:"final_status"`
}

// ListingExpiredEventData represents data for listing expiry events
type ListingExpiredEventData struct {
	ExpiredAt       time.Time `json:"expired_at"`
	WinningBidID    *string   `json:"winning_bid_id,omitempty"`
	WinningBidderID *string   `json:"winning_bidder_id,omitempty"`
	WinningAmount   *float64  `json:"winning_amount,omitempty"`
	TotalBids       int       `json:"total_bids"`
	HasBids         bool      `json:"has_bids"`
}

// BidRemovedEventData represents data for bid removal events
type BidRemovedEventData struct {
	RemovedBidID     string  `json:"removed_bid_id"`
	RemovedBidderID  string  `json:"removed_bidder_id"`
	RemovedAmount    float64 `json:"removed_amount"`
	RemovalReason    string  `json:"removal_reason"`
	WasHighestBid    bool    `json:"was_highest_bid"`
	NewHighestBidID  *string `json:"new_highest_bid_id,omitempty"`
	NewHighestAmount *float64 `json:"new_highest_amount,omitempty"`
}

// ListingUpdatedEventData represents data for listing update events
type ListingUpdatedEventData struct {
	UpdatedFields map[string]interface{} `json:"updated_fields"`
	UpdateReason  string                 `json:"update_reason"`
}

// Helper functions to create specific event types

// CreateListingCreatedEvent creates a listing creation event
func CreateListingCreatedEvent(listingID, actorID string, listing *Listing) (*AuctionEvent, error) {
	eventData := ListingCreatedEventData{
		ProductID:     listing.ProductID,
		AskingPrice:   listing.AskingPrice.InexactFloat64(),
		MinimumBid:    listing.MinimumBid.InexactFloat64(),
		DurationHours: listing.ListingDuration,
		Visibility:    string(listing.Visibility),
		AuctionType:   string(listing.AuctionType),
		BidVisibility: string(listing.BidVisibility),
	}

	event := NewAuctionEvent(listingID, EventListingCreated, actorID, ActorTypeUser, nil)
	if err := event.SetEventData(eventData); err != nil {
		return nil, err
	}

	return event, nil
}

// CreateBidPlacedEvent creates a bid placement event
func CreateBidPlacedEvent(listingID, actorID string, bid *Bid, previousHighest *float64) (*AuctionEvent, error) {
	eventData := BidPlacedEventData{
		BidID:           bid.BidID,
		BidderID:        bid.BidderID,
		BidAmount:       bid.BidAmount.InexactFloat64(),
		PreviousHighest: previousHighest,
		IsAutoBid:       bid.IsAutoBid,
		Message:         bid.Message,
	}

	event := NewAuctionEvent(listingID, EventBidPlaced, actorID, ActorTypeUser, nil)
	if err := event.SetEventData(eventData); err != nil {
		return nil, err
	}

	return event, nil
}

// CreateListingClosedEvent creates a listing closure event
func CreateListingClosedEvent(listingID, actorID string, actorType ActorType, closeReason string, winningBid *Bid, totalBids int, finalStatus ListingStatus) (*AuctionEvent, error) {
	eventData := ListingClosedEventData{
		CloseReason: closeReason,
		TotalBids:   totalBids,
		FinalStatus: string(finalStatus),
	}

	if winningBid != nil {
		eventData.WinningBidID = &winningBid.BidID
		eventData.WinningBidderID = &winningBid.BidderID
		amount := winningBid.BidAmount.InexactFloat64()
		eventData.WinningAmount = &amount
	}

	event := NewAuctionEvent(listingID, EventListingClosed, actorID, actorType, nil)
	if err := event.SetEventData(eventData); err != nil {
		return nil, err
	}

	return event, nil
}

// EventFilter represents filters for event queries
type EventFilter struct {
	ListingID   string            `json:"listing_id,omitempty"`
	EventType   *AuctionEventType `json:"event_type,omitempty"`
	ActorID     string            `json:"actor_id,omitempty"`
	ActorType   *ActorType        `json:"actor_type,omitempty"`
	TimeFrom    *time.Time        `json:"time_from,omitempty"`
	TimeTo      *time.Time        `json:"time_to,omitempty"`
}

// EventSummary represents a summary view of an auction event
type EventSummary struct {
	ID        string           `json:"id"`
	EventID   string           `json:"event_id"`
	ListingID string           `json:"listing_id"`
	EventType AuctionEventType `json:"event_type"`
	ActorID   string           `json:"actor_id"`
	ActorType ActorType        `json:"actor_type"`
	Timestamp time.Time        `json:"timestamp"`
	Summary   string           `json:"summary"` // Human-readable event summary
}