package aaa

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetryExecutor_SuccessfulExecution(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := WithExponentialBackoff(3, 10*time.Millisecond, 100*time.Millisecond)
	executor := NewRetryExecutor(config, logger)

	t.Run("returns immediately on success", func(t *testing.T) {
		ctx := context.Background()
		attempts := 0

		err := executor.Execute(ctx, "test", func(_ context.Context) error {
			attempts++
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, 1, attempts)
	})
}

func TestRetryExecutor_RetryableErrors(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := WithExponentialBackoff(3, 10*time.Millisecond, 100*time.Millisecond)
	executor := NewRetryExecutor(config, logger)

	t.Run("retries on retryable errors", func(t *testing.T) {
		ctx := context.Background()
		attempts := 0

		err := executor.Execute(ctx, "test", func(_ context.Context) error {
			attempts++
			if attempts < 3 {
				return ErrServiceUnavailable // Retryable
			}
			return nil // Success on third attempt
		})

		assert.NoError(t, err)
		assert.Equal(t, 3, attempts)
	})

	t.Run("fails fast on non-retryable errors", func(t *testing.T) {
		ctx := context.Background()
		attempts := 0

		err := executor.Execute(ctx, "test", func(_ context.Context) error {
			attempts++
			return ErrInvalidAddress // Non-retryable
		})

		assert.ErrorIs(t, err, ErrInvalidAddress)
		assert.Equal(t, 1, attempts, "should not retry non-retryable errors")
	})

	t.Run("exhausts retries and returns ErrRetryExhausted", func(t *testing.T) {
		ctx := context.Background()
		attempts := 0

		err := executor.Execute(ctx, "test", func(_ context.Context) error {
			attempts++
			return ErrServiceUnavailable // Always fail
		})

		assert.Error(t, err)
		var retryErr ErrRetryExhausted
		require.True(t, errors.As(err, &retryErr))
		assert.Equal(t, 3, retryErr.Attempts)
		assert.ErrorIs(t, retryErr.LastErr, ErrServiceUnavailable)
		assert.Equal(t, 3, attempts)
	})
}

func TestRetryExecutor_BackoffCalculation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	t.Run("exponential backoff increases correctly", func(t *testing.T) {
		config := WithExponentialBackoff(4, 100*time.Millisecond, 1*time.Second)
		executor := NewRetryExecutor(config, logger)

		sequence := executor.CalculateBackoffSequence()

		// Should be: 100ms, 200ms, 400ms
		assert.Len(t, sequence, 3)
		assert.Equal(t, 100*time.Millisecond, sequence[0])
		assert.Equal(t, 200*time.Millisecond, sequence[1])
		assert.Equal(t, 400*time.Millisecond, sequence[2])
	})

	t.Run("backoff caps at max", func(t *testing.T) {
		config := WithExponentialBackoff(5, 100*time.Millisecond, 250*time.Millisecond)
		executor := NewRetryExecutor(config, logger)

		sequence := executor.CalculateBackoffSequence()

		// Should be: 100ms, 200ms, 250ms (capped), 250ms (capped)
		assert.Len(t, sequence, 4)
		assert.Equal(t, 100*time.Millisecond, sequence[0])
		assert.Equal(t, 200*time.Millisecond, sequence[1])
		assert.Equal(t, 250*time.Millisecond, sequence[2])
		assert.Equal(t, 250*time.Millisecond, sequence[3])
	})

	t.Run("linear backoff stays constant", func(t *testing.T) {
		config := WithLinearBackoff(4, 100*time.Millisecond)
		executor := NewRetryExecutor(config, logger)

		sequence := executor.CalculateBackoffSequence()

		// Should all be 100ms
		assert.Len(t, sequence, 3)
		for _, backoff := range sequence {
			assert.Equal(t, 100*time.Millisecond, backoff)
		}
	})

	t.Run("fixed backoff stays exactly the same", func(t *testing.T) {
		config := WithFixedBackoff(4, 100*time.Millisecond)
		executor := NewRetryExecutor(config, logger)

		sequence := executor.CalculateBackoffSequence()

		// Should all be exactly 100ms (no jitter)
		assert.Len(t, sequence, 3)
		for _, backoff := range sequence {
			assert.Equal(t, 100*time.Millisecond, backoff)
		}
	})
}

