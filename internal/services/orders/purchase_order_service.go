package orders

import (
	"context"
	"fmt"

	orderModels "kisanlink-ecom/entities/models/orders"
	orderRequests "kisanlink-ecom/entities/requests/orders"
	"kisanlink-ecom/internal/repositories/orders"

	"github.com/shopspring/decimal"
)

// PurchaseOrderServiceInterface defines the interface for purchase order operations
type PurchaseOrderServiceInterface interface {
	CreateManualPO(ctx context.Context, req *orderRequests.CreatePurchaseOrderRequest, userID string, fpoOrgID string) (*orderModels.PurchaseOrder, error)
	CreatePOFromOrder(ctx context.Context, order *orderModels.Order, userID string) (*orderModels.PurchaseOrder, error)
	GetPOByID(ctx context.Context, id string, userID string, fpoOrgID string) (*orderModels.PurchaseOrder, error)
	ListPOs(ctx context.Context, filter *orderRequests.ListPurchaseOrdersRequest, userID string, fpoOrgID string, offset, limit int) ([]*orderModels.PurchaseOrder, int, error)
	UpdatePOStatus(ctx context.Context, id string, req *orderRequests.UpdatePurchaseOrderStatusRequest, userID string, fpoOrgID string) error
	CreateGRN(ctx context.Context, poID string, req *orderRequests.CreateGRNRequest, userID string, fpoOrgID string) (*orderModels.GRN, error)
	GetGRN(ctx context.Context, grnID string, userID string, fpoOrgID string) (*orderModels.GRN, error)
	GetGRNsByPO(ctx context.Context, poID string, userID string, fpoOrgID string) ([]*orderModels.GRN, error)
	UpdateGRNStatus(ctx context.Context, grnID string, req *orderRequests.UpdateGRNStatusRequest, userID string, fpoOrgID string) error
	GeneratePOPDF(ctx context.Context, poID string, userID string, fpoOrgID string) ([]byte, error)
}

// PurchaseOrderService provides business logic for purchase order operations
type PurchaseOrderService struct {
	poRepo *orders.PurchaseOrderRepository
}

// NewPurchaseOrderService creates a new purchase order service
func NewPurchaseOrderService(poRepo *orders.PurchaseOrderRepository) *PurchaseOrderService {
	return &PurchaseOrderService{
		poRepo: poRepo,
	}
}

// CreateManualPO creates a manual purchase order for external vendor
func (s *PurchaseOrderService) CreateManualPO(ctx context.Context, req *orderRequests.CreatePurchaseOrderRequest, userID string, fpoOrgID string) (*orderModels.PurchaseOrder, error) {
	// Validate request
	if err := s.validateCreatePORequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Validate organization ID
	if fpoOrgID == "" {
		return nil, fmt.Errorf("FPO organization ID is required")
	}

	// Create purchase order instance
	po := orderModels.NewPurchaseOrder(fpoOrgID, req.VendorName, orderModels.POSourceManual)
	po.VendorContact = req.VendorContact
	po.Notes = req.Notes

	// Set vendor ID if provided
	if req.VendorID != nil {
		po.VendorID.String = *req.VendorID
		po.VendorID.Valid = true
	}

	// Set delivery address if provided
	if req.DeliveryAddress != nil {
		addr := &orderModels.Address{
			Street:     req.DeliveryAddress.Street,
			City:       req.DeliveryAddress.City,
			State:      req.DeliveryAddress.State,
			PostalCode: req.DeliveryAddress.PostalCode,
			Country:    req.DeliveryAddress.Country,
		}
		if err := po.SetDeliveryAddress(addr); err != nil {
			return nil, fmt.Errorf("failed to set delivery address: %w", err)
		}
	}

	// Process PO items
	for _, itemReq := range req.Items {
		poItem := orderModels.NewPOItem(
			po.ID,
			itemReq.ProductName,
			itemReq.ProductSKU,
			itemReq.HSNCode,
			itemReq.Quantity,
			itemReq.UnitPrice,
			itemReq.GSTPercent,
		)
		po.Items = append(po.Items, *poItem)
	}

	// Calculate total
	po.CalculateTotal()

	// Set metadata if provided
	if req.Metadata != nil {
		if err := po.SetMetadata(req.Metadata); err != nil {
			return nil, fmt.Errorf("failed to set metadata: %w", err)
		}
	}

	// Create the purchase order in database
	if err := s.poRepo.CreatePurchaseOrder(ctx, po); err != nil {
		return nil, fmt.Errorf("failed to create purchase order: %w", err)
	}

	return po, nil
}

