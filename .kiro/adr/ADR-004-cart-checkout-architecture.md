# ADR-004: Cart Management and Checkout Architecture

**Status**: Proposed
**Date**: 2025-11-04
**Author**: SDE-3 Backend Architect

## Context

KisanLink e-commerce platform requires a robust cart and checkout system that:
- Manages shopping carts with real-time inventory checks
- Implements price locking for 30-minute windows
- Handles inventory reservation during checkout
- Converts carts to orders atomically
- Supports high concurrency during peak seasons
- Enables abandoned cart recovery

## Problem Statement

1. Maintain cart state across sessions with high availability
2. Lock prices during active shopping to prevent surprises
3. Reserve inventory to prevent overselling
4. Handle concurrent checkouts for same products
5. Recover from payment failures gracefully
6. Support guest checkout and authenticated users
7. Enable cart sharing between devices

## Decision Drivers

1. **Performance**: Sub-100ms cart operations
2. **Consistency**: Prevent race conditions and overselling
3. **Scalability**: Support 50K concurrent carts
4. **Reliability**: Zero cart loss, graceful failure handling
5. **User Experience**: Seamless checkout flow
6. **Cost**: Optimize infrastructure costs

## Considered Options

### Cart Storage Strategy

#### Option 1: Redis-Only Storage
Store carts entirely in Redis with persistence.

**Pros:**
- Ultra-fast operations (<5ms)
- Built-in TTL for expiration
- Atomic operations
- Easy horizontal scaling

**Cons:**
- Risk of data loss on failure
- Limited querying capabilities
- Higher memory costs
- Complex backup strategy

#### Option 2: Database-Only Storage
Store carts in PostgreSQL.

**Pros:**
- ACID guarantees
- Rich querying capabilities
- Consistent with order data
- Easy backup and recovery

**Cons:**
- Higher latency (20-50ms)
- Database load for frequent updates
- Connection pool pressure

#### Option 3: Hybrid (Redis + Database)
Use Redis for active carts, PostgreSQL for persistence.

**Pros:**
- Best of both worlds
- Fast reads from Redis
- Durability from PostgreSQL
- Flexible querying

**Cons:**
- Sync complexity
- Dual-write scenarios
- Cache invalidation challenges

### Price Locking Mechanism

#### Option 1: Timestamp-Based Locking
Store price with timestamp, validate on checkout.

**Pros:**
- Simple implementation
- No background jobs needed
- Minimal storage overhead

**Cons:**
- Price can change between operations
- No guarantee for user

#### Option 2: Reservation System
Create price reservations with expiry.

**Pros:**
- Guaranteed price for duration
- Auditable price locks
- Clear user communication

**Cons:**
- Additional table/storage
- Cleanup job needed
- Complex state management

## Decision

**Adopt Hybrid Storage with Reservation System** for optimal performance and reliability:

### Database Schema

