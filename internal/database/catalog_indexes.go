package database

import (
	"context"
	"fmt"
	"log"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// CreateCatalogIndexes creates optimized indexes for catalog query performance
func CreateCatalogIndexes(dbManager db.DBManager) error {
	log.Println("Creating catalog performance indexes...")

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
		if err := gormDB.Exec(indexSQL).Error; err != nil {
			log.Printf("Warning: Failed to create catalog index %d: %v", i+1, err)
			log.Printf("SQL: %s", indexSQL)
			// Continue with other indexes instead of failing completely
		} else {
			successCount++
		}
	}

	log.Printf("Successfully created %d out of %d catalog performance indexes", successCount, len(indexes))
	return nil
}

// DropCatalogIndexes drops catalog-specific indexes (for testing/cleanup)
func DropCatalogIndexes(dbManager db.DBManager) error {
	log.Println("Dropping catalog performance indexes...")

	postgresManager, ok := dbManager.(*db.PostgresManager)
	if !ok {
		return fmt.Errorf("failed to cast to PostgresManager")
	}

	gormDB, err := postgresManager.GetDB(context.Background(), false)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Drop catalog-specific indexes
	dropStatements := []string{
		"DROP INDEX IF EXISTS idx_catalog_items_tenant_type_status_updated",
		"DROP INDEX IF EXISTS idx_catalog_items_tenant_category_active",
		"DROP INDEX IF EXISTS idx_catalog_items_tenant_vendor_active",
		"DROP INDEX IF EXISTS idx_catalog_items_tenant_visibility_active",
		"DROP INDEX IF EXISTS idx_catalog_items_name_trgm",
		"DROP INDEX IF EXISTS idx_catalog_items_description_trgm",
		"DROP INDEX IF EXISTS idx_catalog_items_name_fts",
		"DROP INDEX IF EXISTS idx_catalog_items_description_fts",
		"DROP INDEX IF EXISTS idx_catalog_items_combined_fts",
		"DROP INDEX IF EXISTS idx_catalog_items_attributes_gin",
		"DROP INDEX IF EXISTS idx_catalog_items_attributes_sku",
		"DROP INDEX IF EXISTS idx_catalog_items_attributes_brand",
		"DROP INDEX IF EXISTS idx_catalog_items_attributes_skills",
		"DROP INDEX IF EXISTS idx_catalog_items_attributes_certification",
		"DROP INDEX IF EXISTS idx_catalog_items_tags_gin",
		"DROP INDEX IF EXISTS idx_catalog_items_tags_array",
		"DROP INDEX IF EXISTS idx_catalog_items_price_range",
		"DROP INDEX IF EXISTS idx_catalog_items_tenant_price_range",
		"DROP INDEX IF EXISTS idx_categories_tenant_parent_active",
		"DROP INDEX IF EXISTS idx_categories_path_active",
		"DROP INDEX IF EXISTS idx_categories_level_active",
		"DROP INDEX IF EXISTS idx_categories_name_trgm",
		"DROP INDEX IF EXISTS idx_variants_tenant_catalog_active",
		"DROP INDEX IF EXISTS idx_variants_sku_tenant",
		"DROP INDEX IF EXISTS idx_variants_attributes_gin",
		"DROP INDEX IF EXISTS idx_variants_price_range",
		"DROP INDEX IF EXISTS idx_availability_catalog_item_active",
		"DROP INDEX IF EXISTS idx_media_catalog_item_type",
		"DROP INDEX IF EXISTS idx_inventory_lots_product_org_active",
		"DROP INDEX IF EXISTS idx_inventory_lots_expiry_active",
		"DROP INDEX IF EXISTS idx_slas_catalog_item_active",
		"DROP INDEX IF EXISTS idx_catalog_items_multi_filter",
		"DROP INDEX IF EXISTS idx_catalog_items_active_tenant_type",
		"DROP INDEX IF EXISTS idx_catalog_items_active_public",
		"DROP INDEX IF EXISTS idx_catalog_items_list_covering",
	}

	for i, statement := range dropStatements {
		if err := gormDB.Exec(statement).Error; err != nil {
			log.Printf("Warning: Failed to drop catalog index %d: %v", i+1, err)
		}
	}

	log.Println("Catalog performance indexes dropped successfully")
	return nil
}

// AnalyzeCatalogIndexUsage provides index usage statistics for catalog tables
func AnalyzeCatalogIndexUsage(dbManager db.DBManager) error {
	log.Println("Analyzing catalog index usage...")

	postgresManager, ok := dbManager.(*db.PostgresManager)
	if !ok {
		return fmt.Errorf("failed to cast to PostgresManager")
	}

	gormDB, err := postgresManager.GetDB(context.Background(), false)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Query to analyze index usage
	query := `
		SELECT
			schemaname,
			tablename,
			indexname,
			idx_tup_read,
			idx_tup_fetch,
			idx_scan
		FROM pg_stat_user_indexes
		WHERE tablename IN ('catalog_items', 'categories', 'variants', 'availability', 'inventory_lots', 'slas')
		ORDER BY idx_scan DESC, idx_tup_read DESC
	`

	rows, err := gormDB.Raw(query).Rows()
	if err != nil {
		return fmt.Errorf("failed to analyze index usage: %w", err)
	}
	defer rows.Close()

	log.Println("Catalog Index Usage Statistics:")
	log.Println("================================")
	log.Printf("%-20s %-20s %-40s %-12s %-12s %-12s",
		"Schema", "Table", "Index", "Tup Read", "Tup Fetch", "Scans")
	log.Println("------------------------------------------------------------------------------------------------")

	for rows.Next() {
		var schema, table, index string
		var tupRead, tupFetch, scans int64

		if err := rows.Scan(&schema, &table, &index, &tupRead, &tupFetch, &scans); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		log.Printf("%-20s %-20s %-40s %-12d %-12d %-12d",
			schema, table, index, tupRead, tupFetch, scans)
	}

	return nil
}
