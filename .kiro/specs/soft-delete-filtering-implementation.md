# Soft Delete Filtering Implementation Specification

## Overview

This specification details the implementation of a consistent soft delete filtering pattern across the Kisanlink e-commerce platform, following ADR-003.

## Design Principles

1. **Secure by Default**: Soft-deleted items are filtered out unless explicitly requested
2. **Least Privilege**: Only authorized users can view deleted items
3. **Consistency**: Same pattern across all layers and models
4. **Performance**: Leverage database indexes, avoid N+1 queries
5. **Observability**: Track and audit access to deleted items

## Package Structure

```
kisanlink-ecom/
├── pkg/
│   ├── repository/
│   │   ├── options.go          # Query options definitions
│   │   ├── options_test.go
│   │   └── base_repository.go  # Base repository with options
│   ├── context/
│   │   ├── query_context.go    # Context helpers for options
│   │   └── query_context_test.go
│   └── security/
│       ├── permissions.go      # Permission checks for deleted items
│       └── audit.go           # Audit logging for deleted access
```

## Core Components

### 1. Query Options Structure

```go
// pkg/repository/options.go
package repository

import (
    "time"
)

// QueryOptions controls query behavior across repositories
type QueryOptions struct {
    // Soft delete control
    IncludeDeleted bool `json:"include_deleted"`

    // Future: Tenant isolation
    TenantID *string `json:"tenant_id,omitempty"`

    // Future: User-specific filtering
    UserID *string `json:"user_id,omitempty"`

    // Future: Time-based filtering
    AsOf *time.Time `json:"as_of,omitempty"`

    // Audit metadata
    RequestID string `json:"request_id"`
    Reason    string `json:"reason,omitempty"` // Why accessing deleted items
}

// DefaultQueryOptions returns safe defaults
func DefaultQueryOptions() QueryOptions {
    return QueryOptions{
        IncludeDeleted: false,
    }
}

// Validate ensures options are valid
func (o QueryOptions) Validate() error {
    if o.IncludeDeleted && o.Reason == "" {
        return fmt.Errorf("reason required when accessing deleted items")
    }
    return nil
}

// Functional options for flexible API
type QueryOption func(*QueryOptions)

func WithDeleted(reason string) QueryOption {
    return func(opts *QueryOptions) {
        opts.IncludeDeleted = true
        opts.Reason = reason
    }
}

func WithTenant(tenantID string) QueryOption {
    return func(opts *QueryOptions) {
        opts.TenantID = &tenantID
    }
}

func WithRequestID(requestID string) QueryOption {
    return func(opts *QueryOptions) {
        opts.RequestID = requestID
    }
}

func WithAsOf(timestamp time.Time) QueryOption {
    return func(opts *QueryOptions) {
        opts.AsOf = &timestamp
    }
}
```

### 2. Context Management

```go
// pkg/context/query_context.go
package context

import (
    "context"
    "kisanlink-ecom/pkg/repository"
)

type contextKey string

const (
    queryOptionsKey contextKey = "queryOptions"
    auditContextKey contextKey = "auditContext"
)

// WithQueryOptions adds query options to context
func WithQueryOptions(ctx context.Context, opts repository.QueryOptions) context.Context {
    // Validate options before storing
    if err := opts.Validate(); err != nil {
        // Log error but use defaults
        logrus.Warnf("Invalid query options: %v, using defaults", err)
        opts = repository.DefaultQueryOptions()
    }

    return context.WithValue(ctx, queryOptionsKey, opts)
}

// GetQueryOptions retrieves options from context
func GetQueryOptions(ctx context.Context) repository.QueryOptions {
    if opts, ok := ctx.Value(queryOptionsKey).(repository.QueryOptions); ok {
        return opts
    }
    return repository.DefaultQueryOptions()
}

// MustGetQueryOptions panics if options not in context (for testing)
func MustGetQueryOptions(ctx context.Context) repository.QueryOptions {
    opts, ok := ctx.Value(queryOptionsKey).(repository.QueryOptions)
    if !ok {
        panic("query options not found in context")
    }
    return opts
}
```

### 3. Base Repository Implementation

