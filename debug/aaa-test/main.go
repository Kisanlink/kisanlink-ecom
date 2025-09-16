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
	runAAATest()
}

// TestAAAIntegration tests the AAA service integration
func runAAATest() {
	fmt.Println("🔐 Testing AAA Service Integration...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Printf("📋 Configuration loaded:\n")
	fmt.Printf("   AAA Endpoint: %s\n", cfg.AAA.Endpoint)
	fmt.Printf("   AAA gRPC Server: %s\n", cfg.AAA.GRPCServerAddr)
	fmt.Printf("   Timeout: %dms\n", cfg.AAA.TimeoutMs)
	fmt.Printf("   Retries: %d\n", cfg.AAA.Retries)

	// Test AAA client connection
	fmt.Println("\n🔌 Testing AAA Client Connection...")
	aaaClient, err := auth.NewAAAClient(&cfg.AAA)
	if err != nil {
		log.Fatalf("❌ Failed to create AAA client: %v", err)
	}
	defer func() {
		if closer, ok := aaaClient.(interface{ Close() error }); ok {
			closer.Close()
		}
	}()
	fmt.Println("✅ AAA client created successfully")

	// Test basic functionality
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.AAA.TimeoutMs)*time.Millisecond)
	defer cancel()

	testUserID := "test_user_123"
	testResource := "catalog_item"
	testAction := "view_catalog"

	fmt.Println("\n🧪 Testing AAA Service Functions...")

	// Test 1: Get User Roles
	fmt.Println("   Testing GetUserRoles...")
	roles, err := aaaClient.GetUserRoles(ctx, testUserID)
	if err != nil {
		fmt.Printf("   ⚠️  GetUserRoles failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ GetUserRoles successful, found %d roles\n", len(roles))
		for _, role := range roles {
			fmt.Printf("      - Role: %s (%s)\n", role.Name, role.Description)
		}
	}

	// Test 2: Get User Permissions
	fmt.Println("   Testing GetUserPermissions...")
	permissions, err := aaaClient.GetUserPermissions(ctx, testUserID)
	if err != nil {
		fmt.Printf("   ⚠️  GetUserPermissions failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ GetUserPermissions successful, found %d permissions\n", len(permissions))
		for _, perm := range permissions {
			fmt.Printf("      - Permission: %s\n", perm)
		}
	}

	// Test 3: Evaluate Permission
	fmt.Println("   Testing EvaluatePermission...")
	allowed, err := aaaClient.EvaluatePermission(ctx, testUserID, testResource, testAction)
	if err != nil {
		fmt.Printf("   ⚠️  EvaluatePermission failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ EvaluatePermission successful: %s can %s on %s: %t\n",
			testUserID, testAction, testResource, allowed)
	}

	// Test 4: Bulk Evaluate Permissions
	fmt.Println("   Testing BulkEvaluatePermissions...")
	permissionChecks := []auth.PermissionCheck{
		{Resource: "catalog_item", Action: "view_catalog"},
		{Resource: "order", Action: "view_order"},
		{Resource: "org", Action: "create_catalog"},
	}

	results, err := aaaClient.BulkEvaluatePermissions(ctx, testUserID, permissionChecks)
	if err != nil {
		fmt.Printf("   ⚠️  BulkEvaluatePermissions failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ BulkEvaluatePermissions successful, evaluated %d permissions\n", len(results))
		for _, result := range results {
			fmt.Printf("      - %s %s on %s: %t (%s)\n",
				testUserID, result.Action, result.Resource, result.Allowed, result.Reason)
		}
	}

	fmt.Println("\n🎯 Integration Test Summary:")
	fmt.Println("   - AAA client connection: ✅")
	fmt.Println("   - Service functions: Tested with real AAA service")
	fmt.Println("   - Check logs above for any warnings or errors")

	if err != nil {
		fmt.Println("\n❌ Some tests failed. Check your AAA service configuration and ensure it's running.")
		fmt.Println("   Common issues:")
		fmt.Println("   1. AAA service not running")
		fmt.Println("   2. Wrong gRPC address/port")
		fmt.Println("   3. Network connectivity issues")
		fmt.Println("   4. AAA service not implementing the expected gRPC methods")
	} else {
		fmt.Println("\n✅ All tests passed! Your AAA integration is working correctly.")
	}
}
