# gRPC Integration with aaa-service

This document describes the integration of the kisanlink-ecom project with the aaa-service for user management, role assignment, and permission management using gRPC.

## Overview

The kisanlink-ecom project now uses the aaa-service as a gRPC backend for all user-related operations, including:

- User creation and management
- Role assignment and management
- Permission management
- Authentication and authorization

## Architecture

```
┌─────────────────┐    gRPC      ┌─────────────────┐
│   kisanlink-ecom│ ──────────► │   aaa-service   │
│   (HTTP API)    │             │   (gRPC Server) │
└─────────────────┘              └─────────────────┘
```

## Features Implemented

### 1. User Management

- ✅ User registration via gRPC
- ✅ User login with JWT token
- ✅ User CRUD operations
- ✅ User role assignment

### 2. Role Management

- ✅ Role creation
- ✅ Role retrieval (all/by ID)
- ✅ Role updates
- ✅ Role deletion

### 3. Permission Management

- ✅ Permission creation
- ✅ Permission retrieval (all/by ID)
- ✅ Permission updates
- ✅ Permission deletion

### 4. Role-Permission Connections

- ✅ Create role-permission connections
- ✅ Retrieve role-permission mappings
- ✅ Update role-permission connections
- ✅ Delete role-permission connections

## API Endpoints

### Authentication

- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/logout` - User logout

### User Management

- `GET /api/v1/users` - Get all users
- `POST /api/v1/users` - Create user
- `GET /api/v1/users/:id` - Get user by ID
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

### Role Management

- `GET /api/v1/roles` - Get all roles
- `POST /api/v1/roles` - Create role
- `GET /api/v1/roles/:id` - Get role by ID
- `PUT /api/v1/roles/:id` - Update role
- `DELETE /api/v1/roles/:id` - Delete role

### Permission Management

- `GET /api/v1/permissions` - Get all permissions
- `POST /api/v1/permissions` - Create permission
- `GET /api/v1/permissions/:id` - Get permission by ID
- `PUT /api/v1/permissions/:id` - Update permission
- `DELETE /api/v1/permissions/:id` - Delete permission

## Configuration

### Environment Variables

Add the following environment variables to your `.env` file:

```bash
# gRPC Server Configuration
AAA_GRPC_SERVER_ADDR=localhost:50051

# Other existing configurations...
PORT=8080
GIN_MODE=debug
API_VERSION=v1
```

### Configuration Structure

The gRPC configuration is integrated into the existing config structure:

```go
type Config struct {
    // ... existing fields
    GRPC GRPCConfig
}

type GRPCConfig struct {
    ServerAddr string
}
```

## Service Architecture

### 1. gRPC Client (`internal/services/grpc_client.go`)

- Handles all gRPC communication with aaa-service
- Provides methods for user, role, and permission operations
- Manages connection lifecycle

### 2. User Service (`internal/services/user_service.go`)

- Business logic for user operations
- Converts between local models and gRPC models
- Handles error mapping and response formatting

### 3. Role/Permission Service (`internal/services/role_permission_service.go`)

- Business logic for role and permission operations
- Manages role-permission connections
- Provides local models for role and permission management

### 4. Service Container (`internal/services/init.go`)

- Initializes all services with proper dependency injection
- Manages service lifecycle
- Provides centralized service access

## Handler Architecture

### 1. User Handler (`internal/handlers/users.go`)

- HTTP handlers for user management
- Uses UserService for business logic
- Provides RESTful API endpoints

### 2. Auth Handler (`internal/handlers/auth.go`)

- HTTP handlers for authentication
- Uses UserService for login/registration
- Returns JWT tokens from aaa-service

### 3. Role Handler (`internal/handlers/roles.go`)

- HTTP handlers for role management
- Uses RolePermissionService for business logic
- Provides role CRUD operations

### 4. Permission Handler (`internal/handlers/permissions.go`)

- HTTP handlers for permission management
- Uses RolePermissionService for business logic
- Provides permission CRUD operations

## Data Models

### Local Models (kisanlink-ecom)

```go
type User struct {
    ID           string
    Username     string
    Email        string
    FullName     string
    Role         UserRole
    Status       Status
}

