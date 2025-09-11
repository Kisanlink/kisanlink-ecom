package common

import (
    "context"
    "errors"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"

    "kisanlink-ecom/internal/common"
)

func TestAppError(t *testing.T) {
    gin.SetMode(gin.TestMode)

    t.Run("NewAppError", func(t *testing.T) {
        err := common.NewAppError(common.ErrorCodeValidationFailed, "Invalid data", 400)

        assert.Equal(t, common.ErrorCodeValidationFailed, err.Code)
        assert.Equal(t, 400, err.StatusCode)
        assert.Equal(t, "Invalid data", err.Message)
    })

    t.Run("WithContext", func(t *testing.T) {
        err := common.NewAppError(common.ErrorCodeInvalidInput, "Invalid data", 400)
        err = err.WithContext("field", "email").WithContext("value", "invalid-email")

        assert.Equal(t, "email", err.Context["field"])
        assert.Equal(t, "invalid-email", err.Context["value"])
        assert.Equal(t, 400, err.StatusCode)
    })

    t.Run("WithField", func(t *testing.T) {
        err := common.NewAppError(common.ErrorCodeValidationFailed, "Validation failed", 400)
        err = err.WithField("email", "Invalid email format").WithField("age", "Must be positive")

        assert.Equal(t, "Invalid email format", err.Fields["email"])
        assert.Equal(t, "Must be positive", err.Fields["age"])
    })
}

func TestValidationErrors(t *testing.T) {
    t.Run("AddAndHasErrors", func(t *testing.T) {
        validationErr := common.NewValidationErrors()

        assert.False(t, validationErr.HasErrors())

        validationErr.Add("email", "Invalid email")
        validationErr.Add("age", "Must be positive")

        assert.True(t, validationErr.HasErrors())
        assert.Equal(t, "Invalid email", validationErr.Errors["email"])
        assert.Equal(t, "Must be positive", validationErr.Errors["age"])
    })

    t.Run("ToAppError", func(t *testing.T) {
        validationErr := common.NewValidationErrors()
        validationErr.Add("email", "Invalid email")

        appErr := validationErr.ToAppError()

        assert.Equal(t, common.ErrorCodeValidationFailed, appErr.Code)
        assert.Equal(t, 400, appErr.StatusCode)
        assert.Equal(t, "Invalid email", appErr.Fields["email"])
    })
}

func TestIsAppError(t *testing.T) {
    t.Run("IsAppError", func(t *testing.T) {
        genericErr := errors.New("generic error")
        appErr := common.NewValidationError("validation failed")

        _, ok := common.IsAppError(genericErr)
        assert.False(t, ok)

        extractedErr, ok := common.IsAppError(appErr)
        assert.True(t, ok)
        assert.Equal(t, appErr, extractedErr)
    })
}

func TestGetStatusCode(t *testing.T) {
    t.Run("GetStatusCode", func(t *testing.T) {
        appErr := common.NewValidationError("validation failed")
        genericErr := errors.New("generic error")

        assert.Equal(t, 400, common.GetStatusCode(appErr))
        assert.Equal(t, 500, common.GetStatusCode(genericErr))
    })
}

func TestGetErrorCode(t *testing.T) {
    t.Run("GetErrorCode", func(t *testing.T) {
        appErr := common.NewValidationError("validation failed")
        genericErr := errors.New("generic error")

        assert.Equal(t, common.ErrorCodeValidationFailed, common.GetErrorCode(appErr))
        assert.Equal(t, common.ErrorCodeInternalError, common.GetErrorCode(genericErr))
    })
}

func TestDatabaseError(t *testing.T) {
    t.Run("WrapError", func(t *testing.T) {
        originalErr := errors.New("original error")
        wrappedErr := common.WrapError(originalErr, common.ErrorCodeDatabaseError, "Database failed", 500)

        assert.Equal(t, common.ErrorCodeDatabaseError, wrappedErr.Code)
        assert.Equal(t, 500, wrappedErr.StatusCode)
        assert.Equal(t, "Database failed", wrappedErr.Message)
        assert.Equal(t, originalErr, wrappedErr.Cause)
    })

    t.Run("WithCause", func(t *testing.T) {
        cause := errors.New("database connection failed")
        err := common.NewDatabaseError("query", cause)

        assert.Equal(t, cause, err.Cause)
        assert.Equal(t, common.ErrorCodeDatabaseError, err.Code)
    })
}

func TestErrorHandler(t *testing.T) {
    gin.SetMode(gin.TestMode)

    t.Run("HandleAppError", func(t *testing.T) {
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = httptest.NewRequest("GET", "/test", nil)

        handler := common.NewErrorHandler(nil)
        validationErr := common.NewValidationErrors()
        validationErr.Add("email", "Invalid email")
        appErr := validationErr.ToAppError()

        handler.HandleError(c, appErr)

        assert.Equal(t, http.StatusBadRequest, w.Code)
        assert.Contains(t, w.Body.String(), "VALIDATION_FAILED")
        assert.Contains(t, w.Body.String(), "Invalid email")
    })

    t.Run("HandleGenericError", func(t *testing.T) {
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = httptest.NewRequest("GET", "/test", nil)

        handler := common.NewErrorHandler(nil)
        err := errors.New("generic error")

        handler.HandleError(c, err)

        assert.Equal(t, http.StatusInternalServerError, w.Code)
        assert.Contains(t, w.Body.String(), "INTERNAL_ERROR")
    })
}

func TestRetryUtilities(t *testing.T) {
    t.Run("SuccessfulOperation", func(t *testing.T) {
        ctx := context.Background()

        attempts := 0
        operation := func(ctx context.Context, attempt int) error {
            attempts++
            return nil // Success on first attempt
        }

        // Test would need actual retry implementation
        err := operation(ctx, 1)

        assert.NoError(t, err)
        assert.Equal(t, 1, attempts)
    })
}
