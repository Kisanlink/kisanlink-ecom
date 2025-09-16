package catalog_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	catalogModels "kisanlink-ecom/entities/models/catalog"
	catalogRequests "kisanlink-ecom/entities/requests/catalog"
	"kisanlink-ecom/internal/handlers/catalog"
	"kisanlink-ecom/internal/services/catalog/mocks"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func TestLabourHandler_CreateLabour(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		request        catalogRequests.CreateLabourRequest
		mockSetup      func(*mocks.CatalogServiceInterface)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful labour creation",
			request: catalogRequests.CreateLabourRequest{
				CreateCatalogItemRequest: catalogRequests.CreateCatalogItemRequest{
					ItemType:    catalogModels.CatalogItemTypeLabour,
					Name:        "Farm Worker",
					Description: "Experienced farm worker for seasonal work",
					BasePrice:   decimal.NewFromFloat(15.0),
					Currency:    "INR",
					Category:    "agricultural-labor",
				},
				SkillLevel: "intermediate",
				HourlyRate: decimalPtr(decimal.NewFromFloat(15.0)),
			},
			mockSetup: func(mockService *mocks.CatalogServiceInterface) {
				expectedLabour := &catalogModels.Labour{
					CatalogItem: catalogModels.CatalogItem{
						Name:      "Farm Worker",
						BasePrice: decimal.NewFromFloat(15.0),
						ItemType:  catalogModels.CatalogItemTypeLabour,
					},
					SkillLevel: "intermediate",
				}
				mockService.On("CreateLabour", mock.Anything, mock.AnythingOfType("*catalog.CreateLabourRequest"), "mock-user-id").
					Return(expectedLabour, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid request body",
			request: catalogRequests.CreateLabourRequest{
				CreateCatalogItemRequest: catalogRequests.CreateCatalogItemRequest{
					ItemType:  catalogModels.CatalogItemTypeLabour,
					BasePrice: decimal.NewFromFloat(-5.0), // Invalid negative price
				},
			},
			mockSetup: func(mockService *mocks.CatalogServiceInterface) {
				// No mock setup needed for validation errors
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_REQUEST",
		},
		{
			name: "wrong item type",
			request: catalogRequests.CreateLabourRequest{
				CreateCatalogItemRequest: catalogRequests.CreateCatalogItemRequest{
					ItemType:  catalogModels.CatalogItemTypeProduct, // Wrong type
					Name:      "Test Labour",
					BasePrice: decimal.NewFromFloat(15.0),
				},
			},
			mockSetup: func(mockService *mocks.CatalogServiceInterface) {
				// No mock setup needed for validation errors
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_TYPE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(mocks.CatalogServiceInterface)
			tt.mockSetup(mockService)

			handler := catalog.NewLabourHandler(mockService)

			// Create request
			reqBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/catalog/labour", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("subjectID", "mock-user-id") // Mock user context

			// Execute
			handler.CreateLabour(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				success, exists := response["success"]
				if exists {
					assert.False(t, success.(bool))
				}
				if errorObj, ok := response["error"].(map[string]interface{}); ok {
					if code, ok := errorObj["code"].(string); ok {
						assert.Contains(t, code, tt.expectedError)
					}
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func floatPtr(f float64) *float64 {
	return &f
}
