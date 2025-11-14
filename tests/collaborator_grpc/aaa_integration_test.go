//go:build integration
// +build integration

package collaborator_grpc_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	aaapb "github.com/Kisanlink/aaa-service/v2/pkg/proto"
	pb "github.com/Kisanlink/kisanlink-ecom/proto/gen/go/collaborator/v1"
)

// MockAAAService simulates AAA service behavior for testing
type MockAAAService struct {
	mu               sync.Mutex
	addresses        map[string]*aaapb.Address
	addressCounter   int
	failureMode      string
	failureCount     int
	latency          time.Duration
	circuitOpen      bool
	createdAddresses []string
}

// TestAAAAddressManagement tests address management integration with AAA service
func TestAAAAddressManagement(t *testing.T) {
	t.Run("Address Immutability", func(t *testing.T) {
		ctx := context.Background()
		server, aaaService := setupServerWithAAA(t)

		// Create collaborator with address
		req := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Test Business",
			},
			Address: &pb.Address{
				Line1:      "123 Main St",
				City:       "Mumbai",
				State:      "Maharashtra",
				Country:    "India",
				PostalCode: "400001",
				Type:       pb.AddressType_ADDRESS_TYPE_BUSINESS,
			},
		}

		resp, err := server.CreateCollaborator(ctx, req)
		require.NoError(t, err)
		require.NotEmpty(t, resp.AddressIds)

		originalAddressId := resp.AddressIds[0]

		// Update address - should create new address, not modify existing
		updateReq := &pb.UpdateCollaboratorRequest{
			Id: resp.Id,
			Address: &pb.Address{
				Line1:      "456 New St",
				City:       "Mumbai",
				State:      "Maharashtra",
				Country:    "India",
				PostalCode: "400002",
				Type:       pb.AddressType_ADDRESS_TYPE_BUSINESS,
			},
		}

		updateResp, err := server.UpdateCollaborator(ctx, updateReq)
		require.NoError(t, err)

		// Should have new address ID
		assert.NotEqual(t, originalAddressId, updateResp.AddressIds[0])

		// Original address should still exist in AAA
		originalAddr := aaaService.GetAddress(originalAddressId)
		assert.NotNil(t, originalAddr)
		assert.Equal(t, "123 Main St", originalAddr.Line1)

		// New address should exist
		newAddr := aaaService.GetAddress(updateResp.AddressIds[0])
		assert.NotNil(t, newAddr)
		assert.Equal(t, "456 New St", newAddr.Line1)
	})

	t.Run("Address ID Persistence", func(t *testing.T) {
		ctx := context.Background()
		server, _ := setupServerWithAAA(t)

		// Create collaborator with address
		req := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Test Business",
			},
			Address: &pb.Address{
				Line1:      "789 Test Ave",
				City:       "Delhi",
				State:      "Delhi",
				Country:    "India",
				PostalCode: "110001",
				Type:       pb.AddressType_ADDRESS_TYPE_HOME,
			},
		}

		resp, err := server.CreateCollaborator(ctx, req)
		require.NoError(t, err)
		require.NotEmpty(t, resp.AddressIds)

		addressId := resp.AddressIds[0]

		// Retrieve collaborator - should have same address ID
		getReq := &pb.GetCollaboratorRequest{
			Id:            resp.Id,
			ExpandAddress: true,
		}

		getResp, err := server.GetCollaborator(ctx, getReq)
		require.NoError(t, err)
		assert.Equal(t, addressId, getResp.AddressIds[0])

		// When expanded, should have full address details
		assert.NotNil(t, getResp.PrimaryAddress)
		assert.Equal(t, "789 Test Ave", getResp.PrimaryAddress.Line1)
	})

	t.Run("Primary Address Flag", func(t *testing.T) {
		ctx := context.Background()
		server, _ := setupServerWithAAA(t)

		// Create collaborator with multiple addresses
		req := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Multi-Address Business",
			},
			Addresses: []*pb.Address{
				{
					Line1:      "Primary Address",
					City:       "Mumbai",
					State:      "Maharashtra",
					Country:    "India",
					PostalCode: "400001",
					Type:       pb.AddressType_ADDRESS_TYPE_BUSINESS,
					IsPrimary:  true,
				},
				{
					Line1:      "Secondary Address",
					City:       "Delhi",
					State:      "Delhi",
					Country:    "India",
					PostalCode: "110001",
					Type:       pb.AddressType_ADDRESS_TYPE_SHIPPING,
					IsPrimary:  false,
				},
			},
		}

		resp, err := server.CreateCollaborator(ctx, req)
		require.NoError(t, err)
		assert.Len(t, resp.AddressIds, 2)

		// Get with expanded addresses
		getReq := &pb.GetCollaboratorRequest{
			Id:            resp.Id,
			ExpandAddress: true,
		}

		getResp, err := server.GetCollaborator(ctx, getReq)
		require.NoError(t, err)

		// Primary address should be set
		assert.NotNil(t, getResp.PrimaryAddress)
		assert.Equal(t, "Primary Address", getResp.PrimaryAddress.Line1)
		assert.True(t, getResp.PrimaryAddress.IsPrimary)
	})

	t.Run("Address Type Validation", func(t *testing.T) {
		ctx := context.Background()
		server, _ := setupServerWithAAA(t)

		testCases := []struct {
			name        string
			addressType pb.AddressType
			shouldPass  bool
		}{
			{
				name:        "Valid HOME type",
				addressType: pb.AddressType_ADDRESS_TYPE_HOME,
				shouldPass:  true,
			},
			{
				name:        "Valid BUSINESS type",
				addressType: pb.AddressType_ADDRESS_TYPE_BUSINESS,
				shouldPass:  true,
			},
			{
				name:        "Valid BILLING type",
				addressType: pb.AddressType_ADDRESS_TYPE_BILLING,
				shouldPass:  true,
			},
			{
				name:        "Valid SHIPPING type",
				addressType: pb.AddressType_ADDRESS_TYPE_SHIPPING,
				shouldPass:  true,
			},
			{
				name:        "Valid WAREHOUSE type",
				addressType: pb.AddressType_ADDRESS_TYPE_WAREHOUSE,
				shouldPass:  true,
			},
			{
				name:        "Unspecified type",
				addressType: pb.AddressType_ADDRESS_TYPE_UNSPECIFIED,
				shouldPass:  false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := &pb.CreateCollaboratorRequest{
					BusinessInfo: &pb.BusinessInfo{
						BusinessName: "Type Test Business",
					},
					Address: &pb.Address{
						Line1:      "Test Address",
						City:       "Mumbai",
						State:      "Maharashtra",
						Country:    "India",
						PostalCode: "400001",
						Type:       tc.addressType,
					},
				}

				_, err := server.CreateCollaborator(ctx, req)
				if tc.shouldPass {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
				}
			})
		}
	})

	t.Run("Required Address Fields", func(t *testing.T) {
		ctx := context.Background()
		server, _ := setupServerWithAAA(t)

		testCases := []struct {
			name       string
			address    *pb.Address
			shouldPass bool
			errorMsg   string
		}{
			{
				name: "All required fields present",
				address: &pb.Address{
					Line1:      "123 Main St",
					City:       "Mumbai",
					State:      "Maharashtra",
					Country:    "India",
					PostalCode: "400001",
					Type:       pb.AddressType_ADDRESS_TYPE_BUSINESS,
				},
				shouldPass: true,
			},
			{
				name: "Missing Line1",
				address: &pb.Address{
					City:       "Mumbai",
					State:      "Maharashtra",
					Country:    "India",
					PostalCode: "400001",
					Type:       pb.AddressType_ADDRESS_TYPE_BUSINESS,
				},
				shouldPass: false,
				errorMsg:   "address line1 is required",
			},
			{
				name: "Missing City",
				address: &pb.Address{
					Line1:      "123 Main St",
					State:      "Maharashtra",
					Country:    "India",
					PostalCode: "400001",
					Type:       pb.AddressType_ADDRESS_TYPE_BUSINESS,
				},
				shouldPass: false,
				errorMsg:   "city is required",
			},
			{
				name: "Missing State",
				address: &pb.Address{
					Line1:      "123 Main St",
					City:       "Mumbai",
					Country:    "India",
					PostalCode: "400001",
					Type:       pb.AddressType_ADDRESS_TYPE_BUSINESS,
				},
				shouldPass: false,
				errorMsg:   "state is required",
			},
			{
				name: "Missing PostalCode",
				address: &pb.Address{
					Line1:   "123 Main St",
					City:    "Mumbai",
					State:   "Maharashtra",
					Country: "India",
					Type:    pb.AddressType_ADDRESS_TYPE_BUSINESS,
				},
				shouldPass: false,
				errorMsg:   "postal code is required",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := &pb.CreateCollaboratorRequest{
					BusinessInfo: &pb.BusinessInfo{
						BusinessName: "Field Test Business",
					},
					Address: tc.address,
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
}

// TestAAAFailureScenarios tests various AAA service failure scenarios
func TestAAAFailureScenarios(t *testing.T) {
	t.Run("AAA Service Unavailable", func(t *testing.T) {
		ctx := context.Background()
		server, aaaService := setupServerWithAAA(t)

		// Set AAA to fail
		aaaService.SetFailureMode("unavailable")

		req := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Test Business",
			},
			Address: &pb.Address{
				Line1:      "123 Main St",
				City:       "Mumbai",
				State:      "Maharashtra",
				Country:    "India",
				PostalCode: "400001",
				Type:       pb.AddressType_ADDRESS_TYPE_BUSINESS,
			},
		}

		_, err := server.CreateCollaborator(ctx, req)
		assert.Error(t, err)

		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unavailable, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "AAA service unavailable")

		// Verify no partial data was saved
		// Collaborator should not exist
		listReq := &pb.ListCollaboratorsRequest{}
		listResp, err := server.ListCollaborators(ctx, listReq)
		require.NoError(t, err)
		assert.Equal(t, 0, len(listResp.Collaborators))
	})

	t.Run("AAA Service Timeout", func(t *testing.T) {
		ctx := context.Background()
		server, aaaService := setupServerWithAAA(t)

		// Set high latency to trigger timeout
		aaaService.SetLatency(10 * time.Second)

		// Use context with timeout
		ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()

		req := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Timeout Test",
			},
			Address: &pb.Address{
				Line1:      "123 Main St",
				City:       "Mumbai",
				State:      "Maharashtra",
				Country:    "India",
				PostalCode: "400001",
				Type:       pb.AddressType_ADDRESS_TYPE_BUSINESS,
			},
		}

		_, err := server.CreateCollaborator(ctxWithTimeout, req)
		assert.Error(t, err)

		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.DeadlineExceeded, statusErr.Code())
	})

	t.Run("Partial Failure Rollback", func(t *testing.T) {
		ctx := context.Background()
		server, aaaService := setupServerWithAAA(t)

		// Allow first address creation, fail on second
		aaaService.SetFailureMode("fail_after_n")
		aaaService.SetFailureCount(1)

		req := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Partial Failure Test",
			},
			Addresses: []*pb.Address{
				{
					Line1:      "First Address",
					City:       "Mumbai",
					State:      "Maharashtra",
					Country:    "India",
					PostalCode: "400001",
					Type:       pb.AddressType_ADDRESS_TYPE_BUSINESS,
				},
				{
					Line1:      "Second Address",
					City:       "Delhi",
					State:      "Delhi",
					Country:    "India",
					PostalCode: "110001",
					Type:       pb.AddressType_ADDRESS_TYPE_SHIPPING,
				},
			},
		}

		_, err := server.CreateCollaborator(ctx, req)
		assert.Error(t, err)

		// Verify rollback - no addresses should be persisted
		assert.Equal(t, 0, len(aaaService.GetCreatedAddresses()))

		// Verify collaborator was not created
		listReq := &pb.ListCollaboratorsRequest{}
		listResp, err := server.ListCollaborators(ctx, listReq)
		require.NoError(t, err)
		assert.Equal(t, 0, len(listResp.Collaborators))
	})

	t.Run("Network Failure Between Calls", func(t *testing.T) {
		ctx := context.Background()
		server, aaaService := setupServerWithAAA(t)

		// Create collaborator successfully
		req := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Network Test",
			},
			Address: &pb.Address{
				Line1:      "123 Main St",
				City:       "Mumbai",
				State:      "Maharashtra",
				Country:    "India",
				PostalCode: "400001",
				Type:       pb.AddressType_ADDRESS_TYPE_BUSINESS,
			},
		}

		resp, err := server.CreateCollaborator(ctx, req)
		require.NoError(t, err)

		// Now simulate network failure for address retrieval
		aaaService.SetFailureMode("unavailable")

		// Try to get collaborator with expanded address
		getReq := &pb.GetCollaboratorRequest{
			Id:            resp.Id,
			ExpandAddress: true,
		}

		getResp, err := server.GetCollaborator(ctx, getReq)
		// Should succeed but without expanded address
		require.NoError(t, err)
		assert.NotEmpty(t, getResp.AddressIds)
		assert.Nil(t, getResp.PrimaryAddress) // Address expansion failed
	})

	t.Run("AAA Service Recovery", func(t *testing.T) {
		ctx := context.Background()
		server, aaaService := setupServerWithAAA(t)

		// Start with AAA failing
		aaaService.SetFailureMode("unavailable")

		req := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Recovery Test",
			},
			Address: &pb.Address{
				Line1:      "123 Main St",
				City:       "Mumbai",
				State:      "Maharashtra",
				Country:    "India",
				PostalCode: "400001",
				Type:       pb.AddressType_ADDRESS_TYPE_BUSINESS,
			},
		}

		// First attempt should fail
		_, err := server.CreateCollaborator(ctx, req)
		assert.Error(t, err)

		// Recover AAA service
		aaaService.SetFailureMode("")

		// Retry should succeed
		resp, err := server.CreateCollaborator(ctx, req)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.AddressIds)
	})
}

