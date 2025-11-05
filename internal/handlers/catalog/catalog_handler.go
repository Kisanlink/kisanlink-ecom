package catalog

import (
	"strconv"
	"strings"

	catalogModels "kisanlink-ecom/entities/models/catalog"
	catalogRequests "kisanlink-ecom/entities/requests/catalog"
	"kisanlink-ecom/internal/common"
	"kisanlink-ecom/internal/middleware"
	catalogService "kisanlink-ecom/internal/services/catalog"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// CatalogHandler handles HTTP requests for generic catalog operations
type CatalogHandler struct {
	catalogService catalogService.CatalogServiceInterface
	etagService    *catalogService.ETagService
}

// NewCatalogHandler creates a new catalog handler
func NewCatalogHandler(catalogService catalogService.CatalogServiceInterface, etagService *catalogService.ETagService) *CatalogHandler {
	return &CatalogHandler{
		catalogService: catalogService,
		etagService:    etagService,
	}
}

// ListCatalogItems godoc
// @Summary List all catalog items
// @Description Retrieve a list of all catalog items with filtering and pagination
// @Tags catalog
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param item_type query string false "Filter by item type (PRODUCT, SERVICE, LABOUR)"
// @Param category query string false "Filter by category"
// @Param subcategory query string false "Filter by subcategory"
// @Param is_active query bool false "Filter by active status"
// @Param visibility query string false "Filter by visibility (PRIVATE, ORG, NETWORK, PUBLIC)"
// @Param min_price query number false "Minimum price filter"
// @Param max_price query number false "Maximum price filter"
// @Param search query string false "Search term"
// @Param tags query []string false "Filter by tags"
// @Param organization_id query string false "Filter by organization ID"
// @Param include_deleted query bool false "Include soft-deleted items (admin only)" default(false)
// @Success 200 {object} common.Response{data=[]interface{},meta=common.ResponseMeta{pagination=common.PaginationMeta}}
// @Router /api/v1/catalog [get]
func (h *CatalogHandler) ListCatalogItems(c *gin.Context) {
	// Extract query options (includes deleted items if user is admin and include_deleted=true)
	middleware.ExtractQueryOptions(c)

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

	// Parse filter parameters
	filter := &catalogRequests.CatalogFilter{}

	// Item type filter
	if itemType := c.Query("item_type"); itemType != "" {
		itemTypeEnum := catalogModels.CatalogItemType(strings.ToUpper(itemType))
		filter.ItemType = &itemTypeEnum
	}

	// Category filters
	if category := c.Query("category"); category != "" {
		filter.Category = &category
	}
	if subcategory := c.Query("subcategory"); subcategory != "" {
		filter.Subcategory = &subcategory
	}

	// Active status filter
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			filter.IsActive = &isActive
		}
	}

	// Visibility filter
	if visibility := c.Query("visibility"); visibility != "" {
		visibilityEnum := catalogModels.VisibilityType(strings.ToUpper(visibility))
		filter.Visibility = &visibilityEnum
	}

	// Price range filters
	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			minPriceDecimal := decimal.NewFromFloat(minPrice)
			filter.MinPrice = &minPriceDecimal
		}
	}
	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			maxPriceDecimal := decimal.NewFromFloat(maxPrice)
			filter.MaxPrice = &maxPriceDecimal
		}
	}

	// Search filter
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	// Tags filter
	if tagsStr := c.Query("tags"); tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		filter.Tags = tags
	}

	// Organization filter (default to context org if present and not specified)
	if orgID := c.Query("organization_id"); orgID != "" {
		filter.OrganizationID = &orgID
	} else if ctxOrg, ok := middleware.GetOrgID(c); ok {
		filter.OrganizationID = &ctxOrg
	}

	// Get catalog items
	items, total, err := h.catalogService.ListCatalogItems(c.Request.Context(), filter, offset, limit)
	if err != nil {
		common.InternalServerError(c, "LIST_FAILED", "Failed to list catalog items", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, items, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
		Pagination: &common.PaginationMeta{
			Page:    page,
			Limit:   limit,
			Total:   total,
			HasNext: len(items) == limit,
		},
	})
}

