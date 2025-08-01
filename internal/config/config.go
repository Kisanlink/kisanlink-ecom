package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Redis    RedisConfig
	Email    EmailConfig
	Upload   UploadConfig
	CORS     CORSConfig
	Logging  LoggingConfig
	GRPC     GRPCConfig
}

// ServerConfig holds server configuration.
type ServerConfig struct {
	Port    string
	Mode    string
	Version string
}

// DatabaseConfig holds database configuration.
type DatabaseConfig struct {
	Provider    string `env:"DB_PROVIDER" envDefault:"inmemory"`
	Host        string `env:"DB_HOST" envDefault:"localhost"`
	Port        int    `env:"DB_PORT" envDefault:"5432"`
	User        string `env:"DB_USER" envDefault:"postgres"`
	Password    string `env:"DB_PASSWORD"`
	Name        string `env:"DB_NAME" envDefault:"kisanlink_ecom"`
	SSLMode     string `env:"DB_SSLMODE" envDefault:"disable"`
	MaxConns    int    `env:"DB_MAX_CONNS" envDefault:"10"`
	MaxIdleTime int    `env:"DB_MAX_IDLE_TIME" envDefault:"30"`

	// For DynamoDB
	Region    string `env:"AWS_REGION" envDefault:"us-east-1"`
	AccessKey string `env:"AWS_ACCESS_KEY_ID"`
	SecretKey string `env:"AWS_SECRET_ACCESS_KEY"`

	// For other NoSQL databases
	ConnectionString string `env:"DB_CONNECTION_STRING"`
}

// JWTConfig holds JWT configuration.
type JWTConfig struct {
	Secret string
	Expiry string
}

// RedisConfig holds Redis configuration.
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       string
}

// EmailConfig holds email configuration.
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
}

// UploadConfig holds file upload configuration.
type UploadConfig struct {
	Directory   string
	MaxFileSize string
}

// CORSConfig holds CORS configuration.
type CORSConfig struct {
	AllowedOrigins string
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	Level  string
	Format string
}

// GRPCConfig holds gRPC configuration.
type GRPCConfig struct {
	ServerAddr string
}

// Load loads configuration from environment variables.
func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		// It's okay if .env file doesn't exist in production
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	dbPort := 5432
	if portStr := getEnv("DB_PORT", "5432"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			dbPort = p
		}
	}

	maxConns := 10
	if maxConnsStr := getEnv("DB_MAX_CONNS", "10"); maxConnsStr != "" {
		if mc, err := strconv.Atoi(maxConnsStr); err == nil {
			maxConns = mc
		}
	}

	maxIdleTime := 30
	if maxIdleTimeStr := getEnv("DB_MAX_IDLE_TIME", "30"); maxIdleTimeStr != "" {
		if mit, err := strconv.Atoi(maxIdleTimeStr); err == nil {
			maxIdleTime = mit
		}
	}

	return &Config{
		Server: ServerConfig{
			Port:    getEnv("PORT", "8080"),
			Mode:    getEnv("GIN_MODE", "debug"),
			Version: getEnv("API_VERSION", "v1"),
		},
		Database: DatabaseConfig{
			Provider:         getEnv("DB_PROVIDER", "inmemory"),
			Host:             getEnv("DB_HOST", "localhost"),
			Port:             dbPort,
			User:             getEnv("DB_USER", "postgres"),
			Password:         getEnv("DB_PASSWORD", "password"),
			Name:             getEnv("DB_NAME", "kisanlink_ecom"),
			SSLMode:          getEnv("DB_SSLMODE", "disable"),
			MaxConns:         maxConns,
			MaxIdleTime:      maxIdleTime,
			Region:           getEnv("AWS_REGION", "us-east-1"),
			AccessKey:        getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretKey:        getEnv("AWS_SECRET_ACCESS_KEY", ""),
			ConnectionString: getEnv("DB_CONNECTION_STRING", ""),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-super-secret-jwt-key-here"),
			Expiry: getEnv("JWT_EXPIRY", "24h"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnv("REDIS_DB", "0"),
		},
		Email: EmailConfig{
			SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			SMTPPort:     getEnv("SMTP_PORT", "587"),
			SMTPUsername: getEnv("SMTP_USERNAME", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		},
		Upload: UploadConfig{
			Directory:   getEnv("UPLOAD_DIR", "./uploads"),
			MaxFileSize: getEnv("MAX_FILE_SIZE", "10MB"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnv("ALLOWED_ORIGINS", "*"),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		GRPC: GRPCConfig{
			ServerAddr: getEnv("AAA_GRPC_SERVER_ADDR", "localhost:50051"),
		},
	}, nil
}

// getEnv gets an environment variable with a fallback value.
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
