package collaborator_grpc_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	collabModel "github.com/Kisanlink/kisanlink-ecom/entities/models/collaborator"
	"github.com/Kisanlink/kisanlink-ecom/internal/aaa"
	pb "github.com/Kisanlink/kisanlink-ecom/proto/gen/go/collaborator/v1"
)

const validTestGST = "27AABCU9603R1ZN"

// TestCreateCollaborator_HappyPath tests successful collaborator creation
func TestCreateCollaborator_HappyPath(t *testing.T) {
	t.Run("Create with all fields", func(t *testing.T) {
		tc := setupTestContext(t)
		defer cleanupTestDB(t, tc.DB)

		ctx := testUserContext("user-1", 1, "USER")
		gst := validTestGST

		req := &pb.CreateCollaboratorRequest{
			UserId:         "user-1",
			Username:       "testuser",
			Email:          "test@example.com",
			FirstName:      "Test",
			LastName:       "User",
			Type:           pb.CollaboratorType_COLLABORATOR_TYPE_VENDOR,
			Phone:          strPtr("+911234567890"),
			Bio:            strPtr("Test bio"),
			ProfilePicture: strPtr("https://example.com/avatar.jpg"),
			BusinessInfo: &pb.CreateBusinessInfoRequest{
				GstNumber:       &gst,
				BusinessName:    "Test Business",
				BusinessType:    pb.BusinessType_BUSINESS_TYPE_PROPRIETORSHIP,
				BusinessPhone:   strPtr("+911234567890"),
				BusinessEmail:   strPtr("business@example.com"),
				BusinessWebsite: strPtr("https://example.com"),
				BusinessLicense: strPtr("LIC123456"),
			},
			Address: &pb.CreateAddressRequest{
				Line1:      "123 Test St",
				Line2:      strPtr("Suite 100"),
				City:       "Mumbai",
				State:      "Maharashtra",
				Country:    "India",
				PostalCode: "400001",
				Type:       pb.AddressType_ADDRESS_TYPE_BUSINESS,
			},
		}

		resp, err := tc.Handler.CreateCollaborator(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Collaborator)

		// Verify response
		assert.Equal(t, "user-1", resp.Collaborator.UserId)
		assert.Equal(t, "testuser", resp.Collaborator.Username)
		assert.Equal(t, "test@example.com", resp.Collaborator.Email)
		assert.NotEmpty(t, resp.Collaborator.Id)
		assert.Equal(t, pb.CollaboratorStatus_COLLABORATOR_STATUS_PENDING_VERIFICATION, resp.Collaborator.Status)

		// Verify business info
		assert.NotNil(t, resp.Collaborator.BusinessInfo)
		assert.Equal(t, gst, *resp.Collaborator.BusinessInfo.GstNumber)
		assert.Equal(t, "Test Business", resp.Collaborator.BusinessInfo.BusinessName)

		// Verify AAA address was created
		assert.Equal(t, 1, len(tc.MockAAA.addresses))
	})

	t.Run("Create with minimal fields", func(t *testing.T) {
		tc := setupTestContext(t)
		defer cleanupTestDB(t, tc.DB)

		ctx := testUserContext("user-2", 1, "USER")

		req := &pb.CreateCollaboratorRequest{
			UserId:    "user-2",
			Username:  "minimaluser",
			Email:     "minimal@example.com",
			FirstName: "Min",
			LastName:  "User",
			Type:      pb.CollaboratorType_COLLABORATOR_TYPE_BUYER,
		}

		resp, err := tc.Handler.CreateCollaborator(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "user-2", resp.Collaborator.UserId)
	})
}

