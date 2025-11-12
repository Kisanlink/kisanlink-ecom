# Collaborator gRPC Service Interceptor Architecture

## Overview

The gRPC interceptor chain provides cross-cutting concerns for all RPC calls, implementing authentication, validation, logging, metrics, and error handling in a consistent manner.

## Interceptor Chain Order

```
Client Request
     │
     ▼
┌─────────────────────┐
│ 1. Recovery         │ - Panic recovery and graceful error handling
└─────────────────────┘
     │
     ▼
┌─────────────────────┐
│ 2. Request ID       │ - Generate unique request ID for tracing
└─────────────────────┘
     │
     ▼
┌─────────────────────┐
│ 3. Logging          │ - Log request entry with context
└─────────────────────┘
     │
     ▼
┌─────────────────────┐
│ 4. Metrics          │ - Start request timer, increment counters
└─────────────────────┘
     │
     ▼
┌─────────────────────┐
│ 5. Tracing          │ - Start distributed trace span
└─────────────────────┘
     │
     ▼
┌─────────────────────┐
│ 6. Rate Limiting    │ - Apply per-user/global rate limits
└─────────────────────┘
     │
     ▼
┌─────────────────────┐
│ 7. Authentication   │ - Validate JWT token, extract claims
└─────────────────────┘
     │
     ▼
┌─────────────────────┐
│ 8. Authorization    │ - Check RBAC permissions
└─────────────────────┘
     │
     ▼
┌─────────────────────┐
│ 9. Validation       │ - Validate request proto messages
└─────────────────────┘
     │
     ▼
┌─────────────────────┐
│ 10. Audit           │ - Log audit trail for mutations
└─────────────────────┘
     │
     ▼
┌─────────────────────┐
│ Service Handler     │ - Actual business logic
└─────────────────────┘
     │
     ▼
┌─────────────────────┐
│ Response Filter     │ - Filter sensitive fields
└─────────────────────┘
     │
     ▼
Client Response
```

## Interceptor Implementations

### 1. Recovery Interceptor

```go
package interceptors

import (
    "context"
    "runtime/debug"

    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "github.com/sirupsen/logrus"
)

// RecoveryInterceptor recovers from panics and returns a proper gRPC error
func RecoveryInterceptor(logger *logrus.Logger) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler) (resp interface{}, err error) {

        defer func() {
            if r := recover(); r != nil {
                // Log the panic with stack trace
                logger.WithFields(logrus.Fields{
                    "method": info.FullMethod,
                    "panic":  r,
                    "stack":  string(debug.Stack()),
                }).Error("Panic recovered in gRPC handler")

                // Return internal error to client
                err = status.Errorf(codes.Internal,
                    "internal error occurred, request ID: %s",
                    getRequestID(ctx))
            }
        }()

        return handler(ctx, req)
    }
}
```

### 2. Request ID Interceptor

```go
package interceptors

import (
    "context"

    "github.com/google/uuid"
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
)

const (
    RequestIDKey     = "x-request-id"
    RequestIDCtxKey  = "request_id"
)

// RequestIDInterceptor generates or extracts request ID
func RequestIDInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler) (interface{}, error) {

        requestID := extractRequestID(ctx)
        if requestID == "" {
            requestID = uuid.New().String()
        }

        // Add to context
        ctx = context.WithValue(ctx, RequestIDCtxKey, requestID)

        // Add to response headers
        header := metadata.Pairs(RequestIDKey, requestID)
        grpc.SendHeader(ctx, header)

        return handler(ctx, req)
    }
}

func extractRequestID(ctx context.Context) string {
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return ""
    }

    ids := md.Get(RequestIDKey)
    if len(ids) > 0 {
        return ids[0]
    }

    return ""
}

func getRequestID(ctx context.Context) string {
    if id, ok := ctx.Value(RequestIDCtxKey).(string); ok {
        return id
    }
    return "unknown"
}
```

### 3. Logging Interceptor

