# Collaborator gRPC Service - Implementation Tasks

**Version**: 1.0
**Date**: 2025-11-12
**Status**: NOT PRODUCTION READY - 7 CRITICAL BLOCKERS

---

## Priority Structure

- **P0 (CRITICAL)**: Security & data integrity blockers - MUST fix before deployment
- **P1 (HIGH)**: Core functionality required for MVP
- **P2 (MEDIUM)**: Important features (not blocking)
- **P3 (LOW)**: Nice-to-have improvements

---

## P0: CRITICAL SECURITY & INTEGRITY FIXES (Week 1)

### P0-1: Implement Distributed Locking for GST Deduplication
**Priority**: CRITICAL
**Effort**: 8 hours
**Assigned**: @agent-sde-backend-engineer
**Review**: @agent-sde3-backend-architect

**Problem**: Multiple FPOs can create duplicate master collaborators when registering same GST simultaneously.

**Files to Create**:
- `internal/services/gst/distributed_lock.go` - Redis-based distributed lock
- `internal/services/gst/gst_service.go` - Update CheckGSTExists with locking
- `tests/collaborator_grpc/gst_race_condition_test.go` - Concurrency tests

**Implementation**:
1. Redis-based distributed lock with SETNX pattern
2. Lock TTL 30 seconds with automatic renewal
3. Timeout handling for lock acquisition
4. Metrics for lock wait times and failures
5. Test with 100+ concurrent GST registrations

**Acceptance Criteria**:
- ✅ No duplicate master records under concurrent load
- ✅ Lock timeout prevents indefinite waits (max 30s)
- ✅ Proper cleanup of expired locks
- ✅ All tests pass with `go test -race`

---

### P0-2: Implement Saga Pattern for AAA-Collaborator Transaction Integrity
**Priority**: CRITICAL
**Effort**: 10 hours
**Assigned**: @agent-sde-backend-engineer
**Review**: @agent-sde3-backend-architect

**Problem**: No rollback when address creation succeeds but collaborator save fails, causing orphaned addresses.

**Files to Create**:
- `internal/saga/saga.go` - Saga pattern implementation
- `internal/saga/executor.go` - Execution engine
- `internal/saga/compensation.go` - Compensation handlers
- `internal/grpc/handlers/collaborator/create.go` - Integrate saga
- `internal/grpc/handlers/collaborator/update.go` - Integrate saga

**Implementation**:
1. Saga with forward/compensation steps
2. Step registration with compensation handlers
3. Transaction executor with rollback on failure
4. Address creation step with deletion compensation
5. Collaborator save step with deletion compensation
6. Error recovery with logging

**Acceptance Criteria**:
- ✅ Partial failure triggers compensation
- ✅ No orphaned addresses in AAA
- ✅ All failures logged with trace IDs
- ✅ Test: AAA succeeds, DB fails → address deleted

---

### P0-3: Strengthen JWT Validation with Signature Verification
**Priority**: CRITICAL
**Effort**: 6 hours
**Assigned**: @agent-sde-backend-engineer

**Problem**: JWT token manipulation allows cross-FPO data access.

**Files to Create**:
- `internal/auth/jwt_validator.go` - Enhanced JWT validation
- `internal/auth/claims.go` - Claims structure
- `internal/grpc/interceptors/auth.go` - Update auth interceptor

**Implementation**:
1. RSA signature verification with public key
2. Algorithm validation (RS256 only)
3. Issuer validation (trusted authority)
4. Audience validation (this service)
5. FPO ID validation against database
6. Expiration validation
7. JTI validation (prevent replay attacks)

**Acceptance Criteria**:
- ✅ Invalid signatures rejected
- ✅ Modified claims detected
- ✅ Token replay prevented
- ✅ Can't access different FPO's data with modified token

---

### P0-4: Implement OTP Verification for Banking Detail Changes
**Priority**: CRITICAL
**Effort**: 8 hours
**Assigned**: @agent-sde-backend-engineer

**Problem**: Banking details can be changed without verification, enabling fraud.

**Files to Create**:
- `internal/services/otp/otp_service.go` - OTP service wrapper
- `internal/grpc/handlers/collaborator/update.go` - Banking update handler
- `proto/collaborator/v1/collaborator_messages.proto` - Add OTP field

