package middleware

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"kisanlink-ecom/entities/models/common"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	// RequestsPerSecond is the rate limit (requests per second)
	RequestsPerSecond int
	// BurstSize is the maximum burst of requests allowed
	BurstSize int
	// KeyGenerator generates keys for rate limiting (e.g., by IP, user ID)
	KeyGenerator func(*gin.Context) string
	// SkipPaths defines paths that skip rate limiting
	SkipPaths []string
	// WindowSize is the time window for rate limiting (for sliding window)
	WindowSize time.Duration
	// OnRateLimitExceeded is called when rate limit is exceeded
	OnRateLimitExceeded func(*gin.Context, string)
	// Headers controls whether to add rate limit headers
	AddHeaders bool
}

// DefaultRateLimiterConfig returns default rate limiter configuration
func DefaultRateLimiterConfig() *RateLimiterConfig {
	return &RateLimiterConfig{
		RequestsPerSecond: 100,
		BurstSize:         10,
		KeyGenerator:      IPBasedKeyGenerator,
		SkipPaths:         []string{"/health", "/metrics"},
		WindowSize:        time.Minute,
		AddHeaders:        true,
		OnRateLimitExceeded: func(c *gin.Context, key string) {
			// Default no-op handler
		},
	}
}

// RateLimiterType defines the type of rate limiter
type RateLimiterType int

const (
	// TokenBucket uses token bucket algorithm
	TokenBucket RateLimiterType = iota
	// SlidingWindow uses sliding window algorithm
	SlidingWindow
	// FixedWindow uses fixed window algorithm
	FixedWindow
)

// RateLimiter interface for different rate limiting algorithms
type RateLimiter interface {
	Allow(key string) bool
	GetStats(key string) RateLimitStats
	Reset(key string)
	Cleanup()
}

// RateLimitStats holds statistics for rate limiting
type RateLimitStats struct {
	Key           string    `json:"key"`
	RequestCount  int64     `json:"request_count"`
	AllowedCount  int64     `json:"allowed_count"`
	BlockedCount  int64     `json:"blocked_count"`
	LastRequest   time.Time `json:"last_request"`
	WindowStart   time.Time `json:"window_start,omitempty"`
	RemainingRate int       `json:"remaining_rate"`
	ResetTime     time.Time `json:"reset_time,omitempty"`
}

// TokenBucketLimiter implements rate limiting using token bucket algorithm
type TokenBucketLimiter struct {
	limiters map[string]*rate.Limiter
	mutex    sync.RWMutex
	config   *RateLimiterConfig
	stats    map[string]*RateLimitStats
}

// NewTokenBucketLimiter creates a new token bucket rate limiter
func NewTokenBucketLimiter(config *RateLimiterConfig) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		limiters: make(map[string]*rate.Limiter),
		config:   config,
		stats:    make(map[string]*RateLimitStats),
	}
}

// Allow checks if a request is allowed for the given key
func (tbl *TokenBucketLimiter) Allow(key string) bool {
	limiter := tbl.getLimiter(key)

	tbl.mutex.Lock()
	stats := tbl.getStats(key)
	stats.RequestCount++
	stats.LastRequest = time.Now()

	allowed := limiter.Allow()
	if allowed {
		stats.AllowedCount++
	} else {
		stats.BlockedCount++
	}

	// Calculate remaining rate
	reservation := limiter.Reserve()
	if reservation.OK() {
		stats.RemainingRate = int(float64(tbl.config.BurstSize) - reservation.Delay().Seconds()*float64(tbl.config.RequestsPerSecond))
		if stats.RemainingRate < 0 {
			stats.RemainingRate = 0
		}
		reservation.Cancel()
	}

	tbl.mutex.Unlock()

	return allowed
}

// GetStats returns statistics for the given key
func (tbl *TokenBucketLimiter) GetStats(key string) RateLimitStats {
	tbl.mutex.RLock()
	defer tbl.mutex.RUnlock()

	if stats, exists := tbl.stats[key]; exists {
		return *stats
	}

	return RateLimitStats{Key: key}
}

// Reset resets the rate limiter for the given key
func (tbl *TokenBucketLimiter) Reset(key string) {
	tbl.mutex.Lock()
	defer tbl.mutex.Unlock()

	delete(tbl.limiters, key)
	delete(tbl.stats, key)
}

// Cleanup removes old entries
func (tbl *TokenBucketLimiter) Cleanup() {
	tbl.mutex.Lock()
	defer tbl.mutex.Unlock()

	cutoff := time.Now().Add(-time.Hour)
	for key, stats := range tbl.stats {
		if stats.LastRequest.Before(cutoff) {
			delete(tbl.limiters, key)
			delete(tbl.stats, key)
		}
	}
}

