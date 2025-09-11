# Build System Documentation

## Overview

The KisanLink E-commerce Service uses a comprehensive build system based on Make, providing targets for development, testing, building, and deployment across different environments.

## Prerequisites

### Required Tools

- **Go 1.24+**: Primary programming language
- **Make**: Build automation tool
- **Git**: Version control (for build metadata)
- **Docker & Docker Compose**: Containerization and local development

### Optional Development Tools

- **golangci-lint**: Code linting
- **air**: Hot reload for development
- **swag**: Swagger documentation generation
- **gosec**: Security scanning
- **mockgen**: Mock generation
- **pre-commit**: Git hooks for code quality

## Quick Start

### 1. Initial Setup

```bash
# Clone the repository
git clone <repository-url>
cd kisanlink-ecom

# Set up development environment
make dev-setup-full
```

### 2. Development Workflow

```bash
# Start development with hot reload
make dev-hot

# Or run without hot reload
make dev

# Run tests
make test

# Run tests with coverage
make test-coverage
```

## Build Targets

### Development Builds

```bash
# Build for development (with debug symbols)
make build

# Run directly without building binary
make dev

# Run with hot reload (requires air)
make dev-hot
```

### Production Builds

```bash
# Build optimized production binary
make build-prod

# Build for staging environment
make build-staging

# Build for multiple platforms
make build-all
```

### Build Optimization

The build system uses different optimization levels:

- **Development**: Fast compilation, debug symbols included
- **Staging**: Optimized but with some debug info
- **Production**: Fully optimized, stripped binaries

## Environment Configuration

### Build Variables

The build system injects metadata into binaries:

- `Version`: Git tag or commit hash
- `BuildTime`: UTC build timestamp
- `GitCommit`: Short commit hash
- `GitBranch`: Current git branch

### Environment-Specific Flags

```makefile
# Development
LDFLAGS_DEV=-ldflags "$(LDFLAGS_BASE)"

# Production (optimized and stripped)
LDFLAGS_PROD=-ldflags "-s -w $(LDFLAGS_BASE)"
```

## Testing

### Test Types

```bash
# Unit tests
make test

# Tests with coverage report
make test-coverage

# Tests with coverage enforcement (90% minimum)
make test-coverage-check

# Race condition detection
make test-race

# Benchmarks
make bench

# Test specific package
make test-package PACKAGE=internal/services/orders
```

### Coverage Requirements

- **Overall Coverage**: Minimum 90%
- **Branch Coverage**: Minimum 80% (enforced by golangci-lint)

## Code Quality

### Linting and Formatting

```bash
# Format code
make fmt

# Vet code
make vet

# Lint with golangci-lint
make lint

# Security check with gosec
make security-check

# Run all pre-commit checks
make pre-commit
```

### Pre-commit Hooks

```bash
# Install pre-commit hooks
make install-pre-commit

# Run pre-commit on all files
make pre-commit-all
```

## Documentation

### API Documentation

```bash
# Generate Swagger documentation
make swagger

# The documentation will be available at:
# - JSON: docs/swagger.json
# - YAML: docs/swagger.yaml
# - UI: http://localhost:8080/docs (when server is running)
```

## Docker Development

### Local Services

```bash
# Start PostgreSQL and Redis
make docker-dev-up

# Stop services
make docker-dev-down

# Reset environment (removes volumes)
make docker-dev-reset
```

### Application Container

```bash
# Build Docker image
make docker-build

# Run container
make docker-run

# Stop container
make docker-stop
```

## CI/CD Pipelines

### Continuous Integration

```bash
# Full CI pipeline (for main branch)
make ci

# PR CI pipeline (faster, for pull requests)
make ci-pr
```

### Continuous Deployment

```bash
# Development deployment
make cd-dev

# Staging deployment
make cd-staging

# Production deployment
make cd-prod
```

## Dependency Management

### Go Modules

```bash
# Download dependencies
make deps

# Clean up dependencies
make deps-tidy

# Update all dependencies
make deps-update
```

### Development Tools

```bash
# Install all development tools
make install-tools

# Individual tool installation
make install-swagger
make install-air
make install-security-tools
make install-mockgen
```

## Performance and Profiling

### Profiling

```bash
# Run with CPU and memory profiling
make profile

# Analyze profiles
make analyze-profile
```

### Benchmarking

```bash
# Run benchmarks
make bench
```

## Troubleshooting

### Common Issues

1. **Build Failures**

   ```bash
   # Clean and rebuild
   make clean
   make build
   ```

2. **Dependency Issues**

   ```bash
   # Clean module cache and retry
   go clean -modcache
   make deps-tidy
   ```

3. **Test Failures**

   ```bash
   # Run tests with verbose output
   make test GOTEST="go test -v"
   ```

4. **Coverage Issues**
   ```bash
   # Check coverage without enforcement
   make test-coverage
   # View coverage report in browser: coverage/coverage.html
   ```

### Environment Variables

Key environment variables that affect builds:

- `GOOS`: Target operating system
- `GOARCH`: Target architecture
- `CGO_ENABLED`: Enable/disable CGO
- `GIN_MODE`: Gin framework mode (debug/release)

## Best Practices

### Development Workflow

1. **Start Development**

   ```bash
   make dev-setup-full
   make dev-hot
   ```

2. **Before Committing**

   ```bash
   make pre-commit
   ```

3. **Before Pushing**
   ```bash
   make ci-pr
   ```

### Build Optimization

1. **Use appropriate build target**:
   - `make build` for development
   - `make build-prod` for production

2. **Enable build caching**:
   - Go module cache is automatically used
   - Docker layer caching in CI/CD

3. **Cross-compilation**:
   ```bash
   # Build for Linux on macOS
   GOOS=linux GOARCH=amd64 make build
   ```

### Testing Strategy

1. **Run tests frequently**:

   ```bash
   make test
   ```

2. **Check coverage regularly**:

   ```bash
   make test-coverage-check
   ```

3. **Use race detection**:
   ```bash
   make test-race
   ```

## Integration with IDEs

### VS Code

Add to `.vscode/tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Build",
      "type": "shell",
      "command": "make build",
      "group": "build"
    },
    {
      "label": "Test",
      "type": "shell",
      "command": "make test",
      "group": "test"
    }
  ]
}
```

### GoLand/IntelliJ

Configure external tools for Make targets in Settings > Tools > External Tools.

## Release Process

### Version Management

Versions are automatically determined from Git tags:

```bash
# Create a release tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# Build will include version metadata
make build-prod
```

### Release Checklist

1. Update CHANGELOG.md
2. Run full test suite: `make ci`
3. Create git tag
4. Build release binaries: `make build-all`
5. Generate documentation: `make swagger`
6. Deploy to staging: `make cd-staging`
7. Deploy to production: `make cd-prod`

## Support

For build system issues:

1. Check this documentation
2. Run `make help` for available targets
3. Check build logs in `tmp/` directory
4. Review CI/CD pipeline logs
5. Contact the development team

## Contributing

When modifying the build system:

1. Update this documentation
2. Test changes across different environments
3. Ensure backward compatibility
4. Update CI/CD pipelines if needed
5. Add appropriate help text to Makefile
