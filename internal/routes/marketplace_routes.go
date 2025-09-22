package routes

import (
	"kisanlink-ecom/internal/auth"
	"kisanlink-ecom/internal/handlers/marketplace"
	"kisanlink-ecom/internal/middleware"
	marketplaceService "kisanlink-ecom/internal/services/marketplace"

	"github.com/gin-gonic/gin"
)

// SetupMarketplaceRoutes configures marketplace routes with proper authentication and authorization
func SetupMarketplaceRoutes(
	router *gin.RouterGroup,
	aaaClient auth.AAAClient,
	marketplaceSvc *marketplaceService.MarketplaceServices,
) {
	if marketplaceSvc == nil {
		setupMarketplaceFallbackHandlers(router)
		return
	}

	// Initialize marketplace handler
	marketplaceHandler := marketplace.NewMarketplaceHandler(marketplaceSvc)

	// Initialize middleware
	authMiddleware := middleware.NewEnhancedAuthMiddleware(aaaClient, nil)
	rbacMiddleware := middleware.NewRBACMiddleware(aaaClient, nil)

	// Marketplace routes group
	marketplaceGroup := router.Group("/marketplace")
	marketplaceGroup.Use(authMiddleware.Middleware()) // Require authentication for all marketplace operations
	{
		// Listing management routes
		listingGroup := marketplaceGroup.Group("/listings")
		{
			// Create new listing
			listingGroup.POST("",
				rbacMiddleware.RequirePermission("marketplace.listing", "create"),
				marketplaceHandler.GetListingHandler().CreateListing,
			)

			// Get active listings (public marketplace view)
			listingGroup.GET("",
				rbacMiddleware.RequirePermission("marketplace.listing", "read"),
				marketplaceHandler.GetListingHandler().GetActiveListings,
			)

			// Get specific listing
			listingGroup.GET("/:id",
				rbacMiddleware.RequirePermission("marketplace.listing", "read"),
				marketplaceHandler.GetListingHandler().GetListing,
			)

			// Update listing (seller only)
			listingGroup.PUT("/:id",
				rbacMiddleware.RequirePermission("marketplace.listing", "update"),
				marketplaceHandler.GetListingHandler().UpdateListing,
			)

			// Close listing (seller only)
			listingGroup.POST("/:id/close",
				rbacMiddleware.RequirePermission("marketplace.listing", "close"),
				marketplaceHandler.GetListingHandler().CloseListing,
			)

			// Get my listings (seller view)
			listingGroup.GET("/my-listings",
				rbacMiddleware.RequirePermission("marketplace.listing", "read"),
				marketplaceHandler.GetListingHandler().GetMyListings,
			)

			// Bidding routes for specific listings
			biddingGroup := listingGroup.Group("/:id/bids")
			{
				// Place bid on listing
				biddingGroup.POST("",
					rbacMiddleware.RequirePermission("marketplace.bid", "create"),
					marketplaceHandler.GetBiddingHandler().PlaceBid,
				)

				// Get bids for listing (with visibility filtering)
				biddingGroup.GET("",
					rbacMiddleware.RequirePermission("marketplace.bid", "read"),
					marketplaceHandler.GetBiddingHandler().GetListingBids,
				)
			}
		}

		// Bid management routes
		bidGroup := marketplaceGroup.Group("/bids")
		{
			// Get specific bid
			bidGroup.GET("/:id",
				rbacMiddleware.RequirePermission("marketplace.bid", "read"),
				marketplaceHandler.GetBiddingHandler().GetBid,
			)

			// Get my bids (bidder view)
			bidGroup.GET("/my-bids",
				rbacMiddleware.RequirePermission("marketplace.bid", "read"),
				marketplaceHandler.GetBiddingHandler().GetMyBids,
			)
		}
	}

	// Admin marketplace routes
	adminGroup := router.Group("/admin/v1/marketplace")
	adminGroup.Use(authMiddleware.Middleware())
	adminGroup.Use(rbacMiddleware.RequireRole("admin", "marketplace_admin", "super_admin"))
	{
		// Admin listing management
		adminListingGroup := adminGroup.Group("/listings")
		{
			// Get all listings (admin view - no visibility restrictions)
			adminListingGroup.GET("",
				rbacMiddleware.RequirePermission("marketplace.admin", "read"),
				marketplaceHandler.GetAdminHandler().GetAllListings,
			)

			// Force close listing (admin operation)
			adminListingGroup.POST("/:id/force-close",
				rbacMiddleware.RequirePermission("marketplace.admin", "force_close"),
				marketplaceHandler.GetAdminHandler().ForceCloseListing,
			)
		}

		// Admin bid management
		adminBidGroup := adminGroup.Group("/bids")
		{
			// Remove fraudulent bid (admin operation)
			adminBidGroup.DELETE("/:id",
				rbacMiddleware.RequirePermission("marketplace.admin", "remove_bid"),
				marketplaceHandler.GetAdminHandler().RemoveBid,
			)
		}

		// Marketplace analytics and monitoring
		adminGroup.GET("/stats",
			rbacMiddleware.RequirePermission("marketplace.admin", "analytics"),
			marketplaceHandler.GetAdminHandler().GetMarketplaceStats,
		)

		// Audit log access
		adminGroup.GET("/audit-log",
			rbacMiddleware.RequirePermission("marketplace.admin", "audit"),
			marketplaceHandler.GetAdminHandler().GetAuditLog,
		)
	}

	// Real-time notification routes
	notificationGroup := marketplaceGroup.Group("/notifications")
	{
		// WebSocket endpoint for real-time notifications
		notificationGroup.GET("/ws",
			marketplaceHandler.GetNotificationHandler().HandleWebSocket,
		)

		// Notification preferences management
		notificationGroup.GET("/preferences",
			rbacMiddleware.RequirePermission("marketplace.notification", "read"),
			marketplaceHandler.GetNotificationHandler().GetNotificationPreferences,
		)

		notificationGroup.PUT("/preferences",
			rbacMiddleware.RequirePermission("marketplace.notification", "update"),
			marketplaceHandler.GetNotificationHandler().UpdateNotificationPreferences,
		)

		// Test notification endpoint (for development/testing)
		notificationGroup.POST("/test",
			rbacMiddleware.RequirePermission("marketplace.notification", "test"),
			marketplaceHandler.GetNotificationHandler().TestNotification,
		)
	}
}

