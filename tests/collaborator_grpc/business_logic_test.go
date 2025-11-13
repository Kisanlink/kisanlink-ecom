//go:build integration
// +build integration

package collaborator_grpc_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "kisanlink-ecom/proto/gen/go/collaborator/v1"
)

// TestBankingInformation tests banking information validation and handling
func TestBankingInformation(t *testing.T) {
	t.Run("Bank Account Number Validation", func(t *testing.T) {
		ctx := context.Background()
		server := setupTestServer(t)

		testCases := []struct {
			name          string
			accountNumber string
			shouldPass    bool
			errorMsg      string
		}{
			{
				name:          "Valid account number",
				accountNumber: "1234567890",
				shouldPass:    true,
			},
			{
				name:          "Valid long account number",
				accountNumber: "12345678901234567890",
				shouldPass:    true,
			},
			{
				name:          "Account number too long",
				accountNumber: strings.Repeat("1", 51),
				shouldPass:    false,
				errorMsg:      "bank account number exceeds maximum length",
			},
			{
				name:          "Account with special chars",
				accountNumber: "1234-5678-90",
				shouldPass:    false,
				errorMsg:      "bank account number contains invalid characters",
			},
			{
				name:          "Empty account number",
				accountNumber: "",
				shouldPass:    false,
				errorMsg:      "bank account number is required for vendors",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := &pb.CreateCollaboratorRequest{
					Type: pb.CollaboratorType_COLLABORATOR_TYPE_VENDOR,
					BusinessInfo: &pb.BusinessInfo{
						BusinessName:      "Test Business",
						BankAccountNumber: tc.accountNumber,
						BankIfscCode:      "SBIN0001234",
					},
				}

				_, err := server.CreateCollaborator(ctx, req)
				if tc.shouldPass {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
					if tc.errorMsg != "" {
						assert.Contains(t, err.Error(), tc.errorMsg)
					}
				}
			})
		}
	})

	t.Run("IFSC Code Validation", func(t *testing.T) {
		ctx := context.Background()
		server := setupTestServer(t)

		testCases := []struct {
			name       string
			ifscCode   string
			shouldPass bool
			errorMsg   string
		}{
			{
				name:       "Valid IFSC - SBI",
				ifscCode:   "SBIN0001234",
				shouldPass: true,
			},
			{
				name:       "Valid IFSC - HDFC",
				ifscCode:   "HDFC0001234",
				shouldPass: true,
			},
			{
				name:       "Valid IFSC - ICICI",
				ifscCode:   "ICIC0001234",
				shouldPass: true,
			},
			{
				name:       "Invalid - lowercase",
				ifscCode:   "sbin0001234",
				shouldPass: false,
				errorMsg:   "IFSC code must be uppercase",
			},
			{
				name:       "Invalid - wrong format",
				ifscCode:   "SBI00001234",
				shouldPass: false,
				errorMsg:   "invalid IFSC code format",
			},
			{
				name:       "Invalid - too short",
				ifscCode:   "SBIN001234",
				shouldPass: false,
				errorMsg:   "IFSC code must be 11 characters",
			},
			{
				name:       "Invalid - too long",
				ifscCode:   "SBIN00012345",
				shouldPass: false,
				errorMsg:   "IFSC code must be 11 characters",
			},
			{
				name:       "Invalid - fifth char not 0",
				ifscCode:   "SBIN1001234",
				shouldPass: false,
				errorMsg:   "IFSC code fifth character must be 0",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := &pb.CreateCollaboratorRequest{
					Type: pb.CollaboratorType_COLLABORATOR_TYPE_VENDOR,
					BusinessInfo: &pb.BusinessInfo{
						BusinessName:      "Test Business",
						BankAccountNumber: "1234567890",
						BankIfscCode:      tc.ifscCode,
					},
				}

				_, err := server.CreateCollaborator(ctx, req)
				if tc.shouldPass {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
					if tc.errorMsg != "" {
						assert.Contains(t, err.Error(), tc.errorMsg)
					}
				}
			})
		}
	})

	t.Run("Banking Details Required for Vendor", func(t *testing.T) {
		ctx := context.Background()
		server := setupTestServer(t)

		// Vendor without banking details should fail
		req := &pb.CreateCollaboratorRequest{
			Type: pb.CollaboratorType_COLLABORATOR_TYPE_VENDOR,
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Vendor Business",
				GstNumber:    "27AABCU9603R1ZM",
				// No banking details
			},
		}

		_, err := server.CreateCollaborator(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "banking details are required for vendors")
	})

	t.Run("Banking Details Audit Trail", func(t *testing.T) {
		ctx := context.Background()
		server, auditLog := setupServerWithAudit(t)

		// Create with initial banking details
		createReq := &pb.CreateCollaboratorRequest{
			Type: pb.CollaboratorType_COLLABORATOR_TYPE_VENDOR,
			BusinessInfo: &pb.BusinessInfo{
				BusinessName:      "Test Business",
				BankAccountNumber: "1234567890",
				BankIfscCode:      "SBIN0001234",
			},
		}

		resp, err := server.CreateCollaborator(ctx, createReq)
		require.NoError(t, err)

		// Update banking details
		updateReq := &pb.UpdateCollaboratorRequest{
			Id: resp.Id,
			BusinessInfo: &pb.BusinessInfo{
				BankAccountNumber: "9876543210",
				BankIfscCode:      "HDFC0005678",
			},
		}

		_, err = server.UpdateCollaborator(ctx, updateReq)
		require.NoError(t, err)

		// Verify audit trail
		entries := auditLog.GetBankingChangeEntries()
		assert.Len(t, entries, 2) // Create and Update

		// Check update audit
		updateEntry := entries[1]
		assert.Equal(t, "BANKING_DETAILS_UPDATE", updateEntry.Action)
		assert.Equal(t, "1234567890", updateEntry.OldBankAccount)
		assert.Equal(t, "9876543210", updateEntry.NewBankAccount)
		assert.Equal(t, "SBIN0001234", updateEntry.OldIFSC)
		assert.Equal(t, "HDFC0005678", updateEntry.NewIFSC)
	})
}

