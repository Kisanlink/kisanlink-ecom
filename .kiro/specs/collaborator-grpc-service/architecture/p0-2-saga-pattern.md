# P0-2: Saga Pattern Architecture for AAA-Collaborator Transaction Integrity

**Version**: 1.0
**Date**: 2025-11-12
**Priority**: P0 CRITICAL
**Author**: SDE-3 Backend Architect

---

## Executive Summary

This document provides production-ready architecture for implementing the Saga pattern to ensure transactional integrity between AAA service (Address) and Collaborator service. The solution implements an orchestrated saga with compensating transactions to handle distributed transaction failures.

---

## Problem Statement

### Current Issue
When creating or updating a collaborator:
1. Address is created in AAA service (success)
2. Collaborator save to database fails
3. Result: Orphaned address in AAA with no corresponding collaborator

### Data Flow Problem
```
┌─────────┐      ┌─────────┐      ┌──────────┐      ┌────────┐
│ Client  │─────>│ Handler │─────>│   AAA    │─────>│   DB   │
└─────────┘      └─────────┘      └──────────┘      └────────┘
                      │                 ✓                ✗
                      └────────────────────────────────────┘
                         No rollback! Orphaned address!
```

### Business Impact
- Data inconsistency between services
- Orphaned addresses consuming resources
- Financial reconciliation errors
- Audit trail gaps

---

## Architecture Design

### Saga Pattern Overview

```
┌──────────────────────────────────────────────────────────────┐
│                     Saga Orchestrator                        │
├──────────────────────────────────────────────────────────────┤
│  ┌───────────────────────────────────────────────────┐       │
│  │           Transaction Definition (T)               │       │
│  │  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐         │       │
│  │  │Step 1│─>│Step 2│─>│Step 3│─>│Step 4│         │       │
│  │  └──────┘  └──────┘  └──────┘  └──────┘         │       │
│  └───────────────────────────────────────────────────┘       │
│                                                               │
│  ┌───────────────────────────────────────────────────┐       │
│  │         Compensation Definition (C)               │       │
│  │  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐         │       │
│  │  │Comp 4│<─│Comp 3│<─│Comp 2│<─│Comp 1│         │       │
│  │  └──────┘  └──────┘  └──────┘  └──────┘         │       │
│  └───────────────────────────────────────────────────┘       │
├──────────────────────────────────────────────────────────────┤
│                    Execution Engine                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐    │
│  │  State   │  │  Logger  │  │  Metrics │  │Recovery  │    │
│  │  Manager │  │          │  │          │  │  Handler │    │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘    │
└──────────────────────────────────────────────────────────────┘
```

### Core Components

```go
// internal/saga/saga.go

type SagaStep struct {
    Name         string
    Execute      StepFunc
    Compensate   StepFunc
    Retryable    bool
    MaxRetries   int
    Timeout      time.Duration
}

type StepFunc func(ctx context.Context, data interface{}) error

type Saga struct {
    ID           string
    Name         string
    Steps        []SagaStep
    State        SagaState
    Context      map[string]interface{}
    CompletedSteps []string
    logger       *zap.Logger
    metrics      *SagaMetrics
}

type SagaState string

const (
    SagaStatePending     SagaState = "pending"
    SagaStateRunning     SagaState = "running"
    SagaStateCompleted   SagaState = "completed"
    SagaStateFailed      SagaState = "failed"
    SagaStateCompensating SagaState = "compensating"
    SagaStateCompensated SagaState = "compensated"
)

type SagaExecutor struct {
    storage      SagaStorage
    logger       *zap.Logger
    metrics      *SagaMetrics
    tracer       trace.Tracer
    retryPolicy  RetryPolicy
}
```

### Saga Implementation

