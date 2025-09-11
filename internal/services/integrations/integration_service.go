package integrations

import (
    "context"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "strconv"
    "time"

    integrationRequests "kisanlink-ecom/entities/requests/integrations"
    integrationResponses "kisanlink-ecom/entities/responses/integrations"
    catalogService "kisanlink-ecom/internal/services/catalog"
    orderService "kisanlink-ecom/internal/services/orders"

    "github.com/google/uuid"
    "github.com/shopspring/decimal"
    "github.com/sirupsen/logrus"
)

// IntegrationServiceInterface defines the interface for integration operations
type IntegrationServiceInterface interface {
    ProcessCatalogProposal(ctx context.Context, req *integrationRequests.CatalogProposalRequest, userID string, orgID string) (*integrationResponses.CatalogProposalResponse, error)
    ProcessOrderAcknowledgement(ctx context.Context, req *integrationRequests.OrderAcknowledgementRequest, userID string, orgID string) (*integrationResponses.OrderAcknowledgementResponse, error)
    ExportCatalogWithDelta(ctx context.Context, req *integrationRequests.CatalogExportRequest, userID string, orgID string) (*integrationResponses.CatalogExportResponse, error)
    ValidateWebhookSignature(ctx context.Context, req *integrationRequests.WebhookSignatureRequest, partnerID string) (*integrationResponses.WebhookValidationResponse, error)
    GetProposalStatus(ctx context.Context, proposalID string, userID string, orgID string) (map[string]interface{}, error)
    ListIntegrationPartners(ctx context.Context, page, limit int, userID string, orgID string) ([]map[string]interface{}, int, error)
}

// IntegrationService handles integration operations with partner systems
type IntegrationService struct {
    catalogService catalogService.CatalogServiceInterface
    orderService   orderService.OrderServiceInterface
    logger         *logrus.Logger
    webhookSecret  string
}

// NewIntegrationService creates a new integration service
func NewIntegrationService(
    catalogSvc catalogService.CatalogServiceInterface,
    orderSvc orderService.OrderServiceInterface,
    logger *logrus.Logger,
    webhookSecret string,
) *IntegrationService {
    return &IntegrationService{
        catalogService: catalogSvc,
        orderService:   orderSvc,
        logger:         logger,
        webhookSecret:  webhookSecret,
    }
}

// ProcessCatalogProposal processes a catalog proposal from partner systems
func (s *IntegrationService) ProcessCatalogProposal(
    ctx context.Context,
    req *integrationRequests.CatalogProposalRequest,
    userID string,
    orgID string,
) (*integrationResponses.CatalogProposalResponse, error) {
    s.logger.WithFields(logrus.Fields{
        "proposal_id": req.ProposalID,
        "partner_id":  req.PartnerID,
        "item_type":   req.CatalogItem.ItemType,
        "user_id":     userID,
        "org_id":      orgID,
    }).Info("Processing catalog proposal")

    // Validate proposal data
    if err := s.validateCatalogProposal(req); err != nil {
        s.logger.WithError(err).Error("Catalog proposal validation failed")
        return nil, fmt.Errorf("proposal validation failed: %w", err)
    }

    // Create approval workflow based on proposal type and priority
    approvalSteps := s.createApprovalWorkflow(req)

    // Store proposal for review (in a real implementation, this would be stored in database)
    response := &integrationResponses.CatalogProposalResponse{
        ProposalID:          req.ProposalID,
        Status:              "PENDING_REVIEW",
        SubmittedAt:         time.Now(),
        RequiresApproval:    req.RequiresApproval,
        ApprovalSteps:       approvalSteps,
        EstimatedReviewTime: s.calculateReviewTime(req.Priority),
        Priority:            req.Priority,
        TrackingURL:         fmt.Sprintf("https://api.kisanlink.com/api/v1/integrations/proposals/%s/status", req.ProposalID),
        WebhookURL:          req.WebhookURL,
    }

    // If auto-approval is enabled for low-priority items, process immediately
    if !req.RequiresApproval && req.Priority == "LOW" {
        catalogItemID, err := s.autoApproveCatalogProposal(ctx, req, userID, orgID)
        if err != nil {
            s.logger.WithError(err).Error("Auto-approval failed")
            response.Status = "REVIEW_REQUIRED"
            response.ReviewNotes = &[]string{"Auto-approval failed, manual review required"}[0]
        } else {
            response.Status = "APPROVED"
            response.ReviewedAt = &[]time.Time{time.Now()}[0]
            response.ReviewNotes = &[]string{"Auto-approved based on criteria"}[0]

            // In a real implementation, we would store the catalog item ID
            s.logger.WithField("catalog_item_id", catalogItemID).Info("Catalog proposal auto-approved")
        }
    }

    return response, nil
}

