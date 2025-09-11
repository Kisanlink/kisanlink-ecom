package catalog

import (
    "context"
    "errors"
    "testing"

    catalogModels "kisanlink-ecom/entities/models/catalog"
    catalogRequests "kisanlink-ecom/entities/requests/catalog"
    catalogService "kisanlink-ecom/internal/services/catalog"

    "github.com/shopspring/decimal"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockCatalogRepository is a comprehensive mock for catalog repository
type MockCatalogRepository struct {
    mock.Mock
}

func (m *MockCatalogRepository) GetByID(ctx context.Context, id string, model interface{}) (interface{}, error) {
    args := m.Called(ctx, id, model)
    return args.Get(0), args.Error(1)
}

func (m *MockCatalogRepository) Create(ctx context.Context, model interface{}) error {
    args := m.Called(ctx, model)
    return args.Error(0)
}

func (m *MockCatalogRepository) Update(ctx context.Context, model interface{}) error {
    args := m.Called(ctx, model)
    return args.Error(0)
}

func (m *MockCatalogRepository) Delete(ctx context.Context, id string, model interface{}) error {
    args := m.Called(ctx, id, model)
    return args.Error(0)
}

func (m *MockCatalogRepository) SoftDelete(ctx context.Context, id string, deletedBy string) error {
    args := m.Called(ctx, id, deletedBy)
    return args.Error(0)
}

func (m *MockCatalogRepository) Restore(ctx context.Context, id string) error {
    args := m.Called(ctx, id)
    return args.Error(0)
}

func (m *MockCatalogRepository) GetBySKU(ctx context.Context, sku string) (*catalogModels.CatalogItem, error) {
    args := m.Called(ctx, sku)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*catalogModels.CatalogItem), args.Error(1)
}

func (m *MockCatalogRepository) ListCatalogItems(ctx context.Context, filter *catalogRequests.CatalogFilter, offset, limit int) ([]*catalogModels.CatalogItem, int, error) {
    args := m.Called(ctx, filter, offset, limit)
    return args.Get(0).([]*catalogModels.CatalogItem), args.Int(1), args.Error(2)
}

func (m *MockCatalogRepository) GetInventoryLevel(ctx context.Context, itemID string) (float64, error) {
    args := m.Called(ctx, itemID)
    return args.Get(0).(float64), args.Error(1)
}

func (m *MockCatalogRepository) ReserveInventory(ctx context.Context, itemID string, quantity float64) error {
    args := m.Called(ctx, itemID, quantity)
    return args.Error(0)
}

func (m *MockCatalogRepository) ReleaseInventory(ctx context.Context, itemID string, quantity float64) error {
    args := m.Called(ctx, itemID, quantity)
    return args.Error(0)
}

// Test helper functions
func createTestCatalogItem(itemType catalogModels.CatalogItemType) *catalogModels.CatalogItem {
    item := &catalogModels.CatalogItem{
        OrganizationID: "org-123",
        ItemType:       itemType,
        Name:           "Test Item",
        Description:    "Test Description",
        BasePrice:      decimal.NewFromFloat(100.0),
        Currency:       "INR",
        Category:       "test-category",
        SKU:            "TEST-001",
        IsActive:       true,
        Visibility:     catalogModels.VisibilityTypeOrg,
    }
    item.ID = "item-123"
    return item
}

func createTestProductRequest() *catalogRequests.CreateProductRequest {
    return &catalogRequests.CreateProductRequest{
        CreateCatalogItemRequest: catalogRequests.CreateCatalogItemRequest{
            Name:        "Test Product",
            Description: "Test Product Description",
            BasePrice:   decimal.NewFromFloat(100.0),
            Currency:    "INR",
            Category:    "agriculture",
            SKU:         "PROD-001",
        },
        Weight:        decimal.NewFromFloat(5.5),
        Perishable:    true,
        ShelfLifeDays: 30,
    }
}

func createTestServiceRequest() *catalogRequests.CreateServiceRequest {
    return &catalogRequests.CreateServiceRequest{
        CreateCatalogItemRequest: catalogRequests.CreateCatalogItemRequest{
            Name:        "Agricultural Consulting",
            Description: "Expert agricultural consulting service",
            BasePrice:   decimal.NewFromFloat(500.0),
            Currency:    "INR",
            Category:    "consulting",
            SKU:         "SERV-001",
        },
        DurationMinutes: 120,
    }
}