// TestAAACircuitBreaker tests circuit breaker behavior for AAA service
func TestAAACircuitBreaker(t *testing.T) {
	t.Run("Circuit Opens After Threshold", func(t *testing.T) {
		ctx := context.Background()
		server, aaaService := setupServerWithCircuitBreaker(t)

		// Set AAA to fail
		aaaService.SetFailureMode("unavailable")

		// Make requests until circuit opens
		failureCount := 0
		for i := 0; i < 10; i++ {
			req := &pb.CreateCollaboratorRequest{
				BusinessInfo: &pb.BusinessInfo{
					BusinessName: fmt.Sprintf("Business %d", i),
				},
				Address: &pb.Address{
					Line1:      "123 Main St",
					City:       "Mumbai",
					State:      "Maharashtra",
					Country:    "India",
					PostalCode: "400001",
					Type:       pb.AddressType_ADDRESS_TYPE_BUSINESS,
				},
			}

			_, err := server.CreateCollaborator(ctx, req)
			if err != nil {
				failureCount++

				statusErr, _ := status.FromError(err)
				// After threshold, should get circuit breaker open error
				if failureCount > 5 {
					assert.Equal(t, codes.Unavailable, statusErr.Code())
					assert.Contains(t, statusErr.Message(), "circuit breaker open")
				}
			}
		}

		assert.Greater(t, failureCount, 5)
	})

	t.Run("Circuit Recovery", func(t *testing.T) {
		ctx := context.Background()
		server, aaaService := setupServerWithCircuitBreaker(t)

		// Open circuit
		aaaService.SetCircuitOpen(true)

		req := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Test Business",
			},
			Address: &pb.Address{
				Line1:      "123 Main St",
				City:       "Mumbai",
				State:      "Maharashtra",
				Country:    "India",
				PostalCode: "400001",
				Type:       pb.AddressType_ADDRESS_TYPE_BUSINESS,
			},
		}

		// Should fail immediately with circuit open
		_, err := server.CreateCollaborator(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circuit breaker open")

		// Close circuit and ensure service is healthy
		aaaService.SetCircuitOpen(false)
		aaaService.SetFailureMode("")

		// Wait for circuit to transition to half-open
		time.Sleep(100 * time.Millisecond)

		// Should succeed now
		resp, err := server.CreateCollaborator(ctx, req)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.AddressIds)
	})
}

