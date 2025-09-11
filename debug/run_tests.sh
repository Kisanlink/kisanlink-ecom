#!/bin/bash

# Debug Tools Runner Script
# This script runs different debug tools without package conflicts

set -e

echo "🔧 Debug Tools Runner"
echo "====================="

# Function to run a test
run_test() {
    local test_name="$1"
    local file="$2"

    echo ""
    echo "🧪 Running: $test_name"
    echo "----------------------------------------"

    if [ -f "debug/$file" ]; then
        cd debug
        go run "$file"
        cd - > /dev/null
    else
        echo "❌ File not found: debug/$file"
        return 1
    fi
}

# Function to show menu
show_menu() {
    echo ""
    echo "Available tests:"
    echo "1. Check integration status"
    echo "2. Health check AAA service"
    echo "3. Test AAA integration"
    echo "4. Seed RBAC data"
    echo "5. Test endpoints"
    echo "6. Run all tests"
    echo "7. Exit"
    echo ""
    read -p "Choose an option (1-7): " choice

    case $choice in
        1) run_test "Integration Status Check" "check_integration_status.go" ;;
        2) run_test "AAA Service Health Check" "health_check.go" ;;
        3) run_test "AAA Integration Test" "aaa_test_runner.go" ;;
        4) run_test "RBAC Data Seeding" "seed_ecommerce_rbac.go" ;;
        5) run_test "API Endpoints Test" "test_ecommerce_endpoints.go" ;;
        6) run_all_tests ;;
        7) echo "👋 Goodbye!"; exit 0 ;;
        *) echo "❌ Invalid option. Please choose 1-7." ;;
    esac
}

# Function to run all tests
run_all_tests() {
    echo ""
    echo "🚀 Running all tests..."
    echo "======================="

    run_test "Integration Status Check" "check_integration_status.go"
    run_test "AAA Service Health Check" "health_check.go"
    run_test "AAA Integration Test" "aaa_test_runner.go"
    run_test "RBAC Data Seeding" "seed_ecommerce_rbac.go"

    echo ""
    echo "✅ All tests completed!"
}

# Main execution
if [ "$1" = "all" ]; then
    run_all_tests
elif [ "$1" = "status" ]; then
    run_test "Integration Status Check" "check_integration_status.go"
elif [ "$1" = "health" ]; then
    run_test "AAA Service Health Check" "health_check.go"
elif [ "$1" = "test" ]; then
    run_test "AAA Integration Test" "aaa_test_runner.go"
elif [ "$1" = "seed" ]; then
    run_test "RBAC Data Seeding" "seed_ecommerce_rbac.go"
elif [ "$1" = "endpoints" ]; then
    run_test "API Endpoints Test" "test_ecommerce_endpoints.go"
elif [ "$1" = "menu" ] || [ -z "$1" ]; then
    show_menu
else
    echo "Usage: $0 [all|status|health|test|seed|endpoints|menu]"
    echo ""
    echo "Examples:"
    echo "  $0 all        # Run all tests"
    echo "  $0 status     # Check integration status"
    echo "  $0 health     # Health check AAA service"
    echo "  $0 test       # Test AAA integration"
    echo "  $0 seed       # Seed RBAC data"
    echo "  $0 endpoints  # Test API endpoints"
    echo "  $0 menu       # Interactive menu (default)"
    exit 1
fi
