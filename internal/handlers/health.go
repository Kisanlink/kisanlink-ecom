package handlers

import (
	"net/http"

	"kisanlink-ecom/internal/models/common"
	"kisanlink-ecom/internal/utils"

	"github.com/gin-gonic/gin"
)

// HealthCheck handles health check requests.
// @Summary      Health Check
// @Description  Check if the API is running and healthy
// @Tags         Health
// @Accept       json
// @Produce      json
// @Success      200  {object}  object  "API is healthy"
// @Router       /health [get]
func HealthCheck(c *gin.Context) {
	response := common.HealthResponse{
		Status:  "ok",
		Message: "KisanLink E-commerce API is running",
		Version: "1.0.0",
	}

	utils.SuccessResponse(c, http.StatusOK, "Health check successful", response)
}
