#!/bin/bash

# Docker Development Helper Script for KisanLink E-commerce Service
# This script provides convenient commands for Docker-based development

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="kisanlink-ecom"
COMPOSE_FILE="docker-compose.yml"
DEV_COMPOSE_FILE="docker-compose.dev.yml"
PROD_COMPOSE_FILE="docker-compose.prod.yml"
TEST_COMPOSE_FILE="docker-compose.test.yml"

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed or not in PATH"
        exit 1
    fi
}

show_help() {
    cat << EOF
Docker Development Helper for KisanLink E-commerce Service

Usage: $0 [COMMAND] [OPTIONS]

Commands:
  dev-up          Start development environment
  dev-down        Stop development environment
  dev-restart     Restart development environment
  dev-logs        Show development logs
  dev-shell       Open shell in development container

  prod-up         Start production environment
  prod-down       Stop production environment
  prod-logs       Show production logs

  test-run        Run tests in container
  test-up         Start test environment
  test-down       Stop test environment

  db-up           Start only database services
  db-down         Stop database services
  db-reset        Reset database (removes all data)
  db-shell        Open PostgreSQL shell
  db-migrate      Run database migrations

  build           Build all images
  build-dev       Build development image
  build-prod      Build production image
  clean           Clean up containers and images
  status          Show status of all services
  logs            Show logs for all services

  help            Show this help message

Options:
  -f, --follow    Follow logs (for log commands)
  -d, --detach    Run in detached mode
  --no-cache      Build without cache
  --pull          Pull latest base images

Examples:
  $0 dev-up                 # Start development environment
  $0 dev-logs -f            # Follow development logs
  $0 test-run               # Run tests
  $0 db-reset               # Reset database
  $0 clean                  # Clean up everything

EOF
}

# Development environment commands
dev_up() {
    log_info "Starting development environment..."
    docker-compose -f $COMPOSE_FILE -f $DEV_COMPOSE_FILE --profile dev up -d
    log_success "Development environment started!"
    log_info "Services available at:"
    echo "  - API: http://localhost:8080"
    echo "  - API Docs: http://localhost:8080/docs"
    echo "  - pgAdmin: http://localhost:5050"
    echo "  - MailHog: http://localhost:8025"
}

dev_down() {
    log_info "Stopping development environment..."
    docker-compose -f $COMPOSE_FILE -f $DEV_COMPOSE_FILE --profile dev down
    log_success "Development environment stopped!"
}

dev_restart() {
    log_info "Restarting development environment..."
    dev_down
    dev_up
}

dev_logs() {
    local follow_flag=""
    if [[ "$1" == "-f" || "$1" == "--follow" ]]; then
        follow_flag="-f"
    fi
    docker-compose -f $COMPOSE_FILE -f $DEV_COMPOSE_FILE --profile dev logs $follow_flag
}

dev_shell() {
    log_info "Opening shell in development container..."
    docker-compose -f $COMPOSE_FILE -f $DEV_COMPOSE_FILE exec app-dev /bin/bash
}

# Production environment commands
prod_up() {
    log_info "Starting production environment..."
    docker-compose -f $COMPOSE_FILE -f $PROD_COMPOSE_FILE --profile prod up -d
    log_success "Production environment started!"
    log_info "Services available at:"
    echo "  - API: http://localhost:8080"
    echo "  - Load Balancer: http://localhost:80"
}

prod_down() {
    log_info "Stopping production environment..."
    docker-compose -f $COMPOSE_FILE -f $PROD_COMPOSE_FILE --profile prod down
    log_success "Production environment stopped!"
}

prod_logs() {
    local follow_flag=""
    if [[ "$1" == "-f" || "$1" == "--follow" ]]; then
        follow_flag="-f"
    fi
    docker-compose -f $COMPOSE_FILE -f $PROD_COMPOSE_FILE --profile prod logs $follow_flag
}

# Test environment commands
test_run() {
    log_info "Running tests in container..."
    docker-compose -f $COMPOSE_FILE -f $TEST_COMPOSE_FILE --profile test up --abort-on-container-exit
    local exit_code=$?
    docker-compose -f $COMPOSE_FILE -f $TEST_COMPOSE_FILE --profile test down

    if [ $exit_code -eq 0 ]; then
        log_success "All tests passed!"
    else
        log_error "Tests failed with exit code $exit_code"
        exit $exit_code
    fi
}

test_up() {
    log_info "Starting test environment..."
    docker-compose -f $COMPOSE_FILE -f $TEST_COMPOSE_FILE --profile test up -d
    log_success "Test environment started!"
}

test_down() {
    log_info "Stopping test environment..."
    docker-compose -f $COMPOSE_FILE -f $TEST_COMPOSE_FILE --profile test down
    log_success "Test environment stopped!"
}