// getLimiter gets or creates a rate limiter for the given key
func (tbl *TokenBucketLimiter) getLimiter(key string) *rate.Limiter {
	tbl.mutex.RLock()
	limiter, exists := tbl.limiters[key]
	tbl.mutex.RUnlock()

	if exists {
		return limiter
	}

	tbl.mutex.Lock()
	defer tbl.mutex.Unlock()

	// Double-check after acquiring write lock
	if limiter, exists := tbl.limiters[key]; exists {
		return limiter
	}

	limiter = rate.NewLimiter(rate.Limit(tbl.config.RequestsPerSecond), tbl.config.BurstSize)
	tbl.limiters[key] = limiter

	return limiter
}

// getStats gets or creates stats for the given key
func (tbl *TokenBucketLimiter) getStats(key string) *RateLimitStats {
	if stats, exists := tbl.stats[key]; exists {
		return stats
	}

	stats := &RateLimitStats{
		Key:         key,
		WindowStart: time.Now(),
	}
	tbl.stats[key] = stats

	return stats
}

// SlidingWindowLimiter implements rate limiting using sliding window algorithm
type SlidingWindowLimiter struct {
	windows map[string]*SlidingWindowData
	mutex   sync.RWMutex
	config  *RateLimiterConfig
	stats   map[string]*RateLimitStats
}

// SlidingWindowData represents a sliding window for rate limiting
type SlidingWindowData struct {
	requests    []time.Time
	maxRequests int
	windowSize  time.Duration
}

// NewSlidingWindowLimiter creates a new sliding window rate limiter
func NewSlidingWindowLimiter(config *RateLimiterConfig) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		windows: make(map[string]*SlidingWindowData),
		config:  config,
		stats:   make(map[string]*RateLimitStats),
	}
}

// Allow checks if a request is allowed for the given key
func (swl *SlidingWindowLimiter) Allow(key string) bool {
	swl.mutex.Lock()
	defer swl.mutex.Unlock()

	window := swl.getWindow(key)
	stats := swl.getStats(key)
	now := time.Now()

	// Remove old requests outside the window
	cutoff := now.Add(-swl.config.WindowSize)
	newRequests := make([]time.Time, 0, len(window.requests))
	for _, requestTime := range window.requests {
		if requestTime.After(cutoff) {
			newRequests = append(newRequests, requestTime)
		}
	}
	window.requests = newRequests

	stats.RequestCount++
	stats.LastRequest = now

	// Check if we can allow this request
	if len(window.requests) < window.maxRequests {
		window.requests = append(window.requests, now)
		stats.AllowedCount++
		stats.RemainingRate = window.maxRequests - len(window.requests)
		return true
	}

	stats.BlockedCount++
	stats.RemainingRate = 0
	return false
}

// GetStats returns statistics for the given key
func (swl *SlidingWindowLimiter) GetStats(key string) RateLimitStats {
	swl.mutex.RLock()
	defer swl.mutex.RUnlock()

	if stats, exists := swl.stats[key]; exists {
		return *stats
	}

	return RateLimitStats{Key: key}
}

// Reset resets the rate limiter for the given key
func (swl *SlidingWindowLimiter) Reset(key string) {
	swl.mutex.Lock()
	defer swl.mutex.Unlock()

	delete(swl.windows, key)
	delete(swl.stats, key)
}

// Cleanup removes old entries
func (swl *SlidingWindowLimiter) Cleanup() {
	swl.mutex.Lock()
	defer swl.mutex.Unlock()

	cutoff := time.Now().Add(-time.Hour)
	for key, stats := range swl.stats {
		if stats.LastRequest.Before(cutoff) {
			delete(swl.windows, key)
			delete(swl.stats, key)
		}
	}
}

// getWindow gets or creates a sliding window for the given key
func (swl *SlidingWindowLimiter) getWindow(key string) *SlidingWindowData {
	if window, exists := swl.windows[key]; exists {
		return window
	}

	maxRequests := swl.config.RequestsPerSecond * int(swl.config.WindowSize.Seconds())
	window := &SlidingWindowData{
		requests:    make([]time.Time, 0),
		maxRequests: maxRequests,
		windowSize:  swl.config.WindowSize,
	}
	swl.windows[key] = window

	return window
}

// getStats gets or creates stats for the given key
func (swl *SlidingWindowLimiter) getStats(key string) *RateLimitStats {
	if stats, exists := swl.stats[key]; exists {
		return stats
	}

	stats := &RateLimitStats{
		Key:         key,
		WindowStart: time.Now(),
	}
	swl.stats[key] = stats

	return stats
}

// RateLimiterMiddleware provides rate limiting functionality
type RateLimiterMiddleware struct {
	limiter RateLimiter
	config  *RateLimiterConfig
	ticker  *time.Ticker
	done    chan bool
}

