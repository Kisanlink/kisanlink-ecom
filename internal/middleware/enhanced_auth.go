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

// EnhancedAuthConfig holds configuration for enhanced authentication
type EnhancedAuthConfig struct {
	// SkipAuthPaths defines paths that skip authentication
	SkipAuthPaths []string
	// TokenCacheEnabled enables JWT token caching
	TokenCacheEnabled bool
	// TokenCacheTTL sets the cache TTL for tokens
	TokenCacheTTL time.Duration
	// PermissionCacheEnabled enables permission caching
	PermissionCacheEnabled bool
	// PermissionCacheTTL sets the cache TTL for permissions (short TTL as per requirement)
	PermissionCacheTTL time.Duration
	// RequireOrganization forces organization validation
	RequireOrganization bool
}

// DefaultEnhancedAuthConfig returns default auth configuration
func DefaultEnhancedAuthConfig() *EnhancedAuthConfig {
	return &EnhancedAuthConfig{
		SkipAuthPaths: []string{
			"/health", "/ready", "/metrics",
			"/docs", "/docs/",
			"/api/v1/auth/login", "/api/v1/auth/register",
		},
		TokenCacheEnabled:      true,
		TokenCacheTTL:          2 * time.Minute,
		PermissionCacheEnabled: true,
		PermissionCacheTTL:     30 * time.Second, // Short TTL for permissions as per requirement
		RequireOrganization:    true,
	}
}

// EnhancedAuthMiddleware provides enhanced authentication with caching and organization validation
type EnhancedAuthMiddleware struct {
	aaaClient       auth.Client
	tokenCache      *gocache.Cache
	permissionCache *gocache.Cache
	cacheManager    *cache.CacheManager
	config          *EnhancedAuthConfig
}

// NewEnhancedAuthMiddleware creates a new enhanced authentication middleware
func NewEnhancedAuthMiddleware(aaaClient auth.Client, config *EnhancedAuthConfig) *EnhancedAuthMiddleware {
	if config == nil {
		config = DefaultEnhancedAuthConfig()
	}

	var tokenCache *gocache.Cache
	if config.TokenCacheEnabled {
		tokenCache = gocache.New(config.TokenCacheTTL, config.TokenCacheTTL*2)
	}

	var permissionCache *gocache.Cache
	if config.PermissionCacheEnabled {
		permissionCache = gocache.New(config.PermissionCacheTTL, config.PermissionCacheTTL*2)
	}

	// Initialize cache manager for advanced caching features
	cacheManager := cache.GetCacheManager()

	return &EnhancedAuthMiddleware{
		aaaClient:       aaaClient,
		tokenCache:      tokenCache,
		permissionCache: permissionCache,
		cacheManager:    cacheManager,
		config:          config,
	}
}

// Middleware returns the authentication middleware function
func (m *EnhancedAuthMiddleware) Middleware() gin.HandlerFunc {
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

		// Validate token and get user context
		userContext, err := m.validateToken(c, token)
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
		c.Set("user_context", userContext)
		c.Set("user_id", userContext.UserID)
		c.Set("organization_id", userContext.OrganizationID)
		c.Set("user_roles", userContext.Roles)
		c.Set("user_permissions", userContext.Permissions)

		// Add user info to response headers for debugging (non-production)
		if gin.Mode() != gin.ReleaseMode {
			c.Header("X-User-ID", userContext.UserID)
			c.Header("X-Organization-ID", userContext.OrganizationID)
		}

		c.Next()
	}
}

// OptionalAuth provides optional authentication (sets user context if token is valid, but doesn't require it)
func (m *EnhancedAuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from request
		token, err := m.extractToken(c)
		if err != nil || token == "" {
			// No token or invalid token format - continue without authentication
			c.Next()
			return
		}

		// Validate token and get user context
		userContext, err := m.validateToken(c, token)
		if err != nil || userContext == nil {
			// Invalid token - continue without authentication
			c.Next()
			return
		}

		// Set user context if validation successful
		c.Set("user_context", userContext)
		c.Set("user_id", userContext.UserID)
		c.Set("organization_id", userContext.OrganizationID)
		c.Set("user_roles", userContext.Roles)
		c.Set("user_permissions", userContext.Permissions)

		c.Next()
	}
}

// RequireActiveUser ensures the authenticated user is active
func (m *EnhancedAuthMiddleware) RequireActiveUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userContext, exists := c.Get("user_context")
		if !exists {
			m.handleUnauthorized(c, "Authentication required")
			return
		}

		userCtx, ok := userContext.(*auth.UserContext)
		if !ok {
			m.handleUnauthorized(c, "Invalid user context")
			return
		}

		if !userCtx.IsActive {
			m.handleUnauthorized(c, "User account is inactive")
			return
		}

		c.Next()
	}
}

// Helper methods

func (m *EnhancedAuthMiddleware) shouldSkipAuth(path string) bool {
	for _, skipPath := range m.config.SkipAuthPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

func (m *EnhancedAuthMiddleware) extractToken(c *gin.Context) (string, error) {
	// Try Authorization header first
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// Support both "Bearer" and "bearer" (case insensitive)
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return strings.TrimSpace(parts[1]), nil
		}
	}

	// Try query parameter
	if token := c.Query("token"); token != "" {
		return token, nil
	}

	// Try cookie
	if token, err := c.Cookie("auth_token"); err == nil && token != "" {
		return token, nil
	}

	return "", nil
}

