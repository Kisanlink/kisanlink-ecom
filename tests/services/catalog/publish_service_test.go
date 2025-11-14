package catalog_test

import (
	"context"
	"errors"
	"testing"

	catalogModels "github.com/Kisanlink/kisanlink-ecom/entities/models/catalog"
	catalogRequests "github.com/Kisanlink/kisanlink-ecom/entities/requests/catalog"
	catalogService "github.com/Kisanlink/kisanlink-ecom/internal/services/catalog"
	"github.com/Kisanlink/kisanlink-ecom/tests/data"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPublishStateRepository mocks the PublishStateRepository interface
type MockPublishStateRepository struct {
	mock.Mock
}

func (m *MockPublishStateRepository) Create(ctx context.Context, state *catalogModels.PublishState) error {
	args := m.Called(ctx, state)
	return args.Error(0)
}

func (m *MockPublishStateRepository) GetByID(ctx context.Context, id string) (*catalogModels.PublishState, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*catalogModels.PublishState), args.Error(1)
}

func (m *MockPublishStateRepository) GetByProductID(ctx context.Context, productID string) (*catalogModels.PublishState, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*catalogModels.PublishState), args.Error(1)
}

func (m *MockPublishStateRepository) Update(ctx context.Context, state *catalogModels.PublishState) error {
	args := m.Called(ctx, state)
	return args.Error(0)
}

func (m *MockPublishStateRepository) UpdateDeliveryCosts(ctx context.Context, productID string, costs map[string]decimal.Decimal) error {
	args := m.Called(ctx, productID, costs)
	return args.Error(0)
}

func (m *MockPublishStateRepository) UpdateFPOAccessList(ctx context.Context, productID string, fpoIDs []string) error {
	args := m.Called(ctx, productID, fpoIDs)
	return args.Error(0)
}

func (m *MockPublishStateRepository) AddFPOAccess(ctx context.Context, productID string, fpoOrgID string, deliveryCost decimal.Decimal) error {
	args := m.Called(ctx, productID, fpoOrgID, deliveryCost)
	return args.Error(0)
}

func (m *MockPublishStateRepository) RemoveFPOAccess(ctx context.Context, productID string, fpoOrgID string) error {
	args := m.Called(ctx, productID, fpoOrgID)
	return args.Error(0)
}

func (m *MockPublishStateRepository) HasFPOAccess(ctx context.Context, productID, fpoOrgID string) (bool, error) {
	args := m.Called(ctx, productID, fpoOrgID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPublishStateRepository) GetProductsVisibleToFPO(ctx context.Context, fpoOrgID string, offset, limit int) ([]*catalogModels.PublishState, int, error) {
	args := m.Called(ctx, fpoOrgID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*catalogModels.PublishState), args.Int(1), args.Error(2)
}

func (m *MockPublishStateRepository) GetAllPublishedProducts(ctx context.Context, offset, limit int) ([]*catalogModels.PublishState, int, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*catalogModels.PublishState), args.Int(1), args.Error(2)
}

func (m *MockPublishStateRepository) Delete(ctx context.Context, productID string, deletedBy string) error {
	args := m.Called(ctx, productID, deletedBy)
	return args.Error(0)
}

// MockCatalogRepository mocks the CatalogRepositoryInterface
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

