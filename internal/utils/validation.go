package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// InitValidator initializes the validator instance.
func InitValidator() {
	validate = validator.New()
}

// ValidateStruct validates a struct and returns formatted error message.
func ValidateStruct(s interface{}) error {
	if validate == nil {
		InitValidator()
	}

	err := validate.Struct(s)
	if err != nil {
		var errors []string
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, formatValidationError(err))
		}
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}
	return nil
}

// formatValidationError formats a validation error into a user-friendly message.
func formatValidationError(err validator.FieldError) string {
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
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// IsValidEmail checks if an email address is valid.
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsValidUsername checks if a username is valid.
func IsValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 50 {
		return false
	}
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return usernameRegex.MatchString(username)
}
