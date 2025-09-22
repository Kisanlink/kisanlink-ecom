package middleware

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"kisanlink-ecom/entities/models/common"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// MarketplaceValidationMiddleware provides marketplace-specific request validation
type MarketplaceValidationMiddleware struct {
	validator *validator.Validate
}

// NewMarketplaceValidationMiddleware creates a new marketplace validation middleware
func NewMarketplaceValidationMiddleware() *MarketplaceValidationMiddleware {
	validate := validator.New()

	// Register marketplace-specific custom validators
	validate.RegisterValidation("listing_status", validateListingStatus)
	validate.RegisterValidation("bid_status", validateBidStatus)
	validate.RegisterValidation("auction_type", validateAuctionType)
	validate.RegisterValidation("listing_visibility", validateListingVisibility)
	validate.RegisterValidation("bid_visibility", validateBidVisibility)
	validate.RegisterValidation("currency_code", validateCurrencyCode)
	validate.RegisterValidation("positive_decimal", validatePositiveDecimal)
	validate.RegisterValidation("duration_hours", validateDurationHours)
	validate.RegisterValidation("marketplace_id", validateMarketplaceID)
	validate.RegisterValidation("no_script_injection", validateNoScriptInjection)
	validate.RegisterValidation("safe_text", validateSafeText)

	return &MarketplaceValidationMiddleware{
		validator: validate,
	}
}

// Marketplace request structures for validation

// CreateListingRequest represents the request to create a marketplace listing
type CreateListingRequest struct {
	ProductID            string  `json:"product_id" validate:"required,marketplace_id"`
	Quantity             float64 `json:"quantity" validate:"required,positive_decimal"`
	AskingPrice          float64 `json:"asking_price" validate:"required,positive_decimal"`
	MinimumBid           float64 `json:"minimum_bid" validate:"required,positive_decimal"`
	Currency             string  `json:"currency" validate:"required,currency_code"`
	ListingDurationHours int     `json:"listing_duration_hours" validate:"required,duration_hours"`
	Visibility           string  `json:"visibility" validate:"required,listing_visibility"`
	AuctionType          string  `json:"auction_type" validate:"required,auction_type"`
	BidVisibility        string  `json:"bid_visibility" validate:"required,bid_visibility"`
	PickupLocation       *struct {
		Address   string  `json:"address" validate:"omitempty,max=500,safe_text"`
		Latitude  float64 `json:"latitude" validate:"omitempty,gte=-90,lte=90"`
		Longitude float64 `json:"longitude" validate:"omitempty,gte=-180,lte=180"`
	} `json:"pickup_location,omitempty"`
	TermsConditions string `json:"terms_conditions" validate:"omitempty,max=2000,safe_text"`
}

// UpdateListingRequest represents the request to update a marketplace listing
type UpdateListingRequest struct {
	AskingPrice          *float64 `json:"asking_price,omitempty" validate:"omitempty,positive_decimal"`
	MinimumBid           *float64 `json:"minimum_bid,omitempty" validate:"omitempty,positive_decimal"`
	ListingDurationHours *int     `json:"listing_duration_hours,omitempty" validate:"omitempty,duration_hours"`
	Visibility           *string  `json:"visibility,omitempty" validate:"omitempty,listing_visibility"`
	BidVisibility        *string  `json:"bid_visibility,omitempty" validate:"omitempty,bid_visibility"`
	TermsConditions      *string  `json:"terms_conditions,omitempty" validate:"omitempty,max=2000,safe_text"`
}

// PlaceBidRequest represents the request to place a bid
type PlaceBidRequest struct {
	BidAmount     float64  `json:"bid_amount" validate:"required,positive_decimal"`
	Currency      string   `json:"currency" validate:"required,currency_code"`
	Quantity      float64  `json:"quantity" validate:"required,positive_decimal"`
	Message       string   `json:"message" validate:"omitempty,max=500,safe_text"`
	AutoBidLimit  *float64 `json:"auto_bid_limit,omitempty" validate:"omitempty,positive_decimal"`
	PaymentMethod string   `json:"payment_method" validate:"omitempty,max=50,safe_text"`
}

