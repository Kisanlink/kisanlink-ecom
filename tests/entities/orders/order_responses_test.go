package orders

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	orderModels "kisanlink-ecom/entities/models/orders"
	orderResponses "kisanlink-ecom/entities/responses/orders"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToOrderResponse(t *testing.T) {
	t.Run("converts complete order model to response", func(t *testing.T) {
		// Create test address
		address := orderModels.Address{
			Street:     "123 Farm Road",
			City:       "Rural City",
			State:      "Maharashtra",
			PostalCode: "411001",
			Country:    "India",
		}
		addressJSON, _ := json.Marshal(address)

		// Create test metadata
		metadata := map[string]interface{}{
			"priority": "high",
			"source":   "mobile_app",
		}
		metadataJSON, _ := json.Marshal(metadata)

		// Create test order
		now := time.Now()
		estimatedDelivery := now.Add(5 * 24 * time.Hour)

		order := &orderModels.Order{
			OrderNumber:           "ORD-2024-001",
			Status:                orderModels.OrderStatusPending,
			BuyerOrganizationID:   "buyer-123",
			SellerOrganizationID:  "seller-456",
			BuyerUserID:           "user-789",
			SubtotalAmount:        decimal.NewFromFloat(255.00),
			TaxAmount:             decimal.NewFromFloat(25.50),
			DiscountAmount:        decimal.NewFromFloat(0.00),
			ShippingAmount:        decimal.NewFromFloat(50.00),
			TotalAmount:           decimal.NewFromFloat(330.50),
			ShippingAddress:       addressJSON,
			EstimatedDeliveryDate: &estimatedDelivery,
			Notes:                 "Special delivery instructions",
			Metadata:              sql.NullString{String: string(metadataJSON), Valid: true},
		}
		order.ID = "order-123"
		order.CreatedAt = now
		order.UpdatedAt = now

		// Add test items
		item := orderModels.OrderItem{
			OrderID:         order.ID,
			CatalogItemID:   "catalog-123",
			CatalogItemType: "product",
			CatalogItemName: "Organic Tomatoes",
			CatalogItemSKU:  "TOM-ORG-001",
			Quantity:        decimal.NewFromFloat(10.5),
			UnitPrice:       decimal.NewFromFloat(25.50),
			TotalPrice:      decimal.NewFromFloat(267.75),
			TaxRate:         decimal.NewFromFloat(0.1000),
			TaxAmount:       decimal.NewFromFloat(26.78),
		}
		item.ID = "item-123"
		item.CreatedAt = now
		item.UpdatedAt = now
		order.Items = []orderModels.OrderItem{item}

		// Add test status history
		history := orderModels.OrderStatusHistory{
			OrderID:                 order.ID,
			FromStatus:              nil,
			ToStatus:                orderModels.OrderStatusPending,
			Reason:                  "Order created",
			ChangedByUserID:         "user-789",
			ChangedByOrganizationID: "buyer-123",
		}
		history.ID = "history-123"
		history.CreatedAt = now
		order.StatusHistory = []orderModels.OrderStatusHistory{history}

		// Convert to response
		response, err := orderResponses.ToOrderResponse(order)
		require.NoError(t, err)
		require.NotNil(t, response)

		// Verify basic fields
		assert.Equal(t, order.ID, response.ID)
		assert.Equal(t, order.OrderNumber, response.OrderNumber)
		assert.Equal(t, string(order.Status), response.Status)
		assert.Equal(t, order.BuyerOrganizationID, response.BuyerOrganizationID)
		assert.Equal(t, order.SellerOrganizationID, response.SellerOrganizationID)
		assert.Equal(t, order.BuyerUserID, response.BuyerUserID)

		// Verify amounts
		assert.True(t, order.SubtotalAmount.Equal(response.SubtotalAmount))
		assert.True(t, order.TaxAmount.Equal(response.TaxAmount))
		assert.True(t, order.DiscountAmount.Equal(response.DiscountAmount))
		assert.True(t, order.ShippingAmount.Equal(response.ShippingAmount))
		assert.True(t, order.TotalAmount.Equal(response.TotalAmount))

		// Verify shipping address
		require.NotNil(t, response.ShippingAddress)
		assert.Equal(t, address.Street, response.ShippingAddress.Street)
		assert.Equal(t, address.City, response.ShippingAddress.City)
		assert.Equal(t, address.State, response.ShippingAddress.State)
		assert.Equal(t, address.PostalCode, response.ShippingAddress.PostalCode)
		assert.Equal(t, address.Country, response.ShippingAddress.Country)

		// Verify metadata
		assert.Equal(t, metadata, response.Metadata)

		// Verify dates
		assert.Equal(t, order.EstimatedDeliveryDate, response.EstimatedDeliveryDate)
		assert.Equal(t, order.CreatedAt, response.CreatedAt)
		assert.Equal(t, order.UpdatedAt, response.UpdatedAt)

		// Verify items
		require.Len(t, response.Items, 1)
		responseItem := response.Items[0]
		assert.Equal(t, item.ID, responseItem.ID)
		assert.Equal(t, item.CatalogItemID, responseItem.CatalogItemID)
		assert.Equal(t, item.CatalogItemType, responseItem.CatalogItemType)
		assert.Equal(t, item.CatalogItemName, responseItem.CatalogItemName)
		assert.Equal(t, item.CatalogItemSKU, responseItem.CatalogItemSKU)
		assert.True(t, item.Quantity.Equal(responseItem.Quantity))
		assert.True(t, item.UnitPrice.Equal(responseItem.UnitPrice))
		assert.True(t, item.TotalPrice.Equal(responseItem.TotalPrice))

		// Verify status history
		require.Len(t, response.StatusHistory, 1)
		responseHistory := response.StatusHistory[0]
		assert.Equal(t, history.ID, responseHistory.ID)
		assert.Equal(t, history.OrderID, responseHistory.OrderID)
		assert.Equal(t, history.FromStatus, responseHistory.FromStatus)
		assert.Equal(t, string(history.ToStatus), responseHistory.ToStatus)
		assert.Equal(t, history.Reason, responseHistory.Reason)
	})

	t.Run("handles nil order", func(t *testing.T) {
		response, err := orderResponses.ToOrderResponse(nil)
		assert.NoError(t, err)
		assert.Nil(t, response)
	})

	t.Run("handles order with empty shipping address", func(t *testing.T) {
		order := &orderModels.Order{
			OrderNumber:          "ORD-2024-001",
			Status:               orderModels.OrderStatusPending,
			BuyerOrganizationID:  "buyer-123",
			SellerOrganizationID: "seller-456",
			BuyerUserID:          "user-789",
			ShippingAddress:      nil,
		}
		order.ID = "order-123"

		response, err := orderResponses.ToOrderResponse(order)
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Nil(t, response.ShippingAddress)
	})

	t.Run("handles order with invalid JSON in shipping address", func(t *testing.T) {
		order := &orderModels.Order{
			OrderNumber:          "ORD-2024-001",
			Status:               orderModels.OrderStatusPending,
			BuyerOrganizationID:  "buyer-123",
			SellerOrganizationID: "seller-456",
			BuyerUserID:          "user-789",
			ShippingAddress:      []byte("invalid json"),
		}
		order.ID = "order-123"

		response, err := orderResponses.ToOrderResponse(order)
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Nil(t, response.ShippingAddress)
	})

	t.Run("handles order with empty metadata", func(t *testing.T) {
		order := &orderModels.Order{
			OrderNumber:          "ORD-2024-001",
			Status:               orderModels.OrderStatusPending,
			BuyerOrganizationID:  "buyer-123",
			SellerOrganizationID: "seller-456",
			BuyerUserID:          "user-789",
			Metadata:             sql.NullString{Valid: false},
		}
		order.ID = "order-123"

		response, err := orderResponses.ToOrderResponse(order)
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Nil(t, response.Metadata)
	})
}

