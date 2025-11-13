# ADR-003: Invoice Generation and Document Management Architecture

**Status**: Proposed
**Date**: 2025-11-04
**Author**: SDE-3 Backend Architect

## Context

KisanLink e-commerce platform requires a robust invoice generation system to:
- Generate PDF invoices for completed orders
- Store invoices securely in S3 with CDN distribution
- Send invoices via email to buyers and sellers
- Maintain compliance with Indian GST regulations
- Support bulk invoice generation for reconciliation

## Problem Statement

1. Generate professional PDF invoices with proper formatting
2. Handle high-volume invoice generation (10K+ daily)
3. Ensure invoice immutability and audit compliance
4. Support multiple invoice formats (Tax Invoice, Proforma, Credit Note)
5. Enable async generation to avoid blocking operations
6. Maintain invoice number sequencing per organization

## Decision Drivers

1. **Performance**: Generate invoices without blocking order completion
2. **Reliability**: Zero invoice loss, guaranteed generation
3. **Compliance**: GST-compliant format with proper tax breakdowns
4. **Storage**: Cost-effective long-term storage with quick retrieval
5. **Scalability**: Handle peak loads during harvest seasons
6. **Security**: Encrypted storage, signed URLs for access

## Considered Options

### PDF Generation Engine

#### Option 1: wkhtmltopdf
HTML to PDF conversion using WebKit rendering engine.

**Pros:**
- High-quality rendering
- Supports complex HTML/CSS
- Battle-tested in production

**Cons:**
- Requires system binary
- Memory intensive
- Slower for complex layouts

#### Option 2: go-pdf (gopdf/gofpdf)
Native Go PDF generation library.

**Pros:**
- Pure Go, no external dependencies
- Fast generation
- Low memory footprint
- Better control over PDF structure

**Cons:**
- More complex for rich layouts
- Limited HTML/CSS support

#### Option 3: Chromium Headless (CDP)
Use Chrome DevTools Protocol for rendering.

**Pros:**
- Perfect HTML/CSS compatibility
- Modern web standards support

**Cons:**
- Heavy resource usage
- Complex deployment
- Overkill for invoices

### Storage Strategy

#### Option 1: S3 + CloudFront
Store in S3, serve via CloudFront CDN.

**Pros:**
- Highly available
- Cost-effective
- Global distribution
- Signed URLs for security

**Cons:**
- Additional CDN costs
- Complexity in URL management

#### Option 2: Database BLOB Storage
Store PDFs in PostgreSQL.

**Pros:**
- Transactional consistency
- Single system to manage

**Cons:**
- Database bloat
- Poor performance
- Expensive storage

## Decision

**Adopt go-pdf for generation and S3 + CloudFront for storage** with the following architecture:

### Database Schema