// ListingFilterRequest represents filtering parameters for listing queries
type ListingFilterRequest struct {
	Status         []string `form:"status" validate:"omitempty,dive,listing_status"`
	Visibility     []string `form:"visibility" validate:"omitempty,dive,listing_visibility"`
	AuctionType    []string `form:"auction_type" validate:"omitempty,dive,auction_type"`
	MinPrice       *float64 `form:"min_price" validate:"omitempty,positive_decimal"`
	MaxPrice       *float64 `form:"max_price" validate:"omitempty,positive_decimal"`
	Currency       string   `form:"currency" validate:"omitempty,currency_code"`
	SellerID       string   `form:"seller_id" validate:"omitempty,marketplace_id"`
	ProductID      string   `form:"product_id" validate:"omitempty,marketplace_id"`
	ExpiringWithin *int     `form:"expiring_within" validate:"omitempty,min=1,max=168"` // hours
}

// BidFilterRequest represents filtering parameters for bid queries
type BidFilterRequest struct {
	Status    []string `form:"status" validate:"omitempty,dive,bid_status"`
	MinAmount *float64 `form:"min_amount" validate:"omitempty,positive_decimal"`
	MaxAmount *float64 `form:"max_amount" validate:"omitempty,positive_decimal"`
	Currency  string   `form:"currency" validate:"omitempty,currency_code"`
	BidderID  string   `form:"bidder_id" validate:"omitempty,marketplace_id"`
	ListingID string   `form:"listing_id" validate:"omitempty,marketplace_id"`
}

// AdminActionRequest represents admin action requests
type AdminActionRequest struct {
	Action string `json:"action" validate:"required,oneof=force_close remove_bid suspend_user"`
	Reason string `json:"reason" validate:"required,min=10,max=500,safe_text"`
}

// Validation middleware functions

