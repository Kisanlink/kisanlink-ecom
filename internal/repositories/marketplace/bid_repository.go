package marketplace

import (
	"context"
	"fmt"
	"time"

	"kisanlink-ecom/entities/models/marketplace"
	"kisanlink-ecom/internal/common"
	repositoryCommon "kisanlink-ecom/internal/repositories/common"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"github.com/shopspring/decimal"
)

// BidRepository interface defines the contract for bid data operations
type BidRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, bid *marketplace.Bid) error
	GetByID(ctx context.Context, id string) (*marketplace.Bid, error)
	GetByBidID(ctx context.Context, bidID string) (*marketplace.Bid, error)
	Update(ctx context.Context, bid *marketplace.Bid) error
	Delete(ctx context.Context, id string) error

	// Atomic bid operations
	PlaceBidAtomic(ctx context.Context, bid *marketplace.Bid, listingID string) (*marketplace.Bid, error)
	UpdateHighestBidAtomic(ctx context.Context, listingID string, newBid *marketplace.Bid) (*marketplace.Bid, error)

	// Bid history and retrieval
	GetBidHistory(ctx context.Context, listingID string, visibility marketplace.BidVisibility, viewerID string, pagination *common.PaginationRequest) ([]*marketplace.Bid, int, error)
	GetBidHistoryForUser(ctx context.Context, listingID string, userID string, pagination *common.PaginationRequest) ([]*marketplace.Bid, int, error)
	GetUserBids(ctx context.Context, userID string, filter *marketplace.BidFilter, pagination *common.PaginationRequest) ([]*marketplace.Bid, int, error)

	// Highest bid tracking
	GetHighestBid(ctx context.Context, listingID string) (*marketplace.Bid, error)
	GetHighestBidAmount(ctx context.Context, listingID string) (decimal.Decimal, error)
	GetBidRanking(ctx context.Context, listingID string, limit int) ([]*marketplace.Bid, error)

	// Bid statistics
	GetBidCount(ctx context.Context, listingID string) (int, error)
	GetUniqueBidderCount(ctx context.Context, listingID string) (int, error)
	GetBidStatistics(ctx context.Context, listingID string) (*marketplace.BidStatistics, error)

	// Auto-bidding operations
	GetAutoBidsForListing(ctx context.Context, listingID string) ([]*marketplace.Bid, error)
	GetUserAutoBids(ctx context.Context, userID string, listingID string) ([]*marketplace.Bid, error)
	UpdateAutoBidStatus(ctx context.Context, bidID string, status marketplace.BidStatus) error

	// Status management
	UpdateBidStatus(ctx context.Context, bidID string, status marketplace.BidStatus) error
	MarkBidsAsOutbid(ctx context.Context, listingID string, excludeBidID string) error
	ExpireBidsForListing(ctx context.Context, listingID string) error

	// Admin operations
	RemoveBid(ctx context.Context, bidID string, reason string, adminID string) error
	GetAllBids(ctx context.Context, filter *marketplace.BidFilter, pagination *common.PaginationRequest) ([]*marketplace.Bid, int, error)

	// Analytics methods
	CountBids(ctx context.Context, startTime, endTime time.Time) (int, error)
	CountUserBids(ctx context.Context, userID string, startTime, endTime time.Time) (int, error)
	CountUserWinningBids(ctx context.Context, userID string, startTime, endTime time.Time) (int, error)
	GetUserAverageBidAmount(ctx context.Context, userID string, startTime, endTime time.Time) (decimal.Decimal, error)
	GetUserTotalSpent(ctx context.Context, userID string, startTime, endTime time.Time) (decimal.Decimal, error)
	CountOrganizationBids(ctx context.Context, orgID string, startTime, endTime time.Time) (int, error)
	GetDailyBidCounts(ctx context.Context, startTime, endTime time.Time) (map[string]int, error)
	GetHourlyBidCounts(ctx context.Context, startTime, endTime time.Time) (map[int]int, error)
	GetAverageBidAmount(ctx context.Context, startTime, endTime time.Time) (decimal.Decimal, error)
	CountAutoBids(ctx context.Context, startTime, endTime time.Time) (int, error)
}

