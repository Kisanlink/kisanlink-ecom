package handlers

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"

    "kisanlink-ecom/entities/models/common"
)

func setupTestRouter() *gin.Engine {
    gin.SetMode(gin.TestMode)
    router := gin.New()
    return router
}

func TestHealthCheck(t *testing.T) {
    router := setupTestRouter()
    router.GET("/health", HealthCheck)

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/health", nil)
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var response common.APIResponse
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.True(t, response.Success)
    assert.Equal(t, "Health check successful", response.Message)

    // Verify the health response data
    data, ok := response.Data.(map[string]interface{})
    assert.True(t, ok)
    assert.Equal(t, "ok", data["status"])
    assert.Equal(t, "KisanLink E-commerce API is running", data["message"])
    assert.Equal(t, "1.0.0", data["version"])
}
