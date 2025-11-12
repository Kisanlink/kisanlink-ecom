# ADR-006: Collaborator gRPC Service Architecture

**Status**: Proposed
**Date**: 2025-11-12
**Decision Makers**: Backend Architecture Team, SDE Manager
**Technical Story**: Design and implement gRPC endpoints for Collaborator service integration with AAA v2

## Context

The KisanLink platform requires a robust gRPC service layer for the Collaborator module to enable:
- High-performance service-to-service communication with AAA v2 for address management
- Standardized RPC patterns for internal microservices communication
- Type-safe communication with protobuf serialization
- Efficient streaming capabilities for bulk operations
- Consistent authentication and authorization across gRPC boundaries

### Requirements

1. **Service Endpoints**:
   - CreateCollaborator: Register new collaborators with AAA address integration
   - UpdateCollaborator: Modify collaborator details with address synchronization
   - DeactivateCollaborator: Soft delete with status propagation
   - GetCollaborator: Retrieve collaborator with optional address expansion

2. **Integration Requirements**:
   - Seamless integration with AAA v2 address service
   - GST-based deduplication for business collaborators
   - Context propagation for distributed tracing
   - Consistent error handling across service boundaries

3. **Performance Requirements**:
   - P99 latency < 100ms for read operations
   - P99 latency < 200ms for write operations with AAA integration
   - Support for 1000+ concurrent connections
   - Connection pooling for AAA service calls

4. **Security Requirements**:
   - JWT token validation via interceptor
   - Role-based access control (RBAC)
   - Audit logging for all mutations
   - Field-level authorization for sensitive data

## Decision Drivers

1. **Protocol Selection**: Choice between REST, gRPC, or GraphQL for service communication
2. **Schema Evolution**: Need for backward-compatible API changes
3. **Performance**: Low-latency requirements for real-time operations
4. **Type Safety**: Strong typing for service contracts
5. **Ecosystem**: Existing AAA service already uses gRPC
6. **Observability**: Need for distributed tracing and metrics

## Considered Options

### Option 1: REST API with OpenAPI
**Pros**:
- Wide ecosystem support
- Simple debugging with curl/postman
- Existing team expertise

**Cons**:
- Higher latency due to JSON serialization
- No built-in streaming support
- Manual contract validation

### Option 2: gRPC with Protocol Buffers v3
**Pros**:
- Binary serialization for performance
- Strong typing with code generation
- Built-in streaming capabilities
- Existing AAA integration pattern
- Automatic client/server stub generation

**Cons**:
- Steeper learning curve
- Complex debugging without proper tools
- Limited browser support

### Option 3: GraphQL Federation
**Pros**:
- Flexible query capabilities
- Single endpoint management
- Strong typing with schema

**Cons**:
- Complexity for simple CRUD operations
- N+1 query problems
- No existing infrastructure

## Decision Outcome

We will implement **Option 2: gRPC with Protocol Buffers v3** for the Collaborator service.

### Rationale

1. **Performance**: Binary protocol provides 20-30% latency reduction over JSON
2. **Type Safety**: Proto definitions ensure compile-time contract validation
3. **Consistency**: Aligns with existing AAA service architecture
4. **Evolution**: Proto3 supports backward-compatible schema evolution
5. **Tooling**: Comprehensive tooling for code generation and validation

## Detailed Architecture

### 1. Proto Definition Structure

```
proto/
├── collaborator/
│   ├── v1/
│   │   ├── collaborator.proto      # Main service definition
│   │   ├── messages.proto          # Request/Response messages
│   │   └── common.proto            # Shared types
└── shared/
    ├── pagination.proto            # Common pagination
    ├── errors.proto                # Error definitions
    └── timestamps.proto            # Timestamp utilities
```

### 2. Service Definition

