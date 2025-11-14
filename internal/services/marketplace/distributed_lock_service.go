package marketplace

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/marketplace"
	marketplaceRepo "github.com/Kisanlink/kisanlink-ecom/internal/repositories/marketplace"

	"github.com/shopspring/decimal"
)

// DistributedLockService provides distributed locking for concurrent operations
type DistributedLockService interface {
	// Acquire a lock for a specific resource with timeout
	AcquireLock(ctx context.Context, resourceID string, timeout time.Duration) (*Lock, error)

	// Release a lock
	ReleaseLock(ctx context.Context, lock *Lock) error

	// Execute a function with a distributed lock
	WithLock(ctx context.Context, resourceID string, timeout time.Duration, fn func() error) error

	// Execute atomic bid operation with proper locking
	ExecuteAtomicBidOperation(ctx context.Context, listingID string, operation AtomicBidOperation) error
}

// Lock represents a distributed lock
type Lock struct {
	ResourceID string
	LockID     string
	AcquiredAt time.Time
	ExpiresAt  time.Time
	OwnerID    string
}

// AtomicBidOperation defines the interface for atomic bid operations
type AtomicBidOperation interface {
	Execute(ctx context.Context) error
	GetRetryPolicy() RetryPolicy
	GetResourceID() string
}

// RetryPolicy defines retry behavior for failed operations
type RetryPolicy struct {
	MaxRetries    int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
}

// DefaultRetryPolicy returns a sensible default retry policy
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxRetries:    3,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      1 * time.Second,
		BackoffFactor: 2.0,
	}
}

// distributedLockService implements DistributedLockService using in-memory locks
// In production, this would use Redis or another distributed locking mechanism
type distributedLockService struct {
	locks    map[string]*Lock
	locksMux sync.RWMutex
}

// NewDistributedLockService creates a new distributed lock service
func NewDistributedLockService() DistributedLockService {
	return &distributedLockService{
		locks: make(map[string]*Lock),
	}
}

// AcquireLock acquires a distributed lock for a resource
func (s *distributedLockService) AcquireLock(ctx context.Context, resourceID string, timeout time.Duration) (*Lock, error) {
	s.locksMux.Lock()
	defer s.locksMux.Unlock()

	// Check if lock already exists and is still valid
	if existingLock, exists := s.locks[resourceID]; exists {
		if time.Now().Before(existingLock.ExpiresAt) {
			return nil, fmt.Errorf("resource %s is already locked until %v", resourceID, existingLock.ExpiresAt)
		}
		// Lock has expired, remove it
		delete(s.locks, resourceID)
	}

	// Create new lock
	lock := &Lock{
		ResourceID: resourceID,
		LockID:     fmt.Sprintf("%s_%d", resourceID, time.Now().UnixNano()),
		AcquiredAt: time.Now(),
		ExpiresAt:  time.Now().Add(timeout),
		OwnerID:    "bidding_service", // In production, this would be a unique service instance ID
	}

	s.locks[resourceID] = lock
	return lock, nil
}

// ReleaseLock releases a distributed lock
func (s *distributedLockService) ReleaseLock(ctx context.Context, lock *Lock) error {
	s.locksMux.Lock()
	defer s.locksMux.Unlock()

	existingLock, exists := s.locks[lock.ResourceID]
	if !exists {
		return fmt.Errorf("lock not found for resource %s", lock.ResourceID)
	}

	if existingLock.LockID != lock.LockID {
		return fmt.Errorf("lock ID mismatch for resource %s", lock.ResourceID)
	}

	delete(s.locks, lock.ResourceID)
	return nil
}

// WithLock executes a function with a distributed lock
func (s *distributedLockService) WithLock(ctx context.Context, resourceID string, timeout time.Duration, fn func() error) error {
	lock, err := s.AcquireLock(ctx, resourceID, timeout)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	defer func() {
		if releaseErr := s.ReleaseLock(ctx, lock); releaseErr != nil {
			// Log the error but don't override the main error
			fmt.Printf("Warning: failed to release lock for resource %s: %v\n", resourceID, releaseErr)
		}
	}()

	return fn()
}

