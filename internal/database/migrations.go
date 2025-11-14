package database

import (
	"context"
	"fmt"
	"log"

	"github.com/Kisanlink/kisanlink-ecom/entities/models"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// RunAutoMigrations runs auto-migrations for all models using the simple GORM-only migrator
func RunAutoMigrations(dbManager *DatabaseManager) error {
	log.Println("Starting simple GORM auto-migration...")

	// Get the PostgreSQL manager
	pgManager := dbManager.GetManager(db.BackendGorm)
	if pgManager == nil {
		return fmt.Errorf("PostgreSQL manager not available")
	}

	// Create simple migrator
	migrator, err := NewSimpleMigrator(pgManager)
	if err != nil {
		return fmt.Errorf("failed to create simple migrator: %w", err)
	}

	// Run migration with simple retry logic
	if err := migrator.RunMigrations(); err != nil {
		log.Printf("Migration failed: %v", err)

		// If we're in development, offer force clean option
		if migrator.isDevelopmentEnvironment() {
			log.Println("Development environment detected. Attempting force clean migration...")
			if forceErr := migrator.ForceCleanMigration(); forceErr != nil {
				return fmt.Errorf("all migration attempts failed - original: %w, force: %v", err, forceErr)
			}
			log.Println("Force clean migration completed successfully")
		} else {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	log.Println("Simple auto-migration completed successfully")
	return nil
}

// RunAutoMigrationsLegacy runs the legacy auto-migration for backward compatibility
func RunAutoMigrationsLegacy(dbManager *DatabaseManager) error {
	log.Println("Starting legacy auto-migration using kisanlink-db...")

	// Get the PostgreSQL manager
	pgManager := dbManager.GetManager(db.BackendGorm)
	if pgManager == nil {
		return fmt.Errorf("PostgreSQL manager not available")
	}

	// Get all models from the registry
	allModels := models.AllModels()

	// Run auto-migration using kisanlink-db
	ctx := context.Background()
	if err := pgManager.AutoMigrateModels(ctx, allModels...); err != nil {
		return fmt.Errorf("auto-migration failed: %w", err)
	}

	log.Printf("Successfully migrated %d models", len(allModels))
	return nil
}

// CreateIndexes creates additional indexes for better performance
func CreateIndexes(dbManager *DatabaseManager) error {
	log.Println("Creating additional performance indexes...")

	pgManager := dbManager.GetManager(db.BackendGorm)
	if pgManager == nil {
		return fmt.Errorf("PostgreSQL manager not available")
	}

	// Cast to PostgresManager to access GetDB method
	postgresManager, ok := pgManager.(*db.PostgresManager)
	if !ok {
		return fmt.Errorf("failed to cast to PostgresManager")
	}

	// Get the raw database connection for executing raw SQL
	database, err := postgresManager.GetDB(context.Background(), false)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Create catalog-specific performance indexes first
	if err := CreateCatalogIndexes(pgManager); err != nil {
		log.Printf("Warning: Failed to create catalog indexes: %v", err)
		// Continue with other indexes
	}

	// Define performance indexes for common query patterns
	indexes := []string{
		// Catalog Items - Organization and type filtering
		"CREATE INDEX IF NOT EXISTS idx_catalog_items_org_type_active ON catalog_items(organization_id, item_type, is_active)",
		"CREATE INDEX IF NOT EXISTS idx_catalog_items_category_subcategory ON catalog_items(category, subcategory)",
		"CREATE INDEX IF NOT EXISTS idx_catalog_items_visibility_active ON catalog_items(visibility, is_active)",
		"CREATE INDEX IF NOT EXISTS idx_catalog_items_sku_org ON catalog_items(sku, organization_id)",

		// Orders - Buyer and seller organization queries
		"CREATE INDEX IF NOT EXISTS idx_orders_buyer_status ON orders(buyer_organization_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_orders_seller_status ON orders(seller_organization_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_orders_buyer_created ON orders(buyer_organization_id, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_orders_seller_created ON orders(seller_organization_id, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_orders_status_created ON orders(status, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_orders_total_amount ON orders(total_amount)",

		// Order Items - Order and catalog item relationships
		"CREATE INDEX IF NOT EXISTS idx_order_items_order_catalog ON order_items(order_id, catalog_item_id)",
		"CREATE INDEX IF NOT EXISTS idx_order_items_catalog_type ON order_items(catalog_item_id, catalog_item_type)",

		// Order Status History - Order and status tracking
		"CREATE INDEX IF NOT EXISTS idx_order_status_history_order_status ON order_status_history(order_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_order_status_history_created ON order_status_history(created_at)",

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

		// Taxation - Rate and rule queries
		"CREATE INDEX IF NOT EXISTS idx_tax_rates_org_type ON tax_rates(organization_id, tax_type)",
		"CREATE INDEX IF NOT EXISTS idx_tax_rates_region ON tax_rates(region)",
		"CREATE INDEX IF NOT EXISTS idx_tax_rules_tax_rate_id ON tax_rules(tax_rate_id)",

		// Discounts - Organization and usage tracking
		"CREATE INDEX IF NOT EXISTS idx_discounts_org_active ON discounts(organization_id, is_active)",
		"CREATE INDEX IF NOT EXISTS idx_discounts_code_org ON discounts(discount_code, organization_id)",
		"CREATE INDEX IF NOT EXISTS idx_discounts_valid_from_to ON discounts(valid_from, valid_to)",
		"CREATE INDEX IF NOT EXISTS idx_discount_usage_discount_id ON discount_usage(discount_id)",
		"CREATE INDEX IF NOT EXISTS idx_discount_usage_order_id ON discount_usage(order_id)",

		// Inventory Lots - Product and organization tracking
		"CREATE INDEX IF NOT EXISTS idx_inventory_lots_product_org ON inventory_lots(product_id, organization_id)",
		"CREATE INDEX IF NOT EXISTS idx_inventory_lots_lot_number ON inventory_lots(lot_number)",
		"CREATE INDEX IF NOT EXISTS idx_inventory_lots_expiry_date ON inventory_lots(expiry_date)",

		// Marketplace Listings - Core auction queries
		"CREATE INDEX IF NOT EXISTS idx_marketplace_listings_status_expires ON marketplace_listings(status, expires_at)",
		"CREATE INDEX IF NOT EXISTS idx_marketplace_listings_seller_status ON marketplace_listings(seller_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_marketplace_listings_org_status ON marketplace_listings(organization_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_marketplace_listings_product_status ON marketplace_listings(product_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_marketplace_listings_visibility_status ON marketplace_listings(visibility, status)",
		"CREATE INDEX IF NOT EXISTS idx_marketplace_listings_auction_type ON marketplace_listings(auction_type, status)",
		"CREATE INDEX IF NOT EXISTS idx_marketplace_listings_expires_status ON marketplace_listings(expires_at, status)",
		"CREATE INDEX IF NOT EXISTS idx_marketplace_listings_asking_price ON marketplace_listings(asking_price)",
		"CREATE INDEX IF NOT EXISTS idx_marketplace_listings_bid_count ON marketplace_listings(bid_count)",

		// Marketplace Bids - Bidding operations and queries
		"CREATE INDEX IF NOT EXISTS idx_marketplace_bids_listing_status ON marketplace_bids(listing_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_marketplace_bids_bidder_status ON marketplace_bids(bidder_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_marketplace_bids_listing_placed ON marketplace_bids(listing_id, placed_at)",
		"CREATE INDEX IF NOT EXISTS idx_marketplace_bids_highest_bid ON marketplace_bids(listing_id, is_highest_bid)",
		"CREATE INDEX IF NOT EXISTS idx_marketplace_bids_auto_bid ON marketplace_bids(listing_id, is_auto_bid)",
		"CREATE INDEX IF NOT EXISTS idx_marketplace_bids_bid_amount ON marketplace_bids(listing_id, bid_amount)",
		"CREATE INDEX IF NOT EXISTS idx_marketplace_bids_placed_at ON marketplace_bids(placed_at)",

		// Auction Events - Event tracking and audit
		"CREATE INDEX IF NOT EXISTS idx_auction_events_listing_type ON auction_events(listing_id, event_type)",
		"CREATE INDEX IF NOT EXISTS idx_auction_events_listing_timestamp ON auction_events(listing_id, timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_auction_events_actor_type ON auction_events(actor_id, actor_type)",
		"CREATE INDEX IF NOT EXISTS idx_auction_events_type_timestamp ON auction_events(event_type, timestamp)",

		// Outbox Events - Event processing and status
		"CREATE INDEX IF NOT EXISTS idx_outbox_events_status_created ON outbox_events(status, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_outbox_events_event_type ON outbox_events(event_type)",
		"CREATE INDEX IF NOT EXISTS idx_outbox_events_retry_count ON outbox_events(retry_count)",

		// Payment Screenshots - Order and verification status queries
		"CREATE INDEX IF NOT EXISTS idx_payment_screenshots_order_id ON payment_screenshots(order_id) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_payment_screenshots_verification_status ON payment_screenshots(verification_status) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_payment_screenshots_buyer_org_created ON payment_screenshots(buyer_organization_id, created_at DESC) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_payment_screenshots_seller_org_status ON payment_screenshots(seller_organization_id, verification_status) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_payment_screenshots_uploaded_by ON payment_screenshots(uploaded_by) WHERE deleted_at IS NULL",
	}

	// Execute each index creation statement
	successCount := 0
	for i, indexSQL := range indexes {
		if err := database.Exec(indexSQL).Error; err != nil {
			log.Printf("Warning: Failed to create index %d: %v", i+1, err)
			log.Printf("SQL: %s", indexSQL)
			// Continue with other indexes instead of failing completely
		} else {
			successCount++
		}
	}

	log.Printf("Successfully created %d out of %d performance indexes", successCount, len(indexes))
	return nil
}
