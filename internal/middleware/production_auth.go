package middleware

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
	"time"

	"kisanlink-ecom/entities/models/common"
	"kisanlink-ecom/internal/auth"
	"kisanlink-ecom/internal/cache"

	"github.com/gin-gonic/gin"
	gocache "github.com/patrickmn/go-cache"
)

// ProductionAuthConfig holds configuration for production authentication
type ProductionAuthConfig struct {
	// SkipAuthPaths defines paths that skip authentication
	SkipAuthPaths []string
	// TokenCacheTTL sets the cache TTL for JWT tokens
	TokenCacheTTL time.Duration
	// PermissionCacheTTL sets the cache TTL for permissions (short TTL as per requirement)
	PermissionCacheTTL time.Duration
	// RequireOrganization forces organization validation
	RequireOrganization bool
	// EnableMetrics enables authentication metrics
	EnableMetrics bool
}

// DefaultProductionAuthConfig returns default production auth configuration
func DefaultProductionAuthConfig() *ProductionAuthConfig {
	return &ProductionAuthConfig{
		SkipAuthPaths: []string{
			"/health", "/ready", "/metrics",
			"/docs", "/docs/",
			"/api/v1/auth/login", "/api/v1/auth/register",
		},
		TokenCacheTTL:       2 * time.Minute,
		PermissionCacheTTL:  30 * time.Second, // Short TTL for permissions as per requirement
		RequireOrganization: true,
		EnableMetrics:       true,
	}
}

// ProductionAuthMiddleware provides production-ready authentication and authorization
type ProductionAuthMiddleware struct {
	aaaClient       auth.Client
	tokenCache      *gocache.Cache
	permissionCache *gocache.Cache
	cacheManager    *cache.CacheManager
	config          *ProductionAuthConfig
}

// NewProductionAuthMiddleware creates a new production authentication middleware
func NewProductionAuthMiddleware(aaaClient auth.Client, config *ProductionAuthConfig) *ProductionAuthMiddleware {
	if config == nil {
		config = DefaultProductionAuthConfig()
	}

	// Initialize token cache
	tokenCache := gocache.New(config.TokenCacheTTL, config.TokenCacheTTL*2)

	// Initialize permission cache with short TTL
	permissionCache := gocache.New(config.PermissionCacheTTL, config.PermissionCacheTTL*2)

	// Initialize cache manager for advanced caching features
	cacheManager := cache.GetCacheManager()

	return &ProductionAuthMiddleware{
		aaaClient:       aaaClient,
		tokenCache:      tokenCache,
		permissionCache: permissionCache,
		cacheManager:    cacheManager,
		config:          config,
	}
}

// JWTAuthenticationMiddleware provides JWT token validation with local verification
func (m *ProductionAuthMiddleware) JWTAuthenticationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for configured paths
		if m.shouldSkipAuth(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Extract token from request
		token, err := m.extractToken(c)
		if err != nil {
			m.handleAuthError(c, "Token extraction failed", err)
			return
		}

		if token == "" {
			m.handleUnauthorized(c, "Authentication token required")
			return
		}

		// Validate token with caching
		userContext, err := m.validateTokenWithCache(c.Request.Context(), token)
		if err != nil {
			m.handleAuthError(c, "Token validation failed", err)
			return
		}

		if userContext == nil {
			m.handleUnauthorized(c, "Invalid token")
			return
		}

		// Validate organization if required
		if m.config.RequireOrganization && userContext.OrganizationID == "" {
			m.handleUnauthorized(c, "Organization membership required")
			return
		}

		// Check if user is active
		if !userContext.IsActive {
			m.handleUnauthorized(c, "User account is inactive")
			return
		}

		// Set user context in gin context
		m.setUserContext(c, userContext)

		c.Next()
	}
}

// AuthorizeMiddleware creates authorization middleware that checks permissions with AAA service
func (m *ProductionAuthMiddleware) AuthorizeMiddleware(resource, action string, resourceIDExtractor func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user context from authentication middleware
		userContext, exists := GetUserContext(c)
		if !exists {
			m.handleUnauthorized(c, "Authentication required for authorization")
			return
		}

		// Extract resource ID if extractor is provided
		resourceID := ""
		if resourceIDExtractor != nil {
			resourceID = resourceIDExtractor(c)
		}

		// Check permission with caching
		allowed, err := m.checkPermissionWithCache(c.Request.Context(), userContext, resource, action, resourceID)
		if err != nil {
			m.handleAuthError(c, "Permission check failed", err)
			return
		}

		if !allowed {
			m.handleForbidden(c, "Insufficient permissions", resource, action, resourceID)
			return
		}

		// Store authorization context for audit logging
		c.Set("auth_resource", resource)
		c.Set("auth_action", action)
		c.Set("auth_resource_id", resourceID)

		c.Next()
	}
}

// RequirePermission creates middleware that requires specific permission
func (m *ProductionAuthMiddleware) RequirePermission(resource, action string) gin.HandlerFunc {
	return m.AuthorizeMiddleware(resource, action, nil)
}

