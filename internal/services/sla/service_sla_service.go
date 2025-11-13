// Package sla provides service layer for service-level agreement management
package sla

import (
	"context"
	"fmt"

	"kisanlink-ecom/entities/models/services"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// ServiceSLARepositoryInterface defines repository operations
type ServiceSLARepositoryInterface interface {
	Create(ctx context.Context, sla *services.SLA) error
	GetByID(ctx context.Context, id string) (*services.SLA, error)
	GetByCatalogItemID(ctx context.Context, catalogItemID string) ([]*services.SLA, error)
	GetByOrganizationID(ctx context.Context, organizationID string) ([]*services.SLA, error)
	Update(ctx context.Context, sla *services.SLA) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter *base.Filter, limit, offset int) ([]*services.SLA, int, error)
}

// ServiceSLAServiceInterface defines service operations
type ServiceSLAServiceInterface interface {
	CreateServiceSLA(ctx context.Context, orgID, catalogItemID, name string, slaType services.SLAType, targetValue float64, unit services.SLAUnit) (*services.SLA, error)
	GetServiceSLA(ctx context.Context, id string) (*services.SLA, error)
	GetCatalogItemSLAs(ctx context.Context, catalogItemID string) ([]*services.SLA, error)
	GetOrganizationSLAs(ctx context.Context, organizationID string) ([]*services.SLA, error)
	UpdateServiceSLA(ctx context.Context, sla *services.SLA) (*services.SLA, error)
	UpdateThresholds(ctx context.Context, id string, warning, critical *float64) (*services.SLA, error)
	DeleteServiceSLA(ctx context.Context, id string) error
	ListServiceSLAs(ctx context.Context, limit, offset int) ([]*services.SLA, int, error)
	IsViolated(ctx context.Context, slaID string, actualValue float64) (bool, error)
	GetActiveSLAs(ctx context.Context, catalogItemID string) ([]*services.SLA, error)
}

// ServiceSLAService provides business logic for service SLA management
type ServiceSLAService struct {
	repo ServiceSLARepositoryInterface
}

// NewServiceSLAService creates a new service SLA service
func NewServiceSLAService(repo ServiceSLARepositoryInterface) *ServiceSLAService {
	return &ServiceSLAService{repo: repo}
}

// CreateServiceSLA creates a new service SLA
func (s *ServiceSLAService) CreateServiceSLA(ctx context.Context, orgID, catalogItemID, name string, slaType services.SLAType, targetValue float64, unit services.SLAUnit) (*services.SLA, error) {
	// Validate input
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}
	if catalogItemID == "" {
		return nil, fmt.Errorf("catalog item ID is required")
	}
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	// Validate SLA type
	validTypes := map[services.SLAType]bool{
		services.SLATypeResponse:     true,
		services.SLATypeResolution:   true,
		services.SLATypeAvailability: true,
		services.SLATypePerformance:  true,
	}
	if !validTypes[slaType] {
		return nil, fmt.Errorf("invalid SLA type: %s", slaType)
	}

	// Validate target value
	if targetValue <= 0 {
		return nil, fmt.Errorf("target value must be greater than 0")
	}

	// Validate unit
	validUnits := map[services.SLAUnit]bool{
		services.SLAUnitMinutes: true,
		services.SLAUnitHours:   true,
		services.SLAUnitDays:    true,
		services.SLAUnitPercent: true,
	}
	if !validUnits[unit] {
		return nil, fmt.Errorf("invalid SLA unit: %s", unit)
	}

	// Validate unit matches SLA type
	if slaType == services.SLATypeAvailability && unit != services.SLAUnitPercent {
		return nil, fmt.Errorf("availability SLA must use percent unit")
	}

	// Create new service SLA
	sla := services.NewSLA(orgID, catalogItemID, name, slaType, targetValue, unit)

	if err := s.repo.Create(ctx, sla); err != nil {
		return nil, fmt.Errorf("failed to create service SLA: %w", err)
	}

	return sla, nil
}

// GetServiceSLA retrieves a service SLA by ID
func (s *ServiceSLAService) GetServiceSLA(ctx context.Context, id string) (*services.SLA, error) {
	if id == "" {
		return nil, fmt.Errorf("service SLA ID is required")
	}

	sla, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get service SLA: %w", err)
	}
	if sla == nil {
		return nil, fmt.Errorf("service SLA not found")
	}

	return sla, nil
}

// GetCatalogItemSLAs retrieves all SLAs for a catalog item
func (s *ServiceSLAService) GetCatalogItemSLAs(ctx context.Context, catalogItemID string) ([]*services.SLA, error) {
	if catalogItemID == "" {
		return nil, fmt.Errorf("catalog item ID is required")
	}

	slas, err := s.repo.GetByCatalogItemID(ctx, catalogItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get catalog item SLAs: %w", err)
	}

	return slas, nil
}

