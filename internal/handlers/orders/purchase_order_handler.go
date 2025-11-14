package orders

import (
	"net/http"
	"strconv"

	orderModels "github.com/Kisanlink/kisanlink-ecom/entities/models/orders"
	orders "github.com/Kisanlink/kisanlink-ecom/entities/requests/orders"
	"github.com/Kisanlink/kisanlink-ecom/internal/common"
	orderService "github.com/Kisanlink/kisanlink-ecom/internal/services/orders"

	"github.com/gin-gonic/gin"
)

// PurchaseOrderHandler handles HTTP requests for purchase order operations
type PurchaseOrderHandler struct {
	poService orderService.PurchaseOrderServiceInterface
}

// NewPurchaseOrderHandler creates a new purchase order handler
func NewPurchaseOrderHandler(poService orderService.PurchaseOrderServiceInterface) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{
		poService: poService,
	}
}

// CreateManualPO godoc
// @Summary Create manual purchase order
// @Description Create a manual purchase order for external vendor
// @Tags purchase-orders
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param purchase_order body object true "Purchase order information"
// @Success 201 {object} common.Response
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/fpo/purchase-orders [post]
func (h *PurchaseOrderHandler) CreateManualPO(c *gin.Context) {
	var req orders.CreatePurchaseOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	fpoOrgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "INVALID_ORG", "Valid organization ID is required", nil)
		return
	}

	po, err := h.poService.CreateManualPO(c.Request.Context(), &req, userID, fpoOrgID)
	if err != nil {
		common.BadRequest(c, "CREATE_FAILED", "Failed to create purchase order", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Created(c, po, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// ListPurchaseOrders godoc
// @Summary List purchase orders
// @Description List purchase orders with filters
// @Tags purchase-orders
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param status query string false "PO Status"
// @Param source query string false "PO Source (KISANLINK/MANUAL)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} common.Response
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/fpo/purchase-orders [get]
func (h *PurchaseOrderHandler) ListPurchaseOrders(c *gin.Context) {
	var req orders.ListPurchaseOrdersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid query parameters", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	fpoOrgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "INVALID_ORG", "Valid organization ID is required", nil)
		return
	}

	// Set defaults
	page := 1
	pageSize := 20
	if req.Page > 0 {
		page = req.Page
	}
	if req.PageSize > 0 {
		pageSize = req.PageSize
	}

	offset := (page - 1) * pageSize
	poList, total, err := h.poService.ListPOs(c.Request.Context(), &req, userID, fpoOrgID, offset, pageSize)
	if err != nil {
		common.BadRequest(c, "LIST_FAILED", "Failed to list purchase orders", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, poList, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
		Pagination: &common.PaginationMeta{
			Page:       page,
			Limit:      pageSize,
			Total:      total,
			TotalPages: (total + pageSize - 1) / pageSize,
			HasNext:    page*pageSize < total,
			HasPrev:    page > 1,
		},
	})
}

// GetPurchaseOrderByID godoc
// @Summary Get purchase order by ID
// @Description Retrieve a purchase order by its ID
// @Tags purchase-orders
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Purchase Order ID"
// @Success 200 {object} common.Response
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/fpo/purchase-orders/{id} [get]
func (h *PurchaseOrderHandler) GetPurchaseOrderByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Purchase order ID is required", nil)
		return
	}

	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	fpoOrgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "INVALID_ORG", "Valid organization ID is required", nil)
		return
	}

	po, err := h.poService.GetPOByID(c.Request.Context(), id, userID, fpoOrgID)
	if err != nil {
		common.NotFound(c, "PO_NOT_FOUND", "Purchase order not found", map[string]interface{}{
			"po_id": id,
			"error": err.Error(),
		})
		return
	}

	common.Success(c, po, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// UpdatePurchaseOrderStatus godoc
// @Summary Update purchase order status
// @Description Update the status of a purchase order
// @Tags purchase-orders
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Purchase Order ID"
// @Param status body object true "Status update"
// @Success 200 {object} common.Response
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/fpo/purchase-orders/{id}/status [patch]
func (h *PurchaseOrderHandler) UpdatePurchaseOrderStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Purchase order ID is required", nil)
		return
	}

	var req orders.UpdatePurchaseOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	fpoOrgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "INVALID_ORG", "Valid organization ID is required", nil)
		return
	}

	if err := h.poService.UpdatePOStatus(c.Request.Context(), id, &req, userID, fpoOrgID); err != nil {
		common.BadRequest(c, "UPDATE_FAILED", "Failed to update purchase order status", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, gin.H{"message": "Purchase order status updated successfully"}, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// DownloadPurchaseOrderPDF godoc
// @Summary Download purchase order PDF
// @Description Generate and download purchase order as PDF
// @Tags purchase-orders
// @Accept json
// @Produce application/pdf
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Purchase Order ID"
// @Success 200 {file} binary "PDF file"
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/fpo/purchase-orders/{id}/pdf [get]
func (h *PurchaseOrderHandler) DownloadPurchaseOrderPDF(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Purchase order ID is required", nil)
		return
	}

	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	fpoOrgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "INVALID_ORG", "Valid organization ID is required", nil)
		return
	}

	pdfData, err := h.poService.GeneratePOPDF(c.Request.Context(), id, userID, fpoOrgID)
	if err != nil {
		common.NotFound(c, "PDF_GENERATION_FAILED", "Failed to generate PDF", map[string]interface{}{
			"po_id": id,
			"error": err.Error(),
		})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=po-"+id+".pdf")
	c.Data(http.StatusOK, "application/pdf", pdfData)
}

