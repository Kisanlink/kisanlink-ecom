package orders

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"time"

	orderModels "kisanlink-ecom/entities/models/orders"
	orderRequests "kisanlink-ecom/entities/requests/orders"
	"kisanlink-ecom/internal/repositories/orders"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// PaymentScreenshotServiceInterface defines the interface for payment screenshot operations
type PaymentScreenshotServiceInterface interface {
	UploadPaymentScreenshot(ctx context.Context, req *orderRequests.UploadPaymentScreenshotRequest, userID, orgID string, file multipart.File, fileHeader *multipart.FileHeader) (*orderModels.PaymentScreenshot, error)
	VerifyPaymentScreenshot(ctx context.Context, screenshotID string, req *orderRequests.VerifyPaymentScreenshotRequest, verifierID string) (*orderModels.PaymentScreenshot, error)
	GetPaymentScreenshot(ctx context.Context, screenshotID string, userID, orgID string, isAdmin bool) (*orderModels.PaymentScreenshot, error)
	ListPaymentScreenshots(ctx context.Context, filters *orderRequests.PaymentScreenshotFilters, userID, orgID string, isAdmin bool) ([]*orderModels.PaymentScreenshot, int64, error)
	GetPaymentScreenshotsByOrder(ctx context.Context, orderID string, userID, orgID string, isAdmin bool) ([]*orderModels.PaymentScreenshot, error)
	GenerateDownloadURL(ctx context.Context, screenshotID string, userID, orgID string, isAdmin bool) (string, time.Time, error)
	DeletePaymentScreenshot(ctx context.Context, screenshotID string, userID, orgID string, isAdmin bool) error
}

// PaymentScreenshotService provides business logic for payment screenshot operations
type PaymentScreenshotService struct {
	screenshotRepo *orders.PaymentScreenshotRepository
	orderRepo      *orders.OrderRepository
	s3Manager      *db.S3Manager
	s3Bucket       string
	logger         *zap.Logger
}

// NewPaymentScreenshotService creates a new payment screenshot service
func NewPaymentScreenshotService(
	screenshotRepo *orders.PaymentScreenshotRepository,
	orderRepo *orders.OrderRepository,
	s3Manager *db.S3Manager,
	s3Bucket string,
	logger *zap.Logger,
) *PaymentScreenshotService {
	return &PaymentScreenshotService{
		screenshotRepo: screenshotRepo,
		orderRepo:      orderRepo,
		s3Manager:      s3Manager,
		s3Bucket:       s3Bucket,
		logger:         logger,
	}
}

// UploadPaymentScreenshot uploads a payment screenshot for an order
func (s *PaymentScreenshotService) UploadPaymentScreenshot(
	ctx context.Context,
	req *orderRequests.UploadPaymentScreenshotRequest,
	userID, orgID string,
	file multipart.File,
	fileHeader *multipart.FileHeader,
) (*orderModels.PaymentScreenshot, error) {
	// Validate input
	if req == nil {
		return nil, fmt.Errorf("upload request cannot be nil")
	}
	if userID == "" || orgID == "" {
		return nil, fmt.Errorf("user ID and organization ID are required")
	}
	if file == nil || fileHeader == nil {
		return nil, fmt.Errorf("file is required")
	}

	// Get and validate order
	order, err := s.orderRepo.GetOrderByID(ctx, req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	// Verify order ownership - only buyer organization can upload payment screenshots
	if order.BuyerOrganizationID != orgID {
		return nil, fmt.Errorf("unauthorized: order does not belong to your organization")
	}

	// Validate file size (10MB limit)
	if fileHeader.Size > 10*1024*1024 {
		return nil, fmt.Errorf("file size exceeds 10MB limit")
	}

	// Validate file type
	mimeType := fileHeader.Header.Get("Content-Type")
	if !isAllowedMimeType(mimeType) {
		return nil, fmt.Errorf("unsupported file type: %s. Allowed types: image/jpeg, image/png, image/webp, application/pdf", mimeType)
	}

	// Calculate file checksum
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return nil, fmt.Errorf("failed to calculate file checksum: %w", err)
	}
	checksum := fmt.Sprintf("%x", hasher.Sum(nil))

	// Reset file pointer after checksum calculation
	if _, err := file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to reset file pointer: %w", err)
	}

	// Check for duplicate checksum
	isDuplicate, err := s.screenshotRepo.CheckDuplicateChecksum(ctx, req.OrderID, checksum)
	if err != nil {
		s.logger.Warn("Failed to check duplicate checksum", zap.Error(err))
	} else if isDuplicate {
		return nil, fmt.Errorf("duplicate file: a payment screenshot with this content already exists for this order")
	}

	// Generate S3 key with organized structure
	screenshotID := uuid.New().String()
	ext := filepath.Ext(fileHeader.Filename)
	now := time.Now()
	s3Key := fmt.Sprintf("payment-screenshots/%d/%02d/%s/%s/%s%s",
		now.Year(), now.Month(), orgID, req.OrderID, screenshotID, ext)

	// Upload file to S3
	if err := s.s3Manager.UploadFile(ctx, s3Key, file, mimeType, map[string]string{
		"order_id":        req.OrderID,
		"uploaded_by":     userID,
		"organization_id": orgID,
		"checksum":        checksum,
	}); err != nil {
		return nil, fmt.Errorf("failed to upload file to S3: %w", err)
	}

	s.logger.Info("File uploaded to S3",
		zap.String("s3_key", s3Key),
		zap.String("order_id", req.OrderID),
		zap.String("user_id", userID))

	// Parse payment date
	paymentDate, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		return nil, fmt.Errorf("invalid payment date format: %w", err)
	}

	// Create payment screenshot record
	screenshot := orderModels.NewPaymentScreenshot(req.OrderID, userID, order.BuyerOrganizationID, order.SellerOrganizationID)
	screenshot.FileName = fileHeader.Filename
	screenshot.FilePath = s3Key
	screenshot.S3Bucket = s.s3Bucket
	screenshot.FileSizeBytes = fileHeader.Size
	screenshot.FileMimeType = mimeType
	screenshot.FileChecksum = checksum
	screenshot.PaymentMethod = orderModels.PaymentMethod(req.PaymentMethod)
	screenshot.PaymentDate = &paymentDate
	screenshot.AmountPaid = req.AmountPaid
	screenshot.TransactionID = req.TransactionID
	screenshot.Description = req.Description

	// Create in database
	if err := s.screenshotRepo.CreatePaymentScreenshot(ctx, screenshot); err != nil {
		// Attempt to clean up S3 file on database error
		if deleteErr := s.s3Manager.Delete(ctx, s3Key); deleteErr != nil {
			s.logger.Error("Failed to clean up S3 file after database error",
				zap.Error(deleteErr),
				zap.String("s3_key", s3Key))
		}
		return nil, fmt.Errorf("failed to create payment screenshot record: %w", err)
	}

	s.logger.Info("Payment screenshot uploaded successfully",
		zap.String("screenshot_id", screenshot.ID),
		zap.String("order_id", req.OrderID))

	return screenshot, nil
}

