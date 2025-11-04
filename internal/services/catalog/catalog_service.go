package catalog

import (
	"context"
	"fmt"

	catalogModels "kisanlink-ecom/entities/models/catalog"
	catalogRequests "kisanlink-ecom/entities/requests/catalog"

	"github.com/shopspring/decimal"
)

// CatalogServiceInterface defines the interface for catalog operations
type CatalogServiceInterface interface {
	CreateProduct(ctx context.Context, product *catalogModels.Product, userID string) (*catalogModels.Product, error)
	GetProductByID(ctx context.Context, productID string) (*catalogModels.Product, error)
	GetProductBySKU(ctx context.Context, sku string) (*catalogModels.Product, error)
	UpdateProduct(ctx context.Context, product *catalogModels.Product, userID string) (*catalogModels.Product, error)
	DeleteProduct(ctx context.Context, productID string, userID string) error
	ListProducts(ctx context.Context, limit, offset int, category, status string) ([]*catalogModels.CatalogItem, error)
	CreateService(ctx context.Context, req *catalogRequests.CreateServiceRequest, userID string) (*catalogModels.Service, error)
	GetServiceByID(ctx context.Context, serviceID string) (*catalogModels.Service, error)
	GetServiceBySKU(ctx context.Context, sku string) (*catalogModels.Service, error)
	UpdateService(ctx context.Context, service *catalogModels.Service, userID string) (*catalogModels.Service, error)
	DeleteService(ctx context.Context, serviceID string, userID string) error
	ListServices(ctx context.Context, limit, offset int, category, status string) ([]*catalogModels.CatalogItem, error)
	CreateLabour(ctx context.Context, req *catalogRequests.CreateLabourRequest, userID string) (*catalogModels.Labour, error)
	GetLabourByID(ctx context.Context, labourID string) (*catalogModels.Labour, error)
	GetLabourBySKU(ctx context.Context, sku string) (*catalogModels.Labour, error)
	UpdateLabour(ctx context.Context, labour *catalogModels.Labour, userID string) (*catalogModels.Labour, error)
	DeleteLabour(ctx context.Context, labourID string, userID string) error
	ListLabour(ctx context.Context, limit, offset int, category, status string) ([]*catalogModels.CatalogItem, error)
	GetCatalogItemByID(ctx context.Context, id string) (*catalogModels.CatalogItem, error)
	UpdateCatalogItem(ctx context.Context, catalogItem *catalogModels.CatalogItem, userID string) (*catalogModels.CatalogItem, error)
	UpdateActiveStatus(ctx context.Context, id string, isActive bool, userID string) error
	DeleteContract(ctx context.Context, contractID string, userID string) error
	ListCatalogItems(ctx context.Context, filter *catalogRequests.CatalogFilter, offset, limit int) ([]*catalogModels.CatalogItem, int, error)
	SearchCatalog(ctx context.Context, query string, filter *catalogRequests.CatalogFilter, offset, limit int) ([]*catalogModels.CatalogItem, int, error)
	GetInventoryLevel(ctx context.Context, itemID string) (float64, error)
	UpdateInventory(ctx context.Context, itemID string, quantity float64, operation string) error
}

// CatalogRepositoryInterface defines the interface for catalog repository operations
type CatalogRepositoryInterface interface {
	GetByID(ctx context.Context, id string, model interface{}) (interface{}, error)
	Create(ctx context.Context, model interface{}) error
	Update(ctx context.Context, model interface{}) error
	Delete(ctx context.Context, id string, model interface{}) error
	SoftDelete(ctx context.Context, id string, deletedBy string) error
	Restore(ctx context.Context, id string) error
	GetBySKU(ctx context.Context, sku string) (*catalogModels.CatalogItem, error)
	ListCatalogItems(ctx context.Context, filter *catalogRequests.CatalogFilter, offset, limit int) ([]*catalogModels.CatalogItem, int, error)
	GetInventoryLevel(ctx context.Context, itemID string) (float64, error)
	ReserveInventory(ctx context.Context, itemID string, quantity float64) error
	ReleaseInventory(ctx context.Context, itemID string, quantity float64) error
}

// CatalogService provides business logic for catalog operations
type CatalogService struct {
	catalogRepo CatalogRepositoryInterface
}

// NewCatalogService creates a new catalog service
func NewCatalogService(catalogRepo CatalogRepositoryInterface) *CatalogService {
	return &CatalogService{
		catalogRepo: catalogRepo,
	}
}

