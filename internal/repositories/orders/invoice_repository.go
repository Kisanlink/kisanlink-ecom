package orders

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/orders"
	repositoryCommon "github.com/Kisanlink/kisanlink-ecom/internal/repositories/common"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"github.com/shopspring/decimal"
)

// InvoiceRepository provides database operations for invoices
type InvoiceRepository struct {
	*base.BaseFilterableRepository[*orders.Invoice]
	*repositoryCommon.BaseRepository
	dbManager db.DBManager
}

// NewInvoiceRepository creates a new invoice repository
func NewInvoiceRepository(dbManager db.DBManager) *InvoiceRepository {
	baseRepo := base.NewBaseFilterableRepository[*orders.Invoice]()
	baseRepo.SetDBManager(dbManager)
	return &InvoiceRepository{
		BaseFilterableRepository: baseRepo,
		BaseRepository:           repositoryCommon.NewBaseRepository(dbManager),
		dbManager:                dbManager,
	}
}

// CreateInvoice creates a new invoice with line items in a transaction
func (r *InvoiceRepository) CreateInvoice(ctx context.Context, invoice *orders.Invoice) error {
	// Validate invoice before creation
	if invoice == nil {
		return fmt.Errorf("invoice cannot be nil")
	}
	if invoice.OrderID == "" {
		return fmt.Errorf("order ID is required")
	}
	if invoice.BuyerOrganizationID == "" {
		return fmt.Errorf("buyer organization ID is required")
	}
	if invoice.SellerOrganizationID == "" {
		return fmt.Errorf("seller organization ID is required")
	}
	if len(invoice.LineItems) == 0 {
		return fmt.Errorf("invoice must have at least one line item")
	}

	// Check if database manager supports transactions
	if txManager, ok := r.dbManager.(interface {
		WithTransaction(ctx context.Context, fn func(tx any) error) error
	}); ok {
		return txManager.WithTransaction(ctx, func(tx any) error {
			return r.createInvoiceWithTransaction(ctx, invoice, tx)
		})
	}

	// Fallback for databases without transaction support
	return r.createInvoiceWithoutTransaction(ctx, invoice)
}

// createInvoiceWithTransaction creates invoice and line items within a transaction
func (r *InvoiceRepository) createInvoiceWithTransaction(ctx context.Context, invoice *orders.Invoice, _ any) error {
	// Calculate total before creating
	invoice.CalculateTotal()

	// Set InvoiceID for all line items before creating (GORM associations require this)
	for i := range invoice.LineItems {
		invoice.LineItems[i].InvoiceID = invoice.ID
	}

	// Create the invoice with all its line items via GORM associations
	if err := r.Create(ctx, invoice); err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	return nil
}

// createInvoiceWithoutTransaction creates invoice without transaction support
func (r *InvoiceRepository) createInvoiceWithoutTransaction(ctx context.Context, invoice *orders.Invoice) error {
	// Calculate total before creating
	invoice.CalculateTotal()

	// Set InvoiceID for all line items before creating (GORM associations require this)
	for i := range invoice.LineItems {
		invoice.LineItems[i].InvoiceID = invoice.ID
	}

	// Create the invoice with all its line items via GORM associations
	if err := r.Create(ctx, invoice); err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	return nil
}

// GetInvoiceByID retrieves an invoice by ID with line items
func (r *InvoiceRepository) GetInvoiceByID(ctx context.Context, id string) (*orders.Invoice, error) {
	invoice := &orders.Invoice{}
	retrievedInvoice, err := r.BaseFilterableRepository.GetByID(ctx, id, invoice)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	invoice = retrievedInvoice

	// Load invoice line items
	lineItems, err := r.GetInvoiceLineItems(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load invoice line items: %w", err)
	}
	invoice.LineItems = make([]orders.InvoiceLineItem, len(lineItems))
	for i, item := range lineItems {
		invoice.LineItems[i] = *item
	}

	return invoice, nil
}

