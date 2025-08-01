package repositories

import (
	"context"

	"kisanlink-ecom/internal/models/order"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// OrderRepository extends BaseFilterableRepository with Order-specific methods
type OrderRepository struct {
	*base.BaseFilterableRepository[*order.Order]
}

// NewOrderRepository creates a new order repository
func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		BaseFilterableRepository: base.NewBaseFilterableRepository[*order.Order](),
	}
}

// GetByUserID retrieves orders by user ID
func (r *OrderRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*order.Order, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "user_id",
			Operator: base.OpEqual,
			Value:    userID,
		},
	}
	filter.Limit = limit
	filter.Offset = offset

	return r.Find(ctx, filter)
}

// GetByStatus retrieves orders by status
func (r *OrderRepository) GetByStatus(ctx context.Context, status order.OrderStatus, limit, offset int) ([]*order.Order, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "status",
			Operator: base.OpEqual,
			Value:    status,
		},
	}
	filter.Limit = limit
	filter.Offset = offset

	return r.Find(ctx, filter)
}

// GetByDateRange retrieves orders within a date range
func (r *OrderRepository) GetByDateRange(ctx context.Context, startDate, endDate string, limit, offset int) ([]*order.Order, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "created_at",
			Operator: base.OpDateBetween,
			Value:    startDate,
			Value2:   endDate,
		},
	}
	filter.Limit = limit
	filter.Offset = offset

	return r.Find(ctx, filter)
}

// UpdateStatus updates the status of an order
func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID string, status order.OrderStatus) error {
	order, err := r.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	order.Status = status
	return r.Update(ctx, order)
}
