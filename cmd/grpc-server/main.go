// Package main implements the gRPC server for the Collaborator service
package main

import (
	"context"
	"kisanlink-ecom/internal/config"
	grpcserver "kisanlink-ecom/internal/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Load configuration
	cfg := config.LoadGRPCServerConfig()

	log.Printf("Starting gRPC server with configuration:")
	log.Printf("  Port: %d", cfg.Port)
	log.Printf("  Max Recv Message Size: %d bytes", cfg.MaxRecvMsgSize)
	log.Printf("  Max Send Message Size: %d bytes", cfg.MaxSendMsgSize)
	log.Printf("  Keepalive Time: %v", cfg.KeepaliveTime)
	log.Printf("  Keepalive Timeout: %v", cfg.KeepaliveTimeout)
	log.Printf("  Reflection Enabled: %v", cfg.ReflectionEnabled)

	// Create listener
	address := grpcserver.GetServerAddress(cfg)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", address, err)
	}

	// Create gRPC server with interceptors (will be added in P1-2)
	var interceptors []grpc.UnaryServerInterceptor
	// TODO: Add interceptors in P1-2

	server := grpcserver.NewServer(cfg, interceptors)

	// Register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register reflection for grpcurl and development
	if cfg.ReflectionEnabled {
		reflection.Register(server)
		log.Println("gRPC reflection enabled")
	}

	// TODO: Register collaborator service handler (will be added in P1-4)
	// collaboratorpb.RegisterCollaboratorServiceServer(server, collaboratorHandler)

	// Start server in goroutine
	go func() {
		log.Printf("gRPC server listening on %s", address)
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	log.Printf("Received signal: %v. Shutting down gRPC server...", sig)

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Set health check to NOT_SERVING during shutdown
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	// Create channel to signal shutdown completion
	done := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(done)
	}()

	// Wait for graceful shutdown to complete or timeout
	select {
	case <-done:
		log.Println("gRPC server stopped gracefully")
	case <-ctx.Done():
		log.Println("Graceful shutdown timeout, forcing stop")
		server.Stop()
	}

	log.Println("Server shutdown complete")
}
