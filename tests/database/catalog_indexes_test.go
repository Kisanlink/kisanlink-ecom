//go:build integration
// +build integration

package database

import (
	"context"
	"testing"

	"github.com/Kisanlink/kisanlink-ecom/internal/database"
	"github.com/Kisanlink/kisanlink-ecom/tests/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCatalogIndexes(t *testing.T) {
	// Setup test database
	dbManager, cleanup := testutils.SetupTestDatabase(t)
	defer cleanup()

	// Skip if no real database manager available
	if dbManager == nil {
		t.Skip("Real database manager not available for integration test")
	}

	// Create catalog indexes
	err := database.CreateCatalogIndexes(dbManager)
	require.NoError(t, err, "Failed to create catalog indexes")

	// Verify indexes were created by checking pg_indexes
	postgresManager := testutils.GetPostgresManager(t, dbManager)
	db, err := postgresManager.GetDB(context.Background(), false)
	require.NoError(t, err)

	// Test some key indexes exist
	expectedIndexes := []string{
		"idx_catalog_items_tenant_type_status_updated",
		"idx_catalog_items_name_trgm",
		"idx_catalog_items_description_trgm",
		"idx_catalog_items_attributes_gin",
		"idx_catalog_items_tags_gin",
		"idx_categories_tenant_parent_active",
		"idx_variants_tenant_catalog_active",
	}

	for _, indexName := range expectedIndexes {
		var exists bool
		query := `
			SELECT EXISTS (
				SELECT 1 FROM pg_indexes
				WHERE indexname = ?
			)
		`
		err := db.Raw(query, indexName).Scan(&exists).Error
		require.NoError(t, err, "Failed to check index existence")
		assert.True(t, exists, "Index %s should exist", indexName)
	}
}

