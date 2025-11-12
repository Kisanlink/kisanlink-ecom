# P0-6: State Machine Architecture for Collaborator Status Transitions

**Version**: 1.0
**Date**: 2025-11-12
**Priority**: P0 CRITICAL
**Author**: SDE-3 Backend Architect

---

## Executive Summary

This document provides production-ready architecture for implementing a finite state machine (FSM) to enforce valid collaborator status transitions, preventing unauthorized status changes and ensuring business rule compliance.

---

## Problem Statement

### Current Issue
Status transitions can be bypassed through direct API calls, allowing:
- Direct transition from PENDING to ACTIVE (bypassing verification)
- Reactivation of SUSPENDED collaborators without approval
- Invalid state transitions breaking business logic
- Lack of audit trail for status changes

### Security Impact
```
Current Vulnerable Flow:
PENDING ──────┐
              ├──> ACTIVE (No verification!)
VERIFIED ─────┘

Required Secure Flow:
PENDING ──> VERIFIED ──> ACTIVE
   │           │           │
   └───────────┴───────────┴──> SUSPENDED
```

### Business Impact
- Compliance violations
- Unverified collaborators conducting transactions
- Financial risk from unvetted partners
- Audit failures

---

## Architecture Design

### State Machine Definition

```
┌─────────────────────────────────────────────────────────────┐
│                  Collaborator State Machine                 │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│    ┌─────────┐      ┌──────────┐      ┌────────┐          │
│    │ PENDING │─────>│ VERIFIED │─────>│ ACTIVE │          │
│    └─────────┘      └──────────┘      └────────┘          │
│         │                 │                 │               │
│         │                 │                 │               │
│         ▼                 ▼                 ▼               │
│    ┌─────────┐      ┌──────────┐      ┌──────────┐        │
│    │ REJECTED│      │ ON_HOLD  │      │ SUSPENDED│        │
│    └─────────┘      └──────────┘      └──────────┘        │
│                           │                 │               │
│                           └────────┬────────┘               │
│                                    ▼                        │
│                              ┌──────────┐                   │
│                              │ INACTIVE │                   │
│                              └──────────┘                   │
└─────────────────────────────────────────────────────────────┘
```

### Core Components

```go
// internal/domain/collaborator/state_machine.go

type CollaboratorStatus string

const (
    StatusPending   CollaboratorStatus = "PENDING"
    StatusVerified  CollaboratorStatus = "VERIFIED"
    StatusActive    CollaboratorStatus = "ACTIVE"
    StatusOnHold    CollaboratorStatus = "ON_HOLD"
    StatusSuspended CollaboratorStatus = "SUSPENDED"
    StatusInactive  CollaboratorStatus = "INACTIVE"
    StatusRejected  CollaboratorStatus = "REJECTED"
)

type StateTransition struct {
    From           CollaboratorStatus
    To             CollaboratorStatus
    Event          string
    RequiredRole   []string
    Prerequisites  []TransitionPrerequisite
    Actions        []TransitionAction
    Compensations  []CompensationAction
}

type TransitionPrerequisite func(ctx context.Context, collaborator *Collaborator) error
type TransitionAction func(ctx context.Context, collaborator *Collaborator) error
type CompensationAction func(ctx context.Context, collaborator *Collaborator) error

type StateMachine struct {
    transitions     map[string]*StateTransition
    logger         *zap.Logger
    metrics        *StateMetrics
    auditLogger    AuditLogger
    notifier       NotificationService
}
```

### State Transition Matrix

