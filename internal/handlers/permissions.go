package handlers

import (
	"context"
	"fmt"
	"net/http"

	"kisanlink-ecom/internal/services"
	"kisanlink-ecom/internal/utils"

	"github.com/gin-gonic/gin"
)

// PermissionHandler handles permission-related HTTP requests
type PermissionHandler struct {
	*BaseHandler
}

// NewPermissionHandler creates a new permission handler instance
func NewPermissionHandler(rolePermissionService *services.RolePermissionService) *PermissionHandler {
	return &PermissionHandler{
		BaseHandler: NewBaseHandler(rolePermissionService),
	}
}

// Wrapper functions for type conversion
func (h *PermissionHandler) createPermissionWrapper(ctx context.Context, name, description string) (interface{}, error) {
	return h.rolePermissionService.CreatePermission(ctx, name, description)
}

func (h *PermissionHandler) getAllPermissionsWrapper(ctx context.Context) (interface{}, error) {
	return h.rolePermissionService.GetAllPermissions(ctx)
}

func (h *PermissionHandler) getPermissionByIDWrapper(ctx context.Context, id string) (interface{}, error) {
	return h.rolePermissionService.GetPermissionByID(ctx, id)
}

func (h *PermissionHandler) updatePermissionWrapper(ctx context.Context, entity interface{}) (interface{}, error) {
	permission, ok := entity.(*services.Permission)
	if !ok {
		return nil, fmt.Errorf("invalid entity type for permission update")
	}
	return h.rolePermissionService.UpdatePermission(ctx, permission)
}

// CreatePermission handles permission creation.
// @Summary      Create Permission
// @Description  Create a new permission with specific resource and action
// @Tags         Permissions
// @Accept       json
// @Produce      json
// @Param        request  body      CreatePermissionRequest  true  "Permission data"
// @Success      201      {object}  object  "Permission created successfully"
// @Failure      400      {object}  object  "Invalid request"
// @Failure      500      {object}  object  "Internal server error"
// @Router       /api/v1/permissions [post]
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var req CreatePermissionRequest
	h.CreateEntity(c, &req, h.createPermissionWrapper, "permission")
}

// GetPermissions handles getting all permissions.
// @Summary      Get All Permissions
// @Description  Retrieve a list of all permissions
// @Tags         Permissions
// @Accept       json
// @Produce      json
// @Success      200      {object}  object  "Permissions retrieved successfully"
// @Failure      500      {object}  object  "Internal server error"
// @Router       /api/v1/permissions [get]
func (h *PermissionHandler) GetPermissions(c *gin.Context) {
	h.GetEntities(c, h.getAllPermissionsWrapper, "permissions")
}

// GetPermission handles getting a specific permission by ID.
// @Summary      Get Permission by ID
// @Description  Retrieve a specific permission by their ID
// @Tags         Permissions
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Permission ID"
// @Success      200  {object}  object  "Permission retrieved successfully"
// @Failure      400  {object}  object  "Invalid permission ID"
// @Failure      404  {object}  object  "Permission not found"
// @Failure      500  {object}  object  "Internal server error"
// @Router       /api/v1/permissions/{id} [get]
func (h *PermissionHandler) GetPermission(c *gin.Context) {
	h.GetEntity(c, h.getPermissionByIDWrapper, "permission")
}

// UpdatePermission handles permission updates.
// @Summary      Update Permission
// @Description  Update an existing permission's information
// @Tags         Permissions
// @Accept       json
// @Produce      json
// @Param        id      path      string                true   "Permission ID"
// @Param        request body      UpdatePermissionRequest     true  "Permission update data"
// @Success      200      {object}  object  "Permission updated successfully"
// @Failure      400      {object}  object  "Invalid request"
// @Failure      404      {object}  object  "Permission not found"
// @Failure      500      {object}  object  "Internal server error"
// @Router       /api/v1/permissions/{id} [put]
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	var req UpdatePermissionRequest
	h.UpdateEntity(c, &req, h.updatePermissionWrapper, "permission")
}

// DeletePermission handles permission deletion.
// @Summary      Delete Permission
// @Description  Delete a permission
// @Tags         Permissions
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Permission ID"
// @Success      200  {object}  object  "Permission deleted successfully"
// @Failure      400  {object}  object  "Invalid permission ID"
// @Failure      404  {object}  object  "Permission not found"
// @Failure      500  {object}  object  "Internal server error"
// @Router       /api/v1/permissions/{id} [delete]
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	h.DeleteEntity(c, h.rolePermissionService.DeletePermission, "permission")
}

// Request/Response models

// CreatePermissionRequest represents the request to create a new permission
type CreatePermissionRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=500"`
}

// UpdatePermissionRequest represents the request to update a permission
type UpdatePermissionRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=500"`
}

// Legacy handler functions for backward compatibility
var defaultPermissionHandler *PermissionHandler

// SetDefaultPermissionHandler sets the default permission handler
func SetDefaultPermissionHandler(handler *PermissionHandler) {
	defaultPermissionHandler = handler
}

// GetPermissions is the legacy function that delegates to the handler
func GetPermissions(c *gin.Context) {
	if defaultPermissionHandler == nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Permission handler not initialized", "")
		return
	}
	defaultPermissionHandler.GetPermissions(c)
}

// CreatePermission is the legacy function that delegates to the handler
func CreatePermission(c *gin.Context) {
	if defaultPermissionHandler == nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Permission handler not initialized", "")
		return
	}
	defaultPermissionHandler.CreatePermission(c)
}

// GetPermission is the legacy function that delegates to the handler
func GetPermission(c *gin.Context) {
	if defaultPermissionHandler == nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Permission handler not initialized", "")
		return
	}
	defaultPermissionHandler.GetPermission(c)
}

// UpdatePermission is the legacy function that delegates to the handler
func UpdatePermission(c *gin.Context) {
	if defaultPermissionHandler == nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Permission handler not initialized", "")
		return
	}
	defaultPermissionHandler.UpdatePermission(c)
}

// DeletePermission is the legacy function that delegates to the handler
func DeletePermission(c *gin.Context) {
	if defaultPermissionHandler == nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Permission handler not initialized", "")
		return
	}
	defaultPermissionHandler.DeletePermission(c)
}
