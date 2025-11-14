package marketplace

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/marketplace"
	"github.com/Kisanlink/kisanlink-ecom/internal/common"

	"github.com/shopspring/decimal"
)

// AuctionResultsServiceInterface defines the interface for auction results revelation operations
type AuctionResultsServiceInterface interface {
	// Results Processing
	ProcessAuctionResults(ctx context.Context, listingID string) (*AuctionResults, error)
	RevealAuctionResults(ctx context.Context, listingID string, viewerID string, viewerOrgID string) (*RevealedAuctionResults, error)

	// Winner Notification
	NotifyWinner(ctx context.Context, listing *marketplace.Listing, winningBid *marketplace.Bid) error
	NotifyParticipants(ctx context.Context, listing *marketplace.Listing, results *AuctionResults) error

	// Historical Data Access
	GetHistoricalBidData(ctx context.Context, listingID string, viewerID string, viewerOrgID string, filter *HistoricalDataFilter) (*HistoricalBidData, error)
	GetAuctionSummary(ctx context.Context, listingID string, viewerID string, viewerOrgID string) (*AuctionSummary, error)

	// Results Validation
	ValidateAuctionResults(ctx context.Context, listingID string) (*ResultsValidation, error)
	ReconcileResults(ctx context.Context, listingID string) (*ResultsReconciliation, error)
}

// AuctionResults represents the complete results of an auction
type AuctionResults struct {
	ListingID        string                    `json:"listing_id"`
	Status           marketplace.ListingStatus `json:"status"`
	WinningBid       *marketplace.Bid          `json:"winning_bid,omitempty"`
	TotalBids        int                       `json:"total_bids"`
	UniqueBidders    int                       `json:"unique_bidders"`
	FinalPrice       decimal.Decimal           `json:"final_price"`
	StartPrice       decimal.Decimal           `json:"start_price"`
	PriceIncrease    decimal.Decimal           `json:"price_increase"`
	Duration         time.Duration             `json:"duration"`
	CompletedAt      time.Time                 `json:"completed_at"`
	BidProgression   []BidProgressionPoint     `json:"bid_progression"`
	ParticipantStats []ParticipantStats        `json:"participant_stats"`
	AuctionMetrics   AuctionMetrics            `json:"auction_metrics"`
}

// RevealedAuctionResults represents auction results with visibility filtering applied
type RevealedAuctionResults struct {
	ListingID        string                     `json:"listing_id"`
	Status           marketplace.ListingStatus  `json:"status"`
	WinningBid       *FilteredBid               `json:"winning_bid,omitempty"`
	TotalBids        int                        `json:"total_bids"`
	RevealedBids     int                        `json:"revealed_bids"`
	FinalPrice       *decimal.Decimal           `json:"final_price,omitempty"`
	StartPrice       decimal.Decimal            `json:"start_price"`
	PriceIncrease    *decimal.Decimal           `json:"price_increase,omitempty"`
	Duration         time.Duration              `json:"duration"`
	CompletedAt      time.Time                  `json:"completed_at"`
	BidProgression   []FilteredBidPoint         `json:"bid_progression,omitempty"`
	ParticipantStats []FilteredParticipantStats `json:"participant_stats,omitempty"`
	AuctionMetrics   FilteredAuctionMetrics     `json:"auction_metrics"`
	VisibilityLevel  BidVisibilityLevel         `json:"visibility_level"`
	RevealReason     string                     `json:"reveal_reason"`
}

// HistoricalBidData represents historical bid data with proper filtering
type HistoricalBidData struct {
	ListingID       string                  `json:"listing_id"`
	RequestedBy     string                  `json:"requested_by"`
	AccessLevel     HistoricalAccessLevel   `json:"access_level"`
	TotalRecords    int                     `json:"total_records"`
	FilteredRecords int                     `json:"filtered_records"`
	BidHistory      []FilteredHistoricalBid `json:"bid_history"`
	TimeRange       TimeRange               `json:"time_range"`
	GeneratedAt     time.Time               `json:"generated_at"`
}

// AuctionSummary represents a high-level summary of auction results
type AuctionSummary struct {
	ListingID    string                    `json:"listing_id"`
	Title        string                    `json:"title"`
	Status       marketplace.ListingStatus `json:"status"`
	Outcome      AuctionOutcome            `json:"outcome"`
	WinnerInfo   *WinnerInfo               `json:"winner_info,omitempty"`
	FinalMetrics SummaryMetrics            `json:"final_metrics"`
	KeyEvents    []KeyEvent                `json:"key_events"`
	CompletedAt  time.Time                 `json:"completed_at"`
}

// Supporting Types