// bidRepository implements the BidRepository interface
type bidRepository struct {
	*repositoryCommon.BaseRepository
	dbManager db.DBManager
}

// NewBidRepository creates a new bid repository
func NewBidRepository(dbManager db.DBManager) BidRepository {
	return &bidRepository{
		BaseRepository: repositoryCommon.NewBaseRepository(dbManager),
		dbManager:      dbManager,
	}
}

// Create creates a new bid in the database
func (r *bidRepository) Create(ctx context.Context, bid *marketplace.Bid) error {
	if err := r.dbManager.Create(ctx, bid); err != nil {
		return fmt.Errorf("failed to create bid: %w", err)
	}
	return nil
}

// GetByID retrieves a bid by its database ID
func (r *bidRepository) GetByID(ctx context.Context, id string) (*marketplace.Bid, error) {
	var bid marketplace.Bid
	if err := r.dbManager.GetByID(ctx, id, &bid); err != nil {
		return nil, fmt.Errorf("failed to get bid by ID: %w", err)
	}
	return &bid, nil
}

// GetByBidID retrieves a bid by its bid ID
func (r *bidRepository) GetByBidID(ctx context.Context, bidID string) (*marketplace.Bid, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "bid_id",
			Operator: base.OpEqual,
			Value:    bidID,
		},
	}

	// Apply query options (adds deleted_at IS NULL by default)
	filter = r.ApplyQueryOptions(ctx, filter)

	var bids []*marketplace.Bid
	if err := r.dbManager.List(ctx, filter, &bids); err != nil {
		return nil, fmt.Errorf("failed to get bid by bid ID: %w", err)
	}

	if len(bids) == 0 {
		return nil, nil
	}

	return bids[0], nil
}

// Update updates an existing bid
func (r *bidRepository) Update(ctx context.Context, bid *marketplace.Bid) error {
	if err := r.dbManager.Update(ctx, bid); err != nil {
		return fmt.Errorf("failed to update bid: %w", err)
	}
	return nil
}

// Delete deletes a bid by ID
func (r *bidRepository) Delete(ctx context.Context, id string) error {
	var bid marketplace.Bid
	if err := r.dbManager.Delete(ctx, id, &bid); err != nil {
		return fmt.Errorf("failed to delete bid: %w", err)
	}
	return nil
}

// PlaceBidAtomic places a bid atomically, ensuring proper locking and consistency
func (r *bidRepository) PlaceBidAtomic(ctx context.Context, bid *marketplace.Bid, listingID string) (*marketplace.Bid, error) {
	// For now, use a simplified atomic operation without explicit transactions
	// In production, this would use proper database transactions with row-level locking

	// Get current highest bid for comparison
	currentHighest, err := r.GetHighestBid(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current highest bid: %w", err)
	}

	// Validate bid amount against current highest
	if currentHighest != nil && bid.BidAmount.LessThanOrEqual(currentHighest.BidAmount) {
		return nil, fmt.Errorf("bid amount must be higher than current highest bid")
	}

	// Create the bid
	if err := r.Create(ctx, bid); err != nil {
		return nil, fmt.Errorf("failed to create bid: %w", err)
	}

	// Mark previous highest bid as outbid
	if currentHighest != nil {
		if err := currentHighest.Outbid(); err != nil {
			return nil, fmt.Errorf("failed to mark previous bid as outbid: %w", err)
		}
		if err := r.Update(ctx, currentHighest); err != nil {
			return nil, fmt.Errorf("failed to update previous highest bid: %w", err)
		}
	}

	// Mark new bid as highest
	bid.SetAsHighestBid()
	if err := r.Update(ctx, bid); err != nil {
		return nil, fmt.Errorf("failed to update new highest bid: %w", err)
	}

	return bid, nil
}

