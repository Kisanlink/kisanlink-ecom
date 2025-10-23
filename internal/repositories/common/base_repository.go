// Package common provides shared repository functionality including soft delete filtering.
package common

import (
	"context"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// BaseRepository provides common functionality for all repositories
type BaseRepository struct {
	dbManager db.DBManager
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(dbManager db.DBManager) *BaseRepository {
	return &BaseRepository{
		dbManager: dbManager,
	}
}

// ApplyQueryOptions applies query options to a filter
// This method modifies the filter to include or exclude soft-deleted items
// based on the QueryOptions in the context
func (r *BaseRepository) ApplyQueryOptions(ctx context.Context, filter *base.Filter) *base.Filter {
	if filter == nil {
		filter = base.NewFilter()
	}

	// Get query options from context
	opts := QueryOptionsFromContext(ctx)

	// If IncludeDeleted is false (default), add a condition to filter out deleted items
	// GORM's soft delete uses deleted_at IS NULL for non-deleted items
	if !opts.IncludeDeleted {
		// Add condition to filter out soft-deleted items
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "deleted_at",
			Operator: base.OpIsNull,
			Value:    nil,
		})
	}

	// Future: Apply tenant filtering
	if opts.TenantID != nil {
		filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
			Field:    "organization_id",
			Operator: base.OpEqual,
			Value:    *opts.TenantID,
		})
	}

	return filter
}

// GetDBManager returns the underlying database manager
func (r *BaseRepository) GetDBManager() db.DBManager {
	return r.dbManager
}

// ExecuteWithQueryOptions executes a database operation with query options applied
// This is a helper method for operations that need to respect query options
func (r *BaseRepository) ExecuteWithQueryOptions(ctx context.Context, filter *base.Filter, operation func(*base.Filter) error) error {
	enhancedFilter := r.ApplyQueryOptions(ctx, filter)
	return operation(enhancedFilter)
}

// CountWithQueryOptions executes a count operation with query options applied
func (r *BaseRepository) CountWithQueryOptions(ctx context.Context, filter *base.Filter, model interface{}) (int64, error) {
	enhancedFilter := r.ApplyQueryOptions(ctx, filter)
	return r.dbManager.Count(ctx, enhancedFilter, model)
}

// ListWithQueryOptions executes a list operation with query options applied
func (r *BaseRepository) ListWithQueryOptions(ctx context.Context, filter *base.Filter, models interface{}) error {
	enhancedFilter := r.ApplyQueryOptions(ctx, filter)
	return r.dbManager.List(ctx, enhancedFilter, models)
}

// GetByIDWithQueryOptions retrieves a single record by ID with query options applied
// Note: For GetByID operations, we need to handle soft deletes differently
// If IncludeDeleted is false, we'll get the record and check if it's deleted
func (r *BaseRepository) GetByIDWithQueryOptions(ctx context.Context, id string, model interface{}) error {
	opts := QueryOptionsFromContext(ctx)

	if opts.IncludeDeleted {
		// Use the special method to include deleted records
		return r.dbManager.GetByID(ctx, id, model)
	}

	// For non-deleted records, we need to get by ID and then verify it's not deleted
	// Unfortunately, kisanlink-db's GetByID doesn't support filters
	// So we'll use List with a filter
	filter := base.NewFilter()
	filter.Group.Conditions = append(filter.Group.Conditions, base.FilterCondition{
		Field:    "id",
		Operator: base.OpEqual,
		Value:    id,
	})

	enhancedFilter := r.ApplyQueryOptions(ctx, filter)
	return r.dbManager.List(ctx, enhancedFilter, model)
}
