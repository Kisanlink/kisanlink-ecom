// Package collaborator implements the gRPC handlers for collaborator management
package collaborator

import (
	"context"
)

type contextKey string

const userContextKey contextKey = "user_context"

// GetUserContext extracts user context from the request context
// This is populated by the authentication interceptor
func GetUserContext(ctx context.Context) *UserContext {
	if uc, ok := ctx.Value(userContextKey).(*UserContext); ok {
		return uc
	}
	// Return empty context if not found (shouldn't happen after auth interceptor)
	return &UserContext{}
}

// WithUserContext adds user context to the request context
func WithUserContext(ctx context.Context, uc *UserContext) context.Context {
	return context.WithValue(ctx, userContextKey, uc)
}
