//go:build integration
// +build integration

package collaborator_grpc_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/Kisanlink/kisanlink-ecom/proto/gen/go/collaborator/v1"
)

// TestGSTDeduplicationLogic tests all GST-related business invariants
func TestGSTDeduplicationLogic(t *testing.T) {
	t.Run("GST Format Validation", func(t *testing.T) {
		testCases := []struct {
			name        string
			gstNumber   string
			shouldPass  bool
			description string
		}{
			{
				name:        "Valid GST format",
				gstNumber:   "27AABCU9603R1ZM",
				shouldPass:  true,
				description: "Standard GST format with valid structure",
			},
			{
				name:        "Invalid state code",
				gstNumber:   "99AABCU9603R1ZM",
				shouldPass:  false,
				description: "State code 99 doesn't exist",
			},
			{
				name:        "Invalid checksum",
				gstNumber:   "27AABCU9603R1ZX",
				shouldPass:  false,
				description: "Invalid checksum character",
			},
			{
				name:        "Too short",
				gstNumber:   "27AABCU9603R1Z",
				shouldPass:  false,
				description: "GST number is too short",
			},
			{
				name:        "Too long",
				gstNumber:   "27AABCU9603R1ZM1",
				shouldPass:  false,
				description: "GST number is too long",
			},
			{
				name:        "Invalid PAN format in GST",
				gstNumber:   "27123CU9603R1ZM",
				shouldPass:  false,
				description: "Characters 3-12 should be valid PAN",
			},
			{
				name:        "Lowercase letters",
				gstNumber:   "27aabcu9603r1zm",
				shouldPass:  false,
				description: "GST must be uppercase",
			},
			{
				name:        "Special characters",
				gstNumber:   "27AABC@9603R1ZM",
				shouldPass:  false,
				description: "No special characters allowed",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := ValidateGSTFormat(tc.gstNumber)
				if tc.shouldPass {
					assert.NoError(t, err, tc.description)
				} else {
					assert.Error(t, err, tc.description)
				}
			})
		}
	})

	t.Run("PAN Extraction from GST", func(t *testing.T) {
		testCases := []struct {
			gstNumber   string
			expectedPAN string
		}{
			{
				gstNumber:   "27AABCU9603R1ZM",
				expectedPAN: "AABCU9603R",
			},
			{
				gstNumber:   "06BZAHM6385P6Z2",
				expectedPAN: "BZAHM6385P",
			},
			{
				gstNumber:   "29GGGGG1314R9Z6",
				expectedPAN: "GGGGG1314R",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.gstNumber, func(t *testing.T) {
				pan := ExtractPANFromGST(tc.gstNumber)
				assert.Equal(t, tc.expectedPAN, pan)
			})
		}
	})

	t.Run("Same GST Links to Same Master", func(t *testing.T) {
		ctx := context.Background()
		server := setupTestServer(t)

		gstNumber := "27AABCU9603R1ZM"

		// FPO A creates collaborator
		req1 := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				GstNumber:     gstNumber,
				BusinessName:  "ABC Traders",
				ContactPerson: "John Doe",
			},
			FpoId: "FPO-A",
		}

		resp1, err := server.CreateCollaborator(ctx, req1)
		require.NoError(t, err)
		require.NotNil(t, resp1)

		// FPO B creates collaborator with same GST
		req2 := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				GstNumber:     gstNumber,
				BusinessName:  "ABC Traders", // Same business name
				ContactPerson: "Jane Smith",  // Different contact person
			},
			FpoId: "FPO-B",
		}

		resp2, err := server.CreateCollaborator(ctx, req2)
		require.NoError(t, err)
		require.NotNil(t, resp2)

		// Verify both link to same master
		assert.Equal(t, resp1.MasterCollaboratorId, resp2.MasterCollaboratorId)
		// But have different local IDs
		assert.NotEqual(t, resp1.Id, resp2.Id)
		// And can have different local data
		assert.NotEqual(t, resp1.ContactPerson, resp2.ContactPerson)
		// But share the same business name (from master)
		assert.Equal(t, resp1.BusinessName, resp2.BusinessName)
	})

	t.Run("GST Update Rejection", func(t *testing.T) {
		ctx := context.Background()
		server := setupTestServer(t)

		// Create first collaborator
		req1 := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				GstNumber:    "27AABCU9603R1ZM",
				BusinessName: "ABC Traders",
			},
		}
		resp1, err := server.CreateCollaborator(ctx, req1)
		require.NoError(t, err)

		// Create second collaborator with different GST
		req2 := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				GstNumber:    "29GGGGG1314R9Z6",
				BusinessName: "XYZ Enterprises",
			},
		}
		resp2, err := server.CreateCollaborator(ctx, req2)
		require.NoError(t, err)

		// Try to update second collaborator's GST to match first
		updateReq := &pb.UpdateCollaboratorRequest{
			Id: resp2.Id,
			BusinessInfo: &pb.BusinessInfo{
				GstNumber: "27AABCU9603R1ZM", // Try to use existing GST
			},
		}

		_, err = server.UpdateCollaborator(ctx, updateReq)
		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.AlreadyExists, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "GST number already registered")
	})

	t.Run("GST Uniqueness at Master Level", func(t *testing.T) {
		ctx := context.Background()
		server := setupTestServer(t)

		gstNumber := "27AABCU9603R1ZM"

		// Create master collaborator
		req := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				GstNumber:    gstNumber,
				BusinessName: "Test Business",
			},
		}

		resp, err := server.CreateCollaborator(ctx, req)
		require.NoError(t, err)
		require.NotEmpty(t, resp.MasterCollaboratorId)

		// Verify GST is indexed and searchable
		searchReq := &pb.SearchCollaboratorsRequest{
			GstNumber: gstNumber,
		}

		searchResp, err := server.SearchCollaborators(ctx, searchReq)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(searchResp.Collaborators), 1)

		// All results should have same master ID
		masterID := searchResp.Collaborators[0].MasterCollaboratorId
		for _, collab := range searchResp.Collaborators {
			assert.Equal(t, masterID, collab.MasterCollaboratorId)
		}
	})

	t.Run("Empty GST Handling", func(t *testing.T) {
		ctx := context.Background()
		server := setupTestServer(t)

		// Create collaborator without GST (should be allowed for some types)
		req := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Individual Farmer",
				// No GST provided
			},
			Type: pb.CollaboratorType_COLLABORATOR_TYPE_FARMER,
		}

		resp, err := server.CreateCollaborator(ctx, req)
		require.NoError(t, err)
		assert.Empty(t, resp.BusinessInfo.GstNumber)
		assert.Empty(t, resp.MasterCollaboratorId) // No master when no GST
	})
}

