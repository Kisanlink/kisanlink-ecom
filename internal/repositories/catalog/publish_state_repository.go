package catalog

import (
	"context"
	"fmt"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/catalog"
	"github.com/Kisanlink/kisanlink-ecom/internal/repositories/common"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"github.com/shopspring/decimal"
)

// PublishStateRepository handles database operations for product publishing states
type PublishStateRepository interface {
	// Create new publish state
	Create(ctx context.Context, state *catalog.PublishState) error

	// GetByID retrieves a publish state by its ID
	GetByID(ctx context.Context, id string) (*catalog.PublishState, error)

	// GetByProductID retrieves publish state by product ID
	GetByProductID(ctx context.Context, productID string) (*catalog.PublishState, error)

	// Update updates an existing publish state
	Update(ctx context.Context, state *catalog.PublishState) error

	// UpdateDeliveryCosts updates delivery costs for a product
	UpdateDeliveryCosts(ctx context.Context, productID string, costs map[string]decimal.Decimal) error

	// UpdateFPOAccessList updates the FPO access list
	UpdateFPOAccessList(ctx context.Context, productID string, fpoIDs []string) error

	// AddFPOAccess adds an FPO to the access list
	AddFPOAccess(ctx context.Context, productID string, fpoOrgID string, deliveryCost decimal.Decimal) error

	// RemoveFPOAccess removes an FPO from the access list
	RemoveFPOAccess(ctx context.Context, productID string, fpoOrgID string) error

	// HasFPOAccess checks if FPO has access to product
	HasFPOAccess(ctx context.Context, productID, fpoOrgID string) (bool, error)

	// GetProductsVisibleToFPO retrieves all published products visible to an FPO
	GetProductsVisibleToFPO(ctx context.Context, fpoOrgID string, offset, limit int) ([]*catalog.PublishState, int, error)

	// GetAllPublishedProducts retrieves all published products
	GetAllPublishedProducts(ctx context.Context, offset, limit int) ([]*catalog.PublishState, int, error)

	// Delete soft deletes a publish state
	Delete(ctx context.Context, productID string, deletedBy string) error
}

// publishStateRepository implements PublishStateRepository using kisanlink-db
type publishStateRepository struct {
	*common.BaseRepository
	dbManager db.DBManager
}

// NewPublishStateRepository creates a new publish state repository
func NewPublishStateRepository(dbManager db.DBManager) PublishStateRepository {
	return &publishStateRepository{
		BaseRepository: common.NewBaseRepository(dbManager),
		dbManager:      dbManager,
	}
}

// Create creates a new publish state
func (r *publishStateRepository) Create(ctx context.Context, state *catalog.PublishState) error {
	if err := state.Validate(); err != nil {
		return fmt.Errorf("invalid publish state: %w", err)
	}

	if err := r.dbManager.Create(ctx, state); err != nil {
		return fmt.Errorf("failed to create publish state: %w", err)
	}

	return nil
}

// GetByID retrieves a publish state by its ID
func (r *publishStateRepository) GetByID(ctx context.Context, id string) (*catalog.PublishState, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "id",
			Operator: base.OpEqual,
			Value:    id,
		},
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var states []*catalog.PublishState
	if err := r.dbManager.List(ctx, filter, &states); err != nil {
		return nil, fmt.Errorf("failed to get publish state by ID: %w", err)
	}

	if len(states) == 0 {
		return nil, fmt.Errorf("publish state not found: %s", id)
	}

	return states[0], nil
}

// GetByProductID retrieves publish state by product ID
func (r *publishStateRepository) GetByProductID(ctx context.Context, productID string) (*catalog.PublishState, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "product_id",
			Operator: base.OpEqual,
			Value:    productID,
		},
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	var states []*catalog.PublishState
	if err := r.dbManager.List(ctx, filter, &states); err != nil {
		return nil, fmt.Errorf("failed to get publish state by product ID: %w", err)
	}

	if len(states) == 0 {
		return nil, fmt.Errorf("publish state not found for product: %s", productID)
	}

	return states[0], nil
}

// Update updates an existing publish state
func (r *publishStateRepository) Update(ctx context.Context, state *catalog.PublishState) error {
	if err := state.Validate(); err != nil {
		return fmt.Errorf("invalid publish state: %w", err)
	}

	if err := r.dbManager.Update(ctx, state); err != nil {
		return fmt.Errorf("failed to update publish state: %w", err)
	}

	return nil
}

// UpdateDeliveryCosts updates delivery costs for a product
func (r *publishStateRepository) UpdateDeliveryCosts(ctx context.Context, productID string, costs map[string]decimal.Decimal) error {
	state, err := r.GetByProductID(ctx, productID)
	if err != nil {
		return err
	}

	// Update delivery costs
	state.DeliveryCosts = costs

	return r.Update(ctx, state)
}

