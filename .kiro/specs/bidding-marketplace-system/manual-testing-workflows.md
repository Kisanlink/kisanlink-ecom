# Manual Workflow Testing Document - Bidding Marketplace System

## Overview

This document provides comprehensive manual testing workflows for the KisanLink E-commerce bidding marketplace system. The system enables Admins to create products, Listers to list them with asking prices, and Buyers to place time-limited bids.

## System Roles & Permissions

- **Admin**: Can create and manage products in the catalog
- **Lister**: Can list products with asking prices (becomes seller)
- **Buyer**: Can place bids on listed products with time limits
- **System**: Manages bid lifecycle, notifications, and automatic closures

## Testing Environment Setup

### Prerequisites

1. **Base URL**: `http://localhost:8080` (or your configured server)
2. **Authentication**: JWT tokens from AAA service
3. **Required Headers**:
   ```
   Authorization: Bearer <jwt_token>
   Content-Type: application/json
   X-Organization-ID: <org_id>
   ```
4. **Test Data**: Sample products, users, and organizations

### Authentication Setup

```bash
# Get authentication token (replace with actual AAA service endpoint)
curl -X POST "http://localhost:8081/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin@kisanlink.com",
    "password": "password123"
  }'

# Extract token from response
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
ORG_ID="org_12345"
```

## Workflow 1: Admin Product Creation

### Test Case 1.1: Create Basic Product

**Objective**: Verify admin can create products in the catalog

**Steps**:

1. **Create Product Request**:

```bash
curl -X POST "http://localhost:8080/api/v1/catalog/products" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Organization-ID: $ORG_ID" \
  -d '{
    "item_type": "PRODUCT",
    "name": "Organic Basmati Rice",
    "description": "Premium quality organic basmati rice, 1kg pack",
    "category": "GRAINS",
    "subcategory": "RICE",
    "base_price": 150.00,
    "currency": "INR",
    "unit_of_measure": "KG",
    "sku": "ORG-RICE-001",
    "visibility": "PUBLIC",
    "tags": ["organic", "basmati", "premium"],
    "attributes": {
      "origin": "Punjab",
      "certification": "Organic India",
      "shelf_life": "12 months"
    },
    "images": ["https://example.com/rice1.jpg"]
  }'
```

2. **Expected Response** (201 Created):

```json
{
  "success": true,
  "data": {
    "id": "prod_001",
    "name": "Organic Basmati Rice",
    "description": "Premium quality organic basmati rice, 1kg pack",
    "category": "GRAINS",
    "base_price": 150.0,
    "currency": "INR",
    "created_at": "2024-01-15T10:30:00Z",
    "organization_id": "org_12345"
  },
  "meta": {
    "trace_id": "trace_123"
  }
}
```

3. **Verification**:

```bash
# Verify product exists
curl -X GET "http://localhost:8080/api/v1/catalog/products/prod_001" \
  -H "Authorization: Bearer $TOKEN"
```

### Test Case 1.2: Create Product with Invalid Data

**Objective**: Verify validation errors are properly handled

**Steps**:

1. **Request with Missing Required Fields**:

```bash
curl -X POST "http://localhost:8080/api/v1/catalog/products" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "item_type": "PRODUCT",
    "description": "Missing name field"
  }'
```

2. **Expected Response** (400 Bad Request):

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request data",
    "details": {
      "name": "Name is required",
      "base_price": "Base price is required"
    }
  }
}
```

## Workflow 2: Lister Product Listing (Ask)

### Test Case 2.1: Create Product Listing

**Objective**: Verify lister can create product listings with asking prices

**API Endpoint**: `POST /api/v1/marketplace/listings`

**Steps**:

1. **Create Listing Request**:

```bash
curl -X POST "http://localhost:8080/api/v1/marketplace/listings" \
  -H "Authorization: Bearer $LISTER_TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Organization-ID: $LISTER_ORG_ID" \
  -d '{
    "product_id": "prod_001",
    "quantity": 100,
    "asking_price": 140.00,
    "currency": "INR",
    "minimum_bid": 120.00,
    "listing_duration_hours": 72,
    "pickup_location": {
      "address": "Farm Gate, Punjab",
      "coordinates": {
        "latitude": 30.7333,
        "longitude": 76.7794
      }
    },
    "terms_conditions": "Payment on delivery, quality guarantee",
    "listing_type": "AUCTION"
  }'
