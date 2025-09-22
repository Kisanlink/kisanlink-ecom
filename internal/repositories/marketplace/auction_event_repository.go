package marketplace

import (
	"context"
	"fmt"
	"time"

	"kisanlink-ecom/entities/models/marketplace"
	"kisanlink-ecom/internal/common"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// AuctionEventRepository interface defines the contract for auction event data operations
type AuctionEventRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, event *marketplace.AuctionEvent) error
	GetByID(ctx context.Context, id string) (*marketplace.AuctionEvent, error)
	GetByEventID(ctx context.Context, eventID string) (*marketplace.AuctionEvent, error)

	// Event logging operations
	LogEvent(ctx context.Context, listingID string, eventType marketplace.AuctionEventType, actorID string, actorType marketplace.ActorType, eventData interface{}) error
	LogListingCreated(ctx context.Context, listingID, actorID string, listing *marketplace.Listing) error
	LogBidPlaced(ctx context.Context, listingID, actorID string, bid *marketplace.Bid, previousHighest *float64) error
	LogBidOutbid(ctx context.Context, listingID, actorID string, outbidBid *marketplace.Bid, newHighestBid *marketplace.Bid) error
	LogAutoBidTriggered(ctx context.Context, listingID, actorID string, originalBid, autoBid *marketplace.Bid) error
	LogListingClosed(ctx context.Context, listingID, actorID string, actorType marketplace.ActorType, closeReason string, winningBid *marketplace.Bid, totalBids int, finalStatus marketplace.ListingStatus) error
	LogListingExpired(ctx context.Context, listingID string, winningBid *marketplace.Bid, totalBids int) error
	LogBidRemoved(ctx context.Context, listingID, actorID string, removedBid *marketplace.Bid, removalReason string, newHighestBid *marketplace.Bid) error

	// Event querying for audit and analytics
	GetEventsByListing(ctx context.Context, listingID string, filter *marketplace.EventFilter, pagination *common.PaginationRequest) ([]*marketplace.AuctionEvent, int, error)
	GetByListingID(ctx context.Context, listingID string, pagination *common.PaginationRequest) ([]*marketplace.AuctionEvent, int, error)
	GetEventsByActor(ctx context.Context, actorID string, filter *marketplace.EventFilter, pagination *common.PaginationRequest) ([]*marketplace.AuctionEvent, int, error)
	GetEventsByType(ctx context.Context, eventType marketplace.AuctionEventType, filter *marketplace.EventFilter, pagination *common.PaginationRequest) ([]*marketplace.AuctionEvent, int, error)
	GetByEventType(ctx context.Context, eventType marketplace.AuctionEventType, pagination *common.PaginationRequest) ([]*marketplace.AuctionEvent, int, error)
	GetEventsByTimeRange(ctx context.Context, startTime, endTime time.Time, filter *marketplace.EventFilter, pagination *common.PaginationRequest) ([]*marketplace.AuctionEvent, int, error)

	// Event streaming for real-time updates
	GetRecentEvents(ctx context.Context, listingID string, since time.Time, limit int) ([]*marketplace.AuctionEvent, error)
	GetEventsSince(ctx context.Context, since time.Time, filter *marketplace.EventFilter, limit int) ([]*marketplace.AuctionEvent, error)

	// Analytics and reporting
	GetEventSummary(ctx context.Context, listingID string) ([]*marketplace.EventSummary, error)
	GetEventStatistics(ctx context.Context, listingID string, timeRange *TimeRange) (*EventStatistics, error)
	GetAuditTrail(ctx context.Context, listingID string, pagination *common.PaginationRequest) ([]*marketplace.AuctionEvent, int, error)

	// Admin operations
	GetAllEvents(ctx context.Context, filter *marketplace.EventFilter, pagination *common.PaginationRequest) ([]*marketplace.AuctionEvent, int, error)
	GetEvents(ctx context.Context, filter *marketplace.EventFilter, pagination *common.PaginationRequest) ([]*marketplace.AuctionEvent, int, error)
	DeleteEventsByListing(ctx context.Context, listingID string) error
	ArchiveEventsBefore(ctx context.Context, beforeDate time.Time) (int, error)
	DeleteEventsBefore(ctx context.Context, beforeDate time.Time) (int, error)
}

