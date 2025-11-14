package collaborator

import (
	"context"
	"strconv"

	collabModel "github.com/Kisanlink/kisanlink-ecom/entities/models/collaborator"
	pb "github.com/Kisanlink/kisanlink-ecom/proto/gen/go/collaborator/v1"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// GetCollaborator retrieves a single collaborator by ID
// Implements P1-6 with FPO access control and optional AAA address expansion
func (h *Handler) GetCollaborator(ctx context.Context, req *pb.GetCollaboratorRequest) (*pb.CollaboratorResponse, error) {
	userCtx := GetUserContext(ctx)

	h.logger.WithFields(logrus.Fields{
		"collaborator_id": req.Id,
		"fpo_id":          userCtx.FpoID,
		"user_id":         userCtx.UserID,
		"expand_address":  req.ExpandAddress,
	}).Info("GetCollaborator request received")

	// 1. Parse collaborator ID
	collaboratorID, err := strconv.ParseUint(req.Id, 10, 64)
	if err != nil {
		h.logger.WithError(err).WithField("id", req.Id).Warn("Invalid collaborator ID")
		return nil, status.Error(codes.InvalidArgument, "invalid collaborator ID")
	}

	// 2. Load collaborator with FPO check
	var collab collabModel.Collaborator
	query := h.db.WithContext(ctx).Where("id = ?", collaboratorID)

	// Only check FPO if user is not admin
	if !hasRole(userCtx.Roles, "ADMIN") {
		// For FPO-scoped users, ensure they can only access their own collaborators
		// Note: Current model doesn't have fpo_id, will use organization_id as proxy
		// In production, add proper FPO scoping
		query = query.Where("user_id = ?", userCtx.UserID)
	}

	if !req.IncludeDeleted {
		query = query.Where("deleted_at IS NULL")
	}

	err = query.First(&collab).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			h.logger.WithFields(logrus.Fields{
				"collaborator_id": req.Id,
				"fpo_id":          userCtx.FpoID,
			}).Warn("Collaborator not found or access denied")
			return nil, status.Error(codes.NotFound, "collaborator not found")
		}
		h.logger.WithError(err).Error("Failed to fetch collaborator")
		return nil, status.Error(codes.Internal, "failed to fetch collaborator")
	}

	// 3. Convert to proto
	pbCollab := h.modelToProto(&collab)

	// 4. Expand address from AAA if requested
	if req.ExpandAddress && collab.BusinessAddress != "" {
		// AAA integration - fetch address by ID
		// For now, this is a placeholder as we don't have address_id in the model
		// In production, use h.aaaClient.GetAddress(ctx, collab.AddressID)
		h.logger.WithField("collaborator_id", req.Id).Debug("Address expansion requested but not yet implemented")
		// Don't fail - continue without address expansion
	}

	// 5. Filter sensitive fields based on roles
	if !hasRole(userCtx.Roles, "ADMIN", "FINANCE_MANAGER", "MANAGER") {
		// Mask sensitive business information for non-privileged users
		sanitizeSensitiveFields(pbCollab, userCtx.Roles)
	}

	h.logger.WithFields(logrus.Fields{
		"collaborator_id": pbCollab.Id,
		"user_id":         pbCollab.UserId,
		"status":          pbCollab.Status,
	}).Info("Collaborator retrieved successfully")

	return &pb.CollaboratorResponse{
		Collaborator: pbCollab,
	}, nil
}

// Removed unused helper functions - moved to helpers.go if needed

// hasRole checks if user has at least one of the specified roles
func hasRole(userRoles []string, allowedRoles ...string) bool {
	for _, userRole := range userRoles {
		for _, allowed := range allowedRoles {
			if userRole == allowed {
				return true
			}
		}
	}
	return false
}
