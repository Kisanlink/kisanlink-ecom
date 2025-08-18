package database

import (
	"context"
	"fmt"
	"log"

	"kisanlink-ecom/internal/config"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// NewDatabaseManager creates a new database manager with proper error handling
func NewDatabaseManager(cfg config.MultiDatabaseConfig) (*db.DatabaseManager, error) {
	log.Printf("Loading database configuration...")

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

	// Connect to database with context
	ctx := context.Background()
	if err := dbManager.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Printf("Database manager initialized successfully")
	return dbManager, nil
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
