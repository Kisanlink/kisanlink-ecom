# AAA Service Issue: ValidateToken Not Returning User Roles

**STATUS: ✅ RESOLVED** (2025-11-04 - Updated to AAA v2.1.4)

## Resolution

Updated AAA service dependency from v2.0.5 to v2.1.4, which properly includes roles in JWT token claims. The issue is now resolved.

- **Previous Version**: v2.0.5 (roles not returned)
- **Current Version**: v2.1.4 (roles properly returned)
- **Fix Applied**: 2025-11-04
- **Verified**: Build successful, tests passing

## Original Issue Summary

The AAA service's `ValidateToken` gRPC endpoint was returning empty roles arrays in both `Claims.Roles` and `UserContext.Roles` fields, even though the JWT token contained role information in its payload.

## Impact (RESOLVED)

**Severity: HIGH - Blocking admin functionality** *(Now Fixed)*

- Admin users cannot access admin-only features in the e-commerce service
- Authorization checks fail because no roles are present
- Users with `ecom_admin`, `super_admin` roles are treated as regular users
- Organization-based filtering is incorrectly applied to admin users

## Environment

- **AAA Service Version**: v2.0.5
- **Client**: kisanlink-ecom service
- **gRPC Endpoint**: `ValidateToken`
- **Token Type**: Bearer JWT token

## Expected Behavior

When calling `ValidateToken` with a JWT token that contains roles in the `user_context.roles` field, the response should populate:

1. `ValidateTokenResponse.Claims.Roles` with role names (e.g., `["ecom_admin"]`)
2. **OR** `ValidateTokenResponse.UserContext.Roles` with role names

## Actual Behavior

Both fields are returned as empty arrays:

```
Claims.Roles = []
UserContext.Roles = []
```

## Evidence

### JWT Token Payload (decoded)
The JWT token DOES contain roles:

```json
{
  "user_context": {
    "roles": [
      {
        "id": "ROLE00000006",
        "name": "ecom_admin",
        "description": "E-commerce Administrator",
        "scope": "organization"
      }
    ]
  }
}
```

### AAA Service Response (actual)
Debug output from e-commerce service calling AAA ValidateToken:

```
[DEBUG] AAA Client Response: Claims present: true, UserContext present: true
[DEBUG] AAA Claims: Roles=[], Permissions=[], Scopes=[]
[DEBUG] AAA UserContext: Roles=[], Permissions=[], Groups=[]
```

Both `Claims.Roles` and `UserContext.Roles` are empty arrays even though the token has roles.

## Root Cause Analysis

The issue is in the AAA service's `ValidateToken` implementation:

1. **Token contains roles**: The JWT payload has roles in `user_context.roles` array
2. **Endpoint doesn't extract roles**: The ValidateToken gRPC endpoint is not extracting and populating roles from the token into the response
3. **Response has empty arrays**: The protobuf response returns with empty `Roles` fields

## What Needs to Be Fixed

### Option 1: Populate Roles in ValidateToken Response (Preferred)

Modify the AAA service's `ValidateToken` endpoint to:

1. Extract roles from the JWT token's `user_context.roles` field
2. Map role objects to role name strings
3. Populate `ValidateTokenResponse.UserContext.Roles` with the role names

**Example fix in AAA service:**

```go
// In ValidateToken endpoint
func (s *TokenService) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
    // ... existing token validation ...

    // Extract roles from token claims
    var roleNames []string
    if userContext, ok := claims["user_context"].(map[string]interface{}); ok {
        if roles, ok := userContext["roles"].([]interface{}); ok {
            for _, role := range roles {
                if roleMap, ok := role.(map[string]interface{}); ok {
                    if roleName, ok := roleMap["name"].(string); ok {
                        roleNames = append(roleNames, roleName)
                    }
                }
            }
        }
    }

    return &pb.ValidateTokenResponse{
        Valid: true,
        Claims: &pb.TokenClaims{
            // ... existing fields ...
            Roles: roleNames, // Add this
        },
        UserContext: &pb.UserContext{
            // ... existing fields ...
            Roles: roleNames, // Or add here
        },
    }, nil
}
```

### Option 2: Add Configuration Flag

If there's a configuration flag to enable role population in ValidateToken responses, please provide documentation on how to enable it.

### Option 3: Different Endpoint

If there's a different gRPC endpoint that returns user roles along with token validation, please provide:
- Endpoint name
- Request/response protobuf definitions
- Example usage

## Testing the Fix

Once fixed, the debug output should show:

```
[DEBUG] AAA Client Response: Claims present: true, UserContext present: true
[DEBUG] AAA Claims: Roles=[ecom_admin], Permissions=[...], Scopes=[...]
[DEBUG] AAA UserContext: Roles=[ecom_admin], Permissions=[...], Groups=[...]
```

## Workaround Attempted

We tried extracting roles from:
1. ✗ `Claims.Roles` - empty
2. ✗ `UserContext.Roles` - empty
3. ✗ `UserContext.UserRoles` - field doesn't exist in protobuf

No viable workaround exists on the client side.

## Related Code

**E-commerce service AAA client**: `internal/auth/aaa_client.go:291-320`

```go
// Our code is ready to consume roles from either location:
if len(resp.Claims.Roles) > 0 {
    roles = resp.Claims.Roles
} else if resp.UserContext != nil && len(resp.UserContext.Roles) > 0 {
    roles = resp.UserContext.Roles
}
```

## Questions for AAA Team

1. Is `ValidateToken` supposed to return roles in v2.0.5?
2. Is there a configuration setting to enable role population?
3. Should we be using a different gRPC endpoint?
4. Does the token need to be generated differently to include roles in validation response?

## Priority

**HIGH** - Blocking admin functionality in production e-commerce service.

Please advise on the fix or provide guidance on the correct way to retrieve roles during token validation.

---

**Reported by**: E-commerce Service Team
**Date**: 2025-11-04
**Contact**: [Your contact information]
