# Implementation Plan

## EPIC A — Domain & Data Modeling (GORM-first)

- [x] A1. Create canonical domain model structure
  - Define core domain entities: Catalog, Category, Variant, Pricing, Availability, Attributes, Media, Vendor/FPO, SLA, PublishState
  - Implement base model with tenant isolation and audit fields
  - Create type-specific attribute schemas for Products, Services, Labour, Contracts
  - **Priority**: P0 | **Estimate**: 5 SP | **Owner**: TBD | **Labels**: [backend], [domain]
  - **Dependencies**: None
  - **Requirements**: 1.1, 1.5

- [x] A2. Implement GORM models with proper tags and constraints
  - Create GORM model structs with table names, indexes, unique constraints
  - Implement soft-delete support using gorm.DeletedAt
  - Add tenant_id/org_id fields
  - **Priority**: P0 | **Estimate**: 8 SP | **Owner**: TBD | **Labels**: [backend], [database]
  - **Dependencies**: A1
  - **Requirements**: 1.2

- [x] A3. Create AutoMigrate bootstrap with custom schema tracking
  - Implement GORM AutoMigrate sequence with idempotent migrator
  - Create custom schema_migrations table for semantic version tracking
  - Build migration runner with rollback safety checks
  - **Priority**: P0 | **Estimate**: 5 SP | **Owner**: TBD | **Labels**: [backend], [database], [ops]
  - **Dependencies**: A2
  - **Requirements**: 1.3

- [x] A4. Optimize database indexes for query performance
  - Create composite indexes for (tenant_id, type, status, updated_at DESC)
  - Implement trigram/FTS indexes for name and description fields
  - Add JSONB indexes for attributes and array indexes for tags
  - **Priority**: P0 | **Estimate**: 3 SP | **Owner**: TBD | **Labels**: [backend], [database], [performance]
  - **Dependencies**: A2
  - **Requirements**: 1.4

- [x] A5. Implement type-specific attribute validation policies
  - Create validation rules for Contract attributes (term/duration)
  - Implement Labour attribute validation (skill, unit rate)
  - Add Service attribute validation (SLA fields)
  - Create Product attribute validation (SKU/variants)
  - **Priority**: P0 | **Estimate**: 6 SP | **Owner**: TBD | **Labels**: [backend], [validation]
  - **Dependencies**: A1, A2
  - **Requirements**: 1.5

## EPIC B — HTTP API & Validations (no gRPC server)

- [x] B1. Create OpenAPI v3 specification
  - Define CRUD and other workflow API handlers and routes to perform all operations
  - Include publish/unpublish and price update operations
  - Generate Swagger documentation
  - **Priority**: P0 | **Estimate**: 6 SP | **Owner**: TBD | **Labels**: [backend], [api], [dx]
  - **Dependencies**: A1
  - **Requirements**: 2.1

- [x] B2. Implement request validation with go-playground/validator
  - Create custom validation rules for catalog-specific fields
  - Implement rate unit validation, contract date validation, SKU format validation
  - Add cross-field validation for complex business rules
  - Add admin approval workflows for complex business rules
  - **Priority**: P0 | **Estimate**: 5 SP | **Owner**: TBD | **Labels**: [backend], [validation]
  - **Dependencies**: A5, B1
  - **Requirements**: 2.2

- [ ] B3. Design structured error handling system
  - Create/Reuse error model with code, message, details, request_id and make it robust
  - Map errors to appropriate HTTP status codes
  - Implement i18n-ready error codes
  - Ensure no database internals are leaked in responses
  - **Priority**: P0 | **Estimate**: 4 SP | **Owner**: TBD | **Labels**: [backend], [error-handling]
  - **Dependencies**: None
  - **Requirements**: 2.3

- [x] B4. Implement idempotency and optimistic locking
  - Add Idempotency-Key header support for POST/PUT operations
  - Implement optimistic locking using updated_at/version fields
  - Create idempotency key storage and validation logic
  - **Priority**: P1 | **Estimate**: 4 SP | **Owner**: TBD | **Labels**: [backend], [reliability]
  - **Dependencies**: A2, B3
  - **Requirements**: 2.4

