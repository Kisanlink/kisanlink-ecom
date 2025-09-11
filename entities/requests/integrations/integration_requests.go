package integrations

import (
    "time"

    "github.com/shopspring/decimal"
)

// CatalogProposalRequest represents a request to propose catalog changes from partner systems
type CatalogProposalRequest struct {
    // Partner identification
    PartnerID   string `json:"partner_id" binding:"required" example:"partner_123"`
    PartnerName string `json:"partner_name" binding:"required" example:"FarmTech Solutions"`

    // Proposal metadata
    ProposalID   string `json:"proposal_id" binding:"required" example:"prop_456"`
    ProposalType string `json:"proposal_type" binding:"required,oneof=CREATE UPDATE DELETE" example:"CREATE"`

    // Catalog item data
    CatalogItem CatalogItemProposal `json:"catalog_item" binding:"required"`

    // Validation and approval
    RequiresApproval bool   `json:"requires_approval" example:"true"`
    Priority         string `json:"priority" binding:"oneof=LOW MEDIUM HIGH URGENT" example:"MEDIUM"`
    Notes            string `json:"notes" example:"New organic product line from certified supplier"`

    // Webhook for status updates
    WebhookURL string `json:"webhook_url,omitempty" example:"https://partner.example.com/webhooks/catalog"`
}

// CatalogItemProposal represents the catalog item data in a proposal
type CatalogItemProposal struct {
    // Item identification
    ExternalID string `json:"external_id" binding:"required" example:"ext_789"`
    ItemType   string `json:"item_type" binding:"required,oneof=PRODUCT SERVICE LABOUR" example:"PRODUCT"`

    // Basic information
    Name        string `json:"name" binding:"required" example:"Organic Tomato Seeds"`
    Description string `json:"description" example:"Premium organic tomato seeds, certified by NPOP"`
    SKU         string `json:"sku" example:"OTS-001"`
    Category    string `json:"category" example:"Seeds"`
    Subcategory string `json:"subcategory" example:"Vegetable Seeds"`

    // Pricing
    BasePrice decimal.Decimal `json:"base_price" binding:"required" example:"25.50"`
    Currency  string          `json:"currency" example:"INR"`

    // Product-specific fields
    Weight        *decimal.Decimal `json:"weight,omitempty" example:"0.1"`
    UnitOfMeasure string           `json:"unit_of_measure" example:"packet"`
    Perishable    *bool            `json:"perishable,omitempty" example:"false"`
    ShelfLifeDays *int             `json:"shelf_life_days,omitempty" example:"730"`

    // Service-specific fields
    DurationMinutes *int    `json:"duration_minutes,omitempty" example:"120"`
    ServiceArea     *string `json:"service_area,omitempty" example:"Maharashtra"`

    // Labour-specific fields
    SkillLevel *string          `json:"skill_level,omitempty" example:"Expert"`
    HourlyRate *decimal.Decimal `json:"hourly_rate,omitempty" example:"75.00"`

    // Metadata
    Tags       []string               `json:"tags,omitempty" example:"organic,certified,premium"`
    Attributes map[string]interface{} `json:"attributes,omitempty"`
    Images     []string               `json:"images,omitempty"`

    // Availability
    IsActive   bool   `json:"is_active" example:"true"`
    Visibility string `json:"visibility" binding:"oneof=PRIVATE ORG NETWORK PUBLIC" example:"NETWORK"`
}

// OrderAcknowledgementRequest represents acknowledgement of order intake from downstream systems
type OrderAcknowledgementRequest struct {
    // Order identification
    OrderID         string `json:"order_id" binding:"required" example:"order_123"`
    ExternalOrderID string `json:"external_order_id" binding:"required" example:"ext_order_456"`

    // Partner identification
    PartnerID   string `json:"partner_id" binding:"required" example:"partner_123"`
    PartnerName string `json:"partner_name" binding:"required" example:"Logistics Partner"`

    // Acknowledgement details
    Status         string    `json:"status" binding:"required,oneof=RECEIVED ACCEPTED REJECTED" example:"ACCEPTED"`
    AcknowledgedAt time.Time `json:"acknowledged_at" binding:"required" example:"2024-01-15T10:30:00Z"`

    // Processing information
    EstimatedProcessingTime *int    `json:"estimated_processing_time,omitempty" example:"24"` // hours
    AssignedTo              *string `json:"assigned_to,omitempty" example:"warehouse_team_1"`

    // Rejection details (if status is REJECTED)
    RejectionReason *string `json:"rejection_reason,omitempty" example:"Insufficient inventory"`
    RejectionCode   *string `json:"rejection_code,omitempty" example:"INV_001"`

    // Additional metadata
    Notes    string                 `json:"notes,omitempty" example:"Order received and queued for processing"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`

    // Webhook for status updates
    WebhookURL string `json:"webhook_url,omitempty" example:"https://partner.example.com/webhooks/orders"`
}

// CatalogExportRequest represents parameters for catalog export with delta synchronization
type CatalogExportRequest struct {
    // Delta synchronization
    SinceWatermark *time.Time `form:"since_watermark" example:"2024-01-15T10:00:00Z"`

    // Filtering
    OrganizationID *string `form:"organization_id" example:"org_123"`
    ItemType       *string `form:"item_type" binding:"omitempty,oneof=PRODUCT SERVICE LABOUR" example:"PRODUCT"`
    Category       *string `form:"category" example:"Seeds"`
    Visibility     *string `form:"visibility" binding:"omitempty,oneof=PRIVATE ORG NETWORK PUBLIC" example:"NETWORK"`
    IsActive       *bool   `form:"is_active" example:"true"`

    // Pagination
    Page     int `form:"page" example:"1"`
    PageSize int `form:"page_size" example:"50"`

    // Output format
    Format          string `form:"format" binding:"oneof=JSON CSV XML" example:"JSON"`
    IncludeMetadata bool   `form:"include_metadata" example:"true"`

    // Partner identification for access control
    PartnerID string `form:"partner_id" binding:"required" example:"partner_123"`
}

// WebhookSignatureRequest represents the webhook signature validation data
type WebhookSignatureRequest struct {
    Signature string `header:"X-Webhook-Signature" binding:"required"`
    Timestamp string `header:"X-Webhook-Timestamp" binding:"required"`
    Body      []byte `json:"-"` // Raw request body for signature verification
}
