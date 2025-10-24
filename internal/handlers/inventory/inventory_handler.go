package inventory

import (
	"net/http"
	"strconv"
	"time"

	catalogModels "kisanlink-ecom/entities/models/catalog"
	"kisanlink-ecom/internal/common"
	"kisanlink-ecom/internal/middleware"
	inventoryService "kisanlink-ecom/internal/services/inventory"
	"kisanlink-ecom/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// InventoryHandler handles HTTP requests for inventory management
type InventoryHandler struct {
	inventoryService inventoryService.InventoryService
}

// NewInventoryHandler creates a new inventory handler
func NewInventoryHandler(inventoryService inventoryService.InventoryService) *InventoryHandler {
	return &InventoryHandler{
		inventoryService: inventoryService,
	}
}

// CreateInventoryLotRequest represents the request body for creating an inventory lot
type CreateInventoryLotRequest struct {
	CatalogItemID     string  `json:"catalog_item_id" binding:"required" example:"prod_123"`
	LotNumber         string  `json:"lot_number" binding:"required" example:"LOT-2024-001"`
	BatchNumber       string  `json:"batch_number" example:"BATCH-001"`
	InitialQuantity   float64 `json:"initial_quantity" binding:"required,gt=0" example:"100.5"`
	QualityGrade      string  `json:"quality_grade" example:"A"`
	HarvestDate       string  `json:"harvest_date" example:"2024-01-15"`
	ExpiryDate        string  `json:"expiry_date" example:"2024-12-31"`
	WarehouseLocation string  `json:"warehouse_location" example:"Warehouse A, Section 1"`
	StorageConditions string  `json:"storage_conditions" example:"Temperature controlled, humidity 60%"`
	LotPrice          float64 `json:"lot_price" example:"25.50"`
	Metadata          string  `json:"metadata" example:"{\"supplier\": \"Farm ABC\"}"`
}

// UpdateInventoryLotRequest represents the request body for updating an inventory lot
type UpdateInventoryLotRequest struct {
	QualityGrade      *string  `json:"quality_grade" example:"A+"`
	HarvestDate       *string  `json:"harvest_date" example:"2024-01-15"`
	ExpiryDate        *string  `json:"expiry_date" example:"2024-12-31"`
	WarehouseLocation *string  `json:"warehouse_location" example:"Warehouse B, Section 2"`
	StorageConditions *string  `json:"storage_conditions" example:"Refrigerated, humidity 50%"`
	LotPrice          *float64 `json:"lot_price" example:"30.00"`
	Metadata          *string  `json:"metadata" example:"{\"notes\": \"Quality improved\"}"`
}

// AdjustInventoryRequest represents the request body for adjusting inventory quantity
type AdjustInventoryRequest struct {
	Adjustment decimal.Decimal `json:"adjustment" binding:"required" example:"10.5"`
	Reason     string          `json:"reason" binding:"required" example:"Stock correction after physical count"`
}

// InventoryAvailabilityResponse represents the response for inventory availability check
type InventoryAvailabilityResponse struct {
	CatalogItemID     string          `json:"catalog_item_id"`
	Available         bool            `json:"available"`
	AvailableQuantity decimal.Decimal `json:"available_quantity"`
	TotalQuantity     decimal.Decimal `json:"total_quantity"`
}

// CreateInventoryLot creates a new inventory lot
// @Summary Create inventory lot
// @Description Create a new inventory lot with product validation
// @Tags inventory
// @Accept json
// @Produce json
// @Param request body CreateInventoryLotRequest true "Inventory lot creation request"
// @Success 201 {object} common.Response
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 500 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/inventory/lots [post]
func (h *InventoryHandler) CreateInventoryLot(c *gin.Context) {
	var req CreateInventoryLotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	// Get user context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", "")
		return
	}

	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		utils.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Organization context required", "")
		return
	}

	// Convert request to service request
	serviceReq := &inventoryService.CreateInventoryLotRequest{
		CatalogItemID:     req.CatalogItemID,
		LotNumber:         req.LotNumber,
		BatchNumber:       req.BatchNumber,
		InitialQuantity:   decimal.NewFromFloat(req.InitialQuantity),
		QualityGrade:      req.QualityGrade,
		WarehouseLocation: req.WarehouseLocation,
		StorageConditions: req.StorageConditions,
		Metadata:          req.Metadata,
	}

	// Parse dates if provided
	if req.HarvestDate != "" {
		if harvestDate, err := time.Parse("2006-01-02", req.HarvestDate); err == nil {
			serviceReq.HarvestDate = &harvestDate
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid harvest date format. Use YYYY-MM-DD", "")
			return
		}
	}

	if req.ExpiryDate != "" {
		if expiryDate, err := time.Parse("2006-01-02", req.ExpiryDate); err == nil {
			serviceReq.ExpiryDate = &expiryDate
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid expiry date format. Use YYYY-MM-DD", "")
			return
		}
	}

	// Set lot price if provided
	if req.LotPrice > 0 {
		lotPrice := decimal.NewFromFloat(req.LotPrice)
		serviceReq.LotPrice = &lotPrice
	}

	// Create inventory lot
	lot, err := h.inventoryService.CreateInventoryLot(
		c.Request.Context(),
		serviceReq,
		userID,
		orgID,
	)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "SERVICE_ERROR", err.Error(), "")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Inventory lot created successfully", lot)
}