// GetInvoiceByNumber retrieves an invoice by invoice number
func (r *InvoiceRepository) GetInvoiceByNumber(ctx context.Context, invoiceNumber string) (*orders.Invoice, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "invoice_number",
			Operator: base.OpEqual,
			Value:    invoiceNumber,
		},
	}

	invoiceList, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(invoiceList) == 0 {
		return nil, fmt.Errorf("invoice not found")
	}

	// Load line items
	invoice := invoiceList[0]
	lineItems, err := r.GetInvoiceLineItems(ctx, invoice.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load invoice line items: %w", err)
	}
	invoice.LineItems = make([]orders.InvoiceLineItem, len(lineItems))
	for i, item := range lineItems {
		invoice.LineItems[i] = *item
	}

	return invoice, nil
}

// GetInvoiceByOrderID retrieves an invoice by order ID
func (r *InvoiceRepository) GetInvoiceByOrderID(ctx context.Context, orderID string) (*orders.Invoice, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "order_id",
			Operator: base.OpEqual,
			Value:    orderID,
		},
	}

	invoiceList, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(invoiceList) == 0 {
		return nil, fmt.Errorf("invoice not found for order %s", orderID)
	}

	// Return the first invoice (most recent)
	invoice := invoiceList[0]

	// Load line items
	lineItems, err := r.GetInvoiceLineItems(ctx, invoice.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load invoice line items: %w", err)
	}
	invoice.LineItems = make([]orders.InvoiceLineItem, len(lineItems))
	for i, item := range lineItems {
		invoice.LineItems[i] = *item
	}

	return invoice, nil
}

// ListInvoices retrieves invoices with filtering and pagination
func (r *InvoiceRepository) ListInvoices(ctx context.Context, filter *orders.InvoiceFilter, offset, limit int) ([]*orders.Invoice, int, error) {
	// Validate input parameters
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if offset < 0 {
		offset = 0
	}

	dbFilter := base.NewFilter()

	// Add conditions based on the invoice filter
	if filter != nil {
		if filter.BuyerID != "" {
			dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
				Field:    "buyer_organization_id",
				Operator: base.OpEqual,
				Value:    filter.BuyerID,
			})
		}

		if filter.SellerID != "" {
			dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
				Field:    "seller_organization_id",
				Operator: base.OpEqual,
				Value:    filter.SellerID,
			})
		}

		if filter.OrderID != "" {
			dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
				Field:    "order_id",
				Operator: base.OpEqual,
				Value:    filter.OrderID,
			})
		}

		if filter.Status != nil {
			dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
				Field:    "status",
				Operator: base.OpEqual,
				Value:    string(*filter.Status),
			})
		}

		if filter.MinAmount != nil {
			dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
				Field:    "total_amount",
				Operator: base.OpGreaterEqual,
				Value:    *filter.MinAmount,
			})
		}

		if filter.MaxAmount != nil {
			dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
				Field:    "total_amount",
				Operator: base.OpLessEqual,
				Value:    *filter.MaxAmount,
			})
		}

		// Date range filtering
		if filter.DateFrom != nil {
			dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
				Field:    "invoice_date",
				Operator: base.OpGreaterEqual,
				Value:    *filter.DateFrom,
			})
		}

		if filter.DateTo != nil {
			dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
				Field:    "invoice_date",
				Operator: base.OpLessEqual,
				Value:    *filter.DateTo,
			})
		}

		// Status-based filters
		if filter.IsPaid != nil && *filter.IsPaid {
			dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
				Field:    "status",
				Operator: base.OpEqual,
				Value:    string(orders.InvoiceStatusPaid),
			})
		}

		// Overdue filter - handled separately after fetching
	}

	// Pagination
	dbFilter.Limit = limit
	dbFilter.Offset = offset

	// Get invoices
	invoiceList, err := r.BaseFilterableRepository.Find(ctx, dbFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find invoices: %w", err)
	}

	// Load line items for each invoice
	for i, invoice := range invoiceList {
		lineItems, err := r.GetInvoiceLineItems(ctx, invoice.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to load line items for invoice %s: %w", invoice.ID, err)
		}
		invoiceList[i].LineItems = make([]orders.InvoiceLineItem, len(lineItems))
		for j, item := range lineItems {
			invoiceList[i].LineItems[j] = *item
		}
	}

	// Apply overdue filter if needed
	if filter != nil && filter.IsOverdue != nil && *filter.IsOverdue {
		var overdueInvoices []*orders.Invoice
		for _, invoice := range invoiceList {
			if invoice.IsOverdue() {
				overdueInvoices = append(overdueInvoices, invoice)
			}
		}
		invoiceList = overdueInvoices
	}

	// Get total count for pagination
	countFilter := *dbFilter
	countFilter.Limit = 0
	countFilter.Offset = 0
	total, err := r.BaseFilterableRepository.CountWithFilter(ctx, &countFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return invoiceList, int(total), nil
}

