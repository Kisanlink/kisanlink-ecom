package catalog

import (
	"encoding/json"
	"testing"
	"time"

	catalogModels "kisanlink-ecom/entities/models/catalog"
	catalogResponses "kisanlink-ecom/entities/responses/catalog"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToCatalogItemResponse(t *testing.T) {
	t.Run("converts complete catalog item model to response", func(t *testing.T) {
		// Create test attributes
		attributes := map[string]interface{}{
			"organic":      true,
			"grade":        "A",
			"harvest_date": "2024-01-15",
		}
		attributesJSON, _ := json.Marshal(attributes)

		now := time.Now()
		item := &catalogModels.CatalogItem{
			OrganizationID: "org-123",
			ItemType:       catalogModels.CatalogItemTypeProduct,
			Category:       "vegetables",
			Subcategory:    "tomatoes",
			Name:           "Organic Tomatoes",
			Description:    "Fresh organic tomatoes grown without pesticides",
			SKU:            "TOM-ORG-001",
			UnitOfMeasure:  "kg",
			BasePrice:      decimal.NewFromFloat(25.50),
			Currency:       "INR",
			IsActive:       true,
			Visibility:     catalogModels.VisibilityOrg,
			Tags:           pq.StringArray{"organic", "fresh", "local"},
			Attributes:     string(attributesJSON),
			Images:         pq.StringArray{"https://example.com/tomato1.jpg", "https://example.com/tomato2.jpg"},
		}
		item.ID = "item-123"
		item.CreatedAt = now
		item.UpdatedAt = now

		response, err := catalogResponses.ToCatalogItemResponse(item)
		require.NoError(t, err)
		require.NotNil(t, response)

		assert.Equal(t, item.ID, response.ID)
		assert.Equal(t, item.OrganizationID, response.OrganizationID)
		assert.Equal(t, string(item.ItemType), response.ItemType)
		assert.Equal(t, item.Category, response.Category)
		assert.Equal(t, item.Subcategory, response.Subcategory)
		assert.Equal(t, item.Name, response.Name)
		assert.Equal(t, item.Description, response.Description)
		assert.Equal(t, item.SKU, response.SKU)
		assert.Equal(t, item.UnitOfMeasure, response.UnitOfMeasure)
		assert.True(t, item.BasePrice.Equal(response.BasePrice))
		assert.Equal(t, item.Currency, response.Currency)
		assert.Equal(t, item.IsActive, response.IsActive)
		assert.Equal(t, string(item.Visibility), response.Visibility)
		assert.Equal(t, []string(item.Tags), response.Tags)
		assert.Equal(t, attributes, response.Attributes)
		assert.Equal(t, []string(item.Images), response.Images)
		assert.Equal(t, item.CreatedAt, response.CreatedAt)
		assert.Equal(t, item.UpdatedAt, response.UpdatedAt)
	})

	t.Run("handles nil item", func(t *testing.T) {
		response, err := catalogResponses.ToCatalogItemResponse(nil)
		assert.NoError(t, err)
		assert.Nil(t, response)
	})

	t.Run("handles item with empty attributes", func(t *testing.T) {
		item := &catalogModels.CatalogItem{
			OrganizationID: "org-123",
			ItemType:       catalogModels.CatalogItemTypeProduct,
			Name:           "Test Item",
			BasePrice:      decimal.NewFromFloat(25.50),
			Attributes:     "",
		}
		item.ID = "item-123"

		response, err := catalogResponses.ToCatalogItemResponse(item)
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Nil(t, response.Attributes)
	})

	t.Run("handles item with invalid JSON in attributes", func(t *testing.T) {
		item := &catalogModels.CatalogItem{
			OrganizationID: "org-123",
			ItemType:       catalogModels.CatalogItemTypeProduct,
			Name:           "Test Item",
			BasePrice:      decimal.NewFromFloat(25.50),
			Attributes:     "invalid json",
		}
		item.ID = "item-123"

		response, err := catalogResponses.ToCatalogItemResponse(item)
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Nil(t, response.Attributes)
	})
}

func TestToProductResponse(t *testing.T) {
	t.Run("converts complete product model to response", func(t *testing.T) {
		dimensions := map[string]interface{}{
			"length": 10.5,
			"width":  8.0,
			"height": 6.5,
		}
		dimensionsJSON, _ := json.Marshal(dimensions)

		now := time.Now()
		weight := decimal.NewFromFloat(2.5)
		shelfLifeDays := 7

		product := &catalogModels.Product{
			CatalogItem: catalogModels.CatalogItem{
				OrganizationID: "org-123",
				ItemType:       catalogModels.CatalogItemTypeProduct,
				Name:           "Organic Tomatoes",
				BasePrice:      decimal.NewFromFloat(25.50),
				Currency:       "INR",
				IsActive:       true,
				Visibility:     catalogModels.VisibilityOrg,
			},
			Weight:        &weight,
			Dimensions:    string(dimensionsJSON),
			Perishable:    true,
			ShelfLifeDays: &shelfLifeDays,
		}
		product.ID = "product-123"
		product.CreatedAt = now
		product.UpdatedAt = now

		response, err := catalogResponses.ToProductResponse(product)
		require.NoError(t, err)
		require.NotNil(t, response)

		assert.Equal(t, product.ID, response.ID)
		assert.Equal(t, string(product.ItemType), response.ItemType)
		assert.Equal(t, product.Name, response.Name)
		assert.True(t, product.BasePrice.Equal(response.BasePrice))
		assert.True(t, weight.Equal(*response.Weight))
		assert.Equal(t, dimensions, response.Dimensions)
		assert.Equal(t, product.Perishable, response.Perishable)
		assert.Equal(t, product.ShelfLifeDays, response.ShelfLifeDays)
	})

	t.Run("handles nil product", func(t *testing.T) {
		response, err := catalogResponses.ToProductResponse(nil)
		assert.NoError(t, err)
		assert.Nil(t, response)
	})

	t.Run("handles product with empty dimensions", func(t *testing.T) {
		product := &catalogModels.Product{
			CatalogItem: catalogModels.CatalogItem{
				OrganizationID: "org-123",
				ItemType:       catalogModels.CatalogItemTypeProduct,
				Name:           "Test Product",
				BasePrice:      decimal.NewFromFloat(25.50),
			},
			Dimensions: "",
		}
		product.ID = "product-123"

		response, err := catalogResponses.ToProductResponse(product)
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Nil(t, response.Dimensions)
	})
}

func TestToServiceResponse(t *testing.T) {
	t.Run("converts complete service model to response", func(t *testing.T) {
		serviceArea := map[string]interface{}{
			"states":   []interface{}{"Maharashtra", "Karnataka"},
			"radius":   float64(50),
			"coverage": "rural",
		}
		serviceAreaJSON, _ := json.Marshal(serviceArea)

		now := time.Now()
		durationMinutes := 120

		service := &catalogModels.Service{
			CatalogItem: catalogModels.CatalogItem{
				OrganizationID: "org-123",
				ItemType:       catalogModels.CatalogItemTypeService,
				Name:           "Farm Consultation",
				BasePrice:      decimal.NewFromFloat(150.00),
				Currency:       "INR",
				IsActive:       true,
				Visibility:     catalogModels.VisibilityNetwork,
			},
			DurationMinutes: &durationMinutes,
			ServiceArea:     string(serviceAreaJSON),
		}
		service.ID = "service-123"
		service.CreatedAt = now
		service.UpdatedAt = now

		response, err := catalogResponses.ToServiceResponse(service)
		require.NoError(t, err)
		require.NotNil(t, response)

		assert.Equal(t, service.ID, response.ID)
		assert.Equal(t, string(service.ItemType), response.ItemType)
		assert.Equal(t, service.Name, response.Name)
		assert.True(t, service.BasePrice.Equal(response.BasePrice))
		assert.Equal(t, service.DurationMinutes, response.DurationMinutes)
		assert.Equal(t, serviceArea, response.ServiceArea)
	})

	t.Run("handles nil service", func(t *testing.T) {
		response, err := catalogResponses.ToServiceResponse(nil)
		assert.NoError(t, err)
		assert.Nil(t, response)
	})

	t.Run("handles service with empty service area", func(t *testing.T) {
		service := &catalogModels.Service{
			CatalogItem: catalogModels.CatalogItem{
				OrganizationID: "org-123",
				ItemType:       catalogModels.CatalogItemTypeService,
				Name:           "Test Service",
				BasePrice:      decimal.NewFromFloat(150.00),
			},
			ServiceArea: "",
		}
		service.ID = "service-123"

		response, err := catalogResponses.ToServiceResponse(service)
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Nil(t, response.ServiceArea)
	})
}

func TestToLabourResponse(t *testing.T) {
	t.Run("converts complete labour model to response", func(t *testing.T) {
		now := time.Now()
		hourlyRate := decimal.NewFromFloat(75.00)

		labour := &catalogModels.Labour{
			CatalogItem: catalogModels.CatalogItem{
				OrganizationID: "org-123",
				ItemType:       catalogModels.CatalogItemTypeLabour,
				Name:           "Farm Worker",
				BasePrice:      decimal.NewFromFloat(75.00),
				Currency:       "INR",
				IsActive:       true,
				Visibility:     catalogModels.VisibilityOrg,
			},
			SkillLevel: "experienced",
			HourlyRate: &hourlyRate,
		}
		labour.ID = "labour-123"
		labour.CreatedAt = now
		labour.UpdatedAt = now

		response, err := catalogResponses.ToLabourResponse(labour)
		require.NoError(t, err)
		require.NotNil(t, response)

		assert.Equal(t, labour.ID, response.ID)
		assert.Equal(t, string(labour.ItemType), response.ItemType)
		assert.Equal(t, labour.Name, response.Name)
		assert.True(t, labour.BasePrice.Equal(response.BasePrice))
		assert.Equal(t, labour.SkillLevel, response.SkillLevel)
		assert.True(t, hourlyRate.Equal(*response.HourlyRate))
	})

	t.Run("handles nil labour", func(t *testing.T) {
		response, err := catalogResponses.ToLabourResponse(nil)
		assert.NoError(t, err)
		assert.Nil(t, response)
	})

	t.Run("handles labour with nil hourly rate", func(t *testing.T) {
		labour := &catalogModels.Labour{
			CatalogItem: catalogModels.CatalogItem{
				OrganizationID: "org-123",
				ItemType:       catalogModels.CatalogItemTypeLabour,
				Name:           "Test Labour",
				BasePrice:      decimal.NewFromFloat(75.00),
			},
			SkillLevel: "beginner",
			HourlyRate: nil,
		}
		labour.ID = "labour-123"

		response, err := catalogResponses.ToLabourResponse(labour)
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Nil(t, response.HourlyRate)
	})
}

func TestToCatalogSummaryResponse(t *testing.T) {
	t.Run("converts catalog item model to summary response", func(t *testing.T) {
		now := time.Now()
		item := &catalogModels.CatalogItem{
			OrganizationID: "org-123",
			ItemType:       catalogModels.CatalogItemTypeProduct,
			Name:           "Organic Tomatoes",
			BasePrice:      decimal.NewFromFloat(25.50),
			Currency:       "INR",
			IsActive:       true,
			Visibility:     catalogModels.VisibilityOrg,
		}
		item.ID = "item-123"
		item.CreatedAt = now

		response := catalogResponses.ToCatalogSummaryResponse(item)
		require.NotNil(t, response)

		assert.Equal(t, item.ID, response.ID)
		assert.Equal(t, item.OrganizationID, response.OrganizationID)
		assert.Equal(t, string(item.ItemType), response.ItemType)
		assert.Equal(t, item.Name, response.Name)
		assert.True(t, item.BasePrice.Equal(response.BasePrice))
		assert.Equal(t, item.Currency, response.Currency)
		assert.Equal(t, item.IsActive, response.IsActive)
		assert.Equal(t, string(item.Visibility), response.Visibility)
		assert.Equal(t, item.CreatedAt, response.CreatedAt)
	})

	t.Run("handles nil item", func(t *testing.T) {
		response := catalogResponses.ToCatalogSummaryResponse(nil)
		assert.Nil(t, response)
	})
}

func TestToCatalogItemResponseList(t *testing.T) {
	t.Run("converts list of catalog items", func(t *testing.T) {
		now := time.Now()
		items := []*catalogModels.CatalogItem{
			{
				OrganizationID: "org-123",
				ItemType:       catalogModels.CatalogItemTypeProduct,
				Name:           "Organic Tomatoes",
				BasePrice:      decimal.NewFromFloat(25.50),
			},
			{
				OrganizationID: "org-123",
				ItemType:       catalogModels.CatalogItemTypeService,
				Name:           "Farm Consultation",
				BasePrice:      decimal.NewFromFloat(150.00),
			},
		}
		items[0].ID = "item-1"
		items[0].CreatedAt = now
		items[1].ID = "item-2"
		items[1].CreatedAt = now

		responses, err := catalogResponses.ToCatalogItemResponseList(items)
		require.NoError(t, err)
		require.Len(t, responses, 2)

		assert.Equal(t, items[0].ID, responses[0].ID)
		assert.Equal(t, items[0].Name, responses[0].Name)
		assert.Equal(t, items[1].ID, responses[1].ID)
		assert.Equal(t, items[1].Name, responses[1].Name)
	})

	t.Run("handles empty list", func(t *testing.T) {
		responses, err := catalogResponses.ToCatalogItemResponseList([]*catalogModels.CatalogItem{})
		assert.NoError(t, err)
		assert.Empty(t, responses)
	})

	t.Run("handles nil list", func(t *testing.T) {
		responses, err := catalogResponses.ToCatalogItemResponseList(nil)
		assert.NoError(t, err)
		assert.Empty(t, responses)
	})
}

func TestToProductResponseList(t *testing.T) {
	t.Run("converts list of products", func(t *testing.T) {
		now := time.Now()
		products := []*catalogModels.Product{
			{
				CatalogItem: catalogModels.CatalogItem{
					OrganizationID: "org-123",
					ItemType:       catalogModels.CatalogItemTypeProduct,
					Name:           "Organic Tomatoes",
					BasePrice:      decimal.NewFromFloat(25.50),
				},
				Perishable: true,
			},
			{
				CatalogItem: catalogModels.CatalogItem{
					OrganizationID: "org-123",
					ItemType:       catalogModels.CatalogItemTypeProduct,
					Name:           "Seeds",
					BasePrice:      decimal.NewFromFloat(10.00),
				},
				Perishable: false,
			},
		}
		products[0].ID = "product-1"
		products[0].CreatedAt = now
		products[1].ID = "product-2"
		products[1].CreatedAt = now

		responses, err := catalogResponses.ToProductResponseList(products)
		require.NoError(t, err)
		require.Len(t, responses, 2)

		assert.Equal(t, products[0].ID, responses[0].ID)
		assert.Equal(t, products[0].Perishable, responses[0].Perishable)
		assert.Equal(t, products[1].ID, responses[1].ID)
		assert.Equal(t, products[1].Perishable, responses[1].Perishable)
	})

	t.Run("handles empty list", func(t *testing.T) {
		responses, err := catalogResponses.ToProductResponseList([]*catalogModels.Product{})
		assert.NoError(t, err)
		assert.Empty(t, responses)
	})
}

func TestBulkOperationResponse(t *testing.T) {
	t.Run("creates and manipulates bulk operation response", func(t *testing.T) {
		response := catalogResponses.NewBulkOperationResponse(10, 8, 2)

		assert.Equal(t, 10, response.TotalItems)
		assert.Equal(t, 8, response.SuccessCount)
		assert.Equal(t, 2, response.FailureCount)
		assert.Empty(t, response.SuccessItems)
		assert.Empty(t, response.FailureItems)
		assert.Empty(t, response.ErrorMessages)

		response.AddSuccessItem("item-1")
		response.AddSuccessItem("item-2")
		response.AddFailureItem("item-3", "Item not found")
		response.AddFailureItem("item-4", "Permission denied")

		assert.Len(t, response.SuccessItems, 2)
		assert.Len(t, response.FailureItems, 2)
		assert.Len(t, response.ErrorMessages, 2)
		assert.Contains(t, response.SuccessItems, "item-1")
		assert.Contains(t, response.SuccessItems, "item-2")
		assert.Contains(t, response.FailureItems, "item-3")
		assert.Contains(t, response.FailureItems, "item-4")
		assert.Contains(t, response.ErrorMessages, "Item not found")
		assert.Contains(t, response.ErrorMessages, "Permission denied")
	})
}
