# AAA Service Integration Testing & Management

This debug folder contains tools to test and manage the integration between kisanlink-ecom and the aaa-service gRPC.

## 🧪 Testing AAA Integration

### 1. Test AAA Connection

```bash
cd debug
go run aaa_test_runner.go
```

This will:

- Test the gRPC connection to your AAA service
- Verify all AAA service methods are working
- Show detailed results for each test

### 2. Seed E-commerce RBAC Data

```bash
cd debug
go run seed_ecommerce_rbac.go
```

This will create the necessary resources, actions, permissions, and roles in your AAA service.

## 🔧 Configuration

Ensure your `.env` file has the correct AAA service configuration:

```env
# AAA Service Configuration
AAA_GRPC_SERVER_ADDR=localhost:50051
AAA_ENDPOINT=localhost:50051
AAA_TIMEOUT_MS=800
AAA_RETRIES=2
```

## 📋 What Gets Seeded

### Actions

- **CRUD Operations**: create, read, update, delete, list
- **Business Operations**: buy, sell, manage_catalog, manage_orders
- **Advanced Operations**: view_analytics, approve_orders, cancel_orders, refund_orders

### Resources

- **Core E-commerce**: product, service, labor, order, customer, seller
- **Management**: organization, catalog, inventory, payment, shipping, analytics

### Roles (to be implemented)

- **Buyer**: Can view catalog, place orders, manage profile
- **Seller**: Can manage catalog, fulfill orders, view analytics
- **Admin**: Full access to all resources
- **Manager**: Limited administrative access

## 🚀 Testing Workflow

1. **Start AAA Service**: Ensure your aaa-service is running
2. **Run Integration Test**: `go run aaa_test_runner.go`
3. **Seed RBAC Data**: `go run seed_ecommerce_rbac.go`
4. **Verify Integration**: Check logs for any errors
5. **Test Real Endpoints**: Use your kisanlink-ecom API endpoints

## 🔍 Troubleshooting

### Common Issues

1. **Connection Failed**
   - Check if AAA service is running
   - Verify gRPC address and port
   - Check network connectivity

2. **Method Not Implemented**
   - Ensure AAA service implements EnhancedRBACService
   - Check protobuf definitions match

3. **Permission Denied**
   - Verify RBAC data is seeded
   - Check user roles and permissions
   - Validate resource/action combinations

### Debug Commands

```bash
# Test gRPC connection
grpcurl -plaintext localhost:50051 list

# Test specific method
grpcurl -plaintext -d '{"user_id": "test_user"}' localhost:50051 EnhancedRBACService/GetUserRoles

# Check AAA service logs
# (in your aaa-service terminal)
```

## 📊 Integration Status

Your current integration has:

- ✅ AAA client implementation
- ✅ Configuration management
- ✅ Middleware integration
- ✅ Permission evaluation
- ⚠️ Some mock implementations (need real AAA calls)
- ⚠️ JWT token validation (needs AAA integration)

## 🔄 Next Steps

1. **Replace Mock Implementations**: Update auth middleware to use real AAA calls
2. **Implement JWT Validation**: Integrate with AAA service for token validation
3. **Add Role-Based Access**: Use seeded roles for different user types
4. **Test End-to-End**: Verify complete authentication/authorization flow

## 📚 Related Files

- `internal/auth/aaa_client.go` - AAA service client
- `internal/middleware/authn.go` - Authentication middleware
- `internal/middleware/authz.go` - Authorization middleware
- `internal/config/config.go` - Configuration management