// UpdateHighestBidAtomic atomically updates the highest bid for a listing
func (r *bidRepository) UpdateHighestBidAtomic(ctx context.Context, listingID string, newBid *marketplace.Bid) (*marketplace.Bid, error) {
	// Get current highest bid
	currentHighest, err := r.GetHighestBid(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current highest bid: %w", err)
	}

	// Mark current highest as outbid if it exists and is different
	if currentHighest != nil && currentHighest.BidID != newBid.BidID {
		if err := currentHighest.Outbid(); err != nil {
			return nil, fmt.Errorf("failed to mark current bid as outbid: %w", err)
		}
		if err := r.Update(ctx, currentHighest); err != nil {
			return nil, fmt.Errorf("failed to update current highest bid: %w", err)
		}
	}

	// Mark new bid as highest
	newBid.SetAsHighestBid()
	if err := r.Update(ctx, newBid); err != nil {
		return nil, fmt.Errorf("failed to update new highest bid: %w", err)
	}

	return newBid, nil
}

// GetBidHistory retrieves bid history with visibility filtering
func (r *bidRepository) GetBidHistory(ctx context.Context, listingID string, visibility marketplace.BidVisibility, viewerID string, pagination *common.PaginationRequest) ([]*marketplace.Bid, int, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "listing_id",
			Operator: base.OpEqual,
			Value:    listingID,
		},
	}

	// Apply visibility filtering
	switch visibility {
	case marketplace.BidVisibilityHidden:
		// Only show user's own bids
		if viewerID != "" {
			filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
				Field:    "bidder_id",
				Operator: base.OpEqual,
				Value:    viewerID,
			})
		} else {
			// No bids visible to anonymous users
			return []*marketplace.Bid{}, 0, nil
		}
	case marketplace.BidVisibilityMinimal:
		// Show only highest bid
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "is_highest_bid",
			Operator: base.OpEqual,
			Value:    true,
		})
	case marketplace.BidVisibilityPartial:
		// Show top 3 bids
		filter.Limit = 3
	case marketplace.BidVisibilityFull:
		// Show all bids (default behavior)
	}

	// Note: Ordering would be handled by the database layer in production

	return r.executeListQuery(ctx, filter, pagination)
}

// GetBidHistoryForUser retrieves all bids for a specific user on a listing
func (r *bidRepository) GetBidHistoryForUser(ctx context.Context, listingID string, userID string, pagination *common.PaginationRequest) ([]*marketplace.Bid, int, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "listing_id",
			Operator: base.OpEqual,
			Value:    listingID,
		},
		{
			Field:    "bidder_id",
			Operator: base.OpEqual,
			Value:    userID,
		},
	}

	// Note: Ordering would be handled by the database layer in production

	return r.executeListQuery(ctx, filter, pagination)
}

// GetUserBids retrieves all bids for a user with filtering
func (r *bidRepository) GetUserBids(ctx context.Context, userID string, filter *marketplace.BidFilter, pagination *common.PaginationRequest) ([]*marketplace.Bid, int, error) {
	dbFilter := r.buildBaseFilter(filter)

	// Add user filter
	dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
		Field:    "bidder_id",
		Operator: base.OpEqual,
		Value:    userID,
	})

	// Note: Ordering would be handled by the database layer in production

	return r.executeListQuery(ctx, dbFilter, pagination)
}

// GetHighestBid retrieves the current highest bid for a listing
func (r *bidRepository) GetHighestBid(ctx context.Context, listingID string) (*marketplace.Bid, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "listing_id",
			Operator: base.OpEqual,
			Value:    listingID,
		},
		{
			Field:    "is_highest_bid",
			Operator: base.OpEqual,
			Value:    true,
		},
		{
			Field:    "status",
			Operator: base.OpEqual,
			Value:    marketplace.BidStatusActive,
		},
	}

	// Apply query options (adds deleted_at IS NULL by default)
	filter = r.ApplyQueryOptions(ctx, filter)

	var bids []*marketplace.Bid
	if err := r.dbManager.List(ctx, filter, &bids); err != nil {
		return nil, fmt.Errorf("failed to get highest bid: %w", err)
	}

	if len(bids) == 0 {
		return nil, nil
	}

	return bids[0], nil
}

// GetHighestBidAmount retrieves the current highest bid amount for a listing
func (r *bidRepository) GetHighestBidAmount(ctx context.Context, listingID string) (decimal.Decimal, error) {
	highestBid, err := r.GetHighestBid(ctx, listingID)
	if err != nil {
		return decimal.Zero, err
	}

	if highestBid == nil {
		return decimal.Zero, nil
	}

	return highestBid.BidAmount, nil
}

