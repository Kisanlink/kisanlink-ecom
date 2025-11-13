package marketplace

import (
	"fmt"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/shopspring/decimal"
)

// BidStatus represents the possible bid states
type BidStatus string

const (
	BidStatusActive  BidStatus = "ACTIVE"
	BidStatusOutbid  BidStatus = "OUTBID"
	BidStatusWinning BidStatus = "WINNING"
	BidStatusExpired BidStatus = "EXPIRED"
	BidStatusRemoved BidStatus = "REMOVED"
)

// Bid represents a bid placed on a marketplace listing
type Bid struct {
	base.BaseModel

	// Basic Information
	BidID     string `json:"bid_id" gorm:"type:varchar(50);uniqueIndex;not null"`
	ListingID string `json:"listing_id" gorm:"type:varchar(50);not null;index"`
	BidderID  string `json:"bidder_id" gorm:"type:varchar(50);not null;index"`

	// Bid Details
	BidAmount decimal.Decimal `json:"bid_amount" gorm:"type:decimal(10,2);not null"`
	Currency  string          `json:"currency" gorm:"type:varchar(3);not null;default:'INR'"`
	Quantity  decimal.Decimal `json:"quantity" gorm:"type:decimal(10,3);not null"`
	Message   string          `json:"message" gorm:"type:text"`

	// Auto-bidding
	AutoBidLimit *decimal.Decimal `json:"auto_bid_limit" gorm:"type:decimal(10,2)"`
	IsAutoBid    bool             `json:"is_auto_bid" gorm:"default:false"`
	ParentBidID  *string          `json:"parent_bid_id" gorm:"type:varchar(50)"`

	// Status
	Status       BidStatus  `json:"status" gorm:"type:varchar(20);not null;default:'ACTIVE';index"`
	IsHighestBid bool       `json:"is_highest_bid" gorm:"default:false;index"`
	OutbidAt     *time.Time `json:"outbid_at"`

	// Payment
	PaymentMethod string `json:"payment_method" gorm:"type:varchar(50)"`

	// Audit
	PlacedAt time.Time `json:"placed_at" gorm:"not null;index"`
}

// TableName returns the table name for GORM
func (Bid) TableName() string {
	return "marketplace_bids"
}

// NewBid creates a new marketplace bid
func NewBid(listingID, bidderID string, bidAmount, quantity decimal.Decimal, message string) *Bid {
	return &Bid{
		BaseModel:    *base.NewBaseModel("BID", "large"),
		BidID:        generateBidID(),
		ListingID:    listingID,
		BidderID:     bidderID,
		BidAmount:    bidAmount,
		Currency:     "INR",
		Quantity:     quantity,
		Message:      message,
		IsAutoBid:    false,
		Status:       BidStatusActive,
		IsHighestBid: false,
		PlacedAt:     time.Now(),
	}
}

// NewAutoBid creates a new auto-bid with a limit
func NewAutoBid(listingID, bidderID string, bidAmount, quantity, autoBidLimit decimal.Decimal, parentBidID string) *Bid {
	bid := NewBid(listingID, bidderID, bidAmount, quantity, "Auto-bid")
	bid.IsAutoBid = true
	bid.AutoBidLimit = &autoBidLimit
	bid.ParentBidID = &parentBidID
	return bid
}

// generateBidID generates a unique bid ID
func generateBidID() string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("BID_%d", timestamp)
}

// IsActive checks if the bid is currently active
func (b *Bid) IsActive() bool {
	return b.Status == BidStatusActive
}

// IsWinning checks if the bid is currently winning
func (b *Bid) IsWinning() bool {
	return b.Status == BidStatusWinning || b.IsHighestBid
}

// CanBeOutbid checks if the bid can be outbid
func (b *Bid) CanBeOutbid() bool {
	return b.Status == BidStatusActive || b.Status == BidStatusWinning
}

// CanTransitionTo checks if the bid can transition to the given status
func (b *Bid) CanTransitionTo(newStatus BidStatus) bool {
	switch b.Status {
	case BidStatusActive:
		return newStatus == BidStatusOutbid || newStatus == BidStatusWinning ||
			newStatus == BidStatusExpired || newStatus == BidStatusRemoved
	case BidStatusWinning:
		return newStatus == BidStatusOutbid || newStatus == BidStatusExpired
	case BidStatusOutbid, BidStatusExpired, BidStatusRemoved:
		return false // Terminal states for most cases
	default:
		return false
	}
}

