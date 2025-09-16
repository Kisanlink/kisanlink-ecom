package validators

import (
	"fmt"
	"strings"

	catalogModels "kisanlink-ecom/entities/models/catalog"
	catalogRequests "kisanlink-ecom/entities/requests/catalog"
	"kisanlink-ecom/internal/utils"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

// CatalogValidator provides validation for catalog-related operations
type CatalogValidator struct {
	validator *validator.Validate
	sanitizer *utils.Sanitizer
}

// NewCatalogValidator creates a new catalog validator
func NewCatalogValidator() *CatalogValidator {
	validate := validator.New()
	sanitizer := utils.NewSanitizer(utils.DefaultSanitizerConfig())

	// Register custom catalog validators
	validate.RegisterValidation("catalog_item_type", validateCatalogItemType)
	validate.RegisterValidation("visibility_type", validateVisibilityType)
	validate.RegisterValidation("category_name", validateCategoryName)
	validate.RegisterValidation("product_sku", validateProductSKU)
	validate.RegisterValidation("price_positive", validatePricePositive)
	validate.RegisterValidation("unit_of_measure", validateUnitOfMeasure)
	validate.RegisterValidation("image_urls", validateImageURLs)
	validate.RegisterValidation("tags_list", validateTagsList)

	return &CatalogValidator{
		validator: validate,
		sanitizer: sanitizer,
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
	return cv.validateCreateCatalogItemBusinessRules(req)
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
	return cv.validateUpdateCatalogItemBusinessRules(req)
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
		return fmt.Sprintf("%s must be PRODUCT, SERVICE, or LABOUR", field)
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
		value == string(catalogModels.CatalogItemTypeLabour)
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

func validateProductSKU(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Optional field
	}
	// SKU should be alphanumeric with hyphens and underscores
	return len(value) <= 50 && !utils.ContainsHTML(value) && !utils.ContainsSQLKeywords(value)
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
