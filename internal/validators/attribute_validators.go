package validators

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"kisanlink-ecom/entities/models/catalog"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

// AttributeValidator provides validation for type-specific catalog attributes
type AttributeValidator struct {
	validator *validator.Validate
}

// NewAttributeValidator creates a new attribute validator
func NewAttributeValidator() *AttributeValidator {
	validate := validator.New()

	// Register custom validators for attributes
	validate.RegisterValidation("product_sku", validateProductSKU)
	validate.RegisterValidation("contract_term", validateContractTerm)
	validate.RegisterValidation("labour_skill", validateLabourSkill)
	validate.RegisterValidation("rate_type", validateRateType)
	validate.RegisterValidation("sla_response_time", validateSLAResponseTime)
	validate.RegisterValidation("duration_format", validateDurationFormat)
	validate.RegisterValidation("temperature_unit", validateTemperatureUnit)
	validate.RegisterValidation("dimension_unit", validateDimensionUnit)
	validate.RegisterValidation("currency_code", validateCurrencyCode)
	validate.RegisterValidation("geographic_type", validateGeographicType)
	validate.RegisterValidation("day_of_week", validateDayOfWeek)
	validate.RegisterValidation("payment_type", validatePaymentType)
	validate.RegisterValidation("penalty_type", validatePenaltyType)

	return &AttributeValidator{
		validator: validate,
	}
}

// ValidateProductAttributes validates product-specific attributes
func (av *AttributeValidator) ValidateProductAttributes(attributes map[string]interface{}) error {
	if attributes == nil {
		return nil
	}

	// Convert to ProductAttributes struct for validation
	var productAttrs catalog.ProductAttributes
	if err := av.mapToStruct(attributes, &productAttrs); err != nil {
		return fmt.Errorf("invalid product attributes structure: %w", err)
	}

	// SKU validation - SKU can be provided either in attributes or at the main request level
	// So it's not required in attributes, but if provided, it should be valid

	// Validate SKU format if provided
	if productAttrs.SKU != "" && !av.isValidSKU(productAttrs.SKU) {
		return fmt.Errorf("sku must be alphanumeric with hyphens and underscores, max 50 characters")
	}

	// Validate weight if provided
	if productAttrs.Weight != nil && productAttrs.Weight.LessThan(decimal.Zero) {
		return fmt.Errorf("weight must be greater than or equal to 0")
	}

	// Validate dimensions if provided
	if productAttrs.Dimensions != nil {
		if err := av.validateDimensions(productAttrs.Dimensions); err != nil {
			return fmt.Errorf("invalid dimensions: %w", err)
		}
	}

	// Validate shelf life for perishable products
	if productAttrs.Perishable && productAttrs.ShelfLife != nil && *productAttrs.ShelfLife <= 0 {
		return fmt.Errorf("shelf_life_days must be greater than 0 for perishable products")
	}

	// Validate storage temperature if provided
	if productAttrs.StorageTemp != nil {
		if err := av.validateTemperatureRange(productAttrs.StorageTemp); err != nil {
			return fmt.Errorf("invalid storage temperature: %w", err)
		}
	}

	// Validate variants if provided
	for i, variant := range productAttrs.Variants {
		if err := av.validateProductVariant(&variant); err != nil {
			return fmt.Errorf("invalid variant at index %d: %w", i, err)
		}
	}

	// Validate certification if provided
	if err := av.validateCertificationList(productAttrs.Certification); err != nil {
		return fmt.Errorf("invalid certification: %w", err)
	}

	return nil
}

