//go:build integration
// +build integration

package collaborator_grpc_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "kisanlink-ecom/proto/gen/go/collaborator/v1"
)

// MockAuthContext simulates authentication context
type MockAuthContext struct {
	UserID      string
	FPOId       string
	Roles       []string
	Permissions []string
	IsAdmin     bool
	IsValid     bool
}

// TestFPOIsolation tests that FPOs cannot access each other's data
func TestFPOIsolation(t *testing.T) {
	t.Run("FPO Cannot Access Other FPO Collaborators", func(t *testing.T) {
		server := setupAuthTestServer(t)

		// FPO-A creates a collaborator
		ctxFpoA := createAuthContext("user1", "FPO-A", []string{"FPO_USER"})
		req := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName:  "FPO-A Business",
				GstNumber:     "27AABCU9603R1ZM",
				ContactPerson: "John Doe",
			},
		}

		respA, err := server.CreateCollaborator(ctxFpoA, req)
		require.NoError(t, err)

		// FPO-B tries to access FPO-A's collaborator
		ctxFpoB := createAuthContext("user2", "FPO-B", []string{"FPO_USER"})
		getReq := &pb.GetCollaboratorRequest{
			Id: respA.Id,
		}

		_, err = server.GetCollaborator(ctxFpoB, getReq)
		assert.Error(t, err)

		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "access denied")
	})

	t.Run("FPO Can Only List Own Collaborators", func(t *testing.T) {
		server := setupAuthTestServer(t)

		// FPO-A creates collaborators
		ctxFpoA := createAuthContext("user1", "FPO-A", []string{"FPO_USER"})
		for i := 0; i < 3; i++ {
			req := &pb.CreateCollaboratorRequest{
				BusinessInfo: &pb.BusinessInfo{
					BusinessName: fmt.Sprintf("FPO-A Business %d", i),
				},
			}
			_, err := server.CreateCollaborator(ctxFpoA, req)
			require.NoError(t, err)
		}

		// FPO-B creates collaborators
		ctxFpoB := createAuthContext("user2", "FPO-B", []string{"FPO_USER"})
		for i := 0; i < 2; i++ {
			req := &pb.CreateCollaboratorRequest{
				BusinessInfo: &pb.BusinessInfo{
					BusinessName: fmt.Sprintf("FPO-B Business %d", i),
				},
			}
			_, err := server.CreateCollaborator(ctxFpoB, req)
			require.NoError(t, err)
		}

		// FPO-A lists collaborators - should only see own
		listReqA := &pb.ListCollaboratorsRequest{}
		listRespA, err := server.ListCollaborators(ctxFpoA, listReqA)
		require.NoError(t, err)
		assert.Equal(t, 3, len(listRespA.Collaborators))

		// Verify all belong to FPO-A
		for _, collab := range listRespA.Collaborators {
			assert.Contains(t, collab.BusinessInfo.BusinessName, "FPO-A")
		}

		// FPO-B lists collaborators - should only see own
		listReqB := &pb.ListCollaboratorsRequest{}
		listRespB, err := server.ListCollaborators(ctxFpoB, listReqB)
		require.NoError(t, err)
		assert.Equal(t, 2, len(listRespB.Collaborators))

		// Verify all belong to FPO-B
		for _, collab := range listRespB.Collaborators {
			assert.Contains(t, collab.BusinessInfo.BusinessName, "FPO-B")
		}
	})

	t.Run("FPO Cannot Update Other FPO Collaborators", func(t *testing.T) {
		server := setupAuthTestServer(t)

		// FPO-A creates a collaborator
		ctxFpoA := createAuthContext("user1", "FPO-A", []string{"FPO_USER"})
		createReq := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "FPO-A Business",
			},
		}

		resp, err := server.CreateCollaborator(ctxFpoA, createReq)
		require.NoError(t, err)

		// FPO-B tries to update FPO-A's collaborator
		ctxFpoB := createAuthContext("user2", "FPO-B", []string{"FPO_USER"})
		updateReq := &pb.UpdateCollaboratorRequest{
			Id: resp.Id,
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Hijacked Business",
			},
		}

		_, err = server.UpdateCollaborator(ctxFpoB, updateReq)
		assert.Error(t, err)

		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, statusErr.Code())
	})

	t.Run("FPO Cannot Deactivate Other FPO Collaborators", func(t *testing.T) {
		server := setupAuthTestServer(t)

		// FPO-A creates a collaborator
		ctxFpoA := createAuthContext("user1", "FPO-A", []string{"FPO_USER"})
		createReq := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "FPO-A Business",
			},
		}

		resp, err := server.CreateCollaborator(ctxFpoA, createReq)
		require.NoError(t, err)

		// FPO-B tries to deactivate FPO-A's collaborator
		ctxFpoB := createAuthContext("user2", "FPO-B", []string{"FPO_USER"})
		deactivateReq := &pb.DeactivateCollaboratorRequest{
			Id:     resp.Id,
			Reason: "Malicious deactivation",
		}

		_, err = server.DeactivateCollaborator(ctxFpoB, deactivateReq)
		assert.Error(t, err)

		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, statusErr.Code())
	})
}