// NewRateLimiterMiddleware creates a new rate limiter middleware
func NewRateLimiterMiddleware(limiterType RateLimiterType, config *RateLimiterConfig) *RateLimiterMiddleware {
	if config == nil {
		config = DefaultRateLimiterConfig()
	}

	var limiter RateLimiter
	switch limiterType {
	case TokenBucket:
		limiter = NewTokenBucketLimiter(config)
	case SlidingWindow:
		limiter = NewSlidingWindowLimiter(config)
	default:
		limiter = NewTokenBucketLimiter(config)
	}

	middleware := &RateLimiterMiddleware{
		limiter: limiter,
		config:  config,
		done:    make(chan bool),
	}

	// Start cleanup routine
	middleware.startCleanup()

	return middleware
}

// Middleware returns the rate limiting middleware function
func (rlm *RateLimiterMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip rate limiting for configured paths
		for _, skipPath := range rlm.config.SkipPaths {
			if c.Request.URL.Path == skipPath {
				c.Next()
				return
			}
		}

		// Generate key for rate limiting
		key := rlm.config.KeyGenerator(c)

		// Check rate limit
		allowed := rlm.limiter.Allow(key)

		// Add rate limit headers if enabled
		if rlm.config.AddHeaders {
			rlm.addRateLimitHeaders(c, key)
		}

		if !allowed {
			// Call the rate limit exceeded handler
			rlm.config.OnRateLimitExceeded(c, key)

			// Return rate limit exceeded response
			c.JSON(http.StatusTooManyRequests, common.APIResponse{
				Success: false,
				Error: &common.APIError{
					Code:    "RATE_LIMIT_EXCEEDED",
					Message: "Rate limit exceeded. Please try again later.",
					Details: fmt.Sprintf("Key: %s", key),
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// addRateLimitHeaders adds rate limit headers to the response
func (rlm *RateLimiterMiddleware) addRateLimitHeaders(c *gin.Context, key string) {
	stats := rlm.limiter.GetStats(key)

	c.Header("X-RateLimit-Limit", strconv.Itoa(rlm.config.RequestsPerSecond))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(stats.RemainingRate))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(stats.ResetTime.Unix(), 10))
	c.Header("X-RateLimit-Window", rlm.config.WindowSize.String())
}

// GetStats returns statistics for all rate limiters
func (rlm *RateLimiterMiddleware) GetStats() map[string]RateLimitStats {
	// This would need to be implemented based on the specific limiter type
	return make(map[string]RateLimitStats)
}

// Reset resets the rate limiter for a given key
func (rlm *RateLimiterMiddleware) Reset(key string) {
	rlm.limiter.Reset(key)
}

// startCleanup starts the cleanup routine
func (rlm *RateLimiterMiddleware) startCleanup() {
	rlm.ticker = time.NewTicker(10 * time.Minute)

	go func() {
		for {
			select {
			case <-rlm.ticker.C:
				rlm.limiter.Cleanup()
			case <-rlm.done:
				rlm.ticker.Stop()
				return
			}
		}
	}()
}

// Stop stops the cleanup routine
func (rlm *RateLimiterMiddleware) Stop() {
	close(rlm.done)
}

// Key generator functions

// IPBasedKeyGenerator generates keys based on client IP
func IPBasedKeyGenerator(c *gin.Context) string {
	return "ip:" + c.ClientIP()
}

// UserBasedKeyGenerator generates keys based on user ID
func UserBasedKeyGenerator(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return IPBasedKeyGenerator(c)
	}
	return "user:" + userID.(string)
}

// APIKeyBasedKeyGenerator generates keys based on API key
func APIKeyBasedKeyGenerator(c *gin.Context) string {
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		return IPBasedKeyGenerator(c)
	}
	return "api_key:" + apiKey
}

// CompositeKeyGenerator generates composite keys
func CompositeKeyGenerator(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if exists {
		return fmt.Sprintf("user:%s:ip:%s", userID.(string), c.ClientIP())
	}
	return IPBasedKeyGenerator(c)
}

// EndpointBasedKeyGenerator generates keys based on endpoint and IP
func EndpointBasedKeyGenerator(c *gin.Context) string {
	return fmt.Sprintf("endpoint:%s:ip:%s", c.Request.URL.Path, c.ClientIP())
}

// RateLimitManagerMiddleware provides endpoints for managing rate limits
type RateLimitManagerMiddleware struct {
	rateLimiter *RateLimiterMiddleware
}

// NewRateLimitManagerMiddleware creates a new rate limit manager middleware
func NewRateLimitManagerMiddleware(rateLimiter *RateLimiterMiddleware) *RateLimitManagerMiddleware {
	return &RateLimitManagerMiddleware{
		rateLimiter: rateLimiter,
	}
}

// GetRateLimitStats returns rate limit statistics
func (rlmm *RateLimitManagerMiddleware) GetRateLimitStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := rlmm.rateLimiter.GetStats()

		c.JSON(http.StatusOK, common.APIResponse{
			Success: true,
			Data: gin.H{
				"stats":     stats,
				"timestamp": time.Now().UTC(),
			},
		})
	}
}