func createTestLabourRequest() *catalogRequests.CreateLabourRequest {
    return &catalogRequests.CreateLabourRequest{
        CreateCatalogItemRequest: catalogRequests.CreateCatalogItemRequest{
            Name:        "Farm Worker",
            Description: "Experienced farm worker",
            BasePrice:   decimal.NewFromFloat(50.0),
            Currency:    "INR",
            Category:    "manual-labor",
            SKU:         "LAB-001",
        },
        SkillLevel: "intermediate",
        HourlyRate: decimal.NewFromFloat(50.0),
    }
}

// Test CreateProduct with comprehensive scenarios
func TestCatalogService_CreateProduct_Success(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)
    req := createTestProductRequest()

    // Mock SKU uniqueness check
    mockRepo.On("GetBySKU", mock.Anything, req.SKU).Return(nil, errors.New("not found"))
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*catalog.CatalogItem")).Return(nil)

    result, err := service.CreateProduct(context.Background(), req, "user-123")

    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, req.Name, result.Name)
    assert.Equal(t, catalogModels.CatalogItemTypeProduct, result.ItemType)
    assert.True(t, result.BasePrice.Equal(req.BasePrice))
    assert.Equal(t, req.Weight, result.Weight)
    assert.Equal(t, req.Perishable, result.Perishable)
    assert.Equal(t, req.ShelfLifeDays, result.ShelfLifeDays)

    mockRepo.AssertExpectations(t)
}

func TestCatalogService_CreateProduct_DuplicateSKU(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)
    req := createTestProductRequest()

    existingItem := createTestCatalogItem(catalogModels.CatalogItemTypeProduct)
    mockRepo.On("GetBySKU", mock.Anything, req.SKU).Return(existingItem, nil)

    result, err := service.CreateProduct(context.Background(), req, "user-123")

    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Contains(t, err.Error(), "SKU already exists")

    mockRepo.AssertExpectations(t)
    mockRepo.AssertNotCalled(t, "Create")
}

func TestCatalogService_CreateProduct_RepositoryError(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)
    req := createTestProductRequest()

    mockRepo.On("GetBySKU", mock.Anything, req.SKU).Return(nil, errors.New("not found"))
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*catalog.CatalogItem")).Return(errors.New("database error"))

    result, err := service.CreateProduct(context.Background(), req, "user-123")

    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Contains(t, err.Error(), "database error")

    mockRepo.AssertExpectations(t)
}

// Test CreateService with comprehensive scenarios
func TestCatalogService_CreateService_Success(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)
    req := createTestServiceRequest()

    mockRepo.On("GetBySKU", mock.Anything, req.SKU).Return(nil, errors.New("not found"))
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*catalog.CatalogItem")).Return(nil)

    result, err := service.CreateService(context.Background(), req, "user-123")

    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, req.Name, result.Name)
    assert.Equal(t, catalogModels.CatalogItemTypeService, result.ItemType)
    assert.True(t, result.BasePrice.Equal(req.BasePrice))
    assert.Equal(t, req.DurationMinutes, result.DurationMinutes)

    mockRepo.AssertExpectations(t)
}

func TestCatalogService_CreateService_InvalidDuration(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)
    req := createTestServiceRequest()
    req.DurationMinutes = -30 // Invalid negative duration

    result, err := service.CreateService(context.Background(), req, "user-123")

    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Contains(t, err.Error(), "duration must be positive")

    mockRepo.AssertNotCalled(t, "GetBySKU")
    mockRepo.AssertNotCalled(t, "Create")
}

// Test CreateLabour with comprehensive scenarios
func TestCatalogService_CreateLabour_Success(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)
    req := createTestLabourRequest()

    mockRepo.On("GetBySKU", mock.Anything, req.SKU).Return(nil, errors.New("not found"))
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*catalog.CatalogItem")).Return(nil)

    result, err := service.CreateLabour(context.Background(), req, "user-123")

    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, req.Name, result.Name)
    assert.Equal(t, catalogModels.CatalogItemTypeLabour, result.ItemType)
    assert.True(t, result.BasePrice.Equal(req.BasePrice))
    assert.Equal(t, req.SkillLevel, result.SkillLevel)
    assert.True(t, result.HourlyRate.Equal(req.HourlyRate))

    mockRepo.AssertExpectations(t)
}

