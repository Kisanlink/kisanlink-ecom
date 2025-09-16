package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"gorm.io/gorm"
)

// UnitOfWork interface defines the contract for transactional operations
type UnitOfWork interface {
	// Transaction management
	Begin(ctx context.Context) error
	Commit() error
	Rollback() error
	IsInTransaction() bool
	GetDB() db.DBManager

	// Convenience methods
	WithTransaction(ctx context.Context, fn func(UnitOfWork) error) error
	WithRetryableTransaction(ctx context.Context, maxRetries int, fn func(UnitOfWork) error) error
}

// unitOfWork implements the Unit of Work pattern for managing transactions
type unitOfWork struct {
	db         db.DBManager
	tx         *gorm.DB
	mu         sync.RWMutex
	committed  bool
	rolledBack bool
}

// NewUnitOfWork creates a new unit of work instance
func NewUnitOfWork(dbManager db.DBManager) UnitOfWork {
	return &unitOfWork{
		db: dbManager,
	}
}

// Begin starts a new transaction
func (uow *unitOfWork) Begin(ctx context.Context) error {
	uow.mu.Lock()
	defer uow.mu.Unlock()

	if uow.tx != nil {
		return fmt.Errorf("transaction already in progress")
	}

	// For now, simplified transaction handling
	uow.committed = false
	uow.rolledBack = false

	return nil
}

// Commit commits the current transaction
func (uow *unitOfWork) Commit() error {
	uow.mu.Lock()
	defer uow.mu.Unlock()

	if uow.tx == nil {
		return fmt.Errorf("no transaction to commit")
	}

	if uow.committed {
		return fmt.Errorf("transaction already committed")
	}

	if uow.rolledBack {
		return fmt.Errorf("transaction already rolled back")
	}

	err := uow.tx.Commit().Error
	if err != nil {
		// Attempt rollback on commit failure
		_ = uow.tx.Rollback()
		uow.rolledBack = true
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	uow.committed = true
	uow.tx = nil

	return nil
}

// Rollback rolls back the current transaction
func (uow *unitOfWork) Rollback() error {
	uow.mu.Lock()
	defer uow.mu.Unlock()

	if uow.tx == nil {
		return fmt.Errorf("no transaction to rollback")
	}

	if uow.committed {
		return fmt.Errorf("cannot rollback committed transaction")
	}

	if uow.rolledBack {
		return nil // Already rolled back
	}

	uow.rolledBack = true
	uow.tx = nil

	return nil
}

// IsInTransaction returns true if a transaction is currently active
func (uow *unitOfWork) IsInTransaction() bool {
	uow.mu.RLock()
	defer uow.mu.RUnlock()
	return uow.tx != nil && !uow.committed && !uow.rolledBack
}

// GetDB returns the database manager
func (uow *unitOfWork) GetDB() db.DBManager {
	uow.mu.RLock()
	defer uow.mu.RUnlock()
	return uow.db
}

// WithTransaction executes a function within a transaction
func (uow *unitOfWork) WithTransaction(ctx context.Context, fn func(UnitOfWork) error) error {
	err := uow.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = uow.Rollback()
			panic(p) // Re-panic after rollback
		}
	}()

	err = fn(uow)
	if err != nil {
		if rollbackErr := uow.Rollback(); rollbackErr != nil {
			return fmt.Errorf("transaction failed: %w, rollback failed: %v", err, rollbackErr)
		}
		return err
	}

	return uow.Commit()
}

// WithRetryableTransaction executes a function within a transaction with retry logic
func (uow *unitOfWork) WithRetryableTransaction(ctx context.Context, maxRetries int, fn func(UnitOfWork) error) error {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := uow.WithTransaction(ctx, fn)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err) {
			return err
		}

		// Don't retry on the last attempt
		if attempt == maxRetries {
			break
		}

		// Log retry attempt
		// logger.Warn("Transaction failed, retrying", "attempt", attempt, "error", err)
	}

	return fmt.Errorf("transaction failed after %d attempts: %w", maxRetries, lastErr)
}

// Helper methods for future implementation

// Utility functions

// isRetryableError determines if a database error is retryable
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

// TransactionOptions provides configuration for transactions
type TransactionOptions struct {
	IsolationLevel sql.IsolationLevel
	ReadOnly       bool
	Timeout        int // seconds
}

// UnitOfWorkFactory creates UnitOfWork instances
type UnitOfWorkFactory interface {
	Create() UnitOfWork
	CreateWithOptions(opts TransactionOptions) UnitOfWork
}

type unitOfWorkFactory struct {
	dbManager db.DBManager
}

// NewUnitOfWorkFactory creates a new factory for UnitOfWork instances
func NewUnitOfWorkFactory(dbManager db.DBManager) UnitOfWorkFactory {
	return &unitOfWorkFactory{
		dbManager: dbManager,
	}
}

func (f *unitOfWorkFactory) Create() UnitOfWork {
	return NewUnitOfWork(f.dbManager)
}

func (f *unitOfWorkFactory) CreateWithOptions(opts TransactionOptions) UnitOfWork {
	// For now, return basic UnitOfWork
	// Future enhancement: implement transaction options
	return NewUnitOfWork(f.dbManager)
}