- [x] B5. Add ETag support for caching validation
  - Implement ETag generation for GET detail responses
  - Add If-None-Match header processing
  - Create cache validation middleware
  - **Priority**: P1 | **Estimate**: 3 SP | **Owner**: TBD | **Labels**: [backend], [caching]
  - **Dependencies**: B1
  - **Requirements**: 2.5

## EPIC C — AAA gRPC Client + AuthZ

- [x] C1. Integrate AAA service gRPC client
  - Import AAA protobufs from github.com/Kisanlink/pkg/proto
  - Generate gRPC client code using buf
  - Configure TLS and mTLS for secure communication
  - **Priority**: P0 | **Estimate**: 4 SP | **Owner**: TBD | **Labels**: [backend], [security], [integration]
  - **Dependencies**: None
  - **Requirements**: 3.1

- [ ] C2. Implement authentication and authorization middleware
  - Create JWT verification middleware (local validation)
  - Implement AAA Authorize calls for resource-level permissions
  - Add permission caching with short TTL
  - **Priority**: P0 | **Estimate**: 6 SP | **Owner**: TBD | **Labels**: [backend], [security], [middleware]
  - **Dependencies**: C1
  - **Requirements**: 3.2

- [ ] C3. Implement tenant extraction and enforcement
  - Extract tenant information from JWT claims and headers
  - Add automatic tenant_id filtering to all repository queries
  - Create tenant-aware middleware for request processing
  - **Priority**: P0 | **Estimate**: 5 SP | **Owner**: TBD | **Labels**: [backend], [security], [multi-tenancy]
  - **Dependencies**: C2, A2
  - **Requirements**: 3.3

- [ ] C4. Create comprehensive audit logging system
  - Implement append-only audit table for all operations
  - Log who/what/when/before/after for all changes
  - Add structured logging with proper context
  - **Priority**: P0 | **Estimate**: 4 SP | **Owner**: TBD | **Labels**: [backend], [security], [audit]
  - **Dependencies**: A3, C3
  - **Requirements**: 3.4

## EPIC D — Persistence & Integrity (GORM)

- [ ] D1. Create repository layer with context-aware queries
  - Implement repository interfaces for all domain entities
  - Add strict SELECT column specification to prevent over-fetching
  - Implement Preload rules to prevent N+1 query problems
  - **Priority**: P0 | **Estimate**: 8 SP | **Owner**: TBD | **Labels**: [backend], [repository]
  - **Dependencies**: A2, C3
  - **Requirements**: 4.1

- [ ] D2. Implement transaction management and retry logic
  - Create transaction wrapper for multi-entity operations
  - Implement compensating logic for failed transactions
  - Add idempotent retry mechanisms with exponential backoff
  - **Priority**: P0 | **Estimate**: 5 SP | **Owner**: TBD | **Labels**: [backend], [reliability], [database]
  - **Dependencies**: D1
  - **Requirements**: 4.2

- [ ] D3. Implement data retention and protection policies
  - Create soft-delete policies with configurable retention periods
  - Implement protected delete checks for referenced entities
  - Document PII mapping and data retention requirements
  - **Priority**: P1 | **Estimate**: 4 SP | **Owner**: TBD | **Labels**: [backend], [compliance], [data-protection]
  - **Dependencies**: A2, D1
  - **Requirements**: 4.3, 4.4

## EPIC E — Performance & Caching

- [ ] E1. Implement Redis caching with tenant namespacing
  - Create cache manager with per-tenant key namespaces
  - Implement TTL policies and explicit cache invalidation on writes
  - Add cache warming strategies for hot data
  - **Priority**: P1 | **Estimate**: 6 SP | **Owner**: TBD | **Labels**: [backend], [caching], [performance]
  - **Dependencies**: D1
  - **Requirements**: 5.1

- [ ] E2. Optimize database queries and implement GORM scopes
  - Use EXPLAIN ANALYZE on realistic datasets for query optimization
  - Create GORM scopes for common filter patterns
  - Implement query result caching for expensive operations
  - **Priority**: P1 | **Estimate**: 5 SP | **Owner**: TBD | **Labels**: [backend], [database], [performance]
  - **Dependencies**: A4, D1
  - **Requirements**: 5.2

