#!/bin/bash

# Quick development check script
# This runs essential checks without being too strict

set -e

# Add Go binary path to PATH
export PATH="$(go env GOPATH)/bin:$PATH"

echo "🚀 Running quick development checks..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}==>${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Check if tools are available
if ! command -v golangci-lint &> /dev/null; then
    print_error "golangci-lint not found. Run 'make install-tools' first."
    exit 1
fi

# Run go mod tidy
print_status "Running go mod tidy..."
if go mod tidy; then
    print_success "go mod tidy passed"
else
    print_error "go mod tidy failed"
    exit 1
fi

# Format code
print_status "Formatting code..."
if gofmt -w .; then
    print_success "gofmt passed"
else
    print_error "gofmt failed"
    exit 1
fi

# Run basic linting with relaxed config
print_status "Running basic linting..."
if [ -f ".golangci-sample.yml" ]; then
    if golangci-lint run --config .golangci-sample.yml; then
        print_success "Basic linting passed"
    else
        print_warning "Linting found issues (not blocking)"
    fi
else
    # Run essential linters only
    if golangci-lint run --disable-all --enable=errcheck,gosimple,govet,ineffassign,staticcheck,typecheck,unused,gofmt,goimports; then
        print_success "Essential linting passed"
    else
        print_warning "Essential linting found issues (not blocking)"
    fi
fi

# Run tests
print_status "Running tests..."
if go test ./...; then
    print_success "Tests passed"
else
    print_warning "Tests failed or no tests found"
fi

print_success "Quick checks completed! 🎉"
print_status "For full checks, run: make check" 