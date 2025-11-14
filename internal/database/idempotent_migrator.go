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

// MigrationDefinition represents a single migration with metadata
type MigrationDefinition struct {
	Version     string
	Description string
	Models      []interface{}
	PreHook     func(*gorm.DB) error // Optional pre-migration hook
	PostHook    func(*gorm.DB) error // Optional post-migration hook
}

// IdempotentMigrator handles safe, idempotent database migrations
type IdempotentMigrator struct {
	runner      *MigrationRunner
	dbManager   db.DBManager
	db          *gorm.DB
	migrations  []MigrationDefinition
	dryRun      bool
	safetyCheck bool
}

// NewIdempotentMigrator creates a new idempotent migrator
func NewIdempotentMigrator(dbManager db.DBManager) (*IdempotentMigrator, error) {
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

	migrator := &IdempotentMigrator{
		runner:      runner,
		dbManager:   dbManager,
		db:          database,
		migrations:  []MigrationDefinition{},
		dryRun:      false,
		safetyCheck: true,
	}

	// Register default migrations
	migrator.registerDefaultMigrations()

	return migrator, nil
}

// SetDryRun enables or disables dry run mode
func (im *IdempotentMigrator) SetDryRun(dryRun bool) {
	im.dryRun = dryRun
}

// SetSafetyCheck enables or disables safety checks
func (im *IdempotentMigrator) SetSafetyCheck(safetyCheck bool) {
	im.safetyCheck = safetyCheck
}

// registerDefaultMigrations registers the default migration definitions
func (im *IdempotentMigrator) registerDefaultMigrations() {
	// Migration v1.0.0 - Initial catalog models
	im.AddMigration(MigrationDefinition{
		Version:     "v1.0.0",
		Description: "Initial catalog models migration",
		Models:      models.ModelsByPriority(),
	})

	// Migration v1.1.0 - Additional indexes and constraints
	im.AddMigration(MigrationDefinition{
		Version:     "v1.1.0",
		Description: "Performance indexes and constraints",
		Models:      []interface{}{}, // No new models, just indexes
		PostHook:    im.createPerformanceIndexes,
	})

	// Migration v1.2.0 - Catalog-specific performance indexes
	im.AddMigration(MigrationDefinition{
		Version:     "v1.2.0",
		Description: "Catalog-specific performance indexes for search and filtering",
		Models:      []interface{}{}, // No new models, just indexes
		PostHook:    im.createCatalogIndexes,
	})
}

// AddMigration adds a new migration definition
func (im *IdempotentMigrator) AddMigration(migration MigrationDefinition) {
	im.migrations = append(im.migrations, migration)
}

// RunMigrations executes all pending migrations
func (im *IdempotentMigrator) RunMigrations() error {
	log.Printf("Starting idempotent migration process (dry-run: %v, safety-check: %v)", im.dryRun, im.safetyCheck)

	// Sort migrations by version
	sort.Slice(im.migrations, func(i, j int) bool {
		return im.migrations[i].Version < im.migrations[j].Version
	})

	// Perform safety checks
	if im.safetyCheck {
		if err := im.performSafetyChecks(); err != nil {
			return fmt.Errorf("safety checks failed: %w", err)
		}
	}

	// Execute each migration
	for _, migration := range im.migrations {
		if err := im.executeMigration(migration); err != nil {
			return fmt.Errorf("migration %s failed: %w", migration.Version, err)
		}
	}

	log.Printf("All migrations completed successfully")
	return nil
}

// executeMigration executes a single migration with enhanced conflict resolution
func (im *IdempotentMigrator) executeMigration(migration MigrationDefinition) error {
	// Check if migration is already applied
	applied, err := im.runner.IsMigrationApplied(migration.Version)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if applied {
		log.Printf("Migration %s already applied, skipping", migration.Version)
		return nil
	}

	log.Printf("Applying migration %s: %s", migration.Version, migration.Description)

	// Calculate checksum for the migration
	checksum := im.calculateMigrationChecksum(migration)

	// Start timing
	startTime := time.Now()

	// Clean up any failed migration attempts for this version first
	if err := im.cleanupFailedMigrationAttempts(migration.Version); err != nil {
		log.Printf("Warning: Failed to cleanup previous attempts: %v", err)
	}

	// Execute migration with enhanced error handling
	err = im.executeWithRetry(migration)

	duration := time.Since(startTime)

	// Record migration result
	var errorMsg string
	success := err == nil

	if err != nil {
		errorMsg = err.Error()
		log.Printf("Migration %s failed after %v: %v", migration.Version, duration, err)
	} else {
		log.Printf("Migration %s completed successfully in %v", migration.Version, duration)
	}

	// Record the migration attempt (even in dry-run mode for tracking)
	if !im.dryRun {
		if recordErr := im.runner.RecordMigration(migration.Version, migration.Description, checksum, success, duration, errorMsg); recordErr != nil {
			log.Printf("Warning: Failed to record migration result: %v", recordErr)
		}
	}

	return err
}

