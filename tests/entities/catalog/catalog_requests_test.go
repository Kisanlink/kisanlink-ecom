package catalog

import (
	"fmt"
	"testing"

	catalogModels "kisanlink-ecom/entities/models/catalog"
	catalogRequests "kisanlink-ecom/entities/requests/catalog"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCreateCatalogItemRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request catalogRequests.CreateCatalogItemRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid product request",
			request: catalogRequests.CreateCatalogItemRequest{
				ItemType:      catalogModels.CatalogItemTypeProduct,
				Category:      "vegetables",
				Subcategory:   "tomatoes",
				Name:          "Organic Tomatoes",
				Description:   "Fresh organic tomatoes grown without pesticides",
				SKU:           "TOM-ORG-001",
				UnitOfMeasure: "kg",
				BasePrice:     decimal.NewFromFloat(25.50),
				Currency:      "INR",
				Visibility:    catalogModels.VisibilityOrg,
				Tags:          []string{"organic", "fresh", "local"},
				Images:        []string{"https://example.com/tomato1.jpg"},
			},
			wantErr: false,
		},
		{
			name: "valid service request",
			request: catalogRequests.CreateCatalogItemRequest{
				ItemType:      catalogModels.CatalogItemTypeService,
				Name:          "Farm Consultation",
				Description:   "Expert agricultural consultation service",
				BasePrice:     decimal.NewFromFloat(150.00),
				UnitOfMeasure: "hour",
			},
			wantErr: false,
		},
		{
			name: "valid labour request",
			request: catalogRequests.CreateCatalogItemRequest{
				ItemType:      catalogModels.CatalogItemTypeLabour,
				Name:          "Farm Worker",
				Description:   "Experienced farm worker for seasonal work",
				BasePrice:     decimal.NewFromFloat(75.00),
				UnitOfMeasure: "hour",
			},
			wantErr: false,
		},
		{
			name: "missing item type",
			request: catalogRequests.CreateCatalogItemRequest{
				Name:      "Test Item",
				BasePrice: decimal.NewFromFloat(25.50),
			},
			wantErr: true,
			errMsg:  "item_type is required",
		},
		{
			name: "invalid item type",
			request: catalogRequests.CreateCatalogItemRequest{
				ItemType:  "INVALID",
				Name:      "Test Item",
				BasePrice: decimal.NewFromFloat(25.50),
			},
			wantErr: true,
			errMsg:  "item_type must be one of: PRODUCT, SERVICE, LABOUR",
		},
		{
			name: "missing name",
			request: catalogRequests.CreateCatalogItemRequest{
				ItemType:  catalogModels.CatalogItemTypeProduct,
				BasePrice: decimal.NewFromFloat(25.50),
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "name too long",
			request: catalogRequests.CreateCatalogItemRequest{
				ItemType:  catalogModels.CatalogItemTypeProduct,
				Name:      string(make([]byte, 256)), // 256 characters
				BasePrice: decimal.NewFromFloat(25.50),
			},
			wantErr: true,
			errMsg:  "name must be at most 255 characters",
		},
		{
			name: "zero base price",
			request: catalogRequests.CreateCatalogItemRequest{
				ItemType:  catalogModels.CatalogItemTypeProduct,
				Name:      "Test Item",
				BasePrice: decimal.Zero,
			},
			wantErr: true,
			errMsg:  "base_price must be greater than 0",
		},
		{
			name: "negative base price",
			request: catalogRequests.CreateCatalogItemRequest{
				ItemType:  catalogModels.CatalogItemTypeProduct,
				Name:      "Test Item",
				BasePrice: decimal.NewFromFloat(-10.00),
			},
			wantErr: true,
			errMsg:  "base_price must be greater than 0",
		},
		{
			name: "invalid currency code",
			request: catalogRequests.CreateCatalogItemRequest{
				ItemType:  catalogModels.CatalogItemTypeProduct,
				Name:      "Test Item",
				BasePrice: decimal.NewFromFloat(25.50),
				Currency:  "INVALID",
			},
			wantErr: true,
			errMsg:  "currency must be exactly 3 characters",
		},
		{
			name: "invalid visibility",
			request: catalogRequests.CreateCatalogItemRequest{
				ItemType:   catalogModels.CatalogItemTypeProduct,
				Name:       "Test Item",
				BasePrice:  decimal.NewFromFloat(25.50),
				Visibility: "INVALID",
			},
			wantErr: true,
			errMsg:  "visibility must be one of: PRIVATE, ORG, NETWORK, PUBLIC",
		},
		{
			name: "description too long",
			request: catalogRequests.CreateCatalogItemRequest{
				ItemType:    catalogModels.CatalogItemTypeProduct,
				Name:        "Test Item",
				Description: string(make([]byte, 2001)), // 2001 characters
				BasePrice:   decimal.NewFromFloat(25.50),
			},
			wantErr: true,
			errMsg:  "description must be at most 2000 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCreateCatalogItemRequest(&tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateProductRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request catalogRequests.CreateProductRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid product request",
			request: catalogRequests.CreateProductRequest{
				CreateCatalogItemRequest: catalogRequests.CreateCatalogItemRequest{
					ItemType:  catalogModels.CatalogItemTypeProduct,
					Name:      "Organic Tomatoes",
					BasePrice: decimal.NewFromFloat(25.50),
				},
				Weight:        decimalPtr(decimal.NewFromFloat(2.5)),
				Perishable:    boolPtr(true),
				ShelfLifeDays: intPtr(7),
			},
			wantErr: false,
		},
		{
			name: "perishable product without shelf life",
			request: catalogRequests.CreateProductRequest{
				CreateCatalogItemRequest: catalogRequests.CreateCatalogItemRequest{
					ItemType:  catalogModels.CatalogItemTypeProduct,
					Name:      "Organic Tomatoes",
					BasePrice: decimal.NewFromFloat(25.50),
				},
				Perishable: boolPtr(true),
			},
			wantErr: true,
			errMsg:  "shelf_life_days is required for perishable products",
		},
		{
			name: "negative weight",
			request: catalogRequests.CreateProductRequest{
				CreateCatalogItemRequest: catalogRequests.CreateCatalogItemRequest{
					ItemType:  catalogModels.CatalogItemTypeProduct,
					Name:      "Test Product",
					BasePrice: decimal.NewFromFloat(25.50),
				},
				Weight: decimalPtr(decimal.NewFromFloat(-1.0)),
			},
			wantErr: true,
			errMsg:  "weight must be greater than or equal to 0",
		},
		{
			name: "wrong item type",
			request: catalogRequests.CreateProductRequest{
				CreateCatalogItemRequest: catalogRequests.CreateCatalogItemRequest{
					ItemType:  catalogModels.CatalogItemTypeService,
					Name:      "Test Product",
					BasePrice: decimal.NewFromFloat(25.50),
				},
			},
			wantErr: true,
			errMsg:  "item_type must be PRODUCT for product creation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := catalogRequests.ValidateCreateProductRequest(&tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateServiceRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request catalogRequests.CreateServiceRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid service request",
			request: catalogRequests.CreateServiceRequest{
				CreateCatalogItemRequest: catalogRequests.CreateCatalogItemRequest{
					ItemType:  catalogModels.CatalogItemTypeService,
					Name:      "Farm Consultation",
					BasePrice: decimal.NewFromFloat(150.00),
				},
				DurationMinutes: intPtr(120),
			},
			wantErr: false,
		},
		{
			name: "zero duration",
			request: catalogRequests.CreateServiceRequest{
				CreateCatalogItemRequest: catalogRequests.CreateCatalogItemRequest{
					ItemType:  catalogModels.CatalogItemTypeService,
					Name:      "Farm Consultation",
					BasePrice: decimal.NewFromFloat(150.00),
				},
				DurationMinutes: intPtr(0),
			},
			wantErr: true,
			errMsg:  "duration_minutes must be greater than 0",
		},
		{
			name: "wrong item type",
			request: catalogRequests.CreateServiceRequest{
				CreateCatalogItemRequest: catalogRequests.CreateCatalogItemRequest{
					ItemType:  catalogModels.CatalogItemTypeProduct,
					Name:      "Test Service",
					BasePrice: decimal.NewFromFloat(150.00),
				},
			},
			wantErr: true,
			errMsg:  "item_type must be SERVICE for service creation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := catalogRequests.ValidateCreateServiceRequest(&tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateLabourRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request catalogRequests.CreateLabourRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid labour request",
			request: catalogRequests.CreateLabourRequest{
				CreateCatalogItemRequest: catalogRequests.CreateCatalogItemRequest{
					ItemType:  catalogModels.CatalogItemTypeLabour,
					Name:      "Farm Worker",
					BasePrice: decimal.NewFromFloat(75.00),
				},
				SkillLevel: "experienced",
				HourlyRate: decimalPtr(decimal.NewFromFloat(75.00)),
			},
			wantErr: false,
		},
		{
			name: "negative hourly rate",
			request: catalogRequests.CreateLabourRequest{
				CreateCatalogItemRequest: catalogRequests.CreateCatalogItemRequest{
					ItemType:  catalogModels.CatalogItemTypeLabour,
					Name:      "Farm Worker",
					BasePrice: decimal.NewFromFloat(75.00),
				},
				HourlyRate: decimalPtr(decimal.NewFromFloat(-10.00)),
			},
			wantErr: true,
			errMsg:  "hourly_rate must be greater than or equal to 0",
		},
		{
			name: "wrong item type",
			request: catalogRequests.CreateLabourRequest{
				CreateCatalogItemRequest: catalogRequests.CreateCatalogItemRequest{
					ItemType:  catalogModels.CatalogItemTypeProduct,
					Name:      "Test Labour",
					BasePrice: decimal.NewFromFloat(75.00),
				},
			},
			wantErr: true,
			errMsg:  "item_type must be LABOUR for labour creation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := catalogRequests.ValidateCreateLabourRequest(&tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSearchCatalogRequest_ToSearchFilters(t *testing.T) {
	t.Run("converts with all fields", func(t *testing.T) {
		itemType := catalogModels.CatalogItemTypeProduct
		category := "vegetables"
		minPrice := decimal.NewFromFloat(10.00)
		maxPrice := decimal.NewFromFloat(100.00)
		sortBy := "name"
		sortOrder := "asc"

		req := &catalogRequests.SearchCatalogRequest{
			Query:     "organic tomatoes",
			ItemType:  &itemType,
			Category:  &category,
			MinPrice:  &minPrice,
			MaxPrice:  &maxPrice,
			Tags:      []string{"organic", "fresh"},
			Page:      2,
			PageSize:  50,
			SortBy:    &sortBy,
			SortOrder: &sortOrder,
		}

		filters := req.ToSearchFilters()

		assert.Equal(t, "organic tomatoes", filters["query"])
		assert.Equal(t, itemType, filters["item_type"])
		assert.Equal(t, category, filters["category"])
		assert.True(t, minPrice.Equal(filters["min_price"].(decimal.Decimal)))
		assert.True(t, maxPrice.Equal(filters["max_price"].(decimal.Decimal)))
		assert.Equal(t, []string{"organic", "fresh"}, filters["tags"])
		assert.Equal(t, 2, filters["page"])
		assert.Equal(t, 50, filters["page_size"])
		assert.Equal(t, "name", filters["sort_by"])
		assert.Equal(t, "asc", filters["sort_order"])
	})

	t.Run("applies defaults for missing values", func(t *testing.T) {
		req := &catalogRequests.SearchCatalogRequest{
			Query: "test",
		}

		filters := req.ToSearchFilters()

		assert.Equal(t, "test", filters["query"])
		assert.Equal(t, 1, filters["page"])
		assert.Equal(t, 20, filters["page_size"])
		assert.Equal(t, "relevance", filters["sort_by"])
		assert.Equal(t, "desc", filters["sort_order"])
	})

	t.Run("handles zero and negative page values", func(t *testing.T) {
		req := &catalogRequests.SearchCatalogRequest{
			Query:    "test",
			Page:     -1,
			PageSize: 0,
		}

		filters := req.ToSearchFilters()

		assert.Equal(t, 1, filters["page"])
		assert.Equal(t, 20, filters["page_size"])
	})
}

func TestListCatalogItemsRequest_ToCatalogFilter(t *testing.T) {
	t.Run("converts to catalog filter", func(t *testing.T) {
		itemType := catalogModels.CatalogItemTypeProduct
		category := "vegetables"
		isActive := true
		visibility := catalogModels.VisibilityOrg
		minPrice := decimal.NewFromFloat(10.00)
		maxPrice := decimal.NewFromFloat(100.00)
		search := "organic"
		orgID := "org-123"

		req := &catalogRequests.ListCatalogItemsRequest{
			Page:     1,
			PageSize: 20,
			CatalogFilter: catalogRequests.CatalogFilter{
				ItemType:       &itemType,
				Category:       &category,
				IsActive:       &isActive,
				Visibility:     &visibility,
				MinPrice:       &minPrice,
				MaxPrice:       &maxPrice,
				Search:         &search,
				Tags:           []string{"organic", "fresh"},
				OrganizationID: &orgID,
			},
		}

		filter := req.ToCatalogFilter()

		assert.Equal(t, &itemType, filter.ItemType)
		assert.Equal(t, &category, filter.Category)
		assert.Equal(t, &isActive, filter.IsActive)
		assert.Equal(t, &visibility, filter.Visibility)
		assert.True(t, minPrice.Equal(*filter.MinPrice))
		assert.True(t, maxPrice.Equal(*filter.MaxPrice))
		assert.Equal(t, &search, filter.Search)
		assert.Equal(t, []string{"organic", "fresh"}, filter.Tags)
		assert.Equal(t, &orgID, filter.OrganizationID)
	})
}

// Helper functions for tests

func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

// Helper validation functions (these would normally be in a separate validation package)

func validateCreateCatalogItemRequest(req *catalogRequests.CreateCatalogItemRequest) error {
	if req.ItemType == "" {
		return fmt.Errorf("item_type is required")
	}
	if req.ItemType != catalogModels.CatalogItemTypeProduct &&
		req.ItemType != catalogModels.CatalogItemTypeService &&
		req.ItemType != catalogModels.CatalogItemTypeLabour {
		return fmt.Errorf("item_type must be one of: PRODUCT, SERVICE, LABOUR")
	}
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(req.Name) > 255 {
		return fmt.Errorf("name must be at most 255 characters")
	}
	if len(req.Description) > 2000 {
		return fmt.Errorf("description must be at most 2000 characters")
	}
	if req.BasePrice.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("base_price must be greater than 0")
	}
	if req.Currency != "" && len(req.Currency) != 3 {
		return fmt.Errorf("currency must be exactly 3 characters")
	}
	if req.Visibility != "" &&
		req.Visibility != catalogModels.VisibilityPrivate &&
		req.Visibility != catalogModels.VisibilityOrg &&
		req.Visibility != catalogModels.VisibilityNetwork &&
		req.Visibility != catalogModels.VisibilityPublic {
		return fmt.Errorf("visibility must be one of: PRIVATE, ORG, NETWORK, PUBLIC")
	}

	return nil
}
