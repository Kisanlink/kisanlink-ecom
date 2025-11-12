// Package gst provides GST number validation and management services for Indian GST numbers.
// It includes comprehensive validation, checksum verification, and deduplication checks.
package gst

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Service handles GST number validation and deduplication
type Service struct {
	db        *gorm.DB
	validator *Validator
	lockMgr   *DistributedLock
	logger    *logrus.Logger
}

// NewService creates a new GST service
func NewService(db *gorm.DB, redisClient redis.UniversalClient, logger *logrus.Logger) *Service {
	return &Service{
		db:        db,
		validator: NewValidator(),
		lockMgr:   NewDistributedLock(redisClient, logger),
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
// Deprecated: Use CheckAndReserveGST instead for race-condition-free checking
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

// CheckAndReserveGST acquires a distributed lock, checks if GST exists, and reserves it if not
// This prevents race conditions where multiple FPOs create duplicate master collaborators
func (s *Service) CheckAndReserveGST(ctx context.Context, gst string, fpoID uint64, requestID string) (bool, error) {
	// Normalize and validate GST first
	normalized, err := s.ValidateAndNormalize(gst)
	if err != nil {
		return false, fmt.Errorf("GST validation failed: %w", err)
	}

	// Create lock key
	lockKey := fmt.Sprintf("gst:lock:%s", normalized)

	// Acquire distributed lock with retry
	lock, err := s.lockMgr.AcquireLock(ctx, LockOptions{
		Key:        lockKey,
		TTL:        30 * time.Second,
		RetryDelay: 100 * time.Millisecond,
		MaxRetries: 10,
		Owner:      fmt.Sprintf("fpo:%d:req:%s", fpoID, requestID),
	})

	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"gst":        normalized,
			"fpo_id":     fpoID,
			"request_id": requestID,
		}).Error("Failed to acquire GST lock")
		return false, fmt.Errorf("failed to acquire GST lock: %w", err)
	}

	defer func() {
		if err := lock.Release(ctx); err != nil {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"gst":        normalized,
				"fpo_id":     fpoID,
				"request_id": requestID,
			}).Error("Failed to release GST lock")
		}
	}()

	// Check if GST exists while holding the lock
	var count int64
	err = s.db.WithContext(ctx).
		Table("collaborators").
		Where("tax_id = ? AND deleted_at IS NULL", normalized).
		Count(&count).Error

	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"gst":        normalized,
			"fpo_id":     fpoID,
			"request_id": requestID,
		}).Error("Failed to check GST existence")
		return false, fmt.Errorf("failed to check GST existence: %w", err)
	}

	// If GST exists, return true (indicating it's already registered)
	if count > 0 {
		s.logger.WithFields(logrus.Fields{
			"gst":        normalized,
			"fpo_id":     fpoID,
			"request_id": requestID,
		}).Warn("GST already exists")
		return true, nil
	}

	// GST does not exist - it's reserved by this lock holder
	s.logger.WithFields(logrus.Fields{
		"gst":        normalized,
		"fpo_id":     fpoID,
		"request_id": requestID,
	}).Info("GST reserved successfully")

	return false, nil
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
