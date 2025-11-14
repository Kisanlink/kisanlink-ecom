package middleware

import (
	"strconv"

	repositoryCommon "github.com/Kisanlink/kisanlink-ecom/internal/repositories/common"

	"github.com/Kisanlink/kisanlink-ecom/internal/auth"

	"github.com/gin-gonic/gin"
)

const (
	// RoleAdmin is the admin role
	RoleAdmin = "admin"
	// RoleSuperAdmin is the super admin role
	RoleSuperAdmin = "super_admin"
)

// ExtractQueryOptions extracts QueryOptions from HTTP query parameters
// and adds them to the context if the user has appropriate permissions
func ExtractQueryOptions(c *gin.Context) {
	// Parse include_deleted query parameter
	includeDeletedStr := c.DefaultQuery("include_deleted", "false")
	includeDeleted, err := strconv.ParseBool(includeDeletedStr)
	if err != nil || !includeDeleted {
		// Default behavior: deleted items are filtered
		return
	}

	// User wants to include deleted items - check permissions
	if !hasAdminPermission(c) {
		// Non-admin users cannot view deleted items
		// Silently ignore the parameter for security
		return
	}

	// Create query options with IncludeDeleted=true
	reason := c.DefaultQuery("reason", "admin query")
	opts := repositoryCommon.WithIncludeDeleted(reason)

	// Add query options to request context
	ctx := repositoryCommon.ContextWithQueryOptions(c.Request.Context(), opts)
	c.Request = c.Request.WithContext(ctx)
}

// hasAdminPermission checks if the current user has admin permissions
func hasAdminPermission(c *gin.Context) bool {
	// Get user context from gin context
	userContext, exists := c.Get("user_context")
	if !exists {
		return false
	}

	userCtx, ok := userContext.(*auth.UserContext)
	if !ok {
		return false
	}

	// Check if user has admin or super_admin role
	for _, role := range userCtx.Roles {
		if role == RoleAdmin || role == RoleSuperAdmin {
			return true
		}
	}

	return false
}
