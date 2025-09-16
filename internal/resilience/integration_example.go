package resilience

import (
	"context"
	"time"

	catalogModels "kisanlink-ecom/entities/models/catalog"
	"kisanlink-ecom/internal/auth"
	catalogService "kisanlink-ecom/internal/services/catalog"
)

// ResilientCatalogService wraps the catalog service with resilience patterns
type ResilientCatalogService struct {
	catalogService catalogService.CatalogServiceInterface
	dbWrapper      *DatabaseWrapper
	authWrapper    *AAAClientWrapper
}

// NewResilientCatalogService creates a resilient catalog service
func NewResilientCatalogService(
	catalogSvc catalogService.CatalogServiceInterface,
	aaaClient auth.AAAClient,
) *ResilientCatalogService {
	// Configure database resilience
	dbConfig := &ServiceConfig{
		ServiceName: "catalog_database",
		CircuitBreakerConfig: &CircuitBreakerConfig{
			Name:        "catalog_db_cb",
			MaxRequests: 3,
			Interval:    60 * time.Second,
			Timeout:     30 * time.Second,
			ReadyToTrip: func(counts Counts) bool {
				failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
				return counts.Requests >= 5 && failureRatio >= 0.6
			},
			OnStateChange: func(name string, from, to CircuitBreakerState) {
				// Log state changes for monitoring
				// logger.Info("Circuit breaker state change", "name", name, "from", from, "to", to)
			},
		},
		RetryConfig: &RetryConfig{
			MaxRetries:      3,
			InitialInterval: 100 * time.Millisecond,
			MaxInterval:     2 * time.Second,
			Multiplier:      2.0,
			RetryableErrorChecker: func(err error) bool {
				// Customize retry logic for database errors
				return DefaultRetryableErrorChecker(err)
			},
		},
		TimeoutConfig: &TimeoutConfig{
			Timeout: 10 * time.Second,
		},
	}

	// Configure AAA service resilience
	aaaConfig := &ServiceConfig{
		ServiceName: "aaa_service",
		CircuitBreakerConfig: &CircuitBreakerConfig{
			Name:        "aaa_service_cb",
			MaxRequests: 2,
			Interval:    30 * time.Second,
			Timeout:     15 * time.Second,
			ReadyToTrip: func(counts Counts) bool {
				return counts.Requests >= 3 && counts.TotalFailures >= 2
			},
		},
		RetryConfig: &RetryConfig{
			MaxRetries:      2,
			InitialInterval: 50 * time.Millisecond,
			MaxInterval:     1 * time.Second,
			Multiplier:      2.0,
		},
		TimeoutConfig: &TimeoutConfig{
			Timeout: 5 * time.Second,
		},
	}

	return &ResilientCatalogService{
		catalogService: catalogSvc,
		dbWrapper:      NewDatabaseWrapper(dbConfig),
		authWrapper:    NewAAAClientWrapper(aaaClient, aaaConfig),
	}
}

// CreateProduct creates a product with resilience patterns
func (r *ResilientCatalogService) CreateProduct(ctx context.Context, product *catalogModels.Product, userID string) (*catalogModels.Product, error) {
	var result *catalogModels.Product

	err := r.dbWrapper.Execute(ctx, func(ctx context.Context) error {
		var err error
		result, err = r.catalogService.CreateProduct(ctx, product, userID)
		return err
	})

	return result, err
}

// GetProductByID retrieves a product by ID with resilience
func (r *ResilientCatalogService) GetProductByID(ctx context.Context, productID string) (*catalogModels.Product, error) {
	var result *catalogModels.Product

	err := r.dbWrapper.Execute(ctx, func(ctx context.Context) error {
		var err error
		result, err = r.catalogService.GetProductByID(ctx, productID)
		return err
	})

	return result, err
}

