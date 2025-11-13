package validators

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	catalogModels "kisanlink-ecom/entities/models/catalog"
	catalogRequests "kisanlink-ecom/entities/requests/catalog"
	"kisanlink-ecom/internal/utils"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

// CatalogValidator provides validation for catalog-related operations
type CatalogValidator struct {
	validator          *validator.Validate
	sanitizer          *utils.Sanitizer
	attributeValidator *AttributeValidator
}

// NewCatalogValidator creates a new catalog validator
func NewCatalogValidator() *CatalogValidator {
	validate := validator.New()
	sanitizer := utils.NewSanitizer(utils.DefaultSanitizerConfig())
	attributeValidator := NewAttributeValidator()

	// Register custom catalog validators
	validate.RegisterValidation("catalog_item_type", validateCatalogItemType)
	validate.RegisterValidation("visibility_type", validateVisibilityType)
	validate.RegisterValidation("category_name", validateCategoryName)
	validate.RegisterValidation("price_positive", validatePricePositive)
	validate.RegisterValidation("unit_of_measure", validateUnitOfMeasure)
	validate.RegisterValidation("image_urls", validateImageURLs)
	validate.RegisterValidation("tags_list", validateTagsList)

	// Register enhanced custom validators for catalog-specific fields
	validate.RegisterValidation("sku_format", validateSKUFormat)
	validate.RegisterValidation("rate_unit", validateRateUnit)
	validate.RegisterValidation("contract_date", validateContractDate)
	validate.RegisterValidation("currency_iso", validateCurrencyISO)
	validate.RegisterValidation("decimal_positive", validateDecimalPositive)
	validate.RegisterValidation("decimal_non_negative", validateDecimalNonNegative)
	validate.RegisterValidation("percentage", validatePercentage)
	validate.RegisterValidation("duration_format", validateDurationFormatValidator)
	validate.RegisterValidation("geographic_coordinates", validateGeographicCoordinates)
	validate.RegisterValidation("time_slot_valid", validateTimeSlotValid)
	validate.RegisterValidation("business_hours", validateBusinessHours)
	validate.RegisterValidation("skill_name", validateSkillName)
	validate.RegisterValidation("equipment_name", validateEquipmentName)
	validate.RegisterValidation("certification_name", validateCertificationName)
	validate.RegisterValidation("language_code", validateLanguageCode)
	validate.RegisterValidation("contract_term_type", validateContractTermType)
	validate.RegisterValidation("payment_method", validatePaymentMethod)
	validate.RegisterValidation("penalty_type", validatePenaltyTypeValidator)
	validate.RegisterValidation("milestone_status", validateMilestoneStatus)
	validate.RegisterValidation("deliverable_status", validateDeliverableStatus)

	return &CatalogValidator{
		validator:          validate,
		sanitizer:          sanitizer,
		attributeValidator: attributeValidator,
	}
}

// ValidateCreateCatalogItemRequest validates and sanitizes a create catalog item request
func (cv *CatalogValidator) ValidateCreateCatalogItemRequest(req *catalogRequests.CreateCatalogItemRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Sanitize string fields
	cv.sanitizeCreateCatalogItemRequest(req)

	// Validate struct
	if err := cv.validator.Struct(req); err != nil {
		return cv.formatValidationError(err)
	}

	// Additional business logic validation
	if err := cv.validateCreateCatalogItemBusinessRules(req); err != nil {
		return err
	}

	// Validate cross-field business rules
	if err := cv.ValidateCrossFieldRules(req); err != nil {
		return err
	}

	// Validate type-specific attributes
	return cv.validateTypeSpecificAttributes(req.ItemType, req.Attributes)
}

// ValidateUpdateCatalogItemRequest validates and sanitizes an update catalog item request
func (cv *CatalogValidator) ValidateUpdateCatalogItemRequest(req *catalogRequests.UpdateCatalogItemRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Sanitize string fields
	cv.sanitizeUpdateCatalogItemRequest(req)

	// Validate struct
	if err := cv.validator.Struct(req); err != nil {
		return cv.formatValidationError(err)
	}

	// Additional business logic validation
	if err := cv.validateUpdateCatalogItemBusinessRules(req); err != nil {
		return err
	}

	// Validate type-specific attributes if provided
	// Note: For updates, we need the item type from the existing item to validate attributes
	// This validation should be done at the service layer where we have access to the existing item
	if req.Attributes != nil {
		// For now, we'll skip attribute validation in update requests
		// This should be handled in the service layer with the full context
	}

	return nil
}

// ValidateCatalogFilter validates and sanitizes catalog filter parameters
func (cv *CatalogValidator) ValidateCatalogFilter(filter *catalogRequests.CatalogFilter) error {
	if filter == nil {
		return fmt.Errorf("filter cannot be nil")
	}

	// Sanitize filter fields
	cv.sanitizeCatalogFilter(filter)

	// Validate struct
	if err := cv.validator.Struct(filter); err != nil {
		return cv.formatValidationError(err)
	}

	// Additional validation
	return cv.validateCatalogFilterBusinessRules(filter)
}

// ValidateListCatalogItemsRequest validates list request parameters
func (cv *CatalogValidator) ValidateListCatalogItemsRequest(req *catalogRequests.ListCatalogItemsRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// Sanitize filter
	cv.sanitizeCatalogFilter(&req.CatalogFilter)

	// Validate struct
	if err := cv.validator.Struct(req); err != nil {
		return cv.formatValidationError(err)
	}

	return nil
}

// ValidateSearchCatalogRequest validates search request parameters
func (cv *CatalogValidator) ValidateSearchCatalogRequest(req *catalogRequests.SearchCatalogRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Sanitize search query
	req.Query = cv.sanitizer.SanitizeSearchQuery(req.Query)

	// Validate minimum query length
	if len(strings.TrimSpace(req.Query)) < 2 {
		return fmt.Errorf("search query must be at least 2 characters long")
	}

	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// Validate struct
	if err := cv.validator.Struct(req); err != nil {
		return cv.formatValidationError(err)
	}

	return nil
}

