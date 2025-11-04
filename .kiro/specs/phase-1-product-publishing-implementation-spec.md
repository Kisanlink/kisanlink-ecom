# Phase 1: Product Publishing Implementation Specification

**Version**: 1.0
**Created**: 2025-11-03
**Status**: Ready for Implementation

## Overview

Implement Product Publishing workflow enabling Super Admins to publish products to selected FPOs with platform fees and delivery costs.

## Implementation Checklist

### Week 1: Database & Models

#### Day 1-2: Database Setup

1. **Create migration file**: `migrations/20251103_create_publish_states.sql`
```sql
-- Full schema from ADR-001
CREATE TABLE publish_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    is_published BOOLEAN NOT NULL DEFAULT false,
    published_at TIMESTAMP,
    published_by VARCHAR(255),
    fpo_access_list JSONB NOT NULL DEFAULT '[]'::jsonb,
    delivery_costs JSONB NOT NULL DEFAULT '{}'::jsonb,
    platform_fee_percent DECIMAL(5,2) NOT NULL DEFAULT 10.00,
    price_lock_duration_minutes INT NOT NULL DEFAULT 30,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255) NOT NULL,
    version INT NOT NULL DEFAULT 1,
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),

    CONSTRAINT unique_product_publish UNIQUE(product_id) WHERE deleted_at IS NULL,
    CONSTRAINT valid_platform_fee CHECK (platform_fee_percent >= 0 AND platform_fee_percent <= 100),
    CONSTRAINT valid_price_lock CHECK (price_lock_duration_minutes > 0 AND price_lock_duration_minutes <= 1440)
);

-- Add foreign key after ensuring catalog_items exists
ALTER TABLE publish_states
ADD CONSTRAINT fk_publish_states_product
FOREIGN KEY (product_id)
REFERENCES catalog_items(id)
ON DELETE CASCADE;

-- Create indexes
CREATE INDEX idx_publish_states_product_id ON publish_states(product_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_publish_states_published ON publish_states(is_published, published_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_publish_states_fpo_access ON publish_states USING GIN(fpo_access_list) WHERE deleted_at IS NULL AND is_published = true;
```

2. **Run migration**:
```bash
make migrate-up
```

#### Day 3-4: Model Implementation

3. **Create model file**: `entities/models/catalog/publish_state.go`
```go
package catalog

import (
    "time"
    "github.com/Kisanlink/kisanlink-db/pkg/base"
    "github.com/shopspring/decimal"
    "gorm.io/gorm"
)

type PublishState struct {
    base.BaseModel

    ProductID          string           `json:"product_id" gorm:"type:varchar(255);not null;uniqueIndex:idx_product_publish,where:deleted_at IS NULL"`
    IsPublished        bool             `json:"is_published" gorm:"not null;default:false;index:idx_publish_active"`
    PublishedAt        *time.Time       `json:"published_at,omitempty"`
    PublishedBy        string           `json:"published_by,omitempty" gorm:"type:varchar(255)"`

    // Access control
    FPOAccessList      []string         `json:"fpo_access_list" gorm:"type:jsonb;default:'[]'"`

    // Pricing
    DeliveryCosts      map[string]*DeliveryCost `json:"delivery_costs" gorm:"type:jsonb;default:'{}'"`
    PlatformFeePercent decimal.Decimal  `json:"platform_fee_percent" gorm:"type:decimal(5,2);not null;default:10.00"`

    // Configuration
    PriceLockMinutes   int              `json:"price_lock_duration_minutes" gorm:"not null;default:30"`

    // Audit
    CreatedBy          string           `json:"created_by" gorm:"type:varchar(255);not null"`
    UpdatedBy          string           `json:"updated_by" gorm:"type:varchar(255);not null"`
    Version            int64            `json:"version" gorm:"not null;default:1"`

    // Soft delete
    DeletedAt          gorm.DeletedAt   `json:"-" gorm:"index"`
    DeletedBy          *string          `json:"-" gorm:"type:varchar(255)"`

    // Relationships
    Product            *CatalogItem     `json:"product,omitempty" gorm:"foreignKey:ProductID;references:ID"`
}

type DeliveryCost struct {
    Amount    decimal.Decimal `json:"amount"`
    Currency  string          `json:"currency"`
    UpdatedAt time.Time       `json:"updated_at"`
}

func (PublishState) TableName() string {
    return "publish_states"
}

// Implement VersionedEntity interface
func (p *PublishState) GetVersion() int64 {
    return p.Version
}

func (p *PublishState) IncrementVersion() {
    p.Version++
    p.UpdatedAt = time.Now()
}
```

