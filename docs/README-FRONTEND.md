# Frontend Integration Documentation - KisanLink E-Commerce API

**Quick Start:** This documentation package will help you integrate the KisanLink E-Commerce backend APIs with your frontend application.

---

## 📚 Documentation Index

### 🚀 **Start Here**

1. **[FRONTEND-INTEGRATION-QUICK-START.md](./FRONTEND-INTEGRATION-QUICK-START.md)** ⭐ **START HERE**
   - 30-minute integration guide
   - Copy-paste ready code examples
   - React hooks and components
   - Debugging tools
   - Common issues & solutions

### 📖 **Complete Reference**

2. **[API-INTEGRATION-GUIDE.md](./API-INTEGRATION-GUIDE.md)**
   - Complete API specification
   - All endpoints with examples
   - Request/response formats
   - Error codes and handling
   - Integration patterns
   - Frontend code examples

### 🧪 **Testing**

3. **[API-TEST-SCENARIOS.md](./API-TEST-SCENARIOS.md)**
   - Comprehensive test scenarios
   - Expected behaviors
   - Edge cases
   - Performance tests
   - Automation scripts

4. **[manual-testing-orders.md](./manual-testing-orders.md)**
   - Manual testing guide
   - curl examples for all endpoints
   - Quick testing workflows

5. **[TESTING-README.md](./TESTING-README.md)**
   - Testing overview
   - Postman collection guide
   - Quick test script

### 🛠️ **Tools**

6. **[Orders-API.postman_collection.json](./Orders-API.postman_collection.json)**
   - Import into Postman
   - Pre-configured requests
   - Auto-saves tokens and IDs

7. **[quick-test-orders.sh](./quick-test-orders.sh)** (Executable)
   - Automated test script
   - Validates complete flow
   - Run: `./quick-test-orders.sh`

---

## 🎯 Getting Started (5 Minutes)

### Prerequisites

- [x] Backend server running on `http://localhost:8080`
- [x] User account created via AAA service
- [x] User assigned to an organization (CRITICAL!)
- [x] Node.js and npm/yarn installed
- [x] Basic understanding of React/JavaScript

### Step 1: Verify Backend

```bash
# Check server is running
curl http://localhost:8080/health

# Expected: {"status":"healthy"}
```

### Step 2: Get Your Token

```bash
# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "your_username",
    "password": "your_password"
  }'
```

**IMPORTANT:** Save the `access_token` and `organization_id` from response!

### Step 3: Verify Organization ID

```bash
# Decode your token to check organization_id
echo "YOUR_TOKEN" | cut -d. -f2 | base64 -d | jq

# Look for: "organization_id": "org_..."
```

**⚠️ If `organization_id` is missing, STOP! Contact backend team to assign user to organization.**

### Step 4: Test Your First Request

```bash
# List orders
curl -X GET http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer YOUR_TOKEN"

# Should return 200 with orders array (may be empty)
# If you get 401 INVALID_ORG - organization_id is missing from token
```

### Step 5: Follow Quick Start Guide

Open **[FRONTEND-INTEGRATION-QUICK-START.md](./FRONTEND-INTEGRATION-QUICK-START.md)** and follow the 7 steps (30 minutes total).

---

## 🏗️ Architecture Overview

```
┌─────────────────┐
│   Frontend      │
│   (React/Vue)   │
└────────┬────────┘
         │
         │ HTTP/REST
         │
┌────────▼────────────────────────────────────┐
│          Backend API Server                 │
│          http://localhost:8080              │
├─────────────────────────────────────────────┤
│                                             │
│  ┌──────────────────────────────────────┐  │
│  │  /api/v1/auth/*                      │  │
│  │  - Login, Register, Token Refresh    │  │
│  └──────────────────────────────────────┘  │
│                                             │
│  ┌──────────────────────────────────────┐  │
│  │  /api/v1/orders/*  (Auth Required)   │  │
│  │  - Create, List, Update, Cancel      │  │
│  │  - Requires: organization_id in token│  │
│  └──────────────────────────────────────┘  │
│                                             │
│  ┌──────────────────────────────────────┐  │
│  │  /api/v1/catalog/*  (Public)         │  │
│  │  - Products, Services, Labour        │  │
│  └──────────────────────────────────────┘  │
│                                             │
│  ┌──────────────────────────────────────┐  │
│  │  /api/v1/marketplace/*               │  │
│  │  - Listings, Bids, Offers            │  │
│  └──────────────────────────────────────┘  │
│                                             │
└─────────────────────────────────────────────┘
         │
         │
┌────────▼────────┐
│   PostgreSQL    │
│   Database      │
└─────────────────┘
```

---

## 🔑 Key Concepts

### Authentication Flow

```javascript
1. Login → Get JWT token + user info
2. Store token in localStorage
3. Store organization_id from user object
4. Add token to all API requests
5. Handle 401 errors (refresh or re-login)
```

### Organization-Based Access

- Every user MUST belong to an organization
- `organization_id` is stored in JWT token
- Orders API requires valid organization_id
- Users can only create orders for their own organization

### Order Lifecycle

```
pending → confirmed → processing → shipped → delivered
   ↓
cancelled (from pending, confirmed, or processing)
```

---

## 📋 API Endpoints Summary

### Authentication (Public)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/login` | Login and get token |
| POST | `/api/v1/auth/refresh` | Refresh access token |

