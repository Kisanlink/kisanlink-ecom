# KisanLink E-commerce API - Operational Runbook

## Overview

This runbook provides operational procedures, troubleshooting guides, and escalation paths for the KisanLink E-commerce API service.

## Service Information

- **Service Name**: KisanLink E-commerce API
- **Version**: 1.0.0
- **Technology Stack**: Go, Gin, PostgreSQL, DynamoDB
- **Dependencies**: AAA Service (gRPC), kisanlink-db package

## Health Check Endpoints

### Basic Health Check

- **Endpoint**: `GET /health`
- **Purpose**: Simple liveness check
- **Expected Response**: 200 OK with basic status

### Detailed Health Check

- **Endpoint**: `GET /health/detailed`
- **Purpose**: Comprehensive health check including dependencies
- **Expected Response**: 200 OK if healthy, 503 if degraded

### Readiness Check

- **Endpoint**: `GET /ready`
- **Purpose**: Kubernetes readiness probe
- **Expected Response**: 200 OK if ready to accept traffic

### Metrics Endpoint

- **Endpoint**: `GET /metrics`
- **Purpose**: System metrics in JSON format
- **Expected Response**: 200 OK with metrics data

### Prometheus Metrics

- **Endpoint**: `GET /metrics/prometheus`
- **Purpose**: Prometheus-compatible metrics for scraping
- **Expected Response**: 200 OK with Prometheus format metrics

### Status Dashboard

- **Endpoint**: `GET /status`
- **Purpose**: Comprehensive status dashboard
- **Expected Response**: 200 OK with dashboard data

## Monitoring and Alerting

### Key Metrics to Monitor

1. **HTTP Metrics**
   - Request rate (requests/second)
   - Error rate (percentage)
   - Response time (95th percentile)
   - Status code distribution

2. **System Metrics**
   - Memory usage
   - CPU usage
   - Goroutine count
   - Garbage collection frequency

3. **Database Metrics**
   - Connection pool usage
   - Query latency
   - Error rate
   - Connection failures

4. **Business Metrics**
   - Order creation rate
   - Inventory operations
   - Authentication failures

### Alert Thresholds

| Metric           | Warning | Critical | Action                         |
| ---------------- | ------- | -------- | ------------------------------ |
| Error Rate       | > 5%    | > 10%    | Check logs, investigate errors |
| Response Time    | > 1s    | > 3s     | Check database performance     |
| Memory Usage     | > 80%   | > 95%    | Check for memory leaks         |
| Database Latency | > 500ms | > 2s     | Check database health          |
| Goroutines       | > 1000  | > 5000   | Check for goroutine leaks      |

## Troubleshooting Procedures

### High Error Rate

1. **Check Error Logs**

   ```bash
   # Check recent error logs
   kubectl logs -f deployment/kisanlink-ecom --since=10m | grep ERROR
   ```

2. **Check Error Distribution**
   - Access `/metrics` endpoint
   - Look at HTTP status code distribution
   - Identify most common error types

3. **Common Causes**
   - Database connectivity issues
   - AAA service unavailable
   - Invalid input validation
   - Business rule violations

4. **Resolution Steps**
   - Check database health via `/health/detailed`
   - Verify AAA service connectivity
   - Review recent deployments
   - Check for configuration changes

### High Response Time

1. **Check Database Performance**

   ```bash
   # Check database metrics
   curl http://localhost:8080/metrics | grep database_latency
   ```

2. **Identify Slow Endpoints**
   - Review application logs
   - Check metrics for endpoint-specific latency
   - Look for database query patterns

3. **Common Causes**
   - Slow database queries
   - High database load
   - Network latency
   - Resource contention

4. **Resolution Steps**
   - Check database connection pool
   - Review query performance
   - Scale database resources if needed
   - Optimize slow queries

### Memory Issues

1. **Check Memory Usage**

   ```bash
   # Check current memory usage
   curl http://localhost:8080/metrics | grep memory
   ```

2. **Identify Memory Leaks**
   - Monitor memory growth over time
   - Check goroutine count
   - Review garbage collection metrics

3. **Common Causes**
   - Goroutine leaks
   - Large object retention
   - Inefficient data structures
   - Connection leaks

4. **Resolution Steps**
   - Restart service if critical
   - Review recent code changes
   - Check for connection leaks
   - Monitor garbage collection

### Database Connectivity Issues

1. **Check Database Health**

   ```bash
   # Check detailed health
   curl http://localhost:8080/health/detailed
   ```

2. **Verify Database Status**
   - Check PostgreSQL connectivity
   - Verify DynamoDB access
   - Test connection pooling

3. **Common Causes**
   - Network connectivity
   - Authentication failures
   - Connection pool exhaustion
   - Database server issues

4. **Resolution Steps**
   - Check database server status
   - Verify connection strings
   - Review authentication credentials
   - Check network connectivity

### Service Unavailable (503 Errors)

1. **Check Service Status**

   ```bash
   # Check if service is running
   kubectl get pods -l app=kisanlink-ecom
   ```

2. **Check Dependencies**
   - Database connectivity
   - AAA service availability
   - External service dependencies

3. **Common Causes**
   - Service not started
   - Failed health checks
   - Dependency failures
   - Resource exhaustion

