# Deployment Documentation

## Overview

The KisanLink E-commerce Service supports multiple deployment strategies using Docker containers. This document covers deployment configurations for different environments.

## Prerequisites

### System Requirements

- **Docker**: Version 20.10+
- **Docker Compose**: Version 2.0+
- **Memory**: Minimum 2GB RAM, recommended 4GB+
- **Storage**: At least 10GB free space
- **Network**: Stable internet connection

### Environment Variables

Create environment-specific `.env` files:

```bash
# .env.development
GIN_MODE=debug
LOG_LEVEL=debug
DB_PASSWORD=dev_password_123
JWT_SECRET=dev_jwt_secret_change_in_production

# .env.staging
GIN_MODE=release
LOG_LEVEL=info
DB_PASSWORD=staging_secure_password
JWT_SECRET=staging_jwt_secret

# .env.production
GIN_MODE=release
LOG_LEVEL=warn
DB_PASSWORD=production_secure_password
JWT_SECRET=production_jwt_secret_very_secure
```

## Development Deployment

### Quick Start

```bash
# Start development environment
docker-compose --profile dev up -d

# Or use the helper script
./scripts/docker-dev.sh dev-up
```

### Services Included

- **Application**: Hot-reload enabled development server
- **PostgreSQL**: Database with development data
- **Redis**: Cache server
- **pgAdmin**: Database management UI
- **MailHog**: Email testing server

### Access Points

- **API**: http://localhost:8080
- **API Documentation**: http://localhost:8080/docs
- **pgAdmin**: http://localhost:5050 (admin@kisanlink.local / admin123)
- **MailHog**: http://localhost:8025

### Development Workflow

```bash
# Start services
docker-compose --profile dev up -d

# View logs
docker-compose --profile dev logs -f app-dev

# Run tests
docker-compose --profile dev exec app-dev make test

# Access application shell
docker-compose --profile dev exec app-dev /bin/bash

# Stop services
docker-compose --profile dev down
```

## Staging Deployment

### Configuration

```bash
# Use staging compose file
docker-compose -f docker-compose.yml -f docker-compose.prod.yml --profile prod up -d
```

### Environment Setup

```bash
# Copy staging environment
cp .env.staging .env

# Update configuration
vim .env

# Start staging environment
docker-compose --profile prod up -d
```

### Health Checks

```bash
# Check service health
curl http://localhost:8080/health

# Check database connectivity
docker-compose exec postgres pg_isready -U postgres

# View application logs
docker-compose logs app-prod
```

## Production Deployment

### Security Considerations

1. **Environment Variables**
   - Use secure passwords
   - Generate strong JWT secrets
   - Configure proper database credentials

2. **Network Security**
   - Use HTTPS in production
   - Configure firewall rules
   - Limit database access

3. **Container Security**
   - Run containers as non-root user
   - Use minimal base images
   - Regular security updates

### Production Configuration

```yaml
# docker-compose.prod.yml
version: "3.8"
services:
  app-prod:
    deploy:
      replicas: 2
      resources:
        limits:
          cpus: "1.0"
          memory: 512M
      restart_policy:
        condition: on-failure
```

### SSL/TLS Setup

```bash
# Create SSL directory
mkdir -p ssl

# Copy SSL certificates
cp your-cert.pem ssl/
cp your-key.pem ssl/

# Update nginx configuration
vim scripts/nginx-prod.conf
```

### Production Deployment Steps

```bash
# 1. Prepare environment
cp .env.production .env
vim .env  # Update with production values

# 2. Build production images
docker-compose -f docker-compose.yml -f docker-compose.prod.yml build

# 3. Start production services
docker-compose -f docker-compose.yml -f docker-compose.prod.yml --profile prod up -d

# 4. Run database migrations
docker-compose exec app-prod make db-migrate

# 5. Verify deployment
curl https://your-domain.com/health
```

## Container Orchestration

### Docker Swarm

```bash
# Initialize swarm
docker swarm init

# Deploy stack
docker stack deploy -c docker-compose.yml -c docker-compose.prod.yml kisanlink

# Scale services
docker service scale kisanlink_app-prod=3

# Update service
docker service update --image kisanlink-ecom:latest kisanlink_app-prod
```

### Kubernetes

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kisanlink-ecom
spec:
  replicas: 3
  selector:
    matchLabels:
      app: kisanlink-ecom
  template:
    metadata:
      labels:
        app: kisanlink-ecom
    spec:
      containers:
        - name: app
          image: kisanlink-ecom:latest
          ports:
            - containerPort: 8080
          env:
            - name: GIN_MODE
              value: "release"
            - name: DB_HOST
              value: "postgres-service"
```

## Monitoring and Logging

### Health Checks

```bash
# Application health
curl http://localhost:8080/health

# Database health
docker-compose exec postgres pg_isready

# Redis health
docker-compose exec redis redis-cli ping
```

### Logging Configuration

```yaml
# docker-compose.yml
services:
  app-prod:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

### Log Aggregation