// executeWithRetry executes migration with retry logic for transient errors
func (im *IdempotentMigrator) executeWithRetry(migration MigrationDefinition) error {
	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			log.Printf("Retry attempt %d/%d for migration %s", attempt, maxRetries, migration.Version)
			time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
		}

		// Clear any cached plans that might be causing conflicts
		if err := im.clearCachedPlans(); err != nil {
			log.Printf("Warning: Failed to clear cached plans: %v", err)
		}

		// Execute the migration
		err := im.executeMigrationAttempt(migration)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if this is a retryable error
		if !im.isRetryableError(err) {
			log.Printf("Non-retryable error encountered: %v", err)
			break
		}

		log.Printf("Retryable error on attempt %d: %v", attempt, err)
	}

	return fmt.Errorf("migration failed after %d attempts: %w", maxRetries, lastErr)
}

// executeMigrationAttempt executes a single migration attempt
func (im *IdempotentMigrator) executeMigrationAttempt(migration MigrationDefinition) error {
	// Use a fresh transaction for each attempt
	return im.db.Transaction(func(tx *gorm.DB) error {
		// Execute pre-hook if present
		if migration.PreHook != nil {
			log.Printf("Executing pre-hook for migration %s", migration.Version)
			if err := migration.PreHook(tx); err != nil {
				return fmt.Errorf("pre-hook failed: %w", err)
			}
		}

		// Run AutoMigrate for models with enhanced error handling
		if len(migration.Models) > 0 {
			log.Printf("Auto-migrating %d models for migration %s", len(migration.Models), migration.Version)

			if im.dryRun {
				log.Printf("DRY RUN: Would migrate models: %v", im.getModelNames(migration.Models))
			} else {
				// Use GORM's batch AutoMigrate - it handles dependencies and conflicts better
				if err := tx.AutoMigrate(migration.Models...); err != nil {
					return fmt.Errorf("auto-migrate failed: %w", err)
				}
			}
		}

		// Execute post-hook if present
		if migration.PostHook != nil {
			log.Printf("Executing post-hook for migration %s", migration.Version)
			if err := migration.PostHook(tx); err != nil {
				return fmt.Errorf("post-hook failed: %w", err)
			}
		}

		return nil
	})
}

// migrateModelSafely migrates a single model using only GORM AutoMigrate
func (im *IdempotentMigrator) migrateModelSafely(tx *gorm.DB, model interface{}, modelName string) error {
	// Use GORM's AutoMigrate directly - it's designed to handle schema changes safely
	log.Printf("Auto-migrating model: %s", modelName)
	return tx.AutoMigrate(model)
}

// Removed complex migration functions - using pure GORM AutoMigrate instead

// Removed - using pure GORM AutoMigrate

// migrateInventoryLotsTable handles inventory_lots table migration
func (im *IdempotentMigrator) migrateInventoryLotsTable(tx *gorm.DB, currentColumns map[string]ColumnInfo) error {
	log.Printf("Performing safe migration for inventory_lots table")

	// Handle catalog_item_id column changes
	if col, exists := currentColumns["catalog_item_id"]; exists {
		if !col.IsNullable {
			// Make catalog_item_id nullable first
			if err := tx.Exec("ALTER TABLE inventory_lots ALTER COLUMN catalog_item_id DROP NOT NULL").Error; err != nil {
				log.Printf("Warning: Failed to make catalog_item_id nullable: %v", err)
			}
		}
	}

	// Add missing columns if they don't exist
	missingColumns := []string{
		"variant_id", "serial_number", "quantity", "reserved_quantity", "available_quantity", "unit_of_measure",
	}

	for _, colName := range missingColumns {
		if _, exists := currentColumns[colName]; !exists {
			if err := im.addColumnSafely(tx, "inventory_lots", colName); err != nil {
				log.Printf("Warning: Failed to add column %s: %v", colName, err)
			}
		}
	}

	return nil
}

// migrateCatalogItemsTable handles catalog_items table migration
func (im *IdempotentMigrator) migrateCatalogItemsTable(tx *gorm.DB, currentColumns map[string]ColumnInfo) error {
	log.Printf("Performing safe migration for catalog_items table")

	// Handle term column if it exists and has null values
	if _, exists := currentColumns["term"]; exists {
		// Update null values before making it NOT NULL
		if err := tx.Exec("UPDATE catalog_items SET term = '' WHERE term IS NULL").Error; err != nil {
			log.Printf("Warning: Failed to update null term values: %v", err)
		}
	}

	// Handle structural changes from old to new catalog_items schema
	if err := im.migrateCatalogItemsStructure(tx, currentColumns); err != nil {
		return fmt.Errorf("failed to migrate catalog_items structure: %w", err)
	}

	return nil
}

// migrateCatalogItemsStructure handles the structural changes in catalog_items table
func (im *IdempotentMigrator) migrateCatalogItemsStructure(tx *gorm.DB, currentColumns map[string]ColumnInfo) error {
	log.Printf("Migrating catalog_items table structure...")

	// Check if we need migration
	if !im.needsCatalogItemsMigration(currentColumns) {
		return nil
	}

	log.Printf("Detected old catalog_items structure, performing structural migration...")

	// Perform migration steps
	im.addCatalogItemsColumns(tx, currentColumns)
	im.removeCatalogItemsOldColumns(tx, currentColumns)
	im.addCatalogItemsConstraints(tx)
	im.addCatalogItemsIndexes(tx)

	log.Printf("Structural migration completed")
	return nil
}

