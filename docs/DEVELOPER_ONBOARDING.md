# Developer Onboarding Guide

## Welcome to KisanLink E-commerce Service

This guide will help you set up your development environment and get started contributing to the KisanLink E-commerce Service.

## Prerequisites

### System Requirements

- **Operating System**: macOS, Linux, or Windows with WSL2
- **RAM**: Minimum 8GB, recommended 16GB
- **Storage**: At least 10GB free space
- **Network**: Stable internet connection for downloading dependencies

### Required Software

1. **Go 1.24+**

   ```bash
   # macOS (using Homebrew)
   brew install go

   # Linux (using package manager)
   sudo apt-get install golang-go  # Ubuntu/Debian
   sudo yum install golang         # CentOS/RHEL

   # Windows
   # Download from https://golang.org/dl/
   ```

2. **Git**

   ```bash
   # macOS
   brew install git

   # Linux
   sudo apt-get install git  # Ubuntu/Debian
   sudo yum install git      # CentOS/RHEL

   # Windows
   # Download from https://git-scm.com/
   ```

3. **Make**

   ```bash
   # macOS (usually pre-installed)
   xcode-select --install

   # Linux (usually pre-installed)
   sudo apt-get install build-essential  # Ubuntu/Debian
   sudo yum groupinstall "Development Tools"  # CentOS/RHEL

   # Windows (with WSL2)
   sudo apt-get install build-essential
   ```

4. **Docker & Docker Compose**

   ```bash
   # macOS
   # Download Docker Desktop from https://www.docker.com/products/docker-desktop

   # Linux
   # Follow instructions at https://docs.docker.com/engine/install/

   # Windows
   # Download Docker Desktop from https://www.docker.com/products/docker-desktop
   ```

## Quick Setup (5 minutes)

### 1. Clone the Repository

```bash
git clone <repository-url>
cd kisanlink-ecom
```

### 2. Environment Setup

```bash
# Copy environment configuration
cp .env.example .env

# Edit .env file with your local settings
# Most defaults should work for development
```

### 3. Full Development Setup

```bash
# This will:
# - Install all development tools
# - Set up pre-commit hooks
# - Generate API documentation
# - Start database services
make dev-setup-full
```

### 4. Verify Setup

```bash
# Run tests to verify everything works
make test

# Start the development server
make dev-hot
```

Visit http://localhost:8080/health to verify the server is running.

## Detailed Setup

### Development Tools Installation

The build system can automatically install development tools:

```bash
# Install all tools at once
make install-tools

# Or install individually
make install-swagger    # API documentation
make install-air        # Hot reload
make install-security-tools  # Security scanning
make install-mockgen    # Mock generation
```

### IDE Setup

#### VS Code (Recommended)

1. **Install Go Extension**
   - Install the official Go extension by Google
   - Install the Go tools when prompted

2. **Recommended Extensions**

   ```json
   {
     "recommendations": [
       "golang.go",
       "ms-vscode.vscode-json",
       "redhat.vscode-yaml",
       "ms-vscode.makefile-tools",
       "humao.rest-client",
       "42crunch.vscode-openapi"
     ]
   }
   ```

3. **Workspace Settings**
   ```json
   {
     "go.useLanguageServer": true,
     "go.lintTool": "golangci-lint",
     "go.lintOnSave": "package",
     "go.formatTool": "goimports",
     "go.testFlags": ["-v"],
     "go.coverOnSave": true
   }
   ```

#### GoLand/IntelliJ IDEA

1. Install Go plugin
2. Configure Go SDK to point to your Go installation
3. Set up external tools for Make targets
4. Configure code style to match project settings

### Database Setup

#### Local Development (Docker)

```bash
# Start PostgreSQL and Redis
make docker-dev-up

# Verify services are running
docker-compose ps

# View logs
docker-compose logs -f postgres
```

#### Manual Database Setup

If you prefer not to use Docker:

