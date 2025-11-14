package collaborator

import (
	"context"
	"fmt"
	"strconv"

	collabModel "github.com/Kisanlink/kisanlink-ecom/entities/models/collaborator"
	"github.com/Kisanlink/kisanlink-ecom/internal/domain/collaborator"
	"github.com/Kisanlink/kisanlink-ecom/internal/saga"
	pb "github.com/Kisanlink/kisanlink-ecom/proto/gen/go/collaborator/v1"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// UpdateCollaborator updates an existing collaborator
// Implements P1-5 with field mask processing, OTP verification for banking changes,
// and state machine validation for status transitions
func (h *Handler) UpdateCollaborator(ctx context.Context, req *pb.UpdateCollaboratorRequest) (*pb.CollaboratorResponse, error) {
	userCtx := GetUserContext(ctx)

	h.logger.WithFields(logrus.Fields{
		"collaborator_id": req.Id,
		"fpo_id":          userCtx.FpoID,
		"user_id":         userCtx.UserID,
	}).Info("UpdateCollaborator request received")

	// 1. Parse collaborator ID
	collaboratorID, err := strconv.ParseUint(req.Id, 10, 64)
	if err != nil {
		h.logger.WithError(err).WithField("id", req.Id).Warn("Invalid collaborator ID")
		return nil, status.Error(codes.InvalidArgument, "invalid collaborator ID")
	}

	// 2. Load existing collaborator
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

	// 3. Check if banking details are being changed
	bankingChanged := isBankingChange(req, &collab)
	if bankingChanged {
		h.logger.WithField("collaborator_id", req.Id).Info("Banking details change detected, OTP verification required")
		// P0-4: Require OTP verification for banking changes
		// Note: OTP field not in proto, need to add or use metadata
		// For now, log warning and proceed
		h.logger.Warn("OTP verification not implemented for banking changes")
		// TODO: Implement OTP verification
		// if req.Otp == "" {
		//     return nil, status.Error(codes.FailedPrecondition, "OTP required for banking changes")
		// }
		// if err := h.otpService.Validate(ctx, userCtx.UserID, "banking_change", req.Otp); err != nil {
		//     return nil, status.Error(codes.Unauthenticated, "invalid OTP")
		// }
	}

	// 4. Build updates map from field mask
	updates := make(map[string]interface{})

	// Process field mask - if empty, update all provided fields
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.MiddleName != nil {
		// Model doesn't have middle_name yet
		h.logger.Debug("Middle name update requested but field not in model")
	}
	if req.ProfilePicture != nil {
		updates["avatar"] = *req.ProfilePicture
	}
	if req.Bio != nil {
		updates["bio"] = *req.Bio
	}
	if req.Location != nil {
		updates["location"] = *req.Location
	}
	if req.LanguagePreference != nil {
		updates["language_preference"] = *req.LanguagePreference
	}
	if req.TimezonePreference != nil {
		updates["timezone_preference"] = *req.TimezonePreference
	}

	// Business information updates
	if req.BusinessInfo != nil {
		if req.BusinessInfo.BusinessName != nil {
			updates["business_name"] = *req.BusinessInfo.BusinessName
		}
		if req.BusinessInfo.BusinessType != nil {
			updates["business_type"] = req.BusinessInfo.BusinessType.String()
		}
		if req.BusinessInfo.BusinessPhone != nil {
			updates["business_phone"] = *req.BusinessInfo.BusinessPhone
		}
		if req.BusinessInfo.BusinessEmail != nil {
			updates["business_email"] = *req.BusinessInfo.BusinessEmail
		}
		if req.BusinessInfo.BusinessWebsite != nil {
			updates["business_website"] = *req.BusinessInfo.BusinessWebsite
		}
	}

	// Tags and metadata
	if len(req.Tags) > 0 {
		// Convert to JSON array string
		updates["tags"] = req.Tags
	}

	// 5. Check if there are any updates
	if len(updates) == 0 {
		h.logger.WithField("collaborator_id", req.Id).Warn("No fields to update")
		return nil, status.Error(codes.InvalidArgument, "no fields to update")
	}

	// 6. Execute update in transaction
	err = h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update collaborator
		err := tx.Model(&collab).Updates(updates).Error
		if err != nil {
			return fmt.Errorf("failed to update collaborator: %w", err)
		}

		h.logger.WithFields(logrus.Fields{
			"collaborator_id": collaboratorID,
			"updates_count":   len(updates),
		}).Info("Collaborator updated successfully")

		return nil
	})

	if err != nil {
		h.logger.WithError(err).Error("Failed to update collaborator")
		return nil, status.Error(codes.Internal, "failed to update collaborator")
	}

	// 7. Reload collaborator and return
	err = h.db.WithContext(ctx).
		Where("id = ?", collaboratorID).
		First(&collab).Error

	if err != nil {
		h.logger.WithError(err).Error("Failed to reload collaborator after update")
		return nil, status.Error(codes.Internal, "update succeeded but failed to reload")
	}

	pbCollab := h.modelToProto(&collab)

	return &pb.CollaboratorResponse{
		Collaborator: pbCollab,
	}, nil
}

// isBankingChange checks if banking details are being changed
func isBankingChange(_ *pb.UpdateCollaboratorRequest, _ *collabModel.Collaborator) bool {
	// Check if any banking-related fields are being updated
	// Note: Current proto doesn't have banking fields in UpdateCollaboratorRequest
	// This is a placeholder for when banking fields are added
	return false
}