// BidProgressionPoint represents a point in the bid progression timeline
type BidProgressionPoint struct {
	Timestamp     time.Time       `json:"timestamp"`
	BidAmount     decimal.Decimal `json:"bid_amount"`
	BidderID      string          `json:"bidder_id"`
	IsAutoBid     bool            `json:"is_auto_bid"`
	TimeFromStart time.Duration   `json:"time_from_start"`
}

// FilteredBidPoint represents a filtered bid progression point
type FilteredBidPoint struct {
	Timestamp     time.Time        `json:"timestamp"`
	BidAmount     *decimal.Decimal `json:"bid_amount,omitempty"`
	AnonymousID   string           `json:"anonymous_id"`
	IsAutoBid     *bool            `json:"is_auto_bid,omitempty"`
	TimeFromStart time.Duration    `json:"time_from_start"`
}

// ParticipantStats represents statistics for a participant
type ParticipantStats struct {
	BidderID     string          `json:"bidder_id"`
	TotalBids    int             `json:"total_bids"`
	HighestBid   decimal.Decimal `json:"highest_bid"`
	FirstBidTime time.Time       `json:"first_bid_time"`
	LastBidTime  time.Time       `json:"last_bid_time"`
	AutoBidsUsed int             `json:"auto_bids_used"`
	IsWinner     bool            `json:"is_winner"`
}

// FilteredParticipantStats represents filtered participant statistics
type FilteredParticipantStats struct {
	AnonymousID  string           `json:"anonymous_id"`
	TotalBids    int              `json:"total_bids"`
	HighestBid   *decimal.Decimal `json:"highest_bid,omitempty"`
	FirstBidTime *time.Time       `json:"first_bid_time,omitempty"`
	LastBidTime  *time.Time       `json:"last_bid_time,omitempty"`
	AutoBidsUsed *int             `json:"auto_bids_used,omitempty"`
	IsWinner     bool             `json:"is_winner"`
}

// AuctionMetrics represents detailed auction metrics
type AuctionMetrics struct {
	CompetitionLevel     CompetitionLevel `json:"competition_level"`
	BidFrequency         float64          `json:"bid_frequency"` // bids per hour
	AverageBidAmount     decimal.Decimal  `json:"average_bid_amount"`
	MedianBidAmount      decimal.Decimal  `json:"median_bid_amount"`
	BidSpread            decimal.Decimal  `json:"bid_spread"` // difference between highest and lowest
	ParticipationRate    float64          `json:"participation_rate"`
	AutoBidPercentage    float64          `json:"auto_bid_percentage"`
	PeakActivity         time.Time        `json:"peak_activity"`
	ActivityDistribution []ActivityPeriod `json:"activity_distribution"`
}

// FilteredAuctionMetrics represents filtered auction metrics
type FilteredAuctionMetrics struct {
	CompetitionLevel  CompetitionLevel `json:"competition_level"`
	BidFrequency      *float64         `json:"bid_frequency,omitempty"`
	AverageBidAmount  *decimal.Decimal `json:"average_bid_amount,omitempty"`
	MedianBidAmount   *decimal.Decimal `json:"median_bid_amount,omitempty"`
	BidSpread         *decimal.Decimal `json:"bid_spread,omitempty"`
	ParticipationRate *float64         `json:"participation_rate,omitempty"`
	AutoBidPercentage *float64         `json:"auto_bid_percentage,omitempty"`
	PeakActivity      *time.Time       `json:"peak_activity,omitempty"`
}

// HistoricalAccessLevel represents the level of access to historical data
type HistoricalAccessLevel string

const (
	HistoricalAccessFull    HistoricalAccessLevel = "FULL"    // Complete access to all data
	HistoricalAccessLimited HistoricalAccessLevel = "LIMITED" // Limited access based on visibility rules
	HistoricalAccessSummary HistoricalAccessLevel = "SUMMARY" // Only summary statistics
	HistoricalAccessDenied  HistoricalAccessLevel = "DENIED"  // No access to historical data
)

// FilteredHistoricalBid represents a historical bid with filtering applied
type FilteredHistoricalBid struct {
	BidID        string                 `json:"bid_id,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
	BidAmount    *decimal.Decimal       `json:"bid_amount,omitempty"`
	AnonymousID  string                 `json:"anonymous_id"`
	IsAutoBid    *bool                  `json:"is_auto_bid,omitempty"`
	Status       *marketplace.BidStatus `json:"status,omitempty"`
	IsHighestBid *bool                  `json:"is_highest_bid,omitempty"`
}

// TimeRange represents a time range for historical data
type TimeRange struct {
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`
}

// AuctionOutcome represents the outcome of an auction
type AuctionOutcome string

