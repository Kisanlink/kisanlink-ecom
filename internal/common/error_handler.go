package common

import (
	"log"
	"net/http"

	"kisanlink-ecom/entities/models/common"

	"github.com/gin-gonic/gin"
)

// ErrorHandler provides centralized error handling
type ErrorHandler struct {
	logger Logger
}

// Logger interface for error logging
type Logger interface {
	Error(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
}

// DefaultLogger implements Logger using standard log package
type DefaultLogger struct{}

func (l *DefaultLogger) Error(msg string, fields ...interface{}) {
	log.Printf("ERROR: %s %v", msg, fields)
}

func (l *DefaultLogger) Warn(msg string, fields ...interface{}) {
	log.Printf("WARN: %s %v", msg, fields)
}

func (l *DefaultLogger) Info(msg string, fields ...interface{}) {
	log.Printf("INFO: %s %v", msg, fields)
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger Logger) *ErrorHandler {
	if logger == nil {
		logger = &DefaultLogger{}
	}
	return &ErrorHandler{
		logger: logger,
	}
}

// HandleError processes errors and returns appropriate HTTP responses
func (h *ErrorHandler) HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// Log the error with context
	h.logError(c, err)

	// Convert to AppError if not already
	appErr := h.toAppError(err)

	// Create API response
	response := common.APIResponse{
		Success: false,
		Error: &common.APIError{
			Code:    string(appErr.Code),
			Message: appErr.Message,
			Details: appErr.Details,
			Fields:  appErr.Fields,
			Context: h.sanitizeContext(appErr.Context),
		},
	}

	c.JSON(appErr.StatusCode, response)
}

// HandleValidationError handles validation errors specifically
func (h *ErrorHandler) HandleValidationError(c *gin.Context, validationErr *ValidationErrors) {
	h.logger.Warn("Validation error", "errors", validationErr.Errors, "path", c.Request.URL.Path)

	response := common.APIResponse{
		Success: false,
		Error: &common.APIError{
			Code:    string(ErrorCodeValidationFailed),
			Message: "Validation failed",
			Fields:  validationErr.Errors,
		},
	}

	c.JSON(http.StatusBadRequest, response)
}

// HandlePanic handles panics and converts them to errors
func (h *ErrorHandler) HandlePanic(c *gin.Context, recovered interface{}) {
	h.logger.Error("Panic recovered", "panic", recovered, "path", c.Request.URL.Path)

	response := common.APIResponse{
		Success: false,
		Error: &common.APIError{
			Code:    string(ErrorCodeInternalError),
			Message: "Internal server error",
		},
	}

	c.JSON(http.StatusInternalServerError, response)
}

// toAppError converts any error to AppError
func (h *ErrorHandler) toAppError(err error) *AppError {
	if appErr, ok := IsAppError(err); ok {
		return appErr
	}

	// Check for common error types
	switch {
	case IsError(err, ErrRecordNotFound):
		return NewNotFoundError("Resource")
	case IsError(err, ErrUnauthorized):
		return NewUnauthorizedError("Authentication required")
	case IsError(err, ErrForbidden):
		return NewForbiddenError("Access denied")
	case IsError(err, ErrInvalidInput):
		return NewValidationError("Invalid input provided")
	case IsError(err, ErrDatabaseConnection), IsError(err, ErrDatabaseQuery):
		return NewDatabaseError("operation", err)
	case IsError(err, ErrExternalServiceUnavailable):
		return NewExternalServiceError("external service", err)
	default:
		return NewInternalError("An unexpected error occurred")
	}
}

// logError logs error with appropriate context
func (h *ErrorHandler) logError(c *gin.Context, err error) {
	fields := []interface{}{
		"error", err.Error(),
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"user_agent", c.Request.UserAgent(),
		"remote_addr", c.ClientIP(),
	}

	// Add request ID if available
	if requestID := c.GetString("request_id"); requestID != "" {
		fields = append(fields, "request_id", requestID)
	}

	// Add user ID if available
	if userID := c.GetString("user_id"); userID != "" {
		fields = append(fields, "user_id", userID)
	}

	// Add organization ID if available
	if orgID := c.GetString("org_id"); orgID != "" {
		fields = append(fields, "org_id", orgID)
	}

	// Log based on error severity
	if appErr, ok := IsAppError(err); ok {
		switch appErr.StatusCode {
		case 400, 401, 403, 404, 422:
			h.logger.Warn("Client error", fields...)
		case 500, 502, 503, 504:
			h.logger.Error("Server error", fields...)
		default:
			h.logger.Info("Request error", fields...)
		}
	} else {
		h.logger.Error("Unhandled error", fields...)
	}
}

// sanitizeContext removes sensitive information from error context
func (h *ErrorHandler) sanitizeContext(context map[string]interface{}) map[string]interface{} {
	if context == nil {
		return nil
	}

	sanitized := make(map[string]interface{})
	sensitiveKeys := map[string]bool{
		"password":     true,
		"token":        true,
		"secret":       true,
		"key":          true,
		"auth":         true,
		"credential":   true,
		"private":      true,
		"confidential": true,
	}

	for k, v := range context {
		// Check if key contains sensitive information
		sensitive := false
		for sensitiveKey := range sensitiveKeys {
			if contains(k, sensitiveKey) {
				sensitive = true
				break
			}
		}

		if sensitive {
			sanitized[k] = "[REDACTED]"
		} else {
			sanitized[k] = v
		}
	}

	return sanitized
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ErrorMiddleware returns a Gin middleware for error handling
func (h *ErrorHandler) ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				h.HandlePanic(c, recovered)
				c.Abort()
			}
		}()

		c.Next()

		// Handle any errors that occurred during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			h.HandleError(c, err)
		}
	}
}

// RecoveryMiddleware returns a Gin middleware for panic recovery
func (h *ErrorHandler) RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		h.HandlePanic(c, recovered)
	})
}