// ResetRateLimit resets rate limit for a specific key
func (rlmm *RateLimitManagerMiddleware) ResetRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Param("key")
		if key == "" {
			c.JSON(http.StatusBadRequest, common.APIResponse{
				Success: false,
				Error: &common.APIError{
					Code:    "MISSING_KEY",
					Message: "Rate limit key is required",
				},
			})
			return
		}

		rlmm.rateLimiter.Reset(key)

		c.JSON(http.StatusOK, common.APIResponse{
			Success: true,
			Data: gin.H{
				"message": "Rate limit reset successfully",
				"key":     key,
			},
		})
	}
}

// RegisterRoutes registers rate limit management routes
func (rlmm *RateLimitManagerMiddleware) RegisterRoutes(router *gin.RouterGroup) {
	rateLimit := router.Group("/rate-limit")
	{
		rateLimit.GET("/stats", rlmm.GetRateLimitStats())
		rateLimit.POST("/reset/:key", rlmm.ResetRateLimit())
	}
}

// Advanced rate limiting features

// AdaptiveRateLimiter adjusts rate limits based on system load
type AdaptiveRateLimiter struct {
	baseLimiter   RateLimiter
	loadThreshold float64
	currentLoad   float64
	config        *RateLimiterConfig
	mutex         sync.RWMutex
}

// NewAdaptiveRateLimiter creates an adaptive rate limiter
func NewAdaptiveRateLimiter(baseLimiter RateLimiter, config *RateLimiterConfig) *AdaptiveRateLimiter {
	return &AdaptiveRateLimiter{
		baseLimiter:   baseLimiter,
		loadThreshold: 0.8,
		config:        config,
	}
}

// Allow checks if a request is allowed, adjusting for system load
func (arl *AdaptiveRateLimiter) Allow(key string) bool {
	arl.mutex.RLock()
	load := arl.currentLoad
	arl.mutex.RUnlock()

	// Reduce rate limit if system is under high load
	if load > arl.loadThreshold {
		loadFactor := (1.0 - load) * 2 // More aggressive reduction under high load
		if loadFactor < 0.1 {
			loadFactor = 0.1 // Minimum 10% of normal rate
		}

		// Simple implementation: randomly reject based on load
		if rand.Float64() > loadFactor {
			return false
		}
	}

	return arl.baseLimiter.Allow(key)
}

// UpdateLoad updates the current system load
func (arl *AdaptiveRateLimiter) UpdateLoad(load float64) {
	arl.mutex.Lock()
	arl.currentLoad = load
	arl.mutex.Unlock()
}

// GetStats returns statistics
func (arl *AdaptiveRateLimiter) GetStats(key string) RateLimitStats {
	return arl.baseLimiter.GetStats(key)
}

// Reset resets the rate limiter
func (arl *AdaptiveRateLimiter) Reset(key string) {
	arl.baseLimiter.Reset(key)
}

// Cleanup performs cleanup
func (arl *AdaptiveRateLimiter) Cleanup() {
	arl.baseLimiter.Cleanup()
}

// Distributed rate limiting (for multi-instance deployments)

// DistributedRateLimiter interface for distributed rate limiting
type DistributedRateLimiter interface {
	RateLimiter
	Sync() error
}

// RedisRateLimiter implements distributed rate limiting using Redis
type RedisRateLimiter struct {
	// Redis client would be here
	config *RateLimiterConfig
	prefix string
}

// NewRedisRateLimiter creates a Redis-based rate limiter
func NewRedisRateLimiter(config *RateLimiterConfig) *RedisRateLimiter {
	return &RedisRateLimiter{
		config: config,
		prefix: "rate_limit:",
	}
}

// Allow checks if a request is allowed using Redis
func (rrl *RedisRateLimiter) Allow(key string) bool {
	// Implementation would use Redis commands for distributed rate limiting
	// This is a placeholder for the actual Redis implementation
	return true
}

// GetStats returns statistics from Redis
func (rrl *RedisRateLimiter) GetStats(key string) RateLimitStats {
	return RateLimitStats{Key: key}
}

// Reset resets the rate limiter in Redis
func (rrl *RedisRateLimiter) Reset(key string) {
	// Redis implementation
}

// Cleanup performs cleanup in Redis
func (rrl *RedisRateLimiter) Cleanup() {
	// Redis implementation
}

// Sync synchronizes with Redis
func (rrl *RedisRateLimiter) Sync() error {
	return nil
}
