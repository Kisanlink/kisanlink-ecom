# Architecture Decision Record: Fix Duplicate ID Generation Issue

**Date:** 2025-11-03
**Status:** Proposed
**Author:** Backend Architecture Team

## Executive Summary

Critical production issue causing duplicate primary key violations across multiple tables due to conflicting ID generation patterns and potential GORM/kisanlink-db double-insert behavior. This ADR proposes a comprehensive fix to establish a single, reliable ID generation strategy.

## 1. Problem Analysis

### Root Cause Identification

#### A. Dual ID Generation Anti-Pattern

The codebase exhibits a critical anti-pattern where IDs are generated twice:

1. **First Generation** (in model constructor):

```go
// In NewOrderItem() at entities/models/orders/order.go:131
baseModel := base.NewBaseModel("ITEM", "large")  // Generates ID like "ITEM-abc123..."
```

2. **Second Override** (in service layer):

```go
// In OrderService.CreateOrder() at internal/services/orders/order_service.go:121
itemID, err := s.sequenceSvc.GenerateID(ctx, "ITEM", &req.SellerOrganizationID)
// Then at line 132:
baseModel.ID = id  // Overrides with "ITEM-20251103144209-000001"
```

This creates a race condition where the original ID might be used before being overridden.

#### B. Sequence Service Collision Risk

The sequence service generates IDs with format: `{PREFIX}-{TIMESTAMP}-{SEQUENCE}`

- Example: `ITEM-20251103144209-000001`
- **Problem**: Timestamp precision is only to seconds (14 digits)
- **Risk**: Multiple requests in the same second get identical timestamps
- **Mitigation**: Only the 6-digit sequence number prevents collisions

#### C. Potential Double-Insert Issue

Based on the error logs showing two INSERT attempts:

1. First INSERT with `ON CONFLICT` clause (succeeds)
2. Second INSERT without `ON CONFLICT` clause (fails)

This suggests either:

- GORM hooks causing double execution
- kisanlink-db transaction handling issues
- Repository pattern creating nested inserts

### Affected Entities

All entities using `base.NewBaseModel()` with ID override pattern:

| Entity             | File                                     | Line    | Override Pattern              |
| ------------------ | ---------------------------------------- | ------- | ----------------------------- |
| OrderItem          | entities/models/orders/order.go          | 131-132 | YES - Overrides with sequence |
| OrderStatusHistory | entities/models/orders/order.go          | 178     | NO - Uses base ID             |
| Order              | entities/models/orders/order.go          | 77      | NO - Uses base ID             |
| Listing            | entities/models/marketplace/listing.go   | 113-114 | Partial - Has ListingID field |
| Bid                | entities/models/marketplace/bid.go       | 62-63   | Partial - Has BidID field     |
| InventoryLot       | entities/models/catalog/inventory_lot.go | 107     | NO - Uses base ID             |
| CatalogItem        | entities/models/catalog/catalog_item.go  | 173     | NO - Uses base ID             |
| All other entities | Various                                  | -       | NO - Uses base ID             |

## 2. Proposed Solution

### Option A: Database-Generated IDs (RECOMMENDED)

**Approach:** Use PostgreSQL's native ID generation capabilities

**Implementation:**

```sql
-- Use BIGSERIAL for auto-increment
ALTER TABLE order_items
  ALTER COLUMN id TYPE BIGINT USING (nextval('order_items_id_seq'));

-- Or use UUID with extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
ALTER TABLE order_items
  ALTER COLUMN id SET DEFAULT uuid_generate_v4();
```

**Pros:**

- Guaranteed uniqueness at database level
- No application-level coordination needed
- Atomic and thread-safe
- Best performance (no round trips)
- Standard industry practice

**Cons:**

- Loss of readable prefixes (ITEM-, ORDER-, etc.)
- Migration complexity for existing data
- Need to handle ID retrieval after insert

### Option B: Fix Application-Level ID Generation

**Approach:** Choose ONE ID generation method and use consistently

**Sub-option B1: Use Only BaseModel IDs**

