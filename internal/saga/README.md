# Saga Pattern Implementation for Distributed Transactions

## Overview

This package implements the Saga pattern for managing distributed transactions with automatic compensation. It ensures transactional integrity when operations span multiple services (e.g., E-commerce service and AAA service).

## Key Features

- **Automatic Compensation**: Failed steps trigger automatic rollback of completed steps
- **State Persistence**: Saga state persisted to database for crash recovery
- **Retry Logic**: Transient failures automatically retried with exponential backoff
- **Metrics**: Comprehensive metrics for monitoring saga execution
- **Type-Safe Steps**: Strongly-typed step definitions with execute and compensate functions

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                     Saga Orchestrator                        │
├──────────────────────────────────────────────────────────────┤
│  ┌───────────────────────────────────────────────────┐       │
│  │           Forward Transaction (T)                  │       │
│  │  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐         │       │
│  │  │Step 1│─>│Step 2│─>│Step 3│─>│Step 4│         │       │
│  │  └──────┘  └──────┘  └──────┘  └──────┘         │       │
│  └───────────────────────────────────────────────────┘       │
│                                                               │
│  ┌───────────────────────────────────────────────────┐       │
│  │         Compensation (on failure)                 │       │
│  │  ┌──────┐  ┌──────┐  ┌──────┐                    │       │
│  │  │Comp 3│<─│Comp 2│<─│Comp 1│                    │       │
│  │  └──────┘  └──────┘  └──────┘                    │       │
│  └───────────────────────────────────────────────────┘       │
└──────────────────────────────────────────────────────────────┘
```

## Usage Example: Collaborator Creation with Address Rollback (P0-7)

### Problem
When creating a collaborator:
1. Address created in AAA service (success)
2. Collaborator save to database fails
3. Result: **Orphaned address** in AAA

### Solution: Saga Pattern

```go
package handlers

import (
    "context"
    "kisanlink-ecom/internal/saga"
    "go.uber.org/zap"
)

// CreateCollaboratorWithSaga creates a collaborator using saga pattern
func (h *CollaboratorHandler) CreateCollaboratorWithSaga(
    ctx context.Context,
    req *CreateCollaboratorRequest,
) (*Collaborator, error) {

    // Create saga
    s := saga.NewSaga("create_collaborator", h.logger)

    // Step 1: Create address in AAA service
    s.AddStep(saga.NewSagaStep("create_address",
        // Execute: Create address
        func(ctx context.Context, data map[string]interface{}) error {
            addressReq := &AAAAddressRequest{
                Street:  req.Address.Street,
                City:    req.Address.City,
                State:   req.Address.State,
                Pincode: req.Address.Pincode,
            }

            addressResp, err := h.aaaClient.CreateAddress(ctx, addressReq)
            if err != nil {
                return fmt.Errorf("AAA address creation failed: %w", err)
            }

            // Store address ID for compensation
            data["address_id"] = addressResp.AddressID
            return nil
        },
        // Compensate: Delete address from AAA
        func(ctx context.Context, data map[string]interface{}) error {
            addressID, ok := data["address_id"].(string)
            if !ok {
                return nil // No address created yet
            }

            err := h.aaaClient.DeleteAddress(ctx, addressID)
            if err != nil {
                h.logger.Error("failed to compensate address deletion",
                    zap.String("address_id", addressID),
                    zap.Error(err))
                // Alert operations team for manual cleanup
                return err
            }

            h.logger.Info("address deleted via compensation",
                zap.String("address_id", addressID))
            return nil
        },
    ).WithRetry(3).WithTimeout(10 * time.Second))

    // Step 2: Save collaborator to database
    s.AddStep(saga.NewSagaStep("save_collaborator",
        // Execute: Save to DB
        func(ctx context.Context, data map[string]interface{}) error {
            collaborator := &Collaborator{
                UserID:    req.UserID,
                Username:  req.Username,
                Email:     req.Email,
                AddressID: data["address_id"].(string),
            }

            err := h.db.Create(collaborator).Error
            if err != nil {
                return fmt.Errorf("database save failed: %w", err)
            }

            data["collaborator"] = collaborator
            return nil
        },
        // Compensate: Delete collaborator (soft delete)
        func(ctx context.Context, data map[string]interface{}) error {
            collaborator, ok := data["collaborator"].(*Collaborator)
            if !ok {
                return nil
            }

            err := h.db.Delete(collaborator).Error
            if err != nil {
                h.logger.Error("failed to compensate collaborator deletion",
                    zap.Uint64("id", collaborator.ID),
                    zap.Error(err))
                return err
            }

            return nil
        },
    ).WithRetry(3))

    // Execute saga
    executor := saga.NewSagaExecutor(h.sagaStorage, h.logger, h.metrics)
    if err := executor.Execute(ctx, s); err != nil {
        return nil, fmt.Errorf("saga execution failed: %w", err)
    }

    // Success! Return collaborator
    collaborator := s.Context["collaborator"].(*Collaborator)
    return collaborator, nil
}
```

## Execution Scenarios

### Scenario 1: Happy Path
```
Step 1 (Create Address) → SUCCESS ✓
Step 2 (Save Collaborator) → SUCCESS ✓
Result: Saga COMPLETED ✓
```

### Scenario 2: Step 1 Fails
```
Step 1 (Create Address) → FAILURE ✗
No compensation needed (nothing to rollback)
Result: Saga COMPENSATED, no side effects
```

### Scenario 3: Step 2 Fails (Critical Case - P0-7)
```
Step 1 (Create Address) → SUCCESS ✓
Step 2 (Save Collaborator) → FAILURE ✗
Compensation triggered:
  → Compensate Step 1 (Delete Address) → SUCCESS ✓
