package actors

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// CollaboratorRole represents the role of a collaborator
type CollaboratorRole string

const (
	CollaboratorRoleOwner      CollaboratorRole = "OWNER"
	CollaboratorRoleAdmin      CollaboratorRole = "ADMIN"
	CollaboratorRoleManager    CollaboratorRole = "MANAGER"
	CollaboratorRoleEmployee   CollaboratorRole = "EMPLOYEE"
	CollaboratorRoleContractor CollaboratorRole = "CONTRACTOR"
	CollaboratorRoleViewer     CollaboratorRole = "VIEWER"
)

// CollaboratorStatus represents the status of a collaborator
type CollaboratorStatus string

const (
	CollaboratorStatusActive    CollaboratorStatus = "ACTIVE"
	CollaboratorStatusInactive  CollaboratorStatus = "INACTIVE"
	CollaboratorStatusSuspended CollaboratorStatus = "SUSPENDED"
	CollaboratorStatusPending   CollaboratorStatus = "PENDING"
)

// Collaborator represents a collaborator entity that references AAA service entities
type Collaborator struct {
	base.BaseModel

	// AAA Service References
	AAAEntityID   string        `json:"aaa_entity_id" gorm:"type:varchar(255);not null;index:idx_aaa_entity;index:idx_aaa_composite,priority:1"`
	AAAEntityType AAAEntityType `json:"aaa_entity_type" gorm:"type:varchar(20);not null;index:idx_aaa_type;index:idx_aaa_composite,priority:2;check:aaa_entity_type IN ('USER', 'ORGANIZATION')"`

	// Internal References (optional, for caching/performance)
	UserID         *string `json:"user_id" gorm:"type:varchar(255);index:idx_user"`        // Reference to user ID from AAA
	OrganizationID *string `json:"organization_id" gorm:"type:varchar(255);index:idx_org"` // Reference to org ID from AAA

	// Organization Context
	ContextOrganizationID string `json:"context_organization_id" gorm:"type:varchar(255);not null;index:idx_context_org;index:idx_org_role,priority:1"` // The org this collaborator belongs to

	// Collaborator-specific fields
	Role        CollaboratorRole   `json:"role" gorm:"type:varchar(20);not null;index:idx_org_role,priority:2;check:role IN ('OWNER', 'ADMIN', 'MANAGER', 'EMPLOYEE', 'CONTRACTOR', 'VIEWER')"`
	Status      CollaboratorStatus `json:"status" gorm:"type:varchar(20);not null;default:'PENDING';index:idx_status;check:status IN ('ACTIVE', 'INACTIVE', 'SUSPENDED', 'PENDING')"`
	DisplayName string             `json:"display_name" gorm:"type:varchar(255);index:idx_collaborators_search_name"`
	Email       string             `json:"email" gorm:"type:varchar(255);index:idx_email"`
	Phone       string             `json:"phone" gorm:"type:varchar(20);index:idx_phone"`

	// Access Control
	Permissions     string `json:"permissions" gorm:"type:jsonb;index:idx_permissions"`                                   // JSON array of specific permissions
	AccessLevel     int    `json:"access_level" gorm:"not null;default:1;check:access_level >= 1 AND access_level <= 10"` // Numeric access level (1-10)
	CanInviteOthers bool   `json:"can_invite_others" gorm:"not null;default:false"`

	// Employment/Contract Information
	EmployeeID   string     `json:"employee_id" gorm:"type:varchar(100);uniqueIndex:idx_employee_id,where:employee_id IS NOT NULL AND employee_id != ''"`
	Department   string     `json:"department" gorm:"type:varchar(100);index:idx_department"`
	JobTitle     string     `json:"job_title" gorm:"type:varchar(100)"`
	StartDate    *time.Time `json:"start_date" gorm:"type:timestamp"`
	EndDate      *time.Time `json:"end_date" gorm:"type:timestamp"`
	ContractType string     `json:"contract_type" gorm:"type:varchar(50);check:contract_type IN ('full-time', 'part-time', 'contract', 'consultant', 'intern') OR contract_type IS NULL"` // full-time, part-time, contract, etc.

	// Invitation Information
	InvitedByUserID string     `json:"invited_by_user_id" gorm:"type:varchar(255);index:idx_invited_by"`
	InvitedAt       *time.Time `json:"invited_at" gorm:"type:timestamp"`
	AcceptedAt      *time.Time `json:"accepted_at" gorm:"type:timestamp"`

	// Metadata
	Metadata string `json:"metadata" gorm:"type:jsonb;index:idx_collaborators_metadata"`

	// Audit fields
	CreatedBy string `json:"created_by" gorm:"type:varchar(255);not null"`
	UpdatedBy string `json:"updated_by" gorm:"type:varchar(255);not null"`

	// Soft delete support
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index:idx_collaborators_deleted"`
}