// SetupMarketplaceRoutesConditional configures marketplace routes with conditional authentication (for development)
func SetupMarketplaceRoutesConditional(
	router *gin.RouterGroup,
	aaaClient auth.AAAClient,
	marketplaceSvc *marketplaceService.MarketplaceServices,
) {
	if marketplaceSvc == nil {
		setupMarketplaceFallbackHandlers(router)
		return
	}

	// Initialize marketplace handler
	marketplaceHandler := marketplace.NewMarketplaceHandler(marketplaceSvc)

	// Marketplace routes group
	marketplaceGroup := router.Group("/marketplace")
	{
		// Listing management routes
		listingGroup := marketplaceGroup.Group("/listings")
		{
			// Create new listing
			listingGroup.POST("",
				conditionalAuthMiddleware(aaaClient),
				marketplaceHandler.GetListingHandler().CreateListing,
			)

			// Get active listings (public marketplace view)
			listingGroup.GET("",
				conditionalAuthMiddleware(aaaClient),
				marketplaceHandler.GetListingHandler().GetActiveListings,
			)

			// Get specific listing
			listingGroup.GET("/:id",
				conditionalAuthMiddleware(aaaClient),
				marketplaceHandler.GetListingHandler().GetListing,
			)

			// Update listing (seller only)
			listingGroup.PUT("/:id",
				conditionalAuthMiddleware(aaaClient),
				marketplaceHandler.GetListingHandler().UpdateListing,
			)

			// Close listing (seller only)
			listingGroup.POST("/:id/close",
				conditionalAuthMiddleware(aaaClient),
				marketplaceHandler.GetListingHandler().CloseListing,
			)

			// Get my listings (seller view)
			listingGroup.GET("/my-listings",
				conditionalAuthMiddleware(aaaClient),
				marketplaceHandler.GetListingHandler().GetMyListings,
			)

			// Bidding routes for specific listings
			biddingGroup := listingGroup.Group("/:id/bids")
			{
				// Place bid on listing
				biddingGroup.POST("",
					conditionalAuthMiddleware(aaaClient),
					marketplaceHandler.GetBiddingHandler().PlaceBid,
				)

				// Get bids for listing (with visibility filtering)
				biddingGroup.GET("",
					conditionalAuthMiddleware(aaaClient),
					marketplaceHandler.GetBiddingHandler().GetListingBids,
				)
			}
		}

		// Bid management routes
		bidGroup := marketplaceGroup.Group("/bids")
		{
			// Get specific bid
			bidGroup.GET("/:id",
				conditionalAuthMiddleware(aaaClient),
				marketplaceHandler.GetBiddingHandler().GetBid,
			)

			// Get my bids (bidder view)
			bidGroup.GET("/my-bids",
				conditionalAuthMiddleware(aaaClient),
				marketplaceHandler.GetBiddingHandler().GetMyBids,
			)
		}
	}

	// Admin marketplace routes (still require proper authentication)
	adminGroup := router.Group("/admin/v1/marketplace")
	adminGroup.Use(conditionalAuthMiddleware(aaaClient))
	{
		// Admin listing management
		adminListingGroup := adminGroup.Group("/listings")
		{
			// Get all listings (admin view - no visibility restrictions)
			adminListingGroup.GET("",
				marketplaceHandler.GetAdminHandler().GetAllListings,
			)

			// Force close listing (admin operation)
			adminListingGroup.POST("/:id/force-close",
				marketplaceHandler.GetAdminHandler().ForceCloseListing,
			)
		}

		// Admin bid management
		adminBidGroup := adminGroup.Group("/bids")
		{
			// Remove fraudulent bid (admin operation)
			adminBidGroup.DELETE("/:id",
				marketplaceHandler.GetAdminHandler().RemoveBid,
			)
		}

		// Marketplace analytics and monitoring
		adminGroup.GET("/stats",
			marketplaceHandler.GetAdminHandler().GetMarketplaceStats,
		)

		// Audit log access
		adminGroup.GET("/audit-log",
			marketplaceHandler.GetAdminHandler().GetAuditLog,
		)
	}
}

