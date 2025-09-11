package catalog

import (
    "encoding/json"
    "time"

    catalogModels "kisanlink-ecom/entities/models/catalog"

    "github.com/shopspring/decimal"
)

// CatalogItemResponse represents a catalog item in API responses
type CatalogItemResponse struct {
    ID             string                 `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
    OrganizationID string                 `json:"organization_id" example:"123e4567-e89b-12d3-a456-426614174001"`
    ItemType       string                 `json:"item_type" example:"PRODUCT"`
    Category       string                 `json:"category" example:"vegetables"`
    Subcategory    string                 `json:"subcategory" example:"tomatoes"`
    Name           string                 `json:"name" example:"Organic Tomatoes"`
    Description    string                 `json:"description" example:"Fresh organic tomatoes grown without pesticides"`
    SKU            string                 `json:"sku" example:"TOM-ORG-001"`
    UnitOfMeasure  string                 `json:"unit_of_measure" example:"kg"`
    BasePrice      decimal.Decimal        `json:"base_price" example:"25.50"`
    Currency       string                 `json:"currency" example:"INR"`
    IsActive       bool                   `json:"is_active" example:"true"`
    Visibility     string                 `json:"visibility" example:"ORG"`
    Tags           []string               `json:"tags" example:"organic,fresh,local"`
    Attributes     map[string]interface{} `json:"attributes,omitempty"`
    Images         []string               `json:"images,omitempty" example:"https://example.com/tomato1.jpg,https://example.com/tomato2.jpg"`
    CreatedAt      time.Time              `json:"created_at" example:"2024-01-15T10:30:00Z"`
    UpdatedAt      time.Time              `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

// ProductResponse represents a product in API responses
type ProductResponse struct {
    CatalogItemResponse
    Weight        *decimal.Decimal       `json:"weight,omitempty" example:"2.5"`
    Dimensions    map[string]interface{} `json:"dimensions,omitempty"`
    Perishable    bool                   `json:"perishable" example:"true"`
    ShelfLifeDays *int                   `json:"shelf_life_days,omitempty" example:"7"`
}

// ServiceResponse represents a service in API responses
type ServiceResponse struct {
    CatalogItemResponse
    DurationMinutes *int                   `json:"duration_minutes,omitempty" example:"120"`
    ServiceArea     map[string]interface{} `json:"service_area,omitempty"`
}

// LabourResponse represents labour in API responses
type LabourResponse struct {
    CatalogItemResponse
    SkillLevel string           `json:"skill_level" example:"experienced"`
    HourlyRate *decimal.Decimal `json:"hourly_rate,omitempty" example:"75.00"`
}

// CatalogSummaryResponse represents a catalog item summary in API responses
type CatalogSummaryResponse struct {
    ID             string          `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
    OrganizationID string          `json:"organization_id" example:"123e4567-e89b-12d3-a456-426614174001"`
    ItemType       string          `json:"item_type" example:"PRODUCT"`
    Name           string          `json:"name" example:"Organic Tomatoes"`
    BasePrice      decimal.Decimal `json:"base_price" example:"25.50"`
    Currency       string          `json:"currency" example:"INR"`
    IsActive       bool            `json:"is_active" example:"true"`
    Visibility     string          `json:"visibility" example:"ORG"`
    CreatedAt      time.Time       `json:"created_at" example:"2024-01-15T10:30:00Z"`
}

// BulkOperationResponse represents the result of a bulk operation
type BulkOperationResponse struct {
    TotalItems    int      `json:"total_items" example:"10"`
    SuccessCount  int      `json:"success_count" example:"8"`
    FailureCount  int      `json:"failure_count" example:"2"`
    SuccessItems  []string `json:"success_items,omitempty" example:"123e4567-e89b-12d3-a456-426614174000,123e4567-e89b-12d3-a456-426614174001"`
    FailureItems  []string `json:"failure_items,omitempty" example:"123e4567-e89b-12d3-a456-426614174002,123e4567-e89b-12d3-a456-426614174003"`
    ErrorMessages []string `json:"error_messages,omitempty" example:"Item not found,Permission denied"`
}

// Transformation functions

// ToCatalogItemResponse converts a CatalogItem model to CatalogItemResponse
func ToCatalogItemResponse(item *catalogModels.CatalogItem) (*CatalogItemResponse, error) {
    if item == nil {
        return nil, nil
    }

    response := &CatalogItemResponse{
        ID:             item.ID,
        OrganizationID: item.OrganizationID,
        ItemType:       string(item.ItemType),
        Category:       item.Category,
        Subcategory:    item.Subcategory,
        Name:           item.Name,
        Description:    item.Description,
        SKU:            item.SKU,
        UnitOfMeasure:  item.UnitOfMeasure,
        BasePrice:      item.BasePrice,
        Currency:       item.Currency,
        IsActive:       item.IsActive,
        Visibility:     string(item.Visibility),
        Tags:           []string(item.Tags),
        Images:         []string(item.Images),
        CreatedAt:      item.CreatedAt,
        UpdatedAt:      item.UpdatedAt,
    }

    // Parse attributes from JSON string
    if item.Attributes != "" {
        var attributes map[string]interface{}
        if err := json.Unmarshal([]byte(item.Attributes), &attributes); err == nil {
            response.Attributes = attributes
        }
    }

    return response, nil
}

// ToProductResponse converts a Product model to ProductResponse
func ToProductResponse(product *catalogModels.Product) (*ProductResponse, error) {
    if product == nil {
        return nil, nil
    }

    catalogResponse, err := ToCatalogItemResponse(&product.CatalogItem)
    if err != nil {
        return nil, err
    }

    response := &ProductResponse{
        CatalogItemResponse: *catalogResponse,
        Weight:              product.Weight,
        Perishable:          product.Perishable,
        ShelfLifeDays:       product.ShelfLifeDays,
    }

    // Parse dimensions from JSON string
    if product.Dimensions != "" {
        var dimensions map[string]interface{}
        if err := json.Unmarshal([]byte(product.Dimensions), &dimensions); err == nil {
            response.Dimensions = dimensions
        }
    }

    return response, nil
}

// ToServiceResponse converts a Service model to ServiceResponse
func ToServiceResponse(service *catalogModels.Service) (*ServiceResponse, error) {
    if service == nil {
        return nil, nil
    }

    catalogResponse, err := ToCatalogItemResponse(&service.CatalogItem)
    if err != nil {
        return nil, err
    }

    response := &ServiceResponse{
        CatalogItemResponse: *catalogResponse,
        DurationMinutes:     service.DurationMinutes,
    }

    // Parse service area from JSON string
    if service.ServiceArea != "" {
        var serviceArea map[string]interface{}
        if err := json.Unmarshal([]byte(service.ServiceArea), &serviceArea); err == nil {
            response.ServiceArea = serviceArea
        }
    }

    return response, nil
}

// ToLabourResponse converts a Labour model to LabourResponse
func ToLabourResponse(labour *catalogModels.Labour) (*LabourResponse, error) {
    if labour == nil {
        return nil, nil
    }

    catalogResponse, err := ToCatalogItemResponse(&labour.CatalogItem)
    if err != nil {
        return nil, err
    }

    response := &LabourResponse{
        CatalogItemResponse: *catalogResponse,
        SkillLevel:          labour.SkillLevel,
        HourlyRate:          labour.HourlyRate,
    }

    return response, nil
}

// ToCatalogSummaryResponse converts a CatalogItem model to CatalogSummaryResponse
func ToCatalogSummaryResponse(item *catalogModels.CatalogItem) *CatalogSummaryResponse {
    if item == nil {
        return nil
    }

    return &CatalogSummaryResponse{
        ID:             item.ID,
        OrganizationID: item.OrganizationID,
        ItemType:       string(item.ItemType),
        Name:           item.Name,
        BasePrice:      item.BasePrice,
        Currency:       item.Currency,
        IsActive:       item.IsActive,
        Visibility:     string(item.Visibility),
        CreatedAt:      item.CreatedAt,
    }
}

// ToCatalogItemResponseList converts a slice of CatalogItem models to CatalogItemResponse slice
func ToCatalogItemResponseList(items []*catalogModels.CatalogItem) ([]*CatalogItemResponse, error) {
    if len(items) == 0 {
        return []*CatalogItemResponse{}, nil
    }

    responses := make([]*CatalogItemResponse, len(items))
    for i, item := range items {
        response, err := ToCatalogItemResponse(item)
        if err != nil {
            return nil, err
        }
        responses[i] = response
    }

    return responses, nil
}

// ToProductResponseList converts a slice of Product models to ProductResponse slice
func ToProductResponseList(products []*catalogModels.Product) ([]*ProductResponse, error) {
    if len(products) == 0 {
        return []*ProductResponse{}, nil
    }

    responses := make([]*ProductResponse, len(products))
    for i, product := range products {
        response, err := ToProductResponse(product)
        if err != nil {
            return nil, err
        }
        responses[i] = response
    }

    return responses, nil
}

// ToServiceResponseList converts a slice of Service models to ServiceResponse slice
func ToServiceResponseList(services []*catalogModels.Service) ([]*ServiceResponse, error) {
    if len(services) == 0 {
        return []*ServiceResponse{}, nil
    }

    responses := make([]*ServiceResponse, len(services))
    for i, service := range services {
        response, err := ToServiceResponse(service)
        if err != nil {
            return nil, err
        }
        responses[i] = response
    }

    return responses, nil
}

// ToLabourResponseList converts a slice of Labour models to LabourResponse slice
func ToLabourResponseList(labours []*catalogModels.Labour) ([]*LabourResponse, error) {
    if len(labours) == 0 {
        return []*LabourResponse{}, nil
    }

    responses := make([]*LabourResponse, len(labours))
    for i, labour := range labours {
        response, err := ToLabourResponse(labour)
        if err != nil {
            return nil, err
        }
        responses[i] = response
    }

    return responses, nil
}

// ToCatalogSummaryResponseList converts a slice of CatalogItem models to CatalogSummaryResponse slice
func ToCatalogSummaryResponseList(items []*catalogModels.CatalogItem) []*CatalogSummaryResponse {
    if len(items) == 0 {
        return []*CatalogSummaryResponse{}
    }

    responses := make([]*CatalogSummaryResponse, len(items))
    for i, item := range items {
        responses[i] = ToCatalogSummaryResponse(item)
    }

    return responses
}

// NewBulkOperationResponse creates a new bulk operation response
func NewBulkOperationResponse(totalItems, successCount, failureCount int) *BulkOperationResponse {
    return &BulkOperationResponse{
        TotalItems:   totalItems,
        SuccessCount: successCount,
        FailureCount: failureCount,
    }
}

// AddSuccessItem adds a successful item ID to the bulk operation response
func (b *BulkOperationResponse) AddSuccessItem(itemID string) {
    b.SuccessItems = append(b.SuccessItems, itemID)
}

// AddFailureItem adds a failed item ID and error message to the bulk operation response
func (b *BulkOperationResponse) AddFailureItem(itemID, errorMessage string) {
    b.FailureItems = append(b.FailureItems, itemID)
    b.ErrorMessages = append(b.ErrorMessages, errorMessage)
}
