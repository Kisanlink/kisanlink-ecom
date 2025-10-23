package inventory

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	catalogModels "kisanlink-ecom/entities/models/catalog"
	inventoryRepo "kisanlink-ecom/internal/repositories/inventory"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/shopspring/decimal"
)

// InventoryService defines the interface for inventory business logic
type InventoryService interface {
	// Inventory lot management
	CreateInventoryLot(ctx context.Context, req *CreateInventoryLotRequest, userID, orgID string) (*catalogModels.InventoryLot, error)
	GetInventoryLot(ctx context.Context, lotID, userID, orgID string) (*catalogModels.InventoryLot, error)
	ListInventoryLots(ctx context.Context, filter *InventoryFilter, userID, orgID string) (*InventoryListResponse, error)
	UpdateInventoryLot(ctx context.Context, lotID string, req *UpdateInventoryLotRequest, userID, orgID string) (*catalogModels.InventoryLot, error)
	DeleteInventoryLot(ctx context.Context, lotID, userID, orgID string) error

	// Inventory operations
	ReserveInventory(ctx context.Context, catalogItemID string, quantity decimal.Decimal, userID, orgID string) ([]*catalogModels.InventoryLot, error)
	ReleaseInventory(ctx context.Context, catalogItemID string, quantity decimal.Decimal, userID, orgID string) error
	SellInventory(ctx context.Context, catalogItemID string, quantity decimal.Decimal, userID, orgID string) error
	AdjustInventory(ctx context.Context, lotID string, adjustment decimal.Decimal, reason, userID, orgID string) (*catalogModels.InventoryLot, error)

	// Inventory queries
	GetAvailableQuantity(ctx context.Context, catalogItemID, userID, orgID string) (decimal.Decimal, error)
	GetTotalQuantity(ctx context.Context, catalogItemID, userID, orgID string) (decimal.Decimal, error)
	CheckAvailability(ctx context.Context, catalogItemID string, requiredQuantity decimal.Decimal, userID, orgID string) (bool, decimal.Decimal, error)

	// Expiration handling
	ProcessExpiringLots(ctx context.Context, orgID string) ([]*catalogModels.InventoryLot, error)
	MarkLotExpired(ctx context.Context, lotID, userID, orgID string) error

	// Audit trail
	GetAuditTrail(ctx context.Context, lotID string, offset, limit int, userID, orgID string) (*AuditTrailResponse, error)
}

// CreateInventoryLotRequest represents a request to create an inventory lot
type CreateInventoryLotRequest struct {
	CatalogItemID     string           `json:"catalog_item_id" validate:"required"`
	LotNumber         string           `json:"lot_number" validate:"required"`
	BatchNumber       string           `json:"batch_number"`
	InitialQuantity   decimal.Decimal  `json:"initial_quantity" validate:"required,gt=0"`
	QualityGrade      string           `json:"quality_grade"`
	HarvestDate       *time.Time       `json:"harvest_date"`
	ExpiryDate        *time.Time       `json:"expiry_date"`
	WarehouseLocation string           `json:"warehouse_location"`
	StorageConditions string           `json:"storage_conditions"`
	LotPrice          *decimal.Decimal `json:"lot_price"`
	Metadata          string           `json:"metadata"`
}

// UpdateInventoryLotRequest represents a request to update an inventory lot
type UpdateInventoryLotRequest struct {
	QualityGrade      *string          `json:"quality_grade"`
	HarvestDate       *time.Time       `json:"harvest_date"`
	ExpiryDate        *time.Time       `json:"expiry_date"`
	WarehouseLocation *string          `json:"warehouse_location"`
	StorageConditions *string          `json:"storage_conditions"`
	LotPrice          *decimal.Decimal `json:"lot_price"`
	Metadata          *string          `json:"metadata"`
}

// InventoryFilter represents filters for listing inventory lots
type InventoryFilter struct {
	CatalogItemID  string                  `json:"catalog_item_id"`
	Status         catalogModels.LotStatus `json:"status"`
	ExpiringBefore *time.Time              `json:"expiring_before"`
	QualityGrade   string                  `json:"quality_grade"`
	Offset         int                     `json:"offset"`
	Limit          int                     `json:"limit"`
}

