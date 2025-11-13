package data

import (
	"database/sql"
	catalogModels "kisanlink-ecom/entities/models/catalog"
	"kisanlink-ecom/entities/models/orders"
	catalogRequests "kisanlink-ecom/entities/requests/catalog"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/shopspring/decimal"
)

// Test data for catalog items

// Sample Product Data
func GetSampleProduct() *catalogModels.Product {
	return &catalogModels.Product{
		CatalogItem: catalogModels.CatalogItem{
			Name:        "Organic Tomatoes",
			Description: "Fresh organic tomatoes from local farms",
			Category:    "vegetables",
			Subcategory: "tomatoes",
			BasePrice:   decimal.NewFromFloat(50.0),
			Currency:    "INR",
			ItemType:    catalogModels.CatalogItemTypeProduct,
			IsActive:    true,
			Visibility:  catalogModels.VisibilityOrg,
		},
		Perishable: true,
	}
}

// Sample Service Data
func GetSampleService() *catalogModels.Service {
	return &catalogModels.Service{
		CatalogItem: catalogModels.CatalogItem{
			Name:        "Agricultural Consulting",
			Description: "Expert agricultural consulting services",
			Category:    "consulting",
			Subcategory: "agriculture",
			BasePrice:   decimal.NewFromFloat(100.0),
			Currency:    "INR",
			ItemType:    catalogModels.CatalogItemTypeService,
			IsActive:    true,
			Visibility:  catalogModels.VisibilityNetwork,
		},
		DurationMinutes: intPtr(60),
	}
}

// Sample Labour Data
func GetSampleLabour() *catalogModels.Labour {
	return &catalogModels.Labour{
		CatalogItem: catalogModels.CatalogItem{
			Name:        "Farm Worker",
			Description: "Experienced farm worker for seasonal work",
			Category:    "agricultural-labor",
			Subcategory: "seasonal",
			BasePrice:   decimal.NewFromFloat(15.0),
			Currency:    "INR",
			ItemType:    catalogModels.CatalogItemTypeLabour,
			IsActive:    true,
			Visibility:  catalogModels.VisibilityPrivate,
		},
		SkillLevel: "intermediate",
		HourlyRate: decimalPtr(decimal.NewFromFloat(15.0)),
	}
}

// Sample Create Requests
func GetSampleCreateProductRequest() catalogRequests.CreateCatalogItemRequest {
	return catalogRequests.CreateCatalogItemRequest{
		ItemType:      catalogModels.CatalogItemTypeProduct,
		Name:          "Organic Tomatoes",
		Description:   "Fresh organic tomatoes from local farms",
		Category:      "vegetables",
		Subcategory:   "tomatoes",
		BasePrice:     decimal.NewFromFloat(50.0),
		Currency:      "INR",
		UnitOfMeasure: "kg",
		Visibility:    catalogModels.VisibilityOrg,
	}
}

func GetSampleCreateServiceRequest() catalogRequests.CreateServiceRequest {
	return catalogRequests.CreateServiceRequest{
		CreateCatalogItemRequest: catalogRequests.CreateCatalogItemRequest{
			ItemType:    catalogModels.CatalogItemTypeService,
			Name:        "Agricultural Consulting",
			Description: "Expert agricultural consulting services",
			Category:    "consulting",
			BasePrice:   decimal.NewFromFloat(100.0),
			Currency:    "INR",
			Visibility:  catalogModels.VisibilityNetwork,
		},
		DurationMinutes: intPtr(60),
	}
}

func GetSampleCreateLabourRequest() catalogRequests.CreateLabourRequest {
	return catalogRequests.CreateLabourRequest{
		CreateCatalogItemRequest: catalogRequests.CreateCatalogItemRequest{
			ItemType:    catalogModels.CatalogItemTypeLabour,
			Name:        "Farm Worker",
			Description: "Experienced farm worker for seasonal work",
			Category:    "agricultural-labor",
			BasePrice:   decimal.NewFromFloat(15.0),
			Currency:    "INR",
			Visibility:  catalogModels.VisibilityPrivate,
		},
		SkillLevel: "intermediate",
		HourlyRate: decimalPtr(decimal.NewFromFloat(15.0)),
	}
}

// Sample Update Requests
func GetSampleUpdateRequest() catalogRequests.UpdateCatalogItemRequest {
	return catalogRequests.UpdateCatalogItemRequest{
		Name:        stringPtr("Updated Product Name"),
		Description: stringPtr("Updated description"),
		BasePrice:   decimalPtr(decimal.NewFromFloat(75.0)),
		IsActive:    boolPtr(true),
	}
}

// Sample Catalog Filters
func GetSampleCatalogFilter() catalogRequests.CatalogFilter {
	itemType := catalogModels.CatalogItemTypeProduct
	category := "vegetables"
	isActive := true
	minPrice := decimal.NewFromFloat(10.0)
	maxPrice := decimal.NewFromFloat(100.0)

	return catalogRequests.CatalogFilter{
		ItemType: &itemType,
		Category: &category,
		IsActive: &isActive,
		MinPrice: &minPrice,
		MaxPrice: &maxPrice,
	}
}