// CreatePOFromOrder auto-creates a purchase order from Kisanlink order
func (s *PurchaseOrderService) CreatePOFromOrder(ctx context.Context, order *orderModels.Order, userID string) (*orderModels.PurchaseOrder, error) {
	// Validate order
	if order == nil {
		return nil, fmt.Errorf("order cannot be nil")
	}
	if order.SellerOrganizationID == "" {
		return nil, fmt.Errorf("order must have a seller organization")
	}

	// Create purchase order instance
	po := orderModels.NewPurchaseOrder(order.BuyerOrganizationID, "Kisanlink Marketplace", orderModels.POSourceKisanlink)

	// Link to Kisanlink order
	po.OrderID.String = order.ID
	po.OrderID.Valid = true

	// Set vendor ID as seller organization
	po.VendorID.String = order.SellerOrganizationID
	po.VendorID.Valid = true

	// Copy shipping address if available
	if addr, err := order.GetShippingAddress(); err == nil && addr != nil {
		if err := po.SetDeliveryAddress(addr); err != nil {
			return nil, fmt.Errorf("failed to set delivery address: %w", err)
		}
	}

	// Convert order items to PO items
	for _, orderItem := range order.Items {
		poItem := orderModels.NewPOItem(
			po.ID,
			orderItem.CatalogItemName,
			orderItem.CatalogItemSKU,
			"", // HSN code not available in order items
			orderItem.Quantity,
			orderItem.UnitPrice,
			orderItem.TaxRate.Mul(decimal.NewFromInt(100)), // Convert tax rate to percentage
		)
		po.Items = append(po.Items, *poItem)
	}

	// Calculate total
	po.CalculateTotal()

	// Set metadata with order reference
	metadata := map[string]interface{}{
		"source":       "kisanlink_marketplace",
		"order_id":     order.ID,
		"order_number": order.OrderNumber,
	}
	if err := po.SetMetadata(metadata); err != nil {
		return nil, fmt.Errorf("failed to set metadata: %w", err)
	}

	// Create the purchase order in database
	if err := s.poRepo.CreatePurchaseOrder(ctx, po); err != nil {
		return nil, fmt.Errorf("failed to create purchase order from order: %w", err)
	}

	return po, nil
}

// GetPOByID retrieves a purchase order by ID with authorization checks
func (s *PurchaseOrderService) GetPOByID(ctx context.Context, id string, userID string, fpoOrgID string) (*orderModels.PurchaseOrder, error) {
	// Validate input parameters
	if id == "" {
		return nil, fmt.Errorf("purchase order ID is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if fpoOrgID == "" {
		return nil, fmt.Errorf("FPO organization ID is required")
	}

	// Get the purchase order
	po, err := s.poRepo.GetPurchaseOrderByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase order: %w", err)
	}

	// Validate that the PO belongs to the user's FPO organization
	if po.FPOOrgID != fpoOrgID {
		return nil, fmt.Errorf("purchase order does not belong to organization")
	}

	return po, nil
}

