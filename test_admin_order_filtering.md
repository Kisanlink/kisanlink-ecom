# Admin Order Filtering Test Plan

## Bug Description
Admins were seeing orders filtered by buyer_organization_id, same as regular users. They should see ALL orders.

## Fix Applied

### 1. Added Helper Functions to `/internal/common/response.go`
```go
// GetUserRoles retrieves the user roles from the gin context
func GetUserRoles(c *gin.Context) ([]string, bool)

// IsAdmin checks if the user has admin or super_admin role
func IsAdmin(c *gin.Context) bool
```

### 2. Modified `/internal/handlers/orders/order_handler.go`
- Changed to detect admin users using `common.IsAdmin(c)`
- Pass `isAdmin` flag to service layer
- Organization ID is optional for admins

### 3. Modified `/internal/services/orders/order_service.go`
- Updated `ListOrders` signature to accept `isAdmin bool` parameter
- **For Admins**: No automatic organization filtering applied
- **For Non-Admins**: Organization filtering applied as before

## Expected Behavior After Fix

### Test Case 1: Super Admin Listing Orders
**Request:**
```
GET /api/v1/orders
Authorization: Bearer <admin_token>
```

**Expected SQL:**
```sql
SELECT * FROM "orders" LIMIT 20
SELECT count(*) FROM "orders"
```
✅ No `buyer_organization_id` filter

### Test Case 2: FPO User Listing Orders
**Request:**
```
GET /api/v1/orders
Authorization: Bearer <fpo_user_token>
```

**Expected SQL:**
```sql
SELECT * FROM "orders" WHERE buyer_organization_id = 'ORGN00000002' LIMIT 20
SELECT count(*) FROM "orders" WHERE buyer_organization_id = 'ORGN00000002'
```
✅ With `buyer_organization_id` filter

### Test Case 3: Admin Filtering by Specific Organization
**Request:**
```
GET /api/v1/orders?buyer_id=ORGN00000002
Authorization: Bearer <admin_token>
```

**Expected SQL:**
```sql
SELECT * FROM "orders" WHERE buyer_organization_id = 'ORGN00000002' LIMIT 20
SELECT count(*) FROM "orders" WHERE buyer_organization_id = 'ORGN00000002'
```
✅ Admin can optionally filter by organization

### Test Case 4: Admin Viewing Single Order
**Request:**
```
GET /api/v1/orders/ORD001
Authorization: Bearer <admin_token>
```

**Expected:** Admin can view any order regardless of organization

### Test Case 5: Admin Updating Order Status
**Request:**
```
PATCH /api/v1/orders/ORD001/status
Authorization: Bearer <admin_token>
```

**Expected:** Admin can update any order regardless of organization

## Files Modified

1. `/internal/common/response.go` - Added helper functions
2. `/internal/handlers/orders/order_handler.go` - Updated ListOrders handler
3. `/internal/services/orders/order_service.go` - Updated ListOrders service method and interface

## Security Considerations

✅ **Non-admin users cannot bypass filters**: Organization filtering is enforced server-side based on JWT roles
✅ **Admin detection is secure**: Uses roles from validated JWT token (not client input)
✅ **Backward compatible**: Existing FPO functionality unchanged
✅ **No breaking changes**: Only affects order listing for admin users

## Role Values Used
- `"admin"` - Regular admin role
- `"super_admin"` - Super admin role

Both roles grant access to view all orders without organization filtering.
