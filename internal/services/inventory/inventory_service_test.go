package inventory

import (
    "context"
    "testing"
    "time"

    catalogModels "kisanlink-ecom/entities/models/catalog"
    inventoryRepo "kisanlink-ecom/internal/repositories/inventory"

    "github.com/shopspring/decimal"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockInventoryRepository is a mock implementation of InventoryRepository
type MockInventoryRepository struct {
    mock.Mock
}

func (m *MockInventoryRepository) Create(ctx context.Context, lot *catalogModels.InventoryLot) error {
    args := m.Called(ctx, lot)
    return args.Error(0)
}

func (m *MockInventoryRepository) GetByID(ctx context.Context, id string) (*catalogModels.InventoryLot, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*catalogModels.InventoryLot), args.Error(1)
}

func (m *MockInventoryRepository) GetByLotNumber(ctx context.Context, orgID, lotNumber string) (*catalogModels.InventoryLot, error) {
    args := m.Called(ctx, orgID, lotNumber)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*catalogModels.InventoryLot), args.Error(1)
}

func (m *MockInventoryRepository) Update(ctx context.Context, lot *catalogModels.InventoryLot) error {
    args := m.Called(ctx, lot)
    return args.Error(0)
}

func (m *MockInventoryRepository) Delete(ctx context.Context, id string) error {
    args := m.Called(ctx, id)
    return args.Error(0)
}

func (m *MockInventoryRepository) ListByOrganization(ctx context.Context, orgID string, offset, limit int) ([]*catalogModels.InventoryLot, int64, error) {
    args := m.Called(ctx, orgID, offset, limit)
    return args.Get(0).([]*catalogModels.InventoryLot), args.Get(1).(int64), args.Error(2)
}

func (m *MockInventoryRepository) ListByCatalogItem(ctx context.Context, catalogItemID string, offset, limit int) ([]*catalogModels.InventoryLot, int64, error) {
    args := m.Called(ctx, catalogItemID, offset, limit)
    return args.Get(0).([]*catalogModels.InventoryLot), args.Get(1).(int64), args.Error(2)
}

func (m *MockInventoryRepository) ListByStatus(ctx context.Context, orgID string, status catalogModels.InventoryStatus, offset, limit int) ([]*catalogModels.InventoryLot, int64, error) {
    args := m.Called(ctx, orgID, status, offset, limit)
    return args.Get(0).([]*catalogModels.InventoryLot), args.Get(1).(int64), args.Error(2)
}

func (m *MockInventoryRepository) GetAvailableQuantity(ctx context.Context, catalogItemID string) (decimal.Decimal, error) {
    args := m.Called(ctx, catalogItemID)
    return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockInventoryRepository) GetTotalQuantity(ctx context.Context, catalogItemID string) (decimal.Decimal, error) {
    args := m.Called(ctx, catalogItemID)
    return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockInventoryRepository) ReserveQuantity(ctx context.Context, catalogItemID string, quantity decimal.Decimal) ([]*catalogModels.InventoryLot, error) {
    args := m.Called(ctx, catalogItemID, quantity)
    return args.Get(0).([]*catalogModels.InventoryLot), args.Error(1)
}

func (m *MockInventoryRepository) ReleaseQuantity(ctx context.Context, catalogItemID string, quantity decimal.Decimal) error {
    args := m.Called(ctx, catalogItemID, quantity)
    return args.Error(0)
}

func (m *MockInventoryRepository) SellQuantity(ctx context.Context, catalogItemID string, quantity decimal.Decimal) error {
    args := m.Called(ctx, catalogItemID, quantity)
    return args.Error(0)
}

func (m *MockInventoryRepository) GetExpiringLots(ctx context.Context, orgID string, beforeDate time.Time) ([]*catalogModels.InventoryLot, error) {
    args := m.Called(ctx, orgID, beforeDate)
    return args.Get(0).([]*catalogModels.InventoryLot), args.Error(1)
}

func (m *MockInventoryRepository) MarkExpired(ctx context.Context, lotID string) error {
    args := m.Called(ctx, lotID)
    return args.Error(0)
}