```go
// internal/saga/executor.go

func (e *SagaExecutor) Execute(ctx context.Context, saga *Saga) error {
    // Start distributed tracing
    ctx, span := e.tracer.Start(ctx, "saga.execute",
        trace.WithAttributes(
            attribute.String("saga.id", saga.ID),
            attribute.String("saga.name", saga.Name),
        ))
    defer span.End()

    // Initialize saga state
    saga.State = SagaStateRunning
    saga.CompletedSteps = make([]string, 0, len(saga.Steps))

    if err := e.storage.SaveState(ctx, saga); err != nil {
        return fmt.Errorf("failed to save initial state: %w", err)
    }

    // Execute forward transaction
    for i, step := range saga.Steps {
        stepCtx, stepSpan := e.tracer.Start(ctx, "saga.step.execute",
            trace.WithAttributes(
                attribute.String("step.name", step.Name),
                attribute.Int("step.index", i),
            ))

        e.logger.Info("executing saga step",
            zap.String("saga_id", saga.ID),
            zap.String("step", step.Name))

        // Execute with retry if configured
        err := e.executeStepWithRetry(stepCtx, step, saga.Context)

        stepSpan.End()

        if err != nil {
            e.logger.Error("saga step failed",
                zap.String("saga_id", saga.ID),
                zap.String("step", step.Name),
                zap.Error(err))

            // Record failure metrics
            e.metrics.RecordStepFailure(saga.Name, step.Name)

            // Trigger compensation
            if compensateErr := e.compensate(ctx, saga, i-1); compensateErr != nil {
                e.logger.Error("compensation failed",
                    zap.String("saga_id", saga.ID),
                    zap.Error(compensateErr))

                saga.State = SagaStateFailed
                e.storage.SaveState(ctx, saga)

                return fmt.Errorf("step failed and compensation failed: %w", compensateErr)
            }

            saga.State = SagaStateCompensated
            e.storage.SaveState(ctx, saga)

            return fmt.Errorf("step %s failed, compensated: %w", step.Name, err)
        }

        // Mark step as completed
        saga.CompletedSteps = append(saga.CompletedSteps, step.Name)

        // Persist state after each successful step
        if err := e.storage.SaveState(ctx, saga); err != nil {
            e.logger.Warn("failed to save intermediate state",
                zap.String("saga_id", saga.ID),
                zap.String("step", step.Name),
                zap.Error(err))
        }

        e.metrics.RecordStepSuccess(saga.Name, step.Name)
    }

    // All steps completed successfully
    saga.State = SagaStateCompleted
    if err := e.storage.SaveState(ctx, saga); err != nil {
        return fmt.Errorf("failed to save final state: %w", err)
    }

    e.metrics.RecordSagaSuccess(saga.Name)

    return nil
}

func (e *SagaExecutor) compensate(ctx context.Context, saga *Saga, fromIndex int) error {
    ctx, span := e.tracer.Start(ctx, "saga.compensate",
        trace.WithAttributes(
            attribute.String("saga.id", saga.ID),
            attribute.Int("from_index", fromIndex),
        ))
    defer span.End()

    saga.State = SagaStateCompensating
    e.storage.SaveState(ctx, saga)

    // Compensate in reverse order
    for i := fromIndex; i >= 0; i-- {
        step := saga.Steps[i]

        // Check if this step was actually completed
        if !contains(saga.CompletedSteps, step.Name) {
            continue
        }

        if step.Compensate == nil {
            e.logger.Warn("no compensation defined for step",
                zap.String("step", step.Name))
            continue
        }

        e.logger.Info("compensating saga step",
            zap.String("saga_id", saga.ID),
            zap.String("step", step.Name))

        // Execute compensation with retry
        err := e.executeCompensationWithRetry(ctx, step, saga.Context)

        if err != nil {
            e.logger.Error("compensation failed for step",
                zap.String("saga_id", saga.ID),
                zap.String("step", step.Name),
                zap.Error(err))

            e.metrics.RecordCompensationFailure(saga.Name, step.Name)

            // Continue with other compensations despite failure
            // Log for manual intervention
            e.alerter.SendCritical(fmt.Sprintf(
                "Compensation failed for saga %s step %s - manual intervention required",
                saga.ID, step.Name))
        } else {
            e.metrics.RecordCompensationSuccess(saga.Name, step.Name)
        }
    }

    return nil
}
```

### Collaborator Creation Saga

