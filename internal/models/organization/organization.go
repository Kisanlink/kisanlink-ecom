package organization

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Organization represents an organization in the system
type Organization struct {
	*base.BaseModel
	Name        string  `json:"name" gorm:"type:varchar(100);not null;uniqueIndex"`
	Description string  `json:"description" gorm:"type:text"`
	ParentID    *string `json:"parent_id" gorm:"type:varchar(255)"` // For org hierarchy
	IsActive    bool    `json:"is_active" gorm:"default:true"`
	Type        string  `json:"type" gorm:"type:varchar(50);not null"` // FPO, collaborator, etc.
	Metadata    string  `json:"metadata" gorm:"type:jsonb"`            // Additional org metadata
}

// NewOrganization creates a new Organization instance
func NewOrganization(name, description, orgType string) *Organization {
	return &Organization{
		BaseModel:   base.NewBaseModel("ORG", hash.Medium),
		Name:        name,
		Description: description,
		Type:        orgType,
		IsActive:    true,
	}
}

// OrganizationType represents the type of organization
type OrganizationType string

const (
	OrganizationTypeFPO          OrganizationType = "FPO"
	OrganizationTypeCollaborator OrganizationType = "collaborator"
	OrganizationTypePlatform     OrganizationType = "platform"
)
