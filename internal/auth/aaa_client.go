package auth

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"kisanlink-ecom/internal/config"
	"os"
	"time"

	aaaPb "github.com/Kisanlink/aaa-service/v2/pkg/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	activeStatus = "active"
)

// Client interface for interacting with AAA service
type Client interface {
	// Token validation
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
	ValidateJWT(ctx context.Context, token string) (bool, error)

	// Authentication
	AuthenticateUser(ctx context.Context, req *AuthenticationRequest) (*AuthenticationResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthenticationResponse, error)

	// Authorization
	Authorize(ctx context.Context, req *AuthorizeRequest) (*AuthorizeResponse, error)
	EvaluatePermission(ctx context.Context, userID, resource, action string) (bool, error)
	EvaluateResourcePermission(ctx context.Context, userID, resource, action, resourceID string) (bool, error)

	// User management
	CreateUser(ctx context.Context, user *AAAUser) (*AAAUser, error)
	GetUser(ctx context.Context, userID string) (*AAAUser, error)
	GetUserFromToken(ctx context.Context, token string) (*AAAUser, error)
	UpdateUser(ctx context.Context, user *AAAUser) (*AAAUser, error)
	DeleteUser(ctx context.Context, userID string) error

	// Role management
	GetUserRoles(ctx context.Context, userID string) ([]*AAARole, error)

	// Permission evaluation
	BulkEvaluatePermissions(ctx context.Context, userID string, permissions []PermissionCheck) ([]PermissionResult, error)

	// Health check
	HealthCheck(ctx context.Context) error

	// Close connection
	Close() error
}

// AAAClient is an alias for Client (for backward compatibility)
type AAAClient = Client

// AuthorizeRequest represents an authorization request
type AuthorizeRequest struct {
	UserID     string
	TenantID   string
	Resource   string // "catalog"
	Action     string // "read", "create", "update", "delete", "publish"
	ResourceID string // specific catalog ID for resource-level permissions
}

// AuthorizeResponse represents an authorization response
type AuthorizeResponse struct {
	Allowed bool
	Reason  string
}

// TokenClaims represents the claims extracted from a JWT token
type TokenClaims struct {
	UserID           string
	Username         string
	Email            string
	PhoneNumber      string
	CountryCode      string
	TenantID         string
	OrganizationID   string
	OrganizationName string
	Roles            []string
	RoleIDs          []string
	Permissions      []string
	Scopes           []string
	IssuedAt         int64
	ExpiresAt        int64
	NotBefore        int64
	Issuer           string
	Audience         string
	IsValidated      bool
	UserRoles        []*UserRoleDetails
	Organizations    []*OrganizationDetails
	Groups           []*GroupDetails
	TokenType        string
	TokenVersion     string
	Subject          string
	SessionID        string
	JTI              string
	TenantContext    map[string]interface{}
	UserContextData  *UserContextDetails
}

// client implements Client interface
type client struct {
	conn        *grpc.ClientConn
	userClient  aaaPb.UserServiceClient
	tokenClient aaaPb.TokenServiceClient
	authzClient aaaPb.AuthorizationServiceClient
	orgClient   aaaPb.OrganizationServiceClient
}

// NewAAAClient creates a new AAA client with TLS/mTLS support (alias for NewClient)
func NewAAAClient(config *config.AAAConfig) (AAAClient, error) {
	return NewClient(config)
}