```go
// internal/domain/collaborator/transitions.go

func NewCollaboratorStateMachine() *StateMachine {
    sm := &StateMachine{
        transitions: make(map[string]*StateTransition),
    }

    // Define all valid transitions
    sm.RegisterTransition(&StateTransition{
        From:  StatusPending,
        To:    StatusVerified,
        Event: "VERIFY",
        RequiredRole: []string{"ADMIN", "MANAGER", "VERIFIER"},
        Prerequisites: []TransitionPrerequisite{
            RequireGSTVerification,
            RequirePANVerification,
            RequireBankVerification,
            RequireAddressVerification,
            RequireDocumentsUploaded,
        },
        Actions: []TransitionAction{
            UpdateVerificationTimestamp,
            NotifyCollaboratorVerified,
            UpdateSearchIndex,
        },
    })

    sm.RegisterTransition(&StateTransition{
        From:  StatusVerified,
        To:    StatusActive,
        Event: "ACTIVATE",
        RequiredRole: []string{"ADMIN", "MANAGER"},
        Prerequisites: []TransitionPrerequisite{
            RequireManagerApproval,
            RequireCreditLimitSet,
            RequireContractSigned,
        },
        Actions: []TransitionAction{
            EnableTransactions,
            SetupPaymentTerms,
            NotifyCollaboratorActivated,
            CreateWelcomeKit,
        },
    })

    sm.RegisterTransition(&StateTransition{
        From:  StatusActive,
        To:    StatusOnHold,
        Event: "HOLD",
        RequiredRole: []string{"ADMIN", "MANAGER", "RISK_MANAGER"},
        Prerequisites: []TransitionPrerequisite{
            RequireHoldReason,
        },
        Actions: []TransitionAction{
            DisableNewTransactions,
            NotifyCollaboratorOnHold,
            FlagForReview,
        },
        Compensations: []CompensationAction{
            ReenableTransactions,
            RemoveHoldFlag,
        },
    })

    sm.RegisterTransition(&StateTransition{
        From:  StatusActive,
        To:    StatusSuspended,
        Event: "SUSPEND",
        RequiredRole: []string{"ADMIN", "RISK_MANAGER"},
        Prerequisites: []TransitionPrerequisite{
            RequireNoActiveOrders,
            RequireNoPendingPayments,
            RequireSuspensionReason,
        },
        Actions: []TransitionAction{
            DisableAllTransactions,
            FreezeCredit,
            NotifyCollaboratorSuspended,
            NotifyAccountManager,
        },
    })

    sm.RegisterTransition(&StateTransition{
        From:  StatusSuspended,
        To:    StatusActive,
        Event: "REINSTATE",
        RequiredRole: []string{"ADMIN"},
        Prerequisites: []TransitionPrerequisite{
            RequireReinstatementApproval,
            RequireIssuesResolved,
            RequireReverification,
        },
        Actions: []TransitionAction{
            ReenableTransactions,
            RestoreCredit,
            NotifyCollaboratorReinstated,
            AuditReinstatement,
        },
    })

    sm.RegisterTransition(&StateTransition{
        From:  StatusPending,
        To:    StatusRejected,
        Event: "REJECT",
        RequiredRole: []string{"ADMIN", "MANAGER", "VERIFIER"},
        Prerequisites: []TransitionPrerequisite{
            RequireRejectionReason,
        },
        Actions: []TransitionAction{
            NotifyCollaboratorRejected,
            CleanupPendingData,
            LogRejectionReason,
        },
    })

    sm.RegisterTransition(&StateTransition{
        From:  StatusOnHold,
        To:    StatusActive,
        Event: "RELEASE_HOLD",
        RequiredRole: []string{"ADMIN", "MANAGER"},
        Prerequisites: []TransitionPrerequisite{
            RequireHoldIssuesResolved,
            RequireManagerApproval,
        },
        Actions: []TransitionAction{
            ReenableTransactions,
            NotifyHoldReleased,
            ClearReviewFlags,
        },
    })

    sm.RegisterTransition(&StateTransition{
        From:  StatusSuspended,
        To:    StatusInactive,
        Event: "DEACTIVATE",
        RequiredRole: []string{"ADMIN"},
        Prerequisites: []TransitionPrerequisite{
            RequireNoOutstandingBalance,
            RequireDataArchived,
            RequireDeactivationApproval,
        },
        Actions: []TransitionAction{
            ArchiveCollaboratorData,
            RevokeAllAccess,
            NotifyDeactivation,
            FinalAuditLog,
        },
    })

    return sm
}
```

### State Machine Implementation