func TestToOrderItemResponse(t *testing.T) {
	t.Run("converts complete order item model to response", func(t *testing.T) {
		metadata := map[string]interface{}{
			"organic": true,
			"grade":   "A",
		}
		metadataJSON, _ := json.Marshal(metadata)

		now := time.Now()
		item := &orderModels.OrderItem{
			OrderID:         "order-123",
			CatalogItemID:   "catalog-123",
			CatalogItemType: "product",
			CatalogItemName: "Organic Tomatoes",
			CatalogItemSKU:  "TOM-ORG-001",
			Quantity:        decimal.NewFromFloat(10.5),
			UnitPrice:       decimal.NewFromFloat(25.50),
			TotalPrice:      decimal.NewFromFloat(267.75),
			TaxRate:         decimal.NewFromFloat(0.1000),
			TaxAmount:       decimal.NewFromFloat(26.78),
			DiscountRate:    decimal.NewFromFloat(0.0500),
			DiscountAmount:  decimal.NewFromFloat(13.39),
			Metadata:        sql.NullString{String: string(metadataJSON), Valid: true},
		}
		item.ID = "item-123"
		item.CreatedAt = now
		item.UpdatedAt = now

		response, err := orderResponses.ToOrderItemResponse(item)
		require.NoError(t, err)
		require.NotNil(t, response)

		assert.Equal(t, item.ID, response.ID)
		assert.Equal(t, item.OrderID, response.OrderID)
		assert.Equal(t, item.CatalogItemID, response.CatalogItemID)
		assert.Equal(t, item.CatalogItemType, response.CatalogItemType)
		assert.Equal(t, item.CatalogItemName, response.CatalogItemName)
		assert.Equal(t, item.CatalogItemSKU, response.CatalogItemSKU)
		assert.True(t, item.Quantity.Equal(response.Quantity))
		assert.True(t, item.UnitPrice.Equal(response.UnitPrice))
		assert.True(t, item.TotalPrice.Equal(response.TotalPrice))
		assert.True(t, item.TaxRate.Equal(response.TaxRate))
		assert.True(t, item.TaxAmount.Equal(response.TaxAmount))
		assert.True(t, item.DiscountRate.Equal(response.DiscountRate))
		assert.True(t, item.DiscountAmount.Equal(response.DiscountAmount))
		assert.Equal(t, metadata, response.Metadata)
		assert.Equal(t, item.CreatedAt, response.CreatedAt)
		assert.Equal(t, item.UpdatedAt, response.UpdatedAt)
	})

	t.Run("handles nil item", func(t *testing.T) {
		response, err := orderResponses.ToOrderItemResponse(nil)
		assert.NoError(t, err)
		assert.Nil(t, response)
	})

	t.Run("handles item with empty metadata", func(t *testing.T) {
		item := &orderModels.OrderItem{
			OrderID:         "order-123",
			CatalogItemID:   "catalog-123",
			CatalogItemType: "product",
			CatalogItemName: "Organic Tomatoes",
			Metadata:        sql.NullString{Valid: false},
		}
		item.ID = "item-123"

		response, err := orderResponses.ToOrderItemResponse(item)
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Nil(t, response.Metadata)
	})
}

