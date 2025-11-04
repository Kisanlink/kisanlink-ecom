package orders

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/shopspring/decimal"
)

// InvoiceStatus represents the possible invoice states
type InvoiceStatus string

const (
	InvoiceStatusDraft  InvoiceStatus = "DRAFT"
	InvoiceStatusFinal  InvoiceStatus = "FINAL"
	InvoiceStatusPaid   InvoiceStatus = "PAID"
	InvoiceStatusVoided InvoiceStatus = "VOIDED"
)

// Invoice represents the main invoice entity
type Invoice struct {
	base.BaseModel

	OrderID       string        `json:"order_id" gorm:"type:varchar(255);not null;index:idx_invoice_order"`
	InvoiceNumber string        `json:"invoice_number" gorm:"type:varchar(50);uniqueIndex;not null"` // INV-YYYYMMDD-XXXXX
	InvoiceDate   time.Time     `json:"invoice_date" gorm:"not null;index:idx_invoice_date"`
	DueDate       *time.Time    `json:"due_date" gorm:"index:idx_invoice_due_date"`
	Status        InvoiceStatus `json:"status" gorm:"type:varchar(20);not null;default:'DRAFT';index:idx_invoice_status"`

	// Organization Information
	BuyerOrganizationID  string `json:"buyer_organization_id" gorm:"type:varchar(255);not null;index:idx_invoice_buyer"`
	SellerOrganizationID string `json:"seller_organization_id" gorm:"type:varchar(255);not null;index:idx_invoice_seller"`

	// Financial Information
	SubtotalAmount    decimal.Decimal `json:"subtotal_amount" gorm:"type:decimal(12,2);not null;default:0.00"`
	GSTAmount         decimal.Decimal `json:"gst_amount" gorm:"type:decimal(12,2);not null;default:0.00"`
	PlatformFeeAmount decimal.Decimal `json:"platform_fee_amount" gorm:"type:decimal(12,2);not null;default:0.00"`
	ShippingAmount    decimal.Decimal `json:"shipping_amount" gorm:"type:decimal(12,2);not null;default:0.00"`
	DiscountAmount    decimal.Decimal `json:"discount_amount" gorm:"type:decimal(12,2);not null;default:0.00"`
	TotalAmount       decimal.Decimal `json:"total_amount" gorm:"type:decimal(12,2);not null;default:0.00"`

	// Payment Information
	PaidAmount     decimal.Decimal `json:"paid_amount" gorm:"type:decimal(12,2);not null;default:0.00"`
	PaymentDate    *time.Time      `json:"payment_date"`
	PaymentMethod  string          `json:"payment_method" gorm:"type:varchar(50)"`
	PaymentDetails sql.NullString  `json:"payment_details" gorm:"type:jsonb"` // NULL if not set

	// Additional Information
	Notes      string         `json:"notes" gorm:"type:text"`
	Terms      string         `json:"terms" gorm:"type:text"`
	TaxDetails sql.NullString `json:"tax_details" gorm:"type:jsonb"` // NULL if not set
	Metadata   sql.NullString `json:"metadata" gorm:"type:jsonb"`    // NULL if not set

	// Relationships
	LineItems []InvoiceLineItem `json:"line_items,omitempty" gorm:"foreignKey:InvoiceID;constraint:OnDelete:CASCADE"`
}

// TableName returns the table name for GORM
func (Invoice) TableName() string {
	return "invoices"
}

// NewInvoice creates a new Invoice instance
func NewInvoice(orderID string, buyerOrgID string, sellerOrgID string) *Invoice {
	now := time.Now()
	return &Invoice{
		BaseModel:            *base.NewBaseModel("INV", "large"),
		OrderID:              orderID,
		InvoiceNumber:        generateInvoiceNumber(),
		InvoiceDate:          now,
		Status:               InvoiceStatusDraft,
		BuyerOrganizationID:  buyerOrgID,
		SellerOrganizationID: sellerOrgID,
		SubtotalAmount:       decimal.Zero,
		GSTAmount:            decimal.Zero,
		PlatformFeeAmount:    decimal.Zero,
		ShippingAmount:       decimal.Zero,
		DiscountAmount:       decimal.Zero,
		TotalAmount:          decimal.Zero,
		PaidAmount:           decimal.Zero,
	}
}

