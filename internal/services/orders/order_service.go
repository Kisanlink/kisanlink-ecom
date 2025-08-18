package orders

import (
	"context"
	"fmt"
	"time"

	"kisanlink-ecom/internal/models/order"
	"kisanlink-ecom/internal/repositories/orders"
	"kisanlink-ecom/internal/services/catalog"
)

// OrderServiceInterface defines the interface for order operations
type OrderServiceInterface interface {
	CreateOrder(ctx context.Context, req *order.CreateOrderRequest, userID string) (*order.Order, error)
	GetOrderByID(ctx context.Context, id string) (*order.Order, error)
	GetOrderByNumber(ctx context.Context, orderNumber string) (*order.Order, error)
	UpdateOrderStatus(ctx context.Context, id string, req *order.UpdateOrderStatusRequest, userID string) error
	ListOrders(ctx context.Context, filter *order.OrderFilter, offset, limit int) ([]*order.Order, int, error)
	GetOrdersByBuyer(ctx context.Context, buyerID string, limit, offset int) ([]*order.Order, error)
	GetOrdersBySeller(ctx context.Context, sellerID string, limit, offset int) ([]*order.Order, error)
	GetOrderSummary(ctx context.Context, id string) (*order.OrderSummary, error)
	CancelOrder(ctx context.Context, id string, userID string) error
	FulfillOrder(ctx context.Context, id string, userID string) error
	GetOrderAnalytics(ctx context.Context, orgID string, startDate, endDate time.Time) (*OrderAnalytics, error)
}

// OrderService provides business logic for order operations
type OrderService struct {
	orderRepo  *orders.OrderRepository
	catalogSvc catalog.CatalogService
}

// NewOrderService creates a new order service
func NewOrderService(orderRepo *orders.OrderRepository, catalogSvc catalog.CatalogService) *OrderService {
	return &OrderService{
		orderRepo:  orderRepo,
		catalogSvc: catalogSvc,
	}
}

