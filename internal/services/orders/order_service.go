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
}

// OrderService provides business logic for order operations
type OrderService struct {
    orderRepo    *orders.OrderRepository
    catalogSvc   catalog.CatalogServiceInterface
    inventorySvc inventory.InventoryService
}

// NewOrderService creates a new order service
func NewOrderService(orderRepo *orders.OrderRepository, catalogSvc catalog.CatalogServiceInterface, inventorySvc inventory.InventoryService) *OrderService {
    return &OrderService{
        orderRepo:    orderRepo,
        catalogSvc:   catalogSvc,
        inventorySvc: inventorySvc,
    }
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

        // Create order item with proper calculations
        quantity := itemReq.Quantity
        orderItem := orderModels.NewOrderItem(
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

// OrderAnalytics represents order analytics data
type OrderAnalytics struct {
    TotalOrders       int     `json:"total_orders"`
    TotalRevenue      float64 `json:"total_revenue"`
    AverageOrderValue float64 `json:"average_order_value"`
    Currency          string  `json:"currency"`
    Period            string  `json:"period"`
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
