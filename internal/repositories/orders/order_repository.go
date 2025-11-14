package orders

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/orders"
	ordersRequests "github.com/Kisanlink/kisanlink-ecom/entities/requests/orders"
	repositoryCommon "github.com/Kisanlink/kisanlink-ecom/internal/repositories/common"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"github.com/shopspring/decimal"
)

// OrderRepository extends BaseFilterableRepository with order-specific methods
type OrderRepository struct {
	*base.BaseFilterableRepository[*orders.Order]
	*repositoryCommon.BaseRepository
	dbManager     db.DBManager
	catalogRepo   CatalogRepository
	inventoryRepo InventoryRepository
}

// CatalogRepository interface for catalog operations
type CatalogRepository interface {
	GetByID(ctx context.Context, id string, model any) (any, error)
	GetCatalogItemOwner(ctx context.Context, id string) (string, error)
}

// InventoryRepository interface for inventory operations
type InventoryRepository interface {
	GetInventoryLevel(ctx context.Context, itemID string) (decimal.Decimal, error)
	ReserveInventory(ctx context.Context, itemID string, quantity decimal.Decimal) error
	ReleaseInventory(ctx context.Context, itemID string, quantity decimal.Decimal) error
}

// NewOrderRepository creates a new order repository
func NewOrderRepository(dbManager db.DBManager) *OrderRepository {
	baseRepo := base.NewBaseFilterableRepository[*orders.Order]()
	baseRepo.SetDBManager(dbManager)
	return &OrderRepository{
		BaseFilterableRepository: baseRepo,
		BaseRepository:           repositoryCommon.NewBaseRepository(dbManager),
		dbManager:                dbManager,
	}
}

// SetCatalogRepository sets the catalog repository dependency
func (r *OrderRepository) SetCatalogRepository(catalogRepo CatalogRepository) {
	r.catalogRepo = catalogRepo
}

// SetInventoryRepository sets the inventory repository dependency
func (r *OrderRepository) SetInventoryRepository(inventoryRepo InventoryRepository) {
	r.inventoryRepo = inventoryRepo
}

// Note: Basic CRUD operations (Create, GetByID, Update, Delete, Find) are inherited from BaseFilterableRepository

// CreateOrder creates a new order with items in a transaction
func (r *OrderRepository) CreateOrder(ctx context.Context, ord *orders.Order) error {
	// Validate order before creation
	if ord == nil {
		return fmt.Errorf("order cannot be nil")
	}
	if ord.BuyerOrganizationID == "" {
		return fmt.Errorf("buyer organization ID is required")
	}
	if ord.SellerOrganizationID == "" {
		return fmt.Errorf("seller organization ID is required")
	}
	if ord.BuyerUserID == "" {
		return fmt.Errorf("buyer user ID is required")
	}
	if len(ord.Items) == 0 {
		return fmt.Errorf("order must have at least one item")
	}

	// Check if database manager supports transactions
	if txManager, ok := r.dbManager.(interface {
		WithTransaction(ctx context.Context, fn func(tx any) error) error
	}); ok {
		return txManager.WithTransaction(ctx, func(tx any) error {
			return r.createOrderWithTransaction(ctx, ord, tx)
		})
	}

	// Fallback for databases without transaction support
	return r.createOrderWithoutTransaction(ctx, ord)
}

