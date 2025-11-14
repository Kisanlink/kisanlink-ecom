package actors

import (
	"time"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/actors"
)

// CreateCollaboratorRequest represents the request to create a collaborator
type CreateCollaboratorRequest struct {
	AAAEntityID           string                  `json:"aaa_entity_id" binding:"required"`
	AAAEntityType         actors.AAAEntityType    `json:"aaa_entity_type" binding:"required,oneof=USER ORGANIZATION"`
	UserID                *string                 `json:"user_id" binding:"omitempty"`
	OrganizationID        *string                 `json:"organization_id" binding:"omitempty"`
	ContextOrganizationID string                  `json:"context_organization_id" binding:"required"`
	Role                  actors.CollaboratorRole `json:"role" binding:"required,oneof=OWNER ADMIN MANAGER EMPLOYEE CONTRACTOR VIEWER"`
	DisplayName           string                  `json:"display_name" binding:"omitempty,max=255"`
	Email                 string                  `json:"email" binding:"omitempty,email"`
	Phone                 string                  `json:"phone" binding:"omitempty,max=20"`
	Permissions           []string                `json:"permissions" binding:"omitempty"`
	AccessLevel           *int                    `json:"access_level" binding:"omitempty,min=1,max=10"`
	CanInviteOthers       *bool                   `json:"can_invite_others"`
	EmployeeID            string                  `json:"employee_id" binding:"omitempty,max=100"`
	Department            string                  `json:"department" binding:"omitempty,max=100"`
	JobTitle              string                  `json:"job_title" binding:"omitempty,max=100"`
	StartDate             *time.Time              `json:"start_date"`
	EndDate               *time.Time              `json:"end_date"`
	ContractType          string                  `json:"contract_type" binding:"omitempty,max=50"`
	Metadata              map[string]interface{}  `json:"metadata" binding:"omitempty"`
}

// UpdateCollaboratorRequest represents the request to update collaborator information
type UpdateCollaboratorRequest struct {
	Role            *actors.CollaboratorRole   `json:"role" binding:"omitempty,oneof=OWNER ADMIN MANAGER EMPLOYEE CONTRACTOR VIEWER"`
	Status          *actors.CollaboratorStatus `json:"status" binding:"omitempty,oneof=ACTIVE INACTIVE SUSPENDED PENDING"`
	DisplayName     *string                    `json:"display_name" binding:"omitempty,max=255"`
	Email           *string                    `json:"email" binding:"omitempty,email"`
	Phone           *string                    `json:"phone" binding:"omitempty,max=20"`
	Permissions     []string                   `json:"permissions" binding:"omitempty"`
	AccessLevel     *int                       `json:"access_level" binding:"omitempty,min=1,max=10"`
	CanInviteOthers *bool                      `json:"can_invite_others"`
	EmployeeID      *string                    `json:"employee_id" binding:"omitempty,max=100"`
	Department      *string                    `json:"department" binding:"omitempty,max=100"`
	JobTitle        *string                    `json:"job_title" binding:"omitempty,max=100"`
	StartDate       *time.Time                 `json:"start_date"`
	EndDate         *time.Time                 `json:"end_date"`
	ContractType    *string                    `json:"contract_type" binding:"omitempty,max=50"`
	Metadata        map[string]interface{}     `json:"metadata" binding:"omitempty"`
}

// InviteCollaboratorRequest represents the request to invite a collaborator
type InviteCollaboratorRequest struct {
	AAAEntityID           string                  `json:"aaa_entity_id" binding:"required"`
	AAAEntityType         actors.AAAEntityType    `json:"aaa_entity_type" binding:"required,oneof=USER ORGANIZATION"`
	ContextOrganizationID string                  `json:"context_organization_id" binding:"required"`
	Role                  actors.CollaboratorRole `json:"role" binding:"required,oneof=OWNER ADMIN MANAGER EMPLOYEE CONTRACTOR VIEWER"`
	Email                 string                  `json:"email" binding:"required,email"`
	DisplayName           string                  `json:"display_name" binding:"omitempty,max=255"`
	AccessLevel           *int                    `json:"access_level" binding:"omitempty,min=1,max=10"`
	CanInviteOthers       *bool                   `json:"can_invite_others"`
	Department            string                  `json:"department" binding:"omitempty,max=100"`
	JobTitle              string                  `json:"job_title" binding:"omitempty,max=100"`
	Message               string                  `json:"message" binding:"omitempty,max=500"`
}

// ListCollaboratorsRequest represents the request to list collaborators
type ListCollaboratorsRequest struct {
	Page                  int                        `form:"page" binding:"omitempty,min=1"`
	PageSize              int                        `form:"page_size" binding:"omitempty,min=1,max=100"`
	ContextOrganizationID *string                    `form:"context_organization_id"`
	Role                  *actors.CollaboratorRole   `form:"role" binding:"omitempty,oneof=OWNER ADMIN MANAGER EMPLOYEE CONTRACTOR VIEWER"`
	Status                *actors.CollaboratorStatus `form:"status" binding:"omitempty,oneof=ACTIVE INACTIVE SUSPENDED PENDING"`
	AAAEntityType         *actors.AAAEntityType      `form:"aaa_entity_type" binding:"omitempty,oneof=USER ORGANIZATION"`
	Department            *string                    `form:"department"`
	ContractType          *string                    `form:"contract_type"`
	Search                *string                    `form:"search"`
	SortBy                *string                    `form:"sort_by" binding:"omitempty,oneof=created_at updated_at display_name role status"`
	SortOrder             *string                    `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

// CollaboratorFilter represents filters for collaborator queries (internal use)
type CollaboratorFilter struct {
	ContextOrganizationID *string
	Role                  *actors.CollaboratorRole
	Status                *actors.CollaboratorStatus
	AAAEntityType         *actors.AAAEntityType
	Department            *string
	ContractType          *string
	Search                *string
	Page                  int
	PageSize              int
	SortBy                string
	SortOrder             string
}
