package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents a standardized API response
// @Description Standard API response structure used across all endpoints
type Response struct {
	Data  interface{}    `json:"data,omitempty"`                       // Response data payload
	Meta  *ResponseMeta  `json:"meta,omitempty" swaggertype:"object"`  // Response metadata including trace ID and pagination
	Error *ResponseError `json:"error,omitempty" swaggertype:"object"` // Error details if request failed
}

// ResponseMeta contains metadata about the response
// @Description Metadata included in API responses for tracing and pagination
type ResponseMeta struct {
	TraceID    string                 `json:"trace_id,omitempty" example:"abc123xyz"`             // Request trace ID for debugging
	Pagination *PaginationMeta        `json:"pagination,omitempty" swaggertype:"object"`          // Pagination information for list endpoints
	Timestamp  string                 `json:"timestamp,omitempty" example:"2025-01-05T10:30:00Z"` // Response timestamp
	Extra      map[string]interface{} `json:"extra,omitempty" swaggertype:"object"`               // Additional metadata
}

// ResponseError represents an error response
// @Description Error details when a request fails
type ResponseError struct {
	Code    string                 `json:"code" example:"INVALID_INPUT"`           // Error code for programmatic handling
	Message string                 `json:"message" example:"Invalid input data"`   // Human-readable error message
	Details map[string]interface{} `json:"details,omitempty" swaggertype:"object"` // Additional error details
}

// PaginationMeta contains pagination information
// @Description Pagination metadata for list endpoints
type PaginationMeta struct {
	Page       int  `json:"page" example:"1"`         // Current page number
	Limit      int  `json:"limit" example:"20"`       // Items per page
	Total      int  `json:"total" example:"100"`      // Total number of items
	TotalPages int  `json:"total_pages" example:"5"`  // Total number of pages
	HasNext    bool `json:"has_next" example:"true"`  // Whether there is a next page
	HasPrev    bool `json:"has_prev" example:"false"` // Whether there is a previous page
}

// Success sends a successful response
func Success(c *gin.Context, data interface{}, meta *ResponseMeta) {
	response := Response{
		Data: data,
		Meta: meta,
	}

	if meta != nil && meta.TraceID == "" {
		meta.TraceID = GetTraceID(c)
	}

	c.JSON(http.StatusOK, response)
}

// Created sends a 201 Created response
func Created(c *gin.Context, data interface{}, meta *ResponseMeta) {
	response := Response{
		Data: data,
		Meta: meta,
	}

	if meta != nil && meta.TraceID == "" {
		meta.TraceID = GetTraceID(c)
	}

	c.JSON(http.StatusCreated, response)
}

// Error sends an error response
func Error(c *gin.Context, statusCode int, code, message string, details map[string]interface{}) {
	response := Response{
		Error: &ResponseError{
			Code:    code,
			Message: message,
			Details: details,
		},
		Meta: &ResponseMeta{
			TraceID: GetTraceID(c),
		},
	}

	c.JSON(statusCode, response)
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, code, message string, details map[string]interface{}) {
	Error(c, http.StatusBadRequest, code, message, details)
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, code, message string, details map[string]interface{}) {
	Error(c, http.StatusUnauthorized, code, message, details)
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *gin.Context, code, message string, details map[string]interface{}) {
	Error(c, http.StatusForbidden, code, message, details)
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, code, message string, details map[string]interface{}) {
	Error(c, http.StatusNotFound, code, message, details)
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(c *gin.Context, code, message string, details map[string]interface{}) {
	Error(c, http.StatusInternalServerError, code, message, details)
}

// GetTraceID extracts the trace ID from the gin context
func GetTraceID(c *gin.Context) string {
	if traceID, exists := c.Get("trace_id"); exists {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}

// GetOrganizationID extracts the organization ID from the gin context
// It tries multiple key variants to ensure compatibility across all handlers
func GetOrganizationID(c *gin.Context) (string, bool) {
	// Try camelCase variant (used by some handlers)
	if orgID, exists := c.Get("organizationID"); exists {
		if id, ok := orgID.(string); ok && id != "" {
			return id, true
		}
	}

	// Try snake_case variant (used by other handlers)
	if orgID, exists := c.Get("organization_id"); exists {
		if id, ok := orgID.(string); ok && id != "" {
			return id, true
		}
	}

	// Try standard key (set by middleware)
	if orgID, exists := c.Get("orgID"); exists {
		if id, ok := orgID.(string); ok && id != "" {
			return id, true
		}
	}

	return "", false
}

// GetSubjectID extracts the subject ID (user ID) from the gin context
// It tries multiple key variants to ensure compatibility
func GetSubjectID(c *gin.Context) (string, bool) {
	// Try standard subjectID key
	if subjectID, exists := c.Get("subjectID"); exists {
		if id, ok := subjectID.(string); ok && id != "" {
			return id, true
		}
	}

	// Try alternative user_id key
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok && id != "" {
			return id, true
		}
	}

	return "", false
}

// GetUserRoles retrieves the user roles from the gin context
func GetUserRoles(c *gin.Context) ([]string, bool) {
	// Try snake_case first (primary key set by middleware)
	if roles, exists := c.Get("user_roles"); exists {
		if roleList, ok := roles.([]string); ok {
			return roleList, true
		}
	}

	// Try camelCase as fallback for backward compatibility
	if roles, exists := c.Get("userRoles"); exists {
		if roleList, ok := roles.([]string); ok {
			return roleList, true
		}
	}

	return nil, false
}

// IsAdmin checks if the user has admin, super_admin, or ecom_admin role
func IsAdmin(c *gin.Context) bool {
	roles, exists := GetUserRoles(c)
	if !exists {
		return false
	}

	for _, role := range roles {
		if role == "admin" || role == "super_admin" || role == "ecom_admin" {
			return true
		}
	}
	return false
}

// NewPaginationMeta creates pagination metadata
func NewPaginationMeta(page, limit, total int) *PaginationMeta {
	totalPages := (total + limit - 1) / limit // Ceiling division

	return &PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}
