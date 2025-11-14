// Package grpc provides gRPC server initialization and configuration
package grpc

import (
	"github.com/Kisanlink/kisanlink-ecom/internal/grpc/interceptors"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// BuildInterceptorChain builds the 10-layer interceptor chain in the correct order
// Order (outermost to innermost):
// 1. Recovery → 2. Request ID → 3. Logging → 4. Metrics → 5. Tracing →
// 6. Rate Limit → 7. Auth → 8. Authorization → 9. Validation → 10. Audit
func BuildInterceptorChain(logger *logrus.Logger, jwtValidator interceptors.JWTValidator) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		interceptors.RecoveryUnaryInterceptor(logger),   // Layer 1: Panic recovery
		interceptors.RequestIDUnaryInterceptor(),        // Layer 2: Request ID generation
		interceptors.LoggingUnaryInterceptor(logger),    // Layer 3: Request/response logging
		interceptors.MetricsUnaryInterceptor(),          // Layer 4: Prometheus metrics
		interceptors.TracingUnaryInterceptor(),          // Layer 5: OpenTelemetry tracing
		interceptors.RateLimitUnaryInterceptor(100, 10), // Layer 6: Rate limiting (100 global, 10 per user)
		interceptors.AuthUnaryInterceptor(jwtValidator), // Layer 7: JWT authentication
		interceptors.AuthorizationUnaryInterceptor(),    // Layer 8: RBAC authorization
		interceptors.ValidationUnaryInterceptor(),       // Layer 9: Proto message validation
		interceptors.AuditUnaryInterceptor(logger),      // Layer 10: Audit logging
	}
}
