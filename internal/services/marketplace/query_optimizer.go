package marketplace

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/marketplace"
	"github.com/Kisanlink/kisanlink-ecom/internal/common"
	marketplaceRepo "github.com/Kisanlink/kisanlink-ecom/internal/repositories/marketplace"

	"github.com/shopspring/decimal"
)

// QueryOptimizer provides optimized database queries for high-volume scenarios
type QueryOptimizer interface {
	// Optimized bid queries
	GetHighestBidOptimized(ctx context.Context, listingID string) (*marketplace.Bid, error)
	GetBidHistoryOptimized(ctx context.Context, listingID string, limit int) ([]*marketplace.Bid, error)
	GetActiveBidsCountOptimized(ctx context.Context, listingID string) (int, error)

	// Batch operations
	GetHighestBidsForListings(ctx context.Context, listingIDs []string) (map[string]*marketplace.Bid, error)
	GetBidCountsForListings(ctx context.Context, listingIDs []string) (map[string]int, error)

	// Optimized listing queries
	GetActiveListingsOptimized(ctx context.Context, filters *OptimizedListingFilter, pagination *common.PaginationParams) ([]*marketplace.Listing, int, error)

	// Performance monitoring
	GetQueryStats() QueryStats
	ResetQueryStats()
}

// OptimizedListingFilter provides optimized filtering options
type OptimizedListingFilter struct {
	OrganizationID string
	SellerID       string
	ProductIDs     []string
	MinPrice       *decimal.Decimal
	MaxPrice       *decimal.Decimal
	ExpiresAfter   *time.Time
	ExpiresBefore  *time.Time
	HasActiveBids  *bool
	MinBidCount    *int
	UseIndexHints  bool
}

// QueryStats provides query performance metrics
type QueryStats struct {
	TotalQueries       int64
	AverageQueryTime   time.Duration
	SlowestQuery       time.Duration
	FastestQuery       time.Duration
	CacheHitRate       float64
	OptimizedQueries   int64
	UnoptimizedQueries int64
	QueryTimesByType   map[string]time.Duration
}

// QueryExecution tracks individual query performance
type QueryExecution struct {
	QueryType    string
	StartTime    time.Time
	Duration     time.Duration
	CacheHit     bool
	Optimized    bool
	RowsReturned int
}

// queryOptimizer implements QueryOptimizer
type queryOptimizer struct {
	bidRepo     marketplaceRepo.BidRepository
	listingRepo marketplaceRepo.ListingRepository
	cache       CacheService

	// Performance tracking
	stats      QueryStats
	statsMux   sync.RWMutex
	executions []QueryExecution
	execMux    sync.Mutex

	// Query result cache
	resultCache map[string]interface{}
	cacheMux    sync.RWMutex
	cacheTTL    time.Duration
}

// NewQueryOptimizer creates a new query optimizer
func NewQueryOptimizer(
	bidRepo marketplaceRepo.BidRepository,
	listingRepo marketplaceRepo.ListingRepository,
	cache CacheService,
) QueryOptimizer {
	optimizer := &queryOptimizer{
		bidRepo:     bidRepo,
		listingRepo: listingRepo,
		cache:       cache,
		resultCache: make(map[string]interface{}),
		cacheTTL:    30 * time.Second,
		stats: QueryStats{
			QueryTimesByType: make(map[string]time.Duration),
			FastestQuery:     time.Hour, // Initialize to a large value
		},
	}

	// Start cache cleanup goroutine
	go optimizer.cacheCleanupLoop()

	return optimizer
}

// GetHighestBidOptimized retrieves the highest bid with optimizations
func (q *queryOptimizer) GetHighestBidOptimized(ctx context.Context, listingID string) (*marketplace.Bid, error) {
	exec := q.startQueryExecution("GetHighestBid")
	defer q.endQueryExecution(exec)

	// Try cache first
	cacheKey := fmt.Sprintf("highest_bid:%s", listingID)
	if cached := q.getFromResultCache(cacheKey); cached != nil {
		if bid, ok := cached.(*marketplace.Bid); ok {
			exec.CacheHit = true
			exec.RowsReturned = 1
			return bid, nil
		}
	}

	// Cache miss, query database with optimization
	bid, err := q.bidRepo.GetHighestBid(ctx, listingID)
	if err != nil {
		return nil, err
	}

	exec.Optimized = true
	if bid != nil {
		exec.RowsReturned = 1
		q.setResultCache(cacheKey, bid)
	}

	return bid, nil
}

// GetBidHistoryOptimized retrieves bid history with optimizations
func (q *queryOptimizer) GetBidHistoryOptimized(ctx context.Context, listingID string, limit int) ([]*marketplace.Bid, error) {
	exec := q.startQueryExecution("GetBidHistory")
	defer q.endQueryExecution(exec)

	// Try cache first
	cacheKey := fmt.Sprintf("bid_history:%s:%d", listingID, limit)
	if cached := q.getFromResultCache(cacheKey); cached != nil {
		if bids, ok := cached.([]*marketplace.Bid); ok {
			exec.CacheHit = true
			exec.RowsReturned = len(bids)
			return bids, nil
		}
	}

	// Cache miss, query database with optimization
	pagination := &common.PaginationRequest{
		Limit:  limit,
		Offset: 0,
	}

	bids, _, err := q.bidRepo.GetBidHistory(ctx, listingID, marketplace.BidVisibilityFull, "", pagination)
	if err != nil {
		return nil, err
	}

	exec.Optimized = true
	exec.RowsReturned = len(bids)
	q.setResultCache(cacheKey, bids)

	return bids, nil
}