// ValidateBulkUpdateRequest validates bulk update request
func (cv *CatalogValidator) ValidateBulkUpdateRequest(req *catalogRequests.BulkUpdateCatalogRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Validate item IDs
	if len(req.ItemIDs) == 0 {
		return fmt.Errorf("at least one item ID is required")
	}

	if len(req.ItemIDs) > 100 {
		return fmt.Errorf("maximum 100 items can be updated at once")
	}

	// Validate at least one update field is provided
	if req.IsActive == nil && req.Visibility == nil && len(req.Tags) == 0 {
		return fmt.Errorf("at least one field to update must be provided")
	}

	// Sanitize tags
	if len(req.Tags) > 0 {
		for i, tag := range req.Tags {
			req.Tags[i] = cv.sanitizer.SanitizeString(tag)
		}
	}

	// Validate struct
	if err := cv.validator.Struct(req); err != nil {
		return cv.formatValidationError(err)
	}

	return nil
}

// Sanitization methods

func (cv *CatalogValidator) sanitizeCreateCatalogItemRequest(req *catalogRequests.CreateCatalogItemRequest) {
	req.Category = cv.sanitizer.SanitizeString(req.Category)
	req.Subcategory = cv.sanitizer.SanitizeString(req.Subcategory)
	req.Name = cv.sanitizer.SanitizeString(req.Name)
	req.Description = cv.sanitizer.SanitizeString(req.Description)
	req.SKU = cv.sanitizer.SanitizeString(req.SKU)
	req.UnitOfMeasure = cv.sanitizer.SanitizeString(req.UnitOfMeasure)
	req.Currency = strings.ToUpper(cv.sanitizer.SanitizeString(req.Currency))

	// Sanitize tags
	for i, tag := range req.Tags {
		req.Tags[i] = cv.sanitizer.SanitizeString(tag)
	}

	// Sanitize image URLs
	for i, img := range req.Images {
		req.Images[i] = cv.sanitizer.SanitizeURL(img)
	}
}

func (cv *CatalogValidator) sanitizeUpdateCatalogItemRequest(req *catalogRequests.UpdateCatalogItemRequest) {
	if req.Category != nil {
		sanitized := cv.sanitizer.SanitizeString(*req.Category)
		req.Category = &sanitized
	}
	if req.Subcategory != nil {
		sanitized := cv.sanitizer.SanitizeString(*req.Subcategory)
		req.Subcategory = &sanitized
	}
	if req.Name != nil {
		sanitized := cv.sanitizer.SanitizeString(*req.Name)
		req.Name = &sanitized
	}
	if req.Description != nil {
		sanitized := cv.sanitizer.SanitizeString(*req.Description)
		req.Description = &sanitized
	}
	if req.SKU != nil {
		sanitized := cv.sanitizer.SanitizeString(*req.SKU)
		req.SKU = &sanitized
	}
	if req.UnitOfMeasure != nil {
		sanitized := cv.sanitizer.SanitizeString(*req.UnitOfMeasure)
		req.UnitOfMeasure = &sanitized
	}
	if req.Currency != nil {
		sanitized := strings.ToUpper(cv.sanitizer.SanitizeString(*req.Currency))
		req.Currency = &sanitized
	}

	// Sanitize tags
	for i, tag := range req.Tags {
		req.Tags[i] = cv.sanitizer.SanitizeString(tag)
	}

	// Sanitize image URLs
	for i, img := range req.Images {
		req.Images[i] = cv.sanitizer.SanitizeURL(img)
	}
}

func (cv *CatalogValidator) sanitizeCatalogFilter(filter *catalogRequests.CatalogFilter) {
	if filter.Category != nil {
		sanitized := cv.sanitizer.SanitizeString(*filter.Category)
		filter.Category = &sanitized
	}
	if filter.Subcategory != nil {
		sanitized := cv.sanitizer.SanitizeString(*filter.Subcategory)
		filter.Subcategory = &sanitized
	}
	if filter.Search != nil {
		sanitized := cv.sanitizer.SanitizeSearchQuery(*filter.Search)
		filter.Search = &sanitized
	}
	if filter.OrganizationID != nil {
		sanitized := cv.sanitizer.SanitizeString(*filter.OrganizationID)
		filter.OrganizationID = &sanitized
	}

	// Sanitize tags
	for i, tag := range filter.Tags {
		filter.Tags[i] = cv.sanitizer.SanitizeString(tag)
	}
}

// Business rule validation methods

func (cv *CatalogValidator) validateCreateCatalogItemBusinessRules(req *catalogRequests.CreateCatalogItemRequest) error {
	// Validate price is positive
	if req.BasePrice.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("base_price must be greater than 0")
	}

	// Validate currency is 3 characters
	if req.Currency != "" && len(req.Currency) != 3 {
		return fmt.Errorf("currency must be a 3-character ISO code")
	}

	// Validate item type specific rules
	switch req.ItemType {
	case catalogModels.CatalogItemTypeProduct:
		return cv.validateProductSpecificRules(req)
	case catalogModels.CatalogItemTypeService:
		return cv.validateServiceSpecificRules(req)
	case catalogModels.CatalogItemTypeLabour:
		return cv.validateLabourSpecificRules(req)
	case catalogModels.CatalogItemTypeContract:
		return cv.validateContractSpecificRules(req)
	default:
		return fmt.Errorf("invalid item type: %s", req.ItemType)
	}
}

func (cv *CatalogValidator) validateUpdateCatalogItemBusinessRules(req *catalogRequests.UpdateCatalogItemRequest) error {
	// Validate price is positive if provided
	if req.BasePrice != nil && req.BasePrice.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("base_price must be greater than 0")
	}

	// Validate currency is 3 characters if provided
	if req.Currency != nil && len(*req.Currency) != 3 {
		return fmt.Errorf("currency must be a 3-character ISO code")
	}

	return nil
}

