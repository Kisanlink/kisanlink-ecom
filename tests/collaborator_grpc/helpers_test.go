package collaborator_grpc_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	collabModel "github.com/Kisanlink/kisanlink-ecom/entities/models/collaborator"
	"github.com/Kisanlink/kisanlink-ecom/internal/aaa"
	domainCollab "github.com/Kisanlink/kisanlink-ecom/internal/domain/collaborator"
	grpcHandler "github.com/Kisanlink/kisanlink-ecom/internal/grpc/handlers/collaborator"
	"github.com/Kisanlink/kisanlink-ecom/internal/saga"
	"github.com/Kisanlink/kisanlink-ecom/internal/services/gst"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// TestContext holds all test dependencies and mocks
type TestContext struct {
	DB              *gorm.DB
	Handler         *grpcHandler.Handler
	MockAAA         *MockAAAClient
	MockGST         *MockGSTService
	MockOTP         *MockOTPService
	MockSagaStorage *MockSagaStorage
	Logger          *logrus.Logger
}

// MockSagaStorage mocks the saga storage interface
type MockSagaStorage struct {
	mu    sync.RWMutex
	sagas map[string]*saga.Saga
}

func newMockSagaStorage() *MockSagaStorage {
	return &MockSagaStorage{
		sagas: make(map[string]*saga.Saga),
	}
}

func (m *MockSagaStorage) SaveState(_ context.Context, s *saga.Saga) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sagas[s.ID] = s
	return nil
}

func (m *MockSagaStorage) LoadState(_ context.Context, id string) (*saga.Saga, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, exists := m.sagas[id]
	if !exists {
		return nil, fmt.Errorf("saga not found: %s", id)
	}
	return s, nil
}