```

2. **Expected Response** (201 Created):

```json
{
  "success": true,
  "data": {
    "listing_id": "list_001",
    "product_id": "prod_001",
    "seller_id": "user_seller_001",
    "quantity": 100,
    "asking_price": 140.0,
    "minimum_bid": 120.0,
    "current_highest_bid": null,
    "status": "ACTIVE",
    "expires_at": "2024-01-18T10:30:00Z",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

3. **Verification**:

```bash
# Verify listing is active
curl -X GET "http://localhost:8080/api/v1/marketplace/listings/list_001" \
  -H "Authorization: Bearer $TOKEN"
```

### Test Case 2.2: List Active Listings

**Objective**: Verify buyers can see active listings

**Steps**:

1. **Get Active Listings**:

```bash
curl -X GET "http://localhost:8080/api/v1/marketplace/listings?status=ACTIVE&page=1&limit=10" \
  -H "Authorization: Bearer $BUYER_TOKEN"
```

2. **Expected Response** (200 OK):

```json
{
  "success": true,
  "data": [
    {
      "listing_id": "list_001",
      "product": {
        "id": "prod_001",
        "name": "Organic Basmati Rice",
        "category": "GRAINS"
      },
      "seller": {
        "id": "user_seller_001",
        "name": "Farm Fresh Co."
      },
      "asking_price": 140.0,
      "minimum_bid": 120.0,
      "current_highest_bid": null,
      "bid_count": 0,
      "time_remaining": "71 hours 45 minutes",
      "expires_at": "2024-01-18T10:30:00Z"
    }
  ],
  "meta": {
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 1,
      "has_next": false
    }
  }
}
```

## Workflow 3: Buyer Bidding System

### Test Case 3.1: Place First Bid

**Objective**: Verify buyer can place initial bid on listing

**API Endpoint**: `POST /api/v1/marketplace/listings/{listing_id}/bids`

**Steps**:

1. **Place Bid Request**:

```bash
curl -X POST "http://localhost:8080/api/v1/marketplace/listings/list_001/bids" \
  -H "Authorization: Bearer $BUYER_TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Organization-ID: $BUYER_ORG_ID" \
  -d '{
    "bid_amount": 125.00,
    "currency": "INR",
    "quantity": 50,
    "message": "Interested in 50kg for immediate delivery",
    "auto_bid_limit": 135.00,
    "payment_method": "BANK_TRANSFER"
  }'
```

2. **Expected Response** (201 Created):

```json
{
  "success": true,
  "data": {
    "bid_id": "bid_001",
    "listing_id": "list_001",
    "bidder_id": "user_buyer_001",
    "bid_amount": 125.0,
    "quantity": 50,
    "status": "ACTIVE",
    "is_highest_bid": true,
    "auto_bid_limit": 135.0,
    "placed_at": "2024-01-15T11:00:00Z"
  }
}
```

3. **Verification - Check Updated Listing**:

```bash
curl -X GET "http://localhost:8080/api/v1/marketplace/listings/list_001" \
  -H "Authorization: Bearer $TOKEN"
```

**Expected Updated Listing**:

```json
{
  "success": true,
  "data": {
    "listing_id": "list_001",
    "current_highest_bid": {
      "bid_id": "bid_001",
      "amount": 125.0,
      "bidder_name": "Anonymous Buyer #1"
    },
    "bid_count": 1,
    "status": "ACTIVE"
  }
}
```

### Test Case 3.2: Place Competing Bid

**Objective**: Verify second buyer can outbid first buyer

**Steps**:

1. **Higher Bid from Different Buyer**:

```bash
curl -X POST "http://localhost:8080/api/v1/marketplace/listings/list_001/bids" \
  -H "Authorization: Bearer $BUYER2_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "bid_amount": 130.00,
    "currency": "INR",
    "quantity": 100,
    "message": "Need full quantity, willing to pay premium"
  }'
```

2. **Expected Response** (201 Created):

```json
{
  "success": true,
  "data": {
    "bid_id": "bid_002",
    "listing_id": "list_001",
    "bid_amount": 130.0,
    "quantity": 100,
    "status": "ACTIVE",
    "is_highest_bid": true,
    "placed_at": "2024-01-15T11:15:00Z"
  }
}
```

3. **Verify First Bid is Outbid**:

```bash
curl -X GET "http://localhost:8080/api/v1/marketplace/bids/bid_001" \
  -H "Authorization: Bearer $BUYER_TOKEN"
```

**Expected Response**:

```json
{
  "success": true,
  "data": {
    "bid_id": "bid_001",
    "status": "OUTBID",
    "is_highest_bid": false,
    "outbid_at": "2024-01-15T11:15:00Z"
  }
}
```

### Test Case 3.3: Auto-Bid Activation

**Objective**: Verify auto-bidding works when configured

**Steps**:

1. **Place Bid that Triggers Auto-Bid**:

```bash
curl -X POST "http://localhost:8080/api/v1/marketplace/listings/list_001/bids" \
  -H "Authorization: Bearer $BUYER3_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "bid_amount": 132.00,
    "currency": "INR",
    "quantity": 100
  }'
```

2. **Expected Auto-Bid Response** (201 Created):

```json
{
  "success": true,
  "data": {
    "bid_id": "bid_003",
    "listing_id": "list_001",
    "bid_amount": 132.0,
    "status": "OUTBID",
    "outbid_by_auto_bid": true,
    "auto_bid_response": {
      "bid_id": "bid_004",
      "bid_amount": 133.0,
      "bidder_id": "user_buyer_001"
    }
  }
}
```

### Test Case 3.4: Invalid Bid Scenarios

**Objective**: Verify proper validation of bid constraints

**Steps**:

1. **Bid Below Minimum**:

```bash
curl -X POST "http://localhost:8080/api/v1/marketplace/listings/list_001/bids" \
  -H "Authorization: Bearer $BUYER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "bid_amount": 100.00,
    "currency": "INR",
    "quantity": 50
  }'
```

**Expected Response** (400 Bad Request):

```json
{
  "success": false,
  "error": {
    "code": "BID_TOO_LOW",
    "message": "Bid amount must be at least 120.00 INR",
    "details": {
      "minimum_bid": 120.0,
      "provided_bid": 100.0
    }
  }
}
```

2. **Bid on Own Listing**:

```bash
curl -X POST "http://localhost:8080/api/v1/marketplace/listings/list_001/bids" \
  -H "Authorization: Bearer $LISTER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "bid_amount": 140.00,
    "currency": "INR",
    "quantity": 50
  }'
```

**Expected Response** (403 Forbidden):

```json
{
  "success": false,
  "error": {
    "code": "CANNOT_BID_OWN_LISTING",
    "message": "Cannot place bid on your own listing"
  }
}
```

## Workflow 4: Bid Management & History

### Test Case 4.1: View Bid History

**Objective**: Verify users can view bid history for a listing

**Steps**:

1. **Get Bid History**:

```bash
curl -X GET "http://localhost:8080/api/v1/marketplace/listings/list_001/bids?page=1&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

2. **Expected Response** (200 OK):

```json
{
  "success": true,
  "data": [
    {
      "bid_id": "bid_004",
      "bid_amount": 133.0,
      "quantity": 100,
      "status": "ACTIVE",
      "is_highest_bid": true,
      "placed_at": "2024-01-15T11:16:00Z",
      "bidder_name": "Anonymous Buyer #1"
    },
    {
      "bid_id": "bid_003",
      "bid_amount": 132.0,
      "quantity": 100,
      "status": "OUTBID",
      "placed_at": "2024-01-15T11:15:30Z",
      "bidder_name": "Anonymous Buyer #3"
    }
  ],
  "meta": {
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 4
    }
  }
}
```

### Test Case 4.2: User's Own Bids

**Objective**: Verify buyer can view their own bid history

**Steps**:

1. **Get My Bids**:

```bash
curl -X GET "http://localhost:8080/api/v1/marketplace/bids/my-bids?status=ALL&page=1" \
  -H "Authorization: Bearer $BUYER_TOKEN"
```

2. **Expected Response** (200 OK):

```json
{
  "success": true,
  "data": [
    {
      "bid_id": "bid_004",
      "listing": {
        "listing_id": "list_001",
        "product_name": "Organic Basmati Rice",
        "seller_name": "Farm Fresh Co."
      },
      "bid_amount": 133.0,
      "quantity": 100,
      "status": "ACTIVE",
      "is_highest_bid": true,
      "placed_at": "2024-01-15T11:16:00Z",
      "expires_at": "2024-01-18T10:30:00Z"
    },
    {
      "bid_id": "bid_001",
      "bid_amount": 125.0,
      "status": "OUTBID",
      "is_highest_bid": false,
      "placed_at": "2024-01-15T11:00:00Z"
    }
  ]
}
```

## Workflow 5: Time-Limited Auctions

### Test Case 5.1: Auction Expiry Simulation

**Objective**: Verify automatic auction closure when time expires

**Steps**:

1. **Check Auction Status Near Expiry**:

```bash
curl -X GET "http://localhost:8080/api/v1/marketplace/listings/list_001" \
  -H "Authorization: Bearer $TOKEN"
```

2. **Wait for Expiry or Trigger Manual Close** (Admin action):

```bash
curl -X POST "http://localhost:8080/api/v1/marketplace/listings/list_001/close" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "EXPIRED",
    "force_close": false
  }'