// InventoryListResponse represents a paginated list of inventory lots
type InventoryListResponse struct {
	Lots   []*catalogModels.InventoryLot `json:"lots"`
	Total  int64                         `json:"total"`
	Offset int                           `json:"offset"`
	Limit  int                           `json:"limit"`
}

// AuditTrailResponse represents a paginated list of audit logs
type AuditTrailResponse struct {
	Logs   []*inventoryRepo.InventoryAuditLog `json:"logs"`
	Total  int64                              `json:"total"`
	Offset int                                `json:"offset"`
	Limit  int                                `json:"limit"`
}

// CatalogRepository interface for inventory service dependencies
type CatalogRepository interface {
	GetByID(ctx context.Context, id string, model interface{}) (interface{}, error)
}

// inventoryService implements the InventoryService interface
type inventoryService struct {
	inventoryRepo inventoryRepo.InventoryRepository
	catalogRepo   CatalogRepository
}

// NewInventoryService creates a new inventory service
func NewInventoryService(inventoryRepo inventoryRepo.InventoryRepository, catalogRepo CatalogRepository) InventoryService {
	return &inventoryService{
		inventoryRepo: inventoryRepo,
		catalogRepo:   catalogRepo,
	}
}

// CreateInventoryLot creates a new inventory lot with product validation
func (s *inventoryService) CreateInventoryLot(ctx context.Context, req *CreateInventoryLotRequest, userID, orgID string) (*catalogModels.InventoryLot, error) {
	// Validate catalog item exists and is a product
	catalogItem := &catalogModels.CatalogItem{}
	_, err := s.catalogRepo.GetByID(ctx, req.CatalogItemID, catalogItem)
	if err != nil {
		return nil, fmt.Errorf("catalog item not found: %w", err)
	}

	// Check if catalog item belongs to the organization
	if catalogItem.OrganizationID != orgID {
		return nil, fmt.Errorf("catalog item does not belong to organization")
	}

	// Validate that it's a product (only products have inventory)
	if catalogItem.ItemType != catalogModels.CatalogItemTypeProduct {
		return nil, fmt.Errorf("inventory lots can only be created for products")
	}

	// Check if lot number is unique within the organization
	existingLot, err := s.inventoryRepo.GetByLotNumber(ctx, orgID, req.LotNumber)
	if err == nil && existingLot != nil {
		return nil, fmt.Errorf("lot number already exists in organization")
	}

	// Validate expiry date is in the future if provided
	if req.ExpiryDate != nil && req.ExpiryDate.Before(time.Now()) {
		return nil, fmt.Errorf("expiry date must be in the future")
	}

	// Create inventory lot
	lot := catalogModels.NewInventoryLot(orgID, req.LotNumber, req.InitialQuantity, "kg") // Default unit
	lot.BatchNumber = req.BatchNumber
	lot.QualityGrade = req.QualityGrade
	lot.HarvestDate = req.HarvestDate
	lot.ExpiryDate = req.ExpiryDate
	lot.CatalogItemID = &req.CatalogItemID
	lot.Warehouse = req.WarehouseLocation
	lot.Condition = req.StorageConditions
	if req.LotPrice != nil {
		lot.UnitCost = *req.LotPrice
	}
	// Set metadata - use sql.NullString to properly handle empty values
	if req.Metadata != "" {
		lot.Metadata = sql.NullString{String: req.Metadata, Valid: true}
	}
	// If empty, leave as NULL (Valid: false) which PostgreSQL jsonb accepts
	lot.CreatedBy = userID

	// Save to database
	if err := s.inventoryRepo.Create(ctx, lot); err != nil {
		return nil, fmt.Errorf("failed to create inventory lot: %w", err)
	}

	// Create audit log
	auditLog := &inventoryRepo.InventoryAuditLog{
		BaseModel:       *base.NewBaseModel("AUDIT", "small"),
		LotID:           lot.ID,
		Operation:       "create",
		QuantityBefore:  decimal.Zero,
		QuantityAfter:   req.InitialQuantity,
		QuantityChanged: req.InitialQuantity,
		Reason:          "Initial inventory lot creation",
		UserID:          userID,
		OrganizationID:  orgID,
	}

	if err := s.inventoryRepo.CreateAuditLog(ctx, auditLog); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to create audit log: %v\n", err)
	}

	return lot, nil
}

