package orders

import (
	"context"
	"fmt"
	"time"

	"kisanlink-ecom/entities/models/orders"
	orderRequests "kisanlink-ecom/entities/requests/orders"
	repositoryCommon "kisanlink-ecom/internal/repositories/common"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// PaymentScreenshotRepository extends BaseFilterableRepository with payment screenshot-specific methods
type PaymentScreenshotRepository struct {
	*base.BaseFilterableRepository[*orders.PaymentScreenshot]
	*repositoryCommon.BaseRepository
	dbManager db.DBManager
}

// NewPaymentScreenshotRepository creates a new payment screenshot repository
func NewPaymentScreenshotRepository(dbManager db.DBManager) *PaymentScreenshotRepository {
	baseRepo := base.NewBaseFilterableRepository[*orders.PaymentScreenshot]()
	baseRepo.SetDBManager(dbManager)
	return &PaymentScreenshotRepository{
		BaseFilterableRepository: baseRepo,
		BaseRepository:           repositoryCommon.NewBaseRepository(dbManager),
		dbManager:                dbManager,
	}
}

// CreatePaymentScreenshot creates a new payment screenshot record
func (r *PaymentScreenshotRepository) CreatePaymentScreenshot(ctx context.Context, screenshot *orders.PaymentScreenshot) error {
	if screenshot == nil {
		return fmt.Errorf("payment screenshot cannot be nil")
	}
	if screenshot.OrderID == "" {
		return fmt.Errorf("order ID is required")
	}
	if screenshot.UploadedBy == "" {
		return fmt.Errorf("uploaded by user ID is required")
	}
	if screenshot.FilePath == "" {
		return fmt.Errorf("file path is required")
	}

	return r.Create(ctx, screenshot)
}

// GetPaymentScreenshot retrieves a payment screenshot by ID
func (r *PaymentScreenshotRepository) GetPaymentScreenshot(ctx context.Context, id string) (*orders.PaymentScreenshot, error) {
	if id == "" {
		return nil, fmt.Errorf("screenshot ID is required")
	}

	screenshot := &orders.PaymentScreenshot{}
	result, err := r.GetByID(ctx, id, screenshot)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment screenshot: %w", err)
	}

	return result, nil
}

// GetByOrderID retrieves all payment screenshots for a specific order
func (r *PaymentScreenshotRepository) GetByOrderID(ctx context.Context, orderID string) ([]*orders.PaymentScreenshot, error) {
	if orderID == "" {
		return nil, fmt.Errorf("order ID is required")
	}

	filter := &base.Filter{
		Group: base.FilterGroup{
			Logic: base.LogicAnd,
			Conditions: []base.FilterCondition{
				{
					Field:    "order_id",
					Operator: base.OpEqual,
					Value:    orderID,
				},
			},
		},
	}

	screenshots, err := r.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment screenshots for order: %w", err)
	}

	return screenshots, nil
}

// ListPaymentScreenshots retrieves payment screenshots with filtering, pagination, and sorting
func (r *PaymentScreenshotRepository) ListPaymentScreenshots(ctx context.Context, filters *orderRequests.PaymentScreenshotFilters) ([]*orders.PaymentScreenshot, int64, error) {
	if filters == nil {
		return nil, 0, fmt.Errorf("filters cannot be nil")
	}

	// Build filter conditions
	conditions := []base.FilterCondition{}

	if filters.OrderID != nil {
		conditions = append(conditions, base.FilterCondition{
			Field:    "order_id",
			Operator: base.OpEqual,
			Value:    *filters.OrderID,
		})
	}

	if filters.VerificationStatus != nil {
		conditions = append(conditions, base.FilterCondition{
			Field:    "verification_status",
			Operator: base.OpEqual,
			Value:    string(*filters.VerificationStatus),
		})
	}

	if filters.BuyerOrganizationID != nil {
		conditions = append(conditions, base.FilterCondition{
			Field:    "buyer_organization_id",
			Operator: base.OpEqual,
			Value:    *filters.BuyerOrganizationID,
		})
	}

	if filters.SellerOrganizationID != nil {
		conditions = append(conditions, base.FilterCondition{
			Field:    "seller_organization_id",
			Operator: base.OpEqual,
			Value:    *filters.SellerOrganizationID,
		})
	}

	if filters.UploadedBy != nil {
		conditions = append(conditions, base.FilterCondition{
			Field:    "uploaded_by",
			Operator: base.OpEqual,
			Value:    *filters.UploadedBy,
		})
	}

	if filters.PaymentMethod != nil {
		conditions = append(conditions, base.FilterCondition{
			Field:    "payment_method",
			Operator: base.OpEqual,
			Value:    string(*filters.PaymentMethod),
		})
	}

	if filters.CreatedAfter != nil {
		createdAfter, err := time.Parse(time.RFC3339, *filters.CreatedAfter)
		if err == nil {
			conditions = append(conditions, base.FilterCondition{
				Field:    "created_at",
				Operator: base.OpGreaterEqual,
				Value:    createdAfter,
			})
		}
	}

	if filters.CreatedBefore != nil {
		createdBefore, err := time.Parse(time.RFC3339, *filters.CreatedBefore)
		if err == nil {
			conditions = append(conditions, base.FilterCondition{
				Field:    "created_at",
				Operator: base.OpLessEqual,
				Value:    createdBefore,
			})
		}
	}

	// Build filter with pagination and sorting
	filter := &base.Filter{
		Group: base.FilterGroup{
			Logic:      base.LogicAnd,
			Conditions: conditions,
		},
		Page:     filters.Page,
		PageSize: filters.PageSize,
		Sort: []base.SortField{
			{
				Field:     filters.SortBy,
				Direction: filters.SortOrder,
			},
		},
	}

	screenshots, err := r.Find(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find payment screenshots: %w", err)
	}

	// Get total count
	countFilter := &base.Filter{
		Group: filter.Group,
	}
	total, err := r.CountWithFilter(ctx, countFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count payment screenshots: %w", err)
	}

	return screenshots, total, nil
}