```go
// internal/domain/collaborator/state_machine.go

func (sm *StateMachine) CanTransition(
    ctx context.Context,
    collaborator *Collaborator,
    toStatus CollaboratorStatus,
    user *User,
) error {
    // Build transition key
    transitionKey := fmt.Sprintf("%s->%s", collaborator.Status, toStatus)

    transition, exists := sm.transitions[transitionKey]
    if !exists {
        return fmt.Errorf("invalid transition from %s to %s",
            collaborator.Status, toStatus)
    }

    // Check role authorization
    if !sm.hasRequiredRole(user, transition.RequiredRole) {
        sm.metrics.RecordUnauthorizedAttempt(user.ID, transitionKey)
        return fmt.Errorf("unauthorized: requires role %v",
            transition.RequiredRole)
    }

    // Check prerequisites
    for _, prereq := range transition.Prerequisites {
        if err := prereq(ctx, collaborator); err != nil {
            sm.metrics.RecordPrerequisiteFailure(transitionKey, err)
            return fmt.Errorf("prerequisite failed: %w", err)
        }
    }

    return nil
}

func (sm *StateMachine) ExecuteTransition(
    ctx context.Context,
    collaborator *Collaborator,
    toStatus CollaboratorStatus,
    user *User,
    reason string,
) error {
    // Validate transition
    if err := sm.CanTransition(ctx, collaborator, toStatus, user); err != nil {
        return err
    }

    transitionKey := fmt.Sprintf("%s->%s", collaborator.Status, toStatus)
    transition := sm.transitions[transitionKey]

    // Start transaction
    tx := sm.db.BeginTx(ctx, nil)
    defer tx.Rollback()

    // Record old status
    oldStatus := collaborator.Status

    // Create audit entry
    auditEntry := &AuditEntry{
        EntityType:   "collaborator",
        EntityID:     collaborator.ID,
        Action:       fmt.Sprintf("status_transition_%s", transition.Event),
        OldValue:     string(oldStatus),
        NewValue:     string(toStatus),
        UserID:       user.ID,
        Reason:       reason,
        Timestamp:    time.Now(),
        IP:           getIPFromContext(ctx),
        UserAgent:    getUserAgentFromContext(ctx),
    }

    // Execute pre-transition actions
    for _, action := range transition.Actions {
        if err := action(ctx, collaborator); err != nil {
            sm.logger.Error("transition action failed",
                zap.String("transition", transitionKey),
                zap.Error(err))

            // Execute compensations
            for _, compensation := range transition.Compensations {
                if compErr := compensation(ctx, collaborator); compErr != nil {
                    sm.logger.Error("compensation failed",
                        zap.Error(compErr))
                }
            }

            return fmt.Errorf("transition action failed: %w", err)
        }
    }

    // Update status
    collaborator.Status = toStatus
    collaborator.StatusUpdatedAt = time.Now()
    collaborator.StatusUpdatedBy = user.ID

    // Save to database
    if err := tx.Save(collaborator).Error; err != nil {
        return fmt.Errorf("failed to save status: %w", err)
    }

    // Save audit entry
    if err := tx.Create(auditEntry).Error; err != nil {
        return fmt.Errorf("failed to save audit: %w", err)
    }

    // Commit transaction
    if err := tx.Commit().Error; err != nil {
        return fmt.Errorf("failed to commit: %w", err)
    }

    // Record metrics
    sm.metrics.RecordTransition(transitionKey, true)

    // Send notifications (async)
    go sm.notifier.NotifyStatusChange(collaborator, oldStatus, toStatus)

    sm.logger.Info("status transition completed",
        zap.Uint64("collaborator_id", collaborator.ID),
        zap.String("from", string(oldStatus)),
        zap.String("to", string(toStatus)),
        zap.String("user", user.Email))

    return nil
}
```

### Prerequisite Implementations

