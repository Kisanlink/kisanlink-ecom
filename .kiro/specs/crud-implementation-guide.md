# CRUD Implementation Guide for Missing Models

## Overview

This document provides a comprehensive implementation guide for creating complete CRUD operations for the 9 models that were recently added to the migration registry but lack full implementation.

## Models Requiring Implementation

1. **User** (`entities/models/user/user.go`) → `users` table [PARTIAL - Needs repository]
2. **UserRole** (`entities/models/roles/user_role.go`) → `user_roles` table
3. **OrganizationRole** (`entities/models/roles/organization_role.go`) → `organization_roles` table
4. **EcommerceRole** (`entities/models/roles/ecommerce_role.go`) → `ecommerce_roles` table
5. **Collaborator** (`entities/models/collaborator/collaborator.go`) → `collaborators` table [EXISTS - Use as reference]
6. **TaxExemption** (`entities/models/taxation/tax.go`) → `tax_exemptions` table
7. **ServiceSLA** (`entities/models/services/sla.go`) → `service_slas` table
8. **DiscountRule** (`entities/models/discounts/discount.go`) → `discount_rules` table
9. **OrganizationCollaborator** (`entities/models/actors/collaborator.go`) → `organization_collaborators` table

## Implementation Pattern

For each model, implement the following layers in order:

### 1. Repository Layer (`internal/repositories/{domain}/{model}_repository.go`)

```go
package {domain}

import (
    "context"
    "fmt"

    "{model_import_path}"
    "kisanlink-ecom/internal/repositories/common"

    "github.com/Kisanlink/kisanlink-db/pkg/base"
    "github.com/Kisanlink/kisanlink-db/pkg/db"
)

// {Model}Repository handles {model} operations
type {Model}Repository struct {
    *common.BaseRepository
    dbManager db.DBManager
}

// New{Model}Repository creates a new repository
func New{Model}Repository(dbManager db.DBManager) *{Model}Repository {
    return &{Model}Repository{
        BaseRepository: common.NewBaseRepository(dbManager),
        dbManager:      dbManager,
    }
}

// Create creates a new {model}
func (r *{Model}Repository) Create(ctx context.Context, model *{Model}) error {
    return r.dbManager.Create(ctx, model)
}

// GetByID retrieves by ID with soft delete filtering
func (r *{Model}Repository) GetByID(ctx context.Context, id string) (*{Model}, error) {
    var item {Model}
    opts := common.QueryOptionsFromContext(ctx)

    if opts.IncludeDeleted {
        if err := r.dbManager.GetByID(ctx, id, &item); err != nil {
            return nil, fmt.Errorf("failed to get {model}: %w", err)
        }
        return &item, nil
    }

    // Filter out deleted items
    filter := base.NewFilter()
    filter.Group.Conditions = []base.FilterCondition{
        {Field: "id", Operator: base.OpEqual, Value: id},
    }

    filter = r.ApplyQueryOptions(ctx, filter)

    var items []*{Model}
    if err := r.dbManager.List(ctx, filter, &items); err != nil {
        return nil, fmt.Errorf("failed to get {model}: %w", err)
    }

    if len(items) == 0 {
        return nil, fmt.Errorf("{model} not found")
    }

    return items[0], nil
}

// Update updates an existing {model}
func (r *{Model}Repository) Update(ctx context.Context, model *{Model}) error {
    return r.dbManager.Update(ctx, model)
}

// Delete soft deletes a {model}
func (r *{Model}Repository) Delete(ctx context.Context, id string) error {
    item := &{Model}{}
    return r.dbManager.Delete(ctx, id, item)
}

// List retrieves {models} with filtering and pagination
func (r *{Model}Repository) List(ctx context.Context, filter *base.Filter, limit, offset int) ([]*{Model}, int, error) {
    dbFilter := filter
    if dbFilter == nil {
        dbFilter = base.NewFilter()
    }

    // Apply soft delete filtering
    dbFilter = r.ApplyQueryOptions(ctx, dbFilter)

    dbFilter.Limit = limit
    dbFilter.Offset = offset

    var items []*{Model}
    if err := r.dbManager.List(ctx, dbFilter, &items); err != nil {
        return nil, 0, fmt.Errorf("failed to list {models}: %w", err)
    }

    total, err := r.dbManager.Count(ctx, dbFilter, &{Model}{})
    if err != nil {
        return nil, 0, fmt.Errorf("failed to count {models}: %w", err)
    }

    return items, int(total), nil
}
```

