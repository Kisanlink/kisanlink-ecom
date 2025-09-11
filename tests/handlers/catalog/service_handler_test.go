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

func TestServiceHandler_CreateService(t *testing.T) {
    gin.SetMode(gin.TestMode)

    tests := []struct {
        name           string
        request        catalogRequests.CreateServiceRequest
        mockSetup      func(*mocks.CatalogServiceInterface)
        expectedStatus int
        expectedError  string
    }{
        {
            name: "successful service creation",
            request: catalogRequests.CreateServiceRequest{
                CreateCatalogItemRequest: catalogRequests.CreateCatalogItemRequest{
                    ItemType:    catalogModels.CatalogItemTypeService,
                    Name:        "Agricultural Consulting",
                    Description: "Expert agricultural consulting services",
                    BasePrice:   decimal.NewFromFloat(100.0),
                    Currency:    "INR",
                    Category:    "consulting",
                },
                DurationMinutes: intPtr(60),
            },
            mockSetup: func(mockService *mocks.CatalogServiceInterface) {
                expectedService := &catalogModels.Service{
                    CatalogItem: catalogModels.CatalogItem{
                        Name:      "Agricultural Consulting",
                        BasePrice: decimal.NewFromFloat(100.0),
                        ItemType:  catalogModels.CatalogItemTypeService,
                    },
                }
                mockService.On("CreateService", mock.Anything, mock.AnythingOfType("*catalog.CreateServiceRequest"), "mock-user-id").
                    Return(expectedService, nil)
            },
            expectedStatus: http.StatusCreated,
        },
        {
            name: "invalid request body",
            request: catalogRequests.CreateServiceRequest{
                CreateCatalogItemRequest: catalogRequests.CreateCatalogItemRequest{
                    ItemType:  catalogModels.CatalogItemTypeService,
                    BasePrice: decimal.NewFromFloat(-10.0), // Invalid negative price
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
            request: catalogRequests.CreateServiceRequest{
                CreateCatalogItemRequest: catalogRequests.CreateCatalogItemRequest{
                    ItemType:  catalogModels.CatalogItemTypeProduct, // Wrong type
                    Name:      "Test Service",
                    BasePrice: decimal.NewFromFloat(100.0),
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

            handler := catalog.NewServiceHandler(mockService)

            // Create request
            reqBody, _ := json.Marshal(tt.request)
            req := httptest.NewRequest(http.MethodPost, "/api/v1/catalog/services", bytes.NewBuffer(reqBody))
            req.Header.Set("Content-Type", "application/json")

            // Create response recorder
            w := httptest.NewRecorder()

            // Create Gin context
            c, _ := gin.CreateTestContext(w)
            c.Request = req
            c.Set("subjectID", "mock-user-id") // Mock user context

            // Execute
            handler.CreateService(c)

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

func intPtr(i int) *int {
    return &i
}
