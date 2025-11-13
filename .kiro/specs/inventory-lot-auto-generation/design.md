# Architecture Decision Record: Lot/Batch Auto-Generation & Database-Initialized ID Sequences

## Status

**Status**: Proposed
**Date**: 2025-10-24
**Author**: Backend Architecture Team

## Context

The current system faces two critical issues:

1. **Lot Number Collisions**: Manual lot number entry causes duplicate key violations on the unique constraint `idx_inventory_lots_org_lot_number`
2. **Random ID Generation**: Entity IDs use random generation via `NewBaseModel(prefix, size)` which lacks sequential ordering and traceability

These issues impact system reliability, user experience, and data auditability.

## Decision

Implement a dual-layer sequence management system:

1. **Application-level sequence management** for lot/batch numbers with format-based generation
2. **Database-backed sequence management** for entity IDs with initialization from existing data

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     API Layer (Handlers)                     │
├─────────────────────────────────────────────────────────────┤
│                    Service Layer                             │
│  ┌──────────────┐  ┌─────────────┐  ┌──────────────┐      │
│  │  Inventory   │  │  Sequence   │  │    Entity    │      │
│  │   Service    │──▶│   Service   │◀──│   Services   │      │
│  └──────────────┘  └─────────────┘  └──────────────┘      │
├─────────────────────────────────────────────────────────────┤
│                   Repository Layer                           │
│  ┌──────────────┐  ┌─────────────┐  ┌──────────────┐      │
│  │  Inventory   │  │  Sequence   │  │    Base      │      │
│  │  Repository  │  │ Repository  │  │  Repository  │      │
│  └──────────────┘  └─────────────┘  └──────────────┘      │
├─────────────────────────────────────────────────────────────┤
│                  Database Abstraction (kisanlink-db)         │
├─────────────────────────────────────────────────────────────┤
│                       PostgreSQL                             │
│  ┌──────────────┐  ┌─────────────┐  ┌──────────────┐      │
│  │  inventory_  │  │  sequence_  │  │   Entity     │      │
│  │     lots     │  │  counters   │  │   Tables     │      │
│  └──────────────┘  └─────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

## Detailed Design

### 1. Sequence Counter Table Schema

```sql
CREATE TABLE IF NOT EXISTS sequence_counters (
    id VARCHAR(255) PRIMARY KEY,
    sequence_key VARCHAR(255) NOT NULL,
    organization_id VARCHAR(255),
    current_value BIGINT NOT NULL DEFAULT 0,
    prefix VARCHAR(50),
    format_pattern VARCHAR(100),
    reset_frequency VARCHAR(20), -- 'daily', 'monthly', 'yearly', 'never'
    last_reset_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Unique constraint for sequence key per organization
    CONSTRAINT uq_sequence_org UNIQUE(sequence_key, organization_id)
);

-- Index for fast lookups
CREATE INDEX idx_sequence_counters_key_org ON sequence_counters(sequence_key, organization_id);
CREATE INDEX idx_sequence_counters_reset ON sequence_counters(reset_frequency, last_reset_at);
```

### 2. Sequence Service Interface

```go
// internal/services/sequence/sequence_service.go

type SequenceService interface {
    // Core sequence operations
    GetNextValue(ctx context.Context, sequenceKey string, orgID *string) (int64, error)
    GetNextFormattedValue(ctx context.Context, sequenceKey string, orgID *string, format SequenceFormat) (string, error)

    // Initialization
    InitializeSequence(ctx context.Context, sequenceKey string, orgID *string, startValue int64) error
    InitializeFromMax(ctx context.Context, sequenceKey, tableName, columnName string, orgID *string) error

    // Lot/Batch specific
    GenerateLotNumber(ctx context.Context, orgID string, orgPrefix *string) (string, error)
    GenerateBatchNumber(ctx context.Context, orgID string) (string, error)
}

type SequenceFormat struct {
    Prefix      string
    DateFormat  string // "YYYYMMDD", "YYYYMM", etc.
    DigitCount  int    // Number of digits in sequence part
    Separator   string // Character between parts
}
```

### 3. Implementation Details

#### 3.1 Lot Number Generation

