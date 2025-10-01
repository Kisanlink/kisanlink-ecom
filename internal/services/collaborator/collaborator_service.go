package collaborator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"kisanlink-ecom/entities/models/collaborator"
	collaboratorRequests "kisanlink-ecom/entities/requests/collaborator"
	collaboratorResponses "kisanlink-ecom/entities/responses/collaborator"
	collaboratorRepo "kisanlink-ecom/internal/repositories/collaborator"

	"github.com/sirupsen/logrus"
)

// CollaboratorServiceInterface defines the interface for collaborator service
type CollaboratorServiceInterface interface {
	// CRUD operations
	CreateCollaborator(ctx context.Context, req *collaboratorRequests.CreateCollaboratorRequest, createdBy string) (*collaboratorResponses.CollaboratorResponse, error)
	GetCollaboratorByID(ctx context.Context, id string) (*collaboratorResponses.CollaboratorResponse, error)
	GetCollaboratorByUserID(ctx context.Context, userID string) (*collaboratorResponses.CollaboratorResponse, error)
	GetCollaboratorByUsername(ctx context.Context, username string) (*collaboratorResponses.CollaboratorResponse, error)
	GetCollaboratorByEmail(ctx context.Context, email string) (*collaboratorResponses.CollaboratorResponse, error)
	UpdateCollaborator(ctx context.Context, id string, req *collaboratorRequests.UpdateCollaboratorRequest, updatedBy string) (*collaboratorResponses.CollaboratorResponse, error)
	DeleteCollaborator(ctx context.Context, id string, deletedBy string) error

	// Listing and filtering
	ListCollaborators(ctx context.Context, filter *collaboratorRequests.CollaboratorFilter, offset, limit int) ([]*collaboratorResponses.CollaboratorSummaryResponse, int, error)
	SearchCollaborators(ctx context.Context, query string, offset, limit int) ([]*collaboratorResponses.CollaboratorSummaryResponse, int, error)

	// Status management
	UpdateCollaboratorStatus(ctx context.Context, id string, req *collaboratorRequests.UpdateCollaboratorStatusRequest, updatedBy string) (*collaboratorResponses.CollaboratorResponse, error)
	ActivateCollaborator(ctx context.Context, id string, updatedBy string) (*collaboratorResponses.CollaboratorResponse, error)
	DeactivateCollaborator(ctx context.Context, id string, updatedBy string) (*collaboratorResponses.CollaboratorResponse, error)
	SuspendCollaborator(ctx context.Context, id string, updatedBy string) (*collaboratorResponses.CollaboratorResponse, error)

	// Verification
	VerifyCollaborator(ctx context.Context, id string, req *collaboratorRequests.VerifyCollaboratorRequest, verifiedBy string) (*collaboratorResponses.CollaboratorVerificationResponse, error)

	// Onboarding
	UpdateOnboardingStep(ctx context.Context, id string, req *collaboratorRequests.UpdateOnboardingStepRequest, updatedBy string) (*collaboratorResponses.OnboardingStepResponse, error)
	CompleteOnboarding(ctx context.Context, id string, updatedBy string) (*collaboratorResponses.OnboardingStepResponse, error)

	// Activity tracking
	UpdateLastActivity(ctx context.Context, id string) error
	UpdateLastLogin(ctx context.Context, id string) error

	// Bulk operations
	BulkUpdateCollaborators(ctx context.Context, req *collaboratorRequests.BulkUpdateCollaboratorsRequest, updatedBy string) (*collaboratorResponses.BulkOperationResponse, error)

	// Statistics and analytics
	GetCollaboratorStats(ctx context.Context, orgID *string) (*collaboratorResponses.CollaboratorStatsResponse, error)

	// Profile operations
	GetCollaboratorProfile(ctx context.Context, id string, requestorID string, isAdmin bool) (*collaboratorResponses.CollaboratorProfileResponse, error)
}

// CollaboratorService handles collaborator business logic
type CollaboratorService struct {
	repo   *collaboratorRepo.CollaboratorRepository
	logger *logrus.Logger
}