// Legacy migration functions - kept for reference but not currently used

// needsCatalogItemsMigration checks if catalog_items needs migration
//
//nolint:unused
func (im *IdempotentMigrator) needsCatalogItemsMigration(currentColumns map[string]ColumnInfo) bool {
	_, hasOldCategory := currentColumns["category"]
	_, hasNewCategoryID := currentColumns["category_id"]
	return hasOldCategory && !hasNewCategoryID
}

// addCatalogItemsColumns adds new columns to catalog_items table
//
//nolint:unused
func (im *IdempotentMigrator) addCatalogItemsColumns(tx *gorm.DB, currentColumns map[string]ColumnInfo) {
	newColumns := map[string]string{
		"category_id":     "ALTER TABLE catalog_items ADD COLUMN category_id varchar(255)",
		"vendor_id":       "ALTER TABLE catalog_items ADD COLUMN vendor_id varchar(255)",
		"version":         "ALTER TABLE catalog_items ADD COLUMN version bigint NOT NULL DEFAULT 1",
		"weight":          "ALTER TABLE catalog_items ADD COLUMN weight decimal(10,3)",
		"dimensions":      "ALTER TABLE catalog_items ADD COLUMN dimensions jsonb",
		"perishable":      "ALTER TABLE catalog_items ADD COLUMN perishable boolean NOT NULL DEFAULT false",
		"shelf_life_days": "ALTER TABLE catalog_items ADD COLUMN shelf_life_days bigint",
	}

	for colName, sql := range newColumns {
		if _, exists := currentColumns[colName]; !exists {
			log.Printf("Adding column: %s", colName)
			if err := tx.Exec(sql).Error; err != nil {
				log.Printf("Warning: Failed to add column %s: %v", colName, err)
			}
		}
	}
}

// removeCatalogItemsOldColumns removes obsolete columns from catalog_items table
//
//nolint:unused
func (im *IdempotentMigrator) removeCatalogItemsOldColumns(tx *gorm.DB, currentColumns map[string]ColumnInfo) {
	oldColumns := []string{"skill_level", "hourly_rate"}
	for _, colName := range oldColumns {
		if _, exists := currentColumns[colName]; exists {
			log.Printf("Removing old column: %s", colName)
			sql := fmt.Sprintf("ALTER TABLE catalog_items DROP COLUMN IF EXISTS %s", colName)
			if err := tx.Exec(sql).Error; err != nil {
				log.Printf("Warning: Failed to drop column %s: %v", colName, err)
			}
		}
	}
}

// addCatalogItemsConstraints adds constraints to catalog_items table
//
//nolint:unused
func (im *IdempotentMigrator) addCatalogItemsConstraints(tx *gorm.DB) {
	constraints := []string{
		"ALTER TABLE catalog_items ADD CONSTRAINT chk_catalog_items_base_price CHECK (base_price >= 0)",
		"ALTER TABLE catalog_items ADD CONSTRAINT chk_catalog_items_weight CHECK (weight IS NULL OR weight >= 0)",
		"ALTER TABLE catalog_items ADD CONSTRAINT chk_catalog_items_shelf_life_days CHECK (shelf_life_days IS NULL OR shelf_life_days > 0)",
	}

	for _, constraintSQL := range constraints {
		constraintName := im.extractConstraintName(constraintSQL)
		if constraintName != "" && !im.constraintExists(tx, "catalog_items", constraintName) {
			log.Printf("Adding constraint: %s", constraintName)
			if err := tx.Exec(constraintSQL).Error; err != nil {
				log.Printf("Warning: Failed to add constraint %s: %v", constraintName, err)
			}
		}
	}
}

// addCatalogItemsIndexes adds indexes to catalog_items table
//
//nolint:unused
func (im *IdempotentMigrator) addCatalogItemsIndexes(tx *gorm.DB) {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_tenant ON catalog_items (organization_id)",
		"CREATE INDEX IF NOT EXISTS idx_tenant_type_status ON catalog_items (organization_id, item_type, is_active)",
		"CREATE INDEX IF NOT EXISTS idx_tenant_category ON catalog_items (organization_id, category_id)",
		"CREATE INDEX IF NOT EXISTS idx_tenant_vendor ON catalog_items (organization_id, vendor_id)",
		"CREATE INDEX IF NOT EXISTS idx_catalog_items_search_name ON catalog_items (name)",
		"CREATE INDEX IF NOT EXISTS idx_search_desc ON catalog_items (description)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_org_sku ON catalog_items (sku) WHERE sku IS NOT NULL AND sku != ''",
		"CREATE INDEX IF NOT EXISTS idx_tags ON catalog_items (tags)",
		"CREATE INDEX IF NOT EXISTS idx_attributes ON catalog_items (attributes)",
		"CREATE INDEX IF NOT EXISTS idx_catalog_items_deleted ON catalog_items (deleted_at)",
	}

	for _, indexSQL := range indexes {
		if err := tx.Exec(indexSQL).Error; err != nil {
			log.Printf("Warning: Failed to create index: %v", err)
		}
	}
}