// TestCalculateRetailPrice tests the retail price calculation logic
func TestCalculateRetailPrice(t *testing.T) {
	tests := []struct {
		name               string
		basePrice          decimal.Decimal
		deliveryCost       decimal.Decimal
		platformFeePercent decimal.Decimal
		expectedRetail     decimal.Decimal
	}{
		{
			name:               "standard case - base=1000, delivery=50, fee=10%",
			basePrice:          decimal.NewFromFloat(1000.00),
			deliveryCost:       decimal.NewFromFloat(50.00),
			platformFeePercent: decimal.NewFromFloat(10.00),
			expectedRetail:     decimal.NewFromFloat(1150.00), // 1000 + 50 + 100
		},
		{
			name:               "zero fee case",
			basePrice:          decimal.NewFromFloat(1000.00),
			deliveryCost:       decimal.NewFromFloat(50.00),
			platformFeePercent: decimal.Zero,
			expectedRetail:     decimal.NewFromFloat(1050.00), // 1000 + 50 + 0
		},
		{
			name:               "high fee case - 20%",
			basePrice:          decimal.NewFromFloat(1000.00),
			deliveryCost:       decimal.NewFromFloat(50.00),
			platformFeePercent: decimal.NewFromFloat(20.00),
			expectedRetail:     decimal.NewFromFloat(1250.00), // 1000 + 50 + 200
		},
		{
			name:               "zero delivery cost",
			basePrice:          decimal.NewFromFloat(500.00),
			deliveryCost:       decimal.Zero,
			platformFeePercent: decimal.NewFromFloat(10.00),
			expectedRetail:     decimal.NewFromFloat(550.00), // 500 + 0 + 50
		},
		{
			name:               "fractional values",
			basePrice:          decimal.NewFromFloat(999.99),
			deliveryCost:       decimal.NewFromFloat(49.99),
			platformFeePercent: decimal.NewFromFloat(10.50),
			expectedRetail:     decimal.NewFromFloat(1154.97895), // 999.99 + 49.99 + 104.99895
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockPublishRepo := new(MockPublishStateRepository)
			mockCatalogRepo := new(MockCatalogRepository)
			logger := logrus.New()

			service := catalogService.NewPublishService(mockPublishRepo, mockCatalogRepo, logger)

			// Execute
			result := service.CalculateRetailPrice(tt.basePrice, tt.deliveryCost, tt.platformFeePercent)

			// Assert
			assert.True(t, tt.expectedRetail.Equal(result),
				"Expected %s but got %s", tt.expectedRetail.String(), result.String())
		})
	}
}

// TestPublishProduct_Success tests successful product publishing
func TestPublishProduct_Success(t *testing.T) {
	tests := []struct {
		name string
		req  catalogService.PublishProductRequest
	}{
		{
			name: "publish to single FPO",
			req: catalogService.PublishProductRequest{
				ProductID: "PROD00000001",
				FPOIDs:    []string{"ORGN00000002"},
				DeliveryCosts: map[string]decimal.Decimal{
					"ORGN00000002": decimal.NewFromFloat(50.00),
				},
				PlatformFeePercent: decimal.NewFromFloat(10.00),
				PublishedBy:        "admin-user-123",
			},
		},
		{
			name: "publish to multiple FPOs",
			req: catalogService.PublishProductRequest{
				ProductID: "PROD00000001",
				FPOIDs:    []string{"ORGN00000002", "ORGN00000003", "ORGN00000004"},
				DeliveryCosts: map[string]decimal.Decimal{
					"ORGN00000002": decimal.NewFromFloat(50.00),
					"ORGN00000003": decimal.NewFromFloat(75.00),
					"ORGN00000004": decimal.NewFromFloat(100.00),
				},
				PlatformFeePercent: decimal.NewFromFloat(10.00),
				PublishedBy:        "admin-user-123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockPublishRepo := new(MockPublishStateRepository)
			mockCatalogRepo := new(MockCatalogRepository)
			logger := logrus.New()

			activeProduct := data.CreateTestActiveProduct()
			activeProduct.BaseModel.ID = tt.req.ProductID

			// Mock expectations
			mockCatalogRepo.On("GetByID", mock.Anything, tt.req.ProductID, mock.AnythingOfType("*catalog.CatalogItem")).
				Run(func(args mock.Arguments) {
					item := args.Get(2).(*catalogModels.CatalogItem)
					*item = *activeProduct
				}).
				Return(activeProduct, nil)

			mockPublishRepo.On("GetByProductID", mock.Anything, tt.req.ProductID).
				Return(nil, errors.New("not found"))

			mockPublishRepo.On("Create", mock.Anything, mock.AnythingOfType("*catalog.PublishState")).
				Return(nil)

			service := catalogService.NewPublishService(mockPublishRepo, mockCatalogRepo, logger)

			// Execute
			result, err := service.PublishProduct(context.Background(), tt.req)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.req.ProductID, result.ProductID)
			assert.Equal(t, len(tt.req.FPOIDs), len(result.FPOAccessList))
			assert.True(t, tt.req.PlatformFeePercent.Equal(result.PlatformFeePercent))

			mockCatalogRepo.AssertExpectations(t)
			mockPublishRepo.AssertExpectations(t)
		})
	}
}

