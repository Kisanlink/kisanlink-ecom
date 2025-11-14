// Package actors provides service layer for organization collaborator management
package actors

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/actors"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// OrganizationCollaboratorRepositoryInterface defines repository operations
type OrganizationCollaboratorRepositoryInterface interface {
	Create(ctx context.Context, collaborator *actors.Collaborator) error
	GetByID(ctx context.Context, id string) (*actors.Collaborator, error)
	GetByAAAEntityID(ctx context.Context, aaaEntityID string, aaaEntityType actors.AAAEntityType) ([]*actors.Collaborator, error)
	GetByContextOrganizationID(ctx context.Context, contextOrgID string) ([]*actors.Collaborator, error)
	Update(ctx context.Context, collaborator *actors.Collaborator) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter *base.Filter, limit, offset int) ([]*actors.Collaborator, int, error)
}

// OrganizationCollaboratorServiceInterface defines service operations
type OrganizationCollaboratorServiceInterface interface {
	CreateCollaborator(ctx context.Context, aaaEntityID string, aaaEntityType actors.AAAEntityType, contextOrgID string, role actors.CollaboratorRole, createdBy string) (*actors.Collaborator, error)
	GetCollaborator(ctx context.Context, id string) (*actors.Collaborator, error)
	GetEntityCollaborators(ctx context.Context, aaaEntityID string, aaaEntityType actors.AAAEntityType) ([]*actors.Collaborator, error)
	GetOrganizationCollaborators(ctx context.Context, contextOrgID string) ([]*actors.Collaborator, error)
	UpdateCollaborator(ctx context.Context, collaborator *actors.Collaborator) (*actors.Collaborator, error)
	UpdateRole(ctx context.Context, id string, role actors.CollaboratorRole) (*actors.Collaborator, error)
	UpdateStatus(ctx context.Context, id string, status actors.CollaboratorStatus) (*actors.Collaborator, error)
	UpdatePermissions(ctx context.Context, id string, permissions []string) (*actors.Collaborator, error)
	UpdateAccessLevel(ctx context.Context, id string, accessLevel int) (*actors.Collaborator, error)
	DeleteCollaborator(ctx context.Context, id string) error
	ListCollaborators(ctx context.Context, limit, offset int) ([]*actors.Collaborator, int, error)
	ActivateCollaborator(ctx context.Context, id string) (*actors.Collaborator, error)
	DeactivateCollaborator(ctx context.Context, id string) (*actors.Collaborator, error)
	SuspendCollaborator(ctx context.Context, id string) (*actors.Collaborator, error)
	CheckHigherRole(ctx context.Context, collaboratorID, otherCollaboratorID string) (bool, error)
	InviteCollaborator(ctx context.Context, aaaEntityID string, aaaEntityType actors.AAAEntityType, contextOrgID string, role actors.CollaboratorRole, invitedByUserID, createdBy string) (*actors.Collaborator, error)
	AcceptInvitation(ctx context.Context, id string) (*actors.Collaborator, error)
}

// OrganizationCollaboratorService provides business logic for organization collaborator management
type OrganizationCollaboratorService struct {
	repo OrganizationCollaboratorRepositoryInterface
}

// NewOrganizationCollaboratorService creates a new organization collaborator service
func NewOrganizationCollaboratorService(repo OrganizationCollaboratorRepositoryInterface) *OrganizationCollaboratorService {
	return &OrganizationCollaboratorService{repo: repo}
}

