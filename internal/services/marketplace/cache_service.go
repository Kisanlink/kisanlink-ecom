package marketplace

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/marketplace"
)

// CacheService provides caching functionality for marketplace operations
type CacheService interface {
	// Listing caching
	GetListing(ctx context.Context, listingID string) (*marketplace.Listing, error)
	SetListing(ctx context.Context, listing *marketplace.Listing, ttl time.Duration) error
	InvalidateListing(ctx context.Context, listingID string) error

	// Bid caching
	GetHighestBid(ctx context.Context, listingID string) (*marketplace.Bid, error)
	SetHighestBid(ctx context.Context, listingID string, bid *marketplace.Bid, ttl time.Duration) error
	InvalidateHighestBid(ctx context.Context, listingID string) error

	// Bid statistics caching
	GetBidStatistics(ctx context.Context, listingID string) (*marketplace.BidStatistics, error)
	SetBidStatistics(ctx context.Context, listingID string, stats *marketplace.BidStatistics, ttl time.Duration) error
	InvalidateBidStatistics(ctx context.Context, listingID string) error

	// Bulk operations
	InvalidateListingCache(ctx context.Context, listingID string) error
	WarmupCache(ctx context.Context, listingIDs []string) error

	// Cache management
	GetCacheStats() CacheStats
	ClearCache(ctx context.Context) error
}

// CacheStats provides cache performance metrics
type CacheStats struct {
	HitCount    int64
	MissCount   int64
	HitRatio    float64
	TotalKeys   int
	MemoryUsage int64 // in bytes
}

// CacheEntry represents a cached item with expiration
type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
	CreatedAt time.Time
}

// IsExpired checks if the cache entry has expired
func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// inMemoryCacheService implements CacheService using in-memory storage
// In production, this would use Redis or another distributed cache
type inMemoryCacheService struct {
	cache    map[string]*CacheEntry
	cacheMux sync.RWMutex
	stats    CacheStats
	statsMux sync.RWMutex
}

// NewCacheService creates a new cache service
func NewCacheService() CacheService {
	service := &inMemoryCacheService{
		cache: make(map[string]*CacheEntry),
	}

	// Start background cleanup goroutine
	go service.cleanupExpiredEntries()

	return service
}

// GetListing retrieves a listing from cache
func (c *inMemoryCacheService) GetListing(ctx context.Context, listingID string) (*marketplace.Listing, error) {
	key := fmt.Sprintf("listing:%s", listingID)

	entry, found := c.get(key)
	if !found {
		c.recordMiss()
		return nil, fmt.Errorf("listing not found in cache")
	}

	listing, ok := entry.Data.(*marketplace.Listing)
	if !ok {
		c.recordMiss()
		return nil, fmt.Errorf("invalid data type in cache")
	}

	c.recordHit()
	return listing, nil
}

// SetListing stores a listing in cache
func (c *inMemoryCacheService) SetListing(ctx context.Context, listing *marketplace.Listing, ttl time.Duration) error {
	key := fmt.Sprintf("listing:%s", listing.ListingID)
	c.set(key, listing, ttl)
	return nil
}

// InvalidateListing removes a listing from cache
func (c *inMemoryCacheService) InvalidateListing(ctx context.Context, listingID string) error {
	key := fmt.Sprintf("listing:%s", listingID)
	c.delete(key)
	return nil
}

// GetHighestBid retrieves the highest bid from cache
func (c *inMemoryCacheService) GetHighestBid(ctx context.Context, listingID string) (*marketplace.Bid, error) {
	key := fmt.Sprintf("highest_bid:%s", listingID)

	entry, found := c.get(key)
	if !found {
		c.recordMiss()
		return nil, fmt.Errorf("highest bid not found in cache")
	}

	bid, ok := entry.Data.(*marketplace.Bid)
	if !ok {
		c.recordMiss()
		return nil, fmt.Errorf("invalid data type in cache")
	}

	c.recordHit()
	return bid, nil
}