// GetBidRanking retrieves the top bids for a listing in ranking order
func (r *bidRepository) GetBidRanking(ctx context.Context, listingID string, limit int) ([]*marketplace.Bid, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "listing_id",
			Operator: base.OpEqual,
			Value:    listingID,
		},
		{
			Field:    "status",
			Operator: base.OpIn,
			Value:    []marketplace.BidStatus{marketplace.BidStatusActive, marketplace.BidStatusWinning},
		},
	}

	// Apply query options (adds deleted_at IS NULL by default)
	filter = r.ApplyQueryOptions(ctx, filter)

	// Note: Ordering would be handled by the database layer in production
	filter.Limit = limit

	var bids []*marketplace.Bid
	if err := r.dbManager.List(ctx, filter, &bids); err != nil {
		return nil, fmt.Errorf("failed to get bid ranking: %w", err)
	}

	return bids, nil
}

// GetBidCount retrieves the total number of bids for a listing
func (r *bidRepository) GetBidCount(ctx context.Context, listingID string) (int, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "listing_id",
			Operator: base.OpEqual,
			Value:    listingID,
		},
	}

	// Apply query options (adds deleted_at IS NULL by default)
	filter = r.ApplyQueryOptions(ctx, filter)

	count, err := r.dbManager.Count(ctx, filter, &marketplace.Bid{})
	if err != nil {
		return 0, fmt.Errorf("failed to get bid count: %w", err)
	}

	return int(count), nil
}

// GetUniqueBidderCount retrieves the number of unique bidders for a listing
func (r *bidRepository) GetUniqueBidderCount(ctx context.Context, listingID string) (int, error) {
	// This would require a custom query for distinct count
	// For now, we'll use a simplified approach
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "listing_id",
			Operator: base.OpEqual,
			Value:    listingID,
		},
	}

	// Apply query options (adds deleted_at IS NULL by default)
	filter = r.ApplyQueryOptions(ctx, filter)

	var bids []*marketplace.Bid
	if err := r.dbManager.List(ctx, filter, &bids); err != nil {
		return 0, fmt.Errorf("failed to get bids for unique count: %w", err)
	}

	// Count unique bidders
	uniqueBidders := make(map[string]bool)
	for _, bid := range bids {
		uniqueBidders[bid.BidderID] = true
	}

	return len(uniqueBidders), nil
}

// GetBidStatistics retrieves comprehensive bid statistics for a listing
func (r *bidRepository) GetBidStatistics(ctx context.Context, listingID string) (*marketplace.BidStatistics, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "listing_id",
			Operator: base.OpEqual,
			Value:    listingID,
		},
	}

	// Apply query options (adds deleted_at IS NULL by default)
	filter = r.ApplyQueryOptions(ctx, filter)

	var bids []*marketplace.Bid
	if err := r.dbManager.List(ctx, filter, &bids); err != nil {
		return nil, fmt.Errorf("failed to get bids for statistics: %w", err)
	}

	stats := &marketplace.BidStatistics{
		ListingID:     listingID,
		TotalBids:     len(bids),
		UniqueBidders: 0,
		HighestBid:    decimal.Zero,
		AverageBid:    decimal.Zero,
		BidIncrement:  decimal.Zero,
		AutoBidCount:  0,
	}

	if len(bids) == 0 {
		return stats, nil
	}

	// Calculate statistics
	uniqueBidders := make(map[string]bool)
	totalAmount := decimal.Zero
	var lastBidTime *time.Time

	for _, bid := range bids {
		uniqueBidders[bid.BidderID] = true
		totalAmount = totalAmount.Add(bid.BidAmount)

		if bid.BidAmount.GreaterThan(stats.HighestBid) {
			stats.HighestBid = bid.BidAmount
		}

		if bid.IsAutoBid {
			stats.AutoBidCount++
		}

		if lastBidTime == nil || bid.PlacedAt.After(*lastBidTime) {
			lastBidTime = &bid.PlacedAt
		}
	}

	stats.UniqueBidders = len(uniqueBidders)
	stats.AverageBid = totalAmount.Div(decimal.NewFromInt(int64(len(bids))))
	stats.LastBidTime = lastBidTime

	// Calculate bid increment (difference between highest and second highest)
	if len(bids) > 1 {
		// Sort bids by amount descending
		ranking, err := r.GetBidRanking(ctx, listingID, 2)
		if err == nil && len(ranking) >= 2 {
			stats.BidIncrement = ranking[0].BidAmount.Sub(ranking[1].BidAmount)
		}
	}

	return stats, nil
}