### 2. Service Layer (`internal/services/{domain}/{model}_service.go`)

```go
package {domain}

import (
    "context"
    "fmt"

    "{model_import_path}"
)

// {Model}RepositoryInterface defines repository operations
type {Model}RepositoryInterface interface {
    Create(ctx context.Context, model *{Model}) error
    GetByID(ctx context.Context, id string) (*{Model}, error)
    Update(ctx context.Context, model *{Model}) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter *base.Filter, limit, offset int) ([]*{Model}, int, error)
}

// {Model}ServiceInterface defines service operations
type {Model}ServiceInterface interface {
    Create{Model}(ctx context.Context, model *{Model}) (*{Model}, error)
    Get{Model}ByID(ctx context.Context, id string) (*{Model}, error)
    Update{Model}(ctx context.Context, model *{Model}) (*{Model}, error)
    Delete{Model}(ctx context.Context, id string) error
    List{Model}s(ctx context.Context, limit, offset int) ([]*{Model}, int, error)
}

// {Model}Service provides business logic
type {Model}Service struct {
    repo {Model}RepositoryInterface
}

// New{Model}Service creates a new service
func New{Model}Service(repo {Model}RepositoryInterface) *{Model}Service {
    return &{Model}Service{repo: repo}
}

// Create{Model} creates a new {model} with validation
func (s *{Model}Service) Create{Model}(ctx context.Context, model *{Model}) (*{Model}, error) {
    // Validation logic here
    if model == nil {
        return nil, fmt.Errorf("{model} cannot be nil")
    }

    // Business rules validation
    if err := s.validate{Model}(model); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    if err := s.repo.Create(ctx, model); err != nil {
        return nil, fmt.Errorf("failed to create {model}: %w", err)
    }

    return model, nil
}

// Get{Model}ByID retrieves a {model} by ID
func (s *{Model}Service) Get{Model}ByID(ctx context.Context, id string) (*{Model}, error) {
    if id == "" {
        return nil, fmt.Errorf("{model} ID is required")
    }

    model, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get {model}: %w", err)
    }

    return model, nil
}

// Update{Model} updates an existing {model}
func (s *{Model}Service) Update{Model}(ctx context.Context, model *{Model}) (*{Model}, error) {
    if model == nil {
        return nil, fmt.Errorf("{model} cannot be nil")
    }

    // Validate exists
    existing, err := s.repo.GetByID(ctx, model.ID)
    if err != nil || existing == nil {
        return nil, fmt.Errorf("{model} not found")
    }

    // Business rules validation
    if err := s.validate{Model}(model); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    if err := s.repo.Update(ctx, model); err != nil {
        return nil, fmt.Errorf("failed to update {model}: %w", err)
    }

    return model, nil
}

// Delete{Model} soft deletes a {model}
func (s *{Model}Service) Delete{Model}(ctx context.Context, id string) error {
    if id == "" {
        return fmt.Errorf("{model} ID is required")
    }

    // Check exists
    _, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return fmt.Errorf("{model} not found")
    }

    if err := s.repo.Delete(ctx, id); err != nil {
        return fmt.Errorf("failed to delete {model}: %w", err)
    }

    return nil
}

// List{Model}s retrieves {models} with pagination
func (s *{Model}Service) List{Model}s(ctx context.Context, limit, offset int) ([]*{Model}, int, error) {
    if limit < 1 {
        limit = 20
    }
    if limit > 100 {
        limit = 100
    }
    if offset < 0 {
        offset = 0
    }

    models, total, err := s.repo.List(ctx, nil, limit, offset)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to list {models}: %w", err)
    }

    return models, total, nil
}

// validate{Model} validates {model} data
func (s *{Model}Service) validate{Model}(model *{Model}) error {
    // Add specific validation rules
    return nil
}
```

