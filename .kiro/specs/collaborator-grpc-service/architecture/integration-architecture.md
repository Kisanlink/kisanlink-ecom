# Integration Architecture for P0 Security Fixes

**Version**: 1.0
**Date**: 2025-11-12
**Priority**: P0 CRITICAL
**Author**: SDE-3 Backend Architect

---

## Executive Summary

This document provides the integration architecture showing how all P0 security fixes work together to create a secure, resilient Collaborator gRPC Service. It covers dependencies, interaction patterns, and deployment strategies.

---

## System Integration Overview

```
┌────────────────────────────────────────────────────────────────────┐
│                     gRPC Request Flow                               │
├────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  Client Request                                                      │
│       │                                                              │
│       ▼                                                              │
│  ┌─────────────┐    P0-3: JWT Validation                           │
│  │   JWT Auth  │────────────────────────────────┐                  │
│  └─────────────┘                                │                  │
│       │                                          ▼                  │
│       │                                    ┌──────────┐            │
│       │                                    │ Authz    │            │
│       │                                    └──────────┘            │
│       ▼                                          │                  │
│  ┌─────────────────────────────────────────────┘                  │
│  │                                                                  │
│  │  Handler Layer                                                   │
│  │  ┌──────────────────────────────────────────────────┐         │
│  │  │                                                    │         │
│  │  │  P0-6: State Machine     P0-4: OTP Verification  │         │
│  │  │       ↓                        ↓                  │         │
│  │  │  ┌─────────┐            ┌──────────┐            │         │
│  │  │  │  FSM    │            │   OTP    │            │         │
│  │  │  └─────────┘            └──────────┘            │         │
│  │  │                                                    │         │
│  │  │  P0-2: Saga Pattern                               │         │
│  │  │       ↓                                           │         │
│  │  │  ┌──────────────────────────────────┐           │         │
│  │  │  │     Saga Orchestrator            │           │         │
│  │  │  └──────────────────────────────────┘           │         │
│  │  │       │              │                           │         │
│  │  │       ▼              ▼                           │         │
│  │  │  P0-1: Lock     P0-7: Compensation              │         │
│  │  │  ┌─────────┐    ┌──────────────┐              │         │
│  │  │  │  Redis  │    │  Rollback    │              │         │
│  │  │  └─────────┘    └──────────────┘              │         │
│  │  │                                                    │         │
│  │  │  P0-5: GST Validation                            │         │
│  │  │       ↓                                           │         │
│  │  │  ┌─────────┐                                     │         │
│  │  │  │Validator│                                     │         │
│  │  │  └─────────┘                                     │         │
│  │  └──────────────────────────────────────────────────┘         │
│  │                                                                  │
│  └──────────────────────────────────────────────────────────────┘ │
│                                                                      │
└────────────────────────────────────────────────────────────────────┘
```

---

## P0 Component Interactions

### Request Flow with All P0 Fixes

