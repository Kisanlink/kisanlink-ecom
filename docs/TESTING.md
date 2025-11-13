# Testing Guide

This document provides comprehensive information about testing in the KisanLink E-commerce project, including test utilities, coverage reporting, and integration tests.

## Table of Contents

- [Overview](#overview)
- [Test Structure](#test-structure)
- [Test Utilities](#test-utilities)
- [Running Tests](#running-tests)
- [Coverage Reporting](#coverage-reporting)
- [Integration Tests](#integration-tests)
- [Test Configuration](#test-configuration)
- [Best Practices](#best-practices)

## Overview

The project uses a comprehensive testing strategy with:

- **Unit Tests**: Testing individual functions and methods
- **Integration Tests**: Testing database operations and service interactions
- **Coverage Reporting**: Ensuring adequate test coverage
- **Test Utilities**: Common helpers for test setup and assertions

## Test Structure

```
kisanlink-ecom/
├── internal/
│   ├── testutils/           # Test utilities and helpers
│   │   └── testutils.go     # Common test functions
│   ├── database/
│   │   └── manager_integration_test.go  # Database manager tests
│   ├── repositories/
│   │   └── user_repository_integration_test.go  # Repository tests
│   └── handlers/
│       └── health_test.go   # Handler tests
├── scripts/
│   ├── test-coverage.sh     # Coverage analysis script
│   └── test-config.sh       # Test configuration script
└── docs/
    └── TESTING.md           # This file
```

## Test Utilities

### TestUtils Package

The `internal/testutils` package provides common testing utilities:

```go
// Test configuration
type TestConfig struct {
    DatabaseProvider string
    DatabaseHost     string
    DatabasePort     int
    DatabaseUser     string
    DatabasePassword string
    DatabaseName     string
    LogLevel         string
}

// Test database manager
func TestDatabaseManager(t *testing.T, cfg *TestConfig) *database.Manager

// Create test user
func CreateTestUser(t *testing.T, repo *repositories.UserRepository, email string) *user.User

// Setup test router
func SetupTestRouter() *gin.Engine

// Make test HTTP request
func MakeTestRequest(t *testing.T, router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder

// Assert JSON response
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedSuccess bool)
```

### Usage Example

```go
func TestUserHandler(t *testing.T) {
    // Setup test suite
    suite := testutils.NewTestSuite(t)
    defer suite.Teardown()

    // Create test user
    userRepo := suite.Manager.GetUserRepository()
    testUser := testutils.CreateTestUser(t, userRepo, "test@example.com")

    // Make HTTP request
    w := testutils.MakeTestRequest(t, suite.Router, "GET", "/users/"+testUser.ID, nil)

    // Assert response
    testutils.AssertJSONResponse(t, w, http.StatusOK, true)
}
```

## Running Tests

### Basic Test Commands

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests with detailed coverage analysis
make test-coverage-analysis

# Setup test environment
make test-setup

# Validate test environment
make test-validate

# Show test configuration
make test-config
```

### Go Test Runner (Script-Free)

```bash
# Run all tests using Go test runner
make test-runner

# Run tests with coverage using Go test runner
make test-runner-coverage

# Validate test environment using Go test runner
make test-runner-validate

# Show test summary using Go test runner
make test-runner-summary

# Direct usage
go run cmd/testrunner/main.go
go run cmd/testrunner/main.go -coverage
go run cmd/testrunner/main.go -run TestHealthCheck
go run cmd/testrunner/main.go -package ./internal/handlers
```

### Direct Go Commands

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -v -run TestHealthCheck ./internal/handlers

# Run tests with race detection
go test -race ./...

# Run tests with timeout
go test -timeout 30s ./...
```

### Coverage Analysis

```bash
# Run detailed coverage analysis using Go test runner
go run cmd/testrunner/main.go -coverage

# Analyze existing coverage report
go run cmd/testrunner/main.go -analyze

# Run with custom threshold
go run cmd/testrunner/main.go -coverage -threshold=80

# Clean up coverage files
go run cmd/testrunner/main.go -cleanup
```

## Coverage Reporting

### Coverage Reports

The coverage system generates multiple reports:

- **coverage.out**: Raw coverage data
- **coverage.html**: HTML coverage report (viewable in browser)
- **coverage-func.txt**: Function-level coverage report

### Coverage Thresholds

- **Default**: 70% coverage required
- **Strict**: 80% coverage for critical packages
- **Custom**: Configurable via script parameters

### Coverage Analysis

The coverage script provides:

1. **Dependency Check**: Validates required tools
2. **Test Execution**: Runs all tests with coverage
3. **Report Generation**: Creates HTML and text reports
4. **Threshold Validation**: Ensures coverage meets requirements
5. **Package Analysis**: Identifies packages with low coverage

### Viewing Coverage Reports

```bash
# Open HTML report in browser
open coverage.html

# View function coverage
cat coverage-func.txt

# View coverage by package
go tool cover -func=coverage.out
```

## Integration Tests

### Database Manager Tests

Tests the database manager functionality:

```go
func TestDatabaseManagerIntegration(t *testing.T) {
    // Test in-memory provider
    t.Run("InMemoryProvider", func(t *testing.T) {
        // Test connection, repositories, health checks
    })

    // Test invalid provider fallback
    t.Run("InvalidProvider", func(t *testing.T) {
        // Test fallback to in-memory
    })
}
```

### Repository Tests

Tests repository CRUD operations:

```go
func TestUserRepositoryIntegration(t *testing.T) {
    t.Run("CreateUser", func(t *testing.T) {
        // Test user creation and validation
    })

    t.Run("GetByID", func(t *testing.T) {
        // Test user retrieval
    })

    t.Run("UpdateUser", func(t *testing.T) {
        // Test user updates
    })

    t.Run("DeleteUser", func(t *testing.T) {
        // Test user deletion
    })
}
```

### Test Coverage Areas

1. **Database Operations**
   - Connection management
   - CRUD operations
   - Transaction handling
   - Error scenarios

2. **Repository Operations**
   - Create, Read, Update, Delete
   - Soft delete and restore
   - Bulk operations
   - Filtering and searching

3. **Handler Operations**
   - HTTP request handling
   - Response formatting
   - Error handling
   - Authentication

4. **Service Operations**
   - Business logic
   - External service integration
   - Error handling

## Test Configuration

### Environment Variables

```bash
# Test environment
export TEST_ENV="test"
export TEST_DATABASE_PROVIDER="inmemory"
export TEST_LOG_LEVEL="debug"
export TEST_TIMEOUT="30s"
```

### Database Providers

- **inmemory**: In-memory storage (default for tests)
- **postgres**: PostgreSQL database
- **dynamodb**: DynamoDB database

### Test Configuration

```bash
# Setup test environment
make test-setup

# Validate environment
make test-validate

# Show configuration
make test-config

# Validate using Go test runner
go run cmd/testrunner/main.go -validate

# Show test summary
go run cmd/testrunner/main.go -summary
```

## Best Practices

### Test Organization

1. **Test Files**: Use `_test.go` suffix
2. **Test Functions**: Use descriptive names
3. **Test Groups**: Use `t.Run()` for sub-tests
4. **Test Data**: Use test utilities for consistent data

### Test Naming

```go
// Good test names
func TestUserRepository_CreateUser_Success(t *testing.T)
func TestUserRepository_CreateUser_DuplicateEmail(t *testing.T)
func TestDatabaseManager_Connect_ValidConfig(t *testing.T)

// Avoid generic names
func TestCreate(t *testing.T)  // Too generic
func TestUser(t *testing.T)    // Too generic
```

### Test Structure

```go
func TestExample(t *testing.T) {
    // 1. Setup
    repo := NewUserRepository()
    ctx := context.Background()

    // 2. Execute
    user := &user.User{...}
    err := repo.Create(ctx, user)

    // 3. Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, user.ID)
}
```

### Test Data Management

1. **Clean State**: Each test should start with clean data
2. **Unique Data**: Use unique identifiers for test data
3. **Cleanup**: Clean up test data after tests
4. **Isolation**: Tests should not depend on each other

### Error Testing

```go
// Test expected errors
err := repo.Create(ctx, invalidUser)
assert.Error(t, err)
assert.Contains(t, err.Error(), "validation failed")

// Test unexpected errors
_, err := repo.GetByID(ctx, "non-existent")
assert.Error(t, err)
```

### Performance Testing

```go
// Benchmark tests
func BenchmarkUserRepository_Create(b *testing.B) {
    repo := NewUserRepository()
    ctx := context.Background()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        user := &user.User{...}
        repo.Create(ctx, user)
    }
}
```

## Troubleshooting

### Common Issues

1. **Import Cycles**: Avoid circular imports in test files
2. **Test Dependencies**: Ensure all test dependencies are available
3. **Database Connections**: Use in-memory database for unit tests
4. **Test Isolation**: Ensure tests don't interfere with each other

### Debugging Tests

```bash
# Run tests with verbose output
go test -v -run TestSpecificTest ./path/to/package

# Run tests with debug logging
go test -v -log.level=debug ./...

# Run tests with timeout
go test -timeout 60s ./...

# Run tests with race detection
go test -race ./...
```

### Coverage Issues

1. **Low Coverage**: Add more test cases
2. **Missing Branches**: Test error conditions
3. **Unreachable Code**: Remove dead code
4. **Integration Gaps**: Test service interactions

## Continuous Integration

### CI/CD Integration

The test suite is designed to work with CI/CD pipelines:

```yaml
# Example GitHub Actions
- name: Run Tests
  run: make test

- name: Run Coverage
  run: make test-coverage

- name: Validate Coverage
  run: ./scripts/test-coverage.sh --threshold=80
```

### Pre-commit Hooks

Tests are automatically run in pre-commit hooks:

```bash
# Pre-commit script includes
make test
make test-coverage
```

## Conclusion

This testing framework provides comprehensive coverage of the application with:

- ✅ **Unit Tests**: Individual component testing
- ✅ **Integration Tests**: Database and service testing
- ✅ **Coverage Reporting**: Detailed coverage analysis
- ✅ **Test Utilities**: Common testing helpers
- ✅ **CI/CD Integration**: Automated testing pipeline

For questions or issues with testing, please refer to the project documentation or create an issue in the repository.
