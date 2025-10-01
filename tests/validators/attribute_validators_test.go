package validators

import (
	"testing"

	"kisanlink-ecom/internal/validators"

	"github.com/stretchr/testify/assert"
)

func TestAttributeValidator_ValidateProductAttributes(t *testing.T) {
	validator := validators.NewAttributeValidator()

	tests := []struct {
		name        string
		attributes  map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil attributes should pass",
			attributes:  nil,
			expectError: false,
		},
		{
			name: "valid product attributes",
			attributes: map[string]interface{}{
				"sku":    "PROD-001",
				"brand":  "TestBrand",
				"weight": 2.5,
				"dimensions": map[string]interface{}{
					"length": 10.0,
					"width":  5.0,
					"height": 3.0,
					"unit":   "cm",
				},
				"perishable":    true,
				"shelf_life":    7,
				"organic":       true,
				"certification": []string{"ISO9001", "Organic"},
			},
			expectError: false,
		},
		{
			name: "valid attributes without SKU",
			attributes: map[string]interface{}{
				"brand": "TestBrand",
			},
			expectError: false,
		},
		{
			name: "invalid SKU format",
			attributes: map[string]interface{}{
				"sku": "PROD@001#",
			},
			expectError: true,
			errorMsg:    "sku must be alphanumeric with hyphens and underscores",
		},
		{
			name: "negative weight",
			attributes: map[string]interface{}{
				"sku":    "PROD-001",
				"weight": -1.0,
			},
			expectError: true,
			errorMsg:    "weight must be greater than or equal to 0",
		},
		{
			name: "invalid dimensions - negative length",
			attributes: map[string]interface{}{
				"sku": "PROD-001",
				"dimensions": map[string]interface{}{
					"length": -10.0,
					"width":  5.0,
					"height": 3.0,
					"unit":   "cm",
				},
			},
			expectError: true,
			errorMsg:    "invalid dimensions",
		},
		{
			name: "invalid dimension unit",
			attributes: map[string]interface{}{
				"sku": "PROD-001",
				"dimensions": map[string]interface{}{
					"length": 10.0,
					"width":  5.0,
					"height": 3.0,
					"unit":   "invalid",
				},
			},
			expectError: true,
			errorMsg:    "unit must be one of",
		},
		{
			name: "too many certifications",
			attributes: map[string]interface{}{
				"sku": "PROD-001",
				"certification": []string{
					"cert1", "cert2", "cert3", "cert4", "cert5",
					"cert6", "cert7", "cert8", "cert9", "cert10",
					"cert11", // This exceeds the limit of 10
				},
			},
			expectError: true,
			errorMsg:    "maximum 10 certifications allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateProductAttributes(tt.attributes)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAttributeValidator_ValidateServiceAttributes(t *testing.T) {
	validator := validators.NewAttributeValidator()

	tests := []struct {
		name        string
		attributes  map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil attributes should pass",
			attributes:  nil,
			expectError: false,
		},
		{
			name: "valid service attributes",
			attributes: map[string]interface{}{
				"duration": "2h30m",
				"sla": map[string]interface{}{
					"response_time_minutes":   30,
					"resolution_time_hours":   4,
					"availability_percentage": 99.9,
					"support_hours":           "24/7",
				},
				"location": "Remote",
				"skills":   []string{"consulting", "analysis"},
				"service_area": map[string]interface{}{
					"type":  "city",
					"value": "Mumbai",
				},
				"equipment":     []string{"laptop", "software"},
				"certification": []string{"PMP", "ITIL"},
			},
			expectError: false,
		},
		{
			name: "missing required duration",
			attributes: map[string]interface{}{
				"location": "Remote",
			},
			expectError: true,
			errorMsg:    "duration is required for services",
		},
		{
			name: "invalid duration format",
			attributes: map[string]interface{}{
				"duration": "invalid",
			},
			expectError: true,
			errorMsg:    "duration must be in format like",
		},
		{
			name: "invalid SLA - negative response time",
			attributes: map[string]interface{}{
				"duration": "2h",
				"sla": map[string]interface{}{
					"response_time_minutes": -30,
				},
			},
			expectError: true,
			errorMsg:    "response_time_minutes must be greater than or equal to 0",
		},
		{
			name: "invalid SLA - availability over 100%",
			attributes: map[string]interface{}{
				"duration": "2h",
				"sla": map[string]interface{}{
					"availability_percentage": 150.0,
				},
			},
			expectError: true,
			errorMsg:    "availability_percentage must be between 0 and 100",
		},
		{
			name: "too many skills",
			attributes: map[string]interface{}{
				"duration": "2h",
				"skills": []string{
					"skill1", "skill2", "skill3", "skill4", "skill5",
					"skill6", "skill7", "skill8", "skill9", "skill10",
					"skill11", "skill12", "skill13", "skill14", "skill15",
					"skill16", "skill17", "skill18", "skill19", "skill20",
					"skill21", // This exceeds the limit of 20
				},
			},
			expectError: true,
			errorMsg:    "maximum 20 skills allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateServiceAttributes(tt.attributes)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAttributeValidator_ValidateLabourAttributes(t *testing.T) {
	validator := validators.NewAttributeValidator()

	tests := []struct {
		name        string
		attributes  map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil attributes should pass",
			attributes:  nil,
			expectError: false,
		},
		{
			name: "valid labour attributes",
			attributes: map[string]interface{}{
				"skills":     []string{"farming", "irrigation"},
				"experience": 5,
				"unit_rate": map[string]interface{}{
					"amount":   75.50,
					"currency": "INR",
				},
				"rate_type": "hourly",
				"location": map[string]interface{}{
					"type":  "city",
					"value": "Pune",
				},
				"languages":     []string{"Hindi", "English"},
				"tools":         []string{"tractor", "plow"},
				"certification": []string{"AgriCert"},
			},
			expectError: false,
		},
		{
			name: "missing required skills",
			attributes: map[string]interface{}{
				"experience": 5,
				"unit_rate": map[string]interface{}{
					"amount":   75.50,
					"currency": "INR",
				},
				"rate_type": "hourly",
			},
			expectError: true,
			errorMsg:    "skills are required for labour",
		},
		{
			name: "empty skills array",
			attributes: map[string]interface{}{
				"skills": []string{},
				"unit_rate": map[string]interface{}{
					"amount":   75.50,
					"currency": "INR",
				},
				"rate_type": "hourly",
			},
			expectError: true,
			errorMsg:    "skills are required for labour",
		},
		{
			name: "zero unit rate",
			attributes: map[string]interface{}{
				"skills": []string{"farming"},
				"unit_rate": map[string]interface{}{
					"amount":   0.0,
					"currency": "INR",
				},
				"rate_type": "hourly",
			},
			expectError: true,
			errorMsg:    "unit_rate amount must be greater than 0",
		},
		{
			name: "invalid currency",
			attributes: map[string]interface{}{
				"skills": []string{"farming"},
				"unit_rate": map[string]interface{}{
					"amount":   75.50,
					"currency": "INVALID",
				},
				"rate_type": "hourly",
			},
			expectError: true,
			errorMsg:    "unit_rate currency must be a valid 3-character ISO code",
		},
		{
			name: "invalid rate type",
			attributes: map[string]interface{}{
				"skills": []string{"farming"},
				"unit_rate": map[string]interface{}{
					"amount":   75.50,
					"currency": "INR",
				},
				"rate_type": "invalid",
			},
			expectError: true,
			errorMsg:    "rate_type must be one of",
		},
		{
			name: "negative experience",
			attributes: map[string]interface{}{
				"skills":     []string{"farming"},
				"experience": -1,
				"unit_rate": map[string]interface{}{
					"amount":   75.50,
					"currency": "INR",
				},
				"rate_type": "hourly",
			},
			expectError: true,
			errorMsg:    "experience must be greater than or equal to 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateLabourAttributes(tt.attributes)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAttributeValidator_ValidateContractAttributes(t *testing.T) {
	validator := validators.NewAttributeValidator()

	tests := []struct {
		name        string
		attributes  map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil attributes should pass",
			attributes:  nil,
			expectError: false,
		},
		{
			name: "valid contract attributes",
			attributes: map[string]interface{}{
				"term":       "fixed",
				"duration":   12,
				"start_date": "2024-01-01T00:00:00Z",
				"end_date":   "2024-12-31T00:00:00Z",
				"terms":      "Contract terms and conditions",
				"deliverables": []map[string]interface{}{
					{
						"id":          "DEL-001",
						"name":        "Phase 1 Delivery",
						"description": "Initial phase deliverable",
						"due_date":    "2024-06-30T00:00:00Z",
						"status":      "pending",
						"value": map[string]interface{}{
							"amount":   50000.0,
							"currency": "INR",
						},
					},
				},
				"milestones": []map[string]interface{}{
					{
						"id":          "MIL-001",
						"name":        "Milestone 1",
						"description": "First milestone",
						"due_date":    "2024-03-31T00:00:00Z",
						"status":      "pending",
						"payment": map[string]interface{}{
							"amount":   25000.0,
							"currency": "INR",
						},
					},
				},
				"payment_terms": map[string]interface{}{
					"type":       "milestone",
					"percentage": 50.0,
					"due_days":   30,
					"method":     []string{"bank_transfer"},
					"currency":   "INR",
				},
				"renewal_terms": map[string]interface{}{
					"auto_renewal":   false,
					"notice_period":  30,
					"renewal_period": 12,
				},
			},
			expectError: false,
		},
		{
			name: "missing required term",
			attributes: map[string]interface{}{
				"duration": 12,
			},
			expectError: true,
			errorMsg:    "term is required for contracts",
		},
		{
			name: "zero duration",
			attributes: map[string]interface{}{
				"term":     "fixed",
				"duration": 0,
			},
			expectError: true,
			errorMsg:    "duration must be greater than 0",
		},
		{
			name: "invalid contract term",
			attributes: map[string]interface{}{
				"term":     "invalid-term",
				"duration": 12,
			},
			expectError: true,
			errorMsg:    "term must be a valid contract term",
		},
		{
			name: "end date before start date",
			attributes: map[string]interface{}{
				"term":       "fixed",
				"duration":   12,
				"start_date": "2024-12-31T00:00:00Z",
				"end_date":   "2024-01-01T00:00:00Z",
			},
			expectError: true,
			errorMsg:    "end_date must be after start_date",
		},
		{
			name: "invalid deliverable - missing ID",
			attributes: map[string]interface{}{
				"term":     "fixed",
				"duration": 12,
				"deliverables": []map[string]interface{}{
					{
						"name": "Phase 1 Delivery",
						"value": map[string]interface{}{
							"amount":   50000.0,
							"currency": "INR",
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "deliverable ID is required",
		},
		{
			name: "invalid milestone - zero payment",
			attributes: map[string]interface{}{
				"term":     "fixed",
				"duration": 12,
				"milestones": []map[string]interface{}{
					{
						"id":   "MIL-001",
						"name": "Milestone 1",
						"payment": map[string]interface{}{
							"amount":   0.0,
							"currency": "INR",
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "milestone payment must be greater than 0",
		},
		{
			name: "invalid payment terms - percentage over 100",
			attributes: map[string]interface{}{
				"term":     "fixed",
				"duration": 12,
				"payment_terms": map[string]interface{}{
					"type":       "advance",
					"percentage": 150.0,
					"currency":   "INR",
				},
			},
			expectError: true,
			errorMsg:    "payment percentage must be between 0 and 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateContractAttributes(tt.attributes)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAttributeValidator_ValidationHelpers(t *testing.T) {
	validator := validators.NewAttributeValidator()

	t.Run("SKU validation", func(t *testing.T) {
		validSKUs := []string{"PROD-001", "SKU_123", "ABC123", "TEST-ITEM_001"}
		invalidSKUs := []string{"PROD@001", "SKU 123", "ABC#123", "", repeat("a", 51)}

		for _, sku := range validSKUs {
			attrs := map[string]interface{}{"sku": sku}
			err := validator.ValidateProductAttributes(attrs)
			assert.NoError(t, err, "SKU %s should be valid", sku)
		}

		for _, sku := range invalidSKUs {
			attrs := map[string]interface{}{"sku": sku}
			err := validator.ValidateProductAttributes(attrs)
			if sku == "" {
				// Empty SKU is now valid (optional field)
				assert.NoError(t, err, "Empty SKU should be valid (optional field)")
			} else {
				assert.Error(t, err, "SKU %s should be invalid", sku)
			}
		}
	})

	t.Run("Duration validation", func(t *testing.T) {
		validDurations := []string{"2h", "30m", "1h30m", "24h", "90m"}
		invalidDurations := []string{"", "2hours", "30mins", "1.5h", "invalid"}

		for _, duration := range validDurations {
			attrs := map[string]interface{}{"duration": duration}
			err := validator.ValidateServiceAttributes(attrs)
			assert.NoError(t, err, "Duration %s should be valid", duration)
		}

		for _, duration := range invalidDurations {
			attrs := map[string]interface{}{"duration": duration}
			err := validator.ValidateServiceAttributes(attrs)
			assert.Error(t, err, "Duration %s should be invalid", duration)
		}
	})

	t.Run("Currency validation", func(t *testing.T) {
		validCurrencies := []string{"INR", "USD", "EUR", "GBP"}
		invalidCurrencies := []string{"", "IN", "INRS", "123"}

		for _, currency := range validCurrencies {
			attrs := map[string]interface{}{
				"skills": []string{"farming"},
				"unit_rate": map[string]interface{}{
					"amount":   100.0,
					"currency": currency,
				},
				"rate_type": "hourly",
			}
			err := validator.ValidateLabourAttributes(attrs)
			assert.NoError(t, err, "Currency %s should be valid", currency)
		}

		for _, currency := range invalidCurrencies {
			attrs := map[string]interface{}{
				"skills": []string{"farming"},
				"unit_rate": map[string]interface{}{
					"amount":   100.0,
					"currency": currency,
				},
				"rate_type": "hourly",
			}
			err := validator.ValidateLabourAttributes(attrs)
			assert.Error(t, err, "Currency %s should be invalid", currency)
		}
	})

	t.Run("Rate type validation", func(t *testing.T) {
		validRateTypes := []string{"hourly", "daily", "weekly", "monthly", "per_hour", "per_day", "per_week", "per_month"}
		invalidRateTypes := []string{"", "yearly", "HOURLY"}

		for _, rateType := range validRateTypes {
			attrs := map[string]interface{}{
				"skills": []string{"farming"},
				"unit_rate": map[string]interface{}{
					"amount":   100.0,
					"currency": "INR",
				},
				"rate_type": rateType,
			}
			err := validator.ValidateLabourAttributes(attrs)
			assert.NoError(t, err, "Rate type %s should be valid", rateType)
		}

		for _, rateType := range invalidRateTypes {
			attrs := map[string]interface{}{
				"skills": []string{"farming"},
				"unit_rate": map[string]interface{}{
					"amount":   100.0,
					"currency": "INR",
				},
				"rate_type": rateType,
			}
			err := validator.ValidateLabourAttributes(attrs)
			assert.Error(t, err, "Rate type %s should be invalid", rateType)
		}
	})

	t.Run("Contract term validation", func(t *testing.T) {
		validTerms := []string{"fixed", "renewable", "indefinite", "project-based", "milestone-based"}
		invalidTerms := []string{"", "temporary", "FIXED", "custom"}

		for _, term := range validTerms {
			attrs := map[string]interface{}{
				"term":     term,
				"duration": 12,
			}
			err := validator.ValidateContractAttributes(attrs)
			assert.NoError(t, err, "Contract term %s should be valid", term)
		}

		for _, term := range invalidTerms {
			attrs := map[string]interface{}{
				"term":     term,
				"duration": 12,
			}
			err := validator.ValidateContractAttributes(attrs)
			if term == "" {
				assert.Error(t, err, "Empty term should be invalid")
				assert.Contains(t, err.Error(), "term is required")
			} else {
				assert.Error(t, err, "Contract term %s should be invalid", term)
			}
		}
	})
}

// Helper function to repeat strings (for testing long strings)
func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
