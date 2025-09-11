# Technology Stack & Build System

## Core Technologies

- **Language**: Go 1.24+
- **Web Framework**: Gin (HTTP router and middleware)
- **Database**: Multi-backend support via kisanlink-db package
  - PostgreSQL (production)
  - DynamoDB (production)
- **Authentication**: JWT tokens with external AAA service integration
- **Documentation**: Swagger/OpenAPI with Scalar API Reference
- **Containerization**: Docker & Docker Compose

## Key Dependencies

- `github.com/Kisanlink/kisanlink-db` - Multi-database abstraction layer
- `github.com/gin-gonic/gin` - HTTP web framework
- `github.com/swaggo/swag` - Swagger documentation generation
- `github.com/sirupsen/logrus` - Structured logging
- `google.golang.org/grpc` - gRPC client for AAA service

## Build System

Uses Make for build automation. Key commands:

### Development

```bash
make dev          # Run in development mode
make dev-hot      # Run with hot reload (requires air)
make run          # Build and run
```

### Building

```bash
make build        # Build for current platform
make build-all    # Build for multiple platforms
make clean        # Clean build artifacts
```

### Testing

```bash
make test         # Run all tests
make test-coverage # Run tests with coverage report
make test-race    # Run tests with race detection
```

### Code Quality

```bash
make lint         # Run golangci-lint
make fmt          # Format code with go fmt
make vet          # Run go vet
make pre-commit   # Run all pre-commit checks
```

### Documentation

```bash
make swagger      # Generate Swagger docs
```

### Docker Operations

```bash
docker-compose up -d     # Start all services
docker-compose down      # Stop all services
docker-compose logs -f   # View logs
```

## Configuration

- Environment-based configuration using `.env` files
- Supports multiple database backends through configuration
- JWT and AAA service configuration
- CORS, logging, and upload settings
