package routes

import (
	"net/http"

	"github.com/Kisanlink/kisanlink-ecom/internal/auth"
	"github.com/Kisanlink/kisanlink-ecom/internal/handlers"
	"github.com/Kisanlink/kisanlink-ecom/internal/handlers/actors"
	"github.com/Kisanlink/kisanlink-ecom/internal/handlers/catalog"
	"github.com/Kisanlink/kisanlink-ecom/internal/handlers/collaborator"
	"github.com/Kisanlink/kisanlink-ecom/internal/handlers/discounts"
	"github.com/Kisanlink/kisanlink-ecom/internal/handlers/health"
	"github.com/Kisanlink/kisanlink-ecom/internal/handlers/integrations"
	"github.com/Kisanlink/kisanlink-ecom/internal/handlers/inventory"
	"github.com/Kisanlink/kisanlink-ecom/internal/handlers/orders"
	"github.com/Kisanlink/kisanlink-ecom/internal/handlers/roles"
	"github.com/Kisanlink/kisanlink-ecom/internal/handlers/sla"
	"github.com/Kisanlink/kisanlink-ecom/internal/handlers/taxation"
	"github.com/Kisanlink/kisanlink-ecom/internal/handlers/user"
	"github.com/Kisanlink/kisanlink-ecom/internal/middleware"
	actorsService "github.com/Kisanlink/kisanlink-ecom/internal/services/actors"
	catalogService "github.com/Kisanlink/kisanlink-ecom/internal/services/catalog"
	collaboratorService "github.com/Kisanlink/kisanlink-ecom/internal/services/collaborator"
	discountsService "github.com/Kisanlink/kisanlink-ecom/internal/services/discounts"
	integrationService "github.com/Kisanlink/kisanlink-ecom/internal/services/integrations"
	inventoryService "github.com/Kisanlink/kisanlink-ecom/internal/services/inventory"
	marketplaceService "github.com/Kisanlink/kisanlink-ecom/internal/services/marketplace"
	orderService "github.com/Kisanlink/kisanlink-ecom/internal/services/orders"
	rolesService "github.com/Kisanlink/kisanlink-ecom/internal/services/roles"
	slaService "github.com/Kisanlink/kisanlink-ecom/internal/services/sla"
	taxationService "github.com/Kisanlink/kisanlink-ecom/internal/services/taxation"
	userService "github.com/Kisanlink/kisanlink-ecom/internal/services/user"

	scalar "github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// Services container for all application services
type Services struct {
	CatalogSvc           catalogService.CatalogServiceInterface
	PublishSvc           catalogService.PublishService
	InventorySvc         inventoryService.InventoryService
	AlertSvc             inventoryService.AlertService
	OrderSvc             orderService.OrderServiceInterface
	PaymentScreenshotSvc orderService.PaymentScreenshotServiceInterface
	InvoiceSvc           orderService.InvoiceServiceInterface
	PurchaseOrderSvc     orderService.PurchaseOrderServiceInterface
	UserSvc              *userService.UserService
	IntegrationSvc       integrationService.IntegrationServiceInterface
	MarketplaceSvc       *marketplaceService.MarketplaceServices
	CollaboratorSvc      collaboratorService.CollaboratorServiceInterface
	UserRoleSvc          rolesService.UserRoleServiceInterface
	OrgRoleSvc           rolesService.OrganizationRoleServiceInterface
	EcomRoleSvc          rolesService.EcommerceRoleServiceInterface
	TaxExemptionSvc      taxationService.TaxExemptionServiceInterface
	ServiceSLASvc        slaService.ServiceSLAServiceInterface
	DiscountRuleSvc      discountsService.DiscountRuleServiceInterface
	OrgCollaboratorSvc   actorsService.OrganizationCollaboratorServiceInterface
}

// SetupRouter configures all routes and middleware
func SetupRouter(aaaClient auth.Client, services *Services) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create router
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(requestid.New()) // Use gin-contrib/requestid
	// CORS middleware - allow all origins for now
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

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
			if services.UserSvc != nil {
				authHandler := handlers.NewAuthHandler(services.UserSvc)
				authGroup.POST("/register", authHandler.Register)
				authGroup.POST("/login", authHandler.Login)
				authGroup.POST("/logout",
					conditionalAuthMiddleware(aaaClient),
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

		// User management routes
		usersGroup := v1.Group("/users")
		{
			if services.UserSvc != nil {
				userHandler := user.NewUserHandler(services.UserSvc)
				usersGroup.POST("",
					conditionalAuthMiddleware(aaaClient),
					userHandler.CreateUser,
				)
				usersGroup.GET("", userHandler.ListUsers)
				usersGroup.GET("/:id", userHandler.GetUser)
				usersGroup.GET("/by-username/:username", userHandler.GetUserByUsername)
				usersGroup.GET("/by-email/:email", userHandler.GetUserByEmail)
				usersGroup.PUT("/:id",
					conditionalAuthMiddleware(aaaClient),
					userHandler.UpdateUser,
				)
				usersGroup.DELETE("/:id",
					conditionalAuthMiddleware(aaaClient),
					userHandler.DeleteUser,
				)
				usersGroup.POST("/:id/activate",
					conditionalAuthMiddleware(aaaClient),
					userHandler.ActivateUser,
				)
				usersGroup.POST("/:id/deactivate",
					conditionalAuthMiddleware(aaaClient),
					userHandler.DeactivateUser,
				)
			}
		}

		// User role management routes
		userRolesGroup := v1.Group("/user-roles")
		{
			if services.UserRoleSvc != nil {
				userRoleHandler := roles.NewUserRoleHandler(services.UserRoleSvc)
				userRolesGroup.POST("",
					conditionalAuthMiddleware(aaaClient),
					userRoleHandler.AssignRole,
				)
				userRolesGroup.GET("", userRoleHandler.ListUserRoles)
				userRolesGroup.GET("/:id", userRoleHandler.GetUserRole)
				userRolesGroup.GET("/user/:userId", userRoleHandler.GetUserRoles)
				userRolesGroup.GET("/role/:roleId", userRoleHandler.GetRoleUsers)
				userRolesGroup.PUT("/:id",
					conditionalAuthMiddleware(aaaClient),
					userRoleHandler.UpdateUserRole,
				)
				userRolesGroup.DELETE("/:id",
					conditionalAuthMiddleware(aaaClient),
					userRoleHandler.DeleteUserRole,
				)
				userRolesGroup.POST("/:id/revoke",
					conditionalAuthMiddleware(aaaClient),
					userRoleHandler.RevokeRole,
				)
			}
		}

		// Organization role management routes
		orgRolesGroup := v1.Group("/organization-roles")
		{
			if services.OrgRoleSvc != nil {
				orgRoleHandler := roles.NewOrganizationRoleHandler(services.OrgRoleSvc)
				orgRolesGroup.POST("",
					conditionalAuthMiddleware(aaaClient),
					orgRoleHandler.CreateOrganizationRole,
				)
				orgRolesGroup.GET("", orgRoleHandler.ListOrganizationRoles)
				orgRolesGroup.GET("/:id", orgRoleHandler.GetOrganizationRole)
				orgRolesGroup.GET("/organization/:orgId", orgRoleHandler.GetOrganizationRoles)
				orgRolesGroup.POST("/organization/:orgId/set-default",
					conditionalAuthMiddleware(aaaClient),
					orgRoleHandler.SetDefaultRole,
				)
				orgRolesGroup.PUT("/:id",
					conditionalAuthMiddleware(aaaClient),
					orgRoleHandler.UpdateOrganizationRole,
				)
				orgRolesGroup.DELETE("/:id",
					conditionalAuthMiddleware(aaaClient),
					orgRoleHandler.DeleteOrganizationRole,
				)
			}
		}

		// E-commerce role management routes
		ecomRolesGroup := v1.Group("/ecommerce-roles")
		{
			if services.EcomRoleSvc != nil {
				ecomRoleHandler := roles.NewEcommerceRoleHandler(services.EcomRoleSvc)
				ecomRolesGroup.POST("",
					conditionalAuthMiddleware(aaaClient),
					ecomRoleHandler.CreateEcommerceRole,
				)
				ecomRolesGroup.GET("", ecomRoleHandler.ListEcommerceRoles)
				ecomRolesGroup.GET("/:id", ecomRoleHandler.GetEcommerceRole)
				ecomRolesGroup.GET("/organization/:orgId", ecomRoleHandler.GetEcommerceRolesByOrganization)
				ecomRolesGroup.POST("/check-permission", ecomRoleHandler.CheckPermission)
				ecomRolesGroup.PUT("/:id",
					conditionalAuthMiddleware(aaaClient),
					ecomRoleHandler.UpdateEcommerceRole,
				)
				ecomRolesGroup.DELETE("/:id",
					conditionalAuthMiddleware(aaaClient),
					ecomRoleHandler.DeleteEcommerceRole,
				)
			}
		}

		// Tax exemption routes
		taxExemptionsGroup := v1.Group("/tax-exemptions")
		{
			if services.TaxExemptionSvc != nil {
				taxExemptionHandler := taxation.NewTaxExemptionHandler(services.TaxExemptionSvc)
				taxExemptionsGroup.POST("",
					conditionalAuthMiddleware(aaaClient),
					taxExemptionHandler.CreateTaxExemption,
				)
				taxExemptionsGroup.GET("", taxExemptionHandler.ListTaxExemptions)
				taxExemptionsGroup.GET("/:id", taxExemptionHandler.GetTaxExemption)
				taxExemptionsGroup.GET("/organization/:orgId/valid", taxExemptionHandler.GetValidTaxExemptions)
				taxExemptionsGroup.POST("/calculate", taxExemptionHandler.CalculateTaxExemption)
				taxExemptionsGroup.PUT("/:id",
					conditionalAuthMiddleware(aaaClient),
					taxExemptionHandler.UpdateTaxExemption,
				)
				taxExemptionsGroup.DELETE("/:id",
					conditionalAuthMiddleware(aaaClient),
					taxExemptionHandler.DeleteTaxExemption,
				)
			}
		}

		// Service SLA routes
		serviceSLAsGroup := v1.Group("/service-slas")
		{
			if services.ServiceSLASvc != nil {
				serviceSLAHandler := sla.NewServiceSLAHandler(services.ServiceSLASvc)
				serviceSLAsGroup.POST("",
					conditionalAuthMiddleware(aaaClient),
					serviceSLAHandler.CreateServiceSLA,
				)
				serviceSLAsGroup.GET("", serviceSLAHandler.ListServiceSLAs)
				serviceSLAsGroup.GET("/:id", serviceSLAHandler.GetServiceSLA)
				serviceSLAsGroup.GET("/catalog-item/:catalogItemId", serviceSLAHandler.GetCatalogItemSLAs)
				serviceSLAsGroup.PUT("/:id",
					conditionalAuthMiddleware(aaaClient),
					serviceSLAHandler.UpdateServiceSLA,
				)
				serviceSLAsGroup.DELETE("/:id",
					conditionalAuthMiddleware(aaaClient),
					serviceSLAHandler.DeleteServiceSLA,
				)
			}
		}

		// Discount rule routes
		discountRulesGroup := v1.Group("/discount-rules")
		{
			if services.DiscountRuleSvc != nil {
				discountRuleHandler := discounts.NewDiscountRuleHandler(services.DiscountRuleSvc)
				discountRulesGroup.POST("",
					conditionalAuthMiddleware(aaaClient),
					discountRuleHandler.CreateDiscountRule,
				)
				discountRulesGroup.GET("", discountRuleHandler.ListDiscountRules)
				discountRulesGroup.GET("/:id", discountRuleHandler.GetDiscountRule)
				discountRulesGroup.GET("/organization/:orgId", discountRuleHandler.GetOrganizationRules)
				discountRulesGroup.GET("/organization/:orgId/active", discountRuleHandler.GetActiveRules)
				discountRulesGroup.PUT("/:id",
					conditionalAuthMiddleware(aaaClient),
					discountRuleHandler.UpdateDiscountRule,
				)
				discountRulesGroup.DELETE("/:id",
					conditionalAuthMiddleware(aaaClient),
					discountRuleHandler.DeleteDiscountRule,
				)
			}
		}

		// Organization collaborator routes
		orgCollaboratorsGroup := v1.Group("/organization-collaborators")
		{
			if services.OrgCollaboratorSvc != nil {
				orgCollaboratorHandler := actors.NewOrganizationCollaboratorHandler(services.OrgCollaboratorSvc)
				orgCollaboratorsGroup.POST("",
					conditionalAuthMiddleware(aaaClient),
					orgCollaboratorHandler.CreateOrganizationCollaborator,
				)
				orgCollaboratorsGroup.GET("", orgCollaboratorHandler.ListOrganizationCollaborators)
				orgCollaboratorsGroup.GET("/:id", orgCollaboratorHandler.GetOrganizationCollaborator)
				orgCollaboratorsGroup.GET("/organization/:orgId", orgCollaboratorHandler.GetOrganizationCollaborators)
				orgCollaboratorsGroup.POST("/:id/invite",
					conditionalAuthMiddleware(aaaClient),
					orgCollaboratorHandler.InviteCollaborator,
				)
				orgCollaboratorsGroup.POST("/:id/activate",
					conditionalAuthMiddleware(aaaClient),
					orgCollaboratorHandler.ActivateCollaborator,
				)
				orgCollaboratorsGroup.PUT("/:id",
					conditionalAuthMiddleware(aaaClient),
					orgCollaboratorHandler.UpdateOrganizationCollaborator,
				)
				orgCollaboratorsGroup.DELETE("/:id",
					conditionalAuthMiddleware(aaaClient),
					orgCollaboratorHandler.DeleteOrganizationCollaborator,
				)
			}
		}

		// Catalog routes
		catalogGroup := v1.Group("/catalog")
		{
			if services.CatalogSvc != nil {
				// Create ETag service for cache validation
				etagService := catalogService.NewETagService(nil)

				// Generic catalog handlers
				catalogHandler := catalog.NewCatalogHandler(services.CatalogSvc, etagService)

				// Create cache validation middleware
				cacheMiddleware := middleware.NewCacheValidationMiddleware(nil)

				// Generic catalog endpoints
				catalogGroup.GET("", catalogHandler.ListCatalogItems)
				catalogGroup.GET("/search", catalogHandler.SearchCatalog)
				catalogGroup.GET("/:type", catalogHandler.ListCatalogItemsByType)
				catalogGroup.GET("/:type/:id",
					cacheMiddleware.CacheValidationHandler(),
					catalogHandler.GetCatalogItemByTypeAndID,
				)
				catalogGroup.PUT("/:type/:id",
					conditionalAuthMiddleware(aaaClient),
					catalogHandler.UpdateCatalogItemByTypeAndID,
				)
				catalogGroup.DELETE("/:type/:id",
					conditionalAuthMiddleware(aaaClient),
					catalogHandler.DeleteCatalogItemByTypeAndID,
				)

				// Publish/Unpublish operations
				catalogGroup.POST("/:type/:id/publish",
					conditionalAuthMiddleware(aaaClient),
					catalogHandler.PublishCatalogItem,
				)
				catalogGroup.POST("/:type/:id/unpublish",
					conditionalAuthMiddleware(aaaClient),
					catalogHandler.UnpublishCatalogItem,
				)

				// Price update operations
				catalogGroup.PUT("/:type/:id/price",
					conditionalAuthMiddleware(aaaClient),
					catalogHandler.UpdateCatalogItemPrice,
				)

				// Bulk operations
				bulkGroup := catalogGroup.Group("/bulk")
				{
					bulkGroup.POST("/update",
						conditionalAuthMiddleware(aaaClient),
						catalogHandler.BulkUpdateCatalogItems,
					)
					bulkGroup.POST("/publish",
						conditionalAuthMiddleware(aaaClient),
						catalogHandler.BulkPublishCatalogItems,
					)
					bulkGroup.POST("/price-update",
						conditionalAuthMiddleware(aaaClient),
						catalogHandler.BulkUpdatePrices,
					)
				}

				// Products
				products := catalogGroup.Group("/products")
				{
					productHandler := catalog.NewProductHandler(services.CatalogSvc, etagService)
					products.POST("",
						conditionalAuthMiddleware(aaaClient),
						// TODO: Add proper authorization middleware
						productHandler.CreateProduct,
					)
					products.GET("",
						productHandler.ListProducts,
					)
					products.GET("/:id",
						cacheMiddleware.CacheValidationHandler(),
						productHandler.GetProductByID,
					)
					products.PUT("/:id",
						conditionalAuthMiddleware(aaaClient),
						// TODO: Add proper authorization middleware
						productHandler.UpdateProduct,
					)
					products.DELETE("/:id",
						conditionalAuthMiddleware(aaaClient),
						// TODO: Add proper authorization middleware
						productHandler.DeleteProduct,
					)
					products.PATCH("/:id/activate",
						conditionalAuthMiddleware(aaaClient),
						// TODO: Add proper admin authorization middleware
						productHandler.ActivateProduct,
					)
					products.PATCH("/:id/deactivate",
						conditionalAuthMiddleware(aaaClient),
						// TODO: Add proper admin authorization middleware
						productHandler.DeactivateProduct,
					)

					// FPO Publishing endpoints (requires PublishService)
					if services.PublishSvc != nil {
						fpoPublishHandler := catalog.NewFPOPublishHandler(services.PublishSvc)

						products.POST("/:id/publish",
							conditionalAuthMiddleware(aaaClient),
							// TODO: Add RBAC for Super Admin only
							fpoPublishHandler.PublishProductToFPOs,
						)
						products.GET("/:id/publish-status",
							conditionalAuthMiddleware(aaaClient),
							fpoPublishHandler.GetPublishStatus,
						)
						products.PATCH("/:id/delivery-costs",
							conditionalAuthMiddleware(aaaClient),
							// TODO: Add RBAC for Super Admin only
							fpoPublishHandler.UpdateProductDeliveryCosts,
						)
						products.DELETE("/:id/fpo-access/:fpo_id",
							conditionalAuthMiddleware(aaaClient),
							// TODO: Add RBAC for Super Admin only
							fpoPublishHandler.RevokeFPOAccess,
						)
					}
				}

				// Services
				servicesGroup := catalogGroup.Group("/services")
				{
					serviceHandler := catalog.NewServiceHandler(services.CatalogSvc, etagService)
					servicesGroup.POST("",
						conditionalAuthMiddleware(aaaClient),
						serviceHandler.CreateService,
					)
					servicesGroup.GET("",
						serviceHandler.ListServices,
					)
					servicesGroup.GET("/:id",
						cacheMiddleware.CacheValidationHandler(),
						serviceHandler.GetServiceByID,
					)
					servicesGroup.PUT("/:id",
						conditionalAuthMiddleware(aaaClient),
						serviceHandler.UpdateService,
					)
					servicesGroup.DELETE("/:id",
						conditionalAuthMiddleware(aaaClient),
						serviceHandler.DeleteService,
					)
					servicesGroup.PATCH("/:id/activate",
						conditionalAuthMiddleware(aaaClient),
						// TODO: Add proper admin authorization middleware
						serviceHandler.ActivateService,
					)
					servicesGroup.PATCH("/:id/deactivate",
						conditionalAuthMiddleware(aaaClient),
						// TODO: Add proper admin authorization middleware
						serviceHandler.DeactivateService,
					)
				}

				// Labour
				labour := catalogGroup.Group("/labour")
				{
					labourHandler := catalog.NewLabourHandler(services.CatalogSvc, etagService)
					labour.POST("",
						conditionalAuthMiddleware(aaaClient),
						labourHandler.CreateLabour,
					)
					labour.GET("",
						labourHandler.ListLabour,
					)
					labour.GET("/:id",
						cacheMiddleware.CacheValidationHandler(),
						labourHandler.GetLabourByID,
					)
					labour.PUT("/:id",
						conditionalAuthMiddleware(aaaClient),
						labourHandler.UpdateLabour,
					)
					labour.DELETE("/:id",
						conditionalAuthMiddleware(aaaClient),
						labourHandler.DeleteLabour,
					)
					labour.PATCH("/:id/activate",
						conditionalAuthMiddleware(aaaClient),
						// TODO: Add proper admin authorization middleware
						labourHandler.ActivateLabour,
					)
					labour.PATCH("/:id/deactivate",
						conditionalAuthMiddleware(aaaClient),
						// TODO: Add proper admin authorization middleware
						labourHandler.DeactivateLabour,
					)
				}
			} else {
				// Fallback handlers when service is not available
				catalogGroup.GET("", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				catalogGroup.GET("/search", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				catalogGroup.GET("/:type", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				catalogGroup.PUT("/:type/:id", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})

				// Products fallback
				products := catalogGroup.Group("/products")
				{
					products.POST("", func(c *gin.Context) {
						c.JSON(503, gin.H{"error": "Service unavailable"})
					})
					products.GET("", func(c *gin.Context) {
						c.JSON(503, gin.H{"error": "Service unavailable"})
					})
					products.GET("/:id", func(c *gin.Context) {
						c.JSON(503, gin.H{"error": "Service unavailable"})
					})
					products.PUT("/:id", func(c *gin.Context) {
						c.JSON(503, gin.H{"error": "Service unavailable"})
					})
					products.DELETE("/:id", func(c *gin.Context) {
						c.JSON(503, gin.H{"error": "Service unavailable"})
					})
				}

				// Services fallback
				servicesGroupFallback := catalogGroup.Group("/services")
				{
					servicesGroupFallback.POST("", func(c *gin.Context) {
						c.JSON(503, gin.H{"error": "Service unavailable"})
					})
					servicesGroupFallback.GET("", func(c *gin.Context) {
						c.JSON(503, gin.H{"error": "Service unavailable"})
					})
					servicesGroupFallback.GET("/:id", func(c *gin.Context) {
						c.JSON(503, gin.H{"error": "Service unavailable"})
					})
					servicesGroupFallback.PUT("/:id", func(c *gin.Context) {
						c.JSON(503, gin.H{"error": "Service unavailable"})
					})
					servicesGroupFallback.DELETE("/:id", func(c *gin.Context) {
						c.JSON(503, gin.H{"error": "Service unavailable"})
					})
				}

				// Labour fallback
				labour := catalogGroup.Group("/labour")
				{
					labour.POST("", func(c *gin.Context) {
						c.JSON(503, gin.H{"error": "Service unavailable"})
					})
					labour.GET("", func(c *gin.Context) {
						c.JSON(503, gin.H{"error": "Service unavailable"})
					})
					labour.GET("/:id", func(c *gin.Context) {
						c.JSON(503, gin.H{"error": "Service unavailable"})
					})
					labour.PUT("/:id", func(c *gin.Context) {
						c.JSON(503, gin.H{"error": "Service unavailable"})
					})
					labour.DELETE("/:id", func(c *gin.Context) {
						c.JSON(503, gin.H{"error": "Service unavailable"})
					})
				}
			}
		}

		// Legacy products routes (for backward compatibility)
		productsGroup := v1.Group("/products")
		{
			if services.CatalogSvc != nil {
				etagService := catalogService.NewETagService(nil)
				productHandler := catalog.NewProductHandler(services.CatalogSvc, etagService)
				productsGroup.POST("",
					conditionalAuthMiddleware(aaaClient),
					productHandler.CreateProduct,
				)
				productsGroup.GET("",
					productHandler.ListProducts,
				)
				productsGroup.GET("/:id",
					productHandler.GetProductByID,
				)
				productsGroup.PUT("/:id",
					conditionalAuthMiddleware(aaaClient),
					productHandler.UpdateProduct,
				)
				productsGroup.DELETE("/:id",
					conditionalAuthMiddleware(aaaClient),
					productHandler.DeleteProduct,
				)
			} else {
				// Fallback handlers when service is not available
				productsGroup.POST("", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				productsGroup.GET("", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				productsGroup.GET("/:id", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				productsGroup.PUT("/:id", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				productsGroup.DELETE("/:id", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
			}
		}

		// Order routes
		ordersGroup := v1.Group("/orders")
		{
			if services.OrderSvc != nil {
				orderHandler := orders.NewOrderHandler(services.OrderSvc)
				ordersGroup.POST("",
					conditionalAuthMiddleware(aaaClient),
					// TODO: Add proper authorization middleware
					orderHandler.CreateOrder,
				)
				ordersGroup.GET("",
					conditionalAuthMiddleware(aaaClient),
					orderHandler.ListOrders,
				)
				ordersGroup.GET("/:id",
					conditionalAuthMiddleware(aaaClient),
					orderHandler.GetOrderByID,
				)
				ordersGroup.PUT("/:id",
					conditionalAuthMiddleware(aaaClient),
					// TODO: Add proper authorization middleware
					orderHandler.UpdateOrder,
				)
				ordersGroup.PATCH("/:id/status",
					conditionalAuthMiddleware(aaaClient),
					// TODO: Add proper authorization middleware
					orderHandler.UpdateOrderStatus,
				)
				ordersGroup.POST("/:id/cancel",
					conditionalAuthMiddleware(aaaClient),
					// TODO: Add proper authorization middleware
					orderHandler.CancelOrder,
				)
				ordersGroup.POST("/from-bid",
					conditionalAuthMiddleware(aaaClient),
					// TODO: Add proper authorization middleware
					orderHandler.CreateOrderFromBid,
				)
				ordersGroup.POST("/validate-bid",
					conditionalAuthMiddleware(aaaClient),
					// TODO: Add proper authorization middleware
					orderHandler.ValidateBidForOrder,
				)
				ordersGroup.POST("/:id/payment",
					conditionalAuthMiddleware(aaaClient),
					// TODO: Add proper authorization middleware
					orderHandler.ProcessPaymentForOrder,
				)
				ordersGroup.GET("/:id/payment/status",
					conditionalAuthMiddleware(aaaClient),
					// TODO: Add proper authorization middleware
					orderHandler.GetPaymentStatus,
				)

				// Payment screenshot routes (buyer upload and view)
				if services.PaymentScreenshotSvc != nil {
					paymentScreenshotHandler := orders.NewPaymentScreenshotHandler(services.PaymentScreenshotSvc, services.OrderSvc)
					ordersGroup.POST("/:order_id/payment-screenshots",
						conditionalAuthMiddleware(aaaClient),
						paymentScreenshotHandler.UploadPaymentScreenshot,
					)
					ordersGroup.GET("/:order_id/payment-screenshots",
						conditionalAuthMiddleware(aaaClient),
						paymentScreenshotHandler.ListPaymentScreenshotsByOrder,
					)
				}

				// Invoice routes for orders
				if services.InvoiceSvc != nil {
					invoiceHandler := orders.NewInvoiceHandler(services.InvoiceSvc)
					ordersGroup.POST("/:id/invoice",
						conditionalAuthMiddleware(aaaClient),
						invoiceHandler.GenerateInvoice,
					)
					ordersGroup.GET("/:id/invoice",
						conditionalAuthMiddleware(aaaClient),
						invoiceHandler.GetInvoiceByOrderID,
					)
					ordersGroup.GET("/:id/invoice/pdf",
						conditionalAuthMiddleware(aaaClient),
						invoiceHandler.DownloadInvoicePDF,
					)
				}
			} else {
				// Fallback handlers when service is not available
				ordersGroup.POST("", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				ordersGroup.GET("", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				ordersGroup.GET("/:id", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				ordersGroup.PUT("/:id", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				ordersGroup.PATCH("/:id/status", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				ordersGroup.POST("/:id/cancel", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				ordersGroup.POST("/from-bid", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				ordersGroup.POST("/validate-bid", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				ordersGroup.POST("/:id/payment", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				ordersGroup.GET("/:id/payment/status", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
			}
		}

		// Invoice routes
		invoicesGroup := v1.Group("/invoices")
		{
			if services.InvoiceSvc != nil {
				invoiceHandler := orders.NewInvoiceHandler(services.InvoiceSvc)
				invoicesGroup.GET("",
					conditionalAuthMiddleware(aaaClient),
					invoiceHandler.ListInvoices,
				)
				invoicesGroup.GET("/:id",
					conditionalAuthMiddleware(aaaClient),
					invoiceHandler.GetInvoiceByID,
				)
				invoicesGroup.POST("/:id/finalize",
					conditionalAuthMiddleware(aaaClient),
					invoiceHandler.FinalizeInvoice,
				)
				invoicesGroup.POST("/:id/mark-paid",
					conditionalAuthMiddleware(aaaClient),
					invoiceHandler.MarkInvoiceAsPaid,
				)
				invoicesGroup.POST("/:id/void",
					conditionalAuthMiddleware(aaaClient),
					invoiceHandler.VoidInvoice,
				)
			} else {
				// Fallback handlers when service is not available
				invoicesGroup.GET("", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				invoicesGroup.GET("/:id", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				invoicesGroup.POST("/:id/finalize", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				invoicesGroup.POST("/:id/mark-paid", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				invoicesGroup.POST("/:id/void", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
			}
		}

		// FPO Purchase Order routes
		fpoGroup := v1.Group("/fpo")
		{
			if services.PurchaseOrderSvc != nil {
				poHandler := orders.NewPurchaseOrderHandler(services.PurchaseOrderSvc)
				purchaseOrdersGroup := fpoGroup.Group("/purchase-orders")
				{
					purchaseOrdersGroup.POST("",
						conditionalAuthMiddleware(aaaClient),
						poHandler.CreateManualPO,
					)
					purchaseOrdersGroup.GET("",
						conditionalAuthMiddleware(aaaClient),
						poHandler.ListPurchaseOrders,
					)
					purchaseOrdersGroup.GET("/:id",
						conditionalAuthMiddleware(aaaClient),
						poHandler.GetPurchaseOrderByID,
					)
					purchaseOrdersGroup.PATCH("/:id/status",
						conditionalAuthMiddleware(aaaClient),
						poHandler.UpdatePurchaseOrderStatus,
					)
					purchaseOrdersGroup.GET("/:id/pdf",
						conditionalAuthMiddleware(aaaClient),
						poHandler.DownloadPurchaseOrderPDF,
					)
					purchaseOrdersGroup.POST("/:id/grn",
						conditionalAuthMiddleware(aaaClient),
						poHandler.CreateGRN,
					)
					purchaseOrdersGroup.GET("/:id/grn",
						conditionalAuthMiddleware(aaaClient),
						poHandler.GetGRNsByPO,
					)
				}
			}
		}

		// Admin Dashboard routes
		adminGroup := v1.Group("/admin")
		{
			if services.OrderSvc != nil && services.PurchaseOrderSvc != nil {
				adminDashboardHandler := orders.NewAdminDashboardHandler(services.OrderSvc, services.PurchaseOrderSvc)
				ordersAdminGroup := adminGroup.Group("/orders")
				{
					ordersAdminGroup.GET("/dashboard",
						conditionalAuthMiddleware(aaaClient),
						adminDashboardHandler.GetOrderDashboard,
					)
				}
			}
		}

		// Inventory routes
		inventoryGroup := v1.Group("/inventory")
		{
			if services.InventorySvc != nil {
				inventoryHandler := inventory.NewInventoryHandler(services.InventorySvc)

				// Inventory lots management
				lots := inventoryGroup.Group("/lots")
				{
					lots.POST("",
						conditionalAuthMiddleware(aaaClient),
						// TODO: Add proper authorization middleware
						inventoryHandler.CreateInventoryLot,
					)
					lots.GET("",
						conditionalAuthMiddleware(aaaClient),
						inventoryHandler.ListInventoryLots,
					)
					lots.GET("/:id",
						conditionalAuthMiddleware(aaaClient),
						inventoryHandler.GetInventoryLot,
					)
					lots.PATCH("/:id",
						conditionalAuthMiddleware(aaaClient),
						// TODO: Add proper authorization middleware
						inventoryHandler.UpdateInventoryLot,
					)
					lots.PATCH("/:id/adjust",
						conditionalAuthMiddleware(aaaClient),
						// TODO: Add proper authorization middleware
						inventoryHandler.AdjustInventoryQuantity,
					)
					lots.GET("/:id/audit",
						conditionalAuthMiddleware(aaaClient),
						inventoryHandler.GetInventoryAuditTrail,
					)
				}

				// Inventory availability checking
				inventoryGroup.GET("/availability/:catalog_item_id",
					conditionalAuthMiddleware(aaaClient),
					inventoryHandler.CheckInventoryAvailability,
				)
			} else {
				// Fallback handlers when service is not available
				lots := inventoryGroup.Group("/lots")
				{
					lots.POST("", func(c *gin.Context) {
						c.JSON(503, gin.H{"error": "Service unavailable"})
					})
					lots.GET("", func(c *gin.Context) {
						c.JSON(503, gin.H{"error": "Service unavailable"})
					})
					lots.GET("/:id", func(c *gin.Context) {
						c.JSON(503, gin.H{"error": "Service unavailable"})
					})
					lots.PATCH("/:id", func(c *gin.Context) {
						c.JSON(503, gin.H{"error": "Service unavailable"})
					})
					lots.PATCH("/:id/adjust", func(c *gin.Context) {
						c.JSON(503, gin.H{"error": "Service unavailable"})
					})
					lots.GET("/:id/audit", func(c *gin.Context) {
						c.JSON(503, gin.H{"error": "Service unavailable"})
					})
				}

				inventoryGroup.GET("/availability/:catalog_item_id", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
			}
		}

		// Integration routes
		integrationsGroup := v1.Group("/integrations")
		{
			if services.IntegrationSvc != nil {
				integrationHandler := integrations.NewIntegrationHandler(services.IntegrationSvc)

				// Catalog integration endpoints
				catalogIntegration := integrationsGroup.Group("/catalog")
				{
					catalogIntegration.POST("/proposals",
						conditionalAuthMiddleware(aaaClient),
						integrationHandler.SubmitCatalogProposal,
					)
					catalogIntegration.GET("/exports",
						conditionalAuthMiddleware(aaaClient),
						integrationHandler.ExportCatalog,
					)
				}

				// Order integration endpoints
				orderIntegration := integrationsGroup.Group("/orders")
				{
					orderIntegration.POST("/acknowledgements",
						conditionalAuthMiddleware(aaaClient),
						integrationHandler.AcknowledgeOrder,
					)
				}

				// Webhook validation endpoints
				integrationsGroup.POST("/webhooks/validate",
					integrationHandler.ValidateWebhookSignature,
				)

				// Proposal status tracking
				integrationsGroup.GET("/proposals/:proposal_id/status",
					conditionalAuthMiddleware(aaaClient),
					integrationHandler.GetProposalStatus,
				)

				// Partner management
				integrationsGroup.GET("/partners",
					conditionalAuthMiddleware(aaaClient),
					integrationHandler.ListIntegrationPartners,
				)
			} else {
				// Fallback handlers when service is not available
				catalogIntegration := integrationsGroup.Group("/catalog")
				{
					catalogIntegration.POST("/proposals", func(c *gin.Context) {
						c.JSON(503, gin.H{"error": "Service unavailable"})
					})
					catalogIntegration.GET("/exports", func(c *gin.Context) {
						c.JSON(503, gin.H{"error": "Service unavailable"})
					})
				}

				orderIntegration := integrationsGroup.Group("/orders")
				{
					orderIntegration.POST("/acknowledgements", func(c *gin.Context) {
						c.JSON(503, gin.H{"error": "Service unavailable"})
					})
				}

				integrationsGroup.POST("/webhooks/validate", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})

				integrationsGroup.GET("/proposals/:proposal_id/status", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})

				integrationsGroup.GET("/partners", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
			}
		}

		// Collaborator routes
		collaboratorsGroup := v1.Group("/collaborators")
		{
			if services.CollaboratorSvc != nil {
				collaboratorHandler := collaborator.NewCollaboratorHandler(services.CollaboratorSvc)

				// Public endpoints (no auth required for some operations)
				collaboratorsGroup.GET("/:id", collaboratorHandler.GetCollaboratorByID)
				collaboratorsGroup.GET("/user/:user_id", collaboratorHandler.GetCollaboratorByUserID)

				// Protected endpoints (require authentication)
				collaboratorsGroup.POST("",
					conditionalAuthMiddleware(aaaClient),
					collaboratorHandler.CreateCollaborator,
				)
				collaboratorsGroup.GET("",
					conditionalAuthMiddleware(aaaClient),
					collaboratorHandler.ListCollaborators,
				)
				collaboratorsGroup.GET("/search",
					conditionalAuthMiddleware(aaaClient),
					collaboratorHandler.SearchCollaborators,
				)
				collaboratorsGroup.GET("/stats",
					conditionalAuthMiddleware(aaaClient),
					collaboratorHandler.GetCollaboratorStats,
				)
				collaboratorsGroup.GET("/:id/profile",
					conditionalAuthMiddleware(aaaClient),
					collaboratorHandler.GetCollaboratorProfile,
				)
				collaboratorsGroup.PUT("/:id",
					conditionalAuthMiddleware(aaaClient),
					collaboratorHandler.UpdateCollaborator,
				)
				collaboratorsGroup.DELETE("/:id",
					conditionalAuthMiddleware(aaaClient),
					collaboratorHandler.DeleteCollaborator,
				)
				collaboratorsGroup.PATCH("/:id/status",
					conditionalAuthMiddleware(aaaClient),
					collaboratorHandler.UpdateCollaboratorStatus,
				)
				collaboratorsGroup.POST("/:id/verify",
					conditionalAuthMiddleware(aaaClient),
					collaboratorHandler.VerifyCollaborator,
				)
				collaboratorsGroup.PATCH("/:id/onboarding",
					conditionalAuthMiddleware(aaaClient),
					collaboratorHandler.UpdateOnboardingStep,
				)
				collaboratorsGroup.POST("/:id/onboarding/complete",
					conditionalAuthMiddleware(aaaClient),
					collaboratorHandler.CompleteOnboarding,
				)
				collaboratorsGroup.PATCH("/bulk",
					conditionalAuthMiddleware(aaaClient),
					collaboratorHandler.BulkUpdateCollaborators,
				)
			} else {
				// Fallback handlers when service is not available
				collaboratorsGroup.GET("/:id", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				collaboratorsGroup.GET("/user/:user_id", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				collaboratorsGroup.POST("", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				collaboratorsGroup.GET("", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				collaboratorsGroup.GET("/search", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				collaboratorsGroup.GET("/stats", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				collaboratorsGroup.GET("/:id/profile", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				collaboratorsGroup.PUT("/:id", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				collaboratorsGroup.DELETE("/:id", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				collaboratorsGroup.PATCH("/:id/status", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				collaboratorsGroup.POST("/:id/verify", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				collaboratorsGroup.PATCH("/:id/onboarding", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				collaboratorsGroup.POST("/:id/onboarding/complete", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
				collaboratorsGroup.PATCH("/bulk", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Service unavailable"})
				})
			}
		}

		// Marketplace routes
		SetupMarketplaceRoutesConditional(v1, aaaClient, services.MarketplaceSvc)
	}

	// Swagger documentation using Scalar API Reference (like aaa-service)
	router.StaticFile("/docs/swagger.json", "docs/swagger.json")
	router.StaticFile("/docs/swagger.yaml", "docs/swagger.yaml")
	router.GET("/docs", func(c *gin.Context) {
		scheme := "http"
		if c.Request.TLS != nil || c.Request.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		specURL := scheme + "://" + c.Request.Host + "/docs/swagger.json"

		htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
			SpecURL:        specURL,
			DarkMode:       true,
			Authentication: "BearerAuth",
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

	return router
}

// conditionalAuthMiddleware returns authentication middleware
func conditionalAuthMiddleware(aaaClient auth.Client) gin.HandlerFunc {
	// Use the standard middleware for now
	return middleware.AuthNMiddleware(aaaClient)
}
