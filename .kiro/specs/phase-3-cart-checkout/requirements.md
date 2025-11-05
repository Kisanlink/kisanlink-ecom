# Phase 3: Cart & Checkout - Requirements Specification

## Overview

The Cart & Checkout system provides a comprehensive shopping experience for the KisanLink agricultural e-commerce platform. It manages shopping carts with real-time inventory checks, implements price locking mechanisms, handles inventory reservations, and ensures atomic cart-to-order conversion with high concurrency support.

## Functional Requirements

### 1. Cart Management

**FR-1.1: Cart Creation & Lifecycle**
- System SHALL create carts for both guest and authenticated users
- System SHALL generate unique cart tokens for identification
- System SHALL maintain cart state across sessions
- System SHALL auto-expire carts after 7 days of inactivity
- System SHALL support cart merging when guest converts to authenticated user

**FR-1.2: Cart Operations**
- System SHALL allow adding products with quantity validation
- System SHALL support updating item quantities
- System SHALL allow removing items from cart
- System SHALL calculate totals automatically (subtotal, tax, shipping, total)
- System SHALL validate product availability in real-time

**FR-1.3: Cart Persistence**
- System SHALL store cart data in hybrid storage (Redis + PostgreSQL)
- System SHALL sync Redis cache with database
- System SHALL recover carts from database if cache fails
- System SHALL maintain cart history for analytics

### 2. Price Management

**FR-2.1: Price Locking**
- System SHALL lock prices for 30-minute windows
- System SHALL generate unique lock tokens
- System SHALL maintain locked prices during checkout
- System SHALL show price lock expiration to users
- System SHALL allow price lock extension during payment

**FR-2.2: Price Validation**
- System SHALL validate current prices before locking
- System SHALL detect price changes and notify users
- System SHALL honor locked prices during checkout
- System SHALL recalculate if lock expires

**FR-2.3: Dynamic Pricing**
- System SHALL apply FPO-specific pricing
- System SHALL calculate delivery costs per location
- System SHALL apply applicable taxes (CGST, SGST, IGST)
- System SHALL support discount codes and promotions

### 3. Inventory Management

**FR-3.1: Availability Checking**
- System SHALL check real-time inventory before adding to cart
- System SHALL validate availability on cart updates
- System SHALL show stock levels to users
- System SHALL handle out-of-stock scenarios gracefully

**FR-3.2: Inventory Reservation**
- System SHALL reserve inventory during checkout initiation
- System SHALL maintain 15-minute reservation window
- System SHALL release expired reservations automatically
- System SHALL confirm reservations on order completion

**FR-3.3: Concurrency Control**
- System SHALL prevent overselling with atomic operations
- System SHALL handle concurrent checkout attempts
- System SHALL implement optimistic locking for updates
- System SHALL queue conflicting operations

### 4. Checkout Process

**FR-4.1: Checkout Flow**
- System SHALL implement multi-step checkout (address → shipping → payment → confirmation)
- System SHALL save progress at each step
- System SHALL validate cart contents before checkout
- System SHALL create checkout sessions with unique tokens

**FR-4.2: Address Management**
- System SHALL collect billing and shipping addresses
- System SHALL validate address formats
- System SHALL save addresses for future use (authenticated users)
- System SHALL calculate shipping based on address

**FR-4.3: Shipping Options**
- System SHALL provide multiple shipping methods
- System SHALL calculate delivery estimates
- System SHALL show shipping costs upfront
- System SHALL validate delivery feasibility

**FR-4.4: Payment Integration**
- System SHALL integrate with payment gateway
- System SHALL support multiple payment methods
- System SHALL handle payment failures gracefully
- System SHALL maintain payment audit trail

### 5. Order Conversion

**FR-5.1: Cart to Order**
- System SHALL atomically convert cart to order
- System SHALL maintain data consistency during conversion
- System SHALL transfer all cart items to order items
- System SHALL preserve locked prices in order

**FR-5.2: Post-Conversion**
- System SHALL mark cart as converted
- System SHALL confirm inventory reservations
- System SHALL clear cart from user session
- System SHALL trigger order confirmation flow

### 6. Cart Recovery

**FR-6.1: Abandoned Cart Detection**
- System SHALL identify carts inactive for 2+ hours
- System SHALL track abandonment patterns
- System SHALL exclude converted/expired carts
- System SHALL prioritize high-value carts

