package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Kisanlink/kisanlink-ecom/internal/config"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// DatabaseManager wraps the kisanlink-db manager with health checking
type DatabaseManager struct {
	manager *db.DatabaseManager
	config  config.MultiDatabaseConfig
}

// NewDatabaseManager creates a new database manager with proper error handling
func NewDatabaseManager(cfg config.MultiDatabaseConfig) (*DatabaseManager, error) {
	log.Printf("Loading database configuration...")

	// Validate configuration
	if err := validateDatabaseConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid database configuration: %w", err)
	}

	// Create database config
	dbConfig := &db.Config{
		PrimaryBackend: db.BackendGorm,

		// PostgreSQL configuration
		PostgresHost:         cfg.PostgreSQL.Host,
		PostgresPort:         fmt.Sprintf("%d", cfg.PostgreSQL.Port),
		PostgresUser:         cfg.PostgreSQL.Username,
		PostgresPassword:     cfg.PostgreSQL.Password,
		PostgresDBName:       cfg.PostgreSQL.Database,
		PostgresSSLMode:      cfg.PostgreSQL.SSLMode,
		PostgresMaxConns:     cfg.PostgreSQL.MaxOpenConns,
		PostgresIdleConns:    cfg.PostgreSQL.MaxIdleConns,
		PostgresReadReplicas: []string{},

		// DynamoDB configuration
		DynamoDBRegion: cfg.DynamoDB.Region,
		DynamoDBTable:  cfg.DynamoDB.Table,
		LogLevel:       "info",
	}

	log.Printf("Database config: PostgreSQL=%s:%d/%s, DynamoDB=%s/%s",
		dbConfig.PostgresHost, cfg.PostgreSQL.Port, dbConfig.PostgresDBName,
		dbConfig.DynamoDBRegion, dbConfig.DynamoDBTable)

	// Create database manager
	dbManager := db.NewDatabaseManagerWithConfig(dbConfig)

	// Connect to database with context and timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := dbManager.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Printf("Database manager initialized successfully")

	wrapper := &DatabaseManager{
		manager: dbManager,
		config:  cfg,
	}

	return wrapper, nil
}

// validateDatabaseConfig validates the database configuration
func validateDatabaseConfig(cfg config.MultiDatabaseConfig) error {
	// Validate PostgreSQL config
	if cfg.PostgreSQL.Host == "" {
		return fmt.Errorf("PostgreSQL host is required")
	}
	if cfg.PostgreSQL.Port <= 0 || cfg.PostgreSQL.Port > 65535 {
		return fmt.Errorf("PostgreSQL port must be between 1 and 65535")
	}
	if cfg.PostgreSQL.Database == "" {
		return fmt.Errorf("PostgreSQL database name is required")
	}
	if cfg.PostgreSQL.Username == "" {
		return fmt.Errorf("PostgreSQL username is required")
	}
	if cfg.PostgreSQL.MaxOpenConns <= 0 {
		return fmt.Errorf("PostgreSQL max open connections must be positive")
	}
	if cfg.PostgreSQL.MaxIdleConns <= 0 {
		return fmt.Errorf("PostgreSQL max idle connections must be positive")
	}
	if cfg.PostgreSQL.MaxIdleConns > cfg.PostgreSQL.MaxOpenConns {
		return fmt.Errorf("PostgreSQL max idle connections cannot exceed max open connections")
	}

	// Validate DynamoDB config
	if cfg.DynamoDB.Region == "" {
		return fmt.Errorf("DynamoDB region is required")
	}
	if cfg.DynamoDB.Table == "" {
		return fmt.Errorf("DynamoDB table name is required")
	}

	return nil
}

// GetManager returns the underlying database manager
func (dm *DatabaseManager) GetManager(backend db.BackendType) db.DBManager {
	return dm.manager.GetManager(backend)
}

// Close closes the database connections
func (dm *DatabaseManager) Close() error {
	if dm.manager != nil {
		return dm.manager.Close()
	}
	return nil
}

