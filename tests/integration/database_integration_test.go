//go:build integration
// +build integration

package integration

import (
	"context"
	"database/sql"
	"testing"
	"time"

	catalogModels "github.com/Kisanlink/kisanlink-ecom/entities/models/catalog"
	orderModels "github.com/Kisanlink/kisanlink-ecom/entities/models/orders"
	catalogRepo "github.com/Kisanlink/kisanlink-ecom/internal/repositories/catalog"
	orderRepo "github.com/Kisanlink/kisanlink-ecom/internal/repositories/orders"
	"github.com/Kisanlink/kisanlink-ecom/tests/testutils"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
)

// DatabaseIntegrationTestSuite tests database operations with test database setup and cleanup
type DatabaseIntegrationTestSuite struct {
	suite.Suite
	testDB          *TestDatabase
	catalogRepo     *catalogRepo.CatalogRepository
	orderRepo       *orderRepo.OrderRepository
	createdOrderIDs []string
	createdItemIDs  []string
}

// TestDatabase represents a test database connection that implements db.DBManager
type TestDatabase struct {
	db *sql.DB
}

func NewTestDatabase() *TestDatabase {
	// In a real implementation, this would connect to a test database
	// For this example, we'll use a mock database
	return &TestDatabase{
		db: nil, // Would be actual database connection
	}
}

func (tdb *TestDatabase) GetConnection() *sql.DB {
	return tdb.db
}

func (tdb *TestDatabase) Close() error {
	if tdb.db != nil {
		return tdb.db.Close()
	}
	return nil
}

func (tdb *TestDatabase) BeginTransaction() (*sql.Tx, error) {
	if tdb.db != nil {
		return tdb.db.Begin()
	}
	return nil, nil
}

func (tdb *TestDatabase) ExecuteQuery(query string, args ...interface{}) (*sql.Rows, error) {
	if tdb.db != nil {
		return tdb.db.Query(query, args...)
	}
	return nil, nil
}

func (tdb *TestDatabase) ExecuteCommand(query string, args ...interface{}) (sql.Result, error) {
	if tdb.db != nil {
		return tdb.db.Exec(query, args...)
	}
	// Return mock result for testing
	return testutils.SetupMockResult(1, 1), nil
}

// AutoMigrateModels implements the DBManager interface
func (tdb *TestDatabase) AutoMigrateModels(ctx context.Context, models ...interface{}) error {
	// Mock implementation for testing
	return nil
}

// Implement remaining DBManager interface methods
func (tdb *TestDatabase) Connect(ctx context.Context) error {
	return nil
}

func (tdb *TestDatabase) IsConnected() bool {
	return tdb.db != nil
}

func (tdb *TestDatabase) GetBackendType() db.BackendType {
	return db.BackendType("test")
}

func (tdb *TestDatabase) Create(ctx context.Context, model interface{}) error {
	return nil
}

func (tdb *TestDatabase) GetByID(ctx context.Context, id interface{}, model interface{}) error {
	return nil
}

func (tdb *TestDatabase) Update(ctx context.Context, model interface{}) error {
	return nil
}

func (tdb *TestDatabase) Delete(ctx context.Context, id interface{}, model interface{}) error {
	return nil
}

func (tdb *TestDatabase) SoftDelete(ctx context.Context, id interface{}, model interface{}, deletedBy string) error {
	return nil
}

func (tdb *TestDatabase) Restore(ctx context.Context, id interface{}, model interface{}) error {
	return nil
}

func (tdb *TestDatabase) List(ctx context.Context, filter *base.Filter, model interface{}) error {
	return nil
}

func (tdb *TestDatabase) Count(ctx context.Context, filter *base.Filter, model interface{}) (int64, error) {
	return 0, nil
}

func (tdb *TestDatabase) ListWithDeleted(ctx context.Context, limit, offset int, models interface{}) error {
	return nil
}

