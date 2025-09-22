# Database Migration System

This document describes the enhanced database migration system implemented for the KisanLink E-Commerce platform.

## Overview

The migration system provides:

- **Idempotent Migrations**: Safe to run multiple times without side effects
- **Version Tracking**: Custom schema_migrations table with semantic versioning
- **Rollback Safety**: Pre-migration safety checks and rollback procedures
- **Dry Run Mode**: Test migrations without applying changes
- **Migration History**: Complete audit trail of all migration attempts
- **Performance Optimization**: Automatic creation of performance indexes

## Architecture

### Components

1. **SchemaMigration Model**: Tracks migration history in the database
2. **MigrationRunner**: Low-level migration execution and tracking
3. **IdempotentMigrator**: High-level migration orchestration with safety checks
4. **MigrationManager**: CLI interface for migration management

### Migration Flow

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   CLI Command   │───▶│ MigrationManager │───▶│ IdempotentMigrator │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                                         │
                                                         ▼
                                               ┌─────────────────┐
                                               │ MigrationRunner │
                                               └─────────────────┘
                                                         │
                                                         ▼
                                               ┌─────────────────┐
                                               │    Database     │
                                               └─────────────────┘
```

## Usage

### CLI Commands

The migration system provides a CLI tool for managing database migrations:

```bash
# Build the migration CLI
go build -o migrate ./cmd/migrate

# Run all pending migrations
./migrate migrate

# Show current migration status
./migrate status

# Show migration history
./migrate history

# Show what migrations would be executed
./migrate plan

# Run migrations in dry-run mode (no changes applied)
./migrate dry-run

# Validate migration integrity
./migrate validate

# Clean up failed migrations (requires --confirm)
./migrate cleanup --confirm

# Rollback a specific migration (use with caution)
./migrate rollback v1.0.0
```

### Programmatic Usage

```go
// Create database manager
dbManager, err := database.NewManager(cfg)
if err != nil {
    return err
}

// Run migrations during application startup
if err := database.RunAutoMigrations(dbManager); err != nil {
    return fmt.Errorf("migration failed: %w", err)
}

// Or use the migration manager for more control
pgManager := dbManager.GetManager(db.BackendGorm)
migrationManager, err := database.NewMigrationManager(pgManager)
if err != nil {
    return err
}

