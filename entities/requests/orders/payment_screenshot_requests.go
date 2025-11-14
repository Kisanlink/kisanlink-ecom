package orders

import (
	"github.com/Kisanlink/kisanlink-ecom/entities/models/orders"

	"github.com/shopspring/decimal"
)

// UploadPaymentScreenshotRequest represents the request to upload a payment screenshot
type UploadPaymentScreenshotRequest struct {
	OrderID       string          `form:"order_id" binding:"required" validate:"required,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
	PaymentMethod string          `form:"payment_method" binding:"required,oneof=bank_transfer upi cheque cash other" validate:"required,oneof=bank_transfer upi cheque cash other" example:"bank_transfer"`
	PaymentDate   string          `form:"payment_date" binding:"required,datetime=2006-01-02" validate:"required,datetime=2006-01-02" example:"2024-01-15"`
	AmountPaid    decimal.Decimal `form:"amount_paid" binding:"required" validate:"required,gt=0" example:"500.00"`
	TransactionID string          `form:"transaction_id" binding:"omitempty,max=100" validate:"omitempty,max=100" example:"TXN123456789"`
	Description   string          `form:"description" binding:"omitempty,max=1000" validate:"omitempty,max=1000" example:"Payment for order ORD-2024-001"`
	File          string          `form:"-" swaggerignore:"true"` // The file will be handled separately via multipart
}

// VerifyPaymentScreenshotRequest represents the request to verify a payment screenshot
type VerifyPaymentScreenshotRequest struct {
	Status orders.VerificationStatus `json:"status" binding:"required,oneof=approved rejected disputed" validate:"required,oneof=approved rejected disputed" example:"approved"`
	Notes  string                    `json:"notes" binding:"required,min=1,max=1000" validate:"required,min=1,max=1000" example:"Payment verified successfully"`
}

// ListPaymentScreenshotsRequest represents the request to list payment screenshots
type ListPaymentScreenshotsRequest struct {
	// Pagination
	Page     int `form:"page" binding:"omitempty,min=1" validate:"omitempty,min=1" example:"1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100" validate:"omitempty,min=1,max=100" example:"20"`

	// Filtering
	OrderID              *string                    `form:"order_id" validate:"omitempty,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
	VerificationStatus   *orders.VerificationStatus `form:"verification_status" binding:"omitempty,oneof=pending approved rejected disputed" validate:"omitempty,oneof=pending approved rejected disputed" example:"pending"`
	BuyerOrganizationID  *string                    `form:"buyer_organization_id" validate:"omitempty,uuid4" example:"123e4567-e89b-12d3-a456-426614174001"`
	SellerOrganizationID *string                    `form:"seller_organization_id" validate:"omitempty,uuid4" example:"123e4567-e89b-12d3-a456-426614174002"`
	UploadedBy           *string                    `form:"uploaded_by" validate:"omitempty,uuid4" example:"123e4567-e89b-12d3-a456-426614174003"`
	PaymentMethod        *orders.PaymentMethod      `form:"payment_method" binding:"omitempty,oneof=bank_transfer upi cheque cash other" validate:"omitempty,oneof=bank_transfer upi cheque cash other" example:"bank_transfer"`

	// Date range
	CreatedAfter  *string `form:"created_after" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00" example:"2024-01-01T00:00:00Z"`
	CreatedBefore *string `form:"created_before" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00" example:"2024-12-31T23:59:59Z"`

	// Sorting
	SortBy    *string `form:"sort_by" binding:"omitempty,oneof=created_at updated_at amount_paid payment_date" validate:"omitempty,oneof=created_at updated_at amount_paid payment_date" example:"created_at"`
	SortOrder *string `form:"sort_order" binding:"omitempty,oneof=asc desc" validate:"omitempty,oneof=asc desc" example:"desc"`
}

// PaymentScreenshotFilters represents filters for payment screenshot queries (internal use)
type PaymentScreenshotFilters struct {
	OrderID              *string
	VerificationStatus   *orders.VerificationStatus
	BuyerOrganizationID  *string
	SellerOrganizationID *string
	UploadedBy           *string
	PaymentMethod        *orders.PaymentMethod
	CreatedAfter         *string
	CreatedBefore        *string
	Page                 int
	PageSize             int
	SortBy               string
	SortOrder            string
}

// ToPaymentScreenshotFilters converts ListPaymentScreenshotsRequest to PaymentScreenshotFilters
func (r *ListPaymentScreenshotsRequest) ToPaymentScreenshotFilters() *PaymentScreenshotFilters {
	filters := &PaymentScreenshotFilters{
		OrderID:              r.OrderID,
		VerificationStatus:   r.VerificationStatus,
		BuyerOrganizationID:  r.BuyerOrganizationID,
		SellerOrganizationID: r.SellerOrganizationID,
		UploadedBy:           r.UploadedBy,
		PaymentMethod:        r.PaymentMethod,
		CreatedAfter:         r.CreatedAfter,
		CreatedBefore:        r.CreatedBefore,
		Page:                 r.Page,
		PageSize:             r.PageSize,
	}

	// Set defaults
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.PageSize <= 0 {
		filters.PageSize = 20
	}

	// Set sorting
	if r.SortBy != nil {
		filters.SortBy = *r.SortBy
	} else {
		filters.SortBy = "created_at"
	}

	if r.SortOrder != nil {
		filters.SortOrder = *r.SortOrder
	} else {
		filters.SortOrder = "desc"
	}

	return filters
}
