package orders

import (
	"time"

	"github.com/shopspring/decimal"
)

// OrderFromBidResponse represents the complete response after creating an order from a bid
type OrderFromBidResponse struct {
	Order   *OrderResponse `json:"order"`
	Bid     *BidInfo       `json:"bid"`
	Listing *ListingInfo   `json:"listing"`
	Message string         `json:"message"`
}

// BidInfo represents bid information in the order response
type BidInfo struct {
	BidID     string          `json:"bid_id"`
	BidAmount decimal.Decimal `json:"bid_amount"`
	Currency  string          `json:"currency"`
	PlacedAt  time.Time       `json:"placed_at"`
	IsWinning bool            `json:"is_winning"`
	BidderID  string          `json:"bidder_id"`
}

// ListingInfo represents listing information in the order response
type ListingInfo struct {
	ListingID   string          `json:"listing_id"`
	ProductID   string          `json:"product_id"`
	ProductName string          `json:"product_name"`
	ProductSKU  string          `json:"product_sku"`
	Quantity    decimal.Decimal `json:"quantity"`
	SellerID    string          `json:"seller_id"`
	SellerOrgID string          `json:"seller_org_id"`
	ClosedAt    *time.Time      `json:"closed_at"`
}

// BidOrderValidationErrorResponse represents validation error response
type BidOrderValidationErrorResponse struct {
	Valid           bool   `json:"valid"`
	ErrorCode       string `json:"error_code"`
	ErrorMessage    string `json:"error_message"`
	ValidationError string `json:"validation_error"`
}