// NewCollaboratorService creates a new collaborator service
func NewCollaboratorService(repo *collaboratorRepo.CollaboratorRepository, logger *logrus.Logger) CollaboratorServiceInterface {
	return &CollaboratorService{
		repo:   repo,
		logger: logger,
	}
}

// CreateCollaborator creates a new collaborator
func (s *CollaboratorService) CreateCollaborator(ctx context.Context, req *collaboratorRequests.CreateCollaboratorRequest, createdBy string) (*collaboratorResponses.CollaboratorResponse, error) {
	// Check if collaborator already exists
	existing, err := s.repo.GetByUserID(ctx, req.UserID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to check existing collaborator")
		return nil, fmt.Errorf("failed to check existing collaborator: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("collaborator with user ID %s already exists", req.UserID)
	}

	// Check for duplicate username
	existingUsername, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil {
		s.logger.WithError(err).Error("Failed to check existing username")
		return nil, fmt.Errorf("failed to check existing username: %w", err)
	}
	if existingUsername != nil {
		return nil, fmt.Errorf("collaborator with username %s already exists", req.Username)
	}

	// Check for duplicate email
	existingEmail, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.logger.WithError(err).Error("Failed to check existing email")
		return nil, fmt.Errorf("failed to check existing email: %w", err)
	}
	if existingEmail != nil {
		return nil, fmt.Errorf("collaborator with email %s already exists", req.Email)
	}

	// Create new collaborator
	collab := collaborator.NewCollaborator(
		req.UserID,
		req.Username,
		req.Email,
		req.FirstName,
		req.LastName,
		req.CollaboratorType,
	)

	// Set optional fields
	if req.Phone != "" {
		collab.Phone = req.Phone
	}
	if req.OrganizationID != "" {
		collab.OrganizationID = req.OrganizationID
	}
	if req.Bio != "" {
		collab.Bio = req.Bio
	}
	if req.Location != "" {
		collab.Location = req.Location
	}
	if req.Coordinates != "" {
		collab.Coordinates = req.Coordinates
	}

	// Set business information
	collab.BusinessName = req.BusinessName
	collab.BusinessType = req.BusinessType
	collab.BusinessLicense = req.BusinessLicense
	collab.TaxID = req.TaxID
	collab.BusinessAddress = req.BusinessAddress
	collab.BusinessPhone = req.BusinessPhone
	collab.BusinessEmail = req.BusinessEmail
	collab.BusinessWebsite = req.BusinessWebsite
	collab.BusinessDescription = req.BusinessDescription

	// Set preferences
	if req.LanguagePreference != "" {
		collab.LanguagePreference = req.LanguagePreference
	}
	if req.TimezonePreference != "" {
		collab.TimezonePreference = req.TimezonePreference
	}

	// Set invitation information
	if req.InvitedBy != "" {
		collab.InvitedBy = req.InvitedBy
		now := time.Now()
		collab.InvitedAt = &now
	}

	// Set audit fields
	collab.SetCreatedBy(createdBy)

	// Save to database
	if err := s.repo.Create(ctx, collab); err != nil {
		s.logger.WithError(err).Error("Failed to create collaborator")
		return nil, fmt.Errorf("failed to create collaborator: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"collaborator_id": collab.ID,
		"user_id":         collab.UserID,
		"username":        collab.Username,
		"type":            collab.CollaboratorType,
	}).Info("Collaborator created successfully")

	return s.toCollaboratorResponse(collab), nil
}

// GetCollaboratorByID retrieves a collaborator by ID
func (s *CollaboratorService) GetCollaboratorByID(ctx context.Context, id string) (*collaboratorResponses.CollaboratorResponse, error) {
	collab, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to get collaborator by ID")
		return nil, fmt.Errorf("failed to get collaborator: %w", err)
	}
	if collab == nil {
		return nil, fmt.Errorf("collaborator not found")
	}

	return s.toCollaboratorResponse(collab), nil
}

// GetCollaboratorByUserID retrieves a collaborator by user ID
func (s *CollaboratorService) GetCollaboratorByUserID(ctx context.Context, userID string) (*collaboratorResponses.CollaboratorResponse, error) {
	collab, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to get collaborator by user ID")
		return nil, fmt.Errorf("failed to get collaborator: %w", err)
	}
	if collab == nil {
		return nil, fmt.Errorf("collaborator not found")
	}

	return s.toCollaboratorResponse(collab), nil
}