// SetHighestBid stores the highest bid in cache
func (c *inMemoryCacheService) SetHighestBid(ctx context.Context, listingID string, bid *marketplace.Bid, ttl time.Duration) error {
	key := fmt.Sprintf("highest_bid:%s", listingID)
	c.set(key, bid, ttl)
	return nil
}

// InvalidateHighestBid removes the highest bid from cache
func (c *inMemoryCacheService) InvalidateHighestBid(ctx context.Context, listingID string) error {
	key := fmt.Sprintf("highest_bid:%s", listingID)
	c.delete(key)
	return nil
}

// GetBidStatistics retrieves bid statistics from cache
func (c *inMemoryCacheService) GetBidStatistics(ctx context.Context, listingID string) (*marketplace.BidStatistics, error) {
	key := fmt.Sprintf("bid_stats:%s", listingID)

	entry, found := c.get(key)
	if !found {
		c.recordMiss()
		return nil, fmt.Errorf("bid statistics not found in cache")
	}

	stats, ok := entry.Data.(*marketplace.BidStatistics)
	if !ok {
		c.recordMiss()
		return nil, fmt.Errorf("invalid data type in cache")
	}

	c.recordHit()
	return stats, nil
}

// SetBidStatistics stores bid statistics in cache
func (c *inMemoryCacheService) SetBidStatistics(ctx context.Context, listingID string, stats *marketplace.BidStatistics, ttl time.Duration) error {
	key := fmt.Sprintf("bid_stats:%s", listingID)
	c.set(key, stats, ttl)
	return nil
}

// InvalidateBidStatistics removes bid statistics from cache
func (c *inMemoryCacheService) InvalidateBidStatistics(ctx context.Context, listingID string) error {
	key := fmt.Sprintf("bid_stats:%s", listingID)
	c.delete(key)
	return nil
}

// InvalidateListingCache invalidates all cache entries for a listing
func (c *inMemoryCacheService) InvalidateListingCache(ctx context.Context, listingID string) error {
	keys := []string{
		fmt.Sprintf("listing:%s", listingID),
		fmt.Sprintf("highest_bid:%s", listingID),
		fmt.Sprintf("bid_stats:%s", listingID),
	}

	for _, key := range keys {
		c.delete(key)
	}

	return nil
}

// WarmupCache preloads cache with frequently accessed data
func (c *inMemoryCacheService) WarmupCache(ctx context.Context, listingIDs []string) error {
	// In a real implementation, this would fetch data from the database
	// and populate the cache for the given listing IDs
	fmt.Printf("Cache warmup requested for %d listings\n", len(listingIDs))
	return nil
}

// GetCacheStats returns cache performance statistics
func (c *inMemoryCacheService) GetCacheStats() CacheStats {
	c.statsMux.RLock()
	defer c.statsMux.RUnlock()

	c.cacheMux.RLock()
	totalKeys := len(c.cache)
	c.cacheMux.RUnlock()

	stats := c.stats
	stats.TotalKeys = totalKeys

	total := stats.HitCount + stats.MissCount
	if total > 0 {
		stats.HitRatio = float64(stats.HitCount) / float64(total)
	}

	return stats
}

// ClearCache removes all entries from cache
func (c *inMemoryCacheService) ClearCache(ctx context.Context) error {
	c.cacheMux.Lock()
	defer c.cacheMux.Unlock()

	c.cache = make(map[string]*CacheEntry)

	c.statsMux.Lock()
	c.stats = CacheStats{}
	c.statsMux.Unlock()

	return nil
}

// Helper methods

// get retrieves an entry from cache
func (c *inMemoryCacheService) get(key string) (*CacheEntry, bool) {
	c.cacheMux.RLock()
	defer c.cacheMux.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		return nil, false
	}

	if entry.IsExpired() {
		// Remove expired entry
		go func() {
			c.cacheMux.Lock()
			delete(c.cache, key)
			c.cacheMux.Unlock()
		}()
		return nil, false
	}

	return entry, true
}

