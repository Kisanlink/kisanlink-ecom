# Payment Screenshot Implementation Status

**Date**: 2025-11-04
**Feature**: Payment Screenshot Upload and Verification
**Status**: 85% Complete - Requires Integration

## Completed Work

### 1. Domain Model (COMPLETED)
- Created `entities/models/orders/payment_screenshot.go`
- Added to model registry in `entities/models/registry.go`
- Added indexes in `internal/database/migrations.go`
- Includes:
  - VerificationStatus enum (pending, approved, rejected, disputed)
  - PaymentMethod enum (bank_transfer, upi, cheque, cash, other)
  - Complete model with S3 file metadata
  - Helper methods (IsApproved, IsRejected, IsPending, CanVerify)

### 2. DTOs (COMPLETED)
- **Requests**: `entities/requests/orders/payment_screenshot_requests.go`
  - UploadPaymentScreenshotRequest
  - VerifyPaymentScreenshotRequest
  - ListPaymentScreenshotsRequest
  - PaymentScreenshotFilters with ToPaymentScreenshotFilters converter

- **Responses**: `entities/responses/orders/payment_screenshot_responses.go`
  - PaymentScreenshotResponse
  - PaymentScreenshotUploadResponse
  - PaymentScreenshotDownloadURLResponse
  - PaymentScreenshotListResponse
  - Transformation functions

### 3. Repository Layer (NEEDS FIX)
- Created `internal/repositories/orders/payment_screenshot_repository.go`
- **Issue**: Needs update to match kisanlink-db base repository API
- Methods implemented:
  - CreatePaymentScreenshot
  - GetPaymentScreenshot
  - GetByOrderID
  - ListPaymentScreenshots (needs filter API fix)
  - GetPendingScreenshots
  - UpdateVerificationStatus
  - DeletePaymentScreenshot
  - GetScreenshotsByBuyerOrganization
  - GetScreenshotsBySellerOrganization
  - CheckDuplicateChecksum

**Required Fix**:
```go
// Current (incorrect):
screenshot := &orders.PaymentScreenshot{}
if err := r.GetByID(ctx, id, screenshot); err != nil

// Correct:
screenshot := &orders.PaymentScreenshot{}
retrievedScreenshot, err := r.GetByID(ctx, id, screenshot)
if err != nil {
    return nil, err
}
screenshot = retrievedScreenshot
```

### 4. Service Layer (COMPLETED)
- Created `internal/services/orders/payment_screenshot_service.go`
- Full S3 integration using kisanlink-db S3Manager
- Methods:
  - UploadPaymentScreenshot (with file validation, S3 upload, checksum)
  - VerifyPaymentScreenshot (admin only)
  - GetPaymentScreenshot (with authorization)
  - ListPaymentScreenshots (with filtering)
  - GetPaymentScreenshotsByOrder
  - GenerateDownloadURL (presigned URLs, 24hr expiry)
  - DeletePaymentScreenshot
  - ValidateAmountMatch (helper)
- Security features:
  - File type validation (JPEG, PNG, WebP, PDF)
  - File size limit (10MB)
  - SHA256 checksum
  - Duplicate detection
  - Organization ownership checks

### 5. Handler Layer (COMPLETED)
- Created `internal/handlers/orders/payment_screenshot_handler.go`
- Endpoints:
  - POST /api/v1/orders/{order_id}/payment-screenshots (buyer upload)
  - GET /api/v1/orders/{order_id}/payment-screenshots (list by order)
  - GET /api/v1/payment-screenshots/{screenshot_id} (get details)
  - GET /api/v1/payment-screenshots/{screenshot_id}/download (presigned URL)
  - GET /api/v1/admin/payment-screenshots (admin list all)
  - POST /api/v1/admin/payment-screenshots/{screenshot_id}/verify (admin verify)
- Authorization:
  - Buyer endpoints check organization ownership
  - Admin endpoints check IsAdmin()
  - Proper error handling and response formatting
- Swagger documentation annotations

### 6. Configuration (COMPLETED)
- Added S3Config to `internal/config/config.go`:
  ```go
  type S3Config struct {
      Bucket string
      Region string
  }
  ```
- Environment variables:
  - S3_BUCKET (default: kisanlink-ecom-files)
  - S3_REGION (default: us-east-1)

## Remaining Work

### 1. Repository API Fixes (CRITICAL)
File: `internal/repositories/orders/payment_screenshot_repository.go`

**Issue**: Methods using BaseFilterableRepository need to match actual API

