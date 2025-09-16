package middleware

import (
	"fmt"
	"html"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"kisanlink-ecom/entities/models/common"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/microcosm-cc/bluemonday"
)

// ValidationConfig holds configuration for validation middleware
type ValidationConfig struct {
	// EnableSanitization enables HTML/XSS sanitization
	EnableSanitization bool
	// MaxRequestSize sets maximum request body size (in bytes)
	MaxRequestSize int64
	// EnableStrictMode enforces stricter validation rules
	EnableStrictMode bool
	// AllowedContentTypes restricts content types
	AllowedContentTypes []string
	// CustomValidators allows custom validation functions
	CustomValidators map[string]validator.Func
}

// DefaultValidationConfig returns default validation configuration
func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		EnableSanitization:  true,
		MaxRequestSize:      10 * 1024 * 1024, // 10MB
		EnableStrictMode:    true,
		AllowedContentTypes: []string{"application/json", "multipart/form-data", "application/x-www-form-urlencoded"},
		CustomValidators:    make(map[string]validator.Func),
	}
}

// ValidationMiddleware provides comprehensive input validation and sanitization
type ValidationMiddleware struct {
	validator  *validator.Validate
	sanitizer  *bluemonday.Policy
	config     *ValidationConfig
	sqlPattern *regexp.Regexp
	xssPattern *regexp.Regexp
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware(config *ValidationConfig) *ValidationMiddleware {
	if config == nil {
		config = DefaultValidationConfig()
	}

	// Initialize validator
	validate := validator.New()

	// Register custom validators
	for tag, fn := range config.CustomValidators {
		validate.RegisterValidation(tag, fn)
	}

	// Register common custom validators
	validate.RegisterValidation("nohtml", validateNoHTML)
	validate.RegisterValidation("nosql", validateNoSQL)
	validate.RegisterValidation("slug", validateSlug)
	validate.RegisterValidation("currency", validateCurrency)
	validate.RegisterValidation("phone", validatePhone)
	validate.RegisterValidation("alphanum_space", validateAlphanumSpace)
	validate.RegisterValidation("safe_string", validateSafeString)

	// Initialize sanitizer
	sanitizer := bluemonday.UGCPolicy()
	if config.EnableStrictMode {
		sanitizer = bluemonday.StrictPolicy()
	}

	// Compile security patterns
	sqlPattern := regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute|script|javascript|vbscript|onload|onerror|onclick)`)
	xssPattern := regexp.MustCompile(`(?i)(<script|javascript:|vbscript:|onload|onerror|onclick|<iframe|<object|<embed)`)

	return &ValidationMiddleware{
		validator:  validate,
		sanitizer:  sanitizer,
		config:     config,
		sqlPattern: sqlPattern,
		xssPattern: xssPattern,
	}
}

// ValidateRequest validates and sanitizes incoming requests
func (m *ValidationMiddleware) ValidateRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check content type
		if !m.isAllowedContentType(c.GetHeader("Content-Type")) {
			m.handleValidationError(c, "UNSUPPORTED_CONTENT_TYPE", "Content type not supported", nil)
			return
		}

		// Check request size
		if c.Request.ContentLength > m.config.MaxRequestSize {
			m.handleValidationError(c, "REQUEST_TOO_LARGE", "Request body too large", map[string]interface{}{
				"max_size":    m.config.MaxRequestSize,
				"actual_size": c.Request.ContentLength,
			})
			return
		}

		// Validate and sanitize path parameters
		if err := m.validatePathParams(c); err != nil {
			m.handleValidationError(c, "INVALID_PATH_PARAMS", "Invalid path parameters", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// Validate and sanitize query parameters
		if err := m.validateQueryParams(c); err != nil {
			m.handleValidationError(c, "INVALID_QUERY_PARAMS", "Invalid query parameters", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// Validate and sanitize headers
		if err := m.validateHeaders(c); err != nil {
			m.handleValidationError(c, "INVALID_HEADERS", "Invalid headers", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		c.Next()
	}
}

// ValidateJSON validates JSON request body
func (m *ValidationMiddleware) ValidateJSON(v interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a new instance of the struct type
		reqType := reflect.TypeOf(v)
		if reqType.Kind() == reflect.Ptr {
			reqType = reqType.Elem()
		}
		req := reflect.New(reqType).Interface()

		// Read and validate JSON
		if err := c.ShouldBindJSON(req); err != nil {
			m.handleValidationError(c, "INVALID_JSON", "Invalid JSON format", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// Validate the struct first (before sanitization)
		if err := m.validator.Struct(req); err != nil {
			m.handleValidationError(c, "VALIDATION_FAILED", "Request validation failed", m.formatValidationErrors(err))
			return
		}

		// Sanitize the request if enabled (after validation)
		if m.config.EnableSanitization {
			m.sanitizeStruct(req)
		}

		// Store validated request in context
		c.Set("validated_request", req)
		c.Next()
	}
}

// ValidateStruct validates any struct and stores it in context
func (m *ValidationMiddleware) ValidateStruct(v interface{}, contextKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate the struct
		if err := m.validator.Struct(v); err != nil {
			m.handleValidationError(c, "VALIDATION_FAILED", "Struct validation failed", m.formatValidationErrors(err))
			return
		}

		// Store in context
		c.Set(contextKey, v)
		c.Next()
	}
}

// SanitizeInput sanitizes input strings
func (m *ValidationMiddleware) SanitizeInput() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize query parameters
		query := c.Request.URL.Query()
		modified := false
		for key, values := range query {
			for i, value := range values {
				sanitized := m.sanitizeString(value)
				if sanitized != value {
					query[key][i] = sanitized
					modified = true
				}
			}
		}

		// Update the URL if query parameters were modified
		if modified {
			c.Request.URL.RawQuery = query.Encode()
		}

		// For POST/PUT requests with form data
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			if err := c.Request.ParseForm(); err == nil {
				for key, values := range c.Request.Form {
					for i, value := range values {
						c.Request.Form[key][i] = m.sanitizeString(value)
					}
				}
			}
		}

		c.Next()
	}
}

// RateLimitValidation validates rate limiting parameters
func (m *ValidationMiddleware) RateLimitValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for suspicious patterns in headers that might indicate bot/attack
		userAgent := c.GetHeader("User-Agent")
		if userAgent == "" || m.isSuspiciousUserAgent(userAgent) {
			c.Header("X-Rate-Limit-Reason", "suspicious_user_agent")
		}

		// Check for rapid requests pattern
		clientIP := c.ClientIP()
		if m.isRapidRequests(clientIP) {
			c.Header("X-Rate-Limit-Reason", "rapid_requests")
		}

		c.Next()
	}
}

// Helper methods

func (m *ValidationMiddleware) isAllowedContentType(contentType string) bool {
	if contentType == "" {
		return true // Allow empty content type for GET requests
	}

	// Extract base content type (ignore charset, boundary, etc.)
	baseType := strings.Split(contentType, ";")[0]
	baseType = strings.TrimSpace(baseType)

	for _, allowed := range m.config.AllowedContentTypes {
		if strings.EqualFold(baseType, allowed) {
			return true
		}
	}
	return false
}

func (m *ValidationMiddleware) validatePathParams(c *gin.Context) error {
	for _, param := range c.Params {
		// Check for basic security threats
		if m.containsSecurityThreats(param.Value) {
			return fmt.Errorf("security threat detected in path parameter '%s'", param.Key)
		}

		// Validate specific path parameter patterns
		if err := m.validatePathParam(param.Key, param.Value); err != nil {
			return fmt.Errorf("invalid path parameter '%s': %w", param.Key, err)
		}
	}
	return nil
}

func (m *ValidationMiddleware) validateQueryParams(c *gin.Context) error {
	for key, values := range c.Request.URL.Query() {
		for _, value := range values {
			if m.containsSecurityThreats(value) {
				return fmt.Errorf("security threat detected in query parameter '%s'", key)
			}

			// Validate specific query parameters
			if err := m.validateQueryParam(key, value); err != nil {
				return fmt.Errorf("invalid query parameter '%s': %w", key, err)
			}
		}
	}
	return nil
}

func (m *ValidationMiddleware) validateHeaders(c *gin.Context) error {
	// Validate critical headers
	if userAgent := c.GetHeader("User-Agent"); userAgent != "" {
		if len(userAgent) > 500 || m.containsSecurityThreats(userAgent) {
			return fmt.Errorf("invalid User-Agent header")
		}
	}

	if referer := c.GetHeader("Referer"); referer != "" {
		if len(referer) > 2000 || m.containsSecurityThreats(referer) {
			return fmt.Errorf("invalid Referer header")
		}
	}

	return nil
}

func (m *ValidationMiddleware) validatePathParam(key, value string) error {
	switch key {
	case "id":
		// Validate UUID format
		if !isValidUUID(value) {
			return fmt.Errorf("must be a valid UUID")
		}
	case "page":
		if page, err := strconv.Atoi(value); err != nil || page < 1 {
			return fmt.Errorf("must be a positive integer")
		}
	case "limit", "size":
		if limit, err := strconv.Atoi(value); err != nil || limit < 1 || limit > 1000 {
			return fmt.Errorf("must be between 1 and 1000")
		}
	}
	return nil
}

func (m *ValidationMiddleware) validateQueryParam(key, value string) error {
	switch key {
	case "page":
		if page, err := strconv.Atoi(value); err != nil || page < 1 {
			return fmt.Errorf("must be a positive integer")
		}
	case "limit", "size", "per_page":
		if limit, err := strconv.Atoi(value); err != nil || limit < 1 || limit > 1000 {
			return fmt.Errorf("must be between 1 and 1000")
		}
	case "sort_order":
		if value != "asc" && value != "desc" {
			return fmt.Errorf("must be 'asc' or 'desc'")
		}
	}
	return nil
}

func (m *ValidationMiddleware) containsSecurityThreats(input string) bool {
	// Check for SQL injection patterns
	if m.sqlPattern.MatchString(input) {
		return true
	}

	// Check for XSS patterns
	if m.xssPattern.MatchString(input) {
		return true
	}

	// Check for path traversal
	if strings.Contains(input, "../") || strings.Contains(input, "..\\") {
		return true
	}

	// Check for null bytes
	if strings.Contains(input, "\x00") {
		return true
	}

	return false
}

func (m *ValidationMiddleware) sanitizeString(input string) string {
	if !m.config.EnableSanitization {
		return input
	}

	// HTML sanitization
	sanitized := m.sanitizer.Sanitize(input)

	// Additional sanitization
	sanitized = html.EscapeString(sanitized)

	// Remove potentially dangerous characters
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")

	return sanitized
}

func (m *ValidationMiddleware) sanitizeStruct(v interface{}) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if !field.CanSet() {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(m.sanitizeString(field.String()))
		case reflect.Ptr:
			if !field.IsNil() && field.Elem().Kind() == reflect.String {
				sanitized := m.sanitizeString(field.Elem().String())
				field.Elem().SetString(sanitized)
			}
		case reflect.Slice:
			if field.Type().Elem().Kind() == reflect.String {
				for j := 0; j < field.Len(); j++ {
					elem := field.Index(j)
					elem.SetString(m.sanitizeString(elem.String()))
				}
			}
		case reflect.Struct:
			m.sanitizeStruct(field.Addr().Interface())
		}
	}
}

func (m *ValidationMiddleware) isSuspiciousUserAgent(userAgent string) bool {
	suspicious := []string{
		"curl", "wget", "python", "bot", "crawler", "spider", "scraper",
		"automated", "test", "scan", "exploit", "injection",
	}

	userAgentLower := strings.ToLower(userAgent)
	for _, pattern := range suspicious {
		if strings.Contains(userAgentLower, pattern) {
			return true
		}
	}
	return false
}

func (m *ValidationMiddleware) isRapidRequests(clientIP string) bool {
	// This would typically integrate with a rate limiter
	// For now, return false - implement with actual rate limiting logic
	return false
}

func (m *ValidationMiddleware) formatValidationErrors(err error) map[string]interface{} {
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

func (m *ValidationMiddleware) formatFieldError(err validator.FieldError) string {
	field := strings.ToLower(err.Field())
	tag := err.Tag()
	param := err.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", field, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uuid4":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "nohtml":
		return fmt.Sprintf("%s cannot contain HTML", field)
	case "nosql":
		return fmt.Sprintf("%s contains potentially dangerous SQL patterns", field)
	case "safe_string":
		return fmt.Sprintf("%s contains unsafe characters", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

func (m *ValidationMiddleware) handleValidationError(c *gin.Context, code, message string, details interface{}) {
	c.AbortWithStatusJSON(http.StatusBadRequest, common.APIResponse{
		Success: false,
		Error: &common.APIError{
			Code:    code,
			Message: message,
			Details: fmt.Sprintf("%v", details),
		},
	})
}

// Custom validator functions

func validateNoHTML(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return !strings.Contains(value, "<") && !strings.Contains(value, ">")
}

func validateNoSQL(fl validator.FieldLevel) bool {
	value := strings.ToLower(fl.Field().String())
	sqlKeywords := []string{"select", "insert", "update", "delete", "drop", "union", "exec", "execute"}
	for _, keyword := range sqlKeywords {
		if strings.Contains(value, keyword) {
			return false
		}
	}
	return true
}

func validateSlug(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	slugPattern := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	return slugPattern.MatchString(value)
}

func validateCurrency(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return len(value) == 3 && regexp.MustCompile(`^[A-Z]{3}$`).MatchString(value)
}

func validatePhone(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	phonePattern := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	return phonePattern.MatchString(value)
}

func validateAlphanumSpace(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	alphanumSpacePattern := regexp.MustCompile(`^[a-zA-Z0-9\s]+$`)
	return alphanumSpacePattern.MatchString(value)
}

func validateSafeString(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	// Allow alphanumeric, spaces, hyphens, underscores, and basic punctuation
	safePattern := regexp.MustCompile(`^[a-zA-Z0-9\s\-_.,!?()]+$`)
	return safePattern.MatchString(value)
}

// Utility functions

func isValidUUID(uuid string) bool {
	uuidPattern := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	return uuidPattern.MatchString(uuid)
}