```sql
-- Invoice master table
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Invoice identifiers
    invoice_number VARCHAR(50) NOT NULL,
    invoice_type VARCHAR(20) NOT NULL, -- 'tax_invoice', 'proforma', 'credit_note', 'debit_note'
    invoice_date DATE NOT NULL,

    -- Order reference
    order_id UUID NOT NULL REFERENCES orders(id),
    order_number VARCHAR(50) NOT NULL,

    -- Organization details
    seller_org_id VARCHAR(255) NOT NULL,
    seller_org_name VARCHAR(255) NOT NULL,
    seller_gstin VARCHAR(15),
    seller_pan VARCHAR(10),

    buyer_org_id VARCHAR(255) NOT NULL,
    buyer_org_name VARCHAR(255) NOT NULL,
    buyer_gstin VARCHAR(15),
    buyer_pan VARCHAR(10),

    -- Financial details
    subtotal_amount DECIMAL(12,2) NOT NULL,
    tax_amount DECIMAL(12,2) NOT NULL,
    cgst_amount DECIMAL(12,2) DEFAULT 0,
    sgst_amount DECIMAL(12,2) DEFAULT 0,
    igst_amount DECIMAL(12,2) DEFAULT 0,
    discount_amount DECIMAL(12,2) DEFAULT 0,
    shipping_amount DECIMAL(12,2) DEFAULT 0,
    total_amount DECIMAL(12,2) NOT NULL,

    amount_in_words TEXT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'INR',

    -- Document details
    pdf_generated BOOLEAN DEFAULT FALSE,
    pdf_generation_status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'processing', 'completed', 'failed'
    pdf_generated_at TIMESTAMP,
    pdf_s3_key VARCHAR(500),
    pdf_s3_bucket VARCHAR(255),
    pdf_cdn_url TEXT,
    pdf_size_bytes INTEGER,
    pdf_checksum VARCHAR(64), -- SHA256

    -- Email status
    email_sent BOOLEAN DEFAULT FALSE,
    email_sent_at TIMESTAMP,
    email_recipients JSONB, -- [{type: 'buyer'|'seller', email: '', sent_at: '', status: ''}]

    -- Address details
    billing_address JSONB NOT NULL,
    shipping_address JSONB NOT NULL,

    -- Payment details
    payment_terms VARCHAR(255),
    payment_due_date DATE,
    payment_status VARCHAR(20), -- 'pending', 'partial', 'paid'

    -- Digital signature (optional)
    digital_signature JSONB, -- {signed_by: '', signed_at: '', certificate: '', signature: ''}

    -- Metadata
    notes TEXT,
    terms_and_conditions TEXT,
    metadata JSONB,

    -- Audit fields
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255) NOT NULL,
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),

    -- Constraints
    CONSTRAINT unique_invoice_number UNIQUE(seller_org_id, invoice_number) WHERE deleted_at IS NULL,
    CONSTRAINT valid_invoice_type CHECK (invoice_type IN ('tax_invoice', 'proforma', 'credit_note', 'debit_note'))
);

-- Indexes for performance
CREATE INDEX idx_invoices_order_id ON invoices(order_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_seller_org ON invoices(seller_org_id, invoice_date DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_buyer_org ON invoices(buyer_org_id, invoice_date DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_generation_status ON invoices(pdf_generation_status) WHERE deleted_at IS NULL AND pdf_generated = FALSE;
CREATE INDEX idx_invoices_number_search ON invoices(invoice_number) WHERE deleted_at IS NULL;

-- Invoice line items
CREATE TABLE invoice_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,

    -- Item details
    item_sequence INTEGER NOT NULL,
    product_id VARCHAR(255) NOT NULL,
    product_name VARCHAR(500) NOT NULL,
    product_sku VARCHAR(100),
    hsn_code VARCHAR(10), -- HSN/SAC code for GST

    -- Quantity and pricing
    quantity DECIMAL(10,3) NOT NULL,
    unit_of_measure VARCHAR(20) NOT NULL,
    unit_price DECIMAL(12,2) NOT NULL,
    gross_amount DECIMAL(12,2) NOT NULL,

    -- Tax details
    tax_rate DECIMAL(5,2) DEFAULT 0,
    cgst_rate DECIMAL(5,2) DEFAULT 0,
    sgst_rate DECIMAL(5,2) DEFAULT 0,
    igst_rate DECIMAL(5,2) DEFAULT 0,
    cgst_amount DECIMAL(10,2) DEFAULT 0,
    sgst_amount DECIMAL(10,2) DEFAULT 0,
    igst_amount DECIMAL(10,2) DEFAULT 0,

    -- Discounts
    discount_rate DECIMAL(5,2) DEFAULT 0,
    discount_amount DECIMAL(10,2) DEFAULT 0,

    net_amount DECIMAL(12,2) NOT NULL,

    -- Metadata
    batch_number VARCHAR(50),
    expiry_date DATE,
    metadata JSONB,

    -- Audit fields
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_invoice_item_sequence UNIQUE(invoice_id, item_sequence)
);

CREATE INDEX idx_invoice_items_invoice_id ON invoice_items(invoice_id);

-- Invoice number sequences per organization
CREATE TABLE invoice_sequences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id VARCHAR(255) NOT NULL,
    sequence_type VARCHAR(50) NOT NULL, -- 'invoice', 'credit_note', 'debit_note'
    prefix VARCHAR(20) NOT NULL,
    current_number INTEGER NOT NULL DEFAULT 0,
    suffix VARCHAR(20),
    fiscal_year VARCHAR(10) NOT NULL,

    -- Format template: {prefix}/{year}/{number:06d}/{suffix}
    format_template VARCHAR(100) NOT NULL DEFAULT '{prefix}/{year}/{number:06d}',

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_org_sequence UNIQUE(organization_id, sequence_type, fiscal_year)
);

CREATE INDEX idx_invoice_sequences_org ON invoice_sequences(organization_id, sequence_type);

-- Invoice generation queue for async processing
CREATE TABLE invoice_generation_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices(id),
    priority INTEGER DEFAULT 5, -- 1-10, 1 being highest
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'processing', 'completed', 'failed'
    attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    error_message TEXT,

    scheduled_at TIMESTAMP DEFAULT NOW(),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invoice_queue_status ON invoice_generation_queue(status, priority, scheduled_at) WHERE status IN ('pending', 'failed');
```

