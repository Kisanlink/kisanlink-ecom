// Package roles provides HTTP handlers for role management operations including user, organization, and ecommerce roles.
package roles //nolint:dupl // stub handlers have similar structure

import (
	"strconv"
	"strings"

	"github.com/Kisanlink/kisanlink-ecom/internal/common"
	rolesService "github.com/Kisanlink/kisanlink-ecom/internal/services/roles"

	"github.com/gin-gonic/gin"
)

// OrganizationRoleHandler handles HTTP requests for organization role operations
type OrganizationRoleHandler struct {
	service rolesService.OrganizationRoleServiceInterface
}

// NewOrganizationRoleHandler creates a new organization role handler
func NewOrganizationRoleHandler(service rolesService.OrganizationRoleServiceInterface) *OrganizationRoleHandler {
	return &OrganizationRoleHandler{
		service: service,
	}
}

// CreateOrganizationRole handles creating a new organization role
func (h *OrganizationRoleHandler) CreateOrganizationRole(c *gin.Context) {
	var req struct {
		OrganizationID    string          `json:"organization_id" binding:"required"`
		AAARoleID         string          `json:"aaa_role_id" binding:"required"`
		ConfiguredBy      string          `json:"configured_by" binding:"required"`
		CustomPermissions map[string]bool `json:"custom_permissions"`
		MaxUsers          *int            `json:"max_users"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	orgRole, err := h.service.CreateOrganizationRole(c.Request.Context(), req.OrganizationID, req.AAARoleID, req.ConfiguredBy, req.CustomPermissions, req.MaxUsers)
	if err != nil {
		if strings.Contains(err.Error(), "already has this role") {
			common.Error(c, 409, string(common.ErrorCodeDuplicateRecord), err.Error(), nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to create organization role", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Created(c, orgRole, nil)
}

// GetOrganizationRole handles getting an organization role by ID
func (h *OrganizationRoleHandler) GetOrganizationRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Organization role ID is required", nil)
		return
	}

	orgRole, err := h.service.GetOrganizationRole(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "Organization role not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get organization role", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, orgRole, nil)
}

// ListOrganizationRoles handles listing all organization roles
func (h *OrganizationRoleHandler) ListOrganizationRoles(c *gin.Context) {
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

	orgRoles, total, err := h.service.ListOrganizationRoles(c.Request.Context(), limit, offset)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to list organization roles", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	meta := &common.ResponseMeta{
		Pagination: common.NewPaginationMeta(page, limit, total),
	}

	common.Success(c, orgRoles, meta)
}

// GetOrganizationRoles handles getting roles for a specific organization
func (h *OrganizationRoleHandler) GetOrganizationRoles(c *gin.Context) {
	organizationID := c.Param("organizationId")
	if organizationID == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Organization ID is required", nil)
		return
	}

	orgRoles, err := h.service.GetOrganizationRoles(c.Request.Context(), organizationID)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get organization roles", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, orgRoles, nil)
}

// SetDefaultRole handles setting the default role for an organization
func (h *OrganizationRoleHandler) SetDefaultRole(c *gin.Context) {
	organizationID := c.Param("organizationId")
	roleID := c.Param("roleId")

	if organizationID == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Organization ID is required", nil)
		return
	}
	if roleID == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Role ID is required", nil)
		return
	}

	err := h.service.SetAsDefault(c.Request.Context(), organizationID, roleID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), err.Error(), nil)
			return
		}
		if strings.Contains(err.Error(), "does not belong") {
			common.BadRequest(c, string(common.ErrorCodeInvalidInput), err.Error(), nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to set default role", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, gin.H{"message": "Default role set successfully"}, nil)
}

// UpdateOrganizationRole handles updating an organization role
func (h *OrganizationRoleHandler) UpdateOrganizationRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Organization role ID is required", nil)
		return
	}

	// Get existing organization role
	existingOrgRole, err := h.service.GetOrganizationRole(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "Organization role not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get organization role", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Bind update request
	var req struct {
		IsActive          *bool           `json:"is_active"`
		CustomPermissions map[string]bool `json:"custom_permissions"`
		MaxUsers          *int            `json:"max_users"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Update fields if provided
	if req.IsActive != nil {
		existingOrgRole.IsActive = *req.IsActive
	}
	if req.MaxUsers != nil {
		existingOrgRole.MaxUsers = req.MaxUsers
	}

	// Update custom permissions separately if provided
	if req.CustomPermissions != nil {
		updatedOrgRole, err := h.service.UpdateCustomPermissions(c.Request.Context(), id, req.CustomPermissions)
		if err != nil {
			common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to update custom permissions", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		existingOrgRole = updatedOrgRole
	}

	updatedOrgRole, err := h.service.UpdateOrganizationRole(c.Request.Context(), existingOrgRole)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to update organization role", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, updatedOrgRole, nil)
}

// DeleteOrganizationRole handles deleting an organization role
func (h *OrganizationRoleHandler) DeleteOrganizationRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Organization role ID is required", nil)
		return
	}

	err := h.service.DeleteOrganizationRole(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "Organization role not found", nil)
			return
		}
		if strings.Contains(err.Error(), "cannot delete default role") {
			common.BadRequest(c, string(common.ErrorCodeBusinessRuleViolation), err.Error(), nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to delete organization role", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	c.Status(204) // No Content
}
