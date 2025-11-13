package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kisanlink-ecom/internal/config"
	"kisanlink-ecom/internal/database"
	"kisanlink-ecom/internal/repositories/inventory"
	inventoryService "kisanlink-ecom/internal/services/inventory"
	"kisanlink-ecom/internal/services/notifications"
)

func main() {
	log.Println("Starting KisanLink Inventory Alert Scheduler...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Convert to multi-DB config
	multiDBConfig := config.MultiDatabaseConfig{
		PostgreSQL: config.PostgreSQLConfig{
			Host:         cfg.Database.Host,
			Port:         cfg.Database.Port,
			Database:     cfg.Database.Name,
			Username:     cfg.Database.User,
			Password:     cfg.Database.Password,
			SSLMode:      cfg.Database.SSLMode,
			MaxOpenConns: cfg.Database.MaxConns,
			MaxIdleConns: cfg.Database.MaxConns / 2,
		},
		DynamoDB: config.DynamoDBConfig{
			Region:          cfg.Database.Region,
			Endpoint:        "",
			AccessKeyID:     cfg.Database.AccessKey,
			SecretAccessKey: cfg.Database.SecretKey,
			DisableSSL:      false,
			Table:           "kisanlink_ecom",
		},
	}

	// Initialize database connection
	dbManager, err := database.NewDatabaseManager(multiDBConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := dbManager.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// Initialize repositories with proper manager
	dbMgr := dbManager.GetManager("postgres")
	alertRepo := inventory.NewAlertRepository(dbMgr)
	inventoryRepo := inventory.NewInventoryRepository(dbMgr)

	// Initialize services
	notificationSvc := notifications.NewNotificationService()
	alertSvc := inventoryService.NewAlertService(alertRepo, inventoryRepo, notificationSvc)

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start background jobs
	log.Println("Starting background jobs...")

	// Low stock check - runs every hour
	lowStockTicker := time.NewTicker(1 * time.Hour)
	defer lowStockTicker.Stop()

	// Expiry check - runs every day at midnight
	expiryTicker := time.NewTicker(24 * time.Hour)
	defer expiryTicker.Stop()

	// Run initial checks
	log.Println("Running initial low stock check...")
	if err := alertSvc.CheckLowStockItems(ctx); err != nil {
		log.Printf("Error in initial low stock check: %v", err)
	}

	log.Println("Running initial expiry check...")
	if err := alertSvc.MarkExpiredItems(ctx); err != nil {
		log.Printf("Error in initial expiry check: %v", err)
	}

	log.Println("Scheduler started successfully. Press Ctrl+C to stop.")

	// Main loop
	for {
		select {
		case <-lowStockTicker.C:
			log.Println("Running scheduled low stock check...")
			if err := alertSvc.CheckLowStockItems(ctx); err != nil {
				log.Printf("Error in low stock check: %v", err)
			}

		case <-expiryTicker.C:
			log.Println("Running scheduled expiry check...")
			if err := alertSvc.MarkExpiredItems(ctx); err != nil {
				log.Printf("Error in expiry check: %v", err)
			}

		case sig := <-sigChan:
			log.Printf("Received signal: %v. Shutting down gracefully...", sig)
			cancel()
			time.Sleep(2 * time.Second) // Give running jobs time to complete
			log.Println("Scheduler stopped.")
			return
		}
	}
}
