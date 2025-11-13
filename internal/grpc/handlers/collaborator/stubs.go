package collaborator

import (
	"context"

	pb "kisanlink-ecom/proto/gen/go/collaborator/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// UpdateCollaborator updates an existing collaborator (stub implementation)
func (h *Handler) UpdateCollaborator(_ context.Context, _ *pb.UpdateCollaboratorRequest) (*pb.CollaboratorResponse, error) {
	return nil, status.Error(codes.Unimplemented, "UpdateCollaborator not yet implemented")
}

// DeactivateCollaborator deactivates or suspends a collaborator (stub implementation)
func (h *Handler) DeactivateCollaborator(_ context.Context, _ *pb.DeactivateCollaboratorRequest) (*pb.StatusResponse, error) {
	return nil, status.Error(codes.Unimplemented, "DeactivateCollaborator not yet implemented")
}

// GetCollaborator retrieves a single collaborator by ID (stub implementation)
func (h *Handler) GetCollaborator(_ context.Context, _ *pb.GetCollaboratorRequest) (*pb.CollaboratorResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetCollaborator not yet implemented")
}

// ListCollaborators lists collaborators with filtering and pagination (stub implementation)
func (h *Handler) ListCollaborators(_ context.Context, _ *pb.ListCollaboratorsRequest) (*pb.ListCollaboratorsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ListCollaborators not yet implemented")
}

// VerifyCollaborator verifies business information (stub implementation)
func (h *Handler) VerifyCollaborator(_ context.Context, _ *pb.VerifyCollaboratorRequest) (*pb.VerifyCollaboratorResponse, error) {
	return nil, status.Error(codes.Unimplemented, "VerifyCollaborator not yet implemented")
}

// GetCollaboratorByGST retrieves a collaborator by GST number (stub implementation)
func (h *Handler) GetCollaboratorByGST(_ context.Context, _ *pb.GetCollaboratorByGSTRequest) (*pb.CollaboratorResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetCollaboratorByGST not yet implemented")
}

// HealthCheck endpoint
func (h *Handler) HealthCheck(_ context.Context, _ *emptypb.Empty) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Status: pb.HealthCheckResponse_STATUS_HEALTHY,
	}, nil
}
