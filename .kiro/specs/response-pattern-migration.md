# Response Pattern Migration Specification

## Objective

Migrate all handlers from legacy `utils.APIResponse` / `common.APIResponse` pattern to modern `common.Response` pattern for API consistency.

## Current State

**Two Incompatible Response Patterns Exist:**

### Legacy Pattern (216 occurrences across 8 files)
```go
// entities/models/common/common.go
type APIResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    Error   *APIError   `json:"error,omitempty"`
}
```

**Example Response:**
```json
{
  "success": true,
  "message": "Listing created successfully",
  "data": {...}
}
```

### Modern Pattern (553 occurrences - STANDARD)
```go
// internal/common/response.go
type Response struct {
    Data  interface{}    `json:"data,omitempty"`
    Meta  *ResponseMeta  `json:"meta,omitempty"`
    Error *ResponseError `json:"error,omitempty"`
}
```

**Example Response:**
```json
{
  "data": {...},
  "meta": {
    "trace_id": "abc123",
    "pagination": {...}
  }
}
```

## Files Requiring Migration

| File | Occurrences | Priority |
|------|-------------|----------|
| internal/handlers/marketplace/listing_handler.go | 42 | HIGH |
| internal/handlers/marketplace/bidding_handler.go | 46 | HIGH |
| internal/handlers/marketplace/admin_handler.go | 25 | HIGH |
| internal/handlers/marketplace/notification_handler.go | 10 | MEDIUM |
| internal/handlers/inventory/inventory_handler.go | 43 | HIGH |
| internal/handlers/inventory/alert_handler.go | 26 | MEDIUM |
| internal/handlers/products.go | 12 | MEDIUM |
| internal/handlers/auth.go | 12 | MEDIUM |

**Total:** 216 occurrences

## Migration Strategy

### Step 1: Import Replacement
```go
// OLD
import "kisanlink-ecom/entities/models/common"
// or
import "kisanlink-ecom/internal/utils"

// NEW
import "kisanlink-ecom/internal/common"
```

### Step 2: Success Response Migration

**OLD:**
```go
utils.SuccessResponse(c, http.StatusOK, "SUCCESS", "Operation successful", data)
// or
c.JSON(http.StatusOK, common.APIResponse{
    Success: true,
    Message: "Operation successful",
    Data: data,
})
```

**NEW:**
```go
common.Success(c, data, nil)
// or with metadata
common.Success(c, data, &common.ResponseMeta{
    TraceID: common.GetTraceID(c),
})
```

### Step 3: Created Response Migration

**OLD:**
```go
utils.SuccessResponse(c, http.StatusCreated, "CREATED", "Resource created", data)
```

**NEW:**
```go
common.Created(c, data, nil)
```

### Step 4: Error Response Migration

**OLD:**
```go
utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_INPUT", "Invalid input", details)
// or
c.JSON(http.StatusBadRequest, common.APIResponse{
    Success: false,
    Error: &common.APIError{
        Code: "INVALID_INPUT",
        Message: "Invalid input",
        Details: details,
    },
})
```

**NEW:**
```go
common.BadRequest(c, "INVALID_INPUT", "Invalid input", detailsMap)
```

### Step 5: Status-Specific Error Migrations

| Old | New |
|-----|-----|
| `utils.ErrorResponse(c, 400, ...)` | `common.BadRequest(c, ...)` |
| `utils.ErrorResponse(c, 401, ...)` | `common.Unauthorized(c, ...)` |
| `utils.ErrorResponse(c, 403, ...)` | `common.Forbidden(c, ...)` |
| `utils.ErrorResponse(c, 404, ...)` | `common.NotFound(c, ...)` |
| `utils.ErrorResponse(c, 500, ...)` | `common.InternalServerError(c, ...)` |

### Step 6: Pagination Response Migration

**OLD:**
```go
utils.SuccessResponse(c, http.StatusOK, "SUCCESS", "Listings retrieved", gin.H{
    "listings": listings,
    "pagination": gin.H{
        "page": page,
        "limit": limit,
        "total": total,
    },
})
```

**NEW:**
```go
common.Success(c, gin.H{
    "listings": listings,
}, &common.ResponseMeta{
    Pagination: common.NewPaginationMeta(page, limit, total),
})
```

## Common Patterns to Look For

### Pattern 1: Direct JSON Response
```go
// OLD
c.JSON(http.StatusOK, common.APIResponse{Success: true, Data: data})

// NEW
common.Success(c, data, nil)
```

### Pattern 2: utils.SuccessResponse
```go
// OLD
utils.SuccessResponse(c, http.StatusOK, "CODE", "message", data)

// NEW
common.Success(c, data, nil)
```

### Pattern 3: utils.ErrorResponse
```go
// OLD
utils.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Resource not found", "")

// NEW
common.NotFound(c, "NOT_FOUND", "Resource not found", nil)
```

### Pattern 4: Error with Details
```go
// OLD
utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", validationErrors)

// NEW
common.BadRequest(c, "VALIDATION_ERROR", "Validation failed", map[string]interface{}{
    "errors": validationErrors,
})
```

## Implementation Checklist

For each file:

- [ ] Replace imports
- [ ] Search for `utils.SuccessResponse` -> Replace with `common.Success` or `common.Created`
- [ ] Search for `utils.ErrorResponse` -> Replace with appropriate `common.*` function
- [ ] Search for `common.APIResponse{` -> Replace with `common.*` function calls
- [ ] Search for `c.JSON(.*common.APIResponse` -> Replace with `common.*` function calls
- [ ] Remove unused `utils` import if no longer needed
- [ ] Verify pagination responses use `common.ResponseMeta`
- [ ] Test build: `go build ./cmd/server`
- [ ] Commit changes for that file

## Edge Cases

### 1. Custom Status Codes
If using custom status codes not covered by helper functions:
```go
common.Error(c, customStatusCode, "CODE", "message", nil)
```

### 2. Details String vs Map
Legacy uses string details, modern uses map:
```go
// OLD
Details: "some error details"

// NEW
Details: map[string]interface{}{
    "message": "some error details",
}
```

### 3. Message Field
Legacy has top-level `message` field. In modern pattern:
- Success: Message not needed (data is self-explanatory)
- Error: Message goes in `error.message`

## Testing Strategy

After migration:

1. **Build Test**: `go build ./cmd/server`
2. **Manual API Test**: Test one endpoint from each migrated file
3. **Response Format Verification**: Check response JSON structure matches new pattern
4. **CI/CD**: Full test suite runs automatically

## Rollout Plan

### Phase 1: Marketplace Handlers (113 occurrences)
1. listing_handler.go (42)
2. bidding_handler.go (46)
3. admin_handler.go (25)

### Phase 2: Inventory Handlers (69 occurrences)
4. inventory_handler.go (43)
5. alert_handler.go (26)

### Phase 3: Remaining Handlers (34 occurrences)
6. notification_handler.go (10)
7. products.go (12)
8. auth.go (12)

### Phase 4: Cleanup
- Remove unused `utils` package response functions
- Update API documentation
- Update integration tests if needed

## Success Criteria

- [ ] All 216 occurrences migrated
- [ ] Build passes: `go build ./cmd/server`
- [ ] All handlers use `common.Response` pattern
- [ ] No references to `utils.APIResponse` or `common.APIResponse` in handlers
- [ ] API responses follow consistent structure
- [ ] Backward compatibility: Consider adding deprecation notice if needed

## Notes

- Do NOT change the underlying business logic
- Only change response formatting
- Maintain all error codes and messages
- Keep trace_id propagation working
- Preserve pagination functionality
