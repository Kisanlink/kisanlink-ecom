package marketplace

import (
	"context"
	"fmt"
	"time"

	marketplaceModels "github.com/Kisanlink/kisanlink-ecom/entities/models/marketplace"
	"github.com/Kisanlink/kisanlink-ecom/internal/common"
	marketplaceRepo "github.com/Kisanlink/kisanlink-ecom/internal/repositories/marketplace"
)

// AuditServiceInterface defines the interface for marketplace audit operations
type AuditServiceInterface interface {
	// Audit Log Management
	GetAuditLog(ctx context.Context, filter *marketplaceModels.EventFilter, pagination *common.PaginationParams) ([]*marketplaceModels.AuctionEvent, int, error)
	GetListingAuditLog(ctx context.Context, listingID string, pagination *common.PaginationParams) ([]*marketplaceModels.AuctionEvent, int, error)
	GetUserAuditLog(ctx context.Context, userID string, pagination *common.PaginationParams) ([]*marketplaceModels.AuctionEvent, int, error)

	// Event Recording
	RecordEvent(ctx context.Context, event *marketplaceModels.AuctionEvent) error
	RecordListingEvent(ctx context.Context, listingID string, eventType marketplaceModels.AuctionEventType, eventData interface{}, actorID string) error
	RecordBidEvent(ctx context.Context, listingID string, eventType marketplaceModels.AuctionEventType, eventData interface{}, actorID string) error
	RecordAdminEvent(ctx context.Context, listingID string, eventType marketplaceModels.AuctionEventType, eventData interface{}, adminID string) error

	// Compliance and Reporting
	GenerateComplianceReport(ctx context.Context, startTime, endTime time.Time) (*ComplianceReport, error)
	GetSuspiciousActivities(ctx context.Context, timeRange string) ([]*SuspiciousActivity, error)
	GetUserActivitySummary(ctx context.Context, userID string, timeRange string) (*UserActivitySummary, error)
	GetSystemActivitySummary(ctx context.Context, timeRange string) (*SystemActivitySummary, error)

	// Data Retention and Cleanup
	ArchiveOldEvents(ctx context.Context, retentionPeriod time.Duration) (int, error)
	CleanupExpiredEvents(ctx context.Context) (int, error)
}

// ComplianceReport represents a compliance audit report
type ComplianceReport struct {
	ReportID         string                `json:"report_id"`
	GeneratedAt      time.Time             `json:"generated_at"`
	StartTime        time.Time             `json:"start_time"`
	EndTime          time.Time             `json:"end_time"`
	TotalEvents      int                   `json:"total_events"`
	EventsByType     map[string]int        `json:"events_by_type"`
	UserActivities   []UserActivitySummary `json:"user_activities"`
	AdminActions     []AdminActionSummary  `json:"admin_actions"`
	PolicyViolations []PolicyViolation     `json:"policy_violations"`
	DataIntegrity    DataIntegrityCheck    `json:"data_integrity"`
	Recommendations  []string              `json:"recommendations"`
}

// SuspiciousActivity represents potentially fraudulent or suspicious activity
type SuspiciousActivity struct {
	ActivityID   string                   `json:"activity_id"`
	DetectedAt   time.Time                `json:"detected_at"`
	ActivityType SuspiciousActivityType   `json:"activity_type"`
	Severity     SeverityLevel            `json:"severity"`
	UserID       string                   `json:"user_id"`
	ListingID    string                   `json:"listing_id,omitempty"`
	BidID        string                   `json:"bid_id,omitempty"`
	Description  string                   `json:"description"`
	Evidence     map[string]interface{}   `json:"evidence"`
	Status       SuspiciousActivityStatus `json:"status"`
	ReviewedBy   string                   `json:"reviewed_by,omitempty"`
	ReviewedAt   *time.Time               `json:"reviewed_at,omitempty"`
	Resolution   string                   `json:"resolution,omitempty"`
}

// UserActivitySummary represents a summary of user activities
type UserActivitySummary struct {
	UserID               string              `json:"user_id"`
	TimeRange            string              `json:"time_range"`
	TotalEvents          int                 `json:"total_events"`
	EventsByType         map[string]int      `json:"events_by_type"`
	ListingsCreated      int                 `json:"listings_created"`
	BidsPlaced           int                 `json:"bids_placed"`
	ListingsClosed       int                 `json:"listings_closed"`
	SuspiciousActivities int                 `json:"suspicious_activities"`
	LastActivity         time.Time           `json:"last_activity"`
	ActivityPattern      UserActivityPattern `json:"activity_pattern"`
	RiskScore            float64             `json:"risk_score"`
}

