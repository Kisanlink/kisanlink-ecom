package marketplace

import (
	"context"
	"fmt"
	"time"

	marketplaceModels "github.com/Kisanlink/kisanlink-ecom/entities/models/marketplace"
	marketplaceRepo "github.com/Kisanlink/kisanlink-ecom/internal/repositories/marketplace"

	"github.com/shopspring/decimal"
)

// AnalyticsServiceInterface defines the interface for marketplace analytics operations
type AnalyticsServiceInterface interface {
	// Marketplace Statistics
	GetMarketplaceStats(ctx context.Context, timeRange string) (*MarketplaceStats, error)
	GetListingStats(ctx context.Context, listingID string) (*ListingStats, error)
	GetUserStats(ctx context.Context, userID string, timeRange string) (*UserStats, error)
	GetOrganizationStats(ctx context.Context, orgID string, timeRange string) (*OrganizationStats, error)

	// Performance Metrics
	GetPerformanceMetrics(ctx context.Context, timeRange string) (*PerformanceMetrics, error)
	GetBiddingTrends(ctx context.Context, timeRange string) (*BiddingTrends, error)
	GetPopularCategories(ctx context.Context, timeRange string) ([]*CategoryStats, error)

	// Revenue Analytics
	GetRevenueStats(ctx context.Context, timeRange string) (*RevenueStats, error)
	GetCommissionStats(ctx context.Context, timeRange string) (*CommissionStats, error)
}

// MarketplaceStats represents overall marketplace statistics
type MarketplaceStats struct {
	TotalListings         int             `json:"total_listings"`
	ActiveListings        int             `json:"active_listings"`
	ClosedListings        int             `json:"closed_listings"`
	ExpiredListings       int             `json:"expired_listings"`
	TotalBids             int             `json:"total_bids"`
	AverageBidsPerListing float64         `json:"average_bids_per_listing"`
	TotalValue            decimal.Decimal `json:"total_value"`
	AverageListingValue   decimal.Decimal `json:"average_listing_value"`
	SuccessRate           float64         `json:"success_rate"`
	TimeRange             string          `json:"time_range"`
	GeneratedAt           time.Time       `json:"generated_at"`
}

// ListingStats represents statistics for a specific listing
type ListingStats struct {
	ListingID         string          `json:"listing_id"`
	TotalBids         int             `json:"total_bids"`
	UniqueBidders     int             `json:"unique_bidders"`
	HighestBid        decimal.Decimal `json:"highest_bid"`
	AverageBid        decimal.Decimal `json:"average_bid"`
	BidIncrement      decimal.Decimal `json:"bid_increment"`
	TimeToFirstBid    *time.Duration  `json:"time_to_first_bid,omitempty"`
	BiddingVelocity   float64         `json:"bidding_velocity"` // bids per hour
	AutoBidPercentage float64         `json:"auto_bid_percentage"`
	ViewCount         int             `json:"view_count"`
	ConversionRate    float64         `json:"conversion_rate"`
}

// UserStats represents statistics for a specific user
type UserStats struct {
	UserID              string          `json:"user_id"`
	TotalListings       int             `json:"total_listings"`
	ActiveListings      int             `json:"active_listings"`
	SuccessfulListings  int             `json:"successful_listings"`
	TotalBids           int             `json:"total_bids"`
	WinningBids         int             `json:"winning_bids"`
	AverageBidAmount    decimal.Decimal `json:"average_bid_amount"`
	TotalSpent          decimal.Decimal `json:"total_spent"`
	TotalEarned         decimal.Decimal `json:"total_earned"`
	SuccessRateAsSeller float64         `json:"success_rate_as_seller"`
	SuccessRateAsBuyer  float64         `json:"success_rate_as_buyer"`
	TimeRange           string          `json:"time_range"`
}

