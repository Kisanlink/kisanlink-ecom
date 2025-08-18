package taxation

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// TaxType represents the type of tax
type TaxType string

const (
	TaxTypeGST     TaxType = "gst"     // Goods and Services Tax (India)
	TaxTypeCGST    TaxType = "cgst"    // Central GST
	TaxTypeSGST    TaxType = "sgst"    // State GST
	TaxTypeIGST    TaxType = "igst"    // Integrated GST
	TaxTypeVAT     TaxType = "vat"     // Value Added Tax
	TaxTypeSales   TaxType = "sales"   // Sales Tax
	TaxTypeExcise  TaxType = "excise"  // Excise Duty
	TaxTypeCustoms TaxType = "customs" // Customs Duty
	TaxTypeOther   TaxType = "other"   // Other taxes
)

// TaxRate represents a tax rate configuration
type TaxRate struct {
	*base.BaseModel
	OrgID    string  `json:"org_id" gorm:"type:varchar(255);not null;index"`  // Organization that sets the tax rate
	TaxType  TaxType `json:"tax_type" gorm:"type:varchar(20);not null;index"` // Type of tax
	Rate     float64 `json:"rate" gorm:"type:decimal(5,4);not null"`          // Tax rate as percentage (e.g., 18.00 for 18%)
	IsActive bool    `json:"is_active" gorm:"default:true"`                   // Whether this tax rate is currently active

	// Applicability
	EntityType  string `json:"entity_type" gorm:"type:varchar(50);index"`  // "product", "service", "labour", "all"
	Category    string `json:"category" gorm:"type:varchar(100);index"`    // Product category this applies to
	Subcategory string `json:"subcategory" gorm:"type:varchar(100);index"` // Product subcategory
	HSNCode     string `json:"hsn_code" gorm:"type:varchar(20);index"`     // HSN/SAC code for GST

	// Validity period
	ValidFrom *time.Time `json:"valid_from" gorm:"type:timestamp"`
	ValidTo   *time.Time `json:"valid_to" gorm:"type:timestamp"`

	// Metadata
	Description string                 `json:"description" gorm:"type:text"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:jsonb"` // Additional tax metadata
}

// NewTaxRate creates a new TaxRate instance
func NewTaxRate(orgID string, taxType TaxType, rate float64) *TaxRate {
	return &TaxRate{
		BaseModel: base.NewBaseModel("TAX", hash.Medium),
		OrgID:     orgID,
		TaxType:   taxType,
		Rate:      rate,
		IsActive:  true,
	}
}

// IsValid checks if the tax rate is currently valid
func (t *TaxRate) IsValid() bool {
	now := time.Now()

	if !t.IsActive {
		return false
	}

	if t.ValidFrom != nil && now.Before(*t.ValidFrom) {
		return false
	}

	if t.ValidTo != nil && now.After(*t.ValidTo) {
		return false
	}

	return true
}

// CalculateTax calculates the tax amount for a given base amount
func (t *TaxRate) CalculateTax(baseAmount float64) float64 {
	if !t.IsValid() {
		return 0.0
	}
	return (baseAmount * t.Rate) / 100.0
}

// TaxRule represents complex tax rules and exemptions
type TaxRule struct {
	*base.BaseModel
	OrgID       string                 `json:"org_id" gorm:"type:varchar(255);not null;index"`
	Name        string                 `json:"name" gorm:"type:varchar(255);not null"`
	Description string                 `json:"description" gorm:"type:text"`
	RuleType    string                 `json:"rule_type" gorm:"type:varchar(50);not null"` // "exemption", "reduction", "threshold"
	RuleData    map[string]interface{} `json:"rule_data" gorm:"type:jsonb"`                // Rule-specific data
	Priority    int                    `json:"priority" gorm:"default:0"`                  // Higher priority rules are applied first
	IsActive    bool                   `json:"is_active" gorm:"default:true"`

	// Conditions
	Conditions map[string]interface{} `json:"conditions" gorm:"type:jsonb"` // When this rule applies
}

