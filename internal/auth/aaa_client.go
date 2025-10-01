package auth

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"kisanlink-ecom/internal/config"
	"os"
	"time"

	aaaPb "kisanlink-ecom/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
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
	authClient  aaaPb.AuthServiceClient
	authzClient aaaPb.AuthorizationServiceClient
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
		authClient:  aaaPb.NewAuthServiceClient(conn),
		authzClient: aaaPb.NewAuthorizationServiceClient(conn),
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

	// Create request
	req := &aaaPb.ValidateTokenRequest{
		Token: token,
	}

	// Call AAA service
	resp, err := c.authClient.ValidateToken(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	if !resp.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Convert protobuf UserRoles to internal UserRoleDetails
	var userRoles []*UserRoleDetails
	for _, ur := range resp.Claims.UserRoles {
		userRole := &UserRoleDetails{
			ID:       ur.Id,
			UserID:   ur.UserId,
			RoleID:   ur.RoleId,
			IsActive: ur.IsActive,
		}

		// Convert nested role if present
		if ur.Role != nil {
			userRole.Role = &RoleDetails{
				ID:          ur.Role.Id,
				Name:        ur.Role.Name,
				Scope:       ur.Role.Scope,
				Description: ur.Role.Description,
				ParentID:    ur.Role.ParentId,
				IsActive:    ur.Role.IsActive,
				Permissions: ur.Role.Permissions,
			}
		}

		userRoles = append(userRoles, userRole)
	}

	// Convert protobuf Organizations to internal OrganizationDetails
	var organizations []*OrganizationDetails
	for _, org := range resp.Claims.Organizations {
		organizations = append(organizations, &OrganizationDetails{
			ID:       org.Id,
			Name:     org.Name,
			TenantID: org.TenantId,
		})
	}

	// Convert protobuf Groups to internal GroupDetails
	var groups []*GroupDetails
	for _, grp := range resp.Claims.Groups {
		groups = append(groups, &GroupDetails{
			ID:             grp.Id,
			Name:           grp.Name,
			OrganizationID: grp.OrganizationId,
			Description:    grp.Description,
		})
	}

	// Convert UserContext if present
	var userContextData *UserContextDetails
	if resp.Claims.UserContext != nil {
		uc := resp.Claims.UserContext

		// Convert roles from UserContext
		var ucRoles []*RoleDetails
		for _, r := range uc.Roles {
			ucRoles = append(ucRoles, &RoleDetails{
				ID:          r.Id,
				Name:        r.Name,
				Scope:       r.Scope,
				Description: r.Description,
				ParentID:    r.ParentId,
				IsActive:    r.IsActive,
				Permissions: r.Permissions,
			})
		}

		// Convert organizations from UserContext
		var ucOrgs []*OrganizationDetails
		for _, o := range uc.Organizations {
			ucOrgs = append(ucOrgs, &OrganizationDetails{
				ID:       o.Id,
				Name:     o.Name,
				TenantID: o.TenantId,
			})
		}

		// Convert groups from UserContext
		var ucGroups []*GroupDetails
		for _, g := range uc.Groups {
			ucGroups = append(ucGroups, &GroupDetails{
				ID:             g.Id,
				Name:           g.Name,
				OrganizationID: g.OrganizationId,
				Description:    g.Description,
			})
		}

		userContextData = &UserContextDetails{
			ID:            uc.Id,
			Username:      uc.Username,
			PhoneNumber:   uc.PhoneNumber,
			CountryCode:   uc.CountryCode,
			IsValidated:   uc.IsValidated,
			Roles:         ucRoles,
			Organizations: ucOrgs,
			Groups:        ucGroups,
		}
	}

	// Parse tenant_context if it's a JSON string
	var tenantContext map[string]interface{}
	if resp.Claims.TenantContext != "" {
		// For now, we'll store it as a map with the raw JSON string
		// In production, you'd want to properly unmarshal this
		tenantContext = make(map[string]interface{})
	}

	// Convert response to TokenClaims
	claims := &TokenClaims{
		UserID:           resp.Claims.UserId,
		Username:         resp.Claims.Username,
		Email:            resp.Claims.Email,
		PhoneNumber:      resp.Claims.PhoneNumber,
		CountryCode:      resp.Claims.CountryCode,
		TenantID:         resp.Claims.TenantId,
		OrganizationID:   resp.Claims.OrganizationId,
		OrganizationName: resp.Claims.OrganizationName,
		Roles:            resp.Claims.Roles,
		RoleIDs:          resp.Claims.RoleIds,
		Permissions:      resp.Claims.Permissions,
		Scopes:           resp.Claims.Scopes,
		IssuedAt:         resp.Claims.IssuedAt,
		ExpiresAt:        resp.Claims.ExpiresAt,
		NotBefore:        resp.Claims.NotBefore,
		Issuer:           resp.Claims.Issuer,
		Audience:         resp.Claims.Audience,
		IsValidated:      resp.Claims.IsValidated,
		UserRoles:        userRoles,
		Organizations:    organizations,
		Groups:           groups,
		TokenType:        resp.Claims.TokenType,
		TokenVersion:     resp.Claims.TokenVersion,
		Subject:          resp.Claims.Sub,
		SessionID:        resp.Claims.SessionId,
		JTI:              resp.Claims.Jti,
		TenantContext:    tenantContext,
		UserContextData:  userContextData,
	}

	return claims, nil
}

