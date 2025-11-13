# Manual Testing Guide - Orders API

## Prerequisites

You need a JWT token with organization information. Here's how to test:

## Step 1: Register/Login to Get Token

### Register a New User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "test_buyer",
    "email": "buyer@test.com",
    "password": "Test@1234",
    "phone_number": "+919876543210",
    "country_code": "+91"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "test_buyer",
    "password": "Test@1234"
  }'
```

**Expected Response:**
```json
{
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc...",
    "token_type": "Bearer",
    "expires_in": 3600,
    "user": {
      "user_id": "user-123",
      "username": "test_buyer",
      "email": "buyer@test.com",
      "organization_id": "org-456",
      "roles": ["buyer"]
    }
  }
}
```

**Important:** Save the `access_token` - you'll use it as `YOUR_TOKEN` below.

---

## Step 2: Verify Token Has Organization

Decode your JWT token at https://jwt.io to verify it contains:
- `organization_id` or `OrganizationID`
- `user_id` or `sub`

If your token doesn't have `organization_id`, you need to assign the user to an organization in the AAA service first.

---

## Step 3: Test Orders API

### 3.1 Create an Order

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "buyer_organization_id": "org-456",
    "seller_organization_id": "org-789",
    "items": [
      {
        "catalog_item_id": "prod-001",
        "catalog_item_type": "product",
        "quantity": 10,
        "unit_price": 250.50,
        "currency": "INR",
        "tax_percentage": 18,
        "discount_percentage": 5
      },
      {
        "catalog_item_id": "serv-001",
        "catalog_item_type": "service",
        "quantity": 2,
        "unit_price": 1500.00,
        "currency": "INR",
        "tax_percentage": 18
      }
    ],
    "delivery_address": {
      "street": "123 Farm Road",
      "city": "Bangalore",
      "state": "Karnataka",
      "postal_code": "560001",
      "country": "India"
    },
    "notes": "Deliver between 9 AM - 5 PM",
    "expected_delivery_date": "2025-11-01T00:00:00Z"
  }'
```

**Expected Success Response (201):**
```json
{
  "data": {
    "id": "ord-123",
    "order_number": "ORD-2025-001",
    "buyer_organization_id": "org-456",
    "seller_organization_id": "org-789",
    "status": "pending",
    "total_amount": 5755.95,
    "currency": "INR",
    "items": [...],
    "delivery_address": {...},
    "created_at": "2025-10-22T20:00:00Z"
  },
  "meta": {
    "trace_id": "abc123"
  }
}
```

**Common Errors:**

- **401 INVALID_ORG:** Your token doesn't have organization_id
  ```json
  {
    "error": {
      "code": "INVALID_ORG",
      "message": "Valid organization ID is required"
    }
  }
  ```

- **400 CREATE_FAILED:** buyer_organization_id doesn't match your token's org
  ```json
  {
    "error": {
      "code": "CREATE_FAILED",
      "message": "Failed to create order",
      "details": {
        "error": "user can only create orders for their own organization"
      }
    }
  }
  ```

---

### 3.2 List Orders

```bash
curl -X GET "http://localhost:8080/api/v1/orders?page=1&limit=20" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**With Filters:**
```bash
curl -X GET "http://localhost:8080/api/v1/orders?status=pending&buyer_id=org-456&page=1&limit=10" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Expected Response (200):**
```json
{
  "data": [
    {
      "id": "ord-123",
      "order_number": "ORD-2025-001",
      "status": "pending",
      "total_amount": 5755.95,
      "created_at": "2025-10-22T20:00:00Z"
    }
  ],
  "meta": {
    "trace_id": "abc123",
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 1,
      "has_next": false
    }
  }
}
```

---

### 3.3 Get Order by ID

```bash
curl -X GET http://localhost:8080/api/v1/orders/ord-123 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Expected Response (200):**
```json
{
  "data": {
    "id": "ord-123",
    "order_number": "ORD-2025-001",
    "buyer_organization_id": "org-456",
    "seller_organization_id": "org-789",
    "status": "pending",
    "items": [...],
    "total_amount": 5755.95,
    "delivery_address": {...}
  }
}
```

---

### 3.4 Update Order Status

```bash
curl -X PATCH http://localhost:8080/api/v1/orders/ord-123/status \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "status": "confirmed",
    "notes": "Order confirmed by seller"
  }'
```

**Valid Status Values:**
- `pending`
- `confirmed`
- `processing`
- `shipped`
- `delivered`
- `cancelled`
- `failed`

**Expected Response (200):**
```json
{
  "data": "Order status updated successfully",
  "meta": {
    "trace_id": "abc123"
  }
}
```

---

### 3.5 Cancel Order

```bash
curl -X POST http://localhost:8080/api/v1/orders/ord-123/cancel \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Expected Response (200):**
```json
{
  "data": "Order cancelled successfully",
  "meta": {
    "trace_id": "abc123"
  }
}
```

---

### 3.6 Update Order Details

```bash
curl -X PUT http://localhost:8080/api/v1/orders/ord-123 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "delivery_address": {
      "street": "456 New Address",
      "city": "Mumbai",
      "state": "Maharashtra",
      "postal_code": "400001",
      "country": "India"
    },
    "notes": "Updated delivery instructions",
    "expected_delivery_date": "2025-11-05T00:00:00Z"
  }'
```