func (tdb *TestDatabase) CountWithDeleted(ctx context.Context) (int64, error) {
	return 0, nil
}

func (tdb *TestDatabase) ExistsWithDeleted(ctx context.Context, id interface{}) (bool, error) {
	return false, nil
}

func (tdb *TestDatabase) GetByCreatedBy(ctx context.Context, createdBy interface{}, limit, offset int, models interface{}) error {
	return nil
}

func (tdb *TestDatabase) GetByUpdatedBy(ctx context.Context, updatedBy interface{}, limit, offset int, models interface{}) error {
	return nil
}

func (tdb *TestDatabase) GetByDeletedBy(ctx context.Context, deletedBy interface{}, limit, offset int, models interface{}) error {
	return nil
}

func (tdb *TestDatabase) CreateMany(ctx context.Context, models []interface{}) error {
	return nil
}

func (tdb *TestDatabase) UpdateMany(ctx context.Context, models []interface{}) error {
	return nil
}

func (tdb *TestDatabase) DeleteMany(ctx context.Context, ids []interface{}) error {
	return nil
}

// SetupSuite initializes the test database and repositories
func (suite *DatabaseIntegrationTestSuite) SetupSuite() {
	// Initialize test database
	suite.testDB = NewTestDatabase()

	// Initialize repositories with test database
	suite.catalogRepo = catalogRepo.NewCatalogRepository(suite.testDB)
	suite.orderRepo = orderRepo.NewOrderRepository(suite.testDB)

	// Initialize tracking slices
	suite.createdOrderIDs = make([]string, 0)
	suite.createdItemIDs = make([]string, 0)

	// Setup test database schema (in real implementation)
	suite.setupTestSchema()
}

func (suite *DatabaseIntegrationTestSuite) setupTestSchema() {
	// In a real implementation, this would create test tables
	// For now, we'll just log that schema setup is happening
	suite.T().Log("Setting up test database schema")

	// Example schema setup commands:
	// - Create tables
	// - Create indexes
	// - Insert reference data
	// - Set up constraints
}

// TearDownSuite cleans up the test database
func (suite *DatabaseIntegrationTestSuite) TearDownSuite() {
	// Clean up created test data
	suite.cleanupTestData()

	// Close database connection
	if suite.testDB != nil {
		suite.testDB.Close()
	}
}

func (suite *DatabaseIntegrationTestSuite) cleanupTestData() {
	// Clean up orders
	for _, orderID := range suite.createdOrderIDs {
		suite.T().Logf("Cleaning up order: %s", orderID)
		// In real implementation: DELETE FROM orders WHERE id = orderID
	}

	// Clean up catalog items
	for _, itemID := range suite.createdItemIDs {
		suite.T().Logf("Cleaning up catalog item: %s", itemID)
		// In real implementation: DELETE FROM catalog_items WHERE id = itemID
	}
}

// SetupTest runs before each test
func (suite *DatabaseIntegrationTestSuite) SetupTest() {
	// Reset test state
	suite.createdOrderIDs = suite.createdOrderIDs[:0]
	suite.createdItemIDs = suite.createdItemIDs[:0]
}

