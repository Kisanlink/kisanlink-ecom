package common

import (
	"fmt"
	"net/http"
)

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	// General errors
	ErrorCodeInternal         ErrorCode = "INTERNAL_ERROR"
	ErrorCodeInvalidRequest   ErrorCode = "INVALID_REQUEST"
	ErrorCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrorCodeNotFound         ErrorCode = "NOT_FOUND"
	ErrorCodeConflict         ErrorCode = "CONFLICT"
	ErrorCodeUnauthorized     ErrorCode = "UNAUTHORIZED"
	ErrorCodeForbidden        ErrorCode = "FORBIDDEN"
	ErrorCodeRateLimited      ErrorCode = "RATE_LIMITED"

	// Business logic errors
	ErrorCodeInsufficientInventory   ErrorCode = "INSUFFICIENT_INVENTORY"
	ErrorCodeInvalidStatusTransition ErrorCode = "INVALID_STATUS_TRANSITION"
	ErrorCodeOrderNotCancellable     ErrorCode = "ORDER_NOT_CANCELLABLE"
	ErrorCodeCatalogItemInactive     ErrorCode = "CATALOG_ITEM_INACTIVE"
	ErrorCodeOrganizationMismatch    ErrorCode = "ORGANIZATION_MISMATCH"
	ErrorCodeDuplicateResource       ErrorCode = "DUPLICATE_RESOURCE"

	// Database errors
	ErrorCodeDatabaseConnection ErrorCode = "DATABASE_CONNECTION_ERROR"
	ErrorCodeDatabaseTimeout    ErrorCode = "DATABASE_TIMEOUT"
	ErrorCodeDatabaseConstraint ErrorCode = "DATABASE_CONSTRAINT_VIOLATION"

	// External service errors
	ErrorCodeExternalServiceUnavailable ErrorCode = "EXTERNAL_SERVICE_UNAVAILABLE"
	ErrorCodeExternalServiceTimeout     ErrorCode = "EXTERNAL_SERVICE_TIMEOUT"
	ErrorCodeAuthenticationServiceError ErrorCode = "AUTHENTICATION_SERVICE_ERROR"

	// Inventory specific errors
	ErrorCodeInventoryLotExpired  ErrorCode = "INVENTORY_LOT_EXPIRED"
	ErrorCodeInventoryLotNotFound ErrorCode = "INVENTORY_LOT_NOT_FOUND"
	ErrorCodeInvalidQuantity      ErrorCode = "INVALID_QUANTITY"
	ErrorCodeReservationFailed    ErrorCode = "RESERVATION_FAILED"

	// Order specific errors
	ErrorCodeOrderAlreadyProcessed ErrorCode = "ORDER_ALREADY_PROCESSED"
	ErrorCodeOrderItemsEmpty       ErrorCode = "ORDER_ITEMS_EMPTY"
	ErrorCodeInvalidOrderStatus    ErrorCode = "INVALID_ORDER_STATUS"
	ErrorCodePaymentRequired       ErrorCode = "PAYMENT_REQUIRED"
)

// AppError represents a structured application error
type AppError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Cause      error                  `json:"-"` // Internal cause, not exposed in JSON
	HTTPStatus int                    `json:"-"` // HTTP status code for this error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause for error unwrapping
func (e *AppError) Unwrap() error {
	return e.Cause
}

// ToAPIError converts AppError to APIError for API responses
func (e *AppError) ToAPIError() *APIError {
	details := ""
	if len(e.Details) > 0 {
		// Convert details map to string representation
		details = fmt.Sprintf("%v", e.Details)
	}

	return &APIError{
		Code:    string(e.Code),
		Message: e.Message,
		Details: details,
	}
}

// NewAppError creates a new application error
func NewAppError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: getDefaultHTTPStatus(code),
	}
}

// NewAppErrorWithCause creates a new application error with an underlying cause
func NewAppErrorWithCause(code ErrorCode, message string, cause error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Cause:      cause,
		HTTPStatus: getDefaultHTTPStatus(code),
	}
}

// NewAppErrorWithDetails creates a new application error with additional details
func NewAppErrorWithDetails(code ErrorCode, message string, details map[string]interface{}) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		HTTPStatus: getDefaultHTTPStatus(code),
	}
}

// WithDetails adds details to an existing error
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	e.Details = details
	return e
}

// WithHTTPStatus sets a custom HTTP status code
func (e *AppError) WithHTTPStatus(status int) *AppError {
	e.HTTPStatus = status
	return e
}

