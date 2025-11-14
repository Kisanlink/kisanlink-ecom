package integrations

import (
	"io"
	"strconv"

	"github.com/Kisanlink/kisanlink-ecom/entities/requests/integrations"
	"github.com/Kisanlink/kisanlink-ecom/internal/common"
	integrationService "github.com/Kisanlink/kisanlink-ecom/internal/services/integrations"

	"github.com/gin-gonic/gin"
)

// IntegrationHandler handles HTTP requests for integration operations
type IntegrationHandler struct {
	integrationService integrationService.IntegrationServiceInterface
}

// NewIntegrationHandler creates a new integration handler
func NewIntegrationHandler(integrationService integrationService.IntegrationServiceInterface) *IntegrationHandler {
	return &IntegrationHandler{
		integrationService: integrationService,
	}
}

// SubmitCatalogProposal godoc
// @Summary Submit catalog proposal
// @Description Submit a catalog proposal from partner systems for federated catalog updates
// @Tags integrations
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param X-Webhook-Signature header string false "Webhook signature for validation"
// @Param X-Webhook-Timestamp header string false "Webhook timestamp"
// @Param proposal body object true "Catalog proposal data"
// @Success 201 {object} common.Response{data=map[string]interface{}}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/integrations/catalog/proposals [post]
func (h *IntegrationHandler) SubmitCatalogProposal(c *gin.Context) {
	var req integrations.CatalogProposalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Validate webhook signature if provided
	if signature := c.GetHeader("X-Webhook-Signature"); signature != "" {
		timestamp := c.GetHeader("X-Webhook-Timestamp")
		if timestamp == "" {
			common.BadRequest(c, "MISSING_TIMESTAMP", "X-Webhook-Timestamp header required when signature is provided", nil)
			return
		}

		// Read request body for signature validation
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			common.InternalServerError(c, "BODY_READ_ERROR", "Failed to read request body", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		webhookReq := &integrations.WebhookSignatureRequest{
			Signature: signature,
			Timestamp: timestamp,
			Body:      body,
		}

		validation, err := h.integrationService.ValidateWebhookSignature(c.Request.Context(), webhookReq, req.PartnerID)
		if err != nil {
			common.InternalServerError(c, "SIGNATURE_VALIDATION_ERROR", "Failed to validate webhook signature", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		if !validation.Valid {
			common.Unauthorized(c, "INVALID_SIGNATURE", validation.Message, map[string]interface{}{
				"partner_id": req.PartnerID,
			})
			return
		}
	}

	// Get user ID from context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Get organization ID from context
	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_ORG", "Organization ID not found in context", nil)
		return
	}

	// Process the catalog proposal
	response, err := h.integrationService.ProcessCatalogProposal(c.Request.Context(), &req, userID, orgID)
	if err != nil {
		common.BadRequest(c, "PROPOSAL_PROCESSING_FAILED", "Failed to process catalog proposal", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Created(c, response, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// AcknowledgeOrder godoc
// @Summary Acknowledge order
// @Description Acknowledge order intake from downstream systems for order confirmations
// @Tags integrations
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param X-Webhook-Signature header string false "Webhook signature for validation"
// @Param X-Webhook-Timestamp header string false "Webhook timestamp"
// @Param acknowledgement body object true "Order acknowledgement data"
// @Success 201 {object} common.Response{data=map[string]interface{}}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/integrations/orders/acknowledgements [post]
func (h *IntegrationHandler) AcknowledgeOrder(c *gin.Context) {
	var req integrations.OrderAcknowledgementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Validate webhook signature if provided
	if signature := c.GetHeader("X-Webhook-Signature"); signature != "" {
		timestamp := c.GetHeader("X-Webhook-Timestamp")
		if timestamp == "" {
			common.BadRequest(c, "MISSING_TIMESTAMP", "X-Webhook-Timestamp header required when signature is provided", nil)
			return
		}

		// Read request body for signature validation
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			common.InternalServerError(c, "BODY_READ_ERROR", "Failed to read request body", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		webhookReq := &integrations.WebhookSignatureRequest{
			Signature: signature,
			Timestamp: timestamp,
			Body:      body,
		}

		validation, err := h.integrationService.ValidateWebhookSignature(c.Request.Context(), webhookReq, req.PartnerID)
		if err != nil {
			common.InternalServerError(c, "SIGNATURE_VALIDATION_ERROR", "Failed to validate webhook signature", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		if !validation.Valid {
			common.Unauthorized(c, "INVALID_SIGNATURE", validation.Message, map[string]interface{}{
				"partner_id": req.PartnerID,
			})
			return
		}
	}

	// Get user ID from context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Get organization ID from context
	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_ORG", "Organization ID not found in context", nil)
		return
	}

	// Process the order acknowledgement
	response, err := h.integrationService.ProcessOrderAcknowledgement(c.Request.Context(), &req, userID, orgID)
	if err != nil {
		common.BadRequest(c, "ACKNOWLEDGEMENT_PROCESSING_FAILED", "Failed to process order acknowledgement", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Created(c, response, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// ExportCatalog godoc
// @Summary Export catalog with delta synchronization
// @Description Export catalog items with delta synchronization for partner systems
// @Tags integrations
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param partner_id query string true "Partner ID for access control"
// @Param since_watermark query string false "Delta sync watermark (RFC3339 format)"
// @Param organization_id query string false "Filter by organization ID"
// @Param item_type query string false "Filter by item type" Enums(PRODUCT, SERVICE, LABOUR)
// @Param category query string false "Filter by category"
// @Param visibility query string false "Filter by visibility" Enums(PRIVATE, ORG, NETWORK, PUBLIC)
// @Param is_active query bool false "Filter by active status"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(50)
// @Param format query string false "Output format" Enums(JSON, CSV, XML) default(JSON)
// @Param include_metadata query bool false "Include metadata in response" default(true)
// @Success 200 {object} common.Response{data=map[string]interface{}}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 403 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/integrations/catalog/exports [get]
func (h *IntegrationHandler) ExportCatalog(c *gin.Context) {
	var req integrations.CatalogExportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid query parameters", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 50
	}
	if req.Format == "" {
		req.Format = "JSON"
	}

	// Get user ID from context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Get organization ID from context
	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_ORG", "Organization ID not found in context", nil)
		return
	}

	// Export catalog with delta synchronization
	response, err := h.integrationService.ExportCatalogWithDelta(c.Request.Context(), &req, userID, orgID)
	if err != nil {
		common.BadRequest(c, "EXPORT_FAILED", "Failed to export catalog", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Handle different output formats
	switch req.Format {
	case "JSON":
		common.Success(c, response, &common.ResponseMeta{
			TraceID: common.GetTraceID(c),
			Pagination: &common.PaginationMeta{
				Page:    response.Page,
				Limit:   response.PageSize,
				Total:   response.TotalItems,
				HasNext: response.HasNextPage,
			},
		})
	case "CSV":
		// In a real implementation, this would generate CSV format
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=catalog_export.csv")
		c.String(200, "CSV export not implemented in this demo")
	case "XML":
		// In a real implementation, this would generate XML format
		c.Header("Content-Type", "application/xml")
		c.String(200, "<?xml version=\"1.0\"?><message>XML export not implemented in this demo</message>")
	default:
		common.BadRequest(c, "UNSUPPORTED_FORMAT", "Unsupported export format", map[string]interface{}{
			"format": req.Format,
		})
	}
}

// ValidateWebhookSignature godoc
// @Summary Validate webhook signature
// @Description Validate webhook signature for secure partner integrations
// @Tags integrations
// @Accept json
// @Produce json
// @Param X-Webhook-Signature header string true "Webhook signature"
// @Param X-Webhook-Timestamp header string true "Webhook timestamp"
// @Param partner_id query string true "Partner ID"
// @Success 200 {object} common.Response{data=map[string]interface{}}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/integrations/webhooks/validate [post]
func (h *IntegrationHandler) ValidateWebhookSignature(c *gin.Context) {
	signature := c.GetHeader("X-Webhook-Signature")
	timestamp := c.GetHeader("X-Webhook-Timestamp")
	partnerID := c.Query("partner_id")

	if signature == "" {
		common.BadRequest(c, "MISSING_SIGNATURE", "X-Webhook-Signature header is required", nil)
		return
	}

	if timestamp == "" {
		common.BadRequest(c, "MISSING_TIMESTAMP", "X-Webhook-Timestamp header is required", nil)
		return
	}

	if partnerID == "" {
		common.BadRequest(c, "MISSING_PARTNER_ID", "partner_id query parameter is required", nil)
		return
	}

	// Read request body for signature validation
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		common.InternalServerError(c, "BODY_READ_ERROR", "Failed to read request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	webhookReq := &integrations.WebhookSignatureRequest{
		Signature: signature,
		Timestamp: timestamp,
		Body:      body,
	}

	// Validate the webhook signature
	response, err := h.integrationService.ValidateWebhookSignature(c.Request.Context(), webhookReq, partnerID)
	if err != nil {
		common.InternalServerError(c, "VALIDATION_ERROR", "Failed to validate webhook signature", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	if response.Valid {
		common.Success(c, response, &common.ResponseMeta{
			TraceID: common.GetTraceID(c),
		})
	} else {
		common.Unauthorized(c, "INVALID_SIGNATURE", response.Message, map[string]interface{}{
			"partner_id": partnerID,
		})
	}
}

// GetProposalStatus godoc
// @Summary Get proposal status
// @Description Get the status of a catalog proposal
// @Tags integrations
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param proposal_id path string true "Proposal ID"
// @Success 200 {object} common.Response{data=map[string]interface{}}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/integrations/proposals/{proposal_id}/status [get]
func (h *IntegrationHandler) GetProposalStatus(c *gin.Context) {
	proposalID := c.Param("proposal_id")
	if proposalID == "" {
		common.BadRequest(c, "MISSING_PROPOSAL_ID", "Proposal ID is required", nil)
		return
	}

	// Get user ID from context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Get organization ID from context
	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_ORG", "Organization ID not found in context", nil)
		return
	}

	// Get proposal status from service
	response, err := h.integrationService.GetProposalStatus(c.Request.Context(), proposalID, userID, orgID)
	if err != nil {
		common.InternalServerError(c, "GET_STATUS_FAILED", "Failed to get proposal status", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, response, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// ListIntegrationPartners godoc
// @Summary List integration partners
// @Description List all registered integration partners
// @Tags integrations
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} common.Response{data=[]map[string]interface{}}
// @Router /api/v1/integrations/partners [get]
func (h *IntegrationHandler) ListIntegrationPartners(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get user ID from context
	userID, exists := common.GetSubjectID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Get organization ID from context
	orgID, exists := common.GetOrganizationID(c)
	if !exists {
		common.Unauthorized(c, "MISSING_ORG", "Organization ID not found in context", nil)
		return
	}

	// Get partners from service
	partners, total, err := h.integrationService.ListIntegrationPartners(c.Request.Context(), page, limit, userID, orgID)
	if err != nil {
		common.InternalServerError(c, "LIST_FAILED", "Failed to list integration partners", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, partners, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
		Pagination: &common.PaginationMeta{
			Page:    page,
			Limit:   limit,
			Total:   total,
			HasNext: len(partners) == limit,
		},
	})
}
