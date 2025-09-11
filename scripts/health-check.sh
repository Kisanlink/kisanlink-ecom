#!/bin/sh

# Health check script for KisanLink E-commerce Service
# This script is used by Docker health checks and monitoring systems

set -e

# Configuration
HOST=${HOST:-localhost}
PORT=${PORT:-8080}
TIMEOUT=${TIMEOUT:-10}
HEALTH_ENDPOINT="/health"

# Colors for output (if terminal supports it)
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    NC='\033[0m'
else
    RED=''
    GREEN=''
    YELLOW=''
    NC=''
fi

# Logging functions
log_info() {
    echo "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo "${RED}[ERROR]${NC} $1"
}

# Check if curl is available
if ! command -v curl >/dev/null 2>&1; then
    log_error "curl is not available"
    exit 1
fi

# Perform health check
log_info "Checking health of service at ${HOST}:${PORT}${HEALTH_ENDPOINT}"

# Make the health check request
response=$(curl -s -w "%{http_code}" --max-time "$TIMEOUT" "http://${HOST}:${PORT}${HEALTH_ENDPOINT}" || echo "000")

# Extract HTTP status code (last 3 characters)
http_code=$(echo "$response" | tail -c 4)

# Extract response body (everything except last 3 characters)
response_body=$(echo "$response" | head -c -4)

# Check the response
case "$http_code" in
    200)
        log_info "Health check passed (HTTP $http_code)"
        echo "$response_body"
        exit 0
        ;;
    000)
        log_error "Failed to connect to service"
        exit 1
        ;;
    *)
        log_error "Health check failed (HTTP $http_code)"
        echo "$response_body"
        exit 1
        ;;
esac