// CreateCollaborator creates a new organization collaborator
func (s *OrganizationCollaboratorService) CreateCollaborator(ctx context.Context, aaaEntityID string, aaaEntityType actors.AAAEntityType, contextOrgID string, role actors.CollaboratorRole, createdBy string) (*actors.Collaborator, error) {
	// Validate input
	if aaaEntityID == "" {
		return nil, fmt.Errorf("AAA entity ID is required")
	}
	if contextOrgID == "" {
		return nil, fmt.Errorf("context organization ID is required")
	}
	if createdBy == "" {
		return nil, fmt.Errorf("created by is required")
	}

	// Validate AAA entity type
	if aaaEntityType != actors.AAAEntityTypeUser && aaaEntityType != actors.AAAEntityTypeOrganization {
		return nil, fmt.Errorf("invalid AAA entity type: %s", aaaEntityType)
	}

	// Validate role
	if err := s.validateRole(role); err != nil {
		return nil, fmt.Errorf("role validation failed: %w", err)
	}

	// Check if collaborator already exists
	existing, err := s.repo.GetByAAAEntityID(ctx, aaaEntityID, aaaEntityType)
	if err == nil && existing != nil {
		for _, c := range existing {
			if c.ContextOrganizationID == contextOrgID {
				return nil, fmt.Errorf("collaborator already exists for this organization")
			}
		}
	}

	// Create new collaborator
	collaborator := actors.NewCollaborator(aaaEntityID, aaaEntityType, contextOrgID, role)
	collaborator.CreatedBy = createdBy
	collaborator.UpdatedBy = createdBy

	if err := s.repo.Create(ctx, collaborator); err != nil {
		return nil, fmt.Errorf("failed to create collaborator: %w", err)
	}

	return collaborator, nil
}

// GetCollaborator retrieves a collaborator by ID
func (s *OrganizationCollaboratorService) GetCollaborator(ctx context.Context, id string) (*actors.Collaborator, error) {
	if id == "" {
		return nil, fmt.Errorf("collaborator ID is required")
	}

	collaborator, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get collaborator: %w", err)
	}
	if collaborator == nil {
		return nil, fmt.Errorf("collaborator not found")
	}

	return collaborator, nil
}

// GetEntityCollaborators retrieves all collaborators for an AAA entity
func (s *OrganizationCollaboratorService) GetEntityCollaborators(ctx context.Context, aaaEntityID string, aaaEntityType actors.AAAEntityType) ([]*actors.Collaborator, error) {
	if aaaEntityID == "" {
		return nil, fmt.Errorf("AAA entity ID is required")
	}

	collaborators, err := s.repo.GetByAAAEntityID(ctx, aaaEntityID, aaaEntityType)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity collaborators: %w", err)
	}

	return collaborators, nil
}

// GetOrganizationCollaborators retrieves all collaborators for an organization
func (s *OrganizationCollaboratorService) GetOrganizationCollaborators(ctx context.Context, contextOrgID string) ([]*actors.Collaborator, error) {
	if contextOrgID == "" {
		return nil, fmt.Errorf("context organization ID is required")
	}

	collaborators, err := s.repo.GetByContextOrganizationID(ctx, contextOrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization collaborators: %w", err)
	}

	return collaborators, nil
}

// UpdateCollaborator updates an existing collaborator
func (s *OrganizationCollaboratorService) UpdateCollaborator(ctx context.Context, collaborator *actors.Collaborator) (*actors.Collaborator, error) {
	if collaborator == nil {
		return nil, fmt.Errorf("collaborator cannot be nil")
	}

	// Validate exists
	existing, err := s.repo.GetByID(ctx, collaborator.ID)
	if err != nil || existing == nil {
		return nil, fmt.Errorf("collaborator not found")
	}

	// Validate role
	if err := s.validateRole(collaborator.Role); err != nil {
		return nil, fmt.Errorf("role validation failed: %w", err)
	}

	// Validate access level
	if collaborator.AccessLevel < 1 || collaborator.AccessLevel > 10 {
		return nil, fmt.Errorf("access level must be between 1 and 10")
	}

	if err := s.repo.Update(ctx, collaborator); err != nil {
		return nil, fmt.Errorf("failed to update collaborator: %w", err)
	}

	return collaborator, nil
}

// UpdateRole updates the role of a collaborator
func (s *OrganizationCollaboratorService) UpdateRole(ctx context.Context, id string, role actors.CollaboratorRole) (*actors.Collaborator, error) {
	if id == "" {
		return nil, fmt.Errorf("collaborator ID is required")
	}

	if err := s.validateRole(role); err != nil {
		return nil, fmt.Errorf("role validation failed: %w", err)
	}

	collaborator, err := s.repo.GetByID(ctx, id)
	if err != nil || collaborator == nil {
		return nil, fmt.Errorf("collaborator not found")
	}

	collaborator.Role = role

	if err := s.repo.Update(ctx, collaborator); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return collaborator, nil
}

