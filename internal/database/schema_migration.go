package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"gorm.io/gorm"
)

// SchemaMigration represents a database schema migration record
type SchemaMigration struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Version     string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"version"`
	Description string    `gorm:"type:text" json:"description"`
	AppliedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"applied_at"`
	Checksum    string    `gorm:"type:varchar(64);not null" json:"checksum"`
	Success     bool      `gorm:"not null;default:false" json:"success"`
	ErrorMsg    string    `gorm:"type:text" json:"error_msg,omitempty"`
	Duration    int64     `gorm:"not null;default:0" json:"duration"` // Duration in milliseconds
}

// TableName returns the table name for SchemaMigration
func (SchemaMigration) TableName() string {
	return "schema_migrations"
}

// MigrationRunner handles database schema migrations with version tracking
type MigrationRunner struct {
	dbManager db.DBManager
	db        *gorm.DB
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(dbManager db.DBManager) (*MigrationRunner, error) {
	if dbManager == nil {
		return nil, fmt.Errorf("database manager cannot be nil")
	}

	// Cast to PostgresManager to access GetDB method
	postgresManager, ok := dbManager.(*db.PostgresManager)
	if !ok {
		return nil, fmt.Errorf("failed to cast to PostgresManager")
	}

	// Get the raw database connection
	database, err := postgresManager.GetDB(context.Background(), false)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	runner := &MigrationRunner{
		dbManager: dbManager,
		db:        database,
	}

	// Initialize the schema_migrations table
	if err := runner.initializeMigrationTable(); err != nil {
		return nil, fmt.Errorf("failed to initialize migration table: %w", err)
	}

	return runner, nil
}

// initializeMigrationTable creates the schema_migrations table if it doesn't exist
func (mr *MigrationRunner) initializeMigrationTable() error {
	log.Println("Initializing schema_migrations table...")

	// Create the schema_migrations table using GORM AutoMigrate
	if err := mr.db.AutoMigrate(&SchemaMigration{}); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	log.Println("Schema migrations table initialized successfully")
	return nil
}

// GetAppliedMigrations returns all successfully applied migrations
func (mr *MigrationRunner) GetAppliedMigrations() ([]SchemaMigration, error) {
	var migrations []SchemaMigration
	if err := mr.db.Where("success = ?", true).Order("applied_at ASC").Find(&migrations).Error; err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}
	return migrations, nil
}

// IsMigrationApplied checks if a specific migration version has been applied
func (mr *MigrationRunner) IsMigrationApplied(version string) (bool, error) {
	var count int64
	if err := mr.db.Model(&SchemaMigration{}).Where("version = ? AND success = ?", version, true).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check migration status: %w", err)
	}
	return count > 0, nil
}

// RecordMigration records a migration attempt in the database
func (mr *MigrationRunner) RecordMigration(version, description, checksum string, success bool, duration time.Duration, errorMsg string) error {
	migration := SchemaMigration{
		Version:     version,
		Description: description,
		AppliedAt:   time.Now(),
		Checksum:    checksum,
		Success:     success,
		ErrorMsg:    errorMsg,
		Duration:    duration.Milliseconds(),
	}

	if err := mr.db.Create(&migration).Error; err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return nil
}

// ValidateMigrationIntegrity checks if the applied migrations match expected checksums
func (mr *MigrationRunner) ValidateMigrationIntegrity(expectedMigrations map[string]string) error {
	appliedMigrations, err := mr.GetAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	for _, migration := range appliedMigrations {
		expectedChecksum, exists := expectedMigrations[migration.Version]
		if !exists {
			log.Printf("Warning: Applied migration %s not found in expected migrations", migration.Version)
			continue
		}

		if migration.Checksum != expectedChecksum {
			return fmt.Errorf("migration integrity check failed for version %s: expected checksum %s, got %s",
				migration.Version, expectedChecksum, migration.Checksum)
		}
	}

	return nil
}

// GetLastAppliedVersion returns the version of the last successfully applied migration
func (mr *MigrationRunner) GetLastAppliedVersion() (string, error) {
	var migration SchemaMigration
	if err := mr.db.Where("success = ?", true).Order("applied_at DESC").First(&migration).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil // No migrations applied yet
		}
		return "", fmt.Errorf("failed to get last applied migration: %w", err)
	}
	return migration.Version, nil
}

// GetFailedMigrations returns all failed migration attempts
func (mr *MigrationRunner) GetFailedMigrations() ([]SchemaMigration, error) {
	var migrations []SchemaMigration
	if err := mr.db.Where("success = ?", false).Order("applied_at DESC").Find(&migrations).Error; err != nil {
		return nil, fmt.Errorf("failed to get failed migrations: %w", err)
	}
	return migrations, nil
}

// CleanupFailedMigrations removes failed migration records (use with caution)
func (mr *MigrationRunner) CleanupFailedMigrations() error {
	result := mr.db.Where("success = ?", false).Delete(&SchemaMigration{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup failed migrations: %w", result.Error)
	}

	log.Printf("Cleaned up %d failed migration records", result.RowsAffected)
	return nil
}

// GetMigrationHistory returns the complete migration history
func (mr *MigrationRunner) GetMigrationHistory() ([]SchemaMigration, error) {
	var migrations []SchemaMigration
	if err := mr.db.Order("applied_at DESC").Find(&migrations).Error; err != nil {
		return nil, fmt.Errorf("failed to get migration history: %w", err)
	}
	return migrations, nil
}

// PrintMigrationStatus prints the current migration status
func (mr *MigrationRunner) PrintMigrationStatus() error {
	history, err := mr.GetMigrationHistory()
	if err != nil {
		return err
	}

	if len(history) == 0 {
		log.Println("No migrations have been applied")
		return nil
	}

	log.Println("Migration History:")
	log.Println("==================")
	for _, migration := range history {
		status := "SUCCESS"
		if !migration.Success {
			status = "FAILED"
		}

		log.Printf("Version: %s | Status: %s | Applied: %s | Duration: %dms",
			migration.Version,
			status,
			migration.AppliedAt.Format("2006-01-02 15:04:05"),
			migration.Duration)

		if migration.Description != "" {
			log.Printf("  Description: %s", migration.Description)
		}

		if !migration.Success && migration.ErrorMsg != "" {
			log.Printf("  Error: %s", migration.ErrorMsg)
		}
	}

	return nil
}