// generateInvoiceNumber generates a unique invoice number in format INV-YYYYMMDD-XXXXX
func generateInvoiceNumber() string {
	now := time.Now()
	dateStr := now.Format("20060102")
	timestampSuffix := now.UnixNano() % 100000
	return fmt.Sprintf("INV-%s-%05d", dateStr, timestampSuffix)
}

// InvoiceLineItem represents individual line items in an invoice
type InvoiceLineItem struct {
	base.BaseModel

	InvoiceID   string `json:"invoice_id" gorm:"type:varchar(255);not null;index:idx_invoice_line_invoice"`
	OrderItemID string `json:"order_item_id" gorm:"type:varchar(255);not null"`

	// Product/Service Information
	ProductName string `json:"product_name" gorm:"type:varchar(255);not null"`
	ProductSKU  string `json:"product_sku" gorm:"type:varchar(100)"`
	HSNCode     string `json:"hsn_code" gorm:"type:varchar(20)"` // Harmonized System Nomenclature for tax
	Description string `json:"description" gorm:"type:text"`

	// Quantity and Pricing
	Quantity  decimal.Decimal `json:"quantity" gorm:"type:decimal(10,3);not null"`
	UnitPrice decimal.Decimal `json:"unit_price" gorm:"type:decimal(10,2);not null"`
	LineTotal decimal.Decimal `json:"line_total" gorm:"type:decimal(12,2);not null"`

	// Tax Information
	TaxRate   decimal.Decimal `json:"tax_rate" gorm:"type:decimal(5,4);default:0.0000"`
	GSTAmount decimal.Decimal `json:"gst_amount" gorm:"type:decimal(10,2);default:0.00"`
	CGSTRate  decimal.Decimal `json:"cgst_rate" gorm:"type:decimal(5,4);default:0.0000"` // Central GST
	SGSTRate  decimal.Decimal `json:"sgst_rate" gorm:"type:decimal(5,4);default:0.0000"` // State GST
	IGSTRate  decimal.Decimal `json:"igst_rate" gorm:"type:decimal(5,4);default:0.0000"` // Integrated GST

	// Discount
	DiscountAmount decimal.Decimal `json:"discount_amount" gorm:"type:decimal(10,2);default:0.00"`

	// Metadata
	Metadata sql.NullString `json:"metadata" gorm:"type:jsonb"` // NULL if not set
}

// TableName returns the table name for GORM
func (InvoiceLineItem) TableName() string {
	return "invoice_line_items"
}

// NewInvoiceLineItem creates a new InvoiceLineItem instance
func NewInvoiceLineItem(invoiceID string, orderItemID string, productName string, productSKU string, quantity decimal.Decimal, unitPrice decimal.Decimal) *InvoiceLineItem {
	return &InvoiceLineItem{
		BaseModel:      *base.NewBaseModel("INVLI", "large"),
		InvoiceID:      invoiceID,
		OrderItemID:    orderItemID,
		ProductName:    productName,
		ProductSKU:     productSKU,
		Quantity:       quantity,
		UnitPrice:      unitPrice,
		LineTotal:      quantity.Mul(unitPrice),
		TaxRate:        decimal.Zero,
		GSTAmount:      decimal.Zero,
		CGSTRate:       decimal.Zero,
		SGSTRate:       decimal.Zero,
		IGSTRate:       decimal.Zero,
		DiscountAmount: decimal.Zero,
	}
}

// Invoice methods

