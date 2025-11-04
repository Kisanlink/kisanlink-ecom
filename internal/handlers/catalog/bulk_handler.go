package catalog

import (
	catalogResponses "kisanlink-ecom/entities/responses/catalog"
	"kisanlink-ecom/internal/common"
	"kisanlink-ecom/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// UpdateCatalogItemRequest represents the request to update a catalog item (for swagger)
type UpdateCatalogItemRequest struct {
	Category      *string                `json:"category,omitempty" binding:"omitempty,max=100"`
	Subcategory   *string                `json:"subcategory,omitempty" binding:"omitempty,max=100"`
	Name          *string                `json:"name,omitempty" binding:"omitempty,max=255"`
	Description   *string                `json:"description,omitempty"`
	SKU           *string                `json:"sku,omitempty" binding:"omitempty,max=100"`
	UnitOfMeasure *string                `json:"unit_of_measure,omitempty" binding:"omitempty,max=50"`
	BasePrice     *decimal.Decimal       `json:"base_price,omitempty" binding:"omitempty"`
	Currency      *string                `json:"currency,omitempty" binding:"omitempty,len=3"`
	IsActive      *bool                  `json:"is_active,omitempty"`
	Visibility    *string                `json:"visibility,omitempty" binding:"omitempty,oneof=PRIVATE ORG NETWORK PUBLIC"`
	Tags          []string               `json:"tags,omitempty" binding:"omitempty"`
	Attributes    map[string]interface{} `json:"attributes,omitempty" binding:"omitempty"`
	Images        []string               `json:"images,omitempty" binding:"omitempty"`
}

// BulkUpdateItem represents a single item in a bulk update request
type BulkUpdateItem struct {
	ID      string                   `json:"id" binding:"required,uuid4"`
	Updates UpdateCatalogItemRequest `json:"updates" binding:"required"`
}

// BulkUpdateRequest represents a bulk update request
type BulkUpdateRequest struct {
	Items  []BulkUpdateItem `json:"items" binding:"required,min=1,max=100"`
	Atomic bool             `json:"atomic,omitempty"`
}

// BulkPublishRequest represents a bulk publish request
type BulkPublishRequest struct {
	ItemIDs       []string `json:"item_ids" binding:"required,min=1,max=100,dive,uuid4"`
	Visibility    string   `json:"visibility,omitempty" binding:"omitempty,oneof=PRIVATE ORG NETWORK PUBLIC"`
	EffectiveDate string   `json:"effective_date,omitempty"`
}

// BulkPriceUpdateItem represents a single price update in a bulk request
type BulkPriceUpdateItem struct {
	ID        string  `json:"id" binding:"required,uuid4"`
	BasePrice float64 `json:"base_price" binding:"required,min=0"`
	Currency  string  `json:"currency,omitempty" binding:"omitempty,len=3"`
}

// BulkPriceUpdateRequest represents a bulk price update request
type BulkPriceUpdateRequest struct {
	Updates       []BulkPriceUpdateItem `json:"updates" binding:"required,min=1,max=100"`
	EffectiveDate string                `json:"effective_date,omitempty"`
	Reason        string                `json:"reason,omitempty" binding:"omitempty,max=500"`
}

// BulkUpdateCatalogItems godoc
// @Summary Bulk update catalog items
// @Description Update multiple catalog items in a single operation
// @Tags Bulk Operations
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param Idempotency-Key header string false "Idempotency key for safe retries"
// @Param request body BulkUpdateRequest true "Bulk update data"
// @Success 200 {object} common.Response{data=catalogResponses.BulkOperationResponse}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 422 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/bulk/update [post]
func (h *CatalogHandler) BulkUpdateCatalogItems(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	var req BulkUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Get organization ID for permission checking
	orgID, hasOrgID := middleware.GetOrgID(c)

	response := catalogResponses.NewBulkOperationResponse(len(req.Items), 0, 0)

	for _, item := range req.Items {
		// Get the catalog item first
		catalogItem, err := h.catalogService.GetCatalogItemByID(c.Request.Context(), item.ID)
		if err != nil {
			response.FailureCount++
			response.AddFailureItem(item.ID, "Item not found")
			continue
		}

		// Check permissions
		if hasOrgID && catalogItem.OrganizationID != orgID {
			response.FailureCount++
			response.AddFailureItem(item.ID, "Insufficient permissions")
			continue
		}

		// Apply updates
		if item.Updates.Name != nil {
			catalogItem.Name = *item.Updates.Name
		}
		if item.Updates.Description != nil {
			catalogItem.Description = *item.Updates.Description
		}
		if item.Updates.Category != nil {
			catalogItem.Category = *item.Updates.Category
		}
		if item.Updates.Subcategory != nil {
			catalogItem.Subcategory = *item.Updates.Subcategory
		}
		if item.Updates.BasePrice != nil {
			catalogItem.BasePrice = *item.Updates.BasePrice
		}
		if item.Updates.Currency != nil {
			catalogItem.Currency = *item.Updates.Currency
		}
		if item.Updates.IsActive != nil {
			catalogItem.IsActive = *item.Updates.IsActive
		}
		if item.Updates.Visibility != nil {
			switch *item.Updates.Visibility {
			case "PRIVATE":
				catalogItem.Visibility = "PRIVATE"
			case "ORG":
				catalogItem.Visibility = "ORG"
			case "NETWORK":
				catalogItem.Visibility = "NETWORK"
			case "PUBLIC":
				catalogItem.Visibility = "PUBLIC"
			}
		}
		if item.Updates.Tags != nil {
			catalogItem.Tags = item.Updates.Tags
		}

		catalogItem.UpdatedBy = userID.(string)

		// Update the item
		_, err = h.catalogService.UpdateCatalogItem(c.Request.Context(), catalogItem, userID.(string))
		if err != nil {
			response.FailureCount++
			response.AddFailureItem(item.ID, err.Error())
			continue
		}

		response.SuccessCount++
		response.AddSuccessItem(item.ID)
	}

	common.Success(c, response, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// BulkPublishCatalogItems godoc
// @Summary Bulk publish catalog items
// @Description Publish multiple catalog items in a single operation
// @Tags Bulk Operations
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body BulkPublishRequest true "Bulk publish data"
// @Success 200 {object} common.Response{data=catalogResponses.BulkOperationResponse}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 422 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/bulk/publish [post]
func (h *CatalogHandler) BulkPublishCatalogItems(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	var req BulkPublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Get organization ID for permission checking
	orgID, hasOrgID := middleware.GetOrgID(c)

	response := catalogResponses.NewBulkOperationResponse(len(req.ItemIDs), 0, 0)

	for _, itemID := range req.ItemIDs {
		// Get the catalog item first
		catalogItem, err := h.catalogService.GetCatalogItemByID(c.Request.Context(), itemID)
		if err != nil {
			response.FailureCount++
			response.AddFailureItem(itemID, "Item not found")
			continue
		}

		// Check permissions
		if hasOrgID && catalogItem.OrganizationID != orgID {
			response.FailureCount++
			response.AddFailureItem(itemID, "Insufficient permissions")
			continue
		}

		// Set as published
		catalogItem.IsActive = true
		catalogItem.UpdatedBy = userID.(string)

		// Update visibility if provided
		if req.Visibility != "" {
			switch req.Visibility {
			case "PRIVATE":
				catalogItem.Visibility = "PRIVATE"
			case "ORG":
				catalogItem.Visibility = "ORG"
			case "NETWORK":
				catalogItem.Visibility = "NETWORK"
			case "PUBLIC":
				catalogItem.Visibility = "PUBLIC"
			}
		}

		// Update the item
		_, err = h.catalogService.UpdateCatalogItem(c.Request.Context(), catalogItem, userID.(string))
		if err != nil {
			response.FailureCount++
			response.AddFailureItem(itemID, err.Error())
			continue
		}

		response.SuccessCount++
		response.AddSuccessItem(itemID)
	}

	common.Success(c, response, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// BulkUpdatePrices godoc
// @Summary Bulk update prices
// @Description Update prices for multiple catalog items
// @Tags Bulk Operations
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body BulkPriceUpdateRequest true "Bulk price update data"
// @Success 200 {object} common.Response{data=catalogResponses.BulkOperationResponse}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 422 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/bulk/price-update [post]
func (h *CatalogHandler) BulkUpdatePrices(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	var req BulkPriceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Get organization ID for permission checking
	orgID, hasOrgID := middleware.GetOrgID(c)

	response := catalogResponses.NewBulkOperationResponse(len(req.Updates), 0, 0)

	for _, update := range req.Updates {
		// Get the catalog item first
		catalogItem, err := h.catalogService.GetCatalogItemByID(c.Request.Context(), update.ID)
		if err != nil {
			response.FailureCount++
			response.AddFailureItem(update.ID, "Item not found")
			continue
		}

		// Check permissions
		if hasOrgID && catalogItem.OrganizationID != orgID {
			response.FailureCount++
			response.AddFailureItem(update.ID, "Insufficient permissions")
			continue
		}

		// Update price
		catalogItem.BasePrice = decimal.NewFromFloat(update.BasePrice)
		if update.Currency != "" {
			catalogItem.Currency = update.Currency
		}
		catalogItem.UpdatedBy = userID.(string)

		// Update the item
		_, err = h.catalogService.UpdateCatalogItem(c.Request.Context(), catalogItem, userID.(string))
		if err != nil {
			response.FailureCount++
			response.AddFailureItem(update.ID, err.Error())
			continue
		}

		response.SuccessCount++
		response.AddSuccessItem(update.ID)
	}

	common.Success(c, response, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}
