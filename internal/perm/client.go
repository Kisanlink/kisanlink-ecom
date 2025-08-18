package perm

import (
	"context"
)

// Client defines the interface for permission checking
type Client interface {
	// Check validates if a subject can perform an action on a resource
	Check(ctx context.Context, subjectID, resourceType, resourceID, action string) (bool, error)

	// ValidateToken validates a JWT token and returns the user ID
	ValidateToken(ctx context.Context, token string) (string, error)
}

// Config holds configuration for the permission client
type Config struct {
	AAAServiceEndpoint string
	Timeout            int // seconds
	MaxRetries         int
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		AAAServiceEndpoint: "localhost:50051",
		Timeout:            30,
		MaxRetries:         3,
	}
}