```go
// internal/domain/collaborator/prerequisites.go

func RequireGSTVerification(ctx context.Context, c *Collaborator) error {
    if c.GSTVerifiedAt == nil {
        return errors.New("GST verification pending")
    }

    // Check if verification is not too old
    if time.Since(*c.GSTVerifiedAt) > 90*24*time.Hour {
        return errors.New("GST verification expired (>90 days)")
    }

    return nil
}

func RequireBankVerification(ctx context.Context, c *Collaborator) error {
    if c.BankDetails == nil {
        return errors.New("bank details not provided")
    }

    if !c.BankDetails.Verified {
        return errors.New("bank account verification pending")
    }

    // Penny drop test must be successful
    if c.BankDetails.PennyDropStatus != "SUCCESS" {
        return errors.New("bank account verification failed")
    }

    return nil
}

func RequireDocumentsUploaded(ctx context.Context, c *Collaborator) error {
    requiredDocs := []string{
        "GST_CERTIFICATE",
        "PAN_CARD",
        "BANK_STATEMENT",
        "ADDRESS_PROOF",
    }

    for _, docType := range requiredDocs {
        if !c.HasDocument(docType) {
            return fmt.Errorf("missing required document: %s", docType)
        }

        doc := c.GetDocument(docType)
        if doc.Status != "VERIFIED" {
            return fmt.Errorf("document %s not verified", docType)
        }
    }

    return nil
}

func RequireNoActiveOrders(ctx context.Context, c *Collaborator) error {
    var activeCount int64

    err := db.Model(&Order{}).
        Where("collaborator_id = ? AND status IN ?",
            c.ID, []string{"PENDING", "PROCESSING", "SHIPPED"}).
        Count(&activeCount).Error

    if err != nil {
        return fmt.Errorf("failed to check active orders: %w", err)
    }

    if activeCount > 0 {
        return fmt.Errorf("collaborator has %d active orders", activeCount)
    }

    return nil
}

func RequireManagerApproval(ctx context.Context, c *Collaborator) error {
    approval, err := getLatestApproval(c.ID, "MANAGER_APPROVAL")
    if err != nil {
        return fmt.Errorf("failed to check approval: %w", err)
    }

    if approval == nil {
        return errors.New("manager approval required")
    }

    if time.Since(approval.CreatedAt) > 30*24*time.Hour {
        return errors.New("manager approval expired (>30 days)")
    }

    return nil
}

func RequireNoPendingPayments(ctx context.Context, c *Collaborator) error {
    var pendingAmount decimal.Decimal

    err := db.Model(&Payment{}).
        Where("collaborator_id = ? AND status = ?", c.ID, "PENDING").
        Select("COALESCE(SUM(amount), 0)").
        Scan(&pendingAmount).Error

    if err != nil {
        return fmt.Errorf("failed to check pending payments: %w", err)
    }

    if pendingAmount.GreaterThan(decimal.Zero) {
        return fmt.Errorf("collaborator has pending payments worth %s",
            pendingAmount.String())
    }

    return nil
}
```

### Transition Actions

```go
// internal/domain/collaborator/actions.go

func UpdateVerificationTimestamp(ctx context.Context, c *Collaborator) error {
    now := time.Now()
    c.VerifiedAt = &now
    c.VerifiedBy = getUserFromContext(ctx)
    return nil
}

func EnableTransactions(ctx context.Context, c *Collaborator) error {
    c.TransactionsEnabled = true
    c.TransactionsEnabledAt = time.Now()

    // Enable in payment gateway
    if err := paymentGateway.EnableMerchant(c.ID); err != nil {
        return fmt.Errorf("failed to enable in payment gateway: %w", err)
    }

    return nil
}

func DisableAllTransactions(ctx context.Context, c *Collaborator) error {
    c.TransactionsEnabled = false
    c.TransactionsDisabledAt = time.Now()

    // Disable in all systems
    errors := make([]error, 0)

    if err := paymentGateway.DisableMerchant(c.ID); err != nil {
        errors = append(errors, err)
    }

    if err := orderService.BlockNewOrders(c.ID); err != nil {
        errors = append(errors, err)
    }

    if err := inventoryService.FreezeAllocations(c.ID); err != nil {
        errors = append(errors, err)
    }

    if len(errors) > 0 {
        return fmt.Errorf("failed to disable transactions: %v", errors)
    }

    return nil
}

func SetupPaymentTerms(ctx context.Context, c *Collaborator) error {
    // Set default payment terms based on verification score
    score := calculateVerificationScore(c)

    terms := &PaymentTerms{
        CollaboratorID: c.ID,
        CreditLimit:    calculateCreditLimit(score),
        PaymentDays:    calculatePaymentDays(score),
        DiscountRate:   calculateDiscountRate(score),
        EffectiveFrom:  time.Now(),
    }

    return db.Create(terms).Error
}

func NotifyCollaboratorActivated(ctx context.Context, c *Collaborator) error {
    notification := &Notification{
        Type:      "COLLABORATOR_ACTIVATED",
        Recipient: c.Email,
        Subject:   "Your account has been activated",
        Template:  "collaborator_activated",
        Data: map[string]interface{}{
            "name":           c.Name,
            "credit_limit":   c.CreditLimit,
            "payment_terms":  c.PaymentTerms,
        },
    }

    return notificationService.Send(notification)
}
```

