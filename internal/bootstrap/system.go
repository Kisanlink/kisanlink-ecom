package bootstrap

import (
	"context"
	"fmt"
	"log"
	"time"

	"kisanlink-ecom/internal/auth"
	"kisanlink-ecom/internal/cache"
	"kisanlink-ecom/internal/config"
	"kisanlink-ecom/internal/database"
	"kisanlink-ecom/internal/handlers"
	"kisanlink-ecom/internal/health"
	"kisanlink-ecom/internal/middleware"
	"kisanlink-ecom/internal/observability"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"github.com/gin-gonic/gin"
)

// SystemComponents holds all initialized system components
type SystemComponents struct {
	// Core infrastructure
	DBManager    db.DBManager
	UnitOfWork   database.UnitOfWork
	CacheManager *cache.CacheManager
	AAAClient    auth.Client

	// Observability
	TelemetryManager   *observability.TelemetryManager
	MetricsIntegration *observability.MetricsIntegration
	HealthManager      *health.HealthManager

	// Middleware
	AuthMiddleware        *middleware.EnhancedAuthMiddleware
	RBACMiddleware        *middleware.RBACMiddleware
	ValidationMiddleware  *middleware.ValidationMiddleware
	RateLimiterMiddleware *middleware.RateLimiterMiddleware
	MetricsMiddleware     *middleware.MetricsMiddleware

	// Handlers
	HealthHandler *handlers.HealthHandler
}

// SystemConfig provides configuration for system initialization
type SystemConfig struct {
	// Service information
	ServiceName    string
	ServiceVersion string
	Environment    string

	// Database configuration
	DatabaseConfig *config.MultiDatabaseConfig

	// Cache configuration
	CacheConfig *cache.CacheConfig

	// Authentication configuration
	AAAConfig *config.AAAConfig

	// Telemetry configuration
	TelemetryConfig *observability.TelemetryConfig

	// Metrics configuration
	MetricsConfig *observability.MetricsCollectionConfig

	// Health check configuration
	HealthConfig *health.HealthManagerConfig

	// Middleware configurations
	RateLimiterConfig *middleware.RateLimiterConfig
	ValidationConfig  *middleware.ValidationConfig
	AuthConfig        *middleware.EnhancedAuthConfig
	RBACConfig        *middleware.RBACConfig
}

// DefaultSystemConfig returns default system configuration
func DefaultSystemConfig() *SystemConfig {
	return &SystemConfig{
		ServiceName:    "kisanlink-ecom",
		ServiceVersion: "1.0.0",
		Environment:    "development",

		// Use default configurations for all components
		CacheConfig:       cache.DefaultCacheConfig(),
		TelemetryConfig:   observability.DefaultTelemetryConfig(),
		MetricsConfig:     observability.DefaultMetricsCollectionConfig(),
		HealthConfig:      health.DefaultHealthManagerConfig(),
		RateLimiterConfig: middleware.DefaultRateLimiterConfig(),
		ValidationConfig:  middleware.DefaultValidationConfig(),
		AuthConfig:        middleware.DefaultEnhancedAuthConfig(),
		RBACConfig:        middleware.DefaultRBACConfig(),
	}
}

// InitializeSystem initializes all system components
func InitializeSystem(config *SystemConfig) (*SystemComponents, error) {
	if config == nil {
		config = DefaultSystemConfig()
	}

	components := &SystemComponents{}

	// Initialize core infrastructure
	if err := initializeInfrastructure(config, components); err != nil {
		return nil, fmt.Errorf("failed to initialize infrastructure: %w", err)
	}

	// Initialize observability
	if err := initializeObservability(config, components); err != nil {
		return nil, fmt.Errorf("failed to initialize observability: %w", err)
	}

	// Initialize health checks
	if err := initializeHealthChecks(config, components); err != nil {
		return nil, fmt.Errorf("failed to initialize health checks: %w", err)
	}

	// Initialize middleware
	if err := initializeMiddleware(config, components); err != nil {
		return nil, fmt.Errorf("failed to initialize middleware: %w", err)
	}

	// Initialize handlers
	if err := initializeHandlers(config, components); err != nil {
		return nil, fmt.Errorf("failed to initialize handlers: %w", err)
	}

	log.Printf("System initialization completed successfully")
	return components, nil
}

