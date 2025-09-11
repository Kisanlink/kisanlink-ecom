package integrations

import (
    "time"

    "github.com/shopspring/decimal"
)

// CatalogProposalResponse represents the response after submitting a catalog proposal
type CatalogProposalResponse struct {
    // Proposal tracking
    ProposalID  string    `json:"proposal_id" example:"prop_456"`
    Status      string    `json:"status" example:"PENDING_REVIEW"`
    SubmittedAt time.Time `json:"submitted_at" example:"2024-01-15T10:30:00Z"`

    // Review information
    ReviewerID *string    `json:"reviewer_id,omitempty" example:"reviewer_123"`
    ReviewedAt *time.Time `json:"reviewed_at,omitempty" example:"2024-01-15T11:00:00Z"`

    // Approval workflow
    RequiresApproval bool           `json:"requires_approval" example:"true"`
    ApprovalSteps    []ApprovalStep `json:"approval_steps,omitempty"`

    // Processing details
    EstimatedReviewTime *int   `json:"estimated_review_time,omitempty" example:"24"` // hours
    Priority            string `json:"priority" example:"MEDIUM"`

    // Feedback
    ReviewNotes *string `json:"review_notes,omitempty" example:"Proposal looks good, pending final approval"`

    // Tracking
    TrackingURL string `json:"tracking_url" example:"https://api.kisanlink.com/api/v1/integrations/proposals/prop_456/status"`
    WebhookURL  string `json:"webhook_url,omitempty" example:"https://partner.example.com/webhooks/catalog"`
}

// ApprovalStep represents a step in the approval workflow
type ApprovalStep struct {
    StepID      string     `json:"step_id" example:"step_1"`
    StepName    string     `json:"step_name" example:"Technical Review"`
    Status      string     `json:"status" example:"PENDING"`
    AssignedTo  *string    `json:"assigned_to,omitempty" example:"tech_reviewer_1"`
    CompletedAt *time.Time `json:"completed_at,omitempty"`
    Notes       *string    `json:"notes,omitempty"`
}

// OrderAcknowledgementResponse represents the response after order acknowledgement
type OrderAcknowledgementResponse struct {
    // Acknowledgement tracking
    AcknowledgementID string    `json:"acknowledgement_id" example:"ack_789"`
    OrderID           string    `json:"order_id" example:"order_123"`
    Status            string    `json:"status" example:"ACKNOWLEDGED"`
    ProcessedAt       time.Time `json:"processed_at" example:"2024-01-15T10:30:00Z"`

    // Partner information
    PartnerID   string `json:"partner_id" example:"partner_123"`
    PartnerName string `json:"partner_name" example:"Logistics Partner"`

    // Processing information
    EstimatedProcessingTime *int    `json:"estimated_processing_time,omitempty" example:"24"`
    AssignedTo              *string `json:"assigned_to,omitempty" example:"warehouse_team_1"`

    // Next steps
    NextSteps []string `json:"next_steps,omitempty" example:"Order will be processed within 24 hours"`

    // Tracking
    TrackingURL string `json:"tracking_url" example:"https://api.kisanlink.com/api/v1/orders/order_123/status"`
    WebhookURL  string `json:"webhook_url,omitempty" example:"https://partner.example.com/webhooks/orders"`
}

// CatalogExportResponse represents the response for catalog export requests
type CatalogExportResponse struct {
    // Export metadata
    ExportID    string    `json:"export_id" example:"export_123"`
    GeneratedAt time.Time `json:"generated_at" example:"2024-01-15T10:30:00Z"`
    Watermark   time.Time `json:"watermark" example:"2024-01-15T10:30:00Z"`

    // Data summary
    TotalItems  int  `json:"total_items" example:"150"`
    ItemsInPage int  `json:"items_in_page" example:"50"`
    Page        int  `json:"page" example:"1"`
    PageSize    int  `json:"page_size" example:"50"`
    HasNextPage bool `json:"has_next_page" example:"true"`

    // Filtering applied
    Filters ExportFilters `json:"filters"`

    // Catalog items
    Items []CatalogExportItem `json:"items"`

    // Delta sync information
    ChangedSince  *time.Time `json:"changed_since,omitempty" example:"2024-01-15T10:00:00Z"`
    NextWatermark time.Time  `json:"next_watermark" example:"2024-01-15T10:30:00Z"`
}