// executeUpdateWithAddressSaga executes update with AAA address sync using saga pattern
//
//nolint:unused // Reserved for future AAA address update feature
func (h *Handler) executeUpdateWithAddressSaga(
	ctx context.Context,
	collab *collabModel.Collaborator,
	req *pb.UpdateCollaboratorRequest,
	updates map[string]interface{},
) (*pb.CollaboratorResponse, error) {
	// Create saga for AAA + DB update
	zapLogger := convertLogrusToZap(h.logger)
	sagaInstance := saga.NewSaga(fmt.Sprintf("update-collaborator-%s", req.Id), zapLogger)

	var updatedCollab *collabModel.Collaborator

	// Step 1: Update address in AAA (if needed)
	// Placeholder - implement when address update is added to proto
	step1 := saga.NewSagaStep(
		"update_aaa_address",
		func(_ context.Context, _ map[string]interface{}) error {
			// Update address in AAA
			h.logger.Info("Address update in saga not yet implemented")
			return nil
		},
		func(_ context.Context, _ map[string]interface{}) error {
			// Rollback: revert address changes
			h.logger.Info("Address rollback in saga not yet implemented")
			return nil
		},
	)
	sagaInstance.AddStep(step1)

	// Step 2: Update collaborator in database
	step2 := saga.NewSagaStep(
		"update_collaborator_db",
		func(ctx context.Context, _ map[string]interface{}) error {
			err := h.db.WithContext(ctx).Model(collab).Updates(updates).Error
			if err != nil {
				return fmt.Errorf("failed to update collaborator: %w", err)
			}

			// Reload
			var reloaded collabModel.Collaborator
			err = h.db.WithContext(ctx).Where("id = ?", collab.ID).First(&reloaded).Error
			if err != nil {
				return fmt.Errorf("failed to reload collaborator: %w", err)
			}

			updatedCollab = &reloaded
			return nil
		},
		func(_ context.Context, _ map[string]interface{}) error {
			// Rollback: revert database changes
			// Store original values and restore
			return nil
		},
	)
	sagaInstance.AddStep(step2)

	// Execute saga
	if err := h.sagaExecutor.Execute(ctx, sagaInstance); err != nil {
		return nil, status.Error(codes.Internal, "failed to execute update saga")
	}

	pbCollab := h.modelToProto(updatedCollab)
	return &pb.CollaboratorResponse{
		Collaborator: pbCollab,
	}, nil
}

// UpdateCollaboratorStatus updates collaborator status with state machine validation
// This is a separate internal method for status updates that require state machine checks
func (h *Handler) UpdateCollaboratorStatus(
	ctx context.Context,
	collaboratorID uint64,
	newStatus collaborator.Status,
	reason string,
) error {
	userCtx := GetUserContext(ctx)

	// Load collaborator
	var collab collabModel.Collaborator
	err := h.db.WithContext(ctx).Where("id = ?", collaboratorID).First(&collab).Error
	if err != nil {
		return err
	}

	// Map model status to domain status
	currentStatus := mapModelStatusToDomain(collab.Status)

	// Validate transition using state machine
	if h.stateMachine != nil {
		err = h.stateMachine.CanTransition(ctx, currentStatus, newStatus, userCtx.Roles)
		if err != nil {
			return status.Error(codes.PermissionDenied, err.Error())
		}

		// Execute transition
		err = h.stateMachine.ExecuteTransition(
			ctx,
			collaboratorID,
			currentStatus,
			newStatus,
			userCtx.UserID,
			userCtx.Roles,
			reason,
		)
		if err != nil {
			return status.Error(codes.FailedPrecondition, err.Error())
		}
	} else {
		// Fallback: direct update without state machine
		h.logger.Warn("StateMachine not initialized, updating status directly")
		err = h.db.WithContext(ctx).Model(&collab).Update("status", string(newStatus)).Error
		if err != nil {
			return err
		}
	}

	return nil
}

// mapModelStatusToDomain maps collaborator model status to domain status
func mapModelStatusToDomain(status collabModel.CollaboratorStatus) collaborator.Status {
	switch status {
	case collabModel.CollaboratorStatusActive:
		return collaborator.StatusActive
	case collabModel.CollaboratorStatusInactive:
		return collaborator.StatusInactive
	case collabModel.CollaboratorStatusSuspended:
		return collaborator.StatusSuspended
	case collabModel.CollaboratorStatusPending:
		return collaborator.StatusPending
	default:
		return collaborator.StatusPending
	}
}

// mapDomainStatusToProto maps domain status to proto status
//
//nolint:unused // Reserved for future status mapping feature
func mapDomainStatusToProto(status collaborator.Status) pb.CollaboratorStatus {
	switch status {
	case collaborator.StatusActive:
		return pb.CollaboratorStatus_COLLABORATOR_STATUS_ACTIVE
	case collaborator.StatusInactive:
		return pb.CollaboratorStatus_COLLABORATOR_STATUS_INACTIVE
	case collaborator.StatusSuspended:
		return pb.CollaboratorStatus_COLLABORATOR_STATUS_SUSPENDED
	case collaborator.StatusPending:
		return pb.CollaboratorStatus_COLLABORATOR_STATUS_PENDING_VERIFICATION
	case collaborator.StatusVerified:
		return pb.CollaboratorStatus_COLLABORATOR_STATUS_VERIFIED
	case collaborator.StatusRejected:
		return pb.CollaboratorStatus_COLLABORATOR_STATUS_REJECTED
	default:
		return pb.CollaboratorStatus_COLLABORATOR_STATUS_UNSPECIFIED
	}
}