// UpdateStatus updates the status of a collaborator
func (s *OrganizationCollaboratorService) UpdateStatus(ctx context.Context, id string, status actors.CollaboratorStatus) (*actors.Collaborator, error) {
	if id == "" {
		return nil, fmt.Errorf("collaborator ID is required")
	}

	if err := s.validateStatus(status); err != nil {
		return nil, fmt.Errorf("status validation failed: %w", err)
	}

	collaborator, err := s.repo.GetByID(ctx, id)
	if err != nil || collaborator == nil {
		return nil, fmt.Errorf("collaborator not found")
	}

	collaborator.Status = status

	if err := s.repo.Update(ctx, collaborator); err != nil {
		return nil, fmt.Errorf("failed to update status: %w", err)
	}

	return collaborator, nil
}

// UpdatePermissions updates the permissions for a collaborator
func (s *OrganizationCollaboratorService) UpdatePermissions(ctx context.Context, id string, permissions []string) (*actors.Collaborator, error) {
	if id == "" {
		return nil, fmt.Errorf("collaborator ID is required")
	}
	if permissions == nil {
		return nil, fmt.Errorf("permissions cannot be nil")
	}

	collaborator, err := s.repo.GetByID(ctx, id)
	if err != nil || collaborator == nil {
		return nil, fmt.Errorf("collaborator not found")
	}

	// Marshal permissions to JSON
	permJSON, err := json.Marshal(permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal permissions: %w", err)
	}

	collaborator.Permissions = string(permJSON)

	if err := s.repo.Update(ctx, collaborator); err != nil {
		return nil, fmt.Errorf("failed to update permissions: %w", err)
	}

	return collaborator, nil
}

// UpdateAccessLevel updates the access level of a collaborator
func (s *OrganizationCollaboratorService) UpdateAccessLevel(ctx context.Context, id string, accessLevel int) (*actors.Collaborator, error) {
	if id == "" {
		return nil, fmt.Errorf("collaborator ID is required")
	}
	if accessLevel < 1 || accessLevel > 10 {
		return nil, fmt.Errorf("access level must be between 1 and 10")
	}

	collaborator, err := s.repo.GetByID(ctx, id)
	if err != nil || collaborator == nil {
		return nil, fmt.Errorf("collaborator not found")
	}

	collaborator.AccessLevel = accessLevel

	if err := s.repo.Update(ctx, collaborator); err != nil {
		return nil, fmt.Errorf("failed to update access level: %w", err)
	}

	return collaborator, nil
}

// DeleteCollaborator soft deletes a collaborator
func (s *OrganizationCollaboratorService) DeleteCollaborator(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("collaborator ID is required")
	}

	// Check exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("collaborator not found")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete collaborator: %w", err)
	}

	return nil
}

// ListCollaborators retrieves collaborators with pagination
func (s *OrganizationCollaboratorService) ListCollaborators(ctx context.Context, limit, offset int) ([]*actors.Collaborator, int, error) {
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	collaborators, total, err := s.repo.List(ctx, nil, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list collaborators: %w", err)
	}

	return collaborators, total, nil
}

// ActivateCollaborator activates a collaborator
func (s *OrganizationCollaboratorService) ActivateCollaborator(ctx context.Context, id string) (*actors.Collaborator, error) {
	if id == "" {
		return nil, fmt.Errorf("collaborator ID is required")
	}

	collaborator, err := s.repo.GetByID(ctx, id)
	if err != nil || collaborator == nil {
		return nil, fmt.Errorf("collaborator not found")
	}

	collaborator.Activate()

	if err := s.repo.Update(ctx, collaborator); err != nil {
		return nil, fmt.Errorf("failed to activate collaborator: %w", err)
	}

	return collaborator, nil
}

// DeactivateCollaborator deactivates a collaborator
func (s *OrganizationCollaboratorService) DeactivateCollaborator(ctx context.Context, id string) (*actors.Collaborator, error) {
	if id == "" {
		return nil, fmt.Errorf("collaborator ID is required")
	}

	collaborator, err := s.repo.GetByID(ctx, id)
	if err != nil || collaborator == nil {
		return nil, fmt.Errorf("collaborator not found")
	}

	collaborator.Deactivate()

	if err := s.repo.Update(ctx, collaborator); err != nil {
		return nil, fmt.Errorf("failed to deactivate collaborator: %w", err)
	}

	return collaborator, nil
}

