package catalog_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	catalogModels "kisanlink-ecom/entities/models/catalog"
	"kisanlink-ecom/internal/handlers/catalog"
	catalogService "kisanlink-ecom/internal/services/catalog"
	"kisanlink-ecom/tests/data"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPublishService mocks the PublishService interface
type MockPublishService struct {
	mock.Mock
}

func (m *MockPublishService) PublishProduct(ctx context.Context, req catalogService.PublishProductRequest) (*catalogModels.PublishState, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*catalogModels.PublishState), args.Error(1)
}

func (m *MockPublishService) GetPublishStatus(ctx context.Context, productID string) (*catalogModels.PublishState, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*catalogModels.PublishState), args.Error(1)
}

func (m *MockPublishService) UpdateDeliveryCosts(ctx context.Context, productID string, costs map[string]decimal.Decimal) error {
	args := m.Called(ctx, productID, costs)
	return args.Error(0)
}

func (m *MockPublishService) RevokeAccess(ctx context.Context, productID, fpoOrgID string) error {
	args := m.Called(ctx, productID, fpoOrgID)
	return args.Error(0)
}

func (m *MockPublishService) CalculateRetailPrice(basePrice, deliveryCost, platformFeePercent decimal.Decimal) decimal.Decimal {
	args := m.Called(basePrice, deliveryCost, platformFeePercent)
	return args.Get(0).(decimal.Decimal)
}

