package product

import (
	"fmt"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Product represents a product in the system.
// @Description Product model representing a product in the catalog
type Product struct {
	*base.BaseModel
	Name        string  `json:"name" validate:"required,min=1,max=200" gorm:"type:varchar(255);not null" example:"Organic Tomatoes"`
	Description string  `json:"description" validate:"max=1000" gorm:"type:text" example:"Fresh organic tomatoes from local farms"`
	Price       float64 `json:"price" validate:"required,min=0" gorm:"type:decimal(10,2);not null" example:"2.99"`
	Currency    string  `json:"currency" validate:"required,len=3" gorm:"type:varchar(3);default:'USD'" example:"USD"`
	Category    string  `json:"category" validate:"required" gorm:"type:varchar(100);not null" example:"Vegetables"`
	Stock       int     `json:"stock" validate:"min=0" gorm:"default:0" example:"100"`
	Status      Status  `json:"status" validate:"required" gorm:"type:varchar(50);default:'active'" example:"active"`
}

// NewProduct creates a new Product with initialized fields.
func NewProduct(name, description, category string, price float64, currency string, stock int) *Product {
	baseModel := base.NewBaseModel("PRODUCT", hash.Small)

	return &Product{
		BaseModel:   baseModel,
		Name:        name,
		Description: description,
		Price:       price,
		Currency:    currency,
		Category:    category,
		Stock:       stock,
		Status:      StatusActive,
	}
}

// BeforeCreate implements the base model interface.
func (p *Product) BeforeCreate() error {
	if p.Name == "" {
		return fmt.Errorf("product name cannot be empty")
	}
	if p.Price < 0 {
		return fmt.Errorf("product price cannot be negative")
	}
	if p.Stock < 0 {
		return fmt.Errorf("product stock cannot be negative")
	}

	return p.BaseModel.BeforeCreate()
}

// BeforeUpdate implements the base model interface.
func (p *Product) BeforeUpdate() error {
	return p.BaseModel.BeforeUpdate()
}

// BeforeDelete implements the base model interface.
func (p *Product) BeforeDelete() error {
	return p.BaseModel.BeforeDelete()
}

// BeforeSoftDelete implements the base model interface.
func (p *Product) BeforeSoftDelete() error {
	return p.BaseModel.BeforeSoftDelete()
}

// Status represents general status for entities.

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusDeleted  Status = "deleted"
)