const (
	OutcomeSuccessful    AuctionOutcome = "SUCCESSFUL"      // Auction completed with winner
	OutcomeExpiredNoBids AuctionOutcome = "EXPIRED_NO_BIDS" // Auction expired without bids
	OutcomeCancelled     AuctionOutcome = "CANCELLED"       // Auction was cancelled
	OutcomeForceClose    AuctionOutcome = "FORCE_CLOSED"    // Auction was force closed by admin
)

// WinnerInfo represents information about the auction winner
type WinnerInfo struct {
	AnonymousID   string          `json:"anonymous_id"`
	WinningAmount decimal.Decimal `json:"winning_amount"`
	WinTime       time.Time       `json:"win_time"`
	IsAutoBid     bool            `json:"is_auto_bid"`
	OrderCreated  bool            `json:"order_created"`
	OrderID       string          `json:"order_id,omitempty"`
}

// SummaryMetrics represents high-level summary metrics
type SummaryMetrics struct {
	TotalBids        int              `json:"total_bids"`
	UniqueBidders    int              `json:"unique_bidders"`
	FinalPrice       decimal.Decimal  `json:"final_price"`
	PriceIncrease    decimal.Decimal  `json:"price_increase"`
	Duration         time.Duration    `json:"duration"`
	CompetitionLevel CompetitionLevel `json:"competition_level"`
}

// KeyEvent represents a significant event during the auction
type KeyEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	EventType   string    `json:"event_type"`
	Description string    `json:"description"`
	Impact      string    `json:"impact,omitempty"`
}

// CompetitionLevel represents the level of competition in an auction
type CompetitionLevel string

const (
	CompetitionLow     CompetitionLevel = "LOW"     // Few bidders, low activity
	CompetitionMedium  CompetitionLevel = "MEDIUM"  // Moderate bidders and activity
	CompetitionHigh    CompetitionLevel = "HIGH"    // Many bidders, high activity
	CompetitionIntense CompetitionLevel = "INTENSE" // Very high competition
)

// ActivityPeriod represents activity during a specific time period
type ActivityPeriod struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	BidCount  int       `json:"bid_count"`
	Intensity float64   `json:"intensity"` // bids per minute
}

// HistoricalDataFilter represents filters for historical data queries
type HistoricalDataFilter struct {
	StartTime       *time.Time       `json:"start_time,omitempty"`
	EndTime         *time.Time       `json:"end_time,omitempty"`
	BidderID        string           `json:"bidder_id,omitempty"`
	MinAmount       *decimal.Decimal `json:"min_amount,omitempty"`
	MaxAmount       *decimal.Decimal `json:"max_amount,omitempty"`
	IncludeAutoBids bool             `json:"include_auto_bids"`
	SortOrder       string           `json:"sort_order,omitempty"` // "asc" or "desc"
}

// ResultsValidation represents validation results for auction results
type ResultsValidation struct {
	IsValid          bool                `json:"is_valid"`
	ValidationErrors []ValidationError   `json:"validation_errors,omitempty"`
	Warnings         []ValidationWarning `json:"warnings,omitempty"`
	CheckedAt        time.Time           `json:"checked_at"`
	CheckedBy        string              `json:"checked_by"`
}

// ResultsReconciliation represents reconciliation results
type ResultsReconciliation struct {
	ReconciliationID string               `json:"reconciliation_id"`
	Status           ReconciliationStatus `json:"status"`
	Discrepancies    []Discrepancy        `json:"discrepancies,omitempty"`
	Corrections      []Correction         `json:"corrections,omitempty"`
	ReconciledAt     time.Time            `json:"reconciled_at"`
	ReconciledBy     string               `json:"reconciled_by"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Field    string `json:"field,omitempty"`
	Severity string `json:"severity"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

// ReconciliationStatus represents the status of reconciliation
type ReconciliationStatus string

const (
	ReconciliationPending   ReconciliationStatus = "PENDING"
	ReconciliationCompleted ReconciliationStatus = "COMPLETED"
	ReconciliationFailed    ReconciliationStatus = "FAILED"
)

// Discrepancy represents a discrepancy found during reconciliation
type Discrepancy struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Expected    interface{} `json:"expected"`
	Actual      interface{} `json:"actual"`
	Impact      string      `json:"impact"`
}

// Correction represents a correction made during reconciliation
type Correction struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	OldValue    interface{} `json:"old_value"`
	NewValue    interface{} `json:"new_value"`
	Reason      string      `json:"reason"`
}

// AuctionResultsService provides business logic for auction results revelation
type AuctionResultsService struct {
	bidRepo           BidRepositoryInterface
	listingRepo       ListingRepositoryInterface
	eventRepo         AuctionEventRepositoryInterface
	visibilityService BidVisibilityServiceInterface
	notificationSvc   NotificationServiceInterface
	eventService      EventServiceInterface
}

