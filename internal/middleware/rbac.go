package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"kisanlink-ecom/entities/models/common"
	"kisanlink-ecom/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

// RBACConfig holds configuration for RBAC middleware
type RBACConfig struct {
	// CacheEnabled enables permission caching
	CacheEnabled bool
	// CacheTTL sets the cache TTL for permissions
	CacheTTL time.Duration
	// SkipAuthPaths defines paths that skip authorization
	SkipAuthPaths []string
	// DefaultAction is the default action to check if none specified
	DefaultAction string
}

// DefaultRBACConfig returns default RBAC configuration
func DefaultRBACConfig() *RBACConfig {
	return &RBACConfig{
		CacheEnabled:  true,
		CacheTTL:      5 * time.Minute,
		SkipAuthPaths: []string{"/health", "/docs", "/metrics"},
		DefaultAction: "read",
	}
}

// RBACMiddleware provides role-based access control
type RBACMiddleware struct {
	aaaClient auth.Client
	cache     *cache.Cache
	config    *RBACConfig
}

// NewRBACMiddleware creates a new RBAC middleware
func NewRBACMiddleware(aaaClient auth.Client, config *RBACConfig) *RBACMiddleware {
	if config == nil {
		config = DefaultRBACConfig()
	}

	var permCache *cache.Cache
	if config.CacheEnabled {
		permCache = cache.New(config.CacheTTL, config.CacheTTL*2)
	}

	return &RBACMiddleware{
		aaaClient: aaaClient,
		cache:     permCache,
		config:    config,
	}
}

// RequirePermission creates middleware that requires specific permission
func (m *RBACMiddleware) RequirePermission(resource string, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authorization for configured paths
		if m.shouldSkipAuth(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Get user context from authentication middleware
		userContext, exists := c.Get("user_context")
		if !exists {
			m.handleUnauthorized(c, "User context not found")
			return
		}

		userCtx, ok := userContext.(*auth.UserContext)
		if !ok {
			m.handleUnauthorized(c, "Invalid user context")
			return
		}

		// Check permission
		allowed, err := m.checkPermission(c.Request.Context(), userCtx, resource, action)
		if err != nil {
			m.handleError(c, "Permission check failed", err)
			return
		}

		if !allowed {
			m.handleForbidden(c, fmt.Sprintf("Permission denied for action '%s' on resource '%s'", action, resource))
			return
		}

		// Store permission info in context for potential use in handlers
		c.Set("permission_resource", resource)
		c.Set("permission_action", action)

		c.Next()
	}
}

// RequireAnyPermission requires user to have any of the specified permissions
func (m *RBACMiddleware) RequireAnyPermission(permissions []PermissionPair) gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.shouldSkipAuth(c.Request.URL.Path) {
			c.Next()
			return
		}

		userContext, exists := c.Get("user_context")
		if !exists {
			m.handleUnauthorized(c, "User context not found")
			return
		}

		userCtx, ok := userContext.(*auth.UserContext)
		if !ok {
			m.handleUnauthorized(c, "Invalid user context")
			return
		}

		// Check if user has any of the required permissions
		for _, perm := range permissions {
			allowed, err := m.checkPermission(c.Request.Context(), userCtx, perm.Resource, perm.Action)
			if err != nil {
				m.handleError(c, "Permission check failed", err)
				return
			}

			if allowed {
				c.Set("permission_resource", perm.Resource)
				c.Set("permission_action", perm.Action)
				c.Next()
				return
			}
		}

		m.handleForbidden(c, "Permission denied for all required permissions")
	}
}

// RequireRole creates middleware that requires specific role
func (m *RBACMiddleware) RequireRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.shouldSkipAuth(c.Request.URL.Path) {
			c.Next()
			return
		}

		userContext, exists := c.Get("user_context")
		if !exists {
			m.handleUnauthorized(c, "User context not found")
			return
		}

		userCtx, ok := userContext.(*auth.UserContext)
		if !ok {
			m.handleUnauthorized(c, "Invalid user context")
			return
		}

		// Check if user has any of the required roles
		userRoles := make(map[string]bool)
		for _, role := range userCtx.Roles {
			userRoles[role] = true
		}

		for _, requiredRole := range requiredRoles {
			if userRoles[requiredRole] {
				c.Set("user_role", requiredRole)
				c.Next()
				return
			}
		}

		m.handleForbidden(c, fmt.Sprintf("Role denied. Required roles: %v", requiredRoles))
	}
}

