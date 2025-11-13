# AAA Service Integration - Complete Guide

## 🎯 Overview

This guide provides everything you need to test, manage, and troubleshoot the integration between kisanlink-ecom and your aaa-service gRPC.

## 🚀 Quick Start Commands

### 1. Check Current Status

```bash
cd debug
./run_tests.sh status
```

### 2. Test AAA Service Health

```bash
cd debug
./run_tests.sh health
```

### 3. Run Full Integration Test

```bash
cd debug
./run_tests.sh all
```

### 4. Interactive Menu

```bash
cd debug
./run_tests.sh menu
```

## 📋 What You Can Test

### ✅ **Integration Status Check** (`check_integration_status.go`)

- Configuration validation
- gRPC connectivity
- AAA client functionality
- Middleware integration
- Identifies areas needing attention

### ✅ **Health Check** (`health_check.go`)

- Basic gRPC connection
- Service availability
- Quick connectivity test

### ✅ **AAA Integration Test** (`aaa_test_runner.go`)

- Full AAA client testing
- All method implementations
- Permission evaluation
- Bulk operations

### ✅ **RBAC Data Seeding** (`seed_ecommerce_rbac.go`)

- Creates resources (products, orders, etc.)
- Creates actions (CRUD operations, business logic)
- Sets up permission structure
- Prepares role-based access control

### ✅ **API Endpoints Test** (`test_ecommerce_endpoints.go`)

- Tests actual API endpoints
- Verifies authentication flow
- Checks authorization logic

## 🔧 Configuration Requirements

### Environment Variables

```env
# AAA Service Configuration
AAA_GRPC_SERVER_ADDR=localhost:50051
AAA_ENDPOINT=localhost:50051
AAA_TIMEOUT_MS=800
AAA_RETRIES=2

# Database (if needed)
DB_PROVIDER=postgres
DB_POSTGRES_HOST=localhost
DB_POSTGRES_PORT=5432
DB_POSTGRES_USER=postgres
DB_POSTGRES_PASSWORD=your_password
DB_POSTGRES_DBNAME=kisanlink_ecom
```

### Prerequisites

1. **AAA Service Running**: On configured gRPC address
2. **Dependencies Installed**: Go modules and protobuf
3. **Network Access**: Firewall allows gRPC communication

## 🧪 Testing Workflow

### Phase 1: Basic Connectivity

```bash
./run_tests.sh health
```

**Expected**: ✅ gRPC connection established, service responding

### Phase 2: Integration Status

```bash
./run_tests.sh status
```

**Expected**: ✅ All basic components working, identifies mock implementations

### Phase 3: Full Integration Test

```bash
./run_tests.sh test
```

**Expected**: ✅ All AAA methods responding correctly

### Phase 4: RBAC Data Setup

```bash
./run_tests.sh seed
```

**Expected**: ✅ Resources, actions, and permissions created

### Phase 5: End-to-End Testing

```bash
./run_tests.sh endpoints
```

**Expected**: ✅ Complete authentication → authorization → access flow

## 🔍 Troubleshooting Guide

### Issue: Connection Refused

```
❌ Failed to connect to AAA service: connection refused
```

**Solutions**:

1. Check if AAA service is running
2. Verify port configuration
3. Check firewall settings
4. Test with: `telnet localhost 50051`

### Issue: Method Not Implemented

```
⚠️ GetUserRoles failed: rpc error: code = Unimplemented
```

**Solutions**:

1. Ensure AAA service implements `EnhancedRBACService`
2. Check protobuf definitions match
3. Verify service registration

### Issue: Permission Denied

```
⚠️ EvaluatePermission failed: rpc error: code = PermissionDenied
```

**Solutions**:

1. Run RBAC seeding: `./run_tests.sh seed`
2. Check user roles and permissions
3. Validate resource/action combinations

### Issue: Timeout Issues

```
⚠️ Service calls timing out
```

**Solutions**:

1. Increase timeout: `AAA_TIMEOUT_MS=5000`
2. Check network latency
3. Verify service performance

## 📊 Current Integration Status

### ✅ **Working Components**

- AAA client implementation
- Configuration management
- gRPC connection handling
- Middleware integration
- Permission evaluation framework

### ⚠️ **Areas Needing Attention**

- JWT token validation (currently mocked)
- User authentication (currently mocked)
- Role assignment (needs AAA service implementation)
- Permission seeding (needs RBAC data)

### 🔄 **Next Steps**

1. **Seed RBAC Data**: `./run_tests.sh seed`
2. **Update Auth Middleware**: Replace mock implementations
3. **Implement JWT Validation**: Integrate with AAA service
4. **Test End-to-End Flow**: Verify complete authentication
5. **Add Role-Based Access**: Use seeded roles

## 🛠️ Development Workflow

### Daily Development

```bash
# Start development session
./run_tests.sh status

# Make changes to code

# Test changes
./run_tests.sh test

# Verify endpoints
./run_tests.sh endpoints
```

### Before Committing

```bash
# Run all tests
./run_tests.sh all

# Check integration status
./run_tests.sh status
```

### CI/CD Integration

```bash
# Add to your CI pipeline
- name: Test AAA Integration
  run: |
    cd debug
    ./run_tests.sh health
    ./run_tests.sh test
    ./run_tests.sh seed
```

## 📚 File Structure

```
debug/
├── README.md                    # Basic usage instructions
├── TESTING_GUIDE.md            # Detailed testing guide
├── INTEGRATION_SUMMARY.md      # This file - complete overview
├── Makefile                    # Make-based testing commands
├── run_tests.sh               # Shell script for running tests
├── check_integration_status.go # Integration status checker
├── health_check.go            # Basic health check
├── aaa_test_runner.go         # Full integration test
├── seed_ecommerce_rbac.go     # RBAC data seeder
└── test_ecommerce_endpoints.go # API endpoint tester
```

## 🎯 Success Criteria

### ✅ **Integration is Working When**

1. Health check passes
2. All AAA methods respond
3. RBAC data is seeded
4. Permission evaluation works
5. End-to-end flow completes

### ❌ **Integration Issues When**

1. Connection fails
2. Methods are missing
3. Data is not seeded
4. Permissions don't work
5. Performance is poor

## 🆘 Getting Help

### Quick Diagnostics

```bash
# Check current status
./run_tests.sh status

# Run health check
./run_tests.sh health

# Test specific component
./run_tests.sh test
```

### Manual Testing

```bash
# Test gRPC directly
grpcurl -plaintext localhost:50051 list

# Test specific method
grpcurl -plaintext -d '{"user_id": "test"}' \
  localhost:50051 EnhancedRBACService/GetUserRoles
```

### Common Commands

```bash
# Run all tests
./run_tests.sh all

# Interactive menu
./run_tests.sh menu

# Check help
./run_tests.sh
```

## 🚀 Ready to Start?

1. **Ensure AAA service is running**
2. **Check your configuration**
3. **Run status check**: `./run_tests.sh status`
4. **Follow the workflow above**
5. **Seed RBAC data**: `./run_tests.sh seed`
6. **Test integration**: `./run_tests.sh test`

Your kisanlink-ecom is ready for full AAA service integration! 🎉