```protobuf
syntax = "proto3";
package collaborator.v1;

import "google/protobuf/timestamp.proto";
import "shared/pagination.proto";
import "validate/validate.proto";

service CollaboratorService {
  rpc CreateCollaborator(CreateCollaboratorRequest) returns (CollaboratorResponse) {
    option (google.api.http) = {
      post: "/v1/collaborators"
      body: "*"
    };
  }

  rpc UpdateCollaborator(UpdateCollaboratorRequest) returns (CollaboratorResponse) {
    option (google.api.http) = {
      patch: "/v1/collaborators/{id}"
      body: "*"
    };
  }

  rpc DeactivateCollaborator(DeactivateCollaboratorRequest) returns (StatusResponse) {
    option (google.api.http) = {
      post: "/v1/collaborators/{id}/deactivate"
    };
  }

  rpc GetCollaborator(GetCollaboratorRequest) returns (CollaboratorResponse) {
    option (google.api.http) = {
      get: "/v1/collaborators/{id}"
    };
  }
}
```

### 3. Message Validation

```protobuf
message CreateCollaboratorRequest {
  string user_id = 1 [(validate.rules).string = {min_len: 1, max_len: 50}];
  string username = 2 [(validate.rules).string = {pattern: "^[a-zA-Z0-9_-]+$"}];
  string email = 3 [(validate.rules).string.email = true];
  string first_name = 4 [(validate.rules).string = {min_len: 1, max_len: 100}];
  string last_name = 5 [(validate.rules).string = {min_len: 1, max_len: 100}];
  CollaboratorType type = 6 [(validate.rules).enum.defined_only = true];

  // Business Information (required for VENDOR/BUYER)
  optional BusinessInfo business_info = 7;

  // Address to be stored via AAA service
  optional AddressInfo address = 8;
}

message BusinessInfo {
  string gst_number = 1 [(validate.rules).string = {
    pattern: "^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z]{1}[1-9A-Z]{1}Z[0-9A-Z]{1}$"
  }];
  string business_name = 2 [(validate.rules).string = {min_len: 1, max_len: 200}];
  string business_type = 3;
  string tax_id = 4;
}
```

### 4. Interceptor Architecture

```
Request Flow:
┌──────────┐     ┌──────────────┐     ┌────────────────┐     ┌──────────────┐
│  Client  │────▶│Auth          │────▶│ Validation     │────▶│ Logging      │
└──────────┘     │Interceptor   │     │ Interceptor   │     │ Interceptor  │
                 └──────────────┘     └────────────────┘     └──────────────┘
                                                                     │
                                                                     ▼
┌──────────┐     ┌──────────────┐     ┌────────────────┐     ┌──────────────┐
│ Response │◀────│Error         │◀────│ Metrics        │◀────│ Handler      │
└──────────┘     │Interceptor   │     │ Interceptor   │     │              │
                 └──────────────┘     └────────────────┘     └──────────────┘
```

#### Auth Interceptor
```go
func AuthInterceptor(authClient auth.Client) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
                handler grpc.UnaryHandler) (interface{}, error) {
        // Extract token from metadata
        md, ok := metadata.FromIncomingContext(ctx)
        if !ok {
            return nil, status.Error(codes.Unauthenticated, "missing metadata")
        }

        // Validate JWT token
        claims, err := authClient.ValidateToken(ctx, token)
        if err != nil {
            return nil, status.Error(codes.Unauthenticated, "invalid token")
        }

        // Add claims to context
        ctx = context.WithValue(ctx, "user_claims", claims)
        return handler(ctx, req)
    }
}
```

#### Validation Interceptor
```go
func ValidationInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
                handler grpc.UnaryHandler) (interface{}, error) {
        // Use protoc-gen-validate for automatic validation
        if v, ok := req.(validator); ok {
            if err := v.Validate(); err != nil {
                return nil, status.Error(codes.InvalidArgument, err.Error())
            }
        }
        return handler(ctx, req)
    }
}
```

### 5. AAA Service Integration