```go
package interceptors

import (
    "context"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/peer"
    "google.golang.org/grpc/status"
    "github.com/sirupsen/logrus"
)

// LoggingInterceptor logs all gRPC requests
func LoggingInterceptor(logger *logrus.Logger) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler) (interface{}, error) {

        startTime := time.Now()
        requestID := getRequestID(ctx)

        // Get client info
        clientIP := "unknown"
        if p, ok := peer.FromContext(ctx); ok {
            clientIP = p.Addr.String()
        }

        // Log request start
        logger.WithFields(logrus.Fields{
            "request_id": requestID,
            "method":     info.FullMethod,
            "client_ip":  clientIP,
            "user_id":    getUserID(ctx),
        }).Info("gRPC request started")

        // Execute handler
        resp, err := handler(ctx, req)

        // Calculate duration
        duration := time.Since(startTime)

        // Get status code
        code := codes.OK
        if err != nil {
            if st, ok := status.FromError(err); ok {
                code = st.Code()
            } else {
                code = codes.Unknown
            }
        }

        // Log request completion
        fields := logrus.Fields{
            "request_id": requestID,
            "method":     info.FullMethod,
            "duration":   duration.Milliseconds(),
            "status":     code.String(),
        }

        if err != nil {
            fields["error"] = err.Error()
            logger.WithFields(fields).Error("gRPC request failed")
        } else {
            logger.WithFields(fields).Info("gRPC request completed")
        }

        return resp, err
    }
}
```

### 4. Metrics Interceptor

```go
package interceptors

import (
    "context"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/status"
    "github.com/prometheus/client_golang/prometheus"
)

var (
    grpcRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "grpc_requests_total",
            Help: "Total number of gRPC requests",
        },
        []string{"method", "status"},
    )

    grpcRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "grpc_request_duration_seconds",
            Help:    "Duration of gRPC requests in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "status"},
    )

    grpcRequestsInFlight = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "grpc_requests_in_flight",
            Help: "Number of gRPC requests currently being processed",
        },
        []string{"method"},
    )
)

func init() {
    prometheus.MustRegister(grpcRequestsTotal)
    prometheus.MustRegister(grpcRequestDuration)
    prometheus.MustRegister(grpcRequestsInFlight)
}

// MetricsInterceptor collects Prometheus metrics
func MetricsInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler) (interface{}, error) {

        startTime := time.Now()

        // Increment in-flight gauge
        grpcRequestsInFlight.WithLabelValues(info.FullMethod).Inc()
        defer grpcRequestsInFlight.WithLabelValues(info.FullMethod).Dec()

        // Execute handler
        resp, err := handler(ctx, req)

        // Get status
        statusCode := "OK"
        if err != nil {
            if st, ok := status.FromError(err); ok {
                statusCode = st.Code().String()
            } else {
                statusCode = "UNKNOWN"
            }
        }

        // Update metrics
        duration := time.Since(startTime).Seconds()
        grpcRequestsTotal.WithLabelValues(info.FullMethod, statusCode).Inc()
        grpcRequestDuration.WithLabelValues(info.FullMethod, statusCode).Observe(duration)

        return resp, err
    }
}
```

### 5. Tracing Interceptor

```go
package interceptors

import (
    "context"

    "google.golang.org/grpc"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("collaborator-grpc")

// TracingInterceptor adds distributed tracing
func TracingInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler) (interface{}, error) {

        // Start span
        ctx, span := tracer.Start(ctx, info.FullMethod,
            trace.WithSpanKind(trace.SpanKindServer))
        defer span.End()

        // Add attributes
        span.SetAttributes(
            attribute.String("rpc.system", "grpc"),
            attribute.String("rpc.method", info.FullMethod),
            attribute.String("request.id", getRequestID(ctx)),
            attribute.String("user.id", getUserID(ctx)),
        )

        // Execute handler
        resp, err := handler(ctx, req)

        // Record error if present
        if err != nil {
            span.RecordError(err)
            span.SetAttributes(attribute.String("error.message", err.Error()))
        }

        return resp, err
    }
}
```

### 6. Rate Limiting Interceptor

