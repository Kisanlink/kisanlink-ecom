#!/bin/bash

# Fast Pre-commit hook script for Go projects
# Runs essential quality checks only, skips slow tests

set -e

echo "🔍 Running pre-commit checks (fast mode)..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
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

# Check if this is an initial commit
if git rev-parse --verify HEAD >/dev/null 2>&1
then
    against=HEAD
else
    # Initial commit: diff against an empty tree object
    against=$(git hash-object -t tree /dev/null)
fi

# Get list of Go files that are being committed
GO_FILES=$(git diff --cached --name-only --diff-filter=ACM $against | grep '\.go$' || true)

if [ -z "$GO_FILES" ]; then
    print_warning "No Go files to check"
    exit 0
fi

print_status "Found Go files to check: $(echo $GO_FILES | wc -w) files"

# Check if required tools are installed
check_tool() {
    if ! command -v $1 &> /dev/null; then
        print_error "$1 is not installed. Please install it first."
        echo "  go install $2"
        exit 1
    fi
}

print_status "Checking required tools..."
check_tool "gofmt" ""
check_tool "goimports" "golang.org/x/tools/cmd/goimports@latest"
check_tool "golangci-lint" "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

# Run go mod tidy
print_status "Running go mod tidy..."
if ! go mod tidy; then
    print_error "go mod tidy failed"
    exit 1
fi
print_success "go mod tidy passed"

# Check if go.mod or go.sum changed
if git diff --name-only | grep -E "go\.(mod|sum)" > /dev/null; then
    print_warning "go.mod or go.sum was modified by 'go mod tidy'"
    print_warning "Please add the changes and commit again"
    exit 1
fi

# Run gofmt
print_status "Running gofmt..."
UNFORMATTED=$(gofmt -l $GO_FILES)
if [ -n "$UNFORMATTED" ]; then
    print_error "gofmt found unformatted files:"
    for file in $UNFORMATTED; do
        echo "  $file"
    done
    print_error "Please run 'make format' or 'gofmt -w .' to fix formatting"
    exit 1
fi
print_success "gofmt passed"

# Run goimports
print_status "Running goimports..."
UNIMPORTED=$(goimports -l $GO_FILES)
if [ -n "$UNIMPORTED" ]; then
    print_error "goimports found files with missing/incorrect imports:"
    for file in $UNIMPORTED; do
        echo "  $file"
    done
    print_error "Please run 'make format' or 'goimports -w .' to fix imports"
    exit 1
fi
print_success "goimports passed"

# Run go vet
print_status "Running go vet..."
if ! go vet ./...; then
    print_error "go vet failed"
    exit 1
fi
print_success "go vet passed"

# Run golangci-lint (only check new code)
print_status "Running golangci-lint..."
if ! golangci-lint run --new-from-rev=HEAD; then
    print_error "golangci-lint failed"
    print_warning "Fix the linting issues or run 'golangci-lint run --fix' to auto-fix some issues"
    exit 1
fi
print_success "golangci-lint passed"

# Build check (fast)
print_status "Running build check..."
if ! go build ./cmd/server > /dev/null 2>&1; then
    print_error "Build failed"
    go build ./cmd/server  # Show error
    exit 1
fi
print_success "Build check passed"

# Unit tests are SKIPPED in fast mode (run in CI)
print_warning "Unit tests skipped (run 'make test-unit' manually or tests run in CI)"

print_success "All pre-commit checks passed! 🎉"
echo ""
print_status "Commit summary:"
echo "  - Files checked: $(echo $GO_FILES | wc -w)"
echo "  - Format: ✓"
echo "  - Imports: ✓"
echo "  - Vet: ✓"
echo "  - Lint: ✓"
echo "  - Build: ✓"
echo "  - Tests: ⚠ Skipped (run in CI)"
echo ""
print_warning "Run 'make test-unit' before pushing to catch test failures early"

exit 0
