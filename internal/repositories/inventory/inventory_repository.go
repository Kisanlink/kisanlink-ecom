package inventory

import (
	"context"
	"fmt"
	"time"

	catalogModels "kisanlink-ecom/entities/models/catalog"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"github.com/shopspring/decimal"
)

// InventoryRepository defines the interface for inventory data operations
type InventoryRepository interface {
	// CRUD operations
	Create(ctx context.Context, lot *catalogModels.InventoryLot) error
	GetByID(ctx context.Context, id string) (*catalogModels.InventoryLot, error)
	GetByLotNumber(ctx context.Context, orgID, lotNumber string) (*catalogModels.InventoryLot, error)
	Update(ctx context.Context, lot *catalogModels.InventoryLot) error
	Delete(ctx context.Context, id string) error

	// Listing and filtering
	ListByOrganization(ctx context.Context, orgID string, offset, limit int) ([]*catalogModels.InventoryLot, int64, error)
	ListByCatalogItem(ctx context.Context, catalogItemID string, offset, limit int) ([]*catalogModels.InventoryLot, int64, error)
	ListByStatus(ctx context.Context, orgID string, status catalogModels.InventoryStatus, offset, limit int) ([]*catalogModels.InventoryLot, int64, error)

	// Inventory operations
	GetAvailableQuantity(ctx context.Context, catalogItemID string) (decimal.Decimal, error)
	GetTotalQuantity(ctx context.Context, catalogItemID string) (decimal.Decimal, error)
	GetInventoryLevel(ctx context.Context, itemID string) (decimal.Decimal, error)
	ReserveQuantity(ctx context.Context, catalogItemID string, quantity decimal.Decimal) ([]*catalogModels.InventoryLot, error)
	ReleaseQuantity(ctx context.Context, catalogItemID string, quantity decimal.Decimal) error
	SellQuantity(ctx context.Context, catalogItemID string, quantity decimal.Decimal) error
	ReserveInventory(ctx context.Context, itemID string, quantity decimal.Decimal) error
	ReleaseInventory(ctx context.Context, itemID string, quantity decimal.Decimal) error

	// Expiration handling
	GetExpiringLots(ctx context.Context, orgID string, beforeDate time.Time) ([]*catalogModels.InventoryLot, error)
	MarkExpired(ctx context.Context, lotID string) error

	// Audit operations
	CreateAuditLog(ctx context.Context, log *InventoryAuditLog) error
	GetAuditLogs(ctx context.Context, lotID string, offset, limit int) ([]*InventoryAuditLog, int64, error)
}

// InventoryAuditLog represents an audit trail entry for inventory operations
type InventoryAuditLog struct {
	base.BaseModel

	LotID           string          `json:"lot_id" gorm:"type:varchar(255);not null;index"`
	Operation       string          `json:"operation" gorm:"type:varchar(50);not null"` // reserve, release, sell, adjust, expire
	QuantityBefore  decimal.Decimal `json:"quantity_before" gorm:"type:decimal(12,3);not null"`
	QuantityAfter   decimal.Decimal `json:"quantity_after" gorm:"type:decimal(12,3);not null"`
	QuantityChanged decimal.Decimal `json:"quantity_changed" gorm:"type:decimal(12,3);not null"`
	Reason          string          `json:"reason" gorm:"type:varchar(500)"`
	UserID          string          `json:"user_id" gorm:"type:varchar(255)"`
	OrganizationID  string          `json:"organization_id" gorm:"type:varchar(255);not null;index"`
	Metadata        string          `json:"metadata" gorm:"type:jsonb"`
}

// TableName returns the table name for GORM
func (InventoryAuditLog) TableName() string {
	return "inventory_audit_logs"
}

// inventoryRepository implements the InventoryRepository interface
type inventoryRepository struct {
	dbManager db.DBManager
}

// NewInventoryRepository creates a new inventory repository
func NewInventoryRepository(dbManager db.DBManager) InventoryRepository {
	return &inventoryRepository{
		dbManager: dbManager,
	}
}

// Create creates a new inventory lot
func (r *inventoryRepository) Create(ctx context.Context, lot *catalogModels.InventoryLot) error {
	if err := r.dbManager.Create(ctx, lot); err != nil {
		return fmt.Errorf("failed to create inventory lot: %w", err)
	}
	return nil
}

