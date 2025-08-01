#!/bin/bash

# KisanLink E-Commerce Development Startup Script
# This script sets up the development environment and starts the services

set -e

echo "🚀 Starting KisanLink E-Commerce Development Environment"
echo "=================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running. Please start Docker and try again.${NC}"
    exit 1
fi

# Check if .env exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}⚠️  .env file not found. Creating from .env.example...${NC}"
    if [ -f .env.example ]; then
        cp .env.example .env
        echo -e "${GREEN}✅ .env file created. Please update the values as needed.${NC}"
    else
        echo -e "${RED}❌ .env.example not found. Please create a .env file manually.${NC}"
        exit 1
    fi
fi

# Start database services
echo -e "${BLUE}📦 Starting database services...${NC}"
docker-compose up -d postgres redis

# Wait for PostgreSQL to be ready
echo -e "${BLUE}⏳ Waiting for PostgreSQL to be ready...${NC}"
timeout=60
counter=0
while ! docker-compose exec -T postgres pg_isready -U postgres -d kisanlink_ecom > /dev/null 2>&1; do
    if [ $counter -ge $timeout ]; then
        echo -e "${RED}❌ PostgreSQL failed to start within $timeout seconds${NC}"
        exit 1
    fi
    echo -e "${YELLOW}⏳ Waiting for PostgreSQL... ($counter/$timeout)${NC}"
    sleep 2
    counter=$((counter + 2))
done

echo -e "${GREEN}✅ PostgreSQL is ready!${NC}"

# Wait for Redis to be ready
echo -e "${BLUE}⏳ Waiting for Redis to be ready...${NC}"
timeout=30
counter=0
while ! docker-compose exec -T redis redis-cli ping > /dev/null 2>&1; do
    if [ $counter -ge $timeout ]; then
        echo -e "${RED}❌ Redis failed to start within $timeout seconds${NC}"
        exit 1
    fi
    echo -e "${YELLOW}⏳ Waiting for Redis... ($counter/$timeout)${NC}"
    sleep 1
    counter=$((counter + 1))
done

echo -e "${GREEN}✅ Redis is ready!${NC}"

# Install Go dependencies
echo -e "${BLUE}📥 Installing Go dependencies...${NC}"
go mod download
go mod tidy

# Run database migrations (if your app supports it)
echo -e "${BLUE}🔄 Running database setup...${NC}"
# Uncomment the following line if you have a migration command
# go run cmd/migrate/main.go

# Build the application
echo -e "${BLUE}🔨 Building the application...${NC}"
go build -o bin/server cmd/server/main.go

echo -e "${GREEN}✅ Build completed successfully!${NC}"

# Display connection information
echo ""
echo -e "${GREEN}🎉 Development environment is ready!${NC}"
echo "=================================================="
echo -e "${BLUE}📊 Service Information:${NC}"
echo "  • API Server:      http://localhost:8080"
echo "  • PostgreSQL:      localhost:5432"
echo "  • Redis:           localhost:6379"
echo "  • pgAdmin:         http://localhost:5050"
echo ""
echo -e "${BLUE}🔐 Database Credentials:${NC}"
echo "  • Database: kisanlink_ecom"
echo "  • Username: postgres" 
echo "  • Password: check your .env file"
echo ""
echo -e "${BLUE}🔐 pgAdmin Credentials:${NC}"
echo "  • Email:    admin@kisanlink.local"
echo "  • Password: admin123"
echo ""
echo -e "${YELLOW}💡 Next Steps:${NC}"
echo "  1. Update your .env file with proper credentials"
echo "  2. Start the API server: ./bin/server"
echo "  3. Or run in development mode: go run cmd/server/main.go"
echo ""
echo -e "${BLUE}🛠️  Useful Commands:${NC}"
echo "  • View logs:        docker-compose logs -f"
echo "  • Stop services:    docker-compose down"
echo "  • Restart services: docker-compose restart"
echo "  • Database shell:   docker-compose exec postgres psql -U postgres -d kisanlink_ecom"
echo ""

# Optionally start the API server
read -p "🚀 Would you like to start the API server now? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${GREEN}🚀 Starting API server...${NC}"
    ./bin/server
fi 