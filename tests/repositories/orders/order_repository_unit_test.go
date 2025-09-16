package orders

import (
	"context"
	"testing"

	"kisanlink-ecom/entities/models/orders"
	ordersRequests "kisanlink-ecom/entities/requests/orders"
	ordersRepo "kisanlink-ecom/internal/repositories/orders"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// TestOrderRepository_ValidateOrderItems_BasicValidation tests basic validation without dependencies
func TestOrderRepository_ValidateOrderItems_BasicValidation(t *testing.T) {
	repo := &ordersRepo.OrderRepository{}

	tests := []struct {
		name          string
		items         []ordersRequests.CreateOrderItemRequest
		sellerOrgID   string
		expectedError string
	}{
		{
			name: "valid items",
			items: []ordersRequests.CreateOrderItemRequest{
				{
					CatalogItemID:   "catalog-item-1",
					CatalogItemType: "product",
					Quantity:        decimal.NewFromFloat(2.0),
					UnitPrice:       decimal.NewFromFloat(50.25),
				},
			},
			sellerOrgID: "seller-org-1",
		},
		{
			name: "empty catalog ID",
			items: []ordersRequests.CreateOrderItemRequest{
				{
					CatalogItemID:   "",
					CatalogItemType: "product",
					Quantity:        decimal.NewFromFloat(1.0),
					UnitPrice:       decimal.NewFromFloat(10.0),
				},
			},
			sellerOrgID:   "seller-org-1",
			expectedError: "item 1: catalog item ID is required",
		},
		{
			name: "zero quantity",
			items: []ordersRequests.CreateOrderItemRequest{
				{
					CatalogItemID:   "catalog-item-1",
					CatalogItemType: "product",
					Quantity:        decimal.NewFromInt(0),
					UnitPrice:       decimal.NewFromFloat(10.0),
				},
			},
			sellerOrgID:   "seller-org-1",
			expectedError: "item 1: quantity must be greater than 0",
		},
		{
			name: "negative price",
			items: []ordersRequests.CreateOrderItemRequest{
				{
					CatalogItemID:   "catalog-item-1",
					CatalogItemType: "product",
					Quantity:        decimal.NewFromFloat(1.0),
					UnitPrice:       decimal.NewFromFloat(-10.0),
				},
			},
			sellerOrgID:   "seller-org-1",
			expectedError: "item 1: unit price cannot be negative",
		},
		{
			name: "invalid item type",
			items: []ordersRequests.CreateOrderItemRequest{
				{
					CatalogItemID:   "catalog-item-1",
					CatalogItemType: "invalid",
					Quantity:        decimal.NewFromFloat(1.0),
					UnitPrice:       decimal.NewFromFloat(10.0),
				},
			},
			sellerOrgID:   "seller-org-1",
			expectedError: "item 1: invalid catalog item type 'invalid'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.ValidateOrderItems(context.Background(), tt.items, tt.sellerOrgID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestOrderRepository_ReleaseInventory_WithoutDependencies tests inventory release without dependencies
func TestOrderRepository_ReleaseInventory_WithoutDependencies(t *testing.T) {
	repo := &ordersRepo.OrderRepository{}

	items := []ordersRequests.CreateOrderItemRequest{
		{
			CatalogItemID:   "catalog-item-1",
			CatalogItemType: "product",
			Quantity:        decimal.NewFromFloat(2.0),
			UnitPrice:       decimal.NewFromFloat(50.25),
		},
	}

	err := repo.ReleaseInventory(context.Background(), items)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "inventory repository not configured")
}

// TestOrderRepository_ReserveInventory_WithoutDependencies tests inventory reservation without dependencies
func TestOrderRepository_ReserveInventory_WithoutDependencies(t *testing.T) {
	repo := &ordersRepo.OrderRepository{}

	items := []ordersRequests.CreateOrderItemRequest{
		{
			CatalogItemID:   "catalog-item-1",
			CatalogItemType: "product",
			Quantity:        decimal.NewFromFloat(2.0),
			UnitPrice:       decimal.NewFromFloat(50.25),
		},
	}

	err := repo.ReserveInventory(context.Background(), items)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "inventory repository not configured")
}

// TestOrder_CalculateTotal tests order total calculation
func TestOrder_CalculateTotal(t *testing.T) {
	order := orders.NewOrder("buyer-org-1", "seller-org-1", "user-1")

	// Add items
	item1 := orders.NewOrderItem(
		order.ID,
		"catalog-item-1",
		"product",
		"Test Product 1",
		"SKU-001",
		decimal.NewFromFloat(2),
		decimal.NewFromFloat(50.25),
	)

	item2 := orders.NewOrderItem(
		order.ID,
		"catalog-item-2",
		"service",
		"Test Service 1",
		"SKU-002",
		decimal.NewFromFloat(1),
		decimal.NewFromFloat(100.0),
	)

	order.Items = []orders.OrderItem{*item1, *item2}
	order.TaxAmount = decimal.NewFromFloat(15.05)
	order.ShippingAmount = decimal.NewFromFloat(10.0)
	order.DiscountAmount = decimal.NewFromFloat(5.0)

	order.CalculateTotal()

	expectedSubtotal := decimal.NewFromFloat(200.50) // 2*50.25 + 1*100.0
	expectedTotal := decimal.NewFromFloat(220.55)    // 200.50 + 15.05 + 10.0 - 5.0

	assert.True(t, order.SubtotalAmount.Equal(expectedSubtotal))
	assert.True(t, order.TotalAmount.Equal(expectedTotal))
}

// TestOrder_StatusTransitions tests order status transitions
func TestOrder_StatusTransitions(t *testing.T) {
	order := orders.NewOrder("buyer-org-1", "seller-org-1", "user-1")

	// Test valid transitions
	assert.True(t, order.CanTransitionTo(orders.OrderStatusConfirmed))
	assert.True(t, order.CanTransitionTo(orders.OrderStatusCancelled))
	assert.False(t, order.CanTransitionTo(orders.OrderStatusDelivered))

	// Update status
	err := order.UpdateStatus(orders.OrderStatusConfirmed, "Order confirmed", "user-1", "seller-org-1")
	assert.NoError(t, err)
	assert.Equal(t, orders.OrderStatusConfirmed, order.Status)
	assert.Len(t, order.StatusHistory, 1)

	// Test invalid transition
	err = order.UpdateStatus(orders.OrderStatusDelivered, "Invalid transition", "user-1", "seller-org-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status transition")
}

// TestOrder_AddRemoveItems tests adding and removing items from order
func TestOrder_AddRemoveItems(t *testing.T) {
	order := orders.NewOrder("buyer-org-1", "seller-org-1", "user-1")

	item := orders.NewOrderItem(
		order.ID,
		"catalog-item-1",
		"product",
		"Test Product",
		"SKU-001",
		decimal.NewFromFloat(2),
		decimal.NewFromFloat(50.25),
	)
	item.ID = "item-1"

	// Add item
	order.AddItem(item)
	assert.Len(t, order.Items, 1)
	assert.True(t, order.SubtotalAmount.Equal(decimal.NewFromFloat(100.50)))

	// Remove item
	order.RemoveItem("item-1")
	assert.Len(t, order.Items, 0)
	assert.True(t, order.SubtotalAmount.Equal(decimal.Zero))
}

// TestOrder_GetValidTransitions tests getting valid status transitions
func TestOrder_GetValidTransitions(t *testing.T) {
	order := orders.NewOrder("buyer-org-1", "seller-org-1", "user-1")

	transitions := order.GetValidTransitions()
	assert.Contains(t, transitions, orders.OrderStatusConfirmed)
	assert.Contains(t, transitions, orders.OrderStatusCancelled)
	assert.NotContains(t, transitions, orders.OrderStatusDelivered)

	// Change status and test again
	order.Status = orders.OrderStatusCompleted
	transitions = order.GetValidTransitions()
	assert.Empty(t, transitions) // Terminal state
}
