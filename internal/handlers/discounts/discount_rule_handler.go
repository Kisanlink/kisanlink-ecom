// Package discounts provides HTTP handlers for discount rule operations.
package discounts

import (
	"strconv"
	"strings"

	"kisanlink-ecom/internal/common"
	discountsService "kisanlink-ecom/internal/services/discounts"

	"github.com/gin-gonic/gin"
)

// DiscountRuleHandler handles HTTP requests for discount rule operations
type DiscountRuleHandler struct {
	service discountsService.DiscountRuleServiceInterface
}

// NewDiscountRuleHandler creates a new discount rule handler
func NewDiscountRuleHandler(service discountsService.DiscountRuleServiceInterface) *DiscountRuleHandler {
	return &DiscountRuleHandler{
		service: service,
	}
}

// CreateDiscountRule handles creating a new discount rule
func (h *DiscountRuleHandler) CreateDiscountRule(c *gin.Context) {
	var req struct {
		OrganizationID string `json:"organization_id" binding:"required"`
		Name           string `json:"name" binding:"required"`
		RuleType       string `json:"rule_type" binding:"required"`
		Priority       int    `json:"priority"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	discountRule, err := h.service.CreateDiscountRule(c.Request.Context(), req.OrganizationID, req.Name, req.RuleType, req.Priority)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			common.BadRequest(c, string(common.ErrorCodeValidationFailed), err.Error(), nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to create discount rule", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Created(c, discountRule, nil)
}

// GetDiscountRule handles getting a discount rule by ID
func (h *DiscountRuleHandler) GetDiscountRule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Discount rule ID is required", nil)
		return
	}

	discountRule, err := h.service.GetDiscountRule(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "Discount rule not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get discount rule", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, discountRule, nil)
}

// ListDiscountRules handles listing all discount rules
func (h *DiscountRuleHandler) ListDiscountRules(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	discountRules, total, err := h.service.ListDiscountRules(c.Request.Context(), limit, offset)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to list discount rules", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	meta := &common.ResponseMeta{
		Pagination: common.NewPaginationMeta(page, limit, total),
	}

	common.Success(c, discountRules, meta)
}

// GetOrganizationRules handles getting discount rules for a specific organization
func (h *DiscountRuleHandler) GetOrganizationRules(c *gin.Context) {
	orgID := c.Param("organizationId")
	if orgID == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Organization ID is required", nil)
		return
	}

	rules, err := h.service.GetOrganizationRules(c.Request.Context(), orgID)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get organization rules", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, rules, nil)
}

// GetActiveRules handles getting active discount rules
func (h *DiscountRuleHandler) GetActiveRules(c *gin.Context) {
	orgID := c.Query("organization_id")
	if orgID == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Organization ID is required", nil)
		return
	}

	rules, err := h.service.GetActiveRules(c.Request.Context(), orgID)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get active rules", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, rules, nil)
}

// UpdateDiscountRule handles updating a discount rule
func (h *DiscountRuleHandler) UpdateDiscountRule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Discount rule ID is required", nil)
		return
	}

	existingRule, err := h.service.GetDiscountRule(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "Discount rule not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get discount rule", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	var req struct {
		Name     *string `json:"name"`
		RuleType *string `json:"rule_type"`
		Priority *int    `json:"priority"`
		IsActive *bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	if req.Name != nil {
		existingRule.Name = *req.Name
	}
	if req.RuleType != nil {
		existingRule.RuleType = *req.RuleType
	}
	if req.Priority != nil {
		existingRule.Priority = *req.Priority
	}
	if req.IsActive != nil {
		existingRule.IsActive = *req.IsActive
	}

	updatedRule, err := h.service.UpdateDiscountRule(c.Request.Context(), existingRule)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to update discount rule", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, updatedRule, nil)
}

// DeleteDiscountRule handles deleting a discount rule
func (h *DiscountRuleHandler) DeleteDiscountRule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Discount rule ID is required", nil)
		return
	}

	err := h.service.DeleteDiscountRule(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "Discount rule not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to delete discount rule", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	c.Status(204)
}
