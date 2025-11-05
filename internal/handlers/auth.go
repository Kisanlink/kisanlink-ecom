package handlers

import (
	"context"

	authRequests "kisanlink-ecom/entities/requests/auth"
	"kisanlink-ecom/internal/common"
	"kisanlink-ecom/internal/services/user"
	"kisanlink-ecom/internal/utils"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	userService *user.UserService
}

// NewAuthHandler creates a new auth handler instance
func NewAuthHandler(userService *user.UserService) *AuthHandler {
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
// @Success      200      {object}  common.Response{data=map[string]interface{}}  "Login successful"
// @Failure      400      {object}  common.Response{error=common.ResponseError}  "Invalid request"
// @Failure      401      {object}  common.Response{error=common.ResponseError}  "Invalid credentials"
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req authRequests.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		common.BadRequest(c, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	ctx := context.Background()
	user, token, err := h.userService.LoginUser(ctx, req.Username, req.Password)
	if err != nil {
		common.Unauthorized(c, "AUTH_ERROR", "Invalid credentials", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, gin.H{
		"user":  user,
		"token": token,
	}, nil)
}

// Register handles user registration.
// @Summary      User Registration
// @Description  Register a new user account
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      auth.RegisterRequest  true  "User registration data"
// @Success      201      {object}  common.Response{data=map[string]interface{}}  "Registration successful"
// @Failure      400      {object}  common.Response{error=common.ResponseError}  "Invalid request"
// @Failure      409      {object}  common.Response{error=common.ResponseError}  "User already exists"
// @Failure      500      {object}  common.Response{error=common.ResponseError}  "Internal server error"
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req authRequests.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		common.BadRequest(c, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	ctx := context.Background()
	user, err := h.userService.CreateUser(ctx, req)
	if err != nil {
		common.InternalServerError(c, "INTERNAL_ERROR", "Failed to register user", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Created(c, gin.H{
		"user": user,
	}, nil)
}

// Logout handles user logout.
// @Summary      User Logout
// @Description  Logout user and invalidate access token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200      {object}  common.Response{data=string}  "Logout successful"
// @Failure      401      {object}  common.Response{error=common.ResponseError}  "Unauthorized"
// @Router       /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// TODO: Implement token invalidation logic
	common.Success(c, nil, nil)
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
		common.InternalServerError(c, "INTERNAL_ERROR", "Auth handler not initialized", nil)
		return
	}
	defaultAuthHandler.Login(c)
}

// Register is the legacy function that delegates to the handler
func Register(c *gin.Context) {
	if defaultAuthHandler == nil {
		common.InternalServerError(c, "INTERNAL_ERROR", "Auth handler not initialized", nil)
		return
	}
	defaultAuthHandler.Register(c)
}

// Logout is the legacy function that delegates to the handler
func Logout(c *gin.Context) {
	if defaultAuthHandler == nil {
		common.InternalServerError(c, "INTERNAL_ERROR", "Auth handler not initialized", nil)
		return
	}
	defaultAuthHandler.Logout(c)
}
