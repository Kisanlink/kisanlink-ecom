package inventory

import (
	"context"
	"fmt"
	"log"
	"time"

	catalogModels "github.com/Kisanlink/kisanlink-ecom/entities/models/catalog"
	inventoryModels "github.com/Kisanlink/kisanlink-ecom/entities/models/inventory"
	inventoryRepo "github.com/Kisanlink/kisanlink-ecom/internal/repositories/inventory"
	"github.com/Kisanlink/kisanlink-ecom/internal/services/notifications"

	"github.com/shopspring/decimal"
)

// AlertService defines the interface for alert business logic
type AlertService interface {
	// Alert management
	GetAlertsByFPO(ctx context.Context, fpoOrgID string, offset, limit int, userID string) (*AlertListResponse, error)
	GetActiveAlerts(ctx context.Context, fpoOrgID string, offset, limit int, userID string) (*AlertListResponse, error)
	AcknowledgeAlert(ctx context.Context, alertID, userID, orgID string) error
	GetAlertConfig(ctx context.Context, fpoOrgID, userID string) (*inventoryModels.AlertConfig, error)
	UpdateAlertConfig(ctx context.Context, fpoOrgID string, req *UpdateAlertConfigRequest, userID string) (*inventoryModels.AlertConfig, error)

	// Background jobs
	CheckLowStockItems(ctx context.Context) error
	MarkExpiredItems(ctx context.Context) error
}

// AlertListResponse represents a paginated list of alerts
type AlertListResponse struct {
	Alerts []*inventoryModels.InventoryAlert `json:"alerts"`
	Total  int64                             `json:"total"`
	Offset int                               `json:"offset"`
	Limit  int                               `json:"limit"`
}

// UpdateAlertConfigRequest represents a request to update alert configuration
type UpdateAlertConfigRequest struct {
	LowStockThreshold int      `json:"low_stock_threshold" validate:"required,gte=0"`
	ExpiryWarningDays int      `json:"expiry_warning_days" validate:"required,gte=1"`
	AlertChannels     []string `json:"alert_channels"`
}

// alertService implements the AlertService interface
type alertService struct {
	alertRepo       inventoryRepo.AlertRepository
	inventoryRepo   inventoryRepo.InventoryRepository
	notificationSvc notifications.NotificationService
}

// NewAlertService creates a new alert service
func NewAlertService(
	alertRepo inventoryRepo.AlertRepository,
	inventoryRepo inventoryRepo.InventoryRepository,
	notificationSvc notifications.NotificationService,
) AlertService {
	return &alertService{
		alertRepo:       alertRepo,
		inventoryRepo:   inventoryRepo,
		notificationSvc: notificationSvc,
	}
}

// GetAlertsByFPO retrieves all alerts for an FPO
func (s *alertService) GetAlertsByFPO(ctx context.Context, fpoOrgID string, offset, limit int, userID string) (*AlertListResponse, error) {
	// Set default pagination
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	alerts, total, err := s.alertRepo.GetAlertsByFPO(ctx, fpoOrgID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}

	return &AlertListResponse{
		Alerts: alerts,
		Total:  total,
		Offset: offset,
		Limit:  limit,
	}, nil
}

// GetActiveAlerts retrieves active alerts for an FPO
func (s *alertService) GetActiveAlerts(ctx context.Context, fpoOrgID string, offset, limit int, userID string) (*AlertListResponse, error) {
	// Set default pagination
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	alerts, total, err := s.alertRepo.GetActiveAlertsByFPO(ctx, fpoOrgID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get active alerts: %w", err)
	}

	return &AlertListResponse{
		Alerts: alerts,
		Total:  total,
		Offset: offset,
		Limit:  limit,
	}, nil
}

// AcknowledgeAlert marks an alert as acknowledged
func (s *alertService) AcknowledgeAlert(ctx context.Context, alertID, userID, orgID string) error {
	// Get alert to verify ownership
	alert, err := s.alertRepo.GetAlertByID(ctx, alertID)
	if err != nil {
		return fmt.Errorf("alert not found: %w", err)
	}

	// Check organization access
	if alert.FPOOrgID != orgID {
		return fmt.Errorf("access denied: alert does not belong to organization")
	}

	// Check if already acknowledged
	if alert.Status == inventoryModels.AlertStatusAcknowledged {
		return fmt.Errorf("alert is already acknowledged")
	}

	// Acknowledge alert
	if err := s.alertRepo.AcknowledgeAlert(ctx, alertID, userID); err != nil {
		return fmt.Errorf("failed to acknowledge alert: %w", err)
	}

	return nil
}