4. **Create request/response DTOs**: `entities/requests/catalog/publish_request.go`
```go
package catalog

import "github.com/shopspring/decimal"

type PublishProductRequest struct {
    FPOIDs             []string                   `json:"fpo_ids" binding:"required,min=1"`
    DeliveryCosts      map[string]decimal.Decimal `json:"delivery_costs" binding:"required"`
    PlatformFeePercent *decimal.Decimal           `json:"platform_fee_percent,omitempty"`
}

type UpdateDeliveryCostRequest struct {
    DeliveryCosts map[string]decimal.Decimal `json:"delivery_costs" binding:"required"`
}

type RevokeAccessRequest struct {
    FPOIDs []string `json:"fpo_ids" binding:"required,min=1"`
}
```

### Week 2: Repository & Service Layer

#### Day 5-6: Repository Implementation

5. **Create repository**: `internal/repositories/catalog/publish_repository.go`
```go
package catalog

import (
    "context"
    "fmt"
    "kisanlink-ecom/entities/models/catalog"
    "kisanlink-ecom/internal/repositories/common"
    "github.com/Kisanlink/kisanlink-db/pkg/db"
)

type PublishRepository struct {
    *common.BaseRepository
    dbManager db.DBManager
}

func NewPublishRepository(dbManager db.DBManager) *PublishRepository {
    return &PublishRepository{
        BaseRepository: common.NewBaseRepository(dbManager),
        dbManager:      dbManager,
    }
}

func (r *PublishRepository) GetByProductID(ctx context.Context, productID string) (*catalog.PublishState, error) {
    var state catalog.PublishState
    err := r.dbManager.GetByField(ctx, "product_id", productID, &state)
    if err != nil {
        return nil, err
    }
    return &state, nil
}

func (r *PublishRepository) PublishProduct(ctx context.Context, state *catalog.PublishState) error {
    // Use transaction for consistency
    return r.WithTransaction(ctx, func(tx interface{}) error {
        // Check if exists
        var existing catalog.PublishState
        err := r.dbManager.GetByField(ctx, "product_id", state.ProductID, &existing)

        if err == nil {
            // Update existing
            existing.FPOAccessList = state.FPOAccessList
            existing.DeliveryCosts = state.DeliveryCosts
            existing.PlatformFeePercent = state.PlatformFeePercent
            existing.IsPublished = true
            existing.PublishedAt = state.PublishedAt
            existing.PublishedBy = state.PublishedBy
            existing.IncrementVersion()
            return r.dbManager.Update(ctx, &existing)
        }

        // Create new
        return r.dbManager.Create(ctx, state)
    })
}

func (r *PublishRepository) GetFPOProducts(ctx context.Context, fpoID string, limit, offset int) ([]*catalog.CatalogItem, error) {
    // Query products accessible to FPO using JSONB contains
    query := `
        SELECT c.* FROM catalog_items c
        JOIN publish_states ps ON c.id = ps.product_id
        WHERE ps.deleted_at IS NULL
        AND ps.is_published = true
        AND ps.fpo_access_list @> $1
        AND c.deleted_at IS NULL
        AND c.is_active = true
        ORDER BY c.created_at DESC
        LIMIT $2 OFFSET $3
    `

    var products []*catalog.CatalogItem
    err := r.dbManager.Raw(ctx, query, fmt.Sprintf(`["%s"]`, fpoID), limit, offset).Scan(&products).Error
    return products, err
}
```

#### Day 7-8: Service Implementation

