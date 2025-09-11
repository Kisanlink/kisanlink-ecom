package orders

import (
    "context"
    "errors"
    "testing"

    orderModels "kisanlink-ecom/entities/models/orders"
    orderRequests "kisanlink-ecom/entities/requests/orders"
    orderService "kisanlink-ecom/internal/services/orders"

    "github.com/shopspring/decimal"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockOrderRepository is a comprehensive mock for order repository
type MockOrderRepository struct {
    mock.Mock
}

func (m *MockOrderRepository) Create(ctx context.Context, order *orderModels.Order) error {
    args := m.Called(ctx, order)
    return args.Error(0)
}

func (m *MockOrderRepository) GetByID(ctx context.Context, id string) (*orderModels.Order, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*orderModels.Order), args.Error(1)
}

func (m *MockOrderRepository) Update(ctx context.Context, order *orderModels.Order) error {
    args := m.Called(ctx, order)
    return args.Error(0)
}

func (m *MockOrderRepository) Delete(ctx context.Context, id string) error {
    args := m.Called(ctx, id)
    return args.Error(0)
}

func (m *MockOrderRepository) List(ctx context.Context, filter *orderRequests.ListOrdersRequest, userID, orgID string, offset, limit int) ([]*orderModels.Order, int64, error) {
    args := m.Called(ctx, filter, userID, orgID, offset, limit)
    return args.Get(0).([]*orderModels.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrderRepository) ValidateOrderItems(ctx context.Context, items []orderRequests.CreateOrderItemRequest, sellerOrgID string) error {
    args := m.Called(ctx, items, sellerOrgID)
    return args.Error(0)
}

func (m *MockOrderRepository) ReserveInventory(ctx context.Context, items []orderRequests.CreateOrderItemRequest) error {
    args := m.Called(ctx, items)
    return args.Error(0)
}

func (m *MockOrderRepository) ReleaseInventory(ctx context.Context, items []orderRequests.CreateOrderItemRequest) error {
    args := m.Called(ctx, items)
    return args.Error(0)
}

// MockCatalogService is a comprehensive mock for catalog service
type MockCatalogService struct {
    mock.Mock
}

func (m *MockCatalogService) ValidateCatalogItems(ctx context.Context, items []orderRequests.CreateOrderItemRequest, sellerOrgID string) error {
    args := m.Called(ctx, items, sellerOrgID)
    return args.Error(0)
}

func (m *MockCatalogService) CheckInventoryAvailability(ctx context.Context, items []orderRequests.CreateOrderItemRequest) error {
    args := m.Called(ctx, items)
    return args.Error(0)
}

// Test helper functions
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

// Test CreateOrder with comprehensive scenarios
func TestOrderService_CreateOrder_Success(t *testing.T) {
    mockOrderRepo := &MockOrderRepository{}
    mockCatalogService := &MockCatalogService{}

    service := orderService.NewOrderService(mockOrderRepo, mockCatalogService)
    req := createTestCreateOrderRequest()

    // Mock successful validation and creation
    mockCatalogService.On("ValidateCatalogItems", mock.Anything, req.Items, req.SellerOrganizationID).Return(nil)
    mockCatalogService.On("CheckInventoryAvailability", mock.Anything, req.Items).Return(nil)
    mockOrderRepo.On("ValidateOrderItems", mock.Anything, req.Items, req.SellerOrganizationID).Return(nil)
    mockOrderRepo.On("ReserveInventory", mock.Anything, req.Items).Return(nil)
    mockOrderRepo.On("Create", mock.Anything, mock.AnythingOfType("*orders.Order")).Return(nil)

    order, err := service.CreateOrder(context.Background(), req, "user-1", "buyer-org-1")

    assert.NoError(t, err)
    assert.NotNil(t, order)
    assert.Equal(t, req.BuyerOrganizationID, order.BuyerOrganizationID)
    assert.Equal(t, req.SellerOrganizationID, order.SellerOrganizationID)
    assert.Equal(t, orderModels.OrderStatusPending, order.Status)
    assert.Len(t, order.Items, 1)
    assert.True(t, order.TotalAmount.Equal(decimal.NewFromFloat(100.50))) // 2 * 50.25

    mockOrderRepo.AssertExpectations(t)
    mockCatalogService.AssertExpectations(t)
}

func TestOrderService_CreateOrder_ValidationFailure(t *testing.T) {
    mockOrderRepo := &MockOrderRepository{}
    mockCatalogService := &MockCatalogService{}

    service := orderService.NewOrderService(mockOrderRepo, mockCatalogService)
    req := createTestCreateOrderRequest()

    // Mock validation failure
    mockCatalogService.On("ValidateCatalogItems", mock.Anything, req.Items, req.SellerOrganizationID).Return(errors.New("catalog item not found"))

    order, err := service.CreateOrder(context.Background(), req, "user-1", "buyer-org-1")

    assert.Error(t, err)
    assert.Nil(t, order)
    assert.Contains(t, err.Error(), "catalog item not found")

    mockCatalogService.AssertExpectations(t)
    // Order repo should not be called if validation fails
    mockOrderRepo.AssertNotCalled(t, "Create")
}

func TestOrderService_CreateOrder_InsufficientInventory(t *testing.T) {
    mockOrderRepo := &MockOrderRepository{}
    mockCatalogService := &MockCatalogService{}

    service := orderService.NewOrderService(mockOrderRepo, mockCatalogService)
    req := createTestCreateOrderRequest()

    // Mock successful catalog validation but insufficient inventory
    mockCatalogService.On("ValidateCatalogItems", mock.Anything, req.Items, req.SellerOrganizationID).Return(nil)
    mockCatalogService.On("CheckInventoryAvailability", mock.Anything, req.Items).Return(errors.New("insufficient inventory"))

    order, err := service.CreateOrder(context.Background(), req, "user-1", "buyer-org-1")

    assert.Error(t, err)
    assert.Nil(t, order)
    assert.Contains(t, err.Error(), "insufficient inventory")

    mockCatalogService.AssertExpectations(t)
    mockOrderRepo.AssertNotCalled(t, "Create")
}

func TestOrderService_CreateOrder_InventoryReservationFailure(t *testing.T) {
    mockOrderRepo := &MockOrderRepository{}
    mockCatalogService := &MockCatalogService{}

    service := orderService.NewOrderService(mockOrderRepo, mockCatalogService)
    req := createTestCreateOrderRequest()

    // Mock successful validation but inventory reservation failure
    mockCatalogService.On("ValidateCatalogItems", mock.Anything, req.Items, req.SellerOrganizationID).Return(nil)
    mockCatalogService.On("CheckInventoryAvailability", mock.Anything, req.Items).Return(nil)
    mockOrderRepo.On("ValidateOrderItems", mock.Anything, req.Items, req.SellerOrganizationID).Return(nil)
    mockOrderRepo.On("ReserveInventory", mock.Anything, req.Items).Return(errors.New("reservation failed"))

    order, err := service.CreateOrder(context.Background(), req, "user-1", "buyer-org-1")

    assert.Error(t, err)
    assert.Nil(t, order)
    assert.Contains(t, err.Error(), "reservation failed")

    mockOrderRepo.AssertExpectations(t)
    mockCatalogService.AssertExpectations(t)
    mockOrderRepo.AssertNotCalled(t, "Create")
}

// Test GetOrderByID with comprehensive scenarios
func TestOrderService_GetOrderByID_Success(t *testing.T) {
    mockOrderRepo := &MockOrderRepository{}
    mockCatalogService := &MockCatalogService{}

    service := orderService.NewOrderService(mockOrderRepo, mockCatalogService)
    expectedOrder := createTestOrder()

    mockOrderRepo.On("GetByID", mock.Anything, "order-123").Return(expectedOrder, nil)

    order, err := service.GetOrderByID(context.Background(), "order-123", "user-1", "buyer-org-1")

    assert.NoError(t, err)
    assert.NotNil(t, order)
    assert.Equal(t, expectedOrder.ID, order.ID)
    assert.Equal(t, expectedOrder.OrderNumber, order.OrderNumber)

    mockOrderRepo.AssertExpectations(t)
}

func TestOrderService_GetOrderByID_NotFound(t *testing.T) {
    mockOrderRepo := &MockOrderRepository{}
    mockCatalogService := &MockCatalogService{}

    service := orderService.NewOrderService(mockOrderRepo, mockCatalogService)

    mockOrderRepo.On("GetByID", mock.Anything, "nonexistent-order").Return(nil, errors.New("order not found"))

    order, err := service.GetOrderByID(context.Background(), "nonexistent-order", "user-1", "buyer-org-1")

    assert.Error(t, err)
    assert.Nil(t, order)
    assert.Contains(t, err.Error(), "order not found")

    mockOrderRepo.AssertExpectations(t)
}

// Test UpdateOrderStatus with comprehensive scenarios
func TestOrderService_UpdateOrderStatus_Success(t *testing.T) {
    mockOrderRepo := &MockOrderRepository{}
    mockCatalogService := &MockCatalogService{}

    service := orderService.NewOrderService(mockOrderRepo, mockCatalogService)
    existingOrder := createTestOrder()

    updateReq := &orderRequests.UpdateOrderStatusRequest{
        Status: orderModels.OrderStatusConfirmed,
        Reason: "Payment confirmed",
    }

    mockOrderRepo.On("GetByID", mock.Anything, "order-123").Return(existingOrder, nil)
    mockOrderRepo.On("Update", mock.Anything, mock.AnythingOfType("*orders.Order")).Return(nil)

    err := service.UpdateOrderStatus(context.Background(), "order-123", updateReq, "user-1", "seller-org-1")

    assert.NoError(t, err)
    mockOrderRepo.AssertExpectations(t)
}

func TestOrderService_UpdateOrderStatus_InvalidTransition(t *testing.T) {
    mockOrderRepo := &MockOrderRepository{}
    mockCatalogService := &MockCatalogService{}

    service := orderService.NewOrderService(mockOrderRepo, mockCatalogService)
    existingOrder := createTestOrder()
    existingOrder.Status = orderModels.OrderStatusCompleted // Terminal state

    updateReq := &orderRequests.UpdateOrderStatusRequest{
        Status: orderModels.OrderStatusPending,
        Reason: "Invalid transition",
    }

    mockOrderRepo.On("GetByID", mock.Anything, "order-123").Return(existingOrder, nil)

    err := service.UpdateOrderStatus(context.Background(), "order-123", updateReq, "user-1", "seller-org-1")

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "invalid status transition")

    mockOrderRepo.AssertExpectations(t)
    mockOrderRepo.AssertNotCalled(t, "Update")
}

