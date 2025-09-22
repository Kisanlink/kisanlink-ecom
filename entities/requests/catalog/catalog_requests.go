package catalog

import (
	"fmt"
	"kisanlink-ecom/entities/models/catalog"
	"time"

	"github.com/shopspring/decimal"
)

// CreateCatalogItemRequest represents the request to create a catalog item
type CreateCatalogItemRequest struct {
	ItemType      catalog.CatalogItemType `json:"item_type" binding:"required,oneof=PRODUCT SERVICE LABOUR CONTRACT" validate:"required,oneof=PRODUCT SERVICE LABOUR CONTRACT" example:"PRODUCT"`
	Category      string                  `json:"category" binding:"omitempty,max=100" validate:"omitempty,max=100" example:"vegetables"`
	Subcategory   string                  `json:"subcategory" binding:"omitempty,max=100" validate:"omitempty,max=100" example:"tomatoes"`
	Name          string                  `json:"name" binding:"required,max=255" validate:"required,min=1,max=255" example:"Organic Tomatoes"`
	Description   string                  `json:"description" binding:"omitempty" validate:"omitempty,max=2000" example:"Fresh organic tomatoes grown without pesticides"`
	SKU           string                  `json:"sku" binding:"omitempty,max=100" validate:"omitempty,max=100" example:"TOM-ORG-001"`
	UnitOfMeasure string                  `json:"unit_of_measure" binding:"omitempty,max=50" validate:"omitempty,max=50" example:"kg"`
	BasePrice     decimal.Decimal         `json:"base_price" binding:"required" validate:"required" example:"25.50"`
	Currency      string                  `json:"currency" binding:"omitempty,len=3" validate:"omitempty,len=3" example:"INR"`
	Visibility    catalog.VisibilityType  `json:"visibility" binding:"omitempty,oneof=PRIVATE ORG NETWORK PUBLIC" validate:"omitempty,oneof=PRIVATE ORG NETWORK PUBLIC" example:"ORG"`
	Tags          []string                `json:"tags" binding:"omitempty" validate:"omitempty,dive,max=50" example:"organic,fresh,local"`
	Attributes    map[string]interface{}  `json:"attributes" binding:"omitempty" validate:"omitempty"`
	Images        []string                `json:"images" binding:"omitempty" validate:"omitempty,dive,url" example:"https://example.com/tomato1.jpg,https://example.com/tomato2.jpg"`
}

// UpdateCatalogItemRequest represents the request to update a catalog item
type UpdateCatalogItemRequest struct {
	Category      *string                 `json:"category" binding:"omitempty,max=100" validate:"omitempty,max=100" example:"vegetables"`
	Subcategory   *string                 `json:"subcategory" binding:"omitempty,max=100" validate:"omitempty,max=100" example:"tomatoes"`
	Name          *string                 `json:"name" binding:"omitempty,max=255" validate:"omitempty,min=1,max=255" example:"Premium Organic Tomatoes"`
	Description   *string                 `json:"description" validate:"omitempty,max=2000" example:"Premium fresh organic tomatoes"`
	SKU           *string                 `json:"sku" binding:"omitempty,max=100" validate:"omitempty,max=100" example:"TOM-ORG-002"`
	UnitOfMeasure *string                 `json:"unit_of_measure" binding:"omitempty,max=50" validate:"omitempty,max=50" example:"kg"`
	BasePrice     *decimal.Decimal        `json:"base_price" binding:"omitempty" validate:"omitempty" example:"30.00"`
	Currency      *string                 `json:"currency" binding:"omitempty,len=3" validate:"omitempty,len=3" example:"INR"`
	IsActive      *bool                   `json:"is_active" example:"true"`
	Visibility    *catalog.VisibilityType `json:"visibility" binding:"omitempty,oneof=PRIVATE ORG NETWORK PUBLIC" validate:"omitempty,oneof=PRIVATE ORG NETWORK PUBLIC" example:"NETWORK"`
	Tags          []string                `json:"tags" binding:"omitempty" validate:"omitempty,dive,max=50" example:"organic,premium,fresh"`
	Attributes    map[string]interface{}  `json:"attributes" binding:"omitempty" validate:"omitempty"`
	Images        []string                `json:"images" binding:"omitempty" validate:"omitempty,dive,url" example:"https://example.com/premium-tomato.jpg"`
}

