# Test Separation: Unit vs Integration Tests

## Overview

This document describes the separation of unit and integration tests in the kisanlink-ecom project.

## Strategy

We use Go build tags to separate unit and integration tests:

- **Unit tests**: Run without any build tags (fast, no external dependencies)
- **Integration tests**: Require `//go:build integration` tag (may require database, external services)

## Build Tags Applied

### Integration Test Files

All files with the `//go:build integration` tag:

1. **tests/integration/**
   - `aaa_service_integration_test.go`
   - `database_integration_test.go`
   - `inventory_integration_test.go`
   - `database_manager_integration_test.go`

2. **tests/database/**
   - `manager_integration_test.go`
   - `catalog_indexes_test.go`
   - `migration_test.go`
   - `marketplace_migration_test.go`

3. **tests/e2e/**
   - `api_contract_test.go`
   - `load_test.go`
   - `simple_flow_test.go`

4. **tests/handlers/**
   - `catalog/integration_test.go`
   - `integrations/integration_handler_test.go`
   - `integrations/integration_test.go`
   - `integrations/integration_handler_working_test.go`

5. **tests/repositories/**
   - `orders/order_repository_integration_test.go`

### Unit Test Files

All other test files remain as unit tests (no build tag required).

## Makefile Targets

### Unit Tests Only (Fast)

```bash
make test-unit                    # Run unit tests
make test-unit-coverage           # Run unit tests with coverage
make test-unit-coverage-check     # Run unit tests with 90% coverage enforcement
make test-unit-race              # Run unit tests with race detection
```

### Integration Tests Only

```bash
make test-integration            # Run integration tests only
```

### All Tests (Unit + Integration)

```bash
make test                        # Run all tests
make test-coverage               # Run all tests with coverage
make test-coverage-check         # Run all tests with 90% coverage enforcement
make test-race                   # Run all tests with race detection
```

## Pre-commit Hooks

The pre-commit hooks have been configured to **only run unit tests** for fast feedback:

```yaml
- repo: local
  hooks:
    - id: go-unit-tests-only
      name: go unit tests (excludes integration tests)
      entry: make test-unit
      language: system
      pass_filenames: false
      types: [go]
```

## CI/CD Pipeline

### Pull Request Pipeline (Fast)

```bash
make ci-pr    # Runs: deps-tidy, fmt, vet, lint, test-unit, test-unit-race
```

### Full CI Pipeline

```bash
make ci       # Runs: deps-tidy, fmt, vet, lint, test-coverage-check, security-check, swagger
```

## Verification

### Check Unit Tests Exclude Integration Tests

```bash
go test ./... -run NonExistentTest 2>&1 | grep "tests/integration"
# Should show: no packages to test
```

### Check Integration Tests Require Tag

```bash
go test ./tests/integration/...
# Should show: matched no packages

go test -tags=integration ./tests/integration/...
# Should run integration tests
```

## Benefits

1. **Faster Pre-commit**: Only unit tests run, providing quick feedback
2. **Clear Separation**: Easy to identify which tests require external dependencies
3. **Flexible CI/CD**: Can run different test suites in different pipeline stages
4. **Better Developer Experience**: Quick local testing with unit tests, comprehensive testing in CI
