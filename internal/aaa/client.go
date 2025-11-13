package aaa

import (
	"context"
	"fmt"
	"time"

	aaaPb "github.com/Kisanlink/aaa-service/v2/pkg/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Client provides methods for interacting with AAA service
type Client struct {
	pool           *ConnectionPool
	circuitBreaker *CircuitBreaker
	retryExecutor  *RetryExecutor
	logger         *logrus.Logger
}

// NewClient creates a new AAA client with connection pool, circuit breaker, and retry logic
func NewClient(ctx context.Context, config *Config, logger *logrus.Logger) (*Client, error) {
	// Create connection pool
	pool, err := NewConnectionPool(ctx, config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Create circuit breaker
	circuitBreaker := NewCircuitBreaker("AAA-AddressService", config.CircuitBreaker, logger)

	// Create retry executor
	retryConfig := RetryConfig{
		MaxAttempts:    config.MaxRetries,
		InitialBackoff: config.InitialBackoff,
		MaxBackoff:     config.MaxBackoff,
		BackoffFactor:  config.BackoffFactor,
		JitterFactor:   0.1,
	}
	retryExecutor := NewRetryExecutor(retryConfig, logger)

	return &Client{
		pool:           pool,
		circuitBreaker: circuitBreaker,
		retryExecutor:  retryExecutor,
		logger:         logger,
	}, nil
}

// CreateAddress creates a new address in AAA service
func (c *Client) CreateAddress(ctx context.Context, req *CreateAddressRequest) (*Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var address *Address

	// Execute with circuit breaker and retry
	err := c.circuitBreaker.Execute(ctx, func(ctx context.Context) error {
		return c.retryExecutor.Execute(ctx, "CreateAddress", func(ctx context.Context) error {
			// Get connection from pool
			conn, err := c.pool.GetConnection()
			if err != nil {
				return err
			}

			// Create client
			client := aaaPb.NewAddressServiceClient(conn)

			// Build request
			aaaReq := &aaaPb.CreateAddressRequest{
				UserId:        req.EntityID,
				Type:          string(req.Type),
				AddressLine_1: req.Line1,
				AddressLine_2: req.Line2,
				City:          req.City,
				State:         req.State,
				PostalCode:    req.PostalCode,
				Country:       req.Country,
				IsPrimary:     req.IsPrimary,
			}

			// Call AAA service with timeout
			callCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			resp, err := client.CreateAddress(callCtx, aaaReq)
			if err != nil {
				return c.mapGRPCError(err)
			}

			// Convert response
			address = c.toAddress(resp.Address)
			return nil
		})
	})

	if err != nil {
		c.logger.WithError(err).WithFields(logrus.Fields{
			"entity_id":   req.EntityID,
			"entity_type": req.EntityType,
		}).Error("Failed to create address")
		return nil, err
	}

	c.logger.WithFields(logrus.Fields{
		"address_id":  address.ID,
		"entity_id":   req.EntityID,
		"entity_type": req.EntityType,
	}).Info("Address created successfully")

	return address, nil
}

// UpdateAddress updates an existing address
func (c *Client) UpdateAddress(ctx context.Context, req *UpdateAddressRequest) (*Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var address *Address

	err := c.circuitBreaker.Execute(ctx, func(ctx context.Context) error {
		return c.retryExecutor.Execute(ctx, "UpdateAddress", func(ctx context.Context) error {
			conn, err := c.pool.GetConnection()
			if err != nil {
				return err
			}

			client := aaaPb.NewAddressServiceClient(conn)

			aaaReq := &aaaPb.UpdateAddressRequest{
				Id:            req.ID,
				Type:          string(req.Type),
				AddressLine_1: req.Line1,
				AddressLine_2: req.Line2,
				City:          req.City,
				State:         req.State,
				PostalCode:    req.PostalCode,
				Country:       req.Country,
			}

			callCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			resp, err := client.UpdateAddress(callCtx, aaaReq)
			if err != nil {
				return c.mapGRPCError(err)
			}

			address = c.toAddress(resp.Address)
			return nil
		})
	})

	if err != nil {
		c.logger.WithError(err).WithField("address_id", req.ID).Error("Failed to update address")
		return nil, err
	}

	c.logger.WithField("address_id", address.ID).Info("Address updated successfully")
	return address, nil
}

// GetAddress retrieves an address by ID
func (c *Client) GetAddress(ctx context.Context, addressID string) (*Address, error) {
	if addressID == "" {
		return nil, ErrInvalidAddressID
	}

	var address *Address

	err := c.circuitBreaker.Execute(ctx, func(ctx context.Context) error {
		return c.retryExecutor.Execute(ctx, "GetAddress", func(ctx context.Context) error {
			conn, err := c.pool.GetConnection()
			if err != nil {
				return err
			}

			client := aaaPb.NewAddressServiceClient(conn)

			callCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			resp, err := client.GetAddress(callCtx, &aaaPb.GetAddressRequest{Id: addressID})
			if err != nil {
				return c.mapGRPCError(err)
			}

			address = c.toAddress(resp.Address)
			return nil
		})
	})

	if err != nil {
		c.logger.WithError(err).WithField("address_id", addressID).Debug("Failed to get address")
		return nil, err
	}

	return address, nil
}

// DeleteAddress deletes an address by ID
func (c *Client) DeleteAddress(ctx context.Context, addressID string) error {
	if addressID == "" {
		return ErrInvalidAddressID
	}

	err := c.circuitBreaker.Execute(ctx, func(ctx context.Context) error {
		return c.retryExecutor.Execute(ctx, "DeleteAddress", func(ctx context.Context) error {
			conn, err := c.pool.GetConnection()
			if err != nil {
				return err
			}

			client := aaaPb.NewAddressServiceClient(conn)

			callCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			_, err = client.DeleteAddress(callCtx, &aaaPb.DeleteAddressRequest{Id: addressID})
			if err != nil {
				return c.mapGRPCError(err)
			}

			return nil
		})
	})

	if err != nil {
		c.logger.WithError(err).WithField("address_id", addressID).Error("Failed to delete address")
		return err
	}

	c.logger.WithField("address_id", addressID).Info("Address deleted successfully")
	return nil
}

// ListAddresses retrieves all addresses for an entity
func (c *Client) ListAddresses(ctx context.Context, userID string) ([]*Address, error) {
	if userID == "" {
		return nil, ErrInvalidRequest
	}

	var addresses []*Address

	err := c.circuitBreaker.Execute(ctx, func(ctx context.Context) error {
		return c.retryExecutor.Execute(ctx, "ListAddresses", func(ctx context.Context) error {
			conn, err := c.pool.GetConnection()
			if err != nil {
				return err
			}

			client := aaaPb.NewAddressServiceClient(conn)

			callCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			resp, err := client.GetAddressesByUser(callCtx, &aaaPb.GetAddressesByUserRequest{
				UserId:     userID,
				ActiveOnly: true,
			})
			if err != nil {
				return c.mapGRPCError(err)
			}

			addresses = make([]*Address, 0, len(resp.Addresses))
			for _, addr := range resp.Addresses {
				addresses = append(addresses, c.toAddress(addr))
			}

			return nil
		})
	})

	if err != nil {
		c.logger.WithError(err).WithField("user_id", userID).Error("Failed to list addresses")
		return nil, err
	}

	return addresses, nil
}

// toAddress converts AAA proto address to domain address
func (c *Client) toAddress(aaaAddr *aaaPb.Address) *Address {
	if aaaAddr == nil {
		return nil
	}

	addr := &Address{
		ID:         aaaAddr.Id,
		EntityID:   aaaAddr.UserId,
		EntityType: "user", // AAA only stores user_id
		Line1:      aaaAddr.AddressLine_1,
		Line2:      aaaAddr.AddressLine_2,
		City:       aaaAddr.City,
		State:      aaaAddr.State,
		Country:    aaaAddr.Country,
		PostalCode: aaaAddr.PostalCode,
		Type:       AddressType(aaaAddr.Type),
		IsPrimary:  aaaAddr.IsPrimary,
		IsVerified: aaaAddr.IsActive, // Map is_active to is_verified
	}

	if aaaAddr.CreatedAt != nil {
		addr.CreatedAt = aaaAddr.CreatedAt.AsTime()
	}
	if aaaAddr.UpdatedAt != nil {
		addr.UpdatedAt = aaaAddr.UpdatedAt.AsTime()
	}

	return addr
}

// fromAddress converts domain address to AAA proto address (unused but kept for future use)
//
//nolint:unused // Will be used for update operations
func (c *Client) fromAddress(addr *Address) *aaaPb.Address {
	if addr == nil {
		return nil
	}

	aaaAddr := &aaaPb.Address{
		Id:            addr.ID,
		UserId:        addr.EntityID,
		AddressLine_1: addr.Line1,
		AddressLine_2: addr.Line2,
		City:          addr.City,
		State:         addr.State,
		Country:       addr.Country,
		PostalCode:    addr.PostalCode,
		Type:          string(addr.Type),
		IsPrimary:     addr.IsPrimary,
		IsActive:      addr.IsVerified, // Map is_verified to is_active
	}

	if !addr.CreatedAt.IsZero() {
		aaaAddr.CreatedAt = timestamppb.New(addr.CreatedAt)
	}
	if !addr.UpdatedAt.IsZero() {
		aaaAddr.UpdatedAt = timestamppb.New(addr.UpdatedAt)
	}

	return aaaAddr
}

// mapGRPCError maps gRPC errors to domain errors
func (c *Client) mapGRPCError(err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return ErrInternal
	}

	switch st.Code() {
	case codes.NotFound:
		return ErrAddressNotFound
	case codes.InvalidArgument:
		return ErrInvalidAddress
	case codes.AlreadyExists:
		return ErrAddressExists
	case codes.Unauthenticated:
		return ErrUnauthenticated
	case codes.PermissionDenied:
		return ErrUnauthorized
	case codes.ResourceExhausted:
		return ErrRateLimited
	case codes.Unavailable:
		return ErrServiceUnavailable
	case codes.DeadlineExceeded:
		return ErrTimeout
	case codes.Internal:
		return ErrInternal
	default:
		c.logger.WithFields(logrus.Fields{
			"grpc_code":    st.Code(),
			"grpc_message": st.Message(),
		}).Warn("Unmapped gRPC error")
		return ErrInternal
	}
}

// Close closes the client and releases all resources
func (c *Client) Close() error {
	return c.pool.Close()
}

// Stats returns statistics about the client
func (c *Client) Stats() *ClientStats {
	poolStats := c.pool.Stats()
	requests, successes, failures, failureRatio := c.circuitBreaker.Counts()

	return &ClientStats{
		PoolStats:           poolStats,
		CircuitBreakerState: c.circuitBreaker.State().String(),
		Requests:            requests,
		Successes:           successes,
		Failures:            failures,
		FailureRatio:        failureRatio,
	}
}

// ClientStats contains statistics about the client
type ClientStats struct {
	PoolStats           *PoolStats
	CircuitBreakerState string
	Requests            uint32
	Successes           uint32
	Failures            uint32
	FailureRatio        float64
}

// String returns a string representation of client stats
func (s *ClientStats) String() string {
	return fmt.Sprintf("AAA Client Stats: CB=%s, Requests=%d, Success=%d, Failures=%d, Ratio=%.2f, Pool=%s",
		s.CircuitBreakerState, s.Requests, s.Successes, s.Failures, s.FailureRatio, s.PoolStats.String())
}
