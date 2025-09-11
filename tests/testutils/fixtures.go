package testutils

import (
    "time"

    catalogModels "kisanlink-ecom/entities/models/catalog"
    orderModels "kisanlink-ecom/entities/models/orders"
    catalogRequests "kisanlink-ecom/entities/requests/catalog"
    orderRequests "kisanlink-ecom/entities/requests/orders"

    "github.com/shopspring/decimal"
)

// Test constants
const (
    TestUserID        = "user-123"
    TestOrgID         = "org-123"
    TestSellerOrgID   = "seller-org-456"
    TestOrderID       = "order-789"
    TestCatalogItemID = "item-456"
    TestSKU           = "TEST-SKU-001"
    TestOrderNumber   = "ORD-2024-001"
)

// Order fixtures
func CreateTestOrder() *orderModels.Order {
    return &orderModels.Order{
        OrderNumber:          TestOrderNumber,
        BuyerOrganizationID:  TestOrgID,
        SellerOrganizationID: TestSellerOrgID,
        BuyerUserID:          TestUserID,
        Status:               orderModels.OrderStatusPending,
        TotalAmount:          decimal.NewFromFloat(1000.0),
        Items:                []orderModels.OrderItem{*CreateTestOrderItem()},
        ShippingAddress:      `{"street": "Test Address", "city": "Test City", "state": "Test State", "postal_code": "12345", "country": "India"}`,
    }
}

func CreateTestOrderItem() *orderModels.OrderItem {
    return &orderModels.OrderItem{
        CatalogItemID:   TestCatalogItemID,
        CatalogItemType: "PRODUCT",
        CatalogItemName: "Test Product",
        CatalogItemSKU:  TestSKU,
        Quantity:        decimal.NewFromFloat(2),
        UnitPrice:       decimal.NewFromFloat(500.0),
        TotalPrice:      decimal.NewFromFloat(1000.0),
    }
}

func CreateTestCreateOrderRequest() *orderRequests.CreateOrderRequest {
    return &orderRequests.CreateOrderRequest{
        BuyerOrganizationID:  TestOrgID,
        SellerOrganizationID: TestSellerOrgID,
        Items: []orderRequests.CreateOrderItemRequest{
            {
                CatalogItemID:   TestCatalogItemID,
                CatalogItemType: "product",
                Quantity:        decimal.NewFromFloat(2),
                UnitPrice:       decimal.NewFromFloat(500.0),
            },
        },
        ShippingAddress: &orderRequests.Address{
            Street:     "Test Address",
            City:       "Test City",
            State:      "Test State",
            PostalCode: "12345",
            Country:    "India",
        },
        Notes: "Test order notes",
    }
}

func CreateTestUpdateOrderRequest() *orderRequests.UpdateOrderRequest {
    notes := "Updated notes"
    return &orderRequests.UpdateOrderRequest{
        ShippingAddress: &orderRequests.Address{
            Street:     "Updated Address",
            City:       "Updated City",
            State:      "Updated State",
            PostalCode: "54321",
            Country:    "India",
        },
        Notes: &notes,
    }
}

func CreateTestUpdateOrderStatusRequest() *orderRequests.UpdateOrderStatusRequest {
    return &orderRequests.UpdateOrderStatusRequest{
        Status: orderModels.OrderStatusConfirmed,
        Reason: "Status updated",
    }
}

func CreateTestListOrdersRequest() *orderRequests.ListOrdersRequest {
    status := orderModels.OrderStatusPending
    buyerOrgID := TestOrgID
    sellerOrgID := TestSellerOrgID
    return &orderRequests.ListOrdersRequest{
        Page:                 1,
        PageSize:             20,
        Status:               &status,
        BuyerOrganizationID:  &buyerOrgID,
        SellerOrganizationID: &sellerOrgID,
    }
}