// GetCollaboratorByUsername retrieves a collaborator by username
func (s *CollaboratorService) GetCollaboratorByUsername(ctx context.Context, username string) (*collaboratorResponses.CollaboratorResponse, error) {
	collab, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		s.logger.WithError(err).WithField("username", username).Error("Failed to get collaborator by username")
		return nil, fmt.Errorf("failed to get collaborator: %w", err)
	}
	if collab == nil {
		return nil, fmt.Errorf("collaborator not found")
	}

	return s.toCollaboratorResponse(collab), nil
}

// GetCollaboratorByEmail retrieves a collaborator by email
func (s *CollaboratorService) GetCollaboratorByEmail(ctx context.Context, email string) (*collaboratorResponses.CollaboratorResponse, error) {
	collab, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		s.logger.WithError(err).WithField("email", email).Error("Failed to get collaborator by email")
		return nil, fmt.Errorf("failed to get collaborator: %w", err)
	}
	if collab == nil {
		return nil, fmt.Errorf("collaborator not found")
	}

	return s.toCollaboratorResponse(collab), nil
}

// UpdateCollaborator updates a collaborator
func (s *CollaboratorService) UpdateCollaborator(ctx context.Context, id string, req *collaboratorRequests.UpdateCollaboratorRequest, updatedBy string) (*collaboratorResponses.CollaboratorResponse, error) {
	collab, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to get collaborator for update")
		return nil, fmt.Errorf("failed to get collaborator: %w", err)
	}
	if collab == nil {
		return nil, fmt.Errorf("collaborator not found")
	}

	// Update fields if provided
	if req.FirstName != nil {
		collab.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		collab.LastName = *req.LastName
	}
	if req.Phone != nil {
		collab.Phone = *req.Phone
	}
	if req.Avatar != nil {
		collab.Avatar = *req.Avatar
	}
	if req.Bio != nil {
		collab.Bio = *req.Bio
	}
	if req.Location != nil {
		collab.Location = *req.Location
	}
	if req.Coordinates != nil {
		collab.Coordinates = *req.Coordinates
	}

	// Update business information
	if req.BusinessName != nil {
		collab.BusinessName = *req.BusinessName
	}
	if req.BusinessType != nil {
		collab.BusinessType = *req.BusinessType
	}
	if req.BusinessLicense != nil {
		collab.BusinessLicense = *req.BusinessLicense
	}
	if req.TaxID != nil {
		collab.TaxID = *req.TaxID
	}
	if req.BusinessAddress != nil {
		collab.BusinessAddress = *req.BusinessAddress
	}
	if req.BusinessPhone != nil {
		collab.BusinessPhone = *req.BusinessPhone
	}
	if req.BusinessEmail != nil {
		collab.BusinessEmail = *req.BusinessEmail
	}
	if req.BusinessWebsite != nil {
		collab.BusinessWebsite = *req.BusinessWebsite
	}
	if req.BusinessDescription != nil {
		collab.BusinessDescription = *req.BusinessDescription
	}

	// Update preferences
	if req.LanguagePreference != nil {
		collab.LanguagePreference = *req.LanguagePreference
	}
	if req.TimezonePreference != nil {
		collab.TimezonePreference = *req.TimezonePreference
	}

	// Update financial information
	if req.BankAccountNumber != nil {
		collab.BankAccountNumber = *req.BankAccountNumber
	}
	if req.BankName != nil {
		collab.BankName = *req.BankName
	}
	if req.BankBranch != nil {
		collab.BankBranch = *req.BankBranch
	}
	if req.IFSCCode != nil {
		collab.IFSCCode = *req.IFSCCode
	}
	if req.UPIId != nil {
		collab.UPIId = *req.UPIId
	}

	// Update metadata
	if req.Tags != nil {
		collab.Tags = *req.Tags
	}
	if req.Notes != nil {
		collab.Notes = *req.Notes
	}

	// Set audit fields
	collab.SetUpdatedBy(updatedBy)

	// Save to database
	if err := s.repo.Update(ctx, collab); err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to update collaborator")
		return nil, fmt.Errorf("failed to update collaborator: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"collaborator_id": collab.ID,
		"updated_by":      updatedBy,
	}).Info("Collaborator updated successfully")

	return s.toCollaboratorResponse(collab), nil
}