```go
// pkg/repository/base_repository.go
package repository

import (
    "context"
    "fmt"
    "gorm.io/gorm"
    "kisanlink-ecom/pkg/context"
    "kisanlink-ecom/pkg/security"
)

// BaseRepository provides common soft delete handling
type BaseRepository struct {
    db *gorm.DB
    auditLogger security.AuditLogger
}

// ApplyQueryOptions applies options to a GORM query
func (r *BaseRepository) ApplyQueryOptions(ctx context.Context, db *gorm.DB) *gorm.DB {
    opts := context.GetQueryOptions(ctx)

    // Handle soft deletes
    if opts.IncludeDeleted {
        // Audit the access
        r.auditLogger.LogDeletedAccess(ctx, opts)
        db = db.Unscoped()
    }

    // Future: Apply tenant filtering
    if opts.TenantID != nil {
        db = db.Where("organization_id = ?", *opts.TenantID)
    }

    // Future: Time-travel queries
    if opts.AsOf != nil {
        db = db.Where("created_at <= ? AND (deleted_at IS NULL OR deleted_at > ?)",
            *opts.AsOf, *opts.AsOf)
    }

    return db
}

// Example method showing usage
func (r *BaseRepository) FindByID(ctx context.Context, id string, model interface{}) error {
    db := r.ApplyQueryOptions(ctx, r.db)
    return db.Where("id = ?", id).First(model).Error
}
```

### 4. Repository Implementation Example

```go
// internal/repositories/catalog/catalog_repository.go
package catalog

import (
    "context"
    "kisanlink-ecom/entities/models/catalog"
    "kisanlink-ecom/pkg/repository"
)

type CatalogRepository struct {
    repository.BaseRepository
    dbManager db.DBManager
}

// GetByID with soft delete filtering
func (r *CatalogRepository) GetByID(ctx context.Context, id string) (*catalog.CatalogItem, error) {
    var item catalog.CatalogItem

    // Get GORM DB instance
    db := r.dbManager.DB()

    // Apply query options (including soft delete filtering)
    db = r.ApplyQueryOptions(ctx, db)

    // Execute query
    if err := db.Where("id = ?", id).First(&item).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrItemNotFound
        }
        return nil, fmt.Errorf("failed to get catalog item: %w", err)
    }

    return &item, nil
}

// Find with comprehensive filtering
func (r *CatalogRepository) Find(ctx context.Context, filter *base.Filter) ([]*catalog.CatalogItem, error) {
    var items []*catalog.CatalogItem

    db := r.dbManager.DB()
    db = r.ApplyQueryOptions(ctx, db)

    // Apply additional filters
    if filter != nil {
        db = r.applyFilters(db, filter)
    }

    if err := db.Find(&items).Error; err != nil {
        return nil, fmt.Errorf("failed to find catalog items: %w", err)
    }

    return items, nil
}

// Explicit soft delete with proper tracking
func (r *CatalogRepository) SoftDelete(ctx context.Context, id string, deletedBy string) error {
    db := r.dbManager.DB()

    // Use Unscoped to update even if already soft-deleted
    result := db.Unscoped().Model(&catalog.CatalogItem{}).
        Where("id = ?", id).
        Updates(map[string]interface{}{
            "deleted_at": gorm.DeletedAt{Time: time.Now(), Valid: true},
            "deleted_by": deletedBy,
        })

    if result.Error != nil {
        return fmt.Errorf("failed to soft delete: %w", result.Error)
    }

    if result.RowsAffected == 0 {
        return ErrItemNotFound
    }

    // Audit the deletion
    r.auditLogger.LogDeletion(ctx, "catalog_item", id, deletedBy)

    return nil
}

// Restore with audit trail
func (r *CatalogRepository) Restore(ctx context.Context, id string, restoredBy string) error {
    db := r.dbManager.DB()

    result := db.Unscoped().Model(&catalog.CatalogItem{}).
        Where("id = ? AND deleted_at IS NOT NULL", id).
        Updates(map[string]interface{}{
            "deleted_at": gorm.DeletedAt{Valid: false},
            "deleted_by": nil,
            "updated_by": restoredBy,
        })

    if result.Error != nil {
        return fmt.Errorf("failed to restore: %w", result.Error)
    }

    if result.RowsAffected == 0 {
        return ErrItemNotFound
    }

    // Audit the restoration
    r.auditLogger.LogRestoration(ctx, "catalog_item", id, restoredBy)

    return nil
}
```

