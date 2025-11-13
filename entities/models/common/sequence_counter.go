package common

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// SequenceCounter tracks sequence numbers for various entity types
// Used for generating unique lot numbers, batch numbers, and entity IDs
type SequenceCounter struct {
	base.BaseModel

	// Unique key for this sequence (e.g., "lot_ORG123_20241024", "batch_20241024")
	SequenceKey string `json:"sequence_key" gorm:"type:varchar(255);not null;uniqueIndex:idx_sequence_key_org,priority:1"`

	// Organization ID for organization-scoped sequences (NULL for global sequences)
	OrganizationID *string `json:"organization_id" gorm:"type:varchar(255);uniqueIndex:idx_sequence_key_org,priority:2"`

	// Current sequence value
	CurrentValue int64 `json:"current_value" gorm:"type:bigint;not null;default:0"`

	// Optional prefix for formatted output (e.g., "ABC" for organization prefix)
	Prefix *string `json:"prefix" gorm:"type:varchar(50)"`

	// Format pattern for this sequence (e.g., "LOT-{PREFIX}-{DATE}-{SEQ}")
	FormatPattern *string `json:"format_pattern" gorm:"type:varchar(100)"`

	// Reset frequency: "daily", "monthly", "yearly", "never"
	ResetFrequency string `json:"reset_frequency" gorm:"type:varchar(20);not null;default:'never'"`

	// Last time this sequence was reset
	LastResetAt *time.Time `json:"last_reset_at" gorm:"type:timestamp"`

	// Soft delete support
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index:idx_sequence_counters_deleted"`
}

// TableName returns the table name for GORM
func (SequenceCounter) TableName() string {
	return "sequence_counters"
}

// NewSequenceCounter creates a new sequence counter
func NewSequenceCounter(sequenceKey string, orgID *string, resetFrequency string) *SequenceCounter {
	now := time.Now()
	return &SequenceCounter{
		BaseModel:      *base.NewBaseModel("SEQ", "small"),
		SequenceKey:    sequenceKey,
		OrganizationID: orgID,
		CurrentValue:   0,
		ResetFrequency: resetFrequency,
		LastResetAt:    &now,
	}
}

// ShouldReset checks if the sequence should be reset based on reset frequency
func (s *SequenceCounter) ShouldReset() bool {
	if s.LastResetAt == nil {
		return false
	}

	now := time.Now()

	switch s.ResetFrequency {
	case "daily":
		return s.LastResetAt.YearDay() != now.YearDay() || s.LastResetAt.Year() != now.Year()
	case "monthly":
		return s.LastResetAt.Month() != now.Month() || s.LastResetAt.Year() != now.Year()
	case "yearly":
		return s.LastResetAt.Year() != now.Year()
	case "never":
		return false
	default:
		return false
	}
}

// Reset resets the sequence counter to 0
func (s *SequenceCounter) Reset() {
	s.CurrentValue = 0
	now := time.Now()
	s.LastResetAt = &now
}

// Increment increments and returns the next sequence value
func (s *SequenceCounter) Increment() int64 {
	s.CurrentValue++
	return s.CurrentValue
}