### 3. Handler Layer (`internal/handlers/{domain}/{model}_handler.go`)

```go
package {domain}

import (
    "strconv"

    "{model_import_path}"
    "kisanlink-ecom/internal/common"
    "{service_import_path}"

    "github.com/gin-gonic/gin"
)

// {Model}Handler handles HTTP requests for {model} operations
type {Model}Handler struct {
    service {service_package}.{Model}ServiceInterface
}

// New{Model}Handler creates a new handler
func New{Model}Handler(service {service_package}.{Model}ServiceInterface) *{Model}Handler {
    return &{Model}Handler{service: service}
}

// Create{Model} godoc
// @Summary Create a new {model}
// @Description Create a new {model}
// @Tags {models}
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param {model} body {Model}CreateRequest true "{Model} data"
// @Success 201 {object} common.Response{data={Model}}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/{models} [post]
func (h *{Model}Handler) Create{Model}(c *gin.Context) {
    var req {Model}CreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
            "error": err.Error(),
        })
        return
    }

    // Convert request to model
    model := req.ToModel()

    // Get user ID from context
    if userID, exists := c.Get("subjectID"); exists {
        model.SetCreatedBy(userID.(string))
    }

    created, err := h.service.Create{Model}(c.Request.Context(), model)
    if err != nil {
        common.InternalServerError(c, "CREATE_FAILED", "Failed to create {model}", map[string]interface{}{
            "error": err.Error(),
        })
        return
    }

    common.Created(c, created, &common.ResponseMeta{
        TraceID: common.GetTraceID(c),
    })
}

// Get{Model} godoc
// @Summary Get {model} by ID
// @Description Retrieve a specific {model} by ID
// @Tags {models}
// @Accept json
// @Produce json
// @Param id path string true "{Model} ID"
// @Success 200 {object} common.Response{data={Model}}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/{models}/{id} [get]
func (h *{Model}Handler) Get{Model}(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        common.BadRequest(c, "MISSING_ID", "{Model} ID is required", nil)
        return
    }

    model, err := h.service.Get{Model}ByID(c.Request.Context(), id)
    if err != nil {
        common.NotFound(c, "NOT_FOUND", "{Model} not found", map[string]interface{}{
            "id": id,
        })
        return
    }

    common.Success(c, model, &common.ResponseMeta{
        TraceID: common.GetTraceID(c),
    })
}

// Update{Model} godoc
// @Summary Update {model}
// @Description Update an existing {model}
// @Tags {models}
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "{Model} ID"
// @Param {model} body {Model}UpdateRequest true "{Model} data"
// @Success 200 {object} common.Response{data={Model}}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/{models}/{id} [put]
func (h *{Model}Handler) Update{Model}(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        common.BadRequest(c, "MISSING_ID", "{Model} ID is required", nil)
        return
    }

    var req {Model}UpdateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
            "error": err.Error(),
        })
        return
    }

    // Get existing model
    model, err := h.service.Get{Model}ByID(c.Request.Context(), id)
    if err != nil {
        common.NotFound(c, "NOT_FOUND", "{Model} not found", nil)
        return
    }

    // Apply updates
    req.ApplyTo(model)

    // Get user ID from context
    if userID, exists := c.Get("subjectID"); exists {
        model.SetUpdatedBy(userID.(string))
    }

    updated, err := h.service.Update{Model}(c.Request.Context(), model)
    if err != nil {
        common.InternalServerError(c, "UPDATE_FAILED", "Failed to update {model}", map[string]interface{}{
            "error": err.Error(),
        })
        return
    }

    common.Success(c, updated, &common.ResponseMeta{
        TraceID: common.GetTraceID(c),
    })
}

// Delete{Model} godoc
// @Summary Delete {model}
// @Description Soft delete a {model}
// @Tags {models}
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "{Model} ID"
// @Success 204 "No content"
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/{models}/{id} [delete]
func (h *{Model}Handler) Delete{Model}(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        common.BadRequest(c, "MISSING_ID", "{Model} ID is required", nil)
        return
    }

    if err := h.service.Delete{Model}(c.Request.Context(), id); err != nil {
        common.InternalServerError(c, "DELETE_FAILED", "Failed to delete {model}", map[string]interface{}{
            "error": err.Error(),
        })
        return
    }

    c.Status(204)
}

// List{Model}s godoc
// @Summary List {models}
// @Description Retrieve a list of {models} with pagination
// @Tags {models}
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} common.Response{data=[]{Model},meta=common.ResponseMeta{pagination=common.PaginationMeta}}
// @Router /api/v1/{models} [get]
func (h *{Model}Handler) List{Model}s(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 20
    }

    offset := (page - 1) * limit

    models, total, err := h.service.List{Model}s(c.Request.Context(), limit, offset)
    if err != nil {
        common.InternalServerError(c, "LIST_FAILED", "Failed to list {models}", map[string]interface{}{
            "error": err.Error(),
        })
        return
    }

    common.Success(c, models, &common.ResponseMeta{
        TraceID: common.GetTraceID(c),
        Pagination: &common.PaginationMeta{
            Page:    page,
            Limit:   limit,
            Total:   total,
            HasNext: len(models) == limit,
        },
    })
}
```

