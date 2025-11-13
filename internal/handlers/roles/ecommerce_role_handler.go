// Package roles provides HTTP handlers for role management operations including user, organization, and ecommerce roles.
package roles //nolint:dupl // stub handlers have similar structure

import (
	"strconv"
	"strings"

	"kisanlink-ecom/internal/common"
	rolesService "kisanlink-ecom/internal/services/roles"

	"github.com/gin-gonic/gin"
)

// EcommerceRoleHandler handles HTTP requests for ecommerce role operations
type EcommerceRoleHandler struct {
	service rolesService.EcommerceRoleServiceInterface
}

// NewEcommerceRoleHandler creates a new ecommerce role handler
func NewEcommerceRoleHandler(service rolesService.EcommerceRoleServiceInterface) *EcommerceRoleHandler {
	return &EcommerceRoleHandler{
		service: service,
	}
}

// CreateEcommerceRole handles creating a new ecommerce role
func (h *EcommerceRoleHandler) CreateEcommerceRole(c *gin.Context) {
	var req struct {
		AAARoleID      string  `json:"aaa_role_id" binding:"required"`
		RoleName       string  `json:"role_name" binding:"required"`
		Description    string  `json:"description"`
		OrganizationID *string `json:"organization_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	ecomRole, err := h.service.CreateEcommerceRole(c.Request.Context(), req.AAARoleID, req.RoleName, req.Description, req.OrganizationID)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			common.Error(c, 409, string(common.ErrorCodeDuplicateRecord), err.Error(), nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to create ecommerce role", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Created(c, ecomRole, nil)
}

// GetEcommerceRole handles getting an ecommerce role by ID
func (h *EcommerceRoleHandler) GetEcommerceRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Ecommerce role ID is required", nil)
		return
	}

	ecomRole, err := h.service.GetEcommerceRole(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "Ecommerce role not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get ecommerce role", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, ecomRole, nil)
}

// ListEcommerceRoles handles listing all ecommerce roles
func (h *EcommerceRoleHandler) ListEcommerceRoles(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	ecomRoles, total, err := h.service.ListEcommerceRoles(c.Request.Context(), limit, offset)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to list ecommerce roles", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	meta := &common.ResponseMeta{
		Pagination: common.NewPaginationMeta(page, limit, total),
	}

	common.Success(c, ecomRoles, meta)
}

// GetEcommerceRolesByOrganization handles getting ecommerce roles for a specific organization
func (h *EcommerceRoleHandler) GetEcommerceRolesByOrganization(c *gin.Context) {
	organizationID := c.Param("organizationId")
	if organizationID == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Organization ID is required", nil)
		return
	}

	ecomRoles, err := h.service.GetOrganizationRoles(c.Request.Context(), organizationID)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get organization roles", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, ecomRoles, nil)
}

// CheckPermission handles checking if a role has a specific permission
func (h *EcommerceRoleHandler) CheckPermission(c *gin.Context) {
	roleID := c.Param("roleId")
	permission := c.Query("permission")

	if roleID == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Role ID is required", nil)
		return
	}
	if permission == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Permission is required", nil)
		return
	}

	hasPermission, err := h.service.CheckPermission(c.Request.Context(), roleID, permission)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "Ecommerce role not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to check permission", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, gin.H{
		"role_id":        roleID,
		"permission":     permission,
		"has_permission": hasPermission,
	}, nil)
}

// UpdateEcommerceRole handles updating an ecommerce role
func (h *EcommerceRoleHandler) UpdateEcommerceRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Ecommerce role ID is required", nil)
		return
	}

	// Get existing ecommerce role
	existingEcomRole, err := h.service.GetEcommerceRole(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "Ecommerce role not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get ecommerce role", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Bind update request
	var req struct {
		RoleName       *string         `json:"role_name"`
		Description    *string         `json:"description"`
		IsActive       *bool           `json:"is_active"`
		Permissions    map[string]bool `json:"permissions"`
		OrganizationID *string         `json:"organization_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Update fields if provided
	if req.RoleName != nil {
		existingEcomRole.RoleName = *req.RoleName
	}
	if req.Description != nil {
		existingEcomRole.Description = *req.Description
	}
	if req.IsActive != nil {
		existingEcomRole.IsActive = *req.IsActive
	}

	// Update permissions separately if provided
	if req.Permissions != nil {
		updatedEcomRole, err := h.service.UpdatePermissions(c.Request.Context(), id, req.Permissions)
		if err != nil {
			common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to update permissions", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		existingEcomRole = updatedEcomRole
	}

	// Update organization scope if provided
	if req.OrganizationID != nil && *req.OrganizationID != "" {
		updatedEcomRole, err := h.service.SetOrganizationScope(c.Request.Context(), id, *req.OrganizationID)
		if err != nil {
			common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to set organization scope", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		existingEcomRole = updatedEcomRole
	}

	updatedEcomRole, err := h.service.UpdateEcommerceRole(c.Request.Context(), existingEcomRole)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to update ecommerce role", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, updatedEcomRole, nil)
}

// DeleteEcommerceRole handles deleting an ecommerce role
func (h *EcommerceRoleHandler) DeleteEcommerceRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Ecommerce role ID is required", nil)
		return
	}

	err := h.service.DeleteEcommerceRole(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "Ecommerce role not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to delete ecommerce role", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	c.Status(204) // No Content
}
