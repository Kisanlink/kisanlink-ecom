package testutils

import (
	"os"
	"strconv"

	"github.com/Kisanlink/kisanlink-ecom/internal/config"
)

// LoadTestDatabaseConfig loads database configuration for integration tests from environment variables
func LoadTestDatabaseConfig() config.MultiDatabaseConfig {
	return config.MultiDatabaseConfig{
		PostgreSQL: config.PostgreSQLConfig{
			Host:            getEnvOrDefault("DB_POSTGRES_HOST", "localhost"),
			Port:            getEnvAsIntOrDefault("DB_POSTGRES_PORT", 5432),
			Database:        getEnvOrDefault("DB_POSTGRES_DBNAME", "kisanlink_ecom_test"),
			Username:        getEnvOrDefault("DB_POSTGRES_USER", "postgres"),
			Password:        getEnvOrDefault("DB_POSTGRES_PASSWORD", "postgres"),
			SSLMode:         getEnvOrDefault("DB_POSTGRES_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsIntOrDefault("DB_POSTGRES_MAX_CONNS", 10),
			MaxIdleConns:    getEnvAsIntOrDefault("DB_POSTGRES_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsIntOrDefault("DB_POSTGRES_CONN_MAX_LIFETIME", 300),
		},
		DynamoDB: config.DynamoDBConfig{
			Region:          getEnvOrDefault("DB_DYNAMO_REGION", "us-east-1"),
			Endpoint:        getEnvOrDefault("DYNAMODB_ENDPOINT", "http://localhost:8000"),
			AccessKeyID:     getEnvOrDefault("DYNAMODB_ACCESS_KEY_ID", "test"),
			SecretAccessKey: getEnvOrDefault("DYNAMODB_SECRET_ACCESS_KEY", "test"),
			DisableSSL:      getEnvAsBoolOrDefault("DYNAMODB_DISABLE_SSL", true),
			Table:           getEnvOrDefault("DYNAMODB_TABLE", "kisanlink_ecom_test"),
		},
	}
}

// GetEnvOrDefault gets an environment variable with a fallback value
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvAsIntOrDefault gets an environment variable as an integer with a fallback value
func GetEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

// GetEnvAsBoolOrDefault gets an environment variable as a boolean with a fallback value
func GetEnvAsBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

// getEnvOrDefault gets an environment variable with a fallback value (private version)
func getEnvOrDefault(key, defaultValue string) string {
	return GetEnvOrDefault(key, defaultValue)
}

// getEnvAsIntOrDefault gets an environment variable as an integer with a fallback value (private version)
func getEnvAsIntOrDefault(key string, defaultValue int) int {
	return GetEnvAsIntOrDefault(key, defaultValue)
}

// getEnvAsBoolOrDefault gets an environment variable as a boolean with a fallback value (private version)
func getEnvAsBoolOrDefault(key string, defaultValue bool) bool {
	return GetEnvAsBoolOrDefault(key, defaultValue)
}