```go
// internal/grpc/handlers/collaborator/integrated_handler.go

type IntegratedCollaboratorHandler struct {
    // P0-1: Distributed Locking
    lockManager *DistributedLock

    // P0-2: Saga Pattern
    sagaExecutor *SagaExecutor

    // P0-3: JWT Validation
    jwtValidator *JWTValidator

    // P0-4: OTP Service
    otpService *OTPService

    // P0-5: GST Validator
    gstValidator *GSTValidator

    // P0-6: State Machine
    stateMachine *StateMachine

    // Core services
    db          *gorm.DB
    aaaClient   aaa.Client
    logger      *zap.Logger
    metrics     *IntegratedMetrics
}

func (h *IntegratedCollaboratorHandler) CreateCollaborator(
    ctx context.Context,
    req *pb.CreateCollaboratorRequest,
) (*pb.CollaboratorResponse, error) {
    // P0-3: JWT is already validated by interceptor
    claims := getClaimsFromContext(ctx)

    // P0-5: Validate GST format first
    if err := h.gstValidator.ValidateFormat(req.GstNumber); err != nil {
        return nil, status.Errorf(codes.InvalidArgument,
            "invalid GST format: %v", err)
    }

    // Define integrated saga with all P0 components
    saga := &Saga{
        ID:   generateSagaID(),
        Name: "create_collaborator_integrated",
        Context: map[string]interface{}{
            "request": req,
            "fpo_id":  claims.FPOID,
            "user_id": claims.UserID,
        },
        Steps: []SagaStep{
            {
                Name: "acquire_gst_lock",
                Execute: func(ctx context.Context, data interface{}) error {
                    sagaCtx := data.(map[string]interface{})
                    req := sagaCtx["request"].(*pb.CreateCollaboratorRequest)

                    // P0-1: Acquire distributed lock
                    lock, err := h.lockManager.AcquireLock(ctx, LockOptions{
                        Key:        fmt.Sprintf("gst:lock:%s", req.GstNumber),
                        TTL:        30 * time.Second,
                        RetryDelay: 100 * time.Millisecond,
                        MaxRetries: 10,
                        Owner:      fmt.Sprintf("saga:%s", sagaCtx["saga_id"]),
                    })

                    if err != nil {
                        return fmt.Errorf("failed to acquire GST lock: %w", err)
                    }

                    sagaCtx["gst_lock"] = lock
                    return nil
                },
                Compensate: func(ctx context.Context, data interface{}) error {
                    sagaCtx := data.(map[string]interface{})
                    if lock, ok := sagaCtx["gst_lock"].(*Lock); ok {
                        return lock.Release(ctx)
                    }
                    return nil
                },
            },
            {
                Name: "validate_and_reserve_gst",
                Execute: func(ctx context.Context, data interface{}) error {
                    sagaCtx := data.(map[string]interface{})
                    req := sagaCtx["request"].(*pb.CreateCollaboratorRequest)

                    // P0-5: Server-side validation
                    if err := h.gstValidator.ValidateChecksum(req.GstNumber); err != nil {
                        return fmt.Errorf("GST checksum validation failed: %w", err)
                    }

                    // Check for existing GST
                    var existing models.Collaborator
                    err := h.db.Where("gst_number = ?", req.GstNumber).First(&existing).Error

                    if err == nil {
                        return ErrGSTAlreadyExists
                    }

                    if !errors.Is(err, gorm.ErrRecordNotFound) {
                        return err
                    }

                    return nil
                },
            },
            {
                Name: "create_addresses_in_aaa",
                Execute: func(ctx context.Context, data interface{}) error {
                    // P0-2: This step is part of the saga
                    // Implementation as shown in P0-2 architecture
                    return createAddressesInAAA(ctx, data)
                },
                Compensate: func(ctx context.Context, data interface{}) error {
                    // P0-7: Compensation for address rollback
                    return h.compensateAddresses(ctx, data)
                },
            },
            {
                Name: "save_collaborator_with_initial_status",
                Execute: func(ctx context.Context, data interface{}) error {
                    sagaCtx := data.(map[string]interface{})
                    req := sagaCtx["request"].(*pb.CreateCollaboratorRequest)

                    // P0-6: Use state machine for initial status
                    collaborator := &models.Collaborator{
                        Name:      req.Name,
                        GSTNumber: req.GstNumber,
                        Status:    models.CollaboratorStatusPending, // Always start as PENDING
                        // ... other fields
                    }

                    // Validate initial state
                    if err := h.stateMachine.ValidateInitialState(collaborator); err != nil {
                        return err
                    }

                    // Save to database
                    if err := h.db.Create(collaborator).Error; err != nil {
                        return fmt.Errorf("failed to save collaborator: %w", err)
                    }

                    sagaCtx["collaborator"] = collaborator
                    return nil
                },
                Compensate: func(ctx context.Context, data interface{}) error {
                    // Soft delete collaborator
                    return h.compensateCollaborator(ctx, data)
                },
            },
            {
                Name: "release_gst_lock",
                Execute: func(ctx context.Context, data interface{}) error {
                    sagaCtx := data.(map[string]interface{})
                    if lock, ok := sagaCtx["gst_lock"].(*Lock); ok {
                        return lock.Release(ctx)
                    }
                    return nil
                },
            },
        },
    }

    // P0-2: Execute saga with all integrated components
    if err := h.sagaExecutor.Execute(ctx, saga); err != nil {
        h.logger.Error("integrated saga failed",
            zap.String("saga_id", saga.ID),
            zap.Error(err))
        return nil, status.Errorf(codes.Internal,
            "failed to create collaborator: %v", err)
    }

    collaborator := saga.Context["collaborator"].(*models.Collaborator)

    return &pb.CollaboratorResponse{
        Collaborator: convertToProto(collaborator),
        Message:      "Collaborator created successfully",
    }, nil
}

func (h *IntegratedCollaboratorHandler) UpdateCollaborator(
    ctx context.Context,
    req *pb.UpdateCollaboratorRequest,
) (*pb.CollaboratorResponse, error) {
    // Load existing collaborator
    var collaborator models.Collaborator
    if err := h.db.First(&collaborator, req.Id).Error; err != nil {
        return nil, status.Errorf(codes.NotFound, "collaborator not found")
    }

    // P0-4: Check if banking details are being updated
    if req.FieldMask.Contains("bank_details") {
        if req.OtpCode == "" {
            // Send OTP and return
            if err := h.otpService.SendOTP(ctx, collaborator.Mobile); err != nil {
                return nil, status.Errorf(codes.Internal,
                    "failed to send OTP: %v", err)
            }

            return nil, status.Errorf(codes.FailedPrecondition,
                "OTP sent to registered mobile. Please provide OTP to update banking details")
        }

        // Verify OTP
        if err := h.otpService.VerifyOTP(ctx, collaborator.Mobile, req.OtpCode); err != nil {
            h.metrics.RecordOTPFailure(collaborator.ID)
            return nil, status.Errorf(codes.PermissionDenied,
                "invalid OTP: %v", err)
        }

        h.metrics.RecordOTPSuccess(collaborator.ID)

        // Log banking change for audit
        h.auditLogger.LogBankingChange(ctx, &collaborator, req.BankDetails)
    }

    // P0-6: Check if status is being updated
    if req.FieldMask.Contains("status") {
        newStatus := models.CollaboratorStatus(req.Status)
        user := getUserFromContext(ctx)

        // Use state machine for transition
        if err := h.stateMachine.ExecuteTransition(
            ctx, &collaborator, newStatus, user, req.Reason,
        ); err != nil {
            return nil, status.Errorf(codes.FailedPrecondition,
                "status transition failed: %v", err)
        }
    }

    // Update other fields...
    // Implementation continues...

    return &pb.CollaboratorResponse{
        Collaborator: convertToProto(&collaborator),
        Message:      "Collaborator updated successfully",
    }, nil
}
```