// Test ListOrders with comprehensive scenarios
func TestOrderService_ListOrders_Success(t *testing.T) {
    mockOrderRepo := &MockOrderRepository{}
    mockCatalogService := &MockCatalogService{}

    service := orderService.NewOrderService(mockOrderRepo, mockCatalogService)

    expectedOrders := []*orderModels.Order{createTestOrder()}
    filter := &orderRequests.ListOrdersRequest{
        Page:     1,
        PageSize: 20,
    }

    mockOrderRepo.On("List", mock.Anything, filter, "user-1", "buyer-org-1", 0, 20).Return(expectedOrders, int64(1), nil)

    orders, total, err := service.ListOrders(context.Background(), filter, "user-1", "buyer-org-1", 0, 20)

    assert.NoError(t, err)
    assert.Len(t, orders, 1)
    assert.Equal(t, int64(1), total)
    assert.Equal(t, expectedOrders[0].ID, orders[0].ID)

    mockOrderRepo.AssertExpectations(t)
}

func TestOrderService_ListOrders_WithFilters(t *testing.T) {
    mockOrderRepo := &MockOrderRepository{}
    mockCatalogService := &MockCatalogService{}

    service := orderService.NewOrderService(mockOrderRepo, mockCatalogService)

    expectedOrders := []*orderModels.Order{createTestOrder()}
    status := orderModels.OrderStatusPending
    filter := &orderRequests.ListOrdersRequest{
        Page:     1,
        PageSize: 20,
        Status:   &status,
    }

    mockOrderRepo.On("List", mock.Anything, filter, "user-1", "buyer-org-1", 0, 20).Return(expectedOrders, int64(1), nil)

    orders, total, err := service.ListOrders(context.Background(), filter, "user-1", "buyer-org-1", 0, 20)

    assert.NoError(t, err)
    assert.Len(t, orders, 1)
    assert.Equal(t, int64(1), total)
    assert.Equal(t, orderModels.OrderStatusPending, orders[0].Status)

    mockOrderRepo.AssertExpectations(t)
}

