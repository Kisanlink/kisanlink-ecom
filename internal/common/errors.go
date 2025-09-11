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

// ErrorCode represents standardized error codes
type ErrorCode string

const (
    // Authentication error codes
    ErrorCodeMissingAuth  ErrorCode = "MISSING_AUTH"
    ErrorCodeInvalidAuth  ErrorCode = "INVALID_AUTH"
    ErrorCodeTokenExpired ErrorCode = "TOKEN_EXPIRED"
    ErrorCodeUnauthorized ErrorCode = "UNAUTHORIZED"

    // Authorization error codes
    ErrorCodeForbidden         ErrorCode = "FORBIDDEN"
    ErrorCodeInsufficientPerms ErrorCode = "INSUFFICIENT_PERMISSIONS"
    ErrorCodeResourceNotFound  ErrorCode = "RESOURCE_NOT_FOUND"
    ErrorCodeAccessDenied      ErrorCode = "ACCESS_DENIED"

    // Validation error codes
    ErrorCodeInvalidInput     ErrorCode = "INVALID_INPUT"
    ErrorCodeMissingField     ErrorCode = "MISSING_FIELD"
    ErrorCodeInvalidFormat    ErrorCode = "INVALID_FORMAT"
    ErrorCodeValidationFailed ErrorCode = "VALIDATION_FAILED"

    // Database error codes
    ErrorCodeDatabaseError     ErrorCode = "DATABASE_ERROR"
    ErrorCodeRecordNotFound    ErrorCode = "RECORD_NOT_FOUND"
    ErrorCodeDuplicateRecord   ErrorCode = "DUPLICATE_RECORD"
    ErrorCodeTransactionFailed ErrorCode = "TRANSACTION_FAILED"

    // Business logic error codes
    ErrorCodeInsufficientInventory ErrorCode = "INSUFFICIENT_INVENTORY"
    ErrorCodeInvalidOrderStatus    ErrorCode = "INVALID_ORDER_STATUS"
    ErrorCodeInvalidPrice          ErrorCode = "INVALID_PRICE"
    ErrorCodeInvalidQuantity       ErrorCode = "INVALID_QUANTITY"
    ErrorCodeBusinessRuleViolation ErrorCode = "BUSINESS_RULE_VIOLATION"

    // External service error codes
    ErrorCodeExternalServiceError ErrorCode = "EXTERNAL_SERVICE_ERROR"
    ErrorCodeServiceUnavailable   ErrorCode = "SERVICE_UNAVAILABLE"
    ErrorCodeServiceTimeout       ErrorCode = "SERVICE_TIMEOUT"

    // Internal error codes
    ErrorCodeInternalError ErrorCode = "INTERNAL_ERROR"
    ErrorCodeConfigError   ErrorCode = "CONFIG_ERROR"
)

// AppError represents a structured application error
type AppError struct {
    Code       ErrorCode              `json:"code"`
    Message    string                 `json:"message"`
    Details    string                 `json:"details,omitempty"`
    Fields     map[string]string      `json:"fields,omitempty"`
    Context    map[string]interface{} `json:"context,omitempty"`
    Cause      error                  `json:"-"`
    StatusCode int                    `json:"-"`
}

func (e *AppError) Error() string {
    if e.Details != "" {
        return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
    }
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
    return e.Cause
}

// WithContext adds context to an AppError
func (e *AppError) WithContext(key string, value interface{}) *AppError {
    if e.Context == nil {
        e.Context = make(map[string]interface{})
    }
    e.Context[key] = value
    return e
}

// WithField adds a field error to an AppError
func (e *AppError) WithField(field, message string) *AppError {
    if e.Fields == nil {
        e.Fields = make(map[string]string)
    }
    e.Fields[field] = message
    return e
}

// WithCause adds a cause error to an AppError
func (e *AppError) WithCause(cause error) *AppError {
    e.Cause = cause
    return e
}

// NewAppError creates a new AppError
func NewAppError(code ErrorCode, message string, statusCode int) *AppError {
    return &AppError{
        Code:       code,
        Message:    message,
        StatusCode: statusCode,
    }
}

// NewValidationError creates a validation error
func NewValidationError(message string) *AppError {
    return &AppError{
        Code:       ErrorCodeValidationFailed,
        Message:    message,
        StatusCode: 400,
    }
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *AppError {
    return &AppError{
        Code:       ErrorCodeResourceNotFound,
        Message:    fmt.Sprintf("%s not found", resource),
        StatusCode: 404,
    }
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *AppError {
    return &AppError{
        Code:       ErrorCodeUnauthorized,
        Message:    message,
        StatusCode: 401,
    }
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *AppError {
    return &AppError{
        Code:       ErrorCodeForbidden,
        Message:    message,
        StatusCode: 403,
    }
}

// NewInternalError creates an internal server error
func NewInternalError(message string) *AppError {
    return &AppError{
        Code:       ErrorCodeInternalError,
        Message:    message,
        StatusCode: 500,
    }
}

// NewDatabaseError creates a database error
func NewDatabaseError(operation string, cause error) *AppError {
    return &AppError{
        Code:       ErrorCodeDatabaseError,
        Message:    fmt.Sprintf("Database operation failed: %s", operation),
        StatusCode: 500,
        Cause:      cause,
    }
}

// NewBusinessRuleError creates a business rule violation error
func NewBusinessRuleError(rule string, details string) *AppError {
    return &AppError{
        Code:       ErrorCodeBusinessRuleViolation,
        Message:    fmt.Sprintf("Business rule violation: %s", rule),
        Details:    details,
        StatusCode: 422,
    }
}

// NewExternalServiceError creates an external service error
func NewExternalServiceError(service string, cause error) *AppError {
    return &AppError{
        Code:       ErrorCodeExternalServiceError,
        Message:    fmt.Sprintf("External service error: %s", service),
        StatusCode: 502,
        Cause:      cause,
    }
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) (*AppError, bool) {
    var appErr *AppError
    if errors.As(err, &appErr) {
        return appErr, true
    }
    return nil, false
}

// GetStatusCode extracts HTTP status code from error
func GetStatusCode(err error) int {
    if appErr, ok := IsAppError(err); ok {
        return appErr.StatusCode
    }
    return 500 // Default to internal server error
}

// GetErrorCode extracts error code from error
func GetErrorCode(err error) ErrorCode {
    if appErr, ok := IsAppError(err); ok {
        return appErr.Code
    }
    return ErrorCodeInternalError
}

// WrapError wraps a generic error as an AppError
func WrapError(err error, code ErrorCode, message string, statusCode int) *AppError {
    return &AppError{
        Code:       code,
        Message:    message,
        StatusCode: statusCode,
        Cause:      err,
    }
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
    Errors map[string]string `json:"errors"`
}

func (v *ValidationErrors) Error() string {
    return "validation failed"
}

func (v *ValidationErrors) Add(field, message string) {
    if v.Errors == nil {
        v.Errors = make(map[string]string)
    }
    v.Errors[field] = message
}

func (v *ValidationErrors) HasErrors() bool {
    return len(v.Errors) > 0
}

// NewValidationErrors creates a new ValidationErrors
func NewValidationErrors() *ValidationErrors {
    return &ValidationErrors{
        Errors: make(map[string]string),
    }
}

// ToAppError converts ValidationErrors to AppError
func (v *ValidationErrors) ToAppError() *AppError {
    return &AppError{
        Code:       ErrorCodeValidationFailed,
        Message:    "Validation failed",
        Fields:     v.Errors,
        StatusCode: 400,
    }
}