---

## Dependency Management

### Service Dependencies

```yaml
dependencies:
  redis:
    purpose: "Distributed locking (P0-1)"
    criticality: HIGH
    fallback: "In-memory locks (single instance only)"
    health_check:
      endpoint: "redis:6379/ping"
      timeout: 5s
      interval: 10s

  aaa_service:
    purpose: "Address management (P0-2, P0-7)"
    criticality: HIGH
    fallback: "Queue for retry"
    health_check:
      endpoint: "aaa:50051/health"
      timeout: 10s
      interval: 30s

  otp_service:
    purpose: "OTP verification (P0-4)"
    criticality: MEDIUM
    fallback: "Email verification"
    health_check:
      endpoint: "otp:8080/health"
      timeout: 5s
      interval: 30s

  database:
    purpose: "Primary data store"
    criticality: CRITICAL
    fallback: "None - fail fast"
    health_check:
      query: "SELECT 1"
      timeout: 3s
      interval: 10s
```

### Initialization Order

```go
// cmd/grpc-server/main.go

func initializeServices(config *Config) (*IntegratedServices, error) {
    // 1. Initialize database (critical)
    db, err := initDatabase(config.Database)
    if err != nil {
        return nil, fmt.Errorf("database init failed: %w", err)
    }

    // 2. Initialize Redis for locking (P0-1)
    redisClient, err := initRedis(config.Redis)
    if err != nil {
        log.Warn("Redis unavailable, using in-memory locks")
        redisClient = initInMemoryRedis()
    }

    lockManager := NewDistributedLock(redisClient)

    // 3. Initialize GST Validator (P0-5)
    gstValidator := NewGSTValidator()

    // 4. Initialize State Machine (P0-6)
    stateMachine := NewCollaboratorStateMachine(db)

    // 5. Initialize AAA Client with circuit breaker
    aaaClient, err := initAAAClient(config.AAA)
    if err != nil {
        return nil, fmt.Errorf("AAA client init failed: %w", err)
    }

    // 6. Initialize OTP Service (P0-4)
    otpService, err := initOTPService(config.OTP)
    if err != nil {
        log.Warn("OTP service unavailable, using fallback")
        otpService = NewFallbackOTPService()
    }

    // 7. Initialize Saga Executor (P0-2)
    sagaStorage := NewDBSagaStorage(db)
    sagaExecutor := NewSagaExecutor(sagaStorage)

    // 8. Initialize JWT Validator (P0-3)
    jwtValidator, err := NewJWTValidator(config.JWT)
    if err != nil {
        return nil, fmt.Errorf("JWT validator init failed: %w", err)
    }

    // 9. Start recovery services
    go startSagaRecovery(sagaExecutor, sagaStorage)
    go startLockMonitor(lockManager)

    return &IntegratedServices{
        DB:           db,
        LockManager:  lockManager,
        GSTValidator: gstValidator,
        StateMachine: stateMachine,
        AAAClient:    aaaClient,
        OTPService:   otpService,
        SagaExecutor: sagaExecutor,
        JWTValidator: jwtValidator,
    }, nil
}
```

