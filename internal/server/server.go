package server

import (
    "fmt"
    "kisanlink-ecom/internal/config"

    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
)

// Server represents the HTTP server.
type Server struct {
    config *config.Config
    router *gin.Engine
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

    // Setup routes with dependency injection
    // Note: Services are now initialized in main.go and passed to SetupRouter
    // This server just needs to run the router

    // Start server
    addr := fmt.Sprintf(":%s", s.config.Server.Port)
    logrus.WithField("port", s.config.Server.Port).Info("Starting server")

    return s.router.Run(addr)
}

// Stop gracefully stops the server and closes connections
func (s *Server) Stop() error {
    // Clean up resources if needed
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