// HealthCheck performs a health check on all database connections
func (dm *DatabaseManager) HealthCheck(ctx context.Context) error {
	if dm.manager == nil {
		return fmt.Errorf("database manager not initialized")
	}

	// Check PostgreSQL health
	pgManager := dm.manager.GetManager(db.BackendGorm)
	if pgManager != nil {
		if err := dm.checkPostgreSQLHealth(ctx, pgManager); err != nil {
			return fmt.Errorf("PostgreSQL health check failed: %w", err)
		}
	}

	// Check DynamoDB health
	dynamoManager := dm.manager.GetManager(db.BackendDynamo)
	if dynamoManager != nil {
		if err := dm.checkDynamoDBHealth(ctx, dynamoManager); err != nil {
			return fmt.Errorf("DynamoDB health check failed: %w", err)
		}
	}

	return nil
}

// checkPostgreSQLHealth checks PostgreSQL connection health
func (dm *DatabaseManager) checkPostgreSQLHealth(ctx context.Context, manager db.DBManager) error {
	postgresManager, ok := manager.(*db.PostgresManager)
	if !ok {
		return fmt.Errorf("failed to cast to PostgresManager")
	}

	db, err := postgresManager.GetDB(ctx, false)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Perform a simple query to check connectivity
	var result int
	if err := db.Raw("SELECT 1").Scan(&result).Error; err != nil {
		return fmt.Errorf("database query failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected query result: %d", result)
	}

	return nil
}

// checkDynamoDBHealth checks DynamoDB connection health
func (dm *DatabaseManager) checkDynamoDBHealth(ctx context.Context, manager db.DBManager) error {
	// For now, just check if the manager is available
	// In a real implementation, you might want to perform a simple DynamoDB operation
	if manager == nil {
		return fmt.Errorf("DynamoDB manager not available")
	}
	return nil
}

// GetPostgresManager returns the PostgreSQL manager specifically
func GetPostgresManager(cfg config.MultiDatabaseConfig) (db.DBManager, error) {
	multiManager, err := NewDatabaseManager(cfg)
	if err != nil {
		return nil, err
	}

	pgManager := multiManager.GetManager(db.BackendGorm)
	if pgManager == nil {
		return nil, fmt.Errorf("failed to get PostgreSQL manager")
	}

	return pgManager, nil
}

// GetDynamoManager returns the DynamoDB manager specifically
func GetDynamoManager(cfg config.MultiDatabaseConfig) (db.DBManager, error) {
	multiManager, err := NewDatabaseManager(cfg)
	if err != nil {
		return nil, err
	}

	dynamoManager := multiManager.GetManager(db.BackendDynamo)
	if dynamoManager == nil {
		return nil, fmt.Errorf("failed to get DynamoDB manager")
	}

	return dynamoManager, nil
}

// NewManager creates a database manager from the regular Config type
func NewManager(cfg *config.Config) (*DatabaseManager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	// Convert Config to MultiDatabaseConfig
	multiCfg := configToMultiDatabaseConfig(cfg)
	return NewDatabaseManager(multiCfg)
}

// configToMultiDatabaseConfig converts a regular Config to MultiDatabaseConfig
func configToMultiDatabaseConfig(cfg *config.Config) config.MultiDatabaseConfig {
	// Set defaults for inmemory provider
	host := cfg.Database.Host
	port := cfg.Database.Port
	user := cfg.Database.User
	password := cfg.Database.Password
	database := cfg.Database.Name

	if cfg.Database.Provider == "inmemory" {
		// Use SQLite in-memory for testing
		host = "localhost"
		port = 5432
		user = "test"
		password = "test"
		database = ":memory:"
	}

	return config.MultiDatabaseConfig{
		PostgreSQL: config.PostgreSQLConfig{
			Host:            host,
			Port:            port,
			Database:        database,
			Username:        user,
			Password:        password,
			SSLMode:         cfg.Database.SSLMode,
			MaxOpenConns:    cfg.Database.MaxConns,
			MaxIdleConns:    cfg.Database.MaxIdleTime,
			ConnMaxLifetime: 300, // Default value
		},
		DynamoDB: config.DynamoDBConfig{
			Region:          cfg.Database.Region,
			Endpoint:        "", // Not specified in regular config
			AccessKeyID:     cfg.Database.AccessKey,
			SecretAccessKey: cfg.Database.SecretKey,
			DisableSSL:      false,            // Default value
			Table:           "kisanlink_ecom", // Default value
		},
	}
}
