# Collaborator gRPC Service Implementation Specification

## Executive Summary

This document provides the complete implementation specification for the Collaborator gRPC service, integrating with AAA v2 for address management and implementing GST-based deduplication for business collaborators.

## Architecture Overview

```
┌──────────────┐       ┌──────────────┐       ┌──────────────┐
│   REST API   │       │  gRPC Client │       │  Admin UI    │
└──────┬───────┘       └──────┬───────┘       └──────┬───────┘
       │                      │                       │
       └──────────────────────┴───────────────────────┘
                              │
                    ┌─────────▼──────────┐
                    │   gRPC Gateway     │
                    └─────────┬──────────┘
                              │
                    ┌─────────▼──────────┐
                    │  Interceptor Chain │
                    └─────────┬──────────┘
                              │
                    ┌─────────▼──────────┐
                    │ Collaborator gRPC  │
                    │     Service         │
                    └─────────┬──────────┘
                              │
            ┌─────────────────┼─────────────────┐
            │                 │                 │
    ┌───────▼────────┐ ┌─────▼──────┐ ┌────────▼────────┐
    │ Service Layer  │ │AAA Service │ │  GST Service    │
    └───────┬────────┘ └────────────┘ └─────────────────┘
            │
    ┌───────▼────────┐
    │   Repository   │
    └───────┬────────┘
            │
    ┌───────▼────────┐
    │   Database     │
    └────────────────┘
```

## Implementation Phases

### Phase 1: Foundation (Days 1-3)

#### Tasks:
1. **Proto Setup**
   - Create proto directory structure
   - Define proto files as specified
   - Set up proto compilation pipeline
   - Generate Go code from protos

2. **Project Structure**
   ```
   internal/
   ├── grpc/
   │   ├── server.go
   │   ├── interceptors/
   │   │   ├── auth.go
   │   │   ├── validation.go
   │   │   ├── logging.go
   │   │   ├── metrics.go
   │   │   ├── ratelimit.go
   │   │   └── audit.go
   │   └── handlers/
   │       └── collaborator/
   │           ├── server.go
   │           ├── create.go
   │           ├── update.go
   │           ├── deactivate.go
   │           └── get.go
   ```

3. **Basic Server Setup**
   ```go
   // cmd/grpc/main.go
   package main

   import (
       "net"
       "os"
       "os/signal"
       "syscall"

       "google.golang.org/grpc"
       "google.golang.org/grpc/reflection"
       "github.com/sirupsen/logrus"
   )

   func main() {
       logger := logrus.New()

       // Load configuration
       cfg := config.Load()

       // Initialize database
       db := database.Initialize(cfg.Database)

       // Initialize AAA client
       aaaClient := auth.NewAAAClient(cfg.AAA)

       // Create gRPC server
       server := grpc.NewGRPCServer(
           grpc.WithLogger(logger),
           grpc.WithAAAClient(aaaClient),
           grpc.WithDatabase(db),
       )

       // Register services
       collaboratorServer := handlers.NewCollaboratorServer(...)
       pb.RegisterCollaboratorServiceServer(server, collaboratorServer)

       // Enable reflection for debugging
       reflection.Register(server)

       // Start server
       lis, err := net.Listen("tcp", cfg.GRPCPort)
       if err != nil {
           logger.Fatal("Failed to listen:", err)
       }

       // Graceful shutdown
       go func() {
           sigChan := make(chan os.Signal, 1)
           signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
           <-sigChan

           logger.Info("Shutting down gRPC server...")
           server.GracefulStop()
       }()

       logger.Info("Starting gRPC server on", cfg.GRPCPort)
       if err := server.Serve(lis); err != nil {
           logger.Fatal("Failed to serve:", err)
       }
   }
   ```

### Phase 2: Core Implementation (Days 4-6)

#### 1. Collaborator Service Handler

```go
// internal/grpc/handlers/collaborator/server.go
package collaborator

import (
    "context"

    pb "kisanlink-ecom/proto/gen/go/collaborator/v1"
    "kisanlink-ecom/internal/services/collaborator"
    "kisanlink-ecom/internal/aaa"
    "github.com/sirupsen/logrus"
)

type Server struct {
    pb.UnimplementedCollaboratorServiceServer

    service        collaborator.CollaboratorServiceInterface
    addressService *aaa.AddressService
    gstService     *GSTDeduplicationService
    logger         *logrus.Logger
}

func NewServer(
    service collaborator.CollaboratorServiceInterface,
    addressService *aaa.AddressService,
    gstService *GSTDeduplicationService,
    logger *logrus.Logger,
) *Server {
    return &Server{
        service:        service,
        addressService: addressService,
        gstService:     gstService,
        logger:         logger,
    }
}
```

