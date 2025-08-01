package testutils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"kisanlink-ecom/internal/config"
	"kisanlink-ecom/internal/database"
	"kisanlink-ecom/internal/models/user"
	"kisanlink-ecom/internal/repositories"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestConfig holds test configuration
type TestConfig struct {
	DatabaseProvider string
	DatabaseHost     string
	DatabasePort     int
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
	LogLevel         string
}

// DefaultTestConfig returns default test configuration
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		DatabaseProvider: "inmemory", // Use in-memory for unit tests
		DatabaseHost:     "localhost",
		DatabasePort:     5432,
		DatabaseUser:     "test",
		DatabasePassword: "test",
		DatabaseName:     "test_db",
		LogLevel:         "debug",
	}
}

// TestDatabaseManager creates a test database manager
func TestDatabaseManager(t *testing.T, cfg *TestConfig) *database.Manager {
	config := &config.Config{
		Database: config.DatabaseConfig{
			Provider: cfg.DatabaseProvider,
			Host:     cfg.DatabaseHost,
			Port:     cfg.DatabasePort,
			User:     cfg.DatabaseUser,
			Password: cfg.DatabasePassword,
			Name:     cfg.DatabaseName,
			SSLMode:  "disable",
		},
	}

	manager, err := database.NewManager(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = manager.Connect(ctx)
	require.NoError(t, err)

	return manager
}

// CleanupTestData cleans up test data
func CleanupTestData(t *testing.T, manager *database.Manager) {
	ctx := context.Background()

	// Clean up users
	users, err := manager.GetUserRepository().Find(ctx, nil)
	if err == nil {
		for _, u := range users {
			_ = manager.GetUserRepository().Delete(ctx, u.ID)
		}
	}

	// Clean up products
	products, err := manager.GetProductRepository().Find(ctx, nil)
	if err == nil {
		for _, p := range products {
			_ = manager.GetProductRepository().Delete(ctx, p.ID)
		}
	}

	// Clean up orders
	orders, err := manager.GetOrderRepository().Find(ctx, nil)
	if err == nil {
		for _, o := range orders {
			_ = manager.GetOrderRepository().Delete(ctx, o.ID)
		}
	}
}

// CreateTestUser creates a test user
func CreateTestUser(t *testing.T, repo *repositories.UserRepository, email string) *user.User {
	user := &user.User{
		Username: "testuser",
		Email:    email,
		FullName: "Test User",
		Role:     user.RoleCustomer,
		Status:   user.StatusActive,
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)
	return user
}

// SetupTestRouter creates a test router with middleware
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add common middleware for testing
	router.Use(gin.Recovery())

	return router
}

// MakeTestRequest makes a test HTTP request and returns the response
func MakeTestRequest(t *testing.T, router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var req *http.Request
	var err error

	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		req, err = http.NewRequest(method, path, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, path, nil)
	}
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// AssertJSONResponse asserts a JSON response
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedSuccess bool) {
	assert.Equal(t, expectedStatus, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	if expectedSuccess {
		assert.True(t, response["success"].(bool))
	} else {
		assert.False(t, response["success"].(bool))
	}
}

// TestLogger creates a test logger
func TestLogger() *zap.Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("failed to create test logger: %v", err))
	}
	return logger
}

// SetTestEnvironment sets test environment variables
func SetTestEnvironment() {
	os.Setenv("ENV", "test")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("DATABASE_PROVIDER", "inmemory")
}

// CleanupTestEnvironment cleans up test environment variables
func CleanupTestEnvironment() {
	os.Unsetenv("ENV")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("DATABASE_PROVIDER")
}

// TestSuite provides a base test suite with common setup/teardown
type TestSuite struct {
	Manager *database.Manager
	Router  *gin.Engine
	T       *testing.T
}

// NewTestSuite creates a new test suite
func NewTestSuite(t *testing.T) *TestSuite {
	SetTestEnvironment()

	manager := TestDatabaseManager(t, DefaultTestConfig())
	router := SetupTestRouter()

	return &TestSuite{
		Manager: manager,
		Router:  router,
		T:       t,
	}
}

// Setup runs before each test
func (ts *TestSuite) Setup() {
	// Clean up any existing test data
	CleanupTestData(ts.T, ts.Manager)
}

// Teardown runs after each test
func (ts *TestSuite) Teardown() {
	CleanupTestData(ts.T, ts.Manager)
	if ts.Manager != nil {
		_ = ts.Manager.Close()
	}
	CleanupTestEnvironment()
}

// RunTest runs a test with setup and teardown
func (ts *TestSuite) RunTest(name string, testFunc func()) {
	ts.T.Run(name, func(t *testing.T) {
		ts.Setup()
		defer ts.Teardown()
		testFunc()
	})
}