### Orders (Requires Auth + Org)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/orders` | Create new order |
| GET | `/api/v1/orders` | List orders (paginated) |
| GET | `/api/v1/orders/{id}` | Get order details |
| PUT | `/api/v1/orders/{id}` | Update order |
| PATCH | `/api/v1/orders/{id}/status` | Update order status |
| POST | `/api/v1/orders/{id}/cancel` | Cancel order |
| POST | `/api/v1/orders/from-bid` | Create order from bid |

### Catalog (Public)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/catalog/products` | List products |
| GET | `/api/v1/catalog/products/{id}` | Get product details |
| GET | `/api/v1/catalog/services` | List services |
| GET | `/api/v1/catalog/labour` | List labour |

### Marketplace (Requires Auth)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/marketplace/listings` | Create listing (ask) |
| POST | `/api/v1/marketplace/bids` | Create bid |
| POST | `/api/v1/marketplace/bids/{id}/accept` | Accept bid |

---

## ⚠️ Critical Requirements

### 1. Organization ID is MANDATORY for Orders API

```javascript
// ❌ WRONG - Will get 401 INVALID_ORG
const response = await api.post('/api/v1/orders', {
  buyer_organization_id: "hardcoded_value",
  ...
});

// ✅ CORRECT - Use from token
const orgId = localStorage.getItem('organization_id');
const response = await api.post('/api/v1/orders', {
  buyer_organization_id: orgId,
  ...
});
```

### 2. Authorization Header Required

```javascript
// ❌ WRONG
fetch('/api/v1/orders')

// ✅ CORRECT
fetch('/api/v1/orders', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
})
```

### 3. Buyer Org Must Match User's Org

```javascript
// User's organization from token: "org_123"

// ❌ WRONG - Will fail
buyer_organization_id: "org_456"

// ✅ CORRECT
buyer_organization_id: "org_123" // Same as user's org
```

---

## 🧪 Testing Checklist

Before integrating, verify:

- [ ] Backend server is accessible
- [ ] Can login successfully
- [ ] Token contains `organization_id`
- [ ] Can create an order
- [ ] Can list orders
- [ ] Can update order status
- [ ] Error handling works for 401, 400, 404, 500
- [ ] Pagination works
- [ ] Filters work

**Run the automated test:**
```bash
./docs/quick-test-orders.sh
```

---

## 🐛 Troubleshooting

### Issue: Getting 401 "INVALID_ORG"

**Problem:** Token doesn't have `organization_id`

**Solution:**
1. Check token contents:
```javascript
const payload = JSON.parse(atob(token.split('.')[1]));
console.log(payload.organization_id);
```

2. If missing, user needs to be assigned to organization in AAA service
3. After assignment, re-login to get new token

### Issue: "user can only create orders for their own organization"

**Problem:** `buyer_organization_id` doesn't match user's org

**Solution:**
```javascript
// Always use:
buyer_organization_id: localStorage.getItem('organization_id')
```

### Issue: CORS Errors

**Problem:** Browser blocking requests

**Solution:** Contact backend team to configure CORS for your frontend URL

### Issue: Token Expired

**Problem:** Token is only valid for 1 hour

**Solution:** Implement token refresh or re-login:
```javascript
api.interceptors.response.use(
  response => response,
  async (error) => {
    if (error.response?.status === 401) {
      // Refresh token or redirect to login
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);
```

---

## 📞 Support

### Documentation Issues

Found an error in docs? File an issue or contact the backend team.

### API Issues

- Check backend logs
- Include `trace_id` from error response
- Share request/response examples

### Integration Help

1. Check **[FRONTEND-INTEGRATION-QUICK-START.md](./FRONTEND-INTEGRATION-QUICK-START.md)** first
2. Review **[API-TEST-SCENARIOS.md](./API-TEST-SCENARIOS.md)** for examples
3. Test with Postman collection
4. Contact backend team with details

---

## 🚦 Integration Status

Track your progress:

- [ ] Authentication implemented
- [ ] Token storage working
- [ ] Organization ID verified
- [ ] Create order working
- [ ] List orders working
- [ ] Order details working
- [ ] Status updates working
- [ ] Error handling complete
- [ ] All tests passing
- [ ] Ready for production

---

## 📊 Quick Reference

### Required Headers

```
Content-Type: application/json
Authorization: Bearer <your_jwt_token>
```

### Common Status Codes

| Code | Meaning | Action |
|------|---------|--------|
| 200 | Success | Continue |
| 201 | Created | Success, resource created |
| 400 | Bad Request | Check request data |
| 401 | Unauthorized | Check token/org_id |
| 403 | Forbidden | Check permissions |
| 404 | Not Found | Check resource ID |
| 500 | Server Error | Contact backend team |

### Order Statuses

```
pending → confirmed → processing → shipped → delivered
   ↓
cancelled
```

### Sample Organization IDs

```
org_buyer_001    (Buyer)
org_seller_001   (Seller)
```

### Sample Catalog IDs

```
prod_tomato_seeds  (Product)
serv_soil_test     (Service)
labour_harvest     (Labour)
```

---

## 🎓 Learning Path

1. **Day 1:** Read Quick Start Guide, implement authentication
2. **Day 2:** Implement order creation and listing
3. **Day 3:** Add order details and status updates
4. **Day 4:** Implement error handling
5. **Day 5:** Testing and refinement

**Total Time:** 5 days for complete integration

---

## 📈 Success Metrics

Your integration is successful when:

✅ Users can login and get valid tokens
✅ Organization ID is present in all requests
✅ Orders can be created, viewed, and updated
✅ Errors are handled gracefully
✅ All test scenarios pass
✅ Performance is acceptable (< 1s response time)
✅ Frontend team is unblocked and productive

---

**Let's build something great! 🚀**

For questions or support, contact the backend team with this documentation as reference.