```go
func NewOrderItem(...) *OrderItem {
    // Remove the ID parameter entirely
    return &OrderItem{
        BaseModel: *base.NewBaseModel("ITEM", "large"),
        // Don't override ID
        // ... other fields
    }
}
```

**Sub-option B2: Use Only Sequence Service**

```go
func NewOrderItem(id string, ...) *OrderItem {
    // Don't call NewBaseModel with prefix
    return &OrderItem{
        BaseModel: base.BaseModel{
            ID: id, // Use provided sequence ID
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        },
        // ... other fields
    }
}
```

### Option C: Hybrid Approach (Application + Database)

**Approach:** Use application-generated UUIDs with database constraints

```go
// In base model
type BaseModel struct {
    ID        string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

This provides fallback safety - application generates UUID, database validates/generates if missing.

## 3. Implementation Plan

### Phase 1: Stop the Bleeding (Immediate)

1. **Hot Fix** for OrderItem only:

```go
// In order_service.go, remove sequence generation for items
// Let BaseModel handle ID generation
orderItem := orderModels.NewOrderItem(
    // Remove ID parameter
    ord.ID,
    itemReq.CatalogItemID,
    // ... rest of parameters
)
```

### Phase 2: Audit and Standardize (Week 1)

1. Audit all 25+ entities for ID generation patterns
2. Create ID generation policy document
3. Update model constructors to follow single pattern
4. Add lint rules to prevent dual generation

### Phase 3: Database Migration (Week 2)

```sql
-- Add proper constraints and defaults
ALTER TABLE order_items
  ADD CONSTRAINT order_items_id_format
  CHECK (id ~ '^[A-Z]+-[0-9a-f]+$');

-- Create standardized ID generation function
CREATE OR REPLACE FUNCTION generate_entity_id(prefix TEXT)
RETURNS TEXT AS $$
BEGIN
  RETURN prefix || '-' || gen_random_uuid();
END;
$$ LANGUAGE plpgsql;

-- Apply to tables
ALTER TABLE order_items
  ALTER COLUMN id SET DEFAULT generate_entity_id('ITEM');
```

### Phase 4: Fix Double-Insert Issue (Week 2)

1. Investigate kisanlink-db Create method implementation
2. Check for GORM hooks or middleware
3. Review transaction handling in repositories
4. Add integration tests for concurrent inserts

## 4. Migration Strategy

### Data Migration

```sql
-- Backup existing data
CREATE TABLE order_items_backup AS SELECT * FROM order_items;

-- Update existing IDs if changing format
UPDATE order_items
SET id = 'ITEM-' || gen_random_uuid()::text
WHERE id !~ '^ITEM-[0-9a-f-]+$';
```

### Code Migration Order

1. **Critical Path** (Day 1):
   - OrderItem
   - OrderStatusHistory
   - Order

2. **High Traffic** (Day 2-3):
   - MarketplaceBid
   - MarketplaceListing
   - InventoryLot

3. **Remaining Entities** (Week 2):
   - All other catalog items
   - Supporting entities

### Rollback Plan

```sql
-- If issues arise, restore from backup
TRUNCATE order_items;
INSERT INTO order_items SELECT * FROM order_items_backup;

-- Or use point-in-time recovery
-- pg_restore --dbname=kisanlink_ecom --clean --create backup_20251103.dump
```

## 5. Testing Strategy

### Unit Tests

```go
func TestOrderItemIDGeneration(t *testing.T) {
    // Test no duplicate IDs in 10000 concurrent creates
    ids := sync.Map{}
    wg := sync.WaitGroup{}

    for i := 0; i < 10000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            item := NewOrderItem(...)
            if _, exists := ids.LoadOrStore(item.ID, true); exists {
                t.Errorf("Duplicate ID generated: %s", item.ID)
            }
        }()
    }
    wg.Wait()
}
```

### Integration Tests

```go
func TestConcurrentOrderCreation(t *testing.T) {
    // Test 100 concurrent order creations
    // Verify no duplicate key violations
    // Check all items are created exactly once
}
```

### Load Tests

- Simulate 1000 orders/second
- Monitor for duplicate key violations
- Check database constraint violations
- Verify sequence counter increments

## 6. Monitoring & Alerts

### Metrics to Track

```yaml
metrics:
  - name: duplicate_key_violations
    query: |
      SELECT COUNT(*)
      FROM pg_stat_database_conflicts
      WHERE conflict_type = 'duplicate_key'
    threshold: 0

  - name: failed_inserts
    query: |
      SELECT COUNT(*)
      FROM pg_stat_user_tables
      WHERE n_tup_ins_failed > 0
    threshold: 0