func TestToOrderStatusHistoryResponse(t *testing.T) {
	t.Run("converts complete status history model to response", func(t *testing.T) {
		metadata := map[string]interface{}{
			"payment_method": "bank_transfer",
			"reference_id":   "PAY-123456",
		}
		metadataJSON, _ := json.Marshal(metadata)

		now := time.Now()
		fromStatus := "pending"
		history := &orderModels.OrderStatusHistory{
			OrderID:                 "order-123",
			FromStatus:              &fromStatus,
			ToStatus:                orderModels.OrderStatusConfirmed,
			Reason:                  "Payment confirmed by bank",
			ChangedByUserID:         "user-789",
			ChangedByOrganizationID: "buyer-123",
			Metadata:                sql.NullString{String: string(metadataJSON), Valid: true},
		}
		history.ID = "history-123"
		history.CreatedAt = now

		response, err := orderResponses.ToOrderStatusHistoryResponse(history)
		require.NoError(t, err)
		require.NotNil(t, response)

		assert.Equal(t, history.ID, response.ID)
		assert.Equal(t, history.OrderID, response.OrderID)
		assert.Equal(t, history.FromStatus, response.FromStatus)
		assert.Equal(t, string(history.ToStatus), response.ToStatus)
		assert.Equal(t, history.Reason, response.Reason)
		assert.Equal(t, history.ChangedByUserID, response.ChangedByUserID)
		assert.Equal(t, history.ChangedByOrganizationID, response.ChangedByOrganizationID)
		assert.Equal(t, metadata, response.Metadata)
		assert.Equal(t, history.CreatedAt, response.CreatedAt)
	})

	t.Run("handles nil history", func(t *testing.T) {
		response, err := orderResponses.ToOrderStatusHistoryResponse(nil)
		assert.NoError(t, err)
		assert.Nil(t, response)
	})

	t.Run("handles history with nil from status", func(t *testing.T) {
		history := &orderModels.OrderStatusHistory{
			OrderID:                 "order-123",
			FromStatus:              nil,
			ToStatus:                orderModels.OrderStatusPending,
			Reason:                  "Order created",
			ChangedByUserID:         "user-789",
			ChangedByOrganizationID: "buyer-123",
		}
		history.ID = "history-123"

		response, err := orderResponses.ToOrderStatusHistoryResponse(history)
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Nil(t, response.FromStatus)
		assert.Equal(t, string(history.ToStatus), response.ToStatus)
	})
}

