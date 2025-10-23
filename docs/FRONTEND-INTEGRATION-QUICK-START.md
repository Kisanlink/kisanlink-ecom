# Frontend Integration Quick Start Guide

**Last Updated:** October 2025
**API Base URL:** `http://localhost:8080`
**Required:** JWT Token with `organization_id`

---

## 🚀 Quick Integration Steps

### Step 1: Authentication (5 minutes)

```javascript
// 1. Configure Axios
import axios from 'axios';

const api = axios.create({
  baseURL: 'http://localhost:8080',
  headers: {
    'Content-Type': 'application/json'
  }
});

// 2. Add auth interceptor
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// 3. Login function
export const login = async (username, password) => {
  const response = await api.post('/api/v1/auth/login', {
    username,
    password
  });

  const { access_token, user } = response.data.data;

  // IMPORTANT: Store these
  localStorage.setItem('access_token', access_token);
  localStorage.setItem('organization_id', user.organization_id);
  localStorage.setItem('user', JSON.stringify(user));

  return user;
};
```

**✅ Test it:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test_user","password":"password"}'
```

---

### Step 2: Verify Organization ID (2 minutes)

```javascript
// Check if user has organization
export const checkOrganization = () => {
  const orgId = localStorage.getItem('organization_id');

  if (!orgId || orgId === 'null' || orgId === 'undefined') {
    throw new Error('User not assigned to organization');
  }

  return orgId;
};

// Use before any orders API call
const createOrder = async (orderData) => {
  const orgId = checkOrganization(); // Will throw if missing

  return await api.post('/api/v1/orders', {
    buyer_organization_id: orgId,
    ...orderData
  });
};
```

**⚠️ CRITICAL:** Orders API will return 401 if `organization_id` is missing from token!

---

### Step 3: Create Your First Order (10 minutes)

```javascript
export const createOrder = async ({
  sellerId,
  items,
  deliveryAddress,
  notes = ''
}) => {
  const orgId = checkOrganization();

  const payload = {
    buyer_organization_id: orgId,
    seller_organization_id: sellerId,
    items: items.map(item => ({
      catalog_item_id: item.id,
      catalog_item_type: item.type, // 'product', 'service', or 'labour'
      quantity: item.quantity,
      unit_price: item.price,
      currency: 'INR',
      tax_percentage: item.tax || 18,
      discount_percentage: item.discount || 0,
      notes: item.notes || ''
    })),
    delivery_address: {
      street: deliveryAddress.street,
      city: deliveryAddress.city,
      state: deliveryAddress.state,
      postal_code: deliveryAddress.postalCode,
      country: deliveryAddress.country || 'India'
    },
    notes
  };

  try {
    const response = await api.post('/api/v1/orders', payload);
    return response.data.data; // Returns created order
  } catch (error) {
    handleApiError(error);
    throw error;
  }
};

// Usage example
const order = await createOrder({
  sellerId: 'org_seller123',
  items: [
    {
      id: 'prod_001',
      type: 'product',
      quantity: 100,
      price: 25.50,
      tax: 18
    }
  ],
  deliveryAddress: {
    street: '123 Farm Road',
    city: 'Pune',
    state: 'Maharashtra',
    postalCode: '411014'
  },
  notes: 'Urgent delivery required'
});

console.log('Order created:', order.order_number);
```

**✅ Test it:**
```bash
# Replace YOUR_TOKEN with actual token
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "buyer_organization_id": "YOUR_ORG_ID",
    "seller_organization_id": "org_seller123",
    "items": [{
      "catalog_item_id": "prod_001",
      "catalog_item_type": "product",
      "quantity": 10,
      "unit_price": 100,
      "currency": "INR"
    }],
    "delivery_address": {
      "street": "123 Test St",
      "city": "Pune",
      "state": "Maharashtra",
      "postal_code": "411014",
      "country": "India"
    }
  }'
