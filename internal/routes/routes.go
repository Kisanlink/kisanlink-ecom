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

// SetupRouter configures all routes and middleware
func SetupRouter(
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
            if userSvc != nil {
                authHandler := handlers.NewAuthHandler(userSvc)
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

        // RBAC routes are now handled by AAA service via gRPC
        // Roles and permissions are managed centrally in the AAA service

        // Catalog routes
        catalogGroup := v1.Group("/catalog")
        {
            if catalogSvc != nil {
                // Generic catalog handlers
                catalogHandler := catalog.NewCatalogHandler(catalogSvc)

                // Generic catalog endpoints
                catalogGroup.GET("", catalogHandler.ListCatalogItems)
                catalogGroup.GET("/search", catalogHandler.SearchCatalog)
                catalogGroup.GET("/:type", catalogHandler.ListCatalogItemsByType)
                catalogGroup.PUT("/:type/:id",
                    conditionalAuthMiddleware(aaaClient),
                    catalogHandler.UpdateCatalogItemByTypeAndID,
                )

                // Products
                products := catalogGroup.Group("/products")
                {
                    productHandler := catalog.NewProductHandler(catalogSvc)
                    products.POST("",
                        conditionalAuthMiddleware(aaaClient),
                        // TODO: Add proper authorization middleware
                        productHandler.CreateProduct,
                    )
                    products.GET("",
                        productHandler.ListProducts,
                    )
                    products.GET("/:id",
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
                }

                // Services
                services := catalogGroup.Group("/services")
                {
                    serviceHandler := catalog.NewServiceHandler(catalogSvc)
                    services.POST("",
                        conditionalAuthMiddleware(aaaClient),
                        serviceHandler.CreateService,
                    )
                    services.GET("",
                        serviceHandler.ListServices,
                    )
                    services.GET("/:id",
                        serviceHandler.GetServiceByID,
                    )
                    services.PUT("/:id",
                        conditionalAuthMiddleware(aaaClient),
                        serviceHandler.UpdateService,
                    )
                    services.DELETE("/:id",
                        conditionalAuthMiddleware(aaaClient),
                        serviceHandler.DeleteService,
                    )
                }

                // Labour
                labour := catalogGroup.Group("/labour")
                {
                    labourHandler := catalog.NewLabourHandler(catalogSvc)
                    labour.POST("",
                        conditionalAuthMiddleware(aaaClient),
                        labourHandler.CreateLabour,
                    )
                    labour.GET("",
                        labourHandler.ListLabour,
                    )
                    labour.GET("/:id",
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
                services := catalogGroup.Group("/services")
                {
                    services.POST("", func(c *gin.Context) {
                        c.JSON(503, gin.H{"error": "Service unavailable"})
                    })
                    services.GET("", func(c *gin.Context) {
                        c.JSON(503, gin.H{"error": "Service unavailable"})
                    })
                    services.GET("/:id", func(c *gin.Context) {
                        c.JSON(503, gin.H{"error": "Service unavailable"})
                    })
                    services.PUT("/:id", func(c *gin.Context) {
                        c.JSON(503, gin.H{"error": "Service unavailable"})
                    })
                    services.DELETE("/:id", func(c *gin.Context) {
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
            if catalogSvc != nil {
                productHandler := catalog.NewProductHandler(catalogSvc)
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
            if orderSvc != nil {
                orderHandler := orders.NewOrderHandler(orderSvc)
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
            }
        }

        // Inventory routes
        inventoryGroup := v1.Group("/inventory")
        {
            if inventorySvc != nil {
                inventoryHandler := inventory.NewInventoryHandler(inventorySvc)

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
            if integrationSvc != nil {
                integrationHandler := integrations.NewIntegrationHandler(integrationSvc)

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

    return router
}

// conditionalAuthMiddleware returns authentication middleware
func conditionalAuthMiddleware(aaaClient auth.AAAClient) gin.HandlerFunc {
    // Use the standard middleware for now
    return middleware.AuthNMiddleware(aaaClient)
}
