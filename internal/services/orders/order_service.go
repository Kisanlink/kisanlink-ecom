package orders

import (
	"context"
	"fmt"
	"time"

	orderModels "kisanlink-ecom/entities/models/orders"
	orderRequests "kisanlink-ecom/entities/requests/orders"
	"kisanlink-ecom/internal/repositories/orders"
	"kisanlink-ecom/internal/services/catalog"
	"kisanlink-ecom/internal/services/inventory"

	"github.com/shopspring/decimal"
)

// OrderServiceInterface defines the interface for order operations
type OrderServiceInterface interface {
	CreateOrder(ctx context.Context, req *orderRequests.CreateOrderRequest, userID string, orgID string) (*orderModels.Order, error)
	CreateOrderFromBid(ctx context.Context, req *orderRequests.CreateOrderFromBidRequest, userID string, orgID string) (*orderModels.Order, error)
	ValidateBidForOrder(ctx context.Context, bidID string, userID string, orgID string) (*BidOrderValidation, error)
	GetOrderByID(ctx context.Context, id string, userID string, orgID string) (*orderModels.Order, error)
	GetOrderByNumber(ctx context.Context, orderNumber string) (*orderModels.Order, error)
	UpdateOrder(ctx context.Context, id string, req *orderRequests.UpdateOrderRequest, userID string, orgID string) (*orderModels.Order, error)
	UpdateOrderStatus(ctx context.Context, id string, req *orderRequests.UpdateOrderStatusRequest, userID string, orgID string) error
	ListOrders(ctx context.Context, filter *orderRequests.ListOrdersRequest, userID string, orgID string, offset, limit int) ([]*orderModels.Order, int, error)
	GetOrdersByBuyer(ctx context.Context, buyerID string, limit, offset int) ([]*orderModels.Order, error)
	GetOrdersBySeller(ctx context.Context, sellerID string, limit, offset int) ([]*orderModels.Order, error)
	GetOrderSummary(ctx context.Context, id string) (*orderModels.OrderSummary, error)
	CancelOrder(ctx context.Context, id string, userID string, orgID string) error
	FulfillOrder(ctx context.Context, id string, userID string, orgID string) error
	GetOrderAnalytics(ctx context.Context, orgID string, startDate, endDate time.Time) (*OrderAnalytics, error)
	ValidateOrderPermissions(ctx context.Context, orderID string, userID string, orgID string, action string) error
	ProcessPaymentForOrder(ctx context.Context, orderID string, paymentMethod string, userID string, orgID string) (*PaymentResult, error)
	GetPaymentStatus(ctx context.Context, orderID string, userID string, orgID string) (*PaymentResult, error)
	HandlePaymentFailure(ctx context.Context, orderID string, reason string, userID string, orgID string) error
}

// SequenceServiceInterface defines the interface for sequence operations
type SequenceServiceInterface interface {
	GenerateID(ctx context.Context, prefix string, orgID *string) (string, error)
}

// OrderService provides business logic for order operations
type OrderService struct {
	orderRepo      *orders.OrderRepository
	catalogSvc     catalog.CatalogServiceInterface
	inventorySvc   inventory.InventoryService
	sequenceSvc    SequenceServiceInterface
	marketplaceSvc MarketplaceServiceInterface
}

// MarketplaceServiceInterface defines the interface for marketplace operations needed by order service
type MarketplaceServiceInterface interface {
	GetBid(ctx context.Context, bidID string, viewerID string) (interface{}, error)
	GetListing(ctx context.Context, listingID string, viewerID string, viewerOrgID string) (interface{}, error)
}

// NewOrderService creates a new order service
func NewOrderService(orderRepo *orders.OrderRepository, catalogSvc catalog.CatalogServiceInterface, inventorySvc inventory.InventoryService, sequenceSvc SequenceServiceInterface) *OrderService {
	return &OrderService{
		orderRepo:    orderRepo,
		catalogSvc:   catalogSvc,
		inventorySvc: inventorySvc,
		sequenceSvc:  sequenceSvc,
	}
}

// SetMarketplaceService sets the marketplace service dependency
func (s *OrderService) SetMarketplaceService(marketplaceSvc MarketplaceServiceInterface) {
	s.marketplaceSvc = marketplaceSvc
}

