//go:build testclient
// +build testclient

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// Simple test client to verify gRPC server functionality
func main() {
	// Create connection
	conn, err := grpc.Dial("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	// Test health check
	healthClient := grpc_health_v1.NewHealthClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		log.Fatalf("health check failed: %v", err)
	}

	fmt.Printf("Health Check Response: %v\n", resp.Status)
	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		log.Fatalf("unexpected health status: %v", resp.Status)
	}

	fmt.Println("✅ Health check passed!")
	fmt.Println("✅ gRPC server is working correctly!")
	os.Exit(0)
}
