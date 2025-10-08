package catalog

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// Category represents a hierarchical category structure
type Category struct {
	base.BaseModel

	// Tenant isolation
	OrganizationID string `json:"organization_id" gorm:"type:varchar(255);not null;index:idx_org_slug,priority:1;index:idx_org_categories"`

	// Category details
	Name        string `json:"name" gorm:"type:varchar(255);not null;index:idx_categories_search_name"`
	Description string `json:"description" gorm:"type:text"`
	Slug        string `json:"slug" gorm:"type:varchar(255);not null;uniqueIndex:idx_org_slug,priority:2"`

	// Hierarchy
	ParentID *string `json:"parent_id" gorm:"type:varchar(255);index:idx_parent"`
	Level    int     `json:"level" gorm:"not null;default:0;check:level >= 0"`
	Path     string  `json:"path" gorm:"type:varchar(1000);not null;index:idx_path"` // materialized path like "/electronics/mobile"

	// Metadata
	IsActive  bool `json:"is_active" gorm:"not null;default:true;index:idx_categories_active"`
	SortOrder int  `json:"sort_order" gorm:"not null;default:0"`

	// Audit fields
	CreatedBy string `json:"created_by" gorm:"type:varchar(255);not null"`
	UpdatedBy string `json:"updated_by" gorm:"type:varchar(255);not null"`

	// Soft delete support
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index:idx_categories_deleted"`

	// Relationships
	Parent   *Category  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []Category `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

// TableName returns the table name for GORM
func (Category) TableName() string {
	return "categories"
}

// NewCategory creates a new category
func NewCategory(orgID, name, slug string) *Category {
	return &Category{
		BaseModel:      *base.NewBaseModel("CAT", "large"),
		OrganizationID: orgID,
		Name:           name,
		Slug:           slug,
		Level:          0,
		Path:           "/" + slug,
		IsActive:       true,
		SortOrder:      0,
	}
}

// IsRoot checks if the category is a root category
func (c *Category) IsRoot() bool {
	return c.ParentID == nil
}

// GetFullPath returns the full category path
func (c *Category) GetFullPath() string {
	return c.Path
}
