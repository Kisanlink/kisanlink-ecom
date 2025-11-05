package orders

import (
	"fmt"
	"strconv"
	"time"

	orderModels "kisanlink-ecom/entities/models/orders"
	orders "kisanlink-ecom/entities/requests/orders"
	orderResponses "kisanlink-ecom/entities/responses/orders"
	"kisanlink-ecom/internal/common"
	"kisanlink-ecom/internal/middleware"
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
// @Param order body object true "Order information"
// @Success 201 {object} common.Response{data=map[string]interface{}}
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
	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Get organization ID from context
	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "INVALID_ORG", "Valid organization ID is required", nil)
		return
	}

	// Create the order
	order, err := h.orderService.CreateOrder(c.Request.Context(), &req, userID, orgID)
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
// @Success 200 {object} common.Response{data=map[string]interface{}}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/orders/{id} [get]
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Order ID is required", nil)
		return
	}

	// Get user ID from context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Get organization ID from context
	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "INVALID_ORG", "Valid organization ID is required", nil)
		return
	}

	order, err := h.orderService.GetOrderByID(c.Request.Context(), id, userID, orgID)
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
// @Param status body object true "Status update"
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
	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Get organization ID from context
	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "INVALID_ORG", "Valid organization ID is required", nil)
		return
	}

	// Update order status
	err := h.orderService.UpdateOrderStatus(c.Request.Context(), id, &req, userID, orgID)
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
// @Param order body object true "Order update data"
// @Success 200 {object} common.Response{data=map[string]interface{}}
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
	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Get organization ID from context
	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "INVALID_ORG", "Valid organization ID is required", nil)
		return
	}

	// Update the order
	order, err := h.orderService.UpdateOrder(c.Request.Context(), id, &req, userID, orgID)
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
// @Param include_deleted query bool false "Include soft-deleted items (admin only)" default(false)
// @Success 200 {object} common.Response{data=[]interface{},meta=common.ResponseMeta{pagination=common.PaginationMeta}}
// @Router /api/v1/orders [get]
func (h *OrderHandler) ListOrders(c *gin.Context) {
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
	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Get organization ID from context
	// For admins, this may not be used for filtering
	orgID, _ := common.GetOrganizationID(c)

	// Check if user is admin
	roles, rolesExist := common.GetUserRoles(c)
	isAdmin := common.IsAdmin(c)

	// Debug logging
	fmt.Printf("[DEBUG] OrderHandler.ListOrders: roles_exist=%v, roles=%v, is_admin=%v, org_id=%s, user_id=%s\n",
		rolesExist, roles, isAdmin, orgID, userID)

	// Get orders
	orders, total, err := h.orderService.ListOrders(c.Request.Context(), filter, userID, orgID, isAdmin, offset, limit)
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
	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Get organization ID from context
	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "INVALID_ORG", "Valid organization ID is required", nil)
		return
	}

	// Cancel the order
	err := h.orderService.CancelOrder(c.Request.Context(), id, userID, orgID)
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

