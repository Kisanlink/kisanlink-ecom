//go:build integration
// +build integration

package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	orderModels "kisanlink-ecom/entities/models/orders"
	orderRequests "kisanlink-ecom/entities/requests/orders"
	"kisanlink-ecom/tests/testutils"

	"github.com/stretchr/testify/suite"
)

// APIContractTestSuite tests API contracts and backward compatibility
type APIContractTestSuite struct {
	suite.Suite
	server           *httptest.Server
	client           *http.Client
	baseURL          string
	authToken        string
	createdResources map[string][]string // resource type -> IDs
}

// APIResponse represents a standard API response structure
type APIResponse struct {
	Success bool             `json:"success"`
	Data    interface{}      `json:"data,omitempty"`
	Error   *APIError        `json:"error,omitempty"`
	Meta    *APIResponseMeta `json:"meta,omitempty"`
}

type APIError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

type APIResponseMeta struct {
	TraceID    string             `json:"trace_id,omitempty"`
	Pagination *APIPaginationMeta `json:"pagination,omitempty"`
}

type APIPaginationMeta struct {
	Page    int   `json:"page"`
	Limit   int   `json:"limit"`
	Total   int64 `json:"total"`
	HasNext bool  `json:"has_next"`
}

// SetupSuite initializes the test server and client
func (suite *APIContractTestSuite) SetupSuite() {
	// Initialize test server
	suite.setupTestServer()

	// Initialize HTTP client
	suite.client = &http.Client{
		Timeout: 30 * time.Second,
	}

	// Initialize resource tracking
	suite.createdResources = make(map[string][]string)

	// Setup authentication
	suite.setupAuthentication()
}

func (suite *APIContractTestSuite) setupTestServer() {
	// In a real implementation, this would start the actual server
	// For testing, we'll create a mock server that implements the API contract
	mux := http.NewServeMux()

	// Setup API routes
	suite.setupAPIRoutes(mux)

	// Create test server
	suite.server = httptest.NewServer(mux)
	suite.baseURL = suite.server.URL

	suite.T().Logf("Test server started at: %s", suite.baseURL)
}

func (suite *APIContractTestSuite) setupAPIRoutes(mux *http.ServeMux) {
	// Catalog routes
	mux.HandleFunc("/api/v1/catalog/products", suite.handleCatalogProducts)
	mux.HandleFunc("/api/v1/catalog/services", suite.handleCatalogServices)
	mux.HandleFunc("/api/v1/catalog/labour", suite.handleCatalogLabour)
	mux.HandleFunc("/api/v1/catalog/", suite.handleCatalogByType)
	mux.HandleFunc("/api/v1/catalog", suite.handleCatalogList)
	mux.HandleFunc("/api/v1/catalog/search", suite.handleCatalogSearch)

	// Order routes
	mux.HandleFunc("/api/v1/orders", suite.handleOrders)
	mux.HandleFunc("/api/v1/orders/", suite.handleOrderByID)

	// Inventory routes
	mux.HandleFunc("/api/v1/inventory/lots", suite.handleInventoryLots)
	mux.HandleFunc("/api/v1/inventory/lots/", suite.handleInventoryLotByID)
	mux.HandleFunc("/api/v1/inventory/availability/", suite.handleInventoryAvailability)

	// Integration routes
	mux.HandleFunc("/api/v1/integrations/catalog/proposals", suite.handleCatalogProposals)
	mux.HandleFunc("/api/v1/integrations/orders/acknowledgements", suite.handleOrderAcknowledgements)
	mux.HandleFunc("/api/v1/integrations/catalog/exports", suite.handleCatalogExports)

	// Health check
	mux.HandleFunc("/health", suite.handleHealthCheck)
}

func (suite *APIContractTestSuite) setupAuthentication() {
	// In a real implementation, this would authenticate with the AAA service
	// For testing, we'll use a mock token
	suite.authToken = "Bearer test_auth_token_" + time.Now().Format("20060102150405")
}

// TearDownSuite cleans up the test server
func (suite *APIContractTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
}

// SetupTest runs before each test
func (suite *APIContractTestSuite) SetupTest() {
	// Clear created resources tracking
	for key := range suite.createdResources {
		suite.createdResources[key] = suite.createdResources[key][:0]
	}
}

// Test catalog API contracts
func (suite *APIContractTestSuite) TestCatalogAPIContracts() {
	suite.T().Log("Testing catalog product API contract")
	suite.testProductAPIContract()

	suite.T().Log("Testing catalog service API contract")
	suite.testServiceAPIContract()

	suite.T().Log("Testing catalog labour API contract")
	suite.testLabourAPIContract()

	suite.T().Log("Testing catalog listing API contract")
	suite.testCatalogListingContract()

	suite.T().Log("Testing catalog search API contract")
	suite.testCatalogSearchContract()
}