// TestAAAAddressRetrieval tests address retrieval from AAA service
func TestAAAAddressRetrieval(t *testing.T) {
	t.Run("Batch Address Retrieval", func(t *testing.T) {
		ctx := context.Background()
		server, _ := setupServerWithAAA(t)

		// Create multiple collaborators with addresses
		collaborators := []string{}
		for i := 0; i < 5; i++ {
			req := &pb.CreateCollaboratorRequest{
				BusinessInfo: &pb.BusinessInfo{
					BusinessName: fmt.Sprintf("Business %d", i),
				},
				Address: &pb.Address{
					Line1:      fmt.Sprintf("%d Main St", i),
					City:       "Mumbai",
					State:      "Maharashtra",
					Country:    "India",
					PostalCode: "400001",
					Type:       pb.AddressType_ADDRESS_TYPE_BUSINESS,
				},
			}

			resp, err := server.CreateCollaborator(ctx, req)
			require.NoError(t, err)
			collaborators = append(collaborators, resp.Id)
		}

		// Batch retrieve with address expansion
		batchReq := &pb.BatchGetCollaboratorsRequest{
			Ids:           collaborators,
			ExpandAddress: true,
		}

		batchResp, err := server.BatchGetCollaborators(ctx, batchReq)
		require.NoError(t, err)
		assert.Len(t, batchResp.Collaborators, 5)

		// All should have expanded addresses
		for i, collab := range batchResp.Collaborators {
			assert.NotNil(t, collab.PrimaryAddress)
			assert.Equal(t, fmt.Sprintf("%d Main St", i), collab.PrimaryAddress.Line1)
		}
	})

	t.Run("Selective Address Expansion", func(t *testing.T) {
		ctx := context.Background()
		server, _ := setupServerWithAAA(t)

		// Create collaborator with address
		req := &pb.CreateCollaboratorRequest{
			BusinessInfo: &pb.BusinessInfo{
				BusinessName: "Test Business",
			},
			Address: &pb.Address{
				Line1:      "123 Main St",
				City:       "Mumbai",
				State:      "Maharashtra",
				Country:    "India",
				PostalCode: "400001",
				Type:       pb.AddressType_ADDRESS_TYPE_BUSINESS,
			},
		}

		resp, err := server.CreateCollaborator(ctx, req)
		require.NoError(t, err)

		// Get without expansion
		getReq := &pb.GetCollaboratorRequest{
			Id:            resp.Id,
			ExpandAddress: false,
		}

		getResp, err := server.GetCollaborator(ctx, getReq)
		require.NoError(t, err)
		assert.NotEmpty(t, getResp.AddressIds)
		assert.Nil(t, getResp.PrimaryAddress)

		// Get with expansion
		getReq.ExpandAddress = true
		getResp, err = server.GetCollaborator(ctx, getReq)
		require.NoError(t, err)
		assert.NotEmpty(t, getResp.AddressIds)
		assert.NotNil(t, getResp.PrimaryAddress)
		assert.Equal(t, "123 Main St", getResp.PrimaryAddress.Line1)
	})
}