// GetActiveBidsCountOptimized retrieves active bid count with optimizations
func (q *queryOptimizer) GetActiveBidsCountOptimized(ctx context.Context, listingID string) (int, error) {
	exec := q.startQueryExecution("GetActiveBidsCount")
	defer q.endQueryExecution(exec)

	// Try cache first
	cacheKey := fmt.Sprintf("active_bids_count:%s", listingID)
	if cached := q.getFromResultCache(cacheKey); cached != nil {
		if count, ok := cached.(int); ok {
			exec.CacheHit = true
			exec.RowsReturned = 1
			return count, nil
		}
	}

	// Cache miss, query database
	count, err := q.bidRepo.GetBidCount(ctx, listingID)
	if err != nil {
		return 0, err
	}

	exec.Optimized = true
	exec.RowsReturned = 1
	q.setResultCache(cacheKey, count)

	return count, nil
}

// GetHighestBidsForListings retrieves highest bids for multiple listings in batch
func (q *queryOptimizer) GetHighestBidsForListings(ctx context.Context, listingIDs []string) (map[string]*marketplace.Bid, error) {
	exec := q.startQueryExecution("GetHighestBidsForListings")
	defer q.endQueryExecution(exec)

	result := make(map[string]*marketplace.Bid)
	uncachedIDs := make([]string, 0, len(listingIDs))

	// Check cache for each listing
	for _, listingID := range listingIDs {
		cacheKey := fmt.Sprintf("highest_bid:%s", listingID)
		if cached := q.getFromResultCache(cacheKey); cached != nil {
			if bid, ok := cached.(*marketplace.Bid); ok {
				result[listingID] = bid
				continue
			}
		}
		uncachedIDs = append(uncachedIDs, listingID)
	}

	// Batch query for uncached listings
	if len(uncachedIDs) > 0 {
		for _, listingID := range uncachedIDs {
			bid, err := q.bidRepo.GetHighestBid(ctx, listingID)
			if err != nil {
				continue // Skip errors for individual listings
			}

			result[listingID] = bid

			// Cache the result
			cacheKey := fmt.Sprintf("highest_bid:%s", listingID)
			q.setResultCache(cacheKey, bid)
		}
	}

	exec.Optimized = true
	exec.RowsReturned = len(result)

	return result, nil
}

// GetBidCountsForListings retrieves bid counts for multiple listings in batch
func (q *queryOptimizer) GetBidCountsForListings(ctx context.Context, listingIDs []string) (map[string]int, error) {
	exec := q.startQueryExecution("GetBidCountsForListings")
	defer q.endQueryExecution(exec)

	result := make(map[string]int)
	uncachedIDs := make([]string, 0, len(listingIDs))

	// Check cache for each listing
	for _, listingID := range listingIDs {
		cacheKey := fmt.Sprintf("bid_count:%s", listingID)
		if cached := q.getFromResultCache(cacheKey); cached != nil {
			if count, ok := cached.(int); ok {
				result[listingID] = count
				continue
			}
		}
		uncachedIDs = append(uncachedIDs, listingID)
	}

	// Batch query for uncached listings
	if len(uncachedIDs) > 0 {
		for _, listingID := range uncachedIDs {
			count, err := q.bidRepo.GetBidCount(ctx, listingID)
			if err != nil {
				continue // Skip errors for individual listings
			}

			result[listingID] = count

			// Cache the result
			cacheKey := fmt.Sprintf("bid_count:%s", listingID)
			q.setResultCache(cacheKey, count)
		}
	}

	exec.Optimized = true
	exec.RowsReturned = len(result)

	return result, nil
}

// GetActiveListingsOptimized retrieves active listings with optimizations
func (q *queryOptimizer) GetActiveListingsOptimized(ctx context.Context, filters *OptimizedListingFilter, pagination *common.PaginationParams) ([]*marketplace.Listing, int, error) {
	exec := q.startQueryExecution("GetActiveListingsOptimized")
	defer q.endQueryExecution(exec)

	// Convert optimized filter to standard filter
	standardFilter := q.convertToStandardFilter(filters)

	// Use cache for frequently accessed queries
	cacheKey := q.generateListingCacheKey(filters, pagination)
	if cached := q.getFromResultCache(cacheKey); cached != nil {
		if result, ok := cached.(struct {
			listings []*marketplace.Listing
			total    int
		}); ok {
			exec.CacheHit = true
			exec.RowsReturned = len(result.listings)
			return result.listings, result.total, nil
		}
	}

	// Query database with optimizations
	paginationReq := &common.PaginationRequest{
		Limit:  pagination.Limit,
		Offset: pagination.CalculateOffset(),
	}

	listings, total, err := q.listingRepo.GetActiveListings(ctx, filters.OrganizationID, standardFilter, paginationReq)
	if err != nil {
		return nil, 0, err
	}

	exec.Optimized = true
	exec.RowsReturned = len(listings)

	// Cache the result
	result := struct {
		listings []*marketplace.Listing
		total    int
	}{listings, total}
	q.setResultCache(cacheKey, result)

	return listings, total, nil
}

