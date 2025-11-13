package orders

import (
	"context"
	"fmt"
	"time"

	"kisanlink-ecom/entities/models/orders"
	ordersRequests "kisanlink-ecom/entities/requests/orders"
	repositoryCommon "kisanlink-ecom/internal/repositories/common"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// PurchaseOrderRepository extends BaseFilterableRepository with purchase order-specific methods
type PurchaseOrderRepository struct {
	*base.BaseFilterableRepository[*orders.PurchaseOrder]
	*repositoryCommon.BaseRepository
	dbManager db.DBManager
}

// NewPurchaseOrderRepository creates a new purchase order repository
func NewPurchaseOrderRepository(dbManager db.DBManager) *PurchaseOrderRepository {
	baseRepo := base.NewBaseFilterableRepository[*orders.PurchaseOrder]()
	baseRepo.SetDBManager(dbManager)
	return &PurchaseOrderRepository{
		BaseFilterableRepository: baseRepo,
		BaseRepository:           repositoryCommon.NewBaseRepository(dbManager),
		dbManager:                dbManager,
	}
}

// CreatePurchaseOrder creates a new purchase order with items in a transaction
func (r *PurchaseOrderRepository) CreatePurchaseOrder(ctx context.Context, po *orders.PurchaseOrder) error {
	// Validate purchase order before creation
	if po == nil {
		return fmt.Errorf("purchase order cannot be nil")
	}
	if po.FPOOrgID == "" {
		return fmt.Errorf("FPO organization ID is required")
	}
	if po.VendorName == "" {
		return fmt.Errorf("vendor name is required")
	}
	if len(po.Items) == 0 {
		return fmt.Errorf("purchase order must have at least one item")
	}

	// Check if database manager supports transactions
	if txManager, ok := r.dbManager.(interface {
		WithTransaction(ctx context.Context, fn func(tx any) error) error
	}); ok {
		return txManager.WithTransaction(ctx, func(tx any) error {
			return r.createPurchaseOrderWithTransaction(ctx, po, tx)
		})
	}

	// Fallback for databases without transaction support
	return r.createPurchaseOrderWithoutTransaction(ctx, po)
}

// createPurchaseOrderWithTransaction creates PO and items within a transaction
func (r *PurchaseOrderRepository) createPurchaseOrderWithTransaction(ctx context.Context, po *orders.PurchaseOrder, _ any) error {
	// Calculate total before creating
	po.CalculateTotal()

	// Set POID for all items before creating (GORM associations require this)
	for i := range po.Items {
		po.Items[i].POID = po.ID
	}

	// Create the purchase order with all its items via GORM associations
	if err := r.Create(ctx, po); err != nil {
		return fmt.Errorf("failed to create purchase order: %w", err)
	}

	return nil
}

// createPurchaseOrderWithoutTransaction creates PO without transaction support
func (r *PurchaseOrderRepository) createPurchaseOrderWithoutTransaction(ctx context.Context, po *orders.PurchaseOrder) error {
	// Calculate total before creating
	po.CalculateTotal()

	// Set POID for all items before creating
	for i := range po.Items {
		po.Items[i].POID = po.ID
	}

	// Create the purchase order with all its items via GORM associations
	if err := r.Create(ctx, po); err != nil {
		return fmt.Errorf("failed to create purchase order: %w", err)
	}

	return nil
}

// GetPurchaseOrderByID retrieves a purchase order by ID with items and GRNs
func (r *PurchaseOrderRepository) GetPurchaseOrderByID(ctx context.Context, id string) (*orders.PurchaseOrder, error) {
	po := &orders.PurchaseOrder{}
	retrievedPO, err := r.BaseFilterableRepository.GetByID(ctx, id, po)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase order: %w", err)
	}
	po = retrievedPO

	// Load PO items
	items, err := r.GetPOItems(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load PO items: %w", err)
	}
	po.Items = make([]orders.POItem, len(items))
	for i, item := range items {
		po.Items[i] = *item
	}

	// Load GRNs
	grns, err := r.GetGRNsByPO(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load GRNs: %w", err)
	}
	po.GRNs = make([]orders.GRN, len(grns))
	for i, grn := range grns {
		po.GRNs[i] = *grn
	}

	return po, nil
}

// GetPurchaseOrderByNumber retrieves a purchase order by PO number
func (r *PurchaseOrderRepository) GetPurchaseOrderByNumber(ctx context.Context, poNumber string) (*orders.PurchaseOrder, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "po_number",
			Operator: base.OpEqual,
			Value:    poNumber,
		},
	}

	poList, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(poList) == 0 {
		return nil, fmt.Errorf("purchase order not found")
	}

	return poList[0], nil
}

