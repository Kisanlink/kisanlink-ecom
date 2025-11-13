// Package aaa provides a production-ready AAA service client with connection pooling,
// circuit breaker, retry logic, and comprehensive error handling for address management.
package aaa

import "time"

// Config contains all configuration for the AAA client
type Config struct {
	// Connection settings
	Address     string        `json:"address" yaml:"address"`
	PoolSize    int           `json:"pool_size" yaml:"pool_size"`
	DialTimeout time.Duration `json:"dial_timeout" yaml:"dial_timeout"`
	CallTimeout time.Duration `json:"call_timeout" yaml:"call_timeout"`

	// Retry settings
	MaxRetries     int           `json:"max_retries" yaml:"max_retries"`
	InitialBackoff time.Duration `json:"initial_backoff" yaml:"initial_backoff"`
	MaxBackoff     time.Duration `json:"max_backoff" yaml:"max_backoff"`
	BackoffFactor  float64       `json:"backoff_factor" yaml:"backoff_factor"`

	// Circuit breaker settings
	CircuitBreaker CircuitBreakerConfig `json:"circuit_breaker" yaml:"circuit_breaker"`

	// TLS settings
	EnableTLS  bool   `json:"enable_tls" yaml:"enable_tls"`
	CertFile   string `json:"cert_file" yaml:"cert_file"`
	KeyFile    string `json:"key_file" yaml:"key_file"`
	CAFile     string `json:"ca_file" yaml:"ca_file"`
	ServerName string `json:"server_name" yaml:"server_name"`
	SkipVerify bool   `json:"skip_verify" yaml:"skip_verify"`

	// Keep-alive settings
	KeepAliveTime    time.Duration `json:"keep_alive_time" yaml:"keep_alive_time"`
	KeepAliveTimeout time.Duration `json:"keep_alive_timeout" yaml:"keep_alive_timeout"`
}

// CircuitBreakerConfig contains circuit breaker configuration
type CircuitBreakerConfig struct {
	// MaxRequests is the maximum number of requests allowed in half-open state
	MaxRequests uint32 `json:"max_requests" yaml:"max_requests"`

	// Interval is the cyclic period of the closed state for the circuit breaker to clear the internal counts
	Interval time.Duration `json:"interval" yaml:"interval"`

	// Timeout is the period of the open state after which the state changes to half-open
	Timeout time.Duration `json:"timeout" yaml:"timeout"`

	// FailureRatio is the ratio of failures that will trip the circuit breaker to open state
	FailureRatio float64 `json:"failure_ratio" yaml:"failure_ratio"`

	// MinRequests is the minimum number of requests before failure ratio is calculated
	MinRequests uint32 `json:"min_requests" yaml:"min_requests"`
}

// DefaultConfig returns a Config with production-ready default values
func DefaultConfig() *Config {
	return &Config{
		Address:        "localhost:50051",
		PoolSize:       5,
		DialTimeout:    5 * time.Second,
		CallTimeout:    10 * time.Second,
		MaxRetries:     3,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     1 * time.Second,
		BackoffFactor:  2.0,
		CircuitBreaker: CircuitBreakerConfig{
			MaxRequests:  5,
			Interval:     60 * time.Second,
			Timeout:      60 * time.Second,
			FailureRatio: 0.6,
			MinRequests:  10,
		},
		EnableTLS:        false,
		KeepAliveTime:    30 * time.Second,
		KeepAliveTimeout: 10 * time.Second,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Address == "" {
		return ErrInvalidConfig{Field: "Address", Message: "address cannot be empty"}
	}
	if c.PoolSize < 1 {
		return ErrInvalidConfig{Field: "PoolSize", Message: "pool size must be at least 1"}
	}
	if c.DialTimeout <= 0 {
		return ErrInvalidConfig{Field: "DialTimeout", Message: "dial timeout must be positive"}
	}
	if c.CallTimeout <= 0 {
		return ErrInvalidConfig{Field: "CallTimeout", Message: "call timeout must be positive"}
	}
	if c.MaxRetries < 0 {
		return ErrInvalidConfig{Field: "MaxRetries", Message: "max retries cannot be negative"}
	}
	if c.CircuitBreaker.FailureRatio < 0 || c.CircuitBreaker.FailureRatio > 1 {
		return ErrInvalidConfig{Field: "CircuitBreaker.FailureRatio", Message: "failure ratio must be between 0 and 1"}
	}
	if c.EnableTLS && c.CertFile == "" {
		return ErrInvalidConfig{Field: "CertFile", Message: "cert file required when TLS is enabled"}
	}
	return nil
}
