package orders

import (
	"time"

	orderModels "github.com/Kisanlink/kisanlink-ecom/entities/models/orders"

	"github.com/shopspring/decimal"
)

// PaymentScreenshotResponse represents a payment screenshot in API responses
type PaymentScreenshotResponse struct {
	ID                   string          `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	OrderID              string          `json:"order_id" example:"123e4567-e89b-12d3-a456-426614174001"`
	UploadedBy           string          `json:"uploaded_by" example:"123e4567-e89b-12d3-a456-426614174002"`
	BuyerOrganizationID  string          `json:"buyer_organization_id" example:"123e4567-e89b-12d3-a456-426614174003"`
	SellerOrganizationID string          `json:"seller_organization_id" example:"123e4567-e89b-12d3-a456-426614174004"`
	FileName             string          `json:"file_name" example:"payment_receipt.jpg"`
	FileSizeBytes        int64           `json:"file_size_bytes" example:"245760"`
	FileMimeType         string          `json:"file_mime_type" example:"image/jpeg"`
	PaymentMethod        string          `json:"payment_method" example:"bank_transfer"`
	PaymentDate          *time.Time      `json:"payment_date,omitempty" example:"2024-01-15T00:00:00Z"`
	AmountPaid           decimal.Decimal `json:"amount_paid" example:"500.00"`
	TransactionID        string          `json:"transaction_id,omitempty" example:"TXN123456789"`
	VerificationStatus   string          `json:"verification_status" example:"pending"`
	VerifiedBy           *string         `json:"verified_by,omitempty" example:"123e4567-e89b-12d3-a456-426614174005"`
	VerifiedAt           *time.Time      `json:"verified_at,omitempty" example:"2024-01-16T10:30:00Z"`
	VerificationNotes    string          `json:"verification_notes,omitempty" example:"Payment verified successfully"`
	Description          string          `json:"description,omitempty" example:"Payment for order ORD-2024-001"`
	CreatedAt            time.Time       `json:"created_at" example:"2024-01-15T14:30:00Z"`
	UpdatedAt            time.Time       `json:"updated_at" example:"2024-01-16T10:30:00Z"`
}

// PaymentScreenshotUploadResponse represents the response after uploading a payment screenshot
type PaymentScreenshotUploadResponse struct {
	PaymentScreenshotResponse
	Message string `json:"message" example:"Payment screenshot uploaded successfully"`
}

// PaymentScreenshotDownloadURLResponse represents the response with a presigned download URL
type PaymentScreenshotDownloadURLResponse struct {
	ScreenshotID string    `json:"screenshot_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	DownloadURL  string    `json:"download_url" example:"https://s3.amazonaws.com/bucket/path?signature=..."`
	ExpiresAt    time.Time `json:"expires_at" example:"2024-01-16T14:30:00Z"`
}

// PaymentScreenshotListResponse represents a paginated list of payment screenshots
type PaymentScreenshotListResponse struct {
	Screenshots []*PaymentScreenshotResponse `json:"screenshots"`
	Total       int64                        `json:"total" example:"25"`
	Page        int                          `json:"page" example:"1"`
	PageSize    int                          `json:"page_size" example:"20"`
	TotalPages  int                          `json:"total_pages" example:"2"`
}

// Transformation functions

// ToPaymentScreenshotResponse converts a PaymentScreenshot model to PaymentScreenshotResponse
func ToPaymentScreenshotResponse(screenshot *orderModels.PaymentScreenshot) (*PaymentScreenshotResponse, error) {
	if screenshot == nil {
		return nil, nil
	}

	return &PaymentScreenshotResponse{
		ID:                   screenshot.ID,
		OrderID:              screenshot.OrderID,
		UploadedBy:           screenshot.UploadedBy,
		BuyerOrganizationID:  screenshot.BuyerOrganizationID,
		SellerOrganizationID: screenshot.SellerOrganizationID,
		FileName:             screenshot.FileName,
		FileSizeBytes:        screenshot.FileSizeBytes,
		FileMimeType:         screenshot.FileMimeType,
		PaymentMethod:        string(screenshot.PaymentMethod),
		PaymentDate:          screenshot.PaymentDate,
		AmountPaid:           screenshot.AmountPaid,
		TransactionID:        screenshot.TransactionID,
		VerificationStatus:   string(screenshot.VerificationStatus),
		VerifiedBy:           screenshot.VerifiedBy,
		VerifiedAt:           screenshot.VerifiedAt,
		VerificationNotes:    screenshot.VerificationNotes,
		Description:          screenshot.Description,
		CreatedAt:            screenshot.CreatedAt,
		UpdatedAt:            screenshot.UpdatedAt,
	}, nil
}

// ToPaymentScreenshotResponseList converts a slice of PaymentScreenshot models to PaymentScreenshotResponse slice
func ToPaymentScreenshotResponseList(screenshots []*orderModels.PaymentScreenshot) ([]*PaymentScreenshotResponse, error) {
	if len(screenshots) == 0 {
		return []*PaymentScreenshotResponse{}, nil
	}

	responses := make([]*PaymentScreenshotResponse, len(screenshots))
	for i, screenshot := range screenshots {
		response, err := ToPaymentScreenshotResponse(screenshot)
		if err != nil {
			return nil, err
		}
		responses[i] = response
	}

	return responses, nil
}
