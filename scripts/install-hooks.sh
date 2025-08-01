#!/bin/bash

# Script to install Git hooks and required development tools

set -e

# Add Go binary path to PATH
export PATH="$(go env GOPATH)/bin:$PATH"

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

echo "🔧 Setting up development environment for KisanLink E-commerce API"
echo ""

# Check if we're in a Git repository
if [ ! -d ".git" ]; then
    print_error "This is not a Git repository. Please run this script from the project root."
    exit 1
fi

# Install Git hooks
print_status "Installing Git hooks..."

# Create hooks directory if it doesn't exist
mkdir -p .git/hooks

# Copy pre-commit hook
if [ -f "scripts/pre-commit.sh" ]; then
    cp scripts/pre-commit.sh .git/hooks/pre-commit
    chmod +x .git/hooks/pre-commit
    print_success "Pre-commit hook installed"
else
    print_error "scripts/pre-commit.sh not found"
    exit 1
fi

# Install required Go tools
print_status "Installing required Go tools..."

install_tool() {
    local tool_name="$1"
    local install_cmd="$2"
    
    if command -v "$tool_name" &> /dev/null; then
        print_success "$tool_name is already installed"
    else
        print_status "Installing $tool_name..."
        if eval "$install_cmd"; then
            print_success "$tool_name installed successfully"
        else
            print_error "Failed to install $tool_name"
            exit 1
        fi
    fi
}

# Install development tools
install_tool "goimports" "go install golang.org/x/tools/cmd/goimports@latest"
install_tool "golangci-lint" "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
install_tool "gosec" "go install github.com/securego/gosec/v2/cmd/gosec@latest"
install_tool "gofumpt" "go install mvdan.cc/gofumpt@latest"

# Check if pre-commit framework is available
print_status "Checking pre-commit framework..."
if command -v pre-commit &> /dev/null; then
    print_success "pre-commit framework is available"
    
    # Install pre-commit hooks
    if [ -f ".pre-commit-config.yaml" ]; then
        print_status "Installing pre-commit hooks..."
        if pre-commit install; then
            print_success "Pre-commit hooks installed"
        else
            print_warning "Failed to install pre-commit hooks"
        fi
    else
        print_warning ".pre-commit-config.yaml not found"
    fi
else
    print_warning "pre-commit framework not found"
    print_warning "You can install it with: pip install pre-commit"
    print_warning "Or use Homebrew: brew install pre-commit"
fi

# Verify installations
print_status "Verifying installations..."

check_tool() {
    if command -v "$1" &> /dev/null; then
        local version=$($1 version 2>/dev/null || $1 --version 2>/dev/null || echo "unknown")
        print_success "$1: $version"
    else
        print_warning "$1: not found"
    fi
}

check_tool "go"
check_tool "gofmt"
check_tool "goimports"
check_tool "golangci-lint"
check_tool "gosec"
check_tool "gofumpt"

# Create .gitignore entries for development files
print_status "Updating .gitignore..."
cat >> .gitignore << 'EOF'

# Development tools
*.log
.DS_Store
.vscode/
.idea/
*.swp
*.swo
*~

# Build artifacts
bin/
dist/
*.exe
*.exe~

# Coverage files
coverage.out
coverage.html

# Environment files
.env.local
.env.development
.env.test
.env.production
EOF

print_success ".gitignore updated"

echo ""
print_success "Development environment setup complete! 🎉"
echo ""
print_status "What was installed:"
echo "  - Git pre-commit hook"
echo "  - golangci-lint (Go linter)"
echo "  - goimports (import formatter)"
echo "  - gosec (security checker)"
echo "  - gofumpt (stricter formatter)"
if command -v pre-commit &> /dev/null; then
    echo "  - pre-commit hooks framework"
fi
echo ""
print_status "Usage:"
echo "  - Run 'make lint' to run linters manually"
echo "  - Run 'make format' to format code"
echo "  - Run 'make check' to run all quality checks"
echo "  - Git hooks will run automatically on commit"
echo ""
print_warning "Note: The pre-commit hook will run tests and linting before each commit."
print_warning "This may take a few seconds but ensures code quality." 