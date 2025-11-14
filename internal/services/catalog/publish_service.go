package catalog

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/catalog"
	catalogRepo "github.com/Kisanlink/kisanlink-ecom/internal/repositories/catalog"
)

// PublishService handles product publishing business logic
type PublishService interface {
	// PublishProduct publishes a product to FPOs with validation
	PublishProduct(ctx context.Context, req PublishProductRequest) (*catalog.PublishState, error)

	// GetPublishStatus retrieves publish status for a product
	GetPublishStatus(ctx context.Context, productID string) (*catalog.PublishState, error)

	// UpdateDeliveryCosts updates delivery costs for specific FPOs
	UpdateDeliveryCosts(ctx context.Context, productID string, costs map[string]decimal.Decimal) error

	// RevokeAccess revokes FPO access to a product
	RevokeAccess(ctx context.Context, productID, fpoOrgID string) error

	// CalculateRetailPrice calculates FPO-specific retail price
	CalculateRetailPrice(basePrice, deliveryCost, platformFeePercent decimal.Decimal) decimal.Decimal

	// ValidateFPOAccess validates if FPO has access to product
	ValidateFPOAccess(ctx context.Context, fpoOrgID, productID string) (bool, error)

	// GetProductsForFPO retrieves all products visible to an FPO with pricing
	GetProductsForFPO(ctx context.Context, fpoOrgID string, offset, limit int) ([]*catalog.ProductWithFPOPricing, error)
}

// PublishProductRequest represents a product publishing request
type PublishProductRequest struct {
	ProductID          string
	FPOIDs             []string
	DeliveryCosts      map[string]decimal.Decimal
	PlatformFeePercent decimal.Decimal
	PublishedBy        string
}

// publishService implements PublishService
type publishService struct {
	publishStateRepo catalogRepo.PublishStateRepository
	catalogRepo      CatalogRepositoryInterface
	logger           *logrus.Logger
}

// NewPublishService creates a new publish service
func NewPublishService(
	publishStateRepo catalogRepo.PublishStateRepository,
	catalogRepo CatalogRepositoryInterface,
	logger *logrus.Logger,
) PublishService {
	return &publishService{
		publishStateRepo: publishStateRepo,
		catalogRepo:      catalogRepo,
		logger:           logger,
	}
}

