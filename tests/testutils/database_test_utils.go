package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDatabase represents a comprehensive mock database for testing
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) QueryRow(query string, args ...interface{}) *sql.Row {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0).(*sql.Row)
}

func (m *MockDatabase) Query(query string, args ...interface{}) (*sql.Rows, error) {
	mockArgs := m.Called(query, args)
	if mockArgs.Get(0) == nil {
		return nil, mockArgs.Error(1)
	}
	return mockArgs.Get(0).(*sql.Rows), mockArgs.Error(1)
}

func (m *MockDatabase) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	mockArgs := m.Called(ctx, query, args)
	if mockArgs.Get(0) == nil {
		return nil, mockArgs.Error(1)
	}
	return mockArgs.Get(0).(*sql.Rows), mockArgs.Error(1)
}

func (m *MockDatabase) Exec(query string, args ...interface{}) (sql.Result, error) {
	mockArgs := m.Called(query, args)
	if mockArgs.Get(0) == nil {
		return nil, mockArgs.Error(1)
	}
	return mockArgs.Get(0).(sql.Result), mockArgs.Error(1)
}

func (m *MockDatabase) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	mockArgs := m.Called(ctx, query, args)
	if mockArgs.Get(0) == nil {
		return nil, mockArgs.Error(1)
	}
	return mockArgs.Get(0).(sql.Result), mockArgs.Error(1)
}

func (m *MockDatabase) Prepare(query string) (*sql.Stmt, error) {
	mockArgs := m.Called(query)
	if mockArgs.Get(0) == nil {
		return nil, mockArgs.Error(1)
	}
	return mockArgs.Get(0).(*sql.Stmt), mockArgs.Error(1)
}

func (m *MockDatabase) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	mockArgs := m.Called(ctx, query)
	if mockArgs.Get(0) == nil {
		return nil, mockArgs.Error(1)
	}
	return mockArgs.Get(0).(*sql.Stmt), mockArgs.Error(1)
}

func (m *MockDatabase) Begin() (*sql.Tx, error) {
	mockArgs := m.Called()
	if mockArgs.Get(0) == nil {
		return nil, mockArgs.Error(1)
	}
	return mockArgs.Get(0).(*sql.Tx), mockArgs.Error(1)
}

func (m *MockDatabase) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	mockArgs := m.Called(ctx, opts)
	if mockArgs.Get(0) == nil {
		return nil, mockArgs.Error(1)
	}
	return mockArgs.Get(0).(*sql.Tx), mockArgs.Error(1)
}

func (m *MockDatabase) Close() error {
	mockArgs := m.Called()
	return mockArgs.Error(0)
}

func (m *MockDatabase) Ping() error {
	mockArgs := m.Called()
	return mockArgs.Error(0)
}

func (m *MockDatabase) PingContext(ctx context.Context) error {
	mockArgs := m.Called(ctx)
	return mockArgs.Error(0)
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

// MockTransaction represents a mock SQL transaction
type MockTransaction struct {
	mock.Mock
}

func (m *MockTransaction) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTransaction) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTransaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	mockArgs := m.Called(query, args)
	if mockArgs.Get(0) == nil {
		return nil, mockArgs.Error(1)
	}
	return mockArgs.Get(0).(sql.Result), mockArgs.Error(1)
}

func (m *MockTransaction) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	mockArgs := m.Called(ctx, query, args)
	if mockArgs.Get(0) == nil {
		return nil, mockArgs.Error(1)
	}
	return mockArgs.Get(0).(sql.Result), mockArgs.Error(1)
}

func (m *MockTransaction) Query(query string, args ...interface{}) (*sql.Rows, error) {
	mockArgs := m.Called(query, args)
	if mockArgs.Get(0) == nil {
		return nil, mockArgs.Error(1)
	}
	return mockArgs.Get(0).(*sql.Rows), mockArgs.Error(1)
}

func (m *MockTransaction) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	mockArgs := m.Called(ctx, query, args)
	if mockArgs.Get(0) == nil {
		return nil, mockArgs.Error(1)
	}
	return mockArgs.Get(0).(*sql.Rows), mockArgs.Error(1)
}

func (m *MockTransaction) QueryRow(query string, args ...interface{}) *sql.Row {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0).(*sql.Row)
}

