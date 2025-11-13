# API Documentation Guide

## Overview

This directory contains comprehensive API documentation for the KisanLink E-commerce platform. The documentation follows OpenAPI 3.0.3 specification and provides detailed information about all available endpoints, request/response schemas, authentication requirements, and business logic.

## Documentation Files

### 1. `openapi.yaml`

Main catalog and marketplace API documentation including:

- Catalog management (Products, Services, Labour, Contracts)
- Order management
- Inventory management
- Marketplace operations (Listings, Bidding)
- Integration endpoints

### 2. `openapi-new-modules.yaml`

Comprehensive documentation for user management and business logic modules:

- **User Management**: User CRUD operations and authentication
- **Role Management**:
  - User Roles: User-to-role assignments
  - Organization Roles: Organization-specific role configurations
  - E-commerce Roles: E-commerce permission management
- **Tax Exemptions**: Tax exemption management and calculations
- **Service SLAs**: Service Level Agreement definitions and tracking
- **Discount Rules**: Discount rule management and evaluation
- **Collaborators**: Organization collaborator management (fully implemented)

### 3. `swagger.yaml` / `swagger.json`

Legacy Swagger documentation (auto-generated from code annotations)

## Viewing the Documentation

### Option 1: Online API Reference (Recommended)

The easiest way to view and interact with the API documentation:

1. **Start the server**:

   ```bash
   cd /Users/kaushik/kisanlink-ecom
   go run cmd/server/main.go
   ```

2. **Access the documentation**:
   - Open your browser and navigate to: http://localhost:8080/docs
   - The Scalar API Reference provides an interactive interface
   - Features:
     - Search across all endpoints
     - Try API calls directly from the browser
     - View request/response examples
     - Download OpenAPI specs

### Option 2: Swagger UI

For traditional Swagger UI experience:

1. **Install Swagger UI**:

   ```bash
   npm install -g swagger-ui-watcher
   ```

2. **Run Swagger UI**:

   ```bash
   swagger-ui-watcher docs/openapi-new-modules.yaml
   ```

3. **Access**: Open browser to http://localhost:3000

### Option 3: VS Code Extension

For development workflow integration:

1. Install "OpenAPI (Swagger) Editor" extension in VS Code
2. Open `openapi-new-modules.yaml`
3. Right-click and select "Preview Swagger"

### Option 4: Postman

Import the OpenAPI specification:

1. Open Postman
2. Click "Import" → "Link"
3. Use: http://localhost:8080/docs/swagger.json
4. Or directly import `openapi-new-modules.yaml`

## Implementation Status

### Fully Implemented

✅ **Collaborator Management** (`/api/v1/collaborators/*`)

- All CRUD operations
- Search and filtering
- Bulk operations
- Status management
- Onboarding workflows
- Profile management

✅ **Catalog Management** (`/api/v1/catalog/*`)

- Product, Service, Labour management
- Publish/Unpublish operations
- Bulk operations
- Search and filtering

✅ **Order Management** (`/api/v1/orders/*`)

- Order lifecycle management
- Payment processing
- Status tracking

✅ **Marketplace** (`/api/v1/marketplace/*`)

- Listing management
- Bidding operations
- Notification system

### Pending Implementation (Returns 501)

⚠️ The following endpoints are defined but not yet implemented:

1. **User Management** (`/api/v1/users/*`)
   - User CRUD operations
   - Activation/deactivation
   - User lookup by username/email

2. **User Roles** (`/api/v1/user-roles/*`)
   - Role assignment
   - Role revocation
   - User role queries

3. **Organization Roles** (`/api/v1/organization-roles/*`)
   - Organization role configuration
   - Default role management
   - Custom permissions

4. **E-commerce Roles** (`/api/v1/ecommerce-roles/*`)
   - Role definition
   - Permission management
   - Permission checking

5. **Tax Exemptions** (`/api/v1/tax-exemptions/*`)
   - Exemption management
   - Tax calculation
   - Validity checking

6. **Service SLAs** (`/api/v1/service-slas/*`)
   - SLA definition
   - SLA tracking
   - Violation detection

7. **Discount Rules** (`/api/v1/discount-rules/*`)
   - Rule management
   - Rule evaluation
   - Stackability logic

8. **Organization Collaborators** (`/api/v1/organization-collaborators/*`)
   - Different from `/api/v1/collaborators`
   - Org-specific collaborator operations
   - Invitation management

## API Structure

### Base URL