// ValidateServiceAttributes validates service-specific attributes
func (av *AttributeValidator) ValidateServiceAttributes(attributes map[string]interface{}) error {
	if attributes == nil {
		return nil
	}

	// Convert to ServiceAttributes struct for validation
	var serviceAttrs catalog.ServiceAttributes
	if err := av.mapToStruct(attributes, &serviceAttrs); err != nil {
		return fmt.Errorf("invalid service attributes structure: %w", err)
	}

	// Validate required fields
	if serviceAttrs.Duration == "" {
		return fmt.Errorf("duration is required for services")
	}

	// Validate duration format
	if !av.isValidDuration(serviceAttrs.Duration) {
		return fmt.Errorf("duration must be in format like '2h', '30m', '1h30m'")
	}

	// Validate SLA fields
	if err := av.validateSLA(&serviceAttrs.SLA); err != nil {
		return fmt.Errorf("invalid SLA: %w", err)
	}

	// Validate skills if provided
	if err := av.validateSkillsList(serviceAttrs.Skills); err != nil {
		return fmt.Errorf("invalid skills: %w", err)
	}

	// Validate service area if provided
	if err := av.validateGeographicArea(&serviceAttrs.ServiceArea); err != nil {
		return fmt.Errorf("invalid service area: %w", err)
	}

	// Validate availability if provided
	for i, slot := range serviceAttrs.Availability {
		if err := av.validateTimeSlot(&slot); err != nil {
			return fmt.Errorf("invalid availability slot at index %d: %w", i, err)
		}
	}

	// Validate equipment list if provided
	if err := av.validateEquipmentList(serviceAttrs.Equipment); err != nil {
		return fmt.Errorf("invalid equipment: %w", err)
	}

	// Validate certification if provided
	if err := av.validateCertificationList(serviceAttrs.Certification); err != nil {
		return fmt.Errorf("invalid certification: %w", err)
	}

	return nil
}

// ValidateLabourAttributes validates labour-specific attributes
func (av *AttributeValidator) ValidateLabourAttributes(attributes map[string]interface{}) error {
	if attributes == nil {
		return nil
	}

	// Convert to LabourAttributes struct for validation
	var labourAttrs catalog.LabourAttributes
	if err := av.mapToStruct(attributes, &labourAttrs); err != nil {
		return fmt.Errorf("invalid labour attributes structure: %w", err)
	}

	// Validate required fields
	if len(labourAttrs.Skills) == 0 {
		return fmt.Errorf("skills are required for labour")
	}

	// Validate skills
	if err := av.validateSkillsList(labourAttrs.Skills); err != nil {
		return fmt.Errorf("invalid skills: %w", err)
	}

	// Validate unit rate
	if labourAttrs.UnitRate.Amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("unit_rate amount must be greater than 0")
	}

	// Validate currency (convert to uppercase for validation)
	currency := strings.ToUpper(labourAttrs.UnitRate.Currency)
	if !av.isValidCurrency(currency) {
		return fmt.Errorf("unit_rate currency must be a valid 3-character ISO code")
	}

	// Validate rate type
	validRateTypes := []string{"hourly", "daily", "weekly", "monthly", "per_hour", "per_day", "per_week", "per_month"}
	if !av.isValidRateType(labourAttrs.RateType, validRateTypes) {
		return fmt.Errorf("rate_type must be one of: %s", strings.Join(validRateTypes, ", "))
	}

	// Validate experience
	if labourAttrs.Experience < 0 {
		return fmt.Errorf("experience must be greater than or equal to 0")
	}

	// Validate location if provided
	if err := av.validateGeographicArea(&labourAttrs.Location); err != nil {
		return fmt.Errorf("invalid location: %w", err)
	}

	// Validate availability if provided
	for i, slot := range labourAttrs.Availability {
		if err := av.validateTimeSlot(&slot); err != nil {
			return fmt.Errorf("invalid availability slot at index %d: %w", i, err)
		}
	}

	// Validate languages if provided
	if err := av.validateLanguagesList(labourAttrs.Languages); err != nil {
		return fmt.Errorf("invalid languages: %w", err)
	}

	// Validate tools if provided
	if err := av.validateToolsList(labourAttrs.Tools); err != nil {
		return fmt.Errorf("invalid tools: %w", err)
	}

	// Validate certification if provided
	if err := av.validateCertificationList(labourAttrs.Certification); err != nil {
		return fmt.Errorf("invalid certification: %w", err)
	}

	return nil
}

