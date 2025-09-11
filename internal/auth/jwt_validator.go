package auth

import (
    "context"
    "crypto/rsa"
    "fmt"
    "strings"
    "time"

    jwt "github.com/golang-jwt/jwt/v5"
)

// JWTValidator handles JWT token validation with AAA service
type JWTValidator struct {
    aaaClient AAAClient
    publicKey *rsa.PublicKey
    issuer    string
    audience  string
    clockSkew time.Duration
}

// NewJWTValidator creates a new JWT validator
func NewJWTValidator(aaaClient AAAClient, issuer, audience string) *JWTValidator {
    return &JWTValidator{
        aaaClient: aaaClient,
        issuer:    issuer,
        audience:  audience,
        clockSkew: 5 * time.Minute, // Allow 5 minutes clock skew
    }
}

// ValidateToken validates a JWT token and returns user context
func (v *JWTValidator) ValidateToken(ctx context.Context, tokenString string) (*TokenValidationResult, error) {
    // Clean the token string
    tokenString = strings.TrimSpace(tokenString)
    if tokenString == "" {
        return &TokenValidationResult{
            Valid: false,
            Error: fmt.Errorf("empty token"),
        }, nil
    }

    // Parse the token without verification first to get claims
    token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
    if err != nil {
        return &TokenValidationResult{
            Valid: false,
            Error: fmt.Errorf("failed to parse token: %w", err),
        }, nil
    }

    // Extract claims
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return &TokenValidationResult{
            Valid: false,
            Error: fmt.Errorf("invalid token claims"),
        }, nil
    }

    // Convert claims to our structure
    jwtClaims, err := v.convertClaims(claims)
    if err != nil {
        return &TokenValidationResult{
            Valid: false,
            Error: fmt.Errorf("failed to convert claims: %w", err),
        }, nil
    }

    // Validate token with AAA service
    isValid, err := v.aaaClient.ValidateJWT(ctx, tokenString)
    if err != nil {
        return &TokenValidationResult{
            Valid: false,
            Error: fmt.Errorf("AAA service validation failed: %w", err),
        }, nil
    }

    if !isValid {
        return &TokenValidationResult{
            Valid: false,
            Error: fmt.Errorf("token validation failed"),
        }, nil
    }

    // Create user context
    userContext := &UserContext{
        UserID:           jwtClaims.UserID,
        Username:         jwtClaims.Username,
        Email:            jwtClaims.Email,
        OrganizationID:   jwtClaims.OrganizationID,
        OrganizationName: jwtClaims.OrganizationName,
        Roles:            jwtClaims.Roles,
        Permissions:      jwtClaims.Permissions,
        IsActive:         true,
    }

    return &TokenValidationResult{
        Valid:       true,
        Claims:      jwtClaims,
        UserContext: userContext,
        Error:       nil,
    }, nil
}

// ValidateTokenOffline validates a JWT token offline (without AAA service call)
func (v *JWTValidator) ValidateTokenOffline(tokenString string) (*TokenValidationResult, error) {
    // This is a fallback method for when AAA service is unavailable
    // Parse and validate the token structure and expiration

    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // For offline validation, we would need the public key
        // For now, return an error to force online validation
        return nil, fmt.Errorf("offline validation requires public key configuration")
    })

    if err != nil {
        return &TokenValidationResult{
            Valid: false,
            Error: fmt.Errorf("offline token validation failed: %w", err),
        }, nil
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return &TokenValidationResult{
            Valid: false,
            Error: fmt.Errorf("invalid token claims"),
        }, nil
    }

    // Convert claims
    jwtClaims, err := v.convertClaims(claims)
    if err != nil {
        return &TokenValidationResult{
            Valid: false,
            Error: fmt.Errorf("failed to convert claims: %w", err),
        }, nil
    }

    // Check expiration
    if time.Now().After(jwtClaims.ExpiresAt.Add(v.clockSkew)) {
        return &TokenValidationResult{
            Valid: false,
            Error: fmt.Errorf("token expired"),
        }, nil
    }

    // Create user context
    userContext := &UserContext{
        UserID:           jwtClaims.UserID,
        Username:         jwtClaims.Username,
        Email:            jwtClaims.Email,
        OrganizationID:   jwtClaims.OrganizationID,
        OrganizationName: jwtClaims.OrganizationName,
        Roles:            jwtClaims.Roles,
        Permissions:      jwtClaims.Permissions,
        IsActive:         true,
    }

    return &TokenValidationResult{
        Valid:       true,
        Claims:      jwtClaims,
        UserContext: userContext,
        Error:       nil,
    }, nil
}

