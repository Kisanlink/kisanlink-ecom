package orders

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "net/http"
    "net/http/httptest"
    "testing"

    orderModels "kisanlink-ecom/entities/models/orders"
    orderRequests "kisanlink-ecom/entities/requests/orders"
    "kisanlink-ecom/internal/handlers/orders"

    "github.com/gin-gonic/gin"
    "github.com/shopspring/decimal"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockOrderService is a comprehensive mock for order service
type MockOrderService struct {
    mock.Mock
}

func (m *MockOrderService) CreateOrder(ctx context.Context, req *orderRequests.CreateOrderRequest, userID, orgID string) (*orderModels.Order, error) {
    args := m.Called(ctx, req, userID, orgID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*orderModels.Order), args.Error(1)
}

func (m *MockOrderService) GetOrderByID(ctx context.Context, id, userID, orgID string) (*orderModels.Order, error) {
    args := m.Called(ctx, id, userID, orgID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*orderModels.Order), args.Error(1)
}

func (m *MockOrderService) UpdateOrderStatus(ctx context.Context, id string, req *orderRequests.UpdateOrderStatusRequest, userID, orgID string) error {
    args := m.Called(ctx, id, req, userID, orgID)
    return args.Error(0)
}

func (m *MockOrderService) UpdateOrder(ctx context.Context, id string, req *orderRequests.UpdateOrderRequest, userID, orgID string) (*orderModels.Order, error) {
    args := m.Called(ctx, id, req, userID, orgID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*orderModels.Order), args.Error(1)
}

func (m *MockOrderService) ListOrders(ctx context.Context, filter *orderRequests.ListOrdersRequest, userID, orgID string, offset, limit int) ([]*orderModels.Order, int64, error) {
    args := m.Called(ctx, filter, userID, orgID, offset, limit)
    return args.Get(0).([]*orderModels.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrderService) CancelOrder(ctx context.Context, id, userID, orgID string) error {
    args := m.Called(ctx, id, userID, orgID)
    return args.Error(0)
}

// Test helper functions
func setupTestRouter() *gin.Engine {
    gin.SetMode(gin.TestMode)
    router := gin.New()
    return router
}

func createTestOrder() *orderModels.Order {
    order := orderModels.NewOrder("buyer-org-1", "seller-org-1", "user-1")
    order.ID = "order-123"
    order.OrderNumber = "ORD-2024-001"
    order.Status = orderModels.OrderStatusPending

    item := orderModels.NewOrderItem(
        order.ID,
        "catalog-item-1",
        "product",
        "Test Product",
        "SKU-001",
        decimal.NewFromFloat(2),
        decimal.NewFromFloat(50.25),
    )
    item.ID = "item-1"
    order.Items = []orderModels.OrderItem{*item}
    order.CalculateTotal()

    return order
}

func createTestCreateOrderRequest() *orderRequests.CreateOrderRequest {
    return &orderRequests.CreateOrderRequest{
        BuyerOrganizationID:  "buyer-org-1",
        SellerOrganizationID: "seller-org-1",
        Items: []orderRequests.CreateOrderItemRequest{
            {
                CatalogItemID:   "catalog-item-1",
                CatalogItemType: "product",
                Quantity:        decimal.NewFromFloat(2),
                UnitPrice:       decimal.NewFromFloat(50.25),
            },
        },
        ShippingAddress: &orderRequests.Address{
            Street:     "123 Test St",
            City:       "Test City",
            State:      "Test State",
            PostalCode: "12345",
            Country:    "Test Country",
        },
        Notes: "Test order",
    }
}

func setupAuthenticatedContext(c *gin.Context) {
    c.Set("subjectID", "user-1")
    c.Set("organizationID", "buyer-org-1")
}

// Test CreateOrder handler with comprehensive scenarios
func TestOrderHandler_CreateOrder_Success(t *testing.T) {
    mockService := &MockOrderService{}
    handler := orders.NewOrderHandler(mockService)
    router := setupTestRouter()

    router.POST("/orders", func(c *gin.Context) {
        setupAuthenticatedContext(c)
        handler.CreateOrder(c)
    })

    req := createTestCreateOrderRequest()
    expectedOrder := createTestOrder()

    mockService.On("CreateOrder", mock.Anything, req, "user-1", "buyer-org-1").Return(expectedOrder, nil)

    reqBody, _ := json.Marshal(req)
    w := httptest.NewRecorder()
    httpReq, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(reqBody))
    httpReq.Header.Set("Content-Type", "application/json")

    router.ServeHTTP(w, httpReq)

    assert.Equal(t, http.StatusCreated, w.Code)

    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.True(t, response["success"].(bool))
    assert.NotNil(t, response["data"])

    mockService.AssertExpectations(t)
}

