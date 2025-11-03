package migrations

import (
	"context"
	"fmt"
	"log"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// CreatePublishStatesTable creates the publish_states table for product publishing workflow
func CreatePublishStatesTable(dbManager db.DBManager) error {
	log.Println("Creating publish_states table...")

	// Cast to PostgresManager to access GetDB method
	postgresManager, ok := dbManager.(*db.PostgresManager)
	if !ok {
		return fmt.Errorf("failed to cast to PostgresManager")
	}

	// Get the raw database connection for executing raw SQL
	gormDB, err := postgresManager.GetDB(context.Background(), false)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Define publish_states schema SQL statements
	statements := []string{
		// Publish States Table
		`CREATE TABLE IF NOT EXISTS publish_states (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			product_id UUID NOT NULL,
			fpo_access_list JSONB NOT NULL DEFAULT '[]',
			delivery_costs JSONB NOT NULL DEFAULT '{}',
			platform_fee_percent DECIMAL(5,2) NOT NULL DEFAULT 10.00,
			published_at TIMESTAMP,
			published_by VARCHAR(255),
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			created_by VARCHAR(255) NOT NULL,
			updated_by VARCHAR(255) NOT NULL,
			deleted_at TIMESTAMP,

			CONSTRAINT unique_product_publish UNIQUE(product_id),
			CONSTRAINT valid_platform_fee CHECK (platform_fee_percent >= 0 AND platform_fee_percent <= 100),
			CONSTRAINT fk_publish_states_product FOREIGN KEY (product_id) REFERENCES catalog_items(id) ON DELETE CASCADE
		)`,

		// Create indexes for publish_states
		`CREATE INDEX IF NOT EXISTS idx_publish_states_product_id
			ON publish_states(product_id) WHERE deleted_at IS NULL`,

		`CREATE INDEX IF NOT EXISTS idx_publish_states_published_at
			ON publish_states(published_at DESC) WHERE deleted_at IS NULL`,

		`CREATE INDEX IF NOT EXISTS idx_publish_states_fpo_access
			ON publish_states USING GIN(fpo_access_list)`,

		`CREATE INDEX IF NOT EXISTS idx_publish_states_deleted
			ON publish_states(deleted_at)`,

		// Add comment to the table
		`COMMENT ON TABLE publish_states IS 'Tracks product publishing state and FPO-specific configurations'`,

		`COMMENT ON COLUMN publish_states.product_id IS 'Reference to the published product in catalog_items'`,

		`COMMENT ON COLUMN publish_states.fpo_access_list IS 'JSONB array of FPO organization IDs with access to this product'`,

		`COMMENT ON COLUMN publish_states.delivery_costs IS 'JSONB map of FPO ID to delivery cost (e.g., {"fpo_123": 50.00})'`,

		`COMMENT ON COLUMN publish_states.platform_fee_percent IS 'Platform commission percentage (0-100)'`,
	}

	// Execute all statements
	for _, statement := range statements {
		if err := gormDB.Exec(statement).Error; err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	log.Println("Successfully created publish_states table and indexes")
	return nil
}

// DropPublishStatesTable drops the publish_states table (for rollback)
func DropPublishStatesTable(dbManager db.DBManager) error {
	log.Println("Dropping publish_states table...")

	postgresManager, ok := dbManager.(*db.PostgresManager)
	if !ok {
		return fmt.Errorf("failed to cast to PostgresManager")
	}

	gormDB, err := postgresManager.GetDB(context.Background(), false)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	statement := `DROP TABLE IF EXISTS publish_states CASCADE`

	if err := gormDB.Exec(statement).Error; err != nil {
		return fmt.Errorf("failed to drop publish_states table: %w", err)
	}

	log.Println("Successfully dropped publish_states table")
	return nil
}
