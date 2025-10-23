# Orders API Testing Guide

This directory contains complete testing resources for the Orders API.

## 📁 Files

### 1. `manual-testing-orders.md`
**Complete API reference** with all endpoints, request/response examples, and error handling.
- All CRUD operations
- Marketplace integration (bid-based orders)
- Payment processing
- Step-by-step testing flows

### 2. `Orders-API.postman_collection.json`
**Postman Collection** - Import this into Postman for GUI-based testing.
- Pre-configured requests with auto-saved variables
- Automatic token extraction from login
- Automatic order ID capture from create response

**How to import:**
1. Open Postman
2. Click **Import** button
3. Select `Orders-API.postman_collection.json`
4. Start testing!

### 3. `quick-test-orders.sh`
**Automated test script** - Run complete order lifecycle test in seconds.
- Checks token for organization ID
- Creates and retrieves orders
- Updates status
- Validates responses

**Run it:**
```bash
./quick-test-orders.sh
```

---

## 🚀 Quick Start

### Option 1: Use the Automated Script

```bash
# Make sure your server is running on localhost:8080
cd docs
./quick-test-orders.sh
```

The script will:
✅ Login and get token
✅ Verify organization ID exists
✅ Create a test order
✅ Retrieve order details
✅ Update order status
✅ List all orders

### Option 2: Manual Testing with curl

```bash
# 1. Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test_buyer","password":"Test@1234"}' \
  | jq -r '.data.access_token')

# 2. Create order
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "buyer_organization_id": "your-org-id",
    "seller_organization_id": "seller-org-id",
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
  }'
```

### Option 3: Import Postman Collection

1. Import `Orders-API.postman_collection.json` into Postman
2. Update `baseUrl` variable if needed (default: `http://localhost:8080`)
3. Run **Login** request first (saves token automatically)
4. Run other requests (token is auto-applied)

---

## ⚠️ Common Issues

### 1. 401 Error: "Valid organization ID is required"

**Problem:** Your JWT token doesn't have an `organization_id` field.

**Check your token:**
```bash
# Decode your token
echo "YOUR_TOKEN" | cut -d. -f2 | base64 -d | jq
```

**Solution:** The user must be assigned to an organization in the AAA service.

### 2. 400 Error: "user can only create orders for their own organization"

**Problem:** The `buyer_organization_id` in your request doesn't match your token's `organization_id`.

**Solution:**
```bash
# Get your org ID from token
ORG_ID=$(echo "$TOKEN" | cut -d. -f2 | base64 -d | jq -r '.organization_id')

# Use it in the request
"buyer_organization_id": "$ORG_ID"
```

### 3. 401 Error: "Authentication required"

**Problem:** Missing or invalid Authorization header.

**Solution:**
```bash
# Always include the header
-H "Authorization: Bearer YOUR_TOKEN_HERE"
```

---

## 📊 Test Data Examples

### Sample Organization IDs
```
org-456  (buyer)
org-789  (seller)
```

### Sample Catalog Item IDs
```
prod-001  (product)
serv-001  (service)
labour-001  (labour)
```

### Valid Order Statuses
```
pending
confirmed
processing
shipped
delivered
cancelled
failed
```

### Valid Payment Methods
```
UPI
CREDIT_CARD
DEBIT_CARD
NET_BANKING
WALLET
```

---

## 🔄 Complete Test Flow

```bash
# 1. Start server
go run cmd/server/main.go

# 2. Run quick test
./quick-test-orders.sh

# 3. Check results
# The script shows:
# - Organization ID from token
# - Created order ID and number
# - Order status updates
# - Total orders count
```

---

## 📝 Expected Responses

### Success - Create Order (201)
```json
{
  "data": {
    "id": "ord-123",
    "order_number": "ORD-2025-001",
    "status": "pending",
    "total_amount": 5755.95,
    "buyer_organization_id": "org-456",
    "seller_organization_id": "org-789"
  }
}
```

### Success - List Orders (200)
```json
{
  "data": [...],
  "meta": {
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 5,
      "has_next": false
    }
  }
}
```

### Error - Invalid Organization (401)
```json
{
  "error": {
    "code": "INVALID_ORG",
    "message": "Valid organization ID is required"
  }
}
```

---

## 🛠️ Customization

### Change Base URL
Edit the variable in your testing method:

**Postman:** Update `baseUrl` variable
**Script:** Edit `BASE_URL` in `quick-test-orders.sh`
**Manual:** Replace `http://localhost:8080` in commands

### Use Different User
1. Register new user via `/api/v1/auth/register`
2. Login to get new token
3. Use new token in subsequent requests

---

## 📚 Additional Resources

- **Swagger/OpenAPI:** Visit http://localhost:8080/docs for interactive API docs
- **Full API Spec:** See `manual-testing-orders.md` for complete endpoint reference
- **Health Check:** http://localhost:8080/health

---

## 💡 Tips

1. **Always check your token** for organization_id before testing orders
2. **Use jq** for better JSON formatting: `brew install jq`
3. **Save common values** as shell variables for easier testing
4. **Check server logs** if you get unexpected errors
5. **Use Postman's Test tab** to add assertions and auto-extract IDs

---

Need help? Check the detailed guide in `manual-testing-orders.md`!