// TestAdminAccess tests admin user access privileges
func TestAdminAccess(t *testing.T) {
	t.Run("Admin Can Access All Collaborators", func(t *testing.T) {
		server := setupAuthTestServer(t)

		// FPO-A creates a collaborator
		ctxFpoA := createAuthContext("user1", "FPO-A", []string{"FPO_USER"})
		reqA := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "FPO-A Business",
			},
		}
		respA, err := server.CreateCollaborator(ctxFpoA, reqA)
		require.NoError(t, err)

		// FPO-B creates a collaborator
		ctxFpoB := createAuthContext("user2", "FPO-B", []string{"FPO_USER"})
		reqB := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "FPO-B Business",
			},
		}
		respB, err := server.CreateCollaborator(ctxFpoB, reqB)
		require.NoError(t, err)

		// Admin can access both
		ctxAdmin := createAuthContext("admin", "", []string{"ADMIN"})

		// Access FPO-A's collaborator
		getReqA := &pb.GetCollaboratorRequest{Id: respA.Id}
		getRespA, err := server.GetCollaborator(ctxAdmin, getReqA)
		require.NoError(t, err)
		assert.Equal(t, respA.Id, getRespA.Id)

		// Access FPO-B's collaborator
		getReqB := &pb.GetCollaboratorRequest{Id: respB.Id}
		getRespB, err := server.GetCollaborator(ctxAdmin, getReqB)
		require.NoError(t, err)
		assert.Equal(t, respB.Id, getRespB.Id)
	})

	t.Run("Admin Can List All Collaborators", func(t *testing.T) {
		server := setupAuthTestServer(t)

		// Multiple FPOs create collaborators
		for fpoIdx := 1; fpoIdx <= 3; fpoIdx++ {
			ctx := createAuthContext(
				fmt.Sprintf("user%d", fpoIdx),
				fmt.Sprintf("FPO-%d", fpoIdx),
				[]string{"FPO_USER"},
			)

			for i := 0; i < 2; i++ {
				req := &pb.CreateCollaboratorRequest{
					BusinessInfo: &pb.BusinessInfo{
						BusinessName: fmt.Sprintf("FPO-%d Business %d", fpoIdx, i),
					},
				}
				_, err := server.CreateCollaborator(ctx, req)
				require.NoError(t, err)
			}
		}

		// Admin lists all collaborators
		ctxAdmin := createAuthContext("admin", "", []string{"ADMIN"})
		listReq := &pb.ListCollaboratorsRequest{}
		listResp, err := server.ListCollaborators(ctxAdmin, listReq)
		require.NoError(t, err)

		// Should see all 6 collaborators
		assert.Equal(t, 6, len(listResp.Collaborators))
	})
}