func (cv *CatalogValidator) validateCatalogFilterBusinessRules(filter *catalogRequests.CatalogFilter) error {
	// Validate price range
	if filter.MinPrice != nil && filter.MaxPrice != nil {
		if filter.MinPrice.GreaterThan(*filter.MaxPrice) {
			return fmt.Errorf("min_price cannot be greater than max_price")
		}
	}

	// Validate price values are non-negative
	if filter.MinPrice != nil && filter.MinPrice.LessThan(decimal.Zero) {
		return fmt.Errorf("min_price cannot be negative")
	}
	if filter.MaxPrice != nil && filter.MaxPrice.LessThan(decimal.Zero) {
		return fmt.Errorf("max_price cannot be negative")
	}

	return nil
}

func (cv *CatalogValidator) validateProductSpecificRules(req *catalogRequests.CreateCatalogItemRequest) error {
	// Products should have categories
	if req.Category == "" {
		return fmt.Errorf("category is required for products")
	}

	// Products should have unit of measure
	if req.UnitOfMeasure == "" {
		return fmt.Errorf("unit_of_measure is required for products")
	}

	// Validate SKU format if provided using custom validation
	if req.SKU != "" {
		if !cv.isValidSKUFormat(req.SKU) {
			return fmt.Errorf("sku format is invalid: must be 3-50 characters, alphanumeric with hyphens and underscores, starting with alphanumeric")
		}
	}

	return nil
}

func (cv *CatalogValidator) validateServiceSpecificRules(req *catalogRequests.CreateCatalogItemRequest) error {
	// Services should have descriptions
	if req.Description == "" {
		return fmt.Errorf("description is required for services")
	}

	return nil
}

func (cv *CatalogValidator) validateLabourSpecificRules(req *catalogRequests.CreateCatalogItemRequest) error {
	// Labour should have descriptions and categories
	if req.Description == "" {
		return fmt.Errorf("description is required for labour")
	}
	if req.Category == "" {
		return fmt.Errorf("category is required for labour")
	}

	return nil
}

func (cv *CatalogValidator) validateContractSpecificRules(req *catalogRequests.CreateCatalogItemRequest) error {
	// Contracts should have descriptions and terms
	if req.Description == "" {
		return fmt.Errorf("description is required for contracts")
	}

	// Contracts should have attributes with term and duration
	if req.Attributes == nil {
		return fmt.Errorf("attributes with term and duration are required for contracts")
	}

	// Check if term and duration are present in attributes
	if term, exists := req.Attributes["term"]; !exists || term == "" {
		return fmt.Errorf("term is required in attributes for contracts")
	}

	if duration, exists := req.Attributes["duration"]; !exists || duration == nil {
		return fmt.Errorf("duration is required in attributes for contracts")
	}

	return nil
}

// Helper method to format validation errors
func (cv *CatalogValidator) formatValidationError(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var messages []string
		for _, fieldError := range validationErrors {
			messages = append(messages, cv.formatFieldError(fieldError))
		}
		return fmt.Errorf("validation failed: %s", strings.Join(messages, ", "))
	}
	return err
}

func (cv *CatalogValidator) formatFieldError(err validator.FieldError) string {
	field := strings.ToLower(err.Field())
	tag := err.Tag()
	param := err.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, param)
	case "catalog_item_type":
		return fmt.Sprintf("%s must be PRODUCT, SERVICE, LABOUR, or CONTRACT", field)
	case "visibility_type":
		return fmt.Sprintf("%s must be PRIVATE, ORG, NETWORK, or PUBLIC", field)
	case "category_name":
		return fmt.Sprintf("%s contains invalid characters", field)
	case "product_sku":
		return fmt.Sprintf("%s must be a valid SKU format", field)
	case "price_positive":
		return fmt.Sprintf("%s must be greater than 0", field)
	case "unit_of_measure":
		return fmt.Sprintf("%s must be a valid unit of measure", field)
	case "image_urls":
		return fmt.Sprintf("%s must contain valid URLs", field)
	case "tags_list":
		return fmt.Sprintf("%s contains invalid tags", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uuid4":
		return fmt.Sprintf("%s must be a valid UUID", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// Custom validator functions

func validateCatalogItemType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return value == string(catalogModels.CatalogItemTypeProduct) ||
		value == string(catalogModels.CatalogItemTypeService) ||
		value == string(catalogModels.CatalogItemTypeLabour) ||
		value == string(catalogModels.CatalogItemTypeContract)
}

func validateVisibilityType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return value == string(catalogModels.VisibilityPrivate) ||
		value == string(catalogModels.VisibilityOrg) ||
		value == string(catalogModels.VisibilityNetwork) ||
		value == string(catalogModels.VisibilityPublic)
}

func validateCategoryName(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Optional field
	}
	// Allow alphanumeric, spaces, hyphens, underscores
	return utils.IsSafeString(value) && !utils.ContainsHTML(value)
}

func validatePricePositive(fl validator.FieldLevel) bool {
	if fl.Field().Type().String() == "decimal.Decimal" {
		// For decimal.Decimal, we need to handle it differently
		return true // Will be validated in business rules
	}
	return fl.Field().Float() > 0
}

func validateUnitOfMeasure(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Optional field
	}
	// Common units of measure validation
	return len(value) <= 20 && !utils.ContainsHTML(value)
}

func validateImageURLs(fl validator.FieldLevel) bool {
	if fl.Field().Kind() != 23 { // Slice
		return true
	}

	for i := 0; i < fl.Field().Len(); i++ {
		url := fl.Field().Index(i).String()
		if !utils.IsValidEmail(url) && !strings.HasPrefix(url, "http") {
			return false
		}
	}
	return true
}

func validateTagsList(fl validator.FieldLevel) bool {
	if fl.Field().Kind() != 23 { // Slice
		return true
	}

	if fl.Field().Len() > 20 {
		return false // Maximum 20 tags
	}

	for i := 0; i < fl.Field().Len(); i++ {
		tag := fl.Field().Index(i).String()
		if len(tag) > 50 || utils.ContainsHTML(tag) || utils.ContainsSQLKeywords(tag) {
			return false
		}
	}
	return true
}