// CatalogFilter represents filters for catalog queries
type CatalogFilter struct {
	ItemType       *catalog.CatalogItemType `form:"item_type" binding:"omitempty,oneof=PRODUCT SERVICE LABOUR CONTRACT" validate:"omitempty,oneof=PRODUCT SERVICE LABOUR CONTRACT" example:"PRODUCT"`
	Category       *string                  `form:"category" validate:"omitempty,max=100" example:"vegetables"`
	Subcategory    *string                  `form:"subcategory" validate:"omitempty,max=100" example:"tomatoes"`
	IsActive       *bool                    `form:"is_active" example:"true"`
	Visibility     *catalog.VisibilityType  `form:"visibility" binding:"omitempty,oneof=PRIVATE ORG NETWORK PUBLIC" validate:"omitempty,oneof=PRIVATE ORG NETWORK PUBLIC" example:"ORG"`
	MinPrice       *decimal.Decimal         `form:"min_price" binding:"omitempty" validate:"omitempty,gte=0" example:"10.00"`
	MaxPrice       *decimal.Decimal         `form:"max_price" binding:"omitempty" validate:"omitempty,gte=0" example:"100.00"`
	Search         *string                  `form:"search" validate:"omitempty,max=255" example:"organic tomatoes"`
	Tags           []string                 `form:"tags" validate:"omitempty,dive,max=50" example:"organic,fresh"`
	OrganizationID *string                  `form:"organization_id" validate:"omitempty,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// CreateProductRequest represents the request to create a product
type CreateProductRequest struct {
	CreateCatalogItemRequest
	Weight        *decimal.Decimal       `json:"weight" binding:"omitempty" validate:"omitempty,gte=0" example:"2.5"`
	Dimensions    map[string]interface{} `json:"dimensions" binding:"omitempty" validate:"omitempty"`
	Perishable    *bool                  `json:"perishable" example:"true"`
	ShelfLifeDays *int                   `json:"shelf_life_days" binding:"omitempty,gte=0" validate:"omitempty,gte=0" example:"7"`
}

// CreateServiceRequest represents the request to create a service
type CreateServiceRequest struct {
	CreateCatalogItemRequest
	DurationMinutes *int                   `json:"duration_minutes" binding:"omitempty,gte=0" validate:"omitempty,gte=0" example:"120"`
	ServiceArea     map[string]interface{} `json:"service_area" binding:"omitempty" validate:"omitempty"`
}

// CreateLabourRequest represents the request to create labour
type CreateLabourRequest struct {
	CreateCatalogItemRequest
	SkillLevel string           `json:"skill_level" binding:"omitempty,max=50" validate:"omitempty,max=50" example:"experienced"`
	HourlyRate *decimal.Decimal `json:"hourly_rate" binding:"omitempty" validate:"omitempty,gte=0" example:"75.00"`
}

// CreateContractRequest represents the request to create a contract
type CreateContractRequest struct {
	CreateCatalogItemRequest
	Term      string `json:"term" binding:"required,max=100" validate:"required,max=100" example:"fixed"`
	Duration  int    `json:"duration" binding:"required,gt=0" validate:"required,gt=0" example:"12"`
	StartDate string `json:"start_date" binding:"omitempty" validate:"omitempty" example:"2024-01-01"`
	EndDate   string `json:"end_date" binding:"omitempty" validate:"omitempty" example:"2024-12-31"`
	Terms     string `json:"terms" binding:"omitempty" validate:"omitempty,max=5000" example:"Contract terms and conditions"`
}

// ListCatalogItemsRequest represents the request to list catalog items
type ListCatalogItemsRequest struct {
	// Pagination
	Page     int `form:"page" binding:"omitempty,min=1" validate:"omitempty,min=1" example:"1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100" validate:"omitempty,min=1,max=100" example:"20"`

	// Filtering
	CatalogFilter

	// Sorting
	SortBy    *string `form:"sort_by" binding:"omitempty,oneof=created_at updated_at name base_price" validate:"omitempty,oneof=created_at updated_at name base_price" example:"created_at"`
	SortOrder *string `form:"sort_order" binding:"omitempty,oneof=asc desc" validate:"omitempty,oneof=asc desc" example:"desc"`
}

// UpdateProductRequest represents the request to update a product
type UpdateProductRequest struct {
	UpdateCatalogItemRequest
	Weight        *decimal.Decimal       `json:"weight" binding:"omitempty" validate:"omitempty,gte=0" example:"3.0"`
	Dimensions    map[string]interface{} `json:"dimensions" binding:"omitempty" validate:"omitempty"`
	Perishable    *bool                  `json:"perishable" example:"false"`
	ShelfLifeDays *int                   `json:"shelf_life_days" binding:"omitempty,gte=0" validate:"omitempty,gte=0" example:"14"`
}

// UpdateServiceRequest represents the request to update a service
type UpdateServiceRequest struct {
	UpdateCatalogItemRequest
	DurationMinutes *int                   `json:"duration_minutes" binding:"omitempty,gte=0" validate:"omitempty,gte=0" example:"180"`
	ServiceArea     map[string]interface{} `json:"service_area" binding:"omitempty" validate:"omitempty"`
}

// UpdateLabourRequest represents the request to update labour
type UpdateLabourRequest struct {
	UpdateCatalogItemRequest
	SkillLevel *string          `json:"skill_level" binding:"omitempty,max=50" validate:"omitempty,max=50" example:"expert"`
	HourlyRate *decimal.Decimal `json:"hourly_rate" binding:"omitempty" validate:"omitempty,gte=0" example:"100.00"`
}

// UpdateContractRequest represents the request to update a contract
type UpdateContractRequest struct {
	UpdateCatalogItemRequest
	Term      *string `json:"term" binding:"omitempty,max=100" validate:"omitempty,max=100" example:"renewable"`
	Duration  *int    `json:"duration" binding:"omitempty,gt=0" validate:"omitempty,gt=0" example:"24"`
	StartDate *string `json:"start_date" binding:"omitempty" validate:"omitempty" example:"2024-02-01"`
	EndDate   *string `json:"end_date" binding:"omitempty" validate:"omitempty" example:"2025-01-31"`
	Terms     *string `json:"terms" binding:"omitempty" validate:"omitempty,max=5000" example:"Updated contract terms"`
}

// SearchCatalogRequest represents the request to search catalog items
type SearchCatalogRequest struct {
	Query     string                   `form:"q" binding:"required" validate:"required,min=1,max=255" example:"organic tomatoes"`
	ItemType  *catalog.CatalogItemType `form:"item_type" binding:"omitempty,oneof=PRODUCT SERVICE LABOUR CONTRACT" validate:"omitempty,oneof=PRODUCT SERVICE LABOUR CONTRACT" example:"PRODUCT"`
	Category  *string                  `form:"category" validate:"omitempty,max=100" example:"vegetables"`
	MinPrice  *decimal.Decimal         `form:"min_price" binding:"omitempty" validate:"omitempty,gte=0" example:"10.00"`
	MaxPrice  *decimal.Decimal         `form:"max_price" binding:"omitempty" validate:"omitempty,gte=0" example:"100.00"`
	Tags      []string                 `form:"tags" validate:"omitempty,dive,max=50" example:"organic,fresh"`
	Page      int                      `form:"page" binding:"omitempty,min=1" validate:"omitempty,min=1" example:"1"`
	PageSize  int                      `form:"page_size" binding:"omitempty,min=1,max=100" validate:"omitempty,min=1,max=100" example:"20"`
	SortBy    *string                  `form:"sort_by" binding:"omitempty,oneof=relevance created_at updated_at name base_price" validate:"omitempty,oneof=relevance created_at updated_at name base_price" example:"relevance"`
	SortOrder *string                  `form:"sort_order" binding:"omitempty,oneof=asc desc" validate:"omitempty,oneof=asc desc" example:"desc"`
}

// BulkUpdateCatalogRequest represents the request to bulk update catalog items
type BulkUpdateCatalogRequest struct {
	ItemIDs    []string                `json:"item_ids" binding:"required,min=1,dive,uuid4" validate:"required,min=1,dive,uuid4" example:"123e4567-e89b-12d3-a456-426614174000,123e4567-e89b-12d3-a456-426614174001"`
	IsActive   *bool                   `json:"is_active" example:"false"`
	Visibility *catalog.VisibilityType `json:"visibility" binding:"omitempty,oneof=PRIVATE ORG NETWORK PUBLIC" validate:"omitempty,oneof=PRIVATE ORG NETWORK PUBLIC" example:"PRIVATE"`
	Tags       []string                `json:"tags" binding:"omitempty" validate:"omitempty,dive,max=50" example:"discontinued,clearance"`
}

// Helper methods

// ToCatalogFilter converts ListCatalogItemsRequest to CatalogFilter
func (r *ListCatalogItemsRequest) ToCatalogFilter() *CatalogFilter {
	return &r.CatalogFilter
}

// ToSearchFilters converts SearchCatalogRequest to filter parameters
func (r *SearchCatalogRequest) ToSearchFilters() map[string]interface{} {
	filters := make(map[string]interface{})

	filters["query"] = r.Query

	if r.ItemType != nil {
		filters["item_type"] = *r.ItemType
	}
	if r.Category != nil {
		filters["category"] = *r.Category
	}
	if r.MinPrice != nil {
		filters["min_price"] = *r.MinPrice
	}
	if r.MaxPrice != nil {
		filters["max_price"] = *r.MaxPrice
	}
	if len(r.Tags) > 0 {
		filters["tags"] = r.Tags
	}

	// Pagination
	page := r.Page
	if page <= 0 {
		page = 1
	}
	pageSize := r.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	filters["page"] = page
	filters["page_size"] = pageSize

	// Sorting
	sortBy := "relevance"
	if r.SortBy != nil {
		sortBy = *r.SortBy
	}
	sortOrder := "desc"
	if r.SortOrder != nil {
		sortOrder = *r.SortOrder
	}
	filters["sort_by"] = sortBy
	filters["sort_order"] = sortOrder

	return filters
}

// Validation helper methods

// ValidateCreateProductRequest validates a create product request
func ValidateCreateProductRequest(req *CreateProductRequest) error {
	if req.ItemType != catalog.CatalogItemTypeProduct {
		return fmt.Errorf("item_type must be PRODUCT for product creation")
	}

	if req.Perishable != nil && *req.Perishable && req.ShelfLifeDays == nil {
		return fmt.Errorf("shelf_life_days is required for perishable products")
	}

	if req.Weight != nil && req.Weight.LessThan(decimal.Zero) {
		return fmt.Errorf("weight must be greater than or equal to 0")
	}

	return nil
}

// ValidateCreateServiceRequest validates a create service request
func ValidateCreateServiceRequest(req *CreateServiceRequest) error {
	if req.ItemType != catalog.CatalogItemTypeService {
		return fmt.Errorf("item_type must be SERVICE for service creation")
	}

	if req.DurationMinutes != nil && *req.DurationMinutes <= 0 {
		return fmt.Errorf("duration_minutes must be greater than 0")
	}

	return nil
}

// ValidateCreateLabourRequest validates a create labour request
func ValidateCreateLabourRequest(req *CreateLabourRequest) error {
	if req.ItemType != catalog.CatalogItemTypeLabour {
		return fmt.Errorf("item_type must be LABOUR for labour creation")
	}

	if req.HourlyRate != nil && req.HourlyRate.LessThan(decimal.Zero) {
		return fmt.Errorf("hourly_rate must be greater than or equal to 0")
	}

	return nil
}

// ValidateCreateContractRequest validates a create contract request
func ValidateCreateContractRequest(req *CreateContractRequest) error {
	if req.ItemType != catalog.CatalogItemTypeContract {
		return fmt.Errorf("item_type must be CONTRACT for contract creation")
	}

	if req.Duration <= 0 {
		return fmt.Errorf("duration must be greater than 0")
	}

	// Validate date formats if provided
	if req.StartDate != "" {
		if _, err := time.Parse("2006-01-02", req.StartDate); err != nil {
			return fmt.Errorf("start_date must be in YYYY-MM-DD format")
		}
	}

	if req.EndDate != "" {
		if _, err := time.Parse("2006-01-02", req.EndDate); err != nil {
			return fmt.Errorf("end_date must be in YYYY-MM-DD format")
		}
	}

	// Validate date consistency
	if req.StartDate != "" && req.EndDate != "" {
		startDate, _ := time.Parse("2006-01-02", req.StartDate)
		endDate, _ := time.Parse("2006-01-02", req.EndDate)
		if endDate.Before(startDate) {
			return fmt.Errorf("end_date must be after start_date")
		}
	}

	return nil
}