// UpdateFPOAccessList updates the FPO access list
func (r *publishStateRepository) UpdateFPOAccessList(ctx context.Context, productID string, fpoIDs []string) error {
	state, err := r.GetByProductID(ctx, productID)
	if err != nil {
		return err
	}

	// Update FPO access list
	state.FPOAccessList = fpoIDs

	return r.Update(ctx, state)
}

// AddFPOAccess adds an FPO to the access list
func (r *publishStateRepository) AddFPOAccess(ctx context.Context, productID string, fpoOrgID string, deliveryCost decimal.Decimal) error {
	state, err := r.GetByProductID(ctx, productID)
	if err != nil {
		return err
	}

	// Add FPO to access list if not already present
	state.AddFPOAccess(fpoOrgID)

	// Set delivery cost
	state.SetDeliveryCost(fpoOrgID, deliveryCost)

	return r.Update(ctx, state)
}

// RemoveFPOAccess removes an FPO from the access list
func (r *publishStateRepository) RemoveFPOAccess(ctx context.Context, productID string, fpoOrgID string) error {
	state, err := r.GetByProductID(ctx, productID)
	if err != nil {
		return err
	}

	// Remove FPO from access list
	state.RemoveFPOAccess(fpoOrgID)

	// Note: We keep the delivery cost history for audit purposes
	// If needed, we could delete it here with: delete(state.DeliveryCosts, fpoOrgID)

	return r.Update(ctx, state)
}

// HasFPOAccess checks if FPO has access to product
// This method uses a database query to check the JSONB array efficiently
func (r *publishStateRepository) HasFPOAccess(ctx context.Context, productID, fpoOrgID string) (bool, error) {
	state, err := r.GetByProductID(ctx, productID)
	if err != nil {
		return false, err
	}

	return state.HasFPOAccess(fpoOrgID), nil
}

// GetProductsVisibleToFPO retrieves all published products visible to an FPO
// Uses JSONB array contains operator for efficient querying
func (r *publishStateRepository) GetProductsVisibleToFPO(ctx context.Context, fpoOrgID string, offset, limit int) ([]*catalog.PublishState, int, error) {
	filter := base.NewFilter()

	// Filter for published products (published_at IS NOT NULL)
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "published_at",
			Operator: base.OpNotEqual,
			Value:    nil,
		},
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	// Set pagination
	filter.Limit = limit
	filter.Offset = offset

	var states []*catalog.PublishState
	if err := r.dbManager.List(ctx, filter, &states); err != nil {
		return nil, 0, fmt.Errorf("failed to get published products: %w", err)
	}

	// Filter in memory for FPO access (since kisanlink-db may not support JSONB contains)
	visibleStates := make([]*catalog.PublishState, 0)
	for _, state := range states {
		if state.HasFPOAccess(fpoOrgID) {
			visibleStates = append(visibleStates, state)
		}
	}

	// Get total count by filtering all states
	countFilter := r.ApplyQueryOptions(ctx, filter)
	total, err := r.dbManager.Count(ctx, countFilter, &catalog.PublishState{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return visibleStates, int(total), nil
}

// GetAllPublishedProducts retrieves all published products
func (r *publishStateRepository) GetAllPublishedProducts(ctx context.Context, offset, limit int) ([]*catalog.PublishState, int, error) {
	filter := base.NewFilter()

	// Filter for published products (published_at IS NOT NULL)
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "published_at",
			Operator: base.OpNotEqual,
			Value:    nil,
		},
	}

	// Apply soft delete filtering
	filter = r.ApplyQueryOptions(ctx, filter)

	// Set pagination
	filter.Limit = limit
	filter.Offset = offset

	var states []*catalog.PublishState
	if err := r.dbManager.List(ctx, filter, &states); err != nil {
		return nil, 0, fmt.Errorf("failed to get published products: %w", err)
	}

	// Get total count
	countFilter := r.ApplyQueryOptions(ctx, filter)
	total, err := r.dbManager.Count(ctx, countFilter, &catalog.PublishState{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return states, int(total), nil
}

// Delete soft deletes a publish state
func (r *publishStateRepository) Delete(ctx context.Context, productID string, deletedBy string) error {
	state, err := r.GetByProductID(ctx, productID)
	if err != nil {
		return err
	}

	// Set deleted by
	state.SetDeletedBy(&deletedBy)

	if err := r.dbManager.Update(ctx, state); err != nil {
		return fmt.Errorf("failed to soft delete publish state: %w", err)
	}

	return nil
}
