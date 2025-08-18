# KisanLink E-commerce Service Makefile

# Variables
BINARY_NAME=kisanlink-ecom
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_WINDOWS=$(BINARY_NAME)_windows.exe
BINARY_DARWIN=$(BINARY_NAME)_darwin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GORUN=$(GOCMD) run
GOGENERATE=$(GOCMD) generate

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(shell git describe --tags --always --dirty) -X main.BuildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S')"

# Directories
CMD_DIR=cmd
INTERNAL_DIR=internal
PKG_DIR=pkg
DOCS_DIR=docs
COVERAGE_DIR=coverage

# Main targets
.PHONY: all build clean test coverage lint swagger run dev docker-build docker-run help

# Default target
all: clean build test

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./$(CMD_DIR)/server
	@echo "Build complete!"

# Build for multiple platforms
build-all: clean
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_UNIX) ./$(CMD_DIR)/server
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_WINDOWS) ./$(CMD_DIR)/server
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DARWIN) ./$(CMD_DIR)/server
	@echo "Multi-platform build complete!"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME) $(BINARY_UNIX) $(BINARY_WINDOWS) $(BINARY_DARWIN)
	rm -rf $(COVERAGE_DIR)
	@echo "Clean complete!"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...
	@echo "Tests complete!"

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report generated in $(COVERAGE_DIR)/coverage.html"

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	$(GOTEST) -race -v ./...

# Run specific test
test-package:
	@echo "Running tests for package: $(PACKAGE)"
	$(GOTEST) -v ./$(PACKAGE)

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Lint the code
lint:
	@echo "Linting code..."
	golangci-lint run
	@echo "Lint complete!"

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...
	@echo "Format complete!"

# Vet the code
vet:
	@echo "Vetting code..."
	$(GOCMD) vet ./...
	@echo "Vet complete!"

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	swag init -g $(CMD_DIR)/server/main.go -o $(DOCS_DIR)
	@echo "Swagger documentation generated!"

# Install Swagger tools
install-swagger:
	@echo "Installing Swagger tools..."
	$(GOGET) -u github.com/swaggo/swag/cmd/swag
	$(GOGET) -u github.com/swaggo/gin-swagger
	@echo "Swagger tools installed!"

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

# Run the application directly (without building)
dev:
	@echo "Running in development mode..."
	$(GORUN) $(CMD_DIR)/server/main.go

# Run with hot reload (requires air)
dev-hot:
	@echo "Running with hot reload..."
	air

# Install air for hot reload
install-air:
	@echo "Installing air for hot reload..."
	$(GOGET) -u github.com/cosmtrek/air
	@echo "Air installed!"

# Database operations
db-migrate:
	@echo "Running database migrations..."
	$(GORUN) $(CMD_DIR)/migrate/main.go

db-seed:
	@echo "Seeding database..."
	$(GORUN) $(CMD_DIR)/seed/main.go

# Docker operations
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME) .
	@echo "Docker build complete!"

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(BINARY_NAME)

docker-stop:
	@echo "Stopping Docker container..."
	docker stop $(shell docker ps -q --filter ancestor=$(BINARY_NAME))

# Dependency management
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	@echo "Dependencies downloaded!"

deps-tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy
	@echo "Dependencies tidied!"

deps-update:
	@echo "Updating dependencies..."
	$(GOMOD) get -u ./...
	$(GOMOD) tidy
	@echo "Dependencies updated!"

# Security
security-check:
	@echo "Running security checks..."
	gosec ./...
	@echo "Security check complete!"

# Install security tools
install-security-tools:
	@echo "Installing security tools..."
	$(GOGET) -u github.com/securecodewarrior/gosec/v2/cmd/gosec
	@echo "Security tools installed!"

# Performance profiling
profile:
	@echo "Running with profiling..."
	$(GORUN) -cpuprofile=cpu.prof -memprofile=mem.prof $(CMD_DIR)/server/main.go

# Analyze profiles
analyze-profile:
	@echo "Analyzing profiles..."
	$(GOCMD) tool pprof cpu.prof
	$(GOCMD) tool pprof mem.prof

# Generate mocks
mocks:
	@echo "Generating mocks..."
	mockgen -source=internal/services/catalog/iface.go -destination=internal/mocks/catalog_mocks.go
	mockgen -source=internal/services/orders/iface.go -destination=internal/mocks/orders_mocks.go
	@echo "Mocks generated!"

# Install mockgen
install-mockgen:
	@echo "Installing mockgen..."
	$(GOGET) -u github.com/golang/mock/mockgen
	@echo "Mockgen installed!"

# Code generation
generate:
	@echo "Generating code..."
	$(GOGENERATE) ./...
	@echo "Code generation complete!"

# Install all development tools
install-tools: install-swagger install-air install-security-tools install-mockgen
	@echo "All development tools installed!"

# Show help
help:
	@echo "Available targets:"
	@echo "  build              - Build the application"
	@echo "  build-all          - Build for multiple platforms"
	@echo "  clean              - Clean build artifacts"
	@echo "  test               - Run tests"
	@echo "  test-coverage      - Run tests with coverage"
	@echo "  test-race          - Run tests with race detection"
	@echo "  lint               - Lint the code"
	@echo "  fmt                - Format code"
	@echo "  vet                - Vet the code"
	@echo "  swagger            - Generate Swagger documentation"
	@echo "  run                - Build and run the application"
	@echo "  dev                - Run in development mode"
	@echo "  dev-hot            - Run with hot reload"
	@echo "  db-migrate         - Run database migrations"
	@echo "  db-seed            - Seed the database"
	@echo "  docker-build       - Build Docker image"
	@echo "  docker-run         - Run Docker container"
	@echo "  deps               - Download dependencies"
	@echo "  deps-tidy          - Tidy dependencies"
	@echo "  deps-update        - Update dependencies"
	@echo "  security-check     - Run security checks"
	@echo "  profile            - Run with profiling"
	@echo "  mocks              - Generate mocks"
	@echo "  generate           - Generate code"
	@echo "  install-tools      - Install all development tools"
	@echo "  help               - Show this help message"

# Development workflow
dev-setup: install-tools deps-tidy swagger
	@echo "Development environment setup complete!"

# CI/CD pipeline
ci: deps-tidy lint test test-coverage security-check
	@echo "CI pipeline complete!"

# Pre-commit hooks
pre-commit: fmt lint test
	@echo "Pre-commit checks complete!"

# Release preparation
release: clean build-all test-coverage
	@echo "Release preparation complete!" 