// NewTaxRule creates a new TaxRule instance
func NewTaxRule(orgID, name, ruleType string) *TaxRule {
	return &TaxRule{
		BaseModel: base.NewBaseModel("TAXRULE", hash.Medium),
		OrgID:     orgID,
		Name:      name,
		RuleType:  ruleType,
		IsActive:  true,
		Priority:  0,
	}
}

// TaxExemption represents tax exemptions
type TaxExemption struct {
	*base.BaseModel
	OrgID       string `json:"org_id" gorm:"type:varchar(255);not null;index"`
	ExemptionID string `json:"exemption_id" gorm:"type:varchar(255);not null;unique"`
	Name        string `json:"name" gorm:"type:varchar(255);not null"`
	Description string `json:"description" gorm:"type:text"`

	// Exemption details
	ExemptionType string  `json:"exemption_type" gorm:"type:varchar(50);not null"` // "full", "partial", "threshold"
	ExemptionRate float64 `json:"exemption_rate" gorm:"type:decimal(5,4)"`         // Exemption rate (0.0 to 100.0)

	// Applicability
	EntityType string `json:"entity_type" gorm:"type:varchar(50);index"` // "product", "service", "labour", "all"
	Category   string `json:"category" gorm:"type:varchar(100);index"`   // Product category
	HSNCode    string `json:"hsn_code" gorm:"type:varchar(20);index"`    // HSN/SAC code

	// Validity
	ValidFrom *time.Time `json:"valid_from" gorm:"type:timestamp"`
	ValidTo   *time.Time `json:"valid_to" gorm:"type:timestamp"`
	IsActive  bool       `json:"is_active" gorm:"default:true"`
}

// NewTaxExemption creates a new TaxExemption instance
func NewTaxExemption(orgID, exemptionID, name, exemptionType string) *TaxExemption {
	return &TaxExemption{
		BaseModel:     base.NewBaseModel("EXEMPT", hash.Medium),
		OrgID:         orgID,
		ExemptionID:   exemptionID,
		Name:          name,
		ExemptionType: exemptionType,
		IsActive:      true,
	}
}

// IsValid checks if the tax exemption is currently valid
func (t *TaxExemption) IsValid() bool {
	now := time.Now()

	if !t.IsActive {
		return false
	}

	if t.ValidFrom != nil && now.Before(*t.ValidFrom) {
		return false
	}

	if t.ValidTo != nil && now.After(*t.ValidTo) {
		return false
	}

	return true
}

// TaxCalculation represents a tax calculation result
type TaxCalculation struct {
	BaseAmount   float64            `json:"base_amount"`   // Amount before tax
	TaxAmount    float64            `json:"tax_amount"`    // Total tax amount
	TotalAmount  float64            `json:"total_amount"`  // Amount after tax
	TaxBreakdown map[string]float64 `json:"tax_breakdown"` // Breakdown by tax type
	Exemptions   []string           `json:"exemptions"`    // Applied exemptions
	AppliedRules []string           `json:"applied_rules"` // Applied tax rules
}

// NewTaxCalculation creates a new TaxCalculation instance
func NewTaxCalculation(baseAmount float64) *TaxCalculation {
	return &TaxCalculation{
		BaseAmount:   baseAmount,
		TaxAmount:    0.0,
		TotalAmount:  baseAmount,
		TaxBreakdown: make(map[string]float64),
		Exemptions:   make([]string, 0),
		AppliedRules: make([]string, 0),
	}
}

// AddTax adds a tax amount for a specific tax type
func (tc *TaxCalculation) AddTax(taxType string, amount float64) {
	tc.TaxBreakdown[taxType] = amount
	tc.TaxAmount += amount
	tc.TotalAmount = tc.BaseAmount + tc.TaxAmount
}

// AddExemption adds an applied exemption
func (tc *TaxCalculation) AddExemption(exemptionID string) {
	tc.Exemptions = append(tc.Exemptions, exemptionID)
}

// AddRule adds an applied tax rule
func (tc *TaxCalculation) AddRule(ruleID string) {
	tc.AppliedRules = append(tc.AppliedRules, ruleID)
}