// ValidateCreateListing validates create listing requests
func (m *MarketplaceValidationMiddleware) ValidateCreateListing() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateListingRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			m.handleValidationError(c, "INVALID_JSON", "Invalid JSON format", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// Validate the request structure
		if err := m.validator.Struct(&req); err != nil {
			m.handleValidationError(c, "VALIDATION_FAILED", "Request validation failed", m.formatValidationErrors(err))
			return
		}

		// Business rule validations
		if err := m.validateCreateListingBusinessRules(&req); err != nil {
			m.handleValidationError(c, "BUSINESS_RULE_VIOLATION", "Business rule validation failed", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		c.Set("validated_request", &req)
		c.Next()
	}
}

// ValidateUpdateListing validates update listing requests
func (m *MarketplaceValidationMiddleware) ValidateUpdateListing() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate listing ID parameter
		listingID := c.Param("id")
		if !isValidMarketplaceID(listingID) {
			m.handleValidationError(c, "INVALID_LISTING_ID", "Invalid listing ID format", map[string]interface{}{
				"listing_id": listingID,
			})
			return
		}

		var req UpdateListingRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			m.handleValidationError(c, "INVALID_JSON", "Invalid JSON format", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// Validate the request structure
		if err := m.validator.Struct(&req); err != nil {
			m.handleValidationError(c, "VALIDATION_FAILED", "Request validation failed", m.formatValidationErrors(err))
			return
		}

		// Business rule validations
		if err := m.validateUpdateListingBusinessRules(&req); err != nil {
			m.handleValidationError(c, "BUSINESS_RULE_VIOLATION", "Business rule validation failed", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		c.Set("validated_request", &req)
		c.Set("listing_id", listingID)
		c.Next()
	}
}

// ValidatePlaceBid validates place bid requests
func (m *MarketplaceValidationMiddleware) ValidatePlaceBid() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate listing ID parameter
		listingID := c.Param("id")
		if !isValidMarketplaceID(listingID) {
			m.handleValidationError(c, "INVALID_LISTING_ID", "Invalid listing ID format", map[string]interface{}{
				"listing_id": listingID,
			})
			return
		}

		var req PlaceBidRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			m.handleValidationError(c, "INVALID_JSON", "Invalid JSON format", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// Validate the request structure
		if err := m.validator.Struct(&req); err != nil {
			m.handleValidationError(c, "VALIDATION_FAILED", "Request validation failed", m.formatValidationErrors(err))
			return
		}

		// Business rule validations
		if err := m.validatePlaceBidBusinessRules(&req); err != nil {
			m.handleValidationError(c, "BUSINESS_RULE_VIOLATION", "Business rule validation failed", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		c.Set("validated_request", &req)
		c.Set("listing_id", listingID)
		c.Next()
	}
}

// ValidateListingFilters validates listing filter parameters
func (m *MarketplaceValidationMiddleware) ValidateListingFilters() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ListingFilterRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			m.handleValidationError(c, "INVALID_QUERY_PARAMS", "Invalid query parameters", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// Validate the request structure
		if err := m.validator.Struct(&req); err != nil {
			m.handleValidationError(c, "VALIDATION_FAILED", "Filter validation failed", m.formatValidationErrors(err))
			return
		}

		// Business rule validations for filters
		if err := m.validateListingFilterBusinessRules(&req); err != nil {
			m.handleValidationError(c, "BUSINESS_RULE_VIOLATION", "Filter business rule validation failed", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		c.Set("validated_filters", &req)
		c.Next()
	}
}

// ValidateBidFilters validates bid filter parameters
func (m *MarketplaceValidationMiddleware) ValidateBidFilters() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req BidFilterRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			m.handleValidationError(c, "INVALID_QUERY_PARAMS", "Invalid query parameters", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// Validate the request structure
		if err := m.validator.Struct(&req); err != nil {
			m.handleValidationError(c, "VALIDATION_FAILED", "Filter validation failed", m.formatValidationErrors(err))
			return
		}

		c.Set("validated_filters", &req)
		c.Next()
	}
}

// ValidateAdminAction validates admin action requests
func (m *MarketplaceValidationMiddleware) ValidateAdminAction() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AdminActionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			m.handleValidationError(c, "INVALID_JSON", "Invalid JSON format", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// Validate the request structure
		if err := m.validator.Struct(&req); err != nil {
			m.handleValidationError(c, "VALIDATION_FAILED", "Request validation failed", m.formatValidationErrors(err))
			return
		}

		c.Set("validated_request", &req)
		c.Next()
	}
}

// ValidateMarketplaceID validates marketplace ID parameters
func (m *MarketplaceValidationMiddleware) ValidateMarketplaceID(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param(paramName)
		if !isValidMarketplaceID(id) {
			m.handleValidationError(c, "INVALID_ID", "Invalid ID format", map[string]interface{}{
				"param": paramName,
				"value": id,
			})
			return
		}

		c.Set(paramName, id)
		c.Next()
	}
}

// ValidatePaginationAndSort validates pagination and sorting parameters
func (m *MarketplaceValidationMiddleware) ValidatePaginationAndSort() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate pagination
		page := 1
		pageSize := 20

		if pageStr := c.Query("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err != nil || p < 1 {
				m.handleValidationError(c, "INVALID_PAGE", "Invalid page parameter", map[string]interface{}{
					"page": pageStr,
				})
				return
			} else {
				page = p
			}
		}

		if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
			if ps, err := strconv.Atoi(pageSizeStr); err != nil || ps < 1 || ps > 100 {
				m.handleValidationError(c, "INVALID_PAGE_SIZE", "Invalid page_size parameter (1-100)", map[string]interface{}{
					"page_size": pageSizeStr,
				})
				return
			} else {
				pageSize = ps
			}
		}

		// Validate sorting
		sortBy := c.Query("sort_by")
		if sortBy != "" && !isValidMarketplaceSortField(sortBy) {
			m.handleValidationError(c, "INVALID_SORT_BY", "Invalid sort_by parameter", map[string]interface{}{
				"sort_by": sortBy,
				"allowed": getValidMarketplaceSortFields(),
			})
			return
		}

		sortOrder := c.Query("sort_order")
		if sortOrder != "" && sortOrder != "asc" && sortOrder != "desc" {
			m.handleValidationError(c, "INVALID_SORT_ORDER", "Invalid sort_order parameter (asc/desc)", map[string]interface{}{
				"sort_order": sortOrder,
			})
			return
		}

		c.Set("page", page)
		c.Set("page_size", pageSize)
		if sortBy != "" {
			c.Set("sort_by", sortBy)
		}
		if sortOrder != "" {
			c.Set("sort_order", sortOrder)
		}

		c.Next()
	}
}