// Helper methods for MockAAAService

func (m *MockAAAService) CreateAddress(ctx context.Context, addr *aaapb.Address) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Simulate latency
	if m.latency > 0 {
		select {
		case <-time.After(m.latency):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	// Check failure modes
	if m.failureMode == "unavailable" {
		return "", status.Error(codes.Unavailable, "AAA service unavailable")
	}

	if m.failureMode == "fail_after_n" && m.failureCount > 0 {
		m.failureCount--
		if m.failureCount == 0 {
			return "", errors.New("AAA service error after n calls")
		}
	}

	if m.circuitOpen {
		return "", errors.New("circuit breaker open")
	}

	// Create address
	m.addressCounter++
	addressId := fmt.Sprintf("addr_%d", m.addressCounter)
	m.addresses[addressId] = addr
	m.createdAddresses = append(m.createdAddresses, addressId)

	return addressId, nil
}

func (m *MockAAAService) GetAddress(addressId string) *aaapb.Address {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.addresses[addressId]
}

func (m *MockAAAService) SetFailureMode(mode string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failureMode = mode
}

func (m *MockAAAService) SetFailureCount(count int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failureCount = count
}

func (m *MockAAAService) SetLatency(latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.latency = latency
}

func (m *MockAAAService) SetCircuitOpen(open bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.circuitOpen = open
}

func (m *MockAAAService) GetCreatedAddresses() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.createdAddresses
}

func setupServerWithAAA(t *testing.T) (*CollaboratorGRPCServer, *MockAAAService) {
	aaaService := &MockAAAService{
		addresses: make(map[string]*aaapb.Address),
	}

	// Setup server with mock AAA service
	server := &CollaboratorGRPCServer{
		aaaService: aaaService,
	}

	return server, aaaService
}

func setupServerWithCircuitBreaker(t *testing.T) (*CollaboratorGRPCServer, *MockAAAService) {
	server, aaaService := setupServerWithAAA(t)
	// Configure circuit breaker with low thresholds for testing
	server.ConfigureCircuitBreaker(5, 0.6, 100*time.Millisecond)
	return server, aaaService
}