// ListPOs retrieves purchase orders with filtering and pagination
func (s *PurchaseOrderService) ListPOs(ctx context.Context, filter *orderRequests.ListPurchaseOrdersRequest, userID string, fpoOrgID string, offset, limit int) ([]*orderModels.PurchaseOrder, int, error) {
	// Validate input parameters
	if userID == "" {
		return nil, 0, fmt.Errorf("user ID is required")
	}
	if fpoOrgID == "" {
		return nil, 0, fmt.Errorf("FPO organization ID is required")
	}

	// Ensure filter is not nil
	if filter == nil {
		filter = &orderRequests.ListPurchaseOrdersRequest{}
	}

	// Call repository with validated filters
	return s.poRepo.ListPurchaseOrders(ctx, filter, fpoOrgID, offset, limit)
}

// UpdatePOStatus updates the status of a purchase order
func (s *PurchaseOrderService) UpdatePOStatus(ctx context.Context, id string, req *orderRequests.UpdatePurchaseOrderStatusRequest, userID string, fpoOrgID string) error {
	// Validate input parameters
	if id == "" {
		return fmt.Errorf("purchase order ID is required")
	}
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if req.Status == "" {
		return fmt.Errorf("status is required")
	}
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	if fpoOrgID == "" {
		return fmt.Errorf("FPO organization ID is required")
	}

	// Validate permissions - verify PO belongs to user's organization
	po, err := s.poRepo.GetPurchaseOrderByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get purchase order: %w", err)
	}

	if po.FPOOrgID != fpoOrgID {
		return fmt.Errorf("purchase order does not belong to organization")
	}

	// Validate status transition using the model's business logic
	if !po.CanTransitionTo(req.Status) {
		return fmt.Errorf("invalid status transition from %s to %s", po.Status, req.Status)
	}

	// Update the PO status
	if err := s.poRepo.UpdatePurchaseOrderStatus(ctx, id, req.Status, req.Reason); err != nil {
		return fmt.Errorf("failed to update purchase order status: %w", err)
	}

	return nil
}

// CreateGRN creates a goods received note for a purchase order
func (s *PurchaseOrderService) CreateGRN(ctx context.Context, poID string, req *orderRequests.CreateGRNRequest, userID string, fpoOrgID string) (*orderModels.GRN, error) {
	// Validate request
	if poID == "" {
		return nil, fmt.Errorf("purchase order ID is required")
	}
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("GRN must have at least one item")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if fpoOrgID == "" {
		return nil, fmt.Errorf("FPO organization ID is required")
	}

	// Validate permissions - verify PO belongs to user's organization
	po, err := s.poRepo.GetPurchaseOrderByID(ctx, poID)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase order: %w", err)
	}

	if po.FPOOrgID != fpoOrgID {
		return nil, fmt.Errorf("purchase order does not belong to organization")
	}

	// Validate that PO status allows GRN creation
	if po.Status != orderModels.POStatusConfirmed && po.Status != orderModels.POStatusDelivered {
		return nil, fmt.Errorf("purchase order must be in CONFIRMED or DELIVERED status to create GRN")
	}

	// Create GRN instance
	grn := orderModels.NewGRN(poID, userID)
	grn.Notes = req.Notes

	// Get PO items for validation
	poItems, err := s.poRepo.GetPOItems(ctx, poID)
	if err != nil {
		return nil, fmt.Errorf("failed to get PO items: %w", err)
	}

	// Create a map of PO items for quick lookup
	poItemsMap := make(map[string]*orderModels.POItem)
	for _, item := range poItems {
		poItemsMap[item.ID] = item
	}

	// Process GRN items
	for i, itemReq := range req.Items {
		// Validate that PO item exists
		poItem, exists := poItemsMap[itemReq.POItemID]
		if !exists {
			return nil, fmt.Errorf("item %d: PO item '%s' not found", i+1, itemReq.POItemID)
		}

		// Validate received quantity
		if itemReq.QuantityReceived.LessThanOrEqual(decimal.Zero) {
			return nil, fmt.Errorf("item %d: received quantity must be greater than zero", i+1)
		}

		if itemReq.QuantityReceived.GreaterThan(poItem.Quantity) {
			return nil, fmt.Errorf("item %d: received quantity cannot exceed ordered quantity", i+1)
		}

		// Create GRN item
		grnItem := orderModels.NewGRNItem(
			grn.ID,
			itemReq.POItemID,
			poItem.Quantity,
			itemReq.QuantityReceived,
			itemReq.Condition,
		)
		grnItem.Notes = itemReq.Notes
		grn.Items = append(grn.Items, *grnItem)
	}

	// Create the GRN in database
	if err := s.poRepo.CreateGRN(ctx, grn); err != nil {
		return nil, fmt.Errorf("failed to create GRN: %w", err)
	}

	// Auto-update PO status to DELIVERED if all items received
	allItemsReceived := true
	for _, grnItem := range grn.Items {
		if grnItem.QuantityReceived.LessThan(grnItem.QuantityOrdered) || grnItem.Condition != orderModels.ItemConditionGood {
			allItemsReceived = false
			break
		}
	}

	if allItemsReceived && po.Status == orderModels.POStatusConfirmed {
		if err := s.poRepo.UpdatePurchaseOrderStatus(ctx, poID, orderModels.POStatusDelivered, "All items received in good condition"); err != nil {
			// Log error but don't fail GRN creation
			fmt.Printf("Warning: failed to update PO status after GRN creation: %v\n", err)
		}
	}

	return grn, nil
}

