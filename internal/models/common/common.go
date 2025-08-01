package common

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// APIResponse represents a standard API response.
// @Description Standard API response structure
type APIResponse struct {
	Success   bool        `json:"success" example:"true"`
	Message   string      `json:"message" example:"Operation completed successfully"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// APIError represents an API error.
// @Description API error structure
type APIError struct {
	Code    string `json:"code" example:"VALIDATION_ERROR"`
	Message string `json:"message" example:"Invalid input data"`
	Details string `json:"details,omitempty" example:"Field 'email' is required"`
}

// PaginatedResponse represents a paginated API response.
// @Description Paginated API response structure
type PaginatedResponse struct {
	APIResponse
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination represents pagination metadata.
// @Description Pagination metadata
type Pagination struct {
	Page       int `json:"page" example:"1"`
	PerPage    int `json:"per_page" example:"20"`
	Total      int `json:"total" example:"100"`
	TotalPages int `json:"total_pages" example:"5"`
}

// HealthResponse represents health check response.
// @Description Health check response structure
type HealthResponse struct {
	Status  string `json:"status" example:"ok"`
	Message string `json:"message" example:"KisanLink E-commerce API is running"`
	Version string `json:"version" example:"1.0.0"`
}

// CreateRequest represents a base structure for creation requests.
type CreateRequest struct {
	base.BaseRequest
}

// UpdateRequest represents a base structure for update requests.
type UpdateRequest struct {
	base.BaseRequest
	ID string `json:"id" validate:"required"`
}

// DeleteRequest represents a base structure for delete requests.
type DeleteRequest struct {
	base.BaseRequest
	ID string `json:"id" validate:"required"`
}

// ListRequest represents a base structure for list requests with pagination.
type ListRequest struct {
	base.BaseRequest
	Page    int `json:"page" validate:"min=1"`
	PerPage int `json:"per_page" validate:"min=1,max=100"`
}

// GetByIDRequest represents a base structure for get by ID requests.
type GetByIDRequest struct {
	base.BaseRequest
	ID string `json:"id" validate:"required"`
}