```

### Alerts to Configure

```yaml
alerts:
  - name: DuplicateKeyViolation
    condition: rate(postgres_errors{type="duplicate_key"}[1m]) > 0
    severity: critical
    action: page_oncall

  - name: HighInsertFailureRate
    condition: rate(insert_failures[5m]) > 0.01
    severity: warning
    action: notify_slack
```

## 7. Decision Matrix

| Criteria                 | Option A (DB IDs) | Option B (App IDs) | Option C (Hybrid) |
| ------------------------ | ----------------- | ------------------ | ----------------- |
| **Uniqueness Guarantee** | ✅ Excellent      | ⚠️ Good            | ✅ Excellent      |
| **Performance**          | ✅ Best           | ⚠️ Good            | ✅ Very Good      |
| **Readable IDs**         | ❌ No             | ✅ Yes             | ⚠️ Partial        |
| **Migration Effort**     | ❌ High           | ✅ Low             | ⚠️ Medium         |
| **Maintenance**          | ✅ Low            | ⚠️ Medium          | ⚠️ Medium         |
| **Industry Standard**    | ✅ Yes            | ⚠️ Partial         | ✅ Yes            |

## 8. Recommendation

**Immediate Action (Today):**

1. Implement Phase 1 hot fix for OrderItem
2. Deploy with monitoring
3. Verify duplicate errors stop

**Short Term (This Week):**

1. Adopt Option A (Database-Generated IDs) for new tables
2. Implement Option B2 (Sequence-only) for existing tables as interim fix
3. Add comprehensive testing

**Long Term (This Month):**

1. Migrate all entities to Option A
2. Deprecate application-level ID generation
3. Implement database-level ID generation function with prefixes

## 9. Risk Assessment

### High Risks

- **Data Loss**: Mitigated by comprehensive backups
- **Service Disruption**: Mitigated by phased rollout
- **ID Format Changes**: Mitigated by maintaining prefixes

### Medium Risks

- **Performance Impact**: Monitor during migration
- **Integration Issues**: Test with downstream services
- **Rollback Complexity**: Practice rollback procedures

### Low Risks

- **Developer Confusion**: Document thoroughly
- **Testing Gaps**: Increase test coverage

## 10. Success Criteria

- Zero duplicate key violations in production
- All tests passing with concurrent operations
- Performance metrics within 5% of baseline
- No data loss or corruption
- Clean rollback capability demonstrated

## Appendix: Affected Files

### Critical Files to Modify

```
/Users/kaushik/kisanlink-ecom/entities/models/orders/order.go:128-149
/Users/kaushik/kisanlink-ecom/internal/services/orders/order_service.go:120-124
/Users/kaushik/kisanlink-ecom/internal/repositories/orders/order_repository.go:104-110
```

### Additional Files Requiring Review

```
/Users/kaushik/kisanlink-ecom/entities/models/marketplace/listing.go
/Users/kaushik/kisanlink-ecom/entities/models/marketplace/bid.go
/Users/kaushik/kisanlink-ecom/internal/services/sequence/sequence_service.go
All 25+ model files using base.NewBaseModel()
```

## References

- PostgreSQL UUID Documentation: https://www.postgresql.org/docs/current/datatype-uuid.html
- GORM ID Generation: https://gorm.io/docs/create.html
- Distributed ID Generation: https://instagram-engineering.com/sharding-ids-at-instagram-1cf5a71e5a5c