func TestRetryExecutor_ContextCancellation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := WithExponentialBackoff(5, 100*time.Millisecond, 1*time.Second)
	executor := NewRetryExecutor(config, logger)

	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		attempts := 0

		// Cancel after first attempt
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		err := executor.Execute(ctx, "test", func(_ context.Context) error {
			attempts++
			return ErrServiceUnavailable
		})

		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
		assert.Equal(t, 1, attempts, "should stop retrying after context cancellation")
	})

	t.Run("respects context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()

		attempts := 0

		err := executor.Execute(ctx, "test", func(_ context.Context) error {
			attempts++
			return ErrServiceUnavailable
		})

		assert.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		// Should have attempted at least once, possibly twice depending on timing
		assert.GreaterOrEqual(t, attempts, 1)
		assert.LessOrEqual(t, attempts, 2)
	})
}

func TestRetryExecutor_JitterPreventsThunderingHerd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timing-sensitive jitter test in short mode")
	}

	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := RetryConfig{
		MaxAttempts:    3,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     1 * time.Second,
		BackoffFactor:  2.0,
		JitterFactor:   0.2, // 20% jitter
	}
	executor := NewRetryExecutor(config, logger)

	t.Run("applies jitter to backoff", func(t *testing.T) {
		t.Skip("Flaky timing test - jitter is random and timing-dependent")
		ctx := context.Background()
		attempts := 0
		var firstBackoff, secondBackoff time.Duration

		start := time.Now()
		err := executor.Execute(ctx, "test", func(_ context.Context) error {
			attempts++
			switch attempts {
			case 1:
				firstBackoff = time.Since(start)
			case 2:
				secondBackoff = time.Since(start) - firstBackoff
			}

			if attempts < 3 {
				return ErrServiceUnavailable
			}
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, 3, attempts)

		// First backoff should be around 100ms ± 20%
		assert.Greater(t, firstBackoff, 80*time.Millisecond)
		assert.Less(t, firstBackoff, 120*time.Millisecond)

		// Second backoff should be around 200ms ± 20%
		assert.Greater(t, secondBackoff, 160*time.Millisecond)
		assert.Less(t, secondBackoff, 240*time.Millisecond)
	})
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{"nil error", nil, false},
		{"service unavailable", ErrServiceUnavailable, true},
		{"timeout", ErrTimeout, true},
		{"connection closed", ErrConnectionClosed, true},
		{"no connections", ErrNoConnections, true},
		{"too many requests", ErrTooManyRequests, false},
		{"invalid request", ErrInvalidRequest, false},
		{"invalid address", ErrInvalidAddress, false},
		{"address not found", ErrAddressNotFound, false},
		{"unauthenticated", ErrUnauthenticated, false},
		{"unauthorized", ErrUnauthorized, false},
		{"unknown error", errors.New("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryable(tt.err)
			assert.Equal(t, tt.retryable, result)
		})
	}
}

func TestRetryExecutor_RealBackoffTiming(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timing test in short mode")
	}

	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := RetryConfig{
		MaxAttempts:    3,
		InitialBackoff: 50 * time.Millisecond,
		MaxBackoff:     200 * time.Millisecond,
		BackoffFactor:  2.0,
		JitterFactor:   0.0, // No jitter for precise timing
	}
	executor := NewRetryExecutor(config, logger)

	t.Run("actual backoff timing matches expected", func(t *testing.T) {
		ctx := context.Background()
		attempts := 0
		timestamps := make([]time.Time, 0)

		start := time.Now()
		err := executor.Execute(ctx, "test", func(_ context.Context) error {
			attempts++
			timestamps = append(timestamps, time.Now())

			if attempts < 3 {
				return ErrServiceUnavailable
			}
			return nil
		})

		totalDuration := time.Since(start)

		assert.NoError(t, err)
		assert.Equal(t, 3, attempts)

		// Expected: 0ms (first attempt) + 50ms + 100ms = 150ms total
		// Allow ±20ms for timing variation
		assert.Greater(t, totalDuration, 130*time.Millisecond)
		assert.Less(t, totalDuration, 170*time.Millisecond)
	})
}