// GetInventoryLot retrieves an inventory lot by ID
func (s *inventoryService) GetInventoryLot(ctx context.Context, lotID, userID, orgID string) (*catalogModels.InventoryLot, error) {
	lot, err := s.inventoryRepo.GetByID(ctx, lotID)
	if err != nil {
		return nil, err
	}

	// Check organization access
	if lot.OrganizationID != orgID {
		return nil, fmt.Errorf("access denied: lot does not belong to organization")
	}

	return lot, nil
}

// ListInventoryLots lists inventory lots with filtering and pagination
func (s *inventoryService) ListInventoryLots(ctx context.Context, filter *InventoryFilter, userID, orgID string) (*InventoryListResponse, error) {
	var lots []*catalogModels.InventoryLot
	var total int64
	var err error

	// Set default pagination
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	// Apply filters
	if filter.CatalogItemID != "" {
		lots, total, err = s.inventoryRepo.ListByCatalogItem(ctx, filter.CatalogItemID, filter.Offset, filter.Limit)
	} else if filter.Status != "" {
		lots, total, err = s.inventoryRepo.ListByStatus(ctx, orgID, filter.Status, filter.Offset, filter.Limit)
	} else {
		lots, total, err = s.inventoryRepo.ListByOrganization(ctx, orgID, filter.Offset, filter.Limit)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list inventory lots: %w", err)
	}

	// Filter by expiring date if specified
	if filter.ExpiringBefore != nil {
		filteredLots := make([]*catalogModels.InventoryLot, 0)
		for _, lot := range lots {
			if lot.ExpiryDate != nil && lot.ExpiryDate.Before(*filter.ExpiringBefore) {
				filteredLots = append(filteredLots, lot)
			}
		}
		lots = filteredLots
		total = int64(len(lots))
	}

	return &InventoryListResponse{
		Lots:   lots,
		Total:  total,
		Offset: filter.Offset,
		Limit:  filter.Limit,
	}, nil
}

// UpdateInventoryLot updates an inventory lot
func (s *inventoryService) UpdateInventoryLot(ctx context.Context, lotID string, req *UpdateInventoryLotRequest, userID, orgID string) (*catalogModels.InventoryLot, error) {
	// Get existing lot
	lot, err := s.inventoryRepo.GetByID(ctx, lotID)
	if err != nil {
		return nil, err
	}

	// Check organization access
	if lot.OrganizationID != orgID {
		return nil, fmt.Errorf("access denied: lot does not belong to organization")
	}

	// Update fields if provided
	if req.QualityGrade != nil {
		lot.QualityGrade = *req.QualityGrade
	}
	if req.HarvestDate != nil {
		lot.HarvestDate = req.HarvestDate
	}
	if req.ExpiryDate != nil {
		// Validate expiry date is in the future
		if req.ExpiryDate.Before(time.Now()) {
			return nil, fmt.Errorf("expiry date must be in the future")
		}
		lot.ExpiryDate = req.ExpiryDate
	}
	if req.WarehouseLocation != nil {
		lot.Warehouse = *req.WarehouseLocation
	}
	if req.StorageConditions != nil {
		lot.Condition = *req.StorageConditions
	}
	if req.LotPrice != nil {
		lot.UnitCost = *req.LotPrice
	}
	if req.Metadata != nil {
		// Set metadata - use sql.NullString to properly handle empty values
		if *req.Metadata != "" {
			lot.Metadata = sql.NullString{String: *req.Metadata, Valid: true}
		}
		// If empty, leave as NULL (Valid: false) which PostgreSQL jsonb accepts
	}

	lot.UpdatedBy = userID

	// Save changes
	if err := s.inventoryRepo.Update(ctx, lot); err != nil {
		return nil, fmt.Errorf("failed to update inventory lot: %w", err)
	}

	// Create audit log
	auditLog := &inventoryRepo.InventoryAuditLog{
		BaseModel:       *base.NewBaseModel("AUDIT", "small"),
		LotID:           lot.ID,
		Operation:       "update",
		QuantityBefore:  lot.AvailableQty.Add(lot.ReservedQty).Add(lot.Quantity),
		QuantityAfter:   lot.AvailableQty.Add(lot.ReservedQty).Add(lot.Quantity),
		QuantityChanged: decimal.Zero,
		Reason:          "Inventory lot metadata update",
		UserID:          userID,
		OrganizationID:  orgID,
	}

	if err := s.inventoryRepo.CreateAuditLog(ctx, auditLog); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to create audit log: %v\n", err)
	}

	return lot, nil
}

