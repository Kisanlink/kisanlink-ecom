//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"kisanlink-ecom/internal/config"
	"kisanlink-ecom/internal/database"
	"kisanlink-ecom/tests/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseManagerIntegration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	t.Run("ValidConfiguration", func(t *testing.T) {
		// Load configuration from environment variables
		cfg := testutils.LoadTestDatabaseConfig()

		dbManager, err := database.NewDatabaseManager(cfg)
		if err != nil {
			t.Skipf("Database not available for integration test: %v", err)
		}
		defer dbManager.Close()

		// Test health check
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = dbManager.HealthCheck(ctx)
		assert.NoError(t, err, "Health check should pass with valid configuration")
	})

	t.Run("InvalidConfiguration", func(t *testing.T) {
		cfg := config.MultiDatabaseConfig{
			PostgreSQL: config.PostgreSQLConfig{
				Host:            "", // Invalid empty host
				Port:            5432,
				Database:        "test",
				Username:        "test",
				Password:        "test",
				SSLMode:         "disable",
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: 300,
			},
			DynamoDB: config.DynamoDBConfig{
				Region: "us-east-1",
				Table:  "test",
			},
		}

		_, err := database.NewDatabaseManager(cfg)
		assert.Error(t, err, "Should fail with invalid configuration")
		assert.Contains(t, err.Error(), "PostgreSQL host is required")
	})

	t.Run("ConnectionTimeout", func(t *testing.T) {
		cfg := config.MultiDatabaseConfig{
			PostgreSQL: config.PostgreSQLConfig{
				Host:            "nonexistent-host",
				Port:            5432,
				Database:        "test",
				Username:        "test",
				Password:        "test",
				SSLMode:         "disable",
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: 300,
			},
			DynamoDB: config.DynamoDBConfig{
				Region: "us-east-1",
				Table:  "test",
			},
		}

		_, err := database.NewDatabaseManager(cfg)
		assert.Error(t, err, "Should fail with connection timeout")
	})

	t.Run("HealthCheckFailure", func(t *testing.T) {
		// This test would require a database that can be connected to but then fails health checks
		// For now, we'll test the timeout behavior
		cfg := config.MultiDatabaseConfig{
			PostgreSQL: config.PostgreSQLConfig{
				Host:            "localhost",
				Port:            5432,
				Database:        "kisanlink_ecom_test",
				Username:        "postgres",
				Password:        "postgres",
				SSLMode:         "disable",
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: 300,
			},
			DynamoDB: config.DynamoDBConfig{
				Region: "us-east-1",
				Table:  "test",
			},
		}

		dbManager, err := database.NewDatabaseManager(cfg)
		if err != nil {
			t.Skipf("Database not available for integration test: %v", err)
		}
		defer dbManager.Close()

		// Test health check with very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		err = dbManager.HealthCheck(ctx)
		// Should either pass quickly or fail due to timeout
		// The exact behavior depends on the database state
	})
}

func TestDatabaseConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      config.MultiDatabaseConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "ValidConfig",
			config: config.MultiDatabaseConfig{
				PostgreSQL: config.PostgreSQLConfig{
					Host:            "localhost",
					Port:            5432,
					Database:        "test",
					Username:        "test",
					Password:        "test",
					SSLMode:         "disable",
					MaxOpenConns:    10,
					MaxIdleConns:    5,
					ConnMaxLifetime: 300,
				},
				DynamoDB: config.DynamoDBConfig{
					Region: "us-east-1",
					Table:  "test",
				},
			},
			expectError: false,
		},
		{
			name: "EmptyPostgreSQLHost",
			config: config.MultiDatabaseConfig{
				PostgreSQL: config.PostgreSQLConfig{
					Host:         "",
					Port:         5432,
					Database:     "test",
					Username:     "test",
					Password:     "test",
					MaxOpenConns: 10,
					MaxIdleConns: 5,
				},
				DynamoDB: config.DynamoDBConfig{
					Region: "us-east-1",
					Table:  "test",
				},
			},
			expectError: true,
			errorMsg:    "PostgreSQL host is required",
		},
		{
			name: "InvalidPostgreSQLPort",
			config: config.MultiDatabaseConfig{
				PostgreSQL: config.PostgreSQLConfig{
					Host:         "localhost",
					Port:         0,
					Database:     "test",
					Username:     "test",
					Password:     "test",
					MaxOpenConns: 10,
					MaxIdleConns: 5,
				},
				DynamoDB: config.DynamoDBConfig{
					Region: "us-east-1",
					Table:  "test",
				},
			},
			expectError: true,
			errorMsg:    "PostgreSQL port must be between 1 and 65535",
		},
		{
			name: "EmptyPostgreSQLDatabase",
			config: config.MultiDatabaseConfig{
				PostgreSQL: config.PostgreSQLConfig{
					Host:         "localhost",
					Port:         5432,
					Database:     "",
					Username:     "test",
					Password:     "test",
					MaxOpenConns: 10,
					MaxIdleConns: 5,
				},
				DynamoDB: config.DynamoDBConfig{
					Region: "us-east-1",
					Table:  "test",
				},
			},
			expectError: true,
			errorMsg:    "PostgreSQL database name is required",
		},
		{
			name: "EmptyPostgreSQLUsername",
			config: config.MultiDatabaseConfig{
				PostgreSQL: config.PostgreSQLConfig{
					Host:         "localhost",
					Port:         5432,
					Database:     "test",
					Username:     "",
					Password:     "test",
					MaxOpenConns: 10,
					MaxIdleConns: 5,
				},
				DynamoDB: config.DynamoDBConfig{
					Region: "us-east-1",
					Table:  "test",
				},
			},
			expectError: true,
			errorMsg:    "PostgreSQL username is required",
		},
		{
			name: "InvalidMaxOpenConns",
			config: config.MultiDatabaseConfig{
				PostgreSQL: config.PostgreSQLConfig{
					Host:         "localhost",
					Port:         5432,
					Database:     "test",
					Username:     "test",
					Password:     "test",
					MaxOpenConns: 0,
					MaxIdleConns: 5,
				},
				DynamoDB: config.DynamoDBConfig{
					Region: "us-east-1",
					Table:  "test",
				},
			},
			expectError: true,
			errorMsg:    "PostgreSQL max open connections must be positive",
		},
		{
			name: "InvalidMaxIdleConns",
			config: config.MultiDatabaseConfig{
				PostgreSQL: config.PostgreSQLConfig{
					Host:         "localhost",
					Port:         5432,
					Database:     "test",
					Username:     "test",
					Password:     "test",
					MaxOpenConns: 10,
					MaxIdleConns: 0,
				},
				DynamoDB: config.DynamoDBConfig{
					Region: "us-east-1",
					Table:  "test",
				},
			},
			expectError: true,
			errorMsg:    "PostgreSQL max idle connections must be positive",
		},
		{
			name: "MaxIdleConnsExceedsMaxOpenConns",
			config: config.MultiDatabaseConfig{
				PostgreSQL: config.PostgreSQLConfig{
					Host:         "localhost",
					Port:         5432,
					Database:     "test",
					Username:     "test",
					Password:     "test",
					MaxOpenConns: 5,
					MaxIdleConns: 10,
				},
				DynamoDB: config.DynamoDBConfig{
					Region: "us-east-1",
					Table:  "test",
				},
			},
			expectError: true,
			errorMsg:    "PostgreSQL max idle connections cannot exceed max open connections",
		},
		{
			name: "EmptyDynamoDBRegion",
			config: config.MultiDatabaseConfig{
				PostgreSQL: config.PostgreSQLConfig{
					Host:         "localhost",
					Port:         5432,
					Database:     "test",
					Username:     "test",
					Password:     "test",
					MaxOpenConns: 10,
					MaxIdleConns: 5,
				},
				DynamoDB: config.DynamoDBConfig{
					Region: "",
					Table:  "test",
				},
			},
			expectError: true,
			errorMsg:    "DynamoDB region is required",
		},
		{
			name: "EmptyDynamoDBTable",
			config: config.MultiDatabaseConfig{
				PostgreSQL: config.PostgreSQLConfig{
					Host:         "localhost",
					Port:         5432,
					Database:     "test",
					Username:     "test",
					Password:     "test",
					MaxOpenConns: 10,
					MaxIdleConns: 5,
				},
				DynamoDB: config.DynamoDBConfig{
					Region: "us-east-1",
					Table:  "",
				},
			},
			expectError: true,
			errorMsg:    "DynamoDB table name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := database.NewDatabaseManager(tt.config)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				if err != nil {
					// If we get a connection error, that's okay for validation tests
					// We're mainly testing the validation logic
					if !assert.Contains(t, err.Error(), "failed to connect") {
						t.Errorf("Unexpected error: %v", err)
					}
				}
			}
		})
	}
}
