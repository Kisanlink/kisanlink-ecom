package inventory_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	catalogModels "github.com/Kisanlink/kisanlink-ecom/entities/models/catalog"
	"github.com/Kisanlink/kisanlink-ecom/internal/handlers/inventory"
	inventoryRepo "github.com/Kisanlink/kisanlink-ecom/internal/repositories/inventory"
	inventoryService "github.com/Kisanlink/kisanlink-ecom/internal/services/inventory"
	"github.com/Kisanlink/kisanlink-ecom/tests/data"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockInventoryService is a mock implementation of InventoryService
type MockInventoryService struct {
	mock.Mock
}

func (m *MockInventoryService) CreateInventoryLot(ctx context.Context, req *inventoryService.CreateInventoryLotRequest, userID, orgID string) (*catalogModels.InventoryLot, error) {
	args := m.Called(ctx, req, userID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*catalogModels.InventoryLot), args.Error(1)
}

func (m *MockInventoryService) GetInventoryLot(ctx context.Context, lotID, userID, orgID string) (*catalogModels.InventoryLot, error) {
	args := m.Called(ctx, lotID, userID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*catalogModels.InventoryLot), args.Error(1)
}

func (m *MockInventoryService) ListInventoryLots(ctx context.Context, filter *inventoryService.InventoryFilter, userID, orgID string) (*inventoryService.InventoryListResponse, error) {
	args := m.Called(ctx, filter, userID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*inventoryService.InventoryListResponse), args.Error(1)
}

func (m *MockInventoryService) UpdateInventoryLot(ctx context.Context, lotID string, req *inventoryService.UpdateInventoryLotRequest, userID, orgID string) (*catalogModels.InventoryLot, error) {
	args := m.Called(ctx, lotID, req, userID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*catalogModels.InventoryLot), args.Error(1)
}

func (m *MockInventoryService) DeleteInventoryLot(ctx context.Context, lotID, userID, orgID string) error {
	args := m.Called(ctx, lotID, userID, orgID)
	return args.Error(0)
}

func (m *MockInventoryService) ReserveInventory(ctx context.Context, catalogItemID string, quantity decimal.Decimal, userID, orgID string) ([]*catalogModels.InventoryLot, error) {
	args := m.Called(ctx, catalogItemID, quantity, userID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*catalogModels.InventoryLot), args.Error(1)
}

func (m *MockInventoryService) ReleaseInventory(ctx context.Context, catalogItemID string, quantity decimal.Decimal, userID, orgID string) error {
	args := m.Called(ctx, catalogItemID, quantity, userID, orgID)
	return args.Error(0)
}

func (m *MockInventoryService) SellInventory(ctx context.Context, catalogItemID string, quantity decimal.Decimal, userID, orgID string) error {
	args := m.Called(ctx, catalogItemID, quantity, userID, orgID)
	return args.Error(0)
}

func (m *MockInventoryService) AdjustInventory(ctx context.Context, lotID string, adjustment decimal.Decimal, reason, userID, orgID string) (*catalogModels.InventoryLot, error) {
	args := m.Called(ctx, lotID, adjustment, reason, userID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*catalogModels.InventoryLot), args.Error(1)
}

func (m *MockInventoryService) GetAvailableQuantity(ctx context.Context, catalogItemID, userID, orgID string) (decimal.Decimal, error) {
	args := m.Called(ctx, catalogItemID, userID, orgID)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockInventoryService) GetTotalQuantity(ctx context.Context, catalogItemID, userID, orgID string) (decimal.Decimal, error) {
	args := m.Called(ctx, catalogItemID, userID, orgID)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockInventoryService) CheckAvailability(ctx context.Context, catalogItemID string, requiredQuantity decimal.Decimal, userID, orgID string) (bool, decimal.Decimal, error) {
	args := m.Called(ctx, catalogItemID, requiredQuantity, userID, orgID)
	return args.Bool(0), args.Get(1).(decimal.Decimal), args.Error(2)
}

func (m *MockInventoryService) ProcessExpiringLots(ctx context.Context, orgID string) ([]*catalogModels.InventoryLot, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*catalogModels.InventoryLot), args.Error(1)
}

func (m *MockInventoryService) MarkLotExpired(ctx context.Context, lotID, userID, orgID string) error {
	args := m.Called(ctx, lotID, userID, orgID)
	return args.Error(0)
}