// DeleteInventoryLot soft deletes an inventory lot
func (s *inventoryService) DeleteInventoryLot(ctx context.Context, lotID, userID, orgID string) error {
	// Get existing lot
	lot, err := s.inventoryRepo.GetByID(ctx, lotID)
	if err != nil {
		return err
	}

	// Check organization access
	if lot.OrganizationID != orgID {
		return fmt.Errorf("access denied: lot does not belong to organization")
	}

	// Check if lot has reserved or sold quantities
	if lot.ReservedQty.GreaterThan(decimal.Zero) || lot.Quantity.GreaterThan(decimal.Zero) {
		return fmt.Errorf("cannot delete lot with reserved or sold quantities")
	}

	// Delete lot
	if err := s.inventoryRepo.Delete(ctx, lotID); err != nil {
		return fmt.Errorf("failed to delete inventory lot: %w", err)
	}

	// Create audit log
	auditLog := &inventoryRepo.InventoryAuditLog{
		BaseModel:       *base.NewBaseModel("AUDIT", "small"),
		LotID:           lot.ID,
		Operation:       "delete",
		QuantityBefore:  lot.AvailableQty,
		QuantityAfter:   decimal.Zero,
		QuantityChanged: lot.AvailableQty.Neg(),
		Reason:          "Inventory lot deletion",
		UserID:          userID,
		OrganizationID:  orgID,
	}

	if err := s.inventoryRepo.CreateAuditLog(ctx, auditLog); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to create audit log: %v\n", err)
	}

	return nil
}

// ReserveInventory reserves inventory with audit trail
func (s *inventoryService) ReserveInventory(ctx context.Context, catalogItemID string, quantity decimal.Decimal, userID, orgID string) ([]*catalogModels.InventoryLot, error) {
	// Validate input
	if quantity.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("quantity must be greater than zero")
	}

	// Validate catalog item exists and belongs to organization
	catalogItem := &catalogModels.CatalogItem{}
	_, err := s.catalogRepo.GetByID(ctx, catalogItemID, catalogItem)
	if err != nil {
		return nil, fmt.Errorf("catalog item not found: %w", err)
	}

	if catalogItem.OrganizationID != orgID {
		return nil, fmt.Errorf("catalog item does not belong to organization")
	}

	if catalogItem.ItemType != catalogModels.CatalogItemTypeProduct {
		return nil, fmt.Errorf("inventory operations are only available for products")
	}

	// Check real-time availability
	available, err := s.inventoryRepo.GetAvailableQuantity(ctx, catalogItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to check availability: %w", err)
	}

	if available.LessThan(quantity) {
		return nil, fmt.Errorf("insufficient inventory: requested %s, available %s", quantity.String(), available.String())
	}

	// Store original quantities for audit trail
	originalLots := make(map[string]decimal.Decimal)

	// Get lots that will be affected for audit trail
	lotsBeforeReservation, _, err := s.inventoryRepo.ListByCatalogItem(ctx, catalogItemID, 0, 1000)
	if err == nil {
		for _, lot := range lotsBeforeReservation {
			if lot.Status == catalogModels.LotStatusActive && lot.AvailableQty.GreaterThan(decimal.Zero) {
				originalLots[lot.ID] = lot.AvailableQty
			}
		}
	}

	// Reserve inventory
	affectedLots, err := s.inventoryRepo.ReserveQuantity(ctx, catalogItemID, quantity)
	if err != nil {
		return nil, fmt.Errorf("failed to reserve inventory: %w", err)
	}

	// Create detailed audit logs for affected lots
	for _, lot := range affectedLots {
		originalAvailable := originalLots[lot.ID]
		quantityReserved := originalAvailable.Sub(lot.AvailableQty)

		auditLog := &inventoryRepo.InventoryAuditLog{
			BaseModel:       *base.NewBaseModel("AUDIT", "small"),
			LotID:           lot.ID,
			Operation:       "reserve",
			QuantityBefore:  originalAvailable,
			QuantityAfter:   lot.AvailableQty,
			QuantityChanged: quantityReserved.Neg(), // Negative because it's a reduction
			Reason:          fmt.Sprintf("Inventory reservation for catalog item %s", catalogItemID),
			UserID:          userID,
			OrganizationID:  orgID,
		}

		if err := s.inventoryRepo.CreateAuditLog(ctx, auditLog); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: failed to create audit log for lot %s: %v\n", lot.ID, err)
		}
	}

	return affectedLots, nil
}