func (m *MockPublishService) ValidateFPOAccess(ctx context.Context, fpoOrgID, productID string) (bool, error) {
	args := m.Called(ctx, fpoOrgID, productID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPublishService) GetProductsForFPO(ctx context.Context, fpoOrgID string, offset, limit int) ([]*catalogModels.ProductWithFPOPricing, error) {
	args := m.Called(ctx, fpoOrgID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*catalogModels.ProductWithFPOPricing), args.Error(1)
}

const (
	testProductID = "PROD00000001"
	testFPOID1    = "ORGN00000002"
	testFPOID2    = "ORGN00000003"
)

func TestPublishProductToFPOs_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		productID      string
		requestBody    catalog.PublishProductToFPOsRequest
		mockSetup      func(*MockPublishService)
		expectedStatus int
	}{
		{
			name:      "publish to single FPO",
			productID: testProductID,
			requestBody: catalog.PublishProductToFPOsRequest{
				FPOIDs: []string{"ORGN00000002"},
				DeliveryCosts: map[string]float64{
					"ORGN00000002": 50.00,
				},
				PlatformFeePercent: 10.00,
			},
			mockSetup: func(ms *MockPublishService) {
				publishState := data.CreateTestPublishStateWithProductID(testProductID)
				ms.On("PublishProduct", mock.Anything, mock.MatchedBy(func(req catalogService.PublishProductRequest) bool {
					return req.ProductID == testProductID &&
						len(req.FPOIDs) == 1 &&
						req.FPOIDs[0] == "ORGN00000002"
				})).Return(publishState, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "publish to multiple FPOs",
			productID: testProductID,
			requestBody: catalog.PublishProductToFPOsRequest{
				FPOIDs: []string{"ORGN00000002", "ORGN00000003", "ORGN00000004"},
				DeliveryCosts: map[string]float64{
					"ORGN00000002": 50.00,
					"ORGN00000003": 75.00,
					"ORGN00000004": 100.00,
				},
				PlatformFeePercent: 10.00,
			},
			mockSetup: func(ms *MockPublishService) {
				fpoIDs := []string{"ORGN00000002", "ORGN00000003", "ORGN00000004"}
				deliveryCosts := data.CreateTestDeliveryCosts(fpoIDs)
				publishState := data.CreateTestPublishStateWithFPOs(testProductID, fpoIDs, deliveryCosts)
				ms.On("PublishProduct", mock.Anything, mock.MatchedBy(func(req catalogService.PublishProductRequest) bool {
					return req.ProductID == testProductID && len(req.FPOIDs) == 3
				})).Return(publishState, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockPublishService)
			tt.mockSetup(mockService)

			handler := catalog.NewFPOPublishHandler(mockService)

			// Create request body
			bodyBytes, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/catalog/products/"+tt.productID+"/publish", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "id", Value: tt.productID},
			}
			c.Set("subjectID", "admin-user-123")

			// Execute
			handler.PublishProductToFPOs(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check response body
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if success, exists := response["success"]; exists {
				assert.True(t, success.(bool))
			}
			assert.NotNil(t, response["data"])

			mockService.AssertExpectations(t)
		})
	}
}

func TestPublishProductToFPOs_ValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		productID      string
		requestBody    interface{}
		setupContext   func(*gin.Context)
		mockSetup      func(*MockPublishService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:      "missing product ID",
			productID: "",
			requestBody: catalog.PublishProductToFPOsRequest{
				FPOIDs: []string{"ORGN00000002"},
				DeliveryCosts: map[string]float64{
					"ORGN00000002": 50.00,
				},
				PlatformFeePercent: 10.00,
			},
			setupContext: func(c *gin.Context) {
				c.Set("subjectID", "admin-user-123")
			},
			mockSetup:      func(_ *MockPublishService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_PRODUCT_ID",
		},
		{
			name:      "missing user authentication",
			productID: testProductID,
			requestBody: catalog.PublishProductToFPOsRequest{
				FPOIDs: []string{"ORGN00000002"},
				DeliveryCosts: map[string]float64{
					"ORGN00000002": 50.00,
				},
				PlatformFeePercent: 10.00,
			},
			setupContext: func(_ *gin.Context) {
				// Don't set subjectID
			},
			mockSetup:      func(_ *MockPublishService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "MISSING_USER",
		},
		{
			name:        "invalid JSON body",
			productID:   "PROD00000001",
			requestBody: `{"invalid": json}`, // Invalid JSON
			setupContext: func(c *gin.Context) {
				c.Set("subjectID", "admin-user-123")
			},
			mockSetup:      func(_ *MockPublishService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_REQUEST",
		},
		{
			name:      "missing delivery cost for FPO",
			productID: testProductID,
			requestBody: catalog.PublishProductToFPOsRequest{
				FPOIDs: []string{"ORGN00000002", "ORGN00000003"},
				DeliveryCosts: map[string]float64{
					"ORGN00000002": 50.00,
					// Missing cost for ORGN00000003
				},
				PlatformFeePercent: 10.00,
			},
			setupContext: func(c *gin.Context) {
				c.Set("subjectID", "admin-user-123")
			},
			mockSetup:      func(_ *MockPublishService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_DELIVERY_COST",
		},
		{
			name:      "service returns error",
			productID: testProductID,
			requestBody: catalog.PublishProductToFPOsRequest{
				FPOIDs: []string{"ORGN00000002"},
				DeliveryCosts: map[string]float64{
					"ORGN00000002": 50.00,
				},
				PlatformFeePercent: 10.00,
			},
			setupContext: func(c *gin.Context) {
				c.Set("subjectID", "admin-user-123")
			},
			mockSetup: func(ms *MockPublishService) {
				ms.On("PublishProduct", mock.Anything, mock.Anything).
					Return(nil, errors.New("product not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "PUBLISH_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockPublishService)
			tt.mockSetup(mockService)

			handler := catalog.NewFPOPublishHandler(mockService)

			// Create request body
			var bodyBytes []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				bodyBytes = []byte(str)
			} else {
				bodyBytes, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/catalog/products/"+tt.productID+"/publish", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "id", Value: tt.productID},
			}
			tt.setupContext(c)

			// Execute
			handler.PublishProductToFPOs(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check response body
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedError != "" {
				errorObj, exists := response["error"]
				assert.True(t, exists)
				if errorMap, ok := errorObj.(map[string]interface{}); ok {
					code, exists := errorMap["code"]
					assert.True(t, exists)
					assert.Equal(t, tt.expectedError, code)
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestGetPublishStatus_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	mockService := new(MockPublishService)
	publishState := data.CreateTestPublishStateWithProductID(testProductID)

	mockService.On("GetPublishStatus", mock.Anything, testProductID).
		Return(publishState, nil)

	handler := catalog.NewFPOPublishHandler(mockService)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/catalog/products/"+testProductID+"/publish-status", nil)
	w := httptest.NewRecorder()

	// Create Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: productID},
	}

	// Execute
	handler.GetPublishStatus(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	if success, exists := response["success"]; exists {
		assert.True(t, success.(bool))
	}
	assert.NotNil(t, response["data"])

	mockService.AssertExpectations(t)
}

func TestGetPublishStatus_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	mockService := new(MockPublishService)
	productID := "PROD00000099"

	mockService.On("GetPublishStatus", mock.Anything, productID).
		Return(nil, errors.New("publish state not found"))

	handler := catalog.NewFPOPublishHandler(mockService)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/catalog/products/"+productID+"/publish-status", nil)
	w := httptest.NewRecorder()

	// Create Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: productID},
	}

	// Execute
	handler.GetPublishStatus(c)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	errorObj, exists := response["error"]
	assert.True(t, exists)
	if errorMap, ok := errorObj.(map[string]interface{}); ok {
		code, exists := errorMap["code"]
		assert.True(t, exists)
		assert.Equal(t, "PUBLISH_STATE_NOT_FOUND", code)
	}

	mockService.AssertExpectations(t)
}

func TestUpdateProductDeliveryCosts_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	mockService := new(MockPublishService)
	productID := "PROD00000001"

	requestBody := catalog.UpdateDeliveryCostsRequest{
		DeliveryCosts: map[string]float64{
			"ORGN00000002": 60.00,
			"ORGN00000003": 80.00,
		},
	}

	mockService.On("UpdateDeliveryCosts", mock.Anything, productID, mock.MatchedBy(func(costs map[string]decimal.Decimal) bool {
		return len(costs) == 2
	})).Return(nil)

	handler := catalog.NewFPOPublishHandler(mockService)

	// Create request
	bodyBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/catalog/products/"+productID+"/delivery-costs", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Create Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: productID},
	}

	// Execute
	handler.UpdateProductDeliveryCosts(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	if success, exists := response["success"]; exists {
		assert.True(t, success.(bool))
	}

	mockService.AssertExpectations(t)
}

func TestUpdateProductDeliveryCosts_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		productID      string
		requestBody    interface{}
		mockSetup      func(*MockPublishService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "missing product ID",
			productID:      "",
			requestBody:    catalog.UpdateDeliveryCostsRequest{DeliveryCosts: map[string]float64{"ORGN00000002": 50.00}},
			mockSetup:      func(_ *MockPublishService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_PRODUCT_ID",
		},
		{
			name:           "invalid JSON body",
			productID:      "PROD00000001",
			requestBody:    `{"invalid": json}`,
			mockSetup:      func(_ *MockPublishService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_REQUEST",
		},
		{
			name:      "service returns error",
			productID: testProductID,
			requestBody: catalog.UpdateDeliveryCostsRequest{
				DeliveryCosts: map[string]float64{"ORGN00000002": 50.00},
			},
			mockSetup: func(ms *MockPublishService) {
				ms.On("UpdateDeliveryCosts", mock.Anything, "PROD00000001", mock.Anything).
					Return(errors.New("product not published"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "UPDATE_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockPublishService)
			tt.mockSetup(mockService)

			handler := catalog.NewFPOPublishHandler(mockService)

			// Create request body
			var bodyBytes []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				bodyBytes = []byte(str)
			} else {
				bodyBytes, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPatch, "/api/v1/catalog/products/"+tt.productID+"/delivery-costs", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "id", Value: tt.productID},
			}

			// Execute
			handler.UpdateProductDeliveryCosts(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedError != "" {
				errorObj, exists := response["error"]
				assert.True(t, exists)
				if errorMap, ok := errorObj.(map[string]interface{}); ok {
					code, exists := errorMap["code"]
					assert.True(t, exists)
					assert.Equal(t, tt.expectedError, code)
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestRevokeFPOAccess_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	mockService := new(MockPublishService)
	productID := "PROD00000001"
	fpoOrgID := "ORGN00000002"

	mockService.On("RevokeAccess", mock.Anything, productID, fpoOrgID).
		Return(nil)

	handler := catalog.NewFPOPublishHandler(mockService)

	// Create request
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/catalog/products/"+productID+"/fpo-access/"+fpoOrgID, nil)
	w := httptest.NewRecorder()

	// Create Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: productID},
		{Key: "fpo_id", Value: fpoOrgID},
	}

	// Execute
	handler.RevokeFPOAccess(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	if success, exists := response["success"]; exists {
		assert.True(t, success.(bool))
	}

	mockService.AssertExpectations(t)
}

func TestRevokeFPOAccess_ValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		productID      string
		fpoOrgID       string
		mockSetup      func(*MockPublishService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "missing product ID",
			productID:      "",
			fpoOrgID:       "ORGN00000002",
			mockSetup:      func(_ *MockPublishService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_PRODUCT_ID",
		},
		{
			name:           "missing FPO ID",
			productID:      "PROD00000001",
			fpoOrgID:       "",
			mockSetup:      func(_ *MockPublishService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_FPO_ID",
		},
		{
			name:      "service returns error",
			productID: testProductID,
			fpoOrgID:  "ORGN00000002",
			mockSetup: func(ms *MockPublishService) {
				ms.On("RevokeAccess", mock.Anything, "PROD00000001", "ORGN00000002").
					Return(errors.New("FPO does not have access"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "REVOKE_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockPublishService)
			tt.mockSetup(mockService)

			handler := catalog.NewFPOPublishHandler(mockService)

			// Create request
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/catalog/products/"+tt.productID+"/fpo-access/"+tt.fpoOrgID, nil)
			w := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "id", Value: tt.productID},
				{Key: "fpo_id", Value: tt.fpoOrgID},
			}

			// Execute
			handler.RevokeFPOAccess(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedError != "" {
				errorObj, exists := response["error"]
				assert.True(t, exists)
				if errorMap, ok := errorObj.(map[string]interface{}); ok {
					code, exists := errorMap["code"]
					assert.True(t, exists)
					assert.Equal(t, tt.expectedError, code)
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}