// GetByID retrieves an inventory lot by ID
func (r *inventoryRepository) GetByID(ctx context.Context, id string) (*catalogModels.InventoryLot, error) {
	var lot catalogModels.InventoryLot
	if err := r.dbManager.GetByID(ctx, id, &lot); err != nil {
		return nil, fmt.Errorf("failed to get inventory lot: %w", err)
	}
	return &lot, nil
}

// GetByLotNumber retrieves an inventory lot by lot number and organization
func (r *inventoryRepository) GetByLotNumber(ctx context.Context, orgID, lotNumber string) (*catalogModels.InventoryLot, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "organization_id",
			Operator: base.OpEqual,
			Value:    orgID,
		},
		{
			Field:    "lot_number",
			Operator: base.OpEqual,
			Value:    lotNumber,
		},
	}

	var lots []*catalogModels.InventoryLot
	if err := r.dbManager.List(ctx, filter, &lots); err != nil {
		return nil, fmt.Errorf("failed to get inventory lot: %w", err)
	}

	if len(lots) == 0 {
		return nil, fmt.Errorf("inventory lot not found")
	}

	return lots[0], nil
}

// Update updates an inventory lot
func (r *inventoryRepository) Update(ctx context.Context, lot *catalogModels.InventoryLot) error {
	if err := r.dbManager.Update(ctx, lot); err != nil {
		return fmt.Errorf("failed to update inventory lot: %w", err)
	}
	return nil
}

// Delete soft deletes an inventory lot
func (r *inventoryRepository) Delete(ctx context.Context, id string) error {
	var lot catalogModels.InventoryLot
	if err := r.dbManager.Delete(ctx, id, &lot); err != nil {
		return fmt.Errorf("failed to delete inventory lot: %w", err)
	}
	return nil
}

// ListByOrganization lists inventory lots by organization with pagination
func (r *inventoryRepository) ListByOrganization(ctx context.Context, orgID string, offset, limit int) ([]*catalogModels.InventoryLot, int64, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "organization_id",
			Operator: base.OpEqual,
			Value:    orgID,
		},
	}
	filter.Offset = offset
	filter.Limit = limit
	filter.Sort = []base.SortField{
		{
			Field:     "created_at",
			Direction: "desc",
		},
	}

	var lots []*catalogModels.InventoryLot
	if err := r.dbManager.List(ctx, filter, &lots); err != nil {
		return nil, 0, fmt.Errorf("failed to list inventory lots: %w", err)
	}

	// Get total count
	total, err := r.dbManager.Count(ctx, filter, &catalogModels.InventoryLot{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count inventory lots: %w", err)
	}

	return lots, total, nil
}

// ListByCatalogItem lists inventory lots by catalog item with pagination
func (r *inventoryRepository) ListByCatalogItem(ctx context.Context, catalogItemID string, offset, limit int) ([]*catalogModels.InventoryLot, int64, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "catalog_item_id",
			Operator: base.OpEqual,
			Value:    catalogItemID,
		},
	}
	filter.Offset = offset
	filter.Limit = limit
	filter.Sort = []base.SortField{
		{
			Field:     "created_at",
			Direction: "desc",
		},
	}

	var lots []*catalogModels.InventoryLot
	if err := r.dbManager.List(ctx, filter, &lots); err != nil {
		return nil, 0, fmt.Errorf("failed to list inventory lots: %w", err)
	}

	// Get total count
	total, err := r.dbManager.Count(ctx, filter, &catalogModels.InventoryLot{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count inventory lots: %w", err)
	}

	return lots, total, nil
}

// ListByStatus lists inventory lots by status with pagination
func (r *inventoryRepository) ListByStatus(ctx context.Context, orgID string, status catalogModels.InventoryStatus, offset, limit int) ([]*catalogModels.InventoryLot, int64, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "organization_id",
			Operator: base.OpEqual,
			Value:    orgID,
		},
		{
			Field:    "status",
			Operator: base.OpEqual,
			Value:    string(status),
		},
	}
	filter.Offset = offset
	filter.Limit = limit
	filter.Sort = []base.SortField{
		{
			Field:     "created_at",
			Direction: "desc",
		},
	}

	var lots []*catalogModels.InventoryLot
	if err := r.dbManager.List(ctx, filter, &lots); err != nil {
		return nil, 0, fmt.Errorf("failed to list inventory lots: %w", err)
	}

	// Get total count
	total, err := r.dbManager.Count(ctx, filter, &catalogModels.InventoryLot{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count inventory lots: %w", err)
	}

	return lots, total, nil
}

