package data

import (
	"time"

	"kisanlink-ecom/entities/models/taxation"

	"github.com/shopspring/decimal"
)

// CreateTestTaxExemption creates a test tax exemption with default values
func CreateTestTaxExemption() *taxation.TaxExemption {
	return &taxation.TaxExemption{
		OrgID:         "org-test-id-123",
		ExemptionID:   "EXEMPT-TEST-123",
		Name:          "Test Tax Exemption",
		Description:   "Test exemption for unit testing",
		ExemptionType: "full",
		ExemptionRate: decimal.NewFromInt(100),
		EntityType:    "product",
		Category:      "agriculture",
		HSNCode:       "1001",
		IsActive:      true,
	}
}

// CreateTestTaxExemptionWithID creates a test tax exemption with a specific ID
func CreateTestTaxExemptionWithID(id string) *taxation.TaxExemption {
	exemption := CreateTestTaxExemption()
	exemption.ID = id
	return exemption
}

// CreateTestTaxExemptionWithOrgID creates a test tax exemption with a specific organization ID
func CreateTestTaxExemptionWithOrgID(orgID string) *taxation.TaxExemption {
	exemption := CreateTestTaxExemption()
	exemption.OrgID = orgID
	return exemption
}

// CreateTestTaxExemptionWithExemptionID creates a test tax exemption with a specific exemption ID
func CreateTestTaxExemptionWithExemptionID(exemptionID string) *taxation.TaxExemption {
	exemption := CreateTestTaxExemption()
	exemption.ExemptionID = exemptionID
	return exemption
}

// CreateTestPartialTaxExemption creates a partial tax exemption
func CreateTestPartialTaxExemption() *taxation.TaxExemption {
	exemption := CreateTestTaxExemption()
	exemption.ExemptionType = "partial"
	exemption.ExemptionRate = decimal.NewFromInt(50)
	return exemption
}

// CreateTestInactiveTaxExemption creates an inactive tax exemption
func CreateTestInactiveTaxExemption() *taxation.TaxExemption {
	exemption := CreateTestTaxExemption()
	exemption.IsActive = false
	return exemption
}

// CreateTestTaxExemptionWithValidity creates a tax exemption with validity period
func CreateTestTaxExemptionWithValidity(validFrom, validTo time.Time) *taxation.TaxExemption {
	exemption := CreateTestTaxExemption()
	exemption.ValidFrom = &validFrom
	exemption.ValidTo = &validTo
	return exemption
}

// CreateTestTaxExemptionWithEntityType creates a tax exemption for a specific entity type
func CreateTestTaxExemptionWithEntityType(entityType string) *taxation.TaxExemption {
	exemption := CreateTestTaxExemption()
	exemption.EntityType = entityType
	return exemption
}

// CreateTestTaxExemptionWithHSNCode creates a tax exemption with a specific HSN code
func CreateTestTaxExemptionWithHSNCode(hsnCode string) *taxation.TaxExemption {
	exemption := CreateTestTaxExemption()
	exemption.HSNCode = hsnCode
	return exemption
}

// CreateTestTaxExemptionWithCategory creates a tax exemption for a specific category
func CreateTestTaxExemptionWithCategory(category string) *taxation.TaxExemption {
	exemption := CreateTestTaxExemption()
	exemption.Category = category
	return exemption
}

// CreateTestTaxExemptionsArray creates a slice of test tax exemptions
func CreateTestTaxExemptionsArray(count int, orgID string) []*taxation.TaxExemption {
	exemptions := make([]*taxation.TaxExemption, count)
	for i := 0; i < count; i++ {
		exemptions[i] = CreateTestTaxExemptionWithOrgID(orgID)
		exemptions[i].ExemptionID = generateExemptionID(i)
		exemptions[i].Name = generateExemptionName(i)
	}
	return exemptions
}

// Helper functions
func generateExemptionID(index int) string {
	return "EXEMPT-TEST-" + string(rune('A'+index))
}

func generateExemptionName(index int) string {
	return "Test Exemption " + string(rune('A'+index))
}