// Test CancelOrder with comprehensive scenarios
func TestOrderService_CancelOrder_Success(t *testing.T) {
    mockOrderRepo := &MockOrderRepository{}
    mockCatalogService := &MockCatalogService{}

    service := orderService.NewOrderService(mockOrderRepo, mockCatalogService)
    existingOrder := createTestOrder()

    mockOrderRepo.On("GetByID", mock.Anything, "order-123").Return(existingOrder, nil)
    mockOrderRepo.On("ReleaseInventory", mock.Anything, mock.AnythingOfType("[]orders.CreateOrderItemRequest")).Return(nil)
    mockOrderRepo.On("Update", mock.Anything, mock.AnythingOfType("*orders.Order")).Return(nil)

    err := service.CancelOrder(context.Background(), "order-123", "user-1", "buyer-org-1")

    assert.NoError(t, err)
    mockOrderRepo.AssertExpectations(t)
}

func TestOrderService_CancelOrder_AlreadyCancelled(t *testing.T) {
    mockOrderRepo := &MockOrderRepository{}
    mockCatalogService := &MockCatalogService{}

    service := orderService.NewOrderService(mockOrderRepo, mockCatalogService)
    existingOrder := createTestOrder()
    existingOrder.Status = orderModels.OrderStatusCancelled

    mockOrderRepo.On("GetByID", mock.Anything, "order-123").Return(existingOrder, nil)

    err := service.CancelOrder(context.Background(), "order-123", "user-1", "buyer-org-1")

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "already cancelled")

    mockOrderRepo.AssertExpectations(t)
    mockOrderRepo.AssertNotCalled(t, "Update")
}