// UpdatePurchaseOrderStatus updates the status of a purchase order
func (r *PurchaseOrderRepository) UpdatePurchaseOrderStatus(ctx context.Context, id string, status orders.POStatus, notes string) error {
	existingPO := &orders.PurchaseOrder{}
	retrievedPO, err := r.BaseFilterableRepository.GetByID(ctx, id, existingPO)
	if err != nil {
		return err
	}
	existingPO = retrievedPO

	// Validate status transition
	if !existingPO.CanTransitionTo(status) {
		return fmt.Errorf("invalid status transition from %s to %s", existingPO.Status, status)
	}

	// Update status
	existingPO.Status = status
	if notes != "" {
		existingPO.Notes = notes
	}

	// Save the updated purchase order
	return r.BaseFilterableRepository.Update(ctx, existingPO)
}

// ListPurchaseOrders retrieves purchase orders with filtering and pagination
func (r *PurchaseOrderRepository) ListPurchaseOrders(ctx context.Context, filter *ordersRequests.ListPurchaseOrdersRequest, fpoOrgID string, offset, limit int) ([]*orders.PurchaseOrder, int, error) {
	// Validate input parameters
	if filter == nil {
		return nil, 0, fmt.Errorf("filter cannot be nil")
	}
	if fpoOrgID == "" {
		return nil, 0, fmt.Errorf("FPO organization ID is required")
	}
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if offset < 0 {
		offset = 0
	}

	dbFilter := base.NewFilter()

	// Always filter by FPO organization
	dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
		Field:    "fpo_org_id",
		Operator: base.OpEqual,
		Value:    fpoOrgID,
	})

	// Add conditions based on the filter
	if filter.Status != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "status",
			Operator: base.OpEqual,
			Value:    string(*filter.Status),
		})
	}

	if filter.Source != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "source",
			Operator: base.OpEqual,
			Value:    string(*filter.Source),
		})
	}

	if filter.VendorID != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "vendor_id",
			Operator: base.OpEqual,
			Value:    *filter.VendorID,
		})
	}

	// Date range filtering
	if filter.CreatedAfter != nil {
		if createdAfter, err := time.Parse(time.RFC3339, *filter.CreatedAfter); err == nil {
			dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
				Field:    "created_at",
				Operator: base.OpGreaterEqual,
				Value:    createdAfter,
			})
		} else {
			return nil, 0, fmt.Errorf("invalid created_after date format: %w", err)
		}
	}

	if filter.CreatedBefore != nil {
		if createdBefore, err := time.Parse(time.RFC3339, *filter.CreatedBefore); err == nil {
			dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
				Field:    "created_at",
				Operator: base.OpLessEqual,
				Value:    createdBefore,
			})
		} else {
			return nil, 0, fmt.Errorf("invalid created_before date format: %w", err)
		}
	}

	// Search functionality
	if filter.Search != nil && *filter.Search != "" {
		searchConditions := []base.FilterCondition{
			{
				Field:    "po_number",
				Operator: base.OpLike,
				Value:    "%" + *filter.Search + "%",
			},
			{
				Field:    "vendor_name",
				Operator: base.OpLike,
				Value:    "%" + *filter.Search + "%",
			},
			{
				Field:    "notes",
				Operator: base.OpLike,
				Value:    "%" + *filter.Search + "%",
			},
		}

		// Create OR group for search conditions
		searchGroup := base.FilterGroup{
			Logic:      base.LogicOr,
			Conditions: searchConditions,
		}

		// Wrap existing conditions with search
		if len(dbFilter.Group.Conditions) > 0 {
			existingGroup := dbFilter.Group
			dbFilter.Group = base.FilterGroup{
				Logic:  base.LogicAnd,
				Groups: []base.FilterGroup{existingGroup, searchGroup},
			}
		} else {
			dbFilter.Group = searchGroup
		}
	}

	// Pagination
	dbFilter.Limit = limit
	dbFilter.Offset = offset

	// Get purchase orders
	poList, err := r.BaseFilterableRepository.Find(ctx, dbFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find purchase orders: %w", err)
	}

	// Load related data if requested
	if filter.IncludeItems != nil && *filter.IncludeItems {
		for i, po := range poList {
			items, err := r.GetPOItems(ctx, po.ID)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to load items for PO %s: %w", po.ID, err)
			}
			poList[i].Items = make([]orders.POItem, len(items))
			for j, item := range items {
				poList[i].Items[j] = *item
			}
		}
	}

	if filter.IncludeGRNs != nil && *filter.IncludeGRNs {
		for i, po := range poList {
			grns, err := r.GetGRNsByPO(ctx, po.ID)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to load GRNs for PO %s: %w", po.ID, err)
			}
			poList[i].GRNs = make([]orders.GRN, len(grns))
			for j, grn := range grns {
				poList[i].GRNs[j] = *grn
			}
		}
	}

	// Get total count for pagination
	countFilter := *dbFilter
	countFilter.Limit = 0
	countFilter.Offset = 0
	total, err := r.BaseFilterableRepository.CountWithFilter(ctx, &countFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return poList, int(total), nil
}