// GetGRN retrieves a GRN by ID with authorization checks
func (s *PurchaseOrderService) GetGRN(ctx context.Context, grnID string, userID string, fpoOrgID string) (*orderModels.GRN, error) {
	// Validate input parameters
	if grnID == "" {
		return nil, fmt.Errorf("GRN ID is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if fpoOrgID == "" {
		return nil, fmt.Errorf("FPO organization ID is required")
	}

	// Get the GRN
	grn, err := s.poRepo.GetGRNByID(ctx, grnID)
	if err != nil {
		return nil, fmt.Errorf("failed to get GRN: %w", err)
	}

	// Verify that the GRN's PO belongs to the user's organization
	po, err := s.poRepo.GetPurchaseOrderByID(ctx, grn.POID)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase order: %w", err)
	}

	if po.FPOOrgID != fpoOrgID {
		return nil, fmt.Errorf("GRN does not belong to organization")
	}

	return grn, nil
}

// GetGRNsByPO retrieves all GRNs for a purchase order
func (s *PurchaseOrderService) GetGRNsByPO(ctx context.Context, poID string, userID string, fpoOrgID string) ([]*orderModels.GRN, error) {
	// Validate input parameters
	if poID == "" {
		return nil, fmt.Errorf("purchase order ID is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if fpoOrgID == "" {
		return nil, fmt.Errorf("FPO organization ID is required")
	}

	// Verify that the PO belongs to the user's organization
	po, err := s.poRepo.GetPurchaseOrderByID(ctx, poID)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase order: %w", err)
	}

	if po.FPOOrgID != fpoOrgID {
		return nil, fmt.Errorf("purchase order does not belong to organization")
	}

	// Get GRNs for the PO
	return s.poRepo.GetGRNsByPO(ctx, poID)
}

// UpdateGRNStatus updates the status of a GRN
func (s *PurchaseOrderService) UpdateGRNStatus(ctx context.Context, grnID string, req *orderRequests.UpdateGRNStatusRequest, userID string, fpoOrgID string) error {
	// Validate input parameters
	if grnID == "" {
		return fmt.Errorf("GRN ID is required")
	}
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if req.Status == "" {
		return fmt.Errorf("status is required")
	}
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	if fpoOrgID == "" {
		return fmt.Errorf("FPO organization ID is required")
	}

	// Verify that the GRN's PO belongs to the user's organization
	grn, err := s.poRepo.GetGRNByID(ctx, grnID)
	if err != nil {
		return fmt.Errorf("failed to get GRN: %w", err)
	}

	po, err := s.poRepo.GetPurchaseOrderByID(ctx, grn.POID)
	if err != nil {
		return fmt.Errorf("failed to get purchase order: %w", err)
	}

	if po.FPOOrgID != fpoOrgID {
		return fmt.Errorf("GRN does not belong to organization")
	}

	// Update the GRN status
	if err := s.poRepo.UpdateGRNStatus(ctx, grnID, req.Status, req.Reason); err != nil {
		return fmt.Errorf("failed to update GRN status: %w", err)
	}

	return nil
}