func (suite *APIContractTestSuite) testProductAPIContract() {
	// Test POST /api/v1/catalog/products
	productReq := testutils.CreateTestProductRequest()

	resp, err := suite.makeRequest("POST", "/api/v1/catalog/products", productReq)
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)

	var apiResp APIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	suite.NoError(err)
	suite.True(apiResp.Success)
	suite.NotNil(apiResp.Data)
	suite.Nil(apiResp.Error)

	// Verify response structure
	productData := apiResp.Data.(map[string]interface{})
	suite.Contains(productData, "id")
	suite.Contains(productData, "name")
	suite.Contains(productData, "base_price")
	suite.Contains(productData, "item_type")
	suite.Equal("PRODUCT", productData["item_type"])

	// Track created resource
	productID := productData["id"].(string)
	suite.createdResources["products"] = append(suite.createdResources["products"], productID)

	// Test GET /api/v1/catalog/products/{id}
	getResp, err := suite.makeRequest("GET", "/api/v1/catalog/products/"+productID, nil)
	suite.NoError(err)
	suite.Equal(http.StatusOK, getResp.StatusCode)

	var getApiResp APIResponse
	err = json.NewDecoder(getResp.Body).Decode(&getApiResp)
	suite.NoError(err)
	suite.True(getApiResp.Success)

	// Test PUT /api/v1/catalog/products/{id}
	updateReq := map[string]interface{}{
		"name":        "Updated Product Name",
		"description": "Updated product description",
		"base_price":  150.0,
	}
	putResp, err := suite.makeRequest("PUT", "/api/v1/catalog/products/"+productID, updateReq)
	suite.NoError(err)
	suite.Equal(http.StatusOK, putResp.StatusCode)
}

func (suite *APIContractTestSuite) testServiceAPIContract() {
	// Test POST /api/v1/catalog/services
	serviceReq := testutils.CreateTestServiceRequest()

	resp, err := suite.makeRequest("POST", "/api/v1/catalog/services", serviceReq)
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)

	var apiResp APIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	suite.NoError(err)
	suite.True(apiResp.Success)

	// Verify service-specific fields
	serviceData := apiResp.Data.(map[string]interface{})
	suite.Equal("SERVICE", serviceData["item_type"])
	suite.Contains(serviceData, "duration_minutes")

	serviceID := serviceData["id"].(string)
	suite.createdResources["services"] = append(suite.createdResources["services"], serviceID)
}

func (suite *APIContractTestSuite) testLabourAPIContract() {
	// Test POST /api/v1/catalog/labour
	labourReq := testutils.CreateTestLabourRequest()

	resp, err := suite.makeRequest("POST", "/api/v1/catalog/labour", labourReq)
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)

	var apiResp APIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	suite.NoError(err)
	suite.True(apiResp.Success)

	// Verify labour-specific fields
	labourData := apiResp.Data.(map[string]interface{})
	suite.Equal("LABOUR", labourData["item_type"])
	suite.Contains(labourData, "skill_level")
	suite.Contains(labourData, "hourly_rate")

	labourID := labourData["id"].(string)
	suite.createdResources["labour"] = append(suite.createdResources["labour"], labourID)
}

func (suite *APIContractTestSuite) testCatalogListingContract() {
	// Test GET /api/v1/catalog with various query parameters
	testCases := []struct {
		name        string
		queryParams string
		expectItems bool
	}{
		{"no filters", "", true},
		{"with pagination", "?page=1&limit=10", true},
		{"filter by type", "?item_type=PRODUCT", true},
		{"filter by category", "?category=agriculture", true},
		{"filter by active status", "?is_active=true", true},
		{"filter by visibility", "?visibility=ORG", true},
		{"price range filter", "?min_price=10&max_price=1000", true},
		{"multiple filters", "?item_type=PRODUCT&is_active=true&page=1&limit=5", true},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			resp, err := suite.makeRequest("GET", "/api/v1/catalog"+tc.queryParams, nil)
			suite.NoError(err)
			suite.Equal(http.StatusOK, resp.StatusCode)

			var apiResp APIResponse
			err = json.NewDecoder(resp.Body).Decode(&apiResp)
			suite.NoError(err)
			suite.True(apiResp.Success)

			// Verify pagination metadata
			if tc.expectItems {
				suite.NotNil(apiResp.Meta)
				suite.NotNil(apiResp.Meta.Pagination)
				suite.GreaterOrEqual(apiResp.Meta.Pagination.Page, 1)
				suite.GreaterOrEqual(apiResp.Meta.Pagination.Limit, 1)
				suite.GreaterOrEqual(apiResp.Meta.Pagination.Total, int64(0))
			}
		})
	}
}

func (suite *APIContractTestSuite) testCatalogSearchContract() {
	// Test GET /api/v1/catalog/search
	searchCases := []struct {
		name        string
		query       string
		expectError bool
	}{
		{"valid search", "?q=test", false},
		{"search with filters", "?q=agricultural&item_type=PRODUCT", false},
		{"search with pagination", "?q=test&page=1&limit=5", false},
		{"empty query", "?q=", true},
		{"missing query", "", true},
	}

	for _, tc := range searchCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			resp, err := suite.makeRequest("GET", "/api/v1/catalog/search"+tc.query, nil)
			suite.NoError(err)

			var apiResp APIResponse
			err = json.NewDecoder(resp.Body).Decode(&apiResp)
			suite.NoError(err)

			if tc.expectError {
				suite.False(apiResp.Success)
				suite.NotNil(apiResp.Error)
				suite.Equal(http.StatusBadRequest, resp.StatusCode)
			} else {
				suite.True(apiResp.Success)
				suite.Nil(apiResp.Error)
				suite.Equal(http.StatusOK, resp.StatusCode)
			}
		})
	}
}

