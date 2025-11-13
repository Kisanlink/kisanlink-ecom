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

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs in tests

	config := CircuitBreakerConfig{
		MaxRequests:  5,
		Interval:     1 * time.Second,
		Timeout:      1 * time.Second,
		FailureRatio: 0.6,
		MinRequests:  10,
	}

	t.Run("starts in closed state", func(t *testing.T) {
		cb := NewCircuitBreaker("test", config, logger)
		assert.Equal(t, StateClosed, cb.State())
	})

	t.Run("transitions to open on high failure ratio", func(t *testing.T) {
		cb := NewCircuitBreaker("test", config, logger)

		// Generate enough failures to trip (need MinRequests = 10)
		ctx := context.Background()
		for i := 0; i < 10; i++ {
			_ = cb.Execute(ctx, func(_ context.Context) error {
				return errors.New("failure")
			})
		}

		// Circuit should be open now
		assert.Equal(t, StateOpen, cb.State())
	})

	t.Run("transitions from open to half-open after timeout", func(t *testing.T) {
		shortConfig := config
		shortConfig.Timeout = 100 * time.Millisecond

		cb := NewCircuitBreaker("test", shortConfig, logger)

		// Trip the circuit
		ctx := context.Background()
		for i := 0; i < 10; i++ {
			_ = cb.Execute(ctx, func(_ context.Context) error {
				return errors.New("failure")
			})
		}

		assert.Equal(t, StateOpen, cb.State())

		// Wait for timeout
		time.Sleep(150 * time.Millisecond)

		// Make a request to trigger state check
		err := cb.Execute(ctx, func(_ context.Context) error {
			return nil
		})

		// Should be half-open now (not open)
		assert.NoError(t, err)
	})

	t.Run("transitions from half-open to closed on consecutive successes", func(t *testing.T) {
		shortConfig := config
		shortConfig.Timeout = 100 * time.Millisecond
		shortConfig.MaxRequests = 3 // Need 3 consecutive successes

		cb := NewCircuitBreaker("test", shortConfig, logger)

		// Trip the circuit
		ctx := context.Background()
		for i := 0; i < 10; i++ {
			_ = cb.Execute(ctx, func(_ context.Context) error {
				return errors.New("failure")
			})
		}

		assert.Equal(t, StateOpen, cb.State())

		// Wait for timeout
		time.Sleep(150 * time.Millisecond)

		// Make MaxRequests successful requests
		for i := 0; i < int(shortConfig.MaxRequests); i++ {
			err := cb.Execute(ctx, func(_ context.Context) error {
				return nil
			})
			require.NoError(t, err, "request %d should succeed", i)
		}

		// Should be closed now
		assert.Equal(t, StateClosed, cb.State())
	})

	t.Run("transitions from half-open to open on failure", func(t *testing.T) {
		shortConfig := config
		shortConfig.Timeout = 100 * time.Millisecond

		cb := NewCircuitBreaker("test", shortConfig, logger)

		// Trip the circuit
		ctx := context.Background()
		for i := 0; i < 10; i++ {
			_ = cb.Execute(ctx, func(_ context.Context) error {
				return errors.New("failure")
			})
		}

		assert.Equal(t, StateOpen, cb.State())

		// Wait for timeout to transition to half-open
		time.Sleep(150 * time.Millisecond)

		// Make one successful request (transitions to half-open)
		err := cb.Execute(ctx, func(_ context.Context) error {
			return nil
		})
		require.NoError(t, err)

		// Make a failing request
		_ = cb.Execute(ctx, func(_ context.Context) error {
			return errors.New("failure")
		})

		// Should be open again
		assert.Equal(t, StateOpen, cb.State())
	})
}

func TestCircuitBreaker_ErrorHandling(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := CircuitBreakerConfig{
		MaxRequests:  5,
		Interval:     1 * time.Second,
		Timeout:      1 * time.Second,
		FailureRatio: 0.6,
		MinRequests:  5, // Lower for faster tests
	}

	t.Run("returns ErrCircuitOpen when open", func(t *testing.T) {
		cb := NewCircuitBreaker("test", config, logger)

		// Trip the circuit
		ctx := context.Background()
		for i := 0; i < 5; i++ {
			_ = cb.Execute(ctx, func(_ context.Context) error {
				return errors.New("failure")
			})
		}

		assert.Equal(t, StateOpen, cb.State())

		// Next request should fail with ErrCircuitOpen
		err := cb.Execute(ctx, func(_ context.Context) error {
			return nil
		})

		assert.ErrorIs(t, err, ErrCircuitOpen)
	})

	t.Run("closes after MaxRequests successes in half-open state", func(t *testing.T) {
		shortConfig := config
		shortConfig.Timeout = 100 * time.Millisecond
		shortConfig.MaxRequests = 2 // Need 2 consecutive successes to close

		cb := NewCircuitBreaker("test", shortConfig, logger)

		// Trip the circuit
		ctx := context.Background()
		for i := 0; i < 5; i++ {
			_ = cb.Execute(ctx, func(_ context.Context) error {
				return errors.New("failure")
			})
		}

		assert.Equal(t, StateOpen, cb.State())

		// Wait for timeout
		time.Sleep(150 * time.Millisecond)

		// First 2 requests should succeed and close the circuit
		for i := 0; i < 2; i++ {
			err := cb.Execute(ctx, func(_ context.Context) error {
				return nil
			})
			assert.NoError(t, err)
		}

		// Circuit should be closed now
		assert.Equal(t, StateClosed, cb.State())

		// Additional requests should also succeed
		err := cb.Execute(ctx, func(_ context.Context) error {
			return nil
		})
		assert.NoError(t, err)
	})
}