---

## Testing Strategy for Integration

### Integration Test Suite

```go
// tests/integration/p0_integration_test.go

func TestP0_FullIntegration(t *testing.T) {
    // Setup all services
    services := setupIntegratedServices(t)
    defer services.Cleanup()

    handler := NewIntegratedCollaboratorHandler(services)

    t.Run("Create collaborator with all P0 checks", func(t *testing.T) {
        // Generate valid JWT (P0-3)
        token := generateTestJWT(t, "admin", "fpo123")
        ctx := contextWithJWT(token)

        req := &pb.CreateCollaboratorRequest{
            Name:      "Test Corp",
            GstNumber: "29ABCDE1234F1Z5", // Valid GST (P0-5)
            PrimaryAddress: &pb.Address{
                Street: "123 Test St",
                City:   "Bangalore",
            },
        }

        // Execute with all P0 components
        resp, err := handler.CreateCollaborator(ctx, req)

        assert.NoError(t, err)
        assert.NotNil(t, resp)
        assert.Equal(t, "PENDING", resp.Collaborator.Status) // P0-6

        // Verify no duplicate on concurrent request (P0-1)
        var wg sync.WaitGroup
        errors := make([]error, 10)

        for i := 0; i < 10; i++ {
            wg.Add(1)
            go func(idx int) {
                defer wg.Done()
                _, err := handler.CreateCollaborator(ctx, req)
                errors[idx] = err
            }(i)
        }

        wg.Wait()

        // All should fail with duplicate error
        for _, err := range errors {
            assert.Error(t, err)
            assert.Contains(t, err.Error(), "already exists")
        }
    })

    t.Run("Update banking with OTP verification", func(t *testing.T) {
        // Create collaborator first
        collaborator := createTestCollaborator(t, services)

        // Attempt update without OTP (P0-4)
        updateReq := &pb.UpdateCollaboratorRequest{
            Id: collaborator.ID,
            BankDetails: &pb.BankDetails{
                AccountNumber: "1234567890",
                IfscCode:     "HDFC0001234",
            },
            FieldMask: &fieldmaskpb.FieldMask{
                Paths: []string{"bank_details"},
            },
        }

        ctx := contextWithJWT(generateTestJWT(t, "admin", "fpo123"))

        // Should fail and send OTP
        _, err := handler.UpdateCollaborator(ctx, updateReq)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "OTP sent")

        // Get OTP from test service
        otp := services.OTPService.(*TestOTPService).GetLastOTP()

        // Retry with OTP
        updateReq.OtpCode = otp
        resp, err := handler.UpdateCollaborator(ctx, updateReq)

        assert.NoError(t, err)
        assert.Equal(t, "1234567890", resp.Collaborator.BankDetails.AccountNumber)
    })

    t.Run("State transitions with authorization", func(t *testing.T) {
        collaborator := createTestCollaborator(t, services)

        // Try invalid transition (P0-6)
        invalidCtx := contextWithJWT(generateTestJWT(t, "viewer", "fpo123"))

        statusReq := &pb.UpdateStatusRequest{
            CollaboratorId: collaborator.ID,
            NewStatus:      "ACTIVE", // Can't go directly to ACTIVE
            Reason:         "Test",
        }

        _, err := handler.UpdateStatus(invalidCtx, statusReq)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "invalid transition")

        // Valid transition path
        adminCtx := contextWithJWT(generateTestJWT(t, "admin", "fpo123"))

        // PENDING -> VERIFIED
        statusReq.NewStatus = "VERIFIED"
        _, err = handler.UpdateStatus(adminCtx, statusReq)
        assert.NoError(t, err)

        // VERIFIED -> ACTIVE
        statusReq.NewStatus = "ACTIVE"
        _, err = handler.UpdateStatus(adminCtx, statusReq)
        assert.NoError(t, err)
    })

    t.Run("Saga compensation on failure", func(t *testing.T) {
        // Simulate AAA service failure
        services.AAAClient.(*MockAAAClient).SetFailure(true)

        req := &pb.CreateCollaboratorRequest{
            Name:      "Fail Test",
            GstNumber: "27ABCDE5678F1Z5",
            PrimaryAddress: &pb.Address{
                Street: "456 Fail St",
            },
        }

        ctx := contextWithJWT(generateTestJWT(t, "admin", "fpo123"))

        // Should fail and compensate
        _, err := handler.CreateCollaborator(ctx, req)
        assert.Error(t, err)

        // Verify no collaborator created
        var count int64
        services.DB.Model(&models.Collaborator{}).
            Where("gst_number = ?", req.GstNumber).
            Count(&count)
        assert.Equal(t, int64(0), count)

        // Verify lock was released (P0-1)
        lockKey := fmt.Sprintf("gst:lock:%s", req.GstNumber)
        exists := services.Redis.Exists(ctx, lockKey).Val()
        assert.Equal(t, int64(0), exists)
    })
}
```