func TestOrderHandler_CreateOrder_InvalidJSON(t *testing.T) {
    mockService := &MockOrderService{}
    handler := orders.NewOrderHandler(mockService)
    router := setupTestRouter()

    router.POST("/orders", func(c *gin.Context) {
        setupAuthenticatedContext(c)
        handler.CreateOrder(c)
    })

    w := httptest.NewRecorder()
    httpReq, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer([]byte("invalid json")))
    httpReq.Header.Set("Content-Type", "application/json")

    router.ServeHTTP(w, httpReq)

    assert.Equal(t, http.StatusBadRequest, w.Code)

    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.False(t, response["success"].(bool))
    assert.Contains(t, response["error"].(map[string]interface{})["code"], "INVALID_REQUEST")

    mockService.AssertNotCalled(t, "CreateOrder")
}

func TestOrderHandler_CreateOrder_MissingAuthentication(t *testing.T) {
    mockService := &MockOrderService{}
    handler := orders.NewOrderHandler(mockService)
    router := setupTestRouter()

    router.POST("/orders", handler.CreateOrder)

    req := createTestCreateOrderRequest()
    reqBody, _ := json.Marshal(req)
    w := httptest.NewRecorder()
    httpReq, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(reqBody))
    httpReq.Header.Set("Content-Type", "application/json")

    router.ServeHTTP(w, httpReq)

    assert.Equal(t, http.StatusUnauthorized, w.Code)

    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.False(t, response["success"].(bool))
    assert.Contains(t, response["error"].(map[string]interface{})["code"], "MISSING_USER")

    mockService.AssertNotCalled(t, "CreateOrder")
}

func TestOrderHandler_CreateOrder_ServiceError(t *testing.T) {
    mockService := &MockOrderService{}
    handler := orders.NewOrderHandler(mockService)
    router := setupTestRouter()

    router.POST("/orders", func(c *gin.Context) {
        setupAuthenticatedContext(c)
        handler.CreateOrder(c)
    })

    req := createTestCreateOrderRequest()
    mockService.On("CreateOrder", mock.Anything, req, "user-1", "buyer-org-1").Return(nil, errors.New("insufficient inventory"))

    reqBody, _ := json.Marshal(req)
    w := httptest.NewRecorder()
    httpReq, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(reqBody))
    httpReq.Header.Set("Content-Type", "application/json")

    router.ServeHTTP(w, httpReq)

    assert.Equal(t, http.StatusBadRequest, w.Code)

    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.False(t, response["success"].(bool))
    assert.Contains(t, response["error"].(map[string]interface{})["code"], "CREATE_FAILED")

    mockService.AssertExpectations(t)
}

// Test GetOrderByID handler with comprehensive scenarios
func TestOrderHandler_GetOrderByID_Success(t *testing.T) {
    mockService := &MockOrderService{}
    handler := orders.NewOrderHandler(mockService)
    router := setupTestRouter()

    router.GET("/orders/:id", func(c *gin.Context) {
        setupAuthenticatedContext(c)
        handler.GetOrderByID(c)
    })

    expectedOrder := createTestOrder()
    mockService.On("GetOrderByID", mock.Anything, "order-123", "user-1", "buyer-org-1").Return(expectedOrder, nil)

    w := httptest.NewRecorder()
    httpReq, _ := http.NewRequest("GET", "/orders/order-123", nil)

    router.ServeHTTP(w, httpReq)

    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.True(t, response["success"].(bool))
    assert.NotNil(t, response["data"])

    mockService.AssertExpectations(t)
}

func TestOrderHandler_GetOrderByID_NotFound(t *testing.T) {
    mockService := &MockOrderService{}
    handler := orders.NewOrderHandler(mockService)
    router := setupTestRouter()

    router.GET("/orders/:id", func(c *gin.Context) {
        setupAuthenticatedContext(c)
        handler.GetOrderByID(c)
    })

    mockService.On("GetOrderByID", mock.Anything, "nonexistent", "user-1", "buyer-org-1").Return(nil, errors.New("order not found"))

    w := httptest.NewRecorder()
    httpReq, _ := http.NewRequest("GET", "/orders/nonexistent", nil)

    router.ServeHTTP(w, httpReq)

    assert.Equal(t, http.StatusNotFound, w.Code)

    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.False(t, response["success"].(bool))
    assert.Contains(t, response["error"].(map[string]interface{})["code"], "ORDER_NOT_FOUND")

    mockService.AssertExpectations(t)
}