// GetInvoiceLineItems retrieves line items for a specific invoice
func (r *InvoiceRepository) GetInvoiceLineItems(ctx context.Context, invoiceID string) ([]*orders.InvoiceLineItem, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "invoice_id",
			Operator: base.OpEqual,
			Value:    invoiceID,
		},
	}

	// Apply query options (adds deleted_at IS NULL by default)
	filter = r.ApplyQueryOptions(ctx, filter)

	var lineItemList []*orders.InvoiceLineItem
	if err := r.dbManager.List(ctx, filter, &lineItemList); err != nil {
		return nil, fmt.Errorf("failed to get invoice line items: %w", err)
	}

	return lineItemList, nil
}

// UpdateInvoiceStatus updates the status of an invoice
func (r *InvoiceRepository) UpdateInvoiceStatus(ctx context.Context, id string, status orders.InvoiceStatus) error {
	existingInvoice := &orders.Invoice{}
	retrievedInvoice, err := r.BaseFilterableRepository.GetByID(ctx, id, existingInvoice)
	if err != nil {
		return err
	}
	existingInvoice = retrievedInvoice

	// Validate status transition
	if !existingInvoice.CanTransitionTo(status) {
		return fmt.Errorf("invalid status transition from %s to %s", existingInvoice.Status, status)
	}

	// Update the invoice status
	existingInvoice.Status = status

	// Save the updated invoice
	return r.BaseFilterableRepository.Update(ctx, existingInvoice)
}

// UpdateInvoice updates an invoice
func (r *InvoiceRepository) UpdateInvoice(ctx context.Context, invoice *orders.Invoice) error {
	// Recalculate total before updating
	invoice.CalculateTotal()

	return r.BaseFilterableRepository.Update(ctx, invoice)
}

// DeleteInvoice soft deletes an invoice
func (r *InvoiceRepository) DeleteInvoice(ctx context.Context, id string) error {
	invoice := &orders.Invoice{}
	return r.BaseFilterableRepository.Delete(ctx, id, invoice)
}

// GetInvoicesByBuyer retrieves invoices for a specific buyer
func (r *InvoiceRepository) GetInvoicesByBuyer(ctx context.Context, buyerOrgID string, limit, offset int) ([]*orders.Invoice, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "buyer_organization_id",
			Operator: base.OpEqual,
			Value:    buyerOrgID,
		},
	}
	filter.Limit = limit
	filter.Offset = offset

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetInvoicesBySeller retrieves invoices for a specific seller
func (r *InvoiceRepository) GetInvoicesBySeller(ctx context.Context, sellerOrgID string, limit, offset int) ([]*orders.Invoice, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "seller_organization_id",
			Operator: base.OpEqual,
			Value:    sellerOrgID,
		},
	}
	filter.Limit = limit
	filter.Offset = offset

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetInvoicesByStatus retrieves invoices by status
func (r *InvoiceRepository) GetInvoicesByStatus(ctx context.Context, status orders.InvoiceStatus, limit, offset int) ([]*orders.Invoice, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "status",
			Operator: base.OpEqual,
			Value:    string(status),
		},
	}
	filter.Limit = limit
	filter.Offset = offset

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetInvoicesByDateRange retrieves invoices within a date range
func (r *InvoiceRepository) GetInvoicesByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*orders.Invoice, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "invoice_date",
			Operator: base.OpGreaterEqual,
			Value:    startDate,
		},
		{
			Field:    "invoice_date",
			Operator: base.OpLessEqual,
			Value:    endDate,
		},
	}
	filter.Limit = limit
	filter.Offset = offset

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetOverdueInvoices retrieves all overdue invoices
func (r *InvoiceRepository) GetOverdueInvoices(ctx context.Context, limit, offset int) ([]*orders.Invoice, error) {
	filter := base.NewFilter()

	// Get all unpaid invoices with due dates
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "status",
			Operator: base.OpIn,
			Value:    []string{string(orders.InvoiceStatusDraft), string(orders.InvoiceStatusFinal)},
		},
		{
			Field:    "due_date",
			Operator: base.OpNotEqual,
			Value:    nil,
		},
	}
	filter.Limit = limit
	filter.Offset = offset

	invoices, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Filter for overdue invoices
	var overdueInvoices []*orders.Invoice
	now := time.Now()
	for _, invoice := range invoices {
		if invoice.DueDate != nil && now.After(*invoice.DueDate) {
			overdueInvoices = append(overdueInvoices, invoice)
		}
	}

	return overdueInvoices, nil
}