func (m *MockInventoryRepository) CreateAuditLog(ctx context.Context, log *inventoryRepo.InventoryAuditLog) error {
    args := m.Called(ctx, log)
    return args.Error(0)
}

func (m *MockInventoryRepository) GetAuditLogs(ctx context.Context, lotID string, offset, limit int) ([]*inventoryRepo.InventoryAuditLog, int64, error) {
    args := m.Called(ctx, lotID, offset, limit)
    return args.Get(0).([]*inventoryRepo.InventoryAuditLog), args.Get(1).(int64), args.Error(2)
}

func (m *MockInventoryRepository) GetInventoryLevel(ctx context.Context, itemID string) (decimal.Decimal, error) {
    args := m.Called(ctx, itemID)
    return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockInventoryRepository) ReserveInventory(ctx context.Context, itemID string, quantity decimal.Decimal) error {
    args := m.Called(ctx, itemID, quantity)
    return args.Error(0)
}

func (m *MockInventoryRepository) ReleaseInventory(ctx context.Context, itemID string, quantity decimal.Decimal) error {
    args := m.Called(ctx, itemID, quantity)
    return args.Error(0)
}

// MockCatalogRepository is a mock implementation of CatalogRepository
type MockCatalogRepository struct {
    mock.Mock
}

func (m *MockCatalogRepository) GetByID(ctx context.Context, id string, model interface{}) (interface{}, error) {
    args := m.Called(ctx, id, model)
    return model, args.Error(1)
}

// Test helper functions
func createTestInventoryLot() *catalogModels.InventoryLot {
    return &catalogModels.InventoryLot{
        CatalogItemID:     "item123",
        OrganizationID:    "org123",
        LotNumber:         "LOT001",
        InitialQuantity:   decimal.NewFromFloat(100),
        AvailableQuantity: decimal.NewFromFloat(100),
        ReservedQuantity:  decimal.Zero,
        SoldQuantity:      decimal.Zero,
        Status:            catalogModels.InventoryStatusAvailable,
    }
}

func createTestCreateRequest() *CreateInventoryLotRequest {
    return &CreateInventoryLotRequest{
        CatalogItemID:   "item123",
        LotNumber:       "LOT001",
        InitialQuantity: decimal.NewFromFloat(100),
        QualityGrade:    "A",
    }
}

// Test CreateInventoryLot with product validation and quantity management
func TestInventoryService_CreateInventoryLot_Success(t *testing.T) {
    mockInventoryRepo := &MockInventoryRepository{}
    mockCatalogRepo := &MockCatalogRepository{}

    service := NewInventoryService(mockInventoryRepo, mockCatalogRepo)

    req := createTestCreateRequest()

    // Mock catalog item validation
    mockCatalogRepo.On("GetByID", mock.Anything, "item123", mock.AnythingOfType("*catalog.CatalogItem")).Return(mock.Anything, nil).Run(func(args mock.Arguments) {
        item := args.Get(2).(*catalogModels.CatalogItem)
        *item = catalogModels.CatalogItem{
            OrganizationID: "org123",
            ItemType:       catalogModels.CatalogItemTypeProduct,
            Name:           "Test Product",
            BasePrice:      decimal.NewFromFloat(100.0),
        }
    })

    // Mock lot number uniqueness check
    mockInventoryRepo.On("GetByLotNumber", mock.Anything, "org123", "LOT001").Return(nil, assert.AnError)

    // Mock lot creation
    mockInventoryRepo.On("Create", mock.Anything, mock.AnythingOfType("*catalog.InventoryLot")).Return(nil)

    // Mock audit log creation
    mockInventoryRepo.On("CreateAuditLog", mock.Anything, mock.AnythingOfType("*inventory.InventoryAuditLog")).Return(nil)

    lot, err := service.CreateInventoryLot(context.Background(), req, "user123", "org123")

    assert.NoError(t, err)
    assert.NotNil(t, lot)
    assert.Equal(t, "item123", lot.CatalogItemID)
    assert.Equal(t, "org123", lot.OrganizationID)
    assert.Equal(t, "LOT001", lot.LotNumber)
    assert.Equal(t, decimal.NewFromFloat(100), lot.InitialQuantity)
    assert.Equal(t, decimal.NewFromFloat(100), lot.AvailableQuantity)

    mockInventoryRepo.AssertExpectations(t)
    mockCatalogRepo.AssertExpectations(t)
}

