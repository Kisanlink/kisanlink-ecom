package gst

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRedis creates a test Redis client (using miniredis for testing)
func setupTestRedis(t *testing.T) redis.UniversalClient {
	// For testing, we'll use a real Redis instance or mock
	// In production tests, use miniredis or testcontainers
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // Use a dedicated test database
	})

	// Clear test database
	ctx := context.Background()
	err := client.FlushDB(ctx).Err()
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}

	return client
}

func TestDistributedLock_AcquireLock_Success(t *testing.T) {
	client := setupTestRedis(t)
	defer func() { _ = client.Close() }()

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	lockMgr := NewDistributedLock(client, logger)
	ctx := context.Background()

	opts := LockOptions{
		Key:        "test:lock:success",
		TTL:        5 * time.Second,
		RetryDelay: 100 * time.Millisecond,
		MaxRetries: 3,
		Owner:      "test-owner-1",
	}

	// Acquire lock
	lock, err := lockMgr.AcquireLock(ctx, opts)
	require.NoError(t, err, "Failed to acquire lock")
	require.NotNil(t, lock)
	defer func() { _ = lock.Release(ctx) }()

	// Verify lock exists in Redis
	val, err := client.Get(ctx, opts.Key).Result()
	require.NoError(t, err)
	assert.Contains(t, val, opts.Owner)
}

func TestDistributedLock_AcquireLock_Concurrent(t *testing.T) {
	client := setupTestRedis(t)
	defer func() { _ = client.Close() }()

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	lockMgr := NewDistributedLock(client, logger)
	ctx := context.Background()

	var wg sync.WaitGroup
	successCount := atomic.Int32{}
	lockKey := "test:lock:concurrent"

	// Spawn 100 goroutines trying to acquire the same lock
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			opts := LockOptions{
				Key:        lockKey,
				TTL:        2 * time.Second,
				RetryDelay: 10 * time.Millisecond,
				MaxRetries: 5,
				Owner:      fmt.Sprintf("owner-%d", id),
			}

			lock, err := lockMgr.AcquireLock(ctx, opts)
			if err == nil {
				successCount.Add(1)
				time.Sleep(50 * time.Millisecond)
				_ = lock.Release(ctx)
			}
		}(i)
	}

	wg.Wait()

	// At least one should succeed
	assert.Greater(t, successCount.Load(), int32(0), "At least one goroutine should acquire the lock")

	// After all released, lock should not exist
	time.Sleep(100 * time.Millisecond)
	exists, err := client.Exists(ctx, lockKey).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), exists, "Lock should be released")
}

func TestDistributedLock_AcquireLock_Timeout(t *testing.T) {
	client := setupTestRedis(t)
	defer func() { _ = client.Close() }()

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	lockMgr := NewDistributedLock(client, logger)
	ctx := context.Background()

	lockKey := "test:lock:timeout"

	// First process acquires lock
	opts1 := LockOptions{
		Key:        lockKey,
		TTL:        5 * time.Second,
		RetryDelay: 50 * time.Millisecond,
		MaxRetries: 3,
		Owner:      "owner-1",
	}

	lock1, err := lockMgr.AcquireLock(ctx, opts1)
	require.NoError(t, err)
	defer func() { _ = lock1.Release(ctx) }()

	// Second process tries to acquire same lock with limited retries
	opts2 := LockOptions{
		Key:        lockKey,
		TTL:        5 * time.Second,
		RetryDelay: 50 * time.Millisecond,
		MaxRetries: 2,
		Owner:      "owner-2",
	}

	lock2, err := lockMgr.AcquireLock(ctx, opts2)
	assert.Error(t, err, "Second lock should fail")
	assert.Nil(t, lock2)
	assert.ErrorIs(t, err, ErrLockAcquisitionTimeout)
}

func TestDistributedLock_Renewal(t *testing.T) {
	client := setupTestRedis(t)
	defer func() { _ = client.Close() }()

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	lockMgr := NewDistributedLock(client, logger)
	ctx := context.Background()

	opts := LockOptions{
		Key:        "test:lock:renewal",
		TTL:        1 * time.Second,
		RetryDelay: 50 * time.Millisecond,
		MaxRetries: 3,
		Owner:      "owner-renewal",
	}

	lock, err := lockMgr.AcquireLock(ctx, opts)
	require.NoError(t, err)
	defer func() { _ = lock.Release(ctx) }()

	// Wait longer than TTL - lock should still exist due to renewal
	time.Sleep(2 * time.Second)

	// Lock should still exist
	exists, err := client.Exists(ctx, opts.Key).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), exists, "Lock should still exist due to renewal")

	// Verify TTL is being refreshed
	ttl, err := client.TTL(ctx, opts.Key).Result()
	require.NoError(t, err)
	assert.Greater(t, ttl, time.Duration(0), "Lock should have valid TTL")
	assert.LessOrEqual(t, ttl, opts.TTL, "TTL should not exceed configured TTL")
}