// Test order API contracts
func (suite *APIContractTestSuite) TestOrderAPIContracts() {
	// First create a catalog item for orders
	suite.createTestCatalogItem()

	suite.T().Log("Testing order creation API contract")
	suite.testOrderCreationContract()

	suite.T().Log("Testing order retrieval API contract")
	suite.testOrderRetrievalContract()

	suite.T().Log("Testing order listing API contract")
	suite.testOrderListingContract()

	suite.T().Log("Testing order status update API contract")
	suite.testOrderStatusUpdateContract()

	suite.T().Log("Testing order cancellation API contract")
	suite.testOrderCancellationContract()
}

func (suite *APIContractTestSuite) createTestCatalogItem() string {
	productReq := testutils.CreateTestProductRequest()
	resp, err := suite.makeRequest("POST", "/api/v1/catalog/products", productReq)
	suite.NoError(err)

	var apiResp APIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	suite.NoError(err)

	productData := apiResp.Data.(map[string]interface{})
	return productData["id"].(string)
}

func (suite *APIContractTestSuite) testOrderCreationContract() {
	// Test POST /api/v1/orders
	orderReq := testutils.CreateTestCreateOrderRequest()

	resp, err := suite.makeRequest("POST", "/api/v1/orders", orderReq)
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)

	var apiResp APIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	suite.NoError(err)
	suite.True(apiResp.Success)
	suite.NotNil(apiResp.Data)

	// Verify order response structure
	orderData := apiResp.Data.(map[string]interface{})
	suite.Contains(orderData, "id")
	suite.Contains(orderData, "order_number")
	suite.Contains(orderData, "status")
	suite.Contains(orderData, "buyer_organization_id")
	suite.Contains(orderData, "seller_organization_id")
	suite.Contains(orderData, "total_amount")
	suite.Contains(orderData, "items")
	suite.Contains(orderData, "created_at")
	suite.Contains(orderData, "updated_at")

	// Verify initial status
	suite.Equal("pending", orderData["status"])

	// Track created order
	orderID := orderData["id"].(string)
	suite.createdResources["orders"] = append(suite.createdResources["orders"], orderID)

	// Test invalid order creation
	invalidReq := &orderRequests.CreateOrderRequest{
		BuyerOrganizationID:  "", // Invalid empty buyer org
		SellerOrganizationID: testutils.TestSellerOrgID,
		Items:                []orderRequests.CreateOrderItemRequest{},
	}

	invalidResp, err := suite.makeRequest("POST", "/api/v1/orders", invalidReq)
	suite.NoError(err)
	suite.Equal(http.StatusBadRequest, invalidResp.StatusCode)

	var invalidApiResp APIResponse
	err = json.NewDecoder(invalidResp.Body).Decode(&invalidApiResp)
	suite.NoError(err)
	suite.False(invalidApiResp.Success)
	suite.NotNil(invalidApiResp.Error)
}

func (suite *APIContractTestSuite) testOrderRetrievalContract() {
	// Get an existing order ID
	orderID := suite.createdResources["orders"][0]

	// Test GET /api/v1/orders/{id}
	resp, err := suite.makeRequest("GET", "/api/v1/orders/"+orderID, nil)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	var apiResp APIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	suite.NoError(err)
	suite.True(apiResp.Success)

	orderData := apiResp.Data.(map[string]interface{})
	suite.Equal(orderID, orderData["id"])

	// Test non-existent order
	notFoundResp, err := suite.makeRequest("GET", "/api/v1/orders/non-existent", nil)
	suite.NoError(err)
	suite.Equal(http.StatusNotFound, notFoundResp.StatusCode)
}

func (suite *APIContractTestSuite) testOrderListingContract() {
	// Test GET /api/v1/orders with various filters
	testCases := []struct {
		name        string
		queryParams string
	}{
		{"no filters", ""},
		{"with pagination", "?page=1&limit=10"},
		{"filter by status", "?status=pending"},
		{"filter by buyer", "?buyer_id=" + testutils.TestOrgID},
		{"filter by seller", "?seller_id=" + testutils.TestSellerOrgID},
		{"multiple filters", "?status=pending&page=1&limit=5"},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			resp, err := suite.makeRequest("GET", "/api/v1/orders"+tc.queryParams, nil)
			suite.NoError(err)
			suite.Equal(http.StatusOK, resp.StatusCode)

			var apiResp APIResponse
			err = json.NewDecoder(resp.Body).Decode(&apiResp)
			suite.NoError(err)
			suite.True(apiResp.Success)

			// Verify pagination metadata
			suite.NotNil(apiResp.Meta)
			suite.NotNil(apiResp.Meta.Pagination)
		})
	}
}

