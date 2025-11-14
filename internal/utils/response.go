package utils

import (
	"net/http"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/common"
	commonErrors "github.com/Kisanlink/kisanlink-ecom/internal/common"

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

// ErrorHandler provides centralized error handling for utils
var errorHandler = commonErrors.NewErrorHandler(nil)

// HandleError processes errors using the centralized error handler
func HandleError(c *gin.Context, err error) {
	errorHandler.HandleError(c, err)
}

// HandleValidationErrors processes validation errors
func HandleValidationErrors(c *gin.Context, validationErr *commonErrors.ValidationErrors) {
	errorHandler.HandleValidationError(c, validationErr)
}

// AppErrorResponse sends an AppError response
func AppErrorResponse(c *gin.Context, appErr *commonErrors.AppError) {
	response := common.APIResponse{
		Success: false,
		Error: &common.APIError{
			Code:    string(appErr.Code),
			Message: appErr.Message,
			Details: appErr.Details,
			Fields:  appErr.Fields,
			Context: appErr.Context,
		},
	}
	c.JSON(appErr.StatusCode, response)
}

// BusinessRuleErrorResponse sends a business rule violation error
func BusinessRuleErrorResponse(c *gin.Context, rule, details string) {
	appErr := commonErrors.NewBusinessRuleError(rule, details)
	AppErrorResponse(c, appErr)
}

// DatabaseErrorResponse sends a database error response
func DatabaseErrorResponse(c *gin.Context, operation string, cause error) {
	appErr := commonErrors.NewDatabaseError(operation, cause)
	AppErrorResponse(c, appErr)
}

// ExternalServiceErrorResponse sends an external service error response
func ExternalServiceErrorResponse(c *gin.Context, service string, cause error) {
	appErr := commonErrors.NewExternalServiceError(service, cause)
	AppErrorResponse(c, appErr)
}

// ValidationErrorResponseWithFields sends a validation error with field details
func ValidationErrorResponseWithFields(c *gin.Context, fields map[string]string) {
	validationErr := commonErrors.NewValidationErrors()
	for field, message := range fields {
		validationErr.Add(field, message)
	}
	HandleValidationErrors(c, validationErr)
}
