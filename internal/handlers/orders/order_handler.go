package orders

import (
	"strconv"

	orderModels "kisanlink-ecom/entities/models/orders"
	orders "kisanlink-ecom/entities/requests/orders"
	_ "kisanlink-ecom/entities/responses/orders" // For Swagger documentation
	"kisanlink-ecom/internal/common"
	orderService "kisanlink-ecom/internal/services/orders"

	"github.com/gin-gonic/gin"
)

// OrderHandler handles HTTP requests for order operations
type OrderHandler struct {
	orderService orderService.OrderServiceInterface
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(orderService orderService.OrderServiceInterface) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// CreateOrder godoc
// @Summary Create a new order
// @Description Create a new order with items
// @Tags orders
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param order body orders.CreateOrderRequest true "Order information"
// @Success 201 {object} common.Response{data=orders.OrderResponse}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req orders.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Normalize catalog item types to handle case sensitivity
	for i := range req.Items {
		req.Items[i].CatalogItemType = normalizeCatalogItemType(req.Items[i].CatalogItemType)
	}

	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Get organization ID from context
	orgID, exists := c.Get("organizationID")
	if !exists {
		common.Unauthorized(c, "MISSING_ORG", "Organization ID not found in context", nil)
		return
	}

	// Create the order
	order, err := h.orderService.CreateOrder(c.Request.Context(), &req, userID.(string), orgID.(string))
	if err != nil {
		common.BadRequest(c, "CREATE_FAILED", "Failed to create order", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Created(c, order, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// normalizeCatalogItemType normalizes catalog item types to handle case sensitivity
func normalizeCatalogItemType(itemType string) string {
	switch itemType {
	case "product", "PRODUCT":
		return "product"
	case "service", "SERVICE":
		return "service"
	case "labour", "LABOUR":
		return "labour"
	default:
		return itemType
	}
}

// GetOrderByID godoc
// @Summary Get order by ID
// @Description Retrieve an order by its ID
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} common.Response{data=orders.OrderResponse}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/orders/{id} [get]
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Order ID is required", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Get organization ID from context
	orgID, exists := c.Get("organizationID")
	if !exists {
		common.Unauthorized(c, "MISSING_ORG", "Organization ID not found in context", nil)
		return
	}

	order, err := h.orderService.GetOrderByID(c.Request.Context(), id, userID.(string), orgID.(string))
	if err != nil {
		common.NotFound(c, "ORDER_NOT_FOUND", "Order not found", map[string]interface{}{
			"order_id": id,
			"error":    err.Error(),
		})
		return
	}

	common.Success(c, order, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// UpdateOrderStatus godoc
// @Summary Update order status
// @Description Update the status of an order
// @Tags orders
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Order ID"
// @Param status body orders.UpdateOrderStatusRequest true "Status update"
// @Success 200 {object} common.Response{data=string}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/orders/{id}/status [patch]
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Order ID is required", nil)
		return
	}

	var req orders.UpdateOrderStatusRequest
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

	// Get organization ID from context
	orgID, exists := c.Get("organizationID")
	if !exists {
		common.Unauthorized(c, "MISSING_ORG", "Organization ID not found in context", nil)
		return
	}

	// Update order status
	err := h.orderService.UpdateOrderStatus(c.Request.Context(), id, &req, userID.(string), orgID.(string))
	if err != nil {
		common.BadRequest(c, "UPDATE_FAILED", "Failed to update order status", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, "Order status updated successfully", &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// UpdateOrder godoc
// @Summary Update an order
// @Description Update order details including status
// @Tags orders
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Order ID"
// @Param order body orders.UpdateOrderRequest true "Order update data"
// @Success 200 {object} common.Response{data=orders.OrderResponse}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/orders/{id} [put]
func (h *OrderHandler) UpdateOrder(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Order ID is required", nil)
		return
	}

	var req orders.UpdateOrderRequest
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

	// Get organization ID from context
	orgID, exists := c.Get("organizationID")
	if !exists {
		common.Unauthorized(c, "MISSING_ORG", "Organization ID not found in context", nil)
		return
	}

	// Update the order
	order, err := h.orderService.UpdateOrder(c.Request.Context(), id, &req, userID.(string), orgID.(string))
	if err != nil {
		common.BadRequest(c, "UPDATE_FAILED", "Failed to update order", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, order, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// ListOrders godoc
// @Summary List orders
// @Description Retrieve a list of orders with filtering and pagination
// @Tags orders
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param buyer_id query string false "Filter by buyer ID"
// @Param seller_id query string false "Filter by seller ID"
// @Param status query string false "Filter by status"
// @Success 200 {object} common.Response{data=[]orders.OrderResponse,meta=common.ResponseMeta{pagination=common.PaginationMeta}}
// @Router /api/v1/orders [get]
func (h *OrderHandler) ListOrders(c *gin.Context) {
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
	filter := &orders.ListOrdersRequest{
		Page:     page,
		PageSize: limit,
	}

	if buyerID := c.Query("buyer_id"); buyerID != "" {
		filter.BuyerOrganizationID = &buyerID
	}
	if sellerID := c.Query("seller_id"); sellerID != "" {
		filter.SellerOrganizationID = &sellerID
	}
	if status := c.Query("status"); status != "" {
		statusValue := orderModels.OrderStatus(status)
		filter.Status = &statusValue
	}

	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Get organization ID from context
	orgID, exists := c.Get("organizationID")
	if !exists {
		common.Unauthorized(c, "MISSING_ORG", "Organization ID not found in context", nil)
		return
	}

	// Get orders
	orders, total, err := h.orderService.ListOrders(c.Request.Context(), filter, userID.(string), orgID.(string), offset, limit)
	if err != nil {
		common.InternalServerError(c, "LIST_FAILED", "Failed to list orders", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, orders, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
		Pagination: &common.PaginationMeta{
			Page:    page,
			Limit:   limit,
			Total:   total,
			HasNext: offset+limit < total,
		},
	})
}

// CancelOrder godoc
// @Summary Cancel an order
// @Description Cancel an order and release inventory
// @Tags orders
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Order ID"
// @Success 200 {object} common.Response{data=string}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/orders/{id}/cancel [post]
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Order ID is required", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Get organization ID from context
	orgID, exists := c.Get("organizationID")
	if !exists {
		common.Unauthorized(c, "MISSING_ORG", "Organization ID not found in context", nil)
		return
	}

	// Cancel the order
	err := h.orderService.CancelOrder(c.Request.Context(), id, userID.(string), orgID.(string))
	if err != nil {
		common.BadRequest(c, "CANCEL_FAILED", "Failed to cancel order", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, "Order cancelled successfully", &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}
