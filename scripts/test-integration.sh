#!/bin/bash

# Integration Test Runner for KisanLink E-Commerce API
# This script runs integration tests with proper environment configuration

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 KisanLink E-Commerce Integration Test Runner${NC}"
echo "=================================================="

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo -e "${RED}❌ .env file not found. Please create one based on .env.example${NC}"
    exit 1
fi

# Load environment variables from .env file
echo -e "${YELLOW}📋 Loading environment variables from .env file...${NC}"
export $(grep -v '^#' .env | xargs)

# Validate required environment variables
echo -e "${YELLOW}🔍 Validating environment configuration...${NC}"

# Database configuration
if [ -z "$DB_POSTGRES_HOST" ]; then
    echo -e "${RED}❌ DB_POSTGRES_HOST is not set${NC}"
    exit 1
fi

if [ -z "$DB_POSTGRES_USER" ]; then
    echo -e "${RED}❌ DB_POSTGRES_USER is not set${NC}"
    exit 1
fi

if [ -z "$DYNAMODB_ACCESS_KEY_ID" ]; then
    echo -e "${RED}❌ DYNAMODB_ACCESS_KEY_ID is not set${NC}"
    exit 1
fi

if [ -z "$DYNAMODB_SECRET_ACCESS_KEY" ]; then
    echo -e "${RED}❌ DYNAMODB_SECRET_ACCESS_KEY is not set${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Environment configuration validated${NC}"

# Print configuration summary (without sensitive data)
echo -e "${YELLOW}📊 Configuration Summary:${NC}"
echo "  PostgreSQL Host: $DB_POSTGRES_HOST"
echo "  PostgreSQL Port: ${DB_POSTGRES_PORT:-5432}"
echo "  PostgreSQL Database: ${DB_POSTGRES_DBNAME:-kisanlink_ecom}"
echo "  PostgreSQL User: $DB_POSTGRES_USER"
echo "  DynamoDB Region: ${DB_DYNAMO_REGION:-us-east-1}"
echo "  DynamoDB Table: ${DB_DYNAMO_TABLE:-kisanlink_ecom}"
echo "  DynamoDB Endpoint: ${DYNAMODB_ENDPOINT:-default}"
echo "  DynamoDB SSL Disabled: ${DYNAMODB_DISABLE_SSL:-false}"

# Check if databases are available
echo -e "${YELLOW}🔌 Checking database connectivity...${NC}"

# Test PostgreSQL connection
if command -v psql &> /dev/null; then
    if PGPASSWORD="$DB_POSTGRES_PASSWORD" psql -h "$DB_POSTGRES_HOST" -p "${DB_POSTGRES_PORT:-5432}" -U "$DB_POSTGRES_USER" -d "${DB_POSTGRES_DBNAME:-kisanlink_ecom}" -c "SELECT 1;" &> /dev/null; then
        echo -e "${GREEN}✅ PostgreSQL connection successful${NC}"
    else
        echo -e "${YELLOW}⚠️  PostgreSQL connection failed - tests may be skipped${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  psql not available - skipping PostgreSQL connectivity check${NC}"
fi

# Test DynamoDB connection (if endpoint is local)
if [[ "$DYNAMODB_ENDPOINT" == *"localhost"* ]] || [[ "$DYNAMODB_ENDPOINT" == *"127.0.0.1"* ]]; then
    if command -v curl &> /dev/null; then
        if curl -s "$DYNAMODB_ENDPOINT" &> /dev/null; then
            echo -e "${GREEN}✅ DynamoDB local endpoint accessible${NC}"
        else
            echo -e "${YELLOW}⚠️  DynamoDB local endpoint not accessible - tests may be skipped${NC}"
        fi
    else
        echo -e "${YELLOW}⚠️  curl not available - skipping DynamoDB connectivity check${NC}"
    fi
fi

# Run integration tests
echo -e "${YELLOW}🧪 Running integration tests...${NC}"

# Set test timeout
TEST_TIMEOUT=${TEST_TIMEOUT:-300s}

# Run specific integration test suites
echo -e "${YELLOW}📦 Running database manager integration tests...${NC}"
go test -v -timeout="$TEST_TIMEOUT" ./tests/integration/database_manager_integration_test.go

echo -e "${YELLOW}📦 Running inventory integration tests...${NC}"
go test -v -timeout="$TEST_TIMEOUT" ./tests/integration/inventory_integration_test.go

echo -e "${YELLOW}📦 Running database integration tests...${NC}"
go test -v -timeout="$TEST_TIMEOUT" ./tests/integration/database_integration_test.go

echo -e "${YELLOW}📦 Running marketplace migration tests...${NC}"
go test -v -timeout="$TEST_TIMEOUT" ./tests/database/marketplace_migration_test.go

# Run all integration tests
echo -e "${YELLOW}📦 Running all integration tests...${NC}"
go test -v -timeout="$TEST_TIMEOUT" -tags=integration ./tests/integration/...

echo -e "${GREEN}🎉 Integration tests completed successfully!${NC}"
echo ""
echo -e "${YELLOW}📋 Test Summary:${NC}"
echo "  - Database manager integration: ✅"
echo "  - Inventory integration: ✅"
echo "  - Database integration: ✅"
echo "  - Marketplace migration: ✅"
echo ""
echo -e "${GREEN}✨ All integration tests passed with environment configuration!${NC}"