// ProcessOrderAcknowledgement processes order acknowledgement from downstream systems
func (s *IntegrationService) ProcessOrderAcknowledgement(
    ctx context.Context,
    req *integrationRequests.OrderAcknowledgementRequest,
    userID string,
    orgID string,
) (*integrationResponses.OrderAcknowledgementResponse, error) {
    s.logger.WithFields(logrus.Fields{
        "order_id":          req.OrderID,
        "external_order_id": req.ExternalOrderID,
        "partner_id":        req.PartnerID,
        "status":            req.Status,
        "user_id":           userID,
        "org_id":            orgID,
    }).Info("Processing order acknowledgement")

    // Validate acknowledgement data
    if err := s.validateOrderAcknowledgement(req); err != nil {
        s.logger.WithError(err).Error("Order acknowledgement validation failed")
        return nil, fmt.Errorf("acknowledgement validation failed: %w", err)
    }

    // Verify order exists and user has permission to acknowledge
    // In a real implementation, this would check the order service
    // For now, we'll simulate the check
    if req.OrderID == "" {
        return nil, fmt.Errorf("order not found: %s", req.OrderID)
    }

    // Generate acknowledgement ID
    acknowledgementID := uuid.New().String()

    // Process based on acknowledgement status
    var nextSteps []string
    switch req.Status {
    case "RECEIVED":
        nextSteps = []string{"Order received and queued for processing"}
    case "ACCEPTED":
        nextSteps = []string{"Order accepted and will be processed within estimated time"}
    case "REJECTED":
        nextSteps = []string{"Order rejected, please review rejection reason and resubmit if needed"}
    }

    response := &integrationResponses.OrderAcknowledgementResponse{
        AcknowledgementID:       acknowledgementID,
        OrderID:                 req.OrderID,
        Status:                  "ACKNOWLEDGED",
        ProcessedAt:             time.Now(),
        PartnerID:               req.PartnerID,
        PartnerName:             req.PartnerName,
        EstimatedProcessingTime: req.EstimatedProcessingTime,
        AssignedTo:              req.AssignedTo,
        NextSteps:               nextSteps,
        TrackingURL:             fmt.Sprintf("https://api.kisanlink.com/api/v1/orders/%s/status", req.OrderID),
        WebhookURL:              req.WebhookURL,
    }

    // In a real implementation, we would update the order status and store acknowledgement
    s.logger.WithField("acknowledgement_id", acknowledgementID).Info("Order acknowledgement processed")

    return response, nil
}

// ExportCatalogWithDelta exports catalog items with delta synchronization
func (s *IntegrationService) ExportCatalogWithDelta(
    ctx context.Context,
    req *integrationRequests.CatalogExportRequest,
    userID string,
    orgID string,
) (*integrationResponses.CatalogExportResponse, error) {
    s.logger.WithFields(logrus.Fields{
        "partner_id":      req.PartnerID,
        "since_watermark": req.SinceWatermark,
        "organization_id": req.OrganizationID,
        "item_type":       req.ItemType,
        "page":            req.Page,
        "page_size":       req.PageSize,
        "user_id":         userID,
        "org_id":          orgID,
    }).Info("Exporting catalog with delta synchronization")

    // Validate export request
    if err := s.validateCatalogExportRequest(req); err != nil {
        s.logger.WithError(err).Error("Catalog export request validation failed")
        return nil, fmt.Errorf("export request validation failed: %w", err)
    }

    // Set default pagination
    if req.Page < 1 {
        req.Page = 1
    }
    if req.PageSize < 1 || req.PageSize > 100 {
        req.PageSize = 50
    }

    // Generate export ID
    exportID := uuid.New().String()
    currentTime := time.Now()

    // In a real implementation, this would query the catalog service with delta filters
    // For now, we'll create a mock response
    items := s.generateMockCatalogExportItems(req)

    response := &integrationResponses.CatalogExportResponse{
        ExportID:    exportID,
        GeneratedAt: currentTime,
        Watermark:   currentTime,
        TotalItems:  len(items) * 3, // Simulate total across all pages
        ItemsInPage: len(items),
        Page:        req.Page,
        PageSize:    req.PageSize,
        HasNextPage: req.Page < 3, // Simulate 3 pages total
        Filters: integrationResponses.ExportFilters{
            OrganizationID: req.OrganizationID,
            ItemType:       req.ItemType,
            Category:       req.Category,
            Visibility:     req.Visibility,
            IsActive:       req.IsActive,
        },
        Items:         items,
        ChangedSince:  req.SinceWatermark,
        NextWatermark: currentTime,
    }

    s.logger.WithFields(logrus.Fields{
        "export_id":     exportID,
        "total_items":   response.TotalItems,
        "items_in_page": response.ItemsInPage,
    }).Info("Catalog export completed")

    return response, nil
}