// Catalog fixtures
func CreateTestCatalogItem(itemType catalogModels.CatalogItemType) *catalogModels.CatalogItem {
    item := &catalogModels.CatalogItem{
        OrganizationID: TestOrgID,
        ItemType:       itemType,
        Name:           "Test " + string(itemType),
        Description:    "Test description for " + string(itemType),
        BasePrice:      decimal.NewFromFloat(100.0),
        Currency:       "INR",
        Category:       "test-category",
        Subcategory:    "test-subcategory",
        SKU:            TestSKU + "-" + string(itemType),
        IsActive:       true,
        Visibility:     catalogModels.VisibilityOrg,
        Tags:           []string{"test", "fixture", string(itemType)},
    }
    item.ID = TestCatalogItemID + "-" + string(itemType)

    return item
}

func CreateTestProduct() *catalogModels.Product {
    baseItem := CreateTestCatalogItem(catalogModels.CatalogItemTypeProduct)
    weight := decimal.NewFromFloat(5.5)
    shelfLife := 30

    product := &catalogModels.Product{
        CatalogItem:   *baseItem,
        Weight:        &weight,
        Perishable:    true,
        ShelfLifeDays: &shelfLife,
        Dimensions:    `{"length": 10.0, "width": 5.0, "height": 3.0}`,
    }

    return product
}

func CreateTestService() *catalogModels.Service {
    baseItem := CreateTestCatalogItem(catalogModels.CatalogItemTypeService)
    duration := 120

    service := &catalogModels.Service{
        CatalogItem:     *baseItem,
        DurationMinutes: &duration,
        ServiceArea:     `{"radius": 50.0, "unit": "km"}`,
    }

    return service
}

func CreateTestLabour() *catalogModels.Labour {
    baseItem := CreateTestCatalogItem(catalogModels.CatalogItemTypeLabour)
    hourlyRate := decimal.NewFromFloat(50.0)

    labour := &catalogModels.Labour{
        CatalogItem: *baseItem,
        SkillLevel:  "intermediate",
        HourlyRate:  &hourlyRate,
    }

    return labour
}

func CreateTestProductRequest() *catalogRequests.CreateProductRequest {
    weight := decimal.NewFromFloat(5.5)
    perishable := true
    shelfLife := 30
    baseReq := CreateTestCatalogItemRequest()

    return &catalogRequests.CreateProductRequest{
        CreateCatalogItemRequest: *baseReq,
        Weight:                   &weight,
        Perishable:               &perishable,
        ShelfLifeDays:            &shelfLife,
        Dimensions: map[string]interface{}{
            "length": 10.0,
            "width":  5.0,
            "height": 3.0,
        },
    }
}

func CreateTestServiceRequest() *catalogRequests.CreateServiceRequest {
    duration := 120
    baseReq := CreateTestCatalogItemRequest()

    return &catalogRequests.CreateServiceRequest{
        CreateCatalogItemRequest: *baseReq,
        DurationMinutes:          &duration,
        ServiceArea: map[string]interface{}{
            "radius": 50.0,
            "unit":   "km",
        },
    }
}

func CreateTestLabourRequest() *catalogRequests.CreateLabourRequest {
    hourlyRate := decimal.NewFromFloat(75.0)
    baseReq := CreateTestCatalogItemRequest()

    return &catalogRequests.CreateLabourRequest{
        CreateCatalogItemRequest: *baseReq,
        SkillLevel:               "intermediate",
        HourlyRate:               &hourlyRate,
    }
}

func CreateTestCatalogItemRequest() *catalogRequests.CreateCatalogItemRequest {
    return &catalogRequests.CreateCatalogItemRequest{
        ItemType:      catalogModels.CatalogItemTypeProduct,
        Name:          "Test Product",
        Description:   "Test product description",
        BasePrice:     decimal.NewFromFloat(100.0),
        Currency:      "INR",
        UnitOfMeasure: "piece",
        Category:      "test-category",
        Subcategory:   "test-subcategory",
        SKU:           TestSKU,
        Tags:          []string{"test", "fixture"},
        Visibility:    catalogModels.VisibilityOrg,
    }
}

// Time helpers
func TimePtr(t time.Time) *time.Time {
    return &t
}

func StringPtr(s string) *string {
    return &s
}

func IntPtr(i int) *int {
    return &i
}

func BoolPtr(b bool) *bool {
    return &b
}