### 5. Service Layer Implementation

```go
// internal/services/catalog/catalog_service.go
package catalog

import (
    "context"
    "kisanlink-ecom/pkg/repository"
    pkgcontext "kisanlink-ecom/pkg/context"
)

type CatalogService struct {
    catalogRepo CatalogRepositoryInterface
}

// ListCatalogItems with optional deleted items
func (s *CatalogService) ListCatalogItems(
    ctx context.Context,
    filter *catalogRequests.CatalogFilter,
    offset, limit int,
    opts ...repository.QueryOption,
) ([]*catalog.CatalogItem, int, error) {
    // Build query options
    queryOpts := repository.DefaultQueryOptions()
    for _, opt := range opts {
        opt(&queryOpts)
    }

    // Validate business rules
    if queryOpts.IncludeDeleted {
        // Additional business validation for deleted items access
        if queryOpts.Reason == "" {
            return nil, 0, fmt.Errorf("reason required for accessing deleted items")
        }
    }

    // Propagate options via context
    ctx = pkgcontext.WithQueryOptions(ctx, queryOpts)

    // Call repository with enriched context
    items, err := s.catalogRepo.Find(ctx, filter)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to list catalog items: %w", err)
    }

    // Get total count for pagination
    total, err := s.catalogRepo.Count(ctx, filter)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to count catalog items: %w", err)
    }

    return items, total, nil
}

// GetCatalogItemByID with controlled deleted access
func (s *CatalogService) GetCatalogItemByID(
    ctx context.Context,
    id string,
    opts ...repository.QueryOption,
) (*catalog.CatalogItem, error) {
    queryOpts := repository.DefaultQueryOptions()
    for _, opt := range opts {
        opt(&queryOpts)
    }

    ctx = pkgcontext.WithQueryOptions(ctx, queryOpts)

    item, err := s.catalogRepo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, ErrItemNotFound) {
            // Check if item exists but is deleted
            if !queryOpts.IncludeDeleted {
                deletedCtx := pkgcontext.WithQueryOptions(ctx,
                    repository.QueryOptions{IncludeDeleted: true})
                if _, err := s.catalogRepo.GetByID(deletedCtx, id); err == nil {
                    return nil, ErrItemDeleted // More specific error
                }
            }
        }
        return nil, err
    }

    return item, nil
}
```

### 6. Handler Layer with Authorization