// TestDeactivationPermissions tests deactivation permission requirements
func TestDeactivationPermissions(t *testing.T) {
	t.Run("Regular User Cannot Deactivate", func(t *testing.T) {
		server := setupAuthTestServer(t)

		// Create collaborator as regular user
		ctx := createAuthContext("user1", "FPO-A", []string{"USER"})
		createReq := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Test Business",
			},
		}

		resp, err := server.CreateCollaborator(ctx, createReq)
		require.NoError(t, err)

		// Try to deactivate as regular user
		deactivateReq := &pb.DeactivateCollaboratorRequest{
			Id:     resp.Id,
			Reason: "Test deactivation",
		}

		_, err = server.DeactivateCollaborator(ctx, deactivateReq)
		assert.Error(t, err)

		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "insufficient permissions")
	})

	t.Run("Manager Can Deactivate Own FPO Collaborators", func(t *testing.T) {
		server := setupAuthTestServer(t)

		// Create collaborator
		ctxUser := createAuthContext("user1", "FPO-A", []string{"USER"})
		createReq := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Test Business",
			},
		}

		resp, err := server.CreateCollaborator(ctxUser, createReq)
		require.NoError(t, err)

		// Manager from same FPO can deactivate
		ctxManager := createAuthContext("manager1", "FPO-A", []string{"MANAGER"})
		deactivateReq := &pb.DeactivateCollaboratorRequest{
			Id:     resp.Id,
			Reason: "Manager deactivation",
		}

		deactivateResp, err := server.DeactivateCollaborator(ctxManager, deactivateReq)
		require.NoError(t, err)
		assert.Equal(t, pb.CollaboratorStatus_COLLABORATOR_STATUS_INACTIVE, deactivateResp.Status)
	})

	t.Run("Admin Can Deactivate Any Collaborator", func(t *testing.T) {
		server := setupAuthTestServer(t)

		// Create collaborator in FPO-A
		ctxFpoA := createAuthContext("user1", "FPO-A", []string{"USER"})
		createReq := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Test Business",
			},
		}

		resp, err := server.CreateCollaborator(ctxFpoA, createReq)
		require.NoError(t, err)

		// Admin can deactivate
		ctxAdmin := createAuthContext("admin", "", []string{"ADMIN"})
		deactivateReq := &pb.DeactivateCollaboratorRequest{
			Id:     resp.Id,
			Reason: "Admin deactivation",
		}

		deactivateResp, err := server.DeactivateCollaborator(ctxAdmin, deactivateReq)
		require.NoError(t, err)
		assert.Equal(t, pb.CollaboratorStatus_COLLABORATOR_STATUS_INACTIVE, deactivateResp.Status)
	})
}

// TestSensitiveFieldAccess tests access control for sensitive fields
func TestSensitiveFieldAccess(t *testing.T) {
	t.Run("Regular User Cannot See Sensitive Fields", func(t *testing.T) {
		server := setupAuthTestServer(t)

		// Admin creates collaborator with sensitive data
		ctxAdmin := createAuthContext("admin", "FPO-A", []string{"ADMIN"})
		createReq := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName:      "Test Business",
				GstNumber:         "27AABCU9603R1ZM",
				PanNumber:         "AABCU9603R",
				BankAccountNumber: "1234567890",
				BankIfscCode:      "SBIN0001234",
			},
		}

		resp, err := server.CreateCollaborator(ctxAdmin, createReq)
		require.NoError(t, err)

		// Regular user from same FPO gets collaborator
		ctxUser := createAuthContext("user1", "FPO-A", []string{"USER"})
		getReq := &pb.GetCollaboratorRequest{Id: resp.Id}
		getResp, err := server.GetCollaborator(ctxUser, getReq)
		require.NoError(t, err)

		// Sensitive fields should be masked or empty
		assert.Equal(t, "27XXXXX03R1ZM", getResp.BusinessInfo.GstNumber) // Partially masked
		assert.Equal(t, "XXXXX9603R", getResp.BusinessInfo.PanNumber)    // Partially masked
		assert.Empty(t, getResp.BusinessInfo.BankAccountNumber)          // Completely hidden
		assert.Empty(t, getResp.BusinessInfo.BankIfscCode)               // Completely hidden
	})

	t.Run("Manager Can See Some Sensitive Fields", func(t *testing.T) {
		server := setupAuthTestServer(t)

		// Create collaborator with sensitive data
		ctxAdmin := createAuthContext("admin", "FPO-A", []string{"ADMIN"})
		createReq := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName:      "Test Business",
				GstNumber:         "27AABCU9603R1ZM",
				PanNumber:         "AABCU9603R",
				BankAccountNumber: "1234567890",
				BankIfscCode:      "SBIN0001234",
			},
		}

		resp, err := server.CreateCollaborator(ctxAdmin, createReq)
		require.NoError(t, err)

		// Manager from same FPO gets collaborator
		ctxManager := createAuthContext("manager1", "FPO-A", []string{"MANAGER"})
		getReq := &pb.GetCollaboratorRequest{Id: resp.Id}
		getResp, err := server.GetCollaborator(ctxManager, getReq)
		require.NoError(t, err)

		// Manager can see GST and PAN but not bank details
		assert.Equal(t, "27AABCU9603R1ZM", getResp.BusinessInfo.GstNumber)
		assert.Equal(t, "AABCU9603R", getResp.BusinessInfo.PanNumber)
		assert.Equal(t, "******7890", getResp.BusinessInfo.BankAccountNumber) // Partially masked
		assert.Equal(t, "SBIN0001234", getResp.BusinessInfo.BankIfscCode)
	})

	t.Run("Admin Can See All Sensitive Fields", func(t *testing.T) {
		server := setupAuthTestServer(t)

		// Create collaborator with sensitive data
		ctxAdmin := createAuthContext("admin", "", []string{"ADMIN"})
		createReq := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName:      "Test Business",
				GstNumber:         "27AABCU9603R1ZM",
				PanNumber:         "AABCU9603R",
				BankAccountNumber: "1234567890",
				BankIfscCode:      "SBIN0001234",
			},
		}

		resp, err := server.CreateCollaborator(ctxAdmin, createReq)
		require.NoError(t, err)

		// Admin gets collaborator
		getReq := &pb.GetCollaboratorRequest{Id: resp.Id}
		getResp, err := server.GetCollaborator(ctxAdmin, getReq)
		require.NoError(t, err)

		// Admin can see all fields unmasked
		assert.Equal(t, "27AABCU9603R1ZM", getResp.BusinessInfo.GstNumber)
		assert.Equal(t, "AABCU9603R", getResp.BusinessInfo.PanNumber)
		assert.Equal(t, "1234567890", getResp.BusinessInfo.BankAccountNumber)
		assert.Equal(t, "SBIN0001234", getResp.BusinessInfo.BankIfscCode)
	})

	t.Run("Finance Role Can See Banking Details", func(t *testing.T) {
		server := setupAuthTestServer(t)

		// Create collaborator with sensitive data
		ctxAdmin := createAuthContext("admin", "FPO-A", []string{"ADMIN"})
		createReq := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName:      "Test Business",
				GstNumber:         "27AABCU9603R1ZM",
				BankAccountNumber: "1234567890",
				BankIfscCode:      "SBIN0001234",
			},
		}

		resp, err := server.CreateCollaborator(ctxAdmin, createReq)
		require.NoError(t, err)

		// Finance user from same FPO gets collaborator
		ctxFinance := createAuthContext("finance1", "FPO-A", []string{"FINANCE"})
		getReq := &pb.GetCollaboratorRequest{Id: resp.Id}
		getResp, err := server.GetCollaborator(ctxFinance, getReq)
		require.NoError(t, err)

		// Finance role can see banking details
		assert.Equal(t, "1234567890", getResp.BusinessInfo.BankAccountNumber)
		assert.Equal(t, "SBIN0001234", getResp.BusinessInfo.BankIfscCode)
	})
}

