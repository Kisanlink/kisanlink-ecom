package catalog

import (
	"context"
	"fmt"

	"kisanlink-ecom/entities/models/catalog"
	catalogRequests "kisanlink-ecom/entities/requests/catalog"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// CatalogRepository handles catalog operations using the database manager
type CatalogRepository struct {
	dbManager db.DBManager
}

// NewCatalogRepository creates a new catalog repository
func NewCatalogRepository(dbManager db.DBManager) *CatalogRepository {
	return &CatalogRepository{
		dbManager: dbManager,
	}
}

// GetByID retrieves a catalog item by ID from the database
func (r *CatalogRepository) GetByID(ctx context.Context, id string, model interface{}) (interface{}, error) {
	if err := r.dbManager.GetByID(ctx, id, model); err != nil {
		return nil, fmt.Errorf("failed to get catalog item by ID: %w", err)
	}
	return model, nil
}

// Create creates a new catalog item in the database
func (r *CatalogRepository) Create(ctx context.Context, model interface{}) error {
	return r.dbManager.Create(ctx, model)
}

// Update updates an existing catalog item in the database
func (r *CatalogRepository) Update(ctx context.Context, model interface{}) error {
	return r.dbManager.Update(ctx, model)
}

// Delete deletes a catalog item from the database
func (r *CatalogRepository) Delete(ctx context.Context, id string, model interface{}) error {
	return r.dbManager.Delete(ctx, id, model)
}

// SoftDelete soft deletes a catalog item in the database
func (r *CatalogRepository) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	// Get the model first to check if it exists
	var item catalog.CatalogItem
	if err := r.dbManager.GetByID(ctx, id, &item); err != nil {
		return fmt.Errorf("failed to get catalog item for soft delete: %w", err)
	}

	// Update the model with soft delete fields
	item.SetDeletedBy(&deletedBy)

	// Save the updated model
	return r.dbManager.Update(ctx, &item)
}

// Restore restores a soft-deleted catalog item
func (r *CatalogRepository) Restore(ctx context.Context, id string) error {
	var item catalog.CatalogItem
	if err := r.dbManager.GetByID(ctx, id, &item); err != nil {
		return fmt.Errorf("failed to get catalog item for restore: %w", err)
	}

	// Clear soft delete fields
	item.SetDeletedBy(nil)

	// Save the updated model
	return r.dbManager.Update(ctx, &item)
}

// Find retrieves catalog items using filters
func (r *CatalogRepository) Find(ctx context.Context, filter *base.Filter) ([]*catalog.CatalogItem, error) {
	var items []*catalog.CatalogItem

	// Use the database manager's List method with the filter
	if err := r.dbManager.List(ctx, filter, &items); err != nil {
		return nil, fmt.Errorf("failed to find catalog items: %w", err)
	}

	return items, nil
}

// GetBySKU retrieves a catalog item by SKU
func (r *CatalogRepository) GetBySKU(ctx context.Context, sku string) (*catalog.CatalogItem, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "sku",
			Operator: base.OpEqual,
			Value:    sku,
		},
	}

	items, err := r.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, nil
	}

	return items[0], nil
}

// GetByOrgID retrieves catalog items by organization ID
func (r *CatalogRepository) GetByOrgID(ctx context.Context, orgID string, limit, offset int) ([]*catalog.CatalogItem, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "organization_id",
			Operator: base.OpEqual,
			Value:    orgID,
		},
		{
			Field:    "is_active",
			Operator: base.OpEqual,
			Value:    true,
		},
	}
	filter.Limit = limit
	filter.Offset = offset

	return r.Find(ctx, filter)
}

// GetByCategory retrieves catalog items by category
func (r *CatalogRepository) GetByCategory(ctx context.Context, category string, limit, offset int) ([]*catalog.CatalogItem, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "category",
			Operator: base.OpEqual,
			Value:    category,
		},
		{
			Field:    "is_active",
			Operator: base.OpEqual,
			Value:    true,
		},
	}
	filter.Limit = limit
	filter.Offset = offset

	return r.Find(ctx, filter)
}