func TestCatalogService_CreateLabour_InvalidSkillLevel(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)
    req := createTestLabourRequest()
    req.SkillLevel = "invalid-skill" // Invalid skill level

    result, err := service.CreateLabour(context.Background(), req, "user-123")

    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Contains(t, err.Error(), "invalid skill level")

    mockRepo.AssertNotCalled(t, "GetBySKU")
    mockRepo.AssertNotCalled(t, "Create")
}

func TestCatalogService_CreateLabour_NegativeHourlyRate(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)
    req := createTestLabourRequest()
    req.HourlyRate = decimal.NewFromFloat(-10.0) // Negative hourly rate

    result, err := service.CreateLabour(context.Background(), req, "user-123")

    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Contains(t, err.Error(), "hourly rate must be positive")

    mockRepo.AssertNotCalled(t, "GetBySKU")
    mockRepo.AssertNotCalled(t, "Create")
}

// Test UpdateProduct with comprehensive scenarios
func TestCatalogService_UpdateProduct_Success(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)

    existingProduct := createTestCatalogItem(catalogModels.CatalogItemTypeProduct)
    existingProduct.Weight = decimal.NewFromFloat(5.0)
    existingProduct.Perishable = true
    existingProduct.ShelfLifeDays = 30

    mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*catalog.CatalogItem")).Return(nil)

    result, err := service.UpdateProduct(context.Background(), existingProduct, "user-123")

    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, existingProduct.ID, result.ID)
    assert.Equal(t, existingProduct.Name, result.Name)

    mockRepo.AssertExpectations(t)
}

func TestCatalogService_UpdateProduct_WrongItemType(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)

    existingService := createTestCatalogItem(catalogModels.CatalogItemTypeService)

    result, err := service.UpdateProduct(context.Background(), existingService, "user-123")

    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Contains(t, err.Error(), "not a product")

    mockRepo.AssertNotCalled(t, "Update")
}

// Test GetProductByID with comprehensive scenarios
func TestCatalogService_GetProductByID_Success(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)

    expectedProduct := createTestCatalogItem(catalogModels.CatalogItemTypeProduct)
    expectedProduct.Weight = decimal.NewFromFloat(5.0)

    mockRepo.On("GetByID", mock.Anything, "item-123", mock.AnythingOfType("*catalog.CatalogItem")).Return(expectedProduct, nil).Run(func(args mock.Arguments) {
        item := args.Get(2).(*catalogModels.CatalogItem)
        *item = *expectedProduct
    })

    result, err := service.GetProductByID(context.Background(), "item-123")

    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, expectedProduct.ID, result.ID)
    assert.Equal(t, catalogModels.CatalogItemTypeProduct, result.ItemType)

    mockRepo.AssertExpectations(t)
}

func TestCatalogService_GetProductByID_NotFound(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)

    mockRepo.On("GetByID", mock.Anything, "nonexistent", mock.AnythingOfType("*catalog.CatalogItem")).Return(nil, errors.New("not found"))

    result, err := service.GetProductByID(context.Background(), "nonexistent")

    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Contains(t, err.Error(), "not found")

    mockRepo.AssertExpectations(t)
}

func TestCatalogService_GetProductByID_WrongItemType(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)

    serviceItem := createTestCatalogItem(catalogModels.CatalogItemTypeService)

    mockRepo.On("GetByID", mock.Anything, "item-123", mock.AnythingOfType("*catalog.CatalogItem")).Return(serviceItem, nil).Run(func(args mock.Arguments) {
        item := args.Get(2).(*catalogModels.CatalogItem)
        *item = *serviceItem
    })

    result, err := service.GetProductByID(context.Background(), "item-123")

    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Contains(t, err.Error(), "not a product")

    mockRepo.AssertExpectations(t)
}

// Test ListCatalogItems with comprehensive scenarios
func TestCatalogService_ListCatalogItems_Success(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)

    expectedItems := []*catalogModels.CatalogItem{
        createTestCatalogItem(catalogModels.CatalogItemTypeProduct),
        createTestCatalogItem(catalogModels.CatalogItemTypeService),
    }
    filter := &catalogRequests.CatalogFilter{}

    mockRepo.On("ListCatalogItems", mock.Anything, filter, 0, 20).Return(expectedItems, 2, nil)

    result, total, err := service.ListCatalogItems(context.Background(), filter, 0, 20)

    assert.NoError(t, err)
    assert.Len(t, result, 2)
    assert.Equal(t, 2, total)
    assert.Equal(t, expectedItems[0].ID, result[0].ID)
    assert.Equal(t, expectedItems[1].ID, result[1].ID)

    mockRepo.AssertExpectations(t)
}