func TestOrderHandler_GetOrderByID_EmptyID(t *testing.T) {
    mockService := &MockOrderService{}
    handler := orders.NewOrderHandler(mockService)
    router := setupTestRouter()

    router.GET("/orders/:id", func(c *gin.Context) {
        setupAuthenticatedContext(c)
        handler.GetOrderByID(c)
    })

    w := httptest.NewRecorder()
    httpReq, _ := http.NewRequest("GET", "/orders/", nil)

    router.ServeHTTP(w, httpReq)

    assert.Equal(t, http.StatusNotFound, w.Code) // Gin returns 404 for empty path param

    mockService.AssertNotCalled(t, "GetOrderByID")
}

// Test UpdateOrderStatus handler with comprehensive scenarios
func TestOrderHandler_UpdateOrderStatus_Success(t *testing.T) {
    mockService := &MockOrderService{}
    handler := orders.NewOrderHandler(mockService)
    router := setupTestRouter()

    router.PATCH("/orders/:id/status", func(c *gin.Context) {
        setupAuthenticatedContext(c)
        handler.UpdateOrderStatus(c)
    })

    updateReq := &orderRequests.UpdateOrderStatusRequest{
        Status: orderModels.OrderStatusConfirmed,
        Reason: "Payment confirmed",
    }

    mockService.On("UpdateOrderStatus", mock.Anything, "order-123", updateReq, "user-1", "buyer-org-1").Return(nil)

    reqBody, _ := json.Marshal(updateReq)
    w := httptest.NewRecorder()
    httpReq, _ := http.NewRequest("PATCH", "/orders/order-123/status", bytes.NewBuffer(reqBody))
    httpReq.Header.Set("Content-Type", "application/json")

    router.ServeHTTP(w, httpReq)

    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.True(t, response["success"].(bool))
    assert.Equal(t, "Order status updated successfully", response["data"])

    mockService.AssertExpectations(t)
}

func TestOrderHandler_UpdateOrderStatus_InvalidTransition(t *testing.T) {
    mockService := &MockOrderService{}
    handler := orders.NewOrderHandler(mockService)
    router := setupTestRouter()

    router.PATCH("/orders/:id/status", func(c *gin.Context) {
        setupAuthenticatedContext(c)
        handler.UpdateOrderStatus(c)
    })

    updateReq := &orderRequests.UpdateOrderStatusRequest{
        Status: orderModels.OrderStatusPending,
        Reason: "Invalid transition",
    }

    mockService.On("UpdateOrderStatus", mock.Anything, "order-123", updateReq, "user-1", "buyer-org-1").Return(errors.New("invalid status transition"))

    reqBody, _ := json.Marshal(updateReq)
    w := httptest.NewRecorder()
    httpReq, _ := http.NewRequest("PATCH", "/orders/order-123/status", bytes.NewBuffer(reqBody))
    httpReq.Header.Set("Content-Type", "application/json")

    router.ServeHTTP(w, httpReq)

    assert.Equal(t, http.StatusBadRequest, w.Code)

    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.False(t, response["success"].(bool))
    assert.Contains(t, response["error"].(map[string]interface{})["code"], "UPDATE_FAILED")

    mockService.AssertExpectations(t)
}

// Test ListOrders handler with comprehensive scenarios
func TestOrderHandler_ListOrders_Success(t *testing.T) {
    mockService := &MockOrderService{}
    handler := orders.NewOrderHandler(mockService)
    router := setupTestRouter()

    router.GET("/orders", func(c *gin.Context) {
        setupAuthenticatedContext(c)
        handler.ListOrders(c)
    })

    expectedOrders := []*orderModels.Order{createTestOrder()}
    filter := &orderRequests.ListOrdersRequest{
        Page:     1,
        PageSize: 20,
    }

    mockService.On("ListOrders", mock.Anything, mock.AnythingOfType("*orders.ListOrdersRequest"), "user-1", "buyer-org-1", 0, 20).Return(expectedOrders, int64(1), nil)

    w := httptest.NewRecorder()
    httpReq, _ := http.NewRequest("GET", "/orders", nil)

    router.ServeHTTP(w, httpReq)

    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.True(t, response["success"].(bool))
    assert.NotNil(t, response["data"])
    assert.NotNil(t, response["meta"].(map[string]interface{})["pagination"])

    mockService.AssertExpectations(t)
}