// CreateGRN godoc
// @Summary Create goods received note
// @Description Create a GRN for a purchase order
// @Tags purchase-orders
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Purchase Order ID"
// @Param grn body object true "GRN information"
// @Success 201 {object} common.Response
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/fpo/purchase-orders/{id}/grn [post]
func (h *PurchaseOrderHandler) CreateGRN(c *gin.Context) {
	poID := c.Param("id")
	if poID == "" {
		common.BadRequest(c, "MISSING_ID", "Purchase order ID is required", nil)
		return
	}

	var req orders.CreateGRNRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	fpoOrgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "INVALID_ORG", "Valid organization ID is required", nil)
		return
	}

	grn, err := h.poService.CreateGRN(c.Request.Context(), poID, &req, userID, fpoOrgID)
	if err != nil {
		common.BadRequest(c, "CREATE_GRN_FAILED", "Failed to create GRN", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Created(c, grn, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// GetGRNsByPO godoc
// @Summary Get GRNs by purchase order
// @Description Retrieve all GRNs for a purchase order
// @Tags purchase-orders
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Purchase Order ID"
// @Success 200 {object} common.Response
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/fpo/purchase-orders/{id}/grn [get]
func (h *PurchaseOrderHandler) GetGRNsByPO(c *gin.Context) {
	poID := c.Param("id")
	if poID == "" {
		common.BadRequest(c, "MISSING_ID", "Purchase order ID is required", nil)
		return
	}

	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	fpoOrgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "INVALID_ORG", "Valid organization ID is required", nil)
		return
	}

	grns, err := h.poService.GetGRNsByPO(c.Request.Context(), poID, userID, fpoOrgID)
	if err != nil {
		common.NotFound(c, "GRNS_NOT_FOUND", "GRNs not found", map[string]interface{}{
			"po_id": poID,
			"error": err.Error(),
		})
		return
	}

	common.Success(c, grns, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// Admin Dashboard Handlers

// AdminDashboardHandler handles admin dashboard operations
type AdminDashboardHandler struct {
	orderService orderService.OrderServiceInterface
	poService    orderService.PurchaseOrderServiceInterface
}

// NewAdminDashboardHandler creates a new admin dashboard handler
func NewAdminDashboardHandler(orderService orderService.OrderServiceInterface, poService orderService.PurchaseOrderServiceInterface) *AdminDashboardHandler {
	return &AdminDashboardHandler{
		orderService: orderService,
		poService:    poService,
	}
}

// GetOrderDashboard godoc
// @Summary Get order management dashboard
// @Description Get order metrics and list for admin dashboard
// @Tags admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param status query string false "Filter by status"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} common.Response
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/admin/orders/dashboard [get]
func (h *AdminDashboardHandler) GetOrderDashboard(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	orgID, _ := common.GetOrganizationID(c)

	// Build filter request
	filter := &orders.ListOrdersRequest{
		Page:     page,
		PageSize: pageSize,
	}

	offset := (page - 1) * pageSize
	orderList, total, err := h.orderService.ListOrders(c.Request.Context(), filter, userID, orgID, true, offset, pageSize)
	if err != nil {
		common.BadRequest(c, "LIST_FAILED", "Failed to list orders", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Calculate basic metrics
	metrics := calculateOrderMetrics(orderList)

	response := gin.H{
		"metrics": metrics,
		"orders":  orderList,
	}

	common.Success(c, response, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
		Pagination: &common.PaginationMeta{
			Page:       page,
			Limit:      pageSize,
			Total:      total,
			TotalPages: (total + pageSize - 1) / pageSize,
			HasNext:    page*pageSize < total,
			HasPrev:    page > 1,
		},
	})
}

// calculateOrderMetrics calculates basic order metrics
func calculateOrderMetrics(orderList []*orderModels.Order) map[string]interface{} {
	totalOrders := len(orderList)
	totalRevenue := 0.0
	statusBreakdown := make(map[string]int)

	for _, order := range orderList {
		revenue, _ := order.TotalAmount.Float64()
		totalRevenue += revenue
		statusBreakdown[string(order.Status)]++
	}

	avgOrderValue := 0.0
	if totalOrders > 0 {
		avgOrderValue = totalRevenue / float64(totalOrders)
	}

	return map[string]interface{}{
		"total_orders":        totalOrders,
		"total_revenue":       totalRevenue,
		"average_order_value": avgOrderValue,
		"status_breakdown":    statusBreakdown,
	}
}