// GetAvailableQuantity gets the total available quantity for a catalog item
func (r *inventoryRepository) GetAvailableQuantity(ctx context.Context, catalogItemID string) (decimal.Decimal, error) {
	// Get all lots for this catalog item with available status
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "catalog_item_id",
			Operator: base.OpEqual,
			Value:    catalogItemID,
		},
	}
	filter.Group.Groups = []base.FilterGroup{
		{
			Logic: base.LogicOr,
			Conditions: []base.FilterCondition{
				{
					Field:    "status",
					Operator: base.OpEqual,
					Value:    string(catalogModels.InventoryStatusAvailable),
				},
				{
					Field:    "status",
					Operator: base.OpEqual,
					Value:    string(catalogModels.InventoryStatusReserved),
				},
			},
		},
	}

	var lots []*catalogModels.InventoryLot
	if err := r.dbManager.List(ctx, filter, &lots); err != nil {
		return decimal.Zero, fmt.Errorf("failed to get inventory lots: %w", err)
	}

	// Sum up available quantities
	total := decimal.Zero
	for _, lot := range lots {
		total = total.Add(lot.AvailableQuantity)
	}

	return total, nil
}

// GetTotalQuantity gets the total quantity (available + reserved) for a catalog item
func (r *inventoryRepository) GetTotalQuantity(ctx context.Context, catalogItemID string) (decimal.Decimal, error) {
	// Get all lots for this catalog item with available or reserved status
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "catalog_item_id",
			Operator: base.OpEqual,
			Value:    catalogItemID,
		},
	}
	filter.Group.Groups = []base.FilterGroup{
		{
			Logic: base.LogicOr,
			Conditions: []base.FilterCondition{
				{
					Field:    "status",
					Operator: base.OpEqual,
					Value:    string(catalogModels.InventoryStatusAvailable),
				},
				{
					Field:    "status",
					Operator: base.OpEqual,
					Value:    string(catalogModels.InventoryStatusReserved),
				},
			},
		},
	}

	var lots []*catalogModels.InventoryLot
	if err := r.dbManager.List(ctx, filter, &lots); err != nil {
		return decimal.Zero, fmt.Errorf("failed to get inventory lots: %w", err)
	}

	// Sum up available and reserved quantities
	total := decimal.Zero
	for _, lot := range lots {
		total = total.Add(lot.AvailableQuantity).Add(lot.ReservedQuantity)
	}

	return total, nil
}

// ReserveQuantity reserves the specified quantity from available lots
// TODO: This implementation needs proper transaction support from dbManager
func (r *inventoryRepository) ReserveQuantity(ctx context.Context, catalogItemID string, quantity decimal.Decimal) ([]*catalogModels.InventoryLot, error) {
	var affectedLots []*catalogModels.InventoryLot
	remainingToReserve := quantity

	// Get available lots ordered by expiry date (FIFO for perishables)
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "catalog_item_id",
			Operator: base.OpEqual,
			Value:    catalogItemID,
		},
		{
			Field:    "status",
			Operator: base.OpEqual,
			Value:    string(catalogModels.InventoryStatusAvailable),
		},
		{
			Field:    "available_quantity",
			Operator: base.OpGreaterThan,
			Value:    decimal.Zero,
		},
	}
	filter.Sort = []base.SortField{
		{
			Field:     "expiry_date",
			Direction: "asc",
		},
		{
			Field:     "created_at",
			Direction: "asc",
		},
	}

	var lots []*catalogModels.InventoryLot
	if err := r.dbManager.List(ctx, filter, &lots); err != nil {
		return nil, fmt.Errorf("failed to get available lots: %w", err)
	}

	// Reserve from lots
	for _, lot := range lots {
		if remainingToReserve.IsZero() {
			break
		}

		quantityToReserve := remainingToReserve
		if lot.AvailableQuantity.LessThan(quantityToReserve) {
			quantityToReserve = lot.AvailableQuantity
		}

		// Reserve quantity from this lot
		if err := lot.Reserve(quantityToReserve); err != nil {
			return nil, fmt.Errorf("failed to reserve from lot %s: %w", lot.ID, err)
		}

		// Update lot in database
		if err := r.dbManager.Update(ctx, lot); err != nil {
			return nil, fmt.Errorf("failed to update lot %s: %w", lot.ID, err)
		}

		affectedLots = append(affectedLots, lot)
		remainingToReserve = remainingToReserve.Sub(quantityToReserve)
	}

	// Check if we could reserve the full quantity
	if remainingToReserve.GreaterThan(decimal.Zero) {
		return nil, fmt.Errorf("insufficient inventory: could not reserve %s units", remainingToReserve.String())
	}

	return affectedLots, nil
}

