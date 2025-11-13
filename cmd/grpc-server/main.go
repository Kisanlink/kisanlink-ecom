// Package main implements the gRPC server for the Collaborator service
package main

import (
	"context"
	"fmt"
	"kisanlink-ecom/internal/aaa"
	"kisanlink-ecom/internal/config"
	collabDomain "kisanlink-ecom/internal/domain/collaborator"
	grpcserver "kisanlink-ecom/internal/grpc"
	"kisanlink-ecom/internal/grpc/handlers/collaborator"
	"kisanlink-ecom/internal/grpc/interceptors"
	"kisanlink-ecom/internal/saga"
	"kisanlink-ecom/internal/services/gst"
	"kisanlink-ecom/internal/services/otp"
	pb "kisanlink-ecom/proto/gen/go/collaborator/v1"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	collabModel "kisanlink-ecom/entities/models/collaborator"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func main() {
	// Load configuration
	cfg := config.LoadGRPCServerConfig()

	log.Printf("Starting gRPC server with configuration:")
	log.Printf("  Port: %d", cfg.Port)
	log.Printf("  Max Recv Message Size: %d bytes", cfg.MaxRecvMsgSize)
	log.Printf("  Max Send Message Size: %d bytes", cfg.MaxSendMsgSize)
	log.Printf("  Keepalive Time: %v", cfg.KeepaliveTime)
	log.Printf("  Keepalive Timeout: %v", cfg.KeepaliveTimeout)
	log.Printf("  Reflection Enabled: %v", cfg.ReflectionEnabled)

	// Create listener
	address := grpcserver.GetServerAddress(cfg)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", address, err)
	}

	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Create mock JWT validator (will be replaced with real validator in production)
	jwtValidator := &interceptors.MockJWTValidator{}

	// Build 10-layer interceptor chain
	interceptorChain := grpcserver.BuildInterceptorChain(logger, jwtValidator)

	server := grpcserver.NewServer(cfg, interceptorChain)

	// Register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register reflection for grpcurl and development
	if cfg.ReflectionEnabled {
		reflection.Register(server)
		log.Println("gRPC reflection enabled")
	}

	// Initialize dependencies for P1-4
	appCfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load application config: %v", err)
	}

	// Initialize PostgreSQL database
	db, err := initDatabase(appCfg, logger)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Redis client
	redisClient := initRedis(appCfg, logger)

	// Initialize AAA client
	aaaClient, err := initAAAClient(context.Background(), appCfg, logger)
	if err != nil {
		log.Fatalf("Failed to initialize AAA client: %v", err)
	}
	defer func() {
		if err := aaaClient.Close(); err != nil {
			logger.WithError(err).Error("Failed to close AAA client")
		}
	}()

	// Initialize GST service
	gstService := gst.NewService(db, redisClient, logger)

	// Initialize Saga executor
	sagaStorage := saga.NewDBSagaStorage(db)
	sagaMetrics := saga.NewSagaMetrics()
	zapLogger, _ := zap.NewProduction()
	sagaExecutor := saga.NewSagaExecutor(sagaStorage, zapLogger, sagaMetrics)

	// Initialize OTP service
	otpService := otp.NewOTPService(zapLogger)

	// Initialize State Machine
	stateMachine := collabDomain.NewStateMachine(db, zapLogger)

	// Create collaborator handler
	collaboratorHandler := collaborator.NewHandler(db, logger, aaaClient, gstService, sagaExecutor, otpService, stateMachine)

	// Register collaborator service
	pb.RegisterCollaboratorServiceServer(server, collaboratorHandler)
	log.Println("Collaborator service registered successfully")

	// Start server in goroutine
	go func() {
		log.Printf("gRPC server listening on %s", address)
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	log.Printf("Received signal: %v. Shutting down gRPC server...", sig)

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Set health check to NOT_SERVING during shutdown
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	// Create channel to signal shutdown completion
	done := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(done)
	}()

	// Wait for graceful shutdown to complete or timeout
	select {
	case <-done:
		log.Println("gRPC server stopped gracefully")
	case <-ctx.Done():
		log.Println("Graceful shutdown timeout, forcing stop")
		server.Stop()
	}

	log.Println("Server shutdown complete")
}

// initDatabase initializes PostgreSQL database connection
func initDatabase(cfg *config.Config, logger *logrus.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	gormCfg := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(cfg.Database.MaxConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleTime)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto-migrate collaborator model
	if err := db.AutoMigrate(&collabModel.Collaborator{}); err != nil {
		return nil, fmt.Errorf("failed to migrate collaborator model: %w", err)
	}

	logger.Info("Database initialized successfully")
	return db, nil
}

// initRedis initializes Redis client
func initRedis(cfg *config.Config, logger *logrus.Logger) redis.UniversalClient {
	redisDB, _ := strconv.Atoi(cfg.Redis.DB)

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       redisDB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.WithError(err).Warn("Redis connection failed, continuing without Redis")
		return nil
	}

	logger.Info("Redis initialized successfully")
	return client
}

// initAAAClient initializes AAA service client
func initAAAClient(ctx context.Context, cfg *config.Config, logger *logrus.Logger) (*aaa.Client, error) {
	aaaConfig := &aaa.Config{
		Address:        cfg.AAA.GRPCServerAddr,
		PoolSize:       10,
		DialTimeout:    5 * time.Second,
		CallTimeout:    10 * time.Second,
		MaxRetries:     cfg.AAA.Retries,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     5 * time.Second,
		BackoffFactor:  2.0,
		EnableTLS:      cfg.AAA.TLSEnabled,
		CertFile:       cfg.AAA.CertPath,
		KeyFile:        cfg.AAA.KeyPath,
		CAFile:         cfg.AAA.CAPath,
		ServerName:     cfg.AAA.ServerName,
		CircuitBreaker: aaa.CircuitBreakerConfig{
			MaxRequests:  10,
			Interval:     10 * time.Second,
			Timeout:      cfg.AAA.CircuitBreakerResetTimeout,
			FailureRatio: 0.6,
			MinRequests:  10,
		},
		KeepAliveTime:    30 * time.Second,
		KeepAliveTimeout: 10 * time.Second,
	}

	client, err := aaa.NewClient(ctx, aaaConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create AAA client: %w", err)
	}

	logger.Info("AAA client initialized successfully")
	return client, nil
}