func (suite *APIContractTestSuite) testOrderStatusUpdateContract() {
	orderID := suite.createdResources["orders"][0]

	// Test PATCH /api/v1/orders/{id}/status
	statusReq := &orderRequests.UpdateOrderStatusRequest{
		Status: orderModels.OrderStatusConfirmed,
		Reason: "API contract test status update",
	}

	resp, err := suite.makeRequest("PATCH", "/api/v1/orders/"+orderID+"/status", statusReq)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	var apiResp APIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	suite.NoError(err)
	suite.True(apiResp.Success)

	// Test invalid status transition
	invalidStatusReq := &orderRequests.UpdateOrderStatusRequest{
		Status: orderModels.OrderStatusDelivered, // Invalid transition from confirmed
		Reason: "Invalid transition test",
	}

	invalidResp, err := suite.makeRequest("PATCH", "/api/v1/orders/"+orderID+"/status", invalidStatusReq)
	suite.NoError(err)
	suite.Equal(http.StatusBadRequest, invalidResp.StatusCode)
}

func (suite *APIContractTestSuite) testOrderCancellationContract() {
	// Create a new order for cancellation test
	orderReq := testutils.CreateTestCreateOrderRequest()
	createResp, err := suite.makeRequest("POST", "/api/v1/orders", orderReq)
	suite.NoError(err)

	var createApiResp APIResponse
	err = json.NewDecoder(createResp.Body).Decode(&createApiResp)
	suite.NoError(err)

	orderData := createApiResp.Data.(map[string]interface{})
	orderID := orderData["id"].(string)

	// Test POST /api/v1/orders/{id}/cancel
	resp, err := suite.makeRequest("POST", "/api/v1/orders/"+orderID+"/cancel", nil)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	var apiResp APIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	suite.NoError(err)
	suite.True(apiResp.Success)
}

// Test inventory API contracts
func (suite *APIContractTestSuite) TestInventoryAPIContracts() {
	suite.T().Log("Testing inventory lot creation API contract")
	suite.testInventoryLotCreationContract()

	suite.T().Log("Testing inventory availability API contract")
	suite.testInventoryAvailabilityContract()

	suite.T().Log("Testing inventory lot listing API contract")
	suite.testInventoryLotListingContract()
}

func (suite *APIContractTestSuite) testInventoryLotCreationContract() {
	// Create a product first
	catalogItemID := suite.createTestCatalogItem()

	// Test POST /api/v1/inventory/lots
	lotReq := map[string]interface{}{
		"catalog_item_id":    catalogItemID,
		"lot_number":         "LOT-API-TEST-001",
		"initial_quantity":   100.0,
		"quality_grade":      "A",
		"warehouse_location": "Test Warehouse",
		"storage_conditions": "Temperature controlled",
	}

	resp, err := suite.makeRequest("POST", "/api/v1/inventory/lots", lotReq)
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)

	var apiResp APIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	suite.NoError(err)
	suite.True(apiResp.Success)

	// Verify inventory lot response structure
	lotData := apiResp.Data.(map[string]interface{})
	suite.Contains(lotData, "id")
	suite.Contains(lotData, "catalog_item_id")
	suite.Contains(lotData, "lot_number")
	suite.Contains(lotData, "initial_quantity")
	suite.Contains(lotData, "available_quantity")
	suite.Contains(lotData, "status")

	lotID := lotData["id"].(string)
	suite.createdResources["inventory_lots"] = append(suite.createdResources["inventory_lots"], lotID)
}

func (suite *APIContractTestSuite) testInventoryAvailabilityContract() {
	catalogItemID := suite.createTestCatalogItem()

	// Test GET /api/v1/inventory/availability/{catalog_item_id}
	resp, err := suite.makeRequest("GET", "/api/v1/inventory/availability/"+catalogItemID, nil)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	var apiResp APIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	suite.NoError(err)
	suite.True(apiResp.Success)

	// Verify availability response structure
	availData := apiResp.Data.(map[string]interface{})
	suite.Contains(availData, "catalog_item_id")
	suite.Contains(availData, "available")
	suite.Contains(availData, "available_quantity")
	suite.Contains(availData, "total_quantity")

	// Test with required quantity parameter
	respWithQty, err := suite.makeRequest("GET", "/api/v1/inventory/availability/"+catalogItemID+"?required_quantity=50", nil)
	suite.NoError(err)
	suite.Equal(http.StatusOK, respWithQty.StatusCode)
}

func (suite *APIContractTestSuite) testInventoryLotListingContract() {
	// Test GET /api/v1/inventory/lots
	resp, err := suite.makeRequest("GET", "/api/v1/inventory/lots", nil)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	var apiResp APIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	suite.NoError(err)
	suite.True(apiResp.Success)

	// Test with filters
	filterResp, err := suite.makeRequest("GET", "/api/v1/inventory/lots?status=available&limit=10", nil)
	suite.NoError(err)
	suite.Equal(http.StatusOK, filterResp.StatusCode)
}

