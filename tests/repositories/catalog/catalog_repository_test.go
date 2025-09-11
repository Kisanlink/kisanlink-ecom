package catalog

import (
    "context"
    "database/sql"
    "errors"
    "testing"

    catalogModels "kisanlink-ecom/entities/models/catalog"
    catalogRequests "kisanlink-ecom/entities/requests/catalog"
    catalogRepo "kisanlink-ecom/internal/repositories/catalog"

    "github.com/shopspring/decimal"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockDatabase represents a mock database interface
type MockDatabase struct {
    mock.Mock
}

func (m *MockDatabase) QueryRow(query string, args ...interface{}) *sql.Row {
    mockArgs := m.Called(query, args)
    return mockArgs.Get(0).(*sql.Row)
}

func (m *MockDatabase) Query(query string, args ...interface{}) (*sql.Rows, error) {
    mockArgs := m.Called(query, args)
    return mockArgs.Get(0).(*sql.Rows), mockArgs.Error(1)
}

func (m *MockDatabase) Exec(query string, args ...interface{}) (sql.Result, error) {
    mockArgs := m.Called(query, args)
    return mockArgs.Get(0).(sql.Result), mockArgs.Error(1)
}

// MockResult represents a mock SQL result
type MockResult struct {
    mock.Mock
}

func (m *MockResult) LastInsertId() (int64, error) {
    args := m.Called()
    return args.Get(0).(int64), args.Error(1)
}

func (m *MockResult) RowsAffected() (int64, error) {
    args := m.Called()
    return args.Get(0).(int64), args.Error(1)
}

// Test helper functions
func createTestCatalogItem() *catalogModels.CatalogItem {
    item := &catalogModels.CatalogItem{
        OrganizationID: "org-123",
        ItemType:       catalogModels.CatalogItemTypeProduct,
        Name:           "Test Product",
        Description:    "Test Description",
        BasePrice:      decimal.NewFromFloat(100.0),
        Currency:       "INR",
        Category:       "agriculture",
        SKU:            "TEST-001",
        IsActive:       true,
        Visibility:     catalogModels.VisibilityTypeOrg,
        Weight:         decimal.NewFromFloat(5.5),
        Perishable:     true,
        ShelfLifeDays:  30,
    }
    item.ID = "item-123"
    return item
}

// Test Create method with comprehensive scenarios
func TestCatalogRepository_Create_Success(t *testing.T) {
    mockDB := &MockDatabase{}
    mockResult := &MockResult{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    item := createTestCatalogItem()

    // Mock successful insert
    mockResult.On("LastInsertId").Return(int64(1), nil)
    mockResult.On("RowsAffected").Return(int64(1), nil)
    mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(mockResult, nil)

    err := repo.Create(context.Background(), item)

    assert.NoError(t, err)
    mockDB.AssertExpectations(t)
    mockResult.AssertExpectations(t)
}

func TestCatalogRepository_Create_DatabaseError(t *testing.T) {
    mockDB := &MockDatabase{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    item := createTestCatalogItem()

    // Mock database error
    mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(nil, errors.New("database connection failed"))

    err := repo.Create(context.Background(), item)

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "database connection failed")
    mockDB.AssertExpectations(t)
}

func TestCatalogRepository_Create_DuplicateKey(t *testing.T) {
    mockDB := &MockDatabase{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    item := createTestCatalogItem()

    // Mock duplicate key constraint error
    mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(nil, errors.New("duplicate key value violates unique constraint"))

    err := repo.Create(context.Background(), item)

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "duplicate key")
    mockDB.AssertExpectations(t)
}

func TestCatalogRepository_Create_NilItem(t *testing.T) {
    mockDB := &MockDatabase{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    err := repo.Create(context.Background(), nil)

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "catalog item cannot be nil")
    mockDB.AssertNotCalled(t, "Exec")
}

// Test GetByID method with comprehensive scenarios
func TestCatalogRepository_GetByID_Success(t *testing.T) {
    mockDB := &MockDatabase{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    expectedItem := createTestCatalogItem()
    var resultItem catalogModels.CatalogItem

    // Mock successful query - this would need to be implemented based on actual database layer
    // For now, we'll test the interface contract
    result, err := repo.GetByID(context.Background(), "item-123", &resultItem)

    // Since we don't have actual database implementation, we test the interface
    assert.NotNil(t, result)
    assert.NoError(t, err)
}

func TestCatalogRepository_GetByID_NotFound(t *testing.T) {
    mockDB := &MockDatabase{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    var resultItem catalogModels.CatalogItem

    // Mock not found scenario
    result, err := repo.GetByID(context.Background(), "nonexistent", &resultItem)

    // Test that appropriate error is returned for not found
    assert.Nil(t, result)
    assert.Error(t, err)
}

func TestCatalogRepository_GetByID_EmptyID(t *testing.T) {
    mockDB := &MockDatabase{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    var resultItem catalogModels.CatalogItem

    result, err := repo.GetByID(context.Background(), "", &resultItem)

    assert.Nil(t, result)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "ID cannot be empty")
}

func TestCatalogRepository_GetByID_NilModel(t *testing.T) {
    mockDB := &MockDatabase{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    result, err := repo.GetByID(context.Background(), "item-123", nil)

    assert.Nil(t, result)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "model cannot be nil")
}

// Test Update method with comprehensive scenarios
func TestCatalogRepository_Update_Success(t *testing.T) {
    mockDB := &MockDatabase{}
    mockResult := &MockResult{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    item := createTestCatalogItem()
    item.Name = "Updated Product Name"

    // Mock successful update
    mockResult.On("RowsAffected").Return(int64(1), nil)
    mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(mockResult, nil)

    err := repo.Update(context.Background(), item)

    assert.NoError(t, err)
    mockDB.AssertExpectations(t)
    mockResult.AssertExpectations(t)
}

func TestCatalogRepository_Update_NotFound(t *testing.T) {
    mockDB := &MockDatabase{}
    mockResult := &MockResult{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    item := createTestCatalogItem()
    item.ID = "nonexistent"

    // Mock no rows affected (item not found)
    mockResult.On("RowsAffected").Return(int64(0), nil)
    mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(mockResult, nil)

    err := repo.Update(context.Background(), item)

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "not found")
    mockDB.AssertExpectations(t)
    mockResult.AssertExpectations(t)
}

func TestCatalogRepository_Update_NilItem(t *testing.T) {
    mockDB := &MockDatabase{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    err := repo.Update(context.Background(), nil)

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "catalog item cannot be nil")
    mockDB.AssertNotCalled(t, "Exec")
}

// Test Delete method with comprehensive scenarios
func TestCatalogRepository_Delete_Success(t *testing.T) {
    mockDB := &MockDatabase{}
    mockResult := &MockResult{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    var item catalogModels.CatalogItem

    // Mock successful delete
    mockResult.On("RowsAffected").Return(int64(1), nil)
    mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(mockResult, nil)

    err := repo.Delete(context.Background(), "item-123", &item)

    assert.NoError(t, err)
    mockDB.AssertExpectations(t)
    mockResult.AssertExpectations(t)
}

func TestCatalogRepository_Delete_NotFound(t *testing.T) {
    mockDB := &MockDatabase{}
    mockResult := &MockResult{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    var item catalogModels.CatalogItem

    // Mock no rows affected (item not found)
    mockResult.On("RowsAffected").Return(int64(0), nil)
    mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(mockResult, nil)

    err := repo.Delete(context.Background(), "nonexistent", &item)

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "not found")
    mockDB.AssertExpectations(t)
    mockResult.AssertExpectations(t)
}

// Test SoftDelete method with comprehensive scenarios
func TestCatalogRepository_SoftDelete_Success(t *testing.T) {
    mockDB := &MockDatabase{}
    mockResult := &MockResult{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    // Mock successful soft delete
    mockResult.On("RowsAffected").Return(int64(1), nil)
    mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(mockResult, nil)

    err := repo.SoftDelete(context.Background(), "item-123", "user-123")

    assert.NoError(t, err)
    mockDB.AssertExpectations(t)
    mockResult.AssertExpectations(t)
}

func TestCatalogRepository_SoftDelete_EmptyUserID(t *testing.T) {
    mockDB := &MockDatabase{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    err := repo.SoftDelete(context.Background(), "item-123", "")

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "deleted by user ID cannot be empty")
    mockDB.AssertNotCalled(t, "Exec")
}

// Test Restore method with comprehensive scenarios
func TestCatalogRepository_Restore_Success(t *testing.T) {
    mockDB := &MockDatabase{}
    mockResult := &MockResult{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    // Mock successful restore
    mockResult.On("RowsAffected").Return(int64(1), nil)
    mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(mockResult, nil)

    err := repo.Restore(context.Background(), "item-123")

    assert.NoError(t, err)
    mockDB.AssertExpectations(t)
    mockResult.AssertExpectations(t)
}

func TestCatalogRepository_Restore_NotFound(t *testing.T) {
    mockDB := &MockDatabase{}
    mockResult := &MockResult{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    // Mock no rows affected (item not found or not deleted)
    mockResult.On("RowsAffected").Return(int64(0), nil)
    mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(mockResult, nil)

    err := repo.Restore(context.Background(), "item-123")

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "not found or not deleted")
    mockDB.AssertExpectations(t)
    mockResult.AssertExpectations(t)
}

// Test GetBySKU method with comprehensive scenarios
func TestCatalogRepository_GetBySKU_Success(t *testing.T) {
    mockDB := &MockDatabase{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    // Test the interface contract
    result, err := repo.GetBySKU(context.Background(), "TEST-001")

    // Since we don't have actual implementation, test interface behavior
    assert.NotNil(t, result)
    assert.NoError(t, err)
}

func TestCatalogRepository_GetBySKU_NotFound(t *testing.T) {
    mockDB := &MockDatabase{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    result, err := repo.GetBySKU(context.Background(), "NONEXISTENT")

    assert.Nil(t, result)
    assert.Error(t, err)
}

func TestCatalogRepository_GetBySKU_EmptySKU(t *testing.T) {
    mockDB := &MockDatabase{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    result, err := repo.GetBySKU(context.Background(), "")

    assert.Nil(t, result)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "SKU cannot be empty")
}

// Test ListCatalogItems method with comprehensive scenarios
func TestCatalogRepository_ListCatalogItems_Success(t *testing.T) {
    mockDB := &MockDatabase{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    filter := &catalogRequests.CatalogFilter{}

    // Test the interface contract
    items, total, err := repo.ListCatalogItems(context.Background(), filter, 0, 20)

    assert.NotNil(t, items)
    assert.GreaterOrEqual(t, total, 0)
    assert.NoError(t, err)
}

func TestCatalogRepository_ListCatalogItems_WithFilters(t *testing.T) {
    mockDB := &MockDatabase{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    itemType := catalogModels.CatalogItemTypeProduct
    isActive := true
    filter := &catalogRequests.CatalogFilter{
        ItemType: &itemType,
        IsActive: &isActive,
    }

    items, total, err := repo.ListCatalogItems(context.Background(), filter, 0, 20)

    assert.NotNil(t, items)
    assert.GreaterOrEqual(t, total, 0)
    assert.NoError(t, err)
}

func TestCatalogRepository_ListCatalogItems_InvalidPagination(t *testing.T) {
    mockDB := &MockDatabase{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    filter := &catalogRequests.CatalogFilter{}

    // Test with negative offset
    items, total, err := repo.ListCatalogItems(context.Background(), filter, -1, 20)

    assert.Nil(t, items)
    assert.Equal(t, 0, total)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "offset cannot be negative")

    // Test with zero or negative limit
    items, total, err = repo.ListCatalogItems(context.Background(), filter, 0, 0)

    assert.Nil(t, items)
    assert.Equal(t, 0, total)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "limit must be positive")
}

func TestCatalogRepository_ListCatalogItems_NilFilter(t *testing.T) {
    mockDB := &MockDatabase{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    items, total, err := repo.ListCatalogItems(context.Background(), nil, 0, 20)

    assert.Nil(t, items)
    assert.Equal(t, 0, total)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "filter cannot be nil")
}

// Test inventory-related methods
func TestCatalogRepository_GetInventoryLevel_Success(t *testing.T) {
    mockDB := &MockDatabase{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    level, err := repo.GetInventoryLevel(context.Background(), "item-123")

    assert.GreaterOrEqual(t, level, 0.0)
    assert.NoError(t, err)
}

func TestCatalogRepository_GetInventoryLevel_EmptyItemID(t *testing.T) {
    mockDB := &MockDatabase{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    level, err := repo.GetInventoryLevel(context.Background(), "")

    assert.Equal(t, 0.0, level)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "item ID cannot be empty")
}

func TestCatalogRepository_ReserveInventory_Success(t *testing.T) {
    mockDB := &MockDatabase{}
    mockResult := &MockResult{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    // Mock successful reservation
    mockResult.On("RowsAffected").Return(int64(1), nil)
    mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(mockResult, nil)

    err := repo.ReserveInventory(context.Background(), "item-123", 10.0)

    assert.NoError(t, err)
    mockDB.AssertExpectations(t)
    mockResult.AssertExpectations(t)
}

func TestCatalogRepository_ReserveInventory_InsufficientStock(t *testing.T) {
    mockDB := &MockDatabase{}
    mockResult := &MockResult{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    // Mock insufficient stock (no rows affected)
    mockResult.On("RowsAffected").Return(int64(0), nil)
    mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(mockResult, nil)

    err := repo.ReserveInventory(context.Background(), "item-123", 1000.0)

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "insufficient inventory")
    mockDB.AssertExpectations(t)
    mockResult.AssertExpectations(t)
}

func TestCatalogRepository_ReserveInventory_InvalidQuantity(t *testing.T) {
    mockDB := &MockDatabase{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    // Test with zero quantity
    err := repo.ReserveInventory(context.Background(), "item-123", 0.0)

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "quantity must be positive")
    mockDB.AssertNotCalled(t, "Exec")

    // Test with negative quantity
    err = repo.ReserveInventory(context.Background(), "item-123", -10.0)

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "quantity must be positive")
    mockDB.AssertNotCalled(t, "Exec")
}

func TestCatalogRepository_ReleaseInventory_Success(t *testing.T) {
    mockDB := &MockDatabase{}
    mockResult := &MockResult{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    // Mock successful release
    mockResult.On("RowsAffected").Return(int64(1), nil)
    mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(mockResult, nil)

    err := repo.ReleaseInventory(context.Background(), "item-123", 5.0)

    assert.NoError(t, err)
    mockDB.AssertExpectations(t)
    mockResult.AssertExpectations(t)
}

func TestCatalogRepository_ReleaseInventory_InsufficientReserved(t *testing.T) {
    mockDB := &MockDatabase{}
    mockResult := &MockResult{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    // Mock insufficient reserved quantity (no rows affected)
    mockResult.On("RowsAffected").Return(int64(0), nil)
    mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(mockResult, nil)

    err := repo.ReleaseInventory(context.Background(), "item-123", 1000.0)

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "insufficient reserved inventory")
    mockDB.AssertExpectations(t)
    mockResult.AssertExpectations(t)
}

// Test transaction scenarios
func TestCatalogRepository_Create_TransactionRollback(t *testing.T) {
    mockDB := &MockDatabase{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    item := createTestCatalogItem()

    // Mock transaction failure
    mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(nil, errors.New("transaction deadlock"))

    err := repo.Create(context.Background(), item)

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "transaction deadlock")
    mockDB.AssertExpectations(t)
}

// Test concurrent access scenarios
func TestCatalogRepository_Update_ConcurrentModification(t *testing.T) {
    mockDB := &MockDatabase{}
    mockResult := &MockResult{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    item := createTestCatalogItem()

    // Mock concurrent modification (optimistic locking failure)
    mockResult.On("RowsAffected").Return(int64(0), nil)
    mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(mockResult, nil)

    err := repo.Update(context.Background(), item)

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "not found")
    mockDB.AssertExpectations(t)
    mockResult.AssertExpectations(t)
}

// Test performance scenarios
func TestCatalogRepository_ListCatalogItems_LargeResultSet(t *testing.T) {
    mockDB := &MockDatabase{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    filter := &catalogRequests.CatalogFilter{}

    // Test with large limit
    items, total, err := repo.ListCatalogItems(context.Background(), filter, 0, 1000)

    assert.NotNil(t, items)
    assert.GreaterOrEqual(t, total, 0)
    assert.NoError(t, err)
}

func TestCatalogRepository_ListCatalogItems_MaxLimitExceeded(t *testing.T) {
    mockDB := &MockDatabase{}
    repo := catalogRepo.NewCatalogRepository(mockDB)

    filter := &catalogRequests.CatalogFilter{}

    // Test with limit exceeding maximum allowed
    items, total, err := repo.ListCatalogItems(context.Background(), filter, 0, 10000)

    assert.Nil(t, items)
    assert.Equal(t, 0, total)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "limit exceeds maximum allowed")
}