func TestCatalogService_ListCatalogItems_WithFilters(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)

    expectedItems := []*catalogModels.CatalogItem{
        createTestCatalogItem(catalogModels.CatalogItemTypeProduct),
    }

    itemType := catalogModels.CatalogItemTypeProduct
    filter := &catalogRequests.CatalogFilter{
        ItemType: &itemType,
    }

    mockRepo.On("ListCatalogItems", mock.Anything, filter, 0, 20).Return(expectedItems, 1, nil)

    result, total, err := service.ListCatalogItems(context.Background(), filter, 0, 20)

    assert.NoError(t, err)
    assert.Len(t, result, 1)
    assert.Equal(t, 1, total)
    assert.Equal(t, catalogModels.CatalogItemTypeProduct, result[0].ItemType)

    mockRepo.AssertExpectations(t)
}

func TestCatalogService_ListCatalogItems_EmptyResult(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)

    filter := &catalogRequests.CatalogFilter{}
    mockRepo.On("ListCatalogItems", mock.Anything, filter, 0, 20).Return([]*catalogModels.CatalogItem{}, 0, nil)

    result, total, err := service.ListCatalogItems(context.Background(), filter, 0, 20)

    assert.NoError(t, err)
    assert.Len(t, result, 0)
    assert.Equal(t, 0, total)

    mockRepo.AssertExpectations(t)
}

// Test SearchCatalog with comprehensive scenarios
func TestCatalogService_SearchCatalog_Success(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)

    expectedItems := []*catalogModels.CatalogItem{
        createTestCatalogItem(catalogModels.CatalogItemTypeProduct),
    }
    expectedItems[0].Name = "Agricultural Product"

    filter := &catalogRequests.CatalogFilter{}
    searchTerm := "agricultural"

    // The search should modify the filter to include the search term
    mockRepo.On("ListCatalogItems", mock.Anything, mock.MatchedBy(func(f *catalogRequests.CatalogFilter) bool {
        return f.Search != nil && *f.Search == searchTerm
    }), 0, 20).Return(expectedItems, 1, nil)

    result, total, err := service.SearchCatalog(context.Background(), searchTerm, filter, 0, 20)

    assert.NoError(t, err)
    assert.Len(t, result, 1)
    assert.Equal(t, 1, total)
    assert.Contains(t, result[0].Name, "Agricultural")

    mockRepo.AssertExpectations(t)
}

func TestCatalogService_SearchCatalog_EmptyQuery(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)

    filter := &catalogRequests.CatalogFilter{}

    result, total, err := service.SearchCatalog(context.Background(), "", filter, 0, 20)

    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Equal(t, 0, total)
    assert.Contains(t, err.Error(), "search query cannot be empty")

    mockRepo.AssertNotCalled(t, "ListCatalogItems")
}

// Test inventory-related methods
func TestCatalogService_GetInventoryLevel_Success(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)

    productItem := createTestCatalogItem(catalogModels.CatalogItemTypeProduct)
    mockRepo.On("GetByID", mock.Anything, "item-123", mock.AnythingOfType("*catalog.CatalogItem")).Return(productItem, nil).Run(func(args mock.Arguments) {
        item := args.Get(2).(*catalogModels.CatalogItem)
        *item = *productItem
    })
    mockRepo.On("GetInventoryLevel", mock.Anything, "item-123").Return(100.0, nil)

    level, err := service.GetInventoryLevel(context.Background(), "item-123")

    assert.NoError(t, err)
    assert.Equal(t, 100.0, level)

    mockRepo.AssertExpectations(t)
}

func TestCatalogService_GetInventoryLevel_NonProduct(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)

    serviceItem := createTestCatalogItem(catalogModels.CatalogItemTypeService)
    mockRepo.On("GetByID", mock.Anything, "item-123", mock.AnythingOfType("*catalog.CatalogItem")).Return(serviceItem, nil).Run(func(args mock.Arguments) {
        item := args.Get(2).(*catalogModels.CatalogItem)
        *item = *serviceItem
    })

    level, err := service.GetInventoryLevel(context.Background(), "item-123")

    assert.Error(t, err)
    assert.Equal(t, 0.0, level)
    assert.Contains(t, err.Error(), "inventory not applicable for non-product items")

    mockRepo.AssertExpectations(t)
    mockRepo.AssertNotCalled(t, "GetInventoryLevel")
}

