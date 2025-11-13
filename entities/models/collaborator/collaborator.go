package collaborator

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// CollaboratorStatus represents the status of a collaborator
type CollaboratorStatus string

const (
	CollaboratorStatusActive    CollaboratorStatus = "ACTIVE"
	CollaboratorStatusInactive  CollaboratorStatus = "INACTIVE"
	CollaboratorStatusSuspended CollaboratorStatus = "SUSPENDED"
	CollaboratorStatusPending   CollaboratorStatus = "PENDING"
)

// CollaboratorType represents the type of collaborator
type CollaboratorType string

const (
	CollaboratorTypeFarmer   CollaboratorType = "FARMER"
	CollaboratorTypeSupplier CollaboratorType = "SUPPLIER"
	CollaboratorTypeBuyer    CollaboratorType = "BUYER"
	CollaboratorTypeAgent    CollaboratorType = "AGENT"
	CollaboratorTypeAdmin    CollaboratorType = "ADMIN"
)

// Collaborator represents a user collaborator in the platform
// This stores platform-specific data while user identity comes from AAA service
type Collaborator struct {
	base.BaseModel `gorm:"embedded"`

	// AAA Service Reference
	UserID   string `json:"user_id" gorm:"type:varchar(255);not null;uniqueIndex" validate:"required"`
	Username string `json:"username" gorm:"type:varchar(100);not null;index" validate:"required"`
	Email    string `json:"email" gorm:"type:varchar(255);not null;index" validate:"required,email"`

	// Collaborator Profile
	FirstName   string `json:"first_name" gorm:"type:varchar(100)" validate:"required"`
	LastName    string `json:"last_name" gorm:"type:varchar(100)" validate:"required"`
	Phone       string `json:"phone" gorm:"type:varchar(20);index"`
	Avatar      string `json:"avatar" gorm:"type:varchar(500)"`
	Bio         string `json:"bio" gorm:"type:text"`
	Location    string `json:"location" gorm:"type:varchar(255)"`
	Coordinates string `json:"coordinates" gorm:"type:varchar(100)"` // lat,lng format

	// Platform Specific
	CollaboratorType CollaboratorType   `json:"collaborator_type" gorm:"type:varchar(50);not null;index" validate:"required"`
	Status           CollaboratorStatus `json:"status" gorm:"type:varchar(50);not null;default:'PENDING'" validate:"required"`
	OrganizationID   string             `json:"organization_id" gorm:"type:varchar(255);index"`

	// Business Information
	BusinessName        string `json:"business_name" gorm:"type:varchar(255)"`
	BusinessType        string `json:"business_type" gorm:"type:varchar(100)"`
	BusinessLicense     string `json:"business_license" gorm:"type:varchar(255)"`
	TaxID               string `json:"tax_id" gorm:"type:varchar(100)"`
	BusinessAddress     string `json:"business_address" gorm:"type:text"`
	BusinessPhone       string `json:"business_phone" gorm:"type:varchar(20)"`
	BusinessEmail       string `json:"business_email" gorm:"type:varchar(255)"`
	BusinessWebsite     string `json:"business_website" gorm:"type:varchar(500)"`
	BusinessDescription string `json:"business_description" gorm:"type:text"`

	// Platform Activity
	LastLoginAt    *time.Time `json:"last_login_at"`
	LastActivityAt *time.Time `json:"last_activity_at"`
	LoginCount     int        `json:"login_count" gorm:"default:0"`

	// Verification & Trust
	IsVerified      bool       `json:"is_verified" gorm:"default:false"`
	VerifiedAt      *time.Time `json:"verified_at"`
	VerifiedBy      string     `json:"verified_by" gorm:"type:varchar(255)"`
	TrustScore      float64    `json:"trust_score" gorm:"default:0.0"`
	CompletedOrders int        `json:"completed_orders" gorm:"default:0"`
	CancelledOrders int        `json:"cancelled_orders" gorm:"default:0"`
	AverageRating   float64    `json:"average_rating" gorm:"default:0.0"`
	TotalReviews    int        `json:"total_reviews" gorm:"default:0"`

	// Preferences & Settings
	Preferences          string `json:"preferences" gorm:"type:jsonb"`           // JSON object for user preferences
	NotificationSettings string `json:"notification_settings" gorm:"type:jsonb"` // JSON object for notification preferences
	LanguagePreference   string `json:"language_preference" gorm:"type:varchar(10);default:'en'"`
	TimezonePreference   string `json:"timezone_preference" gorm:"type:varchar(50);default:'UTC'"`

	// Financial Information
	BankAccountNumber string `json:"bank_account_number" gorm:"type:varchar(100)"`
	BankName          string `json:"bank_name" gorm:"type:varchar(255)"`
	BankBranch        string `json:"bank_branch" gorm:"type:varchar(255)"`
	IFSCCode          string `json:"ifsc_code" gorm:"type:varchar(20)"`
	UPIId             string `json:"upi_id" gorm:"type:varchar(100)"`

	// Metadata
	OnboardingCompleted bool       `json:"onboarding_completed" gorm:"default:false"`
	OnboardingStep      int        `json:"onboarding_step" gorm:"default:0"`
	Tags                string     `json:"tags" gorm:"type:jsonb"` // JSON array of tags
	Notes               string     `json:"notes" gorm:"type:text"`
	InvitedBy           string     `json:"invited_by" gorm:"type:varchar(255)"`
	InvitedAt           *time.Time `json:"invited_at"`
}