// GeneratePOPDF generates a PDF document for a purchase order
func (s *PurchaseOrderService) GeneratePOPDF(ctx context.Context, poID string, userID string, fpoOrgID string) ([]byte, error) {
	// Validate input parameters
	if poID == "" {
		return nil, fmt.Errorf("purchase order ID is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if fpoOrgID == "" {
		return nil, fmt.Errorf("FPO organization ID is required")
	}

	// Get the purchase order
	po, err := s.GetPOByID(ctx, poID, userID, fpoOrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase order: %w", err)
	}

	// MVP: Return simple text-based PDF content
	// In production, use a proper PDF library like go-pdf or pdfcpu
	pdfContent := fmt.Sprintf(`Purchase Order: %s
FPO Organization: %s
Vendor: %s
Status: %s
Total Amount: %s

Items:
`, po.PONumber, po.FPOOrgID, po.VendorName, po.Status, po.TotalAmount.String())

	for i, item := range po.Items {
		pdfContent += fmt.Sprintf("%d. %s - Qty: %s, Price: %s, Total: %s\n",
			i+1, item.ProductName, item.Quantity.String(), item.UnitPrice.String(), item.TotalAmount.String())
	}

	return []byte(pdfContent), nil
}

// validateCreatePORequest validates the create PO request
func (s *PurchaseOrderService) validateCreatePORequest(req *orderRequests.CreatePurchaseOrderRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.VendorName == "" {
		return fmt.Errorf("vendor name is required")
	}

	if len(req.Items) == 0 {
		return fmt.Errorf("purchase order must contain at least one item")
	}

	// Validate individual items
	for i, item := range req.Items {
		if item.ProductName == "" {
			return fmt.Errorf("product name is required for item %d", i+1)
		}

		if item.Quantity.LessThanOrEqual(decimal.Zero) {
			return fmt.Errorf("item quantity must be greater than 0 for item %d", i+1)
		}

		if item.UnitPrice.LessThan(decimal.Zero) {
			return fmt.Errorf("item unit price cannot be negative for item %d", i+1)
		}

		if item.GSTPercent.LessThan(decimal.Zero) || item.GSTPercent.GreaterThan(decimal.NewFromInt(100)) {
			return fmt.Errorf("GST percent must be between 0 and 100 for item %d", i+1)
		}

		// Validate reasonable quantity limits
		maxQuantity := decimal.NewFromInt(1000000)
		if item.Quantity.GreaterThan(maxQuantity) {
			return fmt.Errorf("item quantity exceeds maximum allowed (1,000,000) for item %d", i+1)
		}

		// Validate reasonable price limits
		maxPrice := decimal.NewFromInt(10000000)
		if item.UnitPrice.GreaterThan(maxPrice) {
			return fmt.Errorf("item unit price exceeds maximum allowed (10,000,000) for item %d", i+1)
		}
	}

	// Validate delivery address if provided
	if req.DeliveryAddress != nil {
		if req.DeliveryAddress.Street == "" {
			return fmt.Errorf("street is required in delivery address")
		}
		if req.DeliveryAddress.City == "" {
			return fmt.Errorf("city is required in delivery address")
		}
		if req.DeliveryAddress.State == "" {
			return fmt.Errorf("state is required in delivery address")
		}
		if req.DeliveryAddress.PostalCode == "" {
			return fmt.Errorf("postal code is required in delivery address")
		}
		if req.DeliveryAddress.Country == "" {
			return fmt.Errorf("country is required in delivery address")
		}
	}

	// Validate notes length
	if len(req.Notes) > 1000 {
		return fmt.Errorf("notes cannot exceed 1000 characters")
	}

	return nil
}
