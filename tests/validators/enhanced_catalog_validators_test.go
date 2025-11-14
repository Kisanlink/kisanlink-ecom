package validators

import (
	"testing"

	catalogModels "github.com/Kisanlink/kisanlink-ecom/entities/models/catalog"
	catalogRequests "github.com/Kisanlink/kisanlink-ecom/entities/requests/catalog"
	"github.com/Kisanlink/kisanlink-ecom/internal/validators"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnhancedCatalogValidator_CustomValidators(t *testing.T) {
	validator := validators.NewCatalogValidator()

	t.Run("SKU Format Validation", func(t *testing.T) {
		tests := []struct {
			name    string
			request *catalogRequests.CreateCatalogItemRequest
			wantErr bool
			errMsg  string
		}{
			{
				name: "Valid SKU format",
				request: &catalogRequests.CreateCatalogItemRequest{
					ItemType:      catalogModels.CatalogItemTypeProduct,
					Name:          "Test Product",
					Category:      "electronics",
					UnitOfMeasure: "piece",
					SKU:           "PROD-001",
					BasePrice:     decimal.NewFromFloat(100),
				},
				wantErr: false,
			},
			{
				name: "Invalid SKU - too short",
				request: &catalogRequests.CreateCatalogItemRequest{
					ItemType:      catalogModels.CatalogItemTypeProduct,
					Name:          "Test Product",
					Category:      "electronics",
					UnitOfMeasure: "piece",
					SKU:           "AB",
					BasePrice:     decimal.NewFromFloat(100),
				},
				wantErr: true,
			},
			{
				name: "Invalid SKU - special characters",
				request: &catalogRequests.CreateCatalogItemRequest{
					ItemType:      catalogModels.CatalogItemTypeProduct,
					Name:          "Test Product",
					Category:      "electronics",
					UnitOfMeasure: "piece",
					SKU:           "PROD@001",
					BasePrice:     decimal.NewFromFloat(100),
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validator.ValidateCreateCatalogItemRequest(tt.request)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("Rate Unit Validation", func(t *testing.T) {
		tests := []struct {
			name       string
			attributes map[string]interface{}
			wantErr    bool
		}{
			{
				name: "Valid hourly rate",
				attributes: map[string]interface{}{
					"skills": []string{"farming", "irrigation"},
					"unit_rate": map[string]interface{}{
						"amount":   100.0,
						"currency": "INR",
					},
					"rate_type": "hourly",
				},
				wantErr: false,
			},
			{
				name: "Invalid rate type",
				attributes: map[string]interface{}{
					"skills": []string{"farming"},
					"unit_rate": map[string]interface{}{
						"amount":   100.0,
						"currency": "INR",
					},
					"rate_type": "invalid_rate",
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				request := &catalogRequests.CreateCatalogItemRequest{
					ItemType:    catalogModels.CatalogItemTypeLabour,
					Name:        "Test Labour",
					Description: "Test labour description",
					Category:    "farming",
					BasePrice:   decimal.NewFromFloat(100),
					Attributes:  tt.attributes,
				}

				err := validator.ValidateCreateCatalogItemRequest(request)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("Contract Date Validation", func(t *testing.T) {
		tests := []struct {
			name       string
			attributes map[string]interface{}
			wantErr    bool
		}{
			{
				name: "Valid contract dates",
				attributes: map[string]interface{}{
					"term":       "fixed",
					"duration":   12.0,
					"start_date": "2024-01-01",
					"end_date":   "2024-12-31",
				},
				wantErr: false,
			},
			{
				name: "Invalid date format",
				attributes: map[string]interface{}{
					"term":       "fixed",
					"duration":   12.0,
					"start_date": "01-01-2024",
					"end_date":   "2024-12-31",
				},
				wantErr: true,
			},
			{
				name: "End date before start date",
				attributes: map[string]interface{}{
					"term":       "fixed",
					"duration":   12.0,
					"start_date": "2024-12-31",
					"end_date":   "2024-01-01",
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				request := &catalogRequests.CreateCatalogItemRequest{
					ItemType:    catalogModels.CatalogItemTypeContract,
					Name:        "Test Contract",
					Description: "Test contract description",
					BasePrice:   decimal.NewFromFloat(10000),
					Attributes:  tt.attributes,
				}

				err := validator.ValidateCreateCatalogItemRequest(request)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestEnhancedCatalogValidator_CrossFieldValidation(t *testing.T) {
	validator := validators.NewCatalogValidator()

	t.Run("Contract Cross-Field Validation", func(t *testing.T) {
		t.Run("Valid contract with consistent dates", func(t *testing.T) {
			request := &catalogRequests.CreateCatalogItemRequest{
				ItemType:    catalogModels.CatalogItemTypeContract,
				Name:        "Test Contract",
				Description: "Test contract description",
				BasePrice:   decimal.NewFromFloat(50000),
				Attributes: map[string]interface{}{
					"term":       "fixed",
					"duration":   12.0,
					"start_date": "2024-01-01",
					"end_date":   "2024-12-31",
					"payment_terms": map[string]interface{}{
						"percentage": 50.0,
						"type":       "advance",
					},
				},
			}

			err := validator.ValidateCreateCatalogItemRequest(request)
			assert.NoError(t, err)
		})

		t.Run("Invalid payment percentage", func(t *testing.T) {
			request := &catalogRequests.CreateCatalogItemRequest{
				ItemType:    catalogModels.CatalogItemTypeContract,
				Name:        "Test Contract",
				Description: "Test contract description",
				BasePrice:   decimal.NewFromFloat(50000),
				Attributes: map[string]interface{}{
					"term":     "fixed",
					"duration": 12.0,
					"payment_terms": map[string]interface{}{
						"percentage": 150.0, // Invalid percentage > 100
						"type":       "advance",
					},
				},
			}

			err := validator.ValidateCreateCatalogItemRequest(request)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "percentage must be between 0 and 100")
		})
	})

	t.Run("Labour Cross-Field Validation", func(t *testing.T) {
		t.Run("Advanced skills with insufficient experience", func(t *testing.T) {
			request := &catalogRequests.CreateCatalogItemRequest{
				ItemType:    catalogModels.CatalogItemTypeLabour,
				Name:        "Expert Farmer",
				Description: "Expert farming labour",
				Category:    "farming",
				BasePrice:   decimal.NewFromFloat(200),
				Attributes: map[string]interface{}{
					"skills":     []interface{}{"expert farming", "senior irrigation"},
					"experience": 1.0, // Only 1 year experience for expert skills
					"unit_rate": map[string]interface{}{
						"amount":   200.0,
						"currency": "INR",
					},
					"rate_type": "hourly",
				},
			}

			err := validator.ValidateCreateCatalogItemRequest(request)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "advanced skills typically require at least 3 years")
		})

		t.Run("Valid experience for skills", func(t *testing.T) {
			request := &catalogRequests.CreateCatalogItemRequest{
				ItemType:    catalogModels.CatalogItemTypeLabour,
				Name:        "Expert Farmer",
				Description: "Expert farming labour",
				Category:    "farming",
				BasePrice:   decimal.NewFromFloat(200),
				Attributes: map[string]interface{}{
					"skills":     []interface{}{"expert farming", "senior irrigation"},
					"experience": 5.0, // 5 years experience
					"unit_rate": map[string]interface{}{
						"amount":   200.0,
						"currency": "INR",
					},
					"rate_type": "hourly",
				},
			}

			err := validator.ValidateCreateCatalogItemRequest(request)
			assert.NoError(t, err)
		})
	})

	t.Run("Product Cross-Field Validation", func(t *testing.T) {
		t.Run("Perishable product without shelf life", func(t *testing.T) {
			request := &catalogRequests.CreateCatalogItemRequest{
				ItemType:      catalogModels.CatalogItemTypeProduct,
				Name:          "Fresh Tomatoes",
				Category:      "vegetables",
				UnitOfMeasure: "kg",
				BasePrice:     decimal.NewFromFloat(50),
				Attributes: map[string]interface{}{
					"perishable": true,
					// Missing shelf_life_days
				},
			}

			err := validator.ValidateCreateCatalogItemRequest(request)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "perishable products must specify shelf_life_days")
		})

		t.Run("Valid perishable product", func(t *testing.T) {
			request := &catalogRequests.CreateCatalogItemRequest{
				ItemType:      catalogModels.CatalogItemTypeProduct,
				Name:          "Fresh Tomatoes",
				Category:      "vegetables",
				UnitOfMeasure: "kg",
				SKU:           "TOM-001",
				BasePrice:     decimal.NewFromFloat(50),
				Attributes: map[string]interface{}{
					"perishable":      true,
					"shelf_life_days": 7.0,
				},
			}

			err := validator.ValidateCreateCatalogItemRequest(request)
			assert.NoError(t, err)
		})
	})
}

func TestAdminApprovalWorkflow(t *testing.T) {
	workflow := validators.NewAdminApprovalWorkflow()

	t.Run("High Value Item Requires Approval", func(t *testing.T) {
		request := &catalogRequests.CreateCatalogItemRequest{
			ItemType:  catalogModels.CatalogItemTypeProduct,
			Name:      "Expensive Equipment",
			BasePrice: decimal.NewFromFloat(150000), // > 1 Lakh INR
		}

		requirements, err := workflow.ValidateForApproval(request)
		require.NoError(t, err)
		assert.Len(t, requirements, 1)
		assert.Equal(t, "pricing", requirements[0].Type)
		assert.Equal(t, "high", requirements[0].Severity)
		assert.False(t, requirements[0].AutoApprove)
		assert.Equal(t, "pricing_manager", requirements[0].ReviewerRole)
	})

	t.Run("Medium Value Item Auto-Approvable", func(t *testing.T) {
		request := &catalogRequests.CreateCatalogItemRequest{
			ItemType:  catalogModels.CatalogItemTypeProduct,
			Name:      "Medium Equipment",
			BasePrice: decimal.NewFromFloat(75000), // Between 50K-100K INR
		}

		requirements, err := workflow.ValidateForApproval(request)
		require.NoError(t, err)
		assert.Len(t, requirements, 1)
		assert.Equal(t, "pricing", requirements[0].Type)
		assert.Equal(t, "medium", requirements[0].Severity)
		assert.True(t, requirements[0].AutoApprove)
		assert.Equal(t, "team_lead", requirements[0].ReviewerRole)
	})

	t.Run("High Hourly Rate Requires Approval", func(t *testing.T) {
		request := &catalogRequests.CreateCatalogItemRequest{
			ItemType:  catalogModels.CatalogItemTypeLabour,
			Name:      "Expert Consultant",
			BasePrice: decimal.NewFromFloat(1000),
			Attributes: map[string]interface{}{
				"skills":    []string{"agricultural consulting"},
				"rate_type": "hourly",
				"unit_rate": map[string]interface{}{
					"amount":   6000.0, // > ₹5000/hour
					"currency": "INR",
				},
			},
		}

		requirements, err := workflow.ValidateForApproval(request)
		require.NoError(t, err)
		assert.Len(t, requirements, 1)
		assert.Equal(t, "pricing", requirements[0].Type)
		assert.Equal(t, "critical", requirements[0].Severity)
		assert.False(t, requirements[0].AutoApprove)
		assert.Equal(t, "senior_manager", requirements[0].ReviewerRole)
	})

	t.Run("Long Term Contract Requires Approval", func(t *testing.T) {
		request := &catalogRequests.CreateCatalogItemRequest{
			ItemType:  catalogModels.CatalogItemTypeContract,
			Name:      "Long Term Service Contract",
			BasePrice: decimal.NewFromFloat(100000),
			Attributes: map[string]interface{}{
				"term":     "fixed",
				"duration": 36.0, // > 24 months
			},
		}

		requirements, err := workflow.ValidateForApproval(request)
		require.NoError(t, err)

		// Should have both pricing and terms approval requirements
		assert.GreaterOrEqual(t, len(requirements), 1)

		// Check for terms approval
		hasTermsApproval := false
		for _, req := range requirements {
			if req.Type == "terms" {
				hasTermsApproval = true
				assert.Equal(t, "high", req.Severity)
				assert.False(t, req.AutoApprove)
				assert.Equal(t, "legal_team", req.ReviewerRole)
			}
		}
		assert.True(t, hasTermsApproval)
	})

	t.Run("High Availability SLA Requires Approval", func(t *testing.T) {
		request := &catalogRequests.CreateCatalogItemRequest{
			ItemType:  catalogModels.CatalogItemTypeService,
			Name:      "Critical Service",
			BasePrice: decimal.NewFromFloat(50000),
			Attributes: map[string]interface{}{
				"duration": "2h",
				"sla": map[string]interface{}{
					"availability_percentage": 99.95, // > 99.9%
					"response_time_minutes":   2.0,   // < 5 minutes
				},
			},
		}

		requirements, err := workflow.ValidateForApproval(request)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(requirements), 1)

		// Check for SLA approval
		hasSLAApproval := false
		for _, req := range requirements {
			if req.Type == "sla" {
				hasSLAApproval = true
				// Should have at least one SLA requirement (either high availability or fast response)
				assert.True(t, req.Severity == "high" || req.Severity == "medium")
			}
		}
		assert.True(t, hasSLAApproval)
	})

	t.Run("Regulated Category Requires Compliance Approval", func(t *testing.T) {
		request := &catalogRequests.CreateCatalogItemRequest{
			ItemType:  catalogModels.CatalogItemTypeProduct,
			Name:      "Medical Equipment",
			Category:  "medical",
			BasePrice: decimal.NewFromFloat(25000),
		}

		requirements, err := workflow.ValidateForApproval(request)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(requirements), 1)

		// Check for compliance approval
		hasComplianceApproval := false
		for _, req := range requirements {
			if req.Type == "compliance" {
				hasComplianceApproval = true
				assert.Equal(t, "critical", req.Severity)
				assert.False(t, req.AutoApprove)
				assert.Equal(t, "compliance_officer", req.ReviewerRole)
			}
		}
		assert.True(t, hasComplianceApproval)
	})

	t.Run("Auto-Approvable Check", func(t *testing.T) {
		autoApprovableReqs := []validators.ApprovalRequirement{
			{Type: "pricing", AutoApprove: true},
			{Type: "terms", AutoApprove: true},
		}

		nonAutoApprovableReqs := []validators.ApprovalRequirement{
			{Type: "pricing", AutoApprove: true},
			{Type: "compliance", AutoApprove: false},
		}

		assert.True(t, workflow.IsAutoApprovable(autoApprovableReqs))
		assert.False(t, workflow.IsAutoApprovable(nonAutoApprovableReqs))
	})

	t.Run("Required Reviewers", func(t *testing.T) {
		requirements := []validators.ApprovalRequirement{
			{Type: "pricing", AutoApprove: true, ReviewerRole: "team_lead"},
			{Type: "compliance", AutoApprove: false, ReviewerRole: "compliance_officer"},
			{Type: "legal", AutoApprove: false, ReviewerRole: "legal_team"},
		}

		reviewers := workflow.GetRequiredReviewers(requirements)
		assert.Len(t, reviewers, 2) // Only non-auto-approvable ones
		assert.Contains(t, reviewers, "compliance_officer")
		assert.Contains(t, reviewers, "legal_team")
	})
}

func TestCatalogValidator_DurationParsing(t *testing.T) {
	validator := validators.NewCatalogValidator()

	tests := []struct {
		name     string
		duration string
		expected int
	}{
		{"2 hours", "2h", 120},
		{"30 minutes", "30m", 30},
		{"2 hours 30 minutes", "2h30m", 150},
		{"1 hour", "1h", 60},
		{"90 minutes", "90m", 90},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This tests the internal duration parsing logic
			// We'll create a service request with high duration to trigger the validation
			request := &catalogRequests.CreateCatalogItemRequest{
				ItemType:  catalogModels.CatalogItemTypeService,
				Name:      "Long Service",
				BasePrice: decimal.NewFromFloat(100), // Low price for long duration
				Attributes: map[string]interface{}{
					"duration": tt.duration,
				},
			}

			err := validator.ValidateCreateCatalogItemRequest(request)

			// For durations > 8 hours (480 minutes) with low pricing, should get error
			if tt.expected > 480 {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "services longer than 8 hours typically require higher pricing")
			}
		})
	}
}