**Implementation**:
1. OTP generation and validation
2. Banking field detection in update requests
3. OTP requirement for banking changes
4. Rate limiting (5 attempts per hour)
5. Audit log for banking changes
6. 48-hour payment hold after banking change
7. Email/SMS notifications

**Acceptance Criteria**:
- ✅ Banking change requires valid OTP
- ✅ Invalid OTP rejected with rate limiting
- ✅ Changes logged with user, IP, timestamp
- ✅ Payment hold enforced for 48 hours

---

### P0-5: Implement Server-Side GST Format Validation
**Priority**: CRITICAL
**Effort**: 4 hours
**Assigned**: @agent-sde-backend-engineer

**Problem**: Client-side validation can be bypassed.

**Files to Create**:
- `internal/services/gst/validator.go` - GST validation
- `internal/services/gst/gst_service.go` - Use validator

**Implementation**:
1. GST checksum validation (Luhn variant)
2. State code validation (01-37)
3. PAN format validation (5 letters + 4 digits + 1 letter)
4. Entity number validation (1-9)
5. Normalized storage (uppercase, trimmed)

**Acceptance Criteria**:
- ✅ Invalid checksum rejected
- ✅ Invalid state code rejected
- ✅ Invalid PAN rejected
- ✅ Valid GST accepted

---

### P0-6: Implement State Machine for Collaborator Status Transitions
**Priority**: CRITICAL
**Effort**: 6 hours
**Assigned**: @agent-sde-backend-engineer
**Review**: @agent-sde3-backend-architect

**Problem**: Status transitions can be bypassed via direct updates.

**Files to Create**:
- `internal/domain/collaborator/state_machine.go` - Transition rules
- `internal/grpc/handlers/collaborator/update.go` - Enforce transitions
- `internal/grpc/handlers/collaborator/deactivate.go` - Enforce transitions

**Implementation**:
1. Valid status transition graph (PENDING → VERIFIED → ACTIVE)
2. State machine with allowed transitions
3. Transition requirements (documents for VERIFIED)
4. Manager approval for status changes
5. Validation on DeactivateCollaborator

**Acceptance Criteria**:
- ✅ Only valid transitions allowed
- ✅ Direct VERIFIED update rejected
- ✅ Verification documents required
- ✅ Manager approval logged

---

### P0-7: Implement Address Rollback on Collaborator Save Failure
**Priority**: CRITICAL
**Effort**: 6 hours (integrated with P0-2)
**Assigned**: @agent-sde-backend-engineer
**Dependencies**: P0-2 (Saga Pattern)

**Problem**: Address created but collaborator save fails → orphaned address.

**Files to Modify**:
- `internal/saga/compensation.go` - Address deletion compensation
- `internal/grpc/handlers/collaborator/create.go` - Rollback integration

**Implementation**:
1. Address deletion as compensation step
2. Handle AAA service failures gracefully
3. Retry address deletion with exponential backoff
4. Log failed compensations for manual cleanup
5. Alert on compensation failures

**Acceptance Criteria**:
- ✅ Address deleted when collaborator save fails
- ✅ Multiple addresses all rolled back
- ✅ Failures logged and alerted
- ✅ No orphaned addresses

---

## P1: HIGH PRIORITY - CORE MVP FUNCTIONALITY (Weeks 2-3)

### P1-1: Set Up gRPC Server Infrastructure
**Priority**: HIGH
**Effort**: 6 hours
**Assigned**: @agent-sde-backend-engineer

**Files to Create**:
- `cmd/grpc-server/main.go` - Server entry point
- `internal/grpc/server.go` - Server initialization
- `internal/config/grpc.go` - gRPC configuration

**Acceptance Criteria**:
- ✅ Server starts on configured port
- ✅ Graceful shutdown within 30s
- ✅ Health check endpoint working
- ✅ Reflection enabled for grpcurl

---

### P1-2: Implement Complete 10-Layer Interceptor Chain
**Priority**: HIGH
**Effort**: 16 hours
**Assigned**: @agent-sde-backend-engineer