// createOrderWithTransaction creates order and items within a transaction
func (r *OrderRepository) createOrderWithTransaction(ctx context.Context, ord *orders.Order, _ any) error {
	// Calculate total before creating
	ord.CalculateTotal()

	// Set OrderID for all items before creating (GORM associations require this)
	for i := range ord.Items {
		ord.Items[i].OrderID = ord.ID
	}

	// Create initial status history entry before creating order
	if len(ord.StatusHistory) == 0 {
		initialHistory := orders.NewOrderStatusHistory(
			ord.ID,
			nil,
			ord.Status,
			"Order created",
			ord.BuyerUserID,
			ord.BuyerOrganizationID,
		)
		ord.StatusHistory = append(ord.StatusHistory, *initialHistory)
	}

	// Create the order with all its items and status history via GORM associations
	// r.Create() delegates to BaseFilterableRepository which handles all associated records
	if err := r.Create(ctx, ord); err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

// createOrderWithoutTransaction creates order without transaction support
func (r *OrderRepository) createOrderWithoutTransaction(ctx context.Context, ord *orders.Order) error {
	// Calculate total before creating
	ord.CalculateTotal()

	// Set OrderID for all items before creating (GORM associations require this)
	for i := range ord.Items {
		ord.Items[i].OrderID = ord.ID
	}

	// Create initial status history entry before creating order
	if len(ord.StatusHistory) == 0 {
		initialHistory := orders.NewOrderStatusHistory(
			ord.ID,
			nil,
			ord.Status,
			"Order created",
			ord.BuyerUserID,
			ord.BuyerOrganizationID,
		)
		ord.StatusHistory = append(ord.StatusHistory, *initialHistory)
	}

	// Create the order with all its items and status history via GORM associations
	// r.Create() delegates to BaseFilterableRepository which handles all associated records
	if err := r.Create(ctx, ord); err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

// CreateOrderWithItems creates an order with its items in a transaction (alias for CreateOrder)
func (r *OrderRepository) CreateOrderWithItems(ctx context.Context, order *orders.Order) error {
	return r.CreateOrder(ctx, order)
}

// GetOrderByID retrieves an order by ID with items and status history
func (r *OrderRepository) GetOrderByID(ctx context.Context, id string) (*orders.Order, error) {
	order := &orders.Order{}
	retrievedOrder, err := r.BaseFilterableRepository.GetByID(ctx, id, order)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	order = retrievedOrder

	// Load order items
	items, err := r.GetOrderItems(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load order items: %w", err)
	}
	order.Items = make([]orders.OrderItem, len(items))
	for i, item := range items {
		order.Items[i] = *item
	}

	// Load status history
	history, err := r.GetOrderHistory(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load order history: %w", err)
	}
	order.StatusHistory = make([]orders.OrderStatusHistory, len(history))
	for i, hist := range history {
		order.StatusHistory[i] = *hist
	}

	return order, nil
}

// GetOrderByNumber retrieves an order by order number
func (r *OrderRepository) GetOrderByNumber(ctx context.Context, orderNumber string) (*orders.Order, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "order_number",
			Operator: base.OpEqual,
			Value:    orderNumber,
		},
	}

	orderList, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(orderList) == 0 {
		return nil, fmt.Errorf("order not found")
	}

	return orderList[0], nil
}

// UpdateOrderStatus updates the status of an order
func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, id string, status orders.OrderStatus, expectedDelivery *string, notes string) error {
	existingOrder := &orders.Order{}
	retrievedOrder, err := r.BaseFilterableRepository.GetByID(ctx, id, existingOrder)
	if err != nil {
		return err
	}
	existingOrder = retrievedOrder

	// Validate status transition
	if !existingOrder.CanTransitionTo(status) {
		return fmt.Errorf("invalid status transition from %s to %s", existingOrder.Status, status)
	}

	// Update the order status using the model's method to create history
	err = existingOrder.UpdateStatus(status, notes, "", "") // TODO: Add proper user context
	if err != nil {
		return err
	}

	// Update expected delivery if provided
	if expectedDelivery != nil {
		if parsedTime, parseErr := time.Parse(time.RFC3339, *expectedDelivery); parseErr == nil {
			existingOrder.EstimatedDeliveryDate = &parsedTime
		}
	}

	// Save the updated order
	return r.BaseFilterableRepository.Update(ctx, existingOrder)
}

