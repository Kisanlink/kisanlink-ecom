package integrations

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	integrationRequests "kisanlink-ecom/entities/requests/integrations"
	"kisanlink-ecom/internal/handlers/integrations"
	integrationService "kisanlink-ecom/internal/services/integrations"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSimpleIntegrationHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a real integration service for testing
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	service := integrationService.NewIntegrationService(nil, nil, logger, "test-secret")
	handler := integrations.NewIntegrationHandler(service)

	// Setup router
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("subjectID", "test-user")
		c.Set("organizationID", "test-org")
		c.Next()
	})

	router.POST("/proposals", handler.SubmitCatalogProposal)

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

	// Create HTTP request
	reqBody, err := json.Marshal(req)
	assert.NoError(t, err)

	httpReq, _ := http.NewRequest("POST", "/proposals", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	// Print response for debugging
	t.Logf("Response Status: %d", w.Code)
	t.Logf("Response Body: %s", w.Body.String())

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check the structure
	t.Logf("Response structure: %+v", response)
}
