// Package taxation provides service layer for tax exemption management
package taxation

import (
	"context"
	"fmt"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/taxation"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/shopspring/decimal"
)

// TaxExemptionRepositoryInterface defines repository operations
type TaxExemptionRepositoryInterface interface {
	Create(ctx context.Context, exemption *taxation.TaxExemption) error
	GetByID(ctx context.Context, id string) (*taxation.TaxExemption, error)
	GetByExemptionID(ctx context.Context, exemptionID string) (*taxation.TaxExemption, error)
	GetByOrganizationID(ctx context.Context, orgID string) ([]*taxation.TaxExemption, error)
	Update(ctx context.Context, exemption *taxation.TaxExemption) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter *base.Filter, limit, offset int) ([]*taxation.TaxExemption, int, error)
}

// TaxExemptionServiceInterface defines service operations
type TaxExemptionServiceInterface interface {
	CreateTaxExemption(ctx context.Context, orgID, exemptionID, name, exemptionType string, rate decimal.Decimal) (*taxation.TaxExemption, error)
	GetTaxExemption(ctx context.Context, id string) (*taxation.TaxExemption, error)
	GetTaxExemptionByExemptionID(ctx context.Context, exemptionID string) (*taxation.TaxExemption, error)
	GetOrganizationExemptions(ctx context.Context, orgID string) ([]*taxation.TaxExemption, error)
	UpdateTaxExemption(ctx context.Context, exemption *taxation.TaxExemption) (*taxation.TaxExemption, error)
	DeleteTaxExemption(ctx context.Context, id string) error
	ListTaxExemptions(ctx context.Context, limit, offset int) ([]*taxation.TaxExemption, int, error)
	GetValidExemptions(ctx context.Context, orgID string) ([]*taxation.TaxExemption, error)
	AppliesTo(ctx context.Context, exemptionID, entityType, category, hsnCode string) (bool, error)
	CalculateExemption(ctx context.Context, exemptionID string, taxAmount decimal.Decimal) (decimal.Decimal, error)
}

// TaxExemptionService provides business logic for tax exemption management
type TaxExemptionService struct {
	repo TaxExemptionRepositoryInterface
}

// NewTaxExemptionService creates a new tax exemption service
func NewTaxExemptionService(repo TaxExemptionRepositoryInterface) *TaxExemptionService {
	return &TaxExemptionService{repo: repo}
}

// CreateTaxExemption creates a new tax exemption
func (s *TaxExemptionService) CreateTaxExemption(ctx context.Context, orgID, exemptionID, name, exemptionType string, rate decimal.Decimal) (*taxation.TaxExemption, error) {
	// Validate input
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}
	if exemptionID == "" {
		return nil, fmt.Errorf("exemption ID is required")
	}
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if exemptionType == "" {
		return nil, fmt.Errorf("exemption type is required")
	}

	// Validate exemption type
	validTypes := map[string]bool{
		"full":      true,
		"partial":   true,
		"threshold": true,
	}
	if !validTypes[exemptionType] {
		return nil, fmt.Errorf("invalid exemption type: %s, must be one of: full, partial, threshold", exemptionType)
	}

	// Validate exemption rate (0-100%)
	if rate.LessThan(decimal.Zero) || rate.GreaterThan(decimal.NewFromInt(100)) {
		return nil, fmt.Errorf("exemption rate must be between 0 and 100")
	}

	// Check if exemption ID already exists
	existing, err := s.repo.GetByExemptionID(ctx, exemptionID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("exemption with ID %s already exists", exemptionID)
	}

	// Create new tax exemption
	exemption := taxation.NewTaxExemption(orgID, exemptionID, name, exemptionType)
	exemption.ExemptionRate = rate

	if err := s.repo.Create(ctx, exemption); err != nil {
		return nil, fmt.Errorf("failed to create tax exemption: %w", err)
	}

	return exemption, nil
}

// GetTaxExemption retrieves a tax exemption by ID
func (s *TaxExemptionService) GetTaxExemption(ctx context.Context, id string) (*taxation.TaxExemption, error) {
	if id == "" {
		return nil, fmt.Errorf("tax exemption ID is required")
	}

	exemption, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get tax exemption: %w", err)
	}
	if exemption == nil {
		return nil, fmt.Errorf("tax exemption not found")
	}

	return exemption, nil
}

// GetTaxExemptionByExemptionID retrieves a tax exemption by exemption ID
func (s *TaxExemptionService) GetTaxExemptionByExemptionID(ctx context.Context, exemptionID string) (*taxation.TaxExemption, error) {
	if exemptionID == "" {
		return nil, fmt.Errorf("exemption ID is required")
	}

	exemption, err := s.repo.GetByExemptionID(ctx, exemptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tax exemption: %w", err)
	}
	if exemption == nil {
		return nil, fmt.Errorf("tax exemption not found")
	}

	return exemption, nil
}

