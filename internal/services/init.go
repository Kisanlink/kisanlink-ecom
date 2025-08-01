package services

import (
	"kisanlink-ecom/internal/config"
	"log"
)

// ServiceContainer holds all service instances
type ServiceContainer struct {
	GRPCClient            *GRPCClient
	UserService           *UserService
	RolePermissionService *RolePermissionService
}

// NewServiceContainer creates and initializes all services
func NewServiceContainer(cfg *config.Config) (*ServiceContainer, error) {
	// Initialize gRPC client
	grpcClient, err := NewGRPCClient(cfg.GRPC.ServerAddr)
	if err != nil {
		log.Printf("Failed to initialize gRPC client: %v", err)
		return nil, err
	}

	// Initialize services
	userService := NewUserService(grpcClient)
	rolePermissionService := NewRolePermissionService(grpcClient)

	return &ServiceContainer{
		GRPCClient:            grpcClient,
		UserService:           userService,
		RolePermissionService: rolePermissionService,
	}, nil
}

// Close closes all service connections
func (sc *ServiceContainer) Close() error {
	if sc.GRPCClient != nil {
		return sc.GRPCClient.Close()
	}
	return nil
}