```sql
-- Shopping cart master table
CREATE TABLE carts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Cart identification
    cart_token VARCHAR(100) NOT NULL UNIQUE, -- For guest users
    user_id VARCHAR(255), -- NULL for guest carts
    organization_id VARCHAR(255),
    session_id VARCHAR(255),

    -- Cart state
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- 'active', 'abandoned', 'converted', 'expired', 'merged'

    -- Financial summary (cached for performance)
    subtotal_amount DECIMAL(12,2) DEFAULT 0,
    tax_amount DECIMAL(12,2) DEFAULT 0,
    discount_amount DECIMAL(12,2) DEFAULT 0,
    shipping_amount DECIMAL(12,2) DEFAULT 0,
    total_amount DECIMAL(12,2) DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'INR',

    -- Price lock information
    price_locked_at TIMESTAMP,
    price_lock_expires_at TIMESTAMP,
    price_lock_token VARCHAR(100),

    -- Cart metadata
    source VARCHAR(50), -- 'web', 'mobile', 'api'
    ip_address INET,
    user_agent TEXT,
    referrer TEXT,
    utm_params JSONB,

    -- Conversion tracking
    converted_to_order_id UUID,
    converted_at TIMESTAMP,

    -- Expiry management
    expires_at TIMESTAMP NOT NULL DEFAULT (NOW() + INTERVAL '7 days'),
    last_activity_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Metadata
    notes TEXT,
    metadata JSONB,

    -- Audit fields
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    -- Indexes and constraints
    CONSTRAINT check_price_lock CHECK (
        (price_locked_at IS NULL AND price_lock_expires_at IS NULL) OR
        (price_locked_at IS NOT NULL AND price_lock_expires_at IS NOT NULL)
    )
);

-- Cart performance indexes
CREATE INDEX idx_carts_user_id ON carts(user_id) WHERE deleted_at IS NULL AND status = 'active';
CREATE INDEX idx_carts_token ON carts(cart_token) WHERE deleted_at IS NULL;
CREATE INDEX idx_carts_session ON carts(session_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_carts_expiry ON carts(expires_at) WHERE deleted_at IS NULL AND status = 'active';
CREATE INDEX idx_carts_abandoned ON carts(last_activity_at) WHERE deleted_at IS NULL AND status = 'active' AND user_id IS NOT NULL;
CREATE INDEX idx_carts_price_lock_expiry ON carts(price_lock_expires_at) WHERE price_lock_expires_at IS NOT NULL;

-- Cart items table
CREATE TABLE cart_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id UUID NOT NULL REFERENCES carts(id) ON DELETE CASCADE,

    -- Product information
    catalog_item_id UUID NOT NULL,
    catalog_item_type VARCHAR(20) NOT NULL, -- 'product', 'service', 'labour'
    variant_id UUID,

    -- Snapshot of product details at add time
    product_name VARCHAR(500) NOT NULL,
    product_sku VARCHAR(100),
    product_image_url TEXT,
    seller_org_id VARCHAR(255) NOT NULL,

    -- Quantity and pricing
    quantity DECIMAL(10,3) NOT NULL,
    unit_of_measure VARCHAR(20) NOT NULL,

    -- Price at time of adding to cart
    original_unit_price DECIMAL(12,2) NOT NULL,
    original_total_price DECIMAL(12,2) NOT NULL,

    -- Current/locked price
    current_unit_price DECIMAL(12,2) NOT NULL,
    current_total_price DECIMAL(12,2) NOT NULL,

    -- Price lock details
    price_locked BOOLEAN DEFAULT FALSE,
    locked_unit_price DECIMAL(12,2),
    locked_total_price DECIMAL(12,2),
    price_locked_until TIMESTAMP,

    -- Tax and discounts
    tax_rate DECIMAL(5,2) DEFAULT 0,
    tax_amount DECIMAL(10,2) DEFAULT 0,
    discount_rate DECIMAL(5,2) DEFAULT 0,
    discount_amount DECIMAL(10,2) DEFAULT 0,

    -- Inventory reservation
    inventory_reserved BOOLEAN DEFAULT FALSE,
    reservation_id UUID,
    reserved_until TIMESTAMP,

    -- Item metadata
    customization JSONB, -- Custom options selected
    notes TEXT,
    metadata JSONB,

    -- Timestamps
    added_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT unique_cart_product UNIQUE(cart_id, catalog_item_id, variant_id),
    CONSTRAINT positive_quantity CHECK (quantity > 0),
    CONSTRAINT valid_prices CHECK (current_unit_price >= 0 AND current_total_price >= 0)
);

CREATE INDEX idx_cart_items_cart_id ON cart_items(cart_id);
CREATE INDEX idx_cart_items_product ON cart_items(catalog_item_id);
CREATE INDEX idx_cart_items_reservation ON cart_items(reservation_id) WHERE inventory_reserved = true;

-- Price locks table for guaranteed pricing
CREATE TABLE price_locks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Lock identification
    lock_token VARCHAR(100) NOT NULL UNIQUE,
    cart_id UUID REFERENCES carts(id) ON DELETE CASCADE,

    -- Locked prices
    locked_items JSONB NOT NULL, -- Array of {item_id, quantity, unit_price, total_price}

    -- Lock details
    locked_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    released_at TIMESTAMP,

    -- Financial summary at lock time
    subtotal_amount DECIMAL(12,2) NOT NULL,
    tax_amount DECIMAL(12,2) NOT NULL,
    discount_amount DECIMAL(12,2) NOT NULL,
    shipping_amount DECIMAL(12,2) NOT NULL,
    total_amount DECIMAL(12,2) NOT NULL,

    -- Usage tracking
    used_for_order_id UUID,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT check_expiry CHECK (expires_at > locked_at)
);

CREATE INDEX idx_price_locks_token ON price_locks(lock_token) WHERE released_at IS NULL;
CREATE INDEX idx_price_locks_cart ON price_locks(cart_id) WHERE released_at IS NULL;
CREATE INDEX idx_price_locks_expiry ON price_locks(expires_at) WHERE released_at IS NULL AND used_for_order_id IS NULL;

-- Inventory reservations during checkout
CREATE TABLE inventory_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Reservation details
    reservation_token VARCHAR(100) NOT NULL UNIQUE,
    cart_id UUID REFERENCES carts(id),
    cart_item_id UUID REFERENCES cart_items(id),

    -- Product and inventory
    catalog_item_id UUID NOT NULL,
    inventory_lot_id UUID,
    quantity DECIMAL(10,3) NOT NULL,

    -- Reservation lifecycle
    reserved_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    released_at TIMESTAMP,
    confirmed_at TIMESTAMP,

    -- Order reference if converted
    order_id UUID,
    order_item_id UUID,

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- 'active', 'expired', 'released', 'confirmed'

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT check_status_consistency CHECK (
        (status = 'active' AND released_at IS NULL AND confirmed_at IS NULL) OR
        (status = 'expired' AND expires_at < NOW()) OR
        (status = 'released' AND released_at IS NOT NULL) OR
        (status = 'confirmed' AND confirmed_at IS NOT NULL)
    )
);

CREATE INDEX idx_inventory_reservations_cart ON inventory_reservations(cart_id) WHERE status = 'active';
CREATE INDEX idx_inventory_reservations_product ON inventory_reservations(catalog_item_id) WHERE status = 'active';
CREATE INDEX idx_inventory_reservations_expiry ON inventory_reservations(expires_at) WHERE status = 'active';
CREATE INDEX idx_inventory_reservations_token ON inventory_reservations(reservation_token);

-- Checkout sessions for payment processing
CREATE TABLE checkout_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Session identification
    session_token VARCHAR(100) NOT NULL UNIQUE,
    cart_id UUID NOT NULL REFERENCES carts(id),

    -- Checkout state
    status VARCHAR(20) NOT NULL DEFAULT 'initiated', -- 'initiated', 'processing', 'completed', 'failed', 'expired'
    step VARCHAR(50), -- 'address', 'shipping', 'payment', 'confirmation'

    -- Addresses
    billing_address JSONB,
    shipping_address JSONB,

    -- Shipping method
    shipping_method VARCHAR(50),
    shipping_provider VARCHAR(100),
    shipping_cost DECIMAL(10,2),
    estimated_delivery_date DATE,

    -- Payment information
    payment_method VARCHAR(50),
    payment_provider VARCHAR(100),
    payment_intent_id VARCHAR(255),
    payment_status VARCHAR(50),

    -- Price lock reference
    price_lock_token VARCHAR(100),

    -- Inventory reservations
    inventory_reserved BOOLEAN DEFAULT FALSE,
    reservation_tokens JSONB, -- Array of reservation tokens

    -- Final amounts
    subtotal_amount DECIMAL(12,2),
    tax_amount DECIMAL(12,2),
    discount_amount DECIMAL(12,2),
    shipping_amount DECIMAL(12,2),
    total_amount DECIMAL(12,2),

    -- Conversion
    order_id UUID,
    completed_at TIMESTAMP,

    -- Session management
    expires_at TIMESTAMP NOT NULL DEFAULT (NOW() + INTERVAL '1 hour'),
    last_activity_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Metadata
    ip_address INET,
    user_agent TEXT,
    metadata JSONB,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_checkout_sessions_token ON checkout_sessions(session_token);
CREATE INDEX idx_checkout_sessions_cart ON checkout_sessions(cart_id);
CREATE INDEX idx_checkout_sessions_status ON checkout_sessions(status) WHERE status IN ('initiated', 'processing');
CREATE INDEX idx_checkout_sessions_expiry ON checkout_sessions(expires_at) WHERE status NOT IN ('completed', 'failed');

-- Abandoned cart recovery
CREATE TABLE abandoned_cart_reminders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id UUID NOT NULL REFERENCES carts(id),

    reminder_number INTEGER NOT NULL, -- 1st, 2nd, 3rd reminder
    sent_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Communication details
    channel VARCHAR(20) NOT NULL, -- 'email', 'sms', 'push'
    recipient VARCHAR(255) NOT NULL,

    -- Tracking
    opened_at TIMESTAMP,
    clicked_at TIMESTAMP,
    converted_at TIMESTAMP,

    -- Recovery offer
    discount_code VARCHAR(50),
    discount_amount DECIMAL(10,2),

    metadata JSONB,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_cart_reminder UNIQUE(cart_id, reminder_number)
);

CREATE INDEX idx_abandoned_reminders_cart ON abandoned_cart_reminders(cart_id);
CREATE INDEX idx_abandoned_reminders_sent ON abandoned_cart_reminders(sent_at DESC);
```