func TestCatalogService_UpdateInventory_Reserve_Success(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)

    productItem := createTestCatalogItem(catalogModels.CatalogItemTypeProduct)
    mockRepo.On("GetByID", mock.Anything, "item-123", mock.AnythingOfType("*catalog.CatalogItem")).Return(productItem, nil).Run(func(args mock.Arguments) {
        item := args.Get(2).(*catalogModels.CatalogItem)
        *item = *productItem
    })
    mockRepo.On("ReserveInventory", mock.Anything, "item-123", 10.0).Return(nil)

    err := service.UpdateInventory(context.Background(), "item-123", 10.0, "reserve")

    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}

func TestCatalogService_UpdateInventory_Release_Success(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)

    productItem := createTestCatalogItem(catalogModels.CatalogItemTypeProduct)
    mockRepo.On("GetByID", mock.Anything, "item-123", mock.AnythingOfType("*catalog.CatalogItem")).Return(productItem, nil).Run(func(args mock.Arguments) {
        item := args.Get(2).(*catalogModels.CatalogItem)
        *item = *productItem
    })
    mockRepo.On("ReleaseInventory", mock.Anything, "item-123", 5.0).Return(nil)

    err := service.UpdateInventory(context.Background(), "item-123", 5.0, "release")

    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}

func TestCatalogService_UpdateInventory_InvalidOperation(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)

    productItem := createTestCatalogItem(catalogModels.CatalogItemTypeProduct)
    mockRepo.On("GetByID", mock.Anything, "item-123", mock.AnythingOfType("*catalog.CatalogItem")).Return(productItem, nil).Run(func(args mock.Arguments) {
        item := args.Get(2).(*catalogModels.CatalogItem)
        *item = *productItem
    })

    err := service.UpdateInventory(context.Background(), "item-123", 10.0, "invalid-operation")

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "invalid inventory operation")

    mockRepo.AssertExpectations(t)
    mockRepo.AssertNotCalled(t, "ReserveInventory")
    mockRepo.AssertNotCalled(t, "ReleaseInventory")
}

// Test soft delete and restore operations
func TestCatalogService_SoftDeleteCatalogItem_Success(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)

    mockRepo.On("SoftDelete", mock.Anything, "item-123", "user-123").Return(nil)

    err := service.SoftDeleteCatalogItem(context.Background(), "item-123", "user-123")

    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}

func TestCatalogService_RestoreCatalogItem_Success(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)

    mockRepo.On("Restore", mock.Anything, "item-123").Return(nil)

    err := service.RestoreCatalogItem(context.Background(), "item-123")

    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}

// Test edge cases and validation
func TestCatalogService_CreateProduct_EmptyName(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)
    req := createTestProductRequest()
    req.Name = ""

    result, err := service.CreateProduct(context.Background(), req, "user-123")

    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Contains(t, err.Error(), "name is required")

    mockRepo.AssertNotCalled(t, "GetBySKU")
    mockRepo.AssertNotCalled(t, "Create")
}

func TestCatalogService_CreateProduct_NegativePrice(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)
    req := createTestProductRequest()
    req.BasePrice = decimal.NewFromFloat(-10.0)

    result, err := service.CreateProduct(context.Background(), req, "user-123")

    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Contains(t, err.Error(), "price must be positive")

    mockRepo.AssertNotCalled(t, "GetBySKU")
    mockRepo.AssertNotCalled(t, "Create")
}

func TestCatalogService_CreateProduct_EmptyUserID(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)
    req := createTestProductRequest()

    result, err := service.CreateProduct(context.Background(), req, "")

    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Contains(t, err.Error(), "user ID is required")

    mockRepo.AssertNotCalled(t, "GetBySKU")
    mockRepo.AssertNotCalled(t, "Create")
}

// Test concurrent access scenarios
func TestCatalogService_CreateProduct_ConcurrentSKUCreation(t *testing.T) {
    mockRepo := &MockCatalogRepository{}
    service := catalogService.NewCatalogService(mockRepo)
    req := createTestProductRequest()

    // First check passes (SKU not found), but creation fails due to concurrent creation
    mockRepo.On("GetBySKU", mock.Anything, req.SKU).Return(nil, errors.New("not found"))
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*catalog.CatalogItem")).Return(errors.New("duplicate key constraint"))

    result, err := service.CreateProduct(context.Background(), req, "user-123")

    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Contains(t, err.Error(), "duplicate key constraint")

    mockRepo.AssertExpectations(t)
}
