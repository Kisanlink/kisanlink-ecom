package saga_test

import (
	"context"
	"testing"
	"time"

	"kisanlink-ecom/internal/saga"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewSaga(t *testing.T) {
	logger := zap.NewNop()
	s := saga.NewSaga("test_saga", logger)

	assert.NotEmpty(t, s.ID)
	assert.Equal(t, "test_saga", s.Name)
	assert.Equal(t, saga.SagaStatePending, s.State)
	assert.NotNil(t, s.Context)
	assert.NotNil(t, s.CompletedSteps)
	assert.Empty(t, s.Steps)
}

func TestSaga_AddStep(t *testing.T) {
	logger := zap.NewNop()
	s := saga.NewSaga("test_saga", logger)

	step := saga.NewSagaStep("step1",
		func(ctx context.Context, data map[string]interface{}) error {
			return nil
		},
		nil,
	)

	s.AddStep(step)
	assert.Len(t, s.Steps, 1)
}

func TestSaga_StateTransitions(t *testing.T) {
	logger := zap.NewNop()
	s := saga.NewSaga("test_saga", logger)

	s.SetState(saga.SagaStateRunning)
	assert.Equal(t, saga.SagaStateRunning, s.State)

	s.SetState(saga.SagaStateCompleted)
	assert.Equal(t, saga.SagaStateCompleted, s.State)
	assert.NotNil(t, s.CompletedAt)
}

func TestSaga_MarkCompleted(t *testing.T) {
	logger := zap.NewNop()
	s := saga.NewSaga("test_saga", logger)

	s.MarkCompleted("step1")
	s.MarkCompleted("step2")

	assert.True(t, s.IsStepCompleted("step1"))
	assert.True(t, s.IsStepCompleted("step2"))
	assert.False(t, s.IsStepCompleted("step3"))
}

func TestSaga_ContextManagement(t *testing.T) {
	logger := zap.NewNop()
	s := saga.NewSaga("test_saga", logger)

	s.SetResult("key1", "value1")
	s.SetResult("key2", 42)

	val, ok := s.GetResult("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", val)

	val, ok = s.GetResult("key2")
	assert.True(t, ok)
	assert.Equal(t, 42, val)

	_, ok = s.GetResult("nonexistent")
	assert.False(t, ok)
}

func TestSagaStep_Execute(t *testing.T) {
	executed := false
	step := saga.NewSagaStep("test_step",
		func(ctx context.Context, data map[string]interface{}) error {
			executed = true
			data["result"] = "success"
			return nil
		},
		nil,
	)

	ctx := context.Background()
	data := make(map[string]interface{})

	err := step.Execute(ctx, data)
	assert.NoError(t, err)
	assert.True(t, executed)
	assert.Equal(t, "success", data["result"])
}

func TestSagaStep_Compensate(t *testing.T) {
	compensated := false
	step := saga.NewSagaStep("test_step",
		func(ctx context.Context, data map[string]interface{}) error {
			data["executed"] = true
			return nil
		},
		func(ctx context.Context, data map[string]interface{}) error {
			compensated = true
			delete(data, "executed")
			return nil
		},
	)

	ctx := context.Background()
	data := make(map[string]interface{})

	// Execute
	err := step.Execute(ctx, data)
	assert.NoError(t, err)
	assert.True(t, data["executed"].(bool))

	// Compensate
	err = step.Compensate(ctx, data)
	assert.NoError(t, err)
	assert.True(t, compensated)
	_, exists := data["executed"]
	assert.False(t, exists)
}

func TestSagaStep_WithRetry(t *testing.T) {
	step := saga.NewSagaStep("test_step",
		func(ctx context.Context, data map[string]interface{}) error {
			return nil
		},
		nil,
	).WithRetry(3)

	assert.True(t, step.IsRetryable())
	assert.Equal(t, 3, step.MaxRetries())
}

func TestSagaStep_WithTimeout(t *testing.T) {
	timeout := 10 * time.Second
	step := saga.NewSagaStep("test_step",
		func(ctx context.Context, data map[string]interface{}) error {
			return nil
		},
		nil,
	).WithTimeout(timeout)

	assert.Equal(t, timeout, step.Timeout())
}

func TestSagaStep_HasCompensation(t *testing.T) {
	stepWithComp := saga.NewSagaStep("with_comp",
		func(ctx context.Context, data map[string]interface{}) error {
			return nil
		},
		func(ctx context.Context, data map[string]interface{}) error {
			return nil
		},
	)

	stepWithoutComp := saga.NewSagaStep("without_comp",
		func(ctx context.Context, data map[string]interface{}) error {
			return nil
		},
		nil,
	)

	assert.True(t, stepWithComp.HasCompensation())
	assert.False(t, stepWithoutComp.HasCompensation())
}
