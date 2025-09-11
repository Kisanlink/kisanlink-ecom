# AAA Service Integration Testing Guide

This guide provides step-by-step instructions for testing and managing the integration between kisanlink-ecom and your aaa-service.

## 🚀 Quick Start

### 1. Prerequisites

- AAA service running on `localhost:50051` (or your configured address)
- kisanlink-ecom dependencies installed
- Environment variables configured

### 2. Test AAA Service Health

```bash
cd debug
make health-check
```

### 3. Test Full Integration

```bash
cd debug
make test-all
```

## 🔧 Configuration Setup

### Environment Variables

Create or update your `.env` file:

```env
# AAA Service Configuration
AAA_GRPC_SERVER_ADDR=localhost:50051
AAA_ENDPOINT=localhost:50051
AAA_TIMEOUT_MS=800
AAA_RETRIES=2

# Database Configuration (if needed)
DB_PROVIDER=postgres
DB_POSTGRES_HOST=localhost
DB_POSTGRES_PORT=5432
DB_POSTGRES_USER=postgres
DB_POSTGRES_PASSWORD=your_password
DB_POSTGRES_DBNAME=kisanlink_ecom
```

### Verify Configuration

```bash
cd debug
go run -c "package main; import 'kisanlink-ecom/internal/config'; func main() { cfg, _ := config.Load(); println('AAA Server:', cfg.AAA.GRPCServerAddr) }"
```

## 🧪 Testing Workflow

### Phase 1: Basic Connectivity

1. **Health Check**: Verify AAA service is reachable

   ```bash
   make health-check
   ```

2. **Expected Output**:
   ```
   🏥 AAA Service Health Check...
   📋 Checking AAA service at: localhost:50051
   ✅ gRPC connection established
   🔍 Testing service availability...
   ✅ Service is running and responding to gRPC calls
   🎯 Health Check Summary:
      - gRPC connection: ✅
      - Service availability: ✅
      - AAA service is healthy and ready for integration
   ```

### Phase 2: Integration Testing

1. **Test AAA Client**: Verify all methods work

   ```bash
   make test-aaa
   ```

2. **Expected Output**:
   ```
   🔐 Testing AAA Service Integration...
   📋 Configuration loaded:
      AAA Endpoint: localhost:50051
      AAA gRPC Server: localhost:50051
      Timeout: 800ms
      Retries: 2
   🔌 Testing AAA Client Connection...
   ✅ AAA client created successfully
   🧪 Testing AAA Service Functions...
      Testing GetUserRoles...
      Testing GetUserPermissions...
      Testing EvaluatePermission...
      Testing BulkEvaluatePermissions...
   ```

### Phase 3: RBAC Data Seeding

1. **Seed RBAC Data**: Create resources, actions, and permissions

   ```bash
   make seed-rbac
   ```

2. **Expected Output**:

   ```
   🌱 Seeding RBAC Data to AAA Service...
   Seeding E-commerce RBAC data to AAA Service...
   ✓ Connected to AAA service at localhost:50051

   --- Creating Actions ---
   ✓ Created action: create (ID: action_123)
   ✓ Created action: read (ID: action_124)
   ...

   --- Creating Resources ---
   ✓ Created resource: product (ID: resource_123)
   ✓ Created resource: order (ID: resource_124)
   ...
   ```

### Phase 4: End-to-End Testing

1. **Test API Endpoints**: Verify complete flow
   ```bash
   make test-endpoints
   ```

## 🔍 Troubleshooting

### Common Issues & Solutions

#### 1. Connection Refused

```
❌ Failed to connect to AAA service: connection refused
```

**Solutions**:

- Check if AAA service is running: `ps aux | grep aaa-service`
- Verify port: `netstat -an | grep 50051`
- Check firewall: `sudo ufw status`

#### 2. Method Not Implemented

```
⚠️ GetUserRoles failed: rpc error: code = Unimplemented desc = method GetUserRoles not implemented
```

**Solutions**:

- Ensure AAA service implements `EnhancedRBACService`
- Check protobuf definitions match
- Verify service registration

#### 3. Permission Denied

```
⚠️ EvaluatePermission failed: rpc error: code = PermissionDenied
```

**Solutions**:

- Verify RBAC data is seeded
- Check user roles and permissions
- Validate resource/action combinations

#### 4. Timeout Issues

```
⚠️ Service calls timing out
```

**Solutions**:

- Increase timeout: `AAA_TIMEOUT_MS=5000`
- Check network latency
- Verify service performance

### Debug Commands

#### gRPC Testing

```bash
# List available services
grpcurl -plaintext localhost:50051 list

# Test specific method
grpcurl -plaintext -d '{"user_id": "test_user"}' \
  localhost:50051 EnhancedRBACService/GetUserRoles

# Check service health
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check
```

#### Network Testing

```bash
# Test port connectivity
telnet localhost 50051

# Check DNS resolution
nslookup localhost

# Test network path
traceroute localhost
```

#### Log Analysis

```bash
# Check AAA service logs
# (in your aaa-service terminal)

# Check kisanlink-ecom logs
# (in your kisanlink-ecom terminal)
```

## 📊 Success Criteria

### ✅ Integration is Working When:

1. **Health Check Passes**: gRPC connection established
2. **All Methods Respond**: No "unimplemented" errors
3. **RBAC Data Seeded**: Resources, actions, and permissions created
4. **Permission Evaluation Works**: Users can be authorized/denied
5. **End-to-End Flow**: Complete authentication → authorization → access

### ❌ Integration Issues When:

1. **Connection Fails**: Network or service unavailable
2. **Methods Missing**: Service doesn't implement expected interface
3. **Data Missing**: RBAC data not properly seeded
4. **Permission Errors**: Authorization logic not working
5. **Performance Issues**: Timeouts or slow responses

## 🔄 Continuous Testing

### Development Workflow

```bash
# Before making changes
make health-check

# After changes
make test-all

# Quick validation
make test-aaa
```

### CI/CD Integration

```bash
# Add to your CI pipeline
- name: Test AAA Integration
  run: |
    cd debug
    make health-check
    make test-aaa
    make seed-rbac
```

## 📚 Additional Resources

- [AAA Service Documentation](../aaa-service/README.md)
- [gRPC Testing Guide](https://grpc.io/docs/guides/testing/)
- [RBAC Best Practices](https://en.wikipedia.org/wiki/Role-based_access_control)
- [Go Testing Patterns](https://golang.org/doc/code.html#Testing)

## 🆘 Getting Help

If you encounter issues:

1. **Check Logs**: Both AAA service and kisanlink-ecom
2. **Verify Configuration**: Environment variables and service addresses
3. **Test Manually**: Use grpcurl for direct service testing
4. **Check Dependencies**: Ensure all services are running
5. **Review Changes**: Recent modifications that might affect integration