1. **PostgreSQL**

   ```bash
   # macOS
   brew install postgresql
   brew services start postgresql

   # Create database
   createdb kisanlink_ecom
   ```

2. **Redis** (optional)
   ```bash
   # macOS
   brew install redis
   brew services start redis
   ```

### Environment Configuration

Edit your `.env` file:

```bash
# Database (if using local PostgreSQL)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=kisanlink_ecom

# Development settings
GIN_MODE=debug
LOG_LEVEL=debug
LOG_FORMAT=text

# JWT (use a secure secret in production)
JWT_SECRET=your_development_jwt_secret_here
```

## Development Workflow

### Daily Development

1. **Start Development Environment**

   ```bash
   # Start database services
   make docker-dev-up

   # Start development server with hot reload
   make dev-hot
   ```

2. **Make Changes**
   - Edit code in your preferred IDE
   - The server will automatically reload on changes

3. **Run Tests**

   ```bash
   # Run all tests
   make test

   # Run tests with coverage
   make test-coverage

   # Run specific package tests
   make test-package PACKAGE=internal/services/orders
   ```

4. **Code Quality Checks**

   ```bash
   # Format code
   make fmt

   # Run linter
   make lint

   # Run all pre-commit checks
   make pre-commit
   ```

### Git Workflow

1. **Create Feature Branch**

   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make Changes and Commit**

   ```bash
   # Pre-commit hooks will run automatically
   git add .
   git commit -m "feat: add new feature"
   ```

3. **Before Pushing**

   ```bash
   # Run CI checks locally
   make ci-pr
   ```

4. **Push and Create PR**
   ```bash
   git push origin feature/your-feature-name
   ```

### Testing Strategy

#### Unit Tests

```bash
# Run all unit tests
make test

# Run with race detection
make test-race

# Run with coverage
make test-coverage
```

#### Integration Tests

```bash
# Run integration tests (requires database)
make test GOTEST="go test -tags=integration"
```

#### API Testing

```bash
# Test API endpoints
make test-endpoints

# Test AAA service integration
make test-aaa
```

### API Documentation

```bash
# Generate Swagger docs
make swagger

# View documentation
# Start server: make dev
# Visit: http://localhost:8080/docs
```

## Project Structure

```
kisanlink-ecom/
├── cmd/                    # Application entry points
│   ├── server/            # Main HTTP server
│   ├── migrate/           # Database migrations
│   └── testrunner/        # Test utilities
├── internal/              # Private application code
│   ├── handlers/          # HTTP handlers
│   ├── services/          # Business logic
│   ├── repositories/      # Data access layer
│   ├── middleware/        # HTTP middleware
│   ├── config/           # Configuration management
│   └── auth/             # Authentication logic
├── entities/              # Data models and DTOs
│   ├── models/           # Domain models
│   ├── requests/         # Request DTOs
│   └── responses/        # Response DTOs
├── tests/                # Test files
├── docs/                 # Documentation
├── migrations/           # Database migrations
└── debug/               # Debug utilities
```

## Common Tasks

### Adding a New API Endpoint

1. **Define Request/Response Models**

   ```go
   // entities/requests/your_domain/your_requests.go
   type CreateItemRequest struct {
       Name string `json:"name" validate:"required"`
   }
   ```

2. **Create Handler**

   ```go
   // internal/handlers/your_domain/handler.go
   func (h *Handler) CreateItem(c *gin.Context) {
       // Implementation
   }
   ```

3. **Add Route**

   ```go
   // internal/routes/routes.go
   v1.POST("/items", handlers.CreateItem)
   ```

4. **Add Swagger Documentation**

   ```go
   // @Summary Create item
   // @Description Create a new item
   // @Tags items
   // @Accept json
   // @Produce json
   // @Param item body CreateItemRequest true "Item data"
   // @Success 201 {object} ItemResponse
   // @Router /api/v1/items [post]
   ```

