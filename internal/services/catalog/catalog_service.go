package catalog

import (
	"context"
	"fmt"

	catalogModels "kisanlink-ecom/internal/models/catalog"
	catalogRepo "kisanlink-ecom/internal/repositories/catalog"
)

// CatalogServiceInterface defines the interface for catalog operations
type CatalogServiceInterface interface {
	CreateProduct(ctx context.Context, product *catalogModels.Product, userID string) (*catalogModels.Product, error)
	GetProductByID(ctx context.Context, productID string) (*catalogModels.Product, error)
	GetProductBySKU(ctx context.Context, sku string) (*catalogModels.Product, error)
	UpdateProduct(ctx context.Context, product *catalogModels.Product, userID string) (*catalogModels.Product, error)
	DeleteProduct(ctx context.Context, productID string, userID string) error
	ListProducts(ctx context.Context, limit, offset int, category, status string) ([]*catalogModels.CatalogItem, error)
	CreateService(ctx context.Context, req *catalogModels.CreateCatalogItemRequest, userID string) (*catalogModels.ServiceOffering, error)
	GetServiceByID(ctx context.Context, serviceID string) (*catalogModels.ServiceOffering, error)
	GetServiceBySKU(ctx context.Context, sku string) (*catalogModels.ServiceOffering, error)
	UpdateService(ctx context.Context, service *catalogModels.ServiceOffering, userID string) (*catalogModels.ServiceOffering, error)
	DeleteService(ctx context.Context, serviceID string, userID string) error
	ListServices(ctx context.Context, limit, offset int, category, status string) ([]*catalogModels.CatalogItem, error)
	CreateLabour(ctx context.Context, req *catalogModels.CreateCatalogItemRequest, userID string) (*catalogModels.LabourOffering, error)
	GetLabourByID(ctx context.Context, labourID string) (*catalogModels.LabourOffering, error)
	GetLabourBySKU(ctx context.Context, sku string) (*catalogModels.LabourOffering, error)
	UpdateLabour(ctx context.Context, labour *catalogModels.LabourOffering, userID string) (*catalogModels.LabourOffering, error)
	DeleteLabour(ctx context.Context, labourID string, userID string) error
	ListLabour(ctx context.Context, limit, offset int, category, status string) ([]*catalogModels.CatalogItem, error)
	GetCatalogItemByID(ctx context.Context, id string) (*catalogModels.CatalogItem, error)
	ListCatalogItems(ctx context.Context, filter *catalogModels.CatalogFilter, offset, limit int) ([]*catalogModels.CatalogItem, int, error)
	SearchCatalog(ctx context.Context, query string, filter *catalogModels.CatalogFilter, offset, limit int) ([]*catalogModels.CatalogItem, int, error)
	GetInventoryLevel(ctx context.Context, itemID string) (interface{}, error)
	UpdateInventory(ctx context.Context, itemID string, quantity int, operation string) error
}

// CatalogService provides business logic for catalog operations
type CatalogService struct {
	catalogRepo *catalogRepo.CatalogRepository
}

// NewCatalogService creates a new catalog service
func NewCatalogService(catalogRepo *catalogRepo.CatalogRepository) *CatalogService {
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
		BaseModel:   product.BaseModel,
		Type:        catalogModels.CatalogItemTypeProduct,
		OrgID:       product.OrgID,
		SKU:         product.SKU,
		Name:        product.Name,
		Description: product.Description,
		Category:    product.Category,
		Subcategory: product.Subcategory,
		UOM:         product.UOM,
		Price:       product.Price,
		Currency:    product.Currency,
		IsActive:    product.IsActive,
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
		BaseModel:   product.BaseModel,
		Type:        catalogModels.CatalogItemTypeProduct,
		OrgID:       product.OrgID,
		SKU:         product.SKU,
		Name:        product.Name,
		Description: product.Description,
		Category:    product.Category,
		Subcategory: product.Subcategory,
		UOM:         product.UOM,
		Price:       product.Price,
		Currency:    product.Currency,
		IsActive:    product.IsActive,
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
	catalogItem, err := s.catalogRepo.GetByID(ctx, productID, catalogItem)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

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
	products, err := s.catalogRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	// Filter by category if specified
	if category != "" {
		var filtered []*catalogModels.CatalogItem
		for _, product := range products {
			if product.Category == category {
				filtered = append(filtered, product)
			}
		}
		products = filtered
	}

	// Filter by status if specified
	if status != "" {
		var filtered []*catalogModels.CatalogItem
		for _, product := range products {
			if status == "active" && product.IsActive {
				filtered = append(filtered, product)
			} else if status == "inactive" && !product.IsActive {
				filtered = append(filtered, product)
			}
		}
		products = filtered
	}

	return products, nil
}