// TestCreateCollaborator_ValidationErrors tests request validation
func TestCreateCollaborator_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		req         *pb.CreateCollaboratorRequest
		expectedErr codes.Code
		errContains string
	}{
		{
			name: "Missing user_id",
			req: &pb.CreateCollaboratorRequest{
				Username:  "testuser",
				Email:     "test@example.com",
				FirstName: "Test",
				LastName:  "User",
				Type:      pb.CollaboratorType_COLLABORATOR_TYPE_VENDOR,
			},
			expectedErr: codes.InvalidArgument,
			errContains: "user_id is required",
		},
		{
			name: "Missing username",
			req: &pb.CreateCollaboratorRequest{
				UserId:    "user-1",
				Email:     "test@example.com",
				FirstName: "Test",
				LastName:  "User",
				Type:      pb.CollaboratorType_COLLABORATOR_TYPE_VENDOR,
			},
			expectedErr: codes.InvalidArgument,
			errContains: "username is required",
		},
		{
			name: "Missing email",
			req: &pb.CreateCollaboratorRequest{
				UserId:    "user-1",
				Username:  "testuser",
				FirstName: "Test",
				LastName:  "User",
				Type:      pb.CollaboratorType_COLLABORATOR_TYPE_VENDOR,
			},
			expectedErr: codes.InvalidArgument,
			errContains: "email is required",
		},
		{
			name: "Invalid email format",
			req: &pb.CreateCollaboratorRequest{
				UserId:    "user-1",
				Username:  "testuser",
				Email:     "invalid-email",
				FirstName: "Test",
				LastName:  "User",
				Type:      pb.CollaboratorType_COLLABORATOR_TYPE_VENDOR,
			},
			expectedErr: codes.InvalidArgument,
			errContains: "invalid email format",
		},
		{
			name: "Missing first name",
			req: &pb.CreateCollaboratorRequest{
				UserId:   "user-1",
				Username: "testuser",
				Email:    "test@example.com",
				LastName: "User",
				Type:     pb.CollaboratorType_COLLABORATOR_TYPE_VENDOR,
			},
			expectedErr: codes.InvalidArgument,
			errContains: "first_name is required",
		},
		{
			name: "Missing last name",
			req: &pb.CreateCollaboratorRequest{
				UserId:    "user-1",
				Username:  "testuser",
				Email:     "test@example.com",
				FirstName: "Test",
				Type:      pb.CollaboratorType_COLLABORATOR_TYPE_VENDOR,
			},
			expectedErr: codes.InvalidArgument,
			errContains: "last_name is required",
		},
		{
			name: "Missing type",
			req: &pb.CreateCollaboratorRequest{
				UserId:    "user-1",
				Username:  "testuser",
				Email:     "test@example.com",
				FirstName: "Test",
				LastName:  "User",
				Type:      pb.CollaboratorType_COLLABORATOR_TYPE_UNSPECIFIED,
			},
			expectedErr: codes.InvalidArgument,
			errContains: "type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := setupTestContext(t)
			defer cleanupTestDB(t, tc.DB)

			ctx := testUserContext("user-1", 1, "USER")

			resp, err := tc.Handler.CreateCollaborator(ctx, tt.req)
			require.Error(t, err)
			require.Nil(t, resp)

			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, tt.expectedErr, st.Code())
			assert.Contains(t, st.Message(), tt.errContains)
		})
	}
}

// TestCreateCollaborator_GSTValidation tests GST validation (P0-5)
func TestCreateCollaborator_GSTValidation(t *testing.T) {
	t.Run("Invalid GST format rejected", func(t *testing.T) {
		tc := setupTestContext(t)
		defer cleanupTestDB(t, tc.DB)

		// Setup mock to fail validation
		tc.MockGST.SetValidateFunc(func(_ string) (string, error) {
			return "", fmt.Errorf("invalid GST format")
		})

		ctx := testUserContext("user-1", 1, "USER")
		invalidGST := "INVALID"

		req := &pb.CreateCollaboratorRequest{
			UserId:    "user-1",
			Username:  "testuser",
			Email:     "test@example.com",
			FirstName: "Test",
			LastName:  "User",
			Type:      pb.CollaboratorType_COLLABORATOR_TYPE_VENDOR,
			BusinessInfo: &pb.CreateBusinessInfoRequest{
				GstNumber:    &invalidGST,
				BusinessName: "Test Business",
				BusinessType: pb.BusinessType_BUSINESS_TYPE_PROPRIETORSHIP,
			},
		}

		resp, err := tc.Handler.CreateCollaborator(ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)

		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "invalid GST")
	})

	t.Run("Valid GST normalized", func(t *testing.T) {
		tc := setupTestContext(t)
		defer cleanupTestDB(t, tc.DB)

		ctx := testUserContext("user-1", 1, "USER")
		gst := validTestGST

		req := &pb.CreateCollaboratorRequest{
			UserId:    "user-1",
			Username:  "testuser",
			Email:     "test@example.com",
			FirstName: "Test",
			LastName:  "User",
			Type:      pb.CollaboratorType_COLLABORATOR_TYPE_VENDOR,
			BusinessInfo: &pb.CreateBusinessInfoRequest{
				GstNumber:    &gst,
				BusinessName: "Test Business",
				BusinessType: pb.BusinessType_BUSINESS_TYPE_PROPRIETORSHIP,
			},
		}

		resp, err := tc.Handler.CreateCollaborator(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, gst, *resp.Collaborator.BusinessInfo.GstNumber)
	})
}