```go
func (s *sequenceService) GenerateLotNumber(ctx context.Context, orgID string, orgPrefix *string) (string, error) {
    // Begin transaction for atomicity
    tx := s.db.Begin()
    defer tx.Rollback()

    // Determine prefix
    prefix := s.getOrganizationPrefix(orgID, orgPrefix)

    // Generate date component
    dateStr := time.Now().Format("20060102")

    // Build sequence key: "lot_{orgID}_{date}"
    sequenceKey := fmt.Sprintf("lot_%s_%s", orgID, dateStr)

    // Get next sequence atomically with row lock
    var counter SequenceCounter
    err := tx.Set("gorm:query_option", "FOR UPDATE").
        Where("sequence_key = ? AND organization_id = ?", sequenceKey, orgID).
        First(&counter).Error

    if err == gorm.ErrRecordNotFound {
        // Create new sequence
        counter = SequenceCounter{
            SequenceKey:     sequenceKey,
            OrganizationID:  &orgID,
            CurrentValue:    1,
            ResetFrequency:  "daily",
            LastResetAt:     timePtr(time.Now()),
        }
        if err := tx.Create(&counter).Error; err != nil {
            // Handle concurrent creation
            if isDuplicateKeyError(err) {
                // Retry with SELECT FOR UPDATE
                return s.GenerateLotNumber(ctx, orgID, orgPrefix)
            }
            return "", err
        }
    } else if err != nil {
        return "", err
    } else {
        // Increment existing counter
        counter.CurrentValue++
        if err := tx.Save(&counter).Error; err != nil {
            return "", err
        }
    }

    // Commit transaction
    if err := tx.Commit().Error; err != nil {
        return "", err
    }

    // Format: LOT-{PREFIX}-{YYYYMMDD}-{SEQUENCE}
    return fmt.Sprintf("LOT-%s-%s-%03d", prefix, dateStr, counter.CurrentValue), nil
}
```

#### 3.2 Entity ID Sequence Management

```go
// Enhanced BaseModel initialization
func NewBaseModelWithSequence(prefix string, sequenceService SequenceService) (*BaseModel, error) {
    ctx := context.Background()

    // Get next sequence value
    nextVal, err := sequenceService.GetNextValue(ctx, fmt.Sprintf("entity_%s", strings.ToLower(prefix)), nil)
    if err != nil {
        // Fallback to random generation if sequence service fails
        return NewBaseModel(prefix, "large"), nil
    }

    // Format ID with zero-padding
    id := fmt.Sprintf("%s%06d", prefix, nextVal)

    return &BaseModel{
        ID:        id,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }, nil
}
```

#### 3.3 Database Initialization

```go
func (s *sequenceService) InitializeFromMax(ctx context.Context, sequenceKey, tableName, columnName string, orgID *string) error {
    // Query max ID from table
    var maxID string
    query := fmt.Sprintf("SELECT MAX(%s) FROM %s", columnName, tableName)
    if orgID != nil {
        query += fmt.Sprintf(" WHERE organization_id = '%s'", *orgID)
    }

    err := s.db.Raw(query).Scan(&maxID).Error
    if err != nil || maxID == "" {
        // No existing records, start from 1
        return s.InitializeSequence(ctx, sequenceKey, orgID, 1)
    }

    // Extract numeric part from ID
    numericPart := extractNumericSuffix(maxID)
    startValue := numericPart + 1

    return s.InitializeSequence(ctx, sequenceKey, orgID, startValue)
}
```

### 4. Concurrency Control Strategy

#### 4.1 PostgreSQL Advisory Locks

```go
func (s *sequenceService) GetNextValueWithAdvisoryLock(ctx context.Context, sequenceKey string, orgID *string) (int64, error) {
    // Generate unique lock ID from sequence key
    lockID := hashToInt64(sequenceKey + orgID)

    // Acquire advisory lock
    var lockAcquired bool
    err := s.db.Raw("SELECT pg_try_advisory_lock(?)", lockID).Scan(&lockAcquired).Error
    if err != nil {
        return 0, err
    }

    if !lockAcquired {
        // Wait for lock with timeout
        ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
        defer cancel()

        err = s.db.Raw("SELECT pg_advisory_lock(?)", lockID).Error
        if err != nil {
            return 0, fmt.Errorf("failed to acquire lock: %w", err)
        }
    }

    // Ensure lock is released
    defer s.db.Exec("SELECT pg_advisory_unlock(?)", lockID)

    // Perform sequence operation
    return s.getNextValueInternal(ctx, sequenceKey, orgID)
}
```

#### 4.2 Optimistic Locking with Retry