```go
// internal/handlers/catalog/catalog_handler.go
package catalog

import (
    "kisanlink-ecom/pkg/repository"
    "kisanlink-ecom/internal/middleware"
    "kisanlink-ecom/internal/security"
)

type CatalogHandler struct {
    catalogService CatalogServiceInterface
    authService    security.AuthorizationService
}

// ListCatalogItems with deleted items access control
func (h *CatalogHandler) ListCatalogItems(c *gin.Context) {
    // Parse standard parameters
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

    // Build query options based on request
    var queryOpts []repository.QueryOption

    // Check if user requested deleted items
    if includeDeleted, _ := strconv.ParseBool(c.Query("include_deleted")); includeDeleted {
        // Get user from context
        userClaims, exists := middleware.GetUserClaims(c)
        if !exists {
            common.Unauthorized(c, "AUTH_REQUIRED",
                "Authentication required to view deleted items")
            return
        }

        // Check permission
        canViewDeleted, err := h.authService.CanViewDeleted(c.Request.Context(),
            userClaims.UserID, "catalog_items")
        if err != nil {
            common.InternalServerError(c, "AUTH_CHECK_FAILED",
                "Failed to check permissions", nil)
            return
        }

        if !canViewDeleted {
            common.Forbidden(c, "INSUFFICIENT_PERMISSIONS",
                "You do not have permission to view deleted items")
            return
        }

        // Get reason for audit
        reason := c.Query("reason")
        if reason == "" {
            common.BadRequest(c, "REASON_REQUIRED",
                "Reason required when accessing deleted items", nil)
            return
        }

        // Add option with reason
        queryOpts = append(queryOpts,
            repository.WithDeleted(reason),
            repository.WithRequestID(middleware.GetRequestID(c)))
    }

    // Add tenant filtering if in multi-tenant mode
    if orgID, exists := middleware.GetOrgID(c); exists {
        queryOpts = append(queryOpts, repository.WithTenant(orgID))
    }

    // Parse filter
    filter := h.parseFilter(c)

    // Call service with options
    items, total, err := h.catalogService.ListCatalogItems(
        c.Request.Context(), filter, offset, limit, queryOpts...)

    if err != nil {
        common.InternalServerError(c, "LIST_FAILED",
            "Failed to list catalog items", map[string]interface{}{
                "error": err.Error(),
            })
        return
    }

    // Transform response
    response := h.transformItems(items)

    // Add metadata if viewing deleted
    meta := &common.ResponseMeta{
        TraceID: middleware.GetTraceID(c),
        Pagination: &common.PaginationMeta{
            Page:    page,
            Limit:   limit,
            Total:   total,
            HasNext: len(items) == limit,
        },
    }

    if len(queryOpts) > 0 {
        meta.QueryOptions = map[string]interface{}{
            "include_deleted": includeDeleted,
        }
    }

    common.Success(c, response, meta)
}

// GetCatalogItem with soft delete awareness
func (h *CatalogHandler) GetCatalogItem(c *gin.Context) {
    itemID := c.Param("id")

    var queryOpts []repository.QueryOption

    // Check if should include deleted
    if includeDeleted, _ := strconv.ParseBool(c.Query("include_deleted")); includeDeleted {
        // Similar authorization check as above
        // ... (authorization code)

        queryOpts = append(queryOpts,
            repository.WithDeleted("Direct item access"),
            repository.WithRequestID(middleware.GetRequestID(c)))
    }

    item, err := h.catalogService.GetCatalogItemByID(
        c.Request.Context(), itemID, queryOpts...)

    if err != nil {
        switch {
        case errors.Is(err, ErrItemNotFound):
            common.NotFound(c, "ITEM_NOT_FOUND",
                "Catalog item not found", nil)
        case errors.Is(err, ErrItemDeleted):
            common.Gone(c, "ITEM_DELETED",
                "Catalog item has been deleted", nil)
        default:
            common.InternalServerError(c, "GET_FAILED",
                "Failed to get catalog item", nil)
        }
        return
    }

    common.Success(c, h.transformItem(item), nil)
}
```

### 7. Security and Permissions

```go
// pkg/security/permissions.go
package security

import (
    "context"
    "fmt"
)

type Permission string

const (
    PermissionViewDeleted    Permission = "view_deleted"
    PermissionRestoreDeleted Permission = "restore_deleted"
    PermissionPermanentDelete Permission = "permanent_delete"
)

type AuthorizationService interface {
    CanViewDeleted(ctx context.Context, userID, resource string) (bool, error)
    CanRestoreDeleted(ctx context.Context, userID, resource string) (bool, error)
    CanPermanentDelete(ctx context.Context, userID, resource string) (bool, error)
}

type authorizationService struct {
    permissionChecker PermissionChecker // External service or database
}

func (s *authorizationService) CanViewDeleted(ctx context.Context, userID, resource string) (bool, error) {
    // Check user roles and permissions
    permissions, err := s.permissionChecker.GetUserPermissions(ctx, userID)
    if err != nil {
        return false, fmt.Errorf("failed to get user permissions: %w", err)
    }

    // Check for specific permission
    requiredPerm := fmt.Sprintf("%s:%s", resource, PermissionViewDeleted)
    for _, perm := range permissions {
        if perm == requiredPerm || perm == "*:view_deleted" {
            return true, nil
        }
    }

    // Check if user is admin
    if s.permissionChecker.IsAdmin(ctx, userID) {
        return true, nil
    }

    return false, nil
}
```

### 8. Audit Logging

```go
// pkg/security/audit.go
package security

import (
    "context"
    "time"
    "kisanlink-ecom/pkg/repository"
)

type AuditLogger interface {
    LogDeletedAccess(ctx context.Context, opts repository.QueryOptions)
    LogDeletion(ctx context.Context, resourceType, resourceID, deletedBy string)
    LogRestoration(ctx context.Context, resourceType, resourceID, restoredBy string)
}

type auditLogger struct {
    logger *logrus.Logger
    db     *gorm.DB // For persistent audit trails
}

func (l *auditLogger) LogDeletedAccess(ctx context.Context, opts repository.QueryOptions) {
    entry := AuditEntry{
        Timestamp:   time.Now(),
        RequestID:   opts.RequestID,
        UserID:      opts.UserID,
        Action:      "VIEW_DELETED",
        Reason:      opts.Reason,
        ResourceType: "multiple",
    }

    // Log to structured logger
    l.logger.WithFields(logrus.Fields{
        "request_id": entry.RequestID,
        "user_id":    entry.UserID,
        "action":     entry.Action,
        "reason":     entry.Reason,
    }).Info("Accessed deleted items")

    // Persist to database
    if err := l.db.Create(&entry).Error; err != nil {
        l.logger.WithError(err).Error("Failed to persist audit log")
    }
}
```