// RequireResourcePermission creates middleware that requires permission on a specific resource
func (m *ProductionAuthMiddleware) RequireResourcePermission(resource, action string, resourceIDExtractor func(*gin.Context) string) gin.HandlerFunc {
	return m.AuthorizeMiddleware(resource, action, resourceIDExtractor)
}

// OptionalAuth provides optional authentication (sets user context if token is valid, but doesn't require it)
func (m *ProductionAuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from request
		token, err := m.extractToken(c)
		if err != nil || token == "" {
			// No token or invalid token format - continue without authentication
			c.Next()
			return
		}

		// Validate token and get user context
		userContext, err := m.validateTokenWithCache(c.Request.Context(), token)
		if err != nil || userContext == nil {
			// Invalid token - continue without authentication
			c.Next()
			return
		}

		// Set user context if validation successful
		m.setUserContext(c, userContext)

		c.Next()
	}
}

// Helper methods

func (m *ProductionAuthMiddleware) shouldSkipAuth(path string) bool {
	for _, skipPath := range m.config.SkipAuthPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

func (m *ProductionAuthMiddleware) extractToken(c *gin.Context) (string, error) {
	// Try Authorization header first (Bearer token)
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return strings.TrimSpace(parts[1]), nil
		}
	}

	// Try query parameter as fallback
	if token := c.Query("token"); token != "" {
		return token, nil
	}

	// Try cookie as fallback
	if token, err := c.Cookie("auth_token"); err == nil && token != "" {
		return token, nil
	}

	return "", nil
}

func (m *ProductionAuthMiddleware) validateTokenWithCache(ctx context.Context, token string) (*auth.UserContext, error) {
	// Check cache first
	if cached, found := m.tokenCache.Get(token); found {
		if userCtx, ok := cached.(*auth.UserContext); ok {
			return userCtx, nil
		}
	}

	// Validate with AAA service (local JWT verification + AAA validation)
	claims, err := m.aaaClient.ValidateToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("AAA token validation failed: %w", err)
	}

	// Create user context from claims with all enhanced fields
	userContext := &auth.UserContext{
		UserID:           claims.UserID,
		Username:         claims.Username,
		Email:            claims.Email,
		PhoneNumber:      claims.PhoneNumber,
		CountryCode:      claims.CountryCode,
		TenantID:         claims.TenantID,
		OrganizationID:   claims.OrganizationID,
		OrganizationName: claims.OrganizationName,
		Roles:            claims.Roles,
		RoleIDs:          claims.RoleIDs,
		Permissions:      claims.Permissions,
		Scopes:           claims.Scopes,
		IsActive:         true,
		IsValidated:      claims.IsValidated,
		UserRoles:        claims.UserRoles,
		Organizations:    claims.Organizations,
		Groups:           claims.Groups,
		TokenType:        claims.TokenType,
		TokenVersion:     claims.TokenVersion,
		Subject:          claims.Subject,
		SessionID:        claims.SessionID,
		JTI:              claims.JTI,
		Issuer:           claims.Issuer,
		Audience:         claims.Audience,
		TenantContext:    claims.TenantContext,
		UserContextData:  claims.UserContextData,
	}

	// Cache result
	m.tokenCache.Set(token, userContext, gocache.DefaultExpiration)

	return userContext, nil
}

func (m *ProductionAuthMiddleware) checkPermissionWithCache(ctx context.Context, userContext *auth.UserContext, resource, action, resourceID string) (bool, error) {
	// Generate cache key for permission
	cacheKey := m.generatePermissionCacheKey(userContext.UserID, userContext.TenantID, resource, action, resourceID)

	// Check cache first
	if cached, found := m.permissionCache.Get(cacheKey); found {
		if allowed, ok := cached.(bool); ok {
			return allowed, nil
		}
	}

	// Call AAA service for authorization
	authReq := &auth.AuthorizeRequest{
		UserID:     userContext.UserID,
		TenantID:   userContext.TenantID,
		Resource:   resource,
		Action:     action,
		ResourceID: resourceID,
	}

	authResp, err := m.aaaClient.Authorize(ctx, authReq)
	if err != nil {
		return false, fmt.Errorf("AAA authorization failed: %w", err)
	}

	// Cache the result with short TTL
	m.permissionCache.Set(cacheKey, authResp.Allowed, gocache.DefaultExpiration)

	return authResp.Allowed, nil
}

func (m *ProductionAuthMiddleware) generatePermissionCacheKey(userID, tenantID, resource, action, resourceID string) string {
	// Create a deterministic cache key
	key := fmt.Sprintf("perm:%s:%s:%s:%s:%s", userID, tenantID, resource, action, resourceID)

	// Hash the key to ensure consistent length and avoid special characters
	hash := sha256.Sum256([]byte(key))
	return fmt.Sprintf("permission:%x", hash[:16]) // Use first 16 bytes of hash
}