func TestCatalogIndexPerformance(t *testing.T) {
	// Setup test database with sample data
	dbManager, cleanup := testutils.SetupTestDatabase(t)
	defer cleanup()

	// Skip if no real database manager available
	if dbManager == nil {
		t.Skip("Real database manager not available for integration test")
	}

	// Create catalog indexes
	err := database.CreateCatalogIndexes(dbManager)
	require.NoError(t, err)

	postgresManager := testutils.GetPostgresManager(t, dbManager)
	db, err := postgresManager.GetDB(context.Background(), false)
	require.NoError(t, err)

	// Create sample catalog_items table for testing
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS catalog_items (
			id VARCHAR(255) PRIMARY KEY,
			organization_id VARCHAR(255) NOT NULL,
			item_type VARCHAR(20) NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			category_id VARCHAR(255),
			vendor_id VARCHAR(255),
			is_active BOOLEAN NOT NULL DEFAULT true,
			visibility VARCHAR(20) NOT NULL DEFAULT 'PRIVATE',
			base_price DECIMAL(12,2) NOT NULL,
			currency VARCHAR(3) NOT NULL DEFAULT 'INR',
			tags TEXT[],
			attributes JSONB,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)
	`
	err = db.Exec(createTableSQL).Error
	require.NoError(t, err)

	// Insert sample data
	insertSQL := `
		INSERT INTO catalog_items (
			id, organization_id, item_type, name, description,
			category_id, vendor_id, is_active, visibility,
			base_price, currency, tags, attributes
		) VALUES
		('cat1', 'org1', 'PRODUCT', 'Test Product 1', 'A test product for searching',
		 'cat_food', 'vendor1', true, 'PUBLIC', 100.00, 'INR',
		 ARRAY['organic', 'fresh'], '{"sku": "TEST001", "brand": "TestBrand"}'),
		('cat2', 'org1', 'SERVICE', 'Test Service 1', 'A test service for filtering',
		 'cat_service', 'vendor2', true, 'PRIVATE', 200.00, 'INR',
		 ARRAY['professional', 'certified'], '{"skills": ["farming", "irrigation"]}'),
		('cat3', 'org2', 'LABOUR', 'Test Labour 1', 'A test labour for indexing',
		 'cat_labour', 'vendor3', true, 'PUBLIC', 50.00, 'INR',
		 ARRAY['skilled', 'experienced'], '{"certification": ["organic_farming"]}')
	`
	err = db.Exec(insertSQL).Error
	require.NoError(t, err)

	// Test composite index usage for tenant + type + status queries
	t.Run("CompositeIndexUsage", func(t *testing.T) {
		var count int64
		query := `
			SELECT COUNT(*) FROM catalog_items
			WHERE organization_id = 'org1'
			AND item_type = 'PRODUCT'
			AND is_active = true
		`
		err := db.Raw(query).Scan(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	// Test trigram search functionality
	t.Run("TrigramSearch", func(t *testing.T) {
		var count int64
		query := `
			SELECT COUNT(*) FROM catalog_items
			WHERE name % 'Product'
		`
		err := db.Raw(query).Scan(&count).Error
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(1))
	})

	// Test full-text search functionality
	t.Run("FullTextSearch", func(t *testing.T) {
		var count int64
		query := `
			SELECT COUNT(*) FROM catalog_items
			WHERE to_tsvector('english', name) @@ to_tsquery('english', 'test')
		`
		err := db.Raw(query).Scan(&count).Error
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(1))
	})

	// Test JSONB attribute queries
	t.Run("JSONBAttributeSearch", func(t *testing.T) {
		var count int64
		query := `
			SELECT COUNT(*) FROM catalog_items
			WHERE attributes->>'sku' = 'TEST001'
		`
		err := db.Raw(query).Scan(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	// Test array tag queries
	t.Run("ArrayTagSearch", func(t *testing.T) {
		var count int64
		query := `
			SELECT COUNT(*) FROM catalog_items
			WHERE tags @> ARRAY['organic']
		`
		err := db.Raw(query).Scan(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	// Test price range queries
	t.Run("PriceRangeQuery", func(t *testing.T) {
		var count int64
		query := `
			SELECT COUNT(*) FROM catalog_items
			WHERE base_price BETWEEN 50.00 AND 150.00
			AND currency = 'INR'
			AND is_active = true
		`
		err := db.Raw(query).Scan(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})
}

func TestAnalyzeCatalogIndexUsage(t *testing.T) {
	// Setup test database
	dbManager, cleanup := testutils.SetupTestDatabase(t)
	defer cleanup()

	// Skip if no real database manager available
	if dbManager == nil {
		t.Skip("Real database manager not available for integration test")
	}

	// Create catalog indexes
	err := database.CreateCatalogIndexes(dbManager)
	require.NoError(t, err)

	// Analyze index usage (should not error even with no usage stats)
	err = database.AnalyzeCatalogIndexUsage(dbManager)
	assert.NoError(t, err, "Index usage analysis should not fail")
}

func TestDropCatalogIndexes(t *testing.T) {
	// Setup test database
	dbManager, cleanup := testutils.SetupTestDatabase(t)
	defer cleanup()

	// Skip if no real database manager available
	if dbManager == nil {
		t.Skip("Real database manager not available for integration test")
	}

	// Create catalog indexes first
	err := database.CreateCatalogIndexes(dbManager)
	require.NoError(t, err)

	// Drop catalog indexes
	err = database.DropCatalogIndexes(dbManager)
	assert.NoError(t, err, "Failed to drop catalog indexes")

	// Verify some indexes were dropped
	postgresManager := testutils.GetPostgresManager(t, dbManager)
	db, err := postgresManager.GetDB(context.Background(), false)
	require.NoError(t, err)

	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM pg_indexes
			WHERE indexname = 'idx_catalog_items_tenant_type_status_updated'
		)
	`
	err = db.Raw(query).Scan(&exists).Error
	require.NoError(t, err)
	assert.False(t, exists, "Index should have been dropped")
}

func TestIdempotentCatalogIndexCreation(t *testing.T) {
	// Setup test database
	dbManager, cleanup := testutils.SetupTestDatabase(t)
	defer cleanup()

	// Skip if no real database manager available
	if dbManager == nil {
		t.Skip("Real database manager not available for integration test")
	}

	// Create catalog indexes multiple times (should be idempotent)
	for i := 0; i < 3; i++ {
		err := database.CreateCatalogIndexes(dbManager)
		assert.NoError(t, err, "Catalog index creation should be idempotent (iteration %d)", i+1)
	}

	// Verify indexes still exist after multiple creations
	postgresManager := testutils.GetPostgresManager(t, dbManager)
	db, err := postgresManager.GetDB(context.Background(), false)
	require.NoError(t, err)

	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM pg_indexes
			WHERE indexname = 'idx_catalog_items_name_trgm'
		)
	`
	err = db.Raw(query).Scan(&exists).Error
	require.NoError(t, err)
	assert.True(t, exists, "Index should exist after idempotent creation")
}
