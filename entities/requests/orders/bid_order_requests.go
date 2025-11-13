package orders

import (
	"github.com/shopspring/decimal"
)

// CreateOrderFromBidRequest represents the request to create an order from a winning bid
type CreateOrderFromBidRequest struct {
	BidID           string   `json:"bid_id" binding:"required" validate:"required" example:"BID_1234567890"`
	ShippingAddress *Address `json:"shipping_address" binding:"required" validate:"required"`
	PaymentMethod   string   `json:"payment_method" binding:"required" validate:"required,oneof=credit_card debit_card bank_transfer upi cash_on_delivery" example:"upi"`
	Notes           string   `json:"notes" validate:"omitempty,max=1000" example:"Please deliver during business hours"`
}

// CreateOrderFromBidResponse represents the response after creating an order from a bid
type CreateOrderFromBidResponse struct {
	OrderID     string          `json:"order_id" example:"ORD_1234567890"`
	OrderNumber string          `json:"order_number" example:"1234567890"`
	BidID       string          `json:"bid_id" example:"BID_1234567890"`
	ListingID   string          `json:"listing_id" example:"LST_1234567890"`
	TotalAmount decimal.Decimal `json:"total_amount" example:"150.00"`
	Currency    string          `json:"currency" example:"INR"`
	Status      string          `json:"status" example:"PENDING_PAYMENT"`
	Message     string          `json:"message" example:"Order created successfully from winning bid"`
}

// BidOrderValidationRequest represents the request to validate a bid for order creation
type BidOrderValidationRequest struct {
	BidID string `json:"bid_id" binding:"required" validate:"required" example:"BID_1234567890"`
}

// BidOrderValidationResponse represents the response for bid validation
type BidOrderValidationResponse struct {
	Valid           bool            `json:"valid"`
	BidID           string          `json:"bid_id" example:"BID_1234567890"`
	ListingID       string          `json:"listing_id" example:"LST_1234567890"`
	ProductID       string          `json:"product_id" example:"PROD_1234567890"`
	WinningAmount   decimal.Decimal `json:"winning_amount" example:"150.00"`
	Quantity        decimal.Decimal `json:"quantity" example:"10.5"`
	Currency        string          `json:"currency" example:"INR"`
	SellerID        string          `json:"seller_id" example:"USER_1234567890"`
	BuyerID         string          `json:"buyer_id" example:"USER_0987654321"`
	SellerOrgID     string          `json:"seller_org_id" example:"ORG_1234567890"`
	BuyerOrgID      string          `json:"buyer_org_id" example:"ORG_0987654321"`
	ProductName     string          `json:"product_name" example:"Organic Tomatoes"`
	ProductSKU      string          `json:"product_sku" example:"ORG-TOM-001"`
	ExpiresAt       string          `json:"expires_at" example:"2024-12-31T23:59:59Z"`
	CanCreateOrder  bool            `json:"can_create_order"`
	ValidationError string          `json:"validation_error,omitempty"`
}

// ProcessPaymentRequest represents the request to process payment for an order
type ProcessPaymentRequest struct {
	PaymentMethod string `json:"payment_method" binding:"required" validate:"required,oneof=credit_card debit_card bank_transfer upi cash_on_delivery digital_wallet net_banking" example:"upi"`
}

// PaymentResponse represents the response after processing payment
type PaymentResponse struct {
	PaymentID     string          `json:"payment_id" example:"PAY_1234567890"`
	OrderID       string          `json:"order_id" example:"ORD_1234567890"`
	Status        string          `json:"status" example:"COMPLETED"`
	Amount        decimal.Decimal `json:"amount" example:"150.00"`
	Currency      string          `json:"currency" example:"INR"`
	PaymentMethod string          `json:"payment_method" example:"upi"`
	ProcessedAt   string          `json:"processed_at" example:"2024-12-31T23:59:59Z"`
	Message       string          `json:"message" example:"Payment completed successfully"`
}