// TimeRange represents a time range for analytics
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// EventStatistics represents event statistics for a listing
type EventStatistics struct {
	ListingID       string                               `json:"listing_id"`
	TotalEvents     int                                  `json:"total_events"`
	EventsByType    map[marketplace.AuctionEventType]int `json:"events_by_type"`
	UniqueActors    int                                  `json:"unique_actors"`
	FirstEventTime  *time.Time                           `json:"first_event_time"`
	LastEventTime   *time.Time                           `json:"last_event_time"`
	AuctionDuration *time.Duration                       `json:"auction_duration"`
}

// auctionEventRepository implements the AuctionEventRepository interface
type auctionEventRepository struct {
	dbManager db.DBManager
}

// NewAuctionEventRepository creates a new auction event repository
func NewAuctionEventRepository(dbManager db.DBManager) AuctionEventRepository {
	return &auctionEventRepository{
		dbManager: dbManager,
	}
}

// Create creates a new auction event in the database
func (r *auctionEventRepository) Create(ctx context.Context, event *marketplace.AuctionEvent) error {
	if err := r.dbManager.Create(ctx, event); err != nil {
		return fmt.Errorf("failed to create auction event: %w", err)
	}
	return nil
}

// GetByID retrieves an auction event by its database ID
func (r *auctionEventRepository) GetByID(ctx context.Context, id string) (*marketplace.AuctionEvent, error) {
	var event marketplace.AuctionEvent
	if err := r.dbManager.GetByID(ctx, id, &event); err != nil {
		return nil, fmt.Errorf("failed to get auction event by ID: %w", err)
	}
	return &event, nil
}

// GetByEventID retrieves an auction event by its event ID
func (r *auctionEventRepository) GetByEventID(ctx context.Context, eventID string) (*marketplace.AuctionEvent, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "event_id",
			Operator: base.OpEqual,
			Value:    eventID,
		},
	}

	var events []*marketplace.AuctionEvent
	if err := r.dbManager.List(ctx, filter, &events); err != nil {
		return nil, fmt.Errorf("failed to get auction event by event ID: %w", err)
	}

	if len(events) == 0 {
		return nil, nil
	}

	return events[0], nil
}

// LogEvent logs a generic auction event
func (r *auctionEventRepository) LogEvent(ctx context.Context, listingID string, eventType marketplace.AuctionEventType, actorID string, actorType marketplace.ActorType, eventData interface{}) error {
	event := marketplace.NewAuctionEvent(listingID, eventType, actorID, actorType, eventData)

	if err := event.SetEventData(eventData); err != nil {
		return fmt.Errorf("failed to set event data: %w", err)
	}

	return r.Create(ctx, event)
}

// LogListingCreated logs a listing creation event
func (r *auctionEventRepository) LogListingCreated(ctx context.Context, listingID, actorID string, listing *marketplace.Listing) error {
	event, err := marketplace.CreateListingCreatedEvent(listingID, actorID, listing)
	if err != nil {
		return fmt.Errorf("failed to create listing created event: %w", err)
	}

	return r.Create(ctx, event)
}

// LogBidPlaced logs a bid placement event
func (r *auctionEventRepository) LogBidPlaced(ctx context.Context, listingID, actorID string, bid *marketplace.Bid, previousHighest *float64) error {
	event, err := marketplace.CreateBidPlacedEvent(listingID, actorID, bid, previousHighest)
	if err != nil {
		return fmt.Errorf("failed to create bid placed event: %w", err)
	}

	return r.Create(ctx, event)
}