// ValidateContractAttributes validates contract-specific attributes
func (av *AttributeValidator) ValidateContractAttributes(attributes map[string]interface{}) error {
	if attributes == nil {
		return nil
	}

	// Handle date strings by converting them to time.Time before struct mapping
	processedAttrs := make(map[string]interface{})
	for k, v := range attributes {
		processedAttrs[k] = v
	}

	// Convert date strings to time.Time for proper struct mapping
	if startDateStr, exists := attributes["start_date"]; exists {
		if dateStr, ok := startDateStr.(string); ok && dateStr != "" {
			if parsedDate, err := time.Parse("2006-01-02", dateStr); err == nil {
				processedAttrs["start_date"] = parsedDate
			}
		}
	}

	if endDateStr, exists := attributes["end_date"]; exists {
		if dateStr, ok := endDateStr.(string); ok && dateStr != "" {
			if parsedDate, err := time.Parse("2006-01-02", dateStr); err == nil {
				processedAttrs["end_date"] = parsedDate
			}
		}
	}

	// Convert to ContractAttributes struct for validation
	var contractAttrs catalog.ContractAttributes
	if err := av.mapToStruct(processedAttrs, &contractAttrs); err != nil {
		return fmt.Errorf("invalid contract attributes structure: %w", err)
	}

	// Validate required fields
	if contractAttrs.Term == "" {
		return fmt.Errorf("term is required for contracts")
	}

	if contractAttrs.Duration <= 0 {
		return fmt.Errorf("duration must be greater than 0")
	}

	// Validate term format
	if !av.isValidContractTerm(contractAttrs.Term) {
		return fmt.Errorf("term must be a valid contract term (e.g., 'fixed', 'renewable', 'indefinite')")
	}

	// Validate date consistency
	if !contractAttrs.StartDate.IsZero() && !contractAttrs.EndDate.IsZero() {
		if contractAttrs.EndDate.Before(contractAttrs.StartDate) {
			return fmt.Errorf("end_date must be after start_date")
		}
	}

	// Validate deliverables if provided
	for i, deliverable := range contractAttrs.Deliverables {
		if err := av.validateDeliverable(&deliverable); err != nil {
			return fmt.Errorf("invalid deliverable at index %d: %w", i, err)
		}
	}

	// Validate milestones if provided
	for i, milestone := range contractAttrs.Milestones {
		if err := av.validateMilestone(&milestone); err != nil {
			return fmt.Errorf("invalid milestone at index %d: %w", i, err)
		}
	}

	// Validate payment terms if provided
	if err := av.validatePaymentTerms(&contractAttrs.PaymentTerms); err != nil {
		return fmt.Errorf("invalid payment terms: %w", err)
	}

	// Validate penalties if provided
	for i, penalty := range contractAttrs.Penalties {
		if err := av.validatePenalty(&penalty); err != nil {
			return fmt.Errorf("invalid penalty at index %d: %w", i, err)
		}
	}

	// Validate renewal terms if provided
	if err := av.validateRenewalTerms(&contractAttrs.Renewals); err != nil {
		return fmt.Errorf("invalid renewal terms: %w", err)
	}

	return nil
}

// Helper validation methods

func (av *AttributeValidator) mapToStruct(data map[string]interface{}, target interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, target)
}

func (av *AttributeValidator) isValidSKU(sku string) bool {
	if len(sku) > 50 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, sku)
	return matched
}

func (av *AttributeValidator) isValidDuration(duration string) bool {
	// Validate duration format like "2h", "30m", "1h30m"
	matched, _ := regexp.MatchString(`^(\d+h)?(\d+m)?$`, duration)
	return matched && duration != ""
}

func (av *AttributeValidator) isValidContractTerm(term string) bool {
	validTerms := []string{"fixed", "renewable", "indefinite", "project-based", "milestone-based"}
	for _, validTerm := range validTerms {
		if term == validTerm {
			return true
		}
	}
	return false
}

func (av *AttributeValidator) isValidCurrency(currency string) bool {
	matched, _ := regexp.MatchString(`^[A-Z]{3}$`, currency)
	return len(currency) == 3 && matched
}

