# API Test Scenarios - Complete Testing Guide

**Purpose:** Comprehensive testing scenarios for frontend integration
**Last Updated:** October 2025

---

## 🎯 Test Scenario Categories

1. [Authentication & Authorization](#authentication--authorization)
2. [Orders CRUD Operations](#orders-crud-operations)
3. [Orders Workflow](#orders-workflow)
4. [Marketplace Integration](#marketplace-integration)
5. [Edge Cases & Error Handling](#edge-cases--error-handling)
6. [Performance Testing](#performance-testing)

---

## Authentication & Authorization

### Scenario 1.1: Successful Login

**Objective:** Verify user can login and receive valid token

**Steps:**
1. POST `/api/v1/auth/login` with valid credentials
2. Verify response contains `access_token`
3. Verify response contains user with `organization_id`
4. Store token and organization_id

**Request:**
```json
POST /api/v1/auth/login
{
  "username": "test_buyer",
  "password": "Test@1234"
}
```

**Expected Result:**
- Status: 200 OK
- Response contains: `access_token`, `refresh_token`, `user.organization_id`
- Token is valid JWT format

**Pass Criteria:**
- [ ] Status code is 200
- [ ] `access_token` is present and non-empty
- [ ] `organization_id` is present in user object
- [ ] Token can be decoded and contains `organization_id` claim

---

### Scenario 1.2: Login with Invalid Credentials

**Objective:** Verify proper error for wrong credentials

**Request:**
```json
POST /api/v1/auth/login
{
  "username": "test_buyer",
  "password": "WrongPassword"
}
```

**Expected Result:**
- Status: 401 Unauthorized
- Error code: `INVALID_CREDENTIALS`

**Pass Criteria:**
- [ ] Status code is 401
- [ ] Error message is user-friendly
- [ ] No sensitive information leaked

---

### Scenario 1.3: Access Protected Endpoint Without Token

**Objective:** Verify authentication is enforced

**Request:**
```bash
GET /api/v1/orders
# No Authorization header
```

**Expected Result:**
- Status: 401 Unauthorized
- Error code: `UNAUTHORIZED`

**Pass Criteria:**
- [ ] Status code is 401
- [ ] Clear error message about missing authentication

---

### Scenario 1.4: Access Orders Without Organization

**Objective:** Verify organization requirement is enforced

**Prerequisites:** Login with user that has NO organization_id in token

**Request:**
```bash
GET /api/v1/orders
Authorization: Bearer <token_without_org>
```

**Expected Result:**
- Status: 401 Unauthorized
- Error code: `INVALID_ORG`
- Message: "Valid organization ID is required"

**Pass Criteria:**
- [ ] Status code is 401
- [ ] Error clearly indicates organization is missing
- [ ] Helpful message for resolution

---

## Orders CRUD Operations

### Scenario 2.1: Create Order - Success

**Objective:** Successfully create a new order

**Prerequisites:**
- Valid token with organization_id
- Organization ID: `org_buyer_001`

**Request:**
```json
POST /api/v1/orders
Authorization: Bearer <valid_token>

{
  "buyer_organization_id": "org_buyer_001",
  "seller_organization_id": "org_seller_001",
  "items": [
    {
      "catalog_item_id": "prod_tomato_seeds",
      "catalog_item_type": "product",
      "quantity": 100,
      "unit_price": 25.50,
      "currency": "INR",
      "tax_percentage": 18,
      "discount_percentage": 5
    }
  ],
  "delivery_address": {
    "street": "Farm Plot 45",
    "city": "Pune",
    "state": "Maharashtra",
    "postal_code": "411014",
    "country": "India"
  },
  "notes": "Urgent delivery required"
}
```

**Expected Result:**
- Status: 201 Created
- Response contains:
  - `id` (UUID format)
  - `order_number` (e.g., "ORD-2025-10-001")
  - `status: "pending"`
  - `total_amount` (calculated correctly)
  - All items with calculated tax and discount

**Pass Criteria:**
- [ ] Status code is 201
- [ ] Order ID is generated
- [ ] Order number follows format
- [ ] Total amount = (unit_price * quantity) + tax - discount
- [ ] Status is "pending"
- [ ] Created timestamp is present

**Calculations to Verify:**
```
Subtotal = 100 * 25.50 = 2550.00
Discount = 2550.00 * 5% = 127.50
After Discount = 2550.00 - 127.50 = 2422.50
Tax = 2422.50 * 18% = 436.05
Total = 2422.50 + 436.05 = 2858.55
```

---

### Scenario 2.2: Create Order - Organization Mismatch

**Objective:** Verify user can only create orders for their organization

**Request:**
```json
POST /api/v1/orders
Authorization: Bearer <token_with_org_buyer_001>

{
  "buyer_organization_id": "org_different_org",  // Different from token
  "seller_organization_id": "org_seller_001",
  "items": [...]
}
```

**Expected Result:**
- Status: 400 Bad Request
- Error: "user can only create orders for their own organization"

**Pass Criteria:**
- [ ] Status code is 400
- [ ] Clear error message
- [ ] Order is not created

---

### Scenario 2.3: Create Order - Invalid Items

**Objective:** Verify validation of order items

**Test Cases:**

**2.3a: Empty Items Array**
```json
{
  "buyer_organization_id": "org_buyer_001",
  "seller_organization_id": "org_seller_001",
  "items": [],  // Empty
  "delivery_address": {...}
}
```
Expected: 400, Error: "at least one item is required"

**2.3b: Invalid Quantity**
```json
{
  "items": [{
    "catalog_item_id": "prod_001",
    "catalog_item_type": "product",
    "quantity": 0,  // Invalid
    "unit_price": 100
  }]
}
```
Expected: 400, Error: "quantity must be greater than 0"

**2.3c: Invalid Price**
```json
{
  "items": [{
    "catalog_item_id": "prod_001",
    "catalog_item_type": "product",
    "quantity": 10,
    "unit_price": -50  // Negative
  }]
}
```
Expected: 400, Error: "unit_price must be greater than 0"

**2.3d: Invalid Item Type**
```json
{
  "items": [{
    "catalog_item_id": "prod_001",
    "catalog_item_type": "invalid_type",  // Must be product/service/labour
    "quantity": 10,
    "unit_price": 100
  }]
}
```
Expected: 400, Error: "invalid catalog_item_type"

---

### Scenario 2.4: List Orders - With Pagination

**Objective:** Verify pagination works correctly

**Request:**
```bash
GET /api/v1/orders?page=1&limit=10
Authorization: Bearer <valid_token>
```

**Expected Result:**
- Status: 200 OK
- Response contains:
  - `data`: Array of orders (max 10 items)
  - `meta.pagination.page`: 1
  - `meta.pagination.limit`: 10
  - `meta.pagination.total`: Total count
  - `meta.pagination.has_next`: true/false

**Pass Criteria:**
- [ ] Returns max 10 items
- [ ] Pagination object is correct
- [ ] `has_next` is true if total > 10
- [ ] Can navigate to page 2

**Next Page Test:**
```bash
GET /api/v1/orders?page=2&limit=10
```
- Should return next 10 orders
- Page should be 2

---

### Scenario 2.5: List Orders - With Filters

**Objective:** Verify filtering works

**Test Cases:**

**2.5a: Filter by Status**
```bash
GET /api/v1/orders?status=pending
```
Expected: Only orders with status="pending"

**2.5b: Filter by Buyer**
```bash
GET /api/v1/orders?buyer_id=org_buyer_001
```
Expected: Only orders for this buyer

**2.5c: Multiple Filters**
```bash
GET /api/v1/orders?status=confirmed&seller_id=org_seller_001
```
Expected: Confirmed orders from specific seller

**Pass Criteria:**
- [ ] All returned orders match filter criteria
- [ ] Empty result if no matches
- [ ] Filters work independently and combined

---

### Scenario 2.6: Get Order Details

**Objective:** Verify single order retrieval

**Request:**
```bash
GET /api/v1/orders/{order_id}
Authorization: Bearer <valid_token>
```

**Expected Result:**
- Status: 200 OK
- Complete order details with all fields
- Items array with full details
- Address information
- Timestamps

**Pass Criteria:**
- [ ] All order fields present
- [ ] Items array is complete
- [ ] Amounts match creation values
- [ ] Timestamps are valid

---

### Scenario 2.7: Get Non-Existent Order

**Objective:** Verify 404 handling

**Request:**
```bash
GET /api/v1/orders/non_existent_id
Authorization: Bearer <valid_token>
```

**Expected Result:**
- Status: 404 Not Found
- Error code: `ORDER_NOT_FOUND`

**Pass Criteria:**
- [ ] Status code is 404
- [ ] Clear error message
- [ ] No sensitive data leaked

---

### Scenario 2.8: Update Order Status - Success

**Objective:** Verify status update

**Request:**
```json
PATCH /api/v1/orders/{order_id}/status
Authorization: Bearer <valid_token>

{
  "status": "confirmed",
  "notes": "Order confirmed by seller"
}
```

**Expected Result:**
- Status: 200 OK
- Success message
- Order status updated in database

**Verify:**
```bash
GET /api/v1/orders/{order_id}
```
Status should be "confirmed"

**Pass Criteria:**
- [ ] Status code is 200
- [ ] Status is updated
- [ ] Notes are saved
- [ ] Updated timestamp is current

---

### Scenario 2.9: Update Order Status - Invalid Transition

**Objective:** Verify status transition rules

**Prerequisites:** Order in "delivered" status

**Request:**
```json
PATCH /api/v1/orders/{delivered_order_id}/status
{
  "status": "pending"
}
```

**Expected Result:**
- Status: 400 Bad Request
- Error: "invalid status transition from 'delivered' to 'pending'"

**Pass Criteria:**
- [ ] Status code is 400
- [ ] Clear error about invalid transition
- [ ] Order status unchanged

**Valid Transitions to Test:**
| From | To | Expected |
|------|----|----|
| pending | confirmed | ✅ Success |
| confirmed | processing | ✅ Success |
| processing | shipped | ✅ Success |
| shipped | delivered | ✅ Success |
| pending | cancelled | ✅ Success |
| delivered | pending | ❌ Error |
| cancelled | confirmed | ❌ Error |

---

### Scenario 2.10: Cancel Order

**Objective:** Verify order cancellation

**Prerequisites:** Order in "pending" or "confirmed" status

**Request:**
```bash
POST /api/v1/orders/{order_id}/cancel
Authorization: Bearer <valid_token>
```

**Expected Result:**
- Status: 200 OK
- Success message
- Order status changed to "cancelled"

**Pass Criteria:**
- [ ] Status code is 200
- [ ] Order status is "cancelled"
- [ ] Cannot be cancelled again

---

## Orders Workflow

### Scenario 3.1: Complete Order Lifecycle

**Objective:** Test full order workflow from creation to delivery

**Steps:**

1. **Create Order**
   ```json
   POST /api/v1/orders
   {
     "buyer_organization_id": "org_buyer_001",
     "seller_organization_id": "org_seller_001",
     "items": [...]
   }
   ```
   Expected: 201, status="pending"

2. **Seller Confirms**
   ```json
   PATCH /api/v1/orders/{id}/status
   {"status": "confirmed"}
   ```
   Expected: 200, status="confirmed"

3. **Start Processing**
   ```json
   PATCH /api/v1/orders/{id}/status
   {"status": "processing"}
   ```
   Expected: 200, status="processing"

4. **Ship Order**
   ```json
   PATCH /api/v1/orders/{id}/status
   {"status": "shipped"}
   ```
   Expected: 200, status="shipped"

5. **Deliver Order**
   ```json
   PATCH /api/v1/orders/{id}/status
   {"status": "delivered"}
   ```
   Expected: 200, status="delivered"

6. **Verify Final State**
   ```bash
   GET /api/v1/orders/{id}
   ```
   Verify: status="delivered", all transitions logged

**Pass Criteria:**
- [ ] All transitions succeed
- [ ] Status progresses correctly
- [ ] Timestamps are updated
- [ ] Cannot change status after delivered

---

### Scenario 3.2: Order Cancellation Workflow

**Objective:** Test cancellation at different stages

**Test Cases:**

**3.2a: Cancel Pending Order**
- Create order (pending)
- Cancel immediately
- Expected: Success

**3.2b: Cancel Confirmed Order**
- Create order (pending)
- Confirm order
- Cancel
- Expected: Success

**3.2c: Cannot Cancel Shipped Order**
- Create order
- Progress to "shipped"
- Attempt to cancel
- Expected: 400 Error

**3.2d: Cannot Cancel Delivered Order**
- Create order
- Progress to "delivered"
- Attempt to cancel
- Expected: 400 Error

---

## Marketplace Integration

### Scenario 4.1: Create Order from Bid

**Objective:** Verify bid-to-order flow

**Prerequisites:**
- Accepted bid exists: `bid_123`
- Bid has listing_id, quantity, price

**Request:**
```json
POST /api/v1/orders/from-bid
Authorization: Bearer <valid_token>

{
  "bid_id": "bid_123",
  "delivery_address": {
    "street": "Farm Plot 45",
    "city": "Pune",
    "state": "Maharashtra",
    "postal_code": "411014",
    "country": "India"
  },
  "payment_method": "UPI",
  "notes": "Order from marketplace bid"
}
```

**Expected Result:**
- Status: 201 Created
- Order created with bid details
- Order linked to bid and listing
- Response contains:
  - `order_id`
  - `order_number`
  - `bid_id`
  - `listing_id`
  - `total_amount`

**Pass Criteria:**
- [ ] Status code is 201
- [ ] Order created successfully
- [ ] Order items match bid
- [ ] Price matches bid price
- [ ] Quantity matches bid quantity

---

### Scenario 4.2: Validate Bid Before Order

**Objective:** Verify bid validation endpoint

**Request:**
```json
POST /api/v1/orders/validate-bid
Authorization: Bearer <valid_token>

{
  "bid_id": "bid_123"
}
```

**Expected Result:**
- Status: 200 OK
- Response contains:
  - `valid`: true/false
  - `bid_id`
  - `listing_id`
  - `can_create_order`: true/false
  - Bid details

**Pass Criteria:**
- [ ] Returns bid validation status
- [ ] Shows if order can be created
- [ ] Provides all necessary details

---

## Edge Cases & Error Handling

### Scenario 5.1: Concurrent Status Updates

**Objective:** Test race condition handling

**Steps:**
1. Create order (pending)
2. Send two simultaneous status updates:
   - Request A: Update to "confirmed"
   - Request B: Update to "cancelled"
3. Verify only one succeeds
4. Verify final state is consistent

**Expected:**
- One request succeeds
- One request fails with conflict error
- Order has single, valid status

---

### Scenario 5.2: Large Order (Many Items)

**Objective:** Test with maximum items

**Request:**
```json
POST /api/v1/orders
{
  "buyer_organization_id": "org_buyer_001",
  "seller_organization_id": "org_seller_001",
  "items": [
    /* 100 items */
  ]
}
```

**Pass Criteria:**
- [ ] Request completes within 5 seconds
- [ ] All items are saved
- [ ] Total amount calculated correctly
- [ ] Response size is reasonable

---

### Scenario 5.3: Very Long String Values

**Objective:** Test input validation limits

**Test Cases:**
- `notes`: 10,000 characters
- `delivery_address.street`: 500 characters
- Invalid UTF-8 characters

**Expected:**
- Validation errors for excessive length
- Proper handling of special characters

---

### Scenario 5.4: Malformed JSON

**Objective:** Test request parsing

**Request:**
```bash
POST /api/v1/orders
Content-Type: application/json

{invalid json}
```

**Expected:**
- Status: 400 Bad Request
- Error: "invalid JSON format"

---

### Scenario 5.5: Missing Required Fields

**Objective:** Test field validation

**Request:**
```json
POST /api/v1/orders
{
  "buyer_organization_id": "org_001"
  // Missing seller_organization_id, items, etc.
}
```

**Expected:**
- Status: 400 Bad Request
- Error details list all missing fields

---

## Performance Testing

### Scenario 6.1: List Orders Performance

**Objective:** Verify response time for large datasets

**Test:**
- Database has 10,000 orders
- Request: `GET /api/v1/orders?page=1&limit=20`

**Pass Criteria:**
- [ ] Response time < 500ms
- [ ] Correct pagination
- [ ] No timeout errors

---

### Scenario 6.2: Create Order Performance

**Objective:** Measure order creation time

**Test:**
- Create order with 10 items
- Measure response time

**Pass Criteria:**
- [ ] Response time < 1000ms
- [ ] Database transaction completes
- [ ] No errors

---

### Scenario 6.3: Concurrent Order Creation

**Objective:** Test handling of simultaneous requests

**Test:**
- Create 10 orders simultaneously
- All with different data

**Pass Criteria:**
- [ ] All orders created successfully
- [ ] All have unique order numbers
- [ ] No race conditions
- [ ] Response time < 2000ms per request

---

## Test Summary Template

```
Test Run: [Date]
Tester: [Name]
Environment: [Dev/Staging/Prod]

Authentication Tests: ✅ / ❌
- Scenario 1.1: ✅
- Scenario 1.2: ✅
- ...

Orders CRUD Tests: ✅ / ❌
- Scenario 2.1: ✅
- Scenario 2.2: ✅
- ...

Edge Cases: ✅ / ❌
- Scenario 5.1: ✅
- ...

Performance Tests: ✅ / ❌
- Scenario 6.1: ✅ (450ms avg)
- ...

Overall Pass Rate: 95% (38/40 scenarios)

Issues Found:
1. [Issue description]
2. ...

Recommendations:
1. [Recommendation]
2. ...
```

---

## Automation Script

```bash
#!/bin/bash
# test-all-scenarios.sh

BASE_URL="http://localhost:8080"
TOKEN=""

# Login and get token
echo "Running Scenario 1.1: Login..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"test_user","password":"Test@1234"}')

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.access_token')
ORG_ID=$(echo "$LOGIN_RESPONSE" | jq -r '.data.user.organization_id')

if [ -z "$TOKEN" ]; then
  echo "❌ Scenario 1.1 FAILED"
  exit 1
fi
echo "✅ Scenario 1.1 PASSED"

# Test create order
echo "Running Scenario 2.1: Create Order..."
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/orders" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"buyer_organization_id\": \"$ORG_ID\",
    \"seller_organization_id\": \"org_seller_001\",
    \"items\": [{
      \"catalog_item_id\": \"prod_001\",
      \"catalog_item_type\": \"product\",
      \"quantity\": 10,
      \"unit_price\": 100,
      \"currency\": \"INR\"
    }],
    \"delivery_address\": {
      \"street\": \"Test St\",
      \"city\": \"Pune\",
      \"state\": \"MH\",
      \"postal_code\": \"411001\",
      \"country\": \"India\"
    }
  }")

ORDER_ID=$(echo "$CREATE_RESPONSE" | jq -r '.data.id')

if [ -z "$ORDER_ID" ]; then
  echo "❌ Scenario 2.1 FAILED"
  echo "$CREATE_RESPONSE" | jq
else
  echo "✅ Scenario 2.1 PASSED - Order: $ORDER_ID"
fi

# Continue with more tests...
```

---

This comprehensive test guide ensures smooth frontend integration and catches issues early!
