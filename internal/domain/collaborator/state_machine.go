// Package collaborator provides domain logic for collaborator management including
// state machine enforcement for status transitions with prerequisite validation.
package collaborator

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// StateMachine enforces valid status transitions
type StateMachine struct {
	db            *gorm.DB
	logger        *zap.Logger
	transitions   map[string]*Transition
	prerequisites map[string][]PrerequisiteFunc
}

// Transition represents a state transition
type Transition struct {
	From          Status
	To            Status
	Event         string
	RequiredRoles []string
}

// PrerequisiteFunc checks if a prerequisite is met
type PrerequisiteFunc func(ctx context.Context, db *gorm.DB, collaboratorID uint64) error

// NewStateMachine creates a new state machine
func NewStateMachine(db *gorm.DB, logger *zap.Logger) *StateMachine {
	sm := &StateMachine{
		db:            db,
		logger:        logger,
		transitions:   make(map[string]*Transition),
		prerequisites: make(map[string][]PrerequisiteFunc),
	}

	// Define valid transitions
	sm.registerTransitions()

	return sm
}

// registerTransitions defines all valid state transitions
func (sm *StateMachine) registerTransitions() {
	// PENDING → VERIFIED
	sm.addTransition(StatusPending, StatusVerified, "VERIFY", []string{"ADMIN", "MANAGER", "VERIFIER"})
	sm.addPrerequisite(StatusPending, StatusVerified, RequireDocumentsVerified)

	// PENDING → REJECTED
	sm.addTransition(StatusPending, StatusRejected, "REJECT", []string{"ADMIN", "MANAGER", "VERIFIER"})

	// VERIFIED → ACTIVE
	sm.addTransition(StatusVerified, StatusActive, "ACTIVATE", []string{"ADMIN", "MANAGER"})
	sm.addPrerequisite(StatusVerified, StatusActive, RequireManagerApproval)

	// ACTIVE → ON_HOLD
	sm.addTransition(StatusActive, StatusOnHold, "HOLD", []string{"ADMIN", "MANAGER", "RISK_MANAGER"})

	// ACTIVE → SUSPENDED
	sm.addTransition(StatusActive, StatusSuspended, "SUSPEND", []string{"ADMIN", "RISK_MANAGER"})
	sm.addPrerequisite(StatusActive, StatusSuspended, RequireNoActiveOrders)

	// SUSPENDED → ACTIVE
	sm.addTransition(StatusSuspended, StatusActive, "REINSTATE", []string{"ADMIN"})

	// ON_HOLD → ACTIVE
	sm.addTransition(StatusOnHold, StatusActive, "RELEASE", []string{"ADMIN", "MANAGER"})

	// SUSPENDED → DEACTIVATED
	sm.addTransition(StatusSuspended, StatusDeactivated, "DEACTIVATE", []string{"ADMIN"})
	sm.addPrerequisite(StatusSuspended, StatusDeactivated, RequireNoOutstandingBalance)

	// INACTIVE → ACTIVE
	sm.addTransition(StatusInactive, StatusActive, "REACTIVATE", []string{"ADMIN", "MANAGER"})
}

// addTransition adds a valid transition
func (sm *StateMachine) addTransition(from, to Status, event string, roles []string) {
	key := sm.transitionKey(from, to)
	sm.transitions[key] = &Transition{
		From:          from,
		To:            to,
		Event:         event,
		RequiredRoles: roles,
	}
}

// addPrerequisite adds a prerequisite check for a transition
func (sm *StateMachine) addPrerequisite(from, to Status, prereq PrerequisiteFunc) {
	key := sm.transitionKey(from, to)
	sm.prerequisites[key] = append(sm.prerequisites[key], prereq)
}

// CanTransition checks if a transition is valid
func (sm *StateMachine) CanTransition(_ context.Context, from, to Status, userRoles []string) error {
	key := sm.transitionKey(from, to)

	// Check if transition exists
	transition, exists := sm.transitions[key]
	if !exists {
		return fmt.Errorf("invalid transition from %s to %s", from, to)
	}

	// Check role authorization
	if !sm.hasRequiredRole(userRoles, transition.RequiredRoles) {
		return fmt.Errorf("unauthorized: requires role %v, user has %v",
			transition.RequiredRoles, userRoles)
	}

	return nil
}

