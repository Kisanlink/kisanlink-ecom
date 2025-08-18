package common

import (
	"errors"
	"fmt"
)

// Common error definitions
var (
	// Authentication errors
	ErrMissingAuthorizationHeader = errors.New("missing authorization header")
	ErrInvalidAuthorizationFormat = errors.New("invalid authorization format, expected 'Bearer <token>'")
	ErrEmptyToken                 = errors.New("empty token")
	ErrMissingToken               = errors.New("missing token")
	ErrInvalidToken               = errors.New("invalid token")
	ErrTokenExpired               = errors.New("token expired")
	ErrUnauthorized               = errors.New("unauthorized")

	// Authorization errors
	ErrForbidden               = errors.New("forbidden")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
	ErrResourceNotFound        = errors.New("resource not found")
	ErrResourceAccessDenied    = errors.New("resource access denied")

	// Validation errors
	ErrInvalidInput         = errors.New("invalid input")
	ErrMissingRequiredField = errors.New("missing required field")
	ErrInvalidFormat        = errors.New("invalid format")

	// Database errors
	ErrDatabaseConnection  = errors.New("database connection failed")
	ErrDatabaseQuery       = errors.New("database query failed")
	ErrDatabaseTransaction = errors.New("database transaction failed")
	ErrRecordNotFound      = errors.New("record not found")
	ErrDuplicateRecord     = errors.New("duplicate record")

	// Business logic errors
	ErrInsufficientInventory = errors.New("insufficient inventory")
	ErrInvalidOrderStatus    = errors.New("invalid order status")
	ErrInvalidPrice          = errors.New("invalid price")
	ErrInvalidQuantity       = errors.New("invalid quantity")

	// External service errors
	ErrExternalServiceUnavailable = errors.New("external service unavailable")
	ErrExternalServiceTimeout     = errors.New("external service timeout")
	ErrExternalServiceError       = errors.New("external service error")
)

// ErrorWithContext wraps an error with additional context
type ErrorWithContext struct {
	Err     error
	Context map[string]interface{}
}

func (e *ErrorWithContext) Error() string {
	if e.Context == nil || len(e.Context) == 0 {
		return e.Err.Error()
	}
	return fmt.Sprintf("%s: %v", e.Err.Error(), e.Context)
}

func (e *ErrorWithContext) Unwrap() error {
	return e.Err
}

// NewErrorWithContext creates a new error with context
func NewErrorWithContext(err error, context map[string]interface{}) *ErrorWithContext {
	return &ErrorWithContext{
		Err:     err,
		Context: context,
	}
}

// IsError checks if an error is of a specific type
func IsError(err error, target error) bool {
	return errors.Is(err, target)
}

// AsError attempts to extract an error of a specific type
func AsError(err error, target interface{}) bool {
	return errors.As(err, target)
}