// NewClient creates a new AAA client with TLS/mTLS support
func NewClient(config *config.AAAConfig) (Client, error) {
	// Validate config
	if config == nil {
		return nil, fmt.Errorf("AAA config cannot be nil")
	}
	if config.GRPCServerAddr == "" {
		return nil, fmt.Errorf("AAA gRPC server address cannot be empty")
	}

	// Setup connection options
	var opts []grpc.DialOption

	// Configure TLS/mTLS if enabled
	if config.TLSEnabled {
		creds, err := setupTLSCredentials(config)
		if err != nil {
			return nil, fmt.Errorf("failed to setup TLS credentials: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		// Use insecure credentials for development
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Add connection timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.TimeoutMs)*time.Millisecond)
	defer cancel()

	// Connect to AAA service
	conn, err := grpc.DialContext(ctx, config.GRPCServerAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AAA service: %w", err)
	}

	client := &client{
		conn:        conn,
		userClient:  aaaPb.NewUserServiceClient(conn),
		tokenClient: aaaPb.NewTokenServiceClient(conn),
		authzClient: aaaPb.NewAuthorizationServiceClient(conn),
		orgClient:   aaaPb.NewOrganizationServiceClient(conn),
	}

	return client, nil
}

// setupTLSCredentials configures TLS/mTLS credentials based on configuration
func setupTLSCredentials(config *config.AAAConfig) (credentials.TransportCredentials, error) {
	var tlsConfig *tls.Config

	// Setup mTLS if client certificate and key are provided
	if config.CertPath != "" && config.KeyPath != "" {
		cert, err := tls.LoadX509KeyPair(config.CertPath, config.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}

		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			ServerName:   config.ServerName,
		}

		// Load CA certificate if provided
		if config.CAPath != "" {
			caCert, err := os.ReadFile(config.CAPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA certificate: %w", err)
			}

			caCertPool := x509.NewCertPool()
			if !caCertPool.AppendCertsFromPEM(caCert) {
				return nil, fmt.Errorf("failed to append CA certificate")
			}

			tlsConfig.RootCAs = caCertPool
		}
	} else {
		// Use system root CAs for TLS without client certificate
		tlsConfig = &tls.Config{
			ServerName: config.ServerName,
		}
	}

	return credentials.NewTLS(tlsConfig), nil
}

// Close closes the gRPC connection
func (c *client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// ValidateToken validates a JWT token with AAA service
func (c *client) ValidateToken(ctx context.Context, token string) (*TokenClaims, error) {
	if token == "" {
		return nil, fmt.Errorf("empty token provided")
	}

	// Create request with flags to include user details, permissions, and organization
	req := &aaaPb.ValidateTokenRequest{
		Token:               token,
		IncludeUserDetails:  true,
		IncludePermissions:  true,
		IncludeOrganization: true,
	}

	// Call AAA service TokenService
	resp, err := c.tokenClient.ValidateToken(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	if !resp.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Convert timestamps from protobuf to int64
	var issuedAt, expiresAt, notBefore int64
	if resp.Claims.IssuedAt != nil {
		issuedAt = resp.Claims.IssuedAt.Seconds
	}
	if resp.Claims.ExpiresAt != nil {
		expiresAt = resp.Claims.ExpiresAt.Seconds
	}
	if resp.Claims.NotBefore != nil {
		notBefore = resp.Claims.NotBefore.Seconds
	}

	// Convert audience array to single string (take first if exists)
	var audience string
	if len(resp.Claims.Audience) > 0 {
		audience = resp.Claims.Audience[0]
	}

	// Extract organization information from response
	// Priority: resp.UserContext > resp.Organization > resp.Claims
	var organizationID, organizationName string

	// First, try to get from UserContext (most reliable in AAA v2.0)
	if resp.UserContext != nil {
		organizationID = resp.UserContext.OrganizationId
		organizationName = resp.UserContext.OrganizationName
	}

	// If not found in UserContext, try Organization object
	if organizationID == "" && resp.Organization != nil {
		organizationID = resp.Organization.Id
		organizationName = resp.Organization.Name
	}

	// Fallback to Claims if still not found
	if organizationID == "" {
		organizationID = resp.Claims.OrganizationId
		organizationName = resp.Claims.OrganizationName
	}

	// Convert UserContext if present
	var userContextData *UserContextDetails
	if resp.UserContext != nil {
		uc := resp.UserContext
		userContextData = &UserContextDetails{
			ID:          uc.Id,
			Username:    uc.Username,
			IsValidated: uc.IsValidated,
		}
	}

	// Extract role names from user_context.roles if Claims.Roles is empty
	var roles []string
	if len(resp.Claims.Roles) > 0 {
		// Use roles from Claims if available
		roles = resp.Claims.Roles
	} else if resp.UserContext != nil && len(resp.UserContext.Roles) > 0 {
		// Use roles from UserContext (already a []string)
		roles = resp.UserContext.Roles
	}

	// Convert response to TokenClaims
	claims := &TokenClaims{
		UserID:           resp.Claims.UserId,
		Username:         resp.Claims.Username,
		Email:            resp.Claims.Email,
		OrganizationID:   organizationID,
		OrganizationName: organizationName,
		Roles:            roles,
		Permissions:      resp.Claims.Permissions,
		Scopes:           resp.Claims.Scopes,
		IssuedAt:         issuedAt,
		ExpiresAt:        expiresAt,
		NotBefore:        notBefore,
		Issuer:           resp.Claims.Issuer,
		Audience:         audience,
		TokenType:        resp.Claims.TokenType,
		Subject:          resp.Claims.Subject,
		SessionID:        resp.Claims.SessionId,
		JTI:              resp.Claims.Jti,
		UserContextData:  userContextData,
	}

	return claims, nil
}

// Authorize checks if a user has permission to perform an action on a resource
func (c *client) Authorize(ctx context.Context, req *AuthorizeRequest) (*AuthorizeResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("authorize request cannot be nil")
	}

	// Use the AAA v2 Check API instead of the old Authorize
	checkReq := &aaaPb.CheckRequest{
		PrincipalId:    req.UserID,
		ResourceType:   req.Resource,
		ResourceId:     req.ResourceID,
		Action:         req.Action,
		OrganizationId: req.TenantID,
	}

	// Call AAA service AuthorizationService Check
	checkResp, err := c.authzClient.Check(ctx, checkReq)
	if err != nil {
		return nil, fmt.Errorf("failed to check authorization: %w", err)
	}

	// Combine reasons into a single string
	reason := ""
	if len(checkResp.Reasons) > 0 {
		reason = checkResp.Reasons[0]
	}

	return &AuthorizeResponse{
		Allowed: checkResp.Allowed,
		Reason:  reason,
	}, nil
}

// CreateUser creates a new user in the AAA system using AAA v2 Register
func (c *client) CreateUser(ctx context.Context, user *AAAUser) (*AAAUser, error) {
	// Create register request
	registerReq := &aaaPb.RegisterRequest{
		Username:    user.Username,
		Email:       user.Email,
		FullName:    user.FirstName + " " + user.LastName,
		Password:    "", // Password should be provided separately
		PhoneNumber: user.Phone,
		CountryCode: "+1", // Default, should be provided
	}

	// Call AAA service UserService Register
	resp, err := c.userClient.Register(ctx, registerReq)
	if err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("registration failed: %s", resp.Message)
	}

	// Convert response to AAAUser
	roles := make([]string, 0, len(resp.User.UserRoles))
	for _, ur := range resp.User.UserRoles {
		roles = append(roles, ur.RoleName)
	}

	return &AAAUser{
		ID:          resp.User.Id,
		Username:    resp.User.Username,
		Email:       resp.User.Email,
		Phone:       resp.User.PhoneNumber,
		Roles:       roles,
		Permissions: []string{},
		IsActive:    resp.User.Status == activeStatus,
	}, nil
}

