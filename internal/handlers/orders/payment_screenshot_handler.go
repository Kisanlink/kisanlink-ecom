package orders

import (
	"fmt"

	orderModels "kisanlink-ecom/entities/models/orders"
	orderRequests "kisanlink-ecom/entities/requests/orders"
	orderResponses "kisanlink-ecom/entities/responses/orders"
	"kisanlink-ecom/internal/common"
	orderService "kisanlink-ecom/internal/services/orders"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// PaymentScreenshotHandler handles HTTP requests for payment screenshot operations
type PaymentScreenshotHandler struct {
	paymentScreenshotService orderService.PaymentScreenshotServiceInterface
	orderService             orderService.OrderServiceInterface
}

// NewPaymentScreenshotHandler creates a new payment screenshot handler
func NewPaymentScreenshotHandler(
	paymentScreenshotService orderService.PaymentScreenshotServiceInterface,
	orderService orderService.OrderServiceInterface,
) *PaymentScreenshotHandler {
	return &PaymentScreenshotHandler{
		paymentScreenshotService: paymentScreenshotService,
		orderService:             orderService,
	}
}

// UploadPaymentScreenshot godoc
// @Summary Upload payment screenshot
// @Description Upload a payment screenshot/receipt for an order (buyer only)
// @Tags payment-screenshots
// @Accept multipart/form-data
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param order_id formData string true "Order ID"
// @Param payment_method formData string true "Payment method (bank_transfer, upi, cheque, cash, other)"
// @Param payment_date formData string true "Payment date (YYYY-MM-DD)"
// @Param amount_paid formData number true "Amount paid"
// @Param transaction_id formData string false "Transaction ID"
// @Param description formData string false "Payment description"
// @Param file formData file true "Payment screenshot file (JPEG, PNG, WebP, PDF, max 10MB)"
// @Success 201 {object} common.Response{data=map[string]interface{}}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/orders/{order_id}/payment-screenshots [post]
func (h *PaymentScreenshotHandler) UploadPaymentScreenshot(c *gin.Context) {
	// Get user ID and organization ID from context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "INVALID_ORG", "Valid organization ID is required", nil)
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil { // 10MB
		common.BadRequest(c, "INVALID_FORM", "Failed to parse multipart form", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Extract form fields
	var req orderRequests.UploadPaymentScreenshotRequest
	req.OrderID = c.Param("order_id")
	if req.OrderID == "" {
		req.OrderID = c.PostForm("order_id")
	}
	req.PaymentMethod = c.PostForm("payment_method")
	req.PaymentDate = c.PostForm("payment_date")
	amountPaidStr := c.PostForm("amount_paid")
	req.TransactionID = c.PostForm("transaction_id")
	req.Description = c.PostForm("description")

	// Validate required fields
	if req.OrderID == "" {
		common.BadRequest(c, "MISSING_ORDER_ID", "Order ID is required", nil)
		return
	}
	if req.PaymentMethod == "" {
		common.BadRequest(c, "MISSING_PAYMENT_METHOD", "Payment method is required", nil)
		return
	}
	if req.PaymentDate == "" {
		common.BadRequest(c, "MISSING_PAYMENT_DATE", "Payment date is required", nil)
		return
	}
	if amountPaidStr == "" {
		common.BadRequest(c, "MISSING_AMOUNT", "Amount paid is required", nil)
		return
	}

	// Parse amount paid
	amountPaid, err := parseDecimal(amountPaidStr)
	if err != nil {
		common.BadRequest(c, "INVALID_AMOUNT", "Invalid amount paid format", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	req.AmountPaid = amountPaid

	// Get uploaded file
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		common.BadRequest(c, "MISSING_FILE", "File is required", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	defer file.Close()

	// Upload payment screenshot
	screenshot, err := h.paymentScreenshotService.UploadPaymentScreenshot(
		c.Request.Context(),
		&req,
		userID,
		orgID,
		file,
		fileHeader,
	)
	if err != nil {
		common.BadRequest(c, "UPLOAD_FAILED", "Failed to upload payment screenshot", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Convert to response
	response, err := orderResponses.ToPaymentScreenshotResponse(screenshot)
	if err != nil {
		common.InternalServerError(c, "RESPONSE_ERROR", "Failed to create response", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	uploadResponse := &orderResponses.PaymentScreenshotUploadResponse{
		PaymentScreenshotResponse: *response,
		Message:                   "Payment screenshot uploaded successfully",
	}

	common.Created(c, uploadResponse, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// GetPaymentScreenshot godoc
// @Summary Get payment screenshot details
// @Description Retrieve payment screenshot details by ID
// @Tags payment-screenshots
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param screenshot_id path string true "Payment Screenshot ID"
// @Success 200 {object} common.Response{data=map[string]interface{}}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/payment-screenshots/{screenshot_id} [get]
func (h *PaymentScreenshotHandler) GetPaymentScreenshot(c *gin.Context) {
	screenshotID := c.Param("screenshot_id")
	if screenshotID == "" {
		common.BadRequest(c, "MISSING_ID", "Screenshot ID is required", nil)
		return
	}

	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "INVALID_ORG", "Valid organization ID is required", nil)
		return
	}

	isAdmin := common.IsAdmin(c)

	screenshot, err := h.paymentScreenshotService.GetPaymentScreenshot(c.Request.Context(), screenshotID, userID, orgID, isAdmin)
	if err != nil {
		common.NotFound(c, "SCREENSHOT_NOT_FOUND", "Payment screenshot not found", map[string]interface{}{
			"screenshot_id": screenshotID,
			"error":         err.Error(),
		})
		return
	}

	response, err := orderResponses.ToPaymentScreenshotResponse(screenshot)
	if err != nil {
		common.InternalServerError(c, "RESPONSE_ERROR", "Failed to create response", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, response, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// ListPaymentScreenshotsByOrder godoc
// @Summary List payment screenshots for an order
// @Description Retrieve all payment screenshots for a specific order
// @Tags payment-screenshots
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param order_id path string true "Order ID"
// @Success 200 {object} common.Response{data=[]interface{}}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/orders/{order_id}/payment-screenshots [get]
func (h *PaymentScreenshotHandler) ListPaymentScreenshotsByOrder(c *gin.Context) {
	orderID := c.Param("order_id")
	if orderID == "" {
		common.BadRequest(c, "MISSING_ORDER_ID", "Order ID is required", nil)
		return
	}

	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "INVALID_ORG", "Valid organization ID is required", nil)
		return
	}

	isAdmin := common.IsAdmin(c)

	screenshots, err := h.paymentScreenshotService.GetPaymentScreenshotsByOrder(c.Request.Context(), orderID, userID, orgID, isAdmin)
	if err != nil {
		common.NotFound(c, "ORDER_NOT_FOUND", "Order not found or access denied", map[string]interface{}{
			"order_id": orderID,
			"error":    err.Error(),
		})
		return
	}

	responses, err := orderResponses.ToPaymentScreenshotResponseList(screenshots)
	if err != nil {
		common.InternalServerError(c, "RESPONSE_ERROR", "Failed to create response", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, responses, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// ListPaymentScreenshots godoc
// @Summary List payment screenshots (admin)
// @Description List all payment screenshots with filtering (admin only)
// @Tags payment-screenshots,admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param order_id query string false "Filter by order ID"
// @Param verification_status query string false "Filter by verification status (pending, approved, rejected, disputed)"
// @Param buyer_organization_id query string false "Filter by buyer organization ID"
// @Param seller_organization_id query string false "Filter by seller organization ID"
// @Param payment_method query string false "Filter by payment method"
// @Param sort_by query string false "Sort by field (created_at, updated_at, amount_paid, payment_date)" default("created_at")
// @Param sort_order query string false "Sort order (asc, desc)" default("desc")
// @Success 200 {object} common.Response{data=map[string]interface{}}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/admin/payment-screenshots [get]
func (h *PaymentScreenshotHandler) ListPaymentScreenshots(c *gin.Context) {
	// Admin authorization check
	if !common.IsAdmin(c) {
		common.Forbidden(c, "UNAUTHORIZED", "Admin role required", nil)
		return
	}

	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "INVALID_ORG", "Valid organization ID is required", nil)
		return
	}

	// Parse request
	var req orderRequests.ListPaymentScreenshotsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request parameters", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	filters := req.ToPaymentScreenshotFilters()

	screenshots, total, err := h.paymentScreenshotService.ListPaymentScreenshots(
		c.Request.Context(),
		filters,
		userID,
		orgID,
		true, // isAdmin
	)
	if err != nil {
		common.InternalServerError(c, "LIST_FAILED", "Failed to list payment screenshots", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	responses, err := orderResponses.ToPaymentScreenshotResponseList(screenshots)
	if err != nil {
		common.InternalServerError(c, "RESPONSE_ERROR", "Failed to create response", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	totalPages := int(total) / filters.PageSize
	if int(total)%filters.PageSize > 0 {
		totalPages++
	}

	listResponse := &orderResponses.PaymentScreenshotListResponse{
		Screenshots: responses,
		Total:       total,
		Page:        filters.Page,
		PageSize:    filters.PageSize,
		TotalPages:  totalPages,
	}

	common.Success(c, listResponse, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// VerifyPaymentScreenshot godoc
// @Summary Verify payment screenshot (admin)
// @Description Verify/approve/reject a payment screenshot (admin only)
// @Tags payment-screenshots,admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param screenshot_id path string true "Payment Screenshot ID"
// @Param request body object true "Verification details"
// @Success 200 {object} common.Response{data=map[string]interface{}}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/admin/payment-screenshots/{screenshot_id}/verify [post]
func (h *PaymentScreenshotHandler) VerifyPaymentScreenshot(c *gin.Context) {
	// Admin authorization check
	if !common.IsAdmin(c) {
		common.Forbidden(c, "UNAUTHORIZED", "Admin role required", nil)
		return
	}

	screenshotID := c.Param("screenshot_id")
	if screenshotID == "" {
		common.BadRequest(c, "MISSING_ID", "Screenshot ID is required", nil)
		return
	}

	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	var req orderRequests.VerifyPaymentScreenshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	screenshot, err := h.paymentScreenshotService.VerifyPaymentScreenshot(
		c.Request.Context(),
		screenshotID,
		&req,
		userID,
	)
	if err != nil {
		common.BadRequest(c, "VERIFICATION_FAILED", "Failed to verify payment screenshot", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// If approved, trigger order status update
	if req.Status == orderModels.VerificationStatusApproved {
		// Attempt to update order status to paid
		if updateErr := h.orderService.UpdateOrderStatus(
			c.Request.Context(),
			screenshot.OrderID,
			&orderRequests.UpdateOrderStatusRequest{
				Status: orderModels.OrderStatusPaid,
				Reason: fmt.Sprintf("Payment verified via screenshot %s", screenshotID),
			},
			userID,
			screenshot.SellerOrganizationID,
		); updateErr != nil {
			// Log the error but don't fail the verification
			// The screenshot is still marked as verified
			c.Header("X-Order-Update-Status", "failed")
			c.Header("X-Order-Update-Error", updateErr.Error())
		}
	}

	response, err := orderResponses.ToPaymentScreenshotResponse(screenshot)
	if err != nil {
		common.InternalServerError(c, "RESPONSE_ERROR", "Failed to create response", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, response, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// GetDownloadURL godoc
// @Summary Get payment screenshot download URL
// @Description Generate a presigned URL to download the payment screenshot file
// @Tags payment-screenshots
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param screenshot_id path string true "Payment Screenshot ID"
// @Success 200 {object} common.Response{data=map[string]interface{}}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/payment-screenshots/{screenshot_id}/download [get]
func (h *PaymentScreenshotHandler) GetDownloadURL(c *gin.Context) {
	screenshotID := c.Param("screenshot_id")
	if screenshotID == "" {
		common.BadRequest(c, "MISSING_ID", "Screenshot ID is required", nil)
		return
	}

	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "INVALID_ORG", "Valid organization ID is required", nil)
		return
	}

	isAdmin := common.IsAdmin(c)

	downloadURL, expiresAt, err := h.paymentScreenshotService.GenerateDownloadURL(
		c.Request.Context(),
		screenshotID,
		userID,
		orgID,
		isAdmin,
	)
	if err != nil {
		common.NotFound(c, "SCREENSHOT_NOT_FOUND", "Payment screenshot not found or access denied", map[string]interface{}{
			"screenshot_id": screenshotID,
			"error":         err.Error(),
		})
		return
	}

	response := &orderResponses.PaymentScreenshotDownloadURLResponse{
		ScreenshotID: screenshotID,
		DownloadURL:  downloadURL,
		ExpiresAt:    expiresAt,
	}

	common.Success(c, response, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// Helper function to parse decimal from string
func parseDecimal(s string) (decimal.Decimal, error) {
	return decimal.NewFromString(s)
}
