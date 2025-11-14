package catalog

import (
	"encoding/json"
	"strconv"

	catalogModels "github.com/Kisanlink/kisanlink-ecom/entities/models/catalog"
	catalogRequests "github.com/Kisanlink/kisanlink-ecom/entities/requests/catalog"
	"github.com/Kisanlink/kisanlink-ecom/internal/common"
	"github.com/Kisanlink/kisanlink-ecom/internal/middleware"
	catalogService "github.com/Kisanlink/kisanlink-ecom/internal/services/catalog"

	"github.com/gin-gonic/gin"
)

// ProductHandler handles HTTP requests for product operations
type ProductHandler struct {
	catalogService catalogService.CatalogServiceInterface
	etagService    *catalogService.ETagService
}

// NewProductHandler creates a new product handler
func NewProductHandler(catalogService catalogService.CatalogServiceInterface, etagService *catalogService.ETagService) *ProductHandler {
	return &ProductHandler{
		catalogService: catalogService,
		etagService:    etagService,
	}
}

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product in the catalog
// @Tags products
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param product body object true "Product information"
// @Success 201 {object} common.Response{data=map[string]interface{}}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req catalogRequests.CreateCatalogItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Validate that this is a product request
	if req.ItemType != catalogModels.CatalogItemTypeProduct {
		common.BadRequest(c, "INVALID_TYPE", "Request type must be 'PRODUCT'", nil)
		return
	}

	// Create product from request
	product := catalogModels.NewProduct(
		"", // OrganizationID will be set from context or request
		req.Name,
		req.BasePrice,
	)

	// Set org from context header/middleware if available
	if orgID, ok := middleware.GetOrgID(c); ok {
		product.OrganizationID = orgID
	}

	// Set additional fields from request
	product.Category = req.Category
	product.Subcategory = req.Subcategory
	product.Description = req.Description
	product.SKU = req.SKU
	product.UnitOfMeasure = req.UnitOfMeasure
	product.Currency = req.Currency
	if req.Visibility != "" {
		product.Visibility = req.Visibility
	}
	product.Tags = req.Tags
	if req.Attributes != nil {
		attributesJSON, err := json.Marshal(req.Attributes)
		if err != nil {
			common.BadRequest(c, "INVALID_ATTRIBUTES", "Invalid attributes format", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		product.Attributes = string(attributesJSON)
	}
	product.Images = req.Images

	// Create the product
	createdProduct, err := h.catalogService.CreateProduct(c.Request.Context(), product, userID.(string))
	if err != nil {
		common.InternalServerError(c, "CREATE_FAILED", "Failed to create product", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Created(c, createdProduct, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// GetProductByID godoc
// @Summary Get product by ID
// @Description Retrieve a product by its ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param If-None-Match header string false "ETag for conditional requests"
// @Success 200 {object} common.Response{data=map[string]interface{}}
// @Success 304 "Not modified"
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/products/{id} [get]
func (h *ProductHandler) GetProductByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Product ID is required", nil)
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

	product, err := h.catalogService.GetProductByID(c.Request.Context(), id)
	if err != nil {
		common.NotFound(c, "PRODUCT_NOT_FOUND", "Product not found", map[string]interface{}{
			"product_id": id,
			"error":      err.Error(),
		})
		return
	}

	// Handle ETag validation and conditional response
	if h.etagService != nil {
		// Convert Product to CatalogItem for ETag processing
		catalogItem := &product.CatalogItem

		// Log ETag operation
		h.etagService.LogETagOperation(c, "get_product", catalogItem, map[string]interface{}{
			"product_id": id,
		})

		// Check if client has current version (ETag match)
		if h.etagService.HandleConditionalRequest(c, catalogItem) {
			// 304 Not Modified response was sent
			return
		}
	}

	common.Success(c, product, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// UpdateProduct godoc
// @Summary Update a product
// @Description Update an existing product
// @Tags products
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Product ID"
// @Param product body object true "Product updates"
// @Success 200 {object} common.Response{data=map[string]interface{}}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/products/{id} [put]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Product ID is required", nil)
		return
	}

	var req catalogRequests.UpdateCatalogItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Get existing product
	product, err := h.catalogService.GetProductByID(c.Request.Context(), id)
	if err != nil {
		common.NotFound(c, "PRODUCT_NOT_FOUND", "Product not found", map[string]interface{}{
			"product_id": id,
		})
		return
	}

	// Update fields if provided
	h.updateProductFromRequest(product, &req)

	// Update the product
	updatedProduct, err := h.catalogService.UpdateProduct(c.Request.Context(), product, userID.(string))
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

// DeleteProduct godoc
// @Summary Delete a product
// @Description Delete a product from the catalog
// @Tags products
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Product ID"
// @Success 200 {object} common.Response{data=string}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Product ID is required", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Delete the product
	err := h.catalogService.DeleteProduct(c.Request.Context(), id, userID.(string))
	if err != nil {
		common.InternalServerError(c, "DELETE_FAILED", "Failed to delete product", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, "Product deleted successfully", &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// ListProducts godoc
// @Summary List catalog products
// @Description List products with filtering and pagination. For admin users, returns all products. For FPO users, returns only products published to their organization with FPO-specific pricing (base price + delivery cost + commission).
// @Tags catalog-products
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" example("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Param category query string false "Filter by category ID" example("CAT00000001")
// @Param subcategory query string false "Filter by subcategory ID" example("SUBCAT00000001")
// @Param search query string false "Search in product name and description" example("organic fertilizer")
// @Param is_active query bool false "Filter by active status (admin only)" example(true)
// @Param include_deleted query bool false "Include soft-deleted items (admin only)" default(false)
// @Success 200 {object} common.Response{data=[]interface{},meta=common.ResponseMeta{pagination=common.PaginationMeta}} "Products retrieved successfully"
// @Failure 401 {object} common.Response{error=common.ResponseError} "Unauthorized - missing or invalid token"
// @Failure 500 {object} common.Response{error=common.ResponseError} "Internal server error"
// @Router /api/v1/catalog/products [get]
// @Security BearerAuth
func (h *ProductHandler) ListProducts(c *gin.Context) {
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

	// Check if user is admin
	isAdmin := common.IsAdmin(c)

	// Get organization ID from context
	orgID, hasOrgID := common.GetOrganizationID(c)

	// If user is not admin and has org ID, show FPO-filtered products
	if !isAdmin && hasOrgID {
		// Build filter from query params
		filter := &catalogRequests.CatalogFilter{}

		// Category filter
		if category := c.Query("category"); category != "" {
			filter.Category = &category
		}

		// Subcategory filter
		if subcategory := c.Query("subcategory"); subcategory != "" {
			filter.Subcategory = &subcategory
		}

		// Search filter
		if search := c.Query("search"); search != "" {
			filter.Search = &search
		}

		// Get FPO-filtered products with pricing
		productsWithPricing, err := h.catalogService.ListProductsForFPO(c.Request.Context(), orgID, filter, offset, limit)
		if err != nil {
			common.InternalServerError(c, "LIST_FAILED", "Failed to list products for FPO", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// Calculate total (simplified - in production use proper count query)
		total := len(productsWithPricing)

		common.Success(c, productsWithPricing, &common.ResponseMeta{
			TraceID: common.GetTraceID(c),
			Pagination: &common.PaginationMeta{
				Page:    page,
				Limit:   limit,
				Total:   total,
				HasNext: len(productsWithPricing) == limit,
			},
		})
		return
	}

	// Admin or no org ID - show all products (existing behavior)
	category := c.Query("category")
	status := c.Query("status")

	// Get products
	products, err := h.catalogService.ListProducts(c.Request.Context(), limit, offset, category, status)
	if err != nil {
		common.InternalServerError(c, "LIST_FAILED", "Failed to list products", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Convert to products (since ListProducts returns CatalogItems)
	productList := make([]*catalogModels.Product, 0, len(products))
	for _, item := range products {
		if item.ItemType == catalogModels.CatalogItemTypeProduct {
			product := &catalogModels.Product{
				CatalogItem: *item,
			}
			productList = append(productList, product)
		}
	}

	common.Success(c, productList, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
		Pagination: &common.PaginationMeta{
			Page:    page,
			Limit:   limit,
			Total:   len(productList),
			HasNext: len(productList) == limit,
		},
	})
}

// ActivateProduct godoc
// @Summary Activate a catalog product
// @Description Activate an inactive product to make it available for publishing and ordering. Only admins can activate products. Products are created as inactive by default and must be activated before they can be published to FPOs.
// @Tags catalog-products
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token (Admin only)" example("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
// @Param id path string true "Product ID" example("PROD00000001")
// @Success 200 {object} common.Response{data=string} "Product activated successfully"
// @Failure 400 {object} common.Response{error=common.ResponseError} "Activation failed"
// @Failure 401 {object} common.Response{error=common.ResponseError} "Unauthorized - missing or invalid token"
// @Failure 403 {object} common.Response{error=common.ResponseError} "Forbidden - admin access required"
// @Failure 404 {object} common.Response{error=common.ResponseError} "Product not found"
// @Failure 500 {object} common.Response{error=common.ResponseError} "Internal server error"
// @Router /api/v1/catalog/products/{id}/activate [patch]
// @Security BearerAuth
func (h *ProductHandler) ActivateProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Product ID is required", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Check if user is admin (you can implement proper RBAC here)
	// For now, we'll allow any authenticated user
	// TODO: Add proper admin check using RBAC
	// if !common.IsAdmin(c) {
	// 	common.Forbidden(c, "INSUFFICIENT_PERMISSIONS", "Only admins can activate products", nil)
	// 	return
	// }

	err := h.catalogService.UpdateActiveStatus(c.Request.Context(), id, true, userID.(string))
	if err != nil {
		common.BadRequest(c, "ACTIVATION_FAILED", "Failed to activate product", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, "Product activated successfully", &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// DeactivateProduct godoc
// @Summary Deactivate a catalog product
// @Description Deactivate an active product to make it unavailable for new orders and inventory creation. Existing orders are not affected. Only admins can deactivate products.
// @Tags catalog-products
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token (Admin only)" example("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
// @Param id path string true "Product ID" example("PROD00000001")
// @Success 200 {object} common.Response{data=string} "Product deactivated successfully"
// @Failure 400 {object} common.Response{error=common.ResponseError} "Deactivation failed"
// @Failure 401 {object} common.Response{error=common.ResponseError} "Unauthorized - missing or invalid token"
// @Failure 403 {object} common.Response{error=common.ResponseError} "Forbidden - admin access required"
// @Failure 404 {object} common.Response{error=common.ResponseError} "Product not found"
// @Failure 500 {object} common.Response{error=common.ResponseError} "Internal server error"
// @Router /api/v1/catalog/products/{id}/deactivate [patch]
// @Security BearerAuth
func (h *ProductHandler) DeactivateProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Product ID is required", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Check if user is admin (you can implement proper RBAC here)
	// For now, we'll allow any authenticated user
	// TODO: Add proper admin check using RBAC
	// if !common.IsAdmin(c) {
	// 	common.Forbidden(c, "INSUFFICIENT_PERMISSIONS", "Only admins can deactivate products", nil)
	// 	return
	// }

	err := h.catalogService.UpdateActiveStatus(c.Request.Context(), id, false, userID.(string))
	if err != nil {
		common.BadRequest(c, "DEACTIVATION_FAILED", "Failed to deactivate product", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, "Product deactivated successfully", &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// updateProductFromRequest updates product fields from request
func (h *ProductHandler) updateProductFromRequest(product *catalogModels.Product, req *catalogRequests.UpdateCatalogItemRequest) {
	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Category != nil {
		product.Category = *req.Category
	}
	if req.Subcategory != nil {
		product.Subcategory = *req.Subcategory
	}
	if req.UnitOfMeasure != nil {
		product.UnitOfMeasure = *req.UnitOfMeasure
	}
	if req.BasePrice != nil {
		product.BasePrice = *req.BasePrice
	}
	if req.Currency != nil {
		product.Currency = *req.Currency
	}
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}
	if req.Visibility != nil {
		product.Visibility = *req.Visibility
	}
	if req.Tags != nil {
		product.Tags = req.Tags
	}
	if req.Attributes != nil {
		attributesJSON, err := json.Marshal(req.Attributes)
		if err == nil {
			product.Attributes = string(attributesJSON)
		}
	}
	if req.Images != nil {
		product.Images = req.Images
	}
	if req.SKU != nil {
		product.SKU = *req.SKU
	}
}
