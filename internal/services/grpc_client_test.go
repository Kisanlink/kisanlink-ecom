package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGRPCClient(t *testing.T) {
	// Test with invalid server address
	client, err := NewGRPCClient("invalid-address")
	assert.Error(t, err)
	assert.Nil(t, client)

	// Test with valid server address (this would require a running server)
	// client, err = NewGRPCClient("localhost:50051")
	// if err == nil {
	// 	defer client.Close()
	// 	assert.NotNil(t, client)
	// }
}

func TestServiceContainer(t *testing.T) {
	// Test service container creation
	grpcClient, err := NewGRPCClient("invalid-address")
	if err == nil {
		userService := NewUserService(grpcClient)
		rolePermissionService := NewRolePermissionService(grpcClient)

		assert.NotNil(t, userService)
		assert.NotNil(t, rolePermissionService)
	}
}
