# Collaborator API Documentation

The Collaborator API manages users on the KisanLink e-commerce platform. It stores platform-specific data while user identity and authentication are handled by the external AAA service.

## Overview

Collaborators represent users who participate in the agricultural marketplace. The system supports different types of collaborators with various roles and capabilities.

## Collaborator Types

- **FARMER**: Agricultural producers who grow crops and livestock
- **SUPPLIER**: Vendors who provide agricultural inputs, equipment, and services
- **BUYER**: Purchasers of agricultural products (wholesalers, retailers, etc.)
- **AGENT**: Intermediaries who facilitate transactions between parties
- **ADMIN**: Platform administrators with elevated privileges

## Collaborator Status

- **PENDING**: Newly registered, awaiting verification or activation
- **ACTIVE**: Fully active and can participate in all platform activities
- **INACTIVE**: Temporarily disabled, cannot perform transactions
- **SUSPENDED**: Suspended due to policy violations or other issues

## Key Features

### User Management

- Create and manage collaborator profiles
- Link to AAA service user accounts
- Track platform-specific activity and preferences

### Business Information

- Store business details for commercial users
- Support for licenses, tax information, and financial details
- Business verification and trust scoring

### Onboarding Process

- Step-by-step onboarding workflow
- Track completion status and progress
- Customizable onboarding steps per collaborator type

### Trust and Verification

- Trust score calculation based on activity and ratings
- Manual verification by administrators
- Track order history and performance metrics

### Activity Tracking

- Login and activity timestamps
- Platform usage statistics
- Performance metrics and ratings

## API Endpoints

### Core CRUD Operations

#### Create Collaborator

```
POST /api/v1/collaborators
```

Creates a new collaborator profile linked to an AAA service user.

#### Get Collaborator

```
GET /api/v1/collaborators/{id}
GET /api/v1/collaborators/user/{user_id}
```

Retrieve collaborator information by platform ID or AAA user ID.

#### Update Collaborator

```
PUT /api/v1/collaborators/{id}
```

Update collaborator profile information.

#### Delete Collaborator

```
DELETE /api/v1/collaborators/{id}
```

Soft delete a collaborator (maintains audit trail).

### Listing and Search

#### List Collaborators

```
GET /api/v1/collaborators
```

List collaborators with filtering and pagination support.

**Query Parameters:**

- `page`: Page number (default: 1)
- `limit`: Items per page (default: 20, max: 100)
- `collaborator_type`: Filter by type (FARMER, SUPPLIER, BUYER, AGENT, ADMIN)
- `status`: Filter by status (ACTIVE, INACTIVE, SUSPENDED, PENDING)
- `organization_id`: Filter by organization
- `is_verified`: Filter by verification status
- `onboarding_completed`: Filter by onboarding completion
- `location`: Filter by location (partial match)
- `business_type`: Filter by business type
- `min_trust_score`: Minimum trust score
- `max_trust_score`: Maximum trust score
- `search`: Search across name, email, username, business name
- `tags`: Filter by tags (comma-separated)
- `created_after`: Filter by creation date (ISO format)
- `created_before`: Filter by creation date (ISO format)

#### Search Collaborators

```
GET /api/v1/collaborators/search?q={query}
```

Full-text search across collaborator profiles.

### Status Management

#### Update Status

```
PATCH /api/v1/collaborators/{id}/status
```

Update collaborator status with reason tracking.

#### Verify Collaborator

```
POST /api/v1/collaborators/{id}/verify
```

Mark a collaborator as verified (admin only).

### Onboarding Management

#### Update Onboarding Step

```
PATCH /api/v1/collaborators/{id}/onboarding
```

Update the current onboarding step.

#### Complete Onboarding

```
POST /api/v1/collaborators/{id}/onboarding/complete
```

Mark onboarding as completed.

### Bulk Operations

#### Bulk Update

```
PATCH /api/v1/collaborators/bulk
```

Perform bulk updates on multiple collaborators.

### Analytics

#### Get Statistics

```
GET /api/v1/collaborators/stats
```

Retrieve platform-wide or organization-specific collaborator statistics.

#### Get Profile

```
GET /api/v1/collaborators/{id}/profile
```

Get detailed profile information (includes sensitive data for owner/admin).

## Data Models

### Collaborator Model

The main collaborator model includes:

- **Identity**: UserID (from AAA), username, email, name
- **Profile**: Avatar, bio, location, coordinates
- **Business**: Business information, licenses, financial details
- **Platform**: Status, verification, trust score, activity tracking
- **Preferences**: Language, timezone, notification settings
- **Onboarding**: Progress tracking and completion status
- **Metadata**: Tags, notes, invitation tracking

### Request/Response Models

- `CreateCollaboratorRequest`: Data for creating new collaborators
- `UpdateCollaboratorRequest`: Partial update data
- `CollaboratorResponse`: Full collaborator information
- `CollaboratorSummaryResponse`: Condensed view for listings
- `CollaboratorProfileResponse`: Detailed profile with sensitive data
- `CollaboratorStatsResponse`: Platform statistics

## Authentication and Authorization

- Most endpoints require authentication via Bearer token
- Public endpoints: Get collaborator by ID, Get by user ID
- Admin-only operations: Verification, status changes, bulk operations
- Users can only access their own sensitive information

## Trust Score Calculation

The trust score is calculated based on:

- Base score: 50 points
- Completed orders: +2 points each
- Cancelled orders: -5 points each
- Average rating: rating × 10 points
- Verification status: +20 points if verified
- Score is capped between 0 and 100

## Error Handling

The API uses standard HTTP status codes and returns structured error responses:

- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Collaborator not found
- `409 Conflict`: Duplicate collaborator (username/email)
- `500 Internal Server Error`: Server-side errors

## Integration with AAA Service

The collaborator system integrates with the external AAA service for:

- User authentication and token validation
- User identity management
- Role and permission management
- Single sign-on capabilities

Collaborator records store references to AAA user IDs and cache frequently accessed user data for performance.

## Database Schema

The collaborators table uses the BaseModel from kisanlink-db package, providing:

- Automatic ID generation
- Created/updated timestamps
- Soft delete capability
- Audit trail with created_by/updated_by/deleted_by fields

## Testing

The collaborator system includes:

- Unit tests for models and business logic
- Integration tests for repository operations
- API endpoint tests
- Test data fixtures for consistent testing

## Future Enhancements

Planned improvements include:

- Advanced trust score algorithms
- Integration with rating and review systems
- Enhanced business verification workflows
- Analytics and reporting dashboards
- Mobile app support for profile management