// CreateOrderFromBid godoc
// @Summary Create an order from a winning bid
// @Description Create an order from a marketplace winning bid
// @Tags orders
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param order body object true "Order from bid information"
// @Success 201 {object} common.Response{data=map[string]interface{}}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/orders/from-bid [post]
func (h *OrderHandler) CreateOrderFromBid(c *gin.Context) {
	var req orders.CreateOrderFromBidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Get organization ID from context
	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "INVALID_ORG", "Valid organization ID is required", nil)
		return
	}

	// Create the order from bid
	order, err := h.orderService.CreateOrderFromBid(c.Request.Context(), &req, userID, orgID)
	if err != nil {
		common.BadRequest(c, "CREATE_FROM_BID_FAILED", "Failed to create order from bid", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Build response with bid and listing information
	response := &orders.CreateOrderFromBidResponse{
		OrderID:     order.ID,
		OrderNumber: order.OrderNumber,
		BidID:       req.BidID,
		TotalAmount: order.TotalAmount,
		Currency:    "INR", // Default currency
		Status:      string(order.Status),
		Message:     "Order created successfully from winning bid",
	}

	// Extract listing ID from metadata if available
	if metadata, err := order.GetMetadata(); err == nil {
		if listingID, exists := metadata["listing_id"]; exists {
			if listingIDStr, ok := listingID.(string); ok {
				response.ListingID = listingIDStr
			}
		}
	}

	common.Created(c, response, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// ValidateBidForOrder godoc
// @Summary Validate a bid for order creation
// @Description Validate that a bid can be used to create an order
// @Tags orders
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param validation body object true "Bid validation request"
// @Success 200 {object} common.Response{data=map[string]interface{}}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/orders/validate-bid [post]
func (h *OrderHandler) ValidateBidForOrder(c *gin.Context) {
	var req orders.BidOrderValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Get organization ID from context
	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "INVALID_ORG", "Valid organization ID is required", nil)
		return
	}

	// Validate the bid
	validation, err := h.orderService.ValidateBidForOrder(c.Request.Context(), req.BidID, userID, orgID)
	if err != nil {
		// Return validation error response
		errorResponse := &orderResponses.BidOrderValidationErrorResponse{
			Valid:           false,
			ErrorCode:       "VALIDATION_FAILED",
			ErrorMessage:    "Bid validation failed",
			ValidationError: err.Error(),
		}

		common.BadRequest(c, "BID_VALIDATION_FAILED", "Bid validation failed", map[string]interface{}{
			"validation_error": errorResponse,
		})
		return
	}

	// Build validation response
	response := &orders.BidOrderValidationResponse{
		Valid:           validation.Valid,
		BidID:           validation.BidID,
		ListingID:       validation.ListingID,
		ProductID:       validation.ProductID,
		WinningAmount:   validation.WinningAmount,
		Quantity:        validation.Quantity,
		Currency:        validation.Currency,
		SellerID:        validation.SellerID,
		BuyerID:         validation.BuyerID,
		SellerOrgID:     validation.SellerOrgID,
		BuyerOrgID:      validation.BuyerOrgID,
		ProductName:     validation.ProductName,
		ProductSKU:      validation.ProductSKU,
		ExpiresAt:       validation.ExpiresAt.Format(time.RFC3339),
		CanCreateOrder:  validation.CanCreateOrder,
		ValidationError: validation.ValidationError,
	}

	common.Success(c, response, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// ProcessPaymentForOrder godoc
// @Summary Process payment for an order
// @Description Process payment for an existing order
// @Tags orders
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Order ID"
// @Param payment body object true "Payment processing request"
// @Success 200 {object} common.Response{data=map[string]interface{}}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/orders/{id}/payment [post]
func (h *OrderHandler) ProcessPaymentForOrder(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		common.BadRequest(c, "MISSING_ID", "Order ID is required", nil)
		return
	}

	var req orders.ProcessPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Get organization ID from context
	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "INVALID_ORG", "Valid organization ID is required", nil)
		return
	}

	// Process payment
	paymentResult, err := h.orderService.ProcessPaymentForOrder(c.Request.Context(), orderID, req.PaymentMethod, userID, orgID)
	if err != nil {
		common.BadRequest(c, "PAYMENT_PROCESSING_FAILED", "Failed to process payment", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Build response
	response := &orders.PaymentResponse{
		PaymentID:     paymentResult.PaymentID,
		OrderID:       paymentResult.OrderID,
		Status:        string(paymentResult.Status),
		Amount:        paymentResult.Amount,
		Currency:      paymentResult.Currency,
		PaymentMethod: paymentResult.PaymentMethod,
		ProcessedAt:   paymentResult.ProcessedAt.Format(time.RFC3339),
		Message:       paymentResult.Message,
	}

	common.Success(c, response, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// GetPaymentStatus godoc
// @Summary Get payment status for an order
// @Description Retrieve the payment status for an order
// @Tags orders
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Order ID"
// @Success 200 {object} common.Response{data=map[string]interface{}}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/orders/{id}/payment/status [get]
func (h *OrderHandler) GetPaymentStatus(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		common.BadRequest(c, "MISSING_ID", "Order ID is required", nil)
		return
	}

	// Get user ID from context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Get organization ID from context
	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "INVALID_ORG", "Valid organization ID is required", nil)
		return
	}

	// Get payment status
	paymentResult, err := h.orderService.GetPaymentStatus(c.Request.Context(), orderID, userID, orgID)
	if err != nil {
		common.BadRequest(c, "PAYMENT_STATUS_FAILED", "Failed to get payment status", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Build response
	response := &orders.PaymentResponse{
		PaymentID:     paymentResult.PaymentID,
		OrderID:       paymentResult.OrderID,
		Status:        string(paymentResult.Status),
		Amount:        paymentResult.Amount,
		Currency:      paymentResult.Currency,
		PaymentMethod: paymentResult.PaymentMethod,
		ProcessedAt:   paymentResult.ProcessedAt.Format(time.RFC3339),
		Message:       paymentResult.Message,
	}

	common.Success(c, response, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}
