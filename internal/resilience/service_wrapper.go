package resilience

import (
	"context"
	"fmt"
	"time"

	"kisanlink-ecom/internal/auth"
)

// ServiceConfig holds configuration for resilient service wrappers
type ServiceConfig struct {
	// ServiceName identifies the service
	ServiceName string
	// CircuitBreakerConfig for the service
	CircuitBreakerConfig *CircuitBreakerConfig
	// RetryConfig for the service
	RetryConfig *RetryConfig
	// TimeoutConfig for the service
	TimeoutConfig *TimeoutConfig
}

// DefaultServiceConfig returns default service configuration
func DefaultServiceConfig(serviceName string) *ServiceConfig {
	return &ServiceConfig{
		ServiceName: serviceName,
		CircuitBreakerConfig: &CircuitBreakerConfig{
			Name:        serviceName + "_cb",
			MaxRequests: 3,
			Interval:    60 * time.Second,
			Timeout:     30 * time.Second,
			ReadyToTrip: func(counts Counts) bool {
				failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
				return counts.Requests >= 5 && failureRatio >= 0.6
			},
		},
		RetryConfig: &RetryConfig{
			MaxRetries:      3,
			InitialInterval: 100 * time.Millisecond,
			MaxInterval:     2 * time.Second,
			Multiplier:      2.0,
		},
		TimeoutConfig: &TimeoutConfig{
			Timeout: 10 * time.Second,
		},
	}
}

// AAAClientWrapper wraps AAA client with resilience patterns
type AAAClientWrapper struct {
	client         auth.Client
	circuitBreaker *CircuitBreaker
	retryPolicy    *RetryPolicy
	config         *ServiceConfig
}

// NewAAAClientWrapper creates a new resilient AAA client wrapper
func NewAAAClientWrapper(client auth.Client, config *ServiceConfig) *AAAClientWrapper {
	if config == nil {
		config = DefaultServiceConfig("aaa_service")
	}

	return &AAAClientWrapper{
		client:         client,
		circuitBreaker: GetCircuitBreakerWithConfig(config.ServiceName+"_cb", config.CircuitBreakerConfig),
		retryPolicy:    NewRetryPolicy(config.RetryConfig),
		config:         config,
	}
}

// ValidateToken validates a JWT token with circuit breaker and retry
func (w *AAAClientWrapper) ValidateToken(ctx context.Context, token string) (*auth.TokenClaims, error) {
	var result *auth.TokenClaims
	err := w.circuitBreaker.CallContext(ctx, func(ctx context.Context) error {
		return w.retryPolicy.ExecuteContext(ctx, func(ctx context.Context) error {
			// Apply timeout
			timeoutCtx, cancel := context.WithTimeout(ctx, w.config.TimeoutConfig.Timeout)
			defer cancel()

			var err error
			result, err = w.client.ValidateToken(timeoutCtx, token)
			return err
		})
	})
	return result, err
}

// Authorize performs authorization with resilience
func (w *AAAClientWrapper) Authorize(ctx context.Context, req *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error) {
	var result *auth.AuthorizeResponse
	err := w.circuitBreaker.CallContext(ctx, func(ctx context.Context) error {
		return w.retryPolicy.ExecuteContext(ctx, func(ctx context.Context) error {
			timeoutCtx, cancel := context.WithTimeout(ctx, w.config.TimeoutConfig.Timeout)
			defer cancel()

			var err error
			result, err = w.client.Authorize(timeoutCtx, req)
			return err
		})
	})
	return result, err
}

// HealthCheck performs health check with resilience
func (w *AAAClientWrapper) HealthCheck(ctx context.Context) error {
	return w.circuitBreaker.CallContext(ctx, func(ctx context.Context) error {
		timeoutCtx, cancel := context.WithTimeout(ctx, w.config.TimeoutConfig.Timeout)
		defer cancel()

		return w.client.HealthCheck(timeoutCtx)
	})
}

// DatabaseWrapper wraps database operations with resilience patterns
type DatabaseWrapper struct {
	circuitBreaker *CircuitBreaker
	retryPolicy    *RetryPolicy
	config         *ServiceConfig
}

