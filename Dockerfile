# Multi-stage Dockerfile for KisanLink E-commerce Service
# Optimized for production with development support

# =============================================================================
# Build Stage
# =============================================================================
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    make \
    ca-certificates \
    tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
ARG BUILD_ENV=production
RUN if [ "$BUILD_ENV" = "development" ]; then \
        make build; \
    else \
        make build-prod; \
    fi

# =============================================================================
# Runtime Stage (Production)
# =============================================================================
FROM alpine:3.19 AS runtime

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    && update-ca-certificates

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary from builder stage
COPY --from=builder /app/kisanlink-ecom /app/kisanlink-ecom

# Copy configuration files
COPY --from=builder /app/docs /app/docs

# Create necessary directories
RUN mkdir -p /app/uploads /app/logs /app/tmp && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Set default environment
ENV GIN_MODE=release \
    PORT=8080 \
    LOG_LEVEL=info \
    LOG_FORMAT=json

# Default command
CMD ["./kisanlink-ecom"]

# =============================================================================
# Development Stage
# =============================================================================
FROM golang:1.24-alpine AS development

# Install development dependencies
RUN apk add --no-cache \
    git \
    make \
    ca-certificates \
    tzdata \
    curl \
    bash \
    vim \
    postgresql-client

# Install development tools
RUN go install github.com/cosmtrek/air@latest && \
    go install github.com/swaggo/swag/cmd/swag@latest && \
    go install github.com/golang/mock/mockgen@latest

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Create necessary directories
RUN mkdir -p /app/uploads /app/logs /app/tmp

# Expose port
EXPOSE 8080

# Set development environment
ENV GIN_MODE=debug \
    PORT=8080 \
    LOG_LEVEL=debug \
    LOG_FORMAT=text

# Default command for development (with hot reload)
CMD ["air", "-c", ".air.toml"]

# =============================================================================
# Testing Stage
# =============================================================================
FROM development AS testing

# Install testing tools
RUN go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run tests
RUN make test

# =============================================================================
# Documentation Stage
# =============================================================================
FROM nginx:alpine AS docs

# Copy documentation
COPY docs/ /usr/share/nginx/html/

# Copy custom nginx config
COPY <<EOF /etc/nginx/conf.d/default.conf
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files \$uri \$uri/ /index.html;
    }

    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }
}
EOF

EXPOSE 80
