//go:build integration
// +build integration

package integrations

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	integrationRequests "github.com/Kisanlink/kisanlink-ecom/entities/requests/integrations"
	integrationResponses "github.com/Kisanlink/kisanlink-ecom/entities/responses/integrations"
	"github.com/Kisanlink/kisanlink-ecom/internal/handlers/integrations"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockIntegrationService is a mock implementation of IntegrationServiceInterface
type MockIntegrationService struct {
	mock.Mock
}

func (m *MockIntegrationService) ProcessCatalogProposal(
	ctx context.Context,
	req *integrationRequests.CatalogProposalRequest,
	userID string,
	orgID string,
) (*integrationResponses.CatalogProposalResponse, error) {
	args := m.Called(ctx, req, userID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*integrationResponses.CatalogProposalResponse), args.Error(1)
}

func (m *MockIntegrationService) ProcessOrderAcknowledgement(
	ctx context.Context,
	req *integrationRequests.OrderAcknowledgementRequest,
	userID string,
	orgID string,
) (*integrationResponses.OrderAcknowledgementResponse, error) {
	args := m.Called(ctx, req, userID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*integrationResponses.OrderAcknowledgementResponse), args.Error(1)
}

func (m *MockIntegrationService) ExportCatalogWithDelta(
	ctx context.Context,
	req *integrationRequests.CatalogExportRequest,
	userID string,
	orgID string,
) (*integrationResponses.CatalogExportResponse, error) {
	args := m.Called(ctx, req, userID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*integrationResponses.CatalogExportResponse), args.Error(1)
}

func (m *MockIntegrationService) ValidateWebhookSignature(
	ctx context.Context,
	req *integrationRequests.WebhookSignatureRequest,
	partnerID string,
) (*integrationResponses.WebhookValidationResponse, error) {
	args := m.Called(ctx, req, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*integrationResponses.WebhookValidationResponse), args.Error(1)
}

func (m *MockIntegrationService) GetProposalStatus(ctx context.Context, proposalID string, userID string, orgID string) (map[string]interface{}, error) {
	args := m.Called(ctx, proposalID, userID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockIntegrationService) ListIntegrationPartners(ctx context.Context, page, limit int, userID string, orgID string) ([]map[string]interface{}, int, error) {
	args := m.Called(ctx, page, limit, userID, orgID)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]map[string]interface{}), args.Int(1), args.Error(2)
}

func setupTestRouter(mockService *MockIntegrationService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := integrations.NewIntegrationHandler(mockService)

	// Set up routes
	v1 := router.Group("/api/v1")
	integrationsGroup := v1.Group("/integrations")
	{
		catalogIntegration := integrationsGroup.Group("/catalog")
		{
			catalogIntegration.POST("/proposals", func(c *gin.Context) {
				// Mock auth middleware
				c.Set("subjectID", "test-user-id")
				c.Set("organizationID", "test-org-id")
				handler.SubmitCatalogProposal(c)
			})
			catalogIntegration.GET("/exports", func(c *gin.Context) {
				// Mock auth middleware
				c.Set("subjectID", "test-user-id")
				c.Set("organizationID", "test-org-id")
				handler.ExportCatalog(c)
			})
		}

		orderIntegration := integrationsGroup.Group("/orders")
		{
			orderIntegration.POST("/acknowledgements", func(c *gin.Context) {
				// Mock auth middleware
				c.Set("subjectID", "test-user-id")
				c.Set("organizationID", "test-org-id")
				handler.AcknowledgeOrder(c)
			})
		}

		integrationsGroup.POST("/webhooks/validate", handler.ValidateWebhookSignature)
		integrationsGroup.GET("/proposals/:proposal_id/status", handler.GetProposalStatus)
		integrationsGroup.GET("/partners", handler.ListIntegrationPartners)
	}

	return router
}

func generateTestSignature(payload, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}