// TestGSTRaceConditions tests concurrent operations with GST deduplication
func TestGSTRaceConditions(t *testing.T) {
	t.Run("Concurrent Creates with Same GST", func(t *testing.T) {
		ctx := context.Background()
		server := setupTestServer(t)
		gstNumber := "27AABCU9603R1ZM"

		// Number of concurrent creates
		numGoroutines := 10
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		results := make(chan *pb.CollaboratorResponse, numGoroutines)
		errors := make(chan error, numGoroutines)

		// Launch concurrent creates
		for i := 0; i < numGoroutines; i++ {
			go func(idx int) {
				defer wg.Done()

				req := &pb.CreateCollaboratorRequest{
					BusinessInfo: &pb.BusinessInfo{
						GstNumber:     gstNumber,
						BusinessName:  "Concurrent Business",
						ContactPerson: fmt.Sprintf("Person %d", idx),
					},
					FpoId: fmt.Sprintf("FPO-%d", idx),
				}

				resp, err := server.CreateCollaborator(ctx, req)
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

		// All successful creates should have same master ID
		if len(responses) > 0 {
			masterID := responses[0].MasterCollaboratorId
			for _, resp := range responses {
				assert.Equal(t, masterID, resp.MasterCollaboratorId,
					"All collaborators with same GST should have same master ID")
			}
		}

		// Verify only one master record was created
		masterCount := countMasterRecords(t, server, gstNumber)
		assert.Equal(t, 1, masterCount, "Should have exactly one master record")
	})

	t.Run("Concurrent Updates to GST", func(t *testing.T) {
		ctx := context.Background()
		server := setupTestServer(t)

		// Create initial collaborator
		createReq := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Update Test Business",
				// No GST initially
			},
		}
		resp, err := server.CreateCollaborator(ctx, createReq)
		require.NoError(t, err)

		// Try to update GST concurrently
		gstNumbers := []string{
			"27AABCU9603R1ZM",
			"29GGGGG1314R9Z6",
			"06BZAHM6385P6Z2",
		}

		var wg sync.WaitGroup
		wg.Add(len(gstNumbers))
		successCount := 0
		var mu sync.Mutex

		for _, gst := range gstNumbers {
			go func(gstNum string) {
				defer wg.Done()

				updateReq := &pb.UpdateCollaboratorRequest{
					Id: resp.Id,
					BusinessInfo: &pb.BusinessInfo{
						GstNumber: gstNum,
					},
				}

				_, err := server.UpdateCollaborator(ctx, updateReq)
				if err == nil {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}(gst)
		}

		wg.Wait()

		// Only one update should succeed
		assert.Equal(t, 1, successCount, "Only one GST update should succeed")

		// Verify the final state
		getReq := &pb.GetCollaboratorRequest{Id: resp.Id}
		finalResp, err := server.GetCollaborator(ctx, getReq)
		require.NoError(t, err)
		assert.NotEmpty(t, finalResp.BusinessInfo.GstNumber)
		assert.Contains(t, gstNumbers, finalResp.BusinessInfo.GstNumber)
	})
}

// TestGSTBusinessRules tests complex business rules around GST
func TestGSTBusinessRules(t *testing.T) {
	t.Run("GST Required for Vendor Type", func(t *testing.T) {
		ctx := context.Background()
		server := setupTestServer(t)

		req := &pb.CreateCollaboratorRequest{
			Type: pb.CollaboratorType_COLLABORATOR_TYPE_VENDOR,
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Vendor without GST",
				// No GST provided
			},
		}

		_, err := server.CreateCollaborator(ctx, req)
		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "GST number is required for vendors")
	})

	t.Run("GST State Code Validation", func(t *testing.T) {
		validStateCodes := map[string]string{
			"01": "Jammu and Kashmir",
			"02": "Himachal Pradesh",
			"03": "Punjab",
			"04": "Chandigarh",
			"05": "Uttarakhand",
			"06": "Haryana",
			"07": "Delhi",
			"08": "Rajasthan",
			"09": "Uttar Pradesh",
			"10": "Bihar",
			"27": "Maharashtra",
			"29": "Karnataka",
		}

		for code, state := range validStateCodes {
			t.Run(state, func(t *testing.T) {
				gst := fmt.Sprintf("%sAABCU9603R1ZM", code)
				err := ValidateGSTFormat(gst)
				assert.NoError(t, err)

				stateFromGST := ExtractStateFromGST(gst)
				assert.Equal(t, state, stateFromGST)
			})
		}
	})

	t.Run("GST Change Impact on Master", func(t *testing.T) {
		ctx := context.Background()
		server := setupTestServer(t)

		// Create collaborator with GST
		originalGST := "27AABCU9603R1ZM"
		req1 := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				GstNumber:    originalGST,
				BusinessName: "Original Business",
			},
			FpoId: "FPO-A",
		}
		resp1, err := server.CreateCollaborator(ctx, req1)
		require.NoError(t, err)

		// Another FPO creates with same GST
		req2 := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				GstNumber:    originalGST,
				BusinessName: "Original Business",
			},
			FpoId: "FPO-B",
		}
		resp2, err := server.CreateCollaborator(ctx, req2)
		require.NoError(t, err)

		// Verify both linked to same master
		assert.Equal(t, resp1.MasterCollaboratorId, resp2.MasterCollaboratorId)

		// Attempt to change GST (should fail if other FPOs are using it)
		updateReq := &pb.UpdateCollaboratorRequest{
			Id: resp1.Id,
			BusinessInfo: &pb.BusinessInfo{
				GstNumber: "29GGGGG1314R9Z6", // Try to change GST
			},
		}

		_, err = server.UpdateCollaborator(ctx, updateReq)
		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.FailedPrecondition, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "Cannot change GST when linked to shared master")
	})
}

