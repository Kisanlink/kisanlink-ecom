package utils

import (
	"kisanlink-ecom/internal/models/common"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SuccessResponse sends a successful API response.
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	response := common.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	c.JSON(statusCode, response)
}

// ErrorResponse sends an error API response.
func ErrorResponse(c *gin.Context, statusCode int, code, message, details string) {
	response := common.APIResponse{
		Success: false,
		Message: message,
		Error: &common.APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
	c.JSON(statusCode, response)
}

// ValidationErrorResponse sends a validation error response.
func ValidationErrorResponse(c *gin.Context, details string) {
	ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid input data", details)
}

// NotFoundResponse sends a not found error response.
func NotFoundResponse(c *gin.Context, resource string) {
	ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", resource+" not found", "")
}

// InternalErrorResponse sends an internal server error response.
func InternalErrorResponse(c *gin.Context, details string) {
	ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", details)
}

// UnauthorizedResponse sends an unauthorized error response.
func UnauthorizedResponse(c *gin.Context) {
	ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", "")
}

// ForbiddenResponse sends a forbidden error response.
func ForbiddenResponse(c *gin.Context) {
	ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Access denied", "")
}

// PaginatedResponse sends a paginated response.
func PaginatedResponse(c *gin.Context, statusCode int, message string, data interface{}, pagination *common.Pagination) {
	response := common.PaginatedResponse{
		APIResponse: common.APIResponse{
			Success: true,
			Message: message,
			Data:    data,
		},
		Pagination: pagination,
	}
	c.JSON(statusCode, response)
}