// LogBidOutbid logs a bid outbid event
func (r *auctionEventRepository) LogBidOutbid(ctx context.Context, listingID, actorID string, outbidBid *marketplace.Bid, newHighestBid *marketplace.Bid) error {
	eventData := marketplace.BidOutbidEventData{
		OutbidBidID:      outbidBid.BidID,
		OutbidBidderID:   outbidBid.BidderID,
		OutbidAmount:     outbidBid.BidAmount.InexactFloat64(),
		NewHighestBidID:  newHighestBid.BidID,
		NewHighestAmount: newHighestBid.BidAmount.InexactFloat64(),
	}

	return r.LogEvent(ctx, listingID, marketplace.EventBidOutbid, actorID, marketplace.ActorTypeSystem, eventData)
}

// LogAutoBidTriggered logs an auto-bid trigger event
func (r *auctionEventRepository) LogAutoBidTriggered(ctx context.Context, listingID, actorID string, originalBid, autoBid *marketplace.Bid) error {
	eventData := marketplace.AutoBidTriggeredEventData{
		OriginalBidID:    originalBid.BidID,
		AutoBidID:        autoBid.BidID,
		TriggerBidAmount: originalBid.BidAmount.InexactFloat64(),
		AutoBidAmount:    autoBid.BidAmount.InexactFloat64(),
	}

	if autoBid.AutoBidLimit != nil {
		eventData.AutoBidLimit = autoBid.AutoBidLimit.InexactFloat64()
	}

	return r.LogEvent(ctx, listingID, marketplace.EventAutoBidTriggered, actorID, marketplace.ActorTypeSystem, eventData)
}

// LogListingClosed logs a listing closure event
func (r *auctionEventRepository) LogListingClosed(ctx context.Context, listingID, actorID string, actorType marketplace.ActorType, closeReason string, winningBid *marketplace.Bid, totalBids int, finalStatus marketplace.ListingStatus) error {
	event, err := marketplace.CreateListingClosedEvent(listingID, actorID, actorType, closeReason, winningBid, totalBids, finalStatus)
	if err != nil {
		return fmt.Errorf("failed to create listing closed event: %w", err)
	}

	return r.Create(ctx, event)
}

// LogListingExpired logs a listing expiry event
func (r *auctionEventRepository) LogListingExpired(ctx context.Context, listingID string, winningBid *marketplace.Bid, totalBids int) error {
	eventData := marketplace.ListingExpiredEventData{
		ExpiredAt: time.Now(),
		TotalBids: totalBids,
		HasBids:   totalBids > 0,
	}

	if winningBid != nil {
		eventData.WinningBidID = &winningBid.BidID
		eventData.WinningBidderID = &winningBid.BidderID
		amount := winningBid.BidAmount.InexactFloat64()
		eventData.WinningAmount = &amount
	}

	return r.LogEvent(ctx, listingID, marketplace.EventListingExpired, "system", marketplace.ActorTypeSystem, eventData)
}

// LogBidRemoved logs a bid removal event
func (r *auctionEventRepository) LogBidRemoved(ctx context.Context, listingID, actorID string, removedBid *marketplace.Bid, removalReason string, newHighestBid *marketplace.Bid) error {
	eventData := marketplace.BidRemovedEventData{
		RemovedBidID:    removedBid.BidID,
		RemovedBidderID: removedBid.BidderID,
		RemovedAmount:   removedBid.BidAmount.InexactFloat64(),
		RemovalReason:   removalReason,
		WasHighestBid:   removedBid.IsHighestBid,
	}

	if newHighestBid != nil {
		eventData.NewHighestBidID = &newHighestBid.BidID
		amount := newHighestBid.BidAmount.InexactFloat64()
		eventData.NewHighestAmount = &amount
	}

	return r.LogEvent(ctx, listingID, marketplace.EventBidRemoved, actorID, marketplace.ActorTypeAdmin, eventData)
}

// GetEventsByListing retrieves events for a specific listing
func (r *auctionEventRepository) GetEventsByListing(ctx context.Context, listingID string, filter *marketplace.EventFilter, pagination *common.PaginationRequest) ([]*marketplace.AuctionEvent, int, error) {
	dbFilter := r.buildBaseFilter(filter)

	// Add listing filter
	dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
		Field:    "listing_id",
		Operator: base.OpEqual,
		Value:    listingID,
	})

	// Note: Ordering would be handled by the database layer in production

	return r.executeListQuery(ctx, dbFilter, pagination)
}