### Handler Integration

```go
// internal/grpc/handlers/collaborator/update.go

func (h *UpdateCollaboratorHandler) UpdateStatus(
    ctx context.Context,
    req *pb.UpdateStatusRequest,
) (*pb.StatusResponse, error) {
    // Load collaborator
    var collaborator models.Collaborator
    if err := h.db.First(&collaborator, req.CollaboratorId).Error; err != nil {
        return nil, status.Errorf(codes.NotFound,
            "collaborator not found: %v", err)
    }

    // Get user from context
    user := getUserFromContext(ctx)

    // Parse new status
    newStatus := CollaboratorStatus(req.NewStatus)

    // Execute transition through state machine
    if err := h.stateMachine.ExecuteTransition(
        ctx, &collaborator, newStatus, user, req.Reason,
    ); err != nil {
        h.logger.Error("status transition failed",
            zap.Uint64("collaborator_id", req.CollaboratorId),
            zap.String("to_status", req.NewStatus),
            zap.Error(err))

        return nil, status.Errorf(codes.FailedPrecondition,
            "status transition failed: %v", err)
    }

    return &pb.StatusResponse{
        Success: true,
        Message: fmt.Sprintf("Status updated to %s", newStatus),
        NewStatus: string(collaborator.Status),
        UpdatedAt: timestamppb.New(collaborator.StatusUpdatedAt),
    }, nil
}
```

### State Query Interface

```go
// internal/domain/collaborator/queries.go

type StateQuery struct {
    db *gorm.DB
    sm *StateMachine
}

func (q *StateQuery) GetAvailableTransitions(
    ctx context.Context,
    collaborator *Collaborator,
    user *User,
) []TransitionOption {
    options := []TransitionOption{}

    // Check all possible transitions from current state
    for key, transition := range q.sm.transitions {
        if !strings.HasPrefix(key, string(collaborator.Status)+"->) {
            continue
        }

        // Check if user has required role
        if !q.sm.hasRequiredRole(user, transition.RequiredRole) {
            continue
        }

        // Check prerequisites (non-blocking)
        prereqErrors := []string{}
        for _, prereq := range transition.Prerequisites {
            if err := prereq(ctx, collaborator); err != nil {
                prereqErrors = append(prereqErrors, err.Error())
            }
        }

        options = append(options, TransitionOption{
            ToStatus:      transition.To,
            Event:         transition.Event,
            Available:     len(prereqErrors) == 0,
            Prerequisites: prereqErrors,
            RequiredRole:  transition.RequiredRole,
        })
    }

    return options
}

func (q *StateQuery) GetTransitionHistory(
    collaboratorID uint64,
    limit int,
) ([]TransitionHistory, error) {
    var history []TransitionHistory

    err := q.db.Model(&AuditEntry{}).
        Where("entity_type = ? AND entity_id = ? AND action LIKE ?",
            "collaborator", collaboratorID, "status_transition_%").
        Order("created_at DESC").
        Limit(limit).
        Find(&history).Error

    return history, err
}
```

---

## Testing Strategy

### Unit Tests

