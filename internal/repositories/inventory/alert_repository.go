package inventory

import (
	"context"
	"fmt"

	inventoryModels "kisanlink-ecom/entities/models/inventory"
	"kisanlink-ecom/internal/repositories/common"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// AlertRepository defines the interface for alert data operations
type AlertRepository interface {
	// Alert CRUD operations
	CreateAlert(ctx context.Context, alert *inventoryModels.InventoryAlert) error
	GetAlertByID(ctx context.Context, id string) (*inventoryModels.InventoryAlert, error)
	GetAlertsByFPO(ctx context.Context, fpoOrgID string, offset, limit int) ([]*inventoryModels.InventoryAlert, int64, error)
	GetActiveAlertsByFPO(ctx context.Context, fpoOrgID string, offset, limit int) ([]*inventoryModels.InventoryAlert, int64, error)
	GetAlertsByLot(ctx context.Context, lotID string) ([]*inventoryModels.InventoryAlert, error)
	AcknowledgeAlert(ctx context.Context, alertID, userID string) error
	ResolveAlert(ctx context.Context, alertID string) error
	DeleteAlert(ctx context.Context, id string) error

	// Alert config operations
	GetAlertConfig(ctx context.Context, fpoOrgID string) (*inventoryModels.AlertConfig, error)
	CreateAlertConfig(ctx context.Context, config *inventoryModels.AlertConfig) error
	UpdateAlertConfig(ctx context.Context, config *inventoryModels.AlertConfig) error
	DeleteAlertConfig(ctx context.Context, fpoOrgID string) error
}

// alertRepository implements the AlertRepository interface
type alertRepository struct {
	*common.BaseRepository
	dbManager db.DBManager
}

// NewAlertRepository creates a new alert repository
func NewAlertRepository(dbManager db.DBManager) AlertRepository {
	return &alertRepository{
		BaseRepository: common.NewBaseRepository(dbManager),
		dbManager:      dbManager,
	}
}

// CreateAlert creates a new inventory alert
func (r *alertRepository) CreateAlert(ctx context.Context, alert *inventoryModels.InventoryAlert) error {
	if err := r.dbManager.Create(ctx, alert); err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}
	return nil
}

// GetAlertByID retrieves an alert by ID
func (r *alertRepository) GetAlertByID(ctx context.Context, id string) (*inventoryModels.InventoryAlert, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "id",
			Operator: base.OpEqual,
			Value:    id,
		},
	}

	// Apply query options
	filter = r.ApplyQueryOptions(ctx, filter)

	var alerts []*inventoryModels.InventoryAlert
	if err := r.dbManager.List(ctx, filter, &alerts); err != nil {
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	if len(alerts) == 0 {
		return nil, fmt.Errorf("alert not found")
	}

	return alerts[0], nil
}

// GetAlertsByFPO retrieves all alerts for an FPO with pagination
func (r *alertRepository) GetAlertsByFPO(ctx context.Context, fpoOrgID string, offset, limit int) ([]*inventoryModels.InventoryAlert, int64, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "fpo_org_id",
			Operator: base.OpEqual,
			Value:    fpoOrgID,
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

	// Apply query options
	filter = r.ApplyQueryOptions(ctx, filter)

	var alerts []*inventoryModels.InventoryAlert
	if err := r.dbManager.List(ctx, filter, &alerts); err != nil {
		return nil, 0, fmt.Errorf("failed to list alerts: %w", err)
	}

	// Get total count
	total, err := r.dbManager.Count(ctx, filter, &inventoryModels.InventoryAlert{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count alerts: %w", err)
	}

	return alerts, total, nil
}

// GetActiveAlertsByFPO retrieves active alerts for an FPO with pagination
func (r *alertRepository) GetActiveAlertsByFPO(ctx context.Context, fpoOrgID string, offset, limit int) ([]*inventoryModels.InventoryAlert, int64, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "fpo_org_id",
			Operator: base.OpEqual,
			Value:    fpoOrgID,
		},
		{
			Field:    "status",
			Operator: base.OpEqual,
			Value:    string(inventoryModels.AlertStatusActive),
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

	// Apply query options
	filter = r.ApplyQueryOptions(ctx, filter)

	var alerts []*inventoryModels.InventoryAlert
	if err := r.dbManager.List(ctx, filter, &alerts); err != nil {
		return nil, 0, fmt.Errorf("failed to list active alerts: %w", err)
	}

	// Get total count
	total, err := r.dbManager.Count(ctx, filter, &inventoryModels.InventoryAlert{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count active alerts: %w", err)
	}

	return alerts, total, nil
}

// GetAlertsByLot retrieves all alerts for a specific inventory lot
func (r *alertRepository) GetAlertsByLot(ctx context.Context, lotID string) ([]*inventoryModels.InventoryAlert, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "inventory_lot_id",
			Operator: base.OpEqual,
			Value:    lotID,
		},
	}
	filter.Sort = []base.SortField{
		{
			Field:     "created_at",
			Direction: "desc",
		},
	}

	// Apply query options
	filter = r.ApplyQueryOptions(ctx, filter)

	var alerts []*inventoryModels.InventoryAlert
	if err := r.dbManager.List(ctx, filter, &alerts); err != nil {
		return nil, fmt.Errorf("failed to list alerts for lot: %w", err)
	}

	return alerts, nil
}

