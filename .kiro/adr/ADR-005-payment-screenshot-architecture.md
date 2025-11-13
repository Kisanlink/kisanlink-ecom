# ADR-005: Payment Screenshot Upload and Verification Architecture

**Status**: Approved
**Date**: 2025-11-04
**Decision Makers**: Backend Team, SDE Manager
**Technical Story**: Implement payment screenshot upload for buyers with admin verification workflow

## Context

The e-commerce platform requires a payment verification system where:
- **Buyers** can upload payment screenshots/receipts for their orders
- **Admins** can verify these screenshots and approve/reject payments
- **Order status** should automatically update based on payment verification
- **File storage** must be secure, scalable, and cost-effective

### Requirements

1. **Buyer Functionality**:
   - Upload payment screenshots (JPEG, PNG, WebP, PDF)
   - Associate screenshots with specific orders
   - Provide payment metadata (method, date, amount, transaction ID)
   - View their uploaded screenshots

2. **Admin Functionality**:
   - View all pending payment screenshots
   - Verify screenshots (approve, reject, or mark as disputed)
   - Add verification notes
   - Trigger order status changes upon approval

3. **Security & Authorization**:
   - Buyers can only upload for their own orders
   - Buyers can only view their own screenshots
   - Only admins can verify screenshots
   - File access control via signed URLs

4. **File Storage**:
   - Support max 10MB file size
   - Store files securely with encryption
   - Generate time-limited access URLs
   - Track file metadata (size, checksum, mime type)

## Decision

We will implement a comprehensive payment screenshot management system with the following architecture:

### 1. Database Schema

```sql
CREATE TABLE payment_screenshots (
    -- Identity
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- References
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    uploaded_by VARCHAR(255) NOT NULL,
    buyer_organization_id VARCHAR(255) NOT NULL,
    seller_organization_id VARCHAR(255) NOT NULL,

    -- File Information
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,  -- S3 key
    s3_bucket VARCHAR(255) NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    file_mime_type VARCHAR(100) NOT NULL,
    file_checksum VARCHAR(64),  -- SHA256

    -- Payment Information
    payment_method VARCHAR(50),
    payment_date DATE NOT NULL,
    amount_paid DECIMAL(12,2),
    transaction_id VARCHAR(100),

    -- Verification
    verification_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    verified_by VARCHAR(255),
    verified_at TIMESTAMP,
    verification_notes TEXT,

    -- Description
    description TEXT,

    -- Audit
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255) NOT NULL,
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),

    -- Constraints
    CONSTRAINT valid_verification_status CHECK (verification_status IN ('pending', 'approved', 'rejected', 'disputed')),
    CONSTRAINT valid_mime_type CHECK (file_mime_type IN ('image/jpeg', 'image/png', 'image/webp', 'application/pdf')),
    CONSTRAINT positive_file_size CHECK (file_size_bytes > 0),
    CONSTRAINT positive_amount CHECK (amount_paid > 0)
);

-- Indexes
CREATE INDEX idx_payment_screenshots_order_id ON payment_screenshots(order_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_payment_screenshots_verification_status ON payment_screenshots(verification_status) WHERE deleted_at IS NULL;
CREATE INDEX idx_payment_screenshots_buyer_org ON payment_screenshots(buyer_organization_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_payment_screenshots_seller_org ON payment_screenshots(seller_organization_id, verification_status) WHERE deleted_at IS NULL;
```

### 2. Clean Architecture Layers

Following kisanlink-ecom standards:

```
entities/models/orders/payment_screenshot.go          - Domain model
entities/requests/orders/payment_screenshot_requests.go   - Request DTOs
entities/responses/orders/payment_screenshot_responses.go - Response DTOs

internal/repositories/orders/payment_screenshot_repository.go  - Data access
internal/services/orders/payment_screenshot_service.go         - Business logic
internal/handlers/orders/payment_screenshot_handler.go         - HTTP handlers
```

### 3. File Storage Strategy

**S3 Storage Configuration**:
- **Bucket**: `kisanlink-ecom-payments` (or configured bucket)
- **Key Structure**: `payment-screenshots/{year}/{month}/{org_id}/{order_id}/{screenshot_id}.{ext}`
- **Encryption**: Server-side encryption (AES-256)
- **Access Control**: Private bucket with signed URLs
- **URL Expiry**: 24 hours for downloads
- **Max File Size**: 10MB
- **Allowed Types**: `image/jpeg`, `image/png`, `image/webp`, `application/pdf`

**Example S3 Key**:
```
payment-screenshots/2025/11/ORG00000001/ORD1730725200/550e8400-e29b-41d4-a716-446655440000.jpg
```

### 4. API Endpoints

#### Buyer Endpoints

```
POST /api/v1/orders/{order_id}/payment-screenshots
  - Upload payment screenshot
  - Multipart form-data
  - Auth: Bearer token (buyer role)

GET /api/v1/orders/{order_id}/payment-screenshots
  - List screenshots for an order
  - Auth: Bearer token (buyer - own orders only)

GET /api/v1/payment-screenshots/{screenshot_id}
  - Get specific screenshot details
  - Auth: Bearer token (buyer - own screenshots only)

GET /api/v1/payment-screenshots/{screenshot_id}/download
  - Get signed URL to download screenshot
  - Auth: Bearer token (buyer - own screenshots only)
```

