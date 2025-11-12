package saga

import (
	"context"
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"
)

// SagaExecutor orchestrates saga execution with automatic compensation
type SagaExecutor struct {
	storage SagaStorage
	logger  *zap.Logger
	metrics *SagaMetrics
}

// NewSagaExecutor creates a new saga executor
func NewSagaExecutor(storage SagaStorage, logger *zap.Logger, metrics *SagaMetrics) *SagaExecutor {
	return &SagaExecutor{
		storage: storage,
		logger:  logger,
		metrics: metrics,
	}
}

// Execute runs all saga steps with automatic compensation on failure
func (e *SagaExecutor) Execute(ctx context.Context, saga *Saga) error {
	e.logger.Info("starting saga execution",
		zap.String("saga_id", saga.ID),
		zap.String("saga_name", saga.Name),
		zap.Int("step_count", len(saga.Steps)))

	// Record saga start
	e.metrics.RecordSagaStart(saga.Name)
	startTime := time.Now()

	// Initialize saga state
	saga.SetState(SagaStateRunning)
	if err := e.storage.SaveState(ctx, saga); err != nil {
		e.logger.Error("failed to save initial saga state",
			zap.String("saga_id", saga.ID),
			zap.Error(err))
		return fmt.Errorf("failed to save initial state: %w", err)
	}

	// Execute forward transaction
	lastCompletedIndex := -1
	for i, step := range saga.Steps {
		e.logger.Info("executing saga step",
			zap.String("saga_id", saga.ID),
			zap.String("step", step.Name()),
			zap.Int("step_index", i))

		// Execute step with retry if configured
		err := e.executeStepWithRetry(ctx, step, saga.Context)

		if err != nil {
			e.logger.Error("saga step failed",
				zap.String("saga_id", saga.ID),
				zap.String("step", step.Name()),
				zap.Error(err))

			// Record step failure
			e.metrics.RecordStepFailure(saga.Name, step.Name())

			// Trigger compensation for all completed steps
			if compensateErr := e.compensate(ctx, saga, lastCompletedIndex); compensateErr != nil {
				e.logger.Error("compensation failed",
					zap.String("saga_id", saga.ID),
					zap.Error(compensateErr))

				saga.SetState(SagaStateFailed)
				_ = e.storage.SaveState(ctx, saga)

				e.metrics.RecordSagaFailure(saga.Name, time.Since(startTime))
				return fmt.Errorf("step failed and compensation failed: %w", compensateErr)
			}

			// Compensation successful
			saga.SetState(SagaStateCompensated)
			_ = e.storage.SaveState(ctx, saga)

			e.metrics.RecordSagaCompensated(saga.Name, time.Since(startTime))
			return fmt.Errorf("step %s failed, transaction compensated: %w", step.Name(), err)
		}

		// Mark step as completed
		saga.MarkCompleted(step.Name())
		lastCompletedIndex = i

		// Persist state after each successful step
		if err := e.storage.SaveState(ctx, saga); err != nil {
			e.logger.Warn("failed to save intermediate state",
				zap.String("saga_id", saga.ID),
				zap.String("step", step.Name()),
				zap.Error(err))
		}

		e.metrics.RecordStepSuccess(saga.Name, step.Name())
	}

	// All steps completed successfully
	saga.SetState(SagaStateCompleted)
	if err := e.storage.SaveState(ctx, saga); err != nil {
		e.logger.Error("failed to save final state",
			zap.String("saga_id", saga.ID),
			zap.Error(err))
		return fmt.Errorf("failed to save final state: %w", err)
	}

	e.metrics.RecordSagaSuccess(saga.Name, time.Since(startTime))

	e.logger.Info("saga execution completed successfully",
		zap.String("saga_id", saga.ID),
		zap.Duration("duration", time.Since(startTime)))

	return nil
}