### Load Testing with P0 Components

```go
// tests/load/p0_load_test.go

func TestP0_LoadTest(t *testing.T) {
    services := setupIntegratedServices(t)
    defer services.Cleanup()

    handler := NewIntegratedCollaboratorHandler(services)

    // Generate unique GST numbers
    gstNumbers := generateUniqueGSTs(1000)

    t.Run("Concurrent creation with locking", func(t *testing.T) {
        var wg sync.WaitGroup
        success := atomic.Int32{}
        failed := atomic.Int32{}

        startTime := time.Now()

        for i := 0; i < 1000; i++ {
            wg.Add(1)
            go func(idx int) {
                defer wg.Done()

                ctx := contextWithJWT(generateTestJWT(t, "admin", "fpo123"))

                req := &pb.CreateCollaboratorRequest{
                    Name:      fmt.Sprintf("Test Corp %d", idx),
                    GstNumber: gstNumbers[idx],
                }

                _, err := handler.CreateCollaborator(ctx, req)
                if err == nil {
                    success.Add(1)
                } else {
                    failed.Add(1)
                }
            }(i)
        }

        wg.Wait()

        duration := time.Since(startTime)

        // All should succeed (unique GSTs)
        assert.Equal(t, int32(1000), success.Load())
        assert.Equal(t, int32(0), failed.Load())

        // Performance check
        assert.Less(t, duration, 30*time.Second, "Should complete within 30s")

        // Verify metrics
        lockWaitP99 := services.Metrics.GetP99("lock_acquisition_duration")
        assert.Less(t, lockWaitP99, 200*time.Millisecond)
    })
}
```