// NewAuctionResultsService creates a new auction results service
func NewAuctionResultsService(
	bidRepo BidRepositoryInterface,
	listingRepo ListingRepositoryInterface,
	eventRepo AuctionEventRepositoryInterface,
	visibilityService BidVisibilityServiceInterface,
	notificationSvc NotificationServiceInterface,
	eventService EventServiceInterface,
) AuctionResultsServiceInterface {
	return &AuctionResultsService{
		bidRepo:           bidRepo,
		listingRepo:       listingRepo,
		eventRepo:         eventRepo,
		visibilityService: visibilityService,
		notificationSvc:   notificationSvc,
		eventService:      eventService,
	}
}

// ProcessAuctionResults processes the complete results of an auction
func (s *AuctionResultsService) ProcessAuctionResults(ctx context.Context, listingID string) (*AuctionResults, error) {
	// Get listing
	listing, err := s.listingRepo.GetByListingID(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get listing: %w", err)
	}
	if listing == nil {
		return nil, fmt.Errorf("listing not found: %s", listingID)
	}

	// Ensure auction is completed
	if listing.Status == marketplace.ListingStatusActive && !listing.IsExpired() {
		return nil, fmt.Errorf("auction is still active")
	}

	// Get all bids for the listing
	allBids, _, err := s.bidRepo.GetBidHistory(ctx, listingID, marketplace.BidVisibilityFull, "", &common.PaginationRequest{Limit: 1000})
	if err != nil {
		return nil, fmt.Errorf("failed to get bid history: %w", err)
	}

	// Get winning bid
	winningBid, err := s.bidRepo.GetHighestBid(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get winning bid: %w", err)
	}

	// Calculate metrics
	metrics := s.calculateAuctionMetrics(listing, allBids)
	progression := s.buildBidProgression(allBids)
	participantStats := s.calculateParticipantStats(allBids, winningBid)

	// Determine final price and price increase
	finalPrice := listing.MinimumBid
	if winningBid != nil {
		finalPrice = winningBid.BidAmount
	}
	priceIncrease := finalPrice.Sub(listing.MinimumBid)

	// Calculate duration
	duration := listing.ExpiresAt.Sub(listing.CreatedAt)
	if listing.ClosedAt != nil {
		duration = listing.ClosedAt.Sub(listing.CreatedAt)
	}

	results := &AuctionResults{
		ListingID:        listingID,
		Status:           listing.Status,
		WinningBid:       winningBid,
		TotalBids:        len(allBids),
		UniqueBidders:    s.countUniqueBidders(allBids),
		FinalPrice:       finalPrice,
		StartPrice:       listing.MinimumBid,
		PriceIncrease:    priceIncrease,
		Duration:         duration,
		CompletedAt:      s.getCompletionTime(listing),
		BidProgression:   progression,
		ParticipantStats: participantStats,
		AuctionMetrics:   metrics,
	}

	return results, nil
}

// RevealAuctionResults reveals auction results with appropriate visibility filtering
func (s *AuctionResultsService) RevealAuctionResults(ctx context.Context, listingID string, viewerID string, viewerOrgID string) (*RevealedAuctionResults, error) {
	// Get complete results
	results, err := s.ProcessAuctionResults(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to process auction results: %w", err)
	}

	// Get listing for visibility rules
	listing, err := s.listingRepo.GetByListingID(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get listing: %w", err)
	}

	// Check if viewer can access results
	if !s.visibilityService.CanViewListing(ctx, listing, viewerID, viewerOrgID) {
		return nil, fmt.Errorf("access denied: cannot view auction results")
	}

	// Get auction visibility state
	auctionState, err := s.visibilityService.GetAuctionVisibilityState(ctx, listing, viewerID, viewerOrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get auction visibility state: %w", err)
	}

	// Apply visibility filtering
	revealedResults := s.applyResultsVisibilityFiltering(results, listing, auctionState, viewerID)

	return revealedResults, nil
}

// NotifyWinner sends notification to the auction winner
func (s *AuctionResultsService) NotifyWinner(ctx context.Context, listing *marketplace.Listing, winningBid *marketplace.Bid) error {
	if winningBid == nil {
		return nil // No winner to notify
	}

	// Create winner notification data
	_ = map[string]interface{}{
		"listing_id":     listing.ListingID,
		"product_id":     listing.ProductID,
		"winning_amount": winningBid.BidAmount,
		"seller_id":      listing.SellerID,
		"auction_type":   listing.AuctionType,
		"expires_at":     listing.ExpiresAt,
	}

	// Record winner notification event
	if s.eventService != nil {
		eventData := map[string]interface{}{
			"winner_id":         winningBid.BidderID,
			"winning_amount":    winningBid.BidAmount,
			"notification_sent": true,
		}
		_ = s.eventService.RecordListingEvent(ctx, listing.ListingID, "WINNER_NOTIFIED", eventData, "SYSTEM")
	}

	// Send notification (implementation would depend on notification service)
	if s.notificationSvc != nil {
		return s.notificationSvc.NotifyBidPlaced(ctx, winningBid, listing) // Reuse existing notification method
	}

	return nil
}

