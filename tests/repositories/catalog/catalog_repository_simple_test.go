package catalog

import (
	"testing"

	catalogModels "github.com/Kisanlink/kisanlink-ecom/entities/models/catalog"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// TestCatalogModels tests the catalog model structure
func TestCatalogModels(t *testing.T) {
	t.Run("Create Product", func(t *testing.T) {
		weight := decimal.NewFromFloat(5.5)
		shelfLifeDays := 30

		product := &catalogModels.Product{
			CatalogItem: catalogModels.CatalogItem{
				OrganizationID: "org-123",
				ItemType:       catalogModels.CatalogItemTypeProduct,
				Name:           "Test Product",
				Description:    "Test Description",
				BasePrice:      decimal.NewFromFloat(100.0),
				Currency:       "INR",
				Category:       "agriculture",
				SKU:            "TEST-001",
				IsActive:       true,
				Visibility:     catalogModels.VisibilityOrg,
			},
			Weight:        &weight,
			Perishable:    true,
			ShelfLifeDays: &shelfLifeDays,
		}

		assert.NotNil(t, product)
		assert.Equal(t, "Test Product", product.Name)
		assert.Equal(t, catalogModels.CatalogItemTypeProduct, product.ItemType)
		assert.True(t, product.Perishable)
		assert.Equal(t, decimal.NewFromFloat(5.5), *product.Weight)
	})

	t.Run("Create Service", func(t *testing.T) {
		duration := 60

		service := &catalogModels.Service{
			CatalogItem: catalogModels.CatalogItem{
				OrganizationID: "org-123",
				ItemType:       catalogModels.CatalogItemTypeService,
				Name:           "Test Service",
				Description:    "Test Service Description",
				BasePrice:      decimal.NewFromFloat(200.0),
				Currency:       "INR",
				Category:       "consulting",
				IsActive:       true,
				Visibility:     catalogModels.VisibilityPrivate,
			},
			DurationMinutes: &duration,
		}

		assert.NotNil(t, service)
		assert.Equal(t, "Test Service", service.Name)
		assert.Equal(t, catalogModels.CatalogItemTypeService, service.ItemType)
		assert.Equal(t, 60, *service.DurationMinutes)
	})

	t.Run("Create Labour", func(t *testing.T) {
		hourlyRate := decimal.NewFromFloat(50.0)

		labour := &catalogModels.Labour{
			CatalogItem: catalogModels.CatalogItem{
				OrganizationID: "org-123",
				ItemType:       catalogModels.CatalogItemTypeLabour,
				Name:           "Farm Worker",
				Description:    "Experienced farm worker",
				BasePrice:      decimal.NewFromFloat(500.0),
				Currency:       "INR",
				Category:       "agriculture",
				IsActive:       true,
				Visibility:     catalogModels.VisibilityNetwork,
			},
			SkillLevel: "intermediate",
			HourlyRate: &hourlyRate,
		}

		assert.NotNil(t, labour)
		assert.Equal(t, "Farm Worker", labour.Name)
		assert.Equal(t, catalogModels.CatalogItemTypeLabour, labour.ItemType)
		assert.Equal(t, "intermediate", labour.SkillLevel)
		assert.Equal(t, decimal.NewFromFloat(50.0), *labour.HourlyRate)
	})
}

// TestCatalogItemTypes tests the catalog item type enum
func TestCatalogItemTypes(t *testing.T) {
	assert.Equal(t, catalogModels.CatalogItemType("PRODUCT"), catalogModels.CatalogItemTypeProduct)
	assert.Equal(t, catalogModels.CatalogItemType("SERVICE"), catalogModels.CatalogItemTypeService)
	assert.Equal(t, catalogModels.CatalogItemType("LABOUR"), catalogModels.CatalogItemTypeLabour)
}

// TestVisibilityTypes tests the visibility type enum
func TestVisibilityTypes(t *testing.T) {
	assert.Equal(t, catalogModels.VisibilityType("PRIVATE"), catalogModels.VisibilityPrivate)
	assert.Equal(t, catalogModels.VisibilityType("ORG"), catalogModels.VisibilityOrg)
	assert.Equal(t, catalogModels.VisibilityType("NETWORK"), catalogModels.VisibilityNetwork)
	assert.Equal(t, catalogModels.VisibilityType("PUBLIC"), catalogModels.VisibilityPublic)
}