// convertClaims converts jwt.MapClaims to our JWTClaims structure
func (v *JWTValidator) convertClaims(claims jwt.MapClaims) (*JWTClaims, error) {
    jwtClaims := &JWTClaims{}

    // Extract subject (user ID)
    if sub, ok := claims["sub"].(string); ok {
        jwtClaims.UserID = sub
    } else {
        return nil, fmt.Errorf("missing or invalid subject claim")
    }

    // Extract username
    if username, ok := claims["username"].(string); ok {
        jwtClaims.Username = username
    }

    // Extract email
    if email, ok := claims["email"].(string); ok {
        jwtClaims.Email = email
    }

    // Extract organization ID
    if orgID, ok := claims["org_id"].(string); ok {
        jwtClaims.OrganizationID = orgID
    }

    // Extract organization name
    if orgName, ok := claims["org_name"].(string); ok {
        jwtClaims.OrganizationName = orgName
    }

    // Extract roles
    if rolesInterface, ok := claims["roles"]; ok {
        if rolesList, ok := rolesInterface.([]interface{}); ok {
            roles := make([]string, len(rolesList))
            for i, role := range rolesList {
                if roleStr, ok := role.(string); ok {
                    roles[i] = roleStr
                }
            }
            jwtClaims.Roles = roles
        } else if rolesStr, ok := rolesInterface.(string); ok {
            // Handle single role as string
            jwtClaims.Roles = []string{rolesStr}
        }
    }

    // Extract permissions
    if permsInterface, ok := claims["permissions"]; ok {
        if permsList, ok := permsInterface.([]interface{}); ok {
            permissions := make([]string, len(permsList))
            for i, perm := range permsList {
                if permStr, ok := perm.(string); ok {
                    permissions[i] = permStr
                }
            }
            jwtClaims.Permissions = permissions
        }
    }

    // Extract issued at
    if iat, ok := claims["iat"]; ok {
        if iatFloat, ok := iat.(float64); ok {
            jwtClaims.IssuedAt = time.Unix(int64(iatFloat), 0)
        }
    }

    // Extract expires at
    if exp, ok := claims["exp"]; ok {
        if expFloat, ok := exp.(float64); ok {
            jwtClaims.ExpiresAt = time.Unix(int64(expFloat), 0)
        } else {
            return nil, fmt.Errorf("missing or invalid expiration claim")
        }
    }

    // Extract issuer
    if iss, ok := claims["iss"].(string); ok {
        jwtClaims.Issuer = iss
    }

    // Extract audience
    if aud, ok := claims["aud"].(string); ok {
        jwtClaims.Audience = aud
    }

    // Validate issuer and audience if configured
    if v.issuer != "" && jwtClaims.Issuer != v.issuer {
        return nil, fmt.Errorf("invalid issuer: expected %s, got %s", v.issuer, jwtClaims.Issuer)
    }

    if v.audience != "" && jwtClaims.Audience != v.audience {
        return nil, fmt.Errorf("invalid audience: expected %s, got %s", v.audience, jwtClaims.Audience)
    }

    return jwtClaims, nil
}

// ExtractClaimsWithoutValidation extracts claims from a token without validation (for testing)
func (v *JWTValidator) ExtractClaimsWithoutValidation(tokenString string) (*JWTClaims, error) {
    token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
    if err != nil {
        return nil, fmt.Errorf("failed to parse token: %w", err)
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return nil, fmt.Errorf("invalid token claims")
    }

    return v.convertClaims(claims)
}