### Redis Schema for Active Carts

```go
// Redis key structure
type RedisCartKeys struct {
    // cart:{cart_id} - Hash containing cart data
    CartData string `redis:"cart:{cart_id}"`

    // cart:items:{cart_id} - Sorted set of items by added_at
    CartItems string `redis:"cart:items:{cart_id}"`

    // cart:user:{user_id} - Set of cart IDs for user
    UserCarts string `redis:"cart:user:{user_id}"`

    // cart:session:{session_id} - Current cart for session
    SessionCart string `redis:"cart:session:{session_id}"`

    // cart:locks:{cart_id} - Price lock data
    PriceLock string `redis:"cart:locks:{cart_id}"`

    // cart:reservations:{cart_id} - Set of reservation tokens
    Reservations string `redis:"cart:reservations:{cart_id}"`
}

// Cart structure in Redis
type RedisCart struct {
    ID              string    `redis:"id"`
    UserID          string    `redis:"user_id"`
    Status          string    `redis:"status"`
    SubtotalAmount  float64   `redis:"subtotal_amount"`
    TaxAmount       float64   `redis:"tax_amount"`
    DiscountAmount  float64   `redis:"discount_amount"`
    ShippingAmount  float64   `redis:"shipping_amount"`
    TotalAmount     float64   `redis:"total_amount"`
    ItemCount       int       `redis:"item_count"`
    LastActivityAt  time.Time `redis:"last_activity_at"`
    PriceLocked     bool      `redis:"price_locked"`
    PriceLockExpiry time.Time `redis:"price_lock_expiry"`
}
```