// NewDatabaseWrapper creates a new resilient database wrapper
func NewDatabaseWrapper(config *ServiceConfig) *DatabaseWrapper {
	if config == nil {
		config = DefaultServiceConfig("database")
	}

	return &DatabaseWrapper{
		circuitBreaker: GetCircuitBreakerWithConfig(config.ServiceName+"_cb", config.CircuitBreakerConfig),
		retryPolicy:    NewRetryPolicy(config.RetryConfig),
		config:         config,
	}
}

// Execute executes a database operation with resilience
func (w *DatabaseWrapper) Execute(ctx context.Context, operation func(ctx context.Context) error) error {
	return w.circuitBreaker.CallContext(ctx, func(ctx context.Context) error {
		return w.retryPolicy.ExecuteContext(ctx, func(ctx context.Context) error {
			timeoutCtx, cancel := context.WithTimeout(ctx, w.config.TimeoutConfig.Timeout)
			defer cancel()

			return operation(timeoutCtx)
		})
	})
}

// ExecuteWithResult executes a database operation that returns a result
func (w *DatabaseWrapper) ExecuteWithResult(ctx context.Context, operation func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	var result interface{}
	err := w.circuitBreaker.CallContext(ctx, func(ctx context.Context) error {
		return w.retryPolicy.ExecuteContext(ctx, func(ctx context.Context) error {
			timeoutCtx, cancel := context.WithTimeout(ctx, w.config.TimeoutConfig.Timeout)
			defer cancel()

			var err error
			result, err = operation(timeoutCtx)
			return err
		})
	})
	return result, err
}

// ExternalServiceWrapper provides a generic wrapper for external services
type ExternalServiceWrapper struct {
	serviceName    string
	circuitBreaker *CircuitBreaker
	retryPolicy    *RetryPolicy
	config         *ServiceConfig
}

// NewExternalServiceWrapper creates a new wrapper for external services
func NewExternalServiceWrapper(serviceName string, config *ServiceConfig) *ExternalServiceWrapper {
	if config == nil {
		config = DefaultServiceConfig(serviceName)
	}

	return &ExternalServiceWrapper{
		serviceName:    serviceName,
		circuitBreaker: GetCircuitBreakerWithConfig(serviceName+"_cb", config.CircuitBreakerConfig),
		retryPolicy:    NewRetryPolicy(config.RetryConfig),
		config:         config,
	}
}

// Call executes an external service call with resilience
func (w *ExternalServiceWrapper) Call(ctx context.Context, operation func(ctx context.Context) error) error {
	return w.circuitBreaker.CallContext(ctx, func(ctx context.Context) error {
		return w.retryPolicy.ExecuteContext(ctx, func(ctx context.Context) error {
			timeoutCtx, cancel := context.WithTimeout(ctx, w.config.TimeoutConfig.Timeout)
			defer cancel()

			return operation(timeoutCtx)
		})
	})
}

// CallWithResult executes an external service call that returns a result
func (w *ExternalServiceWrapper) CallWithResult(ctx context.Context, operation func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	var result interface{}
	err := w.circuitBreaker.CallContext(ctx, func(ctx context.Context) error {
		return w.retryPolicy.ExecuteContext(ctx, func(ctx context.Context) error {
			timeoutCtx, cancel := context.WithTimeout(ctx, w.config.TimeoutConfig.Timeout)
			defer cancel()

			var err error
			result, err = operation(timeoutCtx)
			return err
		})
	})
	return result, err
}

// ServiceWrapperFactory creates resilient service wrappers
type ServiceWrapperFactory struct {
	configs map[string]*ServiceConfig
}

// NewServiceWrapperFactory creates a new service wrapper factory
func NewServiceWrapperFactory() *ServiceWrapperFactory {
	return &ServiceWrapperFactory{
		configs: make(map[string]*ServiceConfig),
	}
}

// SetConfig sets configuration for a service
func (f *ServiceWrapperFactory) SetConfig(serviceName string, config *ServiceConfig) {
	f.configs[serviceName] = config
}

// GetConfig gets configuration for a service
func (f *ServiceWrapperFactory) GetConfig(serviceName string) *ServiceConfig {
	if config, exists := f.configs[serviceName]; exists {
		return config
	}
	return DefaultServiceConfig(serviceName)
}

// CreateAAAWrapper creates a resilient AAA client wrapper
func (f *ServiceWrapperFactory) CreateAAAWrapper(client auth.Client) *AAAClientWrapper {
	config := f.GetConfig("aaa_service")
	return NewAAAClientWrapper(client, config)
}

