package orders

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	orderModels "github.com/Kisanlink/kisanlink-ecom/entities/models/orders"
	"github.com/Kisanlink/kisanlink-ecom/internal/repositories/orders"

	"github.com/shopspring/decimal"
)

// InvoiceServiceInterface defines the interface for invoice operations
type InvoiceServiceInterface interface {
	GenerateInvoice(ctx context.Context, orderID string, userID string, orgID string) (*orderModels.Invoice, error)
	GetInvoiceByID(ctx context.Context, invoiceID string, userID string, orgID string) (*orderModels.Invoice, error)
	GetInvoiceByOrderID(ctx context.Context, orderID string, userID string, orgID string) (*orderModels.Invoice, error)
	GetInvoiceByNumber(ctx context.Context, invoiceNumber string, userID string, orgID string) (*orderModels.Invoice, error)
	ListInvoices(ctx context.Context, filter *orderModels.InvoiceFilter, userID string, orgID string, isAdmin bool, offset, limit int) ([]*orderModels.Invoice, int, error)
	GenerateInvoicePDF(ctx context.Context, invoiceID string, userID string, orgID string) ([]byte, error)
	FinalizeInvoice(ctx context.Context, invoiceID string, userID string, orgID string) (*orderModels.Invoice, error)
	MarkInvoiceAsPaid(ctx context.Context, invoiceID string, paidAmount decimal.Decimal, paymentMethod string, userID string, orgID string) (*orderModels.Invoice, error)
	VoidInvoice(ctx context.Context, invoiceID string, userID string, orgID string) (*orderModels.Invoice, error)
	GetInvoiceSummary(ctx context.Context, invoiceID string) (*orderModels.InvoiceSummary, error)
	GetOverdueInvoices(ctx context.Context, orgID string, isBuyer bool, limit, offset int) ([]*orderModels.Invoice, error)
}

// InvoiceService provides business logic for invoice operations
type InvoiceService struct {
	invoiceRepo *orders.InvoiceRepository
	orderRepo   *orders.OrderRepository
}

// NewInvoiceService creates a new invoice service
func NewInvoiceService(invoiceRepo *orders.InvoiceRepository, orderRepo *orders.OrderRepository) *InvoiceService {
	return &InvoiceService{
		invoiceRepo: invoiceRepo,
		orderRepo:   orderRepo,
	}
}