// Helper functions for pointers
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func floatPtr(f float64) *float64 {
	return &f
}

func boolPtr(b bool) *bool {
	return &b
}

// Sample catalog items list
func GetSampleCatalogItems() []*catalogModels.CatalogItem {
	return []*catalogModels.CatalogItem{
		{
			Name:      "Organic Tomatoes",
			ItemType:  catalogModels.CatalogItemTypeProduct,
			Category:  "vegetables",
			BasePrice: decimal.NewFromFloat(50.0),
			IsActive:  true,
		},
		{
			Name:      "Agricultural Consulting",
			ItemType:  catalogModels.CatalogItemTypeService,
			Category:  "consulting",
			BasePrice: decimal.NewFromFloat(100.0),
			IsActive:  true,
		},
		{
			Name:      "Farm Worker",
			ItemType:  catalogModels.CatalogItemTypeLabour,
			Category:  "agricultural-labor",
			BasePrice: decimal.NewFromFloat(15.0),
			IsActive:  true,
		},
	}
}

// Sample Inventory Lot Data
func GetTestInventoryLot() *catalogModels.InventoryLot {
	harvestDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	expiryDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	lotPrice := decimal.NewFromFloat(25.50)

	catalogItemID := "prod_123"
	return &catalogModels.InventoryLot{
		BaseModel:      *base.NewBaseModel("LOT", "large"),
		CatalogItemID:  &catalogItemID,
		OrganizationID: "org_456",
		LotNumber:      "LOT-2024-001",
		BatchNumber:    "BATCH-001",
		Quantity:       decimal.NewFromFloat(100.0),
		AvailableQty:   decimal.NewFromFloat(85.0),
		ReservedQty:    decimal.NewFromFloat(10.0),
		UnitOfMeasure:  "kg",
		QualityGrade:   "A",
		HarvestDate:    &harvestDate,
		ExpiryDate:     &expiryDate,
		Warehouse:      "Warehouse A",
		Zone:           "Section 1",
		UnitCost:       lotPrice,
		Status:         catalogModels.LotStatusActive,
		Metadata:       sql.NullString{String: `{"supplier": "Farm ABC", "notes": "Premium quality"}`, Valid: true},
	}
}

// Sample Inventory Lots List
func GetTestInventoryLots() []*catalogModels.InventoryLot {
	lot1 := GetTestInventoryLot()

	catalogItemID2 := "prod_456"
	lot2 := &catalogModels.InventoryLot{
		BaseModel:      *base.NewBaseModel("LOT", "large"),
		CatalogItemID:  &catalogItemID2,
		OrganizationID: "org_456",
		LotNumber:      "LOT-2024-002",
		BatchNumber:    "BATCH-002",
		Quantity:       decimal.NewFromFloat(50.0),
		AvailableQty:   decimal.NewFromFloat(50.0),
		ReservedQty:    decimal.Zero,
		UnitOfMeasure:  "kg",
		QualityGrade:   "B",
		Status:         catalogModels.LotStatusActive,
		Warehouse:      "Warehouse B",
		Zone:           "Section 2",
		Metadata:       sql.NullString{String: `{"supplier": "Farm XYZ"}`, Valid: true},
	}

	return []*catalogModels.InventoryLot{lot1, lot2}
}

// CreateTestOrder creates a test order with sample data
func CreateTestOrder() *orders.Order {
	return &orders.Order{
		BaseModel:            *base.NewBaseModel("ORD", "large"),
		OrderNumber:          "ORD-2024-001",
		Status:               orders.OrderStatusPending,
		BuyerOrganizationID:  "org-buyer-123",
		SellerOrganizationID: "org-seller-456",
		BuyerUserID:          "user-123",
		SubtotalAmount:       decimal.NewFromFloat(100.00),
		TaxAmount:            decimal.NewFromFloat(10.00),
		DiscountAmount:       decimal.Zero,
		ShippingAmount:       decimal.NewFromFloat(5.00),
		TotalAmount:          decimal.NewFromFloat(115.00),
		Notes:                "Test order for unit testing",
		Items: []orders.OrderItem{
			{
				BaseModel:       *base.NewBaseModel("ITM", "large"),
				CatalogItemID:   "item-123",
				CatalogItemType: "product",
				CatalogItemName: "Test Product",
				CatalogItemSKU:  "SKU-123",
				Quantity:        decimal.NewFromFloat(2.0),
				UnitPrice:       decimal.NewFromFloat(50.00),
				TotalPrice:      decimal.NewFromFloat(100.00),
				TaxRate:         decimal.NewFromFloat(0.10),
				TaxAmount:       decimal.NewFromFloat(10.00),
			},
		},
	}
}
