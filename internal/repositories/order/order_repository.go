package order

import (
	"context"

	"kisanlink-ecom/entities/models/orders"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// OrderRepository extends BaseFilterableRepository with order-specific methods
type OrderRepository struct {
	*base.BaseFilterableRepository[*orders.Order]
	dbManager db.DBManager
}

// NewOrderRepository creates a new order repository
func NewOrderRepository(dbManager db.DBManager) *OrderRepository {
	baseRepo := base.NewBaseFilterableRepository[*orders.Order]()
	baseRepo.SetDBManager(dbManager)
	return &OrderRepository{
		BaseFilterableRepository: baseRepo,
		dbManager:                dbManager,
	}
}

// GetByBuyerID retrieves orders by buyer ID
func (r *OrderRepository) GetByBuyerID(ctx context.Context, buyerID string, limit, offset int) ([]*orders.Order, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "buyer_id",
			Operator: base.OpEqual,
			Value:    buyerID,
		},
	}
	filter.Limit = limit
	filter.Offset = offset

	return r.Find(ctx, filter)
}

// GetBySellerID retrieves orders by seller ID
func (r *OrderRepository) GetBySellerID(ctx context.Context, sellerID string, limit, offset int) ([]*orders.Order, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "seller_id",
			Operator: base.OpEqual,
			Value:    sellerID,
		},
	}
	filter.Limit = limit
	filter.Offset = offset

	return r.Find(ctx, filter)
}

// GetByStatus retrieves orders by status
func (r *OrderRepository) GetByStatus(ctx context.Context, status orders.OrderStatus, limit, offset int) ([]*orders.Order, error) {
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

	return r.Find(ctx, filter)
}

// CreateOrderWithItems creates an order with its items in a transaction
func (r *OrderRepository) CreateOrderWithItems(ctx context.Context, order *orders.Order) error {
	// Get PostgreSQL manager for transaction support
	if pgManager, ok := r.dbManager.(interface {
		WithTransaction(ctx context.Context, fn func(tx interface{}) error) error
	}); ok {
		return pgManager.WithTransaction(ctx, func(tx interface{}) error {
			// Create the order first
			if err := r.Create(ctx, order); err != nil {
				return err
			}

			// Create order items
			for i := range order.Items {
				order.Items[i].OrderID = order.ID
				if err := r.createOrderItem(ctx, &order.Items[i]); err != nil {
					return err
				}
			}

			return nil
		})
	}

	// Fallback for non-PostgreSQL databases
	return r.createOrderWithoutTransaction(ctx, order)
}

// createOrderItem creates a single order item
func (r *OrderRepository) createOrderItem(ctx context.Context, item *orders.OrderItem) error {
	// This would use a separate OrderItemRepository in a real implementation
	// For now, we'll use the base repository methods
	return r.dbManager.Create(ctx, item)
}

// createOrderWithoutTransaction creates order without transaction support
func (r *OrderRepository) createOrderWithoutTransaction(ctx context.Context, order *orders.Order) error {
	// Create the order first
	if err := r.Create(ctx, order); err != nil {
		return err
	}

	// Create order items
	for i := range order.Items {
		order.Items[i].OrderID = order.ID
		if err := r.createOrderItem(ctx, &order.Items[i]); err != nil {
			return err
		}
	}

	return nil
}
