package services

import (
	"context"
	"fmt"

	"kisanlink-ecom/internal/models/order"
	"kisanlink-ecom/internal/repositories"
)

// OrderService handles order-related business logic
type OrderService struct {
	orderRepo   *repositories.OrderRepository
	productRepo *repositories.ProductRepository
}

// NewOrderService creates a new order service instance
func NewOrderService(orderRepo *repositories.OrderRepository, productRepo *repositories.ProductRepository) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

// CreateOrder creates a new order
func (s *OrderService) CreateOrder(ctx context.Context, req order.CreateOrderRequest) (*order.Order, error) {
	// Validate items and calculate total
	var totalPrice float64
	var orderItems []order.OrderItem

	for _, item := range req.Items {
		// Get product to validate and get current price
		product, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("product %s not found: %v", item.ProductID, err)
		}

		// Check stock availability
		if product.Stock < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for product %s", item.ProductID)
		}

		// Create order item with current product price
		orderItem := order.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price,
			Product:   product,
		}
		orderItems = append(orderItems, orderItem)
		totalPrice += product.Price * float64(item.Quantity)
	}

	// Create order
	newOrder := order.NewOrder(req.UserID, orderItems, totalPrice, "USD")

	if err := newOrder.BeforeCreate(); err != nil {
		return nil, fmt.Errorf("order validation failed: %v", err)
	}

	if err := s.orderRepo.Create(ctx, newOrder); err != nil {
		return nil, fmt.Errorf("failed to create order: %v", err)
	}

	return newOrder, nil
}

// GetOrderByID retrieves an order by ID
func (s *OrderService) GetOrderByID(ctx context.Context, orderID string) (*order.Order, error) {
	return s.orderRepo.GetByID(ctx, orderID)
}

// GetOrdersByUserID retrieves orders for a specific user
func (s *OrderService) GetOrdersByUserID(ctx context.Context, userID string, limit, offset int) ([]*order.Order, error) {
	return s.orderRepo.GetByUserID(ctx, userID, limit, offset)
}

// GetAllOrders retrieves all orders with pagination
func (s *OrderService) GetAllOrders(ctx context.Context, limit, offset int) ([]*order.Order, error) {
	filter := repositories.NewFilter()
	filter.Limit = limit
	filter.Offset = offset
	return s.orderRepo.Find(ctx, filter)
}

// UpdateOrderStatus updates the status of an order
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID string, status order.OrderStatus) (*order.Order, error) {
	// Get existing order
	existingOrder, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %v", err)
	}

	// Validate status transition
	if !s.isValidStatusTransition(existingOrder.Status, status) {
		return nil, fmt.Errorf("invalid status transition from %s to %s", existingOrder.Status, status)
	}

	// Update status
	if err := s.orderRepo.UpdateStatus(ctx, orderID, status); err != nil {
		return nil, fmt.Errorf("failed to update order status: %v", err)
	}

	// Return updated order
	return s.orderRepo.GetByID(ctx, orderID)
}

// DeleteOrder deletes an order (only if it's pending or can be cancelled)
func (s *OrderService) DeleteOrder(ctx context.Context, orderID string) error {
	// Get existing order
	existingOrder, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %v", err)
	}

	// Check if order can be cancelled
	if existingOrder.Status == order.OrderStatusShipped || existingOrder.Status == order.OrderStatusDelivered {
		return fmt.Errorf("cannot cancel order with status %s", existingOrder.Status)
	}

	// Delete the order
	return s.orderRepo.Delete(ctx, orderID)
}

// isValidStatusTransition checks if a status transition is valid
func (s *OrderService) isValidStatusTransition(from, to order.OrderStatus) bool {
	validTransitions := map[order.OrderStatus][]order.OrderStatus{
		order.OrderStatusPending:   {order.OrderStatusConfirmed, order.OrderStatusCancelled},
		order.OrderStatusConfirmed: {order.OrderStatusShipped, order.OrderStatusCancelled},
		order.OrderStatusShipped:   {order.OrderStatusDelivered},
		order.OrderStatusDelivered: {}, // No transitions from delivered
		order.OrderStatusCancelled: {}, // No transitions from cancelled
	}

	allowedTransitions, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == to {
			return true
		}
	}
	return false
}