func TestToOrderSummaryResponse(t *testing.T) {
	t.Run("converts order model to summary response", func(t *testing.T) {
		now := time.Now()
		estimatedDelivery := now.Add(5 * 24 * time.Hour)

		order := &orderModels.Order{
			OrderNumber:           "ORD-2024-001",
			Status:                orderModels.OrderStatusPending,
			BuyerOrganizationID:   "buyer-123",
			SellerOrganizationID:  "seller-456",
			TotalAmount:           decimal.NewFromFloat(330.50),
			EstimatedDeliveryDate: &estimatedDelivery,
			Items: []orderModels.OrderItem{
				{CatalogItemID: "item1"},
				{CatalogItemID: "item2"},
				{CatalogItemID: "item3"},
			},
		}
		order.ID = "order-123"
		order.CreatedAt = now

		response := orderResponses.ToOrderSummaryResponse(order)
		require.NotNil(t, response)

		assert.Equal(t, order.ID, response.ID)
		assert.Equal(t, order.OrderNumber, response.OrderNumber)
		assert.Equal(t, string(order.Status), response.Status)
		assert.True(t, order.TotalAmount.Equal(response.TotalAmount))
		assert.Equal(t, 3, response.ItemCount)
		assert.Equal(t, order.BuyerOrganizationID, response.BuyerOrgID)
		assert.Equal(t, order.SellerOrganizationID, response.SellerOrgID)
		assert.Equal(t, order.CreatedAt, response.CreatedAt)
		assert.Equal(t, order.EstimatedDeliveryDate, response.EstimatedDelivery)
	})

	t.Run("handles nil order", func(t *testing.T) {
		response := orderResponses.ToOrderSummaryResponse(nil)
		assert.Nil(t, response)
	})

	t.Run("handles order with no items", func(t *testing.T) {
		order := &orderModels.Order{
			OrderNumber:          "ORD-2024-001",
			Status:               orderModels.OrderStatusPending,
			BuyerOrganizationID:  "buyer-123",
			SellerOrganizationID: "seller-456",
			TotalAmount:          decimal.NewFromFloat(330.50),
			Items:                []orderModels.OrderItem{},
		}
		order.ID = "order-123"

		response := orderResponses.ToOrderSummaryResponse(order)
		require.NotNil(t, response)
		assert.Equal(t, 0, response.ItemCount)
	})
}