#### 2. Create Endpoint Implementation

```go
// internal/grpc/handlers/collaborator/create.go
func (s *Server) CreateCollaborator(
    ctx context.Context,
    req *pb.CreateCollaboratorRequest,
) (*pb.CollaboratorResponse, error) {

    // Step 1: Extract user claims
    claims, err := extractUserClaims(ctx)
    if err != nil {
        return nil, status.Error(codes.Unauthenticated, "unauthorized")
    }

    // Step 2: GST validation and deduplication
    if req.BusinessInfo != nil && req.BusinessInfo.GstNumber != "" {
        if err := s.validateAndCheckGST(ctx, req.BusinessInfo.GstNumber); err != nil {
            return nil, err
        }
    }

    // Step 3: Create address in AAA
    addressIDs, err := s.createAddressesInAAA(ctx, req)
    if err != nil {
        s.logger.WithError(err).Warn("Failed to create addresses")
        // Continue without addresses
    }

    // Step 4: Create collaborator
    domainReq := s.toDomainCreateRequest(req, addressIDs)
    collaborator, err := s.service.CreateCollaborator(ctx, domainReq, claims.UserID)
    if err != nil {
        return nil, s.mapError(err)
    }

    // Step 5: Update address entity IDs asynchronously
    if len(addressIDs) > 0 {
        go s.updateAddressEntities(collaborator.ID, addressIDs)
    }

    // Step 6: Build response
    return s.buildResponse(ctx, collaborator, req.Address != nil), nil
}

func (s *Server) validateAndCheckGST(ctx context.Context, gst string) error {
    // Validate format
    if err := s.gstService.ValidateGST(gst); err != nil {
        return status.Errorf(codes.InvalidArgument, "invalid GST: %v", err)
    }

    // Check for duplicates
    exists, existing, err := s.gstService.CheckGSTExists(ctx, gst)
    if err != nil {
        return status.Error(codes.Internal, "GST check failed")
    }

    if exists {
        return status.Errorf(codes.AlreadyExists,
            "GST already registered to collaborator %s", existing.ID)
    }

    return nil
}
```

#### 3. Update Endpoint Implementation

```go
// internal/grpc/handlers/collaborator/update.go
func (s *Server) UpdateCollaborator(
    ctx context.Context,
    req *pb.UpdateCollaboratorRequest,
) (*pb.CollaboratorResponse, error) {

    // Extract claims and validate permissions
    claims, err := extractUserClaims(ctx)
    if err != nil {
        return nil, status.Error(codes.Unauthenticated, "unauthorized")
    }

    // Check if user can update this collaborator
    if !s.canUpdate(ctx, claims, req.Id) {
        return nil, status.Error(codes.PermissionDenied, "insufficient permissions")
    }

    // Apply field mask
    updateReq := s.applyFieldMask(req)

    // Update collaborator
    collaborator, err := s.service.UpdateCollaborator(ctx, req.Id, updateReq, claims.UserID)
    if err != nil {
        return nil, s.mapError(err)
    }

    return s.buildResponse(ctx, collaborator, false), nil
}

func (s *Server) applyFieldMask(req *pb.UpdateCollaboratorRequest) *services.UpdateRequest {
    update := &services.UpdateRequest{}

    if req.FieldMask != nil {
        for _, path := range req.FieldMask.Paths {
            switch path {
            case "email":
                update.Email = req.Email
            case "phone":
                update.Phone = req.Phone
            case "first_name":
                update.FirstName = req.FirstName
            case "last_name":
                update.LastName = req.LastName
            case "business_info":
                update.BusinessInfo = s.mapBusinessInfo(req.BusinessInfo)
            // ... other fields
            }
        }
    }

    return update
}
```

#### 4. Deactivate Endpoint Implementation

