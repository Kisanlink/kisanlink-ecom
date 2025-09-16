package resilience

import (
	"context"
	"errors"
	"testing"
	"time"

	"kisanlink-ecom/internal/resilience"

	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker_DefaultConfig(t *testing.T) {
	config := resilience.DefaultCircuitBreakerConfig("test")
	assert.Equal(t, "test", config.Name)
	assert.Equal(t, uint32(1), config.MaxRequests)
	assert.Equal(t, 60*time.Second, config.Interval)
	assert.Equal(t, 60*time.Second, config.Timeout)
	assert.NotNil(t, config.ReadyToTrip)
	assert.NotNil(t, config.OnStateChange)
}

func TestCircuitBreaker_NewCircuitBreaker(t *testing.T) {
	config := resilience.DefaultCircuitBreakerConfig("test")
	cb := resilience.NewCircuitBreaker(config)

	assert.Equal(t, "test", cb.Name())
	assert.Equal(t, resilience.StateClosed, cb.State())

	counts := cb.Counts()
	assert.Equal(t, uint32(0), counts.Requests)
	assert.Equal(t, uint32(0), counts.TotalSuccesses)
	assert.Equal(t, uint32(0), counts.TotalFailures)
}

func TestCircuitBreaker_SuccessfulExecution(t *testing.T) {
	config := resilience.DefaultCircuitBreakerConfig("test")
	cb := resilience.NewCircuitBreaker(config)

	result, err := cb.Execute(func() (interface{}, error) {
		return "success", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, resilience.StateClosed, cb.State())

	counts := cb.Counts()
	assert.Equal(t, uint32(1), counts.Requests)
	assert.Equal(t, uint32(1), counts.TotalSuccesses)
	assert.Equal(t, uint32(0), counts.TotalFailures)
}

func TestCircuitBreaker_FailedExecution(t *testing.T) {
	config := resilience.DefaultCircuitBreakerConfig("test")
	cb := resilience.NewCircuitBreaker(config)

	expectedErr := errors.New("test error")
	result, err := cb.Execute(func() (interface{}, error) {
		return nil, expectedErr
	})

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, result)
	assert.Equal(t, resilience.StateClosed, cb.State())

	counts := cb.Counts()
	assert.Equal(t, uint32(1), counts.Requests)
	assert.Equal(t, uint32(0), counts.TotalSuccesses)
	assert.Equal(t, uint32(1), counts.TotalFailures)
}

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	config := &resilience.CircuitBreakerConfig{
		Name:        "test",
		MaxRequests: 1,
		Interval:    100 * time.Millisecond,
		Timeout:     100 * time.Millisecond,
		ReadyToTrip: func(counts resilience.Counts) bool {
			return counts.Requests >= 2 && counts.TotalFailures >= 2
		},
	}
	cb := resilience.NewCircuitBreaker(config)

	// Initial state should be closed
	assert.Equal(t, resilience.StateClosed, cb.State())

	// First failure - should remain closed
	_, err := cb.Execute(func() (interface{}, error) {
		return nil, errors.New("error 1")
	})
	assert.Error(t, err)
	assert.Equal(t, resilience.StateClosed, cb.State())

	// Second failure - should trip to open
	_, err = cb.Execute(func() (interface{}, error) {
		return nil, errors.New("error 2")
	})
	assert.Error(t, err)
	assert.Equal(t, resilience.StateOpen, cb.State())

	// Request while open - should be rejected
	_, err = cb.Execute(func() (interface{}, error) {
		return "should not execute", nil
	})
	assert.Error(t, err)
	assert.Equal(t, resilience.ErrCircuitBreakerOpen, err)

	// Wait for timeout and transition to half-open
	time.Sleep(150 * time.Millisecond)
	assert.Equal(t, resilience.StateHalfOpen, cb.State())

	// Successful request in half-open - should close
	_, err = cb.Execute(func() (interface{}, error) {
		return "success", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, resilience.StateClosed, cb.State())
}

func TestCircuitBreaker_HalfOpenLimitRequests(t *testing.T) {
	config := &resilience.CircuitBreakerConfig{
		Name:        "test",
		MaxRequests: 2,
		Interval:    100 * time.Millisecond,
		Timeout:     100 * time.Millisecond,
		ReadyToTrip: func(counts resilience.Counts) bool {
			return counts.Requests >= 1 && counts.TotalFailures >= 1
		},
	}
	cb := resilience.NewCircuitBreaker(config)

	// Cause circuit to open
	_, err := cb.Execute(func() (interface{}, error) {
		return nil, errors.New("error")
	})
	assert.Error(t, err)
	assert.Equal(t, resilience.StateOpen, cb.State())

	// Wait for transition to half-open
	time.Sleep(150 * time.Millisecond)
	assert.Equal(t, resilience.StateHalfOpen, cb.State())

	// First request in half-open - should succeed and transition to closed
	_, err = cb.Execute(func() (interface{}, error) {
		return "success 1", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, resilience.StateClosed, cb.State())

	// Second request (now in closed state) - should succeed
	_, err = cb.Execute(func() (interface{}, error) {
		return "success 2", nil
	})
	assert.NoError(t, err)

	// Third request (still in closed state) - should succeed
	_, err = cb.Execute(func() (interface{}, error) {
		return "success 3", nil
	})
	assert.NoError(t, err)
}

func TestCircuitBreaker_Context(t *testing.T) {
	config := resilience.DefaultCircuitBreakerConfig("test")
	cb := resilience.NewCircuitBreaker(config)

	ctx, cancel := context.WithCancel(context.Background())

	// Test with valid context
	result, err := cb.ExecuteContext(ctx, func(ctx context.Context) (interface{}, error) {
		return "success", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "success", result)

	// Test with cancelled context
	cancel()
	result, err = cb.ExecuteContext(ctx, func(ctx context.Context) (interface{}, error) {
		return nil, ctx.Err()
	})
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestCircuitBreaker_Call(t *testing.T) {
	config := resilience.DefaultCircuitBreakerConfig("test")
	cb := resilience.NewCircuitBreaker(config)

	// Test successful call
	err := cb.Call(func() error {
		return nil
	})
	assert.NoError(t, err)

	// Test failed call
	expectedErr := errors.New("test error")
	err = cb.Call(func() error {
		return expectedErr
	})
	assert.Equal(t, expectedErr, err)
}

func TestCircuitBreaker_CallContext(t *testing.T) {
	config := resilience.DefaultCircuitBreakerConfig("test")
	cb := resilience.NewCircuitBreaker(config)

	ctx := context.Background()

	// Test successful call
	err := cb.CallContext(ctx, func(ctx context.Context) error {
		return nil
	})
	assert.NoError(t, err)

	// Test failed call
	expectedErr := errors.New("test error")
	err = cb.CallContext(ctx, func(ctx context.Context) error {
		return expectedErr
	})
	assert.Equal(t, expectedErr, err)
}

func TestCircuitBreaker_Panic(t *testing.T) {
	config := resilience.DefaultCircuitBreakerConfig("test")
	cb := resilience.NewCircuitBreaker(config)

	assert.Panics(t, func() {
		cb.Execute(func() (interface{}, error) {
			panic("test panic")
		})
	})

	// Check that failure was recorded
	counts := cb.Counts()
	assert.Equal(t, uint32(1), counts.Requests)
	assert.Equal(t, uint32(1), counts.TotalFailures)
}

func TestCircuitBreakerManager(t *testing.T) {
	manager := resilience.NewCircuitBreakerManager()

	// Test GetOrCreate
	cb1 := manager.GetOrCreate("test1", nil)
	assert.NotNil(t, cb1)
	assert.Equal(t, "test1", cb1.Name())

	// Getting same name should return same instance
	cb2 := manager.GetOrCreate("test1", nil)
	assert.Equal(t, cb1, cb2)

	// Test Get
	cb3, exists := manager.Get("test1")
	assert.True(t, exists)
	assert.Equal(t, cb1, cb3)

	cb4, exists := manager.Get("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, cb4)

	// Test List
	list := manager.List()
	assert.Len(t, list, 1)
	assert.Contains(t, list, "test1")

	// Test Remove
	removed := manager.Remove("test1")
	assert.True(t, removed)

	removed = manager.Remove("nonexistent")
	assert.False(t, removed)

	// Test Clear
	manager.GetOrCreate("test2", nil)
	manager.GetOrCreate("test3", nil)
	assert.Len(t, manager.List(), 2)

	manager.Clear()
	assert.Len(t, manager.List(), 0)
}

func TestGlobalCircuitBreakerFunctions(t *testing.T) {
	// Clean up before test
	resilience.ClearCircuitBreakers()

	// Test GetCircuitBreaker
	cb1 := resilience.GetCircuitBreaker("global1")
	assert.NotNil(t, cb1)
	assert.Equal(t, "global1", cb1.Name())

	// Test GetCircuitBreakerWithConfig
	config := resilience.DefaultCircuitBreakerConfig("global2")
	cb2 := resilience.GetCircuitBreakerWithConfig("global2", config)
	assert.NotNil(t, cb2)
	assert.Equal(t, "global2", cb2.Name())

	// Test ListCircuitBreakers
	list := resilience.ListCircuitBreakers()
	assert.Len(t, list, 2)

	// Test RemoveCircuitBreaker
	removed := resilience.RemoveCircuitBreaker("global1")
	assert.True(t, removed)

	list = resilience.ListCircuitBreakers()
	assert.Len(t, list, 1)

	// Test ClearCircuitBreakers
	resilience.ClearCircuitBreakers()
	list = resilience.ListCircuitBreakers()
	assert.Len(t, list, 0)
}

func TestConvenienceFunctions(t *testing.T) {
	resilience.ClearCircuitBreakers()

	// Test ExecuteWithCircuitBreaker
	err := resilience.ExecuteWithCircuitBreaker("convenience1", func() error {
		return nil
	})
	assert.NoError(t, err)

	// Test ExecuteWithCircuitBreakerContext
	ctx := context.Background()
	err = resilience.ExecuteWithCircuitBreakerContext(ctx, "convenience2", func(ctx context.Context) error {
		return nil
	})
	assert.NoError(t, err)

	// Test ExecuteWithCustomCircuitBreaker
	config := resilience.DefaultCircuitBreakerConfig("convenience3")
	err = resilience.ExecuteWithCustomCircuitBreaker("convenience3", config, func() error {
		return nil
	})
	assert.NoError(t, err)

	// Verify circuit breakers were created
	list := resilience.ListCircuitBreakers()
	assert.Len(t, list, 3)
}

func TestIsCircuitBreakerError(t *testing.T) {
	assert.True(t, resilience.IsCircuitBreakerError(resilience.ErrCircuitBreakerOpen))
	assert.True(t, resilience.IsCircuitBreakerError(resilience.ErrTooManyRequests))
	assert.False(t, resilience.IsCircuitBreakerError(errors.New("other error")))
	assert.False(t, resilience.IsCircuitBreakerError(nil))
}

func TestGetCircuitBreakerStats(t *testing.T) {
	resilience.ClearCircuitBreakers()

	// Create circuit breakers in different states
	_ = resilience.GetCircuitBreaker("stats1")

	// Create cb2 with custom config that will trip on first failure
	config := &resilience.CircuitBreakerConfig{
		Name:        "stats2",
		MaxRequests: 1,
		Interval:    100 * time.Millisecond,
		Timeout:     100 * time.Millisecond,
		ReadyToTrip: func(counts resilience.Counts) bool {
			return counts.TotalFailures >= 1
		},
	}
	cb2 := resilience.GetCircuitBreakerWithConfig("stats2", config)

	// Make cb2 fail and trip
	cb2.Execute(func() (interface{}, error) {
		return nil, errors.New("force failure")
	})

	stats := resilience.GetCircuitBreakerStats()
	assert.Equal(t, 2, stats.TotalBreakers)
	assert.Equal(t, 1, stats.ClosedBreakers)
	assert.Equal(t, 1, stats.OpenBreakers)
	assert.Equal(t, 0, stats.HalfOpenBreakers)
}

func TestResetAndForceOpen(t *testing.T) {
	resilience.ClearCircuitBreakers()

	config := &resilience.CircuitBreakerConfig{
		Name:        "test",
		MaxRequests: 1,
		Interval:    100 * time.Millisecond,
		Timeout:     100 * time.Millisecond,
		ReadyToTrip: func(counts resilience.Counts) bool {
			return counts.TotalFailures >= 1
		},
	}
	cb := resilience.GetCircuitBreakerWithConfig("test", config)

	// Make it fail and trip
	cb.Execute(func() (interface{}, error) {
		return nil, errors.New("force failure")
	})
	assert.Equal(t, resilience.StateOpen, cb.State())

	// Reset to closed
	err := resilience.ResetCircuitBreaker("test")
	assert.NoError(t, err)
	assert.Equal(t, resilience.StateClosed, cb.State())

	// Force open
	err = resilience.ForceOpenCircuitBreaker("test")
	assert.NoError(t, err)
	assert.Equal(t, resilience.StateOpen, cb.State())

	// Test with non-existent circuit breaker
	err = resilience.ResetCircuitBreaker("nonexistent")
	assert.Error(t, err)

	err = resilience.ForceOpenCircuitBreaker("nonexistent")
	assert.Error(t, err)
}

func TestCounts(t *testing.T) {
	var counts resilience.Counts

	// Test initial state
	assert.Equal(t, uint32(0), counts.Requests)
	assert.Equal(t, uint32(0), counts.TotalSuccesses)
	assert.Equal(t, uint32(0), counts.TotalFailures)
	assert.Equal(t, uint32(0), counts.ConsecutiveSuccesses)
	assert.Equal(t, uint32(0), counts.ConsecutiveFailures)

	// Test OnRequest
	counts.OnRequest()
	assert.Equal(t, uint32(1), counts.Requests)

	// Test OnSuccess
	counts.OnSuccess()
	assert.Equal(t, uint32(1), counts.TotalSuccesses)
	assert.Equal(t, uint32(1), counts.ConsecutiveSuccesses)
	assert.Equal(t, uint32(0), counts.ConsecutiveFailures)

	// Test OnFailure
	counts.OnFailure()
	assert.Equal(t, uint32(1), counts.TotalFailures)
	assert.Equal(t, uint32(1), counts.ConsecutiveFailures)
	assert.Equal(t, uint32(0), counts.ConsecutiveSuccesses)

	// Test Reset
	counts.Reset()
	assert.Equal(t, uint32(0), counts.Requests)
	assert.Equal(t, uint32(0), counts.TotalSuccesses)
	assert.Equal(t, uint32(0), counts.TotalFailures)
	assert.Equal(t, uint32(0), counts.ConsecutiveSuccesses)
	assert.Equal(t, uint32(0), counts.ConsecutiveFailures)
}

func TestStateString(t *testing.T) {
	assert.Equal(t, "CLOSED", resilience.StateClosed.String())
	assert.Equal(t, "OPEN", resilience.StateOpen.String())
	assert.Equal(t, "HALF_OPEN", resilience.StateHalfOpen.String())
}

func BenchmarkCircuitBreaker_SuccessfulExecution(b *testing.B) {
	config := resilience.DefaultCircuitBreakerConfig("benchmark")
	cb := resilience.NewCircuitBreaker(config)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cb.Execute(func() (interface{}, error) {
				return "success", nil
			})
		}
	})
}

func BenchmarkCircuitBreaker_FailedExecution(b *testing.B) {
	config := resilience.DefaultCircuitBreakerConfig("benchmark")
	cb := resilience.NewCircuitBreaker(config)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cb.Execute(func() (interface{}, error) {
				return nil, errors.New("error")
			})
		}
	})
}
