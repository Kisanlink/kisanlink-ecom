package integrations

import (
	"bytes"
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

	integrationRequests "kisanlink-ecom/entities/requests/integrations"
	"kisanlink-ecom/internal/handlers/integrations"
	integrationService "kisanlink-ecom/internal/services/integrations"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite provides integration tests that simulate partner system interactions
type IntegrationTestSuite struct {
	suite.Suite
	router        *gin.Engine
	service       *integrationService.IntegrationService
	webhookSecret string
}

func (suite *IntegrationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // Reduce noise in tests

	// Set webhook secret for testing
	suite.webhookSecret = "test-webhook-secret-key"

	// Create integration service with mock dependencies
	suite.service = integrationService.NewIntegrationService(
		nil, // catalogService - not needed for these tests
		nil, // orderService - not needed for these tests
		logger,
		suite.webhookSecret,
	)

	// Create handler
	handler := integrations.NewIntegrationHandler(suite.service)

	// Setup router
	suite.router = gin.New()

	// Add middleware to set auth context
	suite.router.Use(func(c *gin.Context) {
		c.Set("subjectID", "integration-test-user")
		c.Set("organizationID", "integration-test-org")
		c.Next()
	})

	// Setup routes
	v1 := suite.router.Group("/api/v1")
	integrationsGroup := v1.Group("/integrations")
	{
		catalogIntegration := integrationsGroup.Group("/catalog")
		{
			catalogIntegration.POST("/proposals", handler.SubmitCatalogProposal)
			catalogIntegration.GET("/exports", handler.ExportCatalog)
		}

		orderIntegration := integrationsGroup.Group("/orders")
		{
			orderIntegration.POST("/acknowledgements", handler.AcknowledgeOrder)
		}

		integrationsGroup.POST("/webhooks/validate", handler.ValidateWebhookSignature)
		integrationsGroup.GET("/proposals/:proposal_id/status", handler.GetProposalStatus)
		integrationsGroup.GET("/partners", handler.ListIntegrationPartners)
	}
}

func (suite *IntegrationTestSuite) generateWebhookSignature(payload string) string {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signaturePayload := fmt.Sprintf("%s.%s", timestamp, payload)

	h := hmac.New(sha256.New, []byte(suite.webhookSecret))
	h.Write([]byte(signaturePayload))
	signature := "sha256=" + hex.EncodeToString(h.Sum(nil))

	return signature
}

func (suite *IntegrationTestSuite) TestPartnerCatalogProposalWorkflow() {
	// Test Case: Partner submits a new product proposal

	// Step 1: Partner submits catalog proposal
	proposalReq := integrationRequests.CatalogProposalRequest{
		PartnerID:    "farmtech_solutions",
		PartnerName:  "FarmTech Solutions Pvt Ltd",
		ProposalID:   "FTS_PROP_2024_001",
		ProposalType: "CREATE",
		CatalogItem: integrationRequests.CatalogItemProposal{
			ExternalID:    "FTS_ORGANIC_TOMATO_SEEDS_V2",
			ItemType:      "PRODUCT",
			Name:          "Premium Organic Tomato Seeds - Hybrid Variety",
			Description:   "High-yield organic tomato seeds certified by NPOP. Suitable for greenhouse and open field cultivation. Disease resistant variety with 90% germination rate.",
			SKU:           "FTS-OTS-HYB-001",
			Category:      "Seeds",
			Subcategory:   "Vegetable Seeds",
			BasePrice:     decimal.NewFromFloat(45.75),
			Currency:      "INR",
			Weight:        &[]decimal.Decimal{decimal.NewFromFloat(0.25)}[0], // 250g packet
			UnitOfMeasure: "packet",
			Perishable:    &[]bool{false}[0],
			ShelfLifeDays: &[]int{730}[0], // 2 years
			Tags:          []string{"organic", "certified", "hybrid", "disease-resistant", "high-yield"},
			Attributes: map[string]interface{}{
				"germination_rate":   "90%",
				"maturity_days":      "75-80",
				"plant_height":       "4-6 feet",
				"fruit_weight":       "150-200g",
				"certification":      "NPOP Organic",
				"suitable_seasons":   []string{"Kharif", "Rabi"},
				"growing_conditions": "Greenhouse, Open Field",
			},
			Images: []string{
				"https://farmtech.example.com/images/organic-tomato-seeds-front.jpg",
				"https://farmtech.example.com/images/organic-tomato-seeds-back.jpg",
				"https://farmtech.example.com/images/certification.jpg",
			},
			IsActive:   true,
			Visibility: "NETWORK",
		},
		RequiresApproval: true,
		Priority:         "HIGH", // High priority for certified organic products
		Notes:            "New premium organic variety with excellent market demand. Requesting expedited review for upcoming planting season.",
		WebhookURL:       "https://api.farmtech.example.com/webhooks/kisanlink/catalog",
	}

	// Create request with webhook signature
	reqBody, err := json.Marshal(proposalReq)
	suite.NoError(err)

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature := suite.generateWebhookSignature(string(reqBody))

	httpReq, _ := http.NewRequest("POST", "/api/v1/integrations/catalog/proposals", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Webhook-Signature", signature)
	httpReq.Header.Set("X-Webhook-Timestamp", timestamp)
	httpReq.Header.Set("User-Agent", "FarmTech-Integration-Client/1.0")

	// Execute request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Verify response
	suite.Equal(http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.True(response["success"].(bool))

	data := response["data"].(map[string]interface{})
	suite.Equal("FTS_PROP_2024_001", data["proposal_id"])
	suite.Equal("PENDING_REVIEW", data["status"])
	suite.True(data["requires_approval"].(bool))
	suite.Equal("HIGH", data["priority"])
	suite.NotNil(data["estimated_review_time"])
	suite.Contains(data["tracking_url"], "FTS_PROP_2024_001")

	// Step 2: Partner checks proposal status
	statusReq, _ := http.NewRequest("GET", "/api/v1/integrations/proposals/FTS_PROP_2024_001/status", nil)
	statusReq.Header.Set("User-Agent", "FarmTech-Integration-Client/1.0")

	w2 := httptest.NewRecorder()
	suite.router.ServeHTTP(w2, statusReq)

	suite.Equal(http.StatusOK, w2.Code)

	var statusResponse map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &statusResponse)
	suite.NoError(err)
	suite.True(statusResponse["success"].(bool))

	statusData := statusResponse["data"].(map[string]interface{})
	suite.Equal("FTS_PROP_2024_001", statusData["proposal_id"])
	suite.Equal("PENDING_REVIEW", statusData["status"])
}

func (suite *IntegrationTestSuite) TestLogisticsPartnerOrderAcknowledgement() {
	// Test Case: Logistics partner acknowledges order receipt and processing

	ackReq := integrationRequests.OrderAcknowledgementRequest{
		OrderID:                 "ORD_2024_KL_001234",
		ExternalOrderID:         "LOGISTICS_TRACK_789456",
		PartnerID:               "swift_logistics",
		PartnerName:             "Swift Agricultural Logistics",
		Status:                  "ACCEPTED",
		AcknowledgedAt:          time.Now(),
		EstimatedProcessingTime: &[]int{48}[0], // 48 hours
		AssignedTo:              &[]string{"warehouse_team_mumbai"}[0],
		Notes:                   "Order received at Mumbai warehouse. All items verified and in stock. Processing will begin within 2 hours.",
		Metadata: map[string]interface{}{
			"warehouse_location": "Mumbai Central Warehouse",
			"assigned_vehicle":   "MH12AB1234",
			"driver_contact":     "+91-9876543210",
			"expected_dispatch":  time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		},
		WebhookURL: "https://api.swiftlogistics.example.com/webhooks/kisanlink/orders",
	}

	// Create request with webhook signature
	reqBody, err := json.Marshal(ackReq)
	suite.NoError(err)

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature := suite.generateWebhookSignature(string(reqBody))

	httpReq, _ := http.NewRequest("POST", "/api/v1/integrations/orders/acknowledgements", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Webhook-Signature", signature)
	httpReq.Header.Set("X-Webhook-Timestamp", timestamp)
	httpReq.Header.Set("User-Agent", "SwiftLogistics-API-Client/2.1")

	// Execute request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Verify response
	suite.Equal(http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.True(response["success"].(bool))

	data := response["data"].(map[string]interface{})
	suite.Equal("ORD_2024_KL_001234", data["order_id"])
	suite.Equal("ACKNOWLEDGED", data["status"])
	suite.Equal("swift_logistics", data["partner_id"])
	suite.Equal("Swift Agricultural Logistics", data["partner_name"])
	suite.NotEmpty(data["acknowledgement_id"])
	suite.NotEmpty(data["next_steps"])
	suite.Contains(data["tracking_url"], "ORD_2024_KL_001234")
}

func (suite *IntegrationTestSuite) TestCatalogExportWithDeltaSync() {
	// Test Case: Partner system requests catalog export with delta synchronization

	// Step 1: Initial full export
	initialReq, _ := http.NewRequest("GET", "/api/v1/integrations/catalog/exports?partner_id=marketplace_partner&item_type=PRODUCT&visibility=NETWORK&is_active=true&page=1&page_size=25&format=JSON&include_metadata=true", nil)
	initialReq.Header.Set("User-Agent", "MarketplacePartner-Sync/1.5")

	w1 := httptest.NewRecorder()
	suite.router.ServeHTTP(w1, initialReq)

	suite.Equal(http.StatusOK, w1.Code)

	var initialResponse map[string]interface{}
	err := json.Unmarshal(w1.Body.Bytes(), &initialResponse)
	suite.NoError(err)
	suite.True(initialResponse["success"].(bool))

	initialData := initialResponse["data"].(map[string]interface{})
	suite.NotEmpty(initialData["export_id"])
	suite.NotEmpty(initialData["watermark"])
	suite.NotEmpty(initialData["next_watermark"])
	suite.Greater(initialData["total_items"], float64(0))
	suite.NotEmpty(initialData["items"])

	// Extract watermark for delta sync
	nextWatermark := initialData["next_watermark"].(string)

	// Step 2: Delta sync request using watermark
	time.Sleep(100 * time.Millisecond) // Small delay to simulate time passage

	deltaReq, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/integrations/catalog/exports?partner_id=marketplace_partner&since_watermark=%s&item_type=PRODUCT&page=1&page_size=25&format=JSON", nextWatermark), nil)
	deltaReq.Header.Set("User-Agent", "MarketplacePartner-Sync/1.5")

	w2 := httptest.NewRecorder()
	suite.router.ServeHTTP(w2, deltaReq)

	suite.Equal(http.StatusOK, w2.Code)

	var deltaResponse map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &deltaResponse)
	suite.NoError(err)
	suite.True(deltaResponse["success"].(bool))

	deltaData := deltaResponse["data"].(map[string]interface{})
	suite.NotEmpty(deltaData["export_id"])
	suite.NotEmpty(deltaData["watermark"])
	suite.NotEqual(initialData["export_id"], deltaData["export_id"]) // Different export IDs
	suite.NotEmpty(deltaData["changed_since"])

	// Verify pagination metadata
	meta := initialResponse["meta"].(map[string]interface{})
	pagination := meta["pagination"].(map[string]interface{})
	suite.Equal(float64(1), pagination["page"])
	suite.Equal(float64(25), pagination["limit"])
	suite.NotNil(pagination["total"])
}

func (suite *IntegrationTestSuite) TestWebhookSignatureValidation() {
	// Test Case: Validate webhook signatures for secure partner integrations

	testPayload := `{"test": "webhook payload", "timestamp": "2024-01-15T10:30:00Z"}`
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// Generate valid signature
	signaturePayload := fmt.Sprintf("%s.%s", timestamp, testPayload)
	h := hmac.New(sha256.New, []byte(suite.webhookSecret))
	h.Write([]byte(signaturePayload))
	validSignature := "sha256=" + hex.EncodeToString(h.Sum(nil))

	// Test valid signature
	validReq, _ := http.NewRequest("POST", "/api/v1/integrations/webhooks/validate?partner_id=test_partner", bytes.NewBufferString(testPayload))
	validReq.Header.Set("X-Webhook-Signature", validSignature)
	validReq.Header.Set("X-Webhook-Timestamp", timestamp)

	w1 := httptest.NewRecorder()
	suite.router.ServeHTTP(w1, validReq)

	suite.Equal(http.StatusOK, w1.Code)

	var validResponse map[string]interface{}
	err := json.Unmarshal(w1.Body.Bytes(), &validResponse)
	suite.NoError(err)
	suite.True(validResponse["success"].(bool))

	validData := validResponse["data"].(map[string]interface{})
	suite.True(validData["valid"].(bool))
	suite.Equal("test_partner", validData["partner_id"])
	suite.Contains(validData["message"], "validated successfully")

	// Test invalid signature
	invalidSignature := "sha256=invalid_signature_hash"
	invalidReq, _ := http.NewRequest("POST", "/api/v1/integrations/webhooks/validate?partner_id=test_partner", bytes.NewBufferString(testPayload))
	invalidReq.Header.Set("X-Webhook-Signature", invalidSignature)
	invalidReq.Header.Set("X-Webhook-Timestamp", timestamp)

	w2 := httptest.NewRecorder()
	suite.router.ServeHTTP(w2, invalidReq)

	suite.Equal(http.StatusUnauthorized, w2.Code)

	var invalidResponse map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &invalidResponse)
	suite.NoError(err)
	suite.False(invalidResponse["success"].(bool))
	suite.Equal("INVALID_SIGNATURE", invalidResponse["error"].(map[string]interface{})["code"])
}

func (suite *IntegrationTestSuite) TestPartnerManagement() {
	// Test Case: List integration partners

	req, _ := http.NewRequest("GET", "/api/v1/integrations/partners?page=1&limit=10", nil)
	req.Header.Set("User-Agent", "KisanLink-Admin-Panel/1.0")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.True(response["success"].(bool))

	data := response["data"].([]interface{})
	suite.Len(data, 2) // Mock returns 2 partners

	// Verify partner structure
	partner := data[0].(map[string]interface{})
	suite.NotEmpty(partner["partner_id"])
	suite.NotEmpty(partner["partner_name"])
	suite.NotEmpty(partner["status"])
	suite.NotEmpty(partner["created_at"])

	// Verify pagination metadata
	meta := response["meta"].(map[string]interface{})
	pagination := meta["pagination"].(map[string]interface{})
	suite.Equal(float64(1), pagination["page"])
	suite.Equal(float64(10), pagination["limit"])
	suite.Equal(float64(2), pagination["total"])
	suite.False(pagination["has_next"].(bool))
}

func (suite *IntegrationTestSuite) TestErrorHandling() {
	// Test Case: Error handling for various scenarios

	// Test missing webhook signature when timestamp is provided
	req1, _ := http.NewRequest("POST", "/api/v1/integrations/catalog/proposals", bytes.NewBufferString(`{"partner_id": "test"}`))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("X-Webhook-Timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	// Missing X-Webhook-Signature header

	w1 := httptest.NewRecorder()
	suite.router.ServeHTTP(w1, req1)

	suite.Equal(http.StatusBadRequest, w1.Code)

	// Test invalid JSON payload
	req2, _ := http.NewRequest("POST", "/api/v1/integrations/catalog/proposals", bytes.NewBufferString(`{invalid json}`))
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	suite.router.ServeHTTP(w2, req2)

	suite.Equal(http.StatusBadRequest, w2.Code)

	// Test missing required query parameters
	req3, _ := http.NewRequest("GET", "/api/v1/integrations/catalog/exports", nil) // Missing partner_id

	w3 := httptest.NewRecorder()
	suite.router.ServeHTTP(w3, req3)

	suite.Equal(http.StatusBadRequest, w3.Code)
}

func (suite *IntegrationTestSuite) TestConcurrentRequests() {
	// Test Case: Handle concurrent requests from multiple partners

	const numConcurrentRequests = 10
	results := make(chan int, numConcurrentRequests)

	// Create multiple concurrent catalog export requests
	for i := 0; i < numConcurrentRequests; i++ {
		go func(partnerIndex int) {
			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/integrations/catalog/exports?partner_id=partner_%d&page=1&page_size=10", partnerIndex), nil)
			req.Header.Set("User-Agent", fmt.Sprintf("Partner-%d-Client/1.0", partnerIndex))

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			results <- w.Code
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < numConcurrentRequests; i++ {
		statusCode := <-results
		if statusCode == http.StatusOK {
			successCount++
		}
	}

	// All requests should succeed
	suite.Equal(numConcurrentRequests, successCount)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
