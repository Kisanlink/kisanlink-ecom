package taxation

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/shopspring/decimal"
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
	base.BaseModel
	OrgID    string          `json:"org_id" gorm:"type:varchar(255);not null;index"`  // Organization that sets the tax rate
	TaxType  TaxType         `json:"tax_type" gorm:"type:varchar(20);not null;index"` // Type of tax
	Rate     decimal.Decimal `json:"rate" gorm:"type:decimal(5,4);not null"`          // Tax rate as percentage (e.g., 18.00 for 18%)
	IsActive bool            `json:"is_active" gorm:"default:true"`                   // Whether this tax rate is currently active

	// Applicability
	EntityType  string `json:"entity_type" gorm:"type:varchar(50);index"`  // "product", "service", "labour", "all"
	Category    string `json:"category" gorm:"type:varchar(100);index"`    // Product category this applies to
	Subcategory string `json:"subcategory" gorm:"type:varchar(100);index"` // Product subcategory
	HSNCode     string `json:"hsn_code" gorm:"type:varchar(20);index"`     // HSN/SAC code for GST

	// Validity period
	ValidFrom *time.Time `json:"valid_from" gorm:"type:timestamp"`
	ValidTo   *time.Time `json:"valid_to" gorm:"type:timestamp"`

	// Metadata
	Description string `json:"description" gorm:"type:text"`
	Metadata    string `json:"metadata" gorm:"type:jsonb"` // Additional tax metadata
}

// TableName returns the table name for GORM
func (TaxRate) TableName() string {
	return "tax_rates"
}

// NewTaxRate creates a new TaxRate instance
func NewTaxRate(orgID string, taxType TaxType, rate decimal.Decimal) *TaxRate {
	return &TaxRate{
		BaseModel: *base.NewBaseModel("TAX", "medium"),
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
func (t *TaxRate) CalculateTax(baseAmount decimal.Decimal) decimal.Decimal {
	if !t.IsValid() {
		return decimal.Zero
	}
	return baseAmount.Mul(t.Rate).Div(decimal.NewFromInt(100))
}

// TaxRule represents complex tax rules and exemptions
type TaxRule struct {
	base.BaseModel
	OrgID       string `json:"org_id" gorm:"type:varchar(255);not null;index"`
	Name        string `json:"name" gorm:"type:varchar(255);not null"`
	Description string `json:"description" gorm:"type:text"`
	RuleType    string `json:"rule_type" gorm:"type:varchar(50);not null"` // "exemption", "reduction", "threshold"
	RuleData    string `json:"rule_data" gorm:"type:jsonb"`                // Rule-specific data
	Priority    int    `json:"priority" gorm:"default:0"`                  // Higher priority rules are applied first
	IsActive    bool   `json:"is_active" gorm:"default:true"`

	// Conditions
	Conditions string `json:"conditions" gorm:"type:jsonb"` // When this rule applies
}

// TableName returns the table name for GORM
func (TaxRule) TableName() string {
	return "tax_rules"
}

// NewTaxRule creates a new TaxRule instance
func NewTaxRule(orgID, name, ruleType string) *TaxRule {
	return &TaxRule{
		BaseModel: *base.NewBaseModel("TAXRULE", "medium"),
		OrgID:     orgID,
		Name:      name,
		RuleType:  ruleType,
		IsActive:  true,
		Priority:  0,
	}
}

// TaxExemption represents tax exemptions
type TaxExemption struct {
	base.BaseModel
	OrgID       string `json:"org_id" gorm:"type:varchar(255);not null;index"`
	ExemptionID string `json:"exemption_id" gorm:"type:varchar(255);not null;unique"`
	Name        string `json:"name" gorm:"type:varchar(255);not null"`
	Description string `json:"description" gorm:"type:text"`

	// Exemption details
	ExemptionType string          `json:"exemption_type" gorm:"type:varchar(50);not null"` // "full", "partial", "threshold"
	ExemptionRate decimal.Decimal `json:"exemption_rate" gorm:"type:decimal(5,4)"`         // Exemption rate (0.0 to 100.0)

	// Applicability
	EntityType string `json:"entity_type" gorm:"type:varchar(50);index"` // "product", "service", "labour", "all"
	Category   string `json:"category" gorm:"type:varchar(100);index"`   // Product category
	HSNCode    string `json:"hsn_code" gorm:"type:varchar(20);index"`    // HSN/SAC code

	// Validity
	ValidFrom *time.Time `json:"valid_from" gorm:"type:timestamp"`
	ValidTo   *time.Time `json:"valid_to" gorm:"type:timestamp"`
	IsActive  bool       `json:"is_active" gorm:"default:true"`
}

// TableName returns the table name for GORM
func (TaxExemption) TableName() string {
	return "tax_exemptions"
}

// NewTaxExemption creates a new TaxExemption instance
func NewTaxExemption(orgID, exemptionID, name, exemptionType string) *TaxExemption {
	return &TaxExemption{
		BaseModel:     *base.NewBaseModel("EXEMPT", "medium"),
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
	BaseAmount   decimal.Decimal            `json:"base_amount"`   // Amount before tax
	TaxAmount    decimal.Decimal            `json:"tax_amount"`    // Total tax amount
	TotalAmount  decimal.Decimal            `json:"total_amount"`  // Amount after tax
	TaxBreakdown map[string]decimal.Decimal `json:"tax_breakdown"` // Breakdown by tax type
	Exemptions   []string                   `json:"exemptions"`    // Applied exemptions
	AppliedRules []string                   `json:"applied_rules"` // Applied tax rules
}

// NewTaxCalculation creates a new TaxCalculation instance
func NewTaxCalculation(baseAmount decimal.Decimal) *TaxCalculation {
	return &TaxCalculation{
		BaseAmount:   baseAmount,
		TaxAmount:    decimal.Zero,
		TotalAmount:  baseAmount,
		TaxBreakdown: make(map[string]decimal.Decimal),
		Exemptions:   make([]string, 0),
		AppliedRules: make([]string, 0),
	}
}

// AddTax adds a tax amount for a specific tax type
func (tc *TaxCalculation) AddTax(taxType string, amount decimal.Decimal) {
	tc.TaxBreakdown[taxType] = amount
	tc.TaxAmount = tc.TaxAmount.Add(amount)
	tc.TotalAmount = tc.BaseAmount.Add(tc.TaxAmount)
}

// AddExemption adds an applied exemption
func (tc *TaxCalculation) AddExemption(exemptionID string) {
	tc.Exemptions = append(tc.Exemptions, exemptionID)
}

// AddRule adds an applied tax rule
func (tc *TaxCalculation) AddRule(ruleID string) {
	tc.AppliedRules = append(tc.AppliedRules, ruleID)
}
