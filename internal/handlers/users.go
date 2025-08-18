package handlers

import (
	"context"
	"net/http"

	"kisanlink-ecom/internal/models/user"
	"kisanlink-ecom/internal/services"
	"kisanlink-ecom/internal/utils"

	"github.com/gin-gonic/gin"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler creates a new user handler instance
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetUsers handles getting all users.
// @Summary      Get All Users
// @Description  Retrieve a list of all users with pagination
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page     query     int  false  "Page number"     minimum(1)
// @Param        per_page query     int  false  "Items per page"  minimum(1) maximum(100)
// @Success      200      {object}  common.UsersSuccessResponse  "Users retrieved successfully"
// @Failure      401      {object}  common.ErrorResponseModel  "Unauthorized"
// @Failure      403      {object}  common.ErrorResponseModel  "Forbidden"
// @Failure      500      {object}  common.ErrorResponseModel  "Internal server error"
// @Router       /api/v1/users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
	ctx := context.Background()

	users, err := h.userService.GetAllUsers(ctx)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve users", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Users retrieved successfully", gin.H{
		"users": users,
	})
}

// CreateUser handles user creation.
// @Summary      Create User
// @Description  Create a new user account
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      user.CreateUserRequest  true  "User data"
// @Success      201      {object}  common.UserSuccessResponse  "User created successfully"
// @Failure      400      {object}  common.ErrorResponseModel  "Invalid request"
// @Failure      401      {object}  common.ErrorResponseModel  "Unauthorized"
// @Failure      403      {object}  common.ErrorResponseModel  "Forbidden"
// @Failure      409      {object}  common.ErrorResponseModel  "User already exists"
// @Failure      500      {object}  common.ErrorResponseModel  "Internal server error"
// @Router       /api/v1/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req user.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	ctx := context.Background()
	user, err := h.userService.CreateUser(ctx, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create user", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "User created successfully", gin.H{
		"user": user,
	})
}

// GetUser handles getting a specific user by ID.
// @Summary      Get User by ID
// @Description  Retrieve a specific user by their ID
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  common.UserSuccessResponse  "User retrieved successfully"
// @Failure      400  {object}  common.ErrorResponseModel  "Invalid user ID"
// @Failure      401  {object}  common.ErrorResponseModel  "Unauthorized"
// @Failure      403  {object}  common.ErrorResponseModel  "Forbidden"
// @Failure      404  {object}  common.ErrorResponseModel  "User not found"
// @Failure      500  {object}  common.ErrorResponseModel  "Internal server error"
// @Router       /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		utils.ValidationErrorResponse(c, "User ID is required")
		return
	}

	ctx := context.Background()
	user, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve user", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User retrieved successfully", gin.H{
		"user": user,
	})
}

// UpdateUser handles user updates.
// @Summary      Update User
// @Description  Update an existing user's information
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      string                true   "User ID"
// @Param        request body      user.UpdateUserRequest  true  "User update data"
// @Success      200      {object}  common.UserSuccessResponse  "User updated successfully"
// @Failure      400      {object}  common.ErrorResponseModel  "Invalid request"
// @Failure      401      {object}  common.ErrorResponseModel  "Unauthorized"
// @Failure      403      {object}  common.ErrorResponseModel  "Forbidden"
// @Failure      404      {object}  common.ErrorResponseModel  "User not found"
// @Failure      500      {object}  common.ErrorResponseModel  "Internal server error"
// @Router       /api/v1/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		utils.ValidationErrorResponse(c, "User ID is required")
		return
	}

	var req user.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	ctx := context.Background()
	user, err := h.userService.UpdateUser(ctx, userID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update user", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User updated successfully", gin.H{
		"user": user,
	})
}

// DeleteUser handles user deletion.
// @Summary      Delete User
// @Description  Delete a user account
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  common.SuccessResponse  "User deleted successfully"
// @Failure      400  {object}  common.ErrorResponseModel  "Invalid user ID"
// @Failure      401  {object}  common.ErrorResponseModel  "Unauthorized"
// @Failure      403  {object}  common.ErrorResponseModel  "Forbidden"
// @Failure      404  {object}  common.ErrorResponseModel  "User not found"
// @Failure      500  {object}  common.ErrorResponseModel  "Internal server error"
// @Router       /api/v1/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		utils.ValidationErrorResponse(c, "User ID is required")
		return
	}

	ctx := context.Background()
	err := h.userService.DeleteUser(ctx, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete user", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User deleted successfully", nil)
}

// Legacy handler functions for backward compatibility
var defaultUserHandler *UserHandler

// SetDefaultUserHandler sets the default user handler
func SetDefaultUserHandler(handler *UserHandler) {
	defaultUserHandler = handler
}

// GetUsers is the legacy function that delegates to the handler
func GetUsers(c *gin.Context) {
	if defaultUserHandler == nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "User handler not initialized", "")
		return
	}
	defaultUserHandler.GetUsers(c)
}

// CreateUser is the legacy function that delegates to the handler
func CreateUser(c *gin.Context) {
	if defaultUserHandler == nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "User handler not initialized", "")
		return
	}
	defaultUserHandler.CreateUser(c)
}

// GetUser is the legacy function that delegates to the handler
func GetUser(c *gin.Context) {
	if defaultUserHandler == nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "User handler not initialized", "")
		return
	}
	defaultUserHandler.GetUser(c)
}

// UpdateUser is the legacy function that delegates to the handler
func UpdateUser(c *gin.Context) {
	if defaultUserHandler == nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "User handler not initialized", "")
		return
	}
	defaultUserHandler.UpdateUser(c)
}

// DeleteUser is the legacy function that delegates to the handler
func DeleteUser(c *gin.Context) {
	if defaultUserHandler == nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "User handler not initialized", "")
		return
	}
	defaultUserHandler.DeleteUser(c)
}
