# Inventory Lot Auto-Generation - Requirements Document

## Introduction

This feature implements automatic generation of lot numbers and batch numbers for inventory lots to prevent duplicate key constraint violations. The system will generate unique, sequential identifiers for each organization, ensuring traceability and collision-free inventory management.

## Problem Statement

Currently, the system requires manual entry of lot numbers and batch numbers, leading to duplicate key violations when users accidentally reuse the same lot number within an organization:

```
ERROR: duplicate key value violates unique constraint "idx_org_lot_number" (SQLSTATE 23505)
```

This occurs because:

- The `lot_number` field has a unique constraint per organization (`uniqueIndex:idx_inventory_lots_org_lot_number`)
- Users must manually provide lot numbers, which can lead to collisions
- No automatic sequencing mechanism exists

## Requirements

### Requirement 1: Automatic Lot Number Generation

**User Story:** As a collaborator managing inventory, I want lot numbers to be automatically generated when not provided, so that I don't have duplicate key errors and maintain unique identifiers.

#### Acceptance Criteria

1. WHEN creating an inventory lot via POST /api/v1/inventory/lots WITHOUT providing a lot_number THEN the system SHALL auto-generate a unique lot number for the organization
2. WHEN creating an inventory lot WITH a provided lot_number THEN the system SHALL use the provided value and validate uniqueness
3. WHEN auto-generating lot numbers THEN the system SHALL use format: `LOT-{ORG_PREFIX}-{YYYYMMDD}-{SEQUENCE}` (e.g., `LOT-ABC-20251024-001`)
4. WHEN the lot number is auto-generated THEN it SHALL be guaranteed unique within the organization through atomic database operations
5. IF a duplicate lot number is provided manually THEN the system SHALL return HTTP 400 with clear error message

### Requirement 2: Automatic Batch Number Generation

**User Story:** As a collaborator managing inventory, I want batch numbers to be automatically generated when not provided, so that I can easily track related inventory lots.

#### Acceptance Criteria

1. WHEN creating an inventory lot WITHOUT providing a batch_number THEN the system SHALL auto-generate a batch number
2. WHEN creating an inventory lot WITH a provided batch_number THEN the system SHALL use the provided value
3. WHEN auto-generating batch numbers THEN the system SHALL use format: `BATCH-{YYYYMMDD}-{SEQUENCE}` (e.g., `BATCH-20251024-001`)
4. WHEN multiple lots are created on the same day THEN they MAY share the same batch number if created together or have sequential batch numbers
5. The batch_number field SHALL remain non-unique to allow grouping of lots

### Requirement 3: Backwards Compatibility

**User Story:** As an existing user, I want to continue providing my own lot and batch numbers if I prefer, so that my existing workflows are not disrupted.

#### Acceptance Criteria

1. WHEN a user provides both lot_number and batch_number THEN the system SHALL use the provided values
2. WHEN a user provides only lot_number THEN the system SHALL use it and auto-generate batch_number
3. WHEN a user provides only batch_number THEN the system SHALL auto-generate lot_number and use provided batch_number
4. WHEN neither is provided THEN the system SHALL auto-generate both
5. The API documentation SHALL indicate these fields are optional with auto-generation

### Requirement 4: Organization Prefix Management

**User Story:** As an administrator, I want each organization to have a unique prefix for lot numbers, so that lots are easily identifiable across organizations.

#### Acceptance Criteria

1. WHEN an organization is created THEN it MAY have an optional short prefix code (3-5 characters)
2. WHEN auto-generating lot numbers THEN the system SHALL use the organization's prefix if available, otherwise derive from organization ID
3. WHEN no organization prefix exists THEN the system SHALL use first 3-5 characters of organization ID
4. The organization prefix SHALL be alphanumeric and uppercase
5. The prefix derivation SHALL be deterministic and consistent

### Requirement 5: Sequence Management

**User Story:** As a system administrator, I want sequences to be managed reliably using database counters, so that no duplicate numbers are generated even under high concurrency.

#### Acceptance Criteria

1. WHEN generating lot/batch numbers THEN the system SHALL use database-level sequence management or atomic counters
2. WHEN multiple requests create lots concurrently THEN each SHALL receive a unique sequence number
3. WHEN a lot creation fails THEN sequence numbers MAY have gaps (this is acceptable)
4. WHEN querying the next sequence number THEN it SHALL be done atomically within a transaction
5. The sequence SHALL reset daily based on the date component in the format

### Requirement 6: Database-Initialized ID Sequences

**User Story:** As a system administrator, I want all entity IDs (catalog items, lots, orders, etc.) to use database-initialized sequences instead of random generation, so that IDs are predictable, sequential, and traceable.

#### Acceptance Criteria

1. WHEN the application starts THEN it SHALL initialize sequence counters from the database for all entity types
2. WHEN creating any entity (catalog item, inventory lot, order, etc.) THEN the system SHALL use database-backed sequences for ID generation
3. WHEN an entity is created THEN its ID SHALL follow format: `{PREFIX}{SEQUENCE}` (e.g., `CAT000001`, `LOT000001`)
4. WHEN the database is queried for max ID THEN the sequence counter SHALL initialize to max+1
5. IF no entities exist THEN the sequence SHALL start at 1
6. The ID generation SHALL NOT use random number generation (no random UUIDs or timestamps in IDs)
7. WHEN the application restarts THEN ID sequences SHALL resume from the last used ID, not reset to 1

## Non-Functional Requirements

### Performance

- Lot number generation SHALL complete in < 100ms
- SHALL support concurrent lot creation without deadlocks
- Sequence lookup SHALL be optimized with appropriate indexes

### Reliability

- Zero collision guarantee for auto-generated numbers
- Transactional integrity for sequence allocation
- Graceful fallback if sequence service is unavailable

### Auditability

- Log all auto-generated lot and batch numbers
- Include generation metadata in audit trail
- Track whether numbers were auto-generated or user-provided

## Constraints

1. Must maintain existing database schema and unique constraints
2. Must not break existing API contracts
3. Must support PostgreSQL as primary database
4. Must work within the kisanlink-db abstraction layer
5. Auto-generation logic must be in the service layer, not handlers

## Success Criteria

1. Zero duplicate key violations for auto-generated lot numbers
2. Existing manual lot number entry still works
3. All existing tests pass
4. New tests cover auto-generation scenarios
5. API documentation updated to reflect optional fields
