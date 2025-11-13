package collaborator

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	collabModel "kisanlink-ecom/entities/models/collaborator"
	"kisanlink-ecom/internal/aaa"
	"kisanlink-ecom/internal/saga"
	pb "kisanlink-ecom/proto/gen/go/collaborator/v1"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateCollaborator creates a new collaborator with AAA address integration
// Uses saga pattern for transaction integrity and distributed locking for GST deduplication
func (h *Handler) CreateCollaborator(ctx context.Context, req *pb.CreateCollaboratorRequest) (*pb.CollaboratorResponse, error) {
	// 1. Extract user context from interceptor
	userCtx := GetUserContext(ctx)

	h.logger.WithFields(logrus.Fields{
		"user_id":   req.UserId,
		"username":  req.Username,
		"email":     req.Email,
		"type":      req.Type,
		"fpo_id":    userCtx.FpoID,
		"tenant_id": userCtx.TenantID,
	}).Info("CreateCollaborator request received")

	// 2. Validate request
	if err := h.validateCreateRequest(req); err != nil {
		h.logger.WithError(err).Warn("Invalid create collaborator request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 3. If business info provided, validate and deduplicate GST
	var gstNumber string

	if req.BusinessInfo != nil && req.BusinessInfo.GstNumber != nil {
		gstNumber = *req.BusinessInfo.GstNumber

		// Validate GST format (P0-5)
		normalized, err := h.gstService.ValidateAndNormalize(gstNumber)
		if err != nil {
			h.logger.WithError(err).WithField("gst", gstNumber).Warn("GST validation failed")
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid GST number: %v", err))
		}
		gstNumber = normalized

		// Check GST deduplication with distributed lock (P0-1)
		requestID := fmt.Sprintf("%s-%d", req.UserId, time.Now().UnixNano())
		exists, err := h.gstService.CheckAndReserveGST(ctx, gstNumber, userCtx.FpoID, requestID)
		if err != nil {
			h.logger.WithError(err).WithFields(logrus.Fields{
				"gst":        gstNumber,
				"fpo_id":     userCtx.FpoID,
				"request_id": requestID,
			}).Error("Failed to check GST reservation")
			return nil, status.Error(codes.Internal, "failed to check GST availability")
		}

		if exists {
			// GST already exists - fetch existing collaborator
			existingCollab, err := h.getCollaboratorByGST(ctx, gstNumber)
			if err != nil {
				h.logger.WithError(err).WithField("gst", gstNumber).Error("Failed to fetch existing collaborator")
				return nil, status.Error(codes.Internal, "failed to fetch existing collaborator")
			}

			if existingCollab != nil {
				h.logger.WithFields(logrus.Fields{
					"gst":              gstNumber,
					"collaborator_id":  existingCollab.ID,
					"existing_user_id": existingCollab.UserID,
				}).Info("Collaborator with GST already exists")

				// Convert to proto and return
				pbCollab := h.modelToProto(existingCollab)
				return &pb.CollaboratorResponse{
					Collaborator: pbCollab,
				}, nil
			}
		}

		h.logger.WithFields(logrus.Fields{
			"gst":        gstNumber,
			"fpo_id":     userCtx.FpoID,
			"request_id": requestID,
		}).Info("GST reserved successfully")
	}

	// 4. Execute saga for AAA + DB transaction (P0-2)
	sagaResult, err := h.executeCreateSaga(ctx, req, userCtx, gstNumber)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":  req.UserId,
			"username": req.Username,
			"gst":      gstNumber,
		}).Error("Saga execution failed")
		return nil, status.Error(codes.Internal, "failed to create collaborator")
	}

	// 5. Fetch created collaborator and return
	createdCollab, err := h.getCollaboratorByID(ctx, sagaResult.CollaboratorID)
	if err != nil {
		h.logger.WithError(err).WithField("collaborator_id", sagaResult.CollaboratorID).Error("Failed to fetch created collaborator")
		return nil, status.Error(codes.Internal, "collaborator created but failed to fetch")
	}

	h.logger.WithFields(logrus.Fields{
		"collaborator_id": sagaResult.CollaboratorID,
		"address_id":      sagaResult.AddressID,
		"user_id":         req.UserId,
	}).Info("Collaborator created successfully")

	pbCollab := h.modelToProto(createdCollab)
	return &pb.CollaboratorResponse{
		Collaborator: pbCollab,
	}, nil
}

