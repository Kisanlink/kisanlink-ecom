package catalog

import (
	"encoding/json"
	catalogModels "kisanlink-ecom/entities/models/catalog"
	catalogRequests "kisanlink-ecom/entities/requests/catalog"
	"kisanlink-ecom/internal/common"
	"kisanlink-ecom/internal/middleware"
	catalogService "kisanlink-ecom/internal/services/catalog"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ServiceHandler handles HTTP requests for service operations
type ServiceHandler struct {
	catalogService catalogService.CatalogServiceInterface
	etagService    *catalogService.ETagService
}

// NewServiceHandler creates a new service handler
func NewServiceHandler(catalogService catalogService.CatalogServiceInterface, etagService *catalogService.ETagService) *ServiceHandler {
	return &ServiceHandler{
		catalogService: catalogService,
		etagService:    etagService,
	}
}

// CreateService godoc
// @Summary Create a new service
// @Description Create a new service in the catalog
// @Tags services
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param service body catalog.CreateServiceRequest true "Service information"
// @Success 201 {object} common.Response{data=catalog.ServiceResponse}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/services [post]
func (h *ServiceHandler) CreateService(c *gin.Context) {
	var req catalogRequests.CreateServiceRequest
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

	// Validate that this is a service request
	if req.ItemType != catalogModels.CatalogItemTypeService {
		common.BadRequest(c, "INVALID_TYPE", "Request type must be 'SERVICE'", nil)
		return
	}

	// Create the service
	createdService, err := h.catalogService.CreateService(c.Request.Context(), &req, userID.(string))
	if err != nil {
		common.InternalServerError(c, "CREATE_FAILED", "Failed to create service", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Created(c, createdService, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// GetServiceByID godoc
// @Summary Get service by ID
// @Description Retrieve a service by its ID
// @Tags services
// @Accept json
// @Produce json
// @Param id path string true "Service ID"
// @Param If-None-Match header string false "ETag for conditional requests"
// @Success 200 {object} common.Response{data=catalog.ServiceResponse}
// @Success 304 "Not modified"
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/services/{id} [get]
func (h *ServiceHandler) GetServiceByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Service ID is required", nil)
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

	service, err := h.catalogService.GetServiceByID(c.Request.Context(), id)
	if err != nil {
		common.NotFound(c, "SERVICE_NOT_FOUND", "Service not found", map[string]interface{}{
			"service_id": id,
			"error":      err.Error(),
		})
		return
	}

	// Handle ETag validation and conditional response
	if h.etagService != nil {
		// Convert Service to CatalogItem for ETag processing
		catalogItem := &service.CatalogItem

		// Log ETag operation
		h.etagService.LogETagOperation(c, "get_service", catalogItem, map[string]interface{}{
			"service_id": id,
		})

		// Check if client has current version (ETag match)
		if h.etagService.HandleConditionalRequest(c, catalogItem) {
			// 304 Not Modified response was sent
			return
		}
	}

	common.Success(c, service, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// UpdateService godoc
// @Summary Update a service
// @Description Update an existing service
// @Tags services
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Service ID"
// @Param service body catalog.UpdateCatalogItemRequest true "Service updates"
// @Success 200 {object} common.Response{data=catalog.ServiceResponse}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/services/{id} [put]
func (h *ServiceHandler) UpdateService(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Service ID is required", nil)
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

	// Get existing service
	service, err := h.catalogService.GetServiceByID(c.Request.Context(), id)
	if err != nil {
		common.NotFound(c, "SERVICE_NOT_FOUND", "Service not found", map[string]interface{}{
			"service_id": id,
		})
		return
	}

	// Update fields if provided
	h.updateServiceFromRequest(service, &req)

	// Update the service
	updatedService, err := h.catalogService.UpdateService(c.Request.Context(), service, userID.(string))
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

// DeleteService godoc
// @Summary Delete a service
// @Description Delete a service from the catalog
// @Tags services
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Service ID"
// @Success 200 {object} common.Response{data=string}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/services/{id} [delete]
func (h *ServiceHandler) DeleteService(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Service ID is required", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Delete the service
	err := h.catalogService.DeleteService(c.Request.Context(), id, userID.(string))
	if err != nil {
		common.InternalServerError(c, "DELETE_FAILED", "Failed to delete service", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, "Service deleted successfully", &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// ListServices godoc
// @Summary List services
// @Description Retrieve a list of services with filtering and pagination
// @Tags services
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param category query string false "Filter by category"
// @Param org_id query string false "Filter by organization ID"
// @Param is_active query bool false "Filter by active status"
// @Param search query string false "Search term"
// @Param include_deleted query bool false "Include soft-deleted items (admin only)" default(false)
// @Success 200 {object} common.Response{data=[]catalog.ServiceResponse,meta=common.ResponseMeta{pagination=common.PaginationMeta}}
// @Router /api/v1/catalog/services [get]
func (h *ServiceHandler) ListServices(c *gin.Context) {
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

	// Get services
	services, err := h.catalogService.ListServices(c.Request.Context(), limit, offset, category, status)
	if err != nil {
		common.InternalServerError(c, "LIST_FAILED", "Failed to list services", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Convert to services (since ListServices returns CatalogItems)
	serviceList := make([]*catalogModels.Service, 0, len(services))
	for _, item := range services {
		if item.ItemType == catalogModels.CatalogItemTypeService {
			service := &catalogModels.Service{
				CatalogItem: *item,
			}
			serviceList = append(serviceList, service)
		}
	}

	common.Success(c, serviceList, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
		Pagination: &common.PaginationMeta{
			Page:    page,
			Limit:   limit,
			Total:   len(serviceList),
			HasNext: len(serviceList) == limit,
		},
	})
}

// ActivateService godoc
// @Summary Activate a service
// @Description Activate a service to make it visible and usable (Admin only)
// @Tags services
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Service ID"
// @Success 200 {object} common.Response{data=string}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/services/{id}/activate [patch]
func (h *ServiceHandler) ActivateService(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Service ID is required", nil)
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
		common.BadRequest(c, "ACTIVATION_FAILED", "Failed to activate service", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, "Service activated successfully", &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// DeactivateService godoc
// @Summary Deactivate a service
// @Description Deactivate a service to make it invisible and unusable (Admin only)
// @Tags services
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Service ID"
// @Success 200 {object} common.Response{data=string}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/catalog/services/{id}/deactivate [patch]
func (h *ServiceHandler) DeactivateService(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Service ID is required", nil)
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
		common.BadRequest(c, "DEACTIVATION_FAILED", "Failed to deactivate service", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, "Service deactivated successfully", &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// updateServiceFromRequest updates service fields from request
func (h *ServiceHandler) updateServiceFromRequest(service *catalogModels.Service, req *catalogRequests.UpdateCatalogItemRequest) {
	if req.Name != nil {
		service.Name = *req.Name
	}
	if req.Description != nil {
		service.Description = *req.Description
	}
	if req.Category != nil {
		service.Category = *req.Category
	}
	if req.Subcategory != nil {
		service.Subcategory = *req.Subcategory
	}
	if req.UnitOfMeasure != nil {
		service.UnitOfMeasure = *req.UnitOfMeasure
	}
	if req.BasePrice != nil {
		service.BasePrice = *req.BasePrice
	}
	if req.Currency != nil {
		service.Currency = *req.Currency
	}
	if req.IsActive != nil {
		service.IsActive = *req.IsActive
	}
	if req.Visibility != nil {
		service.Visibility = *req.Visibility
	}
	if req.Tags != nil {
		service.Tags = req.Tags
	}
	if req.Attributes != nil {
		attributesJSON, err := json.Marshal(req.Attributes)
		if err == nil {
			service.Attributes = string(attributesJSON)
		}
	}
	if req.Images != nil {
		service.Images = req.Images
	}
	if req.SKU != nil {
		service.SKU = *req.SKU
	}
}