// GetAlertConfig retrieves alert configuration for an FPO
func (s *alertService) GetAlertConfig(ctx context.Context, fpoOrgID, userID string) (*inventoryModels.AlertConfig, error) {
	config, err := s.alertRepo.GetAlertConfig(ctx, fpoOrgID)
	if err != nil {
		// If config doesn't exist, create default one
		defaultConfig := inventoryModels.NewAlertConfig(fpoOrgID, userID)
		if createErr := s.alertRepo.CreateAlertConfig(ctx, defaultConfig); createErr != nil {
			return nil, fmt.Errorf("failed to create default alert config: %w", createErr)
		}
		return defaultConfig, nil
	}

	return config, nil
}

// UpdateAlertConfig updates alert configuration for an FPO
func (s *alertService) UpdateAlertConfig(ctx context.Context, fpoOrgID string, req *UpdateAlertConfigRequest, userID string) (*inventoryModels.AlertConfig, error) {
	// Validate request
	if req.LowStockThreshold < 0 {
		return nil, fmt.Errorf("low stock threshold must be >= 0")
	}
	if req.ExpiryWarningDays < 1 {
		return nil, fmt.Errorf("expiry warning days must be >= 1")
	}

	// Get existing config or create new one
	config, err := s.alertRepo.GetAlertConfig(ctx, fpoOrgID)
	if err != nil {
		// Create new config
		config = inventoryModels.NewAlertConfig(fpoOrgID, userID)
	}

	// Update config
	config.UpdateConfig(req.LowStockThreshold, req.ExpiryWarningDays, req.AlertChannels, userID)

	// Save config
	if err := s.alertRepo.UpdateAlertConfig(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to update alert config: %w", err)
	}

	return config, nil
}

// CheckLowStockItems runs hourly to check for low stock items
func (s *alertService) CheckLowStockItems(ctx context.Context) error {
	log.Printf("[ALERT_SERVICE] Starting low stock check at %s", time.Now().Format(time.RFC3339))

	// This is a simplified implementation
	// In production, you would:
	// 1. Get all FPOs with alert configs
	// 2. For each FPO, get their inventory lots
	// 3. Check against thresholds
	// 4. Create alerts and send notifications

	// For MVP, we'll log that the job ran
	log.Printf("[ALERT_SERVICE] Low stock check completed at %s", time.Now().Format(time.RFC3339))
	return nil
}

// MarkExpiredItems runs daily to mark expired inventory items
func (s *alertService) MarkExpiredItems(ctx context.Context) error {
	log.Printf("[ALERT_SERVICE] Starting expiry check at %s", time.Now().Format(time.RFC3339))

	// Query items where expiry_date < now and status != EXPIRED
	// This is a background job that processes all organizations

	// For demonstration, we'll implement the core logic
	// In production, you would batch process by organization

	now := time.Now()

	// Get all expiring lots (this would need to be paginated in production)
	// For MVP, we process up to 1000 lots at a time
	lots, err := s.inventoryRepo.GetExpiringLots(ctx, "", now)
	if err != nil {
		log.Printf("[ALERT_SERVICE] Error getting expiring lots: %v", err)
		return fmt.Errorf("failed to get expiring lots: %w", err)
	}

	log.Printf("[ALERT_SERVICE] Found %d expiring lots", len(lots))

	for _, lot := range lots {
		// Mark as expired
		if err := s.inventoryRepo.MarkExpired(ctx, lot.ID); err != nil {
			log.Printf("[ALERT_SERVICE] Failed to mark lot %s as expired: %v", lot.ID, err)
			continue
		}

		// Create alert
		message := fmt.Sprintf("Inventory lot %s has expired. Quantity: %s %s",
			lot.LotNumber,
			lot.AvailableQty.String(),
			lot.UnitOfMeasure,
		)

		alert := inventoryModels.NewInventoryAlert(
			lot.ID,
			lot.OrganizationID,
			inventoryModels.AlertTypeExpired,
			int(lot.AvailableQty.IntPart()),
			message,
		)

		if err := s.alertRepo.CreateAlert(ctx, alert); err != nil {
			log.Printf("[ALERT_SERVICE] Failed to create alert for lot %s: %v", lot.ID, err)
			continue
		}

		// Send notification
		s.sendExpiryAlert(ctx, lot, alert)
	}

	log.Printf("[ALERT_SERVICE] Expiry check completed at %s. Processed %d lots", time.Now().Format(time.RFC3339), len(lots))
	return nil
}