---

### 3.7 Create Order from Marketplace Bid

```bash
curl -X POST http://localhost:8080/api/v1/orders/from-bid \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "bid_id": "bid-123",
    "delivery_address": {
      "street": "123 Farm Road",
      "city": "Bangalore",
      "state": "Karnataka",
      "postal_code": "560001",
      "country": "India"
    },
    "payment_method": "UPI",
    "notes": "Order from winning bid"
  }'
```

**Expected Response (201):**
```json
{
  "data": {
    "order_id": "ord-124",
    "order_number": "ORD-2025-002",
    "bid_id": "bid-123",
    "listing_id": "listing-456",
    "total_amount": 25000.00,
    "currency": "INR",
    "status": "pending",
    "message": "Order created successfully from winning bid"
  }
}
```

---

### 3.8 Validate Bid for Order

```bash
curl -X POST http://localhost:8080/api/v1/orders/validate-bid \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "bid_id": "bid-123"
  }'
```

**Expected Response (200):**
```json
{
  "data": {
    "valid": true,
    "bid_id": "bid-123",
    "listing_id": "listing-456",
    "product_id": "prod-001",
    "winning_amount": 25000.00,
    "quantity": 100,
    "currency": "INR",
    "seller_id": "user-789",
    "buyer_id": "user-456",
    "seller_org_id": "org-789",
    "buyer_org_id": "org-456",
    "can_create_order": true
  }
}
```

---

### 3.9 Process Payment for Order

```bash
curl -X POST http://localhost:8080/api/v1/orders/ord-123/payment \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "payment_method": "UPI"
  }'
```

**Payment Methods:**
- `UPI`
- `CREDIT_CARD`
- `DEBIT_CARD`
- `NET_BANKING`
- `WALLET`

**Expected Response (200):**
```json
{
  "data": {
    "payment_id": "pay-123",
    "order_id": "ord-123",
    "status": "completed",
    "amount": 5755.95,
    "currency": "INR",
    "payment_method": "UPI",
    "processed_at": "2025-10-22T20:30:00Z",
    "message": "Payment processed successfully"
  }
}
```

---

### 3.10 Get Payment Status

```bash
curl -X GET http://localhost:8080/api/v1/orders/ord-123/payment/status \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Expected Response (200):**
```json
{
  "data": {
    "payment_id": "pay-123",
    "order_id": "ord-123",
    "status": "completed",
    "amount": 5755.95,
    "currency": "INR",
    "payment_method": "UPI",
    "processed_at": "2025-10-22T20:30:00Z"
  }
}
```

---

## Testing Flow

### Complete Order Lifecycle Test

```bash
# 1. Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test_buyer","password":"Test@1234"}' \
  | jq -r '.data.access_token')

# 2. Create Order
ORDER_ID=$(curl -s -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "buyer_organization_id": "org-456",
    "seller_organization_id": "org-789",
    "items": [{
      "catalog_item_id": "prod-001",
      "catalog_item_type": "product",
      "quantity": 10,
      "unit_price": 250.50,
      "currency": "INR"
    }],
    "delivery_address": {
      "street": "123 Test St",
      "city": "Bangalore",
      "state": "Karnataka",
      "postal_code": "560001",
      "country": "India"
    }
  }' | jq -r '.data.id')

# 3. Get Order Details
curl -X GET http://localhost:8080/api/v1/orders/$ORDER_ID \
  -H "Authorization: Bearer $TOKEN" | jq

# 4. Update Status
curl -X PATCH http://localhost:8080/api/v1/orders/$ORDER_ID/status \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"status": "confirmed"}' | jq

# 5. Process Payment
curl -X POST http://localhost:8080/api/v1/orders/$ORDER_ID/payment \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"payment_method": "UPI"}' | jq

# 6. List All Orders
curl -X GET "http://localhost:8080/api/v1/orders?page=1&limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq
```

---

## Common Issues

### 1. Missing Organization in Token

**Error:**
```json
{
  "error": {
    "code": "INVALID_ORG",
    "message": "Valid organization ID is required"
  }
}
```

**Solution:** Your user needs to be assigned to an organization in the AAA service. Use the AAA service's assign organization endpoint or create a user with organization context.

---

### 2. Organization Mismatch

**Error:**
```json
{
  "error": {
    "code": "CREATE_FAILED",
    "message": "Failed to create order",
    "details": {
      "error": "user can only create orders for their own organization"
    }
  }
}
```

**Solution:** Ensure `buyer_organization_id` in the request matches your token's `organization_id`.

---

### 3. No Authorization Header

**Error:**
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Authentication required"
  }
}
```

**Solution:** Add the `Authorization: Bearer YOUR_TOKEN` header to all requests.

---

## Notes

- Replace `YOUR_TOKEN` with the actual JWT token from login response
- Replace `org-456`, `org-789`, `prod-001` with actual IDs from your system
- All timestamps should be in RFC3339 format (e.g., `2025-10-22T20:00:00Z`)
- The server is running on `http://localhost:8080` (adjust port if different)