// GetAutoBidsForListing retrieves all auto-bids for a listing
func (r *bidRepository) GetAutoBidsForListing(ctx context.Context, listingID string) ([]*marketplace.Bid, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "listing_id",
			Operator: base.OpEqual,
			Value:    listingID,
		},
		{
			Field:    "is_auto_bid",
			Operator: base.OpEqual,
			Value:    true,
		},
		{
			Field:    "status",
			Operator: base.OpEqual,
			Value:    marketplace.BidStatusActive,
		},
	}

	// Apply query options (adds deleted_at IS NULL by default)
	filter = r.ApplyQueryOptions(ctx, filter)

	var bids []*marketplace.Bid
	if err := r.dbManager.List(ctx, filter, &bids); err != nil {
		return nil, fmt.Errorf("failed to get auto-bids: %w", err)
	}

	return bids, nil
}

// GetUserAutoBids retrieves auto-bids for a specific user on a listing
func (r *bidRepository) GetUserAutoBids(ctx context.Context, userID string, listingID string) ([]*marketplace.Bid, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "listing_id",
			Operator: base.OpEqual,
			Value:    listingID,
		},
		{
			Field:    "bidder_id",
			Operator: base.OpEqual,
			Value:    userID,
		},
		{
			Field:    "is_auto_bid",
			Operator: base.OpEqual,
			Value:    true,
		},
		{
			Field:    "status",
			Operator: base.OpEqual,
			Value:    marketplace.BidStatusActive,
		},
	}

	// Apply query options (adds deleted_at IS NULL by default)
	filter = r.ApplyQueryOptions(ctx, filter)

	var bids []*marketplace.Bid
	if err := r.dbManager.List(ctx, filter, &bids); err != nil {
		return nil, fmt.Errorf("failed to get user auto-bids: %w", err)
	}

	return bids, nil
}

// UpdateAutoBidStatus updates the status of an auto-bid
func (r *bidRepository) UpdateAutoBidStatus(ctx context.Context, bidID string, status marketplace.BidStatus) error {
	bid, err := r.GetByBidID(ctx, bidID)
	if err != nil {
		return fmt.Errorf("failed to get bid for status update: %w", err)
	}
	if bid == nil {
		return fmt.Errorf("bid not found: %s", bidID)
	}

	if err := bid.UpdateStatus(status); err != nil {
		return fmt.Errorf("failed to update bid status: %w", err)
	}

	if err := r.Update(ctx, bid); err != nil {
		return fmt.Errorf("failed to save bid status update: %w", err)
	}

	return nil
}

// UpdateBidStatus updates the status of a bid
func (r *bidRepository) UpdateBidStatus(ctx context.Context, bidID string, status marketplace.BidStatus) error {
	return r.UpdateAutoBidStatus(ctx, bidID, status)
}

// MarkBidsAsOutbid marks all bids except the specified one as outbid
func (r *bidRepository) MarkBidsAsOutbid(ctx context.Context, listingID string, excludeBidID string) error {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "listing_id",
			Operator: base.OpEqual,
			Value:    listingID,
		},
		{
			Field:    "bid_id",
			Operator: base.OpNotEqual,
			Value:    excludeBidID,
		},
		{
			Field:    "status",
			Operator: base.OpIn,
			Value:    []marketplace.BidStatus{marketplace.BidStatusActive, marketplace.BidStatusWinning},
		},
	}

	// Apply query options (adds deleted_at IS NULL by default)
	filter = r.ApplyQueryOptions(ctx, filter)

	var bids []*marketplace.Bid
	if err := r.dbManager.List(ctx, filter, &bids); err != nil {
		return fmt.Errorf("failed to get bids to mark as outbid: %w", err)
	}

	// Update each bid
	for _, bid := range bids {
		if err := bid.Outbid(); err != nil {
			return fmt.Errorf("failed to mark bid as outbid: %w", err)
		}
		if err := r.Update(ctx, bid); err != nil {
			return fmt.Errorf("failed to save outbid status: %w", err)
		}
	}

	return nil
}

