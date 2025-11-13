package marketplace

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// ConnectionPoolManager manages database connection pools for high-volume scenarios
type ConnectionPoolManager interface {
	// Get a connection from the pool
	GetConnection(ctx context.Context) (db.DBManager, error)

	// Return a connection to the pool
	ReturnConnection(conn db.DBManager) error

	// Get pool statistics
	GetPoolStats() PoolStats

	// Close all connections in the pool
	Close() error

	// Health check for all connections
	HealthCheck(ctx context.Context) error
}

// PoolStats provides connection pool metrics
type PoolStats struct {
	TotalConnections     int
	ActiveConnections    int
	IdleConnections      int
	WaitingRequests      int
	MaxConnections       int
	ConnectionsCreated   int64
	ConnectionsDestroyed int64
	AverageWaitTime      time.Duration
}

// PooledConnection wraps a database connection with pool metadata
type PooledConnection struct {
	DBManager  db.DBManager
	CreatedAt  time.Time
	LastUsedAt time.Time
	UsageCount int64
	IsActive   bool
}

// connectionPoolManager implements ConnectionPoolManager
type connectionPoolManager struct {
	connections      []*PooledConnection
	availableConns   chan *PooledConnection
	maxConnections   int
	minConnections   int
	connectionTTL    time.Duration
	dbManagerFactory func() (db.DBManager, error)

	// Statistics
	stats    PoolStats
	statsMux sync.RWMutex

	// Pool management
	poolMux sync.RWMutex
	closed  bool

	// Monitoring
	waitTimes []time.Duration
	waitMux   sync.Mutex
}

// ConnectionPoolConfig provides configuration for the connection pool
type ConnectionPoolConfig struct {
	MaxConnections      int
	MinConnections      int
	ConnectionTTL       time.Duration
	IdleTimeout         time.Duration
	HealthCheckInterval time.Duration
}

// DefaultConnectionPoolConfig returns sensible defaults for connection pooling
func DefaultConnectionPoolConfig() ConnectionPoolConfig {
	return ConnectionPoolConfig{
		MaxConnections:      50,
		MinConnections:      5,
		ConnectionTTL:       30 * time.Minute,
		IdleTimeout:         5 * time.Minute,
		HealthCheckInterval: 1 * time.Minute,
	}
}

// NewConnectionPoolManager creates a new connection pool manager
func NewConnectionPoolManager(config ConnectionPoolConfig, dbManagerFactory func() (db.DBManager, error)) (ConnectionPoolManager, error) {
	if config.MaxConnections <= 0 {
		return nil, fmt.Errorf("max connections must be positive")
	}

	if config.MinConnections < 0 || config.MinConnections > config.MaxConnections {
		return nil, fmt.Errorf("min connections must be between 0 and max connections")
	}

	pool := &connectionPoolManager{
		connections:      make([]*PooledConnection, 0, config.MaxConnections),
		availableConns:   make(chan *PooledConnection, config.MaxConnections),
		maxConnections:   config.MaxConnections,
		minConnections:   config.MinConnections,
		connectionTTL:    config.ConnectionTTL,
		dbManagerFactory: dbManagerFactory,
		waitTimes:        make([]time.Duration, 0, 100),
	}

	// Initialize minimum connections
	if err := pool.initializeMinConnections(); err != nil {
		return nil, fmt.Errorf("failed to initialize minimum connections: %w", err)
	}

	// Start background maintenance
	go pool.maintenanceLoop(config.IdleTimeout, config.HealthCheckInterval)

	return pool, nil
}