### Service Architecture

```go
// Cart Service with hybrid storage
type CartService struct {
    cartRepo        CartRepository
    productService  ProductService
    priceService    PriceService
    inventoryService InventoryService
    redisClient     *redis.Client
    eventBus        EventBus
    logger          Logger
}

// Core cart operations
func (s *CartService) AddToCart(ctx context.Context, req AddToCartRequest) (*Cart, error) {
    // 1. Validate product availability
    product, err := s.productService.GetProduct(ctx, req.ProductID)
    if err != nil {
        return nil, err
    }

    // 2. Check inventory
    available, err := s.inventoryService.CheckAvailability(ctx, req.ProductID, req.Quantity)
    if !available {
        return nil, ErrInsufficientInventory
    }

    // 3. Get or create cart
    cart, err := s.GetOrCreateCart(ctx, req.CartToken, req.UserID)
    if err != nil {
        return nil, err
    }

    // 4. Calculate current price
    price, err := s.priceService.CalculatePrice(ctx, product, req.Quantity, cart.OrganizationID)
    if err != nil {
        return nil, err
    }

    // 5. Add item to cart (Redis + DB)
    err = s.withTransaction(ctx, func(tx Transaction) error {
        // Add to database
        item := &CartItem{
            CartID:           cart.ID,
            CatalogItemID:    product.ID,
            Quantity:         req.Quantity,
            CurrentUnitPrice: price.UnitPrice,
            CurrentTotalPrice: price.TotalPrice,
        }

        if err := s.cartRepo.AddItem(ctx, tx, item); err != nil {
            return err
        }

        // Update Redis cache
        if err := s.updateRedisCart(ctx, cart); err != nil {
            // Log but don't fail - DB is source of truth
            s.logger.Warn("Failed to update Redis", "error", err)
        }

        return nil
    })

    // 6. Publish event
    s.eventBus.Publish(CartItemAddedEvent{
        CartID:    cart.ID,
        ProductID: req.ProductID,
        Quantity:  req.Quantity,
    })

    return cart, nil
}

// Price locking implementation
func (s *CartService) LockPrices(ctx context.Context, cartID string) (*PriceLock, error) {
    // 1. Get cart with items
    cart, err := s.GetCart(ctx, cartID)
    if err != nil {
        return nil, err
    }

    // 2. Calculate current prices for all items
    var lockedItems []LockedItem
    totalAmount := decimal.Zero

    for _, item := range cart.Items {
        price, err := s.priceService.GetCurrentPrice(ctx, item.CatalogItemID, item.Quantity)
        if err != nil {
            return nil, err
        }

        lockedItems = append(lockedItems, LockedItem{
            ItemID:     item.ID,
            Quantity:   item.Quantity,
            UnitPrice:  price.UnitPrice,
            TotalPrice: price.TotalPrice,
        })

        totalAmount = totalAmount.Add(price.TotalPrice)
    }

    // 3. Create price lock
    priceLock := &PriceLock{
        LockToken:    generateToken(),
        CartID:       cartID,
        LockedItems:  lockedItems,
        TotalAmount:  totalAmount,
        ExpiresAt:    time.Now().Add(30 * time.Minute),
    }

    // 4. Save to database and Redis
    err = s.withTransaction(ctx, func(tx Transaction) error {
        // Save price lock
        if err := s.cartRepo.CreatePriceLock(ctx, tx, priceLock); err != nil {
            return err
        }

        // Update cart
        cart.PriceLockedAt = &priceLock.LockedAt
        cart.PriceLockExpiresAt = &priceLock.ExpiresAt
        cart.PriceLockToken = priceLock.LockToken

        if err := s.cartRepo.UpdateCart(ctx, tx, cart); err != nil {
            return err
        }

        // Update items with locked prices
        for _, lockedItem := range lockedItems {
            if err := s.cartRepo.UpdateItemPrice(ctx, tx, lockedItem); err != nil {
                return err
            }
        }

        return nil
    })

    // 5. Set Redis TTL
    s.redisClient.Set(ctx, fmt.Sprintf("cart:locks:%s", cartID), priceLock, 30*time.Minute)

    // 6. Publish event
    s.eventBus.Publish(PricesLockedEvent{
        CartID:    cartID,
        LockToken: priceLock.LockToken,
        ExpiresAt: priceLock.ExpiresAt,
    })

    return priceLock, nil
}

// Inventory reservation during checkout
func (s *CartService) ReserveInventory(ctx context.Context, cartID string) ([]string, error) {
    // 1. Get cart items
    cart, err := s.GetCart(ctx, cartID)
    if err != nil {
        return nil, err
    }

    // 2. Create reservations for each item
    var reservationTokens []string

    err = s.withTransaction(ctx, func(tx Transaction) error {
        for _, item := range cart.Items {
            // Check availability
            available, err := s.inventoryService.CheckAvailability(ctx, item.CatalogItemID, item.Quantity)
            if !available {
                return fmt.Errorf("insufficient inventory for %s", item.ProductName)
            }

            // Create reservation
            reservation := &InventoryReservation{
                ReservationToken: generateToken(),
                CartID:           cartID,
                CartItemID:       item.ID,
                CatalogItemID:    item.CatalogItemID,
                Quantity:         item.Quantity,
                ExpiresAt:        time.Now().Add(15 * time.Minute),
            }

            if err := s.inventoryService.CreateReservation(ctx, tx, reservation); err != nil {
                return err
            }

            reservationTokens = append(reservationTokens, reservation.ReservationToken)

            // Update cart item
            item.InventoryReserved = true
            item.ReservationID = reservation.ID
            item.ReservedUntil = reservation.ExpiresAt

            if err := s.cartRepo.UpdateItem(ctx, tx, item); err != nil {
                return err
            }
        }

        return nil
    })

    if err != nil {
        // Rollback any reservations made
        s.releaseReservations(ctx, reservationTokens)
        return nil, err
    }

    // 3. Update Redis
    for _, token := range reservationTokens {
        s.redisClient.SAdd(ctx, fmt.Sprintf("cart:reservations:%s", cartID), token)
        s.redisClient.Expire(ctx, fmt.Sprintf("reservation:%s", token), 15*time.Minute)
    }

    return reservationTokens, nil
}

// Convert cart to order
func (s *CartService) ConvertToOrder(ctx context.Context, checkoutSession *CheckoutSession) (*Order, error) {
    // 1. Validate checkout session
    if checkoutSession.Status != "processing" {
        return nil, ErrInvalidCheckoutStatus
    }

    // 2. Validate price lock
    priceLock, err := s.validatePriceLock(ctx, checkoutSession.PriceLockToken)
    if err != nil {
        return nil, err
    }

    // 3. Validate inventory reservations
    if err := s.validateReservations(ctx, checkoutSession.ReservationTokens); err != nil {
        return nil, err
    }

    // 4. Create order atomically
    var order *Order

    err = s.withTransaction(ctx, func(tx Transaction) error {
        // Get cart
        cart, err := s.cartRepo.GetCart(ctx, tx, checkoutSession.CartID)
        if err != nil {
            return err
        }

        // Create order
        order = &Order{
            BuyerOrganizationID:  cart.OrganizationID,
            BuyerUserID:          cart.UserID,
            SubtotalAmount:       priceLock.SubtotalAmount,
            TaxAmount:            priceLock.TaxAmount,
            DiscountAmount:       priceLock.DiscountAmount,
            ShippingAmount:       priceLock.ShippingAmount,
            TotalAmount:          priceLock.TotalAmount,
            ShippingAddress:      checkoutSession.ShippingAddress,
            EstimatedDeliveryDate: checkoutSession.EstimatedDeliveryDate,
        }

        if err := s.orderService.CreateOrder(ctx, tx, order); err != nil {
            return err
        }

        // Create order items from cart items
        for _, cartItem := range cart.Items {
            orderItem := &OrderItem{
                OrderID:         order.ID,
                CatalogItemID:   cartItem.CatalogItemID,
                CatalogItemType: cartItem.CatalogItemType,
                CatalogItemName: cartItem.ProductName,
                CatalogItemSKU:  cartItem.ProductSKU,
                Quantity:        cartItem.Quantity,
                UnitPrice:       cartItem.LockedUnitPrice,
                TotalPrice:      cartItem.LockedTotalPrice,
            }

            if err := s.orderService.AddOrderItem(ctx, tx, orderItem); err != nil {
                return err
            }
        }

        // Confirm inventory reservations
        for _, token := range checkoutSession.ReservationTokens {
            if err := s.inventoryService.ConfirmReservation(ctx, tx, token, order.ID); err != nil {
                return err
            }
        }

        // Mark cart as converted
        cart.Status = "converted"
        cart.ConvertedToOrderID = order.ID
        cart.ConvertedAt = time.Now()

        if err := s.cartRepo.UpdateCart(ctx, tx, cart); err != nil {
            return err
        }

        // Mark price lock as used
        priceLock.UsedForOrderID = order.ID
        if err := s.cartRepo.UpdatePriceLock(ctx, tx, priceLock); err != nil {
            return err
        }

        // Update checkout session
        checkoutSession.Status = "completed"
        checkoutSession.OrderID = order.ID
        checkoutSession.CompletedAt = time.Now()

        if err := s.cartRepo.UpdateCheckoutSession(ctx, tx, checkoutSession); err != nil {
            return err
        }

        return nil
    })

    if err != nil {
        return nil, err
    }

    // 5. Clear Redis cache
    s.clearCartFromRedis(ctx, checkoutSession.CartID)

    // 6. Publish events
    s.eventBus.Publish(OrderCreatedFromCartEvent{
        CartID:  checkoutSession.CartID,
        OrderID: order.ID,
        UserID:  order.BuyerUserID,
    })

    return order, nil
}

// Abandoned cart recovery
func (s *CartService) RecoverAbandonedCarts(ctx context.Context) error {
    // 1. Find abandoned carts (no activity for 2 hours)
    abandonedCarts, err := s.cartRepo.FindAbandonedCarts(ctx, 2*time.Hour)
    if err != nil {
        return err
    }

    for _, cart := range abandonedCarts {
        // 2. Check if reminder already sent
        reminderCount, err := s.cartRepo.GetReminderCount(ctx, cart.ID)
        if err != nil {
            continue
        }

        if reminderCount >= 3 {
            continue // Max reminders reached
        }

        // 3. Generate recovery offer
        discountCode := s.generateRecoveryDiscount(cart, reminderCount+1)

        // 4. Send reminder
        reminder := &AbandonedCartReminder{
            CartID:         cart.ID,
            ReminderNumber: reminderCount + 1,
            Channel:        "email",
            Recipient:      cart.UserEmail,
            DiscountCode:   discountCode,
        }

        if err := s.sendAbandonedCartReminder(ctx, reminder); err != nil {
            s.logger.Error("Failed to send reminder", "error", err, "cart_id", cart.ID)
            continue
        }

        // 5. Record reminder
        if err := s.cartRepo.CreateReminder(ctx, reminder); err != nil {
            s.logger.Error("Failed to record reminder", "error", err)
        }
    }

    return nil
}
```