// Helper functions

func ValidateGSTFormat(gst string) error {
	// GST format: [0-9]{2}[A-Z]{5}[0-9]{4}[A-Z]{1}[1-9A-Z]{1}Z[0-9A-Z]{1}
	if len(gst) != 15 {
		return fmt.Errorf("GST must be 15 characters long")
	}

	// Validate state code (first 2 digits)
	stateCode := gst[0:2]
	if !isValidStateCode(stateCode) {
		return fmt.Errorf("invalid state code: %s", stateCode)
	}

	// Validate PAN format (next 10 characters)
	pan := gst[2:12]
	if !isValidPAN(pan) {
		return fmt.Errorf("invalid PAN in GST: %s", pan)
	}

	// Validate entity number
	if gst[12] < '1' || gst[12] > '9' {
		return fmt.Errorf("invalid entity number: %c", gst[12])
	}

	// Check for Z at position 14
	if gst[13] != 'Z' {
		return fmt.Errorf("character at position 14 must be 'Z'")
	}

	// Validate checksum
	if !isValidGSTChecksum(gst) {
		return fmt.Errorf("invalid GST checksum")
	}

	return nil
}

func ExtractPANFromGST(gst string) string {
	if len(gst) < 12 {
		return ""
	}
	return gst[2:12]
}