// GetActiveSLAs retrieves all active SLAs for a catalog item
func (s *ServiceSLAService) GetActiveSLAs(ctx context.Context, catalogItemID string) ([]*services.SLA, error) {
	if catalogItemID == "" {
		return nil, fmt.Errorf("catalog item ID is required")
	}

	allSLAs, err := s.repo.GetByCatalogItemID(ctx, catalogItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get catalog item SLAs: %w", err)
	}

	// Filter to only active SLAs
	activeSLAs := make([]*services.SLA, 0)
	for _, sla := range allSLAs {
		if sla.IsActive {
			activeSLAs = append(activeSLAs, sla)
		}
	}

	return activeSLAs, nil
}

// GetOrganizationSLAs retrieves all SLAs for an organization
func (s *ServiceSLAService) GetOrganizationSLAs(ctx context.Context, organizationID string) ([]*services.SLA, error) {
	if organizationID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	slas, err := s.repo.GetByOrganizationID(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization SLAs: %w", err)
	}

	return slas, nil
}

// UpdateServiceSLA updates an existing service SLA
func (s *ServiceSLAService) UpdateServiceSLA(ctx context.Context, sla *services.SLA) (*services.SLA, error) {
	if sla == nil {
		return nil, fmt.Errorf("service SLA cannot be nil")
	}

	// Validate exists
	existing, err := s.repo.GetByID(ctx, sla.ID)
	if err != nil || existing == nil {
		return nil, fmt.Errorf("service SLA not found")
	}

	// Validate target value
	if sla.TargetValue <= 0 {
		return nil, fmt.Errorf("target value must be greater than 0")
	}

	// Validate thresholds
	if err := s.validateThresholds(sla); err != nil {
		return nil, fmt.Errorf("threshold validation failed: %w", err)
	}

	if err := s.repo.Update(ctx, sla); err != nil {
		return nil, fmt.Errorf("failed to update service SLA: %w", err)
	}

	return sla, nil
}

// UpdateThresholds updates the warning and critical thresholds for an SLA
func (s *ServiceSLAService) UpdateThresholds(ctx context.Context, id string, warning, critical *float64) (*services.SLA, error) {
	if id == "" {
		return nil, fmt.Errorf("service SLA ID is required")
	}

	sla, err := s.repo.GetByID(ctx, id)
	if err != nil || sla == nil {
		return nil, fmt.Errorf("service SLA not found")
	}

	// Update thresholds
	if warning != nil {
		sla.WarningThreshold = warning
	}
	if critical != nil {
		sla.CriticalThreshold = critical
	}

	// Validate thresholds
	if err := s.validateThresholds(sla); err != nil {
		return nil, fmt.Errorf("threshold validation failed: %w", err)
	}

	if err := s.repo.Update(ctx, sla); err != nil {
		return nil, fmt.Errorf("failed to update thresholds: %w", err)
	}

	return sla, nil
}

// DeleteServiceSLA soft deletes a service SLA
func (s *ServiceSLAService) DeleteServiceSLA(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("service SLA ID is required")
	}

	// Check exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("service SLA not found")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete service SLA: %w", err)
	}

	return nil
}

// ListServiceSLAs retrieves service SLAs with pagination
func (s *ServiceSLAService) ListServiceSLAs(ctx context.Context, limit, offset int) ([]*services.SLA, int, error) {
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	slas, total, err := s.repo.List(ctx, nil, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list service SLAs: %w", err)
	}

	return slas, total, nil
}

// IsViolated checks if an SLA is violated based on actual value
func (s *ServiceSLAService) IsViolated(ctx context.Context, slaID string, actualValue float64) (bool, error) {
	if slaID == "" {
		return false, fmt.Errorf("service SLA ID is required")
	}

	sla, err := s.repo.GetByID(ctx, slaID)
	if err != nil || sla == nil {
		return false, fmt.Errorf("service SLA not found")
	}

	if !sla.IsActive {
		return false, nil
	}

	// For response/resolution time SLAs, violation occurs when actual > target
	if sla.Type == services.SLATypeResponse || sla.Type == services.SLATypeResolution || sla.Type == services.SLATypePerformance {
		return actualValue > sla.TargetValue, nil
	}

	// For availability SLAs, violation occurs when actual < target
	if sla.Type == services.SLATypeAvailability {
		return actualValue < sla.TargetValue, nil
	}

	return false, nil
}

// validateThresholds validates that thresholds are properly configured
func (s *ServiceSLAService) validateThresholds(sla *services.SLA) error {
	if sla.WarningThreshold != nil && sla.CriticalThreshold != nil {
		// For response/resolution time SLAs, critical should be greater than warning
		if sla.Type == services.SLATypeResponse || sla.Type == services.SLATypeResolution || sla.Type == services.SLATypePerformance {
			if *sla.CriticalThreshold <= *sla.WarningThreshold {
				return fmt.Errorf("critical threshold must be greater than warning threshold")
			}
		}

		// For availability SLAs, warning should be greater than critical
		if sla.Type == services.SLATypeAvailability {
			if *sla.WarningThreshold <= *sla.CriticalThreshold {
				return fmt.Errorf("warning threshold must be greater than critical threshold for availability SLAs")
			}
		}
	}

	return nil
}