// executeStepWithRetry executes a step with retry logic
func (e *SagaExecutor) executeStepWithRetry(ctx context.Context, step Step, data map[string]interface{}) error {
	maxAttempts := 1
	if step.IsRetryable() {
		maxAttempts = step.MaxRetries()
	}

	baseDelay := 100 * time.Millisecond
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Create timeout context for step execution
		stepCtx, cancel := context.WithTimeout(ctx, step.Timeout())
		defer cancel()

		// Execute step
		err := step.Execute(stepCtx, data)

		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err) {
			return err
		}

		// Don't sleep after last attempt
		if attempt < maxAttempts {
			delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt-1)))

			e.logger.Warn("step execution failed, retrying",
				zap.String("step", step.Name()),
				zap.Int("attempt", attempt),
				zap.Int("max_attempts", maxAttempts),
				zap.Duration("delay", delay),
				zap.Error(err))

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				continue
			}
		}
	}

	return fmt.Errorf("step failed after %d attempts: %w", maxAttempts, lastErr)
}

// compensate executes compensation for all completed steps in reverse order
func (e *SagaExecutor) compensate(ctx context.Context, saga *Saga, fromIndex int) error {
	e.logger.Info("starting compensation",
		zap.String("saga_id", saga.ID),
		zap.Int("from_index", fromIndex))

	saga.SetState(SagaStateCompensating)
	e.storage.SaveState(ctx, saga)

	var compensationErrors []error

	// Compensate in reverse order
	for i := fromIndex; i >= 0; i-- {
		step := saga.Steps[i]

		// Check if this step was actually completed
		if !saga.IsStepCompleted(step.Name()) {
			continue
		}

		// Skip if no compensation handler
		if !step.HasCompensation() {
			e.logger.Warn("no compensation defined for step",
				zap.String("saga_id", saga.ID),
				zap.String("step", step.Name()))
			continue
		}

		e.logger.Info("compensating saga step",
			zap.String("saga_id", saga.ID),
			zap.String("step", step.Name()))

		// Execute compensation with retry
		err := e.executeCompensationWithRetry(ctx, step, saga.Context)

		if err != nil {
			e.logger.Error("compensation failed for step",
				zap.String("saga_id", saga.ID),
				zap.String("step", step.Name()),
				zap.Error(err))

			e.metrics.RecordCompensationFailure(saga.Name, step.Name())
			compensationErrors = append(compensationErrors, err)

			// Continue with other compensations despite failure
		} else {
			e.metrics.RecordCompensationSuccess(saga.Name, step.Name())
		}
	}

	if len(compensationErrors) > 0 {
		return fmt.Errorf("compensation failed for %d steps: %v", len(compensationErrors), compensationErrors)
	}

	return nil
}

// executeCompensationWithRetry executes compensation with retry logic
func (e *SagaExecutor) executeCompensationWithRetry(ctx context.Context, step Step, data map[string]interface{}) error {
	maxAttempts := 3 // Always retry compensation
	baseDelay := 100 * time.Millisecond
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Create timeout context
		stepCtx, cancel := context.WithTimeout(ctx, step.Timeout())
		defer cancel()

		// Execute compensation
		err := step.Compensate(stepCtx, data)

		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Don't sleep after last attempt
		if attempt < maxAttempts {
			delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt-1)))

			e.logger.Warn("compensation attempt failed, retrying",
				zap.String("step", step.Name()),
				zap.Int("attempt", attempt),
				zap.Int("max_attempts", maxAttempts),
				zap.Duration("delay", delay),
				zap.Error(err))

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				continue
			}
		}
	}

	return fmt.Errorf("compensation failed after %d attempts: %w", maxAttempts, lastErr)
}

// ExecuteFrom continues saga execution from a specific step (for recovery)
func (e *SagaExecutor) ExecuteFrom(ctx context.Context, saga *Saga, fromIndex int) error {
	// Implementation for crash recovery - to be implemented if needed
	return fmt.Errorf("recovery not yet implemented")
}

// isRetryableError determines if an error is retryable
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errorStr := err.Error()

	retryableErrors := []string{
		"deadlock detected",
		"lock wait timeout exceeded",
		"connection reset by peer",
		"connection refused",
		"temporary failure",
		"serialization failure",
		"could not serialize access",
		"resource temporarily unavailable",
		"timeout",
		"context deadline exceeded",
	}

	for _, retryableErr := range retryableErrors {
		if containsString(errorStr, retryableErr) {
			return true
		}
	}

	return false
}

// containsString checks if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