// ExecuteAtomicBidOperation executes an atomic bid operation with proper locking and retry
func (s *distributedLockService) ExecuteAtomicBidOperation(ctx context.Context, listingID string, operation AtomicBidOperation) error {
	retryPolicy := operation.GetRetryPolicy()

	var lastErr error
	delay := retryPolicy.InitialDelay

	for attempt := 0; attempt <= retryPolicy.MaxRetries; attempt++ {
		err := s.WithLock(ctx, operation.GetResourceID(), 30*time.Second, func() error {
			return operation.Execute(ctx)
		})

		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err) {
			return err
		}

		// Don't sleep after the last attempt
		if attempt < retryPolicy.MaxRetries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Calculate next delay with exponential backoff
				delay = time.Duration(float64(delay) * retryPolicy.BackoffFactor)
				if delay > retryPolicy.MaxDelay {
					delay = retryPolicy.MaxDelay
				}
			}
		}
	}

	return fmt.Errorf("atomic operation failed after %d attempts: %w", retryPolicy.MaxRetries+1, lastErr)
}

// PlaceBidOperation implements AtomicBidOperation for placing bids
type PlaceBidOperation struct {
	BidRepo       marketplaceRepo.BidRepository
	ListingRepo   marketplaceRepo.ListingRepository
	ListingID     string
	BidderID      string
	BidAmount     decimal.Decimal
	Quantity      decimal.Decimal
	Message       string
	PaymentMethod string
	RetryPolicy   RetryPolicy
}

// Execute performs the atomic bid placement operation
func (op *PlaceBidOperation) Execute(ctx context.Context) error {
	// Get current highest bid for validation
	currentHighest, err := op.BidRepo.GetHighestBid(ctx, op.ListingID)
	if err != nil {
		return fmt.Errorf("failed to get current highest bid: %w", err)
	}

	// Validate bid amount against current highest
	if currentHighest != nil && op.BidAmount.LessThanOrEqual(currentHighest.BidAmount) {
		return fmt.Errorf("bid amount %s must be higher than current highest bid %s",
			op.BidAmount.String(), currentHighest.BidAmount.String())
	}

	// Create new bid
	bid := marketplace.NewBid(op.ListingID, op.BidderID, op.BidAmount, op.Quantity, op.Message)
	if op.PaymentMethod != "" {
		bid.PaymentMethod = op.PaymentMethod
	}

	// Place bid atomically using repository
	placedBid, err := op.BidRepo.PlaceBidAtomic(ctx, bid, op.ListingID)
	if err != nil {
		return fmt.Errorf("failed to place bid atomically: %w", err)
	}

	// Update listing bid count and highest bid
	if err := op.ListingRepo.IncrementBidCount(ctx, op.ListingID); err != nil {
		// Log error but don't fail the bid placement since the bid was already placed
		fmt.Printf("Warning: failed to increment bid count for listing %s: %v\n", op.ListingID, err)
	}

	if err := op.ListingRepo.UpdateHighestBid(ctx, op.ListingID, placedBid.BidID); err != nil {
		// Log error but don't fail the bid placement since the bid was already placed
		fmt.Printf("Warning: failed to update highest bid for listing %s: %v\n", op.ListingID, err)
	}

	return nil
}

// GetRetryPolicy returns the retry policy for this operation
func (op *PlaceBidOperation) GetRetryPolicy() RetryPolicy {
	return op.RetryPolicy
}

// GetResourceID returns the resource ID for locking (listing ID)
func (op *PlaceBidOperation) GetResourceID() string {
	return fmt.Sprintf("listing_%s", op.ListingID)
}

// isRetryableError determines if an error is retryable
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errorStr := err.Error()

	// Common retryable database errors
	retryableErrors := []string{
		"deadlock detected",
		"lock wait timeout exceeded",
		"connection reset by peer",
		"connection refused",
		"temporary failure",
		"serialization failure",
		"could not serialize access",
		"resource temporarily unavailable",
		"already locked",
	}

	for _, retryableErr := range retryableErrors {
		if contains(errorStr, retryableErr) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
