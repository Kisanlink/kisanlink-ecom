# Implementation Tasks: Fix Duplicate ID Generation Issue

## Immediate Hot Fix (Priority 1 - Deploy Today)

### Task 1: Fix OrderItem ID Generation

**File:** `/Users/kaushik/kisanlink-ecom/entities/models/orders/order.go`
**Lines:** 128-149

**Current Code:**

```go
func NewOrderItem(id, orderID, catalogItemID, catalogItemType, catalogItemName, catalogItemSKU string, quantity, unitPrice decimal.Decimal) *OrderItem {
    baseModel := base.NewBaseModel("ITEM", "large")
    baseModel.ID = id // Override the auto-generated ID with sequence-based ID
    // ...
}
```

**Fixed Code:**

```go
func NewOrderItem(orderID, catalogItemID, catalogItemType, catalogItemName, catalogItemSKU string, quantity, unitPrice decimal.Decimal) *OrderItem {
    // Remove id parameter, let BaseModel generate it
    return &OrderItem{
        BaseModel:       *base.NewBaseModel("ITEM", "large"),
        OrderID:         orderID,
        CatalogItemID:   catalogItemID,
        CatalogItemType: catalogItemType,
        CatalogItemName: catalogItemName,
        CatalogItemSKU:  catalogItemSKU,
        Quantity:        quantity,
        UnitPrice:       unitPrice,
        TotalPrice:      quantity.Mul(unitPrice),
        TaxRate:         decimal.Zero,
        TaxAmount:       decimal.Zero,
        DiscountRate:    decimal.Zero,
        DiscountAmount:  decimal.Zero,
    }
}
```

### Task 2: Remove Sequence Service Call for OrderItems

**File:** `/Users/kaushik/kisanlink-ecom/internal/services/orders/order_service.go`
**Lines:** 120-144

**Current Code:**

```go
// Generate unique ID for order item using sequence service
itemID, err := s.sequenceSvc.GenerateID(ctx, "ITEM", &req.SellerOrganizationID)
if err != nil {
    return nil, fmt.Errorf("failed to generate order item ID: %w", err)
}

// Create order item with proper calculations
quantity := itemReq.Quantity
orderItem := orderModels.NewOrderItem(
    itemID, // Use sequence-based ID instead of timestamp hash
    ord.ID,
    // ...
)
```

**Fixed Code:**

```go
// Create order item with proper calculations
quantity := itemReq.Quantity
orderItem := orderModels.NewOrderItem(
    // Remove itemID parameter
    ord.ID,
    itemReq.CatalogItemID,
    itemReq.CatalogItemType,
    catalogItem.Name,
    catalogItem.SKU,
    quantity,
    unitPrice,
)
```

### Task 3: Fix Repository Create Method

**File:** `/Users/kaushik/kisanlink-ecom/internal/repositories/orders/order_repository.go`
**Lines:** 104-110

**Investigation Required:**

- Check if `r.dbManager.Create()` is being called multiple times
- Verify transaction handling
- Look for any GORM hooks

## Phase 2: Comprehensive Audit (This Week)

### Task 4: Audit All Entity Constructors

Create a script to find all instances of the pattern:

```bash
grep -r "baseModel.ID = " --include="*.go" .
grep -r "NewBaseModel(" --include="*.go" . | grep -v "BaseModel:"
```

### Task 5: Document ID Generation Policy

Create `/Users/kaushik/kisanlink-ecom/.kiro/standards/id-generation-policy.md`

**Content:**

1. Standard ID format: `{PREFIX}-{UUID}`
2. Approved prefixes per entity
3. No dual generation allowed
4. Database constraints required

### Task 6: Add Linter Rules

**File:** `/Users/kaushik/kisanlink-ecom/.golangci.yml`

Add custom linter to detect:

- `baseModel.ID = ` pattern
- Multiple ID generation in same function
- Missing ID validation

## Phase 3: Database Constraints (Week 2)

### Task 7: Add Database Constraints

**File:** Create migration `/Users/kaushik/kisanlink-ecom/migrations/fix_duplicate_ids.sql`