// ListOrders retrieves orders with filtering and pagination
func (r *OrderRepository) ListOrders(ctx context.Context, filter *ordersRequests.ListOrdersRequest, offset, limit int) ([]*orders.Order, int, error) {
	// Validate input parameters
	if filter == nil {
		return nil, 0, fmt.Errorf("filter cannot be nil")
	}
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if offset < 0 {
		offset = 0
	}

	dbFilter := base.NewFilter()

	// Debug logging
	isAdminPtr := "nil"
	isAdminValue := false
	if filter.IsAdmin != nil {
		isAdminPtr = fmt.Sprintf("%v", *filter.IsAdmin)
		isAdminValue = *filter.IsAdmin
	}
	fmt.Printf("[DEBUG] OrderRepository.ListOrders: filter_is_admin_ptr=%s, filter_is_admin_value=%v, filter_buyer_org=%v, filter_seller_org=%v\n",
		isAdminPtr, isAdminValue, filter.BuyerOrganizationID, filter.SellerOrganizationID)

	// Add conditions based on the order filter
	// Skip organization filters for admin users
	if filter.IsAdmin == nil || !*filter.IsAdmin {
		if filter.BuyerOrganizationID != nil {
			dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
				Field:    "buyer_organization_id",
				Operator: base.OpEqual,
				Value:    *filter.BuyerOrganizationID,
			})
		}

		if filter.SellerOrganizationID != nil {
			dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
				Field:    "seller_organization_id",
				Operator: base.OpEqual,
				Value:    *filter.SellerOrganizationID,
			})
		}
	}

	if filter.BuyerUserID != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "buyer_user_id",
			Operator: base.OpEqual,
			Value:    *filter.BuyerUserID,
		})
	}

	if filter.Status != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "status",
			Operator: base.OpEqual,
			Value:    string(*filter.Status),
		})
	}

	if filter.MinAmount != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "total_amount",
			Operator: base.OpGreaterEqual,
			Value:    *filter.MinAmount,
		})
	}

	if filter.MaxAmount != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "total_amount",
			Operator: base.OpLessEqual,
			Value:    *filter.MaxAmount,
		})
	}

	// Date range filtering with proper error handling
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
				Field:    "order_number",
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

		// If we already have conditions, wrap them in AND logic with search
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

	// Note: Sorting will be handled by the database manager if supported
	// For now, we'll rely on the default ordering
	// TODO: Implement sorting when OrderBy is available in base.Filter
	_ = filter.SortBy // Acknowledge the sorting parameters for future use
	_ = filter.SortOrder

	// Pagination
	dbFilter.Limit = limit
	dbFilter.Offset = offset

	// Get orders
	orderList, err := r.BaseFilterableRepository.Find(ctx, dbFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find orders: %w", err)
	}

	// Load related data if requested
	if filter.IncludeItems != nil && *filter.IncludeItems {
		for i, order := range orderList {
			items, err := r.GetOrderItems(ctx, order.ID)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to load items for order %s: %w", order.ID, err)
			}
			orderList[i].Items = make([]orders.OrderItem, len(items))
			for j, item := range items {
				orderList[i].Items[j] = *item
			}
		}
	}

	if filter.IncludeHistory != nil && *filter.IncludeHistory {
		for i, order := range orderList {
			history, err := r.GetOrderHistory(ctx, order.ID)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to load history for order %s: %w", order.ID, err)
			}
			orderList[i].StatusHistory = make([]orders.OrderStatusHistory, len(history))
			for j, hist := range history {
				orderList[i].StatusHistory[j] = *hist
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

	return orderList, int(total), nil
}

// GetByBuyerID retrieves orders by buyer ID (using BaseFilterableRepository)
func (r *OrderRepository) GetByBuyerID(ctx context.Context, buyerID string, limit, offset int) ([]*orders.Order, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "buyer_organization_id",
			Operator: base.OpEqual,
			Value:    buyerID,
		},
	}
	filter.Limit = limit
	filter.Offset = offset

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetOrdersByBuyer retrieves orders for a specific buyer (alias for GetByBuyerID)
func (r *OrderRepository) GetOrdersByBuyer(ctx context.Context, buyerOrgID string, limit, offset int) ([]*orders.Order, error) {
	return r.GetByBuyerID(ctx, buyerOrgID, limit, offset)
}

// GetBySellerID retrieves orders by seller ID (using BaseFilterableRepository)
func (r *OrderRepository) GetBySellerID(ctx context.Context, sellerID string, limit, offset int) ([]*orders.Order, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "seller_organization_id",
			Operator: base.OpEqual,
			Value:    sellerID,
		},
	}
	filter.Limit = limit
	filter.Offset = offset

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetOrdersBySeller retrieves orders for a specific seller (alias for GetBySellerID)
func (r *OrderRepository) GetOrdersBySeller(ctx context.Context, sellerOrgID string, limit, offset int) ([]*orders.Order, error) {
	return r.GetBySellerID(ctx, sellerOrgID, limit, offset)
}

// GetByStatus retrieves orders by status (using BaseFilterableRepository)
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

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetOrdersByStatus retrieves orders by status (alias for GetByStatus)
func (r *OrderRepository) GetOrdersByStatus(ctx context.Context, status orders.OrderStatus, limit, offset int) ([]*orders.Order, error) {
	return r.GetByStatus(ctx, status, limit, offset)
}

