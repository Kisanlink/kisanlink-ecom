package common

import (
	"context"
	"fmt"
)

// QueryOptions controls query behavior across repositories
type QueryOptions struct {
	// Soft delete control
	IncludeDeleted bool `json:"include_deleted"`

	// Future: Tenant isolation
	TenantID *string `json:"tenant_id,omitempty"`

	// Future: User-specific filtering
	UserID *string `json:"user_id,omitempty"`

	// Audit metadata
	RequestID string `json:"request_id,omitempty"`
	Reason    string `json:"reason,omitempty"` // Why accessing deleted items
}

// NewQueryOptions returns default query options with deleted items filtered
func NewQueryOptions() *QueryOptions {
	return &QueryOptions{
		IncludeDeleted: false,
	}
}

// Validate ensures options are valid
func (o *QueryOptions) Validate() error {
	if o.IncludeDeleted && o.Reason == "" {
		return fmt.Errorf("reason required when accessing deleted items")
	}
	return nil
}

// Clone creates a deep copy of query options
func (o *QueryOptions) Clone() *QueryOptions {
	if o == nil {
		return NewQueryOptions()
	}

	clone := &QueryOptions{
		IncludeDeleted: o.IncludeDeleted,
		RequestID:      o.RequestID,
		Reason:         o.Reason,
	}

	if o.TenantID != nil {
		tid := *o.TenantID
		clone.TenantID = &tid
	}

	if o.UserID != nil {
		uid := *o.UserID
		clone.UserID = &uid
	}

	return clone
}

// Context key type for storing query options
type contextKey string

const queryOptionsKey contextKey = "queryOptions"

// ContextWithQueryOptions adds query options to context
func ContextWithQueryOptions(ctx context.Context, opts *QueryOptions) context.Context {
	if opts == nil {
		opts = NewQueryOptions()
	}

	// Validate options before storing
	if err := opts.Validate(); err != nil {
		// Return context with default options if validation fails
		return context.WithValue(ctx, queryOptionsKey, NewQueryOptions())
	}

	return context.WithValue(ctx, queryOptionsKey, opts)
}

// QueryOptionsFromContext retrieves options from context
// Returns default options if not found in context
func QueryOptionsFromContext(ctx context.Context) *QueryOptions {
	if ctx == nil {
		return NewQueryOptions()
	}

	if opts, ok := ctx.Value(queryOptionsKey).(*QueryOptions); ok {
		return opts
	}

	return NewQueryOptions()
}

// WithIncludeDeleted returns a new QueryOptions with IncludeDeleted set to true
func WithIncludeDeleted(reason string) *QueryOptions {
	opts := NewQueryOptions()
	opts.IncludeDeleted = true
	opts.Reason = reason
	return opts
}

// WithTenantID returns a new QueryOptions with TenantID set
func WithTenantID(tenantID string) *QueryOptions {
	opts := NewQueryOptions()
	opts.TenantID = &tenantID
	return opts
}

// WithRequestID returns a new QueryOptions with RequestID set
func WithRequestID(requestID string) *QueryOptions {
	opts := NewQueryOptions()
	opts.RequestID = requestID
	return opts
}
