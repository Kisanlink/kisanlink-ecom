package orders

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/shopspring/decimal"
)

// VerificationStatus represents the verification state of a payment screenshot
type VerificationStatus string

const (
	VerificationStatusPending  VerificationStatus = "pending"
	VerificationStatusApproved VerificationStatus = "approved"
	VerificationStatusRejected VerificationStatus = "rejected"
	VerificationStatusDisputed VerificationStatus = "disputed"
)

// PaymentMethod represents the method used for payment
type PaymentMethod string

const (
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	PaymentMethodUPI          PaymentMethod = "upi"
	PaymentMethodCheque       PaymentMethod = "cheque"
	PaymentMethodCash         PaymentMethod = "cash"
	PaymentMethodOther        PaymentMethod = "other"
)

// PaymentScreenshot represents a payment proof uploaded by a buyer
type PaymentScreenshot struct {
	base.BaseModel

	// References
	OrderID              string `json:"order_id" gorm:"type:varchar(255);not null;index"`
	UploadedBy           string `json:"uploaded_by" gorm:"type:varchar(255);not null"`
	BuyerOrganizationID  string `json:"buyer_organization_id" gorm:"type:varchar(255);not null;index"`
	SellerOrganizationID string `json:"seller_organization_id" gorm:"type:varchar(255);not null;index"`

	// File Information
	FileName      string `json:"file_name" gorm:"type:varchar(255);not null"`
	FilePath      string `json:"file_path" gorm:"type:varchar(500);not null"`
	S3Bucket      string `json:"s3_bucket" gorm:"type:varchar(255);not null"`
	FileSizeBytes int64  `json:"file_size_bytes" gorm:"not null"`
	FileMimeType  string `json:"file_mime_type" gorm:"type:varchar(100);not null"`
	FileChecksum  string `json:"file_checksum" gorm:"type:varchar(64)"`

	// Payment Information
	PaymentMethod PaymentMethod   `json:"payment_method" gorm:"type:varchar(50)"`
	PaymentDate   *time.Time      `json:"payment_date"`
	AmountPaid    decimal.Decimal `json:"amount_paid" gorm:"type:decimal(12,2)"`
	TransactionID string          `json:"transaction_id" gorm:"type:varchar(100)"`

	// Verification Status
	VerificationStatus VerificationStatus `json:"verification_status" gorm:"type:varchar(20);not null;default:'pending';index"`
	VerifiedBy         *string            `json:"verified_by" gorm:"type:varchar(255)"`
	VerifiedAt         *time.Time         `json:"verified_at"`
	VerificationNotes  string             `json:"verification_notes" gorm:"type:text"`

	// Description
	Description string `json:"description" gorm:"type:text"`

	// Audit fields
	CreatedBy string `json:"created_by" gorm:"type:varchar(255);not null"`
	UpdatedBy string `json:"updated_by" gorm:"type:varchar(255);not null"`
}

// TableName returns the table name for GORM
func (PaymentScreenshot) TableName() string {
	return "payment_screenshots"
}

// IsApproved checks if the screenshot has been approved
func (ps *PaymentScreenshot) IsApproved() bool {
	return ps.VerificationStatus == VerificationStatusApproved
}

// IsRejected checks if the screenshot has been rejected
func (ps *PaymentScreenshot) IsRejected() bool {
	return ps.VerificationStatus == VerificationStatusRejected
}

// IsPending checks if the screenshot is pending verification
func (ps *PaymentScreenshot) IsPending() bool {
	return ps.VerificationStatus == VerificationStatusPending
}

// IsDisputed checks if the screenshot is disputed
func (ps *PaymentScreenshot) IsDisputed() bool {
	return ps.VerificationStatus == VerificationStatusDisputed
}

// CanVerify checks if the screenshot can be verified (is pending or disputed)
func (ps *PaymentScreenshot) CanVerify() bool {
	return ps.IsPending() || ps.IsDisputed()
}

// NewPaymentScreenshot creates a new PaymentScreenshot instance
func NewPaymentScreenshot(orderID, uploadedBy, buyerOrgID, sellerOrgID string) *PaymentScreenshot {
	return &PaymentScreenshot{
		BaseModel:            *base.NewBaseModel("PSCR", "large"),
		OrderID:              orderID,
		UploadedBy:           uploadedBy,
		BuyerOrganizationID:  buyerOrgID,
		SellerOrganizationID: sellerOrgID,
		VerificationStatus:   VerificationStatusPending,
		CreatedBy:            uploadedBy,
		UpdatedBy:            uploadedBy,
	}
}