func TestInventoryService_CreateInventoryLot_ProductValidation_NonProduct(t *testing.T) {
    mockInventoryRepo := &MockInventoryRepository{}
    mockCatalogRepo := &MockCatalogRepository{}

    service := NewInventoryService(mockInventoryRepo, mockCatalogRepo)

    req := createTestCreateRequest()

    // Mock catalog item validation - return service instead of product
    mockCatalogRepo.On("GetByID", mock.Anything, "item123", mock.AnythingOfType("*catalog.CatalogItem")).Return(mock.Anything, nil).Run(func(args mock.Arguments) {
        item := args.Get(2).(*catalogModels.CatalogItem)
        *item = catalogModels.CatalogItem{
            OrganizationID: "org123",
            ItemType:       catalogModels.CatalogItemTypeService, // Service, not product
            Name:           "Test Service",
            BasePrice:      decimal.NewFromFloat(100.0),
        }
    })

    lot, err := service.CreateInventoryLot(context.Background(), req, "user123", "org123")

    assert.Error(t, err)
    assert.Nil(t, lot)
    assert.Contains(t, err.Error(), "inventory lots can only be created for products")

    mockCatalogRepo.AssertExpectations(t)
    // Don't assert inventory repo expectations since it shouldn't be called
}

func TestInventoryService_CreateInventoryLot_ExpiryDateValidation(t *testing.T) {
    mockInventoryRepo := &MockInventoryRepository{}
    mockCatalogRepo := &MockCatalogRepository{}

    service := NewInventoryService(mockInventoryRepo, mockCatalogRepo)

    req := createTestCreateRequest()
    pastDate := time.Now().Add(-24 * time.Hour) // Yesterday
    req.ExpiryDate = &pastDate

    // Mock catalog item validation
    mockCatalogRepo.On("GetByID", mock.Anything, "item123", mock.AnythingOfType("*catalog.CatalogItem")).Return(mock.Anything, nil).Run(func(args mock.Arguments) {
        item := args.Get(2).(*catalogModels.CatalogItem)
        *item = catalogModels.CatalogItem{
            OrganizationID: "org123",
            ItemType:       catalogModels.CatalogItemTypeProduct,
            Name:           "Test Product",
            BasePrice:      decimal.NewFromFloat(100.0),
        }
    })

    // Mock lot number uniqueness check
    mockInventoryRepo.On("GetByLotNumber", mock.Anything, "org123", "LOT001").Return(nil, assert.AnError)

    lot, err := service.CreateInventoryLot(context.Background(), req, "user123", "org123")

    assert.Error(t, err)
    assert.Nil(t, lot)
    assert.Contains(t, err.Error(), "expiry date must be in the future")

    mockInventoryRepo.AssertExpectations(t)
    mockCatalogRepo.AssertExpectations(t)
}