// OrganizationStats represents statistics for an organization
type OrganizationStats struct {
	OrganizationID      string          `json:"organization_id"`
	TotalUsers          int             `json:"total_users"`
	TotalListings       int             `json:"total_listings"`
	TotalBids           int             `json:"total_bids"`
	TotalVolume         decimal.Decimal `json:"total_volume"`
	AverageListingValue decimal.Decimal `json:"average_listing_value"`
	SuccessRate         float64         `json:"success_rate"`
	TopCategories       []string        `json:"top_categories"`
	TimeRange           string          `json:"time_range"`
}

// PerformanceMetrics represents system performance metrics
type PerformanceMetrics struct {
	AverageResponseTime time.Duration `json:"average_response_time"`
	BidProcessingTime   time.Duration `json:"bid_processing_time"`
	ListingCreationTime time.Duration `json:"listing_creation_time"`
	ConcurrentUsers     int           `json:"concurrent_users"`
	PeakBiddingPeriods  []TimeSlot    `json:"peak_bidding_periods"`
	SystemUptime        float64       `json:"system_uptime"`
	ErrorRate           float64       `json:"error_rate"`
	TimeRange           string        `json:"time_range"`
}

// BiddingTrends represents bidding trend analysis
type BiddingTrends struct {
	TotalBids         int                    `json:"total_bids"`
	BidsPerDay        []DailyBidCount        `json:"bids_per_day"`
	BidsPerHour       []HourlyBidCount       `json:"bids_per_hour"`
	AverageBidAmount  decimal.Decimal        `json:"average_bid_amount"`
	BidAmountTrend    []BidAmountTrend       `json:"bid_amount_trend"`
	AutoBidPercentage float64                `json:"auto_bid_percentage"`
	PeakBiddingHours  []int                  `json:"peak_bidding_hours"`
	SeasonalTrends    map[string]interface{} `json:"seasonal_trends"`
	TimeRange         string                 `json:"time_range"`
}

// DailyBidCount represents bid count for a specific day
type DailyBidCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// HourlyBidCount represents bid count for a specific hour
type HourlyBidCount struct {
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

// CategoryStats represents statistics for product categories
type CategoryStats struct {
	Category       string          `json:"category"`
	TotalListings  int             `json:"total_listings"`
	TotalBids      int             `json:"total_bids"`
	AverageValue   decimal.Decimal `json:"average_value"`
	SuccessRate    float64         `json:"success_rate"`
	GrowthRate     float64         `json:"growth_rate"`
	PopularityRank int             `json:"popularity_rank"`
}

// RevenueStats represents revenue and financial statistics
type RevenueStats struct {
	TotalRevenue         decimal.Decimal   `json:"total_revenue"`
	CommissionRevenue    decimal.Decimal   `json:"commission_revenue"`
	ListingFees          decimal.Decimal   `json:"listing_fees"`
	TransactionFees      decimal.Decimal   `json:"transaction_fees"`
	AverageOrderValue    decimal.Decimal   `json:"average_order_value"`
	RevenueGrowthRate    float64           `json:"revenue_growth_rate"`
	TopRevenueCategories []CategoryRevenue `json:"top_revenue_categories"`
	TimeRange            string            `json:"time_range"`
}

// CommissionStats represents commission-specific statistics
type CommissionStats struct {
	TotalCommissions   decimal.Decimal   `json:"total_commissions"`
	AverageCommission  decimal.Decimal   `json:"average_commission"`
	CommissionRate     float64           `json:"commission_rate"`
	TopCommissionUsers []UserCommission  `json:"top_commission_users"`
	CommissionTrends   []CommissionTrend `json:"commission_trends"`
	TimeRange          string            `json:"time_range"`
}

// Supporting types
type TimeSlot struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Activity  int       `json:"activity"`
}

// Duplicate declarations removed - using the ones defined earlier

type BidAmountTrend struct {
	Date   time.Time       `json:"date"`
	Amount decimal.Decimal `json:"amount"`
}

type CategoryRevenue struct {
	Category string          `json:"category"`
	Revenue  decimal.Decimal `json:"revenue"`
}

type UserCommission struct {
	UserID     string          `json:"user_id"`
	Commission decimal.Decimal `json:"commission"`
}

type CommissionTrend struct {
	Date       time.Time       `json:"date"`
	Commission decimal.Decimal `json:"commission"`
}

