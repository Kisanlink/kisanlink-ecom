package config

import "time"

// GRPCServerConfig holds gRPC server configuration
type GRPCServerConfig struct {
	Port                  int
	MaxRecvMsgSize        int
	MaxSendMsgSize        int
	KeepaliveTime         time.Duration
	KeepaliveTimeout      time.Duration
	ReflectionEnabled     bool
	MaxConnectionIdle     time.Duration
	MaxConnectionAge      time.Duration
	MaxConnectionAgeGrace time.Duration
	Timeout               time.Duration
	EnableMetrics         bool
	EnableTracing         bool
}

// LoadGRPCServerConfig loads gRPC server configuration from environment
func LoadGRPCServerConfig() *GRPCServerConfig {
	return &GRPCServerConfig{
		Port:                  getEnvAsInt("GRPC_SERVER_PORT", 50051),
		MaxRecvMsgSize:        getEnvAsInt("GRPC_MAX_RECV_MSG_SIZE", 10*1024*1024), // 10MB
		MaxSendMsgSize:        getEnvAsInt("GRPC_MAX_SEND_MSG_SIZE", 10*1024*1024), // 10MB
		KeepaliveTime:         time.Duration(getEnvAsInt("GRPC_KEEPALIVE_TIME_SECONDS", 30)) * time.Second,
		KeepaliveTimeout:      time.Duration(getEnvAsInt("GRPC_KEEPALIVE_TIMEOUT_SECONDS", 10)) * time.Second,
		ReflectionEnabled:     getEnvAsBool("GRPC_REFLECTION_ENABLED", true),
		MaxConnectionIdle:     time.Duration(getEnvAsInt("GRPC_MAX_CONNECTION_IDLE_SECONDS", 300)) * time.Second,
		MaxConnectionAge:      time.Duration(getEnvAsInt("GRPC_MAX_CONNECTION_AGE_SECONDS", 600)) * time.Second,
		MaxConnectionAgeGrace: time.Duration(getEnvAsInt("GRPC_MAX_CONNECTION_AGE_GRACE_SECONDS", 60)) * time.Second,
		Timeout:               time.Duration(getEnvAsInt("GRPC_TIMEOUT_SECONDS", 30)) * time.Second,
		EnableMetrics:         getEnvAsBool("GRPC_ENABLE_METRICS", true),
		EnableTracing:         getEnvAsBool("GRPC_ENABLE_TRACING", true),
	}
}