// TestStateTransitions tests collaborator state transition rules
func TestStateTransitions(t *testing.T) {
	t.Run("Active to Inactive Transition", func(t *testing.T) {
		ctx := context.Background()
		server := setupTestServer(t)

		// Create active collaborator
		createReq := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Test Business",
			},
			Status: pb.CollaboratorStatus_COLLABORATOR_STATUS_ACTIVE,
		}

		resp, err := server.CreateCollaborator(ctx, createReq)
		require.NoError(t, err)
		assert.Equal(t, pb.CollaboratorStatus_COLLABORATOR_STATUS_ACTIVE, resp.Status)

		// Deactivate
		deactivateReq := &pb.DeactivateCollaboratorRequest{
			Id:     resp.Id,
			Reason: "Test deactivation",
		}

		deactivateResp, err := server.DeactivateCollaborator(ctx, deactivateReq)
		require.NoError(t, err)
		assert.Equal(t, pb.CollaboratorStatus_COLLABORATOR_STATUS_INACTIVE, deactivateResp.Status)
	})

	t.Run("Inactive to Active Transition", func(t *testing.T) {
		ctx := context.Background()
		server := setupTestServer(t)

		// Create inactive collaborator
		createReq := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Test Business",
			},
			Status: pb.CollaboratorStatus_COLLABORATOR_STATUS_INACTIVE,
		}

		resp, err := server.CreateCollaborator(ctx, createReq)
		require.NoError(t, err)

		// Reactivate
		reactivateReq := &pb.ReactivateCollaboratorRequest{
			Id:     resp.Id,
			Reason: "Test reactivation",
		}

		reactivateResp, err := server.ReactivateCollaborator(ctx, reactivateReq)
		require.NoError(t, err)
		assert.Equal(t, pb.CollaboratorStatus_COLLABORATOR_STATUS_ACTIVE, reactivateResp.Status)
	})

	t.Run("Cannot Delete with Active Orders", func(t *testing.T) {
		ctx := context.Background()
		server := setupTestServer(t)

		// Create collaborator
		createReq := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Test Business",
			},
		}

		resp, err := server.CreateCollaborator(ctx, createReq)
		require.NoError(t, err)

		// Simulate active order
		server.AddActiveOrder(resp.Id, "ORDER-001")

		// Try to deactivate
		deactivateReq := &pb.DeactivateCollaboratorRequest{
			Id:     resp.Id,
			Reason: "Try to deactivate with active orders",
		}

		_, err = server.DeactivateCollaborator(ctx, deactivateReq)
		assert.Error(t, err)

		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.FailedPrecondition, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "cannot deactivate collaborator with active orders")
	})

	t.Run("Deactivation Checks Pending Transactions", func(t *testing.T) {
		ctx := context.Background()
		server := setupTestServer(t)

		// Create collaborator
		createReq := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Test Business",
			},
		}

		resp, err := server.CreateCollaborator(ctx, createReq)
		require.NoError(t, err)

		// Add pending transaction
		server.AddPendingTransaction(resp.Id, "TXN-001", 10000.00)

		// Try to deactivate
		deactivateReq := &pb.DeactivateCollaboratorRequest{
			Id:     resp.Id,
			Reason: "Try to deactivate with pending transactions",
		}

		_, err = server.DeactivateCollaborator(ctx, deactivateReq)
		assert.Error(t, err)

		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.FailedPrecondition, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "cannot deactivate collaborator with pending transactions")
	})

	t.Run("Status Transition Audit", func(t *testing.T) {
		ctx := context.Background()
		server, auditLog := setupServerWithAudit(t)

		// Create collaborator
		createReq := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Test Business",
			},
			Status: pb.CollaboratorStatus_COLLABORATOR_STATUS_PENDING_VERIFICATION,
		}

		resp, err := server.CreateCollaborator(ctx, createReq)
		require.NoError(t, err)

		// Verify status
		verifyReq := &pb.VerifyCollaboratorRequest{
			Id:               resp.Id,
			VerificationNote: "Documents verified",
		}

		_, err = server.VerifyCollaborator(ctx, verifyReq)
		require.NoError(t, err)

		// Check audit log
		entries := auditLog.GetStatusChangeEntries()
		assert.Len(t, entries, 1)

		entry := entries[0]
		assert.Equal(t, "STATUS_CHANGE", entry.Action)
		assert.Equal(t, "PENDING_VERIFICATION", entry.OldStatus)
		assert.Equal(t, "VERIFIED", entry.NewStatus)
		assert.Equal(t, "Documents verified", entry.Note)
	})
}