// CreateDatabaseWrapper creates a resilient database wrapper
func (f *ServiceWrapperFactory) CreateDatabaseWrapper() *DatabaseWrapper {
	config := f.GetConfig("database")
	return NewDatabaseWrapper(config)
}

// CreateExternalServiceWrapper creates a resilient external service wrapper
func (f *ServiceWrapperFactory) CreateExternalServiceWrapper(serviceName string) *ExternalServiceWrapper {
	config := f.GetConfig(serviceName)
	return NewExternalServiceWrapper(serviceName, config)
}

// Global service wrapper factory
var defaultFactory = NewServiceWrapperFactory()

// SetServiceConfig sets global configuration for a service
func SetServiceConfig(serviceName string, config *ServiceConfig) {
	defaultFactory.SetConfig(serviceName, config)
}

// CreateResilientAAAClient creates a resilient AAA client wrapper
func CreateResilientAAAClient(client auth.Client) *AAAClientWrapper {
	return defaultFactory.CreateAAAWrapper(client)
}

// CreateResilientDatabaseWrapper creates a resilient database wrapper
func CreateResilientDatabaseWrapper() *DatabaseWrapper {
	return defaultFactory.CreateDatabaseWrapper()
}

// CreateResilientExternalService creates a resilient external service wrapper
func CreateResilientExternalService(serviceName string) *ExternalServiceWrapper {
	return defaultFactory.CreateExternalServiceWrapper(serviceName)
}

// ResilienceMiddleware provides circuit breaker statistics via HTTP endpoints
type ResilienceMiddleware struct {
	// This will be implemented in the middleware package
}

// CircuitBreakerMetrics provides metrics for monitoring
type CircuitBreakerMetrics struct {
	Name                 string    `json:"name"`
	State                string    `json:"state"`
	TotalRequests        uint32    `json:"total_requests"`
	TotalSuccesses       uint32    `json:"total_successes"`
	TotalFailures        uint32    `json:"total_failures"`
	ConsecutiveSuccesses uint32    `json:"consecutive_successes"`
	ConsecutiveFailures  uint32    `json:"consecutive_failures"`
	SuccessRate          float64   `json:"success_rate"`
	FailureRate          float64   `json:"failure_rate"`
	LastStateChange      time.Time `json:"last_state_change,omitempty"`
}

// GetCircuitBreakerMetrics returns detailed metrics for all circuit breakers
func GetCircuitBreakerMetrics() map[string]CircuitBreakerMetrics {
	breakers := ListCircuitBreakers()
	metrics := make(map[string]CircuitBreakerMetrics)

	for name, info := range breakers {
		var successRate, failureRate float64
		if info.Counts.Requests > 0 {
			successRate = float64(info.Counts.TotalSuccesses) / float64(info.Counts.Requests)
			failureRate = float64(info.Counts.TotalFailures) / float64(info.Counts.Requests)
		}

		metrics[name] = CircuitBreakerMetrics{
			Name:                 name,
			State:                info.State.String(),
			TotalRequests:        info.Counts.Requests,
			TotalSuccesses:       info.Counts.TotalSuccesses,
			TotalFailures:        info.Counts.TotalFailures,
			ConsecutiveSuccesses: info.Counts.ConsecutiveSuccesses,
			ConsecutiveFailures:  info.Counts.ConsecutiveFailures,
			SuccessRate:          successRate,
			FailureRate:          failureRate,
		}
	}

	return metrics
}

// CircuitBreakerHealthCheck checks if any circuit breakers are in unhealthy state
func CircuitBreakerHealthCheck() (bool, string) {
	stats := GetCircuitBreakerStats()

	if stats.OpenBreakers > 0 {
		return false, fmt.Sprintf("%d circuit breakers are open", stats.OpenBreakers)
	}

	// Check if any circuit breakers have high failure rates
	metrics := GetCircuitBreakerMetrics()
	unhealthyBreakers := make([]string, 0)

	for name, metric := range metrics {
		if metric.TotalRequests >= 10 && metric.FailureRate >= 0.5 {
			unhealthyBreakers = append(unhealthyBreakers, name)
		}
	}

	if len(unhealthyBreakers) > 0 {
		return false, fmt.Sprintf("circuit breakers with high failure rates: %v", unhealthyBreakers)
	}

	return true, "all circuit breakers healthy"
}
