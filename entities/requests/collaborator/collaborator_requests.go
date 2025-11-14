package collaborator

import (
	"github.com/Kisanlink/kisanlink-ecom/entities/models/collaborator"
)

// CreateCollaboratorRequest represents the request to create a new collaborator
type CreateCollaboratorRequest struct {
	UserID           string                        `json:"user_id" validate:"required"`
	Username         string                        `json:"username" validate:"required,min=3,max=50"`
	Email            string                        `json:"email" validate:"required,email"`
	FirstName        string                        `json:"first_name" validate:"required,min=1,max=100"`
	LastName         string                        `json:"last_name" validate:"required,min=1,max=100"`
	Phone            string                        `json:"phone" validate:"omitempty,min=10,max=20"`
	CollaboratorType collaborator.CollaboratorType `json:"collaborator_type" validate:"required"`
	OrganizationID   string                        `json:"organization_id" validate:"omitempty"`

	// Optional profile information
	Bio         string `json:"bio" validate:"omitempty,max=1000"`
	Location    string `json:"location" validate:"omitempty,max=255"`
	Coordinates string `json:"coordinates" validate:"omitempty,max=100"`

	// Optional business information
	BusinessName        string `json:"business_name" validate:"omitempty,max=255"`
	BusinessType        string `json:"business_type" validate:"omitempty,max=100"`
	BusinessLicense     string `json:"business_license" validate:"omitempty,max=255"`
	TaxID               string `json:"tax_id" validate:"omitempty,max=100"`
	BusinessAddress     string `json:"business_address" validate:"omitempty,max=1000"`
	BusinessPhone       string `json:"business_phone" validate:"omitempty,min=10,max=20"`
	BusinessEmail       string `json:"business_email" validate:"omitempty,email"`
	BusinessWebsite     string `json:"business_website" validate:"omitempty,url"`
	BusinessDescription string `json:"business_description" validate:"omitempty,max=1000"`

	// Optional preferences
	LanguagePreference string `json:"language_preference" validate:"omitempty,len=2"`
	TimezonePreference string `json:"timezone_preference" validate:"omitempty,max=50"`

	// Invitation information
	InvitedBy string `json:"invited_by" validate:"omitempty"`
}

// UpdateCollaboratorRequest represents the request to update a collaborator
type UpdateCollaboratorRequest struct {
	FirstName   *string `json:"first_name" validate:"omitempty,min=1,max=100"`
	LastName    *string `json:"last_name" validate:"omitempty,min=1,max=100"`
	Phone       *string `json:"phone" validate:"omitempty,min=10,max=20"`
	Avatar      *string `json:"avatar" validate:"omitempty,url"`
	Bio         *string `json:"bio" validate:"omitempty,max=1000"`
	Location    *string `json:"location" validate:"omitempty,max=255"`
	Coordinates *string `json:"coordinates" validate:"omitempty,max=100"`

	// Business information updates
	BusinessName        *string `json:"business_name" validate:"omitempty,max=255"`
	BusinessType        *string `json:"business_type" validate:"omitempty,max=100"`
	BusinessLicense     *string `json:"business_license" validate:"omitempty,max=255"`
	TaxID               *string `json:"tax_id" validate:"omitempty,max=100"`
	BusinessAddress     *string `json:"business_address" validate:"omitempty,max=1000"`
	BusinessPhone       *string `json:"business_phone" validate:"omitempty,min=10,max=20"`
	BusinessEmail       *string `json:"business_email" validate:"omitempty,email"`
	BusinessWebsite     *string `json:"business_website" validate:"omitempty,url"`
	BusinessDescription *string `json:"business_description" validate:"omitempty,max=1000"`

	// Preferences updates
	LanguagePreference *string `json:"language_preference" validate:"omitempty,len=2"`
	TimezonePreference *string `json:"timezone_preference" validate:"omitempty,max=50"`

	// Financial information updates
	BankAccountNumber *string `json:"bank_account_number" validate:"omitempty,max=100"`
	BankName          *string `json:"bank_name" validate:"omitempty,max=255"`
	BankBranch        *string `json:"bank_branch" validate:"omitempty,max=255"`
	IFSCCode          *string `json:"ifsc_code" validate:"omitempty,max=20"`
	UPIId             *string `json:"upi_id" validate:"omitempty,max=100"`

	// Metadata updates
	Tags  *string `json:"tags" validate:"omitempty"`
	Notes *string `json:"notes" validate:"omitempty,max=2000"`
}

// UpdateCollaboratorStatusRequest represents the request to update collaborator status
type UpdateCollaboratorStatusRequest struct {
	Status collaborator.CollaboratorStatus `json:"status" validate:"required"`
	Reason string                          `json:"reason" validate:"omitempty,max=500"`
}

// VerifyCollaboratorRequest represents the request to verify a collaborator
type VerifyCollaboratorRequest struct {
	VerificationNotes string `json:"verification_notes" validate:"omitempty,max=1000"`
}

// CollaboratorFilter represents filters for listing collaborators
type CollaboratorFilter struct {
	CollaboratorType    *collaborator.CollaboratorType   `json:"collaborator_type"`
	Status              *collaborator.CollaboratorStatus `json:"status"`
	OrganizationID      *string                          `json:"organization_id"`
	IsVerified          *bool                            `json:"is_verified"`
	OnboardingCompleted *bool                            `json:"onboarding_completed"`
	Location            *string                          `json:"location"`
	BusinessType        *string                          `json:"business_type"`
	MinTrustScore       *float64                         `json:"min_trust_score"`
	MaxTrustScore       *float64                         `json:"max_trust_score"`
	Search              *string                          `json:"search"`             // Search in name, email, username, business name
	Tags                []string                         `json:"tags"`               // Filter by tags
	CreatedAfter        *string                          `json:"created_after"`      // ISO date string
	CreatedBefore       *string                          `json:"created_before"`     // ISO date string
	LastActiveAfter     *string                          `json:"last_active_after"`  // ISO date string
	LastActiveBefore    *string                          `json:"last_active_before"` // ISO date string
}

// UpdateOnboardingStepRequest represents the request to update onboarding step
type UpdateOnboardingStepRequest struct {
	Step int `json:"step" validate:"required,min=0,max=10"`
}

// BulkUpdateCollaboratorsRequest represents the request to bulk update collaborators
type BulkUpdateCollaboratorsRequest struct {
	CollaboratorIDs []string                         `json:"collaborator_ids" validate:"required,min=1"`
	Status          *collaborator.CollaboratorStatus `json:"status"`
	OrganizationID  *string                          `json:"organization_id"`
	Tags            *string                          `json:"tags"`
	Reason          string                           `json:"reason" validate:"omitempty,max=500"`
}