```

---

### Step 4: List Orders (5 minutes)

```javascript
export const fetchOrders = async ({
  page = 1,
  limit = 20,
  status = null,
  buyerId = null,
  sellerId = null
} = {}) => {
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString()
  });

  if (status) params.append('status', status);
  if (buyerId) params.append('buyer_id', buyerId);
  if (sellerId) params.append('seller_id', sellerId);

  const response = await api.get(`/api/v1/orders?${params}`);

  return {
    orders: response.data.data,
    pagination: response.data.meta.pagination
  };
};

// Usage examples
const { orders, pagination } = await fetchOrders();
const pendingOrders = await fetchOrders({ status: 'pending' });
const page2Orders = await fetchOrders({ page: 2, limit: 10 });
```

**✅ Test it:**
```bash
curl -X GET "http://localhost:8080/api/v1/orders?page=1&limit=10" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

### Step 5: Get Order Details (3 minutes)

```javascript
export const getOrderDetails = async (orderId) => {
  const response = await api.get(`/api/v1/orders/${orderId}`);
  return response.data.data;
};

// Usage
const order = await getOrderDetails('ord_abc123');
console.log('Order total:', order.total_amount);
console.log('Status:', order.status);
console.log('Items:', order.items.length);
```

**✅ Test it:**
```bash
curl -X GET http://localhost:8080/api/v1/orders/ord_abc123 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

### Step 6: Update Order Status (5 minutes)

```javascript
export const updateOrderStatus = async (orderId, newStatus, notes = '') => {
  await api.patch(`/api/v1/orders/${orderId}/status`, {
    status: newStatus,
    notes
  });
};

// Valid statuses
const OrderStatus = {
  PENDING: 'pending',
  CONFIRMED: 'confirmed',
  PROCESSING: 'processing',
  SHIPPED: 'shipped',
  DELIVERED: 'delivered',
  CANCELLED: 'cancelled',
  FAILED: 'failed'
};

// Usage
await updateOrderStatus('ord_abc123', OrderStatus.CONFIRMED, 'Order confirmed by seller');
```

**✅ Test it:**
```bash
curl -X PATCH http://localhost:8080/api/v1/orders/ord_abc123/status \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"confirmed","notes":"Confirmed"}'
```

---

### Step 7: Error Handling (10 minutes)

```javascript
// Centralized error handler
export const handleApiError = (error) => {
  if (!error.response) {
    // Network error
    return {
      title: 'Network Error',
      message: 'Please check your internet connection',
      code: 'NETWORK_ERROR'
    };
  }

  const { status, data } = error.response;

  switch (status) {
    case 401:
      if (data.error.code === 'INVALID_ORG') {
        // No organization assigned
        return {
          title: 'Organization Required',
          message: 'Your account is not associated with an organization. Please contact support.',
          code: 'NO_ORGANIZATION',
          action: 'contact_support'
        };
      } else {
        // Invalid/expired token
        localStorage.clear();
        window.location.href = '/login';
        return {
          title: 'Session Expired',
          message: 'Please login again',
          code: 'UNAUTHORIZED',
          action: 'redirect_login'
        };
      }

    case 400:
      return {
        title: 'Invalid Request',
        message: data.error.message,
        details: data.error.details,
        code: data.error.code
      };

    case 403:
      return {
        title: 'Permission Denied',
        message: 'You do not have permission to perform this action',
        code: 'FORBIDDEN'
      };

    case 404:
      return {
        title: 'Not Found',
        message: data.error.message,
        code: data.error.code
      };

    case 500:
      console.error('Server error. Trace ID:', data.meta?.trace_id);
      return {
        title: 'Server Error',
        message: 'Something went wrong. Please try again later.',
        code: 'INTERNAL_ERROR',
        traceId: data.meta?.trace_id
      };

    default:
      return {
        title: 'Error',
        message: 'An unexpected error occurred',
        code: 'UNKNOWN_ERROR'
      };
  }
};

// Use in your UI
try {
  await createOrder(orderData);
  showSuccess('Order created successfully');
} catch (error) {
  const errorInfo = handleApiError(error);
  showError(errorInfo.title, errorInfo.message);

  // Handle specific actions
  if (errorInfo.action === 'contact_support') {
    showSupportDialog();
  }
}
```

---

## 📦 Complete React Hook Example

```javascript
// useOrders.js
import { useState, useEffect, useCallback } from 'react';
import { api, handleApiError, checkOrganization } from './api';