```sql
-- Add unique constraints if missing
ALTER TABLE order_items ADD CONSTRAINT order_items_id_unique UNIQUE (id);
ALTER TABLE order_status_history ADD CONSTRAINT order_status_history_id_unique UNIQUE (id);

-- Add format validation
ALTER TABLE order_items
  ADD CONSTRAINT order_items_id_format
  CHECK (id ~ '^ITEM-[a-zA-Z0-9-]+$');

-- Create ID generation function
CREATE OR REPLACE FUNCTION generate_prefixed_id(prefix TEXT)
RETURNS TEXT AS $$
BEGIN
  RETURN prefix || '-' || gen_random_uuid();
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Set defaults
ALTER TABLE order_items
  ALTER COLUMN id SET DEFAULT generate_prefixed_id('ITEM');
```

## Phase 4: Testing (Ongoing)

### Task 8: Unit Tests for ID Generation

**File:** Create `/Users/kaushik/kisanlink-ecom/tests/unit/id_generation_test.go`

Test cases:

1. Concurrent ID generation (10,000 goroutines)
2. ID format validation
3. Uniqueness guarantee
4. Performance benchmark

### Task 9: Integration Tests

**File:** Create `/Users/kaushik/kisanlink-ecom/tests/integration/order_creation_test.go`

Test cases:

1. Concurrent order creation (100 parallel)
2. Verify no duplicate violations
3. Check transaction rollback
4. Validate all items created

### Task 10: Load Testing

Use k6 or similar:

```javascript
import http from "k6/http";
import { check } from "k6";

export let options = {
  stages: [
    { duration: "30s", target: 100 },
    { duration: "1m", target: 1000 },
    { duration: "30s", target: 0 },
  ],
};

export default function () {
  let response = http.post("http://localhost:8080/api/orders", orderPayload);
  check(response, {
    "no duplicate error": (r) => !r.body.includes("duplicate key"),
    "status is 201": (r) => r.status === 201,
  });
}
```

## Monitoring & Validation

### Task 11: Add Monitoring Queries

Create monitoring dashboard with:

```sql
-- Check for duplicate attempts
SELECT
  schemaname,
  tablename,
  n_tup_ins as inserts,
  n_tup_upd as updates,
  n_tup_del as deletes,
  n_conflict as conflicts
FROM pg_stat_all_tables
WHERE tablename IN ('order_items', 'order_status_history')
ORDER BY n_conflict DESC;

-- Check for constraint violations
SELECT
  conname,
  conrelid::regclass AS table_name,
  pg_get_constraintdef(oid) AS definition
FROM pg_constraint
WHERE conrelid::regclass::text LIKE '%order%'
  AND contype IN ('u', 'p', 'c');
```

### Task 12: Create Alerts

```yaml
# prometheus alerts
groups:
  - name: database
    rules:
      - alert: DuplicateKeyViolations
        expr: rate(pg_stat_database_conflicts_total{datname="kisanlink_ecom"}[5m]) > 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Database experiencing duplicate key violations"
          description: "{{ $labels.instance }} has {{ $value }} conflicts/sec"
```

## Rollback Plan

### Task 13: Prepare Rollback Scripts

**File:** Create `/Users/kaushik/kisanlink-ecom/migrations/rollback_fix_duplicate_ids.sql`

```sql
-- Backup current state
CREATE TABLE order_items_backup_20251103 AS SELECT * FROM order_items;
CREATE TABLE order_status_history_backup_20251103 AS SELECT * FROM order_status_history;

-- Rollback procedure
-- 1. Stop application
-- 2. Restore from backup:
TRUNCATE order_items CASCADE;
INSERT INTO order_items SELECT * FROM order_items_backup_20251103;

-- 3. Deploy previous version
-- 4. Restart application
```

## Success Metrics

1. **Zero** duplicate key violations in 24 hours
2. **<5ms** P99 latency for ID generation
3. **100%** test coverage for ID generation code
4. **Zero** data loss during migration
5. All integration tests passing

## Timeline

- **Day 1 (Today)**: Tasks 1-3 (Hot fix)
- **Day 2-3**: Tasks 4-6 (Audit and documentation)
- **Week 2**: Tasks 7-10 (Database and testing)
- **Ongoing**: Tasks 11-13 (Monitoring and rollback)

## Dependencies

- kisanlink-db package v0.3.0 behavior analysis
- Database migration tooling
- Test environment with production-like load
- Monitoring infrastructure (Prometheus/Grafana)

## Risk Mitigation

| Risk                    | Mitigation                   |
| ----------------------- | ---------------------------- |
| Data corruption         | Full backup before changes   |
| Service outage          | Blue-green deployment        |
| Performance degradation | Load test before production  |
| Rollback failure        | Practice rollback in staging |
