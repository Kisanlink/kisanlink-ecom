package collaborator

import (
	"time"

	"kisanlink-ecom/entities/models/collaborator"
)

// CollaboratorResponse represents the response for a collaborator
type CollaboratorResponse struct {
	ID               string                          `json:"id"`
	UserID           string                          `json:"user_id"`
	Username         string                          `json:"username"`
	Email            string                          `json:"email"`
	FirstName        string                          `json:"first_name"`
	LastName         string                          `json:"last_name"`
	FullName         string                          `json:"full_name"`
	Phone            string                          `json:"phone,omitempty"`
	Avatar           string                          `json:"avatar,omitempty"`
	Bio              string                          `json:"bio,omitempty"`
	Location         string                          `json:"location,omitempty"`
	Coordinates      string                          `json:"coordinates,omitempty"`
	CollaboratorType collaborator.CollaboratorType   `json:"collaborator_type"`
	Status           collaborator.CollaboratorStatus `json:"status"`
	OrganizationID   string                          `json:"organization_id,omitempty"`

	// Business Information
	BusinessName        string `json:"business_name,omitempty"`
	BusinessType        string `json:"business_type,omitempty"`
	BusinessLicense     string `json:"business_license,omitempty"`
	TaxID               string `json:"tax_id,omitempty"`
	BusinessAddress     string `json:"business_address,omitempty"`
	BusinessPhone       string `json:"business_phone,omitempty"`
	BusinessEmail       string `json:"business_email,omitempty"`
	BusinessWebsite     string `json:"business_website,omitempty"`
	BusinessDescription string `json:"business_description,omitempty"`

	// Platform Activity
	LastLoginAt    *time.Time `json:"last_login_at,omitempty"`
	LastActivityAt *time.Time `json:"last_activity_at,omitempty"`
	LoginCount     int        `json:"login_count"`

	// Verification & Trust
	IsVerified      bool       `json:"is_verified"`
	VerifiedAt      *time.Time `json:"verified_at,omitempty"`
	VerifiedBy      string     `json:"verified_by,omitempty"`
	TrustScore      float64    `json:"trust_score"`
	CompletedOrders int        `json:"completed_orders"`
	CancelledOrders int        `json:"cancelled_orders"`
	AverageRating   float64    `json:"average_rating"`
	TotalReviews    int        `json:"total_reviews"`

	// Preferences & Settings
	LanguagePreference string `json:"language_preference"`
	TimezonePreference string `json:"timezone_preference"`

	// Metadata
	OnboardingCompleted bool       `json:"onboarding_completed"`
	OnboardingStep      int        `json:"onboarding_step"`
	Tags                []string   `json:"tags,omitempty"`
	InvitedBy           string     `json:"invited_by,omitempty"`
	InvitedAt           *time.Time `json:"invited_at,omitempty"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// CollaboratorSummaryResponse represents a summary response for a collaborator
type CollaboratorSummaryResponse struct {
	ID               string                          `json:"id"`
	UserID           string                          `json:"user_id"`
	Username         string                          `json:"username"`
	Email            string                          `json:"email"`
	FullName         string                          `json:"full_name"`
	Avatar           string                          `json:"avatar,omitempty"`
	CollaboratorType collaborator.CollaboratorType   `json:"collaborator_type"`
	Status           collaborator.CollaboratorStatus `json:"status"`
	OrganizationID   string                          `json:"organization_id,omitempty"`
	Location         string                          `json:"location,omitempty"`
	BusinessName     string                          `json:"business_name,omitempty"`
	IsVerified       bool                            `json:"is_verified"`
	TrustScore       float64                         `json:"trust_score"`
	LastActivityAt   *time.Time                      `json:"last_activity_at,omitempty"`
	CreatedAt        time.Time                       `json:"created_at"`
}

// CollaboratorStatsResponse represents statistics for collaborators
type CollaboratorStatsResponse struct {
	TotalCollaborators    int                                     `json:"total_collaborators"`
	ActiveCollaborators   int                                     `json:"active_collaborators"`
	PendingCollaborators  int                                     `json:"pending_collaborators"`
	VerifiedCollaborators int                                     `json:"verified_collaborators"`
	CollaboratorsByType   map[collaborator.CollaboratorType]int   `json:"collaborators_by_type"`
	CollaboratorsByStatus map[collaborator.CollaboratorStatus]int `json:"collaborators_by_status"`
	AverageTrustScore     float64                                 `json:"average_trust_score"`
	OnboardingCompletion  float64                                 `json:"onboarding_completion_rate"`
	RecentRegistrations   int                                     `json:"recent_registrations_7_days"`
	ActiveInLast30Days    int                                     `json:"active_in_last_30_days"`
}

// CollaboratorProfileResponse represents a detailed profile response
type CollaboratorProfileResponse struct {
	CollaboratorResponse

	// Additional profile information
	Preferences          map[string]interface{} `json:"preferences,omitempty"`
	NotificationSettings map[string]interface{} `json:"notification_settings,omitempty"`

	// Financial information (only for the collaborator themselves or admins)
	BankAccountNumber string `json:"bank_account_number,omitempty"`
	BankName          string `json:"bank_name,omitempty"`
	BankBranch        string `json:"bank_branch,omitempty"`
	IFSCCode          string `json:"ifsc_code,omitempty"`
	UPIId             string `json:"upi_id,omitempty"`

	// Admin notes (only for admins)
	Notes string `json:"notes,omitempty"`
}

// CollaboratorVerificationResponse represents the response after verification
type CollaboratorVerificationResponse struct {
	ID         string     `json:"id"`
	IsVerified bool       `json:"is_verified"`
	VerifiedAt *time.Time `json:"verified_at"`
	VerifiedBy string     `json:"verified_by"`
	Message    string     `json:"message"`
}

// BulkOperationResponse represents the response for bulk operations
type BulkOperationResponse struct {
	TotalRequested int      `json:"total_requested"`
	Successful     int      `json:"successful"`
	Failed         int      `json:"failed"`
	FailedIDs      []string `json:"failed_ids,omitempty"`
	Errors         []string `json:"errors,omitempty"`
	Message        string   `json:"message"`
}

// OnboardingStepResponse represents the response for onboarding step update
type OnboardingStepResponse struct {
	ID                  string `json:"id"`
	OnboardingStep      int    `json:"onboarding_step"`
	OnboardingCompleted bool   `json:"onboarding_completed"`
	NextStep            string `json:"next_step,omitempty"`
	Message             string `json:"message"`
}
