package catalog

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/shopspring/decimal"
)

// InventoryAuditLog represents an audit trail entry for inventory operations
type InventoryAuditLog struct {
	base.BaseModel

	LotID           string          `json:"lot_id" gorm:"type:varchar(255);not null;index:idx_inventory_audit_lot"`
	Operation       string          `json:"operation" gorm:"type:varchar(50);not null;index:idx_inventory_audit_operation"` // reserve, release, sell, adjust, expire, create, update, delete
	QuantityBefore  decimal.Decimal `json:"quantity_before" gorm:"type:decimal(12,3);not null"`
	QuantityAfter   decimal.Decimal `json:"quantity_after" gorm:"type:decimal(12,3);not null"`
	QuantityChanged decimal.Decimal `json:"quantity_changed" gorm:"type:decimal(12,3);not null"`
	Reason          string          `json:"reason" gorm:"type:varchar(500)"`
	UserID          string          `json:"user_id" gorm:"type:varchar(255);index:idx_inventory_audit_user"`
	OrganizationID  string          `json:"organization_id" gorm:"type:varchar(255);not null;index:idx_inventory_audit_org"`
	Metadata        string          `json:"metadata" gorm:"type:jsonb"`
}

// TableName returns the table name for GORM
func (InventoryAuditLog) TableName() string {
	return "inventory_audit_logs"
}