```

3. **Expected Response** (200 OK):

```json
{
  "success": true,
  "data": {
    "listing_id": "list_001",
    "status": "CLOSED",
    "closed_at": "2024-01-18T10:30:00Z",
    "winning_bid": {
      "bid_id": "bid_004",
      "bidder_id": "user_buyer_001",
      "final_amount": 133.0,
      "quantity": 100
    },
    "total_bids": 4
  }
}
```

### Test Case 5.2: Post-Auction Order Creation

**Objective**: Verify order creation from winning bid

**Steps**:

1. **Create Order from Winning Bid**:

```bash
curl -X POST "http://localhost:8080/api/v1/orders/from-bid" \
  -H "Authorization: Bearer $BUYER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "bid_id": "bid_004",
    "shipping_address": {
      "street": "123 Farm Road",
      "city": "Delhi",
      "state": "Delhi",
      "zip": "110001",
      "country": "India"
    },
    "payment_method": "BANK_TRANSFER",
    "notes": "Please deliver by truck"
  }'
```

2. **Expected Response** (201 Created):

```json
{
  "success": true,
  "data": {
    "order_id": "ord_001",
    "listing_id": "list_001",
    "bid_id": "bid_004",
    "buyer_id": "user_buyer_001",
    "seller_id": "user_seller_001",
    "product_id": "prod_001",
    "quantity": 100,
    "unit_price": 133.0,
    "total_amount": 13300.0,
    "status": "PENDING_PAYMENT",
    "created_at": "2024-01-18T10:35:00Z"
  }
}
```

## Workflow 6: Error Scenarios & Edge Cases

### Test Case 6.1: Expired Listing Bid Attempt

**Objective**: Verify bids are rejected on expired listings

**Steps**:

1. **Attempt Bid on Expired Listing**:

```bash
curl -X POST "http://localhost:8080/api/v1/marketplace/listings/list_001/bids" \
  -H "Authorization: Bearer $BUYER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "bid_amount": 150.00,
    "currency": "INR",
    "quantity": 50
  }'