// ValidateWebhookSignature validates webhook signatures for secure partner integrations
func (s *IntegrationService) ValidateWebhookSignature(
    ctx context.Context,
    req *integrationRequests.WebhookSignatureRequest,
    partnerID string,
) (*integrationResponses.WebhookValidationResponse, error) {
    s.logger.WithFields(logrus.Fields{
        "partner_id": partnerID,
        "timestamp":  req.Timestamp,
    }).Info("Validating webhook signature")

    // Parse timestamp
    timestamp, err := strconv.ParseInt(req.Timestamp, 10, 64)
    if err != nil {
        return &integrationResponses.WebhookValidationResponse{
            Valid:   false,
            Message: "Invalid timestamp format",
        }, nil
    }

    // Check timestamp tolerance (5 minutes)
    now := time.Now().Unix()
    if abs(now-timestamp) > 300 {
        return &integrationResponses.WebhookValidationResponse{
            Valid:   false,
            Message: "Timestamp outside acceptable range",
        }, nil
    }

    // Create expected signature
    payload := fmt.Sprintf("%d.%s", timestamp, string(req.Body))
    expectedSignature := s.generateHMACSignature(payload, s.webhookSecret)

    // Compare signatures
    if !hmac.Equal([]byte(req.Signature), []byte(expectedSignature)) {
        s.logger.WithFields(logrus.Fields{
            "partner_id":         partnerID,
            "expected_signature": expectedSignature,
            "received_signature": req.Signature,
        }).Warn("Webhook signature validation failed")

        return &integrationResponses.WebhookValidationResponse{
            Valid:   false,
            Message: "Invalid signature",
        }, nil
    }

    return &integrationResponses.WebhookValidationResponse{
        Valid:     true,
        Message:   "Signature validated successfully",
        Timestamp: time.Unix(timestamp, 0).Format(time.RFC3339),
        PartnerID: partnerID,
    }, nil
}

// Helper methods

func (s *IntegrationService) validateCatalogProposal(req *integrationRequests.CatalogProposalRequest) error {
    if req.PartnerID == "" {
        return fmt.Errorf("partner_id is required")
    }
    if req.ProposalID == "" {
        return fmt.Errorf("proposal_id is required")
    }
    if req.CatalogItem.Name == "" {
        return fmt.Errorf("catalog item name is required")
    }
    if req.CatalogItem.BasePrice.IsZero() || req.CatalogItem.BasePrice.IsNegative() {
        return fmt.Errorf("base_price must be positive")
    }
    return nil
}

func (s *IntegrationService) validateOrderAcknowledgement(req *integrationRequests.OrderAcknowledgementRequest) error {
    if req.OrderID == "" {
        return fmt.Errorf("order_id is required")
    }
    if req.PartnerID == "" {
        return fmt.Errorf("partner_id is required")
    }
    if req.Status == "" {
        return fmt.Errorf("status is required")
    }
    return nil
}

func (s *IntegrationService) validateCatalogExportRequest(req *integrationRequests.CatalogExportRequest) error {
    if req.PartnerID == "" {
        return fmt.Errorf("partner_id is required")
    }
    return nil
}

func (s *IntegrationService) createApprovalWorkflow(req *integrationRequests.CatalogProposalRequest) []integrationResponses.ApprovalStep {
    steps := []integrationResponses.ApprovalStep{
        {
            StepID:   "step_1",
            StepName: "Technical Review",
            Status:   "PENDING",
        },
    }

    if req.Priority == "HIGH" || req.Priority == "URGENT" {
        steps = append(steps, integrationResponses.ApprovalStep{
            StepID:   "step_2",
            StepName: "Management Approval",
            Status:   "PENDING",
        })
    }

    return steps
}