```
Production:  https://api.kisanlink.com
Staging:     https://staging-api.kisanlink.com
Development: http://localhost:8080
```

### API Versioning

All endpoints are versioned: `/api/v1/*`

### Authentication

All protected endpoints require JWT authentication:

```
Authorization: Bearer <jwt_token>
```

Tokens are obtained from the AAA (Authentication, Authorization, Accounting) service.

### Request Format

- **Content-Type**: `application/json`
- **Accept**: `application/json`

### Response Format

All responses follow a consistent structure:

**Success Response**:

```json
{
  "data": { ... },
  "meta": {
    "trace_id": "uuid",
    "timestamp": "2025-10-29T00:00:00Z",
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 100,
      "total_pages": 5,
      "has_next": true,
      "has_prev": false
    }
  }
}
```

**Error Response**:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request parameters",
    "details": {
      "field_errors": [
        {
          "field": "email",
          "message": "Invalid email format"
        }
      ]
    },
    "trace_id": "uuid",
    "timestamp": "2025-10-29T00:00:00Z"
  }
}
```

## Common Query Parameters

### Pagination

```
?page=1          # Page number (1-based)
&limit=20        # Items per page (default: 20, max: 100)
```

### Filtering

```
?is_active=true              # Filter by active status
&organization_id=org_123     # Filter by organization
&role=ADMIN                  # Filter by role
&status=ACTIVE               # Filter by status
```

### Searching

```
?search=keyword              # Full-text search
?q=search+term              # Query string
```

### Sorting

```
?sort_by=created_at         # Sort field
&sort_order=desc            # Sort direction (asc/desc)
```

## HTTP Status Codes

| Code | Meaning              | Description                                     |
| ---- | -------------------- | ----------------------------------------------- |
| 200  | OK                   | Request succeeded                               |
| 201  | Created              | Resource created successfully                   |
| 204  | No Content           | Request succeeded, no response body             |
| 400  | Bad Request          | Invalid request parameters                      |
| 401  | Unauthorized         | Missing or invalid authentication               |
| 403  | Forbidden            | Insufficient permissions                        |
| 404  | Not Found            | Resource not found                              |
| 409  | Conflict             | Resource already exists or constraint violation |
| 412  | Precondition Failed  | ETag mismatch (optimistic locking)              |
| 422  | Unprocessable Entity | Business rule validation failed                 |
| 429  | Rate Limited         | Too many requests                               |
| 500  | Internal Error       | Server error                                    |
| 501  | Not Implemented      | Endpoint not yet implemented                    |
| 503  | Service Unavailable  | Service temporarily unavailable                 |

## Business Logic Documentation

### Role Hierarchy

```
OWNER (Level 5)
  └─> ADMIN (Level 4)
      └─> MANAGER (Level 3)
          └─> EMPLOYEE (Level 2)
          └─> CONTRACTOR (Level 2)
              └─> VIEWER (Level 1)
```

### E-commerce Permissions

- `catalog.manage`: Create, edit, delete catalog items
- `orders.manage`: View and manage orders
- `inventory.manage`: Manage stock levels
- `pricing.manage`: Set prices and discounts
- `customers.manage`: View customer data
- `analytics.view`: View business analytics
- `users.manage`: Manage user access
- `settings.manage`: Change system settings

### Tax Exemption Types

- **full**: 100% tax exemption
- **partial**: Percentage-based exemption
- **threshold**: Exemption up to a certain amount

### Discount Rule Types

- **stackable**: Can be combined with other discounts
- **exclusive**: Cannot be combined with other discounts
- **conditional**: Applied when conditions are met

### SLA Types

- **response**: Response time requirements
- **resolution**: Resolution time requirements
- **availability**: Uptime percentage requirements
- **performance**: Performance metric requirements

## Testing the API

### Using cURL

```bash
# Get collaborators list
curl -X GET "http://localhost:8080/api/v1/collaborators?page=1&limit=20" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json"

# Create a collaborator
curl -X POST "http://localhost:8080/api/v1/collaborators" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "aaa_entity_id": "user_123",
    "aaa_entity_type": "USER",
    "context_organization_id": "org_456",
    "role": "MANAGER",
    "display_name": "John Doe",
    "email": "john@example.com"
  }'
```

### Using HTTPie

```bash
# List users (not implemented, returns 501)
http GET http://localhost:8080/api/v1/users \
  "Authorization: Bearer <token>"

