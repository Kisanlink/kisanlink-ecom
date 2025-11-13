package aaa

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// ConnectionPool manages a pool of gRPC connections with round-robin load balancing
type ConnectionPool struct {
	config      *Config
	logger      *logrus.Logger
	connections []*grpc.ClientConn
	mu          sync.RWMutex
	closed      atomic.Bool
	currentIdx  atomic.Uint32
	healthCheck *time.Ticker
	stopHealth  chan struct{}
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(ctx context.Context, config *Config, logger *logrus.Logger) (*ConnectionPool, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	pool := &ConnectionPool{
		config:      config,
		logger:      logger,
		connections: make([]*grpc.ClientConn, 0, config.PoolSize),
		stopHealth:  make(chan struct{}),
	}

	// Create connections
	for i := 0; i < config.PoolSize; i++ {
		conn, err := pool.createConnection(ctx, i)
		if err != nil {
			// Clean up already created connections
			_ = pool.closeAllConnections() // Best effort cleanup
			return nil, fmt.Errorf("failed to create connection %d: %w", i, err)
		}
		pool.connections = append(pool.connections, conn)
	}

	// Start health check goroutine
	pool.startHealthCheck()

	logger.WithFields(logrus.Fields{
		"address":   config.Address,
		"pool_size": config.PoolSize,
	}).Info("AAA connection pool created successfully")

	return pool, nil
}

// createConnection creates a single gRPC connection
func (p *ConnectionPool) createConnection(ctx context.Context, idx int) (*grpc.ClientConn, error) {
	// Setup dial options
	opts := []grpc.DialOption{
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                p.config.KeepAliveTime,
			Timeout:             p.config.KeepAliveTimeout,
			PermitWithoutStream: true,
		}),
		grpc.WithDefaultCallOptions(
			grpc.WaitForReady(true),
		),
	}

	// Configure TLS if enabled
	if p.config.EnableTLS {
		creds, err := p.setupTLSCredentials()
		if err != nil {
			return nil, fmt.Errorf("failed to setup TLS: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Create connection context with timeout
	dialCtx, cancel := context.WithTimeout(ctx, p.config.DialTimeout)
	defer cancel()

	// Dial
	//nolint:staticcheck // DialContext is deprecated but still widely used
	conn, err := grpc.DialContext(dialCtx, p.config.Address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", p.config.Address, err)
	}

	p.logger.WithFields(logrus.Fields{
		"address": p.config.Address,
		"index":   idx,
	}).Debug("Created AAA gRPC connection")

	return conn, nil
}

// setupTLSCredentials configures TLS credentials
func (p *ConnectionPool) setupTLSCredentials() (credentials.TransportCredentials, error) {
	tlsConfig := &tls.Config{
		ServerName: p.config.ServerName,
		//nolint:gosec // InsecureSkipVerify is configurable for dev environments
		InsecureSkipVerify: p.config.SkipVerify,
	}

	// Load client certificate if provided
	if p.config.CertFile != "" && p.config.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(p.config.CertFile, p.config.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Load CA certificate if provided
	if p.config.CAFile != "" {
		caCert, err := os.ReadFile(p.config.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	return credentials.NewTLS(tlsConfig), nil
}

// GetConnection returns a connection from the pool using round-robin
func (p *ConnectionPool) GetConnection() (*grpc.ClientConn, error) {
	if p.closed.Load() {
		return nil, ErrPoolShutdown
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.connections) == 0 {
		return nil, ErrNoConnections
	}

	// Round-robin selection
	connLen := len(p.connections)
	//nolint:gosec // Converting positive length to uint32 is safe
	idx := p.currentIdx.Add(1) % uint32(connLen)
	conn := p.connections[idx]

	// Check connection state
	state := conn.GetState()
	if state == connectivity.TransientFailure || state == connectivity.Shutdown {
		p.logger.WithFields(logrus.Fields{
			"index": idx,
			"state": state.String(),
		}).Warn("Connection in bad state")
		return nil, ErrConnectionClosed
	}

	return conn, nil
}

// startHealthCheck starts a goroutine that periodically checks connection health
func (p *ConnectionPool) startHealthCheck() {
	p.healthCheck = time.NewTicker(30 * time.Second)

	go func() {
		for {
			select {
			case <-p.healthCheck.C:
				p.checkAndRecreateConnections()
			case <-p.stopHealth:
				p.healthCheck.Stop()
				return
			}
		}
	}()
}

// checkAndRecreateConnections checks all connections and recreates unhealthy ones
func (p *ConnectionPool) checkAndRecreateConnections() {
	if p.closed.Load() {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	for i, conn := range p.connections {
		state := conn.GetState()

		// Recreate connections in failed states
		if state == connectivity.TransientFailure || state == connectivity.Shutdown {
			p.logger.WithFields(logrus.Fields{
				"index": i,
				"state": state.String(),
			}).Warn("Recreating unhealthy connection")

			// Close old connection
			if err := conn.Close(); err != nil {
				p.logger.WithError(err).Warn("Failed to close unhealthy connection")
			}

			// Create new connection
			ctx, cancel := context.WithTimeout(context.Background(), p.config.DialTimeout)
			newConn, err := p.createConnection(ctx, i)
			cancel()

			if err != nil {
				p.logger.WithError(err).WithField("index", i).Error("Failed to recreate connection")
				continue
			}

			p.connections[i] = newConn
			p.logger.WithField("index", i).Info("Successfully recreated connection")
		}
	}
}

// Close closes all connections in the pool
func (p *ConnectionPool) Close() error {
	if !p.closed.CompareAndSwap(false, true) {
		return nil // Already closed
	}

	// Stop health check
	close(p.stopHealth)

	// Close all connections
	return p.closeAllConnections()
}

// closeAllConnections closes all connections (must be called with write lock or during initialization)
func (p *ConnectionPool) closeAllConnections() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var errs []error
	for i, conn := range p.connections {
		if err := conn.Close(); err != nil {
			p.logger.WithError(err).WithField("index", i).Warn("Failed to close connection")
			errs = append(errs, err)
		}
	}

	p.connections = nil

	if len(errs) > 0 {
		return fmt.Errorf("failed to close %d connections", len(errs))
	}

	p.logger.Info("AAA connection pool closed")
	return nil
}

// Stats returns statistics about the connection pool
func (p *ConnectionPool) Stats() *PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := &PoolStats{
		TotalConnections: len(p.connections),
		ConnectionStates: make(map[string]int),
	}

	for _, conn := range p.connections {
		state := conn.GetState().String()
		stats.ConnectionStates[state]++
	}

	return stats
}

// PoolStats contains statistics about the connection pool
type PoolStats struct {
	TotalConnections int
	ConnectionStates map[string]int
}

// String returns a string representation of pool stats
func (s *PoolStats) String() string {
	return fmt.Sprintf("Connections: %d, States: %v", s.TotalConnections, s.ConnectionStates)
}
