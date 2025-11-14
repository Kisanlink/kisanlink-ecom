// Package sequence provides business logic for generating unique IDs using sequence counters.
package sequence

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/common"
)

// SequenceServiceInterface defines the interface for sequence operations
//
//nolint:revive // Keeping full name for clarity when used outside package
type SequenceServiceInterface interface {
	// GenerateID generates a unique ID using sequence counters
	// prefix: entity type prefix (e.g., "ORDER", "ITEM", "PROD")
	// orgID: organization ID for org-scoped sequences (nil for global)
	GenerateID(ctx context.Context, prefix string, orgID *string) (string, error)

	// GetNextValue gets the next sequence value without formatting
	GetNextValue(ctx context.Context, sequenceKey string, orgID *string) (int64, error)

	// ResetSequence resets a sequence counter
	ResetSequence(ctx context.Context, sequenceKey string, orgID *string) error
}

// SequenceRepositoryInterface defines the repository interface
//
//nolint:revive // Keeping full name for clarity when used outside package
type SequenceRepositoryInterface interface {
	GetNextValue(ctx context.Context, sequenceKey string, orgID *string) (int64, error)
	GetOrCreateSequence(ctx context.Context, sequenceKey string, orgID *string, resetFrequency string) (*common.SequenceCounter, error)
	ResetSequence(ctx context.Context, sequenceKey string, orgID *string) error
}

// SequenceService provides business logic for sequence operations
//
//nolint:revive // Keeping full name for clarity when used outside package
type SequenceService struct {
	sequenceRepo SequenceRepositoryInterface
}

// NewSequenceService creates a new sequence service
func NewSequenceService(sequenceRepo SequenceRepositoryInterface) *SequenceService {
	return &SequenceService{
		sequenceRepo: sequenceRepo,
	}
}

// GenerateID generates a unique ID using sequence counters
// Format: {PREFIX}-{TIMESTAMP}-{SEQUENCE}
// Example: ORDER-20241024153045-000001
func (s *SequenceService) GenerateID(ctx context.Context, prefix string, orgID *string) (string, error) {
	// Get current date for the sequence key
	now := time.Now()
	dateStr := now.Format("20060102") // YYYYMMDD

	// Create sequence key: prefix_date (optionally scoped by org)
	sequenceKey := fmt.Sprintf("%s_%s", prefix, dateStr)

	// Get next sequence value (this will auto-create if not exists with daily reset)
	nextValue, err := s.sequenceRepo.GetNextValue(ctx, sequenceKey, orgID)
	if err != nil {
		return "", fmt.Errorf("failed to get next sequence value: %w", err)
	}

	// Format the ID: PREFIX-YYYYMMDDHHMMSS-SEQNUM
	timestamp := now.Format("20060102150405")
	id := fmt.Sprintf("%s-%s-%06d", prefix, timestamp, nextValue)

	return id, nil
}

// GenerateSimpleID generates a simple sequential ID without timestamp
// Format: {PREFIX}-{SEQUENCE}
// Example: ITEM-000001
func (s *SequenceService) GenerateSimpleID(ctx context.Context, prefix string, orgID *string) (string, error) {
	// Create sequence key that doesn't reset (global counter per prefix)
	sequenceKey := fmt.Sprintf("%s_global", prefix)

	// Get next sequence value
	nextValue, err := s.sequenceRepo.GetNextValue(ctx, sequenceKey, orgID)
	if err != nil {
		return "", fmt.Errorf("failed to get next sequence value: %w", err)
	}

	// Format the ID: PREFIX-SEQNUM
	id := fmt.Sprintf("%s-%06d", prefix, nextValue)

	return id, nil
}

// GetNextValue gets the next sequence value without formatting
func (s *SequenceService) GetNextValue(ctx context.Context, sequenceKey string, orgID *string) (int64, error) {
	return s.sequenceRepo.GetNextValue(ctx, sequenceKey, orgID)
}

// ResetSequence resets a sequence counter
func (s *SequenceService) ResetSequence(ctx context.Context, sequenceKey string, orgID *string) error {
	return s.sequenceRepo.ResetSequence(ctx, sequenceKey, orgID)
}
