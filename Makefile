.PHONY: help build run test clean dev deps lint format check security install-tools setup-hooks pre-commit

# Default target
help:
	@echo "Available commands:"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  dev           - Run in development mode with hot reload"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Download dependencies"
	@echo "  lint          - Run linter"
	@echo "  lint-fix      - Run linter with auto-fix"
	@echo "  format        - Format code"
	@echo "  check         - Run all quality checks"
	@echo "  security      - Run security checks"
	@echo "  install-tools - Install required development tools"
	@echo "  setup-hooks   - Setup Git hooks"
	@echo "  pre-commit    - Run pre-commit checks manually"
	@echo "  quick-check   - Run quick development checks"

# Build the application
build:
	@echo "Building application..."
	go build -ldflags="-s -w" -o bin/server cmd/server/main.go

# Run the application
run: build
	@echo "Running application..."
	./bin/server

# Run in development mode
dev:
	@echo "Running in development mode..."
	go run cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v -race ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out > coverage-func.txt
	@echo "Coverage reports generated:"
	@echo "  - coverage.out (raw data)"
	@echo "  - coverage.html (HTML report)"
	@echo "  - coverage-func.txt (function report)"

# Run tests with coverage analysis
test-coverage-analysis:
	@echo "Running tests with detailed coverage analysis..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out > coverage-func.txt
	@echo "📊 Coverage Summary:"
	@go tool cover -func=coverage.out | grep total
	@echo "📁 Reports generated: coverage.out, coverage.html, coverage-func.txt"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	go mod verify

# Run linter
lint:
	@echo "Running linter..."
	@PATH="$(shell go env GOPATH)/bin:$$PATH" golangci-lint run

# Run linter with auto-fix
lint-fix:
	@echo "Running linter with auto-fix..."
	@PATH="$(shell go env GOPATH)/bin:$$PATH" golangci-lint run --fix

# Format code
format:
	@echo "Formatting code..."
	gofmt -w .
	@PATH="$(shell go env GOPATH)/bin:$$PATH" goimports -w .
	@PATH="$(shell go env GOPATH)/bin:$$PATH" gofumpt -w .

# Run all quality checks
check: format lint test security
	@echo "✅ All quality checks passed!"

# Run security checks
security:
	@echo "Running security checks..."
	@PATH="$(shell go env GOPATH)/bin:$$PATH" gosec -quiet ./...

# Install required development tools
install-tools:
	@echo "Installing development tools..."
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install mvdan.cc/gofumpt@latest
	@echo "✅ Development tools installed!"

# Setup Git hooks
setup-hooks:
	@echo "Setting up Git hooks..."
	chmod +x scripts/install-hooks.sh
	./scripts/install-hooks.sh

# Run pre-commit checks manually
pre-commit:
	@echo "Running pre-commit checks..."
	chmod +x scripts/pre-commit.sh
	@PATH="$(shell go env GOPATH)/bin:$$PATH" ./scripts/pre-commit.sh

# Run quick development checks
quick-check:
	@echo "Running quick development checks..."
	chmod +x scripts/quick-check.sh
	@PATH="$(shell go env GOPATH)/bin:$$PATH" ./scripts/quick-check.sh

# Setup test environment
test-setup:
	@echo "Setting up test environment..."
	go mod download
	go mod tidy
	go mod verify
	@echo "✅ Test environment ready"

# Validate test environment
test-validate:
	@echo "Validating test environment..."
	go build ./...
	go test -v -run TestHealthCheck ./internal/handlers
	@echo "✅ Test environment validated"

# Show test configuration
test-config:
	@echo "Test Configuration:"
	@echo "  - Database Provider: inmemory (for tests)"
	@echo "  - Log Level: debug"
	@echo "  - Test Timeout: 30s"
	@echo "  - Coverage Threshold: 70%"
	@echo ""
	@echo "Test Commands:"
	@echo "  - make test (Run all tests)"
	@echo "  - make test-coverage (Run tests with coverage)"
	@echo "  - make test-coverage-analysis (Detailed coverage)"
	@echo "  - make test-runner (Use Go test runner)"

# Run tests using Go test runner
test-runner:
	@echo "Running tests using Go test runner..."
	go run cmd/testrunner/main.go

# Run tests with coverage using Go test runner
test-runner-coverage:
	@echo "Running tests with coverage using Go test runner..."
	go run cmd/testrunner/main.go -coverage

# Validate test environment using Go test runner
test-runner-validate:
	@echo "Validating test environment using Go test runner..."
	go run cmd/testrunner/main.go -validate

# Show test summary using Go test runner
test-runner-summary:
	@echo "Showing test summary using Go test runner..."
	go run cmd/testrunner/main.go -summary 