// Test ReserveInventory with enhanced audit trail
func TestInventoryService_ReserveInventory_Success_WithAuditTrail(t *testing.T) {
    mockInventoryRepo := &MockInventoryRepository{}
    mockCatalogRepo := &MockCatalogRepository{}

    service := NewInventoryService(mockInventoryRepo, mockCatalogRepo)

    quantity := decimal.NewFromFloat(50)
    affectedLots := []*catalogModels.InventoryLot{createTestInventoryLot()}

    // Mock catalog item validation
    mockCatalogRepo.On("GetByID", mock.Anything, "item123", mock.AnythingOfType("*catalog.CatalogItem")).Return(mock.Anything, nil).Run(func(args mock.Arguments) {
        item := args.Get(2).(*catalogModels.CatalogItem)
        *item = catalogModels.CatalogItem{
            OrganizationID: "org123",
            ItemType:       catalogModels.CatalogItemTypeProduct,
            Name:           "Test Product",
            BasePrice:      decimal.NewFromFloat(100.0),
        }
    })

    // Mock availability check
    mockInventoryRepo.On("GetAvailableQuantity", mock.Anything, "item123").Return(decimal.NewFromFloat(100), nil)

    // Mock getting lots before reservation for audit trail
    mockInventoryRepo.On("ListByCatalogItem", mock.Anything, "item123", 0, 1000).Return(affectedLots, int64(1), nil)

    // Mock reservation
    mockInventoryRepo.On("ReserveQuantity", mock.Anything, "item123", quantity).Return(affectedLots, nil)

    // Mock audit log creation
    mockInventoryRepo.On("CreateAuditLog", mock.Anything, mock.AnythingOfType("*inventory.InventoryAuditLog")).Return(nil)

    lots, err := service.ReserveInventory(context.Background(), "item123", quantity, "user123", "org123")

    assert.NoError(t, err)
    assert.NotNil(t, lots)
    assert.Len(t, lots, 1)

    mockInventoryRepo.AssertExpectations(t)
    mockCatalogRepo.AssertExpectations(t)
}

func TestInventoryService_ReserveInventory_InvalidQuantity(t *testing.T) {
    mockInventoryRepo := &MockInventoryRepository{}
    mockCatalogRepo := &MockCatalogRepository{}

    service := NewInventoryService(mockInventoryRepo, mockCatalogRepo)

    // Test with zero quantity
    lots, err := service.ReserveInventory(context.Background(), "item123", decimal.Zero, "user123", "org123")

    assert.Error(t, err)
    assert.Nil(t, lots)
    assert.Contains(t, err.Error(), "quantity must be greater than zero")

    // Test with negative quantity
    lots, err = service.ReserveInventory(context.Background(), "item123", decimal.NewFromFloat(-10), "user123", "org123")

    assert.Error(t, err)
    assert.Nil(t, lots)
    assert.Contains(t, err.Error(), "quantity must be greater than zero")
}

func TestInventoryService_ReserveInventory_InsufficientInventory(t *testing.T) {
    mockInventoryRepo := &MockInventoryRepository{}
    mockCatalogRepo := &MockCatalogRepository{}

    service := NewInventoryService(mockInventoryRepo, mockCatalogRepo)

    quantity := decimal.NewFromFloat(150) // More than available

    // Mock catalog item validation
    mockCatalogRepo.On("GetByID", mock.Anything, "item123", mock.AnythingOfType("*catalog.CatalogItem")).Return(mock.Anything, nil).Run(func(args mock.Arguments) {
        item := args.Get(2).(*catalogModels.CatalogItem)
        *item = catalogModels.CatalogItem{
            OrganizationID: "org123",
            ItemType:       catalogModels.CatalogItemTypeProduct,
            Name:           "Test Product",
            BasePrice:      decimal.NewFromFloat(100.0),
        }
    })

    // Mock availability check - only 100 available
    mockInventoryRepo.On("GetAvailableQuantity", mock.Anything, "item123").Return(decimal.NewFromFloat(100), nil)

    lots, err := service.ReserveInventory(context.Background(), "item123", quantity, "user123", "org123")

    assert.Error(t, err)
    assert.Nil(t, lots)
    assert.Contains(t, err.Error(), "insufficient inventory")

    mockInventoryRepo.AssertExpectations(t)
    mockCatalogRepo.AssertExpectations(t)
}