// TestDataValidation tests various data validation scenarios
func TestDataValidation(t *testing.T) {
	t.Run("Email Validation", func(t *testing.T) {
		ctx := context.Background()
		server := setupTestServer(t)

		testCases := []struct {
			email      string
			shouldPass bool
		}{
			{"valid@email.com", true},
			{"user.name@domain.co.in", true},
			{"user+tag@example.com", true},
			{"invalid.email", false},
			{"@example.com", false},
			{"user@", false},
			{"user name@example.com", false},
			{"", false},
		}

		for _, tc := range testCases {
			t.Run(tc.email, func(t *testing.T) {
				req := &pb.CreateCollaboratorRequest{
					Email: tc.email,
					BusinessInfo: &pb.BusinessInfo{
						BusinessName: "Test Business",
					},
				}

				_, err := server.CreateCollaborator(ctx, req)
				if tc.shouldPass {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "invalid email")
				}
			})
		}
	})

	t.Run("Phone Number Validation", func(t *testing.T) {
		ctx := context.Background()
		server := setupTestServer(t)

		testCases := []struct {
			phone      string
			shouldPass bool
		}{
			{"+919876543210", true},
			{"9876543210", true},
			{"+91-9876543210", true},
			{"0091-9876543210", true},
			{"+1234567890123456", false}, // Too long
			{"123", false},               // Too short
			{"abcd1234567", false},       // Letters
			{"", true},                   // Optional field
		}

		for _, tc := range testCases {
			t.Run(tc.phone, func(t *testing.T) {
				req := &pb.CreateCollaboratorRequest{
					Phone: tc.phone,
					Email: "test@example.com",
					BusinessInfo: &pb.BusinessInfo{
						BusinessName: "Test Business",
					},
				}

				_, err := server.CreateCollaborator(ctx, req)
				if tc.shouldPass {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "invalid phone")
				}
			})
		}
	})

	t.Run("Field Length Validation", func(t *testing.T) {
		ctx := context.Background()
		server := setupTestServer(t)

		testCases := []struct {
			name       string
			field      string
			value      string
			maxLength  int
			shouldPass bool
		}{
			{
				name:       "Company name within limit",
				field:      "company_name",
				value:      "Valid Company Name",
				maxLength:  200,
				shouldPass: true,
			},
			{
				name:       "Company name exceeds limit",
				field:      "company_name",
				value:      strings.Repeat("a", 201),
				maxLength:  200,
				shouldPass: false,
			},
			{
				name:       "Contact person within limit",
				field:      "contact_person",
				value:      "John Doe",
				maxLength:  100,
				shouldPass: true,
			},
			{
				name:       "Experience text within limit",
				field:      "experience",
				value:      strings.Repeat("x", 1000),
				maxLength:  5000,
				shouldPass: true,
			},
			{
				name:       "Experience text exceeds limit",
				field:      "experience",
				value:      strings.Repeat("x", 5001),
				maxLength:  5000,
				shouldPass: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := createRequestWithField(tc.field, tc.value)
				_, err := server.CreateCollaborator(ctx, req)

				if tc.shouldPass {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "exceeds maximum length")
				}
			})
		}
	})

	t.Run("Unicode and Special Characters", func(t *testing.T) {
		ctx := context.Background()
		server := setupTestServer(t)

		testCases := []struct {
			name         string
			businessName string
			shouldPass   bool
		}{
			{
				name:         "English characters",
				businessName: "ABC Enterprises",
				shouldPass:   true,
			},
			{
				name:         "Hindi characters",
				businessName: "किसान ट्रेडर्स",
				shouldPass:   true,
			},
			{
				name:         "Mixed languages",
				businessName: "Kisan किसान Traders",
				shouldPass:   true,
			},
			{
				name:         "With numbers",
				businessName: "Business 123",
				shouldPass:   true,
			},
			{
				name:         "With allowed special chars",
				businessName: "A&B Co., Ltd.",
				shouldPass:   true,
			},
			{
				name:         "SQL injection attempt",
				businessName: "'; DROP TABLE collaborators; --",
				shouldPass:   false,
			},
			{
				name:         "XSS attempt",
				businessName: "<script>alert('xss')</script>",
				shouldPass:   false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := &pb.CreateCollaboratorRequest{
					BusinessInfo: &pb.BusinessInfo{
						BusinessName: tc.businessName,
					},
				}

				_, err := server.CreateCollaborator(ctx, req)
				if tc.shouldPass {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "invalid characters")
				}
			})
		}
	})

	t.Run("Empty vs Null Value Handling", func(t *testing.T) {
		ctx := context.Background()
		server := setupTestServer(t)

		// Test with empty strings
		req1 := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName:  "Test Business",
				ContactPerson: "", // Empty string
			},
		}

		resp1, err := server.CreateCollaborator(ctx, req1)
		require.NoError(t, err)
		assert.Empty(t, resp1.BusinessInfo.ContactPerson)

		// Test with null (not set)
		req2 := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Test Business 2",
				// ContactPerson not set (null)
			},
		}

		resp2, err := server.CreateCollaborator(ctx, req2)
		require.NoError(t, err)
		assert.Empty(t, resp2.BusinessInfo.ContactPerson)

		// Verify different handling in database
		// Empty string should be stored as empty
		// Null should remain null
		collab1 := server.GetCollaboratorFromDB(resp1.Id)
		collab2 := server.GetCollaboratorFromDB(resp2.Id)

		assert.Equal(t, "", collab1.ContactPerson)
		assert.Nil(t, collab2.ContactPerson)
	})
}