func (av *AttributeValidator) isValidRateType(rateType string, validTypes []string) bool {
	for _, validType := range validTypes {
		if rateType == validType {
			return true
		}
	}
	return false
}

func (av *AttributeValidator) validateDimensions(dimensions *catalog.Dimensions) error {
	if dimensions.Length.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("length must be greater than 0")
	}
	if dimensions.Width.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("width must be greater than 0")
	}
	if dimensions.Height.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("height must be greater than 0")
	}

	validUnits := []string{"cm", "m", "inch", "ft"}
	if !av.isValidUnit(dimensions.Unit, validUnits) {
		return fmt.Errorf("unit must be one of: %s", strings.Join(validUnits, ", "))
	}

	return nil
}

func (av *AttributeValidator) validateTemperatureRange(tempRange *catalog.TemperatureRange) error {
	if tempRange.Max <= tempRange.Min {
		return fmt.Errorf("max temperature must be greater than min temperature")
	}

	validUnits := []string{"celsius", "fahrenheit"}
	if !av.isValidUnit(tempRange.Unit, validUnits) {
		return fmt.Errorf("temperature unit must be one of: %s", strings.Join(validUnits, ", "))
	}

	return nil
}

func (av *AttributeValidator) validateProductVariant(variant *catalog.ProductVariant) error {
	if variant.Name == "" {
		return fmt.Errorf("variant name is required")
	}
	if variant.SKU == "" {
		return fmt.Errorf("variant SKU is required")
	}
	if !av.isValidSKU(variant.SKU) {
		return fmt.Errorf("variant SKU format is invalid")
	}
	if variant.Price.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("variant price must be greater than 0")
	}
	return nil
}

func (av *AttributeValidator) validateSLA(sla *catalog.SLAInfo) error {
	if sla.ResponseTime < 0 {
		return fmt.Errorf("response_time_minutes must be greater than or equal to 0")
	}
	if sla.ResolutionTime < 0 {
		return fmt.Errorf("resolution_time_hours must be greater than or equal to 0")
	}
	if sla.Availability < 0 || sla.Availability > 100 {
		return fmt.Errorf("availability_percentage must be between 0 and 100")
	}
	return nil
}

func (av *AttributeValidator) validateGeographicArea(area *catalog.GeographicArea) error {
	if area.Type == "" {
		return nil // Optional field
	}

	validTypes := []string{"city", "state", "country", "radius"}
	if !av.isValidUnit(area.Type, validTypes) {
		return fmt.Errorf("geographic type must be one of: %s", strings.Join(validTypes, ", "))
	}

	if area.Value == "" {
		return fmt.Errorf("geographic value is required when type is specified")
	}

	if area.Type == "radius" && area.Radius == nil {
		return fmt.Errorf("radius is required when type is 'radius'")
	}

	if area.Radius != nil && *area.Radius <= 0 {
		return fmt.Errorf("radius must be greater than 0")
	}

	return nil
}

func (av *AttributeValidator) validateTimeSlot(slot *catalog.TimeSlot) error {
	validDays := []string{"MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"}
	if !av.isValidUnit(slot.DayOfWeek, validDays) {
		return fmt.Errorf("day_of_week must be one of: %s", strings.Join(validDays, ", "))
	}

	if slot.EndTime.Before(slot.StartTime) {
		return fmt.Errorf("end_time must be after start_time")
	}

	return nil
}

func (av *AttributeValidator) validateDeliverable(deliverable *catalog.Deliverable) error {
	if deliverable.ID == "" {
		return fmt.Errorf("deliverable ID is required")
	}
	if deliverable.Name == "" {
		return fmt.Errorf("deliverable name is required")
	}
	if deliverable.Value.Amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("deliverable value must be greater than 0")
	}
	if !av.isValidCurrency(deliverable.Value.Currency) {
		return fmt.Errorf("deliverable currency must be a valid 3-character ISO code")
	}
	return nil
}