// ListCatalogItemsByType godoc
// @Summary List catalog items by type
// @Description Retrieve a list of catalog items filtered by type with additional filtering and pagination
// @Tags catalog
// @Accept json
// @Produce json
// @Param type path string true "Item type (products, services, labour)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param category query string false "Filter by category"
// @Param subcategory query string false "Filter by subcategory"
// @Param is_active query bool false "Filter by active status"
// @Param visibility query string false "Filter by visibility (PRIVATE, ORG, NETWORK, PUBLIC)"
// @Param min_price query number false "Minimum price filter"
// @Param max_price query number false "Maximum price filter"
// @Param search query string false "Search term"
// @Param tags query []string false "Filter by tags"
// @Param organization_id query string false "Filter by organization ID"
// @Success 200 {object} common.Response{data=[]interface{},meta=common.ResponseMeta{pagination=common.PaginationMeta}}
// @Router /api/v1/catalog/{type} [get]
func (h *CatalogHandler) ListCatalogItemsByType(c *gin.Context) {
	// Get type from path parameter
	typeParam := c.Param("type")
	if typeParam == "" {
		common.BadRequest(c, "MISSING_TYPE", "Catalog type is required", nil)
		return
	}

	// Convert type parameter to CatalogItemType
	var itemType catalogModels.CatalogItemType
	switch strings.ToLower(typeParam) {
	case "products":
		itemType = catalogModels.CatalogItemTypeProduct
	case "services":
		itemType = catalogModels.CatalogItemTypeService
	case "labour":
		itemType = catalogModels.CatalogItemTypeLabour
	default:
		common.BadRequest(c, "INVALID_TYPE", "Invalid catalog type. Must be 'products', 'services', or 'labour'", nil)
		return
	}

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

	// Parse filter parameters
	filter := &catalogRequests.CatalogFilter{
		ItemType: &itemType, // Set the item type filter
	}

	// Category filters
	if category := c.Query("category"); category != "" {
		filter.Category = &category
	}
	if subcategory := c.Query("subcategory"); subcategory != "" {
		filter.Subcategory = &subcategory
	}

	// Active status filter
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			filter.IsActive = &isActive
		}
	}

	// Visibility filter
	if visibility := c.Query("visibility"); visibility != "" {
		visibilityEnum := catalogModels.VisibilityType(strings.ToUpper(visibility))
		filter.Visibility = &visibilityEnum
	}

	// Price range filters
	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			minPriceDecimal := decimal.NewFromFloat(minPrice)
			filter.MinPrice = &minPriceDecimal
		}
	}
	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			maxPriceDecimal := decimal.NewFromFloat(maxPrice)
			filter.MaxPrice = &maxPriceDecimal
		}
	}

	// Search filter
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	// Tags filter
	if tagsStr := c.Query("tags"); tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		filter.Tags = tags
	}

	// Organization filter (default to context org if present and not specified)
	if orgID := c.Query("organization_id"); orgID != "" {
		filter.OrganizationID = &orgID
	} else if ctxOrg, ok := middleware.GetOrgID(c); ok {
		filter.OrganizationID = &ctxOrg
	}

	// Get catalog items
	items, total, err := h.catalogService.ListCatalogItems(c.Request.Context(), filter, offset, limit)
	if err != nil {
		common.InternalServerError(c, "LIST_FAILED", "Failed to list catalog items", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, items, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
		Pagination: &common.PaginationMeta{
			Page:    page,
			Limit:   limit,
			Total:   total,
			HasNext: len(items) == limit,
		},
	})
}

// UpdateCatalogItemByTypeAndID godoc
// @Summary Update catalog item by type and ID
// @Description Update a catalog item by its type and ID
// @Tags catalog
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param type path string true "Item type (products, services, labour)"
// @Param id path string true "Item ID"
// @Param item body object true "Item updates"
// @Success 200 {object} common.Response{data=map[string]interface{}}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/{type}/{id} [put]
func (h *CatalogHandler) UpdateCatalogItemByTypeAndID(c *gin.Context) {
	// Get type and ID from path parameters
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

	var req catalogRequests.UpdateCatalogItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Route to appropriate handler based on type
	switch strings.ToLower(typeParam) {
	case "products":
		h.updateProduct(c, id, &req, userID.(string))
	case "services":
		h.updateService(c, id, &req, userID.(string))
	case "labour":
		h.updateLabour(c, id, &req, userID.(string))
	default:
		common.BadRequest(c, "INVALID_TYPE", "Invalid catalog type. Must be 'products', 'services', or 'labour'", nil)
		return
	}
}