func (m *MockInventoryService) GetAuditTrail(ctx context.Context, lotID string, offset, limit int, userID, orgID string) (*inventoryService.AuditTrailResponse, error) {
	args := m.Called(ctx, lotID, offset, limit, userID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*inventoryService.AuditTrailResponse), args.Error(1)
}

// InventoryHandlerTestSuite defines the test suite for inventory handlers
type InventoryHandlerTestSuite struct {
	suite.Suite
	handler     *inventory.InventoryHandler
	mockService *MockInventoryService
	router      *gin.Engine
	testUserID  string
	testOrgID   string
	testLot     *catalogModels.InventoryLot
}

// SetupSuite sets up the test suite
func (suite *InventoryHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.testUserID = "test-user-123"
	suite.testOrgID = "test-org-456"
}

// SetupTest sets up each test
func (suite *InventoryHandlerTestSuite) SetupTest() {
	suite.mockService = new(MockInventoryService)
	suite.handler = inventory.NewInventoryHandler(suite.mockService)
	suite.router = gin.New()

	// Add middleware to set user context
	suite.router.Use(func(c *gin.Context) {
		c.Set("subjectID", suite.testUserID)
		c.Set("organizationID", suite.testOrgID)
		c.Next()
	})

	// Setup test data
	suite.testLot = data.GetTestInventoryLot()
}

// TearDownTest cleans up after each test
func (suite *InventoryHandlerTestSuite) TearDownTest() {
	suite.mockService.AssertExpectations(suite.T())
}

// TestCreateInventoryLot_Success tests successful inventory lot creation
func (suite *InventoryHandlerTestSuite) TestCreateInventoryLot_Success() {
	// Setup route
	suite.router.POST("/inventory/lots", suite.handler.CreateInventoryLot)

	// Setup request
	requestBody := inventory.CreateInventoryLotRequest{
		CatalogItemID:     "prod_123",
		LotNumber:         "LOT-2024-001",
		BatchNumber:       "BATCH-001",
		InitialQuantity:   100.5,
		QualityGrade:      "A",
		HarvestDate:       "2024-01-15",
		ExpiryDate:        "2024-12-31",
		WarehouseLocation: "Warehouse A",
		StorageConditions: "Temperature controlled",
		LotPrice:          25.50,
		Metadata:          `{"supplier": "Farm ABC"}`,
	}

	// Setup mock expectations
	suite.mockService.On("CreateInventoryLot", mock.Anything, mock.MatchedBy(func(req *inventoryService.CreateInventoryLotRequest) bool {
		return req.CatalogItemID == "prod_123" &&
			req.LotNumber == "LOT-2024-001" &&
			req.InitialQuantity.Equal(decimal.NewFromFloat(100.5))
	}), suite.testUserID, suite.testOrgID).Return(suite.testLot, nil)

	// Make request
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/inventory/lots", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Inventory lot created successfully", response["message"])
	assert.NotNil(suite.T(), response["data"])
}

// TestCreateInventoryLot_ValidationError tests validation error handling
func (suite *InventoryHandlerTestSuite) TestCreateInventoryLot_ValidationError() {
	// Setup route
	suite.router.POST("/inventory/lots", suite.handler.CreateInventoryLot)

	// Setup invalid request (missing required fields)
	requestBody := inventory.CreateInventoryLotRequest{
		// Missing CatalogItemID and LotNumber
		InitialQuantity: -10, // Invalid quantity
	}

	// Make request
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/inventory/lots", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response["success"].(bool))
}

// TestCreateInventoryLot_ServiceError tests service error handling
func (suite *InventoryHandlerTestSuite) TestCreateInventoryLot_ServiceError() {
	// Setup route
	suite.router.POST("/inventory/lots", suite.handler.CreateInventoryLot)

	// Setup request
	requestBody := inventory.CreateInventoryLotRequest{
		CatalogItemID:   "prod_123",
		LotNumber:       "LOT-2024-001",
		InitialQuantity: 100.5,
	}

	// Setup mock expectations
	suite.mockService.On("CreateInventoryLot", mock.Anything, mock.Anything, suite.testUserID, suite.testOrgID).
		Return(nil, fmt.Errorf("catalog item not found"))

	// Make request
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/inventory/lots", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response["error"].(map[string]interface{})["message"], "catalog item not found")
}