// Test ReleaseInventory with enhanced audit trail
func TestInventoryService_ReleaseInventory_Success_WithAuditTrail(t *testing.T) {
    mockInventoryRepo := &MockInventoryRepository{}
    mockCatalogRepo := &MockCatalogRepository{}

    service := NewInventoryService(mockInventoryRepo, mockCatalogRepo)

    quantity := decimal.NewFromFloat(30)
    lot1 := &catalogModels.InventoryLot{
        CatalogItemID:    "item123",
        ReservedQuantity: decimal.NewFromFloat(50),
    }
    lot1.ID = "lot1"
    lotsWithReserved := []*catalogModels.InventoryLot{lot1}

    // Mock catalog item validation
    mockCatalogRepo.On("GetByID", mock.Anything, "item123", mock.AnythingOfType("*catalog.CatalogItem")).Return(mock.Anything, nil)

    // Mock getting lots with reserved quantity
    mockInventoryRepo.On("ListByCatalogItem", mock.Anything, "item123", 0, 1000).Return(lotsWithReserved, int64(1), nil).Twice()

    // Mock release operation
    mockInventoryRepo.On("ReleaseQuantity", mock.Anything, "item123", quantity).Return(nil)

    // Mock audit log creation
    mockInventoryRepo.On("CreateAuditLog", mock.Anything, mock.AnythingOfType("*inventory.InventoryAuditLog")).Return(nil)

    err := service.ReleaseInventory(context.Background(), "item123", quantity, "user123", "org123")

    assert.NoError(t, err)

    mockInventoryRepo.AssertExpectations(t)
    mockCatalogRepo.AssertExpectations(t)
}

func TestInventoryService_ReleaseInventory_InsufficientReserved(t *testing.T) {
    mockInventoryRepo := &MockInventoryRepository{}
    mockCatalogRepo := &MockCatalogRepository{}

    service := NewInventoryService(mockInventoryRepo, mockCatalogRepo)

    quantity := decimal.NewFromFloat(100) // More than reserved
    lot1 := &catalogModels.InventoryLot{
        CatalogItemID:    "item123",
        ReservedQuantity: decimal.NewFromFloat(50), // Only 50 reserved
    }
    lot1.ID = "lot1"
    lotsWithReserved := []*catalogModels.InventoryLot{lot1}

    // Mock catalog item validation
    mockCatalogRepo.On("GetByID", mock.Anything, "item123", mock.AnythingOfType("*catalog.CatalogItem")).Return(mock.Anything, nil)

    // Mock getting lots with reserved quantity
    mockInventoryRepo.On("ListByCatalogItem", mock.Anything, "item123", 0, 1000).Return(lotsWithReserved, int64(1), nil)

    err := service.ReleaseInventory(context.Background(), "item123", quantity, "user123", "org123")

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "insufficient reserved inventory")

    mockInventoryRepo.AssertExpectations(t)
    mockCatalogRepo.AssertExpectations(t)
}

// Test SellInventory with enhanced audit trail
func TestInventoryService_SellInventory_Success_WithAuditTrail(t *testing.T) {
    mockInventoryRepo := &MockInventoryRepository{}
    mockCatalogRepo := &MockCatalogRepository{}

    service := NewInventoryService(mockInventoryRepo, mockCatalogRepo)

    quantity := decimal.NewFromFloat(30)
    lot1 := &catalogModels.InventoryLot{
        CatalogItemID:    "item123",
        ReservedQuantity: decimal.NewFromFloat(50),
        SoldQuantity:     decimal.NewFromFloat(10),
    }
    lot1.ID = "lot1"
    lotsWithReserved := []*catalogModels.InventoryLot{lot1}

    // Mock catalog item validation
    mockCatalogRepo.On("GetByID", mock.Anything, "item123", mock.AnythingOfType("*catalog.CatalogItem")).Return(mock.Anything, nil)

    // Mock getting lots with reserved quantity
    mockInventoryRepo.On("ListByCatalogItem", mock.Anything, "item123", 0, 1000).Return(lotsWithReserved, int64(1), nil).Twice()

    // Mock sell operation
    mockInventoryRepo.On("SellQuantity", mock.Anything, "item123", quantity).Return(nil)

    // Mock audit log creation
    mockInventoryRepo.On("CreateAuditLog", mock.Anything, mock.AnythingOfType("*inventory.InventoryAuditLog")).Return(nil)

    err := service.SellInventory(context.Background(), "item123", quantity, "user123", "org123")

    assert.NoError(t, err)

    mockInventoryRepo.AssertExpectations(t)
    mockCatalogRepo.AssertExpectations(t)
}

