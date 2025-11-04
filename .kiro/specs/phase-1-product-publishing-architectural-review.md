# Architectural Review: Phase 1 Product Publishing Workflow

**Review Date**: 2025-11-03
**Reviewer**: SDE-3 Backend Architect
**Status**: APPROVED WITH MODIFICATIONS

## Executive Summary

The proposed Phase 1 Product Publishing architecture provides a solid foundation with appropriate patterns. However, several critical modifications are required for production readiness, security, and scalability.

## Architecture Assessment

### 1. Database Schema Review

**Current Proposal Issues:**
- Missing referential integrity to `catalog_items` table
- No soft delete support (inconsistent with existing patterns)
- Missing audit fields (`created_by`, `updated_by`)
- No optimistic locking support (`version` field)
- Insufficient indexing strategy

**Required Modifications:**
✅ Add foreign key constraint to `catalog_items(id)`
✅ Include soft delete fields (`deleted_at`, `deleted_by`)
✅ Add audit trail fields
✅ Implement version field for optimistic locking
✅ Add partial indexes for performance
✅ Add `is_published` boolean flag for clear state management

**Verdict**: APPROVED with schema modifications (see ADR-001)

### 2. Model Design Review

**Issues Identified:**
- Missing interface compliance with existing patterns
- No version control mechanism
- Decimal type handling needs standardization

**Required Modifications:**
```go
// entities/models/catalog/publish_state.go
type PublishState struct {
    base.BaseModel  // Inherit standard fields

    // Core fields
    ProductID          string                      `json:"product_id" gorm:"type:varchar(255);not null;uniqueIndex:idx_product_publish,where:deleted_at IS NULL"`
    IsPublished        bool                        `json:"is_published" gorm:"not null;default:false;index:idx_publish_active"`
    FPOAccessList      []string                    `json:"fpo_access_list" gorm:"type:jsonb;default:'[]'"`
    DeliveryCosts      map[string]DeliveryCost    `json:"delivery_costs" gorm:"type:jsonb;default:'{}'"`
    PlatformFeePercent decimal.Decimal             `json:"platform_fee_percent" gorm:"type:decimal(5,2);not null;default:10.00"`

    // Publishing metadata
    PublishedAt *time.Time `json:"published_at"`
    PublishedBy string     `json:"published_by" gorm:"type:varchar(255)"`

    // Audit fields (from BaseModel)
    // CreatedBy, UpdatedBy, Version already included

    // Relationships
    Product *Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
}

type DeliveryCost struct {
    Amount    decimal.Decimal `json:"amount"`
    Currency  string          `json:"currency"`
    UpdatedAt time.Time       `json:"updated_at"`
}
```

**Verdict**: APPROVED with model restructuring

### 3. API Design Review

**Issues:**
- Inconsistent with RESTful conventions (PATCH vs POST)
- Missing validation requirements
- Incomplete error handling specification
- Missing rate limiting considerations

**Required Modifications:**

```yaml
# POST /api/v1/catalog/products/{id}/publish
Request:
  headers:
    Authorization: Bearer <token>  # Super Admin only
  body:
    fpo_ids: string[]              # Required, min 1
    delivery_costs:                # Required
      <fpo_id>:
        amount: decimal            # Required, >= 0
        currency: string           # Default: INR
    platform_fee_percent: decimal # Optional, default from config

Response (201):
  data:
    publish_id: string
    product_id: string
    status: "PUBLISHED"
    fpo_count: integer
    published_at: timestamp
    retail_prices:                # Sample calculations
      <fpo_id>:
        base_price: decimal
        delivery_cost: decimal
        commission: decimal
        retail_price: decimal

# GET /api/v1/catalog/products (FPO View - Modified)
Additional Query Params:
  include_pricing: boolean        # Include FPO-specific pricing
  published_only: boolean         # Filter published products

Response Enhancement:
  - Add pricing_details field when include_pricing=true
  - Add publish_status field
```

**Verdict**: APPROVED with API standardization

### 4. Security & Access Control

**Critical Gaps:**
- No FPO organization validation
- Missing rate limiting
- No audit logging specification
- Insufficient input validation

**Required Security Measures:**

```go
// Security middleware stack
func SecurePublishEndpoint() gin.HandlerFunc {
    return gin.HandlersChain{
        RateLimitMiddleware(10, time.Minute),     // 10 requests/min
        AuthNMiddleware(aaaClient),                // JWT validation
        RequireRole("SUPER_ADMIN"),                // Role check
        AuditLogMiddleware("PRODUCT_PUBLISH"),     // Audit trail
        ValidatePublishRequest(),                  // Input validation
    }
}

// FPO validation service
type OrganizationValidator interface {
    ValidateFPOOrganizations(ctx context.Context, orgIDs []string) ([]string, error)
    GetOrganizationType(ctx context.Context, orgID string) (string, error)
}
```

**Verdict**: APPROVED with security enhancements

### 5. Scalability Analysis

**Performance Considerations:**