type Role struct {
    ID          string
    Name        string
    Description string
}

type Permission struct {
    ID          string
    Name        string
    Description string
}
```

### gRPC Models (aaa-service)

```protobuf
message User {
    string id = 1;
    string username = 2;
    bool is_validated = 3;
    repeated UserRole user_roles = 6;
}

message Role {
    string id = 1;
    string name = 2;
    string description = 3;
}

message Permission {
    string id = 1;
    string name = 2;
    string description = 3;
}
```

## Error Handling

The integration includes comprehensive error handling:

1. **gRPC Connection Errors**: Handled gracefully with retry logic
2. **Service Errors**: Mapped to appropriate HTTP status codes
3. **Validation Errors**: Proper validation with detailed error messages
4. **Timeout Handling**: Configurable timeouts for gRPC calls

## Testing

### Manual Testing

Use the provided test script:

```bash
# Make the script executable
chmod +x scripts/test-grpc-integration.sh

# Run the integration tests
./scripts/test-grpc-integration.sh
```

### API Testing

Test the endpoints using curl or any API testing tool:

```bash
# Test user registration
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "full_name": "Test User",
    "password": "testpassword123",
    "role": "customer"
  }'

# Test user login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpassword123"
  }'
```

## Deployment

### Prerequisites

1. aaa-service must be running and accessible
2. gRPC server address must be configured
3. Network connectivity between services

### Steps

1. Set environment variables
2. Start aaa-service
3. Start kisanlink-ecom
4. Verify connectivity

### Docker Compose Example

```yaml
version: "3.8"
services:
  aaa-service:
    image: kisanlink/aaa-service:latest
    ports:
      - "50051:50051"
    environment:
      - DB_HOST=postgres
      - DB_USER=postgres
      - DB_PASSWORD=password
      - DB_NAME=aaa_service

  kisanlink-ecom:
    image: kisanlink/ecom:latest
    ports:
      - "8080:8080"
    environment:
      - AAA_GRPC_SERVER_ADDR=aaa-service:50051
    depends_on:
      - aaa-service
```

## Monitoring and Logging

### Logging

- gRPC connection status
- Service operation logs
- Error details and stack traces

### Metrics

- gRPC call latency
- Success/failure rates
- Connection health

### Health Checks

- gRPC server connectivity
- Service availability
- Database connectivity (via aaa-service)

## Security Considerations

1. **gRPC Security**: Use TLS for production
2. **Authentication**: JWT tokens from aaa-service
3. **Authorization**: Role-based access control
4. **Input Validation**: Comprehensive validation at all layers
5. **Error Handling**: No sensitive information in error messages

## Troubleshooting

### Common Issues

1. **gRPC Connection Failed**
   - Check aaa-service is running
   - Verify network connectivity
   - Check firewall settings

2. **Authentication Errors**
   - Verify JWT token format
   - Check token expiration
   - Validate user credentials

3. **Role/Permission Issues**
   - Ensure roles exist in aaa-service
   - Check role-permission mappings
   - Verify user role assignments

### Debug Commands

```bash
# Check gRPC server status
grpcurl -plaintext localhost:50051 list

# Test gRPC connection
grpcurl -plaintext localhost:50051 pb.UserService/GetUser

# Check service logs
docker logs aaa-service
docker logs kisanlink-ecom
```

## Future Enhancements

1. **Caching**: Redis caching for frequently accessed data
2. **Load Balancing**: Multiple aaa-service instances
3. **Circuit Breaker**: Resilience patterns for gRPC calls
4. **Metrics**: Prometheus metrics integration
5. **Tracing**: Distributed tracing with Jaeger
6. **Rate Limiting**: API rate limiting
7. **Audit Logging**: Comprehensive audit trail

## Support

For issues related to:

- **gRPC Integration**: Check this document
- **aaa-service**: Refer to aaa-service documentation
- **API Issues**: Check API documentation and logs
- **Deployment**: Review deployment guide

## Contributing

When contributing to the gRPC integration:

1. Follow the existing code structure
2. Add comprehensive tests
3. Update documentation
4. Follow error handling patterns
5. Maintain backward compatibility