// GetInvoiceSummary retrieves a summary of an invoice
func (r *InvoiceRepository) GetInvoiceSummary(ctx context.Context, id string) (*orders.InvoiceSummary, error) {
	invoice, err := r.GetInvoiceByID(ctx, id)
	if err != nil {
		return nil, err
	}

	summary := &orders.InvoiceSummary{
		ID:            invoice.ID,
		InvoiceNumber: invoice.InvoiceNumber,
		OrderID:       invoice.OrderID,
		Status:        invoice.Status,
		TotalAmount:   invoice.TotalAmount,
		PaidAmount:    invoice.PaidAmount,
		BalanceDue:    invoice.GetBalanceDue(),
		InvoiceDate:   invoice.InvoiceDate,
		DueDate:       invoice.DueDate,
		IsOverdue:     invoice.IsOverdue(),
		BuyerOrgID:    invoice.BuyerOrganizationID,
		SellerOrgID:   invoice.SellerOrganizationID,
		CreatedAt:     invoice.CreatedAt,
	}

	return summary, nil
}

// GetTotalsByOrganization retrieves total invoice amounts for an organization
func (r *InvoiceRepository) GetTotalsByOrganization(ctx context.Context, orgID string, isBuyer bool) (decimal.Decimal, error) {
	field := "seller_organization_id"
	if isBuyer {
		field = "buyer_organization_id"
	}

	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    field,
			Operator: base.OpEqual,
			Value:    orgID,
		},
		{
			Field:    "status",
			Operator: base.OpNotEqual,
			Value:    string(orders.InvoiceStatusVoided),
		},
	}

	invoices, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return decimal.Zero, err
	}

	total := decimal.Zero
	for _, invoice := range invoices {
		total = total.Add(invoice.TotalAmount)
	}

	return total, nil
}

// MarkAsPaid marks an invoice as paid
func (r *InvoiceRepository) MarkAsPaid(ctx context.Context, invoiceID string, paidAmount decimal.Decimal, paymentMethod string, paymentDate time.Time) error {
	invoice, err := r.GetInvoiceByID(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	if err := invoice.MarkAsPaid(paidAmount, paymentMethod, paymentDate); err != nil {
		return fmt.Errorf("failed to mark invoice as paid: %w", err)
	}

	return r.UpdateInvoice(ctx, invoice)
}

// FinalizeInvoice finalizes an invoice (converts from draft to final)
func (r *InvoiceRepository) FinalizeInvoice(ctx context.Context, invoiceID string) error {
	invoice, err := r.GetInvoiceByID(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	if err := invoice.Finalize(); err != nil {
		return fmt.Errorf("failed to finalize invoice: %w", err)
	}

	return r.UpdateInvoice(ctx, invoice)
}

// VoidInvoice voids an invoice
func (r *InvoiceRepository) VoidInvoice(ctx context.Context, invoiceID string) error {
	invoice, err := r.GetInvoiceByID(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	if err := invoice.Void(); err != nil {
		return fmt.Errorf("failed to void invoice: %w", err)
	}

	return r.UpdateInvoice(ctx, invoice)
}