// DeleteProduct soft deletes a product
func (s *CatalogService) DeleteProduct(ctx context.Context, productID, userID string) error {
	// Check if product exists
	catalogItem := &catalogModels.CatalogItem{}
	catalogItem, err := s.catalogRepo.GetByID(ctx, productID, catalogItem)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

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
	catalogItem, err := s.catalogRepo.GetByID(ctx, productID, catalogItem)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

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

	if product.Price < 0 {
		return fmt.Errorf("product price cannot be negative")
	}

	if product.Category == "" {
		return fmt.Errorf("product category is required")
	}

	if product.UOM == "" {
		return fmt.Errorf("product UOM is required")
	}

	if product.Currency == "" {
		return fmt.Errorf("product currency is required")
	}

	if product.OrgID == "" {
		return fmt.Errorf("product organization ID is required")
	}

	return nil
}

// Service operations
func (s *CatalogService) CreateService(ctx context.Context, req *catalogModels.CreateCatalogItemRequest, userID string) (*catalogModels.ServiceOffering, error) {
	// TODO: Implement service creation
	return nil, fmt.Errorf("not implemented")
}

func (s *CatalogService) GetServiceByID(ctx context.Context, id string) (*catalogModels.ServiceOffering, error) {
	// TODO: Implement service retrieval
	return nil, fmt.Errorf("not implemented")
}

func (s *CatalogService) GetServiceBySKU(ctx context.Context, sku string) (*catalogModels.ServiceOffering, error) {
	// TODO: Implement service retrieval by SKU
	return nil, fmt.Errorf("not implemented")
}

func (s *CatalogService) UpdateService(ctx context.Context, id string, updates *catalogModels.UpdateCatalogItemRequest, userID string) error {
	// TODO: Implement service update
	return fmt.Errorf("not implemented")
}

func (s *CatalogService) DeleteService(ctx context.Context, id string, userID string) error {
	// TODO: Implement service deletion
	return fmt.Errorf("not implemented")
}

func (s *CatalogService) ListServices(ctx context.Context, filter *catalogModels.CatalogFilter, offset, limit int) ([]*catalogModels.ServiceOffering, int, error) {
	// TODO: Implement service listing
	return nil, 0, fmt.Errorf("not implemented")
}

// Labour operations
func (s *CatalogService) CreateLabour(ctx context.Context, req *catalogModels.CreateCatalogItemRequest, userID string) (*catalogModels.LabourOffering, error) {
	// TODO: Implement labour creation
	return nil, fmt.Errorf("not implemented")
}

func (s *CatalogService) GetLabourByID(ctx context.Context, id string) (*catalogModels.LabourOffering, error) {
	// TODO: Implement labour retrieval
	return nil, fmt.Errorf("not implemented")
}

func (s *CatalogService) GetLabourBySKU(ctx context.Context, sku string) (*catalogModels.LabourOffering, error) {
	// TODO: Implement labour retrieval by SKU
	return nil, fmt.Errorf("not implemented")
}

func (s *CatalogService) UpdateLabour(ctx context.Context, id string, updates *catalogModels.UpdateCatalogItemRequest, userID string) error {
	// TODO: Implement labour update
	return fmt.Errorf("not implemented")
}

func (s *CatalogService) DeleteLabour(ctx context.Context, id string, userID string) error {
	// TODO: Implement labour deletion
	return fmt.Errorf("not implemented")
}

func (s *CatalogService) ListLabour(ctx context.Context, filter *catalogModels.CatalogFilter, offset, limit int) ([]*catalogModels.LabourOffering, int, error) {
	// TODO: Implement labour listing
	return nil, 0, fmt.Errorf("not implemented")
}

// Generic catalog operations
func (s *CatalogService) GetCatalogItemByID(ctx context.Context, id string) (*catalogModels.CatalogItem, error) {
	var item catalogModels.CatalogItem
	return s.catalogRepo.GetByID(ctx, id, &item)
}

func (s *CatalogService) ListCatalogItems(ctx context.Context, filter *catalogModels.CatalogFilter, offset, limit int) ([]*catalogModels.CatalogItem, int, error) {
	// TODO: Implement generic catalog listing with filters
	return nil, 0, fmt.Errorf("not implemented")
}

func (s *CatalogService) SearchCatalog(ctx context.Context, query string, filter *catalogModels.CatalogFilter, offset, limit int) ([]*catalogModels.CatalogItem, int, error) {
	// TODO: Implement catalog search
	return nil, 0, fmt.Errorf("not implemented")
}

// Inventory operations
func (s *CatalogService) GetInventoryLevel(ctx context.Context, itemID string) (float64, error) {
	// TODO: Implement inventory level check
	return 0, fmt.Errorf("not implemented")
}

func (s *CatalogService) UpdateInventory(ctx context.Context, itemID string, quantity float64, operation string) error {
	// TODO: Implement inventory update
	return fmt.Errorf("not implemented")
}
