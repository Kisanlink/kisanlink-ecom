package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Method to required roles mapping
var methodRoles = map[string][]string{
	"/kisanlink.collaborator.v1.CollaboratorService/CreateCollaborator":     {"ADMIN", "MANAGER", "USER"},
	"/kisanlink.collaborator.v1.CollaboratorService/UpdateCollaborator":     {"ADMIN", "MANAGER", "USER"},
	"/kisanlink.collaborator.v1.CollaboratorService/GetCollaborator":        {"ADMIN", "MANAGER", "USER", "VIEWER"},
	"/kisanlink.collaborator.v1.CollaboratorService/ListCollaborators":      {"ADMIN", "MANAGER", "USER", "VIEWER"},
	"/kisanlink.collaborator.v1.CollaboratorService/DeactivateCollaborator": {"ADMIN", "MANAGER"},
	"/kisanlink.collaborator.v1.CollaboratorService/VerifyCollaborator":     {"ADMIN", "VERIFIER"},
}

// AuthorizationUnaryInterceptor checks RBAC permissions
func AuthorizationUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip authorization for public endpoints
		if isPublicEndpoint(info.FullMethod) {
			return handler(ctx, req)
		}

		// Get user claims from context (set by auth interceptor)
		claims := GetUserClaims(ctx)
		if claims == nil {
			return nil, status.Errorf(codes.PermissionDenied, "No user claims found")
		}

		// Check if user has required role
		if !hasPermission(claims, info.FullMethod) {
			return nil, status.Errorf(codes.PermissionDenied, "Insufficient permissions for method: %s", info.FullMethod)
		}

		return handler(ctx, req)
	}
}

func hasPermission(claims map[string]interface{}, method string) bool {
	// Get required roles for method
	requiredRoles, exists := methodRoles[method]
	if !exists {
		// Default: allow if no specific roles defined (will be caught by other security layers)
		return true
	}

	// Get user roles from claims
	var userRoles []string
	if rolesInterface, ok := claims["roles"]; ok {
		switch roles := rolesInterface.(type) {
		case []string:
			userRoles = roles
		case []interface{}:
			for _, r := range roles {
				if roleStr, ok := r.(string); ok {
					userRoles = append(userRoles, roleStr)
				}
			}
		case string:
			userRoles = []string{roles}
		}
	}

	// Check if user has any of the required roles
	for _, userRole := range userRoles {
		for _, requiredRole := range requiredRoles {
			if userRole == requiredRole {
				return true
			}
		}
	}

	return false
}
