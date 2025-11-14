package saga_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Kisanlink/kisanlink-ecom/internal/saga"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// mockSagaStorage is a simple in-memory storage for testing
type mockSagaStorage struct {
	sagas map[string]*saga.Saga
}

func newMockSagaStorage() *mockSagaStorage {
	return &mockSagaStorage{
		sagas: make(map[string]*saga.Saga),
	}
}

func (m *mockSagaStorage) SaveState(ctx context.Context, s *saga.Saga) error {
	m.sagas[s.ID] = s
	return nil
}

func (m *mockSagaStorage) LoadState(ctx context.Context, sagaID string) (*saga.Saga, error) {
	s, ok := m.sagas[sagaID]
	if !ok {
		return nil, errors.New("saga not found")
	}
	return s, nil
}

func (m *mockSagaStorage) ListPendingSagas(ctx context.Context) ([]*saga.Saga, error) {
	pending := make([]*saga.Saga, 0)
	for _, s := range m.sagas {
		if s.State == saga.SagaStateRunning || s.State == saga.SagaStateCompensating {
			pending = append(pending, s)
		}
	}
	return pending, nil
}

func (m *mockSagaStorage) CleanupCompleted(ctx context.Context, olderThan time.Duration) error {
	return nil
}

func TestExecutor_SuccessfulExecution(t *testing.T) {
	storage := newMockSagaStorage()
	logger := zap.NewNop()
	metrics := saga.NewSagaMetrics()
	executor := saga.NewSagaExecutor(storage, logger, metrics)

	s := saga.NewSaga("test_saga", logger)

	// Add steps
	step1Executed := false
	step2Executed := false

	s.AddStep(saga.NewSagaStep("step1",
		func(ctx context.Context, data map[string]interface{}) error {
			step1Executed = true
			data["step1"] = "completed"
			return nil
		},
		nil,
	))

	s.AddStep(saga.NewSagaStep("step2",
		func(ctx context.Context, data map[string]interface{}) error {
			step2Executed = true
			data["step2"] = "completed"
			return nil
		},
		nil,
	))

	// Execute
	err := executor.Execute(context.Background(), s)

	assert.NoError(t, err)
	assert.True(t, step1Executed)
	assert.True(t, step2Executed)
	assert.Equal(t, saga.SagaStateCompleted, s.State)
	assert.Len(t, s.CompletedSteps, 2)
	assert.Equal(t, "completed", s.Context["step1"])
	assert.Equal(t, "completed", s.Context["step2"])
}

func TestExecutor_CompensationOnFailure(t *testing.T) {
	storage := newMockSagaStorage()
	logger := zap.NewNop()
	metrics := saga.NewSagaMetrics()
	executor := saga.NewSagaExecutor(storage, logger, metrics)

	s := saga.NewSaga("test_saga", logger)

	step1Compensated := false
	step2Executed := false

	// Step 1 succeeds
	s.AddStep(saga.NewSagaStep("step1",
		func(ctx context.Context, data map[string]interface{}) error {
			data["step1"] = "completed"
			return nil
		},
		func(ctx context.Context, data map[string]interface{}) error {
			step1Compensated = true
			delete(data, "step1")
			return nil
		},
	))

	// Step 2 fails
	s.AddStep(saga.NewSagaStep("step2",
		func(ctx context.Context, data map[string]interface{}) error {
			step2Executed = true
			return errors.New("deliberate failure")
		},
		nil,
	))

	// Execute
	err := executor.Execute(context.Background(), s)

	assert.Error(t, err)
	assert.True(t, step2Executed)
	assert.True(t, step1Compensated)
	assert.Equal(t, saga.SagaStateCompensated, s.State)
	_, exists := s.Context["step1"]
	assert.False(t, exists)
}

func TestExecutor_NoCompensationIfFirstStepFails(t *testing.T) {
	storage := newMockSagaStorage()
	logger := zap.NewNop()
	metrics := saga.NewSagaMetrics()
	executor := saga.NewSagaExecutor(storage, logger, metrics)

	s := saga.NewSaga("test_saga", logger)

	// First step fails immediately
	s.AddStep(saga.NewSagaStep("step1",
		func(ctx context.Context, data map[string]interface{}) error {
			return errors.New("first step failed")
		},
		func(ctx context.Context, data map[string]interface{}) error {
			t.Fatal("compensation should not be called for failed first step")
			return nil
		},
	))

	// Execute
	err := executor.Execute(context.Background(), s)

	assert.Error(t, err)
	assert.Equal(t, saga.SagaStateCompensated, s.State)
	assert.Empty(t, s.CompletedSteps)
}

func TestExecutor_RetryOnTransientFailure(t *testing.T) {
	storage := newMockSagaStorage()
	logger := zap.NewNop()
	metrics := saga.NewSagaMetrics()
	executor := saga.NewSagaExecutor(storage, logger, metrics)

	s := saga.NewSaga("test_saga", logger)

	attemptCount := 0
	s.AddStep(saga.NewSagaStep("retryable_step",
		func(ctx context.Context, data map[string]interface{}) error {
			attemptCount++
			if attemptCount < 3 {
				return errors.New("timeout: connection failed")
			}
			data["success"] = true
			return nil
		},
		nil,
	).WithRetry(3))

	// Execute
	err := executor.Execute(context.Background(), s)

	assert.NoError(t, err)
	assert.Equal(t, 3, attemptCount)
	assert.Equal(t, saga.SagaStateCompleted, s.State)
	assert.True(t, s.Context["success"].(bool))
}

func TestExecutor_MetricsRecording(t *testing.T) {
	storage := newMockSagaStorage()
	logger := zap.NewNop()
	metrics := saga.NewSagaMetrics()
	executor := saga.NewSagaExecutor(storage, logger, metrics)

	s := saga.NewSaga("test_saga", logger)

	s.AddStep(saga.NewSagaStep("step1",
		func(ctx context.Context, data map[string]interface{}) error {
			return nil
		},
		nil,
	))

	// Execute successful saga
	err := executor.Execute(context.Background(), s)
	assert.NoError(t, err)

	// Check metrics
	stats := metrics.GetSagaStats("test_saga")
	assert.Equal(t, int64(1), stats.Started)
	assert.Equal(t, int64(1), stats.Completed)
	assert.Equal(t, int64(0), stats.Failed)
	assert.Equal(t, int64(0), stats.Compensated)
	assert.Len(t, stats.Durations, 1)
}

func TestExecutor_CompensationMetrics(t *testing.T) {
	storage := newMockSagaStorage()
	logger := zap.NewNop()
	metrics := saga.NewSagaMetrics()
	executor := saga.NewSagaExecutor(storage, logger, metrics)

	s := saga.NewSaga("test_saga", logger)

	s.AddStep(saga.NewSagaStep("step1",
		func(ctx context.Context, data map[string]interface{}) error {
			return nil
		},
		func(ctx context.Context, data map[string]interface{}) error {
			return nil
		},
	))

	s.AddStep(saga.NewSagaStep("step2",
		func(ctx context.Context, data map[string]interface{}) error {
			return errors.New("failure")
		},
		nil,
	))

	// Execute
	err := executor.Execute(context.Background(), s)
	assert.Error(t, err)

	// Check metrics
	stats := metrics.GetSagaStats("test_saga")
	assert.Equal(t, int64(1), stats.Started)
	assert.Equal(t, int64(0), stats.Completed)
	assert.Equal(t, int64(1), stats.Compensated)
}
