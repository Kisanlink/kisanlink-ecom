package middleware

import (
	"strings"
	"time"

	"kisanlink-ecom/entities/models/common"
	"kisanlink-ecom/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

// EnhancedAuthConfig holds configuration for enhanced authentication
type EnhancedAuthConfig struct {
	// SkipAuthPaths defines paths that skip authentication
	SkipAuthPaths []string
	// TokenCacheEnabled enables JWT token caching
	TokenCacheEnabled bool
	// TokenCacheTTL sets the cache TTL for tokens
	TokenCacheTTL time.Duration
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
		TokenCacheEnabled:   true,
		TokenCacheTTL:       2 * time.Minute,
		RequireOrganization: true,
	}
}

// EnhancedAuthMiddleware provides enhanced authentication with caching and organization validation
type EnhancedAuthMiddleware struct {
	aaaClient  auth.AAAClient
	tokenCache *cache.Cache
	config     *EnhancedAuthConfig
}

// NewEnhancedAuthMiddleware creates a new enhanced authentication middleware
func NewEnhancedAuthMiddleware(aaaClient auth.AAAClient, config *EnhancedAuthConfig) *EnhancedAuthMiddleware {
	if config == nil {
		config = DefaultEnhancedAuthConfig()
	}

	var tokenCache *cache.Cache
	if config.TokenCacheEnabled {
		tokenCache = cache.New(config.TokenCacheTTL, config.TokenCacheTTL*2)
	}

	return &EnhancedAuthMiddleware{
		aaaClient:  aaaClient,
		tokenCache: tokenCache,
		config:     config,
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
	valid, err := m.aaaClient.ValidateJWT(c.Request.Context(), token)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, err
	}

	// Get user context from token
	userContext, err := m.aaaClient.GetUserFromToken(c.Request.Context(), token)
	if err != nil {
		return nil, err
	}

	// Cache result if caching is enabled
	if m.tokenCache != nil && userContext != nil {
		m.tokenCache.Set(token, userContext, cache.DefaultExpiration)
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
