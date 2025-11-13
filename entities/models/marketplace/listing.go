package marketplace

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/shopspring/decimal"
)

// ListingStatus represents the possible listing states
type ListingStatus string

const (
	ListingStatusActive        ListingStatus = "ACTIVE"
	ListingStatusClosed        ListingStatus = "CLOSED"
	ListingStatusExpired       ListingStatus = "EXPIRED"
	ListingStatusCancelled     ListingStatus = "CANCELLED"
	ListingStatusExpiredNoBids ListingStatus = "EXPIRED_NO_BIDS"
)

// ListingVisibility represents the visibility level of listings
type ListingVisibility string

const (
	VisibilityPrivate      ListingVisibility = "PRIVATE"      // Only invited participants
	VisibilityPublic       ListingVisibility = "PUBLIC"       // Open to all users
	VisibilityNetwork      ListingVisibility = "NETWORK"      // Network/partner organizations
	VisibilityOrganization ListingVisibility = "ORGANIZATION" // Same organization only
)

// AuctionType represents the auction transparency level
type AuctionType string

const (
	AuctionTypeOpen   AuctionType = "OPEN"   // All bid prices visible during auction
	AuctionTypeClosed AuctionType = "CLOSED" // Bid prices hidden until auction ends
)

// BidVisibility represents the level of bid information visibility
type BidVisibility string

const (
	BidVisibilityFull    BidVisibility = "FULL"    // Show all bid details based on auction type
	BidVisibilityPartial BidVisibility = "PARTIAL" // Show bid count and highest amount only
	BidVisibilityMinimal BidVisibility = "MINIMAL" // Show only bid count
	BidVisibilityHidden  BidVisibility = "HIDDEN"  // No bid information visible
)

// Location represents pickup/delivery location information
type Location struct {
	Address    string   `json:"address"`
	City       string   `json:"city"`
	State      string   `json:"state"`
	PostalCode string   `json:"postal_code"`
	Country    string   `json:"country"`
	Latitude   *float64 `json:"latitude,omitempty"`
	Longitude  *float64 `json:"longitude,omitempty"`
}

// Listing represents a marketplace auction listing
type Listing struct {
	base.BaseModel

	// Basic Information
	ListingID      string `json:"listing_id" gorm:"type:varchar(50);uniqueIndex;not null"`
	ProductID      string `json:"product_id" gorm:"type:varchar(50);not null;index"`
	SellerID       string `json:"seller_id" gorm:"type:varchar(50);not null;index"`
	OrganizationID string `json:"organization_id" gorm:"type:varchar(50);not null;index"`

	// Auction Parameters
	Quantity    decimal.Decimal `json:"quantity" gorm:"type:decimal(10,3);not null"`
	AskingPrice decimal.Decimal `json:"asking_price" gorm:"type:decimal(10,2);not null"`
	MinimumBid  decimal.Decimal `json:"minimum_bid" gorm:"type:decimal(10,2);not null"`
	Currency    string          `json:"currency" gorm:"type:varchar(3);not null;default:'INR'"`

	// Timing
	ListingDuration int       `json:"listing_duration_hours" gorm:"not null"`
	ExpiresAt       time.Time `json:"expires_at" gorm:"not null;index"`

	// Status
	Status              ListingStatus `json:"status" gorm:"type:varchar(20);not null;default:'ACTIVE';index"`
	CurrentHighestBidID *string       `json:"current_highest_bid_id" gorm:"type:varchar(50)"`
	BidCount            int           `json:"bid_count" gorm:"default:0"`

	// Visibility & Auction Configuration
	Visibility    ListingVisibility `json:"visibility" gorm:"type:varchar(20);not null;default:'PUBLIC'"`
	AuctionType   AuctionType       `json:"auction_type" gorm:"type:varchar(10);not null;default:'OPEN'"`
	BidVisibility BidVisibility     `json:"bid_visibility" gorm:"type:varchar(20);not null;default:'FULL'"`

	// Additional Details
	PickupLocation  string `json:"pickup_location" gorm:"type:jsonb"`
	TermsConditions string `json:"terms_conditions" gorm:"type:text"`
	ListingType     string `json:"listing_type" gorm:"type:varchar(20);default:'AUCTION'"`

	// Audit
	ClosedAt    *time.Time `json:"closed_at"`
	CloseReason *string    `json:"close_reason" gorm:"type:varchar(50)"`
}

// TableName returns the table name for GORM
func (Listing) TableName() string {
	return "marketplace_listings"
}

// NewListing creates a new marketplace listing
func NewListing(productID, sellerID, organizationID string, quantity, askingPrice, minimumBid decimal.Decimal, durationHours int) *Listing {
	now := time.Now()
	expiresAt := now.Add(time.Duration(durationHours) * time.Hour)

	return &Listing{
		BaseModel:       *base.NewBaseModel("LST", "large"),
		ListingID:       generateListingID(),
		ProductID:       productID,
		SellerID:        sellerID,
		OrganizationID:  organizationID,
		Quantity:        quantity,
		AskingPrice:     askingPrice,
		MinimumBid:      minimumBid,
		Currency:        "INR",
		ListingDuration: durationHours,
		ExpiresAt:       expiresAt,
		Status:          ListingStatusActive,
		BidCount:        0,
		Visibility:      VisibilityPublic,
		AuctionType:     AuctionTypeOpen,
		BidVisibility:   BidVisibilityFull,
		ListingType:     "AUCTION",
	}
}