- [ ] E3. Create k6 load testing suite with performance targets
  - Develop k6 scripts for all major API endpoints
  - Set P50/P95 latency targets and error budget definitions
  - Create realistic dataset seeding scripts for load testing
  - **Priority**: P1 | **Estimate**: 4 SP | **Owner**: TBD | **Labels**: [testing], [performance]
  - **Dependencies**: B1, E1
  - **Requirements**: 5.3, 5.4

## EPIC F — Search & Discovery (phaseable)

- [ ] F1. Implement PostgreSQL native search with faceted filtering
  - Create ILIKE/trigram search functionality
  - Implement filter facets for type, category, tags, price bands, status, vendor
  - Add search result ranking and relevance scoring
  - **Priority**: P1 | **Estimate**: 6 SP | **Owner**: TBD | **Labels**: [backend], [search]
  - **Dependencies**: A4, D1
  - **Requirements**: 6.1, 6.2

- [ ] F2. Create search abstraction for future OpenSearch integration
  - Design search interface abstraction layer
  - Create backfill and index job framework
  - Implement search result caching strategies
  - **Priority**: P2 | **Estimate**: 4 SP | **Owner**: TBD | **Labels**: [backend], [search], [architecture]
  - **Dependencies**: F1
  - **Requirements**: 6.3, 6.4

## EPIC G — Observability & Operations

- [ ] G1. Implement structured logging with zap
  - Create logging interface with request_id, tenant_id, user_id context
  - Implement PII redaction policies and log sampling
  - Add structured logging for all major operations
  - **Priority**: P0 | **Estimate**: 4 SP | **Owner**: TBD | **Labels**: [backend], [observability]
  - **Dependencies**: C3
  - **Requirements**: 7.1

- [ ] G2. Add Prometheus metrics collection
  - Implement RED metrics (Rate, Errors, Duration) for all endpoints
  - Add USE metrics for system resources
  - Create custom business metrics for catalog operations
  - **Priority**: P0 | **Estimate**: 5 SP | **Owner**: TBD | **Labels**: [backend], [observability], [metrics]
  - **Dependencies**: B1
  - **Requirements**: 7.2

- [ ] G3. Implement OpenTelemetry distributed tracing
  - Add tracing for HTTP→Service→Repository→Database→Cache flows
  - Configure 1% baseline sampling with 100% error sampling
  - Implement trace context propagation
  - **Priority**: P1 | **Estimate**: 4 SP | **Owner**: TBD | **Labels**: [backend], [observability], [tracing]
  - **Dependencies**: G1, G2
  - **Requirements**: 7.3

- [ ] G4. Create health check endpoints
  - Implement /live and /ready endpoints
  - Add health checks for Postgres, Redis, and AAA client connectivity
  - Create dependency health monitoring
  - **Priority**: P0 | **Estimate**: 3 SP | **Owner**: TBD | **Labels**: [backend], [ops], [health]
  - **Dependencies**: C1, E1, D1
  - **Requirements**: 7.4

- [ ] G5. Create operational runbooks
  - Document procedures for AAA service outages
  - Create runbooks for database pressure and Redis eviction scenarios
  - Add cache stampede prevention and recovery procedures
  - **Priority**: P1 | **Estimate**: 3 SP | **Owner**: TBD | **Labels**: [ops], [documentation]
  - **Dependencies**: G1, G2, G3, G4
  - **Requirements**: 7.5

## EPIC H — Security & Compliance

- [ ] H1. Implement secrets management with AWS SSM
  - Configure AWS SSM Parameter Store and Secrets Manager integration
  - Create environment-specific configuration matrices
  - Ensure no secrets are stored in environment files
  - **Priority**: P0 | **Estimate**: 3 SP | **Owner**: TBD | **Labels**: [security], [ops]
  - **Dependencies**: None
  - **Requirements**: 8.1

- [ ] H2. Add request validation and media upload security
  - Implement request size and MIME type validation
  - Create pre-signed S3 URL generation for media uploads
  - Add file type and content validation
  - **Priority**: P0 | **Estimate**: 4 SP | **Owner**: TBD | **Labels**: [security], [backend]
  - **Dependencies**: B1
  - **Requirements**: 8.2

