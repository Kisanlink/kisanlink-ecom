package saga

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SagaStorage interface for persisting saga state
type SagaStorage interface {
	// SaveState persists the current saga state
	SaveState(ctx context.Context, saga *Saga) error

	// LoadState retrieves saga state by ID
	LoadState(ctx context.Context, sagaID string) (*Saga, error)

	// ListPendingSagas retrieves all incomplete sagas
	ListPendingSagas(ctx context.Context) ([]*Saga, error)

	// CleanupCompleted removes old completed sagas
	CleanupCompleted(ctx context.Context, olderThan time.Duration) error
}

// SagaStateRecord represents the database model for saga state
type SagaStateRecord struct {
	ID             string         `gorm:"primaryKey;type:varchar(255)"`
	Name           string         `gorm:"type:varchar(255);not null;index"`
	State          string         `gorm:"type:varchar(50);not null;index"`
	Context        datatypes.JSON `gorm:"type:jsonb"`
	CompletedSteps datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt      time.Time      `gorm:"not null;index"`
	UpdatedAt      time.Time      `gorm:"not null"`
	CompletedAt    *time.Time     `gorm:"index"`
}

// TableName specifies the table name for SagaStateRecord
func (SagaStateRecord) TableName() string {
	return "saga_states"
}

// DBSagaStorage implements SagaStorage using GORM
type DBSagaStorage struct {
	db *gorm.DB
}

// NewDBSagaStorage creates a new database-backed saga storage
func NewDBSagaStorage(db *gorm.DB) SagaStorage {
	return &DBSagaStorage{db: db}
}

// SaveState persists the current saga state
func (s *DBSagaStorage) SaveState(ctx context.Context, saga *Saga) error {
	contextJSON, err := json.Marshal(saga.Context)
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	stepsJSON, err := json.Marshal(saga.CompletedSteps)
	if err != nil {
		return fmt.Errorf("failed to marshal completed steps: %w", err)
	}

	record := &SagaStateRecord{
		ID:             saga.ID,
		Name:           saga.Name,
		State:          string(saga.State),
		Context:        contextJSON,
		CompletedSteps: stepsJSON,
		CreatedAt:      saga.CreatedAt,
		UpdatedAt:      saga.UpdatedAt,
		CompletedAt:    saga.CompletedAt,
	}

	return s.db.WithContext(ctx).Save(record).Error
}

// LoadState retrieves saga state by ID
func (s *DBSagaStorage) LoadState(ctx context.Context, sagaID string) (*Saga, error) {
	var record SagaStateRecord

	err := s.db.WithContext(ctx).First(&record, "id = ?", sagaID).Error
	if err != nil {
		return nil, fmt.Errorf("failed to load saga state: %w", err)
	}

	saga := &Saga{
		ID:          record.ID,
		Name:        record.Name,
		State:       SagaState(record.State),
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
		CompletedAt: record.CompletedAt,
	}

	// Unmarshal context
	if len(record.Context) > 0 {
		if err := json.Unmarshal(record.Context, &saga.Context); err != nil {
			return nil, fmt.Errorf("failed to unmarshal context: %w", err)
		}
	} else {
		saga.Context = make(map[string]interface{})
	}

	// Unmarshal completed steps
	if len(record.CompletedSteps) > 0 {
		if err := json.Unmarshal(record.CompletedSteps, &saga.CompletedSteps); err != nil {
			return nil, fmt.Errorf("failed to unmarshal completed steps: %w", err)
		}
	} else {
		saga.CompletedSteps = make([]string, 0)
	}

	return saga, nil
}

// ListPendingSagas retrieves all incomplete sagas
func (s *DBSagaStorage) ListPendingSagas(ctx context.Context) ([]*Saga, error) {
	var records []SagaStateRecord

	// Find sagas that are running or compensating
	err := s.db.WithContext(ctx).
		Where("state IN ?", []string{
			string(SagaStateRunning),
			string(SagaStateCompensating),
		}).
		Order("created_at ASC").
		Find(&records).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list pending sagas: %w", err)
	}

	sagas := make([]*Saga, 0, len(records))
	for _, record := range records {
		saga, err := s.recordToSaga(&record)
		if err != nil {
			// Log error but continue processing other sagas
			continue
		}
		sagas = append(sagas, saga)
	}

	return sagas, nil
}

// CleanupCompleted removes old completed sagas
func (s *DBSagaStorage) CleanupCompleted(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)

	result := s.db.WithContext(ctx).
		Where("state IN ? AND completed_at < ?",
			[]string{
				string(SagaStateCompleted),
				string(SagaStateCompensated),
				string(SagaStateFailed),
			},
			cutoff).
		Delete(&SagaStateRecord{})

	if result.Error != nil {
		return fmt.Errorf("failed to cleanup completed sagas: %w", result.Error)
	}

	return nil
}

// recordToSaga converts a SagaStateRecord to a Saga
func (s *DBSagaStorage) recordToSaga(record *SagaStateRecord) (*Saga, error) {
	saga := &Saga{
		ID:          record.ID,
		Name:        record.Name,
		State:       SagaState(record.State),
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
		CompletedAt: record.CompletedAt,
	}

	// Unmarshal context
	if len(record.Context) > 0 {
		if err := json.Unmarshal(record.Context, &saga.Context); err != nil {
			return nil, fmt.Errorf("failed to unmarshal context: %w", err)
		}
	} else {
		saga.Context = make(map[string]interface{})
	}

	// Unmarshal completed steps
	if len(record.CompletedSteps) > 0 {
		if err := json.Unmarshal(record.CompletedSteps, &saga.CompletedSteps); err != nil {
			return nil, fmt.Errorf("failed to unmarshal completed steps: %w", err)
		}
	} else {
		saga.CompletedSteps = make([]string, 0)
	}

	return saga, nil
}
