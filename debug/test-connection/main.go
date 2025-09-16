package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"kisanlink-ecom/internal/auth"
	"kisanlink-ecom/internal/config"
)

func main() {
	testAAAConnection()
}

func testAAAConnection() {
	fmt.Println("Testing AAA Service gRPC Connection...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize AAA client
	aaaClient, err := auth.NewAAAClient(&cfg.AAA)
	if err != nil {
		log.Fatalf("Failed to create AAA client: %v", err)
	}

	fmt.Printf("✓ AAA client created successfully\n")
	fmt.Printf("  - Endpoint: %s\n", cfg.AAA.GRPCServerAddr)

	// Test basic connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test user creation (mock)
	testUser := &auth.AAAUser{
		ID:       "test_user_001",
		Username: "test_user",
		Email:    "test@example.com",
		IsActive: true,
	}

	fmt.Println("\n--- Testing User Operations ---")
	user, err := aaaClient.CreateUser(ctx, testUser)
	if err != nil {
		fmt.Printf("✗ CreateUser failed: %v\n", err)
	} else {
		fmt.Printf("✓ CreateUser succeeded: %+v\n", user)
	}

	// Test getting user
	user, err = aaaClient.GetUser(ctx, "test_user_001")
	if err != nil {
		fmt.Printf("✗ GetUser failed: %v\n", err)
	} else {
		fmt.Printf("✓ GetUser succeeded: %+v\n", user)
	}

	// Test authentication
	authUser, err := aaaClient.AuthenticateUser(ctx, "test_user", "password123")
	if err != nil {
		fmt.Printf("✗ AuthenticateUser failed: %v\n", err)
	} else {
		fmt.Printf("✓ AuthenticateUser succeeded: %+v\n", authUser)
	}

	fmt.Println("\n--- Testing Role Operations ---")
	// Test getting user roles
	roles, err := aaaClient.GetUserRoles(ctx, "test_user_001")
	if err != nil {
		fmt.Printf("✗ GetUserRoles failed: %v\n", err)
	} else {
		fmt.Printf("✓ GetUserRoles succeeded: %d roles found\n", len(roles))
		for i, role := range roles {
			fmt.Printf("  Role %d: %+v\n", i+1, role)
		}
	}

	fmt.Println("\n--- Testing Permission Operations ---")
	// Test permission evaluation
	allowed, err := aaaClient.EvaluatePermission(ctx, "test_user_001", "product", "create")
	if err != nil {
		fmt.Printf("✗ EvaluatePermission failed: %v\n", err)
	} else {
		fmt.Printf("✓ EvaluatePermission succeeded: allowed=%v\n", allowed)
	}

	// Test bulk permission evaluation
	checks := []auth.PermissionCheck{
		{Resource: "product", Action: "create"},
		{Resource: "product", Action: "read"},
		{Resource: "order", Action: "create"},
		{Resource: "order", Action: "read"},
	}

	results, err := aaaClient.BulkEvaluatePermissions(ctx, "test_user_001", checks)
	if err != nil {
		fmt.Printf("✗ BulkEvaluatePermissions failed: %v\n", err)
	} else {
		fmt.Printf("✓ BulkEvaluatePermissions succeeded: %d results\n", len(results))
		for i, result := range results {
			fmt.Printf("  Check %d: %s.%s = %v (%s)\n", i+1, result.Resource, result.Action, result.Allowed, result.Reason)
		}
	}

	fmt.Println("\n🎉 AAA Service connection test completed!")
}