// Business rule validation functions

func (m *MarketplaceValidationMiddleware) validateCreateListingBusinessRules(req *CreateListingRequest) error {
	// Minimum bid cannot be higher than asking price
	if req.MinimumBid > req.AskingPrice {
		return fmt.Errorf("minimum bid cannot be higher than asking price")
	}

	// Listing duration must be reasonable (1 hour to 30 days)
	if req.ListingDurationHours < 1 || req.ListingDurationHours > 720 {
		return fmt.Errorf("listing duration must be between 1 and 720 hours")
	}

	// Quantity must be positive
	if req.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	// Validate pickup location if provided
	if req.PickupLocation != nil {
		if req.PickupLocation.Address == "" {
			return fmt.Errorf("pickup location address is required when location is provided")
		}
	}

	return nil
}

func (m *MarketplaceValidationMiddleware) validateUpdateListingBusinessRules(req *UpdateListingRequest) error {
	// If both asking price and minimum bid are provided, validate their relationship
	if req.AskingPrice != nil && req.MinimumBid != nil {
		if *req.MinimumBid > *req.AskingPrice {
			return fmt.Errorf("minimum bid cannot be higher than asking price")
		}
	}

	// Validate duration if provided
	if req.ListingDurationHours != nil {
		if *req.ListingDurationHours < 1 || *req.ListingDurationHours > 720 {
			return fmt.Errorf("listing duration must be between 1 and 720 hours")
		}
	}

	return nil
}

func (m *MarketplaceValidationMiddleware) validatePlaceBidBusinessRules(req *PlaceBidRequest) error {
	// Bid amount must be positive
	if req.BidAmount <= 0 {
		return fmt.Errorf("bid amount must be positive")
	}

	// Quantity must be positive
	if req.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	// Auto bid limit must be higher than bid amount if provided
	if req.AutoBidLimit != nil && *req.AutoBidLimit <= req.BidAmount {
		return fmt.Errorf("auto bid limit must be higher than bid amount")
	}

	return nil
}

func (m *MarketplaceValidationMiddleware) validateListingFilterBusinessRules(req *ListingFilterRequest) error {
	// Price range validation
	if req.MinPrice != nil && req.MaxPrice != nil {
		if *req.MinPrice > *req.MaxPrice {
			return fmt.Errorf("min_price cannot be greater than max_price")
		}
	}

	return nil
}

// Custom validator functions

func validateListingStatus(fl validator.FieldLevel) bool {
	status := fl.Field().String()
	validStatuses := []string{"ACTIVE", "CLOSED", "EXPIRED", "CANCELLED"}
	return contains(validStatuses, status)
}

func validateBidStatus(fl validator.FieldLevel) bool {
	status := fl.Field().String()
	validStatuses := []string{"ACTIVE", "OUTBID", "WINNING", "EXPIRED", "REMOVED"}
	return contains(validStatuses, status)
}