```go
package interceptors

import (
    "context"
    "fmt"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "golang.org/x/time/rate"
)

type RateLimiterConfig struct {
    GlobalLimit      rate.Limit // Global requests per second
    PerUserLimit     rate.Limit // Per-user requests per second
    BurstSize        int        // Token bucket burst size
}

type RateLimiter struct {
    global    *rate.Limiter
    perUser   map[string]*rate.Limiter
    config    RateLimiterConfig
    mu        sync.RWMutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
    return &RateLimiter{
        global:  rate.NewLimiter(config.GlobalLimit, config.BurstSize),
        perUser: make(map[string]*rate.Limiter),
        config:  config,
    }
}

// RateLimitingInterceptor applies rate limiting
func RateLimitingInterceptor(limiter *RateLimiter) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler) (interface{}, error) {

        // Check global rate limit
        if !limiter.global.Allow() {
            return nil, status.Error(codes.ResourceExhausted,
                "global rate limit exceeded")
        }

        // Check per-user rate limit
        userID := getUserID(ctx)
        if userID != "" {
            userLimiter := limiter.getUserLimiter(userID)
            if !userLimiter.Allow() {
                return nil, status.Error(codes.ResourceExhausted,
                    "user rate limit exceeded")
            }
        }

        return handler(ctx, req)
    }
}

func (r *RateLimiter) getUserLimiter(userID string) *rate.Limiter {
    r.mu.RLock()
    limiter, exists := r.perUser[userID]
    r.mu.RUnlock()

    if exists {
        return limiter
    }

    r.mu.Lock()
    defer r.mu.Unlock()

    // Double-check after acquiring write lock
    if limiter, exists := r.perUser[userID]; exists {
        return limiter
    }

    // Create new limiter for user
    limiter = rate.NewLimiter(r.config.PerUserLimit, r.config.BurstSize)
    r.perUser[userID] = limiter

    // Clean up old limiters periodically
    if len(r.perUser) > 10000 {
        r.cleanup()
    }

    return limiter
}

func (r *RateLimiter) cleanup() {
    // Remove limiters that haven't been used recently
    cutoff := time.Now().Add(-1 * time.Hour)
    for userID, limiter := range r.perUser {
        if limiter.Tokens() == float64(r.config.BurstSize) {
            delete(r.perUser, userID)
        }
    }
}
```

### 7. Authentication Interceptor

```go
package interceptors

import (
    "context"
    "strings"

    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/status"
    "kisanlink-ecom/internal/auth"
)

const (
    AuthorizationHeader = "authorization"
    BearerPrefix       = "Bearer "
    UserClaimsKey      = "user_claims"
)

// AuthInterceptor validates JWT tokens
func AuthInterceptor(authClient auth.Client, excludedMethods []string) grpc.UnaryServerInterceptor {
    excludeMap := make(map[string]bool)
    for _, method := range excludedMethods {
        excludeMap[method] = true
    }

    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler) (interface{}, error) {

        // Skip auth for excluded methods (like health check)
        if excludeMap[info.FullMethod] {
            return handler(ctx, req)
        }

        // Extract token from metadata
        token, err := extractToken(ctx)
        if err != nil {
            return nil, status.Error(codes.Unauthenticated, "missing or invalid token")
        }

        // Validate token
        claims, err := authClient.ValidateToken(ctx, token)
        if err != nil {
            return nil, status.Error(codes.Unauthenticated, "token validation failed")
        }

        // Check if token is expired
        if claims.IsExpired() {
            return nil, status.Error(codes.Unauthenticated, "token expired")
        }

        // Add claims to context
        ctx = context.WithValue(ctx, UserClaimsKey, claims)

        return handler(ctx, req)
    }
}

func extractToken(ctx context.Context) (string, error) {
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return "", fmt.Errorf("missing metadata")
    }

    authHeaders := md.Get(AuthorizationHeader)
    if len(authHeaders) == 0 {
        return "", fmt.Errorf("missing authorization header")
    }

    authHeader := authHeaders[0]
    if !strings.HasPrefix(authHeader, BearerPrefix) {
        return "", fmt.Errorf("invalid authorization header format")
    }

    return strings.TrimPrefix(authHeader, BearerPrefix), nil
}

func getUserClaims(ctx context.Context) (*auth.TokenClaims, error) {
    claims, ok := ctx.Value(UserClaimsKey).(*auth.TokenClaims)
    if !ok {
        return nil, fmt.Errorf("user claims not found in context")
    }
    return claims, nil
}

func getUserID(ctx context.Context) string {
    claims, err := getUserClaims(ctx)
    if err != nil {
        return ""
    }
    return claims.UserID
}
```