// SuspendCollaborator suspends a collaborator
func (s *OrganizationCollaboratorService) SuspendCollaborator(ctx context.Context, id string) (*actors.Collaborator, error) {
	if id == "" {
		return nil, fmt.Errorf("collaborator ID is required")
	}

	collaborator, err := s.repo.GetByID(ctx, id)
	if err != nil || collaborator == nil {
		return nil, fmt.Errorf("collaborator not found")
	}

	collaborator.Suspend()

	if err := s.repo.Update(ctx, collaborator); err != nil {
		return nil, fmt.Errorf("failed to suspend collaborator: %w", err)
	}

	return collaborator, nil
}

// CheckHigherRole checks if one collaborator has a higher role than another
func (s *OrganizationCollaboratorService) CheckHigherRole(ctx context.Context, collaboratorID, otherCollaboratorID string) (bool, error) {
	if collaboratorID == "" || otherCollaboratorID == "" {
		return false, fmt.Errorf("both collaborator IDs are required")
	}

	collaborator, err := s.repo.GetByID(ctx, collaboratorID)
	if err != nil || collaborator == nil {
		return false, fmt.Errorf("first collaborator not found")
	}

	otherCollaborator, err := s.repo.GetByID(ctx, otherCollaboratorID)
	if err != nil || otherCollaborator == nil {
		return false, fmt.Errorf("second collaborator not found")
	}

	return collaborator.HasHigherRoleThan(otherCollaborator), nil
}

// InviteCollaborator creates a new collaborator with pending status and invitation details
func (s *OrganizationCollaboratorService) InviteCollaborator(ctx context.Context, aaaEntityID string, aaaEntityType actors.AAAEntityType, contextOrgID string, role actors.CollaboratorRole, invitedByUserID, createdBy string) (*actors.Collaborator, error) {
	// Validate input
	if invitedByUserID == "" {
		return nil, fmt.Errorf("invited by user ID is required")
	}

	// Create collaborator
	collaborator, err := s.CreateCollaborator(ctx, aaaEntityID, aaaEntityType, contextOrgID, role, createdBy)
	if err != nil {
		return nil, err
	}

	// Set invitation details
	now := time.Now()
	collaborator.InvitedByUserID = invitedByUserID
	collaborator.InvitedAt = &now
	collaborator.Status = actors.CollaboratorStatusPending

	if err := s.repo.Update(ctx, collaborator); err != nil {
		return nil, fmt.Errorf("failed to set invitation details: %w", err)
	}

	return collaborator, nil
}

// AcceptInvitation accepts a collaborator invitation
func (s *OrganizationCollaboratorService) AcceptInvitation(ctx context.Context, id string) (*actors.Collaborator, error) {
	if id == "" {
		return nil, fmt.Errorf("collaborator ID is required")
	}

	collaborator, err := s.repo.GetByID(ctx, id)
	if err != nil || collaborator == nil {
		return nil, fmt.Errorf("collaborator not found")
	}

	if collaborator.Status != actors.CollaboratorStatusPending {
		return nil, fmt.Errorf("collaborator is not in pending status")
	}

	collaborator.Activate()

	if err := s.repo.Update(ctx, collaborator); err != nil {
		return nil, fmt.Errorf("failed to accept invitation: %w", err)
	}

	return collaborator, nil
}

// validateRole validates a collaborator role
func (s *OrganizationCollaboratorService) validateRole(role actors.CollaboratorRole) error {
	validRoles := map[actors.CollaboratorRole]bool{
		actors.CollaboratorRoleOwner:      true,
		actors.CollaboratorRoleAdmin:      true,
		actors.CollaboratorRoleManager:    true,
		actors.CollaboratorRoleEmployee:   true,
		actors.CollaboratorRoleContractor: true,
		actors.CollaboratorRoleViewer:     true,
	}

	if !validRoles[role] {
		return fmt.Errorf("invalid collaborator role: %s", role)
	}

	return nil
}

// validateStatus validates a collaborator status
func (s *OrganizationCollaboratorService) validateStatus(status actors.CollaboratorStatus) error {
	validStatuses := map[actors.CollaboratorStatus]bool{
		actors.CollaboratorStatusActive:    true,
		actors.CollaboratorStatusInactive:  true,
		actors.CollaboratorStatusSuspended: true,
		actors.CollaboratorStatusPending:   true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid collaborator status: %s", status)
	}

	return nil
}