Result: Saga COMPENSATED, no orphaned address ✓
```

### Scenario 4: Compensation Fails (Alert Required)
```
Step 1 (Create Address) → SUCCESS ✓
Step 2 (Save Collaborator) → FAILURE ✗
Compensation triggered:
  → Compensate Step 1 (Delete Address) → FAILURE ✗
Result: Saga FAILED, manual cleanup required
Action: Alert operations team, log critical error
```

## State Machine Integration (P0-6)

Saga pattern can be combined with state machine for status transitions:

```go
// Step: Update status with state machine validation
s.AddStep(saga.NewSagaStep("update_status",
    func(ctx context.Context, data map[string]interface{}) error {
        collaborator := data["collaborator"].(*Collaborator)

        // Use state machine to validate transition
        err := h.stateMachine.ExecuteTransition(
            ctx,
            collaborator.ID,
            collaborator.Status,
            StatusActive,
            userID,
            []string{"ADMIN"},
            "Activated via onboarding",
        )

        if err != nil {
            return fmt.Errorf("invalid status transition: %w", err)
        }

        data["previous_status"] = collaborator.Status
        data["new_status"] = StatusActive
        return nil
    },
    func(ctx context.Context, data map[string]interface{}) error {
        // Compensate: Revert status
        collaborator := data["collaborator"].(*Collaborator)
        previousStatus := data["previous_status"].(Status)

        return h.stateMachine.ExecuteTransition(
            ctx,
            collaborator.ID,
            StatusActive,
            previousStatus,
            "system",
            []string{"ADMIN"},
            "Reverted due to saga compensation",
        )
    },
))
```

## OTP Verification Integration (P0-4)

Banking changes require OTP verification within saga:

```go
// Step: Validate OTP before updating banking details
s.AddStep(saga.NewSagaStep("validate_otp",
    func(ctx context.Context, data map[string]interface{}) error {
        otpToken := data["otp_token"].(string)
        userID := data["user_id"].(string)

        // Validate OTP (P0-4)
        err := h.otpService.Validate(ctx, userID, "BANK_UPDATE", otpToken)
        if err != nil {
            return fmt.Errorf("OTP validation failed: %w", err)
        }

        data["otp_validated"] = true
        return nil
    },
    nil, // No compensation for validation
))

// Step: Update banking details
s.AddStep(saga.NewSagaStep("update_banking",
    func(ctx context.Context, data map[string]interface{}) error {
        collaborator := data["collaborator"].(*Collaborator)
        bankDetails := data["bank_details"].(BankDetails)

        // Store old values for compensation
        data["old_bank_details"] = BankDetails{
            AccountNumber: collaborator.BankAccountNumber,
            IFSCCode:      collaborator.IFSCCode,
            BankName:      collaborator.BankName,
        }

        // Update banking details
        collaborator.BankAccountNumber = bankDetails.AccountNumber
        collaborator.IFSCCode = bankDetails.IFSCCode
        collaborator.BankName = bankDetails.BankName

        // Apply 48-hour payment hold (P0-4 requirement)
        collaborator.PaymentHoldUntil = time.Now().Add(48 * time.Hour)

        return h.db.Save(collaborator).Error
    },
    func(ctx context.Context, data map[string]interface{}) error {
        // Compensate: Restore old banking details
        collaborator := data["collaborator"].(*Collaborator)
        oldDetails := data["old_bank_details"].(BankDetails)

        collaborator.BankAccountNumber = oldDetails.AccountNumber
        collaborator.IFSCCode = oldDetails.IFSCCode
        collaborator.BankName = oldDetails.BankName
        collaborator.PaymentHoldUntil = nil

        return h.db.Save(collaborator).Error
    },
))
```

## Monitoring and Metrics

```go
// Get saga statistics
stats := metrics.GetSagaStats("create_collaborator")

fmt.Printf("Started: %d\n", stats.Started)
fmt.Printf("Completed: %d\n", stats.Completed)
fmt.Printf("Failed: %d\n", stats.Failed)
fmt.Printf("Compensated: %d\n", stats.Compensated)
fmt.Printf("Success Rate: %.2f%%\n", stats.SuccessRate() * 100)
fmt.Printf("Avg Duration: %v\n", stats.AverageDuration())
```

## Database Schema

```sql
CREATE TABLE saga_states (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    state VARCHAR(50) NOT NULL,
    context JSONB,
    completed_steps JSONB,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    INDEX idx_state (state),
    INDEX idx_name (name),
    INDEX idx_created_at (created_at)
);
```

## Testing

Run all saga tests:
```bash
go test ./internal/saga/... -v
```

## Security Considerations (P0-3)

All saga operations inherit JWT validation with:
- Cryptographic signature verification
- Issuer/audience validation
- Expiration checking
- JTI replay prevention (via JTI cache)

## Performance Targets

- **P95 Latency**: < 200ms for 2-step sagas
- **P99 Latency**: < 500ms for 4-step sagas
- **Success Rate**: > 99.5% (excluding business logic failures)
- **Compensation Rate**: < 0.5%

## References

- P0-2: Saga Pattern Architecture
- P0-3: JWT Signature Verification
- P0-4: OTP Verification for Banking Changes
- P0-6: State Machine for Status Transitions
- P0-7: Address Rollback on Failure (this implementation)
