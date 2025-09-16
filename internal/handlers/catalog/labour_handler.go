package catalog

import (
	"encoding/json"
	catalogModels "kisanlink-ecom/entities/models/catalog"
	catalogRequests "kisanlink-ecom/entities/requests/catalog"
	"kisanlink-ecom/internal/common"
	catalogService "kisanlink-ecom/internal/services/catalog"
	"strconv"

	"github.com/gin-gonic/gin"
)

// LabourHandler handles HTTP requests for labour operations
type LabourHandler struct {
	catalogService catalogService.CatalogServiceInterface
}

// NewLabourHandler creates a new labour handler
func NewLabourHandler(catalogService catalogService.CatalogServiceInterface) *LabourHandler {
	return &LabourHandler{
		catalogService: catalogService,
	}
}

// CreateLabour godoc
// @Summary Create a new labour offering
// @Description Create a new labour offering in the catalog
// @Tags labour
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param labour body catalog.CreateLabourRequest true "Labour information"
// @Success 201 {object} common.Response{data=catalog.LabourResponse}
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
// @Success 200 {object} common.Response{data=catalog.LabourResponse}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/labour/{id} [get]
func (h *LabourHandler) GetLabourByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Labour ID is required", nil)
		return
	}

	labour, err := h.catalogService.GetLabourByID(c.Request.Context(), id)
	if err != nil {
		common.NotFound(c, "LABOUR_NOT_FOUND", "Labour not found", map[string]interface{}{
			"labour_id": id,
			"error":     err.Error(),
		})
		return
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
// @Param labour body catalog.UpdateCatalogItemRequest true "Labour updates"
// @Success 200 {object} common.Response{data=catalog.LabourResponse}
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
// @Success 200 {object} common.Response{data=[]catalog.LabourResponse,meta=common.ResponseMeta{pagination=common.PaginationMeta}}
// @Router /api/v1/catalog/labour [get]
func (h *LabourHandler) ListLabour(c *gin.Context) {
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