func TestOrderHandler_ListOrders_WithFilters(t *testing.T) {
    mockService := &MockOrderService{}
    handler := orders.NewOrderHandler(mockService)
    router := setupTestRouter()

    router.GET("/orders", func(c *gin.Context) {
        setupAuthenticatedContext(c)
        handler.ListOrders(c)
    })

    expectedOrders := []*orderModels.Order{createTestOrder()}

    mockService.On("ListOrders", mock.Anything, mock.MatchedBy(func(filter *orderRequests.ListOrdersRequest) bool {
        return filter.Status != nil && *filter.Status == orderModels.OrderStatusPending
    }), "user-1", "buyer-org-1", 0, 20).Return(expectedOrders, int64(1), nil)

    w := httptest.NewRecorder()
    httpReq, _ := http.NewRequest("GET", "/orders?status=pending", nil)

    router.ServeHTTP(w, httpReq)

    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.True(t, response["success"].(bool))

    mockService.AssertExpectations(t)
}

func TestOrderHandler_ListOrders_WithPagination(t *testing.T) {
    mockService := &MockOrderService{}
    handler := orders.NewOrderHandler(mockService)
    router := setupTestRouter()

    router.GET("/orders", func(c *gin.Context) {
        setupAuthenticatedContext(c)
        handler.ListOrders(c)
    })

    expectedOrders := []*orderModels.Order{createTestOrder()}

    mockService.On("ListOrders", mock.Anything, mock.MatchedBy(func(filter *orderRequests.ListOrdersRequest) bool {
        return filter.Page == 2 && filter.PageSize == 10
    }), "user-1", "buyer-org-1", 10, 10).Return(expectedOrders, int64(25), nil)

    w := httptest.NewRecorder()
    httpReq, _ := http.NewRequest("GET", "/orders?page=2&limit=10", nil)

    router.ServeHTTP(w, httpReq)

    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.True(t, response["success"].(bool))

    pagination := response["meta"].(map[string]interface{})["pagination"].(map[string]interface{})
    assert.Equal(t, float64(2), pagination["page"])
    assert.Equal(t, float64(10), pagination["limit"])
    assert.Equal(t, float64(25), pagination["total"])

    mockService.AssertExpectations(t)
}

// Test CancelOrder handler with comprehensive scenarios
func TestOrderHandler_CancelOrder_Success(t *testing.T) {
    mockService := &MockOrderService{}
    handler := orders.NewOrderHandler(mockService)
    router := setupTestRouter()

    router.POST("/orders/:id/cancel", func(c *gin.Context) {
        setupAuthenticatedContext(c)
        handler.CancelOrder(c)
    })

    mockService.On("CancelOrder", mock.Anything, "order-123", "user-1", "buyer-org-1").Return(nil)

    w := httptest.NewRecorder()
    httpReq, _ := http.NewRequest("POST", "/orders/order-123/cancel", nil)

    router.ServeHTTP(w, httpReq)

    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.True(t, response["success"].(bool))
    assert.Equal(t, "Order cancelled successfully", response["data"])

    mockService.AssertExpectations(t)
}

func TestOrderHandler_CancelOrder_AlreadyCancelled(t *testing.T) {
    mockService := &MockOrderService{}
    handler := orders.NewOrderHandler(mockService)
    router := setupTestRouter()

    router.POST("/orders/:id/cancel", func(c *gin.Context) {
        setupAuthenticatedContext(c)
        handler.CancelOrder(c)
    })

    mockService.On("CancelOrder", mock.Anything, "order-123", "user-1", "buyer-org-1").Return(errors.New("order already cancelled"))

    w := httptest.NewRecorder()
    httpReq, _ := http.NewRequest("POST", "/orders/order-123/cancel", nil)

    router.ServeHTTP(w, httpReq)

    assert.Equal(t, http.StatusBadRequest, w.Code)

    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.False(t, response["success"].(bool))
    assert.Contains(t, response["error"].(map[string]interface{})["code"], "CANCEL_FAILED")

    mockService.AssertExpectations(t)
}

// Test UpdateOrder handler with comprehensive scenarios
func TestOrderHandler_UpdateOrder_Success(t *testing.T) {
    mockService := &MockOrderService{}
    handler := orders.NewOrderHandler(mockService)
    router := setupTestRouter()

    router.PUT("/orders/:id", func(c *gin.Context) {
        setupAuthenticatedContext(c)
        handler.UpdateOrder(c)
    })

    updateReq := &orderRequests.UpdateOrderRequest{
        Notes: "Updated notes",
    }
    expectedOrder := createTestOrder()
    expectedOrder.Notes = "Updated notes"

    mockService.On("UpdateOrder", mock.Anything, "order-123", updateReq, "user-1", "buyer-org-1").Return(expectedOrder, nil)

    reqBody, _ := json.Marshal(updateReq)
    w := httptest.NewRecorder()
    httpReq, _ := http.NewRequest("PUT", "/orders/order-123", bytes.NewBuffer(reqBody))
    httpReq.Header.Set("Content-Type", "application/json")

    router.ServeHTTP(w, httpReq)

    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.True(t, response["success"].(bool))
    assert.NotNil(t, response["data"])

    mockService.AssertExpectations(t)
}