---

## Monitoring Dashboard

### Grafana Dashboard Configuration

```yaml
dashboard:
  title: "Collaborator Service - P0 Security Metrics"
  refresh: "10s"

  rows:
    - title: "P0-1: Distributed Locking"
      panels:
        - title: "Lock Acquisition Time"
          query: "histogram_quantile(0.99, lock_acquisition_duration)"
          unit: "ms"
        - title: "Lock Failures"
          query: "rate(lock_acquisition_failures[5m])"
        - title: "Active Locks"
          query: "redis_active_locks"

    - title: "P0-2: Saga Pattern"
      panels:
        - title: "Saga Success Rate"
          query: "rate(saga_completed) / rate(saga_started)"
          unit: "%"
        - title: "Compensation Rate"
          query: "rate(saga_compensated[5m])"
        - title: "Saga Duration P99"
          query: "histogram_quantile(0.99, saga_duration)"

    - title: "P0-3: JWT Validation"
      panels:
        - title: "Invalid JWT Attempts"
          query: "rate(jwt_validation_failures[5m])"
        - title: "Token Expiry Rate"
          query: "rate(jwt_expired[5m])"

    - title: "P0-4: OTP Verification"
      panels:
        - title: "OTP Success Rate"
          query: "rate(otp_verification_success) / rate(otp_verification_attempts)"
        - title: "Banking Updates"
          query: "rate(banking_updates_with_otp[1h])"

    - title: "P0-5: GST Validation"
      panels:
        - title: "Invalid GST Rate"
          query: "rate(gst_validation_failures[5m])"
        - title: "Duplicate GST Attempts"
          query: "rate(gst_duplicate_attempts[5m])"

    - title: "P0-6: State Machine"
      panels:
        - title: "Status Distribution"
          query: "collaborator_status_count"
          visualization: "pie"
        - title: "Invalid Transitions"
          query: "rate(state_transition_invalid[5m])"
        - title: "Unauthorized Attempts"
          query: "rate(state_transition_unauthorized[5m])"
```

---

## Deployment Strategy

### Phased Rollout Plan

```yaml
rollout_phases:
  phase_1_validation:
    duration: "2 days"
    traffic: "0%"
    mode: "shadow"
    description: "Run P0 checks in shadow mode, log results"
    success_criteria:
      - "No false positives in GST validation"
      - "Lock acquisition time < 200ms P99"
      - "Saga compensation working correctly"

  phase_2_canary:
    duration: "3 days"
    traffic: "1%"
    mode: "active"
    description: "Enable for 1% of traffic"
    success_criteria:
      - "No increase in error rate"
      - "Latency increase < 10%"
      - "No data inconsistencies"

  phase_3_gradual:
    duration: "5 days"
    traffic: "10% -> 50%"
    mode: "active"
    description: "Gradual rollout"
    success_criteria:
      - "Error rate < 0.1%"
      - "P99 latency < 200ms"
      - "No security incidents"

  phase_4_full:
    duration: "ongoing"
    traffic: "100%"
    mode: "active"
    description: "Full deployment"
    monitoring:
      - "24x7 alerting"
      - "Weekly security reviews"
      - "Monthly performance audits"
```

### Feature Flags