func TestDistributedLock_Release(t *testing.T) {
	client := setupTestRedis(t)
	defer func() { _ = client.Close() }()

	logger := logrus.New()
	lockMgr := NewDistributedLock(client, logger)
	ctx := context.Background()

	opts := LockOptions{
		Key:        "test:lock:release",
		TTL:        5 * time.Second,
		RetryDelay: 50 * time.Millisecond,
		MaxRetries: 3,
		Owner:      "owner-release",
	}

	lock, err := lockMgr.AcquireLock(ctx, opts)
	require.NoError(t, err)

	// Release lock
	err = lock.Release(ctx)
	require.NoError(t, err)

	// Verify lock no longer exists
	exists, err := client.Exists(ctx, opts.Key).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), exists, "Lock should be released")

	// Another process should be able to acquire it
	lock2, err := lockMgr.AcquireLock(ctx, opts)
	require.NoError(t, err)
	require.NotNil(t, lock2)
	_ = lock2.Release(ctx)
}

func TestDistributedLock_ContextCancellation(t *testing.T) {
	client := setupTestRedis(t)
	defer func() { _ = client.Close() }()

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	lockMgr := NewDistributedLock(client, logger)

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	lockKey := "test:lock:cancel"

	// First process holds the lock
	opts1 := LockOptions{
		Key:        lockKey,
		TTL:        10 * time.Second,
		RetryDelay: 100 * time.Millisecond,
		MaxRetries: 20,
		Owner:      "owner-1",
	}

	lock1, err := lockMgr.AcquireLock(context.Background(), opts1)
	require.NoError(t, err)
	defer func() { _ = lock1.Release(context.Background()) }()

	// Second process tries to acquire with cancellable context
	opts2 := LockOptions{
		Key:        lockKey,
		TTL:        10 * time.Second,
		RetryDelay: 100 * time.Millisecond,
		MaxRetries: 50,
		Owner:      "owner-2",
	}

	// Cancel context after short delay
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	lock2, err := lockMgr.AcquireLock(ctx, opts2)
	assert.Error(t, err)
	assert.Nil(t, lock2)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestDistributedLock_MultipleLocksOrdered(t *testing.T) {
	client := setupTestRedis(t)
	defer func() { _ = client.Close() }()

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	lockMgr := NewDistributedLock(client, logger)
	ctx := context.Background()

	// Acquire multiple locks in sorted order to prevent deadlock
	lockKeys := []string{
		"test:lock:multi:a",
		"test:lock:multi:b",
		"test:lock:multi:c",
	}

	locks := make([]*Lock, 0, len(lockKeys))

	for _, key := range lockKeys {
		opts := LockOptions{
			Key:        key,
			TTL:        5 * time.Second,
			RetryDelay: 50 * time.Millisecond,
			MaxRetries: 5,
			Owner:      "multi-owner",
		}

		lock, err := lockMgr.AcquireLock(ctx, opts)
		require.NoError(t, err, "Failed to acquire lock %s", key)
		locks = append(locks, lock)
	}

	// Release in reverse order
	for i := len(locks) - 1; i >= 0; i-- {
		err := locks[i].Release(ctx)
		require.NoError(t, err)
	}

	// Verify all locks released
	for _, key := range lockKeys {
		exists, err := client.Exists(ctx, key).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(0), exists, "Lock %s should be released", key)
	}
}

func TestDistributedLock_Metrics(t *testing.T) {
	client := setupTestRedis(t)
	defer func() { _ = client.Close() }()

	logger := logrus.New()
	lockMgr := NewDistributedLock(client, logger)
	ctx := context.Background()

	opts := LockOptions{
		Key:        "test:lock:metrics",
		TTL:        5 * time.Second,
		RetryDelay: 50 * time.Millisecond,
		MaxRetries: 3,
		Owner:      "metrics-owner",
	}

	lock, err := lockMgr.AcquireLock(ctx, opts)
	require.NoError(t, err)
	defer func() { _ = lock.Release(ctx) }()

	// Get metrics
	metrics := lockMgr.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Contains(t, metrics.acquisitionDuration, opts.Key)
	assert.Greater(t, metrics.acquisitionDuration[opts.Key], time.Duration(0))
}

func TestGenerateLockToken_Uniqueness(t *testing.T) {
	tokens := make(map[string]bool)

	// Generate 1000 tokens and verify uniqueness
	for i := 0; i < 1000; i++ {
		token, err := generateLockToken("test-owner")
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Should not have duplicates
		assert.False(t, tokens[token], "Token should be unique")
		tokens[token] = true
	}
}