```go
func (s *sequenceService) GetNextValueOptimistic(ctx context.Context, sequenceKey string, orgID *string) (int64, error) {
    maxRetries := 3
    for i := 0; i < maxRetries; i++ {
        // Read current value
        var counter SequenceCounter
        err := s.db.Where("sequence_key = ? AND organization_id = ?", sequenceKey, orgID).First(&counter).Error
        if err != nil {
            if err == gorm.ErrRecordNotFound {
                // Create new sequence
                return s.createNewSequence(ctx, sequenceKey, orgID)
            }
            return 0, err
        }

        // Try to update with version check
        oldValue := counter.CurrentValue
        newValue := oldValue + 1

        result := s.db.Model(&SequenceCounter{}).
            Where("sequence_key = ? AND organization_id = ? AND current_value = ?",
                  sequenceKey, orgID, oldValue).
            Update("current_value", newValue)

        if result.Error != nil {
            return 0, result.Error
        }

        if result.RowsAffected == 1 {
            return newValue, nil
        }

        // Concurrent update detected, retry
        time.Sleep(time.Duration(i*10) * time.Millisecond)
    }

    return 0, fmt.Errorf("failed to update sequence after %d retries", maxRetries)
}
```

### 5. Migration Plan

#### Phase 1: Database Schema (Week 1)

1. Create sequence_counters table
2. Add migration for organization prefix column (optional)
3. Create indexes for performance

#### Phase 2: Service Implementation (Week 2)

1. Implement SequenceService interface
2. Add sequence repository
3. Integrate with existing inventory service
4. Add fallback mechanisms

#### Phase 3: ID Migration (Week 3)

1. Initialize sequences from existing max IDs
2. Update BaseModel to use sequence service
3. Gradual rollout with feature flags

#### Phase 4: Testing & Rollout (Week 4)

1. Load testing for concurrent access
2. Integration testing
3. Staged rollout by organization

### 6. API Contract Changes

#### Create Inventory Lot - Before

```json
POST /api/v1/inventory/lots
{
    "catalog_item_id": "CAT123",
    "lot_number": "LOT-2024-001",  // Required
    "batch_number": "BATCH-001",    // Optional
    "initial_quantity": 100.5
}
```

#### Create Inventory Lot - After

```json
POST /api/v1/inventory/lots
{
    "catalog_item_id": "CAT123",
    "lot_number": "LOT-2024-001",  // Optional (auto-generated if not provided)
    "batch_number": "BATCH-001",    // Optional (auto-generated if not provided)
    "initial_quantity": 100.5
}

Response:
{
    "success": true,
    "data": {
        "id": "LOT000001",
        "lot_number": "LOT-ABC-20251024-001",  // Auto-generated
        "batch_number": "BATCH-20251024-001",   // Auto-generated
        ...
    }
}
```

### 7. Performance Considerations

#### 7.1 Caching Strategy

- Cache sequence values in Redis for read-heavy operations
- Pre-allocate sequence blocks (e.g., reserve 10 numbers at once)
- Use write-through cache for sequence updates

#### 7.2 Database Optimizations

- Partition sequence_counters table by organization for large-scale
- Use partial indexes for active sequences
- Regular vacuum and analyze for PostgreSQL

#### 7.3 Monitoring & Metrics

- Track sequence generation latency
- Monitor lock contention rates
- Alert on sequence gaps beyond threshold
- Dashboard for sequence usage patterns

### 8. Security Considerations

#### 8.1 Access Control

- Sequence generation requires authenticated user context
- Organization-level isolation enforced at service layer
- Audit trail for all sequence operations

#### 8.2 Input Validation

- Strict validation for manually provided lot/batch numbers
- Format validation using regex patterns
- SQL injection prevention in dynamic queries

#### 8.3 Rate Limiting

- Implement rate limiting for sequence generation endpoints
- Prevent sequence exhaustion attacks
- Monitor for unusual generation patterns

### 9. Testing Strategy

#### 9.1 Unit Tests

```go
func TestGenerateLotNumber_Concurrent(t *testing.T) {
    service := NewSequenceService(db)
    orgID := "ORG123"

    var wg sync.WaitGroup
    results := make(chan string, 100)

    // Spawn 100 concurrent goroutines
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            lotNumber, err := service.GenerateLotNumber(context.Background(), orgID, nil)
            assert.NoError(t, err)
            results <- lotNumber
        }()
    }

    wg.Wait()
    close(results)

    // Verify uniqueness
    seen := make(map[string]bool)
    for lotNumber := range results {
        assert.False(t, seen[lotNumber], "Duplicate lot number: %s", lotNumber)
        seen[lotNumber] = true
    }

    assert.Equal(t, 100, len(seen))
}
```

#### 9.2 Integration Tests

- Test sequence initialization from existing data
- Verify transaction rollback behavior
- Test failover scenarios
- Validate format compliance

#### 9.3 Load Tests

- Concurrent lot creation (1000 req/s)
- Sequence generation under load
- Database connection pool sizing
- Lock contention measurement

