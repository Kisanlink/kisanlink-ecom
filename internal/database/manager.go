package database

import (
	"context"
	"fmt"
	"strconv"

	"kisanlink-ecom/internal/config"
	"kisanlink-ecom/internal/repositories"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"go.uber.org/zap"
)

// Manager manages database connections and repositories
type Manager struct {
	dbManager         db.DBManager
	userRepository    *repositories.UserRepository
	productRepository *repositories.ProductRepository
	orderRepository   *repositories.OrderRepository
}

// NewManager creates a new database manager
func NewManager(cfg *config.Config) (*Manager, error) {
	var dbManager db.DBManager

	// Create logger
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Create database config
	dbConfig := &db.Config{
		PrimaryBackend: db.BackendType(cfg.Database.Provider),

		// PostgreSQL configuration
		PostgresHost:     cfg.Database.Host,
		PostgresPort:     strconv.Itoa(cfg.Database.Port),
		PostgresUser:     cfg.Database.User,
		PostgresPassword: cfg.Database.Password,
		PostgresDBName:   cfg.Database.Name,
		PostgresSSLMode:  cfg.Database.SSLMode,

		// DynamoDB configuration
		DynamoDBRegion: cfg.Database.Region,
		DynamoDBTable:  cfg.Database.Name, // Using Name as table name for DynamoDB
	}

	// Choose database backend based on configuration
	switch cfg.Database.Provider {
	case "postgres":
		pgManager := db.NewPostgresManager(dbConfig, logger)
		dbManager = pgManager
	case "dynamodb":
		dynamoManager := db.NewDynamoManager(dbConfig, logger)
		dbManager = dynamoManager
	case "inmemory":
		fallthrough
	default:
		// For development and testing, use in-memory repositories
		return &Manager{
			userRepository:    repositories.NewUserRepository(),
			productRepository: repositories.NewProductRepository(),
			orderRepository:   repositories.NewOrderRepository(),
		}, nil
	}

	return &Manager{
		dbManager:         dbManager,
		userRepository:    repositories.NewUserRepository(),
		productRepository: repositories.NewProductRepository(),
		orderRepository:   repositories.NewOrderRepository(),
	}, nil
}

// Connect establishes database connections.
func (m *Manager) Connect(ctx context.Context) error {
	if m.dbManager != nil {
		return m.dbManager.Connect(ctx)
	}
	return nil
}

// Close closes all database connections.
func (m *Manager) Close() error {
	if m.dbManager != nil {
		return m.dbManager.Close()
	}
	return nil
}

// IsConnected checks if the database is connected.
func (m *Manager) IsConnected() bool {
	if m.dbManager != nil {
		return m.dbManager.IsConnected()
	}
	return true // In-memory repositories are always "connected"
}

// GetUserRepository returns the user repository.
func (m *Manager) GetUserRepository() *repositories.UserRepository {
	return m.userRepository
}

// GetProductRepository returns the product repository.
func (m *Manager) GetProductRepository() *repositories.ProductRepository {
	return m.productRepository
}

// GetOrderRepository returns the order repository.
func (m *Manager) GetOrderRepository() *repositories.OrderRepository {
	return m.orderRepository
}

// Migrate runs database migrations (for SQL databases).
func (m *Manager) Migrate(ctx context.Context) error {
	if m.dbManager == nil {
		return nil // No migration needed for in-memory
	}

	// Check if the manager supports migration
	if migrator, ok := m.dbManager.(interface {
		Migrate(ctx context.Context, models ...interface{}) error
	}); ok {
		// Add your models here for migration
		return migrator.Migrate(ctx /* models would go here */)
	}

	return fmt.Errorf("database manager does not support migration")
}

// Health checks the health of all database connections.
func (m *Manager) Health(_ context.Context) map[string]string {
	result := make(map[string]string)

	if m.dbManager != nil {
		if m.dbManager.IsConnected() {
			result["database"] = "healthy"
		} else {
			result["database"] = "unhealthy"
		}
	} else {
		result["inmemory"] = "healthy"
	}

	return result
}
