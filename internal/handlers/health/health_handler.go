package health

import (
	"net/http"

	"kisanlink-ecom/internal/models/common"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthCheck handles the health check endpoint
// @Summary Health check
// @Description Check if the service is running
// @Tags health
// @Produce json
// @Success 200 {object} common.HealthResponse
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	response := common.HealthResponse{
		Status:  "ok",
		Message: "KisanLink E-commerce API is running",
		Version: "1.0.0",
	}

	c.JSON(http.StatusOK, response)
}