// TestBusinessRuleEnforcement tests complex business rules
func TestBusinessRuleEnforcement(t *testing.T) {
	t.Run("Required Fields by Collaborator Type", func(t *testing.T) {
		ctx := context.Background()
		server := setupTestServer(t)

		testCases := []struct {
			name           string
			collabType     pb.CollaboratorType
			requiredFields map[string]bool
		}{
			{
				name:       "Vendor requirements",
				collabType: pb.CollaboratorType_COLLABORATOR_TYPE_VENDOR,
				requiredFields: map[string]bool{
					"gst_number":   true,
					"bank_details": true,
					"address":      true,
				},
			},
			{
				name:       "Buyer requirements",
				collabType: pb.CollaboratorType_COLLABORATOR_TYPE_BUYER,
				requiredFields: map[string]bool{
					"gst_number":   true,
					"bank_details": false,
					"address":      true,
				},
			},
			{
				name:       "Service Provider requirements",
				collabType: pb.CollaboratorType_COLLABORATOR_TYPE_SERVICE_PROVIDER,
				requiredFields: map[string]bool{
					"gst_number":   false,
					"bank_details": true,
					"address":      true,
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Test without required fields
				req := &pb.CreateCollaboratorRequest{
					Type: tc.collabType,
					BusinessInfo: &pb.BusinessInfo{
						BusinessName: "Test Business",
					},
				}

				_, err := server.CreateCollaborator(ctx, req)

				// Check if error messages match requirements
				if tc.requiredFields["gst_number"] {
					assert.Contains(t, err.Error(), "GST number is required")
				}
				if tc.requiredFields["bank_details"] {
					assert.Contains(t, err.Error(), "banking details are required")
				}
				if tc.requiredFields["address"] {
					assert.Contains(t, err.Error(), "address is required")
				}
			})
		}
	})

	t.Run("External ID Generation and Uniqueness", func(t *testing.T) {
		ctx := context.Background()
		server := setupTestServer(t)

		externalIds := make(map[string]bool)

		// Create multiple collaborators
		for i := 0; i < 100; i++ {
			req := &pb.CreateCollaboratorRequest{
				BusinessInfo: &pb.BusinessInfo{
					BusinessName: fmt.Sprintf("Business %d", i),
				},
			}

			resp, err := server.CreateCollaborator(ctx, req)
			require.NoError(t, err)

			// Check external ID format
			assert.NotEmpty(t, resp.ExternalId)
			assert.Regexp(t, "^COL-[A-Z0-9]{8}$", resp.ExternalId)

			// Check uniqueness
			assert.False(t, externalIds[resp.ExternalId], "Duplicate external ID found")
			externalIds[resp.ExternalId] = true
		}
	})

	t.Run("FPO Customization vs Master Data", func(t *testing.T) {
		ctx := context.Background()
		server := setupTestServer(t)

		gstNumber := "27AABCU9603R1ZM"

		// FPO-A creates collaborator
		ctxFpoA := createAuthContext("user1", "FPO-A", []string{"FPO_USER"})
		reqA := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				GstNumber:     gstNumber,
				BusinessName:  "Master Business Name",
				ContactPerson: "John Doe",
				ContactNumber: "9876543210",
				BusinessEmail: "john@fpoa.com",
				CustomField1:  "FPO-A specific data",
			},
		}

		respA, err := server.CreateCollaborator(ctxFpoA, reqA)
		require.NoError(t, err)

		// FPO-B creates with same GST
		ctxFpoB := createAuthContext("user2", "FPO-B", []string{"FPO_USER"})
		reqB := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				GstNumber:     gstNumber,
				BusinessName:  "Different Name", // Should be overridden by master
				ContactPerson: "Jane Smith",     // Local customization
				ContactNumber: "9876543211",     // Local customization
				BusinessEmail: "jane@fpob.com",  // Local customization
				CustomField1:  "FPO-B specific data",
			},
		}

		respB, err := server.CreateCollaborator(ctxFpoB, reqB)
		require.NoError(t, err)

		// Verify master fields are same
		assert.Equal(t, respA.BusinessInfo.BusinessName, respB.BusinessInfo.BusinessName)
		assert.Equal(t, respA.BusinessInfo.GstNumber, respB.BusinessInfo.GstNumber)
		assert.Equal(t, respA.MasterCollaboratorId, respB.MasterCollaboratorId)

		// Verify local fields are different
		assert.NotEqual(t, respA.BusinessInfo.ContactPerson, respB.BusinessInfo.ContactPerson)
		assert.NotEqual(t, respA.BusinessInfo.ContactNumber, respB.BusinessInfo.ContactNumber)
		assert.NotEqual(t, respA.BusinessInfo.BusinessEmail, respB.BusinessInfo.BusinessEmail)
		assert.NotEqual(t, respA.BusinessInfo.CustomField1, respB.BusinessInfo.CustomField1)
	})

	t.Run("Verification Workflow", func(t *testing.T) {
		ctx := context.Background()
		server := setupTestServer(t)

		// Create unverified collaborator
		createReq := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Test Business",
				GstNumber:    "27AABCU9603R1ZM",
			},
			Status: pb.CollaboratorStatus_COLLABORATOR_STATUS_PENDING_VERIFICATION,
		}

		resp, err := server.CreateCollaborator(ctx, createReq)
		require.NoError(t, err)
		assert.Equal(t, pb.CollaboratorStatus_COLLABORATOR_STATUS_PENDING_VERIFICATION, resp.Status)
		assert.False(t, resp.BusinessInfo.IsVerified)

		// Try to perform restricted action while unverified
		// For example, creating orders might be restricted
		orderReq := &pb.CreateOrderRequest{
			CollaboratorId: resp.Id,
			Amount:         10000,
		}

		_, err = server.CreateOrder(ctx, orderReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "collaborator must be verified")

		// Verify collaborator
		verifyReq := &pb.VerifyCollaboratorRequest{
			Id:               resp.Id,
			VerificationNote: "Documents checked",
		}

		verifyResp, err := server.VerifyCollaborator(ctx, verifyReq)
		require.NoError(t, err)
		assert.Equal(t, pb.CollaboratorStatus_COLLABORATOR_STATUS_VERIFIED, verifyResp.Status)
		assert.True(t, verifyResp.BusinessInfo.IsVerified)
		assert.NotNil(t, verifyResp.BusinessInfo.VerifiedAt)

		// Now order creation should work
		_, err = server.CreateOrder(ctx, orderReq)
		assert.NoError(t, err)
	})
}

// Helper functions

func createRequestWithField(field, value string) *pb.CreateCollaboratorRequest {
	req := &pb.CreateCollaboratorRequest{
		Email: "test@example.com",
		BusinessInfo: &pb.BusinessInfo{
			BusinessName: "Default Business",
		},
	}

	switch field {
	case "company_name":
		req.BusinessInfo.BusinessName = value
	case "contact_person":
		req.BusinessInfo.ContactPerson = value
	case "experience":
		req.BusinessInfo.Experience = value
	}

	return req
}