// VerifyPaymentScreenshot allows admins to verify/approve/reject payment screenshots
func (s *PaymentScreenshotService) VerifyPaymentScreenshot(
	ctx context.Context,
	screenshotID string,
	req *orderRequests.VerifyPaymentScreenshotRequest,
	verifierID string,
) (*orderModels.PaymentScreenshot, error) {
	// Validate input
	if screenshotID == "" {
		return nil, fmt.Errorf("screenshot ID is required")
	}
	if req == nil {
		return nil, fmt.Errorf("verification request cannot be nil")
	}
	if verifierID == "" {
		return nil, fmt.Errorf("verifier ID is required")
	}

	// Get screenshot
	screenshot, err := s.screenshotRepo.GetPaymentScreenshot(ctx, screenshotID)
	if err != nil {
		return nil, fmt.Errorf("screenshot not found: %w", err)
	}

	// Check if screenshot can be verified
	if !screenshot.CanVerify() {
		return nil, fmt.Errorf("screenshot with status %s cannot be verified", screenshot.VerificationStatus)
	}

	// Update verification status
	if err := s.screenshotRepo.UpdateVerificationStatus(ctx, screenshotID, req.Status, verifierID, req.Notes); err != nil {
		return nil, fmt.Errorf("failed to update verification status: %w", err)
	}

	// Get updated screenshot
	updatedScreenshot, err := s.screenshotRepo.GetPaymentScreenshot(ctx, screenshotID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated screenshot: %w", err)
	}

	s.logger.Info("Payment screenshot verification updated",
		zap.String("screenshot_id", screenshotID),
		zap.String("status", string(req.Status)),
		zap.String("verifier_id", verifierID))

	return updatedScreenshot, nil
}

// GetPaymentScreenshot retrieves a payment screenshot by ID with authorization checks
func (s *PaymentScreenshotService) GetPaymentScreenshot(
	ctx context.Context,
	screenshotID string,
	userID, orgID string,
	isAdmin bool,
) (*orderModels.PaymentScreenshot, error) {
	if screenshotID == "" {
		return nil, fmt.Errorf("screenshot ID is required")
	}

	screenshot, err := s.screenshotRepo.GetPaymentScreenshot(ctx, screenshotID)
	if err != nil {
		return nil, fmt.Errorf("screenshot not found: %w", err)
	}

	// Authorization check
	if !isAdmin {
		// Non-admin users can only access their own organization's screenshots
		if screenshot.BuyerOrganizationID != orgID && screenshot.SellerOrganizationID != orgID {
			return nil, fmt.Errorf("unauthorized: you do not have access to this screenshot")
		}
	}

	return screenshot, nil
}

