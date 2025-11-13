package saga

import (
	"sync"
	"time"
)

// SagaMetrics tracks saga execution metrics
type SagaMetrics struct {
	sagasStarted           map[string]int64
	sagasCompleted         map[string]int64
	sagasFailed            map[string]int64
	sagasCompensated       map[string]int64
	stepsExecuted          map[string]int64
	stepsFailed            map[string]int64
	compensationsSucceeded map[string]int64
	compensationsFailed    map[string]int64
	sagaDurations          map[string][]time.Duration
	mu                     sync.RWMutex
}

// NewSagaMetrics creates a new saga metrics instance
func NewSagaMetrics() *SagaMetrics {
	return &SagaMetrics{
		sagasStarted:           make(map[string]int64),
		sagasCompleted:         make(map[string]int64),
		sagasFailed:            make(map[string]int64),
		sagasCompensated:       make(map[string]int64),
		stepsExecuted:          make(map[string]int64),
		stepsFailed:            make(map[string]int64),
		compensationsSucceeded: make(map[string]int64),
		compensationsFailed:    make(map[string]int64),
		sagaDurations:          make(map[string][]time.Duration),
	}
}

// RecordSagaStart records a saga start
func (m *SagaMetrics) RecordSagaStart(sagaName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sagasStarted[sagaName]++
}

// RecordSagaSuccess records a successful saga completion
func (m *SagaMetrics) RecordSagaSuccess(sagaName string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sagasCompleted[sagaName]++
	m.sagaDurations[sagaName] = append(m.sagaDurations[sagaName], duration)
}

// RecordSagaFailure records a saga failure
func (m *SagaMetrics) RecordSagaFailure(sagaName string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sagasFailed[sagaName]++
	m.sagaDurations[sagaName] = append(m.sagaDurations[sagaName], duration)
}

// RecordSagaCompensated records a compensated saga
func (m *SagaMetrics) RecordSagaCompensated(sagaName string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sagasCompensated[sagaName]++
	m.sagaDurations[sagaName] = append(m.sagaDurations[sagaName], duration)
}

// RecordStepSuccess records a successful step execution
func (m *SagaMetrics) RecordStepSuccess(sagaName, stepName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := sagaName + ":" + stepName
	m.stepsExecuted[key]++
}

// RecordStepFailure records a step failure
func (m *SagaMetrics) RecordStepFailure(sagaName, stepName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := sagaName + ":" + stepName
	m.stepsFailed[key]++
}

// RecordCompensationSuccess records a successful compensation
func (m *SagaMetrics) RecordCompensationSuccess(sagaName, stepName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := sagaName + ":" + stepName
	m.compensationsSucceeded[key]++
}

// RecordCompensationFailure records a compensation failure
func (m *SagaMetrics) RecordCompensationFailure(sagaName, stepName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := sagaName + ":" + stepName
	m.compensationsFailed[key]++
}

// GetSagaStats returns statistics for a specific saga
func (m *SagaMetrics) GetSagaStats(sagaName string) SagaStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return SagaStats{
		Started:     m.sagasStarted[sagaName],
		Completed:   m.sagasCompleted[sagaName],
		Failed:      m.sagasFailed[sagaName],
		Compensated: m.sagasCompensated[sagaName],
		Durations:   m.sagaDurations[sagaName],
	}
}

// SagaStats contains statistics for a saga
type SagaStats struct {
	Started     int64
	Completed   int64
	Failed      int64
	Compensated int64
	Durations   []time.Duration
}

// SuccessRate calculates the success rate of a saga
func (s *SagaStats) SuccessRate() float64 {
	if s.Started == 0 {
		return 0
	}
	return float64(s.Completed) / float64(s.Started)
}

// AverageDuration calculates the average duration
func (s *SagaStats) AverageDuration() time.Duration {
	if len(s.Durations) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range s.Durations {
		total += d
	}
	return total / time.Duration(len(s.Durations))
}