**Required Changes**:
1. Fix GetByID usage (line 60):
```go
screenshot := &orders.PaymentScreenshot{}
retrievedScreenshot, err := r.GetByID(ctx, id, screenshot)
if err != nil {
    return nil, fmt.Errorf("failed to get payment screenshot: %w", err)
}
return retrievedScreenshot, nil
```

2. Fix Find usage (lines 75-90):
```go
// Check the actual API signature in order_repository.go:196-207
// Use base.NewFilter() and add conditions appropriately
filter := base.NewFilter()
filter.Group.Conditions = []base.FilterCondition{
    {
        Field:    "order_id",
        Operator: base.OpEqual,
        Value:    orderID,
    },
}

screenshotList, err := r.Find(ctx, filter)
if err != nil {
    return nil, fmt.Errorf("failed to find payment screenshots: %w", err)
}

// Convert interface{} slice to typed slice
screenshots := make([]*orders.PaymentScreenshot, len(screenshotList))
for i, item := range screenshotList {
    if screenshot, ok := item.(*orders.PaymentScreenshot); ok {
        screenshots[i] = screenshot
    }
}
return screenshots, nil
```

3. Fix ListPaymentScreenshots filter construction (lines 95-200):
   - Use base.NewFilter()
   - Add conditions using base.OpEqual, base.OpGreaterThan, base.OpLessThan
   - Check if Pagination and Sort are separate methods or part of filter

**Reference**: Study `/Users/kaushik/kisanlink-ecom/internal/repositories/orders/order_repository.go` lines 196-350 for correct patterns

### 2. S3Manager Initialization (CRITICAL)
File: `cmd/server/main.go`

Add after line 132 (after database initialization):

```go
// Initialize S3 Manager
s3Config := &db.Config{
    S3Region: cfg.S3.Region,
    S3Bucket: cfg.S3.Bucket,
}
s3Manager := db.NewS3Manager(s3Config, logrusLogger.WithField("component", "s3"))
if err := s3Manager.Connect(context.Background()); err != nil {
    log.Printf("Warning: S3 connection failed: %v", err)
    log.Printf("Payment screenshot uploads will not work")
} else {
    log.Printf("S3 manager connected successfully")
}
```

Add to repository initialization (after line 140):

```go
paymentScreenshotRepo := orders.NewPaymentScreenshotRepository(dbManager.GetManager(db.BackendGorm))
```

Add to service initialization (after line 162):

```go
paymentScreenshotSvc := orderService.NewPaymentScreenshotService(
    paymentScreenshotRepo,
    orderRepo,
    s3Manager,
    cfg.S3.Bucket,
    logrusLogger.WithField("component", "payment-screenshot"),
)
```

### 3. Route Wiring (CRITICAL)
File: `internal/routes/secure_routes.go`

Add payment screenshot handler parameter to SetupSecureRouter (line 27):

```go
func SetupSecureRouter(
    aaaClient auth.Client,
    catalogSvc catalogService.CatalogServiceInterface,
    inventorySvc inventoryService.InventoryService,
    alertSvc inventoryService.AlertService,
    orderSvc orderService.OrderServiceInterface,
    userSvc *userService.UserService,
    integrationSvc integrationService.IntegrationServiceInterface,
    marketplaceSvc *marketplaceService.MarketplaceServices,
    paymentScreenshotSvc orderService.PaymentScreenshotServiceInterface,  // ADD THIS
) *gin.Engine {
```

Add routes in the orders group (after line 248):

```go
// Payment screenshots (buyer endpoints)
ordersGroup.POST("/:order_id/payment-screenshots",
    rbacMiddleware.RequirePermission("order", "create"),
    paymentScreenshotHandler.UploadPaymentScreenshot,
)
ordersGroup.GET("/:order_id/payment-screenshots",
    rbacMiddleware.RequirePermission("order", "read"),
    paymentScreenshotHandler.ListPaymentScreenshotsByOrder,
)

// Payment screenshot details (buyers and admins)
paymentScreenshotsGroup := v1.Group("/payment-screenshots")
paymentScreenshotsGroup.Use(authMiddleware.Middleware())
{
    paymentScreenshotsGroup.GET("/:screenshot_id",
        rbacMiddleware.RequirePermission("order", "read"),
        paymentScreenshotHandler.GetPaymentScreenshot,
    )
    paymentScreenshotsGroup.GET("/:screenshot_id/download",
        rbacMiddleware.RequirePermission("order", "read"),
        paymentScreenshotHandler.GetDownloadURL,
    )
}
```

Add admin routes in adminGroup (after line 405):

