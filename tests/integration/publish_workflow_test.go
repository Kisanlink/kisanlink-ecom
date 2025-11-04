package integration

import (
	"context"
	"testing"
	"time"

	catalogModels "kisanlink-ecom/entities/models/catalog"
	catalogRepo "kisanlink-ecom/internal/repositories/catalog"
	catalogService "kisanlink-ecom/internal/services/catalog"
	"kisanlink-ecom/tests/data"
	"kisanlink-ecom/tests/mocks"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestPublishWorkflow_EndToEnd tests the complete publishing workflow end-to-end
// This test validates the workflow logic by testing the service layer interactions
func TestPublishWorkflow_EndToEnd(t *testing.T) {
	// This integration test validates the complete product publishing workflow logic:
	// 1. Publish product to FPOs
	// 2. Verify FPO access
	// 3. Update delivery costs
	// 4. Revoke access from one FPO
	// 5. Verify revoked FPO no longer has access

	// Setup
	mockDBManager := mocks.NewMockDBManager()
	mockPublishStateRepo := new(MockPublishStateRepository)
	logger := logrus.New()

	catalogRepository := catalogRepo.NewCatalogRepository(mockDBManager)
	publishService := catalogService.NewPublishService(mockPublishStateRepo, catalogRepository, logger)

	ctx := context.Background()
	const (
		productID = "PROD00000001"
		fpoOrgID1 = "ORGN00000002"
		fpoOrgID2 = "ORGN00000003"
	)
	adminUser := "admin-user-123"

	// Step 1: Verify product is active before publishing
	t.Run("Step 1: Verify product is active", func(t *testing.T) {
		activeProduct := data.CreateTestActiveProduct()
		activeProduct.BaseModel.ID = productID
		assert.True(t, activeProduct.IsActive, "Product should be active for publishing")
	})

	// Step 2: Publish to FPOs
	t.Run("Step 2: Publish product to FPOs", func(t *testing.T) {
		activeProduct := data.CreateTestActiveProduct()
		activeProduct.BaseModel.ID = productID

		// Mock product retrieval for validation
		mockDBManager.On("GetByID", mock.Anything, productID, mock.AnythingOfType("*catalog.CatalogItem")).
			Run(func(args mock.Arguments) {
				item := args.Get(2).(*catalogModels.CatalogItem)
				*item = *activeProduct
			}).
			Return(nil).Once()

		// Mock publish state doesn't exist yet
		mockPublishStateRepo.On("GetByProductID", mock.Anything, productID).
			Return(nil, assert.AnError).Once()

		// Mock create publish state
		publishState := data.CreateTestPublishStateWithFPOs(
			productID,
			[]string{fpoOrgID1, fpoOrgID2},
			map[string]decimal.Decimal{
				fpoOrgID1: decimal.NewFromFloat(50.00),
				fpoOrgID2: decimal.NewFromFloat(75.00),
			},
		)
		mockPublishStateRepo.On("Create", mock.Anything, mock.AnythingOfType("*catalog.PublishState")).
			Run(func(args mock.Arguments) {
				state := args.Get(1).(*catalogModels.PublishState)
				*state = *publishState
			}).
			Return(nil).Once()

		// Publish
		req := catalogService.PublishProductRequest{
			ProductID: productID,
			FPOIDs:    []string{fpoOrgID1, fpoOrgID2},
			DeliveryCosts: map[string]decimal.Decimal{
				fpoOrgID1: decimal.NewFromFloat(50.00),
				fpoOrgID2: decimal.NewFromFloat(75.00),
			},
			PlatformFeePercent: decimal.NewFromFloat(10.00),
			PublishedBy:        adminUser,
		}

		result, err := publishService.PublishProduct(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, productID, result.ProductID)
		assert.Len(t, result.FPOAccessList, 2)
		assert.True(t, result.HasFPOAccess(fpoOrgID1))
		assert.True(t, result.HasFPOAccess(fpoOrgID2))
	})

	// Step 3: FPO lists products (should see with pricing)
	t.Run("Step 3: FPO lists products with pricing", func(t *testing.T) {
		publishState := data.CreateTestPublishStateWithFPOs(
			productID,
			[]string{fpoOrgID1, fpoOrgID2},
			map[string]decimal.Decimal{
				fpoOrgID1: decimal.NewFromFloat(50.00),
				fpoOrgID2: decimal.NewFromFloat(75.00),
			},
		)

		activeProduct := data.CreateTestActiveProduct()
		activeProduct.BaseModel.ID = productID

		// Mock getting products visible to FPO
		mockPublishStateRepo.On("GetProductsVisibleToFPO", mock.Anything, fpoOrgID1, 0, 10).
			Return([]*catalogModels.PublishState{publishState}, 1, nil).Once()

		// Mock product retrieval for enrichment
		mockDBManager.On("GetByID", mock.Anything, productID, mock.AnythingOfType("*catalog.CatalogItem")).
			Run(func(args mock.Arguments) {
				item := args.Get(2).(*catalogModels.CatalogItem)
				*item = *activeProduct
			}).
			Return(nil).Once()

		// Get products for FPO
		products, err := publishService.GetProductsForFPO(ctx, fpoOrgID1, 0, 10)
		assert.NoError(t, err)
		assert.Len(t, products, 1)

		// Validate pricing
		product := products[0]
		assert.Equal(t, productID, product.Product.ID)
		assert.NotNil(t, product.Pricing)

		// Verify price calculations
		expectedDeliveryCost := decimal.NewFromFloat(50.00)
		expectedCommission := decimal.NewFromFloat(100.00)   // 10% of 1000
		expectedRetailPrice := decimal.NewFromFloat(1150.00) // 1000 + 50 + 100

		assert.True(t, product.Pricing.DeliveryCost.Equal(expectedDeliveryCost),
			"Expected delivery cost %s, got %s", expectedDeliveryCost, product.Pricing.DeliveryCost)
		assert.True(t, product.Pricing.CommissionAmount.Equal(expectedCommission),
			"Expected commission %s, got %s", expectedCommission, product.Pricing.CommissionAmount)
		assert.True(t, product.Pricing.RetailPrice.Equal(expectedRetailPrice),
			"Expected retail price %s, got %s", expectedRetailPrice, product.Pricing.RetailPrice)
		assert.NotNil(t, product.Pricing.PriceLockedUntil)
	})

	// Step 4: Update delivery costs
	t.Run("Step 4: Update delivery costs", func(t *testing.T) {
		publishState := data.CreateTestPublishStateWithFPOs(
			productID,
			[]string{fpoOrgID1, fpoOrgID2},
			map[string]decimal.Decimal{
				fpoOrgID1: decimal.NewFromFloat(50.00),
				fpoOrgID2: decimal.NewFromFloat(75.00),
			},
		)

		// Mock getting existing publish state
		mockPublishStateRepo.On("GetByProductID", mock.Anything, productID).
			Return(publishState, nil).Once()

		// Mock update
		mockPublishStateRepo.On("Update", mock.Anything, mock.AnythingOfType("*catalog.PublishState")).
			Return(nil).Once()

		// Update delivery costs
		newCosts := map[string]decimal.Decimal{
			fpoOrgID1: decimal.NewFromFloat(60.00),
			fpoOrgID2: decimal.NewFromFloat(85.00),
		}

		err := publishService.UpdateDeliveryCosts(ctx, productID, newCosts)
		assert.NoError(t, err)

		// Verify the state was updated
		assert.True(t, publishState.GetDeliveryCost(fpoOrgID1).Equal(decimal.NewFromFloat(60.00)))
		assert.True(t, publishState.GetDeliveryCost(fpoOrgID2).Equal(decimal.NewFromFloat(85.00)))
	})

	// Step 5: Revoke access for one FPO
	t.Run("Step 5: Revoke FPO access", func(t *testing.T) {
		publishState := data.CreateTestPublishStateWithFPOs(
			productID,
			[]string{fpoOrgID1, fpoOrgID2},
			map[string]decimal.Decimal{
				fpoOrgID1: decimal.NewFromFloat(60.00),
				fpoOrgID2: decimal.NewFromFloat(85.00),
			},
		)

		// Mock getting existing publish state
		mockPublishStateRepo.On("GetByProductID", mock.Anything, productID).
			Return(publishState, nil).Once()

		// Mock removing FPO access
		mockPublishStateRepo.On("RemoveFPOAccess", mock.Anything, productID, fpoOrgID1).
			Run(func(_ mock.Arguments) {
				publishState.RemoveFPOAccess(fpoOrgID1)
			}).
			Return(nil).Once()

		// Revoke access
		err := publishService.RevokeAccess(ctx, productID, fpoOrgID1)
		assert.NoError(t, err)

		// Verify access was revoked
		assert.False(t, publishState.HasFPOAccess(fpoOrgID1))
		assert.True(t, publishState.HasFPOAccess(fpoOrgID2)) // Still has access
	})

	// Step 6: FPO lists again (should not see revoked product)
	t.Run("Step 6: Revoked FPO should not see product", func(t *testing.T) {
		publishState := data.CreateTestPublishStateWithFPOs(
			productID,
			[]string{fpoOrgID2}, // Only fpoOrgID2 now
			map[string]decimal.Decimal{
				fpoOrgID2: decimal.NewFromFloat(85.00),
			},
		)

		// Mock getting products visible to FPO1 (revoked)
		mockPublishStateRepo.On("GetProductsVisibleToFPO", mock.Anything, fpoOrgID1, 0, 10).
			Return([]*catalogModels.PublishState{}, 0, nil).Once()

		// FPO1 lists products (should be empty)
		products, err := publishService.GetProductsForFPO(ctx, fpoOrgID1, 0, 10)
		assert.NoError(t, err)
		assert.Empty(t, products) // Should not see any products

		// Mock getting products visible to FPO2 (still has access)
		mockPublishStateRepo.On("GetProductsVisibleToFPO", mock.Anything, fpoOrgID2, 0, 10).
			Return([]*catalogModels.PublishState{publishState}, 1, nil).Once()

		activeProduct := data.CreateTestActiveProduct()
		activeProduct.BaseModel.ID = productID

		mockDBManager.On("GetByID", mock.Anything, productID, mock.AnythingOfType("*catalog.CatalogItem")).
			Run(func(args mock.Arguments) {
				item := args.Get(2).(*catalogModels.CatalogItem)
				*item = *activeProduct
			}).
			Return(nil).Once()

		// FPO2 lists products (should still see it)
		products, err = publishService.GetProductsForFPO(ctx, fpoOrgID2, 0, 10)
		assert.NoError(t, err)
		assert.Len(t, products, 1) // Should see the product
		assert.Equal(t, productID, products[0].Product.ID)
	})

	// Verify all expectations
	mockDBManager.AssertExpectations(t)
	mockPublishStateRepo.AssertExpectations(t)
}

// TestPublishWorkflow_PricingCalculations tests pricing calculations across workflow
func TestPublishWorkflow_PricingCalculations(t *testing.T) {
	// Setup
	mockDBManager := mocks.NewMockDBManager()
	mockPublishStateRepo := new(MockPublishStateRepository)
	logger := logrus.New()

	catalogRepository := catalogRepo.NewCatalogRepository(mockDBManager)
	publishService := catalogService.NewPublishService(mockPublishStateRepo, catalogRepository, logger)

	tests := []struct {
		name               string
		basePrice          decimal.Decimal
		deliveryCost       decimal.Decimal
		platformFeePercent decimal.Decimal
		expectedRetail     decimal.Decimal
	}{
		{
			name:               "standard pricing",
			basePrice:          decimal.NewFromFloat(1000.00),
			deliveryCost:       decimal.NewFromFloat(50.00),
			platformFeePercent: decimal.NewFromFloat(10.00),
			expectedRetail:     decimal.NewFromFloat(1150.00),
		},
		{
			name:               "higher delivery cost",
			basePrice:          decimal.NewFromFloat(1000.00),
			deliveryCost:       decimal.NewFromFloat(200.00),
			platformFeePercent: decimal.NewFromFloat(10.00),
			expectedRetail:     decimal.NewFromFloat(1300.00),
		},
		{
			name:               "higher platform fee",
			basePrice:          decimal.NewFromFloat(1000.00),
			deliveryCost:       decimal.NewFromFloat(50.00),
			platformFeePercent: decimal.NewFromFloat(20.00),
			expectedRetail:     decimal.NewFromFloat(1250.00),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retailPrice := publishService.CalculateRetailPrice(tt.basePrice, tt.deliveryCost, tt.platformFeePercent)
			assert.True(t, tt.expectedRetail.Equal(retailPrice),
				"Expected %s but got %s", tt.expectedRetail.String(), retailPrice.String())
		})
	}
}