func TestSubmitCatalogProposal_Success(t *testing.T) {
	mockService := &MockIntegrationService{}
	router := setupTestRouter(mockService)

	// Create test request
	req := integrationRequests.CatalogProposalRequest{
		PartnerID:    "partner_123",
		PartnerName:  "Test Partner",
		ProposalID:   "prop_456",
		ProposalType: "CREATE",
		CatalogItem: integrationRequests.CatalogItemProposal{
			ExternalID:  "ext_789",
			ItemType:    "PRODUCT",
			Name:        "Test Product",
			Description: "Test Description",
			BasePrice:   decimal.NewFromFloat(25.50),
			Currency:    "INR",
			IsActive:    true,
			Visibility:  "NETWORK",
		},
		RequiresApproval: true,
		Priority:         "MEDIUM",
		Notes:            "Test proposal",
	}

	// Mock service response
	expectedResponse := &integrationResponses.CatalogProposalResponse{
		ProposalID:          "prop_456",
		Status:              "PENDING_REVIEW",
		SubmittedAt:         time.Now(),
		RequiresApproval:    true,
		EstimatedReviewTime: &[]int{24}[0],
		Priority:            "MEDIUM",
		TrackingURL:         "https://api.kisanlink.com/api/v1/integrations/proposals/prop_456/status",
	}

	mockService.On("ProcessCatalogProposal", mock.Anything, &req, "test-user-id", "test-org-id").
		Return(expectedResponse, nil)

	// Create HTTP request
	reqBody, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", "/api/v1/integrations/catalog/proposals", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "prop_456", data["proposal_id"])
	assert.Equal(t, "PENDING_REVIEW", data["status"])

	mockService.AssertExpectations(t)
}

