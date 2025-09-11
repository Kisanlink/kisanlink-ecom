package roles

import (
    "time"
)

// OrganizationRole represents the relationship between organizations and roles in the e-commerce service
// This allows organizations to have specific role configurations
type OrganizationRole struct {
    ID             string `json:"id" gorm:"primaryKey;type:varchar(255)"`
    OrganizationID string `json:"organization_id" gorm:"type:varchar(255);not null;index"` // Reference to organization
    AAARoleID      string `json:"role_id" gorm:"type:varchar(255);not null;index"`         // Reference to AAA service role ID
    IsActive       bool   `json:"is_active" gorm:"default:true"`

    // Organization-specific role configuration
    CustomPermissions string `json:"custom_permissions" gorm:"type:jsonb"` // Override default permissions
    MaxUsers          *int   `json:"max_users"`                            // Maximum users that can have this role
    IsDefault         bool   `json:"is_default" gorm:"default:false"`      // Is this a default role for new users

    // Configuration context
    ConfiguredBy string    `json:"configured_by" gorm:"type:varchar(255);index"` // Who configured this role
    ConfiguredAt time.Time `json:"configured_at" gorm:"autoCreateTime"`
    Notes        string    `json:"notes" gorm:"type:text"` // Configuration notes

    // Timestamps
    CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (OrganizationRole) TableName() string {
    return "organization_roles"
}

// NewOrganizationRole creates a new organization-role configuration
func NewOrganizationRole(organizationID, aaaRoleID, configuredBy string) *OrganizationRole {
    return &OrganizationRole{
        OrganizationID: organizationID,
        AAARoleID:      aaaRoleID,
        ConfiguredBy:   configuredBy,
        IsActive:       true,
    }
}

// SetCustomPermissions sets custom permissions for this organization role
func (or *OrganizationRole) SetCustomPermissions(permissions map[string]bool) {
    // Convert to JSON string for storage
    // This would typically use a JSON marshaller
    or.CustomPermissions = "{}" // Placeholder - implement proper JSON handling
}

// SetMaxUsers sets the maximum number of users that can have this role
func (or *OrganizationRole) SetMaxUsers(maxUsers int) {
    or.MaxUsers = &maxUsers
}

// SetAsDefault marks this as a default role for new users in the organization
func (or *OrganizationRole) SetAsDefault() {
    or.IsDefault = true
}