// GetOrdersByDateRange retrieves orders within a date range
func (r *OrderRepository) GetOrdersByDateRange(ctx context.Context, startDate, endDate string, limit, offset int) ([]*orders.Order, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "created_at",
			Operator: base.OpGreaterEqual,
			Value:    startDate,
		},
		{
			Field:    "created_at",
			Operator: base.OpLessEqual,
			Value:    endDate,
		},
	}
	filter.Limit = limit
	filter.Offset = offset

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetOrderHistory retrieves order history
func (r *OrderRepository) GetOrderHistory(ctx context.Context, orderID string) ([]*orders.OrderStatusHistory, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "order_id",
			Operator: base.OpEqual,
			Value:    orderID,
		},
	}

	// Apply query options (adds deleted_at IS NULL by default)
	filter = r.ApplyQueryOptions(ctx, filter)

	// Note: Ordering by created_at will be handled by the database manager if supported

	var historyList []*orders.OrderStatusHistory
	if err := r.dbManager.List(ctx, filter, &historyList); err != nil {
		return nil, fmt.Errorf("failed to get order history: %w", err)
	}

	return historyList, nil
}

// GetOrderItems retrieves items for a specific order
func (r *OrderRepository) GetOrderItems(ctx context.Context, orderID string) ([]*orders.OrderItem, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "order_id",
			Operator: base.OpEqual,
			Value:    orderID,
		},
	}

	// Apply query options (adds deleted_at IS NULL by default)
	filter = r.ApplyQueryOptions(ctx, filter)

	// Note: Ordering by created_at will be handled by the database manager if supported

	var itemList []*orders.OrderItem
	if err := r.dbManager.List(ctx, filter, &itemList); err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}

	return itemList, nil
}

// ValidateOrderItems validates that all items in an order are valid and available
func (r *OrderRepository) ValidateOrderItems(ctx context.Context, items []ordersRequests.CreateOrderItemRequest, sellerOrgID string) error {
	// Validate seller organization ID
	if sellerOrgID == "" {
		return fmt.Errorf("seller organization ID is required")
	}

	for i, item := range items {
		// Basic validation
		if item.CatalogItemID == "" {
			return fmt.Errorf("item %d: catalog item ID is required", i+1)
		}
		if item.Quantity.LessThanOrEqual(decimal.Zero) {
			return fmt.Errorf("item %d: quantity must be greater than 0", i+1)
		}
		if item.UnitPrice.LessThan(decimal.Zero) {
			return fmt.Errorf("item %d: unit price cannot be negative", i+1)
		}
		if item.CatalogItemType == "" {
			return fmt.Errorf("item %d: catalog item type is required", i+1)
		}
		if item.CatalogItemType != "product" && item.CatalogItemType != "service" && item.CatalogItemType != "labour" {
			return fmt.Errorf("item %d: invalid catalog item type '%s'", i+1, item.CatalogItemType)
		}

		// Skip catalog validation if catalog repository is not set
		if r.catalogRepo == nil {
			continue
		}

		// Check if catalog item exists and is owned by seller organization
		var catalogItem any
		retrievedItem, err := r.catalogRepo.GetByID(ctx, item.CatalogItemID, &catalogItem)
		if err != nil {
			return fmt.Errorf("item %d: catalog item '%s' not found: %w", i+1, item.CatalogItemID, err)
		}

		// Verify ownership
		ownerOrgID, err := r.catalogRepo.GetCatalogItemOwner(ctx, item.CatalogItemID)
		if err != nil {
			return fmt.Errorf("item %d: failed to verify catalog item ownership: %w", i+1, err)
		}
		if ownerOrgID != sellerOrgID {
			return fmt.Errorf("item %d: catalog item '%s' does not belong to seller organization", i+1, item.CatalogItemID)
		}

		// For products, check inventory availability
		if item.CatalogItemType == "product" && r.inventoryRepo != nil {
			availableQuantity, err := r.inventoryRepo.GetInventoryLevel(ctx, item.CatalogItemID)
			if err != nil {
				return fmt.Errorf("item %d: failed to check inventory for product '%s': %w", i+1, item.CatalogItemID, err)
			}

			requestedQuantity := item.Quantity
			if availableQuantity.LessThan(requestedQuantity) {
				return fmt.Errorf("item %d: insufficient inventory for product '%s' (available: %s, requested: %s)",
					i+1, item.CatalogItemID, availableQuantity.String(), requestedQuantity.String())
			}
		}

		// Additional validation can be added here for services and labour
		_ = retrievedItem // Use the retrieved item if needed for further validation
	}

	return nil
}