func TestSubmitCatalogProposal_WithWebhookSignature(t *testing.T) {
	mockService := &MockIntegrationService{}
	router := setupTestRouter(mockService)

	// Create test request
	req := integrationRequests.CatalogProposalRequest{
		PartnerID:    "partner_123",
		PartnerName:  "Test Partner",
		ProposalID:   "prop_456",
		ProposalType: "CREATE",
		CatalogItem: integrationRequests.CatalogItemProposal{
			ExternalID: "ext_789",
			ItemType:   "PRODUCT",
			Name:       "Test Product",
			BasePrice:  decimal.NewFromFloat(25.50),
			IsActive:   true,
			Visibility: "NETWORK",
		},
		RequiresApproval: true,
		Priority:         "MEDIUM",
	}

	// Mock webhook validation
	webhookValidation := &integrationResponses.WebhookValidationResponse{
		Valid:     true,
		Message:   "Signature validated successfully",
		Timestamp: time.Now().Format(time.RFC3339),
		PartnerID: "partner_123",
	}

	mockService.On("ValidateWebhookSignature", mock.Anything, mock.AnythingOfType("*integrations.WebhookSignatureRequest"), "partner_123").
		Return(webhookValidation, nil)

	// Mock service response
	expectedResponse := &integrationResponses.CatalogProposalResponse{
		ProposalID:  "prop_456",
		Status:      "PENDING_REVIEW",
		SubmittedAt: time.Now(),
		Priority:    "MEDIUM",
		TrackingURL: "https://api.kisanlink.com/api/v1/integrations/proposals/prop_456/status",
	}

	mockService.On("ProcessCatalogProposal", mock.Anything, &req, "test-user-id", "test-org-id").
		Return(expectedResponse, nil)

	// Create HTTP request with webhook signature
	reqBody, _ := json.Marshal(req)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	payload := fmt.Sprintf("%s.%s", timestamp, string(reqBody))
	signature := generateTestSignature(payload, "test-secret")

	httpReq, _ := http.NewRequest("POST", "/api/v1/integrations/catalog/proposals", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Webhook-Signature", signature)
	httpReq.Header.Set("X-Webhook-Timestamp", timestamp)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	mockService.AssertExpectations(t)
}

func TestAcknowledgeOrder_Success(t *testing.T) {
	mockService := &MockIntegrationService{}
	router := setupTestRouter(mockService)

	// Create test request
	req := integrationRequests.OrderAcknowledgementRequest{
		OrderID:                 "order_123",
		ExternalOrderID:         "ext_order_456",
		PartnerID:               "partner_123",
		PartnerName:             "Logistics Partner",
		Status:                  "ACCEPTED",
		AcknowledgedAt:          time.Now(),
		EstimatedProcessingTime: &[]int{24}[0],
		AssignedTo:              &[]string{"warehouse_team_1"}[0],
		Notes:                   "Order received and queued",
	}

	// Mock service response
	expectedResponse := &integrationResponses.OrderAcknowledgementResponse{
		AcknowledgementID:       "ack_789",
		OrderID:                 "order_123",
		Status:                  "ACKNOWLEDGED",
		ProcessedAt:             time.Now(),
		PartnerID:               "partner_123",
		PartnerName:             "Logistics Partner",
		EstimatedProcessingTime: &[]int{24}[0],
		AssignedTo:              &[]string{"warehouse_team_1"}[0],
		NextSteps:               []string{"Order accepted and will be processed within estimated time"},
		TrackingURL:             "https://api.kisanlink.com/api/v1/orders/order_123/status",
	}

	mockService.On("ProcessOrderAcknowledgement", mock.Anything, &req, "test-user-id", "test-org-id").
		Return(expectedResponse, nil)

	// Create HTTP request
	reqBody, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", "/api/v1/integrations/orders/acknowledgements", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "ack_789", data["acknowledgement_id"])
	assert.Equal(t, "order_123", data["order_id"])
	assert.Equal(t, "ACKNOWLEDGED", data["status"])

	mockService.AssertExpectations(t)
}

func TestExportCatalog_Success(t *testing.T) {
	mockService := &MockIntegrationService{}
	router := setupTestRouter(mockService)

	// Mock service response
	expectedResponse := &integrationResponses.CatalogExportResponse{
		ExportID:    "export_123",
		GeneratedAt: time.Now(),
		Watermark:   time.Now(),
		TotalItems:  150,
		ItemsInPage: 50,
		Page:        1,
		PageSize:    50,
		HasNextPage: true,
		Filters: integrationResponses.ExportFilters{
			ItemType: &[]string{"PRODUCT"}[0],
			IsActive: &[]bool{true}[0],
		},
		Items: []integrationResponses.CatalogExportItem{
			{
				ID:             "item_1",
				GlobalID:       "global_1",
				OrganizationID: "org_123",
				ItemType:       "PRODUCT",
				Name:           "Test Product 1",
				BasePrice:      decimal.NewFromFloat(25.50),
				Currency:       "INR",
				IsActive:       true,
				Visibility:     "NETWORK",
				CreatedAt:      time.Now().Add(-time.Hour * 24),
				UpdatedAt:      time.Now().Add(-time.Hour),
				Version:        1,
				ChangeType:     "UPDATED",
			},
		},
		NextWatermark: time.Now(),
	}

	mockService.On("ExportCatalogWithDelta", mock.Anything, mock.AnythingOfType("*integrations.CatalogExportRequest"), "test-user-id", "test-org-id").
		Return(expectedResponse, nil)

	// Create HTTP request
	httpReq, _ := http.NewRequest("GET", "/api/v1/integrations/catalog/exports?partner_id=partner_123&item_type=PRODUCT&is_active=true&page=1&page_size=50", nil)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "export_123", data["export_id"])
	assert.Equal(t, float64(150), data["total_items"])
	assert.Equal(t, float64(50), data["items_in_page"])

	mockService.AssertExpectations(t)
}

func TestValidateWebhookSignature_Success(t *testing.T) {
	mockService := &MockIntegrationService{}
	router := setupTestRouter(mockService)

	// Mock service response
	expectedResponse := &integrationResponses.WebhookValidationResponse{
		Valid:     true,
		Message:   "Signature validated successfully",
		Timestamp: time.Now().Format(time.RFC3339),
		PartnerID: "partner_123",
	}

	mockService.On("ValidateWebhookSignature", mock.Anything, mock.AnythingOfType("*integrations.WebhookSignatureRequest"), "partner_123").
		Return(expectedResponse, nil)

	// Create HTTP request
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	payload := fmt.Sprintf("%s.%s", timestamp, "test body")
	signature := generateTestSignature(payload, "test-secret")

	httpReq, _ := http.NewRequest("POST", "/api/v1/integrations/webhooks/validate?partner_id=partner_123", bytes.NewBufferString("test body"))
	httpReq.Header.Set("X-Webhook-Signature", signature)
	httpReq.Header.Set("X-Webhook-Timestamp", timestamp)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.True(t, data["valid"].(bool))
	assert.Equal(t, "partner_123", data["partner_id"])

	mockService.AssertExpectations(t)
}