### Service Architecture

```go
// Invoice Service Layer
type InvoiceService struct {
    repo            InvoiceRepository
    orderRepo       OrderRepository
    pdfGenerator    PDFGenerator
    storageService  StorageService
    emailService    EmailService
    sequenceService SequenceService
    queueService    QueueService
    eventBus        EventBus
}

// Core invoice generation flow
func (s *InvoiceService) GenerateInvoice(ctx context.Context, orderID string) (*Invoice, error) {
    // 1. Validate order is in completed state
    // 2. Check if invoice already exists
    // 3. Generate invoice number
    // 4. Create invoice record
    // 5. Queue for PDF generation
    // 6. Publish InvoiceCreatedEvent
    return invoice, nil
}

// Async PDF generation worker
func (s *InvoiceService) ProcessPDFGenerationQueue(ctx context.Context) {
    for {
        // 1. Fetch pending invoices from queue
        // 2. Generate PDF using template
        // 3. Upload to S3
        // 4. Update invoice record with S3 details
        // 5. Generate CloudFront signed URL
        // 6. Queue for email delivery
        // 7. Publish InvoicePDFGeneratedEvent
    }
}
```

### PDF Generation Implementation

```go
// Using gofpdf for generation
type PDFGenerator interface {
    GenerateInvoice(invoice *Invoice, items []InvoiceItem) ([]byte, error)
}

type GoPDFGenerator struct {
    templatePath string
    fontPath     string
}

func (g *GoPDFGenerator) GenerateInvoice(invoice *Invoice, items []InvoiceItem) ([]byte, error) {
    pdf := gofpdf.New("P", "mm", "A4", "")

    // Add company logo
    pdf.Image("logo.png", 10, 10, 40, 0, false, "", 0, "")

    // Add header with invoice details
    pdf.SetFont("Arial", "B", 16)
    pdf.Cell(0, 10, "TAX INVOICE", "", 1, "C")

    // Add invoice number, date
    pdf.SetFont("Arial", "", 10)
    pdf.Cell(0, 5, fmt.Sprintf("Invoice No: %s", invoice.InvoiceNumber), "", 1, "L")
    pdf.Cell(0, 5, fmt.Sprintf("Date: %s", invoice.InvoiceDate.Format("02-Jan-2006")), "", 1, "L")

    // Add seller and buyer details in columns
    // Add item table with HSN codes, tax breakup
    // Add total section with amount in words
    // Add terms and conditions
    // Add digital signature if present

    var buf bytes.Buffer
    pdf.Output(&buf)
    return buf.Bytes(), nil
}
```

### S3 Storage Strategy