func (av *AttributeValidator) validateMilestone(milestone *catalog.Milestone) error {
	if milestone.ID == "" {
		return fmt.Errorf("milestone ID is required")
	}
	if milestone.Name == "" {
		return fmt.Errorf("milestone name is required")
	}
	if milestone.Payment.Amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("milestone payment must be greater than 0")
	}
	if !av.isValidCurrency(milestone.Payment.Currency) {
		return fmt.Errorf("milestone currency must be a valid 3-character ISO code")
	}
	return nil
}

func (av *AttributeValidator) validatePaymentTerms(terms *catalog.PaymentTerms) error {
	if terms.Type == "" {
		return nil // Optional field
	}

	validTypes := []string{"advance", "milestone", "completion", "monthly"}
	if !av.isValidUnit(terms.Type, validTypes) {
		return fmt.Errorf("payment type must be one of: %s", strings.Join(validTypes, ", "))
	}

	if terms.Percentage < 0 || terms.Percentage > 100 {
		return fmt.Errorf("payment percentage must be between 0 and 100")
	}

	if terms.DueDays < 0 {
		return fmt.Errorf("due_days must be greater than or equal to 0")
	}

	// Only validate currency if it's provided
	if terms.Currency != "" && !av.isValidCurrency(terms.Currency) {
		return fmt.Errorf("payment currency must be a valid 3-character ISO code")
	}

	return nil
}

func (av *AttributeValidator) validatePenalty(penalty *catalog.Penalty) error {
	validTypes := []string{"delay", "quality", "breach"}
	if !av.isValidUnit(penalty.Type, validTypes) {
		return fmt.Errorf("penalty type must be one of: %s", strings.Join(validTypes, ", "))
	}

	if penalty.Description == "" {
		return fmt.Errorf("penalty description is required")
	}

	if penalty.Amount.LessThan(decimal.Zero) {
		return fmt.Errorf("penalty amount must be greater than or equal to 0")
	}

	if penalty.Percentage != nil && (*penalty.Percentage < 0 || *penalty.Percentage > 100) {
		return fmt.Errorf("penalty percentage must be between 0 and 100")
	}

	return nil
}

func (av *AttributeValidator) validateRenewalTerms(terms *catalog.RenewalTerms) error {
	// Skip validation if renewal terms are not provided or empty
	if terms == nil || (terms.NoticePeriod == 0 && terms.RenewalPeriod == 0) {
		return nil
	}

	if terms.NoticePeriod < 0 {
		return fmt.Errorf("notice_period_days must be greater than or equal to 0")
	}

	if terms.RenewalPeriod < 0 {
		return fmt.Errorf("renewal_period_months must be greater than or equal to 0")
	}

	if terms.PriceAdjustment != nil && *terms.PriceAdjustment < -100 {
		return fmt.Errorf("price_adjustment_percentage must be greater than -100")
	}

	return nil
}

func (av *AttributeValidator) validateSkillsList(skills []string) error {
	if len(skills) > 20 {
		return fmt.Errorf("maximum 20 skills allowed")
	}

	for i, skill := range skills {
		if len(skill) == 0 || len(skill) > 100 {
			return fmt.Errorf("skill at index %d must be between 1 and 100 characters", i)
		}
		if strings.TrimSpace(skill) != skill {
			return fmt.Errorf("skill at index %d contains leading/trailing whitespace", i)
		}
	}

	return nil
}

func (av *AttributeValidator) validateCertificationList(certifications []string) error {
	if len(certifications) > 10 {
		return fmt.Errorf("maximum 10 certifications allowed")
	}

	for i, cert := range certifications {
		if len(cert) == 0 || len(cert) > 200 {
			return fmt.Errorf("certification at index %d must be between 1 and 200 characters", i)
		}
	}

	return nil
}

func (av *AttributeValidator) validateEquipmentList(equipment []string) error {
	if len(equipment) > 50 {
		return fmt.Errorf("maximum 50 equipment items allowed")
	}

	for i, item := range equipment {
		if len(item) == 0 || len(item) > 100 {
			return fmt.Errorf("equipment item at index %d must be between 1 and 100 characters", i)
		}
	}

	return nil
}