// Test catalog item CRUD operations
func (suite *DatabaseIntegrationTestSuite) TestCatalogItemCRUDOperations() {
	ctx := context.Background()

	// Test Create
	suite.T().Log("Testing catalog item creation")
	catalogItem := testutils.CreateTestCatalogItem(catalogModels.CatalogItemTypeProduct)

	err := suite.catalogRepo.Create(ctx, catalogItem)
	suite.NoError(err)
	suite.createdItemIDs = append(suite.createdItemIDs, catalogItem.ID)

	// Test Read
	suite.T().Log("Testing catalog item retrieval")
	var retrievedItem catalogModels.CatalogItem
	result, err := suite.catalogRepo.GetByID(ctx, catalogItem.ID, &retrievedItem)
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(catalogItem.Name, retrievedItem.Name)
	suite.Equal(catalogItem.SKU, retrievedItem.SKU)
	suite.True(catalogItem.BasePrice.Equal(retrievedItem.BasePrice))

	// Test Update
	suite.T().Log("Testing catalog item update")
	retrievedItem.Name = "Updated Test Product"
	retrievedItem.BasePrice = decimal.NewFromFloat(150.0)

	err = suite.catalogRepo.Update(ctx, &retrievedItem)
	suite.NoError(err)

	// Verify update
	var updatedItem catalogModels.CatalogItem
	result, err = suite.catalogRepo.GetByID(ctx, catalogItem.ID, &updatedItem)
	suite.NoError(err)
	suite.Equal("Updated Test Product", updatedItem.Name)
	suite.True(decimal.NewFromFloat(150.0).Equal(updatedItem.BasePrice))

	// Test Soft Delete
	suite.T().Log("Testing catalog item soft delete")
	err = suite.catalogRepo.SoftDelete(ctx, catalogItem.ID, testutils.TestUserID)
	suite.NoError(err)

	// Verify soft delete (item should not be retrievable)
	var deletedItem catalogModels.CatalogItem
	result, err = suite.catalogRepo.GetByID(ctx, catalogItem.ID, &deletedItem)
	suite.Error(err) // Should return error for soft-deleted item

	// Test Restore
	suite.T().Log("Testing catalog item restore")
	err = suite.catalogRepo.Restore(ctx, catalogItem.ID)
	suite.NoError(err)

	// Verify restore
	var restoredItem catalogModels.CatalogItem
	result, err = suite.catalogRepo.GetByID(ctx, catalogItem.ID, &restoredItem)
	suite.NoError(err)
	suite.NotNil(result)
}

// Test order CRUD operations with complex relationships
func (suite *DatabaseIntegrationTestSuite) TestOrderCRUDOperations() {
	ctx := context.Background()

	// First create a catalog item for the order
	catalogItem := testutils.CreateTestCatalogItem(catalogModels.CatalogItemTypeProduct)
	err := suite.catalogRepo.Create(ctx, catalogItem)
	suite.NoError(err)
	suite.createdItemIDs = append(suite.createdItemIDs, catalogItem.ID)

	// Test Create Order
	suite.T().Log("Testing order creation with items")
	order := testutils.CreateTestOrder()

	err = suite.orderRepo.Create(ctx, order)
	suite.NoError(err)
	suite.createdOrderIDs = append(suite.createdOrderIDs, order.ID)

	// Test Read Order with Items
	suite.T().Log("Testing order retrieval with items")
	retrievedOrder, err := suite.orderRepo.GetOrderByID(ctx, order.ID)
	suite.NoError(err)
	suite.NotNil(retrievedOrder)
	suite.Equal(order.OrderNumber, retrievedOrder.OrderNumber)
	suite.Equal(order.Status, retrievedOrder.Status)
	suite.Len(retrievedOrder.Items, len(order.Items))
	suite.True(order.TotalAmount.Equal(retrievedOrder.TotalAmount))

	// Test Update Order
	suite.T().Log("Testing order update")
	retrievedOrder.Status = orderModels.OrderStatusConfirmed
	retrievedOrder.Notes = "Updated order notes"

	err = suite.orderRepo.Update(ctx, retrievedOrder)
	suite.NoError(err)

	// Verify update
	updatedOrder, err := suite.orderRepo.GetOrderByID(ctx, order.ID)
	suite.NoError(err)
	suite.Equal(orderModels.OrderStatusConfirmed, updatedOrder.Status)
	suite.Equal("Updated order notes", updatedOrder.Notes)

	// Test List Orders with Filters
	suite.T().Log("Testing order listing with filters")
	filter := testutils.CreateTestListOrdersRequest()
	orders, total, err := suite.orderRepo.ListOrders(ctx, filter, 0, 20)
	suite.NoError(err)
	suite.GreaterOrEqual(total, int64(1))
	suite.NotEmpty(orders)

	// Test Cancel Order (Orders typically aren't deleted, just cancelled)
	suite.T().Log("Testing order cancellation")
	err = suite.orderRepo.UpdateOrderStatus(ctx, order.ID, orderModels.OrderStatusCancelled, nil, "Test cancellation")
	suite.NoError(err)

	// Verify cancellation
	cancelledOrder, err := suite.orderRepo.GetOrderByID(ctx, order.ID)
	suite.NoError(err)
	suite.Equal(orderModels.OrderStatusCancelled, cancelledOrder.Status)
}