// TableName returns the table name for GORM
func (Collaborator) TableName() string {
	return "collaborators"
}

// NewCollaborator creates a new Collaborator instance
func NewCollaborator(aaaEntityID string, aaaEntityType AAAEntityType, contextOrgID string, role CollaboratorRole) *Collaborator {
	return &Collaborator{
		BaseModel:             *base.NewBaseModel("COLLAB", "large"),
		AAAEntityID:           aaaEntityID,
		AAAEntityType:         aaaEntityType,
		ContextOrganizationID: contextOrgID,
		Role:                  role,
		Status:                CollaboratorStatusPending,
		AccessLevel:           1,
		CanInviteOthers:       false,
	}
}

// NewUserCollaborator creates a new Collaborator for a user entity
func NewUserCollaborator(aaaEntityID, userID, contextOrgID string, role CollaboratorRole) *Collaborator {
	collaborator := NewCollaborator(aaaEntityID, AAAEntityTypeUser, contextOrgID, role)
	collaborator.UserID = &userID
	return collaborator
}

// NewOrganizationCollaborator creates a new Collaborator for an organization entity
func NewOrganizationCollaborator(aaaEntityID, organizationID, contextOrgID string, role CollaboratorRole) *Collaborator {
	collaborator := NewCollaborator(aaaEntityID, AAAEntityTypeOrganization, contextOrgID, role)
	collaborator.OrganizationID = &organizationID
	return collaborator
}

// IsUserCollaborator checks if the collaborator is associated with a user
func (c *Collaborator) IsUserCollaborator() bool {
	return c.AAAEntityType == AAAEntityTypeUser
}

// IsOrganizationCollaborator checks if the collaborator is associated with an organization
func (c *Collaborator) IsOrganizationCollaborator() bool {
	return c.AAAEntityType == AAAEntityTypeOrganization
}

// GetEntityID returns the appropriate entity ID based on type
func (c *Collaborator) GetEntityID() string {
	if c.IsUserCollaborator() && c.UserID != nil {
		return *c.UserID
	}
	if c.IsOrganizationCollaborator() && c.OrganizationID != nil {
		return *c.OrganizationID
	}
	return c.AAAEntityID
}

// IsActive checks if the collaborator is active
func (c *Collaborator) IsActive() bool {
	return c.Status == CollaboratorStatusActive
}

// CanManageUsers checks if the collaborator can manage other users
func (c *Collaborator) CanManageUsers() bool {
	return c.Role == CollaboratorRoleOwner || c.Role == CollaboratorRoleAdmin || c.Role == CollaboratorRoleManager
}

// CanAccessFinancials checks if the collaborator can access financial information
func (c *Collaborator) CanAccessFinancials() bool {
	return c.Role == CollaboratorRoleOwner || c.Role == CollaboratorRoleAdmin
}

// HasHigherRoleThan checks if this collaborator has a higher role than another
func (c *Collaborator) HasHigherRoleThan(other *Collaborator) bool {
	roleHierarchy := map[CollaboratorRole]int{
		CollaboratorRoleViewer:     1,
		CollaboratorRoleEmployee:   2,
		CollaboratorRoleContractor: 2,
		CollaboratorRoleManager:    3,
		CollaboratorRoleAdmin:      4,
		CollaboratorRoleOwner:      5,
	}

	thisLevel := roleHierarchy[c.Role]
	otherLevel := roleHierarchy[other.Role]

	return thisLevel > otherLevel
}

// Activate activates the collaborator
func (c *Collaborator) Activate() {
	c.Status = CollaboratorStatusActive
	if c.AcceptedAt == nil {
		now := time.Now()
		c.AcceptedAt = &now
	}
}

// Deactivate deactivates the collaborator
func (c *Collaborator) Deactivate() {
	c.Status = CollaboratorStatusInactive
}

// Suspend suspends the collaborator
func (c *Collaborator) Suspend() {
	c.Status = CollaboratorStatusSuspended
}