### 8. Authorization Interceptor

```go
package interceptors

import (
    "context"
    "fmt"

    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "kisanlink-ecom/internal/auth"
)

type Permission struct {
    Resource string
    Action   string
}

type MethodPermissions map[string]Permission

// AuthorizationInterceptor checks RBAC permissions
func AuthorizationInterceptor(authClient auth.Client, permissions MethodPermissions) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler) (interface{}, error) {

        // Get required permission for this method
        perm, ok := permissions[info.FullMethod]
        if !ok {
            // No specific permission required
            return handler(ctx, req)
        }

        // Get user claims
        claims, err := getUserClaims(ctx)
        if err != nil {
            return nil, status.Error(codes.PermissionDenied, "unauthorized")
        }

        // Check permission with AAA service
        allowed, err := authClient.EvaluatePermission(ctx,
            claims.UserID, perm.Resource, perm.Action)
        if err != nil {
            return nil, status.Error(codes.Internal, "permission check failed")
        }

        if !allowed {
            return nil, status.Errorf(codes.PermissionDenied,
                "insufficient permissions for %s:%s", perm.Resource, perm.Action)
        }

        return handler(ctx, req)
    }
}

// Define method permissions
var CollaboratorMethodPermissions = MethodPermissions{
    "/kisanlink.collaborator.v1.CollaboratorService/CreateCollaborator": {
        Resource: "collaborator",
        Action:   "create",
    },
    "/kisanlink.collaborator.v1.CollaboratorService/UpdateCollaborator": {
        Resource: "collaborator",
        Action:   "update",
    },
    "/kisanlink.collaborator.v1.CollaboratorService/DeactivateCollaborator": {
        Resource: "collaborator",
        Action:   "delete",
    },
    "/kisanlink.collaborator.v1.CollaboratorService/GetCollaborator": {
        Resource: "collaborator",
        Action:   "read",
    },
    "/kisanlink.collaborator.v1.CollaboratorService/ListCollaborators": {
        Resource: "collaborator",
        Action:   "read",
    },
    "/kisanlink.collaborator.v1.CollaboratorService/VerifyCollaborator": {
        Resource: "collaborator",
        Action:   "verify",
    },
}
```

### 9. Validation Interceptor

```go
package interceptors

import (
    "context"
    "fmt"

    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// Validator interface for proto message validation
type Validator interface {
    Validate() error
}

// ValidationInterceptor validates request messages
func ValidationInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler) (interface{}, error) {

        // Check if request implements Validator interface
        if v, ok := req.(Validator); ok {
            if err := v.Validate(); err != nil {
                return nil, status.Errorf(codes.InvalidArgument,
                    "validation failed: %v", err)
            }
        }

        // Execute handler
        resp, err := handler(ctx, req)
        if err != nil {
            return nil, err
        }

        // Validate response if applicable
        if v, ok := resp.(Validator); ok {
            if err := v.Validate(); err != nil {
                // Log the error but don't fail the request
                // This indicates a server-side issue
                fmt.Printf("Response validation failed: %v\n", err)
            }
        }

        return resp, nil
    }
}
```

### 10. Audit Interceptor

```go
package interceptors

import (
    "context"
    "encoding/json"
    "time"

    "google.golang.org/grpc"
    "github.com/sirupsen/logrus"
)

type AuditLogger interface {
    LogAuditEvent(event AuditEvent) error
}

type AuditEvent struct {
    Timestamp   time.Time              `json:"timestamp"`
    RequestID   string                 `json:"request_id"`
    UserID      string                 `json:"user_id"`
    Method      string                 `json:"method"`
    Resource    string                 `json:"resource"`
    Action      string                 `json:"action"`
    Request     interface{}            `json:"request,omitempty"`
    Response    interface{}            `json:"response,omitempty"`
    Error       string                 `json:"error,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Methods that should be audited
var auditableMethods = map[string]bool{
    "/kisanlink.collaborator.v1.CollaboratorService/CreateCollaborator":     true,
    "/kisanlink.collaborator.v1.CollaboratorService/UpdateCollaborator":     true,
    "/kisanlink.collaborator.v1.CollaboratorService/DeactivateCollaborator": true,
    "/kisanlink.collaborator.v1.CollaboratorService/VerifyCollaborator":     true,
}

