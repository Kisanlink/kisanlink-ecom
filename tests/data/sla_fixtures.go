package data

import (
	"time"

	"kisanlink-ecom/entities/models/services"
)

// CreateTestSLA creates a test SLA with default values
func CreateTestSLA() *services.SLA {
	return &services.SLA{
		OrganizationID: "org-test-id-123",
		CatalogItemID:  "catalog-item-test-id-123",
		Name:           "Test SLA",
		Type:           services.SLATypeResponse,
		Description:    "Test SLA for unit testing",
		TargetValue:    30.0,
		Unit:           services.SLAUnitMinutes,
		BusinessDays:   "MON-FRI",
		IsActive:       true,
	}
}

// CreateTestSLAWithID creates a test SLA with a specific ID
func CreateTestSLAWithID(id string) *services.SLA {
	sla := CreateTestSLA()
	sla.ID = id
	return sla
}

// CreateTestSLAWithOrgID creates a test SLA with a specific organization ID
func CreateTestSLAWithOrgID(orgID string) *services.SLA {
	sla := CreateTestSLA()
	sla.OrganizationID = orgID
	return sla
}

// CreateTestSLAWithCatalogItemID creates a test SLA with a specific catalog item ID
func CreateTestSLAWithCatalogItemID(catalogItemID string) *services.SLA {
	sla := CreateTestSLA()
	sla.CatalogItemID = catalogItemID
	return sla
}

// CreateTestResponseTimeSLA creates a response time SLA
func CreateTestResponseTimeSLA() *services.SLA {
	sla := CreateTestSLA()
	sla.Type = services.SLATypeResponse
	responseTime := 60
	sla.ResponseTime = &responseTime
	return sla
}

// CreateTestResolutionTimeSLA creates a resolution time SLA
func CreateTestResolutionTimeSLA() *services.SLA {
	sla := CreateTestSLA()
	sla.Type = services.SLATypeResolution
	resolutionTime := 240
	sla.ResolutionTime = &resolutionTime
	return sla
}

// CreateTestAvailabilitySLA creates an availability SLA
func CreateTestAvailabilitySLA() *services.SLA {
	sla := CreateTestSLA()
	sla.Type = services.SLATypeAvailability
	uptime := 99.9
	sla.UptimePercentage = &uptime
	sla.Unit = services.SLAUnitPercent
	sla.TargetValue = 99.9
	return sla
}

// CreateTestPerformanceSLA creates a performance SLA
func CreateTestPerformanceSLA() *services.SLA {
	sla := CreateTestSLA()
	sla.Type = services.SLATypePerformance
	sla.TargetValue = 500.0
	sla.Unit = services.SLAUnitMinutes
	return sla
}

// CreateTestInactiveSLA creates an inactive SLA
func CreateTestInactiveSLA() *services.SLA {
	sla := CreateTestSLA()
	sla.IsActive = false
	return sla
}

// CreateTestSLAWithThresholds creates an SLA with warning and critical thresholds
func CreateTestSLAWithThresholds(warning, critical float64) *services.SLA {
	sla := CreateTestSLA()
	sla.WarningThreshold = &warning
	sla.CriticalThreshold = &critical
	return sla
}

// CreateTestSLAWithBusinessHours creates an SLA with business hours
func CreateTestSLAWithBusinessHours(start, end time.Time) *services.SLA {
	sla := CreateTestSLA()
	sla.BusinessHoursStart = &start
	sla.BusinessHoursEnd = &end
	return sla
}

// CreateTestSLAWithName creates an SLA with a specific name
func CreateTestSLAWithName(orgID, name string) *services.SLA {
	sla := CreateTestSLA()
	sla.OrganizationID = orgID
	sla.Name = name
	return sla
}

// CreateTestSLAWithType creates an SLA with a specific type
func CreateTestSLAWithType(slaType services.SLAType) *services.SLA {
	sla := CreateTestSLA()
	sla.Type = slaType
	return sla
}

// CreateTestSLAsArray creates a slice of test SLAs
func CreateTestSLAsArray(count int, orgID string) []*services.SLA {
	slas := make([]*services.SLA, count)
	for i := 0; i < count; i++ {
		slas[i] = CreateTestSLAWithOrgID(orgID)
		slas[i].Name = generateSLAName(i)
	}
	return slas
}

// CreateTestSLAsArrayForCatalogItem creates a slice of test SLAs for a catalog item
func CreateTestSLAsArrayForCatalogItem(count int, catalogItemID string) []*services.SLA {
	slas := make([]*services.SLA, count)
	for i := 0; i < count; i++ {
		slas[i] = CreateTestSLAWithCatalogItemID(catalogItemID)
		slas[i].Name = generateSLAName(i)
	}
	return slas
}

// Helper functions
func generateSLAName(index int) string {
	return "Test SLA " + string(rune('A'+index))
}