// UpdateProduct updates a product with resilience
func (r *ResilientCatalogService) UpdateProduct(ctx context.Context, product *catalogModels.Product, userID string) (*catalogModels.Product, error) {
	var result *catalogModels.Product

	err := r.dbWrapper.Execute(ctx, func(ctx context.Context) error {
		var err error
		result, err = r.catalogService.UpdateProduct(ctx, product, userID)
		return err
	})

	return result, err
}

// DeleteProduct deletes a product with resilience
func (r *ResilientCatalogService) DeleteProduct(ctx context.Context, productID string, userID string) error {
	return r.dbWrapper.Execute(ctx, func(ctx context.Context) error {
		return r.catalogService.DeleteProduct(ctx, productID, userID)
	})
}

// ValidatePermission validates permissions using resilient AAA client
func (r *ResilientCatalogService) ValidatePermission(ctx context.Context, userID, resource, action string) (bool, error) {
	return r.authWrapper.EvaluatePermission(ctx, userID, resource, action)
}

// ResilientServiceFactory creates resilient service instances
type ResilientServiceFactory struct {
	aaaClient auth.AAAClient
	dbConfig  *ServiceConfig
	aaaConfig *ServiceConfig
}

// NewResilientServiceFactory creates a new resilient service factory
func NewResilientServiceFactory(aaaClient auth.AAAClient) *ResilientServiceFactory {
	return &ResilientServiceFactory{
		aaaClient: aaaClient,
		dbConfig:  DefaultServiceConfig("database"),
		aaaConfig: DefaultServiceConfig("aaa_service"),
	}
}

// SetDatabaseConfig sets the database resilience configuration
func (f *ResilientServiceFactory) SetDatabaseConfig(config *ServiceConfig) {
	f.dbConfig = config
}

// SetAAAConfig sets the AAA service resilience configuration
func (f *ResilientServiceFactory) SetAAAConfig(config *ServiceConfig) {
	f.aaaConfig = config
}

// CreateResilientCatalogService creates a resilient catalog service
func (f *ResilientServiceFactory) CreateResilientCatalogService(catalogSvc catalogService.CatalogServiceInterface) *ResilientCatalogService {
	return &ResilientCatalogService{
		catalogService: catalogSvc,
		dbWrapper:      NewDatabaseWrapper(f.dbConfig),
		authWrapper:    NewAAAClientWrapper(f.aaaClient, f.aaaConfig),
	}
}

// Example usage and configuration templates

// ExampleConfiguration shows how to configure resilience patterns
func ExampleConfiguration() {
	// 1. Basic circuit breaker configuration
	basicConfig := &CircuitBreakerConfig{
		Name:        "example_service",
		MaxRequests: 3,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts Counts) bool {
			return counts.Requests >= 5 && counts.TotalFailures >= 3
		},
	}

	// 2. Database-specific configuration
	dbConfig := &ServiceConfig{
		ServiceName: "database",
		CircuitBreakerConfig: &CircuitBreakerConfig{
			Name:        "database_cb",
			MaxRequests: 5,
			Interval:    120 * time.Second,
			Timeout:     60 * time.Second,
			ReadyToTrip: func(counts Counts) bool {
				failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
				return counts.Requests >= 10 && failureRatio >= 0.5
			},
		},
		RetryConfig: &RetryConfig{
			MaxRetries:      3,
			InitialInterval: 200 * time.Millisecond,
			MaxInterval:     5 * time.Second,
			Multiplier:      2.0,
		},
		TimeoutConfig: &TimeoutConfig{
			Timeout: 15 * time.Second,
		},
	}

	// 3. External service configuration
	externalConfig := &ServiceConfig{
		ServiceName: "external_api",
		CircuitBreakerConfig: &CircuitBreakerConfig{
			Name:        "external_api_cb",
			MaxRequests: 2,
			Interval:    30 * time.Second,
			Timeout:     10 * time.Second,
			ReadyToTrip: func(counts Counts) bool {
				return counts.Requests >= 3 && counts.TotalFailures >= 2
			},
		},
		RetryConfig: &RetryConfig{
			MaxRetries:      2,
			InitialInterval: 100 * time.Millisecond,
			MaxInterval:     2 * time.Second,
			Multiplier:      1.5,
		},
		TimeoutConfig: &TimeoutConfig{
			Timeout: 8 * time.Second,
		},
	}

	// Use the configurations
	_ = basicConfig
	_ = dbConfig
	_ = externalConfig
}

