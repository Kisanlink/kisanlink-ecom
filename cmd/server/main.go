package main

import (
	"kisanlink-ecom/docs"
	"kisanlink-ecom/internal/config"
	"kisanlink-ecom/internal/server"
	"log"
)

// @title           KisanLink E-commerce API
// @version         1.0
// @description     A comprehensive e-commerce API for agricultural products and services.
// @termsOfService  http://swagger.io/terms/

// @contact.name   KisanLink API Support
// @contact.url    http://www.kisanlink.com/support
// @contact.email  support@kisanlink.com

// @license.name  MIT
// @license.url   http://www.opensource.org/licenses/mit-license.php

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Initialize Swagger docs
	docs.SwaggerInfo.Title = "KisanLink E-commerce API"
	docs.SwaggerInfo.Description = "A comprehensive e-commerce API for agricultural products and services."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create and start server
	srv := server.New(cfg)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
