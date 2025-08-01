package server

import (
	"fmt"
	"kisanlink-ecom/internal/config"
	"kisanlink-ecom/internal/handlers"
	"kisanlink-ecom/internal/middleware"
	"kisanlink-ecom/internal/routes"
	"kisanlink-ecom/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Server represents the HTTP server.
type Server struct {
	config           *config.Config
	router           *gin.Engine
	serviceContainer *services.ServiceContainer
}

// New creates a new server instance.
func New(cfg *config.Config) *Server {
	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Create router
	router := gin.New()

	return &Server{
		config: cfg,
		router: router,
	}
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	// Initialize logger
	initLogger(s.config.Logging)

	// Initialize services
	serviceContainer, err := services.NewServiceContainer(s.config)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize services")
	}
	s.serviceContainer = serviceContainer

	// Initialize handlers
	s.initializeHandlers()

	// Setup middleware
	middleware.Setup(s.router, s.config)

	// Setup routes
	routes.Setup(s.router, s.config)

	// Start server
	addr := fmt.Sprintf(":%s", s.config.Server.Port)
	logrus.WithField("port", s.config.Server.Port).Info("Starting server")

	return s.router.Run(addr)
}

// initializeHandlers initializes all handlers with their dependencies
func (s *Server) initializeHandlers() {
	// Initialize user handler
	userHandler := handlers.NewUserHandler(s.serviceContainer.UserService)
	handlers.SetDefaultUserHandler(userHandler)

	// Initialize auth handler
	authHandler := handlers.NewAuthHandler(s.serviceContainer.UserService)
	handlers.SetDefaultAuthHandler(authHandler)

	// Initialize role handler
	roleHandler := handlers.NewRoleHandler(s.serviceContainer.RolePermissionService)
	handlers.SetDefaultRoleHandler(roleHandler)

	// Initialize permission handler
	permissionHandler := handlers.NewPermissionHandler(s.serviceContainer.RolePermissionService)
	handlers.SetDefaultPermissionHandler(permissionHandler)

	logrus.Info("All handlers initialized successfully")
}

// Stop gracefully stops the server and closes connections
func (s *Server) Stop() error {
	if s.serviceContainer != nil {
		return s.serviceContainer.Close()
	}
	return nil
}

// initLogger initializes the logger.
func initLogger(cfg config.LoggingConfig) {
	// Set log level
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}

	logrus.SetLevel(level)

	// Set log format
	if cfg.Format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{})
	}
}
