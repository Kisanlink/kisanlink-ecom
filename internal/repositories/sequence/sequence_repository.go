// Package sequence provides repository operations for managing sequence counters.
package sequence

import (
	"context"
	"fmt"

	"kisanlink-ecom/entities/models/common"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SequenceRepositoryInterface defines the interface for sequence counter repository operations
//
//nolint:revive // Keeping full name for clarity when used outside package
type SequenceRepositoryInterface interface {
	// GetNextValue atomically increments and returns the next sequence value
	GetNextValue(ctx context.Context, sequenceKey string, orgID *string) (int64, error)

	// GetOrCreateSequence gets an existing sequence or creates a new one
	GetOrCreateSequence(ctx context.Context, sequenceKey string, orgID *string, resetFrequency string) (*common.SequenceCounter, error)

	// ResetSequence resets a sequence counter to 0
	ResetSequence(ctx context.Context, sequenceKey string, orgID *string) error
}

// SequenceRepository provides database operations for sequence counters
//
//nolint:revive // Keeping full name for clarity when used outside package
type SequenceRepository struct {
	dbManager db.DBManager
}

// NewSequenceRepository creates a new sequence repository
func NewSequenceRepository(dbManager db.DBManager) *SequenceRepository {
	return &SequenceRepository{
		dbManager: dbManager,
	}
}

// GetNextValue atomically increments and returns the next sequence value
// Uses PostgreSQL row-level locking to ensure thread-safety
func (r *SequenceRepository) GetNextValue(ctx context.Context, sequenceKey string, orgID *string) (int64, error) {
	// Cast to PostgresManager to access GetDB method
	postgresManager, ok := r.dbManager.(*db.PostgresManager)
	if !ok {
		return 0, fmt.Errorf("failed to cast to PostgresManager")
	}

	database, err := postgresManager.GetDB(ctx, false)
	if err != nil {
		return 0, fmt.Errorf("failed to get database connection: %w", err)
	}

	var sequence common.SequenceCounter

	// Use transaction with row-level locking to ensure atomicity
	err = database.Transaction(func(tx *gorm.DB) error {
		// Lock the row for update
		query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("sequence_key = ?", sequenceKey)

		if orgID != nil {
			query = query.Where("organization_id = ?", *orgID)
		} else {
			query = query.Where("organization_id IS NULL")
		}

		if err := query.First(&sequence).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create new sequence if it doesn't exist
				sequence = *common.NewSequenceCounter(sequenceKey, orgID, "never")
				sequence.CurrentValue = 1
				if err := tx.Create(&sequence).Error; err != nil {
					return fmt.Errorf("failed to create sequence: %w", err)
				}
				return nil
			}
			return fmt.Errorf("failed to get sequence: %w", err)
		}

		// Check if sequence should be reset
		if sequence.ShouldReset() {
			sequence.Reset()
		}

		// Increment the sequence
		sequence.Increment()

		// Update the sequence in the database
		if err := tx.Model(&sequence).Updates(map[string]interface{}{
			"current_value": sequence.CurrentValue,
			"last_reset_at": sequence.LastResetAt,
		}).Error; err != nil {
			return fmt.Errorf("failed to update sequence: %w", err)
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return sequence.CurrentValue, nil
}

// GetOrCreateSequence gets an existing sequence or creates a new one
func (r *SequenceRepository) GetOrCreateSequence(ctx context.Context, sequenceKey string, orgID *string, resetFrequency string) (*common.SequenceCounter, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
		Field:    "sequence_key",
		Operator: base.OpEqual,
		Value:    sequenceKey,
	})

	if orgID != nil {
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "organization_id",
			Operator: base.OpEqual,
			Value:    *orgID,
		})
	} else {
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "organization_id",
			Operator: base.OpIsNull,
		})
	}

	var sequences []common.SequenceCounter
	if err := r.dbManager.List(ctx, filter, &sequences); err != nil {
		return nil, fmt.Errorf("failed to get sequence: %w", err)
	}

	if len(sequences) > 0 {
		return &sequences[0], nil
	}

	// Create new sequence
	sequence := common.NewSequenceCounter(sequenceKey, orgID, resetFrequency)
	if err := r.dbManager.Create(ctx, sequence); err != nil {
		return nil, fmt.Errorf("failed to create sequence: %w", err)
	}

	return sequence, nil
}

// ResetSequence resets a sequence counter to 0
func (r *SequenceRepository) ResetSequence(ctx context.Context, sequenceKey string, orgID *string) error {
	sequence, err := r.GetOrCreateSequence(ctx, sequenceKey, orgID, "never")
	if err != nil {
		return err
	}

	sequence.Reset()

	if err := r.dbManager.Update(ctx, sequence); err != nil {
		return fmt.Errorf("failed to reset sequence: %w", err)
	}

	return nil
}
