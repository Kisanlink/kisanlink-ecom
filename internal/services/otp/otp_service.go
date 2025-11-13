// Package otp provides OTP generation, validation, and delivery services
package otp

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Service manages OTP generation, validation, and delivery
type Service interface {
	// Generate creates a new OTP for a user and purpose
	Generate(ctx context.Context, userID, purpose string) (string, error)

	// Validate checks if an OTP is valid
	Validate(ctx context.Context, userID, purpose, otp string) error

	// SendOTP sends the OTP to the user (email/SMS)
	SendOTP(ctx context.Context, userID, otp string) error
}

// otpService implements Service
type otpService struct {
	storage     OTPStorage
	rateLimiter RateLimiter
	notifier    OTPNotifier
	logger      *zap.Logger
	otpLength   int
	otpTTL      time.Duration
}

// NewOTPService creates a new OTP service
func NewOTPService(logger *zap.Logger) Service {
	return &otpService{
		storage:     NewInMemoryOTPStorage(),
		rateLimiter: NewInMemoryRateLimiter(),
		notifier:    NewMockOTPNotifier(logger),
		logger:      logger,
		otpLength:   6,
		otpTTL:      10 * time.Minute,
	}
}

// Generate creates a new OTP
func (s *otpService) Generate(ctx context.Context, userID, purpose string) (string, error) {
	// Check rate limit
	if err := s.rateLimiter.CheckGenerationLimit(ctx, userID); err != nil {
		s.logger.Warn("OTP generation rate limit exceeded",
			zap.String("user_id", userID),
			zap.Error(err))
		return "", fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Generate secure random OTP
	otp, err := generateSecureOTP(s.otpLength)
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Store OTP
	expiresAt := time.Now().Add(s.otpTTL)
	if err := s.storage.Store(ctx, userID, purpose, otp, expiresAt); err != nil {
		return "", fmt.Errorf("failed to store OTP: %w", err)
	}

	// Send OTP
	if err := s.notifier.Send(ctx, userID, otp, purpose); err != nil {
		s.logger.Error("failed to send OTP",
			zap.String("user_id", userID),
			zap.String("purpose", purpose),
			zap.Error(err))
		// Don't fail - OTP is generated
	}

	s.logger.Info("OTP generated",
		zap.String("user_id", userID),
		zap.String("purpose", purpose))

	return otp, nil
}

// Validate checks if an OTP is valid
func (s *otpService) Validate(ctx context.Context, userID, purpose, otp string) error {
	// Check validation rate limit
	if err := s.rateLimiter.CheckValidationLimit(ctx, userID); err != nil {
		s.logger.Warn("OTP validation rate limit exceeded",
			zap.String("user_id", userID),
			zap.Error(err))
		return fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Validate OTP
	if err := s.storage.Validate(ctx, userID, purpose, otp); err != nil {
		s.logger.Warn("OTP validation failed",
			zap.String("user_id", userID),
			zap.String("purpose", purpose),
			zap.Error(err))
		return err
	}

	// Delete OTP after successful validation (one-time use)
	if err := s.storage.Delete(ctx, userID, purpose); err != nil {
		s.logger.Warn("failed to delete used OTP",
			zap.String("user_id", userID),
			zap.Error(err))
	}

	s.logger.Info("OTP validated successfully",
		zap.String("user_id", userID),
		zap.String("purpose", purpose))

	return nil
}

// SendOTP sends an OTP to the user
func (s *otpService) SendOTP(ctx context.Context, userID, otp string) error {
	return s.notifier.Send(ctx, userID, otp, "")
}

// generateSecureOTP generates a cryptographically secure OTP
func generateSecureOTP(length int) (string, error) {
	const digits = "0123456789"
	otp := make([]byte, length)

	for i := range otp {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		otp[i] = digits[num.Int64()]
	}

	return string(otp), nil
}

// OTPStorage interface for storing OTPs
//
//nolint:revive // OTP prefix intentional for clarity
type OTPStorage interface {
	Store(ctx context.Context, userID, purpose, otp string, expiresAt time.Time) error
	Validate(ctx context.Context, userID, purpose, otp string) error
	Delete(ctx context.Context, userID, purpose string) error
}

// inMemoryOTPStorage is a simple in-memory storage for OTPs
type inMemoryOTPStorage struct {
	otps map[string]*OTPRecord
	mu   sync.RWMutex
}

// OTPRecord represents a stored OTP
//
//nolint:revive // OTP prefix intentional for clarity
type OTPRecord struct {
	OTP       string
	ExpiresAt time.Time
	CreatedAt time.Time
	Attempts  int
}

// NewInMemoryOTPStorage creates a new in-memory OTP storage
func NewInMemoryOTPStorage() OTPStorage {
	return &inMemoryOTPStorage{
		otps: make(map[string]*OTPRecord),
	}
}

func (s *inMemoryOTPStorage) Store(_ context.Context, userID, purpose, otp string, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s:%s", userID, purpose)
	s.otps[key] = &OTPRecord{
		OTP:       otp,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		Attempts:  0,
	}

	return nil
}

func (s *inMemoryOTPStorage) Validate(_ context.Context, userID, purpose, otp string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s:%s", userID, purpose)
	record, exists := s.otps[key]

	if !exists {
		return fmt.Errorf("OTP not found")
	}

	// Check expiration
	if time.Now().After(record.ExpiresAt) {
		delete(s.otps, key)
		return fmt.Errorf("OTP expired")
	}

	// Check attempts
	record.Attempts++
	if record.Attempts > 3 {
		delete(s.otps, key)
		return fmt.Errorf("too many validation attempts")
	}

	// Validate OTP
	if record.OTP != otp {
		return fmt.Errorf("invalid OTP")
	}

	return nil
}

func (s *inMemoryOTPStorage) Delete(_ context.Context, userID, purpose string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s:%s", userID, purpose)
	delete(s.otps, key)

	return nil
}

// RateLimiter interface for rate limiting
type RateLimiter interface {
	CheckGenerationLimit(ctx context.Context, userID string) error
	CheckValidationLimit(ctx context.Context, userID string) error
}

// inMemoryRateLimiter is a simple in-memory rate limiter
type inMemoryRateLimiter struct {
	generation map[string][]time.Time
	validation map[string][]time.Time
	mu         sync.RWMutex
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter
func NewInMemoryRateLimiter() RateLimiter {
	return &inMemoryRateLimiter{
		generation: make(map[string][]time.Time),
		validation: make(map[string][]time.Time),
	}
}

func (r *inMemoryRateLimiter) CheckGenerationLimit(_ context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	window := now.Add(-1 * time.Hour)

	// Clean old entries
	timestamps := r.generation[userID]
	validTimestamps := make([]time.Time, 0)
	for _, ts := range timestamps {
		if ts.After(window) {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	// Check limit (5 per hour)
	if len(validTimestamps) >= 5 {
		return fmt.Errorf("generation rate limit exceeded: max 5 per hour")
	}

	// Add current timestamp
	validTimestamps = append(validTimestamps, now)
	r.generation[userID] = validTimestamps

	return nil
}

func (r *inMemoryRateLimiter) CheckValidationLimit(_ context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	window := now.Add(-1 * time.Hour)

	// Clean old entries
	timestamps := r.validation[userID]
	validTimestamps := make([]time.Time, 0)
	for _, ts := range timestamps {
		if ts.After(window) {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	// Check limit (10 per hour)
	if len(validTimestamps) >= 10 {
		return fmt.Errorf("validation rate limit exceeded: max 10 per hour")
	}

	// Add current timestamp
	validTimestamps = append(validTimestamps, now)
	r.validation[userID] = validTimestamps

	return nil
}

// OTPNotifier interface for sending OTPs
//
//nolint:revive // OTP prefix intentional for clarity
type OTPNotifier interface {
	Send(ctx context.Context, userID, otp, purpose string) error
}

// mockOTPNotifier is a mock OTP notifier for testing
type mockOTPNotifier struct {
	logger *zap.Logger
}

// NewMockOTPNotifier creates a new mock OTP notifier
func NewMockOTPNotifier(logger *zap.Logger) OTPNotifier {
	return &mockOTPNotifier{logger: logger}
}

func (n *mockOTPNotifier) Send(_ context.Context, userID, otp, purpose string) error {
	n.logger.Info("OTP sent (mock)",
		zap.String("user_id", userID),
		zap.String("otp", otp),
		zap.String("purpose", purpose))
	return nil
}