```go
// Payment screenshot verification (admin only)
adminPaymentScreenshots := adminGroup.Group("/payment-screenshots")
{
    if paymentScreenshotSvc != nil {
        paymentScreenshotHandler := orders.NewPaymentScreenshotHandler(paymentScreenshotSvc, orderSvc)

        adminPaymentScreenshots.GET("",
            paymentScreenshotHandler.ListPaymentScreenshots,
        )
        adminPaymentScreenshots.POST("/:screenshot_id/verify",
            paymentScreenshotHandler.VerifyPaymentScreenshot,
        )
    }
}
```

Initialize handler (after line 223):

```go
paymentScreenshotHandler := orders.NewPaymentScreenshotHandler(paymentScreenshotSvc, orderSvc)
```

### 4. Order Status Update Logic (ALREADY IMPLEMENTED)
The handler already triggers order status updates on approval (line 450-464 in payment_screenshot_handler.go):

```go
// If approved, trigger order status update
if req.Status == orderModels.VerificationStatusApproved {
    if updateErr := h.orderService.UpdateOrderStatus(
        c.Request.Context(),
        screenshot.OrderID,
        &orderRequests.UpdateOrderStatusRequest{
            Status: orderModels.OrderStatusPaid,
            Reason: fmt.Sprintf("Payment verified via screenshot %s", screenshotID),
        },
        userID,
        screenshot.SellerOrganizationID,
    ); updateErr != nil {
        c.Header("X-Order-Update-Status", "failed")
        c.Header("X-Order-Update-Error", updateErr.Error())
    }
}
```

No additional work needed.

### 5. Unit Tests (TODO)
Create test files:
- `internal/repositories/orders/payment_screenshot_repository_test.go`
- `internal/services/orders/payment_screenshot_service_test.go`
- `internal/handlers/orders/payment_screenshot_handler_test.go`

Test coverage requirements:
- Repository: CRUD operations, filtering, authorization
- Service: Upload workflow, verification, S3 integration, error handling
- Handler: Request validation, authorization, response formatting

### 6. Environment Setup (TODO)
Add to `.env`:
```
S3_BUCKET=kisanlink-ecom-files
S3_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
```

### 7. Documentation (TODO)
- Update Swagger docs: `go generate ./...`
- Add API examples to README
- Document S3 bucket setup requirements
- Add deployment checklist

## Build Status

Current status: **WILL NOT BUILD** due to repository API mismatches

Expected errors:
1. `assignment mismatch: 1 variable but r.GetByID returns 2 values`
2. `undefined: base.GroupAnd`
3. `undefined: base.OpGreaterThanOrEqual`
4. `too many arguments in call to r.Find`

## Next Steps

1. **Fix repository** (30 minutes):
   - Update GetByID calls
   - Fix Find API usage
   - Update filter construction

2. **Wire up services** (20 minutes):
   - Initialize S3Manager in main.go
   - Add payment screenshot service
   - Update routes

3. **Test build** (10 minutes):
   ```bash
   go build ./cmd/server
   ```

4. **Test endpoints** (30 minutes):
   - Upload screenshot
   - List screenshots
   - Verify screenshot
   - Download screenshot
   - Check order status update

5. **Write tests** (2 hours):
   - Unit tests for all layers
   - Integration tests for upload flow

## API Endpoints Summary

### Buyer Endpoints
- `POST /api/v1/orders/{order_id}/payment-screenshots` - Upload
- `GET /api/v1/orders/{order_id}/payment-screenshots` - List by order
- `GET /api/v1/payment-screenshots/{screenshot_id}` - Get details
- `GET /api/v1/payment-screenshots/{screenshot_id}/download` - Download

### Admin Endpoints
- `GET /api/v1/admin/payment-screenshots` - List all with filters
- `POST /api/v1/admin/payment-screenshots/{screenshot_id}/verify` - Verify/approve/reject

## Security Checklist

- [x] File type validation (JPEG, PNG, WebP, PDF only)
- [x] File size limit (10MB)
- [x] Organization ownership checks
- [x] Admin-only verification
- [x] SHA256 checksums
- [x] Duplicate detection
- [x] Presigned URLs with expiry
- [x] Soft delete support
- [ ] Rate limiting on uploads (future)
- [ ] Virus scanning (future)

## Performance Considerations

- File uploads go directly to S3 (not stored in DB)
- Presigned URLs avoid proxying through server
- Indexed queries on common filters
- Pagination support for large result sets
- Efficient duplicate detection using checksums

## Known Limitations

1. No automatic OCR or payment verification
2. Manual admin verification required
3. No integration with payment gateways
4. Single file per upload (no batch)
5. No file preview generation

## Future Enhancements

1. OCR integration for automatic field extraction
2. Integration with payment gateways
3. Automated verification rules
4. Email notifications on verification
5. Batch upload support
6. Thumbnail generation
7. Payment reconciliation dashboard
