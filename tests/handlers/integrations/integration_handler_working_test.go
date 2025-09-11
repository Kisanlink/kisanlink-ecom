package integrations

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    integrationRequests "kisanlink-ecom/entities/requests/integrations"
    "kisanlink-ecom/internal/handlers/integrations"
    integrationService "kisanlink-ecom/internal/services/integrations"

    "github.com/gin-gonic/gin"
    "github.com/shopspring/decimal"
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
)

func TestIntegrationHandlers_Working(t *testing.T) {
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

    // Setup routes
    v1 := router.Group("/api/v1")
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

    t.Run("SubmitCatalogProposal", func(t *testing.T) {
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

        reqBody, err := json.Marshal(req)
        assert.NoError(t, err)

        httpReq, _ := http.NewRequest("POST", "/api/v1/integrations/catalog/proposals", bytes.NewBuffer(reqBody))
        httpReq.Header.Set("Content-Type", "application/json")

        w := httptest.NewRecorder()
        router.ServeHTTP(w, httpReq)

        assert.Equal(t, http.StatusCreated, w.Code)

        var response map[string]interface{}
        err = json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)

        assert.Contains(t, response, "data")
        data := response["data"].(map[string]interface{})
        assert.Equal(t, "prop_456", data["proposal_id"])
        assert.Equal(t, "PENDING_REVIEW", data["status"])
    })

    t.Run("AcknowledgeOrder", func(t *testing.T) {
        req := integrationRequests.OrderAcknowledgementRequest{
            OrderID:         "order_123",
            ExternalOrderID: "ext_order_456",
            PartnerID:       "partner_123",
            PartnerName:     "Logistics Partner",
            Status:          "ACCEPTED",
            AcknowledgedAt:  time.Now(),
            Notes:           "Order received and queued",
        }

        reqBody, err := json.Marshal(req)
        assert.NoError(t, err)

        httpReq, _ := http.NewRequest("POST", "/api/v1/integrations/orders/acknowledgements", bytes.NewBuffer(reqBody))
        httpReq.Header.Set("Content-Type", "application/json")

        w := httptest.NewRecorder()
        router.ServeHTTP(w, httpReq)

        assert.Equal(t, http.StatusCreated, w.Code)

        var response map[string]interface{}
        err = json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)

        assert.Contains(t, response, "data")
        data := response["data"].(map[string]interface{})
        assert.Equal(t, "order_123", data["order_id"])
        assert.Equal(t, "ACKNOWLEDGED", data["status"])
    })

    t.Run("ExportCatalog", func(t *testing.T) {
        httpReq, _ := http.NewRequest("GET", "/api/v1/integrations/catalog/exports?partner_id=partner_123&item_type=PRODUCT&page=1&page_size=10&format=JSON", nil)

        w := httptest.NewRecorder()
        router.ServeHTTP(w, httpReq)

        assert.Equal(t, http.StatusOK, w.Code)

        var response map[string]interface{}
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)

        assert.Contains(t, response, "data")
        data := response["data"].(map[string]interface{})
        assert.Contains(t, data, "export_id")
        assert.Contains(t, data, "items")
    })

    t.Run("GetProposalStatus", func(t *testing.T) {
        httpReq, _ := http.NewRequest("GET", "/api/v1/integrations/proposals/prop_123/status", nil)

        w := httptest.NewRecorder()
        router.ServeHTTP(w, httpReq)

        assert.Equal(t, http.StatusOK, w.Code)

        var response map[string]interface{}
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)

        assert.Contains(t, response, "data")
        data := response["data"].(map[string]interface{})
        assert.Equal(t, "prop_123", data["proposal_id"])
    })

    t.Run("ListIntegrationPartners", func(t *testing.T) {
        httpReq, _ := http.NewRequest("GET", "/api/v1/integrations/partners?page=1&limit=10", nil)

        w := httptest.NewRecorder()
        router.ServeHTTP(w, httpReq)

        assert.Equal(t, http.StatusOK, w.Code)

        var response map[string]interface{}
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)

        assert.Contains(t, response, "data")
        data := response["data"].([]interface{})
        assert.Len(t, data, 2) // Mock returns 2 partners
    })

    t.Run("ValidateWebhookSignature", func(t *testing.T) {
        httpReq, _ := http.NewRequest("POST", "/api/v1/integrations/webhooks/validate?partner_id=test_partner", bytes.NewBufferString("test body"))
        httpReq.Header.Set("X-Webhook-Signature", "sha256=test_signature")
        httpReq.Header.Set("X-Webhook-Timestamp", "1640995200") // Valid timestamp

        w := httptest.NewRecorder()
        router.ServeHTTP(w, httpReq)

        // This will return 401 because signature is invalid, but that's expected
        assert.Equal(t, http.StatusUnauthorized, w.Code)

        var response map[string]interface{}
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)

        assert.Contains(t, response, "error")
    })
}