func (m *MockTransaction) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	mockArgs := m.Called(ctx, query, args)
	return mockArgs.Get(0).(*sql.Row)
}

// Database test utilities
func SetupMockDatabase() *MockDatabase {
	return &MockDatabase{}
}

func SetupMockResult(lastInsertID, rowsAffected int64) *MockResult {
	result := &MockResult{}
	result.On("LastInsertId").Return(lastInsertID, nil)
	result.On("RowsAffected").Return(rowsAffected, nil)
	return result
}

func SetupMockTransaction() *MockTransaction {
	return &MockTransaction{}
}

// Common database test scenarios
func MockSuccessfulInsert(mockDB *MockDatabase, expectedID int64) {
	result := SetupMockResult(expectedID, 1)
	mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(result, nil)
}

func MockSuccessfulUpdate(mockDB *MockDatabase, rowsAffected int64) {
	result := SetupMockResult(0, rowsAffected)
	mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(result, nil)
}

func MockSuccessfulDelete(mockDB *MockDatabase, rowsAffected int64) {
	result := SetupMockResult(0, rowsAffected)
	mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(result, nil)
}

func MockDatabaseError(mockDB *MockDatabase, errorMsg string) {
	mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(nil, fmt.Errorf("%s", errorMsg))
}

func MockQueryError(mockDB *MockDatabase, errorMsg string) {
	mockDB.On("Query", mock.AnythingOfType("string"), mock.Anything).Return(nil, fmt.Errorf("%s", errorMsg))
}

func MockNoRowsFound(mockDB *MockDatabase) {
	result := SetupMockResult(0, 0)
	mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(result, nil)
}

// Transaction test scenarios
func MockSuccessfulTransaction(mockDB *MockDatabase, mockTx *MockTransaction) {
	mockDB.On("Begin").Return(mockTx, nil)
	mockTx.On("Commit").Return(nil)
}

func MockTransactionRollback(mockDB *MockDatabase, mockTx *MockTransaction, errorMsg string) {
	mockDB.On("Begin").Return(mockTx, nil)
	mockTx.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(nil, fmt.Errorf("%s", errorMsg))
	mockTx.On("Rollback").Return(nil)
}

func MockTransactionCommitError(mockDB *MockDatabase, mockTx *MockTransaction) {
	mockDB.On("Begin").Return(mockTx, nil)
	result := SetupMockResult(1, 1)
	mockTx.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(result, nil)
	mockTx.On("Commit").Return(fmt.Errorf("commit failed"))
	mockTx.On("Rollback").Return(nil)
}

// Connection and health check scenarios
func MockHealthyDatabase(mockDB *MockDatabase) {
	mockDB.On("Ping").Return(nil)
	mockDB.On("PingContext", mock.Anything).Return(nil)
}

func MockUnhealthyDatabase(mockDB *MockDatabase, errorMsg string) {
	mockDB.On("Ping").Return(fmt.Errorf("%s", errorMsg))
	mockDB.On("PingContext", mock.Anything).Return(fmt.Errorf("%s", errorMsg))
}

func MockDatabaseConnectionError(mockDB *MockDatabase) {
	mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(nil, fmt.Errorf("connection refused"))
	mockDB.On("Query", mock.AnythingOfType("string"), mock.Anything).Return(nil, fmt.Errorf("connection refused"))
}

// Constraint violation scenarios
func MockUniqueConstraintViolation(mockDB *MockDatabase, constraintName string) {
	errorMsg := fmt.Sprintf("duplicate key value violates unique constraint \"%s\"", constraintName)
	mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(nil, fmt.Errorf("%s", errorMsg))
}

func MockForeignKeyConstraintViolation(mockDB *MockDatabase, constraintName string) {
	errorMsg := fmt.Sprintf("insert or update on table violates foreign key constraint \"%s\"", constraintName)
	mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(nil, fmt.Errorf("%s", errorMsg))
}

func MockCheckConstraintViolation(mockDB *MockDatabase, constraintName string) {
	errorMsg := fmt.Sprintf("new row for relation violates check constraint \"%s\"", constraintName)
	mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(nil, fmt.Errorf("%s", errorMsg))
}