```go
// internal/grpc/handlers/collaborator/create.go

func (h *CreateCollaboratorHandler) createCollaboratorSaga(
    ctx context.Context,
    req *pb.CreateCollaboratorRequest,
) (*models.Collaborator, error) {

    sagaID := generateSagaID()

    saga := &Saga{
        ID:   sagaID,
        Name: "create_collaborator",
        Context: map[string]interface{}{
            "request":     req,
            "fpo_id":      getFPOFromContext(ctx),
            "request_id":  getRequestIDFromContext(ctx),
            "addresses":   []uint64{}, // Will store created address IDs
        },
        Steps: []SagaStep{
            {
                Name: "validate_gst",
                Execute: func(ctx context.Context, data interface{}) error {
                    sagaCtx := data.(map[string]interface{})
                    req := sagaCtx["request"].(*pb.CreateCollaboratorRequest)

                    // Validate GST format
                    if err := h.gstValidator.Validate(req.GstNumber); err != nil {
                        return fmt.Errorf("invalid GST: %w", err)
                    }

                    // Check for duplicates with distributed lock
                    reservation, err := h.gstService.CheckAndReserveGST(
                        ctx, req.GstNumber, sagaCtx["fpo_id"].(uint64))

                    if err != nil {
                        return err
                    }

                    sagaCtx["gst_reservation"] = reservation
                    return nil
                },
                Compensate: func(ctx context.Context, data interface{}) error {
                    sagaCtx := data.(map[string]interface{})
                    if reservation, ok := sagaCtx["gst_reservation"]; ok {
                        return h.gstService.ReleaseReservation(ctx,
                            reservation.(*GSTReservation))
                    }
                    return nil
                },
                Retryable:  false,
                MaxRetries: 0,
                Timeout:    5 * time.Second,
            },
            {
                Name: "create_primary_address",
                Execute: func(ctx context.Context, data interface{}) error {
                    sagaCtx := data.(map[string]interface{})
                    req := sagaCtx["request"].(*pb.CreateCollaboratorRequest)

                    if req.PrimaryAddress == nil {
                        return nil
                    }

                    // Create address in AAA service
                    addressReq := &aaa.CreateAddressRequest{
                        Street:      req.PrimaryAddress.Street,
                        Area:        req.PrimaryAddress.Area,
                        City:        req.PrimaryAddress.City,
                        State:       req.PrimaryAddress.State,
                        Pincode:     req.PrimaryAddress.Pincode,
                        AddressType: "BUSINESS",
                    }

                    resp, err := h.aaaClient.CreateAddress(ctx, addressReq)
                    if err != nil {
                        return fmt.Errorf("failed to create primary address: %w", err)
                    }

                    // Store address ID for potential compensation
                    addresses := sagaCtx["addresses"].([]uint64)
                    addresses = append(addresses, resp.AddressID)
                    sagaCtx["addresses"] = addresses
                    sagaCtx["primary_address_id"] = resp.AddressID

                    return nil
                },
                Compensate: func(ctx context.Context, data interface{}) error {
                    sagaCtx := data.(map[string]interface{})

                    if addressID, ok := sagaCtx["primary_address_id"].(uint64); ok {
                        req := &aaa.DeleteAddressRequest{
                            AddressID: addressID,
                            Reason:    "Saga compensation - collaborator creation failed",
                        }

                        // Retry deletion with exponential backoff
                        return h.retryPolicy.Execute(ctx, func() error {
                            _, err := h.aaaClient.DeleteAddress(ctx, req)
                            return err
                        })
                    }
                    return nil
                },
                Retryable:  true,
                MaxRetries: 3,
                Timeout:    10 * time.Second,
            },
            {
                Name: "create_shipping_address",
                Execute: func(ctx context.Context, data interface{}) error {
                    sagaCtx := data.(map[string]interface{})
                    req := sagaCtx["request"].(*pb.CreateCollaboratorRequest)

                    if req.ShippingAddress == nil {
                        return nil
                    }

                    addressReq := &aaa.CreateAddressRequest{
                        Street:      req.ShippingAddress.Street,
                        Area:        req.ShippingAddress.Area,
                        City:        req.ShippingAddress.City,
                        State:       req.ShippingAddress.State,
                        Pincode:     req.ShippingAddress.Pincode,
                        AddressType: "SHIPPING",
                    }

                    resp, err := h.aaaClient.CreateAddress(ctx, addressReq)
                    if err != nil {
                        return fmt.Errorf("failed to create shipping address: %w", err)
                    }

                    addresses := sagaCtx["addresses"].([]uint64)
                    addresses = append(addresses, resp.AddressID)
                    sagaCtx["addresses"] = addresses
                    sagaCtx["shipping_address_id"] = resp.AddressID

                    return nil
                },
                Compensate: func(ctx context.Context, data interface{}) error {
                    sagaCtx := data.(map[string]interface{})

                    if addressID, ok := sagaCtx["shipping_address_id"].(uint64); ok {
                        req := &aaa.DeleteAddressRequest{
                            AddressID: addressID,
                            Reason:    "Saga compensation - collaborator creation failed",
                        }

                        return h.retryPolicy.Execute(ctx, func() error {
                            _, err := h.aaaClient.DeleteAddress(ctx, req)
                            return err
                        })
                    }
                    return nil
                },
                Retryable:  true,
                MaxRetries: 3,
                Timeout:    10 * time.Second,
            },
            {
                Name: "save_collaborator",
                Execute: func(ctx context.Context, data interface{}) error {
                    sagaCtx := data.(map[string]interface{})
                    req := sagaCtx["request"].(*pb.CreateCollaboratorRequest)

                    collaborator := &models.Collaborator{
                        FPO:               sagaCtx["fpo_id"].(uint64),
                        Name:              req.Name,
                        GSTNumber:         req.GstNumber,
                        PAN:              req.Pan,
                        Mobile:           req.Mobile,
                        Email:            req.Email,
                        PrimaryAddressID:  sagaCtx["primary_address_id"].(uint64),
                        ShippingAddressID: sagaCtx["shipping_address_id"].(uint64),
                        BankDetails:      convertBankDetails(req.BankDetails),
                        Status:           models.CollaboratorStatusPending,
                        IsMaster:         req.IsMaster,
                        CreatedBy:        getUserFromContext(ctx),
                    }

                    // Use transaction for database save
                    err := h.db.Transaction(func(tx *gorm.DB) error {
                        if err := tx.Create(collaborator).Error; err != nil {
                            return err
                        }

                        // Create audit log
                        audit := &models.AuditLog{
                            EntityType: "collaborator",
                            EntityID:   collaborator.ID,
                            Action:     "create",
                            UserID:     getUserFromContext(ctx),
                            Changes:    serializeChanges(nil, collaborator),
                            IP:         getIPFromContext(ctx),
                            UserAgent:  getUserAgentFromContext(ctx),
                        }

                        return tx.Create(audit).Error
                    })

                    if err != nil {
                        return fmt.Errorf("failed to save collaborator: %w", err)
                    }

                    sagaCtx["collaborator_id"] = collaborator.ID
                    sagaCtx["collaborator"] = collaborator

                    return nil
                },
                Compensate: func(ctx context.Context, data interface{}) error {
                    sagaCtx := data.(map[string]interface{})

                    if collaboratorID, ok := sagaCtx["collaborator_id"].(uint64); ok {
                        // Soft delete to maintain audit trail
                        return h.db.Transaction(func(tx *gorm.DB) error {
                            err := tx.Model(&models.Collaborator{}).
                                Where("id = ?", collaboratorID).
                                Updates(map[string]interface{}{
                                    "deleted_at": time.Now(),
                                    "deleted_by": "saga_compensation",
                                }).Error

                            if err != nil {
                                return err
                            }

                            // Log compensation
                            audit := &models.AuditLog{
                                EntityType: "collaborator",
                                EntityID:   collaboratorID,
                                Action:     "compensate_delete",
                                UserID:     "system",
                                Changes:    "Deleted due to saga compensation",
                            }

                            return tx.Create(audit).Error
                        })
                    }
                    return nil
                },
                Retryable:  true,
                MaxRetries: 3,
                Timeout:    5 * time.Second,
            },
            {
                Name: "send_notifications",
                Execute: func(ctx context.Context, data interface{}) error {
                    sagaCtx := data.(map[string]interface{})
                    collaborator := sagaCtx["collaborator"].(*models.Collaborator)

                    // Send welcome email (non-critical, can fail)
                    go h.notificationService.SendWelcomeEmail(collaborator)

                    // Send SMS notification (non-critical, can fail)
                    go h.notificationService.SendWelcomeSMS(collaborator)

                    return nil
                },
                Compensate: nil, // No compensation needed for notifications
                Retryable:  false,
                MaxRetries: 0,
                Timeout:    2 * time.Second,
            },
        },
    }

    // Execute saga
    if err := h.sagaExecutor.Execute(ctx, saga); err != nil {
        h.logger.Error("saga execution failed",
            zap.String("saga_id", sagaID),
            zap.Error(err))
        return nil, err
    }

    return saga.Context["collaborator"].(*models.Collaborator), nil
}
```

