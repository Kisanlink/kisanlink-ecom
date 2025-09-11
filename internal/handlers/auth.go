package handlers

import (
    "context"
    "net/http"

    "kisanlink-ecom/entities/requests/auth"
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
// @Success      200      {object}  common.Response  "Login successful"
// @Failure      400      {object}  common.Response{error=common.ResponseError}  "Invalid request"
// @Failure      401      {object}  common.Response{error=common.ResponseError}  "Invalid credentials"
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
// @Param        request  body      auth.RegisterRequest  true  "User registration data"
// @Success      201      {object}  common.Response  "Registration successful"
// @Failure      400      {object}  common.Response{error=common.ResponseError}  "Invalid request"
// @Failure      409      {object}  common.Response{error=common.ResponseError}  "User already exists"
// @Failure      500      {object}  common.Response{error=common.ResponseError}  "Internal server error"
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
    var req auth.RegisterRequest
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
// @Success      200      {object}  common.Response  "Logout successful"
// @Failure      401      {object}  common.Response{error=common.ResponseError}  "Unauthorized"
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
