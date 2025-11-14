package catalog

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	catalogModels "github.com/Kisanlink/kisanlink-ecom/entities/models/catalog"
	catalogRequests "github.com/Kisanlink/kisanlink-ecom/entities/requests/catalog"
	"github.com/Kisanlink/kisanlink-ecom/internal/handlers/catalog"
	catalogService "github.com/Kisanlink/kisanlink-ecom/internal/services/catalog"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCatalogService is a mock implementation of CatalogServiceInterface
type MockCatalogService struct {
	mock.Mock
}

func (m *MockCatalogService) GetCatalogItemByID(ctx context.Context, id string) (*catalogModels.CatalogItem, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*catalogModels.CatalogItem), args.Error(1)
}

// Add other required methods as no-ops for this test
func (m *MockCatalogService) CreateProduct(_ context.Context, _ *catalogModels.Product, _ string) (*catalogModels.Product, error) {
	return nil, nil
}
func (m *MockCatalogService) GetProductByID(_ context.Context, _ string) (*catalogModels.Product, error) {
	return nil, nil
}
func (m *MockCatalogService) GetProductBySKU(_ context.Context, _ string) (*catalogModels.Product, error) {
	return nil, nil
}
func (m *MockCatalogService) UpdateProduct(_ context.Context, _ *catalogModels.Product, _ string) (*catalogModels.Product, error) {
	return nil, nil
}
func (m *MockCatalogService) DeleteProduct(_ context.Context, _ string, _ string) error {
	return nil
}
func (m *MockCatalogService) ListProducts(_ context.Context, _, _ int, _, _ string) ([]*catalogModels.CatalogItem, error) {
	return nil, nil
}
func (m *MockCatalogService) CreateService(_ context.Context, _ *catalogRequests.CreateServiceRequest, _ string) (*catalogModels.Service, error) {
	return nil, nil
}
func (m *MockCatalogService) GetServiceByID(_ context.Context, _ string) (*catalogModels.Service, error) {
	return nil, nil
}
func (m *MockCatalogService) GetServiceBySKU(_ context.Context, _ string) (*catalogModels.Service, error) {
	return nil, nil
}
func (m *MockCatalogService) UpdateService(_ context.Context, _ *catalogModels.Service, _ string) (*catalogModels.Service, error) {
	return nil, nil
}
func (m *MockCatalogService) DeleteService(_ context.Context, _ string, _ string) error {
	return nil
}
func (m *MockCatalogService) ListServices(_ context.Context, _, _ int, _, _ string) ([]*catalogModels.CatalogItem, error) {
	return nil, nil
}
func (m *MockCatalogService) CreateLabour(_ context.Context, _ *catalogRequests.CreateLabourRequest, _ string) (*catalogModels.Labour, error) {
	return nil, nil
}
func (m *MockCatalogService) GetLabourByID(_ context.Context, _ string) (*catalogModels.Labour, error) {
	return nil, nil
}
func (m *MockCatalogService) GetLabourBySKU(_ context.Context, _ string) (*catalogModels.Labour, error) {
	return nil, nil
}
func (m *MockCatalogService) UpdateLabour(_ context.Context, _ *catalogModels.Labour, _ string) (*catalogModels.Labour, error) {
	return nil, nil
}
func (m *MockCatalogService) DeleteLabour(_ context.Context, _ string, _ string) error {
	return nil
}
func (m *MockCatalogService) ListLabour(_ context.Context, _, _ int, _, _ string) ([]*catalogModels.CatalogItem, error) {
	return nil, nil
}
func (m *MockCatalogService) UpdateCatalogItem(_ context.Context, _ *catalogModels.CatalogItem, _ string) (*catalogModels.CatalogItem, error) {
	return nil, nil
}
func (m *MockCatalogService) DeleteContract(_ context.Context, _ string, _ string) error {
	return nil
}
func (m *MockCatalogService) ListCatalogItems(_ context.Context, _ *catalogRequests.CatalogFilter, _, _ int) ([]*catalogModels.CatalogItem, int, error) {
	return nil, 0, nil
}
func (m *MockCatalogService) SearchCatalog(_ context.Context, _ string, _ *catalogRequests.CatalogFilter, _, _ int) ([]*catalogModels.CatalogItem, int, error) {
	return nil, 0, nil
}
func (m *MockCatalogService) GetInventoryLevel(_ context.Context, _ string) (float64, error) {
	return 0, nil
}
func (m *MockCatalogService) UpdateInventory(_ context.Context, _ string, _ float64, _ string) error {
	return nil
}
func (m *MockCatalogService) UpdateActiveStatus(_ context.Context, _ string, _ bool, _ string) error {
	return nil
}
func (m *MockCatalogService) ListProductsForFPO(_ context.Context, _ string, _ *catalogRequests.CatalogFilter, _, _ int) ([]*catalogModels.ProductWithFPOPricing, error) {
	return nil, nil
}