// GetByListingID is an alias for GetEventsByListing to match the interface
func (r *auctionEventRepository) GetByListingID(ctx context.Context, listingID string, pagination *common.PaginationRequest) ([]*marketplace.AuctionEvent, int, error) {
	return r.GetEventsByListing(ctx, listingID, nil, pagination)
}

// GetEventsByActor retrieves events for a specific actor
func (r *auctionEventRepository) GetEventsByActor(ctx context.Context, actorID string, filter *marketplace.EventFilter, pagination *common.PaginationRequest) ([]*marketplace.AuctionEvent, int, error) {
	dbFilter := r.buildBaseFilter(filter)

	// Add actor filter
	dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
		Field:    "actor_id",
		Operator: base.OpEqual,
		Value:    actorID,
	})

	// Note: Ordering would be handled by the database layer in production

	return r.executeListQuery(ctx, dbFilter, pagination)
}

// GetEventsByType retrieves events of a specific type
func (r *auctionEventRepository) GetEventsByType(ctx context.Context, eventType marketplace.AuctionEventType, filter *marketplace.EventFilter, pagination *common.PaginationRequest) ([]*marketplace.AuctionEvent, int, error) {
	dbFilter := r.buildBaseFilter(filter)

	// Add event type filter
	dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
		Field:    "event_type",
		Operator: base.OpEqual,
		Value:    eventType,
	})

	// Note: Ordering would be handled by the database layer in production

	return r.executeListQuery(ctx, dbFilter, pagination)
}

// GetByEventType is an alias for GetEventsByType to match the interface
func (r *auctionEventRepository) GetByEventType(ctx context.Context, eventType marketplace.AuctionEventType, pagination *common.PaginationRequest) ([]*marketplace.AuctionEvent, int, error) {
	return r.GetEventsByType(ctx, eventType, nil, pagination)
}

// GetEventsByTimeRange retrieves events within a time range
func (r *auctionEventRepository) GetEventsByTimeRange(ctx context.Context, startTime, endTime time.Time, filter *marketplace.EventFilter, pagination *common.PaginationRequest) ([]*marketplace.AuctionEvent, int, error) {
	dbFilter := r.buildBaseFilter(filter)

	// Add time range filter
	dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
		Field:    "timestamp",
		Operator: base.OpGreaterEqual,
		Value:    startTime,
	})
	dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
		Field:    "timestamp",
		Operator: base.OpLessEqual,
		Value:    endTime,
	})

	// Note: Ordering would be handled by the database layer in production

	return r.executeListQuery(ctx, dbFilter, pagination)
}

// GetRecentEvents retrieves recent events for a listing since a specific time
func (r *auctionEventRepository) GetRecentEvents(ctx context.Context, listingID string, since time.Time, limit int) ([]*marketplace.AuctionEvent, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "listing_id",
			Operator: base.OpEqual,
			Value:    listingID,
		},
		{
			Field:    "timestamp",
			Operator: base.OpGreaterThan,
			Value:    since,
		},
	}

	// Note: Ordering would be handled by the database layer in production
	filter.Limit = limit

	var events []*marketplace.AuctionEvent
	if err := r.dbManager.List(ctx, filter, &events); err != nil {
		return nil, fmt.Errorf("failed to get recent events: %w", err)
	}

	return events, nil
}

// GetEventsSince retrieves events since a specific time with filtering
func (r *auctionEventRepository) GetEventsSince(ctx context.Context, since time.Time, filter *marketplace.EventFilter, limit int) ([]*marketplace.AuctionEvent, error) {
	dbFilter := r.buildBaseFilter(filter)

	// Add since time filter
	dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
		Field:    "timestamp",
		Operator: base.OpGreaterThan,
		Value:    since,
	})

	// Note: Ordering would be handled by the database layer in production
	dbFilter.Limit = limit

	var events []*marketplace.AuctionEvent
	if err := r.dbManager.List(ctx, dbFilter, &events); err != nil {
		return nil, fmt.Errorf("failed to get events since: %w", err)
	}

	return events, nil
}