// setupMarketplaceFallbackHandlers creates fallback handlers when marketplace service is unavailable
func setupMarketplaceFallbackHandlers(router *gin.RouterGroup) {
	marketplaceGroup := router.Group("/marketplace")
	{
		// Listing fallback handlers
		listingGroup := marketplaceGroup.Group("/listings")
		{
			listingGroup.POST("", func(c *gin.Context) {
				c.JSON(503, gin.H{
					"error":   "Service unavailable",
					"message": "Marketplace service is currently unavailable",
					"service": "marketplace.listing",
				})
			})

			listingGroup.GET("", func(c *gin.Context) {
				c.JSON(503, gin.H{
					"error":   "Service unavailable",
					"message": "Marketplace service is currently unavailable",
					"service": "marketplace.listing",
				})
			})

			listingGroup.GET("/:id", func(c *gin.Context) {
				c.JSON(503, gin.H{
					"error":   "Service unavailable",
					"message": "Marketplace service is currently unavailable",
					"service": "marketplace.listing",
				})
			})

			listingGroup.PUT("/:id", func(c *gin.Context) {
				c.JSON(503, gin.H{
					"error":   "Service unavailable",
					"message": "Marketplace service is currently unavailable",
					"service": "marketplace.listing",
				})
			})

			listingGroup.POST("/:id/close", func(c *gin.Context) {
				c.JSON(503, gin.H{
					"error":   "Service unavailable",
					"message": "Marketplace service is currently unavailable",
					"service": "marketplace.listing",
				})
			})

			listingGroup.GET("/my-listings", func(c *gin.Context) {
				c.JSON(503, gin.H{
					"error":   "Service unavailable",
					"message": "Marketplace service is currently unavailable",
					"service": "marketplace.listing",
				})
			})

			// Bidding fallback handlers
			biddingGroup := listingGroup.Group("/:id/bids")
			{
				biddingGroup.POST("", func(c *gin.Context) {
					c.JSON(503, gin.H{
						"error":   "Service unavailable",
						"message": "Marketplace service is currently unavailable",
						"service": "marketplace.bidding",
					})
				})

				biddingGroup.GET("", func(c *gin.Context) {
					c.JSON(503, gin.H{
						"error":   "Service unavailable",
						"message": "Marketplace service is currently unavailable",
						"service": "marketplace.bidding",
					})
				})
			}
		}

		// Bid fallback handlers
		bidGroup := marketplaceGroup.Group("/bids")
		{
			bidGroup.GET("/:id", func(c *gin.Context) {
				c.JSON(503, gin.H{
					"error":   "Service unavailable",
					"message": "Marketplace service is currently unavailable",
					"service": "marketplace.bidding",
				})
			})

			bidGroup.GET("/my-bids", func(c *gin.Context) {
				c.JSON(503, gin.H{
					"error":   "Service unavailable",
					"message": "Marketplace service is currently unavailable",
					"service": "marketplace.bidding",
				})
			})
		}
	}

	// Admin fallback handlers
	adminGroup := router.Group("/admin/v1/marketplace")
	{
		adminGroup.Any("/*path", func(c *gin.Context) {
			c.JSON(503, gin.H{
				"error":   "Service unavailable",
				"message": "Marketplace admin service is currently unavailable",
				"service": "marketplace.admin",
				"path":    c.Request.URL.Path,
			})
		})
	}
}
