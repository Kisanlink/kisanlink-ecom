// Package taxation provides HTTP handlers for tax exemption and taxation-related operations.
package taxation

import (
	"strconv"
	"strings"

	"kisanlink-ecom/internal/common"
	taxationService "kisanlink-ecom/internal/services/taxation"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// TaxExemptionHandler handles HTTP requests for tax exemption operations
type TaxExemptionHandler struct {
	service taxationService.TaxExemptionServiceInterface
}

// NewTaxExemptionHandler creates a new tax exemption handler
func NewTaxExemptionHandler(service taxationService.TaxExemptionServiceInterface) *TaxExemptionHandler {
	return &TaxExemptionHandler{
		service: service,
	}
}

// CreateTaxExemption handles creating a new tax exemption
func (h *TaxExemptionHandler) CreateTaxExemption(c *gin.Context) {
	var req struct {
		OrganizationID string          `json:"organization_id" binding:"required"`
		ExemptionID    string          `json:"exemption_id" binding:"required"`
		Name           string          `json:"name" binding:"required"`
		ExemptionType  string          `json:"exemption_type" binding:"required"`
		ExemptionRate  decimal.Decimal `json:"exemption_rate" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	taxExemption, err := h.service.CreateTaxExemption(c.Request.Context(), req.OrganizationID, req.ExemptionID, req.Name, req.ExemptionType, req.ExemptionRate)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			common.Error(c, 409, string(common.ErrorCodeDuplicateRecord), err.Error(), nil)
			return
		}
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "must be") {
			common.BadRequest(c, string(common.ErrorCodeValidationFailed), err.Error(), nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to create tax exemption", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Created(c, taxExemption, nil)
}

// GetTaxExemption handles getting a tax exemption by ID
func (h *TaxExemptionHandler) GetTaxExemption(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Tax exemption ID is required", nil)
		return
	}

	taxExemption, err := h.service.GetTaxExemption(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "Tax exemption not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get tax exemption", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, taxExemption, nil)
}

// ListTaxExemptions handles listing all tax exemptions
func (h *TaxExemptionHandler) ListTaxExemptions(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	taxExemptions, total, err := h.service.ListTaxExemptions(c.Request.Context(), limit, offset)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to list tax exemptions", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	meta := &common.ResponseMeta{
		Pagination: common.NewPaginationMeta(page, limit, total),
	}

	common.Success(c, taxExemptions, meta)
}

// GetValidTaxExemptions handles getting valid tax exemptions for an organization
func (h *TaxExemptionHandler) GetValidTaxExemptions(c *gin.Context) {
	orgID := c.Query("organization_id")
	if orgID == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Organization ID is required", nil)
		return
	}

	taxExemptions, err := h.service.GetValidExemptions(c.Request.Context(), orgID)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get valid exemptions", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, taxExemptions, nil)
}

// CalculateTaxExemption handles calculating tax exemption
func (h *TaxExemptionHandler) CalculateTaxExemption(c *gin.Context) {
	exemptionID := c.Query("exemption_id")
	taxAmountStr := c.Query("tax_amount")

	if exemptionID == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Exemption ID is required", nil)
		return
	}
	if taxAmountStr == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Tax amount is required", nil)
		return
	}

	taxAmount, err := decimal.NewFromString(taxAmountStr)
	if err != nil {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Invalid tax amount", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	exemptionAmount, err := h.service.CalculateExemption(c.Request.Context(), exemptionID, taxAmount)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), err.Error(), nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to calculate exemption", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, gin.H{
		"exemption_id":     exemptionID,
		"tax_amount":       taxAmount,
		"exemption_amount": exemptionAmount,
		"net_tax":          taxAmount.Sub(exemptionAmount),
	}, nil)
}

// UpdateTaxExemption handles updating a tax exemption
func (h *TaxExemptionHandler) UpdateTaxExemption(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Tax exemption ID is required", nil)
		return
	}

	// Get existing tax exemption
	existingExemption, err := h.service.GetTaxExemption(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "Tax exemption not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get tax exemption", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Bind update request
	var req struct {
		Name          *string          `json:"name"`
		ExemptionType *string          `json:"exemption_type"`
		ExemptionRate *decimal.Decimal `json:"exemption_rate"`
		IsActive      *bool            `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Update fields if provided
	if req.Name != nil {
		existingExemption.Name = *req.Name
	}
	if req.ExemptionType != nil {
		existingExemption.ExemptionType = *req.ExemptionType
	}
	if req.ExemptionRate != nil {
		existingExemption.ExemptionRate = *req.ExemptionRate
	}
	if req.IsActive != nil {
		existingExemption.IsActive = *req.IsActive
	}

	updatedExemption, err := h.service.UpdateTaxExemption(c.Request.Context(), existingExemption)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to update tax exemption", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, updatedExemption, nil)
}

// DeleteTaxExemption handles deleting a tax exemption
func (h *TaxExemptionHandler) DeleteTaxExemption(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Tax exemption ID is required", nil)
		return
	}

	err := h.service.DeleteTaxExemption(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "Tax exemption not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to delete tax exemption", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	c.Status(204) // No Content
}