func (m *MockSagaStorage) ListPendingSagas(_ context.Context) ([]*saga.Saga, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*saga.Saga
	for _, s := range m.sagas {
		if s.State == saga.SagaStateRunning || s.State == saga.SagaStateCompensating {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *MockSagaStorage) CleanupCompleted(_ context.Context, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, s := range m.sagas {
		if s.State == saga.SagaStateCompleted || s.State == saga.SagaStateCompensated || s.State == saga.SagaStateFailed {
			delete(m.sagas, id)
		}
	}
	return nil
}

// MockAAAClient mocks the AAA service client
type MockAAAClient struct {
	mu                sync.RWMutex
	createAddressFunc func(ctx context.Context, req *aaa.CreateAddressRequest) (*aaa.Address, error)
	deleteAddressFunc func(ctx context.Context, id string) error
	getAddressFunc    func(ctx context.Context, id string) (*aaa.Address, error)
	updateAddressFunc func(ctx context.Context, id string, req *aaa.UpdateAddressRequest) (*aaa.Address, error)
	addresses         map[string]*aaa.Address
	nextAddressID     int
}

func newMockAAAClient() *MockAAAClient {
	return &MockAAAClient{
		addresses: make(map[string]*aaa.Address),
	}
}

func (m *MockAAAClient) CreateAddress(ctx context.Context, req *aaa.CreateAddressRequest) (*aaa.Address, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.createAddressFunc != nil {
		return m.createAddressFunc(ctx, req)
	}

	// Default behavior
	m.nextAddressID++
	addr := &aaa.Address{
		ID:         fmt.Sprintf("addr-%d", m.nextAddressID),
		EntityID:   req.EntityID,
		EntityType: req.EntityType,
		Line1:      req.Line1,
		Line2:      req.Line2,
		City:       req.City,
		State:      req.State,
		Country:    req.Country,
		PostalCode: req.PostalCode,
		Type:       req.Type,
		IsPrimary:  req.IsPrimary,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	m.addresses[addr.ID] = addr
	return addr, nil
}

func (m *MockAAAClient) DeleteAddress(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.deleteAddressFunc != nil {
		return m.deleteAddressFunc(ctx, id)
	}

	// Default behavior
	delete(m.addresses, id)
	return nil
}

func (m *MockAAAClient) GetAddress(ctx context.Context, id string) (*aaa.Address, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.getAddressFunc != nil {
		return m.getAddressFunc(ctx, id)
	}

	// Default behavior
	addr, exists := m.addresses[id]
	if !exists {
		return nil, fmt.Errorf("address not found: %s", id)
	}

	return addr, nil
}

func (m *MockAAAClient) UpdateAddress(ctx context.Context, id string, req *aaa.UpdateAddressRequest) (*aaa.Address, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.updateAddressFunc != nil {
		return m.updateAddressFunc(ctx, id, req)
	}

	// Default behavior
	addr, exists := m.addresses[id]
	if !exists {
		return nil, fmt.Errorf("address not found: %s", id)
	}

	// Update fields
	if req.Line1 != "" {
		addr.Line1 = req.Line1
	}
	if req.City != "" {
		addr.City = req.City
	}
	if req.State != "" {
		addr.State = req.State
	}
	if req.PostalCode != "" {
		addr.PostalCode = req.PostalCode
	}
	addr.UpdatedAt = time.Now()

	m.addresses[id] = addr
	return addr, nil
}

// SetCreateAddressFunc sets a custom create address function
func (m *MockAAAClient) SetCreateAddressFunc(f func(ctx context.Context, req *aaa.CreateAddressRequest) (*aaa.Address, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createAddressFunc = f
}

// SetDeleteAddressFunc sets a custom delete address function
func (m *MockAAAClient) SetDeleteAddressFunc(f func(ctx context.Context, id string) error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleteAddressFunc = f
}

// MockGSTService mocks the GST validation and locking service
type MockGSTService struct {
	mu                    sync.RWMutex
	validateFunc          func(gst string) (string, error)
	checkAndReserveFunc   func(ctx context.Context, gst string, fpoID uint64, requestID string) (bool, error)
	releaseFunc           func(ctx context.Context, gst string, fpoID uint64, requestID string) error
	reservedGSTs          map[string]string // gst -> requestID
	existingCollaborators map[string]uint64 // gst -> collaboratorID
}

func newMockGSTService() *MockGSTService {
	return &MockGSTService{
		reservedGSTs:          make(map[string]string),
		existingCollaborators: make(map[string]uint64),
	}
}

func (m *MockGSTService) ValidateAndNormalize(gst string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.validateFunc != nil {
		return m.validateFunc(gst)
	}

	// Default: basic validation
	if len(gst) != 15 {
		return "", fmt.Errorf("GST must be 15 characters long")
	}

	return gst, nil
}

func (m *MockGSTService) CheckAndReserveGST(ctx context.Context, gst string, fpoID uint64, requestID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.checkAndReserveFunc != nil {
		return m.checkAndReserveFunc(ctx, gst, fpoID, requestID)
	}

	// Check if GST already exists for a collaborator
	if _, exists := m.existingCollaborators[gst]; exists {
		return true, nil // GST exists
	}

	// Reserve the GST
	m.reservedGSTs[gst] = requestID
	return false, nil // GST doesn't exist
}

func (m *MockGSTService) ReleaseGST(ctx context.Context, gst string, fpoID uint64, requestID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.releaseFunc != nil {
		return m.releaseFunc(ctx, gst, fpoID, requestID)
	}

	delete(m.reservedGSTs, gst)
	return nil
}

// SetExistingCollaborator registers a GST as belonging to an existing collaborator
func (m *MockGSTService) SetExistingCollaborator(gst string, collabID uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.existingCollaborators[gst] = collabID
}

// SetValidateFunc sets a custom validation function
func (m *MockGSTService) SetValidateFunc(f func(gst string) (string, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.validateFunc = f
}

// MockOTPService mocks the OTP verification service
type MockOTPService struct {
	mu           sync.RWMutex
	generateFunc func(ctx context.Context, userID, purpose string) (string, error)
	validateFunc func(ctx context.Context, userID, purpose, otp string) error
	sendFunc     func(ctx context.Context, userID, otp string) error
	otps         map[string]string // key: userID+purpose -> otp
}

func newMockOTPService() *MockOTPService {
	return &MockOTPService{
		otps: make(map[string]string),
	}
}

func (m *MockOTPService) Generate(ctx context.Context, userID, purpose string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.generateFunc != nil {
		return m.generateFunc(ctx, userID, purpose)
	}

	// Default: generate simple OTP
	otp := "123456"
	key := userID + ":" + purpose
	m.otps[key] = otp
	return otp, nil
}

func (m *MockOTPService) Validate(ctx context.Context, userID, purpose, otp string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.validateFunc != nil {
		return m.validateFunc(ctx, userID, purpose, otp)
	}

	// Default: check if OTP matches
	key := userID + ":" + purpose
	expectedOTP, exists := m.otps[key]
	if !exists {
		return fmt.Errorf("no OTP found for user %s and purpose %s", userID, purpose)
	}

	if expectedOTP != otp {
		return fmt.Errorf("invalid OTP")
	}

	return nil
}

// SendOTP sends an OTP to the user (required by otp.Service interface)
func (m *MockOTPService) SendOTP(ctx context.Context, userID, otp string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.sendFunc != nil {
		return m.sendFunc(ctx, userID, otp)
	}

	// Default: mock success
	return nil
}

// SetValidateFunc sets a custom validation function
func (m *MockOTPService) SetValidateFunc(f func(ctx context.Context, userID, purpose, otp string) error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.validateFunc = f
}

// SetSendFunc sets a custom send function
func (m *MockOTPService) SetSendFunc(f func(ctx context.Context, userID, otp string) error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sendFunc = f
}

// setupTestContext creates a new test context with in-memory database and mocks
func setupTestContext(t *testing.T) *TestContext {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Run migrations
	err = db.AutoMigrate(&collabModel.Collaborator{})
	require.NoError(t, err)

	// Create mock services
	mockAAA := newMockAAAClient()
	mockGST := newMockGSTService()
	mockOTP := newMockOTPService()
	mockSagaStorage := newMockSagaStorage()

	// Create logger (handler uses logrus, but saga uses zap)
	testLogger := logrus.New()
	testLogger.SetLevel(logrus.WarnLevel)

	// Create zap logger for saga
	zapLogger := zap.NewNop() // Use no-op logger for tests

	// Create saga metrics
	sagaMetrics := saga.NewSagaMetrics()

	// Create saga executor with correct parameters
	sagaExecutor := saga.NewSagaExecutor(mockSagaStorage, zapLogger, sagaMetrics)

	// Create actual state machine instance
	stateMachine := domainCollab.NewStateMachine(db, zapLogger)

	// Create mock Redis client for GST service (use nil client for tests)
	// GST service will handle nil redis gracefully or we mock it
	var redisClient redis.UniversalClient

	// Create GST service - for now use actual service with mock behavior via mockGST
	// The handler expects *gst.Service, but we control behavior via mockGST in tests
	gstService := gst.NewService(db, redisClient, testLogger)

	// AAA Client - handler expects *aaa.Client
	// We need to create a wrapper or use interface approach
	// For now, we'll create a nil client and the handler should handle gracefully
	// OR we create an interface wrapper
	var aaaClient *aaa.Client // This needs architectural fix

	// Create handler
	handler := grpcHandler.NewHandler(
		db,
		testLogger,
		aaaClient, // TODO: Need interface wrapper for proper mocking
		gstService,
		sagaExecutor,
		mockOTP, // This satisfies otp.Service interface
		stateMachine,
	)

	return &TestContext{
		DB:              db,
		Handler:         handler,
		MockAAA:         mockAAA,
		MockGST:         mockGST,
		MockOTP:         mockOTP,
		MockSagaStorage: mockSagaStorage,
		Logger:          testLogger,
	}
}

// createTestCollaborator creates a test collaborator directly in the database
func createTestCollaborator(db *gorm.DB, userID, gst string, fpoID uint64) *collabModel.Collaborator {
	collab := &collabModel.Collaborator{
		UserID:           userID,
		Username:         "testuser",
		Email:            "test@example.com",
		FirstName:        "Test",
		LastName:         "User",
		CollaboratorType: collabModel.CollaboratorTypeSupplier,
		Status:           collabModel.CollaboratorStatusActive,
		TaxID:            gst,
		OrganizationID:   fmt.Sprintf("%d", fpoID),
		BusinessName:     "Test Business",
		BusinessType:     "MANUFACTURER",
		BusinessPhone:    "+911234567890",
		BusinessEmail:    "business@example.com",
		BusinessWebsite:  "https://example.com",
		BusinessLicense:  "LIC123456",
		BusinessAddress:  "123 Test St",
		Phone:            "+911234567890",
		Bio:              "Test bio",
		Avatar:           "https://example.com/avatar.jpg",
	}

	db.Create(collab)
	return collab
}

// testUserContext creates a user context for testing
func testUserContext(userID string, fpoID uint64, roles ...string) context.Context {
	ctx := context.Background()
	uc := &grpcHandler.UserContext{
		UserID:   userID,
		FpoID:    fpoID,
		Roles:    roles,
		TenantID: "test-tenant",
		Email:    "user@example.com",
	}
	return grpcHandler.WithUserContext(ctx, uc)
}

// cleanupTestDB cleans up the test database
func cleanupTestDB(_ *testing.T, db *gorm.DB) {
	// Clean up all tables
	db.Exec("DELETE FROM collaborators")
	db.Exec("DELETE FROM orders")
	db.Exec("DELETE FROM audit_logs")
}
