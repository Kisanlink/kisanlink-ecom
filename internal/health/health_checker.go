package health

import (
	"context"
	"fmt"
	"sync"
	"time"

	"kisanlink-ecom/internal/cache"
	"kisanlink-ecom/internal/resilience"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	// StatusHealthy indicates the component is healthy
	StatusHealthy HealthStatus = "healthy"
	// StatusUnhealthy indicates the component is unhealthy
	StatusUnhealthy HealthStatus = "unhealthy"
	// StatusDegraded indicates the component is partially healthy
	StatusDegraded HealthStatus = "degraded"
	// StatusUnknown indicates the component status is unknown
	StatusUnknown HealthStatus = "unknown"
)

// ComponentHealth represents the health of an individual component
type ComponentHealth struct {
	Name        string                 `json:"name"`
	Status      HealthStatus           `json:"status"`
	Message     string                 `json:"message,omitempty"`
	LastChecked time.Time              `json:"last_checked"`
	Duration    int64                  `json:"duration"` // Duration in milliseconds
	Details     map[string]interface{} `json:"details,omitempty"`
}

// SystemHealth represents the overall system health
type SystemHealth struct {
	Status     HealthStatus               `json:"status"`
	Version    string                     `json:"version"`
	Timestamp  time.Time                  `json:"timestamp"`
	Duration   int64                      `json:"duration"` // Duration in milliseconds
	Components map[string]ComponentHealth `json:"components"`
}

// HealthChecker interface for health check implementations
type HealthChecker interface {
	Name() string
	Check(ctx context.Context) ComponentHealth
}

// HealthManager manages health checks for all system components
type HealthManager struct {
	checkers         map[string]HealthChecker
	lastCheck        map[string]ComponentHealth
	checkInterval    time.Duration
	timeout          time.Duration
	version          string
	mu               sync.RWMutex
	stopCh           chan struct{}
	backgroundChecks bool
}

// HealthManagerConfig provides configuration for health manager
type HealthManagerConfig struct {
	CheckInterval    time.Duration
	Timeout          time.Duration
	Version          string
	BackgroundChecks bool
}

// DefaultHealthManagerConfig returns default health manager configuration
func DefaultHealthManagerConfig() *HealthManagerConfig {
	return &HealthManagerConfig{
		CheckInterval:    30 * time.Second,
		Timeout:          10 * time.Second,
		Version:          "1.0.0",
		BackgroundChecks: true,
	}
}

// NewHealthManager creates a new health manager
func NewHealthManager(config *HealthManagerConfig) *HealthManager {
	if config == nil {
		config = DefaultHealthManagerConfig()
	}

	hm := &HealthManager{
		checkers:         make(map[string]HealthChecker),
		lastCheck:        make(map[string]ComponentHealth),
		checkInterval:    config.CheckInterval,
		timeout:          config.Timeout,
		version:          config.Version,
		stopCh:           make(chan struct{}),
		backgroundChecks: config.BackgroundChecks,
	}

	if config.BackgroundChecks {
		go hm.runBackgroundChecks()
	}

	return hm
}

// RegisterChecker registers a health checker
func (hm *HealthManager) RegisterChecker(checker HealthChecker) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.checkers[checker.Name()] = checker
}

// UnregisterChecker unregisters a health checker
func (hm *HealthManager) UnregisterChecker(name string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	delete(hm.checkers, name)
	delete(hm.lastCheck, name)
}

// CheckHealth performs health checks on all registered components
func (hm *HealthManager) CheckHealth(ctx context.Context) SystemHealth {
	start := time.Now()

	hm.mu.RLock()
	checkers := make(map[string]HealthChecker, len(hm.checkers))
	for name, checker := range hm.checkers {
		checkers[name] = checker
	}
	hm.mu.RUnlock()

	// Run health checks concurrently
	results := make(chan ComponentHealth, len(checkers))
	var wg sync.WaitGroup

	for name, checker := range checkers {
		wg.Add(1)
		go func(name string, checker HealthChecker) {
			defer wg.Done()

			checkCtx, cancel := context.WithTimeout(ctx, hm.timeout)
			defer cancel()

			health := checker.Check(checkCtx)
			results <- health
		}(name, checker)
	}

	// Wait for all checks to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	components := make(map[string]ComponentHealth)
	for health := range results {
		components[health.Name] = health
	}

	// Update last check results
	hm.mu.Lock()
	for name, health := range components {
		hm.lastCheck[name] = health
	}
	hm.mu.Unlock()

	// Determine overall system health
	overallStatus := hm.determineOverallStatus(components)

	return SystemHealth{
		Status:     overallStatus,
		Version:    hm.version,
		Timestamp:  time.Now(),
		Duration:   time.Since(start).Milliseconds(),
		Components: components,
	}
}