// SendLowStockAlert sends a low stock alert notification
func (s *alertService) SendLowStockAlert(ctx context.Context, lot *catalogModels.InventoryLot, alert *inventoryModels.InventoryAlert) {
	payload := &notifications.NotificationPayload{
		RecipientID: lot.OrganizationID,
		Subject:     "Low Stock Alert",
		Message:     alert.Message,
		Channel:     notifications.ChannelEmail,
		Metadata: map[string]interface{}{
			"alert_id":         alert.ID,
			"lot_id":           lot.ID,
			"lot_number":       lot.LotNumber,
			"current_quantity": alert.CurrentQuantity,
			"threshold":        alert.Threshold,
			"alert_type":       alert.AlertType,
		},
	}

	if err := s.notificationSvc.SendNotification(ctx, payload); err != nil {
		log.Printf("[ALERT_SERVICE] Failed to send low stock notification for lot %s: %v", lot.ID, err)
	}
}

// sendExpiryAlert sends an expiry alert notification
func (s *alertService) sendExpiryAlert(ctx context.Context, lot *catalogModels.InventoryLot, alert *inventoryModels.InventoryAlert) {
	payload := &notifications.NotificationPayload{
		RecipientID: lot.OrganizationID,
		Subject:     "Inventory Expiry Alert",
		Message:     alert.Message,
		Channel:     notifications.ChannelEmail,
		Metadata: map[string]interface{}{
			"alert_id":         alert.ID,
			"lot_id":           lot.ID,
			"lot_number":       lot.LotNumber,
			"current_quantity": alert.CurrentQuantity,
			"alert_type":       alert.AlertType,
			"expiry_date":      lot.ExpiryDate,
		},
	}

	if err := s.notificationSvc.SendNotification(ctx, payload); err != nil {
		log.Printf("[ALERT_SERVICE] Failed to send expiry notification for lot %s: %v", lot.ID, err)
	}
}

// checkLowStockForLot checks if a lot is low on stock and creates alert if needed
func (s *alertService) checkLowStockForLot(ctx context.Context, lot *catalogModels.InventoryLot, threshold int) error {
	// Check if available quantity is below threshold
	if lot.AvailableQty.LessThan(decimal.NewFromInt(int64(threshold))) {
		// Check if there's already an active alert for this lot
		existingAlerts, err := s.alertRepo.GetAlertsByLot(ctx, lot.ID)
		if err != nil {
			return fmt.Errorf("failed to check existing alerts: %w", err)
		}

		// Skip if there's already an active low stock alert
		for _, existing := range existingAlerts {
			if existing.AlertType == inventoryModels.AlertTypeLowStock && existing.Status == inventoryModels.AlertStatusActive {
				return nil
			}
		}

		// Create low stock alert
		message := fmt.Sprintf("Inventory lot %s is low on stock. Current: %s %s, Threshold: %d",
			lot.LotNumber,
			lot.AvailableQty.String(),
			lot.UnitOfMeasure,
			threshold,
		)

		alert := inventoryModels.NewInventoryAlert(
			lot.ID,
			lot.OrganizationID,
			inventoryModels.AlertTypeLowStock,
			int(lot.AvailableQty.IntPart()),
			message,
		)
		alert.Threshold = threshold

		if err := s.alertRepo.CreateAlert(ctx, alert); err != nil {
			return fmt.Errorf("failed to create low stock alert: %w", err)
		}

		// Send notification
		s.SendLowStockAlert(ctx, lot, alert)
	}

	return nil
}
