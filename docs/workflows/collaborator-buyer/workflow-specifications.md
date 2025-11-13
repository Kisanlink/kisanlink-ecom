# Collaborator-Buyer Workflow Specifications

## Overview

Collaborator-Buyers are registered users who participate in the KisanLink marketplace to purchase agricultural products through competitive bidding. They can browse products, participate in auctions, manage bids, and create orders from winning bids.

## User Roles & Permissions

- **Role**: `BUYER`
- **Organization**: Partner organizations or individual buyers
- **Permissions**: Browse public listings, place bids, manage personal bids, create orders

## Workflow 1: Product Discovery & Browsing

### Description

Buyers discover and explore available products and active auction listings in the marketplace.

### Steps

1. **Authentication & Authorization**
   - User logs in with JWT token
   - System validates token and extracts user identity
   - Organization ID is validated if required

2. **Browse Product Catalog**
   - GET `/api/v1/catalog/products`
   - Filter by category, subcategory, price range
   - View product details including base pricing
   - Check product availability and specifications

3. **Search Active Listings**
   - GET `/api/v1/marketplace/listings?status=ACTIVE`
   - Filter by product type, location, price range
   - View listing details: asking price, minimum bid, time remaining
   - Check auction visibility and bidding rules

4. **View Listing Details**
   - GET `/api/v1/marketplace/listings/{listing_id}`
   - Review seller information
   - Check pickup/delivery locations
   - View current bid status (based on visibility settings)
   - Analyze auction rules and terms

### API Endpoints

- `GET /api/v1/catalog/products`
- `GET /api/v1/marketplace/listings`
- `GET /api/v1/marketplace/listings/{listing_id}`

### Success Criteria

- User can browse all accessible products and listings
- Filtering and search work correctly
- Listing details display according to visibility settings
- Performance remains under 500ms for search queries

---

## Workflow 2: Auction Participation & Bidding

### Description

Buyers participate in active auctions by placing competitive bids on listings.

### Steps

1. **Pre-Bid Validation**
   - Verify listing is ACTIVE and not expired
   - Check user is not the seller of the listing
   - Validate minimum bid requirements
   - Confirm inventory availability

2. **Place Initial Bid**
   - POST `/api/v1/marketplace/listings/{listing_id}/bids`
   - Submit bid amount above minimum or current highest bid
   - Include quantity, message, and payment method
   - System validates bid against business rules

3. **Bid Processing**
   - System atomically updates listing state
   - Previous highest bidder is notified of being outbid
   - Bid becomes new highest bid if valid
   - Audit trail is created for bid placement

4. **Real-time Updates**
   - Buyer receives confirmation of bid placement
   - System sends notifications for bid status changes
   - Listing updates with new highest bid (if visible)
   - Time remaining displays updated countdown

5. **Competitive Bidding**
   - Monitor other bidders' activities
   - Place counter-bids as needed
   - Receive outbid notifications
   - Adjust bidding strategy based on competition

### API Endpoints

- `POST /api/v1/marketplace/listings/{listing_id}/bids`
- `GET /api/v1/marketplace/listings/{listing_id}/bids`
- `GET /api/v1/marketplace/bids/{bid_id}`

### Business Rules

- Bid amount must exceed current highest bid
- Cannot bid on own listings
- Must have valid payment method
- Quantity cannot exceed available inventory
- Auction must be ACTIVE and not expired

### Success Criteria

- Bids are processed atomically without race conditions
- Proper validation prevents invalid bids
- Real-time notifications work correctly
- Bid history is accurately maintained

---

## Workflow 3: Auto-Bidding Management

### Description

Buyers can configure automatic bidding to compete efficiently without manual intervention.

### Steps

1. **Configure Auto-Bidding**
   - Set maximum bid limit during initial bid placement
   - Define increment strategy (fixed amount or percentage)
   - Specify response timing preferences
   - Configure notification preferences

2. **Auto-Bid Activation**
   - System monitors for competing bids
   - Automatically places counter-bids when outbid
   - Respects maximum limit constraints
   - Maintains competitive positioning

3. **Auto-Bid Management**
   - Update auto-bid limits as needed
   - Monitor auto-bid performance
   - Receive notifications for auto-bid activities
   - Disable auto-bidding when desired

4. **Limit Management**
   - Track remaining auto-bid capacity
   - Receive warnings when approaching limits
   - Option to increase limits during auction
   - Automatic disabling when limit reached

### API Endpoints

- `PUT /api/v1/marketplace/bids/{bid_id}/auto-bid`
- `GET /api/v1/marketplace/bids/my-bids?auto_bid=true`
- `DELETE /api/v1/marketplace/bids/{bid_id}/auto-bid`

### Success Criteria

- Auto-bidding responds quickly to competing bids
- Limits are respected accurately
- User maintains control over auto-bid settings
- Performance doesn't degrade with multiple auto-bidders

---

## Workflow 4: Bid History Tracking

### Description

Buyers can track their bidding history and analyze their marketplace activities.