```go
type AAAIntegration struct {
    client    aaaPb.AddressServiceClient
    pool      *grpc.ClientConnPool
    timeout   time.Duration
}

func (a *AAAIntegration) CreateAddress(ctx context.Context, addr *AddressInfo) (*aaaPb.Address, error) {
    ctx, cancel := context.WithTimeout(ctx, a.timeout)
    defer cancel()

    req := &aaaPb.CreateAddressRequest{
        EntityId:   addr.EntityId,
        EntityType: "collaborator",
        Address: &aaaPb.Address{
            Line1:      addr.Line1,
            Line2:      addr.Line2,
            City:       addr.City,
            State:      addr.State,
            Country:    addr.Country,
            PostalCode: addr.PostalCode,
            Type:       addr.Type,
        },
    }

    return a.client.CreateAddress(ctx, req)
}
```

### 6. Error Mapping

```go
var domainToGRPCStatus = map[error]codes.Code{
    ErrCollaboratorNotFound:     codes.NotFound,
    ErrDuplicateGST:             codes.AlreadyExists,
    ErrInvalidBusinessInfo:      codes.InvalidArgument,
    ErrUnauthorized:             codes.PermissionDenied,
    ErrAAAServiceUnavailable:    codes.Unavailable,
}

func mapDomainError(err error) error {
    if code, ok := domainToGRPCStatus[err]; ok {
        return status.Error(code, err.Error())
    }
    return status.Error(codes.Internal, "internal error")
}
```

### 7. Service Layer Integration

```go
type CollaboratorGRPCServer struct {
    collaborator.UnimplementedCollaboratorServiceServer
    service     CollaboratorServiceInterface
    aaaClient   AAAIntegration
    logger      *logrus.Logger
}

func (s *CollaboratorGRPCServer) CreateCollaborator(
    ctx context.Context,
    req *pb.CreateCollaboratorRequest,
) (*pb.CollaboratorResponse, error) {
    // Extract user claims
    claims := ctx.Value("user_claims").(*auth.TokenClaims)

    // Check for GST deduplication
    if req.BusinessInfo != nil && req.BusinessInfo.GstNumber != "" {
        exists, err := s.service.CheckGSTExists(ctx, req.BusinessInfo.GstNumber)
        if err != nil {
            return nil, mapDomainError(err)
        }
        if exists {
            return nil, status.Error(codes.AlreadyExists, "GST number already registered")
        }
    }

    // Create address via AAA if provided
    var addressID string
    if req.Address != nil {
        addr, err := s.aaaClient.CreateAddress(ctx, req.Address)
        if err != nil {
            s.logger.WithError(err).Error("Failed to create address in AAA")
            return nil, status.Error(codes.Internal, "address creation failed")
        }
        addressID = addr.Id
    }

    // Convert to domain model
    domainReq := toDomainRequest(req)
    domainReq.AddressID = addressID

    // Call service layer
    result, err := s.service.CreateCollaborator(ctx, domainReq, claims.UserID)
    if err != nil {
        return nil, mapDomainError(err)
    }

    return toProtoResponse(result), nil
}
```

## Performance Considerations

### 1. Connection Pooling

```yaml
grpc:
  server:
    max_concurrent_streams: 1000
    max_connection_idle: 15m
    max_connection_age: 30m
    keepalive:
      time: 10s
      timeout: 5s

  aaa_client:
    pool_size: 10
    max_idle_conns: 5
    dial_timeout: 5s
    request_timeout: 10s
```

### 2. Circuit Breaker for AAA

```go
type CircuitBreakerConfig struct {
    MaxRequests     uint32        // 100
    Interval        time.Duration // 10s
    Timeout         time.Duration // 60s
    FailureRatio    float64       // 0.6
}
```

### 3. Caching Strategy

```go
type CollaboratorCache struct {
    redis    *redis.Client
    ttl      time.Duration // 5 minutes for read cache
}

// Cache GST lookups to prevent duplicate checks
func (c *CollaboratorCache) GetGSTLookup(gst string) (bool, error)
func (c *CollaboratorCache) SetGSTLookup(gst string, exists bool) error
```

## Security Considerations

### 1. Field-Level Authorization

