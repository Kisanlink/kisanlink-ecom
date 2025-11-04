package data

import (
	"time"

	"kisanlink-ecom/entities/models/catalog"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/shopspring/decimal"
)

// CreateTestPublishState creates a test publish state with default values
func CreateTestPublishState() *catalog.PublishState {
	now := time.Now()
	return &catalog.PublishState{
		BaseModel:     *base.NewBaseModel("PUB", "medium"),
		ProductID:     "PROD00000001",
		FPOAccessList: []string{"ORGN00000002", "ORGN00000003"},
		DeliveryCosts: map[string]decimal.Decimal{
			"ORGN00000002": decimal.NewFromFloat(50.00),
			"ORGN00000003": decimal.NewFromFloat(75.00),
		},
		PlatformFeePercent: decimal.NewFromFloat(10.00),
		PublishedAt:        &now,
		PublishedBy:        "admin-user-123",
		CreatedBy:          "admin-user-123",
		UpdatedBy:          "admin-user-123",
	}
}

// CreateTestPublishStateWithProductID creates a test publish state with specific product ID
func CreateTestPublishStateWithProductID(productID string) *catalog.PublishState {
	ps := CreateTestPublishState()
	ps.ProductID = productID
	return ps
}

// CreateTestPublishStateWithFPOs creates a test publish state with specific FPO IDs
func CreateTestPublishStateWithFPOs(productID string, fpoIDs []string, deliveryCosts map[string]decimal.Decimal) *catalog.PublishState {
	now := time.Now()
	return &catalog.PublishState{
		BaseModel:          *base.NewBaseModel("PUB", "medium"),
		ProductID:          productID,
		FPOAccessList:      fpoIDs,
		DeliveryCosts:      deliveryCosts,
		PlatformFeePercent: decimal.NewFromFloat(10.00),
		PublishedAt:        &now,
		PublishedBy:        "admin-user-123",
		CreatedBy:          "admin-user-123",
		UpdatedBy:          "admin-user-123",
	}
}

// CreateTestActiveProduct creates an active product for testing
func CreateTestActiveProduct() *catalog.CatalogItem {
	return &catalog.CatalogItem{
		BaseModel:   *base.NewBaseModel("PROD", "large"),
		Name:        "Test Organic Wheat",
		Description: "Premium organic wheat for testing",
		Category:    "grains",
		Subcategory: "wheat",
		BasePrice:   decimal.NewFromFloat(1000.00),
		Currency:    "INR",
		ItemType:    catalog.CatalogItemTypeProduct,
		IsActive:    true,
		Visibility:  catalog.VisibilityOrg,
	}
}

// CreateTestInactiveProduct creates an inactive product for testing
func CreateTestInactiveProduct() *catalog.CatalogItem {
	product := CreateTestActiveProduct()
	product.IsActive = false
	return product
}

// CreateTestProductWithFPOPricing creates a product with FPO pricing for testing
func CreateTestProductWithFPOPricing() *catalog.ProductWithFPOPricing {
	product := CreateTestActiveProduct()
	basePrice := decimal.NewFromFloat(1000.00)
	deliveryCost := decimal.NewFromFloat(50.00)
	commission := decimal.NewFromFloat(100.00) // 10% of 1000
	retailPrice := decimal.NewFromFloat(1150.00)

	priceLock := time.Now().Add(30 * time.Minute)

	return &catalog.ProductWithFPOPricing{
		Product: catalog.Product{
			CatalogItem: *product,
		},
		Pricing: catalog.FPOPricingDetail{
			BasePrice:        basePrice,
			DeliveryCost:     deliveryCost,
			CommissionAmount: commission,
			RetailPrice:      retailPrice,
			PriceLockedUntil: &priceLock,
		},
	}
}

// CreateTestDeliveryCosts creates test delivery costs map
func CreateTestDeliveryCosts(fpoIDs []string) map[string]decimal.Decimal {
	costs := make(map[string]decimal.Decimal)
	for i, fpoID := range fpoIDs {
		costs[fpoID] = decimal.NewFromFloat(float64(50 + (i * 10)))
	}
	return costs
}

// CreateTestFPOIDs creates test FPO organization IDs
func CreateTestFPOIDs(count int) []string {
	fpoIDs := make([]string, count)
	for i := 0; i < count; i++ {
		fpoIDs[i] = CreateTestFPOID(i + 2) // Start from ORGN00000002
	}
	return fpoIDs
}

// CreateTestFPOID creates a test FPO organization ID
func CreateTestFPOID(index int) string {
	return "ORGN" + padZeros(index, 8)
}

// Helper function to pad zeros
func padZeros(num int, totalLen int) string {
	result := ""
	for i := 0; i < totalLen; i++ {
		result += "0"
	}
	numStr := ""
	temp := num
	for temp > 0 {
		numStr = string(rune('0'+(temp%10))) + numStr
		temp /= 10
	}
	if len(numStr) > totalLen {
		return numStr
	}
	return result[:totalLen-len(numStr)] + numStr
}