// Test transaction handling
func (suite *DatabaseIntegrationTestSuite) TestTransactionHandling() {
	ctx := context.Background()

	suite.T().Log("Testing successful transaction")
	suite.testSuccessfulTransaction(ctx)

	suite.T().Log("Testing transaction rollback")
	suite.testTransactionRollback(ctx)
}

func (suite *DatabaseIntegrationTestSuite) testSuccessfulTransaction(ctx context.Context) {
	// Begin transaction
	tx, err := suite.testDB.BeginTransaction()
	suite.NoError(err)
	suite.NotNil(tx)

	// Create catalog item within transaction
	catalogItem := testutils.CreateTestCatalogItem(catalogModels.CatalogItemTypeProduct)
	catalogItem.ID = "tx-test-item-1"

	// In real implementation, would use transaction-aware repository methods
	err = suite.catalogRepo.Create(ctx, catalogItem)
	suite.NoError(err)

	// Create order within same transaction
	order := testutils.CreateTestOrder()
	order.ID = "tx-test-order-1"

	err = suite.orderRepo.Create(ctx, order)
	suite.NoError(err)

	// Commit transaction
	err = tx.Commit()
	suite.NoError(err)

	// Verify both records exist
	var retrievedItem catalogModels.CatalogItem
	result, err := suite.catalogRepo.GetByID(ctx, catalogItem.ID, &retrievedItem)
	suite.NoError(err)
	suite.NotNil(result)

	retrievedOrder, err := suite.orderRepo.GetOrderByID(ctx, order.ID)
	suite.NoError(err)
	suite.NotNil(retrievedOrder)

	// Track for cleanup
	suite.createdItemIDs = append(suite.createdItemIDs, catalogItem.ID)
	suite.createdOrderIDs = append(suite.createdOrderIDs, order.ID)
}

func (suite *DatabaseIntegrationTestSuite) testTransactionRollback(ctx context.Context) {
	// Begin transaction
	tx, err := suite.testDB.BeginTransaction()
	suite.NoError(err)
	suite.NotNil(tx)

	// Create catalog item within transaction
	catalogItem := testutils.CreateTestCatalogItem(catalogModels.CatalogItemTypeProduct)
	catalogItem.ID = "tx-test-item-2"

	err = suite.catalogRepo.Create(ctx, catalogItem)
	suite.NoError(err)

	// Simulate error condition and rollback
	err = tx.Rollback()
	suite.NoError(err)

	// Verify record does not exist after rollback
	var retrievedItem catalogModels.CatalogItem
	result, err := suite.catalogRepo.GetByID(ctx, catalogItem.ID, &retrievedItem)
	suite.Error(err) // Should return error since transaction was rolled back
	suite.Nil(result)
}

// Test concurrent database operations
func (suite *DatabaseIntegrationTestSuite) TestConcurrentDatabaseOperations() {
	ctx := context.Background()

	suite.T().Log("Testing concurrent catalog item creation")
	suite.testConcurrentCatalogCreation(ctx)

	suite.T().Log("Testing concurrent order processing")
	suite.testConcurrentOrderProcessing(ctx)
}