```bash
# View all logs
docker-compose logs

# Follow specific service logs
docker-compose logs -f app-prod

# Export logs
docker-compose logs --no-color > application.log
```

## Backup and Recovery

### Database Backup

```bash
# Create backup
docker-compose exec postgres pg_dump -U postgres kisanlink_ecom > backup.sql

# Automated backup script
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
docker-compose exec postgres pg_dump -U postgres kisanlink_ecom > "backup_${DATE}.sql"
```

### Volume Backup

```bash
# Backup volumes
docker run --rm -v kisanlink-ecom_postgres_data:/data -v $(pwd):/backup alpine tar czf /backup/postgres_backup.tar.gz -C /data .

# Restore volumes
docker run --rm -v kisanlink-ecom_postgres_data:/data -v $(pwd):/backup alpine tar xzf /backup/postgres_backup.tar.gz -C /data
```

## Performance Optimization

### Resource Limits

```yaml
services:
  app-prod:
    deploy:
      resources:
        limits:
          cpus: "1.0"
          memory: 512M
        reservations:
          cpus: "0.5"
          memory: 256M
```

### Database Optimization

```sql
-- PostgreSQL configuration
shared_buffers = 256MB
effective_cache_size = 1GB
maintenance_work_mem = 64MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
```

### Caching Strategy

```yaml
services:
  redis:
    command: redis-server --maxmemory 256mb --maxmemory-policy allkeys-lru
```

## Troubleshooting

### Common Issues

1. **Container Won't Start**

   ```bash
   # Check logs
   docker-compose logs app-prod

   # Check resource usage
   docker stats

   # Verify configuration
   docker-compose config
   ```

2. **Database Connection Issues**

   ```bash
   # Test database connectivity
   docker-compose exec app-prod nc -zv postgres 5432

   # Check database logs
   docker-compose logs postgres

   # Verify credentials
   docker-compose exec postgres psql -U postgres -l
   ```

3. **Performance Issues**

   ```bash
   # Monitor resource usage
   docker stats

   # Check application metrics
   curl http://localhost:8080/metrics

   # Analyze slow queries
   docker-compose exec postgres psql -U postgres -c "SELECT * FROM pg_stat_activity;"
   ```

### Debug Mode

```bash
# Enable debug logging
docker-compose exec app-prod sh -c 'export LOG_LEVEL=debug && ./kisanlink-ecom'

# Access container shell
docker-compose exec app-prod /bin/sh

# Run diagnostics
docker-compose exec app-prod make test
```

## Scaling

### Horizontal Scaling

```bash
# Scale application containers
docker-compose up -d --scale app-prod=3

# Use load balancer
docker-compose --profile prod up -d nginx
```

### Vertical Scaling

```yaml
services:
  app-prod:
    deploy:
      resources:
        limits:
          cpus: "2.0"
          memory: 1G
```

## Security Best Practices

### Container Security

1. **Use non-root user**
2. **Minimal base images**
3. **Regular updates**
4. **Secret management**
5. **Network isolation**

### Application Security

1. **Environment variables for secrets**
2. **HTTPS in production**
3. **Input validation**
4. **Rate limiting**
5. **Security headers**

## Maintenance

### Updates

```bash
# Pull latest images
docker-compose pull

# Rebuild and restart
docker-compose up -d --build

# Rolling update
docker-compose up -d --no-deps app-prod
```

### Cleanup

```bash
# Remove unused containers
docker container prune

# Remove unused images
docker image prune

# Remove unused volumes
docker volume prune

# Complete cleanup
docker system prune -a
```

## Support

For deployment issues:

1. Check logs: `docker-compose logs`
2. Verify configuration: `docker-compose config`
3. Test connectivity: Health check endpoints
4. Review resource usage: `docker stats`
5. Contact development team with logs and configuration

## Appendix

### Environment Variables Reference

| Variable      | Description        | Default    | Required |
| ------------- | ------------------ | ---------- | -------- |
| `GIN_MODE`    | Gin framework mode | `release`  | No       |
| `PORT`        | Application port   | `8080`     | No       |
| `DB_HOST`     | Database host      | `postgres` | Yes      |
| `DB_PASSWORD` | Database password  | -          | Yes      |
| `JWT_SECRET`  | JWT signing secret | -          | Yes      |
| `LOG_LEVEL`   | Logging level      | `info`     | No       |

### Port Reference

| Service     | Port   | Description       |
| ----------- | ------ | ----------------- |
| Application | 8080   | Main API server   |
| PostgreSQL  | 5432   | Database server   |
| Redis       | 6379   | Cache server      |
| pgAdmin     | 5050   | Database admin UI |
| Nginx       | 80/443 | Load balancer     |

### Volume Reference

| Volume          | Purpose          | Backup Required |
| --------------- | ---------------- | --------------- |
| `postgres_data` | Database data    | Yes             |
| `redis_data`    | Cache data       | No              |
| `app_uploads`   | File uploads     | Yes             |
| `app_logs`      | Application logs | No              |
