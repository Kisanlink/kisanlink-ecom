# Catalog Database Indexes

This document describes the database indexes implemented for the marketplace catalog system to optimize query performance.

## Overview

The catalog indexing strategy focuses on optimizing the most common query patterns:

- Tenant-based filtering (multi-tenancy)
- Catalog type filtering (Product, Service, Labour, Contract)
- Status and visibility filtering
- Full-text search on names and descriptions
- JSONB attribute queries
- Array-based tag searches
- Price range filtering

## Index Categories

### 1. Composite Indexes

These indexes optimize multi-column queries that are common in catalog filtering:

```sql
-- Primary composite index for tenant + type + status + ordering
CREATE INDEX idx_catalog_items_tenant_type_status_updated
ON catalog_items(organization_id, item_type, is_active, updated_at DESC);

-- Tenant + category filtering
CREATE INDEX idx_catalog_items_tenant_category_active
ON catalog_items(organization_id, category_id, is_active);

-- Tenant + vendor filtering
CREATE INDEX idx_catalog_items_tenant_vendor_active
ON catalog_items(organization_id, vendor_id, is_active);

-- Tenant + visibility filtering
CREATE INDEX idx_catalog_items_tenant_visibility_active
ON catalog_items(organization_id, visibility, is_active);
```

**Use Cases:**

- List catalogs by organization and type
- Filter active catalogs by category
- Find vendor-specific catalogs
- Public/private catalog visibility queries

### 2. Full-Text Search Indexes

These indexes enable fast text search across catalog names and descriptions:

```sql
-- Trigram indexes for fuzzy matching
CREATE INDEX idx_catalog_items_name_trgm
ON catalog_items USING gin(name gin_trgm_ops);

CREATE INDEX idx_catalog_items_description_trgm
ON catalog_items USING gin(description gin_trgm_ops);

-- Full-text search indexes
CREATE INDEX idx_catalog_items_name_fts
ON catalog_items USING gin(to_tsvector('english', name));

CREATE INDEX idx_catalog_items_description_fts
ON catalog_items USING gin(to_tsvector('english', description));

-- Combined search index
CREATE INDEX idx_catalog_items_combined_fts
ON catalog_items USING gin(to_tsvector('english', coalesce(name, '') || ' ' || coalesce(description, '')));
```

**Use Cases:**

- Search catalogs by name similarity (`name % 'search_term'`)
- Full-text search across names and descriptions
- Fuzzy matching for typo tolerance
- Combined name + description searches

### 3. JSONB Indexes

These indexes optimize queries on the flexible attributes field:

```sql
-- General JSONB index
CREATE INDEX idx_catalog_items_attributes_gin
ON catalog_items USING gin(attributes);

-- Specific attribute path indexes
CREATE INDEX idx_catalog_items_attributes_sku
ON catalog_items USING gin((attributes->'sku'));

CREATE INDEX idx_catalog_items_attributes_brand
ON catalog_items USING gin((attributes->'brand'));

CREATE INDEX idx_catalog_items_attributes_skills
ON catalog_items USING gin((attributes->'skills'));

CREATE INDEX idx_catalog_items_attributes_certification
ON catalog_items USING gin((attributes->'certification'));
```

**Use Cases:**

- Filter by product SKU: `attributes->>'sku' = 'PROD001'`
- Filter by brand: `attributes->>'brand' = 'BrandName'`
- Filter by skills: `attributes->'skills' @> '["farming"]'`
- Filter by certifications: `attributes ? 'organic_certification'`

### 4. Array Indexes

These indexes optimize tag-based filtering:

```sql
-- GIN index for tag arrays
CREATE INDEX idx_catalog_items_tags_gin
ON catalog_items USING gin(tags);

-- Array operators optimization
CREATE INDEX idx_catalog_items_tags_array
ON catalog_items USING gin(tags array_ops);
```

**Use Cases:**

- Find catalogs with specific tags: `tags @> ARRAY['organic']`
- Find catalogs with any of multiple tags: `tags && ARRAY['fresh', 'local']`
- Check if catalog has tag: `'organic' = ANY(tags)`

### 5. Price Range Indexes

These indexes optimize price-based filtering and sorting:

```sql
-- Price range queries
CREATE INDEX idx_catalog_items_price_range
ON catalog_items(base_price, currency, is_active);

-- Tenant-specific price queries
CREATE INDEX idx_catalog_items_tenant_price_range
ON catalog_items(organization_id, base_price, currency, is_active);
```

**Use Cases:**

- Price range filtering: `base_price BETWEEN 100 AND 500`
- Currency-specific queries: `currency = 'INR' AND base_price < 1000`
- Tenant price comparisons

### 6. Category Hierarchy Indexes

These indexes optimize category navigation and filtering:

```sql
-- Category hierarchy navigation
CREATE INDEX idx_categories_tenant_parent_active
ON categories(organization_id, parent_id, is_active);

-- Path-based queries
CREATE INDEX idx_categories_path_active
ON categories(path, is_active);

-- Level-based queries
CREATE INDEX idx_categories_level_active
ON categories(level, is_active);

-- Category name search
CREATE INDEX idx_categories_name_trgm
ON categories USING gin(name gin_trgm_ops);
```

**Use Cases:**

