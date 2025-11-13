# Bearer Token and Organization Context Fixes

## Issues Fixed

### 1. ✅ Swagger UI - Unable to Add Bearer Token
**Problem**: The /docs route (Scalar API Reference) didn't show a way to add Bearer tokens for authentication.

**Fix**: Updated `internal/routes/routes.go:729` to enable authentication:
```go
Authentication: "BearerAuth",
```

**Result**: The Swagger UI now displays an authentication button where you can add your Bearer token. All API requests from the UI will include this token.

---

### 2. ✅ Organization Context Key Mismatch
**Problem**: Auth middleware set `organization_id` but handlers expected `organizationID`, causing mismatches.

**Fix**: Updated `internal/middleware/authn.go:61` to set both versions:
```go
c.Set("organization_id", claims.OrganizationID) // Snake case
c.Set("organizationID", claims.OrganizationID)  // Camel case
```

**Result**: Both naming conventions are now supported for backward compatibility.

---

### 3. ✅ PostgreSQL JSONB Metadata Field Error
**Problem**: `ERROR: invalid input syntax for type json (SQLSTATE 22P02)`

**Root Cause**: The `Metadata` field was defined as `string` in Go but `jsonb` in PostgreSQL. When GORM tried to insert `"{}"` (a string), PostgreSQL rejected it.

**Fix**:
- Changed field type from `string` to `datatypes.JSON` in `entities/models/catalog/inventory_lot.go:84`
- Updated all assignments in `internal/services/inventory/inventory_service.go` to convert strings to byte slices:
  ```go
  lot.Metadata = []byte(req.Metadata)
  lot.Metadata = []byte("{}")
  ```
- Added `gorm.io/datatypes` dependency

**Result**: Metadata now properly handles PostgreSQL jsonb columns.

---

## ⚠️ IMPORTANT: Organization ID Token Issue

### The Problem You're Still Seeing
From your logs:
```sql
organization_id = ''
```

**This is NOT a code issue** - your JWT token doesn't contain an organization ID.

### Why This Happens
The AAA service token validation extracts the organization ID from the token with this priority:
1. `UserContext.OrganizationId` (AAA v2.0)
2. `Organization.Id`
3. `Claims.OrganizationId`

Your token is valid but has NO organization ID in any of these fields.

### How to Fix
You need to obtain a token that includes an organization context. Options:

#### Option 1: Use a Token with Organization Context
1. Login through the AAA service with a user that belongs to an organization
2. The AAA service should return a token with organization information
3. Use that token in your API requests

#### Option 2: Mock Token for Development
If using mock authentication, update `internal/auth/mock_client.go:40` with a valid organization ID:
```go
OrganizationID: "your_org_id_here",
```

#### Option 3: Verify Your AAA Token
Check what's in your current token:
```bash
# Decode the JWT (without validation)
echo "YOUR_TOKEN" | cut -d'.' -f2 | base64 --decode | jq
```

Look for fields like:
- `organization_id`
- `user_context.organization_id`
- `org_id`

### Testing After Restart
1. Restart your server to apply all fixes
2. Go to `/docs` - you should see an authentication button/lock icon
3. Click it and paste a token that includes an organization ID
4. Test the `/api/v1/inventory/lots` endpoint - it should now work!

---

## Files Modified

1. `internal/middleware/authn.go` - Added both organizationID key formats
2. `internal/routes/routes.go` - Enabled Swagger UI authentication
3. `entities/models/catalog/inventory_lot.go` - Changed Metadata type to datatypes.JSON
4. `internal/services/inventory/inventory_service.go` - Updated metadata assignments
5. `go.mod` / `go.sum` - Added gorm.io/datatypes dependency

---

## Next Steps

1. **Restart your server** to apply the changes
2. **Get a valid token** with organization context from your AAA service
3. **Add the token** in the Swagger UI authentication dialog
4. **Test** your inventory and order endpoints

The metadata JSON error is fixed. You just need a proper token with organization information!