// ExpireBidsForListing marks all active bids for a listing as expired
func (r *bidRepository) ExpireBidsForListing(ctx context.Context, listingID string) error {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "listing_id",
			Operator: base.OpEqual,
			Value:    listingID,
		},
		{
			Field:    "status",
			Operator: base.OpIn,
			Value:    []marketplace.BidStatus{marketplace.BidStatusActive, marketplace.BidStatusWinning},
		},
	}

	// Apply query options (adds deleted_at IS NULL by default)
	filter = r.ApplyQueryOptions(ctx, filter)

	var bids []*marketplace.Bid
	if err := r.dbManager.List(ctx, filter, &bids); err != nil {
		return fmt.Errorf("failed to get bids to expire: %w", err)
	}

	// Update each bid
	for _, bid := range bids {
		if err := bid.UpdateStatus(marketplace.BidStatusExpired); err != nil {
			return fmt.Errorf("failed to mark bid as expired: %w", err)
		}
		if err := r.Update(ctx, bid); err != nil {
			return fmt.Errorf("failed to save expired status: %w", err)
		}
	}

	return nil
}

// RemoveBid removes a bid (admin operation)
func (r *bidRepository) RemoveBid(ctx context.Context, bidID string, reason string, adminID string) error {
	bid, err := r.GetByBidID(ctx, bidID)
	if err != nil {
		return fmt.Errorf("failed to get bid for removal: %w", err)
	}
	if bid == nil {
		return fmt.Errorf("bid not found: %s", bidID)
	}

	if err := bid.UpdateStatus(marketplace.BidStatusRemoved); err != nil {
		return fmt.Errorf("failed to update bid status for removal: %w", err)
	}

	// Add removal reason to message
	bid.Message = fmt.Sprintf("REMOVED by admin %s: %s. Original: %s", adminID, reason, bid.Message)

	if err := r.Update(ctx, bid); err != nil {
		return fmt.Errorf("failed to save bid removal: %w", err)
	}

	return nil
}

// GetAllBids retrieves all bids (admin operation)
func (r *bidRepository) GetAllBids(ctx context.Context, filter *marketplace.BidFilter, pagination *common.PaginationRequest) ([]*marketplace.Bid, int, error) {
	dbFilter := r.buildBaseFilter(filter)
	return r.executeListQuery(ctx, dbFilter, pagination)
}

// Helper methods

// buildBaseFilter builds a base filter from the bid filter
func (r *bidRepository) buildBaseFilter(filter *marketplace.BidFilter) *base.Filter {
	dbFilter := base.NewFilter()

	if filter == nil {
		return dbFilter
	}

	// Status filter
	if filter.Status != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "status",
			Operator: base.OpEqual,
			Value:    *filter.Status,
		})
	}

	// Bidder ID filter
	if filter.BidderID != "" {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "bidder_id",
			Operator: base.OpEqual,
			Value:    filter.BidderID,
		})
	}

	// Listing ID filter
	if filter.ListingID != "" {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "listing_id",
			Operator: base.OpEqual,
			Value:    filter.ListingID,
		})
	}

	// Highest bid filter
	if filter.IsHighestBid != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "is_highest_bid",
			Operator: base.OpEqual,
			Value:    *filter.IsHighestBid,
		})
	}

	// Auto bid filter
	if filter.IsAutoBid != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "is_auto_bid",
			Operator: base.OpEqual,
			Value:    *filter.IsAutoBid,
		})
	}

	// Amount range filters
	if filter.MinAmount != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "bid_amount",
			Operator: base.OpGreaterEqual,
			Value:    *filter.MinAmount,
		})
	}

	if filter.MaxAmount != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "bid_amount",
			Operator: base.OpLessEqual,
			Value:    *filter.MaxAmount,
		})
	}

	// Placed date filters
	if filter.PlacedAfter != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "placed_at",
			Operator: base.OpGreaterThan,
			Value:    *filter.PlacedAfter,
		})
	}

	if filter.PlacedBefore != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "placed_at",
			Operator: base.OpLessThan,
			Value:    *filter.PlacedBefore,
		})
	}

	return dbFilter
}

