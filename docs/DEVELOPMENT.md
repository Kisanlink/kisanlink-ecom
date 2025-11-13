# Development Guide

This document outlines the development workflow, code quality tools, and best practices for the KisanLink E-commerce API.

## 🛠️ Development Tools Setup

The project includes comprehensive code quality tools that run automatically on every commit:

### Pre-commit Hooks

Pre-commit hooks ensure code quality before changes are committed:

```bash
# Install hooks and tools
make setup-hooks

# This installs:
# - Git pre-commit hook
# - golangci-lint (Go linter)
# - goimports (import formatter)
# - gosec (security scanner)
# - gofumpt (code formatter)
```

### Available Commands

| Command              | Description                                              |
| -------------------- | -------------------------------------------------------- |
| `make help`          | Show all available commands                              |
| `make format`        | Format code with gofmt, goimports, and gofumpt           |
| `make lint`          | Run comprehensive linting with golangci-lint             |
| `make lint-fix`      | Run linting with auto-fix enabled                        |
| `make security`      | Run security checks with gosec                           |
| `make test`          | Run tests                                                |
| `make test-coverage` | Run tests with coverage report                           |
| `make check`         | Run all quality checks (format + lint + test + security) |
| `make quick-check`   | Run essential checks quickly                             |
| `make pre-commit`    | Manually run pre-commit checks                           |
| `make install-tools` | Install/update development tools                         |

## 🔍 Code Quality Checks

### What Runs on Every Commit

The pre-commit hook automatically runs:

1. **go mod tidy** - Ensures dependencies are clean
2. **gofmt** - Code formatting
3. **goimports** - Import organization and formatting
4. **go vet** - Static analysis for common issues
5. **golangci-lint** - Comprehensive linting (40+ linters)
6. **go test** - Unit tests
7. **gosec** - Security vulnerability scanning
8. **TODO/FIXME detection** - Tracks technical debt

### Linting Configuration

The project uses `.golangci.yml` for comprehensive linting with:

- 40+ enabled linters including errcheck, govet, gosimple, staticcheck
- Security checks with gosec
- Code style enforcement
- Performance checks
- Best practices validation

For development, use `.golangci-sample.yml` for more relaxed linting:

```bash
golangci-lint run --config .golangci-sample.yml
```

## 🚀 Development Workflow

### Quick Development Loop

```bash
# 1. Make your changes
vim internal/handlers/users.go

# 2. Run quick checks
make quick-check

# 3. Fix any issues found
make format
make lint-fix

# 4. Run full checks before committing
make check

# 5. Commit (pre-commit hook runs automatically)
git add .
git commit -m "feat: add user validation"
```

### Testing the Pre-commit Hook

```bash
# Test the pre-commit hook without committing
./scripts/test-commit.sh

# Or run it manually
make pre-commit
```

## 📋 Code Standards

### Go Style Guidelines

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Use `goimports` for import organization
- Add comments for all exported functions and types
- End comments with periods (enforced by `godot` linter)
- Handle all errors explicitly

### Security Guidelines

- No hardcoded credentials
- Validate all inputs
- Use secure random number generation
- Set appropriate file permissions
- Follow OWASP guidelines

### Testing Guidelines

- Write tests for all business logic
- Use table-driven tests where appropriate
- Aim for >80% code coverage
- Include integration tests for critical flows

## 🛡️ Security Scanning

### gosec - Security Scanner

Automatically scans for:

- Hardcoded credentials (G101)
- SQL injection vulnerabilities (G201, G202)
- Command injection (G204)
- File path traversal (G304)
- Weak cryptography (G401, G402, G405)
- And more...

### Running Security Checks

```bash
# Run security scan
make security

# Run with detailed output
gosec -verbose ./...

# Generate report
gosec -fmt=json -out=security-report.json ./...
```

## 🔧 Troubleshooting

### Pre-commit Hook Too Slow

If the pre-commit hook takes too long:

```bash
# Use quick checks during development
make quick-check

# Skip hook temporarily (not recommended)
git commit --no-verify
```

### Linting Issues

```bash
# Auto-fix what can be fixed
make lint-fix

# Use relaxed config for development
golangci-lint run --config .golangci-sample.yml

# Disable specific rules (in .golangci.yml)
linters:
  disable:
    - dupl  # Disable duplicate code detection
```

### Tools Not Found

```bash
# Reinstall tools
make install-tools

# Check PATH
echo $PATH | grep $(go env GOPATH)/bin

# Manual installation
export PATH="$(go env GOPATH)/bin:$PATH"
```

## 📊 Metrics and Reporting

### Coverage Reports

```bash
# Generate coverage report
make test-coverage

# View in browser
open coverage.html
```

### Code Quality Reports

```bash
# JSON format for CI/CD
golangci-lint run --out-format=json > lint-report.json

# SARIF format for GitHub Security
golangci-lint run --out-format=sarif > lint-report.sarif
```

## 🚀 CI/CD Integration

The project includes GitHub Actions workflows in `.github/workflows/ci.yml`:

- **Lint Job**: Runs golangci-lint
- **Test Job**: Runs tests on multiple Go versions
- **Security Job**: Runs gosec and uploads SARIF results
- **Build Job**: Compiles the application
- **Code Quality Job**: Validates formatting and imports

### Pre-commit Framework (Optional)

If you have [pre-commit](https://pre-commit.com/) installed:

```bash
# Install pre-commit hooks
pre-commit install

# Run on all files
pre-commit run --all-files

# Update hooks
pre-commit autoupdate
```

## 📝 Best Practices

### Commit Messages

Use conventional commits:

```
feat: add user authentication
fix: resolve memory leak in middleware
docs: update API documentation
refactor: simplify error handling
test: add integration tests for orders
```

### Branch Strategy

- `main`: Production-ready code
- `develop`: Integration branch
- `feature/`: Feature branches
- `bugfix/`: Bug fix branches
- `hotfix/`: Emergency fixes

### Code Reviews

All changes should:

- Pass all automated checks
- Have appropriate test coverage
- Follow the style guide
- Include documentation updates
- Be reviewed by at least one team member

## 📚 Additional Resources

- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [golangci-lint Linters](https://golangci-lint.run/usage/linters/)
- [gosec Rules](https://securecodewarrior.github.io/gosec/)
- [Effective Go](https://golang.org/doc/effective_go.html)
