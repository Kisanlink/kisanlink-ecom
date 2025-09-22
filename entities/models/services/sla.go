package services

import (
	"fmt"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// SLAType represents the type of SLA
type SLAType string

const (
	SLATypeResponse     SLAType = "response"     // Response time SLA
	SLATypeResolution   SLAType = "resolution"   // Resolution time SLA
	SLATypeAvailability SLAType = "availability" // Availability SLA
	SLATypePerformance  SLAType = "performance"  // Performance SLA
)

// SLAUnit represents the unit of measurement for SLA
type SLAUnit string

const (
	SLAUnitMinutes SLAUnit = "minutes"
	SLAUnitHours   SLAUnit = "hours"
	SLAUnitDays    SLAUnit = "days"
	SLAUnitPercent SLAUnit = "percent"
)

// SLA represents Service Level Agreement
type SLA struct {
	base.BaseModel

	// Tenant isolation
	OrganizationID string `json:"organization_id" gorm:"type:varchar(255);not null;index:idx_tenant"`

	// Parent catalog item (for services)
	CatalogItemID string `json:"catalog_item_id" gorm:"type:varchar(255);not null;index"`

	// SLA details
	Name        string  `json:"name" gorm:"type:varchar(255);not null"`
	Type        SLAType `json:"type" gorm:"type:varchar(20);not null"`
	Description string  `json:"description" gorm:"type:text"`

	// SLA metrics
	TargetValue float64 `json:"target_value" gorm:"not null"`
	Unit        SLAUnit `json:"unit" gorm:"type:varchar(20);not null"`

	// Thresholds
	WarningThreshold  *float64 `json:"warning_threshold"`
	CriticalThreshold *float64 `json:"critical_threshold"`

	// Time-based SLA
	ResponseTime   *int `json:"response_time"`   // in minutes
	ResolutionTime *int `json:"resolution_time"` // in minutes

	// Availability SLA
	UptimePercentage *float64 `json:"uptime_percentage"`

	// Business hours
	BusinessHoursStart *time.Time `json:"business_hours_start"`
	BusinessHoursEnd   *time.Time `json:"business_hours_end"`
	BusinessDays       string     `json:"business_days" gorm:"type:varchar(50)"` // e.g., "MON-FRI"

	// Penalties and rewards
	PenaltyClause string `json:"penalty_clause" gorm:"type:text"`
	RewardClause  string `json:"reward_clause" gorm:"type:text"`

	// Status
	IsActive bool `json:"is_active" gorm:"default:true"`

	// Audit fields
	CreatedBy string `json:"created_by"`
	UpdatedBy string `json:"updated_by"`
}

// TableName returns the table name for GORM
func (SLA) TableName() string {
	return "slas"
}

// NewSLA creates a new SLA
func NewSLA(orgID, catalogItemID, name string, slaType SLAType, targetValue float64, unit SLAUnit) *SLA {
	return &SLA{
		BaseModel:      *base.NewBaseModel("SLA", "large"),
		OrganizationID: orgID,
		CatalogItemID:  catalogItemID,
		Name:           name,
		Type:           slaType,
		TargetValue:    targetValue,
		Unit:           unit,
		BusinessDays:   "MON-FRI",
		IsActive:       true,
	}
}

// IsResponseTimeSLA checks if this is a response time SLA
func (s *SLA) IsResponseTimeSLA() bool {
	return s.Type == SLATypeResponse
}

// IsAvailabilitySLA checks if this is an availability SLA
func (s *SLA) IsAvailabilitySLA() bool {
	return s.Type == SLATypeAvailability
}

// GetTargetValueWithUnit returns the target value with its unit
func (s *SLA) GetTargetValueWithUnit() string {
	return fmt.Sprintf("%.2f %s", s.TargetValue, s.Unit)
}