// ReleaseQuantity releases the specified reserved quantity back to available
// TODO: This implementation needs proper transaction support from dbManager
func (r *inventoryRepository) ReleaseQuantity(ctx context.Context, catalogItemID string, quantity decimal.Decimal) error {
	remainingToRelease := quantity

	// Get lots with reserved quantity
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "catalog_item_id",
			Operator: base.OpEqual,
			Value:    catalogItemID,
		},
		{
			Field:    "reserved_quantity",
			Operator: base.OpGreaterThan,
			Value:    decimal.Zero,
		},
	}
	filter.Sort = []base.SortField{
		{
			Field:     "created_at",
			Direction: "desc", // Release from most recent reservations first
		},
	}

	var lots []*catalogModels.InventoryLot
	if err := r.dbManager.List(ctx, filter, &lots); err != nil {
		return fmt.Errorf("failed to get reserved lots: %w", err)
	}

	// Release from lots
	for _, lot := range lots {
		if remainingToRelease.IsZero() {
			break
		}

		quantityToRelease := remainingToRelease
		if lot.ReservedQuantity.LessThan(quantityToRelease) {
			quantityToRelease = lot.ReservedQuantity
		}

		// Release quantity from this lot
		if err := lot.Release(quantityToRelease); err != nil {
			return fmt.Errorf("failed to release from lot %s: %w", lot.ID, err)
		}

		// Update lot in database
		if err := r.dbManager.Update(ctx, lot); err != nil {
			return fmt.Errorf("failed to update lot %s: %w", lot.ID, err)
		}

		remainingToRelease = remainingToRelease.Sub(quantityToRelease)
	}

	// Check if we could release the full quantity
	if remainingToRelease.GreaterThan(decimal.Zero) {
		return fmt.Errorf("insufficient reserved inventory: could not release %s units", remainingToRelease.String())
	}

	return nil
}

// SellQuantity converts reserved quantity to sold
// TODO: This implementation needs proper transaction support from dbManager
func (r *inventoryRepository) SellQuantity(ctx context.Context, catalogItemID string, quantity decimal.Decimal) error {
	remainingToSell := quantity

	// Get lots with reserved quantity
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "catalog_item_id",
			Operator: base.OpEqual,
			Value:    catalogItemID,
		},
		{
			Field:    "reserved_quantity",
			Operator: base.OpGreaterThan,
			Value:    decimal.Zero,
		},
	}
	filter.Sort = []base.SortField{
		{
			Field:     "expiry_date",
			Direction: "asc", // Sell oldest/expiring first
		},
		{
			Field:     "created_at",
			Direction: "asc",
		},
	}

	var lots []*catalogModels.InventoryLot
	if err := r.dbManager.List(ctx, filter, &lots); err != nil {
		return fmt.Errorf("failed to get reserved lots: %w", err)
	}

	// Sell from lots
	for _, lot := range lots {
		if remainingToSell.IsZero() {
			break
		}

		quantityToSell := remainingToSell
		if lot.ReservedQuantity.LessThan(quantityToSell) {
			quantityToSell = lot.ReservedQuantity
		}

		// Sell quantity from this lot
		if err := lot.Sell(quantityToSell); err != nil {
			return fmt.Errorf("failed to sell from lot %s: %w", lot.ID, err)
		}

		// Update lot in database
		if err := r.dbManager.Update(ctx, lot); err != nil {
			return fmt.Errorf("failed to update lot %s: %w", lot.ID, err)
		}

		remainingToSell = remainingToSell.Sub(quantityToSell)
	}

	// Check if we could sell the full quantity
	if remainingToSell.GreaterThan(decimal.Zero) {
		return fmt.Errorf("insufficient reserved inventory: could not sell %s units", remainingToSell.String())
	}

	return nil
}