// PublishProduct publishes a product to FPOs with validation
func (s *publishService) PublishProduct(ctx context.Context, req PublishProductRequest) (*catalog.PublishState, error) {
	// Validate product exists
	product := &catalog.CatalogItem{}
	_, err := s.catalogRepo.GetByID(ctx, req.ProductID, product)
	if err != nil {
		s.logger.WithError(err).WithField("product_id", req.ProductID).Error("Product not found for publishing")
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Validate product is active
	if !product.IsActive {
		s.logger.WithField("product_id", req.ProductID).Warn("Attempted to publish inactive product")
		return nil, fmt.Errorf("cannot publish inactive product: %s", req.ProductID)
	}

	// Validate platform fee is within range (0-100%)
	if req.PlatformFeePercent.LessThan(decimal.Zero) || req.PlatformFeePercent.GreaterThan(decimal.NewFromInt(100)) {
		s.logger.WithField("platform_fee", req.PlatformFeePercent).Error("Invalid platform fee percentage")
		return nil, fmt.Errorf("platform fee must be between 0 and 100, got: %s", req.PlatformFeePercent.String())
	}

	// Validate FPO IDs are provided
	if len(req.FPOIDs) == 0 {
		s.logger.Error("No FPO IDs provided for publishing")
		return nil, fmt.Errorf("at least one FPO ID is required")
	}

	// Validate all delivery costs are positive
	for fpoID, cost := range req.DeliveryCosts {
		if cost.LessThan(decimal.Zero) {
			s.logger.WithFields(logrus.Fields{
				"fpo_id":        fpoID,
				"delivery_cost": cost,
			}).Error("Invalid negative delivery cost")
			return nil, fmt.Errorf("delivery cost for FPO %s cannot be negative: %s", fpoID, cost.String())
		}
	}

	// Validate all FPO IDs have delivery costs
	for _, fpoID := range req.FPOIDs {
		if _, exists := req.DeliveryCosts[fpoID]; !exists {
			s.logger.WithField("fpo_id", fpoID).Error("Missing delivery cost for FPO")
			return nil, fmt.Errorf("delivery cost not provided for FPO: %s", fpoID)
		}
	}

	// Check if product is already published
	existingState, err := s.publishStateRepo.GetByProductID(ctx, req.ProductID)
	if err == nil && existingState != nil {
		s.logger.WithField("product_id", req.ProductID).Warn("Product already published, updating publish state")

		// Update existing publish state
		existingState.FPOAccessList = req.FPOIDs
		existingState.DeliveryCosts = req.DeliveryCosts
		existingState.PlatformFeePercent = req.PlatformFeePercent
		existingState.UpdatedBy = req.PublishedBy
		now := time.Now()
		existingState.PublishedAt = &now

		if err := s.publishStateRepo.Update(ctx, existingState); err != nil {
			s.logger.WithError(err).Error("Failed to update existing publish state")
			return nil, fmt.Errorf("failed to update publish state: %w", err)
		}

		s.logger.WithFields(logrus.Fields{
			"product_id": req.ProductID,
			"fpo_count":  len(req.FPOIDs),
		}).Info("Successfully updated product publish state")

		return existingState, nil
	}

	// Create new publish state
	publishState := catalog.NewPublishState(req.ProductID, req.PublishedBy)
	publishState.FPOAccessList = req.FPOIDs
	publishState.DeliveryCosts = req.DeliveryCosts
	publishState.PlatformFeePercent = req.PlatformFeePercent

	if err := s.publishStateRepo.Create(ctx, publishState); err != nil {
		s.logger.WithError(err).WithField("product_id", req.ProductID).Error("Failed to create publish state")
		return nil, fmt.Errorf("failed to create publish state: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"product_id":       req.ProductID,
		"fpo_count":        len(req.FPOIDs),
		"platform_fee":     req.PlatformFeePercent,
		"publish_state_id": publishState.ID,
	}).Info("Successfully published product to FPOs")

	return publishState, nil
}

// GetPublishStatus retrieves publish status for a product
func (s *publishService) GetPublishStatus(ctx context.Context, productID string) (*catalog.PublishState, error) {
	publishState, err := s.publishStateRepo.GetByProductID(ctx, productID)
	if err != nil {
		s.logger.WithError(err).WithField("product_id", productID).Error("Failed to get publish status")
		return nil, fmt.Errorf("failed to get publish status: %w", err)
	}

	return publishState, nil
}

// UpdateDeliveryCosts updates delivery costs for specific FPOs
func (s *publishService) UpdateDeliveryCosts(ctx context.Context, productID string, costs map[string]decimal.Decimal) error {
	// Validate all costs are non-negative
	for fpoID, cost := range costs {
		if cost.LessThan(decimal.Zero) {
			s.logger.WithFields(logrus.Fields{
				"fpo_id":        fpoID,
				"delivery_cost": cost,
			}).Error("Invalid negative delivery cost in update")
			return fmt.Errorf("delivery cost for FPO %s cannot be negative: %s", fpoID, cost.String())
		}
	}

	// Get existing publish state
	publishState, err := s.publishStateRepo.GetByProductID(ctx, productID)
	if err != nil {
		s.logger.WithError(err).WithField("product_id", productID).Error("Failed to get publish state for cost update")
		return fmt.Errorf("failed to get publish state: %w", err)
	}

	// Update delivery costs
	for fpoID, cost := range costs {
		publishState.SetDeliveryCost(fpoID, cost)
	}

	if err := s.publishStateRepo.Update(ctx, publishState); err != nil {
		s.logger.WithError(err).WithField("product_id", productID).Error("Failed to update delivery costs")
		return fmt.Errorf("failed to update delivery costs: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"product_id": productID,
		"fpo_count":  len(costs),
	}).Info("Successfully updated delivery costs")

	return nil
}

// RevokeAccess revokes FPO access to a product
func (s *publishService) RevokeAccess(ctx context.Context, productID, fpoOrgID string) error {
	// Get publish state
	publishState, err := s.publishStateRepo.GetByProductID(ctx, productID)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"product_id": productID,
			"fpo_org_id": fpoOrgID,
		}).Error("Failed to get publish state for revocation")
		return fmt.Errorf("failed to get publish state: %w", err)
	}

	// Check if FPO has access
	if !publishState.HasFPOAccess(fpoOrgID) {
		s.logger.WithFields(logrus.Fields{
			"product_id": productID,
			"fpo_org_id": fpoOrgID,
		}).Warn("Attempted to revoke access for FPO that doesn't have access")
		return fmt.Errorf("FPO %s does not have access to product %s", fpoOrgID, productID)
	}

	// Remove FPO from access list (keeps delivery cost history for audit)
	if err := s.publishStateRepo.RemoveFPOAccess(ctx, productID, fpoOrgID); err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"product_id": productID,
			"fpo_org_id": fpoOrgID,
		}).Error("Failed to revoke FPO access")
		return fmt.Errorf("failed to revoke access: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"product_id": productID,
		"fpo_org_id": fpoOrgID,
	}).Info("Successfully revoked FPO access to product")

	return nil
}