```

2. **Expected Response** (400 Bad Request):

```json
{
  "success": false,
  "error": {
    "code": "LISTING_EXPIRED",
    "message": "Cannot place bid on expired listing",
    "details": {
      "listing_id": "list_001",
      "expired_at": "2024-01-18T10:30:00Z"
    }
  }
}
```

### Test Case 6.2: Insufficient Inventory

**Objective**: Verify listing creation fails with insufficient inventory

**Steps**:

1. **Create Listing with Excessive Quantity**:

```bash
curl -X POST "http://localhost:8080/api/v1/marketplace/listings" \
  -H "Authorization: Bearer $LISTER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "prod_001",
    "quantity": 10000,
    "asking_price": 140.00
  }'
```

2. **Expected Response** (400 Bad Request):

```json
{
  "success": false,
  "error": {
    "code": "INSUFFICIENT_INVENTORY",
    "message": "Requested quantity exceeds available inventory",
    "details": {
      "requested": 10000,
      "available": 500
    }
  }
}
```

## Workflow 7: Admin Management Operations

### Test Case 7.1: Admin Listing Management

**Objective**: Verify admin can manage any listing

**Steps**:

1. **Admin Force Close Listing**:

```bash
curl -X POST "http://localhost:8080/api/v1/marketplace/listings/list_002/close" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "POLICY_VIOLATION",
    "force_close": true,
    "admin_notes": "Seller reported for fraudulent activity"
  }'