// GetEventSummary retrieves a summary of events for a listing
func (r *auctionEventRepository) GetEventSummary(ctx context.Context, listingID string) ([]*marketplace.EventSummary, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "listing_id",
			Operator: base.OpEqual,
			Value:    listingID,
		},
	}

	// Note: Ordering would be handled by the database layer in production

	var events []*marketplace.AuctionEvent
	if err := r.dbManager.List(ctx, filter, &events); err != nil {
		return nil, fmt.Errorf("failed to get events for summary: %w", err)
	}

	// Convert to summaries
	summaries := make([]*marketplace.EventSummary, len(events))
	for i, event := range events {
		summaries[i] = &marketplace.EventSummary{
			ID:        event.ID,
			EventID:   event.EventID,
			ListingID: event.ListingID,
			EventType: event.EventType,
			ActorID:   event.ActorID,
			ActorType: event.ActorType,
			Timestamp: event.Timestamp,
			Summary:   r.generateEventSummary(event),
		}
	}

	return summaries, nil
}

// GetEventStatistics retrieves event statistics for a listing
func (r *auctionEventRepository) GetEventStatistics(ctx context.Context, listingID string, timeRange *TimeRange) (*EventStatistics, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "listing_id",
			Operator: base.OpEqual,
			Value:    listingID,
		},
	}

	// Add time range filter if provided
	if timeRange != nil {
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "timestamp",
			Operator: base.OpGreaterEqual,
			Value:    timeRange.Start,
		})
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "timestamp",
			Operator: base.OpLessEqual,
			Value:    timeRange.End,
		})
	}

	var events []*marketplace.AuctionEvent
	if err := r.dbManager.List(ctx, filter, &events); err != nil {
		return nil, fmt.Errorf("failed to get events for statistics: %w", err)
	}

	stats := &EventStatistics{
		ListingID:    listingID,
		TotalEvents:  len(events),
		EventsByType: make(map[marketplace.AuctionEventType]int),
	}

	if len(events) == 0 {
		return stats, nil
	}

	// Calculate statistics
	uniqueActors := make(map[string]bool)
	var firstTime, lastTime *time.Time

	for _, event := range events {
		// Count by type
		stats.EventsByType[event.EventType]++

		// Track unique actors
		uniqueActors[event.ActorID] = true

		// Track time range
		if firstTime == nil || event.Timestamp.Before(*firstTime) {
			firstTime = &event.Timestamp
		}
		if lastTime == nil || event.Timestamp.After(*lastTime) {
			lastTime = &event.Timestamp
		}
	}

	stats.UniqueActors = len(uniqueActors)
	stats.FirstEventTime = firstTime
	stats.LastEventTime = lastTime

	// Calculate auction duration
	if firstTime != nil && lastTime != nil {
		duration := lastTime.Sub(*firstTime)
		stats.AuctionDuration = &duration
	}

	return stats, nil
}

// GetAuditTrail retrieves the complete audit trail for a listing
func (r *auctionEventRepository) GetAuditTrail(ctx context.Context, listingID string, pagination *common.PaginationRequest) ([]*marketplace.AuctionEvent, int, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "listing_id",
			Operator: base.OpEqual,
			Value:    listingID,
		},
	}

	// Note: Ordering would be handled by the database layer in production

	return r.executeListQuery(ctx, filter, pagination)
}

// GetAllEvents retrieves all events (admin operation)
func (r *auctionEventRepository) GetAllEvents(ctx context.Context, filter *marketplace.EventFilter, pagination *common.PaginationRequest) ([]*marketplace.AuctionEvent, int, error) {
	dbFilter := r.buildBaseFilter(filter)

	// Note: Ordering would be handled by the database layer in production

	return r.executeListQuery(ctx, dbFilter, pagination)
}