// Test AdjustInventory with quantity validation
func TestInventoryService_AdjustInventory_PositiveAdjustment(t *testing.T) {
    mockInventoryRepo := &MockInventoryRepository{}
    mockCatalogRepo := &MockCatalogRepository{}

    service := NewInventoryService(mockInventoryRepo, mockCatalogRepo)

    lot := createTestInventoryLot()
    adjustment := decimal.NewFromFloat(25)

    // Mock lot retrieval
    mockInventoryRepo.On("GetByID", mock.Anything, "lot123").Return(lot, nil)

    // Mock lot update
    mockInventoryRepo.On("Update", mock.Anything, mock.AnythingOfType("*catalog.InventoryLot")).Return(nil)

    // Mock audit log creation
    mockInventoryRepo.On("CreateAuditLog", mock.Anything, mock.AnythingOfType("*inventory.InventoryAuditLog")).Return(nil)

    updatedLot, err := service.AdjustInventory(context.Background(), "lot123", adjustment, "Stock adjustment", "user123", "org123")

    assert.NoError(t, err)
    assert.NotNil(t, updatedLot)
    assert.Equal(t, decimal.NewFromFloat(125), updatedLot.AvailableQuantity) // 100 + 25
    assert.Equal(t, decimal.NewFromFloat(125), updatedLot.InitialQuantity)   // 100 + 25

    mockInventoryRepo.AssertExpectations(t)
}

func TestInventoryService_AdjustInventory_NegativeAdjustment_Valid(t *testing.T) {
    mockInventoryRepo := &MockInventoryRepository{}
    mockCatalogRepo := &MockCatalogRepository{}

    service := NewInventoryService(mockInventoryRepo, mockCatalogRepo)

    lot := createTestInventoryLot()
    adjustment := decimal.NewFromFloat(-25) // Reduce by 25

    // Mock lot retrieval
    mockInventoryRepo.On("GetByID", mock.Anything, "lot123").Return(lot, nil)

    // Mock lot update
    mockInventoryRepo.On("Update", mock.Anything, mock.AnythingOfType("*catalog.InventoryLot")).Return(nil)

    // Mock audit log creation
    mockInventoryRepo.On("CreateAuditLog", mock.Anything, mock.AnythingOfType("*inventory.InventoryAuditLog")).Return(nil)

    updatedLot, err := service.AdjustInventory(context.Background(), "lot123", adjustment, "Stock adjustment", "user123", "org123")

    assert.NoError(t, err)
    assert.NotNil(t, updatedLot)
    assert.Equal(t, decimal.NewFromFloat(75), updatedLot.AvailableQuantity) // 100 - 25
    assert.Equal(t, decimal.NewFromFloat(75), updatedLot.InitialQuantity)   // 100 - 25

    mockInventoryRepo.AssertExpectations(t)
}

func TestInventoryService_AdjustInventory_NegativeResult_Invalid(t *testing.T) {
    mockInventoryRepo := &MockInventoryRepository{}
    mockCatalogRepo := &MockCatalogRepository{}

    service := NewInventoryService(mockInventoryRepo, mockCatalogRepo)

    lot := createTestInventoryLot()
    adjustment := decimal.NewFromFloat(-150) // Would result in negative quantity

    // Mock lot retrieval
    mockInventoryRepo.On("GetByID", mock.Anything, "lot123").Return(lot, nil)

    updatedLot, err := service.AdjustInventory(context.Background(), "lot123", adjustment, "Stock adjustment", "user123", "org123")

    assert.Error(t, err)
    assert.Nil(t, updatedLot)
    assert.Contains(t, err.Error(), "negative available quantity")

    mockInventoryRepo.AssertExpectations(t)
}

