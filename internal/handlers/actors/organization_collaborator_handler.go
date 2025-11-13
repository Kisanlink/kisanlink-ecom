// Package actors provides HTTP handlers for actor-related operations including organization collaborators.
package actors

import (
	"strconv"
	"strings"

	"kisanlink-ecom/entities/models/actors"
	"kisanlink-ecom/internal/common"
	actorsService "kisanlink-ecom/internal/services/actors"

	"github.com/gin-gonic/gin"
)

// OrganizationCollaboratorHandler handles HTTP requests for organization collaborator operations
type OrganizationCollaboratorHandler struct {
	service actorsService.OrganizationCollaboratorServiceInterface
}

// NewOrganizationCollaboratorHandler creates a new organization collaborator handler
func NewOrganizationCollaboratorHandler(service actorsService.OrganizationCollaboratorServiceInterface) *OrganizationCollaboratorHandler {
	return &OrganizationCollaboratorHandler{
		service: service,
	}
}

// CreateOrganizationCollaborator handles creating a new organization collaborator
func (h *OrganizationCollaboratorHandler) CreateOrganizationCollaborator(c *gin.Context) {
	var req struct {
		AAAEntityID    string                  `json:"aaa_entity_id" binding:"required"`
		AAAEntityType  actors.AAAEntityType    `json:"aaa_entity_type" binding:"required"`
		OrganizationID string                  `json:"organization_id" binding:"required"`
		Role           actors.CollaboratorRole `json:"role" binding:"required"`
		CreatedBy      string                  `json:"created_by" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	collaborator, err := h.service.CreateCollaborator(c.Request.Context(), req.AAAEntityID, req.AAAEntityType, req.OrganizationID, req.Role, req.CreatedBy)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			common.Error(c, 409, string(common.ErrorCodeDuplicateRecord), err.Error(), nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to create organization collaborator", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Created(c, collaborator, nil)
}

// GetOrganizationCollaborator handles getting an organization collaborator by ID
func (h *OrganizationCollaboratorHandler) GetOrganizationCollaborator(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Organization collaborator ID is required", nil)
		return
	}

	collaborator, err := h.service.GetCollaborator(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "Organization collaborator not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get organization collaborator", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, collaborator, nil)
}

// ListOrganizationCollaborators handles listing all organization collaborators
func (h *OrganizationCollaboratorHandler) ListOrganizationCollaborators(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	collaborators, total, err := h.service.ListCollaborators(c.Request.Context(), limit, offset)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to list organization collaborators", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	meta := &common.ResponseMeta{
		Pagination: common.NewPaginationMeta(page, limit, total),
	}

	common.Success(c, collaborators, meta)
}

// GetOrganizationCollaborators handles getting collaborators for a specific organization
func (h *OrganizationCollaboratorHandler) GetOrganizationCollaborators(c *gin.Context) {
	orgID := c.Param("organizationId")
	if orgID == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Organization ID is required", nil)
		return
	}

	collaborators, err := h.service.GetOrganizationCollaborators(c.Request.Context(), orgID)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get organization collaborators", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, collaborators, nil)
}

// UpdateOrganizationCollaborator handles updating an organization collaborator
func (h *OrganizationCollaboratorHandler) UpdateOrganizationCollaborator(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Organization collaborator ID is required", nil)
		return
	}

	var req struct {
		Role actors.CollaboratorRole `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Update role using service method
	updatedCollaborator, err := h.service.UpdateRole(c.Request.Context(), id, req.Role)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to update collaborator role", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, updatedCollaborator, nil)
}

// DeleteOrganizationCollaborator handles deleting an organization collaborator
func (h *OrganizationCollaboratorHandler) DeleteOrganizationCollaborator(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Organization collaborator ID is required", nil)
		return
	}

	err := h.service.DeleteCollaborator(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "Organization collaborator not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to delete organization collaborator", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	c.Status(204)
}

// InviteCollaborator handles inviting a collaborator to an organization
func (h *OrganizationCollaboratorHandler) InviteCollaborator(c *gin.Context) {
	var req struct {
		AAAEntityID     string                  `json:"aaa_entity_id" binding:"required"`
		AAAEntityType   actors.AAAEntityType    `json:"aaa_entity_type" binding:"required"`
		OrganizationID  string                  `json:"organization_id" binding:"required"`
		Role            actors.CollaboratorRole `json:"role" binding:"required"`
		InvitedByUserID string                  `json:"invited_by_user_id" binding:"required"`
		CreatedBy       string                  `json:"created_by" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	collaborator, err := h.service.InviteCollaborator(c.Request.Context(), req.AAAEntityID, req.AAAEntityType, req.OrganizationID, req.Role, req.InvitedByUserID, req.CreatedBy)
	if err != nil {
		if strings.Contains(err.Error(), "already invited") || strings.Contains(err.Error(), "already exists") {
			common.Error(c, 409, string(common.ErrorCodeDuplicateRecord), err.Error(), nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to invite collaborator", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Created(c, collaborator, nil)
}

// ActivateCollaborator handles activating a collaborator
func (h *OrganizationCollaboratorHandler) ActivateCollaborator(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Organization collaborator ID is required", nil)
		return
	}

	collaborator, err := h.service.ActivateCollaborator(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "Organization collaborator not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to activate collaborator", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, collaborator, nil)
}