// AnalyticsService provides marketplace analytics and reporting
type AnalyticsService struct {
	listingRepo marketplaceRepo.ListingRepository
	bidRepo     marketplaceRepo.BidRepository
	eventRepo   marketplaceRepo.AuctionEventRepository
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(
	listingRepo marketplaceRepo.ListingRepository,
	bidRepo marketplaceRepo.BidRepository,
	eventRepo marketplaceRepo.AuctionEventRepository,
) AnalyticsServiceInterface {
	return &AnalyticsService{
		listingRepo: listingRepo,
		bidRepo:     bidRepo,
		eventRepo:   eventRepo,
	}
}

// GetMarketplaceStats retrieves comprehensive marketplace statistics
func (s *AnalyticsService) GetMarketplaceStats(ctx context.Context, timeRange string) (*MarketplaceStats, error) {
	// Parse time range
	startTime, endTime, err := s.parseTimeRange(timeRange)
	if err != nil {
		return nil, fmt.Errorf("invalid time range: %w", err)
	}

	// Get listing statistics
	totalListings, err := s.listingRepo.CountListings(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to count total listings: %w", err)
	}

	activeListings, err := s.listingRepo.CountListingsByStatus(ctx, marketplaceModels.ListingStatusActive, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to count active listings: %w", err)
	}

	closedListings, err := s.listingRepo.CountListingsByStatus(ctx, marketplaceModels.ListingStatusClosed, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to count closed listings: %w", err)
	}

	expiredListings, err := s.listingRepo.CountListingsByStatus(ctx, marketplaceModels.ListingStatusExpired, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to count expired listings: %w", err)
	}

	// Get bid statistics
	totalBids, err := s.bidRepo.CountBids(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to count total bids: %w", err)
	}

	// Calculate averages
	var averageBidsPerListing float64
	if totalListings > 0 {
		averageBidsPerListing = float64(totalBids) / float64(totalListings)
	}

	// Get value statistics
	totalValue, err := s.listingRepo.GetTotalListingValue(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get total listing value: %w", err)
	}

	var averageListingValue decimal.Decimal
	if totalListings > 0 {
		averageListingValue = decimal.NewFromFloat(totalValue).Div(decimal.NewFromInt(int64(totalListings)))
	}

	// Calculate success rate
	var successRate float64
	if totalListings > 0 {
		successRate = float64(closedListings) / float64(totalListings) * 100
	}

	return &MarketplaceStats{
		TotalListings:         totalListings,
		ActiveListings:        activeListings,
		ClosedListings:        closedListings,
		ExpiredListings:       expiredListings,
		TotalBids:             totalBids,
		AverageBidsPerListing: averageBidsPerListing,
		TotalValue:            decimal.NewFromFloat(totalValue),
		AverageListingValue:   averageListingValue,
		SuccessRate:           successRate,
		TimeRange:             timeRange,
		GeneratedAt:           time.Now(),
	}, nil
}

// GetListingStats retrieves statistics for a specific listing
func (s *AnalyticsService) GetListingStats(ctx context.Context, listingID string) (*ListingStats, error) {
	// Get listing
	listing, err := s.listingRepo.GetByListingID(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get listing: %w", err)
	}
	if listing == nil {
		return nil, fmt.Errorf("listing not found: %s", listingID)
	}

	// Get bid statistics
	bidStats, err := s.bidRepo.GetBidStatistics(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bid statistics: %w", err)
	}

	// Calculate time to first bid
	var timeToFirstBid *time.Duration
	if bidStats.LastBidTime != nil {
		duration := bidStats.LastBidTime.Sub(listing.CreatedAt)
		timeToFirstBid = &duration
	}

	// Calculate bidding velocity (bids per hour)
	var biddingVelocity float64
	if !listing.CreatedAt.IsZero() {
		hoursElapsed := time.Since(listing.CreatedAt).Hours()
		if hoursElapsed > 0 {
			biddingVelocity = float64(bidStats.TotalBids) / hoursElapsed
		}
	}

	// Calculate auto-bid percentage
	var autoBidPercentage float64
	if bidStats.TotalBids > 0 {
		autoBidPercentage = float64(bidStats.AutoBidCount) / float64(bidStats.TotalBids) * 100
	}

	// Calculate conversion rate (placeholder - would need view tracking)
	conversionRate := 0.0 // This would require view tracking implementation

	return &ListingStats{
		ListingID:         listingID,
		TotalBids:         bidStats.TotalBids,
		UniqueBidders:     bidStats.UniqueBidders,
		HighestBid:        bidStats.HighestBid,
		AverageBid:        bidStats.AverageBid,
		BidIncrement:      bidStats.BidIncrement,
		TimeToFirstBid:    timeToFirstBid,
		BiddingVelocity:   biddingVelocity,
		AutoBidPercentage: autoBidPercentage,
		ViewCount:         0, // Placeholder - would need view tracking
		ConversionRate:    conversionRate,
	}, nil
}

// GetUserStats retrieves statistics for a specific user
func (s *AnalyticsService) GetUserStats(ctx context.Context, userID string, timeRange string) (*UserStats, error) {
	// Parse time range
	startTime, endTime, err := s.parseTimeRange(timeRange)
	if err != nil {
		return nil, fmt.Errorf("invalid time range: %w", err)
	}

	// Get user listing statistics
	totalListings, err := s.listingRepo.CountUserListings(ctx, userID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to count user listings: %w", err)
	}

	activeListings, err := s.listingRepo.CountUserListingsByStatus(ctx, userID, marketplaceModels.ListingStatusActive, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to count active user listings: %w", err)
	}

	successfulListings, err := s.listingRepo.CountUserListingsByStatus(ctx, userID, marketplaceModels.ListingStatusClosed, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to count successful user listings: %w", err)
	}

	// Get user bid statistics
	totalBids, err := s.bidRepo.CountUserBids(ctx, userID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to count user bids: %w", err)
	}

	winningBids, err := s.bidRepo.CountUserWinningBids(ctx, userID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to count user winning bids: %w", err)
	}

	// Get financial statistics
	averageBidAmount, err := s.bidRepo.GetUserAverageBidAmount(ctx, userID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get user average bid amount: %w", err)
	}

	totalSpent, err := s.bidRepo.GetUserTotalSpent(ctx, userID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get user total spent: %w", err)
	}

	totalEarned, err := s.listingRepo.GetUserTotalEarned(ctx, userID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get user total earned: %w", err)
	}

	// Calculate success rates
	var successRateAsSeller, successRateAsBuyer float64
	if totalListings > 0 {
		successRateAsSeller = float64(successfulListings) / float64(totalListings) * 100
	}
	if totalBids > 0 {
		successRateAsBuyer = float64(winningBids) / float64(totalBids) * 100
	}

	return &UserStats{
		UserID:              userID,
		TotalListings:       totalListings,
		ActiveListings:      activeListings,
		SuccessfulListings:  successfulListings,
		TotalBids:           totalBids,
		WinningBids:         winningBids,
		AverageBidAmount:    averageBidAmount,
		TotalSpent:          totalSpent,
		TotalEarned:         totalEarned,
		SuccessRateAsSeller: successRateAsSeller,
		SuccessRateAsBuyer:  successRateAsBuyer,
		TimeRange:           timeRange,
	}, nil
}

// GetOrganizationStats retrieves statistics for an organization
func (s *AnalyticsService) GetOrganizationStats(ctx context.Context, orgID string, timeRange string) (*OrganizationStats, error) {
	// Parse time range
	startTime, endTime, err := s.parseTimeRange(timeRange)
	if err != nil {
		return nil, fmt.Errorf("invalid time range: %w", err)
	}

	// Get organization statistics
	totalUsers, err := s.listingRepo.CountOrganizationUsers(ctx, orgID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to count organization users: %w", err)
	}

	totalListings, err := s.listingRepo.CountOrganizationListings(ctx, orgID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to count organization listings: %w", err)
	}

	totalBids, err := s.bidRepo.CountOrganizationBids(ctx, orgID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to count organization bids: %w", err)
	}

	totalVolume, err := s.listingRepo.GetOrganizationTotalVolume(ctx, orgID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization total volume: %w", err)
	}

	var averageListingValue decimal.Decimal
	if totalListings > 0 {
		averageListingValue = totalVolume.Div(decimal.NewFromInt(int64(totalListings)))
	}

	successfulListings, err := s.listingRepo.CountOrganizationListingsByStatus(ctx, orgID, marketplaceModels.ListingStatusClosed, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to count successful organization listings: %w", err)
	}

	var successRate float64
	if totalListings > 0 {
		successRate = float64(successfulListings) / float64(totalListings) * 100
	}

	// Get top categories (placeholder)
	topCategories := []string{"Products", "Services", "Labour"} // This would be calculated from actual data

	return &OrganizationStats{
		OrganizationID:      orgID,
		TotalUsers:          totalUsers,
		TotalListings:       totalListings,
		TotalBids:           totalBids,
		TotalVolume:         totalVolume,
		AverageListingValue: averageListingValue,
		SuccessRate:         successRate,
		TopCategories:       topCategories,
		TimeRange:           timeRange,
	}, nil
}

// GetPerformanceMetrics retrieves system performance metrics
func (s *AnalyticsService) GetPerformanceMetrics(ctx context.Context, timeRange string) (*PerformanceMetrics, error) {
	// This would integrate with monitoring systems
	// For now, return placeholder data
	return &PerformanceMetrics{
		AverageResponseTime: 150 * time.Millisecond,
		BidProcessingTime:   50 * time.Millisecond,
		ListingCreationTime: 200 * time.Millisecond,
		ConcurrentUsers:     150,
		PeakBiddingPeriods: []TimeSlot{
			{
				StartTime: time.Now().Add(-2 * time.Hour),
				EndTime:   time.Now().Add(-1 * time.Hour),
				Activity:  85,
			},
		},
		SystemUptime: 99.9,
		ErrorRate:    0.1,
		TimeRange:    timeRange,
	}, nil
}

// GetBiddingTrends retrieves bidding trend analysis
func (s *AnalyticsService) GetBiddingTrends(ctx context.Context, timeRange string) (*BiddingTrends, error) {
	// Parse time range
	startTime, endTime, err := s.parseTimeRange(timeRange)
	if err != nil {
		return nil, fmt.Errorf("invalid time range: %w", err)
	}

	// Get total bids
	totalBids, err := s.bidRepo.CountBids(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to count total bids: %w", err)
	}

	// Get daily bid counts
	dailyBids, err := s.bidRepo.GetDailyBidCounts(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily bid counts: %w", err)
	}

	// Get hourly bid counts
	hourlyBids, err := s.bidRepo.GetHourlyBidCounts(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get hourly bid counts: %w", err)
	}

	// Get average bid amount
	averageBidAmount, err := s.bidRepo.GetAverageBidAmount(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get average bid amount: %w", err)
	}

	// Get auto-bid percentage
	autoBidCount, err := s.bidRepo.CountAutoBids(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to count auto bids: %w", err)
	}

	var autoBidPercentage float64
	if totalBids > 0 {
		autoBidPercentage = float64(autoBidCount) / float64(totalBids) * 100
	}

	// Convert maps to slices
	dailyBidCounts := make([]DailyBidCount, 0, len(dailyBids))
	for date, count := range dailyBids {
		dailyBidCounts = append(dailyBidCounts, DailyBidCount{
			Date:  date,
			Count: count,
		})
	}

	hourlyBidCounts := make([]HourlyBidCount, 0, len(hourlyBids))
	for hour, count := range hourlyBids {
		hourlyBidCounts = append(hourlyBidCounts, HourlyBidCount{
			Hour:  hour,
			Count: count,
		})
	}

	// Calculate peak bidding hours
	peakHours := s.calculatePeakHours(hourlyBidCounts)

	return &BiddingTrends{
		TotalBids:         totalBids,
		BidsPerDay:        dailyBidCounts,
		BidsPerHour:       hourlyBidCounts,
		AverageBidAmount:  averageBidAmount,
		BidAmountTrend:    []BidAmountTrend{}, // Would be calculated from historical data
		AutoBidPercentage: autoBidPercentage,
		PeakBiddingHours:  peakHours,
		SeasonalTrends:    make(map[string]interface{}),
		TimeRange:         timeRange,
	}, nil
}

// GetPopularCategories retrieves popular category statistics
func (s *AnalyticsService) GetPopularCategories(ctx context.Context, timeRange string) ([]*CategoryStats, error) {
	// Parse time range
	startTime, endTime, err := s.parseTimeRange(timeRange)
	if err != nil {
		return nil, fmt.Errorf("invalid time range: %w", err)
	}

	// Get category statistics
	categoryStatsMap, err := s.listingRepo.GetCategoryStatistics(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get category statistics: %w", err)
	}

	// Convert map to slice
	categoryStats := make([]*CategoryStats, 0, len(categoryStatsMap))
	for category, count := range categoryStatsMap {
		categoryStats = append(categoryStats, &CategoryStats{
			Category:      category,
			TotalListings: count,
			TotalBids:     0, // Would need additional query to get bid counts per category
		})
	}

	return categoryStats, nil
}

// GetRevenueStats retrieves revenue and financial statistics
func (s *AnalyticsService) GetRevenueStats(ctx context.Context, timeRange string) (*RevenueStats, error) {
	// This would integrate with payment and financial systems
	// For now, return placeholder data
	return &RevenueStats{
		TotalRevenue:      decimal.NewFromFloat(50000.00),
		CommissionRevenue: decimal.NewFromFloat(5000.00),
		ListingFees:       decimal.NewFromFloat(2500.00),
		TransactionFees:   decimal.NewFromFloat(1500.00),
		AverageOrderValue: decimal.NewFromFloat(250.00),
		RevenueGrowthRate: 15.5,
		TopRevenueCategories: []CategoryRevenue{
			{Category: "Products", Revenue: decimal.NewFromFloat(30000.00)},
			{Category: "Services", Revenue: decimal.NewFromFloat(15000.00)},
			{Category: "Labour", Revenue: decimal.NewFromFloat(5000.00)},
		},
		TimeRange: timeRange,
	}, nil
}

// GetCommissionStats retrieves commission-specific statistics
func (s *AnalyticsService) GetCommissionStats(ctx context.Context, timeRange string) (*CommissionStats, error) {
	// This would integrate with commission calculation systems
	// For now, return placeholder data
	return &CommissionStats{
		TotalCommissions:  decimal.NewFromFloat(5000.00),
		AverageCommission: decimal.NewFromFloat(25.00),
		CommissionRate:    10.0,
		TopCommissionUsers: []UserCommission{
			{UserID: "USER_001", Commission: decimal.NewFromFloat(500.00)},
			{UserID: "USER_002", Commission: decimal.NewFromFloat(350.00)},
			{UserID: "USER_003", Commission: decimal.NewFromFloat(275.00)},
		},
		CommissionTrends: []CommissionTrend{},
		TimeRange:        timeRange,
	}, nil
}

// Helper methods

// parseTimeRange parses time range string and returns start and end times
func (s *AnalyticsService) parseTimeRange(timeRange string) (time.Time, time.Time, error) {
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

// calculatePeakHours calculates peak bidding hours from hourly data
func (s *AnalyticsService) calculatePeakHours(hourlyBids []HourlyBidCount) []int {
	if len(hourlyBids) == 0 {
		return []int{}
	}

	// Find the maximum bid count
	maxCount := 0
	for _, hourly := range hourlyBids {
		if hourly.Count > maxCount {
			maxCount = hourly.Count
		}
	}

	// Find hours with bid counts >= 80% of maximum
	threshold := int(float64(maxCount) * 0.8)
	var peakHours []int

	for _, hourly := range hourlyBids {
		if hourly.Count >= threshold {
			peakHours = append(peakHours, hourly.Hour)
		}
	}

	return peakHours
}