// TestPublishProduct_ValidationErrors tests publishing validation failures
func TestPublishProduct_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		req         catalogService.PublishProductRequest
		setupMocks  func(*MockPublishStateRepository, *MockCatalogRepository)
		expectedErr string
	}{
		{
			name: "product not found",
			req: catalogService.PublishProductRequest{
				ProductID: "PROD00000099",
				FPOIDs:    []string{"ORGN00000002"},
				DeliveryCosts: map[string]decimal.Decimal{
					"ORGN00000002": decimal.NewFromFloat(50.00),
				},
				PlatformFeePercent: decimal.NewFromFloat(10.00),
				PublishedBy:        "admin-user-123",
			},
			setupMocks: func(_ *MockPublishStateRepository, mcr *MockCatalogRepository) {
				mcr.On("GetByID", mock.Anything, "PROD00000099", mock.AnythingOfType("*catalog.CatalogItem")).
					Return(nil, errors.New("product not found"))
			},
			expectedErr: "product not found",
		},
		{
			name: "inactive product",
			req: catalogService.PublishProductRequest{
				ProductID: "PROD00000001",
				FPOIDs:    []string{"ORGN00000002"},
				DeliveryCosts: map[string]decimal.Decimal{
					"ORGN00000002": decimal.NewFromFloat(50.00),
				},
				PlatformFeePercent: decimal.NewFromFloat(10.00),
				PublishedBy:        "admin-user-123",
			},
			setupMocks: func(_ *MockPublishStateRepository, mcr *MockCatalogRepository) {
				inactiveProduct := data.CreateTestInactiveProduct()
				mcr.On("GetByID", mock.Anything, "PROD00000001", mock.AnythingOfType("*catalog.CatalogItem")).
					Run(func(args mock.Arguments) {
						item := args.Get(2).(*catalogModels.CatalogItem)
						*item = *inactiveProduct
					}).
					Return(inactiveProduct, nil)
			},
			expectedErr: "cannot publish inactive product",
		},
		{
			name: "platform fee too high",
			req: catalogService.PublishProductRequest{
				ProductID: "PROD00000001",
				FPOIDs:    []string{"ORGN00000002"},
				DeliveryCosts: map[string]decimal.Decimal{
					"ORGN00000002": decimal.NewFromFloat(50.00),
				},
				PlatformFeePercent: decimal.NewFromFloat(150.00),
				PublishedBy:        "admin-user-123",
			},
			setupMocks: func(_ *MockPublishStateRepository, mcr *MockCatalogRepository) {
				activeProduct := data.CreateTestActiveProduct()
				mcr.On("GetByID", mock.Anything, "PROD00000001", mock.AnythingOfType("*catalog.CatalogItem")).
					Run(func(args mock.Arguments) {
						item := args.Get(2).(*catalogModels.CatalogItem)
						*item = *activeProduct
					}).
					Return(activeProduct, nil)
			},
			expectedErr: "platform fee must be between 0 and 100",
		},
		{
			name: "negative platform fee",
			req: catalogService.PublishProductRequest{
				ProductID: "PROD00000001",
				FPOIDs:    []string{"ORGN00000002"},
				DeliveryCosts: map[string]decimal.Decimal{
					"ORGN00000002": decimal.NewFromFloat(50.00),
				},
				PlatformFeePercent: decimal.NewFromFloat(-5.00),
				PublishedBy:        "admin-user-123",
			},
			setupMocks: func(_ *MockPublishStateRepository, mcr *MockCatalogRepository) {
				activeProduct := data.CreateTestActiveProduct()
				mcr.On("GetByID", mock.Anything, "PROD00000001", mock.AnythingOfType("*catalog.CatalogItem")).
					Run(func(args mock.Arguments) {
						item := args.Get(2).(*catalogModels.CatalogItem)
						*item = *activeProduct
					}).
					Return(activeProduct, nil)
			},
			expectedErr: "platform fee must be between 0 and 100",
		},
		{
			name: "no FPO IDs provided",
			req: catalogService.PublishProductRequest{
				ProductID:          "PROD00000001",
				FPOIDs:             []string{},
				DeliveryCosts:      map[string]decimal.Decimal{},
				PlatformFeePercent: decimal.NewFromFloat(10.00),
				PublishedBy:        "admin-user-123",
			},
			setupMocks: func(_ *MockPublishStateRepository, mcr *MockCatalogRepository) {
				activeProduct := data.CreateTestActiveProduct()
				mcr.On("GetByID", mock.Anything, "PROD00000001", mock.AnythingOfType("*catalog.CatalogItem")).
					Run(func(args mock.Arguments) {
						item := args.Get(2).(*catalogModels.CatalogItem)
						*item = *activeProduct
					}).
					Return(activeProduct, nil)
			},
			expectedErr: "at least one FPO ID is required",
		},
		{
			name: "negative delivery cost",
			req: catalogService.PublishProductRequest{
				ProductID: "PROD00000001",
				FPOIDs:    []string{"ORGN00000002"},
				DeliveryCosts: map[string]decimal.Decimal{
					"ORGN00000002": decimal.NewFromFloat(-50.00),
				},
				PlatformFeePercent: decimal.NewFromFloat(10.00),
				PublishedBy:        "admin-user-123",
			},
			setupMocks: func(_ *MockPublishStateRepository, mcr *MockCatalogRepository) {
				activeProduct := data.CreateTestActiveProduct()
				mcr.On("GetByID", mock.Anything, "PROD00000001", mock.AnythingOfType("*catalog.CatalogItem")).
					Run(func(args mock.Arguments) {
						item := args.Get(2).(*catalogModels.CatalogItem)
						*item = *activeProduct
					}).
					Return(activeProduct, nil)
			},
			expectedErr: "delivery cost for FPO ORGN00000002 cannot be negative",
		},
		{
			name: "missing delivery cost for FPO",
			req: catalogService.PublishProductRequest{
				ProductID: "PROD00000001",
				FPOIDs:    []string{"ORGN00000002", "ORGN00000003"},
				DeliveryCosts: map[string]decimal.Decimal{
					"ORGN00000002": decimal.NewFromFloat(50.00),
					// Missing cost for ORGN00000003
				},
				PlatformFeePercent: decimal.NewFromFloat(10.00),
				PublishedBy:        "admin-user-123",
			},
			setupMocks: func(_ *MockPublishStateRepository, mcr *MockCatalogRepository) {
				activeProduct := data.CreateTestActiveProduct()
				mcr.On("GetByID", mock.Anything, "PROD00000001", mock.AnythingOfType("*catalog.CatalogItem")).
					Run(func(args mock.Arguments) {
						item := args.Get(2).(*catalogModels.CatalogItem)
						*item = *activeProduct
					}).
					Return(activeProduct, nil)
			},
			expectedErr: "delivery cost not provided for FPO",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockPublishRepo := new(MockPublishStateRepository)
			mockCatalogRepo := new(MockCatalogRepository)
			logger := logrus.New()

			tt.setupMocks(mockPublishRepo, mockCatalogRepo)

			service := catalogService.NewPublishService(mockPublishRepo, mockCatalogRepo, logger)

			// Execute
			result, err := service.PublishProduct(context.Background(), tt.req)

			// Assert
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.expectedErr)

			mockCatalogRepo.AssertExpectations(t)
			mockPublishRepo.AssertExpectations(t)
		})
	}
}

