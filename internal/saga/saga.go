package saga

import (
	"time"

	"go.uber.org/zap"
)

// SagaState represents the current state of a saga
type SagaState string

const (
	SagaStatePending      SagaState = "pending"
	SagaStateRunning      SagaState = "running"
	SagaStateCompleted    SagaState = "completed"
	SagaStateFailed       SagaState = "failed"
	SagaStateCompensating SagaState = "compensating"
	SagaStateCompensated  SagaState = "compensated"
)

// Saga represents a distributed transaction with compensation
type Saga struct {
	ID             string
	Name           string
	Steps          []Step
	State          SagaState
	Context        map[string]interface{}
	CompletedSteps []string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	CompletedAt    *time.Time
	logger         *zap.Logger
}

// NewSaga creates a new saga instance
func NewSaga(name string, logger *zap.Logger) *Saga {
	now := time.Now()
	return &Saga{
		ID:             generateSagaID(),
		Name:           name,
		Steps:          make([]Step, 0),
		State:          SagaStatePending,
		Context:        make(map[string]interface{}),
		CompletedSteps: make([]string, 0),
		CreatedAt:      now,
		UpdatedAt:      now,
		logger:         logger,
	}
}

// AddStep adds a step to the saga
func (s *Saga) AddStep(step Step) {
	s.Steps = append(s.Steps, step)
}

// GetContext returns the saga context
func (s *Saga) GetContext() map[string]interface{} {
	return s.Context
}

// SetContext sets the saga context
func (s *Saga) SetContext(ctx map[string]interface{}) {
	s.Context = ctx
}

// MarkCompleted marks a step as completed
func (s *Saga) MarkCompleted(stepName string) {
	s.CompletedSteps = append(s.CompletedSteps, stepName)
	s.UpdatedAt = time.Now()
}

// IsStepCompleted checks if a step has been completed
func (s *Saga) IsStepCompleted(stepName string) bool {
	for _, completed := range s.CompletedSteps {
		if completed == stepName {
			return true
		}
	}
	return false
}

// SetState updates the saga state
func (s *Saga) SetState(state SagaState) {
	s.State = state
	s.UpdatedAt = time.Now()

	if state == SagaStateCompleted || state == SagaStateCompensated || state == SagaStateFailed {
		now := time.Now()
		s.CompletedAt = &now
	}
}

// GetResult retrieves a result from the saga context
func (s *Saga) GetResult(key string) (interface{}, bool) {
	val, ok := s.Context[key]
	return val, ok
}

// SetResult stores a result in the saga context
func (s *Saga) SetResult(key string, value interface{}) {
	s.Context[key] = value
	s.UpdatedAt = time.Now()
}

// generateSagaID generates a unique saga identifier
func generateSagaID() string {
	return "saga_" + time.Now().Format("20060102150405") + "_" + randomString(8)
}

// randomString generates a random alphanumeric string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