// TestListInventoryLots_Success tests successful inventory lot listing
func (suite *InventoryHandlerTestSuite) TestListInventoryLots_Success() {
	// Setup route
	suite.router.GET("/inventory/lots", suite.handler.ListInventoryLots)

	// Setup mock response
	mockResponse := &inventoryService.InventoryListResponse{
		Lots:   []*catalogModels.InventoryLot{suite.testLot},
		Total:  1,
		Offset: 0,
		Limit:  20,
	}

	// Setup mock expectations
	suite.mockService.On("ListInventoryLots", mock.Anything, mock.MatchedBy(func(filter *inventoryService.InventoryFilter) bool {
		return filter.Offset == 0 && filter.Limit == 20
	}), suite.testUserID, suite.testOrgID).Return(mockResponse, nil)

	// Make request
	req, _ := http.NewRequest("GET", "/inventory/lots?offset=0&limit=20", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Inventory lots retrieved successfully", response["message"])
	assert.NotNil(suite.T(), response["data"])
}

// TestListInventoryLots_WithFilters tests inventory lot listing with filters
func (suite *InventoryHandlerTestSuite) TestListInventoryLots_WithFilters() {
	// Setup route
	suite.router.GET("/inventory/lots", suite.handler.ListInventoryLots)

	// Setup mock response
	mockResponse := &inventoryService.InventoryListResponse{
		Lots:   []*catalogModels.InventoryLot{suite.testLot},
		Total:  1,
		Offset: 0,
		Limit:  20,
	}

	// Setup mock expectations
	suite.mockService.On("ListInventoryLots", mock.Anything, mock.MatchedBy(func(filter *inventoryService.InventoryFilter) bool {
		return filter.CatalogItemID == "prod_123" &&
			filter.Status == catalogModels.LotStatusActive &&
			filter.QualityGrade == "A"
	}), suite.testUserID, suite.testOrgID).Return(mockResponse, nil)

	// Make request with filters
	req, _ := http.NewRequest("GET", "/inventory/lots?catalog_item_id=prod_123&status=active&quality_grade=A", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
}

// TestGetInventoryLot_Success tests successful inventory lot retrieval
func (suite *InventoryHandlerTestSuite) TestGetInventoryLot_Success() {
	// Setup route
	suite.router.GET("/inventory/lots/:id", suite.handler.GetInventoryLot)

	// Setup mock expectations
	suite.mockService.On("GetInventoryLot", mock.Anything, "lot_123", suite.testUserID, suite.testOrgID).
		Return(suite.testLot, nil)

	// Make request
	req, _ := http.NewRequest("GET", "/inventory/lots/lot_123", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Inventory lot retrieved successfully", response["message"])
	assert.NotNil(suite.T(), response["data"])
}

// TestGetInventoryLot_NotFound tests inventory lot not found
func (suite *InventoryHandlerTestSuite) TestGetInventoryLot_NotFound() {
	// Setup route
	suite.router.GET("/inventory/lots/:id", suite.handler.GetInventoryLot)

	// Setup mock expectations
	suite.mockService.On("GetInventoryLot", mock.Anything, "nonexistent", suite.testUserID, suite.testOrgID).
		Return(nil, fmt.Errorf("inventory lot not found"))

	// Make request
	req, _ := http.NewRequest("GET", "/inventory/lots/nonexistent", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response["success"].(bool))
}

// TestUpdateInventoryLot_Success tests successful inventory lot update
func (suite *InventoryHandlerTestSuite) TestUpdateInventoryLot_Success() {
	// Setup route
	suite.router.PATCH("/inventory/lots/:id", suite.handler.UpdateInventoryLot)

	// Setup request
	qualityGrade := "A+"
	requestBody := inventory.UpdateInventoryLotRequest{
		QualityGrade: &qualityGrade,
	}

	// Setup mock expectations
	suite.mockService.On("UpdateInventoryLot", mock.Anything, "lot_123", mock.MatchedBy(func(req *inventoryService.UpdateInventoryLotRequest) bool {
		return req.QualityGrade != nil && *req.QualityGrade == "A+"
	}), suite.testUserID, suite.testOrgID).Return(suite.testLot, nil)

	// Make request
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("PATCH", "/inventory/lots/lot_123", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Inventory lot updated successfully", response["message"])
}

// TestAdjustInventoryQuantity_Success tests successful inventory quantity adjustment
func (suite *InventoryHandlerTestSuite) TestAdjustInventoryQuantity_Success() {
	// Setup route
	suite.router.PATCH("/inventory/lots/:id/adjust", suite.handler.AdjustInventoryQuantity)

	// Setup request
	requestBody := inventory.AdjustInventoryRequest{
		Adjustment: decimal.NewFromFloat(10.5),
		Reason:     "Stock correction after physical count",
	}

	// Setup mock expectations
	suite.mockService.On("AdjustInventory", mock.Anything, "lot_123", decimal.NewFromFloat(10.5), "Stock correction after physical count", suite.testUserID, suite.testOrgID).
		Return(suite.testLot, nil)

	// Make request
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("PATCH", "/inventory/lots/lot_123/adjust", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Inventory adjusted successfully", response["message"])
}

// TestCheckInventoryAvailability_Success tests successful inventory availability check
func (suite *InventoryHandlerTestSuite) TestCheckInventoryAvailability_Success() {
	// Setup route
	suite.router.GET("/inventory/availability/:catalog_item_id", suite.handler.CheckInventoryAvailability)

	// Setup mock expectations
	suite.mockService.On("CheckAvailability", mock.Anything, "prod_123", mock.AnythingOfType("decimal.Decimal"), suite.testUserID, suite.testOrgID).
		Return(true, decimal.NewFromFloat(100.0), nil)
	suite.mockService.On("GetTotalQuantity", mock.Anything, "prod_123", suite.testUserID, suite.testOrgID).
		Return(decimal.NewFromFloat(150.0), nil)

	// Make request
	req, _ := http.NewRequest("GET", "/inventory/availability/prod_123?required_quantity=5.0", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Inventory availability checked successfully", response["message"])

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), "prod_123", data["catalog_item_id"])
	assert.True(suite.T(), data["available"].(bool))
}

// TestGetInventoryAuditTrail_Success tests successful audit trail retrieval
func (suite *InventoryHandlerTestSuite) TestGetInventoryAuditTrail_Success() {
	// Setup route
	suite.router.GET("/inventory/lots/:id/audit", suite.handler.GetInventoryAuditTrail)

	// Setup mock response
	mockResponse := &inventoryService.AuditTrailResponse{
		Logs:   []*inventoryRepo.InventoryAuditLog{},
		Total:  0,
		Offset: 0,
		Limit:  20,
	}

	// Setup mock expectations
	suite.mockService.On("GetAuditTrail", mock.Anything, "lot_123", 0, 20, suite.testUserID, suite.testOrgID).
		Return(mockResponse, nil)

	// Make request
	req, _ := http.NewRequest("GET", "/inventory/lots/lot_123/audit", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Audit trail retrieved successfully", response["message"])
}

// TestInventoryHandler_NoAuth tests handlers without authentication
func (suite *InventoryHandlerTestSuite) TestInventoryHandler_NoAuth() {
	// Setup router without auth middleware
	router := gin.New()
	router.POST("/inventory/lots", suite.handler.CreateInventoryLot)

	// Make request without auth context
	requestBody := inventory.CreateInventoryLotRequest{
		CatalogItemID:   "prod_123",
		LotNumber:       "LOT-2024-001",
		InitialQuantity: 100.5,
	}

	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/inventory/lots", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response["error"].(map[string]interface{})["message"], "User not authenticated")
}

// TestInventoryHandler_NoOrgContext tests handlers without organization context
func (suite *InventoryHandlerTestSuite) TestInventoryHandler_NoOrgContext() {
	// Setup router with only user auth, no org context
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("subjectID", suite.testUserID)
		// No organizationID set
		c.Next()
	})
	router.POST("/inventory/lots", suite.handler.CreateInventoryLot)

	// Make request
	requestBody := inventory.CreateInventoryLotRequest{
		CatalogItemID:   "prod_123",
		LotNumber:       "LOT-2024-001",
		InitialQuantity: 100.5,
	}

	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/inventory/lots", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response["error"].(map[string]interface{})["message"], "Organization context required")
}

// Run the test suite
func TestInventoryHandlerSuite(t *testing.T) {
	suite.Run(t, new(InventoryHandlerTestSuite))
}