// GenerateInvoice generates an invoice from an order
func (s *InvoiceService) GenerateInvoice(ctx context.Context, orderID string, userID string, orgID string) (*orderModels.Invoice, error) {
	// Validate input parameters
	if orderID == "" {
		return nil, fmt.Errorf("order ID is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	// Get the order
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Validate that the user's organization is either buyer or seller
	if order.BuyerOrganizationID != orgID && order.SellerOrganizationID != orgID {
		return nil, fmt.Errorf("user's organization does not have access to this order")
	}

	// Check if invoice already exists for this order
	existingInvoice, err := s.invoiceRepo.GetInvoiceByOrderID(ctx, orderID)
	if err == nil && existingInvoice != nil {
		return existingInvoice, nil
	}

	// Create new invoice
	invoice := orderModels.NewInvoice(orderID, order.BuyerOrganizationID, order.SellerOrganizationID)

	// Set financial information from order
	invoice.SubtotalAmount = order.SubtotalAmount
	invoice.GSTAmount = order.TaxAmount
	invoice.ShippingAmount = order.ShippingAmount
	invoice.DiscountAmount = order.DiscountAmount
	invoice.TotalAmount = order.TotalAmount

	// Set due date (default: 30 days from invoice date)
	dueDate := time.Now().AddDate(0, 0, 30)
	invoice.DueDate = &dueDate

	// Create invoice line items from order items
	for _, orderItem := range order.Items {
		lineItem := orderModels.NewInvoiceLineItem(
			invoice.ID,
			orderItem.ID,
			orderItem.CatalogItemName,
			orderItem.CatalogItemSKU,
			orderItem.Quantity,
			orderItem.UnitPrice,
		)

		// Set tax information
		lineItem.TaxRate = orderItem.TaxRate
		lineItem.GSTAmount = orderItem.TaxAmount
		lineItem.DiscountAmount = orderItem.DiscountAmount

		// Calculate line total
		lineItem.LineTotal = orderItem.TotalPrice

		invoice.LineItems = append(invoice.LineItems, *lineItem)
	}

	// Calculate totals
	invoice.CalculateTotal()

	// Create the invoice in database
	if err := s.invoiceRepo.CreateInvoice(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	return invoice, nil
}

// GetInvoiceByID retrieves an invoice by ID with authorization checks
func (s *InvoiceService) GetInvoiceByID(ctx context.Context, invoiceID string, userID string, orgID string) (*orderModels.Invoice, error) {
	// Validate input parameters
	if invoiceID == "" {
		return nil, fmt.Errorf("invoice ID is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	// Get the invoice
	invoice, err := s.invoiceRepo.GetInvoiceByID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	// Validate permissions - user can only access invoices from their organization
	if invoice.BuyerOrganizationID != orgID && invoice.SellerOrganizationID != orgID {
		return nil, fmt.Errorf("user's organization does not have access to this invoice")
	}

	return invoice, nil
}

// GetInvoiceByOrderID retrieves an invoice by order ID with authorization checks
func (s *InvoiceService) GetInvoiceByOrderID(ctx context.Context, orderID string, userID string, orgID string) (*orderModels.Invoice, error) {
	// Validate input parameters
	if orderID == "" {
		return nil, fmt.Errorf("order ID is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	// Get the invoice
	invoice, err := s.invoiceRepo.GetInvoiceByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	// Validate permissions - user can only access invoices from their organization
	if invoice.BuyerOrganizationID != orgID && invoice.SellerOrganizationID != orgID {
		return nil, fmt.Errorf("user's organization does not have access to this invoice")
	}

	return invoice, nil
}

// GetInvoiceByNumber retrieves an invoice by invoice number with authorization checks
func (s *InvoiceService) GetInvoiceByNumber(ctx context.Context, invoiceNumber string, userID string, orgID string) (*orderModels.Invoice, error) {
	// Validate input parameters
	if invoiceNumber == "" {
		return nil, fmt.Errorf("invoice number is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	// Get the invoice
	invoice, err := s.invoiceRepo.GetInvoiceByNumber(ctx, invoiceNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	// Validate permissions - user can only access invoices from their organization
	if invoice.BuyerOrganizationID != orgID && invoice.SellerOrganizationID != orgID {
		return nil, fmt.Errorf("user's organization does not have access to this invoice")
	}

	return invoice, nil
}

// ListInvoices retrieves invoices with filtering and pagination
func (s *InvoiceService) ListInvoices(ctx context.Context, filter *orderModels.InvoiceFilter, userID string, orgID string, isAdmin bool, offset, limit int) ([]*orderModels.Invoice, int, error) {
	// Validate input parameters
	if userID == "" {
		return nil, 0, fmt.Errorf("user ID is required")
	}

	// Ensure filter is not nil
	if filter == nil {
		filter = &orderModels.InvoiceFilter{}
	}

	// Apply organization-based filtering for non-admin users
	if !isAdmin {
		if orgID == "" {
			return nil, 0, fmt.Errorf("organization ID is required for non-admin users")
		}

		// Non-admin users can only see invoices from their organization
		// We'll apply both buyer and seller filters - the repository will handle the logic
		// For simplicity, we default to showing invoices where user's org is the buyer
		if filter.BuyerID == "" && filter.SellerID == "" {
			filter.BuyerID = orgID
		} else {
			// Validate that the user can only filter by their own organization
			if filter.BuyerID != "" && filter.BuyerID != orgID {
				return nil, 0, fmt.Errorf("user can only filter invoices for their own organization")
			}
			if filter.SellerID != "" && filter.SellerID != orgID {
				return nil, 0, fmt.Errorf("user can only filter invoices for their own organization")
			}
		}
	}

	// Call repository with validated filters
	return s.invoiceRepo.ListInvoices(ctx, filter, offset, limit)
}

// GenerateInvoicePDF generates a PDF for an invoice
func (s *InvoiceService) GenerateInvoicePDF(ctx context.Context, invoiceID string, userID string, orgID string) ([]byte, error) {
	// Validate input parameters
	if invoiceID == "" {
		return nil, fmt.Errorf("invoice ID is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	// Get the invoice with authorization check
	invoice, err := s.GetInvoiceByID(ctx, invoiceID, userID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	// Generate simple text-based PDF (MVP approach)
	pdfContent := s.generateSimplePDF(invoice)

	return pdfContent, nil
}

// generateSimplePDF generates a simple text-based PDF representation
func (s *InvoiceService) generateSimplePDF(invoice *orderModels.Invoice) []byte {
	var buffer bytes.Buffer

	// PDF Header
	buffer.WriteString("==================================================\n")
	buffer.WriteString("                    INVOICE                       \n")
	buffer.WriteString("==================================================\n\n")

	// Invoice Information
	buffer.WriteString(fmt.Sprintf("Invoice Number: %s\n", invoice.InvoiceNumber))
	buffer.WriteString(fmt.Sprintf("Invoice Date:   %s\n", invoice.InvoiceDate.Format("2006-01-02")))
	if invoice.DueDate != nil {
		buffer.WriteString(fmt.Sprintf("Due Date:       %s\n", invoice.DueDate.Format("2006-01-02")))
	}
	buffer.WriteString(fmt.Sprintf("Status:         %s\n", invoice.Status))
	buffer.WriteString("\n")

	// Organization Information
	buffer.WriteString("--------------------------------------------------\n")
	buffer.WriteString("SELLER INFORMATION\n")
	buffer.WriteString(fmt.Sprintf("Organization ID: %s\n", invoice.SellerOrganizationID))
	buffer.WriteString("\n")
	buffer.WriteString("BUYER INFORMATION\n")
	buffer.WriteString(fmt.Sprintf("Organization ID: %s\n", invoice.BuyerOrganizationID))
	buffer.WriteString("--------------------------------------------------\n\n")

	// Line Items
	buffer.WriteString("LINE ITEMS\n")
	buffer.WriteString("--------------------------------------------------\n")
	buffer.WriteString(fmt.Sprintf("%-30s %8s %12s %12s\n", "Product", "Qty", "Unit Price", "Total"))
	buffer.WriteString("--------------------------------------------------\n")

	for _, item := range invoice.LineItems {
		productName := item.ProductName
		if len(productName) > 30 {
			productName = productName[:27] + "..."
		}
		buffer.WriteString(fmt.Sprintf("%-30s %8s %12s %12s\n",
			productName,
			item.Quantity.StringFixed(2),
			item.UnitPrice.StringFixed(2),
			item.LineTotal.StringFixed(2),
		))
		if item.ProductSKU != "" {
			buffer.WriteString(fmt.Sprintf("  SKU: %s\n", item.ProductSKU))
		}
		if item.HSNCode != "" {
			buffer.WriteString(fmt.Sprintf("  HSN: %s\n", item.HSNCode))
		}
		if !item.GSTAmount.IsZero() {
			buffer.WriteString(fmt.Sprintf("  GST: %s (%.2f%%)\n",
				item.GSTAmount.StringFixed(2),
				item.TaxRate.Mul(decimal.NewFromInt(100)).InexactFloat64(),
			))
		}
	}
	buffer.WriteString("--------------------------------------------------\n\n")

	// Totals
	buffer.WriteString("TOTALS\n")
	buffer.WriteString("--------------------------------------------------\n")
	buffer.WriteString(fmt.Sprintf("Subtotal:         %15s\n", invoice.SubtotalAmount.StringFixed(2)))
	if !invoice.GSTAmount.IsZero() {
		buffer.WriteString(fmt.Sprintf("GST:              %15s\n", invoice.GSTAmount.StringFixed(2)))
	}
	if !invoice.PlatformFeeAmount.IsZero() {
		buffer.WriteString(fmt.Sprintf("Platform Fee:     %15s\n", invoice.PlatformFeeAmount.StringFixed(2)))
	}
	if !invoice.ShippingAmount.IsZero() {
		buffer.WriteString(fmt.Sprintf("Shipping:         %15s\n", invoice.ShippingAmount.StringFixed(2)))
	}
	if !invoice.DiscountAmount.IsZero() {
		buffer.WriteString(fmt.Sprintf("Discount:        -%15s\n", invoice.DiscountAmount.StringFixed(2)))
	}
	buffer.WriteString("--------------------------------------------------\n")
	buffer.WriteString(fmt.Sprintf("TOTAL:            %15s\n", invoice.TotalAmount.StringFixed(2)))
	buffer.WriteString("--------------------------------------------------\n\n")

	// Payment Information
	if invoice.Status == orderModels.InvoiceStatusPaid {
		buffer.WriteString("PAYMENT INFORMATION\n")
		buffer.WriteString("--------------------------------------------------\n")
		buffer.WriteString(fmt.Sprintf("Paid Amount:      %15s\n", invoice.PaidAmount.StringFixed(2)))
		if invoice.PaymentDate != nil {
			buffer.WriteString(fmt.Sprintf("Payment Date:     %s\n", invoice.PaymentDate.Format("2006-01-02")))
		}
		if invoice.PaymentMethod != "" {
			buffer.WriteString(fmt.Sprintf("Payment Method:   %s\n", strings.ToUpper(invoice.PaymentMethod)))
		}
		buffer.WriteString("--------------------------------------------------\n\n")
	} else {
		balanceDue := invoice.GetBalanceDue()
		if !balanceDue.IsZero() {
			buffer.WriteString("PAYMENT DUE\n")
			buffer.WriteString("--------------------------------------------------\n")
			buffer.WriteString(fmt.Sprintf("Balance Due:      %15s\n", balanceDue.StringFixed(2)))
			if invoice.IsOverdue() {
				buffer.WriteString("Status:           OVERDUE\n")
			}
			buffer.WriteString("--------------------------------------------------\n\n")
		}
	}

	// Terms and Notes
	if invoice.Terms != "" {
		buffer.WriteString("TERMS & CONDITIONS\n")
		buffer.WriteString("--------------------------------------------------\n")
		buffer.WriteString(invoice.Terms)
		buffer.WriteString("\n--------------------------------------------------\n\n")
	}

	if invoice.Notes != "" {
		buffer.WriteString("NOTES\n")
		buffer.WriteString("--------------------------------------------------\n")
		buffer.WriteString(invoice.Notes)
		buffer.WriteString("\n--------------------------------------------------\n\n")
	}

	// Footer
	buffer.WriteString("==================================================\n")
	buffer.WriteString("Generated by KisanLink E-commerce Platform\n")
	buffer.WriteString(fmt.Sprintf("Generated on: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	buffer.WriteString("==================================================\n")

	return buffer.Bytes()
}

// FinalizeInvoice finalizes an invoice (converts from draft to final)
func (s *InvoiceService) FinalizeInvoice(ctx context.Context, invoiceID string, userID string, orgID string) (*orderModels.Invoice, error) {
	// Validate input parameters
	if invoiceID == "" {
		return nil, fmt.Errorf("invoice ID is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	// Get the invoice with authorization check
	invoice, err := s.GetInvoiceByID(ctx, invoiceID, userID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	// Only seller can finalize invoice
	if invoice.SellerOrganizationID != orgID {
		return nil, fmt.Errorf("only seller can finalize invoice")
	}

	// Finalize the invoice
	if err := s.invoiceRepo.FinalizeInvoice(ctx, invoiceID); err != nil {
		return nil, fmt.Errorf("failed to finalize invoice: %w", err)
	}

	// Return updated invoice
	return s.invoiceRepo.GetInvoiceByID(ctx, invoiceID)
}

// MarkInvoiceAsPaid marks an invoice as paid
func (s *InvoiceService) MarkInvoiceAsPaid(ctx context.Context, invoiceID string, paidAmount decimal.Decimal, paymentMethod string, userID string, orgID string) (*orderModels.Invoice, error) {
	// Validate input parameters
	if invoiceID == "" {
		return nil, fmt.Errorf("invoice ID is required")
	}
	if paidAmount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("paid amount must be greater than zero")
	}
	if paymentMethod == "" {
		return nil, fmt.Errorf("payment method is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	// Get the invoice with authorization check
	invoice, err := s.GetInvoiceByID(ctx, invoiceID, userID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	// Only buyer can mark invoice as paid
	if invoice.BuyerOrganizationID != orgID {
		return nil, fmt.Errorf("only buyer can mark invoice as paid")
	}

	// Mark as paid
	paymentDate := time.Now()
	if err := s.invoiceRepo.MarkAsPaid(ctx, invoiceID, paidAmount, paymentMethod, paymentDate); err != nil {
		return nil, fmt.Errorf("failed to mark invoice as paid: %w", err)
	}

	// Return updated invoice
	return s.invoiceRepo.GetInvoiceByID(ctx, invoiceID)
}

// VoidInvoice voids an invoice
func (s *InvoiceService) VoidInvoice(ctx context.Context, invoiceID string, userID string, orgID string) (*orderModels.Invoice, error) {
	// Validate input parameters
	if invoiceID == "" {
		return nil, fmt.Errorf("invoice ID is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	// Get the invoice with authorization check
	invoice, err := s.GetInvoiceByID(ctx, invoiceID, userID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	// Only seller can void invoice
	if invoice.SellerOrganizationID != orgID {
		return nil, fmt.Errorf("only seller can void invoice")
	}

	// Void the invoice
	if err := s.invoiceRepo.VoidInvoice(ctx, invoiceID); err != nil {
		return nil, fmt.Errorf("failed to void invoice: %w", err)
	}

	// Return updated invoice
	return s.invoiceRepo.GetInvoiceByID(ctx, invoiceID)
}

// GetInvoiceSummary retrieves a summary of an invoice
func (s *InvoiceService) GetInvoiceSummary(ctx context.Context, invoiceID string) (*orderModels.InvoiceSummary, error) {
	return s.invoiceRepo.GetInvoiceSummary(ctx, invoiceID)
}

// GetOverdueInvoices retrieves overdue invoices for an organization
func (s *InvoiceService) GetOverdueInvoices(ctx context.Context, orgID string, isBuyer bool, limit, offset int) ([]*orderModels.Invoice, error) {
	// Validate input parameters
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	// Get all overdue invoices
	overdueInvoices, err := s.invoiceRepo.GetOverdueInvoices(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue invoices: %w", err)
	}

	// Filter by organization
	var filtered []*orderModels.Invoice
	for _, invoice := range overdueInvoices {
		if isBuyer && invoice.BuyerOrganizationID == orgID {
			filtered = append(filtered, invoice)
		} else if !isBuyer && invoice.SellerOrganizationID == orgID {
			filtered = append(filtered, invoice)
		}
	}

	return filtered, nil
}
