// Package sla provides HTTP handlers for service level agreement operations.
package sla

import (
	"strconv"
	"strings"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/services"
	"github.com/Kisanlink/kisanlink-ecom/internal/common"
	slaService "github.com/Kisanlink/kisanlink-ecom/internal/services/sla"

	"github.com/gin-gonic/gin"
)

// ServiceSLAHandler handles HTTP requests for service SLA operations
type ServiceSLAHandler struct {
	service slaService.ServiceSLAServiceInterface
}

// NewServiceSLAHandler creates a new service SLA handler
func NewServiceSLAHandler(service slaService.ServiceSLAServiceInterface) *ServiceSLAHandler {
	return &ServiceSLAHandler{
		service: service,
	}
}

// CreateServiceSLA handles creating a new service SLA
func (h *ServiceSLAHandler) CreateServiceSLA(c *gin.Context) {
	var req struct {
		OrganizationID string           `json:"organization_id" binding:"required"`
		CatalogItemID  string           `json:"catalog_item_id" binding:"required"`
		Name           string           `json:"name" binding:"required"`
		SLAType        services.SLAType `json:"sla_type" binding:"required"`
		TargetValue    float64          `json:"target_value" binding:"required"`
		Unit           services.SLAUnit `json:"unit" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	serviceSLA, err := h.service.CreateServiceSLA(c.Request.Context(), req.OrganizationID, req.CatalogItemID, req.Name, req.SLAType, req.TargetValue, req.Unit)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			common.Error(c, 409, string(common.ErrorCodeDuplicateRecord), err.Error(), nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to create service SLA", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Created(c, serviceSLA, nil)
}

// GetServiceSLA handles getting a service SLA by ID
func (h *ServiceSLAHandler) GetServiceSLA(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Service SLA ID is required", nil)
		return
	}

	serviceSLA, err := h.service.GetServiceSLA(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "Service SLA not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get service SLA", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, serviceSLA, nil)
}

// ListServiceSLAs handles listing all service SLAs
func (h *ServiceSLAHandler) ListServiceSLAs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	serviceSLAs, total, err := h.service.ListServiceSLAs(c.Request.Context(), limit, offset)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to list service SLAs", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	meta := &common.ResponseMeta{
		Pagination: common.NewPaginationMeta(page, limit, total),
	}

	common.Success(c, serviceSLAs, meta)
}

// GetCatalogItemSLAs handles getting service SLAs for a specific catalog item
func (h *ServiceSLAHandler) GetCatalogItemSLAs(c *gin.Context) {
	catalogItemID := c.Param("catalogItemId")
	if catalogItemID == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Catalog item ID is required", nil)
		return
	}

	slas, err := h.service.GetCatalogItemSLAs(c.Request.Context(), catalogItemID)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get catalog item SLAs", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, slas, nil)
}

// UpdateServiceSLA handles updating a service SLA
func (h *ServiceSLAHandler) UpdateServiceSLA(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Service SLA ID is required", nil)
		return
	}

	existingSLA, err := h.service.GetServiceSLA(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "Service SLA not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get service SLA", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	var req struct {
		Name        *string           `json:"name"`
		SLAType     *services.SLAType `json:"sla_type"`
		TargetValue *float64          `json:"target_value"`
		Unit        *services.SLAUnit `json:"unit"`
		IsActive    *bool             `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	if req.Name != nil {
		existingSLA.Name = *req.Name
	}
	if req.SLAType != nil {
		existingSLA.Type = *req.SLAType
	}
	if req.TargetValue != nil {
		existingSLA.TargetValue = *req.TargetValue
	}
	if req.Unit != nil {
		existingSLA.Unit = *req.Unit
	}
	if req.IsActive != nil {
		existingSLA.IsActive = *req.IsActive
	}

	updatedSLA, err := h.service.UpdateServiceSLA(c.Request.Context(), existingSLA)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to update service SLA", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, updatedSLA, nil)
}

// DeleteServiceSLA handles deleting a service SLA
func (h *ServiceSLAHandler) DeleteServiceSLA(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Service SLA ID is required", nil)
		return
	}

	err := h.service.DeleteServiceSLA(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "Service SLA not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to delete service SLA", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	c.Status(204)
}
