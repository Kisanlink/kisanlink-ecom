package handlers

import (
	"context"
	"net/http"

	"kisanlink-ecom/internal/services"
	"kisanlink-ecom/internal/utils"

	"github.com/gin-gonic/gin"
)

// BaseHandler provides common functionality for role and permission handlers
type BaseHandler struct {
	rolePermissionService *services.RolePermissionService
}

// NewBaseHandler creates a new base handler instance
func NewBaseHandler(rolePermissionService *services.RolePermissionService) *BaseHandler {
	return &BaseHandler{
		rolePermissionService: rolePermissionService,
	}
}

// CreateEntity is a generic method for creating entities (roles or permissions)
func (h *BaseHandler) CreateEntity(c *gin.Context, req interface{}, createFunc func(ctx context.Context, name, description string) (interface{}, error), entityType string) {
	if err := c.ShouldBindJSON(req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	ctx := context.Background()

	// Extract name and description from request
	var name, description string
	switch r := req.(type) {
	case *CreateRoleRequest:
		name = r.Name
		description = r.Description
	case *CreatePermissionRequest:
		name = r.Name
		description = r.Description
	}

	entity, err := createFunc(ctx, name, description)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create "+entityType, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, entityType+" created successfully", gin.H{
		entityType: entity,
	})
}

// GetEntities is a generic method for getting all entities
func (h *BaseHandler) GetEntities(c *gin.Context, getAllFunc func(ctx context.Context) (interface{}, error), entityType string) {
	ctx := context.Background()

	entities, err := getAllFunc(ctx)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve "+entityType, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, entityType+" retrieved successfully", gin.H{
		entityType: entities,
	})
}

// GetEntity is a generic method for getting a specific entity by ID
func (h *BaseHandler) GetEntity(c *gin.Context, getByIDFunc func(ctx context.Context, id string) (interface{}, error), entityType string) {
	entityID := c.Param("id")
	if entityID == "" {
		utils.ValidationErrorResponse(c, entityType+" ID is required")
		return
	}

	ctx := context.Background()
	entity, err := getByIDFunc(ctx, entityID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve "+entityType, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, entityType+" retrieved successfully", gin.H{
		entityType: entity,
	})
}

// UpdateEntity is a generic method for updating entities
func (h *BaseHandler) UpdateEntity(c *gin.Context, req interface{}, updateFunc func(ctx context.Context, entity interface{}) (interface{}, error), entityType string) {
	entityID := c.Param("id")
	if entityID == "" {
		utils.ValidationErrorResponse(c, entityType+" ID is required")
		return
	}

	if err := c.ShouldBindJSON(req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	ctx := context.Background()

	// Create entity object based on type
	var entity interface{}
	switch r := req.(type) {
	case *UpdateRoleRequest:
		entity = &services.Role{
			ID:          entityID,
			Name:        r.Name,
			Description: r.Description,
		}
	case *UpdatePermissionRequest:
		entity = &services.Permission{
			ID:          entityID,
			Name:        r.Name,
			Description: r.Description,
		}
	}

	updatedEntity, err := updateFunc(ctx, entity)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update "+entityType, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, entityType+" updated successfully", gin.H{
		entityType: updatedEntity,
	})
}

// DeleteEntity is a generic method for deleting entities
func (h *BaseHandler) DeleteEntity(c *gin.Context, deleteFunc func(ctx context.Context, id string) error, entityType string) {
	entityID := c.Param("id")
	if entityID == "" {
		utils.ValidationErrorResponse(c, entityType+" ID is required")
		return
	}

	ctx := context.Background()
	err := deleteFunc(ctx, entityID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete "+entityType, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, entityType+" deleted successfully", nil)
}