// constraintExists checks if a constraint exists on a table
//
//nolint:unused
func (im *IdempotentMigrator) constraintExists(tx *gorm.DB, tableName, constraintName string) bool {
	var count int64
	checkSQL := "SELECT COUNT(*) FROM INFORMATION_SCHEMA.table_constraints WHERE table_schema = CURRENT_SCHEMA() AND table_name = ? AND constraint_name = ?"
	err := tx.Raw(checkSQL, tableName, constraintName).Scan(&count).Error
	return err == nil && count > 0
}

// addColumnSafely adds a column with proper error handling
func (im *IdempotentMigrator) addColumnSafely(tx *gorm.DB, tableName, columnName string) error {
	var sql string

	switch columnName {
	case "variant_id":
		sql = "ALTER TABLE " + tableName + " ADD COLUMN variant_id varchar(255)"
	case "serial_number":
		sql = "ALTER TABLE " + tableName + " ADD COLUMN serial_number varchar(100)"
	case "quantity":
		sql = "ALTER TABLE " + tableName + " ADD COLUMN quantity decimal(12,3) NOT NULL DEFAULT 0"
	case "reserved_quantity":
		sql = "ALTER TABLE " + tableName + " ADD COLUMN reserved_quantity decimal(12,3) NOT NULL DEFAULT 0"
	case "available_quantity":
		sql = "ALTER TABLE " + tableName + " ADD COLUMN available_quantity decimal(12,3) NOT NULL DEFAULT 0"
	case "unit_of_measure":
		sql = "ALTER TABLE " + tableName + " ADD COLUMN unit_of_measure varchar(50) NOT NULL DEFAULT ''"
	default:
		return fmt.Errorf("unknown column: %s", columnName)
	}

	return tx.Exec(sql).Error
}

// ColumnInfo represents database column information
type ColumnInfo struct {
	Name       string
	DataType   string
	IsNullable bool
	Default    *string
}

// getTableColumns retrieves current table column information
func (im *IdempotentMigrator) getTableColumns(tx *gorm.DB, tableName string) (map[string]ColumnInfo, error) {
	var columns []ColumnInfo
	query := `
		SELECT
			column_name as name,
			data_type,
			is_nullable = 'YES' as is_nullable,
			column_default as default
		FROM information_schema.columns
		WHERE table_name = ? AND table_schema = CURRENT_SCHEMA()
		ORDER BY ordinal_position
	`

	if err := tx.Raw(query, tableName).Scan(&columns).Error; err != nil {
		return nil, err
	}

	result := make(map[string]ColumnInfo)
	for _, col := range columns {
		result[col.Name] = col
	}

	return result, nil
}