# Database commands
db_up() {
    log_info "Starting database services..."
    docker-compose -f $COMPOSE_FILE --profile db-only up -d
    log_success "Database services started!"
    log_info "Services available at:"
    echo "  - PostgreSQL: localhost:5432"
    echo "  - Redis: localhost:6379"
}

db_down() {
    log_info "Stopping database services..."
    docker-compose -f $COMPOSE_FILE --profile db-only down
    log_success "Database services stopped!"
}

db_reset() {
    log_warning "This will delete all database data. Are you sure? (y/N)"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        log_info "Resetting database..."
        docker-compose -f $COMPOSE_FILE --profile db-only down -v
        docker-compose -f $COMPOSE_FILE --profile db-only up -d
        log_success "Database reset complete!"
    else
        log_info "Database reset cancelled."
    fi
}

db_shell() {
    log_info "Opening PostgreSQL shell..."
    docker-compose -f $COMPOSE_FILE exec postgres psql -U postgres -d kisanlink_ecom
}

db_migrate() {
    log_info "Running database migrations..."
    if docker-compose -f $COMPOSE_FILE -f $DEV_COMPOSE_FILE ps app-dev | grep -q "Up"; then
        docker-compose -f $COMPOSE_FILE -f $DEV_COMPOSE_FILE exec app-dev make db-migrate
    else
        log_error "Development container is not running. Start it first with: $0 dev-up"
        exit 1
    fi
}

# Build commands
build_all() {
    local no_cache=""
    local pull=""

    if [[ "$1" == "--no-cache" ]]; then
        no_cache="--no-cache"
    fi

    if [[ "$1" == "--pull" || "$2" == "--pull" ]]; then
        pull="--pull"
    fi

    log_info "Building all images..."
    docker-compose -f $COMPOSE_FILE build $no_cache $pull
    log_success "All images built successfully!"
}

build_dev() {
    log_info "Building development image..."
    docker-compose -f $COMPOSE_FILE -f $DEV_COMPOSE_FILE build app-dev
    log_success "Development image built successfully!"
}

build_prod() {
    log_info "Building production image..."
    docker-compose -f $COMPOSE_FILE -f $PROD_COMPOSE_FILE build app-prod
    log_success "Production image built successfully!"
}

# Utility commands
clean() {
    log_warning "This will remove all containers, images, and volumes. Are you sure? (y/N)"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        log_info "Cleaning up..."

        # Stop all services
        docker-compose -f $COMPOSE_FILE -f $DEV_COMPOSE_FILE --profile dev down -v 2>/dev/null || true
        docker-compose -f $COMPOSE_FILE -f $PROD_COMPOSE_FILE --profile prod down -v 2>/dev/null || true
        docker-compose -f $COMPOSE_FILE -f $TEST_COMPOSE_FILE --profile test down -v 2>/dev/null || true

        # Remove project containers
        docker ps -a --filter "name=${PROJECT_NAME}" -q | xargs -r docker rm -f

        # Remove project images
        docker images --filter "reference=${PROJECT_NAME}*" -q | xargs -r docker rmi -f

        # Remove unused volumes
        docker volume ls --filter "name=${PROJECT_NAME}" -q | xargs -r docker volume rm

        # Prune unused resources
        docker system prune -f

        log_success "Cleanup complete!"
    else
        log_info "Cleanup cancelled."
    fi
}

status() {
    log_info "Service status:"
    echo
    docker-compose -f $COMPOSE_FILE ps
}

logs() {
    local follow_flag=""
    if [[ "$1" == "-f" || "$1" == "--follow" ]]; then
        follow_flag="-f"
    fi
    docker-compose -f $COMPOSE_FILE logs $follow_flag
}

# Main script logic
main() {
    check_docker

    case "${1:-help}" in
        dev-up)
            dev_up
            ;;
        dev-down)
            dev_down
            ;;
        dev-restart)
            dev_restart
            ;;
        dev-logs)
            dev_logs "$2"
            ;;
        dev-shell)
            dev_shell
            ;;
        prod-up)
            prod_up
            ;;
        prod-down)
            prod_down
            ;;
        prod-logs)
            prod_logs "$2"
            ;;
        test-run)
            test_run
            ;;
        test-up)
            test_up
            ;;
        test-down)
            test_down
            ;;
        db-up)
            db_up
            ;;
        db-down)
            db_down
            ;;
        db-reset)
            db_reset
            ;;
        db-shell)
            db_shell
            ;;
        db-migrate)
            db_migrate
            ;;
        build)
            build_all "$2" "$3"
            ;;
        build-dev)
            build_dev
            ;;
        build-prod)
            build_prod
            ;;
        clean)
            clean
            ;;
        status)
            status
            ;;
        logs)
            logs "$2"
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "Unknown command: $1"
            echo
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
