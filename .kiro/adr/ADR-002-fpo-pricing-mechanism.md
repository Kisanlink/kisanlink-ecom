# ADR-002: FPO-Specific Pricing and Price Locking Mechanism

**Status**: Proposed
**Date**: 2025-11-03
**Author**: SDE-3 Backend Architect

## Context

The KisanLink platform needs to calculate and display FPO-specific retail prices that include:
- Base product price (from catalog)
- FPO-specific delivery costs
- Platform commission (percentage of base price)
- Price locking for 30-minute windows to prevent price changes during checkout

## Problem Statement

1. Prices must be calculated dynamically per FPO
2. Prices need to be locked during active shopping sessions
3. System must handle high-frequency price queries efficiently
4. Price calculations must be consistent and auditable

## Decision Drivers

1. **Performance**: Sub-50ms response time for price queries
2. **Consistency**: Same price throughout shopping session
3. **Scalability**: Support 100K products × 1K FPOs
4. **Auditability**: Track all price changes
5. **Accuracy**: Precise decimal calculations

## Considered Options

### Option 1: Real-time Calculation Only
Calculate prices on every request without caching or locking.

**Pros:**
- Always accurate
- No cache invalidation complexity
- Simple implementation

**Cons:**
- High computational overhead
- Inconsistent prices during session
- Poor performance at scale

### Option 2: Pre-calculated Price Matrix
Store all FPO×Product price combinations in database.

**Pros:**
- Fast queries
- Consistent pricing
- Easy to audit

**Cons:**
- Massive storage (100M+ rows)
- Complex update operations
- Slow bulk price changes

### Option 3: Hybrid with Session-based Locking
Calculate prices dynamically with Redis-based session locking.

**Pros:**
- Balanced performance
- Consistent session pricing
- Reasonable storage
- Flexible updates

**Cons:**
- Redis dependency
- Session management complexity

## Decision

**Adopt Option 3: Hybrid with Session-based Locking**

## Detailed Design

### Price Calculation Formula

```go
// Core pricing formula
RetailPrice = BasePrice + DeliveryCost + (BasePrice × PlatformFeePercent / 100)

// With precision handling
func CalculateRetailPrice(
    basePrice decimal.Decimal,
    deliveryCost decimal.Decimal,
    platformFeePercent decimal.Decimal,
) decimal.Decimal {
    // Calculate commission with proper rounding
    commission := basePrice.
        Mul(platformFeePercent).
        Div(decimal.NewFromInt(100)).
        Round(2) // Round to 2 decimal places

    // Calculate total with precision
    retailPrice := basePrice.
        Add(deliveryCost).
        Add(commission).
        Round(2)

    return retailPrice
}
```

### Price Lock Implementation

```go
// Price lock structure in Redis
type PriceLock struct {
    SessionID       string          `json:"session_id"`
    UserID          string          `json:"user_id"`
    OrganizationID  string          `json:"organization_id"`
    ProductID       string          `json:"product_id"`
    BasePrice       decimal.Decimal `json:"base_price"`
    DeliveryCost    decimal.Decimal `json:"delivery_cost"`
    Commission      decimal.Decimal `json:"commission"`
    RetailPrice     decimal.Decimal `json:"retail_price"`
    LockedAt        time.Time       `json:"locked_at"`
    ExpiresAt       time.Time       `json:"expires_at"`
}

// Redis key pattern
// price_lock:{org_id}:{product_id}:{session_id}
const PriceLockKeyPattern = "price_lock:%s:%s:%s"

// Lock duration from publish_states or default
const DefaultPriceLockMinutes = 30
```

### Service Implementation

```go
type PricingService struct {
    publishRepo    PublishRepository
    catalogRepo    CatalogRepository
    cache          *redis.Client
    logger         *logrus.Logger
}

func (s *PricingService) GetFPOPrice(
    ctx context.Context,
    productID string,
    fpoOrgID string,
    sessionID string,
) (*FPOPricingDetail, error) {
    // 1. Check for existing price lock
    lock := s.checkPriceLock(ctx, productID, fpoOrgID, sessionID)
    if lock != nil && lock.ExpiresAt.After(time.Now()) {
        return s.priceLockToDetail(lock), nil
    }

    // 2. Get product and publish state
    product, err := s.catalogRepo.GetByID(ctx, productID)
    if err != nil {
        return nil, err
    }

    publishState, err := s.publishRepo.GetByProductID(ctx, productID)
    if err != nil {
        return nil, err
    }

    // 3. Validate FPO access
    if !s.hasFPOAccess(publishState, fpoOrgID) {
        return nil, ErrNoAccess
    }

    // 4. Calculate price
    deliveryCost := s.getDeliveryCost(publishState, fpoOrgID)
    commission := product.BasePrice.
        Mul(publishState.PlatformFeePercent).
        Div(decimal.NewFromInt(100)).
        Round(2)

    retailPrice := product.BasePrice.
        Add(deliveryCost).
        Add(commission).
        Round(2)

    // 5. Create price lock
    newLock := &PriceLock{
        SessionID:      sessionID,
        OrganizationID: fpoOrgID,
        ProductID:      productID,
        BasePrice:      product.BasePrice,
        DeliveryCost:   deliveryCost,
        Commission:     commission,
        RetailPrice:    retailPrice,
        LockedAt:       time.Now(),
        ExpiresAt:      time.Now().Add(30 * time.Minute),
    }

    // 6. Store in Redis with TTL
    if err := s.storePriceLock(ctx, newLock); err != nil {
        s.logger.Warnf("Failed to store price lock: %v", err)
        // Continue without lock - graceful degradation
    }

    // 7. Return pricing detail
    return &FPOPricingDetail{
        BasePrice:        product.BasePrice,
        DeliveryCost:     deliveryCost,
        CommissionAmount: commission,
        RetailPrice:      retailPrice,
        PriceLockedUntil: &newLock.ExpiresAt,
    }, nil
}
```

