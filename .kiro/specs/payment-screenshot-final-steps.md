# Payment Screenshot Feature - Final Integration Steps

## Status: 95% Complete - Remaining Integration Only

### ✅ What's Done
- ✅ All code implemented and compiling
- ✅ Routes wired up in `secure_routes.go`
- ✅ Models registered for AutoMigrate
- ✅ Build passing: `go build ./cmd/server`

---

## 🔧 Remaining Tasks

### 1. Initialize S3Manager and PaymentScreenshotService in main.go

**Location**: `/Users/kaushik/kisanlink-ecom/cmd/server/main.go`

**Add after initializing orderService** (around line ~150-200):

```go
// Initialize S3Manager for payment screenshots
s3Config := &db.Config{
    S3Bucket: os.Getenv("AWS_S3_BUCKET"), // e.g., "kisanlink-ecom-payments"
    S3Region: os.Getenv("AWS_S3_REGION"), // e.g., "ap-south-1"
}
s3Manager := db.NewS3Manager(s3Config, logger)
if err := s3Manager.Connect(ctx); err != nil {
    logger.Fatal("Failed to connect to S3", zap.Error(err))
}

// Initialize Payment Screenshot Repository
paymentScreenshotRepo := orderRepos.NewPaymentScreenshotRepository(dbManager)

// Initialize Payment Screenshot Service
paymentScreenshotService := orderServices.NewPaymentScreenshotService(
    paymentScreenshotRepo,
    orderRepo,
    s3Manager,
    os.Getenv("AWS_S3_BUCKET"),
    logger,
)
```

**Update SetupSecureRouter call** to include the new service:

```go
router := routes.SetupSecureRouter(
    aaaClient,
    catalogService,
    inventoryService,
    alertService,
    orderService,
    paymentScreenshotService,  // ADD THIS LINE
    userService,
    integrationService,
    marketplaceServices,
)
```

---

### 2. Environment Variables

**Add to `.env` file**:

```bash
# S3 Configuration for Payment Screenshots
AWS_S3_BUCKET=kisanlink-ecom-payments
AWS_S3_REGION=ap-south-1
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key

# File Upload Configuration
MAX_UPLOAD_SIZE_MB=10
ALLOWED_MIME_TYPES=image/jpeg,image/png,image/webp,application/pdf
```

---

### 3. Run Database Migration

The model is already registered in `entities/models/registry.go`, so just run:

```bash
go run cmd/server/main.go
# The AutoMigrate will create the payment_screenshots table automatically
```

---

### 4. Test the API

#### Upload Payment Screenshot (Buyer):
```bash
curl -X POST http://localhost:8080/api/v1/orders/{order_id}/payment-screenshots \
  -H "Authorization: Bearer $BUYER_TOKEN" \
  -F "file=@payment_receipt.jpg" \
  -F "payment_method=bank_transfer" \
  -F "payment_date=2025-11-04" \
  -F "amount_paid=1000.00" \
  -F "transaction_id=TXN123456"
```

#### List Screenshots for Order (Buyer):
```bash
curl -X GET "http://localhost:8080/api/v1/orders/{order_id}/payment-screenshots" \
  -H "Authorization: Bearer $BUYER_TOKEN"
```

#### List All Pending Screenshots (Admin):
```bash
curl -X GET "http://localhost:8080/api/v1/admin/payment-screenshots?verification_status=pending" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

#### Verify Screenshot (Admin):
```bash
curl -X POST http://localhost:8080/api/v1/admin/payment-screenshots/{id}/verify \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved",
    "notes": "Payment verified successfully"
  }'
```

#### Download Screenshot:
```bash
curl -X GET "http://localhost:8080/api/v1/payment-screenshots/{id}/download" \
  -H "Authorization: Bearer $TOKEN"
```

---

## 📊 Complete API Endpoints

### Buyer Endpoints
```
POST   /api/v1/orders/:order_id/payment-screenshots      - Upload screenshot
GET    /api/v1/orders/:order_id/payment-screenshots      - List for order
GET    /api/v1/payment-screenshots/:id                   - Get details
GET    /api/v1/payment-screenshots/:id/download          - Get download URL
```

### Admin Endpoints
```
GET    /api/v1/admin/payment-screenshots                 - List all (with filters)
POST   /api/v1/admin/payment-screenshots/:id/verify     - Verify (approve/reject)
```

---

## 🧪 Testing (Optional but Recommended)

### Unit Tests Skeleton

Create test files:
- `tests/repositories/orders/payment_screenshot_repository_test.go`
- `tests/services/orders/payment_screenshot_service_test.go`
- `tests/handlers/orders/payment_screenshot_handler_test.go`

Run tests:
```bash
go test ./tests/repositories/orders/... -v
go test ./tests/services/orders/... -v
go test ./tests/handlers/orders/... -v
```

---

## 🎯 Success Criteria

- [x] Code compiles successfully
- [x] Routes wired up
- [x] Models registered
- [ ] S3Manager initialized in main.go
- [ ] PaymentScreenshotService initialized in main.go
- [ ] Environment variables configured
- [ ] Database table created (AutoMigrate)
- [ ] Manual API testing passes

---

## 📝 Notes

1. **S3 Bucket**: Create bucket `kisanlink-ecom-payments` in AWS S3 (or use existing)
2. **IAM Permissions**: Ensure AWS credentials have S3 PutObject, GetObject, DeleteObject permissions
3. **File Upload**: Max size 10MB enforced at application level
4. **Security**: All files stored with server-side encryption (AES-256)
5. **Signed URLs**: 24-hour expiry for downloads

---

## 🔍 Troubleshooting

### If build fails:
```bash
go mod tidy
go build ./cmd/server
```

### If S3 connection fails:
- Check AWS credentials in environment variables
- Verify S3 bucket exists and region is correct
- Check IAM permissions

### If AutoMigrate fails:
- Check database connection
- Verify `entities/models/registry.go` includes `&orders.PaymentScreenshot{}`

---

## 📚 Documentation

- **ADR**: `.kiro/adr/ADR-005-payment-screenshot-architecture.md`
- **API Docs**: Swagger annotations in handlers (auto-generated)
- **Code**: All implementations in:
  - Models: `entities/models/orders/payment_screenshot.go`
  - Requests: `entities/requests/orders/payment_screenshot_requests.go`
  - Responses: `entities/responses/orders/payment_screenshot_responses.go`
  - Repository: `internal/repositories/orders/payment_screenshot_repository.go`
  - Service: `internal/services/orders/payment_screenshot_service.go`
  - Handler: `internal/handlers/orders/payment_screenshot_handler.go`

---

**Implementation Time**: 2-3 hours (from scratch)
**Remaining Time**: 15-30 minutes (just initialization + testing)