4. **Resolution Steps**
   - Check service logs
   - Verify dependencies
   - Check resource limits
   - Restart service if needed

## Escalation Procedures

### Severity Levels

#### P1 - Critical (Service Down)

- **Response Time**: 15 minutes
- **Examples**: Service completely unavailable, data corruption
- **Escalation**: Immediate notification to on-call engineer and team lead

#### P2 - High (Degraded Performance)

- **Response Time**: 1 hour
- **Examples**: High error rate, slow response times
- **Escalation**: Notification to on-call engineer

#### P3 - Medium (Minor Issues)

- **Response Time**: 4 hours
- **Examples**: Non-critical feature failures, minor performance issues
- **Escalation**: Standard ticket assignment

#### P4 - Low (Enhancement/Maintenance)

- **Response Time**: Next business day
- **Examples**: Feature requests, minor improvements
- **Escalation**: Standard development process

### Contact Information

#### On-Call Engineer

- **Primary**: [On-call rotation system]
- **Backup**: [Backup engineer contact]

#### Team Lead

- **Name**: [Team Lead Name]
- **Contact**: [Contact Information]

#### Database Administrator

- **Name**: [DBA Name]
- **Contact**: [Contact Information]

#### Infrastructure Team

- **Contact**: [Infrastructure team contact]

## Recovery Procedures

### Service Restart

```bash
# Kubernetes deployment restart
kubectl rollout restart deployment/kisanlink-ecom

# Check rollout status
kubectl rollout status deployment/kisanlink-ecom
```

### Database Recovery

```bash
# Check database connectivity
kubectl exec -it deployment/kisanlink-ecom -- /app/kisanlink-ecom health-check

# Database migration (if needed)
kubectl exec -it deployment/kisanlink-ecom -- /app/kisanlink-ecom migrate
```

### Rollback Procedures

```bash
# Rollback to previous version
kubectl rollout undo deployment/kisanlink-ecom

# Rollback to specific revision
kubectl rollout undo deployment/kisanlink-ecom --to-revision=2
```

## Maintenance Procedures

### Planned Maintenance

1. **Pre-maintenance Checklist**
   - [ ] Notify stakeholders
   - [ ] Schedule maintenance window
   - [ ] Prepare rollback plan
   - [ ] Backup critical data

2. **During Maintenance**
   - [ ] Monitor service health
   - [ ] Check error rates
   - [ ] Verify functionality
   - [ ] Document any issues

3. **Post-maintenance Checklist**
   - [ ] Verify service health
   - [ ] Check all endpoints
   - [ ] Monitor for 30 minutes
   - [ ] Notify completion

### Database Maintenance

```bash
# Database backup
kubectl exec -it postgres-pod -- pg_dump kisanlink_ecom > backup.sql

# Database optimization
kubectl exec -it postgres-pod -- psql -d kisanlink_ecom -c "VACUUM ANALYZE;"
```

## Performance Tuning

### Database Optimization

1. **Query Performance**
   - Monitor slow queries
   - Add appropriate indexes
   - Optimize query patterns

2. **Connection Pooling**
   - Adjust pool size based on load
   - Monitor connection usage
   - Tune timeout settings

### Application Optimization

1. **Memory Management**
   - Monitor garbage collection
   - Optimize data structures
   - Reduce object allocations

2. **Concurrency**
   - Monitor goroutine usage
   - Optimize concurrent operations
   - Avoid goroutine leaks

## Security Procedures

### Security Incident Response

1. **Immediate Actions**
   - Assess impact and scope
   - Contain the incident
   - Preserve evidence

2. **Investigation**
   - Review access logs
   - Check authentication failures
   - Analyze traffic patterns

3. **Recovery**
   - Apply security patches
   - Update credentials if needed
   - Monitor for further issues

### Regular Security Tasks

- [ ] Review access logs weekly
- [ ] Update dependencies monthly
- [ ] Security scan quarterly
- [ ] Penetration testing annually

## Useful Commands

### Service Management

```bash
# Check service status
kubectl get pods -l app=kisanlink-ecom

# View service logs
kubectl logs -f deployment/kisanlink-ecom

# Scale service
kubectl scale deployment/kisanlink-ecom --replicas=3
```

### Debugging

```bash
# Port forward for local access
kubectl port-forward deployment/kisanlink-ecom 8080:8080

# Execute commands in pod
kubectl exec -it deployment/kisanlink-ecom -- /bin/sh

# Check resource usage
kubectl top pods -l app=kisanlink-ecom
```

### Monitoring

```bash
# Check metrics
curl http://localhost:8080/metrics

# Check health
curl http://localhost:8080/health/detailed

# Check status dashboard
curl http://localhost:8080/status
```

## Documentation Links

- [API Documentation](./api-documentation.md)
- [Deployment Guide](./deployment-guide.md)
- [Configuration Reference](./configuration-reference.md)
- [Architecture Overview](./architecture-overview.md)

## Change Log

| Date       | Version | Changes                  | Author           |
| ---------- | ------- | ------------------------ | ---------------- |
| 2024-01-15 | 1.0.0   | Initial runbook creation | Development Team |

---

**Note**: This runbook should be updated regularly to reflect changes in the system architecture, procedures, and contact information.
