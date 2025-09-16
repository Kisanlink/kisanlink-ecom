package routes

import (
	"net/http"

	"kisanlink-ecom/internal/auth"
	"kisanlink-ecom/internal/handlers"
	"kisanlink-ecom/internal/handlers/catalog"
	"kisanlink-ecom/internal/handlers/health"
	"kisanlink-ecom/internal/handlers/integrations"
	"kisanlink-ecom/internal/handlers/inventory"
	"kisanlink-ecom/internal/handlers/orders"
	"kisanlink-ecom/internal/middleware"
	catalogService "kisanlink-ecom/internal/services/catalog"
	integrationService "kisanlink-ecom/internal/services/integrations"
	inventoryService "kisanlink-ecom/internal/services/inventory"
	orderService "kisanlink-ecom/internal/services/orders"
	userService "kisanlink-ecom/internal/services/user"

	scalar "github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// SetupSecureRouter configures all routes with comprehensive authentication and authorization
func SetupSecureRouter(
	aaaClient auth.AAAClient,
	catalogSvc catalogService.CatalogServiceInterface,
	inventorySvc inventoryService.InventoryService,
	orderSvc orderService.OrderServiceInterface,
	userSvc *userService.UserService,
	integrationSvc integrationService.IntegrationServiceInterface,
) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create router
	router := gin.New()

	// Add basic middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(requestid.New())

	// CORS middleware - configurable for production
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*") // TODO: Configure for production
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Initialize authentication and authorization middleware
	authMiddleware := middleware.NewEnhancedAuthMiddleware(aaaClient, nil)
	rbacMiddleware := middleware.NewRBACMiddleware(aaaClient, nil)

	// Health check endpoints (no auth required)
	healthHandler := health.NewHealthHandler()
	metricsHandler := health.NewMetricsHandler()
	statusHandler := health.NewStatusHandler()

	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/health/detailed", healthHandler.DetailedHealthCheck)
	router.GET("/ready", healthHandler.ReadinessCheck)
	router.GET("/metrics", metricsHandler.GetMetrics)
	router.GET("/metrics/prometheus", metricsHandler.PrometheusMetrics)
	router.GET("/status", statusHandler.GetStatusDashboard)

	// Add metrics middleware to track HTTP metrics
	router.Use(metricsHandler.MetricsMiddleware())

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		authGroup := v1.Group("/auth")
		{
			if userSvc != nil {
				authHandler := handlers.NewAuthHandler(userSvc)
				authGroup.POST("/register", authHandler.Register)
				authGroup.POST("/login", authHandler.Login)
				authGroup.POST("/logout",
					authMiddleware.Middleware(),
					authHandler.Logout,
				)
			} else {
				// Fallback handlers when service is not available
				authGroup.POST("/register", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				authGroup.POST("/login", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				authGroup.POST("/logout", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
			}
		}

		// Catalog routes with proper authorization
		catalogGroup := v1.Group("/catalog")
		catalogGroup.Use(authMiddleware.Middleware()) // Require authentication
		{
			if catalogSvc != nil {
				catalogHandler := catalog.NewCatalogHandler(catalogSvc)

				// Generic catalog endpoints
				catalogGroup.GET("",
					rbacMiddleware.RequirePermission("catalog", "read"),
					catalogHandler.ListCatalogItems,
				)
				catalogGroup.GET("/search",
					rbacMiddleware.RequirePermission("catalog", "read"),
					catalogHandler.SearchCatalog,
				)
				catalogGroup.GET("/:type",
					rbacMiddleware.RequirePermission("catalog", "read"),
					catalogHandler.ListCatalogItemsByType,
				)
				catalogGroup.PUT("/:type/:id",
					rbacMiddleware.RequirePermission("catalog", "update"),
					catalogHandler.UpdateCatalogItemByTypeAndID,
				)

				// Products
				products := catalogGroup.Group("/products")
				{
					productHandler := catalog.NewProductHandler(catalogSvc)

					products.POST("",
						rbacMiddleware.RequirePermission("catalog.product", "create"),
						productHandler.CreateProduct,
					)
					products.GET("",
						rbacMiddleware.RequirePermission("catalog.product", "read"),
						productHandler.ListProducts,
					)
					products.GET("/:id",
						rbacMiddleware.RequirePermission("catalog.product", "read"),
						productHandler.GetProductByID,
					)
					products.PUT("/:id",
						rbacMiddleware.RequirePermission("catalog.product", "update"),
						productHandler.UpdateProduct,
					)
					products.DELETE("/:id",
						rbacMiddleware.RequirePermission("catalog.product", "delete"),
						productHandler.DeleteProduct,
					)
				}

				// Services
				services := catalogGroup.Group("/services")
				{
					serviceHandler := catalog.NewServiceHandler(catalogSvc)

					services.POST("",
						rbacMiddleware.RequirePermission("catalog.service", "create"),
						serviceHandler.CreateService,
					)
					services.GET("",
						rbacMiddleware.RequirePermission("catalog.service", "read"),
						serviceHandler.ListServices,
					)
					services.GET("/:id",
						rbacMiddleware.RequirePermission("catalog.service", "read"),
						serviceHandler.GetServiceByID,
					)
					services.PUT("/:id",
						rbacMiddleware.RequirePermission("catalog.service", "update"),
						serviceHandler.UpdateService,
					)
					services.DELETE("/:id",
						rbacMiddleware.RequirePermission("catalog.service", "delete"),
						serviceHandler.DeleteService,
					)
				}

				// Labour
				labour := catalogGroup.Group("/labour")
				{
					labourHandler := catalog.NewLabourHandler(catalogSvc)

					labour.POST("",
						rbacMiddleware.RequirePermission("catalog.labour", "create"),
						labourHandler.CreateLabour,
					)
					labour.GET("",
						rbacMiddleware.RequirePermission("catalog.labour", "read"),
						labourHandler.ListLabour,
					)
					labour.GET("/:id",
						rbacMiddleware.RequirePermission("catalog.labour", "read"),
						labourHandler.GetLabourByID,
					)
					labour.PUT("/:id",
						rbacMiddleware.RequirePermission("catalog.labour", "update"),
						labourHandler.UpdateLabour,
					)
					labour.DELETE("/:id",
						rbacMiddleware.RequirePermission("catalog.labour", "delete"),
						labourHandler.DeleteLabour,
					)
				}
			} else {
				setupFallbackHandlers(catalogGroup)
			}
		}

		// Order routes with proper authorization
		ordersGroup := v1.Group("/orders")
		ordersGroup.Use(authMiddleware.Middleware()) // Require authentication
		{
			if orderSvc != nil {
				orderHandler := orders.NewOrderHandler(orderSvc)

				ordersGroup.POST("",
					rbacMiddleware.RequirePermission("order", "create"),
					orderHandler.CreateOrder,
				)
				ordersGroup.GET("",
					rbacMiddleware.RequirePermission("order", "read"),
					orderHandler.ListOrders,
				)
				ordersGroup.GET("/:id",
					rbacMiddleware.RequirePermission("order", "read"),
					orderHandler.GetOrderByID,
				)
				ordersGroup.PUT("/:id",
					rbacMiddleware.RequirePermission("order", "update"),
					orderHandler.UpdateOrder,
				)
				ordersGroup.PATCH("/:id/status",
					rbacMiddleware.RequirePermission("order", "update"),
					orderHandler.UpdateOrderStatus,
				)
				ordersGroup.POST("/:id/cancel",
					rbacMiddleware.RequirePermission("order", "cancel"),
					orderHandler.CancelOrder,
				)
			} else {
				setupFallbackHandlers(ordersGroup)
			}
		}

		// Inventory routes with proper authorization
		inventoryGroup := v1.Group("/inventory")
		inventoryGroup.Use(authMiddleware.Middleware()) // Require authentication
		{
			if inventorySvc != nil {
				inventoryHandler := inventory.NewInventoryHandler(inventorySvc)

				// Inventory lots management
				lots := inventoryGroup.Group("/lots")
				{
					lots.POST("",
						rbacMiddleware.RequirePermission("inventory", "create"),
						inventoryHandler.CreateInventoryLot,
					)
					lots.GET("",
						rbacMiddleware.RequirePermission("inventory", "read"),
						inventoryHandler.ListInventoryLots,
					)
					lots.GET("/:id",
						rbacMiddleware.RequirePermission("inventory", "read"),
						inventoryHandler.GetInventoryLot,
					)
					lots.PATCH("/:id",
						rbacMiddleware.RequirePermission("inventory", "update"),
						inventoryHandler.UpdateInventoryLot,
					)
					lots.PATCH("/:id/adjust",
						rbacMiddleware.RequirePermission("inventory", "adjust"),
						inventoryHandler.AdjustInventoryQuantity,
					)
					lots.GET("/:id/audit",
						rbacMiddleware.RequirePermission("inventory", "audit"),
						inventoryHandler.GetInventoryAuditTrail,
					)
				}

				// Inventory availability checking
				inventoryGroup.GET("/availability/:catalog_item_id",
					rbacMiddleware.RequirePermission("inventory", "read"),
					inventoryHandler.CheckInventoryAvailability,
				)
			} else {
				setupFallbackHandlers(inventoryGroup)
			}
		}

		// Integration routes with proper authorization
		integrationsGroup := v1.Group("/integrations")
		integrationsGroup.Use(authMiddleware.Middleware()) // Require authentication
		{
			if integrationSvc != nil {
				integrationHandler := integrations.NewIntegrationHandler(integrationSvc)

				// Catalog integration endpoints
				catalogIntegration := integrationsGroup.Group("/catalog")
				{
					catalogIntegration.POST("/proposals",
						rbacMiddleware.RequirePermission("integration.catalog", "propose"),
						integrationHandler.SubmitCatalogProposal,
					)
					catalogIntegration.GET("/exports",
						rbacMiddleware.RequirePermission("integration.catalog", "export"),
						integrationHandler.ExportCatalog,
					)
				}

				// Order integration endpoints
				orderIntegration := integrationsGroup.Group("/orders")
				{
					orderIntegration.POST("/acknowledgements",
						rbacMiddleware.RequirePermission("integration.order", "acknowledge"),
						integrationHandler.AcknowledgeOrder,
					)
				}

				// Webhook validation endpoints (special permission)
				integrationsGroup.POST("/webhooks/validate",
					rbacMiddleware.RequirePermission("integration.webhook", "validate"),
					integrationHandler.ValidateWebhookSignature,
				)

				// Proposal status tracking
				integrationsGroup.GET("/proposals/:proposal_id/status",
					rbacMiddleware.RequirePermission("integration.proposal", "read"),
					integrationHandler.GetProposalStatus,
				)

				// Partner management (admin only)
				integrationsGroup.GET("/partners",
					rbacMiddleware.RequireRole("admin", "integration_admin"),
					integrationHandler.ListIntegrationPartners,
				)
			} else {
				setupFallbackHandlers(integrationsGroup)
			}
		}

		// Admin routes (require admin role)
		adminGroup := v1.Group("/admin")
		adminGroup.Use(authMiddleware.Middleware())
		adminGroup.Use(rbacMiddleware.RequireRole("admin", "super_admin"))
		{
			// Cache management endpoints
			adminGroup.POST("/cache/clear/permissions", func(c *gin.Context) {
				rbacMiddleware.ClearPermissionCache()
				c.JSON(200, gin.H{"message": "Permission cache cleared"})
			})

			adminGroup.POST("/cache/clear/tokens", func(c *gin.Context) {
				authMiddleware.ClearTokenCache()
				c.JSON(200, gin.H{"message": "Token cache cleared"})
			})

			adminGroup.POST("/cache/clear/all", func(c *gin.Context) {
				rbacMiddleware.ClearPermissionCache()
				authMiddleware.ClearTokenCache()
				c.JSON(200, gin.H{"message": "All caches cleared"})
			})
		}
	}

	// API documentation (public access)
	setupAPIDocumentation(router)

	return router
}

// setupFallbackHandlers creates fallback handlers for unavailable services
func setupFallbackHandlers(group *gin.RouterGroup) {
	group.Any("/*path", func(c *gin.Context) {
		c.JSON(503, gin.H{
			"error":   "Service unavailable",
			"message": "The requested service is currently unavailable",
			"path":    c.Request.URL.Path,
		})
	})
}

// setupAPIDocumentation sets up API documentation endpoints
func setupAPIDocumentation(router *gin.Engine) {
	// Swagger documentation using Scalar API Reference
	router.StaticFile("/docs/swagger.json", "docs/swagger.json")
	router.StaticFile("/docs/swagger.yaml", "docs/swagger.yaml")

	router.GET("/docs", func(c *gin.Context) {
		scheme := "http"
		if c.Request.TLS != nil || c.Request.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		specURL := scheme + "://" + c.Request.Host + "/docs/swagger.json"

		htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
			SpecURL:  specURL,
			DarkMode: true,
			CustomOptions: scalar.CustomOptions{
				PageTitle: "KisanLink E-commerce API Reference",
			},
		})
		if err != nil {
			c.String(http.StatusInternalServerError, "failed to render API docs: %v", err)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlContent))
	})

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/docs")
	})
}