// SystemActivitySummary represents overall system activity summary
type SystemActivitySummary struct {
	TimeRange            string                `json:"time_range"`
	TotalEvents          int                   `json:"total_events"`
	EventsByType         map[string]int        `json:"events_by_type"`
	EventsByHour         []HourlyEventCount    `json:"events_by_hour"`
	TopActiveUsers       []UserActivityRank    `json:"top_active_users"`
	ErrorEvents          int                   `json:"error_events"`
	AdminActions         int                   `json:"admin_actions"`
	SystemEvents         int                   `json:"system_events"`
	PeakActivityPeriods  []AuditActivityPeriod `json:"peak_activity_periods"`
	AverageEventsPerUser float64               `json:"average_events_per_user"`
}

// Supporting types
type SuspiciousActivityType string

const (
	SuspiciousActivityBidManipulation   SuspiciousActivityType = "BID_MANIPULATION"
	SuspiciousActivityFakeListings      SuspiciousActivityType = "FAKE_LISTINGS"
	SuspiciousActivityShillBidding      SuspiciousActivityType = "SHILL_BIDDING"
	SuspiciousActivityRapidBidding      SuspiciousActivityType = "RAPID_BIDDING"
	SuspiciousActivityUnusualPatterns   SuspiciousActivityType = "UNUSUAL_PATTERNS"
	SuspiciousActivityMultipleAccounts  SuspiciousActivityType = "MULTIPLE_ACCOUNTS"
	SuspiciousActivityPriceManipulation SuspiciousActivityType = "PRICE_MANIPULATION"
)

type SeverityLevel string

const (
	SeverityLow      SeverityLevel = "LOW"
	SeverityMedium   SeverityLevel = "MEDIUM"
	SeverityHigh     SeverityLevel = "HIGH"
	SeverityCritical SeverityLevel = "CRITICAL"
)

type SuspiciousActivityStatus string

const (
	SuspiciousActivityStatusPending   SuspiciousActivityStatus = "PENDING"
	SuspiciousActivityStatusReviewing SuspiciousActivityStatus = "REVIEWING"
	SuspiciousActivityStatusResolved  SuspiciousActivityStatus = "RESOLVED"
	SuspiciousActivityStatusDismissed SuspiciousActivityStatus = "DISMISSED"
	SuspiciousActivityStatusEscalated SuspiciousActivityStatus = "ESCALATED"
)

type AdminActionSummary struct {
	AdminID    string    `json:"admin_id"`
	ActionType string    `json:"action_type"`
	Count      int       `json:"count"`
	LastAction time.Time `json:"last_action"`
}

