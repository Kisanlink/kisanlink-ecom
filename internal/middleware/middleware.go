package middleware

import (
	"kisanlink-ecom/internal/config"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// Setup configures all middleware for the application.
func Setup(router *gin.Engine, cfg *config.Config) {
	// Recovery middleware
	router.Use(gin.Recovery())

	// Request ID middleware
	router.Use(requestid.New())

	// Logger middleware
	router.Use(Logger())

	// CORS middleware
	router.Use(CORS(cfg.CORS))
}

// CORS configures CORS middleware.
func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	origins := []string{"*"}
	if cfg.AllowedOrigins != "*" && cfg.AllowedOrigins != "" {
		origins = strings.Split(cfg.AllowedOrigins, ",")
	}

	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
	})
}