**Files to Create**:
- `internal/grpc/interceptors/recovery.go` - Panic recovery
- `internal/grpc/interceptors/request_id.go` - Request ID generation
- `internal/grpc/interceptors/logging.go` - Request logging
- `internal/grpc/interceptors/metrics.go` - Prometheus metrics
- `internal/grpc/interceptors/tracing.go` - OpenTelemetry tracing
- `internal/grpc/interceptors/rate_limit.go` - Rate limiting
- `internal/grpc/interceptors/auth.go` - JWT validation
- `internal/grpc/interceptors/authorization.go` - RBAC checks
- `internal/grpc/interceptors/validation.go` - Proto validation
- `internal/grpc/interceptors/audit.go` - Audit logging

**Layer Order**: Recovery → Request ID → Logging → Metrics → Tracing → Rate Limit → Auth → Authorization → Validation → Audit

**Acceptance Criteria**:
- ✅ All 10 layers execute in order
- ✅ Errors handled properly at each layer
- ✅ Metrics collected for all operations
- ✅ Audit trail complete for mutations

---

### P1-3: Implement AAA Connection Pool with Circuit Breaker
**Priority**: HIGH
**Effort**: 8 hours
**Assigned**: @agent-sde-backend-engineer

**Files to Create**:
- `internal/aaa/pool.go` - Connection pool
- `internal/aaa/circuit_breaker.go` - Circuit breaker
- `internal/aaa/client.go` - AAA client wrapper

**Acceptance Criteria**:
- ✅ Pool creates multiple connections
- ✅ Round-robin load balancing
- ✅ Circuit breaker opens at 60% failure
- ✅ Retries with exponential backoff

---

### P1-4: Implement CreateCollaborator Handler
**Priority**: HIGH
**Effort**: 8 hours
**Assigned**: @agent-sde-backend-engineer
**Validation**: @agent-business-logic-tester
**Dependencies**: P0-1, P0-2, P1-3

**Files to Create**:
- `internal/grpc/handlers/collaborator/create.go`
- `internal/grpc/handlers/collaborator/converters.go`

**Acceptance Criteria**:
- ✅ GST deduplication prevents duplicates
- ✅ Addresses created in AAA via saga
- ✅ Collaborator created in database
- ✅ Audit logged

---

### P1-5: Implement UpdateCollaborator Handler
**Priority**: HIGH
**Effort**: 8 hours
**Assigned**: @agent-sde-backend-engineer
**Dependencies**: P0-4, P0-6

**Files to Create**:
- `internal/grpc/handlers/collaborator/update.go`
- `internal/grpc/handlers/collaborator/field_mask.go`

**Acceptance Criteria**:
- ✅ Field mask processed correctly
- ✅ Banking changes require OTP
- ✅ Status transitions validated

---

### P1-6: Implement GetCollaborator Handler
**Priority**: HIGH
**Effort**: 4 hours
**Assigned**: @agent-sde-backend-engineer

**Files to Create**:
- `internal/grpc/handlers/collaborator/get.go`

**Acceptance Criteria**:
- ✅ Returns correct collaborator
- ✅ Address expansion works
- ✅ Sensitive fields filtered
- ✅ Permissions enforced

---

### P1-7: Implement DeactivateCollaborator Handler
**Priority**: HIGH
**Effort**: 6 hours
**Assigned**: @agent-sde-backend-engineer
**Validation**: @agent-business-logic-tester
**Dependencies**: P0-6, P1-10

**Files to Create**:
- `internal/grpc/handlers/collaborator/deactivate.go`
- `internal/services/collaborator/deactivate_service.go`

**Acceptance Criteria**:
- ✅ Permissions enforced (admin/manager only)
- ✅ Pending transactions prevent deactivation
- ✅ Active orders prevent deactivation
- ✅ Status transition validated

---

### P1-8: Fix Circuit Breaker Reset Logic
**Priority**: HIGH
**Effort**: 4 hours
**Assigned**: @agent-sde-backend-engineer
**Dependencies**: P1-3

**Files to Modify**:
- `internal/aaa/circuit_breaker.go`

**Acceptance Criteria**:
- ✅ Circuit opens after failures
- ✅ Circuit tries recovery after timeout
- ✅ Circuit closes on successful requests

