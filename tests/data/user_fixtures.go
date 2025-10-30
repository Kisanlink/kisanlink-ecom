package data

import (
	"time"

	"kisanlink-ecom/entities/models/user"
)

// CreateTestUser creates a test user with default values
func CreateTestUser() *user.User {
	return &user.User{
		Id:        "user-test-id-123",
		Username:  "testuser",
		Email:     "testuser@example.com",
		FirstName: "Test",
		LastName:  "User",
		Phone:     "+1234567890",
		IsActive:  true,
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now(),
	}
}

// CreateTestUserWithID creates a test user with a specific ID
func CreateTestUserWithID(id string) *user.User {
	u := CreateTestUser()
	u.Id = id
	return u
}

// CreateTestUserWithUsername creates a test user with a specific username
func CreateTestUserWithUsername(username string) *user.User {
	u := CreateTestUser()
	u.Username = username
	u.Email = username + "@example.com"
	return u
}

// CreateTestUserWithEmail creates a test user with a specific email
func CreateTestUserWithEmail(email string) *user.User {
	u := CreateTestUser()
	u.Email = email
	return u
}

// CreateTestInactiveUser creates an inactive test user
func CreateTestInactiveUser() *user.User {
	u := CreateTestUser()
	u.IsActive = false
	return u
}

// CreateTestUsersArray creates a slice of test users
func CreateTestUsersArray(count int) []*user.User {
	users := make([]*user.User, count)
	for i := 0; i < count; i++ {
		users[i] = CreateTestUserWithID(generateID(i))
		users[i].Username = generateUsername(i)
		users[i].Email = generateEmail(i)
	}
	return users
}

// Helper functions
func generateID(index int) string {
	return "user-test-id-" + string(rune('A'+index))
}

func generateUsername(index int) string {
	return "testuser" + string(rune('A'+index))
}

func generateEmail(index int) string {
	return "testuser" + string(rune('A'+index)) + "@example.com"
}
