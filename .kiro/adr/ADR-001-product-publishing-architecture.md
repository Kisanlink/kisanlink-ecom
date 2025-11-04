# ADR-001: Product Publishing Architecture for FPO Marketplace

**Status**: Proposed
**Date**: 2025-11-03
**Author**: SDE-3 Backend Architect

## Context

KisanLink e-commerce platform needs to enable Super Admins to publish products to selected FPOs with access control, platform fees, and delivery costs. The system must support:
- Selective product visibility per FPO
- FPO-specific delivery costs
- Platform commission calculation
- Retail price calculation (base + delivery + commission)
- Price locking for 30-minute windows

## Decision Drivers

1. **Query Performance**: Efficient filtering of products visible to specific FPOs
2. **Data Integrity**: Maintaining consistency between product and publish states
3. **Scalability**: Supporting 100K+ products and 1K+ FPOs
4. **Maintainability**: Clear separation of concerns
5. **Backward Compatibility**: Minimal impact on existing systems

## Considered Options

### Option 1: JSONB Storage (Proposed Design)
Store FPO access list and delivery costs as JSONB in a single table.

**Pros:**
- Single table, fewer joins
- Flexible schema for future extensions
- GIN indexes provide good query performance
- Atomic updates for all publish data

**Cons:**
- JSONB operations slightly slower than normalized tables
- Potential for data inconsistency if not properly validated
- Limited referential integrity

### Option 2: Normalized Junction Tables
Create separate tables: `product_fpo_access` and `product_fpo_delivery_costs`.

**Pros:**
- Strong referential integrity
- Optimal for complex queries
- Clear relational model

**Cons:**
- Multiple joins required
- More complex transaction management
- Higher storage overhead
- More complex update operations

### Option 3: Hybrid Approach
Use JSONB for delivery costs, junction table for FPO access.

**Pros:**
- Balance between performance and integrity
- Optimized for most common query (access check)
- Flexible pricing data

**Cons:**
- Mixed patterns increase complexity
- Two different update mechanisms

## Decision

**Adopt Option 1: JSONB Storage** with the following modifications and safeguards:

### Modified Schema

```sql
CREATE TABLE publish_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES catalog_items(id) ON DELETE CASCADE,

    -- Publishing metadata
    is_published BOOLEAN NOT NULL DEFAULT false,
    published_at TIMESTAMP,
    published_by VARCHAR(255),

    -- FPO Access Control (Array of org IDs)
    fpo_access_list JSONB NOT NULL DEFAULT '[]'::jsonb,

    -- Pricing Configuration
    delivery_costs JSONB NOT NULL DEFAULT '{}'::jsonb,  -- Map: {fpo_id: {amount, currency, updated_at}}
    platform_fee_percent DECIMAL(5,2) NOT NULL DEFAULT 10.00,

    -- Price lock mechanism
    price_lock_duration_minutes INT NOT NULL DEFAULT 30,

    -- Audit fields
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255) NOT NULL,
    version INT NOT NULL DEFAULT 1,

    -- Soft delete
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),

    CONSTRAINT unique_product_publish UNIQUE(product_id) WHERE deleted_at IS NULL,
    CONSTRAINT valid_platform_fee CHECK (platform_fee_percent >= 0 AND platform_fee_percent <= 100),
    CONSTRAINT valid_price_lock CHECK (price_lock_duration_minutes > 0 AND price_lock_duration_minutes <= 1440)
);

-- Performance indexes
CREATE INDEX idx_publish_states_product_id ON publish_states(product_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_publish_states_published ON publish_states(is_published, published_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_publish_states_fpo_access ON publish_states USING GIN(fpo_access_list) WHERE deleted_at IS NULL AND is_published = true;
CREATE INDEX idx_publish_states_updated ON publish_states(updated_at DESC) WHERE deleted_at IS NULL;

-- Function to validate FPO access
CREATE OR REPLACE FUNCTION has_fpo_access(pub_state_id UUID, fpo_org_id TEXT)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM publish_states
        WHERE id = pub_state_id
        AND deleted_at IS NULL
        AND is_published = true
        AND fpo_access_list @> to_jsonb(fpo_org_id)
    );
END;
$$ LANGUAGE plpgsql IMMUTABLE;
```

