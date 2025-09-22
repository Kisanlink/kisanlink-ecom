package catalog

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// SLAType represents the type of SLA
type SLAType string

const (
	SLATypeService      SLAType = "service"
	SLATypeSupport      SLAType = "support"
	SLATypeDelivery     SLAType = "delivery"
	SLATypeAvailability SLAType = "availability"
)

// SLA represents Service Level Agreement information
type SLA struct {
	base.BaseModel

	// Tenant isolation
	OrganizationID string `json:"organization_id" gorm:"type:varchar(255);not null;index:idx_tenant"`

	// Parent catalog item
	CatalogItemID string `json:"catalog_item_id" gorm:"type:varchar(255);not null;index:idx_catalog_item"`

	// SLA details
	Type        SLAType `json:"type" gorm:"type:varchar(20);not null;check:type IN ('service', 'support', 'delivery', 'availability')"`
	Name        string  `json:"name" gorm:"type:varchar(255);not null"`
	Description string  `json:"description" gorm:"type:text"`

	// Response and resolution times (in minutes)
	ResponseTimeMinutes   *int `json:"response_time_minutes" gorm:"check:response_time_minutes IS NULL OR response_time_minutes > 0"`
	ResolutionTimeMinutes *int `json:"resolution_time_minutes" gorm:"check:resolution_time_minutes IS NULL OR resolution_time_minutes > 0"`

	// Availability percentage (0-100)
	AvailabilityPercentage *float64 `json:"availability_percentage" gorm:"type:decimal(5,2);check:availability_percentage IS NULL OR (availability_percentage >= 0 AND availability_percentage <= 100)"`

	// Support hours
	SupportHours string `json:"support_hours" gorm:"type:varchar(100)"` // e.g., "24/7", "9AM-5PM", "Business Hours"

	// Uptime guarantees
	UptimePercentage *float64 `json:"uptime_percentage" gorm:"type:decimal(5,2);check:uptime_percentage IS NULL OR (uptime_percentage >= 0 AND uptime_percentage <= 100)"`

	// Penalties and credits
	PenaltyClause string `json:"penalty_clause" gorm:"type:text"`
	CreditPolicy  string `json:"credit_policy" gorm:"type:text"`

	// Validity period
	ValidFrom *time.Time `json:"valid_from" gorm:"type:timestamp"`
	ValidTo   *time.Time `json:"valid_to" gorm:"type:timestamp"`

	// Status
	IsActive bool `json:"is_active" gorm:"not null;default:true;index:idx_active"`

	// Metadata
	Metadata string `json:"metadata" gorm:"type:jsonb;index:idx_slas_metadata"`

	// Audit fields
	CreatedBy string `json:"created_by" gorm:"type:varchar(255);not null"`
	UpdatedBy string `json:"updated_by" gorm:"type:varchar(255);not null"`

	// Soft delete support
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index:idx_slas_deleted"`

	// Relationships
	CatalogItem *CatalogItem `json:"catalog_item,omitempty" gorm:"foreignKey:CatalogItemID"`
}

// TableName returns the table name for GORM
func (SLA) TableName() string {
	return "slas"
}

// NewSLA creates a new SLA record
func NewSLA(orgID, catalogItemID string, slaType SLAType, name string) *SLA {
	return &SLA{
		BaseModel:      *base.NewBaseModel("SLA", "large"),
		OrganizationID: orgID,
		CatalogItemID:  catalogItemID,
		Type:           slaType,
		Name:           name,
		IsActive:       true,
	}
}

// IsValid checks if the SLA is currently valid
func (s *SLA) IsValid() bool {
	if !s.IsActive {
		return false
	}

	now := time.Now()

	if s.ValidFrom != nil && now.Before(*s.ValidFrom) {
		return false
	}

	if s.ValidTo != nil && now.After(*s.ValidTo) {
		return false
	}

	return true
}

// GetResponseTimeSLA returns the response time SLA in minutes
func (s *SLA) GetResponseTimeSLA() int {
	if s.ResponseTimeMinutes == nil {
		return 0
	}
	return *s.ResponseTimeMinutes
}

// GetResolutionTimeSLA returns the resolution time SLA in minutes
func (s *SLA) GetResolutionTimeSLA() int {
	if s.ResolutionTimeMinutes == nil {
		return 0
	}
	return *s.ResolutionTimeMinutes
}

// GetAvailabilitySLA returns the availability SLA percentage
func (s *SLA) GetAvailabilitySLA() float64 {
	if s.AvailabilityPercentage == nil {
		return 0.0
	}
	return *s.AvailabilityPercentage
}

// GetUptimeSLA returns the uptime SLA percentage
func (s *SLA) GetUptimeSLA() float64 {
	if s.UptimePercentage == nil {
		return 0.0
	}
	return *s.UptimePercentage
}
