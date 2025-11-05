# Swagger Annotation Update Specification

## Objective

Update all Swagger annotations across handler files to reflect the new `common.Response` pattern instead of the legacy `common.APIResponse` pattern.

## Changes Required

### Pattern 1: Success Responses

**OLD:**
```go
// @Success 200 {object} common.APIResponse{data=SomeType} "Success message"
```

**NEW:**
```go
// @Success 200 {object} common.Response{data=SomeType} "Success message"
```

### Pattern 2: Error Responses

**OLD:**
```go
// @Failure 400 {object} common.APIResponse{error=common.APIError} "Error message"
```

**NEW:**
```go
// @Failure 400 {object} common.Response{error=common.ResponseError} "Error message"
```

### Pattern 3: Simple Responses (no specific type)

**OLD:**
```go
// @Success 200 {object} common.APIResponse "Success"
```

**NEW:**
```go
// @Success 200 {object} common.Response "Success"
```

## Files to Update

All handler files with swagger annotations:
- internal/handlers/marketplace/listing_handler.go
- internal/handlers/marketplace/bidding_handler.go
- internal/handlers/marketplace/admin_handler.go
- internal/handlers/marketplace/notification_handler.go
- internal/handlers/inventory/inventory_handler.go
- internal/handlers/inventory/alert_handler.go
- internal/handlers/orders/*.go
- internal/handlers/auth.go
- internal/handlers/products.go
- Any other handlers with swagger annotations

## Search and Replace Patterns

1. `common.APIResponse{data=` → `common.Response{data=`
2. `common.APIResponse{error=common.APIError}` → `common.Response{error=common.ResponseError}`
3. `common.APIResponse "` → `common.Response "`
4. `common.APIResponse}` → `common.Response}`

## Verification

After updates:
1. Run `swag init` to regenerate swagger.json
2. Check that swagger.json contains `common.Response` definitions
3. Verify no references to `common.APIResponse` in swagger.json
4. Test swagger UI to ensure proper rendering

## Notes

- Only update swagger annotations (comments)
- Do NOT change actual code
- Maintain all other annotation details (descriptions, parameters, etc.)
