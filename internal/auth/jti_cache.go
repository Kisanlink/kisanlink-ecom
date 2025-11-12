package auth

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// JTICache manages JWT Token IDs to prevent replay attacks
type JTICache interface {
	// Check validates a JTI hasn't been seen before and stores it
	Check(ctx context.Context, jti string, expiresAt time.Time) error

	// Cleanup removes expired JTIs
	Cleanup(ctx context.Context) error
}

// inMemoryJTICache is an in-memory implementation of JTICache
// In production, this should use Redis for distributed systems
type inMemoryJTICache struct {
	cache map[string]time.Time
	mu    sync.RWMutex
}

// NewInMemoryJTICache creates a new in-memory JTI cache
func NewInMemoryJTICache() JTICache {
	cache := &inMemoryJTICache{
		cache: make(map[string]time.Time),
	}

	// Start cleanup goroutine
	go cache.periodicCleanup()

	return cache
}

// Check validates a JTI hasn't been seen before and stores it
func (c *inMemoryJTICache) Check(ctx context.Context, jti string, expiresAt time.Time) error {
	if jti == "" {
		return fmt.Errorf("empty JTI")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if JTI has been seen before
	if _, exists := c.cache[jti]; exists {
		return fmt.Errorf("token replay detected: JTI %s already used", jti)
	}

	// Store JTI with expiration
	c.cache[jti] = expiresAt

	return nil
}

// Cleanup removes expired JTIs
func (c *inMemoryJTICache) Cleanup(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for jti, expiresAt := range c.cache {
		if now.After(expiresAt) {
			delete(c.cache, jti)
		}
	}

	return nil
}

// periodicCleanup runs cleanup every 5 minutes
func (c *inMemoryJTICache) periodicCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		_ = c.Cleanup(context.Background())
	}
}

// Size returns the current cache size (for testing/monitoring)
func (c *inMemoryJTICache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}
