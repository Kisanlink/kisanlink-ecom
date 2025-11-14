package services

import (
	"github.com/Kisanlink/kisanlink-ecom/internal/auth"
	"github.com/Kisanlink/kisanlink-ecom/internal/config"
	"github.com/Kisanlink/kisanlink-ecom/internal/database"
	catalogRepo "github.com/Kisanlink/kisanlink-ecom/internal/repositories/catalog"
	inventoryRepo "github.com/Kisanlink/kisanlink-ecom/internal/repositories/inventory"
	orderRepo "github.com/Kisanlink/kisanlink-ecom/internal/repositories/orders"
	sequenceRepo "github.com/Kisanlink/kisanlink-ecom/internal/repositories/sequence"
	userRepo "github.com/Kisanlink/kisanlink-ecom/internal/repositories/user"
	"github.com/Kisanlink/kisanlink-ecom/internal/services/catalog"
	"github.com/Kisanlink/kisanlink-ecom/internal/services/inventory"
	"github.com/Kisanlink/kisanlink-ecom/internal/services/orders"
	"github.com/Kisanlink/kisanlink-ecom/internal/services/sequence"
	"github.com/Kisanlink/kisanlink-ecom/internal/services/user"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// RolePermissionService handles role and permission operations
type RolePermissionService struct {
	// Dependencies will be added as role and permission features are implemented
}

// NewRolePermissionService creates a new role permission service
func NewRolePermissionService() *RolePermissionService {
	return &RolePermissionService{}
}

// ServiceContainer holds all service instances
type ServiceContainer struct {
	UserService           *user.UserService
	OrderService          *orders.OrderService
	CatalogService        *catalog.CatalogService
	InventoryService      inventory.InventoryService
	SequenceService       *sequence.SequenceService
	RolePermissionService *RolePermissionService
}

// NewServiceContainer creates and initializes all services
func NewServiceContainer(cfg *config.Config) (*ServiceContainer, error) {
	// Convert DatabaseConfig to MultiDatabaseConfig
	multiDBConfig := config.MultiDatabaseConfig{
		PostgreSQL: config.PostgreSQLConfig{
			Host:         cfg.Database.Host,
			Port:         cfg.Database.Port,
			Database:     cfg.Database.Name,
			Username:     cfg.Database.User,
			Password:     cfg.Database.Password,
			SSLMode:      cfg.Database.SSLMode,
			MaxOpenConns: cfg.Database.MaxConns,
			MaxIdleConns: cfg.Database.MaxIdleTime,
		},
		DynamoDB: config.DynamoDBConfig{
			Region:          cfg.Database.Region,
			AccessKeyID:     cfg.Database.AccessKey,
			SecretAccessKey: cfg.Database.SecretKey,
		},
	}

	// Initialize database manager
	dbManager, err := database.NewDatabaseManager(multiDBConfig)
	if err != nil {
		return nil, err
	}

	// Initialize repositories
	catalogRepository := catalogRepo.NewCatalogRepository(dbManager.GetManager(db.BackendGorm))
	inventoryRepository := inventoryRepo.NewInventoryRepository(dbManager.GetManager(db.BackendGorm))
	orderRepository := orderRepo.NewOrderRepository(dbManager.GetManager(db.BackendGorm))
	sequenceRepository := sequenceRepo.NewSequenceRepository(dbManager.GetManager(db.BackendGorm))
	userRepository := userRepo.NewUserRepository(dbManager.GetManager(db.BackendGorm))

	// Wire repository dependencies
	orderRepository.SetCatalogRepository(catalogRepository)
	orderRepository.SetInventoryRepository(inventoryRepository)

	// Initialize AAA client
	aaaClient, err := auth.NewClient(&cfg.AAA)
	if err != nil {
		// Log warning but continue without AAA client
		aaaClient = nil
	}

	// Initialize services
	userSvc := user.NewUserService(aaaClient, userRepository)
	catalogSvc := catalog.NewCatalogService(catalogRepository)
	inventorySvc := inventory.NewInventoryService(inventoryRepository, catalogRepository)
	sequenceSvc := sequence.NewSequenceService(sequenceRepository)
	orderSvc := orders.NewOrderService(orderRepository, catalogSvc, inventorySvc, sequenceSvc)
	rolePermissionService := NewRolePermissionService()

	return &ServiceContainer{
		UserService:           userSvc,
		OrderService:          orderSvc,
		CatalogService:        catalogSvc,
		InventoryService:      inventorySvc,
		SequenceService:       sequenceSvc,
		RolePermissionService: rolePermissionService,
	}, nil
}

// Close closes all service connections
func (sc *ServiceContainer) Close() error {
	// Close AAA client if available
	// AAA client will be added to service container when centralized service management is implemented
	return nil
}