// ReleaseInventory releases reserved inventory with audit trail
func (s *inventoryService) ReleaseInventory(ctx context.Context, catalogItemID string, quantity decimal.Decimal, userID, orgID string) error {
	// Validate input
	if quantity.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("quantity must be greater than zero")
	}

	// Validate catalog item exists and belongs to organization
	catalogItem := &catalogModels.CatalogItem{}
	_, err := s.catalogRepo.GetByID(ctx, catalogItemID, catalogItem)
	if err != nil {
		return fmt.Errorf("catalog item not found: %w", err)
	}

	if catalogItem.OrganizationID != orgID {
		return fmt.Errorf("catalog item does not belong to organization")
	}

	if catalogItem.ItemType != catalogModels.CatalogItemTypeProduct {
		return fmt.Errorf("inventory operations are only available for products")
	}

	// Check if there's enough reserved quantity to release
	lotsWithReserved, _, err := s.inventoryRepo.ListByCatalogItem(ctx, catalogItemID, 0, 1000)
	if err != nil {
		return fmt.Errorf("failed to get reserved lots: %w", err)
	}

	// Filter to only lots with reserved quantity
	var filteredLots []*catalogModels.InventoryLot
	for _, lot := range lotsWithReserved {
		if lot.ReservedQty.GreaterThan(decimal.Zero) {
			filteredLots = append(filteredLots, lot)
		}
	}
	lotsWithReserved = filteredLots

	totalReserved := decimal.Zero
	originalReserved := make(map[string]decimal.Decimal)
	for _, lot := range lotsWithReserved {
		totalReserved = totalReserved.Add(lot.ReservedQty)
		originalReserved[lot.ID] = lot.ReservedQty
	}

	if totalReserved.LessThan(quantity) {
		return fmt.Errorf("insufficient reserved inventory: requested %s, reserved %s", quantity.String(), totalReserved.String())
	}

	// Release inventory
	if err := s.inventoryRepo.ReleaseQuantity(ctx, catalogItemID, quantity); err != nil {
		return fmt.Errorf("failed to release inventory: %w", err)
	}

	// Create audit logs for affected lots
	// Get updated lots to see what changed
	updatedLots, _, err := s.inventoryRepo.ListByCatalogItem(ctx, catalogItemID, 0, 1000)
	if err == nil {
		for _, lot := range updatedLots {
			originalQty := originalReserved[lot.ID]
			if originalQty.GreaterThan(lot.ReservedQty) {
				quantityReleased := originalQty.Sub(lot.ReservedQty)

				auditLog := &inventoryRepo.InventoryAuditLog{
					BaseModel:       *base.NewBaseModel("AUDIT", "small"),
					LotID:           lot.ID,
					Operation:       "release",
					QuantityBefore:  originalQty,
					QuantityAfter:   lot.ReservedQty,
					QuantityChanged: quantityReleased.Neg(), // Negative because reserved quantity decreased
					Reason:          fmt.Sprintf("Inventory release for catalog item %s", catalogItemID),
					UserID:          userID,
					OrganizationID:  orgID,
				}

				if err := s.inventoryRepo.CreateAuditLog(ctx, auditLog); err != nil {
					// Log error but don't fail the operation
					fmt.Printf("Warning: failed to create audit log for lot %s: %v\n", lot.ID, err)
				}
			}
		}
	}

	return nil
}