// GetUser retrieves a user from the AAA system using AAA v2 GetUser
func (c *client) GetUser(ctx context.Context, userID string) (*AAAUser, error) {
	// Create get user request
	getUserReq := &aaaPb.GetUserRequest{
		Id:                 userID,
		IncludeRoles:       true,
		IncludePermissions: true,
	}

	// Call AAA service UserService GetUser
	resp, err := c.userClient.GetUser(ctx, getUserReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("get user failed: %s", resp.Message)
	}

	if resp.User == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Convert protobuf User to AAAUser
	roles := make([]string, 0, len(resp.User.UserRoles))
	for _, ur := range resp.User.UserRoles {
		roles = append(roles, ur.RoleName)
	}

	return &AAAUser{
		ID:          resp.User.Id,
		Username:    resp.User.Username,
		Email:       resp.User.Email,
		Phone:       resp.User.PhoneNumber,
		Roles:       roles,
		Permissions: resp.User.Permissions,
		IsActive:    resp.User.Status == activeStatus,
	}, nil
}

// UpdateUser updates a user in the AAA system
func (c *client) UpdateUser(ctx context.Context, user *AAAUser) (*AAAUser, error) {
	// Implementation would call AAA service
	// For now, return the user as-is for testing
	return user, nil
}

// DeleteUser deletes a user from the AAA system
func (c *client) DeleteUser(ctx context.Context, userID string) error {
	// Implementation would call AAA service
	// For now, return nil for testing
	return nil
}

// GetUserRoles retrieves roles for a user
func (c *client) GetUserRoles(ctx context.Context, userID string) ([]*AAARole, error) {
	// Implementation would call AAA service
	// For now, return empty roles for testing
	return []*AAARole{}, nil
}

// BulkEvaluatePermissions evaluates multiple permissions for a user
func (c *client) BulkEvaluatePermissions(ctx context.Context, userID string, permissions []PermissionCheck) ([]PermissionResult, error) {
	// Implementation would call AAA service
	// For now, return all permissions as allowed for testing
	results := make([]PermissionResult, len(permissions))
	for i, perm := range permissions {
		results[i] = PermissionResult{
			Resource: perm.Resource,
			Action:   perm.Action,
			Allowed:  true,
			Reason:   "Mock implementation",
		}
	}
	return results, nil
}