// UpdateStatus updates the bid status with audit information
func (b *Bid) UpdateStatus(newStatus BidStatus) error {
	if !b.CanTransitionTo(newStatus) {
		return fmt.Errorf("invalid status transition from %s to %s", b.Status, newStatus)
	}

	b.Status = newStatus

	// Update audit fields based on status change
	switch newStatus {
	case BidStatusOutbid:
		now := time.Now()
		b.OutbidAt = &now
		b.IsHighestBid = false
	case BidStatusWinning:
		b.IsHighestBid = true
	case BidStatusExpired, BidStatusRemoved:
		b.IsHighestBid = false
	}

	return nil
}

// SetAsHighestBid marks this bid as the current highest bid
func (b *Bid) SetAsHighestBid() {
	b.IsHighestBid = true
	if b.Status == BidStatusActive {
		b.Status = BidStatusWinning
	}
}

// Outbid marks this bid as outbid by another bid
func (b *Bid) Outbid() error {
	return b.UpdateStatus(BidStatusOutbid)
}

// HasAutoBidCapacity checks if the auto-bid has remaining capacity
func (b *Bid) HasAutoBidCapacity(newBidAmount decimal.Decimal) bool {
	if !b.IsAutoBid || b.AutoBidLimit == nil {
		return false
	}
	return newBidAmount.LessThan(*b.AutoBidLimit)
}

// GetNextAutoBidAmount calculates the next auto-bid amount
func (b *Bid) GetNextAutoBidAmount(competingBidAmount decimal.Decimal, minimumIncrement decimal.Decimal) decimal.Decimal {
	if !b.IsAutoBid || b.AutoBidLimit == nil {
		return decimal.Zero
	}

	nextBidAmount := competingBidAmount.Add(minimumIncrement)
	if nextBidAmount.LessThanOrEqual(*b.AutoBidLimit) {
		return nextBidAmount
	}

	// If the next increment would exceed the limit, bid up to the limit
	if b.AutoBidLimit.GreaterThan(competingBidAmount) {
		return *b.AutoBidLimit
	}

	return decimal.Zero // Cannot auto-bid further
}

// BidSummary represents a summary view of a bid
type BidSummary struct {
	ID           string          `json:"id"`
	BidID        string          `json:"bid_id"`
	ListingID    string          `json:"listing_id"`
	BidderID     string          `json:"bidder_id"`
	BidAmount    decimal.Decimal `json:"bid_amount"`
	Status       BidStatus       `json:"status"`
	IsHighestBid bool            `json:"is_highest_bid"`
	IsAutoBid    bool            `json:"is_auto_bid"`
	PlacedAt     time.Time       `json:"placed_at"`
	OutbidAt     *time.Time      `json:"outbid_at,omitempty"`
}

// BidFilter represents filters for bid queries
type BidFilter struct {
	Status       *BidStatus       `json:"status,omitempty"`
	BidderID     string           `json:"bidder_id,omitempty"`
	ListingID    string           `json:"listing_id,omitempty"`
	IsHighestBid *bool            `json:"is_highest_bid,omitempty"`
	IsAutoBid    *bool            `json:"is_auto_bid,omitempty"`
	MinAmount    *decimal.Decimal `json:"min_amount,omitempty"`
	MaxAmount    *decimal.Decimal `json:"max_amount,omitempty"`
	PlacedAfter  *time.Time       `json:"placed_after,omitempty"`
	PlacedBefore *time.Time       `json:"placed_before,omitempty"`
}

// BidHistory represents the complete bid history for a listing
type BidHistory struct {
	ListingID   string       `json:"listing_id"`
	TotalBids   int          `json:"total_bids"`
	HighestBid  *BidSummary  `json:"highest_bid,omitempty"`
	Bids        []BidSummary `json:"bids"`
	LastUpdated time.Time    `json:"last_updated"`
}

// AnonymousBidSummary represents a bid summary with anonymized bidder information
type AnonymousBidSummary struct {
	BidAmount    decimal.Decimal `json:"bid_amount"`
	AnonymousID  string          `json:"anonymous_id"` // e.g., "Bidder #1"
	PlacedAt     time.Time       `json:"placed_at"`
	IsHighestBid bool            `json:"is_highest_bid"`
	IsAutoBid    bool            `json:"is_auto_bid"`
}

// BidStatistics represents statistical information about bids for a listing
type BidStatistics struct {
	ListingID     string          `json:"listing_id"`
	TotalBids     int             `json:"total_bids"`
	UniqueBidders int             `json:"unique_bidders"`
	HighestBid    decimal.Decimal `json:"highest_bid"`
	AverageBid    decimal.Decimal `json:"average_bid"`
	BidIncrement  decimal.Decimal `json:"bid_increment"`
	LastBidTime   *time.Time      `json:"last_bid_time,omitempty"`
	AutoBidCount  int             `json:"auto_bid_count"`
}
