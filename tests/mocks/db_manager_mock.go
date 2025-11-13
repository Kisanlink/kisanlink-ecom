// Package mocks provides mock implementations for testing purposes.
// This includes MockDBManager which implements the db.DBManager interface for unit testing.
package mocks

import (
	"context"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockDBManager is a mock implementation of the DBManager interface for testing
type MockDBManager struct {
	mock.Mock
}

// Create mocks the Create method
func (m *MockDBManager) Create(ctx context.Context, entity interface{}) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

// GetByID mocks the GetByID method
func (m *MockDBManager) GetByID(ctx context.Context, id interface{}, result interface{}) error {
	args := m.Called(ctx, id, result)
	return args.Error(0)
}

// Update mocks the Update method
func (m *MockDBManager) Update(ctx context.Context, entity interface{}) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *MockDBManager) Delete(ctx context.Context, id interface{}, entity interface{}) error {
	args := m.Called(ctx, id, entity)
	return args.Error(0)
}

// List mocks the List method
func (m *MockDBManager) List(ctx context.Context, filter *base.Filter, result interface{}) error {
	args := m.Called(ctx, filter, result)
	return args.Error(0)
}

// Count mocks the Count method
func (m *MockDBManager) Count(ctx context.Context, filter *base.Filter, model interface{}) (int64, error) {
	args := m.Called(ctx, filter, model)
	return args.Get(0).(int64), args.Error(1)
}

// ExecuteInTransaction mocks the ExecuteInTransaction method
func (m *MockDBManager) ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

// Close mocks the Close method
func (m *MockDBManager) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Connect mocks the Connect method
func (m *MockDBManager) Connect(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// IsConnected mocks the IsConnected method
func (m *MockDBManager) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

// GetBackendType mocks the GetBackendType method
func (m *MockDBManager) GetBackendType() db.BackendType {
	args := m.Called()
	return args.Get(0).(db.BackendType)
}

// SoftDelete mocks the SoftDelete method
func (m *MockDBManager) SoftDelete(ctx context.Context, id interface{}, model interface{}, deletedBy string) error {
	args := m.Called(ctx, id, model, deletedBy)
	return args.Error(0)
}

// Restore mocks the Restore method
func (m *MockDBManager) Restore(ctx context.Context, id interface{}, model interface{}) error {
	args := m.Called(ctx, id, model)
	return args.Error(0)
}

// ListWithDeleted mocks the ListWithDeleted method
func (m *MockDBManager) ListWithDeleted(ctx context.Context, limit, offset int, models interface{}) error {
	args := m.Called(ctx, limit, offset, models)
	return args.Error(0)
}

// CountWithDeleted mocks the CountWithDeleted method
func (m *MockDBManager) CountWithDeleted(ctx context.Context, model interface{}) (int64, error) {
	args := m.Called(ctx, model)
	return args.Get(0).(int64), args.Error(1)
}

// ExistsWithDeleted mocks the ExistsWithDeleted method
func (m *MockDBManager) ExistsWithDeleted(ctx context.Context, id interface{}) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

// GetByCreatedBy mocks the GetByCreatedBy method
func (m *MockDBManager) GetByCreatedBy(ctx context.Context, createdBy interface{}, limit, offset int, models interface{}) error {
	args := m.Called(ctx, createdBy, limit, offset, models)
	return args.Error(0)
}

// GetByUpdatedBy mocks the GetByUpdatedBy method
func (m *MockDBManager) GetByUpdatedBy(ctx context.Context, updatedBy interface{}, limit, offset int, models interface{}) error {
	args := m.Called(ctx, updatedBy, limit, offset, models)
	return args.Error(0)
}

// GetByDeletedBy mocks the GetByDeletedBy method
func (m *MockDBManager) GetByDeletedBy(ctx context.Context, deletedBy interface{}, limit, offset int, models interface{}) error {
	args := m.Called(ctx, deletedBy, limit, offset, models)
	return args.Error(0)
}

// CreateMany mocks the CreateMany method
func (m *MockDBManager) CreateMany(ctx context.Context, models []interface{}) error {
	args := m.Called(ctx, models)
	return args.Error(0)
}

// UpdateMany mocks the UpdateMany method
func (m *MockDBManager) UpdateMany(ctx context.Context, models []interface{}) error {
	args := m.Called(ctx, models)
	return args.Error(0)
}

// DeleteMany mocks the DeleteMany method
func (m *MockDBManager) DeleteMany(ctx context.Context, ids []interface{}) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

// AutoMigrateModels mocks the AutoMigrateModels method
func (m *MockDBManager) AutoMigrateModels(ctx context.Context, models ...interface{}) error {
	args := m.Called(ctx, models)
	return args.Error(0)
}

// GetDB mocks the GetDB method (optional, for accessing raw GORM DB)
func (m *MockDBManager) GetDB() *gorm.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*gorm.DB)
}

// NewMockDBManager creates a new MockDBManager instance
func NewMockDBManager() *MockDBManager {
	return &MockDBManager{}
}