```go
// internal/grpc/handlers/collaborator/deactivate.go
func (s *Server) DeactivateCollaborator(
    ctx context.Context,
    req *pb.DeactivateCollaboratorRequest,
) (*pb.StatusResponse, error) {

    claims, err := extractUserClaims(ctx)
    if err != nil {
        return nil, status.Error(codes.Unauthenticated, "unauthorized")
    }

    // Check permissions
    if !claims.HasRole("admin") && !claims.HasRole("manager") {
        return nil, status.Error(codes.PermissionDenied, "admin or manager role required")
    }

    // Determine new status
    newStatus := collaborator.StatusInactive
    if req.NewStatus == pb.CollaboratorStatus_COLLABORATOR_STATUS_SUSPENDED {
        newStatus = collaborator.StatusSuspended
    }

    // Deactivate collaborator
    err = s.service.DeactivateCollaborator(ctx, req.Id, newStatus, req.Reason, claims.UserID)
    if err != nil {
        return nil, s.mapError(err)
    }

    return &pb.StatusResponse{
        Success:   true,
        Message:   "Collaborator deactivated successfully",
        NewStatus: s.mapStatus(newStatus),
        UpdatedAt: timestamppb.Now(),
    }, nil
}
```

#### 5. Get Endpoint Implementation

```go
// internal/grpc/handlers/collaborator/get.go
func (s *Server) GetCollaborator(
    ctx context.Context,
    req *pb.GetCollaboratorRequest,
) (*pb.CollaboratorResponse, error) {

    // Get collaborator
    collaborator, err := s.service.GetCollaboratorByID(ctx, req.Id)
    if err != nil {
        return nil, s.mapError(err)
    }

    // Build response
    resp := s.buildResponse(ctx, collaborator, req.ExpandAddress)

    // Filter sensitive fields based on permissions
    claims, _ := extractUserClaims(ctx)
    resp = s.filterSensitiveFields(resp, claims)

    return resp, nil
}

func (s *Server) filterSensitiveFields(
    resp *pb.CollaboratorResponse,
    claims *auth.TokenClaims,
) *pb.CollaboratorResponse {
    // Hide sensitive information based on user role
    if !claims.HasRole("admin") {
        if resp.Collaborator.BusinessInfo != nil {
            resp.Collaborator.BusinessInfo.BankAccountNumber = ""
            resp.Collaborator.BusinessInfo.BankIfscCode = ""
        }
    }

    // Hide PII for non-owners
    if claims.UserID != resp.Collaborator.UserId && !claims.HasRole("manager") {
        resp.Collaborator.Email = maskEmail(resp.Collaborator.Email)
        resp.Collaborator.Phone = maskPhone(resp.Collaborator.Phone)
    }

    return resp
}
```

### Phase 3: AAA Integration (Days 7-9)

#### 1. AAA Service Setup

```go
// internal/aaa/setup.go
package aaa

func InitializeAAAIntegration(cfg *config.AAAConfig) (*AAAIntegration, error) {
    // Create connection pool
    poolConfig := ClientConfig{
        Addresses:      cfg.Addresses,
        MaxConnections: cfg.MaxConnections,
        ConnTimeout:    cfg.ConnTimeout,
        RequestTimeout: cfg.RequestTimeout,
        EnableTLS:      cfg.EnableTLS,
        CertFile:       cfg.CertFile,
    }

    pool, err := NewConnectionPool(poolConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to create AAA connection pool: %w", err)
    }

    // Initialize cache
    cache := NewAddressCache(5*time.Minute, 10*time.Minute)

    // Create address service
    addressService := NewAddressService(pool, cache, logrus.New())

    return &AAAIntegration{
        Pool:           pool,
        AddressService: addressService,
    }, nil
}
```

#### 2. Address Synchronization

```go
// internal/grpc/handlers/collaborator/address_sync.go
func (s *Server) syncAddressWithAAA(
    ctx context.Context,
    collaboratorID string,
    address *pb.CreateAddressRequest,
) (string, error) {

    // Create address in AAA
    aaaReq := &aaa.CreateAddressRequest{
        EntityID:    collaboratorID,
        EntityType:  "collaborator",
        Line1:       address.Line1,
        Line2:       address.Line2,
        City:        address.City,
        State:       address.State,
        Country:     address.Country,
        PostalCode:  address.PostalCode,
        Type:        s.mapAddressType(address.Type),
        IsPrimary:   address.IsPrimary,
    }

    // Add coordinates if available
    if address.Coordinates != nil {
        aaaReq.Coordinates = &aaa.GeoCoordinates{
            Latitude:  address.Coordinates.Latitude,
            Longitude: address.Coordinates.Longitude,
        }
    }

    // Create with retry
    var addressID string
    err := retry.Do(
        func() error {
            addr, err := s.addressService.CreateAddress(ctx, aaaReq)
            if err != nil {
                return err
            }
            addressID = addr.ID
            return nil
        },
        retry.Attempts(3),
        retry.Delay(100*time.Millisecond),
        retry.DelayType(retry.BackOffDelay),
    )

    return addressID, err
}
```