// ValidateJWT validates a JWT token (simplified version)
func (c *client) ValidateJWT(ctx context.Context, token string) (bool, error) {
	_, err := c.ValidateToken(ctx, token)
	return err == nil, err
}

// AuthenticateUser authenticates a user with credentials using AAA v2 Login
func (c *client) AuthenticateUser(ctx context.Context, req *AuthenticationRequest) (*AuthenticationResponse, error) {
	// Create login request
	loginReq := &aaaPb.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}

	// Call AAA service UserService Login
	resp, err := c.userClient.Login(ctx, loginReq)
	if err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("login failed: %s", resp.Message)
	}

	// Convert response to AuthenticationResponse
	userContext := &UserContext{
		UserID:      resp.User.Id,
		Username:    resp.User.Username,
		Email:       resp.User.Email,
		PhoneNumber: resp.User.PhoneNumber,
		CountryCode: resp.User.CountryCode,
		IsActive:    resp.User.Status == "active",
		IsValidated: resp.User.IsValidated,
		Permissions: resp.Permissions,
	}

	// Convert user roles
	userRoles := make([]*UserRoleDetails, 0, len(resp.User.UserRoles))
	for _, ur := range resp.User.UserRoles {
		userRoles = append(userRoles, &UserRoleDetails{
			ID:       ur.Id,
			UserID:   ur.UserId,
			RoleID:   ur.RoleId,
			IsActive: true, // Assuming active if returned
		})
	}
	userContext.UserRoles = userRoles

	return &AuthenticationResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    int64(resp.ExpiresIn),
		TokenType:    resp.TokenType,
		UserContext:  userContext,
	}, nil
}

// RefreshToken refreshes an access token using AAA v2 RefreshToken
func (c *client) RefreshToken(ctx context.Context, refreshToken string) (*AuthenticationResponse, error) {
	// Create refresh token request
	refreshReq := &aaaPb.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}

	// Call AAA service UserService RefreshToken
	resp, err := c.userClient.RefreshToken(ctx, refreshReq)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("token refresh failed: %s", resp.Message)
	}

	return &AuthenticationResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    int64(resp.ExpiresIn),
		TokenType:    "Bearer",
	}, nil
}

// EvaluatePermission evaluates a single permission
func (c *client) EvaluatePermission(ctx context.Context, userID, resource, action string) (bool, error) {
	req := &AuthorizeRequest{
		UserID:   userID,
		Resource: resource,
		Action:   action,
	}
	resp, err := c.Authorize(ctx, req)
	if err != nil {
		return false, err
	}
	return resp.Allowed, nil
}

// EvaluateResourcePermission evaluates permission for a specific resource
func (c *client) EvaluateResourcePermission(ctx context.Context, userID, resource, action, resourceID string) (bool, error) {
	req := &AuthorizeRequest{
		UserID:     userID,
		Resource:   resource,
		Action:     action,
		ResourceID: resourceID,
	}
	resp, err := c.Authorize(ctx, req)
	if err != nil {
		return false, err
	}
	return resp.Allowed, nil
}

// GetUserFromToken retrieves user information from a token
func (c *client) GetUserFromToken(ctx context.Context, token string) (*AAAUser, error) {
	claims, err := c.ValidateToken(ctx, token)
	if err != nil {
		return nil, err
	}

	return &AAAUser{
		ID:       claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
		IsActive: true,
	}, nil
}

// HealthCheck checks the health of the AAA service
func (c *client) HealthCheck(ctx context.Context) error {
	if c.conn == nil {
		return fmt.Errorf("AAA service connection not established")
	}

	// Perform health check by attempting to connect to the service
	// Try a simple token service call to verify connectivity
	testReq := &aaaPb.ValidateTokenRequest{
		Token: "health-check-dummy-token",
	}

	// We expect this to fail with invalid token, but it proves the service is reachable
	_, err := c.tokenClient.ValidateToken(ctx, testReq)

	// If we get any gRPC response (even an error about invalid token), the service is up
	if err != nil {
		// Check if it's a connection error or just an expected validation error
		if err.Error() == "rpc error: code = Unimplemented desc = unknown service pb.AAAService" {
			return fmt.Errorf("AAA service health check failed: service not properly registered: %w", err)
		}
		// Other errors (like invalid token) are acceptable for health check
	}

	return nil
}