// GetOrganizationExemptions retrieves all tax exemptions for an organization
func (s *TaxExemptionService) GetOrganizationExemptions(ctx context.Context, orgID string) ([]*taxation.TaxExemption, error) {
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	exemptions, err := s.repo.GetByOrganizationID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization exemptions: %w", err)
	}

	return exemptions, nil
}

// GetValidExemptions retrieves all currently valid tax exemptions for an organization
func (s *TaxExemptionService) GetValidExemptions(ctx context.Context, orgID string) ([]*taxation.TaxExemption, error) {
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	allExemptions, err := s.repo.GetByOrganizationID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization exemptions: %w", err)
	}

	// Filter to only valid exemptions
	validExemptions := make([]*taxation.TaxExemption, 0)
	for _, exemption := range allExemptions {
		if exemption.IsValid() {
			validExemptions = append(validExemptions, exemption)
		}
	}

	return validExemptions, nil
}

// UpdateTaxExemption updates an existing tax exemption
func (s *TaxExemptionService) UpdateTaxExemption(ctx context.Context, exemption *taxation.TaxExemption) (*taxation.TaxExemption, error) {
	if exemption == nil {
		return nil, fmt.Errorf("tax exemption cannot be nil")
	}

	// Validate exists
	existing, err := s.repo.GetByID(ctx, exemption.ID)
	if err != nil || existing == nil {
		return nil, fmt.Errorf("tax exemption not found")
	}

	// Validate exemption rate
	if exemption.ExemptionRate.LessThan(decimal.Zero) || exemption.ExemptionRate.GreaterThan(decimal.NewFromInt(100)) {
		return nil, fmt.Errorf("exemption rate must be between 0 and 100")
	}

	// Validate validity dates
	if exemption.ValidFrom != nil && exemption.ValidTo != nil {
		if exemption.ValidTo.Before(*exemption.ValidFrom) {
			return nil, fmt.Errorf("valid to date must be after valid from date")
		}
	}

	if err := s.repo.Update(ctx, exemption); err != nil {
		return nil, fmt.Errorf("failed to update tax exemption: %w", err)
	}

	return exemption, nil
}

// DeleteTaxExemption soft deletes a tax exemption
func (s *TaxExemptionService) DeleteTaxExemption(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("tax exemption ID is required")
	}

	// Check exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("tax exemption not found")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete tax exemption: %w", err)
	}

	return nil
}

// ListTaxExemptions retrieves tax exemptions with pagination
func (s *TaxExemptionService) ListTaxExemptions(ctx context.Context, limit, offset int) ([]*taxation.TaxExemption, int, error) {
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	exemptions, total, err := s.repo.List(ctx, nil, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tax exemptions: %w", err)
	}

	return exemptions, total, nil
}

// AppliesTo checks if an exemption applies to a specific entity
func (s *TaxExemptionService) AppliesTo(ctx context.Context, exemptionID, entityType, category, hsnCode string) (bool, error) {
	if exemptionID == "" {
		return false, fmt.Errorf("exemption ID is required")
	}

	exemption, err := s.repo.GetByExemptionID(ctx, exemptionID)
	if err != nil || exemption == nil {
		return false, fmt.Errorf("tax exemption not found")
	}

	// Check if exemption is valid
	if !exemption.IsValid() {
		return false, nil
	}

	// Check entity type match
	if exemption.EntityType != "" && exemption.EntityType != "all" && exemption.EntityType != entityType {
		return false, nil
	}

	// Check category match
	if exemption.Category != "" && exemption.Category != category {
		return false, nil
	}

	// Check HSN code match
	if exemption.HSNCode != "" && exemption.HSNCode != hsnCode {
		return false, nil
	}

	return true, nil
}

// CalculateExemption calculates the exempted tax amount
func (s *TaxExemptionService) CalculateExemption(ctx context.Context, exemptionID string, taxAmount decimal.Decimal) (decimal.Decimal, error) {
	if exemptionID == "" {
		return decimal.Zero, fmt.Errorf("exemption ID is required")
	}
	if taxAmount.LessThan(decimal.Zero) {
		return decimal.Zero, fmt.Errorf("tax amount cannot be negative")
	}

	exemption, err := s.repo.GetByExemptionID(ctx, exemptionID)
	if err != nil || exemption == nil {
		return decimal.Zero, fmt.Errorf("tax exemption not found")
	}

	// Check if exemption is valid
	if !exemption.IsValid() {
		return decimal.Zero, nil
	}

	// Calculate exempted amount based on exemption type
	switch exemption.ExemptionType {
	case "full":
		return taxAmount, nil
	case "partial":
		// Calculate percentage of tax to exempt
		exemptedAmount := taxAmount.Mul(exemption.ExemptionRate).Div(decimal.NewFromInt(100))
		return exemptedAmount, nil
	case "threshold":
		// For threshold exemptions, apply full rate if above threshold
		// This is a simplified implementation - actual logic would depend on business rules
		return taxAmount.Mul(exemption.ExemptionRate).Div(decimal.NewFromInt(100)), nil
	default:
		return decimal.Zero, fmt.Errorf("unknown exemption type: %s", exemption.ExemptionType)
	}
}