## Migration Strategy

### Phase 1: Infrastructure Setup

```bash
# Create new packages
mkdir -p pkg/repository pkg/context pkg/security
# Implement core components with tests
```

### Phase 2: Repository Updates

```go
// Update each repository incrementally
// 1. Add BaseRepository embedding
// 2. Update methods to use ApplyQueryOptions
// 3. Test with both deleted and non-deleted items
```

### Phase 3: Service Layer

```go
// Add optional QueryOption parameters
// Maintain backward compatibility:
func (s *Service) OldMethod(ctx context.Context) {
    // Calls new method with default options
    return s.NewMethod(ctx)
}

func (s *Service) NewMethod(ctx context.Context, opts ...QueryOption) {
    // Implementation with options
}
```

### Phase 4: Handler Updates

```go
// Add query parameter handling
// Implement authorization checks
// Update API documentation
```

## Testing Strategy

### Unit Tests

```go
// pkg/repository/options_test.go
func TestQueryOptions_Validate(t *testing.T) {
    tests := []struct {
        name    string
        opts    QueryOptions
        wantErr bool
    }{
        {
            name: "valid with deleted and reason",
            opts: QueryOptions{
                IncludeDeleted: true,
                Reason: "audit review",
            },
            wantErr: false,
        },
        {
            name: "invalid deleted without reason",
            opts: QueryOptions{
                IncludeDeleted: true,
            },
            wantErr: true,
        },
    }
    // ...
}
```

### Integration Tests

```go
// tests/integration/soft_delete_test.go
func TestSoftDeleteFiltering(t *testing.T) {
    // Setup
    repo := setupTestRepository(t)
    ctx := context.Background()

    // Create and soft delete an item
    item := createTestItem(t)
    repo.Create(ctx, item)
    repo.SoftDelete(ctx, item.ID, "test_user")

    // Test 1: Default query should not return deleted
    items, err := repo.Find(ctx, nil)
    assert.NoError(t, err)
    assert.Len(t, items, 0)

    // Test 2: With deleted option should return item
    ctxWithDeleted := context.WithQueryOptions(ctx,
        repository.QueryOptions{
            IncludeDeleted: true,
            Reason: "test",
        })
    items, err = repo.Find(ctxWithDeleted, nil)
    assert.NoError(t, err)
    assert.Len(t, items, 1)
}
```

### End-to-End Tests

```go
// tests/e2e/catalog_deleted_items_test.go
func TestCatalogDeletedItemsE2E(t *testing.T) {
    // Setup test server
    app := setupTestApp(t)

    // Create and delete item
    itemID := createTestCatalogItem(t)
    deleteTestCatalogItem(t, itemID)

    // Test 1: Regular user cannot see deleted
    req := httptest.NewRequest("GET", "/api/v1/catalog?include_deleted=true", nil)
    req.Header.Set("Authorization", "Bearer " + regularUserToken)

    w := httptest.NewRecorder()
    app.ServeHTTP(w, req)

    assert.Equal(t, http.StatusForbidden, w.Code)

    // Test 2: Admin can see deleted with reason
    req = httptest.NewRequest("GET",
        "/api/v1/catalog?include_deleted=true&reason=audit", nil)
    req.Header.Set("Authorization", "Bearer " + adminToken)

    w = httptest.NewRecorder()
    app.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    // Assert response contains deleted items
}
```

## Performance Considerations

### Database Indexes

