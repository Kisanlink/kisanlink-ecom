package database

import (
	"context"
	"fmt"
	"log"

	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// ModelCounterConfig defines the configuration for each model's counter initialization
type ModelCounterConfig struct {
	TableIdentifier string
	TableName       string
	TableSize       hash.TableSize
}

// GetAllModelConfigs returns configurations for all models that use hash-based IDs
func GetAllModelConfigs() []ModelCounterConfig {
	return []ModelCounterConfig{
		// Order models
		{TableIdentifier: "ORD", TableName: "orders", TableSize: "large"},
		{TableIdentifier: "ITEM", TableName: "order_items", TableSize: "large"},
		{TableIdentifier: "HIST", TableName: "order_status_history", TableSize: "large"},
		{TableIdentifier: "INV", TableName: "invoices", TableSize: "large"},
		{TableIdentifier: "INVLI", TableName: "invoice_line_items", TableSize: "large"},

		// Catalog models
		{TableIdentifier: "CAT", TableName: "catalog_items", TableSize: "large"},
		{TableIdentifier: "VAR", TableName: "variants", TableSize: "medium"},
		{TableIdentifier: "AVL", TableName: "availabilities", TableSize: "medium"},
		{TableIdentifier: "LOT", TableName: "inventory_lots", TableSize: "large"},
		{TableIdentifier: "AUDT", TableName: "inventory_audit_logs", TableSize: "large"},
		{TableIdentifier: "CSLA", TableName: "catalog_slas", TableSize: "medium"},
		{TableIdentifier: "CAT", TableName: "categories", TableSize: "medium"},

		// Service models
		{TableIdentifier: "SSLA", TableName: "service_slas", TableSize: "medium"},

		// Media models
		{TableIdentifier: "MED", TableName: "media", TableSize: "large"},

		// Pricing models
		{TableIdentifier: "PRC", TableName: "prices", TableSize: "large"},
		{TableIdentifier: "TIER", TableName: "price_tiers", TableSize: "medium"},
		{TableIdentifier: "RULE", TableName: "price_rules", TableSize: "medium"},

		// Actor models
		{TableIdentifier: "VEND", TableName: "vendors", TableSize: "medium"},
		{TableIdentifier: "CUST", TableName: "customers", TableSize: "large"},
		{TableIdentifier: "COLB", TableName: "collaborators", TableSize: "medium"},

		// Taxation models
		{TableIdentifier: "TXRT", TableName: "tax_rates", TableSize: "medium"},
		{TableIdentifier: "TXRL", TableName: "tax_rules", TableSize: "medium"},
		{TableIdentifier: "TXEX", TableName: "tax_exemptions", TableSize: "medium"},

		// Discount models
		{TableIdentifier: "DISC", TableName: "discounts", TableSize: "medium"},
		{TableIdentifier: "DRUL", TableName: "discount_rules", TableSize: "medium"},
		{TableIdentifier: "DUSG", TableName: "discount_usages", TableSize: "large"},

		// Marketplace models
		{TableIdentifier: "LIST", TableName: "listings", TableSize: "large"},
		{TableIdentifier: "BID", TableName: "bids", TableSize: "large"},
		{TableIdentifier: "AUCN", TableName: "auction_events", TableSize: "large"},

		// Common models
		{TableIdentifier: "SEQ", TableName: "sequence_counters", TableSize: "small"},
		{TableIdentifier: "AUD", TableName: "audit_logs", TableSize: "large"},

		// User and role models
		{TableIdentifier: "USER", TableName: "users", TableSize: "large"},
		{TableIdentifier: "ROLE", TableName: "ecommerce_roles", TableSize: "small"},
		{TableIdentifier: "URLR", TableName: "user_roles", TableSize: "medium"},
		{TableIdentifier: "ORRL", TableName: "organization_roles", TableSize: "medium"},

		// Outbox models
		{TableIdentifier: "OUTB", TableName: "outbox_events", TableSize: "large"},
	}
}

// InitializeAllCounters initializes counters for all models from the database
func InitializeAllCounters(dbManager db.DBManager) error {
	ctx := context.Background()
	configs := GetAllModelConfigs()

	log.Printf("Initializing ID counters for %d model types...", len(configs))

	for _, config := range configs {
		if err := initializeCounterForModel(ctx, dbManager, config); err != nil {
			log.Printf("Warning: Failed to initialize counter for %s (%s): %v",
				config.TableName, config.TableIdentifier, err)
			// Continue with other models even if one fails
			continue
		}
		log.Printf("✓ Initialized counter for %s (%s)", config.TableName, config.TableIdentifier)
	}

	log.Println("Counter initialization completed")
	return nil
}

// initializeCounterForModel initializes the counter for a single model
func initializeCounterForModel(ctx context.Context, dbManager db.DBManager, config ModelCounterConfig) error {
	// Cast to PostgresManager to access GetDB
	postgresManager, ok := dbManager.(*db.PostgresManager)
	if !ok {
		return fmt.Errorf("database manager is not PostgresManager")
	}

	database, err := postgresManager.GetDB(ctx, true) // Use read replica for querying existing IDs
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Query all existing IDs for this table
	var existingIDs []string
	query := fmt.Sprintf("SELECT id FROM %s WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT 1000", config.TableName)

	rows, err := database.Raw(query).Rows()
	if err != nil {
		// If table doesn't exist yet, skip initialization
		if err.Error() == "relation \""+config.TableName+"\" does not exist" {
			return nil
		}
		return fmt.Errorf("failed to query existing IDs: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			log.Printf("Warning: failed to close rows for %s: %v", config.TableName, closeErr)
		}
	}()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("failed to scan ID: %w", err)
		}
		existingIDs = append(existingIDs, id)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	// Initialize the global counter with existing IDs
	hash.InitializeGlobalCountersFromDatabase(config.TableIdentifier, existingIDs, config.TableSize)

	return nil
}