// TableName returns the table name for GORM
func (Collaborator) TableName() string {
	return "collaborators"
}

// NewCollaborator creates a new collaborator instance
func NewCollaborator(userID, username, email, firstName, lastName string, collaboratorType CollaboratorType) *Collaborator {
	return &Collaborator{
		UserID:             userID,
		Username:           username,
		Email:              email,
		FirstName:          firstName,
		LastName:           lastName,
		CollaboratorType:   collaboratorType,
		Status:             CollaboratorStatusPending,
		TrustScore:         0.0,
		LoginCount:         0,
		CompletedOrders:    0,
		CancelledOrders:    0,
		AverageRating:      0.0,
		TotalReviews:       0,
		OnboardingStep:     0,
		LanguagePreference: "en",
		TimezonePreference: "UTC",
	}
}

// GetFullName returns the full name of the collaborator
func (c *Collaborator) GetFullName() string {
	return c.FirstName + " " + c.LastName
}

// IsActive returns true if the collaborator is active
func (c *Collaborator) IsActive() bool {
	return c.Status == CollaboratorStatusActive
}

// CanLogin returns true if the collaborator can login
func (c *Collaborator) CanLogin() bool {
	return c.Status == CollaboratorStatusActive || c.Status == CollaboratorStatusPending
}

// UpdateLastActivity updates the last activity timestamp
func (c *Collaborator) UpdateLastActivity() {
	now := time.Now()
	c.LastActivityAt = &now
}

// UpdateLastLogin updates the last login timestamp and increments login count
func (c *Collaborator) UpdateLastLogin() {
	now := time.Now()
	c.LastLoginAt = &now
	c.LoginCount++
}

// Activate activates the collaborator
func (c *Collaborator) Activate() {
	c.Status = CollaboratorStatusActive
}

// Suspend suspends the collaborator
func (c *Collaborator) Suspend() {
	c.Status = CollaboratorStatusSuspended
}

// Deactivate deactivates the collaborator
func (c *Collaborator) Deactivate() {
	c.Status = CollaboratorStatusInactive
}

// Verify marks the collaborator as verified
func (c *Collaborator) Verify(verifiedBy string) {
	c.IsVerified = true
	now := time.Now()
	c.VerifiedAt = &now
	c.VerifiedBy = verifiedBy
}

// UpdateTrustScore updates the trust score based on activity
func (c *Collaborator) UpdateTrustScore() {
	// Simple trust score calculation
	// This can be enhanced with more sophisticated algorithms
	baseScore := 50.0

	// Add points for completed orders
	orderScore := float64(c.CompletedOrders) * 2.0

	// Subtract points for cancelled orders
	cancelScore := float64(c.CancelledOrders) * -5.0

	// Add points for ratings
	ratingScore := c.AverageRating * 10.0

	// Add points for verification
	verificationScore := 0.0
	if c.IsVerified {
		verificationScore = 20.0
	}

	c.TrustScore = baseScore + orderScore + cancelScore + ratingScore + verificationScore

	// Ensure score is between 0 and 100
	if c.TrustScore < 0 {
		c.TrustScore = 0
	}
	if c.TrustScore > 100 {
		c.TrustScore = 100
	}
}

// CompleteOnboarding marks onboarding as completed
func (c *Collaborator) CompleteOnboarding() {
	c.OnboardingCompleted = true
	c.OnboardingStep = 0
}

// SetOnboardingStep sets the current onboarding step
func (c *Collaborator) SetOnboardingStep(step int) {
	c.OnboardingStep = step
}