# Get collaborator by ID
http GET http://localhost:8080/api/v1/collaborators/collab_123 \
  "Authorization: Bearer <token>"
```

## Integration Examples

### JavaScript/TypeScript

```typescript
const API_BASE_URL = "http://localhost:8080/api/v1";
const AUTH_TOKEN = "your-jwt-token";

async function getCollaborators(page = 1, limit = 20) {
  const response = await fetch(
    `${API_BASE_URL}/collaborators?page=${page}&limit=${limit}`,
    {
      headers: {
        Authorization: `Bearer ${AUTH_TOKEN}`,
        "Content-Type": "application/json",
      },
    },
  );

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error.message);
  }

  return await response.json();
}

async function createCollaborator(data) {
  const response = await fetch(`${API_BASE_URL}/collaborators`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${AUTH_TOKEN}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error.message);
  }

  return await response.json();
}
```

### Python

```python
import requests

API_BASE_URL = 'http://localhost:8080/api/v1'
AUTH_TOKEN = 'your-jwt-token'

def get_collaborators(page=1, limit=20):
    response = requests.get(
        f'{API_BASE_URL}/collaborators',
        params={'page': page, 'limit': limit},
        headers={
            'Authorization': f'Bearer {AUTH_TOKEN}',
            'Content-Type': 'application/json',
        }
    )
    response.raise_for_status()
    return response.json()

def create_collaborator(data):
    response = requests.post(
        f'{API_BASE_URL}/collaborators',
        json=data,
        headers={
            'Authorization': f'Bearer {AUTH_TOKEN}',
            'Content-Type': 'application/json',
        }
    )
    response.raise_for_status()
    return response.json()
```

## Rate Limiting

Current rate limits:

- Global: 1000 requests/minute
- Per tenant: 100 requests/minute
- Per user: 50 requests/minute

Rate limit headers:

```
X-RateLimit-Limit: 50
X-RateLimit-Remaining: 49
X-RateLimit-Reset: 1640995200
```

## Idempotency

For safe retry of operations, use the `Idempotency-Key` header:

```
Idempotency-Key: <uuid>
```

Supported on:

- POST endpoints for creating resources
- PUT endpoints for updates
- Bulk operations

## Caching and ETags

GET endpoints support conditional requests:

**Request**:

```
If-None-Match: "abc123xyz"
```

**Response** (304 Not Modified if unchanged):

```
ETag: "abc123xyz"
```

## Multi-tenancy

Organization context is determined by:

1. JWT token claims (preferred)
2. `organization_id` in request body
3. `organization_id` query parameter

All resources are isolated by organization unless explicitly shared.

## Soft Delete

Resources support soft delete:

- Deleted resources are not permanently removed
- Set `deleted_at` timestamp
- Filtered from list queries by default
- Can be restored by clearing `deleted_at`

## Troubleshooting

### Common Issues

**401 Unauthorized**

- Check if token is valid and not expired
- Ensure `Bearer` prefix in Authorization header
- Verify token has required claims

**403 Forbidden**

- Check if user has required permissions
- Verify organization context is correct
- Confirm role has necessary e-commerce permissions

**404 Not Found**

- Verify resource ID is correct
- Check if resource was soft deleted
- Confirm organization context

**501 Not Implemented**

- Endpoint is defined but not yet implemented
- Check implementation status above
- Contact development team for timeline

## Contributing

### Adding New Endpoints

1. Update the appropriate OpenAPI spec file
2. Follow existing patterns for consistency
3. Include comprehensive examples
4. Document all possible error responses
5. Update this README if needed

### Generating Client SDKs

Use OpenAPI Generator to create client SDKs:

```bash
# TypeScript/JavaScript
openapi-generator-cli generate \
  -i docs/openapi-new-modules.yaml \
  -g typescript-axios \
  -o clients/typescript

# Python
openapi-generator-cli generate \
  -i docs/openapi-new-modules.yaml \
  -g python \
  -o clients/python

# Go
openapi-generator-cli generate \
  -i docs/openapi-new-modules.yaml \
  -g go \
  -o clients/go
```

## Support

For API support and questions:

- Email: api-support@kisanlink.com
- Documentation: https://docs.kisanlink.com
- GitHub Issues: [Project Repository]

## Version History

### v1.0.0 (2025-10-29)

- Initial comprehensive documentation
- All 9 module endpoints documented
- Collaborator endpoints fully implemented
- 8 modules pending implementation

## License

Proprietary - KisanLink
