package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"kisanlink-ecom/entities/models/common"
	"kisanlink-ecom/internal/handlers/health"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestHealthCheck(t *testing.T) {
	router := setupTestRouter()
	healthHandler := health.NewHealthHandler()
	router.GET("/health", healthHandler.HealthCheck)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response common.HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response.Status)
	assert.Equal(t, "KisanLink E-commerce API is running", response.Message)
	assert.Equal(t, "1.0.0", response.Version)
}