// generateListingID generates a unique listing ID
func generateListingID() string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("LST_%d", timestamp)
}

// IsActive checks if the listing is currently active
func (l *Listing) IsActive() bool {
	return l.Status == ListingStatusActive && time.Now().Before(l.ExpiresAt)
}

// IsExpired checks if the listing has expired
func (l *Listing) IsExpired() bool {
	return time.Now().After(l.ExpiresAt)
}

// CanAcceptBids checks if the listing can accept new bids
func (l *Listing) CanAcceptBids() bool {
	return l.Status == ListingStatusActive && !l.IsExpired()
}

// CanTransitionTo checks if the listing can transition to the given status
func (l *Listing) CanTransitionTo(newStatus ListingStatus) bool {
	switch l.Status {
	case ListingStatusActive:
		return newStatus == ListingStatusClosed || newStatus == ListingStatusExpired ||
			newStatus == ListingStatusCancelled || newStatus == ListingStatusExpiredNoBids
	case ListingStatusClosed, ListingStatusExpired, ListingStatusCancelled, ListingStatusExpiredNoBids:
		return false // Terminal states
	default:
		return false
	}
}

// UpdateStatus updates the listing status with audit information
func (l *Listing) UpdateStatus(newStatus ListingStatus, reason string) error {
	if !l.CanTransitionTo(newStatus) {
		return fmt.Errorf("invalid status transition from %s to %s", l.Status, newStatus)
	}

	l.Status = newStatus
	now := time.Now()
	l.ClosedAt = &now
	if reason != "" {
		l.CloseReason = &reason
	}

	return nil
}

// SetPickupLocation sets the pickup location from a Location struct
func (l *Listing) SetPickupLocation(location *Location) error {
	if location == nil {
		l.PickupLocation = ""
		return nil
	}

	locationJSON, err := json.Marshal(location)
	if err != nil {
		return fmt.Errorf("failed to marshal pickup location: %w", err)
	}

	l.PickupLocation = string(locationJSON)
	return nil
}

// GetPickupLocation returns the pickup location as a Location struct
func (l *Listing) GetPickupLocation() (*Location, error) {
	if l.PickupLocation == "" {
		return nil, nil
	}

	var location Location
	if err := json.Unmarshal([]byte(l.PickupLocation), &location); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pickup location: %w", err)
	}

	return &location, nil
}

// CanBeViewedBy checks if the listing can be viewed by the given user/organization
func (l *Listing) CanBeViewedBy(viewerOrgID string) bool {
	switch l.Visibility {
	case VisibilityPrivate:
		// For private listings, additional invitation logic would be needed
		return l.OrganizationID == viewerOrgID
	case VisibilityPublic:
		return true
	case VisibilityNetwork:
		// For network visibility, network membership logic would be needed
		return true // Simplified for now
	case VisibilityOrganization:
		return l.OrganizationID == viewerOrgID
	default:
		return false
	}
}

// GetTimeRemaining returns the time remaining until expiry
func (l *Listing) GetTimeRemaining() time.Duration {
	if l.IsExpired() {
		return 0
	}
	return time.Until(l.ExpiresAt)
}

// ListingSummary represents a summary view of a listing
type ListingSummary struct {
	ID            string            `json:"id"`
	ListingID     string            `json:"listing_id"`
	ProductID     string            `json:"product_id"`
	Status        ListingStatus     `json:"status"`
	AskingPrice   decimal.Decimal   `json:"asking_price"`
	MinimumBid    decimal.Decimal   `json:"minimum_bid"`
	BidCount      int               `json:"bid_count"`
	TimeRemaining string            `json:"time_remaining"`
	Visibility    ListingVisibility `json:"visibility"`
	AuctionType   AuctionType       `json:"auction_type"`
	CreatedAt     time.Time         `json:"created_at"`
	ExpiresAt     time.Time         `json:"expires_at"`
}

// ListingFilter represents filters for listing queries
type ListingFilter struct {
	Status         *ListingStatus     `json:"status,omitempty"`
	SellerID       string             `json:"seller_id,omitempty"`
	OrganizationID string             `json:"organization_id,omitempty"`
	ProductID      string             `json:"product_id,omitempty"`
	Visibility     *ListingVisibility `json:"visibility,omitempty"`
	AuctionType    *AuctionType       `json:"auction_type,omitempty"`
	MinPrice       *decimal.Decimal   `json:"min_price,omitempty"`
	MaxPrice       *decimal.Decimal   `json:"max_price,omitempty"`
	ExpiresAfter   *time.Time         `json:"expires_after,omitempty"`
	ExpiresBefore  *time.Time         `json:"expires_before,omitempty"`
}
