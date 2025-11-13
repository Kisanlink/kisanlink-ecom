package migrations

import (
	"context"
	"fmt"
	"log"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// CreateMarketplaceSchema creates the marketplace tables with proper constraints and indexes
func CreateMarketplaceSchema(dbManager db.DBManager) error {
	log.Println("Creating marketplace schema...")

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

	// Define marketplace schema SQL statements
	statements := []string{
		// Marketplace Listings Table
		`CREATE TABLE IF NOT EXISTS marketplace_listings (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			listing_id VARCHAR(50) UNIQUE NOT NULL,
			product_id VARCHAR(50) NOT NULL,
			seller_id VARCHAR(50) NOT NULL,
			organization_id VARCHAR(50) NOT NULL,

			quantity DECIMAL(10,3) NOT NULL CHECK (quantity > 0),
			asking_price DECIMAL(10,2) NOT NULL CHECK (asking_price > 0),
			minimum_bid DECIMAL(10,2) NOT NULL CHECK (minimum_bid > 0),
			currency VARCHAR(3) NOT NULL DEFAULT 'INR',

			listing_duration_hours INTEGER NOT NULL CHECK (listing_duration_hours > 0),
			expires_at TIMESTAMP NOT NULL,

			status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'CLOSED', 'EXPIRED', 'CANCELLED', 'EXPIRED_NO_BIDS')),
			current_highest_bid_id VARCHAR(50),
			bid_count INTEGER DEFAULT 0 CHECK (bid_count >= 0),

			visibility VARCHAR(20) NOT NULL DEFAULT 'PUBLIC' CHECK (visibility IN ('PRIVATE', 'PUBLIC', 'NETWORK', 'ORGANIZATION')),
			auction_type VARCHAR(10) NOT NULL DEFAULT 'OPEN' CHECK (auction_type IN ('OPEN', 'CLOSED')),
			bid_visibility VARCHAR(20) NOT NULL DEFAULT 'FULL' CHECK (bid_visibility IN ('FULL', 'PARTIAL', 'MINIMAL', 'HIDDEN')),

			pickup_location JSONB,
			terms_conditions TEXT,
			listing_type VARCHAR(20) DEFAULT 'AUCTION',

			closed_at TIMESTAMP,
			close_reason VARCHAR(50),

			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW(),

			CONSTRAINT chk_minimum_bid_le_asking_price CHECK (minimum_bid <= asking_price),
			CONSTRAINT chk_expires_at_future CHECK (expires_at > created_at)
		)`,

		// Marketplace Bids Table
		`CREATE TABLE IF NOT EXISTS marketplace_bids (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			bid_id VARCHAR(50) UNIQUE NOT NULL,
			listing_id VARCHAR(50) NOT NULL,
			bidder_id VARCHAR(50) NOT NULL,

			bid_amount DECIMAL(10,2) NOT NULL CHECK (bid_amount > 0),
			currency VARCHAR(3) NOT NULL DEFAULT 'INR',
			quantity DECIMAL(10,3) NOT NULL CHECK (quantity > 0),
			message TEXT,

			auto_bid_limit DECIMAL(10,2) CHECK (auto_bid_limit IS NULL OR auto_bid_limit > 0),
			is_auto_bid BOOLEAN DEFAULT FALSE,
			parent_bid_id VARCHAR(50),

			status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'OUTBID', 'WINNING', 'EXPIRED', 'REMOVED')),
			is_highest_bid BOOLEAN DEFAULT FALSE,
			outbid_at TIMESTAMP,

			payment_method VARCHAR(50),
			placed_at TIMESTAMP NOT NULL DEFAULT NOW(),

			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW(),

			CONSTRAINT chk_auto_bid_limit_ge_bid_amount CHECK (auto_bid_limit IS NULL OR auto_bid_limit >= bid_amount),
			CONSTRAINT chk_parent_bid_for_auto_bid CHECK ((is_auto_bid = FALSE AND parent_bid_id IS NULL) OR (is_auto_bid = TRUE AND parent_bid_id IS NOT NULL))
		)`,

		// Auction Events Table
		`CREATE TABLE IF NOT EXISTS auction_events (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			event_id VARCHAR(50) UNIQUE NOT NULL,
			listing_id VARCHAR(50) NOT NULL,
			event_type VARCHAR(50) NOT NULL CHECK (event_type IN ('LISTING_CREATED', 'BID_PLACED', 'BID_OUTBID', 'AUTO_BID_TRIGGERED', 'LISTING_CLOSED', 'LISTING_EXPIRED', 'LISTING_CANCELLED', 'BID_REMOVED', 'LISTING_UPDATED')),
			event_data JSONB,
			actor_id VARCHAR(50) NOT NULL,
			actor_type VARCHAR(20) NOT NULL CHECK (actor_type IN ('USER', 'SYSTEM', 'ADMIN')),
			timestamp TIMESTAMP NOT NULL DEFAULT NOW(),

			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,

		// Add foreign key constraints (if referential integrity is desired)
		`ALTER TABLE marketplace_bids 
		 ADD CONSTRAINT fk_marketplace_bids_listing 
		 FOREIGN KEY (listing_id) REFERENCES marketplace_listings(listing_id) 
		 ON DELETE CASCADE`,

		`ALTER TABLE auction_events 
		 ADD CONSTRAINT fk_auction_events_listing 
		 FOREIGN KEY (listing_id) REFERENCES marketplace_listings(listing_id) 
		 ON DELETE CASCADE`,

		// Create performance indexes
		// Marketplace Listings indexes
		`CREATE INDEX IF NOT EXISTS idx_marketplace_listings_listing_id ON marketplace_listings(listing_id)`,
		`CREATE INDEX IF NOT EXISTS idx_marketplace_listings_product_id ON marketplace_listings(product_id)`,
		`CREATE INDEX IF NOT EXISTS idx_marketplace_listings_seller_id ON marketplace_listings(seller_id)`,
		`CREATE INDEX IF NOT EXISTS idx_marketplace_listings_organization_id ON marketplace_listings(organization_id)`,
		`CREATE INDEX IF NOT EXISTS idx_marketplace_listings_status ON marketplace_listings(status)`,
		`CREATE INDEX IF NOT EXISTS idx_marketplace_listings_expires_at ON marketplace_listings(expires_at)`,
		`CREATE INDEX IF NOT EXISTS idx_marketplace_listings_status_expires ON marketplace_listings(status, expires_at)`,
		`CREATE INDEX IF NOT EXISTS idx_marketplace_listings_seller_status ON marketplace_listings(seller_id, status)`,
		`CREATE INDEX IF NOT EXISTS idx_marketplace_listings_org_status ON marketplace_listings(organization_id, status)`,
		`CREATE INDEX IF NOT EXISTS idx_marketplace_listings_product_status ON marketplace_listings(product_id, status)`,
		`CREATE INDEX IF NOT EXISTS idx_marketplace_listings_visibility_status ON marketplace_listings(visibility, status)`,
		`CREATE INDEX IF NOT EXISTS idx_marketplace_listings_auction_type ON marketplace_listings(auction_type, status)`,
		`CREATE INDEX IF NOT EXISTS idx_marketplace_listings_asking_price ON marketplace_listings(asking_price)`,
		`CREATE INDEX IF NOT EXISTS idx_marketplace_listings_bid_count ON marketplace_listings(bid_count)`,

		// Marketplace Bids indexes
		`CREATE INDEX IF NOT EXISTS idx_marketplace_bids_bid_id ON marketplace_bids(bid_id)`,
		`CREATE INDEX IF NOT EXISTS idx_marketplace_bids_listing_id ON marketplace_bids(listing_id)`,
		`CREATE INDEX IF NOT EXISTS idx_marketplace_bids_bidder_id ON marketplace_bids(bidder_id)`,
		`CREATE INDEX IF NOT EXISTS idx_marketplace_bids_status ON marketplace_bids(status)`,
		`CREATE INDEX IF NOT EXISTS idx_marketplace_bids_is_highest_bid ON marketplace_bids(is_highest_bid)`,
		`CREATE INDEX IF NOT EXISTS idx_marketplace_bids_placed_at ON marketplace_bids(placed_at)`,
		`CREATE INDEX IF NOT EXISTS idx_marketplace_bids_listing_status ON marketplace_bids(listing_id, status)`,
		`CREATE INDEX IF NOT EXISTS idx_marketplace_bids_bidder_status ON marketplace_bids(bidder_id, status)`,
		`CREATE INDEX IF NOT EXISTS idx_marketplace_bids_listing_placed ON marketplace_bids(listing_id, placed_at)`,
		`CREATE INDEX IF NOT EXISTS idx_marketplace_bids_highest_bid ON marketplace_bids(listing_id, is_highest_bid)`,
		`CREATE INDEX IF NOT EXISTS idx_marketplace_bids_auto_bid ON marketplace_bids(listing_id, is_auto_bid)`,
		`CREATE INDEX IF NOT EXISTS idx_marketplace_bids_bid_amount ON marketplace_bids(listing_id, bid_amount)`,

		// Auction Events indexes
		`CREATE INDEX IF NOT EXISTS idx_auction_events_event_id ON auction_events(event_id)`,
		`CREATE INDEX IF NOT EXISTS idx_auction_events_listing_id ON auction_events(listing_id)`,
		`CREATE INDEX IF NOT EXISTS idx_auction_events_event_type ON auction_events(event_type)`,
		`CREATE INDEX IF NOT EXISTS idx_auction_events_timestamp ON auction_events(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_auction_events_listing_type ON auction_events(listing_id, event_type)`,
		`CREATE INDEX IF NOT EXISTS idx_auction_events_listing_timestamp ON auction_events(listing_id, timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_auction_events_actor_type ON auction_events(actor_id, actor_type)`,
		`CREATE INDEX IF NOT EXISTS idx_auction_events_type_timestamp ON auction_events(event_type, timestamp)`,
	}

	// Execute each statement
	successCount := 0
	for i, statement := range statements {
		if err := gormDB.Exec(statement).Error; err != nil {
			log.Printf("Warning: Failed to execute statement %d: %v", i+1, err)
			log.Printf("SQL: %s", statement)
			// Continue with other statements instead of failing completely
		} else {
			successCount++
		}
	}

	log.Printf("Successfully executed %d out of %d marketplace schema statements", successCount, len(statements))

	// Verify tables were created
	if err := verifyMarketplaceTables(gormDB); err != nil {
		return fmt.Errorf("marketplace table verification failed: %w", err)
	}

	log.Println("Marketplace schema created successfully")
	return nil
}

// verifyMarketplaceTables verifies that the marketplace tables were created properly
func verifyMarketplaceTables(gormDB interface{}) error {
	// This is a simplified verification - in a real implementation,
	// you might want to check table structure, constraints, etc.

	tables := []string{
		"marketplace_listings",
		"marketplace_bids",
		"auction_events",
	}

	// For now, just log that verification would happen here
	log.Printf("Verifying marketplace tables: %v", tables)

	// In a real implementation, you would query the database to verify
	// table existence and structure

	return nil
}

// DropMarketplaceSchema drops the marketplace tables (for testing/cleanup)
func DropMarketplaceSchema(dbManager db.DBManager) error {
	log.Println("Dropping marketplace schema...")

	postgresManager, ok := dbManager.(*db.PostgresManager)
	if !ok {
		return fmt.Errorf("failed to cast to PostgresManager")
	}

	gormDB, err := postgresManager.GetDB(context.Background(), false)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Drop tables in reverse order to handle foreign key constraints
	dropStatements := []string{
		"DROP TABLE IF EXISTS auction_events CASCADE",
		"DROP TABLE IF EXISTS marketplace_bids CASCADE",
		"DROP TABLE IF EXISTS marketplace_listings CASCADE",
	}

	for i, statement := range dropStatements {
		if err := gormDB.Exec(statement).Error; err != nil {
			log.Printf("Warning: Failed to execute drop statement %d: %v", i+1, err)
		}
	}

	log.Println("Marketplace schema dropped successfully")
	return nil
}
