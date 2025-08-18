package customer

import (
	"context"

	"kisanlink-ecom/internal/models/user"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// CustomerRepository extends BaseFilterableRepository with Customer-specific methods
type CustomerRepository struct {
	*base.BaseFilterableRepository[*user.Customer]
}

// NewCustomerRepository creates a new customer repository
func NewCustomerRepository() *CustomerRepository {
	return &CustomerRepository{
		BaseFilterableRepository: base.NewBaseFilterableRepository[*user.Customer](),
	}
}

// GetByCustomerCode retrieves a customer by customer code
func (r *CustomerRepository) GetByCustomerCode(ctx context.Context, customerCode string) (*user.Customer, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "customer_code",
			Operator: base.OpEqual,
			Value:    customerCode,
		},
	}

	customers, err := r.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(customers) == 0 {
		return nil, nil
	}

	return customers[0], nil
}

// GetByAAAEntityID retrieves a customer by aaa-service entity ID
func (r *CustomerRepository) GetByAAAEntityID(ctx context.Context, aaaEntityID string) (*user.Customer, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "aaa_entity_id",
			Operator: base.OpEqual,
			Value:    aaaEntityID,
		},
	}

	customers, err := r.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(customers) == 0 {
		return nil, nil
	}

	return customers[0], nil
}

// GetByStatus retrieves customers by status
func (r *CustomerRepository) GetByStatus(ctx context.Context, status string, limit, offset int) ([]*user.Customer, error) {
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

// GetVerifiedCustomers retrieves verified customers
func (r *CustomerRepository) GetVerifiedCustomers(ctx context.Context, limit, offset int) ([]*user.Customer, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "is_verified",
			Operator: base.OpEqual,
			Value:    true,
		},
	}
	filter.Limit = limit
	filter.Offset = offset

	return r.Find(ctx, filter)
}