func TestToOrderResponseList(t *testing.T) {
	t.Run("converts list of orders", func(t *testing.T) {
		now := time.Now()
		orders := []*orderModels.Order{
			{
				OrderNumber:          "ORD-2024-001",
				Status:               orderModels.OrderStatusPending,
				BuyerOrganizationID:  "buyer-123",
				SellerOrganizationID: "seller-456",
				TotalAmount:          decimal.NewFromFloat(100.00),
			},
			{
				OrderNumber:          "ORD-2024-002",
				Status:               orderModels.OrderStatusConfirmed,
				BuyerOrganizationID:  "buyer-123",
				SellerOrganizationID: "seller-456",
				TotalAmount:          decimal.NewFromFloat(200.00),
			},
		}
		orders[0].ID = "order-1"
		orders[0].CreatedAt = now
		orders[1].ID = "order-2"
		orders[1].CreatedAt = now

		responses, err := orderResponses.ToOrderResponseList(orders)
		require.NoError(t, err)
		require.Len(t, responses, 2)

		assert.Equal(t, orders[0].ID, responses[0].ID)
		assert.Equal(t, orders[0].OrderNumber, responses[0].OrderNumber)
		assert.Equal(t, orders[1].ID, responses[1].ID)
		assert.Equal(t, orders[1].OrderNumber, responses[1].OrderNumber)
	})

	t.Run("handles empty list", func(t *testing.T) {
		responses, err := orderResponses.ToOrderResponseList([]*orderModels.Order{})
		assert.NoError(t, err)
		assert.Empty(t, responses)
	})

	t.Run("handles nil list", func(t *testing.T) {
		responses, err := orderResponses.ToOrderResponseList(nil)
		assert.NoError(t, err)
		assert.Empty(t, responses)
	})
}

func TestToOrderSummaryResponseList(t *testing.T) {
	t.Run("converts list of orders to summaries", func(t *testing.T) {
		now := time.Now()
		orders := []*orderModels.Order{
			{
				OrderNumber:          "ORD-2024-001",
				Status:               orderModels.OrderStatusPending,
				BuyerOrganizationID:  "buyer-123",
				SellerOrganizationID: "seller-456",
				TotalAmount:          decimal.NewFromFloat(100.00),
				Items:                []orderModels.OrderItem{{}, {}}, // 2 items
			},
			{
				OrderNumber:          "ORD-2024-002",
				Status:               orderModels.OrderStatusConfirmed,
				BuyerOrganizationID:  "buyer-123",
				SellerOrganizationID: "seller-456",
				TotalAmount:          decimal.NewFromFloat(200.00),
				Items:                []orderModels.OrderItem{{}}, // 1 item
			},
		}
		orders[0].ID = "order-1"
		orders[0].CreatedAt = now
		orders[1].ID = "order-2"
		orders[1].CreatedAt = now

		responses := orderResponses.ToOrderSummaryResponseList(orders)
		require.Len(t, responses, 2)

		assert.Equal(t, orders[0].ID, responses[0].ID)
		assert.Equal(t, 2, responses[0].ItemCount)
		assert.Equal(t, orders[1].ID, responses[1].ID)
		assert.Equal(t, 1, responses[1].ItemCount)
	})

	t.Run("handles empty list", func(t *testing.T) {
		responses := orderResponses.ToOrderSummaryResponseList([]*orderModels.Order{})
		assert.Empty(t, responses)
	})

	t.Run("handles nil list", func(t *testing.T) {
		responses := orderResponses.ToOrderSummaryResponseList(nil)
		assert.Empty(t, responses)
	})
}

func TestValidationErrors(t *testing.T) {
	t.Run("creates validation error", func(t *testing.T) {
		err := orderResponses.NewValidationError("field_name", "Field is required", "invalid_value")
		assert.Equal(t, "field_name", err.Field)
		assert.Equal(t, "Field is required", err.Message)
		assert.Equal(t, "invalid_value", err.Value)
	})

	t.Run("creates validation errors collection", func(t *testing.T) {
		err1 := orderResponses.NewValidationError("field1", "Field 1 is required", nil)
		err2 := orderResponses.NewValidationError("field2", "Field 2 is invalid", "bad_value")

		errors := orderResponses.NewValidationErrors(err1, err2)
		assert.Len(t, errors.Errors, 2)
		assert.Equal(t, "Field 1 is required", errors.Error())
	})

	t.Run("handles empty validation errors", func(t *testing.T) {
		errors := orderResponses.NewValidationErrors()
		assert.Empty(t, errors.Errors)
		assert.Equal(t, "validation failed", errors.Error())
	})
}