func validateAuctionType(fl validator.FieldLevel) bool {
	auctionType := fl.Field().String()
	validTypes := []string{"OPEN", "CLOSED"}
	return contains(validTypes, auctionType)
}

func validateListingVisibility(fl validator.FieldLevel) bool {
	visibility := fl.Field().String()
	validVisibilities := []string{"PRIVATE", "PUBLIC", "NETWORK", "ORGANIZATION"}
	return contains(validVisibilities, visibility)
}

func validateBidVisibility(fl validator.FieldLevel) bool {
	visibility := fl.Field().String()
	validVisibilities := []string{"FULL", "PARTIAL", "MINIMAL", "HIDDEN"}
	return contains(validVisibilities, visibility)
}

func validateCurrencyCode(fl validator.FieldLevel) bool {
	currency := fl.Field().String()
	// ISO 4217 currency codes are 3 uppercase letters
	currencyPattern := regexp.MustCompile(`^[A-Z]{3}$`)
	return currencyPattern.MatchString(currency)
}

func validatePositiveDecimal(fl validator.FieldLevel) bool {
	value := fl.Field().Float()
	return value > 0
}

func validateDurationHours(fl validator.FieldLevel) bool {
	hours := fl.Field().Int()
	return hours >= 1 && hours <= 720 // 1 hour to 30 days
}

func validateMarketplaceID(fl validator.FieldLevel) bool {
	id := fl.Field().String()
	return isValidMarketplaceID(id)
}

func validateNoScriptInjection(fl validator.FieldLevel) bool {
	value := strings.ToLower(fl.Field().String())
	dangerousPatterns := []string{
		"<script", "javascript:", "vbscript:", "onload=", "onerror=", "onclick=",
		"<iframe", "<object", "<embed", "eval(", "alert(", "document.cookie",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(value, pattern) {
			return false
		}
	}
	return true
}

func validateSafeText(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	// Check for script injection
	if !validateNoScriptInjection(fl) {
		return false
	}

	// Check for SQL injection patterns
	sqlPatterns := []string{
		"union select", "drop table", "delete from", "insert into",
		"update set", "create table", "alter table", "exec ", "execute ",
	}

	valueLower := strings.ToLower(value)
	for _, pattern := range sqlPatterns {
		if strings.Contains(valueLower, pattern) {
			return false
		}
	}

	return true
}

// Utility functions