- [ ] H3. Implement rate limiting and basic WAF protection
  - Add rate limiting per IP, token, and tenant
  - Implement basic WAF rules and bot filtering for search endpoints
  - Create rate limit monitoring and alerting
  - **Priority**: P0 | **Estimate**: 4 SP | **Owner**: TBD | **Labels**: [security], [middleware]
  - **Dependencies**: C2
  - **Requirements**: 8.3

- [ ] H4. Add vulnerability scanning and SBOM generation
  - Integrate govulncheck and Trivy in CI pipeline
  - Generate Software Bill of Materials (SBOM)
  - Create vulnerability monitoring and alerting
  - **Priority**: P0 | **Estimate**: 3 SP | **Owner**: TBD | **Labels**: [security], [ci]
  - **Dependencies**: None
  - **Requirements**: 8.4

## EPIC I — Testing Strategy

- [ ] I1. Create comprehensive unit test suite
  - Achieve ≥85% coverage for validators, repositories, and services
  - Create test fixtures and mocks for external dependencies
  - Implement table-driven tests for complex validation logic
  - **Priority**: P0 | **Estimate**: 8 SP | **Owner**: TBD | **Labels**: [testing], [backend]
  - **Dependencies**: A5, B2, C2, D1
  - **Requirements**: 9.1

- [ ] I2. Implement integration tests with Docker Compose
  - Create Docker Compose setup with Postgres and Redis
  - Implement golden file testing for API responses
  - Add database transaction rollback between tests
  - **Priority**: P0 | **Estimate**: 6 SP | **Owner**: TBD | **Labels**: [testing], [integration]
  - **Dependencies**: A3, B1, E1
  - **Requirements**: 9.2

- [ ] I3. Add contract testing and API validation
  - Implement OpenAPI breaking change detection with oasdiff
  - Create AAA client mocks using Buf/connect test harness
  - Add API contract validation in CI pipeline
  - **Priority**: P0 | **Estimate**: 4 SP | **Owner**: TBD | **Labels**: [testing], [contract]
  - **Dependencies**: B1, C1
  - **Requirements**: 9.3

- [ ] I4. Create E2E smoke tests and synthetic monitoring
  - Implement post-deployment smoke tests
  - Create synthetic canary checks for health and basic catalog operations
  - Add monitoring for test results and alerting
  - **Priority**: P1 | **Estimate**: 3 SP | **Owner**: TBD | **Labels**: [testing], [ops]
  - **Dependencies**: G4, B1
  - **Requirements**: 9.4

## EPIC J — CI/CD & Release Management

- [ ] J1. Create build system with Makefile and Docker
  - Implement Makefile targets for build, test, lint, vulnerability checks
  - Create multi-stage Dockerfile with distroless base and non-root user
  - Add OpenAPI generation and validation targets
  - **Priority**: P0 | **Estimate**: 4 SP | **Owner**: TBD | **Labels**: [ci], [ops]
  - **Dependencies**: None
  - **Requirements**: 10.1

- [ ] J2. Implement GitHub Actions CI pipeline
  - Create CI workflow for build, test, lint, and vulnerability scanning
  - Add SBOM generation and container image publishing
  - Implement automated tagging and versioning
  - **Priority**: P0 | **Estimate**: 5 SP | **Owner**: TBD | **Labels**: [ci], [ops]
  - **Dependencies**: J1, H4, I1, I2
  - **Requirements**: 10.2

- [ ] J3. Create ECS blue/green deployment pipeline
  - Implement CD pipeline with ECS blue/green deployment
  - Add safe AutoMigrate execution gates
  - Create feature flag integration for risky deployment paths
  - **Priority**: P0 | **Estimate**: 6 SP | **Owner**: TBD | **Labels**: [cd], [ops]
  - **Dependencies**: J2, A3
  - **Requirements**: 10.3

- [ ] J4. Implement Infrastructure as Code
  - Create CDK/Terraform for ECS service, task definition, ALB
  - Add security groups, SSM parameters, CloudWatch configuration
  - Implement autoscaling policies and monitoring alarms
  - **Priority**: P0 | **Estimate**: 7 SP | **Owner**: TBD | **Labels**: [ops], [infrastructure]
  - **Dependencies**: H1, G2
  - **Requirements**: 10.4

- [ ] J5. Create rollback procedures and database safety
  - Document rollback Standard Operating Procedures
  - Implement database AutoMigrate safety checks
  - Add feature flags for production database operations
  - **Priority**: P0 | **Estimate**: 3 SP | **Owner**: TBD | **Labels**: [ops], [safety]
  - **Dependencies**: J3, A3
  - **Requirements**: 10.5

