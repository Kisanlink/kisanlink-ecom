package catalog_test

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "net/url"
    "testing"

    catalogModels "kisanlink-ecom/entities/models/catalog"
    "kisanlink-ecom/internal/handlers/catalog"
    "kisanlink-ecom/internal/services/catalog/mocks"

    "github.com/gin-gonic/gin"
    "github.com/shopspring/decimal"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestCatalogHandler_ListCatalogItems(t *testing.T) {
    gin.SetMode(gin.TestMode)

    tests := []struct {
        name           string
        queryParams    map[string]string
        mockSetup      func(*mocks.CatalogServiceInterface)
        expectedStatus int
        expectedError  string
    }{
        {
            name: "successful list with no filters",
            queryParams: map[string]string{
                "page":  "1",
                "limit": "20",
            },
            mockSetup: func(mockService *mocks.CatalogServiceInterface) {
                expectedItems := []*catalogModels.CatalogItem{
                    {
                        Name:      "Test Product",
                        ItemType:  catalogModels.CatalogItemTypeProduct,
                        BasePrice: decimal.NewFromFloat(100.0),
                    },
                    {
                        Name:      "Test Service",
                        ItemType:  catalogModels.CatalogItemTypeService,
                        BasePrice: decimal.NewFromFloat(50.0),
                    },
                }
                mockService.On("ListCatalogItems", mock.Anything, mock.AnythingOfType("*catalog.CatalogFilter"), 0, 20).
                    Return(expectedItems, 2, nil)
            },
            expectedStatus: http.StatusOK,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            mockService := new(mocks.CatalogServiceInterface)
            tt.mockSetup(mockService)

            handler := catalog.NewCatalogHandler(mockService)

            // Create request with query parameters
            req := httptest.NewRequest(http.MethodGet, "/api/v1/catalog", nil)
            q := url.Values{}
            for key, value := range tt.queryParams {
                q.Add(key, value)
            }
            req.URL.RawQuery = q.Encode()

            // Create response recorder
            w := httptest.NewRecorder()

            // Create Gin context
            c, _ := gin.CreateTestContext(w)
            c.Request = req

            // Execute
            handler.ListCatalogItems(c)

            // Assert
            assert.Equal(t, tt.expectedStatus, w.Code)

            // Check response body
            var response map[string]interface{}
            err := json.Unmarshal(w.Body.Bytes(), &response)
            assert.NoError(t, err)

            if tt.expectedError != "" {
                success, exists := response["success"]
                if exists {
                    assert.False(t, success.(bool))
                }
            } else {
                success, exists := response["success"]
                if exists {
                    assert.True(t, success.(bool))
                }
                assert.NotNil(t, response["data"])
                assert.NotNil(t, response["meta"])
            }

            mockService.AssertExpectations(t)
        })
    }
}