### Compensation Handlers

```go
// internal/saga/compensation.go

type CompensationHandler struct {
    aaaClient    aaa.Client
    db           *gorm.DB
    logger       *zap.Logger
    metrics      *CompensationMetrics
    retryPolicy  RetryPolicy
}

// Address deletion compensation with retry
func (h *CompensationHandler) CompensateAddressDeletion(
    ctx context.Context,
    addressID uint64,
    reason string,
) error {
    maxAttempts := 5
    baseDelay := 100 * time.Millisecond

    for attempt := 1; attempt <= maxAttempts; attempt++ {
        req := &aaa.DeleteAddressRequest{
            AddressID: addressID,
            Reason:    reason,
            Force:     attempt >= 3, // Force delete after 3 attempts
        }

        _, err := h.aaaClient.DeleteAddress(ctx, req)

        if err == nil {
            h.metrics.RecordCompensationSuccess("address_deletion")
            return nil
        }

        // Check if error is retryable
        if !isRetryableError(err) {
            h.logger.Error("non-retryable error in compensation",
                zap.Uint64("address_id", addressID),
                zap.Error(err))
            return err
        }

        // Exponential backoff
        delay := baseDelay * time.Duration(math.Pow(2, float64(attempt-1)))

        h.logger.Warn("compensation attempt failed, retrying",
            zap.Uint64("address_id", addressID),
            zap.Int("attempt", attempt),
            zap.Duration("delay", delay),
            zap.Error(err))

        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(delay):
            continue
        }
    }

    // All attempts failed - alert for manual intervention
    h.alerter.SendCritical(fmt.Sprintf(
        "Failed to compensate address deletion after %d attempts. AddressID: %d",
        maxAttempts, addressID))

    h.metrics.RecordCompensationFailure("address_deletion")

    return fmt.Errorf("compensation failed after %d attempts", maxAttempts)
}

// GST reservation release compensation
func (h *CompensationHandler) CompensateGSTReservation(
    ctx context.Context,
    reservation *GSTReservation,
) error {
    if reservation == nil {
        return nil
    }

    err := h.gstService.ReleaseReservation(ctx, reservation)
    if err != nil {
        h.logger.Error("failed to release GST reservation",
            zap.String("gst", reservation.GST),
            zap.Error(err))

        h.metrics.RecordCompensationFailure("gst_reservation")
        return err
    }

    h.metrics.RecordCompensationSuccess("gst_reservation")
    return nil
}
```