// SellInventory converts reserved inventory to sold with audit trail
func (s *inventoryService) SellInventory(ctx context.Context, catalogItemID string, quantity decimal.Decimal, userID, orgID string) error {
	// Validate input
	if quantity.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("quantity must be greater than zero")
	}

	// Validate catalog item exists and belongs to organization
	catalogItem := &catalogModels.CatalogItem{}
	_, err := s.catalogRepo.GetByID(ctx, catalogItemID, catalogItem)
	if err != nil {
		return fmt.Errorf("catalog item not found: %w", err)
	}

	if catalogItem.OrganizationID != orgID {
		return fmt.Errorf("catalog item does not belong to organization")
	}

	if catalogItem.ItemType != catalogModels.CatalogItemTypeProduct {
		return fmt.Errorf("inventory operations are only available for products")
	}

	// Check if there's enough reserved quantity to sell
	lotsWithReserved, _, err := s.inventoryRepo.ListByCatalogItem(ctx, catalogItemID, 0, 1000)
	if err != nil {
		return fmt.Errorf("failed to get reserved lots: %w", err)
	}

	// Filter to only lots with reserved quantity
	var filteredLots []*catalogModels.InventoryLot
	for _, lot := range lotsWithReserved {
		if lot.ReservedQty.GreaterThan(decimal.Zero) {
			filteredLots = append(filteredLots, lot)
		}
	}
	lotsWithReserved = filteredLots

	totalReserved := decimal.Zero
	originalReserved := make(map[string]decimal.Decimal)
	originalSold := make(map[string]decimal.Decimal)
	for _, lot := range lotsWithReserved {
		totalReserved = totalReserved.Add(lot.ReservedQty)
		originalReserved[lot.ID] = lot.ReservedQty
		originalSold[lot.ID] = lot.Quantity
	}

	if totalReserved.LessThan(quantity) {
		return fmt.Errorf("insufficient reserved inventory: requested %s, reserved %s", quantity.String(), totalReserved.String())
	}

	// Sell inventory
	if err := s.inventoryRepo.SellQuantity(ctx, catalogItemID, quantity); err != nil {
		return fmt.Errorf("failed to sell inventory: %w", err)
	}

	// Create audit logs for affected lots
	// Get updated lots to see what changed
	updatedLots, _, err := s.inventoryRepo.ListByCatalogItem(ctx, catalogItemID, 0, 1000)
	if err == nil {
		for _, lot := range updatedLots {
			originalReservedQty := originalReserved[lot.ID]
			originalSoldQty := originalSold[lot.ID]

			if lot.Quantity.GreaterThan(originalSoldQty) {
				quantitySold := lot.Quantity.Sub(originalSoldQty)

				auditLog := &inventoryRepo.InventoryAuditLog{
					BaseModel:       *base.NewBaseModel("AUDIT", "small"),
					LotID:           lot.ID,
					Operation:       "sell",
					QuantityBefore:  originalReservedQty,
					QuantityAfter:   lot.ReservedQty,
					QuantityChanged: quantitySold,
					Reason:          fmt.Sprintf("Inventory sale for catalog item %s", catalogItemID),
					UserID:          userID,
					OrganizationID:  orgID,
				}

				if err := s.inventoryRepo.CreateAuditLog(ctx, auditLog); err != nil {
					// Log error but don't fail the operation
					fmt.Printf("Warning: failed to create audit log for lot %s: %v\n", lot.ID, err)
				}
			}
		}
	}

	return nil
}