// DeleteCollaborator soft deletes a collaborator
func (s *CollaboratorService) DeleteCollaborator(ctx context.Context, id string, deletedBy string) error {
	collab, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to get collaborator for deletion")
		return fmt.Errorf("failed to get collaborator: %w", err)
	}
	if collab == nil {
		return fmt.Errorf("collaborator not found")
	}

	if err := s.repo.SoftDelete(ctx, id, deletedBy); err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to delete collaborator")
		return fmt.Errorf("failed to delete collaborator: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"collaborator_id": id,
		"deleted_by":      deletedBy,
	}).Info("Collaborator deleted successfully")

	return nil
}

// ListCollaborators lists collaborators with filtering and pagination
func (s *CollaboratorService) ListCollaborators(ctx context.Context, filter *collaboratorRequests.CollaboratorFilter, offset, limit int) ([]*collaboratorResponses.CollaboratorSummaryResponse, int, error) {
	collaborators, total, err := s.repo.ListCollaborators(ctx, filter, offset, limit)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list collaborators")
		return nil, 0, fmt.Errorf("failed to list collaborators: %w", err)
	}

	responses := make([]*collaboratorResponses.CollaboratorSummaryResponse, len(collaborators))
	for i, collab := range collaborators {
		responses[i] = s.toCollaboratorSummaryResponse(collab)
	}

	return responses, total, nil
}

// SearchCollaborators searches collaborators
func (s *CollaboratorService) SearchCollaborators(ctx context.Context, query string, offset, limit int) ([]*collaboratorResponses.CollaboratorSummaryResponse, int, error) {
	collaborators, total, err := s.repo.SearchCollaborators(ctx, query, limit, offset)
	if err != nil {
		s.logger.WithError(err).WithField("query", query).Error("Failed to search collaborators")
		return nil, 0, fmt.Errorf("failed to search collaborators: %w", err)
	}

	responses := make([]*collaboratorResponses.CollaboratorSummaryResponse, len(collaborators))
	for i, collab := range collaborators {
		responses[i] = s.toCollaboratorSummaryResponse(collab)
	}

	return responses, total, nil
}

// UpdateCollaboratorStatus updates the status of a collaborator
func (s *CollaboratorService) UpdateCollaboratorStatus(ctx context.Context, id string, req *collaboratorRequests.UpdateCollaboratorStatusRequest, updatedBy string) (*collaboratorResponses.CollaboratorResponse, error) {
	collab, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to get collaborator for status update")
		return nil, fmt.Errorf("failed to get collaborator: %w", err)
	}
	if collab == nil {
		return nil, fmt.Errorf("collaborator not found")
	}

	oldStatus := collab.Status
	collab.Status = req.Status
	collab.SetUpdatedBy(updatedBy)

	if err := s.repo.Update(ctx, collab); err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to update collaborator status")
		return nil, fmt.Errorf("failed to update collaborator status: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"collaborator_id": id,
		"old_status":      oldStatus,
		"new_status":      req.Status,
		"updated_by":      updatedBy,
		"reason":          req.Reason,
	}).Info("Collaborator status updated")

	return s.toCollaboratorResponse(collab), nil
}

// ActivateCollaborator activates a collaborator
func (s *CollaboratorService) ActivateCollaborator(ctx context.Context, id string, updatedBy string) (*collaboratorResponses.CollaboratorResponse, error) {
	req := &collaboratorRequests.UpdateCollaboratorStatusRequest{
		Status: collaborator.CollaboratorStatusActive,
		Reason: "Activated by admin",
	}
	return s.UpdateCollaboratorStatus(ctx, id, req, updatedBy)
}