## Implementation Guidelines

### 1. Service Layer Enhancements

```go
type PublishService struct {
    repo            PublishRepository
    catalogRepo     CatalogRepository
    orgValidator    OrganizationValidator
    eventPublisher  EventPublisher
    cache           CacheService
}

// Core business logic with validation
func (s *PublishService) PublishProduct(ctx context.Context, req PublishRequest) (*PublishState, error) {
    // 1. Validate product exists
    // 2. Validate FPO IDs are legitimate
    // 3. Begin transaction
    // 4. Create/update publish state
    // 5. Publish event
    // 6. Invalidate cache
    // 7. Commit transaction
}
```

### 2. Caching Strategy

```go
// Cache keys
const (
    FPOProductListKey = "fpo:products:%s"    // per FPO
    ProductPricingKey = "product:pricing:%s:%s" // product:fpo
    CacheTTL = 5 * time.Minute
)

// Implement cache-aside pattern
func (s *PublishService) GetFPOProducts(ctx context.Context, fpoID string) ([]*ProductWithPricing, error) {
    // Check cache first
    // If miss, query DB
    // Cache results
    // Return
}
```

### 3. Event Publishing

```go
type ProductPublishedEvent struct {
    EventMetadata
    ProductID     string
    FPOCount      int
    PublishedBy   string
    PublishedAt   time.Time
}

type FPOAccessGrantedEvent struct {
    EventMetadata
    ProductID     string
    FPOID         string
    DeliveryCost  decimal.Decimal
    RetailPrice   decimal.Decimal
}
```

### 4. Security Considerations

- **Authorization**: Only Super Admins can publish/unpublish
- **Validation**: FPO IDs must be validated against organization service
- **Audit Trail**: All publish operations logged with user context
- **Input Sanitization**: JSONB inputs validated for structure and content

### 5. Migration Path

```sql
-- Phase 1: Create tables and basic functionality
-- Phase 2: Migrate existing visibility data
UPDATE catalog_items
SET visibility = 'PRIVATE'
WHERE visibility = 'PUBLIC'
AND id IN (SELECT product_id FROM publish_states WHERE is_published = true);

-- Phase 3: Deprecate catalog_items.visibility field (future)
```

## Performance Optimizations

1. **Query Optimization**:
   - Use prepared statements
   - Implement query result pagination
   - Use database connection pooling

2. **Caching**:
   - Redis for FPO product lists
   - In-memory cache for platform fee configuration
   - Cache invalidation on publish events

3. **Database**:
   - Partial indexes for active records
   - GIN indexes for JSONB queries
   - Regular VACUUM and ANALYZE

## Monitoring & Observability

```go
// Metrics to track
- publish_operations_total{status, fpo_count}
- fpo_product_query_duration_seconds
- cache_hit_ratio{cache_type}
- retail_price_calculation_duration_ms
```

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| JSONB query performance degradation | High | Monitor query performance, add specialized indexes |
| Invalid FPO IDs in access list | Medium | Validate against organization service, periodic cleanup job |
| Cache inconsistency | Medium | Event-driven cache invalidation, short TTLs |
| Price calculation errors | High | Comprehensive testing, decimal precision handling |

## Consequences

### Positive
- Simplified data model reduces complexity
- Better performance for common queries
- Easier to implement and maintain
- Flexible for future enhancements

### Negative
- Weaker referential integrity (mitigated by validation)
- Slightly complex JSONB queries (mitigated by helper functions)
- Need for data validation layer

## References
- PostgreSQL JSONB Performance: https://www.postgresql.org/docs/current/datatype-json.html
- GIN Index Documentation: https://www.postgresql.org/docs/current/gin.html
- Decimal Handling in Go: https://github.com/shopspring/decimal
