# Requirements Document

## Introduction

This specification outlines the requirements for taking the Golang + Gin "Marketplace Catalogs" microservice from its current state to production-ready status. The service manages marketplace catalogs with 4 catalog types: Products, Services, Labour, and Contracts. The goal is to create a robust, scalable, secure, and maintainable microservice that meets enterprise production standards with comprehensive observability, security, and operational excellence.

## Requirements

### Requirement 1: Domain & Data Modeling

**User Story:** As a platform architect, I want a well-defined domain model with proper GORM-based persistence, so that the catalog system can handle all four catalog types with appropriate data integrity and performance.

#### Acceptance Criteria

1. WHEN defining the domain model THEN the system SHALL support canonical entities: Catalog, Category, Variant, Pricing, Availability, Attributes (JSONB), Media, Vendor/FPO, SLA, PublishState
2. WHEN implementing GORM models THEN the system SHALL include proper table names, indexes, unique constraints, soft-delete support, audit columns, and tenant isolation
3. WHEN bootstrapping the database THEN the system SHALL use GORM AutoMigrate with a custom schema_migrations table for version tracking
4. WHEN querying data THEN the system SHALL have optimized indexes for composite queries (tenant_id, type, status, updated_at DESC) and full-text search capabilities
5. WHEN handling different catalog types THEN the system SHALL enforce type-specific attribute policies (Contracts: term/duration, Labour: skill/unit rate, Services: SLA fields, Products: SKU/variants)

### Requirement 2: HTTP API & Validation

**User Story:** As an API consumer, I want a well-documented, validated REST API with proper error handling, so that I can reliably integrate with the catalog service.

#### Acceptance Criteria

1. WHEN accessing the API THEN the system SHALL provide OpenAPI v3 specification with CRUD operations, pagination, sorting, filtering, and bulk operations
2. WHEN submitting requests THEN the system SHALL validate input using go-playground/validator with custom rules for catalog-specific fields
3. WHEN errors occur THEN the system SHALL return structured error responses with code, message, details, and request_id without leaking internal details
4. WHEN performing write operations THEN the system SHALL support idempotency using Idempotency-Key headers and optimistic locking
5. WHEN caching responses THEN the system SHALL implement ETag/If-None-Match for GET operations

### Requirement 3: Authentication & Authorization

**User Story:** As a security administrator, I want proper authentication and authorization integrated with the AAA service, so that only authorized users can access appropriate catalog resources.

#### Acceptance Criteria

1. WHEN integrating with AAA THEN the system SHALL use gRPC client only (no gRPC server) with proper TLS/mTLS configuration
2. WHEN processing requests THEN the system SHALL verify JWT tokens locally and call AAA Authorize for resource-level permissions
3. WHEN enforcing tenancy THEN the system SHALL extract tenant information and apply row-level security on all queries
4. WHEN auditing actions THEN the system SHALL log all operations with who/what/when/before/after to an append-only audit table

### Requirement 4: Data Persistence & Integrity

**User Story:** As a data engineer, I want reliable data persistence with proper transaction handling and integrity constraints, so that catalog data remains consistent and recoverable.

#### Acceptance Criteria

1. WHEN accessing data THEN the system SHALL use a repository layer with context-aware queries and prevent N+1 problems
2. WHEN performing multi-entity operations THEN the system SHALL use database transactions with proper rollback and retry logic
3. WHEN managing data lifecycle THEN the system SHALL implement soft-delete policies and prevent deletion of referenced entities
4. WHEN handling PII THEN the system SHALL have documented data retention policies and PII mapping

### Requirement 5: Performance & Caching

**User Story:** As a performance engineer, I want optimized queries and intelligent caching, so that the service can handle production load with acceptable response times.

#### Acceptance Criteria

1. WHEN caching data THEN the system SHALL use Redis with per-tenant namespaces, TTL policies, and cache invalidation on writes
2. WHEN optimizing queries THEN the system SHALL use EXPLAIN ANALYZE on realistic datasets and implement GORM scopes for common filters
3. WHEN load testing THEN the system SHALL meet P50/P95 latency targets with defined error budgets using k6 load profiles
4. WHEN serving requests THEN the system SHALL achieve cache hit ratios ≥70% for hot endpoints

### Requirement 6: Search & Discovery

**User Story:** As an end user, I want to search and filter catalogs efficiently, so that I can quickly find relevant products, services, labor, or contracts.

#### Acceptance Criteria

