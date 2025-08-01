package services

import (
	"context"
	"fmt"
	pb "kisanlink-ecom/internal/proto"
	"log"
)

// RolePermissionService handles role and permission-related business logic using gRPC client
type RolePermissionService struct {
	grpcClient *GRPCClient
}

// NewRolePermissionService creates a new role permission service instance
func NewRolePermissionService(grpcClient *GRPCClient) *RolePermissionService {
	return &RolePermissionService{
		grpcClient: grpcClient,
	}
}

// Role Management Methods

// CreateRole creates a new role via gRPC
func (s *RolePermissionService) CreateRole(ctx context.Context, name, description string) (*Role, error) {
	resp, err := s.grpcClient.CreateRole(ctx, name, description)
	if err != nil {
		log.Printf("Error creating role via gRPC: %v", err)
		return nil, fmt.Errorf("failed to create role: %v", err)
	}

	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("gRPC service returned status %d: %s", resp.StatusCode, resp.Message)
	}

	role := &Role{
		ID:          resp.Role.Id,
		Name:        resp.Role.Name,
		Description: resp.Role.Description,
	}

	return role, nil
}

// GetAllRoles retrieves all roles via gRPC
func (s *RolePermissionService) GetAllRoles(ctx context.Context) ([]*Role, error) {
	resp, err := s.grpcClient.GetAllRoles(ctx)
	if err != nil {
		log.Printf("Error getting all roles via gRPC: %v", err)
		return nil, fmt.Errorf("failed to get roles: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("gRPC service returned status %d: %s", resp.StatusCode, resp.Message)
	}

	roles := make([]*Role, 0, len(resp.Roles))
	for _, grpcRole := range resp.Roles {
		role := &Role{
			ID:          grpcRole.Id,
			Name:        grpcRole.Name,
			Description: grpcRole.Description,
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// GetRoleByID retrieves a role by ID via gRPC
func (s *RolePermissionService) GetRoleByID(ctx context.Context, roleID string) (*Role, error) {
	resp, err := s.grpcClient.GetRoleByID(ctx, roleID)
	if err != nil {
		log.Printf("Error getting role via gRPC: %v", err)
		return nil, fmt.Errorf("failed to get role: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("gRPC service returned status %d: %s", resp.StatusCode, resp.Message)
	}

	role := &Role{
		ID:          resp.Role.Id,
		Name:        resp.Role.Name,
		Description: resp.Role.Description,
	}

	return role, nil
}

// UpdateRole updates a role via gRPC
func (s *RolePermissionService) UpdateRole(ctx context.Context, role *Role) (*Role, error) {
	grpcRole := &pb.Role{
		Id:          role.ID,
		Name:        role.Name,
		Description: role.Description,
	}

	resp, err := s.grpcClient.UpdateRole(ctx, grpcRole)
	if err != nil {
		log.Printf("Error updating role via gRPC: %v", err)
		return nil, fmt.Errorf("failed to update role: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("gRPC service returned status %d: %s", resp.StatusCode, resp.Message)
	}

	updatedRole := &Role{
		ID:          resp.Role.Id,
		Name:        resp.Role.Name,
		Description: resp.Role.Description,
	}

	return updatedRole, nil
}

// DeleteRole deletes a role via gRPC
func (s *RolePermissionService) DeleteRole(ctx context.Context, roleID string) error {
	resp, err := s.grpcClient.DeleteRole(ctx, roleID)
	if err != nil {
		log.Printf("Error deleting role via gRPC: %v", err)
		return fmt.Errorf("failed to delete role: %v", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("gRPC service returned status %d: %s", resp.StatusCode, resp.Message)
	}

	return nil
}

// Permission Management Methods

// CreatePermission creates a new permission via gRPC
func (s *RolePermissionService) CreatePermission(ctx context.Context, name, description string) (*Permission, error) {
	resp, err := s.grpcClient.CreatePermission(ctx, name, description)
	if err != nil {
		log.Printf("Error creating permission via gRPC: %v", err)
		return nil, fmt.Errorf("failed to create permission: %v", err)
	}

	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("gRPC service returned status %d: %s", resp.StatusCode, resp.Message)
	}

	permission := &Permission{
		ID:          resp.Permission.Id,
		Name:        resp.Permission.Name,
		Description: resp.Permission.Description,
	}

	return permission, nil
}

// GetAllPermissions retrieves all permissions via gRPC
func (s *RolePermissionService) GetAllPermissions(ctx context.Context) ([]*Permission, error) {
	resp, err := s.grpcClient.GetAllPermissions(ctx)
	if err != nil {
		log.Printf("Error getting all permissions via gRPC: %v", err)
		return nil, fmt.Errorf("failed to get permissions: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("gRPC service returned status %d: %s", resp.StatusCode, resp.Message)
	}

	permissions := make([]*Permission, 0, len(resp.Permissions))
	for _, grpcPermission := range resp.Permissions {
		permission := &Permission{
			ID:          grpcPermission.Id,
			Name:        grpcPermission.Name,
			Description: grpcPermission.Description,
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// GetPermissionByID retrieves a permission by ID via gRPC
func (s *RolePermissionService) GetPermissionByID(ctx context.Context, permissionID string) (*Permission, error) {
	resp, err := s.grpcClient.GetPermissionByID(ctx, permissionID)
	if err != nil {
		log.Printf("Error getting permission via gRPC: %v", err)
		return nil, fmt.Errorf("failed to get permission: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("gRPC service returned status %d: %s", resp.StatusCode, resp.Message)
	}

	permission := &Permission{
		ID:          resp.Permission.Id,
		Name:        resp.Permission.Name,
		Description: resp.Permission.Description,
	}

	return permission, nil
}

// UpdatePermission updates a permission via gRPC
func (s *RolePermissionService) UpdatePermission(ctx context.Context, permission *Permission) (*Permission, error) {
	grpcPermission := &pb.Permission{
		Id:          permission.ID,
		Name:        permission.Name,
		Description: permission.Description,
	}

	resp, err := s.grpcClient.UpdatePermission(ctx, grpcPermission)
	if err != nil {
		log.Printf("Error updating permission via gRPC: %v", err)
		return nil, fmt.Errorf("failed to update permission: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("gRPC service returned status %d: %s", resp.StatusCode, resp.Message)
	}

	updatedPermission := &Permission{
		ID:          resp.Permission.Id,
		Name:        resp.Permission.Name,
		Description: resp.Permission.Description,
	}

	return updatedPermission, nil
}

// DeletePermission deletes a permission via gRPC
func (s *RolePermissionService) DeletePermission(ctx context.Context, permissionID string) error {
	resp, err := s.grpcClient.DeletePermission(ctx, permissionID)
	if err != nil {
		log.Printf("Error deleting permission via gRPC: %v", err)
		return fmt.Errorf("failed to delete permission: %v", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("gRPC service returned status %d: %s", resp.StatusCode, resp.Message)
	}

	return nil
}

// Role-Permission Connection Methods

// CreateRolePermissionConnection creates a connection between roles and permissions via gRPC
func (s *RolePermissionService) CreateRolePermissionConnection(ctx context.Context, roleIDs, permissionIDs []string) (*RolePermissionConnection, error) {
	resp, err := s.grpcClient.CreateRolePermissionConnection(ctx, roleIDs, permissionIDs)
	if err != nil {
		log.Printf("Error creating role-permission connection via gRPC: %v", err)
		return nil, fmt.Errorf("failed to create role-permission connection: %v", err)
	}

	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("gRPC service returned status %d: %s", resp.StatusCode, resp.Message)
	}

	connection := &RolePermissionConnection{
		ID:            resp.ConnRolePermission.Id,
		RoleIDs:       roleIDs,
		PermissionIDs: permissionIDs,
	}

	return connection, nil
}

// GetAllRolePermissions retrieves all role-permission connections via gRPC
func (s *RolePermissionService) GetAllRolePermissions(ctx context.Context) ([]*RolePermissionConnection, error) {
	resp, err := s.grpcClient.GetAllRolePermissions(ctx)
	if err != nil {
		log.Printf("Error getting all role-permission connections via gRPC: %v", err)
		return nil, fmt.Errorf("failed to get role-permission connections: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("gRPC service returned status %d: %s", resp.StatusCode, resp.Message)
	}

	connections := make([]*RolePermissionConnection, 0, len(resp.ConnRolePermission))
	for _, grpcConnection := range resp.ConnRolePermission {
		connection := &RolePermissionConnection{
			ID: grpcConnection.Id,
		}
		connections = append(connections, connection)
	}

	return connections, nil
}

// GetRolePermissionByID retrieves a role-permission connection by ID via gRPC
func (s *RolePermissionService) GetRolePermissionByID(ctx context.Context, connectionID string) (*RolePermissionConnection, error) {
	resp, err := s.grpcClient.GetRolePermissionByID(ctx, connectionID)
	if err != nil {
		log.Printf("Error getting role-permission connection via gRPC: %v", err)
		return nil, fmt.Errorf("failed to get role-permission connection: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("gRPC service returned status %d: %s", resp.StatusCode, resp.Message)
	}

	connection := &RolePermissionConnection{
		ID: resp.ConnRolePermission.Id,
	}

	return connection, nil
}

// UpdateRolePermissionConnection updates a role-permission connection via gRPC
func (s *RolePermissionService) UpdateRolePermissionConnection(ctx context.Context, connectionID string, roleIDs, permissionIDs []string) (*RolePermissionConnection, error) {
	resp, err := s.grpcClient.UpdateRolePermissionConnection(ctx, connectionID, roleIDs, permissionIDs)
	if err != nil {
		log.Printf("Error updating role-permission connection via gRPC: %v", err)
		return nil, fmt.Errorf("failed to update role-permission connection: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("gRPC service returned status %d: %s", resp.StatusCode, resp.Message)
	}

	connection := &RolePermissionConnection{
		ID:            resp.ConnRolePermission.Id,
		RoleIDs:       roleIDs,
		PermissionIDs: permissionIDs,
	}

	return connection, nil
}

// DeleteRolePermissionConnection deletes a role-permission connection via gRPC
func (s *RolePermissionService) DeleteRolePermissionConnection(ctx context.Context, connectionID string) error {
	resp, err := s.grpcClient.DeleteRolePermissionConnection(ctx, connectionID)
	if err != nil {
		log.Printf("Error deleting role-permission connection via gRPC: %v", err)
		return fmt.Errorf("failed to delete role-permission connection: %v", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("gRPC service returned status %d: %s", resp.StatusCode, resp.Message)
	}

	return nil
}

// Local models for role and permission management

// Role represents a role in the system
type Role struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Permission represents a permission in the system
type Permission struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// RolePermissionConnection represents a connection between roles and permissions
type RolePermissionConnection struct {
	ID            string   `json:"id"`
	RoleIDs       []string `json:"role_ids"`
	PermissionIDs []string `json:"permission_ids"`
}