6. **Create service**: `internal/services/catalog/publish_service.go`
```go
package catalog

import (
    "context"
    "fmt"
    "time"

    "kisanlink-ecom/entities/models/catalog"
    catalogReq "kisanlink-ecom/entities/requests/catalog"
    "kisanlink-ecom/internal/repositories/catalog"
    "kisanlink-ecom/internal/services/events"

    "github.com/shopspring/decimal"
    "github.com/sirupsen/logrus"
)

type PublishService struct {
    publishRepo    *catalog.PublishRepository
    catalogRepo    *catalog.CatalogRepository
    eventPublisher events.Publisher
    logger         *logrus.Logger
    cache          CacheService
}

func NewPublishService(
    publishRepo *catalog.PublishRepository,
    catalogRepo *catalog.CatalogRepository,
    eventPublisher events.Publisher,
    logger *logrus.Logger,
    cache CacheService,
) *PublishService {
    return &PublishService{
        publishRepo:    publishRepo,
        catalogRepo:    catalogRepo,
        eventPublisher: eventPublisher,
        logger:         logger,
        cache:          cache,
    }
}

func (s *PublishService) PublishProduct(ctx context.Context, productID string, req *catalogReq.PublishProductRequest, userID string) error {
    // 1. Validate product exists
    product, err := s.catalogRepo.GetByID(ctx, productID, &catalog.CatalogItem{})
    if err != nil {
        return fmt.Errorf("product not found: %w", err)
    }

    // 2. Validate FPO IDs (would integrate with organization service)
    if err := s.validateFPOIDs(ctx, req.FPOIDs); err != nil {
        return fmt.Errorf("invalid FPO IDs: %w", err)
    }

    // 3. Prepare publish state
    now := time.Now()
    state := &catalog.PublishState{
        ProductID:          productID,
        IsPublished:        true,
        PublishedAt:        &now,
        PublishedBy:        userID,
        FPOAccessList:      req.FPOIDs,
        DeliveryCosts:      s.prepareDeliveryCosts(req.DeliveryCosts),
        PlatformFeePercent: s.getPlatformFee(req.PlatformFeePercent),
        CreatedBy:          userID,
        UpdatedBy:          userID,
    }

    // 4. Save to database
    if err := s.publishRepo.PublishProduct(ctx, state); err != nil {
        return fmt.Errorf("failed to publish product: %w", err)
    }

    // 5. Invalidate cache
    s.invalidateProductCache(ctx, productID, req.FPOIDs)

    // 6. Publish event
    event := events.ProductPublishedEvent{
        ProductID:   productID,
        FPOCount:    len(req.FPOIDs),
        PublishedBy: userID,
        PublishedAt: now,
    }

    if err := s.eventPublisher.Publish(ctx, "catalog.product.published", event); err != nil {
        s.logger.Warnf("Failed to publish event: %v", err)
        // Don't fail the operation
    }

    return nil
}

func (s *PublishService) CalculateRetailPrice(
    basePrice decimal.Decimal,
    deliveryCost decimal.Decimal,
    platformFeePercent decimal.Decimal,
) decimal.Decimal {
    // Calculate commission
    commission := basePrice.
        Mul(platformFeePercent).
        Div(decimal.NewFromInt(100)).
        Round(2)

    // Calculate total
    return basePrice.
        Add(deliveryCost).
        Add(commission).
        Round(2)
}

func (s *PublishService) GetFPOProductPricing(ctx context.Context, productID, fpoID string) (*catalog.FPOPricingDetail, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("pricing:%s:%s", productID, fpoID)
    if cached, found := s.cache.Get(cacheKey); found {
        return cached.(*catalog.FPOPricingDetail), nil
    }

    // Get product and publish state
    product, err := s.catalogRepo.GetByID(ctx, productID, &catalog.CatalogItem{})
    if err != nil {
        return nil, err
    }

    publishState, err := s.publishRepo.GetByProductID(ctx, productID)
    if err != nil {
        return nil, err
    }

    // Check access
    if !s.hasFPOAccess(publishState, fpoID) {
        return nil, fmt.Errorf("FPO does not have access to this product")
    }

    // Get delivery cost
    deliveryCost := s.getDeliveryCostForFPO(publishState, fpoID)

    // Calculate pricing
    commission := product.BasePrice.
        Mul(publishState.PlatformFeePercent).
        Div(decimal.NewFromInt(100)).
        Round(2)

    retailPrice := s.CalculateRetailPrice(
        product.BasePrice,
        deliveryCost,
        publishState.PlatformFeePercent,
    )

    // Prepare response
    lockUntil := time.Now().Add(time.Duration(publishState.PriceLockMinutes) * time.Minute)
    pricing := &catalog.FPOPricingDetail{
        BasePrice:        product.BasePrice,
        DeliveryCost:     deliveryCost,
        CommissionAmount: commission,
        RetailPrice:      retailPrice,
        PriceLockedUntil: &lockUntil,
    }

    // Cache for 5 minutes
    s.cache.Set(cacheKey, pricing, 5*time.Minute)

    return pricing, nil
}

// Helper methods
func (s *PublishService) validateFPOIDs(ctx context.Context, fpoIDs []string) error {
    // TODO: Integrate with organization service to validate FPO IDs
    // For now, basic validation
    if len(fpoIDs) == 0 {
        return fmt.Errorf("at least one FPO ID required")
    }
    return nil
}

func (s *PublishService) hasFPOAccess(state *catalog.PublishState, fpoID string) bool {
    for _, id := range state.FPOAccessList {
        if id == fpoID {
            return true
        }
    }
    return false
}
```

