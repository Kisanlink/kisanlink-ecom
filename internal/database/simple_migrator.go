package database

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/Kisanlink/kisanlink-ecom/entities/models"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"gorm.io/gorm"
)

// SimpleMigrator handles database migrations using only GORM AutoMigrate
type SimpleMigrator struct {
	runner    *MigrationRunner
	dbManager db.DBManager
	db        *gorm.DB
	dryRun    bool
}

// NewSimpleMigrator creates a new simple migrator that uses only GORM
func NewSimpleMigrator(dbManager db.DBManager) (*SimpleMigrator, error) {
	runner, err := NewMigrationRunner(dbManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration runner: %w", err)
	}

	// Cast to PostgresManager to access GetDB method
	postgresManager, ok := dbManager.(*db.PostgresManager)
	if !ok {
		return nil, fmt.Errorf("failed to cast to PostgresManager")
	}

	database, err := postgresManager.GetDB(context.Background(), false)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	return &SimpleMigrator{
		runner:    runner,
		dbManager: dbManager,
		db:        database,
		dryRun:    false,
	}, nil
}

// SetDryRun enables or disables dry run mode
func (sm *SimpleMigrator) SetDryRun(dryRun bool) {
	sm.dryRun = dryRun
}

// RunMigrations executes migrations using only GORM AutoMigrate
func (sm *SimpleMigrator) RunMigrations() error {
	log.Println("Starting simple GORM-only migration...")

	// Clean up any failed migrations first
	if err := sm.runner.CleanupFailedMigrations(); err != nil {
		log.Printf("Warning: Failed to cleanup failed migrations: %v", err)
	}

	// Get all models in dependency order
	allModels := models.ModelsByPriority()

	// Calculate checksum for tracking
	checksum := sm.calculateChecksum(allModels)
	version := "v1.2.0" // Added taxation, discounts, marketplace, outbox, sequence_counter, inventory_audit_log models

	// Check if migration is already applied
	applied, err := sm.runner.IsMigrationApplied(version)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if applied {
		log.Printf("Migration %s already applied, skipping", version)
		return nil
	}

	log.Printf("Applying migration %s with %d models", version, len(allModels))

	// Start timing
	startTime := time.Now()

	// Execute migration
	var migrationErr error
	if sm.dryRun {
		log.Printf("DRY RUN: Would migrate %d models", len(allModels))
		migrationErr = nil
	} else {
		// Use GORM's AutoMigrate with retry logic
		migrationErr = sm.executeWithRetry(allModels)
	}

	duration := time.Since(startTime)

	// Record migration result
	success := migrationErr == nil
	errorMsg := ""
	if migrationErr != nil {
		errorMsg = migrationErr.Error()
		log.Printf("Migration %s failed after %v: %v", version, duration, migrationErr)
	} else {
		log.Printf("Migration %s completed successfully in %v", version, duration)
	}

	// Record the migration attempt
	if !sm.dryRun {
		if recordErr := sm.runner.RecordMigration(version, "GORM AutoMigrate for all models", checksum, success, duration, errorMsg); recordErr != nil {
			log.Printf("Warning: Failed to record migration result: %v", recordErr)
		}
	}

	return migrationErr
}

// executeWithRetry executes GORM AutoMigrate with retry logic
func (sm *SimpleMigrator) executeWithRetry(models []interface{}) error {
	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			log.Printf("Retry attempt %d/%d", attempt, maxRetries)
			time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
		}

		// Clear any cached plans before each attempt
		sm.clearCachedPlans()

		// Execute GORM AutoMigrate
		err := sm.db.AutoMigrate(models...)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if this is a retryable error
		if !sm.isRetryableError(err) {
			log.Printf("Non-retryable error encountered: %v", err)
			break
		}

		log.Printf("Retryable error on attempt %d: %v", attempt, err)
	}

	return fmt.Errorf("migration failed after %d attempts: %w", maxRetries, lastErr)
}

// clearCachedPlans clears PostgreSQL cached plans
func (sm *SimpleMigrator) clearCachedPlans() {
	// Clear cached plans to avoid conflicts
	if err := sm.db.Exec("DISCARD PLANS").Error; err != nil {
		log.Printf("Warning: Failed to clear cached plans: %v", err)
	}

	// Also discard temporary objects
	if err := sm.db.Exec("DISCARD TEMP").Error; err != nil {
		log.Printf("Warning: Failed to clear temp objects: %v", err)
	}
}

// isRetryableError determines if an error is worth retrying
func (sm *SimpleMigrator) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// Retryable error patterns
	retryablePatterns := []string{
		"cached plan must not change result type",
		"current transaction is aborted",
		"deadlock detected",
		"could not serialize access",
		"connection reset",
		"connection refused",
		"timeout",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// calculateChecksum calculates a checksum for the models
func (sm *SimpleMigrator) calculateChecksum(models []interface{}) string {
	hasher := sha256.New()

	// Include model names
	modelNames := make([]string, len(models))
	for i, model := range models {
		t := reflect.TypeOf(model)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		modelNames[i] = t.Name()
	}

	sort.Strings(modelNames)
	hasher.Write([]byte(strings.Join(modelNames, ",")))

	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// ForceCleanMigration drops problematic tables and recreates them (development only)
func (sm *SimpleMigrator) ForceCleanMigration() error {
	log.Println("WARNING: Performing force clean migration - this may cause data loss!")

	// This should only be used in development environments
	if !sm.isDevelopmentEnvironment() {
		return fmt.Errorf("force clean migration is only allowed in development environments")
	}

	// Clean up failed migrations
	if err := sm.runner.CleanupFailedMigrations(); err != nil {
		return fmt.Errorf("failed to cleanup failed migrations: %w", err)
	}

	// Clear cached plans
	sm.clearCachedPlans()

	// Drop problematic tables in dependency order
	problematicTables := []string{
		"inventory_lots", // Has FK to catalog_items
		"catalog_items",  // Main problematic table
	}

	for _, table := range problematicTables {
		log.Printf("Dropping table: %s", table)
		if err := sm.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)).Error; err != nil {
			log.Printf("Warning: Failed to drop table %s: %v", table, err)
		}
	}

	// Clear cached plans again after dropping tables
	sm.clearCachedPlans()

	// Now run migrations normally
	return sm.RunMigrations()
}

// isDevelopmentEnvironment checks if we're in a development environment
func (sm *SimpleMigrator) isDevelopmentEnvironment() bool {
	var dbName string
	if err := sm.db.Raw("SELECT current_database()").Scan(&dbName).Error; err != nil {
		return false
	}

	// Consider it development if database name contains 'dev', 'test', or 'local'
	devIndicators := []string{"dev", "test", "local", "development"}
	dbNameLower := strings.ToLower(dbName)

	for _, indicator := range devIndicators {
		if strings.Contains(dbNameLower, indicator) {
			return true
		}
	}

	return false
}
