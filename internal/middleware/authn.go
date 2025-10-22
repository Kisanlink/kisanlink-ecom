package middleware

import (
	"net/http"

	"kisanlink-ecom/internal/auth"

	"github.com/gin-gonic/gin"
)

const (
	// SubjectIDKey is the key used to store the authenticated user's ID in the context
	SubjectIDKey = "subjectID"
	// UserRolesKey is the key used to store the authenticated user's roles in the context
	UserRolesKey = "userRoles"
	// OrgIDKey is the key used to store the authenticated user's organization ID in the context
	OrgIDKey = "orgID"
)

// AuthNMiddleware creates authentication middleware that validates JWT tokens
func AuthNMiddleware(aaaClient auth.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from request
		token, err := auth.ExtractToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authentication required",
					"details": gin.H{
						"reason": err.Error(),
					},
				},
			})
			c.Abort()
			return
		}

		// Validate token with AAA service
		claims, err := aaaClient.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "INVALID_TOKEN",
					"message": "Token validation failed",
					"details": gin.H{
						"reason": err.Error(),
					},
				},
			})
			c.Abort()
			return
		}

		// Store user information in context for downstream handlers
		c.Set(SubjectIDKey, claims.UserID)
		c.Set(UserRolesKey, claims.Roles)
		c.Set(OrgIDKey, claims.OrganizationID)
		c.Set("user_id", claims.UserID) // For backward compatibility with handlers
		c.Set("organization_id", claims.OrganizationID)

		// Create enhanced user context with all available fields
		userCtx := &auth.UserContext{
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

		c.Set("userContext", userCtx)
		c.Set("user_context", userCtx)

		c.Next()
	}
}

// GetSubjectID retrieves the authenticated user's ID from the context
func GetSubjectID(c *gin.Context) (string, bool) {
	subjectID, exists := c.Get(SubjectIDKey)
	if !exists {
		return "", false
	}

	if id, ok := subjectID.(string); ok {
		return id, true
	}
	return "", false
}

// GetUserRoles retrieves the authenticated user's roles from the context
func GetUserRoles(c *gin.Context) ([]string, bool) {
	roles, exists := c.Get(UserRolesKey)
	if !exists {
		return nil, false
	}

	if userRoles, ok := roles.([]string); ok {
		return userRoles, true
	}
	return nil, false
}

// GetOrgID retrieves the organization ID from the context
func GetOrgID(c *gin.Context) (string, bool) {
	org, exists := c.Get(OrgIDKey)
	if !exists {
		return "", false
	}
	if id, ok := org.(string); ok {
		return id, true
	}
	return "", false
}

// GetOrganizationID retrieves the organization ID from the context with error handling
// Returns the organization ID or an empty string if not found
func GetOrganizationID(c *gin.Context) string {
	// Try to get from OrgIDKey first (set by auth middleware)
	if orgID, exists := GetOrgID(c); exists && orgID != "" {
		return orgID
	}

	// Fallback to organization_id key
	if orgID, exists := c.Get("organization_id"); exists {
		if id, ok := orgID.(string); ok && id != "" {
			return id
		}
	}

	// Last resort: check user context
	if userCtx, exists := c.Get("userContext"); exists {
		if ctx, ok := userCtx.(*auth.UserContext); ok && ctx.OrganizationID != "" {
			return ctx.OrganizationID
		}
	}

	return ""
}

// RequireAuth ensures that the request has been authenticated
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := GetSubjectID(c); !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authentication required",
				},
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