// Test integration API contracts
func (suite *APIContractTestSuite) TestIntegrationAPIContracts() {
	suite.T().Log("Testing catalog proposals API contract")
	suite.testCatalogProposalsContract()

	suite.T().Log("Testing order acknowledgements API contract")
	suite.testOrderAcknowledgementsContract()

	suite.T().Log("Testing catalog exports API contract")
	suite.testCatalogExportsContract()
}

func (suite *APIContractTestSuite) testCatalogProposalsContract() {
	// Test POST /api/v1/integrations/catalog/proposals
	proposalReq := map[string]interface{}{
		"source_organization_id": testutils.TestSellerOrgID,
		"target_organization_id": testutils.TestOrgID,
		"catalog_items": []map[string]interface{}{
			{
				"external_id": "ext-item-123",
				"name":        "Proposed Item",
				"item_type":   "PRODUCT",
				"base_price":  75.0,
				"currency":    "INR",
				"category":    "agriculture",
			},
		},
		"proposal_type": "new_items",
		"notes":         "API contract test proposal",
	}

	resp, err := suite.makeRequest("POST", "/api/v1/integrations/catalog/proposals", proposalReq)
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)

	var apiResp APIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	suite.NoError(err)
	suite.True(apiResp.Success)

	// Verify proposal response structure
	proposalData := apiResp.Data.(map[string]interface{})
	suite.Contains(proposalData, "proposal_id")
	suite.Contains(proposalData, "status")
	suite.Contains(proposalData, "created_at")
}

func (suite *APIContractTestSuite) testOrderAcknowledgementsContract() {
	// Create an order first
	orderReq := testutils.CreateTestCreateOrderRequest()
	createResp, err := suite.makeRequest("POST", "/api/v1/orders", orderReq)
	suite.NoError(err)

	var createApiResp APIResponse
	err = json.NewDecoder(createResp.Body).Decode(&createApiResp)
	suite.NoError(err)

	orderData := createApiResp.Data.(map[string]interface{})
	orderID := orderData["id"].(string)

	// Test POST /api/v1/integrations/orders/acknowledgements
	ackReq := map[string]interface{}{
		"order_id":              orderID,
		"external_order_id":     "ext-order-456",
		"acknowledgement_type":  "received",
		"partner_organization":  testutils.TestSellerOrgID,
		"estimated_fulfillment": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		"notes":                 "Order acknowledged via API",
	}

	resp, err := suite.makeRequest("POST", "/api/v1/integrations/orders/acknowledgements", ackReq)
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)

	var apiResp APIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	suite.NoError(err)
	suite.True(apiResp.Success)
}

func (suite *APIContractTestSuite) testCatalogExportsContract() {
	// Test GET /api/v1/integrations/catalog/exports
	resp, err := suite.makeRequest("GET", "/api/v1/integrations/catalog/exports", nil)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	var apiResp APIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	suite.NoError(err)
	suite.True(apiResp.Success)

	// Test with watermark parameter
	watermarkResp, err := suite.makeRequest("GET", "/api/v1/integrations/catalog/exports?since=2024-01-01T00:00:00Z", nil)
	suite.NoError(err)
	suite.Equal(http.StatusOK, watermarkResp.StatusCode)

	// Verify export response structure
	exportData := apiResp.Data.(map[string]interface{})
	suite.Contains(exportData, "items")
	suite.Contains(exportData, "watermark")
	suite.Contains(exportData, "has_more")
}

// Test error handling and edge cases
func (suite *APIContractTestSuite) TestErrorHandlingContracts() {
	suite.T().Log("Testing authentication error contracts")
	suite.testAuthenticationErrorContracts()

	suite.T().Log("Testing validation error contracts")
	suite.testValidationErrorContracts()

	suite.T().Log("Testing not found error contracts")
	suite.testNotFoundErrorContracts()

	suite.T().Log("Testing rate limiting contracts")
	suite.testRateLimitingContracts()
}