### Concurrency Control

```go
// Optimistic locking for cart updates
type CartUpdateService struct {
    repo CartRepository
}

func (s *CartUpdateService) UpdateCartItem(ctx context.Context, cartID, itemID string, quantity decimal.Decimal) error {
    maxRetries := 3

    for i := 0; i < maxRetries; i++ {
        // Get current version
        item, err := s.repo.GetCartItem(ctx, itemID)
        if err != nil {
            return err
        }

        currentVersion := item.Version

        // Update with version check
        item.Quantity = quantity
        item.Version = currentVersion + 1

        affected, err := s.repo.UpdateCartItemWithVersion(ctx, item, currentVersion)
        if err != nil {
            return err
        }

        if affected > 0 {
            return nil // Success
        }

        // Version conflict, retry
        time.Sleep(time.Millisecond * time.Duration(math.Pow(2, float64(i)) * 100))
    }

    return ErrConcurrentModification
}

// Distributed locking for checkout
type CheckoutService struct {
    redlock *redlock.Client
}

func (s *CheckoutService) ProcessCheckout(ctx context.Context, cartID string) error {
    // Acquire distributed lock
    lock, err := s.redlock.Lock(ctx, fmt.Sprintf("checkout:%s", cartID), 30*time.Second)
    if err != nil {
        return ErrCheckoutInProgress
    }
    defer lock.Unlock(ctx)

    // Process checkout with exclusive access
    return s.processCheckoutInternal(ctx, cartID)
}
```

