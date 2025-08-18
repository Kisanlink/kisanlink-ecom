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
)

// AuthNMiddleware creates authentication middleware that validates JWT tokens
func AuthNMiddleware(aaaClient auth.AAAClient) gin.HandlerFunc {
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
		userID, roles, err := aaaClient.ValidateToken(c.Request.Context(), token)
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
		c.Set(SubjectIDKey, userID)
		c.Set(UserRolesKey, roles)

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