// CalculateRetailPrice calculates FPO-specific retail price
// Formula: RetailPrice = BasePrice + DeliveryCost + (BasePrice × PlatformFee%)
func (s *publishService) CalculateRetailPrice(basePrice, deliveryCost, platformFeePercent decimal.Decimal) decimal.Decimal {
	// Calculate commission
	commission := basePrice.Mul(platformFeePercent).Div(decimal.NewFromInt(100))

	// Calculate retail price
	retailPrice := basePrice.Add(deliveryCost).Add(commission)

	return retailPrice
}

// ValidateFPOAccess validates if FPO has access to product
func (s *publishService) ValidateFPOAccess(ctx context.Context, fpoOrgID, productID string) (bool, error) {
	hasAccess, err := s.publishStateRepo.HasFPOAccess(ctx, productID, fpoOrgID)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"product_id": productID,
			"fpo_org_id": fpoOrgID,
		}).Error("Failed to validate FPO access")
		return false, fmt.Errorf("failed to validate FPO access: %w", err)
	}

	return hasAccess, nil
}

// GetProductsForFPO retrieves all products visible to an FPO with pricing
func (s *publishService) GetProductsForFPO(ctx context.Context, fpoOrgID string, offset, limit int) ([]*catalog.ProductWithFPOPricing, error) {
	// Get all publish states visible to this FPO
	publishStates, total, err := s.publishStateRepo.GetProductsVisibleToFPO(ctx, fpoOrgID, offset, limit)
	if err != nil {
		s.logger.WithError(err).WithField("fpo_org_id", fpoOrgID).Error("Failed to get published products for FPO")
		return nil, fmt.Errorf("failed to get published products: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"fpo_org_id":    fpoOrgID,
		"product_count": total,
		"offset":        offset,
		"limit":         limit,
	}).Debug("Retrieved published products for FPO")

	// Enrich with product details and pricing
	productsWithPricing := make([]*catalog.ProductWithFPOPricing, 0, len(publishStates))

	for _, publishState := range publishStates {
		// Get product details
		productItem := &catalog.CatalogItem{}
		_, err := s.catalogRepo.GetByID(ctx, publishState.ProductID, productItem)
		if err != nil {
			s.logger.WithError(err).WithField("product_id", publishState.ProductID).Warn("Failed to get product details, skipping")
			continue
		}

		// Skip inactive products
		if !productItem.IsActive {
			s.logger.WithField("product_id", publishState.ProductID).Debug("Skipping inactive product")
			continue
		}

		// Convert to Product
		product := catalog.Product{
			CatalogItem: *productItem,
		}

		// Calculate FPO-specific pricing with 30-minute price lock
		pricing := publishState.GetPricingForFPO(productItem.BasePrice, fpoOrgID, 30)

		// Create enriched product
		productWithPricing := &catalog.ProductWithFPOPricing{
			Product: product,
			Pricing: pricing,
		}

		productsWithPricing = append(productsWithPricing, productWithPricing)
	}

	s.logger.WithFields(logrus.Fields{
		"fpo_org_id":      fpoOrgID,
		"enriched_count":  len(productsWithPricing),
		"total_published": total,
	}).Info("Successfully enriched products with FPO pricing")

	return productsWithPricing, nil
}