**FR-6.2: Recovery Campaigns**
- System SHALL send recovery emails (max 3 per cart)
- System SHALL offer incentives (discounts) for recovery
- System SHALL track recovery success rate
- System SHALL support SMS/push notifications

## Non-Functional Requirements

### 1. Performance

**NFR-1.1: Response Times**
- Cart operations SHALL complete within 100ms
- Price calculations SHALL complete within 50ms
- Checkout initiation SHALL complete within 500ms
- Order conversion SHALL complete within 2 seconds

**NFR-1.2: Throughput**
- System SHALL handle 1000 cart operations per second
- System SHALL support 50,000 concurrent active carts
- System SHALL process 100 checkouts per minute
- System SHALL scale horizontally for peak loads

### 2. Reliability

**NFR-2.1: Availability**
- System SHALL maintain 99.9% uptime
- System SHALL implement circuit breakers for external services
- System SHALL provide graceful degradation
- System SHALL maintain cart data during failures

**NFR-2.2: Data Consistency**
- System SHALL ensure ACID properties for critical operations
- System SHALL prevent dirty reads during updates
- System SHALL maintain referential integrity
- System SHALL implement saga pattern for distributed transactions

### 3. Scalability

**NFR-3.1: Horizontal Scaling**
- System SHALL support Redis cluster for caching
- System SHALL implement database read replicas
- System SHALL use connection pooling
- System SHALL support auto-scaling

**NFR-3.2: Load Handling**
- System SHALL handle 10x normal load during sales
- System SHALL implement rate limiting
- System SHALL queue non-critical operations
- System SHALL prioritize checkout operations

### 4. Security

**NFR-4.1: Data Protection**
- System SHALL encrypt sensitive cart data
- System SHALL validate all price calculations server-side
- System SHALL prevent cart manipulation
- System SHALL implement CSRF protection

**NFR-4.2: Access Control**
- System SHALL enforce cart ownership validation
- System SHALL prevent unauthorized cart access
- System SHALL audit cart operations
- System SHALL rotate cart tokens on authentication

## Technical Requirements

### 1. Storage Architecture

**TR-1.1: Hybrid Storage**
```go
// Redis for active carts (hot data)
- cart:{id} - Hash with cart details
- cart:items:{id} - Sorted set of items
- cart:user:{user_id} - User's cart IDs
- cart:locks:{id} - Price lock data

// PostgreSQL for persistence (cold data)
- carts table - Cart master data
- cart_items table - Cart line items
- price_locks table - Price guarantees
- inventory_reservations table - Stock holds
```

**TR-1.2: Synchronization Strategy**
- Write-through cache for cart updates
- Async sync for non-critical data
- Database as source of truth
- Cache TTL of 24 hours

### 2. Concurrency Management

**TR-2.1: Locking Strategies**
```go
// Optimistic locking for cart updates
type CartItem struct {
    Version int `gorm:"version"`
    // ... other fields
}

// Distributed locking for checkout
func ProcessCheckout(cartID string) error {
    lock := acquireDistributedLock(cartID)
    defer lock.Release()
    // ... process checkout
}
```

**TR-2.2: Transaction Isolation**
- Use READ COMMITTED for cart reads
- Use SERIALIZABLE for checkout conversion
- Implement retry logic for conflicts
- Use advisory locks for critical sections

### 3. API Design

**TR-3.1: RESTful Endpoints**
```yaml
# Cart Operations
POST   /api/v1/carts                      # Create cart
GET    /api/v1/carts/{token}              # Get cart
POST   /api/v1/carts/{token}/items        # Add item
PUT    /api/v1/carts/{token}/items/{id}   # Update item
DELETE /api/v1/carts/{token}/items/{id}   # Remove item
POST   /api/v1/carts/{token}/clear        # Clear cart
POST   /api/v1/carts/{token}/merge        # Merge carts

# Price Management
POST   /api/v1/carts/{token}/lock-prices  # Lock prices
GET    /api/v1/carts/{token}/price-lock   # Get lock status

# Checkout
POST   /api/v1/checkout/sessions          # Start checkout
PUT    /api/v1/checkout/{id}/address      # Set address
PUT    /api/v1/checkout/{id}/shipping     # Select shipping
POST   /api/v1/checkout/{id}/reserve      # Reserve inventory
POST   /api/v1/checkout/{id}/payment      # Process payment
POST   /api/v1/checkout/{id}/confirm      # Complete order
```

