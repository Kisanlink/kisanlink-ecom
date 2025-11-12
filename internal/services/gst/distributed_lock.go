package gst

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// DistributedLock provides Redis-based distributed locking for GST deduplication
type DistributedLock struct {
	client  redis.UniversalClient
	logger  *logrus.Logger
	metrics *LockMetrics
}

// LockOptions configures lock acquisition behavior
type LockOptions struct {
	Key        string
	TTL        time.Duration
	RetryDelay time.Duration
	MaxRetries int
	Owner      string // Unique identifier for lock owner
}

// Lock represents an acquired distributed lock
type Lock struct {
	key       string
	owner     string
	token     string // Unique token for this lock instance
	ttl       time.Duration
	renewStop chan struct{}
	client    redis.UniversalClient
	logger    *logrus.Logger
}

// LockMetrics tracks lock performance metrics
type LockMetrics struct {
	acquisitionDuration map[string]time.Duration
	holdDuration        map[string]time.Duration
	acquisitionFailures int64
}

var (
	// ErrLockAcquisitionTimeout indicates lock could not be acquired within retry limit
	ErrLockAcquisitionTimeout = fmt.Errorf("lock acquisition timeout")

	// ErrLockLost indicates the lock was lost during operation
	ErrLockLost = fmt.Errorf("lock lost during operation")
)

// NewDistributedLock creates a new distributed lock manager
func NewDistributedLock(client redis.UniversalClient, logger *logrus.Logger) *DistributedLock {
	return &DistributedLock{
		client: client,
		logger: logger,
		metrics: &LockMetrics{
			acquisitionDuration: make(map[string]time.Duration),
			holdDuration:        make(map[string]time.Duration),
		},
	}
}

// AcquireLock attempts to acquire a distributed lock with retry logic
func (dl *DistributedLock) AcquireLock(ctx context.Context, opts LockOptions) (*Lock, error) {
	startTime := time.Now()

	// Generate unique token for this lock instance
	token, err := generateLockToken(opts.Owner)
	if err != nil {
		return nil, fmt.Errorf("failed to generate lock token: %w", err)
	}

	// Lua script for atomic lock acquisition
	// Returns 1 if lock acquired, 0 if already held by another owner
	script := redis.NewScript(`
		if redis.call("EXISTS", KEYS[1]) == 0 then
			redis.call("SET", KEYS[1], ARGV[1], "PX", ARGV[2])
			return 1
		else
			local current = redis.call("GET", KEYS[1])
			if current == ARGV[1] then
				redis.call("PEXPIRE", KEYS[1], ARGV[2])
				return 1
			end
			return 0
		end
	`)

	// Retry with exponential backoff
	for attempt := 0; attempt <= opts.MaxRetries; attempt++ {
		result, err := script.Run(ctx, dl.client,
			[]string{opts.Key},
			token,
			opts.TTL.Milliseconds()).Result()

		if err != nil {
			dl.logger.WithError(err).WithFields(logrus.Fields{
				"key":     opts.Key,
				"attempt": attempt,
			}).Warn("Lock acquisition script error")

			if attempt >= opts.MaxRetries {
				dl.metrics.acquisitionFailures++
				return nil, fmt.Errorf("lock acquisition failed after %d attempts: %w", opts.MaxRetries+1, err)
			}

			// Exponential backoff
			delay := time.Duration(math.Pow(2, float64(attempt))) * opts.RetryDelay
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				continue
			}
		}

		// Check if lock was acquired
		if result.(int64) == 1 {
			lock := &Lock{
				key:       opts.Key,
				owner:     opts.Owner,
				token:     token,
				ttl:       opts.TTL,
				renewStop: make(chan struct{}),
				client:    dl.client,
				logger:    dl.logger,
			}

			// Start automatic renewal goroutine
			lock.startRenewal(ctx)

			// Record metrics
			duration := time.Since(startTime)
			dl.metrics.acquisitionDuration[opts.Key] = duration
			dl.logger.WithFields(logrus.Fields{
				"key":      opts.Key,
				"duration": duration,
				"attempt":  attempt,
			}).Debug("Lock acquired successfully")

			return lock, nil
		}

		// Lock is held by another process, retry
		if attempt < opts.MaxRetries {
			delay := time.Duration(math.Pow(2, float64(attempt))) * opts.RetryDelay
			if delay > 30*time.Second {
				delay = 30 * time.Second // Cap maximum backoff
			}

			dl.logger.WithFields(logrus.Fields{
				"key":     opts.Key,
				"attempt": attempt,
				"delay":   delay,
			}).Debug("Lock held by another process, retrying")

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				continue
			}
		}
	}

	dl.metrics.acquisitionFailures++
	return nil, ErrLockAcquisitionTimeout
}

// startRenewal starts a background goroutine to automatically renew the lock
func (l *Lock) startRenewal(ctx context.Context) {
	renewInterval := l.ttl / 3 // Renew at 1/3 of TTL

	go func() {
		ticker := time.NewTicker(renewInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-l.renewStop:
				return
			case <-ticker.C:
				if err := l.renew(ctx); err != nil {
					l.logger.WithError(err).WithField("key", l.key).Error("Failed to renew lock")
					return
				}
			}
		}
	}()
}

// renew extends the TTL of the lock if we still own it
func (l *Lock) renew(ctx context.Context) error {
	script := redis.NewScript(`
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("PEXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`)

	result, err := script.Run(ctx, l.client,
		[]string{l.key},
		l.token,
		l.ttl.Milliseconds()).Result()

	if err != nil {
		return fmt.Errorf("lock renewal failed: %w", err)
	}

	if result.(int64) == 0 {
		return ErrLockLost
	}

	l.logger.WithField("key", l.key).Debug("Lock renewed successfully")
	return nil
}

// Release releases the lock if we still own it
func (l *Lock) Release(ctx context.Context) error {
	// Stop renewal goroutine
	if l.renewStop != nil {
		close(l.renewStop)
	}

	// Lua script for atomic lock release
	// Only delete if we own the lock
	script := redis.NewScript(`
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`)

	result, err := script.Run(ctx, l.client,
		[]string{l.key},
		l.token).Result()

	if err != nil {
		return fmt.Errorf("lock release failed: %w", err)
	}

	if result.(int64) == 0 {
		l.logger.WithField("key", l.key).Warn("Lock was already released or taken by another process")
		return nil
	}

	l.logger.WithField("key", l.key).Debug("Lock released successfully")
	return nil
}

// generateLockToken generates a unique token for lock ownership
func generateLockToken(owner string) (string, error) {
	// Generate 16 bytes of randomness
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	// Combine owner and random bytes for uniqueness
	token := fmt.Sprintf("%s:%s:%d",
		owner,
		hex.EncodeToString(randomBytes),
		time.Now().UnixNano())

	return token, nil
}

// GetMetrics returns current lock metrics
func (dl *DistributedLock) GetMetrics() *LockMetrics {
	return dl.metrics
}