func (av *AttributeValidator) validateLanguagesList(languages []string) error {
	if len(languages) > 10 {
		return fmt.Errorf("maximum 10 languages allowed")
	}

	for i, lang := range languages {
		if len(lang) == 0 || len(lang) > 50 {
			return fmt.Errorf("language at index %d must be between 1 and 50 characters", i)
		}
	}

	return nil
}

func (av *AttributeValidator) validateToolsList(tools []string) error {
	if len(tools) > 30 {
		return fmt.Errorf("maximum 30 tools allowed")
	}

	for i, tool := range tools {
		if len(tool) == 0 || len(tool) > 100 {
			return fmt.Errorf("tool at index %d must be between 1 and 100 characters", i)
		}
	}

	return nil
}

func (av *AttributeValidator) isValidUnit(unit string, validUnits []string) bool {
	for _, validUnit := range validUnits {
		if unit == validUnit {
			return true
		}
	}
	return false
}

// Custom validator functions for go-playground/validator

func validateProductSKU(fl validator.FieldLevel) bool {
	sku := fl.Field().String()
	if len(sku) > 50 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, sku)
	return matched
}

func validateContractTerm(fl validator.FieldLevel) bool {
	term := fl.Field().String()
	validTerms := []string{"fixed", "renewable", "indefinite", "project-based", "milestone-based"}
	for _, validTerm := range validTerms {
		if term == validTerm {
			return true
		}
	}
	return false
}

func validateLabourSkill(fl validator.FieldLevel) bool {
	skill := fl.Field().String()
	return len(skill) > 0 && len(skill) <= 100 && strings.TrimSpace(skill) == skill
}

func validateRateType(fl validator.FieldLevel) bool {
	rateType := fl.Field().String()
	validTypes := []string{"hourly", "daily", "weekly", "monthly"}
	for _, validType := range validTypes {
		if rateType == validType {
			return true
		}
	}
	return false
}

func validateSLAResponseTime(fl validator.FieldLevel) bool {
	return fl.Field().Int() >= 0
}

func validateDurationFormat(fl validator.FieldLevel) bool {
	duration := fl.Field().String()
	matched, _ := regexp.MatchString(`^(\d+h)?(\d+m)?$`, duration)
	return matched && duration != ""
}

func validateTemperatureUnit(fl validator.FieldLevel) bool {
	unit := fl.Field().String()
	return unit == "celsius" || unit == "fahrenheit"
}

func validateDimensionUnit(fl validator.FieldLevel) bool {
	unit := fl.Field().String()
	validUnits := []string{"cm", "m", "inch", "ft"}
	for _, validUnit := range validUnits {
		if unit == validUnit {
			return true
		}
	}
	return false
}

func validateCurrencyCode(fl validator.FieldLevel) bool {
	currency := fl.Field().String()
	matched, _ := regexp.MatchString(`^[A-Z]{3}$`, currency)
	return len(currency) == 3 && matched
}

func validateGeographicType(fl validator.FieldLevel) bool {
	geoType := fl.Field().String()
	validTypes := []string{"city", "state", "country", "radius"}
	for _, validType := range validTypes {
		if geoType == validType {
			return true
		}
	}
	return false
}

func validateDayOfWeek(fl validator.FieldLevel) bool {
	day := fl.Field().String()
	validDays := []string{"MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"}
	for _, validDay := range validDays {
		if day == validDay {
			return true
		}
	}
	return false
}

func validatePaymentType(fl validator.FieldLevel) bool {
	paymentType := fl.Field().String()
	validTypes := []string{"advance", "milestone", "completion", "monthly"}
	for _, validType := range validTypes {
		if paymentType == validType {
			return true
		}
	}
	return false
}

func validatePenaltyType(fl validator.FieldLevel) bool {
	penaltyType := fl.Field().String()
	validTypes := []string{"delay", "quality", "breach"}
	for _, validType := range validTypes {
		if penaltyType == validType {
			return true
		}
	}
	return false
}