// TestPublishProduct_UpdateExisting tests updating existing publish state
func TestPublishProduct_UpdateExisting(t *testing.T) {
	// Setup
	mockPublishRepo := new(MockPublishStateRepository)
	mockCatalogRepo := new(MockCatalogRepository)
	logger := logrus.New()

	productID := "PROD00000001"
	activeProduct := data.CreateTestActiveProduct()
	activeProduct.BaseModel.ID = productID

	existingState := data.CreateTestPublishStateWithProductID(productID)

	req := catalogService.PublishProductRequest{
		ProductID: productID,
		FPOIDs:    []string{"ORGN00000002", "ORGN00000004"},
		DeliveryCosts: map[string]decimal.Decimal{
			"ORGN00000002": decimal.NewFromFloat(60.00),
			"ORGN00000004": decimal.NewFromFloat(80.00),
		},
		PlatformFeePercent: decimal.NewFromFloat(12.00),
		PublishedBy:        "admin-user-456",
	}

	// Mock expectations
	mockCatalogRepo.On("GetByID", mock.Anything, productID, mock.AnythingOfType("*catalog.CatalogItem")).
		Run(func(args mock.Arguments) {
			item := args.Get(2).(*catalogModels.CatalogItem)
			*item = *activeProduct
		}).
		Return(activeProduct, nil)

	mockPublishRepo.On("GetByProductID", mock.Anything, productID).
		Return(existingState, nil)

	mockPublishRepo.On("Update", mock.Anything, mock.AnythingOfType("*catalog.PublishState")).
		Return(nil)

	service := catalogService.NewPublishService(mockPublishRepo, mockCatalogRepo, logger)

	// Execute
	result, err := service.PublishProduct(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, productID, result.ProductID)
	assert.Equal(t, 2, len(result.FPOAccessList))
	assert.True(t, req.PlatformFeePercent.Equal(result.PlatformFeePercent))

	mockCatalogRepo.AssertExpectations(t)
	mockPublishRepo.AssertExpectations(t)
}

