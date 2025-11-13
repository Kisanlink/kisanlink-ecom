package orders

import (
	"strconv"

	orderModels "kisanlink-ecom/entities/models/orders"
	"kisanlink-ecom/internal/common"
	"kisanlink-ecom/internal/middleware"
	orderService "kisanlink-ecom/internal/services/orders"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// InvoiceHandler handles HTTP requests for invoice operations
type InvoiceHandler struct {
	invoiceService orderService.InvoiceServiceInterface
}

// NewInvoiceHandler creates a new invoice handler
func NewInvoiceHandler(invoiceService orderService.InvoiceServiceInterface) *InvoiceHandler {
	return &InvoiceHandler{
		invoiceService: invoiceService,
	}
}

// GenerateInvoice godoc
// @Summary Generate invoice for an order
// @Description Generate an invoice from an order
// @Tags invoices
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Order ID"
// @Success 201 {object} common.Response{data=object}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/orders/{id}/invoice [post]
func (h *InvoiceHandler) GenerateInvoice(c *gin.Context) {
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

	// Generate the invoice
	invoice, err := h.invoiceService.GenerateInvoice(c.Request.Context(), orderID, userID, orgID)
	if err != nil {
		common.BadRequest(c, "INVOICE_GENERATION_FAILED", "Failed to generate invoice", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Created(c, invoice, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// GetInvoiceByID godoc
// @Summary Get invoice by ID
// @Description Retrieve an invoice by its ID
// @Tags invoices
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Invoice ID"
// @Success 200 {object} common.Response{data=object}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/invoices/{id} [get]
func (h *InvoiceHandler) GetInvoiceByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Invoice ID is required", nil)
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

	invoice, err := h.invoiceService.GetInvoiceByID(c.Request.Context(), id, userID, orgID)
	if err != nil {
		common.NotFound(c, "INVOICE_NOT_FOUND", "Invoice not found", map[string]interface{}{
			"invoice_id": id,
			"error":      err.Error(),
		})
		return
	}

	common.Success(c, invoice, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// GetInvoiceByOrderID godoc
// @Summary Get invoice by order ID
// @Description Retrieve an invoice by order ID
// @Tags invoices
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Order ID"
// @Success 200 {object} common.Response{data=object}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/orders/{id}/invoice [get]
func (h *InvoiceHandler) GetInvoiceByOrderID(c *gin.Context) {
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

	invoice, err := h.invoiceService.GetInvoiceByOrderID(c.Request.Context(), orderID, userID, orgID)
	if err != nil {
		common.NotFound(c, "INVOICE_NOT_FOUND", "Invoice not found for order", map[string]interface{}{
			"order_id": orderID,
			"error":    err.Error(),
		})
		return
	}

	common.Success(c, invoice, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// DownloadInvoicePDF godoc
// @Summary Download invoice PDF
// @Description Download invoice as PDF
// @Tags invoices
// @Accept json
// @Produce application/pdf
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Order ID"
// @Success 200 {file} application/pdf
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/orders/{id}/invoice/pdf [get]
func (h *InvoiceHandler) DownloadInvoicePDF(c *gin.Context) {
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

	// Get the invoice by order ID first
	invoice, err := h.invoiceService.GetInvoiceByOrderID(c.Request.Context(), orderID, userID, orgID)
	if err != nil {
		common.NotFound(c, "INVOICE_NOT_FOUND", "Invoice not found for order", map[string]interface{}{
			"order_id": orderID,
			"error":    err.Error(),
		})
		return
	}

	// Generate PDF
	pdfContent, err := h.invoiceService.GenerateInvoicePDF(c.Request.Context(), invoice.ID, userID, orgID)
	if err != nil {
		common.InternalServerError(c, "PDF_GENERATION_FAILED", "Failed to generate PDF", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Set headers for PDF download
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=invoice-"+invoice.InvoiceNumber+".txt")
	c.Header("Content-Length", strconv.Itoa(len(pdfContent)))

	// Write PDF content
	c.Data(200, "text/plain; charset=utf-8", pdfContent)
}

// ListInvoices godoc
// @Summary List invoices
// @Description Retrieve a list of invoices with filtering and pagination
// @Tags invoices
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param buyer_id query string false "Filter by buyer ID"
// @Param seller_id query string false "Filter by seller ID"
// @Param order_id query string false "Filter by order ID"
// @Param status query string false "Filter by status"
// @Param include_deleted query bool false "Include soft-deleted items (admin only)" default(false)
// @Success 200 {object} common.Response{data=[]object,meta=common.ResponseMeta{pagination=common.PaginationMeta}}
// @Router /api/v1/invoices [get]
func (h *InvoiceHandler) ListInvoices(c *gin.Context) {
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
	filter := &orderModels.InvoiceFilter{}

	if buyerID := c.Query("buyer_id"); buyerID != "" {
		filter.BuyerID = buyerID
	}
	if sellerID := c.Query("seller_id"); sellerID != "" {
		filter.SellerID = sellerID
	}
	if orderID := c.Query("order_id"); orderID != "" {
		filter.OrderID = orderID
	}
	if status := c.Query("status"); status != "" {
		statusValue := orderModels.InvoiceStatus(status)
		filter.Status = &statusValue
	}

	// Get user ID from context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Get organization ID from context
	orgID, _ := common.GetOrganizationID(c)

	// Check if user is admin
	isAdmin := common.IsAdmin(c)

	// Get invoices
	invoices, total, err := h.invoiceService.ListInvoices(c.Request.Context(), filter, userID, orgID, isAdmin, offset, limit)
	if err != nil {
		common.InternalServerError(c, "LIST_FAILED", "Failed to list invoices", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, invoices, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
		Pagination: &common.PaginationMeta{
			Page:    page,
			Limit:   limit,
			Total:   total,
			HasNext: offset+limit < total,
		},
	})
}

// FinalizeInvoice godoc
// @Summary Finalize invoice
// @Description Finalize an invoice (convert from draft to final)
// @Tags invoices
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Invoice ID"
// @Success 200 {object} common.Response{data=object}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/invoices/{id}/finalize [post]
func (h *InvoiceHandler) FinalizeInvoice(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Invoice ID is required", nil)
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

	// Finalize the invoice
	invoice, err := h.invoiceService.FinalizeInvoice(c.Request.Context(), id, userID, orgID)
	if err != nil {
		common.BadRequest(c, "FINALIZE_FAILED", "Failed to finalize invoice", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, invoice, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// MarkInvoiceAsPaid godoc
// @Summary Mark invoice as paid
// @Description Mark an invoice as paid
// @Tags invoices
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Invoice ID"
// @Param payment body object{paid_amount=decimal.Decimal,payment_method=string} true "Payment information"
// @Success 200 {object} common.Response{data=object}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/invoices/{id}/mark-paid [post]
func (h *InvoiceHandler) MarkInvoiceAsPaid(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Invoice ID is required", nil)
		return
	}

	// Parse request body
	var req struct {
		PaidAmount    decimal.Decimal `json:"paid_amount" binding:"required"`
		PaymentMethod string          `json:"payment_method" binding:"required"`
	}

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

	// Mark the invoice as paid
	invoice, err := h.invoiceService.MarkInvoiceAsPaid(c.Request.Context(), id, req.PaidAmount, req.PaymentMethod, userID, orgID)
	if err != nil {
		common.BadRequest(c, "MARK_PAID_FAILED", "Failed to mark invoice as paid", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, invoice, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// VoidInvoice godoc
// @Summary Void invoice
// @Description Void an invoice
// @Tags invoices
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Invoice ID"
// @Success 200 {object} common.Response{data=object}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/invoices/{id}/void [post]
func (h *InvoiceHandler) VoidInvoice(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Invoice ID is required", nil)
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

	// Void the invoice
	invoice, err := h.invoiceService.VoidInvoice(c.Request.Context(), id, userID, orgID)
	if err != nil {
		common.BadRequest(c, "VOID_FAILED", "Failed to void invoice", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, invoice, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}