func (m *ProductionAuthMiddleware) setUserContext(c *gin.Context, userContext *auth.UserContext) {
	c.Set("user_context", userContext)
	c.Set("user_id", userContext.UserID)
	c.Set("tenant_id", userContext.TenantID)
	c.Set("organization_id", userContext.OrganizationID)
	c.Set("user_roles", userContext.Roles)
	c.Set("user_permissions", userContext.Permissions)

	// Add user info to response headers for debugging (non-production)
	if gin.Mode() != gin.ReleaseMode {
		c.Header("X-User-ID", userContext.UserID)
		c.Header("X-Tenant-ID", userContext.TenantID)
		c.Header("X-Organization-ID", userContext.OrganizationID)
	}
}

// Error handling methods

func (m *ProductionAuthMiddleware) handleUnauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, common.APIResponse{
		Success: false,
		Error: &common.APIError{
			Code:    "UNAUTHORIZED",
			Message: message,
		},
	})
}

func (m *ProductionAuthMiddleware) handleAuthError(c *gin.Context, message string, err error) {
	details := ""
	if err != nil {
		details = err.Error()
	}

	c.AbortWithStatusJSON(http.StatusUnauthorized, common.APIResponse{
		Success: false,
		Error: &common.APIError{
			Code:    "AUTHENTICATION_ERROR",
			Message: message,
			Details: details,
		},
	})
}

func (m *ProductionAuthMiddleware) handleForbidden(c *gin.Context, message, resource, action, resourceID string) {
	c.AbortWithStatusJSON(http.StatusForbidden, common.APIResponse{
		Success: false,
		Error: &common.APIError{
			Code:    "FORBIDDEN",
			Message: message,
			Details: fmt.Sprintf("Access denied for action '%s' on resource '%s' (ID: %s)", action, resource, resourceID),
		},
	})
}

// Cache management methods

// InvalidateTokenCache invalidates cached token
func (m *ProductionAuthMiddleware) InvalidateTokenCache(token string) {
	if m.tokenCache != nil {
		m.tokenCache.Delete(token)
	}
}

// InvalidateUserPermissions invalidates all cached permissions for a specific user
func (m *ProductionAuthMiddleware) InvalidateUserPermissions(userID, tenantID string) {
	if m.permissionCache == nil {
		return
	}

	// Since go-cache doesn't support pattern deletion and TTL is short (30s),
	// we can clear the entire cache for simplicity
	m.permissionCache.Flush()
}

// ClearAllCaches clears both token and permission caches
func (m *ProductionAuthMiddleware) ClearAllCaches() {
	if m.tokenCache != nil {
		m.tokenCache.Flush()
	}
	if m.permissionCache != nil {
		m.permissionCache.Flush()
	}
}

// Common resource ID extractors for catalog operations

// ExtractIDFromParam extracts resource ID from URL parameter
func ExtractIDFromParam(paramName string) func(*gin.Context) string {
	return func(c *gin.Context) string {
		return c.Param(paramName)
	}
}

// ExtractCatalogIDFromParam extracts catalog ID from 'id' parameter
func ExtractCatalogIDFromParam() func(*gin.Context) string {
	return ExtractIDFromParam("id")
}

// ExtractOrderIDFromParam extracts order ID from 'id' parameter
func ExtractOrderIDFromParam() func(*gin.Context) string {
	return ExtractIDFromParam("id")
}

// ExtractTenantIDFromContext extracts tenant ID from user context
func ExtractTenantIDFromContext() func(*gin.Context) string {
	return func(c *gin.Context) string {
		if userContext, exists := GetUserContext(c); exists {
			return userContext.TenantID
		}
		return ""
	}
}

// Catalog-specific permission constants
const (
	// Resources (additional to existing ones in authz.go)
	ResourceCatalog   = "catalog"
	ResourceInventory = "inventory"

	// Actions
	ActionCreate  = "create"
	ActionRead    = "read"
	ActionUpdate  = "update"
	ActionDelete  = "delete"
	ActionPublish = "publish"
	ActionList    = "list"
	ActionSearch  = "search"
)

// Convenience methods for common catalog operations

// RequireCatalogRead requires read permission on catalog
func (m *ProductionAuthMiddleware) RequireCatalogRead() gin.HandlerFunc {
	return m.RequirePermission(ResourceCatalog, ActionRead)
}

// RequireCatalogCreate requires create permission on catalog
func (m *ProductionAuthMiddleware) RequireCatalogCreate() gin.HandlerFunc {
	return m.RequirePermission(ResourceCatalog, ActionCreate)
}

// RequireCatalogUpdate requires update permission on specific catalog item
func (m *ProductionAuthMiddleware) RequireCatalogUpdate() gin.HandlerFunc {
	return m.RequireResourcePermission(ResourceCatalog, ActionUpdate, ExtractCatalogIDFromParam())
}

// RequireCatalogDelete requires delete permission on specific catalog item
func (m *ProductionAuthMiddleware) RequireCatalogDelete() gin.HandlerFunc {
	return m.RequireResourcePermission(ResourceCatalog, ActionDelete, ExtractCatalogIDFromParam())
}

// RequireCatalogPublish requires publish permission on specific catalog item
func (m *ProductionAuthMiddleware) RequireCatalogPublish() gin.HandlerFunc {
	return m.RequireResourcePermission(ResourceCatalog, ActionPublish, ExtractCatalogIDFromParam())
}