// TestTokenValidation tests JWT token validation
func TestTokenValidation(t *testing.T) {
	t.Run("Missing Token", func(t *testing.T) {
		server := setupAuthTestServer(t)

		// Create context without token
		ctx := context.Background()

		req := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Test Business",
			},
		}

		_, err := server.CreateCollaborator(ctx, req)
		assert.Error(t, err)

		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "missing token")
	})

	t.Run("Invalid Token", func(t *testing.T) {
		server := setupAuthTestServer(t)

		// Create context with invalid token
		ctx := metadata.NewIncomingContext(
			context.Background(),
			metadata.Pairs("authorization", "Bearer invalid-token"),
		)

		req := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Test Business",
			},
		}

		_, err := server.CreateCollaborator(ctx, req)
		assert.Error(t, err)

		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "invalid token")
	})

	t.Run("Expired Token", func(t *testing.T) {
		server := setupAuthTestServer(t)

		// Create context with expired token
		ctx := createExpiredAuthContext("user1", "FPO-A", []string{"USER"})

		req := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Test Business",
			},
		}

		_, err := server.CreateCollaborator(ctx, req)
		assert.Error(t, err)

		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "token expired")
	})
}

// TestAuditLogging tests audit logging for sensitive operations
func TestAuditLogging(t *testing.T) {
	t.Run("Create Operation Audit", func(t *testing.T) {
		server, auditLog := setupServerWithAudit(t)

		ctx := createAuthContext("user1", "FPO-A", []string{"USER"})
		req := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Test Business",
				GstNumber:    "27AABCU9603R1ZM",
			},
		}

		resp, err := server.CreateCollaborator(ctx, req)
		require.NoError(t, err)

		// Verify audit log entry
		entry := auditLog.GetLastEntry()
		assert.NotNil(t, entry)
		assert.Equal(t, "CREATE_COLLABORATOR", entry.Action)
		assert.Equal(t, "user1", entry.UserID)
		assert.Equal(t, resp.Id, entry.ResourceID)
		assert.Contains(t, entry.Details, "gst_number")
	})

	t.Run("Update Operation Audit", func(t *testing.T) {
		server, auditLog := setupServerWithAudit(t)

		// Create collaborator
		ctx := createAuthContext("user1", "FPO-A", []string{"USER"})
		createReq := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Original Business",
			},
		}

		resp, err := server.CreateCollaborator(ctx, createReq)
		require.NoError(t, err)

		// Update collaborator
		updateReq := &pb.UpdateCollaboratorRequest{
			Id: resp.Id,
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Updated Business",
			},
		}

		_, err = server.UpdateCollaborator(ctx, updateReq)
		require.NoError(t, err)

		// Verify audit log entry for update
		entry := auditLog.GetLastEntry()
		assert.Equal(t, "UPDATE_COLLABORATOR", entry.Action)
		assert.Equal(t, "user1", entry.UserID)
		assert.Equal(t, resp.Id, entry.ResourceID)
		assert.Contains(t, entry.Changes, "business_name")
		assert.Equal(t, "Original Business", entry.Changes["business_name"]["old"])
		assert.Equal(t, "Updated Business", entry.Changes["business_name"]["new"])
	})

	t.Run("Deactivation Audit", func(t *testing.T) {
		server, auditLog := setupServerWithAudit(t)

		// Create collaborator
		ctxUser := createAuthContext("user1", "FPO-A", []string{"USER"})
		createReq := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Test Business",
			},
		}

		resp, err := server.CreateCollaborator(ctxUser, createReq)
		require.NoError(t, err)

		// Deactivate as manager
		ctxManager := createAuthContext("manager1", "FPO-A", []string{"MANAGER"})
		deactivateReq := &pb.DeactivateCollaboratorRequest{
			Id:     resp.Id,
			Reason: "Test deactivation reason",
		}

		_, err = server.DeactivateCollaborator(ctxManager, deactivateReq)
		require.NoError(t, err)

		// Verify audit log entry
		entry := auditLog.GetLastEntry()
		assert.Equal(t, "DEACTIVATE_COLLABORATOR", entry.Action)
		assert.Equal(t, "manager1", entry.UserID)
		assert.Equal(t, resp.Id, entry.ResourceID)
		assert.Equal(t, "Test deactivation reason", entry.Reason)
	})

	t.Run("Failed Access Attempt Audit", func(t *testing.T) {
		server, auditLog := setupServerWithAudit(t)

		// FPO-A creates collaborator
		ctxFpoA := createAuthContext("user1", "FPO-A", []string{"USER"})
		createReq := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "FPO-A Business",
			},
		}

		resp, err := server.CreateCollaborator(ctxFpoA, createReq)
		require.NoError(t, err)

		// FPO-B tries to access (should fail)
		ctxFpoB := createAuthContext("user2", "FPO-B", []string{"USER"})
		getReq := &pb.GetCollaboratorRequest{Id: resp.Id}

		_, err = server.GetCollaborator(ctxFpoB, getReq)
		assert.Error(t, err)

		// Verify security audit log entry
		entry := auditLog.GetSecurityEntry()
		assert.NotNil(t, entry)
		assert.Equal(t, "UNAUTHORIZED_ACCESS_ATTEMPT", entry.Action)
		assert.Equal(t, "user2", entry.UserID)
		assert.Equal(t, resp.Id, entry.ResourceID)
		assert.Equal(t, "FPO-B", entry.AttemptedFPO)
		assert.Equal(t, "FPO-A", entry.ResourceFPO)
	})
}

