package common

// APIResponse represents the standard API response structure
type APIResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    Error   *APIError   `json:"error,omitempty"`
}

// APIError represents error details in API responses
type APIError struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details string                 `json:"details,omitempty"`
    Fields  map[string]string      `json:"fields,omitempty"`
    Context map[string]interface{} `json:"context,omitempty"`
}

// Pagination represents pagination metadata
type Pagination struct {
    Page       int   `json:"page"`
    PageSize   int   `json:"page_size"`
    TotalItems int64 `json:"total_items"`
    TotalPages int   `json:"total_pages"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
    APIResponse
    Pagination *Pagination `json:"pagination,omitempty"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
    Status  string `json:"status"`
    Message string `json:"message"`
    Version string `json:"version"`
}
