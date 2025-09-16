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
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD)
GIT_BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

# Environment-specific build flags
LDFLAGS_BASE=-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.GitBranch=$(GIT_BRANCH)
LDFLAGS_DEV=-ldflags "$(LDFLAGS_BASE)"
LDFLAGS_PROD=-ldflags "-s -w $(LDFLAGS_BASE)"
LDFLAGS=$(LDFLAGS_DEV)

# Directories
CMD_DIR=cmd
INTERNAL_DIR=internal
PKG_DIR=pkg
DOCS_DIR=docs
COVERAGE_DIR=coverage

# Main targets
.PHONY: all build clean test coverage lint swagger run dev docker-build docker-run help test-aaa seed-aaa test-endpoints

# Default target
all: clean build test

# Build the application
build:
	@echo "Building $(BINARY_NAME) for development..."
	$(GOBUILD) $(LDFLAGS_DEV) -o $(BINARY_NAME) ./$(CMD_DIR)/server
	@echo "Build complete!"

# Build for production (optimized)
build-prod:
	@echo "Building $(BINARY_NAME) for production..."
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS_PROD) -a -installsuffix cgo -o $(BINARY_NAME) ./$(CMD_DIR)/server
	@echo "Production build complete!"

# Build for staging
build-staging:
	@echo "Building $(BINARY_NAME) for staging..."
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS_DEV) -a -installsuffix cgo -o $(BINARY_NAME) ./$(CMD_DIR)/server
	@echo "Staging build complete!"

# Build for multiple platforms
build-all: clean
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS_PROD) -a -installsuffix cgo -o $(BINARY_UNIX) ./$(CMD_DIR)/server
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS_PROD) -a -installsuffix cgo -o $(BINARY_WINDOWS) ./$(CMD_DIR)/server
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS_PROD) -a -installsuffix cgo -o $(BINARY_DARWIN) ./$(CMD_DIR)/server
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS_PROD) -a -installsuffix cgo -o $(BINARY_NAME)_linux_arm64 ./$(CMD_DIR)/server
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS_PROD) -a -installsuffix cgo -o $(BINARY_NAME)_darwin_arm64 ./$(CMD_DIR)/server
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
	$(GOTEST) -v -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	$(GOCMD) tool cover -func=$(COVERAGE_DIR)/coverage.out | grep total | awk '{print "Total coverage: " $$3}'
	@echo "Coverage report generated in $(COVERAGE_DIR)/coverage.html"