func TestCircuitBreaker_Counts(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := CircuitBreakerConfig{
		MaxRequests:  5,
		Interval:     1 * time.Second,
		Timeout:      1 * time.Second,
		FailureRatio: 0.6,
		MinRequests:  10,
	}

	t.Run("tracks request counts correctly", func(t *testing.T) {
		cb := NewCircuitBreaker("test", config, logger)

		ctx := context.Background()

		// 3 successes
		for i := 0; i < 3; i++ {
			_ = cb.Execute(ctx, func(_ context.Context) error {
				return nil
			})
		}

		// 2 failures
		for i := 0; i < 2; i++ {
			_ = cb.Execute(ctx, func(_ context.Context) error {
				return errors.New("failure")
			})
		}

		requests, successes, failures, ratio := cb.Counts()
		assert.Equal(t, uint32(5), requests)
		assert.Equal(t, uint32(3), successes)
		assert.Equal(t, uint32(2), failures)
		assert.InDelta(t, 0.4, ratio, 0.01)
	})

	t.Run("clears counts after interval", func(t *testing.T) {
		shortConfig := config
		shortConfig.Interval = 100 * time.Millisecond

		cb := NewCircuitBreaker("test", shortConfig, logger)

		ctx := context.Background()

		// Make some requests
		for i := 0; i < 3; i++ {
			_ = cb.Execute(ctx, func(_ context.Context) error {
				return nil
			})
		}

		requests, _, _, _ := cb.Counts()
		assert.Equal(t, uint32(3), requests)

		// Wait for interval to expire
		time.Sleep(150 * time.Millisecond)

		// Make another request to trigger count clear
		_ = cb.Execute(ctx, func(_ context.Context) error {
			return nil
		})

		requests, _, _, _ = cb.Counts()
		// Should be reset and only have the last request
		assert.Equal(t, uint32(1), requests)
	})
}

func TestCircuitBreaker_Reset(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := CircuitBreakerConfig{
		MaxRequests:  5,
		Interval:     1 * time.Second,
		Timeout:      1 * time.Second,
		FailureRatio: 0.6,
		MinRequests:  5,
	}

	t.Run("manual reset closes the circuit", func(t *testing.T) {
		cb := NewCircuitBreaker("test", config, logger)

		// Trip the circuit
		ctx := context.Background()
		for i := 0; i < 5; i++ {
			_ = cb.Execute(ctx, func(_ context.Context) error {
				return errors.New("failure")
			})
		}

		assert.Equal(t, StateOpen, cb.State())

		// Reset the circuit
		cb.Reset()

		assert.Equal(t, StateClosed, cb.State())

		// Should accept requests now
		err := cb.Execute(ctx, func(_ context.Context) error {
			return nil
		})
		assert.NoError(t, err)
	})
}

func TestCircuitBreaker_FailureRatioThreshold(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := CircuitBreakerConfig{
		MaxRequests:  5,
		Interval:     1 * time.Second,
		Timeout:      1 * time.Second,
		FailureRatio: 0.5, // 50% threshold
		MinRequests:  10,
	}

	t.Run("does not trip when below failure ratio", func(t *testing.T) {
		cb := NewCircuitBreaker("test", config, logger)

		ctx := context.Background()

		// 6 successes, 4 failures = 40% failure ratio (below 50%)
		for i := 0; i < 6; i++ {
			_ = cb.Execute(ctx, func(_ context.Context) error {
				return nil
			})
		}
		for i := 0; i < 4; i++ {
			_ = cb.Execute(ctx, func(_ context.Context) error {
				return errors.New("failure")
			})
		}

		assert.Equal(t, StateClosed, cb.State())
	})

	t.Run("trips when at or above failure ratio", func(t *testing.T) {
		cb := NewCircuitBreaker("test", config, logger)

		ctx := context.Background()

		// 5 successes, 5 failures = 50% failure ratio (at threshold)
		for i := 0; i < 5; i++ {
			_ = cb.Execute(ctx, func(_ context.Context) error {
				return nil
			})
		}
		for i := 0; i < 5; i++ {
			_ = cb.Execute(ctx, func(_ context.Context) error {
				return errors.New("failure")
			})
		}

		assert.Equal(t, StateOpen, cb.State())
	})
}

func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := CircuitBreakerConfig{
		MaxRequests:  5,
		Interval:     1 * time.Second,
		Timeout:      1 * time.Second,
		FailureRatio: 0.6,
		MinRequests:  100,
	}

	t.Run("handles concurrent requests safely", func(t *testing.T) {
		cb := NewCircuitBreaker("test", config, logger)

		ctx := context.Background()
		concurrency := 50

		done := make(chan bool, concurrency)

		// Launch concurrent requests
		for i := 0; i < concurrency; i++ {
			go func(id int) {
				for j := 0; j < 10; j++ {
					_ = cb.Execute(ctx, func(_ context.Context) error {
						if (id+j)%2 == 0 {
							return nil
						}
						return errors.New("failure")
					})
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < concurrency; i++ {
			<-done
		}

		// Verify state is consistent
		state := cb.State()
		assert.Contains(t, []CircuitState{StateClosed, StateOpen}, state)

		// Verify counts
		requests, _, _, _ := cb.Counts()
		assert.Greater(t, requests, uint32(0))
	})
}