// validateTypeSpecificAttributes validates attributes based on catalog item type
func (cv *CatalogValidator) validateTypeSpecificAttributes(itemType catalogModels.CatalogItemType, attributes map[string]interface{}) error {
	if attributes == nil {
		return nil
	}

	switch itemType {
	case catalogModels.CatalogItemTypeProduct:
		return cv.attributeValidator.ValidateProductAttributes(attributes)
	case catalogModels.CatalogItemTypeService:
		return cv.attributeValidator.ValidateServiceAttributes(attributes)
	case catalogModels.CatalogItemTypeLabour:
		return cv.attributeValidator.ValidateLabourAttributes(attributes)
	case catalogModels.CatalogItemTypeContract:
		return cv.attributeValidator.ValidateContractAttributes(attributes)
	default:
		return fmt.Errorf("unsupported catalog item type: %s", itemType)
	}
}

// ValidateAttributesForType validates attributes for a specific catalog type
// This method can be used by services when they have the item type context
func (cv *CatalogValidator) ValidateAttributesForType(itemType catalogModels.CatalogItemType, attributes map[string]interface{}) error {
	return cv.validateTypeSpecificAttributes(itemType, attributes)
}

// ValidateCrossFieldRules validates complex business rules that span multiple fields
func (cv *CatalogValidator) ValidateCrossFieldRules(req *catalogRequests.CreateCatalogItemRequest) error {
	// Contract-specific cross-field validation
	if req.ItemType == catalogModels.CatalogItemTypeContract {
		return cv.validateContractCrossFieldRules(req)
	}

	// Labour-specific cross-field validation
	if req.ItemType == catalogModels.CatalogItemTypeLabour {
		return cv.validateLabourCrossFieldRules(req)
	}

	// Service-specific cross-field validation
	if req.ItemType == catalogModels.CatalogItemTypeService {
		return cv.validateServiceCrossFieldRules(req)
	}

	// Product-specific cross-field validation
	if req.ItemType == catalogModels.CatalogItemTypeProduct {
		return cv.validateProductCrossFieldRules(req)
	}

	return nil
}

// validateContractCrossFieldRules validates contract-specific cross-field business rules
func (cv *CatalogValidator) validateContractCrossFieldRules(req *catalogRequests.CreateCatalogItemRequest) error {
	if req.Attributes == nil {
		return fmt.Errorf("attributes are required for contracts")
	}

	// Check if start_date and end_date are consistent
	if startDateRaw, hasStart := req.Attributes["start_date"]; hasStart {
		if endDateRaw, hasEnd := req.Attributes["end_date"]; hasEnd {
			startDateStr, ok1 := startDateRaw.(string)
			endDateStr, ok2 := endDateRaw.(string)

			if ok1 && ok2 && startDateStr != "" && endDateStr != "" {
				startDate, err1 := time.Parse("2006-01-02", startDateStr)
				endDate, err2 := time.Parse("2006-01-02", endDateStr)

				if err1 == nil && err2 == nil && endDate.Before(startDate) {
					return fmt.Errorf("contract end_date must be after start_date")
				}
			}
		}
	}

	// Validate payment terms consistency
	if paymentTermsRaw, exists := req.Attributes["payment_terms"]; exists {
		if paymentTermsMap, ok := paymentTermsRaw.(map[string]interface{}); ok {
			if percentage, hasPercentage := paymentTermsMap["percentage"]; hasPercentage {
				if percentageFloat, ok := percentage.(float64); ok && (percentageFloat < 0 || percentageFloat > 100) {
					return fmt.Errorf("payment terms percentage must be between 0 and 100")
				}
			}
		}
	}

	return nil
}

// validateLabourCrossFieldRules validates labour-specific cross-field business rules
func (cv *CatalogValidator) validateLabourCrossFieldRules(req *catalogRequests.CreateCatalogItemRequest) error {
	if req.Attributes == nil {
		return nil
	}

	// Validate rate type and unit rate consistency
	if rateTypeRaw, hasRateType := req.Attributes["rate_type"]; hasRateType {
		if unitRateRaw, hasUnitRate := req.Attributes["unit_rate"]; hasUnitRate {
			rateType, ok1 := rateTypeRaw.(string)
			if ok1 {
				if unitRateMap, ok2 := unitRateRaw.(map[string]interface{}); ok2 {
					if amount, hasAmount := unitRateMap["amount"]; hasAmount {
						if amountFloat, ok := amount.(float64); ok && amountFloat <= 0 {
							return fmt.Errorf("unit_rate amount must be greater than 0 for rate_type %s", rateType)
						}
					}
				}
			}
		}
	}

	// Validate skills and experience consistency
	if skillsRaw, hasSkills := req.Attributes["skills"]; hasSkills {
		if experienceRaw, hasExperience := req.Attributes["experience"]; hasExperience {
			if skills, ok1 := skillsRaw.([]interface{}); ok1 {
				if experience, ok2 := experienceRaw.(float64); ok2 {
					// If advanced skills are listed, experience should be reasonable
					advancedSkills := []string{"expert", "senior", "lead", "architect", "specialist"}
					hasAdvancedSkill := false

					for _, skillRaw := range skills {
						if skill, ok := skillRaw.(string); ok {
							for _, advancedSkill := range advancedSkills {
								if strings.Contains(strings.ToLower(skill), advancedSkill) {
									hasAdvancedSkill = true
									break
								}
							}
						}
					}

					if hasAdvancedSkill && experience < 3 {
						return fmt.Errorf("advanced skills typically require at least 3 years of experience")
					}
				}
			}
		}
	}

	return nil
}