### Phase 4: GST Deduplication (Days 10-11)

#### 1. GST Service Implementation

```go
// internal/services/gst/deduplication.go
package gst

type DeduplicationService struct {
    repo      CollaboratorRepository
    cache     *cache.Cache
    validator *GSTValidator
}

func (s *DeduplicationService) ProcessGST(ctx context.Context, gst string) (*GSTInfo, error) {
    // Normalize GST
    normalized := s.normalizeGST(gst)

    // Validate format
    if err := s.validator.Validate(normalized); err != nil {
        return nil, err
    }

    // Check cache
    if info, found := s.cache.Get(normalized); found {
        return info.(*GSTInfo), nil
    }

    // Extract information
    info := s.extractGSTInfo(normalized)

    // Check for duplicates
    exists, err := s.repo.ExistsByGST(ctx, normalized)
    if err != nil {
        return nil, err
    }

    if exists {
        return nil, ErrDuplicateGST
    }

    // Cache result
    s.cache.Set(normalized, info, 5*time.Minute)

    return info, nil
}

func (s *DeduplicationService) extractGSTInfo(gst string) *GSTInfo {
    return &GSTInfo{
        GST:          gst,
        StateCode:    gst[0:2],
        PAN:          gst[2:12],
        EntityNumber: gst[12:13],
        CheckDigit:   gst[14:15],
        BusinessType: s.inferBusinessType(gst[2:12]),
    }
}
```

### Phase 5: Testing (Days 12-14)

#### 1. Unit Tests

```go
// internal/grpc/handlers/collaborator/server_test.go
package collaborator_test

func TestCreateCollaborator(t *testing.T) {
    // Setup
    mockService := &mocks.CollaboratorService{}
    mockAAA := &mocks.AAAService{}
    mockGST := &mocks.GSTService{}

    server := NewServer(mockService, mockAAA, mockGST, logrus.New())

    t.Run("successful creation", func(t *testing.T) {
        ctx := contextWithClaims(t, &auth.TokenClaims{
            UserID: "user-123",
            Roles:  []string{"admin"},
        })

        req := &pb.CreateCollaboratorRequest{
            UserId:    "user-456",
            Username:  "john.doe",
            Email:     "john@example.com",
            FirstName: "John",
            LastName:  "Doe",
            Type:      pb.CollaboratorType_COLLABORATOR_TYPE_VENDOR,
        }

        mockService.On("CreateCollaborator", mock.Anything, mock.Anything, "user-123").
            Return(&collaborator.Collaborator{
                ID:       "collab-789",
                UserID:   req.UserId,
                Username: req.Username,
            }, nil)

        resp, err := server.CreateCollaborator(ctx, req)

        assert.NoError(t, err)
        assert.NotNil(t, resp)
        assert.Equal(t, "collab-789", resp.Collaborator.Id)
    })

    t.Run("GST deduplication", func(t *testing.T) {
        ctx := contextWithClaims(t, &auth.TokenClaims{
            UserID: "user-123",
        })

        req := &pb.CreateCollaboratorRequest{
            UserId: "user-456",
            BusinessInfo: &pb.CreateBusinessInfoRequest{
                GstNumber: "22AAAAA0000A1Z5",
            },
        }

        mockGST.On("CheckGSTExists", mock.Anything, "22AAAAA0000A1Z5").
            Return(true, &collaborator.Collaborator{ID: "existing-123"}, nil)

        _, err := server.CreateCollaborator(ctx, req)

        assert.Error(t, err)
        st, ok := status.FromError(err)
        assert.True(t, ok)
        assert.Equal(t, codes.AlreadyExists, st.Code())
    })
}
```

#### 2. Integration Tests