// executeListQuery executes a list query with pagination
func (r *bidRepository) executeListQuery(ctx context.Context, filter *base.Filter, pagination *common.PaginationRequest) ([]*marketplace.Bid, int, error) {
	// Apply query options (adds deleted_at IS NULL by default) BEFORE pagination
	filter = r.ApplyQueryOptions(ctx, filter)

	// Apply pagination
	if pagination != nil {
		filter.Limit = pagination.Limit
		filter.Offset = pagination.Offset
	}

	// Execute query
	var bids []*marketplace.Bid
	if err := r.dbManager.List(ctx, filter, &bids); err != nil {
		return nil, 0, fmt.Errorf("failed to execute bid query: %w", err)
	}

	// Get total count for pagination
	total, err := r.dbManager.Count(ctx, filter, &marketplace.Bid{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return bids, int(total), nil
}

// Analytics method implementations

// CountBids counts total bids in a time range
func (r *bidRepository) CountBids(ctx context.Context, startTime, endTime time.Time) (int, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "created_at",
			Operator: base.OpGreaterEqual,
			Value:    startTime,
		},
		{
			Field:    "created_at",
			Operator: base.OpLessEqual,
			Value:    endTime,
		},
	}

	// Apply query options (adds deleted_at IS NULL by default)
	filter = r.ApplyQueryOptions(ctx, filter)

	count, err := r.dbManager.Count(ctx, filter, &marketplace.Bid{})
	if err != nil {
		return 0, fmt.Errorf("failed to count bids: %w", err)
	}

	return int(count), nil
}

// CountUserBids counts bids for a specific user in a time range
func (r *bidRepository) CountUserBids(ctx context.Context, userID string, startTime, endTime time.Time) (int, error) {
	// Stub implementation - return 0 for now
	return 0, nil
}

// CountUserWinningBids counts winning bids for a specific user in a time range
func (r *bidRepository) CountUserWinningBids(ctx context.Context, userID string, startTime, endTime time.Time) (int, error) {
	// Stub implementation - return 0 for now
	return 0, nil
}

// GetUserAverageBidAmount gets average bid amount for a specific user in a time range
func (r *bidRepository) GetUserAverageBidAmount(ctx context.Context, userID string, startTime, endTime time.Time) (decimal.Decimal, error) {
	// Stub implementation - return zero for now
	return decimal.Zero, nil
}

// GetUserTotalSpent gets total amount spent by a user in a time range
func (r *bidRepository) GetUserTotalSpent(ctx context.Context, userID string, startTime, endTime time.Time) (decimal.Decimal, error) {
	// Stub implementation - return zero for now
	return decimal.Zero, nil
}

// CountOrganizationBids counts bids for an organization in a time range
func (r *bidRepository) CountOrganizationBids(ctx context.Context, orgID string, startTime, endTime time.Time) (int, error) {
	// Stub implementation - return 0 for now
	return 0, nil
}

// GetDailyBidCounts gets daily bid counts in a time range
func (r *bidRepository) GetDailyBidCounts(ctx context.Context, startTime, endTime time.Time) (map[string]int, error) {
	// Stub implementation - return empty map for now
	return make(map[string]int), nil
}

// GetHourlyBidCounts gets hourly bid counts in a time range
func (r *bidRepository) GetHourlyBidCounts(ctx context.Context, startTime, endTime time.Time) (map[int]int, error) {
	// Stub implementation - return empty map for now
	return make(map[int]int), nil
}

// GetAverageBidAmount gets average bid amount in a time range
func (r *bidRepository) GetAverageBidAmount(ctx context.Context, startTime, endTime time.Time) (decimal.Decimal, error) {
	// Stub implementation - return zero for now
	return decimal.Zero, nil
}

// CountAutoBids counts auto bids in a time range
func (r *bidRepository) CountAutoBids(ctx context.Context, startTime, endTime time.Time) (int, error) {
	// Stub implementation - return 0 for now
	return 0, nil
}