// GetConnection retrieves a connection from the pool
func (p *connectionPoolManager) GetConnection(ctx context.Context) (db.DBManager, error) {
	if p.closed {
		return nil, fmt.Errorf("connection pool is closed")
	}

	startTime := time.Now()

	select {
	case conn := <-p.availableConns:
		// Got an available connection
		waitTime := time.Since(startTime)
		p.recordWaitTime(waitTime)

		conn.LastUsedAt = time.Now()
		conn.UsageCount++
		conn.IsActive = true

		p.updateStats(func(stats *PoolStats) {
			stats.ActiveConnections++
			stats.IdleConnections--
		})

		return conn.DBManager, nil

	case <-ctx.Done():
		return nil, ctx.Err()

	default:
		// No available connections, try to create a new one
		if p.canCreateNewConnection() {
			conn, err := p.createConnection()
			if err != nil {
				return nil, fmt.Errorf("failed to create new connection: %w", err)
			}

			waitTime := time.Since(startTime)
			p.recordWaitTime(waitTime)

			return conn.DBManager, nil
		}

		// Wait for an available connection
		select {
		case conn := <-p.availableConns:
			waitTime := time.Since(startTime)
			p.recordWaitTime(waitTime)

			conn.LastUsedAt = time.Now()
			conn.UsageCount++
			conn.IsActive = true

			p.updateStats(func(stats *PoolStats) {
				stats.ActiveConnections++
				stats.IdleConnections--
			})

			return conn.DBManager, nil

		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// ReturnConnection returns a connection to the pool
func (p *connectionPoolManager) ReturnConnection(dbManager db.DBManager) error {
	if p.closed {
		return fmt.Errorf("connection pool is closed")
	}

	p.poolMux.RLock()
	defer p.poolMux.RUnlock()

	// Find the connection in our pool
	for _, conn := range p.connections {
		if conn.DBManager == dbManager {
			conn.IsActive = false
			conn.LastUsedAt = time.Now()

			// Check if connection is still healthy and not expired
			if p.isConnectionHealthy(conn) && !p.isConnectionExpired(conn) {
				select {
				case p.availableConns <- conn:
					p.updateStats(func(stats *PoolStats) {
						stats.ActiveConnections--
						stats.IdleConnections++
					})
					return nil
				default:
					// Channel is full, close this connection
					return p.closeConnection(conn)
				}
			} else {
				// Connection is unhealthy or expired, close it
				return p.closeConnection(conn)
			}
		}
	}

	return fmt.Errorf("connection not found in pool")
}

// GetPoolStats returns current pool statistics
func (p *connectionPoolManager) GetPoolStats() PoolStats {
	p.statsMux.RLock()
	defer p.statsMux.RUnlock()

	stats := p.stats
	stats.TotalConnections = len(p.connections)
	stats.MaxConnections = p.maxConnections

	// Calculate average wait time
	p.waitMux.Lock()
	if len(p.waitTimes) > 0 {
		var total time.Duration
		for _, waitTime := range p.waitTimes {
			total += waitTime
		}
		stats.AverageWaitTime = total / time.Duration(len(p.waitTimes))
	}
	p.waitMux.Unlock()

	return stats
}

// Close closes all connections in the pool
func (p *connectionPoolManager) Close() error {
	p.poolMux.Lock()
	defer p.poolMux.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	// Close all connections
	for _, conn := range p.connections {
		if conn.DBManager != nil {
			// In a real implementation, we would close the database connection
			// conn.DBManager.Close()
		}
	}

	// Clear the pool
	p.connections = nil
	close(p.availableConns)

	return nil
}

// HealthCheck performs health checks on all connections
func (p *connectionPoolManager) HealthCheck(ctx context.Context) error {
	p.poolMux.RLock()
	connections := make([]*PooledConnection, len(p.connections))
	copy(connections, p.connections)
	p.poolMux.RUnlock()

	var unhealthyCount int

	for _, conn := range connections {
		if !p.isConnectionHealthy(conn) {
			unhealthyCount++
			// Remove unhealthy connection
			p.closeConnection(conn)
		}
	}

	if unhealthyCount > 0 {
		fmt.Printf("Health check found %d unhealthy connections\n", unhealthyCount)
	}

	return nil
}

// Helper methods

// initializeMinConnections creates the minimum number of connections
func (p *connectionPoolManager) initializeMinConnections() error {
	for i := 0; i < p.minConnections; i++ {
		conn, err := p.createConnection()
		if err != nil {
			return fmt.Errorf("failed to create initial connection %d: %w", i, err)
		}

		// Return connection to pool immediately
		p.availableConns <- conn

		p.updateStats(func(stats *PoolStats) {
			stats.IdleConnections++
		})
	}

	return nil
}

// createConnection creates a new pooled connection
func (p *connectionPoolManager) createConnection() (*PooledConnection, error) {
	dbManager, err := p.dbManagerFactory()
	if err != nil {
		return nil, err
	}

	conn := &PooledConnection{
		DBManager:  dbManager,
		CreatedAt:  time.Now(),
		LastUsedAt: time.Now(),
		UsageCount: 0,
		IsActive:   true,
	}

	p.poolMux.Lock()
	p.connections = append(p.connections, conn)
	p.poolMux.Unlock()

	p.updateStats(func(stats *PoolStats) {
		stats.ConnectionsCreated++
		stats.ActiveConnections++
	})

	return conn, nil
}

// closeConnection closes and removes a connection from the pool
func (p *connectionPoolManager) closeConnection(conn *PooledConnection) error {
	p.poolMux.Lock()
	defer p.poolMux.Unlock()

	// Remove from connections slice
	for i, c := range p.connections {
		if c == conn {
			p.connections = append(p.connections[:i], p.connections[i+1:]...)
			break
		}
	}

	// Close the database connection
	// In a real implementation: conn.DBManager.Close()

	p.updateStats(func(stats *PoolStats) {
		stats.ConnectionsDestroyed++
		if conn.IsActive {
			stats.ActiveConnections--
		} else {
			stats.IdleConnections--
		}
	})

	return nil
}

// canCreateNewConnection checks if we can create a new connection
func (p *connectionPoolManager) canCreateNewConnection() bool {
	p.poolMux.RLock()
	defer p.poolMux.RUnlock()

	return len(p.connections) < p.maxConnections
}

// isConnectionHealthy checks if a connection is healthy
func (p *connectionPoolManager) isConnectionHealthy(conn *PooledConnection) bool {
	// In a real implementation, this would perform a health check on the database connection
	// For now, we'll assume all connections are healthy
	return true
}

// isConnectionExpired checks if a connection has exceeded its TTL
func (p *connectionPoolManager) isConnectionExpired(conn *PooledConnection) bool {
	return time.Since(conn.CreatedAt) > p.connectionTTL
}

// updateStats safely updates pool statistics
func (p *connectionPoolManager) updateStats(updateFn func(*PoolStats)) {
	p.statsMux.Lock()
	defer p.statsMux.Unlock()
	updateFn(&p.stats)
}

// recordWaitTime records a connection wait time for statistics
func (p *connectionPoolManager) recordWaitTime(waitTime time.Duration) {
	p.waitMux.Lock()
	defer p.waitMux.Unlock()

	p.waitTimes = append(p.waitTimes, waitTime)

	// Keep only the last 100 wait times
	if len(p.waitTimes) > 100 {
		p.waitTimes = p.waitTimes[1:]
	}
}

// maintenanceLoop performs periodic maintenance on the connection pool
func (p *connectionPoolManager) maintenanceLoop(idleTimeout, healthCheckInterval time.Duration) {
	healthTicker := time.NewTicker(healthCheckInterval)
	cleanupTicker := time.NewTicker(idleTimeout)

	defer healthTicker.Stop()
	defer cleanupTicker.Stop()

	for {
		select {
		case <-healthTicker.C:
			if !p.closed {
				_ = p.HealthCheck(context.Background())
			}

		case <-cleanupTicker.C:
			if !p.closed {
				p.cleanupIdleConnections(idleTimeout)
			}
		}

		if p.closed {
			return
		}
	}
}

// cleanupIdleConnections removes connections that have been idle too long
func (p *connectionPoolManager) cleanupIdleConnections(idleTimeout time.Duration) {
	p.poolMux.RLock()
	connections := make([]*PooledConnection, len(p.connections))
	copy(connections, p.connections)
	p.poolMux.RUnlock()

	var closedCount int

	for _, conn := range connections {
		if !conn.IsActive && time.Since(conn.LastUsedAt) > idleTimeout {
			// Don't close if we're at minimum connections
			if len(p.connections) > p.minConnections {
				p.closeConnection(conn)
				closedCount++
			}
		}
	}

	if closedCount > 0 {
		fmt.Printf("Cleaned up %d idle connections\n", closedCount)
	}
}
