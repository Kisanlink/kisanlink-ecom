// Package roles provides HTTP handlers for role management operations including user, organization, and ecommerce roles.
package roles

import (
	"strconv"
	"strings"
	"time"

	"kisanlink-ecom/internal/common"
	rolesService "kisanlink-ecom/internal/services/roles"

	"github.com/gin-gonic/gin"
)

// UserRoleHandler handles HTTP requests for user role operations
type UserRoleHandler struct {
	service rolesService.UserRoleServiceInterface
}

// NewUserRoleHandler creates a new user role handler
func NewUserRoleHandler(service rolesService.UserRoleServiceInterface) *UserRoleHandler {
	return &UserRoleHandler{
		service: service,
	}
}

// AssignRole handles assigning a role to a user
func (h *UserRoleHandler) AssignRole(c *gin.Context) {
	var req struct {
		UserID     string     `json:"user_id" binding:"required"`
		RoleID     string     `json:"role_id" binding:"required"`
		AssignedBy string     `json:"assigned_by" binding:"required"`
		ExpiresAt  *time.Time `json:"expires_at"`
		Notes      string     `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	userRole, err := h.service.AssignRole(c.Request.Context(), req.UserID, req.RoleID, req.AssignedBy, req.ExpiresAt, req.Notes)
	if err != nil {
		if strings.Contains(err.Error(), "already has this role") {
			common.Error(c, 409, string(common.ErrorCodeDuplicateRecord), err.Error(), nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to assign role", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Created(c, userRole, nil)
}

// GetUserRole handles getting a user role by ID
func (h *UserRoleHandler) GetUserRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "User role ID is required", nil)
		return
	}

	userRole, err := h.service.GetUserRole(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "User role not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get user role", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, userRole, nil)
}

// ListUserRoles handles listing all user roles
func (h *UserRoleHandler) ListUserRoles(c *gin.Context) {
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

	userRoles, total, err := h.service.ListUserRoles(c.Request.Context(), limit, offset)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to list user roles", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	meta := &common.ResponseMeta{
		Pagination: common.NewPaginationMeta(page, limit, total),
	}

	common.Success(c, userRoles, meta)
}

// GetUserRoles handles getting roles for a specific user
func (h *UserRoleHandler) GetUserRoles(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "User ID is required", nil)
		return
	}

	userRoles, err := h.service.GetUserRoles(c.Request.Context(), userID)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get user roles", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, userRoles, nil)
}

// GetRoleUsers handles getting users for a specific role
func (h *UserRoleHandler) GetRoleUsers(c *gin.Context) {
	roleID := c.Param("roleId")
	if roleID == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Role ID is required", nil)
		return
	}

	userRoles, err := h.service.GetRoleUsers(c.Request.Context(), roleID)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get role users", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, userRoles, nil)
}

// UpdateUserRole handles updating a user role
func (h *UserRoleHandler) UpdateUserRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "User role ID is required", nil)
		return
	}

	// Get existing user role
	existingUserRole, err := h.service.GetUserRole(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "User role not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get user role", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Bind update request
	var req struct {
		IsActive  *bool      `json:"is_active"`
		ExpiresAt *time.Time `json:"expires_at"`
		Notes     *string    `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Update fields if provided
	if req.IsActive != nil {
		existingUserRole.IsActive = *req.IsActive
	}
	if req.ExpiresAt != nil {
		existingUserRole.ExpiresAt = req.ExpiresAt
	}
	if req.Notes != nil {
		existingUserRole.Notes = *req.Notes
	}

	updatedUserRole, err := h.service.UpdateUserRole(c.Request.Context(), existingUserRole)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to update user role", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, updatedUserRole, nil)
}

// DeleteUserRole handles deleting a user role
func (h *UserRoleHandler) DeleteUserRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "User role ID is required", nil)
		return
	}

	err := h.service.RevokeRole(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "User role not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to delete user role", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	c.Status(204) // No Content
}

// RevokeRole handles revoking a role from a user
func (h *UserRoleHandler) RevokeRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "User role ID is required", nil)
		return
	}

	err := h.service.RevokeRole(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "User role not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to revoke role", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	c.Status(204) // No Content
}
