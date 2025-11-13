package validators

import (
	"testing"

	"kisanlink-ecom/entities/models/catalog"
	catalogRequests "kisanlink-ecom/entities/requests/catalog"
	"kisanlink-ecom/internal/validators"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCatalogValidator_ValidateCreateCatalogItemRequest_WithAttributes(t *testing.T) {
	validator := validators.NewCatalogValidator()

	tests := []struct {
		name        string
		request     *catalogRequests.CreateCatalogItemRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid product with attributes",
			request: &catalogRequests.CreateCatalogItemRequest{
				ItemType:      catalog.CatalogItemTypeProduct,
				Category:      "vegetables",
				Name:          "Organic Tomatoes",
				Description:   "Fresh organic tomatoes",
				BasePrice:     decimal.NewFromFloat(25.50),
				Currency:      "INR",
				UnitOfMeasure: "kg",
				Attributes: map[string]interface{}{
					"sku":             "TOM-ORG-001",
					"brand":           "OrganicFarm",
					"perishable":      true,
					"shelf_life_days": 7.0,
				},
			},
			expectError: false,
		},
		{
			name: "valid service with attributes",
			request: &catalogRequests.CreateCatalogItemRequest{
				ItemType:    catalog.CatalogItemTypeService,
				Name:        "Agricultural Consultation",
				Description: "Expert agricultural consultation service",
				BasePrice:   decimal.NewFromFloat(500.00),
				Currency:    "INR",
				Attributes: map[string]interface{}{
					"duration": "2h",
					"sla": map[string]interface{}{
						"response_time_minutes":   30,
						"resolution_time_hours":   4,
						"availability_percentage": 99.5,
					},
					"skills": []string{"crop_management", "soil_analysis"},
				},
			},
			expectError: false,
		},
		{
			name: "valid labour with attributes",
			request: &catalogRequests.CreateCatalogItemRequest{
				ItemType:    catalog.CatalogItemTypeLabour,
				Category:    "farming",
				Name:        "Experienced Farm Worker",
				Description: "Skilled farm worker with 5 years experience",
				BasePrice:   decimal.NewFromFloat(75.00),
				Currency:    "INR",
				Attributes: map[string]interface{}{
					"skills":     []string{"planting", "harvesting", "irrigation"},
					"experience": 5,
					"unit_rate": map[string]interface{}{
						"amount":   75.00,
						"currency": "INR",
					},
					"rate_type": "hourly",
				},
			},
			expectError: false,
		},
		{
			name: "valid contract with attributes",
			request: &catalogRequests.CreateCatalogItemRequest{
				ItemType:    catalog.CatalogItemTypeContract,
				Name:        "Annual Farming Contract",
				Description: "12-month farming service contract",
				BasePrice:   decimal.NewFromFloat(100000.00),
				Currency:    "INR",
				Attributes: map[string]interface{}{
					"term":     "fixed",
					"duration": 12,
					"terms":    "Annual farming service contract with deliverables",
				},
			},
			expectError: false,
		},
		{
			name: "product with invalid SKU in attributes",
			request: &catalogRequests.CreateCatalogItemRequest{
				ItemType:      catalog.CatalogItemTypeProduct,
				Category:      "vegetables",
				Name:          "Organic Tomatoes",
				BasePrice:     decimal.NewFromFloat(25.50),
				Currency:      "INR",
				UnitOfMeasure: "kg",
				Attributes: map[string]interface{}{
					"sku": "INVALID@SKU#",
				},
			},
			expectError: true,
			errorMsg:    "sku must be alphanumeric with hyphens and underscores",
		},
		{
			name: "service with missing duration in attributes",
			request: &catalogRequests.CreateCatalogItemRequest{
				ItemType:    catalog.CatalogItemTypeService,
				Name:        "Agricultural Consultation",
				Description: "Expert agricultural consultation service",
				BasePrice:   decimal.NewFromFloat(500.00),
				Currency:    "INR",
				Attributes: map[string]interface{}{
					"location": "Remote",
				},
			},
			expectError: true,
			errorMsg:    "duration is required for services",
		},
		{
			name: "labour with missing skills in attributes",
			request: &catalogRequests.CreateCatalogItemRequest{
				ItemType:    catalog.CatalogItemTypeLabour,
				Category:    "farming",
				Name:        "Farm Worker",
				Description: "Farm worker",
				BasePrice:   decimal.NewFromFloat(75.00),
				Currency:    "INR",
				Attributes: map[string]interface{}{
					"experience": 5,
					"unit_rate": map[string]interface{}{
						"amount":   75.00,
						"currency": "INR",
					},
					"rate_type": "hourly",
				},
			},
			expectError: true,
			errorMsg:    "skills are required for labour",
		},
		{
			name: "contract with missing term in attributes",
			request: &catalogRequests.CreateCatalogItemRequest{
				ItemType:    catalog.CatalogItemTypeContract,
				Name:        "Farming Contract",
				Description: "Farming service contract",
				BasePrice:   decimal.NewFromFloat(100000.00),
				Currency:    "INR",
				Attributes: map[string]interface{}{
					"duration": 12,
				},
			},
			expectError: true,
			errorMsg:    "term is required in attributes for contracts",
		},
		{
			name: "contract with invalid term in attributes",
			request: &catalogRequests.CreateCatalogItemRequest{
				ItemType:    catalog.CatalogItemTypeContract,
				Name:        "Farming Contract",
				Description: "Farming service contract",
				BasePrice:   decimal.NewFromFloat(100000.00),
				Currency:    "INR",
				Attributes: map[string]interface{}{
					"term":     "invalid-term",
					"duration": 12,
				},
			},
			expectError: true,
			errorMsg:    "term must be a valid contract term",
		},
		{
			name: "labour with invalid rate type in attributes",
			request: &catalogRequests.CreateCatalogItemRequest{
				ItemType:    catalog.CatalogItemTypeLabour,
				Category:    "farming",
				Name:        "Farm Worker",
				Description: "Farm worker",
				BasePrice:   decimal.NewFromFloat(75.00),
				Currency:    "INR",
				Attributes: map[string]interface{}{
					"skills": []string{"farming"},
					"unit_rate": map[string]interface{}{
						"amount":   75.00,
						"currency": "INR",
					},
					"rate_type": "invalid_rate",
				},
			},
			expectError: true,
			errorMsg:    "rate_type must be one of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCreateCatalogItemRequest(tt.request)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCatalogValidator_ValidateAttributesForType(t *testing.T) {
	validator := validators.NewCatalogValidator()

	t.Run("product attributes validation", func(t *testing.T) {
		validAttrs := map[string]interface{}{
			"sku":    "PROD-001",
			"brand":  "TestBrand",
			"weight": 2.5,
		}
		err := validator.ValidateAttributesForType(catalog.CatalogItemTypeProduct, validAttrs)
		assert.NoError(t, err)

		invalidAttrs := map[string]interface{}{
			"sku": "INVALID@SKU",
		}
		err = validator.ValidateAttributesForType(catalog.CatalogItemTypeProduct, invalidAttrs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sku must be alphanumeric")
	})

	t.Run("service attributes validation", func(t *testing.T) {
		validAttrs := map[string]interface{}{
			"duration": "2h30m",
			"location": "Remote",
		}
		err := validator.ValidateAttributesForType(catalog.CatalogItemTypeService, validAttrs)
		assert.NoError(t, err)

		invalidAttrs := map[string]interface{}{
			"duration": "invalid",
		}
		err = validator.ValidateAttributesForType(catalog.CatalogItemTypeService, invalidAttrs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duration must be in format")
	})

	t.Run("labour attributes validation", func(t *testing.T) {
		validAttrs := map[string]interface{}{
			"skills": []string{"farming", "irrigation"},
			"unit_rate": map[string]interface{}{
				"amount":   75.00,
				"currency": "INR",
			},
			"rate_type": "hourly",
		}
		err := validator.ValidateAttributesForType(catalog.CatalogItemTypeLabour, validAttrs)
		assert.NoError(t, err)

		invalidAttrs := map[string]interface{}{
			"skills": []string{},
		}
		err = validator.ValidateAttributesForType(catalog.CatalogItemTypeLabour, invalidAttrs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "skills are required")
	})

	t.Run("contract attributes validation", func(t *testing.T) {
		validAttrs := map[string]interface{}{
			"term":     "fixed",
			"duration": 12,
		}
		err := validator.ValidateAttributesForType(catalog.CatalogItemTypeContract, validAttrs)
		assert.NoError(t, err)

		invalidAttrs := map[string]interface{}{
			"term":     "",
			"duration": 12,
		}
		err = validator.ValidateAttributesForType(catalog.CatalogItemTypeContract, invalidAttrs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "term is required")
	})

	t.Run("unsupported catalog type", func(t *testing.T) {
		attrs := map[string]interface{}{"test": "value"}
		err := validator.ValidateAttributesForType("UNSUPPORTED", attrs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported catalog item type")
	})

	t.Run("nil attributes", func(t *testing.T) {
		err := validator.ValidateAttributesForType(catalog.CatalogItemTypeProduct, nil)
		assert.NoError(t, err)
	})
}

func TestCatalogValidator_ContractSpecificRules(t *testing.T) {
	validator := validators.NewCatalogValidator()

	tests := []struct {
		name        string
		request     *catalogRequests.CreateCatalogItemRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "contract with valid attributes",
			request: &catalogRequests.CreateCatalogItemRequest{
				ItemType:    catalog.CatalogItemTypeContract,
				Name:        "Test Contract",
				Description: "Test contract description",
				BasePrice:   decimal.NewFromFloat(100000.00),
				Currency:    "INR",
				Attributes: map[string]interface{}{
					"term":     "fixed",
					"duration": 12,
				},
			},
			expectError: false,
		},
		{
			name: "contract without description",
			request: &catalogRequests.CreateCatalogItemRequest{
				ItemType:  catalog.CatalogItemTypeContract,
				Name:      "Test Contract",
				BasePrice: decimal.NewFromFloat(100000.00),
				Currency:  "INR",
				Attributes: map[string]interface{}{
					"term":     "fixed",
					"duration": 12,
				},
			},
			expectError: true,
			errorMsg:    "description is required for contracts",
		},
		{
			name: "contract without attributes",
			request: &catalogRequests.CreateCatalogItemRequest{
				ItemType:    catalog.CatalogItemTypeContract,
				Name:        "Test Contract",
				Description: "Test contract description",
				BasePrice:   decimal.NewFromFloat(100000.00),
				Currency:    "INR",
			},
			expectError: true,
			errorMsg:    "attributes with term and duration are required for contracts",
		},
		{
			name: "contract without term in attributes",
			request: &catalogRequests.CreateCatalogItemRequest{
				ItemType:    catalog.CatalogItemTypeContract,
				Name:        "Test Contract",
				Description: "Test contract description",
				BasePrice:   decimal.NewFromFloat(100000.00),
				Currency:    "INR",
				Attributes: map[string]interface{}{
					"duration": 12,
				},
			},
			expectError: true,
			errorMsg:    "term is required in attributes for contracts",
		},
		{
			name: "contract without duration in attributes",
			request: &catalogRequests.CreateCatalogItemRequest{
				ItemType:    catalog.CatalogItemTypeContract,
				Name:        "Test Contract",
				Description: "Test contract description",
				BasePrice:   decimal.NewFromFloat(100000.00),
				Currency:    "INR",
				Attributes: map[string]interface{}{
					"term": "fixed",
				},
			},
			expectError: true,
			errorMsg:    "duration is required in attributes for contracts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCreateCatalogItemRequest(tt.request)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCatalogValidator_UpdateRequestValidation(t *testing.T) {
	validator := validators.NewCatalogValidator()

	t.Run("valid update request", func(t *testing.T) {
		name := "Updated Product"
		price := decimal.NewFromFloat(30.00)
		req := &catalogRequests.UpdateCatalogItemRequest{
			Name:      &name,
			BasePrice: &price,
		}

		err := validator.ValidateUpdateCatalogItemRequest(req)
		assert.NoError(t, err)
	})

	t.Run("update request with attributes", func(t *testing.T) {
		name := "Updated Product"
		req := &catalogRequests.UpdateCatalogItemRequest{
			Name: &name,
			Attributes: map[string]interface{}{
				"sku": "UPDATED-SKU",
			},
		}

		// For update requests, attribute validation is skipped at validator level
		// and should be handled at service layer with full context
		err := validator.ValidateUpdateCatalogItemRequest(req)
		assert.NoError(t, err)
	})

	t.Run("update request with negative price", func(t *testing.T) {
		price := decimal.NewFromFloat(-10.00)
		req := &catalogRequests.UpdateCatalogItemRequest{
			BasePrice: &price,
		}

		err := validator.ValidateUpdateCatalogItemRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "base_price must be greater than 0")
	})
}

func TestCatalogValidator_CatalogItemTypeValidation(t *testing.T) {
	validator := validators.NewCatalogValidator()

	validTypes := []catalog.CatalogItemType{
		catalog.CatalogItemTypeProduct,
		catalog.CatalogItemTypeService,
		catalog.CatalogItemTypeLabour,
		catalog.CatalogItemTypeContract,
	}

	for _, itemType := range validTypes {
		t.Run(string(itemType), func(t *testing.T) {
			req := &catalogRequests.CreateCatalogItemRequest{
				ItemType:    itemType,
				Name:        "Test Item",
				Description: "Test description",
				BasePrice:   decimal.NewFromFloat(100.00),
				Currency:    "INR",
			}

			// Add required attributes based on type
			switch itemType {
			case catalog.CatalogItemTypeProduct:
				req.Category = "test"
				req.UnitOfMeasure = "kg"
				req.Attributes = map[string]interface{}{
					"sku": "TEST-001",
				}
			case catalog.CatalogItemTypeService:
				req.Attributes = map[string]interface{}{
					"duration": "2h",
				}
			case catalog.CatalogItemTypeLabour:
				req.Category = "test"
				req.Attributes = map[string]interface{}{
					"skills": []string{"farming"},
					"unit_rate": map[string]interface{}{
						"amount":   100.00,
						"currency": "INR",
					},
					"rate_type": "hourly",
				}
			case catalog.CatalogItemTypeContract:
				req.Attributes = map[string]interface{}{
					"term":     "fixed",
					"duration": 12,
				}
			}

			err := validator.ValidateCreateCatalogItemRequest(req)
			assert.NoError(t, err, "Item type %s should be valid", itemType)
		})
	}

	t.Run("invalid item type", func(t *testing.T) {
		req := &catalogRequests.CreateCatalogItemRequest{
			ItemType:  "INVALID",
			Name:      "Test Item",
			BasePrice: decimal.NewFromFloat(100.00),
			Currency:  "INR",
		}

		err := validator.ValidateCreateCatalogItemRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "itemtype is invalid")
	})
}