// tableExists checks if a table exists
func (im *IdempotentMigrator) tableExists(tx *gorm.DB, tableName string) (bool, error) {
	var count int64
	query := `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_name = ? AND table_schema = CURRENT_SCHEMA()
	`

	if err := tx.Raw(query, tableName).Scan(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// getTableName extracts table name from model
func (im *IdempotentMigrator) getTableName(model interface{}) string {
	if tabler, ok := model.(interface{ TableName() string }); ok {
		return tabler.TableName()
	}

	// Fallback to GORM's default naming
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return strings.ToLower(t.Name()) + "s"
}

// getModelName extracts model name from interface
func (im *IdempotentMigrator) getModelName(model interface{}) string {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

// clearCachedPlans clears PostgreSQL cached plans that might cause conflicts
func (im *IdempotentMigrator) clearCachedPlans() error {
	// Discard all cached plans
	if err := im.db.Exec("DISCARD PLANS").Error; err != nil {
		return err
	}

	// Also discard temporary objects
	if err := im.db.Exec("DISCARD TEMP").Error; err != nil {
		return err
	}

	return nil
}

// isRetryableError determines if an error is worth retrying
func (im *IdempotentMigrator) isRetryableError(err error) bool {
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

// cleanupFailedMigrationAttempts removes failed attempts for a specific version
func (im *IdempotentMigrator) cleanupFailedMigrationAttempts(version string) error {
	result := im.db.Where("version = ? AND success = ?", version, false).Delete(&SchemaMigration{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected > 0 {
		log.Printf("Cleaned up %d failed attempts for migration %s", result.RowsAffected, version)
	}

	return nil
}

// performSafetyChecks performs pre-migration safety checks
func (im *IdempotentMigrator) performSafetyChecks() error {
	log.Println("Performing safety checks...")

	// Check database connectivity
	if err := im.checkDatabaseConnectivity(); err != nil {
		return fmt.Errorf("database connectivity check failed: %w", err)
	}

	// Check for conflicting migrations
	if err := im.checkForConflictingMigrations(); err != nil {
		return fmt.Errorf("conflicting migrations check failed: %w", err)
	}

	// Check for failed migrations
	if err := im.checkForFailedMigrations(); err != nil {
		return fmt.Errorf("failed migrations check failed: %w", err)
	}

	// Check database locks
	if err := im.checkDatabaseLocks(); err != nil {
		return fmt.Errorf("database locks check failed: %w", err)
	}

	log.Println("All safety checks passed")
	return nil
}

// checkDatabaseConnectivity verifies database connection
func (im *IdempotentMigrator) checkDatabaseConnectivity() error {
	var result int
	if err := im.db.Raw("SELECT 1").Scan(&result).Error; err != nil {
		return fmt.Errorf("database connectivity test failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected connectivity test result: %d", result)
	}

	return nil
}

// checkForConflictingMigrations checks for version conflicts
func (im *IdempotentMigrator) checkForConflictingMigrations() error {
	versions := make(map[string]bool)

	for _, migration := range im.migrations {
		if versions[migration.Version] {
			return fmt.Errorf("duplicate migration version found: %s", migration.Version)
		}
		versions[migration.Version] = true
	}

	return nil
}

// checkForFailedMigrations checks for previous failed migrations
func (im *IdempotentMigrator) checkForFailedMigrations() error {
	failedMigrations, err := im.runner.GetFailedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get failed migrations: %w", err)
	}

	if len(failedMigrations) > 0 {
		log.Printf("Found %d failed migration(s). Cleaning them up automatically.", len(failedMigrations))
		for _, failed := range failedMigrations {
			log.Printf("  - %s: %s", failed.Version, failed.ErrorMsg)
		}

		// Auto-cleanup failed migrations to prevent conflicts
		if err := im.runner.CleanupFailedMigrations(); err != nil {
			log.Printf("Warning: Failed to cleanup failed migrations: %v", err)
		} else {
			log.Printf("Successfully cleaned up failed migrations")
		}
	}

	return nil
}

// checkDatabaseLocks checks for active database locks that might interfere
func (im *IdempotentMigrator) checkDatabaseLocks() error {
	// Check for active locks on tables we'll be modifying
	var lockCount int64
	query := `
		SELECT COUNT(*)
		FROM pg_locks l
		JOIN pg_class c ON l.relation = c.oid
		WHERE c.relkind = 'r'
		AND l.mode IN ('AccessExclusiveLock', 'ShareUpdateExclusiveLock')
	`

	if err := im.db.Raw(query).Scan(&lockCount).Error; err != nil {
		// If the query fails (e.g., not PostgreSQL), just log a warning
		log.Printf("Warning: Could not check database locks: %v", err)
		return nil
	}

	if lockCount > 0 {
		return fmt.Errorf("found %d active exclusive locks on tables", lockCount)
	}

	return nil
}

// calculateMigrationChecksum calculates a checksum for the migration
func (im *IdempotentMigrator) calculateMigrationChecksum(migration MigrationDefinition) string {
	hasher := sha256.New()

	// Include version and description
	hasher.Write([]byte(migration.Version))
	hasher.Write([]byte(migration.Description))

	// Include model names and their structure
	modelNames := im.getModelNames(migration.Models)
	sort.Strings(modelNames)
	hasher.Write([]byte(strings.Join(modelNames, ",")))

	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// getModelNames extracts model names from interfaces
func (im *IdempotentMigrator) getModelNames(models []interface{}) []string {
	names := make([]string, len(models))
	for i, model := range models {
		t := reflect.TypeOf(model)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		names[i] = t.Name()
	}
	return names
}

// createPerformanceIndexes creates additional performance indexes
func (im *IdempotentMigrator) createPerformanceIndexes(tx *gorm.DB) error {
	log.Println("Creating performance indexes...")

	// Define performance indexes for common query patterns
	indexes := []string{
		// Catalog Items - Organization and type filtering
		"CREATE INDEX IF NOT EXISTS idx_catalog_items_org_type_active ON catalog_items(organization_id, item_type, is_active)",
		"CREATE INDEX IF NOT EXISTS idx_catalog_items_category_subcategory ON catalog_items(category, subcategory)",
		"CREATE INDEX IF NOT EXISTS idx_catalog_items_visibility_active ON catalog_items(visibility, is_active)",
		"CREATE INDEX IF NOT EXISTS idx_catalog_items_sku_org ON catalog_items(sku, organization_id)",

		// Pricing - Catalog item and currency queries
		"CREATE INDEX IF NOT EXISTS idx_prices_catalog_currency ON prices(catalog_item_id, currency)",
		"CREATE INDEX IF NOT EXISTS idx_prices_org_currency ON prices(organization_id, currency)",
		"CREATE INDEX IF NOT EXISTS idx_prices_base_price ON prices(base_price)",

		// Price Tiers - Volume-based pricing
		"CREATE INDEX IF NOT EXISTS idx_price_tiers_price_id ON price_tiers(price_id)",
		"CREATE INDEX IF NOT EXISTS idx_price_tiers_min_quantity ON price_tiers(min_quantity)",

		// Price Rules - Rule-based pricing
		"CREATE INDEX IF NOT EXISTS idx_price_rules_price_id ON price_rules(price_id)",
		"CREATE INDEX IF NOT EXISTS idx_price_rules_rule_type ON price_rules(rule_type)",

		// Inventory Lots - Product and organization tracking
		"CREATE INDEX IF NOT EXISTS idx_inventory_lots_product_org ON inventory_lots(product_id, organization_id)",
		"CREATE INDEX IF NOT EXISTS idx_inventory_lots_lot_number ON inventory_lots(lot_number)",
		"CREATE INDEX IF NOT EXISTS idx_inventory_lots_expiry_date ON inventory_lots(expiry_date)",

		// Media - Catalog item relationships
		"CREATE INDEX IF NOT EXISTS idx_media_catalog_item_type ON media(catalog_item_id, media_type)",
		"CREATE INDEX IF NOT EXISTS idx_media_org_type ON media(organization_id, media_type)",

		// Actors - Organization relationships
		"CREATE INDEX IF NOT EXISTS idx_vendors_org_status ON vendors(organization_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_customers_org_status ON customers(organization_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_collaborators_org_role ON collaborators(organization_id, role)",

		// Audit logs - Tracking and compliance
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_entity_action ON audit_logs(entity_type, action)",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_entity_id ON audit_logs(entity_id)",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_user_timestamp ON audit_logs(user_id, timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_org_timestamp ON audit_logs(organization_id, timestamp)",
	}

	// Execute each index creation statement
	successCount := 0
	for i, indexSQL := range indexes {
		if im.dryRun {
			log.Printf("DRY RUN: Would create index: %s", indexSQL)
			successCount++
		} else {
			if err := tx.Exec(indexSQL).Error; err != nil {
				log.Printf("Warning: Failed to create index %d: %v", i+1, err)
				log.Printf("SQL: %s", indexSQL)
				// Continue with other indexes instead of failing completely
			} else {
				successCount++
			}
		}
	}

	log.Printf("Successfully created %d out of %d performance indexes", successCount, len(indexes))
	return nil
}

// createCatalogIndexes creates catalog-specific performance indexes
func (im *IdempotentMigrator) createCatalogIndexes(tx *gorm.DB) error {
	log.Println("Creating catalog-specific performance indexes...")

	// Define catalog-specific performance indexes
	indexes := []string{
		// Composite indexes for common query patterns (tenant_id, type, status, updated_at DESC)
		`CREATE INDEX IF NOT EXISTS idx_catalog_items_tenant_type_status_updated
		 ON catalog_items(organization_id, item_type, is_active, updated_at DESC)`,

		`CREATE INDEX IF NOT EXISTS idx_catalog_items_tenant_category_active
		 ON catalog_items(organization_id, category_id, is_active)`,

		`CREATE INDEX IF NOT EXISTS idx_catalog_items_tenant_vendor_active
		 ON catalog_items(organization_id, vendor_id, is_active)`,

		`CREATE INDEX IF NOT EXISTS idx_catalog_items_tenant_visibility_active
		 ON catalog_items(organization_id, visibility, is_active)`,

		// Full-text search indexes using trigram and GIN
		`CREATE EXTENSION IF NOT EXISTS pg_trgm`,
		`CREATE EXTENSION IF NOT EXISTS btree_gin`,

		// Trigram indexes for name and description fields
		`CREATE INDEX IF NOT EXISTS idx_catalog_items_name_trgm
		 ON catalog_items USING gin(name gin_trgm_ops)`,

		`CREATE INDEX IF NOT EXISTS idx_catalog_items_description_trgm
		 ON catalog_items USING gin(description gin_trgm_ops)`,

		// Full-text search indexes for name and description
		`CREATE INDEX IF NOT EXISTS idx_catalog_items_name_fts
		 ON catalog_items USING gin(to_tsvector('english', name))`,

		`CREATE INDEX IF NOT EXISTS idx_catalog_items_description_fts
		 ON catalog_items USING gin(to_tsvector('english', description))`,

		// Combined full-text search index for name and description
		`CREATE INDEX IF NOT EXISTS idx_catalog_items_combined_fts
		 ON catalog_items USING gin(to_tsvector('english', coalesce(name, '') || ' ' || coalesce(description, '')))`,

		// JSONB indexes for attributes field
		`CREATE INDEX IF NOT EXISTS idx_catalog_items_attributes_gin
		 ON catalog_items USING gin(attributes)`,

		// Specific JSONB path indexes for common attribute queries
		`CREATE INDEX IF NOT EXISTS idx_catalog_items_attributes_sku
		 ON catalog_items USING gin((attributes->'sku'))`,

		`CREATE INDEX IF NOT EXISTS idx_catalog_items_attributes_brand
		 ON catalog_items USING gin((attributes->'brand'))`,

		`CREATE INDEX IF NOT EXISTS idx_catalog_items_attributes_skills
		 ON catalog_items USING gin((attributes->'skills'))`,

		`CREATE INDEX IF NOT EXISTS idx_catalog_items_attributes_certification
		 ON catalog_items USING gin((attributes->'certification'))`,

		// Array indexes for tags field
		`CREATE INDEX IF NOT EXISTS idx_catalog_items_tags_gin
		 ON catalog_items USING gin(tags)`,

		// Specific tag search optimization
		`CREATE INDEX IF NOT EXISTS idx_catalog_items_tags_array
		 ON catalog_items USING gin(tags array_ops)`,

		// Price range queries
		`CREATE INDEX IF NOT EXISTS idx_catalog_items_price_range
		 ON catalog_items(base_price, currency, is_active)`,

		`CREATE INDEX IF NOT EXISTS idx_catalog_items_tenant_price_range
		 ON catalog_items(organization_id, base_price, currency, is_active)`,

		// Category hierarchy indexes
		`CREATE INDEX IF NOT EXISTS idx_categories_tenant_parent_active
		 ON categories(organization_id, parent_id, is_active)`,

		`CREATE INDEX IF NOT EXISTS idx_categories_path_active
		 ON categories(path, is_active)`,

		`CREATE INDEX IF NOT EXISTS idx_categories_level_active
		 ON categories(level, is_active)`,

		// Category name search
		`CREATE INDEX IF NOT EXISTS idx_categories_name_trgm
		 ON categories USING gin(name gin_trgm_ops)`,

		// Variant indexes
		`CREATE INDEX IF NOT EXISTS idx_variants_tenant_catalog_active
		 ON variants(organization_id, catalog_item_id, is_active)`,

		`CREATE INDEX IF NOT EXISTS idx_variants_sku_tenant
		 ON variants(sku, organization_id)`,

		`CREATE INDEX IF NOT EXISTS idx_variants_attributes_gin
		 ON variants USING gin(attributes)`,

		`CREATE INDEX IF NOT EXISTS idx_variants_price_range
		 ON variants(price, currency, is_active)`,

		// Availability indexes (if availability table exists)
		`CREATE INDEX IF NOT EXISTS idx_availability_catalog_item_active
		 ON availability(catalog_item_id, is_available)
		 WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'availability')`,

		// Media indexes (if media table exists)
		`CREATE INDEX IF NOT EXISTS idx_media_catalog_item_type
		 ON media(entity_id, media_type)
		 WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'media')`,

		// Inventory lot indexes for products
		`CREATE INDEX IF NOT EXISTS idx_inventory_lots_product_org_active
		 ON inventory_lots(product_id, organization_id, quantity > 0)`,

		`CREATE INDEX IF NOT EXISTS idx_inventory_lots_expiry_active
		 ON inventory_lots(expiry_date, quantity > 0)
		 WHERE expiry_date IS NOT NULL`,

		// SLA indexes for services
		`CREATE INDEX IF NOT EXISTS idx_slas_catalog_item_active
		 ON slas(catalog_item_id, is_active)
		 WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'slas')`,

		// Composite indexes for complex filtering scenarios
		`CREATE INDEX IF NOT EXISTS idx_catalog_items_multi_filter
		 ON catalog_items(organization_id, item_type, category_id, is_active, visibility, updated_at DESC)`,

		// Partial indexes for active items only (more efficient for common queries)
		`CREATE INDEX IF NOT EXISTS idx_catalog_items_active_tenant_type
		 ON catalog_items(organization_id, item_type, updated_at DESC)
		 WHERE is_active = true`,

		`CREATE INDEX IF NOT EXISTS idx_catalog_items_active_public
		 ON catalog_items(item_type, category_id, updated_at DESC)
		 WHERE is_active = true AND visibility = 'PUBLIC'`,

		// Covering indexes for read-heavy queries
		`CREATE INDEX IF NOT EXISTS idx_catalog_items_list_covering
		 ON catalog_items(organization_id, is_active, item_type)
		 INCLUDE (name, base_price, currency, visibility, updated_at)`,
	}

	// Execute each index creation statement
	successCount := 0
	for i, indexSQL := range indexes {
		if im.dryRun {
			log.Printf("DRY RUN: Would create catalog index: %s", indexSQL)
			successCount++
		} else {
			if err := tx.Exec(indexSQL).Error; err != nil {
				log.Printf("Warning: Failed to create catalog index %d: %v", i+1, err)
				log.Printf("SQL: %s", indexSQL)
				// Continue with other indexes instead of failing completely
			} else {
				successCount++
			}
		}
	}

	log.Printf("Successfully created %d out of %d catalog performance indexes", successCount, len(indexes))
	return nil
}

// RollbackMigration attempts to rollback a specific migration (use with extreme caution)
func (im *IdempotentMigrator) RollbackMigration(version string) error {
	if !im.safetyCheck {
		return fmt.Errorf("rollback operations require safety checks to be enabled")
	}

	log.Printf("WARNING: Attempting to rollback migration %s", version)
	log.Printf("This operation is potentially destructive and should only be used in development")

	// Check if migration was applied
	applied, err := im.runner.IsMigrationApplied(version)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if !applied {
		return fmt.Errorf("migration %s was not applied, cannot rollback", version)
	}

	// For now, we don't implement automatic rollback as it's complex and dangerous
	// Instead, we provide guidance
	return fmt.Errorf("automatic rollback not implemented. Please manually revert changes for migration %s", version)
}

// GetPendingMigrations returns migrations that haven't been applied yet
func (im *IdempotentMigrator) GetPendingMigrations() ([]MigrationDefinition, error) {
	var pending []MigrationDefinition

	for _, migration := range im.migrations {
		applied, err := im.runner.IsMigrationApplied(migration.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to check migration status for %s: %w", migration.Version, err)
		}

		if !applied {
			pending = append(pending, migration)
		}
	}

	return pending, nil
}

// PrintMigrationPlan prints what migrations would be executed
func (im *IdempotentMigrator) PrintMigrationPlan() error {
	pending, err := im.GetPendingMigrations()
	if err != nil {
		return err
	}

	if len(pending) == 0 {
		log.Println("No pending migrations")
		return nil
	}

	log.Println("Migration Plan:")
	log.Println("===============")
	for _, migration := range pending {
		log.Printf("Version: %s", migration.Version)
		log.Printf("  Description: %s", migration.Description)
		log.Printf("  Models: %d", len(migration.Models))
		if len(migration.Models) > 0 {
			log.Printf("  Model Names: %v", im.getModelNames(migration.Models))
		}
		log.Printf("  Has Pre-Hook: %v", migration.PreHook != nil)
		log.Printf("  Has Post-Hook: %v", migration.PostHook != nil)
		log.Println()
	}

	return nil
}

// RecoverFromInconsistentState attempts to recover from database schema inconsistencies
func (im *IdempotentMigrator) RecoverFromInconsistentState() error {
	log.Println("Attempting to recover from inconsistent database state...")

	// Step 1: Clear all cached plans and temporary objects
	if err := im.clearCachedPlans(); err != nil {
		log.Printf("Warning: Failed to clear cached plans: %v", err)
	}

	// Step 2: Clean up failed migrations
	if err := im.runner.CleanupFailedMigrations(); err != nil {
		log.Printf("Warning: Failed to cleanup failed migrations: %v", err)
	}

	// Step 3: Terminate any conflicting connections (be careful with this)
	if err := im.terminateConflictingConnections(); err != nil {
		log.Printf("Warning: Failed to terminate conflicting connections: %v", err)
	}

	// Step 4: Vacuum and analyze to clean up the database
	if err := im.performDatabaseMaintenance(); err != nil {
		log.Printf("Warning: Failed to perform database maintenance: %v", err)
	}

	log.Println("Database recovery attempt completed")
	return nil
}

// terminateConflictingConnections terminates connections that might be holding locks
func (im *IdempotentMigrator) terminateConflictingConnections() error {
	// Only terminate connections that are idle in transaction for too long
	query := `
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE state = 'idle in transaction'
		AND state_change < NOW() - INTERVAL '5 minutes'
		AND pid != pg_backend_pid()
	`

	var terminated []bool
	if err := im.db.Raw(query).Scan(&terminated).Error; err != nil {
		return err
	}

	if len(terminated) > 0 {
		log.Printf("Terminated %d idle connections", len(terminated))
	}

	return nil
}

// performDatabaseMaintenance performs basic database maintenance
func (im *IdempotentMigrator) performDatabaseMaintenance() error {
	// Vacuum analyze to update statistics and clean up
	tables := []string{"schema_migrations", "catalog_items", "inventory_lots"}

	for _, table := range tables {
		if err := im.db.Exec(fmt.Sprintf("VACUUM ANALYZE %s", table)).Error; err != nil {
			log.Printf("Warning: Failed to vacuum table %s: %v", table, err)
		}
	}

	return nil
}

// ForceCleanMigration forces a clean migration by dropping and recreating problematic structures
func (im *IdempotentMigrator) ForceCleanMigration() error {
	log.Println("WARNING: Performing force clean migration - this may cause data loss!")

	// This should only be used in development environments
	if !im.isDevelopmentEnvironment() {
		return fmt.Errorf("force clean migration is only allowed in development environments")
	}

	// Clean up failed migrations
	if err := im.runner.CleanupFailedMigrations(); err != nil {
		return fmt.Errorf("failed to cleanup failed migrations: %w", err)
	}

	// Clear all cached plans first
	if err := im.clearCachedPlans(); err != nil {
		log.Printf("Warning: Failed to clear cached plans: %v", err)
	}

	// Drop and recreate problematic tables in dependency order
	problematicTables := []string{
		"inventory_lots", // Has FK to catalog_items
		"catalog_items",  // Main problematic table
	}

	for _, table := range problematicTables {
		log.Printf("Dropping and recreating table: %s", table)

		// Drop table if exists (CASCADE to handle foreign keys)
		if err := im.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)).Error; err != nil {
			log.Printf("Warning: Failed to drop table %s: %v", table, err)
		}
	}

	// Also drop any related indexes that might cause conflicts
	indexesToDrop := []string{
		"idx_catalog_items_org_type_active",
		"idx_catalog_items_category_subcategory",
		"idx_catalog_items_visibility_active",
		"idx_catalog_items_sku_org",
	}

	for _, index := range indexesToDrop {
		sql := fmt.Sprintf("DROP INDEX IF EXISTS %s", index)
		if err := im.db.Exec(sql).Error; err != nil {
			log.Printf("Warning: Failed to drop index %s: %v", index, err)
		}
	}

	// Clear cached plans again after dropping tables
	if err := im.clearCachedPlans(); err != nil {
		log.Printf("Warning: Failed to clear cached plans after drops: %v", err)
	}

	// Now run migrations normally
	return im.RunMigrations()
}

// isDevelopmentEnvironment checks if we're in a development environment
func (im *IdempotentMigrator) isDevelopmentEnvironment() bool {
	// Check for development indicators
	// This is a simple check - in production you'd want more sophisticated detection
	var dbName string
	if err := im.db.Raw("SELECT current_database()").Scan(&dbName).Error; err != nil {
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

// extractConstraintName extracts the constraint name from an ALTER TABLE ADD CONSTRAINT statement
func (im *IdempotentMigrator) extractConstraintName(sql string) string {
	// Look for pattern "CONSTRAINT constraint_name"
	parts := strings.Fields(sql)
	for i, part := range parts {
		if strings.ToUpper(part) == "CONSTRAINT" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
