package inventory

import (
	"strconv"

	"github.com/Kisanlink/kisanlink-ecom/internal/common"
	inventoryService "github.com/Kisanlink/kisanlink-ecom/internal/services/inventory"

	"github.com/gin-gonic/gin"
)

// AlertHandler handles HTTP requests for inventory alerts
type AlertHandler struct {
	alertService inventoryService.AlertService
}

// NewAlertHandler creates a new alert handler
func NewAlertHandler(alertService inventoryService.AlertService) *AlertHandler {
	return &AlertHandler{
		alertService: alertService,
	}
}

// UpdateAlertConfigRequest represents the request body for updating alert configuration
type UpdateAlertConfigRequest struct {
	LowStockThreshold int      `json:"low_stock_threshold" binding:"required,gte=0" example:"100"`
	ExpiryWarningDays int      `json:"expiry_warning_days" binding:"required,gte=1" example:"30"`
	AlertChannels     []string `json:"alert_channels" example:"[\"EMAIL\",\"SMS\"]"`
}

// GetAlerts lists inventory alerts for an FPO
// @Summary List inventory alerts
// @Description List inventory alerts for an FPO organization with pagination
// @Tags inventory-alerts
// @Accept json
// @Produce json
// @Param offset query int false "Pagination offset" default(0)
// @Param limit query int false "Pagination limit (max 100)" default(20)
// @Success 200 {object} common.Response
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 500 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/fpo/inventory/alerts [get]
func (h *AlertHandler) GetAlerts(c *gin.Context) {
	// Get user context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Forbidden(c, "FORBIDDEN", "Organization context required", nil)
		return
	}

	// Parse pagination
	offset := 0
	limit := 20

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	// Get alerts
	response, err := h.alertService.GetAlertsByFPO(c.Request.Context(), orgID, offset, limit, userID)
	if err != nil {
		common.InternalServerError(c, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	common.Success(c, response, nil)
}

// GetActiveAlerts lists active inventory alerts for an FPO
// @Summary List active inventory alerts
// @Description List active (unacknowledged) inventory alerts for an FPO organization
// @Tags inventory-alerts
// @Accept json
// @Produce json
// @Param offset query int false "Pagination offset" default(0)
// @Param limit query int false "Pagination limit (max 100)" default(20)
// @Success 200 {object} common.Response
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 500 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/fpo/inventory/alerts/active [get]
func (h *AlertHandler) GetActiveAlerts(c *gin.Context) {
	// Get user context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Forbidden(c, "FORBIDDEN", "Organization context required", nil)
		return
	}

	// Parse pagination
	offset := 0
	limit := 20

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	// Get active alerts
	response, err := h.alertService.GetActiveAlerts(c.Request.Context(), orgID, offset, limit, userID)
	if err != nil {
		common.InternalServerError(c, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	common.Success(c, response, nil)
}

// AcknowledgeAlert acknowledges an inventory alert
// @Summary Acknowledge inventory alert
// @Description Mark an inventory alert as acknowledged
// @Tags inventory-alerts
// @Accept json
// @Produce json
// @Param id path string true "Alert ID"
// @Success 200 {object} common.Response
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Failure 500 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/fpo/inventory/alerts/{id}/acknowledge [post]
func (h *AlertHandler) AcknowledgeAlert(c *gin.Context) {
	alertID := c.Param("id")
	if alertID == "" {
		common.BadRequest(c, "VALIDATION_ERROR", "Alert ID is required", nil)
		return
	}

	// Get user context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Forbidden(c, "FORBIDDEN", "Organization context required", nil)
		return
	}

	// Acknowledge alert
	if err := h.alertService.AcknowledgeAlert(c.Request.Context(), alertID, userID, orgID); err != nil {
		common.BadRequest(c, "SERVICE_ERROR", err.Error(), nil)
		return
	}

	common.Success(c, nil, nil)
}

// GetAlertConfig retrieves alert configuration for an FPO
// @Summary Get alert configuration
// @Description Get alert configuration for an FPO organization
// @Tags inventory-alerts
// @Accept json
// @Produce json
// @Success 200 {object} common.Response
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 500 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/fpo/inventory/alert-config [get]
func (h *AlertHandler) GetAlertConfig(c *gin.Context) {
	// Get user context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Forbidden(c, "FORBIDDEN", "Organization context required", nil)
		return
	}

	// Get alert config
	config, err := h.alertService.GetAlertConfig(c.Request.Context(), orgID, userID)
	if err != nil {
		common.InternalServerError(c, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	common.Success(c, config, nil)
}

// UpdateAlertConfig updates alert configuration for an FPO
// @Summary Update alert configuration
// @Description Update alert configuration for an FPO organization
// @Tags inventory-alerts
// @Accept json
// @Produce json
// @Param request body UpdateAlertConfigRequest true "Alert configuration update request"
// @Success 200 {object} common.Response
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 500 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/fpo/inventory/alert-config [put]
func (h *AlertHandler) UpdateAlertConfig(c *gin.Context) {
	var req UpdateAlertConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	// Get user context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Forbidden(c, "FORBIDDEN", "Organization context required", nil)
		return
	}

	// Convert to service request
	serviceReq := &inventoryService.UpdateAlertConfigRequest{
		LowStockThreshold: req.LowStockThreshold,
		ExpiryWarningDays: req.ExpiryWarningDays,
		AlertChannels:     req.AlertChannels,
	}

	// Update alert config
	config, err := h.alertService.UpdateAlertConfig(c.Request.Context(), orgID, serviceReq, userID)
	if err != nil {
		common.BadRequest(c, "SERVICE_ERROR", err.Error(), nil)
		return
	}

	common.Success(c, config, nil)
}

// GetAllAlerts lists all alerts across all FPOs (admin only)
// @Summary List all inventory alerts (admin)
// @Description List all inventory alerts across all FPO organizations (admin only)
// @Tags inventory-alerts
// @Accept json
// @Produce json
// @Param offset query int false "Pagination offset" default(0)
// @Param limit query int false "Pagination limit (max 100)" default(20)
// @Param fpo_org_id query string false "Filter by FPO organization ID"
// @Success 200 {object} common.Response
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 500 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/admin/inventory/alerts [get]
func (h *AlertHandler) GetAllAlerts(c *gin.Context) {
	// Get user context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	// Parse pagination
	offset := 0
	limit := 20

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	// Get FPO filter if provided
	fpoOrgID := c.Query("fpo_org_id")
	if fpoOrgID == "" {
		common.BadRequest(c, "VALIDATION_ERROR", "fpo_org_id is required for admin view", nil)
		return
	}

	// Get alerts for the specified FPO
	response, err := h.alertService.GetAlertsByFPO(c.Request.Context(), fpoOrgID, offset, limit, userID)
	if err != nil {
		common.InternalServerError(c, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	common.Success(c, response, nil)
}