// CreateOrder creates a new order with business logic validation
func (s *OrderService) CreateOrder(ctx context.Context, req *order.CreateOrderRequest, userID string) (*order.Order, error) {
	// Validate request
	if err := s.validateCreateOrderRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Validate order items
	if err := s.orderRepo.ValidateOrderItems(ctx, req.Items, req.SellerID); err != nil {
		return nil, fmt.Errorf("order item validation failed: %w", err)
	}

	// Create order instance
	ord := order.NewOrder(req.BuyerID, req.BuyerType, req.SellerID, req.SellerType, req.Currency)
	ord.Notes = req.Notes
	ord.CreatedBy = userID
	ord.UpdatedBy = userID

	// Add items to order
	for _, itemReq := range req.Items {
		// Get catalog item details
		catalogItem, err := s.catalogSvc.GetCatalogItemByID(ctx, itemReq.ItemID)
		if err != nil {
			return nil, fmt.Errorf("failed to get catalog item %s: %w", itemReq.ItemID, err)
		}

		// Create order item
		orderItem := order.NewOrderItem(
			ord.ID,
			itemReq.ItemID,
			itemReq.Type,
			catalogItem.SKU,
			catalogItem.Name,
			catalogItem.Description,
			itemReq.Quantity,
			itemReq.UnitPrice,
			req.Currency,
			catalogItem.UOM,
		)
		orderItem.Notes = itemReq.Notes
		orderItem.CreatedBy = userID
		orderItem.UpdatedBy = userID

		ord.AddItem(orderItem)
	}

	// Reserve inventory for products
	if err := s.orderRepo.ReserveInventory(ctx, req.Items); err != nil {
		return nil, fmt.Errorf("failed to reserve inventory: %w", err)
	}

	// Create the order
	if err := s.orderRepo.CreateOrder(ctx, ord); err != nil {
		// Release inventory if order creation fails
		_ = s.orderRepo.ReleaseInventory(ctx, req.Items)
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return ord, nil
}

// GetOrderByID retrieves an order by ID
func (s *OrderService) GetOrderByID(ctx context.Context, id string) (*order.Order, error) {
	return s.orderRepo.GetOrderByID(ctx, id)
}

// GetOrderByNumber retrieves an order by order number
func (s *OrderService) GetOrderByNumber(ctx context.Context, orderNumber string) (*order.Order, error) {
	return s.orderRepo.GetOrderByNumber(ctx, orderNumber)
}

// UpdateOrderStatus updates the status of an order with business logic validation
func (s *OrderService) UpdateOrderStatus(ctx context.Context, id string, req *order.UpdateOrderStatusRequest, userID string) error {
	// Validate status transition
	if req.Status == "" {
		return fmt.Errorf("status is required")
	}

	// Convert time.Time to string for the repository method
	var expectedDeliveryStr *string
	if req.ExpectedDelivery != nil {
		deliveryStr := req.ExpectedDelivery.Format(time.RFC3339)
		expectedDeliveryStr = &deliveryStr
	}

	// Update the order status
	return s.orderRepo.UpdateOrderStatus(ctx, id, req.Status, expectedDeliveryStr, req.Notes)
}

// ListOrders retrieves orders with filtering and pagination
func (s *OrderService) ListOrders(ctx context.Context, filter *order.OrderFilter, offset, limit int) ([]*order.Order, int, error) {
	return s.orderRepo.ListOrders(ctx, filter, offset, limit)
}

// GetOrdersByBuyer retrieves orders for a specific buyer
func (s *OrderService) GetOrdersByBuyer(ctx context.Context, buyerID string, limit, offset int) ([]*order.Order, error) {
	return s.orderRepo.GetOrdersByBuyer(ctx, buyerID, limit, offset)
}

// GetOrdersBySeller retrieves orders for a specific seller
func (s *OrderService) GetOrdersBySeller(ctx context.Context, sellerID string, limit, offset int) ([]*order.Order, error) {
	return s.orderRepo.GetOrdersBySeller(ctx, sellerID, limit, offset)
}

// GetOrderSummary retrieves a summary of an order
func (s *OrderService) GetOrderSummary(ctx context.Context, id string) (*order.OrderSummary, error) {
	return s.orderRepo.GetOrderSummary(ctx, id)
}

// CancelOrder cancels an order and releases inventory
func (s *OrderService) CancelOrder(ctx context.Context, id string, userID string) error {
	// Get the order
	ord, err := s.orderRepo.GetOrderByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Check if order can be cancelled
	if !ord.CanTransitionTo(order.OrderStatusCancelled) {
		return fmt.Errorf("order cannot be cancelled from status %s", ord.Status)
	}

	// Update status to cancelled
	if err := s.orderRepo.UpdateOrderStatus(ctx, id, order.OrderStatusCancelled, nil, "Order cancelled by user"); err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	// Release inventory
	items := make([]order.CreateOrderItemRequest, len(ord.Items))
	for i, item := range ord.Items {
		items[i] = order.CreateOrderItemRequest{
			ItemID:    item.ItemID,
			Type:      item.Type,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		}
	}

	if err := s.orderRepo.ReleaseInventory(ctx, items); err != nil {
		// Log error but don't fail the cancellation
		// TODO: Add proper logging
		fmt.Printf("Warning: failed to release inventory for cancelled order %s: %v\n", id, err)
	}

	return nil
}

// FulfillOrder marks an order as fulfilled
func (s *OrderService) FulfillOrder(ctx context.Context, id string, userID string) error {
	// Get the order
	ord, err := s.orderRepo.GetOrderByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Check if order can be fulfilled
	if !ord.CanTransitionTo(order.OrderStatusFulfilled) {
		return fmt.Errorf("order cannot be fulfilled from status %s", ord.Status)
	}

	// Update status to fulfilled
	return s.orderRepo.UpdateOrderStatus(ctx, id, order.OrderStatusFulfilled, nil, "Order fulfilled")
}

// validateCreateOrderRequest validates the create order request
func (s *OrderService) validateCreateOrderRequest(req *order.CreateOrderRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.BuyerID == "" {
		return fmt.Errorf("buyer ID is required")
	}

	if req.BuyerType == "" {
		return fmt.Errorf("buyer type is required")
	}

	if req.SellerID == "" {
		return fmt.Errorf("seller ID is required")
	}

	if req.SellerType == "" {
		return fmt.Errorf("seller type is required")
	}

	if len(req.Items) == 0 {
		return fmt.Errorf("order must contain at least one item")
	}

	if req.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	// Validate individual items
	for i, item := range req.Items {
		if item.ItemID == "" {
			return fmt.Errorf("item ID is required for item %d", i+1)
		}

		if item.Type == "" {
			return fmt.Errorf("item type is required for item %d", i+1)
		}

		if item.Quantity <= 0 {
			return fmt.Errorf("item quantity must be greater than 0 for item %d", i+1)
		}

		if item.UnitPrice < 0 {
			return fmt.Errorf("item unit price cannot be negative for item %d", i+1)
		}
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
