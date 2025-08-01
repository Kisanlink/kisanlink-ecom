#!/bin/bash

# Test script to simulate a commit and test the pre-commit hook

set -e

# Add Go binary path to PATH
export PATH="$(go env GOPATH)/bin:$PATH"

echo "🧪 Testing pre-commit hook functionality..."

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

# Check if we're in a Git repository
if [ ! -d ".git" ]; then
    print_error "This is not a Git repository."
    exit 1
fi

# Stage some Go files to test the hook
print_status "Staging Go files for testing..."
if git add cmd/ internal/ 2>/dev/null; then
    print_success "Files staged successfully"
else
    print_warning "No files to stage or already staged"
fi

# Check what's staged
STAGED_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)
if [ -z "$STAGED_FILES" ]; then
    print_warning "No Go files staged. Staging all Go files for testing..."
    git add **/*.go 2>/dev/null || true
    STAGED_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)
fi

if [ -n "$STAGED_FILES" ]; then
    print_status "Staged Go files:"
    echo "$STAGED_FILES" | sed 's/^/  - /'
    echo ""
    
    # Run the pre-commit hook directly
    print_status "Running pre-commit hook..."
    if ./scripts/pre-commit.sh; then
        print_success "Pre-commit hook passed! ✨"
    else
        print_error "Pre-commit hook failed!"
        print_status "This is expected if there are linting issues to fix."
    fi
else
    print_warning "No Go files found to test"
    print_status "Running pre-commit hook anyway..."
    ./scripts/pre-commit.sh
fi

# Unstage files
print_status "Cleaning up - unstaging files..."
git reset HEAD . 2>/dev/null || true
print_success "Test completed!"

echo ""
print_status "To test a real commit:"
echo "  1. Make some changes to Go files"
echo "  2. git add ."
echo "  3. git commit -m 'test commit'"
echo "  4. The pre-commit hook will run automatically" 