### API Endpoints

```yaml
# Cart Management
POST   /api/v1/carts                    # Create new cart
GET    /api/v1/carts/{token}            # Get cart by token
POST   /api/v1/carts/{token}/items      # Add item to cart
PUT    /api/v1/carts/{token}/items/{id} # Update item quantity
DELETE /api/v1/carts/{token}/items/{id} # Remove item from cart
POST   /api/v1/carts/{token}/clear      # Clear cart
POST   /api/v1/carts/{token}/merge      # Merge guest cart with user cart

# Price Management
POST   /api/v1/carts/{token}/lock-prices    # Lock prices for 30 minutes
GET    /api/v1/carts/{token}/price-lock     # Get current price lock
DELETE /api/v1/carts/{token}/price-lock     # Release price lock

# Checkout Process
POST   /api/v1/checkout/sessions              # Initialize checkout
PUT    /api/v1/checkout/sessions/{id}/address # Update shipping address
PUT    /api/v1/checkout/sessions/{id}/shipping # Select shipping method
POST   /api/v1/checkout/sessions/{id}/reserve # Reserve inventory
POST   /api/v1/checkout/sessions/{id}/payment # Process payment
POST   /api/v1/checkout/sessions/{id}/confirm # Confirm and create order

# Cart Recovery
GET    /api/v1/carts/abandoned         # List abandoned carts (admin)
POST   /api/v1/carts/{token}/recover   # Recover abandoned cart
```

