package catalog

import (
	"context"
	"testing"

	catalogModels "github.com/Kisanlink/kisanlink-ecom/entities/models/catalog"
	catalogRequests "github.com/Kisanlink/kisanlink-ecom/entities/requests/catalog"
	catalogService "github.com/Kisanlink/kisanlink-ecom/internal/services/catalog"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func TestCatalogService_CreateService(t *testing.T) {
	mockRepo := &MockCatalogRepository{}
	mockRepo.On("GetBySKU", mock.Anything, mock.Anything).Return(nil, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*catalog.CatalogItem")).Return(nil)

	service := catalogService.NewCatalogService(mockRepo)

	request := &catalogRequests.CreateServiceRequest{
		CreateCatalogItemRequest: catalogRequests.CreateCatalogItemRequest{
			Name:      "Agricultural Consulting",
			BasePrice: decimal.NewFromFloat(100.0),
			Currency:  "INR",
			Category:  "consulting",
			SKU:       "AGR_CONSULT_001",
		},
	}

	result, err := service.CreateService(context.Background(), request, "user123")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, request.Name, result.Name)
	mockRepo.AssertExpectations(t)
}

func TestCatalogService_CreateLabour(t *testing.T) {
	mockRepo := &MockCatalogRepository{}
	mockRepo.On("GetBySKU", mock.Anything, mock.Anything).Return(nil, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*catalog.CatalogItem")).Return(nil)

	service := catalogService.NewCatalogService(mockRepo)

	request := &catalogRequests.CreateLabourRequest{
		CreateCatalogItemRequest: catalogRequests.CreateCatalogItemRequest{
			Name:      "Farm Worker",
			BasePrice: decimal.NewFromFloat(50.0),
			Currency:  "INR",
			Category:  "manual-labor",
			SKU:       "FARM_WORKER_001",
		},
		SkillLevel: "beginner",
	}

	result, err := service.CreateLabour(context.Background(), request, "user123")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, request.Name, result.Name)
	mockRepo.AssertExpectations(t)
}

func TestCatalogService_GetInventoryLevel(t *testing.T) {
	mockRepo := &MockCatalogRepository{}
	productItem := &catalogModels.CatalogItem{
		ItemType: catalogModels.CatalogItemTypeProduct,
		Name:     "Test Product",
	}
	mockRepo.On("GetByID", mock.Anything, "product123", mock.AnythingOfType("*catalog.CatalogItem")).Return(productItem, nil)
	mockRepo.On("GetInventoryLevel", mock.Anything, "product123").Return(100.0, nil)

	service := catalogService.NewCatalogService(mockRepo)

	level, err := service.GetInventoryLevel(context.Background(), "product123")

	assert.NoError(t, err)
	assert.Equal(t, 100.0, level)
	mockRepo.AssertExpectations(t)
}

func TestCatalogService_UpdateInventory(t *testing.T) {
	mockRepo := &MockCatalogRepository{}
	productItem := &catalogModels.CatalogItem{
		ItemType: catalogModels.CatalogItemTypeProduct,
		Name:     "Test Product",
	}
	mockRepo.On("GetByID", mock.Anything, "product123", mock.AnythingOfType("*catalog.CatalogItem")).Return(productItem, nil)
	mockRepo.On("ReserveInventory", mock.Anything, "product123", 10.0).Return(nil)

	service := catalogService.NewCatalogService(mockRepo)

	err := service.UpdateInventory(context.Background(), "product123", 10.0, "reserve")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCatalogService_ListCatalogItems(t *testing.T) {
	mockRepo := &MockCatalogRepository{}
	items := []*catalogModels.CatalogItem{
		{Name: "Product 1", ItemType: catalogModels.CatalogItemTypeProduct},
		{Name: "Service 1", ItemType: catalogModels.CatalogItemTypeService},
	}
	mockRepo.On("ListCatalogItems", mock.Anything, mock.AnythingOfType("*catalog.CatalogFilter"), 0, 10).Return(items, 2, nil)

	service := catalogService.NewCatalogService(mockRepo)

	filter := &catalogRequests.CatalogFilter{}
	result, total, err := service.ListCatalogItems(context.Background(), filter, 0, 10)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 2, total)
	mockRepo.AssertExpectations(t)
}

func TestCatalogService_SearchCatalog(t *testing.T) {
	mockRepo := &MockCatalogRepository{}
	items := []*catalogModels.CatalogItem{
		{Name: "Agricultural Product", ItemType: catalogModels.CatalogItemTypeProduct},
	}
	mockRepo.On("ListCatalogItems", mock.Anything, mock.MatchedBy(func(f *catalogRequests.CatalogFilter) bool {
		return f.Search != nil && *f.Search == "agricultural"
	}), 0, 10).Return(items, 1, nil)

	service := catalogService.NewCatalogService(mockRepo)

	result, total, err := service.SearchCatalog(context.Background(), "agricultural", &catalogRequests.CatalogFilter{}, 0, 10)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, total)
	mockRepo.AssertExpectations(t)
}