// UpdateInventoryLot updates an inventory lot
// @Summary Update inventory lot
// @Description Update inventory lot metadata (not quantities)
// @Tags inventory
// @Accept json
// @Produce json
// @Param id path string true "Inventory lot ID"
// @Param request body UpdateInventoryLotRequest true "Inventory lot update request"
// @Success 200 {object} common.Response
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Failure 500 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/inventory/lots/{id} [patch]
func (h *InventoryHandler) UpdateInventoryLot(c *gin.Context) {
	lotID := c.Param("id")
	if lotID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Lot ID is required", "")
		return
	}

	var req UpdateInventoryLotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	// Get user context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", "")
		return
	}

	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		utils.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Organization context required", "")
		return
	}

	// Convert request to service request
	serviceReq := &inventoryService.UpdateInventoryLotRequest{
		QualityGrade:      req.QualityGrade,
		WarehouseLocation: req.WarehouseLocation,
		StorageConditions: req.StorageConditions,
		Metadata:          req.Metadata,
	}

	// Parse dates if provided
	if req.HarvestDate != nil && *req.HarvestDate != "" {
		if harvestDate, err := time.Parse("2006-01-02", *req.HarvestDate); err == nil {
			serviceReq.HarvestDate = &harvestDate
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid harvest date format. Use YYYY-MM-DD", "")
			return
		}
	}

	if req.ExpiryDate != nil && *req.ExpiryDate != "" {
		if expiryDate, err := time.Parse("2006-01-02", *req.ExpiryDate); err == nil {
			serviceReq.ExpiryDate = &expiryDate
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid expiry date format. Use YYYY-MM-DD", "")
			return
		}
	}

	// Set lot price if provided
	if req.LotPrice != nil && *req.LotPrice > 0 {
		lotPrice := decimal.NewFromFloat(*req.LotPrice)
		serviceReq.LotPrice = &lotPrice
	}

	// Update inventory lot
	lot, err := h.inventoryService.UpdateInventoryLot(
		c.Request.Context(),
		lotID,
		serviceReq,
		userID,
		orgID,
	)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "SERVICE_ERROR", err.Error(), "")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Inventory lot updated successfully", lot)
}

// ListInventoryLots lists inventory lots with filtering and pagination
// @Summary List inventory lots
// @Description List inventory lots with organization filtering and pagination
// @Tags inventory
// @Accept json
// @Produce json
// @Param catalog_item_id query string false "Filter by catalog item ID"
// @Param status query string false "Filter by status (available, reserved, sold, expired, damaged)"
// @Param expiring_before query string false "Filter lots expiring before date (YYYY-MM-DD)"
// @Param quality_grade query string false "Filter by quality grade"
// @Param offset query int false "Pagination offset" default(0)
// @Param limit query int false "Pagination limit (max 100)" default(20)
// @Param include_deleted query bool false "Include soft-deleted items (admin only)" default(false)
// @Success 200 {object} common.Response
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 500 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/inventory/lots [get]
func (h *InventoryHandler) ListInventoryLots(c *gin.Context) {
	// Extract query options (includes deleted items if user is admin and include_deleted=true)
	middleware.ExtractQueryOptions(c)

	// Get user context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", "")
		return
	}

	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		utils.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Organization context required", "")
		return
	}

	// Parse query parameters
	filter := &inventoryService.InventoryFilter{
		CatalogItemID: c.Query("catalog_item_id"),
		QualityGrade:  c.Query("quality_grade"),
	}

	// Parse status
	if statusStr := c.Query("status"); statusStr != "" {
		filter.Status = catalogModels.LotStatus(statusStr)
	}

	// Parse expiring_before date
	if expiringStr := c.Query("expiring_before"); expiringStr != "" {
		if expiringDate, err := time.Parse("2006-01-02", expiringStr); err == nil {
			filter.ExpiringBefore = &expiringDate
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid expiring_before date format. Use YYYY-MM-DD", "")
			return
		}
	}

	// Parse pagination
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	// List inventory lots
	response, err := h.inventoryService.ListInventoryLots(
		c.Request.Context(),
		filter,
		userID,
		orgID,
	)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), "")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Inventory lots retrieved successfully", response)
}

