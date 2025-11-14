// Package user provides HTTP handlers for user management operations.
package user

import (
	"strconv"
	"strings"

	"github.com/Kisanlink/kisanlink-ecom/entities/requests/auth"
	"github.com/Kisanlink/kisanlink-ecom/internal/common"
	userService "github.com/Kisanlink/kisanlink-ecom/internal/services/user"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for user operations
type Handler struct {
	service *userService.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(service *userService.UserService) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateUser handles creating a new user
func (h *Handler) CreateUser(c *gin.Context) {
	var req auth.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	user, err := h.service.CreateUser(c.Request.Context(), req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			common.Error(c, 409, string(common.ErrorCodeDuplicateRecord), err.Error(), nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to create user", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Created(c, user, nil)
}

// GetUser handles getting a user by ID
func (h *Handler) GetUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "User ID is required", nil)
		return
	}

	user, err := h.service.GetUserByID(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "User not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get user", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, user, nil)
}

// GetUserByUsername handles getting a user by username
func (h *Handler) GetUserByUsername(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Username is required", nil)
		return
	}

	user, err := h.service.GetUserByUsername(c.Request.Context(), username)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "User not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get user", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, user, nil)
}

// GetUserByEmail handles getting a user by email
func (h *Handler) GetUserByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Email is required", nil)
		return
	}

	user, err := h.service.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "User not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get user", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, user, nil)
}

// ListUsers handles listing all users
func (h *Handler) ListUsers(c *gin.Context) {
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

	users, total, err := h.service.ListUsers(c.Request.Context(), limit, offset)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to list users", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	meta := &common.ResponseMeta{
		Pagination: common.NewPaginationMeta(page, limit, total),
	}

	common.Success(c, users, meta)
}

// UpdateUser handles updating a user
func (h *Handler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "User ID is required", nil)
		return
	}

	// Get existing user
	existingUser, err := h.service.GetUserByID(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "User not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get user", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Bind update request
	var req struct {
		FirstName   *string `json:"first_name"`
		LastName    *string `json:"last_name"`
		Phone       *string `json:"phone"`
		Preferences *string `json:"preferences"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Update fields if provided
	if req.FirstName != nil {
		existingUser.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		existingUser.LastName = *req.LastName
	}
	if req.Phone != nil {
		existingUser.Phone = *req.Phone
	}
	if req.Preferences != nil {
		existingUser.Preferences = *req.Preferences
	}

	updatedUser, err := h.service.UpdateUser(c.Request.Context(), existingUser)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to update user", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, updatedUser, nil)
}

// DeleteUser handles deleting a user
func (h *Handler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "User ID is required", nil)
		return
	}

	err := h.service.DeleteUser(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "User not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to delete user", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	c.Status(204) // No Content
}

// ActivateUser handles activating a user
func (h *Handler) ActivateUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "User ID is required", nil)
		return
	}

	user, err := h.service.GetUserByID(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "User not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get user", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	user.IsActive = true
	updatedUser, err := h.service.UpdateUser(c.Request.Context(), user)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to activate user", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, updatedUser, nil)
}

// DeactivateUser handles deactivating a user
func (h *Handler) DeactivateUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, string(common.ErrorCodeInvalidInput), "User ID is required", nil)
		return
	}

	user, err := h.service.GetUserByID(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, string(common.ErrorCodeResourceNotFound), "User not found", nil)
			return
		}
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to get user", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	user.IsActive = false
	updatedUser, err := h.service.UpdateUser(c.Request.Context(), user)
	if err != nil {
		common.InternalServerError(c, string(common.ErrorCodeInternalError), "Failed to deactivate user", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, updatedUser, nil)
}
