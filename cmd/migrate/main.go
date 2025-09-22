package main

import (
	"log"
	"os"

	"kisanlink-ecom/internal/config"
	"kisanlink-ecom/internal/database"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create database manager
	dbManager, err := database.NewManager(cfg)
	if err != nil {
		log.Fatalf("Failed to create database manager: %v", err)
	}
	defer dbManager.Close()

	// Get PostgreSQL manager
	pgManager := dbManager.GetManager(db.BackendGorm)
	if pgManager == nil {
		log.Fatalf("PostgreSQL manager not available")
	}

	// Handle migration CLI commands
	args := os.Args[1:]
	if len(args) == 0 {
		database.PrintUsage()
		os.Exit(1)
	}

	if err := database.RunMigrationCLI(pgManager, args); err != nil {
		log.Fatalf("Migration command failed: %v", err)
	}
}
