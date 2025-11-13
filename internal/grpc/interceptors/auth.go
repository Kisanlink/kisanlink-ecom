package interceptors

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type userClaimsKey struct{}

// JWTValidator defines the interface for JWT validation
type JWTValidator interface {
	ValidateToken(ctx context.Context, token string) (map[string]interface{}, error)
}

// AuthUnaryInterceptor validates JWT tokens and adds claims to context
func AuthUnaryInterceptor(jwtValidator JWTValidator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip auth for public endpoints
		if isPublicEndpoint(info.FullMethod) {
			return handler(ctx, req)
		}

		// Extract token from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "Missing metadata")
		}

		tokens := md.Get("authorization")
		if len(tokens) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "Missing authorization token")
		}

		// Remove "Bearer " prefix if present
		token := strings.TrimPrefix(tokens[0], "Bearer ")

		// Validate JWT
		claims, err := jwtValidator.ValidateToken(ctx, token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "Invalid token: %v", err)
		}

		// Add claims to context
		ctx = context.WithValue(ctx, userClaimsKey{}, claims)

		return handler(ctx, req)
	}
}

// GetUserClaims retrieves user claims from context
func GetUserClaims(ctx context.Context) map[string]interface{} {
	if claims, ok := ctx.Value(userClaimsKey{}).(map[string]interface{}); ok {
		return claims
	}
	return nil
}

func isPublicEndpoint(method string) bool {
	publicEndpoints := []string{
		"/grpc.health.v1.Health/Check",
		"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
	}

	for _, endpoint := range publicEndpoints {
		if method == endpoint {
			return true
		}
	}
	return false
}