// RequireOrganization ensures user belongs to the specified organization
func (m *RBACMiddleware) RequireOrganization(orgID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.shouldSkipAuth(c.Request.URL.Path) {
			c.Next()
			return
		}

		userContext, exists := c.Get("user_context")
		if !exists {
			m.handleUnauthorized(c, "User context not found")
			return
		}

		userCtx, ok := userContext.(*auth.UserContext)
		if !ok {
			m.handleUnauthorized(c, "Invalid user context")
			return
		}

		// Allow if user is in the same organization
		if userCtx.OrganizationID == orgID {
			c.Next()
			return
		}

		// Check with AAA service for cross-organization permissions
		authReq := &auth.AuthorizeRequest{
			UserID:     userCtx.UserID,
			TenantID:   userCtx.TenantID,
			Resource:   "organization",
			Action:     "access",
			ResourceID: orgID,
		}
		authResp, err := m.aaaClient.Authorize(c.Request.Context(), authReq)
		if err != nil {
			m.handleError(c, "Organization validation failed", err)
			return
		}

		if !authResp.Allowed {
			m.handleForbidden(c, fmt.Sprintf("Access denied to organization %s", orgID))
			return
		}

		c.Next()
	}
}

// RequireOwnership ensures user owns the resource or has admin privileges
func (m *RBACMiddleware) RequireOwnership(resourceOwnerIDFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.shouldSkipAuth(c.Request.URL.Path) {
			c.Next()
			return
		}

		userContext, exists := c.Get("user_context")
		if !exists {
			m.handleUnauthorized(c, "User context not found")
			return
		}

		userCtx, ok := userContext.(*auth.UserContext)
		if !ok {
			m.handleUnauthorized(c, "Invalid user context")
			return
		}

		// Get resource owner ID
		resourceOwnerID := resourceOwnerIDFunc(c)
		if resourceOwnerID == "" {
			m.handleError(c, "Resource owner ID not found", nil)
			return
		}

		// Allow if user owns the resource
		if userCtx.UserID == resourceOwnerID {
			c.Next()
			return
		}

		// Check if user has admin role
		for _, role := range userCtx.Roles {
			if role == "admin" || role == "super_admin" {
				c.Next()
				return
			}
		}

		m.handleForbidden(c, "Resource ownership required")
	}
}

// PermissionPair represents a resource-action pair
type PermissionPair struct {
	Resource string
	Action   string
}

// Helper methods

func (m *RBACMiddleware) shouldSkipAuth(path string) bool {
	for _, skipPath := range m.config.SkipAuthPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

func (m *RBACMiddleware) checkPermission(ctx context.Context, userCtx *auth.UserContext, resource, action string) (bool, error) {
	// Check cache first if enabled
	if m.cache != nil {
		cacheKey := fmt.Sprintf("perm:%s:%s:%s", userCtx.UserID, resource, action)
		if cached, found := m.cache.Get(cacheKey); found {
			return cached.(bool), nil
		}
	}

	// Check with AAA service
	authReq := &auth.AuthorizeRequest{
		UserID:   userCtx.UserID,
		TenantID: userCtx.TenantID,
		Resource: resource,
		Action:   action,
	}
	authResp, err := m.aaaClient.Authorize(ctx, authReq)
	if err != nil {
		return false, err
	}
	allowed := authResp.Allowed

	// Cache result if caching is enabled
	if m.cache != nil {
		cacheKey := fmt.Sprintf("perm:%s:%s:%s", userCtx.UserID, resource, action)
		m.cache.Set(cacheKey, allowed, cache.DefaultExpiration)
	}

	return allowed, nil
}

func (m *RBACMiddleware) extractRequestContext(userCtx *auth.UserContext) map[string]interface{} {
	return map[string]interface{}{
		"user_id":         userCtx.UserID,
		"organization_id": userCtx.OrganizationID,
		"roles":           userCtx.Roles,
		"timestamp":       time.Now().Unix(),
	}
}

func (m *RBACMiddleware) handleUnauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(401, common.APIResponse{
		Success: false,
		Error: &common.APIError{
			Code:    "UNAUTHORIZED",
			Message: message,
		},
	})
}

func (m *RBACMiddleware) handleForbidden(c *gin.Context, message string) {
	c.AbortWithStatusJSON(403, common.APIResponse{
		Success: false,
		Error: &common.APIError{
			Code:    "FORBIDDEN",
			Message: message,
		},
	})
}

func (m *RBACMiddleware) handleError(c *gin.Context, message string, err error) {
	details := ""
	if err != nil {
		details = err.Error()
	}

	c.AbortWithStatusJSON(500, common.APIResponse{
		Success: false,
		Error: &common.APIError{
			Code:    "AUTHORIZATION_ERROR",
			Message: message,
			Details: details,
		},
	})
}

// InvalidatePermissionCache invalidates permission cache for a user
func (m *RBACMiddleware) InvalidatePermissionCache(userID string) {
	if m.cache == nil {
		return
	}

	// Get all cache items and remove those for this user
	items := m.cache.Items()
	for key := range items {
		if strings.HasPrefix(key, fmt.Sprintf("perm:%s:", userID)) {
			m.cache.Delete(key)
		}
	}
}

// ClearPermissionCache clears the entire permission cache
func (m *RBACMiddleware) ClearPermissionCache() {
	if m.cache != nil {
		m.cache.Flush()
	}
}