### Week 3: API Layer

#### Day 9-10: Handler Implementation

7. **Create handler**: `internal/handlers/catalog/publish_handler.go`
```go
package catalog

import (
    "net/http"

    "kisanlink-ecom/entities/requests/catalog"
    "kisanlink-ecom/internal/common"
    "kisanlink-ecom/internal/middleware"
    publishService "kisanlink-ecom/internal/services/catalog"

    "github.com/gin-gonic/gin"
)

type PublishHandler struct {
    service *publishService.PublishService
}

func NewPublishHandler(service *publishService.PublishService) *PublishHandler {
    return &PublishHandler{
        service: service,
    }
}

// PublishProduct godoc
// @Summary Publish product to FPOs
// @Description Publish a product to selected FPOs with delivery costs and platform fee
// @Tags products
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Product ID"
// @Param request body catalog.PublishProductRequest true "Publish request"
// @Success 201 {object} common.Response
// @Failure 400 {object} common.Response
// @Failure 401 {object} common.Response
// @Failure 403 {object} common.Response
// @Failure 404 {object} common.Response
// @Router /api/v1/catalog/products/{id}/publish [post]
func (h *PublishHandler) PublishProduct(c *gin.Context) {
    productID := c.Param("id")
    if productID == "" {
        common.BadRequest(c, "INVALID_ID", "Product ID is required", nil)
        return
    }

    var req catalog.PublishProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
            "error": err.Error(),
        })
        return
    }

    // Get user ID from context (set by auth middleware)
    userID := middleware.GetSubjectID(c)

    // Check if user has Super Admin role
    if !middleware.HasRole(c, "SUPER_ADMIN") {
        common.Forbidden(c, "INSUFFICIENT_PERMISSIONS", "Only Super Admins can publish products", nil)
        return
    }

    // Publish product
    if err := h.service.PublishProduct(c.Request.Context(), productID, &req, userID); err != nil {
        common.InternalServerError(c, "PUBLISH_FAILED", "Failed to publish product", map[string]interface{}{
            "error": err.Error(),
        })
        return
    }

    common.Created(c, map[string]interface{}{
        "message":    "Product published successfully",
        "product_id": productID,
        "fpo_count":  len(req.FPOIDs),
    })
}

// GetPublishStatus godoc
// @Summary Get product publish status
// @Description Get the publish status and configuration for a product
// @Tags products
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Product ID"
// @Success 200 {object} common.Response
// @Failure 401 {object} common.Response
// @Failure 404 {object} common.Response
// @Router /api/v1/catalog/products/{id}/publish-status [get]
func (h *PublishHandler) GetPublishStatus(c *gin.Context) {
    productID := c.Param("id")

    status, err := h.service.GetPublishStatus(c.Request.Context(), productID)
    if err != nil {
        common.NotFound(c, "NOT_FOUND", "Product publish status not found", nil)
        return
    }

    common.OK(c, status)
}

// GetFPOPricing godoc
// @Summary Get FPO-specific pricing
// @Description Get pricing details for a product specific to the requesting FPO
// @Tags products
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Product ID"
// @Success 200 {object} common.Response
// @Failure 401 {object} common.Response
// @Failure 403 {object} common.Response
// @Failure 404 {object} common.Response
// @Router /api/v1/catalog/products/{id}/fpo-pricing [get]
func (h *PublishHandler) GetFPOPricing(c *gin.Context) {
    productID := c.Param("id")

    // Get FPO ID from context (extracted from JWT)
    fpoID := middleware.GetOrganizationID(c)
    if fpoID == "" {
        common.BadRequest(c, "MISSING_ORG", "Organization ID not found in token", nil)
        return
    }

    pricing, err := h.service.GetFPOProductPricing(c.Request.Context(), productID, fpoID)
    if err != nil {
        common.Forbidden(c, "NO_ACCESS", "FPO does not have access to this product", nil)
        return
    }

    common.OK(c, pricing)
}
```