// set stores an entry in cache
func (c *inMemoryCacheService) set(key string, data interface{}, ttl time.Duration) {
	c.cacheMux.Lock()
	defer c.cacheMux.Unlock()

	entry := &CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(ttl),
		CreatedAt: time.Now(),
	}

	c.cache[key] = entry
}

// delete removes an entry from cache
func (c *inMemoryCacheService) delete(key string) {
	c.cacheMux.Lock()
	defer c.cacheMux.Unlock()

	delete(c.cache, key)
}

// recordHit increments hit counter
func (c *inMemoryCacheService) recordHit() {
	c.statsMux.Lock()
	defer c.statsMux.Unlock()
	c.stats.HitCount++
}

// recordMiss increments miss counter
func (c *inMemoryCacheService) recordMiss() {
	c.statsMux.Lock()
	defer c.statsMux.Unlock()
	c.stats.MissCount++
}

// cleanupExpiredEntries periodically removes expired cache entries
func (c *inMemoryCacheService) cleanupExpiredEntries() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cacheMux.Lock()

		expiredKeys := make([]string, 0)
		for key, entry := range c.cache {
			if entry.IsExpired() {
				expiredKeys = append(expiredKeys, key)
			}
		}

		for _, key := range expiredKeys {
			delete(c.cache, key)
		}

		c.cacheMux.Unlock()

		if len(expiredKeys) > 0 {
			fmt.Printf("Cleaned up %d expired cache entries\n", len(expiredKeys))
		}
	}
}

// CachedBiddingService wraps BiddingService with caching
type CachedBiddingService struct {
	BiddingServiceInterface
	cache CacheService
}

// NewCachedBiddingService creates a bidding service with caching
func NewCachedBiddingService(biddingService BiddingServiceInterface, cache CacheService) BiddingServiceInterface {
	return &CachedBiddingService{
		BiddingServiceInterface: biddingService,
		cache:                   cache,
	}
}

// GetHighestBid retrieves the highest bid with caching
func (c *CachedBiddingService) GetHighestBid(ctx context.Context, listingID string) (*marketplace.Bid, error) {
	// Try cache first
	if bid, err := c.cache.GetHighestBid(ctx, listingID); err == nil {
		return bid, nil
	}

	// Cache miss, get from underlying service
	bid, err := c.BiddingServiceInterface.GetHighestBid(ctx, listingID)
	if err != nil {
		return nil, err
	}

	// Cache the result for 30 seconds (short TTL for frequently changing data)
	if bid != nil {
		_ = c.cache.SetHighestBid(ctx, listingID, bid, 30*time.Second)
	}

	return bid, nil
}

// GetBidStatistics retrieves bid statistics with caching
func (c *CachedBiddingService) GetBidStatistics(ctx context.Context, listingID string) (*marketplace.BidStatistics, error) {
	// Try cache first
	if stats, err := c.cache.GetBidStatistics(ctx, listingID); err == nil {
		return stats, nil
	}

	// Cache miss, get from underlying service
	stats, err := c.BiddingServiceInterface.GetBidStatistics(ctx, listingID)
	if err != nil {
		return nil, err
	}

	// Cache the result for 2 minutes
	if stats != nil {
		_ = c.cache.SetBidStatistics(ctx, listingID, stats, 2*time.Minute)
	}

	return stats, nil
}

// PlaceBid places a bid and invalidates relevant cache entries
func (c *CachedBiddingService) PlaceBid(ctx context.Context, req *PlaceBidRequest) (*marketplace.Bid, error) {
	// Place the bid using the underlying service
	bid, err := c.BiddingServiceInterface.PlaceBid(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache entries that are now stale
	_ = c.cache.InvalidateHighestBid(ctx, req.ListingID)
	_ = c.cache.InvalidateBidStatistics(ctx, req.ListingID)

	return bid, nil
}