// CreateProduct creates a new product
func (s *CatalogService) CreateProduct(ctx context.Context, product *catalogModels.Product, userID string) (*catalogModels.Product, error) {
	// Validate product data
	if err := s.validateProduct(product); err != nil {
		return nil, fmt.Errorf("product validation failed: %w", err)
	}

	// Check if SKU already exists
	existing, err := s.catalogRepo.GetBySKU(ctx, product.SKU)
	if err != nil {
		return nil, fmt.Errorf("failed to check SKU uniqueness: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("SKU %s already exists", product.SKU)
	}

	// Create the product - since Product embeds CatalogItem, we can access the CatalogItem fields directly
	catalogItem := &catalogModels.CatalogItem{
		BaseModel:      product.BaseModel,
		ItemType:       catalogModels.CatalogItemTypeProduct,
		OrganizationID: product.OrganizationID,
		SKU:            product.SKU,
		Name:           product.Name,
		Description:    product.Description,
		Category:       product.Category,
		Subcategory:    product.Subcategory,
		UnitOfMeasure:  product.UnitOfMeasure,
		BasePrice:      product.BasePrice,
		Currency:       product.Currency,
		IsActive:       product.IsActive,
		Visibility:     product.Visibility,
		Tags:           product.Tags,
		Attributes:     product.Attributes,
		Images:         product.Images,
	}

	if err := s.catalogRepo.Create(ctx, catalogItem); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	// Update the product with the created catalog item data
	product.BaseModel = catalogItem.BaseModel
	return product, nil
}

// UpdateProduct updates an existing product
func (s *CatalogService) UpdateProduct(ctx context.Context, product *catalogModels.Product, userID string) (*catalogModels.Product, error) {
	// Validate product data
	if err := s.validateProduct(product); err != nil {
		return nil, fmt.Errorf("product validation failed: %w", err)
	}

	// Check if SKU conflicts with other products
	existing, err := s.catalogRepo.GetBySKU(ctx, product.SKU)
	if err != nil {
		return nil, fmt.Errorf("failed to check SKU uniqueness: %w", err)
	}
	if existing != nil && existing.ID != product.ID {
		return nil, fmt.Errorf("SKU %s already exists on another product", product.SKU)
	}

	// Update the product - create a CatalogItem from the Product
	catalogItem := &catalogModels.CatalogItem{
		BaseModel:      product.BaseModel,
		ItemType:       catalogModels.CatalogItemTypeProduct,
		OrganizationID: product.OrganizationID,
		SKU:            product.SKU,
		Name:           product.Name,
		Description:    product.Description,
		Category:       product.Category,
		Subcategory:    product.Subcategory,
		UnitOfMeasure:  product.UnitOfMeasure,
		BasePrice:      product.BasePrice,
		Currency:       product.Currency,
		IsActive:       product.IsActive,
		Visibility:     product.Visibility,
		Tags:           product.Tags,
		Attributes:     product.Attributes,
		Images:         product.Images,
	}

	if err := s.catalogRepo.Update(ctx, catalogItem); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	// Update the product with the updated catalog item data
	product.BaseModel = catalogItem.BaseModel
	return product, nil
}

// GetProductByID retrieves a product by ID
func (s *CatalogService) GetProductByID(ctx context.Context, productID string) (*catalogModels.Product, error) {
	catalogItem := &catalogModels.CatalogItem{}
	retrievedItem, err := s.catalogRepo.GetByID(ctx, productID, catalogItem)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Type assertion
	catalogItem = retrievedItem.(*catalogModels.CatalogItem)

	if !catalogItem.IsActive {
		return nil, fmt.Errorf("product %s is not active", productID)
	}

	// Convert CatalogItem to Product
	product := &catalogModels.Product{
		CatalogItem: *catalogItem,
	}

	return product, nil
}

// GetProductBySKU retrieves a product by SKU
func (s *CatalogService) GetProductBySKU(ctx context.Context, sku string) (*catalogModels.Product, error) {
	catalogItem, err := s.catalogRepo.GetBySKU(ctx, sku)
	if err != nil {
		return nil, fmt.Errorf("failed to get product by SKU: %w", err)
	}

	if catalogItem == nil {
		return nil, fmt.Errorf("product with SKU %s not found", sku)
	}

	if !catalogItem.IsActive {
		return nil, fmt.Errorf("product with SKU %s is not active", sku)
	}

	// Convert CatalogItem to Product
	product := &catalogModels.Product{
		CatalogItem: *catalogItem,
	}

	return product, nil
}

// ListProducts retrieves a list of products with optional filtering
func (s *CatalogService) ListProducts(ctx context.Context, limit, offset int, category, status string) ([]*catalogModels.CatalogItem, error) {
	// Create filter for products only
	itemType := catalogModels.CatalogItemTypeProduct
	filter := &catalogRequests.CatalogFilter{
		ItemType: &itemType,
	}

	// Add category filter if specified
	if category != "" {
		filter.Category = &category
	}

	// Add status filter if specified
	if status != "" {
		if status == "active" {
			active := true
			filter.IsActive = &active
		} else if status == "inactive" {
			active := false
			filter.IsActive = &active
		}
	}

	// Use the repository's ListCatalogItems method which queries the database
	products, _, err := s.catalogRepo.ListCatalogItems(ctx, filter, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	return products, nil
}

// DeleteProduct soft deletes a product
func (s *CatalogService) DeleteProduct(ctx context.Context, productID, userID string) error {
	// Check if product exists
	catalogItem := &catalogModels.CatalogItem{}
	retrievedItem, err := s.catalogRepo.GetByID(ctx, productID, catalogItem)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	// Type assertion
	catalogItem = retrievedItem.(*catalogModels.CatalogItem)

	// Soft delete the product
	if err := s.catalogRepo.SoftDelete(ctx, productID, userID); err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

// RestoreProduct restores a soft-deleted product
func (s *CatalogService) RestoreProduct(ctx context.Context, productID, userID string) error {
	// Restore the product
	if err := s.catalogRepo.Restore(ctx, productID); err != nil {
		return fmt.Errorf("failed to restore product: %w", err)
	}

	return nil
}

// UpdateProductStock updates the stock quantity of a product
func (s *CatalogService) UpdateProductStock(ctx context.Context, productID string, newStock int, userID string) error {
	// Get the product
	catalogItem := &catalogModels.CatalogItem{}
	retrievedItem, err := s.catalogRepo.GetByID(ctx, productID, catalogItem)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	// Type assertion
	catalogItem = retrievedItem.(*catalogModels.CatalogItem)

	// Update the product - note: CatalogItem doesn't have stock field, this would need to be handled differently
	// For now, we'll just update the product
	if err := s.catalogRepo.Update(ctx, catalogItem); err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

// validateProduct validates product data
func (s *CatalogService) validateProduct(product *catalogModels.Product) error {
	if product == nil {
		return fmt.Errorf("product cannot be nil")
	}

	if product.Name == "" {
		return fmt.Errorf("product name is required")
	}

	if product.SKU == "" {
		return fmt.Errorf("product SKU is required")
	}

	if product.BasePrice.LessThan(decimal.Zero) {
		return fmt.Errorf("product price cannot be negative")
	}

	if product.Category == "" {
		return fmt.Errorf("product category is required")
	}

	if product.UnitOfMeasure == "" {
		return fmt.Errorf("product unit of measure is required")
	}

	if product.Currency == "" {
		return fmt.Errorf("product currency is required")
	}

	// if product.OrganizationID == "" {
	//  return fmt.Errorf("product organization ID is required")
	// }

	return nil
}

// Service operations
func (s *CatalogService) CreateService(ctx context.Context, req *catalogRequests.CreateServiceRequest, userID string) (*catalogModels.Service, error) {
	// Validate service request
	if err := s.validateServiceRequest(req); err != nil {
		return nil, fmt.Errorf("service validation failed: %w", err)
	}

	// Check if SKU already exists
	if req.SKU != "" {
		existing, err := s.catalogRepo.GetBySKU(ctx, req.SKU)
		if err != nil {
			return nil, fmt.Errorf("failed to check SKU uniqueness: %w", err)
		}
		if existing != nil {
			return nil, fmt.Errorf("SKU %s already exists", req.SKU)
		}
	}

	// Create service from request
	service := s.createServiceFromRequest(req, userID)

	// Create the service catalog item
	catalogItem := &catalogModels.CatalogItem{
		BaseModel:      service.BaseModel,
		ItemType:       catalogModels.CatalogItemTypeService,
		OrganizationID: service.OrganizationID,
		SKU:            service.SKU,
		Name:           service.Name,
		Description:    service.Description,
		Category:       service.Category,
		Subcategory:    service.Subcategory,
		UnitOfMeasure:  service.UnitOfMeasure,
		BasePrice:      service.BasePrice,
		Currency:       service.Currency,
		IsActive:       service.IsActive,
		Visibility:     service.Visibility,
		Tags:           service.Tags,
		Attributes:     service.Attributes,
		Images:         service.Images,
	}

	if err := s.catalogRepo.Create(ctx, catalogItem); err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	// Update the service with the created catalog item data
	service.BaseModel = catalogItem.BaseModel
	return service, nil
}

func (s *CatalogService) GetServiceByID(ctx context.Context, id string) (*catalogModels.Service, error) {
	catalogItem := &catalogModels.CatalogItem{}
	retrievedItem, err := s.catalogRepo.GetByID(ctx, id, catalogItem)
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	catalogItem = retrievedItem.(*catalogModels.CatalogItem)

	if catalogItem.ItemType != catalogModels.CatalogItemTypeService {
		return nil, fmt.Errorf("item %s is not a service", id)
	}

	if !catalogItem.IsActive {
		return nil, fmt.Errorf("service %s is not active", id)
	}

	// Convert CatalogItem to Service
	service := &catalogModels.Service{
		CatalogItem: *catalogItem,
	}

	return service, nil
}

func (s *CatalogService) GetServiceBySKU(ctx context.Context, sku string) (*catalogModels.Service, error) {
	catalogItem, err := s.catalogRepo.GetBySKU(ctx, sku)
	if err != nil {
		return nil, fmt.Errorf("failed to get service by SKU: %w", err)
	}

	if catalogItem == nil {
		return nil, fmt.Errorf("service with SKU %s not found", sku)
	}

	if catalogItem.ItemType != catalogModels.CatalogItemTypeService {
		return nil, fmt.Errorf("item with SKU %s is not a service", sku)
	}

	if !catalogItem.IsActive {
		return nil, fmt.Errorf("service with SKU %s is not active", sku)
	}

	// Convert CatalogItem to Service
	service := &catalogModels.Service{
		CatalogItem: *catalogItem,
	}

	return service, nil
}

func (s *CatalogService) UpdateService(ctx context.Context, service *catalogModels.Service, userID string) (*catalogModels.Service, error) {
	// Validate service data
	if err := s.validateService(service); err != nil {
		return nil, fmt.Errorf("service validation failed: %w", err)
	}

	// Check if SKU conflicts with other services
	if service.SKU != "" {
		existing, err := s.catalogRepo.GetBySKU(ctx, service.SKU)
		if err != nil {
			return nil, fmt.Errorf("failed to check SKU uniqueness: %w", err)
		}
		if existing != nil && existing.ID != service.ID {
			return nil, fmt.Errorf("SKU %s already exists on another item", service.SKU)
		}
	}

	// Update the service - create a CatalogItem from the Service
	catalogItem := &catalogModels.CatalogItem{
		BaseModel:      service.BaseModel,
		ItemType:       catalogModels.CatalogItemTypeService,
		OrganizationID: service.OrganizationID,
		SKU:            service.SKU,
		Name:           service.Name,
		Description:    service.Description,
		Category:       service.Category,
		Subcategory:    service.Subcategory,
		UnitOfMeasure:  service.UnitOfMeasure,
		BasePrice:      service.BasePrice,
		Currency:       service.Currency,
		IsActive:       service.IsActive,
		Visibility:     service.Visibility,
		Tags:           service.Tags,
		Attributes:     service.Attributes,
		Images:         service.Images,
	}

	if err := s.catalogRepo.Update(ctx, catalogItem); err != nil {
		return nil, fmt.Errorf("failed to update service: %w", err)
	}

	// Update the service with the updated catalog item data
	service.BaseModel = catalogItem.BaseModel
	return service, nil
}

func (s *CatalogService) DeleteService(ctx context.Context, serviceID, userID string) error {
	// Check if service exists
	catalogItem := &catalogModels.CatalogItem{}
	retrievedItem, err := s.catalogRepo.GetByID(ctx, serviceID, catalogItem)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	catalogItem = retrievedItem.(*catalogModels.CatalogItem)

	if catalogItem.ItemType != catalogModels.CatalogItemTypeService {
		return fmt.Errorf("item %s is not a service", serviceID)
	}

	// Soft delete the service
	if err := s.catalogRepo.SoftDelete(ctx, serviceID, userID); err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	return nil
}

func (s *CatalogService) ListServices(ctx context.Context, limit, offset int, category, status string) ([]*catalogModels.CatalogItem, error) {
	// Create filter for services only
	itemType := catalogModels.CatalogItemTypeService
	filter := &catalogRequests.CatalogFilter{
		ItemType: &itemType,
	}

	// Add category filter if specified
	if category != "" {
		filter.Category = &category
	}

	// Add status filter if specified
	if status != "" {
		if status == "active" {
			active := true
			filter.IsActive = &active
		} else if status == "inactive" {
			active := false
			filter.IsActive = &active
		}
	}

	// Use the repository's ListCatalogItems method
	services, _, err := s.catalogRepo.ListCatalogItems(ctx, filter, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	return services, nil
}

// Labour operations
func (s *CatalogService) CreateLabour(ctx context.Context, req *catalogRequests.CreateLabourRequest, userID string) (*catalogModels.Labour, error) {
	// Validate labour request
	if err := s.validateLabourRequest(req); err != nil {
		return nil, fmt.Errorf("labour validation failed: %w", err)
	}

	// Check if SKU already exists
	if req.SKU != "" {
		existing, err := s.catalogRepo.GetBySKU(ctx, req.SKU)
		if err != nil {
			return nil, fmt.Errorf("failed to check SKU uniqueness: %w", err)
		}
		if existing != nil {
			return nil, fmt.Errorf("SKU %s already exists", req.SKU)
		}
	}

	// Create labour from request
	labour := s.createLabourFromRequest(req, userID)

	// Create the labour catalog item
	catalogItem := &catalogModels.CatalogItem{
		BaseModel:      labour.BaseModel,
		ItemType:       catalogModels.CatalogItemTypeLabour,
		OrganizationID: labour.OrganizationID,
		SKU:            labour.SKU,
		Name:           labour.Name,
		Description:    labour.Description,
		Category:       labour.Category,
		Subcategory:    labour.Subcategory,
		UnitOfMeasure:  labour.UnitOfMeasure,
		BasePrice:      labour.BasePrice,
		Currency:       labour.Currency,
		IsActive:       labour.IsActive,
		Visibility:     labour.Visibility,
		Tags:           labour.Tags,
		Attributes:     labour.Attributes,
		Images:         labour.Images,
	}

	if err := s.catalogRepo.Create(ctx, catalogItem); err != nil {
		return nil, fmt.Errorf("failed to create labour: %w", err)
	}

	// Update the labour with the created catalog item data
	labour.BaseModel = catalogItem.BaseModel
	return labour, nil
}

func (s *CatalogService) GetLabourByID(ctx context.Context, id string) (*catalogModels.Labour, error) {
	catalogItem := &catalogModels.CatalogItem{}
	retrievedItem, err := s.catalogRepo.GetByID(ctx, id, catalogItem)
	if err != nil {
		return nil, fmt.Errorf("failed to get labour: %w", err)
	}

	catalogItem = retrievedItem.(*catalogModels.CatalogItem)

	if catalogItem.ItemType != catalogModels.CatalogItemTypeLabour {
		return nil, fmt.Errorf("item %s is not labour", id)
	}

	if !catalogItem.IsActive {
		return nil, fmt.Errorf("labour %s is not active", id)
	}

	// Convert CatalogItem to Labour
	labour := &catalogModels.Labour{
		CatalogItem: *catalogItem,
	}

	return labour, nil
}

func (s *CatalogService) GetLabourBySKU(ctx context.Context, sku string) (*catalogModels.Labour, error) {
	catalogItem, err := s.catalogRepo.GetBySKU(ctx, sku)
	if err != nil {
		return nil, fmt.Errorf("failed to get labour by SKU: %w", err)
	}

	if catalogItem == nil {
		return nil, fmt.Errorf("labour with SKU %s not found", sku)
	}

	if catalogItem.ItemType != catalogModels.CatalogItemTypeLabour {
		return nil, fmt.Errorf("item with SKU %s is not labour", sku)
	}

	if !catalogItem.IsActive {
		return nil, fmt.Errorf("labour with SKU %s is not active", sku)
	}

	// Convert CatalogItem to Labour
	labour := &catalogModels.Labour{
		CatalogItem: *catalogItem,
	}

	return labour, nil
}

func (s *CatalogService) UpdateLabour(ctx context.Context, labour *catalogModels.Labour, userID string) (*catalogModels.Labour, error) {
	// Validate labour data
	if err := s.validateLabour(labour); err != nil {
		return nil, fmt.Errorf("labour validation failed: %w", err)
	}

	// Check if SKU conflicts with other items
	if labour.SKU != "" {
		existing, err := s.catalogRepo.GetBySKU(ctx, labour.SKU)
		if err != nil {
			return nil, fmt.Errorf("failed to check SKU uniqueness: %w", err)
		}
		if existing != nil && existing.ID != labour.ID {
			return nil, fmt.Errorf("SKU %s already exists on another item", labour.SKU)
		}
	}

	// Update the labour - create a CatalogItem from the Labour
	catalogItem := &catalogModels.CatalogItem{
		BaseModel:      labour.BaseModel,
		ItemType:       catalogModels.CatalogItemTypeLabour,
		OrganizationID: labour.OrganizationID,
		SKU:            labour.SKU,
		Name:           labour.Name,
		Description:    labour.Description,
		Category:       labour.Category,
		Subcategory:    labour.Subcategory,
		UnitOfMeasure:  labour.UnitOfMeasure,
		BasePrice:      labour.BasePrice,
		Currency:       labour.Currency,
		IsActive:       labour.IsActive,
		Visibility:     labour.Visibility,
		Tags:           labour.Tags,
		Attributes:     labour.Attributes,
		Images:         labour.Images,
	}

	if err := s.catalogRepo.Update(ctx, catalogItem); err != nil {
		return nil, fmt.Errorf("failed to update labour: %w", err)
	}

	// Update the labour with the updated catalog item data
	labour.BaseModel = catalogItem.BaseModel
	return labour, nil
}

func (s *CatalogService) DeleteLabour(ctx context.Context, labourID, userID string) error {
	// Check if labour exists
	catalogItem := &catalogModels.CatalogItem{}
	retrievedItem, err := s.catalogRepo.GetByID(ctx, labourID, catalogItem)
	if err != nil {
		return fmt.Errorf("failed to get labour: %w", err)
	}

	catalogItem = retrievedItem.(*catalogModels.CatalogItem)

	if catalogItem.ItemType != catalogModels.CatalogItemTypeLabour {
		return fmt.Errorf("item %s is not labour", labourID)
	}

	// Soft delete the labour
	if err := s.catalogRepo.SoftDelete(ctx, labourID, userID); err != nil {
		return fmt.Errorf("failed to delete labour: %w", err)
	}

	return nil
}

func (s *CatalogService) ListLabour(ctx context.Context, limit, offset int, category, status string) ([]*catalogModels.CatalogItem, error) {
	// Create filter for labour only
	itemType := catalogModels.CatalogItemTypeLabour
	filter := &catalogRequests.CatalogFilter{
		ItemType: &itemType,
	}

	// Add category filter if specified
	if category != "" {
		filter.Category = &category
	}

	// Add status filter if specified
	if status != "" {
		if status == "active" {
			active := true
			filter.IsActive = &active
		} else if status == "inactive" {
			active := false
			filter.IsActive = &active
		}
	}

	// Use the repository's ListCatalogItems method
	labour, _, err := s.catalogRepo.ListCatalogItems(ctx, filter, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list labour: %w", err)
	}

	return labour, nil
}

// Generic catalog operations
func (s *CatalogService) GetCatalogItemByID(ctx context.Context, id string) (*catalogModels.CatalogItem, error) {
	var item catalogModels.CatalogItem
	retrievedItem, err := s.catalogRepo.GetByID(ctx, id, &item)
	if err != nil {
		return nil, err
	}
	return retrievedItem.(*catalogModels.CatalogItem), nil
}

func (s *CatalogService) ListCatalogItems(ctx context.Context, filter *catalogRequests.CatalogFilter, offset, limit int) ([]*catalogModels.CatalogItem, int, error) {
	// Use the repository's ListCatalogItems method with the provided filter
	items, total, err := s.catalogRepo.ListCatalogItems(ctx, filter, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list catalog items: %w", err)
	}

	return items, total, nil
}

func (s *CatalogService) SearchCatalog(ctx context.Context, query string, filter *catalogRequests.CatalogFilter, offset, limit int) ([]*catalogModels.CatalogItem, int, error) {
	// Add search query to the filter
	if filter == nil {
		filter = &catalogRequests.CatalogFilter{}
	}
	filter.Search = &query

	// Use the repository's ListCatalogItems method with search filter
	items, total, err := s.catalogRepo.ListCatalogItems(ctx, filter, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search catalog: %w", err)
	}

	return items, total, nil
}

// Inventory operations
func (s *CatalogService) GetInventoryLevel(ctx context.Context, itemID string) (float64, error) {
	// Check if the catalog item exists and is a product
	catalogItem := &catalogModels.CatalogItem{}
	retrievedItem, err := s.catalogRepo.GetByID(ctx, itemID, catalogItem)
	if err != nil {
		return 0, fmt.Errorf("failed to get catalog item: %w", err)
	}

	catalogItem = retrievedItem.(*catalogModels.CatalogItem)

	if catalogItem.ItemType != catalogModels.CatalogItemTypeProduct {
		return 0, fmt.Errorf("inventory is only available for products")
	}

	// Get inventory level from repository
	level, err := s.catalogRepo.GetInventoryLevel(ctx, itemID)
	if err != nil {
		return 0, fmt.Errorf("failed to get inventory level: %w", err)
	}

	return level, nil
}

func (s *CatalogService) UpdateInventory(ctx context.Context, itemID string, quantity float64, operation string) error {
	// Validate operation
	if operation != "reserve" && operation != "release" && operation != "sell" {
		return fmt.Errorf("invalid operation: %s. Must be 'reserve', 'release', or 'sell'", operation)
	}

	if quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	// Check if the catalog item exists and is a product
	catalogItem := &catalogModels.CatalogItem{}
	retrievedItem, err := s.catalogRepo.GetByID(ctx, itemID, catalogItem)
	if err != nil {
		return fmt.Errorf("failed to get catalog item: %w", err)
	}

	catalogItem = retrievedItem.(*catalogModels.CatalogItem)

	if catalogItem.ItemType != catalogModels.CatalogItemTypeProduct {
		return fmt.Errorf("inventory operations are only available for products")
	}

	// Perform inventory operation based on the operation type
	switch operation {
	case "reserve":
		if err := s.catalogRepo.ReserveInventory(ctx, itemID, quantity); err != nil {
			return fmt.Errorf("failed to reserve inventory: %w", err)
		}
	case "release":
		if err := s.catalogRepo.ReleaseInventory(ctx, itemID, quantity); err != nil {
			return fmt.Errorf("failed to release inventory: %w", err)
		}
	case "sell":
		// For sell operation, we would typically need to handle this differently
		// For now, we'll treat it as a release operation
		if err := s.catalogRepo.ReleaseInventory(ctx, itemID, quantity); err != nil {
			return fmt.Errorf("failed to sell inventory: %w", err)
		}
	}

	return nil
}

// validateServiceRequest validates service creation request
func (s *CatalogService) validateServiceRequest(req *catalogRequests.CreateServiceRequest) error {
	if req == nil {
		return fmt.Errorf("service request cannot be nil")
	}

	if req.Name == "" {
		return fmt.Errorf("service name is required")
	}

	if req.BasePrice.LessThan(decimal.Zero) {
		return fmt.Errorf("service price cannot be negative")
	}

	if req.Currency == "" {
		req.Currency = "INR" // Set default currency
	}

	return nil
}

// validateService validates service data
func (s *CatalogService) validateService(service *catalogModels.Service) error {
	if service == nil {
		return fmt.Errorf("service cannot be nil")
	}

	if service.Name == "" {
		return fmt.Errorf("service name is required")
	}

	if service.BasePrice.LessThan(decimal.Zero) {
		return fmt.Errorf("service price cannot be negative")
	}

	if service.Currency == "" {
		return fmt.Errorf("service currency is required")
	}

	return nil
}

// validateLabourRequest validates labour creation request
func (s *CatalogService) validateLabourRequest(req *catalogRequests.CreateLabourRequest) error {
	if req == nil {
		return fmt.Errorf("labour request cannot be nil")
	}

	if req.Name == "" {
		return fmt.Errorf("labour name is required")
	}

	if req.BasePrice.LessThan(decimal.Zero) {
		return fmt.Errorf("labour price cannot be negative")
	}

	if req.Currency == "" {
		req.Currency = "INR" // Set default currency
	}

	return nil
}

// validateLabour validates labour data
func (s *CatalogService) validateLabour(labour *catalogModels.Labour) error {
	if labour == nil {
		return fmt.Errorf("labour cannot be nil")
	}

	if labour.Name == "" {
		return fmt.Errorf("labour name is required")
	}

	if labour.BasePrice.LessThan(decimal.Zero) {
		return fmt.Errorf("labour price cannot be negative")
	}

	if labour.Currency == "" {
		return fmt.Errorf("labour currency is required")
	}

	return nil
}

// createServiceFromRequest creates a Service model from CreateServiceRequest
func (s *CatalogService) createServiceFromRequest(req *catalogRequests.CreateServiceRequest, userID string) *catalogModels.Service {
	service := catalogModels.NewService("", req.Name, req.BasePrice)

	// Set basic catalog item fields
	service.Category = req.Category
	service.Subcategory = req.Subcategory
	service.Description = req.Description
	service.SKU = req.SKU
	service.UnitOfMeasure = req.UnitOfMeasure
	service.Currency = req.Currency
	if service.Currency == "" {
		service.Currency = "INR"
	}

	if req.Visibility != "" {
		service.Visibility = req.Visibility
	}

	if req.Tags != nil {
		service.Tags = req.Tags
	}

	if req.Images != nil {
		service.Images = req.Images
	}

	// Set service-specific fields
	if req.DurationMinutes != nil {
		service.DurationMinutes = req.DurationMinutes
	}

	// Set audit fields
	service.SetCreatedBy(userID)
	service.SetUpdatedBy(userID)

	return service
}

// createLabourFromRequest creates a Labour model from CreateLabourRequest
func (s *CatalogService) createLabourFromRequest(req *catalogRequests.CreateLabourRequest, userID string) *catalogModels.Labour {
	labour := catalogModels.NewLabour("", req.Name, req.BasePrice)

	// Set basic catalog item fields
	labour.Category = req.Category
	labour.Subcategory = req.Subcategory
	labour.Description = req.Description
	labour.SKU = req.SKU
	labour.UnitOfMeasure = req.UnitOfMeasure
	labour.Currency = req.Currency
	if labour.Currency == "" {
		labour.Currency = "INR"
	}

	if req.Visibility != "" {
		labour.Visibility = req.Visibility
	}

	if req.Tags != nil {
		labour.Tags = req.Tags
	}

	if req.Images != nil {
		labour.Images = req.Images
	}

	// Set labour-specific fields
	labour.SkillLevel = req.SkillLevel
	if req.HourlyRate != nil {
		labour.HourlyRate = req.HourlyRate
	}

	// Set audit fields
	labour.SetCreatedBy(userID)
	labour.SetUpdatedBy(userID)

	return labour
}

// UpdateCatalogItem updates a catalog item based on its type
func (s *CatalogService) UpdateCatalogItem(ctx context.Context, catalogItem *catalogModels.CatalogItem, userID string) (*catalogModels.CatalogItem, error) {
	if catalogItem == nil {
		return nil, fmt.Errorf("catalog item cannot be nil")
	}

	// Set updated by
	catalogItem.SetUpdatedBy(userID)

	// Update the catalog item
	if err := s.catalogRepo.Update(ctx, catalogItem); err != nil {
		return nil, fmt.Errorf("failed to update catalog item: %w", err)
	}

	return catalogItem, nil
}

// DeleteContract deletes a contract (placeholder implementation)
func (s *CatalogService) DeleteContract(ctx context.Context, contractID string, userID string) error {
	// For now, treat contracts as regular catalog items
	// In the future, this might need special handling
	catalogItem := &catalogModels.CatalogItem{}
	result, err := s.catalogRepo.GetByID(ctx, contractID, catalogItem)
	if err != nil {
		return fmt.Errorf("contract not found: %w", err)
	}
	catalogItem = result.(*catalogModels.CatalogItem)

	// Set deleted by
	catalogItem.SetUpdatedBy(userID)

	// Soft delete the contract
	if err := s.catalogRepo.Delete(ctx, contractID, catalogItem); err != nil {
		return fmt.Errorf("failed to delete contract: %w", err)
	}

	return nil
}

// UpdateActiveStatus updates the is_active field of a catalog item
func (s *CatalogService) UpdateActiveStatus(ctx context.Context, id string, isActive bool, userID string) error {
	catalogItem := &catalogModels.CatalogItem{}
	result, err := s.catalogRepo.GetByID(ctx, id, catalogItem)
	if err != nil {
		return fmt.Errorf("catalog item not found: %w", err)
	}
	catalogItem = result.(*catalogModels.CatalogItem)

	// Update the is_active field
	catalogItem.IsActive = isActive
	catalogItem.SetUpdatedBy(userID)

	// Save the updated item
	if err := s.catalogRepo.Update(ctx, catalogItem); err != nil {
		return fmt.Errorf("failed to update catalog item active status: %w", err)
	}

	return nil
}