```go
type StorageService interface {
    UploadInvoice(ctx context.Context, invoice *Invoice, pdfData []byte) (*StorageResult, error)
    GetSignedURL(ctx context.Context, s3Key string, expiry time.Duration) (string, error)
}

type S3StorageService struct {
    s3Client *s3.Client
    bucket   string
    cdnDomain string
}

// S3 Key Structure: invoices/{year}/{month}/{org_id}/{invoice_number}.pdf
func (s *S3StorageService) UploadInvoice(ctx context.Context, invoice *Invoice, pdfData []byte) (*StorageResult, error) {
    key := fmt.Sprintf("invoices/%d/%02d/%s/%s.pdf",
        invoice.InvoiceDate.Year(),
        invoice.InvoiceDate.Month(),
        invoice.SellerOrgID,
        invoice.InvoiceNumber)

    // Calculate checksum
    checksum := sha256.Sum256(pdfData)

    // Upload with server-side encryption
    _, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
        Bucket:               aws.String(s.bucket),
        Key:                  aws.String(key),
        Body:                 bytes.NewReader(pdfData),
        ContentType:          aws.String("application/pdf"),
        ServerSideEncryption: types.ServerSideEncryptionAes256,
        Metadata: map[string]string{
            "invoice-id":     invoice.ID,
            "invoice-number": invoice.InvoiceNumber,
            "order-id":       invoice.OrderID,
        },
    })

    return &StorageResult{
        S3Key:    key,
        S3Bucket: s.bucket,
        CDNUrl:   fmt.Sprintf("https://%s/%s", s.cdnDomain, key),
        Size:     len(pdfData),
        Checksum: hex.EncodeToString(checksum[:]),
    }, nil
}
```

### Email Integration

```go
type EmailService interface {
    SendInvoiceEmail(ctx context.Context, invoice *Invoice, recipients []string) error
}

type SESEmailService struct {
    sesClient *ses.Client
    fromEmail string
}

func (s *SESEmailService) SendInvoiceEmail(ctx context.Context, invoice *Invoice, recipients []string) error {
    // Generate email body with invoice summary
    subject := fmt.Sprintf("Invoice %s for Order %s", invoice.InvoiceNumber, invoice.OrderNumber)

    // HTML email template with invoice summary
    htmlBody := s.renderInvoiceEmailTemplate(invoice)

    // Send email with PDF attachment link
    _, err := s.sesClient.SendEmail(ctx, &ses.SendEmailInput{
        Source: aws.String(s.fromEmail),
        Destination: &types.Destination{
            ToAddresses: recipients,
        },
        Message: &types.Message{
            Subject: &types.Content{Data: aws.String(subject)},
            Body: &types.Body{
                Html: &types.Content{Data: aws.String(htmlBody)},
            },
        },
    })

    return err
}
```

### Invoice Number Generation

```go
type SequenceService interface {
    GetNextInvoiceNumber(ctx context.Context, orgID string, invoiceType string) (string, error)
}

type InvoiceSequenceService struct {
    repo SequenceRepository
}

func (s *InvoiceSequenceService) GetNextInvoiceNumber(ctx context.Context, orgID string, invoiceType string) (string, error) {
    // Get or create sequence for current fiscal year
    fiscalYear := s.getCurrentFiscalYear()

    sequence, err := s.repo.GetOrCreateSequence(ctx, orgID, invoiceType, fiscalYear)
    if err != nil {
        return "", err
    }

    // Increment with database lock
    nextNumber, err := s.repo.IncrementSequence(ctx, sequence.ID)
    if err != nil {
        return "", err
    }

    // Format invoice number
    invoiceNumber := s.formatInvoiceNumber(sequence, nextNumber)
    return invoiceNumber, nil
}

func (s *InvoiceSequenceService) formatInvoiceNumber(seq *InvoiceSequence, number int) string {
    // Example: INV/2024-25/000001 or ORG001/24-25/INV/000001
    replacer := strings.NewReplacer(
        "{prefix}", seq.Prefix,
        "{year}", seq.FiscalYear,
        "{number:06d}", fmt.Sprintf("%06d", number),
        "{suffix}", seq.Suffix,
    )
    return replacer.Replace(seq.FormatTemplate)
}
```