// CalculateTotal calculates the total amount for the invoice
func (i *Invoice) CalculateTotal() {
	var subtotal decimal.Decimal
	var totalGST decimal.Decimal
	var totalDiscount decimal.Decimal

	for _, item := range i.LineItems {
		subtotal = subtotal.Add(item.LineTotal)
		totalGST = totalGST.Add(item.GSTAmount)
		totalDiscount = totalDiscount.Add(item.DiscountAmount)
	}

	i.SubtotalAmount = subtotal
	i.GSTAmount = totalGST
	i.DiscountAmount = totalDiscount
	i.TotalAmount = subtotal.Add(i.GSTAmount).Add(i.PlatformFeeAmount).Add(i.ShippingAmount).Sub(i.DiscountAmount)
}

// CanTransitionTo checks if the invoice can transition to the given status
func (i *Invoice) CanTransitionTo(newStatus InvoiceStatus) bool {
	switch i.Status {
	case InvoiceStatusDraft:
		return newStatus == InvoiceStatusFinal || newStatus == InvoiceStatusVoided
	case InvoiceStatusFinal:
		return newStatus == InvoiceStatusPaid || newStatus == InvoiceStatusVoided
	case InvoiceStatusPaid:
		return newStatus == InvoiceStatusVoided
	case InvoiceStatusVoided:
		return false // Terminal state
	default:
		return false
	}
}

// UpdateStatus updates the invoice status
func (i *Invoice) UpdateStatus(newStatus InvoiceStatus) error {
	if !i.CanTransitionTo(newStatus) {
		return fmt.Errorf("invalid status transition from %s to %s", i.Status, newStatus)
	}

	i.Status = newStatus
	return nil
}

// MarkAsPaid marks the invoice as paid
func (i *Invoice) MarkAsPaid(paidAmount decimal.Decimal, paymentMethod string, paymentDate time.Time) error {
	if i.Status != InvoiceStatusFinal {
		return fmt.Errorf("only final invoices can be marked as paid")
	}

	i.PaidAmount = paidAmount
	i.PaymentMethod = paymentMethod
	i.PaymentDate = &paymentDate
	i.Status = InvoiceStatusPaid

	return nil
}

// Finalize finalizes the invoice (converts from draft to final)
func (i *Invoice) Finalize() error {
	if i.Status != InvoiceStatusDraft {
		return fmt.Errorf("only draft invoices can be finalized")
	}

	i.Status = InvoiceStatusFinal
	return nil
}

// Void voids the invoice
func (i *Invoice) Void() error {
	if i.Status == InvoiceStatusVoided {
		return fmt.Errorf("invoice is already voided")
	}

	i.Status = InvoiceStatusVoided
	return nil
}

// IsPaid checks if the invoice is paid
func (i *Invoice) IsPaid() bool {
	return i.Status == InvoiceStatusPaid
}

// IsOverdue checks if the invoice is overdue
func (i *Invoice) IsOverdue() bool {
	if i.Status == InvoiceStatusPaid || i.Status == InvoiceStatusVoided || i.DueDate == nil {
		return false
	}
	return time.Now().After(*i.DueDate)
}

// GetBalanceDue returns the balance due on the invoice
func (i *Invoice) GetBalanceDue() decimal.Decimal {
	if i.IsPaid() {
		return decimal.Zero
	}
	return i.TotalAmount.Sub(i.PaidAmount)
}

// Metadata handling methods