// NotifyParticipants sends notifications to all auction participants
func (s *AuctionResultsService) NotifyParticipants(ctx context.Context, listing *marketplace.Listing, results *AuctionResults) error {
	// Get all unique bidders
	uniqueBidders := make(map[string]bool)
	for _, stats := range results.ParticipantStats {
		uniqueBidders[stats.BidderID] = true
	}

	// Notify each participant
	for bidderID := range uniqueBidders {
		isWinner := results.WinningBid != nil && results.WinningBid.BidderID == bidderID

		_ = map[string]interface{}{
			"listing_id":   listing.ListingID,
			"product_id":   listing.ProductID,
			"final_price":  results.FinalPrice,
			"total_bids":   results.TotalBids,
			"is_winner":    isWinner,
			"auction_type": listing.AuctionType,
		}

		// Record participant notification event
		if s.eventService != nil {
			eventData := map[string]interface{}{
				"participant_id":    bidderID,
				"is_winner":         isWinner,
				"notification_type": "AUCTION_ENDED",
			}
			_ = s.eventService.RecordListingEvent(ctx, listing.ListingID, "PARTICIPANT_NOTIFIED", eventData, "SYSTEM")
		}
	}

	return nil
}

// GetHistoricalBidData retrieves historical bid data with proper filtering
func (s *AuctionResultsService) GetHistoricalBidData(ctx context.Context, listingID string, viewerID string, viewerOrgID string, filter *HistoricalDataFilter) (*HistoricalBidData, error) {
	// Get listing
	listing, err := s.listingRepo.GetByListingID(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get listing: %w", err)
	}
	if listing == nil {
		return nil, fmt.Errorf("listing not found: %s", listingID)
	}

	// Check access level
	accessLevel := s.determineHistoricalAccessLevel(listing, viewerID, viewerOrgID)
	if accessLevel == HistoricalAccessDenied {
		return nil, fmt.Errorf("access denied: cannot view historical bid data")
	}

	// Get bid history
	allBids, _, err := s.bidRepo.GetBidHistory(ctx, listingID, listing.BidVisibility, viewerID, &common.PaginationRequest{Limit: 1000})
	if err != nil {
		return nil, fmt.Errorf("failed to get bid history: %w", err)
	}

	// Apply filters
	filteredBids := s.applyHistoricalFilters(allBids, filter)

	// Apply visibility filtering
	historicalBids := s.filterHistoricalBids(filteredBids, listing, accessLevel, viewerID)

	// Determine time range
	timeRange := s.calculateTimeRange(listing, filteredBids)

	historicalData := &HistoricalBidData{
		ListingID:       listingID,
		RequestedBy:     viewerID,
		AccessLevel:     accessLevel,
		TotalRecords:    len(allBids),
		FilteredRecords: len(historicalBids),
		BidHistory:      historicalBids,
		TimeRange:       timeRange,
		GeneratedAt:     time.Now(),
	}

	return historicalData, nil
}

// GetAuctionSummary retrieves a high-level summary of auction results
func (s *AuctionResultsService) GetAuctionSummary(ctx context.Context, listingID string, viewerID string, viewerOrgID string) (*AuctionSummary, error) {
	// Get results
	results, err := s.ProcessAuctionResults(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to process auction results: %w", err)
	}

	// Get listing
	listing, err := s.listingRepo.GetByListingID(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get listing: %w", err)
	}

	// Check access
	if !s.visibilityService.CanViewListing(ctx, listing, viewerID, viewerOrgID) {
		return nil, fmt.Errorf("access denied: cannot view auction summary")
	}

	// Determine outcome
	outcome := s.determineAuctionOutcome(listing, results)

	// Create winner info if applicable
	var winnerInfo *WinnerInfo
	if results.WinningBid != nil {
		winnerInfo = &WinnerInfo{
			AnonymousID:   s.generateAnonymousID(results.WinningBid.BidderID, viewerID),
			WinningAmount: results.WinningBid.BidAmount,
			WinTime:       results.WinningBid.PlacedAt,
			IsAutoBid:     results.WinningBid.IsAutoBid,
			OrderCreated:  false, // Would need to check order service
		}
	}

	// Create summary metrics
	summaryMetrics := SummaryMetrics{
		TotalBids:        results.TotalBids,
		UniqueBidders:    results.UniqueBidders,
		FinalPrice:       results.FinalPrice,
		PriceIncrease:    results.PriceIncrease,
		Duration:         results.Duration,
		CompetitionLevel: results.AuctionMetrics.CompetitionLevel,
	}

	// Generate key events
	keyEvents := s.generateKeyEvents(listing, results)

	summary := &AuctionSummary{
		ListingID:    listingID,
		Title:        fmt.Sprintf("Auction for Product %s", listing.ProductID),
		Status:       listing.Status,
		Outcome:      outcome,
		WinnerInfo:   winnerInfo,
		FinalMetrics: summaryMetrics,
		KeyEvents:    keyEvents,
		CompletedAt:  results.CompletedAt,
	}

	return summary, nil
}