# Run tests with coverage and enforce minimum coverage
test-coverage-check:
	@echo "Running tests with coverage check (minimum 90%)..."
	mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./...
	@COVERAGE=$$($(GOCMD) tool cover -func=$(COVERAGE_DIR)/coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Total coverage: $$COVERAGE%"; \
	if [ $$(echo "$$COVERAGE < 90" | bc -l) -eq 1 ]; then \
	    echo "Coverage $$COVERAGE% is below minimum 90%"; \
	    exit 1; \
	fi

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

# Comprehensive format command aligned with pre-commit hooks
format: fmt
	@echo "Running comprehensive formatting..."
	@echo "  - Running go fmt..."
	$(GOCMD) fmt ./...
	@echo "  - Running goimports..."
	goimports -w .
	@echo "  - Running go mod tidy..."
	$(GOMOD) tidy
	@echo "Comprehensive format complete!"

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

docker-build-dev:
	@echo "Building development Docker image..."
	docker build --target development -t $(BINARY_NAME):dev .
	@echo "Development Docker build complete!"

docker-build-prod:
	@echo "Building production Docker image..."
	docker build --target runtime -t $(BINARY_NAME):prod .
	@echo "Production Docker build complete!"

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(BINARY_NAME)

docker-run-dev:
	@echo "Running development Docker container..."
	docker run -p 8080:8080 -v $(PWD):/app $(BINARY_NAME):dev

docker-stop:
	@echo "Stopping Docker container..."
	docker stop $(shell docker ps -q --filter ancestor=$(BINARY_NAME)) || true

# Docker Compose operations
compose-dev-up:
	@echo "Starting development environment with Docker Compose..."
	docker-compose --profile dev up -d
	@echo "Development environment started!"

compose-dev-down:
	@echo "Stopping development environment..."
	docker-compose --profile dev down
	@echo "Development environment stopped!"

compose-prod-up:
	@echo "Starting production environment with Docker Compose..."
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml --profile prod up -d
	@echo "Production environment started!"

compose-prod-down:
	@echo "Stopping production environment..."
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml --profile prod down
	@echo "Production environment stopped!"

compose-test:
	@echo "Running tests with Docker Compose..."
	docker-compose -f docker-compose.yml -f docker-compose.test.yml --profile test up --abort-on-container-exit
	docker-compose -f docker-compose.yml -f docker-compose.test.yml --profile test down
	@echo "Tests completed!"

compose-logs:
	@echo "Showing Docker Compose logs..."
	docker-compose logs -f

compose-status:
	@echo "Docker Compose service status:"
	docker-compose ps

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
	@echo ""
	@echo "Build targets:"
	@echo "  build              - Build the application for development"
	@echo "  build-prod         - Build optimized production binary"
	@echo "  build-staging      - Build staging binary"
	@echo "  build-all          - Build for multiple platforms"
	@echo "  clean              - Clean build artifacts"
	@echo ""
	@echo "Test targets:"
	@echo "  test               - Run tests"
	@echo "  test-coverage      - Run tests with coverage report"
	@echo "  test-coverage-check - Run tests with coverage enforcement (90%)"
	@echo "  test-race          - Run tests with race detection"
	@echo "  test-package       - Run tests for specific package (PACKAGE=path)"
	@echo "  bench              - Run benchmarks"
	@echo ""
	@echo "Code quality targets:"
	@echo "  lint               - Lint the code with golangci-lint"
	@echo "  fmt                - Format code with go fmt"
	@echo "  vet                - Vet the code with go vet"
	@echo "  security-check     - Run security checks with gosec"
	@echo "  pre-commit         - Run all pre-commit checks"
	@echo "  pre-commit-all     - Run pre-commit on all files"
	@echo ""
	@echo "Development targets:"
	@echo "  run                - Build and run the application"
	@echo "  dev                - Run in development mode"
	@echo "  dev-hot            - Run with hot reload (requires air)"
	@echo "  dev-setup          - Full development environment setup"
	@echo "  dev-setup-quick    - Quick development setup"
	@echo "  dev-setup-full     - Full setup with Docker services"
	@echo ""
	@echo "Docker targets:"
	@echo "  docker-build       - Build Docker image"
	@echo "  docker-build-dev   - Build development Docker image"
	@echo "  docker-build-prod  - Build production Docker image"
	@echo "  docker-run         - Run Docker container"
	@echo "  docker-run-dev     - Run development Docker container"
	@echo "  docker-stop        - Stop Docker container"
	@echo "  docker-dev-up      - Start development services (postgres, redis)"
	@echo "  docker-dev-down    - Stop development services"
	@echo "  docker-dev-reset   - Reset development environment"
	@echo ""
	@echo "Docker Compose targets:"
	@echo "  compose-dev-up     - Start development environment"
	@echo "  compose-dev-down   - Stop development environment"
	@echo "  compose-prod-up    - Start production environment"
	@echo "  compose-prod-down  - Stop production environment"
	@echo "  compose-test       - Run tests in containers"
	@echo "  compose-logs       - Show service logs"
	@echo "  compose-status     - Show service status"
	@echo ""
	@echo "Database targets:"
	@echo "  db-migrate         - Run database migrations"
	@echo "  db-seed            - Seed the database"
	@echo ""
	@echo "Documentation targets:"
	@echo "  swagger            - Generate Swagger documentation"
	@echo ""
	@echo "Dependency targets:"
	@echo "  deps               - Download dependencies"
	@echo "  deps-tidy          - Tidy dependencies"
	@echo "  deps-update        - Update dependencies"
	@echo ""
	@echo "CI/CD targets:"
	@echo "  ci                 - Full CI pipeline"
	@echo "  ci-pr              - PR CI pipeline"
	@echo "  cd-dev             - Development deployment pipeline"
	@echo "  cd-staging         - Staging deployment pipeline"
	@echo "  cd-prod            - Production deployment pipeline"
	@echo ""
	@echo "Utility targets:"
	@echo "  profile            - Run with profiling"
	@echo "  analyze-profile    - Analyze CPU and memory profiles"
	@echo "  mocks              - Generate mocks"
	@echo "  generate           - Generate code"
	@echo "  install-tools      - Install all development tools"
	@echo "  install-pre-commit - Install pre-commit hooks"
	@echo ""
	@echo "AAA Service targets:"
	@echo "  test-aaa           - Test AAA service connection"
	@echo "  seed-aaa           - Seed AAA service with RBAC data"
	@echo "  test-endpoints     - Test e-commerce endpoints"
	@echo "  test-full-aaa      - Run full AAA integration test"
	@echo ""
	@echo "  help               - Show this help message"

# Development workflow
dev-setup: install-tools deps-tidy install-pre-commit swagger
	@echo "Development environment setup complete!"

# Quick development setup (minimal)
dev-setup-quick: deps-tidy swagger
	@echo "Quick development setup complete!"

# Full development environment with Docker
dev-setup-full: dev-setup docker-dev-up
	@echo "Full development environment setup complete!"

# Start development environment with Docker
docker-dev-up:
	@echo "Starting development environment..."
	docker-compose up -d postgres redis
	@echo "Development environment started!"

# Stop development environment
docker-dev-down:
	@echo "Stopping development environment..."
	docker-compose down
	@echo "Development environment stopped!"

# Reset development environment
docker-dev-reset: docker-dev-down
	@echo "Resetting development environment..."
	docker-compose down -v
	docker-compose up -d postgres redis
	@echo "Development environment reset complete!"

# CI/CD pipeline
ci: deps-tidy fmt vet lint test-coverage-check security-check swagger
	@echo "CI pipeline complete!"

# CI pipeline for pull requests
ci-pr: deps-tidy fmt vet lint test test-race
	@echo "PR CI pipeline complete!"

# CD pipeline for deployment
cd-dev: build-staging swagger
	@echo "Development deployment pipeline complete!"

cd-staging: build-staging test-coverage-check swagger
	@echo "Staging deployment pipeline complete!"

cd-prod: build-prod test-coverage-check security-check swagger
	@echo "Production deployment pipeline complete!"

# Pre-commit hooks
pre-commit: fmt vet lint test-coverage-check security-check
	@echo "Pre-commit checks complete!"

# Install pre-commit hooks
install-pre-commit:
	@echo "Installing pre-commit hooks..."
	pre-commit install
	@echo "Pre-commit hooks installed!"

# Run pre-commit on all files
pre-commit-all:
	@echo "Running pre-commit on all files..."
	pre-commit run --all-files

# AAA Service Testing
test-aaa:
	@echo "Testing AAA service connection..."
	$(GORUN) debug/test_aaa_connection.go

seed-aaa:
	@echo "Seeding AAA service with e-commerce RBAC data..."
	$(GORUN) debug/seed_ecommerce_rbac.go

test-endpoints:
	@echo "Testing e-commerce endpoints..."
	$(GORUN) debug/test_ecommerce_endpoints.go

# Full AAA integration test
test-full-aaa: seed-aaa test-aaa test-endpoints
	@echo "Full AAA integration test complete!"

# Release preparation
release: clean build-all test-coverage
	@echo "Release preparation complete!"
