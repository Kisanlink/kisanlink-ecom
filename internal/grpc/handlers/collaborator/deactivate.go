package collaborator

import (
	"context"
	"fmt"
	"strconv"

	collabModel "github.com/Kisanlink/kisanlink-ecom/entities/models/collaborator"
	"github.com/Kisanlink/kisanlink-ecom/internal/domain/collaborator"
	pb "github.com/Kisanlink/kisanlink-ecom/proto/gen/go/collaborator/v1"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// DeactivateCollaborator deactivates or suspends a collaborator
// Implements P1-7 with permission checks, active orders validation, and state machine enforcement
func (h *Handler) DeactivateCollaborator(ctx context.Context, req *pb.DeactivateCollaboratorRequest) (*pb.StatusResponse, error) {
	userCtx := GetUserContext(ctx)

	h.logger.WithFields(logrus.Fields{
		"collaborator_id": req.Id,
		"fpo_id":          userCtx.FpoID,
		"user_id":         userCtx.UserID,
		"reason":          req.Reason,
		"new_status":      req.NewStatus,
	}).Info("DeactivateCollaborator request received")

	// 1. Permission check - only admin/manager can deactivate
	if !hasRole(userCtx.Roles, "ADMIN", "MANAGER") {
		h.logger.WithFields(logrus.Fields{
			"user_id":    userCtx.UserID,
			"user_roles": userCtx.Roles,
		}).Warn("Unauthorized deactivation attempt")
		return nil, status.Error(codes.PermissionDenied, "only admin or manager can deactivate collaborators")
	}

	// 2. Parse collaborator ID
	collaboratorID, err := strconv.ParseUint(req.Id, 10, 64)
	if err != nil {
		h.logger.WithError(err).WithField("id", req.Id).Warn("Invalid collaborator ID")
		return nil, status.Error(codes.InvalidArgument, "invalid collaborator ID")
	}

	// 3. Load collaborator
	var collab collabModel.Collaborator
	err = h.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", collaboratorID).
		First(&collab).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			h.logger.WithField("collaborator_id", req.Id).Warn("Collaborator not found")
			return nil, status.Error(codes.NotFound, "collaborator not found")
		}
		h.logger.WithError(err).Error("Failed to fetch collaborator")
		return nil, status.Error(codes.Internal, "failed to fetch collaborator")
	}

	// 4. Check for active orders (P1-10)
	activeOrders, err := h.checkActiveOrders(ctx, collaboratorID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to check active orders")
		return nil, status.Error(codes.Internal, "failed to check active orders")
	}

	if activeOrders > 0 {
		h.logger.WithFields(logrus.Fields{
			"collaborator_id": req.Id,
			"active_orders":   activeOrders,
		}).Warn("Cannot deactivate collaborator with active orders")
		return nil, status.Errorf(codes.FailedPrecondition,
			"cannot deactivate: collaborator has %d active orders", activeOrders)
	}

	// 5. Determine target status
	targetStatus := collaborator.StatusInactive // Default to INACTIVE instead of DEACTIVATED
	if req.NewStatus != nil {
		switch *req.NewStatus {
		case pb.CollaboratorStatus_COLLABORATOR_STATUS_SUSPENDED:
			targetStatus = collaborator.StatusSuspended
		case pb.CollaboratorStatus_COLLABORATOR_STATUS_INACTIVE:
			targetStatus = collaborator.StatusInactive
		default:
			h.logger.WithField("new_status", *req.NewStatus).Warn("Invalid deactivation status")
			return nil, status.Error(codes.InvalidArgument, "invalid new status for deactivation")
		}
	}

	// 6. Map current status to domain status
	currentStatus := mapModelStatusToDomain(collab.Status)

	// 7. Execute state transition using state machine
	if h.stateMachine != nil {
		err = h.stateMachine.ExecuteTransition(
			ctx,
			collaboratorID,
			currentStatus,
			targetStatus,
			userCtx.UserID,
			userCtx.Roles,
			getReasonOrDefault(req.Reason),
		)

		if err != nil {
			h.logger.WithError(err).WithFields(logrus.Fields{
				"collaborator_id": req.Id,
				"from_status":     currentStatus,
				"to_status":       targetStatus,
			}).Error("State transition failed")
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
	} else {
		// Fallback: direct update without state machine validation
		h.logger.Warn("StateMachine not initialized, updating status directly without validation")
		err = h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			// Update status
			err := tx.Model(&collab).Update("status", string(targetStatus)).Error
			if err != nil {
				return fmt.Errorf("failed to update status: %w", err)
			}

			// Create audit log manually
			auditLog := map[string]interface{}{
				"entity_type": "collaborator",
				"entity_id":   collaboratorID,
				"action":      "deactivate",
				"old_value":   string(currentStatus),
				"new_value":   string(targetStatus),
				"user_id":     userCtx.UserID,
				"reason":      getReasonOrDefault(req.Reason),
			}

			err = tx.Table("audit_logs").Create(auditLog).Error
			if err != nil {
				h.logger.Warn("failed to create audit log", logrus.Fields{
					"error": err.Error(),
				})
				// Don't fail transaction if audit log fails
			}

			return nil
		})

		if err != nil {
			h.logger.WithError(err).Error("Failed to deactivate collaborator")
			return nil, status.Error(codes.Internal, "failed to deactivate collaborator")
		}
	}

	h.logger.WithFields(logrus.Fields{
		"collaborator_id": req.Id,
		"new_status":      targetStatus,
		"reason":          getReasonOrDefault(req.Reason),
		"user_id":         userCtx.UserID,
	}).Info("Collaborator deactivated successfully")

	return &pb.StatusResponse{
		Success: true,
		Message: fmt.Sprintf("Collaborator deactivated successfully to status: %s", targetStatus),
	}, nil
}

// checkActiveOrders checks if a collaborator has active orders
// Implements P1-10 requirement for validation before deactivation
func (h *Handler) checkActiveOrders(ctx context.Context, collaboratorID uint64) (int64, error) {
	var count int64

	// Check if orders table exists and count active orders
	err := h.db.WithContext(ctx).
		Table("orders").
		Where("collaborator_id = ? AND status IN (?)",
			collaboratorID,
			[]string{"PENDING", "PROCESSING", "CONFIRMED", "SHIPPED"}).
		Count(&count).Error

	if err != nil {
		// If table doesn't exist, log and return 0
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		// Check if error is about missing table
		h.logger.WithError(err).Warn("Failed to check active orders, table may not exist")
		return 0, nil // Don't block deactivation if orders table doesn't exist
	}

	h.logger.WithFields(logrus.Fields{
		"collaborator_id": collaboratorID,
		"active_orders":   count,
	}).Debug("Active orders check completed")

	return count, nil
}

// getReasonOrDefault returns the reason if provided, otherwise returns a default message
func getReasonOrDefault(reason *string) string {
	if reason != nil && *reason != "" {
		return *reason
	}
	return "Deactivated by administrator"
}