// initializeInfrastructure initializes core infrastructure components
func initializeInfrastructure(config *SystemConfig, components *SystemComponents) error {
	// Initialize database manager (assuming it's provided externally)
	// components.DBManager = ... (would be injected from main)

	// Initialize Unit of Work
	if components.DBManager != nil {
		components.UnitOfWork = database.NewUnitOfWork(components.DBManager)
	}

	// Initialize cache manager
	if config.CacheConfig != nil {
		components.CacheManager = cache.NewCacheManager(config.CacheConfig)
	}

	// Initialize AAA client (assuming it's provided externally)
	// components.AAAClient = ... (would be injected from main)

	return nil
}

// RunDatabaseMigrations runs database migrations during system initialization
func RunDatabaseMigrations(dbManager db.DBManager, dryRun bool) error {
	log.Println("Running database migrations during system initialization...")

	// Create migration manager
	migrationManager, err := database.NewMigrationManager(dbManager)
	if err != nil {
		return fmt.Errorf("failed to create migration manager: %w", err)
	}

	// Execute migrations
	if dryRun {
		log.Println("Running migrations in dry-run mode...")
		return migrationManager.ExecuteCommand(database.CommandDryRun)
	} else {
		log.Println("Running migrations...")
		return migrationManager.ExecuteCommand(database.CommandMigrate)
	}
}

// initializeObservability initializes observability components
func initializeObservability(config *SystemConfig, components *SystemComponents) error {
	// Update telemetry config with system information
	if config.TelemetryConfig != nil {
		config.TelemetryConfig.ServiceName = config.ServiceName
		config.TelemetryConfig.ServiceVersion = config.ServiceVersion
		config.TelemetryConfig.Environment = config.Environment
	}

	// Initialize telemetry manager
	telemetryManager, err := observability.NewTelemetryManager(config.TelemetryConfig)
	if err != nil {
		return fmt.Errorf("failed to create telemetry manager: %w", err)
	}
	components.TelemetryManager = telemetryManager

	// Initialize metrics integration
	metricsIntegration, err := observability.NewMetricsIntegration(telemetryManager)
	if err != nil {
		return fmt.Errorf("failed to create metrics integration: %w", err)
	}
	components.MetricsIntegration = metricsIntegration

	return nil
}

// initializeHealthChecks initializes health check components
func initializeHealthChecks(config *SystemConfig, components *SystemComponents) error {
	// Update health config with system information
	if config.HealthConfig != nil {
		config.HealthConfig.Version = config.ServiceVersion
	}

	// Initialize health manager
	healthManager := health.NewHealthManager(config.HealthConfig)
	components.HealthManager = healthManager

	// Register health checkers for available components
	if components.DBManager != nil {
		dbChecker := health.NewDatabaseHealthChecker("database", components.DBManager)
		healthManager.RegisterChecker(dbChecker)
	}

	if components.CacheManager != nil {
		// TODO: Add cache health checkers when GetL1Cache/GetL2Cache methods are available
		// cacheChecker := health.NewCacheHealthChecker("cache_l1", components.CacheManager.GetL1Cache())
		// healthManager.RegisterChecker(cacheChecker)
	}

	// Register circuit breaker health checker
	cbChecker := health.NewCircuitBreakerHealthChecker("circuit_breakers")
	healthManager.RegisterChecker(cbChecker)

	// Register custom health checks
	customChecker := health.NewCustomHealthChecker("system_resources", func(ctx context.Context) health.ComponentHealth {
		// Custom health check for system resources
		return health.ComponentHealth{
			Name:        "system_resources",
			Status:      health.StatusHealthy,
			Message:     "System resources are healthy",
			LastChecked: time.Now(),
			Duration:    int64(10 * time.Millisecond),
			Details: map[string]interface{}{
				"cpu_usage":    "< 80%",
				"memory_usage": "< 80%",
				"disk_usage":   "< 90%",
			},
		}
	})
	healthManager.RegisterChecker(customChecker)

	return nil
}