---

## State Persistence

### Saga State Storage

```go
// internal/saga/storage.go

type SagaStorage interface {
    SaveState(ctx context.Context, saga *Saga) error
    LoadState(ctx context.Context, sagaID string) (*Saga, error)
    ListPendingSagas(ctx context.Context) ([]*Saga, error)
    CleanupCompleted(ctx context.Context, olderThan time.Duration) error
}

type DBSagaStorage struct {
    db *gorm.DB
}

type SagaStateRecord struct {
    ID             string         `gorm:"primaryKey"`
    Name           string
    State          string
    Context        datatypes.JSON
    CompletedSteps datatypes.JSON
    CreatedAt      time.Time
    UpdatedAt      time.Time
    CompletedAt    *time.Time
}

func (s *DBSagaStorage) SaveState(ctx context.Context, saga *Saga) error {
    contextJSON, err := json.Marshal(saga.Context)
    if err != nil {
        return err
    }

    stepsJSON, err := json.Marshal(saga.CompletedSteps)
    if err != nil {
        return err
    }

    record := &SagaStateRecord{
        ID:             saga.ID,
        Name:           saga.Name,
        State:          string(saga.State),
        Context:        contextJSON,
        CompletedSteps: stepsJSON,
    }

    if saga.State == SagaStateCompleted || saga.State == SagaStateCompensated {
        now := time.Now()
        record.CompletedAt = &now
    }

    return s.db.WithContext(ctx).Save(record).Error
}
```

---

## Recovery Mechanism

### Saga Recovery on Service Restart