// SearchCatalog godoc
// @Summary Search catalog items
// @Description Search catalog items with advanced filtering
// @Tags catalog
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param item_type query string false "Filter by item type (PRODUCT, SERVICE, LABOUR)"
// @Param category query string false "Filter by category"
// @Param subcategory query string false "Filter by subcategory"
// @Param is_active query bool false "Filter by active status"
// @Param visibility query string false "Filter by visibility (PRIVATE, ORG, NETWORK, PUBLIC)"
// @Param min_price query number false "Minimum price filter"
// @Param max_price query number false "Maximum price filter"
// @Param tags query []string false "Filter by tags"
// @Param organization_id query string false "Filter by organization ID"
// @Success 200 {object} common.Response{data=[]interface{},meta=common.ResponseMeta{pagination=common.PaginationMeta}}
// @Router /api/v1/catalog/search [get]
func (h *CatalogHandler) SearchCatalog(c *gin.Context) {
	// Get search query
	query := c.Query("q")
	if query == "" {
		common.BadRequest(c, "MISSING_QUERY", "Search query is required", nil)
		return
	}

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

	// Parse filter parameters (same as ListCatalogItems)
	filter := &catalogRequests.CatalogFilter{}

	// Item type filter
	if itemType := c.Query("item_type"); itemType != "" {
		itemTypeEnum := catalogModels.CatalogItemType(strings.ToUpper(itemType))
		filter.ItemType = &itemTypeEnum
	}

	// Category filters
	if category := c.Query("category"); category != "" {
		filter.Category = &category
	}
	if subcategory := c.Query("subcategory"); subcategory != "" {
		filter.Subcategory = &subcategory
	}

	// Active status filter
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			filter.IsActive = &isActive
		}
	}

	// Visibility filter
	if visibility := c.Query("visibility"); visibility != "" {
		visibilityEnum := catalogModels.VisibilityType(strings.ToUpper(visibility))
		filter.Visibility = &visibilityEnum
	}

	// Price range filters
	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			minPriceDecimal := decimal.NewFromFloat(minPrice)
			filter.MinPrice = &minPriceDecimal
		}
	}
	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			maxPriceDecimal := decimal.NewFromFloat(maxPrice)
			filter.MaxPrice = &maxPriceDecimal
		}
	}

	// Tags filter
	if tagsStr := c.Query("tags"); tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		filter.Tags = tags
	}

	// Organization filter (default to context org if present and not specified)
	if orgID := c.Query("organization_id"); orgID != "" {
		filter.OrganizationID = &orgID
	} else if ctxOrg, ok := middleware.GetOrgID(c); ok {
		filter.OrganizationID = &ctxOrg
	}

	// Search catalog items
	items, total, err := h.catalogService.SearchCatalog(c.Request.Context(), query, filter, offset, limit)
	if err != nil {
		common.InternalServerError(c, "SEARCH_FAILED", "Failed to search catalog", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, items, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
		Pagination: &common.PaginationMeta{
			Page:    page,
			Limit:   limit,
			Total:   total,
			HasNext: len(items) == limit,
		},
	})
}

// Helper methods for updating different catalog item types

