# ADR-003: Soft Delete Filtering Pattern

## Status

Proposed

## Context

The Kisanlink e-commerce platform extensively uses soft deletes across 13+ models to maintain data integrity and support audit trails. Currently, the implementation is inconsistent:

1. **Inconsistent Repository Patterns**:
   - CatalogRepository: Explicit `SoftDelete()` and `Restore()` methods
   - CollaboratorRepository: Explicit `SoftDelete()` method only
   - InventoryRepository: Implicit deletion via `dbManager.Delete()`
   - OutboxRepository: Manual `deleted_at IS NULL` checks in raw queries

2. **Current State**:
   - All models use `gorm.DeletedAt` field for soft deletes
   - GORM automatically filters soft-deleted records (when using GORM methods)
   - No consistent way to include soft-deleted records when needed
   - No clear security boundaries for who can access deleted data
   - Architecture follows: Routes → Handlers → Services → Repositories → DB

3. **Requirements**:
   - Default filtering of soft-deleted items across all layers
   - Optional inclusion of soft-deleted items with proper authorization
   - Type-safe, compile-time checked solution
   - Backward compatibility with existing code
   - Extensibility for future filtering needs (tenant isolation, permissions)

## Decision

We will implement a **Query Options Pattern** with context-based propagation for soft delete filtering:

### 1. Repository Layer Options

```go
// pkg/repository/options.go
type QueryOptions struct {
    IncludeDeleted bool
    TenantID       *string  // Future: tenant isolation
    UserID         *string  // Future: user-specific filtering
}

func DefaultQueryOptions() QueryOptions {
    return QueryOptions{
        IncludeDeleted: false,
    }
}

// Functional options pattern for flexibility
type QueryOption func(*QueryOptions)

func WithDeleted() QueryOption {
    return func(opts *QueryOptions) {
        opts.IncludeDeleted = true
    }
}

func WithTenant(tenantID string) QueryOption {
    return func(opts *QueryOptions) {
        opts.TenantID = &tenantID
    }
}
```

### 2. Context Propagation

```go
// pkg/context/query_context.go
type contextKey string

const queryOptionsKey contextKey = "queryOptions"

func WithQueryOptions(ctx context.Context, opts QueryOptions) context.Context {
    return context.WithValue(ctx, queryOptionsKey, opts)
}

func GetQueryOptions(ctx context.Context) QueryOptions {
    if opts, ok := ctx.Value(queryOptionsKey).(QueryOptions); ok {
        return opts
    }
    return DefaultQueryOptions()
}
```

### 3. Repository Implementation

```go
// Apply at repository level
func (r *CatalogRepository) Find(ctx context.Context, filter *base.Filter) ([]*catalog.CatalogItem, error) {
    opts := GetQueryOptions(ctx)

    db := r.dbManager.DB()
    if !opts.IncludeDeleted {
        // GORM handles this by default, but be explicit
        db = db.Where("deleted_at IS NULL")
    } else {
        // Include soft-deleted records
        db = db.Unscoped()
    }

    // Apply other filters...
    return items, db.Find(&items).Error
}
```

### 4. Service Layer API

```go
// Service methods accept options as last parameter
func (s *CatalogService) ListCatalogItems(
    ctx context.Context,
    filter *catalogRequests.CatalogFilter,
    offset, limit int,
    opts ...QueryOption,
) ([]*catalog.CatalogItem, int, error) {
    queryOpts := DefaultQueryOptions()
    for _, opt := range opts {
        opt(&queryOpts)
    }

    // Propagate to repository via context
    ctx = WithQueryOptions(ctx, queryOpts)
    return s.catalogRepo.ListCatalogItems(ctx, filter, offset, limit)
}
```

### 5. Handler Layer Authorization

```go
// Handler checks authorization before allowing deleted items
func (h *CatalogHandler) ListCatalogItems(c *gin.Context) {
    var queryOpts []QueryOption

    // Check if user requested deleted items
    if includeDeleted, _ := strconv.ParseBool(c.Query("include_deleted")); includeDeleted {
        // Verify user has permission
        if !h.authService.CanViewDeleted(c) {
            common.Forbidden(c, "INSUFFICIENT_PERMISSIONS",
                "You do not have permission to view deleted items")
            return
        }
        queryOpts = append(queryOpts, WithDeleted())
    }

    items, total, err := h.catalogService.ListCatalogItems(
        c.Request.Context(), filter, offset, limit, queryOpts...)
    // ...
}
```

## Alternatives Considered

### 1. Global GORM Scope Modification

- **Pros**: Minimal code changes, works automatically
- **Cons**: Hard to control per-request, affects all queries globally, security concerns

### 2. Separate Methods for Deleted Items

- **Pros**: Explicit intent, clear API boundaries
- **Cons**: API proliferation (GetByID vs GetByIDWithDeleted), maintenance burden

### 3. Query Parameter in Every Method

- **Pros**: Explicit, simple to understand
- **Cons**: Breaking API changes, parameter pollution, harder to extend

### 4. Middleware-Based Filtering

- **Pros**: Centralized control, clean handlers
- **Cons**: Hidden behavior, harder to debug, less flexible per-endpoint

## Consequences

### Positive

1. **Consistency**: Uniform pattern across all repositories and services
2. **Security**: Clear authorization checks at handler level
3. **Extensibility**: Easy to add new filtering dimensions (tenant, permissions)
4. **Type Safety**: Compile-time checked options
5. **Backward Compatible**: Existing code continues to work (defaults to filtering deleted)
6. **Performance**: Leverages database indexes on deleted_at
7. **Auditability**: Clear when deleted items are accessed

### Negative

1. **Context Propagation**: Requires passing context through all layers
2. **Learning Curve**: Developers need to understand the pattern
3. **Boilerplate**: Some additional code for options handling

### Neutral

1. **Testing Complexity**: Need to test both with and without deleted items
2. **Documentation**: Requires clear documentation of when deleted items are accessible

## Implementation Plan

### Phase 1: Core Infrastructure (Week 1)

1. Create options package with QueryOptions struct
2. Implement context helpers for options propagation
3. Create base repository interface with options support

### Phase 2: Repository Migration (Week 2)

1. Update each repository to use QueryOptions from context
2. Ensure backward compatibility with existing methods
3. Add comprehensive tests for deleted item filtering

### Phase 3: Service Layer Updates (Week 3)

1. Add optional QueryOption parameters to service methods
2. Propagate options to repositories via context
3. Maintain backward compatibility

### Phase 4: Handler Authorization (Week 4)

1. Add include_deleted query parameter support
2. Implement authorization checks for deleted item access
3. Update API documentation

### Phase 5: Monitoring & Rollout (Week 5)

1. Add metrics for deleted item access
2. Gradual rollout with feature flags
3. Monitor for any performance impacts

## Security Considerations

1. **Authorization Levels**:
   - Regular users: Cannot view deleted items
   - Admins: Can view deleted items with audit logging
   - System: Can access deleted items for maintenance

2. **Audit Trail**:
   - Log all access to deleted items
   - Include user, timestamp, and items accessed

3. **Performance**:
   - Ensure deleted_at columns are properly indexed
   - Monitor query performance with Unscoped() calls

## References

- [GORM Soft Delete Documentation](https://gorm.io/docs/delete.html#Soft-Delete)
- [Go Context Best Practices](https://go.dev/blog/context)
- [Functional Options Pattern](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis)
- [OWASP Authorization Best Practices](https://owasp.org/www-project-top-ten/)
