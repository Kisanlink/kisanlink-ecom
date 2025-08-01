package repositories

import (
	"context"
	"testing"

	"kisanlink-ecom/internal/models/user"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepositoryIntegration(t *testing.T) {
	repo := NewUserRepository()
	ctx := context.Background()

	t.Run("CreateUser", func(t *testing.T) {
		// Test successful user creation
		testUser := &user.User{
			Username: "testuser",
			Email:    "test@example.com",
			FullName: "Test User",
			Role:     user.RoleCustomer,
			Status:   user.StatusActive,
		}

		err := repo.Create(ctx, testUser)
		assert.NoError(t, err)
		assert.NotEmpty(t, testUser.ID)
		assert.NotZero(t, testUser.CreatedAt)
		assert.NotZero(t, testUser.UpdatedAt)

		// Test duplicate email
		duplicateUser := &user.User{
			Username: "duplicate",
			Email:    testUser.Email, // Same email
			FullName: "Duplicate User",
			Role:     user.RoleCustomer,
			Status:   user.StatusActive,
		}

		err = repo.Create(ctx, duplicateUser)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("GetByID", func(t *testing.T) {
		// Create test user
		testUser := &user.User{
			Username: "getbyid",
			Email:    "getbyid@example.com",
			FullName: "Get By ID User",
			Role:     user.RoleCustomer,
			Status:   user.StatusActive,
		}

		err := repo.Create(ctx, testUser)
		require.NoError(t, err)

		// Test successful retrieval
		retrievedUser, err := repo.GetByID(ctx, testUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, testUser.ID, retrievedUser.ID)
		assert.Equal(t, testUser.Email, retrievedUser.Email)
		assert.Equal(t, testUser.Username, retrievedUser.Username)
		assert.Equal(t, testUser.FullName, retrievedUser.FullName)
		assert.Equal(t, testUser.Role, retrievedUser.Role)
		assert.Equal(t, testUser.Status, retrievedUser.Status)

		// Test non-existent user
		_, err = repo.GetByID(ctx, "non-existent-id")
		assert.Error(t, err)
	})

	t.Run("GetByEmail", func(t *testing.T) {
		// Create test user
		testUser := &user.User{
			Username: "getbyemail",
			Email:    "getbyemail@example.com",
			FullName: "Get By Email User",
			Role:     user.RoleCustomer,
			Status:   user.StatusActive,
		}

		err := repo.Create(ctx, testUser)
		require.NoError(t, err)

		// Test successful retrieval
		retrievedUser, err := repo.GetByEmail(ctx, testUser.Email)
		assert.NoError(t, err)
		assert.Equal(t, testUser.ID, retrievedUser.ID)
		assert.Equal(t, testUser.Email, retrievedUser.Email)

		// Test non-existent email
		_, err = repo.GetByEmail(ctx, "nonexistent@example.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("UpdateUser", func(t *testing.T) {
		// Create test user
		testUser := &user.User{
			Username: "updateuser",
			Email:    "updateuser@example.com",
			FullName: "Update User",
			Role:     user.RoleCustomer,
			Status:   user.StatusActive,
		}

		err := repo.Create(ctx, testUser)
		require.NoError(t, err)

		originalUpdatedAt := testUser.UpdatedAt

		// Update user
		testUser.FullName = "Updated User Name"
		testUser.Role = user.RoleAdmin
		testUser.Status = user.StatusInactive

		err = repo.Update(ctx, testUser)
		assert.NoError(t, err)

		// Verify update
		updatedUser, err := repo.GetByID(ctx, testUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated User Name", updatedUser.FullName)
		assert.Equal(t, user.RoleAdmin, updatedUser.Role)
		assert.Equal(t, user.StatusInactive, updatedUser.Status)
		assert.True(t, updatedUser.UpdatedAt.After(originalUpdatedAt))

		// Test updating non-existent user
		nonExistentUser := &user.User{
			Username: "nonexistent",
			Email:    "nonexistent@example.com",
			FullName: "Non Existent User",
			Role:     user.RoleCustomer,
			Status:   user.StatusActive,
		}
		nonExistentUser.ID = "non-existent-id"

		err = repo.Update(ctx, nonExistentUser)
		assert.Error(t, err)
	})

	t.Run("DeleteUser", func(t *testing.T) {
		// Create test user
		testUser := &user.User{
			Username: "deleteuser",
			Email:    "deleteuser@example.com",
			FullName: "Delete User",
			Role:     user.RoleCustomer,
			Status:   user.StatusActive,
		}

		err := repo.Create(ctx, testUser)
		require.NoError(t, err)

		// Delete user
		err = repo.Delete(ctx, testUser.ID)
		assert.NoError(t, err)

		// Verify deletion
		_, err = repo.GetByID(ctx, testUser.ID)
		assert.Error(t, err)

		// Test deleting non-existent user
		err = repo.Delete(ctx, "non-existent-id")
		assert.Error(t, err)
	})

	t.Run("FindUsers", func(t *testing.T) {
		// Create multiple test users
		users := []*user.User{
			{
				Username: "user1",
				Email:    "user1@example.com",
				FullName: "User One",
				Role:     user.RoleCustomer,
				Status:   user.StatusActive,
			},
			{
				Username: "user2",
				Email:    "user2@example.com",
				FullName: "User Two",
				Role:     user.RoleAdmin,
				Status:   user.StatusActive,
			},
			{
				Username: "user3",
				Email:    "user3@example.com",
				FullName: "User Three",
				Role:     user.RoleCustomer,
				Status:   user.StatusInactive,
			},
		}

		for _, u := range users {
			err := repo.Create(ctx, u)
			require.NoError(t, err)
		}

		// Test finding all users
		foundUsers, err := repo.Find(ctx, nil)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(foundUsers), len(users))

		// Test finding users with filter
		filter := base.NewFilter()
		filter.Group.Conditions = []base.FilterCondition{
			{
				Field:    "role",
				Operator: base.OpEqual,
				Value:    user.RoleCustomer,
			},
		}

		customerUsers, err := repo.Find(ctx, filter)
		assert.NoError(t, err)
		for _, u := range customerUsers {
			assert.Equal(t, user.RoleCustomer, u.Role)
		}
	})

	t.Run("SoftDeleteAndRestore", func(t *testing.T) {
		// Create test user
		testUser := &user.User{
			Username: "softdelete",
			Email:    "softdelete@example.com",
			FullName: "Soft Delete User",
			Role:     user.RoleCustomer,
			Status:   user.StatusActive,
		}

		err := repo.Create(ctx, testUser)
		require.NoError(t, err)

		// Soft delete user
		err = repo.SoftDelete(ctx, testUser.ID, "test-admin")
		assert.NoError(t, err)

		// Verify soft deletion
		_, err = repo.GetByID(ctx, testUser.ID)
		assert.Error(t, err) // Should not find the user

		// Find with deleted users
		deletedUsers, err := repo.ListWithDeleted(ctx, 100, 0)
		assert.NoError(t, err)
		found := false
		for _, u := range deletedUsers {
			if u.ID == testUser.ID {
				found = true
				assert.NotNil(t, u.DeletedAt)
				assert.Equal(t, "test-admin", *u.DeletedBy)
				break
			}
		}
		assert.True(t, found)

		// Restore user
		err = repo.Restore(ctx, testUser.ID)
		assert.NoError(t, err)

		// Verify restoration
		restoredUser, err := repo.GetByID(ctx, testUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, testUser.ID, restoredUser.ID)
		assert.Nil(t, restoredUser.DeletedAt)
		assert.Nil(t, restoredUser.DeletedBy)
	})

	t.Run("CountOperations", func(t *testing.T) {
		// Test count all users
		totalCount, err := repo.Count(ctx)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, totalCount, int64(0))

		// Test count with deleted users
		totalWithDeleted, err := repo.CountWithDeleted(ctx)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, totalWithDeleted, totalCount)
	})

	t.Run("ExistsOperations", func(t *testing.T) {
		// Create test user
		testUser := &user.User{
			Username: "existsuser",
			Email:    "existsuser@example.com",
			FullName: "Exists User",
			Role:     user.RoleCustomer,
			Status:   user.StatusActive,
		}

		err := repo.Create(ctx, testUser)
		require.NoError(t, err)

		// Test exists
		exists, err := repo.Exists(ctx, testUser.ID)
		assert.NoError(t, err)
		assert.True(t, exists)

		// Test non-existent user
		exists, err = repo.Exists(ctx, "non-existent-id")
		assert.NoError(t, err)
		assert.False(t, exists)

		// Test exists with deleted
		exists, err = repo.ExistsWithDeleted(ctx, testUser.ID)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("BulkOperations", func(t *testing.T) {
		// Create multiple users for bulk operations
		users := []*user.User{
			{
				Username: "bulk1",
				Email:    "bulk1@example.com",
				FullName: "Bulk User 1",
				Role:     user.RoleCustomer,
				Status:   user.StatusActive,
			},
			{
				Username: "bulk2",
				Email:    "bulk2@example.com",
				FullName: "Bulk User 2",
				Role:     user.RoleCustomer,
				Status:   user.StatusActive,
			},
			{
				Username: "bulk3",
				Email:    "bulk3@example.com",
				FullName: "Bulk User 3",
				Role:     user.RoleCustomer,
				Status:   user.StatusActive,
			},
		}

		// Test bulk create
		err := repo.CreateMany(ctx, users)
		assert.NoError(t, err)

		// Verify all users were created
		for _, u := range users {
			assert.NotEmpty(t, u.ID)
		}

		// Test bulk update
		for _, u := range users {
			u.FullName = u.FullName + " (Updated)"
		}

		err = repo.UpdateMany(ctx, users)
		assert.NoError(t, err)

		// Verify updates
		for _, u := range users {
			updatedUser, err := repo.GetByID(ctx, u.ID)
			assert.NoError(t, err)
			assert.Contains(t, updatedUser.FullName, "(Updated)")
		}

		// Test bulk delete
		userIDs := make([]string, len(users))
		for i, u := range users {
			userIDs[i] = u.ID
		}

		err = repo.DeleteMany(ctx, userIDs)
		assert.NoError(t, err)

		// Verify deletions
		for _, id := range userIDs {
			_, err := repo.GetByID(ctx, id)
			assert.Error(t, err)
		}
	})
}
