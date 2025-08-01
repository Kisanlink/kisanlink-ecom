#!/bin/bash

# KisanLink E-Commerce API Testing Script
# This script tests all the main API endpoints

set -e

# Configuration
API_BASE_URL="http://localhost:8080"
CONTENT_TYPE="Content-Type: application/json"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "🧪 KisanLink E-Commerce API Testing"
echo "=================================="
echo "Base URL: $API_BASE_URL"
echo ""

# Function to make HTTP requests and display results
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    echo -e "${BLUE}Testing: $description${NC}"
    echo "  $method $endpoint"
    
    if [ -n "$data" ]; then
        echo "  Data: $data"
        response=$(curl -s -w "HTTPSTATUS:%{http_code}" \
            -X "$method" \
            -H "$CONTENT_TYPE" \
            -d "$data" \
            "$API_BASE_URL$endpoint")
    else
        response=$(curl -s -w "HTTPSTATUS:%{http_code}" \
            -X "$method" \
            "$API_BASE_URL$endpoint")
    fi
    
    # Extract HTTP status code
    http_code=$(echo "$response" | grep -o "HTTPSTATUS:[0-9]*" | cut -d: -f2)
    # Extract response body
    response_body=$(echo "$response" | sed 's/HTTPSTATUS:[0-9]*$//')
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo -e "  ${GREEN}✅ Status: $http_code${NC}"
    elif [ "$http_code" -ge 400 ] && [ "$http_code" -lt 500 ]; then
        echo -e "  ${YELLOW}⚠️  Status: $http_code${NC}"
    else
        echo -e "  ${RED}❌ Status: $http_code${NC}"
    fi
    
    # Pretty print JSON response
    if command -v jq &> /dev/null; then
        echo "  Response: $(echo "$response_body" | jq -c .)"
    else
        echo "  Response: $response_body"
    fi
    echo ""
}

# Check if API server is running
echo -e "${BLUE}🔍 Checking if API server is running...${NC}"
if ! curl -s "$API_BASE_URL/health" > /dev/null; then
    echo -e "${RED}❌ API server is not responding at $API_BASE_URL${NC}"
    echo -e "${YELLOW}💡 Make sure to start the server first:${NC}"
    echo "   go run cmd/server/main.go"
    echo "   or"
    echo "   ./bin/server"
    exit 1
fi

echo -e "${GREEN}✅ API server is running!${NC}"
echo ""

# 1. Health Check
make_request "GET" "/health" "" "Health Check"

# 2. User Management
echo -e "${BLUE}👥 Testing User Management${NC}"
echo "=========================="

# Create a user
USER_DATA='{
    "username": "testuser",
    "email": "test@example.com",
    "full_name": "Test User",
    "role": "customer",
    "password": "password123"
}'
make_request "POST" "/api/v1/users" "$USER_DATA" "Create User"

# Get all users
make_request "GET" "/api/v1/users" "" "Get All Users"

# Get user by ID (this will fail since we need the actual ID)
make_request "GET" "/api/v1/users/test123" "" "Get User by ID"

# Update user (this will also fail without real ID)
UPDATE_USER_DATA='{
    "full_name": "Updated Test User"
}'
make_request "PUT" "/api/v1/users/test123" "$UPDATE_USER_DATA" "Update User"

# 3. Product Management
echo -e "${BLUE}🛍️  Testing Product Management${NC}"
echo "============================="

# Create a product
PRODUCT_DATA='{
    "name": "Test Product",
    "description": "A test product for API testing",
    "price": 29.99,
    "currency": "USD",
    "category": "Electronics",
    "stock": 100
}'
make_request "POST" "/api/v1/products" "$PRODUCT_DATA" "Create Product"

# Get all products
make_request "GET" "/api/v1/products" "" "Get All Products"

# Get product by ID
make_request "GET" "/api/v1/products/test123" "" "Get Product by ID"

# Update product
UPDATE_PRODUCT_DATA='{
    "price": 24.99,
    "stock": 90
}'
make_request "PUT" "/api/v1/products/test123" "$UPDATE_PRODUCT_DATA" "Update Product"

# 4. Order Management
echo -e "${BLUE}📦 Testing Order Management${NC}"
echo "=========================="

# Create an order
ORDER_DATA='{
    "user_id": "test_user_123",
    "items": [
        {
            "product_id": "test_product_123",
            "quantity": 2,
            "price": 29.99
        }
    ],
    "currency": "USD"
}'
make_request "POST" "/api/v1/orders" "$ORDER_DATA" "Create Order"

# Get all orders
make_request "GET" "/api/v1/orders" "" "Get All Orders"

# Get order by ID
make_request "GET" "/api/v1/orders/test123" "" "Get Order by ID"

# Update order status
UPDATE_ORDER_DATA='{
    "status": "confirmed"
}'
make_request "PUT" "/api/v1/orders/test123" "$UPDATE_ORDER_DATA" "Update Order Status"

# 5. Authentication
echo -e "${BLUE}🔐 Testing Authentication${NC}"
echo "======================="

# Register user
REGISTER_DATA='{
    "username": "newuser",
    "email": "newuser@example.com",
    "full_name": "New User",
    "role": "customer",
    "password": "newpassword123"
}'
make_request "POST" "/api/v1/auth/register" "$REGISTER_DATA" "User Registration"

# Login user
LOGIN_DATA='{
    "username": "testuser",
    "password": "password123"
}'
make_request "POST" "/api/v1/auth/login" "$LOGIN_DATA" "User Login"

# Logout user
make_request "POST" "/api/v1/auth/logout" "" "User Logout"

# 6. Error Cases
echo -e "${BLUE}⚠️  Testing Error Cases${NC}"
echo "====================="

# Invalid JSON
make_request "POST" "/api/v1/users" "invalid json" "Invalid JSON"

# Missing required fields
INVALID_USER_DATA='{
    "username": "ab"
}'
make_request "POST" "/api/v1/users" "$INVALID_USER_DATA" "Invalid User Data"

# Non-existent endpoint
make_request "GET" "/api/v1/nonexistent" "" "Non-existent Endpoint"

echo ""
echo -e "${GREEN}🎉 API Testing Complete!${NC}"
echo "======================="
echo ""
echo -e "${YELLOW}💡 Notes:${NC}"
echo "  • Some tests may fail if specific IDs don't exist"
echo "  • This is normal behavior for a clean database"
echo "  • Focus on HTTP status codes and response structure"
echo "  • 200-299: Success"
echo "  • 400-499: Client errors (expected for invalid data)"
echo "  • 500-599: Server errors (should investigate)"
echo ""
echo -e "${BLUE}🔧 Next Steps:${NC}"
echo "  • Test with real data using the actual application"
echo "  • Set up proper test database with fixtures"
echo "  • Implement comprehensive integration tests" 