// ListPaymentScreenshots lists payment screenshots with filtering and authorization
func (s *PaymentScreenshotService) ListPaymentScreenshots(
	ctx context.Context,
	filters *orderRequests.PaymentScreenshotFilters,
	userID, orgID string,
	isAdmin bool,
) ([]*orderModels.PaymentScreenshot, int64, error) {
	if filters == nil {
		filters = &orderRequests.PaymentScreenshotFilters{
			Page:      1,
			PageSize:  20,
			SortBy:    "created_at",
			SortOrder: "desc",
		}
	}

	// Apply authorization filters for non-admin users
	if !isAdmin {
		// Non-admin users can only see their organization's screenshots
		filters.BuyerOrganizationID = &orgID
	}

	return s.screenshotRepo.ListPaymentScreenshots(ctx, filters)
}

// GetPaymentScreenshotsByOrder retrieves all payment screenshots for a specific order
func (s *PaymentScreenshotService) GetPaymentScreenshotsByOrder(
	ctx context.Context,
	orderID string,
	userID, orgID string,
	isAdmin bool,
) ([]*orderModels.PaymentScreenshot, error) {
	if orderID == "" {
		return nil, fmt.Errorf("order ID is required")
	}

	// Verify order access
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	// Authorization check
	if !isAdmin {
		if order.BuyerOrganizationID != orgID && order.SellerOrganizationID != orgID {
			return nil, fmt.Errorf("unauthorized: you do not have access to this order")
		}
	}

	return s.screenshotRepo.GetByOrderID(ctx, orderID)
}

// GenerateDownloadURL generates a presigned URL for downloading a payment screenshot
func (s *PaymentScreenshotService) GenerateDownloadURL(
	ctx context.Context,
	screenshotID string,
	userID, orgID string,
	isAdmin bool,
) (string, time.Time, error) {
	// Get and authorize screenshot access
	screenshot, err := s.GetPaymentScreenshot(ctx, screenshotID, userID, orgID, isAdmin)
	if err != nil {
		return "", time.Time{}, err
	}

	// Generate presigned URL with 24-hour expiry
	expiryDuration := 24 * time.Hour
	downloadURL, err := s.s3Manager.GetPresignedURL(ctx, screenshot.FilePath, expiryDuration)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate download URL: %w", err)
	}

	expiresAt := time.Now().Add(expiryDuration)

	s.logger.Info("Generated download URL for payment screenshot",
		zap.String("screenshot_id", screenshotID),
		zap.String("user_id", userID),
		zap.Time("expires_at", expiresAt))

	return downloadURL, expiresAt, nil
}

// DeletePaymentScreenshot soft deletes a payment screenshot
func (s *PaymentScreenshotService) DeletePaymentScreenshot(
	ctx context.Context,
	screenshotID string,
	userID, orgID string,
	isAdmin bool,
) error {
	if screenshotID == "" {
		return fmt.Errorf("screenshot ID is required")
	}

	// Get and authorize screenshot access
	screenshot, err := s.GetPaymentScreenshot(ctx, screenshotID, userID, orgID, isAdmin)
	if err != nil {
		return err
	}

	// Only allow deletion by uploader or admin
	if !isAdmin && screenshot.UploadedBy != userID {
		return fmt.Errorf("unauthorized: only the uploader or admin can delete this screenshot")
	}

	// Only allow deletion if not approved
	if screenshot.IsApproved() {
		return fmt.Errorf("cannot delete approved payment screenshots")
	}

	// Soft delete in database
	if err := s.screenshotRepo.DeletePaymentScreenshot(ctx, screenshotID, userID); err != nil {
		return fmt.Errorf("failed to delete screenshot: %w", err)
	}

	s.logger.Info("Payment screenshot deleted",
		zap.String("screenshot_id", screenshotID),
		zap.String("deleted_by", userID))

	return nil
}

// isAllowedMimeType checks if the file MIME type is allowed for payment screenshots
func isAllowedMimeType(mimeType string) bool {
	allowed := []string{
		"image/jpeg",
		"image/png",
		"image/webp",
		"application/pdf",
	}

	for _, allowedType := range allowed {
		if allowedType == mimeType {
			return true
		}
	}

	return false
}

// ValidateAmountMatch validates if the payment amount matches the order total (warning only)
func (s *PaymentScreenshotService) ValidateAmountMatch(
	ctx context.Context,
	orderID string,
	paidAmount decimal.Decimal,
) (bool, string, error) {
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return false, "", fmt.Errorf("failed to get order: %w", err)
	}

	if paidAmount.Equal(order.TotalAmount) {
		return true, "Payment amount matches order total", nil
	}

	if paidAmount.LessThan(order.TotalAmount) {
		paidFloat, _ := paidAmount.Float64()
		totalFloat, _ := order.TotalAmount.Float64()
		return false, fmt.Sprintf("Payment amount (%.2f) is less than order total (%.2f). Partial payment detected.",
			paidFloat, totalFloat), nil
	}

	paidFloat, _ := paidAmount.Float64()
	totalFloat, _ := order.TotalAmount.Float64()
	return false, fmt.Sprintf("Payment amount (%.2f) is greater than order total (%.2f). Overpayment detected.",
		paidFloat, totalFloat), nil
}
