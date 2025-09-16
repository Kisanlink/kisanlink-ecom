package catalog

import (
	"encoding/json"
	"strconv"

	catalogModels "kisanlink-ecom/entities/models/catalog"
	catalogRequests "kisanlink-ecom/entities/requests/catalog"
	_ "kisanlink-ecom/entities/responses/catalog" // For Swagger documentation
	"kisanlink-ecom/internal/common"
	"kisanlink-ecom/internal/middleware"
	catalogService "kisanlink-ecom/internal/services/catalog"

	"github.com/gin-gonic/gin"
)

// ProductHandler handles HTTP requests for product operations
type ProductHandler struct {
	catalogService catalogService.CatalogServiceInterface
}

// NewProductHandler creates a new product handler
func NewProductHandler(catalogService catalogService.CatalogServiceInterface) *ProductHandler {
	return &ProductHandler{
		catalogService: catalogService,
	}
}

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product in the catalog
// @Tags products
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param product body catalog.CreateCatalogItemRequest true "Product information"
// @Success 201 {object} common.Response{data=catalog.ProductResponse}
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
// @Success 200 {object} common.Response{data=catalog.ProductResponse}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/products/{id} [get]
func (h *ProductHandler) GetProductByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Product ID is required", nil)
		return
	}

	product, err := h.catalogService.GetProductByID(c.Request.Context(), id)
	if err != nil {
		common.NotFound(c, "PRODUCT_NOT_FOUND", "Product not found", map[string]interface{}{
			"product_id": id,
			"error":      err.Error(),
		})
		return
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
// @Param product body catalog.UpdateCatalogItemRequest true "Product updates"
// @Success 200 {object} common.Response{data=catalog.ProductResponse}
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
// @Summary List products
// @Description Retrieve a list of products with filtering and pagination
// @Tags products
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param category query string false "Filter by category"
// @Param org_id query string false "Filter by organization ID"
// @Param is_active query bool false "Filter by active status"
// @Param search query string false "Search term"
// @Success 200 {object} common.Response{data=[]catalog.ProductResponse,meta=common.ResponseMeta{pagination=common.PaginationMeta}}
// @Router /api/v1/catalog/products [get]
func (h *ProductHandler) ListProducts(c *gin.Context) {
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