func TestETagSupport(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockCatalogService)

	// Create ETag service
	etagService := catalogService.NewETagService(nil)

	// Create handler
	handler := catalog.NewCatalogHandler(mockService, etagService)

	// Create test catalog item
	testItem := &catalogModels.CatalogItem{
		ItemType: catalogModels.CatalogItemTypeProduct,
		Name:     "Test Product",
		Version:  1,
	}
	testItem.ID = "test-id-123"
	testItem.UpdatedAt = time.Now()

	// Mock the service call
	mockService.On("GetCatalogItemByID", mock.Anything, "test-id-123").Return(testItem, nil)

	t.Run("GET without If-None-Match should return 200 with ETag", func(t *testing.T) {
		// Create request
		req, _ := http.NewRequest("GET", "/api/v1/catalog/products/test-id-123", nil)
		w := httptest.NewRecorder()

		// Create gin context
		router := gin.New()
		router.GET("/api/v1/catalog/:type/:id", handler.GetCatalogItemByTypeAndID)
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, w.Header().Get("ETag"))
		assert.Contains(t, w.Header().Get("Cache-Control"), "private")
		assert.Contains(t, w.Header().Get("Cache-Control"), "must-revalidate")

		// Parse response to verify data
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Verify the response has data
		assert.NotNil(t, response["data"])
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "test-id-123", data["id"])
		assert.Equal(t, "Test Product", data["name"])
	})

	t.Run("GET with matching If-None-Match should return 304", func(t *testing.T) {
		// First, get the ETag
		req1, _ := http.NewRequest("GET", "/api/v1/catalog/products/test-id-123", nil)
		w1 := httptest.NewRecorder()

		router := gin.New()
		router.GET("/api/v1/catalog/:type/:id", handler.GetCatalogItemByTypeAndID)
		router.ServeHTTP(w1, req1)

		etag := w1.Header().Get("ETag")
		assert.NotEmpty(t, etag)

		// Now make request with If-None-Match
		req2, _ := http.NewRequest("GET", "/api/v1/catalog/products/test-id-123", nil)
		req2.Header.Set("If-None-Match", etag)
		w2 := httptest.NewRecorder()

		router.ServeHTTP(w2, req2)

		// Should return 304 Not Modified
		assert.Equal(t, http.StatusNotModified, w2.Code)
		assert.Equal(t, etag, w2.Header().Get("ETag"))
		assert.Empty(t, w2.Body.String()) // No body for 304
	})

	t.Run("GET with non-matching If-None-Match should return 200", func(t *testing.T) {
		// Create request with different ETag
		req, _ := http.NewRequest("GET", "/api/v1/catalog/products/test-id-123", nil)
		req.Header.Set("If-None-Match", `"different-etag"`)
		w := httptest.NewRecorder()

		router := gin.New()
		router.GET("/api/v1/catalog/:type/:id", handler.GetCatalogItemByTypeAndID)
		router.ServeHTTP(w, req)

		// Should return 200 with new ETag
		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, w.Header().Get("ETag"))
		assert.NotEqual(t, `"different-etag"`, w.Header().Get("ETag"))
	})

	mockService.AssertExpectations(t)
}

func TestETagGeneration(t *testing.T) {
	etagService := catalogService.NewETagService(nil)

	// Create test catalog item
	testItem := &catalogModels.CatalogItem{
		ItemType: catalogModels.CatalogItemTypeProduct,
		Name:     "Test Product",
		Version:  1,
	}
	testItem.ID = "test-id-123"
	testItem.UpdatedAt = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("ETag generation should be consistent", func(t *testing.T) {
		etag1 := etagService.GenerateETag(testItem)
		etag2 := etagService.GenerateETag(testItem)

		assert.Equal(t, etag1, etag2)
		assert.NotEmpty(t, etag1)
		assert.True(t, len(etag1) > 2) // Should be quoted
	})

	t.Run("ETag should change when version changes", func(t *testing.T) {
		etag1 := etagService.GenerateETag(testItem)

		testItem.Version = 2
		etag2 := etagService.GenerateETag(testItem)

		assert.NotEqual(t, etag1, etag2)
	})

	t.Run("ETag should change when updated time changes", func(t *testing.T) {
		etag1 := etagService.GenerateETag(testItem)

		testItem.UpdatedAt = time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)
		etag2 := etagService.GenerateETag(testItem)

		assert.NotEqual(t, etag1, etag2)
	})
}