| Scenario | Records | Query Pattern | Expected Performance |
|----------|---------|---------------|---------------------|
| FPO Product List | 100K products, 1K FPOs | JSONB @> operator with GIN index | < 50ms with index |
| Retail Price Calc | 1 product, 1 FPO | Single row lookup + calculation | < 10ms |
| Bulk Publish | 1K products | Batch insert/update | < 1s with batching |
| Access Check | 1 product, 1 FPO | Indexed JSONB contains | < 5ms |

**Required Optimizations:**

1. **Database Level:**
   - Implement partial indexes for active records
   - Use prepared statements
   - Connection pooling (min: 10, max: 100)

2. **Application Level:**
   - Redis caching with 5-minute TTL
   - Batch operations for bulk publishing
   - Async event publishing

3. **Infrastructure Level:**
   - Read replicas for FPO queries
   - CDN for static pricing data
   - Queue for event processing

**Verdict**: APPROVED with caching strategy

### 6. Event Architecture

**Required Events:**

```go
// Event definitions
const (
    EventProductPublished   = "catalog.product.published"
    EventProductUnpublished = "catalog.product.unpublished"
    EventFPOAccessGranted   = "catalog.fpo.access_granted"
    EventFPOAccessRevoked   = "catalog.fpo.access_revoked"
    EventPriceUpdated       = "catalog.price.updated"
)

// Publishing pattern
func (s *PublishService) publishEvent(ctx context.Context, event interface{}) error {
    return s.eventPublisher.Publish(ctx, event, WithRetry(3), WithTimeout(5*time.Second))
}
```

**Verdict**: APPROVED with event system integration

## Risk Assessment

| Risk | Severity | Likelihood | Mitigation |
|------|----------|------------|------------|
| JSONB query performance | HIGH | MEDIUM | GIN indexes, monitoring, caching |
| Data inconsistency | HIGH | LOW | Transactions, validation, event sourcing |
| FPO ID validation failure | MEDIUM | MEDIUM | Async validation job, graceful degradation |
| Cache staleness | LOW | HIGH | Event-driven invalidation, short TTL |
| Platform fee calculation errors | HIGH | LOW | Decimal precision, extensive testing |

## Implementation Roadmap

### Phase 1A: Core Infrastructure (Week 1)
1. Create database migrations
2. Implement base models and repositories
3. Set up event publishing infrastructure
4. Create validation services

### Phase 1B: Business Logic (Week 2)
1. Implement PublishService with transactions
2. Add FPO validation integration
3. Implement pricing calculations
4. Create caching layer

### Phase 1C: API Layer (Week 3)
1. Create handlers with proper auth
2. Add request/response validation
3. Implement Swagger documentation
4. Add rate limiting

### Phase 1D: Testing & Monitoring (Week 4)
1. Unit tests (>80% coverage)
2. Integration tests
3. Load testing
4. Monitoring setup

## Final Recommendations

### Critical Requirements (P0)
1. ✅ Implement JSONB with GIN indexes
2. ✅ Add comprehensive validation layer
3. ✅ Implement caching strategy
4. ✅ Add event publishing
5. ✅ Secure with proper RBAC

### Recommended Enhancements (P1)
1. Add GraphQL support for complex queries
2. Implement price history tracking
3. Add bulk operations API
4. Create admin dashboard

### Future Considerations (P2)
1. Machine learning for pricing optimization
2. Automated FPO matching
3. Dynamic commission rates
4. Multi-currency support

## Approval Decision

**Status**: APPROVED WITH MODIFICATIONS

**Conditions for Implementation:**
1. Adopt modified schema from ADR-001
2. Implement security measures as specified
3. Add comprehensive validation layer
4. Include caching from day one
5. Follow existing codebase patterns

**Go-Ahead**: The backend engineer can proceed with implementation following:
- ADR-001 for architectural decisions
- This review document for specific requirements
- Existing codebase patterns for consistency

## Architecture Compliance Checklist

- [x] Follows layered architecture pattern
- [x] Uses existing base models
- [x] Implements soft deletes
- [x] Includes audit fields
- [x] Supports multi-tenancy via OrganizationID
- [x] Compatible with AAA service
- [x] Event-driven architecture support
- [x] Follows naming conventions
- [x] Includes Swagger documentation
- [x] Supports pagination

## Appendix: Code Templates

### Repository Implementation
```go
package catalog

type PublishRepository struct {
    *common.BaseRepository
    dbManager db.DBManager
}

func (r *PublishRepository) PublishProduct(ctx context.Context, state *PublishState) error {
    return r.WithTransaction(ctx, func(tx *gorm.DB) error {
        // Implementation
    })
}
```

### Service Implementation
```go
package catalog

type PublishService struct {
    repo           PublishRepository
    validator      OrganizationValidator
    eventPublisher EventPublisher
    cache          cache.Service
}
```

### Handler Implementation
```go
package catalog

type PublishHandler struct {
    service PublishService
    logger  *logrus.Logger
}
```

---

**Reviewed and Approved by**: SDE-3 Backend Architect
**Next Steps**: Backend engineer to begin implementation with Phase 1A