// DeactivateCollaborator deactivates a collaborator
func (s *CollaboratorService) DeactivateCollaborator(ctx context.Context, id string, updatedBy string) (*collaboratorResponses.CollaboratorResponse, error) {
	req := &collaboratorRequests.UpdateCollaboratorStatusRequest{
		Status: collaborator.CollaboratorStatusInactive,
		Reason: "Deactivated by admin",
	}
	return s.UpdateCollaboratorStatus(ctx, id, req, updatedBy)
}

// SuspendCollaborator suspends a collaborator
func (s *CollaboratorService) SuspendCollaborator(ctx context.Context, id string, updatedBy string) (*collaboratorResponses.CollaboratorResponse, error) {
	req := &collaboratorRequests.UpdateCollaboratorStatusRequest{
		Status: collaborator.CollaboratorStatusSuspended,
		Reason: "Suspended by admin",
	}
	return s.UpdateCollaboratorStatus(ctx, id, req, updatedBy)
}

// VerifyCollaborator verifies a collaborator
func (s *CollaboratorService) VerifyCollaborator(ctx context.Context, id string, req *collaboratorRequests.VerifyCollaboratorRequest, verifiedBy string) (*collaboratorResponses.CollaboratorVerificationResponse, error) {
	collab, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to get collaborator for verification")
		return nil, fmt.Errorf("failed to get collaborator: %w", err)
	}
	if collab == nil {
		return nil, fmt.Errorf("collaborator not found")
	}

	collab.Verify(verifiedBy)
	collab.SetUpdatedBy(verifiedBy)

	if err := s.repo.Update(ctx, collab); err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to verify collaborator")
		return nil, fmt.Errorf("failed to verify collaborator: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"collaborator_id": id,
		"verified_by":     verifiedBy,
		"notes":           req.VerificationNotes,
	}).Info("Collaborator verified successfully")

	return &collaboratorResponses.CollaboratorVerificationResponse{
		ID:         collab.ID,
		IsVerified: collab.IsVerified,
		VerifiedAt: collab.VerifiedAt,
		VerifiedBy: collab.VerifiedBy,
		Message:    "Collaborator verified successfully",
	}, nil
}

// UpdateOnboardingStep updates the onboarding step for a collaborator
func (s *CollaboratorService) UpdateOnboardingStep(ctx context.Context, id string, req *collaboratorRequests.UpdateOnboardingStepRequest, updatedBy string) (*collaboratorResponses.OnboardingStepResponse, error) {
	collab, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to get collaborator for onboarding update")
		return nil, fmt.Errorf("failed to get collaborator: %w", err)
	}
	if collab == nil {
		return nil, fmt.Errorf("collaborator not found")
	}

	collab.SetOnboardingStep(req.Step)
	collab.SetUpdatedBy(updatedBy)

	if err := s.repo.Update(ctx, collab); err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to update onboarding step")
		return nil, fmt.Errorf("failed to update onboarding step: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"collaborator_id": id,
		"onboarding_step": req.Step,
		"updated_by":      updatedBy,
	}).Info("Onboarding step updated")

	return &collaboratorResponses.OnboardingStepResponse{
		ID:                  collab.ID,
		OnboardingStep:      collab.OnboardingStep,
		OnboardingCompleted: collab.OnboardingCompleted,
		Message:             "Onboarding step updated successfully",
	}, nil
}

// CompleteOnboarding marks onboarding as completed
func (s *CollaboratorService) CompleteOnboarding(ctx context.Context, id string, updatedBy string) (*collaboratorResponses.OnboardingStepResponse, error) {
	collab, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to get collaborator for onboarding completion")
		return nil, fmt.Errorf("failed to get collaborator: %w", err)
	}
	if collab == nil {
		return nil, fmt.Errorf("collaborator not found")
	}

	collab.CompleteOnboarding()
	collab.SetUpdatedBy(updatedBy)

	if err := s.repo.Update(ctx, collab); err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to complete onboarding")
		return nil, fmt.Errorf("failed to complete onboarding: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"collaborator_id": id,
		"updated_by":      updatedBy,
	}).Info("Onboarding completed")

	return &collaboratorResponses.OnboardingStepResponse{
		ID:                  collab.ID,
		OnboardingStep:      collab.OnboardingStep,
		OnboardingCompleted: collab.OnboardingCompleted,
		Message:             "Onboarding completed successfully",
	}, nil
}

