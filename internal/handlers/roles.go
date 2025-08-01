package handlers

import (
	"context"
	"fmt"
	"net/http"

	"kisanlink-ecom/internal/services"
	"kisanlink-ecom/internal/utils"

	"github.com/gin-gonic/gin"
)

// RoleHandler handles role-related HTTP requests
type RoleHandler struct {
	*BaseHandler
}

// NewRoleHandler creates a new role handler instance
func NewRoleHandler(rolePermissionService *services.RolePermissionService) *RoleHandler {
	return &RoleHandler{
		BaseHandler: NewBaseHandler(rolePermissionService),
	}
}

// Wrapper functions for type conversion
func (h *RoleHandler) createRoleWrapper(ctx context.Context, name, description string) (interface{}, error) {
	return h.rolePermissionService.CreateRole(ctx, name, description)
}

func (h *RoleHandler) getAllRolesWrapper(ctx context.Context) (interface{}, error) {
	return h.rolePermissionService.GetAllRoles(ctx)
}

func (h *RoleHandler) getRoleByIDWrapper(ctx context.Context, id string) (interface{}, error) {
	return h.rolePermissionService.GetRoleByID(ctx, id)
}

func (h *RoleHandler) updateRoleWrapper(ctx context.Context, entity any) (interface{}, error) {
	role, ok := entity.(*services.Role)
	if !ok {
		return nil, fmt.Errorf("invalid entity type for role update")
	}
	return h.rolePermissionService.UpdateRole(ctx, role)
}

// CreateRole handles role creation.
// @Summary      Create Role
// @Description  Create a new role with hierarchical structure
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Param        request  body      CreateRoleRequest  true  "Role data"
// @Success      201      {object}  object  "Role created successfully"
// @Failure      400      {object}  object  "Invalid request"
// @Failure      500      {object}  object  "Internal server error"
// @Router       /api/v1/roles [post]
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	h.CreateEntity(c, &req, h.createRoleWrapper, "role")
}

// GetRoles handles getting all roles.
// @Summary      Get All Roles
// @Description  Retrieve a list of all roles
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Success      200      {object}  object  "Roles retrieved successfully"
// @Failure      500      {object}  object  "Internal server error"
// @Router       /api/v1/roles [get]
func (h *RoleHandler) GetRoles(c *gin.Context) {
	h.GetEntities(c, h.getAllRolesWrapper, "roles")
}

// GetRole handles getting a specific role by ID.
// @Summary      Get Role by ID
// @Description  Retrieve a specific role by their ID
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Role ID"
// @Success      200  {object}  object  "Role retrieved successfully"
// @Failure      400  {object}  object  "Invalid role ID"
// @Failure      404  {object}  object  "Role not found"
// @Failure      500  {object}  object  "Internal server error"
// @Router       /api/v1/roles/{id} [get]
func (h *RoleHandler) GetRole(c *gin.Context) {
	h.GetEntity(c, h.getRoleByIDWrapper, "role")
}

// UpdateRole handles role updates.
// @Summary      Update Role
// @Description  Update an existing role's information
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Param        id      path      string                true   "Role ID"
// @Param        request body      UpdateRoleRequest     true  "Role update data"
// @Success      200      {object}  object  "Role updated successfully"
// @Failure      400      {object}  object  "Invalid request"
// @Failure      404      {object}  object  "Role not found"
// @Failure      500      {object}  object  "Internal server error"
// @Router       /api/v1/roles/{id} [put]
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	var req UpdateRoleRequest
	h.UpdateEntity(c, &req, h.updateRoleWrapper, "role")
}

// DeleteRole handles role deletion.
// @Summary      Delete Role
// @Description  Delete a role
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Role ID"
// @Success      200  {object}  object  "Role deleted successfully"
// @Failure      400  {object}  object  "Invalid role ID"
// @Failure      404  {object}  object  "Role not found"
// @Failure      500  {object}  object  "Internal server error"
// @Router       /api/v1/roles/{id} [delete]
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	h.DeleteEntity(c, h.rolePermissionService.DeleteRole, "role")
}

// Request/Response models

// CreateRoleRequest represents the request to create a new role
type CreateRoleRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=500"`
}

// UpdateRoleRequest represents the request to update a role
type UpdateRoleRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=500"`
}

// Legacy handler functions for backward compatibility
var defaultRoleHandler *RoleHandler

// SetDefaultRoleHandler sets the default role handler
func SetDefaultRoleHandler(handler *RoleHandler) {
	defaultRoleHandler = handler
}

// GetRoles is the legacy function that delegates to the handler
func GetRoles(c *gin.Context) {
	if defaultRoleHandler == nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Role handler not initialized", "")
		return
	}
	defaultRoleHandler.GetRoles(c)
}

// CreateRole is the legacy function that delegates to the handler
func CreateRole(c *gin.Context) {
	if defaultRoleHandler == nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Role handler not initialized", "")
		return
	}
	defaultRoleHandler.CreateRole(c)
}

// GetRole is the legacy function that delegates to the handler
func GetRole(c *gin.Context) {
	if defaultRoleHandler == nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Role handler not initialized", "")
		return
	}
	defaultRoleHandler.GetRole(c)
}

// UpdateRole is the legacy function that delegates to the handler
func UpdateRole(c *gin.Context) {
	if defaultRoleHandler == nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Role handler not initialized", "")
		return
	}
	defaultRoleHandler.UpdateRole(c)
}

// DeleteRole is the legacy function that delegates to the handler
func DeleteRole(c *gin.Context) {
	if defaultRoleHandler == nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Role handler not initialized", "")
		return
	}
	defaultRoleHandler.DeleteRole(c)
}
