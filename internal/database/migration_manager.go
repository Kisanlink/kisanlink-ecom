package database

import (
	"fmt"
	"log"
	"os"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// MigrationManager provides high-level migration management operations
type MigrationManager struct {
	migrator *IdempotentMigrator
	runner   *MigrationRunner
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(dbManager db.DBManager) (*MigrationManager, error) {
	migrator, err := NewIdempotentMigrator(dbManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create idempotent migrator: %w", err)
	}

	runner, err := NewMigrationRunner(dbManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration runner: %w", err)
	}

	return &MigrationManager{
		migrator: migrator,
		runner:   runner,
	}, nil
}

// MigrationCommand represents available migration commands
type MigrationCommand string

const (
	CommandMigrate  MigrationCommand = "migrate"
	CommandStatus   MigrationCommand = "status"
	CommandHistory  MigrationCommand = "history"
	CommandPlan     MigrationCommand = "plan"
	CommandValidate MigrationCommand = "validate"
	CommandCleanup  MigrationCommand = "cleanup"
	CommandDryRun   MigrationCommand = "dry-run"
	CommandRollback MigrationCommand = "rollback"
)

// ExecuteCommand executes a migration command
func (mm *MigrationManager) ExecuteCommand(command MigrationCommand, args ...string) error {
	switch command {
	case CommandMigrate:
		return mm.executeMigrate(args...)
	case CommandStatus:
		return mm.executeStatus(args...)
	case CommandHistory:
		return mm.executeHistory(args...)
	case CommandPlan:
		return mm.executePlan(args...)
	case CommandValidate:
		return mm.executeValidate(args...)
	case CommandCleanup:
		return mm.executeCleanup(args...)
	case CommandDryRun:
		return mm.executeDryRun(args...)
	case CommandRollback:
		return mm.executeRollback(args...)
	default:
		return fmt.Errorf("unknown migration command: %s", command)
	}
}

// executeMigrate runs all pending migrations
func (mm *MigrationManager) executeMigrate(args ...string) error {
	log.Println("Executing migration command...")

	// Check for safety flags
	safetyCheck := true
	for _, arg := range args {
		if arg == "--no-safety-check" {
			safetyCheck = false
			log.Println("WARNING: Safety checks disabled")
		}
	}

	mm.migrator.SetSafetyCheck(safetyCheck)
	return mm.migrator.RunMigrations()
}

// executeStatus shows current migration status
func (mm *MigrationManager) executeStatus(args ...string) error {
	log.Println("Migration Status:")
	log.Println("=================")

	// Show last applied version
	lastVersion, err := mm.runner.GetLastAppliedVersion()
	if err != nil {
		return fmt.Errorf("failed to get last applied version: %w", err)
	}

	if lastVersion == "" {
		log.Println("No migrations have been applied")
	} else {
		log.Printf("Last applied version: %s", lastVersion)
	}

	// Show pending migrations
	pending, err := mm.migrator.GetPendingMigrations()
	if err != nil {
		return fmt.Errorf("failed to get pending migrations: %w", err)
	}

	if len(pending) == 0 {
		log.Println("No pending migrations")
	} else {
		log.Printf("Pending migrations: %d", len(pending))
		for _, migration := range pending {
			log.Printf("  - %s: %s", migration.Version, migration.Description)
		}
	}

	// Show failed migrations
	failed, err := mm.runner.GetFailedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get failed migrations: %w", err)
	}

	if len(failed) > 0 {
		log.Printf("Failed migrations: %d", len(failed))
		for _, migration := range failed {
			log.Printf("  - %s: %s", migration.Version, migration.ErrorMsg)
		}
	}

	return nil
}

// executeHistory shows migration history
func (mm *MigrationManager) executeHistory(args ...string) error {
	return mm.runner.PrintMigrationStatus()
}

// executePlan shows what migrations would be executed
func (mm *MigrationManager) executePlan(args ...string) error {
	return mm.migrator.PrintMigrationPlan()
}

// executeValidate validates migration integrity
func (mm *MigrationManager) executeValidate(args ...string) error {
	log.Println("Validating migration integrity...")

	// For now, we don't have expected checksums stored externally
	// In a real implementation, these would come from a configuration file
	expectedMigrations := map[string]string{
		// These would be populated from a configuration file or build artifacts
	}

	if len(expectedMigrations) == 0 {
		log.Println("No expected migration checksums configured, skipping integrity check")
		return nil
	}

	if err := mm.runner.ValidateMigrationIntegrity(expectedMigrations); err != nil {
		return fmt.Errorf("migration integrity validation failed: %w", err)
	}

	log.Println("Migration integrity validation passed")
	return nil
}

// executeCleanup cleans up failed migrations
func (mm *MigrationManager) executeCleanup(args ...string) error {
	log.Println("Cleaning up failed migrations...")

	// Check for confirmation flag
	confirmed := false
	for _, arg := range args {
		if arg == "--confirm" {
			confirmed = true
		}
	}

	if !confirmed {
		log.Println("This operation will remove all failed migration records.")
		log.Println("Use --confirm flag to proceed.")
		return nil
	}

	return mm.runner.CleanupFailedMigrations()
}

// executeDryRun runs migrations in dry-run mode
func (mm *MigrationManager) executeDryRun(args ...string) error {
	log.Println("Executing dry-run migration...")

	mm.migrator.SetDryRun(true)
	defer mm.migrator.SetDryRun(false)

	return mm.migrator.RunMigrations()
}

// executeRollback attempts to rollback a migration
func (mm *MigrationManager) executeRollback(args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("rollback command requires a version argument")
	}

	version := args[0]
	log.Printf("Attempting to rollback migration %s...", version)

	return mm.migrator.RollbackMigration(version)
}

// RunMigrationCLI provides a simple CLI interface for migration management
func RunMigrationCLI(dbManager db.DBManager, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("migration command required")
	}

	manager, err := NewMigrationManager(dbManager)
	if err != nil {
		return fmt.Errorf("failed to create migration manager: %w", err)
	}

	command := MigrationCommand(args[0])
	commandArgs := args[1:]

	return manager.ExecuteCommand(command, commandArgs...)
}

// PrintUsage prints usage information for the migration CLI
func PrintUsage() {
	fmt.Println("Migration Management CLI")
	fmt.Println("========================")
	fmt.Println()
	fmt.Println("Usage: migrate <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  migrate     Run all pending migrations")
	fmt.Println("  status      Show current migration status")
	fmt.Println("  history     Show migration history")
	fmt.Println("  plan        Show migration execution plan")
	fmt.Println("  validate    Validate migration integrity")
	fmt.Println("  cleanup     Clean up failed migrations (requires --confirm)")
	fmt.Println("  dry-run     Run migrations in dry-run mode")
	fmt.Println("  rollback    Rollback a specific migration (requires version)")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --no-safety-check    Disable safety checks (migrate command)")
	fmt.Println("  --confirm           Confirm destructive operations (cleanup command)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  migrate migrate")
	fmt.Println("  migrate status")
	fmt.Println("  migrate dry-run")
	fmt.Println("  migrate cleanup --confirm")
	fmt.Println("  migrate rollback v1.0.0")
}

// HandleMigrationCLI handles migration CLI commands with proper error handling
func HandleMigrationCLI(dbManager db.DBManager) {
	args := os.Args[1:]

	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		PrintUsage()
		return
	}

	if err := RunMigrationCLI(dbManager, args); err != nil {
		log.Printf("Migration command failed: %v", err)
		os.Exit(1)
	}
}