// TestGetPublishStatus tests retrieving publish status
func TestGetPublishStatus(t *testing.T) {
	tests := []struct {
		name        string
		productID   string
		setupMocks  func(*MockPublishStateRepository)
		expectError bool
	}{
		{
			name:      "successful retrieval",
			productID: "PROD00000001",
			setupMocks: func(mpr *MockPublishStateRepository) {
				publishState := data.CreateTestPublishStateWithProductID("PROD00000001")
				mpr.On("GetByProductID", mock.Anything, "PROD00000001").
					Return(publishState, nil)
			},
			expectError: false,
		},
		{
			name:      "product not published",
			productID: "PROD00000099",
			setupMocks: func(mpr *MockPublishStateRepository) {
				mpr.On("GetByProductID", mock.Anything, "PROD00000099").
					Return(nil, errors.New("publish state not found"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockPublishRepo := new(MockPublishStateRepository)
			mockCatalogRepo := new(MockCatalogRepository)
			logger := logrus.New()

			tt.setupMocks(mockPublishRepo)

			service := catalogService.NewPublishService(mockPublishRepo, mockCatalogRepo, logger)

			// Execute
			result, err := service.GetPublishStatus(context.Background(), tt.productID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.productID, result.ProductID)
			}

			mockPublishRepo.AssertExpectations(t)
		})
	}
}

// TestUpdateDeliveryCosts tests updating delivery costs
func TestUpdateDeliveryCosts(t *testing.T) {
	tests := []struct {
		name        string
		productID   string
		costs       map[string]decimal.Decimal
		setupMocks  func(*MockPublishStateRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name:      "successful update",
			productID: "PROD00000001",
			costs: map[string]decimal.Decimal{
				"ORGN00000002": decimal.NewFromFloat(60.00),
				"ORGN00000003": decimal.NewFromFloat(80.00),
			},
			setupMocks: func(mpr *MockPublishStateRepository) {
				publishState := data.CreateTestPublishStateWithProductID("PROD00000001")
				mpr.On("GetByProductID", mock.Anything, "PROD00000001").
					Return(publishState, nil)
				mpr.On("Update", mock.Anything, mock.AnythingOfType("*catalog.PublishState")).
					Return(nil)
			},
			expectError: false,
		},
		{
			name:      "negative delivery cost",
			productID: "PROD00000001",
			costs: map[string]decimal.Decimal{
				"ORGN00000002": decimal.NewFromFloat(-50.00),
			},
			setupMocks:  func(mpr *MockPublishStateRepository) {},
			expectError: true,
			errorMsg:    "delivery cost for FPO ORGN00000002 cannot be negative",
		},
		{
			name:      "product not published",
			productID: "PROD00000099",
			costs: map[string]decimal.Decimal{
				"ORGN00000002": decimal.NewFromFloat(50.00),
			},
			setupMocks: func(mpr *MockPublishStateRepository) {
				mpr.On("GetByProductID", mock.Anything, "PROD00000099").
					Return(nil, errors.New("publish state not found"))
			},
			expectError: true,
			errorMsg:    "failed to get publish state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockPublishRepo := new(MockPublishStateRepository)
			mockCatalogRepo := new(MockCatalogRepository)
			logger := logrus.New()

			tt.setupMocks(mockPublishRepo)

			service := catalogService.NewPublishService(mockPublishRepo, mockCatalogRepo, logger)

			// Execute
			err := service.UpdateDeliveryCosts(context.Background(), tt.productID, tt.costs)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockPublishRepo.AssertExpectations(t)
		})
	}
}