```go
func TestStateMachine_ValidTransitions(t *testing.T) {
    sm := NewCollaboratorStateMachine()

    tests := []struct {
        name     string
        from     CollaboratorStatus
        to       CollaboratorStatus
        valid    bool
    }{
        {"pending to verified", StatusPending, StatusVerified, true},
        {"verified to active", StatusVerified, StatusActive, true},
        {"pending to active", StatusPending, StatusActive, false}, // Invalid!
        {"active to suspended", StatusActive, StatusSuspended, true},
        {"suspended to active", StatusSuspended, StatusActive, true},
        {"rejected to active", StatusRejected, StatusActive, false}, // Invalid!
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            collaborator := &Collaborator{Status: tt.from}
            user := &User{Roles: []string{"ADMIN"}}

            err := sm.CanTransition(context.Background(),
                collaborator, tt.to, user)

            if tt.valid {
                // Should pass authorization, may fail prerequisites
                assert.NotContains(t, err.Error(), "invalid transition")
            } else {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), "invalid transition")
            }
        })
    }
}

func TestStateMachine_RoleAuthorization(t *testing.T) {
    sm := NewCollaboratorStateMachine()
    collaborator := &Collaborator{
        Status: StatusActive,
        // All prerequisites met
    }

    tests := []struct {
        role     string
        toStatus CollaboratorStatus
        allowed  bool
    }{
        {"ADMIN", StatusSuspended, true},
        {"MANAGER", StatusSuspended, false}, // Needs ADMIN or RISK_MANAGER
        {"RISK_MANAGER", StatusSuspended, true},
        {"VIEWER", StatusSuspended, false},
    }

    for _, tt := range tests {
        user := &User{Roles: []string{tt.role}}

        err := sm.CanTransition(context.Background(),
            collaborator, tt.toStatus, user)

        if tt.allowed {
            assert.NotContains(t, err.Error(), "unauthorized")
        } else {
            assert.Contains(t, err.Error(), "unauthorized")
        }
    }
}

func TestStateMachine_Prerequisites(t *testing.T) {
    sm := NewCollaboratorStateMachine()

    t.Run("pending to verified requires documents", func(t *testing.T) {
        collaborator := &Collaborator{
            Status: StatusPending,
            // No documents uploaded
        }

        user := &User{Roles: []string{"ADMIN"}}

        err := sm.CanTransition(context.Background(),
            collaborator, StatusVerified, user)

        assert.Error(t, err)
        assert.Contains(t, err.Error(), "document")
    })

    t.Run("active to suspended requires no active orders", func(t *testing.T) {
        collaborator := &Collaborator{
            Status: StatusActive,
            ID:     123,
        }

        // Mock active order
        mockDB.ExpectQuery("SELECT COUNT").
            WithArgs(123, []string{"PENDING", "PROCESSING", "SHIPPED"}).
            WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

        user := &User{Roles: []string{"ADMIN"}}

        err := sm.CanTransition(context.Background(),
            collaborator, StatusSuspended, user)

        assert.Error(t, err)
        assert.Contains(t, err.Error(), "5 active orders")
    })
}
```

### Integration Tests

```go
func TestCollaboratorStatusTransition_EndToEnd(t *testing.T) {
    // Setup
    handler := setupTestHandler()
    collaborator := createTestCollaborator(StatusPending)

    // Step 1: Verify collaborator
    verifyReq := &pb.UpdateStatusRequest{
        CollaboratorId: collaborator.ID,
        NewStatus:      "VERIFIED",
        Reason:         "All documents verified",
    }

    resp, err := handler.UpdateStatus(adminContext(), verifyReq)
    assert.NoError(t, err)
    assert.Equal(t, "VERIFIED", resp.NewStatus)

    // Step 2: Activate collaborator
    activateReq := &pb.UpdateStatusRequest{
        CollaboratorId: collaborator.ID,
        NewStatus:      "ACTIVE",
        Reason:         "Approved for trading",
    }

    resp, err = handler.UpdateStatus(adminContext(), activateReq)
    assert.NoError(t, err)
    assert.Equal(t, "ACTIVE", resp.NewStatus)

    // Step 3: Try invalid transition (should fail)
    invalidReq := &pb.UpdateStatusRequest{
        CollaboratorId: collaborator.ID,
        NewStatus:      "PENDING", // Can't go back to pending
        Reason:         "Test invalid",
    }

    _, err = handler.UpdateStatus(adminContext(), invalidReq)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "invalid transition")

    // Verify audit trail
    var audits []AuditEntry
    db.Where("entity_id = ?", collaborator.ID).Find(&audits)
    assert.Len(t, audits, 2) // Two successful transitions
}
```

