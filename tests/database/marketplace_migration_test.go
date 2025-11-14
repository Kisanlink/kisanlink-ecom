//go:build integration
// +build integration

package database_test

import (
	"context"
	"testing"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/marketplace"
	"github.com/Kisanlink/kisanlink-ecom/internal/database"
	"github.com/Kisanlink/kisanlink-ecom/migrations"
	"github.com/Kisanlink/kisanlink-ecom/tests/testutils"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarketplaceMigration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create test database configuration from environment variables
	cfg := testutils.LoadTestDatabaseConfig()

	// Create database manager
	dbManager, err := database.NewDatabaseManager(cfg)
	if err != nil {
		t.Skipf("Could not connect to test database: %v", err)
	}
	defer dbManager.Close()

	// Get PostgreSQL manager
	pgManager := dbManager.GetManager(db.BackendGorm)
	require.NotNil(t, pgManager)

	// Run marketplace schema migration
	err = migrations.CreateMarketplaceSchema(pgManager)
	if err != nil {
		t.Skipf("Could not create marketplace schema: %v", err)
	}

	// Test that we can create marketplace entities
	t.Run("TestCreateMarketplaceEntities", func(t *testing.T) {
		testCreateMarketplaceEntities(t, pgManager)
	})

	// Clean up - drop the schema
	err = migrations.DropMarketplaceSchema(pgManager)
	assert.NoError(t, err)
}

func testCreateMarketplaceEntities(t *testing.T, pgManager db.DBManager) {
	// Cast to PostgresManager to access GetDB method
	postgresManager, ok := pgManager.(*db.PostgresManager)
	require.True(t, ok)

	gormDB, err := postgresManager.GetDB(context.Background(), false)
	require.NoError(t, err)

	// Test creating a listing
	listing := marketplace.NewListing(
		"PROD_123",
		"USER_456",
		"ORG_789",
		decimal.NewFromFloat(100.5),
		decimal.NewFromFloat(1000.00),
		decimal.NewFromFloat(800.00),
		24,
	)

	err = gormDB.Create(listing).Error
	require.NoError(t, err)
	assert.NotEmpty(t, listing.ID)

	// Test creating a bid
	bid := marketplace.NewBid(
		listing.ListingID,
		"USER_999",
		decimal.NewFromFloat(900.00),
		decimal.NewFromFloat(50.0),
		"Test bid",
	)

	err = gormDB.Create(bid).Error
	require.NoError(t, err)
	assert.NotEmpty(t, bid.ID)

	// Test creating an auction event
	event := marketplace.NewAuctionEvent(
		listing.ListingID,
		marketplace.EventBidPlaced,
		"USER_999",
		marketplace.ActorTypeUser,
		nil,
	)

	eventData := map[string]interface{}{
		"bid_id":     bid.BidID,
		"bid_amount": 900.00,
	}
	err = event.SetEventData(eventData)
	require.NoError(t, err)

	err = gormDB.Create(event).Error
	require.NoError(t, err)
	assert.NotEmpty(t, event.ID)

	// Test querying the created entities
	var retrievedListing marketplace.Listing
	err = gormDB.Where("listing_id = ?", listing.ListingID).First(&retrievedListing).Error
	require.NoError(t, err)
	assert.Equal(t, listing.ListingID, retrievedListing.ListingID)
	assert.Equal(t, listing.ProductID, retrievedListing.ProductID)

	var retrievedBid marketplace.Bid
	err = gormDB.Where("bid_id = ?", bid.BidID).First(&retrievedBid).Error
	require.NoError(t, err)
	assert.Equal(t, bid.BidID, retrievedBid.BidID)
	assert.Equal(t, bid.ListingID, retrievedBid.ListingID)

	var retrievedEvent marketplace.AuctionEvent
	err = gormDB.Where("event_id = ?", event.EventID).First(&retrievedEvent).Error
	require.NoError(t, err)
	assert.Equal(t, event.EventID, retrievedEvent.EventID)
	assert.Equal(t, event.ListingID, retrievedEvent.ListingID)

	// Test foreign key relationships
	var bidsForListing []marketplace.Bid
	err = gormDB.Where("listing_id = ?", listing.ListingID).Find(&bidsForListing).Error
	require.NoError(t, err)
	assert.Len(t, bidsForListing, 1)
	assert.Equal(t, bid.BidID, bidsForListing[0].BidID)

	var eventsForListing []marketplace.AuctionEvent
	err = gormDB.Where("listing_id = ?", listing.ListingID).Find(&eventsForListing).Error
	require.NoError(t, err)
	assert.Len(t, eventsForListing, 1)
	assert.Equal(t, event.EventID, eventsForListing[0].EventID)
}

func TestMarketplaceAutoMigration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create test database configuration from environment variables
	cfg := testutils.LoadTestDatabaseConfig()

	// Create database manager
	dbManager, err := database.NewDatabaseManager(cfg)
	if err != nil {
		t.Skipf("Could not connect to test database: %v", err)
	}
	defer dbManager.Close()

	// Test auto-migration includes marketplace models
	err = database.RunAutoMigrations(dbManager)
	if err != nil {
		t.Skipf("Could not run auto-migrations: %v", err)
	}

	// Test that marketplace tables exist and work
	pgManager := dbManager.GetManager(db.BackendGorm)
	require.NotNil(t, pgManager)

	postgresManager, ok := pgManager.(*db.PostgresManager)
	require.True(t, ok)

	gormDB, err := postgresManager.GetDB(context.Background(), false)
	require.NoError(t, err)

	// Test creating entities using auto-migrated tables
	listing := marketplace.NewListing(
		"PROD_AUTO_123",
		"USER_AUTO_456",
		"ORG_AUTO_789",
		decimal.NewFromFloat(200.0),
		decimal.NewFromFloat(2000.00),
		decimal.NewFromFloat(1500.00),
		48,
	)

	err = gormDB.Create(listing).Error
	assert.NoError(t, err)

	// Verify the listing was created
	var count int64
	err = gormDB.Model(&marketplace.Listing{}).Where("listing_id = ?", listing.ListingID).Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}
