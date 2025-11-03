package catalog

import (
	"time"

	catalogModels "kisanlink-ecom/entities/models/catalog"
	"kisanlink-ecom/internal/common"
	"kisanlink-ecom/internal/middleware"
	catalogService "kisanlink-ecom/internal/services/catalog"

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

// FPO-specific publishing endpoints

// FPOPublishHandler handles FPO-specific product publishing operations
type FPOPublishHandler struct {
	publishService catalogService.PublishService
}

// NewFPOPublishHandler creates a new FPO publish handler
func NewFPOPublishHandler(publishService catalogService.PublishService) *FPOPublishHandler {
	return &FPOPublishHandler{
		publishService: publishService,
	}
}

// PublishProductToFPOsRequest represents the request to publish a product to FPOs
type PublishProductToFPOsRequest struct {
	FPOIDs             []string           `json:"fpo_ids" binding:"required,min=1"`
	DeliveryCosts      map[string]float64 `json:"delivery_costs" binding:"required"`
	PlatformFeePercent float64            `json:"platform_fee_percent" binding:"required,min=0,max=100"`
}

// PublishProductToFPOsResponse represents the response for publishing a product to FPOs
type PublishProductToFPOsResponse struct {
	ProductID      string    `json:"product_id"`
	Status         string    `json:"status"`
	VisibleToFPOs  int       `json:"visible_to_fpos"`
	PublishedAt    time.Time `json:"published_at"`
	PlatformFee    float64   `json:"platform_fee_percent"`
	PublishStateID string    `json:"publish_state_id"`
}

// UpdateDeliveryCostsRequest represents the request to update delivery costs
type UpdateDeliveryCostsRequest struct {
	DeliveryCosts map[string]float64 `json:"delivery_costs" binding:"required"`
}

// PublishProductToFPOs godoc
// @Summary Publish product to FPOs
// @Description Publish a product to specific FPOs with delivery costs and platform fee configuration
// @Tags Product Publishing
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Product ID"
// @Param request body PublishProductToFPOsRequest true "Publish configuration"
// @Success 200 {object} common.Response{data=PublishProductToFPOsResponse}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/products/{id}/publish [post]
func (h *FPOPublishHandler) PublishProductToFPOs(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		common.BadRequest(c, "MISSING_PRODUCT_ID", "Product ID is required", nil)
		return
	}

	// Get user ID from context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	var req PublishProductToFPOsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Validate all FPO IDs have delivery costs
	for _, fpoID := range req.FPOIDs {
		if _, exists := req.DeliveryCosts[fpoID]; !exists {
			common.BadRequest(c, "MISSING_DELIVERY_COST", "Delivery cost required for all FPOs", map[string]interface{}{
				"fpo_id": fpoID,
			})
			return
		}
	}

	// Convert delivery costs to decimal
	deliveryCosts := make(map[string]decimal.Decimal)
	for fpoID, cost := range req.DeliveryCosts {
		deliveryCosts[fpoID] = decimal.NewFromFloat(cost)
	}

	// Create publish request
	publishReq := catalogService.PublishProductRequest{
		ProductID:          productID,
		FPOIDs:             req.FPOIDs,
		DeliveryCosts:      deliveryCosts,
		PlatformFeePercent: decimal.NewFromFloat(req.PlatformFeePercent),
		PublishedBy:        userID,
	}

	// Publish product
	publishState, err := h.publishService.PublishProduct(c.Request.Context(), publishReq)
	if err != nil {
		common.InternalServerError(c, "PUBLISH_FAILED", "Failed to publish product", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Prepare response
	response := PublishProductToFPOsResponse{
		ProductID:      publishState.ProductID,
		Status:         "published",
		VisibleToFPOs:  len(publishState.FPOAccessList),
		PublishedAt:    *publishState.PublishedAt,
		PlatformFee:    publishState.PlatformFeePercent.InexactFloat64(),
		PublishStateID: publishState.ID,
	}

	common.Success(c, response, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// GetPublishStatus godoc
// @Summary Get product publish status
// @Description Retrieve the publish status and FPO access configuration for a product
// @Tags Product Publishing
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Product ID"
// @Success 200 {object} common.Response{data=catalog.PublishState}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/products/{id}/publish-status [get]
func (h *FPOPublishHandler) GetPublishStatus(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		common.BadRequest(c, "MISSING_PRODUCT_ID", "Product ID is required", nil)
		return
	}

	// Get publish status
	publishState, err := h.publishService.GetPublishStatus(c.Request.Context(), productID)
	if err != nil {
		common.NotFound(c, "PUBLISH_STATE_NOT_FOUND", "Product is not published", map[string]interface{}{
			"product_id": productID,
			"error":      err.Error(),
		})
		return
	}

	common.Success(c, publishState, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// UpdateProductDeliveryCosts godoc
// @Summary Update delivery costs for a product
// @Description Update the delivery costs for specific FPOs for a published product
// @Tags Product Publishing
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Product ID"
// @Param request body UpdateDeliveryCostsRequest true "Updated delivery costs"
// @Success 200 {object} common.Response{data=map[string]string}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/products/{id}/delivery-costs [patch]
func (h *FPOPublishHandler) UpdateProductDeliveryCosts(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		common.BadRequest(c, "MISSING_PRODUCT_ID", "Product ID is required", nil)
		return
	}

	var req UpdateDeliveryCostsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Convert delivery costs to decimal
	deliveryCosts := make(map[string]decimal.Decimal)
	for fpoID, cost := range req.DeliveryCosts {
		deliveryCosts[fpoID] = decimal.NewFromFloat(cost)
	}

	// Update delivery costs
	if err := h.publishService.UpdateDeliveryCosts(c.Request.Context(), productID, deliveryCosts); err != nil {
		common.InternalServerError(c, "UPDATE_FAILED", "Failed to update delivery costs", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, map[string]string{
		"message":    "Delivery costs updated successfully",
		"product_id": productID,
	}, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// RevokeFPOAccess godoc
// @Summary Revoke FPO access to a product
// @Description Remove a specific FPO's access to a published product
// @Tags Product Publishing
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Product ID"
// @Param fpo_id path string true "FPO Organization ID"
// @Success 200 {object} common.Response{data=map[string]string}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/products/{id}/fpo-access/{fpo_id} [delete]
func (h *FPOPublishHandler) RevokeFPOAccess(c *gin.Context) {
	productID := c.Param("id")
	fpoOrgID := c.Param("fpo_id")

	if productID == "" {
		common.BadRequest(c, "MISSING_PRODUCT_ID", "Product ID is required", nil)
		return
	}
	if fpoOrgID == "" {
		common.BadRequest(c, "MISSING_FPO_ID", "FPO Organization ID is required", nil)
		return
	}

	// Revoke access
	if err := h.publishService.RevokeAccess(c.Request.Context(), productID, fpoOrgID); err != nil {
		common.InternalServerError(c, "REVOKE_FAILED", "Failed to revoke FPO access", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, map[string]string{
		"message":    "FPO access revoked successfully",
		"product_id": productID,
		"fpo_org_id": fpoOrgID,
	}, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}