// CreateOrder creates a new order with business logic validation
func (s *OrderService) CreateOrder(ctx context.Context, req *orderRequests.CreateOrderRequest, userID string, orgID string) (*orderModels.Order, error) {
	// Validate request
	if err := s.validateCreateOrderRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Validate organization ID
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	// Validate that the user's organization matches the buyer organization
	if req.BuyerOrganizationID != orgID {
		return nil, fmt.Errorf("user can only create orders for their own organization")
	}

	// Validate order items and check inventory availability
	if err := s.validateOrderItems(ctx, req.Items, req.SellerOrganizationID, userID); err != nil {
		return nil, fmt.Errorf("order item validation failed: %w", err)
	}

	// Create order instance
	ord := orderModels.NewOrder(req.BuyerOrganizationID, req.SellerOrganizationID, userID)
	ord.Notes = req.Notes

	// Process order items with proper validation and total calculation
	var totalAmount decimal.Decimal
	for _, itemReq := range req.Items {
		// Get catalog item details for validation and pricing
		catalogItem, err := s.catalogSvc.GetCatalogItemByID(ctx, itemReq.CatalogItemID)
		if err != nil {
			return nil, fmt.Errorf("failed to get catalog item %s: %w", itemReq.CatalogItemID, err)
		}

		// Validate that the catalog item belongs to the seller organization
		if catalogItem.OrganizationID != req.SellerOrganizationID {
			return nil, fmt.Errorf("catalog item %s does not belong to seller organization %s", itemReq.CatalogItemID, req.SellerOrganizationID)
		}

		// Validate pricing - use catalog price if unit price is not provided or is zero
		unitPrice := itemReq.UnitPrice
		if unitPrice.IsZero() {
			unitPrice = catalogItem.BasePrice
		}

		// Generate unique ID for order item using sequence service
		itemID, err := s.sequenceSvc.GenerateID(ctx, "ITEM", &req.SellerOrganizationID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate order item ID: %w", err)
		}

		// Create order item with proper calculations
		quantity := itemReq.Quantity
		orderItem := orderModels.NewOrderItem(
			itemID, // Use sequence-based ID instead of timestamp hash
			ord.ID,
			itemReq.CatalogItemID,
			itemReq.CatalogItemType,
			catalogItem.Name,
			catalogItem.SKU,
			quantity,
			unitPrice,
		)

		// Calculate item total (quantity * unit price)
		orderItem.TotalPrice = quantity.Mul(unitPrice)
		totalAmount = totalAmount.Add(orderItem.TotalPrice)

		ord.Items = append(ord.Items, *orderItem)
	}

	// Set calculated totals
	ord.SubtotalAmount = totalAmount
	ord.TotalAmount = totalAmount.Add(ord.TaxAmount).Add(ord.ShippingAmount).Sub(ord.DiscountAmount)

	// Reserve inventory for products before creating the order
	if err := s.reserveInventoryForOrder(ctx, req.Items, userID, req.SellerOrganizationID); err != nil {
		return nil, fmt.Errorf("failed to reserve inventory: %w", err)
	}

	// Create the order in database
	if err := s.orderRepo.CreateOrder(ctx, ord); err != nil {
		// Release inventory if order creation fails
		if releaseErr := s.releaseInventoryForOrder(ctx, req.Items, userID, req.SellerOrganizationID); releaseErr != nil {
			// Log the release error but don't mask the original error
			fmt.Printf("Warning: failed to release inventory after order creation failure: %v\n", releaseErr)
		}
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return ord, nil
}

// GetOrderByID retrieves an order by ID with authorization checks
func (s *OrderService) GetOrderByID(ctx context.Context, id string, userID string, orgID string) (*orderModels.Order, error) {
	// Validate input parameters
	if id == "" {
		return nil, fmt.Errorf("order ID is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	// Get the order
	order, err := s.orderRepo.GetOrderByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Validate permissions - user can only access orders from their organization
	if err := s.ValidateOrderPermissions(ctx, id, userID, orgID, "read"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	return order, nil
}

// GetOrderByNumber retrieves an order by order number
func (s *OrderService) GetOrderByNumber(ctx context.Context, orderNumber string) (*orderModels.Order, error) {
	return s.orderRepo.GetOrderByNumber(ctx, orderNumber)
}

// UpdateOrder updates an order with business logic validation
func (s *OrderService) UpdateOrder(ctx context.Context, id string, req *orderRequests.UpdateOrderRequest, userID string, orgID string) (*orderModels.Order, error) {
	// Validate input parameters
	if id == "" {
		return nil, fmt.Errorf("order ID is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	// Validate request
	if err := s.validateUpdateOrderRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Validate permissions
	if err := s.ValidateOrderPermissions(ctx, id, userID, orgID, "update"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Get the existing order
	ord, err := s.orderRepo.GetOrderByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Validate that organization changes are not allowed for security
	if req.BuyerOrganizationID != nil && *req.BuyerOrganizationID != ord.BuyerOrganizationID {
		return nil, fmt.Errorf("buyer organization cannot be changed after order creation")
	}
	if req.SellerOrganizationID != nil && *req.SellerOrganizationID != ord.SellerOrganizationID {
		return nil, fmt.Errorf("seller organization cannot be changed after order creation")
	}

	// Update order fields if provided
	if req.Notes != nil {
		ord.Notes = *req.Notes
	}

	// Update status if provided
	if req.Status != nil {
		reason := "Order updated"
		if req.Reason != nil {
			reason = *req.Reason
		}
		if err := s.UpdateOrderStatus(ctx, id, &orderRequests.UpdateOrderStatusRequest{
			Status: *req.Status,
			Reason: reason,
		}, userID, orgID); err != nil {
			return nil, fmt.Errorf("failed to update order status: %w", err)
		}
	}

	// Update the order in the repository
	if err := s.orderRepo.Update(ctx, ord); err != nil {
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	return ord, nil
}

// UpdateOrderStatus updates the status of an order with business logic validation
func (s *OrderService) UpdateOrderStatus(ctx context.Context, id string, req *orderRequests.UpdateOrderStatusRequest, userID string, orgID string) error {
	// Validate input parameters
	if id == "" {
		return fmt.Errorf("order ID is required")
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
	if orgID == "" {
		return fmt.Errorf("organization ID is required")
	}

	// Validate permissions - user can only update orders from their organization
	if err := s.ValidateOrderPermissions(ctx, id, userID, orgID, "update"); err != nil {
		return fmt.Errorf("permission denied: %w", err)
	}

	// Get the current order to validate status transition
	order, err := s.orderRepo.GetOrderByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Validate status transition using the order model's business logic
	if !order.CanTransitionTo(req.Status) {
		return fmt.Errorf("invalid status transition from %s to %s", order.Status, req.Status)
	}

	// Handle inventory release for cancelled orders
	if req.Status == orderModels.OrderStatusCancelled && order.Status != orderModels.OrderStatusCancelled {
		// Convert order items to the format expected by ReleaseInventory
		items := make([]orderRequests.CreateOrderItemRequest, len(order.Items))
		for i, item := range order.Items {
			items[i] = orderRequests.CreateOrderItemRequest{
				CatalogItemID:   item.CatalogItemID,
				CatalogItemType: item.CatalogItemType,
				Quantity:        item.Quantity,
				UnitPrice:       item.UnitPrice,
			}
		}

		// Release inventory
		if err := s.releaseInventoryForOrder(ctx, items, userID, order.SellerOrganizationID); err != nil {
			// Log error but don't fail the status update
			fmt.Printf("Warning: failed to release inventory for cancelled order %s: %v\n", id, err)
		}
	}

	// Handle inventory sale for completed orders
	if req.Status == orderModels.OrderStatusCompleted && order.Status != orderModels.OrderStatusCompleted {
		// Convert order items to the format expected by SellInventory
		items := make([]orderRequests.CreateOrderItemRequest, len(order.Items))
		for i, item := range order.Items {
			items[i] = orderRequests.CreateOrderItemRequest{
				CatalogItemID:   item.CatalogItemID,
				CatalogItemType: item.CatalogItemType,
				Quantity:        item.Quantity,
				UnitPrice:       item.UnitPrice,
			}
		}

		// Sell inventory (convert reserved to sold)
		if err := s.sellInventoryForOrder(ctx, items, userID, order.SellerOrganizationID); err != nil {
			// Log error but don't fail the status update
			fmt.Printf("Warning: failed to sell inventory for completed order %s: %v\n", id, err)
		}
	}

	// Update the order status with proper history tracking
	if err := s.orderRepo.UpdateOrderStatus(ctx, id, req.Status, nil, req.Reason); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

// ListOrders retrieves orders with filtering and pagination
func (s *OrderService) ListOrders(ctx context.Context, filter *orderRequests.ListOrdersRequest, userID string, orgID string, offset, limit int) ([]*orderModels.Order, int, error) {
	// Validate input parameters
	if userID == "" {
		return nil, 0, fmt.Errorf("user ID is required")
	}
	if orgID == "" {
		return nil, 0, fmt.Errorf("organization ID is required")
	}

	// Ensure filter is not nil
	if filter == nil {
		filter = &orderRequests.ListOrdersRequest{}
	}

	// Apply organization-based filtering - users can only see orders from their organization
	// Either as buyer or seller
	if filter.BuyerOrganizationID == nil && filter.SellerOrganizationID == nil {
		// If no organization filter is specified, show orders where user's org is buyer or seller
		filter.BuyerOrganizationID = &orgID
		// Note: We could also create a more complex filter to include both buyer and seller,
		// but for now we'll default to showing orders where the user's org is the buyer
	} else {
		// Validate that the user can only filter by their own organization
		if filter.BuyerOrganizationID != nil && *filter.BuyerOrganizationID != orgID {
			return nil, 0, fmt.Errorf("user can only filter orders for their own organization")
		}
		if filter.SellerOrganizationID != nil && *filter.SellerOrganizationID != orgID {
			return nil, 0, fmt.Errorf("user can only filter orders for their own organization")
		}
	}

	// Call repository with validated filters
	return s.orderRepo.ListOrders(ctx, filter, offset, limit)
}

// GetOrdersByBuyer retrieves orders for a specific buyer
func (s *OrderService) GetOrdersByBuyer(ctx context.Context, buyerID string, limit, offset int) ([]*orderModels.Order, error) {
	return s.orderRepo.GetOrdersByBuyer(ctx, buyerID, limit, offset)
}

// GetOrdersBySeller retrieves orders for a specific seller
func (s *OrderService) GetOrdersBySeller(ctx context.Context, sellerID string, limit, offset int) ([]*orderModels.Order, error) {
	return s.orderRepo.GetOrdersBySeller(ctx, sellerID, limit, offset)
}

// GetOrderSummary retrieves a summary of an order
func (s *OrderService) GetOrderSummary(ctx context.Context, id string) (*orderModels.OrderSummary, error) {
	return s.orderRepo.GetOrderSummary(ctx, id)
}

// CancelOrder cancels an order and releases inventory
func (s *OrderService) CancelOrder(ctx context.Context, id string, userID string, orgID string) error {
	// Validate input parameters
	if id == "" {
		return fmt.Errorf("order ID is required")
	}
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	if orgID == "" {
		return fmt.Errorf("organization ID is required")
	}

	// Validate permissions
	if err := s.ValidateOrderPermissions(ctx, id, userID, orgID, "cancel"); err != nil {
		return fmt.Errorf("permission denied: %w", err)
	}

	// Get the order
	ord, err := s.orderRepo.GetOrderByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Check if order can be cancelled
	if !ord.CanTransitionTo(orderModels.OrderStatusCancelled) {
		return fmt.Errorf("order cannot be cancelled from status %s", ord.Status)
	}

	// Update status to cancelled
	if err := s.orderRepo.UpdateOrderStatus(ctx, id, orderModels.OrderStatusCancelled, nil, "Order cancelled by user"); err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	// Release inventory for product items
	items := make([]orderRequests.CreateOrderItemRequest, len(ord.Items))
	for i, item := range ord.Items {
		items[i] = orderRequests.CreateOrderItemRequest{
			CatalogItemID:   item.CatalogItemID,
			CatalogItemType: item.CatalogItemType,
			Quantity:        item.Quantity,
			UnitPrice:       item.UnitPrice,
		}
	}

	if err := s.releaseInventoryForOrder(ctx, items, userID, ord.SellerOrganizationID); err != nil {
		// Log error but don't fail the cancellation
		fmt.Printf("Warning: failed to release inventory for cancelled order %s: %v\n", id, err)
	}

	return nil
}

// FulfillOrder marks an order as fulfilled
func (s *OrderService) FulfillOrder(ctx context.Context, id string, userID string, orgID string) error {
	// Validate input parameters
	if id == "" {
		return fmt.Errorf("order ID is required")
	}
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	if orgID == "" {
		return fmt.Errorf("organization ID is required")
	}

	// Validate permissions
	if err := s.ValidateOrderPermissions(ctx, id, userID, orgID, "fulfill"); err != nil {
		return fmt.Errorf("permission denied: %w", err)
	}

	// Get the order
	ord, err := s.orderRepo.GetOrderByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Check if order can be fulfilled
	if !ord.CanTransitionTo(orderModels.OrderStatusCompleted) {
		return fmt.Errorf("order cannot be fulfilled from status %s", ord.Status)
	}

	// Update status to fulfilled
	return s.orderRepo.UpdateOrderStatus(ctx, id, orderModels.OrderStatusCompleted, nil, "Order fulfilled")
}

// validateCreateOrderRequest validates the create order request
func (s *OrderService) validateCreateOrderRequest(req *orderRequests.CreateOrderRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.BuyerOrganizationID == "" {
		return fmt.Errorf("buyer organization ID is required")
	}

	if req.SellerOrganizationID == "" {
		return fmt.Errorf("seller organization ID is required")
	}

	// Prevent self-ordering (buyer and seller cannot be the same)
	if req.BuyerOrganizationID == req.SellerOrganizationID {
		return fmt.Errorf("buyer and seller organizations cannot be the same")
	}

	if len(req.Items) == 0 {
		return fmt.Errorf("order must contain at least one item")
	}

	// Validate individual items
	for i, item := range req.Items {
		if item.CatalogItemID == "" {
			return fmt.Errorf("catalog item ID is required for item %d", i+1)
		}

		if item.CatalogItemType == "" {
			return fmt.Errorf("catalog item type is required for item %d", i+1)
		}

		// Validate catalog item type
		validTypes := map[string]bool{
			"product": true,
			"service": true,
			"labour":  true,
		}
		if !validTypes[item.CatalogItemType] {
			return fmt.Errorf("invalid catalog item type '%s' for item %d", item.CatalogItemType, i+1)
		}

		if item.Quantity.LessThanOrEqual(decimal.Zero) {
			return fmt.Errorf("item quantity must be greater than 0 for item %d", i+1)
		}

		if item.UnitPrice.LessThan(decimal.Zero) {
			return fmt.Errorf("item unit price cannot be negative for item %d", i+1)
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

	// Validate shipping address if provided
	if req.ShippingAddress != nil {
		if err := s.validateAddress(req.ShippingAddress); err != nil {
			return fmt.Errorf("invalid shipping address: %w", err)
		}
	}

	// Validate notes length
	if len(req.Notes) > 1000 {
		return fmt.Errorf("notes cannot exceed 1000 characters")
	}

	return nil
}

// validateUpdateOrderRequest validates the update order request
func (s *OrderService) validateUpdateOrderRequest(req *orderRequests.UpdateOrderRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Validate items if provided
	if req.Items != nil && len(req.Items) > 0 {
		// Validate individual items
		for i, item := range req.Items {
			if item.CatalogItemID == "" {
				return fmt.Errorf("catalog item ID is required for item %d", i+1)
			}

			if item.CatalogItemType == "" {
				return fmt.Errorf("catalog item type is required for item %d", i+1)
			}

			// Validate catalog item type
			validTypes := map[string]bool{
				"product": true,
				"service": true,
				"labour":  true,
			}
			if !validTypes[item.CatalogItemType] {
				return fmt.Errorf("invalid catalog item type '%s' for item %d", item.CatalogItemType, i+1)
			}

			if item.Quantity.LessThanOrEqual(decimal.Zero) {
				return fmt.Errorf("item quantity must be greater than 0 for item %d", i+1)
			}

			if item.UnitPrice.LessThan(decimal.Zero) {
				return fmt.Errorf("item unit price cannot be negative for item %d", i+1)
			}

			// Validate reasonable limits
			maxQuantity := decimal.NewFromInt(1000000)
			if item.Quantity.GreaterThan(maxQuantity) {
				return fmt.Errorf("item quantity exceeds maximum allowed (1,000,000) for item %d", i+1)
			}

			maxPrice := decimal.NewFromInt(10000000)
			if item.UnitPrice.GreaterThan(maxPrice) {
				return fmt.Errorf("item unit price exceeds maximum allowed (10,000,000) for item %d", i+1)
			}
		}
	}

	// Validate shipping address if provided
	if req.ShippingAddress != nil {
		if err := s.validateAddress(req.ShippingAddress); err != nil {
			return fmt.Errorf("invalid shipping address: %w", err)
		}
	}

	// Validate notes length if provided
	if req.Notes != nil && len(*req.Notes) > 1000 {
		return fmt.Errorf("notes cannot exceed 1000 characters")
	}

	// Validate reason length if provided
	if req.Reason != nil && len(*req.Reason) > 500 {
		return fmt.Errorf("reason cannot exceed 500 characters")
	}

	return nil
}

// validatePaymentMethod validates payment method for auction orders
func (s *OrderService) validatePaymentMethod(paymentMethod string) error {
	validMethods := map[string]bool{
		"credit_card":      true,
		"debit_card":       true,
		"bank_transfer":    true,
		"upi":              true,
		"cash_on_delivery": true,
		"digital_wallet":   true,
		"net_banking":      true,
	}

	if paymentMethod == "" {
		return fmt.Errorf("payment method is required")
	}

	if !validMethods[paymentMethod] {
		return fmt.Errorf("invalid payment method: %s", paymentMethod)
	}

	return nil
}

// processPaymentForAuctionOrder processes payment for an auction-derived order
func (s *OrderService) processPaymentForAuctionOrder(ctx context.Context, order *orderModels.Order, paymentMethod string) (*PaymentResult, error) {
	// Validate payment method
	if err := s.validatePaymentMethod(paymentMethod); err != nil {
		return nil, fmt.Errorf("payment validation failed: %w", err)
	}

	// Create payment request
	paymentReq := &PaymentRequest{
		OrderID:       order.ID,
		OrderNumber:   order.OrderNumber,
		Amount:        order.TotalAmount,
		Currency:      "INR", // Default currency
		PaymentMethod: paymentMethod,
		BuyerID:       order.BuyerUserID,
		SellerID:      "", // Will be extracted from order metadata
		Description:   fmt.Sprintf("Payment for auction order %s", order.OrderNumber),
		Metadata: map[string]interface{}{
			"source":     "marketplace_auction",
			"order_type": "auction_order",
		},
	}

	// Extract seller information from order metadata
	if metadata, err := order.GetMetadata(); err == nil {
		if bidID, exists := metadata["bid_id"]; exists {
			paymentReq.Metadata["bid_id"] = bidID
		}
		if listingID, exists := metadata["listing_id"]; exists {
			paymentReq.Metadata["listing_id"] = listingID
		}
	}

	// Process payment based on method
	switch paymentMethod {
	case "cash_on_delivery":
		return s.processCashOnDeliveryPayment(ctx, paymentReq)
	case "upi", "digital_wallet":
		return s.processDigitalPayment(ctx, paymentReq)
	case "credit_card", "debit_card":
		return s.processCardPayment(ctx, paymentReq)
	case "bank_transfer", "net_banking":
		return s.processBankTransferPayment(ctx, paymentReq)
	default:
		return nil, fmt.Errorf("payment method %s not implemented", paymentMethod)
	}
}

// processCashOnDeliveryPayment handles cash on delivery payments
func (s *OrderService) processCashOnDeliveryPayment(ctx context.Context, req *PaymentRequest) (*PaymentResult, error) {
	// For COD, we just mark the payment as pending and update order status
	result := &PaymentResult{
		PaymentID:     generatePaymentID(),
		OrderID:       req.OrderID,
		Status:        PaymentStatusPending,
		Amount:        req.Amount,
		Currency:      req.Currency,
		PaymentMethod: req.PaymentMethod,
		ProcessedAt:   time.Now(),
		Message:       "Cash on delivery payment scheduled",
		Metadata: map[string]interface{}{
			"cod_instructions": "Payment will be collected upon delivery",
			"requires_cash":    true,
		},
	}

	return result, nil
}

// processDigitalPayment handles UPI and digital wallet payments
func (s *OrderService) processDigitalPayment(ctx context.Context, req *PaymentRequest) (*PaymentResult, error) {
	// In a real implementation, this would integrate with payment gateways like Razorpay, Paytm, etc.
	// For now, we'll simulate the payment process

	result := &PaymentResult{
		PaymentID:     generatePaymentID(),
		OrderID:       req.OrderID,
		Status:        PaymentStatusPending,
		Amount:        req.Amount,
		Currency:      req.Currency,
		PaymentMethod: req.PaymentMethod,
		ProcessedAt:   time.Now(),
		Message:       "Digital payment initiated",
		Metadata: map[string]interface{}{
			"payment_gateway":       "simulated",
			"requires_confirmation": true,
		},
	}

	// Simulate payment processing delay
	// In real implementation, this would be handled asynchronously via webhooks
	result.Status = PaymentStatusCompleted
	result.Message = "Digital payment completed successfully"
	result.Metadata["transaction_id"] = fmt.Sprintf("TXN_%d", time.Now().Unix())

	return result, nil
}

// processCardPayment handles credit/debit card payments
func (s *OrderService) processCardPayment(ctx context.Context, req *PaymentRequest) (*PaymentResult, error) {
	// In a real implementation, this would integrate with payment gateways
	result := &PaymentResult{
		PaymentID:     generatePaymentID(),
		OrderID:       req.OrderID,
		Status:        PaymentStatusPending,
		Amount:        req.Amount,
		Currency:      req.Currency,
		PaymentMethod: req.PaymentMethod,
		ProcessedAt:   time.Now(),
		Message:       "Card payment processing",
		Metadata: map[string]interface{}{
			"payment_gateway": "simulated",
			"card_type":       "unknown",
		},
	}

	// Simulate payment processing
	result.Status = PaymentStatusCompleted
	result.Message = "Card payment completed successfully"
	result.Metadata["authorization_code"] = fmt.Sprintf("AUTH_%d", time.Now().Unix())

	return result, nil
}

// processBankTransferPayment handles bank transfer and net banking payments
func (s *OrderService) processBankTransferPayment(ctx context.Context, req *PaymentRequest) (*PaymentResult, error) {
	result := &PaymentResult{
		PaymentID:     generatePaymentID(),
		OrderID:       req.OrderID,
		Status:        PaymentStatusPending,
		Amount:        req.Amount,
		Currency:      req.Currency,
		PaymentMethod: req.PaymentMethod,
		ProcessedAt:   time.Now(),
		Message:       "Bank transfer initiated",
		Metadata: map[string]interface{}{
			"requires_manual_verification": true,
			"settlement_time":              "1-3 business days",
		},
	}

	return result, nil
}

// generatePaymentID generates a unique payment ID
func generatePaymentID() string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("PAY_%d", timestamp)
}

// validateAddress validates an address structure
func (s *OrderService) validateAddress(addr *orderRequests.Address) error {
	if addr == nil {
		return fmt.Errorf("address cannot be nil")
	}

	if addr.Street == "" {
		return fmt.Errorf("street is required")
	}

	if addr.City == "" {
		return fmt.Errorf("city is required")
	}

	if addr.State == "" {
		return fmt.Errorf("state is required")
	}

	if addr.PostalCode == "" {
		return fmt.Errorf("postal code is required")
	}

	if addr.Country == "" {
		return fmt.Errorf("country is required")
	}

	// Validate field lengths
	if len(addr.Street) > 255 {
		return fmt.Errorf("street cannot exceed 255 characters")
	}

	if len(addr.City) > 100 {
		return fmt.Errorf("city cannot exceed 100 characters")
	}

	if len(addr.State) > 100 {
		return fmt.Errorf("state cannot exceed 100 characters")
	}

	if len(addr.PostalCode) > 20 {
		return fmt.Errorf("postal code cannot exceed 20 characters")
	}

	if len(addr.Country) > 100 {
		return fmt.Errorf("country cannot exceed 100 characters")
	}

	return nil
}

// ValidateOrderPermissions validates that a user has permission to perform an action on an order
func (s *OrderService) ValidateOrderPermissions(ctx context.Context, orderID string, userID string, orgID string, action string) error {
	// Validate input parameters
	if orderID == "" {
		return fmt.Errorf("order ID is required")
	}
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	if orgID == "" {
		return fmt.Errorf("organization ID is required")
	}
	if action == "" {
		return fmt.Errorf("action is required")
	}

	// Get the order to check ownership
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order for permission check: %w", err)
	}

	// Check if the user's organization is either the buyer or seller
	isAuthorized := false
	switch action {
	case "read", "view":
		// Users can read orders where their org is buyer or seller
		isAuthorized = (order.BuyerOrganizationID == orgID) || (order.SellerOrganizationID == orgID)
	case "update", "cancel":
		// Users can update/cancel orders where their org is the buyer
		isAuthorized = (order.BuyerOrganizationID == orgID)
	case "fulfill", "ship":
		// Users can fulfill/ship orders where their org is the seller
		isAuthorized = (order.SellerOrganizationID == orgID)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	if !isAuthorized {
		return fmt.Errorf("user organization %s does not have permission to %s order %s", orgID, action, orderID)
	}

	return nil
}

// GetOrderAnalytics retrieves order analytics for an organization
func (s *OrderService) GetOrderAnalytics(ctx context.Context, orgID string, startDate, endDate time.Time) (*OrderAnalytics, error) {
	// TODO: Implement order analytics
	return nil, fmt.Errorf("not implemented")
}

// ProcessPaymentForOrder processes payment for an existing order
func (s *OrderService) ProcessPaymentForOrder(ctx context.Context, orderID string, paymentMethod string, userID string, orgID string) (*PaymentResult, error) {
	// Validate input parameters
	if orderID == "" {
		return nil, fmt.Errorf("order ID is required")
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

	// Validate permissions
	if err := s.ValidateOrderPermissions(ctx, orderID, userID, orgID, "update"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Get the order
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Validate order status for payment processing
	if order.Status != orderModels.OrderStatusPending && order.Status != orderModels.OrderStatusConfirmed {
		return nil, fmt.Errorf("order status %s does not allow payment processing", order.Status)
	}

	// Process payment
	paymentResult, err := s.processPaymentForAuctionOrder(ctx, order, paymentMethod)
	if err != nil {
		return nil, fmt.Errorf("payment processing failed: %w", err)
	}

	// Update order metadata with payment information
	metadata, _ := order.GetMetadata()
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	metadata["payment_id"] = paymentResult.PaymentID
	metadata["payment_status"] = string(paymentResult.Status)
	metadata["payment_method"] = paymentResult.PaymentMethod
	metadata["payment_processed_at"] = paymentResult.ProcessedAt.Format(time.RFC3339)

	if err := order.SetMetadata(metadata); err != nil {
		fmt.Printf("Warning: failed to update order metadata for payment: %v\n", err)
	}

	// Update order status based on payment result
	if paymentResult.Status == PaymentStatusCompleted {
		if err := s.UpdateOrderStatus(ctx, orderID, &orderRequests.UpdateOrderStatusRequest{
			Status: orderModels.OrderStatusPaid,
			Reason: "Payment completed successfully",
		}, userID, orgID); err != nil {
			fmt.Printf("Warning: failed to update order status to paid: %v\n", err)
		}
	} else if paymentResult.Status == PaymentStatusFailed {
		if err := s.UpdateOrderStatus(ctx, orderID, &orderRequests.UpdateOrderStatusRequest{
			Status: orderModels.OrderStatusPending,
			Reason: "Payment failed, order reverted to pending",
		}, userID, orgID); err != nil {
			fmt.Printf("Warning: failed to update order status after payment failure: %v\n", err)
		}
	}

	// Save updated order
	if err := s.orderRepo.Update(ctx, order); err != nil {
		fmt.Printf("Warning: failed to save order after payment processing: %v\n", err)
	}

	return paymentResult, nil
}

// GetPaymentStatus retrieves the payment status for an order
func (s *OrderService) GetPaymentStatus(ctx context.Context, orderID string, userID string, orgID string) (*PaymentResult, error) {
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

	// Validate permissions
	if err := s.ValidateOrderPermissions(ctx, orderID, userID, orgID, "read"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Get the order
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Extract payment information from metadata
	metadata, err := order.GetMetadata()
	if err != nil || metadata == nil {
		return nil, fmt.Errorf("no payment information found for order")
	}

	paymentID, _ := metadata["payment_id"].(string)
	paymentStatusStr, _ := metadata["payment_status"].(string)
	paymentMethod, _ := metadata["payment_method"].(string)
	paymentProcessedAtStr, _ := metadata["payment_processed_at"].(string)

	if paymentID == "" {
		return nil, fmt.Errorf("no payment found for order")
	}

	// Parse processed time
	var processedAt time.Time
	if paymentProcessedAtStr != "" {
		if parsedTime, err := time.Parse(time.RFC3339, paymentProcessedAtStr); err == nil {
			processedAt = parsedTime
		}
	}

	// Build payment result
	paymentResult := &PaymentResult{
		PaymentID:     paymentID,
		OrderID:       orderID,
		Status:        PaymentStatus(paymentStatusStr),
		Amount:        order.TotalAmount,
		Currency:      "INR", // Default currency
		PaymentMethod: paymentMethod,
		ProcessedAt:   processedAt,
		Message:       fmt.Sprintf("Payment status: %s", paymentStatusStr),
		Metadata:      metadata,
	}

	return paymentResult, nil
}

// HandlePaymentFailure handles payment failure scenarios
func (s *OrderService) HandlePaymentFailure(ctx context.Context, orderID string, reason string, userID string, orgID string) error {
	// Validate input parameters
	if orderID == "" {
		return fmt.Errorf("order ID is required")
	}
	if reason == "" {
		return fmt.Errorf("failure reason is required")
	}
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	if orgID == "" {
		return fmt.Errorf("organization ID is required")
	}

	// Validate permissions
	if err := s.ValidateOrderPermissions(ctx, orderID, userID, orgID, "update"); err != nil {
		return fmt.Errorf("permission denied: %w", err)
	}

	// Get the order
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Update order metadata with failure information
	metadata, _ := order.GetMetadata()
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	metadata["payment_status"] = string(PaymentStatusFailed)
	metadata["payment_failure_reason"] = reason
	metadata["payment_failed_at"] = time.Now().Format(time.RFC3339)

	if err := order.SetMetadata(metadata); err != nil {
		return fmt.Errorf("failed to update order metadata: %w", err)
	}

	// Update order status to reflect payment failure
	if err := s.UpdateOrderStatus(ctx, orderID, &orderRequests.UpdateOrderStatusRequest{
		Status: orderModels.OrderStatusPending,
		Reason: fmt.Sprintf("Payment failed: %s", reason),
	}, userID, orgID); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	// Save updated order
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	return nil
}

// OrderAnalytics represents order analytics data
type OrderAnalytics struct {
	TotalOrders       int     `json:"total_orders"`
	TotalRevenue      float64 `json:"total_revenue"`
	AverageOrderValue float64 `json:"average_order_value"`
	Currency          string  `json:"currency"`
	Period            string  `json:"period"`
}

// BidOrderValidation represents the validation result for creating an order from a bid
type BidOrderValidation struct {
	Valid           bool            `json:"valid"`
	BidID           string          `json:"bid_id"`
	ListingID       string          `json:"listing_id"`
	ProductID       string          `json:"product_id"`
	WinningAmount   decimal.Decimal `json:"winning_amount"`
	Quantity        decimal.Decimal `json:"quantity"`
	Currency        string          `json:"currency"`
	SellerID        string          `json:"seller_id"`
	BuyerID         string          `json:"buyer_id"`
	SellerOrgID     string          `json:"seller_org_id"`
	BuyerOrgID      string          `json:"buyer_org_id"`
	ProductName     string          `json:"product_name"`
	ProductSKU      string          `json:"product_sku"`
	ExpiresAt       time.Time       `json:"expires_at"`
	CanCreateOrder  bool            `json:"can_create_order"`
	ValidationError string          `json:"validation_error,omitempty"`
}

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusCompleted PaymentStatus = "COMPLETED"
	PaymentStatusFailed    PaymentStatus = "FAILED"
	PaymentStatusCancelled PaymentStatus = "CANCELLED"
	PaymentStatusRefunded  PaymentStatus = "REFUNDED"
)

// PaymentRequest represents a payment processing request
type PaymentRequest struct {
	OrderID       string                 `json:"order_id"`
	OrderNumber   string                 `json:"order_number"`
	Amount        decimal.Decimal        `json:"amount"`
	Currency      string                 `json:"currency"`
	PaymentMethod string                 `json:"payment_method"`
	BuyerID       string                 `json:"buyer_id"`
	SellerID      string                 `json:"seller_id"`
	Description   string                 `json:"description"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// PaymentResult represents the result of a payment processing attempt
type PaymentResult struct {
	PaymentID     string                 `json:"payment_id"`
	OrderID       string                 `json:"order_id"`
	Status        PaymentStatus          `json:"status"`
	Amount        decimal.Decimal        `json:"amount"`
	Currency      string                 `json:"currency"`
	PaymentMethod string                 `json:"payment_method"`
	ProcessedAt   time.Time              `json:"processed_at"`
	Message       string                 `json:"message"`
	ErrorCode     string                 `json:"error_code,omitempty"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// validateOrderItems validates order items and checks inventory availability
func (s *OrderService) validateOrderItems(ctx context.Context, items []orderRequests.CreateOrderItemRequest, sellerOrgID, userID string) error {
	for i, item := range items {
		// Validate catalog item exists
		catalogItem, err := s.catalogSvc.GetCatalogItemByID(ctx, item.CatalogItemID)
		if err != nil {
			return fmt.Errorf("item %d: catalog item '%s' not found: %w", i+1, item.CatalogItemID, err)
		}

		// Validate catalog item belongs to seller organization
		if catalogItem.OrganizationID != sellerOrgID {
			return fmt.Errorf("item %d: catalog item '%s' does not belong to seller organization", i+1, item.CatalogItemID)
		}

		// Validate catalog item is active
		if !catalogItem.IsActive {
			return fmt.Errorf("item %d: catalog item '%s' is not active", i+1, item.CatalogItemID)
		}

		// For products, check inventory availability
		if item.CatalogItemType == "product" && s.inventorySvc != nil {
			available, err := s.inventorySvc.GetAvailableQuantity(ctx, item.CatalogItemID, userID, sellerOrgID)
			if err != nil {
				return fmt.Errorf("item %d: failed to check inventory for product '%s': %w", i+1, item.CatalogItemID, err)
			}

			if available.LessThan(item.Quantity) {
				return fmt.Errorf("item %d: insufficient inventory for product '%s' (available: %s, requested: %s)",
					i+1, item.CatalogItemID, available.String(), item.Quantity.String())
			}
		}

		// Validate quantity is positive
		if item.Quantity.LessThanOrEqual(decimal.Zero) {
			return fmt.Errorf("item %d: quantity must be greater than zero", i+1)
		}

		// Validate unit price is non-negative
		if item.UnitPrice.LessThan(decimal.Zero) {
			return fmt.Errorf("item %d: unit price cannot be negative", i+1)
		}
	}

	return nil
}

// reserveInventoryForOrder reserves inventory for order items
func (s *OrderService) reserveInventoryForOrder(ctx context.Context, items []orderRequests.CreateOrderItemRequest, userID, orgID string) error {
	if s.inventorySvc == nil {
		return fmt.Errorf("inventory service not configured")
	}

	var reservedItems []struct {
		catalogItemID string
		quantity      decimal.Decimal
	}

	for i, item := range items {
		// Only reserve inventory for products
		if item.CatalogItemType == "product" {
			// Reserve the inventory
			_, err := s.inventorySvc.ReserveInventory(ctx, item.CatalogItemID, item.Quantity, userID, orgID)
			if err != nil {
				// Rollback previously reserved items
				s.rollbackInventoryReservations(ctx, reservedItems, userID, orgID)
				return fmt.Errorf("item %d: failed to reserve inventory for product '%s': %w", i+1, item.CatalogItemID, err)
			}

			reservedItems = append(reservedItems, struct {
				catalogItemID string
				quantity      decimal.Decimal
			}{
				catalogItemID: item.CatalogItemID,
				quantity:      item.Quantity,
			})
		}
	}

	return nil
}

// rollbackInventoryReservations releases previously reserved inventory
func (s *OrderService) rollbackInventoryReservations(ctx context.Context, reservedItems []struct {
	catalogItemID string
	quantity      decimal.Decimal
}, userID, orgID string) {
	for _, reserved := range reservedItems {
		if err := s.inventorySvc.ReleaseInventory(ctx, reserved.catalogItemID, reserved.quantity, userID, orgID); err != nil {
			// Log error but don't fail the rollback
			fmt.Printf("Warning: failed to rollback inventory reservation for item %s: %v\n", reserved.catalogItemID, err)
		}
	}
}

// releaseInventoryForOrder releases reserved inventory (e.g., when order is cancelled)
func (s *OrderService) releaseInventoryForOrder(ctx context.Context, items []orderRequests.CreateOrderItemRequest, userID, orgID string) error {
	if s.inventorySvc == nil {
		return fmt.Errorf("inventory service not configured")
	}

	var errors []error

	for i, item := range items {
		// Only release inventory for products
		if item.CatalogItemType == "product" {
			if err := s.inventorySvc.ReleaseInventory(ctx, item.CatalogItemID, item.Quantity, userID, orgID); err != nil {
				errors = append(errors, fmt.Errorf("item %d: failed to release inventory for product '%s': %w", i+1, item.CatalogItemID, err))
			}
		}
	}

	if len(errors) > 0 {
		// Return the first error, but log all errors
		for _, err := range errors[1:] {
			fmt.Printf("Additional inventory release error: %v\n", err)
		}
		return errors[0]
	}

	return nil
}

// CreateOrderFromBid creates an order from a winning bid
func (s *OrderService) CreateOrderFromBid(ctx context.Context, req *orderRequests.CreateOrderFromBidRequest, userID string, orgID string) (*orderModels.Order, error) {
	// Validate input parameters
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if req.BidID == "" {
		return nil, fmt.Errorf("bid ID is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	// Validate the bid for order creation
	validation, err := s.ValidateBidForOrder(ctx, req.BidID, userID, orgID)
	if err != nil {
		return nil, fmt.Errorf("bid validation failed: %w", err)
	}

	if !validation.Valid || !validation.CanCreateOrder {
		return nil, fmt.Errorf("bid is not valid for order creation: %s", validation.ValidationError)
	}

	// Create order request from bid information
	createOrderReq := &orderRequests.CreateOrderRequest{
		BuyerOrganizationID:  validation.BuyerOrgID,
		SellerOrganizationID: validation.SellerOrgID,
		Items: []orderRequests.CreateOrderItemRequest{
			{
				CatalogItemID:   validation.ProductID,
				CatalogItemType: "product", // Assuming marketplace items are products
				Quantity:        validation.Quantity,
				UnitPrice:       validation.WinningAmount,
			},
		},
		ShippingAddress: req.ShippingAddress,
		Notes:           req.Notes,
	}

	// Create the order using existing order creation logic
	order, err := s.CreateOrder(ctx, createOrderReq, userID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to create order from bid: %w", err)
	}

	// Add bid-specific metadata to the order
	bidMetadata := map[string]interface{}{
		"source":            "marketplace_bid",
		"bid_id":            req.BidID,
		"listing_id":        validation.ListingID,
		"winning_amount":    validation.WinningAmount.String(),
		"payment_method":    req.PaymentMethod,
		"auction_closed_at": validation.ExpiresAt.Format(time.RFC3339),
	}

	// Process payment for the auction order
	paymentResult, err := s.processPaymentForAuctionOrder(ctx, order, req.PaymentMethod)
	if err != nil {
		// Log payment error but don't fail order creation
		fmt.Printf("Warning: payment processing failed for order %s: %v\n", order.ID, err)

		// Add payment failure information to metadata
		bidMetadata["payment_status"] = "FAILED"
		bidMetadata["payment_error"] = err.Error()
	} else {
		// Add payment success information to metadata
		bidMetadata["payment_id"] = paymentResult.PaymentID
		bidMetadata["payment_status"] = string(paymentResult.Status)
		bidMetadata["payment_processed_at"] = paymentResult.ProcessedAt.Format(time.RFC3339)

		// Update order status based on payment result
		if paymentResult.Status == PaymentStatusCompleted {
			// Update order status to paid
			if err := s.UpdateOrderStatus(ctx, order.ID, &orderRequests.UpdateOrderStatusRequest{
				Status: orderModels.OrderStatusPaid,
				Reason: "Payment completed successfully",
			}, userID, orgID); err != nil {
				fmt.Printf("Warning: failed to update order status to paid for order %s: %v\n", order.ID, err)
			}
		} else if paymentResult.Status == PaymentStatusPending {
			// Update order status to confirmed (awaiting payment)
			if err := s.UpdateOrderStatus(ctx, order.ID, &orderRequests.UpdateOrderStatusRequest{
				Status: orderModels.OrderStatusConfirmed,
				Reason: "Order confirmed, payment pending",
			}, userID, orgID); err != nil {
				fmt.Printf("Warning: failed to update order status to confirmed for order %s: %v\n", order.ID, err)
			}
		}
	}

	// Update metadata with payment information
	if err := order.SetMetadata(bidMetadata); err != nil {
		// Log error but don't fail order creation
		fmt.Printf("Warning: failed to set updated metadata on order %s: %v\n", order.ID, err)
	}

	// Update the order in the repository to save metadata
	if err := s.orderRepo.Update(ctx, order); err != nil {
		// Log error but don't fail order creation
		fmt.Printf("Warning: failed to update order metadata for order %s: %v\n", order.ID, err)
	}

	return order, nil
}

// ValidateBidForOrder validates that a bid can be used to create an order
func (s *OrderService) ValidateBidForOrder(ctx context.Context, bidID string, userID string, orgID string) (*BidOrderValidation, error) {
	// Initialize validation result
	validation := &BidOrderValidation{
		Valid:          false,
		BidID:          bidID,
		CanCreateOrder: false,
	}

	// Check if marketplace service is available
	if s.marketplaceSvc == nil {
		validation.ValidationError = "marketplace service not available"
		return validation, fmt.Errorf("marketplace service not configured")
	}

	// Get bid information from marketplace service
	bidInterface, err := s.marketplaceSvc.GetBid(ctx, bidID, userID)
	if err != nil {
		validation.ValidationError = fmt.Sprintf("failed to get bid: %v", err)
		return validation, fmt.Errorf("failed to get bid from marketplace: %w", err)
	}

	// Type assert to marketplace bid (we'll need to handle this more gracefully in production)
	bid, ok := bidInterface.(map[string]interface{})
	if !ok {
		validation.ValidationError = "invalid bid data format"
		return validation, fmt.Errorf("invalid bid data format")
	}

	// Extract bid information
	bidderID, _ := bid["bidder_id"].(string)
	listingID, _ := bid["listing_id"].(string)
	bidAmountStr, _ := bid["bid_amount"].(string)
	quantityStr, _ := bid["quantity"].(string)
	currency, _ := bid["currency"].(string)
	status, _ := bid["status"].(string)
	isWinning, _ := bid["is_winning"].(bool)

	// Validate that the user is the bidder
	if bidderID != userID {
		validation.ValidationError = "user is not the bidder"
		return validation, fmt.Errorf("user %s is not the bidder %s", userID, bidderID)
	}

	// Validate that the bid is in winning status
	if status != "WINNING" && !isWinning {
		validation.ValidationError = "bid is not in winning status"
		return validation, fmt.Errorf("bid status is %s, not winning", status)
	}

	// Parse bid amount and quantity
	bidAmount, err := decimal.NewFromString(bidAmountStr)
	if err != nil {
		validation.ValidationError = "invalid bid amount format"
		return validation, fmt.Errorf("invalid bid amount: %w", err)
	}

	quantity, err := decimal.NewFromString(quantityStr)
	if err != nil {
		validation.ValidationError = "invalid quantity format"
		return validation, fmt.Errorf("invalid quantity: %w", err)
	}

	// Get listing information
	listingInterface, err := s.marketplaceSvc.GetListing(ctx, listingID, userID, orgID)
	if err != nil {
		validation.ValidationError = fmt.Sprintf("failed to get listing: %v", err)
		return validation, fmt.Errorf("failed to get listing from marketplace: %w", err)
	}

	listing, ok := listingInterface.(map[string]interface{})
	if !ok {
		validation.ValidationError = "invalid listing data format"
		return validation, fmt.Errorf("invalid listing data format")
	}

	// Extract listing information
	productID, _ := listing["product_id"].(string)
	sellerID, _ := listing["seller_id"].(string)
	sellerOrgID, _ := listing["organization_id"].(string)
	listingStatus, _ := listing["status"].(string)
	expiresAtStr, _ := listing["expires_at"].(string)

	// Validate that the listing is closed
	if listingStatus != "CLOSED" {
		validation.ValidationError = "listing is not closed"
		return validation, fmt.Errorf("listing status is %s, not closed", listingStatus)
	}

	// Parse expires_at time
	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		validation.ValidationError = "invalid expires_at format"
		return validation, fmt.Errorf("invalid expires_at: %w", err)
	}

	// Get product information from catalog service
	catalogItem, err := s.catalogSvc.GetCatalogItemByID(ctx, productID)
	if err != nil {
		validation.ValidationError = fmt.Sprintf("failed to get product: %v", err)
		return validation, fmt.Errorf("failed to get product from catalog: %w", err)
	}

	// Validate that the product belongs to the seller organization
	if catalogItem.OrganizationID != sellerOrgID {
		validation.ValidationError = "product does not belong to seller organization"
		return validation, fmt.Errorf("product organization mismatch")
	}

	// Check if an order already exists for this bid
	existingOrders, _, err := s.orderRepo.ListOrders(ctx, &orderRequests.ListOrdersRequest{
		BuyerOrganizationID: &orgID,
		PageSize:            100,
	}, 0, 100)
	if err == nil {
		for _, existingOrder := range existingOrders {
			if metadata, err := existingOrder.GetMetadata(); err == nil {
				if existingBidID, exists := metadata["bid_id"]; exists && existingBidID == bidID {
					validation.ValidationError = "order already exists for this bid"
					return validation, fmt.Errorf("order already exists for bid %s", bidID)
				}
			}
		}
	}

	// All validations passed
	validation.Valid = true
	validation.CanCreateOrder = true
	validation.ListingID = listingID
	validation.ProductID = productID
	validation.WinningAmount = bidAmount
	validation.Quantity = quantity
	validation.Currency = currency
	validation.SellerID = sellerID
	validation.BuyerID = bidderID
	validation.SellerOrgID = sellerOrgID
	validation.BuyerOrgID = orgID
	validation.ProductName = catalogItem.Name
	validation.ProductSKU = catalogItem.SKU
	validation.ExpiresAt = expiresAt

	return validation, nil
}

// sellInventoryForOrder converts reserved inventory to sold (e.g., when order is completed)
func (s *OrderService) sellInventoryForOrder(ctx context.Context, items []orderRequests.CreateOrderItemRequest, userID, orgID string) error {
	if s.inventorySvc == nil {
		return fmt.Errorf("inventory service not configured")
	}

	var errors []error

	for i, item := range items {
		// Only sell inventory for products
		if item.CatalogItemType == "product" {
			if err := s.inventorySvc.SellInventory(ctx, item.CatalogItemID, item.Quantity, userID, orgID); err != nil {
				errors = append(errors, fmt.Errorf("item %d: failed to sell inventory for product '%s': %w", i+1, item.CatalogItemID, err))
			}
		}
	}

	if len(errors) > 0 {
		// Return the first error, but log all errors
		for _, err := range errors[1:] {
			fmt.Printf("Additional inventory sell error: %v\n", err)
		}
		return errors[0]
	}

	return nil
}
