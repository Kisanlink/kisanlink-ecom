package handlers

import (
	"context"
	"net/http"

	"kisanlink-ecom/internal/models/auth"
	"kisanlink-ecom/internal/models/user"
	"kisanlink-ecom/internal/services"
	"kisanlink-ecom/internal/utils"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	userService *services.UserService
}

// NewAuthHandler creates a new auth handler instance
func NewAuthHandler(userService *services.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// Login handles user login.
// @Summary      User Login
// @Description  Authenticate user and return access token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      auth.LoginRequest  true  "Login credentials"
// @Success      200      {object}  object  "Login successful"
// @Failure      400      {object}  object  "Invalid request"
// @Failure      401      {object}  object  "Invalid credentials"
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	ctx := context.Background()
	user, token, err := h.userService.LoginUser(ctx, req.Username, req.Password)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "AUTH_ERROR", "Invalid credentials", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", gin.H{
		"user":  user,
		"token": token,
	})
}

// Register handles user registration.
// @Summary      User Registration
// @Description  Register a new user account
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      user.CreateUserRequest  true  "User registration data"
// @Success      201      {object}  object  "Registration successful"
// @Failure      400      {object}  object  "Invalid request"
// @Failure      409      {object}  object  "User already exists"
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
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
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to register user", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Registration successful", gin.H{
		"user": user,
	})
}

// Logout handles user logout.
// @Summary      User Logout
// @Description  Logout user and invalidate access token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200      {object}  object  "Logout successful"
// @Failure      401      {object}  object  "Unauthorized"
// @Router       /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// TODO: Implement token invalidation logic
	utils.SuccessResponse(c, http.StatusOK, "Logout successful", nil)
}

// Legacy handler functions for backward compatibility
var defaultAuthHandler *AuthHandler

// SetDefaultAuthHandler sets the default auth handler
func SetDefaultAuthHandler(handler *AuthHandler) {
	defaultAuthHandler = handler
}

// Login is the legacy function that delegates to the handler
func Login(c *gin.Context) {
	if defaultAuthHandler == nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Auth handler not initialized", "")
		return
	}
	defaultAuthHandler.Login(c)
}

// Register is the legacy function that delegates to the handler
func Register(c *gin.Context) {
	if defaultAuthHandler == nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Auth handler not initialized", "")
		return
	}
	defaultAuthHandler.Register(c)
}

// Logout is the legacy function that delegates to the handler
func Logout(c *gin.Context) {
	if defaultAuthHandler == nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Auth handler not initialized", "")
		return
	}
	defaultAuthHandler.Logout(c)
}