## API Endpoints

```yaml
# Invoice Management APIs
POST   /api/v1/invoices/generate        # Generate invoice for order
GET    /api/v1/invoices/{id}           # Get invoice details
GET    /api/v1/invoices                # List invoices with filters
GET    /api/v1/invoices/{id}/download  # Download invoice PDF
POST   /api/v1/invoices/{id}/resend    # Resend invoice email
POST   /api/v1/invoices/bulk-generate  # Bulk invoice generation
GET    /api/v1/invoices/export         # Export invoices (CSV/Excel)

# Credit Note Management
POST   /api/v1/credit-notes            # Create credit note
GET    /api/v1/credit-notes/{id}       # Get credit note details
```

## Security Considerations

1. **Access Control**
   - Only authorized users can view invoices
   - Organization-level access restrictions
   - Role-based permissions (view, generate, export)

2. **Data Protection**
   - Server-side encryption for S3 storage
   - Signed URLs with expiration (24 hours)
   - PII data masking in logs

3. **Audit Trail**
   - Log all invoice access attempts
   - Track invoice generation and modifications
   - Maintain compliance with GST regulations

4. **Input Validation**
   - Validate GST numbers format
   - Verify tax calculations
   - Prevent invoice number manipulation

## Performance Optimizations

1. **Async Processing**
   - Queue-based PDF generation
   - Batch processing for bulk invoices
   - Priority queue for urgent invoices

2. **Caching Strategy**
   ```go
   // Cache invoice PDFs for quick retrieval
   type InvoiceCache struct {
       redis *redis.Client
   }

   func (c *InvoiceCache) GetPDF(invoiceID string) ([]byte, error) {
       key := fmt.Sprintf("invoice:pdf:%s", invoiceID)
       return c.redis.Get(ctx, key).Bytes()
   }

   func (c *InvoiceCache) SetPDF(invoiceID string, data []byte) error {
       key := fmt.Sprintf("invoice:pdf:%s", invoiceID)
       return c.redis.Set(ctx, key, data, 24*time.Hour).Err()
   }
   ```

3. **Database Optimizations**
   - Partial indexes for active invoices
   - Materialized views for reporting
   - Partitioning by fiscal year for large datasets

## Monitoring & Observability

```go
// Key metrics to track
- invoice_generation_duration_seconds
- invoice_generation_errors_total
- invoice_pdf_size_bytes
- invoice_email_delivery_rate
- invoice_s3_upload_duration_seconds
- invoice_queue_depth
- invoice_sequence_conflicts_total
```

## Migration Strategy

1. **Phase 1**: Deploy invoice tables and basic generation
2. **Phase 2**: Integrate PDF generation and S3 storage
3. **Phase 3**: Enable email delivery
4. **Phase 4**: Migrate historical orders to generate invoices
5. **Phase 5**: Enable bulk operations and reporting

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| PDF generation failure | High | Retry mechanism, fallback HTML view |
| S3 unavailability | Medium | Local disk cache, multi-region setup |
| Invoice number conflicts | High | Database-level unique constraints, advisory locks |
| Email delivery failures | Low | Retry queue, alternative channels (SMS) |
| Compliance violations | High | Regular audits, automated validation |

## Consequences

### Positive
- Automated invoice generation reduces manual effort
- Compliant with GST regulations
- Scalable architecture handles growth
- Cost-effective storage with S3
- Professional PDF invoices improve brand image

### Negative
- Additional infrastructure complexity (S3, SES)
- Dependency on external services
- Initial development effort for templates
- Ongoing maintenance of tax rules

## References

- GST Invoice Rules: https://www.gst.gov.in/
- gofpdf Documentation: https://github.com/jung-kurt/gofpdf
- AWS S3 Best Practices: https://docs.aws.amazon.com/AmazonS3/latest/userguide/
- Invoice Design Patterns: https://www.invoiced.com/resources/