#### Day 11: Route Registration

8. **Update routes**: `internal/routes/catalog_routes.go`
```go
func SetupCatalogRoutes(router *gin.RouterGroup, handlers *CatalogHandlers, authMiddleware gin.HandlerFunc) {
    catalog := router.Group("/catalog")
    catalog.Use(authMiddleware)
    {
        products := catalog.Group("/products")
        {
            // Existing routes...

            // Publishing routes (Super Admin only)
            products.POST("/:id/publish", handlers.Publish.PublishProduct)
            products.GET("/:id/publish-status", handlers.Publish.GetPublishStatus)
            products.DELETE("/:id/publish", handlers.Publish.UnpublishProduct)
            products.PATCH("/:id/delivery-costs", handlers.Publish.UpdateDeliveryCosts)

            // FPO-specific routes
            products.GET("/:id/fpo-pricing", handlers.Publish.GetFPOPricing)
        }
    }
}
```

### Week 4: Testing & Deployment

#### Day 12-13: Testing

9. **Unit tests**: `internal/services/catalog/publish_service_test.go`
10. **Integration tests**: `internal/handlers/catalog/publish_handler_test.go`
11. **Repository tests**: `internal/repositories/catalog/publish_repository_test.go`

#### Day 14: Documentation & Deployment

12. **Update Swagger docs**:
```bash
make swagger
```

13. **Deploy to staging**:
```bash
make build
docker build -t kisanlink-ecom:latest .
docker-compose up -d
```

## Testing Checklist

### Unit Tests
- [ ] PublishService.PublishProduct
- [ ] PublishService.CalculateRetailPrice
- [ ] PublishService.GetFPOProductPricing
- [ ] PublishRepository.PublishProduct
- [ ] PublishRepository.GetFPOProducts

### Integration Tests
- [ ] POST /api/v1/catalog/products/{id}/publish
- [ ] GET /api/v1/catalog/products/{id}/publish-status
- [ ] GET /api/v1/catalog/products/{id}/fpo-pricing
- [ ] PATCH /api/v1/catalog/products/{id}/delivery-costs

### Load Tests
- [ ] 1000 concurrent FPO pricing queries
- [ ] 100 concurrent publish operations
- [ ] Cache performance under load

## Monitoring Setup

1. **Prometheus metrics**:
```yaml
- publish_operations_total
- fpo_pricing_queries_total
- pricing_calculation_duration_seconds
- cache_hit_ratio
```

2. **Logging**:
```go
logger.WithFields(logrus.Fields{
    "product_id": productID,
    "fpo_count": len(fpoIDs),
    "user_id": userID,
    "action": "product_published",
}).Info("Product published successfully")
```

3. **Alerts**:
```yaml
- name: HighPricingQueryLatency
  expr: pricing_calculation_duration_seconds > 0.05
  for: 5m
  annotations:
    summary: "High pricing query latency"
```

## Go-Live Checklist

- [ ] Database migrations applied
- [ ] All tests passing (>80% coverage)
- [ ] Swagger documentation updated
- [ ] Security review completed
- [ ] Performance benchmarks met
- [ ] Monitoring configured
- [ ] Rollback plan documented
- [ ] Admin training completed

## Support & Troubleshooting

### Common Issues

1. **JSONB Query Performance**:
   - Ensure GIN indexes are created
   - Monitor query plans with EXPLAIN ANALYZE
   - Consider materialized views for heavy queries

2. **Cache Inconsistency**:
   - Check Redis connectivity
   - Verify cache invalidation events
   - Monitor cache hit ratio

3. **Price Calculation Errors**:
   - Verify decimal precision settings
   - Check for null/zero values
   - Review calculation logs

### Rollback Procedure

1. Revert code deployment
2. Restore database from backup if schema changed
3. Clear Redis cache
4. Verify system functionality

## Next Phase Preview

Phase 2 will include:
- Bulk publishing operations
- Price history tracking
- Advanced FPO matching algorithms
- Multi-currency support
- Dynamic commission rates

---

**Implementation Owner**: Backend Engineer
**Review Required By**: SDE-3 Backend Architect
**Deployment Approval**: DevOps Team