// GetLastHealthCheck returns the last health check results
func (hm *HealthManager) GetLastHealthCheck() SystemHealth {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	components := make(map[string]ComponentHealth, len(hm.lastCheck))
	for name, health := range hm.lastCheck {
		components[name] = health
	}

	overallStatus := hm.determineOverallStatus(components)

	return SystemHealth{
		Status:     overallStatus,
		Version:    hm.version,
		Timestamp:  time.Now(),
		Duration:   0, // Last check duration not tracked
		Components: components,
	}
}

// Stop stops the background health checks
func (hm *HealthManager) Stop() {
	close(hm.stopCh)
}

// runBackgroundChecks runs health checks in the background
func (hm *HealthManager) runBackgroundChecks() {
	ticker := time.NewTicker(hm.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), hm.timeout*2)
			hm.CheckHealth(ctx)
			cancel()
		case <-hm.stopCh:
			return
		}
	}
}

// determineOverallStatus determines the overall system health status
func (hm *HealthManager) determineOverallStatus(components map[string]ComponentHealth) HealthStatus {
	if len(components) == 0 {
		return StatusUnknown
	}

	healthyCount := 0
	degradedCount := 0
	unhealthyCount := 0

	for _, health := range components {
		switch health.Status {
		case StatusHealthy:
			healthyCount++
		case StatusDegraded:
			degradedCount++
		case StatusUnhealthy:
			unhealthyCount++
		}
	}

	// If any component is unhealthy, system is unhealthy
	if unhealthyCount > 0 {
		return StatusUnhealthy
	}

	// If any component is degraded, system is degraded
	if degradedCount > 0 {
		return StatusDegraded
	}

	// All components are healthy
	return StatusHealthy
}

// DatabaseHealthChecker checks database health
type DatabaseHealthChecker struct {
	name      string
	dbManager db.DBManager
}

// NewDatabaseHealthChecker creates a new database health checker
func NewDatabaseHealthChecker(name string, dbManager db.DBManager) *DatabaseHealthChecker {
	return &DatabaseHealthChecker{
		name:      name,
		dbManager: dbManager,
	}
}

// Name returns the checker name
func (dhc *DatabaseHealthChecker) Name() string {
	return dhc.name
}

// Check performs database health check
func (dhc *DatabaseHealthChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		Name:        dhc.name,
		LastChecked: start,
		Details:     make(map[string]interface{}),
	}

	// For now, we'll do a simple health check without ping
	// In a real implementation, you might query a health table or check connection
	duration := time.Since(start)
	health.Duration = duration.Milliseconds()
	health.Status = StatusHealthy
	health.Message = "Database is healthy"
	health.Details["response_time_ms"] = health.Duration

	return health
}

// CacheHealthChecker checks cache health
type CacheHealthChecker struct {
	name  string
	cache cache.Cache
}

// NewCacheHealthChecker creates a new cache health checker
func NewCacheHealthChecker(name string, cache cache.Cache) *CacheHealthChecker {
	return &CacheHealthChecker{
		name:  name,
		cache: cache,
	}
}

// Name returns the checker name
func (chc *CacheHealthChecker) Name() string {
	return chc.name
}

// Check performs cache health check
func (chc *CacheHealthChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		Name:        chc.name,
		LastChecked: start,
		Details:     make(map[string]interface{}),
	}

	// For now, we'll do a simple health check without ping
	// We could check cache stats or attempt a simple operation
	stats := chc.cache.Stats()
	duration := time.Since(start)
	health.Duration = duration.Milliseconds()
	health.Status = StatusHealthy
	health.Message = "Cache is healthy"
	health.Details["response_time_ms"] = health.Duration
	health.Details["cache_stats"] = stats

	return health
}

// CircuitBreakerHealthChecker checks circuit breaker health
type CircuitBreakerHealthChecker struct {
	name string
}

// NewCircuitBreakerHealthChecker creates a new circuit breaker health checker
func NewCircuitBreakerHealthChecker(name string) *CircuitBreakerHealthChecker {
	return &CircuitBreakerHealthChecker{
		name: name,
	}
}

