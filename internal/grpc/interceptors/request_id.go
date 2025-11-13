package interceptors

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type requestIDKey struct{}

// RequestIDUnaryInterceptor generates and injects request ID into context
func RequestIDUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Generate new request ID
		requestID := uuid.New().String()

		// Add to context
		ctx = context.WithValue(ctx, requestIDKey{}, requestID)

		// Add to response headers
		if err := grpc.SetHeader(ctx, metadata.Pairs("x-request-id", requestID)); err != nil {
			// Log error but don't fail request
		}

		return handler(ctx, req)
	}
}

// GetRequestID retrieves request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey{}).(string); ok {
		return requestID
	}
	return ""
}
