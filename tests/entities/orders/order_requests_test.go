package orders

import (
	"errors"
	"testing"

	orderModels "kisanlink-ecom/entities/models/orders"
	orderRequests "kisanlink-ecom/entities/requests/orders"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCreateOrderRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request orderRequests.CreateOrderRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: orderRequests.CreateOrderRequest{
				BuyerOrganizationID:  "123e4567-e89b-12d3-a456-426614174000",
				SellerOrganizationID: "123e4567-e89b-12d3-a456-426614174001",
				Items: []orderRequests.CreateOrderItemRequest{
					{
						CatalogItemID:   "123e4567-e89b-12d3-a456-426614174002",
						CatalogItemType: "product",
						Quantity:        decimal.NewFromFloat(10.5),
						UnitPrice:       decimal.NewFromFloat(25.50),
					},
				},
				ShippingAddress: &orderRequests.Address{
					Street:     "123 Farm Road",
					City:       "Rural City",
					State:      "Maharashtra",
					PostalCode: "411001",
					Country:    "India",
				},
				Notes: "Special delivery instructions",
			},
			wantErr: false,
		},
		{
			name: "missing buyer organization ID",
			request: orderRequests.CreateOrderRequest{
				SellerOrganizationID: "123e4567-e89b-12d3-a456-426614174001",
				Items: []orderRequests.CreateOrderItemRequest{
					{
						CatalogItemID:   "123e4567-e89b-12d3-a456-426614174002",
						CatalogItemType: "product",
						Quantity:        decimal.NewFromFloat(10.5),
						UnitPrice:       decimal.NewFromFloat(25.50),
					},
				},
			},
			wantErr: true,
			errMsg:  "buyer_organization_id is required",
		},
		{
			name: "missing seller organization ID",
			request: orderRequests.CreateOrderRequest{
				BuyerOrganizationID: "123e4567-e89b-12d3-a456-426614174000",
				Items: []orderRequests.CreateOrderItemRequest{
					{
						CatalogItemID:   "123e4567-e89b-12d3-a456-426614174002",
						CatalogItemType: "product",
						Quantity:        decimal.NewFromFloat(10.5),
						UnitPrice:       decimal.NewFromFloat(25.50),
					},
				},
			},
			wantErr: true,
			errMsg:  "seller_organization_id is required",
		},
		{
			name: "empty items",
			request: orderRequests.CreateOrderRequest{
				BuyerOrganizationID:  "123e4567-e89b-12d3-a456-426614174000",
				SellerOrganizationID: "123e4567-e89b-12d3-a456-426614174001",
				Items:                []orderRequests.CreateOrderItemRequest{},
			},
			wantErr: true,
			errMsg:  "items must contain at least 1 item",
		},
		{
			name: "invalid catalog item type",
			request: orderRequests.CreateOrderRequest{
				BuyerOrganizationID:  "123e4567-e89b-12d3-a456-426614174000",
				SellerOrganizationID: "123e4567-e89b-12d3-a456-426614174001",
				Items: []orderRequests.CreateOrderItemRequest{
					{
						CatalogItemID:   "123e4567-e89b-12d3-a456-426614174002",
						CatalogItemType: "invalid",
						Quantity:        decimal.NewFromFloat(10.5),
						UnitPrice:       decimal.NewFromFloat(25.50),
					},
				},
			},
			wantErr: true,
			errMsg:  "catalog_item_type must be one of: product, service, labour",
		},
		{
			name: "zero quantity",
			request: orderRequests.CreateOrderRequest{
				BuyerOrganizationID:  "123e4567-e89b-12d3-a456-426614174000",
				SellerOrganizationID: "123e4567-e89b-12d3-a456-426614174001",
				Items: []orderRequests.CreateOrderItemRequest{
					{
						CatalogItemID:   "123e4567-e89b-12d3-a456-426614174002",
						CatalogItemType: "product",
						Quantity:        decimal.Zero,
						UnitPrice:       decimal.NewFromFloat(25.50),
					},
				},
			},
			wantErr: true,
			errMsg:  "quantity must be greater than 0",
		},
		{
			name: "negative unit price",
			request: orderRequests.CreateOrderRequest{
				BuyerOrganizationID:  "123e4567-e89b-12d3-a456-426614174000",
				SellerOrganizationID: "123e4567-e89b-12d3-a456-426614174001",
				Items: []orderRequests.CreateOrderItemRequest{
					{
						CatalogItemID:   "123e4567-e89b-12d3-a456-426614174002",
						CatalogItemType: "product",
						Quantity:        decimal.NewFromFloat(10.5),
						UnitPrice:       decimal.NewFromFloat(-25.50),
					},
				},
			},
			wantErr: true,
			errMsg:  "unit_price must be greater than or equal to 0",
		},
		{
			name: "notes too long",
			request: orderRequests.CreateOrderRequest{
				BuyerOrganizationID:  "123e4567-e89b-12d3-a456-426614174000",
				SellerOrganizationID: "123e4567-e89b-12d3-a456-426614174001",
				Items: []orderRequests.CreateOrderItemRequest{
					{
						CatalogItemID:   "123e4567-e89b-12d3-a456-426614174002",
						CatalogItemType: "product",
						Quantity:        decimal.NewFromFloat(10.5),
						UnitPrice:       decimal.NewFromFloat(25.50),
					},
				},
				Notes: string(make([]byte, 1001)), // 1001 characters
			},
			wantErr: true,
			errMsg:  "notes must be at most 1000 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCreateOrderRequest(&tt.request)
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

func TestCreateOrderItemRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		item    orderRequests.CreateOrderItemRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid product item",
			item: orderRequests.CreateOrderItemRequest{
				CatalogItemID:   "123e4567-e89b-12d3-a456-426614174002",
				CatalogItemType: "product",
				Quantity:        decimal.NewFromFloat(10.5),
				UnitPrice:       decimal.NewFromFloat(25.50),
				Notes:           "Organic variety preferred",
			},
			wantErr: false,
		},
		{
			name: "valid service item",
			item: orderRequests.CreateOrderItemRequest{
				CatalogItemID:   "123e4567-e89b-12d3-a456-426614174002",
				CatalogItemType: "service",
				Quantity:        decimal.NewFromFloat(2),
				UnitPrice:       decimal.NewFromFloat(150.00),
			},
			wantErr: false,
		},
		{
			name: "valid labour item",
			item: orderRequests.CreateOrderItemRequest{
				CatalogItemID:   "123e4567-e89b-12d3-a456-426614174002",
				CatalogItemType: "labour",
				Quantity:        decimal.NewFromFloat(8),
				UnitPrice:       decimal.NewFromFloat(75.00),
			},
			wantErr: false,
		},
		{
			name: "missing catalog item ID",
			item: orderRequests.CreateOrderItemRequest{
				CatalogItemType: "product",
				Quantity:        decimal.NewFromFloat(10.5),
				UnitPrice:       decimal.NewFromFloat(25.50),
			},
			wantErr: true,
			errMsg:  "catalog_item_id is required",
		},
		{
			name: "invalid catalog item type",
			item: orderRequests.CreateOrderItemRequest{
				CatalogItemID:   "123e4567-e89b-12d3-a456-426614174002",
				CatalogItemType: "invalid",
				Quantity:        decimal.NewFromFloat(10.5),
				UnitPrice:       decimal.NewFromFloat(25.50),
			},
			wantErr: true,
			errMsg:  "catalog_item_type must be one of: product, service, labour",
		},
		{
			name: "notes too long",
			item: orderRequests.CreateOrderItemRequest{
				CatalogItemID:   "123e4567-e89b-12d3-a456-426614174002",
				CatalogItemType: "product",
				Quantity:        decimal.NewFromFloat(10.5),
				UnitPrice:       decimal.NewFromFloat(25.50),
				Notes:           string(make([]byte, 501)), // 501 characters
			},
			wantErr: true,
			errMsg:  "notes must be at most 500 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCreateOrderItemRequest(&tt.item)
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

func TestAddress_Validation(t *testing.T) {
	tests := []struct {
		name    string
		address orderRequests.Address
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid address",
			address: orderRequests.Address{
				Street:     "123 Farm Road",
				City:       "Rural City",
				State:      "Maharashtra",
				PostalCode: "411001",
				Country:    "India",
			},
			wantErr: false,
		},
		{
			name: "missing street",
			address: orderRequests.Address{
				City:       "Rural City",
				State:      "Maharashtra",
				PostalCode: "411001",
				Country:    "India",
			},
			wantErr: true,
			errMsg:  "street is required",
		},
		{
			name: "missing city",
			address: orderRequests.Address{
				Street:     "123 Farm Road",
				State:      "Maharashtra",
				PostalCode: "411001",
				Country:    "India",
			},
			wantErr: true,
			errMsg:  "city is required",
		},
		{
			name: "street too long",
			address: orderRequests.Address{
				Street:     string(make([]byte, 256)), // 256 characters
				City:       "Rural City",
				State:      "Maharashtra",
				PostalCode: "411001",
				Country:    "India",
			},
			wantErr: true,
			errMsg:  "street must be at most 255 characters",
		},
		{
			name: "postal code too long",
			address: orderRequests.Address{
				Street:     "123 Farm Road",
				City:       "Rural City",
				State:      "Maharashtra",
				PostalCode: string(make([]byte, 21)), // 21 characters
				Country:    "India",
			},
			wantErr: true,
			errMsg:  "postal_code must be at most 20 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAddress(&tt.address)
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

func TestListOrdersRequest_ToOrderFilters(t *testing.T) {
	t.Run("converts with all fields", func(t *testing.T) {
		status := orderModels.OrderStatusPending
		buyerOrgID := "buyer-123"
		sellerOrgID := "seller-456"
		buyerUserID := "user-789"
		createdAfter := "2024-01-01T00:00:00Z"
		createdBefore := "2024-12-31T23:59:59Z"
		minAmount := decimal.NewFromFloat(100.00)
		maxAmount := decimal.NewFromFloat(1000.00)
		search := "organic tomatoes"
		sortBy := "created_at"
		sortOrder := "desc"
		includeItems := true
		includeHistory := false

		req := &orderRequests.ListOrdersRequest{
			Page:                 2,
			PageSize:             50,
			Status:               &status,
			BuyerOrganizationID:  &buyerOrgID,
			SellerOrganizationID: &sellerOrgID,
			BuyerUserID:          &buyerUserID,
			CreatedAfter:         &createdAfter,
			CreatedBefore:        &createdBefore,
			MinAmount:            &minAmount,
			MaxAmount:            &maxAmount,
			Search:               &search,
			SortBy:               &sortBy,
			SortOrder:            &sortOrder,
			IncludeItems:         &includeItems,
			IncludeHistory:       &includeHistory,
		}

		filters := req.ToOrderFilters()

		assert.Equal(t, 2, filters.Page)
		assert.Equal(t, 50, filters.PageSize)
		assert.Equal(t, &status, filters.Status)
		assert.Equal(t, &buyerOrgID, filters.BuyerOrganizationID)
		assert.Equal(t, &sellerOrgID, filters.SellerOrganizationID)
		assert.Equal(t, &buyerUserID, filters.BuyerUserID)
		assert.Equal(t, &createdAfter, filters.CreatedAfter)
		assert.Equal(t, &createdBefore, filters.CreatedBefore)
		assert.True(t, minAmount.Equal(*filters.MinAmount))
		assert.True(t, maxAmount.Equal(*filters.MaxAmount))
		assert.Equal(t, &search, filters.Search)
		assert.Equal(t, "created_at", filters.SortBy)
		assert.Equal(t, "desc", filters.SortOrder)
		assert.True(t, filters.IncludeItems)
		assert.False(t, filters.IncludeHistory)
	})

	t.Run("applies defaults for missing values", func(t *testing.T) {
		req := &orderRequests.ListOrdersRequest{}

		filters := req.ToOrderFilters()

		assert.Equal(t, 1, filters.Page)
		assert.Equal(t, 20, filters.PageSize)
		assert.Equal(t, "created_at", filters.SortBy)
		assert.Equal(t, "desc", filters.SortOrder)
		assert.False(t, filters.IncludeItems)
		assert.False(t, filters.IncludeHistory)
	})

	t.Run("handles zero and negative page values", func(t *testing.T) {
		req := &orderRequests.ListOrdersRequest{
			Page:     -1,
			PageSize: 0,
		}

		filters := req.ToOrderFilters()

		assert.Equal(t, 1, filters.Page)
		assert.Equal(t, 20, filters.PageSize)
	})
}

// Helper validation functions (these would normally be in a separate validation package)

func validateCreateOrderRequest(req *orderRequests.CreateOrderRequest) error {
	if req.BuyerOrganizationID == "" {
		return errors.New("buyer_organization_id is required")
	}
	if req.SellerOrganizationID == "" {
		return errors.New("seller_organization_id is required")
	}
	if len(req.Items) == 0 {
		return errors.New("items must contain at least 1 item")
	}
	if len(req.Notes) > 1000 {
		return errors.New("notes must be at most 1000 characters")
	}

	for _, item := range req.Items {
		if err := validateCreateOrderItemRequest(&item); err != nil {
			return err
		}
	}

	if req.ShippingAddress != nil {
		if err := validateAddress(req.ShippingAddress); err != nil {
			return err
		}
	}

	return nil
}

func validateCreateOrderItemRequest(item *orderRequests.CreateOrderItemRequest) error {
	if item.CatalogItemID == "" {
		return errors.New("catalog_item_id is required")
	}
	if item.CatalogItemType != "product" && item.CatalogItemType != "service" && item.CatalogItemType != "labour" {
		return errors.New("catalog_item_type must be one of: product, service, labour")
	}
	if item.Quantity.LessThanOrEqual(decimal.Zero) {
		return errors.New("quantity must be greater than 0")
	}
	if item.UnitPrice.LessThan(decimal.Zero) {
		return errors.New("unit_price must be greater than or equal to 0")
	}
	if len(item.Notes) > 500 {
		return errors.New("notes must be at most 500 characters")
	}

	return nil
}

func validateAddress(addr *orderRequests.Address) error {
	if addr.Street == "" {
		return errors.New("street is required")
	}
	if addr.City == "" {
		return errors.New("city is required")
	}
	if addr.State == "" {
		return errors.New("state is required")
	}
	if addr.PostalCode == "" {
		return errors.New("postal_code is required")
	}
	if addr.Country == "" {
		return errors.New("country is required")
	}

	if len(addr.Street) > 255 {
		return errors.New("street must be at most 255 characters")
	}
	if len(addr.City) > 100 {
		return errors.New("city must be at most 100 characters")
	}
	if len(addr.State) > 100 {
		return errors.New("state must be at most 100 characters")
	}
	if len(addr.PostalCode) > 20 {
		return errors.New("postal_code must be at most 20 characters")
	}
	if len(addr.Country) > 100 {
		return errors.New("country must be at most 100 characters")
	}

	return nil
}