// validateServiceCrossFieldRules validates service-specific cross-field business rules
func (cv *CatalogValidator) validateServiceCrossFieldRules(req *catalogRequests.CreateCatalogItemRequest) error {
	if req.Attributes == nil {
		return nil
	}

	// Validate SLA and pricing consistency
	if slaRaw, hasSLA := req.Attributes["sla"]; hasSLA {
		if slaMap, ok := slaRaw.(map[string]interface{}); ok {
			if availability, hasAvailability := slaMap["availability_percentage"]; hasAvailability {
				if availabilityFloat, ok := availability.(float64); ok && availabilityFloat > 99.5 {
					// High availability services should have premium pricing
					if req.BasePrice.LessThan(decimal.NewFromFloat(100)) {
						return fmt.Errorf("services with >99.5%% availability typically require premium pricing (>100 INR)")
					}
				}
			}
		}
	}

	// Validate duration and pricing consistency
	if durationRaw, hasDuration := req.Attributes["duration"]; hasDuration {
		if duration, ok := durationRaw.(string); ok {
			// Parse duration to minutes for validation
			if minutes := cv.parseDurationToMinutes(duration); minutes > 480 { // 8 hours
				// Long duration services should have appropriate pricing
				if req.BasePrice.LessThan(decimal.NewFromFloat(500)) {
					return fmt.Errorf("services longer than 8 hours typically require higher pricing (>500 INR)")
				}
			}
		}
	}

	return nil
}

