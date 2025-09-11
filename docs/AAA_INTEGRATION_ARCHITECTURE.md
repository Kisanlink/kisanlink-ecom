# AAA Service Integration Architecture

## Overview

This document describes the architecture for integrating the AAA (Authentication, Authorization, and Accounting) service with the KisanLink E-commerce service. The design follows a clear separation of concerns where AAA service manages core user identity and permissions, while the e-commerce service implements role-specific workflows and business logic.

## Architecture Principles

### 1. **Separation of Concerns**

- **AAA Service**: Manages users, role definitions, and permissions
- **E-commerce Service**: Implements role-specific workflows, CRUD operations, and business logic

### 2. **ID Consistency**

- Use AAA service user IDs as primary keys in e-commerce user tables
- Use AAA service role IDs as primary keys in e-commerce role tables
- Maintain referential integrity across services

### 3. **Data Synchronization**

- Cache essential user/role data locally for performance
- Sync with AAA service for authentication and authorization decisions
- Implement eventual consistency for non-critical data

## Data Model

### User Management

#### AAA Service (Source of Truth)

```go
// User identity and authentication
type User struct {
    ID        string
    Username  string
    Email     string
    Password  string
    Status    string
    // ... other AAA-specific fields
}
```

#### E-commerce Service (Local Reference)

```go
// Local user reference with e-commerce specific data
type User struct {
    Id        string    // Primary key - same as AAA service user ID
    Username  string    // Cached from AAA service
    Email     string    // Cached from AAA service
    FirstName string    // Cached from AAA service
    LastName  string    // Cached from AAA service
    Phone     string    // Cached from AAA service

    // E-commerce specific fields
    IsActive     bool
    LastLoginAt  *time.Time
    Preferences  string // JSON preferences
}
```

### Role Management

#### AAA Service (Source of Truth)

```go
// Core role definitions
type Role struct {
    ID          string
    Name        string
    Description string
    Scope       string // GLOBAL, ORG, GROUP
    IsActive    bool
    // ... other AAA-specific fields
}
```

#### E-commerce Service (Role Implementation)

```go
// E-commerce specific role implementation
type EcommerceRole struct {
    Id          string    // Primary key - same as AAA service role ID
    RoleName    string    // Cached from AAA service
    Description string    // Cached from AAA service

    // E-commerce specific permissions
    CanManageCatalog     bool
    CanManageOrders      bool
    CanManageInventory   bool
    CanManagePricing     bool
    CanManageCustomers   bool
    CanViewAnalytics     bool
    CanManageUsers       bool
    CanManageSettings    bool

    // Role scope in e-commerce context
    OrganizationID *string
    IsActive       bool
}
```

### Role Mappings

#### User-Role Assignment

```go
// Maps AAA users to e-commerce roles
type UserRole struct {
    Id        string    // Primary key
    UserId    string    // Reference to AAA service user ID
    RoleId    string    // Reference to AAA service role ID
    IsActive  bool

    // Assignment context
    AssignedBy    string
    AssignedAt    time.Time
    ExpiresAt     *time.Time
    Notes         string
}
```

#### Organization-Role Configuration

```go
// Organization-specific role configurations
type OrganizationRole struct {
    Id             string    // Primary key
    OrganizationID string    // Reference to organization
    RoleId         string    // Reference to AAA service role ID
    IsActive       bool

    // Organization-specific overrides
    CustomPermissions string     // JSON permissions override
    MaxUsers          *int       // Maximum users with this role
    IsDefault         bool       // Default role for new users
}
```

## Service Integration

### AAA Client Interface

```go
type AAAClient interface {
    // Token validation and permission checking
    ValidateToken(ctx context.Context, token string) (userID string, roles []string, err error)
    CheckPermission(ctx context.Context, subjectID, resourceType, resourceID, action string) (allowed bool, reason string, err error)

    // User management
    CreateUser(ctx context.Context, req interface{}) (*AAAUser, error)
    GetUser(ctx context.Context, userID string) (*AAAUser, error)
    UpdateUser(ctx context.Context, userID string, req interface{}) (*AAAUser, error)
    DeleteUser(ctx context.Context, userID string) error

    // Authentication
    AuthenticateUser(ctx context.Context, username, password string) (*AAAUser, string, error)
    RefreshToken(ctx context.Context, refreshToken string) (string, error)

    // Role management
    GetUserRoles(ctx context.Context, userID string) ([]*AAARole, error)
    AssignRole(ctx context.Context, userID, roleID string) error
    RemoveRole(ctx context.Context, userID, roleID string) error

    // Permission management
    GetUserPermissions(ctx context.Context, userID string) ([]*AAAPermission, error)

    Close() error
}
```

### User Service