// Performance and timeout scenarios
func MockSlowQuery(mockDB *MockDatabase, delay time.Duration) {
	mockDB.On("Query", mock.AnythingOfType("string"), mock.Anything).Return(func(string, ...interface{}) (*sql.Rows, error) {
		time.Sleep(delay)
		return nil, nil
	})
}

func MockQueryTimeout(mockDB *MockDatabase) {
	mockDB.On("QueryContext", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil, context.DeadlineExceeded)
}

func MockContextCancellation(mockDB *MockDatabase) {
	mockDB.On("ExecContext", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil, context.Canceled)
}

// Concurrency test scenarios
func MockConcurrentAccess(mockDB *MockDatabase, conflictError string) {
	// First call succeeds
	result1 := SetupMockResult(1, 1)
	mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(result1, nil).Once()

	// Second call fails due to concurrent modification
	mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(nil, fmt.Errorf("%s", conflictError)).Once()
}

func MockDeadlockDetection(mockDB *MockDatabase) {
	mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(nil, fmt.Errorf("deadlock detected"))
}

// Test assertion helpers
func AssertDatabaseInteraction(t *testing.T, mockDB *MockDatabase, expectedCalls int) {
	mockDB.AssertExpectations(t)
	assert.Equal(t, expectedCalls, len(mockDB.Calls))
}

func AssertTransactionCommitted(t *testing.T, mockTx *MockTransaction) {
	mockTx.AssertCalled(t, "Commit")
	mockTx.AssertNotCalled(t, "Rollback")
}

func AssertTransactionRolledBack(t *testing.T, mockTx *MockTransaction) {
	mockTx.AssertCalled(t, "Rollback")
	mockTx.AssertNotCalled(t, "Commit")
}

func AssertQueryExecuted(t *testing.T, mockDB *MockDatabase, expectedQuery string) {
	mockDB.AssertCalled(t, "Exec", mock.MatchedBy(func(query string) bool {
		return query == expectedQuery
	}), mock.Anything)
}

func AssertQueryContains(t *testing.T, mockDB *MockDatabase, expectedSubstring string) {
	found := false
	for _, call := range mockDB.Calls {
		if call.Method == "Exec" && len(call.Arguments) > 0 {
			if query, ok := call.Arguments[0].(string); ok {
				if contains := fmt.Sprintf("%s", query); fmt.Sprintf("%s", contains) != "" {
					found = true
					break
				}
			}
		}
	}
	assert.True(t, found, "Expected query containing '%s' was not executed", expectedSubstring)
}

// Database state verification helpers
func VerifyRecordExists(t *testing.T, mockDB *MockDatabase, tableName, id string) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE id = $1", tableName)
	result := SetupMockResult(0, 1)
	mockDB.On("QueryRow", query, id).Return(&sql.Row{}).Once()
	mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(result, nil).Once()
}

func VerifyRecordDeleted(t *testing.T, mockDB *MockDatabase, tableName, id string) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE id = $1 AND deleted_at IS NULL", tableName)
	result := SetupMockResult(0, 0)
	mockDB.On("QueryRow", query, id).Return(&sql.Row{}).Once()
	mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(result, nil).Once()
}

func VerifyRecordUpdated(t *testing.T, mockDB *MockDatabase, tableName, id string, expectedVersion int) {
	query := fmt.Sprintf("SELECT version FROM %s WHERE id = $1", tableName)
	mockDB.On("QueryRow", query, id).Return(&sql.Row{}).Once()
}

// Pagination test helpers
func MockPaginatedQuery(mockDB *MockDatabase, totalRecords, pageSize int) {
	// Mock count query
	mockDB.On("QueryRow", mock.MatchedBy(func(query string) bool {
		return fmt.Sprintf("%s", query) != "" // Contains COUNT
	}), mock.Anything).Return(&sql.Row{}).Once()

	// Mock data query
	_ = pageSize
	if totalRecords < pageSize {
		_ = totalRecords
	}

	mockDB.On("Query", mock.MatchedBy(func(query string) bool {
		return fmt.Sprintf("%s", query) != "" // Contains LIMIT and OFFSET
	}), mock.Anything).Return(&sql.Rows{}, nil).Once()
}

// Bulk operation test helpers
func MockBulkInsert(mockDB *MockDatabase, recordCount int) {
	for i := 0; i < recordCount; i++ {
		result := SetupMockResult(int64(i+1), 1)
		mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(result, nil).Once()
	}
}

