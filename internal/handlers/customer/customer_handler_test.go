package customer

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"kisanlink-ecom/internal/models/user"
	"kisanlink-ecom/internal/services/customer"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCustomerService is a mock implementation of CustomerServiceInterface
type MockCustomerService struct {
	mock.Mock
}

func (m *MockCustomerService) CreateCustomer(ctx context.Context, aaaEntityID, customerCode string) (*user.Customer, error) {
	args := m.Called(ctx, aaaEntityID, customerCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.Customer), args.Error(1)
}

func (m *MockCustomerService) GetCustomerByID(ctx context.Context, customerID string) (*user.Customer, error) {
	args := m.Called(ctx, customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.Customer), args.Error(1)
}

func (m *MockCustomerService) GetCustomerByAAAEntityID(ctx context.Context, aaaEntityID string) (*user.Customer, error) {
	args := m.Called(ctx, aaaEntityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.Customer), args.Error(1)
}

func (m *MockCustomerService) GetCustomerByCustomerCode(ctx context.Context, customerCode string) (*user.Customer, error) {
	args := m.Called(ctx, customerCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.Customer), args.Error(1)
}

func (m *MockCustomerService) UpdateCustomer(ctx context.Context, customerID string, updates map[string]interface{}) (*user.Customer, error) {
	args := m.Called(ctx, customerID, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.Customer), args.Error(1)
}

func (m *MockCustomerService) ListCustomers(ctx context.Context, limit, offset int, status string) ([]*user.Customer, error) {
	args := m.Called(ctx, limit, offset, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*user.Customer), args.Error(1)
}

func (m *MockCustomerService) DeleteCustomer(ctx context.Context, customerID string) error {
	args := m.Called(ctx, customerID)
	return args.Error(0)
}

// Ensure MockCustomerService implements CustomerServiceInterface
var _ customer.CustomerServiceInterface = (*MockCustomerService)(nil)

func TestCreateCustomer(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	mockService := new(MockCustomerService)
	handler := NewCustomerHandler(mockService)

	router := gin.New()
	router.POST("/customers", handler.CreateCustomer)

	// Test data
	req := CreateCustomerRequest{
		AAAEntityID:  "user_123",
		CustomerCode: "CUST001",
	}

	// Mock service response
	expectedCustomer := user.NewCustomer("user_123", "CUST001")
	mockService.On("CreateCustomer", mock.Anything, "user_123", "CUST001").Return(expectedCustomer, nil)

	// Create request
	jsonData, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/customers", bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, request)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, "Customer created successfully", response["message"])

	// Verify mock was called
	mockService.AssertExpectations(t)
}

func TestGetCustomer(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	mockService := new(MockCustomerService)
	handler := NewCustomerHandler(mockService)

	router := gin.New()
	router.GET("/customers/:id", handler.GetCustomer)

	// Test data
	customerID := "CUST_123"
	expectedCustomer := user.NewCustomer("user_123", "CUST001")

	// Mock service response
	mockService.On("GetCustomerByID", mock.Anything, customerID).Return(expectedCustomer, nil)

	// Create request
	request := httptest.NewRequest("GET", "/customers/"+customerID, nil)

	// Create response recorder
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, request)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Equal(t, "Customer retrieved successfully", response["message"])

	// Verify mock was called
	mockService.AssertExpectations(t)
}