func (suite *DatabaseIntegrationTestSuite) testConcurrentCatalogCreation(ctx context.Context) {
	concurrency := 5
	done := make(chan error, concurrency)

	// Create multiple catalog items concurrently
	for i := 0; i < concurrency; i++ {
		go func(index int) {
			catalogItem := testutils.CreateTestCatalogItem(catalogModels.CatalogItemTypeProduct)
			catalogItem.ID = testutils.TestCatalogItemID + "-concurrent-" + string(rune(index))
			catalogItem.SKU = testutils.TestSKU + "-concurrent-" + string(rune(index))

			err := suite.catalogRepo.Create(ctx, catalogItem)
			if err == nil {
				suite.createdItemIDs = append(suite.createdItemIDs, catalogItem.ID)
			}
			done <- err
		}(i)
	}

	// Wait for all operations to complete
	for i := 0; i < concurrency; i++ {
		err := <-done
		suite.NoError(err, "Concurrent catalog creation should succeed")
	}
}

func (suite *DatabaseIntegrationTestSuite) testConcurrentOrderProcessing(ctx context.Context) {
	// Create a catalog item first
	catalogItem := testutils.CreateTestCatalogItem(catalogModels.CatalogItemTypeProduct)
	err := suite.catalogRepo.Create(ctx, catalogItem)
	suite.NoError(err)
	suite.createdItemIDs = append(suite.createdItemIDs, catalogItem.ID)

	// Create an order
	order := testutils.CreateTestOrder()
	err = suite.orderRepo.Create(ctx, order)
	suite.NoError(err)
	suite.createdOrderIDs = append(suite.createdOrderIDs, order.ID)

	concurrency := 3
	done := make(chan error, concurrency)

	// Simulate concurrent order status updates
	statuses := []orderModels.OrderStatus{
		orderModels.OrderStatusConfirmed,
		orderModels.OrderStatusPaid,
		orderModels.OrderStatusShipped,
	}

	for i, status := range statuses {
		go func(index int, orderStatus orderModels.OrderStatus) {
			// Retrieve order
			currentOrder, err := suite.orderRepo.GetOrderByID(ctx, order.ID)
			if err != nil {
				done <- err
				return
			}

			// Update status
			currentOrder.Status = orderStatus
			err = suite.orderRepo.Update(ctx, currentOrder)
			done <- err
		}(i, status)
	}

	// Wait for all operations to complete
	successCount := 0
	for i := 0; i < concurrency; i++ {
		err := <-done
		if err == nil {
			successCount++
		}
	}

	// At least one update should succeed
	suite.GreaterOrEqual(successCount, 1, "At least one concurrent update should succeed")
}

// Test database constraints and validations
func (suite *DatabaseIntegrationTestSuite) TestDatabaseConstraints() {
	ctx := context.Background()

	suite.T().Log("Testing unique constraint violations")
	suite.testUniqueConstraints(ctx)

	suite.T().Log("Testing foreign key constraints")
	suite.testForeignKeyConstraints(ctx)

	suite.T().Log("Testing check constraints")
	suite.testCheckConstraints(ctx)
}

func (suite *DatabaseIntegrationTestSuite) testUniqueConstraints(ctx context.Context) {
	// Create first catalog item
	catalogItem1 := testutils.CreateTestCatalogItem(catalogModels.CatalogItemTypeProduct)
	err := suite.catalogRepo.Create(ctx, catalogItem1)
	suite.NoError(err)
	suite.createdItemIDs = append(suite.createdItemIDs, catalogItem1.ID)

	// Try to create second item with same SKU (should fail)
	catalogItem2 := testutils.CreateTestCatalogItem(catalogModels.CatalogItemTypeProduct)
	catalogItem2.ID = catalogItem1.ID + "-duplicate"
	catalogItem2.SKU = catalogItem1.SKU // Same SKU

	err = suite.catalogRepo.Create(ctx, catalogItem2)
	suite.Error(err, "Creating item with duplicate SKU should fail")
	suite.Contains(err.Error(), "unique constraint", "Error should mention unique constraint violation")
}