// TestRevokeAccess tests revoking FPO access
func TestRevokeAccess(t *testing.T) {
	tests := []struct {
		name        string
		productID   string
		fpoOrgID    string
		setupMocks  func(*MockPublishStateRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name:      "successful revocation",
			productID: "PROD00000001",
			fpoOrgID:  "ORGN00000002",
			setupMocks: func(mpr *MockPublishStateRepository) {
				publishState := data.CreateTestPublishStateWithProductID("PROD00000001")
				mpr.On("GetByProductID", mock.Anything, "PROD00000001").
					Return(publishState, nil)
				mpr.On("RemoveFPOAccess", mock.Anything, "PROD00000001", "ORGN00000002").
					Return(nil)
			},
			expectError: false,
		},
		{
			name:      "FPO does not have access",
			productID: "PROD00000001",
			fpoOrgID:  "ORGN00000099",
			setupMocks: func(mpr *MockPublishStateRepository) {
				publishState := data.CreateTestPublishStateWithProductID("PROD00000001")
				mpr.On("GetByProductID", mock.Anything, "PROD00000001").
					Return(publishState, nil)
			},
			expectError: true,
			errorMsg:    "does not have access to product",
		},
		{
			name:      "product not published",
			productID: "PROD00000099",
			fpoOrgID:  "ORGN00000002",
			setupMocks: func(mpr *MockPublishStateRepository) {
				mpr.On("GetByProductID", mock.Anything, "PROD00000099").
					Return(nil, errors.New("publish state not found"))
			},
			expectError: true,
			errorMsg:    "failed to get publish state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockPublishRepo := new(MockPublishStateRepository)
			mockCatalogRepo := new(MockCatalogRepository)
			logger := logrus.New()

			tt.setupMocks(mockPublishRepo)

			service := catalogService.NewPublishService(mockPublishRepo, mockCatalogRepo, logger)

			// Execute
			err := service.RevokeAccess(context.Background(), tt.productID, tt.fpoOrgID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockPublishRepo.AssertExpectations(t)
		})
	}
}

