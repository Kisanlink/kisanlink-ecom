package interceptors

import (
	"context"
	"sync"

	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RateLimitUnaryInterceptor implements global and per-user rate limiting
func RateLimitUnaryInterceptor(globalLimit, perUserLimit rate.Limit) grpc.UnaryServerInterceptor {
	globalLimiter := rate.NewLimiter(globalLimit, int(globalLimit))
	userLimiters := sync.Map{}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Check global rate limit
		if !globalLimiter.Allow() {
			return nil, status.Errorf(codes.ResourceExhausted, "Global rate limit exceeded")
		}

		// Check per-user rate limit (if user is authenticated)
		userID := getUserIDFromContext(ctx)
		if userID != "" {
			limiter := getUserLimiter(userID, perUserLimit, &userLimiters)
			if !limiter.Allow() {
				return nil, status.Errorf(codes.ResourceExhausted, "User rate limit exceeded")
			}
		}

		return handler(ctx, req)
	}
}

func getUserLimiter(userID string, limit rate.Limit, limiters *sync.Map) *rate.Limiter {
	if limiter, ok := limiters.Load(userID); ok {
		return limiter.(*rate.Limiter)
	}

	limiter := rate.NewLimiter(limit, int(limit))
	limiters.Store(userID, limiter)
	return limiter
}

func getUserIDFromContext(ctx context.Context) string {
	// Will be set by auth interceptor
	if claims, ok := ctx.Value(userClaimsKey{}).(map[string]interface{}); ok {
		if userID, ok := claims["user_id"].(string); ok {
			return userID
		}
	}
	return ""
}