### 10. Rollback Strategy

#### 10.1 Feature Flags

```go
type FeatureFlags struct {
    AutoGenerateLotNumbers   bool
    AutoGenerateBatchNumbers bool
    UseSequenceForEntityIDs  bool
}
```

#### 10.2 Rollback Procedure

1. Disable feature flags
2. Revert to manual lot number entry
3. Keep sequence_counters table for audit
4. Restore random ID generation for entities

### 11. Monitoring & Observability

#### 11.1 Metrics

- `sequence.generation.duration` - Time to generate sequence
- `sequence.generation.errors` - Error rate
- `sequence.lock.wait_time` - Lock acquisition time
- `sequence.gaps.detected` - Number of sequence gaps

#### 11.2 Logging

```go
log.WithFields(logrus.Fields{
    "sequence_key": sequenceKey,
    "organization_id": orgID,
    "generated_value": value,
    "duration_ms": duration.Milliseconds(),
}).Info("Sequence generated successfully")
```

#### 11.3 Alerts

- Alert on sequence generation failures > 1%
- Alert on lock wait time > 1s
- Alert on sequence gap > 100
- Alert on sequence table size growth

## Consequences

### Positive

- **Zero Collisions**: Guaranteed unique lot numbers within organization
- **Better UX**: Users don't need to manually track lot numbers
- **Auditability**: Sequential IDs improve traceability
- **Performance**: Optimized sequence generation with caching
- **Flexibility**: Supports both auto-generation and manual entry

### Negative

- **Complexity**: Additional service layer and database table
- **Sequence Gaps**: Rollbacks may create gaps in sequences
- **Migration Risk**: Existing systems need careful migration
- **Lock Contention**: High concurrency may cause lock waits

### Risks & Mitigation

| Risk                    | Impact | Probability | Mitigation                            |
| ----------------------- | ------ | ----------- | ------------------------------------- |
| Sequence exhaustion     | High   | Low         | 64-bit integers, daily reset for lots |
| Lock contention         | Medium | Medium      | Advisory locks, optimistic locking    |
| Migration failure       | High   | Low         | Phased rollout, rollback plan         |
| Performance degradation | Medium | Low         | Caching, connection pooling           |

## References

- [PostgreSQL Advisory Locks](https://www.postgresql.org/docs/current/explicit-locking.html#ADVISORY-LOCKS)
- [GORM Transactions](https://gorm.io/docs/transactions.html)
- [Distributed Sequence Generation Patterns](https://martinfowler.com/articles/patterns-of-distributed-systems/generation-clock.html)
- [UUID vs Sequential IDs Performance](https://www.percona.com/blog/2019/11/22/uuids-are-popular-but-bad-for-performance-lets-discuss/)

## Appendix A: Organization Prefix Derivation

```go
func deriveOrgPrefix(orgID string) string {
    // Remove common prefixes
    cleaned := strings.TrimPrefix(orgID, "ORG")
    cleaned = strings.TrimPrefix(cleaned, "org_")

    // Take first 3 alphanumeric characters
    var prefix strings.Builder
    for _, char := range cleaned {
        if unicode.IsLetter(char) || unicode.IsDigit(char) {
            prefix.WriteRune(unicode.ToUpper(char))
            if prefix.Len() >= 3 {
                break
            }
        }
    }

    // Pad with organization ID hash if needed
    if prefix.Len() < 3 {
        hash := sha256.Sum256([]byte(orgID))
        for i := 0; prefix.Len() < 3; i++ {
            prefix.WriteByte('A' + (hash[i] % 26))
        }
    }

    return prefix.String()
}
```

## Appendix B: Database-Specific Implementations

### PostgreSQL

```sql
-- Use native sequences for better performance
CREATE SEQUENCE IF NOT EXISTS lot_sequence_org_123_20241024
    START WITH 1
    INCREMENT BY 1
    NO MAXVALUE
    CACHE 10;

-- Get next value
SELECT nextval('lot_sequence_org_123_20241024');
```

### DynamoDB (Future)

```go
// Use DynamoDB atomic counters
result, err := dynamoClient.UpdateItem(&dynamodb.UpdateItemInput{
    TableName: aws.String("sequence_counters"),
    Key: map[string]*dynamodb.AttributeValue{
        "sequence_key": {S: aws.String(sequenceKey)},
    },
    UpdateExpression: aws.String("ADD current_value :inc"),
    ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
        ":inc": {N: aws.String("1")},
    },
    ReturnValues: aws.String("UPDATED_NEW"),
})
```

---

End of Architecture Decision Record