func isValidMarketplaceID(id string) bool {
	if id == "" {
		return false
	}

	// Accept UUID format or custom marketplace ID format
	uuidPattern := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	customIDPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]{8,50}$`)

	return uuidPattern.MatchString(id) || customIDPattern.MatchString(id)
}

func isValidMarketplaceSortField(field string) bool {
	validFields := getValidMarketplaceSortFields()
	return contains(validFields, field)
}

func getValidMarketplaceSortFields() []string {
	return []string{
		"created_at", "updated_at", "expires_at", "asking_price", "minimum_bid",
		"bid_count", "status", "listing_duration_hours", "quantity",
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Error handling and formatting

func (m *MarketplaceValidationMiddleware) handleValidationError(c *gin.Context, code, message string, details interface{}) {
	c.AbortWithStatusJSON(http.StatusBadRequest, common.APIResponse{
		Success: false,
		Error: &common.APIError{
			Code:    code,
			Message: message,
			Details: fmt.Sprintf("%v", details),
		},
	})
}

func (m *MarketplaceValidationMiddleware) formatValidationErrors(err error) map[string]interface{} {
	errors := make(map[string]interface{})

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		fieldErrors := make(map[string]string)
		for _, fieldError := range validationErrors {
			fieldErrors[strings.ToLower(fieldError.Field())] = m.formatFieldError(fieldError)
		}
		errors["field_errors"] = fieldErrors
	} else {
		errors["general_error"] = err.Error()
	}

	return errors
}

func (m *MarketplaceValidationMiddleware) formatFieldError(err validator.FieldError) string {
	field := strings.ToLower(err.Field())
	tag := err.Tag()
	param := err.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, param)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	case "listing_status":
		return fmt.Sprintf("%s must be a valid listing status (ACTIVE, CLOSED, EXPIRED, CANCELLED)", field)
	case "bid_status":
		return fmt.Sprintf("%s must be a valid bid status (ACTIVE, OUTBID, WINNING, EXPIRED, REMOVED)", field)
	case "auction_type":
		return fmt.Sprintf("%s must be a valid auction type (OPEN, CLOSED)", field)
	case "listing_visibility":
		return fmt.Sprintf("%s must be a valid visibility (PRIVATE, PUBLIC, NETWORK, ORGANIZATION)", field)
	case "bid_visibility":
		return fmt.Sprintf("%s must be a valid bid visibility (FULL, PARTIAL, MINIMAL, HIDDEN)", field)
	case "currency_code":
		return fmt.Sprintf("%s must be a valid 3-letter currency code", field)
	case "positive_decimal":
		return fmt.Sprintf("%s must be a positive number", field)
	case "duration_hours":
		return fmt.Sprintf("%s must be between 1 and 720 hours", field)
	case "marketplace_id":
		return fmt.Sprintf("%s must be a valid marketplace ID", field)
	case "no_script_injection":
		return fmt.Sprintf("%s contains potentially dangerous script content", field)
	case "safe_text":
		return fmt.Sprintf("%s contains unsafe content", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// Convenience functions for extracting validated data

// GetValidatedCreateListingRequest extracts validated create listing request
func GetValidatedCreateListingRequest(c *gin.Context) (*CreateListingRequest, bool) {
	req, exists := c.Get("validated_request")
	if !exists {
		return nil, false
	}
	if typedReq, ok := req.(*CreateListingRequest); ok {
		return typedReq, true
	}
	return nil, false
}

// GetValidatedUpdateListingRequest extracts validated update listing request
func GetValidatedUpdateListingRequest(c *gin.Context) (*UpdateListingRequest, bool) {
	req, exists := c.Get("validated_request")
	if !exists {
		return nil, false
	}
	if typedReq, ok := req.(*UpdateListingRequest); ok {
		return typedReq, true
	}
	return nil, false
}

// GetValidatedPlaceBidRequest extracts validated place bid request
func GetValidatedPlaceBidRequest(c *gin.Context) (*PlaceBidRequest, bool) {
	req, exists := c.Get("validated_request")
	if !exists {
		return nil, false
	}
	if typedReq, ok := req.(*PlaceBidRequest); ok {
		return typedReq, true
	}
	return nil, false
}

// GetValidatedListingFilters extracts validated listing filters
func GetValidatedListingFilters(c *gin.Context) (*ListingFilterRequest, bool) {
	filters, exists := c.Get("validated_filters")
	if !exists {
		return nil, false
	}
	if typedFilters, ok := filters.(*ListingFilterRequest); ok {
		return typedFilters, true
	}
	return nil, false
}

// GetValidatedBidFilters extracts validated bid filters
func GetValidatedBidFilters(c *gin.Context) (*BidFilterRequest, bool) {
	filters, exists := c.Get("validated_filters")
	if !exists {
		return nil, false
	}
	if typedFilters, ok := filters.(*BidFilterRequest); ok {
		return typedFilters, true
	}
	return nil, false
}

// GetValidatedAdminAction extracts validated admin action request
func GetValidatedAdminAction(c *gin.Context) (*AdminActionRequest, bool) {
	req, exists := c.Get("validated_request")
	if !exists {
		return nil, false
	}
	if typedReq, ok := req.(*AdminActionRequest); ok {
		return typedReq, true
	}
	return nil, false
}
