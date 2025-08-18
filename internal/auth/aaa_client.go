package auth

import (
	"context"
	"fmt"
	"time"

	"kisanlink-ecom/internal/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AAAClient interface for AAA service operations
type AAAClient interface {
	ValidateToken(ctx context.Context, token string) (userID string, roles []string, err error)
	CheckPermission(ctx context.Context, subjectID, resourceType, resourceID, action string) (allowed bool, reason string, err error)
	Close() error
}

// aaaClient implements AAAClient
type aaaClient struct {
	conn   *grpc.ClientConn
	config *config.AAAConfig
}

// NewAAAClient creates a new AAA client
func NewAAAClient(cfg *config.AAAConfig) (AAAClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutMs)*time.Millisecond)
	defer cancel()

	conn, err := grpc.DialContext(ctx, cfg.Endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AAA service: %w", err)
	}

	return &aaaClient{
		conn:   conn,
		config: cfg,
	}, nil
}

// ValidateToken validates a JWT token and returns user ID and roles
func (c *aaaClient) ValidateToken(ctx context.Context, token string) (userID string, roles []string, err error) {
	// TODO: Implement actual gRPC call to aaa-service
	// For now, return mock data to allow development to continue
	// This should be replaced with actual proto-generated client calls

	// Mock implementation - remove in production
	if token == "" {
		return "", nil, fmt.Errorf("empty token")
	}

	// Simulate network delay
	time.Sleep(10 * time.Millisecond)

	// Mock successful validation
	return "USR_mock_user", []string{"buyer", "seller"}, nil
}

// CheckPermission checks if a subject has permission to perform an action on a resource
func (c *aaaClient) CheckPermission(ctx context.Context, subjectID, resourceType, resourceID, action string) (allowed bool, reason string, err error) {
	// TODO: Implement actual gRPC call to aaa-service
	// For now, return mock data to allow development to continue

	// Mock implementation - remove in production
	if subjectID == "" || resourceType == "" || action == "" {
		return false, "missing required parameters", fmt.Errorf("missing required parameters")
	}

	// Simulate network delay
	time.Sleep(10 * time.Millisecond)

	// Mock permission check - allow all for now
	// This should be replaced with actual permission logic
	return true, "permission granted", nil
}

// Close closes the gRPC connection
func (c *aaaClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// withRetry executes a function with exponential backoff retry logic
func (c *aaaClient) withRetry(operation func() error) error {
	var lastErr error
	backoff := time.Duration(c.config.TimeoutMs) * time.Millisecond

	for attempt := 0; attempt <= c.config.Retries; attempt++ {
		if err := operation(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		if attempt < c.config.Retries {
			time.Sleep(backoff)
			backoff *= 2 // exponential backoff
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", c.config.Retries+1, lastErr)
}