```go
// internal/config/features.go

type P0Features struct {
    DistributedLockingEnabled bool   `env:"P0_DISTRIBUTED_LOCKING" default:"false"`
    SagaPatternEnabled       bool   `env:"P0_SAGA_PATTERN" default:"false"`
    StrictJWTValidation     bool   `env:"P0_STRICT_JWT" default:"true"`
    OTPForBankingChanges    bool   `env:"P0_OTP_BANKING" default:"false"`
    ServerGSTValidation     bool   `env:"P0_GST_VALIDATION" default:"true"`
    StateMachineEnforcement bool   `env:"P0_STATE_MACHINE" default:"false"`

    // Gradual rollout
    RolloutPercentage       int    `env:"P0_ROLLOUT_PERCENTAGE" default:"0"`
    RolloutUserGroup       string `env:"P0_ROLLOUT_GROUP" default:""`
}

func (f *P0Features) IsEnabledForRequest(ctx context.Context) bool {
    // Check if globally enabled
    if f.RolloutPercentage >= 100 {
        return true
    }

    // Check user group
    user := getUserFromContext(ctx)
    if f.RolloutUserGroup != "" && user.Group == f.RolloutUserGroup {
        return true
    }

    // Check percentage rollout
    requestID := getRequestIDFromContext(ctx)
    hash := crc32.ChecksumIEEE([]byte(requestID))
    return (hash % 100) < uint32(f.RolloutPercentage)
}
```

---

## Rollback Strategy

### Automated Rollback Triggers

```yaml
rollback_triggers:
  - metric: "error_rate"
    threshold: "5%"
    window: "5m"
    action: "immediate_rollback"

  - metric: "p99_latency"
    threshold: "500ms"
    window: "10m"
    action: "gradual_rollback"

  - metric: "saga_compensation_failures"
    threshold: "10"
    window: "5m"
    action: "immediate_rollback"

  - metric: "data_inconsistency_detected"
    threshold: "1"
    window: "1m"
    action: "immediate_rollback"
```

### Rollback Procedure

```bash
#!/bin/bash
# rollback_p0.sh

# 1. Disable feature flags
kubectl set env deployment/collaborator-service \
  P0_DISTRIBUTED_LOCKING=false \
  P0_SAGA_PATTERN=false \
  P0_OTP_BANKING=false \
  P0_STATE_MACHINE=false

# 2. Scale down new version
kubectl scale deployment/collaborator-service-v2 --replicas=0

# 3. Scale up old version
kubectl scale deployment/collaborator-service-v1 --replicas=10

# 4. Clear Redis locks
redis-cli --scan --pattern "gst:lock:*" | xargs redis-cli DEL

# 5. Alert team
curl -X POST $SLACK_WEBHOOK -d '{"text": "P0 rollback executed"}'
```

---

## Performance Benchmarks

### Expected Performance with All P0 Fixes

| Operation | Without P0 | With P0 | Acceptable |
|-----------|-----------|---------|------------|
| Create Collaborator P50 | 20ms | 35ms | ✓ |
| Create Collaborator P99 | 50ms | 150ms | ✓ |
| Update Status P50 | 10ms | 15ms | ✓ |
| Update Status P99 | 30ms | 50ms | ✓ |
| Banking Update P50 | 15ms | 100ms | ✓ |
| Banking Update P99 | 40ms | 250ms | ✓ |
| Concurrent GST Check | Duplicates | No Duplicates | ✓ |
| Saga Compensation Success | N/A | 99.9% | ✓ |

---

## Critical Success Metrics

### Week 1 Goals (P0 Complete)
- ✅ Zero GST duplicates under load
- ✅ 100% saga compensations successful
- ✅ Zero unauthorized status transitions
- ✅ 100% banking changes require OTP
- ✅ All JWT signatures validated
- ✅ P99 latency < 200ms

### Production Readiness Criteria
- ✅ All P0 fixes implemented and tested
- ✅ Load tested with 1000+ concurrent users
- ✅ Security audit passed
- ✅ Monitoring dashboards configured
- ✅ Runbooks documented
- ✅ Rollback procedures tested

---

**Approval Required From**: CTO, Security Team, Platform Team
**Review Schedule**: Daily during rollout
**Go/No-Go Decision**: Week 1 completion
**Production Target**: Week 4