**TR-3.2: Event Publishing**
```yaml
cart.item.added         # Item added to cart
cart.item.removed       # Item removed from cart
cart.abandoned          # Cart marked abandoned
cart.recovered          # Abandoned cart recovered
price.locked            # Prices locked for cart
inventory.reserved      # Stock reserved for checkout
order.created           # Cart converted to order
```

### 4. Database Schema

**TR-4.1: Core Tables**
- carts - Shopping cart master
- cart_items - Cart line items
- price_locks - Price guarantees
- inventory_reservations - Stock holds
- checkout_sessions - Checkout state
- abandoned_cart_reminders - Recovery tracking

**TR-4.2: Indexes**
- Cart lookup by token (unique)
- User carts (user_id, status)
- Abandoned carts (last_activity, status)
- Price lock expiry (expires_at)
- Reservation expiry (expires_at)

## Acceptance Criteria

### AC-1: Cart Operations
- GIVEN a user adds an item to cart
- WHEN the product is available
- THEN item is added with current price
- AND cart total is recalculated
- AND change is reflected in <100ms

### AC-2: Price Locking
- GIVEN a user initiates checkout
- WHEN prices are locked
- THEN prices remain fixed for 30 minutes
- AND lock token is generated
- AND user sees countdown timer

### AC-3: Inventory Reservation
- GIVEN a user proceeds to payment
- WHEN inventory is available
- THEN stock is reserved for 15 minutes
- AND other users cannot purchase reserved items
- AND reservation expires if not confirmed

### AC-4: Concurrent Checkout
- GIVEN multiple users checkout same product
- WHEN limited inventory exists
- THEN first to complete gets the product
- AND others receive out-of-stock message
- AND no overselling occurs

### AC-5: Cart Recovery
- GIVEN a cart is abandoned for 2 hours
- WHEN recovery email is sent
- THEN email contains cart contents
- AND includes discount code
- AND link redirects to saved cart

## Dependencies

1. **External Services**
   - Redis for caching
   - Payment gateway for transactions
   - Email service for notifications

2. **Internal Services**
   - Product catalog service
   - Inventory management service
   - Order management service
   - Pricing service
   - User authentication service

3. **Infrastructure**
   - Redis Cluster/Sentinel
   - PostgreSQL with replicas
   - Message queue for events

## Risks & Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Redis failure | High | Low | Fallback to database, Redis Sentinel |
| Overselling | High | Medium | Atomic operations, reservations |
| Price inconsistency | High | Low | Server-side validation, locking |
| Cart abandonment | Medium | High | Recovery campaigns, persistent carts |
| Payment failures | High | Medium | Retry logic, multiple providers |
| Concurrent updates | Medium | High | Optimistic locking, queue system |

## Success Metrics

1. **Performance Metrics**
   - <100ms average cart operation time
   - <2s checkout completion time
   - 99.9% uptime

2. **Business Metrics**
   - <30% cart abandonment rate
   - >5% recovery success rate
   - >70% checkout completion rate

3. **Quality Metrics**
   - Zero overselling incidents
   - <0.1% price discrepancy issues
   - <1% checkout failures

## Implementation Timeline

- Week 1-2: Database schema and models
- Week 2-3: Redis integration and caching
- Week 3-4: Cart CRUD operations
- Week 4-5: Price locking mechanism
- Week 5-6: Inventory reservation system
- Week 6-7: Checkout flow implementation
- Week 7-8: Order conversion logic
- Week 8-9: Cart recovery system
- Week 9-10: Performance optimization
- Week 10-11: Integration testing
- Week 11-12: UAT and deployment

## Testing Strategy

1. **Unit Tests**
   - Cart operation logic
   - Price calculations
   - Inventory checks

2. **Integration Tests**
   - Redis-Database sync
   - Checkout flow
   - Payment integration

3. **Load Tests**
   - 1000 concurrent carts
   - 100 simultaneous checkouts
   - Cache failover scenarios

4. **Chaos Testing**
   - Redis failure simulation
   - Network partitions
   - Database slowdowns
