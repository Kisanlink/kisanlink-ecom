# KisanLink E-Commerce API

A modern, modular e-commerce API built with Go, Gin, and the kisanlink-db package for flexible data management.

## 🚀 Quick Start

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- Git

### Development Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd kisanlink-ecom
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your preferred settings
   ```

3. **Start development environment**
   ```bash
   ./scripts/start-dev.sh
   ```

   This script will:
   - Start PostgreSQL and Redis containers
   - Install Go dependencies
   - Build the application
   - Display connection information

4. **Run the API server**
   ```bash
   # Using the built binary
   ./bin/server
   
   # Or run directly with Go
   go run cmd/server/main.go
   ```

## 🗄️ Database Setup

### PostgreSQL (Recommended for Development)

The project includes a `docker-compose.yml` file that sets up:
- PostgreSQL 15 database
- Redis for caching
- pgAdmin for database management

**Default Credentials:**
- Database: `kisanlink_ecom`
- Username: `postgres`
- Password: `your_secure_password_here` (change in .env)

**pgAdmin Access:**
- URL: http://localhost:5050
- Email: `admin@kisanlink.local`
- Password: `admin123`

### Alternative Database Options

The application supports multiple database backends:

1. **In-Memory (Testing)**
   ```env
   DB_PROVIDER=inmemory
   ```

2. **DynamoDB (Production)**
   ```env
   DB_PROVIDER=dynamodb
   AWS_REGION=us-east-1
   AWS_ACCESS_KEY_ID=your_key
   AWS_SECRET_ACCESS_KEY=your_secret
   ```

## 🛠️ Development Commands

```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Build the application
go build -o bin/server cmd/server/main.go

# Run linter
make lint

# Format code
make format

# Run all quality checks
make check
```

## 🐳 Docker Commands

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down

# Restart services
docker-compose restart

# Database shell access
docker-compose exec postgres psql -U postgres -d kisanlink_ecom

# Redis CLI access
docker-compose exec redis redis-cli
```

## 🏗️ Project Structure

```
kisanlink-ecom/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Database manager with multi-backend support
│   ├── handlers/        # HTTP request handlers
│   ├── middleware/      # HTTP middleware
│   ├── models/          # Business models extending base.BaseModel
│   ├── repositories/    # Data access layer with filtering
│   ├── routes/          # Route definitions
│   ├── server/          # Server setup and configuration
│   └── utils/           # Utility functions and helpers
├── scripts/             # Development and deployment scripts
├── docs/                # Documentation
├── .env.example         # Environment variables template
├── docker-compose.yml   # Docker services definition
└── Makefile            # Build and development commands
```

## 🔧 Configuration

The application uses environment variables for configuration. See `.env.example` for all available options.

### Key Configuration Sections:

- **Server**: Port, mode, API version
- **Database**: Provider, connection settings, pool configuration
- **JWT**: Secret key, token expiry
- **Redis**: Connection settings for caching
- **Email**: SMTP configuration for notifications
- **File Upload**: Directory and size limits
- **CORS**: Cross-origin request settings
- **Logging**: Level and format configuration

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests with coverage
go test ./... -cover

# Run integration tests
go test ./internal -v

# Run specific test package
go test ./internal/handlers -v
```

## 📊 API Endpoints

### Health Check
- `GET /health` - Service health status

### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/logout` - User logout

### Users
- `GET /api/v1/users` - List users
- `POST /api/v1/users` - Create user
- `GET /api/v1/users/:id` - Get user by ID
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

### Products
- `GET /api/v1/products` - List products
- `POST /api/v1/products` - Create product
- `GET /api/v1/products/:id` - Get product by ID
- `PUT /api/v1/products/:id` - Update product
- `DELETE /api/v1/products/:id` - Delete product

### Orders
- `GET /api/v1/orders` - List orders
- `POST /api/v1/orders` - Create order
- `GET /api/v1/orders/:id` - Get order by ID
- `PUT /api/v1/orders/:id` - Update order
- `DELETE /api/v1/orders/:id` - Delete order

## 🚀 Deployment

The application is production-ready and can be deployed using:

1. **Docker**: Build and run the container
2. **Binary**: Compile and deploy the binary
3. **Cloud Services**: AWS, GCP, Azure with appropriate database services

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run `make check` to ensure code quality
6. Submit a pull request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Troubleshooting

### Database Connection Issues

1. Ensure PostgreSQL container is running:
   ```bash
   docker-compose ps
   ```

2. Check database logs:
   ```bash
   docker-compose logs postgres
   ```

3. Test database connection:
   ```bash
   docker-compose exec postgres pg_isready -U postgres
   ```

### Build Issues

1. Ensure Go modules are up to date:
   ```bash
   go mod tidy
   ```

2. Clear module cache if needed:
   ```bash
   go clean -modcache
   go mod download
   ```

### Port Conflicts

If you encounter port conflicts, update the ports in `docker-compose.yml` and your `.env` file accordingly.

## 📞 Support

For questions and support, please open an issue in the GitHub repository.