// GetExpiringLots gets lots that are expiring before the specified date
func (r *inventoryRepository) GetExpiringLots(ctx context.Context, orgID string, beforeDate time.Time) ([]*catalogModels.InventoryLot, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "organization_id",
			Operator: base.OpEqual,
			Value:    orgID,
		},
		{
			Field:    "expiry_date",
			Operator: base.OpLessThan,
			Value:    beforeDate.Add(24 * time.Hour), // Add one day to include the target date
		},
	}
	// Add NOT IN condition for status
	filter.Group.Groups = []base.FilterGroup{
		{
			Logic: base.LogicAnd,
			Conditions: []base.FilterCondition{
				{
					Field:    "status",
					Operator: base.OpNotEqual,
					Value:    string(catalogModels.InventoryStatusExpired),
				},
				{
					Field:    "status",
					Operator: base.OpNotEqual,
					Value:    string(catalogModels.InventoryStatusSold),
				},
			},
		},
	}
	filter.Sort = []base.SortField{
		{
			Field:     "expiry_date",
			Direction: "asc",
		},
	}

	var lots []*catalogModels.InventoryLot
	if err := r.dbManager.List(ctx, filter, &lots); err != nil {
		return nil, fmt.Errorf("failed to get expiring lots: %w", err)
	}

	return lots, nil
}

// MarkExpired marks a lot as expired
func (r *inventoryRepository) MarkExpired(ctx context.Context, lotID string) error {
	// Get the lot first
	var lot catalogModels.InventoryLot
	if err := r.dbManager.GetByID(ctx, lotID, &lot); err != nil {
		return fmt.Errorf("failed to get lot: %w", err)
	}

	// Update status
	lot.Status = catalogModels.InventoryStatusExpired

	// Save the updated lot
	if err := r.dbManager.Update(ctx, &lot); err != nil {
		return fmt.Errorf("failed to mark lot as expired: %w", err)
	}
	return nil
}

// CreateAuditLog creates an audit log entry
func (r *inventoryRepository) CreateAuditLog(ctx context.Context, log *InventoryAuditLog) error {
	if err := r.dbManager.Create(ctx, log); err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	return nil
}

// GetAuditLogs gets audit logs for a lot with pagination
func (r *inventoryRepository) GetAuditLogs(ctx context.Context, lotID string, offset, limit int) ([]*InventoryAuditLog, int64, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "lot_id",
			Operator: base.OpEqual,
			Value:    lotID,
		},
	}
	filter.Offset = offset
	filter.Limit = limit
	filter.Sort = []base.SortField{
		{
			Field:     "created_at",
			Direction: "desc",
		},
	}

	var logs []*InventoryAuditLog
	if err := r.dbManager.List(ctx, filter, &logs); err != nil {
		return nil, 0, fmt.Errorf("failed to get audit logs: %w", err)
	}

	// Get total count
	total, err := r.dbManager.Count(ctx, filter, &InventoryAuditLog{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	return logs, total, nil
}

// GetInventoryLevel gets the available quantity for a catalog item (alias for GetAvailableQuantity)
func (r *inventoryRepository) GetInventoryLevel(ctx context.Context, itemID string) (decimal.Decimal, error) {
	return r.GetAvailableQuantity(ctx, itemID)
}

// ReserveInventory reserves inventory for a specific item (wrapper around ReserveQuantity)
func (r *inventoryRepository) ReserveInventory(ctx context.Context, itemID string, quantity decimal.Decimal) error {
	_, err := r.ReserveQuantity(ctx, itemID, quantity)
	return err
}

// ReleaseInventory releases reserved inventory for a specific item (wrapper around ReleaseQuantity)
func (r *inventoryRepository) ReleaseInventory(ctx context.Context, itemID string, quantity decimal.Decimal) error {
	return r.ReleaseQuantity(ctx, itemID, quantity)
}