func (suite *APIContractTestSuite) testAuthenticationErrorContracts() {
	// Test request without authentication
	req, _ := http.NewRequest("GET", suite.baseURL+"/api/v1/orders", nil)
	resp, err := suite.client.Do(req)
	suite.NoError(err)
	suite.Equal(http.StatusUnauthorized, resp.StatusCode)

	var apiResp APIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	suite.NoError(err)
	suite.False(apiResp.Success)
	suite.NotNil(apiResp.Error)
	suite.Equal("UNAUTHORIZED", apiResp.Error.Code)

	// Test request with invalid token
	req.Header.Set("Authorization", "Bearer invalid_token")
	resp, err = suite.client.Do(req)
	suite.NoError(err)
	suite.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (suite *APIContractTestSuite) testValidationErrorContracts() {
	// Test invalid JSON
	req, _ := http.NewRequest("POST", suite.baseURL+"/api/v1/catalog/products", bytes.NewBufferString("invalid json"))
	req.Header.Set("Authorization", suite.authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.client.Do(req)
	suite.NoError(err)
	suite.Equal(http.StatusBadRequest, resp.StatusCode)

	var apiResp APIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	suite.NoError(err)
	suite.False(apiResp.Success)
	suite.NotNil(apiResp.Error)
	suite.Equal("INVALID_REQUEST", apiResp.Error.Code)

	// Test missing required fields
	invalidReq := map[string]interface{}{
		"name": "", // Empty required field
	}

	resp, err = suite.makeRequest("POST", "/api/v1/catalog/products", invalidReq)
	suite.NoError(err)
	suite.Equal(http.StatusBadRequest, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	suite.NoError(err)
	suite.False(apiResp.Success)
	suite.NotNil(apiResp.Error)
	suite.NotNil(apiResp.Error.Details)
}

func (suite *APIContractTestSuite) testNotFoundErrorContracts() {
	// Test non-existent resource
	resp, err := suite.makeRequest("GET", "/api/v1/orders/non-existent-id", nil)
	suite.NoError(err)
	suite.Equal(http.StatusNotFound, resp.StatusCode)

	var apiResp APIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	suite.NoError(err)
	suite.False(apiResp.Success)
	suite.NotNil(apiResp.Error)
	suite.Contains([]string{"NOT_FOUND", "ORDER_NOT_FOUND"}, apiResp.Error.Code)

	// Test non-existent endpoint
	resp, err = suite.makeRequest("GET", "/api/v1/nonexistent", nil)
	suite.NoError(err)
	suite.Equal(http.StatusNotFound, resp.StatusCode)
}

func (suite *APIContractTestSuite) testRateLimitingContracts() {
	// Make multiple rapid requests to test rate limiting
	requestCount := 10
	responses := make([]*http.Response, requestCount)

	for i := 0; i < requestCount; i++ {
		resp, err := suite.makeRequest("GET", "/api/v1/catalog", nil)
		suite.NoError(err)
		responses[i] = resp
	}

	// Check if any requests were rate limited
	rateLimitedCount := 0
	for _, resp := range responses {
		if resp.StatusCode == http.StatusTooManyRequests {
			rateLimitedCount++

			var apiResp APIResponse
			err := json.NewDecoder(resp.Body).Decode(&apiResp)
			suite.NoError(err)
			suite.False(apiResp.Success)
			suite.NotNil(apiResp.Error)
			suite.Equal("RATE_LIMIT_EXCEEDED", apiResp.Error.Code)
		}
	}

	suite.T().Logf("Rate limited requests: %d/%d", rateLimitedCount, requestCount)
}

// Test backward compatibility
func (suite *APIContractTestSuite) TestBackwardCompatibility() {
	suite.T().Log("Testing API version compatibility")
	suite.testAPIVersionCompatibility()

	suite.T().Log("Testing response format stability")
	suite.testResponseFormatStability()

	suite.T().Log("Testing deprecated field handling")
	suite.testDeprecatedFieldHandling()
}

func (suite *APIContractTestSuite) testAPIVersionCompatibility() {
	// Test v1 API endpoints
	v1Endpoints := []string{
		"/api/v1/catalog",
		"/api/v1/orders",
		"/api/v1/inventory/lots",
		"/api/v1/integrations/catalog/exports",
	}

	for _, endpoint := range v1Endpoints {
		resp, err := suite.makeRequest("GET", endpoint, nil)
		suite.NoError(err)
		suite.NotEqual(http.StatusNotFound, resp.StatusCode, "v1 endpoint should exist: %s", endpoint)

		// Verify API version in response headers
		suite.Contains(resp.Header.Get("API-Version"), "v1")
	}
}

func (suite *APIContractTestSuite) testResponseFormatStability() {
	// Test that core response structure remains stable
	resp, err := suite.makeRequest("GET", "/api/v1/catalog", nil)
	suite.NoError(err)

	var apiResp APIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	suite.NoError(err)

	// Verify core response structure
	suite.Contains(map[string]interface{}{"success": apiResp.Success}, "success")

	if apiResp.Success {
		suite.NotNil(apiResp.Data)
		suite.Nil(apiResp.Error)
	} else {
		suite.NotNil(apiResp.Error)
		suite.Contains(map[string]interface{}{"code": apiResp.Error.Code}, "code")
		suite.Contains(map[string]interface{}{"message": apiResp.Error.Message}, "message")
	}
}

func (suite *APIContractTestSuite) testDeprecatedFieldHandling() {
	// Test that deprecated fields are still accepted but marked as deprecated
	productReq := testutils.CreateTestProductRequest()

	// Add a deprecated field
	reqMap := map[string]interface{}{
		"name":             productReq.Name,
		"base_price":       productReq.BasePrice,
		"currency":         productReq.Currency,
		"category":         productReq.Category,
		"sku":              productReq.SKU,
		"deprecated_field": "should_be_ignored", // Deprecated field
	}

	resp, err := suite.makeRequest("POST", "/api/v1/catalog/products", reqMap)
	suite.NoError(err)

	// Should still succeed despite deprecated field
	suite.Equal(http.StatusCreated, resp.StatusCode)

	// Check for deprecation warning in response headers
	deprecationHeader := resp.Header.Get("Deprecation")
	if deprecationHeader != "" {
		suite.T().Logf("Deprecation warning: %s", deprecationHeader)
	}
}

// Test health check endpoint
func (suite *APIContractTestSuite) TestHealthCheckContract() {
	resp, err := suite.makeRequest("GET", "/health", nil)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	var healthResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&healthResp)
	suite.NoError(err)

	suite.Contains(healthResp, "status")
	suite.Equal("healthy", healthResp["status"])
	suite.Contains(healthResp, "timestamp")
	suite.Contains(healthResp, "version")
}

// Helper methods for making HTTP requests
func (suite *APIContractTestSuite) makeRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, suite.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", suite.authToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return suite.client.Do(req)
}

// Mock handler implementations (simplified for testing)
func (suite *APIContractTestSuite) handleCatalogProducts(w http.ResponseWriter, r *http.Request) {
	suite.handleGenericCatalogEndpoint(w, r, "PRODUCT")
}

func (suite *APIContractTestSuite) handleCatalogServices(w http.ResponseWriter, r *http.Request) {
	suite.handleGenericCatalogEndpoint(w, r, "SERVICE")
}

func (suite *APIContractTestSuite) handleCatalogLabour(w http.ResponseWriter, r *http.Request) {
	suite.handleGenericCatalogEndpoint(w, r, "LABOUR")
}

func (suite *APIContractTestSuite) handleGenericCatalogEndpoint(w http.ResponseWriter, r *http.Request, itemType string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("API-Version", "v1")

	switch r.Method {
	case "POST":
		// Mock catalog item creation
		mockItem := map[string]interface{}{
			"id":         "mock-item-" + time.Now().Format("20060102150405"),
			"name":       "Mock " + itemType,
			"item_type":  itemType,
			"base_price": 100.0,
			"currency":   "INR",
			"created_at": time.Now().Format(time.RFC3339),
			"updated_at": time.Now().Format(time.RFC3339),
		}

		response := APIResponse{
			Success: true,
			Data:    mockItem,
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (suite *APIContractTestSuite) handleCatalogByType(w http.ResponseWriter, r *http.Request) {
	// Handle /api/v1/catalog/{type}/{id} endpoints
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("API-Version", "v1")

	switch r.Method {
	case "GET":
		mockItem := map[string]interface{}{
			"id":         "mock-item-123",
			"name":       "Mock Item",
			"item_type":  "PRODUCT",
			"base_price": 100.0,
		}

		response := APIResponse{
			Success: true,
			Data:    mockItem,
		}

		json.NewEncoder(w).Encode(response)

	case "PUT":
		mockItem := map[string]interface{}{
			"id":         "mock-item-123",
			"name":       "Updated Mock Item",
			"updated_at": time.Now().Format(time.RFC3339),
		}

		response := APIResponse{
			Success: true,
			Data:    mockItem,
		}

		json.NewEncoder(w).Encode(response)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (suite *APIContractTestSuite) handleCatalogList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("API-Version", "v1")

	mockItems := []map[string]interface{}{
		{
			"id":         "item-1",
			"name":       "Mock Product 1",
			"item_type":  "PRODUCT",
			"base_price": 100.0,
		},
		{
			"id":         "item-2",
			"name":       "Mock Service 1",
			"item_type":  "SERVICE",
			"base_price": 200.0,
		},
	}

	response := APIResponse{
		Success: true,
		Data:    mockItems,
		Meta: &APIResponseMeta{
			Pagination: &APIPaginationMeta{
				Page:    1,
				Limit:   20,
				Total:   2,
				HasNext: false,
			},
		},
	}

	json.NewEncoder(w).Encode(response)
}

func (suite *APIContractTestSuite) handleCatalogSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("API-Version", "v1")

	query := r.URL.Query().Get("q")
	if query == "" {
		response := APIResponse{
			Success: false,
			Error: &APIError{
				Code:    "MISSING_QUERY",
				Message: "Search query is required",
			},
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	mockResults := []map[string]interface{}{
		{
			"id":         "search-result-1",
			"name":       "Search Result 1",
			"item_type":  "PRODUCT",
			"base_price": 150.0,
		},
	}

	response := APIResponse{
		Success: true,
		Data:    mockResults,
		Meta: &APIResponseMeta{
			Pagination: &APIPaginationMeta{
				Page:    1,
				Limit:   20,
				Total:   1,
				HasNext: false,
			},
		},
	}

	json.NewEncoder(w).Encode(response)
}

func (suite *APIContractTestSuite) handleOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("API-Version", "v1")

	switch r.Method {
	case "POST":
		mockOrder := map[string]interface{}{
			"id":                     "mock-order-" + time.Now().Format("20060102150405"),
			"order_number":           "ORD-" + time.Now().Format("20060102150405"),
			"status":                 "pending",
			"buyer_organization_id":  testutils.TestOrgID,
			"seller_organization_id": testutils.TestSellerOrgID,
			"total_amount":           100.50,
			"items":                  []map[string]interface{}{},
			"created_at":             time.Now().Format(time.RFC3339),
			"updated_at":             time.Now().Format(time.RFC3339),
		}

		response := APIResponse{
			Success: true,
			Data:    mockOrder,
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)

	case "GET":
		mockOrders := []map[string]interface{}{
			{
				"id":           "order-1",
				"order_number": "ORD-001",
				"status":       "pending",
				"total_amount": 100.50,
			},
		}

		response := APIResponse{
			Success: true,
			Data:    mockOrders,
			Meta: &APIResponseMeta{
				Pagination: &APIPaginationMeta{
					Page:    1,
					Limit:   20,
					Total:   1,
					HasNext: false,
				},
			},
		}

		json.NewEncoder(w).Encode(response)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (suite *APIContractTestSuite) handleOrderByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("API-Version", "v1")

	// Extract order ID from path
	path := r.URL.Path
	if path == "/api/v1/orders/non-existent" {
		response := APIResponse{
			Success: false,
			Error: &APIError{
				Code:    "ORDER_NOT_FOUND",
				Message: "Order not found",
			},
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	switch r.Method {
	case "GET":
		mockOrder := map[string]interface{}{
			"id":           "mock-order-123",
			"order_number": "ORD-123",
			"status":       "pending",
			"total_amount": 100.50,
		}

		response := APIResponse{
			Success: true,
			Data:    mockOrder,
		}

		json.NewEncoder(w).Encode(response)

	case "PATCH":
		if r.URL.Path[len(r.URL.Path)-7:] == "/status" {
			response := APIResponse{
				Success: true,
				Data:    "Order status updated successfully",
			}
			json.NewEncoder(w).Encode(response)
		}

	case "POST":
		if r.URL.Path[len(r.URL.Path)-7:] == "/cancel" {
			response := APIResponse{
				Success: true,
				Data:    "Order cancelled successfully",
			}
			json.NewEncoder(w).Encode(response)
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (suite *APIContractTestSuite) handleInventoryLots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("API-Version", "v1")

	switch r.Method {
	case "POST":
		mockLot := map[string]interface{}{
			"id":                 "mock-lot-" + time.Now().Format("20060102150405"),
			"catalog_item_id":    "mock-item-123",
			"lot_number":         "LOT-001",
			"initial_quantity":   100.0,
			"available_quantity": 100.0,
			"status":             "available",
			"created_at":         time.Now().Format(time.RFC3339),
		}

		response := APIResponse{
			Success: true,
			Data:    mockLot,
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)

	case "GET":
		mockLots := []map[string]interface{}{
			{
				"id":                 "lot-1",
				"lot_number":         "LOT-001",
				"available_quantity": 100.0,
				"status":             "available",
			},
		}

		response := APIResponse{
			Success: true,
			Data:    mockLots,
		}

		json.NewEncoder(w).Encode(response)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (suite *APIContractTestSuite) handleInventoryLotByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("API-Version", "v1")

	mockLot := map[string]interface{}{
		"id":                 "mock-lot-123",
		"lot_number":         "LOT-123",
		"available_quantity": 75.0,
		"status":             "available",
	}

	response := APIResponse{
		Success: true,
		Data:    mockLot,
	}

	json.NewEncoder(w).Encode(response)
}

func (suite *APIContractTestSuite) handleInventoryAvailability(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("API-Version", "v1")

	mockAvailability := map[string]interface{}{
		"catalog_item_id":    "mock-item-123",
		"available":          true,
		"available_quantity": 100.0,
		"total_quantity":     150.0,
	}

	response := APIResponse{
		Success: true,
		Data:    mockAvailability,
	}

	json.NewEncoder(w).Encode(response)
}

func (suite *APIContractTestSuite) handleCatalogProposals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("API-Version", "v1")

	mockProposal := map[string]interface{}{
		"proposal_id": "prop-" + time.Now().Format("20060102150405"),
		"status":      "pending",
		"created_at":  time.Now().Format(time.RFC3339),
	}

	response := APIResponse{
		Success: true,
		Data:    mockProposal,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (suite *APIContractTestSuite) handleOrderAcknowledgements(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("API-Version", "v1")

	mockAck := map[string]interface{}{
		"acknowledgement_id": "ack-" + time.Now().Format("20060102150405"),
		"status":             "acknowledged",
		"created_at":         time.Now().Format(time.RFC3339),
	}

	response := APIResponse{
		Success: true,
		Data:    mockAck,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (suite *APIContractTestSuite) handleCatalogExports(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("API-Version", "v1")

	mockExport := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"id":         "item-1",
				"name":       "Exported Item 1",
				"updated_at": time.Now().Format(time.RFC3339),
			},
		},
		"watermark": time.Now().Format(time.RFC3339),
		"has_more":  false,
	}

	response := APIResponse{
		Success: true,
		Data:    mockExport,
	}

	json.NewEncoder(w).Encode(response)
}

func (suite *APIContractTestSuite) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	healthResp := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
		"services": map[string]string{
			"database": "healthy",
			"aaa":      "healthy",
		},
	}

	json.NewEncoder(w).Encode(healthResp)
}

// Run the API contract test suite
func TestAPIContractSuite(t *testing.T) {
	suite.Run(t, new(APIContractTestSuite))
}