export const useOrders = () => {
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [pagination, setPagination] = useState({
    page: 1,
    limit: 20,
    total: 0,
    hasNext: false
  });

  // Fetch orders
  const fetchOrders = useCallback(async (filters = {}) => {
    setLoading(true);
    setError(null);

    try {
      const params = {
        page: pagination.page,
        limit: pagination.limit,
        ...filters
      };

      const response = await api.get('/api/v1/orders', { params });

      setOrders(response.data.data);
      setPagination(response.data.meta.pagination);
    } catch (err) {
      const errorInfo = handleApiError(err);
      setError(errorInfo);
    } finally {
      setLoading(false);
    }
  }, [pagination.page, pagination.limit]);

  // Create order
  const createOrder = async (orderData) => {
    setLoading(true);
    setError(null);

    try {
      const orgId = checkOrganization();

      const response = await api.post('/api/v1/orders', {
        buyer_organization_id: orgId,
        ...orderData
      });

      // Refresh orders
      await fetchOrders();

      return response.data.data;
    } catch (err) {
      const errorInfo = handleApiError(err);
      setError(errorInfo);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  // Update status
  const updateStatus = async (orderId, status, notes = '') => {
    try {
      await api.patch(`/api/v1/orders/${orderId}/status`, {
        status,
        notes
      });

      // Refresh orders
      await fetchOrders();
    } catch (err) {
      const errorInfo = handleApiError(err);
      setError(errorInfo);
      throw err;
    }
  };

  // Cancel order
  const cancelOrder = async (orderId) => {
    try {
      await api.post(`/api/v1/orders/${orderId}/cancel`);

      // Refresh orders
      await fetchOrders();
    } catch (err) {
      const errorInfo = handleApiError(err);
      setError(errorInfo);
      throw err;
    }
  };

  // Go to page
  const goToPage = (page) => {
    setPagination(prev => ({ ...prev, page }));
  };

  // Load initial data
  useEffect(() => {
    fetchOrders();
  }, [pagination.page]);

  return {
    orders,
    loading,
    error,
    pagination,
    createOrder,
    updateStatus,
    cancelOrder,
    fetchOrders,
    goToPage
  };
};

// Usage in component
const OrdersPage = () => {
  const {
    orders,
    loading,
    error,
    pagination,
    createOrder,
    updateStatus,
    goToPage
  } = useOrders();

  if (loading) return <Spinner />;
  if (error) return <ErrorAlert error={error} />;

  return (
    <div>
      <OrdersList orders={orders} onUpdateStatus={updateStatus} />
      <Pagination
        current={pagination.page}
        total={pagination.total}
        pageSize={pagination.limit}
        onChange={goToPage}
      />
    </div>
  );
};
```

---

## 🧪 Testing Checklist

### Before You Start

- [ ] Backend server is running on `http://localhost:8080`
- [ ] You have valid login credentials
- [ ] User is assigned to an organization (check with backend team)

### Authentication Tests

- [ ] Login returns access_token
- [ ] Login returns organization_id in user object
- [ ] Token is stored in localStorage
- [ ] Authorization header is added to all requests
- [ ] 401 errors redirect to login

### Orders Tests

- [ ] Can create order with valid data
- [ ] Creating order without org_id fails with INVALID_ORG
- [ ] Can list orders with pagination
- [ ] Can filter orders by status
- [ ] Can get order details
- [ ] Can update order status
- [ ] Invalid status transition shows error
- [ ] Can cancel order
- [ ] Cannot cancel delivered order

### Error Handling Tests

- [ ] Network errors show appropriate message
- [ ] 401 errors handled correctly
- [ ] 400 validation errors show details
- [ ] 404 errors handled
- [ ] 500 errors logged with trace ID

---

## 🔧 Debugging Tools

### Check Token Contents
```javascript
const token = localStorage.getItem('access_token');
if (token) {
  const payload = JSON.parse(atob(token.split('.')[1]));
  console.log('Token payload:', payload);
  console.log('Organization ID:', payload.organization_id);
  console.log('User ID:', payload.user_id || payload.sub);
  console.log('Expires:', new Date(payload.exp * 1000));
}
```

### API Request Logger
```javascript
// Add request/response logging
api.interceptors.request.use((config) => {
  console.log('→ Request:', config.method.toUpperCase(), config.url);
  console.log('  Headers:', config.headers);
  console.log('  Data:', config.data);
  return config;
});

api.interceptors.response.use(
  (response) => {
    console.log('← Response:', response.status, response.config.url);
    console.log('  Data:', response.data);
    return response;
  },
  (error) => {
    console.error('← Error:', error.response?.status, error.config?.url);
    console.error('  Error:', error.response?.data);
    return Promise.reject(error);
  }
);
```

### Check Organization Status
```javascript
const checkUserOrganization = () => {
  const orgId = localStorage.getItem('organization_id');
  const user = JSON.parse(localStorage.getItem('user') || '{}');

  console.log('Organization ID:', orgId);
  console.log('User:', user);

  if (!orgId || orgId === 'null') {
    console.error('❌ No organization assigned!');
    console.log('Contact backend team to assign user to organization');
    return false;
  }

  console.log('✅ Organization:', orgId);
  return true;
};
```

---

## 🚨 Common Issues & Solutions

### Issue 1: "INVALID_ORG" Error

**Symptom:** All orders requests return 401 with code `INVALID_ORG`

**Cause:** User's JWT token doesn't have `organization_id`

**Solution:**
1. Check token: Run `checkUserOrganization()` from debugging tools
2. If missing, user needs to be assigned to organization in AAA service
3. After assignment, logout and login again to get new token

---

### Issue 2: "user can only create orders for their own organization"

**Symptom:** Create order fails with 400 error

**Cause:** `buyer_organization_id` in request doesn't match user's org

**Solution:**
```javascript
// ❌ Wrong - hardcoded
buyer_organization_id: "org_123"

// ✅ Correct - from user's token
buyer_organization_id: localStorage.getItem('organization_id')
```

---

### Issue 3: CORS Errors

**Symptom:** Browser shows CORS policy error

**Solution:** Backend should have CORS configured. If not, ask backend team to add:
```go
// Backend needs this
router.Use(cors.New(cors.Config{
    AllowOrigins: ["http://localhost:3000"],
    AllowMethods: ["GET", "POST", "PUT", "PATCH", "DELETE"],
    AllowHeaders: ["Origin", "Content-Type", "Authorization"],
}))
```

---

### Issue 4: Token Expired

**Symptom:** Requests work initially, then start returning 401

**Cause:** JWT token has expired (default: 1 hour)

**Solution:** Implement token refresh:
```javascript
// Refresh token before it expires
const refreshToken = async () => {
  const refresh_token = localStorage.getItem('refresh_token');

  const response = await api.post('/api/v1/auth/refresh', {
    refresh_token
  });

  localStorage.setItem('access_token', response.data.data.access_token);
};

// Auto-refresh on 401
api.interceptors.response.use(
  response => response,
  async (error) => {
    if (error.response?.status === 401 && error.config && !error.config._retry) {
      error.config._retry = true;
      await refreshToken();
      return api(error.config);
    }
    return Promise.reject(error);
  }
);
```

---

## 📞 Need Help?

1. **Check Token:** Use debugging tools above to verify organization_id
2. **Check Logs:** Look at browser console and network tab
3. **Check Backend:** Verify server is running and accessible
4. **Contact Team:** Share trace_id from error response for debugging

---

## 🎯 Success Criteria

You've successfully integrated when you can:

✅ Login and store token
✅ Create an order
✅ List orders with pagination
✅ View order details
✅ Update order status
✅ Handle all error cases gracefully
✅ Organization ID is present and validated

**Estimated Time:** 30-45 minutes for complete integration

Good luck! 🚀
