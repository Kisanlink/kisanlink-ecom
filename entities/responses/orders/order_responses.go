package orders

import (
    "encoding/json"
    "time"

    orderModels "kisanlink-ecom/entities/models/orders"

    "github.com/shopspring/decimal"
)

// OrderResponse represents an order in API responses
type OrderResponse struct {
    ID                    string                       `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
    OrderNumber           string                       `json:"order_number" example:"ORD-2024-001"`
    Status                string                       `json:"status" example:"pending"`
    BuyerOrganizationID   string                       `json:"buyer_organization_id" example:"123e4567-e89b-12d3-a456-426614174001"`
    SellerOrganizationID  string                       `json:"seller_organization_id" example:"123e4567-e89b-12d3-a456-426614174002"`
    BuyerUserID           string                       `json:"buyer_user_id" example:"123e4567-e89b-12d3-a456-426614174003"`
    SubtotalAmount        decimal.Decimal              `json:"subtotal_amount" example:"255.00"`
    TaxAmount             decimal.Decimal              `json:"tax_amount" example:"25.50"`
    DiscountAmount        decimal.Decimal              `json:"discount_amount" example:"0.00"`
    ShippingAmount        decimal.Decimal              `json:"shipping_amount" example:"50.00"`
    TotalAmount           decimal.Decimal              `json:"total_amount" example:"330.50"`
    ShippingAddress       *AddressResponse             `json:"shipping_address,omitempty"`
    EstimatedDeliveryDate *time.Time                   `json:"estimated_delivery_date,omitempty" example:"2024-01-20T10:30:00Z"`
    ActualDeliveryDate    *time.Time                   `json:"actual_delivery_date,omitempty" example:"2024-01-19T14:15:00Z"`
    Notes                 string                       `json:"notes,omitempty" example:"Special delivery instructions"`
    Metadata              map[string]interface{}       `json:"metadata,omitempty"`
    Items                 []OrderItemResponse          `json:"items,omitempty"`
    StatusHistory         []OrderStatusHistoryResponse `json:"status_history,omitempty"`
    CreatedAt             time.Time                    `json:"created_at" example:"2024-01-15T10:30:00Z"`
    UpdatedAt             time.Time                    `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

// OrderItemResponse represents an order item in API responses
type OrderItemResponse struct {
    ID              string                 `json:"id" example:"123e4567-e89b-12d3-a456-426614174004"`
    OrderID         string                 `json:"order_id" example:"123e4567-e89b-12d3-a456-426614174000"`
    CatalogItemID   string                 `json:"catalog_item_id" example:"123e4567-e89b-12d3-a456-426614174005"`
    CatalogItemType string                 `json:"catalog_item_type" example:"product"`
    CatalogItemName string                 `json:"catalog_item_name" example:"Organic Tomatoes"`
    CatalogItemSKU  string                 `json:"catalog_item_sku" example:"TOM-ORG-001"`
    Quantity        decimal.Decimal        `json:"quantity" example:"10.5"`
    UnitPrice       decimal.Decimal        `json:"unit_price" example:"25.50"`
    TotalPrice      decimal.Decimal        `json:"total_price" example:"267.75"`
    TaxRate         decimal.Decimal        `json:"tax_rate" example:"0.1000"`
    TaxAmount       decimal.Decimal        `json:"tax_amount" example:"26.78"`
    DiscountRate    decimal.Decimal        `json:"discount_rate" example:"0.0000"`
    DiscountAmount  decimal.Decimal        `json:"discount_amount" example:"0.00"`
    Metadata        map[string]interface{} `json:"metadata,omitempty"`
    CreatedAt       time.Time              `json:"created_at" example:"2024-01-15T10:30:00Z"`
    UpdatedAt       time.Time              `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

// AddressResponse represents shipping/billing address in API responses
type AddressResponse struct {
    Street     string `json:"street" example:"123 Farm Road"`
    City       string `json:"city" example:"Rural City"`
    State      string `json:"state" example:"Maharashtra"`
    PostalCode string `json:"postal_code" example:"411001"`
    Country    string `json:"country" example:"India"`
}

// OrderStatusHistoryResponse represents order status history in API responses
type OrderStatusHistoryResponse struct {
    ID                      string                 `json:"id" example:"123e4567-e89b-12d3-a456-426614174006"`
    OrderID                 string                 `json:"order_id" example:"123e4567-e89b-12d3-a456-426614174000"`
    FromStatus              *string                `json:"from_status,omitempty" example:"pending"`
    ToStatus                string                 `json:"to_status" example:"confirmed"`
    Reason                  string                 `json:"reason" example:"Payment confirmed by bank"`
    ChangedByUserID         string                 `json:"changed_by_user_id" example:"123e4567-e89b-12d3-a456-426614174003"`
    ChangedByOrganizationID string                 `json:"changed_by_organization_id" example:"123e4567-e89b-12d3-a456-426614174001"`
    Metadata                map[string]interface{} `json:"metadata,omitempty"`
    CreatedAt               time.Time              `json:"created_at" example:"2024-01-15T10:35:00Z"`
}

// OrderSummaryResponse represents an order summary in API responses
type OrderSummaryResponse struct {
    ID                string          `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
    OrderNumber       string          `json:"order_number" example:"ORD-2024-001"`
    Status            string          `json:"status" example:"pending"`
    TotalAmount       decimal.Decimal `json:"total_amount" example:"330.50"`
    ItemCount         int             `json:"item_count" example:"3"`
    BuyerOrgID        string          `json:"buyer_org_id" example:"123e4567-e89b-12d3-a456-426614174001"`
    SellerOrgID       string          `json:"seller_org_id" example:"123e4567-e89b-12d3-a456-426614174002"`
    CreatedAt         time.Time       `json:"created_at" example:"2024-01-15T10:30:00Z"`
    EstimatedDelivery *time.Time      `json:"estimated_delivery,omitempty" example:"2024-01-20T10:30:00Z"`
}

// Transformation functions

// ToOrderResponse converts an Order model to OrderResponse
func ToOrderResponse(order *orderModels.Order) (*OrderResponse, error) {
    if order == nil {
        return nil, nil
    }

    response := &OrderResponse{
        ID:                    order.ID,
        OrderNumber:           order.OrderNumber,
        Status:                string(order.Status),
        BuyerOrganizationID:   order.BuyerOrganizationID,
        SellerOrganizationID:  order.SellerOrganizationID,
        BuyerUserID:           order.BuyerUserID,
        SubtotalAmount:        order.SubtotalAmount,
        TaxAmount:             order.TaxAmount,
        DiscountAmount:        order.DiscountAmount,
        ShippingAmount:        order.ShippingAmount,
        TotalAmount:           order.TotalAmount,
        EstimatedDeliveryDate: order.EstimatedDeliveryDate,
        ActualDeliveryDate:    order.ActualDeliveryDate,
        Notes:                 order.Notes,
        CreatedAt:             order.CreatedAt,
        UpdatedAt:             order.UpdatedAt,
    }

    // Parse shipping address from JSON string
    if order.ShippingAddress != "" {
        var address orderModels.Address
        if err := json.Unmarshal([]byte(order.ShippingAddress), &address); err == nil {
            response.ShippingAddress = &AddressResponse{
                Street:     address.Street,
                City:       address.City,
                State:      address.State,
                PostalCode: address.PostalCode,
                Country:    address.Country,
            }
        }
    }

    // Parse metadata from JSON string
    if order.Metadata != "" {
        var metadata map[string]interface{}
        if err := json.Unmarshal([]byte(order.Metadata), &metadata); err == nil {
            response.Metadata = metadata
        }
    }

    // Convert items
    if len(order.Items) > 0 {
        response.Items = make([]OrderItemResponse, len(order.Items))
        for i, item := range order.Items {
            itemResponse, err := ToOrderItemResponse(&item)
            if err != nil {
                return nil, err
            }
            response.Items[i] = *itemResponse
        }
    }

    // Convert status history
    if len(order.StatusHistory) > 0 {
        response.StatusHistory = make([]OrderStatusHistoryResponse, len(order.StatusHistory))
        for i, history := range order.StatusHistory {
            historyResponse, err := ToOrderStatusHistoryResponse(&history)
            if err != nil {
                return nil, err
            }
            response.StatusHistory[i] = *historyResponse
        }
    }

    return response, nil
}

// ToOrderItemResponse converts an OrderItem model to OrderItemResponse
func ToOrderItemResponse(item *orderModels.OrderItem) (*OrderItemResponse, error) {
    if item == nil {
        return nil, nil
    }

    response := &OrderItemResponse{
        ID:              item.ID,
        OrderID:         item.OrderID,
        CatalogItemID:   item.CatalogItemID,
        CatalogItemType: item.CatalogItemType,
        CatalogItemName: item.CatalogItemName,
        CatalogItemSKU:  item.CatalogItemSKU,
        Quantity:        item.Quantity,
        UnitPrice:       item.UnitPrice,
        TotalPrice:      item.TotalPrice,
        TaxRate:         item.TaxRate,
        TaxAmount:       item.TaxAmount,
        DiscountRate:    item.DiscountRate,
        DiscountAmount:  item.DiscountAmount,
        CreatedAt:       item.CreatedAt,
        UpdatedAt:       item.UpdatedAt,
    }

    // Parse metadata from JSON string
    if item.Metadata != "" {
        var metadata map[string]interface{}
        if err := json.Unmarshal([]byte(item.Metadata), &metadata); err == nil {
            response.Metadata = metadata
        }
    }

    return response, nil
}

// ToOrderStatusHistoryResponse converts an OrderStatusHistory model to OrderStatusHistoryResponse
func ToOrderStatusHistoryResponse(history *orderModels.OrderStatusHistory) (*OrderStatusHistoryResponse, error) {
    if history == nil {
        return nil, nil
    }

    response := &OrderStatusHistoryResponse{
        ID:                      history.ID,
        OrderID:                 history.OrderID,
        FromStatus:              history.FromStatus,
        ToStatus:                string(history.ToStatus),
        Reason:                  history.Reason,
        ChangedByUserID:         history.ChangedByUserID,
        ChangedByOrganizationID: history.ChangedByOrganizationID,
        CreatedAt:               history.CreatedAt,
    }

    // Parse metadata from JSON string
    if history.Metadata != "" {
        var metadata map[string]interface{}
        if err := json.Unmarshal([]byte(history.Metadata), &metadata); err == nil {
            response.Metadata = metadata
        }
    }

    return response, nil
}

// ToOrderSummaryResponse converts an Order model to OrderSummaryResponse
func ToOrderSummaryResponse(order *orderModels.Order) *OrderSummaryResponse {
    if order == nil {
        return nil
    }

    return &OrderSummaryResponse{
        ID:                order.ID,
        OrderNumber:       order.OrderNumber,
        Status:            string(order.Status),
        TotalAmount:       order.TotalAmount,
        ItemCount:         len(order.Items),
        BuyerOrgID:        order.BuyerOrganizationID,
        SellerOrgID:       order.SellerOrganizationID,
        CreatedAt:         order.CreatedAt,
        EstimatedDelivery: order.EstimatedDeliveryDate,
    }
}

// ToOrderResponseList converts a slice of Order models to OrderResponse slice
func ToOrderResponseList(orders []*orderModels.Order) ([]*OrderResponse, error) {
    if len(orders) == 0 {
        return []*OrderResponse{}, nil
    }

    responses := make([]*OrderResponse, len(orders))
    for i, order := range orders {
        response, err := ToOrderResponse(order)
        if err != nil {
            return nil, err
        }
        responses[i] = response
    }

    return responses, nil
}

// ToOrderSummaryResponseList converts a slice of Order models to OrderSummaryResponse slice
func ToOrderSummaryResponseList(orders []*orderModels.Order) []*OrderSummaryResponse {
    if len(orders) == 0 {
        return []*OrderSummaryResponse{}
    }

    responses := make([]*OrderSummaryResponse, len(orders))
    for i, order := range orders {
        responses[i] = ToOrderSummaryResponse(order)
    }

    return responses
}

// Validation error helpers

// ValidationError represents a field validation error
type ValidationError struct {
    Field   string      `json:"field"`
    Message string      `json:"message"`
    Value   interface{} `json:"value,omitempty"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
    Errors []ValidationError `json:"errors"`
}

// Error implements the error interface
func (ve ValidationErrors) Error() string {
    if len(ve.Errors) == 0 {
        return "validation failed"
    }
    return ve.Errors[0].Message
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string, value interface{}) ValidationError {
    return ValidationError{
        Field:   field,
        Message: message,
        Value:   value,
    }
}

// NewValidationErrors creates a new validation errors collection
func NewValidationErrors(errors ...ValidationError) ValidationErrors {
    return ValidationErrors{
        Errors: errors,
    }
}