// AuditInterceptor logs audit events for mutations
func AuditInterceptor(logger AuditLogger) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler) (interface{}, error) {

        // Check if this method should be audited
        if !auditableMethods[info.FullMethod] {
            return handler(ctx, req)
        }

        event := AuditEvent{
            Timestamp: time.Now(),
            RequestID: getRequestID(ctx),
            UserID:    getUserID(ctx),
            Method:    info.FullMethod,
            Request:   sanitizeRequest(req),
            Metadata:  extractMetadata(ctx),
        }

        // Extract resource and action from method name
        event.Resource, event.Action = parseMethodName(info.FullMethod)

        // Execute handler
        resp, err := handler(ctx, req)

        // Update event with response or error
        if err != nil {
            event.Error = err.Error()
        } else {
            event.Response = sanitizeResponse(resp)
        }

        // Log audit event asynchronously
        go func() {
            if err := logger.LogAuditEvent(event); err != nil {
                logrus.WithError(err).Error("Failed to log audit event")
            }
        }()

        return resp, err
    }
}

// sanitizeRequest removes sensitive fields from request
func sanitizeRequest(req interface{}) interface{} {
    // Convert to JSON and back to remove sensitive fields
    data, err := json.Marshal(req)
    if err != nil {
        return nil
    }

    var sanitized map[string]interface{}
    if err := json.Unmarshal(data, &sanitized); err != nil {
        return nil
    }

    // Remove sensitive fields
    delete(sanitized, "password")
    delete(sanitized, "bank_account_number")
    delete(sanitized, "pan_number")

    return sanitized
}

// sanitizeResponse removes sensitive fields from response
func sanitizeResponse(resp interface{}) interface{} {
    // Similar to sanitizeRequest
    return sanitizeRequest(resp)
}
```

## Interceptor Chain Configuration

```go
package grpc

import (
    "kisanlink-ecom/internal/auth"
    "kisanlink-ecom/internal/grpc/interceptors"

    "google.golang.org/grpc"
    "github.com/sirupsen/logrus"
)

// NewServer creates a new gRPC server with configured interceptors
func NewServer(
    logger *logrus.Logger,
    authClient auth.Client,
    auditLogger interceptors.AuditLogger,
) *grpc.Server {

    // Configure rate limiter
    rateLimiter := interceptors.NewRateLimiter(interceptors.RateLimiterConfig{
        GlobalLimit:  100,  // 100 req/s globally
        PerUserLimit: 10,   // 10 req/s per user
        BurstSize:    20,   // Allow burst of 20 requests
    })

    // Methods that don't require authentication
    excludedFromAuth := []string{
        "/kisanlink.collaborator.v1.CollaboratorService/HealthCheck",
    }

    // Create interceptor chain
    interceptors := []grpc.UnaryServerInterceptor{
        interceptors.RecoveryInterceptor(logger),
        interceptors.RequestIDInterceptor(),
        interceptors.LoggingInterceptor(logger),
        interceptors.MetricsInterceptor(),
        interceptors.TracingInterceptor(),
        interceptors.RateLimitingInterceptor(rateLimiter),
        interceptors.AuthInterceptor(authClient, excludedFromAuth),
        interceptors.AuthorizationInterceptor(authClient,
            interceptors.CollaboratorMethodPermissions),
        interceptors.ValidationInterceptor(),
        interceptors.AuditInterceptor(auditLogger),
    }

    // Chain interceptors
    opts := []grpc.ServerOption{
        grpc.ChainUnaryInterceptor(interceptors...),
        grpc.MaxRecvMsgSize(10 * 1024 * 1024), // 10MB max message size
        grpc.MaxSendMsgSize(10 * 1024 * 1024),
        grpc.MaxConcurrentStreams(1000),
    }

    return grpc.NewServer(opts...)
}
```

## Testing Interceptors

```go
package interceptors_test