// GetPOItems retrieves items for a specific purchase order
func (r *PurchaseOrderRepository) GetPOItems(ctx context.Context, poID string) ([]*orders.POItem, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "po_id",
			Operator: base.OpEqual,
			Value:    poID,
		},
	}

	// Apply query options (adds deleted_at IS NULL by default)
	filter = r.ApplyQueryOptions(ctx, filter)

	var itemList []*orders.POItem
	if err := r.dbManager.List(ctx, filter, &itemList); err != nil {
		return nil, fmt.Errorf("failed to get PO items: %w", err)
	}

	return itemList, nil
}

// CreateGRN creates a new goods received note
func (r *PurchaseOrderRepository) CreateGRN(ctx context.Context, grn *orders.GRN) error {
	// Validate GRN before creation
	if grn == nil {
		return fmt.Errorf("GRN cannot be nil")
	}
	if grn.POID == "" {
		return fmt.Errorf("PO ID is required")
	}
	if grn.ReceivedBy == "" {
		return fmt.Errorf("received by user ID is required")
	}
	if len(grn.Items) == 0 {
		return fmt.Errorf("GRN must have at least one item")
	}

	// Check if database manager supports transactions
	if txManager, ok := r.dbManager.(interface {
		WithTransaction(ctx context.Context, fn func(tx any) error) error
	}); ok {
		return txManager.WithTransaction(ctx, func(tx any) error {
			return r.createGRNWithTransaction(ctx, grn, tx)
		})
	}

	// Fallback for databases without transaction support
	return r.createGRNWithoutTransaction(ctx, grn)
}

// createGRNWithTransaction creates GRN and items within a transaction
func (r *PurchaseOrderRepository) createGRNWithTransaction(ctx context.Context, grn *orders.GRN, _ any) error {
	// Set GRNID for all items before creating
	for i := range grn.Items {
		grn.Items[i].GRNID = grn.ID
	}

	// Create the GRN with all its items via GORM associations
	if err := r.dbManager.Create(ctx, grn); err != nil {
		return fmt.Errorf("failed to create GRN: %w", err)
	}

	return nil
}

// createGRNWithoutTransaction creates GRN without transaction support
func (r *PurchaseOrderRepository) createGRNWithoutTransaction(ctx context.Context, grn *orders.GRN) error {
	// Set GRNID for all items before creating
	for i := range grn.Items {
		grn.Items[i].GRNID = grn.ID
	}

	// Create the GRN with all its items via GORM associations
	if err := r.dbManager.Create(ctx, grn); err != nil {
		return fmt.Errorf("failed to create GRN: %w", err)
	}

	return nil
}

// GetGRNByID retrieves a GRN by ID with items
func (r *PurchaseOrderRepository) GetGRNByID(ctx context.Context, id string) (*orders.GRN, error) {
	grn := &orders.GRN{}
	if err := r.dbManager.GetByID(ctx, id, grn); err != nil {
		return nil, fmt.Errorf("failed to get GRN: %w", err)
	}

	// Load GRN items
	items, err := r.GetGRNItems(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load GRN items: %w", err)
	}
	grn.Items = make([]orders.GRNItem, len(items))
	for i, item := range items {
		grn.Items[i] = *item
	}

	return grn, nil
}

// GetGRNsByPO retrieves all GRNs for a purchase order
func (r *PurchaseOrderRepository) GetGRNsByPO(ctx context.Context, poID string) ([]*orders.GRN, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "po_id",
			Operator: base.OpEqual,
			Value:    poID,
		},
	}

	// Apply query options (adds deleted_at IS NULL by default)
	filter = r.ApplyQueryOptions(ctx, filter)

	var grnList []*orders.GRN
	if err := r.dbManager.List(ctx, filter, &grnList); err != nil {
		return nil, fmt.Errorf("failed to get GRNs: %w", err)
	}

	return grnList, nil
}

// GetGRNItems retrieves items for a specific GRN
func (r *PurchaseOrderRepository) GetGRNItems(ctx context.Context, grnID string) ([]*orders.GRNItem, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "grn_id",
			Operator: base.OpEqual,
			Value:    grnID,
		},
	}

	// Apply query options (adds deleted_at IS NULL by default)
	filter = r.ApplyQueryOptions(ctx, filter)

	var itemList []*orders.GRNItem
	if err := r.dbManager.List(ctx, filter, &itemList); err != nil {
		return nil, fmt.Errorf("failed to get GRN items: %w", err)
	}

	return itemList, nil
}

// UpdateGRNStatus updates the status of a GRN
func (r *PurchaseOrderRepository) UpdateGRNStatus(ctx context.Context, id string, status orders.GRNStatus, notes string) error {
	grn := &orders.GRN{}
	if err := r.dbManager.GetByID(ctx, id, grn); err != nil {
		return fmt.Errorf("failed to get GRN: %w", err)
	}

	// Update status
	grn.Status = status
	if notes != "" {
		grn.Notes = notes
	}

	// Save the updated GRN
	return r.dbManager.Update(ctx, grn)
}
