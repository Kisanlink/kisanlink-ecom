package catalog

import (
	"time"

	catalogModels "kisanlink-ecom/entities/models/catalog"
	"kisanlink-ecom/internal/common"
	"kisanlink-ecom/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// PublishRequest represents the request to publish a catalog item
type PublishRequest struct {
	Visibility    *string    `json:"visibility,omitempty" binding:"omitempty,oneof=PRIVATE ORG NETWORK PUBLIC"`
	EffectiveDate *time.Time `json:"effective_date,omitempty"`
	Notes         string     `json:"notes,omitempty" binding:"omitempty,max=500"`
}

// UnpublishRequest represents the request to unpublish a catalog item
type UnpublishRequest struct {
	Reason             string     `json:"reason,omitempty" binding:"omitempty,max=500"`
	EffectiveDate      *time.Time `json:"effective_date,omitempty"`
	HandleActiveOrders string     `json:"handle_active_orders,omitempty" binding:"omitempty,oneof=fulfill cancel transfer"`
}

// PriceUpdateRequest represents the request to update catalog item price
type PriceUpdateRequest struct {
	BasePrice        float64    `json:"base_price" binding:"required,min=0,default=0"`
	Currency         string     `json:"currency,omitempty" binding:"omitempty,len=3"`
	EffectiveDate    *time.Time `json:"effective_date,omitempty"`
	Reason           string     `json:"reason,omitempty" binding:"omitempty,max=500"`
	Promotional      bool       `json:"promotional,omitempty"`
	PromotionEndDate *time.Time `json:"promotion_end_date,omitempty"`
}

// PublishCatalogItem godoc
// @Summary Publish catalog item
// @Description Publish a catalog item to make it visible according to its visibility settings
// @Tags Catalog Management
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param type path string true "Item type (products, services, labour, contracts)"
// @Param id path string true "Item ID"
// @Param request body PublishRequest false "Publish options"
// @Success 200 {object} common.Response{data=catalog.CatalogItemResponse}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Failure 409 {object} common.Response{error=common.ResponseError}
// @Failure 422 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/{type}/{id}/publish [post]
func (h *CatalogHandler) PublishCatalogItem(c *gin.Context) {
	typeParam := c.Param("type")
	id := c.Param("id")

	if typeParam == "" {
		common.BadRequest(c, "MISSING_TYPE", "Catalog type is required", nil)
		return
	}
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Item ID is required", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	var req PublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Get the catalog item first
	item, err := h.catalogService.GetCatalogItemByID(c.Request.Context(), id)
	if err != nil {
		common.NotFound(c, "ITEM_NOT_FOUND", "Catalog item not found", map[string]interface{}{
			"item_id": id,
		})
		return
	}

	// Check if user has permission to publish this item
	if orgID, ok := middleware.GetOrgID(c); ok {
		if item.OrganizationID != orgID {
			common.Forbidden(c, "INSUFFICIENT_PERMISSIONS", "Cannot publish item from different organization", nil)
			return
		}
	}

	// Update visibility if provided
	if req.Visibility != nil {
		switch *req.Visibility {
		case "PRIVATE":
			item.Visibility = catalogModels.VisibilityPrivate
		case "ORG":
			item.Visibility = catalogModels.VisibilityOrg
		case "NETWORK":
			item.Visibility = catalogModels.VisibilityNetwork
		case "PUBLIC":
			item.Visibility = catalogModels.VisibilityPublic
		}
	}

	// Set item as active (published)
	item.IsActive = true
	item.UpdatedBy = userID.(string)

	// Update the item
	updatedItem, err := h.catalogService.UpdateCatalogItem(c.Request.Context(), item, userID.(string))
	if err != nil {
		common.InternalServerError(c, "PUBLISH_FAILED", "Failed to publish catalog item", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, updatedItem, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// UnpublishCatalogItem godoc
// @Summary Unpublish catalog item
// @Description Unpublish a catalog item to make it inactive
// @Tags Catalog Management
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param type path string true "Item type (products, services, labour, contracts)"
// @Param id path string true "Item ID"
// @Param request body UnpublishRequest false "Unpublish options"
// @Success 200 {object} common.Response{data=catalog.CatalogItemResponse}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Failure 409 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/{type}/{id}/unpublish [post]
func (h *CatalogHandler) UnpublishCatalogItem(c *gin.Context) {
	typeParam := c.Param("type")
	id := c.Param("id")

	if typeParam == "" {
		common.BadRequest(c, "MISSING_TYPE", "Catalog type is required", nil)
		return
	}
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Item ID is required", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	var req UnpublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Get the catalog item first
	item, err := h.catalogService.GetCatalogItemByID(c.Request.Context(), id)
	if err != nil {
		common.NotFound(c, "ITEM_NOT_FOUND", "Catalog item not found", map[string]interface{}{
			"item_id": id,
		})
		return
	}

	// Check if user has permission to unpublish this item
	if orgID, ok := middleware.GetOrgID(c); ok {
		if item.OrganizationID != orgID {
			common.Forbidden(c, "INSUFFICIENT_PERMISSIONS", "Cannot unpublish item from different organization", nil)
			return
		}
	}

	// Set item as inactive (unpublished)
	item.IsActive = false
	item.UpdatedBy = userID.(string)

	// Update the item
	updatedItem, err := h.catalogService.UpdateCatalogItem(c.Request.Context(), item, userID.(string))
	if err != nil {
		common.InternalServerError(c, "UNPUBLISH_FAILED", "Failed to unpublish catalog item", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, updatedItem, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// UpdateCatalogItemPrice godoc
// @Summary Update catalog item price
// @Description Update the price of a catalog item with price history tracking
// @Tags Catalog Management
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param type path string true "Item type (products, services, labour, contracts)"
// @Param id path string true "Item ID"
// @Param request body PriceUpdateRequest true "Price update data"
// @Success 200 {object} common.Response{data=catalog.CatalogItemResponse}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Failure 422 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/{type}/{id}/price [put]
func (h *CatalogHandler) UpdateCatalogItemPrice(c *gin.Context) {
	typeParam := c.Param("type")
	id := c.Param("id")

	if typeParam == "" {
		common.BadRequest(c, "MISSING_TYPE", "Catalog type is required", nil)
		return
	}
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Item ID is required", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	var req PriceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Get the catalog item first
	item, err := h.catalogService.GetCatalogItemByID(c.Request.Context(), id)
	if err != nil {
		common.NotFound(c, "ITEM_NOT_FOUND", "Catalog item not found", map[string]interface{}{
			"item_id": id,
		})
		return
	}

	// Check if user has permission to update price for this item
	if orgID, ok := middleware.GetOrgID(c); ok {
		if item.OrganizationID != orgID {
			common.Forbidden(c, "INSUFFICIENT_PERMISSIONS", "Cannot update price for item from different organization", nil)
			return
		}
	}

	// Update the price
	if req.BasePrice > 0 {
		item.BasePrice = decimal.NewFromFloat(req.BasePrice)
	}
	if req.Currency != "" {
		item.Currency = req.Currency
	}
	item.UpdatedBy = userID.(string)

	// Update the item
	updatedItem, err := h.catalogService.UpdateCatalogItem(c.Request.Context(), item, userID.(string))
	if err != nil {
		common.InternalServerError(c, "PRICE_UPDATE_FAILED", "Failed to update catalog item price", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, updatedItem, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}