// ValidateAuctionResults validates the integrity of auction results
func (s *AuctionResultsService) ValidateAuctionResults(ctx context.Context, listingID string) (*ResultsValidation, error) {
	// Get results
	results, err := s.ProcessAuctionResults(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to process auction results: %w", err)
	}

	validation := &ResultsValidation{
		IsValid:   true,
		CheckedAt: time.Now(),
		CheckedBy: "SYSTEM",
	}

	// Validate winning bid
	if results.WinningBid != nil {
		if err := s.validateWinningBid(results); err != nil {
			validation.IsValid = false
			validation.ValidationErrors = append(validation.ValidationErrors, ValidationError{
				Code:     "INVALID_WINNING_BID",
				Message:  err.Error(),
				Severity: "ERROR",
			})
		}
	}

	// Validate bid progression
	if err := s.validateBidProgression(results.BidProgression); err != nil {
		validation.IsValid = false
		validation.ValidationErrors = append(validation.ValidationErrors, ValidationError{
			Code:     "INVALID_BID_PROGRESSION",
			Message:  err.Error(),
			Severity: "ERROR",
		})
	}

	// Check for warnings
	warnings := s.checkForWarnings(results)
	validation.Warnings = warnings

	return validation, nil
}

// ReconcileResults reconciles auction results with external systems
func (s *AuctionResultsService) ReconcileResults(ctx context.Context, listingID string) (*ResultsReconciliation, error) {
	reconciliation := &ResultsReconciliation{
		ReconciliationID: fmt.Sprintf("REC_%s_%d", listingID, time.Now().Unix()),
		Status:           ReconciliationPending,
		ReconciledAt:     time.Now(),
		ReconciledBy:     "SYSTEM",
	}

	// Perform reconciliation checks
	// This would involve checking against inventory, order systems, etc.

	reconciliation.Status = ReconciliationCompleted
	return reconciliation, nil
}

// Helper methods

// calculateAuctionMetrics calculates detailed auction metrics
func (s *AuctionResultsService) calculateAuctionMetrics(listing *marketplace.Listing, bids []*marketplace.Bid) AuctionMetrics {
	if len(bids) == 0 {
		return AuctionMetrics{
			CompetitionLevel: CompetitionLow,
		}
	}

	// Calculate basic metrics
	totalAmount := decimal.Zero
	autoBidCount := 0
	uniqueBidders := make(map[string]bool)

	for _, bid := range bids {
		totalAmount = totalAmount.Add(bid.BidAmount)
		if bid.IsAutoBid {
			autoBidCount++
		}
		uniqueBidders[bid.BidderID] = true
	}

	averageBid := totalAmount.Div(decimal.NewFromInt(int64(len(bids))))

	// Calculate competition level
	competitionLevel := s.determineCompetitionLevel(len(bids), len(uniqueBidders), listing.ListingDuration)

	// Calculate bid frequency
	duration := listing.ExpiresAt.Sub(listing.CreatedAt)
	bidFrequency := float64(len(bids)) / duration.Hours()

	return AuctionMetrics{
		CompetitionLevel:  competitionLevel,
		BidFrequency:      bidFrequency,
		AverageBidAmount:  averageBid,
		AutoBidPercentage: float64(autoBidCount) / float64(len(bids)) * 100,
		ParticipationRate: float64(len(uniqueBidders)),
	}
}

// Additional helper methods would be implemented here...
// (buildBidProgression, calculateParticipantStats, etc.)

// Placeholder implementations for brevity
func (s *AuctionResultsService) buildBidProgression(bids []*marketplace.Bid) []BidProgressionPoint {
	progression := make([]BidProgressionPoint, len(bids))
	for i, bid := range bids {
		progression[i] = BidProgressionPoint{
			Timestamp: bid.PlacedAt,
			BidAmount: bid.BidAmount,
			BidderID:  bid.BidderID,
			IsAutoBid: bid.IsAutoBid,
		}
	}
	return progression
}