// ExportFilters represents the filters applied to the export
type ExportFilters struct {
    OrganizationID *string `json:"organization_id,omitempty" example:"org_123"`
    ItemType       *string `json:"item_type,omitempty" example:"PRODUCT"`
    Category       *string `json:"category,omitempty" example:"Seeds"`
    Visibility     *string `json:"visibility,omitempty" example:"NETWORK"`
    IsActive       *bool   `json:"is_active,omitempty" example:"true"`
}

// CatalogExportItem represents a catalog item in the export response
type CatalogExportItem struct {
    // Item identification
    ID         string `json:"id" example:"item_123"`
    GlobalID   string `json:"global_id" example:"global_456"`
    ExternalID string `json:"external_id,omitempty" example:"ext_789"`

    // Organization scoping
    OrganizationID string `json:"organization_id" example:"org_123"`

    // Item classification
    ItemType    string `json:"item_type" example:"PRODUCT"`
    Category    string `json:"category,omitempty" example:"Seeds"`
    Subcategory string `json:"subcategory,omitempty" example:"Vegetable Seeds"`

    // Basic information
    Name          string `json:"name" example:"Organic Tomato Seeds"`
    Description   string `json:"description,omitempty" example:"Premium organic tomato seeds"`
    SKU           string `json:"sku,omitempty" example:"OTS-001"`
    UnitOfMeasure string `json:"unit_of_measure,omitempty" example:"packet"`

    // Pricing
    BasePrice decimal.Decimal `json:"base_price" example:"25.50"`
    Currency  string          `json:"currency" example:"INR"`

    // Availability & Visibility
    IsActive   bool   `json:"is_active" example:"true"`
    Visibility string `json:"visibility" example:"NETWORK"`

    // Product-specific fields
    Weight        *decimal.Decimal `json:"weight,omitempty" example:"0.1"`
    Perishable    *bool            `json:"perishable,omitempty" example:"false"`
    ShelfLifeDays *int             `json:"shelf_life_days,omitempty" example:"730"`

    // Service-specific fields
    DurationMinutes *int    `json:"duration_minutes,omitempty" example:"120"`
    ServiceArea     *string `json:"service_area,omitempty" example:"Maharashtra"`

    // Labour-specific fields
    SkillLevel *string          `json:"skill_level,omitempty" example:"Expert"`
    HourlyRate *decimal.Decimal `json:"hourly_rate,omitempty" example:"75.00"`

    // Metadata
    Tags       []string               `json:"tags,omitempty" example:"organic,certified"`
    Attributes map[string]interface{} `json:"attributes,omitempty"`
    Images     []string               `json:"images,omitempty"`

    // Change tracking
    CreatedAt time.Time `json:"created_at" example:"2024-01-15T09:00:00Z"`
    UpdatedAt time.Time `json:"updated_at" example:"2024-01-15T10:00:00Z"`
    Version   int       `json:"version" example:"2"`

    // Delta sync metadata
    ChangeType string `json:"change_type,omitempty" example:"UPDATED"` // CREATED, UPDATED, DELETED
}

// ProposalStatusResponse represents the status of a catalog proposal
type ProposalStatusResponse struct {
    ProposalID  string     `json:"proposal_id" example:"prop_456"`
    Status      string     `json:"status" example:"APPROVED"`
    SubmittedAt time.Time  `json:"submitted_at" example:"2024-01-15T10:30:00Z"`
    ReviewedAt  *time.Time `json:"reviewed_at,omitempty" example:"2024-01-15T11:00:00Z"`
    ApprovedAt  *time.Time `json:"approved_at,omitempty" example:"2024-01-15T11:30:00Z"`

    // Review details
    ReviewerID  *string `json:"reviewer_id,omitempty" example:"reviewer_123"`
    ReviewNotes *string `json:"review_notes,omitempty" example:"Approved with minor modifications"`

    // Approval workflow
    ApprovalSteps []ApprovalStep `json:"approval_steps,omitempty"`

    // Result
    CatalogItemID   *string `json:"catalog_item_id,omitempty" example:"item_123"` // If approved and created
    RejectionReason *string `json:"rejection_reason,omitempty" example:"Duplicate item"`
}

// WebhookValidationResponse represents the response for webhook signature validation
type WebhookValidationResponse struct {
    Valid     bool   `json:"valid" example:"true"`
    Message   string `json:"message" example:"Signature validated successfully"`
    Timestamp string `json:"timestamp" example:"2024-01-15T10:30:00Z"`
    PartnerID string `json:"partner_id" example:"partner_123"`
}