// Helper functions

func createAuthContext(userID, fpoID string, roles []string) context.Context {
	// Create context with authentication metadata
	token := generateMockToken(userID, fpoID, roles)
	return metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", token)),
	)
}

func createExpiredAuthContext(userID, fpoID string, roles []string) context.Context {
	// Create context with expired token
	token := generateExpiredMockToken(userID, fpoID, roles)
	return metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", token)),
	)
}

func generateMockToken(userID, fpoID string, roles []string) string {
	// Generate mock JWT token for testing
	return fmt.Sprintf("mock-token-%s-%s-%v", userID, fpoID, roles)
}

func generateExpiredMockToken(userID, fpoID string, roles []string) string {
	// Generate expired mock JWT token for testing
	return fmt.Sprintf("expired-token-%s-%s-%v", userID, fpoID, roles)
}

func setupAuthTestServer(t *testing.T) *CollaboratorGRPCServer {
	// Setup test server with auth interceptor
	return &CollaboratorGRPCServer{
		authEnabled: true,
	}
}

func setupServerWithAudit(t *testing.T) (*CollaboratorGRPCServer, *MockAuditLog) {
	auditLog := &MockAuditLog{}
	server := &CollaboratorGRPCServer{
		authEnabled: true,
		auditLog:    auditLog,
	}
	return server, auditLog
}
