package data

import (
	"time"

	"kisanlink-ecom/entities/models/roles"
)

// CreateTestUserRole creates a test user role assignment
func CreateTestUserRole() *roles.UserRole {
	return &roles.UserRole{
		Id:         "user-role-test-id-123",
		UserId:     "user-test-id-123",
		RoleId:     "role-test-id-123",
		IsActive:   true,
		AssignedBy: "admin-user-id",
		AssignedAt: time.Now().Add(-24 * time.Hour),
		CreatedAt:  time.Now().Add(-24 * time.Hour),
		UpdatedAt:  time.Now(),
	}
}

// CreateTestUserRoleWithIDs creates a test user role with specific user and role IDs
func CreateTestUserRoleWithIDs(userID, roleID string) *roles.UserRole {
	ur := CreateTestUserRole()
	ur.UserId = userID
	ur.RoleId = roleID
	return ur
}

// CreateTestUserRoleWithExpiration creates a test user role with expiration
func CreateTestUserRoleWithExpiration(expiresIn time.Duration) *roles.UserRole {
	ur := CreateTestUserRole()
	expiresAt := time.Now().Add(expiresIn)
	ur.ExpiresAt = &expiresAt
	return ur
}

// CreateTestInactiveUserRole creates an inactive user role
func CreateTestInactiveUserRole() *roles.UserRole {
	ur := CreateTestUserRole()
	ur.IsActive = false
	return ur
}

// CreateTestOrganizationRole creates a test organization role
func CreateTestOrganizationRole() *roles.OrganizationRole {
	return &roles.OrganizationRole{
		ID:             "org-role-test-id-123",
		OrganizationID: "org-test-id-123",
		AAARoleID:      "aaa-role-test-id-123",
		IsActive:       true,
		ConfiguredBy:   "admin-user-id",
		ConfiguredAt:   time.Now().Add(-24 * time.Hour),
		CreatedAt:      time.Now().Add(-24 * time.Hour),
		UpdatedAt:      time.Now(),
	}
}

// CreateTestOrganizationRoleWithOrgID creates a test organization role with specific org ID
func CreateTestOrganizationRoleWithOrgID(orgID string) *roles.OrganizationRole {
	or := CreateTestOrganizationRole()
	or.OrganizationID = orgID
	return or
}

// CreateTestDefaultOrganizationRole creates a default organization role
func CreateTestDefaultOrganizationRole() *roles.OrganizationRole {
	or := CreateTestOrganizationRole()
	or.IsDefault = true
	return or
}

// CreateTestOrganizationRoleWithMaxUsers creates an organization role with max users limit
func CreateTestOrganizationRoleWithMaxUsers(maxUsers int) *roles.OrganizationRole {
	or := CreateTestOrganizationRole()
	or.MaxUsers = &maxUsers
	return or
}

// CreateTestEcommerceRole creates a test e-commerce role
func CreateTestEcommerceRole() *roles.EcommerceRole {
	return &roles.EcommerceRole{
		AAARoleID:          "ecom-role-test-id-123",
		RoleName:           "Test Role",
		Description:        "Test role description",
		CanManageCatalog:   true,
		CanManageOrders:    true,
		CanManageInventory: false,
		CanManagePricing:   false,
		CanManageCustomers: false,
		CanViewAnalytics:   true,
		CanManageUsers:     false,
		CanManageSettings:  false,
		IsActive:           true,
		CreatedAt:          time.Now().Add(-24 * time.Hour),
		UpdatedAt:          time.Now(),
	}
}

// CreateTestEcommerceRoleWithPermissions creates an e-commerce role with specific permissions
func CreateTestEcommerceRoleWithPermissions(permissions map[string]bool) *roles.EcommerceRole {
	er := CreateTestEcommerceRole()

	if val, ok := permissions["catalog"]; ok {
		er.CanManageCatalog = val
	}
	if val, ok := permissions["orders"]; ok {
		er.CanManageOrders = val
	}
	if val, ok := permissions["inventory"]; ok {
		er.CanManageInventory = val
	}
	if val, ok := permissions["pricing"]; ok {
		er.CanManagePricing = val
	}
	if val, ok := permissions["customers"]; ok {
		er.CanManageCustomers = val
	}
	if val, ok := permissions["analytics"]; ok {
		er.CanViewAnalytics = val
	}
	if val, ok := permissions["users"]; ok {
		er.CanManageUsers = val
	}
	if val, ok := permissions["settings"]; ok {
		er.CanManageSettings = val
	}

	return er
}

// CreateTestAdminEcommerceRole creates an admin e-commerce role with all permissions
func CreateTestAdminEcommerceRole() *roles.EcommerceRole {
	return CreateTestEcommerceRoleWithPermissions(map[string]bool{
		"catalog":   true,
		"orders":    true,
		"inventory": true,
		"pricing":   true,
		"customers": true,
		"analytics": true,
		"users":     true,
		"settings":  true,
	})
}

// CreateTestEcommerceRoleWithOrgScope creates an org-scoped e-commerce role
func CreateTestEcommerceRoleWithOrgScope(orgID string) *roles.EcommerceRole {
	er := CreateTestEcommerceRole()
	er.OrganizationID = &orgID
	return er
}

// CreateTestUserRolesArray creates a slice of test user roles
func CreateTestUserRolesArray(count int, userID string) []*roles.UserRole {
	userRoles := make([]*roles.UserRole, count)
	for i := 0; i < count; i++ {
		userRoles[i] = CreateTestUserRoleWithIDs(userID, generateRoleID(i))
	}
	return userRoles
}

// Helper functions
func generateRoleID(index int) string {
	return "role-test-id-" + string(rune('A'+index))
}
