package routes

import (
	"net/http"

	"kisanlink-ecom/internal/auth"
	"kisanlink-ecom/internal/handlers/catalog"
	"kisanlink-ecom/internal/handlers/orders"
	"kisanlink-ecom/internal/middleware"

	scalar "github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures all routes and middleware
func SetupRouter(
	aaaClient auth.AAAClient,
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

	// Health check endpoint (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "kisanlink-ecom",
		})
	})

	// API v1 routes
	v1 := router.Group("/v1")
	{
		// Catalog routes
		catalogGroup := v1.Group("/catalog")
		{
			// Products
			products := catalogGroup.Group("/products")
			{
				products.POST("",
					conditionalAuthMiddleware(aaaClient),
					// TODO: Add proper authorization middleware
					catalog.NewProductHandler().CreateProduct,
				)
				products.GET("",
					catalog.NewProductHandler().ListProducts,
				)
				products.GET("/:id",
					catalog.NewProductHandler().GetProductByID,
				)
				products.PUT("/:id",
					conditionalAuthMiddleware(aaaClient),
					// TODO: Add proper authorization middleware
					catalog.NewProductHandler().UpdateProduct,
				)
				products.DELETE("/:id",
					conditionalAuthMiddleware(aaaClient),
					// TODO: Add proper authorization middleware
					catalog.NewProductHandler().DeleteProduct,
				)
			}

			// Services
			services := catalogGroup.Group("/services")
			{
				services.POST("",
					conditionalAuthMiddleware(aaaClient),
					// TODO: Add service handler
				)
				services.GET("") // TODO: Add service handler

				services.GET("/:id") // TODO: Add service handler

				services.PUT("/:id",
					conditionalAuthMiddleware(aaaClient),
					// TODO: Add service handler
				)
				services.DELETE("/:id",
					conditionalAuthMiddleware(aaaClient),
					// TODO: Add service handler
				)
			}

			// Labour
			labour := catalogGroup.Group("/labour")
			{
				labour.POST("",
					conditionalAuthMiddleware(aaaClient),
					// TODO: Add labour handler
				)
				labour.GET("") // TODO: Add labour handler

				labour.GET("/:id") // TODO: Add labour handler

				labour.PUT("/:id",
					conditionalAuthMiddleware(aaaClient),
					// TODO: Add labour handler
				)
				labour.DELETE("/:id",
					conditionalAuthMiddleware(aaaClient),
					// TODO: Add labour handler
				)
			}
		}

		// Order routes
		ordersGroup := v1.Group("/orders")
		{
			ordersGroup.POST("",
				conditionalAuthMiddleware(aaaClient),
				// TODO: Add proper authorization middleware
				orders.NewOrderHandler().CreateOrder,
			)
			ordersGroup.GET("",
				conditionalAuthMiddleware(aaaClient),
				orders.NewOrderHandler().ListOrders,
			)
			ordersGroup.GET("/:id",
				conditionalAuthMiddleware(aaaClient),
				orders.NewOrderHandler().GetOrderByID,
			)
			ordersGroup.PATCH("/:id/status",
				conditionalAuthMiddleware(aaaClient),
				// TODO: Add proper authorization middleware
				orders.NewOrderHandler().UpdateOrderStatus,
			)
			ordersGroup.POST("/:id/cancel",
				conditionalAuthMiddleware(aaaClient),
				// TODO: Add proper authorization middleware
				orders.NewOrderHandler().CancelOrder,
			)
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

// conditionalAuthMiddleware returns authentication middleware if aaaClient is available, otherwise a no-op middleware
func conditionalAuthMiddleware(aaaClient auth.AAAClient) gin.HandlerFunc {
	if aaaClient == nil {
		// Return a no-op middleware when AAA client is not available
		return func(c *gin.Context) {
			// Set mock user context for development
			c.Set("subjectID", "mock-user-id")
			c.Set("userRoles", []string{"mock-user"})
			c.Next()
		}
	}

	// Return actual authentication middleware when AAA client is available
	return middleware.AuthNMiddleware(aaaClient)
}