```go
// tests/integration/grpc_test.go
package integration_test

func TestCollaboratorGRPCIntegration(t *testing.T) {
    // Start test server
    server := startTestGRPCServer(t)
    defer server.Stop()

    // Create client
    conn, err := grpc.Dial(server.Address(), grpc.WithInsecure())
    assert.NoError(t, err)
    defer conn.Close()

    client := pb.NewCollaboratorServiceClient(conn)

    t.Run("full lifecycle", func(t *testing.T) {
        ctx := createAuthContext(t)

        // Create collaborator
        createReq := &pb.CreateCollaboratorRequest{
            UserId:    "test-user",
            Username:  "test.user",
            Email:     "test@example.com",
            FirstName: "Test",
            LastName:  "User",
            Type:      pb.CollaboratorType_COLLABORATOR_TYPE_BUYER,
        }

        createResp, err := client.CreateCollaborator(ctx, createReq)
        assert.NoError(t, err)
        assert.NotEmpty(t, createResp.Collaborator.Id)

        collaboratorID := createResp.Collaborator.Id

        // Get collaborator
        getResp, err := client.GetCollaborator(ctx, &pb.GetCollaboratorRequest{
            Id: collaboratorID,
        })
        assert.NoError(t, err)
        assert.Equal(t, collaboratorID, getResp.Collaborator.Id)

        // Update collaborator
        updateReq := &pb.UpdateCollaboratorRequest{
            Id: collaboratorID,
            FieldMask: &fieldmaskpb.FieldMask{
                Paths: []string{"phone"},
            },
            Phone: "+919876543210",
        }

        updateResp, err := client.UpdateCollaborator(ctx, updateReq)
        assert.NoError(t, err)
        assert.Equal(t, "+919876543210", updateResp.Collaborator.Phone)

        // Deactivate collaborator
        deactivateResp, err := client.DeactivateCollaborator(ctx,
            &pb.DeactivateCollaboratorRequest{
                Id:     collaboratorID,
                Reason: "Test deactivation",
            })
        assert.NoError(t, err)
        assert.True(t, deactivateResp.Success)
    })
}
```

#### 3. Load Testing

```go
// tests/load/grpc_load_test.go
package load_test

func TestGRPCLoadTest(t *testing.T) {
    // Setup
    client := setupGRPCClient(t)
    ctx := createAuthContext(t)

    // Load test configuration
    concurrency := 100
    requests := 1000

    var wg sync.WaitGroup
    errors := make(chan error, requests)
    latencies := make(chan time.Duration, requests)

    // Execute load test
    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()

            for j := 0; j < requests/concurrency; j++ {
                start := time.Now()

                _, err := client.GetCollaborator(ctx, &pb.GetCollaboratorRequest{
                    Id: "test-collaborator-1",
                })

                latencies <- time.Since(start)
                if err != nil {
                    errors <- err
                }
            }
        }(i)
    }

    wg.Wait()
    close(errors)
    close(latencies)

    // Analyze results
    var totalLatency time.Duration
    var errorCount int
    latencySlice := []time.Duration{}

    for lat := range latencies {
        totalLatency += lat
        latencySlice = append(latencySlice, lat)
    }

    for range errors {
        errorCount++
    }

    // Calculate metrics
    avgLatency := totalLatency / time.Duration(requests)
    sort.Slice(latencySlice, func(i, j int) bool {
        return latencySlice[i] < latencySlice[j]
    })
    p99Latency := latencySlice[int(float64(len(latencySlice))*0.99)]

    // Assert performance requirements
    assert.Less(t, avgLatency, 50*time.Millisecond, "Average latency exceeds 50ms")
    assert.Less(t, p99Latency, 100*time.Millisecond, "P99 latency exceeds 100ms")
    assert.Less(t, float64(errorCount)/float64(requests), 0.01, "Error rate exceeds 1%")

    t.Logf("Load test results: Avg latency: %v, P99: %v, Errors: %d/%d",
        avgLatency, p99Latency, errorCount, requests)
}
```

### Phase 6: Deployment (Days 15-16)

#### 1. Docker Configuration

```dockerfile
# Dockerfile.grpc
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build gRPC server
RUN CGO_ENABLED=0 GOOS=linux go build -o grpc-server cmd/grpc/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/grpc-server .
COPY --from=builder /app/config ./config

EXPOSE 50051

CMD ["./grpc-server"]
```

#### 2. Kubernetes Deployment

```yaml
# k8s/grpc-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: collaborator-grpc
  namespace: kisanlink
spec:
  replicas: 3
  selector:
    matchLabels:
      app: collaborator-grpc
  template:
    metadata:
      labels:
        app: collaborator-grpc
    spec:
      containers:
      - name: grpc-server
        image: kisanlink/collaborator-grpc:latest
        ports:
        - containerPort: 50051
          name: grpc
        - containerPort: 9090
          name: metrics
        env:
        - name: DB_CONNECTION
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: connection-string
        - name: AAA_SERVICE_URL
          value: "aaa-service.kisanlink.svc.cluster.local:50051"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          grpc:
            port: 50051
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          grpc:
            port: 50051
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: collaborator-grpc
  namespace: kisanlink
spec:
  selector:
    app: collaborator-grpc
  ports:
  - name: grpc
    port: 50051
    targetPort: 50051
  - name: metrics
    port: 9090
    targetPort: 9090
  type: ClusterIP
```