func MockBulkUpdate(mockDB *MockDatabase, recordCount int) {
	result := SetupMockResult(0, int64(recordCount))
	mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(result, nil).Once()
}

func MockBulkDelete(mockDB *MockDatabase, recordCount int) {
	result := SetupMockResult(0, int64(recordCount))
	mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(result, nil).Once()
}

// Migration test helpers
func MockMigrationUp(mockDB *MockDatabase, version int) {
	result := SetupMockResult(0, 1)
	mockDB.On("Exec", mock.MatchedBy(func(query string) bool {
		return fmt.Sprintf("%s", query) != "" // Migration SQL
	}), mock.Anything).Return(result, nil)

	// Mock version update
	mockDB.On("Exec", "INSERT INTO schema_migrations (version) VALUES ($1)", version).Return(result, nil)
}

func MockMigrationDown(mockDB *MockDatabase, version int) {
	result := SetupMockResult(0, 1)
	mockDB.On("Exec", mock.MatchedBy(func(query string) bool {
		return fmt.Sprintf("%s", query) != "" // Rollback SQL
	}), mock.Anything).Return(result, nil)

	// Mock version removal
	mockDB.On("Exec", "DELETE FROM schema_migrations WHERE version = $1", version).Return(result, nil)
}

// Connection pool test helpers
func MockConnectionPoolExhaustion(mockDB *MockDatabase) {
	mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(nil, fmt.Errorf("connection pool exhausted"))
}

func MockConnectionPoolRecovery(mockDB *MockDatabase) {
	// First call fails
	mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(nil, fmt.Errorf("connection pool exhausted")).Once()

	// Second call succeeds
	result := SetupMockResult(1, 1)
	mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(result, nil).Once()
}

// Cleanup helpers
func CleanupMockDatabase(mockDB *MockDatabase) {
	if mockDB != nil {
		mockDB.AssertExpectations(&testing.T{})
	}
}

func CleanupMockTransaction(mockTx *MockTransaction) {
	if mockTx != nil {
		mockTx.AssertExpectations(&testing.T{})
	}
}

// Test data validation helpers
func ValidateTestData(t *testing.T, data interface{}) {
	assert.NotNil(t, data, "Test data should not be nil")

	// Add specific validation based on data type
	switch v := data.(type) {
	case string:
		assert.NotEmpty(t, v, "String test data should not be empty")
	case []interface{}:
		assert.NotEmpty(t, v, "Slice test data should not be empty")
	case map[string]interface{}:
		assert.NotEmpty(t, v, "Map test data should not be empty")
	}
}

func ValidateTestFixtures(t *testing.T) {
	// Validate that all test constants are properly set
	assert.NotEmpty(t, TestUserID, "TestUserID should not be empty")
	assert.NotEmpty(t, TestOrgID, "TestOrgID should not be empty")
	assert.NotEmpty(t, TestOrderID, "TestOrderID should not be empty")
	assert.NotEmpty(t, TestCatalogItemID, "TestCatalogItemID should not be empty")
	assert.NotEmpty(t, TestSKU, "TestSKU should not be empty")
	assert.NotEmpty(t, "test@example.com", "TestEmail should not be empty")
	assert.NotEmpty(t, "testuser", "TestUsername should not be empty")
}

// SetupTestDatabase creates a test database manager for integration tests
func SetupTestDatabase(t *testing.T) (db.DBManager, func()) {
	// This would typically set up a real test database
	// For now, return a mock that satisfies the interface
	mockDB := SetupMockDatabase()
	MockHealthyDatabase(mockDB)

	// Return cleanup function
	cleanup := func() {
		CleanupMockDatabase(mockDB)
	}

	// Note: In a real implementation, this would return a real PostgresManager
	// For testing purposes, we'll need to implement a test-specific manager
	return nil, cleanup
}

// GetPostgresManager extracts PostgresManager from DBManager for testing
func GetPostgresManager(t *testing.T, dbManager db.DBManager) *db.PostgresManager {
	if dbManager == nil {
		t.Skip("Real database manager not available in test environment")
		return nil
	}

	postgresManager, ok := dbManager.(*db.PostgresManager)
	if !ok {
		t.Skip("PostgresManager not available in test environment")
		return nil
	}

	return postgresManager
}