// TestPublishWorkflow_AccessValidation tests FPO access validation
func TestPublishWorkflow_AccessValidation(t *testing.T) {
	// Setup
	mockDBManager := mocks.NewMockDBManager()
	mockPublishStateRepo := new(MockPublishStateRepository)
	logger := logrus.New()

	catalogRepository := catalogRepo.NewCatalogRepository(mockDBManager)
	publishService := catalogService.NewPublishService(mockPublishStateRepo, catalogRepository, logger)

	ctx := context.Background()
	productID := "PROD00000001"
	fpoOrgID := "ORGN00000002"

	t.Run("FPO has valid access", func(t *testing.T) {
		mockPublishStateRepo.On("HasFPOAccess", mock.Anything, productID, fpoOrgID).
			Return(true, nil).Once()

		hasAccess, err := publishService.ValidateFPOAccess(ctx, fpoOrgID, productID)
		assert.NoError(t, err)
		assert.True(t, hasAccess)
	})

	t.Run("FPO does not have access", func(t *testing.T) {
		mockPublishStateRepo.On("HasFPOAccess", mock.Anything, productID, "ORGN00000099").
			Return(false, nil).Once()

		hasAccess, err := publishService.ValidateFPOAccess(ctx, "ORGN00000099", productID)
		assert.NoError(t, err)
		assert.False(t, hasAccess)
	})

	mockPublishStateRepo.AssertExpectations(t)
}

// TestPublishWorkflow_PriceLocking tests price locking mechanism
func TestPublishWorkflow_PriceLocking(t *testing.T) {
	// Setup
	publishState := data.CreateTestPublishState()
	fpoOrgID := "ORGN00000002"
	basePrice := decimal.NewFromFloat(1000.00)

	t.Run("Price lock is set with duration", func(t *testing.T) {
		before := time.Now()
		pricing := publishState.GetPricingForFPO(basePrice, fpoOrgID, 30)
		after := time.Now().Add(31 * time.Minute)

		assert.NotNil(t, pricing.PriceLockedUntil)
		assert.True(t, pricing.PriceLockedUntil.After(before))
		assert.True(t, pricing.PriceLockedUntil.Before(after))
	})

	t.Run("No price lock when duration is zero", func(t *testing.T) {
		pricing := publishState.GetPricingForFPO(basePrice, fpoOrgID, 0)
		assert.Nil(t, pricing.PriceLockedUntil)
	})
}

// MockPublishStateRepository is a mock for testing
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