#### Admin Endpoints

```
GET /api/v1/admin/payment-screenshots?status=pending
  - List all screenshots with filters
  - Auth: Bearer token (admin role)

GET /api/v1/admin/payment-screenshots/{screenshot_id}
  - Get any screenshot details
  - Auth: Bearer token (admin role)

POST /api/v1/admin/payment-screenshots/{screenshot_id}/verify
  - Verify screenshot (approve/reject/dispute)
  - Body: { status, notes }
  - Auth: Bearer token (admin role)
```

### 5. Order Status Update Flow

When admin approves a payment screenshot:

```
Order Status: pending
  ↓
Screenshot uploaded by buyer (status: pending)
  ↓
Admin reviews and approves (status: approved)
  ↓
Order Status: pending → confirmed → paid
  ↓
OrderStatusHistory entry created
  ↓
Event published for downstream processing
```

**Status Transition Logic**:
- Only update order if status is `pending` or `confirmed`
- Create audit trail in `order_status_history` table
- Store reference to screenshot ID in status change
- Publish `PaymentVerifiedEvent` for event-driven workflows

### 6. Authorization Model

```go
// Buyer Authorization
if !isAdmin {
    // Verify order belongs to user's organization
    order, err := orderService.GetOrderByID(ctx, orderID, userID, orgID)
    if err != nil || order.BuyerOrganizationID != orgID {
        return Forbidden
    }
}

// Admin Authorization
if !common.IsAdmin(c) {
    return Forbidden
}
```

### 7. Business Logic Validations

**Upload Validations**:
- Order exists and belongs to buyer
- Order is not already completed/cancelled
- File type and size within limits
- Payment amount matches order total (optional warning)
- No duplicate screenshots (same checksum)

**Verification Validations**:
- Screenshot exists and is pending
- Admin has proper permissions
- Verification notes required for rejection
- Order can accept status transition

## Consequences

### Positive

1. **Secure File Storage**: S3 with encryption protects sensitive payment data
2. **Scalability**: S3 handles file storage, database only stores metadata
3. **Audit Trail**: Complete tracking of uploads and verifications
4. **Clean Separation**: Buyers and admins have distinct, secure workflows
5. **Automatic Status Updates**: Reduces manual intervention in order processing
6. **Standard Compliance**: Follows repository clean architecture patterns

### Negative

1. **S3 Dependency**: Requires S3 configuration and credentials
2. **Additional Complexity**: New tables, models, and business logic
3. **Storage Costs**: S3 storage costs scale with file uploads
4. **Manual Verification**: Admins must manually review each screenshot

### Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| S3 Failure | High | Implement retry logic, fallback storage, monitoring |
| Malicious Files | High | Validate MIME type, file size, scan for malware |
| Authorization Bypass | Critical | Multiple validation layers, comprehensive tests |
| Concurrent Uploads | Medium | Use idempotency keys, transaction locking |
| Data Loss | Medium | S3 versioning, encrypted backups |

## Alternatives Considered

### Alternative 1: Database BLOB Storage
**Rejected**: Does not scale, increases database load, expensive

### Alternative 2: Local File System Storage
**Rejected**: Not cloud-native, difficult to scale horizontally

### Alternative 3: Automatic OCR Verification
**Deferred**: Future enhancement, requires ML infrastructure

### Alternative 4: Third-party Payment Gateway Integration
**Deferred**: Future enhancement, requires business approval

## Implementation Plan

1. **Phase 1**: Database & Models (Days 1-2)
   - Create migration
   - Implement models and DTOs

2. **Phase 2**: Repository Layer (Days 3-4)
   - Implement payment screenshot repository
   - Write repository tests

3. **Phase 3**: Service Layer (Days 5-6)
   - Implement business logic
   - Add S3 upload/download
   - Write service tests

4. **Phase 4**: Handler Layer (Days 7-8)
   - Implement API handlers
   - Add authorization checks
   - Write handler tests

5. **Phase 5**: Integration & Testing (Days 9-10)
   - End-to-end testing
   - Performance testing

6. **Phase 6**: Documentation & Deployment (Days 11-12)
   - Update Swagger docs
   - Deployment

## References

- **ADR-003**: Invoice Generation Architecture (S3 pattern)
- **ADR-004**: Cart and Checkout Architecture (order management)
- **Repository Standards**: `.kiro/steering/structure.md`
- **Security Guidelines**: `.kiro/steering/tech.md`
- **Testing Standards**: `.kiro/steering/testing.md`

## Decision Rationale

This architecture was chosen because:

1. **Proven Pattern**: Follows existing S3 storage pattern from invoice generation
2. **Security First**: Multiple authorization layers, signed URLs, encryption
3. **Scalability**: S3 handles file storage independently
4. **Standard Compliance**: Matches repository clean architecture
5. **User Experience**: Simple upload workflow for buyers, efficient verification for admins
6. **Maintainability**: Clear separation of concerns, testable components

## Approval

- [x] Backend Team Lead
- [x] Security Review
- [x] Architecture Review

**Approved by**: SDE Manager
**Date**: 2025-11-04