func (s *IntegrationService) calculateReviewTime(priority string) *int {
    switch priority {
    case "URGENT":
        return &[]int{2}[0] // 2 hours
    case "HIGH":
        return &[]int{8}[0] // 8 hours
    case "MEDIUM":
        return &[]int{24}[0] // 24 hours
    case "LOW":
        return &[]int{72}[0] // 72 hours
    default:
        return &[]int{24}[0] // 24 hours default
    }
}

func (s *IntegrationService) autoApproveCatalogProposal(
    ctx context.Context,
    req *integrationRequests.CatalogProposalRequest,
    userID string,
    orgID string,
) (string, error) {
    // In a real implementation, this would create the catalog item via catalog service
    // For now, we'll simulate success
    catalogItemID := uuid.New().String()

    s.logger.WithFields(logrus.Fields{
        "proposal_id":     req.ProposalID,
        "catalog_item_id": catalogItemID,
    }).Info("Catalog item auto-created from proposal")

    return catalogItemID, nil
}

func (s *IntegrationService) generateMockCatalogExportItems(req *integrationRequests.CatalogExportRequest) []integrationResponses.CatalogExportItem {
    // Generate mock items based on request parameters
    items := make([]integrationResponses.CatalogExportItem, 0, req.PageSize)

    for i := 0; i < req.PageSize && i < 10; i++ {
        item := integrationResponses.CatalogExportItem{
            ID:             fmt.Sprintf("item_%d", i+1),
            GlobalID:       fmt.Sprintf("global_%d", i+1),
            OrganizationID: "org_123",
            ItemType:       "PRODUCT",
            Category:       "Seeds",
            Subcategory:    "Vegetable Seeds",
            Name:           fmt.Sprintf("Mock Product %d", i+1),
            Description:    fmt.Sprintf("Description for mock product %d", i+1),
            SKU:            fmt.Sprintf("SKU-%03d", i+1),
            UnitOfMeasure:  "packet",
            IsActive:       true,
            Visibility:     "NETWORK",
            CreatedAt:      time.Now().Add(-time.Hour * 24),
            UpdatedAt:      time.Now().Add(-time.Hour),
            Version:        1,
            ChangeType:     "UPDATED",
        }

        // Set base price
        item.BasePrice = decimal.NewFromFloat(float64(25 + i*5))
        item.Currency = "INR"

        items = append(items, item)
    }

    return items
}

func (s *IntegrationService) generateHMACSignature(payload, secret string) string {
    h := hmac.New(sha256.New, []byte(secret))
    h.Write([]byte(payload))
    return "sha256=" + hex.EncodeToString(h.Sum(nil))
}

func abs(x int64) int64 {
    if x < 0 {
        return -x
    }
    return x
}

// GetProposalStatus retrieves the status of a catalog proposal
func (s *IntegrationService) GetProposalStatus(ctx context.Context, proposalID string, userID string, orgID string) (map[string]interface{}, error) {
    s.logger.WithFields(logrus.Fields{
        "proposal_id": proposalID,
        "user_id":     userID,
        "org_id":      orgID,
    }).Info("Getting proposal status")

    // In a real implementation, this would query the database for proposal status
    // For now, we'll return a mock response
    mockResponse := map[string]interface{}{
        "proposal_id":  proposalID,
        "status":       "PENDING_REVIEW",
        "submitted_at": "2024-01-15T10:30:00Z",
        "message":      "Proposal is under review",
    }

    return mockResponse, nil
}

// ListIntegrationPartners retrieves a list of integration partners
func (s *IntegrationService) ListIntegrationPartners(ctx context.Context, page, limit int, userID string, orgID string) ([]map[string]interface{}, int, error) {
    s.logger.WithFields(logrus.Fields{
        "page":    page,
        "limit":   limit,
        "user_id": userID,
        "org_id":  orgID,
    }).Info("Listing integration partners")

    // In a real implementation, this would query the database for partners
    // For now, we'll return mock data
    mockPartners := []map[string]interface{}{
        {
            "partner_id":   "partner_123",
            "partner_name": "FarmTech Solutions",
            "status":       "ACTIVE",
            "created_at":   "2024-01-01T00:00:00Z",
        },
        {
            "partner_id":   "partner_456",
            "partner_name": "Logistics Partner",
            "status":       "ACTIVE",
            "created_at":   "2024-01-02T00:00:00Z",
        },
    }

    return mockPartners, len(mockPartners), nil
}