// AdjustInventory adjusts inventory quantity with audit trail
func (s *inventoryService) AdjustInventory(ctx context.Context, lotID string, adjustment decimal.Decimal, reason, userID, orgID string) (*catalogModels.InventoryLot, error) {
	// Get existing lot
	lot, err := s.inventoryRepo.GetByID(ctx, lotID)
	if err != nil {
		return nil, err
	}

	// Check organization access
	if lot.OrganizationID != orgID {
		return nil, fmt.Errorf("access denied: lot does not belong to organization")
	}

	// Store original quantities for audit
	originalTotal := lot.AvailableQty.Add(lot.ReservedQty).Add(lot.Quantity)

	// Apply adjustment to available quantity
	newAvailable := lot.AvailableQty.Add(adjustment)
	if newAvailable.LessThan(decimal.Zero) {
		return nil, fmt.Errorf("adjustment would result in negative available quantity")
	}

	lot.AvailableQty = newAvailable
	lot.Quantity = lot.Quantity.Add(adjustment)
	lot.UpdatedBy = userID

	// Update status if necessary
	if lot.AvailableQty.IsZero() && lot.ReservedQty.IsZero() {
		lot.Status = catalogModels.LotStatusSold
	} else if lot.AvailableQty.GreaterThan(decimal.Zero) {
		lot.Status = catalogModels.LotStatusActive
	}

	// Save changes
	if err := s.inventoryRepo.Update(ctx, lot); err != nil {
		return nil, fmt.Errorf("failed to adjust inventory: %w", err)
	}

	// Create audit log
	newTotal := lot.AvailableQty.Add(lot.ReservedQty).Add(lot.Quantity)
	auditLog := &inventoryRepo.InventoryAuditLog{
		BaseModel:       *base.NewBaseModel("AUDIT", "small"),
		LotID:           lot.ID,
		Operation:       "adjust",
		QuantityBefore:  originalTotal,
		QuantityAfter:   newTotal,
		QuantityChanged: adjustment,
		Reason:          reason,
		UserID:          userID,
		OrganizationID:  orgID,
	}

	if err := s.inventoryRepo.CreateAuditLog(ctx, auditLog); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to create audit log: %v\n", err)
	}

	return lot, nil
}

// GetAvailableQuantity gets available quantity for a catalog item
func (s *inventoryService) GetAvailableQuantity(ctx context.Context, catalogItemID, userID, orgID string) (decimal.Decimal, error) {
	// Validate catalog item exists and belongs to organization
	catalogItem := &catalogModels.CatalogItem{}
	_, err := s.catalogRepo.GetByID(ctx, catalogItemID, catalogItem)
	if err != nil {
		return decimal.Zero, fmt.Errorf("catalog item not found: %w", err)
	}

	if catalogItem.OrganizationID != orgID {
		return decimal.Zero, fmt.Errorf("catalog item does not belong to organization")
	}

	if catalogItem.ItemType != catalogModels.CatalogItemTypeProduct {
		return decimal.Zero, fmt.Errorf("inventory is only available for products")
	}

	return s.inventoryRepo.GetAvailableQuantity(ctx, catalogItemID)
}

// GetTotalQuantity gets total quantity for a catalog item
func (s *inventoryService) GetTotalQuantity(ctx context.Context, catalogItemID, userID, orgID string) (decimal.Decimal, error) {
	// Validate catalog item exists and belongs to organization
	catalogItem := &catalogModels.CatalogItem{}
	_, err := s.catalogRepo.GetByID(ctx, catalogItemID, catalogItem)
	if err != nil {
		return decimal.Zero, fmt.Errorf("catalog item not found: %w", err)
	}

	if catalogItem.OrganizationID != orgID {
		return decimal.Zero, fmt.Errorf("catalog item does not belong to organization")
	}

	if catalogItem.ItemType != catalogModels.CatalogItemTypeProduct {
		return decimal.Zero, fmt.Errorf("inventory is only available for products")
	}

	return s.inventoryRepo.GetTotalQuantity(ctx, catalogItemID)
}

// CheckAvailability checks if required quantity is available
func (s *inventoryService) CheckAvailability(ctx context.Context, catalogItemID string, requiredQuantity decimal.Decimal, userID, orgID string) (bool, decimal.Decimal, error) {
	available, err := s.GetAvailableQuantity(ctx, catalogItemID, userID, orgID)
	if err != nil {
		return false, decimal.Zero, err
	}

	return available.GreaterThanOrEqual(requiredQuantity), available, nil
}

