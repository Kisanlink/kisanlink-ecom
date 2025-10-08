package common

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// AuditAction represents the type of action performed
type AuditAction string

const (
	AuditActionCreate    AuditAction = "CREATE"
	AuditActionUpdate    AuditAction = "UPDATE"
	AuditActionDelete    AuditAction = "DELETE"
	AuditActionRestore   AuditAction = "RESTORE"
	AuditActionPublish   AuditAction = "PUBLISH"
	AuditActionUnpublish AuditAction = "UNPUBLISH"
	AuditActionApprove   AuditAction = "APPROVE"
	AuditActionReject    AuditAction = "REJECT"
)

// AuditLog represents an append-only audit trail for all operations
type AuditLog struct {
	base.BaseModel

	// Tenant isolation
	OrganizationID string `json:"organization_id" gorm:"type:varchar(255);not null;index:idx_audit_logs_tenant"`

	// Action details
	Action     AuditAction `json:"action" gorm:"type:varchar(20);not null;index:idx_audit_logs_action;check:action IN ('CREATE', 'UPDATE', 'DELETE', 'RESTORE', 'PUBLISH', 'UNPUBLISH', 'APPROVE', 'REJECT')"`
	EntityType string      `json:"entity_type" gorm:"type:varchar(100);not null;index:idx_audit_logs_entity_type"`
	EntityID   string      `json:"entity_id" gorm:"type:varchar(255);not null;index:idx_audit_logs_entity_id"`
	EntityName string      `json:"entity_name" gorm:"type:varchar(255)"`

	// Actor information (who performed the action)
	ActorType  string `json:"actor_type" gorm:"type:varchar(50);not null;check:actor_type IN ('USER', 'SYSTEM', 'API', 'BATCH')"` // USER, SYSTEM, API, BATCH
	ActorID    string `json:"actor_id" gorm:"type:varchar(255);not null;index:idx_audit_logs_actor"`
	ActorName  string `json:"actor_name" gorm:"type:varchar(255)"`
	ActorEmail string `json:"actor_email" gorm:"type:varchar(255)"`

	// Request context
	RequestID string `json:"request_id" gorm:"type:varchar(255);index:idx_audit_logs_request"`
	SessionID string `json:"session_id" gorm:"type:varchar(255);index:idx_audit_logs_session"`
	IPAddress string `json:"ip_address" gorm:"type:varchar(45)"` // IPv6 compatible
	UserAgent string `json:"user_agent" gorm:"type:text"`
	Source    string `json:"source" gorm:"type:varchar(100)"` // web, mobile, api, admin

	// Change details
	FieldsChanged string `json:"fields_changed" gorm:"type:jsonb;index:idx_audit_logs_fields"` // Array of field names that changed
	OldValues     string `json:"old_values" gorm:"type:jsonb"`                                 // Previous values (for UPDATE/DELETE)
	NewValues     string `json:"new_values" gorm:"type:jsonb"`                                 // New values (for CREATE/UPDATE)
	ChangeReason  string `json:"change_reason" gorm:"type:text"`                               // Optional reason for the change

	// Metadata
	Metadata string `json:"metadata" gorm:"type:jsonb;index:idx_audit_logs_metadata"`                                                      // Additional context
	Tags     string `json:"tags" gorm:"type:jsonb;index:idx_audit_logs_tags"`                                                              // Searchable tags
	Severity string `json:"severity" gorm:"type:varchar(20);default:'INFO';check:severity IN ('LOW', 'INFO', 'WARN', 'HIGH', 'CRITICAL')"` // Audit severity level

	// Timestamp (immutable)
	Timestamp time.Time `json:"timestamp" gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;index:idx_audit_logs_timestamp"`

	// Compliance and retention
	RetentionDate *time.Time `json:"retention_date" gorm:"type:timestamp;index:idx_audit_logs_retention"` // When this record can be purged
	IsArchived    bool       `json:"is_archived" gorm:"not null;default:false;index:idx_audit_logs_archived"`

	// Note: No soft delete for audit logs - they are append-only and immutable
	// No UpdatedAt field - audit logs are never updated after creation
}

// TableName returns the table name for GORM
func (AuditLog) TableName() string {
	return "audit_logs"
}

// NewAuditLog creates a new audit log entry
func NewAuditLog(orgID string, action AuditAction, entityType, entityID string) *AuditLog {
	return &AuditLog{
		BaseModel:      *base.NewBaseModel("AUDIT", "large"),
		OrganizationID: orgID,
		Action:         action,
		EntityType:     entityType,
		EntityID:       entityID,
		Timestamp:      time.Now(),
		Severity:       "INFO",
		IsArchived:     false,
	}
}

// SetActor sets the actor information
func (a *AuditLog) SetActor(actorType, actorID, actorName, actorEmail string) {
	a.ActorType = actorType
	a.ActorID = actorID
	a.ActorName = actorName
	a.ActorEmail = actorEmail
}

// SetRequestContext sets the request context information
func (a *AuditLog) SetRequestContext(requestID, sessionID, ipAddress, userAgent, source string) {
	a.RequestID = requestID
	a.SessionID = sessionID
	a.IPAddress = ipAddress
	a.UserAgent = userAgent
	a.Source = source
}

// SetChangeDetails sets the change details
func (a *AuditLog) SetChangeDetails(fieldsChanged, oldValues, newValues, reason string) {
	a.FieldsChanged = fieldsChanged
	a.OldValues = oldValues
	a.NewValues = newValues
	a.ChangeReason = reason
}

// SetMetadata sets additional metadata
func (a *AuditLog) SetMetadata(metadata, tags string, severity string) {
	a.Metadata = metadata
	a.Tags = tags
	if severity != "" {
		a.Severity = severity
	}
}

// SetRetention sets the retention date for compliance
func (a *AuditLog) SetRetention(retentionDate time.Time) {
	a.RetentionDate = &retentionDate
}

// IsRetentionExpired checks if the audit log can be purged
func (a *AuditLog) IsRetentionExpired() bool {
	if a.RetentionDate == nil {
		return false
	}
	return time.Now().After(*a.RetentionDate)
}

// Archive marks the audit log as archived
func (a *AuditLog) Archive() {
	a.IsArchived = true
}

// GetEntityReference returns a string reference to the entity
func (a *AuditLog) GetEntityReference() string {
	return a.EntityType + ":" + a.EntityID
}

// GetActorReference returns a string reference to the actor
func (a *AuditLog) GetActorReference() string {
	return a.ActorType + ":" + a.ActorID
}