// Execute specific commands
err = migrationManager.ExecuteCommand(database.CommandMigrate)
```

### Bootstrap Integration

The migration system integrates with the application bootstrap process:

```go
// Run migrations during system initialization
err := bootstrap.RunDatabaseMigrations(dbManager, false) // false = not dry-run
if err != nil {
    log.Fatalf("Database migration failed: %v", err)
}
```

## Migration Definitions

### Default Migrations

The system includes predefined migrations:

1. **v1.0.0**: Initial catalog models migration
   - Migrates all domain models in dependency order
   - Creates base tables with proper relationships

2. **v1.1.0**: Performance indexes and constraints
   - Creates composite indexes for common query patterns
   - Adds performance optimizations for high-volume operations

### Custom Migrations

You can add custom migrations by extending the `IdempotentMigrator`:

```go
// Add a custom migration
migrator.AddMigration(database.MigrationDefinition{
    Version:     "v1.2.0",
    Description: "Add new feature tables",
    Models:      []interface{}{&MyNewModel{}},
    PreHook: func(tx *gorm.DB) error {
        // Pre-migration logic
        return nil
    },
    PostHook: func(tx *gorm.DB) error {
        // Post-migration logic (e.g., data migration)
        return nil
    },
})
```

## Safety Features

### Pre-Migration Safety Checks

1. **Database Connectivity**: Verifies database connection
2. **Migration Conflicts**: Checks for duplicate version numbers
3. **Failed Migrations**: Warns about previous failed attempts
4. **Database Locks**: Detects conflicting exclusive locks

### Rollback Safety

- Rollback operations require safety checks to be enabled
- Automatic rollback is not implemented to prevent data loss
- Manual rollback guidance is provided for failed migrations

### Idempotency

- Migrations can be run multiple times safely
- Already applied migrations are automatically skipped
- Migration checksums detect changes to existing migrations

## Schema Migrations Table

The system maintains a `schema_migrations` table with the following structure:

```sql
CREATE TABLE schema_migrations (
    id          SERIAL PRIMARY KEY,
    version     VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    applied_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    checksum    VARCHAR(64) NOT NULL,
    success     BOOLEAN NOT NULL DEFAULT FALSE,
    error_msg   TEXT,
    duration    BIGINT NOT NULL DEFAULT 0
);
```

### Fields

- **version**: Semantic version of the migration (e.g., "v1.0.0")
- **description**: Human-readable description of the migration
- **applied_at**: Timestamp when the migration was attempted
- **checksum**: SHA-256 checksum of the migration definition
- **success**: Whether the migration completed successfully
- **error_msg**: Error message if the migration failed
- **duration**: Migration execution time in milliseconds

## Performance Indexes

The system automatically creates performance indexes for common query patterns:

### Catalog Items

- `(organization_id, item_type, is_active)`
- `(category, subcategory)`
- `(visibility, is_active)`
- `(sku, organization_id)`

### Pricing

- `(catalog_item_id, currency)`
- `(organization_id, currency)`
- `(base_price)`

### Inventory

- `(product_id, organization_id)`
- `(lot_number)`
- `(expiry_date)`

### Audit Logs

- `(entity_type, action)`
- `(entity_id)`
- `(user_id, timestamp)`
- `(organization_id, timestamp)`

## Error Handling

### Migration Failures

When a migration fails:

1. The error is logged with full context
2. The failure is recorded in the schema_migrations table
3. Subsequent migrations are blocked until the issue is resolved
4. Failed migrations can be cleaned up using the CLI

### Recovery Procedures

1. **Investigate the failure**: Check logs and error messages
2. **Fix the underlying issue**: Resolve database problems or code issues
3. **Clean up failed records**: Use `./migrate cleanup --confirm`
4. **Retry the migration**: Run `./migrate migrate` again

## Best Practices

### Migration Development

1. **Test migrations thoroughly** in development environments
2. **Use dry-run mode** to validate migration plans
3. **Keep migrations small** and focused on specific changes
4. **Include rollback procedures** for complex migrations
5. **Document breaking changes** in migration descriptions

### Production Deployment

1. **Run migrations during maintenance windows** for large changes
2. **Monitor migration performance** and duration
3. **Have rollback plans ready** for critical migrations
4. **Use feature flags** for risky schema changes
5. **Backup databases** before major migrations

### Monitoring

1. **Check migration status** regularly using `./migrate status`
2. **Monitor migration history** for patterns and issues
3. **Set up alerts** for migration failures
4. **Track migration performance** over time

## Troubleshooting

### Common Issues

1. **Migration stuck**: Check for database locks or long-running transactions
2. **Checksum mismatch**: Migration definition changed after being applied
3. **Permission errors**: Database user lacks necessary privileges
4. **Timeout errors**: Migration taking too long, increase timeout settings

### Debug Commands

```bash
# Show detailed migration status
./migrate status

# Show complete migration history
./migrate history

# Validate migration integrity
./migrate validate

# Show what would be executed
./migrate plan
```

## Configuration

### Environment Variables

The migration system uses the same database configuration as the main application:

```bash
# PostgreSQL Configuration
DB_POSTGRES_HOST=localhost
DB_POSTGRES_PORT=5432
DB_POSTGRES_USER=postgres
DB_POSTGRES_PASSWORD=password
DB_POSTGRES_DBNAME=kisanlink_ecom
DB_POSTGRES_SSLMODE=disable
DB_POSTGRES_MAX_CONNS=25
DB_POSTGRES_IDLE_CONNS=5
```

### Migration Settings

```go
// Disable safety checks (not recommended for production)
migrator.SetSafetyCheck(false)

// Enable dry-run mode
migrator.SetDryRun(true)
```

## Future Enhancements

1. **Migration Templates**: Code generation for common migration patterns
2. **Data Migration Support**: Built-in support for data transformations
3. **Parallel Migrations**: Execute independent migrations in parallel
4. **Migration Dependencies**: Explicit dependency management between migrations
5. **Schema Validation**: Automatic validation of schema changes
6. **Migration Metrics**: Detailed performance and success metrics
7. **Integration Testing**: Automated testing of migration procedures