func (m *EnhancedAuthMiddleware) validateToken(c *gin.Context, token string) (*auth.UserContext, error) {
	// Check cache first if enabled
	if m.tokenCache != nil {
		if cached, found := m.tokenCache.Get(token); found {
			if userCtx, ok := cached.(*auth.UserContext); ok {
				return userCtx, nil
			}
		}
	}

	// Validate with AAA service
	claims, err := m.aaaClient.ValidateToken(c.Request.Context(), token)
	if err != nil {
		return nil, err
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

	// Cache result if caching is enabled
	if m.tokenCache != nil && userContext != nil {
		m.tokenCache.Set(token, userContext, gocache.DefaultExpiration)
	}

	return userContext, nil
}

func (m *EnhancedAuthMiddleware) handleUnauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(401, common.APIResponse{
		Success: false,
		Error: &common.APIError{
			Code:    "UNAUTHORIZED",
			Message: message,
		},
	})
}

func (m *EnhancedAuthMiddleware) handleAuthError(c *gin.Context, message string, err error) {
	details := ""
	if err != nil {
		details = err.Error()
	}

	c.AbortWithStatusJSON(401, common.APIResponse{
		Success: false,
		Error: &common.APIError{
			Code:    "AUTHENTICATION_ERROR",
			Message: message,
			Details: details,
		},
	})
}

// InvalidateTokenCache invalidates cached token
func (m *EnhancedAuthMiddleware) InvalidateTokenCache(token string) {
	if m.tokenCache != nil {
		m.tokenCache.Delete(token)
	}
}

// ClearTokenCache clears the entire token cache
func (m *EnhancedAuthMiddleware) ClearTokenCache() {
	if m.tokenCache != nil {
		m.tokenCache.Flush()
	}
}

// GetUserContext safely extracts user context from gin context
func GetUserContext(c *gin.Context) (*auth.UserContext, bool) {
	userContext, exists := c.Get("user_context")
	if !exists {
		return nil, false
	}

	userCtx, ok := userContext.(*auth.UserContext)
	return userCtx, ok
}

// MustGetUserContext extracts user context or panics (use only when you're sure user is authenticated)
func MustGetUserContext(c *gin.Context) *auth.UserContext {
	userCtx, ok := GetUserContext(c)
	if !ok {
		panic("User context not found - ensure authentication middleware is applied")
	}
	return userCtx
}

// AuthorizeMiddleware creates authorization middleware that checks permissions with caching
func (m *EnhancedAuthMiddleware) AuthorizeMiddleware(resource, action string, resourceIDExtractor func(*gin.Context) string) gin.HandlerFunc {
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

		c.Next()
	}
}

// RequirePermission creates middleware that requires specific permission
func (m *EnhancedAuthMiddleware) RequirePermission(resource, action string) gin.HandlerFunc {
	return m.AuthorizeMiddleware(resource, action, nil)
}

// RequireResourcePermission creates middleware that requires permission on a specific resource
func (m *EnhancedAuthMiddleware) RequireResourcePermission(resource, action string, resourceIDExtractor func(*gin.Context) string) gin.HandlerFunc {
	return m.AuthorizeMiddleware(resource, action, resourceIDExtractor)
}

// checkPermissionWithCache checks permission with caching support
func (m *EnhancedAuthMiddleware) checkPermissionWithCache(ctx context.Context, userContext *auth.UserContext, resource, action, resourceID string) (bool, error) {
	// Generate cache key for permission
	cacheKey := m.generatePermissionCacheKey(userContext.UserID, userContext.TenantID, resource, action, resourceID)

	// Check cache first if enabled
	if m.permissionCache != nil {
		if cached, found := m.permissionCache.Get(cacheKey); found {
			if allowed, ok := cached.(bool); ok {
				return allowed, nil
			}
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

	// Cache the result if caching is enabled
	if m.permissionCache != nil {
		m.permissionCache.Set(cacheKey, authResp.Allowed, gocache.DefaultExpiration)
	}

	return authResp.Allowed, nil
}

// generatePermissionCacheKey generates a cache key for permission checks
func (m *EnhancedAuthMiddleware) generatePermissionCacheKey(userID, tenantID, resource, action, resourceID string) string {
	// Create a deterministic cache key
	key := fmt.Sprintf("perm:%s:%s:%s:%s:%s", userID, tenantID, resource, action, resourceID)

	// Hash the key to ensure consistent length and avoid special characters
	hash := sha256.Sum256([]byte(key))
	return fmt.Sprintf("permission:%x", hash[:16]) // Use first 16 bytes of hash
}

// InvalidatePermissionCache invalidates permission cache for a user
func (m *EnhancedAuthMiddleware) InvalidatePermissionCache(userID string) {
	if m.permissionCache == nil {
		return
	}

	// Since go-cache doesn't support pattern deletion, we'd need to track keys
	// For now, we'll clear the entire cache (acceptable for short TTL)
	m.permissionCache.Flush()
}

// InvalidateUserPermissions invalidates all cached permissions for a specific user
func (m *EnhancedAuthMiddleware) InvalidateUserPermissions(userID, tenantID string) {
	if m.permissionCache == nil {
		return
	}

	// For a more sophisticated implementation, we could maintain a reverse index
	// For now, clear the entire cache since TTL is short (30 seconds)
	m.permissionCache.Flush()
}

// handleForbidden handles forbidden access
func (m *EnhancedAuthMiddleware) handleForbidden(c *gin.Context, message, resource, action, resourceID string) {
	c.AbortWithStatusJSON(http.StatusForbidden, common.APIResponse{
		Success: false,
		Error: &common.APIError{
			Code:    "FORBIDDEN",
			Message: message,
			Details: fmt.Sprintf("Access denied for action '%s' on resource '%s' (ID: %s)", action, resource, resourceID),
		},
	})
}