### Caching Strategy

```go
// Multi-layer caching
type CacheLayer int

const (
    L1_Memory CacheLayer = iota  // In-process cache (5 seconds)
    L2_Redis                      // Distributed cache (5 minutes)
    L3_Database                   // Persistent storage
)

// Cache configuration
type CacheConfig struct {
    L1TTL time.Duration // 5 seconds for hot data
    L2TTL time.Duration // 5 minutes for warm data
    L3TTL time.Duration // Permanent
}

// Cache key patterns
const (
    ProductPriceKey = "product:price:%s:%s"      // product:fpo
    PublishStateKey = "publish:state:%s"         // product
    FPOProductsKey  = "fpo:products:%s:page:%d"  // fpo:page
)
```

### Database Optimization

```sql
-- Materialized view for frequently accessed pricing data
CREATE MATERIALIZED VIEW mv_fpo_product_pricing AS
SELECT
    ps.product_id,
    ps.fpo_access_list,
    ps.delivery_costs,
    ps.platform_fee_percent,
    ci.base_price,
    ci.name,
    ci.sku,
    ps.updated_at
FROM publish_states ps
JOIN catalog_items ci ON ps.product_id = ci.id
WHERE ps.is_published = true
    AND ps.deleted_at IS NULL
    AND ci.deleted_at IS NULL
    AND ci.is_active = true;

-- Refresh strategy
CREATE INDEX idx_mv_pricing_updated ON mv_fpo_product_pricing(updated_at);

-- Refresh every 5 minutes
CREATE OR REPLACE FUNCTION refresh_pricing_view()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_fpo_product_pricing;
END;
$$ LANGUAGE plpgsql;
```

### Monitoring and Metrics

```go
// Key metrics to track
type PricingMetrics struct {
    CalculationDuration   histogram // Time to calculate price
    CacheHitRate          gauge     // L1/L2 cache hit percentage
    PriceLockCount        counter   // Active price locks
    PriceLockExpired      counter   // Expired locks
    CalculationErrors     counter   // Pricing calculation failures
    FPOAccessDenied       counter   // Access control rejections
}

// Prometheus metrics
var (
    pricingDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "pricing_calculation_duration_seconds",
            Help: "Time taken to calculate FPO pricing",
        },
        []string{"fpo_id", "cache_hit"},
    )

    priceLockGauge = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "active_price_locks",
            Help: "Number of active price locks",
        },
        []string{"fpo_id"},
    )
)
```

## Consequences

### Positive
- Consistent pricing during shopping sessions
- Excellent query performance with caching
- Scalable to millions of price queries
- Auditable price history
- Graceful degradation if Redis fails

### Negative
- Additional Redis infrastructure required
- Session management complexity
- Cache invalidation complexity
- Potential for stale prices in edge cases

## Migration Strategy

### Phase 1: Basic Implementation
- Implement price calculation service
- Add Redis for price locks
- Basic monitoring

### Phase 2: Optimization
- Add multi-layer caching
- Implement materialized views
- Advanced monitoring

### Phase 3: Advanced Features
- Dynamic pricing rules
- Bulk pricing updates
- Price history tracking

## Security Considerations

1. **Session Validation**: Ensure session IDs are cryptographically secure
2. **Rate Limiting**: Limit price queries per session
3. **Access Control**: Validate FPO membership before pricing
4. **Audit Trail**: Log all price locks and calculations

## Performance Benchmarks

| Operation | Target | Method |
|-----------|--------|--------|
| Single price calculation | < 10ms | In-memory calculation |
| Cached price retrieval | < 5ms | Redis GET |
| Bulk pricing (100 items) | < 500ms | Batch processing |
| Price lock creation | < 20ms | Redis SET with TTL |

## References
- Redis Documentation: https://redis.io/docs/
- Decimal Precision in E-commerce: https://shopify.engineering/decimal-precision
- Caching Best Practices: https://aws.amazon.com/caching/best-practices/
