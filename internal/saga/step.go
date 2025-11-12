package saga

import (
	"context"
	"time"
)

// StepFunc is a function that executes a step
type StepFunc func(ctx context.Context, data map[string]interface{}) error

// Step represents a single step in a saga
type Step interface {
	// Execute runs the step's main logic
	Execute(ctx context.Context, data map[string]interface{}) error

	// Compensate reverses the step's effects
	Compensate(ctx context.Context, data map[string]interface{}) error

	// Name returns the step's identifier
	Name() string

	// IsRetryable indicates if the step can be retried on failure
	IsRetryable() bool

	// MaxRetries returns the maximum number of retry attempts
	MaxRetries() int

	// Timeout returns the step execution timeout
	Timeout() time.Duration

	// HasCompensation returns true if the step has a compensation handler
	HasCompensation() bool
}

// SagaStep is a concrete implementation of Step
type SagaStep struct {
	StepName       string
	ExecuteFunc    StepFunc
	CompensateFunc StepFunc
	Retryable      bool
	MaxAttempts    int
	StepTimeout    time.Duration
}

// NewSagaStep creates a new saga step
func NewSagaStep(name string, execute StepFunc, compensate StepFunc) *SagaStep {
	return &SagaStep{
		StepName:       name,
		ExecuteFunc:    execute,
		CompensateFunc: compensate,
		Retryable:      false,
		MaxAttempts:    1,
		StepTimeout:    30 * time.Second,
	}
}

// WithRetry configures the step as retryable
func (s *SagaStep) WithRetry(maxRetries int) *SagaStep {
	s.Retryable = true
	s.MaxAttempts = maxRetries
	return s
}

// WithTimeout sets the step timeout
func (s *SagaStep) WithTimeout(timeout time.Duration) *SagaStep {
	s.StepTimeout = timeout
	return s
}

// Execute runs the step's main logic
func (s *SagaStep) Execute(ctx context.Context, data map[string]interface{}) error {
	if s.ExecuteFunc == nil {
		return nil
	}
	return s.ExecuteFunc(ctx, data)
}

// Compensate reverses the step's effects
func (s *SagaStep) Compensate(ctx context.Context, data map[string]interface{}) error {
	if s.CompensateFunc == nil {
		return nil
	}
	return s.CompensateFunc(ctx, data)
}

// Name returns the step's identifier
func (s *SagaStep) Name() string {
	return s.StepName
}

// IsRetryable indicates if the step can be retried on failure
func (s *SagaStep) IsRetryable() bool {
	return s.Retryable
}

// MaxRetries returns the maximum number of retry attempts
func (s *SagaStep) MaxRetries() int {
	return s.MaxAttempts
}

// Timeout returns the step execution timeout
func (s *SagaStep) Timeout() time.Duration {
	return s.StepTimeout
}

// HasCompensation returns true if the step has a compensation handler
func (s *SagaStep) HasCompensation() bool {
	return s.CompensateFunc != nil
}