```go
// internal/saga/recovery.go

type SagaRecoveryService struct {
    executor *SagaExecutor
    storage  SagaStorage
    logger   *zap.Logger
}

func (r *SagaRecoveryService) RecoverPendingSagas(ctx context.Context) error {
    // Load all incomplete sagas
    pendingSagas, err := r.storage.ListPendingSagas(ctx)
    if err != nil {
        return fmt.Errorf("failed to load pending sagas: %w", err)
    }

    r.logger.Info("recovering pending sagas",
        zap.Int("count", len(pendingSagas)))

    for _, saga := range pendingSagas {
        // Check saga age
        if time.Since(saga.CreatedAt) > 24*time.Hour {
            r.logger.Warn("saga too old, marking as failed",
                zap.String("saga_id", saga.ID))

            saga.State = SagaStateFailed
            r.storage.SaveState(ctx, saga)
            continue
        }

        // Attempt to continue or compensate
        go r.recoverSaga(ctx, saga)
    }

    return nil
}

func (r *SagaRecoveryService) recoverSaga(ctx context.Context, saga *Saga) {
    switch saga.State {
    case SagaStateRunning:
        // Determine last successful step and continue
        lastStep := len(saga.CompletedSteps)

        // Restart from the failed step
        if err := r.executor.ExecuteFrom(ctx, saga, lastStep); err != nil {
            r.logger.Error("failed to recover saga",
                zap.String("saga_id", saga.ID),
                zap.Error(err))

            // Trigger compensation
            r.executor.Compensate(ctx, saga)
        }

    case SagaStateCompensating:
        // Continue compensation
        if err := r.executor.Compensate(ctx, saga); err != nil {
            r.logger.Error("failed to complete compensation",
                zap.String("saga_id", saga.ID),
                zap.Error(err))

            // Alert for manual intervention
            r.alerter.SendCritical(fmt.Sprintf(
                "Saga %s requires manual intervention", saga.ID))
        }
    }
}
```

---

## Testing Strategy

### Unit Tests

```go
func TestSaga_SuccessfulExecution(t *testing.T) {
    saga := &Saga{
        ID:   "test-saga-1",
        Name: "test",
        Steps: []SagaStep{
            {
                Name: "step1",
                Execute: func(ctx context.Context, data interface{}) error {
                    data.(map[string]interface{})["step1"] = "completed"
                    return nil
                },
            },
            {
                Name: "step2",
                Execute: func(ctx context.Context, data interface{}) error {
                    data.(map[string]interface{})["step2"] = "completed"
                    return nil
                },
            },
        },
        Context: make(map[string]interface{}),
    }

    executor := NewSagaExecutor(storage, logger, metrics)
    err := executor.Execute(context.Background(), saga)

    assert.NoError(t, err)
    assert.Equal(t, SagaStateCompleted, saga.State)
    assert.Len(t, saga.CompletedSteps, 2)
}

func TestSaga_CompensationOnFailure(t *testing.T) {
    compensationCalled := atomic.Bool{}

    saga := &Saga{
        ID:   "test-saga-2",
        Name: "test",
        Steps: []SagaStep{
            {
                Name: "step1",
                Execute: func(ctx context.Context, data interface{}) error {
                    return nil
                },
                Compensate: func(ctx context.Context, data interface{}) error {
                    compensationCalled.Store(true)
                    return nil
                },
            },
            {
                Name: "step2",
                Execute: func(ctx context.Context, data interface{}) error {
                    return errors.New("deliberate failure")
                },
            },
        },
        Context: make(map[string]interface{}),
    }

    executor := NewSagaExecutor(storage, logger, metrics)
    err := executor.Execute(context.Background(), saga)

    assert.Error(t, err)
    assert.True(t, compensationCalled.Load())
    assert.Equal(t, SagaStateCompensated, saga.State)
}
```

### Integration Tests

