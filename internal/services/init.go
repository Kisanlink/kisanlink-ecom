package services

import (
	"kisanlink-ecom/internal/config"
	"kisanlink-ecom/internal/database"
	"kisanlink-ecom/internal/repositories/catalog"
)

// ServiceContainer holds all service instances
type ServiceContainer struct {
	UserService           *UserService
	OrderService          *OrderService
	CatalogService        *CatalogService
	RolePermissionService *RolePermissionService
}

// NewServiceContainer creates and initializes all services
func NewServiceContainer(cfg *config.Config) (*ServiceContainer, error) {
	// Initialize database manager
	dbManager, err := database.NewManager(cfg)
	if err != nil {
		return nil, err
	}

	// Initialize repositories
	catalogRepo := catalog.NewCatalogRepository(dbManager)

	// Initialize services
	userService := NewUserService()
	orderService := NewOrderService()
	catalogService := NewCatalogService(catalogRepo)
	rolePermissionService := NewRolePermissionService()

	return &ServiceContainer{
		UserService:           userService,
		OrderService:          orderService,
		CatalogService:        catalogService,
		RolePermissionService: rolePermissionService,
	}, nil
}

// Close closes all service connections
func (sc *ServiceContainer) Close() error {
	// Close database connections
	if dbManager := sc.getDBManager(); dbManager != nil {
		return dbManager.Close()
	}
	return nil
}

// getDBManager returns the database manager from one of the repositories
func (sc *ServiceContainer) getDBManager() *database.Manager {
	// Try to get DB manager from catalog service
	if sc.CatalogService != nil && sc.CatalogService.catalogRepo != nil {
		// This would need to be exposed from the repository
		return nil
	}
	return nil
}
