package repositories

import (
	"context"

	"kisanlink-ecom/internal/models/product"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// ProductRepository extends BaseFilterableRepository with Product-specific methods
type ProductRepository struct {
	*base.BaseFilterableRepository[*product.Product]
}

// NewProductRepository creates a new product repository
func NewProductRepository() *ProductRepository {
	return &ProductRepository{
		BaseFilterableRepository: base.NewBaseFilterableRepository[*product.Product](),
	}
}

// GetByCategory retrieves products by category
func (r *ProductRepository) GetByCategory(ctx context.Context, category string, limit, offset int) ([]*product.Product, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "category",
			Operator: base.OpEqual,
			Value:    category,
		},
	}
	filter.Limit = limit
	filter.Offset = offset

	return r.Find(ctx, filter)
}

// GetByPriceRange retrieves products within a price range
func (r *ProductRepository) GetByPriceRange(ctx context.Context, minPrice, maxPrice float64, limit, offset int) ([]*product.Product, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "price",
			Operator: base.OpGreaterEqual,
			Value:    minPrice,
		},
		{
			Field:    "price",
			Operator: base.OpLessEqual,
			Value:    maxPrice,
		},
	}
	filter.Limit = limit
	filter.Offset = offset

	return r.Find(ctx, filter)
}

// GetLowStock retrieves products with low stock
func (r *ProductRepository) GetLowStock(ctx context.Context, threshold int, limit, offset int) ([]*product.Product, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "stock",
			Operator: base.OpLessEqual,
			Value:    threshold,
		},
	}
	filter.Limit = limit
	filter.Offset = offset

	return r.Find(ctx, filter)
}
