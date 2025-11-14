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

// LabourHandler handles HTTP requests for labour operations
type LabourHandler struct {
	catalogService catalogService.CatalogServiceInterface
	etagService    *catalogService.ETagService
}

// NewLabourHandler creates a new labour handler
func NewLabourHandler(catalogService catalogService.CatalogServiceInterface, etagService *catalogService.ETagService) *LabourHandler {
	return &LabourHandler{
		catalogService: catalogService,
		etagService:    etagService,
	}
}

// CreateLabour godoc
// @Summary Create a new labour offering
// @Description Create a new labour offering in the catalog
// @Tags labour
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param labour body object true "Labour information"
// @Success 201 {object} common.Response{data=map[string]interface{}}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/labour [post]
func (h *LabourHandler) CreateLabour(c *gin.Context) {
	var req catalogRequests.CreateLabourRequest
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

	// Validate that this is a labour request
	if req.ItemType != catalogModels.CatalogItemTypeLabour {
		common.BadRequest(c, "INVALID_TYPE", "Request type must be 'LABOUR'", nil)
		return
	}

	// Create the labour
	createdLabour, err := h.catalogService.CreateLabour(c.Request.Context(), &req, userID.(string))
	if err != nil {
		common.InternalServerError(c, "CREATE_FAILED", "Failed to create labour", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Created(c, createdLabour, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// GetLabourByID godoc
// @Summary Get labour by ID
// @Description Retrieve a labour offering by its ID
// @Tags labour
// @Accept json
// @Produce json
// @Param id path string true "Labour ID"
// @Param If-None-Match header string false "ETag for conditional requests"
// @Success 200 {object} common.Response{data=map[string]interface{}}
// @Success 304 "Not modified"
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/labour/{id} [get]
func (h *LabourHandler) GetLabourByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Labour ID is required", nil)
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

	labour, err := h.catalogService.GetLabourByID(c.Request.Context(), id)
	if err != nil {
		common.NotFound(c, "LABOUR_NOT_FOUND", "Labour not found", map[string]interface{}{
			"labour_id": id,
			"error":     err.Error(),
		})
		return
	}

	// Handle ETag validation and conditional response
	if h.etagService != nil {
		// Convert Labour to CatalogItem for ETag processing
		catalogItem := &labour.CatalogItem

		// Log ETag operation
		h.etagService.LogETagOperation(c, "get_labour", catalogItem, map[string]interface{}{
			"labour_id": id,
		})

		// Check if client has current version (ETag match)
		if h.etagService.HandleConditionalRequest(c, catalogItem) {
			// 304 Not Modified response was sent
			return
		}
	}

	common.Success(c, labour, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// UpdateLabour godoc
// @Summary Update a labour offering
// @Description Update an existing labour offering
// @Tags labour
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Labour ID"
// @Param labour body object true "Labour updates"
// @Success 200 {object} common.Response{data=map[string]interface{}}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/labour/{id} [put]
func (h *LabourHandler) UpdateLabour(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Labour ID is required", nil)
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

	// Get existing labour
	labour, err := h.catalogService.GetLabourByID(c.Request.Context(), id)
	if err != nil {
		common.NotFound(c, "LABOUR_NOT_FOUND", "Labour not found", map[string]interface{}{
			"labour_id": id,
		})
		return
	}

	// Update fields if provided
	h.updateLabourFromRequest(labour, &req)

	// Update the labour
	updatedLabour, err := h.catalogService.UpdateLabour(c.Request.Context(), labour, userID.(string))
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

// DeleteLabour godoc
// @Summary Delete a labour offering
// @Description Delete a labour offering from the catalog
// @Tags labour
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Labour ID"
// @Success 200 {object} common.Response{data=string}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/labour/{id} [delete]
func (h *LabourHandler) DeleteLabour(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Labour ID is required", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Delete the labour
	err := h.catalogService.DeleteLabour(c.Request.Context(), id, userID.(string))
	if err != nil {
		common.InternalServerError(c, "DELETE_FAILED", "Failed to delete labour", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, "Labour deleted successfully", &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// ListLabour godoc
// @Summary List labour offerings
// @Description Retrieve a list of labour offerings with filtering and pagination
// @Tags labour
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param category query string false "Filter by category"
// @Param org_id query string false "Filter by organization ID"
// @Param is_active query bool false "Filter by active status"
// @Param search query string false "Search term"
// @Param include_deleted query bool false "Include soft-deleted items (admin only)" default(false)
// @Success 200 {object} common.Response{data=[]interface{},meta=common.ResponseMeta{pagination=common.PaginationMeta}}
// @Router /api/v1/catalog/labour [get]
func (h *LabourHandler) ListLabour(c *gin.Context) {
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
	category := c.Query("category")
	status := c.Query("status")

	// Get labour
	labour, err := h.catalogService.ListLabour(c.Request.Context(), limit, offset, category, status)
	if err != nil {
		common.InternalServerError(c, "LIST_FAILED", "Failed to list labour", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Convert to labour (since ListLabour returns CatalogItems)
	labourList := make([]*catalogModels.Labour, 0, len(labour))
	for _, item := range labour {
		if item.ItemType == catalogModels.CatalogItemTypeLabour {
			labourItem := &catalogModels.Labour{
				CatalogItem: *item,
			}
			labourList = append(labourList, labourItem)
		}
	}

	common.Success(c, labourList, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
		Pagination: &common.PaginationMeta{
			Page:    page,
			Limit:   limit,
			Total:   len(labourList),
			HasNext: len(labourList) == limit,
		},
	})
}

// ActivateLabour godoc
// @Summary Activate a catalog labour offering
// @Description Activate an inactive labour offering to make it available for publishing and ordering. Only admins can activate labour offerings. Labour offerings are created as inactive by default and must be activated before they can be published.
// @Tags catalog-labour
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token (Admin only)" example("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
// @Param id path string true "Labour ID" example("LABR00000001")
// @Success 200 {object} common.Response{data=string} "Labour activated successfully"
// @Failure 400 {object} common.Response{error=common.ResponseError} "Activation failed"
// @Failure 401 {object} common.Response{error=common.ResponseError} "Unauthorized - missing or invalid token"
// @Failure 403 {object} common.Response{error=common.ResponseError} "Forbidden - admin access required"
// @Failure 404 {object} common.Response{error=common.ResponseError} "Labour not found"
// @Failure 500 {object} common.Response{error=common.ResponseError} "Internal server error"
// @Router /api/v1/catalog/labour/{id}/activate [patch]
// @Security BearerAuth
func (h *LabourHandler) ActivateLabour(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Labour ID is required", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	err := h.catalogService.UpdateActiveStatus(c.Request.Context(), id, true, userID.(string))
	if err != nil {
		common.BadRequest(c, "ACTIVATION_FAILED", "Failed to activate labour", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, "Labour activated successfully", &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// DeactivateLabour godoc
// @Summary Deactivate a catalog labour offering
// @Description Deactivate an active labour offering to make it unavailable for new orders. Existing orders are not affected. Only admins can deactivate labour offerings.
// @Tags catalog-labour
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token (Admin only)" example("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
// @Param id path string true "Labour ID" example("LABR00000001")
// @Success 200 {object} common.Response{data=string} "Labour deactivated successfully"
// @Failure 400 {object} common.Response{error=common.ResponseError} "Deactivation failed"
// @Failure 401 {object} common.Response{error=common.ResponseError} "Unauthorized - missing or invalid token"
// @Failure 403 {object} common.Response{error=common.ResponseError} "Forbidden - admin access required"
// @Failure 404 {object} common.Response{error=common.ResponseError} "Labour not found"
// @Failure 500 {object} common.Response{error=common.ResponseError} "Internal server error"
// @Router /api/v1/catalog/labour/{id}/deactivate [patch]
// @Security BearerAuth
func (h *LabourHandler) DeactivateLabour(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Labour ID is required", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	err := h.catalogService.UpdateActiveStatus(c.Request.Context(), id, false, userID.(string))
	if err != nil {
		common.BadRequest(c, "DEACTIVATION_FAILED", "Failed to deactivate labour", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, "Labour deactivated successfully", &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// updateLabourFromRequest updates labour fields from request
func (h *LabourHandler) updateLabourFromRequest(labour *catalogModels.Labour, req *catalogRequests.UpdateCatalogItemRequest) {
	if req.Name != nil {
		labour.Name = *req.Name
	}
	if req.Description != nil {
		labour.Description = *req.Description
	}
	if req.Category != nil {
		labour.Category = *req.Category
	}
	if req.Subcategory != nil {
		labour.Subcategory = *req.Subcategory
	}
	if req.UnitOfMeasure != nil {
		labour.UnitOfMeasure = *req.UnitOfMeasure
	}
	if req.BasePrice != nil {
		labour.BasePrice = *req.BasePrice
	}
	if req.Currency != nil {
		labour.Currency = *req.Currency
	}
	if req.IsActive != nil {
		labour.IsActive = *req.IsActive
	}
	if req.Visibility != nil {
		labour.Visibility = *req.Visibility
	}
	if req.Tags != nil {
		labour.Tags = req.Tags
	}
	if req.Attributes != nil {
		attributesJSON, err := json.Marshal(req.Attributes)
		if err == nil {
			labour.Attributes = string(attributesJSON)
		}
	}
	if req.Images != nil {
		labour.Images = req.Images
	}
	if req.SKU != nil {
		labour.SKU = *req.SKU
	}
}