// SetMetadata sets the metadata from a map
func (i *Invoice) SetMetadata(metadata map[string]interface{}) error {
	if metadata == nil {
		i.Metadata = sql.NullString{Valid: false} // Set to NULL
		return nil
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	i.Metadata = sql.NullString{String: string(metadataJSON), Valid: true}
	return nil
}

// GetMetadata returns the metadata as a map
func (i *Invoice) GetMetadata() (map[string]interface{}, error) {
	if !i.Metadata.Valid || i.Metadata.String == "" {
		return nil, nil
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(i.Metadata.String), &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return metadata, nil
}

// SetTaxDetails sets the tax details from a map
func (i *Invoice) SetTaxDetails(taxDetails map[string]interface{}) error {
	if taxDetails == nil {
		i.TaxDetails = sql.NullString{Valid: false} // Set to NULL
		return nil
	}

	taxDetailsJSON, err := json.Marshal(taxDetails)
	if err != nil {
		return fmt.Errorf("failed to marshal tax details: %w", err)
	}

	i.TaxDetails = sql.NullString{String: string(taxDetailsJSON), Valid: true}
	return nil
}

// GetTaxDetails returns the tax details as a map
func (i *Invoice) GetTaxDetails() (map[string]interface{}, error) {
	if !i.TaxDetails.Valid || i.TaxDetails.String == "" {
		return nil, nil
	}

	var taxDetails map[string]interface{}
	if err := json.Unmarshal([]byte(i.TaxDetails.String), &taxDetails); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tax details: %w", err)
	}

	return taxDetails, nil
}

// SetPaymentDetails sets the payment details from a map
func (i *Invoice) SetPaymentDetails(paymentDetails map[string]interface{}) error {
	if paymentDetails == nil {
		i.PaymentDetails = sql.NullString{Valid: false} // Set to NULL
		return nil
	}

	paymentDetailsJSON, err := json.Marshal(paymentDetails)
	if err != nil {
		return fmt.Errorf("failed to marshal payment details: %w", err)
	}

	i.PaymentDetails = sql.NullString{String: string(paymentDetailsJSON), Valid: true}
	return nil
}

// GetPaymentDetails returns the payment details as a map
func (i *Invoice) GetPaymentDetails() (map[string]interface{}, error) {
	if !i.PaymentDetails.Valid || i.PaymentDetails.String == "" {
		return nil, nil
	}

	var paymentDetails map[string]interface{}
	if err := json.Unmarshal([]byte(i.PaymentDetails.String), &paymentDetails); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment details: %w", err)
	}

	return paymentDetails, nil
}

// InvoiceLineItem metadata handling methods

// SetMetadata sets the metadata from a map
func (ili *InvoiceLineItem) SetMetadata(metadata map[string]interface{}) error {
	if metadata == nil {
		ili.Metadata = sql.NullString{Valid: false} // Set to NULL
		return nil
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	ili.Metadata = sql.NullString{String: string(metadataJSON), Valid: true}
	return nil
}

// GetMetadata returns the metadata as a map
func (ili *InvoiceLineItem) GetMetadata() (map[string]interface{}, error) {
	if !ili.Metadata.Valid || ili.Metadata.String == "" {
		return nil, nil
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(ili.Metadata.String), &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return metadata, nil
}

// InvoiceSummary represents a summary view of an invoice
type InvoiceSummary struct {
	ID            string          `json:"id"`
	InvoiceNumber string          `json:"invoice_number"`
	OrderID       string          `json:"order_id"`
	Status        InvoiceStatus   `json:"status"`
	TotalAmount   decimal.Decimal `json:"total_amount"`
	PaidAmount    decimal.Decimal `json:"paid_amount"`
	BalanceDue    decimal.Decimal `json:"balance_due"`
	InvoiceDate   time.Time       `json:"invoice_date"`
	DueDate       *time.Time      `json:"due_date,omitempty"`
	IsOverdue     bool            `json:"is_overdue"`
	BuyerOrgID    string          `json:"buyer_org_id"`
	SellerOrgID   string          `json:"seller_org_id"`
	CreatedAt     time.Time       `json:"created_at"`
}

// InvoiceFilter represents filters for invoice queries
type InvoiceFilter struct {
	Status    *InvoiceStatus   `json:"status,omitempty"`
	BuyerID   string           `json:"buyer_id,omitempty"`
	SellerID  string           `json:"seller_id,omitempty"`
	OrderID   string           `json:"order_id,omitempty"`
	MinAmount *decimal.Decimal `json:"min_amount,omitempty"`
	MaxAmount *decimal.Decimal `json:"max_amount,omitempty"`
	DateFrom  *time.Time       `json:"date_from,omitempty"`
	DateTo    *time.Time       `json:"date_to,omitempty"`
	IsOverdue *bool            `json:"is_overdue,omitempty"`
	IsPaid    *bool            `json:"is_paid,omitempty"`
}