5. **Write Tests**
   ```go
   // tests/handlers/your_domain/handler_test.go
   func TestCreateItem(t *testing.T) {
       // Test implementation
   }
   ```

### Adding Database Migrations

1. **Create Migration File**

   ```bash
   # migrations/002_add_new_table.go
   ```

2. **Run Migration**
   ```bash
   make db-migrate
   ```

### Adding New Dependencies

1. **Add to go.mod**

   ```bash
   go get github.com/new/dependency
   ```

2. **Tidy Dependencies**
   ```bash
   make deps-tidy
   ```

## Debugging

### Application Debugging

1. **Enable Debug Logging**

   ```bash
   # In .env file
   LOG_LEVEL=debug
   GIN_MODE=debug
   ```

2. **Use Debugger**
   - VS Code: Set breakpoints and use F5
   - GoLand: Set breakpoints and use debug configuration

3. **Profiling**

   ```bash
   # Run with profiling
   make profile

   # Analyze profiles
   make analyze-profile
   ```

### Database Debugging

1. **View Database**

   ```bash
   # Using pgAdmin (if running)
   # Visit: http://localhost:5050

   # Or use psql
   docker exec -it kisanlink-postgres psql -U postgres -d kisanlink_ecom
   ```

2. **Check Logs**
   ```bash
   docker-compose logs postgres
   ```

## Troubleshooting

### Common Issues

1. **Port Already in Use**

   ```bash
   # Find process using port 8080
   lsof -i :8080

   # Kill process
   kill -9 <PID>
   ```

2. **Database Connection Issues**

   ```bash
   # Check if PostgreSQL is running
   docker-compose ps

   # Restart database
   make docker-dev-reset
   ```

3. **Go Module Issues**

   ```bash
   # Clean module cache
   go clean -modcache

   # Re-download dependencies
   make deps
   ```

4. **Build Issues**
   ```bash
   # Clean and rebuild
   make clean
   make build
   ```

### Getting Help

1. **Documentation**
   - Read [BUILD.md](BUILD.md) for build system details
   - Check API documentation at `/docs` endpoint
   - Review code comments and examples

2. **Team Communication**
   - Ask questions in team chat
   - Create GitHub issues for bugs
   - Request code reviews for complex changes

3. **External Resources**
   - [Go Documentation](https://golang.org/doc/)
   - [Gin Framework](https://gin-gonic.com/)
   - [PostgreSQL Documentation](https://www.postgresql.org/docs/)

## Best Practices

### Code Style

1. **Follow Go Conventions**
   - Use `gofmt` for formatting
   - Follow effective Go guidelines
   - Use meaningful variable names

2. **Error Handling**

   ```go
   if err != nil {
       return fmt.Errorf("operation failed: %w", err)
   }
   ```

3. **Testing**
   - Write tests for all public functions
   - Use table-driven tests
   - Mock external dependencies

### Git Practices

1. **Commit Messages**

   ```
   feat: add user authentication
   fix: resolve database connection issue
   docs: update API documentation
   test: add unit tests for order service
   ```

2. **Branch Naming**
   ```
   feature/user-authentication
   bugfix/database-connection
   hotfix/security-vulnerability
   ```

### Security

1. **Never commit secrets**
   - Use `.env` files for local development
   - Use environment variables in production

2. **Validate all inputs**
   - Use struct tags for validation
   - Sanitize user inputs

3. **Follow security best practices**
   - Run security checks: `make security-check`
   - Keep dependencies updated

## Next Steps

1. **Complete the setup** by running `make dev-setup-full`
2. **Explore the codebase** starting with `cmd/server/main.go`
3. **Run the test suite** with `make test`
4. **Try making a small change** and see the hot reload in action
5. **Read the API documentation** at http://localhost:8080/docs
6. **Join the team chat** and introduce yourself

Welcome to the team! 🚀