// ProcessExpiringLots processes lots that are expiring and marks them as expired
func (s *inventoryService) ProcessExpiringLots(ctx context.Context, orgID string) ([]*catalogModels.InventoryLot, error) {
	// Get lots expiring today or earlier
	expiringLots, err := s.inventoryRepo.GetExpiringLots(ctx, orgID, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get expiring lots: %w", err)
	}

	var processedLots []*catalogModels.InventoryLot

	for _, lot := range expiringLots {
		// Mark as expired
		if err := s.inventoryRepo.MarkExpired(ctx, lot.ID); err != nil {
			fmt.Printf("Warning: failed to mark lot %s as expired: %v\n", lot.ID, err)
			continue
		}

		// Create audit log
		auditLog := &inventoryRepo.InventoryAuditLog{
			BaseModel:       *base.NewBaseModel("AUDIT", "small"),
			LotID:           lot.ID,
			Operation:       "expire",
			QuantityBefore:  lot.AvailableQty.Add(lot.ReservedQty),
			QuantityAfter:   lot.AvailableQty.Add(lot.ReservedQty),
			QuantityChanged: decimal.Zero,
			Reason:          "Automatic expiration processing",
			UserID:          "system",
			OrganizationID:  orgID,
		}

		if err := s.inventoryRepo.CreateAuditLog(ctx, auditLog); err != nil {
			fmt.Printf("Warning: failed to create audit log for expired lot %s: %v\n", lot.ID, err)
		}

		lot.Status = catalogModels.LotStatusExpired
		processedLots = append(processedLots, lot)
	}

	return processedLots, nil
}

// MarkLotExpired manually marks a lot as expired
func (s *inventoryService) MarkLotExpired(ctx context.Context, lotID, userID, orgID string) error {
	// Get existing lot
	lot, err := s.inventoryRepo.GetByID(ctx, lotID)
	if err != nil {
		return err
	}

	// Check organization access
	if lot.OrganizationID != orgID {
		return fmt.Errorf("access denied: lot does not belong to organization")
	}

	// Check if lot can be expired
	if lot.Status == catalogModels.LotStatusExpired {
		return fmt.Errorf("lot is already expired")
	}

	if lot.Status == catalogModels.LotStatusSold {
		return fmt.Errorf("cannot expire a sold lot")
	}

	// Mark as expired
	if err := s.inventoryRepo.MarkExpired(ctx, lotID); err != nil {
		return fmt.Errorf("failed to mark lot as expired: %w", err)
	}

	// Create audit log
	auditLog := &inventoryRepo.InventoryAuditLog{
		BaseModel:       *base.NewBaseModel("AUDIT", "small"),
		LotID:           lot.ID,
		Operation:       "expire",
		QuantityBefore:  lot.AvailableQty.Add(lot.ReservedQty),
		QuantityAfter:   lot.AvailableQty.Add(lot.ReservedQty),
		QuantityChanged: decimal.Zero,
		Reason:          "Manual expiration",
		UserID:          userID,
		OrganizationID:  orgID,
	}

	if err := s.inventoryRepo.CreateAuditLog(ctx, auditLog); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to create audit log: %v\n", err)
	}

	return nil
}

// GetAuditTrail gets audit trail for a lot
func (s *inventoryService) GetAuditTrail(ctx context.Context, lotID string, offset, limit int, userID, orgID string) (*AuditTrailResponse, error) {
	// Get lot to check organization access
	lot, err := s.inventoryRepo.GetByID(ctx, lotID)
	if err != nil {
		return nil, err
	}

	// Check organization access
	if lot.OrganizationID != orgID {
		return nil, fmt.Errorf("access denied: lot does not belong to organization")
	}

	// Set default pagination
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Get audit logs
	logs, total, err := s.inventoryRepo.GetAuditLogs(ctx, lotID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit trail: %w", err)
	}

	return &AuditTrailResponse{
		Logs:   logs,
		Total:  total,
		Offset: offset,
		Limit:  limit,
	}, nil
}