// UpdateLastActivity updates the last activity timestamp
func (s *CollaboratorService) UpdateLastActivity(ctx context.Context, id string) error {
	return s.repo.UpdateLastActivity(ctx, id)
}

// UpdateLastLogin updates the last login timestamp and increments login count
func (s *CollaboratorService) UpdateLastLogin(ctx context.Context, id string) error {
	return s.repo.UpdateLastLogin(ctx, id)
}

// BulkUpdateCollaborators performs bulk updates on collaborators
func (s *CollaboratorService) BulkUpdateCollaborators(ctx context.Context, req *collaboratorRequests.BulkUpdateCollaboratorsRequest, updatedBy string) (*collaboratorResponses.BulkOperationResponse, error) {
	if len(req.CollaboratorIDs) == 0 {
		return nil, fmt.Errorf("no collaborator IDs provided")
	}

	successful := 0
	failed := []string{}
	errors := []string{}

	// Handle status updates
	if req.Status != nil {
		successCount, failedIDs, err := s.repo.BulkUpdateStatus(ctx, req.CollaboratorIDs, *req.Status, updatedBy)
		if err != nil {
			s.logger.WithError(err).Error("Failed to bulk update status")
			errors = append(errors, fmt.Sprintf("bulk status update failed: %v", err))
		}
		successful += successCount
		failed = append(failed, failedIDs...)
	}

	// TODO: Handle other bulk operations like organization updates, tags, etc.

	s.logger.WithFields(logrus.Fields{
		"total_requested": len(req.CollaboratorIDs),
		"successful":      successful,
		"failed":          len(failed),
		"updated_by":      updatedBy,
		"reason":          req.Reason,
	}).Info("Bulk update completed")

	return &collaboratorResponses.BulkOperationResponse{
		TotalRequested: len(req.CollaboratorIDs),
		Successful:     successful,
		Failed:         len(failed),
		FailedIDs:      failed,
		Errors:         errors,
		Message:        fmt.Sprintf("Bulk operation completed: %d successful, %d failed", successful, len(failed)),
	}, nil
}

// GetCollaboratorStats retrieves statistics about collaborators
func (s *CollaboratorService) GetCollaboratorStats(ctx context.Context, orgID *string) (*collaboratorResponses.CollaboratorStatsResponse, error) {
	stats, err := s.repo.GetCollaboratorStats(ctx, orgID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get collaborator stats")
		return nil, fmt.Errorf("failed to get collaborator stats: %w", err)
	}

	return &collaboratorResponses.CollaboratorStatsResponse{
		TotalCollaborators:    stats.TotalCollaborators,
		ActiveCollaborators:   stats.ActiveCollaborators,
		PendingCollaborators:  stats.PendingCollaborators,
		VerifiedCollaborators: stats.VerifiedCollaborators,
		RecentRegistrations:   stats.RecentRegistrations,
		ActiveInLast30Days:    stats.ActiveInLast30Days,
		OnboardingCompletion:  stats.OnboardingCompletion,
		CollaboratorsByType:   make(map[collaborator.CollaboratorType]int),
		CollaboratorsByStatus: make(map[collaborator.CollaboratorStatus]int),
		AverageTrustScore:     0.0, // TODO: Calculate from database
	}, nil
}

// GetCollaboratorProfile retrieves detailed profile information
func (s *CollaboratorService) GetCollaboratorProfile(ctx context.Context, id string, requestorID string, isAdmin bool) (*collaboratorResponses.CollaboratorProfileResponse, error) {
	collab, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to get collaborator profile")
		return nil, fmt.Errorf("failed to get collaborator: %w", err)
	}
	if collab == nil {
		return nil, fmt.Errorf("collaborator not found")
	}

	response := &collaboratorResponses.CollaboratorProfileResponse{
		CollaboratorResponse: *s.toCollaboratorResponse(collab),
	}

	// Parse preferences if available
	if collab.Preferences != "" {
		var preferences map[string]interface{}
		if err := json.Unmarshal([]byte(collab.Preferences), &preferences); err == nil {
			response.Preferences = preferences
		}
	}

	// Parse notification settings if available
	if collab.NotificationSettings != "" {
		var notificationSettings map[string]interface{}
		if err := json.Unmarshal([]byte(collab.NotificationSettings), &notificationSettings); err == nil {
			response.NotificationSettings = notificationSettings
		}
	}

	// Include sensitive information only for the collaborator themselves or admins
	if requestorID == collab.UserID || isAdmin {
		response.BankAccountNumber = collab.BankAccountNumber
		response.BankName = collab.BankName
		response.BankBranch = collab.BankBranch
		response.IFSCCode = collab.IFSCCode
		response.UPIId = collab.UPIId
	}

	// Include admin notes only for admins
	if isAdmin {
		response.Notes = collab.Notes
	}

	return response, nil
}