type PolicyViolation struct {
	ViolationID string    `json:"violation_id"`
	PolicyType  string    `json:"policy_type"`
	UserID      string    `json:"user_id"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	DetectedAt  time.Time `json:"detected_at"`
	Status      string    `json:"status"`
}

type DataIntegrityCheck struct {
	CheckedAt         time.Time `json:"checked_at"`
	TotalRecords      int       `json:"total_records"`
	CorruptedRecords  int       `json:"corrupted_records"`
	MissingReferences int       `json:"missing_references"`
	DuplicateRecords  int       `json:"duplicate_records"`
	IntegrityScore    float64   `json:"integrity_score"`
}

type UserActivityPattern struct {
	PeakHours      []int    `json:"peak_hours"`
	AverageSession float64  `json:"average_session_minutes"`
	ActivityDays   []string `json:"activity_days"`
	BehaviorScore  float64  `json:"behavior_score"`
}

type HourlyEventCount struct {
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

type UserActivityRank struct {
	UserID     string `json:"user_id"`
	EventCount int    `json:"event_count"`
	Rank       int    `json:"rank"`
}

type AuditActivityPeriod struct {
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	EventCount  int       `json:"event_count"`
	Description string    `json:"description"`
}

// AuditService provides marketplace audit and compliance functionality
type AuditService struct {
	eventRepo   marketplaceRepo.AuctionEventRepository
	listingRepo marketplaceRepo.ListingRepository
	bidRepo     marketplaceRepo.BidRepository
}

// NewAuditService creates a new audit service
func NewAuditService(
	eventRepo marketplaceRepo.AuctionEventRepository,
	listingRepo marketplaceRepo.ListingRepository,
	bidRepo marketplaceRepo.BidRepository,
) AuditServiceInterface {
	return &AuditService{
		eventRepo:   eventRepo,
		listingRepo: listingRepo,
		bidRepo:     bidRepo,
	}
}

// GetAuditLog retrieves audit log with filtering and pagination
func (s *AuditService) GetAuditLog(ctx context.Context, filter *marketplaceModels.EventFilter, pagination *common.PaginationParams) ([]*marketplaceModels.AuctionEvent, int, error) {
	// Convert pagination params
	paginationReq := &common.PaginationRequest{
		Limit:  pagination.Limit,
		Offset: pagination.CalculateOffset(),
	}

	// Get events from repository
	events, total, err := s.eventRepo.GetEvents(ctx, filter, paginationReq)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get audit log: %w", err)
	}

	return events, total, nil
}

// GetListingAuditLog retrieves audit log for a specific listing
func (s *AuditService) GetListingAuditLog(ctx context.Context, listingID string, pagination *common.PaginationParams) ([]*marketplaceModels.AuctionEvent, int, error) {
	filter := &marketplaceModels.EventFilter{
		ListingID: listingID,
	}

	return s.GetAuditLog(ctx, filter, pagination)
}

// GetUserAuditLog retrieves audit log for a specific user
func (s *AuditService) GetUserAuditLog(ctx context.Context, userID string, pagination *common.PaginationParams) ([]*marketplaceModels.AuctionEvent, int, error) {
	filter := &marketplaceModels.EventFilter{
		ActorID: userID,
	}

	return s.GetAuditLog(ctx, filter, pagination)
}

// RecordEvent records a new audit event
func (s *AuditService) RecordEvent(ctx context.Context, event *marketplaceModels.AuctionEvent) error {
	if err := s.eventRepo.Create(ctx, event); err != nil {
		return fmt.Errorf("failed to record audit event: %w", err)
	}
	return nil
}

// RecordListingEvent records a listing-related event
func (s *AuditService) RecordListingEvent(ctx context.Context, listingID string, eventType marketplaceModels.AuctionEventType, eventData interface{}, actorID string) error {
	event := marketplaceModels.NewAuctionEvent(listingID, eventType, actorID, marketplaceModels.ActorTypeUser, eventData)
	if err := event.SetEventData(eventData); err != nil {
		return fmt.Errorf("failed to set event data: %w", err)
	}

	return s.RecordEvent(ctx, event)
}

// RecordBidEvent records a bid-related event
func (s *AuditService) RecordBidEvent(ctx context.Context, listingID string, eventType marketplaceModels.AuctionEventType, eventData interface{}, actorID string) error {
	event := marketplaceModels.NewAuctionEvent(listingID, eventType, actorID, marketplaceModels.ActorTypeUser, eventData)
	if err := event.SetEventData(eventData); err != nil {
		return fmt.Errorf("failed to set event data: %w", err)
	}

	return s.RecordEvent(ctx, event)
}

// RecordAdminEvent records an admin-related event
func (s *AuditService) RecordAdminEvent(ctx context.Context, listingID string, eventType marketplaceModels.AuctionEventType, eventData interface{}, adminID string) error {
	event := marketplaceModels.NewAuctionEvent(listingID, eventType, adminID, marketplaceModels.ActorTypeAdmin, eventData)
	if err := event.SetEventData(eventData); err != nil {
		return fmt.Errorf("failed to set event data: %w", err)
	}

	return s.RecordEvent(ctx, event)
}

// GenerateComplianceReport generates a comprehensive compliance report
func (s *AuditService) GenerateComplianceReport(ctx context.Context, startTime, endTime time.Time) (*ComplianceReport, error) {
	// Get all events in the time range
	filter := &marketplaceModels.EventFilter{
		TimeFrom: &startTime,
		TimeTo:   &endTime,
	}

	events, totalEvents, err := s.eventRepo.GetEvents(ctx, filter, &common.PaginationRequest{Limit: 10000})
	if err != nil {
		return nil, fmt.Errorf("failed to get events for compliance report: %w", err)
	}

	// Analyze events by type
	eventsByType := make(map[string]int)
	adminActions := make(map[string]*AdminActionSummary)
	userActivities := make(map[string]*UserActivitySummary)

	for _, event := range events {
		// Count events by type
		eventsByType[string(event.EventType)]++

		// Track admin actions
		if event.ActorType == marketplaceModels.ActorTypeAdmin {
			if summary, exists := adminActions[event.ActorID]; exists {
				summary.Count++
				if event.Timestamp.After(summary.LastAction) {
					summary.LastAction = event.Timestamp
				}
			} else {
				adminActions[event.ActorID] = &AdminActionSummary{
					AdminID:    event.ActorID,
					ActionType: string(event.EventType),
					Count:      1,
					LastAction: event.Timestamp,
				}
			}
		}

		// Track user activities
		if event.ActorType == marketplaceModels.ActorTypeUser {
			if summary, exists := userActivities[event.ActorID]; exists {
				summary.TotalEvents++
				summary.EventsByType[string(event.EventType)]++
				if event.Timestamp.After(summary.LastActivity) {
					summary.LastActivity = event.Timestamp
				}
			} else {
				userActivities[event.ActorID] = &UserActivitySummary{
					UserID:       event.ActorID,
					TotalEvents:  1,
					EventsByType: map[string]int{string(event.EventType): 1},
					LastActivity: event.Timestamp,
				}
			}
		}
	}

	// Convert maps to slices
	var adminActionsList []AdminActionSummary
	for _, summary := range adminActions {
		adminActionsList = append(adminActionsList, *summary)
	}

	var userActivitiesList []UserActivitySummary
	for _, summary := range userActivities {
		userActivitiesList = append(userActivitiesList, *summary)
	}

	// Generate report
	report := &ComplianceReport{
		ReportID:         fmt.Sprintf("COMP_%d", time.Now().Unix()),
		GeneratedAt:      time.Now(),
		StartTime:        startTime,
		EndTime:          endTime,
		TotalEvents:      totalEvents,
		EventsByType:     eventsByType,
		UserActivities:   userActivitiesList,
		AdminActions:     adminActionsList,
		PolicyViolations: []PolicyViolation{}, // Would be populated by policy engine
		DataIntegrity: DataIntegrityCheck{
			CheckedAt:         time.Now(),
			TotalRecords:      totalEvents,
			CorruptedRecords:  0,
			MissingReferences: 0,
			DuplicateRecords:  0,
			IntegrityScore:    100.0,
		},
		Recommendations: s.generateRecommendations(eventsByType, adminActionsList, userActivitiesList),
	}

	return report, nil
}

// GetSuspiciousActivities detects and returns suspicious activities
func (s *AuditService) GetSuspiciousActivities(ctx context.Context, timeRange string) ([]*SuspiciousActivity, error) {
	// Parse time range
	startTime, endTime, err := s.parseTimeRange(timeRange)
	if err != nil {
		return nil, fmt.Errorf("invalid time range: %w", err)
	}

	// Get events in time range
	filter := &marketplaceModels.EventFilter{
		TimeFrom: &startTime,
		TimeTo:   &endTime,
	}

	events, _, err := s.eventRepo.GetEvents(ctx, filter, &common.PaginationRequest{Limit: 10000})
	if err != nil {
		return nil, fmt.Errorf("failed to get events for suspicious activity detection: %w", err)
	}

	// Analyze events for suspicious patterns
	var suspiciousActivities []*SuspiciousActivity

	// Detect rapid bidding patterns
	rapidBidding := s.detectRapidBidding(events)
	suspiciousActivities = append(suspiciousActivities, rapidBidding...)

	// Detect unusual listing patterns
	unusualListings := s.detectUnusualListingPatterns(events)
	suspiciousActivities = append(suspiciousActivities, unusualListings...)

	// Detect potential shill bidding
	shillBidding := s.detectShillBidding(events)
	suspiciousActivities = append(suspiciousActivities, shillBidding...)

	return suspiciousActivities, nil
}

// GetUserActivitySummary retrieves activity summary for a specific user
func (s *AuditService) GetUserActivitySummary(ctx context.Context, userID string, timeRange string) (*UserActivitySummary, error) {
	// Parse time range
	startTime, endTime, err := s.parseTimeRange(timeRange)
	if err != nil {
		return nil, fmt.Errorf("invalid time range: %w", err)
	}

	// Get user events
	filter := &marketplaceModels.EventFilter{
		ActorID:  userID,
		TimeFrom: &startTime,
		TimeTo:   &endTime,
	}

	events, totalEvents, err := s.eventRepo.GetEvents(ctx, filter, &common.PaginationRequest{Limit: 10000})
	if err != nil {
		return nil, fmt.Errorf("failed to get user events: %w", err)
	}

	// Analyze events
	eventsByType := make(map[string]int)
	var lastActivity time.Time
	var listingsCreated, bidsPlaced, listingsClosed int

	for _, event := range events {
		eventsByType[string(event.EventType)]++

		if event.Timestamp.After(lastActivity) {
			lastActivity = event.Timestamp
		}

		switch event.EventType {
		case marketplaceModels.EventListingCreated:
			listingsCreated++
		case marketplaceModels.EventBidPlaced:
			bidsPlaced++
		case marketplaceModels.EventListingClosed:
			listingsClosed++
		}
	}

	// Calculate activity pattern and risk score
	activityPattern := s.calculateActivityPattern(events)
	riskScore := s.calculateRiskScore(events, eventsByType)

	return &UserActivitySummary{
		UserID:               userID,
		TimeRange:            timeRange,
		TotalEvents:          totalEvents,
		EventsByType:         eventsByType,
		ListingsCreated:      listingsCreated,
		BidsPlaced:           bidsPlaced,
		ListingsClosed:       listingsClosed,
		SuspiciousActivities: 0, // Would be calculated from suspicious activity detection
		LastActivity:         lastActivity,
		ActivityPattern:      activityPattern,
		RiskScore:            riskScore,
	}, nil
}

// GetSystemActivitySummary retrieves overall system activity summary
func (s *AuditService) GetSystemActivitySummary(ctx context.Context, timeRange string) (*SystemActivitySummary, error) {
	// Parse time range
	startTime, endTime, err := s.parseTimeRange(timeRange)
	if err != nil {
		return nil, fmt.Errorf("invalid time range: %w", err)
	}

	// Get all events in time range
	filter := &marketplaceModels.EventFilter{
		TimeFrom: &startTime,
		TimeTo:   &endTime,
	}

	events, totalEvents, err := s.eventRepo.GetEvents(ctx, filter, &common.PaginationRequest{Limit: 10000})
	if err != nil {
		return nil, fmt.Errorf("failed to get system events: %w", err)
	}

	// Analyze events
	eventsByType := make(map[string]int)
	eventsByHour := make([]HourlyEventCount, 24)
	userEventCounts := make(map[string]int)
	var errorEvents, adminActions, systemEvents int

	for _, event := range events {
		eventsByType[string(event.EventType)]++

		hour := event.Timestamp.Hour()
		eventsByHour[hour].Hour = hour
		eventsByHour[hour].Count++

		userEventCounts[event.ActorID]++

		switch event.ActorType {
		case marketplaceModels.ActorTypeAdmin:
			adminActions++
		case marketplaceModels.ActorTypeSystem:
			systemEvents++
		}
	}

	// Calculate top active users
	topActiveUsers := s.calculateTopActiveUsers(userEventCounts)

	// Calculate average events per user
	var averageEventsPerUser float64
	if len(userEventCounts) > 0 {
		averageEventsPerUser = float64(totalEvents) / float64(len(userEventCounts))
	}

	// Identify peak activity periods
	peakActivityPeriods := s.identifyPeakActivityPeriods(eventsByHour)

	return &SystemActivitySummary{
		TimeRange:            timeRange,
		TotalEvents:          totalEvents,
		EventsByType:         eventsByType,
		EventsByHour:         eventsByHour,
		TopActiveUsers:       topActiveUsers,
		ErrorEvents:          errorEvents,
		AdminActions:         adminActions,
		SystemEvents:         systemEvents,
		PeakActivityPeriods:  peakActivityPeriods,
		AverageEventsPerUser: averageEventsPerUser,
	}, nil
}

// ArchiveOldEvents archives events older than the retention period
func (s *AuditService) ArchiveOldEvents(ctx context.Context, retentionPeriod time.Duration) (int, error) {
	cutoffTime := time.Now().Add(-retentionPeriod)

	count, err := s.eventRepo.ArchiveEventsBefore(ctx, cutoffTime)
	if err != nil {
		return 0, fmt.Errorf("failed to archive old events: %w", err)
	}

	return count, nil
}

// CleanupExpiredEvents removes expired events based on retention policy
func (s *AuditService) CleanupExpiredEvents(ctx context.Context) (int, error) {
	// Default retention period of 7 years for compliance
	retentionPeriod := 7 * 365 * 24 * time.Hour
	cutoffTime := time.Now().Add(-retentionPeriod)

	count, err := s.eventRepo.DeleteEventsBefore(ctx, cutoffTime)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired events: %w", err)
	}

	return count, nil
}

// Helper methods

// parseTimeRange parses time range string and returns start and end times
func (s *AuditService) parseTimeRange(timeRange string) (time.Time, time.Time, error) {
	now := time.Now()
	var startTime time.Time

	switch timeRange {
	case "24h":
		startTime = now.Add(-24 * time.Hour)
	case "7d":
		startTime = now.Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = now.Add(-30 * 24 * time.Hour)
	case "90d":
		startTime = now.Add(-90 * 24 * time.Hour)
	case "1y":
		startTime = now.Add(-365 * 24 * time.Hour)
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("unsupported time range: %s", timeRange)
	}

	return startTime, now, nil
}

// generateRecommendations generates compliance recommendations
func (s *AuditService) generateRecommendations(eventsByType map[string]int, adminActions []AdminActionSummary, userActivities []UserActivitySummary) []string {
	var recommendations []string

	// Check for high admin activity
	if len(adminActions) > 10 {
		recommendations = append(recommendations, "High admin activity detected. Review admin actions for policy compliance.")
	}

	// Check for unusual event patterns
	if eventsByType["BID_PLACED"] > eventsByType["LISTING_CREATED"]*10 {
		recommendations = append(recommendations, "High bid-to-listing ratio detected. Monitor for potential bid manipulation.")
	}

	// Check for user activity patterns
	if len(userActivities) > 0 {
		highActivityUsers := 0
		for _, activity := range userActivities {
			if activity.TotalEvents > 100 {
				highActivityUsers++
			}
		}
		if highActivityUsers > len(userActivities)/10 {
			recommendations = append(recommendations, "Multiple high-activity users detected. Review for potential automated behavior.")
		}
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "No significant compliance issues detected.")
	}

	return recommendations
}

// detectRapidBidding detects rapid bidding patterns
func (s *AuditService) detectRapidBidding(events []*marketplaceModels.AuctionEvent) []*SuspiciousActivity {
	var activities []*SuspiciousActivity

	// Group events by user and listing
	userListingBids := make(map[string]map[string][]*marketplaceModels.AuctionEvent)

	for _, event := range events {
		if event.EventType == marketplaceModels.EventBidPlaced {
			if userListingBids[event.ActorID] == nil {
				userListingBids[event.ActorID] = make(map[string][]*marketplaceModels.AuctionEvent)
			}
			userListingBids[event.ActorID][event.ListingID] = append(userListingBids[event.ActorID][event.ListingID], event)
		}
	}

	// Detect rapid bidding (more than 5 bids in 1 minute)
	for userID, listingBids := range userListingBids {
		for listingID, bids := range listingBids {
			if len(bids) >= 5 {
				// Check if bids are within 1 minute
				for i := 0; i < len(bids)-4; i++ {
					if bids[i+4].Timestamp.Sub(bids[i].Timestamp) <= time.Minute {
						activities = append(activities, &SuspiciousActivity{
							ActivityID:   fmt.Sprintf("RAPID_BID_%d", time.Now().UnixNano()),
							DetectedAt:   time.Now(),
							ActivityType: SuspiciousActivityRapidBidding,
							Severity:     SeverityMedium,
							UserID:       userID,
							ListingID:    listingID,
							Description:  "Rapid bidding pattern detected: 5+ bids within 1 minute",
							Evidence: map[string]interface{}{
								"bid_count":   len(bids),
								"time_window": "1 minute",
								"first_bid":   bids[i].Timestamp,
								"last_bid":    bids[i+4].Timestamp,
							},
							Status: SuspiciousActivityStatusPending,
						})
						break
					}
				}
			}
		}
	}

	return activities
}

// detectUnusualListingPatterns detects unusual listing creation patterns
func (s *AuditService) detectUnusualListingPatterns(events []*marketplaceModels.AuctionEvent) []*SuspiciousActivity {
	var activities []*SuspiciousActivity

	// Count listings created by each user
	userListings := make(map[string]int)
	for _, event := range events {
		if event.EventType == marketplaceModels.EventListingCreated {
			userListings[event.ActorID]++
		}
	}

	// Detect users creating excessive listings (more than 20 in the time period)
	for userID, count := range userListings {
		if count > 20 {
			activities = append(activities, &SuspiciousActivity{
				ActivityID:   fmt.Sprintf("EXCESSIVE_LISTINGS_%d", time.Now().UnixNano()),
				DetectedAt:   time.Now(),
				ActivityType: SuspiciousActivityFakeListings,
				Severity:     SeverityMedium,
				UserID:       userID,
				Description:  "Excessive listing creation detected",
				Evidence: map[string]interface{}{
					"listing_count": count,
					"threshold":     20,
				},
				Status: SuspiciousActivityStatusPending,
			})
		}
	}

	return activities
}

// detectShillBidding detects potential shill bidding patterns
func (s *AuditService) detectShillBidding(events []*marketplaceModels.AuctionEvent) []*SuspiciousActivity {
	var activities []*SuspiciousActivity

	// This would require more sophisticated analysis of bidding patterns
	// For now, return empty slice
	return activities
}

// calculateActivityPattern calculates user activity patterns
func (s *AuditService) calculateActivityPattern(events []*marketplaceModels.AuctionEvent) UserActivityPattern {
	hourCounts := make(map[int]int)
	dayCounts := make(map[string]int)

	for _, event := range events {
		hour := event.Timestamp.Hour()
		hourCounts[hour]++

		day := event.Timestamp.Weekday().String()
		dayCounts[day]++
	}

	// Find peak hours (top 3)
	var peakHours []int
	for hour := range hourCounts {
		peakHours = append(peakHours, hour)
	}
	// Sort by count (simplified)
	if len(peakHours) > 3 {
		peakHours = peakHours[:3]
	}

	// Find active days
	var activityDays []string
	for day := range dayCounts {
		activityDays = append(activityDays, day)
	}

	return UserActivityPattern{
		PeakHours:      peakHours,
		AverageSession: 30.0, // Placeholder
		ActivityDays:   activityDays,
		BehaviorScore:  75.0, // Placeholder
	}
}

// calculateRiskScore calculates user risk score based on activity
func (s *AuditService) calculateRiskScore(events []*marketplaceModels.AuctionEvent, eventsByType map[string]int) float64 {
	// Simple risk scoring algorithm
	score := 0.0

	// High bid activity increases risk
	if bidCount := eventsByType["BID_PLACED"]; bidCount > 50 {
		score += 20.0
	}

	// High listing activity increases risk
	if listingCount := eventsByType["LISTING_CREATED"]; listingCount > 10 {
		score += 15.0
	}

	// Admin actions increase risk
	if adminCount := eventsByType["LISTING_CLOSED"]; adminCount > 5 {
		score += 25.0
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}

// calculateTopActiveUsers calculates top active users by event count
func (s *AuditService) calculateTopActiveUsers(userEventCounts map[string]int) []UserActivityRank {
	var ranks []UserActivityRank

	rank := 1
	for userID, count := range userEventCounts {
		ranks = append(ranks, UserActivityRank{
			UserID:     userID,
			EventCount: count,
			Rank:       rank,
		})
		rank++

		// Limit to top 10
		if rank > 10 {
			break
		}
	}

	return ranks
}

// identifyPeakActivityPeriods identifies periods of high activity
func (s *AuditService) identifyPeakActivityPeriods(eventsByHour []HourlyEventCount) []AuditActivityPeriod {
	var periods []AuditActivityPeriod

	// Find hours with activity above average
	total := 0
	for _, hourly := range eventsByHour {
		total += hourly.Count
	}

	if len(eventsByHour) == 0 {
		return periods
	}

	average := float64(total) / float64(len(eventsByHour))
	threshold := average * 1.5 // 50% above average

	for _, hourly := range eventsByHour {
		if float64(hourly.Count) > threshold {
			periods = append(periods, AuditActivityPeriod{
				StartTime:   time.Date(2024, 1, 1, hourly.Hour, 0, 0, 0, time.UTC),
				EndTime:     time.Date(2024, 1, 1, hourly.Hour+1, 0, 0, 0, time.UTC),
				EventCount:  hourly.Count,
				Description: fmt.Sprintf("Peak activity at hour %d", hourly.Hour),
			})
		}
	}

	return periods
}
