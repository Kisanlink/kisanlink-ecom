package auth

import (
	"context"
	"fmt"
	"time"
)

// JWTValidator handles JWT token validation with AAA service
type JWTValidator struct {
	aaaClient Client
	jtiCache  JTICache
	issuer    string
	audience  string
	clockSkew time.Duration
}

// NewJWTValidator creates a new JWT validator
func NewJWTValidator(aaaClient Client, issuer, audience string) *JWTValidator {
	return &JWTValidator{
		aaaClient: aaaClient,
		jtiCache:  NewInMemoryJTICache(),
		issuer:    issuer,
		audience:  audience,
		clockSkew: 5 * time.Minute, // Allow 5 minutes clock skew
	}
}

// NewJWTValidatorWithJTICache creates a new JWT validator with custom JTI cache
func NewJWTValidatorWithJTICache(aaaClient Client, jtiCache JTICache, issuer, audience string) *JWTValidator {
	return &JWTValidator{
		aaaClient: aaaClient,
		jtiCache:  jtiCache,
		issuer:    issuer,
		audience:  audience,
		clockSkew: 5 * time.Minute,
	}
}

// ValidateToken validates a JWT token and returns validation result
func (v *JWTValidator) ValidateToken(ctx context.Context, tokenString string) (*TokenValidationResult, error) {
	if tokenString == "" {
		return &TokenValidationResult{
			Valid: false,
			Error: fmt.Errorf("empty token provided"),
		}, nil
	}

	// Validate token with AAA service
	tokenClaims, err := v.aaaClient.ValidateToken(ctx, tokenString)
	if err != nil {
		return &TokenValidationResult{
			Valid: false,
			Error: fmt.Errorf("AAA service validation failed: %w", err),
		}, nil
	}

	// Convert TokenClaims to JWTClaims for backward compatibility
	jwtClaims := &JWTClaims{
		UserID:           tokenClaims.UserID,
		Username:         tokenClaims.Username,
		Email:            tokenClaims.Email,
		TenantID:         tokenClaims.TenantID,
		OrganizationID:   tokenClaims.OrganizationID,
		OrganizationName: tokenClaims.OrganizationName,
		Roles:            tokenClaims.Roles,
		Permissions:      tokenClaims.Permissions,
		IssuedAt:         time.Unix(tokenClaims.IssuedAt, 0),
		ExpiresAt:        time.Unix(tokenClaims.ExpiresAt, 0),
		Issuer:           tokenClaims.Issuer,
		Audience:         tokenClaims.Audience,
	}

	// Check expiration
	if time.Now().After(jwtClaims.ExpiresAt.Add(v.clockSkew)) {
		return &TokenValidationResult{
			Valid: false,
			Error: fmt.Errorf("token has expired"),
		}, nil
	}

	// Validate issuer and audience if configured
	if v.issuer != "" && jwtClaims.Issuer != v.issuer {
		return &TokenValidationResult{
			Valid: false,
			Error: fmt.Errorf("invalid issuer: expected %s, got %s", v.issuer, jwtClaims.Issuer),
		}, nil
	}

	if v.audience != "" && jwtClaims.Audience != v.audience {
		return &TokenValidationResult{
			Valid: false,
			Error: fmt.Errorf("invalid audience: expected %s, got %s", v.audience, jwtClaims.Audience),
		}, nil
	}

	// Check for replay attacks using JTI (JWT ID)
	if tokenClaims.JTI != "" && v.jtiCache != nil {
		if err := v.jtiCache.Check(ctx, tokenClaims.JTI, jwtClaims.ExpiresAt); err != nil {
			return &TokenValidationResult{
				Valid: false,
				Error: fmt.Errorf("replay attack detected: %w", err),
			}, nil
		}
	}

	// Create user context with all enhanced fields
	userContext := &UserContext{
		UserID:           jwtClaims.UserID,
		Username:         jwtClaims.Username,
		Email:            jwtClaims.Email,
		PhoneNumber:      tokenClaims.PhoneNumber,
		CountryCode:      tokenClaims.CountryCode,
		TenantID:         jwtClaims.TenantID,
		OrganizationID:   jwtClaims.OrganizationID,
		OrganizationName: jwtClaims.OrganizationName,
		Roles:            jwtClaims.Roles,
		RoleIDs:          tokenClaims.RoleIDs,
		Permissions:      jwtClaims.Permissions,
		Scopes:           tokenClaims.Scopes,
		IsActive:         true,
		IsValidated:      tokenClaims.IsValidated,
		UserRoles:        tokenClaims.UserRoles,
		Organizations:    tokenClaims.Organizations,
		Groups:           tokenClaims.Groups,
		TokenType:        tokenClaims.TokenType,
		TokenVersion:     tokenClaims.TokenVersion,
		Subject:          tokenClaims.Subject,
		SessionID:        tokenClaims.SessionID,
		JTI:              tokenClaims.JTI,
		Issuer:           tokenClaims.Issuer,
		Audience:         tokenClaims.Audience,
		TenantContext:    tokenClaims.TenantContext,
		UserContextData:  tokenClaims.UserContextData,
	}

	return &TokenValidationResult{
		Valid:       true,
		Claims:      tokenClaims,
		UserContext: userContext,
		Error:       nil,
	}, nil
}

// ValidateTokenSimple validates a token and returns true/false
func (v *JWTValidator) ValidateTokenSimple(ctx context.Context, tokenString string) (bool, error) {
	result, err := v.ValidateToken(ctx, tokenString)
	if err != nil {
		return false, err
	}
	return result.Valid, result.Error
}