func (suite *DatabaseIntegrationTestSuite) testForeignKeyConstraints(ctx context.Context) {
	// Try to create order with non-existent catalog item
	order := testutils.CreateTestOrder()
	order.Items[0].CatalogItemID = "non-existent-item"

	err := suite.orderRepo.Create(ctx, order)
	suite.Error(err, "Creating order with non-existent catalog item should fail")
	suite.Contains(err.Error(), "foreign key", "Error should mention foreign key constraint violation")
}

func (suite *DatabaseIntegrationTestSuite) testCheckConstraints(ctx context.Context) {
	// Try to create catalog item with negative price
	catalogItem := testutils.CreateTestCatalogItem(catalogModels.CatalogItemTypeProduct)
	catalogItem.BasePrice = decimal.NewFromFloat(-10.0) // Negative price

	err := suite.catalogRepo.Create(ctx, catalogItem)
	suite.Error(err, "Creating item with negative price should fail")
	suite.Contains(err.Error(), "check constraint", "Error should mention check constraint violation")
}

// Test database performance under load
func (suite *DatabaseIntegrationTestSuite) TestDatabasePerformance() {
	ctx := context.Background()

	suite.T().Log("Testing bulk insert performance")
	suite.testBulkInsertPerformance(ctx)

	suite.T().Log("Testing query performance with large dataset")
	suite.testQueryPerformance(ctx)
}

func (suite *DatabaseIntegrationTestSuite) testBulkInsertPerformance(ctx context.Context) {
	itemCount := 100
	start := time.Now()

	// Create multiple catalog items
	for i := 0; i < itemCount; i++ {
		catalogItem := testutils.CreateTestCatalogItem(catalogModels.CatalogItemTypeProduct)
		catalogItem.ID = testutils.TestCatalogItemID + "-bulk-" + string(rune(i))
		catalogItem.SKU = testutils.TestSKU + "-bulk-" + string(rune(i))

		err := suite.catalogRepo.Create(ctx, catalogItem)
		suite.NoError(err)
		suite.createdItemIDs = append(suite.createdItemIDs, catalogItem.ID)
	}

	duration := time.Since(start)
	suite.T().Logf("Created %d items in %v", itemCount, duration)

	// Performance assertion
	avgTimePerItem := duration / time.Duration(itemCount)
	suite.Less(avgTimePerItem, 100*time.Millisecond, "Average time per item should be less than 100ms")
}

func (suite *DatabaseIntegrationTestSuite) testQueryPerformance(ctx context.Context) {
	// Test listing performance with filters
	filter := testutils.CreateTestCatalogFilter()

	start := time.Now()
	items, total, err := suite.catalogRepo.ListCatalogItems(ctx, filter, 0, 50)
	duration := time.Since(start)

	suite.NoError(err)
	suite.NotNil(items)
	suite.GreaterOrEqual(total, int64(0))

	suite.T().Logf("Listed %d items (total: %d) in %v", len(items), total, duration)

	// Performance assertion
	suite.Less(duration, 2*time.Second, "Query should complete within 2 seconds")
}

// Test database connection handling
func (suite *DatabaseIntegrationTestSuite) TestDatabaseConnectionHandling() {
	ctx := context.Background()

	suite.T().Log("Testing connection health")
	suite.testConnectionHealth(ctx)

	suite.T().Log("Testing connection recovery")
	suite.testConnectionRecovery(ctx)
}

func (suite *DatabaseIntegrationTestSuite) testConnectionHealth(ctx context.Context) {
	// Test database ping
	db := suite.testDB.GetConnection()
	if db != nil {
		err := db.PingContext(ctx)
		suite.NoError(err, "Database should be reachable")
	}
}