func TestOrderService_CancelOrder_InventoryReleaseFailure(t *testing.T) {
    mockOrderRepo := &MockOrderRepository{}
    mockCatalogService := &MockCatalogService{}

    service := orderService.NewOrderService(mockOrderRepo, mockCatalogService)
    existingOrder := createTestOrder()

    mockOrderRepo.On("GetByID", mock.Anything, "order-123").Return(existingOrder, nil)
    mockOrderRepo.On("ReleaseInventory", mock.Anything, mock.AnythingOfType("[]orders.CreateOrderItemRequest")).Return(errors.New("release failed"))

    err := service.CancelOrder(context.Background(), "order-123", "user-1", "buyer-org-1")

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "release failed")

    mockOrderRepo.AssertExpectations(t)
    mockOrderRepo.AssertNotCalled(t, "Update")
}

// Test edge cases and error scenarios
func TestOrderService_CreateOrder_NilRequest(t *testing.T) {
    mockOrderRepo := &MockOrderRepository{}
    mockCatalogService := &MockCatalogService{}

    service := orderService.NewOrderService(mockOrderRepo, mockCatalogService)

    order, err := service.CreateOrder(context.Background(), nil, "user-1", "buyer-org-1")

    assert.Error(t, err)
    assert.Nil(t, order)
    assert.Contains(t, err.Error(), "request cannot be nil")
}