// ExampleUsage demonstrates how to use resilience patterns
func ExampleUsage() {
	// Example 1: Simple circuit breaker
	cb := GetCircuitBreaker("my_service")
	err := cb.Call(func() error {
		// Your service call here
		return nil
	})
	_ = err

	// Example 2: Circuit breaker with context
	ctx := context.Background()
	err = cb.CallContext(ctx, func(ctx context.Context) error {
		// Your service call with context
		return nil
	})
	_ = err

	// Example 3: Using convenience functions
	err = ExecuteWithCircuitBreaker("my_service", func() error {
		// Your operation here
		return nil
	})
	_ = err

	// Example 4: Custom configuration
	config := &CircuitBreakerConfig{
		Name:        "custom_service",
		MaxRequests: 3,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts Counts) bool {
			return counts.TotalFailures >= 5
		},
	}

	err = ExecuteWithCustomCircuitBreaker("custom_service", config, func() error {
		// Your operation here
		return nil
	})
	_ = err
}

// MonitoringExample shows how to monitor circuit breakers
func MonitoringExample() {
	// Get statistics for all circuit breakers
	stats := GetCircuitBreakerStats()
	_ = stats // Use for monitoring dashboards

	// Get detailed metrics
	metrics := GetCircuitBreakerMetrics()
	_ = metrics // Use for detailed analysis

	// Health check
	healthy, message := CircuitBreakerHealthCheck()
	_, _ = healthy, message // Use for service health endpoints

	// Get specific circuit breaker info
	cb := GetCircuitBreaker("my_service")
	state := cb.State()
	counts := cb.Counts()
	_, _ = state, counts // Use for service-specific monitoring
}

// IntegrationPatterns demonstrates common integration patterns
func IntegrationPatterns() {
	// Pattern 1: Database operations with resilience
	dbWrapper := CreateResilientDatabaseWrapper()
	ctx := context.Background()

	err := dbWrapper.Execute(ctx, func(ctx context.Context) error {
		// Database operation
		return nil
	})
	_ = err

	// Pattern 2: External service calls with resilience
	serviceWrapper := CreateResilientExternalService("payment_service")

	result, err := serviceWrapper.CallWithResult(ctx, func(ctx context.Context) (interface{}, error) {
		// External service call
		return "payment_id", nil
	})
	_, _ = result, err

	// Pattern 3: Combining multiple resilience patterns
	// This is handled by the service wrappers automatically
}

// Best practices and recommendations

/*
Best Practices for Circuit Breaker Configuration:

1. **Failure Threshold**: Set based on acceptable failure rate
   - Web services: 50-60% failure rate over 5-10 requests
   - Databases: 30-40% failure rate over 10-20 requests
   - External APIs: 60-80% failure rate over 3-5 requests

2. **Timeout Settings**:
   - Open timeout: 10-60 seconds for recovery
   - Request timeout: 5-30 seconds based on SLA
   - Half-open max requests: 1-5 for gradual recovery

3. **Retry Configuration**:
   - Max retries: 2-3 for external services, 3-5 for databases
   - Initial interval: 100-500ms
   - Max interval: 2-10 seconds
   - Use exponential backoff with jitter

4. **Monitoring and Alerting**:
   - Monitor circuit breaker state changes
   - Alert on circuit breaker opening
   - Track failure rates and response times
   - Dashboard for circuit breaker health

5. **Testing**:
   - Test circuit breaker behavior under load
   - Verify recovery mechanisms
   - Test graceful degradation
   - Chaos engineering for resilience validation
*/