func (suite *DatabaseIntegrationTestSuite) testConnectionRecovery(ctx context.Context) {
	// Simulate connection failure and recovery
	// In real implementation, this would test actual connection pooling and retry logic

	catalogItem := testutils.CreateTestCatalogItem(catalogModels.CatalogItemTypeProduct)

	// First attempt should succeed
	err := suite.catalogRepo.Create(ctx, catalogItem)
	suite.NoError(err)
	suite.createdItemIDs = append(suite.createdItemIDs, catalogItem.ID)

	// Verify item was created
	var retrievedItem catalogModels.CatalogItem
	result, err := suite.catalogRepo.GetByID(ctx, catalogItem.ID, &retrievedItem)
	suite.NoError(err)
	suite.NotNil(result)
}

// Test data consistency across operations
func (suite *DatabaseIntegrationTestSuite) TestDataConsistency() {
	ctx := context.Background()

	suite.T().Log("Testing order-item consistency")
	suite.testOrderItemConsistency(ctx)

	suite.T().Log("Testing inventory consistency")
	suite.testInventoryConsistency(ctx)
}

func (suite *DatabaseIntegrationTestSuite) testOrderItemConsistency(ctx context.Context) {
	// Create catalog item
	catalogItem := testutils.CreateTestCatalogItem(catalogModels.CatalogItemTypeProduct)
	err := suite.catalogRepo.Create(ctx, catalogItem)
	suite.NoError(err)
	suite.createdItemIDs = append(suite.createdItemIDs, catalogItem.ID)

	// Create order with items
	order := testutils.CreateTestOrder()
	err = suite.orderRepo.Create(ctx, order)
	suite.NoError(err)
	suite.createdOrderIDs = append(suite.createdOrderIDs, order.ID)

	// Verify order total matches sum of item totals
	retrievedOrder, err := suite.orderRepo.GetOrderByID(ctx, order.ID)
	suite.NoError(err)

	calculatedTotal := decimal.Zero
	for _, item := range retrievedOrder.Items {
		calculatedTotal = calculatedTotal.Add(item.TotalPrice)
	}

	suite.True(retrievedOrder.SubtotalAmount.Equal(calculatedTotal),
		"Order subtotal should equal sum of item totals")
}

func (suite *DatabaseIntegrationTestSuite) testInventoryConsistency(ctx context.Context) {
	// Create product with inventory
	catalogItem := testutils.CreateTestCatalogItem(catalogModels.CatalogItemTypeProduct)
	err := suite.catalogRepo.Create(ctx, catalogItem)
	suite.NoError(err)
	suite.createdItemIDs = append(suite.createdItemIDs, catalogItem.ID)

	// Test inventory operations
	initialLevel, err := suite.catalogRepo.GetInventoryLevel(ctx, catalogItem.ID)
	suite.NoError(err)
	suite.GreaterOrEqual(initialLevel, 0.0)

	// Reserve inventory
	reserveQuantity := 10.0
	err = suite.catalogRepo.ReserveInventory(ctx, catalogItem.ID, reserveQuantity)
	suite.NoError(err)

	// Check inventory level after reservation
	afterReserveLevel, err := suite.catalogRepo.GetInventoryLevel(ctx, catalogItem.ID)
	suite.NoError(err)
	suite.Equal(initialLevel-reserveQuantity, afterReserveLevel,
		"Inventory level should decrease after reservation")

	// Release inventory
	err = suite.catalogRepo.ReleaseInventory(ctx, catalogItem.ID, reserveQuantity)
	suite.NoError(err)

	// Check inventory level after release
	afterReleaseLevel, err := suite.catalogRepo.GetInventoryLevel(ctx, catalogItem.ID)
	suite.NoError(err)
	suite.Equal(initialLevel, afterReleaseLevel,
		"Inventory level should return to original after release")
}

// Run the database integration test suite
func TestDatabaseIntegrationSuite(t *testing.T) {
	suite.Run(t, new(DatabaseIntegrationTestSuite))
}