// Authorize checks if a user has permission to perform an action on a resource
func (c *client) Authorize(ctx context.Context, req *AuthorizeRequest) (*AuthorizeResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("authorize request cannot be nil")
	}

	// Create gRPC request
	grpcReq := &aaaPb.AuthorizeRequest{
		UserId:     req.UserID,
		TenantId:   req.TenantID,
		Resource:   req.Resource,
		Action:     req.Action,
		ResourceId: req.ResourceID,
	}

	// Call AAA service
	resp, err := c.authzClient.Authorize(ctx, grpcReq)
	if err != nil {
		return nil, fmt.Errorf("failed to authorize: %w", err)
	}

	return &AuthorizeResponse{
		Allowed: resp.Allowed,
		Reason:  resp.Reason,
	}, nil
}

// CreateUser creates a new user in the AAA system
func (c *client) CreateUser(ctx context.Context, user *AAAUser) (*AAAUser, error) {
	// Implementation would call AAA service
	// For now, return the user as-is for testing
	return user, nil
}

// GetUser retrieves a user from the AAA system
func (c *client) GetUser(ctx context.Context, userID string) (*AAAUser, error) {
	// Implementation would call AAA service
	// For now, return a mock user for testing
	return &AAAUser{
		ID:       userID,
		Username: "test-user",
		Email:    "test@example.com",
		IsActive: true,
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

// AuthenticateUser authenticates a user with credentials
func (c *client) AuthenticateUser(ctx context.Context, req *AuthenticationRequest) (*AuthenticationResponse, error) {
	// Implementation would call AAA service
	// For now, return mock response for testing
	return &AuthenticationResponse{
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
		UserContext: &UserContext{
			UserID:   "mock_user_123",
			Username: req.Username,
			Email:    req.Username + "@example.com",
			IsActive: true,
		},
	}, nil
}

// RefreshToken refreshes an access token
func (c *client) RefreshToken(ctx context.Context, refreshToken string) (*AuthenticationResponse, error) {
	// Implementation would call AAA service
	// For now, return mock response for testing
	return &AuthenticationResponse{
		AccessToken:  "new_mock_access_token",
		RefreshToken: "new_mock_refresh_token",
		ExpiresIn:    3600,
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

	// Create health check request
	req := &aaaPb.HealthCheckRequest{}

	// Call AAA service health check
	_, err := c.authClient.HealthCheck(ctx, req)
	if err != nil {
		return fmt.Errorf("AAA service health check failed: %w", err)
	}

	return nil
}
