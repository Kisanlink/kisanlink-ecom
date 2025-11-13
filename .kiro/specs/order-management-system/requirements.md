# Order Management System - Requirements Document

## Introduction

The Order Management System is the core transactional engine of the KisanLink E-commerce Service, providing HTTP REST APIs for FPOs (Farmer Producer Organizations) and farmers to create, track, and fulfill orders for agricultural products and services. Built using Gin framework with layered architecture (main → routes → handlers → services → repositories), this system integrates with aaa-service for authentication/authorization and kisanlink-db for PostgreSQL persistence. The system enforces business rules where collaborators/FPOs can create/update catalog under their org, farmers/customers can buy, and orders validate inventory and ownership with deny-by-default permissions.

## Requirements

### Requirement 0: Catalog Management for Products, Services, and Labour

**User Story:** As collaborator, I want to create and manage catalog entries for products, services, and labour through REST APIs, so that I can offer agricultural items and services to customers within my organization's marketplace.

#### Acceptance Criteria

1. WHEN an authenticated collaborator creates a catalog entry via POST /api/v1/catalog/{type} (where type is products/services/labour) THEN the system SHALL validate ownership permissions and catalog data
2. WHEN creating catalog entries THEN the system SHALL enforce that only collaborators can create/update catalog items under their organization through aaa-service validation
3. WHEN a catalog entry is created THEN the system SHALL support different schemas for Products (name, description, price, quantity, unit), Services (name, description, rate, duration), and Labour (skill_type, hourly_rate, availability) with inheritance from same models where needed.
4. IF user lacks collaborator permissions OR tries to create catalog outside their org THEN the system SHALL return HTTP 403 with deny-by-default enforcement
5. WHEN catalog entries are created/updated THEN the system SHALL use kisanlink-db models with proper audit trail and organization association

### Requirement 1: Order Creation and Management

**User Story:** As an FPO collaborator, I want to create orders through REST APIs and manage the order lifecycle, so that I can facilitate agricultural product transactions efficiently within my organization.

#### Acceptance Criteria

1. WHEN an authenticated user creates an order via POST /api/v1/orders THEN the system SHALL validate product availability, ownership permissions, and calculate total pricing
2. WHEN an order is created THEN the system SHALL validate inventory through kisanlink-db models and generate a unique order ID
3. WHEN an order contains multiple products THEN the system SHALL support mixed catalog types (Products, Services, Labour) in a single order
4. IF inventory is insufficient OR user lacks permissions THEN the system SHALL return HTTP 400/403 with specific error details
5. WHEN an order is created THEN the system SHALL record buyer organization, seller organization, and audit trail using kisanlink-db base models

### Requirement 2: Order Status Tracking and Updates

**User Story:** As a customer, I want to track my order status through REST APIs, so that I can plan my agricultural activities accordingly.

#### Acceptance Criteria

1. WHEN an order status changes via PUT /api/v1/orders/{id}/status THEN the system SHALL validate permissions through aaa-service and update state
2. WHEN querying order status via GET /api/v1/orders/{id} THEN the system SHALL return current status with proper authorization checks
3. WHEN an order progresses through states THEN the system SHALL enforce valid state transitions (Pending → Confirmed → Paid → Shipped → Delivered → Completed)
4. IF an order is cancelled THEN the system SHALL release reserved inventory through repository layer and update financial records
5. WHEN order status is updated THEN the system SHALL return clean JSON response format with updated order details

### Requirement 3: Order Listing and Filtering

**User Story:** As an FPO collaborator, I want to list and filter orders through REST APIs, so that I can efficiently manage transactions within my organization.

#### Acceptance Criteria

1. WHEN listing orders via GET /api/v1/orders THEN the system SHALL filter results based on user's organizational permissions from aaa-service
2. WHEN applying filters THEN the system SHALL support query parameters for date range, status, product type, and organization
3. WHEN displaying order lists THEN the system SHALL paginate results and provide sorting options with clean JSON response format
4. IF user lacks permissions THEN the system SHALL return HTTP 403 with deny-by-default enforcement
5. WHEN querying large datasets THEN the system SHALL maintain performance under 2 seconds using efficient PostgreSQL queries

### Requirement 4: Authentication and Authorization Integration

**User Story:** As a system administrator, I want all order operations to integrate with aaa-service for authentication and authorization, so that access is properly controlled with deny-by-default permissions.

#### Acceptance Criteria

1. WHEN any order API is called THEN the system SHALL validate authentication tokens through aaa-service integration
2. WHEN checking permissions THEN the system SHALL enforce role-based access where FPOs can manage their org's orders and farmers can view their purchases
3. WHEN unauthorized access is attempted THEN the system SHALL return HTTP 401/403 with appropriate error messages
4. IF aaa-service is unavailable THEN the system SHALL gracefully handle failures and deny access by default
5. WHEN permissions are validated THEN the system SHALL use middleware for AuthN/AuthZ with clean separation of concerns

### Requirement 5: Data Persistence and Repository Layer

**User Story:** As a developer, I want order data to be properly persisted using kisanlink-db models and PostgreSQL, so that data integrity and consistency are maintained.

#### Acceptance Criteria

1. WHEN storing order data THEN the system SHALL use kisanlink-db base models with proper entity relationships
2. WHEN performing CRUD operations THEN the system SHALL implement repository pattern with interface-based design
3. WHEN handling database transactions THEN the system SHALL ensure ACID properties for order creation and updates
4. IF database operations fail THEN the system SHALL return appropriate HTTP status codes and error messages
5. WHEN querying orders THEN the system SHALL use efficient PostgreSQL queries with proper indexing and pagination

### Requirement 6: API Documentation and Testing

**User Story:** As a developer/API consumer, I want comprehensive Swagger documentation and unit tests for all order APIs, so that the system is maintainable and reliable.

#### Acceptance Criteria

1. WHEN APIs are developed THEN the system SHALL provide complete Swagger documentation for all order endpoints
2. WHEN testing functionality THEN the system SHALL include unit tests with mocked services and repositories
3. WHEN building the project THEN the system SHALL provide Makefile commands for run/test/lint/swagger operations
4. IF API contracts change THEN the system SHALL update Swagger docs and maintain backward compatibility
5. WHEN deploying THEN the system SHALL ensure all tests pass and documentation is up-to-date with README instructions