```sql
-- Ensure proper indexes for soft delete queries
CREATE INDEX idx_catalog_items_deleted ON catalog_items(deleted_at);
CREATE INDEX idx_catalog_items_org_deleted ON catalog_items(organization_id, deleted_at);

-- Composite indexes for common queries
CREATE INDEX idx_catalog_items_type_deleted ON catalog_items(item_type, deleted_at);
CREATE INDEX idx_catalog_items_category_deleted ON catalog_items(category, deleted_at);
```

### Query Optimization

```go
// Use selective queries to avoid loading unnecessary data
func (r *Repository) FindActive(ctx context.Context) ([]*Model, error) {
    // Explicitly filter at database level
    return r.db.Where("deleted_at IS NULL AND is_active = ?", true).Find(&items).Error
}
```

### Caching Strategy

```go
// Cache keys should include deleted status
func getCacheKey(id string, includeDeleted bool) string {
    if includeDeleted {
        return fmt.Sprintf("item:%s:all", id)
    }
    return fmt.Sprintf("item:%s:active", id)
}
```

## Monitoring and Observability

### Metrics

```go
// Track soft delete access patterns
var (
    deletedItemsAccessed = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "soft_deleted_items_accessed_total",
            Help: "Total number of soft deleted items accessed",
        },
        []string{"resource_type", "user_role"},
    )

    softDeleteOperations = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "soft_delete_operations_total",
            Help: "Total number of soft delete operations",
        },
        []string{"operation", "resource_type"},
    )
)
```

### Logging

```go
// Structured logging for soft delete operations
logger.WithFields(logrus.Fields{
    "operation":     "soft_delete",
    "resource_type": "catalog_item",
    "resource_id":   itemID,
    "deleted_by":    userID,
    "timestamp":     time.Now(),
}).Info("Item soft deleted")
```

### Tracing

```go
// Add spans for soft delete operations
func (r *Repository) SoftDelete(ctx context.Context, id, deletedBy string) error {
    span, ctx := opentracing.StartSpanFromContext(ctx, "repository.soft_delete")
    defer span.Finish()

    span.SetTag("resource.id", id)
    span.SetTag("deleted.by", deletedBy)

    // ... implementation
}
```

## API Documentation Updates

### OpenAPI Specification

```yaml
/api/v1/catalog:
  get:
    parameters:
      - name: include_deleted
        in: query
        required: false
        schema:
          type: boolean
          default: false
        description: Include soft-deleted items (requires admin permission)
      - name: reason
        in: query
        required: false
        schema:
          type: string
        description: Reason for accessing deleted items (required when include_deleted=true)
    responses:
      403:
        description: Forbidden - insufficient permissions to view deleted items
```

## Rollback Plan

### Feature Flags

```go
// Use feature flags for gradual rollout
if featureFlags.IsEnabled("soft_delete_filtering_v2") {
    // New implementation
    ctx = context.WithQueryOptions(ctx, opts)
} else {
    // Fallback to old implementation
}
```

### Backward Compatibility

- All existing methods continue to work (filter deleted by default)
- New optional parameters don't break existing calls
- Database schema unchanged (uses existing deleted_at field)

### Emergency Rollback

```bash
# Disable feature flag
kubectl set env deployment/api FEATURE_SOFT_DELETE_FILTERING_V2=false

# Or revert to previous version
kubectl rollout undo deployment/api
```

## Security Checklist

- [ ] Authorization checks implemented for deleted item access
- [ ] Audit logging for all deleted item access
- [ ] Reason field required and validated
- [ ] No sensitive data exposed in error messages
- [ ] Rate limiting on deleted item queries
- [ ] SQL injection prevention verified
- [ ] Permission checks cannot be bypassed
- [ ] Deleted items properly filtered in all endpoints

## Performance Checklist

- [ ] Indexes created on deleted_at columns
- [ ] Query plans analyzed for common patterns
- [ ] N+1 query problems addressed
- [ ] Batch operations optimized
- [ ] Cache invalidation handles soft deletes
- [ ] Load testing performed with large datasets
- [ ] Monitoring alerts configured

## References

- [GORM Soft Delete Guide](https://gorm.io/docs/delete.html#Soft-Delete)
- [PostgreSQL Index Design](https://www.postgresql.org/docs/current/indexes.html)
- [Go Context Patterns](https://go.dev/blog/context)
- [OWASP Authorization Testing](https://owasp.org/www-project-web-security-testing-guide/latest/4-Web_Application_Security_Testing/05-Authorization_Testing/)