func TestOrderService_CreateOrder_EmptyUserID(t *testing.T) {
    mockOrderRepo := &MockOrderRepository{}
    mockCatalogService := &MockCatalogService{}

    service := orderService.NewOrderService(mockOrderRepo, mockCatalogService)
    req := createTestCreateOrderRequest()

    order, err := service.CreateOrder(context.Background(), req, "", "buyer-org-1")

    assert.Error(t, err)
    assert.Nil(t, order)
    assert.Contains(t, err.Error(), "user ID is required")
}

func TestOrderService_CreateOrder_EmptyOrgID(t *testing.T) {
    mockOrderRepo := &MockOrderRepository{}
    mockCatalogService := &MockCatalogService{}

    service := orderService.NewOrderService(mockOrderRepo, mockCatalogService)
    req := createTestCreateOrderRequest()

    order, err := service.CreateOrder(context.Background(), req, "user-1", "")

    assert.Error(t, err)
    assert.Nil(t, order)
    assert.Contains(t, err.Error(), "organization ID is required")
}

// Test concurrent access scenarios
func TestOrderService_CreateOrder_ConcurrentInventoryReservation(t *testing.T) {
    mockOrderRepo := &MockOrderRepository{}
    mockCatalogService := &MockCatalogService{}

    service := orderService.NewOrderService(mockOrderRepo, mockCatalogService)
    req := createTestCreateOrderRequest()

    // Mock successful validation but concurrent reservation conflict
    mockCatalogService.On("ValidateCatalogItems", mock.Anything, req.Items, req.SellerOrganizationID).Return(nil)
    mockCatalogService.On("CheckInventoryAvailability", mock.Anything, req.Items).Return(nil)
    mockOrderRepo.On("ValidateOrderItems", mock.Anything, req.Items, req.SellerOrganizationID).Return(nil)
    mockOrderRepo.On("ReserveInventory", mock.Anything, req.Items).Return(errors.New("inventory already reserved by another order"))

    order, err := service.CreateOrder(context.Background(), req, "user-1", "buyer-org-1")

    assert.Error(t, err)
    assert.Nil(t, order)
    assert.Contains(t, err.Error(), "inventory already reserved")

    mockOrderRepo.AssertExpectations(t)
    mockCatalogService.AssertExpectations(t)
}

// Test business rule validations
func TestOrderService_CreateOrder_SameBuyerSellerOrganization(t *testing.T) {
    mockOrderRepo := &MockOrderRepository{}
    mockCatalogService := &MockCatalogService{}

    service := orderService.NewOrderService(mockOrderRepo, mockCatalogService)
    req := createTestCreateOrderRequest()
    req.BuyerOrganizationID = "same-org"
    req.SellerOrganizationID = "same-org"

    order, err := service.CreateOrder(context.Background(), req, "user-1", "same-org")

    assert.Error(t, err)
    assert.Nil(t, order)
    assert.Contains(t, err.Error(), "buyer and seller organizations cannot be the same")
}

func TestOrderService_CreateOrder_EmptyItems(t *testing.T) {
    mockOrderRepo := &MockOrderRepository{}
    mockCatalogService := &MockCatalogService{}

    service := orderService.NewOrderService(mockOrderRepo, mockCatalogService)
    req := createTestCreateOrderRequest()
    req.Items = []orderRequests.CreateOrderItemRequest{}

    order, err := service.CreateOrder(context.Background(), req, "user-1", "buyer-org-1")

    assert.Error(t, err)
    assert.Nil(t, order)
    assert.Contains(t, err.Error(), "order must contain at least one item")
}