// validateCreateRequest validates the create collaborator request
func (h *Handler) validateCreateRequest(req *pb.CreateCollaboratorRequest) error {
	if req.UserId == "" {
		return fmt.Errorf("user_id is required")
	}
	if req.Username == "" {
		return fmt.Errorf("username is required")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.FirstName == "" {
		return fmt.Errorf("first_name is required")
	}
	if req.LastName == "" {
		return fmt.Errorf("last_name is required")
	}
	if req.Type == pb.CollaboratorType_COLLABORATOR_TYPE_UNSPECIFIED {
		return fmt.Errorf("type is required")
	}

	// Validate email format
	if !strings.Contains(req.Email, "@") {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// getCollaboratorByGST fetches a collaborator by GST number
func (h *Handler) getCollaboratorByGST(ctx context.Context, gstNumber string) (*collabModel.Collaborator, error) {
	var collab collabModel.Collaborator
	err := h.db.WithContext(ctx).
		Where("tax_id = ? AND deleted_at IS NULL", gstNumber).
		First(&collab).Error

	if err != nil {
		return nil, err
	}

	return &collab, nil
}

// getCollaboratorByID fetches a collaborator by ID
func (h *Handler) getCollaboratorByID(ctx context.Context, id uint64) (*collabModel.Collaborator, error) {
	var collab collabModel.Collaborator
	err := h.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&collab).Error

	if err != nil {
		return nil, err
	}

	return &collab, nil
}

// executeCreateSaga builds and executes the create collaborator saga
func (h *Handler) executeCreateSaga(ctx context.Context, req *pb.CreateCollaboratorRequest, _ *UserContext, gstNumber string) (*CreateSagaResult, error) {
	// Create saga instance (convert logrus logger to zap logger for saga)
	zapLogger := convertLogrusToZap(h.logger)
	sagaInstance := saga.NewSaga(fmt.Sprintf("create-collaborator-%s", req.UserId), zapLogger)

	// Storage for saga results
	var addressID string
	var collaboratorID uint64

	// Step 1: Create address in AAA (if address provided)
	if req.Address != nil {
		step1 := saga.NewSagaStep(
			"create_aaa_address",
			func(ctx context.Context, data map[string]interface{}) error {
				// Build AAA create address request
				aaaReq := &aaa.CreateAddressRequest{
					EntityID:   req.UserId,
					EntityType: "user",
					Line1:      req.Address.Line1,
					Line2:      getStringPtrValue(req.Address.Line2),
					City:       req.Address.City,
					State:      req.Address.State,
					Country:    req.Address.Country,
					PostalCode: req.Address.PostalCode,
					Type:       mapAddressType(req.Address.Type),
					IsPrimary:  true,
				}

				// Call AAA service with circuit breaker and retry
				addr, err := h.aaaClient.CreateAddress(ctx, aaaReq)
				if err != nil {
					h.logger.WithError(err).WithFields(logrus.Fields{
						"user_id":     req.UserId,
						"entity_type": "user",
					}).Error("Failed to create AAA address")
					return fmt.Errorf("failed to create AAA address: %w", err)
				}

				addressID = addr.ID
				data["address_id"] = addressID

				h.logger.WithFields(logrus.Fields{
					"address_id": addressID,
					"user_id":    req.UserId,
				}).Info("AAA address created successfully")

				return nil
			},
			func(ctx context.Context, _ map[string]interface{}) error {
				// Compensation: Delete address from AAA
				if addressID != "" {
					h.logger.WithField("address_id", addressID).Info("Compensating: deleting AAA address")
					if err := h.aaaClient.DeleteAddress(ctx, addressID); err != nil {
						h.logger.WithError(err).WithField("address_id", addressID).Error("Failed to compensate AAA address")
						return err
					}
					h.logger.WithField("address_id", addressID).Info("AAA address compensation successful")
				}
				return nil
			},
		).WithTimeout(10 * time.Second).WithRetry(3)

		sagaInstance.AddStep(step1)
	}

	// Step 2: Create collaborator in database
	step2 := saga.NewSagaStep(
		"create_collaborator_db",
		func(ctx context.Context, data map[string]interface{}) error {
			// Build collaborator model
			collab := collabModel.NewCollaborator(
				req.UserId,
				req.Username,
				req.Email,
				req.FirstName,
				req.LastName,
				h.mapCollaboratorType(req.Type),
			)

			// Set optional fields
			if req.Phone != nil {
				collab.Phone = *req.Phone
			}
			if req.Bio != nil {
				collab.Bio = *req.Bio
			}
			if req.ProfilePicture != nil {
				collab.Avatar = *req.ProfilePicture
			}
			if req.OrganizationId != nil {
				collab.OrganizationID = *req.OrganizationId
			}

			// Set business information if provided
			if req.BusinessInfo != nil {
				collab.BusinessName = req.BusinessInfo.BusinessName
				collab.BusinessType = req.BusinessInfo.BusinessType.String()

				if req.BusinessInfo.GstNumber != nil {
					collab.TaxID = gstNumber // Use normalized GST
				}
				// PAN number not yet supported - extend model if needed
				if req.BusinessInfo.BusinessLicense != nil {
					collab.BusinessLicense = *req.BusinessInfo.BusinessLicense
				}
				if req.BusinessInfo.BusinessPhone != nil {
					collab.BusinessPhone = *req.BusinessInfo.BusinessPhone
				}
				if req.BusinessInfo.BusinessEmail != nil {
					collab.BusinessEmail = *req.BusinessInfo.BusinessEmail
				}
				if req.BusinessInfo.BusinessWebsite != nil {
					collab.BusinessWebsite = *req.BusinessInfo.BusinessWebsite
				}
			}

			// Create collaborator in database
			if err := h.db.WithContext(ctx).Create(collab).Error; err != nil {
				h.logger.WithError(err).WithFields(logrus.Fields{
					"user_id":  req.UserId,
					"username": req.Username,
					"email":    req.Email,
				}).Error("Failed to create collaborator in database")
				return fmt.Errorf("failed to create collaborator: %w", err)
			}

			// Collaborator.ID is string in BaseModel, need to parse
			collabIDStr := fmt.Sprintf("%v", collab.ID)
			if collabIDStr != "" {
				// For now, use the database internal ID
				var dbCollab collabModel.Collaborator
				h.db.WithContext(ctx).Where("user_id = ?", collab.UserID).First(&dbCollab)
				data["collaborator_id"] = dbCollab.ID
				collaboratorID, _ = strconv.ParseUint(fmt.Sprintf("%v", dbCollab.ID), 10, 64)
			} else {
				collaboratorID = 0
			}
			data["collaborator_id"] = collaboratorID

			h.logger.WithFields(logrus.Fields{
				"collaborator_id": collaboratorID,
				"user_id":         req.UserId,
				"username":        req.Username,
			}).Info("Collaborator created in database")

			return nil
		},
		func(ctx context.Context, _ map[string]interface{}) error {
			// Compensation: Delete collaborator from DB (soft delete)
			if collaboratorID != 0 {
				h.logger.WithField("collaborator_id", collaboratorID).Info("Compensating: deleting collaborator")
				if err := h.db.WithContext(ctx).Delete(&collabModel.Collaborator{}, collaboratorID).Error; err != nil {
					h.logger.WithError(err).WithField("collaborator_id", collaboratorID).Error("Failed to compensate collaborator")
					return err
				}
				h.logger.WithField("collaborator_id", collaboratorID).Info("Collaborator compensation successful")
			}
			return nil
		},
	).WithTimeout(10 * time.Second).WithRetry(3)

	sagaInstance.AddStep(step2)

	// Execute saga with executor
	if err := h.sagaExecutor.Execute(ctx, sagaInstance); err != nil {
		return nil, fmt.Errorf("saga execution failed: %w", err)
	}

	return &CreateSagaResult{
		CollaboratorID: collaboratorID,
		AddressID:      addressID,
	}, nil
}

// Helper functions

func getStringPtrValue(ptr *string) string {
	if ptr != nil {
		return *ptr
	}
	return ""
}

func mapAddressType(pbType pb.AddressType) aaa.AddressType {
	switch pbType {
	case pb.AddressType_ADDRESS_TYPE_HOME:
		return aaa.AddressTypeHome
	case pb.AddressType_ADDRESS_TYPE_BUSINESS:
		return aaa.AddressTypeBusiness
	case pb.AddressType_ADDRESS_TYPE_BILLING:
		return aaa.AddressTypeBilling
	case pb.AddressType_ADDRESS_TYPE_SHIPPING:
		return aaa.AddressTypeShipping
	case pb.AddressType_ADDRESS_TYPE_WAREHOUSE:
		return aaa.AddressTypeWarehouse
	default:
		return aaa.AddressTypeBusiness
	}
}

func (h *Handler) mapCollaboratorType(pbType pb.CollaboratorType) collabModel.CollaboratorType {
	switch pbType {
	case pb.CollaboratorType_COLLABORATOR_TYPE_BUYER:
		return collabModel.CollaboratorTypeBuyer
	case pb.CollaboratorType_COLLABORATOR_TYPE_VENDOR:
		return collabModel.CollaboratorTypeSupplier
	case pb.CollaboratorType_COLLABORATOR_TYPE_SERVICE_PROVIDER:
		return collabModel.CollaboratorTypeAgent
	case pb.CollaboratorType_COLLABORATOR_TYPE_ADMIN:
		return collabModel.CollaboratorTypeAdmin
	case pb.CollaboratorType_COLLABORATOR_TYPE_MANAGER:
		return collabModel.CollaboratorTypeAdmin
	default:
		return collabModel.CollaboratorTypeSupplier
	}
}

func (h *Handler) modelToProtoType(modelType collabModel.CollaboratorType) pb.CollaboratorType {
	switch modelType {
	case collabModel.CollaboratorTypeBuyer:
		return pb.CollaboratorType_COLLABORATOR_TYPE_BUYER
	case collabModel.CollaboratorTypeSupplier:
		return pb.CollaboratorType_COLLABORATOR_TYPE_VENDOR
	case collabModel.CollaboratorTypeAgent:
		return pb.CollaboratorType_COLLABORATOR_TYPE_SERVICE_PROVIDER
	case collabModel.CollaboratorTypeAdmin:
		return pb.CollaboratorType_COLLABORATOR_TYPE_ADMIN
	case collabModel.CollaboratorTypeFarmer:
		return pb.CollaboratorType_COLLABORATOR_TYPE_VENDOR
	default:
		return pb.CollaboratorType_COLLABORATOR_TYPE_VENDOR
	}
}

func (h *Handler) modelToProtoStatus(modelStatus collabModel.CollaboratorStatus) pb.CollaboratorStatus {
	switch modelStatus {
	case collabModel.CollaboratorStatusActive:
		return pb.CollaboratorStatus_COLLABORATOR_STATUS_ACTIVE
	case collabModel.CollaboratorStatusInactive:
		return pb.CollaboratorStatus_COLLABORATOR_STATUS_INACTIVE
	case collabModel.CollaboratorStatusSuspended:
		return pb.CollaboratorStatus_COLLABORATOR_STATUS_SUSPENDED
	case collabModel.CollaboratorStatusPending:
		return pb.CollaboratorStatus_COLLABORATOR_STATUS_PENDING_VERIFICATION
	default:
		return pb.CollaboratorStatus_COLLABORATOR_STATUS_PENDING_VERIFICATION
	}
}

// modelToProto converts domain collaborator model to proto message
func (h *Handler) modelToProto(collab *collabModel.Collaborator) *pb.Collaborator {
	if collab == nil {
		return nil
	}

	pbCollab := &pb.Collaborator{
		Id:        fmt.Sprintf("%v", collab.ID),
		UserId:    collab.UserID,
		Username:  collab.Username,
		Email:     collab.Email,
		FirstName: collab.FirstName,
		LastName:  collab.LastName,
		Type:      h.modelToProtoType(collab.CollaboratorType),
		Status:    h.modelToProtoStatus(collab.Status),
	}

	// Optional fields
	if collab.Phone != "" {
		pbCollab.Phone = &collab.Phone
	}
	if collab.Avatar != "" {
		pbCollab.ProfilePicture = &collab.Avatar
	}
	if collab.Bio != "" {
		pbCollab.Bio = &collab.Bio
	}
	if collab.OrganizationID != "" {
		pbCollab.OrganizationId = &collab.OrganizationID
	}

	// Business information
	if collab.BusinessName != "" || collab.TaxID != "" {
		pbCollab.BusinessInfo = &pb.BusinessInfo{
			BusinessName: collab.BusinessName,
			BusinessType: pb.BusinessType_BUSINESS_TYPE_UNSPECIFIED, // Map from string if needed
		}

		if collab.TaxID != "" {
			pbCollab.BusinessInfo.GstNumber = &collab.TaxID
		}
		if collab.BusinessLicense != "" {
			pbCollab.BusinessInfo.BusinessLicense = &collab.BusinessLicense
		}
		if collab.BusinessPhone != "" {
			pbCollab.BusinessInfo.BusinessPhone = &collab.BusinessPhone
		}
		if collab.BusinessEmail != "" {
			pbCollab.BusinessInfo.BusinessEmail = &collab.BusinessEmail
		}
		if collab.BusinessWebsite != "" {
			pbCollab.BusinessInfo.BusinessWebsite = &collab.BusinessWebsite
		}
	}

	// Timestamps
	if !collab.CreatedAt.IsZero() {
		pbCollab.CreatedAt = timestampProto(collab.CreatedAt)
	}
	if !collab.UpdatedAt.IsZero() {
		pbCollab.UpdatedAt = timestampProto(collab.UpdatedAt)
	}

	return pbCollab
}

func timestampProto(t time.Time) *timestamppb.Timestamp {
	nanos := t.Nanosecond()
	if nanos < 0 || nanos > 999999999 {
		nanos = 0
	}
	return &timestamppb.Timestamp{
		Seconds: t.Unix(),
		Nanos:   int32(nanos), // #nosec G115 - nanoseconds are bounded 0-999999999
	}
}

// convertLogrusToZap converts a logrus logger to a zap logger
// This is a temporary adapter until we standardize on one logging library
func convertLogrusToZap(_ *logrus.Logger) *zap.Logger {
	// Create a simple production zap logger
	logger, err := zap.NewProduction()
	if err != nil {
		// Fallback to no-op logger
		logger = zap.NewNop()
	}
	return logger
}