// ListCatalogItems retrieves catalog items with filtering and pagination
func (r *CatalogRepository) ListCatalogItems(ctx context.Context, filter *catalogRequests.CatalogFilter, offset, limit int) ([]*catalog.CatalogItem, int, error) {
	dbFilter := base.NewFilter()

	// Org scoping
	if filter.OrganizationID != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "organization_id",
			Operator: base.OpEqual,
			Value:    *filter.OrganizationID,
		})
	}

	// Type
	if filter.ItemType != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "item_type",
			Operator: base.OpEqual,
			Value:    *filter.ItemType,
		})
	}

	// Category
	if filter.Category != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "category",
			Operator: base.OpEqual,
			Value:    *filter.Category,
		})
	}
	if filter.Subcategory != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "subcategory",
			Operator: base.OpEqual,
			Value:    *filter.Subcategory,
		})
	}

	// Active
	if filter.IsActive != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "is_active",
			Operator: base.OpEqual,
			Value:    *filter.IsActive,
		})
	}

	// Visibility
	if filter.Visibility != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "visibility",
			Operator: base.OpEqual,
			Value:    *filter.Visibility,
		})
	}

	// Price range
	if filter.MinPrice != nil && filter.MaxPrice != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "base_price",
			Operator: base.OpDateBetween,
			Value:    *filter.MinPrice,
			Value2:   *filter.MaxPrice,
		})
	} else if filter.MinPrice != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "base_price",
			Operator: base.OpGreaterEqual,
			Value:    *filter.MinPrice,
		})
	} else if filter.MaxPrice != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "base_price",
			Operator: base.OpLessEqual,
			Value:    *filter.MaxPrice,
		})
	}

	// Tags (contains any)
	if len(filter.Tags) > 0 {
		// Using LIKE on tags JSON/text[] is implementation-specific; for now use contains on tags field string representation
		for _, tag := range filter.Tags {
			dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
				Field:    "tags",
				Operator: base.OpContains,
				Value:    tag,
			})
		}
	}

	// Search across name and description
	if filter.Search != nil && *filter.Search != "" {
		search := *filter.Search
		dbFilter.Group.Groups = append(dbFilter.Group.Groups, base.FilterGroup{
			Logic: base.LogicOr,
			Conditions: []base.FilterCondition{
				{Field: "name", Operator: base.OpContains, Value: search},
				{Field: "description", Operator: base.OpContains, Value: search},
			},
		})
	}

	dbFilter.Limit = limit
	dbFilter.Offset = offset

	items, err := r.Find(ctx, dbFilter)
	if err != nil {
		return nil, 0, err
	}

	// Get total count for pagination
	total, err := r.dbManager.Count(ctx, dbFilter, &catalog.CatalogItem{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return items, int(total), nil
}

// Product operations
func (r *CatalogRepository) CreateProduct(ctx context.Context, product *catalog.Product) error {
	return r.Create(ctx, &product.CatalogItem)
}

func (r *CatalogRepository) GetProductByID(ctx context.Context, id string) (*catalog.Product, error) {
	var item catalog.CatalogItem
	retrievedItem, err := r.GetByID(ctx, id, &item)
	if err != nil {
		return nil, err
	}

	// Convert to Product type
	product := &catalog.Product{
		CatalogItem: *retrievedItem.(*catalog.CatalogItem),
	}

	// Load product-specific fields from database
	// TODO: Implement actual database query for product fields
	return product, nil
}

func (r *CatalogRepository) GetProductBySKU(ctx context.Context, sku string) (*catalog.Product, error) {
	item, err := r.GetBySKU(ctx, sku)
	if err != nil {
		return nil, err
	}
	if item == nil || item.ItemType != catalog.CatalogItemTypeProduct {
		return nil, fmt.Errorf("product not found")
	}

	product := &catalog.Product{
		CatalogItem: *item,
	}

	// TODO: Load product-specific fields
	return product, nil
}

func (r *CatalogRepository) UpdateProduct(ctx context.Context, id string, updates *catalogRequests.UpdateCatalogItemRequest) error {
	// TODO: Implement product update with specific fields
	return fmt.Errorf("not implemented")
}

func (r *CatalogRepository) DeleteProduct(ctx context.Context, id string) error {
	var item catalog.CatalogItem
	return r.Delete(ctx, id, &item)
}

func (r *CatalogRepository) ListProducts(ctx context.Context, filter *catalogRequests.CatalogFilter, offset, limit int) ([]*catalog.Product, int, error) {
	// TODO: Implement product listing with filters
	return nil, 0, fmt.Errorf("not implemented")
}

// Service operations
func (r *CatalogRepository) CreateService(ctx context.Context, service *catalog.Service) error {
	return r.Create(ctx, &service.CatalogItem)
}

func (r *CatalogRepository) GetServiceByID(ctx context.Context, id string) (*catalog.Service, error) {
	var item catalog.CatalogItem
	retrievedItem, err := r.GetByID(ctx, id, &item)
	if err != nil {
		return nil, err
	}
	if retrievedItem.(*catalog.CatalogItem).ItemType != catalog.CatalogItemTypeService {
		return nil, fmt.Errorf("service not found")
	}

	service := &catalog.Service{
		CatalogItem: *retrievedItem.(*catalog.CatalogItem),
	}

	// TODO: Load service-specific fields
	return service, nil
}

func (r *CatalogRepository) GetServiceBySKU(ctx context.Context, sku string) (*catalog.Service, error) {
	item, err := r.GetBySKU(ctx, sku)
	if err != nil {
		return nil, err
	}
	if item == nil || item.ItemType != catalog.CatalogItemTypeService {
		return nil, fmt.Errorf("service not found")
	}

	service := &catalog.Service{
		CatalogItem: *item,
	}

	// TODO: Load service-specific fields
	return service, nil
}

func (r *CatalogRepository) UpdateService(ctx context.Context, id string, updates *catalogRequests.UpdateCatalogItemRequest) error {
	// TODO: Implement service update with specific fields
	return fmt.Errorf("not implemented")
}

func (r *CatalogRepository) DeleteService(ctx context.Context, id string) error {
	var item catalog.Service
	return r.Delete(ctx, id, &item)
}

func (r *CatalogRepository) ListServices(ctx context.Context, filter *catalogRequests.CatalogFilter, offset, limit int) ([]*catalog.Service, int, error) {
	// TODO: Implement service listing with filters
	return nil, 0, fmt.Errorf("not implemented")
}

// Labour operations
func (r *CatalogRepository) CreateLabour(ctx context.Context, labour *catalog.Labour) error {
	return r.Create(ctx, &labour.CatalogItem)
}

func (r *CatalogRepository) GetLabourByID(ctx context.Context, id string) (*catalog.Labour, error) {
	var item catalog.CatalogItem
	retrievedItem, err := r.GetByID(ctx, id, &item)
	if err != nil {
		return nil, err
	}
	if retrievedItem.(*catalog.CatalogItem).ItemType != catalog.CatalogItemTypeLabour {
		return nil, fmt.Errorf("labour not found")
	}

	labour := &catalog.Labour{
		CatalogItem: *retrievedItem.(*catalog.CatalogItem),
	}

	// TODO: Load labour-specific fields
	return labour, nil
}

func (r *CatalogRepository) GetLabourBySKU(ctx context.Context, sku string) (*catalog.Labour, error) {
	item, err := r.GetBySKU(ctx, sku)
	if err != nil {
		return nil, err
	}
	if item == nil || item.ItemType != catalog.CatalogItemTypeLabour {
		return nil, fmt.Errorf("labour not found")
	}

	service := &catalog.Labour{
		CatalogItem: *item,
	}

	// TODO: Load labour-specific fields
	return service, nil
}

func (r *CatalogRepository) UpdateLabour(ctx context.Context, id string, updates *catalogRequests.UpdateCatalogItemRequest) error {
	// TODO: Implement labour update with specific fields
	return fmt.Errorf("not implemented")
}

func (r *CatalogRepository) DeleteLabour(ctx context.Context, id string) error {
	var item catalog.Labour
	return r.Delete(ctx, id, &item)
}

func (r *CatalogRepository) ListLabour(ctx context.Context, filter *catalogRequests.CatalogFilter, offset, limit int) ([]*catalog.Labour, int, error) {
	// TODO: Implement labour listing with filters
	return nil, 0, fmt.Errorf("not implemented")
}

func (r *CatalogRepository) GetCatalogItemOwner(ctx context.Context, id string) (string, error) {
	var item catalog.CatalogItem
	retrievedItem, err := r.GetByID(ctx, id, &item)
	if err != nil {
		return "", err
	}
	return retrievedItem.(*catalog.CatalogItem).OrganizationID, nil
}

// Inventory operations
func (r *CatalogRepository) ReserveInventory(ctx context.Context, itemID string, quantity float64) error {
	// TODO: Implement inventory reservation
	return fmt.Errorf("not implemented")
}

func (r *CatalogRepository) ReleaseInventory(ctx context.Context, itemID string, quantity float64) error {
	// TODO: Implement inventory release
	return fmt.Errorf("not implemented")
}

func (r *CatalogRepository) GetInventoryLevel(ctx context.Context, itemID string) (float64, error) {
	// TODO: Implement inventory level check
	return 0, fmt.Errorf("not implemented")
}