## Model-Specific Implementation Notes

### 1. User Repository

Already implemented: `/Users/kaushik/kisanlink-ecom/internal/repositories/user/user_repository.go`

### 2. UserRole (Role Assignment)

**Key Fields:**

- `UserId` (reference to AAA user)
- `RoleId` (reference to e-commerce role)
- `AssignedBy`, `AssignedAt`, `ExpiresAt`

**Business Rules:**

- Validate role exists before assignment
- Check expiration on role validity checks
- Only admins can assign roles
- Track who assigned the role

**Special Methods:**

- `GetUserRoles(ctx, userID)` - Get all roles for a user
- `GetRoleUsers(ctx, roleID)` - Get all users with a role
- `IsRoleExpired()` - Check if role has expired
- `IsRoleValid()` - Check if active and not expired

### 3. OrganizationRole (Org-specific Role Configuration)

**Key Fields:**

- `OrganizationID`, `AAARoleID`
- `CustomPermissions` (JSON), `MaxUsers`
- `IsDefault` (for new org users)

**Business Rules:**

- Only one default role per organization
- Validate custom permissions structure
- Check max users limit on assignment

**Special Methods:**

- `GetOrgRoles(ctx, orgID)` - Get all roles for org
- `GetDefaultRole(ctx, orgID)` - Get default role for new users
- `SetAsDefault(ctx, orgID, roleID)` - Set role as default (unset others)

### 3. EcommerceRole (E-commerce Specific Roles)

**Key Fields:**

- `AAARoleID` (primary key - same as AAA role)
- Permission booleans: `CanManageCatalog`, `CanManageOrders`, etc.
- `OrganizationID` (optional org scope)

**Business Rules:**

- Sync with AAA service roles
- Permission checks before operations
- Org-scoped roles vs platform-wide roles

**Special Methods:**

- `HasPermission(permission)` - Check specific permission
- `SyncFromAAA(ctx, aaaRoleID)` - Sync role from AAA service

### 5. TaxExemption

**Key Fields:**

- `ExemptionID`, `ExemptionType`, `ExemptionRate`
- `EntityType`, `Category`, `HSNCode`
- `ValidFrom`, `ValidTo`

**Business Rules:**

- Validate exemption rate (0-100%)
- Check validity dates
- Validate HSN code format
- Only applies to specified entity types/categories

**Special Methods:**

- `IsValid()` - Check if currently valid
- `AppliesTo(entityType, category, hsnCode)` - Check applicability
- `CalculateExemption(taxAmount)` - Calculate exempted amount

### 6. ServiceSLA

**Key Fields:**

- `CatalogItemID` (links to service)
- `Type` (response/resolution/availability/performance)
- `TargetValue`, `Unit`, Thresholds
- `BusinessHoursStart`, `BusinessHoursEnd`, `BusinessDays`

**Business Rules:**

- SLA only applies to services (not products/labour)
- Target value must be positive
- Warning threshold < Critical threshold
- Valid business hours format