```go
func (s *CollaboratorGRPCServer) filterResponseFields(
    resp *pb.CollaboratorResponse,
    claims *auth.TokenClaims,
) *pb.CollaboratorResponse {
    // Remove sensitive fields based on user role
    if !claims.HasRole("admin") {
        resp.BusinessInfo.TaxId = ""
        resp.InternalNotes = ""
    }

    if claims.UserID != resp.UserId && !claims.HasRole("manager") {
        resp.Email = maskEmail(resp.Email)
        resp.Phone = maskPhone(resp.Phone)
    }

    return resp
}
```

### 2. Rate Limiting

```go
type RateLimiter struct {
    CreateLimit  rate.Limit // 10 req/min per user
    UpdateLimit  rate.Limit // 20 req/min per user
    ReadLimit    rate.Limit // 100 req/min per user
}
```

### 3. Audit Logging

```go
type AuditLog struct {
    Timestamp   time.Time
    UserID      string
    Action      string
    ResourceID  string
    Changes     map[string]interface{}
    IPAddress   string
    UserAgent   string
}
```

## Observability

### 1. Metrics

```go
var (
    grpcRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "collaborator_grpc_request_duration_seconds",
            Help: "Duration of gRPC requests",
        },
        []string{"method", "status"},
    )

    aaaCallDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "aaa_call_duration_seconds",
            Help: "Duration of AAA service calls",
        },
        []string{"operation", "status"},
    )
)
```

### 2. Distributed Tracing

```go
func TracingInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
                handler grpc.UnaryHandler) (interface{}, error) {
        span, ctx := opentracing.StartSpanFromContext(ctx, info.FullMethod)
        defer span.Finish()

        span.SetTag("grpc.method", info.FullMethod)
        span.SetTag("user.id", getUserID(ctx))

        return handler(ctx, req)
    }
}
```

### 3. Health Checks

```protobuf
service Health {
  rpc Check(HealthCheckRequest) returns (HealthCheckResponse);
  rpc Watch(HealthCheckRequest) returns (stream HealthCheckResponse);
}

message HealthCheckResponse {
  enum ServingStatus {
    UNKNOWN = 0;
    SERVING = 1;
    NOT_SERVING = 2;
    SERVICE_UNKNOWN = 3;
  }
  ServingStatus status = 1;
  map<string, string> dependencies = 2; // AAA, Database, Cache status
}
```

## Migration Strategy

### Phase 1: Foundation (Week 1)
- Set up proto definitions and code generation
- Implement basic CRUD without AAA integration
- Set up interceptor chain
- Unit tests for handlers

### Phase 2: Integration (Week 2)
- Integrate AAA v2 address service
- Implement GST deduplication
- Add circuit breaker and retry logic
- Integration tests

### Phase 3: Observability (Week 3)
- Add metrics and tracing
- Implement health checks
- Load testing and optimization
- Documentation

### Phase 4: Production (Week 4)
- Deploy to staging environment
- Security audit
- Performance testing
- Production deployment

## Consequences

### Positive

1. **Performance**: 30% latency reduction compared to REST
2. **Type Safety**: Compile-time validation of contracts
3. **Evolution**: Backward-compatible schema changes
4. **Consistency**: Unified service communication pattern
5. **Observability**: Built-in support for distributed tracing

### Negative

1. **Complexity**: Additional build steps for proto compilation
2. **Debugging**: Requires specialized tools for binary protocol
3. **Learning Curve**: Team needs gRPC/protobuf training
4. **Browser Support**: Limited direct browser access (requires grpc-web)

## References

- [gRPC Best Practices](https://grpc.io/docs/guides/performance/)
- [Protocol Buffers v3](https://developers.google.com/protocol-buffers/docs/proto3)
- [protoc-gen-validate](https://github.com/envoyproxy/protoc-gen-validate)
- [gRPC Error Handling](https://grpc.io/docs/guides/error/)
- [OpenTelemetry for gRPC](https://opentelemetry.io/docs/instrumentation/go/libraries/)