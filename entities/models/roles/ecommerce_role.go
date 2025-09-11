package roles

import (
    "time"
)

// EcommerceRole represents a role implementation in the e-commerce service
// This references AAA service roles and adds e-commerce specific functionality
type EcommerceRole struct {
    AAARoleID   string `json:"id" gorm:"primaryKey;type:varchar(255)"`            // Primary key - same as AAA service role ID
    RoleName    string `json:"role_name" gorm:"type:varchar(100);not null;index"` // Cached from AAA service
    Description string `json:"description" gorm:"type:text"`                      // Cached from AAA service

    // E-commerce specific role fields
    CanManageCatalog   bool `json:"can_manage_catalog" gorm:"default:false"`   // Can create/edit/delete products
    CanManageOrders    bool `json:"can_manage_orders" gorm:"default:false"`    // Can view/edit orders
    CanManageInventory bool `json:"can_manage_inventory" gorm:"default:false"` // Can manage stock levels
    CanManagePricing   bool `json:"can_manage_pricing" gorm:"default:false"`   // Can set prices and discounts
    CanManageCustomers bool `json:"can_manage_customers" gorm:"default:false"` // Can view customer data
    CanViewAnalytics   bool `json:"can_view_analytics" gorm:"default:false"`   // Can view business analytics
    CanManageUsers     bool `json:"can_manage_users" gorm:"default:false"`     // Can manage user access
    CanManageSettings  bool `json:"can_manage_settings" gorm:"default:false"`  // Can change system settings

    // Role scope in e-commerce context
    OrganizationID *string `json:"organization_id" gorm:"type:varchar(255);index"` // If role is org-scoped
    IsActive       bool    `json:"is_active" gorm:"default:true"`

    // Timestamps
    CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (EcommerceRole) TableName() string {
    return "ecommerce_roles"
}

// NewEcommerceRole creates a new e-commerce role from AAA service role
func NewEcommerceRole(aaaRoleID, roleName, description string) *EcommerceRole {
    return &EcommerceRole{
        AAARoleID:   aaaRoleID,
        RoleName:    roleName,
        Description: description,
        IsActive:    true,
    }
}

// SetOrganizationScope sets the organization scope for this role
func (r *EcommerceRole) SetOrganizationScope(orgID string) {
    r.OrganizationID = &orgID
}

// HasPermission checks if this role has a specific permission
func (r *EcommerceRole) HasPermission(permission string) bool {
    switch permission {
    case "catalog.manage":
        return r.CanManageCatalog
    case "orders.manage":
        return r.CanManageOrders
    case "inventory.manage":
        return r.CanManageInventory
    case "pricing.manage":
        return r.CanManagePricing
    case "customers.manage":
        return r.CanManageCustomers
    case "analytics.view":
        return r.CanViewAnalytics
    case "users.manage":
        return r.CanManageUsers
    case "settings.manage":
        return r.CanManageSettings
    default:
        return false
    }
}

// GetID returns the role ID (same as AAARoleID)
func (r *EcommerceRole) GetID() string {
    return r.AAARoleID
}