// GetPendingScreenshots retrieves payment screenshots with pending verification status
func (r *PaymentScreenshotRepository) GetPendingScreenshots(ctx context.Context, page, limit int) ([]*orders.PaymentScreenshot, int64, error) {
	pendingStatus := orders.VerificationStatusPending
	filters := &orderRequests.PaymentScreenshotFilters{
		VerificationStatus: &pendingStatus,
		Page:               page,
		PageSize:           limit,
		SortBy:             "created_at",
		SortOrder:          "desc",
	}

	return r.ListPaymentScreenshots(ctx, filters)
}

// UpdateVerificationStatus updates the verification status of a payment screenshot
func (r *PaymentScreenshotRepository) UpdateVerificationStatus(ctx context.Context, id string, status orders.VerificationStatus, verifiedBy string, notes string) error {
	if id == "" {
		return fmt.Errorf("screenshot ID is required")
	}
	if status == "" {
		return fmt.Errorf("verification status is required")
	}
	if verifiedBy == "" {
		return fmt.Errorf("verified by user ID is required")
	}

	// Get the screenshot first
	screenshot, err := r.GetPaymentScreenshot(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get payment screenshot: %w", err)
	}

	// Update fields
	now := time.Now()
	screenshot.VerificationStatus = status
	screenshot.VerifiedBy = &verifiedBy
	screenshot.VerifiedAt = &now
	screenshot.VerificationNotes = notes
	screenshot.UpdatedBy = verifiedBy
	screenshot.UpdatedAt = now

	// Update using dbManager
	if err := r.dbManager.Update(ctx, screenshot); err != nil {
		return fmt.Errorf("failed to update verification status: %w", err)
	}

	return nil
}

// DeletePaymentScreenshot soft deletes a payment screenshot
func (r *PaymentScreenshotRepository) DeletePaymentScreenshot(ctx context.Context, id string, deletedBy string) error {
	if id == "" {
		return fmt.Errorf("screenshot ID is required")
	}
	if deletedBy == "" {
		return fmt.Errorf("deleted by user ID is required")
	}

	// Use the Delete method from BaseFilterableRepository
	screenshot := &orders.PaymentScreenshot{}

	if err := r.Delete(ctx, id, screenshot); err != nil {
		return fmt.Errorf("failed to delete payment screenshot: %w", err)
	}

	return nil
}

// GetScreenshotsByBuyerOrganization retrieves payment screenshots for a buyer organization
func (r *PaymentScreenshotRepository) GetScreenshotsByBuyerOrganization(ctx context.Context, orgID string, page, limit int) ([]*orders.PaymentScreenshot, int64, error) {
	if orgID == "" {
		return nil, 0, fmt.Errorf("organization ID is required")
	}

	filters := &orderRequests.PaymentScreenshotFilters{
		BuyerOrganizationID: &orgID,
		Page:                page,
		PageSize:            limit,
		SortBy:              "created_at",
		SortOrder:           "desc",
	}

	return r.ListPaymentScreenshots(ctx, filters)
}

// GetScreenshotsBySellerOrganization retrieves payment screenshots for a seller organization
func (r *PaymentScreenshotRepository) GetScreenshotsBySellerOrganization(ctx context.Context, orgID string, page, limit int) ([]*orders.PaymentScreenshot, int64, error) {
	if orgID == "" {
		return nil, 0, fmt.Errorf("organization ID is required")
	}

	filters := &orderRequests.PaymentScreenshotFilters{
		SellerOrganizationID: &orgID,
		Page:                 page,
		PageSize:             limit,
		SortBy:               "created_at",
		SortOrder:            "desc",
	}

	return r.ListPaymentScreenshots(ctx, filters)
}

// CheckDuplicateChecksum checks if a payment screenshot with the same checksum exists for the order
func (r *PaymentScreenshotRepository) CheckDuplicateChecksum(ctx context.Context, orderID, checksum string) (bool, error) {
	if orderID == "" || checksum == "" {
		return false, fmt.Errorf("order ID and checksum are required")
	}

	// Build filter to find duplicates
	filter := &base.Filter{
		Group: base.FilterGroup{
			Logic: base.LogicAnd,
			Conditions: []base.FilterCondition{
				{
					Field:    "order_id",
					Operator: base.OpEqual,
					Value:    orderID,
				},
				{
					Field:    "file_checksum",
					Operator: base.OpEqual,
					Value:    checksum,
				},
			},
		},
	}

	// Use CountWithFilter to check existence
	count, err := r.CountWithFilter(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check duplicate checksum: %w", err)
	}

	return count > 0, nil
}
