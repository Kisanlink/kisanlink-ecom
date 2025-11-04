package inventory

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// AlertType represents the type of inventory alert
type AlertType string

const (
	AlertTypeLowStock     AlertType = "LOW_STOCK"
	AlertTypeExpiringSoon AlertType = "EXPIRING_SOON"
	AlertTypeExpired      AlertType = "EXPIRED"
)

// AlertStatus represents the status of an alert
type AlertStatus string

const (
	AlertStatusActive       AlertStatus = "ACTIVE"
	AlertStatusAcknowledged AlertStatus = "ACKNOWLEDGED"
	AlertStatusResolved     AlertStatus = "RESOLVED"
)

// InventoryAlert represents an alert for inventory management
type InventoryAlert struct {
	base.BaseModel

	// Alert details
	InventoryLotID  string      `json:"inventory_lot_id" gorm:"type:varchar(255);not null;index:idx_alerts_lot"`
	FPOOrgID        string      `json:"fpo_org_id" gorm:"type:varchar(255);not null;index:idx_alerts_org"`
	AlertType       AlertType   `json:"alert_type" gorm:"type:varchar(50);not null;index:idx_alerts_type;check:alert_type IN ('LOW_STOCK', 'EXPIRING_SOON', 'EXPIRED')"`
	Threshold       int         `json:"threshold" gorm:"type:int;default:0"` // For low stock
	CurrentQuantity int         `json:"current_quantity" gorm:"type:int;not null"`
	Message         string      `json:"message" gorm:"type:text;not null"`
	Status          AlertStatus `json:"status" gorm:"type:varchar(50);not null;default:'ACTIVE';index:idx_alerts_status;check:status IN ('ACTIVE', 'ACKNOWLEDGED', 'RESOLVED')"`

	// Timestamps
	CreatedAt      time.Time  `json:"created_at" gorm:"type:timestamp;not null;index:idx_alerts_created"`
	AcknowledgedAt *time.Time `json:"acknowledged_at" gorm:"type:timestamp"`
	ResolvedAt     *time.Time `json:"resolved_at" gorm:"type:timestamp"`
	AcknowledgedBy string     `json:"acknowledged_by" gorm:"type:varchar(255)"`

	// Soft delete support
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index:idx_alerts_deleted"`
}

// TableName returns the table name for GORM
func (InventoryAlert) TableName() string {
	return "inventory_alerts"
}

// NewInventoryAlert creates a new inventory alert
func NewInventoryAlert(lotID, fpoOrgID string, alertType AlertType, currentQty int, message string) *InventoryAlert {
	return &InventoryAlert{
		BaseModel:       *base.NewBaseModel("ALERT", "small"),
		InventoryLotID:  lotID,
		FPOOrgID:        fpoOrgID,
		AlertType:       alertType,
		CurrentQuantity: currentQty,
		Message:         message,
		Status:          AlertStatusActive,
		CreatedAt:       time.Now(),
	}
}

// Acknowledge marks the alert as acknowledged
func (a *InventoryAlert) Acknowledge(userID string) {
	now := time.Now()
	a.Status = AlertStatusAcknowledged
	a.AcknowledgedAt = &now
	a.AcknowledgedBy = userID
}

// Resolve marks the alert as resolved
func (a *InventoryAlert) Resolve() {
	now := time.Now()
	a.Status = AlertStatusResolved
	a.ResolvedAt = &now
}

// IsActive checks if the alert is still active
func (a *InventoryAlert) IsActive() bool {
	return a.Status == AlertStatusActive
}

// AlertConfig represents configuration for inventory alerts per FPO
type AlertConfig struct {
	base.BaseModel

	// Configuration
	FPOOrgID          string   `json:"fpo_org_id" gorm:"type:varchar(255);not null;uniqueIndex:idx_alert_config_org"`
	LowStockThreshold int      `json:"low_stock_threshold" gorm:"type:int;not null;default:100"`
	ExpiryWarningDays int      `json:"expiry_warning_days" gorm:"type:int;not null;default:30"`
	AlertChannels     []string `json:"alert_channels" gorm:"type:jsonb;not null"` // EMAIL, SMS, WEBHOOK

	// Timestamps
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp;not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"type:timestamp;not null"`
	CreatedBy string    `json:"created_by" gorm:"type:varchar(255);not null"`
	UpdatedBy string    `json:"updated_by" gorm:"type:varchar(255);not null"`

	// Soft delete support
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index:idx_alert_config_deleted"`
}

// TableName returns the table name for GORM
func (AlertConfig) TableName() string {
	return "alert_configs"
}

// NewAlertConfig creates a new alert configuration with default values
func NewAlertConfig(fpoOrgID, userID string) *AlertConfig {
	return &AlertConfig{
		BaseModel:         *base.NewBaseModel("ALRTCFG", "small"),
		FPOOrgID:          fpoOrgID,
		LowStockThreshold: 100,        // Default threshold
		ExpiryWarningDays: 30,         // Default 30 days
		AlertChannels:     []string{}, // Start with no channels
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		CreatedBy:         userID,
		UpdatedBy:         userID,
	}
}

// UpdateConfig updates the alert configuration
func (c *AlertConfig) UpdateConfig(lowStockThreshold, expiryWarningDays int, alertChannels []string, userID string) {
	c.LowStockThreshold = lowStockThreshold
	c.ExpiryWarningDays = expiryWarningDays
	c.AlertChannels = alertChannels
	c.UpdatedBy = userID
	c.UpdatedAt = time.Now()
}
