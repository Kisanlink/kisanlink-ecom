package database_test

import (
	"testing"
	"time"

	"kisanlink-ecom/internal/config"
	"kisanlink-ecom/internal/database"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrationRunner(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Load test configuration
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Provider: "inmemory",
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			Name:     ":memory:",
			SSLMode:  "disable",
		},
	}

	// Create database manager
	dbManager, err := database.NewManager(cfg)
	require.NoError(t, err)
	defer dbManager.Close()

	// Get PostgreSQL manager
	pgManager := dbManager.GetManager(db.BackendGorm)
	require.NotNil(t, pgManager)

	// Create migration runner
	runner, err := database.NewMigrationRunner(pgManager)
	require.NoError(t, err)

	// Test getting applied migrations (should be empty initially)
	applied, err := runner.GetAppliedMigrations()
	require.NoError(t, err)
	assert.Empty(t, applied)

	// Test checking if migration is applied
	isApplied, err := runner.IsMigrationApplied("v1.0.0")
	require.NoError(t, err)
	assert.False(t, isApplied)

	// Test recording a migration
	err = runner.RecordMigration("v1.0.0", "Test migration", "checksum123", true, 100*time.Millisecond, "")
	require.NoError(t, err)

	// Test that migration is now applied
	isApplied, err = runner.IsMigrationApplied("v1.0.0")
	require.NoError(t, err)
	assert.True(t, isApplied)

	// Test getting applied migrations (should have one now)
	applied, err = runner.GetAppliedMigrations()
	require.NoError(t, err)
	assert.Len(t, applied, 1)
	assert.Equal(t, "v1.0.0", applied[0].Version)
	assert.Equal(t, "Test migration", applied[0].Description)
	assert.True(t, applied[0].Success)
}

func TestIdempotentMigrator(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Load test configuration
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Provider: "inmemory",
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			Name:     ":memory:",
			SSLMode:  "disable",
		},
	}

	// Create database manager
	dbManager, err := database.NewManager(cfg)
	require.NoError(t, err)
	defer dbManager.Close()

	// Get PostgreSQL manager
	pgManager := dbManager.GetManager(db.BackendGorm)
	require.NotNil(t, pgManager)

	// Create idempotent migrator
	migrator, err := database.NewIdempotentMigrator(pgManager)
	require.NoError(t, err)

	// Test dry run mode
	migrator.SetDryRun(true)
	migrator.SetSafetyCheck(false) // Disable for testing

	// Test getting pending migrations
	pending, err := migrator.GetPendingMigrations()
	require.NoError(t, err)
	assert.NotEmpty(t, pending) // Should have default migrations

	// Test printing migration plan
	err = migrator.PrintMigrationPlan()
	require.NoError(t, err)

	// Test running migrations in dry-run mode
	err = migrator.RunMigrations()
	require.NoError(t, err)
}

func TestMigrationManager(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Load test configuration
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Provider: "inmemory",
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			Name:     ":memory:",
			SSLMode:  "disable",
		},
	}

	// Create database manager
	dbManager, err := database.NewManager(cfg)
	require.NoError(t, err)
	defer dbManager.Close()

	// Get PostgreSQL manager
	pgManager := dbManager.GetManager(db.BackendGorm)
	require.NotNil(t, pgManager)

	// Create migration manager
	manager, err := database.NewMigrationManager(pgManager)
	require.NoError(t, err)

	// Test status command
	err = manager.ExecuteCommand(database.CommandStatus)
	require.NoError(t, err)

	// Test plan command
	err = manager.ExecuteCommand(database.CommandPlan)
	require.NoError(t, err)

	// Test dry-run command
	err = manager.ExecuteCommand(database.CommandDryRun)
	require.NoError(t, err)

	// Test history command
	err = manager.ExecuteCommand(database.CommandHistory)
	require.NoError(t, err)
}
