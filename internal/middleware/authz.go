package middleware

import (
	"net/http"

	"github.com/Kisanlink/kisanlink-ecom/internal/auth"
	"github.com/Kisanlink/kisanlink-ecom/internal/common"

	"github.com/gin-gonic/gin"
)

// Resource and Action constants for permission checking
type Resource = string
type Action = string

const (
	ResourceOrg   Resource = "org"
	ResourceOrder Resource = "order"
	ResourceItem  Resource = "catalog_item"
)

const (
	ActCreateCatalog Action = "create_catalog"
	ActUpdateCatalog Action = "update_catalog"
	ActViewCatalog   Action = "view_catalog"
	ActBuy           Action = "buy"
	ActSell          Action = "sell"
	ActViewOrder     Action = "view_order"
	ActFulfillOrder  Action = "fulfill_order"
)

// InferResourceID is a function type that extracts the resource ID from the gin context
type InferResourceID func(*gin.Context) (string, error)

// AuthZ creates authorization middleware that checks permissions
func AuthZ(resourceType Resource, action Action, inferResourceID InferResourceID) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the authenticated user's ID from context
		subjectID, exists := GetSubjectID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authentication required",
				},
			})
			c.Abort()
			return
		}

		// Infer the resource ID from the context
		resourceID, err := inferResourceID(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "INVALID_RESOURCE",
					"message": "Failed to identify resource",
					"details": gin.H{
						"reason": err.Error(),
					},
				},
			})
			c.Abort()
			return
		}

		// Get AAA client from context (should be set by dependency injection)
		aaaClient, exists := c.Get("aaaClient")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Authorization service unavailable",
				},
			})
			c.Abort()
			return
		}

		client, ok := aaaClient.(auth.Client)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Invalid authorization client",
				},
			})
			c.Abort()
			return
		}

		// Check permission with AAA service using the new interface
		req := &auth.AuthorizeRequest{
			UserID:     subjectID,
			Resource:   string(resourceType),
			Action:     string(action),
			ResourceID: resourceID,
		}
		resp, err := client.Authorize(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "AUTHORIZATION_ERROR",
					"message": "Failed to check permissions",
					"details": gin.H{
						"reason": err.Error(),
					},
				},
			})
			c.Abort()
			return
		}

		if !resp.Allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "Access denied",
					"details": gin.H{
						"reason":      "insufficient permissions",
						"resource":    resourceType,
						"action":      action,
						"resource_id": resourceID,
					},
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Common resource ID inference functions

// InferOrgIDFromBody extracts org_id from request body
func InferOrgIDFromBody() InferResourceID {
	return func(c *gin.Context) (string, error) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			return "", err
		}

		if orgID, exists := body["org_id"]; exists {
			if id, ok := orgID.(string); ok && id != "" {
				return id, nil
			}
		}

		return "", common.ErrMissingRequiredField
	}
}

// InferOrgIDFromParam extracts org_id from URL parameter
func InferOrgIDFromParam(paramName string) InferResourceID {
	return func(c *gin.Context) (string, error) {
		orgID := c.Param(paramName)
		if orgID == "" {
			return "", common.ErrMissingRequiredField
		}
		return orgID, nil
	}
}

// InferOrderIDFromParam extracts order_id from URL parameter
func InferOrderIDFromParam() InferResourceID {
	return func(c *gin.Context) (string, error) {
		orderID := c.Param("id")
		if orderID == "" {
			return "", common.ErrMissingRequiredField
		}
		return orderID, nil
	}
}

// InferItemIDFromParam extracts item_id from URL parameter
func InferItemIDFromParam() InferResourceID {
	return func(c *gin.Context) (string, error) {
		itemID := c.Param("id")
		if itemID == "" {
			return "", common.ErrMissingRequiredField
		}
		return itemID, nil
	}
}
