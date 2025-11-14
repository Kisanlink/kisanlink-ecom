// Package grpc provides gRPC server initialization and configuration
package grpc

import (
	"fmt"

	"github.com/Kisanlink/kisanlink-ecom/internal/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// NewServer creates a new gRPC server with proper configuration
func NewServer(cfg *config.GRPCServerConfig, interceptors []grpc.UnaryServerInterceptor) *grpc.Server {
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(cfg.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(cfg.MaxSendMsgSize),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     cfg.MaxConnectionIdle,
			MaxConnectionAge:      cfg.MaxConnectionAge,
			MaxConnectionAgeGrace: cfg.MaxConnectionAgeGrace,
			Time:                  cfg.KeepaliveTime,
			Timeout:               cfg.KeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             cfg.KeepaliveTime / 2,
			PermitWithoutStream: true,
		}),
	}

	// Chain interceptors if provided
	if len(interceptors) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(interceptors...))
	}

	return grpc.NewServer(opts...)
}

// GetServerAddress returns the server address string
func GetServerAddress(cfg *config.GRPCServerConfig) string {
	return fmt.Sprintf(":%d", cfg.Port)
}