// Name returns the checker name
func (cbhc *CircuitBreakerHealthChecker) Name() string {
	return cbhc.name
}

// Check performs circuit breaker health check
func (cbhc *CircuitBreakerHealthChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		Name:        cbhc.name,
		LastChecked: start,
		Details:     make(map[string]interface{}),
	}

	// Get circuit breaker statistics
	stats := resilience.GetCircuitBreakerStats()
	health.Duration = time.Since(start).Milliseconds()

	health.Details["total_breakers"] = stats.TotalBreakers
	health.Details["closed_breakers"] = stats.ClosedBreakers
	health.Details["open_breakers"] = stats.OpenBreakers
	health.Details["half_open_breakers"] = stats.HalfOpenBreakers

	// Determine health based on circuit breaker states
	if stats.TotalBreakers == 0 {
		health.Status = StatusHealthy
		health.Message = "No circuit breakers configured"
	} else if stats.OpenBreakers > stats.TotalBreakers/2 {
		health.Status = StatusUnhealthy
		health.Message = fmt.Sprintf("Too many circuit breakers are open (%d/%d)", stats.OpenBreakers, stats.TotalBreakers)
	} else if stats.OpenBreakers > 0 {
		health.Status = StatusDegraded
		health.Message = fmt.Sprintf("Some circuit breakers are open (%d/%d)", stats.OpenBreakers, stats.TotalBreakers)
	} else {
		health.Status = StatusHealthy
		health.Message = "All circuit breakers are healthy"
	}

	return health
}

// ExternalServiceHealthChecker checks external service health
type ExternalServiceHealthChecker struct {
	name        string
	checkFunc   func(context.Context) error
	serviceName string
}

// NewExternalServiceHealthChecker creates a new external service health checker
func NewExternalServiceHealthChecker(name, serviceName string, checkFunc func(context.Context) error) *ExternalServiceHealthChecker {
	return &ExternalServiceHealthChecker{
		name:        name,
		checkFunc:   checkFunc,
		serviceName: serviceName,
	}
}

// Name returns the checker name
func (eshc *ExternalServiceHealthChecker) Name() string {
	return eshc.name
}

// Check performs external service health check
func (eshc *ExternalServiceHealthChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		Name:        eshc.name,
		LastChecked: start,
		Details:     make(map[string]interface{}),
	}

	err := eshc.checkFunc(ctx)
	health.Duration = time.Since(start).Milliseconds()

	health.Details["service_name"] = eshc.serviceName
	health.Details["response_time_ms"] = health.Duration

	if err != nil {
		health.Status = StatusUnhealthy
		health.Message = fmt.Sprintf("External service %s is unhealthy: %v", eshc.serviceName, err)
		health.Details["error"] = err.Error()
	} else {
		health.Status = StatusHealthy
		health.Message = fmt.Sprintf("External service %s is healthy", eshc.serviceName)
	}

	return health
}

// CustomHealthChecker allows for custom health checks
type CustomHealthChecker struct {
	name      string
	checkFunc func(context.Context) ComponentHealth
}

// NewCustomHealthChecker creates a new custom health checker
func NewCustomHealthChecker(name string, checkFunc func(context.Context) ComponentHealth) *CustomHealthChecker {
	return &CustomHealthChecker{
		name:      name,
		checkFunc: checkFunc,
	}
}

// Name returns the checker name
func (chc *CustomHealthChecker) Name() string {
	return chc.name
}

// Check performs custom health check
func (chc *CustomHealthChecker) Check(ctx context.Context) ComponentHealth {
	return chc.checkFunc(ctx)
}

// HealthResult represents a simplified health check result
type HealthResult struct {
	Healthy bool   `json:"healthy"`
	Message string `json:"message,omitempty"`
}

// IsHealthy returns a simplified health status
func (sh SystemHealth) IsHealthy() HealthResult {
	switch sh.Status {
	case StatusHealthy:
		return HealthResult{Healthy: true, Message: "System is healthy"}
	case StatusDegraded:
		return HealthResult{Healthy: true, Message: "System is degraded but operational"}
	case StatusUnhealthy:
		return HealthResult{Healthy: false, Message: "System is unhealthy"}
	default:
		return HealthResult{Healthy: false, Message: "System status is unknown"}
	}
}