// Helper methods for converting models to responses

func (s *CollaboratorService) toCollaboratorResponse(collab *collaborator.Collaborator) *collaboratorResponses.CollaboratorResponse {
	var tags []string
	if collab.Tags != "" {
		json.Unmarshal([]byte(collab.Tags), &tags)
	}

	return &collaboratorResponses.CollaboratorResponse{
		ID:                  collab.ID,
		UserID:              collab.UserID,
		Username:            collab.Username,
		Email:               collab.Email,
		FirstName:           collab.FirstName,
		LastName:            collab.LastName,
		FullName:            collab.GetFullName(),
		Phone:               collab.Phone,
		Avatar:              collab.Avatar,
		Bio:                 collab.Bio,
		Location:            collab.Location,
		Coordinates:         collab.Coordinates,
		CollaboratorType:    collab.CollaboratorType,
		Status:              collab.Status,
		OrganizationID:      collab.OrganizationID,
		BusinessName:        collab.BusinessName,
		BusinessType:        collab.BusinessType,
		BusinessLicense:     collab.BusinessLicense,
		TaxID:               collab.TaxID,
		BusinessAddress:     collab.BusinessAddress,
		BusinessPhone:       collab.BusinessPhone,
		BusinessEmail:       collab.BusinessEmail,
		BusinessWebsite:     collab.BusinessWebsite,
		BusinessDescription: collab.BusinessDescription,
		LastLoginAt:         collab.LastLoginAt,
		LastActivityAt:      collab.LastActivityAt,
		LoginCount:          collab.LoginCount,
		IsVerified:          collab.IsVerified,
		VerifiedAt:          collab.VerifiedAt,
		VerifiedBy:          collab.VerifiedBy,
		TrustScore:          collab.TrustScore,
		CompletedOrders:     collab.CompletedOrders,
		CancelledOrders:     collab.CancelledOrders,
		AverageRating:       collab.AverageRating,
		TotalReviews:        collab.TotalReviews,
		LanguagePreference:  collab.LanguagePreference,
		TimezonePreference:  collab.TimezonePreference,
		OnboardingCompleted: collab.OnboardingCompleted,
		OnboardingStep:      collab.OnboardingStep,
		Tags:                tags,
		InvitedBy:           collab.InvitedBy,
		InvitedAt:           collab.InvitedAt,
		CreatedAt:           collab.CreatedAt,
		UpdatedAt:           collab.UpdatedAt,
		DeletedAt:           collab.DeletedAt,
	}
}

func (s *CollaboratorService) toCollaboratorSummaryResponse(collab *collaborator.Collaborator) *collaboratorResponses.CollaboratorSummaryResponse {
	return &collaboratorResponses.CollaboratorSummaryResponse{
		ID:               collab.ID,
		UserID:           collab.UserID,
		Username:         collab.Username,
		Email:            collab.Email,
		FullName:         collab.GetFullName(),
		Avatar:           collab.Avatar,
		CollaboratorType: collab.CollaboratorType,
		Status:           collab.Status,
		OrganizationID:   collab.OrganizationID,
		Location:         collab.Location,
		BusinessName:     collab.BusinessName,
		IsVerified:       collab.IsVerified,
		TrustScore:       collab.TrustScore,
		LastActivityAt:   collab.LastActivityAt,
		CreatedAt:        collab.CreatedAt,
	}
}