func TestSubmitCatalogProposal_InvalidRequest(t *testing.T) {
	mockService := &MockIntegrationService{}
	router := setupTestRouter(mockService)

	// Create invalid request (missing required fields)
	req := map[string]interface{}{
		"partner_id": "partner_123",
		// Missing required fields
	}

	// Create HTTP request
	reqBody, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", "/api/v1/integrations/catalog/proposals", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Equal(t, "INVALID_REQUEST", response["error"].(map[string]interface{})["code"])
}

func TestAcknowledgeOrder_ServiceError(t *testing.T) {
	mockService := &MockIntegrationService{}
	router := setupTestRouter(mockService)

	// Create test request
	req := integrationRequests.OrderAcknowledgementRequest{
		OrderID:         "order_123",
		ExternalOrderID: "ext_order_456",
		PartnerID:       "partner_123",
		PartnerName:     "Logistics Partner",
		Status:          "ACCEPTED",
		AcknowledgedAt:  time.Now(),
	}

	// Mock service error
	mockService.On("ProcessOrderAcknowledgement", mock.Anything, &req, "test-user-id", "test-org-id").
		Return(nil, fmt.Errorf("service error"))

	// Create HTTP request
	reqBody, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", "/api/v1/integrations/orders/acknowledgements", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Equal(t, "ACKNOWLEDGEMENT_PROCESSING_FAILED", response["error"].(map[string]interface{})["code"])

	mockService.AssertExpectations(t)
}

func TestExportCatalog_CSVFormat(t *testing.T) {
	mockService := &MockIntegrationService{}
	router := setupTestRouter(mockService)

	// Create HTTP request for CSV format
	httpReq, _ := http.NewRequest("GET", "/api/v1/integrations/catalog/exports?partner_id=partner_123&format=CSV", nil)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
}

func TestGetProposalStatus_Success(t *testing.T) {
	mockService := &MockIntegrationService{}
	router := setupTestRouter(mockService)

	// Create HTTP request
	httpReq, _ := http.NewRequest("GET", "/api/v1/integrations/proposals/prop_123/status", nil)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "prop_123", data["proposal_id"])
	assert.Equal(t, "PENDING_REVIEW", data["status"])
}

func TestListIntegrationPartners_Success(t *testing.T) {
	mockService := &MockIntegrationService{}
	router := setupTestRouter(mockService)

	// Create HTTP request
	httpReq, _ := http.NewRequest("GET", "/api/v1/integrations/partners?page=1&limit=20", nil)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	data := response["data"].([]interface{})
	assert.Len(t, data, 2) // Mock returns 2 partners

	partner := data[0].(map[string]interface{})
	assert.Equal(t, "partner_123", partner["partner_id"])
	assert.Equal(t, "FarmTech Solutions", partner["partner_name"])
}

// Benchmark tests for performance validation
func BenchmarkSubmitCatalogProposal(b *testing.B) {
	mockService := &MockIntegrationService{}
	router := setupTestRouter(mockService)

	req := integrationRequests.CatalogProposalRequest{
		PartnerID:    "partner_123",
		ProposalID:   "prop_456",
		ProposalType: "CREATE",
		CatalogItem: integrationRequests.CatalogItemProposal{
			ExternalID: "ext_789",
			ItemType:   "PRODUCT",
			Name:       "Test Product",
			BasePrice:  decimal.NewFromFloat(25.50),
			IsActive:   true,
			Visibility: "NETWORK",
		},
		Priority: "MEDIUM",
	}

	expectedResponse := &integrationResponses.CatalogProposalResponse{
		ProposalID:  "prop_456",
		Status:      "PENDING_REVIEW",
		SubmittedAt: time.Now(),
		Priority:    "MEDIUM",
	}

	mockService.On("ProcessCatalogProposal", mock.Anything, &req, "test-user-id", "test-org-id").
		Return(expectedResponse, nil)

	reqBody, _ := json.Marshal(req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		httpReq, _ := http.NewRequest("POST", "/api/v1/integrations/catalog/proposals", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, httpReq)
	}
}
