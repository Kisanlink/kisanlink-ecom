# Debug Tools Makefile
# This file helps manage different debug tools without package conflicts

.PHONY: help test-aaa health-check seed-rbac test-endpoints clean status test-all dev-setup

help:
	@echo "Available debug tools:"
	@echo "  status          - Check current integration status"
	@echo "  health-check    - Check AAA service health"
	@echo "  test-aaa        - Test AAA service integration"
	@echo "  seed-rbac       - Seed RBAC data to AAA service"
	@echo "  test-endpoints  - Test e-commerce API endpoints"
	@echo "  test-all        - Run all tests in sequence"
	@echo "  dev-setup       - Setup development environment"
	@echo "  clean           - Clean build artifacts"

status:
	@echo "🔍 Checking AAA Integration Status..."
	@go run check_integration_status.go

health-check:
	@echo "🏥 Checking AAA Service Health..."
	@go run health_check.go

test-aaa:
	@echo "🧪 Testing AAA Service Integration..."
	@go run aaa_test_runner.go

seed-rbac:
	@echo "🌱 Seeding RBAC Data to AAA Service..."
	@go run seed_ecommerce_rbac.go

test-endpoints:
	@echo "🔗 Testing E-commerce API Endpoints..."
	@go run test_ecommerce_endpoints.go

clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -f debug/*.exe debug/debug
	@echo "Clean complete"

# Quick test of all components
test-all: health-check test-aaa seed-rbac
	@echo "✅ All tests completed!"

# Development workflow
dev-setup: health-check seed-rbac
	@echo "🚀 Development environment ready!"

# Quick status check
quick-check: status
	@echo "📊 Integration status checked!"

# Troubleshooting commands
troubleshoot:
	@echo "🔧 Running troubleshooting checks..."
	@echo "1. Checking AAA service connectivity..."
	@go run health_check.go
	@echo "2. Checking integration status..."
	@go run check_integration_status.go
	@echo "3. Checking configuration..."
	@go run -c "package main; import 'kisanlink-ecom/internal/config'; func main() { cfg, _ := config.Load(); println('AAA Server:', cfg.AAA.GRPCServerAddr) }"