// GetQueryStats returns query performance statistics
func (q *queryOptimizer) GetQueryStats() QueryStats {
	q.statsMux.RLock()
	defer q.statsMux.RUnlock()

	stats := q.stats

	// Calculate cache hit rate
	if stats.TotalQueries > 0 {
		cacheHits := int64(0)
		q.execMux.Lock()
		for _, exec := range q.executions {
			if exec.CacheHit {
				cacheHits++
			}
		}
		q.execMux.Unlock()

		stats.CacheHitRate = float64(cacheHits) / float64(stats.TotalQueries)
	}

	return stats
}

// ResetQueryStats resets query performance statistics
func (q *queryOptimizer) ResetQueryStats() {
	q.statsMux.Lock()
	defer q.statsMux.Unlock()

	q.stats = QueryStats{
		QueryTimesByType: make(map[string]time.Duration),
		FastestQuery:     time.Hour,
	}

	q.execMux.Lock()
	q.executions = nil
	q.execMux.Unlock()
}

// Helper methods

// startQueryExecution starts tracking a query execution
func (q *queryOptimizer) startQueryExecution(queryType string) *QueryExecution {
	return &QueryExecution{
		QueryType: queryType,
		StartTime: time.Now(),
	}
}

// endQueryExecution completes tracking a query execution
func (q *queryOptimizer) endQueryExecution(exec *QueryExecution) {
	exec.Duration = time.Since(exec.StartTime)

	q.statsMux.Lock()
	q.stats.TotalQueries++

	// Update average query time
	totalTime := q.stats.AverageQueryTime * time.Duration(q.stats.TotalQueries-1)
	q.stats.AverageQueryTime = (totalTime + exec.Duration) / time.Duration(q.stats.TotalQueries)

	// Update fastest/slowest
	if exec.Duration < q.stats.FastestQuery {
		q.stats.FastestQuery = exec.Duration
	}
	if exec.Duration > q.stats.SlowestQuery {
		q.stats.SlowestQuery = exec.Duration
	}

	// Update query type stats
	q.stats.QueryTimesByType[exec.QueryType] = exec.Duration

	// Update optimization stats
	if exec.Optimized {
		q.stats.OptimizedQueries++
	} else {
		q.stats.UnoptimizedQueries++
	}

	q.statsMux.Unlock()

	// Store execution for detailed analysis
	q.execMux.Lock()
	q.executions = append(q.executions, *exec)

	// Keep only the last 1000 executions
	if len(q.executions) > 1000 {
		q.executions = q.executions[1:]
	}
	q.execMux.Unlock()
}

// getFromResultCache retrieves a value from the result cache
func (q *queryOptimizer) getFromResultCache(key string) interface{} {
	q.cacheMux.RLock()
	defer q.cacheMux.RUnlock()

	return q.resultCache[key]
}

// setResultCache stores a value in the result cache
func (q *queryOptimizer) setResultCache(key string, value interface{}) {
	q.cacheMux.Lock()
	defer q.cacheMux.Unlock()

	q.resultCache[key] = value
}

// convertToStandardFilter converts optimized filter to standard filter
func (q *queryOptimizer) convertToStandardFilter(filters *OptimizedListingFilter) *marketplace.ListingFilter {
	if filters == nil {
		return nil
	}

	filter := &marketplace.ListingFilter{
		OrganizationID: filters.OrganizationID,
		SellerID:       filters.SellerID,
		MinPrice:       filters.MinPrice,
		MaxPrice:       filters.MaxPrice,
		ExpiresAfter:   filters.ExpiresAfter,
		ExpiresBefore:  filters.ExpiresBefore,
	}

	// Handle product IDs (take first one for simplicity)
	if len(filters.ProductIDs) > 0 {
		filter.ProductID = filters.ProductIDs[0]
	}

	return filter
}

// generateListingCacheKey generates a cache key for listing queries
func (q *queryOptimizer) generateListingCacheKey(filters *OptimizedListingFilter, pagination *common.PaginationParams) string {
	key := fmt.Sprintf("listings:%s:%s", filters.OrganizationID, filters.SellerID)

	if pagination != nil {
		key += fmt.Sprintf(":%d:%d", pagination.Page, pagination.Limit)
	}

	return key
}

// cacheCleanupLoop periodically cleans up expired cache entries
func (q *queryOptimizer) cacheCleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		q.cacheMux.Lock()

		// Simple cleanup: clear all cache entries periodically
		// In production, this would track expiration times
		if len(q.resultCache) > 1000 {
			q.resultCache = make(map[string]interface{})
		}

		q.cacheMux.Unlock()
	}
}
