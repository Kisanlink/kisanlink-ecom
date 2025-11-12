// Package gst provides GST number validation and management services for Indian GST numbers.
// It includes comprehensive validation, checksum verification, and deduplication checks.
package gst

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Service handles GST number validation and deduplication
type Service struct {
	db        *gorm.DB
	validator *Validator
	logger    *logrus.Logger
}

// NewService creates a new GST service
func NewService(db *gorm.DB, logger *logrus.Logger) *Service {
	return &Service{
		db:        db,
		validator: NewValidator(),
		logger:    logger,
	}
}

// ValidateAndNormalize validates a GST number and returns the normalized version
func (s *Service) ValidateAndNormalize(gst string) (string, error) {
	// Normalize first
	normalized := s.validator.Normalize(gst)

	// Validate
	if err := s.validator.Validate(normalized); err != nil {
		s.logger.WithError(err).WithField("gst", gst).Warn("GST validation failed")
		return "", fmt.Errorf("invalid GST number: %w", err)
	}

	return normalized, nil
}

// CheckGSTExists checks if a GST number is already registered
// This is used for deduplication to prevent multiple collaborators with same GST
func (s *Service) CheckGSTExists(ctx context.Context, gst string) (bool, error) {
	// Normalize GST first
	normalized, err := s.ValidateAndNormalize(gst)
	if err != nil {
		return false, err
	}

	// Check if GST exists in collaborators table
	var count int64
	err = s.db.WithContext(ctx).
		Table("collaborators").
		Where("tax_id = ? AND deleted_at IS NULL", normalized).
		Count(&count).Error

	if err != nil {
		s.logger.WithError(err).WithField("gst", normalized).Error("Failed to check GST existence")
		return false, fmt.Errorf("failed to check GST existence: %w", err)
	}

	return count > 0, nil
}

// Info contains extracted information from a GST number
type Info struct {
	GST        string
	StateCode  string
	StateName  string
	PAN        string
	IsValid    bool
	EntityType string
}

// GetInfo extracts and returns information from a GST number
func (s *Service) GetInfo(gst string) (*Info, error) {
	normalized := s.validator.Normalize(gst)

	info := &Info{
		GST:     normalized,
		IsValid: s.validator.IsValid(normalized),
	}

	if !info.IsValid {
		return info, fmt.Errorf("invalid GST number")
	}

	// Extract state code
	stateCode, err := s.validator.ExtractStateCode(normalized)
	if err != nil {
		return info, err
	}
	info.StateCode = stateCode
	info.StateName = s.validator.GetStateName(stateCode)

	// Extract PAN
	pan, err := s.validator.ExtractPAN(normalized)
	if err != nil {
		return info, err
	}
	info.PAN = pan

	// Determine entity type from PAN
	if len(pan) >= 4 {
		entityTypeChar := pan[3]
		switch entityTypeChar {
		case 'C':
			info.EntityType = "Company"
		case 'F':
			info.EntityType = "Firm"
		case 'H':
			info.EntityType = "Hindu Undivided Family"
		case 'A':
			info.EntityType = "Association of Persons"
		case 'P':
			info.EntityType = "Individual"
		case 'L':
			info.EntityType = "Local Authority"
		case 'J':
			info.EntityType = "Artificial Juridical Person"
		case 'T':
			info.EntityType = "Trust"
		case 'B':
			info.EntityType = "Body of Individuals"
		case 'G':
			info.EntityType = "Government"
		case 'S':
			info.EntityType = "Others"
		default:
			info.EntityType = "Unknown"
		}
	}

	return info, nil
}
