# KisanLink E-Commerce API - Frontend Integration Guide

**Version:** 1.0
**Base URL:** `http://localhost:8080` (Development)
**Last Updated:** October 2025

---

## Table of Contents

1. [Authentication](#authentication)
2. [Orders API](#orders-api)
3. [Catalog API](#catalog-api)
4. [Marketplace API](#marketplace-api)
5. [Error Handling](#error-handling)
6. [Common Patterns](#common-patterns)
7. [Integration Checklist](#integration-checklist)

---

## Authentication

### Overview

All protected endpoints require a JWT token in the `Authorization` header.

```
Authorization: Bearer <your_jwt_token>
```

### Token Requirements

Your JWT **MUST** contain:
- `user_id` or `sub` - User identifier
- `organization_id` - Organization identifier (REQUIRED for Orders API)
- `roles` - User roles array
- `permissions` - User permissions array

### Authentication Flow

#### 1. Register User

**NOT REQUIRED** - Use AAA service for user registration.

#### 2. Login

**Endpoint:** `POST /api/v1/auth/login`

**Request:**
```json
{
  "username": "buyer_user",
  "password": "SecurePass123"
}
```

**Success Response (200):**
```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 3600,
    "user": {
      "user_id": "usr_abc123",
      "username": "buyer_user",
      "email": "buyer@example.com",
      "organization_id": "org_xyz789",
      "organization_name": "ABC Farms Pvt Ltd",
      "roles": ["buyer"],
      "permissions": ["order:create", "order:read"]
    }
  },
  "meta": {
    "trace_id": "req_12345"
  }
}
```

**Error Response (401):**
```json
{
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "Invalid username or password",
    "details": {}
  },
  "meta": {}
}
```

#### 3. Using the Token

**All subsequent requests:**
```bash
curl -X GET https://api.example.com/api/v1/orders \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Frontend Example (JavaScript):**
```javascript
// Store token after login
localStorage.setItem('access_token', response.data.access_token);
localStorage.setItem('organization_id', response.data.user.organization_id);

// Use in subsequent requests
const config = {
  headers: {
    'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
    'Content-Type': 'application/json'
  }
};

const response = await axios.get('/api/v1/orders', config);
```

---

## Orders API

### Base Path: `/api/v1/orders`

### 1. Create Order

**Endpoint:** `POST /api/v1/orders`
**Auth Required:** Yes
**Permissions:** `order:create`

**Request Body:**
```json
{
  "buyer_organization_id": "org_xyz789",
  "seller_organization_id": "org_seller123",
  "items": [
    {
      "catalog_item_id": "prod_001",
      "catalog_item_type": "product",
      "quantity": 100,
      "unit_price": 25.50,
      "currency": "INR",
      "tax_percentage": 18,
      "discount_percentage": 5,
      "notes": "Premium quality seeds"
    },
    {
      "catalog_item_id": "serv_002",
      "catalog_item_type": "service",
      "quantity": 1,
      "unit_price": 5000.00,
      "currency": "INR",
      "tax_percentage": 18,
      "notes": "Soil testing service"
    }
  ],
  "delivery_address": {
    "street": "Plot No. 45, Kharkar Road",
    "city": "Pune",
    "state": "Maharashtra",
    "postal_code": "411014",
    "country": "India"
  },
  "billing_address": {
    "street": "Plot No. 45, Kharkar Road",
    "city": "Pune",
    "state": "Maharashtra",
    "postal_code": "411014",
    "country": "India"
  },
  "notes": "Please deliver during morning hours (6 AM - 10 AM)",
  "expected_delivery_date": "2025-11-15T00:00:00Z"
}
```

**Field Validations:**
- `buyer_organization_id`: Must match authenticated user's organization
- `catalog_item_type`: Must be one of: `product`, `service`, `labour`
- `quantity`: Must be > 0
- `unit_price`: Must be > 0
- `currency`: Currently only `INR` supported
- `tax_percentage`: 0-100
- `discount_percentage`: 0-100

**Success Response (201):**
```json
{
  "data": {
    "id": "ord_abc123def456",
    "order_number": "ORD-2025-10-001",
    "buyer_organization_id": "org_xyz789",
    "seller_organization_id": "org_seller123",
    "status": "pending",
    "items": [
      {
        "id": "item_001",
        "catalog_item_id": "prod_001",
        "catalog_item_type": "product",
        "quantity": 100,
        "unit_price": 25.50,
        "tax_amount": 382.05,
        "discount_amount": 127.50,
        "total_amount": 2804.55,
        "currency": "INR"
      },
      {
        "id": "item_002",
        "catalog_item_id": "serv_002",
        "catalog_item_type": "service",
        "quantity": 1,
        "unit_price": 5000.00,
        "tax_amount": 900.00,
        "discount_amount": 0,
        "total_amount": 5900.00,
        "currency": "INR"
      }
    ],
    "subtotal": 7550.00,
    "tax_amount": 1282.05,
    "discount_amount": 127.50,
    "total_amount": 8704.55,
    "currency": "INR",
    "delivery_address": {
      "street": "Plot No. 45, Kharkar Road",
      "city": "Pune",
      "state": "Maharashtra",
      "postal_code": "411014",
      "country": "India"
    },
    "payment_status": "pending",
    "created_at": "2025-10-22T15:30:00Z",
    "updated_at": "2025-10-22T15:30:00Z"
  },
  "meta": {
    "trace_id": "req_create_order_001"
  }
}
```

**Error Responses:**

```json
// 401 - Missing/Invalid Organization
{
  "error": {
    "code": "INVALID_ORG",
    "message": "Valid organization ID is required",
    "details": {}
  }
}

// 400 - Validation Error
{
  "error": {
    "code": "CREATE_FAILED",
    "message": "Failed to create order",
    "details": {
      "error": "user can only create orders for their own organization"
    }
  }
}

// 400 - Invalid Items
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Invalid request body",
    "details": {
      "error": "items: at least one item is required"
    }
  }
}
```

---

### 2. List Orders

**Endpoint:** `GET /api/v1/orders`
**Auth Required:** Yes
**Permissions:** `order:read`

**Query Parameters:**
```
page=1              // Page number (default: 1)
limit=20            // Items per page (default: 20, max: 100)
status=pending      // Filter by status
buyer_id=org_xyz    // Filter by buyer organization
seller_id=org_abc   // Filter by seller organization
```

**Request Example:**
```bash
GET /api/v1/orders?page=1&limit=20&status=pending
```

**Success Response (200):**
```json
{
  "data": [
    {
      "id": "ord_abc123",
      "order_number": "ORD-2025-10-001",
      "buyer_organization_id": "org_xyz789",
      "seller_organization_id": "org_seller123",
      "status": "pending",
      "total_amount": 8704.55,
      "currency": "INR",
      "items_count": 2,
      "created_at": "2025-10-22T15:30:00Z",
      "expected_delivery_date": "2025-11-15T00:00:00Z"
    },
    {
      "id": "ord_def456",
      "order_number": "ORD-2025-10-002",
      "buyer_organization_id": "org_xyz789",
      "seller_organization_id": "org_seller456",
      "status": "confirmed",
      "total_amount": 15200.00,
      "currency": "INR",
      "items_count": 3,
      "created_at": "2025-10-21T10:15:00Z",
      "expected_delivery_date": "2025-11-10T00:00:00Z"
    }
  ],
  "meta": {
    "trace_id": "req_list_orders_001",
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 45,
      "has_next": true
    }
  }
}
```

**Status Values:**
- `pending` - Order created, awaiting confirmation
- `confirmed` - Seller confirmed the order
- `processing` - Order is being prepared
- `shipped` - Order has been dispatched
- `delivered` - Order delivered successfully
- `cancelled` - Order was cancelled
- `failed` - Order processing failed

---

### 3. Get Order Details

**Endpoint:** `GET /api/v1/orders/{order_id}`
**Auth Required:** Yes
**Permissions:** `order:read`

**Request Example:**
```bash
GET /api/v1/orders/ord_abc123
```

**Success Response (200):**
```json
{
  "data": {
    "id": "ord_abc123",
    "order_number": "ORD-2025-10-001",
    "buyer_organization_id": "org_xyz789",
    "buyer_organization_name": "ABC Farms Pvt Ltd",
    "seller_organization_id": "org_seller123",
    "seller_organization_name": "XYZ Agro Suppliers",
    "status": "pending",
    "items": [
      {
        "id": "item_001",
        "catalog_item_id": "prod_001",
        "catalog_item_type": "product",
        "product_name": "Hybrid Tomato Seeds",
        "product_sku": "HTS-001",
        "quantity": 100,
        "unit": "packets",
        "unit_price": 25.50,
        "tax_percentage": 18,
        "tax_amount": 382.05,
        "discount_percentage": 5,
        "discount_amount": 127.50,
        "total_amount": 2804.55,
        "currency": "INR",
        "notes": "Premium quality seeds"
      }
    ],
    "subtotal": 2550.00,
    "tax_amount": 382.05,
    "discount_amount": 127.50,
    "total_amount": 2804.55,
    "currency": "INR",
    "delivery_address": {
      "street": "Plot No. 45, Kharkar Road",
      "city": "Pune",
      "state": "Maharashtra",
      "postal_code": "411014",
      "country": "India"
    },
    "billing_address": {
      "street": "Plot No. 45, Kharkar Road",
      "city": "Pune",
      "state": "Maharashtra",
      "postal_code": "411014",
      "country": "India"
    },
    "payment_status": "pending",
    "payment_method": null,
    "notes": "Please deliver during morning hours (6 AM - 10 AM)",
    "expected_delivery_date": "2025-11-15T00:00:00Z",
    "actual_delivery_date": null,
    "created_at": "2025-10-22T15:30:00Z",
    "updated_at": "2025-10-22T15:30:00Z",
    "created_by": "usr_abc123",
    "updated_by": "usr_abc123"
  },
  "meta": {
    "trace_id": "req_get_order_001"
  }
}
```

**Error Response (404):**
```json
{
  "error": {
    "code": "ORDER_NOT_FOUND",
    "message": "Order not found",
    "details": {
      "order_id": "ord_invalid",
      "error": "order not found"
    }
  }
}
```

---

### 4. Update Order Status

**Endpoint:** `PATCH /api/v1/orders/{order_id}/status`
**Auth Required:** Yes
**Permissions:** `order:update`

**Request Body:**
```json
{
  "status": "confirmed",
  "notes": "Order confirmed. Expected preparation time: 2 days"
}
```

**Valid Status Transitions:**
```
pending → confirmed → processing → shipped → delivered
pending → cancelled
confirmed → cancelled
processing → cancelled
```

**Success Response (200):**
```json
{
  "data": "Order status updated successfully",
  "meta": {
    "trace_id": "req_update_status_001"
  }
}
```

**Error Response (400):**
```json
{
  "error": {
    "code": "UPDATE_FAILED",
    "message": "Failed to update order status",
    "details": {
      "error": "invalid status transition from 'delivered' to 'pending'"
    }
  }
}
```

---

### 5. Cancel Order

**Endpoint:** `POST /api/v1/orders/{order_id}/cancel`
**Auth Required:** Yes
**Permissions:** `order:cancel`

**Success Response (200):**
```json
{
  "data": "Order cancelled successfully",
  "meta": {
    "trace_id": "req_cancel_order_001"
  }
}
```

**Error Response (400):**
```json
{
  "error": {
    "code": "CANCEL_FAILED",
    "message": "Failed to cancel order",
    "details": {
      "error": "cannot cancel order in 'delivered' status"
    }
  }
}
```

---

## Catalog API

### Base Path: `/api/v1/catalog`

### 1. List Products

**Endpoint:** `GET /api/v1/catalog/products`
**Auth Required:** No
**Permissions:** None

**Query Parameters:**
```
page=1              // Page number
limit=20            // Items per page
search=tomato       // Search query
category=seeds      // Filter by category
min_price=100       // Minimum price
max_price=1000      // Maximum price
```

**Success Response (200):**
```json
{
  "data": [
    {
      "id": "prod_001",
      "name": "Hybrid Tomato Seeds",
      "sku": "HTS-001",
      "description": "High-yield hybrid tomato seeds, disease-resistant",
      "category": "seeds",
      "price": 25.50,
      "currency": "INR",
      "unit": "packet",
      "stock_quantity": 500,
      "is_available": true,
      "images": [
        "https://cdn.example.com/products/hts-001-1.jpg",
        "https://cdn.example.com/products/hts-001-2.jpg"
      ],
      "specifications": {
        "variety": "Hybrid",
        "seed_count": "100 seeds per packet",
        "germination_rate": "90%"
      },
      "created_at": "2025-09-15T10:00:00Z"
    }
  ],
  "meta": {
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 156,
      "has_next": true
    }
  }
}
```

---

### 2. Get Product Details

**Endpoint:** `GET /api/v1/catalog/products/{product_id}`
**Auth Required:** No

**Success Response (200):**
```json
{
  "data": {
    "id": "prod_001",
    "name": "Hybrid Tomato Seeds",
    "sku": "HTS-001",
    "description": "High-yield hybrid tomato seeds, disease-resistant. Suitable for all seasons.",
    "category": "seeds",
    "subcategory": "vegetable-seeds",
    "price": 25.50,
    "currency": "INR",
    "unit": "packet",
    "stock_quantity": 500,
    "low_stock_threshold": 50,
    "is_available": true,
    "images": [
      {
        "url": "https://cdn.example.com/products/hts-001-1.jpg",
        "alt": "Hybrid Tomato Seeds - Main",
        "is_primary": true
      },
      {
        "url": "https://cdn.example.com/products/hts-001-2.jpg",
        "alt": "Hybrid Tomato Seeds - Pack",
        "is_primary": false
      }
    ],
    "specifications": {
      "variety": "Hybrid",
      "seed_count": "100 seeds per packet",
      "germination_rate": "90%",
      "crop_duration": "60-70 days",
      "sowing_season": "All seasons"
    },
    "seller": {
      "organization_id": "org_seller123",
      "organization_name": "XYZ Agro Suppliers",
      "rating": 4.5,
      "total_reviews": 234
    },
    "created_at": "2025-09-15T10:00:00Z",
    "updated_at": "2025-10-20T14:30:00Z"
  },
  "meta": {
    "trace_id": "req_get_product_001"
  }
}
```

---

### 3. List Services

**Endpoint:** `GET /api/v1/catalog/services`
**Auth Required:** No

**Success Response (200):**
```json
{
  "data": [
    {
      "id": "serv_001",
      "name": "Soil Testing Service",
      "description": "Comprehensive soil analysis including NPK levels, pH, organic content",
      "category": "testing",
      "price": 5000.00,
      "currency": "INR",
      "duration": "3-5 days",
      "is_available": true,
      "includes": [
        "Sample collection",
        "Laboratory testing",
        "Detailed report",
        "Recommendations"
      ]
    }
  ],
  "meta": {
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 23,
      "has_next": true
    }
  }
}
```

---

### 4. List Labour

**Endpoint:** `GET /api/v1/catalog/labour`
**Auth Required:** No

**Success Response (200):**
```json
{
  "data": [
    {
      "id": "labour_001",
      "name": "Farm Labour - Harvesting",
      "description": "Experienced farm workers for crop harvesting",
      "category": "harvesting",
      "rate": 500.00,
      "rate_unit": "per_day",
      "currency": "INR",
      "minimum_days": 3,
      "skills": ["harvesting", "sorting", "packing"],
      "is_available": true
    }
  ],
  "meta": {
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 15,
      "has_next": false
    }
  }
}
```

---

## Marketplace API

### Base Path: `/api/v1/marketplace`

### 1. Create Listing (Ask)

**Endpoint:** `POST /api/v1/marketplace/listings`
**Auth Required:** Yes
**Permissions:** `marketplace:create`

**Request Body:**
```json
{
  "type": "ask",
  "catalog_item_id": "prod_001",
  "catalog_item_type": "product",
  "quantity": 1000,
  "unit_price": 25.50,
  "currency": "INR",
  "minimum_quantity": 100,
  "location": {
    "city": "Pune",
    "state": "Maharashtra",
    "country": "India"
  },
  "expires_at": "2025-11-30T23:59:59Z",
  "notes": "Fresh stock available, bulk orders preferred"
}
```

**Success Response (201):**
```json
{
  "data": {
    "id": "listing_abc123",
    "listing_number": "ASK-2025-10-001",
    "type": "ask",
    "status": "active",
    "catalog_item_id": "prod_001",
    "quantity": 1000,
    "unit_price": 25.50,
    "currency": "INR",
    "organization_id": "org_seller123",
    "created_at": "2025-10-22T16:00:00Z",
    "expires_at": "2025-11-30T23:59:59Z"
  },
  "meta": {
    "trace_id": "req_create_listing_001"
  }
}
```

---

### 2. Create Bid

**Endpoint:** `POST /api/v1/marketplace/bids`
**Auth Required:** Yes
**Permissions:** `marketplace:bid`

**Request Body:**
```json
{
  "listing_id": "listing_abc123",
  "quantity": 500,
  "bid_price": 24.00,
  "currency": "INR",
  "notes": "Can pick up from location",
  "expires_at": "2025-11-25T23:59:59Z"
}
```

**Success Response (201):**
```json
{
  "data": {
    "id": "bid_xyz789",
    "bid_number": "BID-2025-10-001",
    "listing_id": "listing_abc123",
    "status": "pending",
    "quantity": 500,
    "bid_price": 24.00,
    "currency": "INR",
    "organization_id": "org_buyer456",
    "created_at": "2025-10-22T16:15:00Z",
    "expires_at": "2025-11-25T23:59:59Z"
  },
  "meta": {
    "trace_id": "req_create_bid_001"
  }
}
```

---

### 3. Accept Bid

**Endpoint:** `POST /api/v1/marketplace/bids/{bid_id}/accept`
**Auth Required:** Yes
**Permissions:** `marketplace:accept_bid`

**Success Response (200):**
```json
{
  "data": {
    "id": "bid_xyz789",
    "status": "accepted",
    "accepted_at": "2025-10-22T17:00:00Z",
    "can_create_order": true,
    "order_creation_expires_at": "2025-10-29T17:00:00Z"
  },
  "meta": {
    "trace_id": "req_accept_bid_001"
  }
}
```

---

### 4. Create Order from Bid

**Endpoint:** `POST /api/v1/orders/from-bid`
**Auth Required:** Yes
**Permissions:** `order:create`

**Request Body:**
```json
{
  "bid_id": "bid_xyz789",
  "delivery_address": {
    "street": "Plot No. 45, Kharkar Road",
    "city": "Pune",
    "state": "Maharashtra",
    "postal_code": "411014",
    "country": "India"
  },
  "payment_method": "UPI",
  "notes": "Order from accepted marketplace bid"
}
```

**Success Response (201):**
```json
{
  "data": {
    "order_id": "ord_bid_001",
    "order_number": "ORD-2025-10-050",
    "bid_id": "bid_xyz789",
    "listing_id": "listing_abc123",
    "total_amount": 12000.00,
    "currency": "INR",
    "status": "pending",
    "message": "Order created successfully from winning bid"
  },
  "meta": {
    "trace_id": "req_order_from_bid_001"
  }
}
```

---

## Error Handling

### Standard Error Response Format

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {
      "field": "Additional error context",
      "validation_errors": []
    }
  },
  "meta": {
    "trace_id": "req_xyz123"
  }
}
```

### Common Error Codes

| HTTP Status | Error Code | Description |
|------------|------------|-------------|
| 400 | `INVALID_REQUEST` | Request validation failed |
| 400 | `CREATE_FAILED` | Resource creation failed |
| 400 | `UPDATE_FAILED` | Resource update failed |
| 401 | `UNAUTHORIZED` | Missing or invalid authentication |
| 401 | `INVALID_ORG` | Missing or invalid organization ID |
| 401 | `INVALID_TOKEN` | Token validation failed |
| 403 | `FORBIDDEN` | Insufficient permissions |
| 404 | `NOT_FOUND` | Resource not found |
| 404 | `ORDER_NOT_FOUND` | Order not found |
| 409 | `CONFLICT` | Resource already exists |
| 500 | `INTERNAL_ERROR` | Internal server error |

### Error Handling Best Practices

**Frontend Example:**
```javascript
try {
  const response = await api.post('/api/v1/orders', orderData);
  // Handle success
  return response.data;
} catch (error) {
  if (error.response) {
    const { status, data } = error.response;

    switch (status) {
      case 401:
        if (data.error.code === 'INVALID_ORG') {
          // User not associated with organization
          showError('Please contact admin to assign organization');
          redirectToSupport();
        } else {
          // Invalid or expired token
          refreshToken();
        }
        break;

      case 400:
        // Validation errors
        showValidationErrors(data.error.details);
        break;

      case 403:
        // Insufficient permissions
        showError('You do not have permission to perform this action');
        break;

      case 404:
        // Resource not found
        showError('Order not found');
        break;

      case 500:
        // Server error
        showError('Something went wrong. Please try again later.');
        logError(data.meta.trace_id); // Log trace ID for debugging
        break;
    }
  }
}
```

---

## Common Patterns

### 1. Pagination

All list endpoints support pagination:

```javascript
const fetchOrders = async (page = 1, limit = 20) => {
  const response = await api.get('/api/v1/orders', {
    params: { page, limit }
  });

  return {
    items: response.data.data,
    pagination: response.data.meta.pagination
  };
};

// Usage
const { items, pagination } = await fetchOrders(1, 20);
console.log(`Showing ${items.length} of ${pagination.total} orders`);
console.log(`Has next page: ${pagination.has_next}`);
```

### 2. Filtering

```javascript
const fetchFilteredOrders = async (filters) => {
  const params = {
    page: filters.page || 1,
    limit: filters.limit || 20,
    ...(filters.status && { status: filters.status }),
    ...(filters.buyerId && { buyer_id: filters.buyerId }),
    ...(filters.sellerId && { seller_id: filters.sellerId })
  };

  return await api.get('/api/v1/orders', { params });
};

// Usage
const orders = await fetchFilteredOrders({
  status: 'pending',
  page: 1,
  limit: 10
});
```

### 3. Creating Resources

```javascript
const createOrder = async (orderData) => {
  // Get organization ID from stored user data
  const organizationId = localStorage.getItem('organization_id');

  const payload = {
    buyer_organization_id: organizationId,
    seller_organization_id: orderData.sellerId,
    items: orderData.items,
    delivery_address: orderData.address,
    notes: orderData.notes
  };

  try {
    const response = await api.post('/api/v1/orders', payload);
    return response.data.data; // Return order object
  } catch (error) {
    handleOrderCreationError(error);
    throw error;
  }
};
```

### 4. Updating Resources

```javascript
const updateOrderStatus = async (orderId, newStatus, notes = '') => {
  try {
    await api.patch(`/api/v1/orders/${orderId}/status`, {
      status: newStatus,
      notes: notes
    });

    return true;
  } catch (error) {
    if (error.response?.status === 400) {
      // Invalid status transition
      const errorMsg = error.response.data.error.details.error;
      showError(errorMsg);
    }
    return false;
  }
};
```

---

## Integration Checklist

### Phase 1: Authentication Setup ✓

- [ ] Implement login flow
- [ ] Store access token securely
- [ ] Store organization_id from user data
- [ ] Add Authorization header to all requests
- [ ] Implement token refresh logic
- [ ] Handle 401 errors (redirect to login)

### Phase 2: Orders Integration ✓

- [ ] Implement create order form
  - [ ] Validate buyer_organization_id matches user's org
  - [ ] Support multiple order items
  - [ ] Validate all required fields
  - [ ] Handle delivery address
- [ ] Implement orders list view
  - [ ] Support pagination
  - [ ] Add status filters
  - [ ] Show order summary cards
- [ ] Implement order details view
  - [ ] Display all order information
  - [ ] Show item details
  - [ ] Display payment status
- [ ] Implement order status updates
  - [ ] Validate status transitions
  - [ ] Add confirmation dialogs
- [ ] Implement order cancellation
  - [ ] Add cancellation reason
  - [ ] Show confirmation dialog

### Phase 3: Catalog Integration ✓

- [ ] Implement product listing
  - [ ] Support search
  - [ ] Add category filters
  - [ ] Implement pagination
- [ ] Implement product details
  - [ ] Show all specifications
  - [ ] Display images gallery
  - [ ] Show seller information
- [ ] Add to cart functionality
- [ ] Implement checkout flow

### Phase 4: Marketplace Integration ✓

- [ ] Implement listing creation (Ask)
- [ ] Implement bid creation
- [ ] Implement bid acceptance
- [ ] Implement order from bid flow

### Phase 5: Error Handling ✓

- [ ] Implement global error handler
- [ ] Add user-friendly error messages
- [ ] Log errors with trace IDs
- [ ] Handle network errors
- [ ] Add retry logic for failed requests

### Phase 6: Testing ✓

- [ ] Test with valid organization ID
- [ ] Test without organization ID (401 error)
- [ ] Test all CRUD operations
- [ ] Test pagination
- [ ] Test filters
- [ ] Test error scenarios
- [ ] Test concurrent operations
- [ ] Load testing

---

## Frontend Code Examples

### Complete Order Management Component (React)

```javascript
import { useState, useEffect } from 'react';
import axios from 'axios';

const OrderManagement = () => {
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [pagination, setPagination] = useState({
    page: 1,
    limit: 20,
    total: 0,
    hasNext: false
  });

  // API client with auth header
  const api = axios.create({
    baseURL: 'http://localhost:8080',
    headers: {
      'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
      'Content-Type': 'application/json'
    }
  });

  // Fetch orders
  const fetchOrders = async (page = 1, filters = {}) => {
    setLoading(true);
    setError(null);

    try {
      const response = await api.get('/api/v1/orders', {
        params: {
          page,
          limit: pagination.limit,
          ...filters
        }
      });

      setOrders(response.data.data);
      setPagination(response.data.meta.pagination);
    } catch (err) {
      handleError(err);
    } finally {
      setLoading(false);
    }
  };

  // Create order
  const createOrder = async (orderData) => {
    const organizationId = localStorage.getItem('organization_id');

    if (!organizationId) {
      setError('Organization ID not found. Please login again.');
      return;
    }

    try {
      const response = await api.post('/api/v1/orders', {
        buyer_organization_id: organizationId,
        ...orderData
      });

      // Refresh orders list
      await fetchOrders(pagination.page);

      return response.data.data;
    } catch (err) {
      handleError(err);
      throw err;
    }
  };

  // Update order status
  const updateStatus = async (orderId, newStatus, notes = '') => {
    try {
      await api.patch(`/api/v1/orders/${orderId}/status`, {
        status: newStatus,
        notes
      });

      // Refresh orders list
      await fetchOrders(pagination.page);
    } catch (err) {
      handleError(err);
      throw err;
    }
  };

  // Cancel order
  const cancelOrder = async (orderId) => {
    try {
      await api.post(`/api/v1/orders/${orderId}/cancel`);

      // Refresh orders list
      await fetchOrders(pagination.page);
    } catch (err) {
      handleError(err);
      throw err;
    }
  };

  // Error handler
  const handleError = (err) => {
    if (err.response) {
      const { status, data } = err.response;

      switch (status) {
        case 401:
          if (data.error.code === 'INVALID_ORG') {
            setError('Your account is not associated with an organization. Please contact support.');
          } else {
            setError('Session expired. Please login again.');
            // Redirect to login
            window.location.href = '/login';
          }
          break;
        case 400:
          setError(data.error.message);
          break;
        case 404:
          setError('Order not found');
          break;
        case 500:
          setError('Server error. Please try again later.');
          console.error('Trace ID:', data.meta?.trace_id);
          break;
        default:
          setError('An error occurred');
      }
    } else {
      setError('Network error. Please check your connection.');
    }
  };

  useEffect(() => {
    fetchOrders();
  }, []);

  return (
    <div>
      {/* Your UI here */}
      {loading && <div>Loading...</div>}
      {error && <div className="error">{error}</div>}
      {orders.map(order => (
        <OrderCard
          key={order.id}
          order={order}
          onUpdateStatus={updateStatus}
          onCancel={cancelOrder}
        />
      ))}
      <Pagination
        current={pagination.page}
        total={pagination.total}
        pageSize={pagination.limit}
        onChange={(page) => fetchOrders(page)}
      />
    </div>
  );
};

export default OrderManagement;
```

---

## Support & Troubleshooting

### Common Issues

**Issue 1: Getting 401 "INVALID_ORG" Error**

**Cause:** Your JWT token doesn't have `organization_id`

**Solution:**
1. Check your token payload:
```javascript
const token = localStorage.getItem('access_token');
const payload = JSON.parse(atob(token.split('.')[1]));
console.log('Organization ID:', payload.organization_id);
```

2. If missing, user needs to be assigned to an organization in AAA service
3. After assignment, re-login to get new token

**Issue 2: Order Creation Fails with "user can only create orders for their own organization"**

**Cause:** `buyer_organization_id` in request doesn't match user's organization

**Solution:**
```javascript
// Always use user's organization from token
const organizationId = localStorage.getItem('organization_id');
payload.buyer_organization_id = organizationId; // Don't hardcode!
```

**Issue 3: Getting 403 Forbidden**

**Cause:** User doesn't have required permissions

**Solution:** Check user roles/permissions, contact admin to grant necessary permissions

---

## API Version & Updates

**Current Version:** v1
**Stability:** Stable
**Breaking Changes:** None planned

For updates and changes, monitor the API changelog or contact the backend team.

---

**Questions?** Contact: dev-team@kisanlink.com
**API Status:** https://status.kisanlink.com
**Documentation:** https://docs.kisanlink.com