## Performance Optimizations

1. **Redis Caching Strategy**
   ```go
   // Write-through cache for active carts
   func (s *CartService) UpdateCart(ctx context.Context, cart *Cart) error {
       // Update database first
       if err := s.cartRepo.UpdateCart(ctx, cart); err != nil {
           return err
       }

       // Update Redis cache
       key := fmt.Sprintf("cart:%s", cart.ID)
       if err := s.redisClient.HSet(ctx, key, cart).Err(); err != nil {
           // Log but don't fail
           s.logger.Warn("Cache update failed", "error", err)
       }

       // Set TTL
       s.redisClient.Expire(ctx, key, 24*time.Hour)

       return nil
   }
   ```

2. **Batch Operations**
   ```go
   // Batch price calculations
   func (s *PriceService) CalculateBulkPrices(ctx context.Context, items []CartItem) ([]Price, error) {
       // Single query for all products
       products := s.productRepo.GetByIDs(ctx, extractProductIDs(items))

       // Parallel price calculation
       prices := make([]Price, len(items))
       var wg sync.WaitGroup

       for i, item := range items {
           wg.Add(1)
           go func(idx int, itm CartItem) {
               defer wg.Done()
               prices[idx] = s.calculatePrice(products[itm.ProductID], itm.Quantity)
           }(i, item)
       }

       wg.Wait()
       return prices, nil
   }
   ```

