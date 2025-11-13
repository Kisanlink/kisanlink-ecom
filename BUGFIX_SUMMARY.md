# Critical Bug Fix: IsAdmin Context Key Mismatch

## Summary
Fixed a critical bug where the `IsAdmin()` function was always returning `false` for admin users due to a context key mismatch between the middleware and helper function.

## Root Cause
**Key Mismatch Between Middleware and Helper Function:**
- **Middleware** (`production_auth.go:335`): Sets `c.Set("user_roles", userContext.Roles)` - snake_case
- **IsAdmin Helper** (`response.go:170`): Looked for `c.Get("userRoles")` - camelCase
- **Result**: `IsAdmin()` always returned `false` because it couldn't find the roles

## Impact
- Admin users were having their order queries incorrectly filtered by `buyer_organization_id`
- Admins could not see ALL orders as intended
- Admin access control was broken across the application

## Fix Applied
**File Modified**: `/Users/kaushik/kisanlink-ecom/internal/common/response.go`

### Updated GetUserRoles Function
```go
// GetUserRoles retrieves the user roles from the gin context
func GetUserRoles(c *gin.Context) ([]string, bool) {
    // Try snake_case first (primary key set by middleware)
    if roles, exists := c.Get("user_roles"); exists {
        if roleList, ok := roles.([]string); ok {
            return roleList, true
        }
    }

    // Try camelCase as fallback for backward compatibility
    if roles, exists := c.Get("userRoles"); exists {
        if roleList, ok := roles.([]string); ok {
            return roleList, true
        }
    }

    return nil, false
}
```

### Key Changes
1. **Primary Lookup**: Now checks `user_roles` (snake_case) first - matches middleware
2. **Fallback**: Still checks `userRoles` (camelCase) for backward compatibility
3. **Priority**: snake_case takes precedence when both exist

## Testing
**New Test File**: `/Users/kaushik/kisanlink-ecom/tests/common/response_test.go`

### Test Coverage
- ✅ GetUserRoles works with snake_case key (primary)
- ✅ GetUserRoles works with camelCase key (fallback)
- ✅ snake_case takes priority when both are present
- ✅ IsAdmin correctly identifies `admin` role
- ✅ IsAdmin correctly identifies `super_admin` role
- ✅ IsAdmin works with both snake_case and camelCase keys
- ✅ IsAdmin returns false for non-admin roles
- ✅ Handles missing or invalid role data gracefully

### Test Results
All tests pass:
```
=== RUN   TestGetUserRoles
=== RUN   TestGetUserRoles/snake_case_key_(primary)
=== RUN   TestGetUserRoles/camelCase_key_(fallback)
=== RUN   TestGetUserRoles/snake_case_takes_priority_over_camelCase
=== RUN   TestGetUserRoles/no_roles_set
=== RUN   TestGetUserRoles/invalid_type_stored
--- PASS: TestGetUserRoles (0.00s)

=== RUN   TestIsAdmin
=== RUN   TestIsAdmin/admin_role_with_snake_case_key
=== RUN   TestIsAdmin/super_admin_role_with_snake_case_key
=== RUN   TestIsAdmin/admin_role_with_camelCase_key
=== RUN   TestIsAdmin/super_admin_role_with_camelCase_key
=== RUN   TestIsAdmin/multiple_roles_including_admin
=== RUN   TestIsAdmin/non-admin_roles
=== RUN   TestIsAdmin/no_roles
=== RUN   TestIsAdmin/roles_not_set
--- PASS: TestIsAdmin (0.00s)
```

## Expected Behavior After Fix

### Test Case 1: Admin Listing ALL Orders
**Before Fix:**
```sql
SELECT * FROM "orders" WHERE buyer_organization_id = 'ORGN00000001' LIMIT 20
```
❌ Admin queries incorrectly filtered by organization

**After Fix:**
```sql
SELECT * FROM "orders" LIMIT 20
```
✅ Admin can see ALL orders without organization filter

### Test Case 2: Non-Admin User Listing Orders
**Before Fix:**
```sql
SELECT * FROM "orders" WHERE buyer_organization_id = 'ORGN00000002' LIMIT 20
```
✅ Already working correctly

**After Fix:**
```sql
SELECT * FROM "orders" WHERE buyer_organization_id = 'ORGN00000002' LIMIT 20
```
✅ Still works correctly (no change for non-admin users)

### Test Case 3: Admin Filtering by Specific Organization
**Request:**
```
GET /api/v1/orders?buyer_id=ORGN00000002
Authorization: Bearer <admin_token>
```

**SQL:**
```sql
SELECT * FROM "orders" WHERE buyer_organization_id = 'ORGN00000002' LIMIT 20
```
✅ Admin can optionally filter by organization

## Verification Steps

1. **Build the application:**
   ```bash
   go build -o bin/server ./cmd/server/main.go
   ```

2. **Run unit tests:**
   ```bash
   go test ./tests/common/... -v
   ```

3. **Test with admin token:**
   ```bash
   curl -X GET http://localhost:8080/api/v1/orders \
     -H "Authorization: Bearer <admin_token>"
   ```
   - Should see ALL orders regardless of organization
   - SQL log should NOT show `WHERE buyer_organization_id = ...`

4. **Test with non-admin token:**
   ```bash
   curl -X GET http://localhost:8080/api/v1/orders \
     -H "Authorization: Bearer <fpo_user_token>"
   ```
   - Should only see orders from user's organization
   - SQL log should show `WHERE buyer_organization_id = '...'`

## Security Considerations
- ✅ Non-admin users CANNOT bypass organization filters
- ✅ Admin detection is based on JWT roles (server-side validation)
- ✅ No client-side input can override admin detection
- ✅ Backward compatible - no breaking changes
- ✅ All existing tests pass

## Files Changed
1. `/Users/kaushik/kisanlink-ecom/internal/common/response.go` - Fixed GetUserRoles function
2. `/Users/kaushik/kisanlink-ecom/tests/common/response_test.go` - Added comprehensive tests

## Commit Information
- **Commit Hash**: `1a71bdae613c5e6b6fc244110303c487fb02f67b`
- **Author**: Kaushik <kaushikkomanduri@gmail.com>
- **Branch**: `feature/bid-ask-marketplace`
- **Message**: fix: correct user_roles context key lookup in IsAdmin function

## Next Steps
1. ✅ Fix applied and committed
2. ✅ Tests passing
3. ✅ Build successful
4. 🔲 Deploy to staging environment
5. 🔲 Verify admin can see ALL orders in staging
6. 🔲 Verify non-admin users still see only their org's orders
7. 🔲 Deploy to production

## Related Files
- `/Users/kaushik/kisanlink-ecom/internal/middleware/production_auth.go` - Sets context keys
- `/Users/kaushik/kisanlink-ecom/internal/handlers/orders/order_handler.go` - Uses IsAdmin
- `/Users/kaushik/kisanlink-ecom/internal/services/orders/order_service.go` - Applies filters
- `/Users/kaushik/kisanlink-ecom/test_admin_order_filtering.md` - Original bug report