// TestCreateCollaborator_GSTDeduplication tests GST deduplication (P0-1)
func TestCreateCollaborator_GSTDeduplication(t *testing.T) {
	t.Run("Duplicate GST returns existing collaborator", func(t *testing.T) {
		tc := setupTestContext(t)
		defer cleanupTestDB(t, tc.DB)

		ctx := testUserContext("user-1", 1, "USER")
		gst := validTestGST

		// Create first collaborator
		existing := createTestCollaborator(tc.DB, "user-existing", gst, 1)

		// Setup mock to indicate GST exists
		existingID := uint64(1) // Mock ID since existing.ID is a string
		if existing.ID != "" {
			// Parse ID if it's numeric
			if id, err := strconv.ParseUint(existing.ID, 10, 64); err == nil {
				existingID = id
			}
		}
		tc.MockGST.SetExistingCollaborator(gst, existingID)

		// Try to create another with same GST
		req := &pb.CreateCollaboratorRequest{
			UserId:    "user-2",
			Username:  "newuser",
			Email:     "new@example.com",
			FirstName: "New",
			LastName:  "User",
			Type:      pb.CollaboratorType_COLLABORATOR_TYPE_VENDOR,
			BusinessInfo: &pb.CreateBusinessInfoRequest{
				GstNumber:    &gst,
				BusinessName: "Test Business",
				BusinessType: pb.BusinessType_BUSINESS_TYPE_PROPRIETORSHIP,
			},
		}

		resp, err := tc.Handler.CreateCollaborator(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Should return existing collaborator
		assert.Equal(t, "user-existing", resp.Collaborator.UserId)
	})

	t.Run("Concurrent creates with same GST", func(t *testing.T) {
		tc := setupTestContext(t)
		defer cleanupTestDB(t, tc.DB)

		gst := validTestGST
		numGoroutines := 10

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		results := make(chan *pb.CollaboratorResponse, numGoroutines)
		errors := make(chan error, numGoroutines)

		// Launch concurrent creates
		for i := 0; i < numGoroutines; i++ {
			go func(idx int) {
				defer wg.Done()

				// Convert idx+1 safely to uint64
				idxVal := idx + 1
				if idxVal < 0 {
					idxVal = 1
				}
				//nolint:gosec // Test code with bounded idx, safe conversion
				fpoID := uint64(idxVal)
				ctx := testUserContext(fmt.Sprintf("user-%d", idx), fpoID, "USER")

				req := &pb.CreateCollaboratorRequest{
					UserId:    fmt.Sprintf("user-%d", idx),
					Username:  fmt.Sprintf("user%d", idx),
					Email:     fmt.Sprintf("user%d@example.com", idx),
					FirstName: "Test",
					LastName:  fmt.Sprintf("User%d", idx),
					Type:      pb.CollaboratorType_COLLABORATOR_TYPE_VENDOR,
					BusinessInfo: &pb.CreateBusinessInfoRequest{
						GstNumber:    &gst,
						BusinessName: "Test Business",
					},
				}

				resp, err := tc.Handler.CreateCollaborator(ctx, req)
				if err != nil {
					errors <- err
				} else {
					results <- resp
				}
			}(i)
		}

		wg.Wait()
		close(results)
		close(errors)

		// Collect results
		var responses []*pb.CollaboratorResponse
		for resp := range results {
			responses = append(responses, resp)
		}

		// At least one should succeed
		assert.GreaterOrEqual(t, len(responses), 1)

		// All successful responses should have the same GST
		if len(responses) > 0 {
			for _, resp := range responses {
				assert.Equal(t, gst, *resp.Collaborator.BusinessInfo.GstNumber)
			}
		}
	})
}

// TestCreateCollaborator_SagaPattern tests saga transaction integrity (P0-2)
func TestCreateCollaborator_SagaRollback(t *testing.T) {
	t.Run("AAA failure triggers rollback", func(t *testing.T) {
		tc := setupTestContext(t)
		defer cleanupTestDB(t, tc.DB)

		// Setup AAA to fail
		tc.MockAAA.SetCreateAddressFunc(func(_ context.Context, _ *aaa.CreateAddressRequest) (*aaa.Address, error) {
			return nil, errors.New("AAA service unavailable")
		})

		ctx := testUserContext("user-1", 1, "USER")

		req := &pb.CreateCollaboratorRequest{
			UserId:    "user-1",
			Username:  "testuser",
			Email:     "test@example.com",
			FirstName: "Test",
			LastName:  "User",
			Type:      pb.CollaboratorType_COLLABORATOR_TYPE_VENDOR,
			Address: &pb.CreateAddressRequest{
				Line1:      "123 Test St",
				City:       "Mumbai",
				State:      "Maharashtra",
				Country:    "India",
				PostalCode: "400001",
				Type:       pb.AddressType_ADDRESS_TYPE_BUSINESS,
			},
		}

		resp, err := tc.Handler.CreateCollaborator(ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)

		// Verify no orphaned collaborator in DB
		var count int64
		tc.DB.Model(&collabModel.Collaborator{}).Where("user_id = ?", "user-1").Count(&count)
		assert.Equal(t, int64(0), count, "No collaborator should be created when AAA fails")
	})

	t.Run("DB failure triggers AAA compensation", func(t *testing.T) {
		tc := setupTestContext(t)
		defer cleanupTestDB(t, tc.DB)

		var createdAddressID string

		// Track AAA address creation
		tc.MockAAA.SetCreateAddressFunc(func(_ context.Context, req *aaa.CreateAddressRequest) (*aaa.Address, error) {
			addr := &aaa.Address{
				ID:         "addr-test",
				EntityID:   req.EntityID,
				EntityType: req.EntityType,
				Line1:      req.Line1,
				City:       req.City,
				State:      req.State,
				Country:    req.Country,
				PostalCode: req.PostalCode,
				Type:       req.Type,
				IsPrimary:  req.IsPrimary,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}
			createdAddressID = addr.ID
			tc.MockAAA.addresses[addr.ID] = addr
			return addr, nil
		})

		// Track AAA address deletion (compensation)
		compensationCalled := false
		tc.MockAAA.SetDeleteAddressFunc(func(_ context.Context, id string) error {
			if id == createdAddressID {
				compensationCalled = true
				delete(tc.MockAAA.addresses, id)
			}
			return nil
		})

		// Close DB to simulate failure - this will cause DB operations to fail
		sqlDB, _ := tc.DB.DB()
		_ = sqlDB.Close()

		ctx := testUserContext("user-1", 1, "USER")

		req := &pb.CreateCollaboratorRequest{
			UserId:    "user-1",
			Username:  "testuser",
			Email:     "test@example.com",
			FirstName: "Test",
			LastName:  "User",
			Type:      pb.CollaboratorType_COLLABORATOR_TYPE_VENDOR,
			Address: &pb.CreateAddressRequest{
				Line1:      "123 Test St",
				City:       "Mumbai",
				State:      "Maharashtra",
				Country:    "India",
				PostalCode: "400001",
				Type:       pb.AddressType_ADDRESS_TYPE_BUSINESS,
			},
		}

		resp, err := tc.Handler.CreateCollaborator(ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)

		// Verify compensation was called
		// Note: This depends on saga implementation details
		assert.True(t, compensationCalled, "Compensation should have been called to delete AAA address")
		// In a real scenario, we'd need to wait for async compensation
		time.Sleep(100 * time.Millisecond)

		// Verify no orphaned address in AAA
		assert.Equal(t, 0, len(tc.MockAAA.addresses), "No address should remain after rollback")
	})
}

// Helper function to create string pointers
func strPtr(s string) *string {
	return &s
}