// validateProductCrossFieldRules validates product-specific cross-field business rules
func (cv *CatalogValidator) validateProductCrossFieldRules(req *catalogRequests.CreateCatalogItemRequest) error {
	if req.Attributes == nil {
		return nil
	}

	// Validate perishable products have shelf life
	if perishableRaw, hasPerishable := req.Attributes["perishable"]; hasPerishable {
		if perishable, ok := perishableRaw.(bool); ok && perishable {
			if shelfLifeRaw, hasShelfLife := req.Attributes["shelf_life_days"]; !hasShelfLife {
				return fmt.Errorf("perishable products must specify shelf_life_days")
			} else if shelfLife, ok := shelfLifeRaw.(float64); ok && shelfLife <= 0 {
				return fmt.Errorf("shelf_life_days must be greater than 0 for perishable products")
			}
		}
	}

	// Validate weight and dimensions consistency
	if weightRaw, hasWeight := req.Attributes["weight"]; hasWeight {
		if dimensionsRaw, hasDimensions := req.Attributes["dimensions"]; hasDimensions {
			if weight, ok1 := weightRaw.(float64); ok1 {
				if dimensionsMap, ok2 := dimensionsRaw.(map[string]interface{}); ok2 {
					// Calculate approximate volume and check weight consistency
					if length, hasLength := dimensionsMap["length"]; hasLength {
						if width, hasWidth := dimensionsMap["width"]; hasWidth {
							if height, hasHeight := dimensionsMap["height"]; hasHeight {
								if l, ok1 := length.(float64); ok1 {
									if w, ok2 := width.(float64); ok2 {
										if h, ok3 := height.(float64); ok3 {
											volume := l * w * h // in cubic units
											// Basic density check (very loose validation)
											if weight > 0 && volume > 0 {
												density := weight / volume
												if density > 10000 { // Extremely dense
													return fmt.Errorf("weight and dimensions seem inconsistent (density too high)")
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return nil
}

// isValidSKUFormat validates SKU format using the same logic as the custom validator
func (cv *CatalogValidator) isValidSKUFormat(sku string) bool {
	if sku == "" {
		return true // Optional field
	}

	// SKU should be 3-50 characters, alphanumeric with hyphens and underscores
	if len(sku) < 3 || len(sku) > 50 {
		return false
	}

	// Must start with alphanumeric character
	if !((sku[0] >= 'A' && sku[0] <= 'Z') || (sku[0] >= 'a' && sku[0] <= 'z') || (sku[0] >= '0' && sku[0] <= '9')) {
		return false
	}

	// Check allowed characters
	for _, char := range sku {
		if !((char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') || char == '-' || char == '_') {
			return false
		}
	}

	return true
}

// parseDurationToMinutes parses duration string to minutes
func (cv *CatalogValidator) parseDurationToMinutes(duration string) int {
	// Simple parser for formats like "2h30m", "90m", "1h"
	var totalMinutes int

	// Extract hours
	if strings.Contains(duration, "h") {
		parts := strings.Split(duration, "h")
		if len(parts) > 0 {
			if hours, err := strconv.Atoi(parts[0]); err == nil {
				totalMinutes += hours * 60
			}
		}
		if len(parts) > 1 {
			duration = parts[1]
		}
	}

	// Extract minutes
	if strings.Contains(duration, "m") {
		parts := strings.Split(duration, "m")
		if len(parts) > 0 {
			if minutes, err := strconv.Atoi(parts[0]); err == nil {
				totalMinutes += minutes
			}
		}
	}

	return totalMinutes
}

// Enhanced custom validator functions for catalog-specific fields

// validateSKUFormat validates SKU format with enhanced rules
func validateSKUFormat(fl validator.FieldLevel) bool {
	sku := fl.Field().String()
	if sku == "" {
		return true // Optional field
	}

	// SKU should be 3-50 characters, alphanumeric with hyphens and underscores
	if len(sku) < 3 || len(sku) > 50 {
		return false
	}

	// Must start with alphanumeric character
	if !((sku[0] >= 'A' && sku[0] <= 'Z') || (sku[0] >= 'a' && sku[0] <= 'z') || (sku[0] >= '0' && sku[0] <= '9')) {
		return false
	}

	// Check allowed characters
	for _, char := range sku {
		if !((char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') || char == '-' || char == '_') {
			return false
		}
	}

	return true
}

// validateRateUnit validates rate unit for labour (hourly, daily, weekly, monthly)
func validateRateUnit(fl validator.FieldLevel) bool {
	rateUnit := strings.ToLower(fl.Field().String())
	validUnits := []string{"hourly", "daily", "weekly", "monthly", "per_hour", "per_day", "per_week", "per_month"}

	for _, validUnit := range validUnits {
		if rateUnit == validUnit {
			return true
		}
	}
	return false
}

// validateContractDate validates contract date format (YYYY-MM-DD)
func validateContractDate(fl validator.FieldLevel) bool {
	dateStr := fl.Field().String()
	if dateStr == "" {
		return true // Optional field
	}

	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}

// validateCurrencyISO validates 3-character ISO currency codes
func validateCurrencyISO(fl validator.FieldLevel) bool {
	currency := strings.ToUpper(fl.Field().String())
	if len(currency) != 3 {
		return false
	}

	// Common currency codes validation
	validCurrencies := []string{
		"INR", "USD", "EUR", "GBP", "JPY", "AUD", "CAD", "CHF", "CNY", "SEK", "NZD", "MXN", "SGD", "HKD", "NOK", "TRY", "ZAR", "BRL", "RUB", "KRW",
	}

	for _, validCurrency := range validCurrencies {
		if currency == validCurrency {
			return true
		}
	}

	// If not in common list, check if it's 3 uppercase letters
	for _, char := range currency {
		if char < 'A' || char > 'Z' {
			return false
		}
	}

	return true
}

// validateDecimalPositive validates that decimal value is positive
func validateDecimalPositive(fl validator.FieldLevel) bool {
	switch fl.Field().Type().String() {
	case "decimal.Decimal":
		// For decimal.Decimal fields, this will be validated in business logic
		return true
	case "float64":
		return fl.Field().Float() > 0
	case "int", "int64":
		return fl.Field().Int() > 0
	default:
		return true
	}
}

// validateDecimalNonNegative validates that decimal value is non-negative
func validateDecimalNonNegative(fl validator.FieldLevel) bool {
	switch fl.Field().Type().String() {
	case "decimal.Decimal":
		// For decimal.Decimal fields, this will be validated in business logic
		return true
	case "float64":
		return fl.Field().Float() >= 0
	case "int", "int64":
		return fl.Field().Int() >= 0
	default:
		return true
	}
}

// validatePercentage validates percentage values (0-100)
func validatePercentage(fl validator.FieldLevel) bool {
	value := fl.Field().Float()
	return value >= 0 && value <= 100
}

// validateDurationFormatValidator validates duration format (e.g., "2h30m", "90m")
func validateDurationFormatValidator(fl validator.FieldLevel) bool {
	duration := fl.Field().String()
	if duration == "" {
		return true // Optional field
	}

	// Check if it matches patterns like "2h", "30m", "2h30m"
	matched := false

	// Pattern 1: Just hours (e.g., "2h")
	if strings.HasSuffix(duration, "h") && !strings.Contains(duration[:len(duration)-1], "m") {
		if _, err := strconv.Atoi(duration[:len(duration)-1]); err == nil {
			matched = true
		}
	}

	// Pattern 2: Just minutes (e.g., "30m")
	if strings.HasSuffix(duration, "m") && !strings.Contains(duration, "h") {
		if _, err := strconv.Atoi(duration[:len(duration)-1]); err == nil {
			matched = true
		}
	}

	// Pattern 3: Hours and minutes (e.g., "2h30m")
	if strings.Contains(duration, "h") && strings.Contains(duration, "m") {
		parts := strings.Split(duration, "h")
		if len(parts) == 2 {
			if _, err := strconv.Atoi(parts[0]); err == nil {
				if strings.HasSuffix(parts[1], "m") {
					if _, err := strconv.Atoi(parts[1][:len(parts[1])-1]); err == nil {
						matched = true
					}
				}
			}
		}
	}

	return matched
}

// validateGeographicCoordinates validates latitude and longitude
func validateGeographicCoordinates(fl validator.FieldLevel) bool {
	// This validator expects a map with latitude and longitude
	if fl.Field().Kind() != 21 { // Map
		return true
	}

	// For now, just return true as this needs complex map validation
	// This should be handled in business logic with proper type assertion
	return true
}

// validateTimeSlotValid validates time slot format
func validateTimeSlotValid(fl validator.FieldLevel) bool {
	// This validator expects a time slot structure
	// For now, just return true as this needs complex struct validation
	// This should be handled in business logic with proper type assertion
	return true
}

// validateBusinessHours validates business hours format
func validateBusinessHours(fl validator.FieldLevel) bool {
	hours := fl.Field().String()
	if hours == "" {
		return true // Optional field
	}

	// Common business hours patterns
	validPatterns := []string{
		"24/7", "9-5", "9AM-5PM", "09:00-17:00", "24x7", "business_hours", "on_demand",
	}

	for _, pattern := range validPatterns {
		if strings.EqualFold(hours, pattern) {
			return true
		}
	}

	// Check if it matches HH:MM-HH:MM pattern
	if strings.Contains(hours, "-") {
		parts := strings.Split(hours, "-")
		if len(parts) == 2 {
			// Basic time format validation
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if len(part) >= 4 && strings.Contains(part, ":") {
					return true // Assume valid time format
				}
			}
		}
	}

	return false
}

// validateSkillName validates skill name format
func validateSkillName(fl validator.FieldLevel) bool {
	skill := strings.TrimSpace(fl.Field().String())
	if len(skill) == 0 || len(skill) > 100 {
		return false
	}

	// Should not contain special characters except spaces, hyphens, and parentheses
	for _, char := range skill {
		if !((char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') || char == ' ' || char == '-' ||
			char == '(' || char == ')' || char == '.' || char == '+') {
			return false
		}
	}

	return true
}

// validateEquipmentName validates equipment name format
func validateEquipmentName(fl validator.FieldLevel) bool {
	equipment := strings.TrimSpace(fl.Field().String())
	if len(equipment) == 0 || len(equipment) > 200 {
		return false
	}

	// Similar to skill name but allows more characters
	return !utils.ContainsHTML(equipment) && !utils.ContainsSQLKeywords(equipment)
}

// validateCertificationName validates certification name format
func validateCertificationName(fl validator.FieldLevel) bool {
	cert := strings.TrimSpace(fl.Field().String())
	if len(cert) == 0 || len(cert) > 300 {
		return false
	}

	return !utils.ContainsHTML(cert) && !utils.ContainsSQLKeywords(cert)
}

// validateLanguageCode validates language codes (ISO 639-1 or language names)
func validateLanguageCode(fl validator.FieldLevel) bool {
	lang := strings.ToLower(strings.TrimSpace(fl.Field().String()))
	if len(lang) == 0 || len(lang) > 50 {
		return false
	}

	// Common language codes and names
	validLanguages := []string{
		"en", "hi", "bn", "te", "mr", "ta", "gu", "kn", "ml", "or", "pa", "as", "ur",
		"english", "hindi", "bengali", "telugu", "marathi", "tamil", "gujarati",
		"kannada", "malayalam", "odia", "punjabi", "assamese", "urdu",
	}

	for _, validLang := range validLanguages {
		if lang == validLang {
			return true
		}
	}

	// If not in common list, check if it's a reasonable language name
	return len(lang) >= 2 && !utils.ContainsHTML(lang) && !utils.ContainsSQLKeywords(lang)
}

// validateContractTermType validates contract term types
func validateContractTermType(fl validator.FieldLevel) bool {
	term := strings.ToLower(fl.Field().String())
	validTerms := []string{
		"fixed", "renewable", "indefinite", "project-based", "milestone-based",
		"seasonal", "temporary", "permanent", "trial", "probationary",
	}

	for _, validTerm := range validTerms {
		if term == validTerm {
			return true
		}
	}
	return false
}

// validatePaymentMethod validates payment method types
func validatePaymentMethod(fl validator.FieldLevel) bool {
	method := strings.ToLower(fl.Field().String())
	validMethods := []string{
		"cash", "bank_transfer", "upi", "card", "cheque", "dd", "neft", "rtgs",
		"imps", "wallet", "credit", "debit", "net_banking", "mobile_payment",
	}

	for _, validMethod := range validMethods {
		if method == validMethod {
			return true
		}
	}
	return false
}

// validatePenaltyTypeValidator validates penalty types
func validatePenaltyTypeValidator(fl validator.FieldLevel) bool {
	penaltyType := strings.ToLower(fl.Field().String())
	validTypes := []string{
		"delay", "quality", "breach", "non_compliance", "late_delivery",
		"poor_quality", "contract_violation", "sla_breach", "performance",
	}

	for _, validType := range validTypes {
		if penaltyType == validType {
			return true
		}
	}
	return false
}

// validateMilestoneStatus validates milestone status values
func validateMilestoneStatus(fl validator.FieldLevel) bool {
	status := strings.ToLower(fl.Field().String())
	validStatuses := []string{
		"pending", "in_progress", "completed", "delayed", "cancelled",
		"on_hold", "approved", "rejected", "under_review",
	}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// validateDeliverableStatus validates deliverable status values
func validateDeliverableStatus(fl validator.FieldLevel) bool {
	status := strings.ToLower(fl.Field().String())
	validStatuses := []string{
		"pending", "in_progress", "completed", "delivered", "accepted",
		"rejected", "under_review", "cancelled", "delayed",
	}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// AdminApprovalWorkflow handles admin approval requirements for complex business rules
type AdminApprovalWorkflow struct {
	validator *validator.Validate
}

// NewAdminApprovalWorkflow creates a new admin approval workflow validator
func NewAdminApprovalWorkflow() *AdminApprovalWorkflow {
	return &AdminApprovalWorkflow{
		validator: validator.New(),
	}
}

// ApprovalRequirement represents an approval requirement
type ApprovalRequirement struct {
	Type         string `json:"type"`          // "pricing", "terms", "sla", "compliance"
	Reason       string `json:"reason"`        // Human-readable reason
	Severity     string `json:"severity"`      // "low", "medium", "high", "critical"
	AutoApprove  bool   `json:"auto_approve"`  // Whether this can be auto-approved
	ReviewerRole string `json:"reviewer_role"` // Required reviewer role
}

// ValidateForApproval checks if a catalog item requires admin approval
func (aaw *AdminApprovalWorkflow) ValidateForApproval(req *catalogRequests.CreateCatalogItemRequest) ([]ApprovalRequirement, error) {
	var requirements []ApprovalRequirement

	// Check pricing thresholds
	if pricingReqs := aaw.checkPricingApproval(req); len(pricingReqs) > 0 {
		requirements = append(requirements, pricingReqs...)
	}

	// Check contract terms approval
	if req.ItemType == catalogModels.CatalogItemTypeContract {
		if contractReqs := aaw.checkContractApproval(req); len(contractReqs) > 0 {
			requirements = append(requirements, contractReqs...)
		}
	}

	// Check SLA approval for services
	if req.ItemType == catalogModels.CatalogItemTypeService {
		if slaReqs := aaw.checkSLAApproval(req); len(slaReqs) > 0 {
			requirements = append(requirements, slaReqs...)
		}
	}

	// Check compliance requirements
	if complianceReqs := aaw.checkComplianceApproval(req); len(complianceReqs) > 0 {
		requirements = append(requirements, complianceReqs...)
	}

	return requirements, nil
}

// checkPricingApproval checks if pricing requires approval
func (aaw *AdminApprovalWorkflow) checkPricingApproval(req *catalogRequests.CreateCatalogItemRequest) []ApprovalRequirement {
	var requirements []ApprovalRequirement

	// High-value items require approval
	if req.BasePrice.GreaterThan(decimal.NewFromFloat(100000)) { // > 1 Lakh INR
		requirements = append(requirements, ApprovalRequirement{
			Type:         "pricing",
			Reason:       "High-value item exceeds automatic approval threshold (₹1,00,000)",
			Severity:     "high",
			AutoApprove:  false,
			ReviewerRole: "pricing_manager",
		})
	} else if req.BasePrice.GreaterThan(decimal.NewFromFloat(50000)) { // > 50K INR
		requirements = append(requirements, ApprovalRequirement{
			Type:         "pricing",
			Reason:       "Medium-value item requires pricing review (₹50,000+)",
			Severity:     "medium",
			AutoApprove:  true,
			ReviewerRole: "team_lead",
		})
	}

	// Check for unusual pricing patterns
	if req.ItemType == catalogModels.CatalogItemTypeLabour {
		if req.Attributes != nil {
			if unitRateRaw, exists := req.Attributes["unit_rate"]; exists {
				if unitRateMap, ok := unitRateRaw.(map[string]interface{}); ok {
					if amount, hasAmount := unitRateMap["amount"]; hasAmount {
						if amountFloat, ok := amount.(float64); ok {
							// Very high hourly rates require approval
							if rateTypeRaw, hasRateType := req.Attributes["rate_type"]; hasRateType {
								if rateType, ok := rateTypeRaw.(string); ok && rateType == "hourly" {
									if amountFloat > 5000 { // > ₹5000/hour
										requirements = append(requirements, ApprovalRequirement{
											Type:         "pricing",
											Reason:       "Exceptionally high hourly rate requires approval (₹5,000+/hour)",
											Severity:     "critical",
											AutoApprove:  false,
											ReviewerRole: "senior_manager",
										})
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return requirements
}

// checkContractApproval checks if contract terms require approval
func (aaw *AdminApprovalWorkflow) checkContractApproval(req *catalogRequests.CreateCatalogItemRequest) []ApprovalRequirement {
	var requirements []ApprovalRequirement

	if req.Attributes == nil {
		return requirements
	}

	// Long-term contracts require approval
	if durationRaw, exists := req.Attributes["duration"]; exists {
		if duration, ok := durationRaw.(float64); ok && duration > 24 { // > 2 years
			requirements = append(requirements, ApprovalRequirement{
				Type:         "terms",
				Reason:       "Long-term contract exceeds 24 months",
				Severity:     "high",
				AutoApprove:  false,
				ReviewerRole: "legal_team",
			})
		}
	}

	// Complex penalty structures require approval
	if penaltiesRaw, exists := req.Attributes["penalties"]; exists {
		if penalties, ok := penaltiesRaw.([]interface{}); ok && len(penalties) > 3 {
			requirements = append(requirements, ApprovalRequirement{
				Type:         "terms",
				Reason:       "Complex penalty structure with multiple penalty types",
				Severity:     "medium",
				AutoApprove:  true,
				ReviewerRole: "contract_manager",
			})
		}
	}

	// High-value contracts require legal review
	if req.BasePrice.GreaterThan(decimal.NewFromFloat(500000)) { // > 5 Lakh INR
		requirements = append(requirements, ApprovalRequirement{
			Type:         "terms",
			Reason:       "High-value contract requires legal review (₹5,00,000+)",
			Severity:     "critical",
			AutoApprove:  false,
			ReviewerRole: "legal_counsel",
		})
	}

	return requirements
}

// checkSLAApproval checks if SLA terms require approval
func (aaw *AdminApprovalWorkflow) checkSLAApproval(req *catalogRequests.CreateCatalogItemRequest) []ApprovalRequirement {
	var requirements []ApprovalRequirement

	if req.Attributes == nil {
		return requirements
	}

	if slaRaw, exists := req.Attributes["sla"]; exists {
		if slaMap, ok := slaRaw.(map[string]interface{}); ok {
			// Very high availability commitments require approval
			if availability, hasAvailability := slaMap["availability_percentage"]; hasAvailability {
				if availabilityFloat, ok := availability.(float64); ok && availabilityFloat > 99.9 {
					requirements = append(requirements, ApprovalRequirement{
						Type:         "sla",
						Reason:       "Extremely high availability commitment (>99.9%)",
						Severity:     "high",
						AutoApprove:  false,
						ReviewerRole: "operations_manager",
					})
				}
			}

			// Very fast response times require approval
			if responseTime, hasResponseTime := slaMap["response_time_minutes"]; hasResponseTime {
				if responseTimeFloat, ok := responseTime.(float64); ok && responseTimeFloat < 5 {
					requirements = append(requirements, ApprovalRequirement{
						Type:         "sla",
						Reason:       "Very fast response time commitment (<5 minutes)",
						Severity:     "medium",
						AutoApprove:  true,
						ReviewerRole: "service_manager",
					})
				}
			}
		}
	}

	return requirements
}

// checkComplianceApproval checks if compliance requirements need approval
func (aaw *AdminApprovalWorkflow) checkComplianceApproval(req *catalogRequests.CreateCatalogItemRequest) []ApprovalRequirement {
	var requirements []ApprovalRequirement

	// Check for regulated categories that require compliance approval
	regulatedCategories := []string{
		"pharmaceuticals", "medical", "healthcare", "financial", "insurance",
		"pesticides", "chemicals", "food_safety", "organic_certification",
	}

	category := strings.ToLower(req.Category)
	for _, regulated := range regulatedCategories {
		if strings.Contains(category, regulated) {
			requirements = append(requirements, ApprovalRequirement{
				Type:         "compliance",
				Reason:       fmt.Sprintf("Regulated category '%s' requires compliance review", regulated),
				Severity:     "critical",
				AutoApprove:  false,
				ReviewerRole: "compliance_officer",
			})
			break
		}
	}

	// Check for export/import items
	if req.Attributes != nil {
		if tags, exists := req.Attributes["tags"]; exists {
			if tagsList, ok := tags.([]interface{}); ok {
				for _, tagRaw := range tagsList {
					if tag, ok := tagRaw.(string); ok {
						tag = strings.ToLower(tag)
						if strings.Contains(tag, "export") || strings.Contains(tag, "import") ||
							strings.Contains(tag, "international") {
							requirements = append(requirements, ApprovalRequirement{
								Type:         "compliance",
								Reason:       "International trade item requires export/import compliance review",
								Severity:     "high",
								AutoApprove:  false,
								ReviewerRole: "trade_compliance",
							})
							break
						}
					}
				}
			}
		}
	}

	return requirements
}

// IsAutoApprovable checks if all requirements can be auto-approved
func (aaw *AdminApprovalWorkflow) IsAutoApprovable(requirements []ApprovalRequirement) bool {
	for _, req := range requirements {
		if !req.AutoApprove {
			return false
		}
	}
	return true
}

// GetRequiredReviewers returns the list of required reviewer roles
func (aaw *AdminApprovalWorkflow) GetRequiredReviewers(requirements []ApprovalRequirement) []string {
	reviewerMap := make(map[string]bool)

	for _, req := range requirements {
		if !req.AutoApprove {
			reviewerMap[req.ReviewerRole] = true
		}
	}

	var reviewers []string
	for reviewer := range reviewerMap {
		reviewers = append(reviewers, reviewer)
	}

	return reviewers
}
