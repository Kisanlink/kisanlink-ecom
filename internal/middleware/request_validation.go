package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/common"
	catalogRequests "github.com/Kisanlink/kisanlink-ecom/entities/requests/catalog"
	"github.com/Kisanlink/kisanlink-ecom/internal/validators"

	"github.com/gin-gonic/gin"
)

// RequestValidationMiddleware provides request-specific validation
type RequestValidationMiddleware struct {
	catalogValidator *validators.CatalogValidator
}

// NewRequestValidationMiddleware creates a new request validation middleware
func NewRequestValidationMiddleware() *RequestValidationMiddleware {
	return &RequestValidationMiddleware{
		catalogValidator: validators.NewCatalogValidator(),
	}
}

// ValidateCreateCatalogItem validates create catalog item requests
func (m *RequestValidationMiddleware) ValidateCreateCatalogItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req catalogRequests.CreateCatalogItemRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			m.handleValidationError(c, "INVALID_JSON", "Invalid JSON format", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// Validate the request
		if err := m.catalogValidator.ValidateCreateCatalogItemRequest(&req); err != nil {
			m.handleValidationError(c, "VALIDATION_FAILED", "Request validation failed", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// Store validated request in context
		c.Set("validated_request", &req)
		c.Next()
	}
}

// ValidateUpdateCatalogItem validates update catalog item requests
func (m *RequestValidationMiddleware) ValidateUpdateCatalogItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate ID parameter
		id := c.Param("id")
		if !isValidUUID(id) {
			m.handleValidationError(c, "INVALID_ID", "Invalid item ID format", map[string]interface{}{
				"id": id,
			})
			return
		}

		var req catalogRequests.UpdateCatalogItemRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			m.handleValidationError(c, "INVALID_JSON", "Invalid JSON format", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// Validate the request
		if err := m.catalogValidator.ValidateUpdateCatalogItemRequest(&req); err != nil {
			m.handleValidationError(c, "VALIDATION_FAILED", "Request validation failed", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// Store validated request in context
		c.Set("validated_request", &req)
		c.Set("item_id", id)
		c.Next()
	}
}

// ValidateListCatalogItems validates list catalog items requests
func (m *RequestValidationMiddleware) ValidateListCatalogItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		req := &catalogRequests.ListCatalogItemsRequest{}

		// Parse pagination parameters
		if pageStr := c.Query("page"); pageStr != "" {
			if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
				req.Page = page
			} else {
				m.handleValidationError(c, "INVALID_PAGE", "Invalid page parameter", map[string]interface{}{
					"page": pageStr,
				})
				return
			}
		}

		if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
			if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
				req.PageSize = pageSize
			} else {
				m.handleValidationError(c, "INVALID_PAGE_SIZE", "Invalid page_size parameter (1-100)", map[string]interface{}{
					"page_size": pageSizeStr,
				})
				return
			}
		}

		// Parse filter parameters
		if err := c.ShouldBindQuery(&req.CatalogFilter); err != nil {
			m.handleValidationError(c, "INVALID_QUERY_PARAMS", "Invalid query parameters", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// Parse sort parameters
		if sortBy := c.Query("sort_by"); sortBy != "" {
			if isValidSortField(sortBy) {
				req.SortBy = &sortBy
			} else {
				m.handleValidationError(c, "INVALID_SORT_BY", "Invalid sort_by parameter", map[string]interface{}{
					"sort_by": sortBy,
				})
				return
			}
		}

		if sortOrder := c.Query("sort_order"); sortOrder != "" {
			if sortOrder == "asc" || sortOrder == "desc" {
				req.SortOrder = &sortOrder
			} else {
				m.handleValidationError(c, "INVALID_SORT_ORDER", "Invalid sort_order parameter (asc/desc)", map[string]interface{}{
					"sort_order": sortOrder,
				})
				return
			}
		}

		// Validate the request
		if err := m.catalogValidator.ValidateListCatalogItemsRequest(req); err != nil {
			m.handleValidationError(c, "VALIDATION_FAILED", "Request validation failed", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// Store validated request in context
		c.Set("validated_request", req)
		c.Next()
	}
}

// ValidateSearchCatalogItems validates search catalog items requests
func (m *RequestValidationMiddleware) ValidateSearchCatalogItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		req := &catalogRequests.SearchCatalogRequest{}

		// Bind query parameters
		if err := c.ShouldBindQuery(req); err != nil {
			m.handleValidationError(c, "INVALID_QUERY_PARAMS", "Invalid query parameters", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// Validate the request
		if err := m.catalogValidator.ValidateSearchCatalogRequest(req); err != nil {
			m.handleValidationError(c, "VALIDATION_FAILED", "Request validation failed", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// Store validated request in context
		c.Set("validated_request", req)
		c.Next()
	}
}

// ValidateBulkUpdateCatalogItems validates bulk update requests
func (m *RequestValidationMiddleware) ValidateBulkUpdateCatalogItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req catalogRequests.BulkUpdateCatalogRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			m.handleValidationError(c, "INVALID_JSON", "Invalid JSON format", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// Validate the request
		if err := m.catalogValidator.ValidateBulkUpdateRequest(&req); err != nil {
			m.handleValidationError(c, "VALIDATION_FAILED", "Request validation failed", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// Store validated request in context
		c.Set("validated_request", &req)
		c.Next()
	}
}

// ValidateGetByID validates get by ID requests
func (m *RequestValidationMiddleware) ValidateGetByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if !isValidUUID(id) {
			m.handleValidationError(c, "INVALID_ID", "Invalid item ID format", map[string]interface{}{
				"id": id,
			})
			return
		}

		c.Set("item_id", id)
		c.Next()
	}
}

// ValidateDeleteByID validates delete by ID requests
func (m *RequestValidationMiddleware) ValidateDeleteByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if !isValidUUID(id) {
			m.handleValidationError(c, "INVALID_ID", "Invalid item ID format", map[string]interface{}{
				"id": id,
			})
			return
		}

		c.Set("item_id", id)
		c.Next()
	}
}

// ValidatePaginationParams validates common pagination parameters
func (m *RequestValidationMiddleware) ValidatePaginationParams() gin.HandlerFunc {
	return func(c *gin.Context) {
		pagination := struct {
			Page     int `form:"page" binding:"omitempty,min=1"`
			PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
		}{
			Page:     1,
			PageSize: 20,
		}

		if err := c.ShouldBindQuery(&pagination); err != nil {
			m.handleValidationError(c, "INVALID_PAGINATION", "Invalid pagination parameters", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		c.Set("page", pagination.Page)
		c.Set("page_size", pagination.PageSize)
		c.Next()
	}
}

// ValidateIDParam validates ID path parameter
func (m *RequestValidationMiddleware) ValidateIDParam(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param(paramName)
		if id == "" {
			m.handleValidationError(c, "MISSING_ID", "ID parameter is required", map[string]interface{}{
				"param": paramName,
			})
			return
		}

		if !isValidUUID(id) {
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

// ValidateQueryParam validates a specific query parameter
func (m *RequestValidationMiddleware) ValidateQueryParam(paramName string, required bool, validator func(string) bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		value := c.Query(paramName)

		if required && value == "" {
			m.handleValidationError(c, "MISSING_PARAM", "Required parameter missing", map[string]interface{}{
				"param": paramName,
			})
			return
		}

		if value != "" && validator != nil && !validator(value) {
			m.handleValidationError(c, "INVALID_PARAM", "Invalid parameter value", map[string]interface{}{
				"param": paramName,
				"value": value,
			})
			return
		}

		if value != "" {
			c.Set(paramName, value)
		}
		c.Next()
	}
}

// ValidateHeaders validates required headers
func (m *RequestValidationMiddleware) ValidateHeaders(requiredHeaders []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, header := range requiredHeaders {
			value := c.GetHeader(header)
			if value == "" {
				m.handleValidationError(c, "MISSING_HEADER", "Required header missing", map[string]interface{}{
					"header": header,
				})
				return
			}
		}
		c.Next()
	}
}

// ValidateContentType validates request content type
func (m *RequestValidationMiddleware) ValidateContentType(allowedTypes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		contentType := c.GetHeader("Content-Type")
		if contentType == "" {
			m.handleValidationError(c, "MISSING_CONTENT_TYPE", "Content-Type header is required", nil)
			return
		}

		// Extract base content type (ignore charset, boundary, etc.)
		baseType := strings.Split(contentType, ";")[0]
		baseType = strings.TrimSpace(baseType)

		for _, allowed := range allowedTypes {
			if strings.EqualFold(baseType, allowed) {
				c.Next()
				return
			}
		}

		m.handleValidationError(c, "UNSUPPORTED_CONTENT_TYPE", "Unsupported content type", map[string]interface{}{
			"content_type":  baseType,
			"allowed_types": allowedTypes,
		})
	}
}

// Helper methods

func (m *RequestValidationMiddleware) handleValidationError(c *gin.Context, code, message string, details interface{}) {
	c.AbortWithStatusJSON(http.StatusBadRequest, common.APIResponse{
		Success: false,
		Error: &common.APIError{
			Code:    code,
			Message: message,
			Details: fmt.Sprintf("%v", details),
		},
	})
}

// Utility functions

func isValidSortField(field string) bool {
	validFields := []string{
		"created_at", "updated_at", "name", "base_price", "category", "subcategory",
	}
	for _, valid := range validFields {
		if field == valid {
			return true
		}
	}
	return false
}

// Convenience functions for common validations

// ValidatePositiveInteger validates that a string is a positive integer
func ValidatePositiveInteger(value string) bool {
	if value == "" {
		return false
	}
	if num, err := strconv.Atoi(value); err != nil || num <= 0 {
		return false
	}
	return true
}

// ValidateNonNegativeInteger validates that a string is a non-negative integer
func ValidateNonNegativeInteger(value string) bool {
	if value == "" {
		return false
	}
	if num, err := strconv.Atoi(value); err != nil || num < 0 {
		return false
	}
	return true
}

// ValidateEnum validates that a value is in a list of allowed values
func ValidateEnum(allowedValues []string) func(string) bool {
	return func(value string) bool {
		for _, allowed := range allowedValues {
			if value == allowed {
				return true
			}
		}
		return false
	}
}

// ValidateStringLength validates string length constraints
func ValidateStringLength(min, max int) func(string) bool {
	return func(value string) bool {
		length := len(value)
		return length >= min && length <= max
	}
}

// ValidateAlphanumeric validates that a string contains only alphanumeric characters
func ValidateAlphanumeric(value string) bool {
	for _, char := range value {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return false
		}
	}
	return true
}

// GetValidatedRequest safely extracts validated request from context
func GetValidatedRequest[T any](c *gin.Context) (*T, bool) {
	value, exists := c.Get("validated_request")
	if !exists {
		return nil, false
	}

	if typedValue, ok := value.(*T); ok {
		return typedValue, true
	}

	return nil, false
}

// MustGetValidatedRequest extracts validated request or panics
func MustGetValidatedRequest[T any](c *gin.Context) *T {
	req, ok := GetValidatedRequest[T](c)
	if !ok {
		panic("Validated request not found or wrong type - ensure validation middleware is applied")
	}
	return req
}