1. WHEN implementing search THEN the system SHALL provide PostgreSQL native search (ILIKE/trigram) with filter facets
2. WHEN filtering results THEN the system SHALL support filters by type, category, tags, price bands, status, and vendor
3. WHEN designing search architecture THEN the system SHALL use abstraction interfaces to allow future OpenSearch integration
4. WHEN indexing data THEN the system SHALL provide backfill and index job capabilities

### Requirement 7: Observability & Operations

**User Story:** As a DevOps engineer, I want comprehensive observability and operational tooling, so that I can monitor, troubleshoot, and maintain the service effectively.

#### Acceptance Criteria

1. WHEN logging events THEN the system SHALL use structured logging with request_id, tenant_id, user_id, and proper redaction policies
2. WHEN collecting metrics THEN the system SHALL expose Prometheus metrics for RPS, latency, error rate, cache hit ratio, and database timings
3. WHEN tracing requests THEN the system SHALL implement OpenTelemetry tracing across HTTP→DB→Redis→AAA client calls
4. WHEN checking health THEN the system SHALL provide /live and /ready endpoints that verify Postgres, Redis, and AAA client connectivity
5. WHEN handling incidents THEN the system SHALL have documented runbooks for common scenarios

### Requirement 8: Security & Compliance

**User Story:** As a security officer, I want the service to follow security best practices and compliance requirements, so that sensitive data and operations are properly protected.

#### Acceptance Criteria

1. WHEN managing secrets THEN the system SHALL use AWS SSM/Secrets Manager with environment-specific configuration matrices
2. WHEN handling uploads THEN the system SHALL validate request sizes, MIME types, and use pre-signed S3 URLs only
3. WHEN protecting against abuse THEN the system SHALL implement rate limiting per IP/token/tenant and basic WAF rules
4. WHEN scanning for vulnerabilities THEN the system SHALL use govulncheck and Trivy with SBOM generation in CI

### Requirement 9: Testing Strategy

**User Story:** As a quality engineer, I want comprehensive test coverage across unit, integration, and contract levels, so that the service is reliable and regression-free.

#### Acceptance Criteria

1. WHEN running unit tests THEN the system SHALL achieve ≥85% coverage for critical packages (validators, repositories, services)
2. WHEN running integration tests THEN the system SHALL use Docker Compose with Postgres and Redis for realistic testing
3. WHEN validating contracts THEN the system SHALL use OpenAPI breaking change detection and AAA client mocks
4. WHEN deploying THEN the system SHALL run E2E smoke tests and synthetic canary checks

### Requirement 10: CI/CD & Release Management

**User Story:** As a release manager, I want automated CI/CD pipelines with safe deployment practices, so that releases are reliable and rollback-capable.

#### Acceptance Criteria

1. WHEN building THEN the system SHALL use Makefile targets and multi-stage Dockerfiles with distroless, non-root images
2. WHEN running CI THEN the system SHALL execute build, test, vulnerability scan, SBOM generation, and image push
3. WHEN deploying THEN the system SHALL use ECS blue/green deployment with safe AutoMigrate gates and feature flags
4. WHEN managing infrastructure THEN the system SHALL use IaC (CDK/Terraform) for ECS, ALB, security groups, and monitoring
5. WHEN rolling back THEN the system SHALL have documented procedures and database migration safety checks

### Requirement 11: Developer Experience & Documentation

**User Story:** As a developer, I want clear documentation and easy local development setup, so that I can quickly understand and contribute to the codebase.

#### Acceptance Criteria

1. WHEN setting up locally THEN the system SHALL provide README quickstart with `make dev-up` using Docker Compose
2. WHEN exploring APIs THEN the system SHALL provide Swagger UI, Postman collections, and gRPC examples
3. WHEN understanding decisions THEN the system SHALL have ADRs documenting architectural choices
4. WHEN extending functionality THEN the system SHALL provide playbooks for adding new catalog types

### Requirement 12: Production Readiness Gates

**User Story:** As a platform owner, I want defined acceptance criteria for production deployment, so that the service meets enterprise standards before going live.

#### Acceptance Criteria

1. WHEN measuring availability THEN the system SHALL achieve 99.9% uptime SLO
2. WHEN measuring performance THEN the system SHALL meet P95 ≤ 120ms for top-10 endpoints on realistic data
3. WHEN checking security THEN the system SHALL have zero critical vulnerabilities and verified secrets management
4. WHEN validating operations THEN the system SHALL have live dashboards, on-call runbooks, and validated blue/green deployment
5. WHEN planning disaster recovery THEN the system SHALL document RPO≤24h and RTO≤2h procedures