import (
    "context"
    "testing"

    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/status"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestAuthInterceptor(t *testing.T) {
    tests := []struct {
        name          string
        setupContext  func() context.Context
        setupAuth     func(*MockAuthClient)
        expectedError codes.Code
    }{
        {
            name: "valid token",
            setupContext: func() context.Context {
                md := metadata.New(map[string]string{
                    "authorization": "Bearer valid-token",
                })
                return metadata.NewIncomingContext(context.Background(), md)
            },
            setupAuth: func(m *MockAuthClient) {
                m.On("ValidateToken", mock.Anything, "valid-token").
                    Return(&auth.TokenClaims{
                        UserID: "user-123",
                        ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
                    }, nil)
            },
            expectedError: codes.OK,
        },
        {
            name: "missing token",
            setupContext: func() context.Context {
                return context.Background()
            },
            setupAuth: func(m *MockAuthClient) {},
            expectedError: codes.Unauthenticated,
        },
        {
            name: "expired token",
            setupContext: func() context.Context {
                md := metadata.New(map[string]string{
                    "authorization": "Bearer expired-token",
                })
                return metadata.NewIncomingContext(context.Background(), md)
            },
            setupAuth: func(m *MockAuthClient) {
                m.On("ValidateToken", mock.Anything, "expired-token").
                    Return(&auth.TokenClaims{
                        UserID: "user-123",
                        ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(),
                    }, nil)
            },
            expectedError: codes.Unauthenticated,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            mockAuth := new(MockAuthClient)
            tt.setupAuth(mockAuth)

            interceptor := AuthInterceptor(mockAuth, []string{})

            // Create test handler
            handler := func(ctx context.Context, req interface{}) (interface{}, error) {
                // Check if claims are in context
                claims, ok := ctx.Value(UserClaimsKey).(*auth.TokenClaims)
                if tt.expectedError == codes.OK {
                    assert.True(t, ok)
                    assert.NotNil(t, claims)
                }
                return "success", nil
            }

            // Execute
            ctx := tt.setupContext()
            info := &grpc.UnaryServerInfo{
                FullMethod: "/test/Method",
            }

            resp, err := interceptor(ctx, nil, info, handler)

            // Assert
            if tt.expectedError != codes.OK {
                assert.Error(t, err)
                st, ok := status.FromError(err)
                assert.True(t, ok)
                assert.Equal(t, tt.expectedError, st.Code())
            } else {
                assert.NoError(t, err)
                assert.Equal(t, "success", resp)
            }

            mockAuth.AssertExpectations(t)
        })
    }
}
```

## Performance Monitoring

```yaml
# Prometheus queries for monitoring

# Request rate by method
rate(grpc_requests_total[5m])

# P95 latency by method
histogram_quantile(0.95,
  rate(grpc_request_duration_seconds_bucket[5m])
)

# Error rate
rate(grpc_requests_total{status!="OK"}[5m]) /
rate(grpc_requests_total[5m])

# Rate limit violations
rate(grpc_requests_total{status="RESOURCE_EXHAUSTED"}[5m])

# Authentication failures
rate(grpc_requests_total{status="UNAUTHENTICATED"}[5m])

# Authorization failures
rate(grpc_requests_total{status="PERMISSION_DENIED"}[5m])

# Concurrent requests
grpc_requests_in_flight
```

## Security Checklist

- ✅ JWT token validation with AAA service
- ✅ Role-based access control (RBAC)
- ✅ Rate limiting (global and per-user)
- ✅ Request validation using protoc-gen-validate
- ✅ Audit logging for all mutations
- ✅ Sensitive field filtering in responses
- ✅ Request ID tracking for debugging
- ✅ Panic recovery with proper error handling
- ✅ Distributed tracing support
- ✅ Prometheus metrics collection
- ✅ Context timeout propagation
- ✅ TLS/mTLS support (configured at server level)

## Configuration Example

```yaml
grpc:
  server:
    port: 50051
    max_recv_msg_size: 10485760  # 10MB
    max_send_msg_size: 10485760  # 10MB
    max_concurrent_streams: 1000

  interceptors:
    rate_limit:
      global_limit: 100
      per_user_limit: 10
      burst_size: 20

    auth:
      excluded_methods:
        - /kisanlink.collaborator.v1.CollaboratorService/HealthCheck

    metrics:
      enabled: true

    tracing:
      enabled: true
      sampling_rate: 0.1

    audit:
      enabled: true
      async: true
      buffer_size: 1000
```