---

### P1-9: Add Retry Logic for AAA Transient Failures
**Priority**: HIGH
**Effort**: 4 hours
**Assigned**: @agent-sde-backend-engineer
**Dependencies**: P1-3

**Files to Create**:
- `internal/aaa/retry.go`

**Acceptance Criteria**:
- ✅ Transient failures retried with backoff
- ✅ Non-retryable errors fail fast
- ✅ Max 3 attempts enforced

---

### P1-10: Add Validation for Active Orders During Deactivation
**Priority**: HIGH
**Effort**: 4 hours
**Assigned**: @agent-business-logic-tester

**Files to Modify**:
- `internal/services/collaborator/deactivate_service.go`

**Acceptance Criteria**:
- ✅ Active orders prevent deactivation
- ✅ Error includes count and value
- ✅ Completed/cancelled orders don't block

---

### P1-11: Achieve 90% Test Coverage
**Priority**: HIGH
**Effort**: 16 hours
**Assigned**: @agent-business-logic-tester
**Dependencies**: All handlers implemented

**Files to Create/Update**:
- All test files updated for complete coverage
- `tests/collaborator_grpc/coverage_report.md`

**Acceptance Criteria**:
- ✅ Overall coverage ≥ 90%
- ✅ Critical path coverage ≥ 95%
- ✅ Branch coverage ≥ 80%
- ✅ All error paths tested

---

## P2: MEDIUM PRIORITY - DEPLOYMENT & OBSERVABILITY (Week 4)

### P2-1: Docker Configuration
**Effort**: 4 hours
**Files**: `Dockerfile`, `docker-compose.yml`

### P2-2: Kubernetes Manifests
**Effort**: 6 hours
**Files**: `k8s/*.yaml`

### P2-3: Prometheus Monitoring Setup
**Effort**: 6 hours
**Files**: `config/prometheus.yml`

### P2-4: Performance Testing & Optimization
**Effort**: 8 hours
**Files**: `tests/load/*`

### P2-5: Load Testing (1000+ concurrent connections)
**Effort**: 6 hours
**Files**: `tests/load/load_test.go`

### P2-6: Documentation Updates
**Effort**: 8 hours
**Files**: `/docs/collaborator-grpc-service.md`

---

## P3: LOW PRIORITY - ADVANCED FEATURES (Post-MVP)

- OpenTelemetry Distributed Tracing
- Advanced Caching Strategies
- Bulk Operations Support
- GraphQL Federation Support

---

## Implementation Sequence

### Week 1: Critical Security Fixes (P0)
All P0 tasks (42 hours total)

### Week 2-3: Core Implementation (P1)
P1-1 through P1-11 (94 hours total)

### Week 4: Deployment (P2)
P2-1 through P2-6 (38 hours total)

---

## Critical Success Metrics

- ✅ GST deduplication prevents duplicates
- ✅ AAA consistency maintained
- ✅ JWT signature verified
- ✅ Banking changes protected with OTP
- ✅ Status transitions enforced
- ✅ All 4 endpoints functional
- ✅ P99 latency < 100ms (reads), < 200ms (writes)
- ✅ Support 1000+ concurrent connections
- ✅ 90% test coverage
- ✅ Security audit passed

---

## Blockers & Dependencies

**Must Be Available**:
- Redis for distributed locking (P0-1)
- AAA service running (P1-3)
- External OTP service (P0-4)
- Order service integration (P1-10)
- Database configured (all)

---

## Commit Strategy

Each task should result in a meaningful, independently committable changeset:

```
feat(grpc): implement distributed locking for GST deduplication

- Add Redis distributed lock implementation
- Integrate with GST deduplication service
- Add comprehensive concurrency tests
- Add metrics for lock operations

Resolves: P0-1
```

Every commit:
- ✅ Passes all tests (`make test`)
- ✅ Passes linting (`make lint`)
- ✅ Doesn't skip pre-commit hooks
- ✅ Has clear, descriptive message
- ✅ No "Claude Code" attribution
- ✅ References task ID

---

**Generated**: 2025-11-12
**Status**: Ready for Implementation
**Next Action**: Launch P0 implementation with @agent-sde-backend-engineer