// Test CheckAvailability with real-time calculations
func TestInventoryService_CheckAvailability_RealTimeCalculation(t *testing.T) {
    mockInventoryRepo := &MockInventoryRepository{}
    mockCatalogRepo := &MockCatalogRepository{}

    service := NewInventoryService(mockInventoryRepo, mockCatalogRepo)

    requiredQuantity := decimal.NewFromFloat(50)
    availableQuantity := decimal.NewFromFloat(100)

    // Mock catalog item validation
    mockCatalogRepo.On("GetByID", mock.Anything, "item123", mock.AnythingOfType("*catalog.CatalogItem")).Return(mock.Anything, nil).Run(func(args mock.Arguments) {
        item := args.Get(2).(*catalogModels.CatalogItem)
        *item = catalogModels.CatalogItem{
            OrganizationID: "org123",
            ItemType:       catalogModels.CatalogItemTypeProduct,
            Name:           "Test Product",
            BasePrice:      decimal.NewFromFloat(100.0),
        }
    })

    // Mock real-time quantity retrieval
    mockInventoryRepo.On("GetAvailableQuantity", mock.Anything, "item123").Return(availableQuantity, nil)

    isAvailable, available, err := service.CheckAvailability(context.Background(), "item123", requiredQuantity, "user123", "org123")

    assert.NoError(t, err)
    assert.True(t, isAvailable)
    assert.Equal(t, availableQuantity, available)

    mockInventoryRepo.AssertExpectations(t)
    mockCatalogRepo.AssertExpectations(t)
}

// Test ProcessExpiringLots with automated status updates
func TestInventoryService_ProcessExpiringLots_AutomatedStatusUpdate(t *testing.T) {
    mockInventoryRepo := &MockInventoryRepository{}
    mockCatalogRepo := &MockCatalogRepository{}

    service := NewInventoryService(mockInventoryRepo, mockCatalogRepo)

    lot1 := &catalogModels.InventoryLot{
        Status:         catalogModels.InventoryStatusAvailable,
        ExpiryDate:     &time.Time{}, // Expired
        OrganizationID: "org123",
    }
    lot1.ID = "lot1"
    expiringLots := []*catalogModels.InventoryLot{lot1}

    // Mock expiring lots retrieval
    mockInventoryRepo.On("GetExpiringLots", mock.Anything, "org123", mock.AnythingOfType("time.Time")).Return(expiringLots, nil)

    // Mock marking as expired
    mockInventoryRepo.On("MarkExpired", mock.Anything, "lot1").Return(nil)

    // Mock audit log creation
    mockInventoryRepo.On("CreateAuditLog", mock.Anything, mock.AnythingOfType("*inventory.InventoryAuditLog")).Return(nil)

    processedLots, err := service.ProcessExpiringLots(context.Background(), "org123")

    assert.NoError(t, err)
    assert.NotNil(t, processedLots)
    assert.Len(t, processedLots, 1)
    assert.Equal(t, catalogModels.InventoryStatusExpired, processedLots[0].Status)

    mockInventoryRepo.AssertExpectations(t)
}

// Test GetAuditTrail
func TestInventoryService_GetAuditTrail_Success(t *testing.T) {
    mockInventoryRepo := &MockInventoryRepository{}
    mockCatalogRepo := &MockCatalogRepository{}

    service := NewInventoryService(mockInventoryRepo, mockCatalogRepo)

    lot := createTestInventoryLot()
    auditLogs := []*inventoryRepo.InventoryAuditLog{
        {
            LotID:     "lot123",
            Operation: "create",
        },
    }

    // Mock lot retrieval for access check
    mockInventoryRepo.On("GetByID", mock.Anything, "lot123").Return(lot, nil)

    // Mock audit logs retrieval
    mockInventoryRepo.On("GetAuditLogs", mock.Anything, "lot123", 0, 20).Return(auditLogs, int64(1), nil)

    response, err := service.GetAuditTrail(context.Background(), "lot123", 0, 20, "user123", "org123")

    assert.NoError(t, err)
    assert.NotNil(t, response)
    assert.Len(t, response.Logs, 1)
    assert.Equal(t, int64(1), response.Total)
    assert.Equal(t, "create", response.Logs[0].Operation)

    mockInventoryRepo.AssertExpectations(t)
}