// DeleteEventsByListing deletes all events for a listing (admin operation)
func (r *auctionEventRepository) DeleteEventsByListing(ctx context.Context, listingID string) error {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "listing_id",
			Operator: base.OpEqual,
			Value:    listingID,
		},
	}

	var events []*marketplace.AuctionEvent
	if err := r.dbManager.List(ctx, filter, &events); err != nil {
		return fmt.Errorf("failed to get events for deletion: %w", err)
	}

	// Delete each event
	for _, event := range events {
		if err := r.dbManager.Delete(ctx, event.ID, event); err != nil {
			return fmt.Errorf("failed to delete event %s: %w", event.EventID, err)
		}
	}

	return nil
}

// Helper methods

// buildBaseFilter builds a base filter from the event filter
func (r *auctionEventRepository) buildBaseFilter(filter *marketplace.EventFilter) *base.Filter {
	dbFilter := base.NewFilter()

	if filter == nil {
		return dbFilter
	}

	// Listing ID filter
	if filter.ListingID != "" {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "listing_id",
			Operator: base.OpEqual,
			Value:    filter.ListingID,
		})
	}

	// Event type filter
	if filter.EventType != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "event_type",
			Operator: base.OpEqual,
			Value:    *filter.EventType,
		})
	}

	// Actor ID filter
	if filter.ActorID != "" {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "actor_id",
			Operator: base.OpEqual,
			Value:    filter.ActorID,
		})
	}

	// Actor type filter
	if filter.ActorType != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "actor_type",
			Operator: base.OpEqual,
			Value:    *filter.ActorType,
		})
	}

	// Time range filters
	if filter.TimeFrom != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "timestamp",
			Operator: base.OpGreaterEqual,
			Value:    *filter.TimeFrom,
		})
	}

	if filter.TimeTo != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "timestamp",
			Operator: base.OpLessEqual,
			Value:    *filter.TimeTo,
		})
	}

	return dbFilter
}

// executeListQuery executes a list query with pagination
func (r *auctionEventRepository) executeListQuery(ctx context.Context, filter *base.Filter, pagination *common.PaginationRequest) ([]*marketplace.AuctionEvent, int, error) {
	// Apply pagination
	if pagination != nil {
		filter.Limit = pagination.Limit
		filter.Offset = pagination.Offset
	}

	// Execute query
	var events []*marketplace.AuctionEvent
	if err := r.dbManager.List(ctx, filter, &events); err != nil {
		return nil, 0, fmt.Errorf("failed to execute event query: %w", err)
	}

	// Get total count for pagination
	total, err := r.dbManager.Count(ctx, filter, &marketplace.AuctionEvent{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return events, int(total), nil
}

// generateEventSummary generates a human-readable summary for an event
func (r *auctionEventRepository) generateEventSummary(event *marketplace.AuctionEvent) string {
	switch event.EventType {
	case marketplace.EventListingCreated:
		return "Listing created"
	case marketplace.EventBidPlaced:
		return "Bid placed"
	case marketplace.EventBidOutbid:
		return "Bid outbid by higher bid"
	case marketplace.EventAutoBidTriggered:
		return "Auto-bid triggered"
	case marketplace.EventListingClosed:
		return "Listing closed"
	case marketplace.EventListingExpired:
		return "Listing expired"
	case marketplace.EventListingCancelled:
		return "Listing cancelled"
	case marketplace.EventBidRemoved:
		return "Bid removed by admin"
	case marketplace.EventListingUpdated:
		return "Listing updated"
	default:
		return string(event.EventType)
	}
}

// GetEvents is an alias for GetAllEvents for compatibility
func (r *auctionEventRepository) GetEvents(ctx context.Context, filter *marketplace.EventFilter, pagination *common.PaginationRequest) ([]*marketplace.AuctionEvent, int, error) {
	return r.GetAllEvents(ctx, filter, pagination)
}

// ArchiveEventsBefore archives events before a certain date
func (r *auctionEventRepository) ArchiveEventsBefore(ctx context.Context, beforeDate time.Time) (int, error) {
	// Stub implementation - return 0 for now
	return 0, nil
}

// DeleteEventsBefore deletes events before a certain date
func (r *auctionEventRepository) DeleteEventsBefore(ctx context.Context, beforeDate time.Time) (int, error) {
	// Stub implementation - return 0 for now
	return 0, nil
}
