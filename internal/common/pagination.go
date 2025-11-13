package common

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	// DefaultPage is the default page number
	DefaultPage = 1
	// DefaultLimit is the default number of items per page
	DefaultLimit = 20
	// MaxLimit is the maximum number of items per page
	MaxLimit = 100
)

// PaginationParams represents pagination parameters from the request
type PaginationParams struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// PaginationRequest represents pagination parameters for repository operations
type PaginationRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// SortParams represents sorting parameters from the request
type SortParams struct {
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"` // "asc" or "desc"
}

// FilterParams represents common filter parameters
type FilterParams struct {
	Search   string            `json:"search"`
	Category string            `json:"category"`
	OrgID    string            `json:"org_id"`
	IsActive *bool             `json:"is_active"`
	DateFrom string            `json:"date_from"`
	DateTo   string            `json:"date_to"`
	Extra    map[string]string `json:"extra"`
}

// GetPaginationParams extracts pagination parameters from the gin context
func GetPaginationParams(c *gin.Context) *PaginationParams {
	page := DefaultPage
	limit := DefaultLimit

	// Parse page parameter
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// Parse limit parameter
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			if l > MaxLimit {
				limit = MaxLimit
			} else {
				limit = l
			}
		}
	}

	return &PaginationParams{
		Page:  page,
		Limit: limit,
	}
}

// GetSortParams extracts sorting parameters from the gin context
func GetSortParams(c *gin.Context) *SortParams {
	sortBy := c.Query("sort_by")
	sortOrder := c.Query("sort_order")

	// Validate sort order
	if sortOrder != "" && sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc" // default to ascending
	}

	return &SortParams{
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}
}

// GetFilterParams extracts filter parameters from the gin context
func GetFilterParams(c *gin.Context) *FilterParams {
	search := c.Query("search")
	category := c.Query("category")
	orgID := c.Query("org_id")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	// Parse is_active boolean
	var isActive *bool
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if isActiveStr == "true" {
			active := true
			isActive = &active
		} else if isActiveStr == "false" {
			active := false
			isActive = &active
		}
	}

	// Extract extra filters
	extra := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if key != "page" && key != "limit" && key != "sort_by" && key != "sort_order" &&
			key != "search" && key != "category" && key != "org_id" && key != "is_active" &&
			key != "date_from" && key != "date_to" {
			if len(values) > 0 {
				extra[key] = values[0]
			}
		}
	}

	return &FilterParams{
		Search:   search,
		Category: category,
		OrgID:    orgID,
		IsActive: isActive,
		DateFrom: dateFrom,
		DateTo:   dateTo,
		Extra:    extra,
	}
}

// CalculateOffset calculates the database offset for pagination
func (p *PaginationParams) CalculateOffset() int {
	return (p.Page - 1) * p.Limit
}

// Validate validates pagination parameters
func (p *PaginationParams) Validate() error {
	if p.Page < 1 {
		return ErrInvalidInput
	}
	if p.Limit < 1 || p.Limit > MaxLimit {
		return ErrInvalidInput
	}
	return nil
}

// BuildPaginationResponse builds a complete pagination response
func BuildPaginationResponse(data interface{}, total int, params *PaginationParams) *Response {
	paginationMeta := NewPaginationMeta(params.Page, params.Limit, total)

	return &Response{
		Data: data,
		Meta: &ResponseMeta{
			Pagination: paginationMeta,
		},
	}
}
