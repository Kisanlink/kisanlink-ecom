package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

// CacheLevel represents different cache levels
type CacheLevel int

const (
	// L1Cache represents in-memory cache (fastest)
	L1Cache CacheLevel = iota
	// L2Cache represents distributed cache (Redis)
	L2Cache
	// L3Cache represents persistent cache (Database)
	L3Cache
)

// CacheConfig holds configuration for caching
type CacheConfig struct {
	// L1Config for in-memory cache
	L1Config *L1CacheConfig
	// L2Config for distributed cache
	L2Config *L2CacheConfig
	// L3Config for persistent cache
	L3Config *L3CacheConfig
	// DefaultTTL is the default time-to-live for cache entries
	DefaultTTL time.Duration
	// MaxRetries for cache operations
	MaxRetries int
	// EnableMetrics enables cache metrics collection
	EnableMetrics bool
}

// L1CacheConfig configures in-memory cache
type L1CacheConfig struct {
	// DefaultExpiration for cache entries
	DefaultExpiration time.Duration
	// CleanupInterval for expired entries
	CleanupInterval time.Duration
	// MaxSize limits the cache size
	MaxSize int64
}

// L2CacheConfig configures distributed cache
type L2CacheConfig struct {
	// Enabled determines if L2 cache is enabled
	Enabled bool
	// DefaultExpiration for cache entries
	DefaultExpiration time.Duration
	// KeyPrefix for cache keys
	KeyPrefix string
	// Serialization format (json, msgpack, etc.)
	Serialization string
}

// L3CacheConfig configures persistent cache
type L3CacheConfig struct {
	// Enabled determines if L3 cache is enabled
	Enabled bool
	// DefaultExpiration for cache entries
	DefaultExpiration time.Duration
	// TableName for persistent cache
	TableName string
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		L1Config: &L1CacheConfig{
			DefaultExpiration: 5 * time.Minute,
			CleanupInterval:   10 * time.Minute,
			MaxSize:           100 * 1024 * 1024, // 100MB
		},
		L2Config: &L2CacheConfig{
			Enabled:           false, // Disabled by default
			DefaultExpiration: 30 * time.Minute,
			KeyPrefix:         "kisanlink:",
			Serialization:     "json",
		},
		L3Config: &L3CacheConfig{
			Enabled:           false, // Disabled by default
			DefaultExpiration: 24 * time.Hour,
			TableName:         "cache_entries",
		},
		DefaultTTL:    5 * time.Minute,
		MaxRetries:    3,
		EnableMetrics: true,
	}
}