**Special Methods:**

- `IsResponseTimeSLA()`, `IsAvailabilitySLA()`
- `IsViolated(actualValue)` - Check if SLA violated
- `GetTargetValueWithUnit()` - Formatted target

### 7. DiscountRule

**Key Fields:**

- `RuleType` (stackable/exclusive/conditional)
- `RuleData` (JSON), `Conditions` (JSON)
- `Priority` (higher first)

**Business Rules:**

- Validate rule data structure
- Check rule conflicts (exclusive vs stackable)
- Apply in priority order
- Validate conditions format

**Special Methods:**

- `IsApplicable(ctx)` - Check if rule applies
- `ApplyRule(discount, order)` - Apply discount rule
- `ConflictsWith(otherRule)` - Check rule conflicts

### 8. OrganizationCollaborator

**Key Fields:**

- `AAAEntityID`, `AAAEntityType` (user/organization)
- `ContextOrganizationID` (org they're collaborating in)
- `Role`, `Status`, `AccessLevel`
- `Department`, `JobTitle`, `ContractType`

**Business Rules:**

- Validate entity exists in AAA
- Role hierarchy for permissions
- Access level 1-10 validation
- Contract type validation

**Special Methods:**

- `IsUserCollaborator()`, `IsOrganizationCollaborator()`
- `CanManageUsers()`, `CanAccessFinancials()`
- `HasHigherRoleThan(other)` - Role comparison
- `Activate()`, `Deactivate()`, `Suspend()`

## Testing Requirements

For each implementation, create tests in `/tests/{domain}/{model}_test.go`:

### Repository Tests

- Create, Read, Update, Delete operations
- Soft delete filtering (include_deleted context)
- List with pagination
- Filter operations
- Error handling

### Service Tests

- Business logic validation
- Error cases
- Edge cases
- Authorization checks
- Mock repository

### Handler Tests

- HTTP request/response
- Status codes
- Request validation
- Error responses
- Authentication required
- Mock service

## API Routes Registration

Add routes in the main router setup:

```go
// Role management
roleRepo := roles.NewUserRoleRepository(dbManager)
roleService := roles.NewUserRoleService(roleRepo)
roleHandler := roles.NewUserRoleHandler(roleService)

roleRoutes := api.Group("/user-roles")
{
    roleRoutes.POST("", middleware.Auth(), roleHandler.CreateUserRole)
    roleRoutes.GET("/:id", roleHandler.GetUserRole)
    roleRoutes.PUT("/:id", middleware.Auth(), roleHandler.UpdateUserRole)
    roleRoutes.DELETE("/:id", middleware.Auth(), roleHandler.DeleteUserRole)
    roleRoutes.GET("", roleHandler.ListUserRoles)
}
```

## Implementation Order

1. **Phase 1: Repositories** (Can be implemented in parallel)
   - UserRole repository
   - OrganizationRole repository
   - EcommerceRole repository

2. **Phase 2: Services** (After repositories)
   - UserRole service
   - OrganizationRole service
   - EcommerceRole service

3. **Phase 3: Handlers** (After services)
   - UserRole handler
   - OrganizationRole handler
   - EcommerceRole handler

4. **Phase 4: Remaining Models** (Repeat pattern)
   - TaxExemption (repository → service → handler)
   - ServiceSLA (repository → service → handler)
   - DiscountRule (repository → service → handler)
   - OrganizationCollaborator (repository → service → handler)

5. **Phase 5: Testing & Documentation**
   - Write tests for all layers
   - Complete API documentation
   - Integration testing

## Success Criteria

- All 9 models have complete CRUD operations
- Soft delete filtering works correctly
- All business rules validated
- Comprehensive tests (90% coverage)
- API documentation complete
- All endpoints return consistent responses
- Authentication/authorization properly implemented
- All precommit hooks pass

## Next Steps

1. Start with UserRole implementation (most critical for role management)
2. Use the implemented User repository as a reference
3. Follow the pattern strictly for consistency
4. Test each layer before moving to the next
5. Create meaningful commits for each completed model