// GetInventoryLot retrieves a specific inventory lot
// @Summary Get inventory lot
// @Description Get inventory lot by ID
// @Tags inventory
// @Accept json
// @Produce json
// @Param id path string true "Inventory lot ID"
// @Success 200 {object} common.Response
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Failure 500 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/inventory/lots/{id} [get]
func (h *InventoryHandler) GetInventoryLot(c *gin.Context) {
	lotID := c.Param("id")
	if lotID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Lot ID is required", "")
		return
	}

	// Get user context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", "")
		return
	}

	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		utils.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Organization context required", "")
		return
	}

	// Get inventory lot
	lot, err := h.inventoryService.GetInventoryLot(
		c.Request.Context(),
		lotID,
		userID,
		orgID,
	)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", err.Error(), "")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Inventory lot retrieved successfully", lot)
}

// AdjustInventoryQuantity adjusts inventory quantity for a specific lot
// @Summary Adjust inventory quantity
// @Description Adjust inventory quantity for a specific lot with audit trail
// @Tags inventory
// @Accept json
// @Produce json
// @Param id path string true "Inventory lot ID"
// @Param request body AdjustInventoryRequest true "Inventory adjustment request"
// @Success 200 {object} common.Response
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Failure 500 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/inventory/lots/{id}/adjust [patch]
func (h *InventoryHandler) AdjustInventoryQuantity(c *gin.Context) {
	lotID := c.Param("id")
	if lotID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Lot ID is required", "")
		return
	}

	var req AdjustInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	// Get user context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", "")
		return
	}

	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		utils.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Organization context required", "")
		return
	}

	// Adjust inventory
	lot, err := h.inventoryService.AdjustInventory(
		c.Request.Context(),
		lotID,
		req.Adjustment,
		req.Reason,
		userID,
		orgID,
	)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "SERVICE_ERROR", err.Error(), "")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Inventory adjusted successfully", lot)
}

// CheckInventoryAvailability checks inventory availability for a catalog item
// @Summary Check inventory availability
// @Description Check real-time inventory availability for a catalog item
// @Tags inventory
// @Accept json
// @Produce json
// @Param catalog_item_id path string true "Catalog item ID"
// @Param required_quantity query number false "Required quantity to check" default(1)
// @Success 200 {object} common.Response
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Failure 500 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/inventory/availability/{catalog_item_id} [get]
func (h *InventoryHandler) CheckInventoryAvailability(c *gin.Context) {
	catalogItemID := c.Param("catalog_item_id")
	if catalogItemID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Catalog item ID is required", "")
		return
	}

	// Get user context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", "")
		return
	}

	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		utils.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Organization context required", "")
		return
	}

	// Parse required quantity
	requiredQuantity := decimal.NewFromInt(1) // Default to 1
	if reqQtyStr := c.Query("required_quantity"); reqQtyStr != "" {
		if reqQty, err := decimal.NewFromString(reqQtyStr); err == nil && reqQty.GreaterThan(decimal.Zero) {
			requiredQuantity = reqQty
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid required_quantity parameter", "")
			return
		}
	}

	// Check availability
	available, availableQty, err := h.inventoryService.CheckAvailability(
		c.Request.Context(),
		catalogItemID,
		requiredQuantity,
		userID,
		orgID,
	)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "SERVICE_ERROR", err.Error(), "")
		return
	}

	// Get total quantity
	totalQty, err := h.inventoryService.GetTotalQuantity(
		c.Request.Context(),
		catalogItemID,
		userID,
		orgID,
	)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), "")
		return
	}

	response := InventoryAvailabilityResponse{
		CatalogItemID:     catalogItemID,
		Available:         available,
		AvailableQuantity: availableQty,
		TotalQuantity:     totalQty,
	}

	utils.SuccessResponse(c, http.StatusOK, "Inventory availability checked successfully", response)
}

// GetInventoryAuditTrail gets audit trail for an inventory lot
// @Summary Get inventory audit trail
// @Description Get audit trail for an inventory lot with pagination
// @Tags inventory
// @Accept json
// @Produce json
// @Param id path string true "Inventory lot ID"
// @Param offset query int false "Pagination offset" default(0)
// @Param limit query int false "Pagination limit (max 100)" default(20)
// @Success 200 {object} common.Response
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Failure 500 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/inventory/lots/{id}/audit [get]
func (h *InventoryHandler) GetInventoryAuditTrail(c *gin.Context) {
	lotID := c.Param("id")
	if lotID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Lot ID is required", "")
		return
	}

	// Get user context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", "")
		return
	}

	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		utils.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Organization context required", "")
		return
	}

	// Parse pagination
	offset := 0
	limit := 20

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	// Get audit trail
	response, err := h.inventoryService.GetAuditTrail(
		c.Request.Context(),
		lotID,
		offset,
		limit,
		userID,
		orgID,
	)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "SERVICE_ERROR", err.Error(), "")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Audit trail retrieved successfully", response)
}