- Navigate category tree: `parent_id = 'cat_electronics'`
- Path-based filtering: `path LIKE '/electronics/mobile%'`
- Level-based queries: `level = 2`
- Category name search

### 7. Variant Indexes

These indexes optimize product variant queries:

```sql
-- Variant filtering by catalog item
CREATE INDEX idx_variants_tenant_catalog_active
ON variants(organization_id, catalog_item_id, is_active);

-- SKU-based variant lookup
CREATE INDEX idx_variants_sku_tenant
ON variants(sku, organization_id);

-- Variant attributes
CREATE INDEX idx_variants_attributes_gin
ON variants USING gin(attributes);

-- Variant price filtering
CREATE INDEX idx_variants_price_range
ON variants(price, currency, is_active);
```

**Use Cases:**

- Find variants for a catalog item
- SKU-based variant lookup
- Variant attribute filtering
- Variant price comparisons

### 8. Performance Optimization Indexes

These indexes provide additional performance optimizations:

```sql
-- Multi-column filtering
CREATE INDEX idx_catalog_items_multi_filter
ON catalog_items(organization_id, item_type, category_id, is_active, visibility, updated_at DESC);

-- Partial index for active items only
CREATE INDEX idx_catalog_items_active_tenant_type
ON catalog_items(organization_id, item_type, updated_at DESC)
WHERE is_active = true;

-- Partial index for public catalogs
CREATE INDEX idx_catalog_items_active_public
ON catalog_items(item_type, category_id, updated_at DESC)
WHERE is_active = true AND visibility = 'PUBLIC';

-- Covering index for list queries
CREATE INDEX idx_catalog_items_list_covering
ON catalog_items(organization_id, is_active, item_type)
INCLUDE (name, base_price, currency, visibility, updated_at);
```

**Use Cases:**

- Complex multi-filter queries
- Active-only catalog listings (most common case)
- Public catalog browsing
- List queries with minimal data fetching

## Query Optimization Examples

### 1. Tenant Catalog Listing

```sql
-- Optimized by: idx_catalog_items_tenant_type_status_updated
SELECT id, name, base_price, currency, updated_at
FROM catalog_items
WHERE organization_id = 'org123'
  AND item_type = 'PRODUCT'
  AND is_active = true
ORDER BY updated_at DESC
LIMIT 20;
```

### 2. Full-Text Search

```sql
-- Optimized by: idx_catalog_items_combined_fts
SELECT id, name, description
FROM catalog_items
WHERE to_tsvector('english', coalesce(name, '') || ' ' || coalesce(description, ''))
      @@ to_tsquery('english', 'organic & farming');
```

### 3. Attribute Filtering

```sql
-- Optimized by: idx_catalog_items_attributes_sku
SELECT id, name, attributes
FROM catalog_items
WHERE attributes->>'sku' = 'PROD001'
  AND organization_id = 'org123';
```

### 4. Tag-Based Filtering

```sql
-- Optimized by: idx_catalog_items_tags_gin
SELECT id, name, tags
FROM catalog_items
WHERE tags @> ARRAY['organic', 'certified']
  AND is_active = true;
```

### 5. Price Range Queries

```sql
-- Optimized by: idx_catalog_items_tenant_price_range
SELECT id, name, base_price
FROM catalog_items
WHERE organization_id = 'org123'
  AND base_price BETWEEN 100.00 AND 500.00
  AND currency = 'INR'
  AND is_active = true;
```

## Index Maintenance

### Monitoring Index Usage

```sql
-- Check index usage statistics
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
WHERE tablename IN ('catalog_items', 'categories', 'variants')
ORDER BY idx_scan DESC;
```

### Index Size Monitoring

```sql
-- Check index sizes
SELECT
    indexname,
    pg_size_pretty(pg_relation_size(indexname::regclass)) as size
FROM pg_indexes
WHERE tablename IN ('catalog_items', 'categories', 'variants')
ORDER BY pg_relation_size(indexname::regclass) DESC;
```

### Analyzing Query Performance

```sql
-- Use EXPLAIN ANALYZE to verify index usage
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM catalog_items
WHERE organization_id = 'org123'
  AND item_type = 'PRODUCT'
  AND is_active = true
ORDER BY updated_at DESC;
```

## Best Practices

1. **Index Selectivity**: Indexes are most effective when they filter out a large percentage of rows
2. **Column Order**: In composite indexes, put the most selective columns first
3. **Partial Indexes**: Use WHERE clauses in indexes to reduce size and improve performance
4. **Covering Indexes**: Include frequently accessed columns to avoid table lookups
5. **Regular Monitoring**: Monitor index usage and performance regularly
6. **Maintenance**: Run VACUUM and ANALYZE regularly to keep statistics up to date

## Extensions Required

The following PostgreSQL extensions are required for optimal performance:

```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;    -- Trigram matching
CREATE EXTENSION IF NOT EXISTS btree_gin;  -- GIN indexes on btree types
```

## Performance Targets

With these indexes in place, the system should achieve:

- P95 ≤ 120ms for catalog list queries
- P95 ≤ 50ms for single catalog lookups
- P95 ≤ 200ms for full-text search queries
- Cache hit ratio ≥ 70% for hot endpoints
- Index scan ratio ≥ 95% for filtered queries