// getDefaultHTTPStatus returns the default HTTP status code for an error code
func getDefaultHTTPStatus(code ErrorCode) int {
	switch code {
	case ErrorCodeInvalidRequest, ErrorCodeValidationFailed, ErrorCodeInvalidQuantity,
		ErrorCodeOrderItemsEmpty, ErrorCodeInvalidOrderStatus:
		return http.StatusBadRequest
	case ErrorCodeUnauthorized, ErrorCodeAuthenticationServiceError:
		return http.StatusUnauthorized
	case ErrorCodeForbidden, ErrorCodeOrganizationMismatch:
		return http.StatusForbidden
	case ErrorCodeNotFound, ErrorCodeInventoryLotNotFound:
		return http.StatusNotFound
	case ErrorCodeConflict, ErrorCodeDuplicateResource, ErrorCodeInvalidStatusTransition,
		ErrorCodeOrderNotCancellable, ErrorCodeOrderAlreadyProcessed:
		return http.StatusConflict
	case ErrorCodeInsufficientInventory, ErrorCodeCatalogItemInactive, ErrorCodeInventoryLotExpired,
		ErrorCodeReservationFailed, ErrorCodePaymentRequired:
		return http.StatusUnprocessableEntity
	case ErrorCodeRateLimited:
		return http.StatusTooManyRequests
	case ErrorCodeExternalServiceUnavailable, ErrorCodeDatabaseConnection, ErrorCodeExternalServiceTimeout,
		ErrorCodeDatabaseTimeout:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// Predefined common errors
var (
	ErrInternal         = NewAppError(ErrorCodeInternal, "An internal error occurred")
	ErrInvalidRequest   = NewAppError(ErrorCodeInvalidRequest, "Invalid request")
	ErrNotFound         = NewAppError(ErrorCodeNotFound, "Resource not found")
	ErrUnauthorized     = NewAppError(ErrorCodeUnauthorized, "Authentication required")
	ErrForbidden        = NewAppError(ErrorCodeForbidden, "Access denied")
	ErrConflict         = NewAppError(ErrorCodeConflict, "Resource conflict")
	ErrValidationFailed = NewAppError(ErrorCodeValidationFailed, "Validation failed")
)

// Validation error helpers
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// NewValidationError creates a validation error with field-specific details
func NewValidationError(field, message string, value interface{}) *AppError {
	return NewAppErrorWithDetails(
		ErrorCodeValidationFailed,
		"Validation failed",
		map[string]interface{}{
			"validation_errors": []ValidationError{
				{
					Field:   field,
					Message: message,
					Value:   value,
				},
			},
		},
	)
}

// NewMultiValidationError creates a validation error with multiple field errors
func NewMultiValidationError(errors []ValidationError) *AppError {
	return NewAppErrorWithDetails(
		ErrorCodeValidationFailed,
		"Multiple validation errors occurred",
		map[string]interface{}{
			"validation_errors": errors,
		},
	)
}

// Business logic error helpers

// NewInsufficientInventoryError creates an inventory shortage error
func NewInsufficientInventoryError(itemID string, requested, available string) *AppError {
	return NewAppErrorWithDetails(
		ErrorCodeInsufficientInventory,
		"Insufficient inventory available",
		map[string]interface{}{
			"catalog_item_id": itemID,
			"requested":       requested,
			"available":       available,
		},
	)
}

// NewOrganizationMismatchError creates an organization access error
func NewOrganizationMismatchError(resourceType, resourceID, userOrg, resourceOrg string) *AppError {
	return NewAppErrorWithDetails(
		ErrorCodeOrganizationMismatch,
		fmt.Sprintf("%s does not belong to your organization", resourceType),
		map[string]interface{}{
			"resource_type": resourceType,
			"resource_id":   resourceID,
			"user_org":      userOrg,
			"resource_org":  resourceOrg,
		},
	)
}

// NewInvalidStatusTransitionError creates a status transition error
func NewInvalidStatusTransitionError(resourceType, resourceID, fromStatus, toStatus string) *AppError {
	return NewAppErrorWithDetails(
		ErrorCodeInvalidStatusTransition,
		fmt.Sprintf("Cannot transition %s from %s to %s", resourceType, fromStatus, toStatus),
		map[string]interface{}{
			"resource_type": resourceType,
			"resource_id":   resourceID,
			"from_status":   fromStatus,
			"to_status":     toStatus,
		},
	)
}

// NewDuplicateResourceError creates a duplicate resource error
func NewDuplicateResourceError(resourceType, field, value string) *AppError {
	return NewAppErrorWithDetails(
		ErrorCodeDuplicateResource,
		fmt.Sprintf("%s with %s '%s' already exists", resourceType, field, value),
		map[string]interface{}{
			"resource_type": resourceType,
			"field":         field,
			"value":         value,
		},
	)
}

// Database error helpers

// NewDatabaseError creates a database-related error
func NewDatabaseError(operation string, cause error) *AppError {
	return NewAppErrorWithCause(
		ErrorCodeDatabaseConnection,
		fmt.Sprintf("Database operation failed: %s", operation),
		cause,
	)
}

// NewDatabaseTimeoutError creates a database timeout error
func NewDatabaseTimeoutError(operation string) *AppError {
	return NewAppError(
		ErrorCodeDatabaseTimeout,
		fmt.Sprintf("Database operation timed out: %s", operation),
	)
}

// External service error helpers

// NewExternalServiceError creates an external service error
func NewExternalServiceError(serviceName string, cause error) *AppError {
	return NewAppErrorWithCause(
		ErrorCodeExternalServiceUnavailable,
		fmt.Sprintf("External service unavailable: %s", serviceName),
		cause,
	)
}

// NewAuthServiceError creates an authentication service error
func NewAuthServiceError(operation string, cause error) *AppError {
	return NewAppErrorWithCause(
		ErrorCodeAuthenticationServiceError,
		fmt.Sprintf("Authentication service error during %s", operation),
		cause,
	)
}

// Error recovery helpers

// IsRetryableError checks if an error is retryable
func IsRetryableError(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		switch appErr.Code {
		case ErrorCodeDatabaseTimeout, ErrorCodeExternalServiceTimeout,
			ErrorCodeExternalServiceUnavailable, ErrorCodeDatabaseConnection:
			return true
		}
	}
	return false
}

// IsClientError checks if an error is a client error (4xx)
func IsClientError(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.HTTPStatus >= 400 && appErr.HTTPStatus < 500
	}
	return false
}

// IsServerError checks if an error is a server error (5xx)
func IsServerError(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.HTTPStatus >= 500
	}
	return false
}