// TestValidateFPOAccess tests FPO access validation
func TestValidateFPOAccess(t *testing.T) {
	tests := []struct {
		name           string
		fpoOrgID       string
		productID      string
		setupMocks     func(*MockPublishStateRepository)
		expectedAccess bool
		expectError    bool
	}{
		{
			name:      "FPO has access",
			fpoOrgID:  "ORGN00000002",
			productID: "PROD00000001",
			setupMocks: func(mpr *MockPublishStateRepository) {
				mpr.On("HasFPOAccess", mock.Anything, "PROD00000001", "ORGN00000002").
					Return(true, nil)
			},
			expectedAccess: true,
			expectError:    false,
		},
		{
			name:      "FPO does not have access",
			fpoOrgID:  "ORGN00000099",
			productID: "PROD00000001",
			setupMocks: func(mpr *MockPublishStateRepository) {
				mpr.On("HasFPOAccess", mock.Anything, "PROD00000001", "ORGN00000099").
					Return(false, nil)
			},
			expectedAccess: false,
			expectError:    false,
		},
		{
			name:      "database error",
			fpoOrgID:  "ORGN00000002",
			productID: "PROD00000001",
			setupMocks: func(mpr *MockPublishStateRepository) {
				mpr.On("HasFPOAccess", mock.Anything, "PROD00000001", "ORGN00000002").
					Return(false, errors.New("database error"))
			},
			expectedAccess: false,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockPublishRepo := new(MockPublishStateRepository)
			mockCatalogRepo := new(MockCatalogRepository)
			logger := logrus.New()

			tt.setupMocks(mockPublishRepo)

			service := catalogService.NewPublishService(mockPublishRepo, mockCatalogRepo, logger)

			// Execute
			hasAccess, err := service.ValidateFPOAccess(context.Background(), tt.fpoOrgID, tt.productID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAccess, hasAccess)
			}

			mockPublishRepo.AssertExpectations(t)
		})
	}
}

// TestGetProductsForFPO tests retrieving products for an FPO
func TestGetProductsForFPO(t *testing.T) {
	// Setup
	mockPublishRepo := new(MockPublishStateRepository)
	mockCatalogRepo := new(MockCatalogRepository)
	logger := logrus.New()

	fpoOrgID := "ORGN00000002"
	publishStates := []*catalogModels.PublishState{
		data.CreateTestPublishStateWithProductID("PROD00000001"),
		data.CreateTestPublishStateWithProductID("PROD00000002"),
	}

	activeProduct1 := data.CreateTestActiveProduct()
	activeProduct1.BaseModel.ID = "PROD00000001"

	activeProduct2 := data.CreateTestActiveProduct()
	activeProduct2.BaseModel.ID = "PROD00000002"
	activeProduct2.Name = "Test Organic Rice"

	// Mock expectations
	mockPublishRepo.On("GetProductsVisibleToFPO", mock.Anything, fpoOrgID, 0, 10).
		Return(publishStates, 2, nil)

	mockCatalogRepo.On("GetByID", mock.Anything, "PROD00000001", mock.AnythingOfType("*catalog.CatalogItem")).
		Run(func(args mock.Arguments) {
			item := args.Get(2).(*catalogModels.CatalogItem)
			*item = *activeProduct1
		}).
		Return(activeProduct1, nil)

	mockCatalogRepo.On("GetByID", mock.Anything, "PROD00000002", mock.AnythingOfType("*catalog.CatalogItem")).
		Run(func(args mock.Arguments) {
			item := args.Get(2).(*catalogModels.CatalogItem)
			*item = *activeProduct2
		}).
		Return(activeProduct2, nil)

	service := catalogService.NewPublishService(mockPublishRepo, mockCatalogRepo, logger)

	// Execute
	results, err := service.GetProductsForFPO(context.Background(), fpoOrgID, 0, 10)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 2)

	// Verify pricing calculations
	for _, result := range results {
		assert.NotNil(t, result.Pricing.BasePrice)
		assert.NotNil(t, result.Pricing.DeliveryCost)
		assert.NotNil(t, result.Pricing.CommissionAmount)
		assert.NotNil(t, result.Pricing.RetailPrice)
		assert.NotNil(t, result.Pricing.PriceLockedUntil)
	}

	mockPublishRepo.AssertExpectations(t)
	mockCatalogRepo.AssertExpectations(t)
}