## EPIC K — DX & Documentation

- [ ] K1. Create developer onboarding documentation
  - Write comprehensive README with quickstart guide
  - Implement `make dev-up` with Docker Compose
  - Create database seeders and example .env files (without secrets)
  - **Priority**: P0 | **Estimate**: 4 SP | **Owner**: TBD | **Labels**: [dx], [documentation]
  - **Dependencies**: J1, A3
  - **Requirements**: 11.1

- [ ] K2. Create API documentation and examples
  - Set up Swagger UI for interactive API exploration
  - Create Postman collection with example requests
  - Add grpcurl examples for AAA service integration
  - **Priority**: P0 | **Estimate**: 3 SP | **Owner**: TBD | **Labels**: [dx], [documentation]
  - **Dependencies**: B1, C1
  - **Requirements**: 11.2

- [ ] K3. Document architectural decisions
  - Create ADRs for AAA client-only choice, GORM AutoMigrate strategy
  - Document search approach, caching strategy, and tenancy enforcement
  - Add decision rationale and trade-offs analysis
  - **Priority**: P1 | **Estimate**: 3 SP | **Owner**: TBD | **Labels**: [documentation], [architecture]
  - **Dependencies**: All major epics
  - **Requirements**: 11.3

- [ ] K4. Create extension playbooks
  - Write "How to add a new catalog type" playbook
  - Document service extension patterns and best practices
  - Create troubleshooting guides for common issues
  - **Priority**: P1 | **Estimate**: 2 SP | **Owner**: TBD | **Labels**: [dx], [documentation]
  - **Dependencies**: A5, K3
  - **Requirements**: 11.4

---

## Dependency Graph

```
A1 → A2 → A3, A4
A1, A2 → A5
A1 → B1 → B2, B5
A5, B1 → B2
B3 (independent)
A2, B3 → B4
C1 → C2 → C3
C2, A2 → C3
A3, C3 → C4
A2, C3 → D1 → D2
A2, D1 → D3
D1 → E1
A4, D1 → E2
B1, E1 → E3
A4, D1 → F1 → F2
C3 → G1
B1 → G2
G1, G2 → G3
C1, E1, D1 → G4
G1, G2, G3, G4 → G5
H1 (independent)
B1 → H2
C2 → H3
H4 (independent)
A5, B2, C2, D1 → I1
A3, B1, E1 → I2
B1, C1 → I3
G4, B1 → I4
J1 (independent)
J1, H4, I1, I2 → J2
J2, A3 → J3
H1, G2 → J4
J3, A3 → J5
J1, A3 → K1
B1, C1 → K2
All major epics → K3
A5, K3 → K4
```

## Two-Week Sprint-1 Plan (Gate 1)

### Week 1: Foundation & Core Domain

**Day 1-2: Domain Modeling**

- A1: Create canonical domain model structure (5 SP)
- Start A2: Begin GORM models implementation

**Day 3-4: Data Layer**

- Complete A2: Finish GORM models with tags and constraints (8 SP)
- A3: Create AutoMigrate bootstrap (5 SP)

**Day 5: Database Optimization**

- A4: Optimize database indexes (3 SP)
- Start B3: Begin error handling system

### Week 2: API & Security Foundation

**Day 6-7: API Design**

- Complete B3: Structured error handling (4 SP)
- B1: Create OpenAPI v3 specification (6 SP)

**Day 8-9: Validation & Security**

- A5: Type-specific attribute validation (6 SP)
- C1: AAA gRPC client integration (4 SP)

**Day 10: Authentication & Health**

- C2: Auth middleware implementation (6 SP)
- G4: Health check endpoints (3 SP)

**Sprint 1 Deliverables:**

- Complete domain models with GORM integration
- Database schema with optimized indexes
- OpenAPI specification
- Basic error handling framework
- AAA service integration
- Authentication middleware
- Health monitoring endpoints

**Total Story Points: 44 SP**
**Risk Mitigation**: Focus on P0 items only, defer P1/P2 items to later sprints
**Success Criteria**: All P0 foundation components working with basic CRUD operations