// Cache interface for different cache implementations
type Cache interface {
	Get(ctx context.Context, key string) (interface{}, bool, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	Exists(ctx context.Context, key string) (bool, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
	Stats() CacheStats
}

// CacheStats holds cache statistics
type CacheStats struct {
	Hits       int64     `json:"hits"`
	Misses     int64     `json:"misses"`
	Sets       int64     `json:"sets"`
	Deletes    int64     `json:"deletes"`
	Errors     int64     `json:"errors"`
	HitRatio   float64   `json:"hit_ratio"`
	Size       int64     `json:"size"`
	KeyCount   int64     `json:"key_count"`
	LastAccess time.Time `json:"last_access"`
	LastSet    time.Time `json:"last_set"`
}

// L1MemoryCache implements in-memory caching
type L1MemoryCache struct {
	cache   *cache.Cache
	config  *L1CacheConfig
	stats   *CacheStats
	mutex   sync.RWMutex
	maxSize int64
	size    int64
}

// NewL1MemoryCache creates a new in-memory cache
func NewL1MemoryCache(config *L1CacheConfig) *L1MemoryCache {
	if config == nil {
		config = &L1CacheConfig{
			DefaultExpiration: 5 * time.Minute,
			CleanupInterval:   10 * time.Minute,
			MaxSize:           100 * 1024 * 1024,
		}
	}

	return &L1MemoryCache{
		cache:   cache.New(config.DefaultExpiration, config.CleanupInterval),
		config:  config,
		stats:   &CacheStats{},
		maxSize: config.MaxSize,
	}
}

// Get retrieves a value from L1 cache
func (l1 *L1MemoryCache) Get(ctx context.Context, key string) (interface{}, bool, error) {
	l1.mutex.Lock()
	l1.stats.LastAccess = time.Now()
	l1.mutex.Unlock()

	value, found := l1.cache.Get(key)

	l1.mutex.Lock()
	if found {
		l1.stats.Hits++
	} else {
		l1.stats.Misses++
	}
	l1.updateHitRatio()
	l1.mutex.Unlock()

	return value, found, nil
}

// Set stores a value in L1 cache
func (l1 *L1MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Estimate size
	estimatedSize := l1.estimateSize(value)

	l1.mutex.Lock()
	// Check if adding this would exceed max size
	if l1.size+estimatedSize > l1.maxSize {
		// Simple eviction: clear some entries
		l1.cache.DeleteExpired()
		// If still too large, reject
		if l1.size+estimatedSize > l1.maxSize {
			l1.mutex.Unlock()
			return fmt.Errorf("cache size limit exceeded")
		}
	}

	l1.cache.Set(key, value, ttl)
	l1.size += estimatedSize
	l1.stats.Sets++
	l1.stats.KeyCount++
	l1.stats.LastSet = time.Now()
	l1.mutex.Unlock()

	return nil
}

// Delete removes a value from L1 cache
func (l1 *L1MemoryCache) Delete(ctx context.Context, key string) error {
	l1.cache.Delete(key)

	l1.mutex.Lock()
	l1.stats.Deletes++
	l1.stats.KeyCount--
	if l1.stats.KeyCount < 0 {
		l1.stats.KeyCount = 0
	}
	l1.mutex.Unlock()

	return nil
}

// Clear clears all entries from L1 cache
func (l1 *L1MemoryCache) Clear(ctx context.Context) error {
	l1.cache.Flush()

	l1.mutex.Lock()
	l1.size = 0
	l1.stats.KeyCount = 0
	l1.mutex.Unlock()

	return nil
}

// Exists checks if a key exists in L1 cache
func (l1 *L1MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	_, found := l1.cache.Get(key)
	return found, nil
}

// TTL returns the time-to-live for a key
func (l1 *L1MemoryCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	_, expiry, found := l1.cache.GetWithExpiration(key)
	if !found {
		return 0, fmt.Errorf("key not found")
	}

	if expiry.IsZero() {
		return -1, nil // No expiration
	}

	ttl := time.Until(expiry)
	if ttl < 0 {
		return 0, nil // Expired
	}

	return ttl, nil
}

// Stats returns cache statistics
func (l1 *L1MemoryCache) Stats() CacheStats {
	l1.mutex.RLock()
	defer l1.mutex.RUnlock()

	stats := *l1.stats
	stats.Size = l1.size
	return stats
}

// estimateSize estimates the size of a value
func (l1 *L1MemoryCache) estimateSize(value interface{}) int64 {
	// Simple estimation - in production, this could be more sophisticated
	data, err := json.Marshal(value)
	if err != nil {
		return 1024 // Default estimate
	}
	return int64(len(data))
}

// updateHitRatio updates the hit ratio
func (l1 *L1MemoryCache) updateHitRatio() {
	total := l1.stats.Hits + l1.stats.Misses
	if total > 0 {
		l1.stats.HitRatio = float64(l1.stats.Hits) / float64(total)
	}
}

// MultiLevelCache implements multi-level caching
type MultiLevelCache struct {
	l1Cache Cache
	l2Cache Cache
	l3Cache Cache
	config  *CacheConfig
	stats   map[CacheLevel]*CacheStats
	mutex   sync.RWMutex
}

// NewMultiLevelCache creates a new multi-level cache
func NewMultiLevelCache(config *CacheConfig) *MultiLevelCache {
	if config == nil {
		config = DefaultCacheConfig()
	}

	mlc := &MultiLevelCache{
		config: config,
		stats:  make(map[CacheLevel]*CacheStats),
	}

	// Initialize L1 cache (always enabled)
	mlc.l1Cache = NewL1MemoryCache(config.L1Config)

	// Initialize L2 cache if enabled
	if config.L2Config.Enabled {
		// L2 cache would be Redis or similar distributed cache
		// For now, create another in-memory cache as placeholder
		mlc.l2Cache = NewL1MemoryCache(&L1CacheConfig{
			DefaultExpiration: config.L2Config.DefaultExpiration,
			CleanupInterval:   10 * time.Minute,
		})
	}

	// Initialize L3 cache if enabled
	if config.L3Config.Enabled {
		// L3 cache would be database-backed
		// For now, create another in-memory cache as placeholder
		mlc.l3Cache = NewL1MemoryCache(&L1CacheConfig{
			DefaultExpiration: config.L3Config.DefaultExpiration,
			CleanupInterval:   30 * time.Minute,
		})
	}

	return mlc
}

// Get retrieves a value from the multi-level cache
func (mlc *MultiLevelCache) Get(ctx context.Context, key string) (interface{}, bool, error) {
	// Try L1 cache first
	if value, found, err := mlc.l1Cache.Get(ctx, key); err == nil && found {
		return value, true, nil
	}

	// Try L2 cache if enabled
	if mlc.l2Cache != nil {
		if value, found, err := mlc.l2Cache.Get(ctx, key); err == nil && found {
			// Promote to L1 cache
			mlc.l1Cache.Set(ctx, key, value, mlc.config.DefaultTTL)
			return value, true, nil
		}
	}

	// Try L3 cache if enabled
	if mlc.l3Cache != nil {
		if value, found, err := mlc.l3Cache.Get(ctx, key); err == nil && found {
			// Promote to L2 and L1 caches
			if mlc.l2Cache != nil {
				mlc.l2Cache.Set(ctx, key, value, mlc.config.DefaultTTL*2)
			}
			mlc.l1Cache.Set(ctx, key, value, mlc.config.DefaultTTL)
			return value, true, nil
		}
	}

	return nil, false, nil
}

// Set stores a value in all cache levels
func (mlc *MultiLevelCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	var lastErr error

	// Set in L1 cache
	if err := mlc.l1Cache.Set(ctx, key, value, ttl); err != nil {
		lastErr = err
	}

	// Set in L2 cache if enabled
	if mlc.l2Cache != nil {
		if err := mlc.l2Cache.Set(ctx, key, value, ttl*2); err != nil {
			lastErr = err
		}
	}

	// Set in L3 cache if enabled
	if mlc.l3Cache != nil {
		if err := mlc.l3Cache.Set(ctx, key, value, ttl*4); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// Delete removes a value from all cache levels
func (mlc *MultiLevelCache) Delete(ctx context.Context, key string) error {
	var lastErr error

	// Delete from all cache levels
	if err := mlc.l1Cache.Delete(ctx, key); err != nil {
		lastErr = err
	}

	if mlc.l2Cache != nil {
		if err := mlc.l2Cache.Delete(ctx, key); err != nil {
			lastErr = err
		}
	}

	if mlc.l3Cache != nil {
		if err := mlc.l3Cache.Delete(ctx, key); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// Clear clears all cache levels
func (mlc *MultiLevelCache) Clear(ctx context.Context) error {
	var lastErr error

	if err := mlc.l1Cache.Clear(ctx); err != nil {
		lastErr = err
	}

	if mlc.l2Cache != nil {
		if err := mlc.l2Cache.Clear(ctx); err != nil {
			lastErr = err
		}
	}

	if mlc.l3Cache != nil {
		if err := mlc.l3Cache.Clear(ctx); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// Exists checks if a key exists in any cache level
func (mlc *MultiLevelCache) Exists(ctx context.Context, key string) (bool, error) {
	if exists, err := mlc.l1Cache.Exists(ctx, key); err == nil && exists {
		return true, nil
	}

	if mlc.l2Cache != nil {
		if exists, err := mlc.l2Cache.Exists(ctx, key); err == nil && exists {
			return true, nil
		}
	}

	if mlc.l3Cache != nil {
		if exists, err := mlc.l3Cache.Exists(ctx, key); err == nil && exists {
			return true, nil
		}
	}

	return false, nil
}

// TTL returns the time-to-live for a key
func (mlc *MultiLevelCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	// Check L1 first
	if ttl, err := mlc.l1Cache.TTL(ctx, key); err == nil {
		return ttl, nil
	}

	// Check L2
	if mlc.l2Cache != nil {
		if ttl, err := mlc.l2Cache.TTL(ctx, key); err == nil {
			return ttl, nil
		}
	}

	// Check L3
	if mlc.l3Cache != nil {
		if ttl, err := mlc.l3Cache.TTL(ctx, key); err == nil {
			return ttl, nil
		}
	}

	return 0, fmt.Errorf("key not found")
}

// Stats returns statistics for all cache levels
func (mlc *MultiLevelCache) Stats() map[CacheLevel]CacheStats {
	mlc.mutex.RLock()
	defer mlc.mutex.RUnlock()

	stats := make(map[CacheLevel]CacheStats)

	stats[L1Cache] = mlc.l1Cache.Stats()

	if mlc.l2Cache != nil {
		stats[L2Cache] = mlc.l2Cache.Stats()
	}

	if mlc.l3Cache != nil {
		stats[L3Cache] = mlc.l3Cache.Stats()
	}

	return stats
}

// CacheManager manages application-wide caching
type CacheManager struct {
	cache    *MultiLevelCache
	config   *CacheConfig
	patterns map[string]*CachePattern
	mutex    sync.RWMutex
}

// CachePattern defines caching behavior for specific patterns
type CachePattern struct {
	Pattern string
	TTL     time.Duration
	Levels  []CacheLevel
}

// NewCacheManager creates a new cache manager
func NewCacheManager(config *CacheConfig) *CacheManager {
	return &CacheManager{
		cache:    NewMultiLevelCache(config),
		config:   config,
		patterns: make(map[string]*CachePattern),
	}
}

// SetPattern sets a caching pattern
func (cm *CacheManager) SetPattern(pattern string, ttl time.Duration, levels []CacheLevel) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.patterns[pattern] = &CachePattern{
		Pattern: pattern,
		TTL:     ttl,
		Levels:  levels,
	}
}

// GetWithPattern retrieves a value using pattern-based caching
func (cm *CacheManager) GetWithPattern(ctx context.Context, key string) (interface{}, bool, error) {
	return cm.cache.Get(ctx, key)
}

// SetWithPattern stores a value using pattern-based caching
func (cm *CacheManager) SetWithPattern(ctx context.Context, key string, value interface{}) error {
	ttl := cm.config.DefaultTTL

	// Find matching pattern
	cm.mutex.RLock()
	for pattern, cachePattern := range cm.patterns {
		if matchesPattern(key, pattern) {
			ttl = cachePattern.TTL
			break
		}
	}
	cm.mutex.RUnlock()

	return cm.cache.Set(ctx, key, value, ttl)
}

// matchesPattern checks if a key matches a pattern
func matchesPattern(key, pattern string) bool {
	// Simple pattern matching - could be enhanced with regex
	return key == pattern || (len(pattern) > 0 && pattern[len(pattern)-1] == '*' &&
		len(key) >= len(pattern)-1 && key[:len(pattern)-1] == pattern[:len(pattern)-1])
}

// Global cache manager instance
var defaultCacheManager *CacheManager
var once sync.Once

// GetCacheManager returns the global cache manager
func GetCacheManager() *CacheManager {
	once.Do(func() {
		defaultCacheManager = NewCacheManager(DefaultCacheConfig())
	})
	return defaultCacheManager
}

// Convenience functions

// Get retrieves a value from the global cache
func Get(ctx context.Context, key string) (interface{}, bool, error) {
	return GetCacheManager().GetWithPattern(ctx, key)
}

// Set stores a value in the global cache
func Set(ctx context.Context, key string, value interface{}) error {
	return GetCacheManager().SetWithPattern(ctx, key, value)
}

// Delete removes a value from the global cache
func Delete(ctx context.Context, key string) error {
	return GetCacheManager().cache.Delete(ctx, key)
}

// Clear clears the global cache
func Clear(ctx context.Context) error {
	return GetCacheManager().cache.Clear(ctx)
}

// SetWithTTL stores a value with custom TTL
func SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return GetCacheManager().cache.Set(ctx, key, value, ttl)
}

// GetStats returns cache statistics
func GetStats() map[CacheLevel]CacheStats {
	return GetCacheManager().cache.Stats()
}