### Steps

1. **View Personal Bid History**
   - GET `/api/v1/marketplace/bids/my-bids`
   - Filter by status (ACTIVE, OUTBID, WON, LOST)
   - Sort by date, amount, or status
   - View detailed bid information

2. **Analyze Bid Performance**
   - Track win/loss ratios
   - Review bidding patterns
   - Analyze spending trends
   - Identify successful strategies

3. **Monitor Active Bids**
   - View current active bids across all auctions
   - Check time remaining for each auction
   - Monitor bid status changes
   - Receive real-time updates

4. **Historical Analysis**
   - Review completed auctions
   - Analyze winning/losing bid amounts
   - Track seasonal pricing trends
   - Export bid history for analysis

### API Endpoints

- `GET /api/v1/marketplace/bids/my-bids`
- `GET /api/v1/marketplace/bids/my-bids/analytics`
- `GET /api/v1/marketplace/bids/{bid_id}/history`

### Success Criteria

- Complete bid history is available
- Filtering and sorting work efficiently
- Analytics provide meaningful insights
- Data export functions correctly

---

## Workflow 5: Order Creation from Winning Bids

### Description

When buyers win auctions, they can convert winning bids into purchase orders.

### Steps

1. **Auction Completion Notification**
   - Receive notification of winning bid
   - System presents order creation opportunity
   - Review final bid details and terms
   - Access order creation interface

2. **Order Details Configuration**
   - Pre-filled order form with bid details
   - Add shipping/delivery address
   - Select payment method
   - Specify delivery preferences
   - Add special instructions or notes

3. **Order Creation**
   - POST `/api/v1/orders/from-bid`
   - Submit complete order information
   - System validates all required fields
   - Order is created with PENDING_PAYMENT status

4. **Order Confirmation**
   - Receive order confirmation with details
   - Order ID and tracking information provided
   - Seller is notified of new order
   - Payment process is initiated

5. **Order Tracking**
   - Monitor order status updates
   - Track payment processing
   - Receive fulfillment notifications
   - Access order history

### API Endpoints

- `POST /api/v1/orders/from-bid`
- `GET /api/v1/orders/{order_id}`
- `GET /api/v1/orders/my-orders`
- `PUT /api/v1/orders/{order_id}/shipping`

### Business Rules

- Only winning bidders can create orders
- Order must be created within specified timeframe
- Shipping address must be valid and deliverable
- Payment method must be verified
- Order amount matches winning bid total

### Success Criteria

- Seamless transition from winning bid to order
- All order details are accurately captured
- Payment integration works correctly
- Order tracking provides real-time updates

---

## Workflow 6: Payment Processing

### Description

Buyers complete payments for orders created from winning bids.

### Steps

1. **Payment Method Selection**
   - Choose from available payment options
   - Bank transfer, digital wallets, or credit cards
   - Verify payment method details
   - Review payment terms and conditions

2. **Payment Initiation**
   - POST `/api/v1/payments/initiate`
   - Submit payment details and order information
   - System generates payment reference
   - Integration with payment gateway

3. **Payment Processing**
   - Real-time payment status monitoring
   - Handle payment confirmations and failures
   - Retry mechanisms for failed payments
   - Secure handling of payment credentials

4. **Payment Confirmation**
   - Receive payment confirmation notification
   - Order status updates to PAID
   - Seller receives payment notification
   - Fulfillment process is triggered

5. **Payment History**
   - Track all payment transactions
   - Download payment receipts
   - Review payment method usage
   - Monitor refund status if applicable

### API Endpoints

- `POST /api/v1/payments/initiate`
- `GET /api/v1/payments/{payment_id}/status`
- `GET /api/v1/payments/my-payments`
- `POST /api/v1/payments/{payment_id}/retry`

### Success Criteria

- Secure payment processing
- Multiple payment method support
- Real-time status updates
- Proper error handling and recovery
- Complete payment audit trail

---

## Error Handling & Edge Cases

### Common Error Scenarios

1. **Bid Validation Errors**
   - Bid below minimum amount
   - Bidding on expired auctions
   - Insufficient funds validation
   - Bidding on own listings

2. **System Errors**
   - Network connectivity issues
   - Database transaction failures
   - Payment gateway errors
   - Authentication/authorization failures

3. **Business Rule Violations**
   - Inventory shortfalls
   - Auction rule violations
   - Organization policy restrictions
   - Concurrent bidding conflicts

### Error Response Format

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "User-friendly error message",
    "details": {
      "field": "specific validation error"
    }
  },
  "meta": {
    "trace_id": "unique_trace_id",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

## Performance Requirements

- API response times under 500ms
- Support for 100+ concurrent bidders per auction
- Real-time notification delivery under 2 seconds
- Database query optimization for bid history
- Efficient caching for frequently accessed data

## Security Considerations

- JWT token validation for all requests
- Organization-based access controls
- Secure payment processing
- Audit trails for all bid activities
- Rate limiting to prevent abuse
- Input validation and sanitization