3. **Database Query Optimization**
   ```sql
   -- Materialized view for cart summaries
   CREATE MATERIALIZED VIEW cart_summaries AS
   SELECT
       c.id,
       c.user_id,
       c.status,
       COUNT(ci.id) as item_count,
       SUM(ci.quantity) as total_quantity,
       SUM(ci.current_total_price) as subtotal,
       c.updated_at
   FROM carts c
   LEFT JOIN cart_items ci ON ci.cart_id = c.id
   WHERE c.deleted_at IS NULL
   GROUP BY c.id;

   CREATE INDEX idx_cart_summaries_user ON cart_summaries(user_id);
   ```

## Security Considerations

1. **Cart Token Security**
   - Use cryptographically secure random tokens
   - Rotate tokens on user authentication
   - Validate token ownership

2. **Price Manipulation Prevention**
   - Never trust client-side prices
   - Validate all prices server-side
   - Log price discrepancies

3. **Inventory Protection**
   - Atomic inventory operations
   - Reservation expiry cleanup
   - Rate limiting on cart operations

4. **Session Security**
   - HTTPS-only cookies
   - CSRF protection
   - Session timeout enforcement

## Monitoring & Observability

```go
// Key metrics
- cart_operations_total{operation, status}
- cart_conversion_rate
- cart_abandonment_rate
- price_lock_duration_seconds
- inventory_reservation_conflicts_total
- checkout_duration_seconds
- cart_value_distribution
- items_per_cart_distribution
```

## Migration Strategy

1. **Phase 1**: Deploy cart tables and basic CRUD
2. **Phase 2**: Implement Redis caching layer
3. **Phase 3**: Enable price locking mechanism
4. **Phase 4**: Add inventory reservation system
5. **Phase 5**: Implement checkout flow
6. **Phase 6**: Enable abandoned cart recovery

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Redis failure | High | Fallback to database, Redis Sentinel for HA |
| Price lock expiry during payment | Medium | Extend lock during payment, honor expired lock with warning |
| Inventory overselling | High | Pessimistic locking, reservation system, monitoring |
| Cart abandonment | Medium | Recovery emails, persistent carts, incentives |
| Concurrent checkout | High | Distributed locking, idempotency keys |

## Consequences

### Positive
- Fast cart operations with Redis caching
- Guaranteed pricing with lock mechanism
- Prevented overselling with reservations
- Improved conversion with cart recovery
- Scalable architecture for growth

### Negative
- Complex dual-storage management
- Additional infrastructure (Redis)
- Cleanup job requirements
- Higher operational complexity

## References

- Redis Best Practices: https://redis.io/docs/
- E-commerce Cart Patterns: https://www.enterpriseintegrationpatterns.com/
- Distributed Locking: https://martin.kleppmann.com/2016/02/08/how-to-do-distributed-locking.html
- PostgreSQL Advisory Locks: https://www.postgresql.org/docs/current/explicit-locking.html