---

## Monitoring & Metrics

### Required Metrics

```go
type StateMetrics struct {
    TransitionsAttempted   counter
    TransitionsSucceeded   counter
    TransitionsFailed      counter
    UnauthorizedAttempts   counter
    PrerequisiteFailures   counter
    TransitionDuration     histogram
    StateDistribution      gauge // Current count per state
}
```

### Dashboards

```yaml
grafana_panels:
  - title: "Status Distribution"
    query: "collaborator_status_distribution"
    type: "pie_chart"

  - title: "Transition Success Rate"
    query: "rate(transitions_succeeded) / rate(transitions_attempted)"
    type: "stat"

  - title: "Unauthorized Attempts"
    query: "increase(unauthorized_attempts[1h])"
    type: "time_series"

  - title: "Common Prerequisite Failures"
    query: "topk(10, increase(prerequisite_failures[1h]))"
    type: "bar_chart"

  - title: "Average Time in Each State"
    query: "collaborator_state_duration_avg"
    type: "heatmap"
```

### Alerts

```yaml
alerts:
  - name: HighTransitionFailureRate
    expr: |
      rate(transitions_failed[5m]) / rate(transitions_attempted[5m]) > 0.1
    severity: warning
    annotations:
      summary: "High transition failure rate: {{ $value | humanizePercentage }}"

  - name: UnauthorizedStatusChangeAttempts
    expr: increase(unauthorized_attempts[5m]) > 10
    severity: critical
    annotations:
      summary: "Multiple unauthorized status change attempts detected"

  - name: StuckInPendingState
    expr: collaborator_status_distribution{status="PENDING"} > 100
    for: 24h
    severity: warning
    annotations:
      summary: "{{ $value }} collaborators stuck in PENDING for >24h"
```

---

## Security Considerations

### Authorization Matrix

| Role | VERIFY | ACTIVATE | SUSPEND | REINSTATE | REJECT | DEACTIVATE |
|------|--------|----------|---------|-----------|---------|------------|
| ADMIN | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| MANAGER | ✓ | ✓ | ✗ | ✗ | ✓ | ✗ |
| VERIFIER | ✓ | ✗ | ✗ | ✗ | ✓ | ✗ |
| RISK_MANAGER | ✗ | ✗ | ✓ | ✗ | ✗ | ✗ |
| VIEWER | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ |

### Audit Requirements

Every state transition MUST:
1. Record old and new status
2. Capture user ID and role
3. Include timestamp and IP
4. Store reason/justification
5. Be immutable once written

---

## Implementation Checklist

- [ ] Define all status constants
- [ ] Implement StateMachine core
- [ ] Define all transitions
- [ ] Implement all prerequisites
- [ ] Implement all actions
- [ ] Add compensation handlers
- [ ] Integrate with handlers
- [ ] Add comprehensive tests
- [ ] Add metrics collection
- [ ] Create monitoring dashboards
- [ ] Document state diagram
- [ ] Add audit logging
- [ ] Create admin UI

---

## Risk Mitigation

### Risk 1: Bypassed Transitions
**Mitigation**: All status updates MUST go through state machine

### Risk 2: Stuck States
**Mitigation**: Timeout mechanisms and manual override with audit

### Risk 3: Race Conditions
**Mitigation**: Database-level locks on status updates

### Risk 4: Authorization Bypass
**Mitigation**: Role checks at multiple levels

---

## References

- Martin Fowler: State Pattern
- Microsoft: Workflow Patterns
- AWS Step Functions: State Machines
- BPMN 2.0 Specification

---

**Approval Required From**: Security Team, Compliance Team, Product Manager
**Implementation Timeline**: 6 hours
**Testing Timeline**: 4 hours
**Rollout Strategy**: Feature flag with gradual enablement