## Configuration

### 1. Application Configuration

```yaml
# config/grpc.yaml
server:
  port: 50051
  max_recv_msg_size: 10485760  # 10MB
  max_send_msg_size: 10485760  # 10MB
  max_concurrent_streams: 1000
  keepalive:
    time: 10s
    timeout: 5s

database:
  driver: postgres
  host: localhost
  port: 5432
  database: kisanlink
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m

aaa:
  addresses:
    - aaa-service-1:50051
    - aaa-service-2:50051
  max_connections: 10
  conn_timeout: 5s
  request_timeout: 10s
  enable_tls: true
  cert_file: /certs/aaa.crt

cache:
  redis:
    address: localhost:6379
    password: ""
    db: 0
    pool_size: 10

interceptors:
  rate_limit:
    global_limit: 100
    per_user_limit: 10
    burst_size: 20

  auth:
    excluded_methods:
      - /kisanlink.collaborator.v1.CollaboratorService/HealthCheck

observability:
  metrics:
    enabled: true
    port: 9090

  tracing:
    enabled: true
    sampling_rate: 0.1
    jaeger_endpoint: http://jaeger:14268/api/traces

  logging:
    level: info
    format: json
```

### 2. Environment Variables

```bash
# .env.production
GRPC_PORT=50051
DB_HOST=postgres.production.svc.cluster.local
DB_PORT=5432
DB_NAME=kisanlink_prod
DB_USER=collaborator_service
DB_PASSWORD=${DB_PASSWORD}

AAA_SERVICE_URLS=aaa-1.prod:50051,aaa-2.prod:50051
AAA_TLS_ENABLED=true
AAA_CERT_PATH=/etc/certs/aaa.crt

REDIS_HOST=redis.production.svc.cluster.local
REDIS_PORT=6379
REDIS_PASSWORD=${REDIS_PASSWORD}

LOG_LEVEL=info
METRICS_ENABLED=true
TRACING_ENABLED=true
JAEGER_ENDPOINT=http://jaeger.monitoring:14268/api/traces
```

## Security Checklist

- ✅ **Authentication**: JWT validation via AAA service
- ✅ **Authorization**: RBAC with permission checks
- ✅ **Rate Limiting**: Global and per-user limits
- ✅ **Input Validation**: Proto validation rules
- ✅ **Audit Logging**: All mutations logged
- ✅ **TLS/mTLS**: Encrypted communication
- ✅ **Secret Management**: Environment variables and K8s secrets
- ✅ **Field Filtering**: Sensitive data protection
- ✅ **Error Handling**: No sensitive data in errors
- ✅ **OWASP Compliance**: Following OWASP guidelines

## Monitoring Dashboard

```yaml
# Grafana dashboard queries

# Request Rate
rate(grpc_requests_total[5m])

# Error Rate
rate(grpc_requests_total{status!="OK"}[5m]) / rate(grpc_requests_total[5m])

# P95 Latency
histogram_quantile(0.95, rate(grpc_request_duration_seconds_bucket[5m]))

# AAA Service Health
aaa_circuit_breaker_state

# GST Deduplication Rate
rate(gst_duplicates_detected_total[5m])

# Cache Hit Rate
rate(aaa_cache_hits_total[5m]) / (rate(aaa_cache_hits_total[5m]) + rate(aaa_cache_misses_total[5m]))
```

## Rollback Strategy

1. **Blue-Green Deployment**: Maintain two identical production environments
2. **Canary Release**: Route 10% traffic to new version initially
3. **Feature Flags**: Disable new features without deployment
4. **Database Migration**: Backward-compatible schema changes
5. **Circuit Breaker**: Automatic failover for AAA service issues

## Success Criteria

- ✅ All 4 gRPC endpoints functional
- ✅ P99 latency < 100ms for reads, < 200ms for writes
- ✅ 99.9% uptime SLA
- ✅ GST deduplication working correctly
- ✅ AAA integration stable with circuit breaker
- ✅ Load test passing (1000 RPS)
- ✅ Security audit passed
- ✅ Documentation complete

## Contact and Support

- **Architecture Team**: architecture@kisanlink.com
- **DevOps Team**: devops@kisanlink.com
- **On-Call**: Use PagerDuty escalation
- **Documentation**: https://docs.kisanlink.com/grpc