```

2. **Admin View All Listings**:

```bash
curl -X GET "http://localhost:8080/api/v1/admin/marketplace/listings?status=ALL&page=1" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

3. **Admin Bid Management**:

```bash
curl -X DELETE "http://localhost:8080/api/v1/admin/marketplace/bids/bid_005" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "FRAUDULENT_BID",
    "notify_user": true
  }'
```

## Monitoring & Analytics Workflows

### Test Case 8.1: Marketplace Metrics

**Objective**: Verify analytics endpoints work correctly

**Steps**:

1. **Get Marketplace Statistics**:

```bash
curl -X GET "http://localhost:8080/api/v1/analytics/marketplace/stats?period=7d" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

2. **Expected Response**:

```json
{
  "success": true,
  "data": {
    "period": "7d",
    "total_listings": 45,
    "active_listings": 12,
    "completed_auctions": 28,
    "total_bids": 234,
    "average_bid_count": 5.2,
    "top_categories": [
      { "category": "GRAINS", "listings": 15 },
      { "category": "VEGETABLES", "listings": 12 }
    ],
    "total_value_traded": 125000.0
  }
}
```

## Performance Testing Scenarios

### Test Case 9.1: Concurrent Bidding

**Objective**: Test system behavior under concurrent bid load

**Steps**:

1. **Simulate Concurrent Bids** (Use testing tools like Apache Bench):

```bash
# Create test script with multiple curl commands
# Run 10 concurrent requests with 0.1s intervals
for i in {1..10}; do
  curl -X POST "http://localhost:8080/api/v1/marketplace/listings/list_001/bids" \
    -H "Authorization: Bearer $BUYER_TOKEN_$i" \
    -H "Content-Type: application/json" \
    -d "{\"bid_amount\": $((130 + i)).00, \"quantity\": 10}" &
done
wait
```

2. **Verify Data Consistency**:
   - Check that only one bid is marked as highest
   - Verify bid sequence is maintained
   - Ensure no race conditions occurred

## Test Data Cleanup

### Cleanup Scripts

```bash
# Clean up test data after testing
curl -X DELETE "http://localhost:8080/api/v1/admin/test-data/cleanup" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "delete_test_products": true,
    "delete_test_listings": true,
    "delete_test_bids": true,
    "delete_test_orders": true
  }'
```

## Conclusion

This manual testing document covers the complete bidding marketplace workflow from product creation to order fulfillment. Each test case includes:

- Clear objectives
- Step-by-step procedures
- Expected responses
- Error scenarios
- Performance considerations

The tests validate the complete business workflow while ensuring system reliability, security, and user experience standards are met.

## Notes for Testers

1. **Authentication Tokens**: Ensure you have valid tokens for different user roles
2. **Test Environment**: Run tests in a dedicated test environment
3. **Data Persistence**: Verify data persists correctly across operations
4. **Error Handling**: Test both success and failure scenarios
5. **Performance**: Monitor response times during testing
6. **Security**: Verify authorization controls work correctly
7. **Cleanup**: Always clean up test data after testing sessions