// AcknowledgeAlert marks an alert as acknowledged
func (r *alertRepository) AcknowledgeAlert(ctx context.Context, alertID, userID string) error {
	// Get the alert first
	alert, err := r.GetAlertByID(ctx, alertID)
	if err != nil {
		return err
	}

	// Update status
	alert.Acknowledge(userID)

	// Save the updated alert
	if err := r.dbManager.Update(ctx, alert); err != nil {
		return fmt.Errorf("failed to acknowledge alert: %w", err)
	}

	return nil
}

// ResolveAlert marks an alert as resolved
func (r *alertRepository) ResolveAlert(ctx context.Context, alertID string) error {
	// Get the alert first
	alert, err := r.GetAlertByID(ctx, alertID)
	if err != nil {
		return err
	}

	// Update status
	alert.Resolve()

	// Save the updated alert
	if err := r.dbManager.Update(ctx, alert); err != nil {
		return fmt.Errorf("failed to resolve alert: %w", err)
	}

	return nil
}

// DeleteAlert soft deletes an alert
func (r *alertRepository) DeleteAlert(ctx context.Context, id string) error {
	var alert inventoryModels.InventoryAlert
	if err := r.dbManager.Delete(ctx, id, &alert); err != nil {
		return fmt.Errorf("failed to delete alert: %w", err)
	}
	return nil
}

// GetAlertConfig retrieves alert configuration for an FPO
func (r *alertRepository) GetAlertConfig(ctx context.Context, fpoOrgID string) (*inventoryModels.AlertConfig, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "fpo_org_id",
			Operator: base.OpEqual,
			Value:    fpoOrgID,
		},
	}

	// Apply query options
	filter = r.ApplyQueryOptions(ctx, filter)

	var configs []*inventoryModels.AlertConfig
	if err := r.dbManager.List(ctx, filter, &configs); err != nil {
		return nil, fmt.Errorf("failed to get alert config: %w", err)
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("alert config not found")
	}

	return configs[0], nil
}

// CreateAlertConfig creates a new alert configuration
func (r *alertRepository) CreateAlertConfig(ctx context.Context, config *inventoryModels.AlertConfig) error {
	if err := r.dbManager.Create(ctx, config); err != nil {
		return fmt.Errorf("failed to create alert config: %w", err)
	}
	return nil
}

// UpdateAlertConfig updates an alert configuration
func (r *alertRepository) UpdateAlertConfig(ctx context.Context, config *inventoryModels.AlertConfig) error {
	if err := r.dbManager.Update(ctx, config); err != nil {
		return fmt.Errorf("failed to update alert config: %w", err)
	}
	return nil
}

// DeleteAlertConfig soft deletes an alert configuration
func (r *alertRepository) DeleteAlertConfig(ctx context.Context, fpoOrgID string) error {
	config, err := r.GetAlertConfig(ctx, fpoOrgID)
	if err != nil {
		return err
	}

	var alertConfig inventoryModels.AlertConfig
	if err := r.dbManager.Delete(ctx, config.ID, &alertConfig); err != nil {
		return fmt.Errorf("failed to delete alert config: %w", err)
	}
	return nil
}