```go
func TestCollaboratorCreation_WithSaga(t *testing.T) {
    // Setup test environment
    aaaServer := setupMockAAAServer()
    defer aaaServer.Close()

    handler := setupCollaboratorHandler()

    req := &pb.CreateCollaboratorRequest{
        Name:      "Test Collaborator",
        GstNumber: "29ABCDE1234F1Z5",
        PrimaryAddress: &pb.Address{
            Street:  "123 Test St",
            City:    "Bangalore",
            State:   "Karnataka",
            Pincode: "560001",
        },
    }

    // Execute
    resp, err := handler.CreateCollaborator(context.Background(), req)

    assert.NoError(t, err)
    assert.NotNil(t, resp)

    // Verify collaborator created
    var collaborator models.Collaborator
    err = db.Where("gst_number = ?", req.GstNumber).First(&collaborator).Error
    assert.NoError(t, err)

    // Verify address created in AAA
    address, err := aaaClient.GetAddress(ctx, collaborator.PrimaryAddressID)
    assert.NoError(t, err)
    assert.Equal(t, req.PrimaryAddress.Street, address.Street)
}

func TestCollaboratorCreation_CompensationOnDBFailure(t *testing.T) {
    // Setup
    aaaServer := setupMockAAAServer()
    defer aaaServer.Close()

    // Simulate DB failure
    db.Close()

    handler := setupCollaboratorHandler()

    req := &pb.CreateCollaboratorRequest{
        Name:      "Test Collaborator",
        GstNumber: "29ABCDE1234F1Z5",
        PrimaryAddress: &pb.Address{
            Street: "123 Test St",
        },
    }

    // Execute - should fail
    _, err := handler.CreateCollaborator(context.Background(), req)

    assert.Error(t, err)

    // Verify no address exists in AAA (compensated)
    time.Sleep(500 * time.Millisecond) // Wait for compensation

    addresses, err := aaaClient.ListAddresses(ctx)
    assert.NoError(t, err)
    assert.Empty(t, addresses)
}
```

---

## Monitoring & Metrics

### Required Metrics

```go
type SagaMetrics struct {
    // Saga level metrics
    SagasStarted    counter
    SagasCompleted  counter
    SagasFailed     counter
    SagasCompensated counter
    SagaDuration    histogram

    // Step level metrics
    StepsExecuted   counter
    StepsFailed     counter
    StepDuration    histogram

    // Compensation metrics
    CompensationsTriggered counter
    CompensationsSucceeded counter
    CompensationsFailed    counter
    CompensationDuration   histogram

    // Recovery metrics
    RecoveriesAttempted counter
    RecoveriesSucceeded counter
    RecoveriesFailed    counter
}
```

### Alerts

```yaml
alerts:
  - name: HighSagaFailureRate
    expr: rate(saga_failures[5m]) / rate(saga_started[5m]) > 0.05
    severity: warning
    annotations:
      summary: "High saga failure rate: {{ $value | humanizePercentage }}"

  - name: CompensationFailure
    expr: increase(compensation_failures[5m]) > 0
    severity: critical
    annotations:
      summary: "Saga compensation failures detected"

  - name: OrphanedAddresses
    expr: orphaned_addresses > 0
    severity: warning
    annotations:
      summary: "{{ $value }} orphaned addresses detected"

  - name: LongRunningSaga
    expr: saga_duration_p99 > 30s
    severity: warning
    annotations:
      summary: "Saga taking too long: P99 {{ $value }}s"
```

---

## Implementation Checklist

- [ ] Implement Saga core components
- [ ] Implement SagaExecutor with forward execution
- [ ] Implement compensation logic
- [ ] Add state persistence with database
- [ ] Implement recovery service
- [ ] Create collaborator saga definition
- [ ] Add AAA integration with compensation
- [ ] Implement retry policies
- [ ] Add distributed tracing
- [ ] Add comprehensive metrics
- [ ] Create monitoring dashboards
- [ ] Add integration tests
- [ ] Test failure scenarios
- [ ] Document runbooks

---

## Risk Mitigation

### Risk 1: Partial Compensation Failure
**Mitigation**: Log failures, alert operations, provide manual compensation tools

### Risk 2: Infinite Retry Loops
**Mitigation**: Max retry limits, circuit breakers, timeout policies

### Risk 3: State Corruption
**Mitigation**: Versioned state, checksums, regular backups

### Risk 4: Performance Degradation
**Mitigation**: Async compensation, batching, connection pooling

---

## References

- Microservices.io: Saga Pattern
- Chris Richardson: Microservices Patterns
- Google Cloud: Distributed Transactions
- AWS Step Functions: Saga Implementation

---

**Approval Required From**: Platform Architect, Data Team, SRE Team
**Implementation Timeline**: 10 hours
**Testing Timeline**: 6 hours
**Rollout Strategy**: Shadow mode → 1% traffic → 10% → 50% → 100%