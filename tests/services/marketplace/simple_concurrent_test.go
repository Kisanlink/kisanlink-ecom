package marketplace

import (
	"context"
	"testing"
	"time"

	"kisanlink-ecom/internal/services/marketplace"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDistributedLockServiceBasic(t *testing.T) {
	lockService := marketplace.NewDistributedLockService()

	t.Run("AcquireLock", func(t *testing.T) {
		ctx := context.Background()
		resourceID := "test_resource"
		timeout := 5 * time.Second

		lock, err := lockService.AcquireLock(ctx, resourceID, timeout)
		require.NoError(t, err)
		assert.NotNil(t, lock)
		assert.Equal(t, resourceID, lock.ResourceID)
		assert.False(t, lock.AcquiredAt.IsZero())
		assert.False(t, lock.ExpiresAt.IsZero())

		// Clean up
		err = lockService.ReleaseLock(ctx, lock)
		assert.NoError(t, err)
	})

	t.Run("WithLock", func(t *testing.T) {
		ctx := context.Background()
		resourceID := "test_resource_3"
		timeout := 5 * time.Second

		executed := false
		err := lockService.WithLock(ctx, resourceID, timeout, func() error {
			executed = true
			return nil
		})

		assert.NoError(t, err)
		assert.True(t, executed)
	})
}

func TestCacheServiceBasic(t *testing.T) {
	cacheService := marketplace.NewCacheService()

	t.Run("CacheStats", func(t *testing.T) {
		stats := cacheService.GetCacheStats()
		assert.Equal(t, int64(0), stats.HitCount)
		assert.Equal(t, int64(0), stats.MissCount)
		assert.Equal(t, 0, stats.TotalKeys)
	})

	t.Run("ClearCache", func(t *testing.T) {
		ctx := context.Background()
		err := cacheService.ClearCache(ctx)
		assert.NoError(t, err)
	})
}

func TestPerformanceMonitorBasic(t *testing.T) {
	monitor := marketplace.NewPerformanceMonitor()

	t.Run("RecordBidPlacement", func(t *testing.T) {
		ctx := context.Background()
		duration := 100 * time.Millisecond

		monitor.RecordBidPlacement(ctx, duration, true)

		report := monitor.GetPerformanceReport()
		assert.Equal(t, int64(1), report.TotalBidsPlaced)
		assert.Equal(t, int64(1), report.SuccessfulBids)
		assert.Equal(t, int64(0), report.FailedBids)
		assert.Equal(t, 1.0, report.SuccessRate)
	})

	t.Run("GetRealTimeMetrics", func(t *testing.T) {
		metrics := monitor.GetRealTimeMetrics()
		assert.False(t, metrics.CurrentTime.IsZero())
		assert.GreaterOrEqual(t, metrics.ActiveBidOperations, 0)
		assert.GreaterOrEqual(t, metrics.QueuedOperations, 0)
	})
}

func TestPlaceBidOperationBasic(t *testing.T) {
	t.Run("GetRetryPolicy", func(t *testing.T) {
		operation := &marketplace.PlaceBidOperation{
			ListingID:   "test_listing",
			BidderID:    "test_bidder",
			BidAmount:   decimal.NewFromFloat(100.0),
			Quantity:    decimal.NewFromFloat(1.0),
			RetryPolicy: marketplace.DefaultRetryPolicy(),
		}

		policy := operation.GetRetryPolicy()
		assert.Equal(t, 3, policy.MaxRetries)
		assert.Equal(t, 10*time.Millisecond, policy.InitialDelay)
		assert.Equal(t, 1*time.Second, policy.MaxDelay)
		assert.Equal(t, 2.0, policy.BackoffFactor)
	})

	t.Run("GetResourceID", func(t *testing.T) {
		operation := &marketplace.PlaceBidOperation{
			ListingID: "test_listing_123",
		}

		resourceID := operation.GetResourceID()
		assert.Equal(t, "listing_test_listing_123", resourceID)
	})
}

func TestDefaultRetryPolicy(t *testing.T) {
	policy := marketplace.DefaultRetryPolicy()
	assert.Equal(t, 3, policy.MaxRetries)
	assert.Equal(t, 10*time.Millisecond, policy.InitialDelay)
	assert.Equal(t, 1*time.Second, policy.MaxDelay)
	assert.Equal(t, 2.0, policy.BackoffFactor)
}

func TestConnectionPoolConfig(t *testing.T) {
	config := marketplace.DefaultConnectionPoolConfig()
	assert.Equal(t, 50, config.MaxConnections)
	assert.Equal(t, 5, config.MinConnections)
	assert.Equal(t, 30*time.Minute, config.ConnectionTTL)
}