func (h *CatalogHandler) updateProduct(c *gin.Context, id string, req *catalogRequests.UpdateCatalogItemRequest, userID string) {
	// Get existing product
	product, err := h.catalogService.GetProductByID(c.Request.Context(), id)
	if err != nil {
		common.NotFound(c, "PRODUCT_NOT_FOUND", "Product not found", map[string]interface{}{
			"product_id": id,
		})
		return
	}

	// Update fields if provided
	productHandler := &ProductHandler{catalogService: h.catalogService}
	productHandler.updateProductFromRequest(product, req)

	// Update the product
	updatedProduct, err := h.catalogService.UpdateProduct(c.Request.Context(), product, userID)
	if err != nil {
		common.InternalServerError(c, "UPDATE_FAILED", "Failed to update product", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, updatedProduct, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

func (h *CatalogHandler) updateService(c *gin.Context, id string, req *catalogRequests.UpdateCatalogItemRequest, userID string) {
	// Get existing service
	service, err := h.catalogService.GetServiceByID(c.Request.Context(), id)
	if err != nil {
		common.NotFound(c, "SERVICE_NOT_FOUND", "Service not found", map[string]interface{}{
			"service_id": id,
		})
		return
	}

	// Update fields if provided
	serviceHandler := &ServiceHandler{catalogService: h.catalogService}
	serviceHandler.updateServiceFromRequest(service, req)

	// Update the service
	updatedService, err := h.catalogService.UpdateService(c.Request.Context(), service, userID)
	if err != nil {
		common.InternalServerError(c, "UPDATE_FAILED", "Failed to update service", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, updatedService, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

func (h *CatalogHandler) updateLabour(c *gin.Context, id string, req *catalogRequests.UpdateCatalogItemRequest, userID string) {
	// Get existing labour
	labour, err := h.catalogService.GetLabourByID(c.Request.Context(), id)
	if err != nil {
		common.NotFound(c, "LABOUR_NOT_FOUND", "Labour not found", map[string]interface{}{
			"labour_id": id,
		})
		return
	}

	// Update fields if provided
	labourHandler := &LabourHandler{catalogService: h.catalogService}
	labourHandler.updateLabourFromRequest(labour, req)

	// Update the labour
	updatedLabour, err := h.catalogService.UpdateLabour(c.Request.Context(), labour, userID)
	if err != nil {
		common.InternalServerError(c, "UPDATE_FAILED", "Failed to update labour", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, updatedLabour, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// GetCatalogItemByTypeAndID godoc
// @Summary Get catalog item by type and ID
// @Description Retrieve a specific catalog item by its type and ID
// @Tags catalog
// @Accept json
// @Produce json
// @Param type path string true "Item type (products, services, labour, contracts)"
// @Param id path string true "Item ID"
// @Param If-None-Match header string false "ETag for conditional requests"
// @Success 200 {object} common.Response{data=map[string]interface{}}
// @Success 304 "Not modified"
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/{type}/{id} [get]
func (h *CatalogHandler) GetCatalogItemByTypeAndID(c *gin.Context) {
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

	// Validate conditional request headers if ETag service is available
	if h.etagService != nil {
		if err := h.etagService.ValidateConditionalRequest(c); err != nil {
			common.BadRequest(c, "INVALID_CONDITIONAL_HEADER", "Invalid conditional request header", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
	}

	// Get catalog item
	item, err := h.catalogService.GetCatalogItemByID(c.Request.Context(), id)
	if err != nil {
		common.NotFound(c, "ITEM_NOT_FOUND", "Catalog item not found", map[string]interface{}{
			"item_id": id,
		})
		return
	}

	// Validate type matches
	expectedType := ""
	switch strings.ToLower(typeParam) {
	case "products":
		expectedType = "PRODUCT"
	case "services":
		expectedType = "SERVICE"
	case "labour":
		expectedType = "LABOUR"
	case "contracts":
		expectedType = "CONTRACT"
	default:
		common.BadRequest(c, "INVALID_TYPE", "Invalid catalog type", nil)
		return
	}

	if string(item.ItemType) != expectedType {
		common.BadRequest(c, "TYPE_MISMATCH", "Item type does not match requested type", map[string]interface{}{
			"requested_type": typeParam,
			"actual_type":    item.ItemType,
		})
		return
	}

	// Handle ETag validation and conditional response
	if h.etagService != nil {
		// Log ETag operation
		h.etagService.LogETagOperation(c, "get_catalog_item", item, map[string]interface{}{
			"type_param": typeParam,
		})

		// Check if client has current version (ETag match)
		if h.etagService.HandleConditionalRequest(c, item) {
			// 304 Not Modified response was sent
			return
		}
	}

	common.Success(c, item, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// DeleteCatalogItemByTypeAndID godoc
// @Summary Delete catalog item by type and ID
// @Description Soft delete a catalog item by its type and ID
// @Tags catalog
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param type path string true "Item type (products, services, labour, contracts)"
// @Param id path string true "Item ID"
// @Param force query bool false "Force delete even with references (admin only)"
// @Success 204 "Catalog item deleted successfully"
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Failure 409 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/{type}/{id} [delete]
func (h *CatalogHandler) DeleteCatalogItemByTypeAndID(c *gin.Context) {
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

	// Route to appropriate handler based on type
	switch strings.ToLower(typeParam) {
	case "products":
		h.deleteProduct(c, id, userID.(string))
	case "services":
		h.deleteService(c, id, userID.(string))
	case "labour":
		h.deleteLabour(c, id, userID.(string))
	case "contracts":
		h.deleteContract(c, id, userID.(string))
	default:
		common.BadRequest(c, "INVALID_TYPE", "Invalid catalog type. Must be 'products', 'services', 'labour', or 'contracts'", nil)
		return
	}
}

// Helper methods for deleting different catalog item types

func (h *CatalogHandler) deleteProduct(c *gin.Context, id string, userID string) {
	err := h.catalogService.DeleteProduct(c.Request.Context(), id, userID)
	if err != nil {
		common.InternalServerError(c, "DELETE_FAILED", "Failed to delete product", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	c.Status(204)
}

func (h *CatalogHandler) deleteService(c *gin.Context, id string, userID string) {
	err := h.catalogService.DeleteService(c.Request.Context(), id, userID)
	if err != nil {
		common.InternalServerError(c, "DELETE_FAILED", "Failed to delete service", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	c.Status(204)
}

func (h *CatalogHandler) deleteLabour(c *gin.Context, id string, userID string) {
	err := h.catalogService.DeleteLabour(c.Request.Context(), id, userID)
	if err != nil {
		common.InternalServerError(c, "DELETE_FAILED", "Failed to delete labour", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	c.Status(204)
}

func (h *CatalogHandler) deleteContract(c *gin.Context, id string, userID string) {
	err := h.catalogService.DeleteContract(c.Request.Context(), id, userID)
	if err != nil {
		common.InternalServerError(c, "DELETE_FAILED", "Failed to delete contract", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	c.Status(204)
}