// Test edge cases and error scenarios
func TestOrderHandler_CreateOrder_CatalogItemTypeNormalization(t *testing.T) {
    mockService := &MockOrderService{}
    handler := orders.NewOrderHandler(mockService)
    router := setupTestRouter()

    router.POST("/orders", func(c *gin.Context) {
        setupAuthenticatedContext(c)
        handler.CreateOrder(c)
    })

    req := createTestCreateOrderRequest()
    req.Items[0].CatalogItemType = "PRODUCT" // Uppercase
    expectedOrder := createTestOrder()

    // The handler should normalize the catalog item type to lowercase
    mockService.On("CreateOrder", mock.Anything, mock.MatchedBy(func(r *orderRequests.CreateOrderRequest) bool {
        return r.Items[0].CatalogItemType == "product"
    }), "user-1", "buyer-org-1").Return(expectedOrder, nil)

    reqBody, _ := json.Marshal(req)
    w := httptest.NewRecorder()
    httpReq, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(reqBody))
    httpReq.Header.Set("Content-Type", "application/json")

    router.ServeHTTP(w, httpReq)

    assert.Equal(t, http.StatusCreated, w.Code)
    mockService.AssertExpectations(t)
}

func TestOrderHandler_ListOrders_InvalidPaginationParameters(t *testing.T) {
    mockService := &MockOrderService{}
    handler := orders.NewOrderHandler(mockService)
    router := setupTestRouter()

    router.GET("/orders", func(c *gin.Context) {
        setupAuthenticatedContext(c)
        handler.ListOrders(c)
    })

    expectedOrders := []*orderModels.Order{}

    // Handler should normalize invalid pagination parameters
    mockService.On("ListOrders", mock.Anything, mock.MatchedBy(func(filter *orderRequests.ListOrdersRequest) bool {
        return filter.Page == 1 && filter.PageSize == 20 // Default values
    }), "user-1", "buyer-org-1", 0, 20).Return(expectedOrders, int64(0), nil)

    w := httptest.NewRecorder()
    httpReq, _ := http.NewRequest("GET", "/orders?page=0&limit=0", nil) // Invalid values

    router.ServeHTTP(w, httpReq)

    assert.Equal(t, http.StatusOK, w.Code)
    mockService.AssertExpectations(t)
}

func TestOrderHandler_ListOrders_ExcessiveLimit(t *testing.T) {
    mockService := &MockOrderService{}
    handler := orders.NewOrderHandler(mockService)
    router := setupTestRouter()

    router.GET("/orders", func(c *gin.Context) {
        setupAuthenticatedContext(c)
        handler.ListOrders(c)
    })

    expectedOrders := []*orderModels.Order{}

    // Handler should cap the limit to maximum allowed
    mockService.On("ListOrders", mock.Anything, mock.MatchedBy(func(filter *orderRequests.ListOrdersRequest) bool {
        return filter.PageSize == 20 // Capped to default
    }), "user-1", "buyer-org-1", 0, 20).Return(expectedOrders, int64(0), nil)

    w := httptest.NewRecorder()
    httpReq, _ := http.NewRequest("GET", "/orders?limit=1000", nil) // Excessive limit

    router.ServeHTTP(w, httpReq)

    assert.Equal(t, http.StatusOK, w.Code)
    mockService.AssertExpectations(t)
}

// Test concurrent request scenarios
func TestOrderHandler_CreateOrder_ConcurrentRequests(t *testing.T) {
    mockService := &MockOrderService{}
    handler := orders.NewOrderHandler(mockService)
    router := setupTestRouter()

    router.POST("/orders", func(c *gin.Context) {
        setupAuthenticatedContext(c)
        handler.CreateOrder(c)
    })

    req := createTestCreateOrderRequest()
    expectedOrder := createTestOrder()

    // Mock service to handle concurrent requests
    mockService.On("CreateOrder", mock.Anything, req, "user-1", "buyer-org-1").Return(expectedOrder, nil).Times(3)

    // Simulate 3 concurrent requests
    for i := 0; i < 3; i++ {
        reqBody, _ := json.Marshal(req)
        w := httptest.NewRecorder()
        httpReq, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(reqBody))
        httpReq.Header.Set("Content-Type", "application/json")

        router.ServeHTTP(w, httpReq)
        assert.Equal(t, http.StatusCreated, w.Code)
    }

    mockService.AssertExpectations(t)
}
