# Project Structure & Architecture

## Directory Organization

### Core Application

- `cmd/` - Application entry points
  - `cmd/server/` - Main HTTP server
  - `cmd/testrunner/` - Test runner utility
- `internal/` - Private application code (not importable by other projects)
- `pkg/` - Public packages (importable by other projects)

### Internal Architecture (Clean Architecture Pattern)

```
internal/
├── config/          # Configuration management
├── server/          # Server setup and initialization
├── routes/          # Route definitions and setup
├── middleware/      # HTTP middleware (auth, logging, CORS)
├── handlers/        # HTTP request handlers (controllers)
├── services/        # Business logic layer
├── repositories/    # Data access layer
├── database/        # Database connection management
├── auth/           # Authentication client integration
├── utils/          # Utility functions and helpers
└── testutils/      # Testing utilities
```

## Architecture Patterns

### Layered Architecture

1. **Handlers** - HTTP request/response handling
2. **Services** - Business logic and orchestration
3. **Repositories** - Data access and persistence
4. **Models** - Data structures and domain objects

### Dependency Flow

- Handlers depend on Services
- Services depend on Repositories
- Repositories depend on Database abstractions
- All layers can use Models and Utils

### Key Conventions

#### File Naming

- Use snake_case for file names: `user_handler.go`, `order_service.go`
- Group related functionality in packages: `handlers/catalog/`, `services/orders/`
- Test files follow `*_test.go` pattern
- Integration tests use `*_integration_test.go` pattern

#### Package Structure

- Each domain has its own package structure:
  - `handlers/catalog/` - Catalog-related handlers
  - `services/catalog/` - Catalog business logic
  - `repositories/catalog/` - Catalog data access
  - `models/catalog/` - Catalog domain models

#### Base Types

- All models extend `base.BaseModel` from kisanlink-db package
- Common response structures in `internal/models/common/`
- Standard API response format with success/error structure

#### Error Handling

- Use structured error responses via `utils.ErrorResponse()`
- Validation errors via `utils.ValidationErrorResponse()`
- Success responses via `utils.SuccessResponse()`

#### Authentication & Authorization

- JWT-based authentication through AAA service
- Conditional middleware for development (mock auth when AAA unavailable)
- Authorization checks in handlers before business logic

## Configuration Management

- Environment-based configuration in `internal/config/`
- Support for multiple database backends
- Centralized configuration loading with defaults
- Docker Compose for development environment setup
