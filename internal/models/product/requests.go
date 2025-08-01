package product

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// CreateProductRequest represents request to create a product.
// @Description Request structure for creating a new product
type CreateProductRequest struct {
	base.BaseRequest
	Name        string  `json:"name" validate:"required,min=1,max=200" example:"Organic Tomatoes"`
	Description string  `json:"description,omitempty" validate:"max=1000" example:"Fresh organic tomatoes from local farms"`
	Price       float64 `json:"price" validate:"required,min=0" example:"2.99"`
	Currency    string  `json:"currency" validate:"required,len=3" example:"USD"`
	Category    string  `json:"category" validate:"required" example:"Vegetables"`
	Stock       int     `json:"stock" validate:"min=0" example:"100"`
}

// UpdateProductRequest represents request to update a product.
// @Description Request structure for updating an existing product
type UpdateProductRequest struct {
	base.BaseRequest
	Name        string  `json:"name,omitempty" validate:"omitempty,min=1,max=200" example:"Organic Tomatoes"`
	Description string  `json:"description,omitempty" validate:"max=1000" example:"Fresh organic tomatoes from local farms"`
	Price       float64 `json:"price,omitempty" validate:"omitempty,min=0" example:"2.99"`
	Currency    string  `json:"currency,omitempty" validate:"omitempty,len=3" example:"USD"`
	Category    string  `json:"category,omitempty" example:"Vegetables"`
	Stock       int     `json:"stock,omitempty" validate:"min=0" example:"100"`
	Status      Status  `json:"status,omitempty" example:"active"`
}

// SimpleProduct represents a simplified product structure for API responses.
// @Description Simplified product structure for API responses
type SimpleProduct struct {
	ID          string  `json:"id" example:"PRODUCT123456789"`
	Name        string  `json:"name" example:"Organic Tomatoes"`
	Description string  `json:"description" example:"Fresh organic tomatoes from local farms"`
	Price       float64 `json:"price" example:"2.99"`
	Currency    string  `json:"currency" example:"USD"`
	Category    string  `json:"category" example:"Vegetables"`
	Stock       int     `json:"stock" example:"100"`
	Status      Status  `json:"status" example:"active"`
}