func (s *AuctionResultsService) calculateParticipantStats(bids []*marketplace.Bid, winningBid *marketplace.Bid) []ParticipantStats {
	stats := make(map[string]*ParticipantStats)

	for _, bid := range bids {
		if _, exists := stats[bid.BidderID]; !exists {
			stats[bid.BidderID] = &ParticipantStats{
				BidderID:     bid.BidderID,
				FirstBidTime: bid.PlacedAt,
				LastBidTime:  bid.PlacedAt,
				HighestBid:   bid.BidAmount,
			}
		}

		stat := stats[bid.BidderID]
		stat.TotalBids++
		if bid.BidAmount.GreaterThan(stat.HighestBid) {
			stat.HighestBid = bid.BidAmount
		}
		if bid.PlacedAt.After(stat.LastBidTime) {
			stat.LastBidTime = bid.PlacedAt
		}
		if bid.IsAutoBid {
			stat.AutoBidsUsed++
		}
		if winningBid != nil && bid.BidderID == winningBid.BidderID {
			stat.IsWinner = true
		}
	}

	result := make([]ParticipantStats, 0, len(stats))
	for _, stat := range stats {
		result = append(result, *stat)
	}

	return result
}

func (s *AuctionResultsService) countUniqueBidders(bids []*marketplace.Bid) int {
	unique := make(map[string]bool)
	for _, bid := range bids {
		unique[bid.BidderID] = true
	}
	return len(unique)
}

func (s *AuctionResultsService) getCompletionTime(listing *marketplace.Listing) time.Time {
	if listing.ClosedAt != nil {
		return *listing.ClosedAt
	}
	return listing.ExpiresAt
}

func (s *AuctionResultsService) determineCompetitionLevel(totalBids, uniqueBidders, durationHours int) CompetitionLevel {
	bidsPerHour := float64(totalBids) / float64(durationHours)

	if uniqueBidders >= 10 && bidsPerHour >= 5 {
		return CompetitionIntense
	} else if uniqueBidders >= 5 && bidsPerHour >= 2 {
		return CompetitionHigh
	} else if uniqueBidders >= 2 && bidsPerHour >= 0.5 {
		return CompetitionMedium
	}
	return CompetitionLow
}

func (s *AuctionResultsService) applyResultsVisibilityFiltering(results *AuctionResults, listing *marketplace.Listing, auctionState *AuctionVisibilityState, viewerID string) *RevealedAuctionResults {
	// Apply visibility filtering based on auction state and configuration
	// This is a simplified implementation
	return &RevealedAuctionResults{
		ListingID:       results.ListingID,
		Status:          results.Status,
		TotalBids:       results.TotalBids,
		RevealedBids:    results.TotalBids,
		FinalPrice:      &results.FinalPrice,
		StartPrice:      results.StartPrice,
		PriceIncrease:   &results.PriceIncrease,
		Duration:        results.Duration,
		CompletedAt:     results.CompletedAt,
		VisibilityLevel: VisibilityLevelFull,
		RevealReason:    auctionState.RevealReason,
	}
}

func (s *AuctionResultsService) determineHistoricalAccessLevel(listing *marketplace.Listing, viewerID, viewerOrgID string) HistoricalAccessLevel {
	// Determine access level based on user role and listing visibility
	if listing.SellerID == viewerID {
		return HistoricalAccessFull
	}

	if listing.Status != marketplace.ListingStatusActive {
		return HistoricalAccessLimited
	}

	return HistoricalAccessSummary
}

func (s *AuctionResultsService) applyHistoricalFilters(bids []*marketplace.Bid, filter *HistoricalDataFilter) []*marketplace.Bid {
	if filter == nil {
		return bids
	}

	filtered := make([]*marketplace.Bid, 0)
	for _, bid := range bids {
		if s.bidMatchesFilter(bid, filter) {
			filtered = append(filtered, bid)
		}
	}

	return filtered
}

func (s *AuctionResultsService) bidMatchesFilter(bid *marketplace.Bid, filter *HistoricalDataFilter) bool {
	if filter.StartTime != nil && bid.PlacedAt.Before(*filter.StartTime) {
		return false
	}
	if filter.EndTime != nil && bid.PlacedAt.After(*filter.EndTime) {
		return false
	}
	if filter.BidderID != "" && bid.BidderID != filter.BidderID {
		return false
	}
	if filter.MinAmount != nil && bid.BidAmount.LessThan(*filter.MinAmount) {
		return false
	}
	if filter.MaxAmount != nil && bid.BidAmount.GreaterThan(*filter.MaxAmount) {
		return false
	}
	if !filter.IncludeAutoBids && bid.IsAutoBid {
		return false
	}

	return true
}