// ReserveInventory reserves inventory for order items
func (r *OrderRepository) ReserveInventory(ctx context.Context, items []ordersRequests.CreateOrderItemRequest) error {
	if r.inventoryRepo == nil {
		return fmt.Errorf("inventory repository not configured")
	}

	// Track reserved items for rollback in case of failure
	var reservedItems []struct {
		itemID   string
		quantity decimal.Decimal
	}

	for i, item := range items {
		// Only reserve inventory for products
		if item.CatalogItemType == "product" {
			quantity := item.Quantity

			// Check current inventory levels before reservation
			availableQuantity, err := r.inventoryRepo.GetInventoryLevel(ctx, item.CatalogItemID)
			if err != nil {
				// Rollback previously reserved items
				r.rollbackReservations(ctx, reservedItems)
				return fmt.Errorf("item %d: failed to check inventory for product '%s': %w", i+1, item.CatalogItemID, err)
			}

			if availableQuantity.LessThan(quantity) {
				// Rollback previously reserved items
				r.rollbackReservations(ctx, reservedItems)
				return fmt.Errorf("item %d: insufficient inventory for product '%s' (available: %s, requested: %s)",
					i+1, item.CatalogItemID, availableQuantity.String(), quantity.String())
			}

			// Reserve the inventory
			if err := r.inventoryRepo.ReserveInventory(ctx, item.CatalogItemID, quantity); err != nil {
				// Rollback previously reserved items
				r.rollbackReservations(ctx, reservedItems)
				return fmt.Errorf("item %d: failed to reserve inventory for product '%s': %w", i+1, item.CatalogItemID, err)
			}

			// Track this reservation for potential rollback
			reservedItems = append(reservedItems, struct {
				itemID   string
				quantity decimal.Decimal
			}{
				itemID:   item.CatalogItemID,
				quantity: quantity,
			})
		}
	}

	return nil
}

// rollbackReservations releases previously reserved inventory
func (r *OrderRepository) rollbackReservations(ctx context.Context, reservedItems []struct {
	itemID   string
	quantity decimal.Decimal
}) {
	for _, reserved := range reservedItems {
		if err := r.inventoryRepo.ReleaseInventory(ctx, reserved.itemID, reserved.quantity); err != nil {
			// Log error but don't fail the rollback
			fmt.Printf("Warning: failed to rollback inventory reservation for item %s: %v\n", reserved.itemID, err)
		}
	}
}

// ReleaseInventory releases reserved inventory (e.g., when order is cancelled)
func (r *OrderRepository) ReleaseInventory(ctx context.Context, items []ordersRequests.CreateOrderItemRequest) error {
	if r.inventoryRepo == nil {
		return fmt.Errorf("inventory repository not configured")
	}

	var errors []error

	for i, item := range items {
		// Only release inventory for products
		if item.CatalogItemType == "product" {
			quantity := item.Quantity

			if err := r.inventoryRepo.ReleaseInventory(ctx, item.CatalogItemID, quantity); err != nil {
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

// ReleaseInventoryForOrder releases inventory for all items in an order
func (r *OrderRepository) ReleaseInventoryForOrder(ctx context.Context, orderID string) error {
	// Get order items
	items, err := r.GetOrderItems(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order items: %w", err)
	}

	if r.inventoryRepo == nil {
		return fmt.Errorf("inventory repository not configured")
	}

	var errors []error

	for i, item := range items {
		// Only release inventory for products
		if item.CatalogItemType == "product" {
			if err := r.inventoryRepo.ReleaseInventory(ctx, item.CatalogItemID, item.Quantity); err != nil {
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

// GetOrderSummary retrieves a summary of an order
func (r *OrderRepository) GetOrderSummary(ctx context.Context, id string) (*orders.OrderSummary, error) {
	order, err := r.GetOrderByID(ctx, id)
	if err != nil {
		return nil, err
	}

	summary := &orders.OrderSummary{
		ID:                order.ID,
		OrderNumber:       order.OrderNumber,
		BuyerOrgID:        order.BuyerOrganizationID,
		SellerOrgID:       order.SellerOrganizationID,
		Status:            order.Status,
		TotalAmount:       order.TotalAmount,
		ItemCount:         len(order.Items),
		EstimatedDelivery: order.EstimatedDeliveryDate,
		CreatedAt:         order.CreatedAt,
	}

	return summary, nil
}
