package routes

import (
	"kisanlink-ecom/internal/config"
	"kisanlink-ecom/internal/handlers"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Setup configures all routes for the application.
func Setup(router *gin.Engine, cfg *config.Config) {
	// Health check endpoint
	router.GET("/health", handlers.HealthCheck)

	// Swagger documentation endpoints
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/docs", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	api := router.Group("/api/" + cfg.Server.Version)
	{
		setupAuthRoutes(api)
		setupUserRoutes(api)
		setupRoleRoutes(api)
		setupPermissionRoutes(api)
		setupProductRoutes(api)
		setupOrderRoutes(api)
	}
}

// setupAuthRoutes configures authentication routes.
func setupAuthRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		auth.POST("/login", handlers.Login)
		auth.POST("/register", handlers.Register)
		auth.POST("/logout", handlers.Logout)
	}
}

// setupUserRoutes configures user routes.
func setupUserRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	{
		users.GET("", handlers.GetUsers)
		users.POST("", handlers.CreateUser)
		users.GET("/:id", handlers.GetUser)
		users.PUT("/:id", handlers.UpdateUser)
		users.DELETE("/:id", handlers.DeleteUser)
	}
}

// setupRoleRoutes configures role routes.
func setupRoleRoutes(rg *gin.RouterGroup) {
	roles := rg.Group("/roles")
	{
		roles.GET("", handlers.GetRoles)
		roles.POST("", handlers.CreateRole)
		roles.GET("/:id", handlers.GetRole)
		roles.PUT("/:id", handlers.UpdateRole)
		roles.DELETE("/:id", handlers.DeleteRole)
	}
}

// setupPermissionRoutes configures permission routes.
func setupPermissionRoutes(rg *gin.RouterGroup) {
	permissions := rg.Group("/permissions")
	{
		permissions.GET("", handlers.GetPermissions)
		permissions.POST("", handlers.CreatePermission)
		permissions.GET("/:id", handlers.GetPermission)
		permissions.PUT("/:id", handlers.UpdatePermission)
		permissions.DELETE("/:id", handlers.DeletePermission)
	}
}

// setupProductRoutes configures product routes.
func setupProductRoutes(rg *gin.RouterGroup) {
	products := rg.Group("/products")
	{
		products.GET("", handlers.GetProducts)
		products.POST("", handlers.CreateProduct)
		products.GET("/:id", handlers.GetProduct)
		products.PUT("/:id", handlers.UpdateProduct)
		products.DELETE("/:id", handlers.DeleteProduct)
	}
}

// setupOrderRoutes configures order routes.
func setupOrderRoutes(rg *gin.RouterGroup) {
	orders := rg.Group("/orders")
	{
		orders.GET("", handlers.GetOrders)
		orders.POST("", handlers.CreateOrder)
		orders.GET("/:id", handlers.GetOrder)
		orders.PUT("/:id", handlers.UpdateOrder)
		orders.DELETE("/:id", handlers.DeleteOrder)
	}
}