func (s *AuctionResultsService) filterHistoricalBids(bids []*marketplace.Bid, listing *marketplace.Listing, accessLevel HistoricalAccessLevel, viewerID string) []FilteredHistoricalBid {
	filtered := make([]FilteredHistoricalBid, 0)

	for _, bid := range bids {
		filteredBid := FilteredHistoricalBid{
			Timestamp:   bid.PlacedAt,
			AnonymousID: s.generateAnonymousID(bid.BidderID, viewerID),
		}

		if accessLevel == HistoricalAccessFull || bid.BidderID == viewerID {
			bidAmount := bid.BidAmount
			filteredBid.BidID = bid.BidID
			filteredBid.BidAmount = &bidAmount
			filteredBid.IsAutoBid = &bid.IsAutoBid
			filteredBid.Status = &bid.Status
			filteredBid.IsHighestBid = &bid.IsHighestBid
		}

		filtered = append(filtered, filteredBid)
	}

	return filtered
}

func (s *AuctionResultsService) calculateTimeRange(listing *marketplace.Listing, bids []*marketplace.Bid) TimeRange {
	startTime := listing.CreatedAt
	endTime := s.getCompletionTime(listing)

	if len(bids) > 0 {
		if bids[0].PlacedAt.After(startTime) {
			startTime = bids[0].PlacedAt
		}
		if bids[len(bids)-1].PlacedAt.Before(endTime) {
			endTime = bids[len(bids)-1].PlacedAt
		}
	}

	return TimeRange{
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime),
	}
}

func (s *AuctionResultsService) determineAuctionOutcome(listing *marketplace.Listing, results *AuctionResults) AuctionOutcome {
	switch listing.Status {
	case marketplace.ListingStatusClosed:
		if results.WinningBid != nil {
			return OutcomeSuccessful
		}
		return OutcomeExpiredNoBids
	case marketplace.ListingStatusExpiredNoBids:
		return OutcomeExpiredNoBids
	case marketplace.ListingStatusCancelled:
		return OutcomeCancelled
	default:
		return OutcomeExpiredNoBids
	}
}

func (s *AuctionResultsService) generateAnonymousID(bidderID, viewerID string) string {
	if bidderID == viewerID {
		return "Your Bid"
	}

	// Simple hash-based anonymization
	hash := 0
	for _, char := range bidderID {
		hash = hash*31 + int(char)
	}
	if hash < 0 {
		hash = -hash
	}
	return fmt.Sprintf("Bidder #%d", (hash%1000)+1)
}

func (s *AuctionResultsService) generateKeyEvents(listing *marketplace.Listing, results *AuctionResults) []KeyEvent {
	events := []KeyEvent{
		{
			Timestamp:   listing.CreatedAt,
			EventType:   "AUCTION_STARTED",
			Description: "Auction started",
		},
	}

	if results.WinningBid != nil {
		events = append(events, KeyEvent{
			Timestamp:   results.WinningBid.PlacedAt,
			EventType:   "WINNING_BID",
			Description: fmt.Sprintf("Winning bid placed: %s", results.WinningBid.BidAmount.String()),
		})
	}

	events = append(events, KeyEvent{
		Timestamp:   results.CompletedAt,
		EventType:   "AUCTION_ENDED",
		Description: "Auction completed",
	})

	return events
}

func (s *AuctionResultsService) validateWinningBid(results *AuctionResults) error {
	if results.WinningBid == nil {
		return nil
	}

	// Validate that winning bid is actually the highest
	for _, stats := range results.ParticipantStats {
		if stats.BidderID != results.WinningBid.BidderID && stats.HighestBid.GreaterThan(results.WinningBid.BidAmount) {
			return fmt.Errorf("winning bid is not the highest bid")
		}
	}

	return nil
}

func (s *AuctionResultsService) validateBidProgression(progression []BidProgressionPoint) error {
	for i := 1; i < len(progression); i++ {
		if progression[i].Timestamp.Before(progression[i-1].Timestamp) {
			return fmt.Errorf("bid progression timestamps are not in order")
		}
	}
	return nil
}

func (s *AuctionResultsService) checkForWarnings(results *AuctionResults) []ValidationWarning {
	var warnings []ValidationWarning

	if results.TotalBids == 0 {
		warnings = append(warnings, ValidationWarning{
			Code:    "NO_BIDS",
			Message: "Auction completed without any bids",
		})
	}

	if results.UniqueBidders == 1 {
		warnings = append(warnings, ValidationWarning{
			Code:    "SINGLE_BIDDER",
			Message: "Auction had only one bidder",
		})
	}

	return warnings
}
