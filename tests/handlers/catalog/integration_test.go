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
	catalogService "kisanlink-ecom/internal/services/catalog"
	"kisanlink-ecom/internal/services/catalog/mocks"
	"kisanlink-ecom/tests/data"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCatalogHandlers_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Complete catalog workflow", func(t *testing.T) {
		// Setup mock service
		mockService := new(mocks.CatalogServiceInterface)

		// Setup handlers
		etagService := &catalogService.ETagService{}
		productHandler := catalog.NewProductHandler(mockService, etagService)
		serviceHandler := catalog.NewServiceHandler(mockService, etagService)
		labourHandler := catalog.NewLabourHandler(mockService, etagService)
		catalogHandler := catalog.NewCatalogHandler(mockService, etagService)

		// Test 1: Create a product
		t.Run("Create Product", func(t *testing.T) {
			req := data.GetSampleCreateProductRequest()
			expectedProduct := data.GetSampleProduct()

			mockService.On("CreateProduct", mock.Anything, mock.AnythingOfType("*catalog.Product"), "test-user").
				Return(expectedProduct, nil).Once()

			reqBody, _ := json.Marshal(req)
			httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/catalog/products", bytes.NewBuffer(reqBody))
			httpReq.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httpReq
			c.Set("subjectID", "test-user")

			productHandler.CreateProduct(c)

			assert.Equal(t, http.StatusCreated, w.Code)
		})

		// Test 2: Create a service
		t.Run("Create Service", func(t *testing.T) {
			req := data.GetSampleCreateServiceRequest()
			expectedService := data.GetSampleService()

			mockService.On("CreateService", mock.Anything, mock.AnythingOfType("*catalog.CreateServiceRequest"), "test-user").
				Return(expectedService, nil).Once()

			reqBody, _ := json.Marshal(req)
			httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/catalog/services", bytes.NewBuffer(reqBody))
			httpReq.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httpReq
			c.Set("subjectID", "test-user")

			serviceHandler.CreateService(c)

			assert.Equal(t, http.StatusCreated, w.Code)
		})

		// Test 3: Create labour
		t.Run("Create Labour", func(t *testing.T) {
			req := data.GetSampleCreateLabourRequest()
			expectedLabour := data.GetSampleLabour()

			mockService.On("CreateLabour", mock.Anything, mock.AnythingOfType("*catalog.CreateLabourRequest"), "test-user").
				Return(expectedLabour, nil).Once()

			reqBody, _ := json.Marshal(req)
			httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/catalog/labour", bytes.NewBuffer(reqBody))
			httpReq.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httpReq
			c.Set("subjectID", "test-user")

			labourHandler.CreateLabour(c)

			assert.Equal(t, http.StatusCreated, w.Code)
		})

		// Test 4: List all catalog items
		t.Run("List All Catalog Items", func(t *testing.T) {
			expectedItems := data.GetSampleCatalogItems()

			mockService.On("ListCatalogItems", mock.Anything, mock.AnythingOfType("*catalog.CatalogFilter"), 0, 20).
				Return(expectedItems, len(expectedItems), nil).Once()

			httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/catalog?page=1&limit=20", nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httpReq

			catalogHandler.ListCatalogItems(c)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			success, exists := response["success"]
			if exists {
				assert.True(t, success.(bool))
			}
			assert.NotNil(t, response["data"])
		})

		// Test 5: List products by type
		t.Run("List Products by Type", func(t *testing.T) {
			expectedItems := []*catalogModels.CatalogItem{
				{
					Name:     "Organic Tomatoes",
					ItemType: catalogModels.CatalogItemTypeProduct,
					Category: "vegetables",
					IsActive: true,
				},
			}

			mockService.On("ListCatalogItems", mock.Anything, mock.MatchedBy(func(filter *catalogRequests.CatalogFilter) bool {
				return filter.ItemType != nil && *filter.ItemType == catalogModels.CatalogItemTypeProduct
			}), 0, 20).Return(expectedItems, 1, nil).Once()

			httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/catalog/products?page=1&limit=20", nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httpReq
			c.Params = gin.Params{
				{Key: "type", Value: "products"},
			}

			catalogHandler.ListCatalogItemsByType(c)

			assert.Equal(t, http.StatusOK, w.Code)
		})

		// Test 6: Search catalog
		t.Run("Search Catalog", func(t *testing.T) {
			expectedItems := []*catalogModels.CatalogItem{
				{
					Name:     "Organic Tomatoes",
					ItemType: catalogModels.CatalogItemTypeProduct,
					Category: "vegetables",
					IsActive: true,
				},
			}

			mockService.On("SearchCatalog", mock.Anything, "tomato", mock.AnythingOfType("*catalog.CatalogFilter"), 0, 20).
				Return(expectedItems, 1, nil).Once()

			httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/catalog/search?q=tomato&page=1&limit=20", nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httpReq

			catalogHandler.SearchCatalog(c)

			assert.Equal(t, http.StatusOK, w.Code)
		})

		// Verify all expectations were met
		mockService.AssertExpectations(t)
	})
}
