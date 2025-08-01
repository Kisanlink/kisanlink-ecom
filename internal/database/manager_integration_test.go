package database

import (
	"context"
	"testing"
	"time"

	"kisanlink-ecom/internal/config"
	"kisanlink-ecom/internal/models/order"
	"kisanlink-ecom/internal/models/product"
	"kisanlink-ecom/internal/models/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseManagerIntegration(t *testing.T) {
	// Test with in-memory provider
	t.Run("InMemoryProvider", func(t *testing.T) {
		cfg := &config.Config{
			Database: config.DatabaseConfig{
				Provider: "inmemory",
			},
		}

		manager, err := NewManager(cfg)
		require.NoError(t, err)
		require.NotNil(t, manager)

		// Test connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = manager.Connect(ctx)
		assert.NoError(t, err)

		// Test connection status
		assert.True(t, manager.IsConnected())

		// Test repositories
		userRepo := manager.GetUserRepository()
		assert.NotNil(t, userRepo)

		productRepo := manager.GetProductRepository()
		assert.NotNil(t, productRepo)

		orderRepo := manager.GetOrderRepository()
		assert.NotNil(t, orderRepo)

		// Test health check
		health := manager.Health(ctx)
		assert.NotEmpty(t, health)
		assert.Equal(t, "ok", health["status"])

		// Test close
		err = manager.Close()
		assert.NoError(t, err)
	})

	// Test with invalid provider
	t.Run("InvalidProvider", func(t *testing.T) {
		cfg := &config.Config{
			Database: config.DatabaseConfig{
				Provider: "invalid_provider",
			},
		}

		manager, err := NewManager(cfg)
		require.NoError(t, err) // Should fallback to in-memory
		require.NotNil(t, manager)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = manager.Connect(ctx)
		assert.NoError(t, err)
		assert.True(t, manager.IsConnected())
	})
}

func TestDatabaseManagerRepositories(t *testing.T) {
	// Setup test manager
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Provider: "inmemory",
		},
	}

	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = manager.Connect(ctx)
	require.NoError(t, err)

	t.Run("UserRepositoryOperations", func(t *testing.T) {
		userRepo := manager.GetUserRepository()
		ctx := context.Background()

		// Create test user
		testUser := &user.User{
			Username: "testuser",
			Email:    "test@example.com",
			FullName: "Test User",
			Role:     user.RoleCustomer,
			Status:   user.StatusActive,
		}

		err := userRepo.Create(ctx, testUser)
		assert.NoError(t, err)
		assert.NotEmpty(t, testUser.ID)

		// Retrieve user by ID
		retrievedUser, err := userRepo.GetByID(ctx, testUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, testUser.Email, retrievedUser.Email)
		assert.Equal(t, testUser.Username, retrievedUser.Username)

		// Retrieve user by email
		userByEmail, err := userRepo.GetByEmail(ctx, testUser.Email)
		assert.NoError(t, err)
		assert.Equal(t, testUser.ID, userByEmail.ID)

		// Test duplicate email
		duplicateUser := &user.User{
			Username: "duplicate",
			Email:    testUser.Email, // Same email
			FullName: "Duplicate User",
			Role:     user.RoleCustomer,
			Status:   user.StatusActive,
		}
		err = userRepo.Create(ctx, duplicateUser)
		assert.Error(t, err) // Should fail due to duplicate email

		// Update user
		testUser.FullName = "Updated User"
		err = userRepo.Update(ctx, testUser)
		assert.NoError(t, err)

		// Verify update
		updatedUser, err := userRepo.GetByID(ctx, testUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated User", updatedUser.FullName)

		// Delete user
		err = userRepo.Delete(ctx, testUser.ID)
		assert.NoError(t, err)

		// Verify deletion
		_, err = userRepo.GetByID(ctx, testUser.ID)
		assert.Error(t, err) // Should not find the user
	})

	t.Run("ProductRepositoryOperations", func(t *testing.T) {
		productRepo := manager.GetProductRepository()
		ctx := context.Background()

		// Create test product
		testProduct := &product.Product{
			Name:        "Test Product",
			Description: "Test Description",
			Price:       99.99,
			Category:    "Electronics",
			Stock:       100,
		}

		err := productRepo.Create(ctx, testProduct)
		assert.NoError(t, err)
		assert.NotEmpty(t, testProduct.ID)

		// Retrieve product
		retrievedProduct, err := productRepo.GetByID(ctx, testProduct.ID)
		assert.NoError(t, err)
		assert.Equal(t, testProduct.Name, retrievedProduct.Name)
		assert.Equal(t, testProduct.Price, retrievedProduct.Price)
	})

	t.Run("OrderRepositoryOperations", func(t *testing.T) {
		orderRepo := manager.GetOrderRepository()
		userRepo := manager.GetUserRepository()
		ctx := context.Background()

		// Create test user
		testUser := &user.User{
			Username: "orderuser",
			Email:    "order@example.com",
			FullName: "Order User",
			Role:     user.RoleCustomer,
			Status:   user.StatusActive,
		}

		err := userRepo.Create(ctx, testUser)
		assert.NoError(t, err)

		// Create test order
		testOrder := &order.Order{
			UserID:     testUser.ID,
			Status:     order.OrderStatusPending,
			TotalPrice: 199.99,
			Items:      []order.OrderItem{},
		}

		err = orderRepo.Create(ctx, testOrder)
		assert.NoError(t, err)
		assert.NotEmpty(t, testOrder.ID)

		// Retrieve order
		retrievedOrder, err := orderRepo.GetByID(ctx, testOrder.ID)
		assert.NoError(t, err)
		assert.Equal(t, testOrder.UserID, retrievedOrder.UserID)
		assert.Equal(t, testOrder.Status, retrievedOrder.Status)
	})
}

func TestDatabaseManagerHealth(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Provider: "inmemory",
		},
	}

	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = manager.Connect(ctx)
	require.NoError(t, err)

	t.Run("HealthCheck", func(t *testing.T) {
		health := manager.Health(ctx)

		// Check required health fields
		assert.NotEmpty(t, health)
		assert.Contains(t, health, "status")
		assert.Contains(t, health, "timestamp")
		assert.Equal(t, "ok", health["status"])
	})
}

func TestDatabaseManagerMigration(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Provider: "inmemory",
		},
	}

	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = manager.Connect(ctx)
	require.NoError(t, err)

	t.Run("Migration", func(t *testing.T) {
		err := manager.Migrate(ctx)
		assert.NoError(t, err)
	})
}

func TestDatabaseManagerErrorHandling(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Provider: "inmemory",
		},
	}

	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = manager.Connect(ctx)
	require.NoError(t, err)

	t.Run("InvalidOperations", func(t *testing.T) {
		userRepo := manager.GetUserRepository()
		ctx := context.Background()

		// Test getting non-existent user
		_, err := userRepo.GetByID(ctx, "non-existent-id")
		assert.Error(t, err)

		// Test getting user by non-existent email
		_, err = userRepo.GetByEmail(ctx, "nonexistent@example.com")
		assert.Error(t, err)

		// Test updating non-existent user
		nonExistentUser := &user.User{
			Username: "test",
			Email:    "test@example.com",
			FullName: "Test User",
			Role:     user.RoleCustomer,
			Status:   user.StatusActive,
		}
		// Set ID manually for testing
		nonExistentUser.ID = "non-existent-id"
		err = userRepo.Update(ctx, nonExistentUser)
		assert.Error(t, err)

		// Test deleting non-existent user
		err = userRepo.Delete(ctx, "non-existent-id")
		assert.Error(t, err)
	})
}