// ExecuteTransition executes a state transition with all checks
func (sm *StateMachine) ExecuteTransition(
	ctx context.Context,
	collaboratorID uint64,
	from, to Status,
	userID string,
	userRoles []string,
	reason string,
) error {
	// Validate transition is allowed
	if err := sm.CanTransition(ctx, from, to, userRoles); err != nil {
		return err
	}

	// Check prerequisites
	key := sm.transitionKey(from, to)
	if prereqs, exists := sm.prerequisites[key]; exists {
		for _, prereq := range prereqs {
			if err := prereq(ctx, sm.db, collaboratorID); err != nil {
				return fmt.Errorf("prerequisite failed: %w", err)
			}
		}
	}

	// Execute transition in database transaction
	return sm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update status
		now := time.Now()
		updates := map[string]interface{}{
			"status":     string(to),
			"updated_at": now,
		}

		// Update specific fields based on transition
		switch to {
		case StatusVerified:
			updates["verified_at"] = &now
			updates["is_verified"] = true
		case StatusActive:
			updates["status_updated_at"] = &now
		case StatusSuspended, StatusDeactivated:
			updates["status_updated_at"] = &now
		}

		err := tx.Table("collaborators").
			Where("id = ? AND status = ?", collaboratorID, string(from)).
			Updates(updates).Error

		if err != nil {
			return fmt.Errorf("failed to update status: %w", err)
		}

		// Create audit log
		auditLog := map[string]interface{}{
			"entity_type": "collaborator",
			"entity_id":   collaboratorID,
			"action":      fmt.Sprintf("status_transition_%s", sm.transitions[key].Event),
			"old_value":   string(from),
			"new_value":   string(to),
			"user_id":     userID,
			"reason":      reason,
			"created_at":  now,
		}

		err = tx.Table("audit_logs").Create(auditLog).Error
		if err != nil {
			sm.logger.Warn("failed to create audit log", zap.Error(err))
			// Don't fail the transaction if audit log fails
		}

		sm.logger.Info("status transition completed",
			zap.Uint64("collaborator_id", collaboratorID),
			zap.String("from", string(from)),
			zap.String("to", string(to)),
			zap.String("user_id", userID))

		return nil
	})
}

// GetAvailableTransitions returns all valid transitions from a status
func (sm *StateMachine) GetAvailableTransitions(from Status, userRoles []string) []Transition {
	available := make([]Transition, 0)

	for _, transition := range sm.transitions {
		if transition.From == from && sm.hasRequiredRole(userRoles, transition.RequiredRoles) {
			available = append(available, *transition)
		}
	}

	return available
}

// transitionKey creates a unique key for a transition
func (sm *StateMachine) transitionKey(from, to Status) string {
	return fmt.Sprintf("%s->%s", from, to)
}

// hasRequiredRole checks if user has at least one required role
func (sm *StateMachine) hasRequiredRole(userRoles, requiredRoles []string) bool {
	for _, userRole := range userRoles {
		for _, requiredRole := range requiredRoles {
			if userRole == requiredRole {
				return true
			}
		}
	}
	return false
}

// Prerequisite functions

// RequireDocumentsVerified checks if all documents are verified
func RequireDocumentsVerified(_ context.Context, db *gorm.DB, collaboratorID uint64) error {
	// Simplified check - in production, check actual documents table
	var isVerified bool
	err := db.Table("collaborators").
		Where("id = ?", collaboratorID).
		Select("is_verified").
		Scan(&isVerified).Error

	if err != nil {
		return fmt.Errorf("failed to check verification status: %w", err)
	}

	// For now, allow transition - implement proper document checks in production
	return nil
}

// RequireManagerApproval checks if manager approval exists
func RequireManagerApproval(_ context.Context, _ *gorm.DB, _ uint64) error {
	// Simplified check - in production, check approvals table
	// For now, allow transition
	return nil
}

// RequireNoActiveOrders checks if collaborator has no active orders
func RequireNoActiveOrders(_ context.Context, db *gorm.DB, collaboratorID uint64) error {
	var count int64
	err := db.Table("orders").
		Where("collaborator_id = ? AND status IN (?)",
			collaboratorID,
			[]string{"PENDING", "PROCESSING", "SHIPPED"}).
		Count(&count).Error

	if err != nil {
		return fmt.Errorf("failed to check active orders: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("collaborator has %d active orders, cannot suspend", count)
	}

	return nil
}

// RequireNoOutstandingBalance checks if there's no outstanding balance
func RequireNoOutstandingBalance(_ context.Context, _ *gorm.DB, _ uint64) error {
	// Simplified check - in production, check payments/invoices table
	// For now, allow transition
	return nil
}