The user service acts as a bridge between AAA service and e-commerce business logic:

```go
type UserService struct {
    aaaClient auth.AAAClient
    // TODO: Add user repository
}

// Key methods:
// - CreateUser: Registers with AAA service, creates local reference
// - LoginUser: Authenticates via AAA service, updates local data
// - GetUserByID: Fetches from local cache, syncs with AAA if needed
// - GetUserRoles: Retrieves roles from local role mapping tables
// - AssignRole: Creates local role assignment, validates with AAA
```

## Authentication Flow

### 1. User Registration

```
1. Client → E-commerce Service: POST /api/v1/auth/register
2. E-commerce Service → AAA Service: CreateUser gRPC call
3. AAA Service: Creates user, returns user data
4. E-commerce Service: Creates local user reference
5. Response: User created successfully
```

### 2. User Login

```
1. Client → E-commerce Service: POST /api/v1/auth/login
2. E-commerce Service → AAA Service: AuthenticateUser gRPC call
3. AAA Service: Validates credentials, returns token + user data
4. E-commerce Service: Updates local user reference, returns token
5. Response: Login successful with access token
```

### 3. Permission Checking

```
1. Client → E-commerce Service: Protected endpoint with token
2. E-commerce Service → AAA Service: ValidateToken gRPC call
3. AAA Service: Validates token, returns user ID + roles
4. E-commerce Service: Checks local role mappings for permissions
5. Response: Access granted/denied based on role permissions
```

## Authorization Workflow

### Role-Based Access Control (RBAC)

1. **User Authentication**: Token validation via AAA service
2. **Role Resolution**: Map AAA roles to e-commerce permissions
3. **Permission Checking**: Check if user's roles allow the requested action
4. **Access Decision**: Grant or deny access based on permissions

### Permission Matrix

| Permission         | Buyer Role | Seller Role | Admin Role | Manager Role |
| ------------------ | ---------- | ----------- | ---------- | ------------ |
| `catalog.view`     | ✅         | ✅          | ✅         | ✅           |
| `catalog.manage`   | ❌         | ✅          | ✅         | ✅           |
| `orders.create`    | ✅         | ❌          | ✅         | ✅           |
| `orders.manage`    | ❌         | ✅          | ✅         | ✅           |
| `inventory.manage` | ❌         | ✅          | ✅         | ❌           |
| `analytics.view`   | ❌         | ❌          | ✅         | ✅           |
| `users.manage`     | ❌         | ❌          | ✅         | ❌           |

## Implementation Status

### ✅ Completed

- [x] User model with AAA service ID as primary key
- [x] E-commerce role model with AAA service role ID as primary key
- [x] User-role mapping table
- [x] Organization-role configuration table
- [x] AAA client interface definition
- [x] User service with AAA integration methods
- [x] Basic authentication endpoints

### 🔄 In Progress

- [ ] gRPC client implementation with protobuf imports
- [ ] Database migrations for new tables
- [ ] Role-based middleware implementation
- [ ] Permission checking in business logic

### 📋 Planned

- [ ] Complete gRPC integration with AAA service
- [ ] Role seeding and permission configuration
- [ ] User role assignment workflows
- [ ] Organization role management
- [ ] Audit logging for role changes
- [ ] Performance optimization with caching

## Database Schema

### Tables

1. **users** - Local user references
2. **ecommerce_roles** - Role implementations
3. **user_roles** - User-role assignments
4. **organization_roles** - Organization-specific role configs

### Key Relationships

- `users.id` → AAA service user ID
- `ecommerce_roles.id` → AAA service role ID
- `user_roles.user_id` → `users.id`
- `user_roles.role_id` → `ecommerce_roles.id`
- `organization_roles.role_id` → `ecommerce_roles.id`

## Security Considerations

1. **Token Validation**: All requests validated via AAA service
2. **Permission Granularity**: Fine-grained permissions per resource/action
3. **Role Isolation**: Organization-scoped roles prevent cross-org access
4. **Audit Trail**: Log all role assignments and permission changes
5. **Token Expiration**: Implement proper token refresh mechanisms

## Performance Considerations

1. **Local Caching**: Cache user/role data locally for fast access
2. **Batch Operations**: Batch permission checks where possible
3. **Connection Pooling**: Maintain gRPC connection pools
4. **Async Updates**: Update local cache asynchronously
5. **Database Indexing**: Proper indexes on foreign keys and frequently queried fields

## Future Enhancements

1. **Multi-Tenancy**: Support for multiple organizations with isolated data
2. **Dynamic Permissions**: Runtime permission configuration
3. **Role Inheritance**: Hierarchical role structures
4. **Permission Delegation**: Allow users to delegate permissions
5. **Advanced Auditing**: Detailed audit trails with compliance reporting