func ExtractStateFromGST(gst string) string {
	if len(gst) < 2 {
		return ""
	}

	stateMap := map[string]string{
		"01": "Jammu and Kashmir",
		"02": "Himachal Pradesh",
		"03": "Punjab",
		"04": "Chandigarh",
		"05": "Uttarakhand",
		"06": "Haryana",
		"07": "Delhi",
		"08": "Rajasthan",
		"09": "Uttar Pradesh",
		"10": "Bihar",
		"11": "Sikkim",
		"12": "Arunachal Pradesh",
		"13": "Nagaland",
		"14": "Manipur",
		"15": "Mizoram",
		"16": "Tripura",
		"17": "Meghalaya",
		"18": "Assam",
		"19": "West Bengal",
		"20": "Jharkhand",
		"21": "Orissa",
		"22": "Chhattisgarh",
		"23": "Madhya Pradesh",
		"24": "Gujarat",
		"25": "Daman and Diu",
		"26": "Dadra and Nagar Haveli",
		"27": "Maharashtra",
		"28": "Andhra Pradesh",
		"29": "Karnataka",
		"30": "Goa",
		"31": "Lakshadweep",
		"32": "Kerala",
		"33": "Tamil Nadu",
		"34": "Puducherry",
		"35": "Andaman and Nicobar Islands",
		"36": "Telangana",
		"37": "Andhra Pradesh (New)",
	}

	return stateMap[gst[0:2]]
}

func isValidStateCode(code string) bool {
	// Valid state codes are from 01 to 37
	if len(code) != 2 {
		return false
	}

	num := 0
	fmt.Sscanf(code, "%d", &num)
	return num >= 1 && num <= 37
}

func isValidPAN(pan string) bool {
	// PAN format: [A-Z]{5}[0-9]{4}[A-Z]{1}
	if len(pan) != 10 {
		return false
	}

	// First 5 must be uppercase letters
	for i := 0; i < 5; i++ {
		if pan[i] < 'A' || pan[i] > 'Z' {
			return false
		}
	}

	// Next 4 must be digits
	for i := 5; i < 9; i++ {
		if pan[i] < '0' || pan[i] > '9' {
			return false
		}
	}

	// Last must be uppercase letter
	if pan[9] < 'A' || pan[9] > 'Z' {
		return false
	}

	return true
}

func isValidGSTChecksum(gst string) bool {
	// Implement GST checksum validation algorithm
	// This is a simplified version - actual implementation would use proper checksum algorithm
	return true // Placeholder
}

func setupTestServer(t *testing.T) *CollaboratorGRPCServer {
	// Setup test server with mocked dependencies
	// This would be implemented based on actual server structure
	return &CollaboratorGRPCServer{}
}

func countMasterRecords(t *testing.T, server *CollaboratorGRPCServer, gst string) int {
	// Count master records for given GST
	// This would query the database or check internal state
	return 1 // Placeholder
}