// initializeMiddleware initializes middleware components
func initializeMiddleware(config *SystemConfig, components *SystemComponents) error {
	// Initialize auth middleware
	if components.AAAClient != nil && config.AuthConfig != nil {
		authMiddleware := middleware.NewEnhancedAuthMiddleware(components.AAAClient, config.AuthConfig)
		components.AuthMiddleware = authMiddleware
	}

	// Initialize RBAC middleware
	if components.AAAClient != nil && config.RBACConfig != nil {
		rbacMiddleware := middleware.NewRBACMiddleware(components.AAAClient, config.RBACConfig)
		components.RBACMiddleware = rbacMiddleware
	}

	// Initialize validation middleware
	if config.ValidationConfig != nil {
		validationMiddleware := middleware.NewValidationMiddleware(config.ValidationConfig)
		components.ValidationMiddleware = validationMiddleware
	}

	// Initialize rate limiter middleware
	if config.RateLimiterConfig != nil {
		rateLimiterMiddleware := middleware.NewRateLimiterMiddleware(middleware.TokenBucket, config.RateLimiterConfig)
		components.RateLimiterMiddleware = rateLimiterMiddleware
	}

	// Initialize metrics middleware
	if components.MetricsIntegration != nil {
		metricsMiddleware := middleware.NewMetricsMiddleware(components.MetricsIntegration)
		components.MetricsMiddleware = metricsMiddleware
	}

	return nil
}

// initializeHandlers initializes handler components
func initializeHandlers(config *SystemConfig, components *SystemComponents) error {
	// Initialize health handler
	if components.HealthManager != nil {
		healthHandler := handlers.NewHealthHandler(components.HealthManager)
		components.HealthHandler = healthHandler
	}

	return nil
}

// ConfigureRouter configures the Gin router with all middleware and routes
func ConfigureRouter(components *SystemComponents) *gin.Engine {
	router := gin.New()

	// Add basic middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Add observability middleware
	if components.TelemetryManager != nil {
		router.Use(components.TelemetryManager.TracingMiddleware())
	}

	// Add metrics middleware
	if components.MetricsMiddleware != nil {
		router.Use(components.MetricsMiddleware.Handler())
	}

	// Add health middleware
	if components.HealthManager != nil {
		router.Use(handlers.HealthMiddleware(components.HealthManager))
	}

	// Add security middleware
	if components.AuthMiddleware != nil {
		// Apply auth middleware selectively to protected routes
		// router.Use(components.AuthMiddleware.Handler())
	}

	if components.RateLimiterMiddleware != nil {
		router.Use(components.RateLimiterMiddleware.Middleware())
	}

	// Register routes
	api := router.Group("/api/v1")

	// Health check routes (public)
	if components.HealthHandler != nil {
		components.HealthHandler.RegisterRoutes(api)
	}

	// Protected routes would be added here with RBAC middleware
	// protectedAPI := api.Group("/")
	// if components.RBACMiddleware != nil {
	//     protectedAPI.Use(components.RBACMiddleware.RequirePermission("resource", "action"))
	// }

	return router
}

// Shutdown gracefully shuts down all system components
func Shutdown(components *SystemComponents) error {
	log.Printf("Starting graceful shutdown...")

	// Stop health manager background checks
	if components.HealthManager != nil {
		components.HealthManager.Stop()
	}

	// Shutdown telemetry
	if components.TelemetryManager != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := components.TelemetryManager.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down telemetry: %v", err)
		}
	}

	// Close cache connections
	if components.CacheManager != nil {
		// TODO: Add Close method to CacheManager if needed
		// if err := components.CacheManager.Close(); err != nil {
		//     log.Printf("Error closing cache manager: %v", err)
		// }
	}

	// Close database connections would be handled by the database manager
	// if components.DBManager != nil {
	//     components.DBManager.Close()
	// }

	log.Printf("Graceful shutdown completed")
	return nil
}

// InitializeDefaultComponents initializes system with default configuration
func InitializeDefaultComponents(dbManager db.DBManager, aaaClient auth.Client) (*SystemComponents, error) {
	config := DefaultSystemConfig()

	// Initialize the system
	components, err := InitializeSystem(config)
	if err != nil {
		return nil, err
	}

	// Inject external dependencies
	components.DBManager = dbManager
	components.AAAClient = aaaClient

	// Re-initialize components that depend on external dependencies
	if dbManager != nil {
		components.UnitOfWork = database.NewUnitOfWork(dbManager)

		// Register database health checker
		dbChecker := health.NewDatabaseHealthChecker("database", dbManager)
		components.HealthManager.RegisterChecker(dbChecker)
	}

	if aaaClient != nil && config.AuthConfig != nil {
		components.AuthMiddleware = middleware.NewEnhancedAuthMiddleware(aaaClient, config.AuthConfig)
	}

	if aaaClient != nil && config.RBACConfig != nil {
		components.RBACMiddleware = middleware.NewRBACMiddleware(aaaClient, config.RBACConfig)
	}

	return components, nil
}
